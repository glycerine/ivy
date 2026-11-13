package int

import (
	"math/rand"
	"testing"
)

func TestMiniSatInterpolantsAgainstCraigOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(114123))
	checked := 0
	jointOnly := 0
	for trial := 0; trial < 400; trial++ {
		var clauses [][]int32
		var parts []int32
		s := &miniSat114pBackend{}
		s.SetVarNum(3)
		for part := int32(0); part < 2; part++ {
			for i := 0; i < 1+rng.Intn(5); i++ {
				clause := make([]int32, 1+rng.Intn(3))
				for j := range clause {
					v := int32(2)
					if rng.Intn(2) == 0 {
						v = part
					}
					clause[j] = v*2 + int32(rng.Intn(2))
				}
				clauses = append(clauses, clause)
				parts = append(parts, part)
				if !s.AddClause(clause) {
					break
				}
			}
		}
		if s.Solve() != 0 {
			continue
		}
		indices := s.rootInputIndices()
		rootParts := make([]int32, len(indices))
		for i, index := range indices {
			rootParts[i] = parts[index]
		}
		partitionSatisfiable := [2]bool{}
		for _, pudlak := range []bool{false, true} {
			g, err := InterpolateM114p(s, rootParts, []int32{-1, -2, 0}, pudlak)
			if err != nil {
				t.Fatalf("trial=%d pudlak=%v: %v clauses=%v", trial, pudlak, err, clauses)
			}
			for assignment := 0; assignment < 8; assignment++ {
				sat := [2]bool{true, true}
				for i, c := range clauses {
					ok := false
					for _, lit := range c {
						if (assignment>>uint(lit>>1))&1 != int(lit&1) {
							ok = true
							break
						}
					}
					if !ok {
						sat[parts[i]] = false
					}
				}
				partitionSatisfiable[0] = partitionSatisfiable[0] || sat[0]
				partitionSatisfiable[1] = partitionSatisfiable[1] || sat[1]
				out, err := g.Evaluate([]bool{assignment&4 != 0})
				if err != nil || sat[0] && !out[0] || sat[1] && out[0] {
					t.Fatalf("trial=%d pudlak=%v assignment=%d A=%v B=%v I=%v err=%v clauses=%v parts=%v", trial, pudlak, assignment, sat[0], sat[1], out, err, clauses, parts)
				}
			}
		}
		checked++
		if partitionSatisfiable[0] && partitionSatisfiable[1] {
			jointOnly++
		}
	}
	if checked < 50 {
		t.Fatalf("only %d UNSAT proofs checked", checked)
	}
	if jointOnly < 10 {
		t.Fatalf("only %d proofs had individually satisfiable partitions", jointOnly)
	}
	t.Logf("checked both interpolants for %d UNSAT formulas (%d with individually satisfiable partitions)", checked, jointOnly)
}
