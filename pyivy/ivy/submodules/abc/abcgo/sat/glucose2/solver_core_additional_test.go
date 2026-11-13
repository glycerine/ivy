package glucose2

import (
	"bytes"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Independently verify that every reported failed-assumption core contains
// only negated assumptions and itself makes the original formula unsatisfiable.
func TestSolverAssumptionCoresAgainstExhaustiveOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(98412))
	checked := 0
	for trial := 0; trial < 1000; trial++ {
		n := 2 + rng.Intn(5)
		s := NewSolver()
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
		}
		var clauses [][]Lit
		for j := rng.Intn(14); j > 0; j-- {
			c := make([]Lit, 1+rng.Intn(3))
			for i := range c {
				c[i] = MkLit(Var(rng.Intn(n)), rng.Intn(2) != 0)
			}
			clauses = append(clauses, c)
			s.AddClause(c)
		}
		if !bruteCNF(n, clauses, nil) {
			continue
		}
		for k := 0; k < 4; k++ {
			assumptions := make([]Lit, 1+rng.Intn(n+1))
			for i := range assumptions {
				assumptions[i] = MkLit(Var(rng.Intn(n)), rng.Intn(2) != 0)
			}
			status := s.SolveLimited(assumptions)
			want := bruteCNF(n, clauses, assumptions)
			if (status == LTrue) != want {
				t.Fatalf("trial %d assumptions=%v status=%v", trial, assumptions, status)
			}
			if want {
				continue
			}
			checked++
			if len(s.Conflict) == 0 {
				t.Fatalf("trial %d: missing core for satisfiable base", trial)
			}
			core := make([]Lit, len(s.Conflict))
			for i, lit := range s.Conflict {
				core[i] = lit.Not()
				found := false
				for _, a := range assumptions {
					if a == core[i] {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("trial %d: core=%v assumptions=%v", trial, s.Conflict, assumptions)
				}
			}
			if bruteCNF(n, clauses, core) {
				t.Fatalf("trial %d: satisfiable core %v, CNF=%v", trial, core, clauses)
			}
			if !s.Okay() || s.SolveLimited(nil) != LTrue {
				t.Fatalf("trial %d: assumptions poisoned satisfiable base", trial)
			}
		}
	}
	if checked < 100 {
		t.Fatalf("only %d unsatisfiable assumption sets tested", checked)
	}
}

func TestSolverDeadlineAndBudgetRecovery(t *testing.T) {
	s := NewSolver()
	x := s.NewVar(true, true)
	s.AddClause([]Lit{MkLit(x, false)})
	s.Deadline = time.Now().Add(-time.Second)
	if got := s.SolveLimited(nil); got != LUndef {
		t.Fatalf("expired deadline: %v", got)
	}
	s.Deadline = time.Time{}
	if got := s.SolveLimited(nil); got != LTrue || s.ModelValue(MkLit(x, false)) != LTrue {
		t.Fatalf("recovery: %v model=%v", got, s.Model)
	}
	s.SetPropBudget(0)
	if got := s.SolveLimited(nil); got != LUndef {
		t.Fatalf("zero propagation budget: %v", got)
	}
	s.BudgetOff()
	if got := s.SolveLimited(nil); got != LTrue {
		t.Fatalf("budget recovery: %v", got)
	}
}

// Learned clauses may be removed while they are not reasons for live
// assignments. Validate proof additions independently even during reduction.
func TestSolverReductionProducesValidDRUPProof(t *testing.T) {
	const pigeons, holes = 6, 5
	s := NewSolver()
	s.CertifiedUNSAT = true
	var proof bytes.Buffer
	s.CertifiedOutput = &proof
	s.nextReduce = 1
	s.FirstReduceDB = 1
	s.IncReduceDB = 1
	s.SpecialIncReduceDB = 0
	for i := 0; i < pigeons*holes; i++ {
		s.NewVar(true, true)
	}
	var cnf [][]Lit
	add := func(c []Lit) {
		cnf = append(cnf, append([]Lit(nil), c...))
		if !s.AddClause(c) {
			t.Fatal("unexpected root conflict")
		}
	}
	for p := 0; p < pigeons; p++ {
		var c []Lit
		for h := 0; h < holes; h++ {
			c = append(c, MkLit(Var(p*holes+h), false))
		}
		add(c)
	}
	for h := 0; h < holes; h++ {
		for a := 0; a < pigeons; a++ {
			for b := a + 1; b < pigeons; b++ {
				add([]Lit{MkLit(Var(a*holes+h), true), MkLit(Var(b*holes+h), true)})
			}
		}
	}
	if got := s.SolveLimited(nil); got != LFalse {
		t.Fatalf("pigeonhole status %v", got)
	}
	if s.ReduceDBs == 0 || s.RemovedClauses == 0 {
		t.Fatalf("no actual reduction: runs=%d removed=%d", s.ReduceDBs, s.RemovedClauses)
	}
	var empty bool
	for _, line := range strings.Split(strings.TrimSpace(proof.String()), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] == "d" {
			continue
		}
		var c []Lit
		for _, field := range fields {
			n, err := strconv.Atoi(field)
			if err != nil {
				t.Fatal(err)
			}
			if n == 0 {
				break
			}
			negative := n < 0
			if negative {
				n = -n
			}
			c = append(c, MkLit(Var(n-1), negative))
		}
		if !rupConflict(pigeons*holes, cnf, c) {
			t.Fatalf("invalid proof step %q", line)
		}
		cnf = append(cnf, c)
		if len(c) == 0 {
			empty = true
		}
	}
	if !empty {
		t.Fatal("proof missing final empty clause")
	}
}
