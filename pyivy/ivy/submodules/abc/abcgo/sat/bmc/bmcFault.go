package bmc

/**CFile****************************************************************

  FileName    [bmcFault.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [Checking for functional faults.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcFault.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

// Line-by-line port of src/sat/bmc/bmcFault.c. The SAT solver is driven
// through the sat_solver_* bridges of bsat (AddClauseC, AddAndC, ...), so
// that clause order, variable allocation and failed-clause behaviour are C's.

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const FFTEST_MAX_VARS = 2
const FFTEST_MAX_PARS = 8

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

func faultB2i(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// Gia_DeriveFormula_rec writes the formula of iLit using the CI names.
func Gia_DeriveFormula_rec(pGia *gia.Gia_Man_t, ppNamesIn []string, vStr *[]byte, iLit int32) {
	pObj := gia.Gia_ManObj(pGia, util.Abc_Lit2Var(iLit))
	fCompl := util.Abc_LitIsCompl(iLit)
	if gia.Gia_ObjIsAnd(pObj) != 0 {
		*vStr = append(*vStr, '(')
		if gia.Gia_ObjIsMux(pGia, pObj) != 0 {
			Gia_DeriveFormula_rec(pGia, ppNamesIn, vStr, gia.Gia_ObjFaninLit0p(pGia, pObj))
			*vStr = append(*vStr, '?')
			Gia_DeriveFormula_rec(pGia, ppNamesIn, vStr, util.Abc_LitNotCond(gia.Gia_ObjFaninLit1p(pGia, pObj), fCompl))
			*vStr = append(*vStr, ':')
			Gia_DeriveFormula_rec(pGia, ppNamesIn, vStr, util.Abc_LitNotCond(gia.Gia_ObjFaninLit2p(pGia, pObj), fCompl))
		} else {
			Gia_DeriveFormula_rec(pGia, ppNamesIn, vStr, util.Abc_LitNotCond(gia.Gia_ObjFaninLit0p(pGia, pObj), fCompl))
			if gia.Gia_ObjIsXor(pObj) != 0 {
				*vStr = append(*vStr, '^')
			} else if fCompl != 0 {
				*vStr = append(*vStr, '|')
			} else {
				*vStr = append(*vStr, '&')
			}
			Gia_DeriveFormula_rec(pGia, ppNamesIn, vStr, util.Abc_LitNotCond(gia.Gia_ObjFaninLit1p(pGia, pObj), fCompl))
		}
		*vStr = append(*vStr, ')')
	} else {
		if fCompl != 0 {
			*vStr = append(*vStr, '~')
		}
		*vStr = append(*vStr, ppNamesIn[gia.Gia_ObjCioId(pObj)]...)
	}
}

// Gia_DeriveFormula returns the formula of the first CO.
func Gia_DeriveFormula(pGia *gia.Gia_Man_t, ppNamesIn []string) string {
	vStr := make([]byte, 0, 1000)
	pMuxes := gia.Gia_ManDupMuxes(pGia, 2)
	pObj := gia.Gia_ManCo(pGia, 0)
	vStr = append(vStr, '(')
	Gia_DeriveFormula_rec(pGia, ppNamesIn, &vStr, gia.Gia_ObjFaninLit0p(pGia, pObj))
	vStr = append(vStr, ')')
	gia.Gia_ManStop(pMuxes)
	return string(vStr)
}

/*
*Function*************************************************************

	Synopsis    [This procedure sets default parameters.]

**********************************************************************
*/
func Gia_ParFfSetDefault(p *Bmc_ParFf_t) {
	*p = Bmc_ParFf_t{}
}

/**Function*************************************************************

  Synopsis    [Adds a general cardinality constraint in terms of vVars.]

  Description [Strict is "exactly K out of N". Non-strict is
  "less or equal than K out of N".]

***********************************************************************/

func Cnf_AddSorder(p *bsat.Solver, pVars []int32, i, k int32, pnVars *int32) {
	iVar1 := *pnVars
	*pnVars++
	iVar2 := *pnVars
	*pnVars++
	p.AddAndC(iVar1, pVars[i], pVars[k], 1, 1, 1)
	p.AddAndC(iVar2, pVars[i], pVars[k], 0, 0, 0)
	pVars[i] = iVar1
	pVars[k] = iVar2
}
func Cnf_AddCardinConstrMerge(p *bsat.Solver, pVars []int32, lo, hi, r int32, pnVars *int32) {
	step := r * 2
	if step < hi-lo {
		Cnf_AddCardinConstrMerge(p, pVars, lo, hi-r, step, pnVars)
		Cnf_AddCardinConstrMerge(p, pVars, lo+r, hi, step, pnVars)
		for i := lo + r; i < hi-r; i += step {
			Cnf_AddSorder(p, pVars, i, i+r, pnVars)
		}
	}
}
func Cnf_AddCardinConstrRange(p *bsat.Solver, pVars []int32, lo, hi int32, pnVars *int32) {
	if hi-lo >= 1 {
		mid := lo + (hi-lo)/2
		for i := lo; i <= mid; i++ {
			Cnf_AddSorder(p, pVars, i, i+(hi-lo+1)/2, pnVars)
		}
		Cnf_AddCardinConstrRange(p, pVars, lo, mid, pnVars)
		Cnf_AddCardinConstrRange(p, pVars, mid+1, hi, pnVars)
		Cnf_AddCardinConstrMerge(p, pVars, lo, hi, 1, pnVars)
	}
}
func Cnf_AddCardinConstrPairWise(p *bsat.Solver, vVars *vec.Vec_Int_t, K, fStrict int32) {
	nVars := p.NumVars()
	nSizeOld := int32(vec.Vec_IntSize(vVars))
	assert(nSizeOld >= 2, "nSizeOld >= 2")
	// check that vars are ok
	for _, iVar := range vec.Vec_IntArray(vVars) {
		assert(iVar >= 0 && iVar < nVars, "iVar >= 0 && iVar < nVars")
	}
	// make the size a degree of two
	nSize := int32(1)
	for nSize < nSizeOld {
		nSize *= 2
	}
	// extend
	p.SetNumVars(nVars + 1 + nSize*nSize/2)
	if nSize != nSizeOld {
		// fill in with const0
		vec.Vec_IntFillExtra(vVars, int(nSize), nVars)
		// add const0 variable
		Lit := util.Abc_Var2Lit(nVars, 1)
		nVars++
		if p.AddClauseC([]bsat.Literal{Lit}) == 0 {
			assert(false, "0")
		}
	}
	// construct recursively
	Cnf_AddCardinConstrRange(p, vec.Vec_IntArray(vVars), 0, nSize-1, &nVars)
	// add final constraint
	assert(K > 0 && K < nSizeOld, "K > 0 && K < nSizeOld")
	Lit := util.Abc_Var2Lit(vec.Vec_IntEntry(vVars, int(K)), 1)
	if p.AddClauseC([]bsat.Literal{Lit}) == 0 {
		assert(false, "0")
	}
	if fStrict != 0 {
		Lit = util.Abc_Var2Lit(vec.Vec_IntEntry(vVars, int(K-1)), 0)
		if p.AddClauseC([]bsat.Literal{Lit}) == 0 {
			assert(false, "0")
		}
	}
	// return to the old size
	vec.Vec_IntShrink(vVars, 0) // make it unusable
}

func Cnf_AddCardinVar(Level, Base int32, vVars *vec.Vec_Int_t, k int32) int32 {
	if Level != 0 {
		return Base + k
	}
	return vec.Vec_IntEntry(vVars, int(k))
}
func Cnf_AddCardinConstrGeneral(p *bsat.Solver, vVars *vec.Vec_Int_t, K, fStrict int32) {
	nVars := p.NumVars()
	nBits := int32(vec.Vec_IntSize(vVars))
	for _, iVar := range vec.Vec_IntArray(vVars) {
		assert(iVar >= 0 && iVar < nVars, "iVar >= 0 && iVar < nVars")
	}
	p.SetNumVars(nVars + nBits*nBits)
	var k int32
	for i := int32(0); i < nBits; i++ {
		iCur := nVars + nBits*(i-1)
		iNext := nVars + nBits*i
		if i&1 != 0 {
			p.AddBufferC(iNext, Cnf_AddCardinVar(i, iCur, vVars, 0), 0)
		}
		for k = i & 1; k+1 < nBits; k += 2 {
			p.AddAndC(iNext+k, Cnf_AddCardinVar(i, iCur, vVars, k), Cnf_AddCardinVar(i, iCur, vVars, k+1), 1, 1, 1)
			p.AddAndC(iNext+k+1, Cnf_AddCardinVar(i, iCur, vVars, k), Cnf_AddCardinVar(i, iCur, vVars, k+1), 0, 0, 0)
		}
		if k == nBits-1 {
			p.AddBufferC(iNext+nBits-1, Cnf_AddCardinVar(i, iCur, vVars, nBits-1), 0)
		}
	}
	// add final constraint
	assert(K > 0 && K < nBits, "K > 0 && K < nBits")
	Lit := util.Abc_Var2Lit(nVars+nBits*(nBits-1)+K, 1)
	if p.AddClauseC([]bsat.Literal{Lit}) == 0 {
		assert(false, "0")
	}
	if fStrict != 0 {
		Lit = util.Abc_Var2Lit(nVars+nBits*(nBits-1)+K-1, 0)
		if p.AddClauseC([]bsat.Literal{Lit}) == 0 {
			assert(false, "0")
		}
	}
}

