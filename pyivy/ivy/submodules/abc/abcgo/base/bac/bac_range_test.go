package bac

import "testing"

func TestBacSourceRangeWidths(t *testing.T) {
	for _, tc := range []struct {
		text string
		want int32
	}{{"[7=3]", 5}, {"[2=5]", 4}, {"[4]", 1}, {"[-2=2]", 5}} {
		if got := bacRangeWidth(tc.text); got != tc.want {
			t.Fatalf("%q => %d, want %d", tc.text, got, tc.want)
		}
	}
}
