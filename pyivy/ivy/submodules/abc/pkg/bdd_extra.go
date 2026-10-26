package abc

// Stand-alone decision-diagram utilities from src/bdd/extrab/extraBdd.h,
// extraBddAuto.c, extraBddCas.c, extraBddImage.c, extraBddKmap.c,
// extraBddMaxMin.c, extraBddMisc.c, extraBddSet.c, extraBddSymm.c,
// extraBddThresh.c, extraBddTime.c, and extraBddUnate.c.
//
// This file ports the concrete truth, family, transfer, image, symmetry,
// unateness, threshold, and formatting semantics that stand alone. The
// automorphism-space algorithms, column encoders, native image-tree schedule,
// pointer-cache implementation, dynamic reorder restarts, and stream side
// effects remain outside this fixed-order owned representation.

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

var extraBDDClusterSources = [...]string{
	"extraBddAuto.c", "extraBddCas.c", "extraBddImage.c", "extraBddKmap.c",
	"extraBddMaxMin.c", "extraBddMisc.c", "extraBddSet.c", "extraBddSymm.c",
	"extraBddThresh.c", "extraBddTime.c", "extraBddUnate.c",
}

func ExtraBDDClusterBuildSources() []string {
	result := make([]string, len(extraBDDClusterSources))
	copy(result, extraBDDClusterSources[:])
	return result
}

func extraBDDValidateGraph(manager *ExtraBDDManager, roots []ExtraBDDLit, work *extraBDDWork) error {
	if manager == nil {
		return ErrDimension
	}
	if len(manager.nodes) > work.options.MaxBytes/2 {
		return ErrLimit
	}
	state := make([]uint8, len(manager.nodes))
	var visit func(ExtraBDDLit, int) error
	visit = func(root ExtraBDDLit, parent int) error {
		if err := work.step(); err != nil {
			return err
		}
		id := root.ID()
		if uint64(id) >= uint64(len(manager.nodes)) {
			return ErrInvalidLiteral
		}
		if id == 0 {
			return nil
		}
		node := manager.nodes[id]
		variable := int(node.variable)
		if variable >= manager.variables || variable <= parent || node.then == node.otherwise || node.otherwise.Complemented() {
			return ErrInvariant
		}
		if state[id] == 1 {
			return ErrInvariant
		}
		if state[id] == 2 {
			return nil
		}
		state[id] = 1
		if err := visit(node.then, variable); err != nil {
			return err
		}
		if err := visit(node.otherwise, variable); err != nil {
			return err
		}
		state[id] = 2
		return nil
	}
	for _, root := range roots {
		if err := visit(root, -1); err != nil {
			return err
		}
	}
	return nil
}

// ExtraTranslationSpace returns the set of XOR translations that leave root
// invariant. It is the owned fixed-order form of Extra_bddSpaceFromFunctionFast.
func (manager *ExtraBDDManager) ExtraTranslationSpace(root ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.analysisRun([]ExtraBDDLit{root}, 3, options, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		if err := extraBDDValidateGraph(manager, []ExtraBDDLit{root}, work); err != nil {
			return 0, err
		}
		values, err := manager.analysisDense(root, count, work)
		if err != nil {
			return 0, err
		}
		translations := make([]bool, count)
		for shift := 0; shift < count; shift++ {
			invariant := true
			for assignment := 0; assignment < count; assignment++ {
				if err := work.step(); err != nil {
					return 0, err
				}
				if values[assignment] != values[assignment^shift] {
					invariant = false
					break
				}
			}
			translations[shift] = invariant
		}
		return manager.analysisBuild(translations, work)
	})
}

