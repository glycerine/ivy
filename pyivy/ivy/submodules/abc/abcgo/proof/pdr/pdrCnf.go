// Go port of pdrCnf.c by Alan Mishchenko, UC Berkeley (2010).
package pdr

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Pdr_ObjSatVar1 returns the SAT variable assigned by the monolithic CNF.
func Pdr_ObjSatVar1(p *Pdr_Man_t, k int32, object aig.Edge) int32 {
	_ = k
	id := object.NodeID()
	if p.pCnf1 == nil || id < 0 || id >= len(p.pCnf1.VarNums) {
		panic("pdr: object missing from monolithic CNF")
	}
	return p.pCnf1.VarNums[id]
}

// Pdr_ObjSatVar2FindOrAdd maintains the object-to-frame and reverse mappings.
// SAT variable zero is reserved; a zero forward entry means not yet allocated.
func Pdr_ObjSatVar2FindOrAdd(p *Pdr_Man_t, k int32, object aig.Edge, isNew *bool) int32 {
	id := object.NodeID()
	if k < 0 || int(k) >= len(p.vVar2Ids) || p.vVar2Ids[k] == nil || id < 0 || id >= len(p.pvId2Vars) {
		panic("pdr: missing frame or object mapping")
	}
	forward := p.pvId2Vars[id]
	if forward == nil {
		forward = vec.Vec_IntAlloc(0)
		p.pvId2Vars[id] = forward
	}
	variable := vec.Vec_IntGetEntry(forward, int(k))
	if variable != 0 {
		return variable
	}
	reverse := p.vVar2Ids[k]
	variable = int32(vec.Vec_IntSize(reverse))
	if variable <= 0 {
		panic("pdr: SAT variable zero must be reserved")
	}
	vec.Vec_IntPush(reverse, int32(id))
	vec.Vec_IntWriteEntry(forward, int(k), variable)
	solver := p.vSolvers[k]
	for solver.NumVars() <= variable {
		solver.NewVariable()
	}
	if k == 0 && saig.Saig_ObjIsLo(p.pAig, object) {
		if !solver.AddClause(bsat.Negative(variable)) || !solver.Compress() {
			panic("pdr: inconsistent initial register")
		}
	}
	*isNew = true
	return variable
}

// Pdr_ObjSatVar2 lazily installs a node's CNF and recursively installs its
// fanin cones. The Go CNF representation maps object IDs to SAT variables, so
// an inverse map is required when translating its clause literals back to AIG
// object IDs (the original C representation used object IDs directly).
func Pdr_ObjSatVar2(p *Pdr_Man_t, k int32, object aig.Edge, level, polarity int32) int32 {
	_, _ = level, polarity // The source's default non-PG path emits both phases.
	if p.pCnf2 == nil {
		panic("pdr: dynamic CNF not initialized")
	}
	inverse := p.cnfVarToObj
	var install func(aig.Edge) int32
	install = func(edge aig.Edge) int32 {
		edge = aig.Aig_Regular(edge)
		var fresh bool
		variable := Pdr_ObjSatVar2FindOrAdd(p, k, edge, &fresh)
		if !fresh || aig.Aig_ObjIsCi(p.pAig, edge) {
			return variable
		}
		solver := p.vSolvers[k]
		id := edge.NodeID()
		if id == 0 {
			if !solver.AddClause(bsat.Positive(variable)) {
				panic("pdr: inconsistent AIG constant")
			}
			return variable
		}
		if id >= len(p.pCnf2.Obj2Clause) || id >= len(p.pCnf2.Obj2Count) {
			panic("pdr: missing dynamic CNF object metadata")
		}
		start, count := p.pCnf2.Obj2Clause[id], p.pCnf2.Obj2Count[id]
		if start < 0 || count <= 0 || int(start+count) > len(p.pCnf2.Clauses) {
			panic("pdr: invalid dynamic CNF object clause range")
		}
		for _, clause := range p.pCnf2.Clauses[start : start+count] {
			if len(clause) == 0 || clause[0]>>1 != p.pCnf2.VarNums[id] {
				panic("pdr: dynamic CNF output literal is missing")
			}
			literals := make([]bsat.Literal, len(clause))
			literals[0] = variable<<1 | clause[0]&1
			for i, literal := range clause[1:] {
				cnfVariable := literal >> 1
				if cnfVariable < 0 || int(cnfVariable) >= len(inverse) || inverse[cnfVariable] < 0 {
					panic("pdr: dynamic CNF literal has no AIG object")
				}
				fanin := aig.Aig_ManObj(p.pAig, inverse[cnfVariable])
				literals[i+1] = install(fanin)<<1 | literal&1
			}
			if !solver.AddClause(literals...) {
				panic("pdr: dynamic CNF is inconsistent")
			}
		}
		return variable
	}
	return install(object)
}

func Pdr_ObjSatVar(p *Pdr_Man_t, k, polarity int32, object aig.Edge) int32 {
	if p.pPars.FMonoCnf != 0 {
		return Pdr_ObjSatVar1(p, k, object)
	}
	return Pdr_ObjSatVar2(p, k, object, 0, polarity)
}

func Pdr_ObjRegNum1(p *Pdr_Man_t, k, variable int32) int32 {
	_ = k
	if variable < 0 {
		panic("pdr: negative SAT variable")
	}
	if int(variable) >= p.pCnf1.Variables {
		return -1
	}
	reg := vec.Vec_IntEntry(p.vVar2Reg, int(variable))
	if reg < 0 || reg >= aig.Aig_ManRegNum(p.pAig) {
		panic("pdr: SAT variable is not a register input")
	}
	return reg
}

