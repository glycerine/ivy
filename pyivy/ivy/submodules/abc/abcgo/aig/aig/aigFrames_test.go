package aig

import "testing"

func TestFrameExpansionInitialized(t *testing.T) {
	p := NewGraph()
	input, state := p.NewInput(), p.NewInput()
	p.AddOutput(state)
	p.AddOutput(p.NewXor(input, state))
	Aig_ManSetRegNum(p, 1)
	var mapping []Edge
	q := Aig_ManFrames(p, 3, 1, 1, 0, 0, &mapping)
	if q.InputCount() != 3 || q.OutputCount() != 3 {
		t.Fatal("wrong expansion terminals")
	}
	for mask := 0; mask < 8; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
		out, err := q.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{false, inputs[0], inputs[0] != inputs[1]}
		for i := range want {
			if out[i] != want[i] {
				t.Fatalf("assignment %d frame %d", mask, i)
			}
		}
	}
	if Aig_ObjFrames(mapping, 3, state, 0) != q.Const0() {
		t.Fatal("initial state map")
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestFrameExpansionFreeStateAndEnlarge(t *testing.T) {
	p := NewGraph()
	input, state := p.NewInput(), p.NewInput()
	p.AddOutput(state)
	p.AddOutput(p.NewXor(input, state))
	Aig_ManSetRegNum(p, 1)
	q := Aig_ManFrames(p, 2, 0, 1, 1, 1, nil)
	if q.InputCount() != 3 || q.OutputCount() != 2 || q.nRegs != 1 {
		t.Fatal("wrong enlarged terminals")
	}
	if q.TruePiCount() != 2 || q.TruePoCount() != 1 || Aig_ObjCioId(q, q.inputs[1]) != 1 || Aig_ObjCioId(q, q.outputs[1]) != 1 {
		t.Fatal("frame expansion did not finalize terminal metadata")
	}
	for mask := 0; mask < 8; mask++ {
		a, b, s := mask&1 != 0, mask&2 != 0, mask&4 != 0
		out, err := q.Evaluate([]bool{a, b, s})
		if err != nil || out[0] != (a != s) || out[1] != (a != s) {
			t.Fatal(out, err)
		}
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestInitializedFrameWithLatchOutputsIsCombinational(t *testing.T) {
	p := NewGraph()
	state := p.NewInput()
	p.AddOutput(state)
	p.AddOutput(state.Not())
	Aig_ManSetRegNum(p, 1)
	q := Aig_ManFrames(p, 1, 1, 0, 1, 0, nil)
	if q.nRegs != 0 || q.InputCount() != 0 || q.OutputCount() != 1 {
		t.Fatal("initialized expansion invented state inputs")
	}
	if err := q.Check(); err != nil {
		t.Fatal(err)
	}
	out, err := q.Evaluate(nil)
	if err != nil || !out[0] {
		t.Fatal(out, err)
	}
}
