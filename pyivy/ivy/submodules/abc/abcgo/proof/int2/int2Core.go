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
// Int2_ManUnroll is unsupported: the checked-in C body refers to an
// undeclared AIG throughout. It does not establish an executable contract.
func Int2_ManUnroll(_ *Gia_Man_t, _ int32) *Gia_Man_t { return nil }

// Int2_GoUnroll is the Go-only reconstruction; it is not a C source port.
func Int2_GoUnroll(p *Gia_Man_t, nFrames int32) *Gia_Man_t {
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

// Int2_ManPreparePrefix is unsupported because the C body has undeclared
// locals and reads CNF from a different manager than the one it constructs.
func Int2_ManPreparePrefix(_ *Gia_Man_t, _ int32, pvCiMap **Vec_Int_t) *sat_solver {
	if pvCiMap != nil {
		*pvCiMap = nil
	}
	return nil
}

// Int2_GoPreparePrefix uses the Go-only unrolling reconstruction.
func Int2_GoPreparePrefix(p *Gia_Man_t, f int32, pvCiMap **Vec_Int_t) *sat_solver {
	assert(pvCiMap != nil, "CI map output is provided")
	pPref := Int2_GoUnroll(p, f)
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
// Int2_ManPrepareSuffix is unsupported: the C implementation refers to an
// undeclared SOP variable, so its clause mapping has no reliable contract.
// A nil solver and nil output pointers indicate that no suffix was produced.
func Int2_ManPrepareSuffix(_ *Gia_Man_t, _, _ *Vec_Int_t, pvCoMap **Vec_Int_t, ppSuff **Gia_Man_t) *sat_solver {
	if pvCoMap != nil {
		*pvCoMap = nil
	}
	if ppSuff != nil {
		*ppSuff = nil
	}
	return nil
}

/**Function*************************************************************

  Synopsis    [Returns the cube cover and status.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

// Int2_ManComputePreimage is unavailable: the C implementation references
// an undeclared manager and cannot establish a reliable mapping contract.
func Int2_ManComputePreimage(_ *Gia_Man_t, _, _ *sat_solver, _, _, _ *Vec_Int_t) (*Vec_Int_t, int32) {
	return nil, INT2_TIME_OUT
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
