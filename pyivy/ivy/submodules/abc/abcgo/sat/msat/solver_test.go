package msat

import (
	"bytes"
	"math/rand"
	"strings"
	"testing"
)

func newSolver(n int) *Msat_Solver_t {
	p := Msat_SolverAlloc(n, 1, 1, 1, 1, 0)
	Msat_SolverClean(p, n)
	return p
}
func brute(n int, cs [][]int, assumptions []int) bool {
	for mask := 0; mask < 1<<n; mask++ {
		ok := true
		for _, lit := range assumptions {
			if ((mask>>(lit>>1))&1)^(lit&1) == 0 {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for _, c := range cs {
			sat := false
			for _, lit := range c {
				if ((mask>>(lit>>1))&1)^(lit&1) == 1 {
					sat = true
				}
			}
			if !sat {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
func TestRandomCNFAndAssumptions(t *testing.T) {
	rng := rand.New(rand.NewSource(432))
	for run := 0; run < 1000; run++ {
		n := 1 + rng.Intn(8)
		p := newSolver(n)
		var cs [][]int
		for k := 0; k < rng.Intn(30); k++ {
			c := make([]int, 1+rng.Intn(5))
			for j := range c {
				c[j] = rng.Intn(2 * n)
			}
			cs = append(cs, c)
			Msat_SolverAddClause(p, &Msat_IntVec_t{data: append([]int(nil), c...)})
		}
		for trial := 0; trial < 4; trial++ {
			var as []int
			for k := 0; k < rng.Intn(4); k++ {
				as = append(as, rng.Intn(2*n))
			}
			want := brute(n, cs, as)
			got := Msat_SolverSolve(p, &Msat_IntVec_t{data: as}, 0, 0)
			if (got == int(MSAT_TRUE)) != want || got == 0 {
				t.Fatalf("case %d trial %d got %d SAT %v clauses %v assumptions %v", run, trial, got, want, cs, as)
			}
			if got == int(MSAT_TRUE) {
				for _, c := range cs {
					sat := false
					for _, lit := range c {
						if p.pModel[lit>>1] == lit {
							sat = true
						}
					}
					if !sat {
						t.Fatal("invalid model", c, p.pModel)
					}
				}
			}
		}
	}
}

func TestIncrementalClausesAndAssumptionsAgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(715203))
	for trial := 0; trial < 500; trial++ {
		const n = 5
		p := New(n)
		var clauses [][]int
		for step := 0; step < 12; step++ {
			if step%2 == 0 {
				clause := []int{rng.Intn(2 * n), rng.Intn(2 * n), rng.Intn(2 * n)}
				clauses = append(clauses, append([]int(nil), clause...))
				p.AddClause(clause)
			}
			var assumptions []int
			for i := 0; i < rng.Intn(5); i++ {
				assumptions = append(assumptions, rng.Intn(2*n))
			}
			want, got := brute(n, clauses, assumptions), p.Solve(assumptions, 0, 0)
			if (got == MSAT_TRUE) != want || got == MSAT_UNKNOWN {
				t.Fatalf("trial=%d step=%d got=%v want=%v clauses=%v assumptions=%v", trial, step, got, want, clauses, assumptions)
			}
			if got == MSAT_TRUE {
				for _, lit := range assumptions {
					if p.pModel[lit>>1] != lit {
						t.Fatalf("model %v violates assumption %d", p.pModel, lit)
					}
				}
				for _, clause := range clauses {
					satisfied := false
					for _, lit := range clause {
						if p.pModel[lit>>1] == lit {
							satisfied = true
							break
						}
					}
					if !satisfied {
						t.Fatalf("model %v violates clause %v", p.pModel, clause)
					}
				}
			}
		}
	}
}

func TestClauseValidationDoesNotSkipSatisfiedPrefix(t *testing.T) {
	p := New(1)
	if !p.AddClause([]int{0}) {
		t.Fatal("unit rejected")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("out-of-range literal accepted")
			}
		}()
		p.AddClause([]int{0, 4})
	}()
	if p.Solve(nil, 0, 0) != MSAT_TRUE || p.ModelValue(0) != 1 {
		t.Fatal("invalid clause changed formula")
	}
}

func TestInvalidAssumptionsDoNotDiscardModel(t *testing.T) {
	p := New(2)
	if !p.AddClause([]int{0}) || p.Solve(nil, 0, 0) != MSAT_TRUE {
		t.Fatal("initial solve")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("invalid assumption accepted")
			}
		}()
		p.Solve([]int{2, 4}, 0, 0)
	}()
	if p.ModelValue(0) != 1 || p.Solve(nil, 0, 0) != MSAT_TRUE {
		t.Fatal("invalid assumptions changed solver state")
	}
}
func TestPigeonholeLearning(t *testing.T) {
	const pigeons = 5
	const holes = 4
	p := newSolver(pigeons * holes)
	for i := 0; i < pigeons; i++ {
		var c []int
		for h := 0; h < holes; h++ {
			c = append(c, 2*(i*holes+h))
		}
		Msat_SolverAddClause(p, &Msat_IntVec_t{data: c})
	}
	for h := 0; h < holes; h++ {
		for a := 0; a < pigeons; a++ {
			for b := a + 1; b < pigeons; b++ {
				Msat_SolverAddClause(p, &Msat_IntVec_t{data: []int{2*(a*holes+h) + 1, 2*(b*holes+h) + 1}})
			}
		}
	}
	if Msat_SolverSolve(p, nil, 0, 0) != int(MSAT_FALSE) || len(p.vLearned.data) == 0 {
		t.Fatal("learning failed", p.Stats)
	}
}
func TestDIMACSRoundTrip(t *testing.T) {
	p, err := ParseDIMACS(strings.NewReader("p cnf 3 3\n1 -2 0\n2 3 0\n-1 0\n"))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := p.WriteDIMACS(&out); err != nil {
		t.Fatal(err)
	}
	q, err := ParseDIMACS(&out)
	if err != nil || Msat_SolverSolve(q, nil, 0, 0) != int(MSAT_TRUE) {
		t.Fatal(err)
	}
}

