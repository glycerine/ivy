package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCbaFlopRankingUsesTopologicalOrder(t *testing.T) {
	g := aig.NewGraph()
	pi, q0, q1 := g.NewInput(), g.NewInput(), g.NewInput()
	early := g.AND(pi, q0)
	later := g.AND(q0, q1)
	aig.Aig_ObjDisconnect(g, early)
	aig.Aig_ObjConnect(g, early, later, pi)
	g.AddOutput(g.Const0())
	g.AddOutput(early)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 2)
	output, err := g.Evaluate([]bool{true, true, true})
	if err != nil || !output[1] {
		t.Fatalf("rewired graph: outputs=%v error=%v", output, err)
	}
	cex := util.Abc_CexAlloc(0, 3, 2)
	cex.IFrame = 1
	for _, bit := range []int32{0, 1, 2, 4, 5} {
		util.Abc_InfoSetBit(cex.PData, bit)
	}
	classes := vec.Vec_IntStart(2)
	candidates := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(candidates, 0)
	vec.Vec_IntPush(candidates, 1)
	result := Saig_ManCbaFilterFlops(g, cex, classes, candidates, 1)
	if got := vec.Vec_IntEntry(result, 0); got != 1 {
		t.Fatalf("selected candidate=%d, want 1", got)
	}
}