// ExtraProfileWidth counts distinct regular nodes crossing each level cut.
// Complemented edges share their regular node, extending the source routine
// (which only accepted uncomplemented diagrams) to this concrete manager.
func (manager *ExtraBDDManager) ExtraProfileWidth(root ExtraBDDLit, cutLevel int, options ExtraBDDOptions) ([]int, int, error) {
	if err := manager.checkLiteral(root); err != nil {
		return nil, 0, err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return nil, 0, err
	}
	if cutLevel < -1 || cutLevel > manager.variables {
		return nil, 0, ErrDimension
	}
	if len(manager.nodes) > options.MaxBytes/16 {
		return nil, 0, ErrLimit
	}
	work := extraBDDWork{options: options}
	if err := extraBDDValidateGraph(manager, []ExtraBDDLit{root}, &work); err != nil {
		return nil, 0, err
	}
	top := make([]int, len(manager.nodes))
	for index := range top {
		top[index] = manager.variables + 1
	}
	top[root.ID()] = 0
	seen := make([]bool, len(manager.nodes))
	stack := []ExtraBDDLit{root}
	for len(stack) != 0 {
		if err := work.step(); err != nil {
			return nil, 0, err
		}
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		id := current.ID()
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		level := manager.variableOf(current) + 1
		for _, child := range [...]ExtraBDDLit{manager.elseOf(current), manager.thenOf(current)} {
			childID := child.ID()
			if level < top[childID] {
				top[childID] = level
			}
			stack = append(stack, child)
		}
	}
	profile := make([]int, manager.variables+1)
	for id, start := range top {
		if start > manager.variables {
			continue
		}
		limit := manager.variables
		if id != 0 {
			limit = int(manager.nodes[id].variable)
		}
		for level := start; level <= limit; level++ {
			if err := work.step(); err != nil {
				return nil, 0, err
			}
			profile[level]++
		}
	}
	limit := len(profile)
	if cutLevel > 0 {
		limit = cutLevel
	}
	maximum := 0
	for _, width := range profile[:limit] {
		if width > maximum {
			maximum = width
		}
	}
	return profile, maximum, nil
}

func (manager *ExtraBDDManager) ExtraSupport(root ExtraBDDLit, options ExtraBDDOptions) ([]int, error) {
	if err := manager.checkLiteral(root); err != nil {
		return nil, err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return nil, err
	}
	if manager.variables > options.MaxBytes/2 || len(manager.nodes) > options.MaxBytes/9 {
		return nil, ErrLimit
	}
	work := extraBDDWork{options: options}
	if err := extraBDDValidateGraph(manager, []ExtraBDDLit{root}, &work); err != nil {
		return nil, err
	}
	present := make([]bool, manager.variables)
	seen := make([]bool, len(manager.nodes))
	stack := []ExtraBDDLit{root}
	for len(stack) != 0 {
		if err := work.step(); err != nil {
			return nil, err
		}
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		id := current.ID()
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		variable := manager.variableOf(current)
		if variable < 0 || variable >= manager.variables {
			return nil, ErrInvariant
		}
		present[variable] = true
		stack = append(stack, manager.thenOf(current), manager.elseOf(current))
	}
	result := make([]int, 0, manager.variables)
	for variable, found := range present {
		if found {
			result = append(result, variable)
		}
	}
	return result, nil
}

func (manager *ExtraBDDManager) ExtraOneCube(root ExtraBDDLit, options ExtraBDDOptions) (CuddCube, error) {
	if err := manager.checkLiteral(root); err != nil {
		return CuddCube{}, err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return CuddCube{}, err
	}
	work := extraBDDWork{options: options}
	if err := extraBDDValidateGraph(manager, []ExtraBDDLit{root}, &work); err != nil {
		return CuddCube{}, err
	}
	if root == ExtraBDDFalse {
		return CuddCube{}, nil
	}
	if manager.variables > options.MaxBytes/16 {
		return CuddCube{}, ErrLimit
	}
	assignments := make([]ExtraBDDAssignment, 0, manager.variables)
	current := root
	for current >= 2 {
		if err := work.step(); err != nil {
			return CuddCube{}, err
		}
		variable := manager.variableOf(current)
		low, high := manager.elseOf(current), manager.thenOf(current)
		value := false
		if low == ExtraBDDFalse {
			value, current = true, high
		} else {
			current = low
		}
		assignments = append(assignments, ExtraBDDAssignment{Variable: variable, Value: value})
	}
	if current != ExtraBDDTrue {
		return CuddCube{}, ErrInvariant
	}
	dontCares := manager.variables - len(assignments)
	if dontCares >= 64 {
		return CuddCube{}, ErrLimit
	}
	transaction := manager.beginTransaction()
	work.transaction = &transaction
	cube := ExtraBDDTrue
	for index := len(assignments) - 1; index >= 0; index-- {
		item := assignments[index]
		var createErr error
		if item.Value {
			cube, createErr = manager.uniqueCreate(item.Variable, cube, ExtraBDDFalse, &transaction)
		} else {
			cube, createErr = manager.uniqueCreate(item.Variable, ExtraBDDFalse, cube, &transaction)
		}
		if createErr != nil {
			manager.rollback(&transaction)
			return CuddCube{}, createErr
		}
	}
	return CuddCube{Assignments: assignments, Root: cube, Exists: true, Minterms: uint64(1) << uint(dontCares)}, nil
}

