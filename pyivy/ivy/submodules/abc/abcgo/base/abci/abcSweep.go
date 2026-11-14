package abci

import (
	"fmt"
	"os"
	"sort"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/fraig"
)

// Abc_NtkFraigSweep sweeps equivalent nodes while retaining the C routine's
// mapped-network restoration and cleanup order.
func Abc_NtkFraigSweep(network *Abc_Ntk_t, useInv, useExdc, verbose, veryVerbose int32) int32 {
	if abc.Abc_NtkIsStrash(network) != 0 {
		panic("Abc_NtkFraigSweep requires a non-strashed network")
	}
	var savedMappedData map[*Abc_Obj_t]any
	if abc.Abc_NtkIsMappedLogic(network) != 0 {
		savedMappedData = make(map[*Abc_Obj_t]any)
		for _, node := range abc.Abc_NtkNodes(network) {
			savedMappedData[node] = node.Data
		}
	}
	aig := abc.Abc_NtkStrash(network, 0, 1, 0)
	if savedMappedData != nil {
		if hopMan, ok := network.FunctionManager.(*hop.Hop_Man_t); ok {
			hop.Hop_ManStop(hopMan)
		}
		network.FunctionManager = abcmain.Abc_FrameReadLibGen()
		network.Function = uint8(abc.ABC_FUNC_MAP)
		for node, data := range savedMappedData {
			node.Data = data
		}
	}
	params := &Fraig_Params_t{}
	fraig.Fraig_ParamsSetDefault(params)
	params.FInternal = 1
	manager := Abc_NtkToFraig(aig, params, 0, 0).(*Fraig_Man_t)
	// The FRAIG conversion stores mappings on internal AIG nodes. Mirror the
	// C pCopy links for CIs and the constant so complemented buffer nodes can
	// participate in the same equivalence table.
	aigConstant := abc.Abc_AigConst1(aig).Object
	aigConstant.CopyData = fraig.Fraig_ManReadConst1(manager)
	for i, input := range aig.Inputs() {
		input.CopyData = fraig.Fraig_ManReadIthVar(manager, int32(i))
	}
	if useExdc != 0 {
		if network.Exdc == nil {
			fmt.Println("Warning: Networks has no EXDC.")
		} else {
			Abc_NtkFraigSweepUsingExdc(manager, network)
		}
	}
	abc.Abc_NtkLevel(network)
	classes := Abc_NtkFraigEquiv(network, useInv, verbose, veryVerbose)
	Abc_NtkFraigTransform(network, classes, useInv, verbose)
	fraig.Fraig_ManFree(manager)
	abc.Abc_NtkDelete(aig)
	for _, node := range network.Objects() {
		node.Copy, node.CopyEdge = nil, abc.Edge{}
	}
	if abc.Abc_NtkHasMapping(network) != 0 {
		abc.Abc_NtkCleanup(network, verbose)
	} else {
		Abc_NtkSweep(network, verbose)
	}
	if abc.Abc_NtkCheck(network) == 0 {
		fmt.Println("Abc_NtkFraigSweep: The network check has failed.")
		return 0
	}
	return 1
}

func Abc_NtkFraigSweepUsingExdc(manager *Fraig_Man_t, network *Abc_Ntk_t) {
	if network.Exdc == nil {
		panic("Abc_NtkFraigSweepUsingExdc requires EXDC")
	}
	exdc := Abc_NtkToFraigExdc(manager, network, network.Exdc)
	for _, node := range abc.Abc_NtkNodes(network) {
		if len(node.Fanins()) == 0 {
			continue
		}
		aigNode := node.Copy
		if aigNode == nil {
			continue
		}
		gNode, ok := aigNode.CopyData.(*Fraig_Node_t)
		if !ok || gNode == nil {
			panic("FRAIG conversion did not annotate a strashed node")
		}
		gNode = fraig.Fraig_NotCond(gNode, abc.Abc_ObjIsComplement(abc.Abc_ObjCopyEdge(node)) != 0)
		result := fraig.Fraig_NodeAnd(manager, gNode, fraig.Fraig_Not(exdc))
		aigNode.CopyData = fraig.Fraig_NotCond(result, abc.Abc_ObjIsComplement(abc.Abc_ObjCopyEdge(node)) != 0)
	}
}

