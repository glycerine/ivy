package aig

import "sort"

func satPartitionNode(kind NodeKind) bool {
	return kind == And || kind == ExclusiveOr || kind == Buffer
}
func Aig_ManPartitionMonolithic(g *Graph) []int32 { return make([]int32, len(g.nodes)) }
func Aig_ManPartitionLevelized(g *Graph, size int32) []int32 {
	if size <= 0 {
		panic("aig: partition size must be positive")
	}
	result := make([]int32, len(g.nodes))
	// C enumerates every live object in reverse level/ID order. Terminals
	// consume positions too, even though only gates are split into partitions.
	var ids []int
	for id := range g.nodes {
		if !g.deleted[id] {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := ids[i], ids[j]
		if g.nodes[a].Level != g.nodes[b].Level {
			return g.nodes[a].Level > g.nodes[b].Level
		}
		return a > b
	})
	for i, id := range ids {
		result[id] = int32(i) / size
	}
	return result
}
func Aig_ManPartitionDfs(g *Graph, size, preorder int32) []int32 {
	if size <= 0 {
		panic("aig: partition size must be positive")
	}
	// Use the source traversal setup and recursive helpers: a choice is visited
	// in postorder before ordinary fanins even during preorder partitioning.
	// Unlike the strict whole-manager DFS entry points, permit dangling gates
	// and unary buffers in this typed-graph extension.
	Aig_ManIncrementTravId(g)
	g.node(g.Const1()).TravID = g.travID
	for _, e := range g.inputs {
		g.node(e).TravID = g.travID
	}
	var visited, nodes []Edge
	for _, e := range g.outputs {
		root := Aig_Regular(g.outputDriver(e))
		if preorder != 0 {
			Aig_ManDfsPreorder_rec(g, root, &visited)
		} else {
			Aig_ManDfs_rec(g, root, &visited)
		}
	}
	for _, e := range visited {
		if satPartitionNode(g.node(e).Kind) {
			nodes = append(nodes, e)
		}
	}
	if preorder == 0 {
		for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	}
	result := make([]int32, len(g.nodes))
	for i, e := range nodes {
		result[e.NodeID()] = int32(i) / size
	}
	return result
}
func Aig_ManPartSplitOne_rec(destination, source *Graph, node Edge, pio *[]int32) {
	node = Aig_Regular(node)
	n := source.node(node)
	if n.Kind == Output {
		Aig_ManPartSplitOne_rec(destination, source, n.Fanins[0], pio)
		source.copies[node.NodeID()] = Aig_ObjRealCopy(source, n.Fanins[0])
		return
	}
	if n.TravID != source.travID {
		n.TravID = source.travID
		copy := destination.NewInput()
		destination.node(copy).MarkA = n.MarkA
		destination.node(copy).MarkB = n.MarkB
		source.copies[node.NodeID()] = copy
		*pio = append(*pio, int32(node.NodeID()))
		return
	}
	if _, exists := source.copies[node.NodeID()]; exists {
		return
	}
	Aig_ManPartSplitOne_rec(destination, source, n.Fanins[0], pio)
	a := Aig_ObjRealCopy(source, n.Fanins[0])
	var copy Edge
	if n.Kind == Buffer {
		copy = a
	} else {
		Aig_ManPartSplitOne_rec(destination, source, n.Fanins[1], pio)
		b := Aig_ObjRealCopy(source, n.Fanins[1])
		if n.Kind == ExclusiveOr {
			copy = destination.NewXor(a, b)
		} else {
			copy = destination.AND(a, b)
		}
	}
	source.copies[node.NodeID()] = copy
}
func Aig_ManPartSplitOne(source *Graph, nodes []Edge) (*Graph, []int32) {
	Aig_ManIncrementTravId(source)
	source.copies = map[int]Edge{}
	for _, node := range nodes {
		if node.Complemented() || !satPartitionNode(source.node(node).Kind) {
			panic("aig: invalid partition node")
		}
		if source.node(node).TravID == source.travID {
			panic("aig: duplicate partition node")
		}
		source.node(node).TravID = source.travID
	}
	// Count crossings in the original graph. A buffer may collapse into its
	// driver and simplify a later AND, so copied reference counts cannot tell
	// which original object IDs still need boundary terminals.
	externalRefs := make(map[int]int, len(nodes))
	for _, e := range nodes {
		externalRefs[e.NodeID()] = source.node(e).Refs
	}
	for _, e := range nodes {
		for _, fanin := range nodeFanins(*source.node(e)) {
			if _, selected := externalRefs[fanin.NodeID()]; selected {
				externalRefs[fanin.NodeID()]--
			}
		}
	}
	destination := NewGraph()
	var pio []int32
	for _, node := range nodes {
		Aig_ManPartSplitOne_rec(destination, source, node, &pio)
	}
	for _, node := range nodes {
		copy := source.copies[node.NodeID()]
		if externalRefs[node.NodeID()] > 0 {
			Aig_ObjCreateCo(destination, copy)
			pio = append(pio, int32(node.NodeID()))
		}
	}
	return destination, pio
}

