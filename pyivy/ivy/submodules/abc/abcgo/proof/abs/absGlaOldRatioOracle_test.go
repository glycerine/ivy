package abs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

// This reaches refinement, rollback, and the source's ratio cutoff. The
// companion fixture runs the same network and options through ABC's C driver.
func TestOldGlaRatioLimitSourceOracle(t *testing.T) {
	for _, full := range []int32{0, 1} {
		for _, ratio := range []int32{0, 50, 100} {
			t.Run(fmt.Sprintf("full%d_ratio%d", full, ratio), func(t *testing.T) {
				g := gia.Gia_ManStart(20)
				var x [5]int32
				for i := range x {
					x[i] = gia.Gia_ManAppendCi(g)
				}
				left := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, x[0], x[1]), x[2])
				right := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, x[0]^1, x[3]), x[4])
				gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, left, right))
				p := NewParameters()
				p.FUseFullProof, p.FAddLayer, p.FPropFanout = full, 1, 0
				p.NRatioMin, p.NRatioMin2, p.NFramesMax, p.NTimeOut = ratio, 0, 3, 1
				status := Gia_ManPerformGlaOld(g, p, 0)
				var classes strings.Builder
				for _, c := range vec.Vec_IntArray(gia.Gia_ManGateClasses(g)) {
					fmt.Fprint(&classes, c)
				}
				got := fmt.Sprintf("SOURCE status=%d frame=%d stable=%d cex=%d classes=%s", status, p.IFrame, p.nFramesNoChange, boolToInt32(g.SequentialCounterexample() != nil), classes.String())
				want := "SOURCE status=-1 frame=3 stable=2 cex=0 classes=300000111110"
				if ratio == 100 {
					want = "SOURCE status=-1 frame=0 stable=0 cex=0 classes=100000111110"
				} else if full == 1 {
					want = "SOURCE status=-1 frame=3 stable=2 cex=0 classes=300000333330"
				}
				if got != want {
					t.Fatalf("got %s, want %s", got, want)
				}
				if oracle := os.Getenv("ABS_OLD_RATIO_ORACLE"); oracle != "" {
					out, err := exec.Command(oracle, fmt.Sprint(full), fmt.Sprint(ratio)).CombinedOutput()
					if err != nil || !strings.Contains(string(out), want+"\n") {
						t.Fatalf("C oracle: %v: %s", err, out)
					}
				}
			})
		}
	}
}
