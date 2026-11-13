package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
)

// The C derivation walks pNext in chronological order, copies positive PI
// literals, and ignores abstraction PPIs. This trace crosses word boundaries.
func TestDeriveCexMultiFrameWordBoundariesAndPPI(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 33)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	first, second := g.NewInput(), g.NewInput()
	g.AddOutput(second)
	g.AddOutput(inputs[0])
	g.AddOutput(first)
	aig.Aig_ManSetRegNum(g, 2)
	p := Pdr_ManStart(g, &Pdr_Par_t{}, nil)
	defer Pdr_ManStop(p)
	last := Pdr_OblStart(2, 0, testPdrSet(nil, []int32{64, 66}), nil)
	middle := Pdr_OblStart(1, 0, testPdrSet(nil, []int32{1, 65}), last)
	firstFrame := Pdr_OblStart(0, 0, testPdrSet(nil, []int32{0, 64, 66}), middle)
	Pdr_QueuePush(p, firstFrame)
	cex := Pdr_ManDeriveCex(p)
	if cex.IFrame != 2 || cex.NPis != 33 || cex.NRegs != 2 || saig.Saig_ManVerifyCex(g, cex) != 1 {
		t.Fatalf("invalid trace: %+v", cex)
	}
	for frame := int32(0); frame < 3; frame++ {
		for input := int32(0); input < cex.NPis; input++ {
			want := frame == 0 && (input == 0 || input == 32) || frame == 2 && input == 32
			if got := util.Abc_InfoHasBit(cex.PData, cex.NRegs+frame*cex.NPis+input); got != want {
				t.Fatalf("frame %d input %d: got %v, want %v", frame, input, got, want)
			}
		}
	}
}
