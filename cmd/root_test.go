package cmd

import (
	"testing"
	"time"

	"github.com/robfig/cron"
	"github.com/stretchr/testify/assert"
)

// periodicSchedule is a cron.Schedule that returns t+period for any input t.
// It lets tests model arbitrary inter-tick deltas without depending on real
// clocks or cron-spec parsing.
type periodicSchedule struct {
	period time.Duration
}

func (s periodicSchedule) Next(t time.Time) time.Time { return t.Add(s.period) }

func TestRunScheduleLoop(t *testing.T) {
	base := time.Date(2026, 5, 3, 3, 0, 0, 0, time.UTC)

	cases := []struct {
		name      string
		next      time.Time
		ticks     []time.Time
		wantFires int
	}{
		{
			name:      "tick before next does not fire",
			next:      base,
			ticks:     []time.Time{base.Add(-time.Second)},
			wantFires: 0,
		},
		{
			name:      "tick at or after next fires once",
			next:      base,
			ticks:     []time.Time{base.Add(time.Second)},
			wantFires: 1,
		},
		{
			name:      "two ticks in same window fire once",
			next:      base,
			ticks:     []time.Time{base.Add(time.Second), base.Add(2 * time.Second)},
			wantFires: 1,
		},
		{
			name:      "tick after long sleep collapses missed windows into one fire",
			next:      base,
			ticks:     []time.Time{base.Add(3 * 24 * time.Hour)},
			wantFires: 1,
		},
		{
			name:      "tick crossing two separate windows fires twice",
			next:      base,
			ticks:     []time.Time{base.Add(time.Second), base.Add(2 * time.Minute)},
			wantFires: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tick := make(chan time.Time, len(tc.ticks))
			for _, ts := range tc.ticks {
				tick <- ts
			}
			close(tick)

			fired := 0
			runScheduleLoop(
				periodicSchedule{period: time.Minute},
				tc.next,
				tick,
				nil,
				func(time.Time) { fired++ },
			)

			assert.Equal(t, tc.wantFires, fired)
		})
	}
}

func TestRunScheduleLoop_StopChannelReturns(t *testing.T) {
	base := time.Date(2026, 5, 3, 3, 0, 0, 0, time.UTC)
	tick := make(chan time.Time)
	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		runScheduleLoop(
			periodicSchedule{period: time.Minute},
			base,
			tick,
			stop,
			func(time.Time) {},
		)
		close(done)
	}()

	close(stop)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runScheduleLoop did not return after stop was closed")
	}
}

func TestRunScheduleLoop_RealCronScheduleAdvancesPastSlept(t *testing.T) {
	// "0 0 3 * * *" — daily at 03:00. Mirrors the original bug report.
	schedule, err := cron.Parse("0 0 3 * * *")
	assert.NoError(t, err)

	loc := time.UTC
	scheduledAt := time.Date(2026, 5, 1, 3, 0, 0, 0, loc)
	wakeAt := time.Date(2026, 5, 3, 10, 30, 0, 0, loc) // 2 days, 7.5 hours past

	tick := make(chan time.Time, 1)
	tick <- wakeAt
	close(tick)

	var fireAt time.Time
	fireCount := 0
	runScheduleLoop(schedule, scheduledAt, tick, nil, func(now time.Time) {
		fireCount++
		fireAt = now
	})

	assert.Equal(t, 1, fireCount, "should fire exactly once even after multiple missed windows")
	assert.Equal(t, wakeAt, fireAt, "should fire with the wake-time tick value")

	nextAfterFire := schedule.Next(fireAt)
	expectedNext := time.Date(2026, 5, 4, 3, 0, 0, 0, loc)
	assert.Equal(t, expectedNext, nextAfterFire, "next window after fire should be the upcoming 03:00, not a past one")
}