func Pdr_ObjRegNum2(p *Pdr_Man_t, k, variable int32) int32 {
	if k < 0 || int(k) >= len(p.vVar2Ids) || p.vVar2Ids[k] == nil || variable <= 0 || int(variable) >= vec.Vec_IntSize(p.vVar2Ids[k]) {
		panic("pdr: SAT variable outside frame mapping")
	}
	id := vec.Vec_IntEntry(p.vVar2Ids[k], int(variable))
	if id == -1 {
		return -1
	}
	object := aig.Aig_ManObj(p.pAig, id)
	if !saig.Saig_ObjIsLi(p.pAig, object) {
		panic("pdr: SAT variable is not a register input")
	}
	return aig.Aig_ObjCioId(p.pAig, object) - saig.Saig_ManPoNum(p.pAig)
}

func Pdr_ObjRegNum(p *Pdr_Man_t, k, variable int32) int32 {
	if p.pPars.FMonoCnf != 0 {
		return Pdr_ObjRegNum1(p, k, variable)
	}
	return Pdr_ObjRegNum2(p, k, variable)
}

func Pdr_ManFreeVar(p *Pdr_Man_t, k int32) int32 {
	if p.pPars.FMonoCnf != 0 {
		return p.vSolvers[k].NumVars()
	}
	reverse := p.vVar2Ids[k]
	vec.Vec_IntPush(reverse, -1)
	return int32(vec.Vec_IntSize(reverse) - 1)
}

func pdrConfigureSolver(solver *bsat.Solver, p *Pdr_Man_t) {
	if p.runtimeClock != nil {
		solver.SetRuntimeDeadline(p.runtimeClock, p.timeToStop)
	} else if p.timeToStop != 0 {
		solver.SetDeadline(time.Now().Add(time.Duration(int64(p.timeToStop) - util.Abc_Clock())))
	} else {
		solver.SetDeadline(time.Time{})
	}
	solver.SetStopFunc(p.pPars.RunId, p.pPars.pFuncStop)
}

func Pdr_ManNewSolver1(solver *bsat.Solver, p *Pdr_Man_t, k, initialize int32) *bsat.Solver {
	if solver == nil {
		panic("pdr: nil SAT solver")
	}
	if p.pCnf1 == nil {
		p.pCnf1 = cnf.Cnf_DeriveWithMan(p.pCnfMan, p.pAig, aig.Aig_ManCoNum(p.pAig))
		if p.pCnf1 == nil {
			return nil
		}
		p.vVar2Reg = vec.Vec_IntStartFull(p.pCnf1.Variables)
		for i := int32(0); i < aig.Aig_ManRegNum(p.pAig); i++ {
			object := aig.Aig_ManLi(p.pAig, i)
			variable := Pdr_ObjSatVar1(p, k, object)
			if variable < 0 {
				panic("pdr: register input missing from CNF")
			}
			vec.Vec_IntWriteEntry(p.vVar2Reg, int(variable), i)
		}
	}
	solver = cnf.Cnf_DataWriteIntoSolverInt(solver, p.pCnf1, 1, initialize)
	if solver != nil {
		pdrConfigureSolver(solver, p)
	}
	return solver
}

func Pdr_ManNewSolver2(solver *bsat.Solver, p *Pdr_Man_t, k, initialize int32) *bsat.Solver {
	_ = initialize // Frame zero is initialized when its register variables are allocated.
	if solver == nil || k < 0 {
		panic("pdr: invalid SAT solver or frame")
	}
	if p.pCnf2 == nil {
		p.pCnf2 = cnf.Cnf_DeriveOtherWithMan(p.pCnfMan, p.pAig, 0)
		if p.pCnf2 == nil {
			return nil
		}
		p.pvId2Vars = make([]*vec.Vec_Int_t, len(p.pAig.Nodes()))
		p.cnfVarToObj = make([]int32, p.pCnf2.Variables)
		for i := range p.cnfVarToObj {
			p.cnfVarToObj[i] = -1
		}
		for id, variable := range p.pCnf2.VarNums {
			if variable >= 0 {
				p.cnfVarToObj[variable] = int32(id)
			}
		}
	}
	for len(p.vVar2Ids) <= int(k) {
		p.vVar2Ids = append(p.vVar2Ids, nil)
	}
	reverse := p.vVar2Ids[k]
	if reverse == nil {
		reverse = vec.Vec_IntAlloc(500)
		p.vVar2Ids[k] = reverse
	}
	for _, id := range vec.Vec_IntArray(reverse) {
		if id < 0 {
			continue
		}
		forward := p.pvId2Vars[id]
		if forward == nil || vec.Vec_IntGetEntry(forward, int(k)) <= 0 {
			panic(fmt.Sprintf("pdr: inconsistent frame mapping for object %d", id))
		}
		vec.Vec_IntWriteEntry(forward, int(k), 0)
	}
	vec.Vec_IntClear(reverse)
	vec.Vec_IntPush(reverse, -1)
	for solver.NumVars() < 500 {
		solver.NewVariable()
	}
	pdrConfigureSolver(solver, p)
	return solver
}

func Pdr_ManNewSolver(solver *bsat.Solver, p *Pdr_Man_t, k, initialize int32) *bsat.Solver {
	if p.pPars.FMonoCnf != 0 {
		return Pdr_ManNewSolver1(solver, p, k, initialize)
	}
	return Pdr_ManNewSolver2(solver, p, k, initialize)
}
