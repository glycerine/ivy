package abc

import (
	"fmt"
	"os"
)

// Port of the SimpSolver based part of src/sat/glucose2/AbcGlucose2.cpp: the
// glucose2_solver_* wrappers, the extern "C" bmcg2_sat_solver_* entry points and the
// CNF-file driver. `#define USE_SIMP_SOLVER 1` is set at the top of the file, so
// only the SimpSolver variant is live. The functions that need the Gia package
// (Glucose_SolverFromAig*, Glucose_GenerateCubes, the quantification and
// equivalence helpers, Glucose_SolveAig) remain stubs in skel_sat_glucose2_AbcGlucose2*.go.

// GLUCOSE_UNSAT, GLUCOSE_SAT and GLUCOSE_UNDEC of AbcGlucose2.h.
const (
	Glucose2Unsat = -1
	Glucose2Sat   = 1
	Glucose2Undec = 0
)

// glucose2Simp is the cast `(Gluco2::SimpSolver*)s` of a bmcg2_sat_solver pointer.
func glucose2Simp(s Bmcg2SatSolver) *Gluco2SimpSolver { return s.(*Gluco2SimpSolver) }

// gluco2AbcVar2Lit is Abc_Var2Lit (misc/util/abc_global.h).
func gluco2AbcVar2Lit(v int, c int) int { return v + v + c }

// gluco2AbcPrintTime is Abc_PrintTime (misc/util/abc_global.h): ABC_PRT(pStr, time).
func gluco2AbcPrintTime(level int, pStr string, t int64) {
	fmt.Printf("%s =", pStr)
	fmt.Printf("%9.2f sec\n", 1.0*float64(t)/float64(AbcClockTicksPerSecond))
}

// glucose2_solver_start
// C: SimpSolver * glucose2_solver_start()
func glucose2_solver_start() *Gluco2SimpSolver {
	S := NewGluco2SimpSolver()
	S.SetIncrementalMode()
	return S
}

// glucose2_solver_stop
// C: void glucose2_solver_stop(Gluco2::SimpSolver* S)
func glucose2_solver_stop(S *Gluco2SimpSolver) {
	S.Destroy()
}

// glucose2_solver_reset
// C: void glucose2_solver_reset(Gluco2::SimpSolver* S)
func glucose2_solver_reset(S *Gluco2SimpSolver) {
	S.Reset()
}

// glucose2_solver_addclause
// C: int glucose2_solver_addclause(Gluco2::SimpSolver* S, int * plits, int nlits)
func glucose2_solver_addclause(S *Gluco2SimpSolver, plits []int, nlits int) int {
	var lits Gluco2Vec[Gluco2Lit]
	for i := 0; i < nlits; i++ {
		// note: Glucose uses the same var->lit conventiaon as ABC
		for plits[i]/2 >= S.NVars() {
			S.NewVar(true, true)
		}
		if plits[i]/2 >= S.NVars() { // NOTE: since we explicitely use new function bmc_add_var
			panic("glucose2_solver_addclause: literal out of range")
		}
		lits.PushElem(Gluco2Lit{x: int32(plits[i])})
	}
	if S.AddClause(&lits) { // returns 0 if the problem is UNSAT
		return 1
	}
	return 0
}

// glucose2_solver_setcallback
// C: void glucose2_solver_setcallback(Gluco2::SimpSolver* S, void * pman, int(*pfunc)(void*, int, int*))
func glucose2_solver_setcallback(S *Gluco2SimpSolver, pman any, pfunc func(any, int, []int) int) {
	S.PCnfMan = pman
	S.PCnfFunc = pfunc
	S.NCallConfl = 1000
}

// glucose2_solver_solve
// C: int glucose2_solver_solve(Gluco2::SimpSolver* S, int * plits, int nlits)
func glucose2_solver_solve(S *Gluco2SimpSolver, plits []int, nlits int) int {
	var lits Gluco2Vec[Gluco2Lit]
	for i := 0; i < nlits; i++ {
		lits.PushElem(Gluco2Lit{x: int32(plits[i])})
	}
	res := S.SolveLimited(&lits, false, false)
	if res == Gluco2LTrue {
		return 1
	} else if res == Gluco2LFalse {
		return -1
	}
	return 0
}

// glucose2_solver_addvar
// C: int glucose2_solver_addvar(Gluco2::SimpSolver* S)
func glucose2_solver_addvar(S *Gluco2SimpSolver) int {
	S.NewVar(true, true)
	return S.NVars() - 1
}

