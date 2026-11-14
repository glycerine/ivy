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
	return interCopy_rec(source, e, destination, mapping)
}

func interCopy_rec(decisions *Graph, function Edge, destination *Graph, mapping map[int]Edge) Edge {
	id := function.NodeID()
	value, ok := mapping[id]
	if !ok {
		n := *decisions.node(function)
		if n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer && n.Kind != Output {
			panic("aig: unmapped decision variable")
		}
		a := interCopy_rec(decisions, n.Fanins[0], destination, mapping)
		if n.Kind == Buffer || n.Kind == Output {
			value = a
		} else {
			b := interCopy_rec(decisions, n.Fanins[1], destination, mapping)
			if n.Kind == And {
				value = destination.AND(a, b)
			} else {
				value = destination.NewXor(a, b)
			}
		}
		mapping[id] = value
	}
	if function.Complemented() {
		value = value.Not()
	}
	return value
}