// PIO mappings contain original object IDs. Output groups contain original
// driver literals, preserving output complement phases without tagged pointers.
func Aig_ManPartSplit(source *Graph, node2part []int32) (parts []*Graph, pio [][]int32, partOutputs [][]Edge) {
	if len(node2part) < len(source.nodes) {
		panic("aig: incomplete partition map")
	}
	groups := map[int32][]Edge{}
	for id, n := range source.nodes {
		if source.deleted[id] || !satPartitionNode(n.Kind) {
			continue
		}
		label := node2part[id]
		if label < 0 {
			panic("aig: negative partition label")
		}
		groups[label] = append(groups[label], Edge(id)<<1)
	}
	if len(groups) == 0 {
		part := NewGraph()
		source.copies = map[int]Edge{0: 0}
		for _, input := range source.inputs {
			source.copies[input.NodeID()] = part.NewInput()
		}
		var mapping []int32
		for _, input := range source.inputs {
			mapping = append(mapping, int32(input.NodeID()))
		}
		for _, output := range source.outputs {
			driver := satOutputDriver(source, output)
			part.AddOutput(dupEdge(part, source, Aig_Regular(driver)))
			mapping = append(mapping, int32(driver.NodeID()))
		}
		return []*Graph{part}, [][]int32{mapping}, [][]Edge{append([]Edge(nil), outputDrivers(source)...)}
	}
	// Retain pre-existing client marks rather than using persistent graph flags
	// as temporary partitioning state.
	type marks struct{ a, b bool }
	saved := make([]marks, len(source.nodes))
	for id := range source.nodes {
		n := &source.nodes[id]
		saved[id] = marks{n.MarkA, n.MarkB}
		n.MarkA, n.MarkB = false, false
	}
	defer func() {
		for id, m := range saved {
			source.nodes[id].MarkA, source.nodes[id].MarkB = m.a, m.b
		}
	}()
	for _, output := range source.outputs {
		driver := satOutputDriver(source, output)
		n := source.node(driver)
		if n.Kind != Input {
			continue
		}
		if driver.Complemented() {
			n.MarkA = true
		} else {
			n.MarkB = true
		}
	}
	labels := make([]int32, 0, len(groups))
	for label := range groups {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool { return labels[i] < labels[j] })
	compact := map[int32]int{}
	for _, label := range labels {
		part, mapping := Aig_ManPartSplitOne(source, groups[label])
		compact[label] = len(parts)
		parts = append(parts, part)
		pio = append(pio, mapping)
		partOutputs = append(partOutputs, nil)
	}
	for _, output := range source.outputs {
		driver := satOutputDriver(source, output)
		if !satPartitionNode(source.node(driver).Kind) {
			continue
		}
		part := compact[node2part[driver.NodeID()]]
		partOutputs[part] = append(partOutputs[part], driver)
	}
	return
}

// Output objects are transparent in the Boolean network, including COs that
// appear as fanins after graph transformations.
func satOutputDriver(g *Graph, edge Edge) Edge {
	for g.node(edge).Kind == Output {
		edge = dupPhase(g.node(edge).Fanins[0], edge.Complemented())
	}
	return edge
}
func outputDrivers(g *Graph) []Edge {
	result := make([]Edge, len(g.outputs))
	for i, output := range g.outputs {
		result[i] = satOutputDriver(g, output)
	}
	return result
}
func Aig_ManPartResetNodePolarity(part *Graph) {
	for id := range part.nodes {
		part.nodes[id].Phase = false
	}
}
func Aig_ManPartSetNodePolarity(source, part *Graph, pio []int32) {
	if len(pio) != len(part.inputs)+len(part.outputs) {
		panic("aig: invalid PIO mapping")
	}
	part.nodes[0].Phase = true
	for i, input := range part.inputs {
		part.node(input).Phase = source.nodes[pio[i]].Phase
	}
	state := make([]uint8, len(part.nodes))
	state[0] = 2
	for _, input := range part.inputs {
		state[input.NodeID()] = 2
	}
	var phase func(Edge) bool
	phase = func(edge Edge) bool {
		id := edge.NodeID()
		if state[id] == 1 {
			panic("aig: cyclic polarity graph")
		}
		if state[id] != 2 {
			state[id] = 1
			n := part.node(edge)
			switch n.Kind {
			case And:
				n.Phase = phase(n.Fanins[0]) && phase(n.Fanins[1])
			case ExclusiveOr:
				n.Phase = phase(n.Fanins[0]) != phase(n.Fanins[1])
			case Buffer, Output:
				n.Phase = phase(n.Fanins[0])
			default:
				panic("aig: unknown polarity node")
			}
			state[id] = 2
		}
		return part.nodes[id].Phase != edge.Complemented()
	}
	for id := range part.nodes {
		if !part.deleted[id] {
			phase(Edge(id) << 1)
		}
	}
	for i, output := range part.outputs {
		if phase(output) != source.nodes[pio[len(part.inputs)+i]].Phase {
			panic("aig: partition output phase mismatch")
		}
	}
}
