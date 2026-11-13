package glucose2

import (
	"bytes"
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func bruteCNF(n int, clauses [][]Lit, assumptions []Lit) bool {
	for bits := 0; bits < 1<<n; bits++ {
		ok := true
		for _, p := range assumptions {
			if ((bits>>p.Var())&1 != 0) == p.Sign() {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for _, clause := range clauses {
			sat := false
			for _, p := range clause {
				if ((bits>>p.Var())&1 != 0) != p.Sign() {
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
func TestSolverRandomCNFDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(2381))
	for trial := 0; trial < 1500; trial++ {
		n := 1 + rng.Intn(9)
		s := NewSolver()
		s.FirstReduceDB = 4
		s.nextReduce = 4
		s.IncReduceDB = 2
		s.lbdQueue.InitSize(2)
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
		}
		var cnf [][]Lit
		for c := rng.Intn(45); c > 0; c-- {
			var clause []Lit
			for k := 1 + rng.Intn(5); k > 0; k-- {
				clause = append(clause, MkLit(Var(rng.Intn(n)), rng.Intn(2) != 0))
			}
			cnf = append(cnf, clause)
			s.AddClause(clause)
		}
		for round := 0; round < 5; round++ {
			var assumptions []Lit
			for k := rng.Intn(4); k > 0; k-- {
				assumptions = append(assumptions, MkLit(Var(rng.Intn(n)), rng.Intn(2) != 0))
			}
			got := s.Solve(assumptions)
			want := bruteCNF(n, cnf, assumptions)
			if got != want {
				t.Fatalf("trial %d round %d got %v want %v cnf=%v assumptions=%v", trial, round, got, want, cnf, assumptions)
			}
			if got {
				for _, clause := range cnf {
					sat := false
					for _, p := range clause {
						if s.ModelValue(p) == LTrue {
							sat = true
						}
					}
					if !sat {
						t.Fatalf("bad model trial %d clause %v", trial, clause)
					}
				}
			}
		}
	}
}
func TestSolverIncrementalAndBudget(t *testing.T) {
	s := NewSolver()
	x := s.NewVar(false, true)
	y := s.NewVar(false, true)
	s.AddClause([]Lit{MkLit(x, false), MkLit(y, false)})
	if !s.Solve(nil) {
		t.Fatal("SAT")
	}
	if s.Solve([]Lit{MkLit(x, true), MkLit(y, true)}) {
		t.Fatal("assumptions should conflict")
	}
	if !s.Okay() || !s.Solve(nil) {
		t.Fatal("assumption UNSAT poisoned formula")
	}
	s.SetConfBudget(0)
	if s.SolveLimited(nil) != LUndef {
		t.Fatal("zero budget")
	}
	s.BudgetOff()
	s.Interrupt()
	if s.SolveLimited(nil) != LUndef {
		t.Fatal("interrupt")
	}
	s.ClearInterrupt()
	s.AddClause([]Lit{MkLit(x, true)})
	s.AddClause([]Lit{MkLit(y, true)})
	if s.Solve(nil) || s.Okay() {
		t.Fatal("incremental contradiction")
	}
}

// Golden statuses from the unchanged C++ Solver compiled with Glucose2.cpp,
// Options2.cpp and System2.cpp. The LCG and complete request sequence are kept
// here so repeated assumptions and formula reuse are covered by the comparison.
func TestSolverNativeStatusOracle(t *testing.T) {
	state := uint32(2381)
	next := func() uint32 { state = state*1664525 + 1013904223; return state }
	hash := uint64(1469598103934665603)
	for trial := 0; trial < 500; trial++ {
		s := NewSolver()
		n := 1 + int(next()%9)
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
		}
		count := int(next() % 45)
		for c := 0; c < count; c++ {
			length := 1 + int(next()%5)
			var ps []Lit
			for k := 0; k < length; k++ {
				v := Var(next() % uint32(n))
				sign := next()>>31 != 0
				ps = append(ps, MkLit(v, sign))
			}
			s.AddClause(ps)
		}
		for round := 0; round < 5; round++ {
			count := int(next() % 4)
			var assumptions []Lit
			for k := 0; k < count; k++ {
				v := Var(next() % uint32(n))
				sign := next()>>31 != 0
				assumptions = append(assumptions, MkLit(v, sign))
			}
			if s.Solve(assumptions) {
				hash ^= 1
			}
			hash *= 1099511628211
		}
	}
	if hash != 4376511358172924066 {
		t.Fatalf("native status oracle: %d", hash)
	}
}

func TestSolverCollectionAndReset(t *testing.T) {
	s := NewSolver()
	s.NewVar(true, true)
	s.NewVar(true, true)
	s.AddClause([]Lit{0, 2})
	s.AddClause([]Lit{0})
	if !s.Simplify() {
		t.Fatal("simplify")
	}
	s.GarbageCollect()
	for _, c := range s.allocator.clauses {
		if c.freed {
			t.Fatal("dead clause retained")
		}
	}
	if !s.Solve(nil) {
		t.Fatal("SAT after collect")
	}
	s.Reset()
	if !s.Okay() || s.NVars() != 0 || s.NClauses() != 0 || s.Conflicts != 0 {
		t.Fatal("reset")
	}
	s.NewVar(false, true)
	s.AddClause([]Lit{1})
	if !s.Solve(nil) || s.Model[0] != LFalse {
		t.Fatal("reuse after reset")
	}
}

// rupConflict checks the proof rule independently of Solver propagation.
func rupConflict(n int, clauses [][]Lit, addition []Lit) bool {
	values := make([]int8, n)
	assign := func(p Lit) bool {
		v := int8(1)
		if p.Sign() {
			v = -1
		}
		old := values[p.Var()]
		if old != 0 {
			return old == v
		}
		values[p.Var()] = v
		return true
	}
	for _, p := range addition {
		if !assign(p.Not()) {
			return true
		}
	}
	for {
		changed := false
		for _, c := range clauses {
			unknown := 0
			unit := LitUndef
			sat := false
			for _, p := range c {
				v := values[p.Var()]
				if v == 0 {
					unknown++
					unit = p
				} else if (v == 1) != p.Sign() {
					sat = true
					break
				}
			}
			if sat {
				continue
			}
			if unknown == 0 {
				return true
			}
			if unknown == 1 {
				if !assign(unit) {
					return true
				}
				changed = true
			}
		}
		if !changed {
			return false
		}
	}
}
func TestSolverVisibleDRUPProof(t *testing.T) {
	s := NewSolver()
	var proof bytes.Buffer
	s.CertifiedUNSAT = true
	s.CertifiedOutput = &proof
	const pigeons = 4
	const holes = 3
	for i := 0; i < pigeons*holes; i++ {
		s.NewVar(true, true)
	}
	var cnf [][]Lit
	add := func(c []Lit) { cnf = append(cnf, append([]Lit(nil), c...)); s.AddClause(c) }
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
	if s.Solve(nil) {
		t.Fatal("pigeonhole should be UNSAT")
	}
	empty := false
	for _, line := range strings.Split(strings.TrimSpace(proof.String()), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "d" {
			continue
		}
		var clause []Lit
		for _, f := range fields {
			n, err := strconv.Atoi(f)
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
			clause = append(clause, MkLit(Var(n-1), negative))
		}
		if !rupConflict(pigeons*holes, cnf, clause) {
			t.Fatalf("non-RUP proof step %q", line)
		}
		cnf = append(cnf, clause)
		if len(clause) == 0 {
			empty = true
		}
	}
	if !empty {
		t.Fatal("proof did not conclude with empty clause")
	}
}

func TestSolvePreservesUndefinedNonDecisionVariables(t *testing.T) {
	s := NewSolver()
	x := s.NewVar(true, false)
	y := s.NewVar(true, false)
	if !s.AddClause([]Lit{MkLit(x, false), MkLit(y, false)}) || !s.AddClause([]Lit{MkLit(x, true), MkLit(y, true)}) {
		t.Fatal("unexpected inconsistency")
	}
	if got := s.SolveLimited(nil); got != LTrue {
		t.Fatalf("got %v", got)
	}
	if len(s.Model) != 2 || !s.Model[0].IsUndef() || !s.Model[1].IsUndef() {
		t.Fatalf("unassigned variables should remain undefined: %v", s.Model)
	}
}
