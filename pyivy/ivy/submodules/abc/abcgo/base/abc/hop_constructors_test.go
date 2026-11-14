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