func (manager *ExtraBDDManager) ExtraRangeCube(start, stop int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if start < 0 || stop < start || stop > manager.variables {
		return 0, ErrDimension
	}
	assignments := make([]ExtraBDDAssignment, stop-start)
	for index := range assignments {
		assignments[index] = ExtraBDDAssignment{Variable: start + index, Value: true}
	}
	return manager.extraCube(assignments, options)
}

func (manager *ExtraBDDManager) ExtraBitsCube(code uint64, variables []int, mostSignificantFirst bool, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if len(variables) > 64 {
		return 0, ErrDimension
	}
	assignments := make([]ExtraBDDAssignment, len(variables))
	seen := make([]bool, manager.variables)
	for index, variable := range variables {
		if variable < 0 || variable >= manager.variables || seen[variable] {
			return 0, ErrDimension
		}
		seen[variable] = true
		bit := index
		if mostSignificantFirst {
			bit = len(variables) - 1 - index
		}
		assignments[index] = ExtraBDDAssignment{Variable: variable, Value: code&(uint64(1)<<uint(bit)) != 0}
	}
	sort.Slice(assignments, func(i, j int) bool { return assignments[i].Variable < assignments[j].Variable })
	return manager.extraCube(assignments, options)
}

func (manager *ExtraBDDManager) extraCube(assignments []ExtraBDDAssignment, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.runBooleanOperation(nil, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		result := ExtraBDDTrue
		for index := len(assignments) - 1; index >= 0; index-- {
			if err := work.step(); err != nil {
				return 0, err
			}
			item := assignments[index]
			if item.Variable < 0 || item.Variable >= manager.variables || index+1 < len(assignments) && item.Variable >= assignments[index+1].Variable {
				return 0, ErrDimension
			}
			var err error
			if item.Value {
				result, err = manager.uniqueCreate(item.Variable, result, ExtraBDDFalse, work.transaction)
			} else {
				result, err = manager.uniqueCreate(item.Variable, ExtraBDDFalse, result, work.transaction)
			}
			if err != nil {
				return 0, err
			}
		}
		return result, nil
	})
}

func (manager *ExtraBDDManager) ExtraNaryAnd(count int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if count < 0 || count > manager.variables {
		return 0, ErrDimension
	}
	if count == 0 {
		return ExtraBDDTrue, manager.check()
	}
	return manager.ExtraRangeCube(0, count, options)
}

func (manager *ExtraBDDManager) ExtraNaryOr(count int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if count < 0 || count > manager.variables {
		return 0, ErrDimension
	}
	if count == 0 {
		return ExtraBDDFalse, nil
	}
	assignments := make([]ExtraBDDAssignment, count)
	for variable := range assignments {
		assignments[variable] = ExtraBDDAssignment{Variable: variable, Value: false}
	}
	result, err := manager.extraCube(assignments, options)
	if err != nil {
		return 0, err
	}
	return result.Not(), nil
}

func (manager *ExtraBDDManager) ExtraNaryXor(count int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if count < 0 || count > manager.variables {
		return 0, ErrDimension
	}
	return manager.runBooleanOperation(nil, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		if count > work.options.MaxBytes/4 {
			return 0, ErrLimit
		}
		roots := make([]ExtraBDDLit, count)
		for variable := range roots {
			roots[variable] = ExtraBDDLit(variable+1) << 1
		}
		if err := extraBDDValidateGraph(manager, roots, work); err != nil {
			return 0, err
		}
		result := ExtraBDDFalse
		memo := make(map[extraBDDIteKey]ExtraBDDLit)
		for variable := count - 1; variable >= 0; variable-- {
			literal := ExtraBDDLit(variable+1) << 1
			var err error
			result, err = manager.iteRecursive(literal, result.Not(), result, work, memo)
			if err != nil {
				return 0, err
			}
		}
		return result, nil
	})
}

