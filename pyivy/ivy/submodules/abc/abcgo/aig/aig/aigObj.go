package aig

import "fmt"

func Aig_ObjCreateCi(p *Graph) Edge              { return p.NewInput() }
func Aig_ObjCreateCo(p *Graph, driver Edge) Edge { return p.AddOutput(driver) }

func Aig_ObjCreate(p *Graph, ghost *Node) Edge {
	if ghost.Kind == And {
		return p.AND(ghost.Fanins[0], ghost.Fanins[1])
	}
	if ghost.Kind != ExclusiveOr && ghost.Kind != Buffer {
		panic("aig: invalid ghost kind")
	}
	n := *ghost
	if !p.valid(n.Fanins[0]) || (n.Kind == ExclusiveOr && !p.valid(n.Fanins[1])) {
		panic("aig: invalid ghost fanins")
	}
	if n.Kind == ExclusiveOr {
		if n.Fanins[0] > n.Fanins[1] {
			n.Fanins[0], n.Fanins[1] = n.Fanins[1], n.Fanins[0]
		}
		if old := Aig_TableLookup(p, &n); old != InvalidEdge {
			return old
		}
	}
	e := Edge(len(p.nodes)) << 1
	n.Refs = 0
	p.nodes = append(p.nodes, n)
	if n.Kind == ExclusiveOr {
		Aig_TableInsert(p, e)
	}
	p.recountReferences()
	p.recomputeLevels()
	p.refreshFanouts()
	return e
}
func Aig_ObjConnect(p *Graph, e, a, b Edge) {
	if !p.valid(a) {
		panic("aig: invalid first fanin")
	}
	n := p.node(e)
	if (n.Kind == And || n.Kind == ExclusiveOr) && !p.valid(b) {
		panic("aig: invalid second fanin")
	}
	if n.Kind == And || n.Kind == ExclusiveOr {
		if a > b {
			a, b = b, a
		}
	}
	n.Fanins = [2]Edge{a, b}
	p.recountReferences()
	p.recomputeLevels()
	if n.Kind == And || n.Kind == ExclusiveOr {
		Aig_TableInsert(p, e)
	}
	p.refreshFanouts()
}
func Aig_ObjDisconnect(p *Graph, e Edge) {
	n := p.node(e)
	if n.Kind == And || n.Kind == ExclusiveOr {
		table := p.nodeTable(n.Kind)
		if table[tableKey(n)] == Aig_Regular(e) {
			delete(table, tableKey(n))
		}
	}
	if p.fanouts != nil {
		for pin, fanin := range nodeFanins(*n) {
			if p.valid(fanin) {
				Aig_ObjRemoveFanoutPin(p, Aig_Regular(fanin), Aig_Regular(e), int32(pin))
			}
		}
	}
	for _, fanin := range nodeFanins(*n) {
		if p.valid(fanin) {
			if p.node(fanin).Refs <= 0 {
				panic("aig: reference underflow")
			}
			p.node(fanin).Refs--
		}
	}
	n.Fanins = [2]Edge{InvalidEdge, InvalidEdge}
}
func Aig_ObjDelete(p *Graph, e Edge) {
	if e.Complemented() || p.node(e).Refs != 0 {
		panic("aig: deleting complemented/referenced object")
	}
	n := p.node(e)
	if n.Kind == Input || n.Kind == Constant {
		panic("aig: cannot delete terminal directly")
	}
	if n.Kind == And || n.Kind == ExclusiveOr {
		table := p.nodeTable(n.Kind)
		if table[tableKey(n)] == e {
			delete(table, tableKey(n))
		}
	}
	p.deleted[e.NodeID()] = true
	*n = Node{}
	p.refreshFanouts()
}
func Aig_ObjDelete_rec(p *Graph, e Edge, freeTop int32) {
	p.editDepth++
	defer func() {
		p.editDepth--
		if freeTop != 0 {
			p.refreshFanouts()
		}
	}()
	n := *p.node(e)
	if n.Kind == Constant || n.Kind == Input {
		return
	}
	if n.Kind == Output {
		panic("aig: recursive deletion of output")
	}
	fanins := nodeFanins(n)
	Aig_ObjDisconnect(p, e)
	if freeTop != 0 {
		Aig_ObjDelete(p, e)
	}
	for _, x := range fanins {
		if p.valid(x) && p.node(x).Refs == 0 {
			Aig_ObjDelete_rec(p, Aig_Regular(x), 1)
		}
	}
}
func Aig_ObjDeletePo(p *Graph, e Edge) {
	index := -1
	for i, x := range p.outputs {
		if x == e {
			index = i
			break
		}
	}
	if index < 0 {
		panic("aig: output not found")
	}
	p.outputs = append(p.outputs[:index], p.outputs[index+1:]...)
	p.node(e).Refs--
	if p.node(e).Kind == Output {
		Aig_ObjDisconnect(p, e)
		Aig_ObjDelete(p, e)
	}
	p.refreshFanouts()
}
func Aig_ObjPatchFanin0(p *Graph, e, driver Edge) {
	if p.node(e).Kind != Output || !p.valid(driver) {
		panic("aig: output patch requires valid output/driver")
	}
	old := p.node(e).Fanins[0]
	p.node(old).Refs--
	p.node(e).Fanins[0] = driver
	p.node(driver).Refs++
	p.recomputeLevels()
	p.refreshFanouts()
	if p.node(old).Refs == 0 && p.node(old).Kind != Input && p.node(old).Kind != Constant {
		Aig_ObjDelete_rec(p, Aig_Regular(old), 1)
	}
}
func Aig_ObjPrint(p *Graph, e Edge) {
	n := p.node(e)
	fmt.Printf("%d: kind=%d fanins=(%d,%d) level=%d refs=%d\n", e.NodeID(), n.Kind, n.Fanins[0], n.Fanins[1], n.Level, n.Refs)
}
func (p *Graph) real(e Edge) Edge {
	seen := make(map[int]bool)
	for p.node(e).Kind == Buffer {
		if seen[e.NodeID()] {
			panic("aig: buffer cycle")
		}
		seen[e.NodeID()] = true
		next := p.node(e).Fanins[0]
		if e.Complemented() {
			next = next.Not()
		}
		e = next
	}
	return e
}
func Aig_NodeFixBufferFanins(p *Graph, e Edge, updateLevel int32) {
	n := *p.node(e)
	if n.Kind == Output {
		Aig_ObjPatchFanin0(p, e, p.real(n.Fanins[0]))
		return
	}
	if n.Kind != And && n.Kind != ExclusiveOr {
		panic("aig: buffer propagation requires node")
	}
	a, b := p.real(n.Fanins[0]), p.real(n.Fanins[1])
	result := Aig_Oper(p, a, b, int32(Aig_ObjType(p, e)))
	Aig_ObjReplace(p, e, result, updateLevel)
}
func Aig_ManPropagateBuffers(p *Graph, updateLevel int32) int32 {
	var changed int32
	for {
		progress := false
		for id, n := range p.nodes {
			if p.deleted[id] {
				continue
			}
			if n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Output {
				continue
			}
			needs := false
			for _, e := range nodeFanins(n) {
				if p.node(e).Kind == Buffer {
					needs = true
				}
			}
			if needs {
				Aig_NodeFixBufferFanins(p, Edge(id)<<1, updateLevel)
				changed++
				progress = true
			}
		}
		if !progress {
			break
		}
	}
	for i, e := range p.outputs {
		if p.node(e).Kind == Buffer {
			p.outputs[i] = p.real(e)
			changed++
		}
	}
	p.recountReferences()
	Aig_ManCleanup(p)
	p.recomputeLevels()
	p.refreshFanouts()
	return changed
}
func Aig_ObjReplace(p *Graph, old, replacement Edge, updateLevel int32) {
	p.editDepth++
	defer func() { p.editDepth--; p.refreshFanouts() }()
	if old.Complemented() || old.NodeID() == replacement.NodeID() {
		panic("aig: invalid replacement")
	}
	if p.node(old).Kind == Input || p.node(old).Kind == Output || p.node(old).Kind == Constant {
		panic("aig: cannot replace terminal")
	}
	if p.node(replacement).Kind == Buffer || p.node(replacement).Kind == Output {
		panic("aig: invalid replacement object")
	}
	for _, e := range p.dfs([]Edge{replacement}, false, false, false) {
		if e.NodeID() == old.NodeID() {
			panic("aig: replacement introduces cycle")
		}
	}
	p.node(replacement).Refs++
	Aig_ObjDelete_rec(p, old, 0)
	p.node(replacement).Refs--
	n := p.node(old)
	r := p.node(replacement)
	if replacement.Complemented() || r.Refs != 0 || r.Kind != And && r.Kind != ExclusiveOr {
		// The source keeps a buffer when the replacement is complemented,
		// referenced, or a terminal.
		n.Kind = Buffer
		n.Fanins = [2]Edge{replacement, InvalidEdge}
		r.Refs++
	} else {
		// Move an unused positive node into the old object slot, preserving
		// the old ID and its fanouts as in Aig_ObjReplace in aigObj.c.
		kind, fanins := r.Kind, r.Fanins
		Aig_ObjDisconnect(p, replacement)
		n.Kind, n.Fanins = kind, fanins
		for _, fanin := range nodeFanins(*n) {
			p.node(fanin).Refs++
		}
		Aig_TableInsert(p, old)
		Aig_ObjDelete(p, replacement)
	}
	p.recomputeLevels()
	p.refreshFanouts()
}
func nodeFanins(n Node) []Edge {
	if n.Kind == And || n.Kind == ExclusiveOr {
		return n.Fanins[:]
	}
	if n.Kind == Buffer || n.Kind == Output {
		return n.Fanins[:1]
	}
	return nil
}
func (p *Graph) recomputeLevels() {
	state := make([]uint8, len(p.nodes))
	var visit func(int)
	visit = func(id int) {
		if state[id] == 2 {
			return
		}
		if state[id] == 1 {
			panic("aig: cycle in graph")
		}
		state[id] = 1
		n := p.nodes[id]
		level := 0
		if n.Kind == Input {
			level = n.Level
		}
		phase := n.Kind == Constant
		fanins := nodeFanins(n)
		for _, e := range fanins {
			if !p.valid(e) {
				panic("aig: disconnected graph node")
			}
			visit(e.NodeID())
			if l := p.node(e).Level; l > level {
				level = l
			}
		}
		if len(fanins) > 0 {
			phase = Aig_ObjPhaseReal(p, fanins[0])
			if len(fanins) == 2 {
				level++
				if n.Kind == And {
					phase = phase && Aig_ObjPhaseReal(p, fanins[1])
				} else {
					phase = phase != Aig_ObjPhaseReal(p, fanins[1])
				}
			}
		}
		p.nodes[id].Level = level
		p.nodes[id].Phase = phase
		state[id] = 2
	}
	for id := range p.nodes {
		if !p.deleted[id] {
			visit(id)
		}
	}
}
func (p *Graph) refreshFanouts() {
	if p.editDepth == 0 && p.fanouts != nil {
		p.fanouts.Rebuild(p)
	}
}

