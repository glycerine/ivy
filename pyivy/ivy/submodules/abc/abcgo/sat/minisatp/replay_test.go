package minisatp

import (
	"math/rand"
	"testing"
)

// Replay exported root-first proofs independently of the internal event log.
func TestExportedProofReplay(t *testing.T) {
	rng := rand.New(rand.NewSource(11457))
	unsat := 0
	for trial := 0; trial < 2000; trial++ {
		n := 1 + rng.Intn(7)
		s := New(n)
		var inputs [][]int32
		for i, count := 0, 1+rng.Intn(30); i < count; i++ {
			length := 1 + rng.Intn(4)
			if rng.Intn(30) == 0 {
				length = 0
			}
			c := make([]int32, length)
			for j := range c {
				c[j] = int32(rng.Intn(2 * n))
			}
			inputs = append(inputs, c)
			if !s.AddClause(c) {
				break
			}
		}
		if s.Solve() != 0 {
			continue
		}
		unsat++
		roots, chains, indexes := s.Proof()
		if len(roots) != len(indexes) {
			t.Fatalf("trial %d: root index count", trial)
		}
		clauses := make([]map[int32]bool, 0, len(roots)+len(chains))
		for i, root := range roots {
			if indexes[i] < 0 || indexes[i] >= len(inputs) {
				t.Fatalf("trial %d: invalid root index", trial)
			}
			want := make(map[int32]bool)
			for _, lit := range inputs[indexes[i]] {
				want[lit] = true
			}
			got := make(map[int32]bool)
			for _, lit := range root {
				got[lit] = true
			}
			if len(got) != len(want) {
				t.Fatalf("trial %d: root mismatch", trial)
			}
			for lit := range got {
				if !want[lit] {
					t.Fatalf("trial %d: root literal mismatch", trial)
				}
			}
			clauses = append(clauses, got)
		}
		for _, chain := range chains {
			if len(chain.Clauses) != len(chain.Pivots)+1 || len(chain.Clauses) == 0 {
				t.Fatalf("trial %d: malformed chain", trial)
			}
			first := int(chain.Clauses[0])
			if first < 0 || first >= len(clauses) {
				t.Fatalf("trial %d: invalid antecedent", trial)
			}
			result := make(map[int32]bool)
			for lit := range clauses[first] {
				result[lit] = true
			}
			for i, pivot := range chain.Pivots {
				index := int(chain.Clauses[i+1])
				if index < 0 || index >= len(clauses) {
					t.Fatalf("trial %d: invalid antecedent", trial)
				}
				other := clauses[index]
				a, b := 2*pivot, 2*pivot+1
				if !(result[a] && other[b] || result[b] && other[a]) {
					t.Fatalf("trial %d: invalid pivot %d", trial, pivot)
				}
				delete(result, a)
				delete(result, b)
				for lit := range other {
					if lit != a && lit != b {
						result[lit] = true
					}
				}
			}
			clauses = append(clauses, result)
		}
		if len(clauses) == 0 || len(clauses[len(clauses)-1]) != 0 {
			t.Fatalf("trial %d: proof does not end in contradiction", trial)
		}
	}
	if unsat < 100 {
		t.Fatalf("insufficient UNSAT examples: %d", unsat)
	}
}
