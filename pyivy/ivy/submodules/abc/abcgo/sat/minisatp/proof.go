// Package minisatp ports MiniSat-p 1.14, by Niklas Een and Niklas Sorensson.
// The original MIT license is reproduced in LICENSE in this directory.
package minisatp

import "sort"

// ClauseID identifies a root or derived clause in proof insertion order.
type ClauseID int32

// Chain is a sequence of resolutions: Clauses[0] resolves with Clauses[i+1]
// on Pivots[i]. References point to previously recorded clauses.
type Chain struct {
	Clauses []ClauseID
	Pivots  []int32
}

// proofEvent retains deletions and the original insertion order, as in
// Proof::traverse in MiniSat-p's Proof.C. Clause data is kept in memory rather
// than serialized to a temporary file.
type proofEvent struct {
	id      ClauseID
	root    []int32
	chain   Chain
	deleted bool
}

type proof struct {
	next   ClauseID
	events []proofEvent
	chain  Chain
}

func (p *proof) addRoot(lits []int32) ClauseID {
	root := append([]int32(nil), lits...)
	sort.Slice(root, func(i, j int) bool { return root[i] < root[j] })
	n := 0
	for _, lit := range root {
		if n == 0 || root[n-1] != lit {
			root[n] = lit
			n++
		}
	}
	id := p.next
	p.next++
	p.events = append(p.events, proofEvent{id: id, root: root[:n]})
	return id
}

func (p *proof) beginChain(start ClauseID) {
	p.chain.Clauses = append(p.chain.Clauses[:0], start)
	p.chain.Pivots = p.chain.Pivots[:0]
}

func (p *proof) resolve(next ClauseID, variable int32) {
	p.chain.Clauses = append(p.chain.Clauses, next)
	p.chain.Pivots = append(p.chain.Pivots, variable)
}

func (p *proof) endChain() ClauseID {
	if len(p.chain.Pivots) == 0 {
		return p.chain.Clauses[0]
	}
	id := p.next
	p.next++
	chain := Chain{
		Clauses: append([]ClauseID(nil), p.chain.Clauses...),
		Pivots:  append([]int32(nil), p.chain.Pivots...),
	}
	p.events = append(p.events, proofEvent{id: id, chain: chain})
	return id
}

func (p *proof) deleted(id ClauseID) {
	p.events = append(p.events, proofEvent{id: id, deleted: true})
}
