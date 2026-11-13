package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// fraSim.c asserts that a frame model identifies one of the unrolled COs.
func TestCopyCounterExampleRequiresFrameOutputID(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	frames := aig.NewGraph()
	frames.NewInput()
	frames.AddOutput(frames.Const0())
	defer func() {
		if got := recover(); got != "fra: counterexample output ID not found" {
			t.Fatalf("panic = %v, want missing output ID assertion", got)
		}
	}()
	Fra_SmlCopyCounterExample(g, frames, []int32{0, 999})
}

// fraSim.c asserts the caller has detected a nonzero output first.
func TestGetCounterExampleRequiresNonconstantOutputFlag(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const0())
	s := Fra_SmlStart(g, 0, 1, 1)
	defer func() {
		if got := recover(); got != "fra: no nonconstant output detected" {
			t.Fatalf("panic = %v, want missing output flag assertion", got)
		}
	}()
	Fra_SmlGetCounterExample(s)
}

func TestCopyCounterExampleRequiresSequentialSource(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	frames := aig.NewGraph()
	frames.NewInput()
	out := frames.AddOutput(frames.Const0())
	defer func() {
		if got := recover(); got != "fra: invalid counterexample frames or model" {
			t.Fatalf("panic = %v, want invalid frame assertion", got)
		}
	}()
	Fra_SmlCopyCounterExample(g, frames, []int32{0, int32(out.NodeID())})
}
