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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GitProvider is the Git hosting service of a repository.
// +kubebuilder:validation:Enum=GitLab
type GitProvider string

const (
	// GitProviderGitLab is GitLab.com or a self-managed GitLab instance.
	GitProviderGitLab GitProvider = "GitLab"
)

// MergeMode defines how a promotion into an environment is merged.
// +kubebuilder:validation:Enum=Manual;Automatic
type MergeMode string

const (
	// MergeModeManual leaves the merge to a person.
	MergeModeManual MergeMode = "Manual"
	// MergeModeAutomatic merges the promotion as soon as every gate passes.
	MergeModeAutomatic MergeMode = "Automatic"
)

// PromotableKind is a kind of value that travels between environments.
// +kubebuilder:validation:Enum=Images;HelmChartVersions
type PromotableKind string

const (
	// PromotableImages promotes container image tags.
	PromotableImages PromotableKind = "Images"
	// PromotableHelmChartVersions promotes Helm chart versions.
	PromotableHelmChartVersions PromotableKind = "HelmChartVersions"
)

// SecretRef refers to a Secret in the same namespace.
type SecretRef struct {
	// name of the Secret.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name,omitempty"`
}

// Repository is the GitOps repository that holds every environment.
type Repository struct {
	// provider is the Git hosting service of the repository.
	// +required
	Provider GitProvider `json:"provider,omitempty"`

	// url is the HTTPS URL of the repository.
	// +required
	// +kubebuilder:validation:MinLength=9
	// +kubebuilder:validation:MaxLength=2048
	// +kubebuilder:validation:XValidation:rule="self.startsWith('https://')",message="url must start with https://"
	URL string `json:"url,omitempty"`

	// branch is the branch that holds every environment.
	// +optional
	// +kubebuilder:default=main
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	Branch string `json:"branch,omitempty"`

	// credentialsSecretRef refers to a Secret with the access token under the key token.
	// +required
	CredentialsSecretRef SecretRef `json:"credentialsSecretRef,omitzero"`
}

// FreezeWindow is a recurring period in which promotions are blocked.
type FreezeWindow struct {
	// schedule is a cron expression, in UTC, for the start of the window.
	// +required
	// +kubebuilder:validation:MinLength=9
	// +kubebuilder:validation:MaxLength=100
	Schedule string `json:"schedule,omitempty"`

	// durationMinutes is the length of the window in minutes.
	// +required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=43200
	DurationMinutes int32 `json:"durationMinutes,omitempty"`
}

// ArgoCDGate requires an Argo CD Application of the previous environment to be Healthy and Synced.
type ArgoCDGate struct {
	// application is the name of the Argo CD Application.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Application string `json:"application,omitempty"`

	// namespace is the namespace of the Argo CD Application.
	// +optional
	// +kubebuilder:default=argocd
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	Namespace string `json:"namespace,omitempty"`
}

// Gates are the conditions a change must meet to be promoted into an environment.
// +kubebuilder:validation:MinProperties=1
type Gates struct {
	// minSoakTimeMinutes is the minimum time, in minutes, that the change must have run in the previous environment.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=43200
	MinSoakTimeMinutes int32 `json:"minSoakTimeMinutes,omitempty"`

	// requiredApprovals is the number of approvals required on the merge request.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	RequiredApprovals int32 `json:"requiredApprovals,omitempty"`

	// freezeWindows are the periods in which promotions into the environment are blocked.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	FreezeWindows []FreezeWindow `json:"freezeWindows,omitempty"`

	// argoCD requires the previous environment to be healthy in Argo CD.
	// +optional
	ArgoCD ArgoCDGate `json:"argoCD,omitzero"`
}

// Environment is one step of the promotion chain.
type Environment struct {
	// name identifies the environment. It is a DNS label, unique within the policy.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Name string `json:"name,omitempty"`

	// path is the directory of the environment in the repository.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Path string `json:"path,omitempty"`

	// merge defines how promotions into this environment are merged.
	// +optional
	// +kubebuilder:default=Manual
	Merge MergeMode `json:"merge,omitempty"`

	// gates are the conditions a change must meet to be promoted into this environment.
	// +optional
	Gates Gates `json:"gates,omitzero"`
}

// PromotionPolicySpec defines the desired state of PromotionPolicy.
type PromotionPolicySpec struct {
	// repository is the GitOps repository that holds every environment.
	// +required
	Repository Repository `json:"repository,omitzero"`

	// environments is the ordered chain of environments. Each one is promoted to the next.
	// +required
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=2
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:XValidation:rule="!has(self[0].gates)",message="the first environment cannot have gates, because nothing is promoted into it"
	// +kubebuilder:validation:XValidation:rule="!has(self[size(self) - 1].merge) || self[size(self) - 1].merge != 'Automatic'",message="promotions into the last environment must be merged manually"
	Environments []Environment `json:"environments,omitempty"`

	// promotable lists what travels between environments.
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=2
	// +kubebuilder:default={Images,HelmChartVersions}
	Promotable []PromotableKind `json:"promotable,omitempty"`
}

// EnvironmentStatus is the observed state of one environment.
type EnvironmentStatus struct {
	// name of the environment.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	Name string `json:"name,omitempty"`

	// revision is the commit of the repository last observed for the environment.
	// +optional
	// +kubebuilder:validation:MinLength=7
	// +kubebuilder:validation:MaxLength=64
	Revision string `json:"revision,omitempty"`

	// pendingReview is the name of the open PromotionReview from this environment to the next.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	PendingReview string `json:"pendingReview,omitempty"`
}

// PromotionPolicyStatus defines the observed state of PromotionPolicy.
// +kubebuilder:validation:MinProperties=1
type PromotionPolicyStatus struct {
	// conditions represent the current state of the PromotionPolicy.
	// Ready is True when the repository is reachable and every environment is tracked.
	// +listType=map
	// +listMapKey=type
	// +optional
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// observedGeneration is the generation of the spec that this status reflects.
	// +optional
	// +kubebuilder:validation:Minimum=1
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// environments reports the observed state of each environment.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=10
	Environments []EnvironmentStatus `json:"environments,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// PromotionPolicy describes a chain of environments in a GitOps repository and
// the rules a change must meet to be promoted from one environment to the next.
type PromotionPolicy struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PromotionPolicy
	// +required
	Spec PromotionPolicySpec `json:"spec,omitzero"`

	// status defines the observed state of PromotionPolicy
	// +optional
	Status PromotionPolicyStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PromotionPolicyList contains a list of PromotionPolicy
type PromotionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PromotionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PromotionPolicy{}, &PromotionPolicyList{})
		return nil
	})
}
