package webui

import (
	"fmt"
	goivy "github.com/glycerine/ivy/goivy"
	"strings"
)

// parseConstraintString parses an Ivy formula string to logic.Expr.
// The LALR parser produces ast.Node; if the result satisfies logic.Expr
// directly (which it does for simple formulas), we return it.
// For complex formulas requiring sort inference, a Compiler is needed
// (use Graph.AddConstraintsExpr / SetFactsExpr instead).
func parseConstraintString(s string) (goivy.Expr, error) {
	astNode, err := goivy.ToFormula(s)
	if err != nil {
		return nil, fmt.Errorf("parse constraint %q: %w", s, err)
	}
	if expr, ok := astNode.(goivy.Expr); ok {
		return expr, nil
	}
	return nil, fmt.Errorf("constraint %q: parsed to %T, not logic.Expr (use *Expr variant with compiled context)", s, astNode)
}

func parseBMCErrorConditionString(s string) (goivy.Expr, error) {
	text := strings.TrimSpace(s)
	switch strings.ToLower(text) {
	case "true":
		return goivy.True, nil
	case "false":
		return goivy.False, nil
	}
	astNode, err := goivy.ToFormula(text)
	if err != nil {
		return nil, fmt.Errorf("parse error condition: %w", err)
	}
	if expr, ok := astNode.(goivy.Expr); ok {
		return expr, nil
	}
	return nil, fmt.Errorf("error condition is not a logic expression")
}
