package abc

// Bounded Boolean approximations and priority operations from
// src/bdd/cudd/cuddApprox.c, cuddClip.c, cuddSubsetHB.c,
// cuddSubsetSP.c, cuddPriority.c, and cuddSign.c.
// ExtraBDDManager has a fixed order, so CUDD's reordering restarts do not
// apply. ADD-valued routines are outside this Boolean manager. The quality
// and safety settings are validated source-shaped configuration; deterministic
// weight/path ties define the adapted heuristics rather than reproducing the
// original floating-point node-ranking decisions.

import (
	"math"
	"slices"
)

type CuddApproxOptions struct {
	ExtraBDDOptions
	Threshold int
	Quality   float64
	Safe      bool
}

type CuddClipDirection uint8

const (
	CuddClipUnder CuddClipDirection = iota
	CuddClipOver
)

type CuddPriorityOrder uint8

const (
	CuddPrioritySmallest CuddPriorityOrder = iota
	CuddPriorityLargest
)

func normalizeCuddApproxOptions(options CuddApproxOptions) (CuddApproxOptions, error) {
	if options.Threshold < 0 || math.IsNaN(options.Quality) || math.IsInf(options.Quality, 0) || options.Quality < 0 {
		return options, ErrDimension
	}
	normalized, err := normalizeExtraBDDOptions(options.ExtraBDDOptions)
	options.ExtraBDDOptions = normalized
	return options, err
}

func (manager *ExtraBDDManager) cuddApproxRun(roots []ExtraBDDLit, options CuddApproxOptions, operation func(*extraBDDWork) (ExtraBDDLit, error)) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	options, err := normalizeCuddApproxOptions(options)
	if err != nil {
		return 0, err
	}
	for _, root := range roots {
		if !manager.valid(root) {
			return 0, ErrInvalidLiteral
		}
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: options.ExtraBDDOptions, transaction: &transaction}
	result, err := operation(&work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	if options.Threshold > 0 {
		nodes, countErr := manager.cuddApproxNodeCount(result, &work)
		if countErr != nil || nodes > options.Threshold {
			manager.rollback(&transaction)
			if countErr != nil {
				return 0, countErr
			}
			return 0, ErrLimit
		}
	}
	return result, nil
}

func (manager *ExtraBDDManager) cuddApproxNodeCount(root ExtraBDDLit, work *extraBDDWork) (int, error) {
	seen := make(map[uint32]struct{})
	stack := []ExtraBDDLit{root}
	for len(stack) != 0 {
		if err := work.step(); err != nil {
			return 0, err
		}
		literal := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if literal < 2 {
			continue
		}
		id := literal.ID()
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		stack = append(stack, manager.elseOf(literal), manager.thenOf(literal))
	}
	return len(seen), nil
}

func (manager *ExtraBDDManager) cuddSatWeight(root ExtraBDDLit, work *extraBDDWork, memo map[ExtraBDDLit]uint64) (uint64, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if root == ExtraBDDFalse {
		return 0, nil
	}
	if root == ExtraBDDTrue {
		return 1, nil
	}
	if value, ok := memo[root]; ok {
		return value, nil
	}
	low, err := manager.cuddSatWeight(manager.elseOf(root), work, memo)
	if err != nil {
		return 0, err
	}
	high, err := manager.cuddSatWeight(manager.thenOf(root), work, memo)
	if err != nil {
		return 0, err
	}
	value := low + high
	if value < low {
		value = math.MaxUint64
	}
	memo[root] = value
	return value, nil
}

