package sfm

// Port of src/opt/sfm/sfmSat.c.

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
)

var s_Truths6 = [6]uint64{
	0xAAAAAAAAAAAAAAAA,
	0xCCCCCCCCCCCCCCCC,
	0xF0F0F0F0F0F0F0F0,
	0xFF00FF00FF00FF00,
	0xFFFF0000FFFF0000,
	0xFFFFFFFF00000000,
}

// satVarLiteral mirrors sat_solver_var_literal.
func satVarLiteral(s *bsat.Solver, v int32) int32 {
	if satVarValue(s, v) != 0 {
		return abcVar2Lit(v, 0)
	}
	return abcVar2Lit(v, 1)
}

// satVarValue mirrors sat_solver_var_value.
func satVarValue(s *bsat.Solver, v int32) int32 {
	val, ok := s.Value(v)
	assert(ok, "sat_solver_var_value after a satisfiable call")
	if val {
		return 1
	}
	return 0
}

// satSolve mirrors sat_solver_solve( s, begin, end, nConfLimit, 0, 0, 0 ).
func satSolve(s *bsat.Solver, lits []int32, nConfLimit int64) bsat.Status {
	return s.SolveWithResourceLimits(lits, nConfLimit, 0, 0, 0)
}

// Converts a window into a SAT solver.
func Sfm_NtkWindowToSolver(p *Sfm_Ntk_t) bool {
	// p.vOrder contains all variables in the window in a good order
	// p.vDivs is a subset of nodes in p.vOrder used as divisor candidates
	// p.vTfo contains TFO of the node (does not include node)
	// p.vRoots contains roots of the TFO of the node (may include node)
	clk := time.Now()
	//    if ( p->pSat )
	//        printf( "%d  ", p->pSat->stats.learnts );
	p.pSat.RestartSearch()
	p.pSat.SetNumVars(int32(1 + len(p.vOrder) + len(p.vTfo) + len(p.vRoots) + 10))
	// create SAT variables
	Sfm_NtkCleanVars(p)
	p.nSatVars = 1
	for _, iNode := range p.vOrder {
		Sfm_ObjSetSatVar(p, iNode, p.nSatVars)
		p.nSatVars++
	}
	// collect divisor variables
	p.vDivVars = p.vDivVars[:0]
	for _, iNode := range p.vDivs {
		p.vDivVars = append(p.vDivVars, Sfm_ObjSatVar(p, iNode))
	}
	// add CNF clauses for the TFI
	for _, iNode := range p.vOrder {
		if Sfm_ObjIsPi(p, iNode) {
			continue
		}
		// collect fanin variables
		p.vFaninMap = p.vFaninMap[:0]
		for _, iFanin := range p.Fanins[iNode] {
			p.vFaninMap = append(p.vFaninMap, Sfm_ObjSatVar(p, iFanin))
		}
		p.vFaninMap = append(p.vFaninMap, Sfm_ObjSatVar(p, iNode))
		// generate CNF
		Sfm_TranslateCnf(&p.vClauses, p.vCnfs[iNode], p.vFaninMap, -1)
		// add clauses
		for _, vClause := range p.vClauses {
			if len(vClause) == 0 {
				break
			}
			if !p.pSat.AddClause(vClause...) {
				return false
			}
		}
	}
	if len(p.vTfo) > 0 {
		assert(p.pPars.TfoLevMax > 0, "p->pPars->nTfoLevMax > 0")
		assert(len(p.vRoots) > 0, "Vec_IntSize(p->vRoots) > 0")
		assert(p.vTfo[0] != p.iPivotNode, "Vec_IntEntry(p->vTfo, 0) != p->iPivotNode")
		// collect variables of root nodes
		p.vLits = p.vLits[:0]
		for _, iNode := range p.vRoots {
			p.vLits = append(p.vLits, Sfm_ObjSatVar(p, iNode))
		}
		// assign new variables to the TFO nodes
		for _, iNode := range p.vTfo {
			Sfm_ObjCleanSatVar(p, Sfm_ObjSatVar(p, iNode))
			Sfm_ObjSetSatVar(p, iNode, p.nSatVars)
			p.nSatVars++
		}
		// add CNF clauses for the TFO
		for _, iNode := range p.vTfo {
			assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
			// collect fanin variables
			p.vFaninMap = p.vFaninMap[:0]
			for _, iFanin := range p.Fanins[iNode] {
				p.vFaninMap = append(p.vFaninMap, Sfm_ObjSatVar(p, iFanin))
			}
			p.vFaninMap = append(p.vFaninMap, Sfm_ObjSatVar(p, iNode))
			// generate CNF
			Sfm_TranslateCnf(&p.vClauses, p.vCnfs[iNode], p.vFaninMap, Sfm_ObjSatVar(p, p.iPivotNode))
			// add clauses
			for _, vClause := range p.vClauses {
				if len(vClause) == 0 {
					break
				}
				if !p.pSat.AddClause(vClause...) {
					return false
				}
			}
		}
		// create XOR clauses for the roots
		for i, iNode := range p.vRoots {
			p.pSat.AddXor(p.vLits[i], Sfm_ObjSatVar(p, iNode), p.nSatVars, false)
			p.nSatVars++
			p.vLits[i] = abcVar2Lit(p.nSatVars-1, 0)
		}
		// make OR clause for the last nRoots variables
		if !p.pSat.AddClause(p.vLits...) {
			return false
		}
	}
	// finalize
	RetValue := p.pSat.Simplify()
	p.timeCnf += time.Since(clk)
	return RetValue
}

