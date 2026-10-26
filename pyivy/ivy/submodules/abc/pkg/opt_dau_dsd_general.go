package abc

// General truth-table DSD decomposition from opt/dau/dauDsd.c.

import (
	"fmt"
	"slices"
)

// DauDSDDecomposeOptions bounds general DSD decomposition. Zero limits select
// package defaults. OmitPrimeTruth preserves the source's analysis-only
// spelling for prime blocks; such a spelling is intentionally not parseable.
// SplitPrime enables the source's one-step prime splitting policy.
type DauDSDDecomposeOptions struct {
	MaxBytes, MaxNodes, MaxDepth, MaxWords int
	OmitPrimeTruth, SplitPrime             bool
	VariableLevels                         []int
	Stop                                   func() bool
}

// DauDSDDecomposition owns the formula and reports the largest irreducible
// prime block. A zero LargestPrime means the function is fully decomposable.
// SplitPrimeFailed records the source status in which no variable has two
// fully decomposable cofactors; Formula still contains the unsplit prime.
type DauDSDDecomposition struct {
	Formula          string
	LargestPrime     int
	SplitPrimeFailed bool
}

type dauDSDGeneralDecomposer struct {
	options DauDSDDecomposeOptions
	nodes   int
}

type dauDSDExpression struct {
	text        string
	level       int
	synthesized bool
}

func normalizeDauDSDDecomposeOptions(options DauDSDDecomposeOptions) (DauDSDDecomposeOptions, error) {
	if options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxDepth < 0 || options.MaxWords < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = DauMaxFormulaBytes
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 4096
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 128
	}
	if options.MaxWords == 0 {
		options.MaxWords = DauMaxTruthWords
	}
	return options, nil
}

// DauDSDDecompose decomposes an owned copy of truth using the source scan order:
// output variables, input pairs, mux structures, then an irreducible prime.
// The input table is never modified. The default truth-bearing result can be
// parsed by ParseDauDSD and replayed independently.
func DauDSDDecompose(truth *TruthTable, options DauDSDDecomposeOptions) (DauDSDDecomposition, error) {
	fail := func(err error) (DauDSDDecomposition, error) { return DauDSDDecomposition{}, err }
	o, err := normalizeDauDSDDecomposeOptions(options)
	if err != nil {
		return fail(err)
	}
	if truth == nil {
		return fail(ErrDimension)
	}
	if err = truth.Check(); err != nil {
		return fail(err)
	}
	if truth.variables < 0 || truth.variables > DauMaxVariables || len(truth.words) > o.MaxWords {
		return fail(ErrLimit)
	}
	if len(o.VariableLevels) != 0 && len(o.VariableLevels) != truth.variables {
		return fail(ErrDimension)
	}
	if o.Stop != nil && o.Stop() {
		return fail(ErrCancelled)
	}
	if truth.IsConstant(false) {
		return DauDSDDecomposition{Formula: "0"}, nil
	}
	if truth.IsConstant(true) {
		return DauDSDDecomposition{Formula: "1"}, nil
	}
	expressions := make([]dauDSDExpression, truth.variables)
	for variable := range expressions {
		expressions[variable].text = string(rune('a' + variable))
		if len(o.VariableLevels) != 0 {
			expressions[variable].level = o.VariableLevels[variable]
		}
	}
	d := dauDSDGeneralDecomposer{options: o}
	formula, largest, err := d.decompose(truth.Clone(), expressions, 0, o.SplitPrime)
	if err != nil {
		return fail(err)
	}
	if len(formula) > o.MaxBytes {
		return fail(ErrLimit)
	}
	if !o.OmitPrimeTruth {
		formula, err = DauDSDRemoveBraces(formula, DauDSDOptions{
			MaxBytes: o.MaxBytes, MaxNodes: o.MaxNodes, MaxDepth: o.MaxDepth, MaxVariables: DauMaxVariables, MaxWords: o.MaxWords, Stop: o.Stop,
		})
		if err != nil {
			return fail(fmt.Errorf("simplify generated DSD formula: %w", err))
		}
		parsed, parseErr := ParseDauDSD(formula, DauDSDOptions{
			MaxBytes: o.MaxBytes, MaxNodes: o.MaxNodes, MaxDepth: o.MaxDepth, MaxVariables: DauMaxVariables, Stop: o.Stop,
		})
		if parseErr != nil {
			return fail(fmt.Errorf("generated DSD formula: %w", parseErr))
		}
		if parsed.variables > truth.variables {
			return fail(ErrInvariant)
		}
	}
	failed := largest < 0
	if failed {
		largest = -largest
	}
	return DauDSDDecomposition{Formula: formula, LargestPrime: largest, SplitPrimeFailed: failed}, nil
}

