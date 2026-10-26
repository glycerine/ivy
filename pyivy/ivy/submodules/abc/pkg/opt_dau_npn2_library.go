package abc

// Owned formula-library generation replacing Dtt_DumpLibrary's file output.

import (
	"fmt"
	"strings"

	"github.com/glycerine/insdict"
)

type DttFormulaOptions struct {
	MaxBytes, MaxDepth int
	MaxTransforms      uint64
	Stop               func() bool
}

func normalizeDttFormulaOptions(options DttFormulaOptions) (DttFormulaOptions, error) {
	if options.MaxBytes < 0 || options.MaxDepth < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 256
	}
	return options, nil
}

func formatDttOperation(kind uint8, first, second string, maximum int) (string, error) {
	var result string
	switch kind {
	case 0:
		result = "(" + first + "&" + second + ")"
	case 1:
		result = "(~" + first + "&" + second + ")"
	case 2:
		result = "(" + first + "&~" + second + ")"
	case 3:
		result = "~(~" + first + "&~" + second + ")"
	case 4:
		result = "(" + first + "^" + second + ")"
	case 5:
		result = "~(" + first + "&" + second + ")"
	case 6:
		result = "~(~" + first + "&" + second + ")"
	case 7:
		result = "~(" + first + "&~" + second + ")"
	case 8:
		result = "(~" + first + "&~" + second + ")"
	case 9:
		result = "~(" + first + "^" + second + ")"
	default:
		return "", ErrDimension
	}
	if len(result) > maximum {
		return "", ErrLimit
	}
	return result, nil
}

func (enumeration *DttEnumeration) FunctionFormula(index int, options DttFormulaOptions) (string, error) {
	if enumeration == nil || index < 0 || index >= len(enumeration.functions) {
		return "", ErrDimension
	}
	o, err := normalizeDttFormulaOptions(options)
	if err != nil {
		return "", err
	}
	if index == 0 {
		return "", ErrUnsupported
	}
	top, output, err := FindDttNPNTransform(enumeration.functions[index].Truth, enumeration.functions[index].Canonical, enumeration.variables, DttTransformSearchOptions{MaxTransforms: o.MaxTransforms, Stop: o.Stop})
	if err != nil {
		return "", err
	}
	var render func(int, DttNPNTransform, int) (string, error)
	render = func(functionIndex int, outer DttNPNTransform, depth int) (string, error) {
		if o.Stop != nil && o.Stop() {
			return "", ErrCancelled
		}
		if depth > o.MaxDepth {
			return "", ErrLimit
		}
		function := enumeration.functions[functionIndex]
		if functionIndex == 0 {
			return "", ErrUnsupported
		}
		if functionIndex == 1 {
			variable, complement, err := outer.Variable(0)
			if err != nil {
				return "", err
			}
			result := string(rune('a' + variable))
			if complement {
				result = "~" + result
			}
			return result, nil
		}
		leftTransform, err := ComposeDttNPNTransforms(function.LeftTransform, outer)
		if err != nil {
			return "", err
		}
		left, err := render(function.Fanin0, leftTransform, depth+1)
		if err != nil {
			return "", err
		}
		if function.LeftOutputComplement {
			left = "~" + left
		}
		right, err := render(function.Fanin1, outer, depth+1)
		if err != nil {
			return "", err
		}
		return formatDttOperation(function.Type, left, right, o.MaxBytes)
	}
	formula, err := render(index, top, 1)
	if err != nil {
		return "", err
	}
	if output {
		formula = "~" + formula
	}
	if len(formula) > o.MaxBytes {
		return "", ErrLimit
	}
	return formula, nil
}

type DttLibraryOptions struct {
	MaxFormulas, MaxBytes int
	Formula               DttFormulaOptions
}

type DttLibraryEntry struct {
	Function        int
	Truth           uint64
	Frontier, Nodes int
	Formula         string
}

type dttLibraryImplementation struct {
	function              int
	type_                 uint8
	leftClass, rightClass uint64
	leftTransform         DttNPNTransform
	rightTransform        DttNPNTransform
	leftComplement        bool
	rightComplement       bool
	outputComplement      bool
}