// glucose2_solver_read_cex
// C: int * glucose2_solver_read_cex(Gluco2::SimpSolver* S )
func glucose2_solver_read_cex(S *Gluco2SimpSolver) []int {
	return S.GetCex()
}

// glucose2_solver_read_cex_varvalue
// C: int glucose2_solver_read_cex_varvalue(Gluco2::SimpSolver* S, int ivar)
func glucose2_solver_read_cex_varvalue(S *Gluco2SimpSolver, ivar int) int {
	if *S.Model.At(ivar) == Gluco2LTrue {
		return 1
	}
	return 0
}

// glucose2_solver_setstop
// C: void glucose2_solver_setstop(Gluco2::SimpSolver* S, int * pstop)
func glucose2_solver_setstop(S *Gluco2SimpSolver, pstop *int) {
	S.Pstop = pstop
}

// bmcg2_sat_solver_start
// C: extern bmcg2_sat_solver * bmcg2_sat_solver_start()
func bmcg2_sat_solver_start() Bmcg2SatSolver {
	return glucose2_solver_start()
}

// bmcg2_sat_solver_stop
// C: extern void bmcg2_sat_solver_stop(bmcg2_sat_solver* s)
func bmcg2_sat_solver_stop(s Bmcg2SatSolver) {
	glucose2_solver_stop(glucose2Simp(s))
}

// bmcg2_sat_solver_reset
// C: extern void bmcg2_sat_solver_reset(bmcg2_sat_solver* s)
func bmcg2_sat_solver_reset(s Bmcg2SatSolver) {
	glucose2_solver_reset(glucose2Simp(s))
}

// bmcg2_sat_solver_addclause
// C: extern int bmcg2_sat_solver_addclause(bmcg2_sat_solver* s, int * plits, int nlits)
func bmcg2_sat_solver_addclause(s Bmcg2SatSolver, plits []int, nlits int) int {
	return glucose2_solver_addclause(glucose2Simp(s), plits, nlits)
}

// bmcg2_sat_solver_setcallback
// C: extern void bmcg2_sat_solver_setcallback(bmcg2_sat_solver* s, void * pman, int(*pfunc)(void*, int, int*))
func bmcg2_sat_solver_setcallback(s Bmcg2SatSolver, pman any, pfunc func(any, int, []int) int) {
	glucose2_solver_setcallback(glucose2Simp(s), pman, pfunc)
}

// bmcg2_sat_solver_solve
// C: extern int bmcg2_sat_solver_solve(bmcg2_sat_solver* s, int * plits, int nlits)
func bmcg2_sat_solver_solve(s Bmcg2SatSolver, plits []int, nlits int) int {
	return glucose2_solver_solve(glucose2Simp(s), plits, nlits)
}

// bmcg2_sat_solver_final: *ppArray receives the final conflict as literal codes (a copy of conflict's Lit.x).
// C: extern int bmcg2_sat_solver_final(bmcg2_sat_solver* s, int ** ppArray)
func bmcg2_sat_solver_final(s Bmcg2SatSolver, ppArray *[]int) int {
	c := &glucose2Simp(s).Conflict
	arr := make([]int, c.Size())
	for i := range arr {
		arr[i] = int(c.At(i).x)
	}
	*ppArray = arr
	return c.Size()
}

// bmcg2_sat_solver_addvar
// C: extern int bmcg2_sat_solver_addvar(bmcg2_sat_solver* s)
func bmcg2_sat_solver_addvar(s Bmcg2SatSolver) int {
	return glucose2_solver_addvar(glucose2Simp(s))
}

// bmcg2_sat_solver_set_nvars
// C: extern void bmcg2_sat_solver_set_nvars(bmcg2_sat_solver* s, int nvars)
func bmcg2_sat_solver_set_nvars(s Bmcg2SatSolver, nvars int) {
	for i := bmcg2_sat_solver_varnum(s); i < nvars; i++ {
		bmcg2_sat_solver_addvar(s)
	}
}

// bmcg2_sat_solver_eliminate
// C: extern int bmcg2_sat_solver_eliminate(bmcg2_sat_solver* s, int turn_off_elim)
func bmcg2_sat_solver_eliminate(s Bmcg2SatSolver, turn_off_elim int) int {
	if glucose2Simp(s).Eliminate(turn_off_elim != 0) {
		return 1
	}
	return 0
}

// bmcg2_sat_solver_var_is_elim
// C: extern int bmcg2_sat_solver_var_is_elim(bmcg2_sat_solver* s, int v)
func bmcg2_sat_solver_var_is_elim(s Bmcg2SatSolver, v int) int {
	if glucose2Simp(s).IsEliminated(v) {
		return 1
	}
	return 0
}

