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

			result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: zomboidServerName})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

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

				It("should mount the game data volume", func() {
					Expect(container.VolumeMounts).To(ContainElement(corev1.VolumeMount{
						Name:      "game-data",
						MountPath: "/game-data",
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
						InitialDelaySeconds: 0,
						PeriodSeconds:       2,
						TimeoutSeconds:      1,
						SuccessThreshold:    1,
						FailureThreshold:    60,
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
	result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      obj.Name,
		Namespace: obj.Namespace,
	}})
	Expect(err).NotTo(HaveOccurred())
	Expect(result).To(Equal(ctrl.Result{}))
}
