package abc

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestDerivePropertyMonitor(t *testing.T) {
	for _, function := range []Abc_NtkFunc_t{ABC_FUNC_SOP, ABC_FUNC_BDD, ABC_FUNC_AIG} {
		src := Abc_NtkAlloc(ABC_NTK_LOGIC, function, 1)
		input := src.NewObject(ObjectPI)
		input.SetName("in")
		state := Abc_NtkAddLatch(src, input, ABC_INIT_ZERO)
		ids, values := vec.Vec_IntAlloc(2), vec.Vec_IntAlloc(2)
		vec.Vec_IntPush(ids, int32(input.ID))
		vec.Vec_IntPush(ids, int32(state.ID))
		vec.Vec_IntPush(values, 1)
		vec.Vec_IntPush(values, 0)
		result := Abc_NtkDeriveWithOnePo(src, ids, values)
		if len(result.ObjectsOfKind(ObjectLatch)) != 1 || len(result.ObjectsOfKind(ObjectPO)) != 1 {
			t.Fatal("monitor interface mismatch")
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
		ci := result.cis
		for _, a := range []bool{false, true} {
			for _, q := range []bool{false, true} {
				got := eval(result.ObjectsOfKind(ObjectPO)[0], map[*Object]bool{ci[0]: a, ci[1]: q})
				if got != (a && !q) {
					t.Fatalf("function=%d input=%v state=%v got=%v", function, a, q, got)
				}
			}
		}
		Abc_NtkDelete(result)
		Abc_NtkDelete(src)
	}
}
