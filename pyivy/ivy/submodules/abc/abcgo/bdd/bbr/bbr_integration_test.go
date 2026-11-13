package bbr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestReachabilityProvesAndFindsCounterexample(t *testing.T) {
	for _, fail := range []bool{false, true} {
		g := aig.NewGraph()
		lo := g.NewInput()
		g.AddOutput(lo)
		next := g.Const0()
		if fail {
			next = g.Const1()
		}
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		var p Saig_ParBbr_t
		Bbr_ManSetDefaultParams(&p)
		p.FSilent = 1
		got := Aig_ManVerifyUsingBdds(g, &p)
		if fail && got != 0 {
			t.Fatalf("expected failing reachable state, got %d", got)
		}
		if !fail && got != 1 {
			t.Fatalf("expected safe invariant, got %d", got)
		}
	}
}
