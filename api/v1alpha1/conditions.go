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

// CredentialsTokenKey is the key of the access token in the Secret that
// credentialsSecretRef refers to.
const CredentialsTokenKey = "token"

// Condition types.
const (
	// ConditionTypeReady summarises whether a PromotionPolicy is working.
	ConditionTypeReady = "Ready"
)

// Reasons of the Ready condition of a PromotionPolicy.
const (
	// ReasonSecretNotFound means that the credentials Secret does not exist.
	ReasonSecretNotFound = "SecretNotFound"

	// ReasonTokenMissing means that the credentials Secret has no token, or
	// the token is empty.
	ReasonTokenMissing = "TokenMissing"

	// ReasonRepositoryNotChecked means that the credentials are present, but
	// access to the repository has not been checked yet.
	ReasonRepositoryNotChecked = "RepositoryNotChecked"
)
