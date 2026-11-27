package abci

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/dsd"
	"github.com/glycerine/abcgo/bdd/extrab"
)

// abcDsd.c: decomposes the network using disjoint-support decomposition.
//
// The DSD marks (Dsd_NodeSetMark) hold Abc_Obj_t pointers in C; here they
// hold object ids of the network being built.

// Abc_NtkDsdGlobal derives the DSD network. Takes the strashed network
// (pNtk), derives global BDDs for the combinational outputs of this
// network, and decomposes these BDDs using disjoint support decomposition.
// Finally, constructs and return a new network, which is topologically
// equivalent to the decomposition tree. Allocates and frees a new BDD
// manager and a new DSD manager.
func Abc_NtkDsdGlobal(pNtk *abc.Network, fVerbose, fPrint, fShort int32) *abc.Network {
	var dd *core.DdManager
	var pNtkNew *abc.Network
	if abc.Abc_NtkIsStrash(pNtk) == 0 {
		panic("Abc_NtkDsdGlobal: Abc_NtkIsStrash(pNtk)")
	}
	dd = Abc_NtkBuildGlobalBdds(pNtk, 10000000, 1, 1, 0, fVerbose)
	if dd == nil {
		return nil
	}
	if fVerbose != 0 {
		fmt.Printf("Shared BDD size = %6d nodes.\n", core.Cudd_ReadKeys(dd)-core.Cudd_ReadDead(dd))
	}
	// transform the result of mapping into a BDD network
	pNtkNew = Abc_NtkDsdInternal(pNtk, fVerbose, fPrint, fShort)
	extrab.Extra_StopManager(dd)
	if pNtkNew == nil {
		return nil
	}
	// copy EXDC network
	if pNtk.Exdc != nil {
		pNtkNew.Exdc = abc.Abc_NtkDup(pNtk.Exdc)
	}
	if abc.Abc_NtkCheck(pNtkNew) == 0 {
		fmt.Printf("Abc_NtkDsdGlobal: The network check has failed.\n")
		abc.Abc_NtkDelete(pNtkNew)
		return nil
	}
	return pNtkNew
}

// Abc_NtkDsdInternal constructs the decomposed network.
func Abc_NtkDsdInternal(pNtk *abc.Network, fVerbose, fPrint, fShort int32) *abc.Network {
	var pManDsd *dsd.Dsd_Manager_t
	var pNtkNew *abc.Network
	var dd *core.DdManager

	// complement the global functions
	nCos := int(abc.Abc_NtkCoNum(pNtk))
	vFuncsGlob := make([]*core.DdNode, 0, nCos)
	for i := 0; i < nCos; i++ {
		pObj := abc.Abc_NtkCo(pNtk, int32(i))
		vFuncsGlob = append(vFuncsGlob, core.Cudd_NotCond(Abc_ObjGlobalBdd(pObj), abc.Abc_ObjFaninC0(pObj) != 0))
	}

	// perform the decomposition
	dd = Abc_NtkGlobalBddMan(pNtk)
	pManDsd = dsd.Dsd_ManagerStart(dd, int(abc.Abc_NtkCiNum(pNtk)), int(fVerbose))
	if pManDsd == nil {
		core.Cudd_Quit(dd)
		return nil
	}
	dsd.Dsd_Decompose(pManDsd, vFuncsGlob, nCos)
	Abc_NtkFreeGlobalBdds(pNtk, 0)

	// start the new network
	pNtkNew = abc.Abc_NtkStartFrom(pNtk, abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD)
	// make sure the new manager has enough inputs
	core.Cudd_bddIthVar(pNtkNew.FunctionManager.(*core.DdManager), len(dd.Vars)-1)
	// put the results into the new network (save new CO drivers in old CO drivers)
	Abc_NtkDsdConstruct(pManDsd, pNtk, pNtkNew)
	// finalize the new network
	abc.Abc_NtkFinalize(pNtk, pNtkNew)
	// fix the problem with complemented and duplicated CO edges
	abc.Abc_NtkLogicMakeSimpleCos(pNtkNew, 0)
	if fPrint != 0 {
		ppNamesCi := abc.Abc_NtkCollectCioNames(pNtk, 0)
		ppNamesCo := abc.Abc_NtkCollectCioNames(pNtk, 1)
		if fVerbose != 0 {
			dsd.Dsd_TreePrint(os.Stdout, pManDsd, ppNamesCi, ppNamesCo, int(fShort), -1)
		} else {
			dsd.Dsd_TreePrint2(os.Stdout, pManDsd, ppNamesCi, ppNamesCo, -1)
		}
	}

	// stop the DSD manager
	dsd.Dsd_ManagerStop(pManDsd)
	return pNtkNew
}

