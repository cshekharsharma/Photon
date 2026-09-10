package flashdb

import (
	"testing"
	"time"
)

func TestRecord_IsExpired(t *testing.T) {
	now := time.Now()

	t.Run("no TTL (zero expiry) => not expired", func(t *testing.T) {
		r := &record{value: map[string]any{"x": 1}, size: 10, expiry: time.Time{}}
		if r.isExpired(now) {
			t.Fatalf("zero expiry should not be expired")
		}
	})

	t.Run("expiry in the future => not expired", func(t *testing.T) {
		r := &record{value: map[string]any{"x": 1}, size: 10, expiry: now.Add(10 * time.Second)}
		if r.isExpired(now) {
			t.Fatalf("future expiry should not be expired")
		}
	})

	t.Run("expiry exactly at 'now' => not expired (strict After)", func(t *testing.T) {
		exp := now
		r := &record{value: map[string]any{"x": 1}, size: 10, expiry: exp}

		if r.isExpired(now) {
			t.Fatalf("expiry == now should not be expired")
		}
	})

	t.Run("expiry in the past => expired", func(t *testing.T) {
		r := &record{value: map[string]any{"x": 1}, size: 10, expiry: now.Add(-1 * time.Nanosecond)}
		if !r.isExpired(now) {
			t.Fatalf("past expiry should be expired")
		}
	})
}
