package clients

import (
	"testing"
	"time"
)

func TestWaitTime_FixedBackoff(t *testing.T) {
	r := &RetryInfo{
		MaxRetries: 3,
		Wait:       100,
	}

	for i := 0; i < 3; i++ {
		got := r.WaitTime(i)
		want := 100 * time.Millisecond
		if got != want {
			t.Errorf("WaitTime(%d) = %v, want %v", i, got, want)
		}
	}
}

func TestWaitTime_ExponentialBackoff_DefaultMultiplier(t *testing.T) {
	r := &RetryInfo{
		MaxRetries:  5,
		Wait:        100,
		Exponential: true,
		// Multiplier defaults to 2 when <= 0
	}

	tests := []struct {
		retryCount int
		want       time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, 1600 * time.Millisecond},
	}

	for _, tt := range tests {
		got := r.WaitTime(tt.retryCount)
		if got != tt.want {
			t.Errorf("WaitTime(%d) = %v, want %v", tt.retryCount, got, tt.want)
		}
	}
}

func TestWaitTime_ExponentialBackoff_CustomMultiplier(t *testing.T) {
	r := &RetryInfo{
		MaxRetries:  3,
		Wait:        100,
		Exponential: true,
		Multiplier:  3,
	}

	tests := []struct {
		retryCount int
		want       time.Duration
	}{
		{0, 100 * time.Millisecond},  // base * multiplier^0
		{1, 300 * time.Millisecond},  // base * multiplier^1
		{2, 900 * time.Millisecond},  // base * multiplier^2
		{3, 2700 * time.Millisecond}, // base * multiplier^3
	}

	for _, tt := range tests {
		got := r.WaitTime(tt.retryCount)
		if got != tt.want {
			t.Errorf("WaitTime(%d) = %v, want %v", tt.retryCount, got, tt.want)
		}
	}
}

func TestWaitTime_ExponentialBackoff_MaxWaitCap(t *testing.T) {
	r := &RetryInfo{
		MaxRetries:  5,
		Wait:        100,
		Exponential: true,
		MaxWait:     500,
	}

	tests := []struct {
		retryCount int
		want       time.Duration
	}{
		{0, 100 * time.Millisecond}, // base * multiplier^0
		{1, 200 * time.Millisecond}, // base * multiplier^1
		{2, 400 * time.Millisecond}, // base * multiplier^2
		{3, 500 * time.Millisecond}, // base * multiplier^3, capped to MaxWait
		{4, 500 * time.Millisecond}, // base * multiplier^4, capped to MaxWait
	}

	for _, tt := range tests {
		got := r.WaitTime(tt.retryCount)
		if got != tt.want {
			t.Errorf("WaitTime(%d) = %v, want %v", tt.retryCount, got, tt.want)
		}
	}
}

func TestWaitTime_Jitter(t *testing.T) {
	r := &RetryInfo{
		MaxRetries:  3,
		Wait:        100,
		Exponential: true,
		Jitter:      true,
	}

	// Jitter adds [0, backoff) to the computed backoff, so the result
	// should be in [backoff, 2*backoff).
	for i := 0; i < 3; i++ {
		rNoJitter := &RetryInfo{
			Wait:        100,
			Exponential: true,
		}
		baseWait := rNoJitter.WaitTime(i)

		// Run multiple times to check the range
		for j := 0; j < 20; j++ {
			got := r.WaitTime(i)
			if got < baseWait || got >= 2*baseWait {
				t.Errorf("WaitTime(%d) with jitter = %v, want in [%v, %v)", i, got, baseWait, 2*baseWait)
			}
		}
	}
}

func TestWaitTime_ZeroWait(t *testing.T) {
	r := &RetryInfo{
		MaxRetries: 3,
		Wait:       0,
	}

	got := r.WaitTime(0)
	if got != 0 {
		t.Errorf("WaitTime(0) with Wait=0 = %v, want 0", got)
	}
}

func TestWaitTime_FixedBackoff_IgnoresMaxWait(t *testing.T) {
	r := &RetryInfo{
		MaxRetries: 3,
		Wait:       1000,
		MaxWait:    500, // only applies when Exponential=true
	}

	got := r.WaitTime(0)
	want := 1000 * time.Millisecond
	if got != want {
		t.Errorf("WaitTime(0) = %v, want %v (MaxWait should be ignored for fixed backoff)", got, want)
	}
}
