package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRefinementJustificationChoosesOneControllingInput(t *testing.T) {
	for _, propagate := range []int32{0, 1} {
		g := gia.Gia_ManStart(16)
		a, b, c, d := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x, y := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, c, d)
		root := gia.Gia_ManAppendAnd(g, x, y)
		gia.Gia_ManAppendCo(g, root^1)
		p := Rnm_ManStart(g)
		p.vMap = vec.Vec_IntAlloc(2)
		vec.Vec_IntPush(p.vMap, x>>1)
		vec.Vec_IntPush(p.vMap, y>>1)
		p.pCex = util.Abc_CexAlloc(0, 2, 1)
		p.fPropFanout = propagate
		Rnm_ManCollect(p)
		p.nObjsFrame = int32(1 + vec.Vec_IntSize(p.vMap) + vec.Vec_IntSize(p.vObjs))
		p.pObjs = make([]Rnm_Obj_t, p.nObjsFrame)
		Rnm_ManSensitize(p)
		selected := vec.Vec_IntAlloc(2)
		Rnm_ManJustify_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManPo(g, 0)), 0, selected)
		vec.Vec_IntUniqify(selected)
		got := vec.Vec_IntArray(selected)
		if len(got) != 1 || got[0] != x>>1 {
			t.Fatalf("fanout=%d selected=%v", propagate, got)
		}
		Rnm_ManStop(p, 0)
	}
}
