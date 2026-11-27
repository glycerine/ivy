package sbd

// Port of src/opt/sbd/sbdLut.c: QBF-based synthesis of LUT structures.

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// count the number of parameter variables in the structure
func Sbd_ProblemCountParams(nStrs int32, pStr0 []Sbd_Str_t) int32 {
	nPars := int32(0)
	for i := int32(0); i < nStrs; i++ {
		pStr := &pStr0[i]
		if pStr.fLut != 0 {
			nPars += 1 << uint(pStr.nVarIns)
		} else {
			nPars += pStr.nVarIns
		}
	}
	return nPars
}

// add clauses for the structure
func Sbd_ProblemAddClauses(pSat *sat_solver, nVars, nStrs int32, pVars []int32, pStr0 []Sbd_Str_t) int32 {
	// variable order:  inputs, structure outputs, parameters
	VarOut := nVars
	VarPar := nVars + nStrs
	// C sized this SBD_SIZE_MAX+2, too small for LUTs of 5 and 6 inputs
	// (see CBUGS_FIXED.md); the fixed source uses SBD_DIV_MAX+2.
	var pLits [SBD_DIV_MAX + 2]int32
	for s := int32(0); s < nStrs; s, VarOut = s+1, VarOut+1 {
		pStr := &pStr0[s]
		if pStr.fLut != 0 {
			nMints := int32(1) << uint(pStr.nVarIns)
			assert(pStr.nVarIns <= 6, "pStr->nVarIns <= 6")
			for m := int32(0); m < nMints; m, VarPar = m+1, VarPar+1 {
				for k := int32(0); k < pStr.nVarIns; k++ {
					pLits[k] = Abc_Var2Lit(pVars[pStr.VarIns[k]], (m>>uint(k))&1)
				}
				for n := int32(0); n < 2; n++ {
					pLits[pStr.nVarIns] = Abc_Var2Lit(pVars[VarPar], n)
					pLits[pStr.nVarIns+1] = Abc_Var2Lit(pVars[VarOut], 1-n)
					status := sat_solver_addclause(pSat, pLits[:pStr.nVarIns+2])
					if status == 0 {
						return 0
					}
				}
			}
		} else {
			assert(pStr.nVarIns <= SBD_DIV_MAX, "pStr->nVarIns <= SBD_DIV_MAX")
			for k := int32(0); k < pStr.nVarIns; k, VarPar = k+1, VarPar+1 {
				for n := int32(0); n < 2; n++ {
					pLits[0] = Abc_Var2Lit(pVars[VarPar], 1)
					pLits[1] = Abc_Var2Lit(pVars[VarOut], n)
					pLits[2] = Abc_Var2Lit(pVars[pStr.VarIns[k]], 1-n)
					status := sat_solver_addclause(pSat, pLits[:3])
					if status == 0 {
						return 0
					}
				}
			}
		}
	}
	return 1
}
func Sbd_ProblemAddClausesInit(pSat *sat_solver, nVars, nStrs int32, pVars []int32, pStr0 []Sbd_Str_t) {
	VarPar := nVars + nStrs
	var pLits [SBD_DIV_MAX]int32
	// make sure selector parameters are mutually exclusive
	for s := int32(0); s < nStrs; s++ {
		pStr := &pStr0[s]
		if pStr.fLut == 0 {
			// one variable should be selected
			assert(pStr.nVarIns <= SBD_DIV_MAX, "pStr->nVarIns <= SBD_DIV_MAX")
			for m := int32(0); m < pStr.nVarIns; m++ {
				pLits[m] = Abc_Var2Lit(pVars[VarPar+m], 0)
			}
			status := sat_solver_addclause(pSat, pLits[:pStr.nVarIns])
			assert(status != 0, "status")
			// two variables cannot be selected
			for m := int32(0); m < pStr.nVarIns; m++ {
				for m2 := m + 1; m2 < pStr.nVarIns; m2++ {
					pLits[0] = Abc_Var2Lit(pVars[VarPar+m], 1)
					pLits[1] = Abc_Var2Lit(pVars[VarPar+m2], 1)
					status = sat_solver_addclause(pSat, pLits[:2])
					assert(status != 0, "status")
				}
			}
		}
		if pStr.fLut != 0 {
			VarPar += 1 << uint(pStr.nVarIns)
		} else {
			VarPar += pStr.nVarIns
		}
	}
}
func Sbd_ProblemPrintSolution(nStrs int32, pStr0 []Sbd_Str_t, vLits *Vec_Int_t) {
	iLit := 0
	fmt.Printf("Solution found:\n")
	for s := int32(0); s < nStrs; s++ {
		pStr := &pStr0[s]
		nIters := pStr.nVarIns
		name := "SEL"
		if pStr.fLut != 0 {
			nIters = 1 << uint(pStr.nVarIns)
			name = "LUT"
		}
		fmt.Printf("%s%d : ", name, s)
		for m := int32(0); m < nIters; m, iLit = m+1, iLit+1 {
			fmt.Printf("%d", 1-Abc_LitIsCompl(vec.Vec_IntEntry(vLits, iLit)))
		}
		fmt.Printf("    {")
		for m := int32(0); m < pStr.nVarIns; m++ {
			fmt.Printf(" %d", pStr.VarIns[m])
		}
		fmt.Printf(" }\n")
	}
	assert(iLit == vec.Vec_IntSize(vLits), "iLit == Vec_IntSize(vLits)")
}
func Sbd_ProblemCollectSolution(nStrs int32, pStr0 []Sbd_Str_t, vLits *Vec_Int_t) {
	iLit := 0
	for s := int32(0); s < nStrs; s++ {
		pStr := &pStr0[s]
		pStr.Res = 0
		if pStr.fLut != 0 {
			nIters := int32(1) << uint(pStr.nVarIns)
			for m := int32(0); m < nIters; m, iLit = m+1, iLit+1 {
				if Abc_LitIsCompl(vec.Vec_IntEntry(vLits, iLit)) == 0 {
					pStr.Res |= word(1) << uint(m)
				}
			}
			pStr.Res = util.Abc_Tt6Stretch(pStr.Res, pStr.nVarIns)
		} else {
			nIters := 0
			for m := int32(0); m < pStr.nVarIns; m, iLit = m+1, iLit+1 {
				if Abc_LitIsCompl(vec.Vec_IntEntry(vLits, iLit)) == 0 {
					pStr.Res = word(pStr.VarIns[m])
					nIters++
				}
			}
			assert(nIters == 1, "nIters == 1")
		}
	}
	assert(iLit == vec.Vec_IntSize(vLits), "iLit == Vec_IntSize(vLits)")
}

