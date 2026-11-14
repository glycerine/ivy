package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// Results characterize the checked-in C source with sec_options_source.c.
func TestSecInterpolationAndTimeLimitSourceOptions(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		inter, separate, budget int32
		timeLimit               int32
		status, output, frame   int32
		model                   bool
	}{
		{"separate", 1, 1, 10000, 0, 0, 0, 0, true},
		{"merged", 1, 0, 10000, 0, 0, 0, 1, true},
		{"interpolation disabled", 0, 0, 10000, 0, -1, -1, -1, false},
		{"limited separate", 1, 1, 1, 0, 0, 0, 0, true},
		{"limited merged", 1, 0, 1, 0, 0, 0, 1, true},
		{"negative merged budget", 1, 0, -1, 0, -1, -1, -1, false},
		{"source ignores time limit", 0, 0, 10000, -1, -1, -1, -1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			x, a, b := g.NewInput(), g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(x, a))
			g.AddOutput(g.AND(x, b))
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
			p.FInterpolation = tc.inter
			p.FInterSeparate = tc.separate
			p.NBTLimitInter = tc.budget
			p.TimeLimit = tc.timeLimit
			var result *aig.Graph
			status := Fra_FraigSec(g, &p, &result)
			cex := g.SequentialCounterexample()
			if status != tc.status || (cex != nil) != tc.model || result == nil || result.RegisterCount() != 2 || result.OutputCount() != 4 {
				t.Fatalf("status=%d cex=%+v result=%v", status, cex, result)
			}
			if cex != nil && (cex.IPo != tc.output || cex.IFrame != tc.frame) {
				t.Fatalf("cex=%+v want output=%d frame=%d", cex, tc.output, tc.frame)
			}
		})
	}
}
