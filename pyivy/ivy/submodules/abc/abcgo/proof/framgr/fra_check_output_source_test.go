package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSimulationOutputCheckRequiresSourceReferencePhase(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x.Not())
	p := &Fra_Man_t{PManAig: g, PSml: Fra_SmlStart(g, 0, 1, 1)}
	defer func() {
		if recover() == nil {
			t.Error("source asserts when the first output's reference pattern is true")
		}
	}()
	Fra_SmlCheckOutput(p)
}
