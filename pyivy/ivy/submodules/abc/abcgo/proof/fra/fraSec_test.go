package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

func TestSecDefaultsAndTrivialMiter(t *testing.T) {
	var opts Fra_Sec_t
	Fra_SecSetDefaultParams(&opts)
	if opts.nFramesMax != 4 || opts.nBTLimit != 1000 || opts.fUsePdr != 1 || opts.fInterpolation != 1 {
		t.Fatalf("defaults %+v", opts)
	}
	for _, fails := range []bool{false, true} {
		g := aig.NewGraph()
		g.NewInput()
		g.NewInput()
		g.AddOutput(g.Const0())
		if fails {
			g.AddOutput(g.Const1())
		} else {
			g.AddOutput(g.Const0())
		}
		g.AddOutput(g.Const0())
		aig.Aig_ManSetRegNum(g, 1)
		opts.fSilent = 1
		got := Fra_FraigSec(g, &opts, nil)
		want := int32(1)
		if fails {
			want = 0
		}
		if got != want {
			t.Fatalf("fails=%v got=%d want=%d", fails, got, want)
		}
		if fails {
			cex := g.SequentialCounterexample()
			if cex == nil || cex.IPo != 1 || saig.Saig_ManVerifyCexNoClear(g, cex) == 0 {
				t.Fatalf("invalid witness %+v", cex)
			}
		}
	}
}