// Cnf_AddCardinConstr adds the constraint that no more than 1 variable is 1.
func Cnf_AddCardinConstr(p *bsat.Solver, vVars *vec.Vec_Int_t) int32 {
	nVars := p.NumVars()
	for _, iVar := range vec.Vec_IntArray(vVars) {
		assert(iVar >= 0 && iVar < nVars, "iVar >= 0 && iVar < nVars")
	}
	iVar := nVars
	p.SetNumVars(nVars + int32(vec.Vec_IntSize(vVars)) - 1)
	for vec.Vec_IntSize(vVars) > 1 {
		k := 0
		for i := 0; i < vec.Vec_IntSize(vVars)/2; i++ {
			pLits := []bsat.Literal{
				util.Abc_Var2Lit(vec.Vec_IntEntry(vVars, 2*i), 1),
				util.Abc_Var2Lit(vec.Vec_IntEntry(vVars, 2*i+1), 1),
			}
			p.AddClauseC(pLits)
			p.AddAndC(iVar, vec.Vec_IntEntry(vVars, 2*i), vec.Vec_IntEntry(vVars, 2*i+1), 1, 1, 1)
			vec.Vec_IntWriteEntry(vVars, k, iVar)
			k++
			iVar++
		}
		if vec.Vec_IntSize(vVars)&1 != 0 {
			vec.Vec_IntWriteEntry(vVars, k, vec.Vec_IntEntryLast(vVars))
			k++
		}
		vec.Vec_IntShrink(vVars, k)
	}
	return iVar
}
func Cnf_AddCardinConstrTest() {
	Count, nVars := 1, int32(8)
	vVars := vec.Vec_IntStartNatural(int(nVars))
	pSat := bsat.NewSolver()
	pSat.SetNumVars(nVars)
	//Cnf_AddCardinConstr( pSat, vVars );
	Cnf_AddCardinConstrPairWise(pSat, vVars, 2, 1)
	for {
		status := pSat.SolveWithResourceLimits(nil, 0, 0, 0, 0)
		if status != bsat.Satisfiable {
			break
		}
		vec.Vec_IntClear(vVars)
		fmt.Printf("%3d : ", Count)
		Count++
		for i := int32(0); i < nVars; i++ {
			vec.Vec_IntPush(vVars, util.Abc_Var2Lit(i, faultB2i(pSat.VarValue(i))))
			fmt.Printf("%d", faultB2i(pSat.VarValue(i)))
		}
		fmt.Printf("\n")
		if pSat.AddClauseC(vec.Vec_IntArray(vVars)) == 0 {
			break
		}
	}
	pSat.Delete()
}

// Cnf_DeriveGiaRemapped derives the CNF of a GIA through Gia_ManToAigSimple.
func Cnf_DeriveGiaRemapped(p *gia.Gia_Man_t) *cnf.Cnf_Dat_t {
	pAig := gia.Gia_ManToAigSimple(p)
	aig.Aig_ManSetRegNumForConstruction(pAig, 0) // pAig->nRegs = 0;
	pCnf := cnf.Cnf_Derive(pAig, aig.Aig_ManCoNum(pAig))
	aig.Aig_ManStop(pAig)
	return pCnf
	//    return (Cnf_Dat_t *)Mf_ManGenerateCnf( p, 8, 0, 0, 0, 0 );
}

// Cnf_DataLiftGia shifts the CNF variables by nVarsPlus.
func Cnf_DataLiftGia(p *cnf.Cnf_Dat_t, pGia *gia.Gia_Man_t, nVarsPlus int32) {
	cnf.Cnf_DataLift(p, nVarsPlus)
}

func Gia_ManFaultUnfold(p *gia.Gia_Man_t, fUseMuxes, fFfOnly int32) *gia.Gia_Man_t {
	pNew := gia.Gia_ManStart((2 + 3*fUseMuxes) * gia.Gia_ManObjNum(p))
	gia.Gia_ManSetNameSpec(pNew, gia.Gia_ManName(p), "")
	gia.Gia_ManHashAlloc(pNew)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	// add first timeframe
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManRo(p, i), gia.Gia_ManAppendCi(pNew))
	}
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManPi(p, i), gia.Gia_ManAppendCi(pNew))
	}
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		pObj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(pObj) != 0 {
			gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))
		}
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		pObj := gia.Gia_ManCo(p, i)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ObjFanin0Copy(pObj))
	}
	// add second timeframe
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		pObj := gia.Gia_ManRo(p, i)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ObjValue(gia.Gia_ObjRoToRi(p, pObj)))
	}
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManPi(p, i), gia.Gia_ManAppendCi(pNew))
	}
	if fFfOnly != 0 {
		for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
			pObj := gia.Gia_ManObj(p, i)
			if gia.Gia_ObjIsAnd(pObj) != 0 {
				gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))
			}
		}
		for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
			pObj := gia.Gia_ManPo(p, i)
			gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
		}
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			pObj := gia.Gia_ManRi(p, i)
			iCtrl := gia.Gia_ManAppendCi(pNew)
			iThis := gia.Gia_ObjFanin0Copy(pObj)
			if fUseMuxes != 0 {
				gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashMux(pNew, iCtrl, gia.Gia_ObjValue(pObj), iThis))
			} else {
				gia.Gia_ObjSetValue(pObj, iThis)
			}
			gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjValue(pObj)))
		}
	} else {
		for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
			pObj := gia.Gia_ManObj(p, i)
			if gia.Gia_ObjIsAnd(pObj) == 0 {
				continue
			}
			iCtrl := gia.Gia_ManAppendCi(pNew)
			iThis := gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj))
			if fUseMuxes != 0 {
				gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashMux(pNew, iCtrl, gia.Gia_ObjValue(pObj), iThis))
			} else {
				gia.Gia_ObjSetValue(pObj, iThis)
			}
		}
		for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
			pObj := gia.Gia_ManCo(p, i)
			gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
		}
	}
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManStop(pTemp)
	nCtrl := gia.Gia_ManAndNum(p)
	if fFfOnly != 0 {
		nCtrl = gia.Gia_ManRegNum(p)
	}
	assert(gia.Gia_ManPiNum(pNew) == gia.Gia_ManRegNum(p)+2*gia.Gia_ManPiNum(p)+nCtrl, "Gia_ManPiNum(pNew) == ...")
	return pNew
}

func Gia_ManStuckAtUnfold(p *gia.Gia_Man_t, vMap *vec.Vec_Int_t) *gia.Gia_Man_t {
	iFuncVars := 0
	pNew := gia.Gia_ManStart(3 * gia.Gia_ManObjNum(p))
	gia.Gia_ManSetNameSpec(pNew, gia.Gia_ManName(p), "")
	gia.Gia_ManHashAlloc(pNew)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManCi(p, i), gia.Gia_ManAppendCi(pNew))
	}
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		pObj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(pObj) == 0 {
			continue
		}
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))

		if vec.Vec_IntEntry(vMap, iFuncVars) != 0 {
			iFuncVars++
			gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashAnd(pNew, util.Abc_LitNot(gia.Gia_ManAppendCi(pNew)), gia.Gia_ObjValue(pObj)))
		} else {
			iFuncVars++
			gia.Gia_ManAppendCi(pNew)
		}

		if vec.Vec_IntEntry(vMap, iFuncVars) != 0 {
			iFuncVars++
			gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashOr(pNew, gia.Gia_ManAppendCi(pNew), gia.Gia_ObjValue(pObj)))
		} else {
			iFuncVars++
			gia.Gia_ManAppendCi(pNew)
		}
	}
	assert(iFuncVars == vec.Vec_IntSize(vMap), "iFuncVars == Vec_IntSize(vMap)")
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		pObj := gia.Gia_ManCo(p, i)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
	}
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManStop(pTemp)
	assert(gia.Gia_ManPiNum(pNew) == gia.Gia_ManCiNum(p)+2*gia.Gia_ManAndNum(p), "Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + 2 * Gia_ManAndNum(p)")
	return pNew
}

func Gia_ManFlipUnfold(p *gia.Gia_Man_t, vMap *vec.Vec_Int_t) *gia.Gia_Man_t {
	iFuncVars := 0
	pNew := gia.Gia_ManStart(4 * gia.Gia_ManObjNum(p))
	gia.Gia_ManSetNameSpec(pNew, gia.Gia_ManName(p), "")
	gia.Gia_ManHashAlloc(pNew)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManCi(p, i), gia.Gia_ManAppendCi(pNew))
	}
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		pObj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(pObj) == 0 {
			continue
		}
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))
		if vec.Vec_IntEntry(vMap, iFuncVars) != 0 {
			iFuncVars++
			gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashXor(pNew, gia.Gia_ManAppendCi(pNew), gia.Gia_ObjValue(pObj)))
		} else {
			iFuncVars++
			gia.Gia_ManAppendCi(pNew)
		}
	}
	assert(iFuncVars == vec.Vec_IntSize(vMap), "iFuncVars == Vec_IntSize(vMap)")
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		pObj := gia.Gia_ManCo(p, i)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
	}
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManStop(pTemp)
	assert(gia.Gia_ManPiNum(pNew) == gia.Gia_ManCiNum(p)+gia.Gia_ManAndNum(p), "Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + Gia_ManAndNum(p)")
	return pNew
}

