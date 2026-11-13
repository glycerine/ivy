package ssw

import "testing"

func TestSignalFilterSourceRemainingLimits(t *testing.T) {
	tick := int64(sswTicksPerSecond)
	for _, tc := range []struct {
		limit, cap int32
		remaining  int64
		want       int32
	}{
		{0, 0, 0, 0}, {0, 4, 0, 4}, {10, 0, 3 * tick, 3}, {10, 2, 3 * tick, 2},
		{10, 4, tick / 2, 4}, {10, 4, -tick / 2, 4}, {10, 4, -2 * tick, -2},
	} {
		got := sswFilterPartLimit(tc.limit, tc.cap, 10*tick, 10*tick-tc.remaining)
		if got != tc.want {
			t.Fatalf("limit=%d cap=%d remaining=%d got=%d want=%d", tc.limit, tc.cap, tc.remaining, got, tc.want)
		}
	}
}

func TestSourceDeadlineStrictBoundaryAndZeroConvention(t *testing.T) {
	for _, tc := range []struct {
		name     string
		limit    int32
		now, end int64
		want     bool
	}{
		{"disabled", 0, 100, 0, false},
		{"exact boundary", 1, 10, 10, false},
		{"one tick over", 1, 11, 10, true},
		{"negative nonzero limit", -1, 0, -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sswDeadlineExceeded(tc.limit, tc.now, tc.end); got != tc.want {
				t.Fatalf("deadline result %t, want %t", got, tc.want)
			}
		})
	}
}
