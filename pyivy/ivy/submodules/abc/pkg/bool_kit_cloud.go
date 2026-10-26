package abc

// Canonical complemented-edge BDD manager from src/bool/kit/cloud.c and cloud.h.

import (
	"math"
	"math/bits"
	"strconv"
	"strings"
)

// CloudEdge is an owned-manager BDD edge.  The low bit is the complement bit;
// the high word is a manager generation so Restart can reject stale edges.
type CloudEdge uint64

func (edge CloudEdge) Not() CloudEdge     { return edge ^ 1 }
func (edge CloudEdge) Complemented() bool { return edge&1 != 0 }
func (edge CloudEdge) nodeID() uint32     { return uint32(edge) >> 1 }
func (edge CloudEdge) generation() uint32 { return uint32(uint64(edge) >> 32) }
func cloudEdge(generation, id uint32) CloudEdge {
	return CloudEdge(uint64(generation)<<32 | uint64(id)<<1)
}

type CloudOperation uint8

const (
	CloudOperationAnd CloudOperation = iota
	CloudOperationXor
	CloudOperationBooleanDifference
	CloudOperationLessOrEqual
	cloudOperationCount
)

type CloudOptions struct {
	MaxOperations int
	Stop          func() bool
}

type CloudStatistics struct {
	Variables, Nodes, NodeLimit int
	Generation                  uint32
	UniqueHits, UniqueMisses    uint64
	CacheHits, CacheMisses      uint64
	MemoryBytes                 uint64
}

type cloudNode struct {
	variable        int
	then, otherwise CloudEdge
}

type cloudNodeKey struct {
	variable        int
	then, otherwise CloudEdge
}

type cloudCacheEntry struct {
	first, second, result CloudEdge
	valid                 bool
}

// CloudManager owns all nodes and caches.  maxNodes replaces the source's
// eager 2^(bits+1) allocation with an exact checked limit.
type CloudManager struct {
	variables, maxNodes int
	generation          uint32
	nodes               []cloudNode
	unique              map[cloudNodeKey]uint32
	caches              [cloudOperationCount][]cloudCacheEntry
	uniqueHits          uint64
	uniqueMisses        uint64
	cacheHits           uint64
	cacheMisses         uint64
	closed              bool
}

func NewCloudManager(variables, maxNodes int) (*CloudManager, error) {
	if variables < 0 || variables > 4096 || maxNodes < variables+1 || maxNodes > math.MaxInt32/2 {
		return nil, ErrDimension
	}
	manager := &CloudManager{variables: variables, maxNodes: maxNodes, generation: 1}
	if err := manager.initialize(); err != nil {
		return nil, err
	}
	return manager, nil
}

func (manager *CloudManager) initialize() error {
	manager.nodes = make([]cloudNode, 1, min(manager.maxNodes, manager.variables+32))
	manager.nodes[0].variable = manager.variables
	manager.unique = make(map[cloudNodeKey]uint32, min(manager.maxNodes, manager.variables+32))
	for operation := range manager.caches {
		if manager.caches[operation] != nil {
			clear(manager.caches[operation])
		}
	}
	for variable := 0; variable < manager.variables; variable++ {
		edge, err := manager.makeNodeRaw(variable, manager.One(), manager.Zero())
		if err != nil || edge.nodeID() != uint32(variable+1) {
			return ErrInvariant
		}
	}
	return nil
}

func (manager *CloudManager) check() error {
	if manager == nil {
		return ErrDimension
	}
	if manager.closed || manager.generation == 0 || manager.variables < 0 || manager.maxNodes < manager.variables+1 || len(manager.nodes) < manager.variables+1 || len(manager.nodes) > manager.maxNodes || manager.nodes[0].variable != manager.variables || manager.unique == nil {
		return ErrInvariant
	}
	return nil
}

func (manager *CloudManager) One() CloudEdge {
	if manager == nil {
		return 0
	}
	return cloudEdge(manager.generation, 0)
}

func (manager *CloudManager) Zero() CloudEdge { return manager.One().Not() }