func Gia_ManFOFUnfold(p *gia.Gia_Man_t, vMap *vec.Vec_Int_t) *gia.Gia_Man_t {
	var iCtrl0, iCtrl1, iCtrl2, iCtrl3, iMuxA, iMuxB int32
	iFuncVars := int32(0)
	VarLimit := 4 * gia.Gia_ManAndNum(p)
	pNew := gia.Gia_ManStart(9 * gia.Gia_ManObjNum(p))
	gia.Gia_ManSetNameSpec(pNew, gia.Gia_ManName(p), "")
	gia.Gia_ManHashAlloc(pNew)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManCi(p, i), gia.Gia_ManAppendCi(pNew))
	}
	// C: if ( Vec_IntEntry(vMap, iFuncVars++) && iFuncVars < VarLimit )
	ctrl := func() int32 {
		e := vec.Vec_IntEntry(vMap, int(iFuncVars))
		iFuncVars++
		if e != 0 && iFuncVars < VarLimit {
			return gia.Gia_ManAppendCi(pNew)
		}
		gia.Gia_ManAppendCi(pNew)
		return 0
	}
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		pObj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(pObj) == 0 {
			continue
		}
		iCtrl0 = ctrl()
		iCtrl1 = ctrl()
		iCtrl2 = ctrl()
		iCtrl3 = ctrl()

		if gia.Gia_ObjFaninC0(pObj) != 0 && gia.Gia_ObjFaninC1(pObj) != 0 {
			iCtrl0 = util.Abc_LitNot(iCtrl0)
		} else if gia.Gia_ObjFaninC0(pObj) == 0 && gia.Gia_ObjFaninC1(pObj) != 0 {
			iCtrl1 = util.Abc_LitNot(iCtrl1)
		} else if gia.Gia_ObjFaninC0(pObj) != 0 && gia.Gia_ObjFaninC1(pObj) == 0 {
			iCtrl2 = util.Abc_LitNot(iCtrl2)
		} else { //if ( !Gia_ObjFaninC0(pObj) && !Gia_ObjFaninC1(pObj) )
			iCtrl3 = util.Abc_LitNot(iCtrl3)
		}

		iMuxA = gia.Gia_ManHashMux(pNew, gia.Gia_ObjValue(gia.Gia_ObjFanin0(pObj)), iCtrl1, iCtrl0)
		iMuxB = gia.Gia_ManHashMux(pNew, gia.Gia_ObjValue(gia.Gia_ObjFanin0(pObj)), iCtrl3, iCtrl2)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashMux(pNew, gia.Gia_ObjValue(gia.Gia_ObjFanin1(pObj)), iMuxB, iMuxA))
	}
	assert(int(iFuncVars) == vec.Vec_IntSize(vMap), "iFuncVars == Vec_IntSize(vMap)")
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		pObj := gia.Gia_ManCo(p, i)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
	}
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManStop(pTemp)
	assert(gia.Gia_ManPiNum(pNew) == gia.Gia_ManCiNum(p)+4*gia.Gia_ManAndNum(p), "Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + 4 * Gia_ManAndNum(p)")
	return pNew
}

// faultStr returns the byte of a C string at i (0 past the end).
func faultStr(s []byte, i int) byte {
	if i < 0 || i >= len(s) {
		return 0
	}
	return s[i]
}

func Gia_FormStrCount(pStr string, pnVars, pnPars *int32) int32 {
	Counter := 0
	s := []byte(pStr)
	if faultStr(s, 0) != '(' {
		fmt.Printf("The first symbol should be the opening parenthesis \"(\".\n")
		return 1
	}
	if len(s) == 0 || s[len(s)-1] != ')' {
		fmt.Printf("The last symbol should be the closing parenthesis \")\".\n")
		return 1
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '(' {
			Counter++
		} else if s[i] == ')' {
			Counter--
		}
	}
	if Counter != 0 {
		fmt.Printf("The number of opening and closing parentheses is not equal.\n")
		return 1
	}
	*pnVars = 0
	*pnPars = 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'b' {
			*pnVars = util.Abc_MaxInt(*pnVars, int32(c-'a'+1))
		} else if c >= 'p' && c <= 's' {
			*pnPars = util.Abc_MaxInt(*pnPars, int32(c-'p'+1))
		} else if c == '(' || c == ')' {
		} else if c == '&' || c == '|' || c == '^' {
		} else if c == '?' || c == ':' {
		} else if c == '~' {
			if next := faultStr(s, i+1); next < 'a' || next > 'z' {
				fmt.Printf("Expecting alphabetic symbol (instead of \"%c\") after negation (~)\n", next)
				return 1
			}
		} else {
			fmt.Printf("Unknown symbol (%c) in the formula (%s)\n", c, pStr)
			return 1
		}
	}
	if *pnVars != FFTEST_MAX_VARS {
		fmt.Printf("The number of input variables (%d) should be 2\n", *pnVars)
		return 1
	}
	if *pnPars < 1 || *pnPars > FFTEST_MAX_PARS {
		fmt.Printf("The number of parameters should be between 1 and %d\n", FFTEST_MAX_PARS)
		return 1
	}
	return 0
}

// Gia_FormStrTransform replaces "~x" by the upper-case letter X.
func Gia_FormStrTransform(pForm string) string {
	pStr := make([]byte, 0, len(pForm))
	for i := 0; i < len(pForm); i++ {
		if pForm[i] == '~' {
			i++
			assert(i < len(pForm) && pForm[i] >= 'a' && pForm[i] <= 'z', "pForm[i] >= 'a' && pForm[i] <= 'z'")
			pStr = append(pStr, 'A'+pForm[i]-'a')
		} else {
			pStr = append(pStr, pForm[i])
		}
	}
	return string(pStr)
}

// Gia_ManFormulaEndToken returns the index just past the token starting at beg.
func Gia_ManFormulaEndToken(pForm string, beg int) int {
	Counter := 0
	for pThis := beg; pThis < len(pForm); pThis++ {
		assert(pForm[pThis] != '~', "*pThis != '~'")
		if pForm[pThis] == '(' {
			Counter++
		} else if pForm[pThis] == ')' {
			Counter--
		}
		if Counter == 0 {
			return pThis + 1
		}
	}
	assert(false, "0")
	return -1
}
func Gia_ManPrintFormula_rec(s string, pBeg, pEnd int) {
	if pBeg+1 == pEnd {
		c := s[pBeg]
		if c >= 'a' && c <= 'b' {
			fmt.Printf("%c", c)
		} else if c >= 'A' && c <= 'B' {
			fmt.Printf("~%c", c-'A'+'a')
		} else if c >= 'p' && c <= 'w' { // pqrstuvw
			fmt.Printf("%c", c)
		} else if c >= 'P' && c <= 'W' {
			fmt.Printf("~%c", c-'A'+'a')
		}
		return
	}
	if s[pBeg] == '(' {
		pEndNew := Gia_ManFormulaEndToken(s, pBeg)
		if pEndNew == pEnd {
			assert(s[pBeg] == '(', "pBeg[0] == '('")
			assert(s[pEnd-1] == ')', "pBeg[pEnd-pBeg-1] == ')'")
			Gia_ManPrintFormula_rec(s, pBeg+1, pEnd-1)
			return
		}
	}
	// get first part
	pEndNew := Gia_ManFormulaEndToken(s, pBeg)
	fmt.Printf("(")
	Gia_ManPrintFormula_rec(s, pBeg, pEndNew)
	fmt.Printf(")")
	Oper := faultStr([]byte(s), pEndNew)
	// derive the formula
	if Oper == '&' {
		fmt.Printf("&")
	} else if Oper == '|' {
		fmt.Printf("|")
	} else if Oper == '^' {
		fmt.Printf("^")
	} else if Oper == '?' {
		fmt.Printf("?")
	} else {
		assert(false, "0")
	}
	// get second part
	pBeg = pEndNew + 1
	pEndNew = Gia_ManFormulaEndToken(s, pBeg)
	fmt.Printf("(")
	Gia_ManPrintFormula_rec(s, pBeg, pEndNew)
	fmt.Printf(")")
	if Oper == '?' {
		fmt.Printf(":")
		// get third part
		assert(faultStr([]byte(s), pEndNew) == ':', "pEndNew[0] == ':'")
		pBeg = pEndNew + 1
		pEndNew = Gia_ManFormulaEndToken(s, pBeg)
		fmt.Printf("(")
		Gia_ManPrintFormula_rec(s, pBeg, pEndNew)
		fmt.Printf(")")
	}
}
func Gia_ManPrintFormula(pStr string) {
	fmt.Printf("Using formula: ")
	fmt.Printf("(")
	Gia_ManPrintFormula_rec(pStr, 0, len(pStr))
	fmt.Printf(")")
	fmt.Printf("\n")
}

