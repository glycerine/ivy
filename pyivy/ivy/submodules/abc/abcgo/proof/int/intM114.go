package int

// Line-by-line port of src/proof/int/intM114.c: one interpolation step with
// the internal proof-logging sat_solver (Sto_Man_t) and Inta_ManInterpolate.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

func toLitCond(v, c int32) bsat.Literal { return bsat.Literal(v + v + c) }

/*
*Function*************************************************************

	Synopsis    [Returns the SAT solver for one interpolation run.]

	Description [pInter is the previous interpolant. pAig is one time frame.
	pFrames is the unrolled time frames.]

**********************************************************************
*/
func Inter_ManDeriveSatSolver(
	pInter *Aig_Man_t, pCnfInter *Cnf_Dat_t,
	pAig *Aig_Man_t, pCnfAig *Cnf_Dat_t,
	pFrames *Aig_Man_t, pCnfFrames *Cnf_Dat_t,
	vVarsAB *Vec_Int_t, fUseBackward int32) *sat_solver {
	var pSat *sat_solver
	var pObj, pObj2 Aig_Obj_t
	var i int32
	var Lits [2]bsat.Literal

	// sanity checks
	assert(aig.Aig_ManRegNum(pInter) == 0, "Aig_ManRegNum(pInter) == 0")
	assert(aig.Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0")
	assert(aig.Aig_ManRegNum(pFrames) == 0, "Aig_ManRegNum(pFrames) == 0")
	assert(aig.Aig_ManCoNum(pInter) == 1, "Aig_ManCoNum(pInter) == 1")
	assert(fUseBackward != 0 || aig.Aig_ManCiNum(pInter) == aig.Aig_ManRegNum(pAig), "fUseBackward || Aig_ManCiNum(pInter) == Aig_ManRegNum(pAig)")

	// prepare CNFs
	cnf.Cnf_DataLift(pCnfAig, int32(pCnfFrames.Variables))
	cnf.Cnf_DataLift(pCnfInter, int32(pCnfFrames.Variables+pCnfAig.Variables))

	// start the solver
	pSat = bsat.NewSolver()
	pSat.StoreAllocC()
	pSat.SetnvarsC(int32(pCnfInter.Variables + pCnfAig.Variables + pCnfFrames.Variables))

	// add clauses of A
	// interpolant
	for i = 0; i < int32(len(pCnfInter.Clauses)); i++ {
		if pSat.AddClauseC(pCnfInter.Clauses[i]) == 0 {
			pSat.Close()
			// return clauses to the original state
			cnf.Cnf_DataLift(pCnfAig, -int32(pCnfFrames.Variables))
			cnf.Cnf_DataLift(pCnfInter, -int32(pCnfFrames.Variables)-int32(pCnfAig.Variables))
			return nil
		}
	}
	// connector clauses
	if fUseBackward != 0 {
		for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
			pObj2 = saig.Saig_ManLi(pAig, i)
			if saig.Saig_ManRegNum(pAig) == aig.Aig_ManCiNum(pInter) {
				pObj = aig.Aig_ManCi(pInter, i)
			} else {
				assert(aig.Aig_ManCiNum(pAig) == aig.Aig_ManCiNum(pInter), "Aig_ManCiNum(pAig) == Aig_ManCiNum(pInter)")
				pObj = aig.Aig_ManCi(pInter, aig.Aig_ManCiNum(pAig)-saig.Saig_ManRegNum(pAig)+i)
			}

			Lits[0] = toLitCond(pCnfInter.VarNums[aig.Aig_ObjId(pObj)], 0)
			Lits[1] = toLitCond(pCnfAig.VarNums[aig.Aig_ObjId(pObj2)], 1)
			if pSat.AddClauseC(Lits[:]) == 0 {
				assert(false, "0")
			}
			Lits[0] = toLitCond(pCnfInter.VarNums[aig.Aig_ObjId(pObj)], 1)
			Lits[1] = toLitCond(pCnfAig.VarNums[aig.Aig_ObjId(pObj2)], 0)
			if pSat.AddClauseC(Lits[:]) == 0 {
				assert(false, "0")
			}
		}
	} else {
		for i = 0; i < aig.Aig_ManCiNum(pInter); i++ {
			pObj = aig.Aig_ManCi(pInter, i)
			pObj2 = saig.Saig_ManLo(pAig, i)

			Lits[0] = toLitCond(pCnfInter.VarNums[aig.Aig_ObjId(pObj)], 0)
			Lits[1] = toLitCond(pCnfAig.VarNums[aig.Aig_ObjId(pObj2)], 1)
			if pSat.AddClauseC(Lits[:]) == 0 {
				assert(false, "0")
			}
			Lits[0] = toLitCond(pCnfInter.VarNums[aig.Aig_ObjId(pObj)], 1)
			Lits[1] = toLitCond(pCnfAig.VarNums[aig.Aig_ObjId(pObj2)], 0)
			if pSat.AddClauseC(Lits[:]) == 0 {
				assert(false, "0")
			}
		}
	}
	// one timeframe
	for i = 0; i < int32(len(pCnfAig.Clauses)); i++ {
		if pSat.AddClauseC(pCnfAig.Clauses[i]) == 0 {
			assert(false, "0")
		}
	}
	// connector clauses
	vec.Vec_IntClear(vVarsAB)
	if fUseBackward != 0 {
		for i = 0; i < aig.Aig_ManCoNum(pFrames); i++ {
			pObj = aig.Aig_ManCo(pFrames, i)
			assert(pCnfFrames.VarNums[aig.Aig_ObjId(pObj)] >= 0, "pCnfFrames->pVarNums[pObj->Id] >= 0")
			vec.Vec_IntPush(vVarsAB, pCnfFrames.VarNums[aig.Aig_ObjId(pObj)])

			pObj2 = saig.Saig_ManLo(pAig, i)
			Lits[0] = toLitCond(pCnfFrames.VarNums[aig.Aig_ObjId(pObj)], 0)
			Lits[1] = toLitCond(pCnfAig.VarNums[aig.Aig_ObjId(pObj2)], 1)
			if pSat.AddClauseC(Lits[:]) == 0 {
				assert(false, "0")
			}
			Lits[0] = toLitCond(pCnfFrames.VarNums[aig.Aig_ObjId(pObj)], 1)
			Lits[1] = toLitCond(pCnfAig.VarNums[aig.Aig_ObjId(pObj2)], 0)
			if pSat.AddClauseC(Lits[:]) == 0 {
				assert(false, "0")
			}
		}
	} else {
		for i = 0; i < aig.Aig_ManCiNum(pFrames); i++ {
			pObj = aig.Aig_ManCi(pFrames, i)
			if i == aig.Aig_ManRegNum(pAig) {
				break
			}
			vec.Vec_IntPush(vVarsAB, pCnfFrames.VarNums[aig.Aig_ObjId(pObj)])

			pObj2 = saig.Saig_ManLi(pAig, i)
			Lits[0] = toLitCond(pCnfFrames.VarNums[aig.Aig_ObjId(pObj)], 0)
			Lits[1] = toLitCond(pCnfAig.VarNums[aig.Aig_ObjId(pObj2)], 1)
			if pSat.AddClauseC(Lits[:]) == 0 {
				assert(false, "0")
			}
			Lits[0] = toLitCond(pCnfFrames.VarNums[aig.Aig_ObjId(pObj)], 1)
			Lits[1] = toLitCond(pCnfAig.VarNums[aig.Aig_ObjId(pObj2)], 0)
			if pSat.AddClauseC(Lits[:]) == 0 {
				assert(false, "0")
			}
		}
	}
	// add clauses of B
	pSat.StoreMarkClausesAC()
	for i = 0; i < int32(len(pCnfFrames.Clauses)); i++ {
		if pSat.AddClauseC(pCnfFrames.Clauses[i]) == 0 {
			pSat.SetSolvedC()
			break
		}
	}
	pSat.StoreMarkRootsC()
	// return clauses to the original state
	cnf.Cnf_DataLift(pCnfAig, -int32(pCnfFrames.Variables))
	cnf.Cnf_DataLift(pCnfInter, -int32(pCnfFrames.Variables)-int32(pCnfAig.Variables))
	return pSat
}

