package bmc

import (
	"fmt"
	"io"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
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
	pBridgeWriter  io.Writer // nil selects stdout, matching ABC's bridge stream
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

// Saig_ManBmcScalable performs bounded sequential checking with one growing
// frame unrolling and one persistent BSAT solver. It returns 0 on a witness,
// 1 only after the finite-state completeness bound, and -1 on bounded
// exhaustion or interrupted/unknown solving, matching the C entry point.
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
	completeAt := int32(-1)
	if unbounded {
		params.nFramesMax = 1000000000 // ABC_INFINITY
		maxFrames = params.nFramesMax
	}
	if regs < 30 {
		completeAt = int32(1) << uint(regs)
	}
	if maxFrames < 1 {
		maxFrames = 1
	}
	start := time.Now()
	if params.nTimeOutOne > 0 && params.nTimeOut == 0 {
		params.nTimeOut = params.nTimeOutOne*pos/1000 + 1
	}
	if params.nTimeOutOne > 0 && params.fSolveAll == 0 {
		params.nTimeOutOne = 0
	}
	params.timeLastSolved = start.UnixNano()
	var deadline time.Time
	if params.nTimeOut > 0 {
		deadline = start.Add(time.Duration(params.nTimeOut) * time.Second)
	}
	var found bool
	failed := make([]bool, pos)
	spent := make([]time.Duration, pos)
	var storedCEX *vec.Vec_Ptr_t
	if params.fSolveAll != 0 && params.fStoreCex != 0 {
		storedCEX = vec.Vec_PtrStart(int(pos))
		defer graph.SetSequentialCounterexamples(storedCEX)
	}
	var jumpFrame int32
	cutManager, err := newBmc3CutManager(graph, regs, pis, pos)
	if err != nil {
		return -1
	}
	solver := newBmc3Solver(params)
	// C initializes each BMC3 backend with 1000 SAT variables; variable 0 is
	// reserved, and the cut manager's literals keep their native variable IDs.
	for solver.NumVars() < 1000 {
		solver.NewVariable()
	}
	syncCutCNF := func() bool {
		for solver.NumVars() <= cutManager.variableCount() {
			solver.NewVariable()
		}
		for _, clause := range cutManager.drainClauses() {
			lits := make([]bsat.Literal, len(clause))
			for i, cLit := range clause {
				lit, err := bmcCToBSATLiteral(cLit)
				if err != nil {
					return false
				}
				lits[i] = bsat.Literal(lit)
			}
			if !solver.AddClause(lits...) {
				return false
			}
		}
		return true
	}
	for frame := int32(0); frame < maxFrames; frame++ {
		// C checks completeness before adding the frame at 2^nRegs.
		if completeAt >= 0 && params.nFramesJump == 0 && frame == completeAt {
			if params.fUseBridge != 0 {
				for output := int32(0); output < pos; output++ {
					if !failed[output] && spent[output] >= 0 {
						_ = writeBmcBridgeResult(bmcBridgeWriter(params.pBridgeWriter), 1, nil, output)
					}
				}
			}
			if params.nFailOuts > 0 {
				return 0
			}
			return 1
		}
		if params.fSolveAll != 0 && params.nFailOuts+params.nDropOuts >= pos {
			if params.nFailOuts > 0 {
				return 0
			}
			return 1
		}
		if frame < params.nStart || frame < jumpFrame {
			params.iFrame = frame
			continue
		}
		targets := make([]int32, pos)
		if params.fSolveAll != 0 {
			for output := int32(0); output < pos; output++ {
				if failed[output] || spent[output] < 0 {
					continue
				}
				target, ok := cutManager.outputLiteral(int(frame), output)
				if !ok {
					return -1
				}
				targets[output] = target
			}
			if !syncCutCNF() {
				return 1
			}
		}
		for output := int32(0); output < pos; output++ {
			if params.fSolveAll != 0 && (failed[output] || spent[output] < 0) {
				continue
			}
			if params.nTimeOutGap > 0 && time.Since(time.Unix(0, params.timeLastSolved)) > time.Duration(params.nTimeOutGap)*time.Second {
				return resultFromFailures(params)
			}
			if params.nTimeOut > 0 && time.Since(start) >= time.Duration(params.nTimeOut)*time.Second {
				return resultFromFailures(params)
			}
			if params.pFuncStop != nil && params.pFuncStop(params.RunId) != 0 {
				return resultFromFailures(params)
			}
			if params.fSolveAll == 0 {
				target, ok := cutManager.outputLiteral(int(frame), output)
				if !ok {
					return -1
				}
				targets[output] = target
				if !syncCutCNF() {
					return 1
				}
			}
			solveStart := time.Now()
			solveDeadline := deadline
			if params.nTimeOutGap > 0 {
				gapDeadline := time.Unix(0, params.timeLastSolved).Add(time.Duration(params.nTimeOutGap) * time.Second)
				solveDeadline = bmcSolveDeadline(solveDeadline, solveStart, time.Until(gapDeadline))
			}
			if params.fSolveAll != 0 && params.nTimeOutOne > 0 {
				remaining := time.Duration(params.nTimeOutOne)*time.Millisecond - spent[output]
				if remaining <= 0 {
					if !failed[output] && spent[output] >= 0 {
						params.nDropOuts++
						spent[output] = -1
					}
					continue
				}
				solveDeadline = bmcSolveDeadline(solveDeadline, solveStart, remaining)
			}
			solver.SetDeadline(solveDeadline)
			solver.SetLimits(int64(params.nConfLimit), 0)
			variable := targets[output]
			var status bsat.Status
			switch variable {
			case 0:
				status = bsat.Unsatisfiable
			case 1:
				status = bsat.Satisfiable
			default:
				status = solver.Solve(bmcCToBSATAssumption(variable))
			}
			spent[output] += time.Since(solveStart)
			params.iFrame = frame
			jumped := false
			switch status {
			case bsat.Satisfiable:
				found = true
				failed[output] = true
				params.nFailOuts++
				model := cutManager.counterexample(solver, frame, output)
				if params.fSolveAll != 0 && params.fUseBridge != 0 && model != nil {
					_ = writeBmcBridgeResult(bmcBridgeWriter(params.pBridgeWriter), 0, model, output)
				}
				callbackModel := model
				if params.fSolveAll != 0 {
					if storedCEX != nil && model != nil {
						vec.Vec_PtrWriteEntry(storedCEX, int(output), model)
					} else {
						callbackModel = nil
					}
				} else if model != nil {
					_ = graph.SetSequentialCounterexample(model)
				}
				if params.fSolveAll != 0 && params.pFuncOnFail != nil && params.pFuncOnFail(output, callbackModel) != 0 {
					return 0
				}
				params.timeLastSolved = time.Now().UnixNano()
				if params.fSolveAll == 0 {
					return 0
				}
				// The source tests all remaining outputs against this same SAT
				// assignment before advancing the frame.
				for other := int32(0); other < pos; other++ {
					if failed[other] {
						continue
					}
					otherLiteral := targets[other]
					otherSat := otherLiteral == 1
					if otherLiteral >= 2 {
						value, known := solver.Value(otherLiteral >> 1)
						otherSat = known && value != (otherLiteral&1 != 0)
					}
					if !otherSat {
						continue
					}
					failed[other] = true
					params.nFailOuts++
					otherModel := model
					if otherModel != nil {
						copy := *otherModel
						copy.PData = append([]uint32(nil), otherModel.PData...)
						copy.IPo = other
						otherModel = &copy
					}
					if params.fUseBridge != 0 && otherModel != nil {
						_ = writeBmcBridgeResult(bmcBridgeWriter(params.pBridgeWriter), 0, otherModel, other)
					}
					if storedCEX != nil && otherModel != nil {
						vec.Vec_PtrWriteEntry(storedCEX, int(other), otherModel)
					}
				}
			case bsat.Unsatisfiable:
				// C retains the failed target as a permanent unit so subsequent
				// incremental solves need not rediscover this output/frame result.
				if variable >= 2 {
					unit, err := bmcCToBSATLiteral(variable ^ 1)
					if err != nil {
						return -1
					}
					if !solver.AddClause(bsat.Literal(unit)) {
						return 1
					}
					if defaultSolver, ok := solver.(*bsat.Solver); ok && !transferBmcLearnedUnits(defaultSolver) {
						return 1
					}
				}
				if params.fUseBridge != 0 {
					_ = writeBmcBridgeProgress(bmcBridgeWriter(params.pBridgeWriter), output, frame)
				}
			case bsat.Unknown:
				if params.nFramesJump > 0 {
					params.nConfLimit = params.nConfLimitJump
					jumpFrame = frame + params.nFramesJump
					jumped = true
				} else if params.fSolveAll != 0 && params.nTimeOutOne > 0 {
					if spent[output] >= time.Duration(params.nTimeOutOne)*time.Millisecond {
						params.nDropOuts++
						spent[output] = -1
					}
				} else {
					return -1
				}
			}
			if jumped {
				break
			}
		}
		if params.fSolveAll != 0 && found && params.nFailOuts >= pos {
			return 0
		}
		if params.nFramesJump > 0 && jumpFrame > frame {
			continue
		}
	}
	if found {
		return 0
	}
	return -1
}