// Gia_ManRealizeFormula_rec implements the fault model formula using
// functional/parameter vars.
func Gia_ManRealizeFormula_rec(p *gia.Gia_Man_t, pVars, pPars []int32, s string, pBeg, pEnd int, nPars int32) int32 {
	var iFans [3]int32
	if pBeg+1 == pEnd {
		c := s[pBeg]
		if c >= 'a' && c <= 'b' {
			return pVars[c-'a']
		}
		if c >= 'A' && c <= 'B' {
			return util.Abc_LitNot(pVars[c-'A'])
		}
		if c >= 'p' && c <= 'w' { // pqrstuvw
			return pPars[c-'p']
		}
		if c >= 'P' && c <= 'W' {
			return util.Abc_LitNot(pPars[c-'P'])
		}
		assert(false, "0")
		return -1
	}
	if s[pBeg] == '(' {
		pEndNew := Gia_ManFormulaEndToken(s, pBeg)
		if pEndNew == pEnd {
			assert(s[pBeg] == '(', "pBeg[0] == '('")
			assert(s[pEnd-1] == ')', "pBeg[pEnd-pBeg-1] == ')'")
			return Gia_ManRealizeFormula_rec(p, pVars, pPars, s, pBeg+1, pEnd-1, nPars)
		}
	}
	// get first part
	pEndNew := Gia_ManFormulaEndToken(s, pBeg)
	iFans[0] = Gia_ManRealizeFormula_rec(p, pVars, pPars, s, pBeg, pEndNew, nPars)
	Oper := faultStr([]byte(s), pEndNew)
	// get second part
	pBeg = pEndNew + 1
	pEndNew = Gia_ManFormulaEndToken(s, pBeg)
	iFans[1] = Gia_ManRealizeFormula_rec(p, pVars, pPars, s, pBeg, pEndNew, nPars)
	// derive the formula
	if Oper == '&' {
		return gia.Gia_ManHashAnd(p, iFans[0], iFans[1])
	}
	if Oper == '|' {
		return gia.Gia_ManHashOr(p, iFans[0], iFans[1])
	}
	if Oper == '^' {
		return gia.Gia_ManHashXor(p, iFans[0], iFans[1])
	}
	// get third part
	assert(Oper == '?', "Oper == '?'")
	assert(faultStr([]byte(s), pEndNew) == ':', "pEndNew[0] == ':'")
	pBeg = pEndNew + 1
	pEndNew = Gia_ManFormulaEndToken(s, pBeg)
	iFans[2] = Gia_ManRealizeFormula_rec(p, pVars, pPars, s, pBeg, pEndNew, nPars)
	return gia.Gia_ManHashMux(p, iFans[0], iFans[1], iFans[2])
}
func Gia_ManRealizeFormula(p *gia.Gia_Man_t, pVars, pPars []int32, pStr string, nPars int32) int32 {
	return Gia_ManRealizeFormula_rec(p, pVars, pPars, pStr, 0, len(pStr), nPars)
}
func Gia_ManFormulaUnfold(p *gia.Gia_Man_t, pForm string, fFfOnly int32) *gia.Gia_Man_t {
	Count := int32(0)
	var iCtrl [FFTEST_MAX_PARS]int32
	var iFans [FFTEST_MAX_VARS]int32
	var nVars, nPars int32
	assert(len(pForm) < 100, "strlen(pForm) < 100")
	Gia_FormStrCount(pForm, &nVars, &nPars)
	assert(nVars == 2, "nVars == 2")
	pStr := Gia_FormStrTransform(pForm)
	Gia_ManPrintFormula(pStr)
	pNew := gia.Gia_ManStart(5 * gia.Gia_ManObjNum(p))
	gia.Gia_ManSetNameSpec(pNew, gia.Gia_ManName(p), "")
	gia.Gia_ManHashAlloc(pNew)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		gia.Gia_ObjSetValue(gia.Gia_ManCi(p, i), gia.Gia_ManAppendCi(pNew))
	}
	if fFfOnly != 0 {
		gia.Gia_ManCleanMark0(p)
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			gia.Gia_ObjSetMark0(gia.Gia_ObjFanin0(gia.Gia_ManRi(p, i)), 1)
		}
		for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
			pObj := gia.Gia_ManObj(p, i)
			if gia.Gia_ObjIsAnd(pObj) == 0 {
				continue
			}
			if gia.Gia_ObjMark0(pObj) != 0 {
				for k := int32(0); k < nPars; k++ {
					iCtrl[k] = gia.Gia_ManAppendCi(pNew)
				}
				iFans[0] = gia.Gia_ObjFanin0Copy(pObj)
				iFans[1] = gia.Gia_ObjFanin1Copy(pObj)
				gia.Gia_ObjSetValue(pObj, Gia_ManRealizeFormula(pNew, iFans[:], iCtrl[:], pStr, nPars))
				Count++
			} else {
				gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))
			}
		}
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			gia.Gia_ObjSetMark0(gia.Gia_ObjFanin0(gia.Gia_ManRi(p, i)), 0)
		}
	} else {
		for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
			pObj := gia.Gia_ManObj(p, i)
			if gia.Gia_ObjIsAnd(pObj) == 0 {
				continue
			}
			for k := int32(0); k < nPars; k++ {
				iCtrl[k] = gia.Gia_ManAppendCi(pNew)
			}
			iFans[0] = gia.Gia_ObjFanin0Copy(pObj)
			iFans[1] = gia.Gia_ObjFanin1Copy(pObj)
			gia.Gia_ObjSetValue(pObj, Gia_ManRealizeFormula(pNew, iFans[:], iCtrl[:], pStr, nPars))
		}
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		pObj := gia.Gia_ManCo(p, i)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
	}
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManStop(pTemp)
	nCount := gia.Gia_ManAndNum(p)
	if fFfOnly != 0 {
		nCount = Count
	}
	assert(gia.Gia_ManPiNum(pNew) == gia.Gia_ManCiNum(p)+nPars*nCount, "Gia_ManPiNum(pNew) == Gia_ManCiNum(p) + nPars * ...")
	return pNew
}

func Gia_ManFaultCofactor(p *gia.Gia_Man_t, vValues *vec.Vec_Int_t) *gia.Gia_Man_t {
	pNew := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	gia.Gia_ManSetNameSpec(pNew, gia.Gia_ManName(p), "")
	gia.Gia_ManHashAlloc(pNew)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		pObj := gia.Gia_ManPi(p, i)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCi(pNew))
		if int(i) < vec.Vec_IntSize(vValues) {
			gia.Gia_ObjSetValue(pObj, vec.Vec_IntEntry(vValues, int(i)))
		}
	}
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		pObj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(pObj) != 0 {
			gia.Gia_ObjSetValue(pObj, gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(pObj), gia.Gia_ObjFanin1Copy(pObj)))
		}
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		pObj := gia.Gia_ManCo(p, i)
		gia.Gia_ObjSetValue(pObj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(pObj)))
	}
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManStop(pTemp)
	assert(gia.Gia_ManPiNum(pNew) == gia.Gia_ManPiNum(p), "Gia_ManPiNum(pNew) == Gia_ManPiNum(p)")
	return pNew
}

func Gia_ManDumpTests(vTests *vec.Vec_Int_t, nIter int32, pFileName string) {
	pFile, err := os.Create(pFileName)
	if err != nil {
		panic(err)
	}
	defer pFile.Close()
	nVars := int32(vec.Vec_IntSize(vTests)) / nIter
	assert(int32(vec.Vec_IntSize(vTests))%nIter == 0, "Vec_IntSize(vTests) % nIter == 0")
	buf := make([]byte, 0, vec.Vec_IntSize(vTests)+int(nIter))
	v := 0
	for i := int32(0); i < nIter; i++ {
		for k := int32(0); k < nVars; k++ {
			buf = fmt.Appendf(buf, "%d", vec.Vec_IntEntry(vTests, v))
			v++
		}
		buf = append(buf, '\n')
	}
	pFile.Write(buf)
}

