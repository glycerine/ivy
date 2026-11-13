package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

func TestInterpolantStepIdentityTransition(t *testing.T) {
	inter := aig.NewGraph()
	x := inter.NewInput()
	inter.AddOutput(x.Not())
	trans := aig.NewGraph()
	state := trans.NewInput()
	trans.AddOutput(state)
	aig.Aig_ManSetRegNum(trans, 1)
	frames := aig.NewGraph()
	next := frames.NewInput()
	frames.AddOutput(next)
	p := &Inter_Man_t{pInter: inter, pCnfInter: cnf.Cnf_Derive(inter, 0), pAigTrans: trans, pCnfAig: cnf.Cnf_Derive(trans, 1), pFrames: frames, pCnfFrames: cnf.Cnf_Derive(frames, 0), vVarsAB: vec.Vec_IntAlloc(1)}
	beforeInter := append([]int32(nil), p.pCnfInter.VarNums...)
	beforeTrans := append([]int32(nil), p.pCnfAig.VarNums...)
	result := Inter_ManPerformOneStep(p, 0, 0, 0)
	if result != 1 || p.pInterNew == nil {
		t.Fatalf("expected interpolant, got %d graph=%v", result, p.pInterNew)
	}
	for mask := 0; mask < 2; mask++ {
		output, err := p.pInterNew.Evaluate([]bool{mask != 0})
		if err != nil || len(output) != 1 || output[0] != (mask == 0) {
			t.Fatalf("input %d: %v %v", mask, output, err)
		}
	}
	for i, v := range beforeInter {
		if p.pCnfInter.VarNums[i] != v {
			t.Fatal("inter CNF was not restored")
		}
	}
	for i, v := range beforeTrans {
		if p.pCnfAig.VarNums[i] != v {
			t.Fatal("transition CNF was not restored")
		}
	}
	satFrames := aig.NewGraph()
	satState := satFrames.NewInput()
	satFrames.AddOutput(satState.Not())
	p.pFrames, p.pCnfFrames, p.pInterNew = satFrames, cnf.Cnf_Derive(satFrames, 0), nil
	if result := Inter_ManPerformOneStep(p, 1, 0, 0); result != 0 || p.pInterNew != nil {
		t.Fatalf("expected satisfiable interpolation query, got %d graph=%v", result, p.pInterNew)
	}
}
