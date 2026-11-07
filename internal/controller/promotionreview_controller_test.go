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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
)

var _ = Describe("PromotionReview Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-resource"
			resourceNamespace = "default"
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}
		promotionreview := &promoloupev1alpha1.PromotionReview{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind PromotionReview")
			err := k8sClient.Get(ctx, typeNamespacedName, promotionreview)
			if err != nil && errors.IsNotFound(err) {
				resource := validPromotionReview(resourceName)
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &promoloupev1alpha1.PromotionReview{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance PromotionReview")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &PromotionReviewReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
