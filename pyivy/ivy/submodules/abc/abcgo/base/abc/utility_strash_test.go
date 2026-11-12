package abc

import "testing"

func TestStrashToGiaTruth(t *testing.T) {
	n := NewNetwork("strash")
	n.Type = uint8(ABC_NTK_STRASH)
	m := Abc_AigAlloc(n)
	a, b := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	edge := Abc_AigAnd(m, a, Abc_ObjNot(b))
	o := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(o, edge.Not())
	Abc_NtkOrderCisCos(n)
	g := Abc_NtkStrashToGia(n)
	for i := 0; i < 4; i++ {
		inputs := []bool{i&1 != 0, i&2 != 0}
		out, err := g.Evaluate(inputs)
		if err != nil || out[0] != !(inputs[0] && !inputs[1]) {
			t.Fatalf("input %v output %v err %v", inputs, out, err)
		}
	}
}
