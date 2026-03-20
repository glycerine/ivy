package actions

import (
	"fmt"
	"strings"

	"github.com/glycerine/goivy/ast"
	co "github.com/glycerine/goivy/clauseops"
)

func init() {
	// Register annotation conjunction callback to break import cycle.
	// clauseops cannot import actions, so it exposes AnnotConjFunc as a hook.
	co.AnnotConjFunc = func(a, b interface{}) interface{} {
		aa, ok1 := a.(Annotation)
		bb, ok2 := b.(Annotation)
		if ok1 && ok2 {
			return aa.Conj(bb)
		}
		return a // fallback: keep first
	}
}

// RegisterAnnotConj sets the annotation conjunction callback on a clauseops
// OpsConfig. Replaces the init()-based global assignment for per-session use.
func RegisterAnnotConj(coCfg *co.OpsConfig) {
	coCfg.AnnotConjFunc = func(a, b interface{}) interface{} {
		aa, ok1 := a.(Annotation)
		bb, ok2 := b.(Annotation)
		if ok1 && ok2 {
			return aa.Conj(bb)
		}
		return a // fallback: keep first
	}
}

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
	Rename(m map[string]string) Annotation
	// Ite creates an if-then-else annotation.
	Ite(cond string, other Annotation) Annotation
}

// --- EmptyAnnotation ---

// EmptyAnnotation is the trivial annotation.
type EmptyAnnotation struct{}

func (EmptyAnnotation) annotationMarker()   {}
func (EmptyAnnotation) String() string      { return "()" }

func (e EmptyAnnotation) Conj(other Annotation) Annotation {
	return &ConjAnnotation{Args: []Annotation{e, other}}
}
func (e EmptyAnnotation) Compose(other Annotation) Annotation {
	return &ComposeAnnotation{Args: []Annotation{e, other}}
}
func (e EmptyAnnotation) Rename(m map[string]string) Annotation {
	if len(m) == 0 {
		return e
	}
	return &RenameAnnotation{Arg: e, Map: m}
}
func (e EmptyAnnotation) Ite(cond string, other Annotation) Annotation {
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
func (c *ConjAnnotation) Rename(m map[string]string) Annotation {
	if len(m) == 0 {
		return c
	}
	return &RenameAnnotation{Arg: c, Map: m}
}
func (c *ConjAnnotation) Ite(cond string, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: c, ElseB: other}
}

// --- ComposeAnnotation ---

// ComposeAnnotation sequentially composes annotations.
type ComposeAnnotation struct {
	Args   []Annotation
	Lineno *ast.Location
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
func (c *ComposeAnnotation) Rename(m map[string]string) Annotation {
	if len(m) == 0 {
		return c
	}
	return &RenameAnnotation{Arg: c, Map: m}
}
func (c *ComposeAnnotation) Ite(cond string, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: c, ElseB: other}
}

// --- RenameAnnotation ---

// RenameAnnotation renames symbols according to a map.
type RenameAnnotation struct {
	Arg Annotation
	Map map[string]string
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
func (r *RenameAnnotation) Rename(m map[string]string) Annotation {
	if len(m) == 0 {
		return r
	}
	return &RenameAnnotation{Arg: r, Map: m}
}
func (r *RenameAnnotation) Ite(cond string, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: r, ElseB: other}
}

// --- IteAnnotation ---

// IteAnnotation represents an if-then-else over annotations, keyed on a
// branch condition variable.
type IteAnnotation struct {
	Cond  string
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
func (i *IteAnnotation) Rename(m map[string]string) Annotation {
	if len(m) == 0 {
		return i
	}
	return &RenameAnnotation{Arg: i, Map: m}
}
func (i *IteAnnotation) Ite(cond string, other Annotation) Annotation {
	return &IteAnnotation{Cond: cond, ThenB: i, ElseB: other}
}

// AnnotationError is returned when annotation processing fails.
type AnnotationError struct {
	Msg string
}

func (e *AnnotationError) Error() string { return e.Msg }
