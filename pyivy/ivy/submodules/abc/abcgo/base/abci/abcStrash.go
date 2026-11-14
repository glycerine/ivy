package abci

import (
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func strashAig(n *abc.Network) *abc.Abc_Aig_t {
	m, _ := n.FunctionManager.(*abc.Abc_Aig_t)
	return m
}

func strashCopyChild(o *abc.Object, index int32) abc.Edge {
	fanin := abc.Abc_ObjFanin(o, index)
	edge := abc.Abc_ObjCopyEdge(fanin)
	if abc.Abc_ObjFaninC(o, index) != 0 {
		edge = edge.Not()
	}
	return edge
}

func strashNodeCopy(n *abc.Network, o *abc.Object) {
	abc.Abc_ObjSetCopyData(o, abc.Abc_AigAnd(strashAig(n), strashCopyChild(o, 0), strashCopyChild(o, 1)))
}

func abcStrashRandBit() uint32 {
	cascadeRandState.Lock()
	defer cascadeRandState.Unlock()
	return cascadeRandState.rng.next() & 1
}

func abcHopChildCopy(node *Hop_Obj_t, index int) abc.Edge {
	var child *Hop_Obj_t
	var complemented int32
	if index == 0 {
		child, complemented = hop.Hop_ObjFanin0(node), hop.Hop_ObjFaninC0(node)
	} else {
		child, complemented = hop.Hop_ObjFanin1(node), hop.Hop_ObjFaninC1(node)
	}
	copy, ok := hop.Hop_ObjData(child).(abc.Edge)
	assert(ok, "Hop child has ABC edge copy")
	return abc.Abc_ObjNotCond(copy, complemented)
}

// Abc_NtkRestrash reapplies structural hashing to an existing AIG.
func Abc_NtkRestrash(pNtk *Abc_Ntk_t, fCleanup int32) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	if abc.Abc_NtkGetChoiceNum(pNtk) != 0 {
		fmt.Println("Warning: The choice nodes in the original AIG are removed by strashing.")
	}
	pNtkAig := abc.Abc_NtkStartFrom(pNtk, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	for _, node := range abc.Abc_NtkDfs(pNtk, 0) {
		strashNodeCopy(pNtkAig, node)
	}
	abc.Abc_NtkFinalize(pNtk, pNtkAig)
	if fCleanup != 0 {
		abc.Abc_AigCleanup(strashAig(pNtkAig))
	}
	if pNtk.Exdc != nil {
		pNtkAig.Exdc = abc.Abc_NtkDup(pNtk.Exdc)
	}
	if abc.Abc_NtkCheck(pNtkAig) == 0 {
		fmt.Println("Abc_NtkStrash: The network check has failed.")
		abc.Abc_NtkDelete(pNtkAig)
		return nil
	}
	return pNtkAig
}

// Abc_NtkRestrashRandom_rec performs the source's randomized DFS ordering.
func Abc_NtkRestrashRandom_rec(pNtk *Abc_Ntk_t, pObj *Abc_Obj_t) {
	if abc.Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return
	}
	abc.Abc_NodeSetTravIdCurrent(pObj)
	if abc.Abc_ObjIsNode(pObj) == 0 {
		return
	}
	if abcStrashRandBit() != 0 {
		Abc_NtkRestrashRandom_rec(pNtk, abc.Abc_ObjFanin0(pObj))
		Abc_NtkRestrashRandom_rec(pNtk, abc.Abc_ObjFanin1(pObj))
	} else {
		Abc_NtkRestrashRandom_rec(pNtk, abc.Abc_ObjFanin1(pObj))
		Abc_NtkRestrashRandom_rec(pNtk, abc.Abc_ObjFanin0(pObj))
	}
	strashNodeCopy(pNtk, pObj)
}

