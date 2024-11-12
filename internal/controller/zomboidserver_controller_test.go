package controller

import (
	"context"
	"fmt"

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

			When("creating the SQLite Service", func() {
				var sqliteService *corev1.Service

				BeforeEach(func() {
					sqliteService = &corev1.Service{}
					Expect(k8sClient.Get(ctx, types.NamespacedName{
						Name:      zomboidServer.Name + "-sqlite",
						Namespace: zomboidServer.Namespace,
					}, sqliteService)).To(Succeed())
				})

				It("should create the SQLite service with correct port", func() {
					Expect(sqliteService.Spec.Ports).To(ConsistOf(
						corev1.ServicePort{
							Name:       "ws4sqlite",
							Port:       12321,
							Protocol:   corev1.ProtocolTCP,
							TargetPort: intstr.FromString("ws4sqlite"),
						},
					))
				})

				It("should set the correct selector", func() {
					expectedLabels := map[string]string{
						"app.kubernetes.io/name":       "zomboidserver",
						"app.kubernetes.io/instance":   zomboidServer.Name,
						"app.kubernetes.io/managed-by": "zomboid-operator",
					}
					Expect(sqliteService.Spec.Selector).To(Equal(expectedLabels))
				})

				It("should set the correct labels", func() {
					expectedLabels := map[string]string{
						"app.kubernetes.io/name":       "zomboidserver",
						"app.kubernetes.io/instance":   zomboidServer.Name,
						"app.kubernetes.io/managed-by": "zomboid-operator",
					}
					Expect(sqliteService.Labels).To(Equal(expectedLabels))
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

			Context("ws4sqlite sidecar container", func() {
				var container corev1.Container

				BeforeEach(func() {
					deployment := &appsv1.Deployment{}
					Expect(k8sClient.Get(ctx, zomboidServerName, deployment)).To(Succeed())

					Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(2))

					container = deployment.Spec.Template.Spec.Containers[1]
					Expect(container.Name).To(Equal("ws4sqlite"))
				})

				It("should use the correct image", func() {
					Expect(container.Image).To(Equal("germanorizzo/ws4sqlite:v0.16.2"))
				})

				It("should set the correct command", func() {
					expectedDBPath := fmt.Sprintf("/game-data/db/%s.db", zomboidServer.Name)
					Expect(container.Args).To(Equal([]string{"--db", expectedDBPath}))
				})

				It("should mount the game-data volume", func() {
					Expect(container.VolumeMounts).To(ContainElement(corev1.VolumeMount{
						Name:      "game-data",
						MountPath: "/game-data",
					}))
				})

				It("should expose the HTTP port", func() {
					Expect(container.Ports).To(ContainElement(corev1.ContainerPort{
						Name:          "ws4sqlite",
						ContainerPort: 12321,
						Protocol:      corev1.ProtocolTCP,
					}))
				})

				It("should run as the correct user and group", func() {
					Expect(container.SecurityContext.RunAsUser).To(Equal(ptr.To(int64(1000))))
					Expect(container.SecurityContext.RunAsGroup).To(Equal(ptr.To(int64(1000))))
				})

				Context("service", func() {
					var sqliteService *corev1.Service

					BeforeEach(func() {
						sqliteService = &corev1.Service{}
						Expect(k8sClient.Get(ctx, types.NamespacedName{
							Name:      zomboidServer.Name + "-sqlite",
							Namespace: zomboidServer.Namespace,
						}, sqliteService)).To(Succeed())
					})

					It("should create the service with correct port", func() {
						Expect(sqliteService.Spec.Ports).To(ConsistOf(
							corev1.ServicePort{
								Name:       "ws4sqlite",
								Port:       12321,
								Protocol:   corev1.ProtocolTCP,
								TargetPort: intstr.FromString("ws4sqlite"),
							},
						))
					})

					It("should set the correct selector", func() {
						expectedLabels := map[string]string{
							"app.kubernetes.io/name":       "zomboidserver",
							"app.kubernetes.io/instance":   zomboidServer.Name,
							"app.kubernetes.io/managed-by": "zomboid-operator",
						}
						Expect(sqliteService.Spec.Selector).To(Equal(expectedLabels))
					})

					It("should set the correct labels", func() {
						expectedLabels := map[string]string{
							"app.kubernetes.io/name":       "zomboidserver",
							"app.kubernetes.io/instance":   zomboidServer.Name,
							"app.kubernetes.io/managed-by": "zomboid-operator",
						}
						Expect(sqliteService.Labels).To(Equal(expectedLabels))
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
})

func updateAndReconcile(ctx context.Context, k8sClient client.Client, reconciler *ZomboidServerReconciler, obj *zomboidv1.ZomboidServer) {
	Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}})
	Expect(err).NotTo(HaveOccurred())
}
