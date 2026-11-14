package abci

import (
	"fmt"
	"math"
	"time"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/extrab"
)

func abcReachVarMap(dd *core.DdManager, piNum, ciNum, latchNum int, function *core.DdNode) *core.DdNode {
	mapping := make([]int, len(dd.Vars))
	for i := range mapping {
		mapping[i] = i
	}
	for i := 0; i < latchNum; i++ {
		x, y := piNum+i, ciNum+i
		if x >= len(mapping) || y >= len(mapping) {
			return nil
		}
		mapping[x], mapping[y] = y, x
	}
	return core.Cudd_bddPermute(dd, function, mapping)
}

// Abc_NtkInitStateVarMap returns the source-defined initial-state cube and
// configures the present/next-state variable swap used by image iteration.
func Abc_NtkInitStateVarMap(dd *core.DdManager, network *abc.Network, verbose int32) *core.DdNode {
	if dd == nil || network == nil {
		return nil
	}
	piNum, ciNum, latches := len(network.ObjectsOfKind(abc.ObjectPI)), len(network.Inputs()), network.ObjectsOfKind(abc.ObjectLatch)
	for i := 0; i < ciNum+len(latches); i++ {
		core.Cudd_bddIthVar(dd, i)
	}
	initial := core.Cudd_ReadOne(dd)
	core.Cudd_Ref(initial)
	for i, latch := range latches {
		variable := core.Cudd_bddIthVar(dd, piNum+i)
		if abc.Abc_LatchIsInit1(latch) == 0 {
			variable = core.Cudd_Not(variable)
		}
		next := core.Cudd_bddAnd(dd, initial, variable)
		core.Cudd_Ref(next)
		core.Cudd_RecursiveDeref(dd, initial)
		initial = next
	}
	_ = verbose
	core.Cudd_Deref(initial)
	return initial
}

// Abc_NtkCreatePartitions builds one XNOR partition per latch and releases
// global object BDD references after extracting the transition functions.
func Abc_NtkCreatePartitions(dd *core.DdManager, network *abc.Network, reorder, verbose int32) []*core.DdNode {
	if dd == nil || network == nil {
		return nil
	}
	ciNum := len(network.Inputs())
	latches := network.ObjectsOfKind(abc.ObjectLatch)
	for i := 0; i < ciNum+len(latches); i++ {
		core.Cudd_bddIthVar(dd, i)
	}
	if reorder != 0 {
		core.Cudd_AutodynEnable(dd, core.CUDD_REORDER_SYMM_SIFT)
	} else {
		core.Cudd_AutodynDisable(dd)
	}
	parts := make([]*core.DdNode, len(latches))
	for i, latch := range latches {
		if len(latch.Fanins()) != 1 {
			Abc_NtkFreeGlobalBdds(network, 0)
			return nil
		}
		bi := latch.Fanins()[0].Object
		function := Abc_ObjGlobalBdd(bi)
		if function == nil {
			Abc_NtkFreeGlobalBdds(network, 0)
			return nil
		}
		parts[i] = core.Cudd_bddXnor(dd, core.Cudd_bddIthVar(dd, ciNum+i), function)
		core.Cudd_Ref(parts[i])
	}
	Abc_NtkFreeGlobalBdds(network, 0)
	if reorder != 0 {
		if verbose != 0 {
			fmt.Printf("BDD nodes in the partitions before reordering %d.\n", abcBDDSharingSize(parts))
		}
		core.Cudd_ReduceHeap(dd, core.CUDD_REORDER_SYMM_SIFT, 100)
		core.Cudd_AutodynDisable(dd)
		if verbose != 0 {
			fmt.Printf("BDD nodes in the partitions after reordering %d.\n", abcBDDSharingSize(parts))
		}
	}
	return parts
}