// abcDsdSetCopy is pObj->pCopy = pCopy (possibly complemented).
func abcDsdSetCopy(pObj *abc.Object, pCopy abc.Edge) {
	pObj.Copy = pCopy.Object
	pObj.CopyEdge = pCopy
}

// Abc_NtkDsdConstruct constructs the decomposed network.
func Abc_NtkDsdConstruct(pManDsd *dsd.Dsd_Manager_t, pNtk, pNtkNew *abc.Network) {
	var pNodeDsd *dsd.Dsd_Node_t
	var pNodeNew, pDriver *abc.Object

	// save the CI nodes in the DSD nodes
	pNodeNew = abc.Abc_NtkCreateNodeConst1(pNtkNew)
	abcDsdSetCopy(abc.Abc_AigConst1(pNtk).Object, pNodeNew.Signal())
	dsd.Dsd_NodeSetMark(dsd.Dsd_ManagerReadConst1(pManDsd), pNodeNew.ID)
	for i := 0; i < int(abc.Abc_NtkCiNum(pNtk)); i++ {
		pNode := abc.Abc_NtkCi(pNtk, int32(i))
		pNodeDsd = dsd.Dsd_ManagerReadInput(pManDsd, i)
		dsd.Dsd_NodeSetMark(pNodeDsd, abc.Abc_ObjCopyEdge(pNode).Object.ID)
	}

	// collect DSD nodes in DFS order (leaves and const1 are not collected)
	nNodesDsd := 0
	ppNodesDsd := dsd.Dsd_TreeCollectNodesDfs(pManDsd, &nNodesDsd)
	for i := 0; i < nNodesDsd; i++ {
		Abc_NtkDsdConstructNode(pManDsd, ppNodesDsd[i], pNtkNew, nil)
	}

	// set the pointers to the CO drivers
	for i := 0; i < int(abc.Abc_NtkCoNum(pNtk)); i++ {
		pNode := abc.Abc_NtkCo(pNtk, int32(i))
		pDriver = abc.Abc_ObjFanin0(pNode)
		if abc.Abc_ObjIsNode(pDriver) == 0 {
			continue
		}
		if abc.Abc_AigNodeIsAnd(pDriver) == 0 {
			continue
		}
		pNodeDsd = dsd.Dsd_ManagerReadRoot(pManDsd, i)
		pNodeNew = abc.Abc_NtkObj(pNtkNew, int32(dsd.Dsd_NodeReadMark(dsd.Dsd_Regular(pNodeDsd))))
		abcDsdSetCopy(pDriver, abc.Abc_ObjNotCond(pNodeNew, abcBoolInt32(dsd.Dsd_IsComplement(pNodeDsd))))
	}
}

