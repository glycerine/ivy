package abs

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// The C layer strategy repeats a SAT query when no pseudo-input remains;
// it neither performs counterexample refinement nor makes further progress.
func TestOldGlaEmptyLayerWaitsForTimeout(t *testing.T) {
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
			if got != -1 || g.SequentialCounterexample() != nil {
				t.Fatalf("result=%d cex=%v, want source unknown without witness", got, g.SequentialCounterexample())
			}
			if binary := os.Getenv("ABS_OLD_EMPTY_LAYER_ORACLE"); binary != "" {
				out, err := exec.Command(binary, strconv.Itoa(int(full))).CombinedOutput()
				if err != nil || !strings.Contains(string(out), "SOURCE_STATUS=-1 CEX=0") {
					t.Fatalf("C oracle err=%v output=%s", err, out)
				}
			}
		})
	}
}