func ExtraBDDTransferPermute(source *ExtraBDDManager, destination *ExtraBDDManager, root ExtraBDDLit, permutation []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if source == nil || destination == nil || source.variables != destination.variables || len(permutation) != source.variables {
		return 0, ErrDimension
	}
	if err := source.checkLiteral(root); err != nil {
		return 0, err
	}
	if err := destination.check(); err != nil {
		return 0, err
	}
	seen := make([]bool, len(permutation))
	for _, variable := range permutation {
		if variable < 0 || variable >= len(permutation) || seen[variable] {
			return 0, ErrDimension
		}
		seen[variable] = true
	}
	count, normalized, err := source.analysisCount(options, 2)
	if err != nil {
		return 0, err
	}
	work := extraBDDWork{options: normalized}
	if err := extraBDDValidateGraph(source, []ExtraBDDLit{root}, &work); err != nil {
		return 0, err
	}
	values, err := source.analysisDense(root, count, &work)
	if err != nil {
		return 0, err
	}
	permuted := make([]bool, count)
	for assignment := 0; assignment < count; assignment++ {
		if err := work.step(); err != nil {
			return 0, err
		}
		sourceAssignment := 0
		for old, next := range permutation {
			if assignment&(1<<uint(next)) != 0 {
				sourceAssignment |= 1 << uint(old)
			}
		}
		permuted[assignment] = values[sourceAssignment]
	}
	transaction := destination.beginTransaction()
	work.transaction = &transaction
	result, err := destination.analysisBuild(permuted, &work)
	if err != nil {
		destination.rollback(&transaction)
		return 0, err
	}
	return result, nil
}

func (manager *ExtraBDDManager) ExtraImage(care ExtraBDDLit, parts []ExtraBDDLit, quantify []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	if len(parts) > normalized.MaxBytes/8 || len(quantify) > normalized.MaxBytes/8 {
		return 0, ErrLimit
	}
	roots := make([]ExtraBDDLit, 1, len(parts)+1)
	roots[0] = care
	roots = append(roots, parts...)
	seen := make([]bool, manager.variables)
	mask := 0
	for _, variable := range quantify {
		if variable < 0 || variable >= manager.variables || seen[variable] || variable >= strconv.IntSize-1 {
			return 0, ErrDimension
		}
		seen[variable] = true
		mask |= 1 << uint(variable)
	}
	return manager.analysisRun(roots, len(roots)+2, normalized, func(count int, work *extraBDDWork) (ExtraBDDLit, error) {
		if err := extraBDDValidateGraph(manager, roots, work); err != nil {
			return 0, err
		}
		tables := make([][]bool, len(roots))
		for index, root := range roots {
			var err error
			tables[index], err = manager.analysisDense(root, count, work)
			if err != nil {
				return 0, err
			}
		}
		projected := make([]bool, count)
		for assignment := 0; assignment < count; assignment++ {
			if err := work.step(); err != nil {
				return 0, err
			}
			value := true
			for _, table := range tables {
				value = value && table[assignment]
			}
			if value {
				projected[assignment&^mask] = true
			}
		}
		values := make([]bool, count)
		for assignment := range values {
			if err := work.step(); err != nil {
				return 0, err
			}
			values[assignment] = projected[assignment&^mask]
		}
		return manager.analysisBuild(values, work)
	})
}

type ExtraBDDUnateVariable struct {
	Variable           int
	Positive, Negative bool
}

