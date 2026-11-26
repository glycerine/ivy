package mfs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Port of mfsInter.c.

// Abc_MfsSatAddXor adds constraints for the two-input AND-gate.
func Abc_MfsSatAddXor(pSat *bsat.Solver, iVarA, iVarB, iVarC int32) int32 {
	var Lits [3]int32

	Lits[0] = toLitCond(iVarA, 1)
	Lits[1] = toLitCond(iVarB, 1)
	Lits[2] = toLitCond(iVarC, 1)
	if sat_solver_addclause(pSat, Lits[:]) == 0 {
		return 0
	}

	Lits[0] = toLitCond(iVarA, 1)
	Lits[1] = toLitCond(iVarB, 0)
	Lits[2] = toLitCond(iVarC, 0)
	if sat_solver_addclause(pSat, Lits[:]) == 0 {
		return 0
	}

	Lits[0] = toLitCond(iVarA, 0)
	Lits[1] = toLitCond(iVarB, 1)
	Lits[2] = toLitCond(iVarC, 0)
	if sat_solver_addclause(pSat, Lits[:]) == 0 {
		return 0
	}

	Lits[0] = toLitCond(iVarA, 0)
	Lits[1] = toLitCond(iVarB, 0)
	Lits[2] = toLitCond(iVarC, 1)
	if sat_solver_addclause(pSat, Lits[:]) == 0 {
		return 0
	}
	return 1
}

// cnfShiftLiterals adds delta to every literal of the CNF (the C code does
// this through pClauses[0][i] for i < nLiterals).
func cnfShiftLiterals(pCnf *cnf.Cnf_Dat_t, delta int32) {
	for _, c := range cnf.Cnf_DataClauses(pCnf) {
		for i := range c {
			c[i] += delta
		}
	}
}

// Abc_MfsCreateSolverResub creates miter for the resubstitution problem.
// pCands == nil corresponds to the C call with pCands == NULL.
func Abc_MfsCreateSolverResub(p *Mfs_Man_t, pCands []int32, fInvert int32) *bsat.Solver {
	var Lits [2]int32
	nVars := cnf.Cnf_DataVarNum(p.pCnf)
	nDivs := int32(len(p.vDivs))
	clauses := cnf.Cnf_DataClauses(p.pCnf)

	// get the literal for the output of F
	nCos := aig.Aig_ManCoNum(p.pAigWin)
	pObjPo := aig.Aig_ManCo(p.pAigWin, nCos-nDivs-1)
	Lits[0] = toLitCond(p.pCnf.VarNums[pObjPo.NodeID()], fInvert)

	// collect the outputs of the divisors
	p.vProjVarsCnf = p.vProjVarsCnf[:0]
	for i := nCos - nDivs; i < nCos; i++ {
		pObjPo = aig.Aig_ManCo(p.pAigWin, i)
		assert(p.pCnf.VarNums[pObjPo.NodeID()] >= 0, "p->pCnf->pVarNums[pObjPo->Id] >= 0")
		p.vProjVarsCnf = append(p.vProjVarsCnf, p.pCnf.VarNums[pObjPo.NodeID()])
	}
	assert(int32(len(p.vProjVarsCnf)) == nDivs, "Vec_IntSize(p->vProjVarsCnf) == Vec_PtrSize(p->vDivs)")

	// start the solver
	pSat := bsat.NewSolver()
	pSat.SetNumVars(2*nVars + nDivs)
	if pCands != nil {
		pSat.EnableClauseStore()
	}

	// load the first copy of the clauses
	for _, c := range clauses {
		if sat_solver_addclause(pSat, c) == 0 {
			pSat.Delete()
			return nil
		}
	}
	// add the clause for the first output of F
	if sat_solver_addclause(pSat, Lits[:1]) == 0 {
		pSat.Delete()
		return nil
	}
	// add one-hotness constraints
	if p.pPars.FOneHotness != 0 {
		p.pSat = pSat
		if Abc_NtkAddOneHotness(p) == 0 {
			return nil
		}
		p.pSat = nil
	}
	// bookmark the clauses of A
	if pCands != nil {
		pSat.MarkClauseStorePartitionA()
	}
	// transform the literals
	cnfShiftLiterals(p.pCnf, 2*nVars)
	// load the second copy of the clauses
	for _, c := range clauses {
		if sat_solver_addclause(pSat, c) == 0 {
			cnfShiftLiterals(p.pCnf, -2*nVars)
			pSat.Delete()
			return nil
		}
	}
	// add one-hotness constraints
	if p.pPars.FOneHotness != 0 {
		p.pSat = pSat
		if Abc_NtkAddOneHotness(p) == 0 {
			cnfShiftLiterals(p.pCnf, -2*nVars)
			return nil
		}
		p.pSat = nil
	}
	// transform the literals
	cnfShiftLiterals(p.pCnf, -2*nVars)
	// add the clause for the second output of F
	Lits[0] = 2*nVars + lit_neg(Lits[0])
	if sat_solver_addclause(pSat, Lits[:1]) == 0 {
		pSat.Delete()
		return nil
	}

	if pCands != nil {
		// add relevant clauses for EXOR gates
		for c := range pCands {
			// get the variable number of this divisor
			i := lit_var(pCands[c]) - 2*nVars
			// get the corresponding SAT variable
			iVar := p.vProjVarsCnf[i]
			// add the corresponding EXOR gate
			if Abc_MfsSatAddXor(pSat, iVar, iVar+nVars, 2*nVars+i) == 0 {
				pSat.Delete()
				return nil
			}
			// add the corresponding clause
			if sat_solver_addclause(pSat, pCands[c:c+1]) == 0 {
				pSat.Delete()
				return nil
			}
		}
		// bookmark the roots
		pSat.MarkClauseStoreRoots()
	} else {
		// add the clauses for the EXOR gates - and remember their outputs
		p.vProjVarsSat = p.vProjVarsSat[:0]
		for i, iVar := range p.vProjVarsCnf {
			if Abc_MfsSatAddXor(pSat, iVar, iVar+nVars, 2*nVars+int32(i)) == 0 {
				pSat.Delete()
				return nil
			}
			p.vProjVarsSat = append(p.vProjVarsSat, 2*nVars+int32(i))
		}
		assert(len(p.vProjVarsCnf) == len(p.vProjVarsSat), "Vec_IntSize(p->vProjVarsCnf) == Vec_IntSize(p->vProjVarsSat)")
		// simplify the solver
		if !pSat.Simplify() {
			pSat.Delete()
			return nil
		}
	}
	return pSat
}