/*
*Function*************************************************************

	Synopsis    [Performs one SAT run with interpolation.]

	Description [Returns 1 if proven. 0 if failed. -1 if undecided.]

**********************************************************************
*/
func Inter_ManPerformOneStep(p *Inter_Man_t, fUseBias, fUseBackward int32, nTimeNewOut abctime) int32 {
	var pSat *sat_solver
	var pSatCnf *bsat.Sto_Man_t
	var pManInterA *bsat.Inta_Man_t
	//    Intb_Man_t * pManInterB;
	var pGlobalVars []int32
	var status, RetValue int32
	var i, Var int32
	var clk abctime
	//    assert( p->pInterNew == NULL );

	// derive the SAT solver
	pSat = Inter_ManDeriveSatSolver(p.pInter, p.pCnfInter, p.pAigTrans, p.pCnfAig, p.pFrames, p.pCnfFrames, p.vVarsAB, fUseBackward)
	if pSat == nil {
		p.pInterNew = nil
		return 1
	}

	// set runtime limit
	if nTimeNewOut != 0 {
		pSat.SetRuntimeLimitC(util.Abc_Clock, nTimeNewOut)
	}

	// collect global variables
	pGlobalVars = make([]int32, pSat.NVarsC())
	for i = 0; i < int32(vec.Vec_IntSize(p.vVarsAB)); i++ {
		Var = vec.Vec_IntEntry(p.vVarsAB, int(i))
		pGlobalVars[Var] = 1
	}
	if fUseBias != 0 {
		pSat.SetGlobalVarsC(pGlobalVars)
	} else {
		pSat.SetGlobalVarsC(nil)
	}

	// solve the problem
	clk = util.Abc_Clock()
	status = pSat.SolveC(nil, int64(p.nConfLimit), 0, 0, 0)
	p.nConfCur = pSat.NConflicts()
	p.timeSat += util.Abc_Clock() - clk

	pSat.SetGlobalVarsC(nil)
	pGlobalVars = nil
	if status == int32(bsat.Unsatisfiable) {
		pSatCnf = pSat.StoreReleaseC()
		RetValue = 1
	} else if status == int32(bsat.Satisfiable) {
		RetValue = 0
	} else {
		RetValue = -1
	}
	pSat.Close()
	if pSatCnf == nil {
		return RetValue
	}

	// create the resulting manager
	clk = util.Abc_Clock()
	pManInterA = bsat.Inta_ManAlloc()
	p.pInterNew = bsat.Inta_ManInterpolate(pManInterA, pSatCnf, nTimeNewOut, vec.Vec_IntArray(p.vVarsAB), 0)
	bsat.Inta_ManFree(pManInterA)

	p.timeInt += util.Abc_Clock() - clk
	bsat.Sto_ManFree(pSatCnf)
	if p.pInterNew == nil {
		RetValue = -1
	}
	return RetValue
}
