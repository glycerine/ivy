package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Ssw_ManSetConstrPhases finds PI assignments satisfying the unrolled constraints.
// It returns 0 for satisfiable, 1 for unsatisfiable and -1 for unknown.
func Ssw_ManSetConstrPhases(g *aig.Graph, frames int32, inits **vec.Vec_Int_t) int32 {

	if inits != nil {
		*inits = nil
	}
	unrolled := Ssw_FramesWithConstraints(g, frames)
	data := cnf.Cnf_Derive(unrolled, 0)
	if data == nil {
		panic("ssw: cannot derive constraint CNF")
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
		return 1
	}
	solver.SetLimits(1000000, 0)
	switch solver.Solve() {
	case bsat.Unsatisfiable:
		return 1
	case bsat.Satisfiable:
		if inits != nil {
			values := vec.Vec_IntAlloc(1000)
			for _, obj := range unrolled.Inputs() {
				v, ok := solver.Value(data.VarNums[obj.NodeID()])
				if ok && v {
					vec.Vec_IntPush(values, 1)
				} else {
					vec.Vec_IntPush(values, 0)
				}
			}
			*inits = values
		}
		return 0
	default:
		return -1
	}
}

// Ssw_ManSetConstrPhases_ uses one CNF copied across frames with SAT assumptions.
func Ssw_ManSetConstrPhases_(g *aig.Graph, frames int32, inits **vec.Vec_Int_t) int32 {
	if inits != nil {
		*inits = nil
	}
	if aig.Aig_ManConstrNum(g) <= 0 || frames <= 0 {
		panic("ssw: constraints and frames required")
	}
	data := cnf.Cnf_Derive(g, int32(g.OutputCount()))
	if data == nil {
		panic("ssw: cannot derive constraint CNF")
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, frames, 0)
	if solver == nil {
		return 1
	}
	inputs, outputs := g.Inputs(), g.Outputs()
	regs := int(g.RegisterCount())
	nPis, nPOs := len(inputs)-regs, len(outputs)-regs
	stride := int32(data.Variables)
	assumptions := make([]bsat.Literal, 0, regs+int(frames)*int(aig.Aig_ManConstrNum(g)))
	for _, lo := range inputs[nPis:] {
		assumptions = append(assumptions, bsat.Negative(data.VarNums[lo.NodeID()]))
	}
	for f := int32(0); f < frames; f++ {
		for _, out := range outputs[nPOs-int(aig.Aig_ManConstrNum(g)) : nPOs] {
			assumptions = append(assumptions, bsat.Negative(data.VarNums[out.NodeID()]+stride*f))
		}
	}
	solver.SetLimits(1000000, 0)
	switch solver.Solve(assumptions...) {
	case bsat.Unsatisfiable:
		return 1
	case bsat.Satisfiable:
		if inits != nil {
			values := vec.Vec_IntAlloc(1000)
			for f := int32(0); f < frames; f++ {
				for _, pi := range inputs[:nPis] {
					value, ok := solver.Value(data.VarNums[pi.NodeID()] + stride*f)
					if ok && value {
						vec.Vec_IntPush(values, 1)
					} else {
						vec.Vec_IntPush(values, 0)
					}
				}
			}
			*inits = values
		}
		return 0
	default:
		return -1
	}
}
