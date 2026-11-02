package abc

import (
	"fmt"
	"os"
)

// Port of the SimpSolver based part of src/sat/glucose/AbcGlucose.cpp: the
// glucose_solver_* wrappers, the extern "C" bmcg_sat_solver_* entry points and the
// CNF-file driver. `#define USE_SIMP_SOLVER 1` is set at the top of the file, so
// only the SimpSolver variant is live. The functions that need the Gia package
// (Glucose_SolverFromAig*, Glucose_GenerateCubes, the quantification and
// equivalence helpers, Glucose_SolveAig) remain stubs in skel_sat_glucose_AbcGlucose*.go.

// GLUCOSE_UNSAT, GLUCOSE_SAT and GLUCOSE_UNDEC of AbcGlucose.h.
const (
	GlucoseUnsat = -1
	GlucoseSat   = 1
	GlucoseUndec = 0
)

// glucoseSimp is the cast `(Gluco::SimpSolver*)s` of a bmcg_sat_solver pointer.
func glucoseSimp(s BmcgSatSolver) *GlucoSimpSolver { return s.(*GlucoSimpSolver) }

// glucoAbcVar2Lit is Abc_Var2Lit (misc/util/abc_global.h).
func glucoAbcVar2Lit(v int, c int) int { return v + v + c }

// glucoAbcPrintTime is Abc_PrintTime (misc/util/abc_global.h): ABC_PRT(pStr, time).
func glucoAbcPrintTime(level int, pStr string, t int64) {
	fmt.Printf("%s =", pStr)
	fmt.Printf("%9.2f sec\n", 1.0*float64(t)/float64(AbcClockTicksPerSecond))
}

// glucose_solver_start
// C: SimpSolver * glucose_solver_start()
func glucose_solver_start() *GlucoSimpSolver {
	S := NewGlucoSimpSolver()
	S.SetIncrementalMode()
	return S
}

// glucose_solver_stop
// C: void glucose_solver_stop(Gluco::SimpSolver* S)
func glucose_solver_stop(S *GlucoSimpSolver) {
	S.Destroy()
}

// glucose_solver_reset
// C: void glucose_solver_reset(Gluco::SimpSolver* S)
func glucose_solver_reset(S *GlucoSimpSolver) {
	S.Reset()
}

// glucose_solver_addclause
// C: int glucose_solver_addclause(Gluco::SimpSolver* S, int * plits, int nlits)
func glucose_solver_addclause(S *GlucoSimpSolver, plits []int, nlits int) int {
	var lits GlucoVec[GlucoLit]
	for i := 0; i < nlits; i++ {
		// note: Glucose uses the same var->lit conventiaon as ABC
		for plits[i]/2 >= S.NVars() {
			S.NewVar(true, true)
		}
		if plits[i]/2 >= S.NVars() { // NOTE: since we explicitely use new function bmc_add_var
			panic("glucose_solver_addclause: literal out of range")
		}
		lits.PushElem(GlucoLit{x: int32(plits[i])})
	}
	if S.AddClause(&lits) { // returns 0 if the problem is UNSAT
		return 1
	}
	return 0
}

// glucose_solver_setcallback
// C: void glucose_solver_setcallback(Gluco::SimpSolver* S, void * pman, int(*pfunc)(void*, int, int*))
func glucose_solver_setcallback(S *GlucoSimpSolver, pman any, pfunc func(any, int, []int) int) {
	S.PCnfMan = pman
	S.PCnfFunc = pfunc
	S.NCallConfl = 1000
}

// glucose_solver_solve
// C: int glucose_solver_solve(Gluco::SimpSolver* S, int * plits, int nlits)
func glucose_solver_solve(S *GlucoSimpSolver, plits []int, nlits int) int {
	var lits GlucoVec[GlucoLit]
	for i := 0; i < nlits; i++ {
		lits.PushElem(GlucoLit{x: int32(plits[i])})
	}
	res := S.SolveLimited(&lits, false, false)
	if res == GlucoLTrue {
		return 1
	} else if res == GlucoLFalse {
		return -1
	}
	return 0
}

// glucose_solver_addvar
// C: int glucose_solver_addvar(Gluco::SimpSolver* S)
func glucose_solver_addvar(S *GlucoSimpSolver) int {
	S.NewVar(true, true)
	return S.NVars() - 1
}

