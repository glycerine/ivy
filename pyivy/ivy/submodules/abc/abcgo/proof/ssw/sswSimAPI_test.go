package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestNonConstantOutputExcludesConstraints(t *testing.T) {
	g := aig.NewGraph()
	a := g.NewInput()
	property, constraint := g.AddOutput(a), g.AddOutput(a)
	aig.Aig_ManSetConstrNum(g, 1)
	p := Ssw_SmlStart(g, 0, 1, 1)
	Ssw_SmlSimInfo(p, constraint)[0] = 1
	if Ssw_SmlCheckNonConstOutputs(p) != 0 {
		t.Fatal("constraint counted as property")
	}
	Ssw_SmlSimInfo(p, property)[0] = 1
	if Ssw_SmlCheckNonConstOutputs(p) != 1 {
		t.Fatal("nonconstant property missed")
	}
}
