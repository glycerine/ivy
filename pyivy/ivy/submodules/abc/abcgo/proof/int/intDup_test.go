package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestInitialStateGraph(t *testing.T) {
	g := Inter_ManStartInitState(3)
	for mask := 0; mask < 8; mask++ {
		values := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		outputs, err := g.Evaluate(values)
		if err != nil || len(outputs) != 1 || outputs[0] != (mask == 0) {
			t.Fatalf("inputs %03b got %v, err %v", mask, outputs, err)
		}
	}
}

func TestTransitionDuplication(t *testing.T) {
	src := aig.NewGraph()
	x, lo := src.NewInput(), src.NewInput()
	src.AddOutput(x)
	src.AddOutput(x.Not()) // constraint, inverted by the duplication routines
	src.AddOutput(src.XOR(x, lo))
	aig.Aig_ManSetRegNum(src, 1)
	aig.Aig_ManSetConstrNum(src, 1)

	for _, mode := range []string{"duplicate", "loop", "first"} {
		var got *aig.Graph
		switch mode {
		case "duplicate":
			got = Inter_ManStartDuplicated(src)
		case "loop":
			got = Inter_ManStartOneOutput(src, 0)
		case "first":
			got = Inter_ManStartOneOutput(src, 1)
		}
		for mask := 0; mask < 8; mask++ {
			xv, cv, lv := mask&1 != 0, mask&2 != 0, mask&4 != 0
			inputs := []bool{xv, lv}
			if mode != "duplicate" {
				inputs = []bool{xv, cv, lv}
			}
			outputs, err := got.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			wantSecond := xv != lv
			if mode == "loop" && cv {
				wantSecond = lv
			}
			if mode == "first" {
				wantSecond = xv
			}
			if len(outputs) != 2 || outputs[0] != xv || outputs[1] != wantSecond {
				t.Fatalf("%s inputs %v: got %v want [%v %v]", mode, inputs, outputs, xv, wantSecond)
			}
		}
		wantRegs := int32(1)
		if mode == "first" {
			wantRegs = 0
		}
		if got.RegisterCount() != wantRegs {
			t.Fatalf("%s register count %d", mode, got.RegisterCount())
		}
	}
}
