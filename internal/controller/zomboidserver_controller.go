package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
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
	var err error

	gameDataPVC := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zomboidServer.Name + "-game-data",
			Namespace: zomboidServer.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, gameDataPVC, func() error {
		gameDataPVC.Labels = commonLabels(zomboidServer)

		storageRequest := zomboidServer.Spec.Storage.Request

		if gameDataPVC.CreationTimestamp.IsZero() {
			gameDataPVC.Spec = corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: zomboidServer.Spec.Storage.StorageClassName,
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: storageRequest,
					},
				},
			}
		} else {
			gameDataPVC.Spec.Resources.Requests[corev1.ResourceStorage] = storageRequest
		}
		return ctrl.SetControllerReference(zomboidServer, gameDataPVC, r.Scheme)
	})

	if err != nil {
		return err
	}

	if zomboidServer.Spec.Storage.WorkshopRequest != nil {
		modsPVC := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      zomboidServer.Name + "-workshop",
				Namespace: zomboidServer.Namespace,
			},
		}

		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, modsPVC, func() error {
			modsPVC.Labels = commonLabels(zomboidServer)

			storageRequest := *zomboidServer.Spec.Storage.WorkshopRequest

			if modsPVC.CreationTimestamp.IsZero() {
				modsPVC.Spec = corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					StorageClassName: zomboidServer.Spec.Storage.StorageClassName,
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: storageRequest,
						},
					},
				}
			} else {
				modsPVC.Spec.Resources.Requests[corev1.ResourceStorage] = storageRequest
			}
			return ctrl.SetControllerReference(zomboidServer, modsPVC, r.Scheme)
		})
	}

	if err != nil {
		return err
	}

	if zomboidServer.Spec.Backups.Request != nil {
		backupPVC := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      zomboidServer.Name + "-backups",
				Namespace: zomboidServer.Namespace,
			},
		}

		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, backupPVC, func() error {
			backupPVC.Labels = commonLabels(zomboidServer)

			storageRequest := *zomboidServer.Spec.Backups.Request
			storageClassName := zomboidServer.Spec.Backups.StorageClassName
			if storageClassName == nil {
				storageClassName = zomboidServer.Spec.Storage.StorageClassName
			}

			if backupPVC.CreationTimestamp.IsZero() {
				backupPVC.Spec = corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteMany,
					},
					StorageClassName: storageClassName,
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: storageRequest,
						},
					},
				}
			} else {
				backupPVC.Spec.Resources.Requests[corev1.ResourceStorage] = storageRequest
			}
			return ctrl.SetControllerReference(zomboidServer, backupPVC, r.Scheme)
		})
	}

	if err != nil {
		return err
	}

	return nil
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
			// The General log includes every connection and disconnection to
			// the RCON server, which creates a lot of ongoing noise that also
			// includes the admin password.  Unfortunately, we miss some of the
			// more useful startup logs, but this is better than spamming the
			// admin password constantly.
			{
				Name:  "ZOMBOID_SERVER_DISABLE_LOG",
				Value: "General",
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

		// Add Discord environment variables if configured
		if zomboidServer.Spec.Discord != nil {
			if zomboidServer.Spec.Discord.DiscordToken != nil {
				envVars = append(envVars, corev1.EnvVar{
					Name: "ZOMBOID_DISCORD_TOKEN",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: zomboidServer.Spec.Discord.DiscordToken,
					},
				})
			}
			if zomboidServer.Spec.Discord.DiscordChannel != nil {
				envVars = append(envVars, corev1.EnvVar{
					Name: "ZOMBOID_DISCORD_CHANNEL",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: zomboidServer.Spec.Discord.DiscordChannel,
					},
				})
			}
			if zomboidServer.Spec.Discord.DiscordChannelID != nil {
				envVars = append(envVars, corev1.EnvVar{
					Name: "ZOMBOID_DISCORD_CHANNEL_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: zomboidServer.Spec.Discord.DiscordChannelID,
					},
				})
			}
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

		image := fmt.Sprintf("hordehost/zomboid-server:%s", zomboidServer.Spec.Version)

		var workshopVolumeSource corev1.VolumeSource
		if zomboidServer.Spec.Storage.WorkshopRequest != nil {
			workshopVolumeSource = corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: zomboidServer.Name + "-workshop",
				},
			}
		} else {
			workshopVolumeSource = corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}
		}

		replicas := int32(1)
		if zomboidServer.Spec.Suspended != nil && *zomboidServer.Spec.Suspended {
			replicas = 0
		}

		// Create init containers slice with existing containers
		initContainers := []corev1.Container{
			{
				Name:            "game-data-set-owner",
				Image:           image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/usr/bin/chown", "-R", "1000:1000", "/game-data"},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: ptr.To(int64(0)),
				},
				VolumeMounts: []corev1.VolumeMount{{Name: "game-data", MountPath: "/game-data"}},
			},
			{
				Name:            "game-data-set-permissions",
				Image:           image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/usr/bin/chmod", "-R", "755", "/game-data"},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: ptr.To(int64(0)),
				},
				VolumeMounts: []corev1.VolumeMount{{Name: "game-data", MountPath: "/game-data"}},
			},
			{
				Name:            "workshop-set-owner",
				Image:           image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/usr/bin/chown", "-R", "1000:1000", "/server/steamapps"},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: ptr.To(int64(0)),
				},
				VolumeMounts: []corev1.VolumeMount{{Name: "workshop", MountPath: "/server/steamapps"}},
			},
			{
				Name:            "workshop-set-permissions",
				Image:           image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command:         []string{"/usr/bin/chmod", "-R", "755", "/server/steamapps"},
				SecurityContext: &corev1.SecurityContext{
					RunAsUser: ptr.To(int64(0)),
				},
				VolumeMounts: []corev1.VolumeMount{{Name: "workshop", MountPath: "/server/steamapps"}},
			},
		}

		// Add backup volume init containers if backup storage is requested
		if zomboidServer.Spec.Backups.Request != nil {
			initContainers = append(initContainers,
				corev1.Container{
					Name:            "backup-set-owner",
					Image:           image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"/usr/bin/chown", "-R", "1000:1000", "/game-data/backups"},
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: ptr.To(int64(0)),
					},
					VolumeMounts: []corev1.VolumeMount{{Name: "backups", MountPath: "/game-data/backups"}},
				},
				corev1.Container{
					Name:            "backup-set-permissions",
					Image:           image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"/usr/bin/chmod", "-R", "755", "/game-data/backups"},
					SecurityContext: &corev1.SecurityContext{
						RunAsUser: ptr.To(int64(0)),
					},
					VolumeMounts: []corev1.VolumeMount{{Name: "backups", MountPath: "/game-data/backups"}},
				},
			)
		}

		// Create volumes slice with existing volumes
		volumes := []corev1.Volume{
			{
				Name: "game-data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: zomboidServer.Name + "-game-data",
					},
				},
			},
			{
				Name:         "workshop",
				VolumeSource: workshopVolumeSource,
			},
		}

		// Create volume mounts slice with existing mounts
		volumeMounts := []corev1.VolumeMount{
			{
				Name:      "game-data",
				MountPath: "/game-data",
			},
			{
				Name:      "workshop",
				MountPath: "/server/steamapps",
			},
		}

		// Add backup volume and mount if requested
		if zomboidServer.Spec.Backups.Request != nil {
			volumes = append(volumes, corev1.Volume{
				Name: "backups",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: zomboidServer.Name + "-backups",
					},
				},
			})
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      "backups",
				MountPath: "/game-data/backups",
			})
		}

		// Update deployment spec with all containers, volumes, and mounts
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
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
					InitContainers: initContainers,
					Containers: []corev1.Container{
						{
							Name:            "zomboid",
							Image:           image,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Resources:       zomboidServer.Spec.Resources,
							Env:             envVars,
							VolumeMounts:    volumeMounts,
							StartupProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"/server/health"},
									},
								},
								InitialDelaySeconds: 20,
								PeriodSeconds:       5,
								TimeoutSeconds:      2,
								SuccessThreshold:    1,
								FailureThreshold:    120, // 5 seconds * 120 attempts = 10 minutes
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
					Volumes: volumes,
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

