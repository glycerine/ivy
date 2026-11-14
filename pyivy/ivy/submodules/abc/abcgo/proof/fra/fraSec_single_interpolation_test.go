package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

// Compiled sec_options_source.c with its single-output switch exercises the
// direct interpolation branch after sequential cleanup removes one register.
func TestSecSingleOutputInterpolationSource(t *testing.T) {
	for _, tc := range []struct {
		name                            string
		enable, separate, limit, status int32
		model                           bool
	}{
		{"single", 1, 0, 10000, 0, true},
		{"separate option ignored for single", 1, 1, 10000, 0, true},
		{"negative conflict limit", 1, 0, -1, -1, false},
		{"disabled", 0, 0, 10000, -1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(x, a))
			g.AddOutput(g.XOR(x, a))
			g.AddOutput(g.XOR(a, b))
			aig.Aig_ManSetRegNum(g, 2)
			var p Fra_Sec_t
			Fra_SecSetDefaultParams(&p)
			p.fSilent = 1
			p.FRetimeFirst = 0
			p.FRetimeRegs = 0
			p.FFraiging = 0
			p.FInduction = 0
			p.fReachability = 0
			p.FUsePdr = 0
			p.FInterpolation = tc.enable
			p.FInterSeparate = tc.separate
			p.NBTLimitInter = tc.limit
			var reduced *aig.Graph
			status := Fra_FraigSec(g, &p, &reduced)
			cex := g.SequentialCounterexample()
			if status != tc.status || (cex != nil) != tc.model || reduced == nil || reduced.RegisterCount() != 1 || reduced.OutputCount() != 2 {
				t.Fatalf("status=%d model=%+v reduced=%v", status, cex, reduced)
			}
			if cex != nil && (cex.IPo != 0 || cex.IFrame != 1 || saig.Saig_ManVerifyCexNoClear(g, cex) == 0) {
				t.Fatalf("model=%+v", cex)
			}
		})
	}
}