func (manager *ExtraBDDManager) cuddSubsetBranch(root ExtraBDDLit, budget int, shortest bool, work *extraBDDWork, weights map[ExtraBDDLit]uint64, lengths map[ExtraBDDLit]int) (ExtraBDDLit, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if root < 2 || budget <= 0 {
		if root == ExtraBDDTrue && budget >= 0 {
			return root, nil
		}
		return ExtraBDDFalse, nil
	}
	low, high := manager.elseOf(root), manager.thenOf(root)
	chooseHigh := true
	if shortest {
		lowLength := manager.cuddShortestToOne(low, lengths)
		highLength := manager.cuddShortestToOne(high, lengths)
		chooseHigh = highLength <= lowLength
	} else {
		lowWeight, err := manager.cuddSatWeight(low, work, weights)
		if err != nil {
			return 0, err
		}
		highWeight, err := manager.cuddSatWeight(high, work, weights)
		if err != nil {
			return 0, err
		}
		chooseHigh = highWeight >= lowWeight
	}
	chosen := low
	if chooseHigh {
		chosen = high
	}
	child, err := manager.cuddSubsetBranch(chosen, budget-1, shortest, work, weights, lengths)
	if err != nil || child == ExtraBDDFalse {
		return child, err
	}
	if chooseHigh {
		return manager.uniqueCreate(manager.variableOf(root), child, ExtraBDDFalse, work.transaction)
	}
	return manager.uniqueCreate(manager.variableOf(root), ExtraBDDFalse, child, work.transaction)
}

func (manager *ExtraBDDManager) cuddShortestToOne(root ExtraBDDLit, memo map[ExtraBDDLit]int) int {
	if root == ExtraBDDTrue {
		return 0
	}
	if root == ExtraBDDFalse {
		return math.MaxInt / 4
	}
	if value, ok := memo[root]; ok {
		return value
	}
	value := 1 + min(manager.cuddShortestToOne(manager.elseOf(root), memo), manager.cuddShortestToOne(manager.thenOf(root), memo))
	memo[root] = value
	return value
}

func (manager *ExtraBDDManager) cuddSubset(root ExtraBDDLit, shortest bool, options CuddApproxOptions) (ExtraBDDLit, error) {
	return manager.cuddApproxRun([]ExtraBDDLit{root}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		if options.Threshold == 0 {
			return ExtraBDDFalse, nil
		}
		nodes, err := manager.cuddApproxNodeCount(root, work)
		if err != nil || nodes <= options.Threshold {
			return root, err
		}
		return manager.cuddSubsetBranch(root, options.Threshold, shortest, work, make(map[ExtraBDDLit]uint64), make(map[ExtraBDDLit]int))
	})
}

func (manager *ExtraBDDManager) UnderApprox(root ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	return manager.cuddSubset(root, false, options)
}

func (manager *ExtraBDDManager) OverApprox(root ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	result, err := manager.cuddSubset(root.Not(), false, options)
	if err != nil {
		return 0, err
	}
	return result.Not(), err
}

func (manager *ExtraBDDManager) RemapUnderApprox(root ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	return manager.cuddSubset(root, false, options)
}

func (manager *ExtraBDDManager) RemapOverApprox(root ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	return manager.OverApprox(root, options)
}

func (manager *ExtraBDDManager) BiasedUnderApprox(root, bias ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	return manager.cuddApproxRun([]ExtraBDDLit{root, bias}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		if options.Threshold > 0 {
			nodes, countErr := manager.cuddApproxNodeCount(root, work)
			if countErr != nil {
				return 0, countErr
			}
			if nodes <= options.Threshold {
				return root, nil
			}
		}
		mandatory, err := manager.andRecursive(root, bias, work)
		if err != nil {
			return 0, err
		}
		if options.Threshold > 0 {
			nodes, countErr := manager.cuddApproxNodeCount(mandatory, work)
			if countErr != nil || nodes > options.Threshold {
				if countErr != nil {
					return 0, countErr
				}
				return 0, ErrLimit
			}
		}
		return mandatory, nil
	})
}

func (manager *ExtraBDDManager) BiasedOverApprox(root, bias ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	result, err := manager.BiasedUnderApprox(root.Not(), bias, options)
	if err != nil {
		return 0, err
	}
	return result.Not(), err
}

func (manager *ExtraBDDManager) SubsetHeavyBranch(root ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	return manager.cuddSubset(root, false, options)
}