func (manager *ExtraBDDManager) ExtraUnateness(root ExtraBDDLit, options ExtraBDDOptions) ([]ExtraBDDUnateVariable, error) {
	count, normalized, err := manager.analysisCount(options, 1)
	if err != nil {
		return nil, err
	}
	work := extraBDDWork{options: normalized}
	if err := extraBDDValidateGraph(manager, []ExtraBDDLit{root}, &work); err != nil {
		return nil, err
	}
	values, err := manager.analysisDense(root, count, &work)
	if err != nil {
		return nil, err
	}
	result := make([]ExtraBDDUnateVariable, 0, manager.variables)
	for variable := 0; variable < manager.variables; variable++ {
		positive, negative := true, true
		dependent := false
		bit := 1 << uint(variable)
		for assignment := 0; assignment < count; assignment++ {
			if assignment&bit != 0 {
				continue
			}
			if err := work.step(); err != nil {
				return nil, err
			}
			low, high := values[assignment], values[assignment|bit]
			dependent = dependent || low != high
			positive = positive && (!low || high)
			negative = negative && (!high || low)
		}
		if dependent {
			result = append(result, ExtraBDDUnateVariable{Variable: variable, Positive: positive, Negative: negative})
		}
	}
	return result, nil
}

type ExtraBDDSymmetryPair struct{ First, Second int }

func (manager *ExtraBDDManager) ExtraSymmetry(root ExtraBDDLit, options ExtraBDDOptions) ([]ExtraBDDSymmetryPair, error) {
	count, normalized, err := manager.analysisCount(options, 1)
	if err != nil {
		return nil, err
	}
	work := extraBDDWork{options: normalized}
	if err := extraBDDValidateGraph(manager, []ExtraBDDLit{root}, &work); err != nil {
		return nil, err
	}
	values, err := manager.analysisDense(root, count, &work)
	if err != nil {
		return nil, err
	}
	support := make([]int, 0, manager.variables)
	for variable := 0; variable < manager.variables; variable++ {
		bit := 1 << uint(variable)
		dependent := false
		for assignment := 0; assignment < count; assignment++ {
			if err := work.step(); err != nil {
				return nil, err
			}
			if assignment&bit == 0 && values[assignment] != values[assignment|bit] {
				dependent = true
				break
			}
		}
		if dependent {
			support = append(support, variable)
		}
	}
	result := make([]ExtraBDDSymmetryPair, 0)
	for left := 0; left < len(support); left++ {
		for right := left + 1; right < len(support); right++ {
			first, second := support[left], support[right]
			symmetric := true
			for assignment := 0; assignment < count; assignment++ {
				if err := work.step(); err != nil {
					return nil, err
				}
				firstBit := assignment >> uint(first) & 1
				secondBit := assignment >> uint(second) & 1
				swapped := assignment
				if firstBit != secondBit {
					swapped ^= 1<<uint(first) | 1<<uint(second)
				}
				if values[assignment] != values[swapped] {
					symmetric = false
					break
				}
			}
			if symmetric {
				result = append(result, ExtraBDDSymmetryPair{First: first, Second: second})
			}
		}
	}
	return result, nil
}

func ExtraBDDThresholdChow(values []bool, variables int, options ExtraBDDOptions) (int, []int, error) {
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, nil, err
	}
	if variables < 1 || variables >= strconv.IntSize-1 || len(values) != 1<<uint(variables) {
		return 0, nil, ErrDimension
	}
	if variables > options.MaxBytes/8 {
		return 0, nil, ErrLimit
	}
	work := extraBDDWork{options: options}
	coefficients := make([]int, variables)
	ones := 0
	for assignment, value := range values {
		if err := work.step(); err != nil {
			return 0, nil, err
		}
		if !value {
			continue
		}
		ones++
		for variable := 0; variable < variables; variable++ {
			if assignment&(1<<uint(variable)) != 0 {
				coefficients[variable]++
			}
		}
	}
	for variable := range coefficients {
		coefficients[variable] = 2*coefficients[variable] - ones
	}
	return ones - (1 << uint(variables-1)), coefficients, nil
}

func ExtraBDDThresholdCheck(values []bool, weights []int, threshold int, options ExtraBDDOptions) (bool, error) {
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return false, err
	}
	if len(weights) >= strconv.IntSize-1 || len(values) != 1<<uint(len(weights)) {
		return false, ErrDimension
	}
	work := extraBDDWork{options: options}
	for assignment, value := range values {
		if err := work.step(); err != nil {
			return false, err
		}
		sum := int64(0)
		for variable, weight := range weights {
			if assignment&(1<<uint(variable)) != 0 {
				value := int64(weight)
				if value > 0 && sum > math.MaxInt64-value || value < 0 && sum < math.MinInt64-value {
					return false, ErrLimit
				}
				sum += value
			}
		}
		if (sum >= int64(threshold)) != value {
			return false, nil
		}
	}
	return true, nil
}

