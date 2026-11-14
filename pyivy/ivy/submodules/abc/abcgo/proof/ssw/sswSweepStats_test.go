package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestSweepStatsUseSourceIterationDeltas(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 0)
	classes := Ssw_ClassesStart(g)
	defer Ssw_ClassesStop(classes)
	Ssw_ObjSetConst1Cand(g, x)
	classes.nCands1 = 1
	p := &Ssw_Man_t{pAig: g, pFrames: g, ppClasses: classes, pPars: &Ssw_Pars_t{FDynamic: 1}, nSatProof: 12, nSatCallsSat: 15, nRecycles: 8, nSatFailsReal: 9, nConstrReduced: 7}
	got := sswSweepStats(p, 2, 10, 11, 5, 3)
	want := "  2 : C =      1. Cl =      0. LR =     7. NR =     0. Cex =    4. R =   3. F =    6. + "
	if got != want {
		t.Fatalf("dynamic stats = %q, want %q", got, want)
	}
	p.pPars.FLatchCorrOpt = 1
	got = sswSweepStats(p, 2, 10, 11, 5, 3)
	want = "  2 : C =      1. Cl =      0. Pr =     2. Cex =    4. R =   3. F =   6. "
	if got != want {
		t.Fatalf("latch stats = %q, want %q", got, want)
	}
}
