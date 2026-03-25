package ast

import iu "github.com/glycerine/goivy/ivyutils"

// Labeler generates unique label atoms for proof subgoals.
// Corresponds to Python's ivy_ast.Labeler (ivy_ast.py:1938-1942).
type Labeler struct {
	rn *iu.UniqueRenamer
}

// NewLabeler creates a new Labeler.
func NewLabeler() *Labeler {
	return &Labeler{rn: iu.NewUniqueRenamer("", nil)}
}

// Call returns a fresh Atom with a unique name.
// Corresponds to Python's Labeler.__call__ which returns Atom(self.rn(), []).
func (lb *Labeler) Call() *Atom {
	name := lb.rn.Rename("")
	return &Atom{Rep: name}
}
