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
	if n.Kind == Buffer {
		n.Fanins[1] = InvalidEdge
	}
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
	n.Level, n.Phase = p.levelNew(n), p.phaseNew(n)
	p.nodes = append(p.nodes, n)
	for _, fanin := range nodeFanins(n) {
		p.node(fanin).Refs++
	}
	if n.Kind == ExclusiveOr {
		Aig_TableInsert(p, e)
	}
	p.linkFanins(e.NodeID())
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
	// ABC connects only disconnected objects; the Go API also rewires.
	for _, fanin := range nodeFanins(*n) {
		if p.valid(fanin) {
			Aig_ObjDisconnect(p, e)
			break
		}
	}
	if n.Kind == And || n.Kind == ExclusiveOr {
		if a > b {
			a, b = b, a
		}
	}
	if n.Kind == Buffer {
		b = InvalidEdge
	}
	n.Fanins = [2]Edge{a, b}
	for _, fanin := range nodeFanins(*n) {
		p.node(fanin).Refs++
	}
	n.Level, n.Phase = p.levelNew(*n), p.phaseNew(*n)
	p.linkFanins(e.NodeID())
	if n.Kind == And || n.Kind == ExclusiveOr {
		Aig_TableInsert(p, e)
	}
	p.refreshFanouts()
}
func Aig_ObjDisconnect(p *Graph, e Edge) {
	n := p.node(e)
	if n.Kind == And || n.Kind == ExclusiveOr {
		table := p.nodeTable(n.Kind)
		if old, _ := table.get(tableKey(n)); old == Aig_Regular(e) {
			table.del(tableKey(n))
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
	if n.Kind == Output {
		for _, out := range p.outputs {
			if out.NodeID() == e.NodeID() {
				panic("aig: use output deletion for an attached CO")
			}
		}
	}
	if n.Kind == Buffer {
		for i, b := range p.bufs {
			if b.NodeID() == e.NodeID() {
				p.bufs = append(p.bufs[:i], p.bufs[i+1:]...)
				break
			}
		}
	}
	// ABC deletes only disconnected objects; release any remaining fanins.
	for _, fanin := range nodeFanins(*n) {
		if p.valid(fanin) {
			Aig_ObjDisconnect(p, e)
			break
		}
	}
	p.markDeleted(e.NodeID())
	*n = Node{}
	p.refreshFanouts()
}
func Aig_ObjDelete_rec(p *Graph, e Edge, freeTop int32) {
	p.editDepth++
	// Disconnect updates each circular fanout link. Rebuilding afterward would
	// change the surviving iteration order after a CO pin was patched.
	defer func() { p.editDepth-- }()
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
	if p.node(e).Kind == Output {
		Aig_ObjDisconnect(p, e)
		Aig_ObjDelete(p, e)
	} else {
		p.node(e).Refs--
	}
	p.refreshFanouts()
}
func Aig_ObjPatchFanin0(p *Graph, e, driver Edge) {
	if p.node(e).Kind != Output || !p.valid(driver) {
		panic("aig: output patch requires valid output/driver")
	}
	old := p.node(e).Fanins[0]
	if p.fanouts != nil {
		Aig_ObjRemoveFanoutPin(p, Aig_Regular(old), Aig_Regular(e), 0)
	}
	p.node(old).Refs--
	p.node(e).Fanins[0] = driver
	p.node(e).Level = p.node(driver).Level
	p.node(e).Phase = Aig_ObjPhaseReal(p, driver)
	p.node(driver).Refs++
	if p.fanouts != nil {
		Aig_ObjAddFanoutPin(p, Aig_Regular(driver), Aig_Regular(e), 0)
	}
	if p.node(old).Refs == 0 && p.node(old).Kind != Input && p.node(old).Kind != Constant {
		Aig_ObjDelete_rec(p, Aig_Regular(old), 1)
	}
}
func Aig_ObjPrint(p *Graph, e Edge) {
	if e == InvalidEdge {
		fmt.Print("Object is NULL.")
		return
	}
	if e.Complemented() {
		fmt.Print("Compl ")
		e = Aig_Regular(e)
	}
	n := p.node(e)
	fmt.Printf("Node %4d : ", e.NodeID())
	switch n.Kind {
	case Constant:
		fmt.Print("constant 1")
	case Input:
		fmt.Print("PI")
	case Output:
		f := n.Fanins[0]
		fmt.Printf("PO( %4d%s )", f.NodeID(), aigPrintComplement(f))
	case Buffer:
		f := n.Fanins[0]
		fmt.Printf("BUF( %d%s )", f.NodeID(), aigPrintComplement(f))
	default:
		kind := "AND"
		if n.Kind == ExclusiveOr {
			kind = "XOR"
		}
		a, b := n.Fanins[0], n.Fanins[1]
		fmt.Printf("%s( %4d%s, %4d%s )", kind, a.NodeID(), aigPrintComplement(a), b.NodeID(), aigPrintComplement(b))
	}
	fmt.Printf(" (refs = %3d)", n.Refs)
	if next, ok := p.equivs[e.NodeID()]; ok && next != InvalidEdge {
		fmt.Printf("  { %4d ", e.NodeID())
		seen := map[int]bool{e.NodeID(): true}
		for next != InvalidEdge {
			id := next.NodeID()
			if seen[id] {
				panic("aig: cyclic diagnostic equivalence class")
			}
			seen[id] = true
			suffix := " "
			if (p.node(next).Phase != next.Complemented()) != n.Phase {
				suffix = "'"
			}
			fmt.Printf(" %4d%s", id, suffix)
			next = Aig_ObjEquiv(p, next)
		}
		fmt.Print(" }")
	} else if repr, ok := p.reprs[e.NodeID()]; ok && repr != InvalidEdge {
		fmt.Printf("  class of %d", e.NodeID())
	}
}

func aigPrintComplement(e Edge) string {
	if e.Complemented() {
		return "'"
	}
	return " "
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
	if p.fanouts != nil {
		return p.propagateBufferList(updateLevel)
	}
	var changed int32
	for {
		progress := false
		for id, n := range p.nodes {
			if p.isDeleted(id) {
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
// Aig_ObjReplace follows aigObj.c: the old object keeps its ID and fanouts.
// It becomes a buffer when the replacement is complemented, referenced or not
// a logic node; otherwise it absorbs the replacement's fanins. With fanouts
// active, levels update incrementally and buffers are propagated at once.
func Aig_ObjReplace(p *Graph, old, replacement Edge, updateLevel int32) {
	p.editDepth++
	defer func() { p.editDepth--; p.refreshFanouts() }()
	newR := Aig_Regular(replacement)
	if old.Complemented() || old.NodeID() == newR.NodeID() {
		panic("aig: invalid replacement")
	}
	if k := p.node(old).Kind; k == Input || k == Output || k == Constant {
		panic("aig: cannot replace terminal")
	}
	if k := p.node(newR).Kind; k == Buffer || k == Output {
		panic("aig: invalid replacement object")
	}
	for _, fanin := range nodeFanins(*p.node(newR)) {
		if fanin.NodeID() == old.NodeID() {
			panic("aig: replacement introduces cycle")
		}
	}
	p.node(newR).Refs++
	Aig_ObjDelete_rec(p, old, 0)
	p.node(newR).Refs--
	n := p.node(old)
	if r := p.node(newR); replacement.Complemented() || r.Refs > 0 || r.Kind != And && r.Kind != ExclusiveOr {
		n.Kind = Buffer
		n.Fanins = [2]Edge{InvalidEdge, InvalidEdge}
		Aig_ObjConnect(p, old, replacement, InvalidEdge)
	} else {
		kind, fanins, levelOld := r.Kind, r.Fanins, n.Level
		Aig_ObjDisconnect(p, newR)
		n.Kind = kind
		Aig_ObjConnect(p, old, fanins[0], fanins[1])
		Aig_ObjDelete(p, newR)
		if p.fanouts != nil {
			n.Level = levelOld
			Aig_ManUpdateLevel(p, old)
		}
		// ABC asserts that reverse levels exist; the Go API tolerates callers
		// that request level updates without them.
		if updateLevel != 0 && p.reverseLevels != nil {
			Aig_ObjClearReverseLevel(p, old)
			Aig_ManUpdateReverseLevel(p, old)
		}
	}
	if p.fanouts != nil && p.node(old).Kind == Buffer {
		p.bufs = append(p.bufs, old)
		Aig_ManPropagateBuffers(p, updateLevel)
	}
	if p.fanouts == nil {
		// ABC leaves fanout levels stale here; the Go graph keeps its exact
		// level invariant (checked by Graph.Check) instead.
		p.recomputeLevels()
	}
}
// propagateBufferList is aigObj.c's Aig_ManPropagateBuffers: repeatedly
// take the last pending buffer, walk first fanouts to a non-buffer consumer
// and rebuild that consumer on the buffers' real drivers.
func (p *Graph) propagateBufferList(updateLevel int32) int32 {
	var steps int32
	for ; len(p.bufs) > 0; steps++ {
		e := p.bufs[len(p.bufs)-1]
		for p.node(e).Kind == Buffer {
			fan := Aig_ObjFanout0Int(p, int32(e.NodeID()))
			if fan == 0 {
				// Only bare interface edges (no CO object) use this buffer.
				p.bypassBareOutputs(e)
				break
			}
			e = Aig_ManObj(p, fan>>1)
		}
		if p.node(e).Kind != Buffer {
			Aig_NodeFixBufferFanins(p, e, updateLevel)
		}
		if steps > 1000000 {
			fmt.Println("Error: A cycle is encountered while propagating buffers.")
			break
		}
	}
	return steps
}

// bypassBareOutputs points interface entries that name buffer b at its real
// driver and deletes b once unreferenced.
func (p *Graph) bypassBareOutputs(b Edge) {
	for i, out := range p.outputs {
		if out.NodeID() == b.NodeID() {
			p.outputs[i] = p.real(out)
			p.node(p.outputs[i]).Refs++
			p.node(b).Refs--
		}
	}
	if p.node(b).Refs == 0 {
		Aig_ObjDelete_rec(p, b, 1)
	} else {
		panic("aig: buffer has references but no fanouts")
	}
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
	p.levelModes = nil
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
				level += nodeLevelCost(n.Kind)
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
		if !p.isDeleted(id) {
			visit(id)
		}
	}
}
// Fanouts are maintained incrementally, as in ABC: every connection appends
// its fanout at the tail of the driver's list. Tests set verifyFanouts to
// compare the index against a full rebuild after each edit.
var verifyFanouts bool

func (p *Graph) refreshFanouts() {
	if verifyFanouts && p.editDepth == 0 && p.fanouts != nil {
		p.checkFanouts()
	}
}

// linkFanins records the fanouts of node id's fanins (Aig_ObjConnect).
func (p *Graph) linkFanins(id int) {
	if p.fanouts == nil {
		return
	}
	p.fanouts.ensure(id)
	n := p.nodes[id]
	for pin := 0; pin < fanoutPins(n); pin++ {
		p.fanouts.add(n.Fanins[pin].NodeID(), id, int32(pin))
	}
}

// unlinkFanins removes the fanouts of node id's fanins.
func (p *Graph) unlinkFanins(id int) {
	if p.fanouts == nil {
		return
	}
	n := p.nodes[id]
	for pin := 0; pin < fanoutPins(n); pin++ {
		p.fanouts.remove(n.Fanins[pin].NodeID(), id, int32(pin))
	}
}

// checkFanouts panics unless the index holds exactly the live connections.
func (p *Graph) checkFanouts() {
	want := new(FanoutIndex)
	want.Rebuild(p)
	for id := range p.nodes {
		if p.isDeleted(id) {
			continue
		}
		got, exp := map[int32]int{}, map[int32]int{}
		for _, h := range p.fanouts.Handles(Edge(id) << 1) {
			got[h]++
		}
		for _, h := range want.Handles(Edge(id) << 1) {
			exp[h]++
		}
		if len(got) != len(exp) {
			panic(fmt.Sprintf("aig: incremental fanouts of %d differ: %v vs %v", id, got, exp))
		}
		for h, c := range exp {
			if got[h] != c {
				panic(fmt.Sprintf("aig: incremental fanouts of %d differ: %v vs %v", id, got, exp))
			}
		}
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

// levelNew is Aig_ObjLevelNew for a node about to be connected.
func (p *Graph) levelNew(n Node) int {
	fanins := nodeFanins(n)
	if len(fanins) == 0 {
		return n.Level
	}
	level := p.node(fanins[0]).Level
	if len(fanins) == 2 {
		level = max(level, p.node(fanins[1]).Level) + nodeLevelCost(n.Kind)
	}
	return level
}

// phaseNew is the simulation phase Aig_ObjConnect assigns.
func (p *Graph) phaseNew(n Node) bool {
	fanins := nodeFanins(n)
	if len(fanins) == 0 {
		return n.Kind == Constant
	}
	phase := Aig_ObjPhaseReal(p, fanins[0])
	if len(fanins) == 2 {
		if n.Kind == And {
			phase = phase && Aig_ObjPhaseReal(p, fanins[1])
		} else {
			phase = phase != Aig_ObjPhaseReal(p, fanins[1])
		}
	}
	return phase
}