func (manager *ExtraBDDManager) SupersetHeavyBranch(root ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	return manager.OverApprox(root, options)
}

func (manager *ExtraBDDManager) SubsetShortPaths(root ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	return manager.cuddSubset(root, true, options)
}

func (manager *ExtraBDDManager) SupersetShortPaths(root ExtraBDDLit, options CuddApproxOptions) (ExtraBDDLit, error) {
	result, err := manager.cuddSubset(root.Not(), true, options)
	if err != nil {
		return 0, err
	}
	return result.Not(), err
}

func (manager *ExtraBDDManager) cuddClipAnd(first, second ExtraBDDLit, depth int, direction CuddClipDirection, work *extraBDDWork, memo map[[3]ExtraBDDLit]ExtraBDDLit) (ExtraBDDLit, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if first == ExtraBDDFalse || second == ExtraBDDFalse || first == second.Not() {
		return ExtraBDDFalse, nil
	}
	if first == second || second == ExtraBDDTrue {
		return first, nil
	}
	if first == ExtraBDDTrue {
		return second, nil
	}
	if depth == 0 {
		leq, err := manager.leqRecursive(first, second, work, make(map[extraBDDPairKey]bool))
		if err != nil {
			return 0, err
		}
		if leq {
			return first, nil
		}
		leq, err = manager.leqRecursive(second, first, work, make(map[extraBDDPairKey]bool))
		if err != nil {
			return 0, err
		}
		if leq {
			return second, nil
		}
		if direction == CuddClipUnder {
			return ExtraBDDFalse, nil
		}
		disjoint, err := manager.leqRecursive(first, second.Not(), work, make(map[extraBDDPairKey]bool))
		if err != nil {
			return 0, err
		}
		if !disjoint {
			disjoint, err = manager.leqRecursive(second, first.Not(), work, make(map[extraBDDPairKey]bool))
			if err != nil {
				return 0, err
			}
		}
		if disjoint {
			return ExtraBDDFalse, nil
		}
		return ExtraBDDTrue, nil
	}
	if first > second {
		first, second = second, first
	}
	key := [3]ExtraBDDLit{first, second, ExtraBDDLit(depth)<<1 | ExtraBDDLit(direction)}
	if result, ok := memo[key]; ok {
		return result, nil
	}
	variable := min(manager.variableOf(first), manager.variableOf(second))
	firstLow, firstHigh := first, first
	secondLow, secondHigh := second, second
	if manager.variableOf(first) == variable {
		firstLow, firstHigh = manager.elseOf(first), manager.thenOf(first)
	}
	if manager.variableOf(second) == variable {
		secondLow, secondHigh = manager.elseOf(second), manager.thenOf(second)
	}
	low, err := manager.cuddClipAnd(firstLow, secondLow, depth-1, direction, work, memo)
	if err != nil {
		return 0, err
	}
	high, err := manager.cuddClipAnd(firstHigh, secondHigh, depth-1, direction, work, memo)
	if err != nil {
		return 0, err
	}
	result, err := manager.uniqueCreate(variable, high, low, work.transaction)
	if err == nil {
		memo[key] = result
	}
	return result, err
}

func (manager *ExtraBDDManager) ClippingAnd(first, second ExtraBDDLit, depth int, direction CuddClipDirection, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if depth < 0 || direction > CuddClipOver {
		return 0, ErrDimension
	}
	return manager.runBooleanOperation([]ExtraBDDLit{first, second}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		return manager.cuddClipAnd(first, second, depth, direction, work, make(map[[3]ExtraBDDLit]ExtraBDDLit))
	})
}

