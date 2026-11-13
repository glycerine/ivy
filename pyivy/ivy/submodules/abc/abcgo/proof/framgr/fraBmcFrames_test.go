package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestInitializedTimeframes(t *testing.T) {
	source := aig.NewGraph()
	input, latch := source.NewInput(), source.NewInput()
	source.AddOutput(latch)
	source.AddOutput(input)
	aig.Aig_ManSetRegNum(source, 1)
	graph, mapping := fraUnrollInitialized(source, 3, true)
	if graph.InputCount() != 3 || graph.OutputCount() != 3 {
		t.Fatalf("unexpected unrolled interface %d/%d", graph.InputCount(), graph.OutputCount())
	}
	for bits := 0; bits < 8; bits++ {
		values := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, err := graph.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{false, values[0], values[1]}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("inputs %v frame %d = %v, want %v", values, i, got[i], want[i])
			}
		}
	}
	if mapping[latch.NodeID()*3] != graph.Const0() {
		t.Fatal("initial latch was not zero")
	}
}

func TestInitializedTimeframesNativeXor(t *testing.T) {
	source := aig.NewGraph()
	input, latch := source.NewInput(), source.NewInput()
	source.AddOutput(latch)
	source.AddOutput(source.NewXor(latch, input))
	aig.Aig_ManSetRegNum(source, 1)
	graph, _ := fraUnrollInitialized(source, 3, true)
	for bits := 0; bits < 8; bits++ {
		values := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		got, err := graph.Evaluate(values)
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{false, values[0], values[0] != values[1]}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("input %v output %d = %v, want %v", values, i, got[i], want[i])
			}
		}
	}
}

func TestInitializedTimeframesTwoRegisterOracle(t *testing.T) {
	source := aig.NewGraph()
	x, y, q, r := source.NewInput(), source.NewInput(), source.NewInput(), source.NewInput()
	source.AddOutput(source.AND(q.Not(), x))
	source.AddOutput(source.NewXor(r, y))
	source.AddOutput(source.NewXor(q, y))
	source.AddOutput(source.AND(r.Not(), x))
	aig.Aig_ManSetRegNum(source, 2)
	graph, _ := fraUnrollInitialized(source, 4, true)
	for mask := 0; mask < 256; mask++ {
		inputs := make([]bool, 8)
		for i := range inputs {
			inputs[i] = mask&(1<<i) != 0
		}
		got, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		qv, rv := false, false
		for frame := 0; frame < 4; frame++ {
			xv, yv := inputs[2*frame], inputs[2*frame+1]
			want0, want1 := !qv && xv, rv != yv
			if got[2*frame] != want0 || got[2*frame+1] != want1 {
				t.Fatalf("mask %d frame %d: got %v, want %v %v", mask, frame, got[2*frame:2*frame+2], want0, want1)
			}
			qv, rv = qv != yv, !rv && xv
		}
	}
}
