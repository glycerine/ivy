package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// sec_options_source.c compiled against ABC characterizes the early-return
// contract: simulation can report failure before assigning ppResult.
func TestSecGlobalBudgetAndResultSource(t *testing.T) {
	for _, tc := range []struct {
		name                                  string
		global, frames, status, output, frame int32
		result                                bool
	}{
		{"negative global stops before induction", -1, 1, -1, -1, -1, true},
		{"zero frames skips induction", 0, 0, 0, 0, 1, true},
		{"zero global permits induction", 0, 1, 0, 0, 1, false},
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
			p.FInduction = 1
			p.NFramesMax = tc.frames
			p.NBTLimitGlobal = tc.global
			p.fReachability = 0
			p.FUsePdr = 0
			p.FInterpolation = 1
			p.FInterSeparate = 0
			sentinel := aig.NewGraph()
			result := sentinel
			status := Fra_FraigSec(g, &p, &result)
			cex := g.SequentialCounterexample()
			if status != tc.status || (result != sentinel) != tc.result {
				t.Fatalf("status=%d result present=%v", status, result != sentinel)
			}
			if tc.output < 0 {
				if cex != nil {
					t.Fatalf("unexpected model %+v", cex)
				}
			} else if cex == nil || cex.IPo != tc.output || cex.IFrame != tc.frame {
				t.Fatalf("model=%+v want output=%d frame=%d", cex, tc.output, tc.frame)
			}
			if result != sentinel && (result.RegisterCount() != 2 || result.OutputCount() != 4) {
				t.Fatalf("result regs=%d outputs=%d", result.RegisterCount(), result.OutputCount())
			}
		})
	}
}
