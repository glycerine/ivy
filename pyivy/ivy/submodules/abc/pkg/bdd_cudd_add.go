package abc

// Numeric ADD operations from src/bdd/cudd/cuddAddAbs.c,
// src/bdd/cudd/cuddAddApply.c, src/bdd/cudd/cuddAddFind.c,
// src/bdd/cudd/cuddAddInv.c, src/bdd/cudd/cuddAddIte.c,
// src/bdd/cudd/cuddAddNeg.c, and src/bdd/cudd/cuddAddWalsh.c.

import "math"

type CuddADDEdge uint32

type CuddADDOperation uint8

const (
	CuddADDPlus CuddADDOperation = iota
	CuddADDTimes
	CuddADDThreshold
	CuddADDSetNZ
	CuddADDDivide
	CuddADDMinus
	CuddADDMinimum
	CuddADDMaximum
	CuddADDOneZeroMaximum
	CuddADDDiff
	CuddADDAgreement
	CuddADDOr
	CuddADDNand
	CuddADDNor
	CuddADDXor
	CuddADDXnor
)

const (
	CuddADDResidueMSB = 1 << iota
	CuddADDResidueTwosComplement
)

type CuddADDNode struct {
	Variable   int
	Then, Else CuddADDEdge
	Value      float64
	Terminal   bool
}

type cuddADDInternalKey struct {
	variable  int
	high, low CuddADDEdge
}

type CuddADDManager struct {
	variables int
	maxNodes  int
	nodes     []CuddADDNode
	terminals map[uint64]CuddADDEdge
	unique    map[cuddADDInternalKey]CuddADDEdge
	inputs    []CuddADDEdge
	closed    bool
}

type cuddADDWork struct {
	options    ExtraBDDOptions
	operations int
}

