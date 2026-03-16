// Package logicparser provides a public API for parsing Ivy formula and term
// strings into AST nodes. This wraps the existing hand-written parser's
// expression parsing capability.
//
// This corresponds to the Python functions to_formula(), to_term(), to_clause(),
// to_literal() in ivy_logic_utils.py, which use the PLY-based ivy_logic_parser.
package logicparser

import (
	"fmt"

	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/parser"
)

// DefaultVersion is the version used by the convenience functions
// (ToFormula, ToTerm, etc.) when no explicit version is provided.
var DefaultVersion = lexer.Version{1, 7}

// ParseFormula parses a formula string and returns the AST node.
// The entire input must be consumed; trailing tokens cause an error.
func ParseFormula(input string, version lexer.Version) (ast.Node, error) {
	return parseString(input, version)
}

// ParseTerm parses a term string and returns the AST node.
// For v1.7+ where terms subsume formulas, this is identical to ParseFormula.
// For earlier versions the start symbol differs, but the hand-written parser
// uses unified expression parsing regardless.
func ParseTerm(input string, version lexer.Version) (ast.Node, error) {
	return parseString(input, version)
}

// ToFormula parses a formula string using the default version.
func ToFormula(s string) (ast.Node, error) {
	return ParseFormula(s, DefaultVersion)
}

// ToTerm parses a term string using the default version.
func ToTerm(s string) (ast.Node, error) {
	return ParseTerm(s, DefaultVersion)
}

// ToFormulaV parses a formula string using the specified version.
func ToFormulaV(s string, version lexer.Version) (ast.Node, error) {
	return ParseFormula(s, version)
}

// ToTermV parses a term string using the specified version.
func ToTermV(s string, version lexer.Version) (ast.Node, error) {
	return ParseTerm(s, version)
}

// parseString creates a parser, parses a single expression, and verifies
// the entire input was consumed.
func parseString(input string, version lexer.Version) (ast.Node, error) {
	p := parser.New(input, version)
	result := p.ParseExpr(0)

	if result == nil {
		errs := p.Errors()
		if len(errs) > 0 {
			return nil, fmt.Errorf("%s", errs[0].Error())
		}
		return nil, fmt.Errorf("failed to parse expression from %q", input)
	}

	if !p.AtEOF() {
		return nil, fmt.Errorf("unexpected trailing input after expression: token type %v", p.CurrentTokenType())
	}

	errs := p.Errors()
	if len(errs) > 0 {
		return nil, fmt.Errorf("%s", errs[0].Error())
	}

	return result, nil
}