func Gia_ManDumpTestsSimulate(p *gia.Gia_Man_t, vValues *vec.Vec_Int_t) {
	assert(int32(vec.Vec_IntSize(vValues)) == gia.Gia_ManCiNum(p), "Vec_IntSize(vValues) == Gia_ManCiNum(p)")
	gia.Gia_ObjSetMark0(gia.Gia_ManConst0(p), 0)
	for k := int32(0); k < gia.Gia_ManCiNum(p); k++ {
		gia.Gia_ObjSetMark0(gia.Gia_ManCi(p, k), vec.Vec_IntEntry(vValues, int(k)))
	}
	for k := int32(0); k < gia.Gia_ManObjNum(p); k++ {
		pObj := gia.Gia_ManObj(p, k)
		if gia.Gia_ObjIsAnd(pObj) == 0 {
			continue
		}
		gia.Gia_ObjSetMark0(pObj, (gia.Gia_ObjMark0(gia.Gia_ObjFanin0(pObj))^gia.Gia_ObjFaninC0(pObj))&
			(gia.Gia_ObjMark0(gia.Gia_ObjFanin1(pObj))^gia.Gia_ObjFaninC1(pObj)))
	}
	for k := int32(0); k < gia.Gia_ManCoNum(p); k++ {
		pObj := gia.Gia_ManCo(p, k)
		gia.Gia_ObjSetMark0(pObj, gia.Gia_ObjMark0(gia.Gia_ObjFanin0(pObj))^gia.Gia_ObjFaninC0(pObj))
	}
	// collect flop input values
	vec.Vec_IntClear(vValues)
	for k := int32(0); k < gia.Gia_ManRegNum(p); k++ {
		vec.Vec_IntPush(vValues, gia.Gia_ObjMark0(gia.Gia_ManRi(p, k)))
	}
	assert(int32(vec.Vec_IntSize(vValues)) == gia.Gia_ManRegNum(p), "Vec_IntSize(vValues) == Gia_ManRegNum(p)")
}
func Gia_ManDumpTestsDelay(vTests *vec.Vec_Int_t, nIter int32, pFileName string, p *gia.Gia_Man_t) {
	pFile, err := os.Create(pFileName)
	if err != nil {
		panic(err)
	}
	defer pFile.Close()
	vValues := vec.Vec_IntAlloc(int(gia.Gia_ManCiNum(p)))
	nVars := int32(vec.Vec_IntSize(vTests)) / nIter
	assert(int32(vec.Vec_IntSize(vTests))%nIter == 0, "Vec_IntSize(vTests) % nIter == 0")
	assert(nVars == 2*gia.Gia_ManPiNum(p)+gia.Gia_ManRegNum(p), "nVars == 2 * Gia_ManPiNum(p) + Gia_ManRegNum(p)")
	var buf []byte
	for i := int32(0); i < nIter; i++ {
		// collect PIs followed by flops
		vec.Vec_IntClear(vValues)
		for v := gia.Gia_ManRegNum(p); v < gia.Gia_ManCiNum(p); v++ {
			buf = fmt.Appendf(buf, "%d", vec.Vec_IntEntry(vTests, int(i*nVars+v)))
			vec.Vec_IntPush(vValues, vec.Vec_IntEntry(vTests, int(i*nVars+v)))
		}
		for v := int32(0); v < gia.Gia_ManRegNum(p); v++ {
			buf = fmt.Appendf(buf, "%d", vec.Vec_IntEntry(vTests, int(i*nVars+v)))
			vec.Vec_IntPush(vValues, vec.Vec_IntEntry(vTests, int(i*nVars+v)))
		}
		buf = append(buf, '\n')
		// derive next-state values
		Gia_ManDumpTestsSimulate(p, vValues)
		// collect PIs followed by flops
		for v := gia.Gia_ManCiNum(p); v < nVars; v++ {
			buf = fmt.Appendf(buf, "%d", vec.Vec_IntEntry(vTests, int(i*nVars+v)))
		}
		for v := 0; v < vec.Vec_IntSize(vValues); v++ {
			buf = fmt.Appendf(buf, "%d", vec.Vec_IntEntry(vValues, v))
		}
		buf = append(buf, '\n')
	}
	gia.Gia_ManCleanMark0(p)
	pFile.Write(buf)
}

func Gia_ManPrintResults(p *gia.Gia_Man_t, pSat *bsat.Solver, nIter int32, clk int64) {
	pTable, err := os.OpenFile("fault_stats.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer pTable.Close()
	fmt.Fprintf(pTable, "%s ", gia.Gia_ManName(p))
	fmt.Fprintf(pTable, "%d ", gia.Gia_ManPiNum(p))
	fmt.Fprintf(pTable, "%d ", gia.Gia_ManPoNum(p))
	fmt.Fprintf(pTable, "%d ", gia.Gia_ManRegNum(p))
	fmt.Fprintf(pTable, "%d ", gia.Gia_ManAndNum(p))

	fmt.Fprintf(pTable, "%d ", pSat.NumVars())
	fmt.Fprintf(pTable, "%d ", pSat.NClauses())
	fmt.Fprintf(pTable, "%d ", pSat.NConflicts())

	fmt.Fprintf(pTable, "%d ", nIter)
	fmt.Fprintf(pTable, "%.2f", 1.0*float64(clk)/float64(util.CLOCKS_PER_SEC))
	fmt.Fprintf(pTable, "\n")
}

func faultAddCnfClauses(pSat *bsat.Solver, pCnf *cnf.Cnf_Dat_t) bool {
	for _, clause := range pCnf.Clauses {
		if pSat.AddClauseC(clause) == 0 {
			return false
		}
	}
	return true
}

func Gia_ManFaultAddOne(pM *gia.Gia_Man_t, pCnf *cnf.Cnf_Dat_t, pSat *bsat.Solver, vLits *vec.Vec_Int_t, nFuncVars, fAddOr int32, pGiaCnf *gia.Gia_Man_t) int32 {
	// derive the cofactor
	pC := Gia_ManFaultCofactor(pM, vLits)
	// derive new CNF
	pCnf2 := Cnf_DeriveGiaRemapped(pC)
	Cnf_DataLiftGia(pCnf2, pC, pSat.NumVars())
	// add timeframe clauses
	if !faultAddCnfClauses(pSat, pCnf2) {
		cnf.Cnf_DataFree(pCnf2)
		gia.Gia_ManStop(pC)
		return 0
	}
	// add constraint clauses
	if fAddOr != 0 {
		// add one OR gate to assert difference of at least one output of the miter
		vOrGate := make([]bsat.Literal, 0, gia.Gia_ManPoNum(pC))
		for i := int32(0); i < gia.Gia_ManPoNum(pC); i++ {
			pObj := gia.Gia_ManPo(pC, i)
			vOrGate = append(vOrGate, util.Abc_Var2Lit(pCnf2.VarNums[gia.Gia_ObjId(pC, pObj)], 0))
		}
		if pSat.AddClauseC(vOrGate) == 0 {
			assert(false, "0")
		}
	} else {
		// add negative literals to assert equality of all outputs of the miter
		for i := int32(0); i < gia.Gia_ManPoNum(pC); i++ {
			pObj := gia.Gia_ManPo(pC, i)
			Lit := util.Abc_Var2Lit(pCnf2.VarNums[gia.Gia_ObjId(pC, pObj)], 1)
			if pSat.AddClauseC([]bsat.Literal{Lit}) == 0 {
				cnf.Cnf_DataFree(pCnf2)
				gia.Gia_ManStop(pC)
				return 0
			}
		}
	}
	// add connection clauses
	if pGiaCnf != nil {
		for i := int32(0); i < gia.Gia_ManPiNum(pGiaCnf); i++ {
			if i >= nFuncVars {
				pObj := gia.Gia_ManPi(pGiaCnf, i)
				pSat.AddBufferC(pCnf.VarNums[gia.Gia_ObjId(pGiaCnf, pObj)], pCnf2.VarNums[gia.Gia_ObjId(pC, gia.Gia_ManPi(pC, i))], 0)
			}
		}
	}
	cnf.Cnf_DataFree(pCnf2)
	gia.Gia_ManStop(pC)
	return 1
}

func Gia_ManDumpUntests(pM *gia.Gia_Man_t, pCnf *cnf.Cnf_Dat_t, pSat *bsat.Solver, nFuncVars int32, pFileName string, fVerbose int32) int32 {
	pFile, err := os.Create(pFileName)
	if err != nil {
		panic(err)
	}
	defer pFile.Close()
	var out []byte
	defer func() { pFile.Write(out) }()
	nItersMax := int32(10000)
	Count := 0
	vLits := vec.Vec_IntAlloc(int(gia.Gia_ManPiNum(pM) - nFuncVars))
	var nIters int32
	for nIters = 0; nIters < nItersMax; nIters++ {
		status := pSat.SolveWithResourceLimits(nil, 0, 0, 0, 0)
		if status == bsat.Unknown {
			fmt.Printf("Timeout reached after dumping %d untestable faults.\n", nIters)
		}
		if status == bsat.Unknown {
			break
		}
		if status == bsat.Unsatisfiable {
			break
		}
		// collect literals
		vec.Vec_IntClear(vLits)
		for i := int32(0); i < gia.Gia_ManPiNum(pM); i++ {
			if i >= nFuncVars {
				v := pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManPi(pM, i))]
				vec.Vec_IntPush(vLits, util.Abc_Var2Lit(v, faultB2i(pSat.VarValue(v))))
			}
		}
		// dump the fault
		lits := vec.Vec_IntArray(vLits)
		i := 0
		for i = 0; i < len(lits); i++ {
			if util.Abc_LitIsCompl(lits[i]) != 0 {
				break
			}
		}
		if i < len(lits) {
			if fVerbose != 0 {
				Count++
				fmt.Printf("Untestable fault %4d : ", Count)
				for i, Value := range lits {
					if util.Abc_LitIsCompl(Value) != 0 {
						fmt.Printf("%d ", i)
					}
				}
				fmt.Printf("\n")
			}
			for i, Value := range lits {
				if util.Abc_LitIsCompl(Value) != 0 {
					out = fmt.Appendf(out, "%d ", i)
				}
			}
			out = append(out, '\n')
		}
		// add this clause
		if pSat.AddClauseC(lits) == 0 {
			break
		}
	}
	return nIters
}

func Gia_ManGetTestPatterns(pFileName string) *vec.Vec_Int_t {
	data, err := os.ReadFile(pFileName)
	if err != nil {
		fmt.Printf("Cannot open input file \"%s\".\n", pFileName)
		return nil
	}
	vTests := vec.Vec_IntAlloc(10000)
	for _, c := range data {
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			continue
		}
		if c != '0' && c != '1' {
			fmt.Printf("Wrong symbol (%c) in the input file.\n", c)
			return nil
		}
		vec.Vec_IntPush(vTests, int32(c-'0'))
	}
	return vTests
}

