package aig

// Representative maps store regular edges; absence is InvalidEdge, since edge
// zero is the valid constant-one representative. Canonical links decrease IDs.
func Aig_ManReprStart(g *Graph, capacity int32) {
	if capacity < 0 || g.reprs != nil {
		panic("aig: representative storage already started or invalid capacity")
	}
	for id, n := range g.nodes {
		if !g.deleted[id] && n.Kind == Buffer {
			panic("aig: representatives require buffer-free graph")
		}
	}
	g.reprs = make(map[int]Edge, int(capacity))
}
func Aig_ManReprStop(g *Graph) {
	if g.reprs == nil {
		panic("aig: representative storage not started")
	}
	g.reprs = nil
}
func reprValidate(g *Graph, e Edge) {
	if g.reprs == nil || e.Complemented() || !g.valid(e) {
		panic("aig: invalid representative operand")
	}
}
func Aig_ObjCreateRepr(g *Graph, representative, node Edge) {
	reprValidate(g, representative)
	reprValidate(g, node)
	if representative.NodeID() >= node.NodeID() {
		panic("aig: representative ID must decrease")
	}
	g.reprs[node.NodeID()] = representative
}
func Aig_ObjSetRepr_(g *Graph, a, b Edge) {
	reprValidate(g, a)
	reprValidate(g, b)
	if a == b {
		return
	}
	if a.NodeID() > b.NodeID() {
		a, b = b, a
	}
	g.reprs[b.NodeID()] = a
}
func Aig_ObjFindRepr(g *Graph, node Edge) Edge {
	reprValidate(g, node)
	if representative, ok := g.reprs[node.NodeID()]; ok {
		if representative != InvalidEdge {
			reprValidate(g, representative)
		}
		return representative
	}
	return InvalidEdge
}
func Aig_ObjClearRepr(g *Graph, node Edge) { reprValidate(g, node); delete(g.reprs, node.NodeID()) }
func Aig_ObjFindReprTransitive(g *Graph, node Edge) Edge {
	reprValidate(g, node)
	representative := Aig_ObjFindRepr(g, node)
	if representative == InvalidEdge {
		return InvalidEdge
	}
	seen := map[int]bool{node.NodeID(): true}
	for {
		if seen[representative.NodeID()] {
			panic("aig: cyclic representatives")
		}
		seen[representative.NodeID()] = true
		next := Aig_ObjFindRepr(g, representative)
		if next == InvalidEdge {
			return representative
		}
		representative = next
	}
}
func Aig_ObjGetRepr(g *Graph, node Edge) Edge {
	regular := Aig_Regular(node)
	representative := Aig_ObjFindRepr(g, regular)
	if representative == InvalidEdge {
		representative = regular
	}
	copy, ok := g.copies[representative.NodeID()]
	if !ok {
		return InvalidEdge
	}
	phase := g.node(regular).Phase != g.node(representative).Phase
	return dupPhase(copy, phase != node.Complemented())
}
func Aig_ObjChild0Repr(g *Graph, node Edge) Edge { return Aig_ObjGetRepr(g, g.node(node).Fanins[0]) }
func Aig_ObjChild1Repr(g *Graph, node Edge) Edge { return Aig_ObjGetRepr(g, g.node(node).Fanins[1]) }
func Aig_ManTransferRepr(destination, source *Graph) {
	if destination.reprs == nil || source.reprs == nil {
		panic("aig: representative storage not started")
	}
	for id := range source.nodes {
		if source.deleted[id] {
			continue
		}
		representative := Aig_ObjFindRepr(source, Edge(id)<<1)
		if representative == InvalidEdge {
			continue
		}
		a, aok := source.copies[id]
		b, bok := source.copies[representative.NodeID()]
		if !aok || !bok || a == InvalidEdge || b == InvalidEdge {
			continue
		}
		a, b = Aig_Regular(a), Aig_Regular(b)
		if destination.valid(a) && destination.valid(b) {
			Aig_ObjSetRepr_(destination, a, b)
		}
	}
}

