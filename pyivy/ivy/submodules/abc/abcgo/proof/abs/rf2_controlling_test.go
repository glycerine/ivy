package abs

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAlternateRefinementControllingValues(t *testing.T) {
	for _, tc := range []struct {
		bits int
		want []int32
	}{{0, []int32{1, 2}}, {1, []int32{2}}, {2, []int32{1}}} {
		g := gia.Gia_ManStart(16)
		a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, c, d)
		root := gia.Gia_ManAppendAnd(g, x, y)
		gia.Gia_ManAppendCo(g, root^1)
		p := Rf2_ManStart(g)
		p.vMap = vec.Vec_IntAlloc(2)
		vec.Vec_IntPush(p.vMap, x>>1)
		vec.Vec_IntPush(p.vMap, y>>1)
		p.pCex = util.Abc_CexAlloc(0, 2, 1)
		for i := int32(0); i < 2; i++ {
			if tc.bits&(1<<i) != 0 {
				util.Abc_InfoSetBit(p.pCex.PData, i)
			}
		}
		Rf2_ManCollect(p)
		Rf2_ManAssignJustIds(p)
		got := vec.Vec_IntArray(Rf2_ManPropagate(p, 32))
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("bits=%d masks=%v want=%v", tc.bits, got, tc.want)
		}
		Rf2_ManStop(p, 0)
	}
}
