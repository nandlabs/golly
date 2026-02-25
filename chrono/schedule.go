package chrono

import (
	"time"
)

// IntervalSchedule represents a fixed-interval schedule.
// Jobs using this schedule will be executed repeatedly at the specified interval.
type IntervalSchedule struct {
	interval time.Duration
}

// NewIntervalSchedule creates a new IntervalSchedule with the given interval.
// The interval must be positive. Returns ErrInvalidInterval if the interval is <= 0.
func NewIntervalSchedule(interval time.Duration) (*IntervalSchedule, error) {
	if interval <= 0 {
		return nil, ErrInvalidInterval
	}
	return &IntervalSchedule{interval: interval}, nil
}

// Next returns the next activation time, which is from + interval.
func (s *IntervalSchedule) Next(from time.Time) time.Time {
	return from.Add(s.interval)
}

// Interval returns the configured interval duration.
func (s *IntervalSchedule) Interval() time.Duration {
	return s.interval
}

// OneShotSchedule represents a one-time schedule that fires once at a computed time.
// After the target time has passed, Next returns the zero time, indicating no more activations.
type OneShotSchedule struct {
	runAt time.Time
}

// NewOneShotSchedule creates a new OneShotSchedule that fires after the given delay from now.
// The delay must be non-negative. Returns ErrInvalidDelay if the delay is negative.
func NewOneShotSchedule(delay time.Duration) (*OneShotSchedule, error) {
	if delay < 0 {
		return nil, ErrInvalidDelay
	}
	return &OneShotSchedule{
		runAt: time.Now().Add(delay),
	}, nil
}

// NewOneShotScheduleAt creates a new OneShotSchedule that fires at the specified time.
func NewOneShotScheduleAt(at time.Time) *OneShotSchedule {
	return &OneShotSchedule{
		runAt: at,
	}
}

// Next returns the target time if from is before it, otherwise returns the zero time.
func (s *OneShotSchedule) Next(from time.Time) time.Time {
	if from.Before(s.runAt) {
		return s.runAt
	}
	return time.Time{}
}

// RunAt returns the configured target execution time.
func (s *OneShotSchedule) RunAt() time.Time {
	return s.runAt
}
