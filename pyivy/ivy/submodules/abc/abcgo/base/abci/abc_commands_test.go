package abci

import (
	"reflect"
	"testing"
)

func TestAbcTruth32ToWords(t *testing.T) {
	tests := []struct {
		name  string
		input []uint32
		want  []uint64
	}{
		{name: "paired words", input: []uint32{0x89abcdef, 0x01234567}, want: []uint64{0x0123456789abcdef}},
		{name: "odd final word", input: []uint32{0xffffffff, 0x80000000, 0x12345678}, want: []uint64{0x80000000ffffffff, 0x12345678}},
		{name: "empty", input: nil, want: []uint64{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := abcTruth32ToWords(tc.input); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("abcTruth32ToWords(%08x) = %016x, want %016x", tc.input, got, tc.want)
			}
		})
	}
}

func TestAbcAtoiMatchesCDecimalPrefix(t *testing.T) {
	for input, want := range map[string]int32{
		"17":         17,
		"-4":         -4,
		"  +23tail":  23,
		"not-a-num":  0,
		"+no-digits": 0,
	} {
		if got := abcAtoi(input); got != want {
			t.Errorf("abcAtoi(%q) = %d, want %d", input, got, want)
		}
	}
}
