package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

func TestBMCOutcomesWithAndWithoutRewrite(t *testing.T) {
	for _, rewrite := range []int32{0, 1} {
		for _, mode := range []string{"constant true", "constant false", "sequential witness"} {
			t.Run(mode+string(rune('0'+rewrite)), func(t *testing.T) {
				g := aig.NewGraph()
				x, y, q := g.NewInput(), g.NewInput(), g.NewInput()
				switch mode {
				case "constant true":
					g.AddOutput(g.Const1())
				case "constant false":
					g.AddOutput(g.Const0())
				default:
					g.AddOutput(q)
				}
				g.AddOutput(g.AND(x, y))
				aig.Aig_ManSetRegNum(g, 1)
				Fra_BmcPerformSimple(g, 2, 1000, rewrite, 0)
				cex := g.SequentialCounterexample()
				if mode == "constant false" {
					if cex != nil {
						t.Fatalf("unexpected witness %+v", cex)
					}
					return
				}
				if cex == nil {
					t.Fatal("missing witness")
				}
				if mode == "constant true" && (cex.IPo != 0 || cex.IFrame != 0) {
					t.Fatalf("trivial witness %+v", cex)
				}
				if mode == "sequential witness" && (cex.IPo != 0 || cex.IFrame != 1) {
					t.Fatalf("sequential witness %+v", cex)
				}
				if saig.Saig_ManVerifyCexNoClear(g, cex) != 1 {
					t.Fatalf("witness does not replay: %+v", cex)
				}
			})
		}
	}
}
