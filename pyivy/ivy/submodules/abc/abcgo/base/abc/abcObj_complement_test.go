package abc

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestNodeComplementPreservesBDDRepresentation(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_BDD, 1)
	input := Abc_NtkCreatePi(n)
	node := Abc_NtkCreateNode(n)
	Abc_ObjAddFanin(node, input)
	manager := n.FunctionManager.(*core.DdManager)
	root := core.Cudd_bddIthVar(manager, 0)
	core.Cudd_Ref(root)
	node.Data = root

	Abc_NodeComplement(node)
	complement, ok := node.Data.(*core.DdNode)
	if !ok {
		t.Fatalf("BDD complement changed node data type to %T", node.Data)
	}
	for _, value := range []bool{false, true} {
		got, err := manager.Evaluate(complement, []bool{value})
		if err != nil || got == value {
			t.Fatalf("complement(%t) = %t, err %v", value, got, err)
		}
	}
}

func TestNodeComplementInputPreservesBDDRepresentation(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_BDD, 1)
	a, b := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
	node := Abc_NtkCreateNode(n)
	Abc_ObjAddFanin(node, a)
	Abc_ObjAddFanin(node, b)
	manager := n.FunctionManager.(*core.DdManager)
	root := core.Cudd_bddAnd(manager, core.Cudd_bddIthVar(manager, 0), core.Cudd_bddIthVar(manager, 1))
	core.Cudd_Ref(root)
	node.Data = root

	Abc_NodeComplementInput(node, a)
	complement, ok := node.Data.(*core.DdNode)
	if !ok {
		t.Fatalf("BDD input complement changed node data type to %T", node.Data)
	}
	for assignment := 0; assignment < 4; assignment++ {
		values := []bool{assignment&1 != 0, assignment&2 != 0}
		got, err := manager.Evaluate(complement, values)
		want := !values[0] && values[1]
		if err != nil || got != want {
			t.Fatalf("input-complement assignment %v = %t, want %t (err %v)", values, got, want, err)
		}
	}
}

func TestNodeComplementRetainsFunctionDataBehavior(t *testing.T) {
	n := NewNetwork("function")
	input := Abc_NtkCreatePi(n)
	node := Abc_NtkCreateNode(n)
	Abc_ObjAddFanin(node, input)
	node.Data = func(values []bool) bool { return values[0] }

	Abc_NodeComplement(node)
	if _, ok := node.Data.(func([]bool) bool); !ok {
		t.Fatalf("function-valued node data changed to %T", node.Data)
	}
	for _, value := range []bool{false, true} {
		got, err := n.Evaluate(node.Signal(), map[*Object]bool{input: value})
		if err != nil || got == value {
			t.Fatalf("complement(%t) = %t, err %v", value, got, err)
		}
	}
}

func TestBDDConstantAndBufferConstructorsKeepBDDData(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_BDD, 1)
	manager := n.FunctionManager.(*core.DdManager)
	input := Abc_NtkCreatePi(n)
	for _, tc := range []struct {
		name string
		node *Object
		want func([]bool) bool
	}{
		{"const0", Abc_NtkCreateNodeConst0(n), func([]bool) bool { return false }},
		{"const1", Abc_NtkCreateNodeConst1(n), func([]bool) bool { return true }},
		{"buffer", Abc_NtkCreateNodeBuf(n, input), func(v []bool) bool { return v[0] }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bdd, ok := tc.node.Data.(*core.DdNode)
			if !ok {
				t.Fatalf("constructor data = %T, want BDD node", tc.node.Data)
			}
			for _, value := range []bool{false, true} {
				got, err := manager.Evaluate(bdd, []bool{value})
				if err != nil || got != tc.want([]bool{value}) {
					t.Fatalf("input %t: got %t, err %v", value, got, err)
				}
			}
		})
	}
}
