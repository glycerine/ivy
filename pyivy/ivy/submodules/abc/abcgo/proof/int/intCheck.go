package int

// Incremental inductive containment checking from proof/int/intCheck.c.
import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

type Inter_Check_t struct {
	nFramesK, nVars   int32
	pFrames           *aig.Graph
	pCnf              *cnf.Cnf_Dat_t
	pSat              *bsat.Solver
	vOrLits, vAndLits []int32
}

// Inter_ManUnrollFrames creates register inputs followed by each frame's PIs,
// and emits the register inputs at the end of each frame as outputs.
func Inter_ManUnrollFrames(trans *aig.Graph, nFrames int32) *aig.Graph {
	var mapping []aig.Edge
	frames := Inter_ManFramesLatches(trans, nFrames, &mapping)
	for _, edge := range mapping[int(trans.RegisterCount()):] {
		frames.AddOutput(edge)
	}
	return frames
}

func Inter_CheckStart(trans *aig.Graph, nFrames int32) *Inter_Check_t {
	if trans == nil || trans.RegisterCount() <= 0 || nFrames <= 0 {
		panic("invalid induction frame count")
	}
	frames := Inter_ManUnrollFrames(trans, nFrames)
	data := cnf.Cnf_Derive(frames, int32(frames.OutputCount()))
	if data == nil {
		return nil
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
		return nil
	}
	return &Inter_Check_t{nFramesK: nFrames, nVars: int32(data.Variables), pFrames: frames, pCnf: data, pSat: solver}
}

func Inter_CheckStop(p *Inter_Check_t) {
	if p == nil {
		return
	}
	if p.pCnf != nil {
		cnf.Cnf_DataFree(p.pCnf)
	}
	*p = Inter_Check_t{}
}

func Inter_CheckAddOrGate(p *Inter_Check_t, a, b, c int32) {
	if !p.pSat.AddClause(bsat.Negative(a), bsat.Positive(c)) ||
		!p.pSat.AddClause(bsat.Negative(b), bsat.Positive(c)) ||
		!p.pSat.AddClause(bsat.Positive(a), bsat.Positive(b), bsat.Negative(c)) {
		panic("could not add interpolation OR gate")
	}
}

func Inter_CheckAddEqual(p *Inter_Check_t, a, b int32) {
	if !p.pSat.AddClause(bsat.Negative(a), bsat.Positive(b)) ||
		!p.pSat.AddClause(bsat.Positive(a), bsat.Negative(b)) {
		panic("could not add interpolation equality")
	}
}

// Inter_CheckPerform adds one interpolant to the existing k-induction query.
// It returns 1 only when the resulting assumptions are proven UNSAT.
func Inter_CheckPerform(p *Inter_Check_t, data *cnf.Cnf_Dat_t, deadline abctime) int32 {
	if p == nil || p.pSat == nil || data == nil || data.SourceGraph() == nil {
		panic("invalid interpolant CNF")
	}
	graph := data.SourceGraph()
	regs := graph.InputCount()
	if graph.OutputCount() != 1 || p.pFrames.OutputCount() != int(p.nFramesK)*regs {
		panic("interpolant dimensions do not match frames")
	}
	if deadline != 0 {
		p.pSat.SetDeadline(time.Now().Add(time.Duration(deadline - util.Abc_Clock())))
	}
	base := p.nVars
	cnf.Cnf_DataLift(data, base)
	offset := base
	defer func() { cnf.Cnf_DataLift(data, -offset) }()
	for frame := int32(0); frame <= p.nFramesK; frame++ {
		for _, clause := range data.Clauses {
			if !p.pSat.AddClause(clause...) {
				panic("inconsistent interpolation clauses")
			}
		}
		for i, input := range graph.Inputs() {
			var frameNode aig.Edge
			if frame == 0 {
				frameNode = p.pFrames.Inputs()[i]
			} else {
				frameNode = p.pFrames.Outputs()[i+int(frame-1)*regs]
			}
			Inter_CheckAddEqual(p, data.VarNums[input.NodeID()], p.pCnf.VarNums[frameNode.NodeID()])
		}
		output := data.VarNums[graph.Outputs()[0].NodeID()]
		if frame < p.nFramesK {
			if int(frame) == len(p.vOrLits) {
				p.vOrLits = append(p.vOrLits, output)
			} else {
				combined := p.nVars + int32(data.Variables)
				Inter_CheckAddOrGate(p, p.vOrLits[frame], output, combined)
				p.vOrLits[frame] = combined
			}
		} else {
			p.vAndLits = append(p.vAndLits, output)
		}
		step := int32(data.Variables) + 1
		cnf.Cnf_DataLift(data, step)
		offset += step
		p.nVars += step
	}
	assumptions := make([]bsat.Literal, 0, len(p.vOrLits)+len(p.vAndLits))
	for _, v := range p.vOrLits {
		assumptions = append(assumptions, bsat.Positive(v))
	}
	for _, v := range p.vAndLits {
		assumptions = append(assumptions, bsat.Negative(v))
	}
	if p.pSat.Solve(assumptions...) == bsat.Unsatisfiable {
		return 1
	}
	return 0
}
