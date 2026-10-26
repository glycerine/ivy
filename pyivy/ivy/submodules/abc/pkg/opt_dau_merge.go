package abc

// Owned two-formula DSD merging from opt/dau/dauMerge.c.

import (
	"fmt"
	"strings"
)

type DauMergeOptions struct {
	MaxBytes, MaxNodes, MaxDepth, MaxVariables, MaxWords int
	Stop                                                 func() bool
}

func normalizeDauMergeOptions(options DauMergeOptions) (DauMergeOptions, error) {
	if options.MaxBytes < 0 || options.MaxNodes < 0 || options.MaxDepth < 0 || options.MaxVariables < 0 || options.MaxWords < 0 {
		return options, ErrDimension
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 2000
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 2000
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 256
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 12
	}
	if options.MaxVariables > 26 {
		return options, ErrDimension
	}
	return options, nil
}

func dauMergeDSDOptions(options DauMergeOptions) DauDSDOptions {
	return DauDSDOptions{
		MaxBytes:     options.MaxBytes,
		MaxNodes:     options.MaxNodes,
		MaxDepth:     options.MaxDepth,
		MaxVariables: options.MaxVariables,
		MaxWords:     options.MaxWords,
		Stop:         options.Stop,
	}
}

func checkDauMergePermutation(mapping []int, variables int) error {
	if len(mapping) != variables {
		return ErrDimension
	}
	seen := make([]bool, variables)
	for _, target := range mapping {
		if target < 0 || target >= variables || seen[target] {
			return ErrDimension
		}
		seen[target] = true
	}
	return nil
}

func complementDauMergeFormula(formula *DauDSDFormula, complement bool) {
	if !complement {
		return
	}
	node := &formula.nodes[formula.root]
	if node.kind == DauDSDConstant {
		node.value ^= 1
		return
	}
	node.complement = !node.complement
}

func decomposeDauMergeTable(table *TruthTable, options DauMergeOptions) (string, error) {
	result, err := DauDSDDecompose(table, DauDSDDecomposeOptions{
		MaxBytes: options.MaxBytes, MaxNodes: options.MaxNodes, MaxDepth: options.MaxDepth,
		MaxWords: options.MaxWords, Stop: options.Stop,
	})
	if err != nil {
		return "", err
	}
	return result.Formula, nil
}

func dauDSDMergeInline(formula *DauDSDFormula, index, shared int, definitions []string, stop func() bool) (string, error) {
	if stop != nil && stop() {
		return "", ErrCancelled
	}
	node := &formula.nodes[index]
	if node.kind == DauDSDVariable {
		if node.value < shared {
			return formula.render(index, false), nil
		}
		definition := node.value - shared
		if definition < 0 || definition >= len(definitions) {
			return "", ErrInvariant
		}
		if node.complement {
			return "!" + definitions[definition], nil
		}
		return definitions[definition], nil
	}
	if node.kind == DauDSDConstant {
		return formula.render(index, false), nil
	}
	prefix, suffix, children := dauDSDMergeGroupSpelling(formula, index)
	parts := make([]string, len(children))
	for i, child := range children {
		part, err := dauDSDMergeInline(formula, child, shared, definitions, stop)
		if err != nil {
			return "", err
		}
		parts[i] = part
	}
	return prefix + strings.Join(parts, "") + suffix, nil
}

func normalizeDauMergedText(text string, options DauMergeOptions) (string, error) {
	simplified, err := DauDSDRemoveBraces(text, dauMergeDSDOptions(options))
	if err != nil {
		return "", err
	}
	formula, err := ParseDauDSD(simplified, dauMergeDSDOptions(options))
	if err != nil {
		return "", err
	}
	result := formula.Normalized()
	if len(result) > options.MaxBytes {
		return "", ErrLimit
	}
	return result, nil
}

