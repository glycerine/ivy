package acb

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type sat_solver = bsat.Solver
type abctime = int64

func Cnf_DataLiftAndFlipLits(p *Cnf_Dat_t, offset int32, flips *Vec_Int_t) {
	cnf.Cnf_DataLiftAndFlipLits(p, offset, flips)
}
func Cnf_DataCollectFlipLits(p *Cnf_Dat_t, pivot int32, flips *Vec_Int_t) {
	cnf.Cnf_DataCollectFlipLits(p, pivot, flips)
}
func sat_solver_add_buffer_enable(s *sat_solver, a, b, enable, complement int32) int32 {
	first := bsat.Positive(a)
	if complement != 0 {
		first = bsat.Negative(a)
	}
	second := bsat.Positive(b)
	if complement == 0 {
		second = bsat.Negative(b)
	}
	if !s.AddClause(first, second, bsat.Negative(enable)) {
		return 0
	}
	first = bsat.Negative(a)
	if complement == 0 {
		second = bsat.Positive(b)
	} else {
		second = bsat.Negative(b)
	}
	if !s.AddClause(first, second, bsat.Negative(enable)) {
		return 0
	}
	return 2
}
func sat_solver_simplify(s *sat_solver) int32 {
	if s.Simplify() {
		return 1
	}
	return 0
}
func sat_solver_restart(s *sat_solver) { s.RestartSearch() }
func ABC_PRTP(name string, elapsed, total int64) {
	percent := 0.0
	if total != 0 {
		percent = 100 * float64(elapsed) / float64(total)
	}
	fmt.Printf("%s = %9.2f sec (%6.2f %%)\n", name, float64(elapsed)/1e9, percent)
}

