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