func (r *ZomboidServerReconciler) isRunningInCluster() bool {
	// If running in a pod, this env var will be set
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return true
	}

	// Alternatively, check if the serviceaccount token exists
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		return true
	}

	return false
}

func (r *ZomboidServerReconciler) observeCurrentSettings(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil {
		return nil, nil
	}

	// If we're not running against a real cluster, don't try to get settings
	if r.Config == nil {
		return nil, nil
	}

	password, err := r.getRCONPassword(ctx, zomboidServer)
	if err != nil {
		return nil, err
	}

	hostname := fmt.Sprintf("%s-rcon.%s.svc.cluster.local", zomboidServer.Name, zomboidServer.Namespace)
	port := 27015

	// Replace the hardcoded "true" with the actual check
	if !r.isRunningInCluster() {
		parts := strings.Split(hostname, ".")
		localPort, cleanup, err := SetupPortForwarder(ctx, r.Config, r.Client, parts[1], parts[0], port)
		if err != nil {
			return nil, fmt.Errorf("failed to setup port forwarder: %w", err)
		}
		defer cleanup()

		hostname = "localhost"
		port = localPort
	}

	observed := zomboidv1.ZomboidSettings{}

	if err := settings.ReadServerOptions(hostname, port, password, &observed); err != nil {
		return nil, err
	}

	zomboidServer.Status.Settings = &observed
	zomboidServer.Status.SettingsLastObserved = &metav1.Time{Time: time.Now()}

	return nil, nil
}

