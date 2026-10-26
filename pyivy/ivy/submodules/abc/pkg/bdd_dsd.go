package abc

// Disjoint-support decomposition from src/bdd/dsd.

import (
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
)

// DSDNodeType is the source node classification.
type DSDNodeType uint8

const (
	DSDNodeNone DSDNodeType = iota
	DSDNodeConst1
	DSDNodeBuffer
	DSDNodeOR
	DSDNodeXOR
	DSDNodePrime
)

// DSDEdge replaces complemented Dsd_Node_t pointers with an explicit bit.
type DSDEdge struct {
	Node       int
	Complement bool
}

func (edge DSDEdge) Not() DSDEdge {
	edge.Complement = !edge.Complement
	return edge
}

// DSDNode is an owned decomposition node snapshot. Function uses the manager's
// original variable space. PrimeFunction uses one variable per child.
type DSDNode struct {
	Type          DSDNodeType
	Variable      int
	Function      *TruthTable
	PrimeFunction *TruthTable
	Support       uint64
	Children      []DSDEdge
	Mark          int
}

func cloneDSDNode(node DSDNode) DSDNode {
	node.Function = node.Function.Clone()
	node.PrimeFunction = node.PrimeFunction.Clone()
	node.Children = slices.Clone(node.Children)
	return node
}

// DSDOptions bounds decomposition work. Zero numeric fields select defaults.
type DSDOptions struct {
	MaxVariables  int
	MaxNodes      int
	MaxRoots      int
	MaxOperations int
	MaxWords      int
	Stop          func() bool
}

