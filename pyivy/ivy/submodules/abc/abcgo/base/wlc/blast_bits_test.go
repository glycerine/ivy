package wlc

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestConstantBitHelpers(t *testing.T) {
	bits := []int32{1, 0, 1, 8, 1}
	if got := Wlc_NtkCountConstBits(bits, 5); got != 4 {
		t.Fatalf("count %d", got)
	}
	if got := Wlc_BlastGetConst(bits, 3); got != 5 {
		t.Fatalf("value %d", got)
	}
	if got := Wlc_BlastGetConst(bits, 5); got != -1 {
		t.Fatalf("nonconstant value %d", got)
	}
}
func TestLoadFaninsSignAndZeroExtension(t *testing.T) {
	out := vec.Vec_IntAlloc(0)
	for _, tc := range []struct {
		signed int32
		want   []int32
	}{{0, []int32{3, 5, 0, 0}}, {1, []int32{3, 5, 5, 5}}} {
		got := Wlc_VecLoadFanins(out, []int32{3, 5}, 2, 4, tc.signed)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("signed %d: %v", tc.signed, got)
		}
	}
	if got := Wlc_VecCopy(out, []int32{2, 4, 6}, 2); !reflect.DeepEqual(got, []int32{2, 4}) {
		t.Fatal(got)
	}
}
