package minisatp

import (
	"fmt"
	"math/rand"
	"testing"
)

func checkProof(s *solver) error {
	clauses := map[ClauseID]map[int32]bool{}
	var final map[int32]bool
	for _, event := range s.proof.events {
		if event.deleted {
			continue
		}
		clause := map[int32]bool{}
		if len(event.chain.Clauses) == 0 {
			for _, lit := range event.root {
				clause[lit] = true
			}
		} else {
			first, ok := clauses[event.chain.Clauses[0]]
			if !ok {
				return fmt.Errorf("missing antecedent")
			}
			for lit := range first {
				clause[lit] = true
			}
			for i, v := range event.chain.Pivots {
				other, ok := clauses[event.chain.Clauses[i+1]]
				if !ok {
					return fmt.Errorf("missing antecedent")
				}
				a, b := v*2, v*2+1
				if !(clause[a] && other[b] || clause[b] && other[a]) {
					return fmt.Errorf("invalid pivot %d", v)
				}
				delete(clause, a)
				delete(clause, b)
				for lit := range other {
					if lit != a && lit != b {
						clause[lit] = true
					}
				}
			}
		}
		clauses[event.id] = clause
		final = clause
	}
	if len(final) != 0 {
		return fmt.Errorf("last clause is not empty: %v", final)
	}
	return nil
}

func TestSmallFormulasAgainstExhaustiveAssignments(t *testing.T) {
	rng := rand.New(rand.NewSource(114))
	for trial := 0; trial < 1000; trial++ {
		n := 1 + rng.Intn(5)
		var formula [][]int32
		s := newSolver(n)
		for j := 0; j < 1+rng.Intn(18); j++ {
			clause := make([]int32, 1+rng.Intn(4))
			for i := range clause {
				clause[i] = int32(rng.Intn(2 * n))
			}
			formula = append(formula, clause)
			s.addClause(clause)
		}
		sat := false
		for assignment := 0; assignment < 1<<n; assignment++ {
			all := true
			for _, clause := range formula {
				any := false
				for _, lit := range clause {
					if ((assignment >> uint(lit>>1)) & 1) != int(lit&1) {
						any = true
						break
					}
				}
				if !any {
					all = false
					break
				}
			}
			if all {
				sat = true
				break
			}
		}
		got := s.solve()
		if (got == 1) != sat {
			t.Fatalf("trial %d formula %v: got %d, sat %v", trial, formula, got, sat)
		}
		if !sat {
			if err := checkProof(s); err != nil {
				t.Fatalf("trial %d formula %v: %v", trial, formula, err)
			}
		}
	}
}