// bmcg2_sat_solver_var_set_frozen
// C: extern void bmcg2_sat_solver_var_set_frozen(bmcg2_sat_solver* s, int v, int freeze)
func bmcg2_sat_solver_var_set_frozen(s Bmcg2SatSolver, v int, freeze int) {
	glucose2Simp(s).SetFrozen(v, freeze != 0)
}

// bmcg2_sat_solver_elim_varnum
// C: extern int bmcg2_sat_solver_elim_varnum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_elim_varnum(s Bmcg2SatSolver) int {
	return glucose2Simp(s).Eliminated_vars
}

// bmcg2_sat_solver_read_cex
// C: extern int * bmcg2_sat_solver_read_cex(bmcg2_sat_solver* s)
func bmcg2_sat_solver_read_cex(s Bmcg2SatSolver) []int {
	return glucose2_solver_read_cex(glucose2Simp(s))
}

// bmcg2_sat_solver_read_cex_varvalue
// C: extern int bmcg2_sat_solver_read_cex_varvalue(bmcg2_sat_solver* s, int)
func bmcg2_sat_solver_read_cex_varvalue(s Bmcg2SatSolver, ivar int) int {
	return glucose2_solver_read_cex_varvalue(glucose2Simp(s), ivar)
}

// bmcg2_sat_solver_set_stop (pstop points at a single stop flag)
// C: extern void bmcg2_sat_solver_set_stop(bmcg2_sat_solver* s, int * pstop)
func bmcg2_sat_solver_set_stop(s Bmcg2SatSolver, pstop *int) {
	glucose2_solver_setstop(glucose2Simp(s), pstop)
}

// bmcg2_sat_solver_set_runtime_limit
// C: extern abctime bmcg2_sat_solver_set_runtime_limit(bmcg2_sat_solver* s, abctime Limit)
func bmcg2_sat_solver_set_runtime_limit(s Bmcg2SatSolver, Limit int64) int64 {
	nRuntimeLimit := glucose2Simp(s).NRuntimeLimit
	glucose2Simp(s).NRuntimeLimit = Limit
	return nRuntimeLimit
}

// bmcg2_sat_solver_set_conflict_budget
// C: extern void bmcg2_sat_solver_set_conflict_budget(bmcg2_sat_solver* s, int Limit)
func bmcg2_sat_solver_set_conflict_budget(s Bmcg2SatSolver, Limit int) {
	if Limit > 0 {
		glucose2Simp(s).SetConfBudget(int64(Limit))
	} else {
		glucose2Simp(s).BudgetOff()
	}
}

// bmcg2_sat_solver_varnum
// C: extern int bmcg2_sat_solver_varnum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_varnum(s Bmcg2SatSolver) int { return glucose2Simp(s).NVars() }

// bmcg2_sat_solver_clausenum
// C: extern int bmcg2_sat_solver_clausenum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_clausenum(s Bmcg2SatSolver) int { return glucose2Simp(s).NClauses() }

// bmcg2_sat_solver_learntnum
// C: extern int bmcg2_sat_solver_learntnum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_learntnum(s Bmcg2SatSolver) int { return glucose2Simp(s).NLearnts() }

// bmcg2_sat_solver_conflictnum
// C: extern int bmcg2_sat_solver_conflictnum(bmcg2_sat_solver* s)
func bmcg2_sat_solver_conflictnum(s Bmcg2SatSolver) int { return int(int32(glucose2Simp(s).Conflicts)) }

