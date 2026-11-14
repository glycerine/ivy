package aig

// CNF/SAT/proof interpolation drivers live in proof/aiginterp so they can
// depend on the graph, CNF mapper and solver without an import cycle.

func interCopy(source, destination *Graph, e Edge, inputs []Edge) Edge {
	if len(source.inputs) != len(inputs) {
		panic("aig: inconsistent shared-variable map")
	}
	mapping := map[int]Edge{0: destination.Const1()}
	for i, input := range source.inputs {
		mapping[input.NodeID()] = inputs[i]
	}
	return Aig_NodeBddToMuxes_rec(source, e, destination, mapping)
}
