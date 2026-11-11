/*
Copyright 2026 The promoloupe Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
)

// createPolicy creates a valid PromotionPolicy that uses the Secret
// secretName for credentials, and deletes it when the spec ends.
func createPolicy(ctx context.Context, secretName string) *promoloupev1alpha1.PromotionPolicy {
	GinkgoHelper()
	policy := validPromotionPolicy()
	policy.GenerateName = "policy-"
	policy.Spec.Repository.CredentialsSecretRef.Name = secretName
	Expect(k8sClient.Create(ctx, policy)).To(Succeed())
	DeferCleanup(func(ctx context.Context) error {
		return client.IgnoreNotFound(k8sClient.Delete(ctx, policy))
	})
	return policy
}

// createSecret creates a Secret with data in the test namespace, and deletes
// it when the spec ends unless the spec deleted it first.
func createSecret(ctx context.Context, name string, data map[string][]byte) *corev1.Secret {
	GinkgoHelper()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: testNamespace},
		Data:       data,
	}
	Expect(k8sClient.Create(ctx, secret)).To(Succeed())
	DeferCleanup(func(ctx context.Context) error {
		return client.IgnoreNotFound(k8sClient.Delete(ctx, secret))
	})
	return secret
}

// getReadyCondition reads the PromotionPolicy from the API server and
// returns its Ready condition, or nil if it has none.
func getReadyCondition(
	ctx context.Context,
	g Gomega,
	policy *promoloupev1alpha1.PromotionPolicy,
) *metav1.Condition {
	current := &promoloupev1alpha1.PromotionPolicy{}
	g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(policy), current)).To(Succeed())
	return meta.FindStatusCondition(current.Status.Conditions, promoloupev1alpha1.ConditionTypeReady)
}

var _ = Describe("PromotionPolicy reconciler", func() {
	var (
		ctx        context.Context
		recorder   *events.FakeRecorder
		reconciler *PromotionPolicyReconciler
	)

	BeforeEach(func() {
		ctx = context.Background()
		recorder = events.NewFakeRecorder(10)
		reconciler = &PromotionPolicyReconciler{
			Client:    k8sClient,
			Scheme:    k8sClient.Scheme(),
			APIReader: k8sClient,
			Recorder:  recorder,
		}
	})

	reconcilePolicy := func(policy *promoloupev1alpha1.PromotionPolicy) {
		GinkgoHelper()
		_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(policy)})
		Expect(err).NotTo(HaveOccurred())
	}

	It("reports a missing Secret", func() {
		policy := createPolicy(ctx, "missing-secret")

		reconcilePolicy(policy)

		ready := getReadyCondition(ctx, Default, policy)
		Expect(ready).NotTo(BeNil())
		Expect(ready.Status).To(Equal(metav1.ConditionFalse))
		Expect(ready.Reason).To(Equal(promoloupev1alpha1.ReasonSecretNotFound))
		Expect(ready.ObservedGeneration).To(Equal(policy.Generation))
		Expect(recorder.Events).To(Receive(HavePrefix("Warning SecretNotFound")))
	})

	DescribeTable("reports a Secret without a usable token",
		func(secretName string, data map[string][]byte) {
			createSecret(ctx, secretName, data)
			policy := createPolicy(ctx, secretName)

			reconcilePolicy(policy)

			ready := getReadyCondition(ctx, Default, policy)
			Expect(ready).NotTo(BeNil())
			Expect(ready.Status).To(Equal(metav1.ConditionFalse))
			Expect(ready.Reason).To(Equal(promoloupev1alpha1.ReasonTokenMissing))
			Expect(recorder.Events).To(Receive(HavePrefix("Warning TokenMissing")))
		},
		Entry("without the token key", "no-token-key", map[string][]byte{"username": []byte("bot")}),
		Entry("with an empty token", "empty-token", map[string][]byte{
			promoloupev1alpha1.CredentialsTokenKey: {},
		}),
	)

	It("marks the repository as not checked when the credentials are present", func() {
		createSecret(ctx, "valid-credentials", map[string][]byte{
			promoloupev1alpha1.CredentialsTokenKey: []byte("test-token"),
		})
		policy := createPolicy(ctx, "valid-credentials")

		reconcilePolicy(policy)

		ready := getReadyCondition(ctx, Default, policy)
		Expect(ready).NotTo(BeNil())
		Expect(ready.Status).To(Equal(metav1.ConditionUnknown))
		Expect(ready.Reason).To(Equal(promoloupev1alpha1.ReasonRepositoryNotChecked))
		Expect(recorder.Events).To(Receive(HavePrefix("Normal RepositoryNotChecked")))
	})

	It("is idempotent", func() {
		policy := createPolicy(ctx, "idempotent-secret")
		reconcilePolicy(policy)
		first := getReadyCondition(ctx, Default, policy)
		Expect(recorder.Events).To(Receive())

		reconcilePolicy(policy)

		second := getReadyCondition(ctx, Default, policy)
		Expect(second.LastTransitionTime).To(Equal(first.LastTransitionTime))
		Expect(recorder.Events).NotTo(Receive())
	})

	It("ignores a PromotionPolicy that no longer exists", func() {
		_, err := reconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{Namespace: testNamespace, Name: "deleted-policy"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(recorder.Events).NotTo(Receive())
	})
})

var _ = Describe("PromotionPolicy controller", Ordered, func() {
	var (
		cancelManager  context.CancelFunc
		managerStopped chan struct{}
	)

	BeforeAll(func() {
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme.Scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
		})
		Expect(err).NotTo(HaveOccurred())

		Expect((&PromotionPolicyReconciler{
			Client:    mgr.GetClient(),
			Scheme:    mgr.GetScheme(),
			APIReader: mgr.GetAPIReader(),
			Recorder:  events.NewFakeRecorder(100),
		}).SetupWithManager(mgr)).To(Succeed())

		var managerCtx context.Context
		managerCtx, cancelManager = context.WithCancel(context.Background())
		managerStopped = make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(managerStopped)
			Expect(mgr.Start(managerCtx)).To(Succeed())
		}()
	})

	AfterAll(func() {
		cancelManager()
		Eventually(managerStopped).Should(BeClosed())
	})

	It("reconciles a PromotionPolicy when its Secret is created, changed or deleted", func() {
		ctx := context.Background()
		policy := createPolicy(ctx, "watched-secret")

		expectReady := func(status metav1.ConditionStatus, reason string) {
			GinkgoHelper()
			Eventually(func(g Gomega) {
				ready := getReadyCondition(ctx, g, policy)
				g.Expect(ready).NotTo(BeNil())
				g.Expect(ready.Status).To(Equal(status))
				g.Expect(ready.Reason).To(Equal(reason))
			}).Should(Succeed())
		}

		By("reporting the missing Secret")
		expectReady(metav1.ConditionFalse, promoloupev1alpha1.ReasonSecretNotFound)

		By("creating the Secret with a token")
		secret := createSecret(ctx, "watched-secret", map[string][]byte{
			promoloupev1alpha1.CredentialsTokenKey: []byte("test-token"),
		})
		expectReady(metav1.ConditionUnknown, promoloupev1alpha1.ReasonRepositoryNotChecked)

		By("removing the token from the Secret")
		secret.Data = map[string][]byte{"username": []byte("bot")}
		Expect(k8sClient.Update(ctx, secret)).To(Succeed())
		expectReady(metav1.ConditionFalse, promoloupev1alpha1.ReasonTokenMissing)

		By("deleting the Secret")
		Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
		expectReady(metav1.ConditionFalse, promoloupev1alpha1.ReasonSecretNotFound)
	})
})