// Abc_NtkFraigEquiv forms one class per regular FRAIG function. Phases remain
// attached to members so the merge routines can distinguish complemented
// representatives exactly as ABC does.
func Abc_NtkFraigEquiv(network *Abc_Ntk_t, useInv, verbose, veryVerbose int32) [][]*Abc_Obj_t {
	byFunction := make(map[*Fraig_Node_t][]*Abc_Obj_t)
	functionCount := 0
	for _, node := range abc.Abc_NtkNodes(network) {
		if len(node.Fanins()) == 0 || abc.Abc_NodeFindCoFanout(node) != nil {
			continue
		}
		aigNode := node.Copy
		if aigNode == nil {
			continue
		}
		function, ok := aigNode.CopyData.(*Fraig_Node_t)
		if !ok || function == nil {
			panic("FRAIG conversion did not annotate a strashed node")
		}
		function = fraig.Fraig_NotCond(function, abc.Abc_ObjIsComplement(abc.Abc_ObjCopyEdge(node)) != 0)
		regular := fraig.Fraig_Regular(function)
		if _, exists := byFunction[regular]; !exists {
			functionCount++
		}
		node.Phase = fraig.Fraig_IsComplement(function)
		byFunction[regular] = append(byFunction[regular], node)
	}
	var classes [][]*Abc_Obj_t
	var members int
	for _, group := range byFunction {
		if len(group) < 2 {
			continue
		}
		classes = append(classes, group)
		members += len(group)
		if veryVerbose != 0 {
			fmt.Printf("Class %2d : {", len(classes)-1)
			for _, node := range group {
				fmt.Printf(" %s(%c)(%d)", abc.Abc_ObjName(node), phaseSign(node.Phase), node.Level)
				node.Copy = nil
			}
			fmt.Println(" }")
		}
	}
	if verbose != 0 || veryVerbose != 0 {
		fmt.Printf("Sweeping stats for network %q:\n", network.Name)
		fmt.Printf("Internal nodes = %d. Different functions (up to compl) = %d.\n", abc.Abc_NtkNodeNum(network), functionCount)
		fmt.Printf("Non-trivial classes = %d. Nodes in non-trivial classes = %d.\n", len(classes), members)
	}
	_ = useInv // retained for the C ABI; phase classes drive the transformation.
	return classes
}

func phaseSign(phase bool) byte {
	if phase {
		return '-'
	}
	return '+'
}

func Abc_NtkFraigTransform(network *Abc_Ntk_t, classes [][]*Abc_Obj_t, useInv, verbose int32) {
	if len(classes) == 0 {
		return
	}
	if abc.Abc_NtkHasMapping(network) != 0 {
		Abc_NtkDelayTrace(network, nil, nil, 0)
		for _, group := range classes {
			Abc_NtkFraigMergeClassMapped(network, group, useInv, verbose)
		}
		return
	}
	for _, group := range classes {
		Abc_NtkFraigMergeClass(network, group, useInv, verbose)
	}
}

func Abc_NtkFraigMergeClassMapped(_ *Abc_Ntk_t, chain []*Abc_Obj_t, _, _ int32) {
	if len(chain) < 2 {
		return
	}
	direct, inverse := splitFraigPhases(chain)
	for _, group := range [][]*Abc_Obj_t{direct, inverse} {
		if len(group) < 2 {
			continue
		}
		best := bestMappedRepresentative(group)
		for _, node := range group {
			if node != best {
				abc.Abc_ObjTransferFanout(node, best)
			}
		}
	}
}

func bestMappedRepresentative(nodes []*Abc_Obj_t) *Abc_Obj_t {
	best := nodes[0]
	for _, node := range nodes[1:] {
		a, b := Abc_NodeReadArrivalWorst(best), Abc_NodeReadArrivalWorst(node)
		if a > b || (a == b && best.Level > node.Level) || (a == b && best.Level == node.Level && Abc_NodeDroppingCost(best) < Abc_NodeDroppingCost(node)) {
			best = node
		}
	}
	return best
}

func splitFraigPhases(nodes []*Abc_Obj_t) ([]*Abc_Obj_t, []*Abc_Obj_t) {
	var direct, inverse []*Abc_Obj_t
	for _, node := range nodes {
		if node.Phase {
			direct = append(direct, node)
		} else {
			inverse = append(inverse, node)
		}
	}
	return direct, inverse
}