// Takes SAT solver and returns interpolant.
// If interpolant does not exist, records difference variables.
func Sfm_ComputeInterpolant(p *Sfm_Ntk_t) uint64 {
	var uCube, uTruth uint64
	var status bsat.Status
	nIter := 0
	var pLits [2]int32
	nVars := p.pSat.NumVars()
	p.pSat.SetNumVars(nVars + 1)
	pLits[0] = abcVar2Lit(Sfm_ObjSatVar(p, p.iPivotNode), 0) // F = 1
	pLits[1] = abcVar2Lit(nVars, 0)                          // iNewLit
	assert(len(p.vDivIds) <= 6, "Vec_IntSize(p->vDivIds) <= 6")
	for {
		// find onset minterm
		p.Stats.SATCalls++
		status = satSolve(p.pSat, pLits[:], int64(p.pPars.BTLimit))
		if status == bsat.Unknown {
			return SFM_SAT_UNDEC
		}
		if status == bsat.Unsatisfiable {
			return uTruth
		}
		assert(status == bsat.Satisfiable, "status == l_True")
		// remember variable values
		p.vValues = p.vValues[:0]
		for _, iVar := range p.vDivVars {
			p.vValues = append(p.vValues, satVarValue(p.pSat, iVar))
		}
		// collect divisor literals
		p.vLits = p.vLits[:0]
		p.vLits = append(p.vLits, abcLitNot(pLits[0])) // F = 0
		for _, Div := range p.vDivIds {
			p.vLits = append(p.vLits, satVarLiteral(p.pSat, Div))
		}
		// check against offset
		p.Stats.SATCalls++
		status = satSolve(p.pSat, p.vLits, int64(p.pPars.BTLimit))
		if status == bsat.Unknown {
			return SFM_SAT_UNDEC
		}
		if status == bsat.Satisfiable {
			break
		}
		assert(status == bsat.Unsatisfiable, "status == l_False")
		// compute cube and add clause
		pFinal := p.pSat.UnsatCore()
		uCube = ^uint64(0)
		p.vLits = p.vLits[:0]
		p.vLits = append(p.vLits, abcLitNot(pLits[1])) // NOT(iNewLit)
		for i := 0; i < len(pFinal); i++ {
			if pFinal[i] == pLits[0] {
				continue
			}
			p.vLits = append(p.vLits, pFinal[i])
			iVar := int32Find(p.vDivIds, abcLit2Var(pFinal[i]))
			assert(iVar >= 0, "iVar >= 0")
			if abcLitIsCompl(pFinal[i]) {
				uCube &= s_Truths6[iVar]
			} else {
				uCube &= ^s_Truths6[iVar]
			}
		}
		uTruth |= uCube
		ok := p.pSat.AddClause(p.vLits...)
		assert(ok, "status")
		nIter++
	}
	assert(status == bsat.Satisfiable, "status == l_True")
	// store the counter-example
	for i, iVar := range p.vDivVars {
		if p.vValues[i]^satVarValue(p.pSat, iVar) != 0 { // insert 1
			assert(p.vDivCexes[i]&(uint64(1)<<uint(p.nCexes)) == 0, "!Abc_InfoHasBit( (unsigned *)pSign, p->nCexes)")
			p.vDivCexes[i] ^= uint64(1) << uint(p.nCexes)
		}
	}
	p.nCexes++
	return SFM_SAT_SAT
}

