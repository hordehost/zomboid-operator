package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
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
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
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
			}),
		).
		Complete(r)
}

// +kubebuilder:rbac:groups=horde.host,resources=zomboidservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=horde.host,resources=zomboidservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=horde.host,resources=zomboidservers/finalizers,verbs=update

// Reconcile is the main function that reconciles a ZomboidServer resource
func (r *ZomboidServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error

	logger := log.FromContext(ctx)

	zomboidServer := &zomboidv1.ZomboidServer{}
	err = r.Get(ctx, req.NamespacedName, zomboidServer)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("ZomboidServer not found", "name", req.NamespacedName)
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

	result, err = r.reconcileSettings(ctx, zomboidServer)
	if result != nil {
		return r.status(ctx, zomboidServer, result, err)
	}
	if err != nil {
		return r.status(ctx, zomboidServer, &ctrl.Result{}, err)
	}

	return r.status(ctx, zomboidServer, &ctrl.Result{}, nil)
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

func (r *ZomboidServerReconciler) reconcileInfrastructure(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if err := r.reconcilePVC(ctx, zomboidServer); err != nil {
		if errors.IsConflict(err) {
			return &ctrl.Result{Requeue: true}, nil
		}
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:    zomboidv1.TypeInfrastructureReady,
			Status:  metav1.ConditionFalse,
			Reason:  zomboidv1.ReasonMissingPVC,
			Message: fmt.Sprintf("Failed to reconcile PersistentVolumeClaim: %v", err),
		})
		return nil, err
	}

	if err := r.reconcileDeployment(ctx, zomboidServer); err != nil {
		if errors.IsConflict(err) {
			return &ctrl.Result{Requeue: true}, nil
		}
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:    zomboidv1.TypeInfrastructureReady,
			Status:  metav1.ConditionFalse,
			Reason:  zomboidv1.ReasonMissingDeployment,
			Message: fmt.Sprintf("Failed to reconcile Deployment: %v", err),
		})
		return nil, err
	}

	if err := r.reconcileRCONService(ctx, zomboidServer); err != nil {
		if errors.IsConflict(err) {
			return &ctrl.Result{Requeue: true}, nil
		}
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:    zomboidv1.TypeInfrastructureReady,
			Status:  metav1.ConditionFalse,
			Reason:  zomboidv1.ReasonMissingRCONService,
			Message: fmt.Sprintf("Failed to reconcile RCON Service: %v", err),
		})
		return nil, err
	}

	if err := r.reconcileGameService(ctx, zomboidServer); err != nil {
		if errors.IsConflict(err) {
			return &ctrl.Result{Requeue: true}, nil
		}
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:    zomboidv1.TypeInfrastructureReady,
			Status:  metav1.ConditionFalse,
			Reason:  zomboidv1.ReasonMissingGameService,
			Message: fmt.Sprintf("Failed to reconcile Game Service: %v", err),
		})
		return nil, err
	}
	return nil, nil
}

func (r *ZomboidServerReconciler) reconcilePVC(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) error {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zomboidServer.Name + "-game-data",
			Namespace: zomboidServer.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		pvc.Labels = commonLabels(zomboidServer)
		if pvc.CreationTimestamp.IsZero() {
			pvc.Spec = corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: zomboidServer.Spec.Storage.StorageClassName,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: zomboidServer.Spec.Storage.Request,
					},
				},
			}
		} else {
			pvc.Spec.Resources.Requests[corev1.ResourceStorage] = zomboidServer.Spec.Storage.Request
		}
		return ctrl.SetControllerReference(zomboidServer, pvc, r.Scheme)
	})

	return err
}

