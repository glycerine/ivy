package aig

import (
	"slices"
	"sort"
)

// Duplication uses graph-local edge maps rather than mutating node payloads.
// Source copy maps remain available to callers, as pData did in ABC.
func dupMetadata(dst, src *Graph) {
	dst.Name, dst.Spec = src.Name, src.Spec
	dst.Time2Quit = src.Time2Quit
	dst.RetimingChanged = src.RetimingChanged
	dst.nRegs, dst.nConstrs = src.nRegs, src.nConstrs
	dst.Assertions, dst.BarrierBuffers = src.Assertions, src.BarrierBuffers
	dst.PartitionInputs = slices.Clone(src.PartitionInputs)
	dst.PartitionOutputs = slices.Clone(src.PartitionOutputs)
	dst.RetimingChanged = src.RetimingChanged
	dst.FlopNumbers = slices.Clone(src.FlopNumbers)
	dst.FlopReprs = slices.Clone(src.FlopReprs)
	dst.ClockDomains = slices.Clone(src.ClockDomains)
	for i, domain := range dst.ClockDomains {
		dst.ClockDomains[i] = slices.Clone(domain)
	}
	dst.OneHotGroups = slices.Clone(src.OneHotGroups)
	for i, group := range dst.OneHotGroups {
		dst.OneHotGroups[i] = slices.Clone(group)
	}
	dst.OneHotInputGroups = slices.Clone(src.OneHotInputGroups)
	for i, group := range dst.OneHotInputGroups {
		dst.OneHotInputGroups[i] = slices.Clone(group)
	}
	if src.Timing != nil {
		if c, ok := src.Timing.(interface{ CloneTiming() any }); ok {
			dst.Timing = c.CloneTiming()
		} else {
			panic("aig: timing metadata must implement CloneTiming")
		}
	}
	if src.History != nil {
		if c, ok := src.History.(interface{ CloneHistory() any }); ok {
			dst.History = c.CloneHistory()
		} else {
			panic("aig: history metadata must implement CloneHistory")
		}
	}
}
func dupStart(src *Graph) *Graph {
	dst := NewGraph()
	dupMetadata(dst, src)
	src.copies = map[int]Edge{0: 0}
	for _, e := range src.inputs {
		x := dst.NewInput()
		dst.node(x).Level = src.node(e).Level
		src.copies[e.NodeID()] = x
	}
	return dst
}
func dupPhase(e Edge, flip bool) Edge {
	if flip {
		return e.Not()
	}
	return e
}
func dupVisit(dst, src *Graph, e Edge, choices, representatives bool, active map[int]bool) Edge {
	id := e.NodeID()
	if !src.valid(e) {
		panic("aig: invalid duplication edge")
	}
	if representatives {
		if rep, ok := src.reprs[id]; ok && rep != InvalidEdge && rep.NodeID() != id {
			if active[id] {
				panic("aig: cyclic representative")
			}
			active[id] = true
			result := dupVisit(dst, src, rep, choices, representatives, active)
			delete(active, id)
			result = dupPhase(result, src.node(e).Phase != src.node(rep).Phase)
			src.copies[id] = result
			return dupPhase(result, e.Complemented())
		}
	}
	if result, ok := src.copies[id]; ok {
		return dupPhase(result, e.Complemented())
	}
	if active[id] {
		panic("aig: cyclic duplication graph")
	}
	active[id] = true
	defer delete(active, id)
	n := src.node(e)
	var equivalent Edge = InvalidEdge
	if choices {
		if other, ok := src.equivs[id]; ok && other != InvalidEdge && other.NodeID() != id {
			equivalent = dupVisit(dst, src, other, choices, representatives, active)
		}
	}
	var result Edge
	switch n.Kind {
	case Constant:
		result = dst.Const1()
	case Input:
		if choices {
			return InvalidEdge
		}
		panic("aig: unbound input in duplication cone")
	case Buffer, Output:
		result = dupVisit(dst, src, n.Fanins[0], choices, representatives, active)
	case And:
		a := dupVisit(dst, src, n.Fanins[0], choices, representatives, active)
		if a == InvalidEdge {
			return InvalidEdge
		}
		b := dupVisit(dst, src, n.Fanins[1], choices, representatives, active)
		if b == InvalidEdge {
			return InvalidEdge
		}
		result = dst.AND(a, b)
	case ExclusiveOr:
		a := dupVisit(dst, src, n.Fanins[0], choices, representatives, active)
		if a == InvalidEdge {
			return InvalidEdge
		}
		b := dupVisit(dst, src, n.Fanins[1], choices, representatives, active)
		if b == InvalidEdge {
			return InvalidEdge
		}
		result = dst.NewXor(a, b)
	default:
		panic("aig: unsupported duplication node")
	}
	if result == InvalidEdge {
		return InvalidEdge
	}
	src.copies[id] = result
	if equivalent != InvalidEdge && equivalent.NodeID() != result.NodeID() {
		if dst.equivs == nil {
			dst.equivs = map[int]Edge{}
		}
		if dst.reprs == nil {
			dst.reprs = map[int]Edge{}
		}
		dst.equivs[result.NodeID()] = Aig_Regular(equivalent)
		dst.reprs[equivalent.NodeID()] = Aig_Regular(result)
	}
	return dupPhase(result, e.Complemented())
}
func dupEdge(dst, src *Graph, e Edge) Edge {
	return dupVisit(dst, src, e, false, false, map[int]bool{})
}
func dupOutputDriver(src *Graph, e Edge) Edge {
	if src.node(e).Kind == Output {
		return dupPhase(src.node(e).Fanins[0], e.Complemented())
	}
	return e
}
func dupOutputs(dst, src *Graph, outputs []Edge, choices, reps bool) {
	for _, e := range outputs {
		result := dupVisit(dst, src, dupOutputDriver(src, e), choices, reps, map[int]bool{})
		if result == InvalidEdge {
			panic("aig: output has unavailable input")
		}
		if src.node(e).Kind == Output {
			src.copies[e.NodeID()] = Aig_ObjCreateCo(dst, result)
		} else {
			dst.AddOutput(result)
		}
	}
}
func Aig_ManDupSimple(src *Graph) *Graph {
	dst := dupStart(src)
	for id, n := range src.nodes {
		if !src.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			dupEdge(dst, src, Edge(id)<<1)
		}
	}
	dupOutputs(dst, src, src.outputs, false, false)
	return dst
}
func Aig_ManDupSimpleWithHints(src *Graph, hints []int32) *Graph {
	dst := dupStart(src)
	for id, n := range src.nodes {
		if src.deleted[id] || n.Kind == Constant || n.Kind == Output {
			continue
		}
		if n.Kind != Input {
			dupEdge(dst, src, Edge(id)<<1)
		}
		if id < len(hints) && (hints[id] == 0 || hints[id] == 1) {
			src.copies[id] = Edge(hints[id])
		}
	}
	dupOutputs(dst, src, src.outputs, false, false)
	Aig_ManCleanup(dst)
	return dst
}
func Aig_ManDupSimpleDfs_rec(dst, src *Graph, e Edge) Edge { return dupEdge(dst, src, e) }
func Aig_ManDupSimpleDfs(src *Graph) *Graph {
	dst := dupStart(src)
	dupOutputs(dst, src, src.outputs, false, false)
	return dst
}
func Aig_ManDupSimpleDfsPart(src *Graph, inputs, outputs []Edge) *Graph {
	dst := NewGraph()
	src.copies = map[int]Edge{0: 0}
	seenInputs := map[int]bool{}
	for _, e := range inputs {
		if seenInputs[e.NodeID()] {
			panic("aig: duplicate partial CI")
		}
		seenInputs[e.NodeID()] = true
		if src.node(e).Kind != Input {
			panic("aig: partial input must be CI")
		}
		src.copies[e.NodeID()] = dst.NewInput()
	}
	dupOutputs(dst, src, outputs, false, false)
	return dst
}
func Aig_ManDupOrdered(src *Graph) *Graph {
	dst := dupStart(src)
	for id, n := range src.nodes {
		if src.deleted[id] {
			continue
		}
		e := Edge(id) << 1
		switch n.Kind {
		case And, ExclusiveOr, Buffer:
			dupEdge(dst, src, e)
		case Output:
			src.copies[id] = Aig_ObjCreateCo(dst, dupEdge(dst, src, n.Fanins[0]))
		}
	}
	dst.outputs = nil
	for _, out := range src.outputs {
		if src.node(out).Kind == Output {
			dst.outputs = append(dst.outputs, src.copies[out.NodeID()])
		} else {
			dst.outputs = append(dst.outputs, dupEdge(dst, src, out))
		}
	}
	dst.recountReferences()
	Aig_ManCleanup(dst)
	dst.refreshFanouts()
	return dst
}
func Aig_ManDupCof(src *Graph, index, value int32) *Graph {
	if index < 0 || int(index) >= len(src.inputs) || (value != 0 && value != 1) {
		panic("aig: invalid cofactor")
	}
	dst := NewGraph()
	dupMetadata(dst, src)
	src.copies = map[int]Edge{0: 0}
	for i, e := range src.inputs {
		if i == int(index) {
			src.copies[e.NodeID()] = Edge(1 - value)
		} else {
			x := dst.NewInput()
			dst.node(x).Level = src.node(e).Level
			src.copies[e.NodeID()] = x
		}
	}
	dupOutputs(dst, src, src.outputs, false, false)
	Aig_ManCleanup(dst)
	return dst
}
func Aig_ManDupTrim(src *Graph) *Graph {
	dst := NewGraph()
	dupMetadata(dst, src)
	src.copies = map[int]Edge{0: 0}
	for i, e := range src.inputs {
		if src.node(e).Refs > 0 || i >= len(src.inputs)-int(src.nRegs) {
			x := dst.NewInput()
			dst.node(x).Level = src.node(e).Level
			src.copies[e.NodeID()] = x
		}
	}
	for id, n := range src.nodes {
		if !src.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			dupEdge(dst, src, Edge(id)<<1)
		}
	}
	dupOutputs(dst, src, src.outputs, false, false)
	Aig_ManCleanup(dst)
	return dst
}
func Aig_ManDupExor(src *Graph) *Graph {
	dst := dupStart(src)
	for id, n := range src.nodes {
		if src.deleted[id] || (n.Kind != And && n.Kind != ExclusiveOr && n.Kind != Buffer) {
			continue
		}
		if n.Kind == And && n.Fanins[0].Complemented() && n.Fanins[1].Complemented() {
			a, b := src.node(n.Fanins[0]), src.node(n.Fanins[1])
			if a.Kind == And && b.Kind == And && ((a.Fanins[0] == b.Fanins[0].Not() && a.Fanins[1] == b.Fanins[1].Not()) || (a.Fanins[0] == b.Fanins[1].Not() && a.Fanins[1] == b.Fanins[0].Not())) {
				src.copies[id] = dst.NewXor(dupEdge(dst, src, a.Fanins[0]), dupEdge(dst, src, a.Fanins[1]))
				continue
			}
		}
		dupEdge(dst, src, Edge(id)<<1)
	}
	dupOutputs(dst, src, src.outputs, false, false)
	Aig_ManCleanup(dst)
	return dst
}
func Aig_ManDupDfs_rec(dst, src *Graph, e Edge) Edge {
	return dupVisit(dst, src, e, true, false, map[int]bool{})
}
func Aig_ManDupDfs(src *Graph) *Graph {
	dst := dupStart(src)
	dupOutputs(dst, src, src.outputs, true, false)
	if len(src.equivs) == 0 && len(src.reprs) == 0 {
		Aig_ManCleanup(dst)
	}
	return dst
}

