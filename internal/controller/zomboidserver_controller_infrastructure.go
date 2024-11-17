package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

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

	if err := r.reconcileSqliteService(ctx, zomboidServer); err != nil {
		if errors.IsConflict(err) {
			return &ctrl.Result{Requeue: true}, nil
		}
		meta.SetStatusCondition(&zomboidServer.Status.Conditions, metav1.Condition{
			Type:    zomboidv1.TypeInfrastructureReady,
			Status:  metav1.ConditionFalse,
			Reason:  zomboidv1.ReasonMissingSQLiteService,
			Message: fmt.Sprintf("Failed to reconcile SQLite Service: %v", err),
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

		serverPort := int32(16261)
		if zomboidServer.Spec.ServerPort != nil {
			serverPort = *zomboidServer.Spec.ServerPort
		}

		udpPort := int32(16262)
		if zomboidServer.Spec.UDPPort != nil {
			udpPort = *zomboidServer.Spec.UDPPort
		}

		envVars = append(envVars,
			corev1.EnvVar{
				Name:  "ZOMBOID_SERVER_PORT",
				Value: fmt.Sprintf("%d", serverPort),
			},
			corev1.EnvVar{
				Name:  "ZOMBOID_UDP_PORT",
				Value: fmt.Sprintf("%d", udpPort),
			},
		)

		// Admin credentials
		adminSecret := &corev1.Secret{}
		err := r.Get(ctx, client.ObjectKey{
			Namespace: zomboidServer.Namespace,
			Name:      zomboidServer.Spec.Administrator.Password.Name,
		}, adminSecret)
		if err != nil {
			return fmt.Errorf("failed to get admin password secret: %w", err)
		}
		adminHash := sha256.Sum256(adminSecret.Data[zomboidServer.Spec.Administrator.Password.Key])

		annotations := map[string]string{
			"secret/administrator": hex.EncodeToString(adminHash[:]),
		}

		envVars = append(
			envVars,
			corev1.EnvVar{
				Name:  "ZOMBOID_SERVER_ADMIN_USERNAME",
				Value: zomboidServer.Spec.Administrator.Username,
			},
			corev1.EnvVar{
				Name: "ZOMBOID_SERVER_ADMIN_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &zomboidServer.Spec.Administrator.Password,
				},
			},
		)

		// Server password if configured
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

			envVars = append(envVars, corev1.EnvVar{
				Name: "ZOMBOID_SERVER_PASSWORD",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: zomboidServer.Spec.Password,
				},
			})
		}

		if zomboidServer.Spec.Discord != nil {
			if zomboidServer.Spec.Discord.DiscordToken != nil {
				discordTokenSecret := &corev1.Secret{}
				err := r.Get(ctx, client.ObjectKey{
					Namespace: zomboidServer.Namespace,
					Name:      zomboidServer.Spec.Discord.DiscordToken.Name,
				}, discordTokenSecret)
				if err != nil {
					return fmt.Errorf("failed to get Discord token secret: %w", err)
				}
				discordTokenHash := sha256.Sum256(discordTokenSecret.Data[zomboidServer.Spec.Discord.DiscordToken.Key])
				annotations["secret/discord-token"] = hex.EncodeToString(discordTokenHash[:])

				envVars = append(envVars, corev1.EnvVar{
					Name: "ZOMBOID_DISCORD_TOKEN",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: zomboidServer.Spec.Discord.DiscordToken,
					},
				})
			}

			if zomboidServer.Spec.Discord.DiscordChannel != nil {
				discordChannelSecret := &corev1.Secret{}
				err := r.Get(ctx, client.ObjectKey{
					Namespace: zomboidServer.Namespace,
					Name:      zomboidServer.Spec.Discord.DiscordChannel.Name,
				}, discordChannelSecret)
				if err != nil {
					return fmt.Errorf("failed to get Discord channel secret: %w", err)
				}
				discordChannelHash := sha256.Sum256(discordChannelSecret.Data[zomboidServer.Spec.Discord.DiscordChannel.Key])
				annotations["secret/discord-channel"] = hex.EncodeToString(discordChannelHash[:])

				envVars = append(envVars, corev1.EnvVar{
					Name: "ZOMBOID_DISCORD_CHANNEL",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: zomboidServer.Spec.Discord.DiscordChannel,
					},
				})
			}

			if zomboidServer.Spec.Discord.DiscordChannelID != nil {
				discordChannelIDSecret := &corev1.Secret{}
				err := r.Get(ctx, client.ObjectKey{
					Namespace: zomboidServer.Namespace,
					Name:      zomboidServer.Spec.Discord.DiscordChannelID.Name,
				}, discordChannelIDSecret)
				if err != nil {
					return fmt.Errorf("failed to get Discord channel ID secret: %w", err)
				}
				discordChannelIDHash := sha256.Sum256(discordChannelIDSecret.Data[zomboidServer.Spec.Discord.DiscordChannelID.Key])
				annotations["secret/discord-channel-id"] = hex.EncodeToString(discordChannelIDHash[:])

				envVars = append(envVars, corev1.EnvVar{
					Name: "ZOMBOID_DISCORD_CHANNEL_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: zomboidServer.Spec.Discord.DiscordChannelID,
					},
				})
			}
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
									ContainerPort: serverPort,
									Protocol:      corev1.ProtocolUDP,
								},
								{
									Name:          "raknet",
									ContainerPort: udpPort,
									Protocol:      corev1.ProtocolUDP,
								},
							},
						},
						{
							Name:            "ws4sqlite",
							Image:           "germanorizzo/ws4sqlite:v0.16.2",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args:            []string{"--db", fmt.Sprintf("/game-data/db/%s.db", zomboidServer.Name)},
							SecurityContext: &corev1.SecurityContext{
								RunAsUser:  ptr.To(int64(1000)),
								RunAsGroup: ptr.To(int64(1000)),
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "ws4sqlite",
									ContainerPort: 12321,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "game-data",
									MountPath: "/game-data",
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

func (r *ZomboidServerReconciler) reconcileSqliteService(ctx context.Context, zomboidServer *zomboidv1.ZomboidServer) error {
	sqliteService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      zomboidServer.Name + "-sqlite",
			Namespace: zomboidServer.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sqliteService, func() error {
		labels := commonLabels(zomboidServer)
		sqliteService.Labels = labels
		sqliteService.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "ws4sqlite",
					Port:       12321,
					Protocol:   corev1.ProtocolTCP,
					TargetPort: intstr.FromString("ws4sqlite"),
				},
			},
		}
		return ctrl.SetControllerReference(zomboidServer, sqliteService, r.Scheme)
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

		serverPort := int32(16261)
		if zomboidServer.Spec.ServerPort != nil {
			serverPort = *zomboidServer.Spec.ServerPort
		}

		udpPort := int32(16262)
		if zomboidServer.Spec.UDPPort != nil {
			udpPort = *zomboidServer.Spec.UDPPort
		}

		gameService.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "steam",
					Port:       serverPort,
					Protocol:   corev1.ProtocolUDP,
					TargetPort: intstr.FromString("steam"),
				},
				{
					Name:       "raknet",
					Port:       udpPort,
					Protocol:   corev1.ProtocolUDP,
					TargetPort: intstr.FromString("raknet"),
				},
			},
		}
		return ctrl.SetControllerReference(zomboidServer, gameService, r.Scheme)
	})

	return err
}
