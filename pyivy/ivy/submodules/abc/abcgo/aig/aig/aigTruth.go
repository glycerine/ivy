package aig

import "fmt"

// Aig_ManCutTruthOne combines the already computed child truth tables. Bit i
// corresponds to assignment i, with leaf zero in the least-significant bit.
func Aig_ManCutTruthOne(p *Graph, node Edge, truth []uint32, children map[int][]uint32) []uint32 {
	n := p.nodes[node.NodeID()]
	if n.Kind != And && n.Kind != ExclusiveOr {
		panic("aig: cut truth requires an AND node")
	}
	a, b := children[n.Fanins[0].NodeID()], children[n.Fanins[1].NodeID()]
	if len(a) < len(truth) || len(b) < len(truth) {
		panic("aig: missing child truth table")
	}
	for i := range truth {
		x, y := a[i], b[i]
		if n.Fanins[0].Complemented() {
			x = ^x
		}
		if n.Fanins[1].Complemented() {
			y = ^y
		}
		if n.Kind == ExclusiveOr {
			truth[i] = x ^ y
		} else {
			truth[i] = x & y
		}
	}
	return truth
}

// Aig_ManCutTruth computes a truth table over cut leaves, without storing
// temporary truth pointers in graph nodes. Inputs outside the cut are errors.
func Aig_ManCutTruth(p *Graph, root Edge, leaves []Edge) ([]uint32, error) {
	if !p.valid(root) {
		return nil, fmt.Errorf("aig: root outside graph")
	}
	if len(leaves) > 20 {
		return nil, fmt.Errorf("aig: cut truth exceeds 20 leaves")
	}
	bits := 1 << len(leaves)
	words := (bits + 31) / 32
	values := make(map[int][]uint32)
	for i, e := range leaves {
		if !p.valid(e) || e.Complemented() {
			return nil, fmt.Errorf("aig: cut leaves must be regular graph nodes")
		}
		if _, exists := values[e.NodeID()]; exists {
			return nil, fmt.Errorf("aig: duplicate cut leaf")
		}
		table := make([]uint32, words)
		for bit := 0; bit < words*32; bit++ {
			if bit&(1<<i) != 0 {
				table[bit/32] |= 1 << uint(bit%32)
			}
		}
		values[e.NodeID()] = table
	}
	var visit func(Edge) error
	visit = func(e Edge) error {
		id := e.NodeID()
		if _, ok := values[id]; ok {
			return nil
		}
		n := p.nodes[id]
		if n.Kind == Constant {
			table := make([]uint32, words)
			for i := range table {
				table[i] = ^uint32(0)
			}
			values[id] = table
			return nil
		}
		if n.Kind == Output || n.Kind == Buffer {
			if err := visit(n.Fanins[0]); err != nil {
				return err
			}
			table := append([]uint32(nil), values[n.Fanins[0].NodeID()]...)
			if n.Fanins[0].Complemented() {
				for i := range table {
					table[i] = ^table[i]
				}
			}
			values[id] = table
			return nil
		}
		if n.Kind != And && n.Kind != ExclusiveOr {
			return fmt.Errorf("aig: input %d outside cut", id)
		}
		if err := visit(n.Fanins[0]); err != nil {
			return err
		}
		if err := visit(n.Fanins[1]); err != nil {
			return err
		}
		values[id] = Aig_ManCutTruthOne(p, e, make([]uint32, words), values)
		return nil
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	result := append([]uint32(nil), values[root.NodeID()]...)
	if root.Complemented() {
		for i := range result {
			result[i] = ^result[i]
		}
	}
	return result, nil
}
