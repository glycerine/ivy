package sbd

// Port of src/opt/sbd/sbdWin.c: SAT-based window processing.

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// s_Truths6 from misc/util/utilTruth.h.
var s_Truths6 = [6]word{
	0xAAAAAAAAAAAAAAAA,
	0xCCCCCCCCCCCCCCCC,
	0xF0F0F0F0F0F0F0F0,
	0xFF00FF00FF00FF00,
	0xFFFF0000FFFF0000,
	0xFFFFFFFF00000000,
}

// Sbd_ManSatSolver constructs SAT solver for the window.
//
// The window for the pivot node (Pivot) is represented as a DFS ordered
// array of objects (vWinObjs) whose indexed in the array (which will be used
// as SAT variables) are given in array vObj2Var. The TFO nodes are listed as
// the last ones in vWinObjs. The root nodes are labeled with Abc_LitIsCompl()
// in vTfo and also given in vRoots. If fQbf is 1, returns the instance meant
// for QBF solving. It is using the last variable (LastVar) as the placeholder
// for the second copy of the pivot node.
func Sbd_ManSatSolver(pSat *sat_solver, p *Gia_Man_t, vMirrors *Vec_Int_t,
	Pivot int32, vWinObjs, vObj2Var, vTfo, vRoots *Vec_Int_t, fQbf int32) *sat_solver {
	iLit := int32(1)
	TfoStart := int32(vec.Vec_IntSize(vWinObjs) - vec.Vec_IntSize(vTfo))
	PivotVar := vec.Vec_IntEntry(vObj2Var, int(Pivot))
	LastVar := int32(vec.Vec_IntSize(vWinObjs) + vec.Vec_IntSize(vTfo) + vec.Vec_IntSize(vRoots))
	nTfo := int32(vec.Vec_IntSize(vTfo))
	// create SAT solver
	if pSat == nil {
		pSat = sat_solver_new()
	} else {
		sat_solver_restart(pSat)
	}
	sat_solver_setnvars(pSat, int32(vec.Vec_IntSize(vWinObjs)+vec.Vec_IntSize(vTfo)+vec.Vec_IntSize(vRoots))+SBD_FVAR_MAX)
	// create constant 0 clause
	sat_solver_addclause(pSat, []int32{iLit})
	// add clauses for all nodes
	for i := 1; i < vec.Vec_IntSize(vWinObjs); i++ {
		iObj := vec.Vec_IntEntry(vWinObjs, i)
		pObj := Gia_ManObj(p, iObj)
		if Gia_ObjIsCi(pObj) {
			continue
		}
		assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
		assert(vec.Vec_IntEntry(vMirrors, int(iObj)) < 0, "Vec_IntEntry( vMirrors, iObj ) < 0")
		Node := vec.Vec_IntEntry(vObj2Var, int(iObj))
		Lit0m := vec.Vec_IntEntry(vMirrors, int(Gia_ObjFaninId0(pObj, iObj)))
		Lit1m := vec.Vec_IntEntry(vMirrors, int(Gia_ObjFaninId1(pObj, iObj)))
		Fan0 := Gia_ObjFaninId0(pObj, iObj)
		if Lit0m >= 0 {
			Fan0 = Abc_Lit2Var(Lit0m)
		}
		Fan1 := Gia_ObjFaninId1(pObj, iObj)
		if Lit1m >= 0 {
			Fan1 = Abc_Lit2Var(Lit1m)
		}
		Fan0 = vec.Vec_IntEntry(vObj2Var, int(Fan0))
		Fan1 = vec.Vec_IntEntry(vObj2Var, int(Fan1))
		fCompl0 := Gia_ObjFaninC0(pObj) ^ b2i(Lit0m >= 0 && Abc_LitIsCompl(Lit0m) != 0)
		fCompl1 := Gia_ObjFaninC1(pObj) ^ b2i(Lit1m >= 0 && Abc_LitIsCompl(Lit1m) != 0)
		if Gia_ObjIsXor(pObj) {
			sat_solver_add_xor(pSat, Node, Fan0, Fan1, fCompl0^fCompl1)
		} else {
			sat_solver_add_and(pSat, Node, Fan0, Fan1, fCompl0, fCompl1, 0)
		}
	}
	// add second clauses for the TFO
	for i := int(TfoStart); i < vec.Vec_IntSize(vWinObjs); i++ {
		iObj := vec.Vec_IntEntry(vWinObjs, i)
		pObj := Gia_ManObj(p, iObj)
		assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
		assert(vec.Vec_IntEntry(vMirrors, int(iObj)) < 0, "Vec_IntEntry( vMirrors, iObj ) < 0")
		Node := vec.Vec_IntEntry(vObj2Var, int(iObj)) + nTfo
		Lit0m := vec.Vec_IntEntry(vMirrors, int(Gia_ObjFaninId0(pObj, iObj)))
		Lit1m := vec.Vec_IntEntry(vMirrors, int(Gia_ObjFaninId1(pObj, iObj)))
		Fan0 := Gia_ObjFaninId0(pObj, iObj)
		if Lit0m >= 0 {
			Fan0 = Abc_Lit2Var(Lit0m)
		}
		Fan1 := Gia_ObjFaninId1(pObj, iObj)
		if Lit1m >= 0 {
			Fan1 = Abc_Lit2Var(Lit1m)
		}
		Fan0 = vec.Vec_IntEntry(vObj2Var, int(Fan0))
		Fan1 = vec.Vec_IntEntry(vObj2Var, int(Fan1))
		if Fan0 >= TfoStart {
			Fan0 += nTfo
		}
		if Fan1 >= TfoStart {
			Fan1 += nTfo
		}
		if fQbf != 0 {
			if Fan0 == PivotVar {
				Fan0 = LastVar
			}
			if Fan1 == PivotVar {
				Fan1 = LastVar
			}
		}
		fCompl0 := Gia_ObjFaninC0(pObj) ^ b2i(fQbf == 0 && Fan0 == PivotVar) ^ b2i(Lit0m >= 0 && Abc_LitIsCompl(Lit0m) != 0)
		fCompl1 := Gia_ObjFaninC1(pObj) ^ b2i(fQbf == 0 && Fan1 == PivotVar) ^ b2i(Lit1m >= 0 && Abc_LitIsCompl(Lit1m) != 0)
		if Gia_ObjIsXor(pObj) {
			sat_solver_add_xor(pSat, Node, Fan0, Fan1, fCompl0^fCompl1)
		} else {
			sat_solver_add_and(pSat, Node, Fan0, Fan1, fCompl0, fCompl1, 0)
		}
	}
	if vec.Vec_IntSize(vRoots) > 0 {
		// create XOR clauses for the roots
		nVars := int32(vec.Vec_IntSize(vWinObjs) + vec.Vec_IntSize(vTfo))
		vFaninVars := make([]int32, 0, vec.Vec_IntSize(vRoots))
		for _, iObj := range vec.Vec_IntArray(vRoots) {
			assert(vec.Vec_IntEntry(vMirrors, int(iObj)) < 0, "Vec_IntEntry( vMirrors, iObj ) < 0")
			Node := vec.Vec_IntEntry(vObj2Var, int(iObj))
			vFaninVars = append(vFaninVars, Abc_Var2Lit(nVars, 0))
			sat_solver_add_xor(pSat, Node, Node+nTfo, nVars, 0)
			nVars++
		}
		// make OR clause for the last nRoots variables
		RetValue := sat_solver_addclause(pSat, vFaninVars)
		if RetValue == 0 {
			sat_solver_delete(pSat)
			return nil
		}
		assert(sat_solver_nvars(pSat) == nVars+SBD_FVAR_MAX, "sat_solver_nvars(pSat) == nVars + SBD_FVAR_MAX")
	} else if fQbf != 0 {
		for n := int32(0); n < 2; n++ {
			pLits := []int32{Abc_Var2Lit(PivotVar, n), Abc_Var2Lit(LastVar, n)}
			RetValue := sat_solver_addclause(pSat, pLits)
			assert(RetValue != 0, "RetValue")
		}
	}
	// finalize
	RetValue := sat_solver_simplify(pSat)
	if RetValue == 0 {
		sat_solver_delete(pSat)
		return nil
	}
	return pSat
}