func (r *ZomboidServerReconciler) reconcileDeployment(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zomboidServer.Name,
			Namespace: zomboidServer.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		labels := commonLabels(zomboidServer)
		deployment.Labels = labels

		envVars := []corev1.EnvVar{
			{
				Name:  "ZOMBOID_JVM_MAX_HEAP",
				Value: fmt.Sprintf("%dm", zomboidServer.Spec.Resources.Limits.Memory().Value()/(1024*1024)),
			},
			{
				Name:  "ZOMBOID_SERVER_NAME",
				Value: zomboidServer.Name,
			},
			{
				Name:  "ZOMBOID_SERVER_ADMIN_USERNAME",
				Value: zomboidServer.Spec.Administrator.Username,
			},
			{
				Name: "ZOMBOID_SERVER_ADMIN_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &zomboidServer.Spec.Administrator.Password,
				},
			},
		}

		if zomboidServer.Spec.Password != nil {
			envVars = append(envVars, corev1.EnvVar{
				Name: "ZOMBOID_SERVER_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: zomboidServer.Spec.Password,
				},
			})
		}

		// Get admin password secret
		adminSecret := &corev1.Secret{}
		err := r.Get(ctx, client.ObjectKey{
			Namespace: zomboidServer.Namespace,
			Name:      zomboidServer.Spec.Administrator.Password.Name,
		}, adminSecret)
		if err != nil {
			return fmt.Errorf("failed to get admin password secret: %w", err)
		}
		adminHash := sha256.Sum256(adminSecret.Data[zomboidServer.Spec.Administrator.Password.Key])

		// Initialize annotations map
		annotations := map[string]string{
			"secret/admin": hex.EncodeToString(adminHash[:]),
		}

		// Get server password secret if it exists
		if zomboidServer.Spec.Password != nil {
			serverSecret := &corev1.Secret{}
			err := r.Get(ctx, client.ObjectKey{
				Namespace: zomboidServer.Namespace,
				Name:      zomboidServer.Spec.Password.Name,
			}, serverSecret)
			if err != nil {
				return fmt.Errorf("failed to get server password secret: %w", err)
			}
			serverHash := sha256.Sum256(serverSecret.Data[zomboidServer.Spec.Password.Key])
			annotations["secret/server"] = hex.EncodeToString(serverHash[:])
		}

		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "zomboid",
							Image:           fmt.Sprintf("hordehost/zomboid-server:%s", zomboidServer.Spec.Version),
							ImagePullPolicy: corev1.PullIfNotPresent,
							Resources:       zomboidServer.Spec.Resources,
							Env:             envVars,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "game-data",
									MountPath: "/game-data",
								},
							},
							StartupProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/server/health"},
									},
								},
								InitialDelaySeconds: 0,
								PeriodSeconds:       2,
								TimeoutSeconds:      1,
								SuccessThreshold:    1,
								FailureThreshold:    60,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/server/health"},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       15,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							Lifecycle: &corev1.Lifecycle{
								PreStop: &corev1.LifecycleHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/server/rcon", "quit"},
									},
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "rcon",
									ContainerPort: 27015,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "steam",
									ContainerPort: 16261,
									Protocol:      corev1.ProtocolUDP,
								},
								{
									Name:          "raknet",
									ContainerPort: 16262,
									Protocol:      corev1.ProtocolUDP,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "game-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: zomboidServer.Name + "-game-data",
								},
							},
						},
					},
				},
			},
		}

		return ctrl.SetControllerReference(zomboidServer, deployment, r.Scheme)
	})

	return err
}

func (r *ZomboidServerReconciler) reconcileRCONService(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) error {
	rconService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zomboidServer.Name + "-rcon",
			Namespace: zomboidServer.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, rconService, func() error {
		labels := commonLabels(zomboidServer)
		rconService.Labels = labels
		rconService.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "rcon",
					Port:       27015,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromString("rcon"),
				},
			},
		}
		return ctrl.SetControllerReference(zomboidServer, rconService, r.Scheme)
	})

	return err
}

