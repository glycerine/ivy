package core
import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
)

// Dar_RefactTryGraph reuses an existing strashed node outside the MFFC,
// but rejects reusing the root currently being refactored.
func TestRefactorCandidateReusesStrashedNode(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	shared := g.AND(a, b)
	root := g.AND(shared, c)
	g.AddOutput(root)
	g.AddOutput(shared)
	aig.Aig_ManIncrementTravId(g)
	var cover []uint32
	if kit.Kit_TruthIsop([]uint32{0x88888888}, 2, &cover, 0) < 0 {
		t.Fatal("ISOP failed")
	}
	candidate := kit.Kit_SopFactor(cover, 0, 2, &cover)
	cost, _, ok := refactorTryKit(g, root, []aig.Edge{a, b}, candidate, 0, 100)
	if !ok || cost != 0 {
		t.Fatalf("existing shared node: cost=%d ok=%v", cost, ok)
	}
	_, _, ok = refactorTryKit(g, shared, []aig.Edge{a, b}, candidate, 1, 100)
	if ok {
		t.Fatal("candidate matching root must be rejected")
	}
}