// DauDSDCheckOneStep returns the source status codes for natural variable
// priority: -1 for a fully decomposable function, -2 when no one-variable split
// has two fully decomposable cofactors, or the selected variable position.
func DauDSDCheckOneStep(truth *TruthTable, options DauDSDDecomposeOptions) (int, error) {
	if truth == nil {
		return 0, ErrDimension
	}
	o, err := normalizeDauDSDDecomposeOptions(options)
	if err != nil {
		return 0, err
	}
	o.SplitPrime = false
	base, err := DauDSDDecompose(truth, o)
	if err != nil {
		return 0, err
	}
	if base.LargestPrime == 0 {
		return -1, nil
	}
	expressions := make([]dauDSDExpression, truth.variables)
	for variable := range expressions {
		expressions[variable] = dauDSDExpression{text: string(rune('a' + variable))}
		if len(o.VariableLevels) != 0 {
			expressions[variable].level = o.VariableLevels[variable]
		}
	}
	best, err := dauDSDOneStepVariable(truth, o, dauDSDLevelPriority(expressions, len(o.VariableLevels) != 0))
	if err != nil {
		return 0, err
	}
	if best < 0 {
		return -2, nil
	}
	return best, nil
}

func dauDSDOneStepVariable(table *TruthTable, options DauDSDDecomposeOptions, priority []int) (int, error) {
	best, bestSupport := -1, int(^uint(0)>>1)
	if len(priority) == 0 {
		priority = make([]int, table.variables)
		for variable := range priority {
			priority[variable] = variable
		}
	}
	cofactorOptions := options
	cofactorOptions.VariableLevels = nil
	for _, variable := range priority {
		if options.Stop != nil && options.Stop() {
			return 0, ErrCancelled
		}
		low, err := table.Cofactor(variable, false)
		if err != nil {
			return 0, err
		}
		high, err := table.Cofactor(variable, true)
		if err != nil {
			return 0, err
		}
		lowSupport, err := low.Support()
		if err != nil {
			return 0, err
		}
		highSupport, err := high.Support()
		if err != nil {
			return 0, err
		}
		lowResult, err := DauDSDDecompose(low, cofactorOptions)
		if err != nil {
			return 0, err
		}
		highResult, err := DauDSDDecompose(high, cofactorOptions)
		if err != nil {
			return 0, err
		}
		if lowResult.LargestPrime != 0 || highResult.LargestPrime != 0 {
			continue
		}
		support := len(lowSupport) + len(highSupport)
		if support < bestSupport {
			best, bestSupport = variable, support
		}
	}
	return best, nil
}

func dauDSDLevelPriority(expressions []dauDSDExpression, enabled bool) []int {
	if !enabled {
		return nil
	}
	priority := make([]int, len(expressions))
	costs := make([]int, len(expressions))
	for index := range priority {
		priority[index] = index
		costs[index] = -expressions[index].level
	}
	for index := 0; index+1 < len(priority); index++ {
		best := index
		for next := index + 1; next < len(priority); next++ {
			if costs[next] < costs[best] {
				best = next
			}
		}
		priority[index], priority[best] = priority[best], priority[index]
		costs[index], costs[best] = costs[best], costs[index]
	}
	return priority
}

func dauDSDCompoundExpression(text string, children ...dauDSDExpression) dauDSDExpression {
	result := dauDSDExpression{text: text, synthesized: true}
	for _, child := range children {
		level := child.level
		if child.synthesized {
			level++
		}
		result.level = max(result.level, level)
	}
	return result
}

func (d *dauDSDGeneralDecomposer) step(depth int) error {
	if d.options.Stop != nil && d.options.Stop() {
		return ErrCancelled
	}
	if depth >= d.options.MaxDepth {
		return ErrLimit
	}
	d.nodes++
	if d.nodes > d.options.MaxNodes {
		return ErrLimit
	}
	return nil
}

func (d *dauDSDGeneralDecomposer) checked(parts ...string) (string, error) {
	size := 0
	for _, part := range parts {
		var err error
		size, err = checkedSize(size, 1, len(part), d.options.MaxBytes)
		if err != nil {
			return "", err
		}
	}
	result := ""
	for _, part := range parts {
		result += part
	}
	return result, nil
}

func dauDSDRemovePosition(table *TruthTable, expressions []dauDSDExpression, position int) (*TruthTable, []dauDSDExpression, error) {
	if table == nil || position < 0 || position >= table.variables || len(expressions) != table.variables {
		return nil, nil, ErrDimension
	}
	last := table.variables - 1
	result := table
	owned := slices.Clone(expressions)
	var err error
	if position != last {
		result, err = result.SwapVariables(position, last)
		if err != nil {
			return nil, nil, err
		}
		owned[position], owned[last] = owned[last], owned[position]
	}
	keep := make([]int, last)
	for i := range keep {
		keep[i] = i
	}
	result, err = result.ProjectVariables(keep)
	if err != nil {
		return nil, nil, err
	}
	return result, owned[:last], nil
}