func resultFromFailures(params *Saig_ParBmc_t) int32 {
	if params != nil && params.nFailOuts > 0 {
		return 0
	}
	return -1
}

func bmcCToBSATAssumption(cLit int32) bsat.Literal {
	variable := cLit >> 1
	if cLit&1 != 0 {
		return bsat.Negative(variable)
	}
	return bsat.Positive(variable)
}

// bmcIncrementalUnroller appends a single frame at a time. Node IDs and CNF
// variables for earlier frames remain stable, allowing clauses and learned
// solver state to be reused as in the source scalable-BMC manager.
type bmcIncrementalUnroller struct {
	src            *aig.Graph
	graph          *aig.Graph
	regs, pis, pos int32
	mapping        []aig.Edge // frame-major source object map
	outputs        []aig.Edge
	frame          int32
}

func newBmcIncrementalUnroller(src *aig.Graph, regs, pis, pos int32) *bmcIncrementalUnroller {
	return &bmcIncrementalUnroller{src: src, graph: aig.NewGraph(), regs: regs, pis: pis, pos: pos}
}

func (u *bmcIncrementalUnroller) addFrame() error {
	nodes := u.src.Nodes()
	inputs, outputs := u.src.Inputs(), u.src.Outputs()
	if len(nodes) == 0 || len(inputs) < int(u.regs) || len(outputs) < int(u.regs) {
		return fmt.Errorf("invalid sequential graph")
	}
	old := len(u.mapping)
	u.mapping = append(u.mapping, make([]aig.Edge, len(nodes))...)
	for i := old; i < len(u.mapping); i++ {
		u.mapping[i] = aig.InvalidEdge
	}
	base := int(u.frame) * len(nodes)
	for _, e := range inputs[:int(u.pis)] {
		u.mapping[base+e.NodeID()] = u.graph.NewInput()
	}
	for i, e := range inputs[int(u.pis):] {
		v := u.graph.Const0()
		if u.frame > 0 {
			li := outputs[int(u.pos)+i]
			n := nodes[li.NodeID()]
			v = u.mapping[int(u.frame-1)*len(nodes)+n.Fanins[0].NodeID()]
			if n.Fanins[0].Complemented() {
				v = v.Not()
			}
		}
		u.mapping[base+e.NodeID()] = v
	}
	for id, n := range nodes {
		var v aig.Edge
		switch n.Kind {
		case aig.Constant:
			u.mapping[base+id] = u.graph.Const1()
			continue
		case aig.Input:
			if u.mapping[base+id] == aig.InvalidEdge {
				return fmt.Errorf("unmapped input %d", id)
			}
			continue
		case aig.And, aig.ExclusiveOr:
			a, b := u.mappedChild(base, n.Fanins[0]), u.mappedChild(base, n.Fanins[1])
			if n.Kind == aig.And {
				v = u.graph.AND(a, b)
			} else {
				v = u.graph.NewXor(a, b)
			}
		case aig.Buffer, aig.Output:
			v = u.mappedChild(base, n.Fanins[0])
		default:
			continue
		}
		u.mapping[base+id] = v
	}
	for _, e := range outputs[:int(u.pos)] {
		driver := nodes[e.NodeID()].Fanins[0]
		v := u.mapping[base+driver.NodeID()]
		if driver.Complemented() {
			v = v.Not()
		}
		u.outputs = append(u.outputs, u.graph.AddOutput(v))
	}
	u.frame++
	return nil
}

func (u *bmcIncrementalUnroller) mappedChild(base int, edge aig.Edge) aig.Edge {
	v := u.mapping[base+edge.NodeID()]
	if edge.Complemented() {
		v = v.Not()
	}
	return v
}

func makeBmcCounterexample(graph, unrolled *aig.Graph, mapping []aig.Edge, varNums []int32, solver bmc3Solver, frame, output int32, nPis int32) *util.Abc_Cex_t {
	if graph == nil || unrolled == nil || solver == nil {
		return nil
	}
	regs := int32(aig.Aig_ManRegNum(graph))
	model := &util.Abc_Cex_t{NRegs: regs, NPis: nPis, IFrame: frame, IPo: output, NBits: regs + (frame+1)*nPis}
	model.PData = make([]uint32, (int(model.NBits)+31)/32)
	for f := int32(0); f <= frame; f++ {
		for i, input := range graph.Inputs()[:int(nPis)] {
			idx := int(f)*len(graph.Nodes()) + input.NodeID()
			if idx < 0 || idx >= len(mapping) {
				return nil
			}
			mapped := mapping[idx]
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
