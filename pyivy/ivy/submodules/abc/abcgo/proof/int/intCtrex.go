package int

// Counterexample generation ported from ABC's proof/int/intCtrex.c.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Inter_ManFramesBmc unrolls from the all-zero initial state and exposes the
// property output of the last frame. Inputs are ordered frame by frame.
func Inter_ManFramesBmc(src *aig.Graph, nFrames int32) *aig.Graph {
	if src == nil || src.RegisterCount() <= 0 || nFrames <= 0 {
		panic("counterexample construction requires registers and positive frames")
	}
	regs := int(src.RegisterCount())
	inputs, outputs := src.Inputs(), src.Outputs()
	pis, pos := len(inputs)-regs, len(outputs)-regs
	if pos != 1 {
		panic("counterexample construction requires one output")
	}
	dst := aig.NewGraph()
	state := make([]aig.Edge, regs)
	for i := range state {
		state[i] = dst.Const0()
	}
	var copies map[int]aig.Edge
	for frame := int32(0); frame < nFrames; frame++ {
		copies = map[int]aig.Edge{0: dst.Const1()}
		for i := 0; i < pis; i++ {
			copies[inputs[i].NodeID()] = dst.NewInput()
		}
		for i, lo := range inputs[pis:] {
			copies[lo.NodeID()] = state[i]
		}
		if frame+1 == nFrames {
			break
		}
		next := make([]aig.Edge, regs)
		for i, li := range outputs[pos:] {
			next[i] = duplicateTransitionEdge(src, dst, copies, src.Node(li).Fanins[0])
		}
		state = next
	}
	dst.AddOutput(duplicateTransitionEdge(src, dst, copies, src.Node(outputs[0]).Fanins[0]))
	return dst
}

// Inter_ManGetCounterExample returns a frame-major trace when the unrolled
// final-frame property is satisfiable, and nil on UNSAT or resource exhaustion.
func Inter_ManGetCounterExample(src *aig.Graph, nFrames, verbose int32) *util.Abc_Cex_t {
	started := time.Now()
	frames := Inter_ManFramesBmc(src, nFrames)
	data, err := cnf.EncodeAIG(frames, 1)
	if err != nil {
		return nil
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
		return nil
	}
	solver.SetLimits(1000000, 0)
	if solver.Solve() != bsat.Satisfiable {
		return nil
	}
	regs := src.RegisterCount()
	pis := int32(src.InputCount()) - regs
	result := util.Abc_CexAlloc(regs, pis, nFrames)
	result.IFrame = nFrames - 1
	for i, input := range frames.Inputs() {
		value, ok := solver.Value(data.VarNums[input.NodeID()])
		if !ok {
			return nil
		}
		if value {
			util.Abc_InfoSetBit(result.PData, regs+int32(i))
		}
	}
	if verbose != 0 {
		fmt.Printf("Counterexample generation: %s\n", time.Since(started))
	}
	return result
}
