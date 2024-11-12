package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	"github.com/hordehost/zomboid-operator/internal/players"
	"github.com/hordehost/zomboid-operator/internal/settings"
)

// ZomboidServerReconciler reconciles a ZomboidServer object
type ZomboidServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *rest.Config
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZomboidServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&zomboidv1.ZomboidServer{}).
		Named("zomboidserver").
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.findZomboidServersForSecret),
		).
		Complete(r)
}

// findZomboidServersForSecret returns reconciliation requests for ZomboidServers that reference a Secret
func (r *ZomboidServerReconciler) findZomboidServersForSecret(ctx context.Context, obj client.Object) []reconcile.Request {
	secret := obj.(*corev1.Secret)

	zomboidList := &zomboidv1.ZomboidServerList{}
	if err := r.List(ctx, zomboidList); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, zs := range zomboidList.Items {
		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      zs.Name,
				Namespace: zs.Namespace,
			},
		}

		if zs.Namespace == secret.Namespace &&
			(zs.Spec.Administrator.Password.LocalObjectReference.Name == secret.Name ||
				(zs.Spec.Password != nil && zs.Spec.Password.LocalObjectReference.Name == secret.Name)) {
			requests = append(requests, request)
		}
	}
	return requests
}

// +kubebuilder:rbac:groups=horde.host,resources=zomboidservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=horde.host,resources=zomboidservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=horde.host,resources=zomboidservers/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

// Reconcile is the main function that reconciles a ZomboidServer resource
func (r *ZomboidServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error

	zomboidServer := &zomboidv1.ZomboidServer{}
	err = r.Get(ctx, req.NamespacedName, zomboidServer)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	result, err := r.reconcileInfrastructure(ctx, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
		Type:               zomboidv1.TypeInfrastructureReady,
		ObservedGeneration: zomboidServer.Generation,
		Status:             metav1.ConditionTrue,
		Reason:             zomboidv1.ReasonInfrastructureReady,
		Message:            "All required infrastructure components are ready",
	})

	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: zomboidServer.Name, Namespace: zomboidServer.Namespace}, deployment); err != nil {
		zomboidServer.Status.Ready = false
	} else {
		zomboidServer.Status.Ready = deployment.Status.ReadyReplicas >= 1
	}

	if !zomboidServer.Status.Ready {
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:               zomboidv1.TypeReadyForPlayers,
			ObservedGeneration: zomboidServer.Generation,
			Status:             metav1.ConditionFalse,
			Reason:             zomboidv1.ReasonServerStarting,
			Message:            "Server is starting up",
		})
	} else {
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:               zomboidv1.TypeReadyForPlayers,
			ObservedGeneration: zomboidServer.Generation,
			Status:             metav1.ConditionTrue,
			Reason:             zomboidv1.ReasonServerReady,
			Message:            "Server is ready to accept players",
		})
	}

	if !zomboidServer.Status.Ready {
		return r.status(ctx, zomboidServer, &ctrl.Result{RequeueAfter: 1 * time.Second}, nil)
	}

	result, err = r.observeCurrentAllowlist(ctx, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	result, err = r.observeConnectedPlayers(ctx, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	result, err = r.observeCurrentSettings(ctx, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	result, err = r.applyDesiredSettings(ctx, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	// By default, requeue to poll for new setting updates
	return r.status(ctx, zomboidServer, &ctrl.Result{RequeueAfter: 15 * time.Second}, nil)
}

func (r *ZomboidServerReconciler) status(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer, result *ctrl.Result, err error) (ctrl.Result, error) {
	if statusErr := r.Status().Update(ctx, zomboidServer); statusErr != nil {
		if errors.IsConflict(statusErr) {
			return ctrl.Result{Requeue: true}, nil
		}
		return *result, statusErr
	}
	return *result, err
}

func commonLabels(zomboidServer *zomboidv1.ZomboidServer) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "zomboidserver",
		"app.kubernetes.io/instance":   zomboidServer.Name,
		"app.kubernetes.io/managed-by": "zomboid-operator",
	}
}

func (r *ZomboidServerReconciler) getRCONPassword(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (string, error) {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      zomboidServer.Spec.Administrator.Password.LocalObjectReference.Name,
		Namespace: zomboidServer.Namespace,
	}, secret); err != nil {
		return "", fmt.Errorf("failed to get RCON secret: %w", err)
	}

	password := string(secret.Data[zomboidServer.Spec.Administrator.Password.Key])
	if password == "" {
		return "", fmt.Errorf(
			"RCON password not found in secret %s",
			zomboidServer.Spec.Administrator.Password.LocalObjectReference.Name,
		)
	}

	return password, nil
}