func ExtraBDDGrayCode(value uint64) uint64 { return value ^ value>>1 }

func ExtraBDDBinaryCode(gray uint64) uint64 {
	result := gray
	for shift := uint(1); shift < 64; shift <<= 1 {
		result ^= result >> shift
	}
	return result
}

func (manager *ExtraBDDManager) ExtraKMap(root ExtraBDDLit, options ExtraBDDOptions) (string, error) {
	count, normalized, err := manager.analysisCount(options, 1)
	if err != nil {
		return "", err
	}
	if manager.variables > 8 || count > normalized.MaxBytes/2 {
		return "", ErrLimit
	}
	work := extraBDDWork{options: normalized}
	if err := extraBDDValidateGraph(manager, []ExtraBDDLit{root}, &work); err != nil {
		return "", err
	}
	values, err := manager.analysisDense(root, count, &work)
	if err != nil {
		return "", err
	}
	columns := 1 << uint((manager.variables+1)/2)
	rows := 1 << uint(manager.variables/2)
	var builder strings.Builder
	for row := 0; row < rows; row++ {
		for column := 0; column < columns; column++ {
			if err := work.step(); err != nil {
				return "", err
			}
			assignment := int(ExtraBDDGrayCode(uint64(column))) | int(ExtraBDDGrayCode(uint64(row)))<<uint((manager.variables+1)/2)
			if values[assignment] {
				builder.WriteByte('1')
			} else {
				builder.WriteByte('0')
			}
			if column+1 < columns {
				builder.WriteByte(' ')
			}
		}
		builder.WriteByte('\n')
	}
	if builder.Len() > normalized.MaxBytes {
		return "", ErrLimit
	}
	return builder.String(), nil
}

type extraZDDRelation uint8

const (
	extraZDDSubsets extraZDDRelation = iota
	extraZDDSupersets
	extraZDDNotSubsets
	extraZDDNotSupersets
)

func extraZDDValidateGraphs(manager *CuddZDDManager, roots []CuddZDDEdge, work *cuddZDDWork) error {
	if len(roots) == 0 {
		return nil
	}
	if len(manager.nodes) > work.options.MaxBytes-work.bytes {
		return ErrLimit
	}
	work.bytes += len(manager.nodes)
	state := make([]uint8, len(manager.nodes))
	var visit func(CuddZDDEdge, int) error
	visit = func(edge CuddZDDEdge, parent int) error {
		if err := work.step(); err != nil {
			return err
		}
		if !manager.valid(edge) {
			return ErrInvalidLiteral
		}
		if edge <= CuddZDDBase {
			return nil
		}
		node := manager.nodes[edge]
		variable := int(node.variable)
		if variable >= manager.variables || variable <= parent || node.high == CuddZDDEmpty {
			return ErrInvariant
		}
		key := cuddZDDKey{variable: node.variable, high: node.high, low: node.low}
		if canonical, found := manager.unique[key]; !found || canonical != edge {
			return ErrInvariant
		}
		if state[edge] == 1 {
			return ErrInvariant
		}
		if state[edge] == 2 {
			return nil
		}
		state[edge] = 1
		if err := visit(node.high, variable); err != nil {
			return err
		}
		if err := visit(node.low, variable); err != nil {
			return err
		}
		state[edge] = 2
		return nil
	}
	for _, root := range roots {
		if err := visit(root, -1); err != nil {
			return err
		}
	}
	return nil
}

func extraZDDRun(manager *CuddZDDManager, roots []CuddZDDEdge, options CuddZDDOptions, operation func([]map[uint64]struct{}, *cuddZDDWork) (map[uint64]struct{}, error)) (CuddZDDEdge, error) {
	if manager == nil {
		return 0, ErrDimension
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return 0, err
	}
	for _, root := range roots {
		if !manager.valid(root) {
			return 0, ErrInvalidLiteral
		}
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return 0, err
	}
	work := &cuddZDDWork{options: normalized}
	if err := extraZDDValidateGraphs(manager, roots, work); err != nil {
		return 0, err
	}
	families := make([]map[uint64]struct{}, len(roots))
	memo := make(map[CuddZDDEdge]map[uint64]struct{})
	for index, root := range roots {
		families[index], err = manager.decode(root, work, memo)
		if err != nil {
			return 0, err
		}
	}
	resultFamily, err := operation(families, work)
	if err != nil {
		return 0, err
	}
	oldCount := len(manager.nodes)
	result, err := manager.build(resultFamily, 0, work)
	if err != nil {
		manager.rollback(oldCount)
		return 0, err
	}
	manager.operations += uint64(work.steps)
	return result, nil
}

