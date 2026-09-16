package goivy

import (
	"fmt"
)

// IvyError is the general error type for Ivy.
type IvyError struct {
	Msg    string
	Loc    Location
	HasLoc bool
}

func (e *IvyError) Error() string {
	if e.HasLoc {
		return formatIvyErrorAt(e.Loc, e.Msg)
	}
	return e.Msg
}

func formatIvyErrorAt(loc Location, msg string) string {
	if loc.Reference != nil {
		return formatIvyErrorAt(*loc.Reference, msg) + "\n" + formatLocationWithoutReference(loc) + "error: instantiated here"
	}
	return formatLocationWithoutReference(loc) + "error: " + msg
}

func formatLocationWithoutReference(loc Location) string {
	return Location{Filename: loc.Filename, Line: loc.Line}.String()
}

// NewIvyError creates an IvyError with location extracted from an AST node.
// Matches Python: IvyError(ast, msg) which does ast.lineno if hasattr(ast,'lineno').
func NewIvyError(node Node, msg string) *IvyError {
	if node != nil {
		loc := node.GetLineno()
		if loc.Line > 0 || loc.Filename != "" {
			return &IvyError{Msg: msg, Loc: loc, HasLoc: true}
		}
	}
	return &IvyError{Msg: msg}
}

// SortError is raised for sort-related errors (type mismatches, etc.).
type SortError struct {
	Msg    string
	Loc    Location
	HasLoc bool
}

func (e *SortError) Error() string {
	if e.HasLoc {
		return fmt.Sprintf("%serror: %s", e.Loc, e.Msg)
	}
	return e.Msg
}

// NewSortError creates a SortError with location extracted from an AST node.
func NewSortError(node Node, msg string) *SortError {
	if node != nil {
		loc := node.GetLineno()
		if loc.Line > 0 || loc.Filename != "" {
			return &SortError{Msg: msg, Loc: loc, HasLoc: true}
		}
	}
	return &SortError{Msg: msg}
}
