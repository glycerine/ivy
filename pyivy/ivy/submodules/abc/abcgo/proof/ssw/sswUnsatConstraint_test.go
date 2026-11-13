package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestConstrainedSignalCorrespondenceRejectsImpossibleConstraint(t *testing.T) {
	g := aig.NewGraph()
	pi, q := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(pi, q))
	g.AddOutput(g.Const1())
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	pars.fConstrs = 1
	pars.nStepsMax = 2
	if result := Ssw_SignalCorrespondence(g, &pars); result != nil {
		t.Fatal("impossible constraint accepted")
	}
}