// Gia_ManDeriveDup derives the second AIG.
func Gia_ManDeriveDup(p *gia.Gia_Man_t, nPisNew int32) *gia.Gia_Man_t {
	pNew := gia.Gia_ManDup(p)
	for i := int32(0); i < nPisNew; i++ {
		gia.Gia_ManAppendCi(pNew)
	}
	return pNew
}

func Gia_ManFaultAnalyze(pSat *bsat.Solver, vPars, vMap, vLits *vec.Vec_Int_t, Iter int32) int32 {
	nConfLimit := int64(100)
	nUnsats, nRuns := int32(0), 0
	clk := util.Abc_Clock()
	assert(vec.Vec_IntSize(vPars) == vec.Vec_IntSize(vMap), "Vec_IntSize(vPars) == Vec_IntSize(vMap)")
	// check presence of each variable
	vec.Vec_IntClear(vLits)
	vec.Vec_IntAppend(vLits, vMap)
	for v := 0; v < vec.Vec_IntSize(vPars); v++ {
		if vec.Vec_IntEntry(vLits, v) == 0 {
			continue
		}
		assert(vec.Vec_IntEntry(vLits, v) == 1, "Vec_IntEntry(vLits, v) == 1")
		nRuns++
		Lit := util.Abc_Var2Lit(vec.Vec_IntEntry(vPars, v), 0)
		status := pSat.SolveWithResourceLimits([]bsat.Literal{Lit}, nConfLimit, 0, 0, 0)
		if status == bsat.Unknown {
			continue
		}
		if status == bsat.Unsatisfiable {
			nUnsats++
			assert(vec.Vec_IntEntry(vMap, v) == 1, "Vec_IntEntry(vMap, v) == 1")
			vec.Vec_IntWriteEntry(vMap, v, 0)
			//status = sat_solver_addclause( pSat, &Lit, &Lit+1 );
			continue
		}
		for i, iVar := range vec.Vec_IntArray(vPars) {
			if vec.Vec_IntEntry(vLits, i) != 0 && pSat.VarValue(iVar) {
				vec.Vec_IntWriteEntry(vLits, i, 0)
			}
		}
		assert(vec.Vec_IntEntry(vLits, v) == 0, "Vec_IntEntry(vLits, v) == 0")
	}
	fmt.Printf("Iteration %3d has determined %5d (out of %5d) parameters after %6d SAT calls.  ", Iter, vec.Vec_IntSize(vMap)-vec.Vec_IntCountPositive(vMap), vec.Vec_IntSize(vPars), nRuns)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	return nUnsats
}

func faultSetRuntimeLimit(pSat *bsat.Solver, nTimeOut int32) {
	if nTimeOut != 0 {
		pSat.SetRuntimeDeadline(util.Abc_Clock, int64(nTimeOut)*util.CLOCKS_PER_SEC+util.Abc_Clock())
	} else {
		pSat.SetRuntimeDeadline(util.Abc_Clock, 0)
	}
}

// Gia_ManFaultDumpNewFaults dumps faults detected by the new test, which are
// not detected by previous tests.
func Gia_ManFaultDumpNewFaults(pM *gia.Gia_Man_t, nFuncVars int32, vTests, vTestNew *vec.Vec_Int_t, pPars *Bmc_ParFf_t) int32 {
	pFileName := "newfaults.txt"

	// derive the cofactor
	pC := Gia_ManFaultCofactor(pM, vTestNew)
	// derive new CNF
	pCnf2 := Cnf_DeriveGiaRemapped(pC)

	// create SAT solver
	pSat := bsat.NewSolver()
	pSat.SetNumVars(1)
	faultSetRuntimeLimit(pSat, pPars.NTimeOut)
	// create the first cofactor
	Gia_ManFaultAddOne(pM, nil, pSat, vTestNew, nFuncVars, 1, nil)

	// add other test patterns
	assert(vec.Vec_IntSize(vTests)%int(nFuncVars) == 0, "Vec_IntSize(vTests) % nFuncVars == 0")
	IterMax := int32(vec.Vec_IntSize(vTests)) / nFuncVars
	vLits := vec.Vec_IntAlloc(int(nFuncVars))
	for Iter := int32(0); Iter < IterMax; Iter++ {
		// get pattern
		vec.Vec_IntClear(vLits)
		for i := int32(0); i < nFuncVars; i++ {
			vec.Vec_IntPush(vLits, vec.Vec_IntEntry(vTests, int(Iter*nFuncVars+i)))
		}
		// the resulting problem cannot be UNSAT, because the new test is guaranteed
		// to detect something that the given test set does not detect
		if Gia_ManFaultAddOne(pM, pCnf2, pSat, vLits, nFuncVars, 0, pC) == 0 {
			assert(false, "0")
		}
	}

	// dump the new faults
	clk := util.Abc_Clock()
	nNewFaults := Gia_ManDumpUntests(pC, pCnf2, pSat, nFuncVars, pFileName, pPars.FVerbose)
	fmt.Printf("Dumped %d new multiple faults into file \"%s\".  ", nNewFaults, pFileName)
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)

	// cleanup
	pSat.Delete()
	cnf.Cnf_DataFree(pCnf2)
	gia.Gia_ManStop(pC)
	return 1
}

// Gia_ManFaultPrepare generates miter, CNF and solver.
func Gia_ManFaultPrepare(p, pG *gia.Gia_Man_t, pPars *Bmc_ParFf_t, nFuncVars int32, vMap, vTests, vLits *vec.Vec_Int_t, ppMiter **gia.Gia_Man_t, ppCnf **cnf.Cnf_Dat_t, ppSat **bsat.Solver, fWarmUp int32) int32 {
	var p0, p1 *gia.Gia_Man_t
	var clkSat int64

	if vec.Vec_IntSize(vTests) != 0 && (vec.Vec_IntSize(vTests)%int(nFuncVars) != 0) {
		fmt.Printf("The number of symbols in the input patterns (%d) does not divide evenly on the number of test variables (%d).\n", vec.Vec_IntSize(vTests), nFuncVars)
		return 0
	}

	// select algorithm
	if pPars.Algo == 0 {
		p1 = Gia_ManFormulaUnfold(p, pPars.PFormStr, pPars.FFfOnly)
	} else if pPars.Algo == 1 {
		assert(gia.Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0")
		p0 = Gia_ManFaultUnfold(pG, 0, pPars.FFfOnly)
		p1 = Gia_ManFaultUnfold(p, 1, pPars.FFfOnly)
	} else if pPars.Algo == 2 {
		p1 = Gia_ManStuckAtUnfold(p, vMap)
	} else if pPars.Algo == 3 {
		p1 = Gia_ManFlipUnfold(p, vMap)
	} else if pPars.Algo == 4 {
		p1 = Gia_ManFOFUnfold(p, vMap)
	}
	if pPars.Algo != 1 {
		p0 = Gia_ManDeriveDup(pG, gia.Gia_ManCiNum(p1)-gia.Gia_ManCiNum(pG))
	}
	//    Gia_AigerWrite( p1, "newfault.aig", 0, 0, 0 );
	//    printf( "Dumped circuit with fault parameters into file \"newfault.aig\".\n" );

	// create miter
	pM := gia.Gia_ManMiter(p0, p1, 0, 0, 0, 0, 0)
	pCnf := Cnf_DeriveGiaRemapped(pM)
	gia.Gia_ManStop(p0)
	gia.Gia_ManStop(p1)

	// start the SAT solver
	pSat := bsat.NewSolver()
	pSat.SetNumVars(cnf.Cnf_DataVarNum(pCnf))
	faultSetRuntimeLimit(pSat, pPars.NTimeOut)
	// add timeframe clauses
	if !faultAddCnfClauses(pSat, pCnf) {
		assert(false, "0")
	}

	// add one large OR clause
	vec.Vec_IntClear(vLits)
	for i := int32(0); i < gia.Gia_ManCoNum(pM); i++ {
		vec.Vec_IntPush(vLits, util.Abc_Var2Lit(pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManCo(pM, i))], 0))
	}
	pSat.AddClauseC(vec.Vec_IntArray(vLits))

	// save return data
	*ppMiter = pM
	*ppCnf = pCnf
	*ppSat = pSat

	// add cardinality constraint
	if pPars.FBasic != 0 {
		vec.Vec_IntClear(vLits)
		for i := int32(0); i < gia.Gia_ManPiNum(pM); i++ {
			if i >= nFuncVars {
				vec.Vec_IntPush(vLits, pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManPi(pM, i))])
			}
		}
		Cnf_AddCardinConstr(pSat, vLits)
	} else if pPars.NCardConstr != 0 {
		vec.Vec_IntClear(vLits)
		for i := int32(0); i < gia.Gia_ManPiNum(pM); i++ {
			if i >= nFuncVars {
				vec.Vec_IntPush(vLits, pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManPi(pM, i))])
			}
		}
		Cnf_AddCardinConstrGeneral(pSat, vLits, pPars.NCardConstr, faultB2i(pPars.FNonStrict == 0))
	}

	// add available test-patterns
	if vec.Vec_IntSize(vTests) > 0 {
		nTests := int32(vec.Vec_IntSize(vTests)) / nFuncVars
		assert(vec.Vec_IntSize(vTests)%int(nFuncVars) == 0, "Vec_IntSize(vTests) % nFuncVars == 0")
		if pPars.PFileName != "" {
			fmt.Printf("Reading %d pre-computed test patterns from file \"%s\".\n", int32(vec.Vec_IntSize(vTests))/nFuncVars, pPars.PFileName)
		} else {
			fmt.Printf("Reading %d pre-computed test patterns from previous rounds.\n", int32(vec.Vec_IntSize(vTests))/nFuncVars)
		}
		for Iter := int32(0); Iter < nTests; Iter++ {
			if fWarmUp != 0 {
				clk := util.Abc_Clock()
				status := pSat.SolveWithResourceLimits(nil, 0, 0, 0, 0)
				clkSat += util.Abc_Clock() - clk
				if status == bsat.Unknown {
					if pPars.FVerbose != 0 {
						fmt.Printf("\n")
					}
					fmt.Printf("Timeout reached after %d seconds and adding %d tests.\n", pPars.NTimeOut, Iter)
					vec.Vec_IntShrink(vTests, int(Iter*nFuncVars))
					return 0
				}
				if status == bsat.Unsatisfiable {
					if pPars.FVerbose != 0 {
						fmt.Printf("\n")
					}
					fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter)
					vec.Vec_IntShrink(vTests, int(Iter*nFuncVars))
					return 0
				}
			}
			// get pattern
			vec.Vec_IntClear(vLits)
			for i := int32(0); i < nFuncVars; i++ {
				vec.Vec_IntPush(vLits, vec.Vec_IntEntry(vTests, int(Iter*nFuncVars+i)))
			}
			if Gia_ManFaultAddOne(pM, pCnf, pSat, vLits, nFuncVars, 0, pM) == 0 {
				if pPars.FVerbose != 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter)
				vec.Vec_IntShrink(vTests, int(Iter*nFuncVars))
				return 0
			}
			if pPars.FVerbose != 0 {
				faultPrintIter(Iter, pSat, clkSat)
			}
		}
	} else if pPars.FStartPats != 0 {
		for Iter := int32(0); Iter < 2; Iter++ {
			status := pSat.SolveWithResourceLimits(nil, 0, 0, 0, 0)
			if status == bsat.Unknown {
				if pPars.FVerbose != 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("Timeout reached after %d seconds and %d iterations.\n", pPars.NTimeOut, Iter)
				vec.Vec_IntShrink(vTests, int(Iter*nFuncVars))
				return 0
			}
			if status == bsat.Unsatisfiable {
				if pPars.FVerbose != 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("The problem is UNSAT after %d iterations.\n", Iter)
				vec.Vec_IntShrink(vTests, int(Iter*nFuncVars))
				return 0
			}
			// initialize simple pattern
			vec.Vec_IntFill(vLits, int(nFuncVars), Iter)
			vec.Vec_IntAppend(vTests, vLits)
			if Gia_ManFaultAddOne(pM, pCnf, pSat, vLits, nFuncVars, 0, pM) == 0 {
				if pPars.FVerbose != 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter)
				vec.Vec_IntShrink(vTests, int(Iter*nFuncVars))
				return 0
			}
		}
	}

	fmt.Printf("Using miter with:  AIG nodes = %6d.  CNF variables = %6d.  CNF clauses = %8d.\n", gia.Gia_ManAndNum(pM), cnf.Cnf_DataVarNum(pCnf), cnf.Cnf_DataClauseNum(pCnf))
	return 1
}

