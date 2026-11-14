package pdr

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func TestPdrDynamicCNFAgainstGraphEvaluation(t *testing.T) {
	rng := rand.New(rand.NewSource(59121))
	for trial := 0; trial < 50; trial++ {
		g := aig.NewGraph()
		pool := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.Const0(), g.Const1()}
		for i := 0; i < 5; i++ {
			a, b := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
			if rng.Intn(2) == 0 {
				a = a.Not()
			}
			if rng.Intn(2) == 0 {
				b = b.Not()
			}
			if rng.Intn(2) == 0 {
				pool = append(pool, g.AND(a, b))
			} else {
				pool = append(pool, g.NewXor(a, b))
			}
		}
		for i := 0; i < 2; i++ {
			g.AddOutput(pool[rng.Intn(len(pool))])
		}
		s := bsat.NewSolver()
		p := &Pdr_Man_t{pAig: g, pPars: &Pdr_Par_t{}, pCnfMan: cnf.Cnf_ManStart(), vSolvers: []*bsat.Solver{nil, s}}
		if Pdr_ManNewSolver(s, p, 1, 0) == nil {
			t.Fatal("frame solver")
		}
		var vars []int32
		for _, edge := range g.Inputs() {
			vars = append(vars, Pdr_ObjSatVar(p, 1, 3, edge))
		}
		for _, edge := range g.Outputs() {
			vars = append(vars, Pdr_ObjSatVar(p, 1, 3, edge))
		}
		for mask := 0; mask < 32; mask++ {
			inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			outputs, err := g.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			want := outputs[0] == (mask&8 != 0) && outputs[1] == (mask&16 != 0)
			assumptions := make([]bsat.Literal, len(vars))
			for i, variable := range vars {
				assumptions[i] = variable << 1
				if mask&(1<<i) == 0 {
					assumptions[i] ^= 1
				}
			}
			if got := s.Solve(assumptions...) == bsat.Satisfiable; got != want {
				t.Fatalf("trial=%d assignment=%05b got=%v want=%v", trial, mask, got, want)
			}
		}
	}
}

func TestPdrFrameCNFEncodesSequentialCone(t *testing.T) {
	for _, mono := range []int32{0, 1} {
		g := aig.NewGraph()
		pi, lo := g.NewInput(), g.NewInput()
		out := g.AddOutput(g.AND(pi, lo))
		li := g.AddOutput(pi.Not())
		aig.Aig_ManSetRegNum(g, 1)
		solver := bsat.NewSolver()
		manager := &Pdr_Man_t{pAig: g, pPars: &Pdr_Par_t{FMonoCnf: mono}, pCnfMan: cnf.Cnf_ManStart(), vSolvers: []*bsat.Solver{nil, solver}}
		if Pdr_ManNewSolver(solver, manager, 1, 0) == nil {
			t.Fatalf("mono=%d: no solver", mono)
		}
		x, state := Pdr_ObjSatVar(manager, 1, 3, pi), Pdr_ObjSatVar(manager, 1, 3, lo)
		y, next := Pdr_ObjSatVar(manager, 1, 3, out), Pdr_ObjSatVar(manager, 1, 3, li)
		if Pdr_ObjRegNum(manager, 1, next) != 0 {
			t.Fatalf("mono=%d: wrong register mapping", mono)
		}
		if mono == 0 && Pdr_ObjSatVar(manager, 1, 3, out) != y {
			t.Fatal("dynamic variable was duplicated")
		}
		for mask := 0; mask < 16; mask++ {
			variables := []int32{x, state, y, next}
			assumptions := make([]bsat.Literal, len(variables))
			for i, variable := range variables {
				assumptions[i] = variable << 1
				if mask&(1<<i) == 0 {
					assumptions[i] ^= 1
				}
			}
			want := (mask&4 != 0) == (mask&1 != 0 && mask&2 != 0) && (mask&8 != 0) == (mask&1 == 0)
			if got := solver.Solve(assumptions...) == bsat.Satisfiable; got != want {
				t.Fatalf("mono=%d assignment=%04b got=%v want=%v", mono, mask, got, want)
			}
		}
		free := Pdr_ManFreeVar(manager, 1)
		if Pdr_ObjRegNum(manager, 1, free) != -1 {
			t.Fatalf("mono=%d: activation variable mapped to register", mono)
		}
	}
}

func TestPdrFrameZeroInitializesLatchAndRecyclesMapping(t *testing.T) {
	g := aig.NewGraph()
	lo := g.NewInput()
	li := g.AddOutput(lo.Not())
	aig.Aig_ManSetRegNum(g, 1)
	s := bsat.NewSolver()
	p := &Pdr_Man_t{pAig: g, pPars: &Pdr_Par_t{}, pCnfMan: cnf.Cnf_ManStart(), vSolvers: []*bsat.Solver{s}}
	if Pdr_ManNewSolver(s, p, 0, 1) == nil {
		t.Fatal("frame initialization")
	}
	variable := Pdr_ObjSatVar(p, 0, 3, lo)
	if s.Solve(bsat.Positive(variable)) != bsat.Unsatisfiable {
		t.Fatal("initial latch was not constrained to zero")
	}
	Pdr_ObjSatVar(p, 0, 3, li)
	old := vec.Vec_IntSize(p.vVar2Ids[0])
	if old < 3 {
		t.Fatal("expected object mappings")
	}
	replacement := bsat.NewSolver()
	p.vSolvers[0] = replacement
	if Pdr_ManNewSolver(replacement, p, 0, 1) == nil || vec.Vec_IntSize(p.vVar2Ids[0]) != 1 {
		t.Fatal("frame recycling retained reverse mappings")
	}
	if Pdr_ObjSatVar(p, 0, 3, lo) != 1 {
		t.Fatal("frame recycling retained stale forward mapping")
	}
}

func TestPdrDynamicConstantHasPositiveUnit(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const1())
	s := bsat.NewSolver()
	p := &Pdr_Man_t{pAig: g, pPars: &Pdr_Par_t{}, pCnfMan: cnf.Cnf_ManStart(), vSolvers: []*bsat.Solver{nil, s}}
	if Pdr_ManNewSolver(s, p, 1, 0) == nil {
		t.Fatal("frame solver")
	}
	variable := Pdr_ObjSatVar(p, 1, 3, g.Const1())
	if s.Solve(bsat.Negative(variable)) != bsat.Unsatisfiable || s.Solve(bsat.Positive(variable)) != bsat.Satisfiable {
		t.Fatal("AIG constant does not have its required SAT value")
	}
}
