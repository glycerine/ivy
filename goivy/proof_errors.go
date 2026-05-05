// Package proof implements Ivy's built-in proof checker.
//
// It handles schema instantiation, goal management, matching/unification,
// and skolemization. Ported from Python ivy_proof.py.
package goivy

import "fmt"

// Redefinition is raised when a symbol is defined more than once.
type Redefinition struct {
	Msg  string
	Node interface{} // the AST node where the error occurred (may be nil)
}

func (e *Redefinition) Error() string {
	if e.Node != nil {
		return fmt.Sprintf("redefinition: %s (at %v)", e.Msg, e.Node)
	}
	return fmt.Sprintf("redefinition: %s", e.Msg)
}

// Circular is raised when a definition creates a circular dependency.
type Circular struct {
	Msg  string
	Node interface{}
}

func (e *Circular) Error() string {
	if e.Node != nil {
		return fmt.Sprintf("circular: %s (at %v)", e.Msg, e.Node)
	}
	return fmt.Sprintf("circular: %s", e.Msg)
}

// NoMatch is raised when a schema does not match the goal.
type NoMatch struct {
	Msg  string
	Node interface{}
}

func (e *NoMatch) Error() string {
	if e.Node != nil {
		return fmt.Sprintf("no match: %s (at %v)", e.Msg, e.Node)
	}
	return fmt.Sprintf("no match: %s", e.Msg)
}

// ProofError is raised for general proof-related errors.
type ProofError struct {
	Msg  string
	Node interface{}
}

func (e *ProofError) Error() string {
	if e.Node != nil {
		return fmt.Sprintf("proof error: %s (at %v)", e.Msg, e.Node)
	}
	return fmt.Sprintf("proof error: %s", e.Msg)
}

// CaptureError is raised when a substitution would cause variable capture.
type CaptureError struct {
	Msg  string
	Node interface{}
}

func (e *CaptureError) Error() string {
	if e.Node != nil {
		return fmt.Sprintf("capture error: %s (at %v)", e.Msg, e.Node)
	}
	return fmt.Sprintf("capture error: %s", e.Msg)
}