func orderedDttLibraryPair(first, second uint64) uint64 {
	if second < first {
		first, second = second, first
	}
	return first<<32 | second
}

func transformDttFormula(formula string, transform DttNPNTransform, complement bool, maximum int) (string, error) {
	var result strings.Builder
	for _, token := range formula {
		if token < 'a' || token >= 'a'+rune(transform.Variables()) {
			result.WriteRune(token)
			continue
		}
		variable, phase, err := transform.Variable(int(token - 'a'))
		if err != nil {
			return "", err
		}
		if phase {
			result.WriteByte('~')
		}
		result.WriteByte(byte('a' + variable))
		if result.Len() > maximum {
			return "", ErrLimit
		}
	}
	transformed := result.String()
	if complement {
		transformed = "~" + transformed
	}
	if len(transformed) > maximum {
		return "", ErrLimit
	}
	return transformed, nil
}

func (enumeration *DttEnumeration) libraryImplementations(options DttFormulaOptions) (*insdict.Dict[uint64, []dttLibraryImplementation], error) {
	result := insdict.NewDict[uint64, []dttLibraryImplementation]()
	seen := insdict.NewDict[uint64, *insdict.Dict[uint64, bool]]()
	mask := dauTruthMask(enumeration.variables)
	for index := 2; index < len(enumeration.functions); index++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		function := enumeration.functions[index]
		leftClass := enumeration.functions[function.Fanin0].Canonical
		rightClass := enumeration.functions[function.Fanin1].Canonical
		pairs, found := seen.Get2(function.Canonical)
		if !found {
			pairs = insdict.NewDict[uint64, bool]()
			seen.Set(function.Canonical, pairs)
		}
		pair := orderedDttLibraryPair(leftClass, rightClass)
		if _, found = pairs.Get2(pair); found {
			continue
		}
		pairs.Set(pair, true)

		leftTruth, err := function.LeftTransform.ApplyTruth(enumeration.functions[function.Fanin0].Truth)
		if err != nil {
			return nil, err
		}
		if function.LeftOutputComplement {
			leftTruth = (^leftTruth) & mask
		}
		leftTransform, leftComplement, err := FindDttNPNTransform(leftClass, leftTruth, enumeration.variables, DttTransformSearchOptions{MaxTransforms: options.MaxTransforms, Stop: options.Stop})
		if err != nil {
			return nil, err
		}
		rightTransform, rightComplement, err := FindDttNPNTransform(rightClass, enumeration.functions[function.Fanin1].Truth, enumeration.variables, DttTransformSearchOptions{MaxTransforms: options.MaxTransforms, Stop: options.Stop})
		if err != nil {
			return nil, err
		}
		topTransform, outputComplement, err := FindDttNPNTransform(function.Truth, function.Canonical, enumeration.variables, DttTransformSearchOptions{MaxTransforms: options.MaxTransforms, Stop: options.Stop})
		if err != nil {
			return nil, err
		}
		leftTransform, err = ComposeDttNPNTransforms(leftTransform, topTransform)
		if err != nil {
			return nil, err
		}
		rightTransform, err = ComposeDttNPNTransforms(rightTransform, topTransform)
		if err != nil {
			return nil, err
		}
		implementation := dttLibraryImplementation{
			function: index, type_: function.Type,
			leftClass: leftClass, rightClass: rightClass,
			leftTransform: leftTransform, rightTransform: rightTransform,
			leftComplement: leftComplement, rightComplement: rightComplement,
			outputComplement: outputComplement,
		}
		leftCheck, err := implementation.leftTransform.ApplyTruth(leftClass)
		if err != nil {
			return nil, err
		}
		if implementation.leftComplement {
			leftCheck = (^leftCheck) & mask
		}
		rightCheck, err := implementation.rightTransform.ApplyTruth(rightClass)
		if err != nil {
			return nil, err
		}
		if implementation.rightComplement {
			rightCheck = (^rightCheck) & mask
		}
		truthCheck, err := ApplyDttType(implementation.type_, leftCheck, rightCheck, mask)
		if err != nil {
			return nil, err
		}
		if implementation.outputComplement {
			truthCheck = (^truthCheck) & mask
		}
		if truthCheck != function.Canonical {
			return nil, ErrInvariant
		}
		implementations, _ := result.Get2(function.Canonical)
		result.Set(function.Canonical, append(implementations, implementation))
	}
	return result, nil
}

