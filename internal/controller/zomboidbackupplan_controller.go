package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"

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
		"googledrive": {
			sourceName:      "googledrive-application",
			sourceNamespace: "zomboid-system",
			targetName:      fmt.Sprintf("%s-googledrive-application", backupPlan.Name),
			keys:            []string{"client-id", "client-secret"},
		},
	}

	var desiredSecret *applicationSecret
	if server != nil && destination != nil {
		switch {
		case destination.Spec.Dropbox != nil:
			desiredSecret = possibleSecrets["dropbox"]
		case destination.Spec.GoogleDrive != nil:
			desiredSecret = possibleSecrets["googledrive"]
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

	var env []corev1.EnvVar
	var remotePath string
	if destination != nil {
		switch {
		case destination.Spec.Dropbox != nil:
			env, remotePath = r.dropboxConfiguration(*destination.Spec.Dropbox, backupPlan)
		case destination.Spec.GoogleDrive != nil:
			env, remotePath = r.googleDriveConfiguration(*destination.Spec.GoogleDrive, backupPlan)
		case destination.Spec.S3 != nil:
			env, remotePath = r.s3Configuration(*destination.Spec.S3)
		}
	}

	// If no provider is active or server is missing, we shouldn't have a CronJob
	shouldExist := server != nil && len(env) > 0

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

		container := corev1.Container{
			Name:  "backup",
			Image: "rclone/rclone:1.68.1",
			Command: []string{
				"rclone",
				"sync",
				"/backup",
				remotePath,
			},
			Env: env,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "backup-data",
					MountPath: "/backup",
					ReadOnly:  true,
				},
			},
		}

		cronJob.Spec = batchv1.CronJobSpec{
			Schedule: backupPlan.Spec.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers:    []corev1.Container{container},
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

func (r *ZomboidBackupPlanReconciler) dropboxConfiguration(dropbox hordehostv1.Dropbox, backupPlan *hordehostv1.ZomboidBackupPlan) ([]corev1.EnvVar, string) {
	var remotePath string
	if dropbox.Path != "" {
		// Strip leading slash for app folder scoping
		remotePath = strings.TrimPrefix(dropbox.Path, "/")
	} else {
		remotePath = fmt.Sprintf("%s/zomboid/%s", backupPlan.Namespace, backupPlan.Spec.Server.Name)
	}

	env := []corev1.EnvVar{
		{
			Name:  "RCLONE_CONFIG_DROPBOX_TYPE",
			Value: "dropbox",
		},
		{
			Name: "RCLONE_CONFIG_DROPBOX_CLIENT_ID",
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
			Name: "RCLONE_CONFIG_DROPBOX_CLIENT_SECRET",
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
			Name: "RCLONE_CONFIG_DROPBOX_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &dropbox.Token,
			},
		},
	}

	return env, fmt.Sprintf("dropbox:%s", remotePath)
}

func (r *ZomboidBackupPlanReconciler) s3Configuration(s3 hordehostv1.S3) ([]corev1.EnvVar, string) {
	env := []corev1.EnvVar{
		{
			Name:  "RCLONE_CONFIG_S3_TYPE",
			Value: "s3",
		},
		{
			Name:  "RCLONE_CONFIG_S3_PROVIDER",
			Value: s3.Provider,
		},
	}

	if s3.Region != "" {
		env = append(env, corev1.EnvVar{
			Name:  "RCLONE_CONFIG_S3_REGION",
			Value: s3.Region,
		})
	}

	if s3.Endpoint != "" {
		env = append(env, corev1.EnvVar{
			Name:  "RCLONE_CONFIG_S3_ENDPOINT",
			Value: s3.Endpoint,
		})
	}

	if s3.AccessKeyID != nil {
		env = append(env, corev1.EnvVar{
			Name: "RCLONE_CONFIG_S3_ACCESS_KEY_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: s3.AccessKeyID,
			},
		})
	}

	if s3.SecretAccessKey != nil {
		env = append(env, corev1.EnvVar{
			Name: "RCLONE_CONFIG_S3_SECRET_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: s3.SecretAccessKey,
			},
		})
	}

	if s3.StorageClass != "" {
		env = append(env, corev1.EnvVar{
			Name:  "RCLONE_CONFIG_S3_STORAGE_CLASS",
			Value: s3.StorageClass,
		})
	}

	if s3.ServerSideEncryption != "" {
		env = append(env, corev1.EnvVar{
			Name:  "RCLONE_CONFIG_S3_SERVER_SIDE_ENCRYPTION",
			Value: s3.ServerSideEncryption,
		})
	}

	s3Path := s3.Path
	if s3Path != "" && !strings.HasSuffix(s3Path, "/") {
		s3Path += "/"
	}

	return env, fmt.Sprintf("s3:%s/%s", s3.BucketName, s3Path)
}

func (r *ZomboidBackupPlanReconciler) googleDriveConfiguration(googleDrive hordehostv1.GoogleDrive, backupPlan *hordehostv1.ZomboidBackupPlan) ([]corev1.EnvVar, string) {
	var remotePath string
	if googleDrive.Path != "" {
		remotePath = strings.TrimPrefix(googleDrive.Path, "/")
	} else {
		remotePath = fmt.Sprintf("%s/zomboid/%s", backupPlan.Namespace, backupPlan.Spec.Server.Name)
	}

	env := []corev1.EnvVar{
		{
			Name:  "RCLONE_CONFIG_GDRIVE_TYPE",
			Value: "drive",
		},
		{
			Name: "RCLONE_CONFIG_GDRIVE_CLIENT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-googledrive-application", backupPlan.Name),
					},
					Key: "client-id",
				},
			},
		},
		{
			Name: "RCLONE_CONFIG_GDRIVE_CLIENT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-googledrive-application", backupPlan.Name),
					},
					Key: "client-secret",
				},
			},
		},
		{
			Name: "RCLONE_CONFIG_GDRIVE_TOKEN",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &googleDrive.Token,
			},
		},
		{
			Name:  "RCLONE_CONFIG_GDRIVE_SCOPE",
			Value: "drive,drive.metadata.readonly",
		},
	}

	if googleDrive.RootFolderID != "" {
		env = append(env, corev1.EnvVar{
			Name:  "RCLONE_CONFIG_GDRIVE_ROOT_FOLDER_ID",
			Value: googleDrive.RootFolderID,
		})
	}

	if googleDrive.TeamDriveID != "" {
		env = append(env, corev1.EnvVar{
			Name:  "RCLONE_CONFIG_GDRIVE_TEAM_DRIVE",
			Value: googleDrive.TeamDriveID,
		})
	}

	return env, fmt.Sprintf("gdrive:%s", remotePath)
}
