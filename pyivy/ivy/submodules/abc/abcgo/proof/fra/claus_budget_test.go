package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Four pigeons cannot fit in three holes; proving this requires search.
func clausBudgetFixture(t *testing.T, limit int32) *Clu_Man_t {
	t.Helper()
	g := aig.NewGraph()
	g.AddOutput(g.NewInput())
	s := bsat.NewSolver()
	s.NewVariable() // Output, independent of the pigeonhole formula.
	for pigeon := int32(0); pigeon < 4; pigeon++ {
		var clause []int32
		for hole := int32(0); hole < 3; hole++ {
			clause = append(clause, bsat.Positive(1+pigeon*3+hole))
		}
		if !s.AddClause(clause...) {
			t.Fatal("unexpected contradiction during setup")
		}
	}
	for hole := int32(0); hole < 3; hole++ {
		for a := int32(0); a < 4; a++ {
			for b := a + 1; b < 4; b++ {
				if !s.AddClause(bsat.Negative(1+a*3+hole), bsat.Negative(1+b*3+hole)) {
					t.Fatal("unexpected contradiction during setup")
				}
			}
		}
	}
	s.NewVariable() // The second frame output used by the induction check.
	vars := make([]int32, len(g.Nodes()))
	return &Clu_Man_t{pAig: g, pCnf: &cnf.Cnf_Dat_t{Variables: 13, VarNums: vars}, pSatMain: s, pSatBmc: s, nFrames: 1, nBTLimit: limit}
}

func TestClausPropertyChecksHonorConflictLimit(t *testing.T) {
	for _, tc := range []struct {
		name  string
		check func(*Clu_Man_t) int32
	}{
		{"bmc", Fra_ClausRunBmc}, {"inductive", Fra_ClausRunSat}, {"single frame", Fra_ClausRunSat0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "inductive" {
				func() {
					defer func() {
						if recover() == nil {
							t.Fatal("source asserts on undefined induction result")
						}
					}()
					tc.check(clausBudgetFixture(t, 1))
				}()
			} else if got := tc.check(clausBudgetFixture(t, 1)); got != 0 {
				t.Fatalf("limited search returned proof %d", got)
			}
			if got := tc.check(clausBudgetFixture(t, 0)); got != 1 {
				t.Fatalf("unlimited search failed to prove unsatisfiability: %d", got)
			}
		})
	}
}

func TestClausBMCFilteringHonorsConflictLimit(t *testing.T) {
	for _, tc := range []struct{ limit, removed int32 }{{1, 1}, {0, 0}} {
		p := clausBudgetFixture(t, tc.limit)
		p.vLits, p.vClauses, p.vCosts = vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
		vec.Vec_IntPush(p.vLits, bsat.Positive(0))
		vec.Vec_IntPush(p.vClauses, 1)
		vec.Vec_IntPush(p.vCosts, 1)
		if got := Fra_ClausBmcClauses(p); got != tc.removed {
			t.Fatalf("limit %d: removed %d, want %d", tc.limit, got, tc.removed)
		}
	}
}
