package abc

// Canonical complemented-edge BDDs from src/misc/extra/extraUtilMult.c.

import (
	"math"
	"strconv"
	"strings"
)

type ExtraBDDLit uint32

const (
	ExtraBDDFalse ExtraBDDLit = 0
	ExtraBDDTrue  ExtraBDDLit = 1
)

func (literal ExtraBDDLit) Not() ExtraBDDLit { return literal ^ 1 }
func (literal ExtraBDDLit) ID() uint32       { return uint32(literal >> 1) }
func (literal ExtraBDDLit) Complemented() bool {
	return literal&1 != 0
}

type ExtraBDDOptions struct {
	MaxOperations, MaxBytes int
	Stop                    func() bool
}

type ExtraBDDStatistics struct {
	Variables, Nodes, Capacity int
	CacheLookups, CacheMisses  uint64
	MemoryBytes                uint64
}

type extraBDDNode struct {
	variable        uint8
	then, otherwise ExtraBDDLit
}

type extraBDDCacheEntry struct {
	first, second, result ExtraBDDLit
	valid                 bool
}

type ExtraBDDManager struct {
	variables, capacity int
	nodes               []extraBDDNode
	uniqueHeads         []uint32
	next                []uint32
	cache               []extraBDDCacheEntry
	cacheLookups        uint64
	cacheMisses         uint64
	closed              bool
}

func normalizeExtraBDDOptions(options ExtraBDDOptions) (ExtraBDDOptions, error) {
	if options.MaxOperations < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 26
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 16 << 20
	}
	return options, nil
}

func extraBDDTableSize(capacity int) (int, error) {
	if capacity <= 0 {
		return 0, ErrDimension
	}
	size := 1
	for size < capacity {
		if size > math.MaxInt/2 {
			return 0, ErrLimit
		}
		size <<= 1
	}
	return size, nil
}

func NewExtraBDDManager(variables, capacity int) (*ExtraBDDManager, error) {
	if variables < 0 || variables > 255 || capacity < variables+1 || capacity > int(math.MaxInt32>>1) {
		return nil, ErrDimension
	}
	tableSize, err := extraBDDTableSize(capacity)
	if err != nil {
		return nil, err
	}
	manager := &ExtraBDDManager{
		variables: variables, capacity: capacity,
		nodes:       make([]extraBDDNode, 1, capacity),
		next:        make([]uint32, 1, capacity),
		uniqueHeads: make([]uint32, tableSize),
		cache:       make([]extraBDDCacheEntry, tableSize),
	}
	manager.nodes[0].variable = 0xff
	for variable := 0; variable < variables; variable++ {
		literal, createErr := manager.uniqueCreate(variable, ExtraBDDTrue, ExtraBDDFalse, nil)
		if createErr != nil || literal != ExtraBDDLit(variable+1)<<1 {
			return nil, ErrInvariant
		}
	}
	return manager, nil
}

func (manager *ExtraBDDManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed || manager.variables < 0 || manager.variables > 255 || manager.capacity < manager.variables+1 || len(manager.nodes) == 0 || len(manager.nodes) != len(manager.next) || len(manager.nodes) > manager.capacity || len(manager.uniqueHeads) == 0 || len(manager.uniqueHeads)&(len(manager.uniqueHeads)-1) != 0 || len(manager.cache) != len(manager.uniqueHeads) || manager.nodes[0].variable != 0xff {
		return ErrInvariant
	}
	return nil
}

func (manager *ExtraBDDManager) valid(literal ExtraBDDLit) bool {
	return uint64(literal.ID()) < uint64(len(manager.nodes))
}

func (manager *ExtraBDDManager) checkLiteral(literal ExtraBDDLit) error {
	if err := manager.check(); err != nil {
		return err
	}
	if !manager.valid(literal) {
		return ErrInvalidLiteral
	}
	return nil
}

func extraBDDHash(first, second, third ExtraBDDLit) uint32 {
	return 12582917*uint32(first) + 4256249*uint32(second) + 741457*uint32(third)
}

func (manager *ExtraBDDManager) variableOf(literal ExtraBDDLit) int {
	return int(manager.nodes[literal.ID()].variable)
}

func (manager *ExtraBDDManager) thenOf(literal ExtraBDDLit) ExtraBDDLit {
	return manager.nodes[literal.ID()].then ^ ExtraBDDLit(uint32(literal)&1)
}

func (manager *ExtraBDDManager) elseOf(literal ExtraBDDLit) ExtraBDDLit {
	return manager.nodes[literal.ID()].otherwise ^ ExtraBDDLit(uint32(literal)&1)
}

type extraBDDUniqueLink struct {
	head  bool
	index int
	old   uint32
}

type extraBDDCacheChange struct {
	index int
	old   extraBDDCacheEntry
}

