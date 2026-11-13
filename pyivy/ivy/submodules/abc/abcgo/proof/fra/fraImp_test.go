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
	p := &Fra_Man_t{PManAig: g, PPars: &Fra_Par_t{NFramesK: 1}, NFramesAll: 2, NSatVars: 3, PSat: bsat.NewSolver(), PMemFraig: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.PMemFraig {
		p.PMemFraig[i] = aig.InvalidEdge
	}
	Fra_ObjSetFraig(p, left, 0, left)
	Fra_ObjSetFraig(p, right, 0, right)
	vars := make([]int32, len(g.Nodes()))
	vars[left.NodeID()] = 1
	vars[right.NodeID()] = 2
	imps := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(imps, Fra_ImpCreate(int32(left.NodeID()), int32(right.NodeID())))
	Fra_ImpAddToSolver(p, imps, vars)
	if p.PSat == nil {
		t.Fatal("consistent implication rejected")
	}
	if got := p.PSat.Solve(bsat.Positive(1), bsat.Negative(2)); got != bsat.Unsatisfiable {
		t.Fatalf("counterexample status = %v", got)
	}
	if got := p.PSat.Solve(bsat.Negative(1), bsat.Negative(2)); got != bsat.Satisfiable {
		t.Fatalf("valid state status = %v", got)
	}
}