func abcBoolInt32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// Abc_NtkDsdConstructNode creates the node of the new network for one DSD
// node, whose fanins carry the new objects in their marks.
func Abc_NtkDsdConstructNode(pManDsd *dsd.Dsd_Manager_t, pNodeDsd *dsd.Dsd_Node_t, pNtkNew *abc.Network, pCounters []int) *abc.Object {
	ddDsd := dsd.Dsd_ManagerReadDd(pManDsd)
	ddNew := pNtkNew.FunctionManager.(*core.DdManager)
	var pFaninDsd *dsd.Dsd_Node_t
	var pNodeNew, pFanin *abc.Object
	var bLocal, bTemp, bVar *core.DdNode

	// create the new node
	pNodeNew = abc.Abc_NtkCreateNode(pNtkNew)
	// add the fanins
	Type := dsd.Dsd_NodeReadType(pNodeDsd)
	nDecs := dsd.Dsd_NodeReadDecsNum(pNodeDsd)
	if nDecs <= 1 {
		panic("Abc_NtkDsdConstructNode: nDecs > 1")
	}
	for i := 0; i < nDecs; i++ {
		pFaninDsd = dsd.Dsd_NodeReadDec(pNodeDsd, i)
		pFanin = abc.Abc_NtkObj(pNtkNew, int32(dsd.Dsd_NodeReadMark(dsd.Dsd_Regular(pFaninDsd))))
		abc.Abc_ObjAddFanin(pNodeNew, pFanin)
		if !(Type == dsd.DSD_NODE_OR || !dsd.Dsd_IsComplement(pFaninDsd)) {
			panic("Abc_NtkDsdConstructNode: Type == DSD_NODE_OR || !Dsd_IsComplement(pFaninDsd)")
		}
	}

	// create the local function depending on the type of the node
	switch Type {
	case dsd.DSD_NODE_CONST1:
		bLocal = ddNew.One
		core.Cudd_Ref(bLocal)
	case dsd.DSD_NODE_OR:
		bLocal = core.Cudd_Not(ddNew.One)
		core.Cudd_Ref(bLocal)
		for i := 0; i < nDecs; i++ {
			pFaninDsd = dsd.Dsd_NodeReadDec(pNodeDsd, i)
			bVar = core.Cudd_NotCond(ddNew.Vars[i], dsd.Dsd_IsComplement(pFaninDsd))
			bTemp = bLocal
			bLocal = core.Cudd_bddOr(ddNew, bTemp, bVar)
			core.Cudd_Ref(bLocal)
			core.Cudd_RecursiveDeref(ddNew, bTemp)
		}
	case dsd.DSD_NODE_EXOR:
		bLocal = core.Cudd_Not(ddNew.One)
		core.Cudd_Ref(bLocal)
		for i := 0; i < nDecs; i++ {
			bTemp = bLocal
			bLocal = core.Cudd_bddXor(ddNew, bTemp, ddNew.Vars[i])
			core.Cudd_Ref(bLocal)
			core.Cudd_RecursiveDeref(ddNew, bTemp)
		}
	case dsd.DSD_NODE_PRIME:
		if pCounters != nil {
			if nDecs < 10 {
				pCounters[nDecs]++
			} else {
				pCounters[10]++
			}
		}
		bLocal = dsd.Dsd_TreeGetPrimeFunction(ddDsd, pNodeDsd)
		core.Cudd_Ref(bLocal)
		bTemp = bLocal
		bLocal = extrab.Extra_TransferLevelByLevel(ddDsd, ddNew, bTemp)
		core.Cudd_Ref(bLocal)
		core.Cudd_RecursiveDeref(ddDsd, bTemp)
		// bLocal is now in the new BDD manager
	default:
		panic("Abc_NtkDsdConstructNode: unknown DSD node type")
	}
	pNodeNew.Data = bLocal
	dsd.Dsd_NodeSetMark(pNodeDsd, pNodeNew.ID)
	return pNodeNew
}

// Abc_NtkDsdLocal recursively decomposes internal nodes.
func Abc_NtkDsdLocal(pNtk *abc.Network, fVerbose, fRecursive int32) int32 {
	var pManDsd *dsd.Dsd_Manager_t
	dd := pNtk.FunctionManager.(*core.DdManager)
	pCounters := make([]int, 11)

	if abc.Abc_NtkIsBddLogic(pNtk) == 0 {
		panic("Abc_NtkDsdLocal: Abc_NtkIsBddLogic(pNtk)")
	}

	// make the network minimum base
	abc.Abc_NtkMinimumBase(pNtk)

	// start the DSD manager
	pManDsd = dsd.Dsd_ManagerStart(dd, len(dd.Vars), 0)

	// collect nodes for decomposition
	vNodes := Abc_NtkCollectNodesForDsd(pNtk)
	for i := 0; i < len(vNodes); i++ {
		Abc_NodeDecompDsdAndMux(vNodes[i], &vNodes, pManDsd, fRecursive != 0, pCounters)
	}

	if fVerbose != 0 {
		fmt.Printf("Number of non-decomposable functions:\n")
		for i := 3; i < 10; i++ {
			fmt.Printf("Inputs = %d.  Functions = %6d.\n", i, pCounters[i])
		}
		fmt.Printf("Inputs > %d.  Functions = %6d.\n", 9, pCounters[10])
	}

	// stop the DSD manager
	dsd.Dsd_ManagerStop(pManDsd)

	// make sure everything is okay
	if abc.Abc_NtkCheck(pNtk) == 0 {
		fmt.Printf("Abc_NtkDsdRecursive: The network check has failed.\n")
		return 0
	}
	return 1
}

