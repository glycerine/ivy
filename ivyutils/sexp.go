package ivyutils

// NodeKey is a structural identity key for logic nodes.
// It is a DISTINCT TYPE (not a string alias) so the compiler enforces
// that map[NodeKey] cannot accept plain strings and vice versa.
// Two nodes with the same NodeKey are structurally equal,
// matching Python's recstruct == and hash behavior.
type NodeKey string

// String returns the NodeKey as a plain string for printing.
func (k NodeKey) String() string {
	return string(k)
}

// Sexper supports rendering as an S-expression
// string that is the special NodeKey typed string.
// The NodeKey is distinct to enforce that just
// the name will not do (need the "Sort" type too).
type Sexper interface {
	Sexp() NodeKey
}
