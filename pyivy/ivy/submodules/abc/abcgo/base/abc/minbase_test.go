package abc

import (
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestMinimumBasePreservesBDDAndFaninOrder(t *testing.T) {
	n := NewNetwork("support")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_BDD)
	dd := core.NewBDDManager(3)
	n.FunctionManager = dd
	inputs := []*Object{n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)}
	node := n.NewObject(ObjectNode)
	for _, input := range inputs {
		Abc_ObjAddFanin(node, input)
	}
	old := core.Cudd_bddAnd(dd, dd.Vars[0], core.Cudd_Not(dd.Vars[2]))
	core.Cudd_Ref(old)
	node.Data = old
	if Abc_NodeMinimumBase(node) != 1 || len(node.fanins) != 2 || node.fanins[0].Object != inputs[0] || node.fanins[1].Object != inputs[2] {
		t.Fatal("fanin support not compacted")
	}
	for a := 0; a < 2; a++ {
		for b := 0; b < 2; b++ {
			got, err := dd.Evaluate(node.Data.(*core.DdNode), []bool{a != 0, b != 0, false})
			if err != nil || got != (a != 0 && b == 0) {
				t.Fatalf("a=%d b=%d got=%v err=%v", a, b, got, err)
			}
		}
	}
	if Abc_NodeMinimumBase(node) != 0 {
		t.Fatal("already minimal node changed")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestRemoveDuplicateComplementedBDDInputs(t *testing.T) {
	n := NewNetwork("duplicates")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_BDD)
	dd := core.NewBDDManager(2)
	n.FunctionManager = dd
	input := n.NewObject(ObjectPI)
	node := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(node, input)
	Abc_ObjAddFanin(node, input.Signal().Not())
	old := core.Cudd_bddAnd(dd, dd.Vars[0], core.Cudd_Not(dd.Vars[1]))
	core.Cudd_Ref(old)
	node.Data = old
	if Abc_NodeRemoveDupFanins(node) != 1 || len(node.fanins) != 1 {
		t.Fatal("duplicate was not removed")
	}
	for _, value := range []bool{false, true} {
		got, err := dd.Evaluate(node.Data.(*core.DdNode), []bool{value, false})
		if err != nil || got != value {
			t.Fatalf("input=%v got=%v err=%v", value, got, err)
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestCollapseBDDNodesPreservesComplementedEdges(t *testing.T) {
	n := NewNetwork("collapse")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_BDD)
	dd := core.NewBDDManager(3)
	n.FunctionManager = dd
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	first := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(first, a.Signal().Not())
	Abc_ObjAddFanin(first, b)
	first.Data = core.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
	core.Cudd_Ref(first.Data.(*core.DdNode))
	second := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(second, first.Signal().Not())
	Abc_ObjAddFanin(second, c)
	second.Data = core.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	core.Cudd_Ref(second.Data.(*core.DdNode))
	firstFunction, secondFunction := first.Data.(*core.DdNode), second.Data.(*core.DdNode)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, second)
	values := make([]bool, 8)
	for mask := 0; mask < 8; mask++ {
		var err error
		values[mask], err = n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
	}
	fanins := vec.Vec_PtrAlloc(3)
	if got := Abc_NodeCollapse(first, second, fanins, make([]int32, 3), make([]int32, 3)); got != 1 {
		t.Fatal("collapse rejected")
	}
	if core.Cudd_Regular(firstFunction).Ref != 0 || core.Cudd_Regular(secondFunction).Ref != 0 {
		t.Fatal("deleted BDD nodes retained their function references")
	}
	for mask := 0; mask < 8; mask++ {
		got, err := n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
		if err != nil || got != values[mask] {
			t.Fatalf("mask=%d got=%v want=%v err=%v", mask, got, values[mask], err)
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestCollapseHopNodesPreservesComplementedEdges(t *testing.T) {
	n := NewNetwork("hop collapse")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_AIG)
	manager := hop.Hop_ManStart()
	n.FunctionManager = manager
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	first := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(first, a.Signal().Not())
	Abc_ObjAddFanin(first, b)
	first.Data = hop.Hop_Exor(manager, hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1))
	second := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(second, first.Signal().Not())
	Abc_ObjAddFanin(second, c)
	second.Data = hop.Hop_And(manager, hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1))
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, second)
	values := make([]bool, 8)
	for mask := 0; mask < 8; mask++ {
		var err error
		values[mask], err = n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := Abc_NodeCollapse1(first, second, vec.Vec_PtrAlloc(3), make([]int32, 3), make([]int32, 3)); got != 1 {
		t.Fatal("collapse rejected")
	}
	for mask := 0; mask < 8; mask++ {
		got, err := n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
		if err != nil || got != values[mask] {
			t.Fatalf("mask=%d got=%v want=%v err=%v", mask, got, values[mask], err)
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestHopEliminationRespectsSizeLimitAndTruth(t *testing.T) {
	for _, limit := range []int32{2, 3} {
		n := NewNetwork("elimination")
		n.Type = uint8(ABC_NTK_LOGIC)
		n.Function = uint8(ABC_FUNC_AIG)
		manager := hop.Hop_ManStart()
		n.FunctionManager = manager
		a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
		first := n.NewObject(ObjectNode)
		Abc_ObjAddFanin(first, a.Signal().Not())
		Abc_ObjAddFanin(first, b)
		first.Data = hop.Hop_And(manager, hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1))
		second := n.NewObject(ObjectNode)
		Abc_ObjAddFanin(second, first.Signal().Not())
		Abc_ObjAddFanin(second, c)
		second.Data = hop.Hop_Or(manager, hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1))
		po := n.NewObject(ObjectPO)
		Abc_ObjAddFanin(po, second)
		Abc_NtkOrderCisCos(n)
		expected := make([]bool, 8)
		for mask := 0; mask < 8; mask++ {
			var err error
			expected[mask], err = n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
			if err != nil {
				t.Fatal(err)
			}
		}
		if Abc_NtkEliminate1(n, 0, limit, 3, 0, 0) != 1 {
			t.Fatal("elimination failed")
		}
		wantNodes := int32(1)
		if limit == 2 {
			wantNodes = 2
		}
		if got := Abc_NtkNodeNum(n); got != wantNodes {
			t.Fatalf("limit=%d nodes=%d want=%d", limit, got, wantNodes)
		}
		for mask := 0; mask < 8; mask++ {
			got, err := n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
			if err != nil || got != expected[mask] {
				t.Fatalf("limit=%d mask=%d got=%v want=%v err=%v", limit, mask, got, expected[mask], err)
			}
		}
		if err := n.Check(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBDDEliminationSizeAndReordering(t *testing.T) {
	for _, limit := range []int32{2, 3} {
		n := NewNetwork("BDD elimination")
		n.Type = uint8(ABC_NTK_LOGIC)
		n.Function = uint8(ABC_FUNC_BDD)
		dd := core.NewBDDManager(3)
		n.FunctionManager = dd
		a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
		first := n.NewObject(ObjectNode)
		Abc_ObjAddFanin(first, a.Signal().Not())
		Abc_ObjAddFanin(first, b)
		first.Data = core.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[1])
		core.Cudd_Ref(first.Data.(*core.DdNode))
		second := n.NewObject(ObjectNode)
		Abc_ObjAddFanin(second, first.Signal().Not())
		Abc_ObjAddFanin(second, c)
		second.Data = core.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
		core.Cudd_Ref(second.Data.(*core.DdNode))
		po := n.NewObject(ObjectPO)
		Abc_ObjAddFanin(po, second)
		Abc_NtkOrderCisCos(n)
		values := make([]bool, 8)
		for mask := 0; mask < 8; mask++ {
			var err error
			values[mask], err = n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
			if err != nil {
				t.Fatal(err)
			}
		}
		if Abc_NtkEliminate(n, limit, 0, 0) != 1 {
			t.Fatal("BDD elimination failed")
		}
		want := int32(1)
		if limit == 2 {
			want = 2
		}
		if got := Abc_NtkNodeNum(n); got != want {
			t.Fatalf("limit=%d got %d nodes want %d", limit, got, want)
		}
		for mask := 0; mask < 8; mask++ {
			got, err := n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
			if err != nil || got != values[mask] {
				t.Fatalf("limit=%d mask=%d got=%v want=%v err=%v", limit, mask, got, values[mask], err)
			}
		}
		if err := n.Check(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBDDSupportDoesNotMutateSharedNodes(t *testing.T) {
	dd := core.NewBDDManager(4)
	root := core.Cudd_bddXor(dd, dd.Vars[0], dd.Vars[3])
	support := vec.Vec_StrAlloc(0)
	for _, edge := range []*core.DdNode{root, core.Cudd_Not(root), root} {
		if count := Abc_NodeSupport(edge, support, 4); count != 2 {
			t.Fatalf("support count %d", count)
		}
		for i, want := range []byte{1, 0, 0, 1} {
			if got := vec.Vec_StrEntry(support, i); got != want {
				t.Fatalf("support[%d]=%d", i, got)
			}
		}
	}
}

func TestSpecialEliminationPreservesFunction(t *testing.T) {
	n := NewNetwork("special")
	n.Type = uint8(ABC_NTK_LOGIC)
	n.Function = uint8(ABC_FUNC_BDD)
	dd := core.NewBDDManager(3)
	n.FunctionManager = dd
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	first := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(first, a)
	Abc_ObjAddFanin(first, b)
	first.Data = core.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	core.Cudd_Ref(first.Data.(*core.DdNode))
	second := n.NewObject(ObjectNode)
	Abc_ObjAddFanin(second, first)
	Abc_ObjAddFanin(second, c)
	second.Data = core.Cudd_bddAnd(dd, dd.Vars[0], dd.Vars[1])
	core.Cudd_Ref(second.Data.(*core.DdNode))
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, second)
	Abc_NtkOrderCisCos(n)
	values := make([]bool, 8)
	for mask := 0; mask < 8; mask++ {
		var err error
		values[mask], err = n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
		if err != nil {
			t.Fatal(err)
		}
	}
	if Abc_NtkEliminateSpecial(n, 3, 0) != 1 {
		t.Fatal("special elimination failed")
	}
	if got := Abc_NtkNodeNum(n); got != 1 {
		t.Fatalf("got %d nodes", got)
	}
	for mask := 0; mask < 8; mask++ {
		got, err := n.Evaluate(po.Signal(), map[*Object]bool{a: mask&1 != 0, b: mask&2 != 0, c: mask&4 != 0})
		if err != nil || got != values[mask] {
			t.Fatalf("mask=%d got=%v want=%v err=%v", mask, got, values[mask], err)
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
