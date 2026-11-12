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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	promoloupev1alpha1 "github.com/promoloupe/promoloupe/api/v1alpha1"
)

// cronParser parses five-field cron expressions, the same syntax as
// Kubernetes CronJobs. Descriptors such as @daily and @every are not
// accepted, because a freeze window needs a fixed start time.
var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// freezeWindows blocks promotions while any of its windows is active.
type freezeWindows struct {
	windows []promoloupev1alpha1.FreezeWindow
}

// NewFreezeWindows is the Factory of the FreezeWindows gate. It builds the
// gate when the environment sets freezeWindows.
func NewFreezeWindows(config promoloupev1alpha1.Gates) (Gate, bool) {
	if len(config.FreezeWindows) == 0 {
		return nil, false
	}
	return freezeWindows{windows: config.FreezeWindows}, true
}

func (freezeWindows) Name() promoloupev1alpha1.GateName {
	return promoloupev1alpha1.GateFreezeWindows
}

func (g freezeWindows) Evaluate(_ context.Context, promotion Promotion) Outcome {
	now := promotion.Now.UTC()
	var activeUntil time.Time

	for _, window := range g.windows {
		schedule, err := parseSchedule(window.Schedule)
		if err != nil {
			return Outcome{
				Result:  promoloupev1alpha1.GateResultUnknown,
				Message: fmt.Sprintf("Freeze window %q is not valid: %v.", window.Schedule, err),
			}
		}

		// A window is active if it started within the last duration, that is,
		// if its first start after now minus the duration is not after now.
		duration := time.Duration(window.DurationMinutes) * time.Minute
		start := schedule.Next(now.Add(-duration))
		if start.IsZero() || start.After(now) {
			continue
		}
		if end := start.Add(duration); end.After(activeUntil) {
			activeUntil = end
		}
	}

	if activeUntil.IsZero() {
		return Outcome{
			Result:  promoloupev1alpha1.GateResultPassed,
			Message: "No freeze window is active.",
		}
	}

	return Outcome{
		Result:       promoloupev1alpha1.GateResultFailed,
		Message:      fmt.Sprintf("A freeze window is active until %s.", activeUntil.Format(time.RFC3339)),
		RecheckAfter: activeUntil.Sub(now),
	}
}

// parseSchedule parses the cron expression of a freeze window. Time zone
// prefixes are rejected, because freeze windows are always in UTC.
func parseSchedule(spec string) (cron.Schedule, error) {
	if strings.HasPrefix(spec, "TZ=") || strings.HasPrefix(spec, "CRON_TZ=") {
		return nil, errors.New("time zones are not supported, schedules are in UTC")
	}
	return cronParser.Parse(spec)
}
