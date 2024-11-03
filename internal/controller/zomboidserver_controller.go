package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
)

// ZomboidServerReconciler reconciles a ZomboidServer object
type ZomboidServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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

	defer func() {
		if err := r.Status().Update(ctx, zomboidServer); err != nil {
			logger.Error(err, "Failed to update ZomboidServer status")
		}
	}()

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zomboidServer.Name + "-game-data",
			Namespace: zomboidServer.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
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
	if err != nil {
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:    zomboidv1.TypeInfrastructureReady,
			Status:  metav1.ConditionFalse,
			Reason:  zomboidv1.ReasonMissingPVC,
			Message: fmt.Sprintf("Failed to create or update PVC: %v", err),
		})
		return ctrl.Result{}, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zomboidServer.Name,
			Namespace: zomboidServer.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
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

		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": zomboidServer.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": zomboidServer.Name,
					},
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
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "game-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvc.Name,
								},
							},
						},
					},
				},
			},
		}

		return ctrl.SetControllerReference(zomboidServer, deployment, r.Scheme)
	})
	if err != nil {
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:    zomboidv1.TypeInfrastructureReady,
			Status:  metav1.ConditionFalse,
			Reason:  zomboidv1.ReasonMissingDeployment,
			Message: fmt.Sprintf("Failed to create or update Deployment: %v", err),
		})
		return ctrl.Result{}, err
	}

	meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
		Type:               zomboidv1.TypeInfrastructureReady,
		ObservedGeneration: zomboidServer.Generation,
		Status:             metav1.ConditionTrue,
		Reason:             zomboidv1.ReasonInfrastructureReady,
		Message:            "All required infrastructure components are ready",
	})

	zomboidServer.Status.Ready = deployment != nil && deployment.Status.ReadyReplicas == 1

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

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZomboidServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&zomboidv1.ZomboidServer{}).
		Named("zomboidserver").
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
