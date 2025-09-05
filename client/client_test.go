package client

import (
	"testing"
	"time"
)

func TestPctDue(t *testing.T) {
	for i, td := range []struct {
		now      time.Time
		pct      float64
		expiry   time.Time
		validity time.Duration
		expected bool
	}{
		{
			now:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			pct:      0.8,
			expiry:   time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC),
			validity: 14 * 24 * time.Hour,
			expected: false,
		},
		{
			now:      time.Date(2022, 1, 13, 0, 0, 0, 0, time.UTC),
			pct:      0.8,
			expiry:   time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC),
			validity: 14 * 24 * time.Hour,
			expected: true,
		},
	} {
		pctDueFunc := newPctDueAt(func() time.Time { return td.now }, td.pct)

		if have, want := pctDueFunc(td.expiry, td.validity), td.expected; have != want {
			t.Errorf("expected %v but got %v for index %d", want, have, i)
		}
	}
}