func TestImplicationConstraintAllowsSATVariableZero(t *testing.T) {
	g := aig.NewGraph()
	left, right := g.NewInput(), g.NewInput()
	p := &Fra_Man_t{PManAig: g, PPars: &Fra_Par_t{NFramesK: 1}, NFramesAll: 2, NSatVars: 2, PSat: bsat.NewSolver(), PMemFraig: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.PMemFraig {
		p.PMemFraig[i] = aig.InvalidEdge
	}
	Fra_ObjSetFraig(p, left, 0, left)
	Fra_ObjSetFraig(p, right, 0, right)
	vars := make([]int32, len(g.Nodes()))
	vars[left.NodeID()] = 0
	vars[right.NodeID()] = 1
	imps := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(imps, Fra_ImpCreate(int32(left.NodeID()), int32(right.NodeID())))
	Fra_ImpAddToSolver(p, imps, vars)
	if p.PSat == nil {
		t.Fatal("consistent implication rejected")
	}
	if got := p.PSat.Solve(bsat.Positive(0), bsat.Negative(1)); got != bsat.Unsatisfiable {
		t.Fatalf("counterexample status = %v", got)
	}
}

func TestImplicationCompactionAfterSimplificationFailure(t *testing.T) {
	g := aig.NewGraph()
	left, right := g.NewInput(), g.NewInput()
	solver := bsat.NewSolver()
	// Each source-valid addition succeeds. Root propagation then discovers
	// the contradiction between the two implications of a.
	if !solver.AddClause(bsat.Negative(0), bsat.Positive(1)) ||
		!solver.AddClause(bsat.Negative(0), bsat.Negative(1)) ||
		!solver.AddClause(bsat.Positive(0)) {
		t.Fatal("fixture clauses rejected before simplification")
	}
	p := &Fra_Man_t{PManAig: g, PPars: &Fra_Par_t{NFramesK: 1}, NFramesAll: 1, PSat: solver, PMemFraig: make([]aig.Edge, len(g.Nodes()))}
	for i := range p.PMemFraig {
		p.PMemFraig[i] = aig.InvalidEdge
	}
	imps := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(imps, Fra_ImpCreate(int32(left.NodeID()), int32(right.NodeID())))
	Fra_ImpAddToSolver(p, imps, nil)
	if p.PSat != nil || vec.Vec_IntSize(imps) != 0 {
		t.Fatalf("source simplifies then compacts: solver=%v implications=%v", p.PSat, vec.Vec_IntArray(imps))
	}
}

// fraImp.c discards an implication if either endpoint lacks a mapping in
// any induction frame, then compacts the vector after simplification.
func TestImplicationMissingOneFrameIsCompacted(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	p := &Fra_Man_t{PManAig: g, PPars: &Fra_Par_t{NFramesK: 2}, NFramesAll: 2, NSatVars: 4, PSat: bsat.NewSolver(), PMemFraig: make([]aig.Edge, len(g.Nodes())*2)}
	for i := range p.PMemFraig {
		p.PMemFraig[i] = aig.InvalidEdge
	}
	for frame := int32(0); frame < 2; frame++ {
		Fra_ObjSetFraig(p, x, frame, x)
		Fra_ObjSetFraig(p, y, frame, y)
	}
	Fra_ObjSetFraig(p, z, 0, z) // The second frame is missing.
	variables := make([]int32, len(g.Nodes()))
	variables[x.NodeID()], variables[y.NodeID()], variables[z.NodeID()] = 1, 2, 3
	keep := Fra_ImpCreate(int32(x.NodeID()), int32(y.NodeID()))
	imps := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(imps, Fra_ImpCreate(int32(x.NodeID()), int32(z.NodeID())))
	vec.Vec_IntPush(imps, keep)
	Fra_ImpAddToSolver(p, imps, variables)
	if p.PSat == nil || vec.Vec_IntSize(imps) != 1 || vec.Vec_IntEntry(imps, 0) != keep {
		t.Fatalf("unexpected remaining implications: %v", vec.Vec_IntArray(imps))
	}
	if p.PSat.Solve(bsat.Positive(1), bsat.Negative(2)) != bsat.Unsatisfiable {
		t.Fatal("surviving implication was not added")
	}
}

// fraImp.c applies each normalized implication separately in every induction
// frame, accounting for complemented copies of both endpoints.
func TestImplicationConstraintsFramePolarity(t *testing.T) {
	for polarity := 0; polarity < 16; polarity++ {
		source, frames := aig.NewGraph(), aig.NewGraph()
		left, right := source.NewInput(), source.NewInput()
		copies := []aig.Edge{frames.NewInput(), frames.NewInput(), frames.NewInput(), frames.NewInput()}
		p := &Fra_Man_t{PManAig: source, PManFraig: frames, PPars: &Fra_Par_t{NFramesK: 2}, NFramesAll: 2, NSatVars: 4, PSat: bsat.NewSolver(), PMemFraig: make([]aig.Edge, len(source.Nodes())*2)}
		for i := range p.PMemFraig {
			p.PMemFraig[i] = aig.InvalidEdge
		}
		vars := make([]int32, len(frames.Nodes()))
		for i, edge := range copies {
			vars[edge.NodeID()] = int32(i)
			if polarity&(1<<i) != 0 {
				edge = edge.Not()
			}
			if i%2 == 0 {
				Fra_ObjSetFraig(p, left, int32(i/2), edge)
			} else {
				Fra_ObjSetFraig(p, right, int32(i/2), edge)
			}
		}
		imps := vec.Vec_IntAlloc(1)
		vec.Vec_IntPush(imps, Fra_ImpCreate(int32(left.NodeID()), int32(right.NodeID())))
		Fra_ImpAddToSolver(p, imps, vars)
		if p.PSat == nil {
			t.Fatalf("polarity %d: formula unexpectedly inconsistent", polarity)
		}
		for values := 0; values < 16; values++ {
			assumptions := make([]int32, 4)
			for i := range assumptions {
				if values&(1<<i) != 0 {
					assumptions[i] = bsat.Positive(int32(i))
				} else {
					assumptions[i] = bsat.Negative(int32(i))
				}
			}
			want := true
			for frame := 0; frame < 2; frame++ {
				a := ((values ^ polarity) & (1 << (frame * 2))) != 0
				b := ((values ^ polarity) & (1 << (frame*2 + 1))) != 0
				want = want && (!a || b)
			}
			if got := p.PSat.Solve(assumptions...); (got == bsat.Satisfiable) != want {
				t.Fatalf("polarity=%d values=%d status=%v want SAT=%v", polarity, values, got, want)
			}
		}
	}
}

// The original C checks the limit after appending. Thus a zero limit still
// returns all candidates at the maximum cost, rather than an empty vector.
func TestSelectHighestCostImplicationsZeroLimit(t *testing.T) {
	values := vec.Vec_IntAlloc(3)
	for _, v := range []int32{11, 22, 33} {
		vec.Vec_IntPush(values, v)
	}
	var pivot int32
	got := Fra_SmlSelectMaxCost(values, []int32{5, 2, 5}, 5, 0, &pivot)
	if pivot != 5 || vec.Vec_IntSize(got) != 2 || vec.Vec_IntEntry(got, 0) != 11 || vec.Vec_IntEntry(got, 1) != 33 {
		t.Fatalf("pivot=%d selected=%v", pivot, vec.Vec_IntArray(got))
	}
}