/**CFile****************************************************************

  FileName    [abcMfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Hierarchical word-level netlist.]

  Synopsis    [Optimization with don't-cares.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - July 21, 2015.]

  Revision    [$Id: abcMfs.c,v 1.00 2014/11/29 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "acb.h"
//#include "bool/kit/kit.h"
//#include "sat/bsat/satSolver.h"
//#include "sat/cnf/cnf.h"
//#include "misc/util/utilTruth.h"
//#include "acbPar.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Acb_ObjIsDelayCriticalFanin(p *Acb_Ntk_t, i int32, f int32) int32 {
	return boolInt(Acb_ObjIsCi(p, f) == 0 && Acb_ObjLevelR(p, i)+Acb_ObjLevelD(p, f) == p.LevelMax, 1, 0)
}
func Acb_ObjIsAreaCritical(p *Acb_Ntk_t, f int32) int32 {
	return boolInt(Acb_ObjIsCi(p, f) == 0 && Acb_ObjFanoutNum(p, f) == 1, 1, 0)
}
func Acb_ObjIsCritical(p *Acb_Ntk_t, i int32, f int32, fDel int32) int32 {
	if fDel != 0 {
		return Acb_ObjIsDelayCriticalFanin(p, i, f)
	} else {
		return Acb_ObjIsAreaCritical(p, f)
	}
}

// Kit_TruthIsop adapts ABC's word truth table and cube vector to the shared
// Kit implementation. Kit stores the same two-bit-per-variable cube encoding.
func Kit_TruthIsop(truth *word, nVars int32, cover *Vec_Int_t, tryBoth int32) int32 {
	words := []uint32{uint32(*truth), uint32(*truth >> 32)}
	var cubes []uint32
	phase := kit.Kit_TruthIsop(words, nVars, &cubes, tryBoth)
	Vec_IntClear(cover)
	for _, cube := range cubes {
		Vec_IntPush(cover, int32(cube))
	}
	return phase
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    [Derive CNF for nodes in the window.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_DeriveCnfFromTruth(Truth word, nVars int32, vCover *Vec_Int_t, vCnf *Vec_Str_t) int32 {
	Vec_StrClear(vCnf)
	if Truth == 0 || ^Truth == 0 {
		//        assert( nVars == 0 );
		// A constant still constrains the output at index nVars when its
		// redundant fanins have not yet been removed.
		Vec_StrPush(vCnf, byte(Abc_Var2Lit(nVars, boolInt(Truth == 0, 1, 0))))
		Vec_StrPush(vCnf, byte(255))
		return 1

	} else {
		var i, k, c, RetValue, Literal, Cube, nCubes int32
		assert(nVars > 0, "nVars > 0")
		for c = 0; c < 2; c++ {
			if c != 0 {
				Truth = ^Truth
			} else {
				Truth = Truth
			}
			RetValue = Kit_TruthIsop(&Truth, nVars, vCover, 0)
			assert(RetValue == 0, "RetValue == 0")
			nCubes += Vec_IntSize(vCover)
			for i = 0; i < Vec_IntSize(vCover); i++ {
				Cube = Vec_IntEntry(vCover, i)
				for k = 0; k < nVars; k++ {
					Literal = 3 & (Cube >> (k << 1))
					if Literal == 1 { // '0'  -> pos lit
						Vec_StrPush(vCnf, byte(Abc_Var2Lit(k, 0)))
					} else if Literal == 2 { // '1'  -> neg lit
						Vec_StrPush(vCnf, byte(Abc_Var2Lit(k, 1)))
					} else if Literal != 0 {
						assert(false, "0")
					}
				}
				Vec_StrPush(vCnf, byte(Abc_Var2Lit(nVars, c)))
				Vec_StrPush(vCnf, byte(255))
			}
		}
		return nCubes
	}
}

func Acb_DeriveCnfForWindowOne(p *Acb_Ntk_t, iObj int32) {
	var vCnfs []Vec_Str_t = p.vCnfs
	var vCnfBase *Vec_Str_t = Acb_ObjCnfs(p, iObj)
	assert(Vec_StrSize(vCnfBase) == 0, "Vec_StrSize(vCnfBase) == 0") // unassigned
	assert(len(vCnfs) == int(Acb_NtkObjNumMax(p)), "len(vCnfs) == int(Acb_NtkObjNumMax(p))")
	Acb_DeriveCnfFromTruth(Acb_ObjTruth(p, iObj), Acb_ObjFaninNum(p, iObj), &p.vCover, &p.vCnf)
	Vec_StrGrow(vCnfBase, Vec_StrSize(&p.vCnf))
	Vec_StrClear(vCnfBase)
	for _, entry := range Vec_StrArray(&p.vCnf)[:Vec_StrSize(&p.vCnf)] {
		Vec_StrPush(vCnfBase, entry)
	}
}
func Acb_DeriveCnfForWindow(p *Acb_Ntk_t, vWin *Vec_Int_t, PivotVar int32) []Vec_Str_t {
	var vCnfs []Vec_Str_t = p.vCnfs
	var vCnfBase *Vec_Str_t
	var i, iObj int32
	assert(len(vCnfs) == int(Acb_NtkObjNumMax(p)), "len(vCnfs) == int(Acb_NtkObjNumMax(p))")
	for i = 0; i < Vec_IntSize(vWin); i++ {
		iObj = Vec_IntEntry(vWin, i)
		if Abc_LitIsCompl(iObj) != 0 && i < PivotVar {
			continue
		}
		iObj = Abc_Lit2Var(iObj)
		vCnfBase = Acb_ObjCnfs(p, iObj)
		if Vec_StrSize(vCnfBase) > 0 {
			continue
		}
		Acb_DeriveCnfForWindowOne(p, iObj)
	}
	return vCnfs
}

/*
*Function*************************************************************

	Synopsis    [Constructs CNF for the window.]

	Description [The window for the pivot node is represented as a DFS ordered array
	of objects (vWinObjs) whose indexes are used as SAT variable IDs (stored in p->vCopies).
	PivotVar is the index of the pivot node in array vWinObjs.
	The nodes before (after) PivotVar are TFI (TFO) nodes.
	The leaf (root) nodes are labeled with Abc_LitIsCompl().
	If fQbf is 1, returns the instance meant for QBF solving. It uses the last
	variable (LastVar) as the placeholder for the second copy of the pivot node.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_TranslateCnf(vClas *Vec_Int_t, vLits *Vec_Int_t, vCnf *Vec_Str_t, vSatVars *Vec_Int_t, iPivotVar int32) {
	var i, Lit int32
	for i = 0; i < Vec_StrSize(vCnf); i++ {
		Entry := int8(Vec_StrEntry(vCnf, i))
		if int32(Entry) == -1 {
			Vec_IntPush(vClas, Vec_IntSize(vLits))
			continue
		}
		Lit = util.Abc_Lit2LitV(Vec_IntArray(vSatVars), int32(Entry))
		Lit = Abc_LitNotCond(Lit, Abc_Lit2Var(Lit) == iPivotVar)
		Vec_IntPush(vLits, Lit)
	}
}
func Acb_NtkCountRoots(vWinObjs *Vec_Int_t, PivotVar int32) int32 {
	var i, iObjLit, nRoots int32
	for i = PivotVar + 1; i < Vec_IntSize(vWinObjs); i++ {
		iObjLit = Vec_IntEntry(vWinObjs, i)
		nRoots += Abc_LitIsCompl(iObjLit)
	}
	return nRoots
}
func Acb_DeriveCnfForNode(p *Acb_Ntk_t, iObj int32, pSat *sat_solver, OutVar int32) {
	var vCnfs []Vec_Str_t = p.vCnfs
	var vFaninVars *Vec_Int_t = &p.vCover
	var vClas *Vec_Int_t = Vec_IntAlloc(100)
	var vLits *Vec_Int_t = Vec_IntAlloc(100)
	var k int32
	var iFanin int32
	var pFanins []int32
	var Prev int32
	var This int32
	// collect SAT variables
	Vec_IntClear(vFaninVars)
	for k = 0; k < Acb_ObjFaninNum(p, iObj); k++ {
		pFanins = Acb_ObjFanins(p, iObj)
		iFanin = pFanins[k+1]
		assert(Acb_ObjFunc(p, iFanin) >= 0, "Acb_ObjFunc(p, iFanin) >= 0")
		Vec_IntPush(vFaninVars, Acb_ObjFunc(p, iFanin))
	}
	Vec_IntPush(vFaninVars, OutVar)
	// derive CNF for the node
	Acb_TranslateCnf(vClas, vLits, &vCnfs[iObj], vFaninVars, -1)
	// add clauses
	Prev = 0
	for k = 0; k < Vec_IntSize(vClas); k++ {
		This = Vec_IntEntry(vClas, k)
		if !sat_solver_addclause(pSat, Vec_IntArray(vLits)[Prev:This]) {
			fmt.Printf("Error: SAT solver became UNSAT at a wrong place (while adding new CNF).\n")
		}
		Prev = This
	}
	Vec_IntFree(vClas)
	Vec_IntFree(vLits)
}
func Acb_NtkWindow2Cnf(p *Acb_Ntk_t, vWinObjs *Vec_Int_t, Pivot int32) *Cnf_Dat_t {
	var pCnf *Cnf_Dat_t
	var vFaninVars *Vec_Int_t = Vec_IntAlloc(8)
	var PivotVar int32 = Vec_IntFind(vWinObjs, Abc_Var2Lit(Pivot, 0))
	var nRoots int32 = Acb_NtkCountRoots(vWinObjs, PivotVar)
	var TfoStart int32 = PivotVar + 1
	var nTfoSize int32 = Vec_IntSize(vWinObjs) - TfoStart
	var nVarsAll int32 = Vec_IntSize(vWinObjs) + nTfoSize + nRoots
	var i int32
	var k int32
	var iObj int32
	var iObjLit int32
	var iFanin int32
	var pFanins []int32
	var vCnfs []Vec_Str_t = Acb_DeriveCnfForWindow(p, vWinObjs, PivotVar)
	var vClas *Vec_Int_t = Vec_IntAlloc(100)
	var vLits *Vec_Int_t = Vec_IntAlloc(1000)
	// mark new SAT variables
	for i = 0; i < Vec_IntSize(vWinObjs); i++ {
		iObj = Vec_IntEntry(vWinObjs, i)
		Acb_ObjSetFunc(p, Abc_Lit2Var(iObj), i)
	}
	// add clauses for all nodes
	Vec_IntPush(vClas, Vec_IntSize(vLits))
	for i = 0; i < Vec_IntSize(vWinObjs); i++ {
		iObjLit = Vec_IntEntry(vWinObjs, i)
		if Abc_LitIsCompl(iObjLit) != 0 && i < PivotVar {
			continue
		}
		iObj = Abc_Lit2Var(iObjLit)
		assert(Acb_ObjIsCio(p, iObj) == 0, "Acb_ObjIsCio(p, iObj) == 0")
		// collect SAT variables
		Vec_IntClear(vFaninVars)
		for k = 0; k < Acb_ObjFaninNum(p, iObj); k++ {
			pFanins = Acb_ObjFanins(p, iObj)
			iFanin = pFanins[k+1]
			Vec_IntPush(vFaninVars, Acb_ObjFunc(p, iFanin))
		}
		Vec_IntPush(vFaninVars, Acb_ObjFunc(p, iObj))
		// derive CNF for the node
		Acb_TranslateCnf(vClas, vLits, &vCnfs[iObj], vFaninVars, -1)
	}
	// add second clauses for the TFO
	for i = TfoStart; i < Vec_IntSize(vWinObjs); i++ {
		iObjLit = Vec_IntEntry(vWinObjs, i)
		iObj = Abc_Lit2Var(iObjLit)
		assert(Acb_ObjIsCio(p, iObj) == 0, "Acb_ObjIsCio(p, iObj) == 0")
		// collect SAT variables
		Vec_IntClear(vFaninVars)
		for k = 0; k < Acb_ObjFaninNum(p, iObj); k++ {
			pFanins = Acb_ObjFanins(p, iObj)
			iFanin = pFanins[k+1]
			Vec_IntPush(vFaninVars, Acb_ObjFunc(p, iFanin)+boolInt(Acb_ObjFunc(p, iFanin) > PivotVar, nTfoSize, 0))
		}
		Vec_IntPush(vFaninVars, Acb_ObjFunc(p, iObj)+nTfoSize)
		// derive CNF for the node
		Acb_TranslateCnf(vClas, vLits, &vCnfs[iObj], vFaninVars, PivotVar)
	}
	if nRoots > 0 {
		// create XOR clauses for the roots
		var nVars int32 = Vec_IntSize(vWinObjs) + nTfoSize
		Vec_IntClear(vFaninVars)
		for i = TfoStart; i < Vec_IntSize(vWinObjs); i++ {
			iObjLit = Vec_IntEntry(vWinObjs, i)
			if Abc_LitIsCompl(iObjLit) == 0 {
				continue
			}
			iObj = Abc_Lit2Var(iObjLit)
			// add clauses
			Vec_IntPushThree(vLits, Abc_Var2Lit(Acb_ObjFunc(p, iObj), 1), Abc_Var2Lit(Acb_ObjFunc(p, iObj)+nTfoSize, 0), Abc_Var2Lit(nVars, 0))
			Vec_IntPush(vClas, Vec_IntSize(vLits))
			Vec_IntPushThree(vLits, Abc_Var2Lit(Acb_ObjFunc(p, iObj), 0), Abc_Var2Lit(Acb_ObjFunc(p, iObj)+nTfoSize, 1), Abc_Var2Lit(nVars, 0))
			Vec_IntPush(vClas, Vec_IntSize(vLits))
			Vec_IntPushThree(vLits, Abc_Var2Lit(Acb_ObjFunc(p, iObj), 0), Abc_Var2Lit(Acb_ObjFunc(p, iObj)+nTfoSize, 0), Abc_Var2Lit(nVars, 1))
			Vec_IntPush(vClas, Vec_IntSize(vLits))
			Vec_IntPushThree(vLits, Abc_Var2Lit(Acb_ObjFunc(p, iObj), 1), Abc_Var2Lit(Acb_ObjFunc(p, iObj)+nTfoSize, 1), Abc_Var2Lit(nVars, 1))
			Vec_IntPush(vClas, Vec_IntSize(vLits))
			Vec_IntPush(vFaninVars, Abc_Var2Lit(nVars, 0))
			nVars++
		}
		Vec_IntAppend(vLits, vFaninVars)
		Vec_IntPush(vClas, Vec_IntSize(vLits))
		assert(nRoots == Vec_IntSize(vFaninVars), "nRoots == Vec_IntSize(vFaninVars)")
		assert(nVars == nVarsAll, "nVars == nVarsAll")
	}
	Vec_IntFree(vFaninVars)
	// create CNF structure from the source format's packed literals and offsets.
	pCnf = new(Cnf_Dat_t)
	Cnf_DataSetPackedClauses(pCnf, nVarsAll, Vec_IntArray(vClas), Vec_IntArray(vLits))
	// cleanup
	Vec_IntFree(vClas)
	Vec_IntFree(vLits)
	//Cnf_DataPrint( pCnf, 1 );
	return pCnf
}
func Acb_NtkWindowUndo(p *Acb_Ntk_t, vWin *Vec_Int_t) {
	var i, iObj int32
	for i = 0; i < Vec_IntSize(vWin); i++ {
		iObj = Vec_IntEntry(vWin, i)
		assert(Vec_IntEntry(&p.vObjFunc, Abc_Lit2Var(iObj)) != -1, "Vec_IntEntry(&p.vObjFunc, Abc_Lit2Var(iObj)) != -1")
		Vec_IntWriteEntry(&p.vObjFunc, Abc_Lit2Var(iObj), -1)
	}
}

/*
*Function*************************************************************

	Synopsis    [Creates SAT solver containing several copies of the window.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkWindow2Solver(pSat *sat_solver, pCnf *Cnf_Dat_t, vFlip *Vec_Int_t, PivotVar int32, nDivs int32, nTimes int32) int32 {
	var n int32
	var i int32
	var RetValue int32
	var Test int32 = Cnf_DataClause(pCnf, 0)[0]
	var nGroups int32
	if nTimes <= 2 {
		nGroups = nTimes - 1
	} else {
		nGroups = 2
	}
	var nRounds int32
	if nTimes <= 2 {
		nRounds = nTimes - 1
	} else {
		nRounds = nTimes
	}
	assert(sat_solver_nvars(pSat) == 0, "sat_solver_nvars(pSat) == 0")
	sat_solver_setnvars(pSat, nTimes*Cnf_DataVarNum(pCnf)+nGroups*nDivs+2)
	assert(nTimes == 1 || nTimes == 2 || nTimes == 6, "nTimes == 1 || nTimes == 2 || nTimes == 6")
	for n = 0; n < nTimes; n++ {
		if n&1 != 0 {
			Cnf_DataLiftAndFlipLits(pCnf, -Cnf_DataVarNum(pCnf), vFlip)
		}
		for i = 0; i < Cnf_DataClauseNum(pCnf); i++ {
			if !sat_solver_addclause(pSat, Cnf_DataClause(pCnf, i)) {
				fmt.Printf("Error: SAT solver became UNSAT at a wrong place.\n")
			}
		}
		if n&1 != 0 {
			Cnf_DataLiftAndFlipLits(pCnf, Cnf_DataVarNum(pCnf), vFlip)
		}
		if n < nTimes-1 {
			Cnf_DataLift(pCnf, Cnf_DataVarNum(pCnf))
		} else if n != 0 { // if ( n == nTimes - 1 )
			Cnf_DataLift(pCnf, -(nTimes-1)*Cnf_DataVarNum(pCnf))
		}
	}
	assert(Test == Cnf_DataClause(pCnf, 0)[0], "Test == Cnf_DataClause(pCnf, 0)[0]")
	// add conditional buffers
	for n = 0; n < nRounds; n++ {
		var BaseA int32 = n * Cnf_DataVarNum(pCnf)
		var BaseB int32 = ((n + 1) % nTimes) * Cnf_DataVarNum(pCnf)
		var BaseC int32 = nTimes*Cnf_DataVarNum(pCnf) + (n&1)*nDivs
		for i = 0; i < nDivs; i++ {
			sat_solver_add_buffer_enable(pSat, BaseA+i, BaseB+i, BaseC+i, 0)
		}
	}
	// finalize
	RetValue = sat_solver_simplify(pSat)
	if !(RetValue != 0) {
		fmt.Printf("Error: SAT solver became UNSAT at a wrong place.\n")
	}
	return 1
}

/*
*Function*************************************************************

	Synopsis    [Computes function of the node]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ComputeFunction(pSat *sat_solver, PivotVar int32, FreeVar int32, vDivVars *Vec_Int_t, fCompl int32) word {
	var fExpand int32
	var uCube, uTruth word
	var vTempLits *Vec_Int_t = Vec_IntAlloc(100)
	var status int32
	var i int32
	var iVar int32
	var iLit int32
	var nFinal int32
	var pFinal []int32
	var pLits [2]int32
	assert(FreeVar < sat_solver_nvars(pSat), "FreeVar < sat_solver_nvars(pSat)")
	//    if ( fCompl )
	//        pLits[0] = Abc_Var2Lit( sat_solver_nvars(pSat)-2, 0 ); // F = 1
	//    else
	pLits[0] = Abc_Var2Lit(PivotVar, fCompl) // F = 1
	pLits[1] = Abc_Var2Lit(FreeVar, 0)       // iNewLit
	for {
		// find onset minterm
		status = sat_solver_solve(pSat, pLits[:], nil, 0, 0, 0, 0)
		if status == l_False {
			Vec_IntFree(vTempLits)
			return uTruth
		}
		assert(status == l_True, "status == l_True")
		if fExpand != 0 {
			// collect divisor literals
			Vec_IntFill(vTempLits, 1, Abc_LitNot(pLits[0])) // F = 0
			for i = 0; i < Vec_IntSize(vDivVars); i++ {
				iVar = Vec_IntEntry(vDivVars, i)
				Vec_IntPush(vTempLits, sat_solver_var_literal(pSat, iVar))
			}
			// check against offset
			status = sat_solver_solve(pSat, Vec_IntArray(vTempLits)[:Vec_IntSize(vTempLits)], nil, 0, 0, 0, 0)
			if status != l_False {
				fmt.Printf("Failed internal check during function comptutation.\n")
			}
			assert(status == l_False, "status == l_False")
			// compute cube and add clause
			nFinal = sat_solver_final(pSat, &pFinal)
			Vec_IntFill(vTempLits, 1, Abc_LitNot(pLits[1])) // NOT(iNewLit)
			for i = 0; i < nFinal; i++ {
				if pFinal[i] != pLits[0] {
					Vec_IntPush(vTempLits, pFinal[i])
				}
			}

		} else {
			// collect divisor literals
			Vec_IntFill(vTempLits, 1, Abc_LitNot(pLits[1])) // NOT(iNewLit)
			for i = 0; i < Vec_IntSize(vDivVars); i++ {
				iVar = Vec_IntEntry(vDivVars, i)
				Vec_IntPush(vTempLits, Abc_LitNot(sat_solver_var_literal(pSat, iVar)))
			}
		}
		uCube = ^word(0)
		for i = 1; i < Vec_IntSize(vTempLits); i++ {
			iLit = Vec_IntEntry(vTempLits, i)
			iVar = Vec_IntFind(vDivVars, Abc_Lit2Var(iLit))
			assert(iVar >= 0, "iVar >= 0")
			if Abc_LitIsCompl(iLit) != 0 {
				uCube &= s_Truths6[iVar]
			} else {
				uCube &= ^s_Truths6[iVar]
			}
		}
		uTruth |= uCube
		if !sat_solver_addclause(pSat, Vec_IntArray(vTempLits)[:Vec_IntSize(vTempLits)]) {
			Vec_IntFree(vTempLits)
			return uTruth
		}
	}
	Vec_IntFree(vTempLits)
	assert(false, "0")
	return ^uint64(0)
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkPrintVec(p *Acb_Ntk_t, vVec *Vec_Int_t, pName string) {
	var i int32
	fmt.Printf("%s: ", pName)
	for i = 0; i < Vec_IntSize(vVec); i++ {
		fmt.Printf("%d ", Vec_IntEntry(vVec, i))
	}
	fmt.Printf("\n")
}
func Acb_NtkPrintVec2(p *Acb_Ntk_t, vVec *Vec_Int_t, pName string) {
	var i int32
	fmt.Printf("%s: \n", pName)
	for i = 0; i < Vec_IntSize(vVec); i++ {
		Acb_NtkPrintNode(p, Vec_IntEntry(vVec, i))
	}
	fmt.Printf("\n")
}
func Acb_NtkPrintVecWin(p *Acb_Ntk_t, vVec *Vec_Int_t, pName string) {
	var i int32
	fmt.Printf("%s: \n", pName)
	for i = 0; i < Vec_IntSize(vVec); i++ {
		Acb_NtkPrintNode(p, Abc_Lit2Var(Vec_IntEntry(vVec, i)))
	}
	fmt.Printf("\n")
}

/*
*Function*************************************************************

	Synopsis    [Collects divisors in a non-topo order.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkDivisors_rec(p *Acb_Ntk_t, iObj int32, nTfiLevMin int32, vDivs *Vec_Int_t) {
	var k int32
	var iFanin int32
	var pFanins []int32
	//    if ( Acb_ObjIsCi(p, iObj) == 0 && Acb_ObjLevelD(p, iObj) < nTfiLevMin )
	if Acb_ObjIsCi(p, iObj) == 0 && nTfiLevMin < 0 {
		return
	}
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return
	}
	for k = 0; k < Acb_ObjFaninNum(p, iObj); k++ {
		pFanins = Acb_ObjFanins(p, iObj)
		iFanin = pFanins[k+1]
		Acb_NtkDivisors_rec(p, iFanin, nTfiLevMin-1, vDivs)
	}
	Vec_IntPush(vDivs, iObj)
}
func Acb_NtkDivisors(p *Acb_Ntk_t, Pivot int32, nTfiLevMin int32, fDelay int32) *Vec_Int_t {
	var k int32
	var iFanin int32
	var pFanins []int32
	var vDivs *Vec_Int_t = Vec_IntAlloc(100)
	Acb_NtkIncTravId(p)
	//    if ( fDelay ) // delay-oriented
	if false { // delay-oriented

		// start from critical fanins
		assert(Acb_ObjLevelD(p, Pivot) > 1, "Acb_ObjLevelD( p, Pivot ) > 1")
		for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
			pFanins = Acb_ObjFanins(p, Pivot)
			iFanin = pFanins[k+1]
			if Acb_ObjIsDelayCriticalFanin(p, Pivot, iFanin) != 0 {
				Acb_NtkDivisors_rec(p, iFanin, nTfiLevMin, vDivs)
			}
		}
		// add non-critical fanins
		for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
			pFanins = Acb_ObjFanins(p, Pivot)
			iFanin = pFanins[k+1]
			if Acb_ObjIsDelayCriticalFanin(p, Pivot, iFanin) == 0 && Acb_ObjSetTravIdCur(p, iFanin) == 0 {
				Vec_IntPush(vDivs, iFanin)
			}
		}

	} else {
		Acb_NtkDivisors_rec(p, Pivot, nTfiLevMin, vDivs)
		assert(Vec_IntEntryLast(vDivs) == Pivot, "Vec_IntEntryLast(vDivs) == Pivot")
		Vec_IntPop(vDivs)
		// add remaining fanins of the node
		for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
			pFanins = Acb_ObjFanins(p, Pivot)
			iFanin = pFanins[k+1]
			if Acb_ObjSetTravIdCur(p, iFanin) == 0 {
				Vec_IntPush(vDivs, iFanin)
			}
		}
		/*
		   // start from critical fanins
		   assert( Acb_ObjLevelD( p, Pivot ) > 1 );
		   Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
		       if ( Acb_ObjIsAreaCritical( p, iFanin ) )
		           Acb_NtkDivisors_rec( p, iFanin, nTfiLevMin, vDivs );
		   // add non-critical fanins
		   Acb_ObjForEachFaninFast( p, Pivot, pFanins, iFanin, k )
		       if ( Acb_ObjIsAreaCritical( p, iFanin ) == 0 )
		           if ( Acb_ObjSetTravIdCur(p, iFanin) == 0 )
		               Vec_IntPush( vDivs, iFanin );
		*/
	}
	return vDivs
}

