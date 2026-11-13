package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestSelectHighestCostImplications(t *testing.T) {
	values := vec.Vec_IntAlloc(4)
	for _, v := range []int32{11, 22, 33, 44} {
		vec.Vec_IntPush(values, v)
	}
	var pivot int32
	got := Fra_SmlSelectMaxCost(values, []int32{2, 5, 5, 1}, 5, 2, &pivot)
	if pivot != 5 || vec.Vec_IntSize(got) != 2 || vec.Vec_IntEntry(got, 0) != 22 || vec.Vec_IntEntry(got, 1) != 33 {
		t.Fatalf("selection = %v pivot %d", vec.Vec_IntArray(got), pivot)
	}
}

func TestImplicationConstraintsAcrossFrames(t *testing.T) {
	g := aig.NewGraph()
	left, right := g.NewInput(), g.NewInput()
	p := &Fra_Man_t{pManAig: g, pPars: &Fra_Par_t{nFramesK: 1}, nFramesAll: 2, nSatVars: 3, pSat: bsat.NewSolver(), pMemFraig: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.pMemFraig {
		p.pMemFraig[i] = aig.InvalidEdge
	}
	Fra_ObjSetFraig(p, left, 0, left)
	Fra_ObjSetFraig(p, right, 0, right)
	vars := make([]int32, len(g.Nodes()))
	vars[left.NodeID()] = 1
	vars[right.NodeID()] = 2
	imps := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(imps, Fra_ImpCreate(int32(left.NodeID()), int32(right.NodeID())))
	Fra_ImpAddToSolver(p, imps, vars)
	if p.pSat == nil {
		t.Fatal("consistent implication rejected")
	}
	if got := p.pSat.Solve(bsat.Positive(1), bsat.Negative(2)); got != bsat.Unsatisfiable {
		t.Fatalf("counterexample status = %v", got)
	}
	if got := p.pSat.Solve(bsat.Negative(1), bsat.Negative(2)); got != bsat.Satisfiable {
		t.Fatalf("valid state status = %v", got)
	}
}

func TestImplicationConstraintAllowsSATVariableZero(t *testing.T) {
	g := aig.NewGraph()
	left, right := g.NewInput(), g.NewInput()
	p := &Fra_Man_t{pManAig: g, pPars: &Fra_Par_t{nFramesK: 1}, nFramesAll: 2, nSatVars: 2, pSat: bsat.NewSolver(), pMemFraig: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.pMemFraig {
		p.pMemFraig[i] = aig.InvalidEdge
	}
	Fra_ObjSetFraig(p, left, 0, left)
	Fra_ObjSetFraig(p, right, 0, right)
	vars := make([]int32, len(g.Nodes()))
	vars[left.NodeID()] = 0
	vars[right.NodeID()] = 1
	imps := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(imps, Fra_ImpCreate(int32(left.NodeID()), int32(right.NodeID())))
	Fra_ImpAddToSolver(p, imps, vars)
	if p.pSat == nil {
		t.Fatal("consistent implication rejected")
	}
	if got := p.pSat.Solve(bsat.Positive(0), bsat.Negative(1)); got != bsat.Unsatisfiable {
		t.Fatalf("counterexample status = %v", got)
	}
}