func abcBDDSharingSize(functions []*core.DdNode) int {
	seen := map[*core.DdNode]bool{}
	var visit func(*core.DdNode)
	visit = func(edge *core.DdNode) {
		node := core.Cudd_Regular(edge)
		if node == nil || core.Cudd_IsConstant(node) || seen[node] {
			return
		}
		seen[node] = true
		visit(node.T)
		visit(node.E)
	}
	for _, f := range functions {
		visit(f)
	}
	return len(seen)
}

// Abc_NtkComputeReachable performs the C partitioned or monolithic image loop
// and returns the computed reached-state BDD.
func Abc_NtkComputeReachable(dd *core.DdManager, network *abc.Network, parts []*core.DdNode, initial, output *core.DdNode, maxBDD, maxIterations, partitioned, reorder, verbose int32) *core.DdNode {
	if dd == nil || network == nil || initial == nil || output == nil || maxIterations <= 1 || len(parts) != len(network.ObjectsOfKind(abc.ObjectLatch)) {
		return nil
	}
	piNum, ciNum, latches := len(network.ObjectsOfKind(abc.ObjectPI)), len(network.Inputs()), network.ObjectsOfKind(abc.ObjectLatch)
	care := extrab.Extra_bddComputeRangeCube(dd, piNum, ciNum)
	if care == nil {
		return nil
	}
	core.Cudd_Ref(care)
	vars := make([]*core.DdNode, len(latches))
	for i := range vars {
		vars[i] = core.Cudd_bddIthVar(dd, ciNum+i)
	}
	var tree *extrab.Extra_ImageTree_t
	var mono *extrab.Extra_ImageTree2_t
	if partitioned != 0 {
		tree = extrab.Extra_bddImageStart(dd, care, len(parts), parts, len(vars), vars, verbose != 0)
		if tree == nil {
			core.Cudd_RecursiveDeref(dd, care)
			return nil
		}
		defer extrab.Extra_bddImageTreeDelete(tree)
	} else {
		mono = extrab.Extra_bddImageStart2(dd, care, len(parts), parts, len(vars), vars, verbose != 0)
		if mono == nil {
			core.Cudd_RecursiveDeref(dd, care)
			return nil
		}
		defer extrab.Extra_bddImageTreeDelete2(mono)
	}
	core.Cudd_RecursiveDeref(dd, care)
	current, reached := initial, initial
	core.Cudd_Ref(current)
	core.Cudd_Ref(reached)
	var next *core.DdNode
	hasCurrent := true
	iterations := 0
	converged := false
	for iteration := 1; iteration <= int(maxIterations); iteration++ {
		iterations = iteration
		if tree != nil {
			next = extrab.Extra_bddImageCompute(tree, current)
		} else {
			next = extrab.Extra_bddImageCompute2(mono, current)
		}
		if next == nil {
			core.Cudd_RecursiveDeref(dd, current)
			core.Cudd_RecursiveDeref(dd, reached)
			return nil
		}
		core.Cudd_Ref(next)
		core.Cudd_RecursiveDeref(dd, current)
		hasCurrent = false
		mapped := abcReachVarMap(dd, piNum, ciNum, len(latches), next)
		if mapped == nil {
			core.Cudd_RecursiveDeref(dd, next)
			core.Cudd_RecursiveDeref(dd, reached)
			return nil
		}
		core.Cudd_Ref(mapped)
		core.Cudd_RecursiveDeref(dd, next)
		next = mapped
		if core.Cudd_bddLeq(dd, next, reached) {
			converged = true
			break
		}
		nextSize := core.Cudd_DagSize(next)
		if nextSize > int(maxBDD) {
			break
		}
		current = core.Cudd_bddAnd(dd, next, core.Cudd_Not(reached))
		core.Cudd_Ref(current)
		hasCurrent = true
		newReached := core.Cudd_bddOr(dd, reached, next)
		core.Cudd_Ref(newReached)
		core.Cudd_RecursiveDeref(dd, reached)
		reached = newReached
		core.Cudd_RecursiveDeref(dd, next)
		next = nil
		if verbose != 0 {
			fmt.Printf("Iteration = %3d. BDD = %5d. ", iteration, nextSize)
		}
	}
	exhausted := !converged && iterations == int(maxIterations) && next == nil
	reportedIterations := iterations
	if exhausted {
		reportedIterations++
	}
	if next != nil {
		core.Cudd_RecursiveDeref(dd, next)
	}
	if hasCurrent {
		core.Cudd_RecursiveDeref(dd, current)
	}
	if verbose != 0 {
		minterms := core.Cudd_CountMinterm(dd, reached, len(latches))
		if exhausted || core.Cudd_DagSize(reached) > int(maxBDD) {
			fmt.Printf("Reachability analysis is stopped after %d iterations.\n", reportedIterations)
		} else {
			fmt.Printf("Reachability analysis completed in %d iterations.\n", iterations)
		}
		fmt.Printf("Reachable states = %.0f. (Ratio = %.4f %%)\n", minterms, 100*minterms/math.Pow(2, float64(len(latches))))
	}
	if exhausted || core.Cudd_DagSize(reached) > int(maxBDD) {
		fmt.Printf("Verified ONLY FOR STATES REACHED in %d iterations. \n", reportedIterations)
	}
	fmt.Printf("The miter is proved unreachable in %d iteration.  ", reportedIterations)
	core.Cudd_Deref(reached)
	return reached
}

