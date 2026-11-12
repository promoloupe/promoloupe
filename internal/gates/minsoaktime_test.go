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

package gates_test

import (
	"testing"
	"time"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
	"github.com/promoloupe/promoloupe/internal/gates"
)

// deployedAt is when the change reaches the source environment in the tests,
// on Friday 13 November 2026 at 09:00 UTC.
var deployedAt = time.Date(2026, time.November, 13, 9, 0, 0, 0, time.UTC)

func TestNewMinSoakTime(t *testing.T) {
	if _, ok := gates.NewMinSoakTime(promoloupev1alpha1.Gates{}); ok {
		t.Error("NewMinSoakTime built the gate without minSoakTimeMinutes")
	}
	if _, ok := gates.NewMinSoakTime(promoloupev1alpha1.Gates{MinSoakTimeMinutes: 60}); !ok {
		t.Error("NewMinSoakTime did not build the gate with minSoakTimeMinutes")
	}
}

func TestMinSoakTimeEvaluate(t *testing.T) {
	gate, _ := gates.NewMinSoakTime(promoloupev1alpha1.Gates{MinSoakTimeMinutes: 60})

	tests := []struct {
		name      string
		promotion gates.Promotion
		want      gates.Outcome
	}{
		{
			name:      "deployment time not known",
			promotion: gates.Promotion{Now: deployedAt},
			want: gates.Outcome{
				Result:  promoloupev1alpha1.GateResultUnknown,
				Message: "The time the change reached the source environment is not known yet.",
			},
		},
		{
			name:      "still soaking",
			promotion: gates.Promotion{Now: deployedAt.Add(45 * time.Minute), SourceDeployedAt: deployedAt},
			want: gates.Outcome{
				Result:       promoloupev1alpha1.GateResultFailed,
				Message:      "The change has run for 45 of 60 minutes.",
				RecheckAfter: 15 * time.Minute,
			},
		},
		{
			name:      "exactly the minimum",
			promotion: gates.Promotion{Now: deployedAt.Add(60 * time.Minute), SourceDeployedAt: deployedAt},
			want: gates.Outcome{
				Result:  promoloupev1alpha1.GateResultPassed,
				Message: "The change has run for 60 minutes, and 60 are required.",
			},
		},
		{
			name:      "longer than the minimum",
			promotion: gates.Promotion{Now: deployedAt.Add(90 * time.Minute), SourceDeployedAt: deployedAt},
			want: gates.Outcome{
				Result:  promoloupev1alpha1.GateResultPassed,
				Message: "The change has run for 90 minutes, and 60 are required.",
			},
		},
		{
			name:      "deployment time in the future",
			promotion: gates.Promotion{Now: deployedAt, SourceDeployedAt: deployedAt.Add(5 * time.Minute)},
			want: gates.Outcome{
				Result:       promoloupev1alpha1.GateResultFailed,
				Message:      "The change has run for 0 of 60 minutes.",
				RecheckAfter: 60 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gate.Evaluate(t.Context(), tt.promotion); got != tt.want {
				t.Errorf("Evaluate() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
