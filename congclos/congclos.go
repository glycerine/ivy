// Ported to Go from ivy_congclos.py.

// Package congclos implements congruence closure using union-find.
// Representative terms are minimal in term order (lexicographic by name).
// Supports push/pop for backtracking.
package congclos

import (
	"github.com/glycerine/goivy/logic"
)

// node is an entry in the union-find structure.
// node[0] = the term, node[1] = pointer to representative node (nil if root).
type node struct {
	term Term
	rep  *node // nil means this node is its own representative
}

// undoRep records a previous rep pointer for undo.
type undoRep struct {
	n      *node
	oldRep *node
}

func (u *undoRep) undoit() {
	u.n.rep = u.oldRep
}

// Term is a logic.Expr (Var, Const, Apply, etc.)
type Term = logic.Expr

// CongClos implements congruence closure with union-find.
// For now there are no function symbols, so this is just union-find.
type CongClos struct {
	tab    map[string]*node
	trail  []*undoRep
	pushes []int
}

// New creates a new CongClos.
func New() *CongClos {
	return &CongClos{
		tab: make(map[string]*node),
	}
}

// setRep sets the representative of n to rep, recording the old value
// on the trail for backtracking.
func (cc *CongClos) setRep(n *node, rep *node) {
	if n.rep != rep {
		cc.trail = append(cc.trail, &undoRep{n: n, oldRep: n.rep})
		n.rep = rep
	}
}

// getRepRec recursively finds the representative, performing path compression.
func (cc *CongClos) getRepRec(n *node) *node {
	if n.rep == nil {
		return n
	}
	rep := cc.getRepRec(n.rep)
	cc.setRep(n, rep)
	return rep
}

// getRepNode retrieves (or creates) the node for a term and returns its
// representative node.
func (cc *CongClos) getRepNode(term Term) *node {
	name := nodeName(term)
	n, ok := cc.tab[name]
	if !ok {
		n = &node{term: term, rep: nil}
		cc.tab[name] = n
	}
	return cc.getRepRec(n)
}

// Find returns the representative term for the given term.
func (cc *CongClos) Find(term Term) Term {
	return cc.getRepNode(term).term
}

// FindByName looks up a constant by name (creating one if needed) and
// returns its representative term.
func (cc *CongClos) FindByName(name string) Term {
	n, ok := cc.tab[name]
	if !ok {
		n = &node{term: logic.NewSymbol(name, logic.TopS), rep: nil}
		cc.tab[name] = n
	}
	return cc.getRepRec(n).term
}

// Union merges the equivalence classes of term1 and term2.
// The representative with the lexicographically smaller name wins.
func (cc *CongClos) Union(term1, term2 Term) {
	rep1 := cc.getRepNode(term1)
	rep2 := cc.getRepNode(term2)
	if rep1 == rep2 {
		return
	}
	name1 := nodeName(rep1.term)
	name2 := nodeName(rep2.term)
	if name1 < name2 {
		cc.setRep(rep2, rep1)
	} else {
		cc.setRep(rep1, rep2)
	}
}

// Literal wraps an atom with a polarity for Theory output.
type Literal struct {
	Polarity int
	Atom     *EqAtom
}

// EqAtom is a simplified equality atom for Theory output.
type EqAtom struct {
	RelName string
	Args    [2]Term
}

// Theory returns the list of equalities implied by the current state.
// Each equality says "var = representative".
func (cc *CongClos) Theory() []Literal {
	var result []Literal
	for _, n := range cc.tab {
		varTerm := n.term
		repTerm := cc.getRepRec(n).term
		if varTerm != repTerm {
			result = append(result, Literal{
				Polarity: 1,
				Atom:     &EqAtom{RelName: "=", Args: [2]Term{varTerm, repTerm}},
			})
		}
	}
	return result
}

// Push saves the current state for later restoration with Pop.
func (cc *CongClos) Push() {
	cc.pushes = append(cc.pushes, len(cc.trail))
}

// Pop restores the state saved by the most recent Push.
func (cc *CongClos) Pop() {
	if len(cc.pushes) == 0 {
		return
	}
	newLen := cc.pushes[len(cc.pushes)-1]
	cc.pushes = cc.pushes[:len(cc.pushes)-1]
	for len(cc.trail) > newLen {
		cc.trail[len(cc.trail)-1].undoit()
		cc.trail = cc.trail[:len(cc.trail)-1]
	}
}

// nodeName extracts the name used for table lookup from a term.
func nodeName(t Term) string {
	switch v := t.(type) {
	case *logic.Variable:
		return v.Name
	case *logic.Symbol:
		return v.Name
	default:
		return t.String()
	}
}
