package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func pdrTestSATManager(t *testing.T, mono bool) (*Pdr_Man_t, aig.Edge) {
	t.Helper()
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(x)
	g.AddOutput(lo.Not())
	aig.Aig_ManSetRegNum(g, 1)
	p := &Pdr_Man_t{pAig: g, pPars: &Pdr_Par_t{NRecycle: 100}, vLits: vec.Vec_IntAlloc(0), vClauses: vec.Vec_VecAlloc(0), vActVars: vec.Vec_IntAlloc(0)}
	if mono {
		p.pPars.FMonoCnf = 1
	}
	if Pdr_ManCreateSolver(p, 0) == nil {
		t.Fatal("failed to create initial frame")
	}
	return p, lo
}

func TestPdrSATInitialCubeAndProperty(t *testing.T) {
	for _, mono := range []bool{false, true} {
		p, lo := pdrTestSATManager(t, mono)
		pos := testPdrSet([]int32{0}, nil)
		neg := testPdrSet([]int32{1}, nil)
		if got := Pdr_ManCheckCubeCs(p, 0, pos); got != 1 {
			t.Fatalf("mono=%v positive initial cube got %d", mono, got)
		}
		if got := Pdr_ManCheckCubeCs(p, 0, neg); got != 0 {
			t.Fatalf("mono=%v negative initial cube got %d", mono, got)
		}
		if got := Pdr_ManCheckCube(p, 0, nil, nil, 0, 0, 0); got != 0 {
			t.Fatalf("mono=%v unconstrained property got %d", mono, got)
		}
		variable := Pdr_ObjSatVar(p, 0, 3, lo)
		if v, ok := p.vSolvers[0].Value(variable); !ok || v {
			t.Fatalf("mono=%v initial latch value=%v valid=%v", mono, v, ok)
		}
	}
}

func TestPdrSATRecyclesAndRestoresFrameClauses(t *testing.T) {
	p, _ := pdrTestSATManager(t, false)
	if Pdr_ManCreateSolver(p, 1) == nil {
		t.Fatal("failed to create frame one")
	}
	cube := testPdrSet([]int32{0}, nil)
	vec.Vec_VecPush(p.vClauses, 0, cube)
	Pdr_ManSolverAddClause(p, 0, cube)
	p.pPars.NRecycle = 1
	vec.Vec_IntWriteEntry(p.vActVars, 0, 1)
	solver := Pdr_ManFetchSolver(p, 0)
	if solver == nil || p.nStarts != 1 || vec.Vec_IntEntry(p.vActVars, 0) != 0 {
		t.Fatal("recycle state")
	}
	if got := Pdr_ManCheckCubeCs(p, 0, cube); got != 1 {
		t.Fatalf("recycled clause result %d", got)
	}
	if solver.Solve() != bsat.Satisfiable {
		t.Fatal("recycled solver unusable")
	}
}

// C's Pdr_ManFetchSolver rebuilds property constraints and clauses from the
// recycled frame onward, but does not import clauses from earlier frames.
func TestPdrRecycleRestoresPropertyAndCurrentAndLaterFrames(t *testing.T) {
	for _, mono := range []bool{false, true} {
		g := aig.NewGraph()
		input := g.NewInput()
		for i := 0; i < 3; i++ {
			g.NewInput()
		}
		g.AddOutput(input)
		for i := 0; i < 3; i++ {
			g.AddOutput(g.Const0())
		}
		aig.Aig_ManSetRegNum(g, 3)
		var pars Pdr_Par_t
		Pdr_ManSetDefaultParams(&pars)
		pars.NRecycle, pars.FMonoCnf = 1, pdrBool(mono)
		p := Pdr_ManStart(g, &pars, nil)
		for frame := int32(0); frame < 3; frame++ {
			if Pdr_ManCreateSolver(p, frame) == nil {
				t.Fatal("could not create frame")
			}
			vec.Vec_VecPush(p.vClauses, int(frame), testPdrSet([]int32{2 * frame}, nil))
		}
		old := p.vSolvers[1]
		vec.Vec_IntWriteEntry(p.vActVars, 1, 1)
		if Pdr_ManFetchSolver(p, 1) != old || p.nStarts != 1 || vec.Vec_IntEntry(p.vActVars, 1) != 0 {
			t.Fatalf("mono=%v solver was not recycled", mono)
		}
		if got := Pdr_ManCheckCube(p, 1, nil, nil, 0, 0, 0); got != 1 {
			t.Fatalf("mono=%v property constraint was not restored: %d", mono, got)
		}
		for i, want := range []int32{0, 1, 1} {
			if got := Pdr_ManCheckCubeCs(p, 1, testPdrSet([]int32{int32(2 * i)}, nil)); got != want {
				t.Fatalf("mono=%v cube %d result=%d want=%d", mono, i, got, want)
			}
		}
		Pdr_ManStop(p)
	}
}