// Abc_NtkCollectNodesForDsd collects the nodes that may need decomposition.
func Abc_NtkCollectNodesForDsd(pNtk *abc.Network) []*abc.Object {
	vNodes := make([]*abc.Object, 0, 100)
	for i := 0; i < int(abc.Abc_NtkObjNumMax(pNtk)); i++ {
		pNode := abc.Abc_NtkObj(pNtk, int32(i))
		if pNode == nil || abc.Abc_ObjIsNode(pNode) == 0 {
			continue
		}
		if Abc_NodeIsForDsd(pNode) != 0 {
			vNodes = append(vNodes, pNode)
		}
	}
	return vNodes
}

// Abc_NodeDecompDsdAndMux performs decomposition of one node.
func Abc_NodeDecompDsdAndMux(pNode *abc.Object, vNodes *[]*abc.Object, pManDsd *dsd.Dsd_Manager_t, fRecursive bool, pCounters []int) {
	dd := pNode.Network().FunctionManager.(*core.DdManager)
	var pRoot, pFanin, pNode1, pNode2, pNodeC *abc.Object
	var pNodeDsd, pFaninDsd *dsd.Dsd_Node_t

	// try disjoint support decomposition
	pNodeDsd = dsd.Dsd_DecomposeOne(pManDsd, pNode.Data.(*core.DdNode))
	fCompl := dsd.Dsd_IsComplement(pNodeDsd)
	pNodeDsd = dsd.Dsd_Regular(pNodeDsd)

	// determine what decomposition to use
	if !fRecursive || dsd.Dsd_NodeReadDecsNum(pNodeDsd) != int(abc.Abc_ObjFaninNum(pNode)) {
		// perform DSD

		// set the inputs
		for i := 0; i < int(abc.Abc_ObjFaninNum(pNode)); i++ {
			pFanin = abc.Abc_ObjFanin(pNode, int32(i))
			pFaninDsd = dsd.Dsd_ManagerReadInput(pManDsd, i)
			dsd.Dsd_NodeSetMark(pFaninDsd, pFanin.ID)
		}

		// construct the intermediate nodes
		nNodesDsd := 0
		ppNodesDsd := dsd.Dsd_TreeCollectNodesDfsOne(pManDsd, pNodeDsd, &nNodesDsd)
		for i := 0; i < nNodesDsd; i++ {
			pRoot = Abc_NtkDsdConstructNode(pManDsd, ppNodesDsd[i], pNode.Network(), pCounters)
			if Abc_NodeIsForDsd(pRoot) != 0 && fRecursive {
				*vNodes = append(*vNodes, pRoot)
			}
		}
		if pRoot == nil {
			panic("Abc_NodeDecompDsdAndMux: pRoot")
		}

		// remove the current fanins
		abc.Abc_ObjRemoveFanins(pNode)
		// add fanin to the root
		abc.Abc_ObjAddFanin(pNode, pRoot)
		// update the function to be that of buffer
		core.Cudd_RecursiveDeref(dd, pNode.Data.(*core.DdNode))
		pNode.Data = core.Cudd_NotCond(dd.Vars[0], fCompl)
		core.Cudd_Ref(pNode.Data.(*core.DdNode))
	} else { // perform MUX-decomposition
		// get the cofactoring variable
		iVar := Abc_NodeFindMuxVar(dd, pNode.Data.(*core.DdNode), int(abc.Abc_ObjFaninNum(pNode)))
		pNodeC = abc.Abc_ObjFanin(pNode, int32(iVar))

		// get the negative cofactor
		pNode1 = abc.Abc_NtkCloneObj(pNode)
		pNode1.Data = core.Cudd_Cofactor(dd, pNode.Data.(*core.DdNode), core.Cudd_Not(dd.Vars[iVar]))
		core.Cudd_Ref(pNode1.Data.(*core.DdNode))
		abc.Abc_NodeMinimumBase(pNode1)
		if Abc_NodeIsForDsd(pNode1) != 0 {
			*vNodes = append(*vNodes, pNode1)
		}

		// get the positive cofactor
		pNode2 = abc.Abc_NtkCloneObj(pNode)
		pNode2.Data = core.Cudd_Cofactor(dd, pNode.Data.(*core.DdNode), dd.Vars[iVar])
		core.Cudd_Ref(pNode2.Data.(*core.DdNode))
		abc.Abc_NodeMinimumBase(pNode2)
		if Abc_NodeIsForDsd(pNode2) != 0 {
			*vNodes = append(*vNodes, pNode2)
		}

		// remove the current fanins
		abc.Abc_ObjRemoveFanins(pNode)
		// add new fanins
		abc.Abc_ObjAddFanin(pNode, pNodeC)
		abc.Abc_ObjAddFanin(pNode, pNode2)
		abc.Abc_ObjAddFanin(pNode, pNode1)
		// update the function to be that of MUX
		core.Cudd_RecursiveDeref(dd, pNode.Data.(*core.DdNode))
		pNode.Data = core.Cudd_bddIte(dd, dd.Vars[0], dd.Vars[1], dd.Vars[2])
		core.Cudd_Ref(pNode.Data.(*core.DdNode))
	}
}

