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
	"slices"
	"testing"
	"time"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
	"github.com/promoloupe/promoloupe/internal/gates"
)

// gateNames returns the names of the gates, in order.
func gateNames(built []gates.Gate) []promoloupev1alpha1.GateName {
	names := make([]promoloupev1alpha1.GateName, 0, len(built))
	for _, gate := range built {
		names = append(names, gate.Name())
	}
	return names
}

func TestDefaultRegistryBuild(t *testing.T) {
	tests := []struct {
		name   string
		config promoloupev1alpha1.Gates
		want   []promoloupev1alpha1.GateName
	}{
		{
			name:   "no gates",
			config: promoloupev1alpha1.Gates{},
			want:   []promoloupev1alpha1.GateName{},
		},
		{
			name: "every shipped gate, in registry order",
			config: promoloupev1alpha1.Gates{
				FreezeWindows:      []promoloupev1alpha1.FreezeWindow{weekendFreeze},
				MinSoakTimeMinutes: 60,
			},
			want: []promoloupev1alpha1.GateName{
				promoloupev1alpha1.GateMinSoakTime,
				promoloupev1alpha1.GateFreezeWindows,
			},
		},
		{
			name:   "a gate that is not shipped yet",
			config: promoloupev1alpha1.Gates{RequiredApprovals: 1},
			want:   []promoloupev1alpha1.GateName{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gateNames(gates.DefaultRegistry().Build(tt.config)); !slices.Equal(got, tt.want) {
				t.Errorf("Build() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	built := gates.DefaultRegistry().Build(promoloupev1alpha1.Gates{
		MinSoakTimeMinutes: 60,
		FreezeWindows:      []promoloupev1alpha1.FreezeWindow{weekendFreeze},
	})

	tests := []struct {
		name             string
		promotion        gates.Promotion
		wantResults      []promoloupev1alpha1.GateResult
		wantRecheckAfter time.Duration
		wantAllPassed    bool
	}{
		{
			name:          "every gate passes",
			promotion:     gates.Promotion{Now: utc(13, 11, 0), SourceDeployedAt: utc(13, 9, 0)},
			wantResults:   []promoloupev1alpha1.GateResult{promoloupev1alpha1.GateResultPassed, promoloupev1alpha1.GateResultPassed},
			wantAllPassed: true,
		},
		{
			name:      "the shortest recheck wins",
			promotion: gates.Promotion{Now: utc(13, 18, 0), SourceDeployedAt: utc(13, 17, 45)},
			wantResults: []promoloupev1alpha1.GateResult{
				promoloupev1alpha1.GateResultFailed,
				promoloupev1alpha1.GateResultFailed,
			},
			wantRecheckAfter: 45 * time.Minute,
		},
		{
			name:      "an unknown result blocks the promotion",
			promotion: gates.Promotion{Now: utc(13, 11, 0)},
			wantResults: []promoloupev1alpha1.GateResult{
				promoloupev1alpha1.GateResultUnknown,
				promoloupev1alpha1.GateResultPassed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statuses, recheckAfter := gates.Evaluate(t.Context(), built, tt.promotion)

			results := make([]promoloupev1alpha1.GateResult, 0, len(statuses))
			for _, status := range statuses {
				results = append(results, status.Result)
			}
			if !slices.Equal(results, tt.wantResults) {
				t.Errorf("results = %v, want %v", results, tt.wantResults)
			}
			if recheckAfter != tt.wantRecheckAfter {
				t.Errorf("recheckAfter = %v, want %v", recheckAfter, tt.wantRecheckAfter)
			}
			if got := gates.AllPassed(statuses); got != tt.wantAllPassed {
				t.Errorf("AllPassed() = %v, want %v", got, tt.wantAllPassed)
			}
		})
	}
}

func TestAllPassedWithoutGates(t *testing.T) {
	if !gates.AllPassed(nil) {
		t.Error("AllPassed(nil) = false, want true")
	}
}