func TestPdrSATCubeLiteralConversion(t *testing.T) {
	for _, mono := range []bool{false, true} {
		p, _ := pdrTestSATManager(t, mono)
		Pdr_ManCreateSolver(p, 1)
		cube := testPdrSet([]int32{0}, nil)
		current := vec.Vec_IntEntry(Pdr_ManCubeToLits(p, 1, cube, 0, 0), 0)
		complement := vec.Vec_IntEntry(Pdr_ManCubeToLits(p, 1, cube, 1, 0), 0)
		if complement != current^1 {
			t.Fatalf("mono=%v complement %d, want %d", mono, complement, current^1)
		}
		next := vec.Vec_IntEntry(Pdr_ManCubeToLits(p, 1, cube, 0, 1), 0)
		got := Pdr_ManLitsToCube(p, 1, []int32{next}, 1)
		if vec.Vec_IntSize(got) != 1 || vec.Vec_IntEntry(got, 0) != 1 {
			t.Fatalf("mono=%v recovered cube %v", mono, vec.Vec_IntArray(got))
		}
	}
}

func TestPdrSATPropertyConstraint(t *testing.T) {
	for _, mono := range []bool{false, true} {
		p, _ := pdrTestSATManager(t, mono)
		Pdr_ManCreateSolver(p, 1)
		p.pPars.FUsePropOut = 1
		Pdr_ManSetPropertyOutput(p, 1)
		if got := Pdr_ManCheckCube(p, 1, nil, nil, 0, 0, 0); got != 1 {
			t.Fatalf("mono=%v constrained property result %d, want UNSAT", mono, got)
		}
	}
}

func TestPdrSATPrepareModelMaterializesStructuralCone(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(x, y)
	root := g.AND(inner, z)
	g.AddOutput(root)
	aig.Aig_ManSetRegNum(g, 0)
	p := &Pdr_Man_t{pAig: g, pPars: &Pdr_Par_t{NRecycle: 100}, vLits: vec.Vec_IntAlloc(0), vClauses: vec.Vec_VecAlloc(0), vActVars: vec.Vec_IntAlloc(0)}
	Pdr_ManCreateSolver(p, 0)
	// Represent the root as a three-input cut. Its intermediate AND is then
	// absent from the lazy SAT cone, although ternary simulation may read it.
	cnf := p.pCnf2
	id := root.NodeID()
	v := func(e aig.Edge) int32 { return cnf.VarNums[e.NodeID()] << 1 }
	cnf.Obj2Clause[id] = int32(len(cnf.Clauses))
	cnf.Obj2Count[id] = 4
	cnf.Clauses = append(cnf.Clauses,
		[]int32{v(root) ^ 1, v(x)}, []int32{v(root) ^ 1, v(y)},
		[]int32{v(root) ^ 1, v(z)}, []int32{v(root), v(x) ^ 1, v(y) ^ 1, v(z) ^ 1})
	Pdr_ManCreateSolver(p, 1)
	pdrPrepareModel(p, 1, nil)
	output := Pdr_ObjSatVar(p, 1, 2, aig.Aig_ManCo(g, 0))
	if p.vSolvers[1].Solve(bsat.Positive(output)) != bsat.Satisfiable {
		t.Fatal("expected satisfiable output")
	}
	for _, obj := range []aig.Edge{x, y, z, inner, root, aig.Aig_ManCo(g, 0)} {
		variable := Pdr_ObjSatVar(p, 1, 3, obj)
		value, ok := p.vSolvers[1].Value(variable)
		if !ok || !value {
			t.Fatalf("object %v: value %v, model valid %v", obj, value, ok)
		}
	}
}

func TestPdrSATUnknownGeneralizationIsTreatedAsSAT(t *testing.T) {
	p, _ := pdrTestSATManager(t, false)
	Pdr_ManCreateSolver(p, 1)
	p.pPars.NConfGenLimit = 1
	p.timeToStop = util.Abc_Clock() - 1
	if got := Pdr_ManCheckCube(p, 1, testPdrSet([]int32{0}, nil), nil, 0, 1, 0); got != 0 {
		t.Fatalf("timed out generalization returned %d, want source's SAT result", got)
	}
}
