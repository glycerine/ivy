package abc

import "testing"

func TestLogicNetlistRoundTrip(t *testing.T) {
	src := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	a, b := src.NewObject(ObjectPI), src.NewObject(ObjectPI)
	a.SetName("a")
	b.SetName("b")
	node := Abc_NtkCreateNodeAnd(src, []*Object{a, b})
	output := src.NewObject(ObjectPO)
	output.SetName("result")
	Abc_ObjAddFanin(output, node)
	netlist := Abc_NtkToNetlist(src)
	if Abc_NtkCheck(netlist) == 0 {
		t.Fatal("invalid netlist")
	}
	logic := Abc_NtkToLogic(netlist)
	if Abc_NtkCheck(logic) == 0 {
		t.Fatal("invalid reconstructed logic")
	}
	inputs, outputs := logic.ObjectsOfKind(ObjectPI), logic.ObjectsOfKind(ObjectPO)
	if inputs[0].Name() != "a" || inputs[1].Name() != "b" || outputs[0].Name() != "result" {
		t.Fatalf("terminal names lost: %q %q %q", inputs[0].Name(), inputs[1].Name(), outputs[0].Name())
	}
	driver := outputs[0].fanins[0].Object
	fn := localFunction(driver)
	for _, v := range []struct{ a, b, want bool }{{false, false, false}, {false, true, false}, {true, false, false}, {true, true, true}} {
		if got := fn([]bool{v.a, v.b}); got != v.want {
			t.Fatalf("AND(%v,%v) = %v", v.a, v.b, got)
		}
	}
}

func TestAigToLogicPreservesComplementedEdges(t *testing.T) {
	src := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	a, b := src.NewObject(ObjectPI), src.NewObject(ObjectPI)
	a.SetName("a")
	b.SetName("b")
	and := src.AndEdges(a.Signal().Not(), b.Signal())
	output := src.NewObject(ObjectPO)
	output.SetName("result")
	Abc_ObjAddFanin(output, and.Not())
	logic := Abc_NtkAigToLogicSop(src)
	if Abc_NtkCheck(logic) == 0 {
		t.Fatal("invalid converted logic")
	}
	var eval func(*Object, map[*Object]bool) bool
	eval = func(o *Object, values map[*Object]bool) bool {
		if o.IsCI() {
			return values[o]
		}
		args := make([]bool, len(o.fanins))
		for i, e := range o.fanins {
			args[i] = eval(e.Object, values) != e.Complement
		}
		if o.IsCO() {
			return args[0]
		}
		return localFunction(o)(args)
	}
	inputs := logic.ObjectsOfKind(ObjectPI)
	for _, v := range []struct{ a, b bool }{{false, false}, {false, true}, {true, false}, {true, true}} {
		got := eval(logic.ObjectsOfKind(ObjectPO)[0], map[*Object]bool{inputs[0]: v.a, inputs[1]: v.b})
		if got != !(!v.a && v.b) {
			t.Fatalf("inputs %v,%v: got %v", v.a, v.b, got)
		}
	}
}

func TestAigBenchConversionsPreserveLogicAndSource(t *testing.T) {
	for _, nand := range []bool{false, true} {
		src := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
		a, b := src.NewObject(ObjectPI), src.NewObject(ObjectPI)
		a.SetName("a")
		b.SetName("b")
		first := src.AndEdges(a.Signal().Not(), b.Signal())
		second := src.AndEdges(first.Not(), a.Signal())
		for i, edge := range []Edge{first, second.Not()} {
			out := src.NewObject(ObjectPO)
			out.SetName([]string{"first", "second"}[i])
			Abc_ObjAddFanin(out, edge)
		}
		originalFirst, originalSecond := first.Object.Fanins(), second.Object.Fanins()
		var logic *Network
		if nand {
			logic = Abc_NtkAigToLogicSopNand(src)
		} else {
			logic = Abc_NtkAigToLogicSopBench(src)
		}
		if Abc_NtkCheck(logic) == 0 {
			t.Fatal("invalid bench conversion")
		}
		for i, edge := range first.Object.Fanins() {
			if edge != originalFirst[i] {
				t.Fatal("first source node was changed")
			}
		}
		for i, edge := range second.Object.Fanins() {
			if edge != originalSecond[i] {
				t.Fatal("second source node was changed")
			}
		}
		var eval func(*Object, map[*Object]bool) bool
		eval = func(o *Object, inputs map[*Object]bool) bool {
			if o.IsCI() {
				return inputs[o]
			}
			args := make([]bool, len(o.fanins))
			for i, edge := range o.fanins {
				args[i] = eval(edge.Object, inputs) != edge.Complement
			}
			if o.IsCO() {
				return args[0]
			}
			return localFunction(o)(args)
		}
		inputs, outputs := logic.ObjectsOfKind(ObjectPI), logic.ObjectsOfKind(ObjectPO)
		for _, v := range []struct{ a, b bool }{{false, false}, {false, true}, {true, false}, {true, true}} {
			firstValue := !v.a && v.b
			want := []bool{firstValue, !(!firstValue && v.a)}
			for i, out := range outputs {
				if got := eval(out, map[*Object]bool{inputs[0]: v.a, inputs[1]: v.b}); got != want[i] {
					t.Fatalf("nand=%v input=%v,%v output=%d got=%v want=%v", nand, v.a, v.b, i, got, want[i])
				}
			}
		}
	}
}
