package goivy

import "testing"

// Python's ivy_parser.py has 'complexact : IF fmla sequence' only for
// get_numeric_version() <= [1,4]; later versions use somefmla.
func TestParser16Pre15(t *testing.T) {
	for _, tc := range []struct {
		v    Version
		want bool
	}{{Version{1, 3}, true}, {Version{1, 4}, true}, {Version{1, 5}, false}, {Version{1, 6}, false}} {
		if got := parser16Pre15(newParser16LexAdapter("", tc.v)); got != tc.want {
			t.Errorf("parser16Pre15(%v) = %v, want %v", tc.v, got, tc.want)
		}
	}
}