// bmcg2_sat_solver_minimize_assumptions
// C: extern int bmcg2_sat_solver_minimize_assumptions(bmcg2_sat_solver * s, int * plits, int nlits, int pivot)
func bmcg2_sat_solver_minimize_assumptions(s Bmcg2SatSolver, plits []int, nlits int, pivot int) int {
	array := &glucose2Simp(s).User_vec
	var i, nlitsL, nlitsR, nresL, nresR, status int
	if pivot < 0 || nlits-pivot < 1 {
		panic("bmcg2_sat_solver_minimize_assumptions: precondition")
	}
	if nlits-pivot == 1 {
		// since the problem is UNSAT, we try to solve it without assuming the last literal
		// if the result is UNSAT, the last literal can be dropped; otherwise, it is needed
		status = bmcg2_sat_solver_solve(s, plits, pivot)
		if status != Glucose2Unsat { // return 1 if the problem is not UNSAT
			return 1
		}
		return 0
	}
	nlitsL = (nlits - pivot) / 2
	nlitsR = (nlits - pivot) - nlitsL
	if nlitsL+nlitsR != nlits-pivot {
		panic("bmcg2_sat_solver_minimize_assumptions: split")
	}
	// solve with these assumptions
	status = bmcg2_sat_solver_solve(s, plits, pivot+nlitsL)
	if status == Glucose2Unsat { // these are enough
		return bmcg2_sat_solver_minimize_assumptions(s, plits, pivot+nlitsL, pivot)
	}
	// these are not enough
	// solve for the right lits
	if nlitsR == 1 {
		nresL = 1
	} else {
		nresL = bmcg2_sat_solver_minimize_assumptions(s, plits, nlits, pivot+nlitsL)
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
	status = bmcg2_sat_solver_solve(s, plits, pivot+nresL)
	if status == Glucose2Unsat { // these are enough
		return nresL
	}
	// solve for the left lits
	if nlitsL == 1 {
		nresR = 1
	} else {
		nresR = bmcg2_sat_solver_minimize_assumptions(s, plits, pivot+nresL+nlitsL, pivot+nresL)
	}
	return nresL + nresR
}

// bmcg2_sat_solver_add_and
// C: extern int bmcg2_sat_solver_add_and(bmcg2_sat_solver * s, int iVar, int iVar0, int iVar1, int fCompl0, int fCompl1, int fCompl)
func bmcg2_sat_solver_add_and(s Bmcg2SatSolver, iVar int, iVar0 int, iVar1 int, fCompl0 int, fCompl1 int, fCompl int) int {
	var Lits [3]int
	b2i := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	Lits[0] = gluco2AbcVar2Lit(iVar, b2i(fCompl == 0))
	Lits[1] = gluco2AbcVar2Lit(iVar0, fCompl0)
	if bmcg2_sat_solver_addclause(s, Lits[:], 2) == 0 {
		return 0
	}
	Lits[0] = gluco2AbcVar2Lit(iVar, b2i(fCompl == 0))
	Lits[1] = gluco2AbcVar2Lit(iVar1, fCompl1)
	if bmcg2_sat_solver_addclause(s, Lits[:], 2) == 0 {
		return 0
	}
	Lits[0] = gluco2AbcVar2Lit(iVar, fCompl)
	Lits[1] = gluco2AbcVar2Lit(iVar0, b2i(fCompl0 == 0))
	Lits[2] = gluco2AbcVar2Lit(iVar1, b2i(fCompl1 == 0))
	if bmcg2_sat_solver_addclause(s, Lits[:], 3) == 0 {
		return 0
	}
	return 1
}

// bmcg2_sat_solver_add_xor
// C: int bmcg2_sat_solver_add_xor( bmcg2_sat_solver * pSat, int iVarA, int iVarB, int iVarC, int fCompl )
func bmcg2_sat_solver_add_xor(pSat Bmcg2SatSolver, iVarA int, iVarB int, iVarC int, fCompl int) int {
	var Lits [3]int
	b2i := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	if iVarA < 0 || iVarB < 0 || iVarC < 0 {
		panic("bmcg2_sat_solver_add_xor: negative variable")
	}
	Lits[0] = gluco2AbcVar2Lit(iVarA, b2i(fCompl == 0))
	Lits[1] = gluco2AbcVar2Lit(iVarB, 1)
	Lits[2] = gluco2AbcVar2Lit(iVarC, 1)
	if bmcg2_sat_solver_addclause(pSat, Lits[:], 3) == 0 {
		panic("bmcg2_sat_solver_add_xor: clause 1")
	}
	Lits[0] = gluco2AbcVar2Lit(iVarA, b2i(fCompl == 0))
	Lits[1] = gluco2AbcVar2Lit(iVarB, 0)
	Lits[2] = gluco2AbcVar2Lit(iVarC, 0)
	if bmcg2_sat_solver_addclause(pSat, Lits[:], 3) == 0 {
		panic("bmcg2_sat_solver_add_xor: clause 2")
	}
	Lits[0] = gluco2AbcVar2Lit(iVarA, fCompl)
	Lits[1] = gluco2AbcVar2Lit(iVarB, 1)
	Lits[2] = gluco2AbcVar2Lit(iVarC, 0)
	if bmcg2_sat_solver_addclause(pSat, Lits[:], 3) == 0 {
		panic("bmcg2_sat_solver_add_xor: clause 3")
	}
	Lits[0] = gluco2AbcVar2Lit(iVarA, fCompl)
	Lits[1] = gluco2AbcVar2Lit(iVarB, 0)
	Lits[2] = gluco2AbcVar2Lit(iVarC, 1)
	if bmcg2_sat_solver_addclause(pSat, Lits[:], 3) == 0 {
		panic("bmcg2_sat_solver_add_xor: clause 4")
	}
	return 4
}

// bmcg2_sat_solver_jftr
// C: extern int bmcg2_sat_solver_jftr(bmcg2_sat_solver * s)
func bmcg2_sat_solver_jftr(s Bmcg2SatSolver) int { return glucose2Simp(s).Jftr }

// bmcg2_sat_solver_set_jftr
// C: extern void bmcg2_sat_solver_set_jftr(bmcg2_sat_solver * s, int jftr)
func bmcg2_sat_solver_set_jftr(s Bmcg2SatSolver, jftr int) { glucose2Simp(s).Jftr = jftr }

// bmcg2_sat_solver_set_var_fanin_lit
// C: extern void bmcg2_sat_solver_set_var_fanin_lit(bmcg2_sat_solver * s, int var, int lit0, int lit1)
func bmcg2_sat_solver_set_var_fanin_lit(s Bmcg2SatSolver, var_ int, lit0 int, lit1 int) {
	glucose2Simp(s).Sat_solver_set_var_fanin_lit(var_, lit0, lit1)
}

// bmcg2_sat_solver_start_new_round
// C: extern void bmcg2_sat_solver_start_new_round(bmcg2_sat_solver * s)
func bmcg2_sat_solver_start_new_round(s Bmcg2SatSolver) {
	glucose2Simp(s).Sat_solver_start_new_round()
}

// bmcg2_sat_solver_mark_cone
// C: extern void bmcg2_sat_solver_mark_cone(bmcg2_sat_solver * s, int var)
func bmcg2_sat_solver_mark_cone(s Bmcg2SatSolver, var_ int) {
	glucose2Simp(s).Sat_solver_mark_cone(var_)
}

// glucose2_print_stats
// C: void glucose2_print_stats(SimpSolver& s, abctime clk)
func glucose2_print_stats(s *Gluco2SimpSolver, clk int64) {
	cpuTime := float64(uint32(clk)) / float64(AbcClockTicksPerSecond)
	memUsed := Gluco2MemUsed()
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

// GlucoseReadDimacs_glucose2 is Glucose_ReadDimacs: a tolerant DIMACS reader that adds the
// clauses with addClause (the C code reads past the end of the buffer on a file
// without a final newline; the Go port stops at the end of the data).
// C: void Glucose_ReadDimacs( char * pFileName, SimpSolver& s )
func GlucoseReadDimacs_glucose2(pFileName string, s *Gluco2SimpSolver) {
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
			lits.PushElem(Gluco2ToLit(gluco2AbcVar2Lit(Var, fComp)))
		}
		for at(pTemp) >= '0' && at(pTemp) <= '9' {
			pTemp++
		}
	}
}