func (manager *CloudManager) valid(edge CloudEdge) bool {
	return manager != nil && edge.generation() == manager.generation && uint64(edge.nodeID()) < uint64(len(manager.nodes))
}

func (manager *CloudManager) checkEdge(edge CloudEdge) error {
	if err := manager.check(); err != nil {
		return err
	}
	if !manager.valid(edge) {
		return ErrInvalidLiteral
	}
	return nil
}

func (manager *CloudManager) Variable(variable int) (CloudEdge, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	if variable < 0 || variable >= manager.variables {
		return 0, ErrDimension
	}
	return cloudEdge(manager.generation, uint32(variable+1)), nil
}

func (manager *CloudManager) variableOf(edge CloudEdge) int {
	return manager.nodes[edge.nodeID()].variable
}

func (manager *CloudManager) cofactors(edge CloudEdge) (CloudEdge, CloudEdge) {
	node := manager.nodes[edge.nodeID()]
	if edge.Complemented() {
		return node.then.Not(), node.otherwise.Not()
	}
	return node.then, node.otherwise
}

func (manager *CloudManager) makeNodeRaw(variable int, then, otherwise CloudEdge) (CloudEdge, error) {
	complement := then.Complemented()
	if complement {
		then, otherwise = then.Not(), otherwise.Not()
	}
	key := cloudNodeKey{variable: variable, then: then, otherwise: otherwise}
	if id, found := manager.unique[key]; found {
		manager.uniqueHits++
		edge := cloudEdge(manager.generation, id)
		if complement {
			edge = edge.Not()
		}
		return edge, nil
	}
	manager.uniqueMisses++
	if len(manager.nodes) >= manager.maxNodes {
		return 0, ErrLimit
	}
	id := uint32(len(manager.nodes))
	manager.nodes = append(manager.nodes, cloudNode{variable: variable, then: then, otherwise: otherwise})
	manager.unique[key] = id
	edge := cloudEdge(manager.generation, id)
	if complement {
		edge = edge.Not()
	}
	return edge, nil
}

func (manager *CloudManager) MakeNode(variable int, then, otherwise CloudEdge) (CloudEdge, error) {
	if err := manager.checkEdge(then); err != nil {
		return 0, err
	}
	if err := manager.checkEdge(otherwise); err != nil {
		return 0, err
	}
	if variable < 0 || variable >= manager.variables || variable >= manager.variableOf(then) || variable >= manager.variableOf(otherwise) {
		return 0, ErrDimension
	}
	if then == otherwise {
		return then, nil
	}
	transaction := manager.beginTransaction()
	result, err := manager.makeNodeRaw(variable, then, otherwise)
	if err != nil {
		manager.rollback(transaction)
		return 0, err
	}
	return result, nil
}

func normalizeCloudOptions(options CloudOptions) (CloudOptions, error) {
	if options.MaxOperations < 0 {
		return options, ErrDimension
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 1 << 26
	}
	return options, nil
}

type cloudWork struct {
	options    CloudOptions
	operations int
}

func (work *cloudWork) charge() error {
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	if work.operations >= work.options.MaxOperations {
		return ErrLimit
	}
	work.operations++
	return nil
}

type cloudTransaction struct {
	nodes                    int
	uniqueHits, uniqueMisses uint64
	cacheHits, cacheMisses   uint64
}

func (manager *CloudManager) beginTransaction() cloudTransaction {
	return cloudTransaction{len(manager.nodes), manager.uniqueHits, manager.uniqueMisses, manager.cacheHits, manager.cacheMisses}
}

func (manager *CloudManager) rollback(transaction cloudTransaction) {
	manager.nodes = manager.nodes[:transaction.nodes]
	manager.unique = make(map[cloudNodeKey]uint32, len(manager.nodes))
	for id := 1; id < len(manager.nodes); id++ {
		node := manager.nodes[id]
		manager.unique[cloudNodeKey{variable: node.variable, then: node.then, otherwise: node.otherwise}] = uint32(id)
	}
	for operation := range manager.caches {
		clear(manager.caches[operation])
	}
	manager.uniqueHits, manager.uniqueMisses = transaction.uniqueHits, transaction.uniqueMisses
	manager.cacheHits, manager.cacheMisses = transaction.cacheHits, transaction.cacheMisses
}

