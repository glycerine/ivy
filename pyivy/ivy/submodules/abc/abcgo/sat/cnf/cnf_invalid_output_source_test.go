package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// The original C writer asserts for either value on this two-output
// combinational graph (verified with the direct Cnf_Derive harness).
func TestSourceRejectsInvalidFreeOutputCounts(t *testing.T) {
	for _, n := range []int32{-1, 1} {
		t.Run(map[int32]string{-1: "negative", 1: "partial"}[n], func(t *testing.T) {
			g := aig.NewGraph()
			x := g.NewInput()
			g.AddOutput(x)
			g.AddOutput(x.Not())
			defer func() {
				if recover() == nil {
					t.Error("expected source assertion")
				}
			}()
			Cnf_Derive(g, n)
		})
	}
}
