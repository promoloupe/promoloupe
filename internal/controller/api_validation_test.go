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
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
)

// expectInvalid asserts that err is an Invalid error from the API server that contains message.
func expectInvalid(err error, message string) {
	GinkgoHelper()
	Expect(err).To(HaveOccurred())
	Expect(apierrors.IsInvalid(err)).To(BeTrue(), "expected an Invalid error, got: %v", err)
	Expect(err.Error()).To(ContainSubstring(message))
}

var _ = Describe("API validation", func() {
	ctx := context.Background()

	Context("PromotionPolicy", func() {
		It("accepts a valid policy and applies the defaults", func() {
			policy := validPromotionPolicy()
			policy.GenerateName = "valid-"
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			DeferCleanup(k8sClient.Delete, ctx, policy)

			Expect(policy.Spec.Repository.Branch).To(Equal("main"))
			Expect(policy.Spec.Environments[0].Merge).To(Equal(promoloupev1alpha1.MergeModeManual))
			Expect(policy.Spec.Promotable).To(ConsistOf(
				promoloupev1alpha1.PromotableImages,
				promoloupev1alpha1.PromotableHelmChartVersions,
			))
		})

		It("defaults the Argo CD namespace", func() {
			policy := validPromotionPolicy()
			policy.GenerateName = "argocd-"
			policy.Spec.Environments[2].Gates.ArgoCD = promoloupev1alpha1.ArgoCDGate{Application: "checkout-staging"}
			Expect(k8sClient.Create(ctx, policy)).To(Succeed())
			DeferCleanup(k8sClient.Delete, ctx, policy)

			Expect(policy.Spec.Environments[2].Gates.ArgoCD.Namespace).To(Equal("argocd"))
		})

		DescribeTable("rejects an invalid policy",
			func(mutate func(*promoloupev1alpha1.PromotionPolicy), message string) {
				policy := validPromotionPolicy()
				policy.GenerateName = "invalid-"
				mutate(policy)
				expectInvalid(k8sClient.Create(ctx, policy), message)
			},
			Entry("with a single environment",
				func(p *promoloupev1alpha1.PromotionPolicy) { p.Spec.Environments = p.Spec.Environments[:1] },
				"should have at least 2 items"),
			Entry("with gates on the first environment",
				func(p *promoloupev1alpha1.PromotionPolicy) {
					p.Spec.Environments[0].Gates = promoloupev1alpha1.Gates{MinSoakTimeMinutes: 10}
				},
				"the first environment cannot have gates"),
			Entry("with automatic merges into the last environment",
				func(p *promoloupev1alpha1.PromotionPolicy) {
					p.Spec.Environments[2].Merge = promoloupev1alpha1.MergeModeAutomatic
				},
				"promotions into the last environment must be merged manually"),
			Entry("with duplicate environment names",
				func(p *promoloupev1alpha1.PromotionPolicy) { p.Spec.Environments[1].Name = "dev" },
				"Duplicate value"),
			Entry("with an environment name that is not a DNS label",
				func(p *promoloupev1alpha1.PromotionPolicy) { p.Spec.Environments[0].Name = "Dev_1" },
				"should match"),
			Entry("with a repository URL that is not HTTPS",
				func(p *promoloupev1alpha1.PromotionPolicy) { p.Spec.Repository.URL = "http://gitlab.com/acme/gitops" },
				"url must start with https://"),
			Entry("with an unsupported provider",
				func(p *promoloupev1alpha1.PromotionPolicy) { p.Spec.Repository.Provider = "GitHub" },
				"Unsupported value"),
			Entry("with too many required approvals",
				func(p *promoloupev1alpha1.PromotionPolicy) { p.Spec.Environments[2].Gates.RequiredApprovals = 11 },
				"should be less than or equal to 10"),
		)
	})

	Context("PromotionReview", func() {
		It("accepts a valid review and allows status updates", func() {
			review := validPromotionReview("")
			review.GenerateName = "valid-"
			Expect(k8sClient.Create(ctx, review)).To(Succeed())
			DeferCleanup(k8sClient.Delete, ctx, review)

			review.Status.MergeRequest = promoloupev1alpha1.MergeRequest{
				URL: "https://gitlab.com/acme/gitops/-/merge_requests/42",
			}
			Expect(k8sClient.Status().Update(ctx, review)).To(Succeed())
		})

		It("rejects changes to the spec", func() {
			review := validPromotionReview("")
			review.GenerateName = "immutable-"
			Expect(k8sClient.Create(ctx, review)).To(Succeed())
			DeferCleanup(k8sClient.Delete, ctx, review)

			review.Spec.SourceRevision = "1234abc"
			expectInvalid(k8sClient.Update(ctx, review), "spec is immutable")
		})

		DescribeTable("rejects an invalid review",
			func(mutate func(*promoloupev1alpha1.PromotionReview), message string) {
				review := validPromotionReview("")
				review.GenerateName = "invalid-"
				mutate(review)
				expectInvalid(k8sClient.Create(ctx, review), message)
			},
			Entry("with the same source and target environment",
				func(r *promoloupev1alpha1.PromotionReview) { r.Spec.To = r.Spec.From },
				"from and to must be different environments"),
			Entry("with a source revision that is not a lowercase hex commit",
				func(r *promoloupev1alpha1.PromotionReview) { r.Spec.SourceRevision = "NOTASHA1" },
				"should match"),
			Entry("without changes",
				func(r *promoloupev1alpha1.PromotionReview) { r.Spec.Changes = nil },
				"Required value"),
		)
	})
})
