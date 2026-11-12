package abc

import "testing"

func TestSOPToCubesTruthTables(t *testing.T) {
	for _, sop := range []string{"10 1\n01 1\n", "10 0\n01 0\n"} {
		for _, xor := range []int32{0, 1} {
			n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
			a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
			a.SetName("a")
			b.SetName("b")
			node := n.NewObject(ObjectNode)
			node.Data = sop
			Abc_ObjAddFanin(node, a)
			Abc_ObjAddFanin(node, b)
			out := n.NewObject(ObjectPO)
			out.SetName("out")
			Abc_ObjAddFanin(out, node)
			converted := Abc_NtkSopToCubes(n, xor)
			if converted == nil {
				t.Fatal("cube conversion failed")
			}
			var eval func(*Object, map[*Object]bool) bool
			eval = func(o *Object, inputs map[*Object]bool) bool {
				if o.IsCI() {
					return inputs[o]
				}
				values := make([]bool, len(o.fanins))
				for i, e := range o.fanins {
					values[i] = eval(e.Object, inputs) != e.Complement
				}
				if o.IsCO() {
					return values[0]
				}
				return localFunction(o)(values)
			}
			inputs := converted.ObjectsOfKind(ObjectPI)
			for _, av := range []bool{false, true} {
				for _, bv := range []bool{false, true} {
					want := av != bv
					if sop[3] == '0' {
						want = !want
					}
					if got := eval(converted.ObjectsOfKind(ObjectPO)[0], map[*Object]bool{inputs[0]: av, inputs[1]: bv}); got != want {
						t.Fatalf("sop=%q xor=%d inputs %v,%v got %v", sop, xor, av, bv, got)
					}
				}
			}
		}
	}
}

func TestComputeDelay(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	first := Abc_NtkCreateNodeAnd(n, []*Object{a, b})
	second := Abc_NtkCreateNodeAnd(n, []*Object{first, b})
	out := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(out, second)
	if got := Abc_NtkComputeDelay(n); got != 2 {
		t.Fatalf("delay = %v, want 2", got)
	}
}

func TestSplitSopPreservesPositiveAndNegativeCovers(t *testing.T) {
	for _, phase := range []byte{'0', '1'} {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
		for _, name := range []string{"a", "b", "c"} {
			input := n.NewObject(ObjectPI)
			input.SetName(name)
		}
		node := n.NewObject(ObjectNode)
		node.Data = "000 " + string(phase) + "\n001 " + string(phase) + "\n010 " + string(phase) + "\n111 " + string(phase) + "\n"
		for _, input := range n.ObjectsOfKind(ObjectPI) {
			Abc_ObjAddFanin(node, input)
		}
		output := n.NewObject(ObjectPO)
		output.SetName("out")
		Abc_ObjAddFanin(output, node)
		split := Abc_NtkSplitSop(n, 2, 0)
		if split == nil || len(split.ObjectsOfKind(ObjectNode)) != 3 {
			t.Fatal("expected root and two split nodes")
		}
		var eval func(*Object, map[*Object]bool) bool
		eval = func(o *Object, inputs map[*Object]bool) bool {
			if o.IsCI() {
				return inputs[o]
			}
			values := make([]bool, len(o.fanins))
			for i, e := range o.fanins {
				values[i] = eval(e.Object, inputs) != e.Complement
			}
			if o.IsCO() {
				return values[0]
			}
			return localFunction(o)(values)
		}
		inputs := split.ObjectsOfKind(ObjectPI)
		for mask := 0; mask < 8; mask++ {
			values := make(map[*Object]bool)
			for i, input := range inputs {
				values[input] = mask&(1<<i) != 0
			}
			want := mask == 0 || mask == 4 || mask == 2 || mask == 7
			if phase == '0' {
				want = !want
			}
			if got := eval(split.ObjectsOfKind(ObjectPO)[0], values); got != want {
				t.Fatalf("phase=%c mask=%d got=%v want=%v", phase, mask, got, want)
			}
		}
	}
}
