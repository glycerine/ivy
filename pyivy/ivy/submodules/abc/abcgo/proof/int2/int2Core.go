package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [int2Core.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Core procedures.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 1, 2013.]

  Revision    [$Id: int2Core.c,v 1.00 2013/12/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "int2Int.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default values of interpolation parameters.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManUnroll(p *Gia_Man_t, nFrames int32) *Gia_Man_t {
	assert(p != nil && gia.Gia_ManRegNum(p) > 0, "GIA has registers")
	assert(nFrames >= 0, "frame count is nonnegative")
	pFrames := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	pFrames.CopyIdentityFrom(p)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	gia.Gia_ManHashAlloc(pFrames)
	for frame := int32(0); frame < nFrames; frame++ {
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			ro := gia.Gia_ManRo(p, i)
			value := int32(0)
			if frame > 0 {
				value = gia.Gia_ObjValue(gia.Gia_ObjRoToRi(p, ro))
			}
			gia.Gia_ObjSetValue(ro, value)
		}
		for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
			pi := gia.Gia_ManPi(p, i)
			gia.Gia_ObjSetValue(pi, gia.Gia_ManAppendCi(pFrames))
		}
		for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
			obj := gia.Gia_ManObj(p, i)
			if gia.Gia_ObjIsAnd(obj) != 0 {
				lit := int2HashCopyGate(pFrames, p, i)
				gia.Gia_ObjSetValue(obj, lit)
			}
		}
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			ri := gia.Gia_ManRi(p, i)
			gia.Gia_ObjSetValue(ri, gia.Gia_ObjFanin0Copy(ri))
		}
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		ri := gia.Gia_ManRi(p, i)
		value := int32(0)
		if nFrames > 0 {
			value = gia.Gia_ObjValue(ri)
		}
		gia.Gia_ManAppendCo(pFrames, value)
	}
	gia.Gia_ManHashStop(pFrames)
	temp := pFrames
	pFrames = gia.Gia_ManCleanup(temp)
	gia.Gia_ManStop(temp)
	return pFrames
}
func Int2_ManPreparePrefix(p *Gia_Man_t, f int32, pvCiMap **Vec_Int_t) *sat_solver {
	assert(pvCiMap != nil, "CI map output is provided")
	pPref := Int2_ManUnroll(p, f)
	pNew := gia.Jf_ManDeriveCnf(pPref, 0)
	cnf, ok := pNew.TakeData().(*Cnf_Dat_t)
	gia.Gia_ManStop(pPref)
	if !ok || cnf == nil {
		gia.Gia_ManStop(pNew)
		*pvCiMap = nil
		return nil
	}
	solver := sat_solver_from_clauses(int32(cnf.Variables), cnf.Clauses)
	if solver == nil {
		gia.Gia_ManStop(pNew)
		*pvCiMap = nil
		return nil
	}
	ciMap := vec.Vec_IntAlloc(int(gia.Gia_ManPoNum(pNew)))
	for i := int32(0); i < gia.Gia_ManPoNum(pNew); i++ {
		po := gia.Gia_ManPo(pNew, i)
		vec.Vec_IntPush(ciMap, cnf.VarNums[gia.Gia_ObjId(pNew, po)])
	}
	*pvCiMap = ciMap
	gia.Gia_ManStop(pNew)
	return solver
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Int2_ManPrepareSuffix(p *Gia_Man_t, vImageOne *Vec_Int_t, vImagesAll *Vec_Int_t, pvCoMap **Vec_Int_t, ppSuff **Gia_Man_t) *sat_solver {
	assert(pvCoMap != nil, "CO map output is provided")
	pSuff := Int2_ManProbToGia(p, vImageOne)
	assert(gia.Gia_ManPiNum(pSuff) == gia.Gia_ManCiNum(p), "suffix inputs match original CIs")
	pNew := gia.Jf_ManDeriveCnf(pSuff, 0)
	cnf, ok := pNew.TakeData().(*Cnf_Dat_t)
	gia.Gia_ManStop(pSuff)
	if !ok || cnf == nil {
		gia.Gia_ManStop(pNew)
		*pvCoMap = nil
		return nil
	}
	solver := sat_solver_from_clauses(int32(cnf.Variables), cnf.Clauses)
	if solver == nil {
		gia.Gia_ManStop(pNew)
		*pvCoMap = nil
		return nil
	}
	// JF encodes each CO but does not assert it; the suffix query asks for
	// assignments satisfying its single property/image output.
	output := gia.Gia_ManPo(pNew, 0)
	outputVar := cnf.VarNums[gia.Gia_ObjId(pNew, output)]
	if outputVar < 0 || sat_solver_addclause(solver, []int32{sat_literal(outputVar, 0)}) == 0 {
		gia.Gia_ManStop(pNew)
		*pvCoMap = nil
		return nil
	}
	coMap := vec.Vec_IntAlloc(int(gia.Gia_ManCiNum(pNew)))
	for i := int32(0); i < gia.Gia_ManCiNum(pNew); i++ {
		pi := gia.Gia_ManCi(pNew, i)
		variable := cnf.VarNums[gia.Gia_ObjId(pNew, pi)]
		if variable < 0 {
			variable = sat_solver_nvars(solver)
			sat_solver_setnvars(solver, variable+1)
		}
		vec.Vec_IntPush(coMap, variable)
	}
	// Image covers reserve entry zero; each subsequent cube contains state
	// literals, which must be translated to this solver's CI variables.
	for cursor := 1; cursor < vec.Vec_IntSize(vImagesAll); {
		count := vec.Vec_IntEntry(vImagesAll, cursor)
		cursor++
		if count < 0 || cursor+int(count) > vec.Vec_IntSize(vImagesAll) {
			panic("int2: malformed accumulated image cubes")
		}
		clause := make([]int32, count)
		for k := range clause {
			lit := vec.Vec_IntEntry(vImagesAll, cursor)
			state := lit >> 1
			if lit < 0 || state >= gia.Gia_ManRegNum(p) {
				panic("int2: image cube state index out of range")
			}
			variable := vec.Vec_IntEntry(coMap, int(gia.Gia_ManPiNum(p)+state))
			clause[k] = sat_literal(variable, (lit&1)^1)
			cursor++
		}
		if sat_solver_addclause(solver, clause) == 0 {
			gia.Gia_ManStop(pNew)
			*pvCoMap = nil
			return nil
		}
	}
	*pvCoMap = coMap
	if ppSuff != nil {
		*ppSuff = pNew
	} else {
		gia.Gia_ManStop(pNew)
	}
	return solver
}

/**Function*************************************************************

  Synopsis    [Returns the cube cover and status.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

func Int2_ManComputePreimage(pSuff *Gia_Man_t, pSatPref, pSatSuff *sat_solver, vCiMap, vCoMap, vPrio *Vec_Int_t) (*Vec_Int_t, int32) {
	assert(pSuff != nil && pSatPref != nil && pSatSuff != nil, "preimage managers are initialized")
	assert(vCiMap != nil && vCoMap != nil && vPrio != nil, "preimage maps and priorities are initialized")
	assert(vec.Vec_IntSize(vCoMap) == vec.Vec_IntSize(vPrio), "suffix CI map and priority sizes match")
	assert(vec.Vec_IntSize(vCiMap) <= int(gia.Gia_ManCiNum(pSuff)), "prefix map fits suffix state variables")
	firstStateCI := gia.Gia_ManCiNum(pSuff) - int32(vec.Vec_IntSize(vCiMap))
	image := vec.Vec_IntAlloc(1000)
	vec.Vec_IntPush(image, 0) // reserved SOP header
	for {
		status := sat_solver_solve(pSatSuff, nil, 0, 0, 0, 0)
		if status == l_Undef {
			return nil, INT2_TIME_OUT
		}
		if status == l_False {
			return image, INT2_COMPUTED
		}
		assert(status == l_True, "suffix SAT result is recognized")

		assignment := vec.Vec_IntAlloc(vec.Vec_IntSize(vCoMap))
		for i := 0; i < vec.Vec_IntSize(vCoMap); i++ {
			variable := vec.Vec_IntEntry(vCoMap, i)
			value := sat_solver_var_value(pSatSuff, variable)
			// An unassigned model variable is a don't-care; zero is a valid
			// deterministic completion for cube minimization.
			if value == l_True {
				vec.Vec_IntPush(assignment, 1)
			} else {
				vec.Vec_IntPush(assignment, 0)
			}
		}
		prioritized := Int2_ManRefineCube(pSuff, assignment, vPrio)

		assumptions := make([]int32, 0, vec.Vec_IntSize(prioritized))
		for i := 0; i < vec.Vec_IntSize(prioritized); i++ {
			ci := vec.Vec_IntEntry(prioritized, i)
			assert(ci >= firstStateCI && ci < int32(vec.Vec_IntSize(vCoMap)), "selected CI is a state variable")
			stateIndex := ci - firstStateCI
			variable := vec.Vec_IntEntry(vCiMap, int(stateIndex))
			value := vec.Vec_IntEntry(assignment, int(ci))
			assumptions = append(assumptions, sat_literal(variable, 1-value))
		}
		vec.Vec_IntFree(prioritized)
		vec.Vec_IntFree(assignment)
		status = sat_solver_solve(pSatPref, assumptions, 0, 0, 0, 0)
		if status == l_False {
			var core []int32
			sat_solver_final(pSatPref, &core)
			clause := make([]int32, 0, len(core))
			vec.Vec_IntPush(image, int32(len(core)))
			for _, literal := range core {
				stateIndex := -1
				for i := 0; i < vec.Vec_IntSize(vCiMap); i++ {
					if vec.Vec_IntEntry(vCiMap, i) == literal>>1 {
						stateIndex = i
						break
					}
				}
				assert(stateIndex >= 0, "core literal maps to a state variable")
				suffixVariable := vec.Vec_IntEntry(vCoMap, int(firstStateCI)+stateIndex)
				clause = append(clause, sat_literal(suffixVariable, literal&1))
				vec.Vec_IntPush(image, sat_literal(int32(stateIndex), (literal&1)^1))
			}
			if sat_solver_addclause(pSatSuff, clause) == 0 {
				return image, INT2_COMPUTED
			}
			continue
		}
		if status == l_Undef {
			return nil, INT2_TIME_OUT
		}
		if status == l_True {
			return nil, INT2_FALSE_NEG
		}
		assert(false, "prefix SAT result is recognized")
	}
}

/*
*Function*************************************************************

	Synopsis    [Interpolates while the number of conflicts is not exceeded.]

	Description [Returns 1 if proven. 0 if failed. -1 if undecided.]

	SideEffects [Does not check the property in 0-th frame.]

	SeeAlso     []

**********************************************************************
*/
// Int2_ManPerformInterpolation returns undecided because the checked-in C
// interpolation driver is incomplete. See SOURCE_FIDELITY.md for details.
func Int2_ManPerformInterpolation(_ *Gia_Man_t, _ *Int2_ManPars_t) int32 {
	return INT2_TIME_OUT
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
