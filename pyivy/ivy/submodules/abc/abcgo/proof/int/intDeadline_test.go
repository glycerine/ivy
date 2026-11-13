package int

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestRewriteExpiredDeadline(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	if got := rewriteForInterpolation(g, util.Abc_Clock()-1); got != nil {
		t.Fatal("expired rewrite deadline returned a graph")
	}
}
