package abs

import (
	"strconv"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// With -a and every pseudo-input a PI, the C layer strategy added an empty
// set: with incremental proofs it divided by zero in Prf_ManSize, with full
// proofs (-p) it repeated the SAT query until the timeout. After the fix
// (CBUGS_FIXED.md #35) the counter-example is checked as without -a.
func TestOldGlaEmptyLayerChecksCounterexample(t *testing.T) {
	for _, full := range []int32{0, 1} {
		t.Run(strconv.Itoa(int(full)), func(t *testing.T) {
			g := gia.Gia_ManStart(100)
			in := make([]int32, 6)
			for i := range in {
				in[i] = gia.Gia_ManAppendCi(g)
			}
			root := in[0]
			for _, input := range in[1:] {
				root = gia.Gia_ManAppendAnd(g, root, input)
			}
			gia.Gia_ManAppendCo(g, root)
			var p Abs_Par_t
			Abs_ParSetDefaults(&p)
			p.FAddLayer = 1
			p.FUseFullProof = full
			p.NFramesMax = 1
			p.NTimeOut = 1
			p.NRatioMin = 0
			p.NRatioMin2 = 0
			got := Gia_ManPerformGlaOld(g, &p, 0)
			if got != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) == 0 {
				t.Fatalf("result=%d cex=%v, want a verified witness", got, g.SequentialCounterexample())
			}
		})
	}
}
