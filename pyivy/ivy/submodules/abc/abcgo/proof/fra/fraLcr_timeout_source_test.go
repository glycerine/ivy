package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// The C timeout is checked inside the partition loop. For this graph the
// initial simulation produces no candidate partition, so even a negative
// deadline completes one iteration; a compiled C fixture confirms this.
func TestLatchCorrespondenceEmptyPartitionsIgnoreExpiredDeadline(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, state))
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	iterations := int32(-7)
	result := Fra_FraigLatchCorrespondence(g, 0, 100, 0, 0, &iterations, -1)
	if result == nil || iterations != 1 {
		t.Fatalf("result=%v iterations=%d, want a graph and one iteration", result, iterations)
	}
}

func TestLatchCorrespondenceDeadlineUsesCFloatRounding(t *testing.T) {
	// C converts the integer Abc_Clock result to float before adding the
	// float timeout. At this boundary 16777217 rounds down to 16777216.
	now := int64(16777217)
	want := int64(16777216 + fraLcrTicksPerSecond)
	if got := fraLcrDeadline(now, 1); got != want {
		t.Fatalf("deadline=%d, want %d", got, want)
	}
}