func (work *cuddADDWork) step() error {
	work.operations++
	if work.operations > work.options.MaxOperations {
		return ErrLimit
	}
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func cuddADDValueBits(value float64) (uint64, error) {
	if math.IsNaN(value) {
		return 0, ErrDimension
	}
	if value == 0 {
		value = 0
	}
	return math.Float64bits(value), nil
}

func NewCuddADDManager(variables, maxNodes int) (*CuddADDManager, error) {
	if variables < 0 || variables > 30 || maxNodes < variables+2 {
		return nil, ErrDimension
	}
	manager := &CuddADDManager{
		variables: variables,
		maxNodes:  maxNodes,
		terminals: make(map[uint64]CuddADDEdge),
		unique:    make(map[cuddADDInternalKey]CuddADDEdge),
	}
	zero, err := manager.constantUnchecked(0)
	if err != nil || zero != 0 {
		return nil, ErrInvariant
	}
	one, err := manager.constantUnchecked(1)
	if err != nil || one != 1 {
		return nil, ErrInvariant
	}
	for variable := 0; variable < variables; variable++ {
		edge, createErr := manager.nodeUnchecked(variable, one, zero)
		if createErr != nil {
			return nil, createErr
		}
		manager.inputs = append(manager.inputs, edge)
	}
	return manager, nil
}

func (manager *CuddADDManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed || manager.variables < 0 || manager.variables > 30 || manager.maxNodes < manager.variables+2 || len(manager.nodes) < manager.variables+2 || len(manager.nodes) > manager.maxNodes || len(manager.inputs) != manager.variables || manager.terminals == nil || manager.unique == nil {
		return ErrInvariant
	}
	for index, node := range manager.nodes {
		if node.Terminal {
			bits, err := cuddADDValueBits(node.Value)
			if err != nil || manager.terminals[bits] != CuddADDEdge(index) || node.Variable != -1 {
				return ErrInvariant
			}
			continue
		}
		if node.Variable < 0 || node.Variable >= manager.variables || int(node.Then) >= index || int(node.Else) >= index || node.Then == node.Else || manager.unique[cuddADDInternalKey{node.Variable, node.Then, node.Else}] != CuddADDEdge(index) {
			return ErrInvariant
		}
	}
	for variable, edge := range manager.inputs {
		node := manager.nodes[edge]
		if node.Terminal || node.Variable != variable || node.Then != 1 || node.Else != 0 {
			return ErrInvariant
		}
	}
	return nil
}

func (manager *CuddADDManager) rebuild() {
	manager.terminals = make(map[uint64]CuddADDEdge)
	manager.unique = make(map[cuddADDInternalKey]CuddADDEdge)
	for index, node := range manager.nodes {
		if node.Terminal {
			bits, _ := cuddADDValueBits(node.Value)
			manager.terminals[bits] = CuddADDEdge(index)
		} else {
			manager.unique[cuddADDInternalKey{node.Variable, node.Then, node.Else}] = CuddADDEdge(index)
		}
	}
}

func (manager *CuddADDManager) Close() error {
	if err := manager.check(); err != nil {
		return err
	}
	manager.closed = true
	manager.nodes = nil
	manager.inputs = nil
	manager.terminals = nil
	manager.unique = nil
	return nil
}

func (manager *CuddADDManager) Variables() int {
	if manager == nil {
		return 0
	}
	return manager.variables
}
func (manager *CuddADDManager) NodeCount() int {
	if manager == nil {
		return 0
	}
	return len(manager.nodes)
}

func (manager *CuddADDManager) valid(edge CuddADDEdge) bool {
	return uint64(edge) < uint64(len(manager.nodes))
}

func (manager *CuddADDManager) checkEdges(edges ...CuddADDEdge) error {
	if err := manager.check(); err != nil {
		return err
	}
	for _, edge := range edges {
		if !manager.valid(edge) {
			return ErrInvalidLiteral
		}
	}
	return nil
}

func (manager *CuddADDManager) constantUnchecked(value float64) (CuddADDEdge, error) {
	bits, err := cuddADDValueBits(value)
	if err != nil {
		return 0, err
	}
	if edge, ok := manager.terminals[bits]; ok {
		return edge, nil
	}
	if len(manager.nodes) >= manager.maxNodes {
		return 0, ErrLimit
	}
	edge := CuddADDEdge(len(manager.nodes))
	manager.nodes = append(manager.nodes, CuddADDNode{Variable: -1, Value: math.Float64frombits(bits), Terminal: true})
	manager.terminals[bits] = edge
	return edge, nil
}

func (manager *CuddADDManager) nodeUnchecked(variable int, high, low CuddADDEdge) (CuddADDEdge, error) {
	if high == low {
		return high, nil
	}
	key := cuddADDInternalKey{variable, high, low}
	if edge, ok := manager.unique[key]; ok {
		return edge, nil
	}
	if len(manager.nodes) >= manager.maxNodes {
		return 0, ErrLimit
	}
	edge := CuddADDEdge(len(manager.nodes))
	manager.nodes = append(manager.nodes, CuddADDNode{Variable: variable, Then: high, Else: low})
	manager.unique[key] = edge
	return edge, nil
}

func (manager *CuddADDManager) Constant(value float64) (CuddADDEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	start := len(manager.nodes)
	edge, err := manager.constantUnchecked(value)
	if err != nil {
		manager.nodes = manager.nodes[:start]
		manager.rebuild()
	}
	return edge, err
}

func (manager *CuddADDManager) Variable(variable int) (CuddADDEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if variable < 0 || variable >= manager.variables {
		return 0, ErrDimension
	}
	return manager.inputs[variable], nil
}

func (manager *CuddADDManager) Node(edge CuddADDEdge) (CuddADDNode, error) {
	if err := manager.checkEdges(edge); err != nil {
		return CuddADDNode{}, err
	}
	return manager.nodes[edge], nil
}

func (manager *CuddADDManager) Evaluate(edge CuddADDEdge, assignment uint64) (float64, error) {
	if err := manager.checkEdges(edge); err != nil {
		return 0, err
	}
	for {
		node := manager.nodes[edge]
		if node.Terminal {
			return node.Value, nil
		}
		if assignment&(uint64(1)<<uint(node.Variable)) != 0 {
			edge = node.Then
		} else {
			edge = node.Else
		}
	}
}

func (manager *CuddADDManager) evaluateUnchecked(edge CuddADDEdge, assignment uint64) float64 {
	for {
		node := manager.nodes[edge]
		if node.Terminal {
			return node.Value
		}
		if assignment&(uint64(1)<<uint(node.Variable)) != 0 {
			edge = node.Then
		} else {
			edge = node.Else
		}
	}
}

func (manager *CuddADDManager) Dense(edge CuddADDEdge) ([]float64, error) {
	if err := manager.checkEdges(edge); err != nil {
		return nil, err
	}
	size := 1 << uint(manager.variables)
	values := make([]float64, size)
	for assignment := range size {
		value, err := manager.Evaluate(edge, uint64(assignment))
		if err != nil {
			return nil, err
		}
		values[assignment] = value
	}
	return values, nil
}

func (manager *CuddADDManager) top(edge CuddADDEdge) int {
	node := manager.nodes[edge]
	if node.Terminal {
		return manager.variables
	}
	return node.Variable
}

func (manager *CuddADDManager) branch(edge CuddADDEdge, variable int) (CuddADDEdge, CuddADDEdge) {
	node := manager.nodes[edge]
	if !node.Terminal && node.Variable == variable {
		return node.Then, node.Else
	}
	return edge, edge
}

func (manager *CuddADDManager) run(edges []CuddADDEdge, options ExtraBDDOptions, operation func(*cuddADDWork) (CuddADDEdge, error)) (CuddADDEdge, error) {
	if err := manager.checkEdges(edges...); err != nil {
		return 0, err
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	start := len(manager.nodes)
	work := &cuddADDWork{options: normalized}
	edge, err := operation(work)
	if err == nil && len(manager.nodes)*32 > normalized.MaxBytes {
		err = ErrLimit
	}
	if err != nil {
		manager.nodes = manager.nodes[:start]
		manager.rebuild()
		return 0, err
	}
	return edge, nil
}

func cuddADDTerminalOperation(operation CuddADDOperation, left, right float64) (float64, error) {
	switch operation {
	case CuddADDPlus:
		return left + right, nil
	case CuddADDTimes:
		return left * right, nil
	case CuddADDThreshold:
		if left >= right {
			return left, nil
		}
		return 0, nil
	case CuddADDSetNZ:
		if right != 0 {
			return right, nil
		}
		return left, nil
	case CuddADDDivide:
		// CUDD defines zero divided by every leaf, including zero, as zero.
		if left == 0 {
			return 0, nil
		}
		return left / right, nil
	case CuddADDMinus:
		return left - right, nil
	case CuddADDMinimum:
		return math.Min(left, right), nil
	case CuddADDMaximum:
		return math.Max(left, right), nil
	case CuddADDOneZeroMaximum:
		if left > right {
			return 1, nil
		}
		return 0, nil
	case CuddADDDiff:
		if left == right {
			return math.Inf(1), nil
		}
		return math.Min(left, right), nil
	case CuddADDAgreement:
		if left == right {
			return left, nil
		}
		return 0, nil
	case CuddADDOr:
		if left != 0 || right != 0 {
			return 1, nil
		}
		return 0, nil
	case CuddADDNand:
		if left != 0 && right != 0 {
			return 0, nil
		}
		return 1, nil
	case CuddADDNor:
		if left == 0 && right == 0 {
			return 1, nil
		}
		return 0, nil
	case CuddADDXor:
		if (left != 0) != (right != 0) {
			return 1, nil
		}
		return 0, nil
	case CuddADDXnor:
		if (left != 0) == (right != 0) {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, ErrDimension
	}
}

type cuddADDPair struct {
	left, right CuddADDEdge
	operation   CuddADDOperation
}

func (manager *CuddADDManager) applyRecursive(operation CuddADDOperation, left, right CuddADDEdge, work *cuddADDWork, memo map[cuddADDPair]CuddADDEdge) (CuddADDEdge, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	key := cuddADDPair{left, right, operation}
	if edge, ok := memo[key]; ok {
		return edge, nil
	}
	leftNode, rightNode := manager.nodes[left], manager.nodes[right]
	if leftNode.Terminal && rightNode.Terminal {
		value, err := cuddADDTerminalOperation(operation, leftNode.Value, rightNode.Value)
		if err != nil {
			return 0, err
		}
		edge, err := manager.constantUnchecked(value)
		if err == nil {
			memo[key] = edge
		}
		return edge, err
	}
	top := manager.top(left)
	if candidate := manager.top(right); candidate < top {
		top = candidate
	}
	leftHigh, leftLow := manager.branch(left, top)
	rightHigh, rightLow := manager.branch(right, top)
	high, err := manager.applyRecursive(operation, leftHigh, rightHigh, work, memo)
	if err != nil {
		return 0, err
	}
	low, err := manager.applyRecursive(operation, leftLow, rightLow, work, memo)
	if err != nil {
		return 0, err
	}
	edge, err := manager.nodeUnchecked(top, high, low)
	if err == nil {
		memo[key] = edge
	}
	return edge, err
}

func (manager *CuddADDManager) Apply(operation CuddADDOperation, left, right CuddADDEdge, options ExtraBDDOptions) (CuddADDEdge, error) {
	if operation > CuddADDXnor {
		return 0, ErrDimension
	}
	if operation >= CuddADDOr {
		if err := manager.checkEdges(left, right); err != nil {
			return 0, err
		}
		seen := make(map[CuddADDEdge]bool)
		if !manager.boolean(left, seen) || !manager.boolean(right, seen) {
			return 0, ErrDimension
		}
	}
	return manager.run([]CuddADDEdge{left, right}, options, func(work *cuddADDWork) (CuddADDEdge, error) {
		return manager.applyRecursive(operation, left, right, work, make(map[cuddADDPair]CuddADDEdge))
	})
}

type cuddADDIteKey struct{ condition, then, otherwise CuddADDEdge }

func (manager *CuddADDManager) boolean(edge CuddADDEdge, seen map[CuddADDEdge]bool) bool {
	if result, ok := seen[edge]; ok {
		return result
	}
	node := manager.nodes[edge]
	if node.Terminal {
		result := node.Value == 0 || node.Value == 1
		seen[edge] = result
		return result
	}
	result := manager.boolean(node.Then, seen) && manager.boolean(node.Else, seen)
	seen[edge] = result
	return result
}

func (manager *CuddADDManager) iteRecursive(condition, then, otherwise CuddADDEdge, work *cuddADDWork, memo map[cuddADDIteKey]CuddADDEdge) (CuddADDEdge, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if condition == 1 {
		return then, nil
	}
	if condition == 0 {
		return otherwise, nil
	}
	if then == otherwise {
		return then, nil
	}
	key := cuddADDIteKey{condition, then, otherwise}
	if edge, ok := memo[key]; ok {
		return edge, nil
	}
	top := manager.top(condition)
	if v := manager.top(then); v < top {
		top = v
	}
	if v := manager.top(otherwise); v < top {
		top = v
	}
	ch, cl := manager.branch(condition, top)
	th, tl := manager.branch(then, top)
	oh, ol := manager.branch(otherwise, top)
	high, err := manager.iteRecursive(ch, th, oh, work, memo)
	if err != nil {
		return 0, err
	}
	low, err := manager.iteRecursive(cl, tl, ol, work, memo)
	if err != nil {
		return 0, err
	}
	edge, err := manager.nodeUnchecked(top, high, low)
	if err == nil {
		memo[key] = edge
	}
	return edge, err
}

func (manager *CuddADDManager) Ite(condition, then, otherwise CuddADDEdge, options ExtraBDDOptions) (CuddADDEdge, error) {
	if err := manager.checkEdges(condition, then, otherwise); err != nil {
		return 0, err
	}
	if !manager.boolean(condition, make(map[CuddADDEdge]bool)) {
		return 0, ErrDimension
	}
	return manager.run([]CuddADDEdge{condition, then, otherwise}, options, func(work *cuddADDWork) (CuddADDEdge, error) {
		return manager.iteRecursive(condition, then, otherwise, work, make(map[cuddADDIteKey]CuddADDEdge))
	})
}

func (manager *CuddADDManager) IteConstant(condition, then, otherwise CuddADDEdge, options ExtraBDDOptions) (float64, bool, error) {
	if err := manager.checkEdges(condition, then, otherwise); err != nil {
		return 0, false, err
	}
	if !manager.boolean(condition, make(map[CuddADDEdge]bool)) {
		return 0, false, ErrDimension
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, false, err
	}
	work := cuddADDWork{options: normalized}
	found := false
	result := 0.0
	for assignment := 0; assignment < 1<<uint(manager.variables); assignment++ {
		if err = work.step(); err != nil {
			return 0, false, err
		}
		selected, _ := manager.Evaluate(condition, uint64(assignment))
		edge := otherwise
		if selected != 0 {
			edge = then
		}
		value, _ := manager.Evaluate(edge, uint64(assignment))
		if !found {
			result, found = value, true
		} else if result != value {
			return 0, false, nil
		}
	}
	return result, true, nil
}

func (manager *CuddADDManager) EvalConst(condition, value CuddADDEdge, background float64, options ExtraBDDOptions) (float64, bool, error) {
	if err := manager.checkEdges(condition, value); err != nil {
		return 0, false, err
	}
	if !manager.boolean(condition, make(map[CuddADDEdge]bool)) {
		return 0, false, ErrDimension
	}
	found := false
	result := background
	workOptions, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, false, err
	}
	work := cuddADDWork{options: workOptions}
	for assignment := 0; assignment < 1<<uint(manager.variables); assignment++ {
		if err = work.step(); err != nil {
			return 0, false, err
		}
		if manager.evaluateUnchecked(condition, uint64(assignment)) != 0 {
			selected := manager.evaluateUnchecked(value, uint64(assignment))
			if !found {
				result, found = selected, true
			} else if result != selected {
				return 0, false, nil
			}
		}
	}
	return result, true, nil
}

func (manager *CuddADDManager) unary(edge CuddADDEdge, options ExtraBDDOptions, terminal func(float64) (float64, error)) (CuddADDEdge, error) {
	return manager.run([]CuddADDEdge{edge}, options, func(work *cuddADDWork) (CuddADDEdge, error) {
		memo := make(map[CuddADDEdge]CuddADDEdge)
		var visit func(CuddADDEdge) (CuddADDEdge, error)
		visit = func(current CuddADDEdge) (CuddADDEdge, error) {
			if err := work.step(); err != nil {
				return 0, err
			}
			if result, ok := memo[current]; ok {
				return result, nil
			}
			node := manager.nodes[current]
			if node.Terminal {
				value, err := terminal(node.Value)
				if err != nil {
					return 0, err
				}
				result, err := manager.constantUnchecked(value)
				if err == nil {
					memo[current] = result
				}
				return result, err
			}
			high, err := visit(node.Then)
			if err != nil {
				return 0, err
			}
			low, err := visit(node.Else)
			if err != nil {
				return 0, err
			}
			result, err := manager.nodeUnchecked(node.Variable, high, low)
			if err == nil {
				memo[current] = result
			}
			return result, err
		}
		return visit(edge)
	})
}

func (manager *CuddADDManager) Complement(edge CuddADDEdge, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.unary(edge, options, func(value float64) (float64, error) {
		if value == 0 {
			return 1, nil
		}
		return 0, nil
	})
}
func (manager *CuddADDManager) Negate(edge CuddADDEdge, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.unary(edge, options, func(value float64) (float64, error) { return -value, nil })
}
func (manager *CuddADDManager) Log(edge CuddADDEdge, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.unary(edge, options, func(value float64) (float64, error) { return math.Log(value), nil })
}
func (manager *CuddADDManager) RoundOff(edge CuddADDEdge, digits int, options ExtraBDDOptions) (CuddADDEdge, error) {
	if digits < 0 || digits > 15 {
		return 0, ErrDimension
	}
	scale := math.Pow10(digits)
	return manager.unary(edge, options, func(value float64) (float64, error) { return math.Ceil(value*scale) / scale, nil })
}
func (manager *CuddADDManager) ScalarInverse(edge CuddADDEdge, epsilon float64, options ExtraBDDOptions) (CuddADDEdge, error) {
	if math.IsNaN(epsilon) || epsilon < 0 {
		return 0, ErrDimension
	}
	return manager.unary(edge, options, func(value float64) (float64, error) {
		if math.Abs(value) < epsilon {
			return 0, ErrDimension
		}
		return 1 / value, nil
	})
}

func (manager *CuddADDManager) Leq(left, right CuddADDEdge, options ExtraBDDOptions) (bool, error) {
	if err := manager.checkEdges(left, right); err != nil {
		return false, err
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return false, err
	}
	work := cuddADDWork{options: normalized}
	type pair struct{ left, right CuddADDEdge }
	memo := make(map[pair]bool)
	var visit func(CuddADDEdge, CuddADDEdge) (bool, error)
	visit = func(a, b CuddADDEdge) (bool, error) {
		if err := work.step(); err != nil {
			return false, err
		}
		key := pair{a, b}
		if result, ok := memo[key]; ok {
			return result, nil
		}
		leftNode, rightNode := manager.nodes[a], manager.nodes[b]
		if leftNode.Terminal && rightNode.Terminal {
			result := leftNode.Value <= rightNode.Value
			memo[key] = result
			return result, nil
		}
		top := min(manager.top(a), manager.top(b))
		aHigh, aLow := manager.branch(a, top)
		bHigh, bLow := manager.branch(b, top)
		high, err := visit(aHigh, bHigh)
		if err != nil || !high {
			return high, err
		}
		low, err := visit(aLow, bLow)
		if err == nil {
			memo[key] = low
		}
		return low, err
	}
	return visit(left, right)
}

func (manager *CuddADDManager) FindMax(edge CuddADDEdge, options ExtraBDDOptions) (float64, error) {
	return manager.find(edge, options, true)
}
func (manager *CuddADDManager) FindMin(edge CuddADDEdge, options ExtraBDDOptions) (float64, error) {
	return manager.find(edge, options, false)
}
func (manager *CuddADDManager) find(edge CuddADDEdge, options ExtraBDDOptions, maximum bool) (float64, error) {
	if err := manager.checkEdges(edge); err != nil {
		return 0, err
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	work := cuddADDWork{options: normalized}
	memo := make(map[CuddADDEdge]float64)
	var visit func(CuddADDEdge) (float64, error)
	visit = func(current CuddADDEdge) (float64, error) {
		if err := work.step(); err != nil {
			return 0, err
		}
		if result, ok := memo[current]; ok {
			return result, nil
		}
		node := manager.nodes[current]
		if node.Terminal {
			memo[current] = node.Value
			return node.Value, nil
		}
		high, err := visit(node.Then)
		if err != nil {
			return 0, err
		}
		low, err := visit(node.Else)
		if err != nil {
			return 0, err
		}
		result := low
		if maximum && high > low || !maximum && high < low {
			result = high
		}
		memo[current] = result
		return result, nil
	}
	return visit(edge)
}

func (manager *CuddADDManager) IthBit(edge CuddADDEdge, bit int, options ExtraBDDOptions) (CuddADDEdge, error) {
	if bit < 0 || bit >= 63 {
		return 0, ErrDimension
	}
	return manager.unary(edge, options, func(value float64) (float64, error) {
		if math.IsInf(value, 0) || value < math.MinInt64 || value >= -float64(math.MinInt64) {
			return 0, ErrDimension
		}
		integer := int64(value)
		if uint64(integer)&(uint64(1)<<uint(bit)) != 0 {
			return 1, nil
		}
		return 0, nil
	})
}

func (manager *CuddADDManager) abstract(edge CuddADDEdge, variables []int, operation CuddADDOperation, options ExtraBDDOptions) (CuddADDEdge, error) {
	if err := manager.checkEdges(edge); err != nil {
		return 0, err
	}
	selected := make([]bool, manager.variables)
	for _, variable := range variables {
		if variable < 0 || variable >= manager.variables {
			return 0, ErrDimension
		}
		selected[variable] = true
	}
	if operation == CuddADDOr && !manager.boolean(edge, make(map[CuddADDEdge]bool)) {
		return 0, ErrDimension
	}
	return manager.run([]CuddADDEdge{edge}, options, func(work *cuddADDWork) (CuddADDEdge, error) {
		type abstractionKey struct {
			edge     CuddADDEdge
			variable int
		}
		memo := make(map[abstractionKey]CuddADDEdge)
		applyMemo := make(map[cuddADDPair]CuddADDEdge)
		var visit func(CuddADDEdge, int) (CuddADDEdge, error)
		visit = func(current CuddADDEdge, variable int) (CuddADDEdge, error) {
			if err := work.step(); err != nil {
				return 0, err
			}
			if variable == manager.variables {
				return current, nil
			}
			key := abstractionKey{current, variable}
			if result, ok := memo[key]; ok {
				return result, nil
			}
			node := manager.nodes[current]
			highEdge, lowEdge := current, current
			if !node.Terminal && node.Variable == variable {
				highEdge, lowEdge = node.Then, node.Else
			}
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
				result, err = manager.applyRecursive(operation, high, low, work, applyMemo)
			} else {
				result, err = manager.nodeUnchecked(variable, high, low)
			}
			if err == nil {
				memo[key] = result
			}
			return result, err
		}
		return visit(edge, 0)
	})
}
func (manager *CuddADDManager) ExistAbstract(edge CuddADDEdge, variables []int, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.abstract(edge, variables, CuddADDPlus, options)
}
func (manager *CuddADDManager) UnivAbstract(edge CuddADDEdge, variables []int, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.abstract(edge, variables, CuddADDTimes, options)
}
func (manager *CuddADDManager) OrAbstract(edge CuddADDEdge, variables []int, options ExtraBDDOptions) (CuddADDEdge, error) {
	return manager.abstract(edge, variables, CuddADDOr, options)
}

func (manager *CuddADDManager) fromDense(values []float64, work *cuddADDWork, extraBytes int64) (CuddADDEdge, error) {
	if int64(len(manager.nodes))*32+extraBytes+int64(len(values))*4 > int64(work.options.MaxBytes) {
		return 0, ErrLimit
	}
	current := make([]CuddADDEdge, len(values))
	for index, value := range values {
		if err := work.step(); err != nil {
			return 0, err
		}
		edge, err := manager.constantUnchecked(value)
		if err != nil {
			return 0, err
		}
		current[index] = edge
		if int64(len(manager.nodes))*32+extraBytes+int64(len(current))*4 > int64(work.options.MaxBytes) {
			return 0, ErrLimit
		}
	}
	for variable, stride := manager.variables-1, len(current)/2; variable >= 0; variable, stride = variable-1, stride/2 {
		if int64(len(manager.nodes))*32+extraBytes+int64(len(current)+stride)*4 > int64(work.options.MaxBytes) {
			return 0, ErrLimit
		}
		next := make([]CuddADDEdge, stride)
		for i := range stride {
			if err := work.step(); err != nil {
				return 0, err
			}
			edge, err := manager.nodeUnchecked(variable, current[i+stride], current[i])
			if err != nil {
				return 0, err
			}
			next[i] = edge
			if int64(len(manager.nodes))*32+extraBytes+int64(len(current)+len(next))*4 > int64(work.options.MaxBytes) {
				return 0, ErrLimit
			}
		}
		current = next
	}
	return current[0], nil
}

func (manager *CuddADDManager) FromDense(values []float64, options ExtraBDDOptions) (CuddADDEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if len(values) != 1<<uint(manager.variables) {
		return 0, ErrDimension
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	start := len(manager.nodes)
	edge, err := manager.fromDense(values, &cuddADDWork{options: normalized}, 0)
	if err != nil {
		manager.nodes = manager.nodes[:start]
		manager.rebuild()
		return 0, err
	}
	return edge, nil
}

func (manager *CuddADDManager) Walsh(xVariables, yVariables []int, options ExtraBDDOptions) (CuddADDEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if len(xVariables) != len(yVariables) {
		return 0, ErrDimension
	}
	used := make([]bool, manager.variables)
	for _, list := range [][]int{xVariables, yVariables} {
		for _, variable := range list {
			if variable < 0 || variable >= manager.variables || used[variable] {
				return 0, ErrDimension
			}
			used[variable] = true
		}
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	size := 1 << uint(manager.variables)
	if int64(len(manager.nodes))*32+int64(size)*12 > int64(normalized.MaxBytes) {
		return 0, ErrLimit
	}
	start := len(manager.nodes)
	work := cuddADDWork{options: normalized}
	values := make([]float64, size)
	for assignment := range values {
		if err = work.step(); err != nil {
			break
		}
		parity := 0
		for index := range xVariables {
			if assignment&(1<<uint(xVariables[index])) != 0 && assignment&(1<<uint(yVariables[index])) != 0 {
				parity ^= 1
			}
		}
		if parity == 0 {
			values[assignment] = 1
		} else {
			values[assignment] = -1
		}
	}
	var edge CuddADDEdge
	if err == nil {
		edge, err = manager.fromDense(values, &work, int64(size)*8)
	}
	if err != nil {
		manager.nodes = manager.nodes[:start]
		manager.rebuild()
	}
	return edge, err
}

func (manager *CuddADDManager) Residue(bits, modulus, flags, top int, options ExtraBDDOptions) (CuddADDEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if bits < 1 || modulus < 2 || top < 0 || top+bits > manager.variables || flags&^(CuddADDResidueMSB|CuddADDResidueTwosComplement) != 0 {
		return 0, ErrDimension
	}
	normalized, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	size := 1 << uint(manager.variables)
	if int64(len(manager.nodes))*32+int64(size)*12 > int64(normalized.MaxBytes) {
		return 0, ErrLimit
	}
	start := len(manager.nodes)
	work := cuddADDWork{options: normalized}
	values := make([]float64, size)
	for assignment := range values {
		if err = work.step(); err != nil {
			break
		}
		number := 0
		for index := 0; index < bits; index++ {
			variable := top + index
			weight := index
			if flags&CuddADDResidueMSB != 0 {
				weight = bits - index - 1
			}
			if assignment&(1<<uint(variable)) != 0 {
				number += 1 << uint(weight)
			}
		}
		if flags&CuddADDResidueTwosComplement != 0 {
			signVariable := top + bits - 1
			if flags&CuddADDResidueMSB != 0 {
				signVariable = top
			}
			if assignment&(1<<uint(signVariable)) != 0 {
				number -= 1 << uint(bits)
			}
		}
		residue := number % modulus
		if residue < 0 {
			residue += modulus
		}
		values[assignment] = float64(residue)
	}
	var edge CuddADDEdge
	if err == nil {
		edge, err = manager.fromDense(values, &work, int64(size)*8)
	}
	if err != nil {
		manager.nodes = manager.nodes[:start]
		manager.rebuild()
	}
	return edge, err
}
