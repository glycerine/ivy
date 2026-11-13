package glucose2

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestABCAndEncodingTruthTable(t *testing.T) {
	for _, phases := range []struct{ a, b, out bool }{
		{}, {true, false, false}, {false, true, true}, {true, true, true},
	} {
		for a := int32(0); a < 2; a++ {
			for b := int32(0); b < 2; b++ {
				for out := int32(0); out < 2; out++ {
					s := NewABCSolver()
					if !s.AddABCAnd(2, 0, 1, phases.a, phases.b, phases.out) {
						t.Fatal("encoding unexpectedly made the formula inconsistent")
					}
					want := ((a ^ boolBit(phases.a)) & (b ^ boolBit(phases.b))) ^ boolBit(phases.out)
					assumptions := []int32{abcLit(0, a == 0), abcLit(1, b == 0), abcLit(2, out == 0)}
					got := s.SolveABCLiterals(assumptions)
					if (got == GLUCOSE_SAT) != (out == want) {
						t.Fatalf("AND phases %+v, inputs (%d,%d,%d): status %d, want output %d", phases, a, b, out, got, want)
					}
				}
			}
		}
	}
}

func TestABCXorEncodingTruthTable(t *testing.T) {
	for complement := 0; complement < 2; complement++ {
		for a := int32(0); a < 2; a++ {
			for b := int32(0); b < 2; b++ {
				for out := int32(0); out < 2; out++ {
					s := NewABCSolver()
					if !s.AddABCXor(2, 0, 1, complement != 0) {
						t.Fatal("encoding unexpectedly made the formula inconsistent")
					}
					want := a ^ b ^ int32(complement)
					assumptions := []int32{abcLit(0, a == 0), abcLit(1, b == 0), abcLit(2, out == 0)}
					got := s.SolveABCLiterals(assumptions)
					if (got == GLUCOSE_SAT) != (out == want) {
						t.Fatalf("XOR complement %d, inputs (%d,%d,%d): status %d, want output %d", complement, a, b, out, got, want)
					}
				}
			}
		}
	}
}

func TestMinimizeABCAssumptionsRetainsFixedPrefix(t *testing.T) {
	s := NewABCSolver()
	s.EnsureABCVariables(3)
	// a and b may not both be true. c is unrelated and should be removed.
	if !s.AddABCClause([]int32{abcLit(0, true), abcLit(1, true)}) {
		t.Fatal("could not add base clause")
	}
	input := []int32{abcLit(0, false), abcLit(1, false), abcLit(2, false)}
	got := s.MinimizeABCAssumptions(input, 1)
	want := []int32{input[0], input[1]}
	if !equalABCInts(got, want) {
		t.Fatalf("minimized assumptions %v, want %v", got, want)
	}
	if !equalABCInts(input, []int32{abcLit(0, false), abcLit(1, false), abcLit(2, false)}) {
		t.Fatalf("input assumptions were mutated: %v", input)
	}
}

func TestSolveABCLiteralsReturnsStatusAndConflictCopy(t *testing.T) {
	s := NewABCSolver()
	s.EnsureABCVariables(1)
	if !s.AddABCClause([]int32{abcLit(0, false)}) {
		t.Fatal("could not add unit clause")
	}
	if got := s.SolveABCLiterals([]int32{abcLit(0, true)}); got != GLUCOSE_UNSAT {
		t.Fatalf("status = %d, want UNSAT", got)
	}
	core := s.ABCConflict()
	if len(core) != 1 || core[0] != 0 {
		t.Fatalf("conflict = %v, want the solver's unit conflict clause", core)
	}
	core[0] = 123
	if int32(s.Conflict[0]) != 0 {
		t.Fatal("ABCConflict returned storage owned by the solver")
	}
}

func boolBit(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func equalABCInts(a, b []int32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestMinimizeABCAssumptionsRemovesFinalRedundantLiteral(t *testing.T) {
	s := NewABCSolver()
	s.EnsureABCVariables(2)
	if !s.AddABCClause([]int32{1}) {
		t.Fatal("unexpected contradiction")
	}
	// Fixed x contradicts the permanent clause; y is unnecessary even when it
	// is the sole removable assumption.
	got := s.MinimizeABCAssumptions([]int32{0, 2}, 1)
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("got %v; want fixed assumption only", got)
	}
}

// Graph loading must register each clause exactly once with the simplifier.
func TestSolverFromAigRegistersOccurrencesOnce(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.AND(a, b))
	s := NewSimpSolver()
	Glucose2_SolverFromAig(graph, s)
	counts := make([]int, len(s.literalOccurrences))
	perVar := make([]int, s.NVars())
	for _, c := range s.clauses {
		for _, lit := range c.Lits {
			counts[lit]++
			perVar[lit.Var()]++
		}
	}
	for lit, want := range counts {
		if got := s.literalOccurrences[lit]; got != want {
			t.Fatalf("literal %d: occurrences = %d, want %d", lit, got, want)
		}
	}
	for v, want := range perVar {
		if got := len(s.occurs[v]); got != want {
			t.Fatalf("variable %d: occurrence list = %d, want %d", v, got, want)
		}
	}
	if !s.Eliminate(true) || s.SolveLimited(nil, false) != LTrue {
		t.Fatal("satisfiable AND graph lost during preprocessing")
	}
	model := make([]bool, graph.InputCount())
	for i, input := range graph.Inputs() {
		model[i] = s.Model[input.NodeID()] == LTrue
	}
	if !model[0] || !model[1] {
		t.Fatalf("bad AND model %v", model)
	}
}

func TestInterruptedEquivalenceDoesNotProveEquality(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	s := NewABCSolver()
	s.Stop = func() bool { return true }
	if got := bmcg2_sat_solver_equiv_overlap_check(s, graph, int32(a), int32(b), 1); got != 0 {
		t.Fatalf("interrupted check incorrectly proved equality: %d", got)
	}
}

func TestInterruptedQuantificationIsNotAConstant(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	root := graph.AND(a, b)
	solvers := [2]*SimpSolver{NewABCSolver(), NewABCSolver()}
	solvers[0].Stop = func() bool { return true }
	keep := func(_ any, i int32) int32 {
		if i == 0 {
			return 1
		}
		return 0
	}
	result := bmcg2_sat_solver_quantify(solvers, graph, int32(root), 0, keep, nil, nil)
	// Exists b: a AND b equals a, not constant one.
	if result != int32(a) {
		t.Fatalf("quantified result=%d, want input %d", result, a)
	}
	var divisors []int32
	if got := bmcg2_sat_solver_quantify(solvers, graph, int32(root), 0, keep, nil, &divisors); got != -1 {
		t.Fatalf("interrupted divisor request returned %d, want failure", got)
	}
}

func TestGenerateCubesRejectsOverlapAndInterruptedOffset(t *testing.T) {
	for _, interrupted := range []bool{false, true} {
		offset, onset := NewABCSolver(), NewABCSolver()
		offset.EnsureABCVariables(1)
		onset.EnsureABCVariables(1)
		if interrupted {
			offset.Stop = func() bool { return true }
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("interrupted=%v: expected rejection", interrupted)
				}
			}()
			Glucose2_GenerateCubes([2]*SimpSolver{offset, onset}, []int32{0}, []int32{0}, 0)
		}()
	}
}
