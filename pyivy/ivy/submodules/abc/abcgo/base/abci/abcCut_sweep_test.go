package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/opt/cut"
)

func TestCutWrapperRestoresAdjustedMuxReferencesAndTruth(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	control := abc.Abc_NtkCreatePi(network)
	thenInput := abc.Abc_NtkCreatePi(network)
	elseInput := abc.Abc_NtkCreatePi(network)
	manager := network.FunctionManager.(*abc.Abc_Aig_t)
	mux := abc.Abc_AigMux(manager, control, thenInput, elseInput)
	po := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(po, mux)

	controlFanouts := abc.Abc_ObjFanoutNum(control)
	adjacency := control.Fanouts()
	Abc_NtkCutsSubtractFanunt(network)
	if got := abc.Abc_ObjFanoutNum(control); got != controlFanouts-1 {
		t.Fatalf("adjusted control fanout count=%d, want %d", got, controlFanouts-1)
	}
	if got := control.Fanouts(); len(got) != len(adjacency) || got[0] != adjacency[0] || got[1] != adjacency[1] {
		t.Fatal("temporary cut adjustment changed graph adjacency")
	}
	Abc_NtkCutsAddFanunt(network)
	if got := abc.Abc_ObjFanoutNum(control); got != controlFanouts {
		t.Fatalf("restored control fanout count=%d, want %d", got, controlFanouts)
	}

	params := cut.DefaultParams(3)
	params.FTruth = 1
	params.FAdjust = 1
	managerCuts := Abc_NtkCuts(network, params)
	if managerCuts == nil {
		t.Fatal("cut manager is nil")
	}
	if params.NIdsMax != abc.Abc_NtkObjNumMax(network) {
		t.Fatalf("NIdsMax=%d, want object capacity %d", params.NIdsMax, abc.Abc_NtkObjNumMax(network))
	}
	if got := abc.Abc_ObjFanoutNum(control); got != controlFanouts {
		t.Fatalf("Abc_NtkCuts leaked temporary fanout adjustment: %d != %d", got, controlFanouts)
	}

	var threeLeaf *cut.Cut_Cut_t
	for current := Abc_NodeReadCuts(managerCuts, mux.Object); current != nil; current = current.Next {
		if cut.Cut_CutReadLeaveNum(current) == 3 {
			threeLeaf = current
			break
		}
	}
	if threeLeaf == nil {
		t.Fatal("mux cut with all three leaves was not computed")
	}
	leaves, truth := cut.Cut_CutReadLeaves(threeLeaf), cut.Cut_CutReadTruth(threeLeaf)
	if len(truth) == 0 {
		t.Fatal("three-leaf mux cut has no truth table")
	}
	for assignment := 0; assignment < 8; assignment++ {
		values := map[int32]bool{}
		for pin, id := range leaves {
			values[id] = assignment&(1<<uint(pin)) != 0
		}
		want := evaluateCutBoundary(mux.Object, values, map[*abc.Object]bool{})
		index := assignment
		got := truth[index>>5]&(1<<uint(index&31)) != 0
		if threeLeaf.FCompl {
			got = !got
		}
		if got != want {
			t.Fatalf("cut truth at assignment %03b = %v, want %v (leaves=%v FCompl=%v truth=%08x)", assignment, got, want, leaves, threeLeaf.FCompl, truth[0])
		}
	}
}

func evaluateCutBoundary(node *abc.Object, boundary map[int32]bool, memo map[*abc.Object]bool) bool {
	if value, ok := boundary[int32(node.ID)]; ok {
		return value
	}
	if value, ok := memo[node]; ok {
		return value
	}
	if node.Kind == abc.ObjectConst1 {
		return true
	}
	fanins := node.Fanins()
	if len(fanins) != 2 {
		panic("cut truth boundary reaches a non-leaf that is not a two-input AIG node")
	}
	a, b := evaluateCutBoundary(fanins[0].Object, boundary, memo), evaluateCutBoundary(fanins[1].Object, boundary, memo)
	if fanins[0].Complement {
		a = !a
	}
	if fanins[1].Complement {
		b = !b
	}
	value := a && b
	memo[node] = value
	return value
}

func TestSweepBDDNodeReferencesRemainBalanced(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(network), abc.Abc_NtkCreatePi(network)
	inverter := abc.Abc_NtkCreateNodeInv(network, a)
	root := abc.Abc_NtkCreateNodeAnd(network, []*abc.Object{inverter, b})
	po := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(po, root)

	if removed := abc.Abc_NtkSweep(network, 0); removed != 1 {
		t.Fatalf("sweep removed %d nodes, want 1", removed)
	}
	if got := abc.Abc_ObjFaninNum(abc.Abc_ObjFanin0(po)); got != 2 {
		t.Fatalf("swept root fanin count=%d, want 2", got)
	}
	dd, ok := network.FunctionManager.(*core.DdManager)
	if !ok {
		t.Fatal("BDD sweep did not retain its CUDD manager")
	}
	function, ok := root.Data.(*core.DdNode)
	if !ok || core.Cudd_ExternalRefCount(dd, function) == 0 {
		t.Fatal("live network node lost its external BDD reference")
	}
	for bits := 0; bits < 4; bits++ {
		inputs := map[*abc.Object]bool{a: bits&1 != 0, b: bits&2 != 0}
		got, err := network.Evaluate(abc.Abc_ObjFanin0(po).Signal(), inputs)
		if err != nil {
			t.Fatal(err)
		}
		if want := !inputs[a] && inputs[b]; got != want {
			t.Fatalf("a=%v b=%v got=%v want=%v", inputs[a], inputs[b], got, want)
		}
	}
}

func TestSweepBufferAndInverterPreserveAIGTruth(t *testing.T) {
	network := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	a, b := abc.Abc_NtkCreatePi(network), abc.Abc_NtkCreatePi(network)
	inverter := abc.Abc_NtkCreateNodeInv(network, a)
	root := abc.Abc_NtkCreateNodeAnd(network, []*abc.Object{inverter, b})
	po := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(po, root)

	if removed := Abc_NtkSweepBufsInvs(network, 0); removed != 1 {
		t.Fatalf("buffer/inverter sweep removed %d nodes, want 1", removed)
	}
	if root.Fanins()[0].Object != a {
		t.Fatal("inverter was not removed from the root fanin")
	}
	for bits := 0; bits < 4; bits++ {
		inputs := map[*abc.Object]bool{a: bits&1 != 0, b: bits&2 != 0}
		got, err := network.Evaluate(abc.Abc_ObjFanin0(po).Signal(), inputs)
		if err != nil {
			t.Fatal(err)
		}
		if want := !inputs[a] && inputs[b]; got != want {
			t.Fatalf("a=%v b=%v got=%v want=%v", inputs[a], inputs[b], got, want)
		}
	}
}
