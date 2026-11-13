package abs

import (
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
