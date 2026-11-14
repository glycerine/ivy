package cba

import "testing"

func TestSourceNaturalNameComparison(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int32
	}{
		{"wire2", "wire10", -1}, {"wire19", "wire3", 1}, {"a_1_x", "a_2_x", -1},
		{"wire", "wire1", -1}, {"wire1", "wire", 1}, {"node3a", "node3b", -1},
		{"a#2", "a$1", -1}, {"x1000000000", "x999999999", 1},
	} {
		if got := Cba_StrCmp(&tc.a, &tc.b); got != tc.want {
			t.Errorf("%q %q => %d want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
func TestSourceNaturalNameEqualNumericSegmentAsserts(t *testing.T) {
	a, b := "node01", "node1"
	defer func() {
		if recover() == nil {
			t.Fatal("expected source equal-number assertion")
		}
	}()
	Cba_StrCmp(&a, &b)
}
