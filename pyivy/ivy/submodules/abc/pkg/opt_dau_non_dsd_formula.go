package abc

// Deterministic DSD formula derivation for non-DSD decompositions.

type DauSetFormulaOptions struct {
	MaxBytes, MaxNodes, MaxDepth, MaxWords int
	Stop                                   func() bool
}

type DauSetFormulas struct {
	Composition, Decomposition string
	SelectorVariable           int
}

func normalizeDauSetFormulaOptions(options DauSetFormulaOptions) (DauSetFormulaOptions, error) {
	if options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxDepth < 0 || options.MaxWords < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 20
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 256
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 20
	}
	return options, nil
}

func dauSetFormulaMergeOptions(options DauSetFormulaOptions) DauMergeOptions {
	return DauMergeOptions{
		MaxBytes:     options.MaxBytes,
		MaxNodes:     options.MaxNodes,
		MaxDepth:     options.MaxDepth,
		MaxVariables: 26,
		MaxWords:     options.MaxWords,
		Stop:         options.Stop,
	}
}

func remapDauSetFormula(text string, mapping []int, options DauSetFormulaOptions) (string, error) {
	mergeOptions := dauSetFormulaMergeOptions(options)
	formula, err := ParseDauDSD(text, dauMergeDSDOptions(mergeOptions))
	if err != nil {
		return "", err
	}
	formula, err = dauDSDMergeRemap(formula, mapping, options.Stop)
	if err != nil {
		return "", err
	}
	result := formula.Normalized()
	if len(result) > options.MaxBytes {
		return "", ErrLimit
	}
	return result, nil
}

// DSDFormulas derives deterministic formulas from the owned truth tables.
// Composition uses original free/shared variable letters and the synthetic
// selector letter 'a'+Variables; decomposition uses original unique/shared
// letters. Both components use the general bounded decomposer through twelve
// inputs; irreducible components retain their exact prime truth tables.
func (result DauSetDecomposition) DSDFormulas(options DauSetFormulaOptions) (DauSetFormulas, error) {
	o, err := normalizeDauSetFormulaOptions(options)
	if err != nil {
		return DauSetFormulas{}, err
	}
	if !result.Found || result.Composition == nil || result.Decomposition == nil || result.Variables < 3 || result.Variables >= 26 {
		return DauSetFormulas{}, ErrDimension
	}
	if _, err = result.Recompose(o.MaxWords, o.Stop); err != nil {
		return DauSetFormulas{}, err
	}
	mergeOptions := dauSetFormulaMergeOptions(o)
	composition, err := decomposeDauMergeTable(result.Composition, mergeOptions)
	if err != nil {
		return DauSetFormulas{}, err
	}
	decomposition, err := decomposeDauMergeTable(result.Decomposition, mergeOptions)
	if err != nil {
		return DauSetFormulas{}, err
	}
	compositionMapping := make([]int, 0, len(result.FreeVariables)+len(result.SharedVariables)+1)
	compositionMapping = append(compositionMapping, result.FreeVariables...)
	compositionMapping = append(compositionMapping, result.SharedVariables...)
	compositionMapping = append(compositionMapping, result.Variables)
	composition, err = remapDauSetFormula(composition, compositionMapping, o)
	if err != nil {
		return DauSetFormulas{}, err
	}
	decompositionMapping := make([]int, 0, len(result.UniqueVariables)+len(result.SharedVariables))
	decompositionMapping = append(decompositionMapping, result.UniqueVariables...)
	decompositionMapping = append(decompositionMapping, result.SharedVariables...)
	decomposition, err = remapDauSetFormula(decomposition, decompositionMapping, o)
	if err != nil {
		return DauSetFormulas{}, err
	}
	return DauSetFormulas{Composition: composition, Decomposition: decomposition, SelectorVariable: result.Variables}, nil
}