func extraSetSubset(first, second uint64) bool { return first&^second == 0 }

func extraZDDFilter(manager *CuddZDDManager, first, second CuddZDDEdge, relation extraZDDRelation, maximal bool, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDRun(manager, []CuddZDDEdge{first, second}, options, func(families []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		result := make(map[uint64]struct{})
		for candidate := range families[0] {
			matched := false
			for reference := range families[1] {
				if err := work.step(); err != nil {
					return nil, err
				}
				subset := extraSetSubset(candidate, reference)
				superset := extraSetSubset(reference, candidate)
				switch relation {
				case extraZDDSubsets:
					matched = matched || subset
				case extraZDDSupersets:
					matched = matched || superset
				case extraZDDNotSubsets:
					matched = matched || subset
				case extraZDDNotSupersets:
					matched = matched || superset
				}
			}
			keep := matched
			if relation == extraZDDNotSubsets || relation == extraZDDNotSupersets {
				keep = !matched
			}
			if keep {
				if err := work.addSet(result, candidate); err != nil {
					return nil, err
				}
			}
		}
		if maximal {
			return extraZDDReduce(result, true, work)
		}
		return result, nil
	})
}

func extraZDDReduce(family map[uint64]struct{}, maximal bool, work *cuddZDDWork) (map[uint64]struct{}, error) {
	result := make(map[uint64]struct{})
	for candidate := range family {
		keep := true
		for other := range family {
			if err := work.step(); err != nil {
				return nil, err
			}
			if candidate == other {
				continue
			}
			if maximal && extraSetSubset(candidate, other) || !maximal && extraSetSubset(other, candidate) {
				keep = false
			}
		}
		if keep {
			if err := work.addSet(result, candidate); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func (manager *CuddZDDManager) ExtraMaximal(root CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDRun(manager, []CuddZDDEdge{root}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		return extraZDDReduce(f[0], true, work)
	})
}

func (manager *CuddZDDManager) ExtraMinimal(root CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDRun(manager, []CuddZDDEdge{root}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		return extraZDDReduce(f[0], false, work)
	})
}

func (manager *CuddZDDManager) extraReducedUnion(first, second CuddZDDEdge, maximal bool, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDRun(manager, []CuddZDDEdge{first, second}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		combined := make(map[uint64]struct{})
		for _, family := range f {
			for set := range family {
				if err := work.addSet(combined, set); err != nil {
					return nil, err
				}
			}
		}
		return extraZDDReduce(combined, maximal, work)
	})
}

func (manager *CuddZDDManager) ExtraMaxUnion(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.extraReducedUnion(first, second, true, options)
}

func (manager *CuddZDDManager) ExtraMinUnion(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.extraReducedUnion(first, second, false, options)
}

func (manager *CuddZDDManager) ExtraSubsets(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDFilter(manager, first, second, extraZDDSubsets, false, options)
}

func (manager *CuddZDDManager) ExtraSupersets(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDFilter(manager, first, second, extraZDDSupersets, false, options)
}

func (manager *CuddZDDManager) ExtraNotSubsets(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDFilter(manager, first, second, extraZDDNotSubsets, false, options)
}

func (manager *CuddZDDManager) ExtraNotSupersets(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDFilter(manager, first, second, extraZDDNotSupersets, false, options)
}

func (manager *CuddZDDManager) ExtraMaxNotSupersets(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDFilter(manager, first, second, extraZDDNotSupersets, true, options)
}

func (manager *CuddZDDManager) extraFamilyProduct(first, second CuddZDDEdge, union, maximal bool, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDRun(manager, []CuddZDDEdge{first, second}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		result := make(map[uint64]struct{})
		for left := range f[0] {
			for right := range f[1] {
				value := left & right
				if union {
					value = left | right
				}
				if err := work.addSet(result, value); err != nil {
					return nil, err
				}
			}
		}
		if maximal {
			return extraZDDReduce(result, true, work)
		}
		return result, nil
	})
}

