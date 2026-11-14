package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func multiTestAig(t *testing.T) (*abc.Network, []*abc.Object) {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
	inputs := []*abc.Object{abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)}
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	ab := abc.Abc_AigAnd(manager, inputs[0], inputs[1])
	cd := abc.Abc_AigAnd(manager, inputs[2], inputs[3])
	root := abc.Abc_AigAnd(manager, ab, cd)
	shared := abc.Abc_AigAnd(manager, ab, inputs[2])
	for i, driver := range []abc.Edge{root, shared} {
		po := abc.Abc_NtkCreatePo(n)
		po.SetName([]string{"wide", "shared"}[i])
		abc.Abc_ObjAddFanin(po, driver)
	}
	return n, inputs
}

func evalMultiNode(t *testing.T, n *abc.Network, node *abc.Object, input map[*abc.Object]bool, memo map[*abc.Object]bool) bool {
	t.Helper()
	if value, ok := memo[node]; ok {
		return value
	}
	if node.Kind == abc.ObjectPI {
		return input[node]
	}
	if node.Kind == abc.ObjectConst1 {
		return true
	}
	if node.Kind != abc.ObjectNode {
		t.Fatalf("unexpected node kind %v", node.Kind)
	}
	values := make([]bool, len(node.Fanins()))
	for i, edge := range node.Fanins() {
		values[i] = evalMultiNode(t, n, edge.Object, input, memo) != edge.Complement
	}
	var result bool
	if bdd, ok := node.Data.(*core.DdNode); ok {
		manager := n.FunctionManager.(*core.DdManager)
		var err error
		result, err = manager.Evaluate(bdd, values)
		if err != nil {
			t.Fatalf("evaluate node %d: %v", node.ID, err)
		}
	} else {
		if len(values) != 2 {
			t.Fatalf("AIG node %d has %d fanins", node.ID, len(values))
		}
		result = values[0] && values[1]
	}
	memo[node] = result
	return result
}

func TestAbcNtkMultiPreservesMultiOutputFunctions(t *testing.T) {
	for _, mode := range []struct {
		name  string
		multi int32
	}{{"fanin_limit", 0}, {"multi_boundary", 1}} {
		t.Run(mode.name, func(t *testing.T) {
			n, inputs := multiTestAig(t)
			converted := Abc_NtkMulti(n, 0, 3, 0, mode.multi, 0, 0)
			if converted == nil {
				t.Fatal("Abc_NtkMulti returned nil")
			}
			if got := abc.Abc_NtkPoNum(converted); got != 2 {
				t.Fatalf("output count = %d, want 2", got)
			}
			pis := converted.ObjectsOfKind(abc.ObjectPI)
			pos := converted.ObjectsOfKind(abc.ObjectPO)
			for pattern := 0; pattern < 1<<len(inputs); pattern++ {
				oldValues := make(map[*abc.Object]bool, len(inputs))
				newValues := make(map[*abc.Object]bool, len(inputs))
				for i, pi := range inputs {
					oldValues[pi] = pattern&(1<<i) != 0
					newValues[pis[i]] = oldValues[pi]
				}
				for i, po := range n.ObjectsOfKind(abc.ObjectPO) {
					want := evalMultiNode(t, n, abc.Abc_ObjFanin0(po), oldValues, map[*abc.Object]bool{}) != (abc.Abc_ObjFaninC0(po) != 0)
					got := evalMultiNode(t, converted, abc.Abc_ObjFanin0(pos[i]), newValues, map[*abc.Object]bool{}) != (abc.Abc_ObjFaninC0(pos[i]) != 0)
					if got != want {
						t.Fatalf("pattern %04b output %d: got %v, want %v", pattern, i, got, want)
					}
				}
			}
		})
	}
}
