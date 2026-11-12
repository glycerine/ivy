package aig

import (
	"github.com/glycerine/abcgo/aig/ioa"
	"testing"
)

func TestAigerModelNormalizationAndRegisters(t *testing.T) {
	p := NewGraph()
	p.Name = "counter"
	a, state := p.NewInput(), p.NewInput()
	next := p.NewXor(a, state)
	Aig_ObjCreateCo(p, next.Not())
	Aig_ObjCreateCo(p, next)
	Aig_ManSetRegNum(p, 1)
	before := len(p.nodes)
	model, err := p.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Inputs) != 1 || len(model.Latches) != 1 || len(model.Outputs) != 1 || len(p.nodes) != before {
		t.Fatal(model)
	}
	data, err := ioa.Marshal(model, ioa.WriteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := ioa.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := decoded.Check(); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Latches) != 1 || decoded.Latches[0].Init != 0 {
		t.Fatal(decoded)
	}
}
func TestAigerConstantConvention(t *testing.T) {
	p := NewGraph()
	p.AddOutput(p.Const1())
	p.AddOutput(p.Const0())
	model, err := p.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	if model.Outputs[0] != 1 || model.Outputs[1] != 0 {
		t.Fatal(model.Outputs)
	}
}

func TestAigerRoundtripPreservesBooleanFunctions(t *testing.T) {
	p := NewGraph()
	a, state := p.NewInput(), p.NewInput()
	next := p.NewXor(a, state)
	Aig_ObjCreateCo(p, next.Not())
	Aig_ObjCreateCo(p, next)
	Aig_ManSetRegNum(p, 1)
	model, err := p.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	data, err := ioa.Marshal(model, ioa.WriteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	model, err = ioa.Read(data)
	if err != nil {
		t.Fatal(err)
	}
	for mask := 0; mask < 4; mask++ {
		inputs := []bool{mask&1 != 0, mask&2 != 0}
		want, err := p.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		values := map[uint32]bool{0: false}
		values[model.Inputs[0]>>1] = inputs[0]
		values[model.Latches[0].Literal>>1] = inputs[1]
		value := func(lit uint32) bool { return values[lit>>1] != (lit&1 != 0) }
		for _, gate := range model.ANDs {
			values[gate.Literal>>1] = value(gate.Left) && value(gate.Right)
		}
		got := []bool{value(model.Outputs[0]), value(model.Latches[0].Next)}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("assignment %d output %d: got %v want %v", mask, i, got[i], want[i])
			}
		}
	}
}
func TestAigerConstraintPolarity(t *testing.T) {
	p := NewGraph()
	a, b := p.NewInput(), p.NewInput()
	p.AddOutput(a)
	p.AddOutput(b.Not())
	p.nConstrs = 1
	model, err := p.AigerModel()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Outputs) != 0 || len(model.Bad) != 1 || len(model.Constraints) != 1 || model.Bad[0] != 2 || model.Constraints[0] != 4 {
		t.Fatal(model)
	}
}