// glucose_solver_read_cex
// C: int * glucose_solver_read_cex(Gluco::SimpSolver* S )
func glucose_solver_read_cex(S *GlucoSimpSolver) []int {
	return S.GetCex()
}

// glucose_solver_read_cex_varvalue
// C: int glucose_solver_read_cex_varvalue(Gluco::SimpSolver* S, int ivar)
func glucose_solver_read_cex_varvalue(S *GlucoSimpSolver, ivar int) int {
	if *S.Model.At(ivar) == GlucoLTrue {
		return 1
	}
	return 0
}

// glucose_solver_setstop
// C: void glucose_solver_setstop(Gluco::SimpSolver* S, int * pstop)
func glucose_solver_setstop(S *GlucoSimpSolver, pstop *int) {
	S.Pstop = pstop
}

// bmcg_sat_solver_start
// C: extern bmcg_sat_solver * bmcg_sat_solver_start()
func bmcg_sat_solver_start() BmcgSatSolver {
	return glucose_solver_start()
}

// bmcg_sat_solver_stop
// C: extern void bmcg_sat_solver_stop(bmcg_sat_solver* s)
func bmcg_sat_solver_stop(s BmcgSatSolver) {
	glucose_solver_stop(glucoseSimp(s))
}

// bmcg_sat_solver_reset
// C: extern void bmcg_sat_solver_reset(bmcg_sat_solver* s)
func bmcg_sat_solver_reset(s BmcgSatSolver) {
	glucose_solver_reset(glucoseSimp(s))
}

// bmcg_sat_solver_addclause
// C: extern int bmcg_sat_solver_addclause(bmcg_sat_solver* s, int * plits, int nlits)
func bmcg_sat_solver_addclause(s BmcgSatSolver, plits []int, nlits int) int {
	return glucose_solver_addclause(glucoseSimp(s), plits, nlits)
}

// bmcg_sat_solver_setcallback
// C: extern void bmcg_sat_solver_setcallback(bmcg_sat_solver* s, void * pman, int(*pfunc)(void*, int, int*))
func bmcg_sat_solver_setcallback(s BmcgSatSolver, pman any, pfunc func(any, int, []int) int) {
	glucose_solver_setcallback(glucoseSimp(s), pman, pfunc)
}

// bmcg_sat_solver_solve
// C: extern int bmcg_sat_solver_solve(bmcg_sat_solver* s, int * plits, int nlits)
func bmcg_sat_solver_solve(s BmcgSatSolver, plits []int, nlits int) int {
	return glucose_solver_solve(glucoseSimp(s), plits, nlits)
}

// bmcg_sat_solver_final: *ppArray receives the final conflict as literal codes (a copy of conflict's Lit.x).
// C: extern int bmcg_sat_solver_final(bmcg_sat_solver* s, int ** ppArray)
func bmcg_sat_solver_final(s BmcgSatSolver, ppArray *[]int) int {
	c := &glucoseSimp(s).Conflict
	arr := make([]int, c.Size())
	for i := range arr {
		arr[i] = int(c.At(i).x)
	}
	*ppArray = arr
	return c.Size()
}

// bmcg_sat_solver_addvar
// C: extern int bmcg_sat_solver_addvar(bmcg_sat_solver* s)
func bmcg_sat_solver_addvar(s BmcgSatSolver) int {
	return glucose_solver_addvar(glucoseSimp(s))
}

// bmcg_sat_solver_set_nvars
// C: extern void bmcg_sat_solver_set_nvars(bmcg_sat_solver* s, int nvars)
func bmcg_sat_solver_set_nvars(s BmcgSatSolver, nvars int) {
	for i := bmcg_sat_solver_varnum(s); i < nvars; i++ {
		bmcg_sat_solver_addvar(s)
	}
}

// bmcg_sat_solver_eliminate
// C: extern int bmcg_sat_solver_eliminate(bmcg_sat_solver* s, int turn_off_elim)
func bmcg_sat_solver_eliminate(s BmcgSatSolver, turn_off_elim int) int {
	if glucoseSimp(s).Eliminate(turn_off_elim != 0) {
		return 1
	}
	return 0
}

// bmcg_sat_solver_var_is_elim
// C: extern int bmcg_sat_solver_var_is_elim(bmcg_sat_solver* s, int v)
func bmcg_sat_solver_var_is_elim(s BmcgSatSolver, v int) int {
	if glucoseSimp(s).IsEliminated(v) {
		return 1
	}
	return 0
}