func Abc_NtkRestrashRandom(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	if abc.Abc_NtkGetChoiceNum(pNtk) != 0 {
		fmt.Println("Warning: The choice nodes in the original AIG are removed by strashing.")
	}
	pNtkAig := abc.Abc_NtkStartFrom(pNtk, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	abc.Abc_NtkIncrementTravId(pNtk)
	for _, co := range pNtk.Outputs() {
		if len(co.Fanins()) != 0 {
			Abc_NtkRestrashRandom_rec(pNtkAig, abc.Abc_ObjFanin0(co))
		}
	}
	abc.Abc_NtkFinalize(pNtk, pNtkAig)
	if pNtk.Exdc != nil {
		pNtkAig.Exdc = abc.Abc_NtkDup(pNtk.Exdc)
	}
	if abc.Abc_NtkCheck(pNtkAig) == 0 {
		fmt.Println("Abc_NtkStrash: The network check has failed.")
		abc.Abc_NtkDelete(pNtkAig)
		return nil
	}
	return pNtkAig
}

func Abc_NtkRestrashZero(pNtk *Abc_Ntk_t, fCleanup int32) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	if abc.Abc_NtkGetChoiceNum(pNtk) != 0 {
		fmt.Println("Warning: The choice nodes in the original AIG are removed by strashing.")
	}
	pNtkAig := abc.Abc_NtkStartFrom(pNtk, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	counter := 0
	for _, latch := range pNtk.ObjectsOfKind(abc.ObjectLatch) {
		if abc.Abc_LatchIsInitDc(latch) != 0 {
			counter++
		} else if abc.Abc_LatchIsInit1(latch) != 0 && len(latch.Fanouts()) != 0 {
			bo := latch.Fanouts()[0]
			abc.Abc_ObjSetCopyData(bo, abc.Abc_ObjNot(abc.Abc_ObjCopyEdge(bo)))
		}
	}
	if counter != 0 {
		fmt.Printf("Converting %d flops from don't-care to zero initial value.\n", counter)
	}
	for _, obj := range pNtk.Objects() {
		if obj.Kind == abc.ObjectNode {
			strashNodeCopy(pNtkAig, obj)
		}
	}
	abc.Abc_NtkFinalize(pNtk, pNtkAig)
	for _, latch := range pNtkAig.ObjectsOfKind(abc.ObjectLatch) {
		if abc.Abc_LatchIsInit1(latch) != 0 && len(latch.Fanins()) != 0 {
			bi := latch.Fanins()[0].Object
			abc.Abc_ObjXorFaninC(bi, 0)
			if len(bi.Fanouts()) != 0 {
				bo := bi.Fanouts()[0]
				for _, po := range bo.Fanouts() {
					if po.Kind == abc.ObjectPO {
						bo.SetName(abc.Abc_ObjName(bo) + "_inv")
						break
					}
				}
			}
		}
		abc.Abc_LatchSetInit0(latch)
	}
	if fCleanup != 0 {
		if nNodes := abc.Abc_AigCleanup(strashAig(pNtkAig)); nNodes != 0 {
			fmt.Printf("Abc_NtkRestrash(): AIG cleanup removed %d nodes (this is a bug).\n", nNodes)
		}
	}
	if pNtk.Exdc != nil {
		pNtkAig.Exdc = abc.Abc_NtkDup(pNtk.Exdc)
	}
	if pNtk.NameIDs != nil {
		abc.Abc_NtkTransferNameIds(pNtk, pNtkAig)
		abc.Abc_NtkUpdateNameIds(pNtkAig)
	}
	if abc.Abc_NtkCheck(pNtkAig) == 0 {
		fmt.Println("Abc_NtkStrash: The network check has failed.")
		abc.Abc_NtkDelete(pNtkAig)
		return nil
	}
	return pNtkAig
}

func Abc_NtkStrash(pNtk *Abc_Ntk_t, fAllNodes, fCleanup, fRecord int32) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsLogic(pNtk) != 0 || abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsLogic(pNtk) || Abc_NtkIsStrash(pNtk)")
	if abc.Abc_NtkIsStrash(pNtk) != 0 {
		return Abc_NtkRestrash(pNtk, fCleanup)
	}
	if abc.Abc_NtkToAig(pNtk) == 0 {
		fmt.Println("Converting to AIGs has failed.")
		return nil
	}
	pNtkAig := abc.Abc_NtkStartFrom(pNtk, abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG)
	Abc_NtkStrashPerform(pNtk, pNtkAig, fAllNodes, fRecord)
	abc.Abc_NtkFinalize(pNtk, pNtkAig)
	if pNtk.NameIDs != nil {
		abc.Abc_NtkTransferNameIds(pNtk, pNtkAig)
	}
	if fCleanup != 0 {
		abc.Abc_AigCleanup(strashAig(pNtkAig))
	}
	if pNtk.Exdc != nil {
		pNtkAig.Exdc = Abc_NtkStrash(pNtk.Exdc, fAllNodes, fCleanup, fRecord)
	}
	if abc.Abc_NtkCheck(pNtkAig) == 0 {
		fmt.Println("Abc_NtkStrash: The network check has failed.")
		abc.Abc_NtkDelete(pNtkAig)
		return nil
	}
	return pNtkAig
}