func (manager *ExtraBDDManager) ClippingAndAbstract(first, second ExtraBDDLit, variables []int, depth int, direction CuddClipDirection, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if depth < 0 || direction > CuddClipOver {
		return 0, ErrDimension
	}
	set, err := manager.abstractionSet(variables)
	if err != nil {
		return 0, err
	}
	return manager.runBooleanOperation([]ExtraBDDLit{first, second}, options, func(work *extraBDDWork) (ExtraBDDLit, error) {
		clipped, clipErr := manager.cuddClipAnd(first, second, depth, direction, work, make(map[[3]ExtraBDDLit]ExtraBDDLit))
		if clipErr != nil {
			return 0, clipErr
		}
		return manager.abstractRecursive(clipped, set, false, false, work, make(map[extraBDDIteKey]ExtraBDDLit), make(map[ExtraBDDLit]ExtraBDDLit))
	})
}

func (manager *ExtraBDDManager) cuddCheckVariables(groups ...[]int) error {
	if err := manager.check(); err != nil {
		return err
	}
	seen := make([]bool, manager.variables)
	for _, group := range groups {
		for _, variable := range group {
			if variable < 0 || variable >= manager.variables || seen[variable] {
				return ErrDimension
			}
			seen[variable] = true
		}
	}
	return nil
}

