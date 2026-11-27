package int

// Line-by-line port of src/proof/int/intDup.c: duplication of the AIG for
// interpolation. The C pData copy pointers are the graph's copy map
// (Aig_ObjSetCopy / Aig_ObjChild0Copy).
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

// interForEachNode is Aig_ManForEachNode: the internal (AND/EXOR) objects in
// increasing ID order.
func interForEachNode(p *Aig_Man_t, f func(pObj Aig_Obj_t)) {
	for i := int32(0); i < aig.Aig_ManObjNumMax(p); i++ {
		pObj := aig.Aig_ManObj(p, i)
		if !p.ValidEdge(pObj) || !aig.Aig_ObjIsNode(p, pObj) {
			continue
		}
		f(pObj)
	}
}

// interCopyNodes is the common loop
// Aig_ManForEachNode( p, pObj, i ) pObj->pData = Aig_And( pNew, Aig_ObjChild0Copy(pObj), Aig_ObjChild1Copy(pObj) );
func interCopyNodes(p, pNew *Aig_Man_t) {
	interForEachNode(p, func(pObj Aig_Obj_t) {
		aig.Aig_ObjSetCopy(p, pObj, aig.Aig_And(pNew, aig.Aig_ObjChild0Copy(p, pObj), aig.Aig_ObjChild1Copy(p, pObj)))
	})
}

/*
*Function*************************************************************

	Synopsis    [Create trivial AIG manager for the init state.]

**********************************************************************
*/
func Inter_ManStartInitState(nRegs int32) *Aig_Man_t {
	var p *Aig_Man_t
	var pRes Aig_Obj_t
	var ppInputs []Aig_Obj_t
	var i int32
	assert(nRegs > 0, "nRegs > 0")
	ppInputs = make([]Aig_Obj_t, nRegs)
	p = aig.Aig_ManStart(nRegs)
	for i = 0; i < nRegs; i++ {
		ppInputs[i] = aig.Aig_Not(aig.Aig_ObjCreateCi(p))
	}
	pRes = aig.Aig_Multi(p, ppInputs, nRegs, int32(aig.AIG_OBJ_AND))
	aig.Aig_ObjCreateCo(p, pRes)
	return p
}

/*
*Function*************************************************************

	Synopsis    [Duplicate the AIG w/o POs.]

**********************************************************************
*/
func Inter_ManStartDuplicated(p *Aig_Man_t) *Aig_Man_t {
	var pNew *Aig_Man_t
	var pObj Aig_Obj_t
	var i int32
	assert(aig.Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0")
	// create the new manager
	pNew = aig.Aig_ManStart(aig.Aig_ManObjNumMax(p))
	pNew.Name = p.Name
	pNew.Spec = p.Spec
	// create the PIs
	aig.Aig_ManCleanData(p)
	aig.Aig_ObjSetCopy(p, aig.Aig_ManConst1(p), aig.Aig_ManConst1(pNew))
	for i = 0; i < aig.Aig_ManCiNum(p); i++ {
		aig.Aig_ObjSetCopy(p, aig.Aig_ManCi(p, i), aig.Aig_ObjCreateCi(pNew))
	}
	// set registers
	pNew.SetTruePiCount(p.TruePiCount())
	pNew.SetTruePoCount(saig.Saig_ManConstrNum(p))
	aig.Aig_ManSetRegNumForConstruction(pNew, aig.Aig_ManRegNum(p))
	// duplicate internal nodes
	interCopyNodes(p, pNew)
	// create constraint outputs
	for i = 0; i < saig.Saig_ManPoNum(p); i++ {
		if i < saig.Saig_ManPoNum(p)-saig.Saig_ManConstrNum(p) {
			continue
		}
		pObj = aig.Aig_ManCo(p, i)
		aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(aig.Aig_ObjChild0Copy(p, pObj)))
	}
	// create register inputs with MUXes
	for i = 0; i < saig.Saig_ManRegNum(p); i++ {
		pObj = saig.Saig_ManLi(p, i)
		aig.Aig_ObjCreateCo(pNew, aig.Aig_ObjChild0Copy(p, pObj))
	}
	aig.Aig_ManCleanup(pNew)
	aig.Aig_ManSetCioIds(pNew)
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Duplicate the AIG w/o POs and transforms to transit into init state.]

**********************************************************************
*/
func Inter_ManStartOneOutput(p *Aig_Man_t, fAddFirstPo int32) *Aig_Man_t {
	var pNew *Aig_Man_t
	var pObj, pObjLi, pObjLo Aig_Obj_t
	pCtrl := aig.InvalidEdge // Suppress "might be used uninitialized"
	var i int32
	assert(aig.Aig_ManRegNum(p) > 0, "Aig_ManRegNum(p) > 0")
	// create the new manager
	pNew = aig.Aig_ManStart(aig.Aig_ManObjNumMax(p))
	pNew.Name = p.Name
	pNew.Spec = p.Spec
	// create the PIs
	aig.Aig_ManCleanData(p)
	aig.Aig_ObjSetCopy(p, aig.Aig_ManConst1(p), aig.Aig_ManConst1(pNew))
	for i = 0; i < aig.Aig_ManCiNum(p); i++ {
		pObj = aig.Aig_ManCi(p, i)
		if i == saig.Saig_ManPiNum(p) {
			pCtrl = aig.Aig_ObjCreateCi(pNew)
		}
		aig.Aig_ObjSetCopy(p, pObj, aig.Aig_ObjCreateCi(pNew))
	}
	// set registers
	if fAddFirstPo != 0 {
		aig.Aig_ManSetRegNumForConstruction(pNew, 0)
		pNew.SetTruePiCount(aig.Aig_ManCiNum(p) + 1)
	} else {
		aig.Aig_ManSetRegNumForConstruction(pNew, aig.Aig_ManRegNum(p))
		pNew.SetTruePiCount(p.TruePiCount() + 1)
	}
	pNew.SetTruePoCount(fAddFirstPo + saig.Saig_ManConstrNum(p))
	// duplicate internal nodes
	interCopyNodes(p, pNew)
	// create constraint outputs
	for i = 0; i < saig.Saig_ManPoNum(p); i++ {
		if i < saig.Saig_ManPoNum(p)-saig.Saig_ManConstrNum(p) {
			continue
		}
		pObj = aig.Aig_ManCo(p, i)
		aig.Aig_ObjCreateCo(pNew, aig.Aig_Not(aig.Aig_ObjChild0Copy(p, pObj)))
	}
	// add the PO
	if fAddFirstPo != 0 {
		pObj = aig.Aig_ManCo(p, 0)
		aig.Aig_ObjCreateCo(pNew, aig.Aig_ObjChild0Copy(p, pObj))
	} else {
		// create register inputs with MUXes
		for i = 0; i < saig.Saig_ManRegNum(p); i++ {
			pObjLi, pObjLo = saig.Saig_ManLi(p, i), saig.Saig_ManLo(p, i)
			pObj = aig.Aig_Mux(pNew, pCtrl, aig.Aig_ObjCopy(p, pObjLo), aig.Aig_ObjChild0Copy(p, pObjLi))
			//        pObj = Aig_Mux( pNew, pCtrl, Aig_ManConst0(pNew), Aig_ObjChild0Copy(pObjLi) );
			aig.Aig_ObjCreateCo(pNew, pObj)
		}
	}
	aig.Aig_ManCleanup(pNew)
	aig.Aig_ManSetCioIds(pNew)
	return pNew
}