// Sbd_ManSolve solves one SAT problem.
//
// Computes node function for PivotVar with fanins in vDivSet using
// don't-care represented in the SAT solver. Uses array vValues to return the
// values of the first Vec_IntSize(vValues) SAT variables in case the
// implementation of the node with the given fanins does not exist.
func Sbd_ManSolve(pSat *sat_solver, PivotVar, FreeVar int32, vDivSet, vDivVars, vDivValues, vTemp *Vec_Int_t) word {
	nBTLimit := int64(0)
	var uTruth word
	var status int32
	assert(FreeVar < sat_solver_nvars(pSat), "FreeVar < sat_solver_nvars(pSat)")
	assert(vec.Vec_IntSize(vDivVars) == vec.Vec_IntSize(vDivValues), "Vec_IntSize(vDivVars) == Vec_IntSize(vDivValues)")
	pLits := []int32{Abc_Var2Lit(PivotVar, 0), Abc_Var2Lit(FreeVar, 0)} // F = 1; iNewLit
	for {
		// find onset minterm
		status = sat_solver_solve(pSat, pLits, nBTLimit, 0, 0, 0)
		if status == l_Undef {
			return SBD_SAT_UNDEC
		}
		if status == l_False {
			return uTruth
		}
		assert(status == l_True, "status == l_True")
		// remember variable values
		for i, iVar := range vec.Vec_IntArray(vDivVars) {
			vec.Vec_IntWriteEntry(vDivValues, i, 2*sat_solver_var_value(pSat, iVar))
		}
		// collect divisor literals
		vec.Vec_IntClear(vTemp)
		vec.Vec_IntPush(vTemp, Abc_LitNot(pLits[0])) // F = 0
		for _, iVar := range vec.Vec_IntArray(vDivSet) {
			vec.Vec_IntPush(vTemp, sat_solver_var_literal(pSat, iVar))
		}
		// check against offset
		status = sat_solver_solve(pSat, vec.Vec_IntArray(vTemp), nBTLimit, 0, 0, 0)
		if status == l_Undef {
			return SBD_SAT_UNDEC
		}
		if status == l_True {
			break
		}
		assert(status == l_False, "status == l_False")
		// compute cube and add clause
		pFinal := sat_solver_final(pSat)
		uCube := ^word(0)
		vec.Vec_IntClear(vTemp)
		vec.Vec_IntPush(vTemp, Abc_LitNot(pLits[1])) // NOT(iNewLit)
		for _, lit := range pFinal {
			if lit == pLits[0] {
				continue
			}
			vec.Vec_IntPush(vTemp, lit)
			iVar := int32(vec.Vec_IntFind(vDivSet, Abc_Lit2Var(lit)))
			assert(iVar >= 0, "iVar >= 0")
			// The fixed C source skips divisors beyond the six elementary
			// truth tables: with more divisors the truth table is not
			// meaningful and callers only test it against SBD_SAT_*.
			if iVar < 6 {
				if Abc_LitIsCompl(lit) != 0 {
					uCube &= s_Truths6[iVar]
				} else {
					uCube &= ^s_Truths6[iVar]
				}
			}
		}
		uTruth |= uCube
		status = sat_solver_addclause(pSat, vec.Vec_IntArray(vTemp))
		assert(status != 0, "status")
	}
	assert(status == l_True, "status == l_True")
	// store the counter-example
	for i, iVar := range vec.Vec_IntArray(vDivVars) {
		vec.Vec_IntAddToEntry(vDivValues, i, sat_solver_var_value(pSat, iVar))
	}
	for i := 0; i < vec.Vec_IntSize(vDivValues); i++ {
		vec.Vec_IntAddToEntry(vDivValues, i, 0xC)
	}
	return SBD_SAT_SAT
}

