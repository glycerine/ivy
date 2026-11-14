package cba

import "testing"

func TestSourceVerilogCharacterClasses(t *testing.T) {
	for _, tc := range []struct {
		name     string
		fn       func(byte) int32
		accepted string
	}{
		{"space", Prs_CharIsSpace, " \t\r\n"},
		{"digit", Prs_CharIsDigit, "0123456789"},
		{"binary", Prs_CharIsDigitB, "01xz"},
		{"hex", Prs_CharIsDigitH, "0123456789ABCDEFabcdef"},
		{"letter", Prs_CharIsChar, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"symbol1", Prs_CharIsSymb1, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"},
		{"symbol2", Prs_CharIsSymb2, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789$"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var accepted [256]bool
			for _, c := range []byte(tc.accepted) {
				accepted[c] = true
			}
			for i := 0; i < 256; i++ {
				want := int32(0)
				if accepted[i] {
					want = 1
				}
				if got := tc.fn(byte(i)); got != want {
					t.Fatalf("byte %d got %d want %d", i, got, want)
				}
			}
		})
	}
}