/*
*Function*************************************************************

	Synopsis    [Marks TFO of divisors.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjMarkTfo_rec(p *Acb_Ntk_t, iObj int32, nTfoLevMax int32, nFanMax int32, vMarked *Vec_Int_t) {
	var iFanout, i int32
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return
	}
	Vec_IntPush(vMarked, iObj)
	if Acb_ObjLevelD(p, iObj) > nTfoLevMax || Acb_ObjFanoutNum(p, iObj) > nFanMax {
		return
	}
	for i = 0; i < Acb_ObjFanoutNum(p, iObj); i++ {
		iFanout = Acb_ObjFanout(p, iObj, i)
		if Acb_ObjType(p, iFanout) != 0 {
			Acb_ObjMarkTfo_rec(p, iFanout, nTfoLevMax, nFanMax, vMarked)
		}
	}
}
func Acb_ObjMarkTfo(p *Acb_Ntk_t, vDivs *Vec_Int_t, Pivot int32, nTfoLevMax int32, nFanMax int32) *Vec_Int_t {
	var vMarked *Vec_Int_t = Vec_IntAlloc(1000)
	var i, iObj int32
	Acb_NtkIncTravId(p)
	Acb_ObjSetTravIdCur(p, Pivot)
	Vec_IntPush(vMarked, Pivot)
	for i = 0; i < Vec_IntSize(vDivs); i++ {
		iObj = Vec_IntEntry(vDivs, i)
		Acb_ObjMarkTfo_rec(p, iObj, nTfoLevMax, nFanMax, vMarked)
	}
	return vMarked
}
func Acb_ObjMarkTfo2(p *Acb_Ntk_t, vMarked *Vec_Int_t) {
	var i, Node int32
	Acb_NtkIncTravId(p)
	for i = 0; i < Vec_IntSize(vMarked); i++ {
		Node = Vec_IntEntry(vMarked, i)
		Acb_ObjSetTravIdCur(p, Node)
	}
}

/*
*Function*************************************************************

	Synopsis    [Labels TFO nodes with {none, root, inner} based on their type.]

	Description [Assuming TFO of TFI is marked with the current trav ID.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjLabelTfo_rec(p *Acb_Ntk_t, iObj int32, nTfoLevMax int32, nFanMax int32, fFirst int32) int32 {
	var iFanout, i, Diff, fHasNone int32
	Diff = Acb_ObjTravIdDiff(p, iObj)
	if Diff <= 2 {
		return Diff
	}
	Acb_ObjSetTravIdDiff(p, iObj, 2)
	if Acb_ObjIsCo(p, iObj) != 0 || Acb_ObjLevelD(p, iObj) > nTfoLevMax {
		return 2
	}
	if Acb_ObjLevelD(p, iObj) == nTfoLevMax || Acb_ObjFanoutNum(p, iObj) > nFanMax {
		if Diff == 3 { // belongs to TFO of TFI
			Acb_ObjSetTravIdDiff(p, iObj, 1)
		} // root
		return Acb_ObjTravIdDiff(p, iObj)
	}
	for i = 0; i < Acb_ObjFanoutNum(p, iObj); i++ {
		iFanout = Acb_ObjFanout(p, iObj, i)
		if Acb_ObjType(p, iFanout) != 0 && (fFirst == 0 || Acb_ObjIsDelayCriticalFanin(p, iFanout, iObj) != 0) {
			if Acb_ObjLabelTfo_rec(p, iFanout, nTfoLevMax, nFanMax, 0) == 2 {
				fHasNone |= 2
			}
		}
	}
	if fHasNone != 0 && Diff == 3 { // belongs to TFO of TFI
		Acb_ObjSetTravIdDiff(p, iObj, 1) // root
	} else if fHasNone == 0 {
		Acb_ObjSetTravIdDiff(p, iObj, 0)
	} // inner
	return Acb_ObjTravIdDiff(p, iObj)
}
func Acb_ObjLabelTfo(p *Acb_Ntk_t, Root int32, nTfoLevMax int32, nFanMax int32, fDelay int32) int32 {
	Acb_NtkIncTravId(p) // none  (2)    marked (3)  unmarked (4)
	Acb_NtkIncTravId(p) // root  (1)
	Acb_NtkIncTravId(p) // inner (0)
	assert(Acb_ObjTravIdDiff(p, Root) > 2, "Acb_ObjTravIdDiff(p, Root) > 2")
	return Acb_ObjLabelTfo_rec(p, Root, nTfoLevMax, nFanMax, fDelay)
}

/*
*Function*************************************************************

	Synopsis    [Collects labeled TFO.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_ObjDeriveTfo_rec(p *Acb_Ntk_t, iObj int32, vTfo *Vec_Int_t, vRoots *Vec_Int_t, fFirst int32) {
	var iFanout int32
	var i int32
	var Diff int32 = Acb_ObjTravIdDiff(p, iObj)
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return
	}
	if Diff == 2 { // root

		Vec_IntPush(vRoots, iObj)
		Vec_IntPush(vTfo, iObj)
		return
	}
	assert(Diff == 1, "Diff == 1")
	for i = 0; i < Acb_ObjFanoutNum(p, iObj); i++ {
		iFanout = Acb_ObjFanout(p, iObj, i)
		if Acb_ObjType(p, iFanout) != 0 && (fFirst == 0 || Acb_ObjIsDelayCriticalFanin(p, iFanout, iObj) != 0) {
			Acb_ObjDeriveTfo_rec(p, iFanout, vTfo, vRoots, 0)
		}
	}
	Vec_IntPush(vTfo, iObj)
}
func Acb_ObjDeriveTfo(p *Acb_Ntk_t, Pivot int32, nTfoLevMax int32, nFanMax int32, pvTfo **Vec_Int_t, pvRoots **Vec_Int_t, fDelay int32) {
	var Res int32 = Acb_ObjLabelTfo(p, Pivot, nTfoLevMax, nFanMax, fDelay)
	vTfo := Vec_IntAlloc(10)
	vRoots := Vec_IntAlloc(10)
	*pvTfo = vTfo
	*pvRoots = vRoots
	if Res != 0 { // none or root
		return
	}
	Acb_NtkIncTravId(p) // root (2)   inner (1)  visited (0)
	Acb_ObjDeriveTfo_rec(p, Pivot, vTfo, vRoots, fDelay)
	assert(Vec_IntEntryLast(vTfo) == Pivot, "Vec_IntEntryLast(vTfo) == Pivot")
	Vec_IntPop(vTfo)
	assert(Vec_IntEntryLast(vRoots) != Pivot, "Vec_IntEntryLast(vRoots) != Pivot")
	Vec_IntReverseOrder(vTfo)
	Vec_IntReverseOrder(vRoots)
}

/*
*Function*************************************************************

	Synopsis    [Collect side-inputs of the TFO, except the node.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkCollectTfoSideInputs(p *Acb_Ntk_t, Pivot int32, vTfo *Vec_Int_t) *Vec_Int_t {
	var vSide *Vec_Int_t = Vec_IntAlloc(100)
	var i int32
	var k int32
	var Node int32
	var iFanin int32
	var pFanins []int32
	Acb_NtkIncTravId(p)
	Vec_IntPush(vTfo, Pivot)
	for i = 0; i < Vec_IntSize(vTfo); i++ {
		Node = Vec_IntEntry(vTfo, i)
		Acb_ObjSetTravIdCur(p, Node)
	}
	for i = 0; i < Vec_IntSize(vTfo); i++ {
		Node = Vec_IntEntry(vTfo, i)
		pFanins = Acb_ObjFanins(p, Node)
		for k = 0; k < Acb_ObjFaninNum(p, Node); k++ {
			iFanin = pFanins[k+1]
			if Acb_ObjSetTravIdCur(p, iFanin) == 0 && iFanin != Pivot {
				Vec_IntPush(vSide, iFanin)
			}
		}
	}
	Vec_IntPop(vTfo)
	return vSide
}

/*
*Function*************************************************************

	Synopsis    [From side inputs, collect marked nodes and their unmarked fanins.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkCollectNewTfi1_rec(p *Acb_Ntk_t, iObj int32, vTfiNew *Vec_Int_t) {
	var i int32
	var iFanin int32
	var pFanins []int32
	if Acb_ObjIsTravIdPrev(p, iObj) == 0 {
		return
	}
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return
	}
	for i = 0; i < Acb_ObjFaninNum(p, iObj); i++ {
		pFanins = Acb_ObjFanins(p, iObj)
		iFanin = pFanins[i+1]
		Acb_NtkCollectNewTfi1_rec(p, iFanin, vTfiNew)
	}
	Vec_IntPush(vTfiNew, iObj)
}
func Acb_NtkCollectNewTfi2_rec(p *Acb_Ntk_t, iObj int32, vTfiNew *Vec_Int_t) {
	var i int32
	var iFanin int32
	var pFanins []int32
	var fTravIdPrev int32 = Acb_ObjIsTravIdPrev(p, iObj)
	if Acb_ObjSetTravIdCur(p, iObj) != 0 {
		return
	}
	if (fTravIdPrev != 0) && Acb_ObjIsCi(p, iObj) == 0 {
		for i = 0; i < Acb_ObjFaninNum(p, iObj); i++ {
			pFanins = Acb_ObjFanins(p, iObj)
			iFanin = pFanins[i+1]
			Acb_NtkCollectNewTfi2_rec(p, iFanin, vTfiNew)
		}
	}
	Vec_IntPush(vTfiNew, iObj)
}
func Acb_NtkCollectNewTfi(p *Acb_Ntk_t, Pivot int32, vDivs *Vec_Int_t, vSide *Vec_Int_t, pnDivs *int32) *Vec_Int_t {
	var vTfiNew *Vec_Int_t = Vec_IntAlloc(100)
	var i, Node int32
	Acb_NtkIncTravId(p)
	//Acb_NtkPrintVec( p, vDivs, "vDivs" );
	for i = 0; i < Vec_IntSize(vDivs); i++ {
		Node = Vec_IntEntry(vDivs, i)
		Acb_NtkCollectNewTfi1_rec(p, Node, vTfiNew)
	}
	//Acb_NtkPrintVec( p, vTfiNew, "vTfiNew" );
	Acb_NtkCollectNewTfi1_rec(p, Pivot, vTfiNew)
	//Acb_NtkPrintVec( p, vTfiNew, "vTfiNew" );
	assert(Vec_IntEntryLast(vTfiNew) == Pivot, "Vec_IntEntryLast(vTfiNew) == Pivot")
	Vec_IntPop(vTfiNew)
	/*
	   Vec_IntForEachEntry( vDivs, Node, i )
	   {
	       Acb_ObjSetTravIdCur( p, Node );
	       Vec_IntPush( vTfiNew, Node );
	   }
	*/
	*pnDivs = Vec_IntSize(vTfiNew)
	for i = 0; i < Vec_IntSize(vSide); i++ {
		Node = Vec_IntEntry(vSide, i)
		Acb_NtkCollectNewTfi2_rec(p, Node, vTfiNew)
	}
	Vec_IntPush(vTfiNew, Pivot)
	return vTfiNew
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkCollectWindow(p *Acb_Ntk_t, Pivot int32, vTfi *Vec_Int_t, vTfo *Vec_Int_t, vRoots *Vec_Int_t) *Vec_Int_t {
	var vWin *Vec_Int_t = Vec_IntAlloc(100)
	var i int32
	var k int32
	var iObj int32
	var iFanin int32
	var pFanins []int32
	assert(Vec_IntEntryLast(vTfi) == Pivot, "Vec_IntEntryLast(vTfi) == Pivot")
	// mark nodes
	Acb_NtkIncTravId(p)
	for i = 0; i < Vec_IntSize(vTfi); i++ {
		iObj = Vec_IntEntry(vTfi, i)
		Acb_ObjSetTravIdCur(p, iObj)
	}
	// add TFI
	for i = 0; i < Vec_IntSize(vTfi); i++ {
		iObj = Vec_IntEntry(vTfi, i)
		var fIsTfiInput int32
		for k = 0; k < Acb_ObjFaninNum(p, iObj); k++ {
			pFanins = Acb_ObjFanins(p, iObj)
			iFanin = pFanins[k+1]
			if Acb_ObjIsTravIdCur(p, iFanin) == 0 {
				fIsTfiInput = 1
			}
		}
		Vec_IntPush(vWin, Abc_Var2Lit(iObj, boolInt(Acb_ObjIsCi(p, iObj) != 0 || fIsTfiInput != 0, 1, 0)))
	}
	// mark roots
	Acb_NtkIncTravId(p)
	for i = 0; i < Vec_IntSize(vRoots); i++ {
		iObj = Vec_IntEntry(vRoots, i)
		Acb_ObjSetTravIdCur(p, iObj)
	}
	// add TFO
	for i = 0; i < Vec_IntSize(vTfo); i++ {
		iObj = Vec_IntEntry(vTfo, i)
		assert(Acb_ObjIsCo(p, iObj) == 0, "Acb_ObjIsCo(p, iObj) == 0")
		Vec_IntPush(vWin, Abc_Var2Lit(iObj, Acb_ObjIsTravIdCur(p, iObj)))
	}
	return vWin
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkWindow(p *Acb_Ntk_t, Pivot int32, nTfiLevs int32, nTfoLevs int32, nFanMax int32, fDelay int32, pnDivs *int32) *Vec_Int_t {
	var fVerbose int32
	//int nTfiLevMin = Acb_ObjLevelD(p, Pivot) - nTfiLevs;
	var nTfoLevMax int32 = Acb_ObjLevelD(p, Pivot) + nTfoLevs
	var vWin, vDivs, vMarked, vTfo, vRoots, vSide, vTfi *Vec_Int_t
	// collect divisors by traversing limited TFI
	vDivs = Acb_NtkDivisors(p, Pivot, nTfiLevs, fDelay)
	if fVerbose != 0 {
		Acb_NtkPrintVec(p, vDivs, "vDivs")
	}
	// mark limited TFO of the divisors
	vMarked = Acb_ObjMarkTfo(p, vDivs, Pivot, nTfoLevMax, nFanMax)
	// collect TFO and roots
	Acb_ObjDeriveTfo(p, Pivot, nTfoLevMax, nFanMax, &vTfo, &vRoots, 0) //fDelay );
	if fVerbose != 0 {
		Acb_NtkPrintVec(p, vTfo, "vTfo")
	}
	if fVerbose != 0 {
		Acb_NtkPrintVec(p, vRoots, "vRoots")
	}
	// collect side inputs of the TFO
	vSide = Acb_NtkCollectTfoSideInputs(p, Pivot, vTfo)
	if fVerbose != 0 {
		Acb_NtkPrintVec(p, vSide, "vSide")
	}
	// mark limited TFO of the divisors
	//Acb_ObjMarkTfo( p, vDivs, Pivot, nTfoLevMax, nFanMax );
	Acb_ObjMarkTfo2(p, vMarked)
	Vec_IntFree(vMarked)
	// collect new TFI
	vTfi = Acb_NtkCollectNewTfi(p, Pivot, vDivs, vSide, pnDivs)
	if fVerbose != 0 {
		Acb_NtkPrintVec(p, vTfi, "vTfi")
	}
	Vec_IntFree(vSide)
	Vec_IntFree(vDivs)
	// collect all nodes
	vWin = Acb_NtkCollectWindow(p, Pivot, vTfi, vTfo, vRoots)
	// cleanup
	Vec_IntFree(vTfi)
	Vec_IntFree(vTfo)
	Vec_IntFree(vRoots)
	return vWin
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Vec_IntVars2Vars(p *Vec_Int_t, Shift int32) {
	var i int32
	for i = 0; i < Vec_IntSize(p); i++ {
		Vec_IntWriteEntry(p, i, Vec_IntEntry(p, i)+Shift)
	}
}
func Vec_IntVars2Lits(p *Vec_Int_t, Shift int32, fCompl int32) {
	var i int32
	for i = 0; i < Vec_IntSize(p); i++ {
		Vec_IntWriteEntry(p, i, Abc_Var2Lit(Vec_IntEntry(p, i)+Shift, fCompl))
	}
}
func Vec_IntLits2Vars(p *Vec_Int_t, Shift int32) {
	var i int32
	for i = 0; i < Vec_IntSize(p); i++ {
		Vec_IntWriteEntry(p, i, Abc_Lit2Var(Vec_IntEntry(p, i))+Shift)
	}
}
func Vec_IntRemap(p *Vec_Int_t, vMap *Vec_Int_t) {
	var i int32
	for i = 0; i < Vec_IntSize(p); i++ {
		Vec_IntWriteEntry(p, i, Vec_IntEntry(vMap, Vec_IntEntry(p, i)))
	}
}

func Acb_WinPrint(p *Acb_Ntk_t, vWin *Vec_Int_t, Pivot int32, nDivs int32) {
	var i, Node int32
	fmt.Printf("Window for node %d with %d divisors:\n", Pivot, nDivs)
	for i = 0; i < Vec_IntSize(vWin); i++ {
		Node = Vec_IntEntry(vWin, i)
		if i == nDivs {
			fmt.Printf(" | ")
		}
		if Abc_Lit2Var(Node) == Pivot {
			fmt.Printf("(%d) ", Pivot)
		} else {
			fmt.Printf("%s%d ", map[bool]string{true: "*", false: ""}[Abc_LitIsCompl(Node) != 0], Abc_Lit2Var(Node))
		}
	}
	fmt.Printf("\n")
}

func Acb_NtkOrderByRefCount(p *Acb_Ntk_t, vSupp *Vec_Int_t) {
	var i int32
	var j int32
	var best_i int32
	var nSize int32 = Vec_IntSize(vSupp)
	var pArray []int32 = Vec_IntArray(vSupp)
	for i = 0; i < nSize-1; i++ {
		best_i = i
		for j = i + 1; j < nSize; j++ {
			if Acb_ObjFanoutNum(p, pArray[j]) > Acb_ObjFanoutNum(p, pArray[best_i]) {
				best_i = j
			}
		}
		pArray[i], pArray[best_i] = pArray[best_i], pArray[i]
	}
}

func Acb_NtkRemapIntoSatVariables(p *Acb_Ntk_t, vSupp *Vec_Int_t) {
	var k, iFanin int32
	for k = 0; k < Vec_IntSize(vSupp); k++ {
		iFanin = Vec_IntEntry(vSupp, k)
		assert(Acb_ObjFunc(p, iFanin) >= 0, "Acb_ObjFunc(p, iFanin) >= 0")
		Vec_IntWriteEntry(vSupp, k, Acb_ObjFunc(p, iFanin))
	}
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkFindSupp1(p *Acb_Ntk_t, Pivot int32, pSat *sat_solver, nVars int32, nDivs int32, vWin *Vec_Int_t, vSupp *Vec_Int_t) int32 {
	var nSuppNew int32
	var status int32
	var k int32
	var iFanin int32
	var pFanins []int32
	Vec_IntClear(vSupp)
	for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
		pFanins = Acb_ObjFanins(p, Pivot)
		iFanin = pFanins[k+1]
		Vec_IntPush(vSupp, iFanin)
	}
	Acb_NtkOrderByRefCount(p, vSupp)
	Acb_NtkRemapIntoSatVariables(p, vSupp)
	Vec_IntVars2Lits(vSupp, 2*nVars, 0)
	status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
	if status != l_False {
		fmt.Printf("Failed internal check at node %d.\n", Pivot)
	}
	assert(status == l_False, "status == l_False")
	nSuppNew = sat_solver_minimize_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0)
	Vec_IntShrink(vSupp, nSuppNew)
	Vec_IntLits2Vars(vSupp, -2*nVars)
	return boolInt(Vec_IntSize(vSupp) < Acb_ObjFaninNum(p, Pivot), 1, 0)
}

var StrCount int32

func Acb_NtkFindSupp2(p *Acb_Ntk_t, Pivot int32, pSat *sat_solver, nVars int32, nDivs int32, vWin *Vec_Int_t, vSupp *Vec_Int_t, nLutSize int32, fDelay int32) int32 {
	var nSuppNew int32
	var status int32
	var k int32
	var iFanin int32
	var pFanins []int32
	var k2 int32
	var iFanin2 int32
	var pFanins2 []int32
	for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
		pFanins = Acb_ObjFanins(p, Pivot)
		iFanin = pFanins[k+1]
		assert(Acb_ObjFunc(p, iFanin) >= 0 && Acb_ObjFunc(p, iFanin) < nDivs, "Acb_ObjFunc(p, iFanin) >= 0 && Acb_ObjFunc(p, iFanin) < nDivs")
	}
	if fDelay != 0 {
		// add non-timing-critical fanins
		var nNonCrits int32
		var k2 int32
		var iFanin2 int32
		var pFanins2 []int32
		assert(Acb_ObjLevelD(p, Pivot) > 1, "Acb_ObjLevelD( p, Pivot ) > 1")
		Vec_IntClear(vSupp)
		for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
			pFanins = Acb_ObjFanins(p, Pivot)
			iFanin = pFanins[k+1]
			if Acb_ObjIsDelayCriticalFanin(p, Pivot, iFanin) == 0 {
				Vec_IntPush(vSupp, iFanin)
			}
		}
		nNonCrits = Vec_IntSize(vSupp)
		// add fanins of timing critical fanins
		for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
			pFanins = Acb_ObjFanins(p, Pivot)
			iFanin = pFanins[k+1]
			if Acb_ObjIsDelayCriticalFanin(p, Pivot, iFanin) != 0 {
				for k2 = 0; k2 < Acb_ObjFaninNum(p, iFanin); k2++ {
					pFanins2 = Acb_ObjFanins(p, iFanin)
					iFanin2 = pFanins2[k2+1]
					Vec_IntPushUnique(vSupp, iFanin2)
				}
			}
		}
		assert(nNonCrits < Vec_IntSize(vSupp), "nNonCrits < Vec_IntSize(vSupp)")
		// sort additional fanins by level
		Vec_IntSelectSortCost(Vec_IntArray(vSupp)[nNonCrits:], Vec_IntSize(vSupp)-nNonCrits, &p.vLevelD)
		// translate to SAT vars
		for k = 0; k < Vec_IntSize(vSupp); k++ {
			iFanin = Vec_IntEntry(vSupp, k)
			assert(Acb_ObjFunc(p, iFanin) >= 0, "Acb_ObjFunc(p, iFanin) >= 0")
			Vec_IntWriteEntry(vSupp, k, Acb_ObjFunc(p, iFanin))
		}
		// solve for these fanins
		Vec_IntVars2Lits(vSupp, 2*nVars, 0)
		status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
		if status != l_False {
			fmt.Printf("Failed internal check at node %d.\n", Pivot)
		}
		assert(status == l_False, "status == l_False")
		nSuppNew = sat_solver_minimize_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0)
		Vec_IntShrink(vSupp, nSuppNew)
		Vec_IntLits2Vars(vSupp, -2*nVars)
		return boolInt(Vec_IntSize(vSupp) <= nLutSize, 1, 0)
	}
	// iterate through different fanout free cones
	for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
		pFanins = Acb_ObjFanins(p, Pivot)
		iFanin = pFanins[k+1]
		if !(Acb_ObjIsAreaCritical(p, iFanin) != 0) {
			continue
		}
		// collect fanins of the root node
		Vec_IntClear(vSupp)
		for k2 = 0; k2 < Acb_ObjFaninNum(p, Pivot); k2++ {
			pFanins2 = Acb_ObjFanins(p, Pivot)
			iFanin2 = pFanins2[k2+1]
			if iFanin != iFanin2 {
				Vec_IntPush(vSupp, iFanin2)
			}
		}
		// collect fanins of the selected node
		for k2 = 0; k2 < Acb_ObjFaninNum(p, iFanin); k2++ {
			pFanins2 = Acb_ObjFanins(p, iFanin)
			iFanin2 = pFanins2[k2+1]
			Vec_IntPushUnique(vSupp, iFanin2)
		}
		// sort fanins by level
		Vec_IntSelectSortCost(Vec_IntArray(vSupp), Vec_IntSize(vSupp), &p.vLevelD)
		//Acb_NtkOrderByRefCount( p, vSupp );
		Acb_NtkRemapIntoSatVariables(p, vSupp)
		// solve for these fanins
		Vec_IntVars2Lits(vSupp, 2*nVars, 0)
		status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
		if status != l_False {
			fmt.Printf("Failed internal check at node %d.\n", Pivot)
		}
		assert(status == l_False, "status == l_False")
		nSuppNew = sat_solver_minimize_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0)
		Vec_IntShrink(vSupp, nSuppNew)
		Vec_IntLits2Vars(vSupp, -2*nVars)
		if Vec_IntSize(vSupp) <= nLutSize {
			return 1
		}
	}
	return 0
}

