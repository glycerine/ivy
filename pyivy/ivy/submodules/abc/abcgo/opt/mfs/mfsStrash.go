package mfs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
)

// Port of mfsStrash.c. The C code keeps AIG pointers in the Hop pData
// fields and in the pCopy/pNext fields of the network objects; here AIG
// literals are kept in the Hop iData fields and in Mfs_Man_t.objCopy/objNext.
// (Abc_MfsConvertAigToHop, Abc_NtkAigForConstraints and
// Abc_NtkConstraintRatio are only referenced from commented-out code.)

// Abc_MfsConvertHopToAig_rec converts the Hop cone into the AIG manager.
func Abc_MfsConvertHopToAig_rec(pObj *hop.Hop_Obj_t, pMan *aig.Graph) {
	assert(hop.Hop_IsComplement(pObj) == 0, "!Hop_IsComplement(pObj)")
	if hop.Hop_ObjIsNode(pObj) == 0 || hop.Hop_ObjIsMarkA(pObj) != 0 {
		return
	}
	Abc_MfsConvertHopToAig_rec(hop.Hop_ObjFanin0(pObj), pMan)
	Abc_MfsConvertHopToAig_rec(hop.Hop_ObjFanin1(pObj), pMan)
	hop.Hop_ObjSetIData(pObj, int32(aig.Aig_And(pMan, aig.Edge(hop.Hop_ObjChild0CopyI(pObj)), aig.Edge(hop.Hop_ObjChild1CopyI(pObj)))))
	assert(hop.Hop_ObjIsMarkA(pObj) == 0, "loop detection")
	hop.Hop_ObjSetMarkA(pObj)
}

// Abc_MfsConvertHopToAig converts the local function of the node twice:
// once over the fanin pCopy literals and once over the fanin pNext literals.
func Abc_MfsConvertHopToAig(p *Mfs_Man_t, pObjOld *abc.Object, pMan *aig.Graph) {
	// get the local AIG
	pHopMan := pObjOld.Network().FunctionManager.(*hop.Hop_Man_t)
	pRoot := pObjOld.Data.(*hop.Hop_Obj_t)
	id := pObjOld.ID
	// check the case of a constant
	if hop.Hop_ObjIsConst1(hop.Hop_Regular(pRoot)) != 0 {
		p.objCopy[id] = aig.Aig_NotCond(aig.Aig_ManConst1(pMan), hop.Hop_IsComplement(pRoot))
		p.objNext[id] = p.objCopy[id]
		return
	}
	nFanins := abc.Abc_ObjFaninNum(pObjOld)
	// assign the fanin nodes
	for i := int32(0); i < nFanins; i++ {
		hop.Hop_ObjSetIData(hop.Hop_ManPi(pHopMan, i), int32(p.objCopy[abc.Abc_ObjFanin(pObjOld, i).ID]))
	}
	// construct the AIG
	Abc_MfsConvertHopToAig_rec(hop.Hop_Regular(pRoot), pMan)
	p.objCopy[id] = aig.Aig_NotCond(aig.Edge(hop.Hop_ObjIData(hop.Hop_Regular(pRoot))), hop.Hop_IsComplement(pRoot))
	hop.Hop_ConeUnmark_rec(hop.Hop_Regular(pRoot))
	// assign the fanin nodes
	for i := int32(0); i < nFanins; i++ {
		hop.Hop_ObjSetIData(hop.Hop_ManPi(pHopMan, i), int32(p.objNext[abc.Abc_ObjFanin(pObjOld, i).ID]))
	}
	// construct the AIG
	Abc_MfsConvertHopToAig_rec(hop.Hop_Regular(pRoot), pMan)
	p.objNext[id] = aig.Aig_NotCond(aig.Edge(hop.Hop_ObjIData(hop.Hop_Regular(pRoot))), hop.Hop_IsComplement(pRoot))
	hop.Hop_ConeUnmark_rec(hop.Hop_Regular(pRoot))
}

// Abc_NtkConstructAig_rec computes the observability condition of the node.
func Abc_NtkConstructAig_rec(p *Mfs_Man_t, pNode *abc.Object, pMan *aig.Graph) aig.Edge {
	p.ensureObj(int(abc.Abc_NtkObjNumMax(p.pNtk)) - 1)
	// assign AIG nodes to the leaves
	for _, pObj := range p.vSupp {
		e := aig.Aig_ObjCreateCi(pMan)
		p.objCopy[pObj.ID] = e
		p.objNext[pObj.ID] = e
	}
	// strash intermediate nodes
	abc.Abc_NtkIncrementTravId(abc.Abc_ObjNtk(pNode))
	for _, pObj := range p.vNodes {
		Abc_MfsConvertHopToAig(p, pObj, pMan)
		if pObj == pNode {
			p.objNext[pObj.ID] = p.objNext[pObj.ID].Not()
		}
	}
	// create the observability condition
	pRoot := aig.Aig_ManConst0(pMan)
	for _, pObj := range p.vRoots {
		pExor := aig.Aig_Exor(pMan, p.objCopy[pObj.ID], p.objNext[pObj.ID])
		pRoot = aig.Aig_Or(pMan, pRoot, pExor)
	}
	return pRoot
}

