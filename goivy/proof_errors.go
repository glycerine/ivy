// Package proof implements Ivy's built-in proof checker.
//
// It handles schema instantiation, goal management, matching/unification,
// and skolemization. Ported from Python ivy_proof.py.
package goivy

// Redefinition is raised when a symbol is defined more than once.
type Redefinition struct {
	Msg  string
	Node interface{} // the AST node where the error occurred (may be nil)
}

// Error formats like Python's iu.IvyError, of which Redefinition is a plain
// subclass: str(lineno) + "error: " + msg, or msg when unlocated.
func (e *Redefinition) Error() string { return proofErrorString(e.Node, e.Msg) }

// HasLocation reports whether the error carries a source location.
func (e *Redefinition) HasLocation() bool { _, ok := proofErrorLoc(e.Node); return ok }

// Circular is raised when a definition creates a circular dependency.
type Circular struct {
	Msg  string
	Node interface{}
}

// Error formats like Python's iu.IvyError, of which Circular is a plain
// subclass: str(lineno) + "error: " + msg, or msg when unlocated.
func (e *Circular) Error() string { return proofErrorString(e.Node, e.Msg) }

// HasLocation reports whether the error carries a source location.
func (e *Circular) HasLocation() bool { _, ok := proofErrorLoc(e.Node); return ok }

// NoMatch is raised when a schema does not match the goal.
type NoMatch struct {
	Msg  string
	Node interface{}
}

// Error formats like Python's iu.IvyError, of which NoMatch is a plain
// subclass: str(lineno) + "error: " + msg, or msg when unlocated.
func (e *NoMatch) Error() string { return proofErrorString(e.Node, e.Msg) }

// HasLocation reports whether the error carries a source location.
func (e *NoMatch) HasLocation() bool { _, ok := proofErrorLoc(e.Node); return ok }

// ProofError is raised for general proof-related errors.
type ProofError struct {
	Msg  string
	Node interface{}
}

// Error formats like Python's iu.IvyError, of which ProofError is a plain
// subclass: str(lineno) + "error: " + msg, or msg when unlocated.
func (e *ProofError) Error() string { return proofErrorString(e.Node, e.Msg) }

// HasLocation reports whether the error carries a source location.
func (e *ProofError) HasLocation() bool { _, ok := proofErrorLoc(e.Node); return ok }

// CaptureError is raised when a substitution would cause variable capture.
type CaptureError struct {
	Msg  string
	Node interface{}
}

// Error formats like Python's iu.IvyError, of which CaptureError is a plain
// subclass: str(lineno) + "error: " + msg, or msg when unlocated.
func (e *CaptureError) Error() string { return proofErrorString(e.Node, e.Msg) }

// HasLocation reports whether the error carries a source location.
func (e *CaptureError) HasLocation() bool { _, ok := proofErrorLoc(e.Node); return ok }

// proofErrorLoc is the location of the node a proof error was raised at,
// if it has one (Python: ast.lineno when hasattr(ast,'lineno')).
func proofErrorLoc(node interface{}) (Location, bool) {
	n, ok := node.(Node)
	if !ok || n == nil {
		return Location{}, false
	}
	loc := n.GetLineno()
	if loc.Line > 0 || loc.Filename != "" {
		return loc, true
	}
	return Location{}, false
}

// proofErrorString formats a proof error the way Python's IvyError does.
func proofErrorString(node interface{}, msg string) string {
	if loc, ok := proofErrorLoc(node); ok {
		return formatIvyErrorAt(loc, msg)
	}
	return msg
}
