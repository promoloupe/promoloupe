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

import (
	"context"
	"fmt"
	"time"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
)

// minSoakTime requires the change to have run in the source environment for
// a minimum time.
type minSoakTime struct {
	minimum time.Duration
}

// NewMinSoakTime is the Factory of the MinSoakTime gate. It builds the gate
// when the environment sets minSoakTimeMinutes.
func NewMinSoakTime(config promoloupev1alpha1.Gates) (Gate, bool) {
	if config.MinSoakTimeMinutes == 0 {
		return nil, false
	}
	return minSoakTime{minimum: time.Duration(config.MinSoakTimeMinutes) * time.Minute}, true
}

func (minSoakTime) Name() promoloupev1alpha1.GateName {
	return promoloupev1alpha1.GateMinSoakTime
}

func (g minSoakTime) Evaluate(_ context.Context, promotion Promotion) Outcome {
	if promotion.SourceDeployedAt.IsZero() {
		return Outcome{
			Result:  promoloupev1alpha1.GateResultUnknown,
			Message: "The time the change reached the source environment is not known yet.",
		}
	}

	soaked := max(promotion.Now.Sub(promotion.SourceDeployedAt), 0)
	if soaked >= g.minimum {
		return Outcome{
			Result: promoloupev1alpha1.GateResultPassed,
			Message: fmt.Sprintf("The change has run for %d minutes, and %d are required.",
				wholeMinutes(soaked), wholeMinutes(g.minimum)),
		}
	}

	return Outcome{
		Result: promoloupev1alpha1.GateResultFailed,
		Message: fmt.Sprintf("The change has run for %d of %d minutes.",
			wholeMinutes(soaked), wholeMinutes(g.minimum)),
		RecheckAfter: g.minimum - soaked,
	}
}

// wholeMinutes returns d in whole minutes, rounded down.
func wholeMinutes(d time.Duration) int64 {
	return int64(d / time.Minute)
}
