package int

// Incremental inductive containment checking from proof/int/intCheck.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Inter_Check_t struct {
	nFramesK int32          // the number of timeframes (K=1 for simple induction)
	nVars    int32          // the current number of variables in the solver
	pFrames  *Aig_Man_t     // unrolled timeframes
	pCnf     *Cnf_Dat_t     // CNF of unrolled timeframes
	pSat     *sat_solver    // SAT solver
	vOrLits  *vec.Vec_Int_t // OR vars in each time frame (total number is the number nFrames)
	vAndLits *vec.Vec_Int_t // AND vars in the last timeframe (total number is the number of interpolants)
	vAssLits *vec.Vec_Int_t // assumptions (the union of the two)
}

/*
*Function*************************************************************

	Synopsis    [Create timeframes of the manager for interpolation.]

	Description [The resulting manager is combinational. The primary inputs
	corresponding to register outputs are ordered first.]

**********************************************************************
*/
func Inter_ManUnrollFrames(pAig *Aig_Man_t, nFrames int32) *Aig_Man_t {
	var pFrames *Aig_Man_t
	var pObj, pObjLi, pObjLo Aig_Obj_t
	var i, f int32
	assert(saig.Saig_ManRegNum(pAig) > 0, "Saig_ManRegNum(pAig) > 0")
	pFrames = aig.Aig_ManStart(aig.Aig_ManNodeNum(pAig) * nFrames)
	// map the constant node
	aig.Aig_ObjSetCopy(pAig, aig.Aig_ManConst1(pAig), aig.Aig_ManConst1(pFrames))
	// create variables for register outputs
	for i = 0; i < saig.Saig_ManRegNum(pAig); i++ {
		aig.Aig_ObjSetCopy(pAig, saig.Saig_ManLo(pAig, i), aig.Aig_ObjCreateCi(pFrames))
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
			aig.Aig_ObjCreateCo(pFrames, aig.Aig_ObjCopy(pAig, pObjLo))
		}
	}
	aig.Aig_ManCleanup(pFrames)
	return pFrames
}

/*
*Function*************************************************************

	Synopsis    [This procedure sets default values of interpolation parameters.]

**********************************************************************
*/
func Inter_CheckStart(pTrans *Aig_Man_t, nFramesK int32) *Inter_Check_t {
	var p *Inter_Check_t
	// create solver
	p = &Inter_Check_t{}
	p.vOrLits = vec.Vec_IntAlloc(100)
	p.vAndLits = vec.Vec_IntAlloc(100)
	p.vAssLits = vec.Vec_IntAlloc(100)
	// generate the timeframes
	p.pFrames = Inter_ManUnrollFrames(pTrans, nFramesK)
	assert(aig.Aig_ManCiNum(p.pFrames) == nFramesK*saig.Saig_ManPiNum(pTrans)+saig.Saig_ManRegNum(pTrans), "Aig_ManCiNum(p->pFrames) == nFramesK * Saig_ManPiNum(pTrans) + Saig_ManRegNum(pTrans)")
	assert(aig.Aig_ManCoNum(p.pFrames) == nFramesK*saig.Saig_ManRegNum(pTrans), "Aig_ManCoNum(p->pFrames) == nFramesK * Saig_ManRegNum(pTrans)")
	// convert to CNF
	p.pCnf = cnf.Cnf_Derive(p.pFrames, aig.Aig_ManCoNum(p.pFrames))
	p.pSat = cnf.Cnf_DataWriteIntoSolver(p.pCnf, 1, 0)
	// assign parameters
	p.nFramesK = nFramesK
	p.nVars = int32(p.pCnf.Variables)
	return p
}

/*
*Function*************************************************************

	Synopsis    [This procedure sets default values of interpolation parameters.]

**********************************************************************
*/
func Inter_CheckStop(p *Inter_Check_t) {
	if p == nil {
		return
	}
	cnf.Cnf_DataFree(p.pCnf)
	aig.Aig_ManStop(p.pFrames)
	if p.pSat != nil {
		p.pSat.Close()
	}
	*p = Inter_Check_t{}
}

/*
*Function*************************************************************

	Synopsis    [Creates one OR-gate: A + B = C.]

**********************************************************************
*/
func Inter_CheckAddOrGate(p *Inter_Check_t, iVarA, iVarB, iVarC int32) {
	var RetValue int32
	var pLits [3]bsat.Literal
	// add A => C   or   !A + C
	pLits[0] = toLitCond(iVarA, 1)
	pLits[1] = toLitCond(iVarC, 0)
	RetValue = p.pSat.AddClauseC(pLits[:2])
	assert(RetValue != 0, "RetValue")
	// add B => C   or   !B + C
	pLits[0] = toLitCond(iVarB, 1)
	pLits[1] = toLitCond(iVarC, 0)
	RetValue = p.pSat.AddClauseC(pLits[:2])
	assert(RetValue != 0, "RetValue")
	// add !A & !B => !C     or A + B + !C
	pLits[0] = toLitCond(iVarA, 0)
	pLits[1] = toLitCond(iVarB, 0)
	pLits[2] = toLitCond(iVarC, 1)
	RetValue = p.pSat.AddClauseC(pLits[:3])
	assert(RetValue != 0, "RetValue")
}