type extraBDDTransaction struct {
	nodeCount       int
	lookups, misses uint64
	uniqueLinks     []extraBDDUniqueLink
	cacheChanges    []extraBDDCacheChange
}

func (manager *ExtraBDDManager) beginTransaction() extraBDDTransaction {
	return extraBDDTransaction{nodeCount: len(manager.nodes), lookups: manager.cacheLookups, misses: manager.cacheMisses}
}

func (manager *ExtraBDDManager) rollback(transaction *extraBDDTransaction) {
	for index := len(transaction.cacheChanges) - 1; index >= 0; index-- {
		change := transaction.cacheChanges[index]
		manager.cache[change.index] = change.old
	}
	for index := len(transaction.uniqueLinks) - 1; index >= 0; index-- {
		change := transaction.uniqueLinks[index]
		if change.head {
			manager.uniqueHeads[change.index] = change.old
		} else {
			manager.next[change.index] = change.old
		}
	}
	manager.nodes = manager.nodes[:transaction.nodeCount]
	manager.next = manager.next[:transaction.nodeCount]
	manager.cacheLookups, manager.cacheMisses = transaction.lookups, transaction.misses
}

func (manager *ExtraBDDManager) uniqueCreateInternal(variable int, then, otherwise ExtraBDDLit, transaction *extraBDDTransaction) (ExtraBDDLit, error) {
	slot := int(extraBDDHash(ExtraBDDLit(variable), then, otherwise) & uint32(len(manager.uniqueHeads)-1))
	current := manager.uniqueHeads[slot]
	if current == 0 {
		if len(manager.nodes) >= manager.capacity {
			return 0, ErrLimit
		}
		id := uint32(len(manager.nodes))
		if transaction != nil {
			transaction.uniqueLinks = append(transaction.uniqueLinks, extraBDDUniqueLink{head: true, index: slot, old: 0})
		}
		manager.uniqueHeads[slot] = id
		manager.nodes = append(manager.nodes, extraBDDNode{variable: uint8(variable), then: then, otherwise: otherwise})
		manager.next = append(manager.next, 0)
		return ExtraBDDLit(id << 1), nil
	}
	for {
		node := manager.nodes[current]
		if int(node.variable) == variable && node.then == then && node.otherwise == otherwise {
			return ExtraBDDLit(current << 1), nil
		}
		if manager.next[current] == 0 {
			break
		}
		current = manager.next[current]
	}
	if len(manager.nodes) >= manager.capacity {
		return 0, ErrLimit
	}
	id := uint32(len(manager.nodes))
	if transaction != nil {
		transaction.uniqueLinks = append(transaction.uniqueLinks, extraBDDUniqueLink{index: int(current), old: 0})
	}
	manager.next[current] = id
	manager.nodes = append(manager.nodes, extraBDDNode{variable: uint8(variable), then: then, otherwise: otherwise})
	manager.next = append(manager.next, 0)
	return ExtraBDDLit(id << 1), nil
}

func (manager *ExtraBDDManager) uniqueCreate(variable int, then, otherwise ExtraBDDLit, transaction *extraBDDTransaction) (ExtraBDDLit, error) {
	if variable < 0 || variable >= manager.variables || !manager.valid(then) || !manager.valid(otherwise) || variable >= manager.variableOf(then) || variable >= manager.variableOf(otherwise) {
		return 0, ErrInvariant
	}
	if then == otherwise {
		return otherwise, nil
	}
	if !otherwise.Complemented() {
		return manager.uniqueCreateInternal(variable, then, otherwise, transaction)
	}
	result, err := manager.uniqueCreateInternal(variable, then.Not(), otherwise.Not(), transaction)
	if err != nil {
		return 0, err
	}
	return result.Not(), nil
}

func (manager *ExtraBDDManager) Variable(index int) (ExtraBDDLit, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if index < 0 || index >= manager.variables {
		return 0, ErrDimension
	}
	return ExtraBDDLit(index+1) << 1, nil
}

func (manager *ExtraBDDManager) cacheLookup(first, second ExtraBDDLit) (ExtraBDDLit, bool) {
	slot := int(extraBDDHash(0, first, second) & uint32(len(manager.cache)-1))
	manager.cacheLookups++
	entry := manager.cache[slot]
	return entry.result, entry.valid && entry.first == first && entry.second == second
}

func (manager *ExtraBDDManager) cacheInsert(first, second, result ExtraBDDLit, transaction *extraBDDTransaction) ExtraBDDLit {
	slot := int(extraBDDHash(0, first, second) & uint32(len(manager.cache)-1))
	transaction.cacheChanges = append(transaction.cacheChanges, extraBDDCacheChange{index: slot, old: manager.cache[slot]})
	manager.cache[slot] = extraBDDCacheEntry{first: first, second: second, result: result, valid: true}
	manager.cacheMisses++
	return result
}

