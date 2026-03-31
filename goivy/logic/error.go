package logic

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
)

// IvyError is the general error type for Ivy.
type IvyError struct {
	Msg    string
	Loc    ast.Location
	HasLoc bool
}

func (e *IvyError) Error() string {
	if e.HasLoc {
		return fmt.Sprintf("%s: error: %s", e.Loc, e.Msg)
	}
	return e.Msg
}

// NewIvyError creates an IvyError with location extracted from an AST node.
// Matches Python: IvyError(ast, msg) which does ast.lineno if hasattr(ast,'lineno').
func NewIvyError(node ast.Node, msg string) *IvyError {
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
	Loc    ast.Location
	HasLoc bool
}

func (e *SortError) Error() string {
	if e.HasLoc {
		return fmt.Sprintf("%s: error: %s", e.Loc, e.Msg)
	}
	return e.Msg
}

// NewSortError creates a SortError with location extracted from an AST node.
func NewSortError(node ast.Node, msg string) *SortError {
	if node != nil {
		loc := node.GetLineno()
		if loc.Line > 0 || loc.Filename != "" {
			return &SortError{Msg: msg, Loc: loc, HasLoc: true}
		}
	}
	return &SortError{Msg: msg}
}