/*
*Function*************************************************************

	Synopsis    [Creates equality: A = B.]

**********************************************************************
*/
func Inter_CheckAddEqual(p *Inter_Check_t, iVarA, iVarB int32) {
	var RetValue int32
	var pLits [3]bsat.Literal
	// add A => B   or   !A + B
	pLits[0] = toLitCond(iVarA, 1)
	pLits[1] = toLitCond(iVarB, 0)
	RetValue = p.pSat.AddClauseC(pLits[:2])
	assert(RetValue != 0, "RetValue")
	// add B => A   or   !B + A
	pLits[0] = toLitCond(iVarB, 1)
	pLits[1] = toLitCond(iVarA, 0)
	RetValue = p.pSat.AddClauseC(pLits[:2])
	assert(RetValue != 0, "RetValue")
}

/*
*Function*************************************************************

	Synopsis    [Perform the checking.]

	Description [Returns 1 if the check has passed.]

**********************************************************************
*/
func Inter_CheckPerform(p *Inter_Check_t, pCnfInt *Cnf_Dat_t, nTimeNewOut abctime) int32 {
	var pObj, pObj2 Aig_Obj_t
	var i, f, VarA, VarB, RetValue, Entry, status int32
	pMan := pCnfInt.SourceGraph()
	nRegs := aig.Aig_ManCiNum(pMan)
	assert(aig.Aig_ManCoNum(p.pCnf.SourceGraph()) == p.nFramesK*nRegs, "Aig_ManCoNum(p->pCnf->pMan) == p->nFramesK * nRegs")
	assert(aig.Aig_ManCoNum(pMan) == 1, "Aig_ManCoNum(pCnfInt->pMan) == 1")

	// set runtime limit
	if nTimeNewOut != 0 {
		p.pSat.SetRuntimeLimitC(util.Abc_Clock, nTimeNewOut)
	}

	// add clauses to the SAT solver
	cnf.Cnf_DataLift(pCnfInt, p.nVars)
	for f = 0; f <= p.nFramesK; f++ {
		// add clauses to the solver
		for i = 0; i < int32(len(pCnfInt.Clauses)); i++ {
			RetValue = p.pSat.AddClauseC(pCnfInt.Clauses[i])
			assert(RetValue != 0, "RetValue")
		}
		// add equality clauses for the flop variables
		for i = 0; i < aig.Aig_ManCiNum(pMan); i++ {
			pObj = aig.Aig_ManCi(pMan, i)
			if f != 0 {
				pObj2 = aig.Aig_ManCo(p.pFrames, i+(f-1)*nRegs)
			} else {
				pObj2 = aig.Aig_ManCi(p.pFrames, i)
			}
			Inter_CheckAddEqual(p, pCnfInt.VarNums[aig.Aig_ObjId(pObj)], p.pCnf.VarNums[aig.Aig_ObjId(pObj2)])
		}
		// add final clauses
		if f < p.nFramesK {
			if f == int32(vec.Vec_IntSize(p.vOrLits)) { // find time here
				// add literal to this frame
				VarB = pCnfInt.VarNums[aig.Aig_ObjId(aig.Aig_ManCo(pMan, 0))]
				vec.Vec_IntPush(p.vOrLits, VarB)
			} else {
				// add OR gate for this frame
				VarA = vec.Vec_IntEntry(p.vOrLits, int(f))
				VarB = pCnfInt.VarNums[aig.Aig_ObjId(aig.Aig_ManCo(pMan, 0))]
				Inter_CheckAddOrGate(p, VarA, VarB, p.nVars+int32(pCnfInt.Variables))
				vec.Vec_IntWriteEntry(p.vOrLits, int(f), p.nVars+int32(pCnfInt.Variables)) // using var ID!
			}
		} else {
			// add AND gate for this frame
			VarB = pCnfInt.VarNums[aig.Aig_ObjId(aig.Aig_ManCo(pMan, 0))]
			vec.Vec_IntPush(p.vAndLits, VarB)
		}
		// update variable IDs
		cnf.Cnf_DataLift(pCnfInt, int32(pCnfInt.Variables)+1)
		p.nVars += int32(pCnfInt.Variables) + 1
	}
	cnf.Cnf_DataLift(pCnfInt, -p.nVars)
	assert(int32(vec.Vec_IntSize(p.vOrLits)) == p.nFramesK, "Vec_IntSize(p->vOrLits) == p->nFramesK")

	// collect the assumption literals
	vec.Vec_IntClear(p.vAssLits)
	for i = 0; i < int32(vec.Vec_IntSize(p.vOrLits)); i++ {
		Entry = vec.Vec_IntEntry(p.vOrLits, int(i))
		vec.Vec_IntPush(p.vAssLits, int32(toLitCond(Entry, 0)))
	}
	for i = 0; i < int32(vec.Vec_IntSize(p.vAndLits)); i++ {
		Entry = vec.Vec_IntEntry(p.vAndLits, int(i))
		vec.Vec_IntPush(p.vAssLits, int32(toLitCond(Entry, 1)))
	}
	/*
	   if ( pCnfInt->nLiterals == 3635 )
	   {
	       int s = 0;
	   }
	*/
	// call the SAT solver
	status = p.pSat.SolveC(vec.Vec_IntArray(p.vAssLits), 0, 0, 0, 0)

	return b2i(status == int32(bsat.Unsatisfiable))
}
