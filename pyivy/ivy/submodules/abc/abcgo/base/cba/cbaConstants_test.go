package cba

import (
	"reflect"
	"testing"
)

func TestConstantBitDecoding(t *testing.T) {
	for _, tc := range []struct {
		text   string
		total  int32
		signed bool
		want   []int32
	}{
		{"4'b1010", 4, false, []int32{0, 1, 0, 1}},
		{"'b1", 1, false, []int32{1}},
		{"4'b1x0z", 4, false, []int32{1, 0, 1, 1}},
		{"7'h2F", 7, false, []int32{1, 1, 1, 1, 0, 1, 0}},
		{"6'd19", 6, false, []int32{1, 1, 0, 0, 1, 0}},
		// The source passes total-current to Vec_IntFillExtra, a target size.
		{"3'sb101", 8, true, []int32{1, 0, 1, 1, 1}},
	} {
		if got := cbaConstBits(tc.text, tc.total, tc.signed); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s => %v, want %v", tc.text, got, tc.want)
		}
	}
}

func TestReadHexDigit(t *testing.T) {
	for _, digits := range []string{"0123456789ABCDEF", "0123456789abcdef"} {
		for want := range digits {
			if got := Cba_ReadHexDigit(digits[want]); got != int32(want) {
				t.Fatalf("%q=%d want %d", digits[want], got, want)
			}
		}
	}
}
