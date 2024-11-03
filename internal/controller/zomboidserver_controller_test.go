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
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	zomboidv1 "github.com/hordehost/zomboid-operator/api/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Describe("ZomboidServer Controller", func() {
	It("should have the CRD available", func() {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		err := k8sClient.Get(context.Background(), types.NamespacedName{
			Name: "zomboidservers.horde.host",
		}, crd)
		Expect(err).NotTo(HaveOccurred())
		Expect(crd.Spec.Names.Kind).To(Equal("ZomboidServer"))
	})

	Context("When managing ZomboidServer resources", func() {
		var (
			ctx            context.Context
			namespace      string
			zomboidServer  *zomboidv1.ZomboidServer
			namespacedName types.NamespacedName
			reconciler     *ZomboidServerReconciler
			adminSecret    *corev1.Secret
		)

		BeforeEach(func() {
			ctx = context.Background()
			namespace = fmt.Sprintf("zomboid-test-ns-%s", uuid.New().String())

			// Create namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			Expect(k8sClient.Create(ctx, ns)).To(Succeed())

			// Create admin secret
			adminSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-admin-secret",
					Namespace: namespace,
				},
				StringData: map[string]string{
					"password": "testpassword",
				},
			}
			Expect(k8sClient.Create(ctx, adminSecret)).To(Succeed())

			namespacedName = types.NamespacedName{
				Name:      "test-server",
				Namespace: namespace,
			}

			zomboidServer = &zomboidv1.ZomboidServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      namespacedName.Name,
					Namespace: namespacedName.Namespace,
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
				},
			}

			reconciler = &ZomboidServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
		})

		Context("Creating a new ZomboidServer", func() {
			It("should create PVC with correct specifications", func() {
				Expect(k8sClient.Create(ctx, zomboidServer)).To(Succeed())

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				pvc := &corev1.PersistentVolumeClaim{}
				Eventually(func() error {
					return k8sClient.Get(ctx, types.NamespacedName{
						Name:      zomboidServer.Name + "-game-data",
						Namespace: zomboidServer.Namespace,
					}, pvc)
				}).Should(Succeed())

				Expect(pvc.Spec.Resources.Requests[corev1.ResourceStorage]).To(Equal(zomboidServer.Spec.Storage.Request))
				Expect(pvc.Spec.AccessModes).To(ContainElement(corev1.ReadWriteOnce))
			})

			It("should create Deployment with correct specifications", func() {
				Expect(k8sClient.Create(ctx, zomboidServer)).To(Succeed())

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				deploy := &appsv1.Deployment{}
				Eventually(func() error {
					return k8sClient.Get(ctx, namespacedName, deploy)
				}).Should(Succeed())

				container := deploy.Spec.Template.Spec.Containers[0]
				Expect(container.Image).To(Equal("hordehost/zomboid-server:" + zomboidServer.Spec.Version))
				Expect(container.Resources).To(Equal(zomboidServer.Spec.Resources))

				// Verify volume mount
				Expect(container.VolumeMounts).To(ContainElement(corev1.VolumeMount{
					Name:      "game-data",
					MountPath: "/game-data",
				}))

				// Verify environment variables
				expectedEnvVars := []corev1.EnvVar{
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
				for _, envVar := range expectedEnvVars {
					Expect(container.Env).To(ContainElement(envVar))
				}
			})
		})

		Context("Updating an existing ZomboidServer", func() {
			BeforeEach(func() {
				Expect(k8sClient.Create(ctx, zomboidServer)).To(Succeed())
				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should update Deployment when resources are changed", func() {
				// Update the ZomboidServer resource requirements
				Eventually(func() error {
					if err := k8sClient.Get(ctx, namespacedName, zomboidServer); err != nil {
						return err
					}
					zomboidServer.Spec.Resources.Limits[corev1.ResourceMemory] = resource.MustParse("4Gi")
					zomboidServer.Spec.Resources.Requests[corev1.ResourceMemory] = resource.MustParse("4Gi")
					return k8sClient.Update(ctx, zomboidServer)
				}).Should(Succeed())

				// Reconcile the changes
				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				// Verify the deployment was updated
				deploy := &appsv1.Deployment{}
				Eventually(func() resource.Quantity {
					Expect(k8sClient.Get(ctx, namespacedName, deploy)).To(Succeed())
					return deploy.Spec.Template.Spec.Containers[0].Resources.Limits[corev1.ResourceMemory]
				}).Should(Equal(resource.MustParse("4Gi")))
			})

			It("should update Deployment when version is changed", func() {
				// Update the ZomboidServer version
				Eventually(func() error {
					if err := k8sClient.Get(ctx, namespacedName, zomboidServer); err != nil {
						return err
					}
					zomboidServer.Spec.Version = "41.78.17"
					return k8sClient.Update(ctx, zomboidServer)
				}).Should(Succeed())

				// Reconcile the changes
				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))

				// Verify the deployment was updated
				deploy := &appsv1.Deployment{}
				Eventually(func() string {
					Expect(k8sClient.Get(ctx, namespacedName, deploy)).To(Succeed())
					return deploy.Spec.Template.Spec.Containers[0].Image
				}).Should(Equal("hordehost/zomboid-server:41.78.17"))
			})

		})
	})
})
