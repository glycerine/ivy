package abc

import "testing"

func TestMuxRecognitionAndParity(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	c, a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	manager := n.FunctionManager.(*Abc_Aig_t)
	mux := Abc_AigMux(manager, c, a, b)
	var thenEdge, elseEdge Edge
	control := Abc_NodeRecognizeMux(mux.Object, &thenEdge, &elseEdge)
	if control != c || Abc_NodeIsMuxType(mux.Object) != 1 || Abc_NodeIsMuxControlType(c) != 1 {
		t.Fatal("mux structure not recognized")
	}
	for mask := 0; mask < 8; mask++ {
		inputs := map[*Object]bool{c: mask&1 != 0, a: mask&2 != 0, b: mask&4 != 0}
		tv, err := n.Evaluate(thenEdge, inputs)
		if err != nil {
			t.Fatal(err)
		}
		ev, err := n.Evaluate(elseEdge, inputs)
		if err != nil {
			t.Fatal(err)
		}
		got, err := n.Evaluate(mux.Object.Signal(), inputs)
		if err != nil {
			t.Fatal(err)
		}
		want := ev
		if inputs[control] {
			want = tv
		}
		if got != want {
			t.Fatalf("mask %d: got %v want %v", mask, got, want)
		}
	}
	xor := Abc_AigXor(manager, a, b)
	if Abc_NodeIsExorType(xor.Object) != 1 || Abc_NtkGetExorNum(n) != 1 || Abc_NtkCountMuxes(n) < 2 {
		t.Fatal("XOR structure not recognized")
	}
}
