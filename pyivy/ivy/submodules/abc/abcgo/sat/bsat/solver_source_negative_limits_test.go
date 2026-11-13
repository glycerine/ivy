package bsat

import "testing"

type sourceBudgetSolver interface {
	NewVariable() int32
	Solve(...Literal) Status
	SolveWithSourceLimits([]Literal, int64, int64, int64, int64) Status
}

// Expected values were obtained from the compiled ABC sat_solver and
// sat_solver2 with two free variables and the corresponding signed limits.
func TestSourceSignedBudgetsAgainstC(t *testing.T) {
	for _, factory := range []struct {
		name string
		new  func() sourceBudgetSolver
	}{{"solver", func() sourceBudgetSolver { return NewSolver() }}, {"solver2", func() sourceBudgetSolver { return NewSolver2() }}} {
		t.Run(factory.name, func(t *testing.T) {
			for position := 0; position < 4; position++ {
				for _, warm := range []bool{false, true} {
					s := factory.new()
					s.NewVariable()
					s.NewVariable()
					if warm && s.Solve() != Satisfiable {
						t.Fatal("warm solve failed")
					}
					limits := [4]int64{}
					limits[position] = -1
					if got := s.SolveWithSourceLimits(nil, limits[0], limits[1], limits[2], limits[3]); got != Unknown {
						t.Fatalf("position=%d warm=%t status=%v", position, warm, got)
					}
				}
			}
		})
	}
}

// C computes local limits relative to existing counters. After one warm
// propagation, -1 yields the zero sentinel (unlimited), so SAT is possible.
// Contradictory assumptions are also resolved before the budget is checked.
func TestSourceSignedBudgetsCanReturnSATAndUNSAT(t *testing.T) {
	for _, s := range []sourceBudgetSolver{NewSolver(), NewSolver2()} {
		x := s.NewVariable()
		if got := s.Solve(); got != Satisfiable {
			t.Fatalf("warm status=%v", got)
		}
		if got := s.SolveWithSourceLimits(nil, 0, -1, 0, 0); got != Satisfiable {
			t.Fatalf("zero-sentinel status=%v", got)
		}
		if got := s.SolveWithSourceLimits([]Literal{Positive(x), Negative(x)}, -1, 0, 0, 0); got != Unsatisfiable {
			t.Fatalf("contradictory assumptions status=%v", got)
		}
	}
}

// These source-oracle cases set both counters to three before a solve. Local
// -3 cancels the counter to ABC's zero/unlimited sentinel, unlike -4 or -2.
func TestSourceSignedBudgetsZeroSentinelAgainstC(t *testing.T) {
	for _, proof := range []bool{false, true} {
		for position := 0; position < 2; position++ {
			for _, limit := range []int64{-4, -3, -2} {
				var s sourceBudgetSolver
				if proof {
					p := NewSolver2()
					p.solver.stats.conflicts, p.solver.stats.propagations = 3, 3
					s = p
				} else {
					p := NewSolver()
					p.stats.conflicts, p.stats.propagations = 3, 3
					s = p
				}
				var conflicts, propagations int64
				if position == 0 {
					conflicts = limit
				} else {
					propagations = limit
				}
				want := Unknown
				if limit == -3 {
					want = Satisfiable
				}
				if got := s.SolveWithSourceLimits(nil, conflicts, propagations, 0, 0); got != want {
					t.Fatalf("proof=%t position=%d local=%d status=%v want=%v", proof, position, limit, got, want)
				}
			}
		}
	}
}

func TestSourceSignedBudgetsDoNotReplacePersistentLimits(t *testing.T) {
	for _, proof := range []bool{false, true} {
		var s sourceBudgetSolver
		var internal *sat_solver_t
		if proof {
			p := NewSolver2()
			p.SetLimits(7, 8)
			internal = &p.solver.sat_solver_t
			s = p
		} else {
			p := NewSolver()
			p.SetLimits(7, 8)
			internal = p
			s = p
		}
		s.NewVariable()
		if got := s.SolveWithSourceLimits(nil, -1, 0, 0, 0); got != Unknown {
			t.Fatalf("proof=%t status=%v", proof, got)
		}
		if c, p := internal.Limits(); c != 7 || p != 8 {
			t.Fatalf("proof=%t persistent limits=%d/%d", proof, c, p)
		}
	}
}

func TestSourceSignedGlobalCapsOverrideZeroLocalSentinel(t *testing.T) {
	for _, proof := range []bool{false, true} {
		for position := 0; position < 2; position++ {
			var s sourceBudgetSolver
			if proof {
				p := NewSolver2()
				p.solver.stats.conflicts, p.solver.stats.propagations = 3, 3
				s = p
			} else {
				p := NewSolver()
				p.stats.conflicts, p.stats.propagations = 3, 3
				s = p
			}
			limits := [4]int64{}
			limits[position], limits[position+2] = -3, -1
			if got := s.SolveWithSourceLimits(nil, limits[0], limits[1], limits[2], limits[3]); got != Unknown {
				t.Fatalf("proof=%t position=%d status=%v", proof, position, got)
			}
		}
	}
}

func TestSourceSignedBudgetsDoNotMaskEarlyRootConflict(t *testing.T) {
	for _, proof := range []bool{false, true} {
		for position := 0; position < 4; position++ {
			var solver sourceBudgetSolver
			var add func(...Literal) bool
			if proof {
				p := NewSolver2()
				solver, add = p, p.AddClause
			} else {
				p := NewSolver()
				solver, add = p, p.AddClause
			}
			x, y := solver.NewVariable(), solver.NewVariable()
			if !add(Positive(x), Positive(y)) || !add(Negative(x)) || !add(Negative(y)) {
				t.Fatal("fixture unexpectedly rejected during clause insertion")
			}
			limits := [4]int64{}
			limits[position] = -1
			if got := solver.SolveWithSourceLimits([]Literal{Positive(x)}, limits[0], limits[1], limits[2], limits[3]); got != Unsatisfiable {
				t.Fatalf("proof=%t position=%d early root conflict returned %v", proof, position, got)
			}
		}
	}
}

func TestSolver2PublicLimitsRejectNegativeBudgets(t *testing.T) {
	for _, limits := range [][2]int64{{-1, 0}, {0, -1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("SetLimits(%d, %d) did not reject negative budget", limits[0], limits[1])
				}
			}()
			NewSolver2().SetLimits(limits[0], limits[1])
		}()
	}
}
