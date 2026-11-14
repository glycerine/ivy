package abci

import (
	"fmt"
	"math"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
)

// Abc_NtkTransitionRelation builds the monolithic transition relation over
// present and next state, then existentially quantifies primary inputs.
func Abc_NtkTransitionRelation(dd *core.DdManager, network *abc.Network, verbose int32) *core.DdNode {
	if dd == nil || network == nil {
		return nil
	}
	piNum, ciNum := len(network.ObjectsOfKind(abc.ObjectPI)), len(network.Inputs())
	latches := network.ObjectsOfKind(abc.ObjectLatch)
	for i := 0; i < ciNum+len(latches); i++ {
		core.Cudd_bddIthVar(dd, i)
	}
	core.Cudd_AutodynEnable(dd, core.CUDD_REORDER_SYMM_SIFT)
	relation := core.Cudd_ReadOne(dd)
	core.Cudd_Ref(relation)
	for i, latch := range latches {
		if len(latch.Fanins()) != 1 {
			core.Cudd_RecursiveDeref(dd, relation)
			Abc_NtkFreeGlobalBdds(network, 0)
			return nil
		}
		bi := latch.Fanins()[0].Object
		function := Abc_ObjGlobalBdd(bi)
		if function == nil {
			core.Cudd_RecursiveDeref(dd, relation)
			Abc_NtkFreeGlobalBdds(network, 0)
			return nil
		}
		product := core.Cudd_bddXnor(dd, core.Cudd_bddIthVar(dd, ciNum+i), function)
		core.Cudd_Ref(product)
		next := core.Cudd_bddAnd(dd, relation, product)
		core.Cudd_Ref(next)
		core.Cudd_RecursiveDeref(dd, relation)
		core.Cudd_RecursiveDeref(dd, product)
		relation = next
	}
	Abc_NtkFreeGlobalBdds(network, 0)
	inputs := extrab.Extra_bddComputeRangeCube(dd, 0, piNum)
	if inputs == nil {
		core.Cudd_RecursiveDeref(dd, relation)
		return nil
	}
	core.Cudd_Ref(inputs)
	result := core.Cudd_bddExistAbstract(dd, relation, inputs)
	if result == nil {
		core.Cudd_RecursiveDeref(dd, relation)
		core.Cudd_RecursiveDeref(dd, inputs)
		return nil
	}
	core.Cudd_Ref(result)
	core.Cudd_RecursiveDeref(dd, relation)
	core.Cudd_RecursiveDeref(dd, inputs)
	relation = result
	if verbose != 0 {
		fmt.Printf("BDD nodes in the transition relation before reordering %d.\n", core.Cudd_DagSize(relation))
	}
	core.Cudd_ReduceHeap(dd, core.CUDD_REORDER_SYMM_SIFT, 100)
	core.Cudd_AutodynDisable(dd)
	if verbose != 0 {
		fmt.Printf("BDD nodes in the transition relation after reordering %d.\n", core.Cudd_DagSize(relation))
	}
	core.Cudd_Deref(relation)
	return relation
}

// Abc_NtkInitStateAndVarMap shares the source ordering and init semantics of
// Abc_NtkInitStateVarMap: don't-care initialization is treated as zero.
func Abc_NtkInitStateAndVarMap(dd *core.DdManager, network *abc.Network, verbose int32) *core.DdNode {
	return Abc_NtkInitStateVarMap(dd, network, verbose)
}

