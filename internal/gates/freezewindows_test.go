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
	"strings"
	"testing"
	"time"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
	"github.com/promoloupe/promoloupe/internal/gates"
)

// weekendFreeze starts on Friday at 18:00 UTC and lasts 64 hours, until
// Monday at 10:00 UTC.
var weekendFreeze = promoloupev1alpha1.FreezeWindow{Schedule: "0 18 * * 5", DurationMinutes: 3840}

// Messages of the FreezeWindows gate in the tests.
const (
	notFrozen         = "No freeze window is active."
	frozenUntilMonday = "A freeze window is active until 2026-11-16T10:00:00Z."
)

// utc returns a time on the given day of November 2026, in UTC.
func utc(day, hour, minute int) time.Time {
	return time.Date(2026, time.November, day, hour, minute, 0, 0, time.UTC)
}

func TestNewFreezeWindows(t *testing.T) {
	if _, ok := gates.NewFreezeWindows(promoloupev1alpha1.Gates{}); ok {
		t.Error("NewFreezeWindows built the gate without freezeWindows")
	}
	config := promoloupev1alpha1.Gates{FreezeWindows: []promoloupev1alpha1.FreezeWindow{weekendFreeze}}
	if _, ok := gates.NewFreezeWindows(config); !ok {
		t.Error("NewFreezeWindows did not build the gate with freezeWindows")
	}
}

func TestFreezeWindowsEvaluate(t *testing.T) {
	tests := []struct {
		name             string
		windows          []promoloupev1alpha1.FreezeWindow
		now              time.Time
		wantResult       promoloupev1alpha1.GateResult
		wantMessage      string
		wantRecheckAfter time.Duration
	}{
		{
			name:        "before the window",
			windows:     []promoloupev1alpha1.FreezeWindow{weekendFreeze},
			now:         utc(13, 17, 59),
			wantResult:  promoloupev1alpha1.GateResultPassed,
			wantMessage: notFrozen,
		},
		{
			name:             "when the window starts",
			windows:          []promoloupev1alpha1.FreezeWindow{weekendFreeze},
			now:              utc(13, 18, 0),
			wantResult:       promoloupev1alpha1.GateResultFailed,
			wantMessage:      frozenUntilMonday,
			wantRecheckAfter: 64 * time.Hour,
		},
		{
			name:             "during the window",
			windows:          []promoloupev1alpha1.FreezeWindow{weekendFreeze},
			now:              utc(15, 12, 0),
			wantResult:       promoloupev1alpha1.GateResultFailed,
			wantMessage:      frozenUntilMonday,
			wantRecheckAfter: 22 * time.Hour,
		},
		{
			name:        "when the window ends",
			windows:     []promoloupev1alpha1.FreezeWindow{weekendFreeze},
			now:         utc(16, 10, 0),
			wantResult:  promoloupev1alpha1.GateResultPassed,
			wantMessage: notFrozen,
		},
		{
			name:             "now in another time zone",
			windows:          []promoloupev1alpha1.FreezeWindow{weekendFreeze},
			now:              time.Date(2026, time.November, 13, 19, 0, 0, 0, time.FixedZone("CET", 3600)),
			wantResult:       promoloupev1alpha1.GateResultFailed,
			wantMessage:      frozenUntilMonday,
			wantRecheckAfter: 64 * time.Hour,
		},
		{
			name: "overlapping windows",
			windows: []promoloupev1alpha1.FreezeWindow{
				weekendFreeze,
				{Schedule: "0 0 * * 6", DurationMinutes: 4320},
			},
			now:              utc(14, 1, 0),
			wantResult:       promoloupev1alpha1.GateResultFailed,
			wantMessage:      "A freeze window is active until 2026-11-17T00:00:00Z.",
			wantRecheckAfter: 71 * time.Hour,
		},
		{
			name:        "schedule that never fires",
			windows:     []promoloupev1alpha1.FreezeWindow{{Schedule: "0 0 30 2 *", DurationMinutes: 60}},
			now:         utc(13, 18, 0),
			wantResult:  promoloupev1alpha1.GateResultPassed,
			wantMessage: notFrozen,
		},
		{
			name:        "invalid schedule",
			windows:     []promoloupev1alpha1.FreezeWindow{{Schedule: "0 18 * *", DurationMinutes: 60}},
			now:         utc(13, 18, 0),
			wantResult:  promoloupev1alpha1.GateResultUnknown,
			wantMessage: `Freeze window "0 18 * *" is not valid`,
		},
		{
			name:        "time zone prefix",
			windows:     []promoloupev1alpha1.FreezeWindow{{Schedule: "CRON_TZ=UTC", DurationMinutes: 60}},
			now:         utc(13, 18, 0),
			wantResult:  promoloupev1alpha1.GateResultUnknown,
			wantMessage: "time zones are not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gate, _ := gates.NewFreezeWindows(promoloupev1alpha1.Gates{FreezeWindows: tt.windows})

			got := gate.Evaluate(t.Context(), gates.Promotion{Now: tt.now})

			if got.Result != tt.wantResult {
				t.Errorf("Result = %q, want %q", got.Result, tt.wantResult)
			}
			if !strings.Contains(got.Message, tt.wantMessage) {
				t.Errorf("Message = %q, want it to contain %q", got.Message, tt.wantMessage)
			}
			if got.RecheckAfter != tt.wantRecheckAfter {
				t.Errorf("RecheckAfter = %v, want %v", got.RecheckAfter, tt.wantRecheckAfter)
			}
		})
	}
}
