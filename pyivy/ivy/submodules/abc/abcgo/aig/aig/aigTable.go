package aig

import "fmt"

func Aig_Hash(node *Node, size int32) uint64 {
	if size <= 0 {
		panic("aig: invalid hash table size")
	}
	var key uint64
	if node.Kind == ExclusiveOr {
		key = 1699
	}
	a, b := node.Fanins[0], node.Fanins[1]
	key ^= uint64(a.NodeID()) * 7937
	key ^= uint64(b.NodeID()) * 2971
	key ^= uint64(a&1) * 911
	key ^= uint64(b&1) * 353
	return key % uint64(size)
}
func tableKey(node *Node) edgePair {
	a, b := node.Fanins[0], node.Fanins[1]
	if a > b {
		a, b = b, a
	}
	return edgePair{a, b}
}
func (p *Graph) nodeTable(kind NodeKind) map[edgePair]Edge {
	if kind == And {
		if p.hash == nil {
			p.hash = make(map[edgePair]Edge)
		}
		return p.hash
	}
	if kind == ExclusiveOr {
		if p.xorHash == nil {
			p.xorHash = make(map[edgePair]Edge)
		}
		return p.xorHash
	}
	panic("aig: only AND/XOR nodes are hashable")
}

// Aig_TableFind returns the existing entry and presence. Map storage replaces
// C's mutable pointer into a bucket chain; insert/delete are explicit operations.
func Aig_TableFind(p *Graph, node *Node) (Edge, bool) {
	e, ok := p.nodeTable(node.Kind)[tableKey(node)]
	return e, ok
}
func Aig_TableResize(p *Graph) {
	old := p.hash
	p.hash = make(map[edgePair]Edge, len(old))
	for key, e := range old {
		p.hash[key] = e
	}
	old = p.xorHash
	p.xorHash = make(map[edgePair]Edge, len(old))
	for key, e := range old {
		p.xorHash[key] = e
	}
}
func Aig_TableLookup(p *Graph, ghost *Node) Edge {
	if e, ok := Aig_TableFind(p, ghost); ok {
		return e
	}
	return InvalidEdge
}
func Aig_TableLookupTwo(p *Graph, a, b Edge) Edge {
	if e, ok := Aig_TableLookupInt(p, a, b); ok {
		return e
	}
	return InvalidEdge
}
func Aig_TableInsert(p *Graph, e Edge) {
	n := p.node(e)
	table := p.nodeTable(n.Kind)
	key := tableKey(n)
	if old, exists := table[key]; exists && old != e {
		panic("aig: duplicate hash entry")
	}
	table[key] = Aig_Regular(e)
}
func Aig_TableDelete(p *Graph, e Edge) {
	n := p.node(e)
	table := p.nodeTable(n.Kind)
	key := tableKey(n)
	if old, exists := table[key]; !exists || old != Aig_Regular(e) {
		panic("aig: missing hash entry")
	}
	delete(table, key)
}
func Aig_TableCountEntries(p *Graph) int32 { return int32(len(p.hash) + len(p.xorHash)) }
func Aig_TableProfile(p *Graph) {
	fmt.Printf("AIG hash: %d AND entries, %d XOR entries\n", len(p.hash), len(p.xorHash))
}
func Aig_TableClear(p *Graph) { clear(p.hash); clear(p.xorHash) }