func Acb_NtkFindSupp3(p *Acb_Ntk_t, Pivot int32, pSat *sat_solver, nVars int32, nDivs int32, vWin *Vec_Int_t, vSupp *Vec_Int_t, nLutSize int32, fDelay int32) int32 {
	var nSuppNew int32
	var status int32
	var k int32
	var iFanin int32
	var pFanins []int32
	var k2 int32
	var iFanin2 int32
	var pFanins2 []int32
	var k3 int32
	var iFanin3 int32
	var pFanins3 []int32
	var NodeMark int32

	if fDelay != 0 {
		return 0
	}

	// iterate through pairs of fanins with one fanouts
	for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
		pFanins = Acb_ObjFanins(p, Pivot)
		iFanin = pFanins[k+1]
		if !(Acb_ObjIsAreaCritical(p, iFanin) != 0) {
			continue
		}
		for k2 = 0; k2 < Acb_ObjFaninNum(p, Pivot); k2++ {
			pFanins2 = Acb_ObjFanins(p, Pivot)
			iFanin2 = pFanins2[k2+1]
			if !(Acb_ObjIsAreaCritical(p, iFanin2) != 0) || k2 == k {
				continue
			}
			// iFanin and iFanin2 have 1 fanout
			assert(iFanin != iFanin2, "iFanin != iFanin2")

			// collect fanins of the root node
			Vec_IntClear(vSupp)
			for k3 = 0; k3 < Acb_ObjFaninNum(p, Pivot); k3++ {
				pFanins3 = Acb_ObjFanins(p, Pivot)
				iFanin3 = pFanins3[k3+1]
				if iFanin3 != iFanin && iFanin3 != iFanin2 {
					assert(Acb_ObjFunc(p, iFanin3) >= 0, "Acb_ObjFunc(p, iFanin3) >= 0")
					Vec_IntPush(vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3)+6*nVars, 0))
				}
			}
			NodeMark = Vec_IntSize(vSupp)

			// collect fanins of the second node
			for k3 = 0; k3 < Acb_ObjFaninNum(p, iFanin); k3++ {
				pFanins3 = Acb_ObjFanins(p, iFanin)
				iFanin3 = pFanins3[k3+1]
				assert(Acb_ObjFunc(p, iFanin3) >= 0, "Acb_ObjFunc(p, iFanin3) >= 0")
				Vec_IntPush(vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3)+6*nVars+nDivs, 0))
			}
			// collect fanins of the third node
			for k3 = 0; k3 < Acb_ObjFaninNum(p, iFanin2); k3++ {
				pFanins3 = Acb_ObjFanins(p, iFanin2)
				iFanin3 = pFanins3[k3+1]
				assert(Acb_ObjFunc(p, iFanin3) >= 0, "Acb_ObjFunc(p, iFanin3) >= 0")
				Vec_IntPushUnique(vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3)+6*nVars+nDivs, 0))
			}
			assert(Vec_IntCheckUniqueSmall(vSupp) != 0, "Vec_IntCheckUniqueSmall(vSupp)")

			// sort fanins by level
			//Vec_IntSelectSortCost( Vec_IntArray(vSupp) + NodeMark, Vec_IntSize(vSupp) - NodeMark, &p->vLevelD );
			// solve for these fanins
			status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
			if status != l_False {
				continue
			}
			assert(status == l_False, "status == l_False")
			nSuppNew = sat_solver_minimize_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0)
			Vec_IntShrink(vSupp, nSuppNew)
			Vec_IntLits2Vars(vSupp, -6*nVars)
			Vec_IntSort(vSupp, 1)
			// count how many belong to H; the rest belong to G
			NodeMark = 0
			for k3 = 0; k3 < Vec_IntSize(vSupp); k3++ {
				iFanin3 = Vec_IntEntry(vSupp, k3)
				if iFanin3 >= nDivs {
					Vec_IntWriteEntry(vSupp, k3, iFanin3-nDivs)
				} else {
					NodeMark++
				}
			}
			if NodeMark == 0 {
				//printf( "Obj %d: Special case 1 (vars = %d)\n", Pivot, Vec_IntSize(vSupp) );
				continue
			}
			assert(NodeMark > 0, "NodeMark > 0")
			if Vec_IntSize(vSupp)-NodeMark <= nLutSize {
				return NodeMark
			}
		}
	}

	// iterate through fanins with one fanout and their fanins with one fanout
	for k = 0; k < Acb_ObjFaninNum(p, Pivot); k++ {
		pFanins = Acb_ObjFanins(p, Pivot)
		iFanin = pFanins[k+1]
		if !(Acb_ObjIsAreaCritical(p, iFanin) != 0) {
			continue
		}
		for k2 = 0; k2 < Acb_ObjFaninNum(p, iFanin); k2++ {
			pFanins2 = Acb_ObjFanins(p, iFanin)
			iFanin2 = pFanins2[k2+1]
			if !(Acb_ObjIsAreaCritical(p, iFanin2) != 0) {
				continue
			}
			// iFanin and iFanin2 have 1 fanout
			assert(iFanin != iFanin2, "iFanin != iFanin2")

			// collect fanins of the root node
			Vec_IntClear(vSupp)
			for k3 = 0; k3 < Acb_ObjFaninNum(p, Pivot); k3++ {
				pFanins3 = Acb_ObjFanins(p, Pivot)
				iFanin3 = pFanins3[k3+1]
				if iFanin3 != iFanin && iFanin3 != iFanin2 {
					Vec_IntPush(vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3)+6*nVars, 0))
				}
			}
			NodeMark = Vec_IntSize(vSupp)

			// collect fanins of the second node
			for k3 = 0; k3 < Acb_ObjFaninNum(p, iFanin); k3++ {
				pFanins3 = Acb_ObjFanins(p, iFanin)
				iFanin3 = pFanins3[k3+1]
				if iFanin3 != iFanin2 {
					Vec_IntPush(vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3)+6*nVars+nDivs, 0))
				}
			}
			// collect fanins of the third node
			for k3 = 0; k3 < Acb_ObjFaninNum(p, iFanin2); k3++ {
				pFanins3 = Acb_ObjFanins(p, iFanin2)
				iFanin3 = pFanins3[k3+1]
				assert(Acb_ObjFunc(p, iFanin3) >= 0, "Acb_ObjFunc(p, iFanin3) >= 0")
				Vec_IntPushUnique(vSupp, Abc_Var2Lit(Acb_ObjFunc(p, iFanin3)+6*nVars+nDivs, 0))
			}
			assert(Vec_IntCheckUniqueSmall(vSupp) != 0, "Vec_IntCheckUniqueSmall(vSupp)")

			// sort fanins by level
			//Vec_IntSelectSortCost( Vec_IntArray(vSupp) + NodeMark, Vec_IntSize(vSupp) - NodeMark, &p->vLevelD );
			//Sat_SolverWriteDimacs( pSat, NULL, Vec_IntArray(vSupp), Vec_IntLimit(vSupp), 0 );
			// solve for these fanins
			status = sat_solver_solve(pSat, Vec_IntArray(vSupp)[:Vec_IntSize(vSupp)], nil, 0, 0, 0, 0)
			if status != l_False {
				fmt.Printf("Failed internal check at node %d.\n", Pivot)
			}
			assert(status == l_False, "status == l_False")
			nSuppNew = sat_solver_minimize_assumptions(pSat, Vec_IntArray(vSupp), Vec_IntSize(vSupp), 0)
			Vec_IntShrink(vSupp, nSuppNew)
			Vec_IntLits2Vars(vSupp, -6*nVars)
			Vec_IntSort(vSupp, 1)
			// count how many belong to H; the rest belong to G
			NodeMark = 0
			for k3 = 0; k3 < Vec_IntSize(vSupp); k3++ {
				iFanin3 = Vec_IntEntry(vSupp, k3)
				if iFanin3 >= nDivs {
					Vec_IntWriteEntry(vSupp, k3, iFanin3-nDivs)
				} else {
					NodeMark++
				}
			}
			if NodeMark == 0 {
				//printf( "Obj %d: Special case 2 (vars = %d)\n", Pivot, Vec_IntSize(vSupp) );
				continue
			}
			assert(NodeMark > 0, "NodeMark > 0")
			if Vec_IntSize(vSupp)-NodeMark <= nLutSize {
				return NodeMark
			}
		}
	}

	return 0
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