// Abc_NodeIsForDsd checks if the node should be decomposed by DSD.
func Abc_NodeIsForDsd(pNode *abc.Object) int32 {
	if abc.Abc_ObjIsNode(pNode) == 0 {
		panic("Abc_NodeIsForDsd: Abc_ObjIsNode(pNode)")
	}
	if abc.Abc_ObjFaninNum(pNode) > 2 {
		return 1
	}
	return 0
}

// Abc_NodeFindMuxVar determines a cofactoring variable.
func Abc_NodeFindMuxVar(dd *core.DdManager, bFunc *core.DdNode, nVars int) int {
	var bVar, bCof0, bCof1 *core.DdNode
	SuppSumMin := 1000000
	var nSSD, nSSQ int

	iVar := -1
	for i := 0; i < nVars; i++ {
		bVar = dd.Vars[i]

		bCof0 = core.Cudd_Cofactor(dd, bFunc, core.Cudd_Not(bVar))
		core.Cudd_Ref(bCof0)
		bCof1 = core.Cudd_Cofactor(dd, bFunc, bVar)
		core.Cudd_Ref(bCof1)

		nSSD = core.Cudd_SupportSize(dd, bCof0)
		nSSQ = core.Cudd_SupportSize(dd, bCof1)

		core.Cudd_RecursiveDeref(dd, bCof0)
		core.Cudd_RecursiveDeref(dd, bCof1)

		if SuppSumMin > nSSD+nSSQ {
			SuppSumMin = nSSD + nSSQ
			iVar = i
		}
	}
	return iVar
}

// Extra_bddComputeSum computes the OR of the cubes.
func Extra_bddComputeSum(dd *core.DdManager, pbCubes []*core.DdNode) *core.DdNode {
	var bRes, bTemp *core.DdNode
	bRes = core.Cudd_Not(dd.One)
	core.Cudd_Ref(bRes)
	for i := 0; i < len(pbCubes); i++ {
		bTemp = bRes
		bRes = core.Cudd_bddOr(dd, bTemp, pbCubes[i])
		core.Cudd_Ref(bRes)
		core.Cudd_RecursiveDeref(dd, bTemp)
	}
	core.Cudd_Deref(bRes)
	return bRes
}

// abcTransferBDD transfers a BDD level by level (used by Abc_NtkSparsify).
func abcTransferBDD(source, destination *core.DdManager, function *core.DdNode) *core.DdNode {
	memo := map[*core.DdNode]*core.DdNode{}
	var transfer func(*core.DdNode) *core.DdNode
	transfer = func(edge *core.DdNode) *core.DdNode {
		if result := memo[edge]; result != nil {
			return result
		}
		regular := core.Cudd_Regular(edge)
		if core.Cudd_IsConstant(regular) {
			result := core.Cudd_NotCond(core.Cudd_ReadOne(destination), edge != regular)
			memo[edge] = result
			return result
		}
		low := core.Cudd_NotCond(regular.E, edge != regular)
		high := core.Cudd_NotCond(regular.T, edge != regular)
		level := source.Perm[regular.Index]
		index := destination.InvPerm[level]
		result := core.Cudd_bddIte(destination, core.Cudd_bddIthVar(destination, index), transfer(high), transfer(low))
		memo[edge] = result
		return result
	}
	return transfer(function)
}