func TestRootConflictPersistsAcrossSolves(t *testing.T) {
	p := newSolver(2)
	for _, c := range [][]int{{1, 3}, {0}, {2}} {
		Msat_SolverAddClause(p, &Msat_IntVec_t{data: c})
	}
	for i := 0; i < 3; i++ {
		if Msat_SolverSolve(p, nil, 0, 0) != int(MSAT_FALSE) {
			t.Fatal("root conflict forgotten on solve", i)
		}
	}
}
func TestVectorOwnershipAndOrdering(t *testing.T) {
	a := []int{3, 1}
	p := Msat_IntVecAllocArray(a, 2)
	p.data[0] = 5
	if a[0] != 5 {
		t.Fatal("adoption copied array")
	}
	copy := Msat_IntVecDup(p)
	copy.data[0] = 9
	if a[0] != 5 {
		t.Fatal("duplicate aliased array")
	}
	moved := Msat_IntVecDupArray(p)
	if len(p.data) != 0 || moved.data[0] != 5 {
		t.Fatal("transfer")
	}
	Msat_IntVecSort(moved, 0)
	Msat_IntVecPushUniqueOrder(moved, 4, 1)
	Msat_IntVecPushUniqueOrder(moved, 4, 1)
	if len(moved.data) != 3 || moved.data[0] != 1 || moved.data[1] != 4 || moved.data[2] != 5 {
		t.Fatal(moved.data)
	}
}
func TestVariableHeapUpdatesAndDelete(t *testing.T) {
	weights := []float64{1, 8, 3, 4}
	h := Msat_VarHeapAlloc()
	Msat_VarHeapSetActivity(h, weights)
	Msat_VarHeapStart(h, []int{0, 1, 2, 3}, 4, 4)
	Msat_VarHeapCheck(h)
	if Msat_VarHeapGetMax(h) != 1 {
		t.Fatal("max")
	}
	weights[0] = 9
	Msat_VarHeapUpdate(h, 0)
	Msat_VarHeapDelete(h, 3)
	Msat_VarHeapCheck(h)
	if Msat_VarHeapGetMax(h) != 0 || Msat_VarHeapGetMax(h) != 2 {
		t.Fatal("update/delete")
	}
}
func TestPoolsRecycleAndRestart(t *testing.T) {
	p := Msat_MmFixedStart(8)
	a := Msat_MmFixedEntryFetch(p)
	a[0] = 17
	Msat_MmFixedEntryRecycle(p, a)
	b := Msat_MmFixedEntryFetch(p)
	if &a[0] != &b[0] || b[0] != 17 {
		t.Fatal("pool did not reuse entry")
	}
	Msat_MmFixedRestart(p)
	if Msat_MmFixedReadMemUsage(p) != 8 {
		t.Fatal("restart discarded allocation")
	}
	step := Msat_MmStepStart(3)
	x := Msat_MmStepEntryFetch(step, 9)
	Msat_MmStepEntryRecycle(step, x, 9)
	y := Msat_MmStepEntryFetch(step, 9)
	if &x[0] != &y[0] {
		t.Fatal("step pool reuse")
	}
}

func TestTautologyNormalizationRegression(t *testing.T) {
	p := New(6)
	for _, c := range [][]int{{0, 11, 10, 3, 5}, {0, 4, 10}, {1}} {
		if !p.AddClause(c) {
			t.Fatal("tautology or valid clause rejected")
		}
	}
	if p.Solve([]int{11, 2}, 0, 0) != MSAT_TRUE {
		t.Fatal("tautology restricted the CNF")
	}
}

func TestHardRandomThreeSATAgainstExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(3745))
	for run := 0; run < 400; run++ {
		n := 4 + rng.Intn(5)
		p := New(n)
		var cs [][]int
		for c := 0; c < 4*n; c++ {
			vars := rng.Perm(n)[:3]
			lits := make([]int, 3)
			for i, v := range vars {
				lits[i] = 2*v + rng.Intn(2)
			}
			cs = append(cs, lits)
			p.AddClause(lits)
		}
		for trial := 0; trial < 4; trial++ {
			var as []int
			if trial > 0 {
				as = []int{rng.Intn(2 * n)}
			}
			want := brute(n, cs, as)
			got := p.Solve(as, 0, 0)
			if (got == MSAT_TRUE) != want || got == MSAT_UNKNOWN {
				t.Fatalf("run %d trial %d got %d SAT %v clauses %v assumptions %v", run, trial, got, want, cs, as)
			}
		}
	}
}

func TestModelInvalidatedByContradictoryAddition(t *testing.T) {
	p := New(1)
	p.AddClause([]int{0})
	if p.Solve(nil, 0, 0) != MSAT_TRUE || p.ModelValue(0) != 1 {
		t.Fatal("initial model")
	}
	p.AddClause([]int{1})
	if p.ModelValue(0) != -1 || p.Solve(nil, 0, 0) != MSAT_FALSE || p.ModelValue(0) != -1 {
		t.Fatal("stale model after contradiction", p.pModel)
	}
}

func TestRemoveMarkedRestoresUnitAndSimplifiedFormula(t *testing.T) {
	p := New(2)
	p.AddClause([]int{0, 2})
	Msat_SolverMarkClausesStart(p)
	p.AddClause([]int{1})
	if p.Solve(nil, 0, 0) != MSAT_TRUE {
		t.Fatal("incremental formula")
	}
	if Msat_SolverSimplifyDB(p) == 0 {
		t.Fatal("simplification")
	}
	Msat_SolverRemoveMarked(p)
	if p.Solve([]int{0}, 0, 0) != MSAT_TRUE {
		t.Fatal("removed unit still constrained variable")
	}
	if p.Solve([]int{1, 3}, 0, 0) != MSAT_FALSE {
		t.Fatal("simplification erased marked base formula")
	}
}

func TestMarkedRollbackAgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(93932))
	for trial := 0; trial < 250; trial++ {
		const n = 5
		p := New(n)
		var base [][]int
		for i := 0; i < 7; i++ {
			clause := make([]int, 1+rng.Intn(3))
			for j := range clause {
				clause[j] = rng.Intn(2 * n)
			}
			base = append(base, append([]int(nil), clause...))
			p.AddClause(clause)
		}
		Msat_SolverMarkClausesStart(p)
		for i := 0; i < 4; i++ {
			p.AddClause([]int{rng.Intn(2 * n), rng.Intn(2 * n)})
		}
		p.Solve(nil, 0, 0)
		Msat_SolverRemoveMarked(p)
		for query := 0; query < 8; query++ {
			var assumptions []int
			for i := 0; i < rng.Intn(4); i++ {
				assumptions = append(assumptions, rng.Intn(2*n))
			}
			want := brute(n, base, assumptions)
			got := p.Solve(assumptions, 0, 0)
			if (got == MSAT_TRUE) != want || got == MSAT_UNKNOWN {
				t.Fatalf("trial=%d query=%d got=%v want=%v base=%v assumptions=%v", trial, query, got, want, base, assumptions)
			}
		}
	}
}

func TestActivityRescaleBeforeNewClauseInsertion(t *testing.T) {
	p := New(1)
	p.dVarInc = 1e101
	Msat_SolverVarBumpActivity(p, 0)
	if p.pdActivity[0] != 10 || p.dVarInc != 10 {
		t.Fatal("variable rescale", p.pdActivity, p.dVarInc)
	}
	old := &Msat_Clause_t{learned: true, activity: 1e20}
	p.vLearned.data = append(p.vLearned.data, old)
	fresh := &Msat_Clause_t{learned: true}
	p.dClaInc = 1e21
	Msat_SolverClaBumpActivity(p, fresh)
	if fresh.activity != 10 || old.activity > 1.001 || old.activity < 0.999 {
		t.Fatal("new clause missed activity rescale", fresh.activity, old.activity)
	}
	if Msat_SolverReadClauseNum(p) != 1 {
		t.Fatal("learned clause omitted from total clause count")
	}
}
func TestTruthTableSetup(t *testing.T) {
	var truths [6][2]uint32
	Msat_SolverSetupTruthTables(&truths)
	if truths[0][0] != 0xaaaaaaaa || truths[4][0] != 0xffff0000 || truths[5][0] != 0 || truths[5][1] != 0xffffffff {
		t.Fatal(truths)
	}
}
