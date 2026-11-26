package gia

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Gia_ManFromAigChoices allocates pSibls for every object up front, so a
// choice graph has a full sibling array (and Gia_ManHasChoices holds) even
// when siblings precede later objects such as the COs, or are absent.
func TestFromAigChoicesAllocatesFullSiblingArray(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x := g.AND(a, b)
	y := g.AND(g.AND(a, c), b) // a different structure for the same class
	aig.Aig_ObjSetEquiv(g, aig.Aig_Regular(y), aig.Aig_Regular(x))
	g.AddOutput(g.AND(y, c))
	g.AddOutput(c)
	p := Gia_ManFromAigChoices(g)
	if len(p.pSibls) != int(p.nObjs) {
		t.Fatalf("pSibls has %d entries for %d objects", len(p.pSibls), p.nObjs)
	}
	if err := Gia_AigerWrite(p, filepath.Join(t.TempDir(), "c.aig"), 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	plain := aig.NewGraph()
	plain.AddOutput(plain.AND(plain.NewInput(), plain.NewInput()))
	if q := Gia_ManFromAigChoices(plain); Gia_ManHasChoices(q) == 0 || len(q.pSibls) != int(q.nObjs) {
		t.Fatal("choice import without choices must still allocate siblings")
	}
}
