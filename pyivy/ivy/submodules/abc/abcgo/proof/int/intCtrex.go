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
	return interManFramesBmc(src, nFrames, 0)
}

func interManFramesBmc(src *aig.Graph, nFrames int32, deadline int64) *aig.Graph {
	if src == nil || src.RegisterCount() <= 0 || nFrames <= 0 {
		panic("counterexample construction requires registers and positive frames")
	}
	regs := int(src.RegisterCount())
	inputs, outputs := src.Inputs(), src.Outputs()
	pis, pos := len(inputs)-regs, len(outputs)-regs
	constraints := int(aig.Aig_ManConstrNum(src))
	if pos-constraints != 1 {
		panic("counterexample construction requires one output")
	}
	dst := aig.NewGraph()
	legal := dst.Const1()
	state := make([]aig.Edge, regs)
	for i := range state {
		state[i] = dst.Const0()
	}
	var copies map[int]aig.Edge
	for frame := int32(0); frame < nFrames; frame++ {
		if deadline != 0 && util.Abc_Clock() >= deadline {
			return nil
		}
		copies = map[int]aig.Edge{0: dst.Const1()}
		for i := 0; i < pis; i++ {
			copies[inputs[i].NodeID()] = dst.NewInput()
		}
		for i, lo := range inputs[pis:] {
			copies[lo.NodeID()] = state[i]
		}
		for _, output := range outputs[1:pos] {
			violation := duplicateTransitionEdge(src, dst, copies, src.Node(output).Fanins[0])
			legal = dst.AND(legal, violation.Not())
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
	bad := duplicateTransitionEdge(src, dst, copies, src.Node(outputs[0]).Fanins[0])
	dst.AddOutput(dst.AND(legal, bad))
	return dst
}

// Inter_ManGetCounterExample returns a frame-major trace when the unrolled
// final-frame property is satisfiable, and nil on UNSAT or resource exhaustion.
func Inter_ManGetCounterExample(src *aig.Graph, nFrames, verbose int32) *util.Abc_Cex_t {
	return interManGetCounterExample(src, nFrames, verbose, 0)
}

func interManGetCounterExample(src *aig.Graph, nFrames, verbose int32, deadline int64) *util.Abc_Cex_t {
	started := time.Now()
	frames := interManFramesBmc(src, nFrames, deadline)
	if frames == nil || deadline != 0 && util.Abc_Clock() >= deadline {
		return nil
	}
	data, err := cnf.EncodeAIG(frames, 1)
	if err != nil {
		return nil
	}
	if deadline != 0 && util.Abc_Clock() >= deadline {
		return nil
	}
	solver := cnf.Cnf_DataWriteIntoSolver(data, 1, 0)
	if solver == nil {
		return nil
	}
	solver.SetLimits(1000000, 0)
	if deadline != 0 {
		solver.SetDeadline(time.Now().Add(time.Duration(deadline - util.Abc_Clock())))
	}
	if solver.Solve() != bsat.Satisfiable {
		return nil
	}
	regs := src.RegisterCount()
	pis := int32(src.InputCount()) - regs
	result := util.Abc_CexAlloc(regs, pis, nFrames)
	result.IFrame = nFrames - 1
	for i, input := range frames.Inputs() {
		if deadline != 0 && util.Abc_Clock() >= deadline {
			return nil
		}
		value, ok := solver.Value(data.VarNums[input.NodeID()])
		if !ok {
			return nil
		}
		if value {
			util.Abc_InfoSetBit(result.PData, regs+int32(i))
		}
	}
	if !interVerifyCounterExample(src, result) {
		if verbose != 0 {
			fmt.Println("Counterexample generation: extracted witness failed verification")
		}
		return nil
	}
	if verbose != 0 {
		fmt.Printf("Counterexample generation: %s\n", time.Since(started))
	}
	return result
}

// interVerifyCounterExample mirrors the source's Saig_ManVerifyCex call and
// additionally checks the constraint outputs supported by this Go path.
func interVerifyCounterExample(src *aig.Graph, cex *util.Abc_Cex_t) bool {
	if src == nil || cex == nil || cex.IFrame < 0 || cex.IPo != 0 || cex.NRegs != src.RegisterCount() {
		return false
	}
	regs := int(src.RegisterCount())
	pis := src.InputCount() - regs
	outputs := src.Outputs()
	pos := len(outputs) - regs
	constraints := int(aig.Aig_ManConstrNum(src))
	neededBits := int64(regs) + (int64(cex.IFrame)+1)*int64(pis)
	if pis < 0 || pos-constraints != 1 || int(cex.NPis) != pis || neededBits != int64(cex.NBits) || int64(len(cex.PData))*32 < neededBits {
		return false
	}
	state := make([]bool, regs)
	for i := int32(0); i <= cex.IFrame; i++ {
		inputs := make([]bool, 0, pis+regs)
		for j := 0; j < pis; j++ {
			inputs = append(inputs, util.Abc_InfoHasBit(cex.PData, int32(regs)+i*int32(pis)+int32(j)))
		}
		inputs = append(inputs, state...)
		values, err := src.Evaluate(inputs)
		if err != nil || len(values) != len(outputs) {
			return false
		}
		for j := 0; j < constraints; j++ {
			if values[1+j] {
				return false
			}
		}
		if i == cex.IFrame {
			return values[0]
		}
		for j := 0; j < regs; j++ {
			state[j] = values[pos+j]
		}
	}
	return false
}
