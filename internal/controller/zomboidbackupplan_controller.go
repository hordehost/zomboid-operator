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

	if err := r.reconcileApplicationSecret(ctx, backupPlan, server, destination); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile application secret: %w", err)
	}

	if err := r.reconcileCronJob(ctx, backupPlan, server, destination); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to reconcile CronJob: %w", err)
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

func (r *ZomboidBackupPlanReconciler) reconcileApplicationSecret(ctx context.Context, backupPlan *hordehostv1.ZomboidBackupPlan, server *hordehostv1.ZomboidServer, destination *hordehostv1.BackupDestination) error {
	type applicationSecret struct {
		sourceName      string
		sourceNamespace string
		targetName      string
		keys            []string
	}

	possibleSecrets := map[string]*applicationSecret{
		"dropbox": {
			sourceName:      "dropbox-application",
			sourceNamespace: "zomboid-system",
			targetName:      fmt.Sprintf("%s-dropbox-application", backupPlan.Name),
			keys:            []string{"app-key", "app-secret"},
		},
	}

	var desiredSecret *applicationSecret
	if server != nil && destination != nil {
		switch {
		case destination.Spec.Dropbox != nil:
			desiredSecret = possibleSecrets["dropbox"]
		}
	}

	// Delete any existing secrets that shouldn't exist
	for _, possibleSecret := range possibleSecrets {
		if desiredSecret != nil && possibleSecret.targetName == desiredSecret.targetName {
			continue
		}

		secret := &corev1.Secret{}
		err := r.Get(ctx, types.NamespacedName{
			Name:      possibleSecret.targetName,
			Namespace: backupPlan.Namespace,
		}, secret)
		if err == nil {
			if err := r.Delete(ctx, secret); err != nil {
				return fmt.Errorf("failed to delete secret %s: %w", possibleSecret.targetName, err)
			}
		} else if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get secret %s: %w", possibleSecret.targetName, err)
		}
	}

	// If no secret should exist, we're done
	if desiredSecret == nil {
		return nil
	}

	// Create/update the desired secret
	sourceSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Name:      desiredSecret.sourceName,
		Namespace: desiredSecret.sourceNamespace,
	}, sourceSecret); err != nil {
		return fmt.Errorf("failed to get source credentials: %w", err)
	}

	targetSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      desiredSecret.targetName,
			Namespace: backupPlan.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, targetSecret, func() error {
		targetSecret.Data = make(map[string][]byte)
		for _, key := range desiredSecret.keys {
			targetSecret.Data[key] = sourceSecret.Data[key]
		}
		return controllerutil.SetOwnerReference(backupPlan, targetSecret, r.Scheme)
	})

	return err
}

func (r *ZomboidBackupPlanReconciler) reconcileCronJob(ctx context.Context, backupPlan *hordehostv1.ZomboidBackupPlan, server *hordehostv1.ZomboidServer, destination *hordehostv1.BackupDestination) error {
	var err error

	var container *corev1.Container
	if destination != nil {
		switch {
		case destination.Spec.Dropbox != nil:
			container = r.dropboxContainer(*destination.Spec.Dropbox, backupPlan)
		case destination.Spec.S3 != nil:
			container = r.s3Container(*destination.Spec.S3)
		}
	}

	// If no provider is active or server is missing, we shouldn't have a CronJob
	shouldExist := server != nil && container != nil

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

		cronJob.Spec = batchv1.CronJobSpec{
			Schedule: backupPlan.Spec.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers:    []corev1.Container{*container},
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

func (r *ZomboidBackupPlanReconciler) dropboxContainer(dropbox hordehostv1.Dropbox, backupPlan *hordehostv1.ZomboidBackupPlan) *corev1.Container {
	remotePath := fmt.Sprintf("/%s/zomboid/%s", backupPlan.Namespace, backupPlan.Spec.Server.Name)
	if dropbox.RemotePath != "" {
		remotePath = dropbox.RemotePath
	}

	env := []corev1.EnvVar{
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
						Name: dropbox.RefreshToken.Name,
					},
					Key: dropbox.RefreshToken.Key,
				},
			},
		},
	}

	return &corev1.Container{
		Name:    "backup",
		Image:   "offen/docker-volume-backup:v2.43.0",
		Command: []string{"/usr/bin/backup"},
		Env:     env,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "backup-data",
				MountPath: "/backup",
			},
		},
	}
}

func (r *ZomboidBackupPlanReconciler) s3Container(s3 hordehostv1.S3) *corev1.Container {
	env := []corev1.EnvVar{
		{
			Name:  "AWS_S3_BUCKET_NAME",
			Value: s3.BucketName,
		},
	}

	if s3.Path != "" {
		env = append(env, corev1.EnvVar{
			Name:  "AWS_S3_PATH",
			Value: s3.Path,
		})
	}

	if s3.AccessKeyID != nil {
		env = append(env, corev1.EnvVar{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: s3.AccessKeyID,
			},
		})
	}

	if s3.SecretAccessKey != nil {
		env = append(env, corev1.EnvVar{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: s3.SecretAccessKey,
			},
		})
	}

	if s3.IAMRoleEndpoint != "" {
		env = append(env, corev1.EnvVar{
			Name:  "AWS_IAM_ROLE_ENDPOINT",
			Value: s3.IAMRoleEndpoint,
		})
	}

	if s3.Endpoint != "" {
		env = append(env, corev1.EnvVar{
			Name:  "AWS_ENDPOINT",
			Value: s3.Endpoint,
		})
	}

	if s3.EndpointProtocol != "" {
		env = append(env, corev1.EnvVar{
			Name:  "AWS_ENDPOINT_PROTO",
			Value: s3.EndpointProtocol,
		})

		// Only set AWS_ENDPOINT_INSECURE when using HTTPS protocol
		if s3.EndpointProtocol == "https" && s3.EndpointInsecure {
			env = append(env, corev1.EnvVar{
				Name:  "AWS_ENDPOINT_INSECURE",
				Value: "true",
			})
		}
	}

	if s3.EndpointCACert != "" {
		env = append(env, corev1.EnvVar{
			Name:  "AWS_ENDPOINT_CA_CERT",
			Value: s3.EndpointCACert,
		})
	}

	if s3.StorageClass != "" {
		env = append(env, corev1.EnvVar{
			Name:  "AWS_STORAGE_CLASS",
			Value: s3.StorageClass,
		})
	}

	if s3.PartSize != nil {
		env = append(env, corev1.EnvVar{
			Name:  "AWS_PART_SIZE",
			Value: fmt.Sprintf("%d", *s3.PartSize),
		})
	}

	return &corev1.Container{
		Name:    "backup",
		Image:   "offen/docker-volume-backup:v2.43.0",
		Command: []string{"/usr/bin/backup"},
		Env:     env,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "backup-data",
				MountPath: "/backup",
			},
		},
	}
}
