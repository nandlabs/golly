package messaging

import (
	"testing"
	"time"
)

func TestAddDeliveryGuarantee(t *testing.T) {
	opts := NewOptionsBuilder().AddDeliveryGuarantee(AtLeastOnce).Build()
	got, ok := GetOptValue[DeliveryGuarantee](DeliveryGuaranteeOpt, opts...)
	if !ok || got != AtLeastOnce {
		t.Errorf("DeliveryGuarantee roundtrip wrong: ok=%v val=%v", ok, got)
	}
}

func TestAddDeadLetter(t *testing.T) {
	opts := NewOptionsBuilder().AddDeadLetter("dlq.events").Build()
	got, ok := GetOptValue[string](DeadLetterOpt, opts...)
	if !ok || got != "dlq.events" {
		t.Errorf("DeadLetter roundtrip wrong: ok=%v val=%q", ok, got)
	}
}

func TestAddMaxDeliveryAttempts(t *testing.T) {
	opts := NewOptionsBuilder().AddMaxDeliveryAttempts(5).Build()
	got, ok := GetOptValue[int](MaxDeliveryAttemptsOpt, opts...)
	if !ok || got != 5 {
		t.Errorf("MaxDeliveryAttempts roundtrip wrong: ok=%v val=%d", ok, got)
	}
}

func TestAddMaxDeliveryAttempts_ZeroIsNoop(t *testing.T) {
	opts := NewOptionsBuilder().AddMaxDeliveryAttempts(0).Build()
	if _, ok := GetOptValue[int](MaxDeliveryAttemptsOpt, opts...); ok {
		t.Errorf("zero MaxDeliveryAttempts should not be recorded")
	}
}

func TestAddConsumerMode(t *testing.T) {
	cases := []ConsumerMode{ConsumerPush, ConsumerPull}
	for _, m := range cases {
		opts := NewOptionsBuilder().AddConsumerMode(m).Build()
		got, ok := GetOptValue[ConsumerMode](ConsumerModeOpt, opts...)
		if !ok || got != m {
			t.Errorf("ConsumerMode(%v) roundtrip wrong: ok=%v val=%v", m, ok, got)
		}
	}
}

func TestAddAckTimeout(t *testing.T) {
	opts := NewOptionsBuilder().AddAckTimeout(30 * time.Second).Build()
	got, ok := GetOptValue[time.Duration](AckTimeoutOpt, opts...)
	if !ok || got != 30*time.Second {
		t.Errorf("AckTimeout roundtrip wrong: ok=%v val=%v", ok, got)
	}
}

func TestAddAckTimeout_ZeroIsNoop(t *testing.T) {
	opts := NewOptionsBuilder().AddAckTimeout(0).Build()
	if _, ok := GetOptValue[time.Duration](AckTimeoutOpt, opts...); ok {
		t.Errorf("zero AckTimeout should not be recorded")
	}
}

func TestBrokerOptions_CompositeBuild(t *testing.T) {
	opts := NewOptionsBuilder().
		AddDeliveryGuarantee(ExactlyOnce).
		AddDeadLetter("dlq").
		AddMaxDeliveryAttempts(3).
		AddConsumerMode(ConsumerPull).
		AddAckTimeout(10 * time.Second).
		Build()

	if len(opts) != 5 {
		t.Fatalf("expected 5 options recorded; got %d", len(opts))
	}
	if g, _ := GetOptValue[DeliveryGuarantee](DeliveryGuaranteeOpt, opts...); g != ExactlyOnce {
		t.Errorf("DeliveryGuarantee wrong: %v", g)
	}
	if d, _ := GetOptValue[string](DeadLetterOpt, opts...); d != "dlq" {
		t.Errorf("DeadLetter wrong: %q", d)
	}
}