func cloudPowerOfTwoFloor(value int) int {
	if value < 2 {
		return 2
	}
	result := 1
	for result <= value/2 {
		result <<= 1
	}
	return result
}

// AllocateCache replaces the direct-mapped cache for one operation.  entries
// is rounded down to a power of two, matching the source's masked indexing.
func (manager *CloudManager) AllocateCache(operation CloudOperation, entries int) error {
	if err := manager.check(); err != nil {
		return err
	}
	if operation >= cloudOperationCount || entries < 2 || entries > 1<<24 {
		return ErrDimension
	}
	manager.caches[operation] = make([]cloudCacheEntry, cloudPowerOfTwoFloor(entries))
	return nil
}

// AllocateCacheRatio preserves Cloud_CacheAllocate's logarithmic sizing while
// rejecting ratios that would produce fewer than two entries.
func (manager *CloudManager) AllocateCacheRatio(operation CloudOperation, logRatio int) error {
	if err := manager.check(); err != nil {
		return err
	}
	bitsNode := bits.Len(uint(manager.maxNodes - 1))
	if operation >= cloudOperationCount || logRatio <= 0 || logRatio >= bitsNode {
		return ErrDimension
	}
	return manager.AllocateCache(operation, 1<<uint(bitsNode-logRatio))
}

func (manager *CloudManager) ensureCache(operation CloudOperation) {
	if manager.caches[operation] != nil {
		return
	}
	entries := manager.maxNodes / 4
	if entries > 1<<16 {
		entries = 1 << 16
	}
	manager.caches[operation] = make([]cloudCacheEntry, cloudPowerOfTwoFloor(entries))
}

func cloudCacheHash(first, second CloudEdge) uint64 {
	return uint64(first)*12582917 + uint64(second)*4256249
}

func (manager *CloudManager) andRecursive(first, second CloudEdge, work *cloudWork) (CloudEdge, error) {
	if err := work.charge(); err != nil {
		return 0, err
	}
	if first > second {
		first, second = second, first
	}
	if first == second {
		return first, nil
	}
	if first == second.Not() {
		return manager.Zero(), nil
	}
	if first == manager.Zero() || second == manager.Zero() {
		return manager.Zero(), nil
	}
	if first == manager.One() {
		return second, nil
	}
	if second == manager.One() {
		return first, nil
	}
	manager.ensureCache(CloudOperationAnd)
	cache := manager.caches[CloudOperationAnd]
	slot := int(cloudCacheHash(first, second) & uint64(len(cache)-1))
	manager.cacheHits++
	entry := cache[slot]
	if entry.valid && entry.first == first && entry.second == second {
		return entry.result, nil
	}
	manager.cacheMisses++
	manager.cacheHits--
	variable := min(manager.variableOf(first), manager.variableOf(second))
	firstThen, firstElse := first, first
	if manager.variableOf(first) == variable {
		firstThen, firstElse = manager.cofactors(first)
	}
	secondThen, secondElse := second, second
	if manager.variableOf(second) == variable {
		secondThen, secondElse = manager.cofactors(second)
	}
	then, err := manager.andRecursive(firstThen, secondThen, work)
	if err != nil {
		return 0, err
	}
	otherwise, err := manager.andRecursive(firstElse, secondElse, work)
	if err != nil {
		return 0, err
	}
	result := then
	if then != otherwise {
		result, err = manager.makeNodeRaw(variable, then, otherwise)
		if err != nil {
			return 0, err
		}
	}
	manager.caches[CloudOperationAnd][slot] = cloudCacheEntry{first: first, second: second, result: result, valid: true}
	return result, nil
}

