package abc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func TestInductionWitnessHandoff(t *testing.T) {
	g := aig.NewGraph()
	for i := 0; i < 3; i++ {
		g.NewInput()
	}
	for i := 0; i < 3; i++ {
		g.AddOutput(g.Const0())
	}
	aig.Aig_ManSetRegNum(g, 2)
	witness := util.Abc_CexAlloc(1, 1, 1)
	if err := g.SetInductionCounterexample(witness); err != nil {
		t.Fatal(err)
	}
	n := NewNetwork("witness")
	n.Data = "preserve generic user data"
	n.Model = []int32{1}
	n.TakeInductionCounterexample(g)
	if n.SequenceModel != witness || g.InductionCounterexample() != nil || n.Model != nil || n.Data != "preserve generic user data" {
		t.Fatal("typed witness ownership not transferred")
	}
}
