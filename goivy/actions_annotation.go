package goivy

import (
	"fmt"
	"strings"
)

// Annotation lets us reconstruct an execution trace from a satisfying assignment.
// It contains two kinds of information:
//   - For each symbol in the update formula corresponding to a program variable,
//     the original program symbol and the execution point at which it was introduced.
//   - For each symbol corresponding to a branch decision, the program execution
//     point of the branch.
type Annotation interface {
	annotationMarker()
	String() string
	// Conj merges two annotations (requires disjoint domains).
	Conj(other Annotation) Annotation
	// Compose sequentially composes two annotations.
	Compose(other Annotation) Annotation
	// Rename renames symbols according to the map.
	// Keys are lg.NodeKey (Sexp-based structural identity, matching Python's
	// symbol __hash__/__eq__). Values are lg.Expr.
	Rename(m map[NodeKey]Expr) Annotation
	// Ite creates an if-then-else annotation.
	// cond is an lg.Expr (typically *lg.Const) — matching the Python source of truth
	// where IteAnnotation.cond is an lg.Const, not a string.
	Ite(cond Expr, other Annotation) Annotation
}

// ActionAnnotation mirrors Python's lf.annot tuple `(action, annot)` used by
// tactics-generated goals.
type ActionAnnotation struct {
	Action ActionsAction
	Annot  Annotation
}

// --- EmptyAnnotation ---

// EmptyAnnotation is the trivial annotation.
type EmptyAnnotation struct{}

func (EmptyAnnotation) annotationMarker() {}
func (EmptyAnnotation) String() string    { return "()" }

func (e EmptyAnnotation) Conj(other Annotation) Annotation {
	return &ConjAnnotation{Args: []Annotation{e, other}}
}
func (e EmptyAnnotation) Compose(other Annotation) Annotation {
	return &ComposeAnnotation{Args: []Annotation{e, other}}
}
func (e EmptyAnnotation) Rename(m map[NodeKey]Expr) Annotation {
	if len(m) == 0 {
		return e
	}
	return newRenameAnnotation(e, m)
}
func (e EmptyAnnotation) Ite(cond Expr, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: e, ElseB: other}
}

// --- ConjAnnotation ---

// ConjAnnotation merges annotations with disjoint domains.
type ConjAnnotation struct {
	Args []Annotation
}

func (ConjAnnotation) annotationMarker() {}
func (c *ConjAnnotation) String() string {
	parts := make([]string, len(c.Args))
	for i, a := range c.Args {
		parts[i] = a.String()
	}
	return "And(" + strings.Join(parts, ",") + ")"
}

func (c *ConjAnnotation) Conj(other Annotation) Annotation {
	return &ConjAnnotation{Args: []Annotation{c, other}}
}
func (c *ConjAnnotation) Compose(other Annotation) Annotation {
	return &ComposeAnnotation{Args: []Annotation{c, other}}
}
func (c *ConjAnnotation) Rename(m map[NodeKey]Expr) Annotation {
	if len(m) == 0 {
		return c
	}
	return newRenameAnnotation(c, m)
}
func (c *ConjAnnotation) Ite(cond Expr, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: c, ElseB: other}
}

// --- ComposeAnnotation ---

// ComposeAnnotation sequentially composes annotations.
type ComposeAnnotation struct {
	Args   []Annotation
	Lineno *Location
}

func (ComposeAnnotation) annotationMarker() {}
func (c *ComposeAnnotation) String() string {
	prefix := ""
	if c.Lineno != nil {
		prefix = c.Lineno.String()
	}
	parts := make([]string, len(c.Args))
	for i, a := range c.Args {
		parts[i] = a.String()
	}
	return prefix + "Compose(" + strings.Join(parts, ",") + ")"
}

func (c *ComposeAnnotation) Conj(other Annotation) Annotation {
	return &ConjAnnotation{Args: []Annotation{c, other}}
}
func (c *ComposeAnnotation) Compose(other Annotation) Annotation {
	return &ComposeAnnotation{Args: []Annotation{c, other}}
}
func (c *ComposeAnnotation) Rename(m map[NodeKey]Expr) Annotation {
	if len(m) == 0 {
		return c
	}
	return newRenameAnnotation(c, m)
}
func (c *ComposeAnnotation) Ite(cond Expr, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: c, ElseB: other}
}

// --- RenameAnnotation ---

// RenameAnnotation renames symbols according to a map.
// In Python, self.map maps lg.Const → lg.Const using structural equality.
// In Go, keys are lg.NodeKey (Sexp-based structural identity) and values
// are lg.Expr, matching Python's __hash__/__eq__ behavior.
type RenameAnnotation struct {
	Arg Annotation
	Map map[NodeKey]Expr
}

func newRenameAnnotation(arg Annotation, m map[NodeKey]Expr) *RenameAnnotation {
	copied := make(map[NodeKey]Expr, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return &RenameAnnotation{Arg: arg, Map: copied}
}

func (RenameAnnotation) annotationMarker() {}
func (r *RenameAnnotation) String() string {
	pairs := make([]string, 0, len(r.Map))
	for k, v := range r.Map {
		pairs = append(pairs, fmt.Sprintf("%s:%s", k, v))
	}
	return fmt.Sprintf("Rename(%s,{%s})", r.Arg, strings.Join(pairs, ","))
}

func (r *RenameAnnotation) Conj(other Annotation) Annotation {
	return &ConjAnnotation{Args: []Annotation{r, other}}
}
func (r *RenameAnnotation) Compose(other Annotation) Annotation {
	return &ComposeAnnotation{Args: []Annotation{r, other}}
}
func (r *RenameAnnotation) Rename(m map[NodeKey]Expr) Annotation {
	if len(m) == 0 {
		return r
	}
	return newRenameAnnotation(r, m)
}
func (r *RenameAnnotation) Ite(cond Expr, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: r, ElseB: other}
}

// --- IteAnnotation ---

// IteAnnotation represents an if-then-else over annotations, keyed on a
// branch condition variable.
// In Python, cond is an lg.Const (set via a.ite(v, annot) where v is an
// lg.Const). The Go port now matches the Python source of truth.
type IteAnnotation struct {
	Cond  Expr
	ThenB Annotation
	ElseB Annotation
}

func (IteAnnotation) annotationMarker() {}
func (i *IteAnnotation) String() string {
	return fmt.Sprintf("Ite(%s,%s,%s)", i.Cond, i.ThenB, i.ElseB)
}

func (i *IteAnnotation) Conj(other Annotation) Annotation {
	return &ConjAnnotation{Args: []Annotation{i, other}}
}
func (i *IteAnnotation) Compose(other Annotation) Annotation {
	return &ComposeAnnotation{Args: []Annotation{i, other}}
}
func (i *IteAnnotation) Rename(m map[NodeKey]Expr) Annotation {
	if len(m) == 0 {
		return i
	}
	return newRenameAnnotation(i, m)
}
func (i *IteAnnotation) Ite(cond Expr, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: i, ElseB: other}
}

// AnnotationError is returned when annotation processing fails.
type AnnotationError struct {
	Msg string
}

func (e *AnnotationError) Error() string { return e.Msg }
