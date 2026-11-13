package minisatp

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestConflictAndDeadlineLimits(t *testing.T) {
	formula := [][]int32{{0, 2}, {0, 3}, {1, 2}, {1, 3}}
	for _, limited := range []bool{false, true} {
		s := New(2)
		for _, c := range formula {
			if !s.AddClause(c) {
				t.Fatal("premature conflict")
			}
		}
		if limited {
			s.SetLimits(1, time.Time{})
		}
		got := s.Solve()
		want := int32(0)
		if limited {
			want = -1
		}
		if got != want {
			t.Fatalf("limited=%v got %d want %d", limited, got, want)
		}
	}
	s := New(1)
	s.SetLimits(0, time.Now().Add(-time.Second))
	if s.Solve() != -1 || !s.Interrupted() {
		t.Fatal("expired deadline was not honored")
	}
}

func TestProofRootIndexAndInterleavedChainRemapping(t *testing.T) {
	s := New(2)
	for _, c := range [][]int32{{0}, {0, 2}, {2, 3}, {1, 2}} {
		if !s.AddClause(c) {
			t.Fatalf("unexpected conflict adding %v", c)
		}
	}
	if s.AddClause([]int32{3}) {
		t.Fatal("expected contradiction")
	}
	roots, chains, inputs := s.Proof()
	if !reflect.DeepEqual(inputs, []int{0, 3, 4}) {
		t.Fatalf("root inputs %v", inputs)
	}
	if !reflect.DeepEqual(roots, [][]int32{{0}, {1, 2}, {3}}) {
		t.Fatalf("roots %v", roots)
	}
	want := []Chain{{Clauses: []ClauseID{1, 0}, Pivots: []int32{0}}, {Clauses: []ClauseID{2, 3}, Pivots: []int32{1}}}
	if !reflect.DeepEqual(chains, want) {
		t.Fatalf("chains %+v want %+v", chains, want)
	}
}

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
		if sat {
			for _, clause := range formula {
				satisfied := false
				for _, lit := range clause {
					if (s.model[lit>>1] > 0) != (lit&1 != 0) {
						satisfied = true
						break
					}
				}
				if !satisfied {
					t.Fatalf("trial %d: model %v does not satisfy %v", trial, s.model, clause)
				}
			}
		}
		if !sat {
			if err := checkProof(s); err != nil {
				t.Fatalf("trial %d formula %v: %v", trial, formula, err)
			}
		}
	}
}