type extraBDDWork struct {
	options     ExtraBDDOptions
	operations  int
	transaction *extraBDDTransaction
}

func (work *extraBDDWork) step() error {
	work.operations++
	if work.operations > work.options.MaxOperations {
		return ErrLimit
	}
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (manager *ExtraBDDManager) andRecursive(first, second ExtraBDDLit, work *extraBDDWork) (ExtraBDDLit, error) {
	if err := work.step(); err != nil {
		return 0, err
	}
	if first == ExtraBDDFalse || second == ExtraBDDFalse {
		return ExtraBDDFalse, nil
	}
	if first == ExtraBDDTrue {
		return second, nil
	}
	if second == ExtraBDDTrue || first == second {
		return first, nil
	}
	if first > second {
		return manager.andRecursive(second, first, work)
	}
	if cached, ok := manager.cacheLookup(first, second); ok {
		return cached, nil
	}
	firstVar, secondVar := manager.variableOf(first), manager.variableOf(second)
	var low, high ExtraBDDLit
	var err error
	switch {
	case firstVar < secondVar:
		low, err = manager.andRecursive(manager.elseOf(first), second, work)
		if err == nil {
			high, err = manager.andRecursive(manager.thenOf(first), second, work)
		}
	case firstVar > secondVar:
		low, err = manager.andRecursive(first, manager.elseOf(second), work)
		if err == nil {
			high, err = manager.andRecursive(first, manager.thenOf(second), work)
		}
	default:
		low, err = manager.andRecursive(manager.elseOf(first), manager.elseOf(second), work)
		if err == nil {
			high, err = manager.andRecursive(manager.thenOf(first), manager.thenOf(second), work)
		}
	}
	if err != nil {
		return 0, err
	}
	result, err := manager.uniqueCreate(min(firstVar, secondVar), high, low, work.transaction)
	if err != nil {
		return 0, err
	}
	return manager.cacheInsert(first, second, result, work.transaction), nil
}

func (manager *ExtraBDDManager) And(first, second ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	if err := manager.checkLiteral(first); err != nil {
		return 0, err
	}
	if err := manager.checkLiteral(second); err != nil {
		return 0, err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	transaction := manager.beginTransaction()
	work := extraBDDWork{options: options, transaction: &transaction}
	result, err := manager.andRecursive(first, second, &work)
	if err != nil {
		manager.rollback(&transaction)
		return 0, err
	}
	return result, nil
}

func (manager *ExtraBDDManager) Or(first, second ExtraBDDLit, options ExtraBDDOptions) (ExtraBDDLit, error) {
	result, err := manager.And(first.Not(), second.Not(), options)
	if err != nil {
		return 0, err
	}
	return result.Not(), nil
}

func (manager *ExtraBDDManager) Evaluate(root ExtraBDDLit, assignment []bool) (bool, error) {
	if err := manager.checkLiteral(root); err != nil {
		return false, err
	}
	if len(assignment) != manager.variables {
		return false, ErrDimension
	}
	current := root
	for current >= 2 {
		variable := manager.variableOf(current)
		if assignment[variable] {
			current = manager.thenOf(current)
		} else {
			current = manager.elseOf(current)
		}
	}
	return current == ExtraBDDTrue, nil
}

func (manager *ExtraBDDManager) countRoots(roots []ExtraBDDLit, shared bool, options ExtraBDDOptions) (int, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return 0, err
	}
	seen := make([]bool, len(manager.nodes))
	total, operations := 0, 0
	for _, root := range roots {
		if !manager.valid(root) {
			return 0, ErrInvalidLiteral
		}
		if !shared {
			clear(seen)
		}
		stack := []ExtraBDDLit{root}
		for len(stack) > 0 {
			operations++
			if operations > options.MaxOperations {
				return 0, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return 0, ErrCancelled
			}
			literal := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if literal < 2 || seen[literal.ID()] {
				continue
			}
			seen[literal.ID()] = true
			total++
			stack = append(stack, manager.elseOf(literal), manager.thenOf(literal))
		}
	}
	return total, nil
}

func (manager *ExtraBDDManager) CountNodes(root ExtraBDDLit, options ExtraBDDOptions) (int, error) {
	return manager.countRoots([]ExtraBDDLit{root}, true, options)
}

func (manager *ExtraBDDManager) CountSharedNodes(roots []ExtraBDDLit, options ExtraBDDOptions) (int, error) {
	return manager.countRoots(roots, true, options)
}

func (manager *ExtraBDDManager) CountSeparateNodes(roots []ExtraBDDLit, options ExtraBDDOptions) (int, error) {
	return manager.countRoots(roots, false, options)
}

func extraBDDWrite(builder *strings.Builder, limit int, parts ...string) error {
	length := 0
	for _, part := range parts {
		if len(part) > limit-builder.Len()-length {
			return ErrLimit
		}
		length += len(part)
	}
	for _, part := range parts {
		builder.WriteString(part)
	}
	return nil
}

func (manager *ExtraBDDManager) Format(root ExtraBDDLit, options ExtraBDDOptions) (string, error) {
	if err := manager.checkLiteral(root); err != nil {
		return "", err
	}
	options, err := normalizeExtraBDDOptions(options)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err = extraBDDWrite(&builder, options.MaxBytes, "BDD ", strconv.FormatUint(uint64(root), 10), " = "); err != nil {
		return "", err
	}
	path := make([]int8, manager.variables)
	for index := range path {
		path[index] = -1
	}
	operations := 0
	var visit func(ExtraBDDLit) error
	visit = func(literal ExtraBDDLit) error {
		operations++
		if operations > options.MaxOperations {
			return ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if literal == ExtraBDDFalse {
			return nil
		}
		if literal == ExtraBDDTrue {
			for variable, value := range path {
				if value < 0 {
					continue
				}
				sign := "-"
				if value == 1 {
					sign = "+"
				}
				if err := extraBDDWrite(&builder, options.MaxBytes, sign, strconv.Itoa(variable)); err != nil {
					return err
				}
			}
			return extraBDDWrite(&builder, options.MaxBytes, " ")
		}
		variable := manager.variableOf(literal)
		path[variable] = 0
		if err := visit(manager.elseOf(literal)); err != nil {
			return err
		}
		path[variable] = 1
		if err := visit(manager.thenOf(literal)); err != nil {
			return err
		}
		path[variable] = -1
		return nil
	}
	if err = visit(root); err != nil {
		return "", err
	}
	if err = extraBDDWrite(&builder, options.MaxBytes, "\n"); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func (manager *ExtraBDDManager) Statistics() (ExtraBDDStatistics, error) {
	if err := manager.check(); err != nil {
		return ExtraBDDStatistics{}, err
	}
	memory := uint64(len(manager.nodes))*uint64(1+4+4) + uint64(len(manager.next)+len(manager.uniqueHeads))*4 + uint64(len(manager.cache))*16
	return ExtraBDDStatistics{Variables: manager.variables, Nodes: len(manager.nodes), Capacity: manager.capacity, CacheLookups: manager.cacheLookups, CacheMisses: manager.cacheMisses, MemoryBytes: memory}, nil
}

func (statistics ExtraBDDStatistics) Format() string {
	hits := uint64(0)
	if statistics.CacheLookups >= statistics.CacheMisses {
		hits = statistics.CacheLookups - statistics.CacheMisses
	}
	megabytes := float64(statistics.MemoryBytes) / float64(1<<20)
	return "BDD stats: Var = " + strconv.Itoa(statistics.Variables) + "  Obj = " + strconv.Itoa(statistics.Nodes) + "  Alloc = " + strconv.Itoa(statistics.Capacity) + "  Hit = " + strconv.FormatUint(hits, 10) + "  Miss = " + strconv.FormatUint(statistics.CacheMisses, 10) + "  Mem = " + strconv.FormatFloat(megabytes, 'f', 2, 64) + " MB\n"
}

func (manager *ExtraBDDManager) Close() (ExtraBDDStatistics, error) {
	statistics, err := manager.Statistics()
	if err != nil {
		return ExtraBDDStatistics{}, err
	}
	manager.nodes, manager.next, manager.uniqueHeads, manager.cache = nil, nil, nil, nil
	manager.closed = true
	return statistics, nil
}

type ExtraBDDExampleResult struct {
	Root  ExtraBDDLit
	Text  string
	Nodes int
}

func (manager *ExtraBDDManager) Example(options ExtraBDDOptions) (ExtraBDDExampleResult, error) {
	if manager.variables < 4 {
		return ExtraBDDExampleResult{}, ErrDimension
	}
	a, _ := manager.Variable(0)
	b, _ := manager.Variable(1)
	c, _ := manager.Variable(2)
	d, _ := manager.Variable(3)
	ab, err := manager.And(a, b, options)
	if err != nil {
		return ExtraBDDExampleResult{}, err
	}
	cd, err := manager.And(c, d, options)
	if err != nil {
		return ExtraBDDExampleResult{}, err
	}
	root, err := manager.Or(ab, cd, options)
	if err != nil {
		return ExtraBDDExampleResult{}, err
	}
	text, err := manager.Format(root, options)
	if err != nil {
		return ExtraBDDExampleResult{}, err
	}
	nodes, err := manager.CountNodes(root, options)
	if err != nil {
		return ExtraBDDExampleResult{}, err
	}
	return ExtraBDDExampleResult{Root: root, Text: text, Nodes: nodes}, nil
}
