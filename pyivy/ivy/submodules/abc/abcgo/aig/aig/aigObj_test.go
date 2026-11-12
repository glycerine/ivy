package aig

import "testing"

func TestObjectOutputPatchAndDelete(t *testing.T) {
	p := NewGraph()
	a, b := Aig_ObjCreateCi(p), Aig_ObjCreateCi(p)
	n := Aig_And(p, a, b)
	output := Aig_ObjCreateCo(p, n)
	Aig_ManFanoutStart(p)
	values, err := p.Evaluate([]bool{true, false})
	if err != nil || values[0] {
		t.Fatal(values, err)
	}
	Aig_ObjPatchFanin0(p, output, a.Not())
	if p.valid(n) {
		t.Fatal("unused cone retained")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	Aig_ManCheckPhase(p)
	values, err = p.Evaluate([]bool{false, true})
	if err != nil || !values[0] {
		t.Fatal(values, err)
	}
	Aig_ObjDeletePo(p, output)
	if len(p.outputs) != 0 || p.valid(output) {
		t.Fatal("output retained")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestObjectReplacementAndFanoutPropagation(t *testing.T) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	old := p.AND(a, b)
	root := p.AND(old, c)
	p.AddOutput(root)
	Aig_ManFanoutStart(p)
	Aig_ObjReplace(p, old, a.Not(), 1)
	values, err := p.Evaluate([]bool{false, false, true})
	if err != nil || !values[0] {
		t.Fatal(values, err)
	}
	Aig_ManPropagateBuffers(p, 1)
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	Aig_ManCheckPhase(p)
	for mask := 0; mask < 8; mask++ {
		out, err := p.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
		if err != nil || out[0] != (mask&1 == 0 && mask&4 != 0) {
			t.Fatal(out, err)
		}
	}
}
func TestExplicitXorObjects(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	e := p.NewXor(a, b)
	if p.node(e).Kind != ExclusiveOr {
		t.Fatal("XOR not retained")
	}
	p.AddOutput(e)
	for mask := 0; mask < 4; mask++ {
		out, err := p.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || out[0] != ((mask&1 != 0) != (mask&2 != 0)) {
			t.Fatal(out, err)
		}
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	Aig_ManCheckPhase(p)
}

func TestExplicitXorTruthAndDanglingCleanup(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	e := p.NewXor(a, b)
	table, err := Aig_ManCutTruth(p, e, []Edge{a, b})
	if err != nil || table[0]&15 != 6 {
		t.Fatal(table, err)
	}
	if Aig_ManCleanup(p) != 1 || len(p.xorHash) != 0 {
		t.Fatal("explicit XOR hash retained")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestExplicitConstantOutputCleanup(t *testing.T) {
	p := NewGraph()
	a := p.NewInput()
	zero := Aig_ObjCreateCo(p, p.Const0())
	Aig_ObjCreateCo(p, a)
	Aig_ManFanoutStart(p)
	if Aig_ManCoCleanup(p) != 1 || p.valid(zero) {
		t.Fatal("constant output retained")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	if Aig_ObjFanout0Int(p, int32(a.NodeID())) == 0 {
		t.Fatal("unary output fanout missing")
	}
}