// bmcg_sat_solver_var_set_frozen
// C: extern void bmcg_sat_solver_var_set_frozen(bmcg_sat_solver* s, int v, int freeze)
func bmcg_sat_solver_var_set_frozen(s BmcgSatSolver, v int, freeze int) {
	glucoseSimp(s).SetFrozen(v, freeze != 0)
}

// bmcg_sat_solver_elim_varnum
// C: extern int bmcg_sat_solver_elim_varnum(bmcg_sat_solver* s)
func bmcg_sat_solver_elim_varnum(s BmcgSatSolver) int {
	return glucoseSimp(s).Eliminated_vars
}

// bmcg_sat_solver_read_cex
// C: extern int * bmcg_sat_solver_read_cex(bmcg_sat_solver* s)
func bmcg_sat_solver_read_cex(s BmcgSatSolver) []int {
	return glucose_solver_read_cex(glucoseSimp(s))
}

// bmcg_sat_solver_read_cex_varvalue
// C: extern int bmcg_sat_solver_read_cex_varvalue(bmcg_sat_solver* s, int)
func bmcg_sat_solver_read_cex_varvalue(s BmcgSatSolver, ivar int) int {
	return glucose_solver_read_cex_varvalue(glucoseSimp(s), ivar)
}

// bmcg_sat_solver_set_stop (pstop points at a single stop flag)
// C: extern void bmcg_sat_solver_set_stop(bmcg_sat_solver* s, int * pstop)
func bmcg_sat_solver_set_stop(s BmcgSatSolver, pstop *int) {
	glucose_solver_setstop(glucoseSimp(s), pstop)
}

// bmcg_sat_solver_set_runtime_limit
// C: extern abctime bmcg_sat_solver_set_runtime_limit(bmcg_sat_solver* s, abctime Limit)
func bmcg_sat_solver_set_runtime_limit(s BmcgSatSolver, Limit int64) int64 {
	nRuntimeLimit := glucoseSimp(s).NRuntimeLimit
	glucoseSimp(s).NRuntimeLimit = Limit
	return nRuntimeLimit
}

// bmcg_sat_solver_set_conflict_budget
// C: extern void bmcg_sat_solver_set_conflict_budget(bmcg_sat_solver* s, int Limit)
func bmcg_sat_solver_set_conflict_budget(s BmcgSatSolver, Limit int) {
	if Limit > 0 {
		glucoseSimp(s).SetConfBudget(int64(Limit))
	} else {
		glucoseSimp(s).BudgetOff()
	}
}

// bmcg_sat_solver_varnum
// C: extern int bmcg_sat_solver_varnum(bmcg_sat_solver* s)
func bmcg_sat_solver_varnum(s BmcgSatSolver) int { return glucoseSimp(s).NVars() }

// bmcg_sat_solver_clausenum
// C: extern int bmcg_sat_solver_clausenum(bmcg_sat_solver* s)
func bmcg_sat_solver_clausenum(s BmcgSatSolver) int { return glucoseSimp(s).NClauses() }

// bmcg_sat_solver_learntnum
// C: extern int bmcg_sat_solver_learntnum(bmcg_sat_solver* s)
func bmcg_sat_solver_learntnum(s BmcgSatSolver) int { return glucoseSimp(s).NLearnts() }

// bmcg_sat_solver_conflictnum
// C: extern int bmcg_sat_solver_conflictnum(bmcg_sat_solver* s)
func bmcg_sat_solver_conflictnum(s BmcgSatSolver) int { return int(int32(glucoseSimp(s).Conflicts)) }

