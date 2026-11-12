package abc

import "testing"

func TestSimpleOutputsRepair(t *testing.T) {
	for _, duplicate := range []int32{0, 1} {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
		input := n.NewObject(ObjectPI)
		input.SetName("input")
		driver := Abc_NtkCreateNodeBuf(n, input)
		for _, spec := range []struct {
			name string
			edge Edge
		}{
			{"first", driver.Signal()},
			{"second", driver.Signal()},
			{"inverse", driver.Signal().Not()},
			{"direct", input.Signal()},
		} {
			output := n.NewObject(ObjectPO)
			output.SetName(spec.name)
			Abc_ObjAddFanin(output, spec.edge)
		}
		if Abc_NtkLogicHasSimpleCos(n) != 0 {
			t.Fatal("expected non-simple outputs")
		}
		if got := Abc_NtkLogicMakeSimpleCos2(n, duplicate); got != 3 {
			t.Fatalf("duplicate %d: repaired %d outputs, want 3", duplicate, got)
		}
		if Abc_NtkLogicHasSimpleCos(n) != 1 {
			t.Fatal("outputs remain non-simple")
		}
		for i, output := range n.ObjectsOfKind(ObjectPO) {
			node := output.fanins[0].Object
			if output.fanins[0].Complement {
				t.Fatal("output edge remains complemented")
			}
			if i > 0 && localFunction(node)([]bool{true}) != (i != 2) {
				t.Fatalf("output %d has incorrect polarity", i)
			}
		}
	}
}

func TestMakeSimpleCosPreservesFunctions(t *testing.T) {
	for _, duplicate := range []int32{0, 1} {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
		input := n.NewObject(ObjectPI)
		input.SetName("input")
		driver := Abc_NtkCreateNodeBuf(n, input)
		other := Abc_NtkCreateNodeBuf(n, driver)
		for _, spec := range []struct {
			name string
			edge Edge
		}{
			{"negated", driver.Signal().Not()},
			{"other", other.Signal()},
			{"other2", other.Signal()},
			{"constant", Abc_NtkCreateNodeConst0(n).Signal().Not()},
		} {
			out := n.NewObject(ObjectPO)
			out.SetName(spec.name)
			Abc_ObjAddFanin(out, spec.edge)
		}
		if got := Abc_NtkLogicMakeSimpleCos(n, duplicate); got != 1 {
			t.Fatalf("duplicate %d: added %d nodes, want 1", duplicate, got)
		}
		if Abc_NtkLogicHasSimpleCos(n) != 1 {
			t.Fatal("outputs remain non-simple")
		}
		var eval func(*Object, bool) bool
		eval = func(obj *Object, value bool) bool {
			if obj.IsCI() {
				return value
			}
			args := make([]bool, len(obj.fanins))
			for i, edge := range obj.fanins {
				args[i] = eval(edge.Object, value) != edge.Complement
			}
			if obj.IsCO() {
				return args[0]
			}
			return localFunction(obj)(args)
		}
		for _, value := range []bool{false, true} {
			want := []bool{!value, value, value, true}
			for i, output := range n.ObjectsOfKind(ObjectPO) {
				if got := eval(output, value); got != want[i] {
					t.Fatalf("duplicate %d output %d input %v: got %v", duplicate, i, value, got)
				}
			}
		}
	}
}

func TestMakeSimpleCosSharesInvertedDriver(t *testing.T) {
	for _, duplicate := range []int32{0, 1} {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
		input := n.NewObject(ObjectPI)
		input.SetName("in")
		driver := Abc_NtkCreateNodeBuf(n, input)
		for i, inverted := range []bool{false, true, true} {
			out := n.NewObject(ObjectPO)
			out.SetName([]string{"positive", "negative", "negative"}[i])
			Abc_ObjAddFanin(out, Edge{Object: driver, Complement: inverted})
		}
		if got := Abc_NtkLogicMakeSimpleCos(n, duplicate); got != 1 {
			t.Fatalf("duplicate %d: added %d nodes, want 1", duplicate, got)
		}
		outputs := n.ObjectsOfKind(ObjectPO)
		if outputs[1].fanins[0].Object != outputs[2].fanins[0].Object || Abc_NtkLogicHasSimpleCos(n) == 0 {
			t.Fatal("inverted outputs do not share a simple driver")
		}
		if got := localFunction(outputs[1].fanins[0].Object)([]bool{true}); got {
			t.Fatal("inverted driver has wrong polarity")
		}
	}
}