func int32Find(a []int32, Entry int32) int {
	for i, e := range a {
		if e == Entry {
			return i
		}
	}
	return -1
}

// Sfm_ComputeInterpolantInt takes the SAT solver and computes the care onset
// and offset (Truth[1] and Truth[0]) over the divisors in vDivIds.
func Sfm_ComputeInterpolantInt(p *Sfm_Ntk_t, Truth *[2]uint64) bsat.Status {
	var iMint int
	var fOnSet int32
	nVars := p.pSat.NumVars()
	iVarPivot := Sfm_ObjSatVar(p, p.iPivotNode)
	var status bsat.Status
	nIter := 0
	Truth[0], Truth[1] = 0, 0
	p.pSat.SetNumVars(nVars + 1)
	iNewLit := abcVar2Lit(nVars, 0) // iNewLit
	assert(len(p.vDivIds) <= 6, "Vec_IntSize(p->vDivIds) <= 6")
	nValues := (1 << uint(len(p.vDivIds))) * len(p.vDivVars)
	p.vValues = p.vValues[:0]
	for i := 0; i < nValues; i++ {
		p.vValues = append(p.vValues, -1)
	}
	for {
		// find care minterm
		p.Stats.SATCalls++
		nIter++
		status = satSolve(p.pSat, []int32{iNewLit}, int64(p.pPars.BTLimit))
		if status == bsat.Unknown {
			return bsat.Unknown
		}
		if status == bsat.Unsatisfiable {
			return bsat.Unsatisfiable
		}
		assert(status == bsat.Satisfiable, "status == l_True")
		// collect values
		iMint = 0
		fOnSet = satVarValue(p.pSat, iVarPivot)
		p.vLits = p.vLits[:0]
		p.vLits = append(p.vLits, abcLitNot(iNewLit)) // NOT(iNewLit)
		p.vLits = append(p.vLits, abcLitNot(satVarLiteral(p.pSat, iVarPivot)))
		for i, Div := range p.vDivIds {
			p.vLits = append(p.vLits, abcLitNot(satVarLiteral(p.pSat, Div)))
			if satVarValue(p.pSat, Div) != 0 {
				iMint |= 1 << uint(i)
			}
		}
		if Truth[1-fOnSet]&(uint64(1)<<uint(iMint)) != 0 {
			break
		}
		assert(Truth[fOnSet]&(uint64(1)<<uint(iMint)) == 0, "!(Truth[fOnSet] & ((word)1 << iMint))")
		Truth[fOnSet] |= uint64(1) << uint(iMint)
		// remember variable values
		for i, iVar := range p.vDivVars {
			p.vValues[iMint*len(p.vDivVars)+i] = satVarValue(p.pSat, iVar)
		}
		if !p.pSat.AddClause(p.vLits...) {
			return bsat.Unsatisfiable
		}
	}
	assert(status == bsat.Satisfiable, "status == l_True")
	// store the counter-example
	assert(iMint < (1<<uint(len(p.vDivIds))), "iMint < (1 << Vec_IntSize(p->vDivIds))")
	for i, iVar := range p.vDivVars {
		Value := p.vValues[iMint*len(p.vDivVars)+i]
		assert(Value != -1, "Value != -1")
		if Value^satVarValue(p.pSat, iVar) != 0 { // insert 1
			assert(p.vDivCexes[i]&(uint64(1)<<uint(p.nCexes)) == 0, "!Abc_InfoHasBit( (unsigned *)pSign, p->nCexes)")
			p.vDivCexes[i] ^= uint64(1) << uint(p.nCexes)
		}
	}
	p.nCexes++
	return bsat.Satisfiable
}
func Sfm_ComputeInterpolant2(p *Sfm_Ntk_t) uint64 {
	var Res, ResP, ResN uint64
	var Truth [2]uint64
	var nCubesP, nCubesN int32
	RetValue := Sfm_ComputeInterpolantInt(p, &Truth)
	if RetValue == bsat.Unknown {
		return SFM_SAT_UNDEC
	}
	if RetValue == bsat.Satisfiable {
		return SFM_SAT_SAT
	}
	assert(RetValue == bsat.Unsatisfiable, "RetValue == l_False")
	//printf( "Offset = %2d.  Onset = %2d.  DC = %2d.  Total = %2d.\n",
	//    Abc_TtCountOnes(Truth[0]), Abc_TtCountOnes(Truth[1]),
	//    (1<<Vec_IntSize(p->vDivIds)) - (Abc_TtCountOnes(Truth[0]) + Abc_TtCountOnes(Truth[1])),
	//     1<<Vec_IntSize(p->vDivIds) );
	n := int32(len(p.vDivIds))
	Truth[0] = util.Abc_Tt6Stretch(Truth[0], n)
	Truth[1] = util.Abc_Tt6Stretch(Truth[1], n)
	ResP = util.Abc_Tt6Isop(Truth[1], ^Truth[0], n, &nCubesP)
	ResN = util.Abc_Tt6Isop(Truth[0], ^Truth[1], n, &nCubesN)
	if nCubesP <= nCubesN {
		Res = ResP
	} else {
		Res = ^ResN
	}
	//Dau_DsdPrintFromTruth( &Res, Vec_IntSize(p->vDivIds) );
	return Res
}