// faultPrintIter prints C's verbose iteration line, ending in ABC_PRTr.
func faultPrintIter(Iter int32, pSat *bsat.Solver, clkSat int64) {
	fmt.Printf("Iter%6d : ", Iter)
	fmt.Printf("Var =%10d  ", pSat.NumVars())
	fmt.Printf("Clause =%10d  ", pSat.NClauses())
	fmt.Printf("Conflict =%10d  ", pSat.NConflicts())
	//Abc_PrintTime( 1, "Time", clkSat );
	fmt.Printf("%s =", "Solver time")
	fmt.Printf("%9.2f sec\r", 1.0*float64(clkSat)/float64(util.CLOCKS_PER_SEC))
}

func Gia_ManFaultTest(p, pG *gia.Gia_Man_t, pPars *Bmc_ParFf_t) {
	nIterMax := int32(1000000)
	var nVars, nPars int32
	var Iter, Iter2 int32
	var status bsat.Status
	nFuncVars := int32(-1)
	var clk, clkSat int64
	clkTotal := util.Abc_Clock()
	var vLits, vMap, vTests, vPars *vec.Vec_Int_t
	var pM *gia.Gia_Man_t
	var pCnf *cnf.Cnf_Dat_t
	var pSat *bsat.Solver

	if pPars.Algo == 0 && Gia_FormStrCount(pPars.PFormStr, &nVars, &nPars) != 0 {
		return
	}

	single := ""
	if pPars.FBasic != 0 {
		single = "single "
	}
	// select algorithm
	if pPars.Algo == 0 {
		fmt.Printf("FFTEST is computing test patterns for fault model \"%s\"...\n", pPars.PFormStr)
	} else if pPars.Algo == 1 {
		fmt.Printf("FFTEST is computing test patterns for %sdelay faults...\n", single)
	} else if pPars.Algo == 2 {
		fmt.Printf("FFTEST is computing test patterns for %sstuck-at faults...\n", single)
	} else if pPars.Algo == 3 {
		fmt.Printf("FFTEST is computing test patterns for %scomplement faults...\n", single)
	} else if pPars.Algo == 4 {
		fmt.Printf("FFTEST is computing test patterns for %sfunctionally observable faults...\n", single)
	} else {
		fmt.Printf("Unrecognized algorithm (%d).\n", pPars.Algo)
		return
	}

	fmt.Printf("Options: ")
	if pPars.FCheckUntest != 0 || pPars.FDumpDelay != 0 {
		fmt.Printf("Untestable faults = %s. ", "yes")
	} else {
		fmt.Printf("Untestable faults = %s. ", "no")
	}
	if pPars.NCardConstr != 0 {
		nonStrict := ""
		if pPars.FNonStrict != 0 {
			nonStrict = "non-"
		}
		fmt.Printf("Using %sstrict cardinality %d. ", nonStrict, pPars.NCardConstr)
	}
	if pPars.FFfOnly != 0 {
		fmt.Printf("Faults at FF outputs only = yes. ")
	}
	if pPars.NTimeOut != 0 {
		fmt.Printf("Runtime limit = %d sec.  ", pPars.NTimeOut)
	}
	if p != pG && gia.Gia_ManSpec(pG) != "" {
		fmt.Printf("Golden model = %s. ", gia.Gia_ManSpec(pG))
	}
	if pPars.FVerbose != 0 {
		fmt.Printf("Verbose = %s. ", "yes")
	} else {
		fmt.Printf("Verbose = %s. ", "no")
	}
	fmt.Printf("\n")

	// select algorithm
	if pPars.Algo == 0 {
		nFuncVars = gia.Gia_ManCiNum(p)
	} else if pPars.Algo == 1 {
		nFuncVars = gia.Gia_ManRegNum(p) + 2*gia.Gia_ManPiNum(p)
	} else if pPars.Algo == 2 {
		nFuncVars = gia.Gia_ManCiNum(p)
	} else if pPars.Algo == 3 {
		nFuncVars = gia.Gia_ManCiNum(p)
	} else if pPars.Algo == 4 {
		nFuncVars = gia.Gia_ManCiNum(p)
	}

	// collect test patterns from file
	if pPars.PFileName != "" {
		vTests = Gia_ManGetTestPatterns(pPars.PFileName)
	} else {
		vTests = vec.Vec_IntAlloc(10000)
	}
	if vTests == nil {
		return
	}

	// select algorithm
	vMap = vec.Vec_IntAlloc(0)
	if pPars.Algo == 2 {
		vec.Vec_IntFill(vMap, int(2*gia.Gia_ManAndNum(p)), 1)
	} else if pPars.Algo == 3 {
		vec.Vec_IntFill(vMap, int(gia.Gia_ManAndNum(p)), 1)
	} else if pPars.Algo == 4 {
		vec.Vec_IntFill(vMap, int(4*gia.Gia_ManAndNum(p)), 1)
	}

	// prepare SAT solver
	vLits = vec.Vec_IntAlloc(int(gia.Gia_ManCoNum(p)))

	// add user-speicified test-vectors (if given) and create missing ones (if needed)
	if Gia_ManFaultPrepare(p, pG, pPars, nFuncVars, vMap, vTests, vLits, &pM, &pCnf, &pSat, 1) != 0 {
		if pPars.FStartPats != 0 {
			Iter = 2
		} else {
			Iter = int32(vec.Vec_IntSize(vTests)) / nFuncVars
		}
		for ; Iter < nIterMax; Iter++ {
			// collect parameter variables
			if pPars.NIterCheck != 0 && vPars == nil {
				vPars = vec.Vec_IntAlloc(int(gia.Gia_ManPiNum(pM) - nFuncVars))
				for i := int32(0); i < gia.Gia_ManPiNum(pM); i++ {
					if i >= nFuncVars {
						vec.Vec_IntPush(vPars, pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManPi(pM, i))])
					}
				}
				assert(int32(vec.Vec_IntSize(vPars)) == gia.Gia_ManPiNum(pM)-nFuncVars, "Vec_IntSize(vPars) == Gia_ManPiNum(pM) - nFuncVars")
			}
			// derive unit parameter variables (only fault models with a fault map have them;
			// C read past the empty vMap of -A 0/-A 1 here: CBUGS_FIXED.md #24)
			if Iter != 0 && pPars.NIterCheck != 0 && (Iter%pPars.NIterCheck) == 0 && vec.Vec_IntSize(vMap) > 0 {
				Gia_ManFaultAnalyze(pSat, vPars, vMap, vLits, Iter)
				// cleanup
				gia.Gia_ManStop(pM)
				cnf.Cnf_DataFree(pCnf)
				pSat.Delete()
				// recompute
				if Gia_ManFaultPrepare(p, pG, pPars, nFuncVars, vMap, vTests, vLits, &pM, &pCnf, &pSat, 0) == 0 {
					fmt.Printf("This should never happen.\n")
					return
				}
				vec.Vec_IntFreeP(&vPars)
			}
			// solve
			clk = util.Abc_Clock()
			status = pSat.SolveWithResourceLimits(nil, 0, 0, 0, 0)
			clkSat += util.Abc_Clock() - clk
			if pPars.FVerbose != 0 {
				faultPrintIter(Iter, pSat, clkSat)
			}
			if status == bsat.Unknown {
				if pPars.FVerbose != 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("Timeout reached after %d seconds and %d iterations.\n", pPars.NTimeOut, Iter)
				break // goto finish
			}
			if status == bsat.Unsatisfiable {
				if pPars.FVerbose != 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("The problem is UNSAT after %d iterations.  ", Iter)
				break
			}
			assert(status == bsat.Satisfiable, "status == l_True")
			// collect SAT assignment
			vec.Vec_IntClear(vLits)
			for i := int32(0); i < gia.Gia_ManPiNum(pM); i++ {
				if i < nFuncVars {
					vec.Vec_IntPush(vLits, faultB2i(pSat.VarValue(pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManPi(pM, i))])))
				}
			}
			// if the user selected to generate new faults detected by this test (vLits)
			// and not detected by the given test set (vTests), we compute the new faults here and quit
			if pPars.FDumpNewFaults != 0 {
				if vec.Vec_IntSize(vTests) == 0 {
					fmt.Printf("The input test patterns are not given.\n")
				} else {
					Gia_ManFaultDumpNewFaults(pM, nFuncVars, vTests, vLits, pPars)
				}
				return // goto finish_all
			}
			vec.Vec_IntAppend(vTests, vLits)
			// add constraint
			if Gia_ManFaultAddOne(pM, pCnf, pSat, vLits, nFuncVars, 0, pM) == 0 {
				if pPars.FVerbose != 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter)
				break
			}
		}
	} else {
		Iter = int32(vec.Vec_IntSize(vTests)) / nFuncVars
	}
