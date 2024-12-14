package controller

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	zomboidv1 "github.com/zomboidhost/zomboid-operator/api/v1"
)

var _ = Describe("ZomboidServer Status Tests", func() {
	var (
		ctx        context.Context
		reconciler *ZomboidServerReconciler
	)

	Context("When managing a ZomboidServer", func() {
		var (
			zomboidServerName types.NamespacedName
			zomboidServer     *zomboidv1.ZomboidServer
		)

		BeforeEach(func() {
			ctx = context.Background()
			reconciler = &ZomboidServerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
		})

		BeforeEach(func() {
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-status-" + uuid.New().String(),
				},
			}
			Expect(k8sClient.Create(ctx, namespace)).Should(Succeed())

			zomboidServerName = types.NamespacedName{
				Name:      "test-server",
				Namespace: namespace.Name,
			}

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "admin-pass",
					Namespace: namespace.Name,
				},
				StringData: map[string]string{
					"password": "test123",
				},
			}
			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

			zomboidServer = &zomboidv1.ZomboidServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-server",
					Namespace: namespace.Name,
				},
				Spec: zomboidv1.ZomboidServerSpec{
					Version: "latest",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("2"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
					},
					Storage: zomboidv1.Storage{
						Request: resource.MustParse("10Gi"),
					},
					Administrator: zomboidv1.Administrator{
						Username: "admin",
						Password: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "admin-pass",
							},
							Key: "password",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, zomboidServer)).Should(Succeed())

			Expect(reconcileAndReload(ctx, reconciler, zomboidServerName, zomboidServer)).Should(Succeed())
		})

		It("Should update infrastructure readiness condition when PVC creation fails", func() {
			// Create a PVC with invalid storage request to trigger failure
			zomboidServer.Spec.Storage.Request.Set(-1)
			Expect(k8sClient.Update(ctx, zomboidServer)).Should(Succeed())

			Expect(reconcileAndReload(ctx, reconciler, zomboidServerName, zomboidServer)).To(HaveOccurred())

			infraCondition := meta.FindStatusCondition(zomboidServer.Status.Conditions, zomboidv1.TypeInfrastructureReady)
			Expect(infraCondition).NotTo(BeNil())
			Expect(infraCondition.Status).To(Equal(metav1.ConditionFalse))
			Expect(infraCondition.Reason).To(Equal(zomboidv1.ReasonMissingPVC))
			Expect(infraCondition.Message).To(ContainSubstring("Failed to reconcile PersistentVolumeClaim"))
		})

		It("Should update infrastructure readiness condition", func() {
			fmt.Println(zomboidServer.Status.Conditions)
			infraCondition := meta.FindStatusCondition(zomboidServer.Status.Conditions, zomboidv1.TypeInfrastructureReady)
			Expect(infraCondition).NotTo(BeNil())
			Expect(infraCondition.Status).To(Equal(metav1.ConditionTrue))
			Expect(infraCondition.Reason).To(Equal(zomboidv1.ReasonInfrastructureReady))
		})

		It("Should update infrastructure readiness condition when Deployment creation fails", func() {
			// Create an invalid container resource request to trigger failure
			zomboidServer.Spec.Resources.Requests[corev1.ResourceMemory] = resource.MustParse("-1")
			Expect(k8sClient.Update(ctx, zomboidServer)).Should(Succeed())

			Expect(reconcileAndReload(ctx, reconciler, zomboidServerName, zomboidServer)).To(HaveOccurred())

			infraCondition := meta.FindStatusCondition(zomboidServer.Status.Conditions, zomboidv1.TypeInfrastructureReady)
			Expect(infraCondition).NotTo(BeNil())
			Expect(infraCondition.Status).To(Equal(metav1.ConditionFalse))
			Expect(infraCondition.Reason).To(Equal(zomboidv1.ReasonMissingDeployment))
			Expect(infraCondition.Message).To(ContainSubstring("Failed to reconcile Deployment"))
		})

		It("Should update ready for players condition when deployment is not ready", func() {
			deployment := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: zomboidServer.Name, Namespace: zomboidServerName.Namespace}, deployment)).Should(Succeed())

			deployment.Status.Replicas = 1
			deployment.Status.ReadyReplicas = 0

			Expect(k8sClient.Status().Update(ctx, deployment)).Should(Succeed())

			Expect(reconcileAndReload(ctx, reconciler, zomboidServerName, zomboidServer)).Should(Succeed())

			Expect(zomboidServer.Status.Ready).To(BeFalse())

			readyCondition := meta.FindStatusCondition(zomboidServer.Status.Conditions, zomboidv1.TypeReadyForPlayers)
			Expect(readyCondition).NotTo(BeNil())
			Expect(readyCondition.Status).To(Equal(metav1.ConditionFalse))
			Expect(readyCondition.Reason).To(Equal(zomboidv1.ReasonServerStarting))
		})

		It("Should update ready for players condition when deployment is ready", func() {
			deployment := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: zomboidServer.Name, Namespace: zomboidServerName.Namespace}, deployment)).Should(Succeed())

			deployment.Status.Replicas = 1
			deployment.Status.ReadyReplicas = 1

			Expect(k8sClient.Status().Update(ctx, deployment)).Should(Succeed())

			Expect(reconcileAndReload(ctx, reconciler, zomboidServerName, zomboidServer)).Should(Succeed())

			Expect(zomboidServer.Status.Ready).To(BeTrue())

			readyCondition := meta.FindStatusCondition(zomboidServer.Status.Conditions, zomboidv1.TypeReadyForPlayers)
			Expect(readyCondition).NotTo(BeNil())
			Expect(readyCondition.Status).To(Equal(metav1.ConditionTrue))
			Expect(readyCondition.Reason).To(Equal(zomboidv1.ReasonServerReady))
		})
	})
})

func reconcileAndReload(
	ctx context.Context,
	reconciler *ZomboidServerReconciler,
	key types.NamespacedName,
	server *zomboidv1.ZomboidServer,
) error {
	result, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: key})
	Expect(result.Requeue).To(BeFalse())
	Expect(k8sClient.Get(ctx, key, server)).Should(Succeed())
	return err
}