func (r *ZomboidServerReconciler) reconcileGameService(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) error {
	gameService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zomboidServer.Name,
			Namespace: zomboidServer.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, gameService, func() error {
		labels := commonLabels(zomboidServer)
		gameService.Labels = labels
		gameService.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "steam",
					Port:       16261,
					Protocol:   corev1.ProtocolUDP,
					TargetPort: intstr.FromString("steam"),
				},
				{
					Name:       "raknet",
					Port:       16262,
					Protocol:   corev1.ProtocolUDP,
					TargetPort: intstr.FromString("raknet"),
				},
			},
		}
		return ctrl.SetControllerReference(zomboidServer, gameService, r.Scheme)
	})

	return err
}

func (r *ZomboidServerReconciler) reconcileSettings(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil {
		return nil, nil
	}

	// TODO: remove this once we have a way to test the settings
	if true {
		return nil, nil
	}

	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      zomboidServer.Spec.Administrator.Password.LocalObjectReference.Name,
		Namespace: zomboidServer.Namespace,
	}, secret); err != nil {
		return nil, fmt.Errorf("failed to get RCON secret: %w", err)
	}

	password := string(secret.Data[zomboidServer.Spec.Administrator.Password.Key])
	if password == "" {
		return nil, fmt.Errorf(
			"RCON password not found in secret %s",
			zomboidServer.Spec.Administrator.Password.LocalObjectReference.Name,
		)
	}

	hostname := fmt.Sprintf("%s-rcon.%s.svc.cluster.local", zomboidServer.Name, zomboidServer.Namespace)
	port := 27015

	// TODO: how to distinguish between local development and in-cluster?
	if true {
		// For local development from a host, we need a port-forwarder to the RCON service
		// so that we can connect to it.
		parts := strings.Split(hostname, ".")
		localPort, cleanup, err := setupPortForwarder(ctx, r.Config, r.Client, parts[1], parts[0], port)
		if err != nil {
			return nil, fmt.Errorf("failed to setup port forwarder: %w", err)
		}
		defer cleanup()

		hostname = "localhost"
		port = localPort
	}

	settings, err := settings.GetServerOptions(hostname, port, password)
	if err != nil {
		return nil, err
	}

	zomboidServer.Status.SettingsLastObserved = &metav1.Time{Time: time.Now()}
	zomboidServer.Status.Settings = settings

	return nil, nil
}

func setupPortForwarder(ctx context.Context, config *rest.Config, k8sClient client.Client, namespace string, serviceName string, port int) (int, func(), error) {
	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create round tripper: %w", err)
	}

	logger := log.FromContext(ctx)
	logger.Info("Setting up port forwarder", "namespace", namespace, "serviceName", serviceName, "port", port)

	// Get the service to find its selector labels
	svc := &corev1.Service{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: serviceName}, svc); err != nil {
		return 0, nil, fmt.Errorf("failed to get service: %w", err)
	}

	// List pods matching the service selector
	pods := &corev1.PodList{}
	if err := k8sClient.List(ctx, pods, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labels.SelectorFromSet(svc.Spec.Selector),
	}); err != nil {
		return 0, nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return 0, nil, fmt.Errorf("no pods found for service %s", serviceName)
	}

	// Use the first pod's name for port forwarding
	podName := pods.Items[0].Name
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", namespace, podName)
	hostIP := strings.TrimLeft(config.Host, "htps:/")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	logger.Info("Server URL", "url", serverURL.String())

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, &serverURL)

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	fw, err := portforward.NewOnAddresses(
		dialer,
		[]string{"localhost"}, []string{fmt.Sprintf("%d:%d", 0, port)},
		stopChan, readyChan,
		os.Stdout, os.Stderr,
	)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create port forwarder: %w", err)
	}

	go func() {
		err := fw.ForwardPorts()
		if err != nil {
			log.Log.Error(err, "Error forwarding ports")
		}
	}()

	select {
	case <-readyChan:
	case <-ctx.Done():
		return 0, nil, ctx.Err()
	}

	ports, err := fw.GetPorts()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to get forwarded ports: %w", err)
	}

	cleanup := func() {
		logger.Info("Stopping port forwarder", "namespace", namespace, "serviceName", serviceName)
		close(stopChan)
	}

	return int(ports[0].Local), cleanup, nil
}
