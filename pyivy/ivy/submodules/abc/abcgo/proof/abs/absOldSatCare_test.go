package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestRefinementSatAssumptionsMapToInputsAndFrames(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.NewInput()
	frames := aig.NewGraph()
	frames.NewInput()
	frames.NewInput()
	frames.NewInput()
	cex := util.Abc_CexAlloc(0, 2, 2)
	p := &Saig_RefMan_t{pAig: g, pFrames: frames, pCex: cex, vMapPiF2A: vec.Vec_IntAllocArrayCopy([]int32{1, 0, 0, 1, 1, 1}, 6)}
	vars := vec.Vec_IntAllocArrayCopy([]int32{0, 1, 2}, 3)
	assumptions := vec.Vec_IntAllocArrayCopy([]int32{0, 2, 5}, 3)
	grouped := Saig_RefManOrderLiterals(p, vars, assumptions)
	if vec.Vec_VecSize(grouped) != 2 || !reflect.DeepEqual(vec.Vec_IntArray(vec.Vec_VecEntryInt(grouped, 0)), []int32{0, 5}) || !reflect.DeepEqual(vec.Vec_IntArray(vec.Vec_VecEntryInt(grouped, 1)), []int32{2}) {
		t.Fatal("wrong grouped literals")
	}
	care := Saig_RefManCreateCex(p, vars, assumptions)
	for i := int32(0); i < 4; i++ {
		if util.Abc_InfoHasBit(care.PData, i) != (i != 0) {
			t.Fatalf("incorrect care bit %d", i)
		}
	}
}

func TestRefinementReasonsDeduplicateAcrossFrames(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.NewInput()
	frames := aig.NewGraph()
	frames.NewInput()
	frames.NewInput()
	frames.NewInput()
	p := &Saig_RefMan_t{pAig: g, pFrames: frames, pCex: util.Abc_CexAlloc(0, 2, 2), vMapPiF2A: vec.Vec_IntAllocArrayCopy([]int32{1, 0, 0, 1, 1, 1}, 6)}
	reasons := vec.Vec_IntAllocArrayCopy([]int32{0, 1, 2}, 3)
	if got := vec.Vec_IntArray(Saig_RefManReason2Inputs(p, reasons)); !reflect.DeepEqual(got, []int32{1, 0}) {
		t.Fatalf("inputs: %v", got)
	}
	care := Saig_RefManReason2Cex(p, reasons)
	for i := int32(0); i < 4; i++ {
		if util.Abc_InfoHasBit(care.PData, i) != (i != 0) {
			t.Fatalf("care bit %d", i)
		}
	}
}

func TestRefinementSetPhasesRespectsCare(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, y))
	cex := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(cex.PData, 0)
	util.Abc_InfoSetBit(cex.PData, 1)
	p := &Saig_RefMan_t{pFrames: g, pCex: cex, vMapPiF2A: vec.Vec_IntAllocArrayCopy([]int32{0, 0, 1, 0}, 4)}
	care := util.Abc_CexAlloc(0, 2, 1)
	util.Abc_InfoSetBit(care.PData, 0)
	if Saig_RefManSetPhases(p, care, 0) != 0 {
		t.Fatal("don't-care set to zero should disable AND")
	}
	if Saig_RefManSetPhases(p, care, 1) != 1 {
		t.Fatal("don't-care set to one should enable AND")
	}
	if Saig_RefManSetPhases(p, nil, 0) != 1 {
		t.Fatal("nil care should preserve counterexample")
	}
}
