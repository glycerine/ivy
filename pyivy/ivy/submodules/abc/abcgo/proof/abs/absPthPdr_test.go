package abs

import (
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestAbsProverRunsPDR(t *testing.T) {
	for _, bad := range []bool{false, true} {
		g := aig.NewGraph()
		g.NewInput()
		g.NewInput()
		out := g.Const0()
		if bad {
			out = g.Const1()
		}
		g.AddOutput(out)
		g.AddOutput(g.Const0())
		aig.Aig_ManSetRegNum(g, 1)
		done := make(chan int32, 1)
		absProver.start(func(id int32) int32 { result := absRunPDR(id, g, 0); done <- result; return result })
		select {
		case got := <-done:
			want := int32(1)
			if bad {
				want = 0
			}
			if got != want {
				t.Fatalf("bad=%v: PDR status=%d want=%d", bad, got, want)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("PDR did not finish")
		}
	}
}
