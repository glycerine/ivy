package llb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCounterexampleTraceReconstructsAndReplays(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state.Not())
	aig.Aig_ManSetRegNum(g, 1)
	states := vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(states, []uint32{0})
	vec.Vec_PtrPush(states, []uint32{1})
	cex := Llb4_Nonlin4TransformCex(g, states, 0, 0)
	if cex == nil {
		t.Fatal("expected a valid trace")
	}
	if replay := Llb4_Nonlin4VerifyCex(g, cex); replay == nil || vec.Vec_PtrSize(replay) != 2 {
		t.Fatal("reconstructed counterexample did not replay")
	}
}
