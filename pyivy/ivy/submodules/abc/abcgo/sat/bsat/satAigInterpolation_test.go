package bsat

import (
	"math/rand"
	"testing"
)

func TestAigInterpolationCraig(t *testing.T) {
	rng := rand.New(rand.NewSource(338))
	for trial := 0; trial < 150; trial++ {
		partitions := [2][][]lit{}
		for part := range partitions {
			for count := 0; count < 9; count++ {
				clause := []lit{}
				for variable := 0; variable < 4; variable++ {
					if rng.Intn(2) == 0 {
						clause = append(clause, lit(2*variable+rng.Intn(2)))
					}
				}
				if len(clause) > 0 {
					partitions[part] = append(partitions[part], clause)
				}
			}
		}
		cnf := partitionedProof(t, partitions[0], partitions[1])
		if cnf == nil {
			continue
		}
		for _, mux := range []bool{false, true} {
			manager := aigInterpolation{}
			graph := manager.interpolate(cnf, nil, 0, mux, 0)
			if manager.lastErr != nil {
				t.Fatal(manager.lastErr)
			}
			table := []uint32{0}
			for point := 0; point < 32; point++ {
				values := make([]bool, len(manager.vVarsAB))
				for i := range values {
					values[i] = point&(1<<uint(i)) != 0
				}
				out, err := graph.Evaluate(values)
				if err != nil {
					t.Fatal(err)
				}
				if out[0] {
					table[0] |= uint32(1) << uint(point)
				}
			}
			checkTruthCraig(t, cnf, table, manager.vVarsAB)
			if err := graph.Check(); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestInterpolationUsesExplicitClockDomain(t *testing.T) {
	cnf := partitionedProof(t, [][]lit{{0}}, [][]lit{{1}})
	if cnf == nil {
		t.Fatal("expected UNSAT proof")
	}
	manager := Inta_ManAlloc()
	if got := Inta_ManInterpolateWithClock(manager, cnf, 100, []int32{0}, 0, func() int64 { return 99 }); got == nil {
		t.Fatalf("live custom deadline rejected: %v", Inta_ManError(manager))
	}
	if got := Inta_ManInterpolateWithClock(manager, cnf, 100, []int32{0}, 0, func() int64 { return 101 }); got != nil {
		t.Fatal("expired custom deadline accepted")
	}
}
