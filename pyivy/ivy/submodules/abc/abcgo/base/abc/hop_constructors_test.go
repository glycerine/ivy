package abc

import (
	"github.com/glycerine/abcgo/aig/hop"
	"testing"
)

func TestHopConstantAndBufferConstructors(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_AIG, 1)
	defer Abc_NtkDelete(n)
	input := Abc_NtkCreatePi(n)
	manager := n.FunctionManager.(*hop.Hop_Man_t)
	for _, tc := range []struct {
		node *Object
		want *hop.Hop_Obj_t
	}{
		{Abc_NtkCreateNodeConst0(n), hop.Hop_Not(hop.Hop_ManConst1(manager))},
		{Abc_NtkCreateNodeConst1(n), hop.Hop_ManConst1(manager)},
		{Abc_NtkCreateNodeBuf(n, input), hop.Hop_IthVar(manager, 0)},
	} {
		if got, ok := tc.node.Data.(*hop.Hop_Obj_t); !ok || got != tc.want {
			t.Fatalf("constructor produced %T %v, want Hop function %v", tc.node.Data, tc.node.Data, tc.want)
		}
	}
}

func TestComplementRetainsHopAndSopRepresentations(t *testing.T) {
	for _, function := range []Abc_NtkFunc_t{ABC_FUNC_AIG, ABC_FUNC_SOP} {
		n := Abc_NtkAlloc(ABC_NTK_LOGIC, function, 1)
		a, b := Abc_NtkCreatePi(n), Abc_NtkCreatePi(n)
		node := Abc_NtkCreateNodeAnd(n, []*Object{a, b})
		Abc_NodeComplementInput(node, a)
		Abc_NodeComplement(node)
		switch function {
		case ABC_FUNC_AIG:
			if _, ok := node.Data.(*hop.Hop_Obj_t); !ok {
				t.Fatalf("lost Hop function: %T", node.Data)
			}
		case ABC_FUNC_SOP:
			if _, ok := node.Data.(string); !ok {
				t.Fatalf("lost SOP function: %T", node.Data)
			}
		}
		for bits := 0; bits < 4; bits++ {
			av, bv := bits&1 != 0, bits&2 != 0
			got, err := n.Evaluate(node.Signal(), map[*Object]bool{a: av, b: bv})
			if err != nil || got != !(!av && bv) {
				t.Fatalf("function=%d bits=%d got=%v err=%v", function, bits, got, err)
			}
		}
		Abc_NtkDelete(n)
	}
}