func Abc_NtkVerifyUsingBdds(network *abc.Network, maxBDD, maxIterations, partitioned, reorder, verbose int32) {
	started := time.Now()
	if network == nil || abc.Abc_NtkIsStrash(network) == 0 || len(network.ObjectsOfKind(abc.ObjectPO)) != 1 || len(network.ObjectsOfKind(abc.ObjectLatch)) == 0 {
		return
	}
	if abc.Abc_ObjFanoutNum(network.ObjectsOfKind(abc.ObjectPO)[0]) != 0 {
		return
	}
	dd := Abc_NtkBuildGlobalBdds(network, maxBDD, 1, reorder, 0, verbose)
	if dd == nil {
		fmt.Printf("The number of intermediate BDD nodes exceeded the limit (%d).\n", maxBDD)
		return
	}
	defer func() { Abc_NtkFreeGlobalBdds(network, 1) }()
	if verbose != 0 {
		fmt.Printf("Shared BDD size is %6d nodes.\n", core.Cudd_ReadKeys(dd)-core.Cudd_ReadDead(dd))
	}
	outputs := network.ObjectsOfKind(abc.ObjectPO)
	function := Abc_ObjGlobalBdd(outputs[0])
	if function == nil {
		return
	}
	core.Cudd_Ref(function)
	parts := Abc_NtkCreatePartitions(dd, network, reorder, verbose)
	if parts == nil {
		core.Cudd_RecursiveDeref(dd, function)
		return
	}
	initial := Abc_NtkInitStateVarMap(dd, network, verbose)
	if initial == nil {
		core.Cudd_RecursiveDeref(dd, function)
		for _, part := range parts {
			core.Cudd_RecursiveDeref(dd, part)
		}
		return
	}
	core.Cudd_Ref(initial)
	if !core.Cudd_bddLeq(dd, initial, core.Cudd_Not(function)) {
		fmt.Print("The miter is proved REACHABLE in the initial state.  ")
	} else {
		reached := Abc_NtkComputeReachable(dd, network, parts, initial, function, maxBDD, maxIterations, partitioned, reorder, verbose)
		if reached != nil {
			core.Cudd_Ref(reached)
			core.Cudd_RecursiveDeref(dd, reached)
		}
	}
	core.Cudd_RecursiveDeref(dd, function)
	core.Cudd_RecursiveDeref(dd, initial)
	for _, part := range parts {
		core.Cudd_RecursiveDeref(dd, part)
	}
	fmt.Printf("Time: %.2fs\n", time.Since(started).Seconds())
}
