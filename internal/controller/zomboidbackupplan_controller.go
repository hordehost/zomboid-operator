package controller

import (
	"context"
	"fmt"
	"reflect"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hordehostv1 "github.com/hordehost/zomboid-operator/api/v1"
)

// ZomboidBackupPlanReconciler reconciles a ZomboidBackupPlan object
type ZomboidBackupPlanReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZomboidBackupPlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hordehostv1.ZomboidBackupPlan{}).
		Owns(&batchv1.CronJob{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.findBackupPlansForGlobalSecret)).
		Watches(&hordehostv1.ZomboidServer{}, handler.EnqueueRequestsFromMapFunc(r.findBackupPlansForServer)).
		Watches(&hordehostv1.BackupDestination{}, handler.EnqueueRequestsFromMapFunc(r.findBackupPlansForDestination)).
		Named("ZomboidBackupPlan").
		Complete(r)
}

func (r *ZomboidBackupPlanReconciler) findBackupPlansForGlobalSecret(ctx context.Context, obj client.Object) []reconcile.Request {
	secret := obj.(*corev1.Secret)

	// We only need to re-reconcile when the global dropbox-application secret changes
	if !(secret.Namespace == "zomboid-system" && secret.Name == "dropbox-application") {
		return nil
	}

	backupPlans := &hordehostv1.ZomboidBackupPlanList{}
	if err := r.List(ctx, backupPlans); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, plan := range backupPlans.Items {
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      plan.Name,
				Namespace: plan.Namespace,
			},
		})
	}
	return requests
}

func (r *ZomboidBackupPlanReconciler) findBackupPlansForServer(ctx context.Context, obj client.Object) []reconcile.Request {
	server := obj.(*hordehostv1.ZomboidServer)

	backupPlans := &hordehostv1.ZomboidBackupPlanList{}
	if err := r.List(ctx, backupPlans); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, plan := range backupPlans.Items {
		if plan.Namespace == server.Namespace && plan.Spec.Server.Name == server.Name {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      plan.Name,
					Namespace: plan.Namespace,
				},
			})
		}
	}
	return requests
}

func (r *ZomboidBackupPlanReconciler) findBackupPlansForDestination(ctx context.Context, obj client.Object) []reconcile.Request {
	destination := obj.(*hordehostv1.BackupDestination)

	backupPlans := &hordehostv1.ZomboidBackupPlanList{}
	if err := r.List(ctx, backupPlans); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, plan := range backupPlans.Items {
		if plan.Namespace == destination.Namespace && plan.Spec.Destination.Name == destination.Name {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      plan.Name,
					Namespace: plan.Namespace,
				},
			})
		}
	}
	return requests
}

// +kubebuilder:rbac:groups=horde.host,resources=zomboidbackupplans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=horde.host,resources=zomboidbackupplans/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=horde.host,resources=zomboidbackupplans/finalizers,verbs=update
// +kubebuilder:rbac:groups=horde.host,resources=zomboidservers,verbs=get;list;watch
// +kubebuilder:rbac:groups=horde.host,resources=backupdestinations,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=[""],resources=secrets,verbs=get;list;watch;create;update;patch;delete

func (r *ZomboidBackupPlanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	backupPlan := &hordehostv1.ZomboidBackupPlan{}
	if err := r.Get(ctx, req.NamespacedName, backupPlan); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get ZomboidBackupPlan: %w", err)
	}

	server := &hordehostv1.ZomboidServer{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      backupPlan.Spec.Server.Name,
		Namespace: backupPlan.Namespace,
	}, server); err != nil {
		if errors.IsNotFound(err) {
			server = nil
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get ZomboidServer: %w", err)
		}
	}

	destination := &hordehostv1.BackupDestination{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      backupPlan.Spec.Destination.Name,
		Namespace: backupPlan.Namespace,
	}, destination); err != nil {
		if errors.IsNotFound(err) {
			destination = nil
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get BackupDestination: %w", err)
		}
	}

	if err := r.setOwnerReferences(ctx, backupPlan, server, destination); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner references: %w", err)
	}

	if err := r.reconcileDropbox(ctx, backupPlan, server, destination); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile Dropbox: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *ZomboidBackupPlanReconciler) setOwnerReferences(ctx context.Context, backupPlan *hordehostv1.ZomboidBackupPlan, server *hordehostv1.ZomboidServer, destination *hordehostv1.BackupDestination) error {
	originalRefs := backupPlan.GetOwnerReferences()

	ownerRefs := []metav1.OwnerReference{}

	if server != nil {
		ownerRefs = append(ownerRefs, metav1.OwnerReference{
			APIVersion: hordehostv1.GroupVersion.String(),
			Kind:       "ZomboidServer",
			Name:       server.Name,
			UID:        server.UID,
		})
	}

	if destination != nil {
		ownerRefs = append(ownerRefs, metav1.OwnerReference{
			APIVersion: hordehostv1.GroupVersion.String(),
			Kind:       "BackupDestination",
			Name:       destination.Name,
			UID:        destination.UID,
		})
	}

	if !reflect.DeepEqual(originalRefs, ownerRefs) {
		backupPlan.OwnerReferences = ownerRefs
		return r.Update(ctx, backupPlan)
	}

	return nil
}

