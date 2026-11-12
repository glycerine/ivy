package abc

import "testing"

func TestAddBuffersPreservesLogicAndAlignsLevels(t *testing.T) {
	for _, reverse := range []int32{0, 1} {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
		a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
		a.SetName("a")
		b.SetName("b")
		first := Abc_NtkCreateNodeAnd(n, []*Object{a, b})
		second := Abc_NtkCreateNodeAnd(n, []*Object{first, a})
		for i, driver := range []*Object{first, second} {
			out := n.NewObject(ObjectPO)
			out.SetName([]string{"short", "long"}[i])
			Abc_ObjAddFanin(out, driver)
		}
		buffered := Abc_NtkAddBuffsInt(n, reverse, 0, 0)
		if len(buffered.ObjectsOfKind(ObjectNode)) <= 2 {
			t.Fatalf("reverse=%d: expected buffers", reverse)
		}
		var eval func(*Object, map[*Object]bool) bool
		eval = func(o *Object, inputs map[*Object]bool) bool {
			if o.IsCI() {
				return inputs[o]
			}
			args := make([]bool, len(o.fanins))
			for i, e := range o.fanins {
				args[i] = eval(e.Object, inputs) != e.Complement
			}
			if o.IsCO() {
				return args[0]
			}
			return localFunction(o)(args)
		}
		inputs := buffered.ObjectsOfKind(ObjectPI)
		for _, aValue := range []bool{false, true} {
			for _, bValue := range []bool{false, true} {
				for _, out := range buffered.ObjectsOfKind(ObjectPO) {
					if got := eval(out, map[*Object]bool{inputs[0]: aValue, inputs[1]: bValue}); got != (aValue && bValue) {
						t.Fatalf("reverse=%d input=%v,%v got=%v", reverse, aValue, bValue, got)
					}
				}
			}
		}
		for _, node := range buffered.ObjectsOfKind(ObjectNode) {
			for _, edge := range node.fanins {
				if node.Level != edge.Object.Level+1 {
					t.Fatalf("unbalanced edge %d -> %d", edge.Object.ID, node.ID)
				}
			}
		}
	}
}