func Abc_NtkFraigMergeClass(network *Abc_Ntk_t, chain []*Abc_Obj_t, _, _ int32) {
	if len(chain) < 2 {
		return
	}
	best := chain[0]
	for _, node := range chain[1:] {
		if best.Level > node.Level || (best.Level == node.Level && Abc_NodeDroppingCost(best) < Abc_NodeDroppingCost(node)) {
			best = node
		}
	}
	var direct, inverse []*Abc_Obj_t
	for _, node := range chain {
		if node == best {
			continue
		}
		if node.Phase == best.Phase {
			direct = append(direct, node)
		} else {
			inverse = append(inverse, node)
		}
	}
	for _, node := range direct {
		abc.Abc_ObjTransferFanout(node, best)
	}
	if len(inverse) == 0 {
		return
	}
	inverted := abc.Abc_NtkCreateNodeInv(network, best)
	for _, node := range inverse {
		abc.Abc_ObjTransferFanout(node, inverted)
	}
}

func Abc_NodeDroppingCost(_ *Abc_Obj_t) int32 { return 1 }

func Abc_NtkCleanup(network *Abc_Ntk_t, verbose int32) int32 {
	if abc.Abc_NtkIsLogic(network) == 0 {
		panic("Abc_NtkCleanup requires logic network")
	}
	visited := abc.Abc_NtkDfs(network, 0)
	vector := objectsToVec(visited)
	defer vec.Vec_PtrFree(vector)
	removed := Abc_NtkReduceNodes(network, vector)
	if verbose != 0 {
		fmt.Printf("Cleanup removed %d dangling nodes.\n", removed)
	}
	return removed
}

func Abc_NtkCleanupNodes(network *Abc_Ntk_t, roots *Vec_Ptr_t, verbose int32) int32 {
	if abc.Abc_NtkIsLogic(network) == 0 {
		panic("Abc_NtkCleanupNodes requires logic network")
	}
	starts := make([]*Abc_Obj_t, 0, len(network.Objects()))
	for _, object := range network.Objects() {
		if object.IsCO() {
			starts = append(starts, object)
		}
	}
	for _, raw := range vec.Vec_PtrArray(roots) {
		if root, ok := raw.(*Abc_Obj_t); ok && root != nil {
			starts = append(starts, root)
		}
	}
	visited := abc.Abc_NtkDfsNodes(network, starts, int32(len(starts)))
	vector := objectsToVec(visited)
	defer vec.Vec_PtrFree(vector)
	removed := Abc_NtkReduceNodes(network, vector)
	if verbose != 0 {
		fmt.Printf("Cleanup removed %d dangling nodes.\n", removed)
	}
	return removed
}

func objectsToVec(objects []*Abc_Obj_t) *Vec_Ptr_t {
	v := vec.Vec_PtrAlloc(len(objects))
	for _, object := range objects {
		vec.Vec_PtrPush(v, object)
	}
	return v
}

func Abc_NtkReduceNodes(network *Abc_Ntk_t, reachable *Vec_Ptr_t) int32 {
	if abc.Abc_NtkIsLogic(network) == 0 {
		panic("Abc_NtkReduceNodes requires logic network")
	}
	keep := make(map[*Abc_Obj_t]bool, vec.Vec_PtrSize(reachable))
	for _, raw := range vec.Vec_PtrArray(reachable) {
		if node, ok := raw.(*Abc_Obj_t); ok && node != nil {
			keep[node] = true
			node.MarkA = true
		}
	}
	var removed int32
	for _, node := range abc.Abc_NtkNodes(network) {
		if !keep[node] {
			abc.Abc_NtkDeleteObj(node)
			removed++
		}
	}
	for node := range keep {
		node.MarkA = false
	}
	if abc.Abc_NtkCheck(network) == 0 {
		fmt.Println("Abc_NtkCleanup: The network check has failed.")
	}
	return removed
}

func Abc_NodeConstantInput(node, fanin *Abc_Obj_t, constZero int32) {
	abc.Abc_NodeConstantInput(node, fanin, constZero)
}

func Abc_NodeComplementInput(node, fanin *Abc_Obj_t) {
	abc.Abc_NodeComplementInput(node, fanin)
}

