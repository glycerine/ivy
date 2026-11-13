package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

// fraSim.c packs solver values by reduced-graph CI position, then projects
// first-frame PIs and the reduced graph's final register inputs into vCex.
func TestSavePatternPacksSolverModelAndProjectsCounterexample(t *testing.T) {
	original := aig.NewGraph()
	for i := 0; i < 5; i++ {
		original.NewInput()
	}
	original.AddOutput(original.Const0())
	original.AddOutput(original.Const0())
	aig.Aig_ManSetRegNum(original, 2)
	reduced := aig.NewGraph()
	for i := 0; i < 40; i++ {
		reduced.NewInput()
	}
	reduced.AddOutput(reduced.Const0())
	reduced.AddOutput(reduced.Const0())
	aig.Aig_ManSetRegNum(reduced, 2)
	solver := bsat.NewSolver()
	variables := make([]int32, 40)
	for i := range variables {
		variables[i] = solver.NewVariable()
	}
	assumptions := make([]bsat.Literal, 40)
	for i := range assumptions {
		v := variables[39-i]
		if i%3 == 0 {
			assumptions[i] = bsat.Positive(v)
		} else {
			assumptions[i] = bsat.Negative(v)
		}
	}
	if status := solver.Solve(assumptions...); status != bsat.Satisfiable {
		t.Fatalf("solve = %v", status)
	}
	p := &Fra_Man_t{PManAig: original, PManFraig: reduced, PSat: solver, PMemSatNums: make([]int32, len(reduced.Nodes())), PPatWords: []uint32{^uint32(0), ^uint32(0)}, VCex: vec.Vec_IntAlloc(1)}
	vec.Vec_IntPush(p.VCex, 9)
	for i, input := range reduced.Inputs() {
		Fra_ObjSetSatNum(p, input, variables[39-i])
	}
	Fra_SmlSavePattern(p)
	var want [2]uint32
	for i := 0; i < 40; i++ {
		if i%3 == 0 {
			want[i/32] |= uint32(1) << uint(i%32)
		}
	}
	if p.PPatWords[0] != want[0] || p.PPatWords[1] != want[1] {
		t.Fatalf("pattern = %#v, want %#v", p.PPatWords, want)
	}
	expected := []int32{1, 0, 0, 0, 1}
	got := vec.Vec_IntArray(p.VCex)
	if len(got) != len(expected) {
		t.Fatalf("cex = %v", got)
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("cex = %v, want %v", got, expected)
		}
	}
	p.VCex = nil
	p.PPatWords[1] = ^uint32(0)
	Fra_SmlSavePattern(p)
	if p.PPatWords[1] != want[1] {
		t.Fatalf("nil vCex pattern = %#v, want %#v", p.PPatWords, want)
	}
}
