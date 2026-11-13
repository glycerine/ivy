package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestBackwardRetimeRejectsIncompleteCut(t *testing.T) {
	g := aig.NewGraph()
	input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(input)
	g.AddOutput(g.AND(input, b))
	aig.Aig_ManSetRegNum(g, 2)
	if result := Saig_ManRetimeMinAreaBackward(g, 0); result != nil {
		t.Fatal("incomplete cut should not be retimed")
	}
}
