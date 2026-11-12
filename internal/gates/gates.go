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

// Package gates evaluates the conditions that a change must meet to be
// promoted into an environment. Each condition is a Gate. A Registry builds
// the gates that an environment configures, and Evaluate runs them.
package gates

import (
	"context"
	"time"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
)

// Promotion is what the gates know about a promotion when they evaluate it.
type Promotion struct {
	// Now is the time of the evaluation. Gates never read the clock
	// themselves, so the same promotion always gives the same outcome.
	Now time.Time

	// SourceDeployedAt is when the change reached the source environment.
	// It is zero while it is not known.
	SourceDeployedAt time.Time
}

// Outcome is the result of evaluating a gate.
type Outcome struct {
	// Result is Passed, Failed, or Unknown when the gate could not be
	// evaluated.
	Result promoloupev1alpha1.GateResult

	// Message explains the result to a person.
	Message string

	// RecheckAfter is how long until the result may change on its own, for
	// example when a soak time or a freeze window ends. It is zero when the
	// result only changes if something else changes.
	RecheckAfter time.Duration
}

// Gate is a condition that a change must meet to be promoted into an
// environment.
type Gate interface {
	// Name identifies the gate in the status of a PromotionReview.
	Name() promoloupev1alpha1.GateName

	// Evaluate checks the gate for a promotion. It never returns an error:
	// a gate that cannot be evaluated returns an Unknown outcome.
	Evaluate(ctx context.Context, promotion Promotion) Outcome
}

// Factory builds a gate from the gates configured for an environment. It
// returns false when the environment does not configure the gate.
type Factory func(config promoloupev1alpha1.Gates) (Gate, bool)

// Registry builds the gates that an environment configures, from a fixed
// list of factories.
type Registry struct {
	factories []Factory
}

// NewRegistry returns a Registry with the given factories. Gates are built
// and evaluated in the order of the factories.
func NewRegistry(factories ...Factory) *Registry {
	return &Registry{factories: factories}
}

// Build returns the gates that config configures.
func (r *Registry) Build(config promoloupev1alpha1.Gates) []Gate {
	built := make([]Gate, 0, len(r.factories))
	for _, factory := range r.factories {
		if gate, ok := factory(config); ok {
			built = append(built, gate)
		}
	}
	return built
}

// Evaluate evaluates every gate and returns their statuses, in the order of
// the gates, and the shortest RecheckAfter among them, or zero if none has
// one.
func Evaluate(
	ctx context.Context,
	toEvaluate []Gate,
	promotion Promotion,
) ([]promoloupev1alpha1.GateStatus, time.Duration) {
	statuses := make([]promoloupev1alpha1.GateStatus, 0, len(toEvaluate))
	var recheckAfter time.Duration

	for _, gate := range toEvaluate {
		outcome := gate.Evaluate(ctx, promotion)
		statuses = append(statuses, promoloupev1alpha1.GateStatus{
			Name:    gate.Name(),
			Result:  outcome.Result,
			Message: outcome.Message,
		})
		if outcome.RecheckAfter > 0 && (recheckAfter == 0 || outcome.RecheckAfter < recheckAfter) {
			recheckAfter = outcome.RecheckAfter
		}
	}

	return statuses, recheckAfter
}

// AllPassed reports whether every status passed. It is true when there are
// no statuses, because an environment without gates has nothing to wait for.
func AllPassed(statuses []promoloupev1alpha1.GateStatus) bool {
	for _, status := range statuses {
		if status.Result != promoloupev1alpha1.GateResultPassed {
			return false
		}
	}
	return true
}
