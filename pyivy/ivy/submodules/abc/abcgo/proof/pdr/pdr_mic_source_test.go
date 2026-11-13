package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/hash"
	"github.com/glycerine/abcgo/misc/util"
)

// C returns immediately on an inconclusive MIC query, before restoring the
// removed literal to the caller's cube.
func TestSimpleMicTimeoutLeavesAttemptedLiteralRemovedAsInC(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.NewInput()
	g.NewInput()
	g.AddOutput(x)
	g.AddOutput(x)
	g.AddOutput(y)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	p := Pdr_ManStart(g, &pars, nil)
	defer Pdr_ManStop(p)
	Pdr_ManCreateSolver(p, 0)
	Pdr_ManCreateSolver(p, 1)
	p.timeToStop = util.Abc_Clock() - 1
	cube := testPdrSet([]int32{0, 2}, nil)
	if got := ZPdr_ManSimpleMic(p, 1, &cube); got != -1 {
		t.Fatalf("expired query returned %d", got)
	}
	if cube.Lits[0] != -1 && cube.Lits[1] != -1 {
		t.Fatalf("source leaves the attempted literal removed: %v", cube.Lits)
	}
}

// In C an inconclusive CTG check breaks out of the CTG loop; it then joins
// with the predecessor, and this keep-set makes the join fail with result 0.
func TestDownCtgTimeoutContinuesToJoinAsInC(t *testing.T) {
	g := aig.NewGraph()
	inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
	for range inputs {
		g.NewInput()
	}
	g.AddOutput(g.Const0())
	for _, input := range inputs {
		g.AddOutput(input)
	}
	aig.Aig_ManSetRegNum(g, 3)
	var pars Pdr_Par_t
	Pdr_ManSetDefaultParams(&pars)
	pars.fCtgs = 1
	p := Pdr_ManStart(g, &pars, nil)
	defer Pdr_ManStop(p)
	for i := int32(0); i < 3; i++ {
		Pdr_ManCreateSolver(p, i)
	}
	p.timeToStop = util.Abc_Clock() - 1
	cube := testPdrSet([]int32{0, 2}, nil)
	pred := testPdrSet([]int32{0, 4}, nil)
	keep := hash.Hash_IntAlloc(5)
	hash.Hash_IntWriteEntry(keep, 2, 1)
	added := int32(1)
	if got := ZPdr_ManDown(p, 2, &cube, pred, keep, nil, &added); got != 0 || cube != nil {
		t.Fatalf("got result=%d cube=%v, want failed join", got, cube)
	}
}
