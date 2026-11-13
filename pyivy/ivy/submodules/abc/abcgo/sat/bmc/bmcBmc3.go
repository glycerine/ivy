package bmc

import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Saig_ParBmc_t contains options and accumulated status for scalable BMC.
// It lives beside the native implementation so this algorithm can be tested
// without importing the unrelated C-derived unrolling declarations.
type Saig_ParBmc_t struct {
	nStart         int32
	nFramesMax     int32
	nConfLimit     int32
	nConfLimitJump int32
	nFramesJump    int32
	nTimeOut       int32
	nTimeOutGap    int32
	nTimeOutOne    int32
	nPisAbstract   int32
	fSolveAll      int32
	fStoreCex      int32
	fUseBridge     int32
	fDropSatOuts   int32
	nFfToAddMax    int32
	fSkipRand      int32
	fNoRestarts    int32
	fUseSatoko     int32
	fUseGlucose    int32
	nLearnedStart  int32
	nLearnedDelta  int32
	nLearnedPerce  int32
	fVerbose       int32
	fNotVerbose    int32
	pLogFileName   string
	fSilent        int32
	iFrame         int32
	nFailOuts      int32
	nDropOuts      int32
	timeLastSolved int64
	pFuncOnFail    func(int32, *util.Abc_Cex_t) int32
	RunId          int32
	pFuncStop      func(int32) int32
}

// Saig_ParBmcSetDefaultParams initializes the public BMC options.
func Saig_ParBmcSetDefaultParams(p *Saig_ParBmc_t) {
	if p == nil {
		return
	}
	*p = Saig_ParBmc_t{nLearnedStart: 10000, nLearnedDelta: 2000, nLearnedPerce: 80, iFrame: -1}
}

// bmcSolveDeadline applies the global budget even inside a single SAT call.
func bmcSolveDeadline(global, now time.Time, remaining time.Duration) time.Time {
	if remaining <= 0 {
		return global
	}
	local := now.Add(remaining)
	if global.IsZero() || local.Before(global) {
		return local
	}
	return global
}

// Saig_ManBmcScalable performs bounded sequential checking with a fresh native
// unrolling per depth and an incremental BSAT solver for that depth. It returns
// 0 when a property is disproved, 1 when the requested bound is safe, and -1
// when solving stops early.
func Saig_ManBmcScalable(graph *aig.Graph, params *Saig_ParBmc_t) int32 {
	if graph == nil || params == nil || graph.OutputCount() == int(aig.Aig_ManRegNum(graph)) {
		return -1
	}
	regs := int32(aig.Aig_ManRegNum(graph))
	pis := int32(graph.InputCount()) - regs
	pos := int32(graph.OutputCount()) - regs
	if regs < 0 || regs > int32(graph.InputCount()) || regs > int32(graph.OutputCount()) {
		return -1
	}
	maxFrames := params.nFramesMax
	unbounded := maxFrames <= 0
	if unbounded {
		if regs < 30 {
			maxFrames = int32(1) << uint(regs)
		} else {
			maxFrames = 10000
		}
		if maxFrames < 1 {
			maxFrames = 1
		}
	}
	start := time.Now()
	var deadline time.Time
	if params.nTimeOut > 0 {
		deadline = start.Add(time.Duration(params.nTimeOut) * time.Second)
	}
	var found bool
	failed := make([]bool, pos)
	spent := make([]time.Duration, pos)
	for frame := int32(0); frame < maxFrames; frame++ {
		if frame < params.nStart {
			params.iFrame = frame
			continue
		}
		if params.nTimeOut > 0 && time.Since(start) >= time.Duration(params.nTimeOut)*time.Second {
			return -1
		}
		if params.pFuncStop != nil && params.pFuncStop(params.RunId) != 0 {
			return -1
		}
		var frameMap []aig.Edge
		unrolled := aig.Aig_ManFrames(graph, frame+1, 1, 1, 0, 0, &frameMap)
		data, err := cnf.EncodeAIG(unrolled, 0)
		if err != nil {
			return -1
		}
		solver := bsat.NewSolver()
		for solver.NumVars() < int32(data.Variables) {
			solver.NewVariable()
		}
		baseUnsat := false
		for _, clause := range data.Clauses {
			if !solver.AddClause(clause...) {
				baseUnsat = true
				break
			}
		}
		if baseUnsat {
			continue
		}
		for output := int32(0); output < pos; output++ {
			if params.fSolveAll != 0 && failed[output] {
				continue
			}
			solveStart := time.Now()
			solveDeadline := deadline
			if params.fSolveAll != 0 && params.nTimeOutOne > 0 {
				remaining := time.Duration(params.nTimeOutOne)*time.Millisecond - spent[output]
				if remaining <= 0 {
					return -1
				}
				solveDeadline = bmcSolveDeadline(deadline, solveStart, remaining)
			}
			solver.SetDeadline(solveDeadline)
			solver.SetLimits(int64(params.nConfLimit), 0)
			coIndex := int(frame)*int(pos) + int(output)
			if coIndex < 0 || coIndex >= len(unrolled.Outputs()) {
				return -1
			}
			co := unrolled.Outputs()[coIndex]
			variable := data.VarNums[co.NodeID()]
			if variable < 0 {
				return -1
			}
			status := solver.Solve(bsat.Positive(variable))
			spent[output] += time.Since(solveStart)
			params.iFrame = frame
			switch status {
			case bsat.Satisfiable:
				found = true
				failed[output] = true
				params.nFailOuts++
				model := makeBmcCounterexample(graph, unrolled, frameMap, data.VarNums, solver, frame, output, pis)
				if model != nil {
					_ = graph.SetSequentialCounterexample(model)
				}
				if params.pFuncOnFail != nil && params.pFuncOnFail(output, model) != 0 {
					return 0
				}
				if params.fSolveAll == 0 {
					return 0
				}
			case bsat.Unsatisfiable:
			case bsat.Unknown:
				return -1
			}
		}
		if params.fSolveAll != 0 && found && params.nFailOuts >= pos {
			return 0
		}
	}
	if found {
		return 0
	}
	if unbounded && regs >= 30 {
		return -1
	}
	return 1
}

func makeBmcCounterexample(graph, unrolled *aig.Graph, mapping []aig.Edge, varNums []int32, solver *bsat.Solver, frame, output int32, nPis int32) *util.Abc_Cex_t {
	if graph == nil || unrolled == nil || solver == nil {
		return nil
	}
	regs := int32(aig.Aig_ManRegNum(graph))
	model := &util.Abc_Cex_t{NRegs: regs, NPis: nPis, IFrame: frame, IPo: output, NBits: regs + (frame+1)*nPis}
	model.PData = make([]uint32, (int(model.NBits)+31)/32)
	for f := int32(0); f <= frame; f++ {
		for i, input := range graph.Inputs()[:int(nPis)] {
			mapped := aig.Aig_ObjFrames(mapping, frame+1, input, f)
			id := mapped.NodeID()
			if id < 0 || id >= len(varNums) {
				return nil
			}
			variable := varNums[id]
			if variable < 0 {
				return nil
			}
			value, known := solver.Value(variable)
			if !known || !value {
				continue
			}
			bit := int(regs + f*nPis + int32(i))
			model.PData[bit>>5] |= 1 << uint(bit&31)
		}
	}
	return model
}
