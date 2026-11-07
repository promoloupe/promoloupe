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

// GateName identifies a gate.
// +kubebuilder:validation:Enum=MinSoakTime;RequiredApprovals;FreezeWindows;ArgoCD
type GateName string

const (
	// GateMinSoakTime checks the time the change has run in the previous environment.
	GateMinSoakTime GateName = "MinSoakTime"
	// GateRequiredApprovals checks the approvals on the merge request.
	GateRequiredApprovals GateName = "RequiredApprovals"
	// GateFreezeWindows checks that no freeze window is active.
	GateFreezeWindows GateName = "FreezeWindows"
	// GateArgoCD checks the health of the previous environment in Argo CD.
	GateArgoCD GateName = "ArgoCD"
)

// GateResult is the outcome of evaluating a gate.
// +kubebuilder:validation:Enum=Passed;Failed;Unknown
type GateResult string

const (
	// GateResultPassed means the gate is met.
	GateResultPassed GateResult = "Passed"
	// GateResultFailed means the gate is not met.
	GateResultFailed GateResult = "Failed"
	// GateResultUnknown means the gate could not be evaluated, for example because Argo CD did not respond.
	GateResultUnknown GateResult = "Unknown"
)

// Recommendation is the reviewer's advice on a promotion.
// +kubebuilder:validation:Enum=Approve;Hold;Block
type Recommendation string

const (
	// RecommendationApprove advises merging the promotion.
	RecommendationApprove Recommendation = "Approve"
	// RecommendationHold advises waiting, for example until a metric recovers.
	RecommendationHold Recommendation = "Hold"
	// RecommendationBlock advises not merging the promotion.
	RecommendationBlock Recommendation = "Block"
)

// Risk is the reviewer's assessment of how risky a promotion is.
// +kubebuilder:validation:Enum=Low;Medium;High
type Risk string

const (
	// RiskLow is a routine change.
	RiskLow Risk = "Low"
	// RiskMedium is a change that deserves attention.
	RiskMedium Risk = "Medium"
	// RiskHigh is a change likely to cause an incident.
	RiskHigh Risk = "High"
)

// PolicyRef refers to a PromotionPolicy in the same namespace.
type PolicyRef struct {
	// name of the PromotionPolicy.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name,omitempty"`
}

// Change is one promotable value that changes in the target environment.
type Change struct {
	// path is the file that changes in the target environment.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Path string `json:"path,omitempty"`

	// field is the promotable field within the file, for example images[name=checkout].newTag.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=512
	Field string `json:"field,omitempty"`

	// from is the current value in the target environment. It is omitted when the field does not exist yet.
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=512
	From string `json:"from,omitempty"`

	// to is the value promoted from the source environment.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=512
	To string `json:"to,omitempty"`
}

// PromotionReviewSpec defines the desired state of PromotionReview.
// +kubebuilder:validation:XValidation:rule="self.from != self.to",message="from and to must be different environments"
type PromotionReviewSpec struct {
	// policyRef refers to the PromotionPolicy that the promotion belongs to.
	// +required
	PolicyRef PolicyRef `json:"policyRef,omitzero"`

	// from is the source environment.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	From string `json:"from,omitempty"`

	// to is the target environment, the next one in the chain.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	To string `json:"to,omitempty"`

	// sourceRevision is the commit of the source environment being promoted.
	// +required
	// +kubebuilder:validation:MinLength=7
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=`^[0-9a-f]+$`
	SourceRevision string `json:"sourceRevision,omitempty"`

	// changes are the promotable values that change in the target environment.
	// +required
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	Changes []Change `json:"changes,omitempty"`
}

// MergeRequest is the merge request opened for a promotion.
type MergeRequest struct {
	// url of the merge request.
	// +required
	// +kubebuilder:validation:MinLength=9
	// +kubebuilder:validation:MaxLength=2048
	URL string `json:"url,omitempty"`
}

// GateStatus is the result of evaluating one gate.
type GateStatus struct {
	// name of the gate.
	// +required
	Name GateName `json:"name,omitempty"`

	// result of the evaluation.
	// +required
	Result GateResult `json:"result,omitempty"`

	// message explains the result, for example "0 of 1 approvals".
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Message string `json:"message,omitempty"`
}

// Evidence is one fact behind a verdict.
type Evidence struct {
	// source is where the fact comes from, for example Prometheus or Argo CD.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	Source string `json:"source,omitempty"`

	// detail describes the fact.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Detail string `json:"detail,omitempty"`
}

// Verdict is the reviewer's assessment of a promotion. It only recommends and never promotes.
type Verdict struct {
	// recommendation is the reviewer's advice.
	// +required
	Recommendation Recommendation `json:"recommendation,omitempty"`

	// risk is how risky the reviewer considers the promotion.
	// +required
	Risk Risk `json:"risk,omitempty"`

	// summary explains the change and its risk in plain language.
	// +required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Summary string `json:"summary,omitempty"`

	// evidence lists the facts behind the verdict.
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=20
	Evidence []Evidence `json:"evidence,omitempty"`
}

// PromotionReviewStatus defines the observed state of PromotionReview.
// The controller owns mergeRequest, gates and the GatesPassed and Merged conditions.
// The reviewer owns verdict and the Reviewed condition.
// +kubebuilder:validation:MinProperties=1
type PromotionReviewStatus struct {
	// conditions represent the current state of the PromotionReview.
	// GatesPassed and Merged are set by the controller, Reviewed by the reviewer.
	// +listType=map
	// +listMapKey=type
	// +optional
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// observedGeneration is the generation of the spec that this status reflects.
	// +optional
	// +kubebuilder:validation:Minimum=1
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// mergeRequest is the merge request opened for the promotion.
	// +optional
	MergeRequest MergeRequest `json:"mergeRequest,omitzero"`

	// gates reports the result of each gate of the target environment.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=4
	Gates []GateStatus `json:"gates,omitempty"`

	// verdict is the reviewer's assessment of the promotion.
	// +optional
	Verdict Verdict `json:"verdict,omitzero"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="From",type=string,JSONPath=`.spec.from`
// +kubebuilder:printcolumn:name="To",type=string,JSONPath=`.spec.to`
// +kubebuilder:printcolumn:name="Recommendation",type=string,JSONPath=`.status.verdict.recommendation`
// +kubebuilder:printcolumn:name="Merged",type=string,JSONPath=`.status.conditions[?(@.type=="Merged")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// PromotionReview records one promotion between two adjacent environments.
// It is created by the controller and is the contract between the controller and the reviewer.
type PromotionReview struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec describes the promotion. It is immutable.
	// +required
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec is immutable"
	Spec PromotionReviewSpec `json:"spec,omitzero"`

	// status defines the observed state of PromotionReview
	// +optional
	Status PromotionReviewStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PromotionReviewList contains a list of PromotionReview
type PromotionReviewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PromotionReview `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PromotionReview{}, &PromotionReviewList{})
		return nil
	})
}