func Sbd_ManSolve2(pSat *sat_solver, PivotVar, FreeVar int32, vDivVars, vDivValues, vTemp, vSop *Vec_Int_t) int32 {
	nBTLimit := int64(0)
	var status int32
	assert(FreeVar < sat_solver_nvars(pSat), "FreeVar < sat_solver_nvars(pSat)")
	assert(vec.Vec_IntSize(vDivVars) == vec.Vec_IntSize(vDivValues), "Vec_IntSize(vDivVars) == Vec_IntSize(vDivValues)")
	pLits := []int32{Abc_Var2Lit(PivotVar, 0), Abc_Var2Lit(FreeVar, 0)}
	vec.Vec_IntClear(vSop)
	for {
		// find onset minterm
		status = sat_solver_solve(pSat, pLits, nBTLimit, 0, 0, 0)
		if status == l_Undef {
			return 0
		}
		if status == l_False {
			return 1
		}
		assert(status == l_True, "status == l_True")
		// collect divisor literals
		vec.Vec_IntClear(vTemp)
		vec.Vec_IntPush(vTemp, Abc_LitNot(pLits[0])) // F = 0
		for _, iVar := range vec.Vec_IntArray(vDivVars) {
			vec.Vec_IntPush(vTemp, sat_solver_var_literal(pSat, iVar))
		}
		// check against offset
		status = sat_solver_solve(pSat, vec.Vec_IntArray(vTemp), nBTLimit, 0, 0, 0)
		if status == l_Undef {
			return 0
		}
		if status == l_True {
			break
		}
		assert(status == l_False, "status == l_False")
		// compute cube and add clause
		pFinal := sat_solver_final(pSat)
		vec.Vec_IntClear(vTemp)
		vec.Vec_IntPush(vTemp, Abc_LitNot(pLits[1])) // NOT(iNewLit)
		for _, lit := range pFinal {
			if lit == pLits[0] {
				continue
			}
			vec.Vec_IntPush(vTemp, lit)
			iVar := int32(vec.Vec_IntFind(vDivVars, Abc_Lit2Var(lit)))
			assert(iVar >= 0, "iVar >= 0")
			vec.Vec_IntPush(vSop, Abc_Var2Lit(iVar, 1-Abc_LitIsCompl(lit)))
		}
		vec.Vec_IntPush(vSop, -1)
		status = sat_solver_addclause(pSat, vec.Vec_IntArray(vTemp))
		assert(status != 0, "status")
	}
	assert(status == l_True, "status == l_True")
	return 0
}
func Sbd_ManSolverSupp(vSop *Vec_Int_t, pInds []int32, pnVars *int32) word {
	var Supp word
	nVars := int32(0)
	for _, Entry := range vec.Vec_IntArray(vSop) {
		if Entry == -1 {
			continue
		}
		assert(Abc_Lit2Var(Entry) < 64, "Abc_Lit2Var(Entry) < 64")
		if (Supp>>uint(Abc_Lit2Var(Entry)))&1 != 0 {
			continue
		}
		pInds[Abc_Lit2Var(Entry)] = nVars
		nVars++
		Supp |= word(1) << uint(Abc_Lit2Var(Entry))
	}
	*pnVars = nVars
	return Supp
}
func Sbd_ManSolverPrint(vSop *Vec_Int_t) {
	var nVars int32
	var pInds [64]int32
	Sbd_ManSolverSupp(vSop, pInds[:], &nVars)
	var Cube [65]byte
	for v := int32(0); v < nVars; v++ {
		Cube[v] = '-'
	}
	for _, Entry := range vec.Vec_IntArray(vSop) {
		if Entry == -1 {
			fmt.Printf("%s\n", string(Cube[:nVars]))
			for v := int32(0); v < nVars; v++ {
				Cube[v] = '-'
			}
			continue
		}
		Cube[pInds[Abc_Lit2Var(Entry)]] = '1' - byte(Abc_LitIsCompl(Entry))
	}
}
func Sbd_ManSolveSelect(p *Gia_Man_t, vMirrors *Vec_Int_t, Pivot int32, vDivVars, vDivValues, vWinObjs, vObj2Var, vTfo, vRoots *Vec_Int_t) {
	vSop := vec.Vec_IntAlloc(100)
	vTemp := vec.Vec_IntAlloc(100)
	pSat := Sbd_ManSatSolver(nil, p, vMirrors, Pivot, vWinObjs, vObj2Var, vTfo, vRoots, 0)
	PivotVar := vec.Vec_IntEntry(vObj2Var, int(Pivot))
	FreeVar := int32(vec.Vec_IntSize(vWinObjs) + vec.Vec_IntSize(vTfo) + vec.Vec_IntSize(vRoots))
	Status := Sbd_ManSolve2(pSat, PivotVar, FreeVar, vDivVars, vDivValues, vTemp, vSop)
	fmt.Printf("Pivot = %4d. Divs = %4d.  ", Pivot, vec.Vec_IntSize(vDivVars))
	if Status == 0 {
		fmt.Printf("UNSAT.\n")
	} else {
		var nVars int32
		var pInds [64]int32
		Sbd_ManSolverSupp(vSop, pInds[:], &nVars)
		fmt.Printf("SAT with %d vars and %d cubes.\n", nVars, vec.Vec_IntCountEntry(vSop, -1))
	}
	sat_solver_delete(pSat)
}

