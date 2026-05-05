// Parse-and-compile convenience functions.
// These correspond to Python's to_clause(s), to_clauses(s), to_literal(s)
// from ivy_logic_utils.py which combine parsing + compilation + conversion.
package goivy

// ToFormula parses a string, compiles the AST, and applies sort inference.
// Python equivalent: to_formula(s) = formula_parser.parse(s).compile_with_sort_inference()
func (c *Compiler) ToFormula(s string) (Expr, error) {
	return c.ToFormulaV(s, DefaultVersion)
}

// ToFormulaV parses a string at the given version, compiles and infers sorts.
func (c *Compiler) ToFormulaV(s string, version Version) (Expr, error) {
	astNode, err := ParseFormula(s, version)
	if err != nil {
		return nil, err
	}
	return c.SortifyWithInference(astNode)
}

// ToTerm parses a term string, compiles it, and applies sort inference.
// Python equivalent: to_term(s) = term_parser.parse(s)
func (c *Compiler) ToTerm(s string) (Expr, error) {
	return c.ToTermV(s, DefaultVersion)
}

// ToTermV parses a term string at the given version, compiles and infers sorts.
func (c *Compiler) ToTermV(s string, version Version) (Expr, error) {
	astNode, err := ParseTerm(s, version)
	if err != nil {
		return nil, err
	}
	return c.SortifyWithInference(astNode)
}

// ToClause parses a formula string and converts it to clause form (disjunction of literals).
// Python equivalent: to_clause(s) = formula_to_clause(to_formula(s))
func (c *Compiler) ToClause(s string) ([]Expr, error) {
	f, err := c.ToFormula(s)
	if err != nil {
		return nil, err
	}
	tc := NewTseitinContext(nil)
	return FormulaToClause(tc, f), nil
}

// ToClauses parses a formula string and converts it to Clauses form.
// Python equivalent: to_clauses(s) = formula_to_clauses(to_formula(s))
func (c *Compiler) ToClauses(s string) (*Clauses, error) {
	f, err := c.ToFormula(s)
	if err != nil {
		return nil, err
	}
	return FormulaToClauses(f, nil), nil
}

// ToLiteral parses a formula string and converts it to a literal (positive or negated atom).
// Python equivalent: to_literal(s) = formula_to_lit(to_formula(s))
func (c *Compiler) ToLiteral(s string) (Expr, error) {
	f, err := c.ToFormula(s)
	if err != nil {
		return nil, err
	}
	tc := NewTseitinContext(nil)
	return FormulaToLit(tc, f), nil
}
