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

		// This secret needs to exist just because we happen to be using
		// dropbox for the common backup destination tests.
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

				// Uses a Dropbox destination just because we need something, but that
				// doesn't matter for these tests
				secret := &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "dropbox-token",
						Namespace: namespace,
					},
					StringData: map[string]string{
						"token": "test-refresh-token",
					},
				}
				Expect(k8sClient.Create(ctx, secret)).To(Succeed())

				destination = &hordehostv1.BackupDestination{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-destination",
						Namespace: namespace,
					},
					Spec: hordehostv1.BackupDestinationSpec{
						Dropbox: &hordehostv1.Dropbox{
							RefreshToken: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: secret.Name,
								},
								Key: "token",
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

			It("should configure the volume mounts correctly", func() {
				Expect(container.VolumeMounts).To(ConsistOf(
					corev1.VolumeMount{
						Name:      "backup-data",
						MountPath: "/backup",
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
				dropboxApplicationSecret *corev1.Secret
				dropboxBackupPlanName    types.NamespacedName
				dropboxDestination       *hordehostv1.BackupDestination
				dropboxBackupPlan        *hordehostv1.ZomboidBackupPlan
				dropboxSecret            *corev1.Secret
				container                corev1.Container
			)

			BeforeEach(func() {
				// This works because we created the application secret in the
				// BeforeEach block for the whole test suite.  Other backup
				// providers would need to create and delete application secrets
				// here as appropriate.
				dropboxApplicationSecret = applicationSecret
			})

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
							RefreshToken: corev1.SecretKeySelector{
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

				Expect(k8sClient.Get(ctx, dropboxBackupPlanName, dropboxBackupPlan)).To(Succeed())

				cronJob := &batchv1.CronJob{}
				err = k8sClient.Get(ctx, dropboxBackupPlanName, cronJob)
				Expect(err).NotTo(HaveOccurred())

				container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
			})

			It("should use the correct container image", func() {
				Expect(container.Image).To(Equal("offen/docker-volume-backup:v2.43.0"))
			})

			Context("with default remote path", func() {
				It("should set the Dropbox remote path to the default format", func() {
					Expect(container.Env).To(ContainElement(corev1.EnvVar{
						Name:  "DROPBOX_REMOTE_PATH",
						Value: fmt.Sprintf("/%s/zomboid/%s", namespace, server.Name),
					}))
				})
			})

			Context("with custom remote path", func() {
				BeforeEach(func() {
					dropboxDestination.Spec.Dropbox.RemotePath = "/custom/backup/path"
					Expect(k8sClient.Update(ctx, dropboxDestination)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, dropboxBackupPlanName, cronJob)
					Expect(err).NotTo(HaveOccurred())

					container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
				})

				It("should use the configured remote path", func() {
					Expect(container.Env).To(ContainElement(corev1.EnvVar{
						Name:  "DROPBOX_REMOTE_PATH",
						Value: "/custom/backup/path",
					}))
				})
			})

			It("should configure the Dropbox app credentials", func() {
				Expect(container.Env).To(ContainElements(
					corev1.EnvVar{
						Name: "DROPBOX_APP_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
								},
								Key: "app-key",
							},
						},
					},
					corev1.EnvVar{
						Name: "DROPBOX_APP_SECRET",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
								},
								Key: "app-secret",
							},
						},
					},
				))
			})

			It("should configure the Dropbox refresh token from the secret", func() {
				Expect(container.Env).To(ContainElement(corev1.EnvVar{
					Name: "DROPBOX_REFRESH_TOKEN",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &dropboxDestination.Spec.Dropbox.RefreshToken,
					},
				}))
			})

			It("should copy the application credentials to the target namespace", func() {
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dropboxBackupPlan.Name,
						Namespace: dropboxBackupPlan.Namespace,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				copiedSecret := &corev1.Secret{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
					Namespace: dropboxBackupPlan.Namespace,
				}, copiedSecret)).To(Succeed())

				Expect(copiedSecret.Data).To(HaveKeyWithValue("app-key", []byte("test-app-key")))
				Expect(copiedSecret.Data).To(HaveKeyWithValue("app-secret", []byte("test-app-secret")))

				Expect(copiedSecret.OwnerReferences).To(HaveLen(1))
				Expect(copiedSecret.OwnerReferences[0].Name).To(Equal(dropboxBackupPlan.Name))
			})

			It("should update the copied secret when the source changes", func() {
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dropboxBackupPlan.Name,
						Namespace: dropboxBackupPlan.Namespace,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				dropboxApplicationSecret.Data["app-key"] = []byte("updated-app-key")
				Expect(k8sClient.Update(ctx, dropboxApplicationSecret)).To(Succeed())

				_, err = reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dropboxBackupPlan.Name,
						Namespace: dropboxBackupPlan.Namespace,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				copiedSecret := &corev1.Secret{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
					Namespace: dropboxBackupPlan.Namespace,
				}, copiedSecret)).To(Succeed())
				Expect(copiedSecret.Data).To(HaveKeyWithValue("app-key", []byte("updated-app-key")))
			})

			It("should delete the copied secret when switching to a non-Dropbox destination", func() {
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dropboxBackupPlan.Name,
						Namespace: dropboxBackupPlan.Namespace,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				dropboxDestination.Spec.Dropbox = nil
				Expect(k8sClient.Update(ctx, dropboxDestination)).To(Succeed())

				_, err = reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      dropboxBackupPlan.Name,
						Namespace: dropboxBackupPlan.Namespace,
					},
				})
				Expect(err).NotTo(HaveOccurred())

				copiedSecret := &corev1.Secret{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
					Namespace: dropboxBackupPlan.Namespace,
				}, copiedSecret)
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})

			Context("when switching from Dropbox to non-Dropbox destination", func() {
				BeforeEach(func() {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: dropboxBackupPlanName,
					})
					Expect(err).NotTo(HaveOccurred())

					secret := &corev1.Secret{}
					err = k8sClient.Get(ctx, types.NamespacedName{
						Name:      fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
						Namespace: dropboxBackupPlan.Namespace,
					}, secret)
					Expect(err).NotTo(HaveOccurred())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, types.NamespacedName{
						Name:      dropboxBackupPlan.Name,
						Namespace: dropboxBackupPlan.Namespace,
					}, cronJob)
					Expect(err).NotTo(HaveOccurred())

					dropboxDestination.Spec.Dropbox = nil
					Expect(k8sClient.Update(ctx, dropboxDestination)).To(Succeed())

					_, err = reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: dropboxBackupPlanName,
					})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should delete both the secret and CronJob", func() {
					secret := &corev1.Secret{}
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
						Namespace: dropboxBackupPlan.Namespace,
					}, secret)
					Expect(errors.IsNotFound(err)).To(BeTrue())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, types.NamespacedName{
						Name:      dropboxBackupPlan.Name,
						Namespace: dropboxBackupPlan.Namespace,
					}, cronJob)
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})

			Context("when the server is deleted", func() {
				BeforeEach(func() {
					// First create everything normally
					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					// Then delete the server
					Expect(k8sClient.Delete(ctx, server)).To(Succeed())

					// Reconcile again after server deletion
					_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should delete both the CronJob and application secret", func() {
					cronJob := &batchv1.CronJob{}
					err := k8sClient.Get(ctx, dropboxBackupPlanName, cronJob)
					Expect(errors.IsNotFound(err)).To(BeTrue())

					secret := &corev1.Secret{}
					err = k8sClient.Get(ctx, types.NamespacedName{
						Name:      fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
						Namespace: dropboxBackupPlan.Namespace,
					}, secret)
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})

			Context("when the destination is deleted", func() {
				BeforeEach(func() {
					// First create everything normally
					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					// Then delete the destination
					Expect(k8sClient.Delete(ctx, dropboxDestination)).To(Succeed())

					// Reconcile again after destination deletion
					_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: dropboxBackupPlanName})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should delete both the CronJob and application secret", func() {
					cronJob := &batchv1.CronJob{}
					err := k8sClient.Get(ctx, dropboxBackupPlanName, cronJob)
					Expect(errors.IsNotFound(err)).To(BeTrue())

					secret := &corev1.Secret{}
					err = k8sClient.Get(ctx, types.NamespacedName{
						Name:      fmt.Sprintf("%s-dropbox-application", dropboxBackupPlan.Name),
						Namespace: dropboxBackupPlan.Namespace,
					}, secret)
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
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
							Endpoint:         "minio.example.com",
							EndpointProtocol: "https",
							StorageClass:     "STANDARD",
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

				Expect(k8sClient.Get(ctx, s3BackupPlanName, s3BackupPlan)).To(Succeed())

				cronJob := &batchv1.CronJob{}
				err = k8sClient.Get(ctx, s3BackupPlanName, cronJob)
				Expect(err).NotTo(HaveOccurred())

				container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
			})

			It("should use the correct container image", func() {
				Expect(container.Image).To(Equal("offen/docker-volume-backup:v2.43.0"))
			})

			It("should configure the S3 bucket and path", func() {
				Expect(container.Env).To(ContainElements(
					corev1.EnvVar{
						Name:  "AWS_S3_BUCKET_NAME",
						Value: "test-bucket",
					},
					corev1.EnvVar{
						Name:  "AWS_S3_PATH",
						Value: "backups/test",
					},
				))
			})

			It("should configure the S3 credentials", func() {
				Expect(container.Env).To(ContainElements(
					corev1.EnvVar{
						Name: "AWS_ACCESS_KEY_ID",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: s3Destination.Spec.S3.AccessKeyID,
						},
					},
					corev1.EnvVar{
						Name: "AWS_SECRET_ACCESS_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: s3Destination.Spec.S3.SecretAccessKey,
						},
					},
				))
			})

			It("should configure the S3 endpoint settings", func() {
				Expect(container.Env).To(ContainElements(
					corev1.EnvVar{
						Name:  "AWS_ENDPOINT",
						Value: "minio.example.com",
					},
					corev1.EnvVar{
						Name:  "AWS_ENDPOINT_PROTO",
						Value: "https",
					},
				))
			})

			It("should configure the S3 storage class", func() {
				Expect(container.Env).To(ContainElement(
					corev1.EnvVar{
						Name:  "AWS_STORAGE_CLASS",
						Value: "STANDARD",
					},
				))
			})

			Context("when using IAM role authentication", func() {
				BeforeEach(func() {
					s3Destination.Spec.S3.AccessKeyID = nil
					s3Destination.Spec.S3.SecretAccessKey = nil
					s3Destination.Spec.S3.IAMRoleEndpoint = "http://169.254.169.254"
					Expect(k8sClient.Update(ctx, s3Destination)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: s3BackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, s3BackupPlanName, cronJob)
					Expect(err).NotTo(HaveOccurred())

					container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
				})

				It("should configure the IAM role endpoint", func() {
					Expect(container.Env).To(ContainElement(
						corev1.EnvVar{
							Name:  "AWS_IAM_ROLE_ENDPOINT",
							Value: "http://169.254.169.254",
						},
					))
				})

				It("should not include AWS credentials", func() {
					Expect(container.Env).NotTo(ContainElement(HaveField("Name", "AWS_ACCESS_KEY_ID")))
					Expect(container.Env).NotTo(ContainElement(HaveField("Name", "AWS_SECRET_ACCESS_KEY")))
				})
			})

			Context("when using insecure endpoints", func() {
				BeforeEach(func() {
					s3Destination.Spec.S3.EndpointProtocol = "https"
					s3Destination.Spec.S3.EndpointInsecure = true
					s3Destination.Spec.S3.EndpointCACert = "-----BEGIN CERTIFICATE-----\nMIIE...\n-----END CERTIFICATE-----"
					Expect(k8sClient.Update(ctx, s3Destination)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: s3BackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, s3BackupPlanName, cronJob)
					Expect(err).NotTo(HaveOccurred())

					container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
				})

				It("should configure insecure endpoint settings only for HTTPS", func() {
					Expect(container.Env).To(ContainElements(
						corev1.EnvVar{
							Name:  "AWS_ENDPOINT_PROTO",
							Value: "https",
						},
						corev1.EnvVar{
							Name:  "AWS_ENDPOINT_INSECURE",
							Value: "true",
						},
						corev1.EnvVar{
							Name:  "AWS_ENDPOINT_CA_CERT",
							Value: "-----BEGIN CERTIFICATE-----\nMIIE...\n-----END CERTIFICATE-----",
						},
					))
				})
			})

			Context("when using HTTP protocol", func() {
				BeforeEach(func() {
					s3Destination.Spec.S3.EndpointProtocol = "http"
					s3Destination.Spec.S3.EndpointInsecure = true
					Expect(k8sClient.Update(ctx, s3Destination)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: s3BackupPlanName})
					Expect(err).NotTo(HaveOccurred())

					cronJob := &batchv1.CronJob{}
					err = k8sClient.Get(ctx, s3BackupPlanName, cronJob)
					Expect(err).NotTo(HaveOccurred())

					container = cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0]
				})

				It("should not set insecure flag for HTTP protocol", func() {
					Expect(container.Env).To(ContainElement(
						corev1.EnvVar{
							Name:  "AWS_ENDPOINT_PROTO",
							Value: "http",
						},
					))
					Expect(container.Env).NotTo(ContainElement(HaveField("Name", "AWS_ENDPOINT_INSECURE")))
				})
			})

			Context("when the server is deleted", func() {
				BeforeEach(func() {
					Expect(k8sClient.Delete(ctx, server)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: s3BackupPlanName})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should delete the CronJob", func() {
					cronJob := &batchv1.CronJob{}
					err := k8sClient.Get(ctx, s3BackupPlanName, cronJob)
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})

			Context("when the destination is deleted", func() {
				BeforeEach(func() {
					Expect(k8sClient.Delete(ctx, s3Destination)).To(Succeed())

					_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: s3BackupPlanName})
					Expect(err).NotTo(HaveOccurred())
				})

				It("should delete the CronJob", func() {
					cronJob := &batchv1.CronJob{}
					err := k8sClient.Get(ctx, s3BackupPlanName, cronJob)
					Expect(errors.IsNotFound(err)).To(BeTrue())
				})
			})
		})
	})
})