func dauDSDMinimumBase(table *TruthTable, expressions []dauDSDExpression) (*TruthTable, []dauDSDExpression, error) {
	result := table
	owned := slices.Clone(expressions)
	for variable := result.variables - 1; variable >= 0; variable-- {
		used, err := result.HasVariable(variable)
		if err != nil {
			return nil, nil, err
		}
		if used {
			continue
		}
		result, owned, err = dauDSDRemovePosition(result, owned, variable)
		if err != nil {
			return nil, nil, err
		}
	}
	return result, owned, nil
}

func dauDSDPairExpression(kind TruthPairOperator, first, second string) string {
	switch kind {
	case TruthPairAnd:
		return "(" + second + first + ")"
	case TruthPairNotFirstAnd:
		return "(" + second + "!" + first + ")"
	case TruthPairNotSecondAnd:
		return "(!" + second + first + ")"
	case TruthPairNotBothAnd:
		return "(!" + second + "!" + first + ")"
	case TruthPairXor:
		return "[" + second + first + "]"
	default:
		return ""
	}
}

func dauDSDPrimeFormula(table *TruthTable, expressions []dauDSDExpression, includeTruth bool) (string, error) {
	if table == nil || table.variables != len(expressions) || table.variables < 3 {
		return "", ErrInvariant
	}
	prefix := ""
	if includeTruth {
		prefix = dauTreePrimeDigits(table.words, table.variables)
	}
	result := prefix + "{"
	for _, expression := range expressions {
		result += expression.text
	}
	return result + "}", nil
}

