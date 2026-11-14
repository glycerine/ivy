package cba

import "testing"

func TestSourceBlifCharacterClasses(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fn       func(byte) int32
		accepted string
	}{
		{"space", prsBlifCharIsSpace, " \t\r"}, {"stop", Prs_CharIsStop, "#\\\n="}, {"literal", Prs_CharIsLit, "01-"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var set [256]bool
			for _, b := range []byte(tc.accepted) {
				set[b] = true
			}
			for i := 0; i < 256; i++ {
				want := int32(0)
				if set[i] {
					want = 1
				}
				if got := tc.fn(byte(i)); got != want {
					t.Fatalf("byte %d got %d want %d", i, got, want)
				}
			}
		})
	}
}
