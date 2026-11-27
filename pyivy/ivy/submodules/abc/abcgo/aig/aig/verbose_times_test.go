package aig

import (
	"regexp"
	"testing"
)

// Verbose reports end in ABC times ("T =%9.2f sec", "Runtime =%9.2f sec")
// after the same text as ABC; the ports printed Go durations.
func TestAigVerboseTimesLikeABC(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.AND(a, b), c))
	var p *Aig_ManCut_t
	out := captureAigOutput(t, func() { p = Aig_ComputeCuts(g, 8, 4, 1, 1) })
	Aig_ManCutStop(p)
	// one 64-entry chunk of nCutSize*nCutsMax = (24+4*4+4*1)*8 bytes
	want := `^Nodes = +[0-9]+\. Total cuts = +[0-9]+\. 4-input cuts = +[0-9]+\.\n` +
		`Cut size = 44\. Truth size =  4\. Total mem =  0\.02 MB  Runtime = +[0-9]+\.[0-9]{2} sec\n$`
	if !regexp.MustCompile(want).MatchString(out) {
		t.Errorf("Aig_ComputeCuts -v:\n%q", out)
	}

	out = captureAigOutput(t, func() { Aig_ManJustExperiment(g) })
	if !regexp.MustCompile(`\nPO = +1\. C0 = +[0-9]+\. C0f = +[0-9]+\. C1 = +[0-9]+\. C1f = +[0-9]+\. \( *[0-9.]+ %\) Ave = *[0-9.]+ T = +[0-9]+\.[0-9]{2} sec\n`).MatchString(out) {
		t.Errorf("Aig_ManJustExperiment:\n%q", out)
	}
}