func Abc_NtkAppend(pNtk1, pNtk2 *Abc_Ntk_t, fAddPos int32) int32 {
	assert(abc.Abc_NtkIsStrash(pNtk1) != 0, "Abc_NtkIsStrash(pNtk1)")
	assert(abc.Abc_NtkIsLogic(pNtk2) != 0 || abc.Abc_NtkIsStrash(pNtk2) != 0, "Abc_NtkIsLogic(pNtk2) || Abc_NtkIsStrash(pNtk2)")
	if abc.Abc_NtkIsLogic(pNtk2) != 0 && abc.Abc_NtkToAig(pNtk2) == 0 {
		fmt.Println("Converting to AIGs has failed.")
		return 0
	}
	if abc.Abc_NtkCompareSignals(pNtk1, pNtk2, 1, 1) == 0 {
		fmt.Println("Abc_NtkAppend(): The union of the network PIs is computed (warning).")
	}
	abc.Abc_NtkCleanCopy(pNtk2)
	if abc.Abc_NtkIsStrash(pNtk2) != 0 {
		abc.Abc_ObjSetCopyData(abc.Abc_AigConst1(pNtk2).Object, abc.Abc_AigConst1(pNtk1))
	}
	newCIs := 0
	for _, ci := range pNtk2.Inputs() {
		copy := abc.Abc_NtkFindCi(pNtk1, abc.Abc_ObjName(ci))
		if copy == nil {
			copy = abc.Abc_NtkDupObj(pNtk1, ci, 1)
			newCIs++
		}
		abc.Abc_ObjSetCopyData(ci, copy)
	}
	if newCIs != 0 {
		fmt.Printf("Warning: Procedure Abc_NtkAppend() added %d new CIs.\n", newCIs)
	}
	if abc.Abc_NtkIsLogic(pNtk2) != 0 {
		Abc_NtkStrashPerform(pNtk2, pNtk1, 1, 0)
	} else {
		for _, obj := range pNtk2.Objects() {
			if obj.Kind == abc.ObjectNode {
				strashNodeCopy(pNtk1, obj)
			}
		}
	}
	if fAddPos != 0 {
		for _, po := range pNtk2.ObjectsOfKind(abc.ObjectPO) {
			newPO := abc.Abc_NtkDupObj(pNtk1, po, 0)
			abc.Abc_ObjAddFanin(newPO, strashCopyChild(po, 0))
			abc.Abc_ObjAssignName(newPO, abc.Abc_ObjName(po), "")
		}
	} else {
		for _, co := range pNtk2.Outputs() {
			var old *abc.Object
			for _, candidate := range pNtk1.Objects() {
				if (candidate.Kind == abc.ObjectPO || candidate.Kind == abc.ObjectBI) && abc.Abc_ObjName(candidate) == abc.Abc_ObjName(co) {
					old = candidate
					break
				}
			}
			assert(old != nil, "matching PO/BI name exists")
			oldDriver := abc.Abc_ObjChild0(old)
			newDriver := strashCopyChild(co, 0)
			orDriver := abc.Abc_AigOr(strashAig(pNtk1), oldDriver, newDriver)
			if abc.Abc_ObjRegular(oldDriver) == abc.Abc_ObjRegular(orDriver) {
				continue
			}
			compl := int32(0)
			if abc.Abc_ObjRegular(oldDriver).Phase != abc.Abc_ObjRegular(orDriver).Phase {
				compl = 1
			}
			abc.Abc_ObjPatchFanin(old, abc.Abc_ObjRegular(oldDriver), abc.Abc_ObjNotCond(abc.Abc_ObjRegular(orDriver), compl))
		}
	}
	if abc.Abc_NtkCheck(pNtk1) == 0 {
		fmt.Println("Abc_NtkAppend: The network check has failed.")
		return 0
	}
	return 1
}