func mergeWorkshopMods(settings *zomboidv1.ZomboidSettings) {
	if len(settings.WorkshopMods) == 0 {
		return
	}

	var modIDs []string
	var workshopIDs []string

	// First collect any existing mods from the semicolon-separated lists
	if settings.Mods.Mods != nil && *settings.Mods.Mods != "" {
		modIDs = append(modIDs, strings.Split(*settings.Mods.Mods, ";")...)
	}
	if settings.Mods.WorkshopItems != nil && *settings.Mods.WorkshopItems != "" {
		workshopIDs = append(workshopIDs, strings.Split(*settings.Mods.WorkshopItems, ";")...)
	}

	// Add the structured workshop mods
	for _, mod := range settings.WorkshopMods {
		if mod.ModID != nil {
			modIDs = append(modIDs, *mod.ModID)
		}
		if mod.WorkshopID != nil {
			workshopIDs = append(workshopIDs, *mod.WorkshopID)
		}
	}

	// Convert back to semicolon-separated strings if we have any items
	if len(modIDs) > 0 {
		modString := strings.Join(modIDs, ";")
		settings.Mods.Mods = &modString
	}
	if len(workshopIDs) > 0 {
		workshopString := strings.Join(workshopIDs, ";")
		settings.Mods.WorkshopItems = &workshopString
	}
}

func (r *ZomboidServerReconciler) applyDesiredSettings(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) (*ctrl.Result, error) {
	if zomboidServer == nil {
		return nil, nil
	}

	specSettings := zomboidServer.Spec.Settings
	statusSettings := zomboidServer.Status.Settings

	if statusSettings == nil {
		return nil, nil
	}

	// Special cases: if the user isn't specifying a ResetID or ServerPlayerID, backfill
	// it with the observed value, because it was a server-generated unique ID, we can't
	// assume there's a default
	if specSettings.Identity.ResetID == nil && statusSettings.Identity.ResetID != nil {
		specSettings.Identity.ResetID = ptr.To(*statusSettings.Identity.ResetID)
	}
	if specSettings.Identity.ServerPlayerID == nil && statusSettings.Identity.ServerPlayerID != nil {
		specSettings.Identity.ServerPlayerID = ptr.To(*statusSettings.Identity.ServerPlayerID)
	}

	// Merge WorkshopMods into Mods before calculating differences
	mergeWorkshopMods(&specSettings)

	// Calculate differences between current and desired settings
	updates := settings.SettingsDiff(*statusSettings, specSettings)
	if len(updates) == 0 {
		return nil, nil
	}

	password, err := r.getRCONPassword(ctx, zomboidServer)
	if err != nil {
		return nil, err
	}

	// Setup RCON connection details
	hostname := fmt.Sprintf("%s-rcon.%s.svc.cluster.local", zomboidServer.Name, zomboidServer.Namespace)
	port := 27015

	// Replace the hardcoded "true" with the actual check
	if !r.isRunningInCluster() {
		parts := strings.Split(hostname, ".")
		localPort, cleanup, err := SetupPortForwarder(ctx, r.Config, r.Client, parts[1], parts[0], port)
		if err != nil {
			return nil, fmt.Errorf("failed to setup port forwarder: %w", err)
		}
		defer cleanup()

		hostname = "localhost"
		port = localPort
	}

	if err := settings.ApplySettingsUpdates(ctx, hostname, port, password, updates, statusSettings); err != nil {
		return nil, err
	}

	// If we got here, we have applied one or more settings and gotten confirmed
	// updated values for those settings, so bump the observed time
	zomboidServer.Status.SettingsLastObserved = &metav1.Time{Time: time.Now()}

	// Check if any mod-related settings were changed
	needsRestart := false
	for _, update := range updates {
		fieldName := update[0]
		if fieldName == "Mods" || fieldName == "WorkshopItems" {
			needsRestart = true
			break
		}
	}

	// If mod settings changed, restart the server using RCON quit command
	if needsRestart {
		if err := settings.RestartServer(ctx, hostname, port, password); err != nil {
			return nil, fmt.Errorf("failed to restart server after mod changes: %w", err)
		}
		return &ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	return nil, nil
}