// Sbd_ManCollectConstants returns a bunch of positive/negative random care
// minterms. Returns 0/1 if the functions is const 0/1.
func Sbd_ManCollectConstants(pSat *sat_solver, nCareMints [2]int32, PivotVar int32, pVarSims [][]word, vInds *Vec_Int_t) int32 {
	nBTLimit := int64(0)
	assert(int32(vec.Vec_IntSize(vInds)) == nCareMints[0]+nCareMints[1], "Vec_IntSize(vInds) == nCareMints[0] + nCareMints[1]")
	for i, Ind := range vec.Vec_IntArray(vInds) {
		fOffSet := b2i(int32(i) < nCareMints[0])
		iLit := Abc_Var2Lit(PivotVar, fOffSet)
		sat_solver_random_polarity(pSat)
		status := sat_solver_solve(pSat, []int32{iLit}, nBTLimit, 0, 0, 0)
		if status == l_Undef {
			return -2
		}
		if status == l_False {
			return fOffSet
		}
		for k := int32(0); k <= PivotVar; k++ {
			if util.Abc_TtGetBit(pVarSims[k], Ind) != sat_solver_var_value(pSat, k) {
				util.Abc_TtXorBit(pVarSims[k], Ind)
			}
		}
	}
	return -1
}
func Sbd_ManCollectConstantsNew(pSat *sat_solver, vDivVars *Vec_Int_t, nConsts, PivotVar int32, pOnset, pOffset []word) int32 {
	nBTLimit := int64(0)
	pPats := [2][]word{pOnset, pOffset}
	assert(vec.Vec_IntSize(vDivVars) < 64, "Vec_IntSize(vDivVars) < 64")
	for n := int32(0); n < 2; n++ {
		for i := int32(0); i < nConsts; i++ {
			sat_solver_random_polarity(pSat)
			iLit := Abc_Var2Lit(PivotVar, n)
			status := sat_solver_solve(pSat, []int32{iLit}, nBTLimit, 0, 0, 0)
			if status == l_Undef {
				return -2
			}
			if status == l_False {
				return n
			}
			pPats[n][i] = word(b2i(n == 0)) << uint(vec.Vec_IntSize(vDivVars))
			for k, iVar := range vec.Vec_IntArray(vDivVars) {
				if sat_solver_var_value(pSat, iVar) != 0 {
					util.Abc_TtXorBit(pPats[n][i:], int32(k))
				}
			}
		}
	}
	return -1
}