func (r *ZomboidServerReconciler) observeCurrentAllowlist(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil || r.Config == nil {
		return nil, nil
	}

	hostname, port, cleanup, err := r.getServiceEndpoint(ctx,
		zomboidServer.Name+"-sqlite",
		zomboidServer.Namespace,
		12321,
	)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	allowlist, err := players.GetAllowlist(hostname, port, zomboidServer.Name)
	if err != nil {
		return nil, err
	}

	zomboidServer.Status.Allowlist = &allowlist
	return nil, nil
}

func (r *ZomboidServerReconciler) observeCurrentSettings(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil || r.Config == nil {
		return nil, nil
	}

	password, err := r.getRCONPassword(ctx, zomboidServer)
	if err != nil {
		return nil, err
	}

	hostname, port, cleanup, err := r.getServiceEndpoint(ctx,
		zomboidServer.Name+"-rcon",
		zomboidServer.Namespace,
		27015,
	)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	observed := zomboidv1.ZomboidSettings{}
	if err := settings.ReadServerOptions(hostname, port, password, &observed); err != nil {
		return nil, err
	}

	zomboidServer.Status.Settings = &observed
	zomboidServer.Status.SettingsLastObserved = &metav1.Time{Time: time.Now()}
	return nil, nil
}

func (r *ZomboidServerReconciler) observeConnectedPlayers(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil || r.Config == nil {
		return nil, nil
	}

	password, err := r.getRCONPassword(ctx, zomboidServer)
	if err != nil {
		return nil, err
	}

	hostname, port, cleanup, err := r.getServiceEndpoint(ctx,
		zomboidServer.Name+"-rcon",
		zomboidServer.Namespace,
		27015,
	)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	players, err := players.GetConnectedPlayers(ctx, hostname, port, password)
	if err != nil {
		return nil, err
	}

	connectedPlayers := make([]zomboidv1.ConnectedPlayer, len(players))
	for i, username := range players {
		connectedPlayers[i] = zomboidv1.ConnectedPlayer{
			Username: username,
		}
	}

	zomboidServer.Status.ConnectedPlayers = &connectedPlayers
	return nil, nil
}

func (r *ZomboidServerReconciler) applyDesiredSettings(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil || zomboidServer.Status.Settings == nil {
		return nil, nil
	}

	specSettings := zomboidServer.Spec.Settings
	statusSettings := zomboidServer.Status.Settings

	// Special cases handling...
	if specSettings.Identity.ResetID == nil && statusSettings.Identity.ResetID != nil {
		specSettings.Identity.ResetID = ptr.To(*statusSettings.Identity.ResetID)
	}
	if specSettings.Identity.ServerPlayerID == nil && statusSettings.Identity.ServerPlayerID != nil {
		specSettings.Identity.ServerPlayerID = ptr.To(*statusSettings.Identity.ServerPlayerID)
	}

	settings.MergeWorkshopMods(&specSettings)

	updates := settings.SettingsDiff(*statusSettings, specSettings)
	if len(updates) == 0 {
		return nil, nil
	}

	password, err := r.getRCONPassword(ctx, zomboidServer)
	if err != nil {
		return nil, err
	}

	hostname, port, cleanup, err := r.getServiceEndpoint(ctx,
		zomboidServer.Name+"-rcon",
		zomboidServer.Namespace,
		27015,
	)
	defer cleanup()
	if err != nil {
		return nil, err
	}

	if err := settings.ApplySettingsUpdates(ctx, hostname, port, password, updates, statusSettings); err != nil {
		return nil, err
	}

	zomboidServer.Status.SettingsLastObserved = &metav1.Time{Time: time.Now()}

	needsRestart := false
	for _, update := range updates {
		fieldName := update[0]
		if fieldName == "Mods" || fieldName == "WorkshopItems" {
			needsRestart = true
			break
		}
	}

	if needsRestart {
		if err := settings.RestartServer(ctx, hostname, port, password); err != nil {
			return nil, fmt.Errorf("failed to restart server after mod changes: %w", err)
		}
		return &ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	return nil, nil
}
