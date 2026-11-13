// Go port of pdrSat.c by Alan Mishchenko, UC Berkeley (2010).
package pdr

import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func Pdr_ManCreateSolver(p *Pdr_Man_t, k int32) *bsat.Solver {
	if int(k) != len(p.vSolvers) || int(k) != vec.Vec_VecSize(p.vClauses) || int(k) != vec.Vec_IntSize(p.vActVars) {
		panic("pdr: inconsistent solver frame count")
	}
	solver := Pdr_ManNewSolver(bsat.NewSolverSeed(float64(p.pPars.nRandomSeed)), p, k, pdrBool(k == 0))
	if solver == nil {
		return nil
	}
	p.vSolvers = append(p.vSolvers, solver)
	vec.Vec_VecExpand(p.vClauses, int(k))
	vec.Vec_IntPush(p.vActVars, 0)
	for i := int32(0); i < saig.Saig_ManPoNum(p.pAig); i++ {
		Pdr_ObjSatVar(p, k, 1, aig.Aig_ManCo(p.pAig, i))
	}
	return solver
}

func Pdr_ManFetchSolver(p *Pdr_Man_t, k int32) *bsat.Solver {
	solver := Pdr_ManSolver(p, k)
	if vec.Vec_IntEntry(p.vActVars, int(k)) < p.pPars.nRecycle {
		return solver
	}
	if int(k) >= len(p.vSolvers)-1 {
		panic("pdr: cannot recycle last frame solver")
	}
	p.nStarts++
	// ABC restarts this frame with the configured random seed. Reset() uses
	// the fixed default seed, so construct the equivalent seeded solver.
	solver = bsat.NewSolverSeed(float64(p.pPars.nRandomSeed))
	solver = Pdr_ManNewSolver(solver, p, k, pdrBool(k == 0))
	if solver == nil {
		return nil
	}
	p.vSolvers[k] = solver
	vec.Vec_IntWriteEntry(p.vActVars, int(k), 0)
	Pdr_ManSetPropertyOutput(p, k)
	for frame := int(k); frame < vec.Vec_VecSize(p.vClauses); frame++ {
		for _, item := range vec.Vec_PtrArray(vec.Vec_VecEntry(p.vClauses, frame)) {
			Pdr_ManSolverAddClause(p, k, item.(*Pdr_Set_t))
		}
	}
	return solver
}

func Pdr_ManLitsToCube(p *Pdr_Man_t, k int32, lits []int32, count int32) *vec.Vec_Int_t {
	if count < 0 || int(count) > len(lits) {
		panic("pdr: invalid literal count")
	}
	vec.Vec_IntClear(p.vLits)
	for _, lit := range lits[:count] {
		reg := Pdr_ObjRegNum(p, k, lit>>1)
		if reg == -1 {
			continue
		}
		if reg < 0 || reg >= aig.Aig_ManRegNum(p.pAig) {
			panic("pdr: SAT variable mapped to invalid register")
		}
		vec.Vec_IntPush(p.vLits, reg<<1|(lit&1)^1)
	}
	return p.vLits
}

func Pdr_ManCubeToLits(p *Pdr_Man_t, k int32, cube *Pdr_Set_t, complement, next int32) *vec.Vec_Int_t {
	start := util.Abc_Clock()
	vec.Vec_IntClear(p.vLits)
	for _, lit := range cube.Lits[:cube.nLits] {
		if lit == -1 {
			continue
		}
		reg := lit >> 1
		if reg < 0 || reg >= aig.Aig_ManRegNum(p.pAig) {
			panic("pdr: cube register out of range")
		}
		object := saig.Saig_ManLo(p.pAig, reg)
		polarity := int32(3)
		if next != 0 {
			object, polarity = saig.Saig_ManLi(p.pAig, reg), 2-(lit&1)
		}
		variable := Pdr_ObjSatVar(p, k, polarity, object)
		if variable < 0 {
			panic("pdr: cube object lacks SAT variable")
		}
		vec.Vec_IntPush(p.vLits, variable<<1|(complement^(lit&1))&1)
	}
	p.tCnf += util.Abc_Clock() - start
	return p.vLits
}

func Pdr_ManSetPropertyOutput(p *Pdr_Man_t, k int32) {
	if p.pPars.fUsePropOut == 0 {
		return
	}
	solver := Pdr_ManSolver(p, k)
	for i := int32(0); i < saig.Saig_ManPoNum(p.pAig); i++ {
		if p.vCexes != nil && vec.Vec_PtrEntry(p.vCexes, int(i)) != nil {
			continue
		}
		if p.pPars.vOutMap != nil && vec.Vec_IntEntry(p.pPars.vOutMap, int(i)) == -1 {
			continue
		}
		variable := Pdr_ObjSatVar(p, k, 1, aig.Aig_ManCo(p.pAig, i))
		if !solver.AddClause(bsat.Negative(variable)) {
			panic("pdr: inconsistent property output")
		}
	}
	solver.Compress()
}

func pdrLiterals(v *vec.Vec_Int_t) []bsat.Literal {
	return vec.Vec_IntArray(v)
}