// Sbd_ProblemSolve solves QBF problem for the given window.
func Sbd_ProblemSolve(p *Gia_Man_t, vMirrors *Vec_Int_t,
	Pivot int32, vWinObjs, vObj2Var *Vec_Int_t,
	vTfo, vRoots *Vec_Int_t,
	vDivSet *Vec_Int_t, nStrs int32, pStr0 []Sbd_Str_t) int32 { // divisors, structures
	vLits := vec.Vec_IntAlloc(100)
	pSatCec := Sbd_ManSatSolver(nil, p, vMirrors, Pivot, vWinObjs, vObj2Var, vTfo, vRoots, 1)
	pSatQbf := sat_solver_new()
	nVars := int32(vec.Vec_IntSize(vDivSet))
	nPars := Sbd_ProblemCountParams(nStrs, pStr0)
	VarCecOut := int32(vec.Vec_IntSize(vWinObjs) + vec.Vec_IntSize(vTfo) + vec.Vec_IntSize(vRoots))
	VarCecPar := VarCecOut + nStrs
	VarQbfPar := int32(0)
	VarQbfFree := nPars
	var pVarsCec [256]int32
	var pVarsQbf [256]int32
	RetValue := int32(0)
	assert(vec.Vec_IntSize(vDivSet) <= SBD_DIV_MAX, "Vec_IntSize(vDivSet) <= SBD_DIV_MAX")
	assert(nVars+nStrs+nPars <= 256, "nVars + nStrs + nPars <= 256")
	// collect CEC variables
	for i, iVar := range vec.Vec_IntArray(vDivSet) {
		pVarsCec[i] = iVar
	}
	for i := int32(0); i < nStrs; i++ {
		pVarsCec[nVars+i] = VarCecOut + i
	}
	for i := int32(0); i < nPars; i++ {
		pVarsCec[nVars+nStrs+i] = VarCecPar + i
	}
	// collect QBF variables
	for i := int32(0); i < nVars+nStrs; i++ {
		pVarsQbf[i] = -1
	}
	for i := int32(0); i < nPars; i++ {
		pVarsQbf[nVars+nStrs+i] = VarQbfPar + i
	}
	// add clauses to the CEC problem
	Sbd_ProblemAddClauses(pSatCec, nVars, nStrs, pVarsCec[:], pStr0)
	// create QBF solver
	sat_solver_setnvars(pSatQbf, 1000)
	Sbd_ProblemAddClausesInit(pSatQbf, nVars, nStrs, pVarsQbf[:], pStr0)
	// assume all parameter variables are 0
	vec.Vec_IntClear(vLits)
	for i := int32(0); i < nPars; i++ {
		vec.Vec_IntPush(vLits, Abc_Var2Lit(VarCecPar+i, 1))
	}
	for nIters := int32(0); nIters < (1 << uint(nVars)); nIters++ {
		// check if these parameters solve the problem
		status := sat_solver_solve(pSatCec, vec.Vec_IntArray(vLits), 0, 0, 0, 0)
		if status == l_False { // solution found
			break
		}
		assert(status == l_True, "status == l_True")
		vec.Vec_IntClear(vLits)
		// create new QBF variables
		for i := int32(0); i < nVars+nStrs; i++ {
			pVarsQbf[i] = VarQbfFree
			VarQbfFree++
		}
		// set their values
		for i, iVar := range vec.Vec_IntArray(vDivSet) {
			iLit := Abc_Var2Lit(pVarsQbf[i], 1-sat_solver_var_value(pSatCec, iVar))
			status = sat_solver_addclause(pSatQbf, []int32{iLit})
			assert(status != 0, "status")
		}
		iLit := Abc_Var2Lit(pVarsQbf[nVars], sat_solver_var_value(pSatCec, VarCecOut))
		status = sat_solver_addclause(pSatQbf, []int32{iLit})
		assert(status != 0, "status")
		// add clauses to the QBF problem
		if Sbd_ProblemAddClauses(pSatQbf, nVars, nStrs, pVarsQbf[:], pStr0) == 0 {
			break // solution does not exist
		}
		// check if solution still exists
		status = sat_solver_solve(pSatQbf, nil, 0, 0, 0, 0)
		if status == l_False { // solution does not exist
			break
		}
		assert(status == l_True, "status == l_True")
		// find the new values of parameters
		assert(vec.Vec_IntSize(vLits) == 0, "Vec_IntSize(vLits) == 0")
		for i := int32(0); i < nPars; i++ {
			vec.Vec_IntPush(vLits, Abc_Var2Lit(VarCecPar+i, 1-sat_solver_var_value(pSatQbf, VarQbfPar+i)))
		}
	}
	if vec.Vec_IntSize(vLits) > 0 {
		Sbd_ProblemCollectSolution(nStrs, pStr0, vLits)
		RetValue = 1
	}
	sat_solver_delete(pSatCec)
	sat_solver_delete(pSatQbf)
	return RetValue
}
