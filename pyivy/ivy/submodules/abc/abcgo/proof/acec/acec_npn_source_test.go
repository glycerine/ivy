package acec

import (
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestSourceNPNOrbitOrdering(t *testing.T) {
	for _, seeded := range []bool{false, true} {
		v := vec.Vec_IntAlloc(128)
		if seeded {
			vec.Vec_IntPush(v, 86)
			vec.Vec_IntPush(v, 222)
		}
		min := Extra_TruthCanonNPN2(0x56, 3, v)
		want := []int32{86, 54, 30, 166, 198, 154, 210, 180, 156, 106, 108, 120}
		if seeded {
			want = append([]int32{86, 222}, want[1:]...)
		}
		if min != 30 || !reflect.DeepEqual(vec.Vec_IntArray(v), want) {
			t.Fatalf("seeded=%v min=%d orbit=%v; ABC min=30 orbit=%v", seeded, min, vec.Vec_IntArray(v), want)
		}
	}
}
