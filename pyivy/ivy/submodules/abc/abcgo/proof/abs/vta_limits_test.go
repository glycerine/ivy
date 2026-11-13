package abs

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
	"time"
)

func TestVtaUnsatCoreReportsDeadlineAndRecovers(t *testing.T) {
	s := bsat.NewLabeledSolver2()
	s.AddClause(1, bsat.Positive(1), bsat.Positive(2))
	s.SetDeadline(time.Unix(0, 1))
	var status, conflicts int32
	if core := Vta_ManUnsatCore(int32(bsat.Negative(1)), s, 0, 0, &status, &conflicts); core != nil || status != -1 {
		t.Fatalf("expired deadline: status %d, core %v", status, core)
	}
	s.SetDeadline(time.Time{})
	if core := Vta_ManUnsatCore(int32(bsat.Negative(1)), s, 0, 0, &status, &conflicts); core != nil || status != 0 {
		t.Fatalf("recovery: status %d, core %v", status, core)
	}
}

// An implication chain yields a source-style labeled proof whose core must
// retain every link. It also crosses common small-vector allocation sizes.
func TestVtaUnsatCoreLongLabeledChain(t *testing.T) {
	const links = 256
	s := bsat.NewLabeledSolver2()
	for i := 1; i <= links; i++ {
		if !s.AddClause(int32(i), bsat.Negative(int32(i)), bsat.Positive(int32(i+1))) {
			t.Fatalf("adding link %d", i)
		}
	}
	if !s.AddClause(links+1, bsat.Negative(links+1)) {
		t.Fatal("adding final unit")
	}
	var status, conflicts int32
	core := Vta_ManUnsatCore(int32(bsat.Positive(1)), s, 0, 0, &status, &conflicts)
	if status != 1 || core == nil {
		t.Fatalf("status=%d core=%v conflicts=%d", status, core, conflicts)
	}
	seen := make(map[int32]bool)
	for _, label := range vec.Vec_IntArray(core) {
		seen[label] = true
	}
	for i := int32(1); i <= links+1; i++ {
		if !seen[i] {
			t.Fatalf("missing label %d from %d-label proof core", i, vec.Vec_IntSize(core))
		}
	}
}

// test_vectors/vta_conflict_oracle.c constructs the same guarded pigeonhole
// instance and invokes the C solver call used by Vta_ManUnsatCore. Compiled
// ABC returns unknown at budget 1 (2 conflicts) and UNSAT without a limit
// (7 conflicts); the individual conflict count is implementation dependent.
func TestVtaUnsatCoreConflictBudgetAgainstC(t *testing.T) {
	for _, budget := range []int32{1, 0} {
		s := bsat.NewLabeledSolver2()
		for pigeon := int32(0); pigeon < 4; pigeon++ {
			clause := []bsat.Literal{bsat.Negative(0)}
			for hole := int32(0); hole < 3; hole++ {
				clause = append(clause, bsat.Positive(1+pigeon*3+hole))
			}
			if !s.AddClause(1+pigeon, clause...) {
				t.Fatal("adding pigeon clause")
			}
		}
		for hole := int32(0); hole < 3; hole++ {
			for pigeon := int32(0); pigeon < 4; pigeon++ {
				for other := pigeon + 1; other < 4; other++ {
					if !s.AddClause(5+hole*6+pigeon*4+other, bsat.Negative(1+pigeon*3+hole), bsat.Negative(1+other*3+hole)) {
						t.Fatal("adding hole clause")
					}
				}
			}
		}
		var status, conflicts int32
		core := Vta_ManUnsatCore(int32(bsat.Positive(0)), s, budget, 0, &status, &conflicts)
		if budget == 1 {
			if status != -1 || core != nil || conflicts < budget {
				t.Fatalf("budget=%d status=%d conflicts=%d core=%v", budget, status, conflicts, core)
			}
		} else if status != 1 || core == nil || conflicts == 0 {
			t.Fatalf("budget=%d status=%d conflicts=%d core=%v", budget, status, conflicts, core)
		}
	}
}
