package int

// Line-by-line port of src/proof/int/intCtrex.c. (The int command's real
// counterexamples come from Saig_ManBmcScalable in intCore.c; these
// routines are kept as in C.)
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

/*
*Function*************************************************************

	Synopsis    [Unroll the circuit the given number of timeframes.]

**********************************************************************
*/
func Inter_ManFramesBmc(pAig *Aig_Man_t, nFrames int32) *Aig_Man_t {
	var pFrames *Aig_Man_t
	var pObj, pObjLi, pObjLo Aig_Obj_t
	var i, f int32
	assert(saig.Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0")
	assert(saig.Saig_ManPoNum(pAig) == 1, "Saig_ManPoNum(pAig) == 1")
	pFrames = aig.Aig_ManStart(aig.Aig_ManNodeNum(pAig) * nFrames)
	// map the constant node
	aig.Aig_ObjSetCopy(pAig, aig.Aig_ManConst1(pAig), aig.Aig_ManConst1(pFrames))
	// create variables for register outputs
	for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
		aig.Aig_ObjSetCopy(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ManConst0(pFrames))
	}
	// add timeframes
	for f = 0; f < nFrames; f++ {
		// create PI nodes for this frame
		for i = 0; i < saig.Saig_ManPiNum(pAig); i++ {
			aig.Aig_ObjSetCopy(pAig, aig.Aig_ManCi(pAig, i), aig.Aig_ObjCreateCi(pFrames))
		}
		// add internal nodes of this frame
		interCopyNodes(pAig, pFrames)
		if f == nFrames-1 {
			break
		}
		// transfer to register outputs
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			pObjLi = saig.Saig_ManLi(pAig, i)
			aig.Aig_ObjSetCopy(pAig, pObjLi, aig.Aig_ObjChild0Copy(pAig, pObjLi))
		}
		// transfer to register outputs
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			pObjLi, pObjLo = saig.Saig_ManLi(pAig, i), saig.Saig_ManLo(pAig, i)
			aig.Aig_ObjSetCopy(pAig, pObjLo, aig.Aig_ObjCopy(pAig, pObjLi))
		}
	}
	// create POs for the output of the last frame
	pObj = aig.Aig_ManCo(pAig, 0)
	aig.Aig_ObjCreateCo(pFrames, aig.Aig_ObjChild0Copy(pAig, pObj))
	aig.Aig_ManCleanup(pFrames)
	return pFrames
}

/*
*Function*************************************************************

	Synopsis    [Run the SAT solver on the unrolled instance.]

**********************************************************************
*/
func Inter_ManGetCounterExample(pAig *Aig_Man_t, nFrames, fVerbose int32) *util.Abc_Cex_t {
	nConfLimit := int64(1000000)
	var pCtrex *util.Abc_Cex_t
	var pFrames *Aig_Man_t
	var pSat *sat_solver
	var pCnf *Cnf_Dat_t
	var status int32
	clk := util.Abc_Clock()
	var vCiIds *Vec_Int_t
	// create timeframes
	assert(saig.Saig_ManPoNum(pAig) == 1, "Saig_ManPoNum(pAig) == 1")
	pFrames = Inter_ManFramesBmc(pAig, nFrames)
	// derive CNF
	pCnf = cnf.Cnf_Derive(pFrames, 0)
	cnf.Cnf_DataTranformPolarity(pCnf, 0)
	vCiIds = cnf.Cnf_DataCollectPiSatNums(pCnf, pFrames)
	aig.Aig_ManStop(pFrames)
	// convert into SAT solver
	pSat = cnf.Cnf_DataWriteIntoSolver(pCnf, 1, 0)
	cnf.Cnf_DataFree(pCnf)
	if pSat == nil {
		fmt.Printf("Counter-example generation in command \"int\" has failed.\n")
		fmt.Printf("Use command \"bmc2\" to produce a valid counter-example.\n")
		return nil
	}
	// simplify the problem
	status = pSat.SimplifyC()
	if status == 0 {
		pSat.Close()
		return nil
	}
	// solve the miter
	status = pSat.SolveC(nil, nConfLimit, 0, 0, 0)
	// if the problem is SAT, get the counterexample
	if status == int32(bsat.Satisfiable) {
		pModel := bsat.Sat_SolverGetModel(pSat, vec.Vec_IntArray(vCiIds), int32(vec.Vec_IntSize(vCiIds)))
		pCtrex = util.Abc_CexAlloc(saig.Saig_ManRegNum(pAig), saig.Saig_ManPiNum(pAig), nFrames)
		pCtrex.IFrame = nFrames - 1
		pCtrex.IPo = 0
		for i := int32(0); i < int32(vec.Vec_IntSize(vCiIds)); i++ {
			if pModel[i] != 0 {
				util.Abc_InfoSetBit(pCtrex.PData, saig.Saig_ManRegNum(pAig)+i)
			}
		}
	}
	// free the sat_solver
	pSat.Close()
	// verify counter-example
	status = saig.Saig_ManVerifyCex(pAig, pCtrex)
	if status == 0 {
		fmt.Printf("Inter_ManGetCounterExample(): Counter-example verification has FAILED.\n")
	}
	// report the results
	if fVerbose != 0 {
		util.Abc_PrintTime(1, "Total ctrex generation time", util.Abc_Clock()-clk)
	}
	return pCtrex
}
