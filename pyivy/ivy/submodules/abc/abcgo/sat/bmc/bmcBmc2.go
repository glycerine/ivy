package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

const (
	ABS_ZER int32 = absZero
	ABS_ONE int32 = absOne
	ABS_UND int32 = absUnknown
)

func Abs_ManSimInfoNot(value int32) int32          { return ternaryNot(value) }
func Abs_ManSimInfoAnd(value0, value1 int32) int32 { return ternaryAnd(value0, value1) }

// Abs_ManSimInfoGet reads an object's packed ternary value for one frame.
func Abs_ManSimInfoGet(frames *vec.Vec_Ptr_t, object *aig.Edge, frame int32) int32 {
	if frames == nil || object == nil || frame < 0 || int(frame) >= vec.Vec_PtrSize(frames) {
		return 0
	}
	words, ok := vec.Vec_PtrEntry(frames, int(frame)).([]uint32)
	if !ok || words == nil {
		return 0
	}
	id := uint32(object.NodeID())
	word := int(id >> 4)
	if word >= len(words) {
		return 0
	}
	return int32(packed2Get(words[word], id))
}

// Abs_ManSimInfoSet replaces an object's packed ternary value for one frame.
func Abs_ManSimInfoSet(frames *vec.Vec_Ptr_t, object *aig.Edge, frame, value int32) {
	if frames == nil || object == nil || frame < 0 || int(frame) >= vec.Vec_PtrSize(frames) {
		return
	}
	if value < ABS_ZER || value > ABS_UND {
		panic("bmc: invalid ternary state")
	}
	words, ok := vec.Vec_PtrEntry(frames, int(frame)).([]uint32)
	if !ok || words == nil {
		panic("bmc: simulation frame does not contain packed words")
	}
	id := uint32(object.NodeID())
	word := int(id >> 4)
	if word >= len(words) {
		panic("bmc: object outside simulation frame")
	}
	words[word] = packed2Set(words[word], id, uint32(value))
}

// Abs_ManTernarySimulate propagates the zero initial state and unknown primary
// inputs through the requested frames. It stops once a property is not proved
// false, matching the source routine's early-return contract.
func Abs_ManTernarySimulate(graph *aig.Graph, maxFrames, verbose int32) *vec.Vec_Ptr_t {
	result := vec.Vec_PtrAlloc(0)
	if graph == nil || maxFrames <= 0 || aig.Aig_ManRegNum(graph) <= 0 {
		return result
	}
	inputs, outputs, nodes := graph.Inputs(), graph.Outputs(), graph.Nodes()
	regs := int(aig.Aig_ManRegNum(graph))
	if regs > len(inputs) || regs > len(outputs) {
		return result
	}
	primaryInputs := len(inputs) - regs
	primaryOutputs := len(outputs) - regs
	state := make([]int32, regs) // C's default initial state is all zero.
	for i := range state {
		state[i] = ABS_ZER
	}
	for frame := int32(0); frame < maxFrames; frame++ {
		values := make([]int32, len(nodes))
		values[0] = ABS_ONE
		for i, input := range inputs {
			if i < primaryInputs {
				values[input.NodeID()] = ABS_UND
			} else {
				values[input.NodeID()] = state[i-primaryInputs]
			}
		}
		for id, node := range nodes {
			switch node.Kind {
			case aig.Constant, aig.Input:
			case aig.And:
				values[id] = ternaryAnd(ternaryAIGValue(values, node.Fanins[0]), ternaryAIGValue(values, node.Fanins[1]))
			case aig.ExclusiveOr:
				left, right := ternaryAIGValue(values, node.Fanins[0]), ternaryAIGValue(values, node.Fanins[1])
				if left == ABS_UND || right == ABS_UND {
					values[id] = ABS_UND
				} else if left == right {
					values[id] = ABS_ZER
				} else {
					values[id] = ABS_ONE
				}
			case aig.Buffer, aig.Output:
				values[id] = ternaryAIGValue(values, node.Fanins[0])
			}
		}
		packed := make([]uint32, (len(nodes)+15)/16)
		for id, value := range values {
			if value != 0 {
				packed[id>>4] = packed2Set(packed[id>>4], uint32(id), uint32(value))
			}
		}
		vec.Vec_PtrPush(result, packed)
		proved := true
		for i := 0; i < primaryOutputs; i++ {
			if ternaryAIGValue(values, graph.Node(outputs[i]).Fanins[0]) != ABS_ZER {
				proved = false
				break
			}
		}
		if !proved {
			if verbose != 0 {
				fmt.Printf("Ternary simulation found a non-zero or unknown output in frame %d.\n", frame)
			}
			return result
		}
		for i := 0; i < regs; i++ {
			driver := graph.Node(outputs[primaryOutputs+i]).Fanins[0]
			state[i] = ternaryAIGValue(values, driver)
		}
	}
	if verbose != 0 {
		fmt.Printf("Ternary simulation proved all outputs false for %d frames.\n", vec.Vec_PtrSize(result))
	}
	return result
}

func ternaryAIGValue(values []int32, edge aig.Edge) int32 {
	value := values[edge.NodeID()]
	if edge.Complemented() && value != 0 {
		return ternaryNot(value)
	}
	return value
}

func Abs_ManFreeAray(frames *vec.Vec_Ptr_t) { vec.Vec_PtrFree(frames) }

// Saig_BmcPerform preserves the established API while using the shared native
// AIG unroller, CNF encoder, and SAT solver. Backend-specific legacy knobs that
// the shared solver does not expose are accepted and ignored.
func Saig_BmcPerform(graph *aig.Graph, start, framesMax, nodesMax, timeout, conflictsOne, conflictsAll, verbose, overwrite int32, failedFrame *int32, silent, useSatoko int32) int32 {
	if graph == nil {
		return -1
	}
	params := Saig_ParBmc_t{
		NStart:      start,
		NFramesMax:  framesMax,
		NConfLimit:  conflictsOne,
		NTimeOut:    timeout,
		FVerbose:    verbose,
		FNotVerbose: silent,
	}
	status := Saig_ManBmcScalable(graph, &params)
	if failedFrame != nil {
		*failedFrame = params.IFrame
		if status == 0 {
			*failedFrame--
		}
	}
	_ = nodesMax
	_ = conflictsAll
	_ = overwrite
	_ = useSatoko
	if status == 0 {
		return 0
	}
	// The C entry point used -1 for a bounded-safe run and for early stop.
	return -1
}