func (r *ZomboidBackupPlanReconciler) reconcileDropbox(ctx context.Context, backupPlan *hordehostv1.ZomboidBackupPlan, server *hordehostv1.ZomboidServer, destination *hordehostv1.BackupDestination) error {
	if err := r.reconcileDropboxApplicationSecret(ctx, backupPlan, server, destination); err != nil {
		return fmt.Errorf("failed to reconcile Dropbox application secret: %w", err)
	}

	if err := r.reconcileDropboxCronJob(ctx, backupPlan, server, destination); err != nil {
		return fmt.Errorf("failed to reconcile Dropbox CronJob: %w", err)
	}

	return nil
}

func (r *ZomboidBackupPlanReconciler) reconcileDropboxApplicationSecret(ctx context.Context, backupPlan *hordehostv1.ZomboidBackupPlan, server *hordehostv1.ZomboidServer, destination *hordehostv1.BackupDestination) error {
	var err error

	shouldExist := server != nil && destination != nil && destination.Spec.Dropbox != nil

	targetSecret := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-dropbox-application", backupPlan.Name),
		Namespace: backupPlan.Namespace,
	}, targetSecret)
	if err == nil && !shouldExist {
		return r.Delete(ctx, targetSecret)
	} else if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Secret: %w", err)
	}

	if !shouldExist {
		return nil
	}

	sourceSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      "dropbox-application",
		Namespace: "zomboid-system",
	}, sourceSecret); err != nil {
		return fmt.Errorf("failed to get source Dropbox credentials: %w", err)
	}

	targetSecret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-dropbox-application", backupPlan.Name),
			Namespace: backupPlan.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, targetSecret, func() error {
		targetSecret.Data = make(map[string][]byte)
		targetSecret.Data["app-key"] = sourceSecret.Data["app-key"]
		targetSecret.Data["app-secret"] = sourceSecret.Data["app-secret"]

		return controllerutil.SetOwnerReference(backupPlan, targetSecret, r.Scheme)
	})

	return err
}

func (r *ZomboidBackupPlanReconciler) reconcileDropboxCronJob(ctx context.Context, backupPlan *hordehostv1.ZomboidBackupPlan, server *hordehostv1.ZomboidServer, destination *hordehostv1.BackupDestination) error {
	var err error

	shouldExist := server != nil && destination != nil && destination.Spec.Dropbox != nil

	cronJob := &batchv1.CronJob{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      backupPlan.Name,
		Namespace: backupPlan.Namespace,
	}, cronJob)

	if err == nil && !shouldExist {
		return r.Delete(ctx, cronJob)
	} else if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get CronJob: %w", err)
	}

	if !shouldExist {
		return nil
	}

	cronJob = &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      backupPlan.Name,
			Namespace: backupPlan.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, cronJob, func() error {
		if err := controllerutil.SetControllerReference(backupPlan, cronJob, r.Scheme); err != nil {
			return err
		}

		remotePath := fmt.Sprintf("/%s/zomboid/%s", backupPlan.Namespace, backupPlan.Spec.Server.Name)
		if destination.Spec.Dropbox != nil && destination.Spec.Dropbox.RemotePath != "" {
			remotePath = destination.Spec.Dropbox.RemotePath
		}

		cronJob.Spec = batchv1.CronJobSpec{
			Schedule: backupPlan.Spec.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  "backup",
									Image: "offen/docker-volume-backup:v2.43.0",
									Env: []corev1.EnvVar{
										{
											Name:  "DROPBOX_REMOTE_PATH",
											Value: remotePath,
										},
										{
											Name: "DROPBOX_APP_KEY",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: fmt.Sprintf("%s-dropbox-application", backupPlan.Name),
													},
													Key: "app-key",
												},
											},
										},
										{
											Name: "DROPBOX_APP_SECRET",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: fmt.Sprintf("%s-dropbox-application", backupPlan.Name),
													},
													Key: "app-secret",
												},
											},
										},
										{
											Name: "DROPBOX_REFRESH_TOKEN",
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: destination.Spec.Dropbox.RefreshToken.Name,
													},
													Key: destination.Spec.Dropbox.RefreshToken.Key,
												},
											},
										},
									},
									Command: []string{"/usr/bin/backup"},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "backup-data",
											MountPath: "/backup",
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "backup-data",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: backupPlan.Spec.Server.Name + "-backups",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		return nil
	})

	return err
}