// DauDSDMerge returns an owned DSD formula for the conjunction of first and
// second after applying their source-to-global variable permutations and
// optional output complements. The source's shared-variable substitution
// pipeline is retained and reduced conjunctions use the general DSD decomposer
// throughout the supported variable domain.
func DauDSDMerge(first string, firstPermutation []int, second string, secondPermutation []int, complementFirst, complementSecond bool, variables int, options DauMergeOptions) (string, error) {
	o, err := normalizeDauMergeOptions(options)
	if err != nil {
		return "", err
	}
	if variables < 0 || variables > o.MaxVariables || variables > 26 {
		return "", ErrDimension
	}
	if err = checkDauMergePermutation(firstPermutation, variables); err != nil {
		return "", err
	}
	if err = checkDauMergePermutation(secondPermutation, variables); err != nil {
		return "", err
	}
	if o.Stop != nil && o.Stop() {
		return "", ErrCancelled
	}
	dsdOptions := dauMergeDSDOptions(o)
	left, err := ParseDauDSD(first, dsdOptions)
	if err != nil {
		return "", fmt.Errorf("first DSD: %w", err)
	}
	right, err := ParseDauDSD(second, dsdOptions)
	if err != nil {
		return "", fmt.Errorf("second DSD: %w", err)
	}
	complementDauMergeFormula(left, complementFirst)
	complementDauMergeFormula(right, complementSecond)
	left, err = dauDSDMergeRemap(left, firstPermutation, o.Stop)
	if err != nil {
		return "", fmt.Errorf("first permutation: %w", err)
	}
	right, err = dauDSDMergeRemap(right, secondPermutation, o.Stop)
	if err != nil {
		return "", fmt.Errorf("second permutation: %w", err)
	}
	leftNode, rightNode := left.nodes[left.root], right.nodes[right.root]
	if leftNode.kind == DauDSDConstant {
		if leftNode.value == 0 {
			return "0", nil
		}
		return normalizeDauMergedText(right.String(), o)
	}
	if rightNode.kind == DauDSDConstant {
		if rightNode.value == 0 {
			return "0", nil
		}
		return normalizeDauMergedText(left.String(), o)
	}
	leftPresence, err := dauDSDMergePresence(left, variables, o.Stop)
	if err != nil {
		return "", err
	}
	rightPresence, err := dauDSDMergePresence(right, variables, o.Stop)
	if err != nil {
		return "", err
	}
	shared, oldToNew, newToOld, err := dauDSDMergeMaps(leftPresence, rightPresence)
	if err != nil {
		return "", err
	}
	if shared == 0 {
		combined := "(" + left.String() + right.String() + ")"
		if len(combined) > o.MaxBytes {
			return "", ErrLimit
		}
		return normalizeDauMergedText(combined, o)
	}
	left, err = dauDSDMergeRemap(left, oldToNew, o.Stop)
	if err != nil {
		return "", err
	}
	right, err = dauDSDMergeRemap(right, oldToNew, o.Stop)
	if err != nil {
		return "", err
	}
	leftStatus, err := dauDSDMergeStatuses(left, shared, o.Stop)
	if err != nil {
		return "", err
	}
	rightStatus, err := dauDSDMergeStatuses(right, shared, o.Stop)
	if err != nil {
		return "", err
	}
	definitions := &dauDSDMergeDefinitions{shared: shared, maxVariables: o.MaxVariables, maxBytes: o.MaxBytes}
	reducedLeft, err := dauDSDMergeExtract(left, leftStatus, left.root, definitions, o.Stop)
	if err != nil {
		return "", err
	}
	reducedRight, err := dauDSDMergeExtract(right, rightStatus, right.root, definitions, o.Stop)
	if err != nil {
		return "", err
	}
	reducedText := "(" + reducedLeft + reducedRight + ")"
	if len(reducedText) > o.MaxBytes {
		return "", ErrLimit
	}
	reduced, err := ParseDauDSD(reducedText, dsdOptions)
	if err != nil {
		return "", err
	}
	reducedVariables := shared + len(definitions.values)
	reducedTruth, err := reduced.TruthTable(reducedVariables, o.MaxWords, o.Stop)
	if err != nil {
		return "", err
	}
	decomposedText, err := decomposeDauMergeTable(reducedTruth, o)
	if err != nil {
		return "", err
	}
	decomposed, err := ParseDauDSD(decomposedText, dsdOptions)
	if err != nil {
		return "", err
	}
	inlined, err := dauDSDMergeInline(decomposed, decomposed.root, shared, definitions.values, o.Stop)
	if err != nil {
		return "", err
	}
	inlined, err = DauDSDRemoveBraces(inlined, dsdOptions)
	if err != nil {
		return "", err
	}
	inlinedFormula, err := ParseDauDSD(inlined, dsdOptions)
	if err != nil {
		return "", err
	}
	restored, err := dauDSDMergeRemap(inlinedFormula, newToOld, o.Stop)
	if err != nil {
		return "", err
	}
	return normalizeDauMergedText(restored.String(), o)
}
