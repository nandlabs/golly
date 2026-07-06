package genai

import (
	"context"
	"errors"
	"testing"
)

type typedErr struct{ retry bool }

func (e typedErr) Error() string   { return "typed error" }
func (e typedErr) Retryable() bool { return e.retry }

func TestRetryable_Table(t *testing.T) {
	cases := []struct {
		name string
		err  error
		resp *GenResponse
		want bool
	}{
		{"nil error", nil, nil, false},
		{"429 rate limit", errors.New("openai API error: status 429 too many requests"), nil, true},
		{"503 unavailable", errors.New("claude API error: status 503 unavailable"), nil, true},
		{"500 server", errors.New("provider error: status 500"), nil, true},
		{"timeout", errors.New("request timed out"), nil, true},
		{"400 bad request", errors.New("openai API error: status 400 bad request"), nil, false},
		{"401 unauthorized", errors.New("status 401 unauthorized"), nil, false},
		{"content policy text", errors.New("blocked by content policy"), nil, false},
		{"typed retryable wins", typedErr{retry: true}, nil, true},
		{"typed non-retryable wins", typedErr{retry: false}, nil, false},
		{"context canceled", context.Canceled, nil, false},
		{"deadline exceeded", context.DeadlineExceeded, nil, false},
		{"unknown error", errors.New("something weird happened"), nil, false},
		{
			"content_filter finish reason not retryable",
			errors.New("status 503 unavailable"), // even a transient-looking error...
			&GenResponse{Candidates: []Candidate{{FinishReason: FinishReasonContentFilter}}},
			false, // ...is overridden by the content-filter finish reason
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := retryable(tc.err, tc.resp); got != tc.want {
				t.Errorf("retryable(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
