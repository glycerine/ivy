package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Abc_NodeConstantInput replaces a BDD node's function by its cofactor at the
// given constant input. Fanin removal is handled by minimum-base processing.
func Abc_NodeConstantInput(node, fanin *Object, constZero int32) {
	if Abc_NtkIsBddLogic(node.network) == 0 {
		panic("abc: constant input requires BDD logic")
	}
	index := Abc_ObjFanoutFaninNum(node, fanin)
	if index < 0 {
		fmt.Printf("Node %s should be among the fanins of node %s...\n", fanin.Name(), node.Name())
		return
	}
	dd := node.network.FunctionManager.(*core.DdManager)
	variable := core.Cudd_NotCond(core.Cudd_bddIthVar(dd, int(index)), constZero != 0)
	old := node.Data.(*core.DdNode)
	replacement := core.Cudd_Cofactor(dd, old, variable)
	core.Cudd_Ref(replacement)
	core.Cudd_RecursiveDeref(dd, old)
	node.Data = replacement
}

// sweepComplementInput is the BDD branch of the C Abc_NodeComplementInput.
func sweepComplementInput(node, fanin *Object) {
	index := Abc_ObjFanoutFaninNum(node, fanin)
	if index < 0 {
		fmt.Printf("Node %s should be among the fanins of node %s...\n", fanin.Name(), node.Name())
		return
	}
	dd := node.network.FunctionManager.(*core.DdManager)
	variable := core.Cudd_bddIthVar(dd, int(index))
	old := node.Data.(*core.DdNode)
	cof0 := core.Cudd_Cofactor(dd, old, core.Cudd_Not(variable))
	core.Cudd_Ref(cof0)
	cof1 := core.Cudd_Cofactor(dd, old, variable)
	core.Cudd_Ref(cof1)
	core.Cudd_RecursiveDeref(dd, old)
	replacement := core.Cudd_bddIte(dd, variable, cof0, cof1)
	core.Cudd_Ref(replacement)
	core.Cudd_RecursiveDeref(dd, cof0)
	core.Cudd_RecursiveDeref(dd, cof1)
	node.Data = replacement
}

// Abc_NtkSweep ports the ABC_USE_CUDD implementation in abcSweep.c.
func Abc_NtkSweep(n *Network, verbose int32) int32 {
	if Abc_NtkIsLogic(n) == 0 {
		panic("abc: sweep requires logic network")
	}
	if Abc_NtkToBdd(n) == 0 {
		fmt.Println("Converting to BDD has failed.")
		return 1
	}
	oldCount := Abc_NtkNodeNum(n)
	Abc_NtkCleanup(n, 0)
	Abc_NtkRemoveDupFanins(n)
	Abc_NtkMinimumBase(n)
	stack := make([]*Object, 0, 100)
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if Abc_ObjFaninNum(node) < 2 {
			stack = append(stack, node)
		}
	}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if !n.contains(node) || Abc_ObjIsNode(node) == 0 {
			continue
		}
		fanout := Abc_NodeFindNonCoFanout(node)
		if fanout == nil {
			continue
		}
		if Abc_ObjIsNode(fanout) == 0 {
			panic("abc: non-CO fanout must be node")
		}
		if Abc_ObjFaninNum(node) == 0 {
			Abc_NodeConstantInput(fanout, node, Abc_NodeIsConst0(node))
		} else {
			if Abc_ObjFaninNum(node) != 1 {
				panic("abc: sweep node must have at most one input")
			}
			driver := Abc_ObjFanin0(node)
			if Abc_NodeIsInv(node) != 0 {
				sweepComplementInput(fanout, node)
			}
			Abc_ObjPatchFanin(fanout, node, driver)
		}
		Abc_NodeRemoveDupFanins(fanout)
		Abc_NodeMinimumBase(fanout)
		if Abc_ObjFaninNum(fanout) < 2 {
			stack = append(stack, fanout)
		}
		if Abc_ObjFanoutNum(node) > 0 {
			stack = append(stack, node)
		} else {
			Abc_NtkDeleteObj_rec(node, 1)
		}
	}
	for i := int32(0); i < Abc_NtkCoNum(n); i++ {
		co := Abc_NtkCo(n, i)
		node := Abc_ObjFanin0(co)
		if Abc_ObjFaninNum(node) != 1 {
			continue
		}
		driver := Abc_ObjFanin0(node)
		if Abc_ObjFanoutNum(driver) != 1 || Abc_ObjIsNode(driver) == 0 {
			continue
		}
		if Abc_NodeIsInv(node) != 0 {
			driver.Data = core.Cudd_Not(driver.Data.(*core.DdNode))
		}
		Abc_ObjPatchFanin(co, node, driver)
	}
	Abc_NtkCleanup(n, 0)
	removed := oldCount - Abc_NtkNodeNum(n)
	if verbose != 0 {
		fmt.Printf("Sweep removed %d nodes.\n", removed)
	}
	return removed
}