func (manager *ExtraBDDManager) cuddPredicate(variables []int, options ExtraBDDOptions, predicate func(uint64) bool) (ExtraBDDLit, error) {
	if len(variables) > 63 || manager.cuddCheckVariables(variables) != nil {
		return 0, ErrDimension
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	if len(variables) >= 63 || (uint64(1)<<uint(len(variables)+1))-1 > uint64(normalized.MaxOperations) {
		return 0, ErrLimit
	}
	ordered := append([]int(nil), variables...)
	slices.Sort(ordered)
	return manager.runBooleanOperation(nil, normalized, func(work *extraBDDWork) (ExtraBDDLit, error) {
		var build func(int, uint64) (ExtraBDDLit, error)
		build = func(position int, value uint64) (ExtraBDDLit, error) {
			if err := work.step(); err != nil {
				return 0, err
			}
			if position == len(ordered) {
				var original uint64
				for originalPosition, variable := range variables {
					orderedPosition, _ := slices.BinarySearch(ordered, variable)
					original |= (value >> uint(orderedPosition) & 1) << uint(originalPosition)
				}
				if predicate(original) {
					return ExtraBDDTrue, nil
				}
				return ExtraBDDFalse, nil
			}
			low, err := build(position+1, value)
			if err != nil {
				return 0, err
			}
			high, err := build(position+1, value|(uint64(1)<<uint(position)))
			if err != nil {
				return 0, err
			}
			return manager.uniqueCreate(ordered[position], high, low, work.transaction)
		}
		return build(0, 0)
	})
}

func (manager *ExtraBDDManager) CuddXGreaterY(x, y []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if len(x) != len(y) || len(x) > 31 || manager.cuddCheckVariables(x, y) != nil {
		return 0, ErrDimension
	}
	variables := append(append([]int(nil), x...), y...)
	return manager.cuddPredicate(variables, options, func(value uint64) bool {
		var xv, yv uint64
		for i := range x {
			xv = xv<<1 | (value >> uint(i) & 1)
			yv = yv<<1 | (value >> uint(len(x)+i) & 1)
		}
		return xv > yv
	})
}

func (manager *ExtraBDDManager) CuddXEqualY(x, y []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if len(x) != len(y) || len(x) > 31 || manager.cuddCheckVariables(x, y) != nil {
		return 0, ErrDimension
	}
	variables := append(append([]int(nil), x...), y...)
	return manager.cuddPredicate(variables, options, func(value uint64) bool {
		for i := range x {
			if value>>uint(i)&1 != value>>uint(len(x)+i)&1 {
				return false
			}
		}
		return true
	})
}

func (manager *ExtraBDDManager) CuddDistanceXYGreaterXZ(x, y, z []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.cuddDistancePriority(x, y, z, false, options)
}

func (manager *ExtraBDDManager) CuddDistanceXYGreaterYZ(x, y, z []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.cuddDistancePriority(x, y, z, true, options)
}

func (manager *ExtraBDDManager) cuddDistancePriority(x, y, z []int, compareYZ bool, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if len(x) != len(y) || len(x) != len(z) || len(x) > 21 || manager.cuddCheckVariables(x, y, z) != nil {
		return 0, ErrDimension
	}
	variables := append(append(append([]int(nil), x...), y...), z...)
	return manager.cuddPredicate(variables, options, func(value uint64) bool {
		dxy, other := 0, 0
		for bit := range x {
			xb := value >> uint(bit) & 1
			yb := value >> uint(len(x)+bit) & 1
			zb := value >> uint(2*len(x)+bit) & 1
			if xb != yb {
				dxy++
			}
			if compareYZ {
				if yb != zb {
					other++
				}
			} else if xb != zb {
				other++
			}
		}
		return dxy > other
	})
}

func (manager *ExtraBDDManager) CuddDifferenceAtLeast(x, y []int, constant int64, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.cuddDifference(x, y, constant, false, options)
}

func (manager *ExtraBDDManager) CuddDifferenceNotEqual(x, y []int, constant int64, options ExtraBDDOptions) (ExtraBDDLit, error) {
	return manager.cuddDifference(x, y, constant, true, options)
}

func (manager *ExtraBDDManager) cuddDifference(x, y []int, constant int64, unequal bool, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if len(x) != len(y) || len(x) > 31 || manager.cuddCheckVariables(x, y) != nil {
		return 0, ErrDimension
	}
	variables := append(append([]int(nil), x...), y...)
	return manager.cuddPredicate(variables, options, func(value uint64) bool {
		var xv, yv int64
		for i := range x {
			xv = xv<<1 | int64(value>>uint(i)&1)
			yv = yv<<1 | int64(value>>uint(len(x)+i)&1)
		}
		if unequal {
			return xv-yv != constant
		}
		return xv-yv >= constant
	})
}

func (manager *ExtraBDDManager) CuddInterval(x []int, lower, upper uint64, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if len(x) > 63 || manager.cuddCheckVariables(x) != nil || lower > upper {
		return 0, ErrDimension
	}
	return manager.cuddPredicate(x, options, func(value uint64) bool {
		var number uint64
		for i := range x {
			number = number<<1 | (value >> uint(i) & 1)
		}
		return number >= lower && number <= upper
	})
}

func (manager *ExtraBDDManager) CofactorMintermSignatures(root ExtraBDDLit, options ExtraBDDOptions) ([]float64, error) {
	count, normalized, err := manager.analysisCount(options, 1)
	if err != nil || !manager.valid(root) {
		if err == nil {
			err = ErrInvalidLiteral
		}
		return nil, err
	}
	work := extraBDDWork{options: normalized}
	values, err := manager.analysisDense(root, count, &work)
	if err != nil {
		return nil, err
	}
	result := make([]float64, manager.variables+1)
	trueCount := 0
	for assignment, value := range values {
		if err = work.step(); err != nil {
			return nil, err
		}
		if !value {
			continue
		}
		trueCount++
		for variable := 0; variable < manager.variables; variable++ {
			if assignment&(1<<uint(variable)) != 0 {
				result[variable]++
			}
		}
	}
	if manager.variables > 0 {
		denominator := float64(uint64(1) << uint(manager.variables-1))
		for variable := 0; variable < manager.variables; variable++ {
			result[variable] /= denominator
		}
	}
	result[manager.variables] = float64(trueCount) / float64(count)
	return result, nil
}

func (manager *ExtraBDDManager) PrioritySelect(relation ExtraBDDLit, x, y []int, order CuddPriorityOrder, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if len(x) != len(y) || len(x) > 30 || order > CuddPriorityLargest || manager.cuddCheckVariables(x, y) != nil {
		return 0, ErrDimension
	}
	count, normalized, err := manager.analysisCount(options, 4)
	if err != nil || !manager.valid(relation) {
		if err == nil {
			err = ErrInvalidLiteral
		}
		return 0, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	values, err := manager.analysisDense(relation, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	selected := make([]bool, count)
	best := make(map[uint64]uint64)
	present := make(map[uint64]bool)
	for assignment, value := range values {
		if err = work.step(); err != nil {
			manager.rollback(&transaction)
			return 0, err
		}
		if !value {
			continue
		}
		var xv, yv uint64
		for bit := range x {
			xv = xv<<1 | uint64(assignment>>uint(x[bit])&1)
			yv = yv<<1 | uint64(assignment>>uint(y[bit])&1)
		}
		if !present[xv] || order == CuddPrioritySmallest && yv < best[xv] || order == CuddPriorityLargest && yv > best[xv] {
			best[xv], present[xv] = yv, true
		}
	}
	for assignment, value := range values {
		if err = work.step(); err != nil {
			manager.rollback(&transaction)
			return 0, err
		}
		if !value {
			continue
		}
		var xv, yv uint64
		for bit := range x {
			xv = xv<<1 | uint64(assignment>>uint(x[bit])&1)
			yv = yv<<1 | uint64(assignment>>uint(y[bit])&1)
		}
		selected[assignment] = yv == best[xv]
	}
	result, err := manager.analysisBuild(selected, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	return result, nil
}

func (manager *ExtraBDDManager) CompatibleProjection(relation ExtraBDDLit, projected []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if manager.cuddCheckVariables(projected) != nil {
		return 0, ErrDimension
	}
	marked := make([]bool, manager.variables)
	for _, variable := range projected {
		marked[variable] = true
	}
	retained := make([]int, 0, manager.variables-len(projected))
	for variable := 0; variable < manager.variables; variable++ {
		if !marked[variable] {
			retained = append(retained, variable)
		}
	}
	if len(projected) > 63 || len(retained) > 63 {
		return 0, ErrLimit
	}
	count, normalized, err := manager.analysisCount(options, 4)
	if err != nil || !manager.valid(relation) {
		if err == nil {
			err = ErrInvalidLiteral
		}
		return 0, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	values, err := manager.analysisDense(relation, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	best := make(map[uint64]uint64)
	present := make(map[uint64]bool)
	for assignment, value := range values {
		if err = work.step(); err != nil {
			manager.rollback(&transaction)
			return 0, err
		}
		if !value {
			continue
		}
		var key, choice uint64
		for _, variable := range retained {
			key = key<<1 | uint64(assignment>>uint(variable)&1)
		}
		for _, variable := range projected {
			choice = choice<<1 | uint64(assignment>>uint(variable)&1)
		}
		if !present[key] || choice < best[key] {
			best[key], present[key] = choice, true
		}
	}
	selected := make([]bool, count)
	for assignment, value := range values {
		if err = work.step(); err != nil {
			manager.rollback(&transaction)
			return 0, err
		}
		if !value {
			continue
		}
		var key, choice uint64
		for _, variable := range retained {
			key = key<<1 | uint64(assignment>>uint(variable)&1)
		}
		for _, variable := range projected {
			choice = choice<<1 | uint64(assignment>>uint(variable)&1)
		}
		selected[assignment] = choice == best[key]
	}
	result, err := manager.analysisBuild(selected, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	return result, nil
}

func (manager *ExtraBDDManager) PrioritySelectByBDD(relation, priority ExtraBDDLit, x, y, z []int, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if len(x) != len(y) || len(x) != len(z) || len(x) > 20 || manager.cuddCheckVariables(x, y, z) != nil {
		return 0, ErrDimension
	}
	count, normalized, err := manager.analysisCount(options, 4)
	if err != nil || !manager.valid(relation) || !manager.valid(priority) {
		if err == nil {
			err = ErrInvalidLiteral
		}
		return 0, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	relationValues, err := manager.analysisDense(relation, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	priorityValues, err := manager.analysisDense(priority, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	selected := make([]bool, count)
	zCount := uint64(1) << uint(len(z))
	for assignment, related := range relationValues {
		if err = work.step(); err != nil {
			manager.rollback(&transaction)
			return 0, err
		}
		if !related {
			continue
		}
		selected[assignment] = true
		for candidate := uint64(0); candidate < zCount; candidate++ {
			if err = work.step(); err != nil {
				manager.rollback(&transaction)
				return 0, err
			}
			relationAssignment, priorityAssignment := assignment, assignment
			for bit := range z {
				candidateValue := candidate>>uint(len(z)-1-bit)&1 != 0
				if candidateValue {
					relationAssignment |= 1 << uint(y[bit])
					priorityAssignment |= 1 << uint(z[bit])
				} else {
					relationAssignment &^= 1 << uint(y[bit])
					priorityAssignment &^= 1 << uint(z[bit])
				}
			}
			if relationValues[relationAssignment] && priorityValues[priorityAssignment] {
				selected[assignment] = false
				break
			}
		}
	}
	result, err := manager.analysisBuild(selected, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	return result, nil
}

func (manager *ExtraBDDManager) MinHammingDistance(root ExtraBDDLit, minterm []bool, options ExtraBDDOptions) (int, error) {
	if len(minterm) != manager.variables {
		return 0, ErrDimension
	}
	count, normalized, err := manager.analysisCount(options, 1)
	if err != nil || !manager.valid(root) {
		if err == nil {
			err = ErrInvalidLiteral
		}
		return 0, err
	}
	work := extraBDDWork{options: normalized}
	values, err := manager.analysisDense(root, count, &work)
	if err != nil {
		return 0, err
	}
	best := manager.variables + 1
	for assignment, value := range values {
		if err = work.step(); err != nil {
			return 0, err
		}
		if !value {
			continue
		}
		distance := 0
		for variable, expected := range minterm {
			if (assignment&(1<<uint(variable)) != 0) != expected {
				distance++
			}
		}
		best = min(best, distance)
	}
	if best > manager.variables {
		return -1, nil
	}
	return best, nil
}

func (manager *ExtraBDDManager) ClosestCube(first, second ExtraBDDLit, options ExtraBDDOptions) (CuddCube, error) {
	count, normalized, err := manager.analysisCount(options, 4)
	if err != nil || !manager.valid(first) || !manager.valid(second) {
		if err == nil {
			err = ErrInvalidLiteral
		}
		return CuddCube{}, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: normalized, transaction: &transaction}
	firstValues, err := manager.analysisDense(first, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return CuddCube{}, err
	}
	secondValues, err := manager.analysisDense(second, count, &work)
	if err != nil {
		manager.rollback(&transaction)
		return CuddCube{}, err
	}
	bestAssignment, bestDistance := -1, manager.variables+1
	for candidate, present := range firstValues {
		if err = work.step(); err != nil {
			manager.rollback(&transaction)
			return CuddCube{}, err
		}
		if !present {
			continue
		}
		for target, targetPresent := range secondValues {
			if !targetPresent {
				continue
			}
			if err = work.step(); err != nil {
				manager.rollback(&transaction)
				return CuddCube{}, err
			}
			distance := cuddApproxOnesCount(uint(candidate ^ target))
			if distance < bestDistance {
				bestAssignment, bestDistance = candidate, distance
			}
		}
	}
	if bestAssignment < 0 {
		return CuddCube{Root: ExtraBDDFalse}, nil
	}
	assignments := make([]ExtraBDDAssignment, manager.variables)
	root := ExtraBDDTrue
	for variable := manager.variables - 1; variable >= 0; variable-- {
		value := bestAssignment>>uint(variable)&1 != 0
		assignments[variable] = ExtraBDDAssignment{Variable: variable, Value: value}
		literal := ExtraBDDLit(variable+1) << 1
		if !value {
			literal = literal.Not()
		}
		root, err = manager.andRecursive(root, literal, &work)
		if err != nil {
			manager.rollback(&transaction)
			return CuddCube{}, err
		}
	}
	return CuddCube{Assignments: assignments, Root: root, Minterms: 1, Exists: true}, nil
}

func cuddApproxOnesCount(value uint) int {
	count := 0
	for value != 0 {
		value &= value - 1
		count++
	}
	return count
}
