package controller

import (
	"context"
	"crypto/sha256"
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
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gorcon/rcon"
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
		if zs.Namespace != secret.Namespace {
			continue
		}

		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      zs.Name,
				Namespace: zs.Namespace,
			},
		}

		logger := log.FromContext(ctx)

		// Check administrator password
		if zs.Spec.Administrator.Password.LocalObjectReference.Name == secret.Name {
			logger.Info("requeueing to update administrator password", "name", zs.Name)
			requests = append(requests, request)
			continue
		}

		// Check server password if set
		if zs.Spec.Password != nil && zs.Spec.Password.LocalObjectReference.Name == secret.Name {
			logger.Info("requeueing to update server password", "name", zs.Name)
			requests = append(requests, request)
			continue
		}

		// Check Discord token, channel and channel ID if set
		if zs.Spec.Discord != nil {
			if zs.Spec.Discord.DiscordToken != nil && zs.Spec.Discord.DiscordToken.LocalObjectReference.Name == secret.Name {
				logger.Info("requeueing to update Discord token", "name", zs.Name)
				requests = append(requests, request)
				continue
			}
			if zs.Spec.Discord.DiscordChannel != nil && zs.Spec.Discord.DiscordChannel.LocalObjectReference.Name == secret.Name {
				logger.Info("requeueing to update Discord channel", "name", zs.Name)
				requests = append(requests, request)
				continue
			}
			if zs.Spec.Discord.DiscordChannelID != nil && zs.Spec.Discord.DiscordChannelID.LocalObjectReference.Name == secret.Name {
				logger.Info("requeueing to update Discord channel ID", "name", zs.Name)
				requests = append(requests, request)
				continue
			}
		}

		// Check user passwords
		for _, user := range zs.Spec.Users {
			if user.Password != nil && user.Password.LocalObjectReference.Name == secret.Name {
				logger.Info("requeueing to update user password", "name", zs.Name)
				requests = append(requests, request)
				break
			}
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

	logger := log.FromContext(ctx)
	logger.Info("reconciling", "name", req.NamespacedName)

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

	// If we're not pointing to a real cluster (like in tests), we can't do anything else
	if r.Config == nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, nil)
	}

	// Establish RCON connection for all subsequent operations
	conn, cleanup, err := r.connectRCON(ctx, zomboidServer)
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}
	defer cleanup()

	if conn == nil {
		logger.Info("no RCON connection, skipping reconciliation")
	}

	result, err = r.observeCurrentSettings(ctx, conn, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	result, err = r.applyDesiredSettings(ctx, conn, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	result, err = r.observeCurrentAllowlist(ctx, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	result, err = r.observeConnectedPlayers(ctx, conn, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	result, err = r.reconcileUsers(ctx, conn, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	// By default, requeue to poll for new setting updates
	logger.Info("reconciled", "name", req.NamespacedName)
	return r.status(ctx, zomboidServer, &ctrl.Result{RequeueAfter: 10 * time.Second}, nil)
}

func (r *ZomboidServerReconciler) status(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer, result *ctrl.Result, err error) (ctrl.Result, error) {
	if statusErr := r.Status().Update(ctx, zomboidServer); statusErr != nil {
		if errors.IsConflict(statusErr) {
			return ctrl.Result{Requeue: true}, nil
		}
		logger := log.FromContext(ctx)
		logger.Info("unrecognized status update error", "error", statusErr)
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

func (r *ZomboidServerReconciler) observeCurrentSettings(ctx context.Context, conn *rcon.Conn, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil {
		return nil, nil
	}

	observed := zomboidv1.ZomboidSettings{}
	if err := settings.ReadServerOptions(ctx, conn, &observed); err != nil {
		return nil, err
	}

	zomboidServer.Status.Settings = &observed
	zomboidServer.Status.SettingsLastObserved = &metav1.Time{Time: time.Now()}
	return nil, nil
}

func (r *ZomboidServerReconciler) applyDesiredSettings(ctx context.Context, conn *rcon.Conn, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
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

	if err := settings.ApplySettingsUpdates(ctx, conn, updates, statusSettings); err != nil {
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
		if err := settings.RestartServer(ctx, conn); err != nil {
			return nil, fmt.Errorf("failed to restart server after mod changes: %w", err)
		}
		return &ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	return nil, nil
}

func (r *ZomboidServerReconciler) observeCurrentAllowlist(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil {
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

	// Create map of existing users to preserve hashed passwords
	existingUsers := make(map[string]zomboidv1.AllowlistUser)
	for _, user := range zomboidServer.Status.Allowlist {
		existingUsers[user.Username] = user
	}

	// Merge new allowlist with existing hashed passwords
	for i, user := range allowlist {
		if existing, ok := existingUsers[user.Username]; ok {
			allowlist[i].HashedPassword = existing.HashedPassword
		}
	}

	zomboidServer.Status.Allowlist = allowlist

	return nil, nil
}

func (r *ZomboidServerReconciler) reconcileUsers(ctx context.Context, conn *rcon.Conn, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil {
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

	currentUsers := make(map[string]zomboidv1.AllowlistUser)
	for i := range zomboidServer.Status.Allowlist {
		user := &zomboidServer.Status.Allowlist[i]
		currentUsers[user.Username] = *user
	}

	desiredUsers := zomboidServer.Spec.Users[:]

	// Add the administrator to the desired users to make sure they
	// are always present and unbanned.
	desiredUsers = append(desiredUsers, zomboidv1.User{
		Username:    zomboidServer.Spec.Administrator.Username,
		Password:    &zomboidServer.Spec.Administrator.Password,
		AccessLevel: "admin",
		Banned:      false,
	})

	for _, desiredUser := range desiredUsers {
		current, exists := currentUsers[desiredUser.Username]

		if desiredUser.Password == nil {
			continue
		}

		userSecret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{
			Name:      desiredUser.Password.Name,
			Namespace: zomboidServer.Namespace,
		}, userSecret); err != nil {
			return nil, fmt.Errorf("failed to get user secret for %s: %w", desiredUser.Username, err)
		}

		password := string(userSecret.Data[desiredUser.Password.Key])
		hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))

		if !exists {
			current = zomboidv1.AllowlistUser{
				Username:       desiredUser.Username,
				HashedPassword: hashedPassword,
			}

			if err := players.AddUser(ctx, conn, desiredUser.Username, password); err != nil {
				return nil, fmt.Errorf("failed to add user %s: %w", desiredUser.Username, err)
			}

			currentUsers[desiredUser.Username] = current
			zomboidServer.Status.Allowlist = append(zomboidServer.Status.Allowlist, current)
		} else {
			if current.HashedPassword != hashedPassword {
				if err := players.SetPassword(ctx, hostname, port, zomboidServer.Name, desiredUser.Username, password); err != nil {
					return nil, fmt.Errorf("failed to set password for user %s: %w", desiredUser.Username, err)
				}
			}
			for i := range zomboidServer.Status.Allowlist {
				if zomboidServer.Status.Allowlist[i].Username == desiredUser.Username {
					zomboidServer.Status.Allowlist[i].HashedPassword = hashedPassword
					break
				}
			}
		}

		if desiredUser.AccessLevel != "" && (!exists || current.AccessLevel != desiredUser.AccessLevel) {
			if err := players.SetAccessLevel(ctx, conn, desiredUser.Username, desiredUser.AccessLevel); err != nil {
				return nil, fmt.Errorf("failed to set access level for %s: %w", desiredUser.Username, err)
			}
		}

		if exists && current.Banned != desiredUser.Banned {
			if desiredUser.Banned {
				if err := players.BanUser(ctx, conn, desiredUser.Username); err != nil {
					return nil, fmt.Errorf("failed to ban user %s: %w", desiredUser.Username, err)
				}
			} else {
				if err := players.UnbanUser(ctx, conn, desiredUser.Username); err != nil {
					return nil, fmt.Errorf("failed to unban user %s: %w", desiredUser.Username, err)
				}
			}
		}
	}

	// For open servers, we won't remove unlisted users, so we're done here
	// The default is open
	if zomboidServer.Spec.Settings.Player.Open == nil || *zomboidServer.Spec.Settings.Player.Open {
		return nil, nil
	}

	desiredUsersMap := make(map[string]struct{})
	for _, user := range zomboidServer.Spec.Users {
		desiredUsersMap[user.Username] = struct{}{}
	}

	for username := range currentUsers {
		if username == zomboidServer.Spec.Administrator.Username {
			continue
		}

		if _, desired := desiredUsersMap[username]; !desired {
			if err := players.RemoveUser(ctx, conn, username); err != nil {
				return nil, fmt.Errorf("failed to remove user %s: %w", username, err)
			}
		}
	}

	return nil, nil
}

func (r *ZomboidServerReconciler) observeConnectedPlayers(ctx context.Context, conn *rcon.Conn, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil {
		return nil, nil
	}

	players, err := players.GetConnectedPlayers(ctx, conn)
	if err != nil {
		return nil, err
	}

	connectedPlayers := make([]zomboidv1.ConnectedPlayer, len(players))
	for i, username := range players {
		connectedPlayers[i] = zomboidv1.ConnectedPlayer{
			Username: username,
		}
	}

	zomboidServer.Status.ConnectedPlayers = connectedPlayers
	return nil, nil
}
