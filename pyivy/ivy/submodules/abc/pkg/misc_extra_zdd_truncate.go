package abc

// Finite ZDD subset-family truncation from src/misc/extra/extraZddTrunc.c.

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
)

type ExtraZDDLimits struct {
	MaxVariables, MaxSubsets, MaxElements, MaxOutputElements int
	Stop                                                     func() bool
}

func normalizeExtraZDDLimits(limits ExtraZDDLimits) (ExtraZDDLimits, error) {
	if limits.MaxVariables < 0 || limits.MaxSubsets < 0 || limits.MaxElements < 0 || limits.MaxOutputElements < 0 {
		return limits, ErrDimension
	}
	if limits.MaxVariables == 0 {
		limits.MaxVariables = 1 << 20
	}
	if limits.MaxSubsets == 0 {
		limits.MaxSubsets = 1 << 16
	}
	if limits.MaxElements == 0 {
		limits.MaxElements = 1 << 22
	}
	if limits.MaxOutputElements == 0 {
		limits.MaxOutputElements = 1 << 22
	}
	return limits, nil
}

// ExtraZDD is an immutable, explicitly represented family of finite subsets.
// Its order is the canonical ZDD traversal order: exclude before include at
// each increasing variable index.
type ExtraZDD struct {
	variables int
	subsets   [][]int
}

func extraZDDSubsetLess(first, second []int) bool {
	i, j := 0, 0
	for i < len(first) && j < len(second) {
		if first[i] == second[j] {
			i++
			j++
			continue
		}
		if first[i] < second[j] {
			return false
		}
		return true
	}
	return i == len(first) && j < len(second)
}

func extraZDDSubsetKey(subset []int) string {
	var builder strings.Builder
	for _, variable := range subset {
		fmt.Fprintf(&builder, "%d,", variable)
	}
	return builder.String()
}

func NewExtraZDD(variables int, subsets [][]int, limits ExtraZDDLimits) (*ExtraZDD, error) {
	limits, e := normalizeExtraZDDLimits(limits)
	if e != nil {
		return nil, e
	}
	if variables < 0 {
		return nil, ErrDimension
	}
	if variables > limits.MaxVariables || len(subsets) > limits.MaxSubsets {
		return nil, ErrLimit
	}
	result := &ExtraZDD{variables: variables, subsets: make([][]int, 0, len(subsets))}
	seen := make(map[string]struct{}, len(subsets))
	elements := 0
	for _, input := range subsets {
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		subset := slices.Clone(input)
		sort.Ints(subset)
		write := 0
		for _, variable := range subset {
			if variable < 0 || variable >= variables {
				return nil, ErrDimension
			}
			if write == 0 || subset[write-1] != variable {
				subset[write] = variable
				write++
			}
		}
		subset = subset[:write]
		key := extraZDDSubsetKey(subset)
		if _, duplicate := seen[key]; duplicate {
			continue
		}
		seen[key] = struct{}{}
		elements += len(subset)
		if elements > limits.MaxElements {
			return nil, ErrLimit
		}
		result.subsets = append(result.subsets, subset)
	}
	sort.Slice(result.subsets, func(i, j int) bool {
		return extraZDDSubsetLess(result.subsets[i], result.subsets[j])
	})
	return result, nil
}

func NewExtraZDDVariable(variables, variable int, limits ExtraZDDLimits) (*ExtraZDD, error) {
	if variable < 0 || variable >= variables {
		return nil, ErrDimension
	}
	return NewExtraZDD(variables, [][]int{{variable}}, limits)
}

func (z *ExtraZDD) Variables() int {
	if z == nil {
		return 0
	}
	return z.variables
}

func (z *ExtraZDD) Subsets() [][]int {
	if z == nil {
		return nil
	}
	result := make([][]int, len(z.subsets))
	for i := range result {
		result[i] = slices.Clone(z.subsets[i])
	}
	return result
}

type ExtraZDDTruncation struct {
	Subsets [][]int
	Flat    []int
}

func flattenExtraZDDSubsets(subsets [][]int, maxElements int) ([]int, error) {
	length := 1 + len(subsets)
	for _, subset := range subsets {
		if len(subset) > maxElements-length {
			return nil, ErrLimit
		}
		length += len(subset)
	}
	if length > maxElements {
		return nil, ErrLimit
	}
	flat := make([]int, 1, length)
	flat[0] = len(subsets)
	for _, subset := range subsets {
		flat = append(flat, subset...)
		flat = append(flat, -1)
	}
	return flat, nil
}

// Truncate keeps subsets whose product of included-variable probabilities is
// at least limit. Excluded variables do not affect the product, matching the
// source recursion rather than an independent-event probability measure.
func (z *ExtraZDD) Truncate(probabilities []float64, limit float64, limits ExtraZDDLimits) (ExtraZDDTruncation, error) {
	fail := func(e error) (ExtraZDDTruncation, error) { return ExtraZDDTruncation{}, e }
	limits, e := normalizeExtraZDDLimits(limits)
	if e != nil {
		return fail(e)
	}
	if z == nil || z.variables < 0 || len(probabilities) < z.variables || !(limit > 0 && limit <= 1) || math.IsNaN(limit) {
		return fail(ErrDimension)
	}
	for variable := 0; variable < z.variables; variable++ {
		probability := probabilities[variable]
		if !(probability > 0 && probability <= 1) || math.IsNaN(probability) {
			return fail(ErrDimension)
		}
	}
	if z.variables > limits.MaxVariables || len(z.subsets) > limits.MaxSubsets {
		return fail(ErrLimit)
	}
	result := make([][]int, 0, len(z.subsets))
	elements := 0
	for _, subset := range z.subsets {
		if limits.Stop != nil && limits.Stop() {
			return fail(ErrCancelled)
		}
		probability := 1.0
		for _, variable := range subset {
			probability *= probabilities[variable]
			if probability < limit {
				break
			}
		}
		if probability < limit {
			continue
		}
		elements += len(subset)
		if elements > limits.MaxElements {
			return fail(ErrLimit)
		}
		result = append(result, slices.Clone(subset))
	}
	flat, e := flattenExtraZDDSubsets(result, limits.MaxOutputElements)
	if e != nil {
		return fail(e)
	}
	return ExtraZDDTruncation{Subsets: result, Flat: flat}, nil
}

func FormatExtraZDDSubsets(subsets [][]int) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "The set contains %d subsets:\n", len(subsets))
	for index, subset := range subsets {
		fmt.Fprintf(&builder, "Subset %3d : {", index)
		for _, variable := range subset {
			fmt.Fprintf(&builder, " %d", variable)
		}
		builder.WriteString(" }\n")
	}
	return builder.String()
}

type ExtraZDDExample struct {
	Initial   [][]int
	Result    ExtraZDDTruncation
	Formatted string
}

func ExtraZDDTruncateExample() (ExtraZDDExample, error) {
	initial := [][]int{{0, 3, 5}, {1, 2, 3, 6, 9}, {1, 5, 7, 8}, {2, 4}, {0, 5, 6, 9}}
	zdd, e := NewExtraZDD(10, initial, ExtraZDDLimits{})
	if e != nil {
		return ExtraZDDExample{}, e
	}
	probabilities := make([]float64, 10)
	for i := range probabilities {
		probabilities[i] = 0.1
	}
	result, e := zdd.Truncate(probabilities, 0.001, ExtraZDDLimits{})
	if e != nil {
		return ExtraZDDExample{}, e
	}
	return ExtraZDDExample{Initial: zdd.Subsets(), Result: result, Formatted: FormatExtraZDDSubsets(result.Subsets)}, nil
}
