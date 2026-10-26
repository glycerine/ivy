package abc

// Matrix operations from src/bdd/cudd/cuddMatMult.c.

// The source unit contains four public operations and three recursive helpers.
// The helpers are absorbed here by the canonical ADD apply and abstraction
// engines. MatrixMultiply and TimesPlus retain the same arithmetic-semiring
// meaning, Triangle retains min-plus multiplication, and OuterSum retains the
// pointwise min(M,r+c) operation. Variable indices replace source node pointers
// for identifying the shared matrix dimension. Dynamic reorder restarts and
// pointer-identity cache shortcuts do not apply to the fixed-order Go manager.

import "math"

type cuddMatrixCombine uint8

const (
	cuddMatrixProduct cuddMatrixCombine = iota
	cuddMatrixDistanceSum
	cuddMatrixPlus
	cuddMatrixMinimum
)

type cuddMatrixPair struct {
	left, right CuddADDEdge
	combine     cuddMatrixCombine
}

func cuddMatrixTerminal(combine cuddMatrixCombine, left, right float64) float64 {
	switch combine {
	case cuddMatrixProduct:
		// CUDD's times callback treats zero as absorbing, including when the
		// other leaf is infinite.
		if left == 0 || right == 0 {
			return 0
		}
		return left * right
	case cuddMatrixDistanceSum:
		// The shortest-path routines treat positive infinity as absorbing,
		// including the mixed signed-infinity case.
		if math.IsInf(left, 1) || math.IsInf(right, 1) {
			return math.Inf(1)
		}
		return left + right
	case cuddMatrixPlus:
		return left + right
	case cuddMatrixMinimum:
		return math.Min(left, right)
	default:
		return math.NaN()
	}
}

type cuddMatrixMemory struct {
	maxBytes    int64
	fixedBytes  int64
	memoEntries int64
}

func (memory *cuddMatrixMemory) memo(manager *CuddADDManager) error {
	memory.memoEntries++
	if int64(len(manager.nodes))*32+memory.fixedBytes+memory.memoEntries*32 > memory.maxBytes {
		return ErrLimit
	}
	return nil
}

func (memory *cuddMatrixMemory) resetMemo() { memory.memoEntries = 0 }

func (manager *CuddADDManager) matrixApply(
	combine cuddMatrixCombine,
	left, right CuddADDEdge,
	work *cuddADDWork,
	memo map[cuddMatrixPair]CuddADDEdge,
	memory *cuddMatrixMemory,
) (CuddADDEdge, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	key := cuddMatrixPair{left: left, right: right, combine: combine}
	if edge, ok := memo[key]; ok {
		return edge, nil
	}
	leftNode, rightNode := manager.nodes[left], manager.nodes[right]
	if leftNode.Terminal && rightNode.Terminal {
		edge, err := manager.constantUnchecked(cuddMatrixTerminal(combine, leftNode.Value, rightNode.Value))
		if err == nil {
			memo[key] = edge
			err = memory.memo(manager)
		}
		return edge, err
	}
	top := manager.top(left)
	if candidate := manager.top(right); candidate < top {
		top = candidate
	}
	leftHigh, leftLow := manager.branch(left, top)
	rightHigh, rightLow := manager.branch(right, top)
	high, err := manager.matrixApply(combine, leftHigh, rightHigh, work, memo, memory)
	if err != nil {
		return 0, err
	}
	low, err := manager.matrixApply(combine, leftLow, rightLow, work, memo, memory)
	if err != nil {
		return 0, err
	}
	edge, err := manager.nodeUnchecked(top, high, low)
	if err == nil {
		memo[key] = edge
		err = memory.memo(manager)
	}
	return edge, err
}

