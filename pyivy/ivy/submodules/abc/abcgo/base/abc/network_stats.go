package abc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

func Abc_NtkAttrFree(n *Network, attribute, freeManager int32) any {
	if attribute < 0 || int(attribute) >= len(n.Attributes) {
		panic("abc: attribute index out of range")
	}
	item := n.Attributes[attribute]
	n.Attributes[attribute] = nil
	return vec.Vec_AttFree(item, int(freeManager))
}

func sopNodes(n *Network, skipConstants bool, count func(string) int32) int32 {
	if Abc_NtkHasSop(n) == 0 {
		panic("abc: SOP statistics require an SOP network")
	}
	var total int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if skipConstants && Abc_NodeIsConst(node) != 0 {
			continue
		}
		sop, ok := node.Data.(string)
		if !ok {
			panic("abc: node has no SOP")
		}
		total += count(sop)
	}
	return total
}

func Abc_NtkGetCubeNum(n *Network) int32 {
	return sopNodes(n, true, Abc_SopGetCubeNum)
}

func Abc_NtkGetCubePairNum(n *Network) int32 {
	if Abc_NtkHasSop(n) == 0 {
		panic("abc: cube-pair statistics require an SOP network")
	}
	var total int64
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if Abc_NodeIsConst(node) != 0 {
			continue
		}
		sop, ok := node.Data.(string)
		if !ok {
			panic("abc: node has no SOP")
		}
		cubes := int64(Abc_SopGetCubeNum(sop))
		total += cubes * (cubes - 1) / 2
		if total >= 1<<30 {
			return 1 << 30
		}
	}
	return int32(total)
}

func Abc_NtkGetLitNum(n *Network) int32 { return sopNodes(n, false, Abc_SopGetLitNum) }

func Abc_NtkGetLitFactNum(n *Network) int32 {
	return sopNodes(n, true, func(sop string) int32 {
		factor := dec.Dec_Factor(sop)
		defer dec.Dec_GraphFree(factor)
		return 1 + dec.Dec_GraphNodeNum(factor)
	})
}

func Abc_NtkGetMultiRefNum(n *Network) int32 {
	if Abc_NtkIsStrash(n) == 0 {
		panic("abc: shared-node count requires an AIG")
	}
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if len(node.fanouts) > 1 {
			count++
		}
	}
	return count
}

func Abc_NtkGetBddNodeNum(n *Network) int32 {
	if Abc_NtkIsBddLogic(n) == 0 {
		panic("abc: BDD statistics require BDD logic")
	}
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if len(node.fanins) >= 2 {
			root, ok := node.Data.(*core.DdNode)
			if !ok || root == nil {
				panic("abc: node has no BDD")
			}
			count += int32(core.Cudd_DagSize(root) - 1)
		}
	}
	return count
}

func Abc_NtkGetAigNodeNum(n *Network) int32 {
	if Abc_NtkIsAigLogic(n) == 0 {
		panic("abc: AIG statistics require AIG logic")
	}
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if len(node.fanins) >= 2 {
			root, ok := node.Data.(*hop.Hop_Obj_t)
			if !ok || root == nil {
				panic("abc: node has no local AIG")
			}
			count += hop.Hop_DagSize(root)
		}
	}
	return count
}

func Abc_NtkGetClauseNum(n *Network) int32 {
	if Abc_NtkIsBddLogic(n) == 0 {
		panic("abc: clause count requires BDD logic")
	}
	dd, ok := n.FunctionManager.(*core.DdManager)
	if !ok || dd == nil {
		panic("abc: BDD manager missing")
	}
	if dd.ZddSize < 2*len(dd.Vars) && core.Cudd_zddVarsFromBddVars(dd, 2) == 0 {
		panic("abc: cannot initialize ZDD variables")
	}
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		root, ok := node.Data.(*core.DdNode)
		if !ok || root == nil {
			panic("abc: node has no BDD")
		}
		for _, function := range []*core.DdNode{root, core.Cudd_Not(root)} {
			var cover *core.DdNode
			bdd := core.Cudd_zddIsop(dd, function, function, &cover)
			if bdd == nil || cover == nil {
				panic("abc: cannot construct ISOP cover")
			}
			core.Cudd_Ref(bdd)
			core.Cudd_Ref(cover)
			count += Abc_CountZddCubes(dd, cover)
			core.Cudd_RecursiveDeref(dd, bdd)
			core.Cudd_RecursiveDeref(dd, cover)
		}
	}
	return count
}

func Abc_NtkFetchTwinNode(node *Object) *Object {
	if node == nil || node.network == nil || Abc_NtkHasMapping(node.network) == 0 {
		panic("abc: twin lookup requires a mapped node")
	}
	gate, ok := node.Data.(*mio.Mio_Gate_t)
	if !ok || gate == nil || mio.Mio_GateReadTwin(gate) == nil {
		return nil
	}
	twin := node.network.Object(node.ID + 1)
	if twin == nil || twin.Kind != ObjectNode || len(twin.fanins) != len(node.fanins) {
		return nil
	}
	twinGate, ok := twin.Data.(*mio.Mio_Gate_t)
	if !ok || twinGate != mio.Mio_GateReadTwin(gate) {
		return nil
	}
	return twin
}

func Abc_NtkGetMappedArea(n *Network) float64 {
	if Abc_NtkHasMapping(n) == 0 {
		panic("abc: mapped area requires a mapped network")
	}
	nodes := n.ObjectsOfKind(ObjectNode)
	var area float64
	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		if node.Barrier {
			continue
		}
		gate, ok := node.Data.(*mio.Mio_Gate_t)
		if !ok || gate == nil {
			fmt.Println("Node without mapping is encountered.")
			continue
		}
		area += mio.Mio_GateReadArea(gate)
		if Abc_NtkFetchTwinNode(node) != nil {
			i++
		}
	}
	return area
}

func Abc_NtkGetBufNum(n *Network) int32 {
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if len(node.fanins) == 1 {
			count++
		}
	}
	return count
}

func Abc_NtkGetLargeNodeNum(n *Network) int32 {
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if len(node.fanins) > 1 {
			count++
		}
	}
	return count
}

func Abc_NtkGetChoiceNum(n *Network) int32 {
	if Abc_NtkIsStrash(n) == 0 {
		return 0
	}
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		count += Abc_AigNodeIsChoice(node)
	}
	return count
}

func Abc_NtkGetFaninMax(n *Network) int32 {
	var maximum int
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		maximum = max(maximum, len(node.fanins))
	}
	return int32(maximum)
}

func Abc_NtkGetFanoutMax(n *Network) int32 {
	var maximum int
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		maximum = max(maximum, len(node.fanouts))
	}
	return int32(maximum)
}

func Abc_NtkGetTotalFanins(n *Network) int32 {
	var count int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		count += int32(len(node.fanins))
	}
	return count
}