// bmcg_sat_solver_minimize_assumptions
// C: extern int bmcg_sat_solver_minimize_assumptions(bmcg_sat_solver * s, int * plits, int nlits, int pivot)
func bmcg_sat_solver_minimize_assumptions(s BmcgSatSolver, plits []int, nlits int, pivot int) int {
	array := &glucoseSimp(s).User_vec
	var i, nlitsL, nlitsR, nresL, nresR, status int
	if pivot < 0 || nlits-pivot < 1 {
		panic("bmcg_sat_solver_minimize_assumptions: precondition")
	}
	if nlits-pivot == 1 {
		// since the problem is UNSAT, we try to solve it without assuming the last literal
		// if the result is UNSAT, the last literal can be dropped; otherwise, it is needed
		status = bmcg_sat_solver_solve(s, plits, pivot)
		if status != GlucoseUnsat { // return 1 if the problem is not UNSAT
			return 1
		}
		return 0
	}
	nlitsL = (nlits - pivot) / 2
	nlitsR = (nlits - pivot) - nlitsL
	if nlitsL+nlitsR != nlits-pivot {
		panic("bmcg_sat_solver_minimize_assumptions: split")
	}
	// solve with these assumptions
	status = bmcg_sat_solver_solve(s, plits, pivot+nlitsL)
	if status == GlucoseUnsat { // these are enough
		return bmcg_sat_solver_minimize_assumptions(s, plits, pivot+nlitsL, pivot)
	}
	// these are not enough
	// solve for the right lits
	if nlitsR == 1 {
		nresL = 1
	} else {
		nresL = bmcg_sat_solver_minimize_assumptions(s, plits, nlits, pivot+nlitsL)
	}
	// swap literals
	array.Clear(false)
	for i = 0; i < nlitsL; i++ {
		array.PushElem(plits[pivot+i])
	}
	for i = 0; i < nresL; i++ {
		plits[pivot+i] = plits[pivot+nlitsL+i]
	}
	for i = 0; i < nlitsL; i++ {
		plits[pivot+nresL+i] = *array.At(i)
	}
	// solve with these assumptions
	status = bmcg_sat_solver_solve(s, plits, pivot+nresL)
	if status == GlucoseUnsat { // these are enough
		return nresL
	}
	// solve for the left lits
	if nlitsL == 1 {
		nresR = 1
	} else {
		nresR = bmcg_sat_solver_minimize_assumptions(s, plits, pivot+nresL+nlitsL, pivot+nresL)
	}
	return nresL + nresR
}

// bmcg_sat_solver_add_and
// C: extern int bmcg_sat_solver_add_and(bmcg_sat_solver * s, int iVar, int iVar0, int iVar1, int fCompl0, int fCompl1, int fCompl)
func bmcg_sat_solver_add_and(s BmcgSatSolver, iVar int, iVar0 int, iVar1 int, fCompl0 int, fCompl1 int, fCompl int) int {
	var Lits [3]int
	b2i := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	Lits[0] = glucoAbcVar2Lit(iVar, b2i(fCompl == 0))
	Lits[1] = glucoAbcVar2Lit(iVar0, fCompl0)
	if bmcg_sat_solver_addclause(s, Lits[:], 2) == 0 {
		return 0
	}
	Lits[0] = glucoAbcVar2Lit(iVar, b2i(fCompl == 0))
	Lits[1] = glucoAbcVar2Lit(iVar1, fCompl1)
	if bmcg_sat_solver_addclause(s, Lits[:], 2) == 0 {
		return 0
	}
	Lits[0] = glucoAbcVar2Lit(iVar, fCompl)
	Lits[1] = glucoAbcVar2Lit(iVar0, b2i(fCompl0 == 0))
	Lits[2] = glucoAbcVar2Lit(iVar1, b2i(fCompl1 == 0))
	if bmcg_sat_solver_addclause(s, Lits[:], 3) == 0 {
		return 0
	}
	return 1
}

// bmcg_solver_add_xor (the C file defines this name; the header declares bmcg_sat_solver_add_xor)
// C: int bmcg_solver_add_xor( bmcg_sat_solver * pSat, int iVarA, int iVarB, int iVarC, int fCompl )
func bmcg_solver_add_xor(pSat BmcgSatSolver, iVarA int, iVarB int, iVarC int, fCompl int) int {
	var Lits [3]int
	b2i := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	if iVarA < 0 || iVarB < 0 || iVarC < 0 {
		panic("bmcg_solver_add_xor: negative variable")
	}
	Lits[0] = glucoAbcVar2Lit(iVarA, b2i(fCompl == 0))
	Lits[1] = glucoAbcVar2Lit(iVarB, 1)
	Lits[2] = glucoAbcVar2Lit(iVarC, 1)
	if bmcg_sat_solver_addclause(pSat, Lits[:], 3) == 0 {
		panic("bmcg_solver_add_xor: clause 1")
	}
	Lits[0] = glucoAbcVar2Lit(iVarA, b2i(fCompl == 0))
	Lits[1] = glucoAbcVar2Lit(iVarB, 0)
	Lits[2] = glucoAbcVar2Lit(iVarC, 0)
	if bmcg_sat_solver_addclause(pSat, Lits[:], 3) == 0 {
		panic("bmcg_solver_add_xor: clause 2")
	}
	Lits[0] = glucoAbcVar2Lit(iVarA, fCompl)
	Lits[1] = glucoAbcVar2Lit(iVarB, 1)
	Lits[2] = glucoAbcVar2Lit(iVarC, 0)
	if bmcg_sat_solver_addclause(pSat, Lits[:], 3) == 0 {
		panic("bmcg_solver_add_xor: clause 3")
	}
	Lits[0] = glucoAbcVar2Lit(iVarA, fCompl)
	Lits[1] = glucoAbcVar2Lit(iVarB, 0)
	Lits[2] = glucoAbcVar2Lit(iVarC, 1)
	if bmcg_sat_solver_addclause(pSat, Lits[:], 3) == 0 {
		panic("bmcg_solver_add_xor: clause 4")
	}
	return 4
}

