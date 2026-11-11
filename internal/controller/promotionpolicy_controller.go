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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
)

// credentialsSecretIndex indexes PromotionPolicies by the name of their
// credentials Secret, to find the policies that use a Secret when it changes.
const credentialsSecretIndex = ".spec.repository.credentialsSecretRef.name"

// PromotionPolicyReconciler reconciles a PromotionPolicy object.
type PromotionPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// APIReader reads from the API server directly. It is used for Secrets,
	// so that their content is never kept in the cache.
	APIReader client.Reader

	// Recorder emits Kubernetes events about PromotionPolicies.
	Recorder events.EventRecorder
}

// +kubebuilder:rbac:groups=promoloupe.com,resources=promotionpolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=promoloupe.com,resources=promotionpolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=events.k8s.io,resources=events,verbs=create;patch

// Reconcile checks the credentials of a PromotionPolicy and reports the
// result in its Ready condition.
func (r *PromotionPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	policy := &promoloupev1alpha1.PromotionPolicy{}
	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ready, err := r.readyCondition(ctx, policy)
	if err != nil {
		return ctrl.Result{}, err
	}

	var previousReason string
	if previous := meta.FindStatusCondition(policy.Status.Conditions, promoloupev1alpha1.ConditionTypeReady); previous != nil {
		previousReason = previous.Reason
	}

	base := policy.DeepCopy()
	meta.SetStatusCondition(&policy.Status.Conditions, ready)
	policy.Status.ObservedGeneration = policy.Generation

	if !equality.Semantic.DeepEqual(base.Status, policy.Status) {
		if err := r.Status().Patch(ctx, policy, client.MergeFrom(base)); err != nil {
			return ctrl.Result{}, fmt.Errorf("updating status: %w", err)
		}
	}

	if ready.Reason != previousReason {
		eventType := corev1.EventTypeNormal
		if ready.Status == metav1.ConditionFalse {
			eventType = corev1.EventTypeWarning
		}
		r.Recorder.Eventf(policy, nil, eventType, ready.Reason, "CheckCredentials", "%s", ready.Message)
		log.Info("Ready condition changed", "status", ready.Status, "reason", ready.Reason)
	}

	return ctrl.Result{}, nil
}

// readyCondition builds the Ready condition from the credentials Secret.
// A missing Secret or token is a user error, reported in the condition. Any
// other error is returned, so that the request is retried with backoff.
func (r *PromotionPolicyReconciler) readyCondition(
	ctx context.Context,
	policy *promoloupev1alpha1.PromotionPolicy,
) (metav1.Condition, error) {
	condition := metav1.Condition{
		Type:               promoloupev1alpha1.ConditionTypeReady,
		ObservedGeneration: policy.Generation,
	}

	name := policy.Spec.Repository.CredentialsSecretRef.Name
	secret := &corev1.Secret{}
	err := r.APIReader.Get(ctx, types.NamespacedName{Namespace: policy.Namespace, Name: name}, secret)

	switch {
	case apierrors.IsNotFound(err):
		condition.Status = metav1.ConditionFalse
		condition.Reason = promoloupev1alpha1.ReasonSecretNotFound
		condition.Message = fmt.Sprintf("Secret %q does not exist in namespace %q.", name, policy.Namespace)
	case err != nil:
		return metav1.Condition{}, fmt.Errorf("getting credentials Secret %q: %w", name, err)
	case len(secret.Data[promoloupev1alpha1.CredentialsTokenKey]) == 0:
		condition.Status = metav1.ConditionFalse
		condition.Reason = promoloupev1alpha1.ReasonTokenMissing
		condition.Message = fmt.Sprintf("Secret %q has no %q key, or its value is empty.",
			name, promoloupev1alpha1.CredentialsTokenKey)
	default:
		condition.Status = metav1.ConditionUnknown
		condition.Reason = promoloupev1alpha1.ReasonRepositoryNotChecked
		condition.Message = "The credentials are present. Access to the repository is not checked yet."
	}

	return condition, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PromotionPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(context.Background(), &promoloupev1alpha1.PromotionPolicy{},
		credentialsSecretIndex, func(obj client.Object) []string {
			policy, ok := obj.(*promoloupev1alpha1.PromotionPolicy)
			if !ok {
				return nil
			}
			return []string{policy.Spec.Repository.CredentialsSecretRef.Name}
		})
	if err != nil {
		return fmt.Errorf("indexing PromotionPolicies by credentials Secret: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&promoloupev1alpha1.PromotionPolicy{}).
		WatchesMetadata(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(r.policiesForSecret)).
		Named("promotionpolicy").
		Complete(r)
}

// policiesForSecret returns a request for every PromotionPolicy in the
// namespace of the Secret that uses it for credentials.
func (r *PromotionPolicyReconciler) policiesForSecret(ctx context.Context, secret client.Object) []reconcile.Request {
	policies := &promoloupev1alpha1.PromotionPolicyList{}
	if err := r.List(ctx, policies,
		client.InNamespace(secret.GetNamespace()),
		client.MatchingFields{credentialsSecretIndex: secret.GetName()},
	); err != nil {
		logf.FromContext(ctx).Error(err, "Failed to list PromotionPolicies for Secret",
			"secret", client.ObjectKeyFromObject(secret))
		return nil
	}

	requests := make([]reconcile.Request, 0, len(policies.Items))
	for i := range policies.Items {
		requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&policies.Items[i])})
	}
	return requests
}