func (d *dauDSDGeneralDecomposer) decompose(table *TruthTable, expressions []dauDSDExpression, depth int, allowSplit bool) (string, int, error) {
	if err := d.step(depth); err != nil {
		return "", 0, err
	}
	var err error
	table, expressions, err = dauDSDMinimumBase(table, expressions)
	if err != nil {
		return "", 0, err
	}
	if table.IsConstant(false) {
		return "0", 0, nil
	}
	if table.IsConstant(true) {
		return "1", 0, nil
	}
	if table.variables == 1 {
		value, err := table.Get(0)
		if err != nil {
			return "", 0, err
		}
		if value {
			formula, err := d.checked("!", expressions[0].text)
			return formula, 0, err
		}
		return expressions[0].text, 0, nil
	}

	// Dau_DsdDecomposeSingleVar scans from the highest current position.
	for variable := table.variables - 1; variable >= 0; variable-- {
		output, found, err := table.DecomposeVariable(variable)
		if err != nil {
			return "", 0, err
		}
		if !found {
			continue
		}
		leaf := expressions[variable].text
		remainderTable := output.Remainder
		if output.Operator == TruthOutputNotOr || output.Operator == TruthOutputOr {
			remainderTable = remainderTable.Not()
		}
		next, nextExpressions, err := dauDSDRemovePosition(remainderTable, expressions, variable)
		if err != nil {
			return "", 0, err
		}
		remainder, largest, err := d.decompose(next, nextExpressions, depth+1, allowSplit)
		if err != nil {
			return "", 0, err
		}
		var formula string
		switch output.Operator {
		case TruthOutputAnd:
			formula, err = d.checked("(", leaf, remainder, ")")
		case TruthOutputNotAnd:
			formula, err = d.checked("(!", leaf, remainder, ")")
		case TruthOutputNotOr:
			formula, err = d.checked("!(", leaf, remainder, ")")
		case TruthOutputOr:
			formula, err = d.checked("!(!", leaf, remainder, ")")
		case TruthOutputXor:
			formula, err = d.checked("[", leaf, remainder, "]")
		default:
			err = ErrInvariant
		}
		return formula, largest, err
	}

	// Dau_DsdDecomposeDoubleVars scans v then u in descending order and stores
	// the replacement in u's position.
	for second := table.variables - 1; second > 0; second-- {
		for first := second - 1; first >= 0; first-- {
			pair, found, err := table.DecomposeInputPair(first, second)
			if err != nil {
				return "", 0, err
			}
			if !found {
				continue
			}
			compound := dauDSDPairExpression(pair.Operator, expressions[first].text, expressions[second].text)
			if compound == "" {
				return "", 0, ErrInvariant
			}
			nextExpressions := slices.Clone(expressions)
			nextExpressions[first] = dauDSDCompoundExpression(compound, expressions[first], expressions[second])
			next, nextExpressions, err := dauDSDRemovePosition(pair.Remainder, nextExpressions, second)
			if err != nil {
				return "", 0, err
			}
			return d.decompose(next, nextExpressions, depth+1, allowSplit)
		}
	}

	// Dau_DsdDecomposeTripleVars first recognizes a control whose cofactor
	// supports do not overlap, then the one-exclusive-input mux case.
	for control := table.variables - 1; control >= 0; control-- {
		low, err := table.Cofactor(control, false)
		if err != nil {
			return "", 0, err
		}
		high, err := table.Cofactor(control, true)
		if err != nil {
			return "", 0, err
		}
		overlap := false
		lowOnly, highOnly := []int{}, []int{}
		for variable := 0; variable < table.variables; variable++ {
			if variable == control {
				continue
			}
			inLow, err := low.HasVariable(variable)
			if err != nil {
				return "", 0, err
			}
			inHigh, err := high.HasVariable(variable)
			if err != nil {
				return "", 0, err
			}
			overlap = overlap || inLow && inHigh
			if inLow && !inHigh {
				lowOnly = append(lowOnly, variable)
			}
			if inHigh && !inLow {
				highOnly = append(highOnly, variable)
			}
		}
		if !overlap {
			controlExpression := expressions[control].text
			highTable, highExpressions, err := dauDSDRemovePosition(high, expressions, control)
			if err != nil {
				return "", 0, err
			}
			lowTable, lowExpressions, err := dauDSDRemovePosition(low, expressions, control)
			if err != nil {
				return "", 0, err
			}
			highFormula, highPrime, err := d.decompose(highTable, highExpressions, depth+1, false)
			if err != nil {
				return "", 0, err
			}
			lowFormula, lowPrime, err := d.decompose(lowTable, lowExpressions, depth+1, false)
			if err != nil {
				return "", 0, err
			}
			formula, err := d.checked("<", controlExpression, highFormula, lowFormula, ">")
			return formula, max(highPrime, lowPrime), err
		}
		if len(lowOnly) == 1 && len(highOnly) == 1 {
			mux, found, err := table.DecomposeMuxInput(control)
			if err != nil {
				return "", 0, err
			}
			if !found {
				continue
			}
			compound := "<" + expressions[control].text + expressions[mux.WhenTrue].text
			if mux.ComplementFalse {
				compound += "!"
			}
			compound += expressions[mux.WhenFalse].text + ">"
			nextExpressions := slices.Clone(expressions)
			nextExpressions[control] = dauDSDCompoundExpression(compound, expressions[control], expressions[mux.WhenTrue], expressions[mux.WhenFalse])
			tags := make([]int, len(expressions))
			for i := range tags {
				tags[i] = i
			}
			next := mux.Remainder
			for _, removedTag := range []int{mux.WhenTrue, mux.WhenFalse} {
				position := slices.Index(tags, removedTag)
				if position < 0 {
					return "", 0, ErrInvariant
				}
				last := len(tags) - 1
				next, nextExpressions, err = dauDSDRemovePosition(next, nextExpressions, position)
				if err != nil {
					return "", 0, err
				}
				tags[position], tags[last] = tags[last], tags[position]
				tags = tags[:last]
			}
			return d.decompose(next, nextExpressions, depth+1, allowSplit)
		}
	}
	if allowSplit && d.options.SplitPrime {
		plainOptions := d.options
		plainOptions.SplitPrime = false
		priority := dauDSDLevelPriority(expressions, len(d.options.VariableLevels) != 0)
		plainOptions.VariableLevels = nil
		best, err := dauDSDOneStepVariable(table, plainOptions, priority)
		if err != nil {
			return "", 0, err
		}
		if best >= 0 {
			low, err := table.Cofactor(best, false)
			if err != nil {
				return "", 0, err
			}
			high, err := table.Cofactor(best, true)
			if err != nil {
				return "", 0, err
			}
			lowResult, err := DauDSDDecompose(low, plainOptions)
			if err != nil {
				return "", 0, err
			}
			highResult, err := DauDSDDecompose(high, plainOptions)
			if err != nil {
				return "", 0, err
			}
			if lowResult.LargestPrime != 0 || highResult.LargestPrime != 0 {
				return "", 0, ErrInvariant
			}
			bindings := ""
			for _, expression := range expressions {
				bindings += expression.text
			}
			formula, err := d.checked("<", string(rune('a'+best)), highResult.Formula, lowResult.Formula, ">{", bindings, "}")
			return formula, 0, err
		}
	}

	formula, err := dauDSDPrimeFormula(table, expressions, !d.options.OmitPrimeTruth)
	if err != nil {
		return "", 0, err
	}
	formula, err = d.checked(formula)
	if allowSplit && d.options.SplitPrime {
		return formula, -table.variables, err
	}
	return formula, table.variables, err
}