// Materialize the structural cone before solving. Ternary simulation reads
// values for cone objects, including nodes omitted by CNF cuts; adding their
// SAT variables after Solve would invalidate the model in the Go solver.
func pdrPrepareModel(p *Pdr_Man_t, k int32, cube *Pdr_Set_t) {
	if p.pPars.fMonoCnf != 0 {
		return
	}
	var stack []aig.Edge
	if cube == nil {
		stack = append(stack, aig.Aig_ManCo(p.pAig, p.iOutCur))
	} else {
		for _, lit := range cube.Lits[:cube.nLits] {
			if lit >= 0 {
				stack = append(stack, saig.Saig_ManLi(p.pAig, lit>>1))
			}
		}
	}
	seen := make(map[int]bool)
	for len(stack) != 0 {
		obj := aig.Aig_Regular(stack[len(stack)-1])
		stack = stack[:len(stack)-1]
		if seen[obj.NodeID()] {
			continue
		}
		seen[obj.NodeID()] = true
		Pdr_ObjSatVar(p, k, 3, obj)
		if aig.Aig_ObjIsCo(p.pAig, obj) {
			stack = append(stack, aig.Aig_ObjFanin0(p.pAig, obj))
		} else if aig.Aig_ObjIsNode(p.pAig, obj) {
			stack = append(stack, aig.Aig_ObjFanin0(p.pAig, obj), aig.Aig_ObjFanin1(p.pAig, obj))
		}
	}
}

func Pdr_ManSolverAddClause(p *Pdr_Man_t, k int32, cube *Pdr_Set_t) {
	solver := Pdr_ManSolver(p, k)
	if !solver.AddClause(pdrLiterals(Pdr_ManCubeToLits(p, k, cube, 1, 0))...) {
		panic("pdr: inconsistent frame clause")
	}
	solver.Compress()
}

func Pdr_ManCollectValues(p *Pdr_Man_t, k int32, objects, values *vec.Vec_Int_t) {
	vec.Vec_IntClear(values)
	solver := Pdr_ManSolver(p, k)
	for _, id := range vec.Vec_IntArray(objects) {
		variable := Pdr_ObjSatVar(p, k, 3, aig.Edge(uint32(id)<<1))
		value, ok := solver.Value(variable)
		if !ok {
			panic("pdr: SAT model unavailable")
		}
		vec.Vec_IntPush(values, pdrBool(value))
	}
}

// pdrSolve temporarily applies the current global/per-output time limit and
// the caller's conflict limit, then restores the solver's prior limits.
func pdrSolve(p *Pdr_Man_t, solver *bsat.Solver, assumptions []bsat.Literal, limit int32) bsat.Status {
	oldConflicts, oldPropagations := solver.Limits()
	oldDeadline := solver.Deadline()
	solver.SetLimits(int64(limit), 0)
	deadline := Pdr_ManTimeLimit(p)
	if deadline == 0 {
		solver.SetDeadline(time.Time{})
	} else {
		solver.SetDeadline(time.Now().Add(time.Duration(deadline - util.Abc_Clock())))
	}
	defer func() { solver.SetLimits(oldConflicts, oldPropagations); solver.SetDeadline(oldDeadline) }()
	return solver.Solve(assumptions...)
}

func Pdr_ManCheckCubeCs(p *Pdr_Man_t, k int32, cube *Pdr_Set_t) int32 {
	solver := Pdr_ManFetchSolver(p, k)
	status := pdrSolve(p, solver, pdrLiterals(Pdr_ManCubeToLits(p, k, cube, 0, 0)), 0)
	if status == bsat.Unknown {
		return -1
	}
	return pdrBool(status == bsat.Unsatisfiable)
}

func Pdr_ManCheckCube(p *Pdr_Man_t, k int32, cube *Pdr_Set_t, predecessor **Pdr_Set_t, conflictLimit, tryConf, useLit int32) int32 {
	p.nCalls++
	solver := Pdr_ManFetchSolver(p, k)
	if predecessor != nil {
		pdrPrepareModel(p, k, cube)
	}
	var assumptions []bsat.Literal
	limit := conflictLimit
	if cube == nil {
		variable := Pdr_ObjSatVar(p, k, 2, aig.Aig_ManCo(p.pAig, p.iOutCur))
		assumptions = []bsat.Literal{bsat.Positive(variable)}
	} else {
		var activation bsat.Literal
		if useLit != 0 {
			vec.Vec_IntAddToEntry(p.vActVars, int(k), 1)
			clause := pdrLiterals(Pdr_ManCubeToLits(p, k, cube, 1, 0))
			activation = bsat.Positive(Pdr_ManFreeVar(p, k))
			clause = append(clause, activation)
			if !solver.AddClause(clause...) {
				panic("pdr: activation clause is inconsistent")
			}
			solver.Compress()
		}
		assumptions = pdrLiterals(Pdr_ManCubeToLits(p, k, cube, 0, 1))
		if useLit != 0 {
			assumptions = append(assumptions, activation^1)
		}
		if tryConf != 0 {
			limit = p.pPars.nConfGenLimit
		}
	}
	start := util.Abc_Clock()
	status := pdrSolve(p, solver, assumptions, limit)
	if status == bsat.Unknown {
		if cube != nil && tryConf != 0 && p.pPars.nConfGenLimit != 0 {
			// ABC treats an exhausted generalization budget as SAT, even
			// when the caller requests a predecessor.
			status = bsat.Satisfiable
		} else {
			return -1
		}
	}
	elapsed := util.Abc_Clock() - start
	p.tSat += elapsed
	if status == bsat.Unsatisfiable {
		p.tSatUnsat += elapsed
		p.nCallsU++
		if predecessor != nil {
			*predecessor = nil
		}
		return 1
	}
	p.tSatSat += elapsed
	p.nCallsS++
	if predecessor != nil {
		start = util.Abc_Clock()
		if p.pPars.fNewXSim != 0 {
			*predecessor = Txs3_ManTernarySim(p.pTxs3, k, cube)
		} else {
			*predecessor = Pdr_ManTernarySim(p, k, cube)
		}
		p.tTsim += util.Abc_Clock() - start
		if *predecessor == nil {
			return -1
		}
		p.nXsimLits += (*predecessor).nLits
		p.nXsimRuns++
	}
	return 0
}

func pdrBool(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