func Abc_NtkSweep(network *Abc_Ntk_t, verbose int32) int32 { return abc.Abc_NtkSweep(network, verbose) }

func Abc_NodeRemoveNonCurrentObjects(network *Abc_Ntk_t) int32 {
	var removed int32
	for _, object := range network.Objects() {
		if abc.Abc_NodeIsTravIdCurrent(object) == 0 {
			abc.Abc_NtkDeleteObj(object)
			removed++
		}
	}
	return removed
}

func Abc_NtkSetTravId_rec(object *Abc_Obj_t) {
	abc.Abc_NodeSetTravIdCurrent(object)
	if abc.Abc_ObjFaninNum(object) == 0 {
		return
	}
	if abc.Abc_ObjFaninNum(object) != 1 {
		panic("Abc_NtkSetTravId_rec requires zero or one fanin")
	}
	Abc_NtkSetTravId_rec(abc.Abc_ObjFanin0(object))
}

func Abc_NtkCheckConstant_rec(object *Abc_Obj_t) int32 {
	if abc.Abc_ObjFaninNum(object) == 0 {
		if abc.Abc_ObjIsNode(object) == 0 {
			return -1
		}
		if abc.Abc_NodeIsConst0(object) != 0 {
			return 0
		}
		if abc.Abc_NodeIsConst1(object) != 0 {
			return 1
		}
		panic("zero-input node is neither constant zero nor one")
	}
	if abc.Abc_ObjIsLatch(object) != 0 || abc.Abc_ObjFaninNum(object) > 1 {
		return -1
	}
	if abc.Abc_ObjIsNode(object) == 0 || abc.Abc_NodeIsBuf(object) != 0 {
		return Abc_NtkCheckConstant_rec(abc.Abc_ObjFanin0(object))
	}
	if abc.Abc_NodeIsInv(object) != 0 {
		value := Abc_NtkCheckConstant_rec(abc.Abc_ObjFanin0(object))
		if value == 0 {
			return 1
		}
		if value == 1 {
			return 0
		}
		return value
	}
	panic("unsupported single-input logic node")
}

func Abc_NtkLatchSweep(network *Abc_Ntk_t) int32 {
	var removed int32
	var pivot *Abc_Obj_t
	for _, latch := range network.ObjectsOfKind(abc.ObjectLatch) {
		value := Abc_NtkCheckConstant_rec(abc.Abc_ObjFanin0(latch))
		if value < 0 {
			continue
		}
		var replacement *Abc_Obj_t
		if (value == 1 && abc.Abc_LatchIsInit0(latch) != 0) || (value == 0 && abc.Abc_LatchIsInit1(latch) != 0) {
			if pivot == nil {
				pivot = latch
				continue
			}
			if latchInitCode(latch) != latchInitCode(pivot) {
				replacement = abc.Abc_NtkCreateNodeInv(network, abc.Abc_ObjFanout0(pivot))
			} else {
				replacement = abc.Abc_ObjFanout0(pivot)
			}
		} else {
			replacement = abc.Abc_ObjFanin0(abc.Abc_ObjFanin0(latch))
		}
		bo := abc.Abc_ObjFanout0(latch)
		abc.Abc_ObjTransferFanout(bo, replacement)
		abc.Abc_NtkDeleteObj_rec(bo, 0)
		removed++
	}
	return removed
}

func latchInitCode(latch *Abc_Obj_t) int32 {
	if abc.Abc_LatchIsInit0(latch) != 0 {
		return 1
	}
	if abc.Abc_LatchIsInit1(latch) != 0 {
		return 2
	}
	if abc.Abc_LatchIsInitDc(latch) != 0 {
		return 3
	}
	return 0
}

