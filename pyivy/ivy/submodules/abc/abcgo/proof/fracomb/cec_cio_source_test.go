package fracomb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// fraCore.c tests each output driver's stored CioId against nTruePis. A
// freshly allocated second CI retains CioId zero until IDs are assigned.
func TestMiterStatusUsesStoredCioID(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	second := g.NewInput()
	g.AddOutput(second)
	g.SetTruePiCount(1)
	if got := Fra_FraigMiterStatus(g); got != 0 {
		t.Fatalf("uninitialized second CioId: status=%d, want C status 0", got)
	}
	aig.Aig_ManSetCioIds(g)
	if got := Fra_FraigMiterStatus(g); got != -1 {
		t.Fatalf("assigned second CioId: status=%d, want -1", got)
	}
}
