package aig

func frameIndex(mapping []Edge, frames int32, e Edge, frame int32) int {
	if frames <= 0 || frame < 0 || frame >= frames {
		panic("aig: invalid frame index")
	}
	index := int(frames)*e.NodeID() + int(frame)
	if index < 0 || index >= len(mapping) {
		panic("aig: frame map outside range")
	}
	return index
}
func Aig_ObjFrames(mapping []Edge, frames int32, e Edge, frame int32) Edge {
	return mapping[frameIndex(mapping, frames, e, frame)]
}
func Aig_ObjSetFrames(mapping []Edge, frames int32, e Edge, frame int32, node Edge) {
	mapping[frameIndex(mapping, frames, e, frame)] = node
}
func Aig_ObjChild0Frames(p *Graph, mapping []Edge, frames int32, e Edge, frame int32) Edge {
	return frameChild(p, mapping, frames, e, frame, 0)
}
func Aig_ObjChild1Frames(p *Graph, mapping []Edge, frames int32, e Edge, frame int32) Edge {
	return frameChild(p, mapping, frames, e, frame, 1)
}
func frameChild(p *Graph, mapping []Edge, frames int32, e Edge, frame int32, pin int) Edge {
	fanins := nodeFanins(*p.node(e))
	if pin >= len(fanins) {
		return InvalidEdge
	}
	child := fanins[pin]
	value := Aig_ObjFrames(mapping, frames, child, frame)
	if value == InvalidEdge {
		return value
	}
	if child.Complemented() {
		value = value.Not()
	}
	return value
}

// Aig_ManFrames unrolls sequential state with ABC's original flag conventions.
// init fixes frame-zero state to zero; outs exports ordinary outputs; regs
// appends latch inputs. enlarge selects last-frame POs and first-frame LIs.
func Aig_ManFrames(p *Graph, frames, init, outs, regs, enlarge int32, returnedMap *[]Edge) *Graph {
	if frames <= 0 {
		panic("aig: frame count must be positive")
	}
	if p.nRegs < 0 || int(p.nRegs) > len(p.inputs) || int(p.nRegs) > len(p.outputs) {
		panic("aig: invalid sequential terminal counts")
	}
	mapping := make([]Edge, len(p.nodes)*int(frames))
	for i := range mapping {
		mapping[i] = InvalidEdge
	}
	q := Aig_ManStart(int32(len(p.nodes)) * frames)
	q.Name, q.Spec = p.Name, p.Spec
	piCount, poCount := len(p.inputs)-int(p.nRegs), len(p.outputs)-int(p.nRegs)
	for f := int32(0); f < frames; f++ {
		Aig_ObjSetFrames(mapping, frames, p.Const1(), f, q.Const1())
		for _, e := range p.inputs[:piCount] {
			Aig_ObjSetFrames(mapping, frames, e, f, q.NewInput())
		}
	}
	for _, e := range p.inputs[piCount:] {
		value := q.Const0()
		if init == 0 {
			value = q.NewInput()
		}
		Aig_ObjSetFrames(mapping, frames, e, 0, value)
	}
	active := make(map[int]bool)
	var visit func(Edge, int32) Edge
	visit = func(e Edge, f int32) Edge {
		index := frameIndex(mapping, frames, e, f)
		value := mapping[index]
		if value == InvalidEdge {
			if active[index] {
				panic("aig: cyclic frame expansion")
			}
			active[index] = true
			n := *p.node(e)
			switch n.Kind {
			case And, ExclusiveOr:
				a, b := visit(n.Fanins[0], f), visit(n.Fanins[1], f)
				if n.Kind == And {
					value = q.AND(a, b)
				} else {
					value = q.NewXor(a, b)
				}
			case Output, Buffer:
				value = visit(n.Fanins[0], f)
			default:
				panic("aig: unmapped frame terminal")
			}
			delete(active, index)
			mapping[index] = value
		}
		if e.Complemented() {
			value = value.Not()
		}
		return value
	}
	for f := int32(0); f < frames; f++ {
		// Include every internal object as in the source; cleanup later removes
		// unused cones while keeping frame-map indices stable.
		for id, n := range p.nodes {
			if !p.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
				visit(Edge(id)<<1, f)
			}
		}
		if f+1 < frames {
			for i, e := range p.inputs[piCount:] {
				next := visit(p.outputDriver(p.outputs[poCount+i]), f)
				Aig_ObjSetFrames(mapping, frames, e, f+1, next)
			}
		}
	}
	if outs != 0 {
		first := int32(0)
		if enlarge != 0 {
			first = frames - 1
		}
		for f := first; f < frames; f++ {
			for _, e := range p.outputs[:poCount] {
				output := q.AddOutput(visit(p.outputDriver(e), f))
				Aig_ObjSetFrames(mapping, frames, e, f, output)
			}
		}
	}
	if regs != 0 {
		sourceFrame := frames - 1
		if enlarge != 0 {
			sourceFrame = 0
		}
		for _, e := range p.outputs[poCount:] {
			output := q.AddOutput(visit(p.outputDriver(e), sourceFrame))
			Aig_ObjSetFrames(mapping, frames, e, frames-1, output)
		}
		// Initialized expansions have no state CIs and are combinational BMC
		// managers; do not label ordinary PIs as register terminals.
		if init == 0 {
			Aig_ManSetRegNum(q, p.nRegs)
		}
	}
	Aig_ManCleanup(q)
	if returnedMap != nil {
		*returnedMap = mapping
	}
	return q
}
