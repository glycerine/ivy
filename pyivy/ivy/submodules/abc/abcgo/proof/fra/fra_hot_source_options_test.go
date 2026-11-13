package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestOneHotKnownConstraintSourceTruthTable(t *testing.T) {
	g := aig.NewGraph()
	inputs := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
	p := &Fra_Man_t{PManFraig: g, PSat: bsat.NewSolver(), PMemSatNums: make([]int32, len(g.Nodes()))}
	for i, x := range inputs {
		p.PMemSatNums[x.NodeID()] = int32(i + 1)
	}
	groups := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(groups, vec.Vec_IntAllocArrayCopy([]int32{0, 1, 2}, 3))
	Fra_OneHotAddKnownConstraint(p, groups)
	for mask := 0; mask < 8; mask++ {
		lits := make([]bsat.Literal, 3)
		for i := range lits {
			if mask&(1<<i) != 0 {
				lits[i] = bsat.Positive(int32(i + 1))
			} else {
				lits[i] = bsat.Negative(int32(i + 1))
			}
		}
		got := p.PSat.Solve(lits...) == bsat.Satisfiable
		want := mask == 0 || mask == 1 || mask == 2 || mask == 4
		if got != want {
			t.Fatalf("mask %03b: got %v want %v", mask, got, want)
		}
	}
}

func TestOneHotAssumeContradictionClearsSolver(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 2)
	p := &Fra_Man_t{PManFraig: g, PPars: &Fra_Par_t{NFramesK: 1}, PSat: bsat.NewSolver(), PMemSatNums: make([]int32, len(g.Nodes()))}
	p.PMemSatNums[a.NodeID()] = 1
	p.PMemSatNums[b.NodeID()] = 2
	if !p.PSat.AddClause(bsat.Positive(1)) || !p.PSat.AddClause(bsat.Positive(2)) {
		t.Fatal("fixture contradiction")
	}
	clauses := vec.Vec_IntAllocArrayCopy([]int32{0, 0, Fra_RegToLit(0, 1), Fra_RegToLit(1, 1)}, 4)
	Fra_OneHotAssume(p, clauses)
	if p.PSat != nil {
		t.Fatal("C discards the contradictory solver")
	}
}

func TestOneHotKnownConstraintContradictionClearsSolver(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := &Fra_Man_t{PManFraig: g, PSat: bsat.NewSolver(), PMemSatNums: make([]int32, len(g.Nodes()))}
	p.PMemSatNums[a.NodeID()] = 1
	p.PMemSatNums[b.NodeID()] = 2
	if !p.PSat.AddClause(bsat.Positive(1)) || !p.PSat.AddClause(bsat.Positive(2)) {
		t.Fatal("fixture contradiction")
	}
	groups := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(groups, vec.Vec_IntAllocArrayCopy([]int32{0, 1}, 2))
	Fra_OneHotAddKnownConstraint(p, groups)
	if p.PSat != nil {
		t.Fatal("C discards the contradictory solver")
	}
}

func TestOneHotCheckSkipsRemovedClauses(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	g.AddOutput(a)
	aig.Aig_ManSetRegNum(g, 1)
	p := &Fra_Man_t{PManFraig: g, PCla: Fra_ClassesStart(g)}
	Fra_OneHotCheck(p, vec.Vec_IntAllocArrayCopy([]int32{0, 0}, 2))
	if p.PCla.FRefinement != 0 {
		t.Fatal("removed clause changed refinement state")
	}
}

// A one-step machine can move from a one-hot state to a non-one-hot state.
// The source simulation must consume the SAT counterexample and delete the
// disproved clause rather than claiming an inductive invariant.
func oneHotCheckFixture() *Fra_Man_t {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.NewInput()
	g.NewInput()
	g.AddOutput(x)
	g.AddOutput(y)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	pars.NFramesK = 1
	pars.FUse1Hot = 1
	p := Fra_ManStart(g, &pars)
	p.PSml = Fra_SmlStart(g, 0, 2, 2)
	p.PManFraig = Fra_FramesWithClasses(p)
	data := cnf.Cnf_Derive(p.PManFraig, p.PManFraig.RegisterCount())
	p.PSat = cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	p.NSatVars = int32(data.Variables)
	Fra_ManClean(p, int32(len(p.PManFraig.Nodes())))
	for _, edge := range p.PManFraig.LiveEdges() {
		if n := data.VarNums[edge.NodeID()]; n >= 0 {
			Fra_ObjSetSatNum(p, edge, n)
			Fra_ObjSetFaninVec(p, edge, []aig.Edge{})
		}
	}
	p.VOneHots = vec.Vec_IntAllocArrayCopy([]int32{Fra_RegToLit(0, 1), Fra_RegToLit(1, 1)}, 2)
	Fra_OneHotAssume(p, p.VOneHots)
	return p
}

func TestOneHotCheckRefinesSATCounterexample(t *testing.T) {
	p := oneHotCheckFixture()
	Fra_OneHotCheck(p, p.VOneHots)
	if p.PCla.FRefinement != 1 || Fra_OneHotCount(p, p.VOneHots) != 0 || p.NSatCallsSat != 1 {
		t.Fatalf("refinement=%d clauses=%v SAT=%d", p.PCla.FRefinement, vec.Vec_IntArray(p.VOneHots), p.NSatCallsSat)
	}
}

func TestOneHotCheckUndefinedAsserts(t *testing.T) {
	p := oneHotCheckFixture()
	p.PPars.NBTLimitNode = -1
	defer func() {
		if recover() == nil {
			t.Error("C asserts when an undefined query leaves the clause unrefined")
		}
		if p.PCla.FRefinement != 1 || p.NSatFailsReal != 1 || Fra_OneHotCount(p, p.VOneHots) != 1 {
			t.Errorf("refinement=%d failures=%d clauses=%v", p.PCla.FRefinement, p.NSatFailsReal, vec.Vec_IntArray(p.VOneHots))
		}
	}()
	Fra_OneHotCheck(p, p.VOneHots)
}