// PIO describes a logical terminal even when outputs are stored as driver edges.
type PIO struct {
	Input bool
	Index int
}

func Aig_ManOrderPios(src, order *Graph) []PIO {
	if len(src.inputs) != len(order.inputs) || len(src.outputs) != len(order.outputs) {
		panic("aig: PIO count mismatch")
	}
	inputs := map[int]int{}
	outputs := map[int]int{}
	for i, e := range order.inputs {
		inputs[e.NodeID()] = i
	}
	for i, e := range order.outputs {
		if order.node(e).Kind == Output {
			outputs[e.NodeID()] = i
		}
	}
	var result []PIO
	used := map[int]bool{}
	for id := range order.nodes {
		if i, ok := inputs[id]; ok {
			result = append(result, PIO{true, i})
		}
		if i, ok := outputs[id]; ok {
			result = append(result, PIO{false, i})
			used[i] = true
		}
	}
	for i := range order.outputs {
		if !used[i] {
			result = append(result, PIO{false, i})
		}
	}
	return result
}
func Aig_ManDupDfsGuided_rec(dst, src *Graph, e Edge) Edge {
	if src.node(e).Kind == Input {
		if copy, ok := src.copies[e.NodeID()]; ok {
			return dupPhase(copy, e.Complemented())
		}
		return InvalidEdge
	}
	// A guided cone may depend on a CI not yet introduced by the PIO ordering.
	seen := map[int]bool{}
	var available func(Edge) bool
	available = func(x Edge) bool {
		if _, ok := src.copies[x.NodeID()]; ok {
			return true
		}
		if seen[x.NodeID()] {
			return true
		}
		seen[x.NodeID()] = true
		n := src.node(x)
		if n.Kind == Input {
			return false
		}
		if n.Kind == Constant {
			return true
		}
		if !available(n.Fanins[0]) {
			return false
		}
		return n.Kind == Buffer || n.Kind == Output || available(n.Fanins[1])
	}
	if !available(e) {
		return InvalidEdge
	}
	return dupVisit(dst, src, e, true, false, map[int]bool{})
}
func Aig_ManDupDfsGuided(src *Graph, order []PIO) *Graph {
	dst := NewGraph()
	dupMetadata(dst, src)
	src.copies = map[int]Edge{0: 0}
	seenInputs, seenOutputs := map[int]bool{}, map[int]bool{}
	for _, pio := range order {
		if pio.Input {
			if pio.Index < 0 || pio.Index >= len(src.inputs) || seenInputs[pio.Index] {
				panic("aig: invalid guided CI order")
			}
			seenInputs[pio.Index] = true
			e := src.inputs[pio.Index]
			x := dst.NewInput()
			dst.node(x).Level = src.node(e).Level
			src.copies[e.NodeID()] = x
		} else {
			if pio.Index < 0 || pio.Index >= len(src.outputs) || seenOutputs[pio.Index] {
				panic("aig: invalid guided CO order")
			}
			seenOutputs[pio.Index] = true
			e := Aig_ManDupDfsGuided_rec(dst, src, dupOutputDriver(src, src.outputs[pio.Index]))
			if e == InvalidEdge {
				panic("aig: output precedes required CI")
			}
			dst.AddOutput(e)
		}
	}
	if len(dst.inputs) != len(src.inputs) || len(dst.outputs) != len(src.outputs) {
		panic("aig: incomplete PIO order")
	}
	return dst
}
func Aig_ManDupLevelized(src *Graph) *Graph {
	dst := dupStart(src)
	var nodes []int
	for id, n := range src.nodes {
		if !src.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			nodes = append(nodes, id)
		}
	}
	sort.SliceStable(nodes, func(i, j int) bool { return src.nodes[nodes[i]].Level < src.nodes[nodes[j]].Level })
	for _, id := range nodes {
		dupEdge(dst, src, Edge(id)<<1)
	}
	dupOutputs(dst, src, src.outputs, false, false)
	return dst
}
func Aig_ManDupWithoutPos(src *Graph) *Graph {
	dst := dupStart(src)
	dst.nRegs = 0
	dst.nConstrs = 0
	for id, n := range src.nodes {
		if !src.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			dupEdge(dst, src, Edge(id)<<1)
		}
	}
	return dst
}
func Aig_ManDupFlopsOnly(src *Graph) *Graph {
	dst := Aig_ManDupWithoutPos(src)
	dupOutputs(dst, src, src.outputs[len(src.outputs)-int(src.nRegs):], false, false)
	dst.nRegs = src.nRegs
	Aig_ManCleanup(dst)
	return dst
}
func Aig_ObjGetRepres(src *Graph, e Edge) Edge {
	if rep, ok := src.reprs[e.NodeID()]; ok && rep != InvalidEdge {
		copy, found := src.copies[rep.NodeID()]
		if !found {
			return InvalidEdge
		}
		phaseDiff := src.node(e).Phase != src.node(rep).Phase
		phaseDiff = phaseDiff != e.Complemented()
		return dupPhase(copy, phaseDiff != rep.Complemented())
	}
	if copy, ok := src.copies[e.NodeID()]; ok {
		return dupPhase(copy, e.Complemented())
	}
	return InvalidEdge
}
func Aig_ObjChild0Repres(src *Graph, e Edge) Edge {
	return Aig_ObjGetRepres(src, src.node(e).Fanins[0])
}
func Aig_ObjChild1Repres(src *Graph, e Edge) Edge {
	return Aig_ObjGetRepres(src, src.node(e).Fanins[1])
}
func Aig_ManDupRepres(src *Graph) *Graph {
	dst := dupStart(src)
	for id, n := range src.nodes {
		if !src.deleted[id] && (n.Kind == Input || n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			dupVisit(dst, src, Edge(id)<<1, false, true, map[int]bool{})
		}
	}
	dupOutputs(dst, src, src.outputs, false, true)
	return dst
}
func Aig_ManDupRepres_rec(dst, src *Graph, e Edge) Edge {
	return dupVisit(dst, src, e, false, true, map[int]bool{})
}
func Aig_ManDupRepresDfs(src *Graph) *Graph {
	dst := dupStart(src)
	dupOutputs(dst, src, src.outputs, false, true)
	return dst
}
func Aig_ManCreateMiter(a, b *Graph, operation int32) *Graph {
	if a.nRegs != 0 || b.nRegs != 0 || len(a.inputs) != len(b.inputs) || len(a.outputs) != 1 || len(b.outputs) != 1 {
		panic("aig: invalid miter operands")
	}
	dst := dupStart(a)
	x := dupEdge(dst, a, dupOutputDriver(a, a.outputs[0]))
	b.copies = map[int]Edge{0: 0}
	for i, e := range b.inputs {
		b.copies[e.NodeID()] = dst.inputs[i]
	}
	y := dupEdge(dst, b, dupOutputDriver(b, b.outputs[0]))
	var out Edge
	switch operation {
	case 0:
		out = dst.XOR(x, y)
	case 1:
		out = dst.AND(x, y.Not())
	case 2:
		out = dst.OR(x, y)
	case 3:
		out = dst.AND(x, y)
	default:
		panic("aig: invalid miter operation")
	}
	dst.AddOutput(out)
	Aig_ManCleanup(dst)
	return dst
}
func dupSequentialOutputs(src *Graph, outputs []Edge, addRegs int32) *Graph {
	dst := dupStart(src)
	dst.nConstrs = 0
	if addRegs == 0 {
		dst.nRegs = 0
	}
	dupOutputs(dst, src, outputs, false, false)
	if addRegs != 0 {
		dupOutputs(dst, src, src.outputs[len(src.outputs)-int(src.nRegs):], false, false)
	}
	Aig_ManCleanup(dst)
	return dst
}
func Aig_ManDupOrpos(src *Graph, addRegs int32) *Graph {
	if src.nConstrs > 0 {
		return nil
	}
	dst := dupStart(src)
	dst.nConstrs = 0
	if addRegs == 0 {
		dst.nRegs = 0
	}
	out := dst.Const0()
	for _, e := range src.outputs[:len(src.outputs)-int(src.nRegs)] {
		out = dst.OR(out, dupEdge(dst, src, dupOutputDriver(src, e)))
	}
	dst.AddOutput(out)
	if addRegs != 0 {
		dupOutputs(dst, src, src.outputs[len(src.outputs)-int(src.nRegs):], false, false)
	}
	Aig_ManCleanup(dst)
	return dst
}
func Aig_ManDupOneOutput(src *Graph, index, addRegs int32) *Graph {
	if index < 0 || int(index) >= len(src.outputs)-int(src.nRegs) {
		panic("aig: PO index")
	}
	return dupSequentialOutputs(src, []Edge{src.outputs[index]}, addRegs)
}
func Aig_ManDupUnsolvedOutputs(src *Graph, addRegs int32) *Graph {
	if src.nConstrs > 0 {
		return nil
	}
	var outputs []Edge
	for _, e := range src.outputs[:len(src.outputs)-int(src.nRegs)] {
		if dupOutputDriver(src, e).NodeID() != 0 {
			outputs = append(outputs, e)
		}
	}
	return dupSequentialOutputs(src, outputs, addRegs)
}
func Aig_ManDupArray(graphs []*Graph) *Graph {
	if len(graphs) == 0 {
		return nil
	}
	dst := NewGraph()
	dst.Name = graphs[0].Name
	for range graphs[0].inputs {
		dst.NewInput()
	}
	for _, src := range graphs {
		if src.nRegs != 0 || len(src.inputs) != len(dst.inputs) {
			panic("aig: incompatible graph array")
		}
		src.copies = map[int]Edge{0: 0}
		for i, e := range src.inputs {
			src.copies[e.NodeID()] = dst.inputs[i]
		}
		dupOutputs(dst, src, src.outputs, false, false)
	}
	return dst
}
func Aig_ManDupNodes(src *Graph, roots []Edge) *Graph {
	if len(roots) == 0 {
		return nil
	}
	dst := NewGraph()
	dst.Name = src.Name
	src.copies = map[int]Edge{0: 0}
	visited := map[int]bool{}
	var leaves []Edge
	var visit func(Edge)
	visit = func(e Edge) {
		if visited[e.NodeID()] {
			return
		}
		visited[e.NodeID()] = true
		n := src.node(e)
		if n.Kind == Input {
			leaves = append(leaves, Aig_Regular(e))
			return
		}
		if n.Kind == Constant {
			return
		}
		visit(n.Fanins[0])
		if n.Kind == And || n.Kind == ExclusiveOr {
			visit(n.Fanins[1])
		}
	}
	for _, e := range roots {
		visit(dupOutputDriver(src, e))
	}
	for _, e := range leaves {
		src.copies[e.NodeID()] = dst.NewInput()
	}
	dupOutputs(dst, src, roots, false, false)
	return dst
}
