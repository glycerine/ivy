package abc

import "testing"

func TestLogicStrash(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 0)
	a, b := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	node := n.NewObject(ObjectNode)
	node.Data = "10 1\n"
	Abc_ObjAddFanin(node, a)
	Abc_ObjAddFanin(node, b)
	out := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(out, node)
	Abc_NtkOrderCisCos(n)
	dst := Abc_NtkStrash(n, 0, 1, 0)
	if Abc_NtkIsStrash(dst) == 0 {
		t.Fatal("not strashed")
	}
	inputs := dst.ObjectsOfKind(ObjectPI)
	output := dst.ObjectsOfKind(ObjectPO)[0]
	for x := 0; x < 4; x++ {
		values := map[*Object]bool{inputs[0]: x&1 != 0, inputs[1]: x&2 != 0}
		got, err := dst.Evaluate(output.Signal(), values)
		if err != nil || got != (x == 1) {
			t.Fatalf("input %d result %v err %v", x, got, err)
		}
	}
}