func Abc_NtkStrashPerform(pNtkOld, pNtkNew *Abc_Ntk_t, fAllNodes, fRecord int32) {
	assert(abc.Abc_NtkIsLogic(pNtkOld) != 0, "Abc_NtkIsLogic(pNtkOld)")
	assert(abc.Abc_NtkIsStrash(pNtkNew) != 0, "Abc_NtkIsStrash(pNtkNew)")
	for _, node := range abc.Abc_NtkDfsIter(pNtkOld, fAllNodes) {
		if abc.Abc_ObjIsBarBuf(node) != 0 {
			abc.Abc_ObjSetCopyData(node, strashCopyChild(node, 0))
		} else {
			abc.Abc_ObjSetCopyData(node, Abc_NodeStrash(pNtkNew, node, fRecord))
		}
	}
}

func Abc_NodeStrash_rec(pMan *abc.Abc_Aig_t, pObj *Hop_Obj_t) {
	assert(hop.Hop_IsComplement(pObj) == 0, "!Hop_IsComplement(pObj)")
	if hop.Hop_ObjIsNode(pObj) == 0 || hop.Hop_ObjIsMarkA(pObj) != 0 {
		return
	}
	Abc_NodeStrash_rec(pMan, hop.Hop_ObjFanin0(pObj))
	Abc_NodeStrash_rec(pMan, hop.Hop_ObjFanin1(pObj))
	a := abcHopChildCopy(pObj, 0)
	b := abcHopChildCopy(pObj, 1)
	hop.Hop_ObjSetData(pObj, abc.Abc_AigAnd(pMan, a, b))
	assert(hop.Hop_ObjIsMarkA(pObj) == 0, "!Hop_ObjIsMarkA(pObj)")
	hop.Hop_ObjSetMarkA(pObj)
}

func Abc_NodeStrash(pNtkNew *Abc_Ntk_t, pNodeOld *Abc_Obj_t, fRecord int32) abc.Edge {
	assert(abc.Abc_ObjIsNode(pNodeOld) != 0, "Abc_ObjIsNode(pNodeOld)")
	assert(abc.Abc_NtkHasAig(pNodeOld.Network()) != 0 && abc.Abc_NtkIsStrash(pNodeOld.Network()) == 0, "Abc_NtkHasAig && !Abc_NtkIsStrash")
	pMan, ok := pNodeOld.Network().FunctionManager.(*hop.Hop_Man_t)
	assert(ok, "Hop manager exists")
	pRoot, ok := pNodeOld.Data.(*Hop_Obj_t)
	assert(ok, "node function is Hop object")
	if abc.Abc_NodeIsConst(pNodeOld) != 0 || hop.Hop_Regular(pRoot) == hop.Hop_ManConst1(pMan) {
		return abc.Abc_ObjNotCond(abc.Abc_AigConst1(pNtkNew), hop.Hop_IsComplement(pRoot))
	}
	for i, fanin := range pNodeOld.Fanins() {
		hop.Hop_ObjSetData(hop.Hop_IthVar(pMan, int32(i)), abc.Abc_ObjCopyEdge(fanin.Object))
	}
	hop.Hop_ObjSetData(hop.Hop_ManConst1(pMan), abc.Abc_AigConst1(pNtkNew))
	root := hop.Hop_Regular(pRoot)
	Abc_NodeStrash_rec(strashAig(pNtkNew), root)
	hop.Hop_ConeUnmark_rec(root)
	result, ok := hop.Hop_ObjData(root).(abc.Edge)
	assert(ok, "strashed Hop node has ABC edge copy")
	return abc.Abc_ObjNotCond(result, hop.Hop_IsComplement(pRoot))
}

func Abc_NtkTopmost_rec(pNtkNew *Abc_Ntk_t, pNode *Abc_Obj_t, levelCut int32) abc.Edge {
	if edge := abc.Abc_ObjCopyEdge(pNode); edge.Object != nil {
		return edge
	}
	if pNode.Level <= int(levelCut) {
		pi := abc.Abc_NtkCreatePi(pNtkNew)
		abc.Abc_ObjSetCopyData(pNode, pi)
		return pi.Signal()
	}
	Abc_NtkTopmost_rec(pNtkNew, abc.Abc_ObjFanin0(pNode), levelCut)
	Abc_NtkTopmost_rec(pNtkNew, abc.Abc_ObjFanin1(pNode), levelCut)
	copy := abc.Abc_AigAnd(strashAig(pNtkNew), strashCopyChild(pNode, 0), strashCopyChild(pNode, 1))
	abc.Abc_ObjSetCopyData(pNode, copy)
	return copy
}