func (manager *CloudManager) binary(first, second CloudEdge, options CloudOptions, operation CloudOperation) (CloudEdge, error) {
	if err := manager.checkEdge(first); err != nil {
		return 0, err
	}
	if err := manager.checkEdge(second); err != nil {
		return 0, err
	}
	options, err := normalizeCloudOptions(options)
	if err != nil {
		return 0, err
	}
	transaction := manager.beginTransaction()
	work := cloudWork{options: options}
	var result CloudEdge
	switch operation {
	case CloudOperationAnd:
		result, err = manager.andRecursive(first, second, &work)
	case CloudOperationXor:
		var left, right CloudEdge
		left, err = manager.andRecursive(first, second.Not(), &work)
		if err == nil {
			right, err = manager.andRecursive(first.Not(), second, &work)
		}
		if err == nil {
			result, err = manager.andRecursive(left.Not(), right.Not(), &work)
			result = result.Not()
		}
	default:
		err = ErrDimension
	}
	if err != nil {
		manager.rollback(transaction)
		return 0, err
	}
	return result, nil
}

func (manager *CloudManager) And(first, second CloudEdge, options CloudOptions) (CloudEdge, error) {
	return manager.binary(first, second, options, CloudOperationAnd)
}

func (manager *CloudManager) Or(first, second CloudEdge, options CloudOptions) (CloudEdge, error) {
	result, err := manager.binary(first.Not(), second.Not(), options, CloudOperationAnd)
	if err != nil {
		return 0, err
	}
	return result.Not(), nil
}

func (manager *CloudManager) Xor(first, second CloudEdge, options CloudOptions) (CloudEdge, error) {
	return manager.binary(first, second, options, CloudOperationXor)
}

func (manager *CloudManager) Evaluate(root CloudEdge, assignment []bool) (bool, error) {
	if err := manager.checkEdge(root); err != nil {
		return false, err
	}
	if len(assignment) < manager.variables {
		return false, ErrDimension
	}
	current := root
	for current.nodeID() != 0 {
		node := manager.nodes[current.nodeID()]
		then, otherwise := manager.cofactors(current)
		if assignment[node.variable] {
			current = then
		} else {
			current = otherwise
		}
	}
	return !current.Complemented(), nil
}

func (manager *CloudManager) Support(root CloudEdge, options CloudOptions) ([]int, error) {
	if err := manager.checkEdge(root); err != nil {
		return nil, err
	}
	options, err := normalizeCloudOptions(options)
	if err != nil {
		return nil, err
	}
	work := cloudWork{options: options}
	seen := make([]bool, len(manager.nodes))
	present := make([]bool, manager.variables)
	stack := []CloudEdge{root}
	for len(stack) != 0 {
		if err := work.charge(); err != nil {
			return nil, err
		}
		edge := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		id := edge.nodeID()
		if seen[id] {
			continue
		}
		seen[id] = true
		if id == 0 {
			continue
		}
		node := manager.nodes[id]
		present[node.variable] = true
		stack = append(stack, node.then, node.otherwise)
	}
	result := make([]int, 0)
	for variable, yes := range present {
		if yes {
			result = append(result, variable)
		}
	}
	return result, nil
}

func (manager *CloudManager) SupportSize(root CloudEdge, options CloudOptions) (int, error) {
	support, err := manager.Support(root, options)
	return len(support), err
}

func (manager *CloudManager) SupportCube(root CloudEdge, options CloudOptions) (CloudEdge, error) {
	support, err := manager.Support(root, options)
	if err != nil {
		return 0, err
	}
	transaction := manager.beginTransaction()
	result := manager.One()
	for index := len(support) - 1; index >= 0; index-- {
		result, err = manager.makeNodeRaw(support[index], result, manager.Zero())
		if err != nil {
			manager.rollback(transaction)
			return 0, err
		}
	}
	return result, nil
}