func (manager *CuddADDManager) matrixRun(
	edges []CuddADDEdge,
	variables []int,
	options ExtraBDDOptions,
	operation func(*cuddADDWork, []bool, *cuddMatrixMemory) (CuddADDEdge, error),
) (result CuddADDEdge, err error) {
	if err = manager.checkEdges(edges...); err != nil {
		return 0, err
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	selected := make([]bool, manager.variables)
	for _, variable := range variables {
		if variable < 0 || variable >= manager.variables {
			return 0, ErrDimension
		}
		selected[variable] = true
	}
	memory := &cuddMatrixMemory{maxBytes: int64(normalized.MaxBytes), fixedBytes: int64(len(selected))}
	if int64(len(manager.nodes))*32+memory.fixedBytes > memory.maxBytes {
		return 0, ErrLimit
	}

	// ADD nodes have a stable 32-byte accounting unit throughout this manager.
	// Tighten the node ceiling for this transaction so intermediate nodes cannot
	// exceed MaxBytes and then disappear before the final run check.
	byteNodeLimit := normalized.MaxBytes / 32
	if byteNodeLimit < len(manager.nodes) {
		return 0, ErrLimit
	}
	originalLimit := manager.maxNodes
	if byteNodeLimit < manager.maxNodes {
		manager.maxNodes = byteNodeLimit
		defer func() { manager.maxNodes = originalLimit }()
	}
	return manager.run(edges, normalized, func(work *cuddADDWork) (CuddADDEdge, error) {
		return operation(work, selected, memory)
	})
}

func (manager *CuddADDManager) matrixAbstract(
	edge CuddADDEdge,
	selected []bool,
	join cuddMatrixCombine,
	work *cuddADDWork,
	memory *cuddMatrixMemory,
) (CuddADDEdge, error) {
	type abstractionKey struct {
		edge     CuddADDEdge
		variable int
	}
	memo := make(map[abstractionKey]CuddADDEdge)
	applyMemo := make(map[cuddMatrixPair]CuddADDEdge)
	var visit func(CuddADDEdge, int) (CuddADDEdge, error)
	visit = func(current CuddADDEdge, variable int) (CuddADDEdge, error) {
		if err := work.step(); err != nil {
			return 0, err
		}
		if variable == manager.variables {
			return current, nil
		}
		key := abstractionKey{edge: current, variable: variable}
		if result, ok := memo[key]; ok {
			return result, nil
		}
		highEdge, lowEdge := manager.branch(current, variable)
		high, err := visit(highEdge, variable+1)
		if err != nil {
			return 0, err
		}
		low, err := visit(lowEdge, variable+1)
		if err != nil {
			return 0, err
		}
		var result CuddADDEdge
		if selected[variable] {
			result, err = manager.matrixApply(join, high, low, work, applyMemo, memory)
		} else {
			result, err = manager.nodeUnchecked(variable, high, low)
		}
		if err == nil {
			memo[key] = result
			err = memory.memo(manager)
		}
		return result, err
	}
	return visit(edge, 0)
}

func (manager *CuddADDManager) matrixProduct(
	left, right CuddADDEdge,
	variables []int,
	combine cuddMatrixCombine,
	join cuddMatrixCombine,
	options ExtraBDDOptions,
) (CuddADDEdge, error) {
	return manager.matrixRun(
		[]CuddADDEdge{left, right}, variables, options,
		func(work *cuddADDWork, selected []bool, memory *cuddMatrixMemory) (CuddADDEdge, error) {
			combined, err := manager.matrixApply(combine, left, right, work, make(map[cuddMatrixPair]CuddADDEdge), memory)
			if err != nil {
				return 0, err
			}
			memory.resetMemo()
			return manager.matrixAbstract(combined, selected, join, work, memory)
		},
	)
}

// MatrixMultiply returns the arithmetic matrix product after summing out the
// shared-dimension variables.
func (manager *CuddADDManager) MatrixMultiply(left, right CuddADDEdge, variables []int, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.matrixProduct(left, right, variables, cuddMatrixProduct, cuddMatrixPlus, options)
}

// TimesPlus implements the source unit's alternative arithmetic matrix-product
// entry point. Its result is canonical and therefore identical to
// MatrixMultiply for the same operands and variables.
func (manager *CuddADDManager) TimesPlus(left, right CuddADDEdge, variables []int, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.matrixProduct(left, right, variables, cuddMatrixProduct, cuddMatrixPlus, options)
}

// Triangle performs min-plus matrix multiplication after minimizing over the
// shared-dimension variables.
func (manager *CuddADDManager) Triangle(left, right CuddADDEdge, variables []int, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.matrixProduct(left, right, variables, cuddMatrixDistanceSum, cuddMatrixMinimum, options)
}

// OuterSum returns the pointwise minimum of matrix and row+column.
func (manager *CuddADDManager) OuterSum(matrix, row, column CuddADDEdge, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.matrixRun(
		[]CuddADDEdge{matrix, row, column}, nil, options,
		func(work *cuddADDWork, _ []bool, memory *cuddMatrixMemory) (CuddADDEdge, error) {
			sum, err := manager.matrixApply(cuddMatrixDistanceSum, row, column, work, make(map[cuddMatrixPair]CuddADDEdge), memory)
			if err != nil {
				return 0, err
			}
			memory.resetMemo()
			return manager.matrixApply(cuddMatrixMinimum, matrix, sum, work, make(map[cuddMatrixPair]CuddADDEdge), memory)
		},
	)
}