// Existing copy bindings take precedence here, matching aigRepr.c: CI copies
// remain distinct, while representatives are applied at internal fanouts. This
// differs intentionally from aigDup.c's eager representative duplicator.
func reprDuplicate(dst, src *Graph, node Edge, active map[int]bool) Edge {
	id := node.NodeID()
	if copy, ok := src.copies[id]; ok {
		return dupPhase(copy, node.Complemented())
	}
	if active[id] {
		panic("aig: cyclic representative cone")
	}
	active[id] = true
	defer delete(active, id)
	if representative := Aig_ObjFindRepr(src, Aig_Regular(node)); representative != InvalidEdge {
		copy := reprDuplicate(dst, src, representative, active)
		copy = dupPhase(copy, src.node(node).Phase != src.node(representative).Phase)
		src.copies[id] = copy
		return dupPhase(copy, node.Complemented())
	}
	n := src.node(node)
	var copy Edge
	switch n.Kind {
	case Constant:
		copy = dst.Const1()
	case Input:
		panic("aig: missing CI copy")
	case Output, Buffer:
		reprDuplicate(dst, src, Aig_Regular(n.Fanins[0]), active)
		copy = Aig_ObjChild0Repr(src, node)
	case And, ExclusiveOr:
		reprDuplicate(dst, src, Aig_Regular(n.Fanins[0]), active)
		reprDuplicate(dst, src, Aig_Regular(n.Fanins[1]), active)
		a, b := Aig_ObjChild0Repr(src, node), Aig_ObjChild1Repr(src, node)
		if n.Kind == And {
			copy = dst.AND(a, b)
		} else {
			copy = dst.NewXor(a, b)
		}
	default:
		panic("aig: invalid representative node")
	}
	src.copies[id] = copy
	return dupPhase(copy, node.Complemented())
}
func Aig_ManDupRepr_rec(dst, src *Graph, node Edge) Edge {
	return reprDuplicate(dst, src, node, map[int]bool{})
}
func Aig_ManDupRepr(src *Graph, ordered int32) *Graph {
	if src.reprs == nil {
		panic("aig: representative storage not started")
	}
	dst := dupStart(src)
	defer dst.updateTrueTerminalCounts()
	if ordered != 0 {
		for id, n := range src.nodes {
			if src.deleted[id] {
				continue
			}
			if n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer {
				continue
			}
			a := Aig_ObjChild0Repr(src, Edge(id)<<1)
			if a == InvalidEdge {
				panic("aig: unordered representative graph")
			}
			var copy Edge
			if n.Kind == Buffer {
				copy = a
			} else {
				b := Aig_ObjChild1Repr(src, Edge(id)<<1)
				if b == InvalidEdge {
					panic("aig: unordered representative graph")
				}
				if n.Kind == And {
					copy = dst.AND(a, b)
				} else {
					copy = dst.NewXor(a, b)
				}
			}
			src.copies[id] = copy
		}
	} else {
		for _, output := range src.outputs {
			Aig_ManDupRepr_rec(dst, src, src.outputDriver(output))
		}
	}
	for _, output := range src.outputs {
		driver := Aig_ObjGetRepr(src, src.outputDriver(output))
		if driver == InvalidEdge {
			panic("aig: missing representative output")
		}
		copy := Aig_ObjCreateCo(dst, driver)
		if src.node(output).Kind == Output {
			src.copies[output.NodeID()] = copy
		}
	}
	return dst
}

