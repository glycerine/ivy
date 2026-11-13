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
