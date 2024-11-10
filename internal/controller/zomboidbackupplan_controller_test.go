package controller

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	hordehostv1 "github.com/hordehost/zomboid-operator/api/v1"
)

var _ = Describe("ZomboidBackupPlan Controller", func() {
	var (
		ctx               context.Context
		reconciler        *ZomboidBackupPlanReconciler
		namespace         string
		server            *hordehostv1.ZomboidServer
		operatorNS        *corev1.Namespace
		applicationSecret *corev1.Secret
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create the operator's namespace
		operatorNS = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "zomboid-system",
			},
		}
		err := k8sClient.Create(ctx, operatorNS)
		if err != nil && !errors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		// Create application secrets in operator namespace
		applicationSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dropbox-application",
				Namespace: "zomboid-system",
			},
			Data: map[string][]byte{
				"app-key":    []byte("test-app-key"),
				"app-secret": []byte("test-app-secret"),
			},
		}
		Expect(k8sClient.Create(ctx, applicationSecret)).To(Succeed())
	})

	AfterEach(func() {
		// Clean up application secrets
		Expect(k8sClient.Delete(ctx, applicationSecret)).To(Succeed())
	})

	BeforeEach(func() {
		reconciler = &ZomboidBackupPlanReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		namespace = "test-namespace-" + uuid.New().String()
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}
		Expect(k8sClient.Create(ctx, ns)).To(Succeed())

		server = &hordehostv1.ZomboidServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-server",
				Namespace: namespace,
			},
		}
		Expect(k8sClient.Create(ctx, server)).To(Succeed())
	})

	When("reconciling a backup plan", func() {
		When("using any type of destination", func() {
			var (
				backupPlanName types.NamespacedName
				destination    *hordehostv1.BackupDestination
				backupPlan     *hordehostv1.ZomboidBackupPlan
				cronJob        *batchv1.CronJob
				container      corev1.Container
			)

			BeforeEach(func() {
				backupPlanName = types.NamespacedName{
					Name:      "test-backup-plan",
					Namespace: namespace,
				}

				// Create an S3 destination for testing common functionality
				s3Secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "s3-credentials",
						Namespace: namespace,
					},
					StringData: map[string]string{
						"access-key":    "test-access-key",
						"access-secret": "test-access-secret",
					},
				}
				Expect(k8sClient.Create(ctx, s3Secret)).To(Succeed())

				destination = &hordehostv1.BackupDestination{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-destination",
						Namespace: namespace,
					},
					Spec: hordehostv1.BackupDestinationSpec{
						S3: &hordehostv1.S3{
							Provider:   "AWS",
							BucketName: "test-bucket",
							AccessKeyID: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: s3Secret.Name,
								},
								Key: "access-key",
							},
							SecretAccessKey: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: s3Secret.Name,
								},
								Key: "access-secret",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, destination)).To(Succeed())

				backupPlan = &hordehostv1.ZomboidBackupPlan{
					ObjectMeta: metav1.ObjectMeta{
						Name:      backupPlanName.Name,
						Namespace: backupPlanName.Namespace,
					},
					Spec: hordehostv1.ZomboidBackupPlanSpec{
						Server: corev1.LocalObjectReference{
							Name: server.Name,
						},
						Destination: corev1.LocalObjectReference{
							Name: destination.Name,
						},
						Schedule: "*/15 * * * *",
					},
				}
				Expect(k8sClient.Create(ctx, backupPlan)).To(Succeed())

				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: backupPlanName})
				Expect(err).NotTo(HaveOccurred())

				Expect(k8sClient.Get(ctx, backupPlanName, backupPlan)).To(Succeed())

				cronJob = &batchv1.CronJob{}
				Expect(k8sClient.Get(ctx, backupPlanName, cronJob)).To(Succeed())
				container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
			})

			It("should do nothing when the server doesn't exist", func() {
				backupPlan.Spec.Server.Name = "non-existent-server"
				Expect(k8sClient.Update(ctx, backupPlan)).To(Succeed())

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: backupPlanName})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})

			It("should do nothing when the destination doesn't exist", func() {
				backupPlan.Spec.Destination.Name = "non-existent-destination"
				Expect(k8sClient.Update(ctx, backupPlan)).To(Succeed())

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: backupPlanName})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})

			It("should set owner references correctly", func() {
				Expect(backupPlan.OwnerReferences).To(HaveLen(2))

				var serverRef, destRef *metav1.OwnerReference
				for i := range backupPlan.OwnerReferences {
					ref := &backupPlan.OwnerReferences[i]
					switch ref.Name {
					case server.Name:
						serverRef = ref
					case destination.Name:
						destRef = ref
					}
				}

				Expect(serverRef).NotTo(BeNil())
				Expect(serverRef.Controller).To(BeNil())
				Expect(destRef).NotTo(BeNil())
				Expect(destRef.Controller).To(BeNil())
			})

			It("should set the CronJob owner reference correctly", func() {
				Expect(cronJob.OwnerReferences).To(HaveLen(1))
				ownerRef := cronJob.OwnerReferences[0]
				Expect(ownerRef.Name).To(Equal(backupPlan.Name))
				Expect(ownerRef.Controller).NotTo(BeNil())
				Expect(*ownerRef.Controller).To(BeTrue())
			})

			It("should set the CronJob schedule", func() {
				Expect(cronJob.Spec.Schedule).To(Equal("*/15 * * * *"))
			})

			It("should set the restart policy to Never", func() {
				Expect(cronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy).To(Equal(corev1.RestartPolicyNever))
			})

			It("should set the container name", func() {
				Expect(container.Name).To(Equal("backup"))
			})

			It("should use the correct container image", func() {
				Expect(container.Image).To(Equal("rclone/rclone:1.68.1"))
			})

			It("should configure the volume mounts correctly", func() {
				Expect(container.VolumeMounts).To(ConsistOf(
					corev1.VolumeMount{
						Name:      "backup-data",
						MountPath: "/backup",
						ReadOnly:  true,
					},
				))
			})

			It("should configure the volumes correctly", func() {
				Expect(cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes).To(ConsistOf(
					corev1.Volume{
						Name: "backup-data",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: backupPlan.Spec.Server.Name + "-backups",
							},
						},
					},
				))
			})

			Context("when the server is deleted", func() {
				BeforeEach(func() {
					// First create everything normally
					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: backupPlanName})
					Expect(err).NotTo(HaveOccurred())

					// Then delete the server
					Expect(k8sClient.Delete(ctx, server)).To(Succeed())

					// Reconcile again after server deletion
					_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: backupPlanName})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should delete the CronJob", func() {
					cronJob := &batchv1.CronJob{}
					err := k8sClient.Get(ctx, backupPlanName, cronJob)
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})

			Context("when the destination is deleted", func() {
				BeforeEach(func() {
					// First create everything normally
					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: backupPlanName})
					Expect(err).NotTo(HaveOccurred())

					// Then delete the destination
					Expect(k8sClient.Delete(ctx, destination)).To(Succeed())

					// Reconcile again after destination deletion
					_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: backupPlanName})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should delete the CronJob", func() {
					cronJob := &batchv1.CronJob{}
					err := k8sClient.Get(ctx, backupPlanName, cronJob)
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})

			It("should update the CronJob schedule when the backup plan schedule changes", func() {
				// First verify initial schedule
				Expect(cronJob.Spec.Schedule).To(Equal("*/15 * * * *"))

				// Update the schedule
				backupPlan.Spec.Schedule = "0 2 * * *"
				Expect(k8sClient.Update(ctx, backupPlan)).To(Succeed())

				// Reconcile and verify the change
				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: backupPlanName})
				Expect(err).NotTo(HaveOccurred())

				// Get the updated CronJob
				updatedCronJob := &batchv1.CronJob{}
				Expect(k8sClient.Get(ctx, backupPlanName, updatedCronJob)).To(Succeed())
				Expect(updatedCronJob.Spec.Schedule).To(Equal("0 2 * * *"))
			})
		})

		When("using a Dropbox destination", func() {
			var (
				dropboxBackupPlanName types.NamespacedName
				dropboxDestination    *hordehostv1.BackupDestination
				dropboxBackupPlan     *hordehostv1.ZomboidBackupPlan
				dropboxSecret         *corev1.Secret
				container             corev1.Container
			)

			BeforeEach(func() {
				dropboxBackupPlanName = types.NamespacedName{
					Name:      "dropbox-backup-plan",
					Namespace: namespace,
				}

				dropboxSecret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dropbox-token",
						Namespace: namespace,
					},
					StringData: map[string]string{
						"token": "test-refresh-token",
					},
				}
				Expect(k8sClient.Create(ctx, dropboxSecret)).To(Succeed())

				dropboxDestination = &hordehostv1.BackupDestination{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dropbox-destination",
						Namespace: namespace,
					},
					Spec: hordehostv1.BackupDestinationSpec{
						Dropbox: &hordehostv1.Dropbox{
							Token: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: dropboxSecret.Name,
								},
								Key: "token",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, dropboxDestination)).To(Succeed())

				dropboxBackupPlan = &hordehostv1.ZomboidBackupPlan{
					ObjectMeta: metav1.ObjectMeta{
						Name:      dropboxBackupPlanName.Name,
						Namespace: dropboxBackupPlanName.Namespace,
					},
					Spec: hordehostv1.ZomboidBackupPlanSpec{
						Server: corev1.LocalObjectReference{
							Name: server.Name,
						},
						Destination: corev1.LocalObjectReference{
							Name: dropboxDestination.Name,
						},
						Schedule: "0 3 * * *",
					},
				}
				Expect(k8sClient.Create(ctx, dropboxBackupPlan)).To(Succeed())

				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
				Expect(err).NotTo(HaveOccurred())

				cronJob := &batchv1.CronJob{}
				err = k8sClient.Get(ctx, dropboxBackupPlanName, cronJob)
				Expect(err).NotTo(HaveOccurred())

				container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
			})

			It("should configure the correct rclone command", func() {
				Expect(container.Command).To(Equal([]string{
					"rclone",
					"sync",
					"/backup",
					fmt.Sprintf("dropbox:%s/zomboid/test-server", namespace),
				}))
			})

			Context("with custom path", func() {
				Context("with leading slash", func() {
					BeforeEach(func() {
						dropboxDestination.Spec.Dropbox.Path = "/custom/backup/path"
						Expect(k8sClient.Update(ctx, dropboxDestination)).To(Succeed())

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
						Expect(err).NotTo(HaveOccurred())

						cronJob := &batchv1.CronJob{}
						err = k8sClient.Get(ctx, dropboxBackupPlanName, cronJob)
						Expect(err).NotTo(HaveOccurred())

						container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
					})

					It("should strip the leading slash from the path", func() {
						Expect(container.Command).To(Equal([]string{
							"rclone",
							"sync",
							"/backup",
							"dropbox:custom/backup/path",
						}))
					})
				})

				Context("without leading slash", func() {
					BeforeEach(func() {
						dropboxDestination.Spec.Dropbox.Path = "custom/backup/path"
						Expect(k8sClient.Update(ctx, dropboxDestination)).To(Succeed())

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
						Expect(err).NotTo(HaveOccurred())

						cronJob := &batchv1.CronJob{}
						err = k8sClient.Get(ctx, dropboxBackupPlanName, cronJob)
						Expect(err).NotTo(HaveOccurred())

						container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
					})

					It("should use the path as-is", func() {
						Expect(container.Command).To(Equal([]string{
							"rclone",
							"sync",
							"/backup",
							"dropbox:custom/backup/path",
						}))
					})
				})

				Context("with default path", func() {
					BeforeEach(func() {
						dropboxDestination.Spec.Dropbox.Path = ""
						Expect(k8sClient.Update(ctx, dropboxDestination)).To(Succeed())

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
						Expect(err).NotTo(HaveOccurred())

						cronJob := &batchv1.CronJob{}
						err = k8sClient.Get(ctx, dropboxBackupPlanName, cronJob)
						Expect(err).NotTo(HaveOccurred())

						container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
					})

					It("should use the default path format", func() {
						Expect(container.Command).To(Equal([]string{
							"rclone",
							"sync",
							"/backup",
							fmt.Sprintf("dropbox:%s/zomboid/test-server", namespace),
						}))
					})
				})
			})

			It("should configure the Dropbox environment variables", func() {
				Expect(container.Env).To(ContainElements(
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_DROPBOX_TYPE",
						Value: "dropbox",
					},
					corev1.EnvVar{
						Name: "RCLONE_CONFIG_DROPBOX_CLIENT_ID",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "dropbox-backup-plan-dropbox-application",
								},
								Key: "app-key",
							},
						},
					},
					corev1.EnvVar{
						Name: "RCLONE_CONFIG_DROPBOX_CLIENT_SECRET",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "dropbox-backup-plan-dropbox-application",
								},
								Key: "app-secret",
							},
						},
					},
					corev1.EnvVar{
						Name: "RCLONE_CONFIG_DROPBOX_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &dropboxDestination.Spec.Dropbox.Token,
						},
					},
				))
			})
		})

		When("using an S3 destination", func() {
			var (
				s3BackupPlanName types.NamespacedName
				s3Destination    *hordehostv1.BackupDestination
				s3BackupPlan     *hordehostv1.ZomboidBackupPlan
				s3Secret         *corev1.Secret
				container        corev1.Container
			)

			BeforeEach(func() {
				s3BackupPlanName = types.NamespacedName{
					Name:      "s3-backup-plan",
					Namespace: namespace,
				}

				s3Secret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "s3-credentials",
						Namespace: namespace,
					},
					StringData: map[string]string{
						"access-key":    "test-access-key",
						"access-secret": "test-access-secret",
					},
				}
				Expect(k8sClient.Create(ctx, s3Secret)).To(Succeed())

				s3Destination = &hordehostv1.BackupDestination{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "s3-destination",
						Namespace: namespace,
					},
					Spec: hordehostv1.BackupDestinationSpec{
						S3: &hordehostv1.S3{
							Provider:   "Minio",
							BucketName: "test-bucket",
							Path:       "backups/test",
							AccessKeyID: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: s3Secret.Name,
								},
								Key: "access-key",
							},
							SecretAccessKey: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: s3Secret.Name,
								},
								Key: "access-secret",
							},
							Endpoint:             "minio.example.com",
							StorageClass:         "STANDARD",
							ServerSideEncryption: "AES256",
						},
					},
				}
				Expect(k8sClient.Create(ctx, s3Destination)).To(Succeed())

				s3BackupPlan = &hordehostv1.ZomboidBackupPlan{
					ObjectMeta: metav1.ObjectMeta{
						Name:      s3BackupPlanName.Name,
						Namespace: s3BackupPlanName.Namespace,
					},
					Spec: hordehostv1.ZomboidBackupPlanSpec{
						Server: corev1.LocalObjectReference{
							Name: server.Name,
						},
						Destination: corev1.LocalObjectReference{
							Name: s3Destination.Name,
						},
						Schedule: "0 3 * * *",
					},
				}
				Expect(k8sClient.Create(ctx, s3BackupPlan)).To(Succeed())

				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: s3BackupPlanName})
				Expect(err).NotTo(HaveOccurred())

				cronJob := &batchv1.CronJob{}
				err = k8sClient.Get(ctx, s3BackupPlanName, cronJob)
				Expect(err).NotTo(HaveOccurred())

				container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
			})

			It("should configure the correct rclone command", func() {
				Expect(container.Command).To(Equal([]string{
					"rclone",
					"sync",
					"/backup",
					"s3:test-bucket/backups/test/",
				}))
			})

			It("should configure the S3 environment variables", func() {
				Expect(container.Env).To(ContainElements(
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_S3_TYPE",
						Value: "s3",
					},
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_S3_PROVIDER",
						Value: "Minio",
					},
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_S3_ENDPOINT",
						Value: "minio.example.com",
					},
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_S3_STORAGE_CLASS",
						Value: "STANDARD",
					},
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_S3_SERVER_SIDE_ENCRYPTION",
						Value: "AES256",
					},
					corev1.EnvVar{
						Name: "RCLONE_CONFIG_S3_ACCESS_KEY_ID",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: s3Destination.Spec.S3.AccessKeyID,
						},
					},
					corev1.EnvVar{
						Name: "RCLONE_CONFIG_S3_SECRET_ACCESS_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: s3Destination.Spec.S3.SecretAccessKey,
						},
					},
				))
			})

			Context("when using IAM role authentication", func() {
				BeforeEach(func() {
					s3Destination.Spec.S3.AccessKeyID = nil
					s3Destination.Spec.S3.SecretAccessKey = nil
					Expect(k8sClient.Update(ctx, s3Destination)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: s3BackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, s3BackupPlanName, cronJob)
					Expect(err).NotTo(HaveOccurred())

					container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
				})

				It("should not include AWS credentials", func() {
					Expect(container.Env).NotTo(ContainElement(HaveField("Name", "RCLONE_CONFIG_S3_ACCESS_KEY_ID")))
					Expect(container.Env).NotTo(ContainElement(HaveField("Name", "RCLONE_CONFIG_S3_SECRET_ACCESS_KEY")))
				})
			})
		})

		When("using a Google Drive destination", func() {
			var (
				googleDriveBackupPlanName types.NamespacedName
				googleDriveDestination    *hordehostv1.BackupDestination
				googleDriveBackupPlan     *hordehostv1.ZomboidBackupPlan
				googleDriveSecret         *corev1.Secret
				googleDriveAppSecret      *corev1.Secret
				container                 corev1.Container
			)

			BeforeEach(func() {
				// Create Google Drive application secret in operator namespace
				googleDriveAppSecret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "googledrive-application",
						Namespace: "zomboid-system",
					},
					Data: map[string][]byte{
						"client-id":     []byte("test-client-id"),
						"client-secret": []byte("test-client-secret"),
					},
				}
				Expect(k8sClient.Create(ctx, googleDriveAppSecret)).To(Succeed())

				googleDriveBackupPlanName = types.NamespacedName{
					Name:      "googledrive-backup-plan",
					Namespace: namespace,
				}

				googleDriveSecret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "googledrive-token",
						Namespace: namespace,
					},
					StringData: map[string]string{
						"token": "test-token",
					},
				}
				Expect(k8sClient.Create(ctx, googleDriveSecret)).To(Succeed())

				googleDriveDestination = &hordehostv1.BackupDestination{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "googledrive-destination",
						Namespace: namespace,
					},
					Spec: hordehostv1.BackupDestinationSpec{
						GoogleDrive: &hordehostv1.GoogleDrive{
							Token: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: googleDriveSecret.Name,
								},
								Key: "token",
							},
							Path:         "backups/test",
							RootFolderID: "test-root-folder",
							TeamDriveID:  "test-team-drive",
						},
					},
				}
				Expect(k8sClient.Create(ctx, googleDriveDestination)).To(Succeed())

				googleDriveBackupPlan = &hordehostv1.ZomboidBackupPlan{
					ObjectMeta: metav1.ObjectMeta{
						Name:      googleDriveBackupPlanName.Name,
						Namespace: googleDriveBackupPlanName.Namespace,
					},
					Spec: hordehostv1.ZomboidBackupPlanSpec{
						Server: corev1.LocalObjectReference{
							Name: server.Name,
						},
						Destination: corev1.LocalObjectReference{
							Name: googleDriveDestination.Name,
						},
						Schedule: "0 3 * * *",
					},
				}
				Expect(k8sClient.Create(ctx, googleDriveBackupPlan)).To(Succeed())

				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: googleDriveBackupPlanName})
				Expect(err).NotTo(HaveOccurred())

				cronJob := &batchv1.CronJob{}
				err = k8sClient.Get(ctx, googleDriveBackupPlanName, cronJob)
				Expect(err).NotTo(HaveOccurred())

				container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, googleDriveAppSecret)).To(Succeed())
			})

			It("should configure the correct rclone command", func() {
				Expect(container.Command).To(Equal([]string{
					"rclone",
					"sync",
					"/backup",
					"gdrive:backups/test",
				}))
			})

			It("should configure the Google Drive environment variables", func() {
				Expect(container.Env).To(ContainElements(
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_GDRIVE_TYPE",
						Value: "drive",
					},
					corev1.EnvVar{
						Name: "RCLONE_CONFIG_GDRIVE_CLIENT_ID",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "googledrive-backup-plan-googledrive-application",
								},
								Key: "client-id",
							},
						},
					},
					corev1.EnvVar{
						Name: "RCLONE_CONFIG_GDRIVE_CLIENT_SECRET",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "googledrive-backup-plan-googledrive-application",
								},
								Key: "client-secret",
							},
						},
					},
					corev1.EnvVar{
						Name: "RCLONE_CONFIG_GDRIVE_TOKEN",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &googleDriveDestination.Spec.GoogleDrive.Token,
						},
					},
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_GDRIVE_SCOPE",
						Value: "drive,drive.metadata.readonly",
					},
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_GDRIVE_ROOT_FOLDER_ID",
						Value: "test-root-folder",
					},
					corev1.EnvVar{
						Name:  "RCLONE_CONFIG_GDRIVE_TEAM_DRIVE",
						Value: "test-team-drive",
					},
				))
			})

			Context("with default path", func() {
				BeforeEach(func() {
					googleDriveDestination.Spec.GoogleDrive.Path = ""
					Expect(k8sClient.Update(ctx, googleDriveDestination)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: googleDriveBackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, googleDriveBackupPlanName, cronJob)
					Expect(err).NotTo(HaveOccurred())

					container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
				})

				It("should use the default path format", func() {
					Expect(container.Command).To(Equal([]string{
						"rclone",
						"sync",
						"/backup",
						fmt.Sprintf("gdrive:%s/zomboid/test-server", namespace),
					}))
				})
			})

			Context("without optional parameters", func() {
				BeforeEach(func() {
					googleDriveDestination.Spec.GoogleDrive.RootFolderID = ""
					googleDriveDestination.Spec.GoogleDrive.TeamDriveID = ""
					Expect(k8sClient.Update(ctx, googleDriveDestination)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: googleDriveBackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, googleDriveBackupPlanName, cronJob)
					Expect(err).NotTo(HaveOccurred())

					container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
				})

				It("should not include optional environment variables", func() {
					Expect(container.Env).NotTo(ContainElement(HaveField("Name", "RCLONE_CONFIG_GDRIVE_ROOT_FOLDER_ID")))
					Expect(container.Env).NotTo(ContainElement(HaveField("Name", "RCLONE_CONFIG_GDRIVE_TEAM_DRIVE")))
				})
			})
		})
	})
})
