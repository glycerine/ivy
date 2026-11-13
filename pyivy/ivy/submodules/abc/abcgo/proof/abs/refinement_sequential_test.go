package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRefinementPropagatesAcrossFrames(t *testing.T) {
	g := gia.Gia_ManStart(8)
	input, flop := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, flop)
	gia.Gia_ManAppendCo(g, input)
	gia.Gia_ManSetRegNum(g, 1)
	mapping := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(mapping, input>>1)
	cex := util.Abc_CexAlloc(0, 1, 2)
	cex.IFrame = 1
	util.Abc_InfoSetBit(cex.PData, 0)
	p := Rnm_ManStart(g)
	if result := Rnm_ManRefine(p, cex, mapping, 0, 0, 0); vec.Vec_IntSize(result) != 0 {
		t.Fatalf("true counterexample refined %v", vec.Vec_IntArray(result))
	}
	if gia.Gia_ObjTerSimGet1(gia.Gia_ManPo(g, 0)) == 0 {
		t.Fatal("ternary verification lost previous-frame register input")
	}
	Rnm_ManStop(p, 0)
	q := Rf2_ManStart(g)
	defer Rf2_ManStop(q, 0)
	q.pCex, q.vMap = cex, mapping
	Rf2_ManCollect(q)
	Rf2_ManAssignJustIds(q)
	if result := vec.Vec_IntArray(Rf2_ManPropagate(q, 32)); len(result) != 1 || result[0] != 0 {
		t.Fatalf("alternate masks %v", result)
	}
	if gia.Gia_ObjMark0(gia.Gia_ManPo(g, 0)) != 1 {
		t.Fatal("alternate propagation lost previous-frame register input")
	}
}

func TestRefinementSelectsPseudoInput(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	p := Rnm_ManStart(g)
	defer Rnm_ManStop(p, 0)
	mapping := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(mapping, gate>>1)
	cex := util.Abc_CexAlloc(0, 1, 1)
	util.Abc_InfoSetBit(cex.PData, 0)
	old := refinementSelectionCounter.Swap(8)
	defer refinementSelectionCounter.Store(old)
	got := vec.Vec_IntArray(Rnm_ManRefine(p, cex, mapping, 0, 1, 0))
	if len(got) != 1 || got[0] != gate>>1 {
		t.Fatalf("refinement=%v", got)
	}
}
