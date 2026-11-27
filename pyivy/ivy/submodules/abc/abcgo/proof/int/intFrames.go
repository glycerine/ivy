package int

// Line-by-line port of src/proof/int/intFrames.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

/*
*Function*************************************************************

	Synopsis    [Create timeframes of the manager for interpolation.]

	Description [The resulting manager is combinational. The primary inputs
	corresponding to register outputs are ordered first. The only POs of the
	manager is the property output of the last timeframe.]

**********************************************************************
*/
func Inter_ManFramesInter(pAig *Aig_Man_t, nFrames, fAddRegOuts, fUseTwoFrames int32) *Aig_Man_t {
	var pFrames *Aig_Man_t
	var pObj, pObjLi, pObjLo Aig_Obj_t
	pLastPo := aig.InvalidEdge
	var i, f int32
	assert(saig.Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0")
	assert(saig.Saig_ManPoNum(pAig)-saig.Saig_ManConstrNum(pAig) == 1, "Saig_ManPoNum(pAig)-Saig_ManConstrNum(pAig) == 1")
	pFrames = aig.Aig_ManStart(aig.Aig_ManNodeNum(pAig) * nFrames)
	// map the constant node
	aig.Aig_ObjSetCopy(pAig, aig.Aig_ManConst1(pAig), aig.Aig_ManConst1(pFrames))
	// create variables for register outputs
	if fAddRegOuts != 0 {
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			aig.Aig_ObjSetCopy(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ManConst0(pFrames))
		}
	} else {
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			aig.Aig_ObjSetCopy(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ObjCreateCi(pFrames))
		}
	}
	// add timeframes
	for f = 0; f < nFrames; f++ {
		// create PI nodes for this frame
		for i = 0; i < saig.Saig_ManPiNum(pAig); i++ {
			aig.Aig_ObjSetCopy(pAig, aig.Aig_ManCi(pAig, i), aig.Aig_ObjCreateCi(pFrames))
		}
		// add internal nodes of this frame
		interCopyNodes(pAig, pFrames)
		// add outputs for constraints
		for i = 0; i < saig.Saig_ManPoNum(pAig); i++ {
			if i < saig.Saig_ManPoNum(pAig)-saig.Saig_ManConstrNum(pAig) {
				continue
			}
			pObj = aig.Aig_ManCo(pAig, i)
			aig.Aig_ObjCreateCo(pFrames, aig.Aig_Not(aig.Aig_ObjChild0Copy(pAig, pObj)))
		}
		if f == nFrames-1 {
			break
		}
		// remember the last PO
		pObj = aig.Aig_ManCo(pAig, 0)
		pLastPo = aig.Aig_ObjChild0Copy(pAig, pObj)
		// save register inputs
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			pObj = saig.Saig_ManLi(pAig, i)
			aig.Aig_ObjSetCopy(pAig, pObj, aig.Aig_ObjChild0Copy(pAig, pObj))
		}
		// transfer to register outputs
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			pObjLi, pObjLo = saig.Saig_ManLi(pAig, i), saig.Saig_ManLo(pAig, i)
			aig.Aig_ObjSetCopy(pAig, pObjLo, aig.Aig_ObjCopy(pAig, pObjLi))
		}
	}
	// create POs for each register output
	if fAddRegOuts != 0 {
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			pObj = saig.Saig_ManLi(pAig, i)
			aig.Aig_ObjCreateCo(pFrames, aig.Aig_ObjChild0Copy(pAig, pObj))
		}
	} else {
		// create the only PO of the manager
		pObj = aig.Aig_ManCo(pAig, 0)
		// add the last PO
		if pLastPo == aig.InvalidEdge || fUseTwoFrames == 0 {
			pLastPo = aig.Aig_ObjChild0Copy(pAig, pObj)
		} else {
			pLastPo = aig.Aig_Or(pFrames, pLastPo, aig.Aig_ObjChild0Copy(pAig, pObj))
		}
		aig.Aig_ObjCreateCo(pFrames, pLastPo)
	}
	aig.Aig_ManCleanup(pFrames)
	return pFrames
}
