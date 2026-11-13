package fra

import "github.com/glycerine/abcgo/aig/aig"

// fraUnrollInitialized builds the initialized timeframes used by Fra_BmcFrames.
// The returned mapping is indexed by original node ID * frames + frame.
func fraUnrollInitialized(original *aig.Graph, frames int, keepOutputs bool) (*aig.Graph, []aig.Edge) {
	if original == nil || frames <= 0 {
		panic("fra: invalid BMC unrolling dimensions")
	}
	regs := int(original.RegisterCount())
	inputs, outputs, nodes := original.Inputs(), original.Outputs(), original.Nodes()
	if regs > len(inputs) || regs > len(outputs) {
		panic("fra: invalid sequential interface")
	}
	graph := aig.NewGraph()
	graph.Name, graph.Spec = original.Name, original.Spec
	mapping := make([]aig.Edge, len(nodes)*frames)
	for i := range mapping {
		mapping[i] = aig.InvalidEdge
	}
	set := func(edge aig.Edge, frame int, value aig.Edge) { mapping[edge.NodeID()*frames+frame] = value }
	get := func(edge aig.Edge, frame int) aig.Edge {
		if edge == aig.InvalidEdge {
			return edge
		}
		value := mapping[edge.NodeID()*frames+frame]
		if value == aig.InvalidEdge {
			panic("fra: missing unrolled node")
		}
		if edge.Complemented() {
			return value.Not()
		}
		return value
	}
	for f := 0; f < frames; f++ {
		set(original.Const1(), f, graph.Const1())
		for _, input := range inputs[:len(inputs)-regs] {
			set(input, f, graph.NewInput())
		}
	}
	for _, latch := range inputs[len(inputs)-regs:] {
		set(latch, 0, graph.Const0())
	}
	for f := 0; f < frames; f++ {
		for _, edge := range original.LiveEdges() {
			node := nodes[edge.NodeID()]
			switch node.Kind {
			case aig.And:
				set(edge, f, graph.AND(get(node.Fanins[0], f), get(node.Fanins[1], f)))
			case aig.ExclusiveOr:
				set(edge, f, graph.XOR(get(node.Fanins[0], f), get(node.Fanins[1], f)))
			case aig.Buffer:
				set(edge, f, get(node.Fanins[0], f))
			}
		}
		if f+1 < frames {
			for i, output := range outputs[len(outputs)-regs:] {
				set(inputs[len(inputs)-regs+i], f+1, get(nodes[output.NodeID()].Fanins[0], f))
			}
		}
	}
	if keepOutputs {
		for f := 0; f < frames; f++ {
			for _, output := range outputs[:len(outputs)-regs] {
				graph.AddOutput(get(nodes[output.NodeID()].Fanins[0], f))
			}
		}
		aig.Aig_ManCleanup(graph)
	} else {
		for _, edge := range graph.LiveEdges() {
			node := graph.Node(edge)
			if (node.Kind == aig.And || node.Kind == aig.ExclusiveOr) && node.Refs == 0 {
				graph.AddOutput(edge)
			}
		}
	}
	return graph, mapping
}