func Abc_NtkTopmost(pNtk *Abc_Ntk_t, nLevels int32) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	levelCut := abc.Abc_AigLevel(pNtk) - nLevels
	if levelCut < 0 {
		levelCut = 0
	}
	out := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	out.Name = pNtk.Name
	abc.Abc_NtkCleanCopy(pNtk)
	abc.Abc_ObjSetCopyData(abc.Abc_AigConst1(pNtk).Object, abc.Abc_AigConst1(out))
	for _, co := range pNtk.Outputs() {
		if len(co.Fanins()) == 0 {
			continue
		}
		driver := Abc_NtkTopmost_rec(out, abc.Abc_ObjFanin0(co), levelCut)
		copy := abc.Abc_ObjNotCond(driver, abc.Abc_ObjFaninC0(co))
		po := abc.Abc_NtkCreatePo(out)
		abc.Abc_ObjAddFanin(po, copy)
		abc.Abc_ObjSetCopyData(co, po)
	}
	abc.Abc_NtkAddDummyPiNames(out)
	for _, co := range pNtk.Outputs() {
		if co.Copy != nil {
			abc.Abc_ObjAssignName(co.Copy, abc.Abc_ObjName(co), "")
		}
	}
	if abc.Abc_NtkCheck(out) == 0 {
		fmt.Println("Abc_NtkTopmost: The network check has failed.")
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}

func Abc_NtkBottommost_rec(pNtkNew *Abc_Ntk_t, pNode *Abc_Obj_t, levelCut int32) abc.Edge {
	if edge := abc.Abc_ObjCopyEdge(pNode); edge.Object != nil {
		return edge
	}
	Abc_NtkBottommost_rec(pNtkNew, abc.Abc_ObjFanin0(pNode), levelCut)
	Abc_NtkBottommost_rec(pNtkNew, abc.Abc_ObjFanin1(pNode), levelCut)
	if pNode.Level > int(levelCut) {
		return abc.Edge{}
	}
	copy := abc.Abc_AigAnd(strashAig(pNtkNew), strashCopyChild(pNode, 0), strashCopyChild(pNode, 1))
	abc.Abc_ObjSetCopyData(pNode, copy)
	return copy
}

func Abc_NtkBottommost(pNtk *Abc_Ntk_t, nLevels int32) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	assert(nLevels >= 0, "nLevels >= 0")
	out := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	out.Name = pNtk.Name
	abc.Abc_NtkCleanCopy(pNtk)
	abc.Abc_ObjSetCopyData(abc.Abc_AigConst1(pNtk).Object, abc.Abc_AigConst1(out))
	for _, ci := range pNtk.Inputs() {
		abc.Abc_ObjSetCopyData(ci, abc.Abc_NtkCreatePi(out))
	}
	for _, co := range pNtk.Outputs() {
		if len(co.Fanins()) != 0 {
			Abc_NtkBottommost_rec(out, abc.Abc_ObjFanin0(co), nLevels)
		}
	}
	for _, node := range out.ObjectsOfKind(abc.ObjectNode) {
		if abc.Abc_ObjFanoutNum(node) == 0 {
			po := abc.Abc_NtkCreatePo(out)
			abc.Abc_ObjAddFanin(po, node)
		}
	}
	abc.Abc_NtkAddDummyPiNames(out)
	abc.Abc_NtkAddDummyPoNames(out)
	if abc.Abc_NtkCheck(out) == 0 {
		fmt.Println("Abc_NtkBottommost: The network check has failed.")
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}

func Vec_CompareNodeIds(pp1 **Abc_Obj_t, pp2 **Abc_Obj_t) int32 {
	if abc.Abc_ObjRegular(*pp1).ID < abc.Abc_ObjRegular(*pp2).ID {
		return -1
	}
	if abc.Abc_ObjRegular(*pp1).ID > abc.Abc_ObjRegular(*pp2).ID {
		return 1
	}
	return 0
}

func abcCompareObjIDs(a, b *any) int {
	x := abc.Abc_ObjRegular(*a)
	y := abc.Abc_ObjRegular(*b)
	if x.ID < y.ID {
		return -1
	}
	if x.ID > y.ID {
		return 1
	}
	return 0
}

