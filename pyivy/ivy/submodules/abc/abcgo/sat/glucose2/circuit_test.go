package glucose2

import (
	"math/rand"
	"testing"
)

func TestCircuitAllGateAssignments(t *testing.T) {
	for _, xor := range []bool{false, true} {
		for phase := 0; phase < 4; phase++ {
			for bits := 0; bits < 8; bits++ {
				s := NewSolver()
				s.JustificationMode = 2
				for v := 0; v < 3; v++ {
					s.NewVar(true, true)
				}
				a, b := MkLit(0, phase&1 != 0), MkLit(1, phase&2 != 0)
				if xor {
					a, b = b, a
				}
				s.SetVarFaninLits(2, a, b)
				s.SetNewRound()
				s.MarkCone(2)
				assumptions := make([]Lit, 3)
				for v := 0; v < 3; v++ {
					assumptions[v] = MkLit(Var(v), bits&(1<<v) == 0)
				}
				x, y := ((bits&1) != 0) != (phase&1 != 0), ((bits&2) != 0) != (phase&2 != 0)
				value := x && y
				if xor {
					value = x != y
				}
				want := value == ((bits & 4) != 0)
				if got := s.Solve(assumptions); got != want {
					t.Fatalf("xor%v phase%d bits%d got%v want%v", xor, phase, bits, got, want)
				}
			}
		}
	}
}

func TestCircuitJustificationAndGateReplacement(t *testing.T) {
	s := NewSolver()
	s.JustificationMode = 2
	for v := 0; v < 4; v++ {
		s.NewVar(true, true)
	}
	s.SetVarFaninLits(2, 0, 2)
	s.SetVarFaninLits(3, 0, 4)
	s.SetNewRound()
	s.MarkCone(3)
	if !s.Solve([]Lit{MkLit(3, true)}) {
		t.Fatal("false AND should be satisfiable")
	}
	if len(s.JustModel) < 2 {
		t.Fatal("missing PI justification", s.JustModel)
	}
	// A second call must not retain a stale justification watch or heap entry.
	if !s.Solve([]Lit{MkLit(3, false)}) {
		t.Fatal("true AND should be satisfiable")
	}
	s.SetVarFaninLits(3, 2, 4)
	if len(s.circuit.fanouts[0]) != 1 || len(s.circuit.fanouts[1]) != 2 {
		t.Fatal("old gate fanouts retained", s.circuit.fanouts)
	}
}

func TestCircuitRandomCones(t *testing.T) {
	rng := rand.New(rand.NewSource(71))
	for trial := 0; trial < 500; trial++ {
		s := NewSolver()
		s.JustificationMode = 2
		for i := 0; i < 9; i++ {
			s.NewVar(true, true)
		}
		var cnf [][]Lit
		for v := Var(3); v < 9; v++ {
			a, b := Var(rng.Intn(int(v))), Var(rng.Intn(int(v)))
			for b == a {
				b = Var(rng.Intn(int(v)))
			}
			x, y := MkLit(a, rng.Intn(2) != 0), MkLit(b, rng.Intn(2) != 0)
			s.SetVarFaninLits(v, x, y)
			z := MkLit(v, false)
			if a < b {
				cnf = append(cnf, []Lit{z.Not(), x}, []Lit{z.Not(), y}, []Lit{z, x.Not(), y.Not()})
			} else {
				cnf = append(cnf, []Lit{x, y, z.Not()}, []Lit{x.Not(), y.Not(), z.Not()}, []Lit{x, y.Not(), z}, []Lit{x.Not(), y, z})
			}
		}
		s.SetNewRound()
		for v := Var(0); v < 9; v++ {
			s.MarkCone(v)
		}
		for round := 0; round < 10; round++ {
			var assumptions []Lit
			for k := 0; k < 1+rng.Intn(5); k++ {
				assumptions = append(assumptions, MkLit(Var(rng.Intn(9)), rng.Intn(2) != 0))
			}
			want := bruteCNF(9, cnf, assumptions)
			if got := s.Solve(assumptions); got != want {
				t.Fatalf("trial%d round%d got%v want%v assumptions%v", trial, round, got, want, assumptions)
			}
		}
	}
}

func TestCircuitNativeStatusOracle(t *testing.T) {
	state := uint32(71)
	next := func() uint32 { state = state*1664525 + 1013904223; return state }
	hash := uint64(1469598103934665603)
	for trial := 0; trial < 500; trial++ {
		s := NewSolver()
		s.JustificationMode = 2
		for i := 0; i < 9; i++ {
			s.NewVar(true, true)
		}
		for v := Var(3); v < 9; v++ {
			a, b := Var(next()%uint32(v)), Var(next()%uint32(v))
			for b == a {
				b = Var(next() % uint32(v))
			}
			sa, sb := next()>>31 != 0, next()>>31 != 0
			s.SetVarFaninLits(v, MkLit(a, sa), MkLit(b, sb))
		}
		s.SetNewRound()
		for v := Var(0); v < 9; v++ {
			s.MarkCone(v)
		}
		for round := 0; round < 10; round++ {
			count := 1 + int(next()%5)
			var as []Lit
			for k := 0; k < count; k++ {
				v := Var(next() % 9)
				sign := next()>>31 != 0
				as = append(as, MkLit(v, sign))
			}
			if s.Solve(as) {
				hash ^= 1
			}
			hash *= 1099511628211
		}
	}
	if hash != 2660800957835849238 {
		t.Fatalf("native circuit status oracle %d", hash)
	}
}

func TestCircuitJustificationWithCNF(t *testing.T) {
	for mode := 1; mode <= 2; mode++ {
		for bits := 0; bits < 8; bits++ {
			s := NewSolver()
			s.JustificationMode = mode
			for i := 0; i < 3; i++ {
				s.NewVar(true, true)
			}
			s.SetVarFaninLits(2, 0, 2)
			s.SetNewRound()
			s.MarkCone(2)
			for _, c := range [][]Lit{{5, 0}, {5, 2}, {4, 1, 3}} {
				s.AddClause(c)
			}
			as := []Lit{MkLit(0, bits&1 == 0), MkLit(1, bits&2 == 0), MkLit(2, bits&4 == 0)}
			want := (bits&3 == 3) == (bits&4 != 0)
			if got := s.Solve(as); got != want {
				t.Fatalf("mode%d bits%d got%v", mode, bits, got)
			}
		}
	}
}
