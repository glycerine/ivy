package int

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestRewriteInterpolationHasNoDeadlineAsInC(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	if got := interRwsat(g); got == nil {
		t.Fatal("source untimed rewrite failed")
	}
}