func (enumeration *DttEnumeration) GenerateLibrary(options DttLibraryOptions) ([]DttLibraryEntry, error) {
	if enumeration == nil || options.MaxFormulas < 0 || options.MaxBytes < 0 {
		return nil, ErrDimension
	}
	formulaOptions, err := normalizeDttFormulaOptions(options.Formula)
	if err != nil {
		return nil, err
	}
	maximum := options.MaxFormulas
	if maximum == 0 {
		maximum = 1 << 20
	}
	maximumBytes := options.MaxBytes
	if maximumBytes == 0 {
		maximumBytes = 1 << 24
	}
	implementations, err := enumeration.libraryImplementations(formulaOptions)
	if err != nil {
		return nil, err
	}
	result := make([]DttLibraryEntry, 0, min(maximum, max(0, len(enumeration.functions)-2)))
	totalBytes := 0
	variableClass := enumeration.functions[1].Canonical
	variableFormula, err := enumeration.FunctionFormula(1, formulaOptions)
	if err != nil {
		return nil, err
	}
	var expand func(uint64, int, func(string) error) error
	var renderImplementation func(dttLibraryImplementation, int, func(string) error) error
	renderImplementation = func(implementation dttLibraryImplementation, depth int, emit func(string) error) error {
		return expand(implementation.leftClass, depth+1, func(leftCanonical string) error {
			left, err := transformDttFormula(leftCanonical, implementation.leftTransform, implementation.leftComplement, formulaOptions.MaxBytes)
			if err != nil {
				return err
			}
			return expand(implementation.rightClass, depth+1, func(rightCanonical string) error {
				right, err := transformDttFormula(rightCanonical, implementation.rightTransform, implementation.rightComplement, formulaOptions.MaxBytes)
				if err != nil {
					return err
				}
				formula, err := formatDttOperation(implementation.type_, left, right, formulaOptions.MaxBytes)
				if err != nil {
					return err
				}
				if implementation.outputComplement {
					formula = "~" + formula
				}
				if len(formula) > formulaOptions.MaxBytes {
					return ErrLimit
				}
				return emit(formula)
			})
		})
	}
	expand = func(class uint64, depth int, emit func(string) error) error {
		if formulaOptions.Stop != nil && formulaOptions.Stop() {
			return ErrCancelled
		}
		if depth > formulaOptions.MaxDepth {
			return ErrLimit
		}
		if class == variableClass {
			return emit(variableFormula)
		}
		choices, found := implementations.Get2(class)
		if !found || len(choices) == 0 {
			return ErrInvariant
		}
		for _, implementation := range choices {
			if err := renderImplementation(implementation, depth, emit); err != nil {
				return err
			}
		}
		return nil
	}
	for _, class := range enumeration.classes {
		choices, found := implementations.Get2(class.Canonical)
		if !found {
			continue
		}
		for _, root := range choices {
			function := enumeration.functions[root.function]
			err = renderImplementation(root, 1, func(formula string) error {
				if len(result) >= maximum {
					return ErrLimit
				}
				if totalBytes > maximumBytes-len(formula) {
					return ErrLimit
				}
				totalBytes += len(formula)
				result = append(result, DttLibraryEntry{Function: root.function, Truth: class.Canonical, Frontier: function.Frontier, Nodes: function.Nodes, Formula: formula})
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func FormatDttLibrary(entries []DttLibraryEntry, maxBytes int) (string, error) {
	if maxBytes < 0 {
		return "", ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 1 << 24
	}
	var result strings.Builder
	for _, entry := range entries {
		line := fmt.Sprintf("%08x = %s\n", entry.Truth, entry.Formula)
		if result.Len() > maxBytes-len(line) {
			return "", ErrLimit
		}
		result.WriteString(line)
	}
	return result.String(), nil
}
