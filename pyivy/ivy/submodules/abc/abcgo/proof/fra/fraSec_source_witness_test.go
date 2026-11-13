package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

// Compiled fraSec.c returns failure without copying its temporary model when
// the source graph's nTruePis does not match the reduced graph's count.
func TestSecSourceDirectInputDoesNotTransferWitness(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	var pars Fra_Sec_t
	Fra_SecSetDefaultParams(&pars)
	pars.fSilent = 1
	if status := Fra_FraigSec(g, &pars, nil); status != 0 || g.SequentialCounterexample() != nil {
		t.Fatalf("status=%d model=%+v", status, g.SequentialCounterexample())
	}
}

// ABC creates its trivial model on the reduced graph without setting iPo to
// the failing output; the optional verified wrapper can repair that model.
func TestSecSourceTrivialWitnessRetainsDefaultOutput(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const1())
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	var pars Fra_Sec_t
	Fra_SecSetDefaultParams(&pars)
	pars.fSilent = 1
	if status := Fra_FraigSec(g, &pars, nil); status != 0 {
		t.Fatalf("status=%d", status)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IPo != 0 || saig.Saig_ManVerifyCexNoClear(g, cex) != 0 {
		t.Fatalf("source model=%+v", cex)
	}
	if status := Fra_FraigSecVerified(g, &pars, nil); status != 0 {
		t.Fatalf("verified status=%d", status)
	}
	cex = g.SequentialCounterexample()
	if cex == nil || cex.IPo != 1 || saig.Saig_ManVerifyCexNoClear(g, cex) == 0 {
		t.Fatalf("verified model=%+v", cex)
	}
}