func normalizeDSDOptions(options DSDOptions) (DSDOptions, error) {
	if options.MaxVariables < 0 || options.MaxNodes < 0 || options.MaxRoots < 0 || options.MaxOperations < 0 || options.MaxWords < 0 {
		return options, ErrDimension
	}
	if options.MaxVariables == 0 {
		options.MaxVariables = 12
	}
	if options.MaxVariables > 20 || options.MaxVariables > 63 {
		return options, ErrLimit
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxRoots == 0 {
		options.MaxRoots = 1 << 16
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 28
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 14
	}
	return options, nil
}

// DSDManager owns all nodes, roots, and structural lookup state.
type DSDManager struct {
	variables  int
	options    DSDOptions
	nodes      []DSDNode
	inputs     []DSDEdge
	roots      []DSDEdge
	cache      map[string]DSDEdge
	operations int
	closed     bool
}

// NewDSDManager creates constant-one and input nodes in source order.
func NewDSDManager(variables int, options DSDOptions) (*DSDManager, error) {
	options, err := normalizeDSDOptions(options)
	if err != nil {
		return nil, err
	}
	if variables < 0 || variables > options.MaxVariables {
		return nil, ErrDimension
	}
	if variables+1 > options.MaxNodes {
		return nil, ErrLimit
	}
	manager := &DSDManager{
		variables: variables,
		options:   options,
		cache:     make(map[string]DSDEdge, variables+1),
	}
	one, err := NewConstantTruthTable(variables, true, options.MaxWords)
	if err != nil {
		return nil, err
	}
	manager.nodes = append(manager.nodes, DSDNode{Type: DSDNodeConst1, Variable: -1, Function: one})
	manager.cache[dsdTruthKey(one)] = DSDEdge{Node: 0}
	for variable := 0; variable < variables; variable++ {
		function, makeErr := NewVariableTruthTable(variables, variable, options.MaxWords)
		if makeErr != nil {
			return nil, makeErr
		}
		edge := DSDEdge{Node: len(manager.nodes)}
		manager.nodes = append(manager.nodes, DSDNode{Type: DSDNodeBuffer, Variable: variable, Function: function, Support: uint64(1) << uint(variable)})
		manager.inputs = append(manager.inputs, edge)
		manager.cache[dsdTruthKey(function)] = edge
	}
	return manager, nil
}

func (manager *DSDManager) check() error {
	if manager == nil || manager.closed || manager.variables < 0 || len(manager.nodes) < manager.variables+1 || len(manager.inputs) != manager.variables || manager.cache == nil {
		return ErrInvariant
	}
	for index, node := range manager.nodes {
		if node.Type < DSDNodeConst1 || node.Type > DSDNodePrime || node.Function == nil || node.Function.Variables() != manager.variables {
			return ErrInvariant
		}
		if node.Type == DSDNodeBuffer && (node.Variable < 0 || node.Variable >= manager.variables) {
			return ErrInvariant
		}
		for _, child := range node.Children {
			if child.Node < 0 || child.Node >= index {
				return ErrInvariant
			}
		}
		if node.Type == DSDNodePrime && (node.PrimeFunction == nil || node.PrimeFunction.Variables() != len(node.Children)) {
			return ErrInvariant
		}
	}
	for _, edge := range manager.roots {
		if edge.Node < 0 || edge.Node >= len(manager.nodes) {
			return ErrInvariant
		}
	}
	return nil
}

func (manager *DSDManager) tick(count int) error {
	if count < 0 || manager.operations > manager.options.MaxOperations-count {
		return ErrLimit
	}
	manager.operations += count
	if manager.options.Stop != nil && manager.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func dsdTruthKey(function *TruthTable) string {
	words := function.Words()
	bytes := make([]byte, 8*len(words)+2)
	bytes[0] = byte(function.Variables())
	bytes[1] = byte(function.Variables() >> 8)
	for index, word := range words {
		for shift := 0; shift < 8; shift++ {
			bytes[2+8*index+shift] = byte(word >> uint(8*shift))
		}
	}
	return string(bytes)
}

func dsdTruthEqual(first, second *TruthTable) bool {
	if first == nil || second == nil || first.Variables() != second.Variables() {
		return false
	}
	return slices.Equal(first.Words(), second.Words())
}

func (manager *DSDManager) truthSupport(function *TruthTable) (uint64, error) {
	if err := function.Check(); err != nil {
		return 0, err
	}
	variables := function.Variables()
	if variables > 63 {
		return 0, ErrLimit
	}
	assignments := uint64(1) << uint(variables)
	var support uint64
	for variable := 0; variable < variables; variable++ {
		bit := uint64(1) << uint(variable)
		for assignment := uint64(0); assignment < assignments; assignment++ {
			if assignment&bit != 0 {
				continue
			}
			if err := manager.tick(1); err != nil {
				return 0, err
			}
			zero, _ := function.Get(assignment)
			one, _ := function.Get(assignment | bit)
			if zero != one {
				support |= bit
				break
			}
		}
	}
	return support, nil
}

type dsdTransaction struct {
	nodes, roots, operations int
	cacheKeys                []string
}

func (manager *DSDManager) rollback(transaction dsdTransaction) {
	manager.nodes = manager.nodes[:transaction.nodes]
	manager.roots = manager.roots[:transaction.roots]
	manager.operations = transaction.operations
	for _, key := range transaction.cacheKeys {
		delete(manager.cache, key)
	}
}

func (manager *DSDManager) build(function *TruthTable, transaction *dsdTransaction) (DSDEdge, error) {
	if err := manager.tick(1); err != nil {
		return DSDEdge{}, err
	}
	if function == nil || function.Variables() != manager.variables {
		return DSDEdge{}, ErrDimension
	}
	if err := function.Check(); err != nil {
		return DSDEdge{}, err
	}
	base, _ := function.Get(0)
	regular := function
	complement := false
	if base {
		regular = function.Not()
		complement = true
	}
	key := dsdTruthKey(regular)
	if edge, exists := manager.cache[key]; exists {
		edge.Complement = edge.Complement != complement
		return edge, nil
	}
	support, err := manager.truthSupport(regular)
	if err != nil {
		return DSDEdge{}, err
	}
	if support == 0 {
		return DSDEdge{Node: 0, Complement: true != complement}, nil
	}
	children, kind, found, err := manager.findDecomposition(regular, support, transaction)
	if err != nil {
		return DSDEdge{}, err
	}
	if !found {
		children = make([]DSDEdge, 0, dsdBitsCount(support))
		for variable := 0; variable < manager.variables; variable++ {
			if support&(uint64(1)<<uint(variable)) != 0 {
				children = append(children, manager.inputs[variable])
			}
		}
		kind = DSDNodePrime
	}
	if len(manager.nodes) >= manager.options.MaxNodes {
		return DSDEdge{}, ErrLimit
	}
	node := DSDNode{Type: kind, Variable: -1, Function: regular.Clone(), Support: support, Children: children}
	if kind == DSDNodePrime {
		node.PrimeFunction, err = manager.localPrime(regular, support)
		if err != nil {
			return DSDEdge{}, err
		}
	}
	edge := DSDEdge{Node: len(manager.nodes)}
	manager.nodes = append(manager.nodes, node)
	manager.cache[key] = edge
	transaction.cacheKeys = append(transaction.cacheKeys, key)
	edge.Complement = complement
	return edge, nil
}

func dsdBitsCount(value uint64) int {
	count := 0
	for value != 0 {
		value &= value - 1
		count++
	}
	return count
}

func dsdDeposit(local, mask uint64) uint64 {
	var global uint64
	for localBit := uint(0); mask != 0; localBit++ {
		lowest := mask & -mask
		if local&(uint64(1)<<localBit) != 0 {
			global |= lowest
		}
		mask &^= lowest
	}
	return global
}

func (manager *DSDManager) partitionFunctions(function *TruthTable, firstMask, secondMask uint64, xor bool) (*TruthTable, *TruthTable, bool, error) {
	first, err := NewConstantTruthTable(manager.variables, false, manager.options.MaxWords)
	if err != nil {
		return nil, nil, false, err
	}
	second, err := NewConstantTruthTable(manager.variables, false, manager.options.MaxWords)
	if err != nil {
		return nil, nil, false, err
	}
	firstCount := uint64(1) << uint(dsdBitsCount(firstMask))
	secondCount := uint64(1) << uint(dsdBitsCount(secondMask))
	firstValues := make([]bool, firstCount)
	secondValues := make([]bool, secondCount)
	for a := uint64(0); a < firstCount; a++ {
		firstValues[a], _ = function.Get(dsdDeposit(a, firstMask))
	}
	for b := uint64(0); b < secondCount; b++ {
		secondValues[b], _ = function.Get(dsdDeposit(b, secondMask))
	}
	for a := uint64(0); a < firstCount; a++ {
		for b := uint64(0); b < secondCount; b++ {
			if err = manager.tick(1); err != nil {
				return nil, nil, false, err
			}
			actual, _ := function.Get(dsdDeposit(a, firstMask) | dsdDeposit(b, secondMask))
			expected := firstValues[a] != secondValues[b]
			if !xor {
				expected = firstValues[a] || secondValues[b]
			}
			if actual != expected {
				return nil, nil, false, nil
			}
		}
	}
	assignments := uint64(1) << uint(manager.variables)
	for assignment := uint64(0); assignment < assignments; assignment++ {
		firstValue, _ := function.Get(assignment & firstMask)
		secondValue, _ := function.Get(assignment & secondMask)
		_ = first.Set(assignment, firstValue)
		_ = second.Set(assignment, secondValue)
	}
	return first, second, true, nil
}

func (manager *DSDManager) findDecomposition(function *TruthTable, support uint64, transaction *dsdTransaction) ([]DSDEdge, DSDNodeType, bool, error) {
	lowest := support & -support
	for subset := (support - 1) & support; subset != 0; subset = (subset - 1) & support {
		if subset&lowest == 0 || subset == support {
			continue
		}
		other := support &^ subset
		for _, kind := range []DSDNodeType{DSDNodeOR, DSDNodeXOR} {
			first, second, ok, err := manager.partitionFunctions(function, subset, other, kind == DSDNodeXOR)
			if err != nil {
				return nil, DSDNodeNone, false, err
			}
			if !ok {
				continue
			}
			firstEdge, err := manager.build(first, transaction)
			if err != nil {
				return nil, DSDNodeNone, false, err
			}
			secondEdge, err := manager.build(second, transaction)
			if err != nil {
				return nil, DSDNodeNone, false, err
			}
			children := manager.flatten(kind, firstEdge, secondEdge)
			return children, kind, true, nil
		}
	}
	return nil, DSDNodeNone, false, nil
}

func (manager *DSDManager) flatten(kind DSDNodeType, edges ...DSDEdge) []DSDEdge {
	result := make([]DSDEdge, 0, len(edges))
	for _, edge := range edges {
		node := manager.nodes[edge.Node]
		if !edge.Complement && node.Type == kind {
			result = append(result, node.Children...)
		} else {
			result = append(result, edge)
		}
	}
	return result
}

func (manager *DSDManager) localPrime(function *TruthTable, support uint64) (*TruthTable, error) {
	variables := dsdBitsCount(support)
	local, err := NewConstantTruthTable(variables, false, manager.options.MaxWords)
	if err != nil {
		return nil, err
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
		value, _ := function.Get(dsdDeposit(assignment, support))
		if err = local.Set(assignment, value); err != nil {
			return nil, err
		}
	}
	return local, nil
}

// Decompose atomically replaces the manager's roots.
func (manager *DSDManager) Decompose(functions []*TruthTable) ([]DSDEdge, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	if len(functions) > manager.options.MaxRoots {
		return nil, ErrLimit
	}
	transaction := dsdTransaction{nodes: len(manager.nodes), roots: len(manager.roots), operations: manager.operations}
	newRoots := make([]DSDEdge, 0, len(functions))
	for _, function := range functions {
		edge, err := manager.build(function, &transaction)
		if err != nil {
			manager.rollback(transaction)
			return nil, err
		}
		newRoots = append(newRoots, edge)
	}
	manager.roots = slices.Clone(newRoots)
	return slices.Clone(newRoots), nil
}

// DecomposeOne adds nodes transactionally without changing the root array.
func (manager *DSDManager) DecomposeOne(function *TruthTable) (DSDEdge, error) {
	if err := manager.check(); err != nil {
		return DSDEdge{}, err
	}
	transaction := dsdTransaction{nodes: len(manager.nodes), roots: len(manager.roots), operations: manager.operations}
	edge, err := manager.build(function, &transaction)
	if err != nil {
		manager.rollback(transaction)
	}
	return edge, err
}

// DecomposeBDD converts a bounded concrete BDD to an owned truth table.
func (manager *DSDManager) DecomposeBDD(bdd *ExtraBDDManager, root ExtraBDDLit) (DSDEdge, error) {
	if err := manager.check(); err != nil {
		return DSDEdge{}, err
	}
	if bdd == nil || bdd.variables != manager.variables {
		return DSDEdge{}, ErrDimension
	}
	if err := bdd.checkLiteral(root); err != nil {
		return DSDEdge{}, err
	}
	operations := manager.operations
	function, err := NewConstantTruthTable(manager.variables, false, manager.options.MaxWords)
	if err != nil {
		return DSDEdge{}, err
	}
	assignment := make([]bool, manager.variables)
	for value := uint64(0); value < uint64(1)<<uint(manager.variables); value++ {
		if err = manager.tick(1); err != nil {
			manager.operations = operations
			return DSDEdge{}, err
		}
		for variable := range assignment {
			assignment[variable] = value&(uint64(1)<<uint(variable)) != 0
		}
		result, evaluateErr := bdd.Evaluate(root, assignment)
		if evaluateErr != nil {
			manager.operations = operations
			return DSDEdge{}, evaluateErr
		}
		_ = function.Set(value, result)
	}
	edge, err := manager.DecomposeOne(function)
	if err != nil {
		manager.operations = operations
	}
	return edge, err
}

// Root and Input return checked explicit edges.
func (manager *DSDManager) Root(index int) (DSDEdge, error) {
	if err := manager.check(); err != nil {
		return DSDEdge{}, err
	}
	if index < 0 || index >= len(manager.roots) {
		return DSDEdge{}, ErrDimension
	}
	return manager.roots[index], nil
}

func (manager *DSDManager) Input(index int) (DSDEdge, error) {
	if err := manager.check(); err != nil {
		return DSDEdge{}, err
	}
	if index < 0 || index >= len(manager.inputs) {
		return DSDEdge{}, ErrDimension
	}
	return manager.inputs[index], nil
}

func (manager *DSDManager) ConstantOne() (DSDEdge, error) {
	if err := manager.check(); err != nil {
		return DSDEdge{}, err
	}
	return DSDEdge{Node: 0}, nil
}

// Node returns an owned snapshot.
func (manager *DSDManager) Node(index int) (DSDNode, error) {
	if err := manager.check(); err != nil {
		return DSDNode{}, err
	}
	if index < 0 || index >= len(manager.nodes) {
		return DSDNode{}, ErrDimension
	}
	return cloneDSDNode(manager.nodes[index]), nil
}

func (manager *DSDManager) SetMark(index, mark int) error {
	if err := manager.check(); err != nil {
		return err
	}
	if index < 0 || index >= len(manager.nodes) {
		return ErrDimension
	}
	manager.nodes[index].Mark = mark
	return nil
}

func (manager *DSDManager) Unmark() error {
	if err := manager.check(); err != nil {
		return err
	}
	for index := range manager.nodes {
		manager.nodes[index].Mark = 0
	}
	return nil
}

func (manager *DSDManager) evaluateEdge(edge DSDEdge, assignment uint64, memo map[int]bool) (bool, error) {
	if edge.Node < 0 || edge.Node >= len(manager.nodes) {
		return false, ErrInvalidLiteral
	}
	value, exists := memo[edge.Node]
	if !exists {
		node := manager.nodes[edge.Node]
		switch node.Type {
		case DSDNodeConst1:
			value = true
		case DSDNodeBuffer:
			value = assignment&(uint64(1)<<uint(node.Variable)) != 0
		case DSDNodeOR:
			for _, child := range node.Children {
				childValue, err := manager.evaluateEdge(child, assignment, memo)
				if err != nil {
					return false, err
				}
				value = value || childValue
			}
		case DSDNodeXOR:
			for _, child := range node.Children {
				childValue, err := manager.evaluateEdge(child, assignment, memo)
				if err != nil {
					return false, err
				}
				value = value != childValue
			}
		case DSDNodePrime:
			var local uint64
			for index, child := range node.Children {
				childValue, err := manager.evaluateEdge(child, assignment, memo)
				if err != nil {
					return false, err
				}
				if childValue {
					local |= uint64(1) << uint(index)
				}
			}
			value, _ = node.PrimeFunction.Get(local)
		default:
			return false, ErrInvariant
		}
		memo[edge.Node] = value
	}
	if edge.Complement {
		value = !value
	}
	return value, nil
}

// Reconstruct evaluates the decomposition into the original variable space.
func (manager *DSDManager) Reconstruct(edge DSDEdge) (*TruthTable, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	result, err := NewConstantTruthTable(manager.variables, false, manager.options.MaxWords)
	if err != nil {
		return nil, err
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(manager.variables); assignment++ {
		value, evaluateErr := manager.evaluateEdge(edge, assignment, make(map[int]bool))
		if evaluateErr != nil {
			return nil, evaluateErr
		}
		_ = result.Set(assignment, value)
	}
	return result, nil
}

// DSDTreeInfo summarizes source-shaped depth and gate-size metrics.
type DSDTreeInfo struct {
	Depth, MaximumGateSize int
}

func (manager *DSDManager) TreeInfo(edges []DSDEdge) (DSDTreeInfo, error) {
	if err := manager.check(); err != nil {
		return DSDTreeInfo{}, err
	}
	info := DSDTreeInfo{}
	var walk func(DSDEdge, int) error
	walk = func(edge DSDEdge, depth int) error {
		if edge.Node < 0 || edge.Node >= len(manager.nodes) {
			return ErrInvalidLiteral
		}
		node := manager.nodes[edge.Node]
		gateSize := len(node.Children)
		if node.Type == DSDNodeOR || node.Type == DSDNodeXOR {
			gateSize = 2
		}
		if gateSize > info.MaximumGateSize {
			info.MaximumGateSize = gateSize
		}
		if len(node.Children) >= 2 && depth+1 > info.Depth {
			info.Depth = depth + 1
		}
		for _, child := range node.Children {
			if err := walk(child, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	for _, edge := range edges {
		if err := walk(edge, 0); err != nil {
			return DSDTreeInfo{}, err
		}
	}
	return info, nil
}

// AIGCost preserves the source cost model for decomposed gates.
func (manager *DSDManager) AIGCost(edge DSDEdge) (int, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	visited := make(map[int]bool)
	var walk func(DSDEdge) (int, error)
	walk = func(current DSDEdge) (int, error) {
		if current.Node < 0 || current.Node >= len(manager.nodes) {
			return 0, ErrInvalidLiteral
		}
		if visited[current.Node] {
			return 0, nil
		}
		visited[current.Node] = true
		node := manager.nodes[current.Node]
		cost := 0
		switch node.Type {
		case DSDNodeOR:
			cost = len(node.Children) - 1
		case DSDNodeXOR:
			cost = 3 * (len(node.Children) - 1)
		case DSDNodePrime:
			if len(node.Children) == 3 {
				cost = 3
			}
		}
		for _, child := range node.Children {
			childCost, err := walk(child)
			if err != nil {
				return 0, err
			}
			cost += childCost
		}
		return cost, nil
	}
	return walk(edge)
}

// DFS returns unique nonterminal nodes in children-first source order.
func (manager *DSDManager) DFS(edge DSDEdge) ([]int, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	visited := make(map[int]bool)
	result := make([]int, 0)
	var walk func(DSDEdge) error
	walk = func(current DSDEdge) error {
		if current.Node < 0 || current.Node >= len(manager.nodes) {
			return ErrInvalidLiteral
		}
		if visited[current.Node] {
			return nil
		}
		visited[current.Node] = true
		node := manager.nodes[current.Node]
		if node.Type == DSDNodeBuffer || node.Type == DSDNodeConst1 {
			return nil
		}
		for _, child := range node.Children {
			if err := walk(child); err != nil {
				return err
			}
		}
		result = append(result, current.Node)
		return nil
	}
	if err := walk(edge); err != nil {
		return nil, err
	}
	return result, nil
}

func (manager *DSDManager) NodeCounts(edge DSDEdge) (nonterminal, prime int, err error) {
	nodes, err := manager.DFS(edge)
	if err != nil {
		return 0, 0, err
	}
	for _, index := range nodes {
		typeOf := manager.nodes[index].Type
		if typeOf != DSDNodeBuffer && typeOf != DSDNodeConst1 {
			nonterminal++
		}
		if typeOf == DSDNodePrime {
			prime++
		}
	}
	return nonterminal, prime, nil
}

// DecomposableVariables preserves the source's children-first collection.
// A prime node with more than four children blocks collection in its ancestors.
func (manager *DSDManager) DecomposableVariables(edge DSDEdge) ([]int, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	if edge.Node < 0 || edge.Node >= len(manager.nodes) {
		return nil, ErrInvalidLiteral
	}
	result := make([]int, 0, manager.variables)
	var walk func(int) (bool, error)
	walk = func(index int) (bool, error) {
		if index < 0 || index >= len(manager.nodes) {
			return false, ErrInvalidLiteral
		}
		node := manager.nodes[index]
		if len(node.Children) <= 1 {
			return false, nil
		}
		skip := false
		for _, child := range node.Children {
			childSkip, err := walk(child.Node)
			if err != nil {
				return false, err
			}
			skip = skip || childSkip
		}
		if skip || (node.Type != DSDNodeOR && node.Type != DSDNodeXOR && len(node.Children) > 4) {
			return true, nil
		}
		for _, child := range node.Children {
			childNode := manager.nodes[child.Node]
			if childNode.Type == DSDNodeBuffer {
				result = append(result, childNode.Variable)
			}
		}
		return false, nil
	}
	_, err := walk(edge.Node)
	return result, err
}

// PrimeFunction returns an owned local function for a prime node.
func (manager *DSDManager) PrimeFunction(index int) (*TruthTable, error) {
	node, err := manager.Node(index)
	if err != nil {
		return nil, err
	}
	if node.Type != DSDNodePrime {
		return nil, ErrDimension
	}
	return node.PrimeFunction, nil
}

// EquivalentOnCare compares two functions only where both care sets hold.
func DSDEquivalentOnCare(first, second, firstCare, secondCare *TruthTable) (bool, error) {
	if first == nil || second == nil || firstCare == nil || secondCare == nil || first.checkPair(second) != nil || first.checkPair(firstCare) != nil || first.checkPair(secondCare) != nil {
		return false, ErrDimension
	}
	assignments := uint64(1) << uint(first.Variables())
	for assignment := uint64(0); assignment < assignments; assignment++ {
		care1, _ := firstCare.Get(assignment)
		care2, _ := secondCare.Get(assignment)
		if !care1 || !care2 {
			continue
		}
		value1, _ := first.Get(assignment)
		value2, _ := second.Get(assignment)
		if value1 != value2 {
			return false, nil
		}
	}
	return true, nil
}

func (manager *DSDManager) formatEdge(edge DSDEdge, names []string, active map[int]bool) (string, error) {
	if edge.Node < 0 || edge.Node >= len(manager.nodes) || active[edge.Node] {
		return "", ErrInvariant
	}
	active[edge.Node] = true
	defer delete(active, edge.Node)
	node := manager.nodes[edge.Node]
	var result string
	switch node.Type {
	case DSDNodeConst1:
		result = "1"
	case DSDNodeBuffer:
		if node.Variable < len(names) && names[node.Variable] != "" {
			result = names[node.Variable]
		} else {
			result = fmt.Sprintf("x%d", node.Variable)
		}
	case DSDNodeOR, DSDNodeXOR, DSDNodePrime:
		parts := make([]string, len(node.Children))
		for index, child := range node.Children {
			text, err := manager.formatEdge(child, names, active)
			if err != nil {
				return "", err
			}
			parts[index] = text
		}
		if node.Type == DSDNodePrime {
			words := node.PrimeFunction.Words()
			bytes := make([]byte, 8*len(words))
			for index, word := range words {
				for shift := 0; shift < 8; shift++ {
					bytes[8*index+shift] = byte(word >> uint(8*shift))
				}
			}
			result = "P" + hex.EncodeToString(bytes) + "(" + strings.Join(parts, ",") + ")"
		} else {
			operator := " | "
			if node.Type == DSDNodeXOR {
				operator = " ^ "
			}
			result = "(" + strings.Join(parts, operator) + ")"
		}
	default:
		return "", ErrInvariant
	}
	if edge.Complement {
		result = "!" + result
	}
	return result, nil
}

// Format returns deterministic owned text instead of writing through FILE.
func (manager *DSDManager) Format(edge DSDEdge, names []string) (string, error) {
	if err := manager.check(); err != nil {
		return "", err
	}
	if len(names) != 0 && len(names) != manager.variables {
		return "", ErrDimension
	}
	return manager.formatEdge(edge, names, make(map[int]bool))
}

// Statistics summarizes manager-owned state.
type DSDStatistics struct {
	Variables, Nodes, Roots, Operations int
}

func (manager *DSDManager) Statistics() (DSDStatistics, error) {
	if err := manager.check(); err != nil {
		return DSDStatistics{}, err
	}
	return DSDStatistics{Variables: manager.variables, Nodes: len(manager.nodes), Roots: len(manager.roots), Operations: manager.operations}, nil
}

func (manager *DSDManager) Close() (DSDStatistics, error) {
	statistics, err := manager.Statistics()
	if err != nil {
		return DSDStatistics{}, err
	}
	manager.closed = true
	manager.nodes = nil
	manager.inputs = nil
	manager.roots = nil
	manager.cache = nil
	return statistics, nil
}