// careData holds the pData fields of the care-set AIG objects.
type careData map[int]aig.Edge

// Abc_NtkConstructCare_rec creates the care-set constraint in the window AIG.
// aig.InvalidEdge plays the role of NULL.
func Abc_NtkConstructCare_rec(pCare *aig.Graph, pObj aig.Edge, pMan *aig.Graph, data careData) aig.Edge {
	if aig.Aig_ObjIsTravIdCurrent(pCare, pObj) {
		if e, ok := data[pObj.NodeID()]; ok {
			return e
		}
		return aig.InvalidEdge
	}
	aig.Aig_ObjSetTravIdCurrent(pCare, pObj)
	if aig.Aig_ObjIsCi(pCare, pObj) {
		data[pObj.NodeID()] = aig.InvalidEdge
		return aig.InvalidEdge
	}
	pObj0 := Abc_NtkConstructCare_rec(pCare, aig.Aig_ObjFanin0(pCare, pObj), pMan, data)
	if pObj0 == aig.InvalidEdge {
		data[pObj.NodeID()] = aig.InvalidEdge
		return aig.InvalidEdge
	}
	pObj1 := Abc_NtkConstructCare_rec(pCare, aig.Aig_ObjFanin1(pCare, pObj), pMan, data)
	if pObj1 == aig.InvalidEdge {
		data[pObj.NodeID()] = aig.InvalidEdge
		return aig.InvalidEdge
	}
	pObj0 = aig.Aig_NotCond(pObj0, aig.Aig_ObjFaninC0(pCare, pObj))
	pObj1 = aig.Aig_NotCond(pObj1, aig.Aig_ObjFaninC1(pCare, pObj))
	r := aig.Aig_And(pMan, pObj0, pObj1)
	data[pObj.NodeID()] = r
	return r
}

// Abc_NtkConstructAig creates AIG for the window with constraints.
func Abc_NtkConstructAig(p *Mfs_Man_t, pNode *abc.Object) *aig.Graph {
	// start the new manager
	pMan := aig.Aig_ManStart(1000)
	// construct the root node's AIG cone
	pObjAig := Abc_NtkConstructAig_rec(p, pNode, pMan)
	aig.Aig_ObjCreateCo(pMan, pObjAig)
	if p.pCare != nil {
		data := careData{}
		// mark the care set
		aig.Aig_ManIncrementTravId(p.pCare)
		for _, pFanin := range p.vSupp {
			pPi := aig.Aig_ManCi(p.pCare, p.ciIndex[pFanin.ID])
			aig.Aig_ObjSetTravIdCurrent(p.pCare, pPi)
			data[pPi.NodeID()] = p.objCopy[pFanin.ID]
		}
		// construct the constraints
		for _, pFanin := range p.vSupp {
			vOuts := p.vSuppsInv[p.ciIndex[pFanin.ID]]
			for _, iOut := range vOuts {
				pPo := aig.Aig_ManCo(p.pCare, iOut)
				if aig.Aig_ObjIsTravIdCurrent(p.pCare, pPo) {
					continue
				}
				aig.Aig_ObjSetTravIdCurrent(p.pCare, pPo)
				if aig.Aig_ObjFanin0(p.pCare, pPo) == aig.Aig_ManConst1(p.pCare) {
					continue
				}
				pObjAig = Abc_NtkConstructCare_rec(p.pCare, aig.Aig_ObjFanin0(p.pCare, pPo), pMan, data)
				if pObjAig == aig.InvalidEdge {
					continue
				}
				pObjAig = aig.Aig_NotCond(pObjAig, aig.Aig_ObjFaninC0(p.pCare, pPo))
				aig.Aig_ObjCreateCo(pMan, pObjAig)
			}
		}
	}
	if p.pPars.FResub != 0 {
		// construct the node
		aig.Aig_ObjCreateCo(pMan, p.objCopy[pNode.ID])
		// construct the divisors
		for _, pFanin := range p.vDivs {
			aig.Aig_ObjCreateCo(pMan, p.objCopy[pFanin.ID])
		}
	} else {
		// construct the fanins
		for i := int32(0); i < abc.Abc_ObjFaninNum(pNode); i++ {
			aig.Aig_ObjCreateCo(pMan, p.objCopy[abc.Abc_ObjFanin(pNode, i).ID])
		}
	}
	aig.Aig_ManCleanup(pMan)
	return pMan
}