// Cleanup retains register terminals, but disconnects latch cones unreachable
// from primary outputs. Latch dependencies are followed to their next-state COs.
func reprSequentialCleanup(g *Graph) {
	regs := int(g.nRegs)
	if regs < 0 || regs > len(g.inputs) || regs > len(g.outputs) {
		panic("aig: invalid register count")
	}
	latchInputs := map[int]Edge{}
	for i := 0; i < regs; i++ {
		latchInputs[g.inputs[len(g.inputs)-regs+i].NodeID()] = g.outputs[len(g.outputs)-regs+i]
	}
	live := map[int]bool{}
	queue := append([]Edge(nil), g.outputs[:len(g.outputs)-regs]...)
	var visit func(Edge)
	visit = func(e Edge) {
		id := e.NodeID()
		if live[id] {
			return
		}
		live[id] = true
		n := g.node(e)
		switch n.Kind {
		case Input:
			if next, ok := latchInputs[id]; ok {
				queue = append(queue, next)
			}
		case And, ExclusiveOr:
			visit(n.Fanins[0])
			visit(n.Fanins[1])
		case Buffer, Output:
			visit(n.Fanins[0])
		}
	}
	for i := 0; i < len(queue); i++ {
		visit(queue[i])
	}
	for i, output := range g.outputs {
		if live[output.NodeID()] {
			continue
		}
		if g.node(output).Kind == Output {
			Aig_ObjPatchFanin0(g, output, g.Const0())
		} else {
			g.outputs[i] = g.Const0()
		}
	}
	g.recountReferences()
	Aig_ManCleanup(g)
	g.refreshFanouts()
}
func Aig_ManDupReprBasic(src *Graph) *Graph {
	dst := Aig_ManDupRepr(src, 0)
	reprSequentialCleanup(dst)
	for id, copy := range src.copies {
		if copy == InvalidEdge || !dst.valid(copy) {
			delete(src.copies, id)
		}
	}
	return dst
}
func Aig_ManRemapRepr(g *Graph) int32 {
	var fanouts int32
	for id, n := range g.nodes {
		if g.deleted[id] || (n.Kind != And && n.Kind != ExclusiveOr) {
			continue
		}
		node := Edge(id) << 1
		representative := Aig_ObjFindReprTransitive(g, node)
		if representative == InvalidEdge {
			continue
		}
		if representative.NodeID() >= id {
			panic("aig: representative ID must decrease")
		}
		Aig_ObjSetRepr_(g, node, representative)
		if n.Refs > 0 {
			fanouts++
		}
	}
	return fanouts
}
func Aig_ManCountReprs(g *Graph) int32 {
	var count int32
	for id, representative := range g.reprs {
		if id >= 0 && id < len(g.nodes) && !g.deleted[id] && representative != InvalidEdge {
			count++
		}
	}
	return count
}
func reprTfi(g *Graph, node, target Edge, visited map[int]bool) bool {
	if node == InvalidEdge || !g.valid(node) || g.node(node).Kind == Input {
		return false
	}
	node = Aig_Regular(node)
	if node == target {
		return true
	}
	if visited[node.NodeID()] {
		return false
	}
	visited[node.NodeID()] = true
	n := g.node(node)
	if n.Kind == And || n.Kind == ExclusiveOr {
		if reprTfi(g, n.Fanins[0], target, visited) || reprTfi(g, n.Fanins[1], target, visited) {
			return true
		}
	} else if n.Kind == Buffer || n.Kind == Output {
		if reprTfi(g, n.Fanins[0], target, visited) {
			return true
		}
	}
	if equivalent, ok := g.equivs[node.NodeID()]; ok {
		return reprTfi(g, equivalent, target, visited)
	}
	return false
}
func Aig_ObjCheckTfi_rec(g *Graph, node, target Edge) int32 {
	if reprTfi(g, node, Aig_Regular(target), map[int]bool{}) {
		return 1
	}
	return 0
}
func Aig_ObjCheckTfi(g *Graph, node, target Edge) int32 {
	if node.Complemented() || target.Complemented() {
		panic("aig: TFI roots must be regular")
	}
	return Aig_ObjCheckTfi_rec(g, node, target)
}
func Aig_ManRehash(g *Graph) *Graph {
	if g.reprs == nil {
		panic("aig: representative storage not started")
	}
	for Aig_ManRemapRepr(g) != 0 {
		old := g
		g = Aig_ManDupRepr(old, 1)
		Aig_ManReprStart(g, int32(len(g.nodes)))
		Aig_ManTransferRepr(g, old)
		Aig_ManStop(old)
	}
	return g
}
func Aig_ManMarkValidChoices(g *Graph) {
	if g.reprs == nil || g.equivs != nil {
		panic("aig: invalid choice conversion state")
	}
	g.equivs = map[int]Edge{}
	for id, n := range g.nodes {
		if g.deleted[id] || (n.Kind != And && n.Kind != ExclusiveOr) {
			continue
		}
		node := Edge(id) << 1
		representative := Aig_ObjFindRepr(g, node)
		if representative == InvalidEdge {
			continue
		}
		kind := g.node(representative).Kind
		if (kind != And && kind != ExclusiveOr) || Aig_ObjCheckTfi(g, node, representative) != 0 || n.Refs > 0 {
			Aig_ObjClearRepr(g, node)
			continue
		}
		if next, ok := g.equivs[representative.NodeID()]; ok {
			g.equivs[id] = next
		}
		g.equivs[representative.NodeID()] = node
	}
}
func Aig_TransferMappedClasses(destination, part *Graph, mapping []int32) int32 {
	if part.reprs == nil {
		return 0
	}
	if destination.reprs == nil {
		panic("aig: representative storage not started")
	}
	var count int32
	for id := range part.nodes {
		if part.deleted[id] {
			continue
		}
		representative := Aig_ObjFindRepr(part, Edge(id)<<1)
		if representative == InvalidEdge {
			continue
		}
		if id >= len(mapping) || representative.NodeID() >= len(mapping) {
			panic("aig: incomplete class mapping")
		}
		a, b := mapping[id], mapping[representative.NodeID()]
		if a < 0 || b < 0 || int(a) >= len(destination.nodes) || int(b) >= len(destination.nodes) {
			panic("aig: mapped class outside destination")
		}
		Aig_ObjSetRepr_(destination, Edge(a)<<1, Edge(b)<<1)
		count++
	}
	return count
}
