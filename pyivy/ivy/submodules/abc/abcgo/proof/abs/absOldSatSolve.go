package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Saig_RefManRunSat computes a care set from an assumption UNSAT core.
func Saig_RefManRunSat(p *Saig_RefMan_t, newOrder int32) *util.Abc_Cex_t {
	formula := cnf.Cnf_DeriveSimple(p.pFrames, 0)
	if formula == nil {
		return nil
	}
	if Saig_RefManSetPhases(p, nil, 0) != 0 {
		fmt.Println("Constructed frames are incorrect.")
		return nil
	}
	// ABC transforms each non-output variable by the phase of its AIG node.
	// The phase values are manager-local here rather than stored in the graph.
	flip := make(map[int32]bool)
	for id, n := range p.pFrames.Nodes() {
		if id < len(formula.VarNums) && formula.VarNums[id] >= 0 && n.Kind != aig.Output {
			flip[formula.VarNums[id]] = p.phase[id]
		}
	}
	for _, clause := range formula.Clauses {
		for i, lit := range clause {
			if flip[lit>>1] {
				clause[i] ^= 1
			}
		}
	}
	solver := cnf.Cnf_DataWriteIntoSolver(formula, 1, 0)
	if solver == nil {
		return nil
	}
	solver.SetLimits(1000000, 0)
	if p.nInputs > 0 && solver.Solve() == bsat.Unsatisfiable {
		care := util.Abc_CexDup(p.pCex, p.pCex.NRegs)
		clear(care.PData)
		return care
	}
	varToPI := vec.Vec_IntStartFull(formula.Variables)
	assumptions := vec.Vec_IntAlloc(len(p.pFrames.Inputs()))
	for i, e := range p.pFrames.Inputs() {
		v := formula.VarNums[e.NodeID()]
		vec.Vec_IntPush(assumptions, 2*v+1)
		vec.Vec_IntWriteEntry(varToPI, int(v), int32(i))
	}
	if newOrder != 0 {
		groups := Saig_RefManOrderLiterals(p, varToPI, assumptions)
		vec.Vec_VecSort(groups, 1)
		vec.Vec_IntClear(assumptions)
		for i := 0; i < vec.Vec_VecSize(groups); i++ {
			for _, lit := range vec.Vec_IntArray(vec.Vec_VecEntryInt(groups, i)) {
				vec.Vec_IntPush(assumptions, lit)
			}
		}
	}
	lits := make([]bsat.Literal, vec.Vec_IntSize(assumptions))
	for i, lit := range vec.Vec_IntArray(assumptions) {
		lits[i] = bsat.Literal(lit)
	}
	status := solver.Solve(lits...)
	if status != bsat.Unsatisfiable {
		if status == bsat.Satisfiable {
			fmt.Println("Internal Error!!! The resulting problem is SAT.")
		} else {
			fmt.Println("Internal Error!!! SAT solver timed out.")
		}
		return nil
	}
	core := solver.UnsatCore()
	if len(core) == 0 {
		return nil
	}
	vec.Vec_IntClear(assumptions)
	for _, lit := range core {
		vec.Vec_IntPush(assumptions, int32(lit))
	}
	groups := Saig_RefManOrderLiterals(p, varToPI, assumptions)
	vec.Vec_IntClear(assumptions)
	for i := 0; i < vec.Vec_VecSize(groups); i++ {
		for _, lit := range vec.Vec_IntArray(vec.Vec_VecEntryInt(groups, i)) {
			vec.Vec_IntPush(assumptions, lit)
		}
	}
	care := Saig_RefManCreateCex(p, varToPI, assumptions)
	zero := Saig_RefManSetPhases(p, care, 0)
	one := Saig_RefManSetPhases(p, care, 1)
	if zero != 0 || one != 0 {
		fmt.Println("Reduced CEX verification has failed.")
	}
	return care
}

// Saig_RefManRefineWithSat greedily drops groups of frame assumptions
// belonging to the same original input while preserving unsatisfiability.
func Saig_RefManRefineWithSat(p *Saig_RefMan_t, aigPIs *vec.Vec_Int_t) *vec.Vec_Int_t {
	formula := cnf.Cnf_DeriveSimple(p.pFrames, 0)
	if formula == nil {
		return nil
	}
	solver := cnf.Cnf_DataWriteIntoSolver(formula, 1, 0)
	if solver == nil {
		return nil
	}
	solver.SetLimits(1000000, 0)
	selected := make(map[int32]bool)
	for _, i := range vec.Vec_IntArray(aigPIs) {
		if i < 0 || int(i) >= len(p.pAig.Inputs()) {
			panic("invalid original input")
		}
		selected[i] = true
	}
	varToPI := vec.Vec_IntStartFull(formula.Variables)
	assumptions := vec.Vec_IntAlloc(len(p.pFrames.Inputs()))
	for i, e := range p.pFrames.Inputs() {
		input, frame := vec.Vec_IntEntry(p.vMapPiF2A, 2*i), vec.Vec_IntEntry(p.vMapPiF2A, 2*i+1)
		if !selected[input] {
			continue
		}
		variable := formula.VarNums[e.NodeID()]
		lit := 2 * variable
		if !util.Abc_InfoHasBit(p.pCex.PData, p.pCex.NRegs+p.pCex.NPis*frame+input) {
			lit++
		}
		vec.Vec_IntPush(assumptions, lit)
		vec.Vec_IntWriteEntry(varToPI, int(variable), int32(i))
	}
	groups := Saig_RefManOrderLiterals(p, varToPI, assumptions)
	solve := func(skip int) bsat.Status {
		var lits []bsat.Literal
		for i := 0; i < vec.Vec_VecSize(groups); i++ {
			if i == skip {
				continue
			}
			for _, lit := range vec.Vec_IntArray(vec.Vec_VecEntryInt(groups, i)) {
				lits = append(lits, bsat.Literal(lit))
			}
		}
		return solver.Solve(lits...)
	}
	for i := 0; i < vec.Vec_VecSize(groups); i++ {
		if solve(i) == bsat.Unsatisfiable {
			vec.Vec_IntClear(vec.Vec_VecEntryInt(groups, i))
		}
	}
	result := vec.Vec_IntAlloc(100)
	for i := 0; i < vec.Vec_VecSize(groups); i++ {
		for _, lit := range vec.Vec_IntArray(vec.Vec_VecEntryInt(groups, i)) {
			vec.Vec_IntPush(result, vec.Vec_IntEntry(varToPI, int(lit>>1)))
		}
	}
	return result
}