// Abc_NtkMfsInterplateTruth performs interpolation and returns its truth table.
func Abc_NtkMfsInterplateTruth(p *Mfs_Man_t, pCands []int32, fInvert int32) []uint32 {
	nCands := int32(len(pCands))
	nVars := cnf.Cnf_DataVarNum(p.pCnf)
	// derive the SAT solver for interpolation
	pSat := Abc_MfsCreateSolverResub(p, pCands, fInvert)

	// solve the problem
	status := sat_solver_solve(pSat, nil, int64(p.pPars.NBTLimit), 0, 0, 0)
	if status != l_False {
		p.nTimeOuts++
		return nil
	}
	// get the learned clauses
	pCnf := pSat.ReleaseClauseStore()
	pSat.Delete()

	// set the global variables
	pGloVars := bsat.Int_ManSetGlobalVars(p.pMan, nCands)
	for c := int32(0); c < nCands; c++ {
		// get the variable number of this divisor
		i := lit_var(pCands[c]) - 2*nVars
		// get the corresponding SAT variable
		pGloVars[c] = p.vProjVarsCnf[i]
	}

	// derive the interpolant
	var puTruth []uint32
	nFanins := bsat.Int_ManInterpolate(p.pMan, pCnf, 0, &puTruth)
	bsat.Sto_ManFree(pCnf)
	assert(nFanins == nCands, "nFanins == nCands")
	return puTruth
}

// Abc_NtkMfsInterplateEval counts the number of don't-care minterms.
func Abc_NtkMfsInterplateEval(p *Mfs_Man_t, pCands []int32) int32 {
	var uTruth0, uTruth1 [2]uint32
	nCands := len(pCands)
	pTruth := Abc_NtkMfsInterplateTruth(p, pCands, 0)
	if nCands == 6 {
		uTruth1[0] = pTruth[0]
		uTruth1[1] = pTruth[1]
	} else {
		uTruth1[0] = pTruth[0]
		uTruth1[1] = pTruth[0]
	}
	pTruth = Abc_NtkMfsInterplateTruth(p, pCands, 1)
	if nCands == 6 {
		uTruth0[0] = ^pTruth[0]
		uTruth0[1] = ^pTruth[1]
	} else {
		uTruth0[0] = ^pTruth[0]
		uTruth0[1] = ^pTruth[0]
	}
	nCounter := extra.Extra_WordCountOnes(uTruth0[0] ^ uTruth1[0])
	nCounter += extra.Extra_WordCountOnes(uTruth0[1] ^ uTruth1[1])
	return nCounter
}

// Abc_NtkMfsInterplate performs interpolation and returns the new function.
func Abc_NtkMfsInterplate(p *Mfs_Man_t, pCands []int32) *hop.Hop_Obj_t {
	nCands := int32(len(pCands))
	nVars := cnf.Cnf_DataVarNum(p.pCnf)

	// derive the SAT solver for interpolation
	pSat := Abc_MfsCreateSolverResub(p, pCands, 0)

	// solve the problem
	status := sat_solver_solve(pSat, nil, int64(p.pPars.NBTLimit), 0, 0, 0)
	if status != l_False {
		p.nTimeOuts++
		return nil
	}
	// get the learned clauses
	pCnf := pSat.ReleaseClauseStore()
	pSat.Delete()

	// set the global variables
	pGloVars := bsat.Int_ManSetGlobalVars(p.pMan, nCands)
	for c := int32(0); c < nCands; c++ {
		// get the variable number of this divisor
		i := lit_var(pCands[c]) - 2*nVars
		// get the corresponding SAT variable
		pGloVars[c] = p.vProjVarsCnf[i]
	}

	// derive the interpolant
	var puTruth []uint32
	nFanins := bsat.Int_ManInterpolate(p.pMan, pCnf, 0, &puTruth)
	bsat.Sto_ManFree(pCnf)
	assert(nFanins == nCands, "nFanins == nCands")

	// transform interpolant into AIG
	pGraph := kit.Kit_TruthToGraph(puTruth, nFanins, &p.vMem)
	pFunc := kit.Kit_GraphToHop(p.pNtk.FunctionManager.(*hop.Hop_Man_t), pGraph)
	kit.Kit_GraphFree(pGraph)
	return pFunc
}
