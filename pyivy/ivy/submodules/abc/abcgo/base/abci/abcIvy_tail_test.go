package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/ivy"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func evalIvyRoundtrip(network *abc.Network, object *abc.Object, assignment map[*abc.Object]bool, memo map[*abc.Object]bool) bool {
	if value, ok := memo[object]; ok {
		return value
	}
	var value bool
	switch object.Kind {
	case abc.ObjectPI:
		value = assignment[object]
	case abc.ObjectConst1:
		value = true
	case abc.ObjectNode:
		fanins := object.Fanins()
		if len(fanins) != 2 {
			panic("expected AIG node")
		}
		left := evalIvyRoundtrip(network, fanins[0].Object, assignment, memo)
		right := evalIvyRoundtrip(network, fanins[1].Object, assignment, memo)
		if fanins[0].Complement {
			left = !left
		}
		if fanins[1].Complement {
			right = !right
		}
		value = left && right
	default:
		panic("unexpected AIG driver kind")
	}
	memo[object] = value
	return value
}

func evalIvyRoundtripOutputs(network *abc.Network, assignment map[*abc.Object]bool) []bool {
	outputs := make([]bool, abc.Abc_NtkCoNum(network))
	for i := range outputs {
		co := abc.Abc_NtkCo(network, int32(i))
		fanin := co.Fanins()[0]
		outputs[i] = evalIvyRoundtrip(network, fanin.Object, assignment, map[*abc.Object]bool{})
		if fanin.Complement {
			outputs[i] = !outputs[i]
		}
	}
	return outputs
}

func evalIvySignal(manager *ivy.Ivy_Man_t, signal *ivy.Ivy_Obj_t, inputs map[*ivy.Ivy_Obj_t]bool) bool {
	complement := ivy.Ivy_IsComplement(signal) != 0
	object := ivy.Ivy_Regular(signal)
	var value bool
	switch {
	case ivy.Ivy_ObjIsConst1(object) != 0:
		value = true
	case ivy.Ivy_ObjIsPi(object) != 0:
		value = inputs[object]
	case ivy.Ivy_ObjIsNode(object) != 0:
		left := evalIvySignal(manager, ivy.Ivy_ObjChild0(object), inputs)
		right := evalIvySignal(manager, ivy.Ivy_ObjChild1(object), inputs)
		if ivy.Ivy_ObjIsExor(object) != 0 {
			value = left != right
		} else {
			value = left && right
		}
	default:
		panic("unexpected Ivy signal")
	}
	return value != complement
}

func TestIvyConversionRoundTripPreservesCIOrderAndComplement(t *testing.T) {
	source := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a := abc.Abc_NtkCreatePi(source)
	b := abc.Abc_NtkCreatePi(source)
	a.SetName("a")
	b.SetName("b")
	positive := abc.Abc_NtkCreatePo(source)
	negative := abc.Abc_NtkCreatePo(source)
	positive.SetName("positive")
	negative.SetName("negative")
	aigManager := source.FunctionManager.(*abc.Abc_Aig_t)
	root := abc.Abc_AigAnd(aigManager, a.Signal(), b.Signal().Not())
	abc.Abc_ObjAddFanin(positive, root)
	abc.Abc_ObjAddFanin(negative, root.Not())

	manager := Abc_NtkToIvy(source)
	defer ivy.Ivy_ManStop(manager)
	if ivy.Ivy_ManPiNum(manager) != 2 || ivy.Ivy_ManPoNum(manager) != 2 {
		t.Fatalf("Ivy interface = %d PIs, %d POs; want 2 and 2", ivy.Ivy_ManPiNum(manager), ivy.Ivy_ManPoNum(manager))
	}
	if got := []int32{ivy.Ivy_ObjFaninC0(ivy.Ivy_ManPo(manager, 0)), ivy.Ivy_ObjFaninC0(ivy.Ivy_ManPo(manager, 1))}; got[0] != 0 || got[1] != 1 {
		t.Fatalf("Ivy output complement flags = %v, want [0 1]", got)
	}

	mapped := Abc_NtkFromIvy(source, manager)
	defer abc.Abc_NtkDelete(mapped)
	for i, want := range []string{"a", "b"} {
		if got := abc.Abc_NtkCi(mapped, int32(i)).Name(); got != want {
			t.Fatalf("CI %d name = %q, want %q", i, got, want)
		}
	}
	for i, want := range []string{"positive", "negative"} {
		if got := abc.Abc_NtkCo(mapped, int32(i)).Name(); got != want {
			t.Fatalf("CO %d name = %q, want %q", i, got, want)
		}
	}
	inputs := []*abc.Object{abc.Abc_NtkCi(mapped, 0), abc.Abc_NtkCi(mapped, 1)}
	for bits := 0; bits < 4; bits++ {
		assignment := map[*abc.Object]bool{inputs[0]: bits&1 != 0, inputs[1]: bits&2 != 0}
		got := evalIvyRoundtripOutputs(mapped, assignment)
		want := (bits&1 != 0) && (bits&2 == 0)
		if got[0] != want || got[1] != !want {
			t.Fatalf("assignment %02b gives %v, want [%v %v]", bits, got, want, !want)
		}
	}
}

func TestIvyStrashSOPPreservesXORFunction(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(network), abc.Abc_NtkCreatePi(network)
	node := abc.Abc_NtkCreateNode(network)
	node.Data = abc.Abc_SopCreateXor(nil, 2)
	abc.Abc_ObjAddFanin(node, a)
	abc.Abc_ObjAddFanin(node, b)
	output := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(output, node)

	manager := Abc_NtkToIvy(network)
	defer ivy.Ivy_ManStop(manager)
	root := ivy.Ivy_ObjChild0(ivy.Ivy_ManPo(manager, 0))
	for bits := 0; bits < 4; bits++ {
		assignment := map[*ivy.Ivy_Obj_t]bool{
			ivy.Ivy_ManPi(manager, 0): bits&1 != 0,
			ivy.Ivy_ManPi(manager, 1): bits&2 != 0,
		}
		got := evalIvySignal(manager, root, assignment)
		want := (bits&1 != 0) != (bits&2 != 0)
		if got != want {
			t.Fatalf("input %02b evaluates to %v, want XOR %v; SOP %q", bits, got, want, node.Data)
		}
	}
}

func TestIvyLatchValueCollectionFollowsLatchOrderAndDontCare(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_NETLIST, abc.ABC_FUNC_SOP, 1)
	first := abc.Abc_NtkCreateLatch(network)
	second := abc.Abc_NtkCreateLatch(network)
	third := abc.Abc_NtkCreateLatch(network)
	abc.Abc_LatchSetInit0(first)
	abc.Abc_LatchSetInit1(second)
	abc.Abc_LatchSetInitDc(third)
	values := Abc_NtkCollectLatchValuesIvy(network, 0)
	if got := []int32{vec.Vec_IntEntry(values, 0), vec.Vec_IntEntry(values, 1), vec.Vec_IntEntry(values, 2)}; got[0] != int32(ivy.IVY_INIT_0) || got[1] != int32(ivy.IVY_INIT_1) || got[2] != int32(ivy.IVY_INIT_DC) {
		t.Fatalf("latch initial values = %v", got)
	}
	forced := Abc_NtkCollectLatchValuesIvy(network, 1)
	for i := 0; i < vec.Vec_IntSize(forced); i++ {
		if got := vec.Vec_IntEntry(forced, i); got != int32(ivy.IVY_INIT_DC) {
			t.Fatalf("forced don't-care latch %d = %d", i, got)
		}
	}
}