func (manager *CuddZDDManager) ExtraDotProduct(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.extraFamilyProduct(first, second, true, false, options)
}

func (manager *CuddZDDManager) ExtraCrossProduct(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.extraFamilyProduct(first, second, false, false, options)
}

func (manager *CuddZDDManager) ExtraMaxDotProduct(first, second CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return manager.extraFamilyProduct(first, second, true, true, options)
}

func (manager *CuddZDDManager) ExtraTuples(variables []int, choose int, options CuddZDDOptions) (CuddZDDEdge, error) {
	if manager == nil || choose < 0 || choose > len(variables) {
		return 0, ErrDimension
	}
	seen := make([]bool, manager.variables)
	for _, variable := range variables {
		if variable < 0 || variable >= manager.variables || seen[variable] {
			return 0, ErrDimension
		}
		seen[variable] = true
	}
	return extraZDDRun(manager, nil, options, func(_ []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		result := make(map[uint64]struct{})
		var visit func(int, int, uint64) error
		visit = func(index, remaining int, set uint64) error {
			if err := work.step(); err != nil {
				return err
			}
			if remaining == 0 {
				return work.addSet(result, set)
			}
			if len(variables)-index < remaining {
				return nil
			}
			if err := visit(index+1, remaining-1, set|uint64(1)<<uint(variables[index])); err != nil {
				return err
			}
			return visit(index+1, remaining, set)
		}
		if err := visit(0, choose, 0); err != nil {
			return nil, err
		}
		return result, nil
	})
}

func (manager *CuddZDDManager) ExtraSelectOne(root CuddZDDEdge, options CuddZDDOptions) (CuddZDDEdge, error) {
	return extraZDDRun(manager, []CuddZDDEdge{root}, options, func(f []map[uint64]struct{}, work *cuddZDDWork) (map[uint64]struct{}, error) {
		if len(f[0]) == 0 {
			return map[uint64]struct{}{}, nil
		}
		selected := uint64(math.MaxUint64)
		for set := range f[0] {
			if err := work.step(); err != nil {
				return nil, err
			}
			if set < selected {
				selected = set
			}
		}
		result := make(map[uint64]struct{}, 1)
		if err := work.addSet(result, selected); err != nil {
			return nil, err
		}
		return result, nil
	})
}

func (manager *CuddZDDManager) ExtraEmptyBelongs(root CuddZDDEdge, options CuddZDDOptions) (bool, error) {
	if manager == nil {
		return false, ErrDimension
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return false, err
	}
	if !manager.valid(root) {
		return false, ErrInvalidLiteral
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return false, err
	}
	work := &cuddZDDWork{options: normalized}
	if err := extraZDDValidateGraphs(manager, []CuddZDDEdge{root}, work); err != nil {
		return false, err
	}
	current := root
	for current > CuddZDDBase {
		if err := work.step(); err != nil {
			return false, err
		}
		current = manager.nodes[current].low
	}
	manager.operations += uint64(work.steps)
	return current == CuddZDDBase, nil
}

func (manager *CuddZDDManager) ExtraIsOneSubset(root CuddZDDEdge, options CuddZDDOptions) (bool, error) {
	if manager == nil {
		return false, ErrDimension
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if err := manager.check(); err != nil {
		return false, err
	}
	if !manager.valid(root) {
		return false, ErrInvalidLiteral
	}
	normalized, err := normalizeCuddZDDOptions(options)
	if err != nil {
		return false, err
	}
	work := &cuddZDDWork{options: normalized}
	if err := extraZDDValidateGraphs(manager, []CuddZDDEdge{root}, work); err != nil {
		return false, err
	}
	current := root
	for current > CuddZDDBase {
		if err := work.step(); err != nil {
			return false, err
		}
		node := manager.nodes[current]
		if node.low != CuddZDDEmpty {
			manager.operations += uint64(work.steps)
			return false, nil
		}
		current = node.high
	}
	manager.operations += uint64(work.steps)
	return current == CuddZDDBase, nil
}
