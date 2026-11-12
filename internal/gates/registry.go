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

package gates

// DefaultRegistry returns a Registry with every gate that promoloupe ships,
// in the order in which they appear in the status of a PromotionReview.
// A new gate is added here.
func DefaultRegistry() *Registry {
	return NewRegistry(
		NewMinSoakTime,
		NewFreezeWindows,
	)
}
