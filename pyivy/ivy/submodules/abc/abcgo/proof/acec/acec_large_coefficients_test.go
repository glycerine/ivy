package acec

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// C stores coefficients as signed power-of-two exponents. These large values
// stay within C int arithmetic bounds and exercise carries and cancellation.
func TestSourceLargeSignedCoefficientMerging(t *testing.T) {
	cases := []struct{ in, want []int32 }{
		{[]int32{1000000000, 1000000000, -1000000000}, []int32{1000000000}},
		{[]int32{-1000000000, -1000000000, 999999999}, []int32{-1000000001, 999999999}},
		{[]int32{1000000000, -1000000000, 29, 29, -30}, []int32{}},
		{[]int32{-900000000, 899999999}, []int32{-899999999}},
	}
	for i, tc := range cases {
		v := vec.Vec_IntAlloc(len(tc.in))
		for _, n := range tc.in {
			Gia_PolynMergeConstOne(v, n)
		}
		got := vec.Vec_IntArray(v)
		if len(got) != len(tc.want) || len(got) > 0 && !reflect.DeepEqual(got, tc.want) {
			t.Errorf("case %d got %v, compiled C %v", i, got, tc.want)
		}
	}
}
