package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSignalCorrespondenceWithoutGates(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	pars := &Ssw_Pars_t{}
	out := Ssw_SignalCorrespondence(g, pars)
	if out == nil || aig.Aig_ManRegNum(out) != 1 || pars.nIters != 0 || !aig.Aig_ManHasReprs(g) {
		t.Fatal("empty graph correspondence failed")
	}
}
func TestCorrespondenceCallback(t *testing.T) {
	n := 0
	p := &Ssw_Pars_t{pData: &n, pFunc: func(data any) int32 { *data.(*int)++; return 0 }}
	sswInvokeCallback(p)
	if n != 1 {
		t.Fatal("callback was not invoked")
	}
}
