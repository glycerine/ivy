package int

// Line-by-line port of src/proof/int/intContain.c.
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/proof/framgr"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

/*
*Function*************************************************************

	Synopsis    [Checks constainment of two interpolants.]

**********************************************************************
*/
func Inter_ManCheckContainment(pNew, pOld *Aig_Man_t) int32 {
	var pMiter, pAigTemp *Aig_Man_t
	var RetValue int32
	pMiter = aig.Aig_ManCreateMiter(pNew, pOld, 1)
	//    pMiter = Dar_ManRwsat( pAigTemp = pMiter, 1, 0 );
	//    Aig_ManStop( pAigTemp );
	RetValue = framgr.Fra_FraigMiterStatus(pMiter)
	if RetValue == -1 {
		pAigTemp = framgr.Fra_FraigEquivence(pMiter, 1000000, 1)
		RetValue = framgr.Fra_FraigMiterStatus(pAigTemp)
		aig.Aig_ManStop(pAigTemp)
		//        RetValue = Fra_FraigSat( pMiter, 1000000, 0, 0, 0, 0, 0, 0, 0, 0 );
	}
	assert(RetValue != -1, "RetValue != -1")
	aig.Aig_ManStop(pMiter)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Checks constainment of two interpolants.]

**********************************************************************
*/
func Inter_ManCheckEquivalence(pNew, pOld *Aig_Man_t) int32 {
	var pMiter, pAigTemp *Aig_Man_t
	var RetValue int32
	pMiter = aig.Aig_ManCreateMiter(pNew, pOld, 0)
	RetValue = framgr.Fra_FraigMiterStatus(pMiter)
	if RetValue == -1 {
		pAigTemp = framgr.Fra_FraigEquivence(pMiter, 1000000, 1)
		RetValue = framgr.Fra_FraigMiterStatus(pAigTemp)
		aig.Aig_ManStop(pAigTemp)
	}
	assert(RetValue != -1, "RetValue != -1")
	aig.Aig_ManStop(pMiter)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    [Create timeframes of the manager for interpolation.]

	Description [The resulting manager is combinational. The primary inputs
	corresponding to register outputs are ordered first.]

**********************************************************************
*/
func Inter_ManFramesLatches(pAig *Aig_Man_t, nFrames int32, pvMapReg *[]Aig_Obj_t) *Aig_Man_t {
	var pFrames *Aig_Man_t
	var pObj, pObjLi, pObjLo Aig_Obj_t
	var i, f int32
	assert(saig.Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0")
	pFrames = aig.Aig_ManStart(aig.Aig_ManNodeNum(pAig) * nFrames)
	// map the constant node
	aig.Aig_ObjSetCopy(pAig, aig.Aig_ManConst1(pAig), aig.Aig_ManConst1(pFrames))
	// create variables for register outputs
	*pvMapReg = make([]Aig_Obj_t, 0, (nFrames+1)*saig.Saig_ManRegNum(pAig))
	for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
		pObj = saig.Saig_ManLo(pAig, i)
		aig.Aig_ObjSetCopy(pAig, pObj, aig.Aig_ObjCreateCi(pFrames))
		*pvMapReg = append(*pvMapReg, aig.Aig_ObjCopy(pAig, pObj))
	}
	// add timeframes
	for f = 0; f < nFrames; f++ {
		// create PI nodes for this frame
		for i = 0; i < saig.Saig_ManPiNum(pAig); i++ {
			aig.Aig_ObjSetCopy(pAig, aig.Aig_ManCi(pAig, i), aig.Aig_ObjCreateCi(pFrames))
		}
		// add internal nodes of this frame
		interCopyNodes(pAig, pFrames)
		// save register inputs
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			pObj = saig.Saig_ManLi(pAig, i)
			aig.Aig_ObjSetCopy(pAig, pObj, aig.Aig_ObjChild0Copy(pAig, pObj))
		}
		// transfer to register outputs
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			pObjLi, pObjLo = saig.Saig_ManLi(pAig, i), saig.Saig_ManLo(pAig, i)
			aig.Aig_ObjSetCopy(pAig, pObjLo, aig.Aig_ObjCopy(pAig, pObjLi))
			*pvMapReg = append(*pvMapReg, aig.Aig_ObjCopy(pAig, pObjLo))
		}
	}
	return pFrames
}