// Glucose2SolveCnf is Glucose2_SolveCnf.
// C: extern void Glucose2_SolveCnf( char * pFilename, Glucose_Pars * pPars )
func Glucose2SolveCnf(pFilename string, pPars *Glucose2Pars) {
	clk := AbcClock()
	S := NewGluco2SimpSolver()
	S.Verbosity = pPars.verb
	if pPars.nConfls > 0 {
		S.SetConfBudget(int64(pPars.nConfls))
	} else {
		S.SetConfBudget(-1)
	}
	GlucoseReadDimacs_glucose2(pFilename, S)
	if pPars.verb != 0 {
		fmt.Printf("c ============================[ Problem Statistics ]=============================\n")
		fmt.Printf("c |                                                                             |\n")
		fmt.Printf("c |  Number of variables:  %12d                                         |\n", S.NVars())
		fmt.Printf("c |  Number of clauses:    %12d                                         |\n", S.NClauses())
	}
	if pPars.pre != 0 {
		S.Eliminate(true)
		fmt.Printf("c Simplication removed %d variables and %d clauses.  ", S.Eliminated_vars, S.Eliminated_clauses)
		gluco2AbcPrintTime(1, "Time", AbcClock()-clk)
	}
	var dummy Gluco2Vec[Gluco2Lit]
	ret := S.SolveLimited(&dummy, false, false)
	if pPars.verb != 0 {
		glucose2_print_stats(S, AbcClock()-clk)
	}
	if ret == Gluco2LTrue {
		fmt.Printf("SATISFIABLE")
	} else if ret == Gluco2LFalse {
		fmt.Printf("UNSATISFIABLE")
	} else {
		fmt.Printf("INDETERMINATE")
	}
	gluco2AbcPrintTime(1, "      Time", AbcClock()-clk)
}