// Abc_NtkComputeUnreachable returns the complement of the least fixed-point
// reachable-state set, represented in current-state variables.
func Abc_NtkComputeUnreachable(dd *core.DdManager, network *abc.Network, transition, initial *core.DdNode, verbose int32) *core.DdNode {
	if dd == nil || network == nil || transition == nil || initial == nil {
		return nil
	}
	piNum, ciNum := len(network.ObjectsOfKind(abc.ObjectPI)), len(network.Inputs())
	latchNum := len(network.ObjectsOfKind(abc.ObjectLatch))
	current, reached, relation := initial, initial, transition
	core.Cudd_Ref(current)
	core.Cudd_Ref(reached)
	core.Cudd_Ref(relation)
	cube := extrab.Extra_bddComputeRangeCube(dd, piNum, ciNum)
	if cube == nil {
		core.Cudd_RecursiveDeref(dd, current)
		core.Cudd_RecursiveDeref(dd, reached)
		core.Cudd_RecursiveDeref(dd, relation)
		return nil
	}
	core.Cudd_Ref(cube)
	iterations := 0
	var next *core.DdNode
	for {
		iterations++
		next = core.Cudd_bddAndAbstract(dd, relation, current, cube)
		if next == nil {
			core.Cudd_RecursiveDeref(dd, current)
			core.Cudd_RecursiveDeref(dd, reached)
			core.Cudd_RecursiveDeref(dd, relation)
			core.Cudd_RecursiveDeref(dd, cube)
			return nil
		}
		core.Cudd_Ref(next)
		core.Cudd_RecursiveDeref(dd, current)
		mapped := abcReachVarMap(dd, piNum, ciNum, latchNum, next)
		if mapped == nil {
			core.Cudd_RecursiveDeref(dd, next)
			core.Cudd_RecursiveDeref(dd, reached)
			core.Cudd_RecursiveDeref(dd, relation)
			core.Cudd_RecursiveDeref(dd, cube)
			return nil
		}
		core.Cudd_Ref(mapped)
		core.Cudd_RecursiveDeref(dd, next)
		next = mapped
		if core.Cudd_bddLeq(dd, next, reached) {
			break
		}
		current = core.Cudd_bddAnd(dd, next, core.Cudd_Not(reached))
		core.Cudd_Ref(current)
		newReached := core.Cudd_bddOr(dd, reached, next)
		core.Cudd_Ref(newReached)
		core.Cudd_RecursiveDeref(dd, reached)
		reached = newReached
		core.Cudd_RecursiveDeref(dd, next)
	}
	core.Cudd_RecursiveDeref(dd, relation)
	core.Cudd_RecursiveDeref(dd, cube)
	core.Cudd_RecursiveDeref(dd, next)
	if verbose != 0 {
		minterms := int(core.Cudd_CountMinterm(dd, reached, latchNum))
		fmt.Printf("Reachability analysis completed in %d iterations.\n", iterations)
		fmt.Printf("The number of minterms in the reachable state set = %d. (%6.2f %%)\n", minterms, 100.0*float64(minterms)/math.Pow(2, float64(latchNum)))
	}
	core.Cudd_Deref(reached)
	return core.Cudd_Not(reached)
}

// Abc_NtkConstructExdc represents unreachable states as one BDD node driven
// by latch-output PIs, with POs for eligible outputs and latch inputs.
func Abc_NtkConstructExdc(dd *core.DdManager, network *abc.Network, unreachable *core.DdNode) *abc.Network {
	if dd == nil || network == nil || unreachable == nil {
		return nil
	}
	latches := network.ObjectsOfKind(abc.ObjectLatch)
	out := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	out.Name = "exdc"
	destDD := out.FunctionManager.(*core.DdManager)
	inputs := make([]*abc.Object, len(latches))
	for i, latch := range latches {
		output := latchPredecessorObject(latch)
		if output == nil {
			return nil
		}
		inputs[i] = out.NewObject(abc.ObjectPI)
		inputs[i].SetName(output.Name())
		core.Cudd_bddIthVar(destDD, i)
	}
	node := out.NewObject(abc.ObjectNode)
	for _, input := range inputs {
		if err := out.AddFanin(node, abc.Edge{Object: input}); err != nil {
			return nil
		}
	}
	permutation := make([]int, len(dd.Vars))
	for i := range permutation {
		permutation[i] = -1
	}
	piNum := len(network.ObjectsOfKind(abc.ObjectPI))
	for i := range latches {
		index := piNum + i
		if index >= len(permutation) {
			return nil
		}
		permutation[index] = i
	}
	node.Data = extrab.Extra_TransferPermute(dd, destDD, unreachable, permutation)
	if node.Data == nil {
		return nil
	}
	core.Cudd_Ref(node.Data.(*core.DdNode))
	abc.Abc_NodeMinimumBase(node)
	for _, po := range network.ObjectsOfKind(abc.ObjectPO) {
		if len(po.Fanins()) == 0 {
			return nil
		}
		driver := po.Fanins()[0].Object
		if driver.IsCI() {
			continue
		}
		newPO := out.NewObject(abc.ObjectPO)
		newPO.SetName(po.Name())
		if err := out.AddFanin(newPO, abc.Edge{Object: node}); err != nil {
			return nil
		}
	}
	for _, bi := range network.ObjectsOfKind(abc.ObjectBI) {
		newPO := out.NewObject(abc.ObjectPO)
		newPO.SetName(bi.Name())
		if err := out.AddFanin(newPO, abc.Edge{Object: node}); err != nil {
			return nil
		}
	}
	abc.Abc_NtkLogicMakeSimpleCos(out, 0)
	if abc.Abc_NtkBddToSop(out, -1, math.MaxInt32) == 0 {
		fmt.Println("Abc_NtkConstructExdc(): Converting to SOPs has failed.")
		return nil
	}
	return out
}

