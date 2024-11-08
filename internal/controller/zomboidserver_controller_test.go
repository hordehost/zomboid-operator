package controller

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Describe("ZomboidServer Controller", func() {
	var (
		ctx        context.Context
		reconciler *ZomboidServerReconciler
	)

	BeforeEach(func() {
		ctx = context.Background()
		reconciler = &ZomboidServerReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
	})

	It("should have the CRD available", func() {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Name: "zomboidservers.horde.host",
		}, crd)).To(Succeed())

		Expect(crd.Spec.Names.Kind).To(Equal("ZomboidServer"))
	})

	When("managing ZomboidServer resources", func() {
		var (
			zomboidServerName types.NamespacedName
			zomboidServer     *zomboidv1.ZomboidServer
		)

		It("should do nothing when the ZomboidServer isn't found", func() {
			nonExistentName := types.NamespacedName{
				Name:      "does-not-exist",
				Namespace: "anyhoo",
			}

			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: nonExistentName,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})

		BeforeEach(func() {
			namespace := corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace-" + uuid.New().String(),
				},
			}
			Expect(k8sClient.Create(ctx, &namespace)).To(Succeed())

			adminSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "the-admin-secret",
					Namespace: namespace.Name,
				},
				StringData: map[string]string{
					"password": "the-extremely-secure-password",
				},
			}
			Expect(k8sClient.Create(ctx, adminSecret)).To(Succeed())

			serverSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "server-secret",
					Namespace: namespace.Name,
				},
				StringData: map[string]string{
					"password": "server-password",
				},
			}
			Expect(k8sClient.Create(ctx, serverSecret)).To(Succeed())

			zomboidServerName = types.NamespacedName{
				Name:      "test-server",
				Namespace: namespace.Name,
			}

			zomboidServer = &zomboidv1.ZomboidServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      zomboidServerName.Name,
					Namespace: zomboidServerName.Namespace,
				},
				Spec: zomboidv1.ZomboidServerSpec{
					Version: "41.78.16",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("2Gi"),
							corev1.ResourceCPU:    resource.MustParse("500m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("2Gi"),
							corev1.ResourceCPU:    resource.MustParse("1"),
						},
					},
					Storage: zomboidv1.Storage{
						StorageClassName: ptr.To("standard"),
						Request:          resource.MustParse("10Gi"),
					},
					Administrator: zomboidv1.Administrator{
						Username: "admin",
						Password: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: adminSecret.Name,
							},
							Key: "password",
						},
					},
					Password: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: serverSecret.Name,
						},
						Key: "password",
					},
				},
			}

			reconciler = &ZomboidServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			Expect(k8sClient.Create(ctx, zomboidServer)).To(Succeed())

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: zomboidServerName})
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Get(ctx, zomboidServerName, zomboidServer)).To(Succeed())
		})

		When("Creating a new ZomboidServer", func() {
			When("Creating the PersistentVolumeClaim", func() {
				var pvc *corev1.PersistentVolumeClaim

				BeforeEach(func() {
					pvc = &corev1.PersistentVolumeClaim{}
					Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      zomboidServer.Name + "-game-data",
						Namespace: zomboidServer.Namespace,
					}, pvc)).To(Succeed())
				})

				It("should create PVC with correct storage class", func() {
					Expect(*pvc.Spec.StorageClassName).To(Equal("standard"))
				})

				It("should create PVC with correct storage size", func() {
					Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("10Gi")))
				})

				It("should create PVC with correct access mode", func() {
					Expect(pvc.Spec.AccessModes).To(ConsistOf(corev1.ReadWriteOnce))
				})

				It("should set the correct labels", func() {
					expectedLabels := map[string]string{
						"app.kubernetes.io/name":       "zomboidserver",
						"app.kubernetes.io/instance":   zomboidServer.Name,
						"app.kubernetes.io/managed-by": "zomboid-operator",
					}
					Expect(pvc.Labels).To(Equal(expectedLabels))
				})
			})

			When("creating the RCON Service", func() {
				var rconService *corev1.Service

				BeforeEach(func() {
					rconService = &corev1.Service{}
					Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      zomboidServer.Name + "-rcon",
						Namespace: zomboidServer.Namespace,
					}, rconService)).To(Succeed())
				})

				It("should create the RCON service with correct port", func() {
					Expect(rconService.Spec.Ports).To(ConsistOf(
						corev1.ServicePort{
							Name:       "rcon",
							Port:       27015,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromString("rcon"),
						},
					))
				})

				It("should set the correct selector", func() {
					Expect(rconService.Spec.Selector).To(Equal(map[string]string{
						"app.kubernetes.io/name":       "zomboidserver",
						"app.kubernetes.io/instance":   zomboidServer.Name,
						"app.kubernetes.io/managed-by": "zomboid-operator",
					}))
				})

				It("should set the correct labels", func() {
					expectedLabels := map[string]string{
						"app.kubernetes.io/name":       "zomboidserver",
						"app.kubernetes.io/instance":   zomboidServer.Name,
						"app.kubernetes.io/managed-by": "zomboid-operator",
					}
					Expect(rconService.Labels).To(Equal(expectedLabels))
					Expect(rconService.Spec.Selector).To(Equal(expectedLabels))
				})
			})

			When("creating the Deployment", func() {
				var (
					deployment *appsv1.Deployment
					container  corev1.Container
				)

				BeforeEach(func() {
					deployment = &appsv1.Deployment{}
					Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())

					Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
					container = deployment.Spec.Template.Spec.Containers[0]
					Expect(container.Name).To(Equal("zomboid"))
				})

				Context("init containers", func() {
					var initContainers []corev1.Container

					BeforeEach(func() {
						initContainers = deployment.Spec.Template.Spec.InitContainers
						Expect(initContainers).To(HaveLen(4))
					})

					It("should configure game-data init containers correctly", func() {
						setOwner := initContainers[0]
						Expect(setOwner.Name).To(Equal("game-data-set-owner"))
						Expect(setOwner.Image).To(Equal("hordehost/zomboid-server:" + zomboidServer.Spec.Version))
						Expect(setOwner.Command).To(Equal([]string{"/usr/bin/chown", "-R", "1000:1000", "/game-data"}))
						Expect(setOwner.SecurityContext.RunAsUser).To(Equal(ptr.To(int64(0))))
						Expect(setOwner.VolumeMounts).To(ConsistOf(corev1.VolumeMount{
							Name:      "game-data",
							MountPath: "/game-data",
						}))

						setPermissions := initContainers[1]
						Expect(setPermissions.Name).To(Equal("game-data-set-permissions"))
						Expect(setPermissions.Image).To(Equal("hordehost/zomboid-server:" + zomboidServer.Spec.Version))
						Expect(setPermissions.Command).To(Equal([]string{"/usr/bin/chmod", "-R", "755", "/game-data"}))
						Expect(setPermissions.SecurityContext.RunAsUser).To(Equal(ptr.To(int64(0))))
						Expect(setPermissions.VolumeMounts).To(ConsistOf(corev1.VolumeMount{
							Name:      "game-data",
							MountPath: "/game-data",
						}))
					})

					It("should configure workshop init containers correctly", func() {
						setOwner := initContainers[2]
						Expect(setOwner.Name).To(Equal("workshop-set-owner"))
						Expect(setOwner.Image).To(Equal("hordehost/zomboid-server:" + zomboidServer.Spec.Version))
						Expect(setOwner.Command).To(Equal([]string{"/usr/bin/chown", "-R", "1000:1000", "/server/steamapps"}))
						Expect(setOwner.SecurityContext.RunAsUser).To(Equal(ptr.To(int64(0))))
						Expect(setOwner.VolumeMounts).To(ConsistOf(corev1.VolumeMount{
							Name:      "workshop",
							MountPath: "/server/steamapps",
						}))

						setPermissions := initContainers[3]
						Expect(setPermissions.Name).To(Equal("workshop-set-permissions"))
						Expect(setPermissions.Image).To(Equal("hordehost/zomboid-server:" + zomboidServer.Spec.Version))
						Expect(setPermissions.Command).To(Equal([]string{"/usr/bin/chmod", "-R", "755", "/server/steamapps"}))
						Expect(setPermissions.SecurityContext.RunAsUser).To(Equal(ptr.To(int64(0))))
						Expect(setPermissions.VolumeMounts).To(ConsistOf(corev1.VolumeMount{
							Name:      "workshop",
							MountPath: "/server/steamapps",
						}))
					})
				})

				It("should mount both game-data and workshop volumes", func() {
					Expect(container.VolumeMounts).To(ConsistOf(
						corev1.VolumeMount{
							Name:      "game-data",
							MountPath: "/game-data",
						},
						corev1.VolumeMount{
							Name:      "workshop",
							MountPath: "/server/steamapps",
						},
					))
				})

				Context("workshop volume configuration", func() {
					It("should use emptyDir when WorkshopRequest is not specified", func() {
						volumes := deployment.Spec.Template.Spec.Volumes
						workshopVolume := volumes[1]
						Expect(workshopVolume.Name).To(Equal("workshop"))
						Expect(workshopVolume.EmptyDir).NotTo(BeNil())
						Expect(workshopVolume.PersistentVolumeClaim).To(BeNil())
					})

					When("WorkshopRequest is specified", func() {
						BeforeEach(func() {
							zomboidServer.Spec.Storage.WorkshopRequest = ptr.To(resource.MustParse("20Gi"))
							updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)
						})

						It("should create a PVC for workshop data", func() {
							workshopPVC := &corev1.PersistentVolumeClaim{}
							Expect(k8sClient.Get(ctx, types.NamespacedName{
								Name:      zomboidServer.Name + "-workshop",
								Namespace: zomboidServer.Namespace,
							}, workshopPVC)).To(Succeed())

							Expect(workshopPVC.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("20Gi")))
							Expect(*workshopPVC.Spec.StorageClassName).To(Equal("standard"))
							Expect(workshopPVC.Spec.AccessModes).To(ConsistOf(corev1.ReadWriteOnce))

							expectedLabels := map[string]string{
								"app.kubernetes.io/name":       "zomboidserver",
								"app.kubernetes.io/instance":   zomboidServer.Name,
								"app.kubernetes.io/managed-by": "zomboid-operator",
							}
							Expect(workshopPVC.Labels).To(Equal(expectedLabels))
						})

						It("should use the workshop PVC in the deployment", func() {
							deployment := &appsv1.Deployment{}
							Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())

							volumes := deployment.Spec.Template.Spec.Volumes
							workshopVolume := volumes[1]
							Expect(workshopVolume.Name).To(Equal("workshop"))
							Expect(workshopVolume.PersistentVolumeClaim).NotTo(BeNil())
							Expect(workshopVolume.PersistentVolumeClaim.ClaimName).To(Equal(zomboidServer.Name + "-workshop"))
							Expect(workshopVolume.EmptyDir).To(BeNil())
						})
					})
				})

				It("should set the correct container image", func() {
					Expect(container.Image).To(Equal("hordehost/zomboid-server:" + zomboidServer.Spec.Version))
				})

				It("should set the correct resource requirements and set the JVM max heap size", func() {
					Expect(container.Resources).To(Equal(zomboidServer.Spec.Resources))
					Expect(container.Env).To(ContainElement(corev1.EnvVar{
						Name:  "ZOMBOID_JVM_MAX_HEAP",
						Value: "2048m",
					}))
				})

				It("should set the server name", func() {
					Expect(container.Env).To(ContainElement(corev1.EnvVar{
						Name:  "ZOMBOID_SERVER_NAME",
						Value: zomboidServer.Name,
					}))
				})

				It("should set up the admin user", func() {
					Expect(container.Env).To(ContainElement(corev1.EnvVar{
						Name:  "ZOMBOID_SERVER_ADMIN_USERNAME",
						Value: "admin",
					}))
					Expect(container.Env).To(ContainElement(corev1.EnvVar{
						Name: "ZOMBOID_SERVER_ADMIN_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "the-admin-secret",
								},
								Key: "password",
							},
						},
					}))
				})

				It("should set a startup probe", func() {
					Expect(container.StartupProbe).To(Equal(&corev1.Probe{
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
					}))
				})

				It("should set a liveness probe", func() {
					Expect(container.LivenessProbe).To(Equal(&corev1.Probe{
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
					}))
				})

				It("should configure graceful shutdown via RCON", func() {
					Expect(container.Lifecycle).To(Equal(&corev1.Lifecycle{
						PreStop: &corev1.LifecycleHandler{
							Exec: &corev1.ExecAction{
								Command: []string{"/server/rcon", "quit"},
							},
						},
					}))
				})

				It("should conditionally set the server password", func() {
					// First verify password is set when Spec.Password is configured
					Expect(container.Env).To(ContainElement(corev1.EnvVar{
						Name: "ZOMBOID_SERVER_PASSWORD",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "server-secret",
								},
								Key: "password",
							},
						},
					}))

					// Now remove the password and verify it's not set
					zomboidServer.Spec.Password = nil
					updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)

					updatedDeployment := &appsv1.Deployment{}
					Expect(k8sClient.Get(ctx, zomboidServerName, updatedDeployment)).To(Succeed())

					updatedContainer := updatedDeployment.Spec.Template.Spec.Containers[0]
					for _, env := range updatedContainer.Env {
						Expect(env.Name).NotTo(Equal("ZOMBOID_SERVER_PASSWORD"))
					}
				})

				It("should set the admin password hash annotation", func() {
					Expect(deployment.Spec.Template.Annotations["secret/admin"]).To(
						Equal("a0052321048e12ed3bf3e2d264e41762a0547fceacfb97c04ed058c0edc39a8b"),
					)
				})

				When("server password is set", func() {
					It("should set both admin and server password hash annotations", func() {
						Expect(deployment.Spec.Template.Annotations["secret/server"]).To(
							Equal("32b7f0192280ba7f3529cf2cd5e381ab68db4a50acf636b7f32524364a1e98cc"),
						)
					})
				})

				When("server password is not set", func() {
					It("should only set admin password hash annotation", func() {
						zomboidServer.Spec.Password = nil
						updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)

						updatedDeployment := &appsv1.Deployment{}
						Expect(k8sClient.Get(ctx, zomboidServerName, updatedDeployment)).To(Succeed())

						Expect(updatedDeployment.Spec.Template.Annotations["secret/admin"]).NotTo(BeEmpty())
						Expect(updatedDeployment.Spec.Template.Annotations["secret/server"]).To(BeEmpty())
					})
				})

				It("should expose the RCON port", func() {
					Expect(container.Ports).To(ContainElement(corev1.ContainerPort{
						Name:          "rcon",
						ContainerPort: 27015,
						Protocol:      corev1.ProtocolTCP,
					}))
				})

				It("should set the correct labels", func() {
					expectedLabels := map[string]string{
						"app.kubernetes.io/name":       "zomboidserver",
						"app.kubernetes.io/instance":   zomboidServer.Name,
						"app.kubernetes.io/managed-by": "zomboid-operator",
					}
					Expect(deployment.Labels).To(Equal(expectedLabels))
					Expect(deployment.Spec.Selector.MatchLabels).To(Equal(expectedLabels))
					Expect(deployment.Spec.Template.Labels).To(Equal(expectedLabels))
				})

				It("should set replicas to 1 when not suspended", func() {
					Expect(*deployment.Spec.Replicas).To(Equal(int32(1)))
				})

				When("suspended is true", func() {
					BeforeEach(func() {
						zomboidServer.Spec.Suspended = ptr.To(true)
						updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)

						deployment = &appsv1.Deployment{}
						Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())
					})

					It("should set replicas to 0", func() {
						Expect(*deployment.Spec.Replicas).To(Equal(int32(0)))
					})
				})

				When("suspended is false", func() {
					BeforeEach(func() {
						zomboidServer.Spec.Suspended = ptr.To(false)
						updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)

						deployment = &appsv1.Deployment{}
						Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())
					})

					It("should set replicas to 1", func() {
						Expect(*deployment.Spec.Replicas).To(Equal(int32(1)))
					})
				})

				When("suspended is nil", func() {
					BeforeEach(func() {
						zomboidServer.Spec.Suspended = nil
						updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)

						deployment = &appsv1.Deployment{}
						Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())
					})

					It("should set replicas to 1", func() {
						Expect(*deployment.Spec.Replicas).To(Equal(int32(1)))
					})
				})
			})

			When("creating the Game Service", func() {
				var gameService *corev1.Service

				BeforeEach(func() {
					gameService = &corev1.Service{}
					Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      zomboidServer.Name,
						Namespace: zomboidServer.Namespace,
					}, gameService)).To(Succeed())
				})

				It("should create the Game service with correct ports", func() {
					Expect(gameService.Spec.Ports).To(ConsistOf(
						corev1.ServicePort{
							Name:       "steam",
							Port:       16261,
							Protocol:   corev1.ProtocolUDP,
							TargetPort: intstr.FromString("steam"),
						},
						corev1.ServicePort{
							Name:       "raknet",
							Port:       16262,
							Protocol:   corev1.ProtocolUDP,
							TargetPort: intstr.FromString("raknet"),
						},
					))
				})

				It("should set the correct selector", func() {
					expectedLabels := map[string]string{
						"app.kubernetes.io/name":       "zomboidserver",
						"app.kubernetes.io/instance":   zomboidServer.Name,
						"app.kubernetes.io/managed-by": "zomboid-operator",
					}
					Expect(gameService.Spec.Selector).To(Equal(expectedLabels))
				})

				It("should set the correct labels", func() {
					expectedLabels := map[string]string{
						"app.kubernetes.io/name":       "zomboidserver",
						"app.kubernetes.io/instance":   zomboidServer.Name,
						"app.kubernetes.io/managed-by": "zomboid-operator",
					}
					Expect(gameService.Labels).To(Equal(expectedLabels))
				})
			})

			Context("backup volume configuration", func() {
				When("BackupRequest is specified", func() {
					BeforeEach(func() {
						zomboidServer.Spec.Backups.Request = ptr.To(resource.MustParse("5Gi"))
						zomboidServer.Spec.Backups.StorageClassName = ptr.To("rwx-storage")
						updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)
					})

					It("should create a PVC for backups with RWX access mode", func() {
						backupPVC := &corev1.PersistentVolumeClaim{}
						Expect(k8sClient.Get(ctx, types.NamespacedName{
							Name:      zomboidServer.Name + "-backups",
							Namespace: zomboidServer.Namespace,
						}, backupPVC)).To(Succeed())

						Expect(backupPVC.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(resource.MustParse("5Gi")))
						Expect(*backupPVC.Spec.StorageClassName).To(Equal("rwx-storage"))
						Expect(backupPVC.Spec.AccessModes).To(ConsistOf(corev1.ReadWriteMany))

						expectedLabels := map[string]string{
							"app.kubernetes.io/name":       "zomboidserver",
							"app.kubernetes.io/instance":   zomboidServer.Name,
							"app.kubernetes.io/managed-by": "zomboid-operator",
						}
						Expect(backupPVC.Labels).To(Equal(expectedLabels))
					})

					It("should mount the backup PVC and add init containers", func() {
						deployment := &appsv1.Deployment{}
						Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())

						// Check init containers
						initContainers := deployment.Spec.Template.Spec.InitContainers
						Expect(initContainers).To(HaveLen(6)) // Original 4 + 2 new ones

						backupSetOwner := initContainers[4]
						Expect(backupSetOwner.Name).To(Equal("backup-set-owner"))
						Expect(backupSetOwner.Image).To(Equal("hordehost/zomboid-server:" + zomboidServer.Spec.Version))
						Expect(backupSetOwner.Command).To(Equal([]string{"/usr/bin/chown", "-R", "1000:1000", "/game-data/backups"}))
						Expect(backupSetOwner.SecurityContext.RunAsUser).To(Equal(ptr.To(int64(0))))
						Expect(backupSetOwner.VolumeMounts).To(ConsistOf(corev1.VolumeMount{
							Name:      "backups",
							MountPath: "/game-data/backups",
						}))

						backupSetPermissions := initContainers[5]
						Expect(backupSetPermissions.Name).To(Equal("backup-set-permissions"))
						Expect(backupSetPermissions.Image).To(Equal("hordehost/zomboid-server:" + zomboidServer.Spec.Version))
						Expect(backupSetPermissions.Command).To(Equal([]string{"/usr/bin/chmod", "-R", "755", "/game-data/backups"}))
						Expect(backupSetPermissions.SecurityContext.RunAsUser).To(Equal(ptr.To(int64(0))))
						Expect(backupSetPermissions.VolumeMounts).To(ConsistOf(corev1.VolumeMount{
							Name:      "backups",
							MountPath: "/game-data/backups",
						}))

						// Check container volume mounts
						container := deployment.Spec.Template.Spec.Containers[0]
						Expect(container.VolumeMounts).To(ContainElement(corev1.VolumeMount{
							Name:      "backups",
							MountPath: "/game-data/backups",
						}))

						// Check volumes
						volumes := deployment.Spec.Template.Spec.Volumes
						backupVolume := volumes[2]
						Expect(backupVolume.Name).To(Equal("backups"))
						Expect(backupVolume.PersistentVolumeClaim).NotTo(BeNil())
						Expect(backupVolume.PersistentVolumeClaim.ClaimName).To(Equal(zomboidServer.Name + "-backups"))
					})
				})

				When("BackupRequest is not specified", func() {
					It("should not create a backup PVC or mount", func() {
						deployment := &appsv1.Deployment{}
						Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())

						// Check that there are only the original init containers
						Expect(deployment.Spec.Template.Spec.InitContainers).To(HaveLen(4))

						// Check that there is no backup volume mount
						container := deployment.Spec.Template.Spec.Containers[0]
						for _, mount := range container.VolumeMounts {
							Expect(mount.Name).NotTo(Equal("backups"))
						}

						// Check that there is no backup volume
						for _, volume := range deployment.Spec.Template.Spec.Volumes {
							Expect(volume.Name).NotTo(Equal("backups"))
						}
					})
				})
			})
		})

		Context("Updating an existing ZomboidServer", func() {
			BeforeEach(func() {
				Expect(k8sClient.Get(ctx, zomboidServerName, zomboidServer)).To(Succeed())
			})

			It("should update Deployment when resources are changed", func() {
				zomboidServer.Spec.Resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")
				zomboidServer.Spec.Resources.Requests[corev1.ResourceMemory] = resource.MustParse("4Gi")

				updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)

				deployment := &appsv1.Deployment{}
				Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())

				container := deployment.Spec.Template.Spec.Containers[0]
				Expect(container.Resources.Limits[corev1.ResourceMemory]).To(Equal(resource.MustParse("4Gi")))
			})

			It("should update Deployment when version is changed", func() {
				zomboidServer.Spec.Version = "41.78.17"

				updateAndReconcile(ctx, k8sClient, reconciler, zomboidServer)

				deployment := &appsv1.Deployment{}
				Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())

				container := deployment.Spec.Template.Spec.Containers[0]
				Expect(container.Image).To(Equal("hordehost/zomboid-server:41.78.17"))
			})
		})
	})

	When("applying server settings", func() {
		It("should merge WorkshopMods into Mods strings", func() {
			settings := &zomboidv1.ZomboidSettings{
				// Start with some existing mods in the classic format
				Mods: zomboidv1.Mods{
					Mods:          ptr.To("ExistingMod1;ExistingMod2"),
					WorkshopItems: ptr.To("111111;222222"),
				},
				// Add some workshop mods in the structured format
				WorkshopMods: []zomboidv1.WorkshopMod{
					{
						ModID:      ptr.To("NewMod1"),
						WorkshopID: ptr.To("333333"),
					},
					{
						ModID:      ptr.To("NewMod2"),
						WorkshopID: ptr.To("444444"),
					},
				},
			}

			mergeWorkshopMods(settings)

			// Verify the mods were merged correctly
			Expect(*settings.Mods.Mods).To(Equal("ExistingMod1;ExistingMod2;NewMod1;NewMod2"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("111111;222222;333333;444444"))
		})

		It("should handle empty initial Mods strings", func() {
			settings := &zomboidv1.ZomboidSettings{
				WorkshopMods: []zomboidv1.WorkshopMod{
					{
						ModID:      ptr.To("NewMod1"),
						WorkshopID: ptr.To("333333"),
					},
				},
			}

			mergeWorkshopMods(settings)

			Expect(*settings.Mods.Mods).To(Equal("NewMod1"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("333333"))
		})

		It("should handle empty WorkshopMods", func() {
			settings := &zomboidv1.ZomboidSettings{
				Mods: zomboidv1.Mods{
					Mods:          ptr.To("ExistingMod1;ExistingMod2"),
					WorkshopItems: ptr.To("111111;222222"),
				},
			}

			mergeWorkshopMods(settings)

			// Verify the existing mods remain unchanged
			Expect(*settings.Mods.Mods).To(Equal("ExistingMod1;ExistingMod2"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("111111;222222"))
		})

		It("should handle nil ModID or WorkshopID", func() {
			settings := &zomboidv1.ZomboidSettings{
				WorkshopMods: []zomboidv1.WorkshopMod{
					{
						ModID:      ptr.To("NewMod1"),
						WorkshopID: nil,
					},
					{
						ModID:      nil,
						WorkshopID: ptr.To("444444"),
					},
				},
			}

			mergeWorkshopMods(settings)

			// Verify only non-nil values are merged
			Expect(*settings.Mods.Mods).To(Equal("NewMod1"))
			Expect(*settings.Mods.WorkshopItems).To(Equal("444444"))
		})
	})
})

func updateAndReconcile(ctx context.Context, k8sClient client.Client, reconciler *ZomboidServerReconciler, obj *zomboidv1.ZomboidServer) {
	Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}})
	Expect(err).NotTo(HaveOccurred())
}
