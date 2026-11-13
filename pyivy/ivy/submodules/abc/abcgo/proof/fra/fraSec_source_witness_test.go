package fra

import (
	"math/rand"
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

// A checked-in-C Fra_FraigSec run on trial 1 of this generated fixture reaches
// separate interpolation and reports status=0, model output/frame 0/0, and a
// reduced graph with one register and three total outputs.
func TestSecSeparateInterpolationMatchesSourceWitness(t *testing.T) {
	random := rand.New(rand.NewSource(7411))
	var g *aig.Graph
	for trial := 0; trial < 2; trial++ {
		g = aig.NewGraph()
		edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 7; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		for i := 0; i < 4; i++ {
			e := edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				e = e.Not()
			}
			g.AddOutput(e)
		}
		aig.Aig_ManSetRegNum(g, 2)
	}
	var pars Fra_Sec_t
	Fra_SecSetDefaultParams(&pars)
	pars.fSilent, pars.fInterSeparate = 1, 1
	pars.fRetimeFirst, pars.fRetimeRegs, pars.fInduction, pars.fFraiging = 0, 0, 0, 0
	pars.fReachability, pars.fUsePdr = 0, 0
	var reduced *aig.Graph
	status := Fra_FraigSec(g, &pars, &reduced)
	cex := g.SequentialCounterexample()
	if status != 0 || cex == nil || cex.IPo != 0 || cex.IFrame != 0 || reduced == nil || reduced.RegisterCount() != 1 || reduced.OutputCount() != 3 {
		t.Fatalf("status=%d model=%+v reduced=%v", status, cex, reduced)
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

// fraSec.c sets bit i rather than nRegs+i in its fallback all-one model;
// Abc_CexDup then discards that initial-register bit on transfer.
func TestSecSourceTrivialWitnessUsesRawBitOffset(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.NewInput()
	g.AddOutput(x)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	var pars Fra_Sec_t
	Fra_SecSetDefaultParams(&pars)
	pars.fSilent = 1
	if status := Fra_FraigSec(g, &pars, nil); status != 0 {
		t.Fatalf("status=%d", status)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || len(cex.PData) == 0 || cex.PData[0]&3 != 0 || saig.Saig_ManVerifyCexNoClear(g, cex) != 0 {
		t.Fatalf("source transfers an all-zero, invalid witness: %+v", cex)
	}
}