func latchPredecessorObject(latch *abc.Object) *abc.Object {
	if latch == nil || len(latch.Fanouts()) == 0 {
		return nil
	}
	for _, object := range latch.Network().ObjectsOfKind(abc.ObjectBO) {
		fans := object.Fanins()
		if len(fans) == 1 && fans[0].Object == latch {
			return object
		}
	}
	return nil
}

// Abc_NtkExtractSequentialDcs attaches the unreachable-state function as EXDC.
func Abc_NtkExtractSequentialDcs(network *abc.Network, verbose int32) int32 {
	if network == nil || abc.Abc_NtkIsStrash(network) == 0 {
		return 0
	}
	if network.Exdc != nil {
		abc.Abc_NtkDelete(network.Exdc)
		network.Exdc = nil
	}
	dd := Abc_NtkBuildGlobalBdds(network, 10000000, 1, 1, 0, verbose)
	if dd == nil {
		return 0
	}
	if verbose != 0 {
		fmt.Printf("Shared BDD size = %6d nodes.\n", core.Cudd_ReadKeys(dd)-core.Cudd_ReadDead(dd))
	}
	transition := Abc_NtkTransitionRelation(dd, network, verbose)
	if transition == nil {
		Abc_NtkFreeGlobalBdds(network, 1)
		return 0
	}
	core.Cudd_Ref(transition)
	initial := Abc_NtkInitStateAndVarMap(dd, network, verbose)
	if initial == nil {
		core.Cudd_RecursiveDeref(dd, transition)
		Abc_NtkFreeGlobalBdds(network, 1)
		return 0
	}
	core.Cudd_Ref(initial)
	unreachable := Abc_NtkComputeUnreachable(dd, network, transition, initial, verbose)
	if unreachable == nil {
		core.Cudd_RecursiveDeref(dd, transition)
		core.Cudd_RecursiveDeref(dd, initial)
		Abc_NtkFreeGlobalBdds(network, 1)
		return 0
	}
	core.Cudd_Ref(unreachable)
	core.Cudd_RecursiveDeref(dd, transition)
	core.Cudd_RecursiveDeref(dd, initial)
	core.Cudd_ReduceHeap(dd, core.CUDD_REORDER_SYMM_SIFT, 1)
	core.Cudd_AutodynDisable(dd)
	core.Cudd_zddVarsFromBddVars(dd, 2)
	network.Exdc = Abc_NtkConstructExdc(dd, network, unreachable)
	core.Cudd_RecursiveDeref(dd, unreachable)
	Abc_NtkFreeGlobalBdds(network, 1)
	if network.Exdc != nil && abc.Abc_NtkCheck(network.Exdc) == 0 {
		fmt.Println("Abc_NtkExtractSequentialDcs: The network check has failed.")
		abc.Abc_NtkDelete(network.Exdc)
		network.Exdc = nil
		return 0
	}
	return 1
}