// Checks resubstitution.
func Sfm_ComputeInterpolantCheck(p *Sfm_Ntk_t) {
	iNode := 3
	iDiv0 := 6
	iDiv1 := 7
	//    int i;
	//    Sfm_NtkForEachNode( p, i )
	{
		Sfm_NtkCreateWindow(p, iNode, true)
		Sfm_NtkWindowToSolver(p)
		// collect SAT variables of divisors
		p.vDivIds = p.vDivIds[:0]
		p.vDivIds = append(p.vDivIds, Sfm_ObjSatVar(p, iDiv0))
		p.vDivIds = append(p.vDivIds, Sfm_ObjSatVar(p, iDiv1))
		uTruth := Sfm_ComputeInterpolant(p)
		if uTruth == SFM_SAT_SAT {
			fmt.Printf("The problem is SAT.\n")
		} else if uTruth == SFM_SAT_UNDEC {
			fmt.Printf("The problem is UNDEC.\n")
		} else {
			kitDsdPrintFromTruth(uTruth, 2)
			fmt.Printf("\n")
		}
	}
}

func kitDsdPrintFromTruth(uTruth uint64, nVars int) {
	kit.Kit_DsdPrintFromTruth([]uint32{uint32(uTruth), uint32(uTruth >> 32)}, int32(nVars))
}
