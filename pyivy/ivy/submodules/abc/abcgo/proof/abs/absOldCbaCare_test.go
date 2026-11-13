package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func TestCbaReasonsMapAndDeduplicate(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.NewInput()
	g.NewInput()
	frames := aig.NewGraph()
	frames.NewInput()
	frames.NewInput()
	frames.NewInput()
	p := &Saig_ManCba_t{pAig: g, pFrames: frames, nInputs: 1, pCex: util.Abc_CexAlloc(0, 3, 2), vMapPiF2A: vec.Vec_IntAllocArrayCopy([]int32{2, 0, 1, 1, 2, 1}, 6)}
	reasons := vec.Vec_IntAllocArrayCopy([]int32{0, 1, 2}, 3)
	if got := vec.Vec_IntArray(Saig_ManCbaReason2Inputs(p, reasons)); !reflect.DeepEqual(got, []int32{1, 0}) {
		t.Fatalf("inputs %v", got)
	}
	care := Saig_ManCbaReason2Cex(p, reasons)
	for i := int32(0); i < 6; i++ {
		want := i == 2 || i == 4 || i == 5
		if util.Abc_InfoHasBit(care.PData, i) != want {
			t.Fatalf("care bit %d", i)
		}
	}
}