type Acb_Mfs_t struct {
	pNtk      *Acb_Ntk_t     // network
	pPars     *Acb_Par_t     // parameters
	pSat      [3]*sat_solver // SAT solvers
	vSupp     *Vec_Int_t     // support
	vFlip     *Vec_Int_t     // support
	vValues   *Vec_Int_t     // support
	nNodes    int32          // nodes
	nWins     int32          // windows
	nWinsAll  int32          // windows
	nDivsAll  int32          // windows
	nChanges  [8]int32       // changes
	nOvers    int32          // overflows
	nTwoNodes int32          // two nodes
	timeTotal abctime
	timeCnf   abctime
	timeSol   abctime
	timeWin   abctime
	timeSat   abctime
	timeSatU  abctime
	timeSatS  abctime
}

func Acb_MfsStart(pNtk *Acb_Ntk_t, pPars *Acb_Par_t) *Acb_Mfs_t {
	var p *Acb_Mfs_t = new(Acb_Mfs_t)
	p.pNtk = pNtk
	p.pPars = pPars
	p.timeTotal = Abc_Clock()
	p.pSat[0] = sat_solver_new()
	p.pSat[1] = sat_solver_new()
	p.pSat[2] = sat_solver_new()
	p.vSupp = Vec_IntAlloc(100)
	p.vFlip = Vec_IntAlloc(100)
	p.vValues = Vec_IntAlloc(100)
	return p
}
func Acb_MfsStop(p *Acb_Mfs_t) {
	Vec_IntFree(p.vFlip)
	Vec_IntFree(p.vSupp)
	Vec_IntFree(p.vValues)
	sat_solver_delete(p.pSat[0])
	sat_solver_delete(p.pSat[1])
	sat_solver_delete(p.pSat[2])
	p = nil
}
func Acb_NtkObjMffcEstimate(pNtk *Acb_Ntk_t, iObj int32) int32 {
	var k int32
	var iFanin int32
	var pFanins []int32
	var Count int32
	var iFaninCrit int32 = -1
	for k = 0; k < Acb_ObjFaninNum(pNtk, iObj); k++ {
		pFanins = Acb_ObjFanins(pNtk, iObj)
		iFanin = pFanins[k+1]
		if Acb_ObjIsAreaCritical(pNtk, iFanin) != 0 {
			iFaninCrit = iFanin
			Count++
		}
	}
	if Count != 1 {
		return Count
	}
	for k = 0; k < Acb_ObjFaninNum(pNtk, iFaninCrit); k++ {
		pFanins = Acb_ObjFanins(pNtk, iFaninCrit)
		iFanin = pFanins[k+1]
		if Acb_ObjIsAreaCritical(pNtk, iFanin) != 0 {
			Count++
		}
	}
	return Count
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkOptNodeAnalyze(p *Acb_Mfs_t, PivotVar int32, nDivs int32, nValues int32, pValues []int32, vSupp *Vec_Int_t) {
	var OnSet [64]word
	var OffSet [64]word
	var Diffs [64]word
	var s int32
	var nScope int32 = 1 + 2*nDivs
	var d int32
	var i int32
	var f int32
	var nFrames int32 = nValues / nScope
	var start int32
	if nDivs < 64 {
		start = 0
	} else {
		start = nDivs - 64
	}
	var stop int32
	if nDivs < 64 {
		stop = nDivs
	} else {
		stop = 64
	}
	assert(nValues%nScope == 0, "nValues % nScope == 0")
	assert(nFrames <= 16, "nFrames <= 16")
	for f = 0; f < nFrames; f++ {
		pStart := pValues[int(f*nScope):]
		pOnSet := pStart[1+boolInt(pStart[0] != 0, 0, nDivs):]
		pOffSet := pStart[1+boolInt(pStart[0] != 0, nDivs, 0):]

		fmt.Printf("%2d:", f)
		for s = start; s < stop; s++ {
			fmt.Printf("%d", pOnSet[s])
		}
		fmt.Printf("\n")

		fmt.Printf("%2d:", f)
		for s = start; s < stop; s++ {
			fmt.Printf("%d", pOffSet[s])
		}
		fmt.Printf("\n")

		for s = start; s < stop; s++ {
			if pOnSet[s] != 0 {
				OnSet[f] |= word(1) << uint(s-start)
			}
			if pOffSet[s] != 0 {
				OffSet[f] |= word(1) << uint(s-start)
			}
		}
	}
	d = 0
	for f = 0; f < nFrames; f++ {
		for s = 0; s < nFrames; s++ {
			for i = 0; i < d; i++ {
				if Diffs[i] == (OnSet[f] ^ OffSet[s]) {
					break
				}
			}
			if i < d {
				continue
			}
			if d < 64 {
				Diffs[d] = OnSet[f] ^ OffSet[s]
				d++
			}
		}
	}

	fmt.Printf("Divisors = %d.  Frames = %d.  Patterns = %d.\n", nDivs, nFrames, d)
	fmt.Printf("   ")
	for s = start; s < stop; s++ {
		fmt.Printf("%d", s/10)
	}
	fmt.Printf("\n")
	fmt.Printf("   ")
	for s = start; s < stop; s++ {
		fmt.Printf("%d", s%10)
	}
	fmt.Printf("\n")
	fmt.Printf("   ")
	for s = start; s < stop; s++ {
		idx := Vec_IntFind(vSupp, s)
		ch := rune(' ')
		if idx >= 0 {
			ch = rune('a') + rune(idx)
		}
		fmt.Printf("%c", ch)
	}
	fmt.Printf("\n")
	for s = 0; s < d; s++ {
		fmt.Printf("%2d:", s)
		for f = 0; f < stop; f++ {
			ch := rune(' ')
			if (Diffs[s]>>uint(f))&1 != 0 {
				ch = '*'
			}
			fmt.Printf("%c", ch)
		}
		fmt.Printf("\n")
	}
}

func Acb_NtkOptNode(p *Acb_Mfs_t, Pivot int32) int32 {
	var pCnf *Cnf_Dat_t = nil
	var clk abctime
	var vWin *Vec_Int_t = nil
	var uTruth word
	var Result, PivotVar, nDivs, RetValue, c int32
	assert(Acb_ObjFanoutNum(p.pNtk, Pivot) > 0, "Acb_ObjFanoutNum(p.pNtk, Pivot) > 0")
	p.nWins++

	// compute divisors and window for this target node with these taboo nodes
	clk = Abc_Clock()
	vWin = Acb_NtkWindow(p.pNtk, Pivot, p.pPars.nTfiLevMax, p.pPars.nTfoLevMax, p.pPars.nFanoutMax, boolInt(p.pPars.fArea == 0, 1, 0), &nDivs)
	p.nWinsAll += Vec_IntSize(vWin)
	p.nDivsAll += nDivs
	p.timeWin += Abc_Clock() - clk
	PivotVar = Vec_IntFind(vWin, Abc_Var2Lit(Pivot, 0))
	if p.pPars.fVerbose != 0 {
		fmt.Printf("Node %d: Window contains %d objects and %d divisors.  ", Pivot, Vec_IntSize(vWin), nDivs)
	}
	//    Acb_WinPrint( p->pNtk, vWin, Pivot, nDivs );
	//    Acb_NtkPrintVecWin( p->pNtk, vWin, "Win" );
	if Vec_IntSize(vWin) > p.pPars.nWinNodeMax {
		p.nOvers++
		if p.pPars.fVerbose != 0 {
			fmt.Printf("Too many divisors.\n")
		}
		goto cleanup
	}

	// derive CNF
	clk = Abc_Clock()
	pCnf = Acb_NtkWindow2Cnf(p.pNtk, vWin, Pivot)
	assert(PivotVar == Acb_ObjFunc(p.pNtk, Pivot), "PivotVar == Acb_ObjFunc(p.pNtk, Pivot)")
	Cnf_DataCollectFlipLits(pCnf, PivotVar, p.vFlip)
	p.timeCnf += Abc_Clock() - clk

	// derive SAT solver
	clk = Abc_Clock()
	Acb_NtkWindow2Solver(p.pSat[0], pCnf, p.vFlip, PivotVar, nDivs, 1)
	p.timeSol += Abc_Clock() - clk
	// check constants
	for c = 0; c < 2; c++ {
		var Lit int32 = Abc_Var2Lit(PivotVar, c)
		var status int32 = sat_solver_solve(p.pSat[0], []int32{Lit}, nil, 0, 0, 0, 0)
		if status == l_False {
			p.nChanges[0]++
			if p.pPars.fVerbose != 0 {
				fmt.Printf("Found constant %d.\n", c)
			}
			Acb_NtkUpdateNode(p.pNtk, Pivot, word(boolInt(c != 0, -1, 0)), nil)
			RetValue = 1
			goto cleanup
		}
		assert(status == l_True, "status == l_True")
	}

	// derive SAT solver
	clk = Abc_Clock()
	Acb_NtkWindow2Solver(p.pSat[1], pCnf, p.vFlip, PivotVar, nDivs, 2)
	p.timeSol += Abc_Clock() - clk

	// try to remove useless fanins
	if p.pPars.fArea != 0 {
		var fEnableProfile int32
		if fEnableProfile != 0 {
			userVars := make([]int32, 0, 1+2*nDivs)
			userVars = append(userVars, PivotVar)
			for c = 0; c < nDivs; c++ {
				userVars = append(userVars, c)
			}
			for c = 0; c < nDivs; c++ {
				userVars = append(userVars, c+Cnf_DataVarNum(pCnf))
			}
			p.pSat[1].SetUserVariables(userVars)
		}

		// perform solving
		clk = Abc_Clock()
		Result = Acb_NtkFindSupp1(p.pNtk, Pivot, p.pSat[1], Cnf_DataVarNum(pCnf), nDivs, vWin, p.vSupp)
		p.timeSat += Abc_Clock() - clk
		// undo variables
		p.pSat[1].ClearUserVariables()
		if Result != 0 {
			if Vec_IntSize(p.vSupp) == 0 {
				p.nChanges[0]++
			} else {
				p.nChanges[1]++
			}
			assert(Vec_IntSize(p.vSupp) < p.pPars.nLutSize, "Vec_IntSize(p.vSupp) < p.pPars.nLutSize")
			if p.pPars.fVerbose != 0 {
				fmt.Printf("Found %d inputs: ", Vec_IntSize(p.vSupp))
			}
			uTruth = Acb_ComputeFunction(p.pSat[0], PivotVar, sat_solver_nvars(p.pSat[0])-1, p.vSupp, 0)
			if p.pPars.fVerbose != 0 {
				extra.Extra_PrintHex(os.Stdout, []uint32{uint32(uTruth), uint32(uTruth >> 32)}, Vec_IntSize(p.vSupp))
			}
			if p.pPars.fVerbose != 0 {
				fmt.Printf("\n")
			}
			// create support in terms of nodes
			Vec_IntRemap(p.vSupp, vWin)
			Vec_IntLits2Vars(p.vSupp, 0)
			Acb_NtkUpdateNode(p.pNtk, Pivot, uTruth, p.vSupp)
			RetValue = 1
			goto cleanup
		}
		if fEnableProfile != 0 {
			// analyze the resulting values
			Acb_NtkOptNodeAnalyze(p, PivotVar, nDivs, p.pSat[1].UserValueCount(), p.pSat[1].UserValues(), p.vSupp)
			p.pSat[1].ClearUserValues()
		}
	}

	if Acb_NtkObjMffcEstimate(p.pNtk, Pivot) >= 1 {
		// check for one-node implementation
		clk = Abc_Clock()
		Result = Acb_NtkFindSupp2(p.pNtk, Pivot, p.pSat[1], Cnf_DataVarNum(pCnf), nDivs, vWin, p.vSupp, p.pPars.nLutSize, boolInt(p.pPars.fArea == 0, 1, 0))
		p.timeSat += Abc_Clock() - clk
		if Result != 0 {
			p.nChanges[2]++
			assert(Vec_IntSize(p.vSupp) <= p.pPars.nLutSize, "Vec_IntSize(p.vSupp) <= p.pPars.nLutSize")
			if p.pPars.fVerbose != 0 {
				fmt.Printf("Found %d inputs: ", Vec_IntSize(p.vSupp))
			}
			uTruth = Acb_ComputeFunction(p.pSat[0], PivotVar, sat_solver_nvars(p.pSat[0])-1, p.vSupp, 0)
			if p.pPars.fVerbose != 0 {
				extra.Extra_PrintHex(os.Stdout, []uint32{uint32(uTruth), uint32(uTruth >> 32)}, Vec_IntSize(p.vSupp))
			}
			if p.pPars.fVerbose != 0 {
				fmt.Printf("\n")
			}
			// create support in terms of nodes
			Vec_IntRemap(p.vSupp, vWin)
			Vec_IntLits2Vars(p.vSupp, 0)
			Acb_NtkUpdateNode(p.pNtk, Pivot, uTruth, p.vSupp)
			RetValue = 1
			goto cleanup
		}
	}

	//#if 0
	if p.pPars.fUseAshen != 0 && Acb_NtkObjMffcEstimate(p.pNtk, Pivot) >= 2 { // && Pivot != 70 )

		p.nTwoNodes++
		// derive SAT solver
		clk = Abc_Clock()
		Acb_NtkWindow2Solver(p.pSat[2], pCnf, p.vFlip, PivotVar, nDivs, 6)
		p.timeSol += Abc_Clock() - clk

		// check for two-node implementation
		clk = Abc_Clock()
		Result = Acb_NtkFindSupp3(p.pNtk, Pivot, p.pSat[2], Cnf_DataVarNum(pCnf), nDivs, vWin, p.vSupp, p.pPars.nLutSize, boolInt(p.pPars.fArea == 0, 1, 0))
		p.timeSat += Abc_Clock() - clk
		if Result != 0 {
			var fVerbose int32 = 1
			var i, k, Lit, Var, Var2, status, NodeNew, fBecameUnsat, fCompl int32
			assert(Result < p.pPars.nLutSize, "Result                       <  p.pPars.nLutSize")
			assert(Vec_IntSize(p.vSupp)-Result <= p.pPars.nLutSize, "Vec_IntSize(p.vSupp)-Result <= p.pPars.nLutSize")
			if fVerbose != 0 || p.pPars.fVerbose != 0 {
				fmt.Printf("Obj %5d: Found %d Hvars and %d Gvars: ", Pivot, Result, Vec_IntSize(p.vSupp)-Result)
			}
			// p->vSupp contains G variables (Vec_IntSize(p->vSupp)-Result) followed by H variables (Result)
			//sat_solver_restart( p->pSat[1] );
			//Acb_NtkWindow2Solver( p->pSat[1], pCnf, p->vFlip, PivotVar, nDivs, 2 );

			// constrain H-variables to be equal
			for i = Vec_IntSize(p.vSupp) - Result; i < Vec_IntSize(p.vSupp); i++ {
				Var = Vec_IntEntry(p.vSupp, i)
				assert(Var >= 0 && Var < nDivs, "Var >= 0 && Var < nDivs")
				assert(Var+2*Cnf_DataVarNum(pCnf) < sat_solver_nvars(p.pSat[1]), "Var + 2*Cnf_DataVarNum(pCnf) < sat_solver_nvars(p.pSat[1])")
				Lit = Abc_Var2Lit(Var+2*Cnf_DataVarNum(pCnf), 0) // HVars are the same
				if !sat_solver_addclause(p.pSat[1], []int32{Lit}) {
					if fVerbose != 0 || p.pPars.fVerbose != 0 {
						fmt.Printf("Error: SAT solver became UNSAT at a wrong place (place 2).  ")
					}
					fBecameUnsat = 1
				}
			}
			// find one satisfying assighment
			status = sat_solver_solve(p.pSat[1], nil, nil, 0, 0, 0, 0)
			assert(status == l_True, "status == l_True")
			// get assignment of the function
			fCompl = boolInt(sat_solver_var_value(p.pSat[1], PivotVar) == 0, 1, 0)
			// The C solver retains its model while clauses are added. The Go
			// API invalidates model access on mutation, so save this assignment.
			model := p.pSat[1].Model()
			// constrain second set of G-vars to have values equal to the assignment
			for i = 0; i < Vec_IntSize(p.vSupp)-Result && i < Vec_IntSize(p.vSupp); i++ {
				Var = Vec_IntEntry(p.vSupp, i)
				// check if this is a C-var
				for k = Vec_IntSize(p.vSupp) - Result; k < Vec_IntSize(p.vSupp); k++ {
					Var2 = Vec_IntEntry(p.vSupp, k)
					if Var == Var2 {
						break
					}
				}
				if k < Vec_IntSize(p.vSupp) { // do not constrain a C-var

					if fVerbose != 0 || p.pPars.fVerbose != 0 {
						fmt.Printf("Found C-var in object %d.  ", Pivot)
					}
					continue
				}
				assert(Var >= 0 && Var < nDivs, "Var >= 0 && Var < nDivs")
				Lit = Abc_Var2Lit(Var+Cnf_DataVarNum(pCnf), boolInt(!model[Var+Cnf_DataVarNum(pCnf)], 1, 0))
				if !sat_solver_addclause(p.pSat[1], []int32{Lit}) {
					if fVerbose != 0 || p.pPars.fVerbose != 0 {
						fmt.Printf("Error: SAT solver became UNSAT at a wrong place (place 1).  ")
					}
					fBecameUnsat = 1
				}
			}
			if fBecameUnsat != 0 {
				StrCount++
				if fVerbose != 0 || p.pPars.fVerbose != 0 {
					fmt.Printf(" Quitting.\n")
				}
				goto cleanup
			}
			// consider only G variables
			oldSupport := Vec_IntSize(p.vSupp)
			preservedTail := append([]int32(nil), Vec_IntArray(p.vSupp)[oldSupport-Result:]...)
			Vec_IntShrink(p.vSupp, oldSupport-Result)
			// truth table
			uTruth = Acb_ComputeFunction(p.pSat[1], PivotVar, sat_solver_nvars(p.pSat[1])-1, p.vSupp, fCompl)
			if fVerbose != 0 || p.pPars.fVerbose != 0 {
				extra.Extra_PrintHex(os.Stdout, []uint32{uint32(uTruth), uint32(uTruth >> 32)}, Vec_IntSize(p.vSupp))
			}
			if uTruth == 0 || ^uTruth == 0 {
				if fVerbose != 0 || p.pPars.fVerbose != 0 {
					fmt.Printf(" Quitting.\n")
				}
				goto cleanup
			}
			p.nChanges[3]++
			// create new node
			Vec_IntRemap(p.vSupp, vWin)
			Vec_IntLits2Vars(p.vSupp, 0)
			NodeNew = Acb_NtkCreateNode(p.pNtk, uTruth, p.vSupp)
			Acb_DeriveCnfForWindowOne(p.pNtk, NodeNew)
			Acb_DeriveCnfForNode(p.pNtk, NodeNew, p.pSat[0], sat_solver_nvars(p.pSat[0])-2)
			Vec_IntPushArray(p.vSupp, preservedTail, int32(len(preservedTail)))
			// collect new variables
			for i = Vec_IntSize(p.vSupp) - Result; i < Vec_IntSize(p.vSupp); i++ {
				Var = Vec_IntEntry(p.vSupp, i)
				Vec_IntWriteEntry(p.vSupp, i-(Vec_IntSize(p.vSupp)-Result), Var)
			}
			Vec_IntShrink(p.vSupp, Result)
			Vec_IntPush(p.vSupp, sat_solver_nvars(p.pSat[0])-2)
			// truth table
			uTruth = Acb_ComputeFunction(p.pSat[0], PivotVar, sat_solver_nvars(p.pSat[0])-1, p.vSupp, 0)
			// create new fanins of the node
			if fVerbose != 0 || p.pPars.fVerbose != 0 {
				fmt.Printf("  ")
			}
			if fVerbose != 0 || p.pPars.fVerbose != 0 {
				extra.Extra_PrintHex(os.Stdout, []uint32{uint32(uTruth), uint32(uTruth >> 32)}, Vec_IntSize(p.vSupp))
			}
			if fVerbose != 0 || p.pPars.fVerbose != 0 {
				fmt.Printf("\n")
			}
			// create support in terms of nodes
			Vec_IntPop(p.vSupp)
			Vec_IntRemap(p.vSupp, vWin)
			Vec_IntLits2Vars(p.vSupp, 0)
			Vec_IntPush(p.vSupp, NodeNew)
			Acb_NtkUpdateNode(p.pNtk, Pivot, uTruth, p.vSupp)
			RetValue = 2
			goto cleanup
		}
	}
	//#endif

	if p.pPars.fVerbose != 0 {
		fmt.Printf("\n")
	}

cleanup:
	sat_solver_restart(p.pSat[0])
	sat_solver_restart(p.pSat[1])
	sat_solver_restart(p.pSat[2])
	if pCnf != nil {
		Cnf_DataFree(pCnf)
		Acb_NtkWindowUndo(p.pNtk, vWin)
	}
	Vec_IntFreeP(&vWin)
	return RetValue
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Acb_NtkOpt(pNtk *Acb_Ntk_t, pPars *Acb_Par_t) {
	var pMan *Acb_Mfs_t = Acb_MfsStart(pNtk, pPars)
	if pPars.fVerbose != 0 {
		mode := "Delay"
		if pMan.pPars.fArea != 0 {
			mode = "Area"
		}
		fmt.Printf("%s-optimization parameters: TfiLev(I) = %d  TfoLev(O) = %d  WinMax(W) = %d  LutSize = %d\n", mode, pMan.pPars.nTfiLevMax, pMan.pPars.nTfoLevMax, pMan.pPars.nWinNodeMax, pMan.pPars.nLutSize)
	}
	Acb_NtkCreateFanout(pNtk)  // fanout data structure
	Acb_NtkCleanObjFuncs(pNtk) // SAT variables
	Acb_NtkCleanObjCnfs(pNtk)  // CNF representations
	if pMan.pPars.fArea != 0 {
		var n int32
		var iObj int32
		var nNodes int32 = Acb_NtkObjNumMax(pNtk)
		var vVisited *Vec_Bit_t = Vec_BitStart(Acb_NtkObjNumMax(pNtk))
		Acb_NtkUpdateLevelD(pNtk, -1) // compute forward logic level
		for n = 2; n >= 0; n-- {
			for iObj = 1; iObj < Vec_StrSize(&pNtk.vObjType); iObj++ {
				if Acb_ObjType(pNtk, iObj) == 0 || Acb_ObjIsCio(pNtk, iObj) != 0 || iObj >= nNodes || Vec_BitEntry(vVisited, iObj) != 0 || Acb_NtkObjMffcEstimate(pNtk, iObj) < n {
					continue
				}
				pMan.nNodes++
				for Acb_NtkOptNode(pMan, iObj) != 0 && Acb_ObjFaninNum(pNtk, iObj) > 0 {
				}
				Vec_BitWriteEntry(vVisited, iObj, 1)
			}
		}
		Vec_BitFree(vVisited)

	} else {
		var Value int32
		Acb_NtkUpdateTiming(pNtk, -1) // compute delay information
		for pNtk.vQue != nil {
			Value = int32(Vec_QueTopPriority(pNtk.vQue))
			if Value <= 0 {
				break
			}
			var iObj int32 = Vec_QuePop(pNtk.vQue)
			if Acb_ObjType(pNtk, iObj) == 0 {
				continue
			}
			//if ( iObj != 103 )
			//    continue;
			//printf( "Trying node %4d (%4d) ", iObj, Value );
			Acb_NtkOptNode(pMan, iObj)
		}
	}
	if pPars.fVerbose != 0 {
		pMan.timeTotal = Abc_Clock() - pMan.timeTotal
		fmt.Printf("Node = %d  Win = %d (Ave = %d)  DivAve = %d   Change = %d  C = %d  N1 = %d  N2 = %d  N3 = %d   Over = %d  Str = %d  2Node = %d.\n", pMan.nNodes, pMan.nWins, pMan.nWinsAll/Abc_MaxInt(1, pMan.nWins), pMan.nDivsAll/Abc_MaxInt(1, pMan.nWins), pMan.nChanges[0]+pMan.nChanges[1]+pMan.nChanges[2]+pMan.nChanges[3], pMan.nChanges[0], pMan.nChanges[1], pMan.nChanges[2], pMan.nChanges[3], pMan.nOvers, StrCount, pMan.nTwoNodes)
		ABC_PRTP("Windowing  ", pMan.timeWin, pMan.timeTotal)
		ABC_PRTP("CNF compute", pMan.timeCnf, pMan.timeTotal)
		ABC_PRTP("Make solver", pMan.timeSol, pMan.timeTotal)
		ABC_PRTP("SAT solving", pMan.timeSat, pMan.timeTotal)
		//        ABC_PRTP( "  unsat    ", pMan->timeSatU,   pMan->timeTotal );
		//        ABC_PRTP( "  sat      ", pMan->timeSatS,   pMan->timeTotal );
		ABC_PRTP("TOTAL      ", pMan.timeTotal, pMan.timeTotal)
		// fmt.Printf writes directly to os.Stdout, so no explicit flush is needed.
	}
	Acb_MfsStop(pMan)
	StrCount = 0
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
