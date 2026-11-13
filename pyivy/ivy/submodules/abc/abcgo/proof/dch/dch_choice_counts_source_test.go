package dch

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// dchChoice.c:75-101 counts each link and asserts that its destination has a
// lower object ID than its owner.
func TestChoiceLinkCountsEnforceSourceOrder(t *testing.T) {
	for _, kind := range []string{"representative", "equivalence"} {
		t.Run(kind, func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			count := Dch_DeriveChoiceCountReprs
			set := aig.Aig_ObjSetRepr
			if kind == "equivalence" {
				count = Dch_DeriveChoiceCountEquivs
				set = aig.Aig_ObjSetEquiv
			}
			set(g, b, a)
			if got := count(g); got != 1 {
				t.Fatalf("count = %d, want 1", got)
			}
			set(g, a, b)
			defer func() {
				if recover() == nil {
					t.Fatal("expected source order assertion")
				}
			}()
			count(g)
		})
	}
}