/*
*Function*************************************************************

	Synopsis    [Duplicates AIG while mapping PIs into the given array.]

**********************************************************************
*/
func Inter_ManAppendCone(pOld, pNew *Aig_Man_t, ppNewPis []Aig_Obj_t, fCompl int32) {
	var pObj Aig_Obj_t
	var i int32
	assert(aig.Aig_ManCoNum(pOld) == 1, "Aig_ManCoNum(pOld) == 1")
	// create the PIs
	aig.Aig_ManCleanData(pOld)
	aig.Aig_ObjSetCopy(pOld, aig.Aig_ManConst1(pOld), aig.Aig_ManConst1(pNew))
	for i = 0; i < aig.Aig_ManCiNum(pOld); i++ {
		aig.Aig_ObjSetCopy(pOld, aig.Aig_ManCi(pOld, i), ppNewPis[i])
	}
	// duplicate internal nodes
	interCopyNodes(pOld, pNew)
	// add one PO to new
	pObj = aig.Aig_ManCo(pOld, 0)
	aig.Aig_ObjCreateCo(pNew, aig.Aig_NotCond(aig.Aig_ObjChild0Copy(pOld, pObj), fCompl))
}

/*
*Function*************************************************************

	Synopsis    [Checks constainment of two interpolants inductively.]

**********************************************************************
*/
func Inter_ManCheckInductiveContainment(pTrans, pInter *Aig_Man_t, nSteps, fBackward int32) int32 {
	var pFrames *Aig_Man_t
	var ppNodes []Aig_Obj_t
	var vMapRegs []Aig_Obj_t
	var pCnf *Cnf_Dat_t
	var pSat *sat_solver
	var f, nRegs, status int32
	nRegs = saig.Saig_ManRegNum(pTrans)
	assert(nRegs > 0, "nRegs > 0")
	// generate the timeframes
	pFrames = Inter_ManFramesLatches(pTrans, nSteps, &vMapRegs)
	assert(int32(len(vMapRegs)) == (nSteps+1)*nRegs, "Vec_PtrSize(vMapRegs) == (nSteps + 1) * nRegs")
	// add main constraints to the timeframes
	ppNodes = vMapRegs
	if fBackward == 0 {
		// forward inductive check: p -> p -> ... -> !p
		for f = 0; f < nSteps; f++ {
			Inter_ManAppendCone(pInter, pFrames, ppNodes[f*nRegs:], 0)
		}
		Inter_ManAppendCone(pInter, pFrames, ppNodes[f*nRegs:], 1)
	} else {
		// backward inductive check: p -> !p -> ... -> !p
		Inter_ManAppendCone(pInter, pFrames, ppNodes[0*nRegs:], 1)
		for f = 1; f <= nSteps; f++ {
			Inter_ManAppendCone(pInter, pFrames, ppNodes[f*nRegs:], 0)
		}
	}
	vMapRegs = nil
	aig.Aig_ManCleanup(pFrames)

	// convert to CNF
	pCnf = cnf.Cnf_Derive(pFrames, 0)
	pSat = cnf.Cnf_DataWriteIntoSolver(pCnf, 1, 0)
	//    Cnf_DataFree( pCnf );
	//    Aig_ManStop( pFrames );

	if pSat == nil {
		cnf.Cnf_DataFree(pCnf)
		aig.Aig_ManStop(pFrames)
		return 1
	}

	// solve the problem
	status = pSat.SolveC(nil, 0, 0, 0, 0)

	//    Inter_ManCheckUniqueness( pTrans, pSat, pCnf, nSteps );

	cnf.Cnf_DataFree(pCnf)
	aig.Aig_ManStop(pFrames)

	pSat.Close()
	return b2i(status == int32(bsat.Unsatisfiable))
}

/*
*Function*************************************************************

	Synopsis    [Check if cex satisfies uniqueness constraints.]

	Description [The C routine asserts for nFrames > 1 before its unfinished
	simulation code.]

**********************************************************************
*/
func Inter_ManCheckUniqueness(p *Aig_Man_t, pSat *sat_solver, pCnf *Cnf_Dat_t, nFrames int32) int32 {
	if nFrames == 1 {
		return 1
	}
	assert(false, "0")
	fmt.Printf("Uniquness does not hold in %d frames.\n", 0)
	return 1
}
