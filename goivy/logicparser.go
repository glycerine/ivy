// Package logicparser provides a public API for parsing Ivy formula and term
// strings into AST nodes. This wraps the existing hand-written parser's
// expression parsing capability.
//
// This corresponds to the Python functions to_formula(), to_term(), to_clause(),
// to_literal() in ivy_logic_utils.py, which use the PLY-based ivy_logic_parser.
package goivy

// DefaultVersion is the version used by the convenience functions
// (ToFormula, ToTerm, etc.) when no explicit version is provided.
var DefaultVersion = Version{1, 7}

// ParseFormula parses a formula string and returns the AST node.
// The entire input must be consumed; trailing tokens cause an error.
func ParseFormula(input string, version Version) (Node, error) {
	return parseString(input, version)
}

// ParseTerm parses a term string and returns the AST node.
// For v1.7+ where terms subsume formulas, this is identical to ParseFormula.
// For earlier versions the start symbol differs, but the hand-written parser
// uses unified expression parsing regardless.
func ParseTerm(input string, version Version) (Node, error) {
	return parseTermString(input, version)
}

// ToFormula parses a formula string using the default version.
func ToFormula(s string) (Node, error) {
	return ParseFormula(s, DefaultVersion)
}

// ToTerm parses a term string using the default version.
func ToTerm(s string) (Node, error) {
	return ParseTerm(s, DefaultVersion)
}

// ToFormulaV parses a formula string using the specified version.
func ToFormulaV(s string, version Version) (Node, error) {
	return ParseFormula(s, version)
}

// ToTermV parses a term string using the specified version.
func ToTermV(s string, version Version) (Node, error) {
	return ParseTerm(s, version)
}

// parseString dispatches to the LALR logic parser for all versions.
// The lalr_logicparser package has version-specific grammars (v1.2, v1.6, v1.7+)
// matching Python's ivy_logic_parser.py.
func parseString(input string, version Version) (Node, error) {
	return ParseLogic(input, version)
}

func parseTermString(input string, version Version) (Node, error) {
	return ParseLogicTerm(input, version)
}
