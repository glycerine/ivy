package parse

import (
	"fmt"
	"io"
)

// Parse_FormulaParser retains the original parser's diagnostic-stream and
// failure-zero convention, using typed callbacks instead of a concrete C BDD
// manager. vars supplies the original zero-rank variable nodes; Shift implements
// the backend's variable-index permutation. New callers may use ParseFormula
// and Compile directly to receive structured errors.
func Parse_FormulaParser[T any](output io.Writer, formula string, nVars, nRanks int, names []string, ops Operations[T], vars []T) T {
	var zero T
	fail := func(err error) T {
		if output != nil {
			fmt.Fprintln(output, "Parse_FormulaParser():", err)
		}
		return zero
	}
	if nVars < 0 || nRanks < 0 || len(names) < nVars || len(vars) < nVars {
		return fail(fmt.Errorf("invalid variable/rank inventory"))
	}
	e, err := ParseFormula(formula, names[:nVars])
	if err != nil {
		return fail(err)
	}
	ops.Variable = func(index int) T { return vars[index] }
	result, err := Compile(e, ops, nVars)
	if err != nil {
		return fail(err)
	}
	return result
}
func Parse_FormulaParserEqn[T any](output io.Writer, formula string, names []string, ops Operations[T]) T {
	result, err := CompileEquation(formula, names, ops)
	if err != nil {
		if output != nil {
			fmt.Fprintln(output, "Parse_FormulaParserEqn():", err)
		}
		var zero T
		return zero
	}
	return result
}
