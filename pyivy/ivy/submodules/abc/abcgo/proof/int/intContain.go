package int

// Containment and latch-frame helpers ported from proof/int/intContain.c.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// checkMiter asks whether the source miter can be true. A proof of UNSAT
// establishes containment/equivalence; budget exhaustion remains undecided.
func checkMiter(a, b *aig.Graph, operation int32) int32 {
	miter := aig.Aig_ManCreateMiter(a, b, operation)
	data, err := cnf.EncodeAIG(miter, 1)
	if err != nil {
		return -1
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
		return 0
	default:
		return -1
	}
}

// Inter_ManCheckContainment reports whether every state in newer is in older.
func Inter_ManCheckContainment(newer, older *aig.Graph) int32 {
	return checkMiter(newer, older, 1)
}

func Inter_ManCheckEquivalence(a, b *aig.Graph) int32 {
	return checkMiter(a, b, 0)
}

// Inter_ManFramesLatches builds an unrolled graph and records each register's
// value from frame zero through the end of the last frame.
func Inter_ManFramesLatches(src *aig.Graph, nFrames int32, mapping *[]aig.Edge) *aig.Graph {
	if src == nil || src.RegisterCount() <= 0 || nFrames < 0 || mapping == nil {
		panic("invalid latch unrolling request")
	}
	regs := int(src.RegisterCount())
	inputs, outputs := src.Inputs(), src.Outputs()
	pis, pos := len(inputs)-regs, len(outputs)-regs
	dst := aig.NewGraph()
	state := make([]aig.Edge, regs)
	*mapping = make([]aig.Edge, 0, (int(nFrames)+1)*regs)
	for i := range state {
		state[i] = dst.NewInput()
		*mapping = append(*mapping, state[i])
	}
	for frame := int32(0); frame < nFrames; frame++ {
		copies := map[int]aig.Edge{0: dst.Const1()}
		for i := 0; i < pis; i++ {
			copies[inputs[i].NodeID()] = dst.NewInput()
		}
		for i, lo := range inputs[pis:] {
			copies[lo.NodeID()] = state[i]
		}
		next := make([]aig.Edge, regs)
		for i, li := range outputs[pos:] {
			next[i] = duplicateTransitionEdge(src, dst, copies, src.Node(li).Fanins[0])
		}
		state = next
		*mapping = append(*mapping, state...)
	}
	return dst
}

// Inter_ManAppendCone appends the single output of old, mapping its inputs to
// the supplied nodes in dst and optionally complementing the output.
func Inter_ManAppendCone(old, dst *aig.Graph, inputs []aig.Edge, complement int32) {
	if old == nil || dst == nil || old.OutputCount() != 1 || old.InputCount() != len(inputs) {
		panic("invalid interpolant cone mapping")
	}
	copies := map[int]aig.Edge{0: dst.Const1()}
	for i, edge := range old.Inputs() {
		copies[edge.NodeID()] = inputs[i]
	}
	output := duplicateTransitionEdge(old, dst, copies, old.Node(old.Outputs()[0]).Fanins[0])
	if complement != 0 {
		output = output.Not()
	}
	dst.AddOutput(output)
}

// Inter_ManCheckInductiveContainment constructs ABC's forward or backward
// chain of interpolant constraints and asks whether their conjunction is UNSAT.
func Inter_ManCheckInductiveContainment(trans, interpolant *aig.Graph, steps, backward int32) int32 {
	if trans == nil || interpolant == nil || steps < 0 {
		panic("invalid inductive containment request")
	}
	regs := int(trans.RegisterCount())
	if regs <= 0 || interpolant.InputCount() != regs {
		panic("interpolant register count mismatch")
	}
	var mapping []aig.Edge
	frames := Inter_ManFramesLatches(trans, steps, &mapping)
	for frame := int32(0); frame <= steps; frame++ {
		complement := int32(0)
		if backward == 0 && frame == steps || backward != 0 && frame == 0 {
			complement = 1
		}
		Inter_ManAppendCone(interpolant, frames, mapping[int(frame)*regs:int(frame+1)*regs], complement)
	}
	data := cnf.Cnf_Derive(frames, 0)
	if data == nil {
		return -1
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
		return 1
	}
	switch solver.Solve() {
	case bsat.Unsatisfiable:
		return 1
	case bsat.Satisfiable:
		return 0
	default:
		return -1
	}
}

// The C implementation deliberately asserts for multi-frame uniqueness: its
// simulation code is unreachable pending the original unresolved question.
func Inter_ManCheckUniqueness(_ *aig.Graph, _ *bsat.Solver, _ *cnf.Cnf_Dat_t, nFrames int32) int32 {
	if nFrames == 1 {
		return 1
	}
	panic("multi-frame uniqueness is not implemented in the original source")
}
