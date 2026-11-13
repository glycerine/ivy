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