func Abc_NtkReplaceAutonomousLogic(network *Abc_Ntk_t) int32 {
	var roots []*Abc_Obj_t
	seen := make(map[*Abc_Obj_t]bool)
	for _, node := range network.Objects() {
		if abc.Abc_NodeIsTravIdCurrent(node) == 0 {
			continue
		}
		for _, edge := range node.Fanins() {
			fanin := edge.Object
			if abc.Abc_NodeIsTravIdCurrent(fanin) != 0 || seen[fanin] {
				continue
			}
			isConstant := Abc_NtkCheckConstant_rec(fanin) != -1
			if abc.Abc_ObjIsBo(fanin) != 0 {
				boFanin := abc.Abc_ObjFanin0(fanin)
				if abc.Abc_ObjFaninNum(boFanin) > 0 {
					isConstant = isConstant || Abc_NtkCheckConstant_rec(abc.Abc_ObjFanin0(boFanin)) != -1
				}
			}
			if isConstant {
				Abc_NtkSetTravId_rec(fanin)
				continue
			}
			if abc.Abc_ObjIsLatch(fanin) != 0 {
				panic("autonomous logic traversal reached a latch")
			}
			seen[fanin] = true
			roots = append(roots, fanin)
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		return abc.Abc_ObjPointerCompare(roots[i], roots[j]) < 0
	})
	for _, node := range roots {
		pi := abc.Abc_NtkCreatePi(network)
		abc.Abc_ObjAssignName(pi, abc.Abc_ObjName(pi), "")
		abc.Abc_NodeSetTravIdCurrent(pi)
		abc.Abc_ObjTransferFanout(node, pi)
	}
	return int32(len(roots))
}

func Abc_NtkCleanupSeq(network *Abc_Ntk_t, latchSweep, autoSweep, verbose int32) int32 {
	if abc.Abc_NtkIsLogic(network) == 0 {
		panic("Abc_NtkCleanupSeq requires logic network")
	}
	abc.Abc_NtkDfsSeq(network)
	removed := Abc_NodeRemoveNonCurrentObjects(network)
	if verbose != 0 {
		fmt.Printf("Cleanup removed %4d dangling objects.\n", removed)
	}
	if latchSweep != 0 {
		removed = Abc_NtkLatchSweep(network)
		if verbose != 0 {
			fmt.Printf("Cleanup removed %4d redundant latches.\n", removed)
		}
	}
	if autoSweep != 0 {
		abc.Abc_NtkDfsSeqReverse(network)
		added := Abc_NtkReplaceAutonomousLogic(network)
		if verbose != 0 {
			fmt.Printf("Cleanup added   %4d additional PIs.\n", added)
		}
		removed = Abc_NodeRemoveNonCurrentObjects(network)
		if verbose != 0 {
			fmt.Printf("Cleanup removed %4d autonomous objects.\n", removed)
		}
	}
	if abc.Abc_NtkCheck(network) == 0 {
		fmt.Println("Abc_NtkCleanupSeq: The network check has failed.")
	}
	return 1
}

func Abc_NtkSweepBufsInvs(network *Abc_Ntk_t, verbose int32) int32 {
	if abc.Abc_NtkIsLogic(network) == 0 {
		panic("Abc_NtkSweepBufsInvs requires logic network")
	}
	if abc.Abc_NtkToAig(network) == 0 {
		fmt.Fprintln(os.Stdout, "Converting to SOP has failed.")
		return 1
	}
	manager, ok := network.FunctionManager.(*hop.Hop_Man_t)
	if !ok {
		panic("AIG network has no Hop manager")
	}
	abc.Abc_NtkIncrementTravId(network)
	var changes, removed int32 = 1, 0
	for changes != 0 {
		changes = 0
		for _, object := range network.Objects() {
			for pin, edge := range object.Fanins() {
				fanin := edge.Object
				if abc.Abc_NodeIsTravIdCurrent(fanin) != 0 || abc.Abc_ObjIsNode(fanin) == 0 || abc.Abc_ObjFaninNum(fanin) != 1 {
					continue
				}
				if abc.Abc_ObjIsCo(object) != 0 && abc.Abc_NodeIsInv(fanin) != 0 {
					continue
				}
				changes++
				removed++
				if abc.Abc_NodeIsInv(fanin) != 0 {
					old, ok := object.Data.(*hop.Hop_Obj_t)
					if !ok {
						panic("node has no Hop function")
					}
					object.Data = hop.Hop_Compose(manager, old, hop.Hop_Not(hop.Hop_IthVar(manager, int32(pin))), int32(pin))
				}
				driver := abc.Abc_ObjFanin0(fanin)
				abc.Abc_ObjPatchFanin(object, fanin, driver)
				if abc.Abc_ObjFanoutNum(fanin) == 0 {
					abc.Abc_NtkDeleteObj(fanin)
				}
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Removed %d single input nodes.\n", removed)
	}
	return removed
}