// Aig_ObjReplaceInputFanouts redirects all uses of an input, keeping the
// original input in the interface. The replacement must not depend on it.
func Aig_ObjReplaceInputFanouts(p *Graph, input, replacement Edge) {
	if input.Complemented() || !p.valid(input) || p.node(input).Kind != Input || !p.valid(replacement) {
		panic("aig: invalid input replacement")
	}
	for _, e := range p.dfs([]Edge{replacement}, true, false, false) {
		if e.NodeID() == input.NodeID() {
			panic("aig: input replacement introduces cycle")
		}
	}
	// Snapshot the original consumers; rebuilding one may delete another.
	for _, e := range p.LiveEdges() {
		if !p.valid(e) {
			continue
		}
		n := *p.node(e)
		if n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer && n.Kind != Output {
			continue
		}
		fanins := n.Fanins
		changed := false
		for i, f := range nodeFanins(n) {
			if f.NodeID() == input.NodeID() {
				fanins[i] = Aig_NotCond(replacement, int32(f&1))
				changed = true
			}
		}
		if !changed {
			continue
		}
		switch n.Kind {
		case Output:
			Aig_ObjPatchFanin0(p, e, fanins[0])
		case Buffer:
			Aig_ObjReplace(p, e, fanins[0], 0)
		case And:
			Aig_ObjReplace(p, e, p.AND(fanins[0], fanins[1]), 0)
		case ExclusiveOr:
			Aig_ObjReplace(p, e, p.NewXor(fanins[0], fanins[1]), 0)
		}
	}
}
