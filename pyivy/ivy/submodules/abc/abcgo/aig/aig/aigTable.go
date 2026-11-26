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
func (p *Graph) nodeTable(kind NodeKind) *strashTable {
	if kind == And {
		return &p.hash
	}
	if kind == ExclusiveOr {
		return &p.xorHash
	}
	panic("aig: only AND/XOR nodes are hashable")
}

// Aig_TableFind returns the existing entry and presence. Map storage replaces
// C's mutable pointer into a bucket chain; insert/delete are explicit operations.
func Aig_TableFind(p *Graph, node *Node) (Edge, bool) {
	return p.nodeTable(node.Kind).get(tableKey(node))
}

// Aig_TableResize is a no-op: the table grows itself as entries are added.
func Aig_TableResize(p *Graph) {}

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
	if old, exists := table.get(key); exists && old != e {
		panic("aig: duplicate hash entry")
	}
	table.set(key, Aig_Regular(e))
}
func Aig_TableDelete(p *Graph, e Edge) {
	n := p.node(e)
	table := p.nodeTable(n.Kind)
	key := tableKey(n)
	if old, exists := table.get(key); !exists || old != Aig_Regular(e) {
		panic("aig: missing hash entry")
	}
	table.del(key)
}
func Aig_TableCountEntries(p *Graph) int32 { return int32(p.hash.len() + p.xorHash.len()) }
func Aig_TableProfile(p *Graph) {
	fmt.Printf("AIG hash: %d AND entries, %d XOR entries\n", p.hash.len(), p.xorHash.len())
}
func Aig_TableClear(p *Graph) { p.hash.clear(); p.xorHash.clear() }
