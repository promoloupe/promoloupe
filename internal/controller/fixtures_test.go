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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
)

// testNamespace is the namespace where the tests create their resources.
const testNamespace = "default"

// validPromotionPolicy returns a PromotionPolicy that passes every validation rule.
func validPromotionPolicy(name string) *promoloupev1alpha1.PromotionPolicy {
	return &promoloupev1alpha1.PromotionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: promoloupev1alpha1.PromotionPolicySpec{
			Repository: promoloupev1alpha1.Repository{
				Provider:             promoloupev1alpha1.GitProviderGitLab,
				URL:                  "https://gitlab.com/acme/gitops",
				CredentialsSecretRef: promoloupev1alpha1.SecretRef{Name: "gitlab-token"},
			},
			Environments: []promoloupev1alpha1.Environment{
				{
					Name: "dev",
					Path: "apps/checkout/overlays/dev",
				},
				{
					Name:  "staging",
					Path:  "apps/checkout/overlays/staging",
					Merge: promoloupev1alpha1.MergeModeAutomatic,
					Gates: promoloupev1alpha1.Gates{MinSoakTimeMinutes: 60},
				},
				{
					Name: "prod",
					Path: "apps/checkout/overlays/prod",
					Gates: promoloupev1alpha1.Gates{
						MinSoakTimeMinutes: 1440,
						RequiredApprovals:  1,
					},
				},
			},
		},
	}
}

// validPromotionReview returns a PromotionReview that passes every validation rule.
func validPromotionReview(name string) *promoloupev1alpha1.PromotionReview {
	return &promoloupev1alpha1.PromotionReview{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: promoloupev1alpha1.PromotionReviewSpec{
			PolicyRef:      promoloupev1alpha1.PolicyRef{Name: "checkout"},
			From:           "staging",
			To:             "prod",
			SourceRevision: "9f2c1ab",
			Changes: []promoloupev1alpha1.Change{
				{
					Path:  "apps/checkout/overlays/prod/kustomization.yaml",
					Field: "images[name=checkout].newTag",
					From:  "1.4.2",
					To:    "1.5.0",
				},
			},
		},
	}
}
