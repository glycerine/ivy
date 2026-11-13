package psat

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func brute(variables int, clauses [][]int, as []int) bool {
	for mask := 0; mask < 1<<variables; mask++ {
		ok := true
		for _, lit := range as {
			if ((mask>>(lit>>1))&1)^(lit&1) == 0 {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for _, c := range clauses {
			sat := false
			for _, lit := range c {
				if ((mask>>(lit>>1))&1)^(lit&1) == 1 {
					sat = true
					break
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
func TestRandomSATAndResolutionProofs(t *testing.T) {
	rng := rand.New(rand.NewSource(831))
	for run := 0; run < 600; run++ {
		variables := 1 + rng.Intn(7)
		s := New(true)
		s.SetVariableCount(variables)
		var roots [][]int
		count := rng.Intn(20)
		for c := 0; c < count; c++ {
			lits := make([]int, 1+rng.Intn(4))
			for i := range lits {
				lits[i] = rng.Intn(2 * variables)
			}
			roots = append(roots, lits)
			s.AddClause(lits)
		}
		for trial := 0; trial < 3; trial++ {
			var as []int
			for i := 0; i < rng.Intn(4); i++ {
				as = append(as, rng.Intn(2*variables))
			}
			want := brute(variables, roots, as)
			got := s.Solve(as, 0)
			if (got == Satisfiable) != want || got == Undecided {
				t.Fatalf("run %d trial %d got %d wantSAT %v roots %v assumptions %v", run, trial, got, want, roots, as)
			}
			if got == Unsatisfiable {
				if !s.ProofReady() {
					t.Fatal("missing proof")
				}
				if err := s.Proof().Validate(); err != nil {
					t.Fatal(err)
				}
			} else {
				for _, clause := range append(append([][]int(nil), roots...), asClauses(as)...) {
					sat := false
					for _, lit := range clause {
						if s.ModelValue(lit>>1)^(lit&1) == 1 {
							sat = true
						}
					}
					if !sat {
						t.Fatal("invalid model", clause, s.model)
					}
				}
			}
		}
	}
}

func TestIncrementalQueriesAndProofAnchoring(t *testing.T) {
	rng := rand.New(rand.NewSource(612057))
	for trial := 0; trial < 300; trial++ {
		const variables = 5
		s := New(true)
		s.SetVariableCount(variables)
		var roots [][]int
		for step := 0; step < 10; step++ {
			if step%2 == 0 {
				clause := []int{rng.Intn(2 * variables), rng.Intn(2 * variables), rng.Intn(2 * variables)}
				roots = append(roots, append([]int(nil), clause...))
				s.AddClause(clause)
			}
			var assumptions []int
			for i := 0; i < rng.Intn(5); i++ {
				assumptions = append(assumptions, rng.Intn(2*variables))
			}
			want, got := brute(variables, roots, assumptions), s.Solve(assumptions, 0)
			if (got == Satisfiable) != want || got == Undecided {
				t.Fatalf("trial=%d step=%d got=%d want=%v", trial, step, got, want)
			}
			if got == Unsatisfiable {
				if err := s.Proof().ValidateAgainst(roots, assumptions); err != nil {
					t.Fatalf("trial=%d step=%d proof: %v", trial, step, err)
				}
			} else {
				for _, clause := range append(append([][]int(nil), roots...), asClauses(assumptions)...) {
					valid := false
					for _, lit := range clause {
						if s.ModelValue(lit>>1)^(lit&1) == 1 {
							valid = true
							break
						}
					}
					if !valid {
						t.Fatalf("trial=%d step=%d invalid model %v clause %v", trial, step, s.model, clause)
					}
				}
			}
		}
	}
}
func asClauses(as []int) [][]int {
	out := make([][]int, len(as))
	for i, lit := range as {
		out[i] = []int{lit}
	}
	return out
}
func TestProofTraversalMatchesInterpolationContract(t *testing.T) {
	s := M114p_SolverNew(1)
	M114p_SolverSetVarNum(s, 2)
	roots := [][]int{{0, 2}, {1, 2}, {0, 3}, {1, 3}}
	for _, root := range roots {
		if M114p_SolverAddClause(s, root, len(root)) == 0 {
			t.Fatal("premature contradiction")
		}
	}
	if M114p_SolverSolve(s, nil, 0, 0) != Unsatisfiable || M114p_SolverProofIsReady(s) == 0 {
		t.Fatal("UNSAT proof missing")
	}
	var clauses [][]int
	var lits []int
	for n := M114p_SolverGetFirstRoot(s, &lits); n != 0; n = M114p_SolverGetNextRoot(s, &lits) {
		clauses = append(clauses, append([]int(nil), lits...))
	}
	if len(clauses) != len(roots) {
		t.Fatal("root order/count")
	}
	var parents, pivots []int
	for n := M114p_SolverGetFirstChain(s, &parents, &pivots); n != 0; n = M114p_SolverGetNextChain(s, &parents, &pivots) {
		if len(parents) != n+1 {
			t.Fatal("chain arity")
		}
		current := clauses[parents[0]]
		for i, pivot := range pivots {
			var err error
			current, err = Resolve(current, clauses[parents[i+1]], pivot)
			if err != nil {
				t.Fatal(err)
			}
		}
		clauses = append(clauses, current)
	}
	if len(clauses) != M114p_SolverProofClauseNum(s) || len(clauses[len(clauses)-1]) != 0 {
		t.Fatal("last interpolant clause is not empty", clauses)
	}
}
func TestProofRoundTripAndOwnership(t *testing.T) {
	s := New(true)
	s.AddClause([]int{0})
	s.AddClause([]int{1})
	if s.Solve(nil, 0) != Unsatisfiable {
		t.Fatal("expected UNSAT")
	}
	proof := s.Proof()
	var b bytes.Buffer
	if err := proof.Write(&b); err != nil {
		t.Fatal(err)
	}
	restored, err := ReadProof(&b)
	if err != nil || !reflect.DeepEqual(restored, proof) {
		t.Fatal(err, restored, proof)
	}
	proof.Roots[0][0] = 8
	if s.Proof().Roots[0][0] != 0 {
		t.Fatal("proof aliases producer")
	}
	M114p_SolverDelete(s)
	M114p_SolverDelete(s)
	if s.ProofReady() {
		t.Fatal("deleted solver retains proof")
	}
}
func TestInvalidProofsRejected(t *testing.T) {
	for _, p := range []*Proof{{Version: 1, Roots: [][]int{{0}, {1}}, Chains: []Chain{{Clauses: []int{0, 2}, Pivots: []int{0}}}, Final: 2}, {Version: 1, Roots: [][]int{{0}, {2}}, Chains: []Chain{{Clauses: []int{0, 1}, Pivots: []int{0}}}, Final: 2}, {Version: 1, Roots: [][]int{{0}}, Final: 0}} {
		if p.Validate() == nil {
			t.Fatal("accepted invalid proof", p)
		}
	}
	if _, err := Resolve([]int{0, 1}, []int{0}, 0); err == nil {
		t.Fatal("ambiguous pivot accepted")
	}
}
func TestLimitsAndProofMode(t *testing.T) {
	s := New(true)
	for _, root := range [][]int{{0, 2}, {1, 2}, {0, 3}, {1, 3}} {
		s.AddClause(root)
	}
	if s.Solve(nil, 1) != Undecided || s.ProofReady() {
		t.Fatal("work limit ignored")
	}
	if s.Solve(nil, 0) != Unsatisfiable || !s.ProofReady() {
		t.Fatal("resume failed")
	}
	q := New(false)
	q.AddClause([]int{0})
	q.AddClause([]int{1})
	if q.Solve(nil, 0) != Unsatisfiable || q.ProofReady() {
		t.Fatal("proof mode ignored")
	}
}

func TestPigeonholeProof(t *testing.T) {
	const pigeons = 4
	const holes = 3
	s := New(true)
	for p := 0; p < pigeons; p++ {
		var clause []int
		for h := 0; h < holes; h++ {
			clause = append(clause, 2*(p*holes+h))
		}
		s.AddClause(clause)
	}
	for h := 0; h < holes; h++ {
		for a := 0; a < pigeons; a++ {
			for b := a + 1; b < pigeons; b++ {
				s.AddClause([]int{2*(a*holes+h) + 1, 2*(b*holes+h) + 1})
			}
		}
	}
	if s.Solve(nil, 0) != Unsatisfiable || !s.ProofReady() {
		t.Fatal("missing pigeonhole refutation")
	}
	if len(s.Proof().Chains) == 0 {
		t.Fatal("no resolution derivations")
	}
	if err := s.Proof().ValidateAgainst(s.roots, nil); err != nil {
		t.Fatal(err)
	}
}

func TestProofSaveFile(t *testing.T) {
	s := New(true)
	s.AddClause([]int{0})
	s.AddClause([]int{1})
	s.Solve(nil, 0)
	path := filepath.Join(t.TempDir(), "proof.json")
	M114p_SolverProofSave(s, path)
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	proof, err := ReadProof(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := proof.ValidateAgainst(s.roots, nil); err != nil {
		t.Fatal(err)
	}
}