func (manager *CloudManager) DagCollect(root CloudEdge, options CloudOptions) ([]CloudEdge, error) {
	if err := manager.checkEdge(root); err != nil {
		return nil, err
	}
	options, err := normalizeCloudOptions(options)
	if err != nil {
		return nil, err
	}
	work := cloudWork{options: options}
	seen := make([]bool, len(manager.nodes))
	result := make([]CloudEdge, 0)
	var visit func(CloudEdge) error
	visit = func(edge CloudEdge) error {
		if err := work.charge(); err != nil {
			return err
		}
		id := edge.nodeID()
		if seen[id] {
			return nil
		}
		seen[id] = true
		if id != 0 {
			node := manager.nodes[id]
			if err := visit(node.then); err != nil {
				return err
			}
			if err := visit(node.otherwise); err != nil {
				return err
			}
		}
		result = append(result, cloudEdge(manager.generation, id))
		return nil
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	return result, nil
}

func (manager *CloudManager) DagSize(root CloudEdge, options CloudOptions) (int, error) {
	nodes, err := manager.DagCollect(root, options)
	return len(nodes), err
}

func (manager *CloudManager) SharingSize(roots []CloudEdge, options CloudOptions) (int, error) {
	if err := manager.check(); err != nil {
		return 0, err
	}
	options, err := normalizeCloudOptions(options)
	if err != nil {
		return 0, err
	}
	work := cloudWork{options: options}
	seen := make([]bool, len(manager.nodes))
	stack := make([]CloudEdge, len(roots))
	copy(stack, roots)
	for _, root := range roots {
		if !manager.valid(root) {
			return 0, ErrInvalidLiteral
		}
	}
	count := 0
	for len(stack) != 0 {
		if err := work.charge(); err != nil {
			return 0, err
		}
		edge := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		id := edge.nodeID()
		if seen[id] {
			continue
		}
		seen[id] = true
		count++
		if id != 0 {
			node := manager.nodes[id]
			stack = append(stack, node.then, node.otherwise)
		}
	}
	return count, nil
}

func (manager *CloudManager) GetOneCube(root CloudEdge, options CloudOptions) (CloudEdge, error) {
	if err := manager.checkEdge(root); err != nil {
		return 0, err
	}
	options, err := normalizeCloudOptions(options)
	if err != nil {
		return 0, err
	}
	if root == manager.Zero() || root == manager.One() {
		return root, nil
	}
	work := cloudWork{options: options}
	type literal struct {
		variable int
		positive bool
	}
	literals := make([]literal, 0)
	current := root
	for current.nodeID() != 0 {
		if err := work.charge(); err != nil {
			return 0, err
		}
		node := manager.nodes[current.nodeID()]
		then, otherwise := manager.cofactors(current)
		if otherwise != manager.Zero() {
			literals = append(literals, literal{variable: node.variable})
			current = otherwise
		} else {
			if then == manager.Zero() {
				return 0, ErrInvariant
			}
			literals = append(literals, literal{variable: node.variable, positive: true})
			current = then
		}
	}
	if current == manager.Zero() {
		return manager.Zero(), nil
	}
	transaction := manager.beginTransaction()
	result := manager.One()
	for index := len(literals) - 1; index >= 0; index-- {
		item := literals[index]
		if item.positive {
			result, err = manager.makeNodeRaw(item.variable, result, manager.Zero())
		} else {
			result, err = manager.makeNodeRaw(item.variable, manager.Zero(), result)
		}
		if err != nil {
			manager.rollback(transaction)
			return 0, err
		}
	}
	return result, nil
}

func (manager *CloudManager) FormatCube(cube CloudEdge) (string, error) {
	if err := manager.checkEdge(cube); err != nil {
		return "", err
	}
	if cube.nodeID() == 0 {
		return "", ErrDimension
	}
	var builder strings.Builder
	current := cube
	for current.nodeID() != 0 {
		node := manager.nodes[current.nodeID()]
		then, otherwise := manager.cofactors(current)
		if (then == manager.Zero()) == (otherwise == manager.Zero()) {
			return "", ErrDimension
		}
		builder.WriteByte('[')
		builder.WriteString(strconv.Itoa(node.variable))
		builder.WriteByte(']')
		if otherwise != manager.Zero() {
			builder.WriteByte('\'')
			current = otherwise
		} else {
			current = then
		}
	}
	if current != manager.One() {
		return "", ErrDimension
	}
	return builder.String(), nil
}

// FormatDNF returns the source print notation without writing process-global
// stdout.  Root-to-one paths are disjoint cubes.
func (manager *CloudManager) FormatDNF(root CloudEdge, options CloudOptions) (string, error) {
	if err := manager.checkEdge(root); err != nil {
		return "", err
	}
	if root == manager.Zero() {
		return "Constant 0.\n", nil
	}
	if root == manager.One() {
		return "Constant 1.\n", nil
	}
	options, err := normalizeCloudOptions(options)
	if err != nil {
		return "", err
	}
	work := cloudWork{options: options}
	type item struct {
		variable int
		positive bool
	}
	path := make([]item, 0)
	cubes := make([]string, 0)
	var visit func(CloudEdge) error
	visit = func(edge CloudEdge) error {
		if err := work.charge(); err != nil {
			return err
		}
		if edge == manager.Zero() {
			return nil
		}
		if edge == manager.One() {
			var builder strings.Builder
			for _, literal := range path {
				builder.WriteByte('[')
				builder.WriteString(strconv.Itoa(literal.variable))
				builder.WriteByte(']')
				if !literal.positive {
					builder.WriteByte('\'')
				}
			}
			cubes = append(cubes, builder.String())
			return nil
		}
		node := manager.nodes[edge.nodeID()]
		then, otherwise := manager.cofactors(edge)
		path = append(path, item{variable: node.variable})
		if err := visit(otherwise); err != nil {
			return err
		}
		path[len(path)-1].positive = true
		if err := visit(then); err != nil {
			return err
		}
		path = path[:len(path)-1]
		return nil
	}
	if err := visit(root); err != nil {
		return "", err
	}
	return strings.Join(cubes, " + ") + "\n", nil
}

func (manager *CloudManager) HashOccupancy() ([]bool, error) {
	if err := manager.check(); err != nil {
		return nil, err
	}
	result := make([]bool, manager.maxNodes)
	for id := 1; id < len(manager.nodes); id++ {
		result[id] = true
	}
	return result, nil
}

func (manager *CloudManager) Statistics() (CloudStatistics, error) {
	if err := manager.check(); err != nil {
		return CloudStatistics{}, err
	}
	cacheEntries := 0
	for _, cache := range manager.caches {
		cacheEntries += len(cache)
	}
	memory := uint64(len(manager.nodes))*32 + uint64(len(manager.unique))*40 + uint64(cacheEntries)*32
	return CloudStatistics{Variables: manager.variables, Nodes: len(manager.nodes), NodeLimit: manager.maxNodes, Generation: manager.generation, UniqueHits: manager.uniqueHits, UniqueMisses: manager.uniqueMisses, CacheHits: manager.cacheHits, CacheMisses: manager.cacheMisses, MemoryBytes: memory}, nil
}

func (statistics CloudStatistics) Format() string {
	return "variables=" + strconv.Itoa(statistics.Variables) + " nodes=" + strconv.Itoa(statistics.Nodes) + "/" + strconv.Itoa(statistics.NodeLimit) + " generation=" + strconv.FormatUint(uint64(statistics.Generation), 10) + " unique_hits=" + strconv.FormatUint(statistics.UniqueHits, 10) + " unique_misses=" + strconv.FormatUint(statistics.UniqueMisses, 10) + " cache_hits=" + strconv.FormatUint(statistics.CacheHits, 10) + " cache_misses=" + strconv.FormatUint(statistics.CacheMisses, 10) + " memory=" + strconv.FormatUint(statistics.MemoryBytes, 10)
}

func (manager *CloudManager) Restart() error {
	if err := manager.check(); err != nil {
		return err
	}
	if manager.generation == math.MaxUint32 {
		return ErrLimit
	}
	hits, misses := manager.uniqueHits, manager.uniqueMisses
	manager.generation++
	if err := manager.initialize(); err != nil {
		return err
	}
	manager.uniqueHits, manager.uniqueMisses = hits, misses
	return nil
}

func (manager *CloudManager) Close() (CloudStatistics, error) {
	statistics, err := manager.Statistics()
	if err != nil {
		return CloudStatistics{}, err
	}
	manager.closed = true
	manager.nodes = nil
	manager.unique = nil
	for operation := range manager.caches {
		manager.caches[operation] = nil
	}
	return statistics, nil
}