// Abc_NtkSparsify exposes the source API. Percentages outside [1,100] are
// retain one minterm for zero-sized on/off sets, matching ABC behavior.
func Abc_NtkSparsify(n *abc.Network, percentage, verbose int32) *abc.Network {
	if n == nil || abc.Abc_NtkIsComb(n) == 0 || abc.Abc_NtkIsBddLogic(n) == 0 {
		return nil
	}
	return Abc_NtkSparsifyInternal(n, percentage, verbose)
}

func Abc_NtkSparsifyInternal(n *abc.Network, percentage, verbose int32) *abc.Network {
	if n == nil {
		return nil
	}
	sourceDD, ok := n.FunctionManager.(*core.DdManager)
	if !ok {
		return nil
	}
	out := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	out.Name, out.Spec = n.Name, n.Spec
	destDD := out.FunctionManager.(*core.DdManager)
	oldToNew := map[*abc.Object]*abc.Object{}
	oldToNew[n.One().Object] = out.One().Object
	for _, old := range n.Inputs() {
		pi := out.NewObject(abc.ObjectPI)
		pi.SetName(old.Name())
		oldToNew[old] = pi
	}
	var roots []*abc.Object
	for _, co := range n.Outputs() {
		for _, edge := range co.Fanins() {
			roots = append(roots, edge.Object)
		}
	}
	order, err := n.Topological(roots)
	if err != nil {
		return nil
	}
	for _, old := range order {
		if old.IsCI() || old.Kind == abc.ObjectConst1 {
			continue
		}
		if old.Kind != abc.ObjectNode {
			continue
		}
		copy := out.NewObject(abc.ObjectNode)
		for _, edge := range old.Fanins() {
			fanin := oldToNew[edge.Object]
			if fanin == nil {
				return nil
			}
			if err := out.AddFanin(copy, abc.Edge{Object: fanin, Complement: edge.Complement}); err != nil {
				return nil
			}
		}
		function, ok := old.Data.(*core.DdNode)
		if !ok {
			return nil
		}
		copy.Data = abcTransferBDD(sourceDD, destDD, function)
		oldToNew[old] = copy
	}
	for _, co := range n.Outputs() {
		if co.FaninNum() != 1 {
			return nil
		}
		edge := co.FaninAt(int32(0))
		name := co.Name()
		if edge.Object.IsCI() {
			driver := oldToNew[edge.Object]
			for c := 0; c < 2; c++ {
				po := out.NewObject(abc.ObjectPO)
				suffix := "_on"
				if c == 1 {
					suffix = "_off"
				}
				po.SetName(name + suffix)
				if err := out.AddFanin(po, abc.Edge{Object: driver, Complement: edge.Complement != (c == 1)}); err != nil {
					return nil
				}
			}
			continue
		}
		if edge.Object.Kind == abc.ObjectConst1 || edge.Object.FaninNum() == 0 {
			value, evalErr := n.Evaluate(edge, map[*abc.Object]bool{})
			if evalErr != nil {
				return nil
			}
			for c := 0; c < 2; c++ {
				complement := !(value != (c == 1))
				po := out.NewObject(abc.ObjectPO)
				suffix := "_on"
				if c == 1 {
					suffix = "_off"
				}
				po.SetName(name + suffix)
				if err := out.AddFanin(po, abc.Edge{Object: out.One().Object, Complement: complement}); err != nil {
					return nil
				}
			}
			continue
		}
		base := oldToNew[edge.Object]
		if base == nil {
			return nil
		}
		original, ok := base.Data.(*core.DdNode)
		if !ok {
			return nil
		}
		original = core.Cudd_NotCond(original, edge.Complement)
		for c := 0; c < 2; c++ {
			core.Cudd_Srandom(0)
			copy := out.NewObject(abc.ObjectNode)
			for _, fanin := range base.Fanins() {
				if err := out.AddFanin(copy, fanin); err != nil {
					return nil
				}
			}
			function := core.Cudd_NotCond(original, c == 1)
			copy.Data = Abc_NtkSparsifyInternalOne(destDD, function, copy.FaninNum(), percentage)
			if copy.Data == nil {
				return nil
			}
			po := out.NewObject(abc.ObjectPO)
			suffix := "_on"
			if c == 1 {
				suffix = "_off"
			}
			po.SetName(name + suffix)
			if err := out.AddFanin(po, abc.Edge{Object: copy}); err != nil {
				return nil
			}
		}
	}
	_ = verbose
	if abc.Abc_NtkCheck(out) == 0 {
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}

func Abc_NtkSparsifyInternalOne(dd *core.DdManager, function *core.DdNode, fanins int, percentage int32) *core.DdNode {
	space := int(core.Cudd_CountMinterm(dd, function, fanins))
	count := max(1, int(float64(percentage)*0.01*float64(space)))
	cubes := abcPickSourceMinterms(dd, function, fanins, count)
	if cubes == nil {
		return nil
	}
	result := core.Cudd_Not(core.Cudd_ReadOne(dd))
	for _, cube := range cubes {
		result = core.Cudd_bddOr(dd, result, cube)
	}
	return result
}

// abcPickSourceMinterms follows CUDD's recursive cofactor allocation and
// source ordering, then randomizes don't-cares with CUDD's seeded generator.
func abcPickSourceMinterms(dd *core.DdManager, function *core.DdNode, nvars, count int) []*core.DdNode {
	if count < 0 || nvars < 0 || float64(count) > core.Cudd_CountMinterm(dd, function, nvars) {
		return nil
	}
	strings := make([][]byte, count)
	for i := range strings {
		strings[i] = make([]byte, len(dd.Vars))
		for j := range strings[i] {
			strings[i][j] = '2'
		}
	}
	var pick func(*core.DdNode, [][]byte) bool
	pick = func(f *core.DdNode, rows [][]byte) bool {
		if len(rows) == 0 || f == core.Cudd_Not(core.Cudd_ReadOne(dd)) || f == core.Cudd_ReadOne(dd) {
			return true
		}
		node := core.Cudd_Regular(f)
		t, e := node.T, node.E
		if core.Cudd_IsComplement(f) {
			t, e = core.Cudd_Not(t), core.Cudd_Not(e)
		}
		oneCount := int(float64(len(rows))*core.Cudd_CountMinterm(dd, t, nvars)/(core.Cudd_CountMinterm(dd, t, nvars)+core.Cudd_CountMinterm(dd, e, nvars)) + 0.5)
		for i := 0; i < oneCount; i++ {
			rows[i][node.Index] = '1'
		}
		for i := oneCount; i < len(rows); i++ {
			rows[i][node.Index] = '0'
		}
		return pick(t, rows[:oneCount]) && pick(e, rows[oneCount:])
	}
	if !pick(function, strings) {
		return nil
	}
	selected := make([]*core.DdNode, count)
	saved := ""
	savePoint := -1
	saveFlag := false
	for i, row := range strings {
		isSame := false
		if !saveFlag {
			for j := i + 1; j < count; j++ {
				if string(row) == string(strings[j]) {
					savePoint = i
					saved = string(row)
					saveFlag = true
					break
				}
			}
		} else if string(row) == saved {
			isSame = true
		} else {
			saveFlag = false
			for j := i + 1; j < count; j++ {
				if string(row) == string(strings[j]) {
					savePoint = i
					saved = string(row)
					saveFlag = true
					break
				}
			}
		}
		for j := 0; j < nvars; j++ {
			index := int(core.Cudd_bddIthVar(dd, j).Index)
			if row[index] == '2' {
				if core.Cudd_Random()&0x20 != 0 {
					row[index] = '1'
				} else {
					row[index] = '0'
				}
			}
		}
		for isSame {
			isSame = false
			for j := savePoint; j < i; j++ {
				if string(row) == string(strings[j]) {
					isSame = true
					break
				}
			}
			if isSame {
				copy(row, []byte(saved))
				for j := 0; j < nvars; j++ {
					index := int(core.Cudd_bddIthVar(dd, j).Index)
					if row[index] == '2' {
						if core.Cudd_Random()&0x20 != 0 {
							row[index] = '1'
						} else {
							row[index] = '0'
						}
					}
				}
			}
		}
		cube := core.Cudd_ReadOne(dd)
		for j := 0; j < nvars; j++ {
			v := core.Cudd_bddIthVar(dd, j)
			if row[v.Index] == '0' {
				v = core.Cudd_Not(v)
			}
			cube = core.Cudd_bddAnd(dd, cube, v)
		}
		if !core.Cudd_bddLeq(dd, cube, function) {
			return nil
		}
		selected[i] = cube
	}
	return selected
}