finish:
	// print results
	//    if ( status == l_False )
	//        Gia_ManPrintResults( p, pSat, Iter, Abc_Clock() - clkTotal );
	// cleanup
	util.Abc_PrintTime(1, "Testing runtime", util.Abc_Clock()-clkTotal)
	// dump the test suite
	if pPars.FDump != 0 {
		pFileName := "tests.txt"
		if gia.Gia_ManSpec(p) != "" {
			pFileName = extra.Extra_FileNameGenericAppend(gia.Gia_ManSpec(p), "_tests.txt")
		}
		if pPars.FDumpDelay != 0 && pPars.Algo == 1 {
			Gia_ManDumpTestsDelay(vTests, Iter, pFileName, p)
			fmt.Printf("Dumping %d pairs of test patterns (total %d pattern) into file \"%s\".\n", int32(vec.Vec_IntSize(vTests))/nFuncVars, 2*int32(vec.Vec_IntSize(vTests))/nFuncVars, pFileName)
		} else {
			Gia_ManDumpTests(vTests, Iter, pFileName)
			fmt.Printf("Dumping %d test patterns into file \"%s\".\n", int32(vec.Vec_IntSize(vTests))/nFuncVars, pFileName)
		}
	}

	// compute untestable faults
	if Iter != 0 && (p != pG || pPars.FDumpUntest != 0 || pPars.FCheckUntest != 0) {
		clkTotal := util.Abc_Clock()
		// restart the SAT solver
		pSat.Delete()
		pSat = bsat.NewSolver()
		pSat.SetNumVars(cnf.Cnf_DataVarNum(pCnf))
		faultSetRuntimeLimit(pSat, pPars.NTimeOut)
		// add timeframe clauses
		if !faultAddCnfClauses(pSat, pCnf) {
			assert(false, "0")
		}
		// add constraint to rule out no fault
		//        if ( p == pG )
		{
			vec.Vec_IntClear(vLits)
			for i := int32(0); i < gia.Gia_ManPiNum(pM); i++ {
				if i >= nFuncVars {
					vec.Vec_IntPush(vLits, util.Abc_Var2Lit(pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManPi(pM, i))], 0))
				}
			}
			pSat.AddClauseC(vec.Vec_IntArray(vLits))
		}
		// add cardinality constraint
		if pPars.FBasic != 0 {
			vec.Vec_IntClear(vLits)
			for i := int32(0); i < gia.Gia_ManPiNum(pM); i++ {
				if i >= nFuncVars {
					vec.Vec_IntPush(vLits, pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManPi(pM, i))])
				}
			}
			Cnf_AddCardinConstr(pSat, vLits)
		}
		// add output clauses
		for i := int32(0); i < gia.Gia_ManCoNum(pM); i++ {
			Lit := util.Abc_Var2Lit(pCnf.VarNums[gia.Gia_ObjId(pM, gia.Gia_ManCo(pM, i))], 1)
			pSat.AddClauseC([]bsat.Literal{Lit})
		}
		// simplify the SAT solver
		ok := pSat.SimplifyC()
		assert(ok != 0, "status")

		// add test patterns
		assert(int32(vec.Vec_IntSize(vTests)) == Iter*nFuncVars, "Vec_IntSize(vTests) == Iter * nFuncVars")
		for Iter2 = 0; ; Iter2++ {
			clk := util.Abc_Clock()
			status = pSat.SolveWithResourceLimits(nil, 0, 0, 0, 0)
			clkSat += util.Abc_Clock() - clk
			if pPars.FVerbose != 0 {
				faultPrintIter(Iter2, pSat, clkSat)
			}
			if status == bsat.Unknown {
				if pPars.FVerbose != 0 {
					fmt.Printf("\n")
				}
				fmt.Printf("Timeout reached after %d seconds and %d iterations.\n", pPars.NTimeOut, Iter2)
				goto finish
			}
			if Iter2 == Iter {
				break
			}
			assert(status == bsat.Satisfiable, "status == l_True")
			// get pattern
			vec.Vec_IntClear(vLits)
			for i := int32(0); i < nFuncVars; i++ {
				vec.Vec_IntPush(vLits, vec.Vec_IntEntry(vTests, int(Iter2*nFuncVars+i)))
			}
			if Gia_ManFaultAddOne(pM, pCnf, pSat, vLits, nFuncVars, 0, pM) == 0 {
				fmt.Printf("The problem is UNSAT after adding %d tests.\n", Iter2)
				goto finish
			}
		}
		assert(Iter2 == Iter, "Iter2 == Iter")
		if pPars.FVerbose != 0 {
			fmt.Printf("\n")
		}
		if p == pG {
			if status == bsat.Satisfiable {
				fmt.Printf("There are untestable faults.  ")
			} else if status == bsat.Unsatisfiable {
				fmt.Printf("There is no untestable faults.  ")
			} else {
				assert(false, "0")
			}
			util.Abc_PrintTime(1, "Fault computation runtime", util.Abc_Clock()-clkTotal)
		} else {
			if status == bsat.Satisfiable {
				fmt.Printf("The circuit is rectifiable.  ")
			} else if status == bsat.Unsatisfiable {
				fmt.Printf("The circuit is not rectifiable (or equivalent to the golden one).  ")
			} else {
				assert(false, "0")
			}
			util.Abc_PrintTime(1, "Rectification runtime", util.Abc_Clock()-clkTotal)
		}
		// dump untestable faults
		if pPars.FDumpUntest != 0 && status == bsat.Satisfiable {
			clk := util.Abc_Clock()
			pFileName := "untest.txt"
			if gia.Gia_ManSpec(p) != "" {
				pFileName = extra.Extra_FileNameGenericAppend(gia.Gia_ManSpec(p), "_untest.txt")
			}
			nUntests := Gia_ManDumpUntests(pM, pCnf, pSat, nFuncVars, pFileName, pPars.FVerbose)
			if p == pG {
				fmt.Printf("Dumped %d untestable multiple faults into file \"%s\".  ", nUntests, pFileName)
			} else {
				fmt.Printf("Dumped %d ways of rectifying the circuit into file \"%s\".  ", nUntests, pFileName)
			}
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
		}
	}
	// finish_all:
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