// bmcg_sat_solver_jftr
// C: extern int bmcg_sat_solver_jftr(bmcg_sat_solver * s)
func bmcg_sat_solver_jftr(s BmcgSatSolver) int { return glucoseSimp(s).Jftr }

// bmcg_sat_solver_set_jftr
// C: extern void bmcg_sat_solver_set_jftr(bmcg_sat_solver * s, int jftr)
func bmcg_sat_solver_set_jftr(s BmcgSatSolver, jftr int) { glucoseSimp(s).Jftr = jftr }

// bmcg_sat_solver_set_var_fanin_lit
// C: extern void bmcg_sat_solver_set_var_fanin_lit(bmcg_sat_solver * s, int var, int lit0, int lit1)
func bmcg_sat_solver_set_var_fanin_lit(s BmcgSatSolver, var_ int, lit0 int, lit1 int) {
	glucoseSimp(s).Sat_solver_set_var_fanin_lit(var_, lit0, lit1)
}

// bmcg_sat_solver_start_new_round
// C: extern void bmcg_sat_solver_start_new_round(bmcg_sat_solver * s)
func bmcg_sat_solver_start_new_round(s BmcgSatSolver) {
	glucoseSimp(s).Sat_solver_start_new_round()
}

// bmcg_sat_solver_mark_cone
// C: extern void bmcg_sat_solver_mark_cone(bmcg_sat_solver * s, int var)
func bmcg_sat_solver_mark_cone(s BmcgSatSolver, var_ int) {
	glucoseSimp(s).Sat_solver_mark_cone(var_)
}

// glucose_print_stats
// C: void glucose_print_stats(SimpSolver& s, abctime clk)
func glucose_print_stats(s *GlucoSimpSolver, clk int64) {
	cpuTime := float64(uint32(clk)) / float64(AbcClockTicksPerSecond)
	memUsed := GlucoMemUsed()
	starts := 0
	if s.Starts > 0 {
		starts = int(s.Conflicts / s.Starts)
	}
	fmt.Printf("c restarts              : %d (%d conflicts on average)\n", int32(s.Starts), int32(starts))
	fmt.Printf("c blocked restarts      : %d (multiple: %d) \n", int32(s.Nbstopsrestarts), int32(s.Nbstopsrestartssame))
	fmt.Printf("c last block at restart : %d\n", int32(s.Lastblockatrestart))
	fmt.Printf("c nb ReduceDB           : %-12d\n", int32(s.NbReduceDB))
	fmt.Printf("c nb removed Clauses    : %-12d\n", int32(s.NbRemovedClauses))
	fmt.Printf("c nb learnts DL2        : %-12d\n", int32(s.NbDL2))
	fmt.Printf("c nb learnts size 2     : %-12d\n", int32(s.NbBin))
	fmt.Printf("c nb learnts size 1     : %-12d\n", int32(s.NbUn))
	fmt.Printf("c conflicts             : %-12d  (%.0f /sec)\n", int32(s.Conflicts), float64(s.Conflicts)/cpuTime)
	fmt.Printf("c decisions             : %-12d  (%4.2f %% random) (%.0f /sec)\n", int32(s.Decisions), float64(float32(s.Rnd_decisions)*100/float32(s.Decisions)), float64(s.Decisions)/cpuTime)
	fmt.Printf("c propagations          : %-12d  (%.0f /sec)\n", int32(s.Propagations), float64(s.Propagations)/cpuTime)
	fmt.Printf("c conflict literals     : %-12d  (%4.2f %% deleted)\n", int32(s.Tot_literals), float64((s.Max_literals-s.Tot_literals)*100)/float64(s.Max_literals))
	fmt.Printf("c nb reduced Clauses    : %-12d\n", int32(s.NbReducedClauses))
	if memUsed != 0 {
		fmt.Printf("Memory used           : %.2f MB\n", memUsed)
	}
}

