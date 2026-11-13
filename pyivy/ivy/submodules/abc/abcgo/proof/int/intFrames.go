package int

// Inter_ManFramesInter ports the timeframe construction in intFrames.c. The
// combinational inputs are initial register values (unless fixed to zero),
// followed by the primary inputs of each successive timeframe.
import "github.com/glycerine/abcgo/aig/aig"

func Inter_ManFramesInter(src *aig.Graph, nFrames, addRegOuts, useTwoFrames int32) *aig.Graph {
	if src == nil || src.RegisterCount() <= 0 || nFrames <= 0 {
		panic("interpolation requires a sequential graph and positive frame count")
	}
	registers := int(src.RegisterCount())
	inputs, outputs := src.Inputs(), src.Outputs()
	primaryInputs := len(inputs) - registers
	primaryOutputs := len(outputs) - registers
	constraints := int(aig.Aig_ManConstrNum(src))
	if primaryOutputs-constraints != 1 {
		panic("interpolation requires one property output")
	}
	dst := aig.NewGraph()
	state := make([]aig.Edge, registers)
	for i := range state {
		if addRegOuts != 0 {
			state[i] = dst.Const0()
		} else {
			state[i] = dst.NewInput()
		}
	}
	previousProperty := aig.InvalidEdge
	var copies map[int]aig.Edge
	for frame := int32(0); frame < nFrames; frame++ {
		copies = map[int]aig.Edge{0: dst.Const1()}
		for i := 0; i < primaryInputs; i++ {
			copies[inputs[i].NodeID()] = dst.NewInput()
		}
		for i, edge := range inputs[primaryInputs:] {
			copies[edge.NodeID()] = state[i]
		}
		for _, output := range outputs[primaryOutputs-constraints : primaryOutputs] {
			dst.AddOutput(duplicateTransitionEdge(src, dst, copies, src.Node(output).Fanins[0]).Not())
		}
		if frame == nFrames-1 {
			break
		}
		previousProperty = duplicateTransitionEdge(src, dst, copies, src.Node(outputs[0]).Fanins[0])
		next := make([]aig.Edge, registers)
		for i, output := range outputs[primaryOutputs:] {
			next[i] = duplicateTransitionEdge(src, dst, copies, src.Node(output).Fanins[0])
		}
		state = next
	}
	if addRegOuts != 0 {
		for _, output := range outputs[primaryOutputs:] {
			dst.AddOutput(duplicateTransitionEdge(src, dst, copies, src.Node(output).Fanins[0]))
		}
	} else {
		property := duplicateTransitionEdge(src, dst, copies, src.Node(outputs[0]).Fanins[0])
		if previousProperty != aig.InvalidEdge && useTwoFrames != 0 {
			property = dst.OR(previousProperty, property)
		}
		dst.AddOutput(property)
	}
	return dst
}
