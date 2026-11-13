package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestUniqueRegisterPairInfoSourceBranches(t *testing.T) {
	original := aig.NewGraph()
	regs := make([]aig.Edge, 5)
	for i := range regs {
		regs[i] = original.NewInput()
	}
	for range regs {
		original.AddOutput(original.Const0())
	}
	aig.Aig_ManSetRegNum(original, int32(len(regs)))
	frames := aig.NewGraph()
	x, y := frames.NewInput(), frames.NewInput()
	both := frames.AND(x, y)
	same := frames.AND(both, x)
	pairs := [][2]aig.Edge{{x, x}, {x, x.Not()}, {x, y.Not()}, {both, same}, {x, y}}
	sat := Ssw_SatStart(0)
	defer Ssw_SatStop(sat)
	sat.pAig = frames
	p := &Ssw_Man_t{pAig: original, pFrames: frames, pMSat: sat, pPars: &Ssw_Pars_t{nBTLimit: 1000}, nFrames: 2, iOutputLit: -1, pNodeToFrames: make([]aig.Edge, len(original.Nodes())*2)}
	for i, node := range regs {
		Ssw_ObjSetFrame(p, node, 0, pairs[i][0])
		Ssw_ObjSetFrame(p, node, 1, pairs[i][1])
	}
	Ssw_UniqueRegisterPairInfo(p)
	want := []int32{0, 1, 1, 0, 1}
	for i, v := range want {
		if got := vec.Vec_IntEntry(p.vDiffPairs, i); got != v {
			t.Fatalf("pair %d: got %d want %d", i, got, v)
		}
	}
}

func TestUniqueOneRequiresFeasibleEqualLatchValues(t *testing.T) {
	for _, feasible := range []int32{0, 1} {
		for _, second := range []bool{false, true} {
			original := aig.NewGraph()
			pi, lo := original.NewInput(), original.NewInput()
			first, other := original.AND(pi, lo), original.AND(pi.Not(), lo)
			original.AddOutput(first)
			original.AddOutput(lo)
			aig.Aig_ManSetRegNum(original, 1)
			frames := aig.NewGraph()
			a, b := frames.NewInput(), frames.NewInput()
			sat := Ssw_SatStart(0)
			sat.pAig = frames
			p := &Ssw_Man_t{pAig: original, pFrames: frames, pMSat: sat, pPars: &Ssw_Pars_t{nFramesK: 2}, nFrames: 2, iOutputLit: -1, pNodeToFrames: make([]aig.Edge, len(original.Nodes())*2), vDiffPairs: vec.Vec_IntAllocArrayCopy([]int32{feasible}, 1), vCommon: vec.Vec_PtrAlloc(1)}
			Ssw_ObjSetFrame(p, lo, 0, a)
			Ssw_ObjSetFrame(p, lo, 1, b)
			Ssw_CnfNodeAddToSolver(sat, a)
			Ssw_CnfNodeAddToSolver(sat, b)
			secondLit := sswLit(sat, b, !second)
			if sat.pSat.Solve(sswLit(sat, a, true), secondLit) != bsat.Satisfiable {
				t.Fatal("SAT model unavailable")
			}
			got := Ssw_ManUniqueOne(p, first, other, 0)
			want := int32(0)
			if feasible != 0 && !second {
				want = 1
			}
			if got != want || vec.Vec_PtrSize(p.vCommon) != 1 {
				t.Fatalf("feasible=%d second=%t got=%d support=%d", feasible, second, got, vec.Vec_PtrSize(p.vCommon))
			}
			Ssw_SatStop(sat)
		}
	}
}
