package abci

import (
	"fmt"
	"os"
	"strings"

	"github.com/glycerine/abcgo/aig/ivy"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
)

type ivyHopAdapter struct{ manager *ivy.Ivy_Man_t }

func (n ivyHopAdapter) ConstantOne() any         { return ivy.Ivy_ManConst1(n.manager) }
func (n ivyHopAdapter) Complement(value any) any { return ivy.Ivy_Not(value.(*ivy.Ivy_Obj_t)) }
func (n ivyHopAdapter) And(a, b any) any {
	return ivy.Ivy_And(n.manager, a.(*ivy.Ivy_Obj_t), b.(*ivy.Ivy_Obj_t))
}

func ivyABCEdge(edge abc.Edge) int32 {
	assert(edge.Object != nil, "ABC edge has an object")
	value := int32(edge.Object.ID << 1)
	if edge.Complement {
		value |= 1
	}
	return value
}

func ivyCopyOf(object *abc.Object) *ivy.Ivy_Obj_t {
	copy, ok := object.CopyData.(*ivy.Ivy_Obj_t)
	assert(ok && copy != nil, "ABC object has an Ivy copy")
	return copy
}

func setIvyCopy(object *abc.Object, copy *ivy.Ivy_Obj_t) { object.CopyData = copy }

// Abc_NtkFromIvy reconstructs a combinational Ivy graph in ABC's hashed AIG.
func Abc_NtkFromIvy(old *abc.Network, manager *ivy.Ivy_Man_t) *abc.Network {
	network := abc.Abc_NtkStartFrom(old, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	ivy.Ivy_ManConst1(manager).TravId = ivyABCEdge(abc.Abc_AigConst1(network))
	for i := int32(0); i < abc.Abc_NtkCiNum(old); i++ {
		ivy.Ivy_ManPi(manager, i).TravId = ivyABCEdge(abc.Abc_ObjCopyEdge(abc.Abc_NtkCi(old, i)))
	}
	aigManager := network.FunctionManager.(*abc.Abc_Aig_t)
	for _, id := range ivy.Ivy_ManDfs(manager) {
		node := ivy.Ivy_ManObj(manager, id)
		fanin0 := Abc_ObjFanin0Ivy(network, node)
		if ivy.Ivy_ObjIsBuf(node) != 0 {
			node.TravId = ivyABCEdge(fanin0)
			continue
		}
		fanin1 := Abc_ObjFanin1Ivy(network, node)
		var result abc.Edge
		if ivy.Ivy_ObjIsExor(node) != 0 {
			result = abc.Abc_AigXor(aigManager, fanin0, fanin1)
		} else {
			result = abc.Abc_AigAnd(aigManager, fanin0, fanin1)
		}
		node.TravId = ivyABCEdge(result)
	}
	for i := int32(0); i < abc.Abc_NtkCoNum(old); i++ {
		oldCo := abc.Abc_NtkCo(old, i)
		abc.Abc_ObjAddFanin(oldCo.Copy, Abc_ObjFanin0Ivy(network, ivy.Ivy_ManPo(manager, i)))
	}
	if abc.Abc_NtkCheck(network) == 0 {
		fmt.Fprintln(os.Stdout, "Abc_NtkFromIvy(): Network check has failed.")
	}
	return network
}

// Abc_NtkFromIvySeq reconstructs Ivy logic, latch state, and optional choice chains.
func Abc_NtkFromIvySeq(old *abc.Network, manager *ivy.Ivy_Man_t, haig int32) *abc.Network {
	network := abc.Abc_NtkStartFromNoLatches(old, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	ivy.Ivy_ManConst1(manager).TravId = ivyABCEdge(abc.Abc_AigConst1(network))
	for i := int32(0); i < abc.Abc_NtkPiNum(old); i++ {
		ivy.Ivy_ManPi(manager, i).TravId = ivyABCEdge(abc.Abc_ObjCopyEdge(abc.Abc_NtkPi(old, i)))
	}
	var latchIDs []int32
	nodeIDs := ivy.Ivy_ManDfsSeq(manager, &latchIDs)
	for _, id := range latchIDs {
		latch := ivy.Ivy_ManObj(manager, id)
		box := abc.Abc_NtkCreateLatch(network)
		bi, bo := abc.Abc_NtkCreateBi(network), abc.Abc_NtkCreateBo(network)
		abc.Abc_ObjAddFanin(box, bi)
		abc.Abc_ObjAddFanin(bo, box)
		switch {
		case haig != 0 || ivy.Ivy_ObjInit(latch) == ivy.IVY_INIT_DC:
			abc.Abc_LatchSetInitDc(box)
		case ivy.Ivy_ObjInit(latch) == ivy.IVY_INIT_1:
			abc.Abc_LatchSetInit1(box)
		case ivy.Ivy_ObjInit(latch) == ivy.IVY_INIT_0:
			abc.Abc_LatchSetInit0(box)
		default:
			assert(false, "Ivy latch has an initial value")
		}
		latch.TravId = ivyABCEdge(bo.Signal())
	}
	abc.Abc_NtkAddDummyBoxNames(network)
	aigManager := network.FunctionManager.(*abc.Abc_Aig_t)
	for _, id := range nodeIDs {
		node := ivy.Ivy_ManObj(manager, id)
		fanin0 := Abc_ObjFanin0Ivy(network, node)
		if ivy.Ivy_ObjIsBuf(node) != 0 {
			node.TravId = ivyABCEdge(fanin0)
			continue
		}
		fanin1 := Abc_ObjFanin1Ivy(network, node)
		var result abc.Edge
		if ivy.Ivy_ObjIsExor(node) != 0 {
			result = abc.Abc_AigXor(aigManager, fanin0, fanin1)
		} else {
			result = abc.Abc_AigAnd(aigManager, fanin0, fanin1)
		}
		node.TravId = ivyABCEdge(result)
		choice := ivy.Ivy_ObjEquiv(node)
		if haig != 0 && choice != nil && ivy.Ivy_ObjRefs(node) > 0 {
			assert(ivy.Ivy_IsComplement(choice) == 0, "choice representative is regular")
			previous := Abc_EdgeToNode(network, Abc_Edge_t(node.TravId)).Object
			for alternative := choice; alternative != node; alternative = ivy.Ivy_Regular(ivy.Ivy_ObjEquiv(alternative)) {
				current := Abc_EdgeToNode(network, Abc_Edge_t(alternative.TravId)).Object
				previous.Choice = current
				previous = current
			}
			previous.Choice = nil
		}
	}
	for i := int32(0); i < abc.Abc_NtkPoNum(old); i++ {
		abc.Abc_ObjAddFanin(abc.Abc_NtkPo(network, i), Abc_ObjFanin0Ivy(network, ivy.Ivy_ManPo(manager, i)))
	}
	for i, id := range latchIDs {
		latch := ivy.Ivy_ManObj(manager, id)
		box := abc.Abc_NtkBox(network, int32(i))
		abc.Abc_ObjAddFanin(abc.Abc_ObjFanin0(box), Abc_ObjFanin0Ivy(network, latch))
	}
	if abc.Abc_NtkCheck(network) == 0 {
		fmt.Fprintln(os.Stdout, "Abc_NtkFromIvySeq(): Network check has failed.")
	}
	return network
}

// Abc_NtkToIvy converts a SOP logic network or strashed graph into Ivy.
func Abc_NtkToIvy(old *abc.Network) *ivy.Ivy_Man_t {
	assert(abc.Abc_NtkHasSop(old) != 0 || abc.Abc_NtkIsStrash(old) != 0, "network has SOP logic or is strashed")
	manager := ivy.Ivy_ManStart()
	abc.Abc_NtkCleanCopy(old)
	if abc.Abc_NtkIsStrash(old) != 0 {
		setIvyCopy(abc.Abc_AigConst1(old).Object, ivy.Ivy_ManConst1(manager))
	}
	for i := int32(0); i < abc.Abc_NtkCiNum(old); i++ {
		setIvyCopy(abc.Abc_NtkCi(old, i), ivy.Ivy_ObjCreatePi(manager))
	}
	Abc_NtkStrashPerformAig(old, manager)
	for i := int32(0); i < abc.Abc_NtkCoNum(old); i++ {
		co := abc.Abc_NtkCo(old, i)
		fanin := ivyCopyOf(abc.Abc_ObjFanin0(co))
		fanin = ivy.Ivy_NotCond(fanin, abc.Abc_ObjFaninC0(co))
		ivy.Ivy_ObjCreatePo(manager, fanin)
	}
	ivy.Ivy_ManCleanup(manager)
	return manager
}

// Abc_NtkStrashPerformAig processes DFS nodes in topological order.
func Abc_NtkStrashPerformAig(network *abc.Network, manager *ivy.Ivy_Man_t) {
	for _, node := range abc.Abc_NtkDfs(network, 0) {
		setIvyCopy(node, Abc_NodeStrashAig(manager, node))
	}
}

func Abc_NodeStrashAig(manager *ivy.Ivy_Man_t, node *abc.Object) *ivy.Ivy_Obj_t {
	assert(abc.Abc_ObjIsNode(node) != 0, "strash target is a logic node")
	if abc.Abc_NtkIsStrash(node.Network()) != 0 {
		if abc.Abc_AigNodeIsConst(node) != 0 {
			return ivy.Ivy_ManConst1(manager)
		}
		fanins := node.Fanins()
		left := ivy.Ivy_NotCond(ivyCopyOf(fanins[0].Object), boolInt(fanins[0].Complement))
		right := ivy.Ivy_NotCond(ivyCopyOf(fanins[1].Object), boolInt(fanins[1].Complement))
		return ivy.Ivy_And(manager, left, right)
	}
	sop := ""
	if abc.Abc_NtkHasMapping(node.Network()) != 0 {
		gate, ok := node.Data.(*mio.Mio_Gate_t)
		assert(ok && gate != nil, "mapped node has a Mio gate")
		sop = mio.Mio_GateReadSop(gate)
	} else {
		sop, _ = node.Data.(string)
	}
	if abc.Abc_NodeIsConst(node) != 0 {
		return ivy.Ivy_NotCond(ivy.Ivy_ManConst1(manager), abc.Abc_SopIsConst0(sop))
	}
	if abc.Abc_ObjFaninNum(node) > 2 && abc.Abc_SopGetCubeNum(sop) > 1 && abc.Abc_SopIsExorType(sop) == 0 {
		return Abc_NodeStrashAigFactorAig(manager, node, sop)
	}
	return Abc_NodeStrashAigSopAig(manager, node, sop)
}

func Abc_NodeStrashAigSopAig(manager *ivy.Ivy_Man_t, node *abc.Object, sop string) *ivy.Ivy_Obj_t {
	assert(int32(len(node.Fanins())) == abc.Abc_SopGetVarNum(sop), "SOP fanin count matches")
	var sum = ivy.Ivy_Not(ivy.Ivy_ManConst1(manager))
	for _, cube := range ivySopCubes(sop, int(abc.Abc_SopGetVarNum(sop))) {
		product := ivy.Ivy_ManConst1(manager)
		for i, edge := range node.Fanins() {
			switch cube[i] {
			case '1':
				product = ivy.Ivy_And(manager, product, ivyCopyOf(edge.Object))
			case '0':
				product = ivy.Ivy_And(manager, product, ivy.Ivy_Not(ivyCopyOf(edge.Object)))
			}
		}
		if abc.Abc_SopIsExorType(sop) != 0 {
			sum = ivy.Ivy_Exor(manager, sum, product)
		} else {
			sum = ivy.Ivy_Or(manager, sum, product)
		}
	}
	return ivy.Ivy_NotCond(sum, abc.Abc_SopIsComplement(sop))
}

func ivySopCubes(sop string, variables int) []string {
	rows := strings.Split(strings.TrimSuffix(sop, "\n"), "\n")
	for i := range rows {
		if len(rows[i]) < variables {
			panic("abc: malformed SOP cube")
		}
		rows[i] = rows[i][:variables]
	}
	return rows
}

func Abc_NodeStrashAigExorAig(manager *ivy.Ivy_Man_t, node *abc.Object, sop string) *ivy.Ivy_Obj_t {
	assert(int32(len(node.Fanins())) == abc.Abc_SopGetVarNum(sop), "XOR SOP fanin count matches")
	sum := ivy.Ivy_Not(ivy.Ivy_ManConst1(manager))
	for _, edge := range node.Fanins() {
		sum = ivy.Ivy_Exor(manager, sum, ivyCopyOf(edge.Object))
	}
	return ivy.Ivy_NotCond(sum, abc.Abc_SopIsComplement(sop))
}

func Abc_NodeStrashAigFactorAig(manager *ivy.Ivy_Man_t, root *abc.Object, sop string) *ivy.Ivy_Obj_t {
	graph := dec.Dec_Factor(sop)
	defer dec.Dec_GraphFree(graph)
	for i, edge := range root.Fanins() {
		dec.Dec_GraphSetLeafFunc(graph, int32(i), ivyCopyOf(edge.Object))
	}
	result := dec.Dec_GraphToNetworkIvy(ivyHopAdapter{manager}, graph)
	ivyNode, ok := result.(*ivy.Ivy_Obj_t)
	assert(ok && ivyNode != nil, "factored SOP builds an Ivy node")
	return ivyNode
}

// Abc_NtkCollectLatchValuesIvy returns latch initialization in network order.
func Abc_NtkCollectLatchValuesIvy(network *abc.Network, useDCs int32) *vec.Vec_Int_t {
	values := vec.Vec_IntAlloc(int(abc.Abc_NtkLatchNum(network)))
	for _, latch := range network.ObjectsOfKind(abc.ObjectLatch) {
		switch {
		case useDCs != 0 || abc.Abc_LatchIsInitDc(latch) != 0:
			vec.Vec_IntPush(values, int32(ivy.IVY_INIT_DC))
		case abc.Abc_LatchIsInit1(latch) != 0:
			vec.Vec_IntPush(values, int32(ivy.IVY_INIT_1))
		case abc.Abc_LatchIsInit0(latch) != 0:
			vec.Vec_IntPush(values, int32(ivy.IVY_INIT_0))
		default:
			assert(false, "latch initialization is known")
		}
	}
	return values
}