// GlucoseReadDimacs is Glucose_ReadDimacs: a tolerant DIMACS reader that adds the
// clauses with addClause (the C code reads past the end of the buffer on a file
// without a final newline; the Go port stops at the end of the data).
// C: void Glucose_ReadDimacs( char * pFileName, SimpSolver& s )
func GlucoseReadDimacs(pFileName string, s *GlucoSimpSolver) {
	lits := &s.User_lits
	data, err := os.ReadFile(pFileName) // Extra_FileReadContents: NULL if the file cannot be opened
	if err != nil {
		return
	}
	at := func(i int) byte { // the NUL terminated buffer
		if i < len(data) {
			return data[i]
		}
		return 0
	}
	var fComp, Var, VarMax int
	lits.Clear(false)
	for pTemp := 0; at(pTemp) != 0; pTemp++ {
		if at(pTemp) == 'c' || at(pTemp) == 'p' {
			for at(pTemp) != '\n' && at(pTemp) != 0 {
				pTemp++
			}
			continue
		}
		for at(pTemp) == ' ' || at(pTemp) == '\t' || at(pTemp) == '\r' || at(pTemp) == '\n' {
			pTemp++
		}
		fComp = 0
		if at(pTemp) == '-' {
			fComp = 1
			pTemp++
		}
		if at(pTemp) == '+' {
			pTemp++
		}
		Var = 0 // atoi(pTemp)
		for j := pTemp; at(j) >= '0' && at(j) <= '9'; j++ {
			Var = Var*10 + int(at(j)-'0')
		}
		if Var == 0 {
			if lits.Size() > 0 {
				s.AddVar(VarMax)
				s.AddClause(lits)
				lits.Clear(false)
			}
		} else {
			Var--
			if Var > VarMax {
				VarMax = Var
			}
			lits.PushElem(GlucoToLit(glucoAbcVar2Lit(Var, fComp)))
		}
		for at(pTemp) >= '0' && at(pTemp) <= '9' {
			pTemp++
		}
	}
}

// GlucoseSolveCnf is Glucose_SolveCnf.
// C: extern void Glucose_SolveCnf( char * pFilename, Glucose_Pars * pPars )
func GlucoseSolveCnf(pFilename string, pPars *GlucosePars) {
	clk := AbcClock()
	S := NewGlucoSimpSolver()
	S.Verbosity = pPars.verb
	if pPars.nConfls > 0 {
		S.SetConfBudget(int64(pPars.nConfls))
	} else {
		S.SetConfBudget(-1)
	}
	GlucoseReadDimacs(pFilename, S)
	if pPars.verb != 0 {
		fmt.Printf("c ============================[ Problem Statistics ]=============================\n")
		fmt.Printf("c |                                                                             |\n")
		fmt.Printf("c |  Number of variables:  %12d                                         |\n", S.NVars())
		fmt.Printf("c |  Number of clauses:    %12d                                         |\n", S.NClauses())
	}
	if pPars.pre != 0 {
		S.Eliminate(true)
		fmt.Printf("c Simplication removed %d variables and %d clauses.  ", S.Eliminated_vars, S.Eliminated_clauses)
		glucoAbcPrintTime(1, "Time", AbcClock()-clk)
	}
	var dummy GlucoVec[GlucoLit]
	ret := S.SolveLimited(&dummy, false, false)
	if pPars.verb != 0 {
		glucose_print_stats(S, AbcClock()-clk)
	}
	if ret == GlucoLTrue {
		fmt.Printf("SATISFIABLE")
	} else if ret == GlucoLFalse {
		fmt.Printf("UNSATISFIABLE")
	} else {
		fmt.Printf("INDETERMINATE")
	}
	glucoAbcPrintTime(1, "      Time", AbcClock()-clk)
}
