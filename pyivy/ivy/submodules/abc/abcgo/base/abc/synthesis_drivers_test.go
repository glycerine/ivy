package abc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	abcmain "github.com/glycerine/abcgo/base/main"
)

type synthesisTestDispatcher struct{ command string }

func (d *synthesisTestDispatcher) Execute(_ *abcmain.Frame, command string) int32 {
	d.command = command
	return 0
}

func TestSopSynthesizeConstant(t *testing.T) {
	for _, tc := range []struct {
		sop string
		lit int32
	}{{" 0\n", 0}, {" 1\n", 1}} {
		g := Abc_SopSynthesizeOne(tc.sop, 0)
		if name, _ := g.Identity(); name != "top" {
			t.Fatalf("name = %q", name)
		}
		co := gia.Gia_ManCo(g, 0)
		if gia.Gia_ManCoNum(g) != 1 || gia.Gia_ObjFaninLit0(co, gia.Gia_ObjId(g, co)) != tc.lit {
			t.Fatalf("wrong constant for %q", tc.sop)
		}
		gia.Gia_ManStop(g)
	}
}

func TestGiaSynthesizeInterUsesCollapse(t *testing.T) {
	prior := abcmain.Abc_FrameReadGlobalFrame()
	defer abcmain.Abc_FrameSetGlobalFrame(prior)
	collapseGia.RLock()
	old := collapseGia.fn
	collapseGia.RUnlock()
	defer RegisterCollapseGia(old)
	frame := abcmain.NewFrame()
	dispatch := &synthesisTestDispatcher{}
	frame.SetDispatcher(dispatch)
	abcmain.Abc_FrameSetGlobalFrame(frame)
	g := gia.Gia_ManStart(4)
	a := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	called := false
	RegisterCollapseGia(func(n *Network) *gia.Gia_Man_t {
		called = true
		if Abc_NtkIsStrash(n) == 0 || Abc_NtkPiNum(n) != 1 || Abc_NtkPoNum(n) != 1 {
			t.Fatal("incorrect synthesis input")
		}
		return g
	})
	if Abc_GiaSynthesizeInter(g) != g || !called || dispatch.command != "balance; collapse; muxes; strash; dc2" {
		t.Fatal("intermediate synthesis did not run")
	}
}
