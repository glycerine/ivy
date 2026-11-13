package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSswBmcUnrollSequentialFrames(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	out := g.AddOutput(lo)
	g.AddOutput(g.XOR(pi, lo))
	aig.Aig_ManSetRegNum(g, 1)
	f := Ssw_FrmStart(g)
	defer Ssw_FrmStop(f)
	f.pFrames = aig.NewGraph()
	for frame := int32(0); frame < 4; frame++ {
		result := Ssw_BmcUnroll_rec(f, out, frame)
		if result != Ssw_BmcUnroll_rec(f, out, frame) {
			t.Fatal("cached frame changed")
		}
		f.pFrames.AddOutput(result)
	}
	if got := len(f.pFrames.Inputs()); got != 3 {
		t.Fatalf("unrolling four outputs needs three PI frames; got %d", got)
	}
	for bits := 0; bits < 8; bits++ {
		inputs := make([]bool, 3)
		for frame := int32(0); frame < 3; frame++ {
			e := Ssw_ObjFrame_(f, pi, frame)
			idx := int(aig.Aig_ObjCioId(f.pFrames, e))
			inputs[idx] = bits&(1<<frame) != 0
		}
		got, err := f.pFrames.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		state := false
		for frame := 0; frame < 4; frame++ {
			if got[frame] != state {
				t.Fatalf("bits=%03b frame=%d got=%v want=%v", bits, frame, got[frame], state)
			}
			if frame < 3 {
				state = state != (bits&(1<<frame) != 0)
			}
		}
	}
}
func TestSswBmcUnrollComplementedOutput(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	out := g.AddOutput(lo.Not())
	g.AddOutput(pi.Not())
	aig.Aig_ManSetRegNum(g, 1)
	f := Ssw_FrmStart(g)
	defer Ssw_FrmStop(f)
	f.pFrames = aig.NewGraph()
	first := Ssw_BmcUnroll_rec(f, out, 0)
	second := Ssw_BmcUnroll_rec(f, out, 1)
	f.pFrames.AddOutput(first)
	f.pFrames.AddOutput(second)
	for _, value := range []bool{false, true} {
		got, err := f.pFrames.Evaluate([]bool{value})
		if err != nil || !got[0] || got[1] != value {
			t.Fatalf("input %v: %v %v", value, got, err)
		}
	}
}