func Abc_NodeGetSuper(pNode *Abc_Obj_t) *Vec_Ptr_t {
	assert(abc.Abc_ObjIsNode(pNode) != 0 && !pNode.Signal().Complement, "node is regular")
	super, front := vec.Vec_PtrAlloc(100), vec.Vec_PtrAlloc(100)
	vec.Vec_PtrPush(front, pNode)
	for i := 0; i < vec.Vec_PtrSize(front); i++ {
		and := vec.Vec_PtrEntry(front, i).(*abc.Object)
		for _, fanin := range and.Fanins() {
			if fanin.Object.Kind == abc.ObjectNode && !fanin.Complement && abc.Abc_ObjFanoutNum(fanin.Object) == 1 {
				vec.Vec_PtrPush(front, fanin.Object)
			} else {
				vec.Vec_PtrPush(super, fanin)
			}
		}
	}
	vec.Vec_PtrFree(front)
	vec.Vec_PtrReverseOrder(super)
	vec.Vec_PtrUniqify(super, abcCompareObjIDs)
	return super
}

func Abc_NtkTopAnd(pNtk *Abc_Ntk_t) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	po := abc.Abc_NtkPo(pNtk, 0)
	root := abc.Abc_ObjChild0(po)
	assert(!root.Complement, "top PO root is regular")
	roots := Abc_NodeGetSuper(root.Object)
	assert(vec.Vec_PtrSize(roots) >= 2, "Vec_PtrSize(vNodes) >= 2")
	abc.Abc_NtkCleanCopy(pNtk)
	out := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	out.Name, out.Spec = pNtk.Name, pNtk.Spec
	abc.Abc_ObjSetCopyData(abc.Abc_AigConst1(pNtk).Object, abc.Abc_AigConst1(out))
	for _, pi := range pNtk.ObjectsOfKind(abc.ObjectPI) {
		abc.Abc_NtkDupObj(out, pi, 1)
	}
	rootObjects := make([]*abc.Object, vec.Vec_PtrSize(roots))
	for i := range rootObjects {
		rootObjects[i] = abc.Abc_ObjRegular(vec.Vec_PtrEntry(roots, i))
	}
	for _, node := range abc.Abc_NtkDfsIterNodes(pNtk, rootObjects) {
		strashNodeCopy(out, node)
	}
	for i := 0; i < vec.Vec_PtrSize(roots); i++ {
		edge := vec.Vec_PtrEntry(roots, i).(abc.Edge)
		driver := abc.Abc_ObjNotCond(abc.Abc_ObjCopyEdge(edge.Object), abcBool(edge.Complement))
		newPO := abc.Abc_NtkCreatePo(out)
		abc.Abc_ObjAddFanin(newPO, driver)
		abc.Abc_ObjAssignName(newPO, abc.Abc_ObjName(newPO), "")
	}
	vec.Vec_PtrFree(roots)
	if n := abc.Abc_AigCleanup(strashAig(out)); n != 0 {
		fmt.Printf("Abc_NtkTopAnd(): AIG cleanup removed %d nodes (this is a bug).\n", n)
	}
	if abc.Abc_NtkCheck(out) == 0 {
		fmt.Println("Abc_NtkStrash: The network check has failed.")
		abc.Abc_NtkDelete(out)
		return nil
	}
	return out
}

func abcBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func Abc_NtkWriteAig(pNtk *Abc_Ntk_t, pFileName string) {
	assert(abc.Abc_NtkIsStrash(pNtk) != 0, "Abc_NtkIsStrash(pNtk)")
	assert(abc.Abc_NtkLatchNum(pNtk) == 0, "Abc_NtkLatchNum(pNtk)==0")
	var file *os.File
	var writer io.Writer
	if pFileName == "" {
		file = os.Stdout
		writer = file
	} else {
		opened, err := os.OpenFile(pFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			fmt.Println("Cannot open output file.")
			return
		}
		file = opened
		writer = file
	}
	if file == nil {
		fmt.Println("Cannot open output file.")
		return
	}
	if file != os.Stdout {
		defer file.Close()
	}
	limit := int(2 * abc.Abc_NtkObjNumMax(pNtk))
	idToLit := make([]int32, limit)
	for i := range idToLit {
		idToLit[i] = -1
	}
	lit := int32(0)
	one := abc.Abc_AigConst1(pNtk)
	idToLit[2*one.Object.ID+1] = lit
	lit++
	idToLit[2*one.Object.ID] = lit
	lit++
	for _, pi := range pNtk.ObjectsOfKind(abc.ObjectPI) {
		idToLit[2*pi.ID] = lit
		lit++
		idToLit[2*pi.ID+1] = lit
		lit++
	}
	ands := make([]*abc.Object, 0)
	for id := int32(0); id < abc.Abc_NtkObjNumMax(pNtk); id++ {
		o := abc.Abc_NtkObj(pNtk, id)
		if o != nil && abc.Abc_AigNodeIsAnd(o) != 0 {
			ands = append(ands, o)
			idToLit[2*o.ID] = lit
			lit++
			idToLit[2*o.ID+1] = lit
			lit++
		}
	}
	fprintf := func(format string, args ...any) { _, _ = fmt.Fprintf(writer, format, args...) }
	fprintf("{\n    \"%s\", ", pNtk.Name)
	fprintf("//  pi=%d  po=%d  and=%d\n", abc.Abc_NtkPiNum(pNtk), abc.Abc_NtkPoNum(pNtk), abc.Abc_NtkNodeNum(pNtk))
	fprintf("    { ")
	for _, pi := range pNtk.ObjectsOfKind(abc.ObjectPI) {
		fprintf("\"%s\",", abc.Abc_ObjName(pi))
	}
	fprintf("NULL },\n    { ")
	for _, po := range pNtk.ObjectsOfKind(abc.ObjectPO) {
		fprintf("\"%s\",", abc.Abc_ObjName(po))
	}
	fprintf("NULL },\n    { ")
	for _, n := range ands {
		e := abc.Abc_ObjChild0(n)
		fprintf("%d,", idToLit[2*e.Object.ID+int(abcBool(e.Complement))])
	}
	fprintf("0 },\n    { ")
	for _, n := range ands {
		e := abc.Abc_ObjChild1(n)
		fprintf("%d,", idToLit[2*e.Object.ID+int(abcBool(e.Complement))])
	}
	fprintf("0 },\n    { ")
	for _, po := range pNtk.ObjectsOfKind(abc.ObjectPO) {
		e := abc.Abc_ObjChild0(po)
		fprintf("%d,", idToLit[2*e.Object.ID+int(abcBool(e.Complement))])
	}
	fprintf("0 },\n},\n")
}

func Abc_NtkPutOnTop(pNtk, pNtk2 *Abc_Ntk_t) *Abc_Ntk_t {
	assert(abc.Abc_NtkIsLogic(pNtk) != 0, "Abc_NtkIsLogic(pNtk)")
	assert(abc.Abc_NtkIsLogic(pNtk2) != 0, "Abc_NtkIsLogic(pNtk2)")
	assert(abc.Abc_NtkPoNum(pNtk) == abc.Abc_NtkPiNum(pNtk2), "matching interface size")
	abc.Abc_NtkCleanCopy(pNtk)
	abc.Abc_NtkCleanCopy(pNtk2)
	out := abc.Abc_NtkAlloc(abc.Abc_NtkType_t(pNtk.Type), abc.Abc_NtkFunc_t(pNtk.Function), 1)
	out.Name, out.Spec = pNtk.Name, pNtk.Spec
	for _, pi := range pNtk.ObjectsOfKind(abc.ObjectPI) {
		abc.Abc_NtkDupObj(out, pi, 1)
	}
	for _, node := range abc.Abc_NtkDfs(pNtk, 0) {
		copy := abc.Abc_NtkDupObj(out, node, 0)
		for i, e := range node.Fanins() {
			abc.Abc_ObjAddFanin(copy, strashCopyChild(node, int32(i)))
			_ = e
		}
	}
	for i, pi := range pNtk2.ObjectsOfKind(abc.ObjectPI) {
		abc.Abc_ObjSetCopyData(pi, strashCopyChild(abc.Abc_NtkPo(pNtk, int32(i)), 0))
	}
	for _, node := range abc.Abc_NtkDfs(pNtk2, 0) {
		copy := abc.Abc_NtkDupObj(out, node, 0)
		for i := range node.Fanins() {
			abc.Abc_ObjAddFanin(copy, strashCopyChild(node, int32(i)))
		}
	}
	for _, po := range pNtk2.ObjectsOfKind(abc.ObjectPO) {
		copy := abc.Abc_NtkDupObj(out, po, 1)
		abc.Abc_ObjAddFanin(copy, strashCopyChild(po, 0))
	}
	if abc.Abc_NtkCheck(out) == 0 {
		fmt.Fprintln(os.Stdout, "Abc_NtkPutOnTop(): Network check has failed.")
	}
	return out
}
