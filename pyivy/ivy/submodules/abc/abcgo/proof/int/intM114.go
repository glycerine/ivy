package int

// The ordinary ABC interpolation path from proof/int/intM114.c uses the
// internal BSAT proof store, not the optional MiniSat-1.14p interface.
import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Inter_ManDeriveSatSolver assembles the A (interpolant and transition) and B
// (unrolled frames) clause partitions and records their shared variables.
func Inter_ManDeriveSatSolver(inter *aig.Graph, interCNF *cnf.Cnf_Dat_t, trans *aig.Graph, transCNF *cnf.Cnf_Dat_t, frames *aig.Graph, framesCNF *cnf.Cnf_Dat_t, globals *vec.Vec_Int_t, backward int32) *bsat.Solver {
	if inter == nil || interCNF == nil || trans == nil || transCNF == nil || frames == nil || framesCNF == nil || globals == nil || trans.RegisterCount() <= 0 {
		panic("invalid interpolation SAT construction")
	}
	regs := int(trans.RegisterCount())
	if inter.RegisterCount() != 0 || frames.RegisterCount() != 0 || inter.OutputCount() != 1 || backward == 0 && inter.InputCount() != regs {
		panic("incompatible interpolation graphs")
	}
	frameVars, transVars := int32(framesCNF.Variables), int32(transCNF.Variables)
	cnf.Cnf_DataLift(transCNF, frameVars)
	cnf.Cnf_DataLift(interCNF, frameVars+transVars)
	defer cnf.Cnf_DataLift(transCNF, -frameVars)
	defer cnf.Cnf_DataLift(interCNF, -frameVars-transVars)
	solver := bsat.NewSolver()
	solver.EnableClauseStore()
	for n := framesCNF.Variables + transCNF.Variables + interCNF.Variables; int(solver.NumVars()) < n; {
		solver.NewVariable()
	}
	addCNF := func(data *cnf.Cnf_Dat_t) bool {
		for _, clause := range data.Clauses {
			if !solver.AddClause(clause...) {
				return false
			}
		}
		return true
	}
	equal := func(a, b int32) bool {
		if a < 0 || b < 0 {
			panic("interpolation variable missing from CNF")
		}
		return solver.AddClause(bsat.Positive(a), bsat.Negative(b)) && solver.AddClause(bsat.Negative(a), bsat.Positive(b))
	}
	if !addCNF(interCNF) {
		return nil
	}
	transInputs, transOutputs := trans.Inputs(), trans.Outputs()
	transPI, transPO := len(transInputs)-regs, len(transOutputs)-regs
	interInputs := inter.Inputs()
	for i := 0; i < regs; i++ {
		interIndex := i
		transNode := transInputs[transPI+i]
		if backward != 0 {
			transNode = transOutputs[transPO+i]
			if len(interInputs) != regs {
				interIndex = len(interInputs) - regs + i
			}
		}
		if interIndex < 0 || interIndex >= len(interInputs) || !equal(interCNF.VarNums[interInputs[interIndex].NodeID()], transCNF.VarNums[transNode.NodeID()]) {
			return nil
		}
	}
	if !addCNF(transCNF) {
		return nil
	}
	vec.Vec_IntClear(globals)
	frameInputs, frameOutputs := frames.Inputs(), frames.Outputs()
	for i := 0; i < regs; i++ {
		var frameNode, transNode aig.Edge
		if backward != 0 {
			if i >= len(frameOutputs) {
				panic("missing backward frame output")
			}
			frameNode, transNode = frameOutputs[i], transInputs[transPI+i]
		} else {
			if i >= len(frameInputs) {
				panic("missing frame register input")
			}
			frameNode, transNode = frameInputs[i], transOutputs[transPO+i]
		}
		v := framesCNF.VarNums[frameNode.NodeID()]
		vec.Vec_IntPush(globals, v)
		if !equal(v, transCNF.VarNums[transNode.NodeID()]) {
			return nil
		}
	}
	solver.MarkClauseStorePartitionA()
	addCNF(framesCNF) // A contradiction in B is recorded as a solved UNSAT solver.
	solver.MarkClauseStoreRoots()
	return solver
}

// Inter_ManPerformOneStep runs BSAT and builds an interpolant from its proof.
// It returns 1 for a proof, 0 for a counterexample, and -1 for unknown.
func Inter_ManPerformOneStep(p *Inter_Man_t, bias, backward int32, deadline abctime) int32 {
	solver := Inter_ManDeriveSatSolver(p.pInter, p.pCnfInter, p.pAigTrans, p.pCnfAig, p.pFrames, p.pCnfFrames, p.vVarsAB, backward)
	if solver == nil {
		p.pInterNew = nil
		return 1
	}
	if deadline != 0 {
		solver.SetDeadline(time.Now().Add(time.Duration(deadline - util.Abc_Clock())))
	}
	if bias != 0 {
		solver.SetGlobalVariableBias(vec.Vec_IntArray(p.vVarsAB))
	}
	start := util.Abc_Clock()
	status := solver.SolveWithSourceLimits(nil, int64(p.nConfLimit), 0, 0, 0)
	p.nConfCur = int32(solver.ConflictCount())
	p.timeSat += util.Abc_Clock() - start
	if status == bsat.Satisfiable {
		return 0
	}
	if status != bsat.Unsatisfiable {
		return -1
	}
	store := solver.ReleaseClauseStore()
	start = util.Abc_Clock()
	manager := bsat.Inta_ManAlloc()
	p.pInterNew = bsat.Inta_ManInterpolate(manager, store, deadline, vec.Vec_IntArray(p.vVarsAB), 0)
	bsat.Inta_ManFree(manager)
	p.timeInt += util.Abc_Clock() - start
	if p.pInterNew == nil {
		return -1
	}
	return 1
}
