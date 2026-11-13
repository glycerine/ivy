package fra

import "github.com/glycerine/abcgo/aig/aig"

// Fra_FramesConstrainNode performs one speculative class merge and emits the
// associated equality constraint into the time-frame graph.
func Fra_FramesConstrainNode(p *Fra_Man_t, graph *aig.Graph, obj aig.Edge, frame int32) {
	repr := Fra_ClassObjRepr(p, obj)
	if repr == aig.InvalidEdge {
		return
	}
	if repr.NodeID() >= obj.NodeID() {
		panic("fra: representative is not topologically earlier")
	}
	node, representative := Fra_ObjFraig(p, obj, frame), Fra_ObjFraig(p, repr, frame)
	if node == aig.InvalidEdge || representative == aig.InvalidEdge {
		panic("fra: missing speculative frame node")
	}
	if aig.Aig_Regular(node) == aig.Aig_Regular(representative) {
		return
	}
	reduced := representative
	if p.pManAig.Node(obj).Phase != p.pManAig.Node(repr).Phase {
		reduced = reduced.Not()
	}
	Fra_ObjSetFraig(p, obj, frame, reduced)
	constraint := graph.XOR(node, representative)
	if graph.Node(constraint).Phase == constraint.Complemented() {
		constraint = constraint.Not()
	}
	graph.AddOutput(constraint)
}

func Fra_FramesWithClasses(p *Fra_Man_t) *aig.Graph {
	source := p.pManAig
	if p.pManFraig != nil || source.RegisterCount() <= 0 || int(source.RegisterCount()) >= source.InputCount() {
		panic("fra: invalid induction frame manager")
	}
	graph := aig.NewGraph()
	graph.Name, graph.Spec = source.Name, source.Spec
	inputs, outputs := source.Inputs(), source.Outputs()
	regs := int(source.RegisterCount())
	trueInputs := len(inputs) - regs
	for f := int32(0); f < p.nFramesAll; f++ {
		Fra_ObjSetFraig(p, source.Const1(), f, graph.Const1())
		for _, input := range inputs[:trueInputs] {
			Fra_ObjSetFraig(p, input, f, graph.NewInput())
		}
	}
	for _, latch := range inputs[trueInputs:] {
		Fra_ObjSetFraig(p, latch, 0, graph.NewInput())
	}
	aig.Aig_ManSetRegNumForConstruction(graph, source.RegisterCount())
	for f := int32(0); f < p.nFramesAll-1; f++ {
		for _, latch := range inputs[trueInputs:] {
			Fra_FramesConstrainNode(p, graph, latch, f)
		}
		for _, obj := range source.LiveEdges() {
			node := source.Node(obj)
			var mapped aig.Edge
			switch node.Kind {
			case aig.And:
				mapped = graph.AND(Fra_ObjChild0Fra(p, obj, f), Fra_ObjChild1Fra(p, obj, f))
			case aig.ExclusiveOr:
				mapped = graph.NewXor(Fra_ObjChild0Fra(p, obj, f), Fra_ObjChild1Fra(p, obj, f))
			case aig.Buffer:
				mapped = Fra_ObjChild0Fra(p, obj, f)
			default:
				continue
			}
			Fra_ObjSetFraig(p, obj, f, mapped)
			Fra_FramesConstrainNode(p, graph, obj, f)
		}
		for i, output := range outputs[len(outputs)-regs:] {
			Fra_ObjSetFraig(p, inputs[trueInputs+i], f+1, Fra_ObjChild0Fra(p, output, f))
		}
	}
	graph.Assertions = int32(graph.OutputCount())
	for _, latch := range inputs[trueInputs:] {
		graph.AddOutput(Fra_ObjFraig(p, latch, p.nFramesAll-1))
	}
	aig.Aig_ManSetRegNum(graph, source.RegisterCount())
	aig.Aig_ManCleanup(graph)
	// Cleanup can remove a frame node that was mapped before it became
	// dangling. C leaves an AIG_OBJ_NONE marker; clear the Go mapping.
	for i, edge := range p.pMemFraig {
		if edge != aig.InvalidEdge && aig.Aig_ObjIsNone(graph, edge) {
			p.pMemFraig[i] = aig.InvalidEdge
		}
	}
	return graph
}
