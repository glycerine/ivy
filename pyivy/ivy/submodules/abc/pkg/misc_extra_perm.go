package abc

// Permutation ZDD computation from src/misc/extra/extraUtilPerm.c.

import (
	"fmt"
	"math/big"
	"slices"
	"sort"
	"strings"
)

type ExtraZDDOperation uint8

const (
	extraZDDDifference ExtraZDDOperation = iota + 1
	extraZDDUnion
	extraZDDMinimalUnion
	extraZDDIntersection
	extraZDDPermutation
	extraZDDPermutationProduct
	extraZDDCofactor0
	extraZDDCofactor1
	extraZDDThreshold
	extraZDDDotProduct
	extraZDDDotMinimalProduct6
)

type ExtraZDDNode struct {
	Variable, True, False int
}

type extraZDDNodeKey struct{ variable, whenTrue, whenFalse int }
type extraZDDCacheKey struct {
	first, second int
	operation     ExtraZDDOperation
}
type extraZDDCacheEntry struct {
	key    extraZDDCacheKey
	result int
	valid  bool
}

type ExtraPermutationZDDOptions struct {
	MaxNodes, CacheEntries, MaxOperations int
	Stop                                  func() bool
}

type ExtraPermutationZDDStats struct {
	Variables, Nodes, NodeLimit, CacheEntries int
	CacheLookups, CacheHits, CacheMisses      int
	Operations                                int
}

type ExtraPermutationZDD struct {
	variables                                   int
	nodes                                       []ExtraZDDNode
	unique                                      map[extraZDDNodeKey]int
	cache                                       []extraZDDCacheEntry
	options                                     ExtraPermutationZDDOptions
	cacheLookups, cacheMisses, operations       int
	permutationSize                             int
	variableFirst, variableSecond, pairVariable []int
}

func normalizeExtraPermutationZDDOptions(variables int, options ExtraPermutationZDDOptions) (ExtraPermutationZDDOptions, error) {
	if variables < 0 || options.MaxNodes < 0 || options.CacheEntries < 0 || options.MaxOperations < 0 {
		return options, ErrDimension
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.CacheEntries == 0 {
		options.CacheEntries = options.MaxNodes
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 64 << 20
	}
	if options.MaxNodes < variables+2 || options.CacheEntries < 1 || options.MaxOperations < 1 {
		return options, ErrLimit
	}
	return options, nil
}

func NewExtraPermutationZDD(variables int, options ExtraPermutationZDDOptions) (*ExtraPermutationZDD, error) {
	options, err := normalizeExtraPermutationZDDOptions(variables, options)
	if err != nil {
		return nil, err
	}
	cacheSize := 1
	for cacheSize < options.CacheEntries {
		if cacheSize > 1<<29 {
			return nil, ErrLimit
		}
		cacheSize <<= 1
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	m := &ExtraPermutationZDD{variables: variables, options: options, unique: make(map[extraZDDNodeKey]int, variables), cache: make([]extraZDDCacheEntry, cacheSize)}
	m.nodes = append(m.nodes, ExtraZDDNode{Variable: variables}, ExtraZDDNode{Variable: variables})
	for variable := 0; variable < variables; variable++ {
		if _, err = m.create(variable, 1, 0); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *ExtraPermutationZDD) clone() *ExtraPermutationZDD {
	copyManager := *m
	copyManager.nodes = slices.Clone(m.nodes)
	copyManager.cache = slices.Clone(m.cache)
	copyManager.unique = make(map[extraZDDNodeKey]int, len(m.unique))
	for key, value := range m.unique {
		copyManager.unique[key] = value
	}
	copyManager.variableFirst = slices.Clone(m.variableFirst)
	copyManager.variableSecond = slices.Clone(m.variableSecond)
	copyManager.pairVariable = slices.Clone(m.pairVariable)
	return &copyManager
}

func (m *ExtraPermutationZDD) usable() bool {
	return m != nil && len(m.nodes) >= 2 && len(m.cache) != 0
}
func (m *ExtraPermutationZDD) validID(id int) bool { return m.usable() && id >= 0 && id < len(m.nodes) }
func (m *ExtraPermutationZDD) tick() error {
	if !m.usable() {
		return ErrDimension
	}
	if m.options.Stop != nil && m.options.Stop() {
		return ErrCancelled
	}
	if m.operations == m.options.MaxOperations {
		return ErrLimit
	}
	m.operations++
	return nil
}

func extraZDDHash(first, second, third int) uint32 {
	return uint32(12582917)*uint32(first) + uint32(4256249)*uint32(second) + uint32(741457)*uint32(third)
}

func (m *ExtraPermutationZDD) cacheLookup(first, second int, operation ExtraZDDOperation) (int, bool) {
	m.cacheLookups++
	key := extraZDDCacheKey{first, second, operation}
	entry := m.cache[int(extraZDDHash(first, second, int(operation)))&(len(m.cache)-1)]
	return entry.result, entry.valid && entry.key == key
}

func (m *ExtraPermutationZDD) cacheInsert(first, second int, operation ExtraZDDOperation, result int) int {
	key := extraZDDCacheKey{first, second, operation}
	m.cache[int(extraZDDHash(first, second, int(operation)))&(len(m.cache)-1)] = extraZDDCacheEntry{key: key, result: result, valid: true}
	m.cacheMisses++
	return result
}

func (m *ExtraPermutationZDD) create(variable, whenTrue, whenFalse int) (int, error) {
	if variable < 0 || variable >= m.variables || !m.validID(whenTrue) || !m.validID(whenFalse) {
		return 0, ErrDimension
	}
	if whenTrue == 0 {
		return whenFalse, nil
	}
	if variable >= m.nodes[whenTrue].Variable || variable >= m.nodes[whenFalse].Variable {
		return 0, ErrInvariant
	}
	key := extraZDDNodeKey{variable, whenTrue, whenFalse}
	if id, found := m.unique[key]; found {
		return id, nil
	}
	if len(m.nodes) == m.options.MaxNodes {
		return 0, ErrLimit
	}
	id := len(m.nodes)
	m.nodes = append(m.nodes, ExtraZDDNode{variable, whenTrue, whenFalse})
	m.unique[key] = id
	return id, nil
}

func (m *ExtraPermutationZDD) transactional(run func(*ExtraPermutationZDD) (int, error)) (int, error) {
	if !m.usable() {
		return 0, ErrDimension
	}
	trial := m.clone()
	result, err := run(trial)
	if err != nil {
		return 0, err
	}
	*m = *trial
	return result, nil
}

func (m *ExtraPermutationZDD) Variables() int {
	if !m.usable() {
		return 0
	}
	return m.variables
}
func (m *ExtraPermutationZDD) NodeCount() int {
	if !m.usable() {
		return 0
	}
	return len(m.nodes)
}
func (m *ExtraPermutationZDD) Variable(variable int) (int, error) {
	if !m.usable() || variable < 0 || variable >= m.variables {
		return 0, ErrDimension
	}
	return variable + 2, nil
}
func (m *ExtraPermutationZDD) Node(id int) (ExtraZDDNode, error) {
	if !m.validID(id) {
		return ExtraZDDNode{}, ErrDimension
	}
	return m.nodes[id], nil
}
func (m *ExtraPermutationZDD) Stats() ExtraPermutationZDDStats {
	if !m.usable() {
		return ExtraPermutationZDDStats{}
	}
	return ExtraPermutationZDDStats{m.variables, len(m.nodes), m.options.MaxNodes, len(m.cache), m.cacheLookups, m.cacheLookups - m.cacheMisses, m.cacheMisses, m.operations}
}

func (m *ExtraPermutationZDD) BuildSet(set []int) (int, error) {
	owned := slices.Clone(set)
	sort.Ints(owned)
	for index, variable := range owned {
		if variable < 0 || variable >= m.variables || index > 0 && owned[index-1] == variable {
			return 0, ErrDimension
		}
	}
	return m.transactional(func(trial *ExtraPermutationZDD) (int, error) {
		result := 1
		for index := len(owned) - 1; index >= 0; index-- {
			var err error
			result, err = trial.create(owned[index], result, 0)
			if err != nil {
				return 0, err
			}
		}
		return result, nil
	})
}

func (m *ExtraPermutationZDD) difference(a, b int) (int, error) {
	if err := m.tick(); err != nil {
		return 0, err
	}
	if a == 0 {
		return 0, nil
	}
	if b == 0 {
		return a, nil
	}
	if a == b {
		return 0, nil
	}
	if result, found := m.cacheLookup(a, b, extraZDDDifference); found {
		return result, nil
	}
	A, B := m.nodes[a], m.nodes[b]
	var result int
	var err error
	if A.Variable < B.Variable {
		var zero int
		zero, err = m.difference(A.False, b)
		if err == nil {
			result, err = m.create(A.Variable, A.True, zero)
		}
	} else if A.Variable > B.Variable {
		result, err = m.difference(a, B.False)
	} else {
		var zero, one int
		zero, err = m.difference(A.False, B.False)
		if err == nil {
			one, err = m.difference(A.True, B.True)
		}
		if err == nil {
			result, err = m.create(A.Variable, one, zero)
		}
	}
	if err != nil {
		return 0, err
	}
	return m.cacheInsert(a, b, extraZDDDifference, result), nil
}

func (m *ExtraPermutationZDD) union(a, b int, minimal bool) (int, error) {
	if err := m.tick(); err != nil {
		return 0, err
	}
	if a == 0 {
		return b, nil
	}
	if b == 0 {
		return a, nil
	}
	if a == b {
		return a, nil
	}
	if a > b {
		return m.union(b, a, minimal)
	}
	op := extraZDDUnion
	if minimal {
		op = extraZDDMinimalUnion
	}
	if r, found := m.cacheLookup(a, b, op); found {
		return r, nil
	}
	A, B := m.nodes[a], m.nodes[b]
	var zero, one int
	var err error
	if A.Variable < B.Variable {
		zero, err = m.union(A.False, b, minimal)
		one = A.True
	} else if A.Variable > B.Variable {
		zero, err = m.union(a, B.False, minimal)
		one = B.True
	} else {
		zero, err = m.union(A.False, B.False, minimal)
		if err == nil {
			one, err = m.union(A.True, B.True, minimal)
		}
	}
	if err == nil && minimal {
		one, err = m.difference(one, zero)
	}
	if err != nil {
		return 0, err
	}
	r, err := m.create(min(A.Variable, B.Variable), one, zero)
	if err != nil {
		return 0, err
	}
	return m.cacheInsert(a, b, op, r), nil
}

func (m *ExtraPermutationZDD) intersection(a, b int) (int, error) {
	if err := m.tick(); err != nil {
		return 0, err
	}
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a == b {
		return a, nil
	}
	if a > b {
		return m.intersection(b, a)
	}
	if r, f := m.cacheLookup(a, b, extraZDDIntersection); f {
		return r, nil
	}
	A, B := m.nodes[a], m.nodes[b]
	var zero, one int
	var err error
	if A.Variable < B.Variable {
		zero, err = m.intersection(A.False, b)
		one = A.True
	} else if A.Variable > B.Variable {
		zero, err = m.intersection(a, B.False)
		one = B.True
	} else {
		zero, err = m.intersection(A.False, B.False)
		if err == nil {
			one, err = m.intersection(A.True, B.True)
		}
	}
	if err != nil {
		return 0, err
	}
	r, err := m.create(min(A.Variable, B.Variable), one, zero)
	if err != nil {
		return 0, err
	}
	return m.cacheInsert(a, b, extraZDDIntersection, r), nil
}

func (m *ExtraPermutationZDD) cofactor(a, variable int, one bool) (int, error) {
	if err := m.tick(); err != nil {
		return 0, err
	}
	if a < 2 {
		return a, nil
	}
	A := m.nodes[a]
	if A.Variable > variable {
		return a, nil
	}
	op := extraZDDCofactor0
	if one {
		op = extraZDDCofactor1
	}
	if r, f := m.cacheLookup(a, variable, op); f {
		return r, nil
	}
	var zero, high, result int
	var err error
	if A.Variable < variable {
		zero, err = m.cofactor(A.False, variable, one)
		if err == nil {
			high, err = m.cofactor(A.True, variable, one)
		}
		if err == nil {
			result, err = m.create(A.Variable, high, zero)
		}
	} else if one {
		high, err = m.cofactor(A.True, variable, true)
		if err == nil {
			result, err = m.create(A.Variable, high, 0)
		}
	} else {
		result, err = m.cofactor(A.False, variable, false)
	}
	if err != nil {
		return 0, err
	}
	return m.cacheInsert(a, variable, op, result), nil
}

func (m *ExtraPermutationZDD) threshold(a, bound int) (int, error) {
	if err := m.tick(); err != nil {
		return 0, err
	}
	if a < 2 {
		return a, nil
	}
	if bound == 0 {
		return 0, nil
	}
	if r, f := m.cacheLookup(a, bound, extraZDDThreshold); f {
		return r, nil
	}
	A := m.nodes[a]
	zero, err := m.threshold(A.False, bound)
	if err != nil {
		return 0, err
	}
	one, err := m.threshold(A.True, bound-1)
	if err != nil {
		return 0, err
	}
	r, err := m.create(A.Variable, one, zero)
	if err != nil {
		return 0, err
	}
	return m.cacheInsert(a, bound, extraZDDThreshold, r), nil
}

func (m *ExtraPermutationZDD) dot(a, b int, minimal6 bool) (int, error) {
	if err := m.tick(); err != nil {
		return 0, err
	}
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a == 1 {
		return b, nil
	}
	if b == 1 {
		return a, nil
	}
	if a > b {
		return m.dot(b, a, minimal6)
	}
	op := extraZDDDotProduct
	if minimal6 {
		op = extraZDDDotMinimalProduct6
	}
	if r, f := m.cacheLookup(a, b, op); f {
		return r, nil
	}
	A, B := m.nodes[a], m.nodes[b]
	var zero, one int
	var err error
	if A.Variable < B.Variable {
		zero, err = m.dot(A.False, b, minimal6)
		if err == nil {
			one, err = m.dot(A.True, b, minimal6)
		}
	} else if A.Variable > B.Variable {
		zero, err = m.dot(a, B.False, minimal6)
		if err == nil {
			one, err = m.dot(a, B.True, minimal6)
		}
	} else {
		zero, err = m.dot(A.False, B.False, minimal6)
		var both, t1, t2 int
		if err == nil {
			both, err = m.union(B.False, B.True, minimal6)
		}
		if err == nil {
			t1, err = m.dot(A.True, both, minimal6)
		}
		if err == nil {
			t2, err = m.dot(A.False, B.True, minimal6)
		}
		if err == nil {
			one, err = m.union(t1, t2, minimal6)
		}
	}
	if err == nil && minimal6 {
		one, err = m.threshold(one, 5)
	}
	if err == nil && minimal6 {
		one, err = m.difference(one, zero)
	}
	if err != nil {
		return 0, err
	}
	r, err := m.create(min(A.Variable, B.Variable), one, zero)
	if err != nil {
		return 0, err
	}
	return m.cacheInsert(a, b, op, r), nil
}

func (m *ExtraPermutationZDD) binary(a, b int, run func(*ExtraPermutationZDD, int, int) (int, error)) (int, error) {
	if !m.validID(a) || !m.validID(b) {
		return 0, ErrDimension
	}
	return m.transactional(func(t *ExtraPermutationZDD) (int, error) { return run(t, a, b) })
}
func (m *ExtraPermutationZDD) Difference(a, b int) (int, error) {
	return m.binary(a, b, func(t *ExtraPermutationZDD, x, y int) (int, error) { return t.difference(x, y) })
}
func (m *ExtraPermutationZDD) Union(a, b int) (int, error) {
	return m.binary(a, b, func(t *ExtraPermutationZDD, x, y int) (int, error) { return t.union(x, y, false) })
}
func (m *ExtraPermutationZDD) MinimalUnion(a, b int) (int, error) {
	return m.binary(a, b, func(t *ExtraPermutationZDD, x, y int) (int, error) { return t.union(x, y, true) })
}
func (m *ExtraPermutationZDD) Intersection(a, b int) (int, error) {
	return m.binary(a, b, func(t *ExtraPermutationZDD, x, y int) (int, error) { return t.intersection(x, y) })
}
func (m *ExtraPermutationZDD) Cofactor0(a, variable int) (int, error) {
	if !m.validID(a) || variable < 0 || variable >= m.variables {
		return 0, ErrDimension
	}
	return m.transactional(func(t *ExtraPermutationZDD) (int, error) { return t.cofactor(a, variable, false) })
}
func (m *ExtraPermutationZDD) Cofactor1(a, variable int) (int, error) {
	if !m.validID(a) || variable < 0 || variable >= m.variables {
		return 0, ErrDimension
	}
	return m.transactional(func(t *ExtraPermutationZDD) (int, error) { return t.cofactor(a, variable, true) })
}
func (m *ExtraPermutationZDD) Threshold(a, bound int) (int, error) {
	if !m.validID(a) || bound < 0 {
		return 0, ErrDimension
	}
	return m.transactional(func(t *ExtraPermutationZDD) (int, error) { return t.threshold(a, bound) })
}
func (m *ExtraPermutationZDD) DotProduct(a, b int) (int, error) {
	return m.binary(a, b, func(t *ExtraPermutationZDD, x, y int) (int, error) { return t.dot(x, y, false) })
}
func (m *ExtraPermutationZDD) DotMinimalProduct6(a, b int) (int, error) {
	return m.binary(a, b, func(t *ExtraPermutationZDD, x, y int) (int, error) { return t.dot(x, y, true) })
}

func (m *ExtraPermutationZDD) CountPaths(root int) (*big.Int, error) {
	if !m.validID(root) {
		return nil, ErrDimension
	}
	memo := make([]*big.Int, len(m.nodes))
	var visit func(int) *big.Int
	visit = func(id int) *big.Int {
		if id == 0 {
			return new(big.Int)
		}
		if id == 1 {
			return big.NewInt(1)
		}
		if memo[id] != nil {
			return new(big.Int).Set(memo[id])
		}
		value := new(big.Int).Add(visit(m.nodes[id].False), visit(m.nodes[id].True))
		memo[id] = new(big.Int).Set(value)
		return value
	}
	return visit(root), nil
}
func (m *ExtraPermutationZDD) CountNodes(root int) (int, error) {
	return m.CountSharedNodes([]int{root})
}
func (m *ExtraPermutationZDD) CountSharedNodes(roots []int) (int, error) {
	if !m.usable() {
		return 0, ErrDimension
	}
	seen := make([]bool, len(m.nodes))
	var visit func(int) int
	visit = func(id int) int {
		if id < 2 || seen[id] {
			return 0
		}
		seen[id] = true
		return 1 + visit(m.nodes[id].False) + visit(m.nodes[id].True)
	}
	total := 0
	for _, root := range roots {
		if !m.validID(root) {
			return 0, ErrDimension
		}
		total += visit(root)
	}
	return total, nil
}

func (m *ExtraPermutationZDD) Subsets(root, maxSubsets, maxElements int) ([][]int, error) {
	if !m.validID(root) || maxSubsets < 0 || maxElements < 0 {
		return nil, ErrDimension
	}
	if maxSubsets == 0 {
		maxSubsets = 1 << 20
	}
	if maxElements == 0 {
		maxElements = 16 << 20
	}
	result := make([][]int, 0)
	path := make([]int, 0)
	elements := 0
	var visit func(int) error
	visit = func(id int) error {
		if m.options.Stop != nil && m.options.Stop() {
			return ErrCancelled
		}
		if id == 0 {
			return nil
		}
		if id == 1 {
			if len(result) == maxSubsets || elements > maxElements-len(path) {
				return ErrLimit
			}
			result = append(result, slices.Clone(path))
			elements += len(path)
			return nil
		}
		node := m.nodes[id]
		if err := visit(node.False); err != nil {
			return err
		}
		path = append(path, node.Variable)
		err := visit(node.True)
		path = path[:len(path)-1]
		return err
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	return result, nil
}

type ExtraTransposition struct{ First, Second int }

func ExtraPermutationToTranspositions(permutation []int) ([]ExtraTransposition, error) {
	owned := slices.Clone(permutation)
	seen := make([]bool, len(owned))
	for _, value := range owned {
		if value < 0 || value >= len(owned) || seen[value] {
			return nil, ErrDimension
		}
		seen[value] = true
	}
	result := make([]ExtraTransposition, 0, len(owned))
	for i := range owned {
		if owned[i] == i {
			continue
		}
		j := i + 1
		for j < len(owned) && owned[j] != i {
			j++
		}
		if j == len(owned) {
			return nil, ErrInvariant
		}
		result = append(result, ExtraTransposition{i, j})
		owned[i], owned[j] = owned[j], owned[i]
	}
	return result, nil
}
func ExtraTranspositionsToPermutation(transpositions []ExtraTransposition, size int) ([]int, error) {
	if size < 0 {
		return nil, ErrDimension
	}
	result := make([]int, size)
	for i := range result {
		result[i] = i
	}
	for index := len(transpositions) - 1; index >= 0; index-- {
		pair := transpositions[index]
		if pair.First < 0 || pair.First >= size || pair.Second < 0 || pair.Second >= size || pair.First >= pair.Second {
			return nil, ErrDimension
		}
		result[pair.First], result[pair.Second] = result[pair.Second], result[pair.First]
	}
	return result, nil
}
func FormatExtraPermutation(permutation []int) string {
	var b strings.Builder
	b.WriteByte('{')
	for _, value := range permutation {
		fmt.Fprintf(&b, " %2d", value)
	}
	b.WriteString(" }\n")
	return b.String()
}
func FormatExtraTranspositions(transpositions []ExtraTransposition) string {
	if len(transpositions) == 0 {
		return "Empty set\n"
	}
	var b strings.Builder
	for _, pair := range transpositions {
		fmt.Fprintf(&b, "(%d %d)", pair.First, pair.Second)
	}
	b.WriteByte('\n')
	return b.String()
}

func (m *ExtraPermutationZDD) ConfigurePermutations(size int) error {
	if !m.usable() || size < 1 || size*(size-1)/2 != m.variables || m.permutationSize != 0 {
		return ErrDimension
	}
	first := make([]int, m.variables)
	second := make([]int, m.variables)
	pairs := make([]int, size*size)
	for i := range pairs {
		pairs[i] = -1
	}
	variable := 0
	for i := 0; i < size; i++ {
		for j := i + 1; j < size; j++ {
			first[variable] = i
			second[variable] = j
			pairs[i*size+j] = variable
			variable++
		}
	}
	m.permutationSize = size
	m.variableFirst = first
	m.variableSecond = second
	m.pairVariable = pairs
	return nil
}
func (m *ExtraPermutationZDD) TranspositionVariable(first, second int) (int, error) {
	if !m.usable() || m.permutationSize == 0 || first < 0 || first >= second || second >= m.permutationSize {
		return 0, ErrDimension
	}
	return m.pairVariable[first*m.permutationSize+second], nil
}

func (m *ExtraPermutationZDD) permute(a, variable int) (int, error) {
	if err := m.tick(); err != nil {
		return 0, err
	}
	if a == 0 {
		return 0, nil
	}
	if a == 1 {
		return variable + 2, nil
	}
	if r, f := m.cacheLookup(a, variable, extraZDDPermutation); f {
		return r, nil
	}
	A := m.nodes[a]
	var result int
	var err error
	if m.variableFirst[A.Variable] > m.variableFirst[variable] {
		result, err = m.create(variable, a, 0)
	} else if A.Variable == variable {
		var zero int
		zero, err = m.permute(A.False, variable)
		if err == nil {
			result, err = m.union(zero, A.True, false)
		}
	} else {
		ai, aj := m.variableFirst[A.Variable], m.variableSecond[A.Variable]
		bi, bj := m.variableFirst[variable], m.variableSecond[variable]
		variablePerm, variableTop := 0, 0
		if aj == bi {
			variablePerm = variable
			variableTop = m.pairVariable[ai*m.permutationSize+bj]
		} else if aj == bj {
			variablePerm = variable
			variableTop = m.pairVariable[ai*m.permutationSize+bi]
		} else if ai == bi {
			x, y := aj, bj
			if x > y {
				x, y = y, x
			}
			variablePerm = m.pairVariable[x*m.permutationSize+y]
			variableTop = A.Variable
		} else {
			variablePerm = variable
			variableTop = A.Variable
		}
		zero, e := m.permute(A.False, variable)
		if e != nil {
			return 0, e
		}
		one, e := m.permute(A.True, variablePerm)
		if e != nil {
			return 0, e
		}
		if m.nodes[one].Variable <= variableTop {
			return 0, fmt.Errorf("%w: permutation child order", ErrInvariant)
		}
		if m.nodes[zero].Variable > variableTop {
			result, err = m.create(variableTop, one, zero)
			if err != nil {
				return 0, fmt.Errorf("%w: permute create top=%d one=%d/%d zero=%d/%d a=%d variable=%d", err, variableTop, one, m.nodes[one].Variable, zero, m.nodes[zero].Variable, a, variable)
			}
		} else {
			one, err = m.create(variableTop, one, 0)
			if err != nil {
				return 0, fmt.Errorf("%w: permute insert top=%d one=%d/%d a=%d variable=%d", err, variableTop, one, m.nodes[one].Variable, a, variable)
			}
			if err == nil {
				result, err = m.union(zero, one, false)
			}
		}
	}
	if err != nil {
		return 0, err
	}
	return m.cacheInsert(a, variable, extraZDDPermutation, result), nil
}
func (m *ExtraPermutationZDD) permutationProduct(a, b int) (int, error) {
	if err := m.tick(); err != nil {
		return 0, err
	}
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a == 1 {
		return b, nil
	}
	if b == 1 {
		return a, nil
	}
	if r, f := m.cacheLookup(a, b, extraZDDPermutationProduct); f {
		return r, nil
	}
	B := m.nodes[b]
	zero, err := m.permutationProduct(a, B.False)
	if err != nil {
		return 0, err
	}
	one, err := m.permutationProduct(a, B.True)
	if err != nil {
		return 0, err
	}
	one, err = m.permute(one, B.Variable)
	if err != nil {
		return 0, err
	}
	result, err := m.union(zero, one, false)
	if err != nil {
		return 0, err
	}
	return m.cacheInsert(a, b, extraZDDPermutationProduct, result), nil
}
func (m *ExtraPermutationZDD) Permute(root, variable int) (int, error) {
	if !m.validID(root) || m.permutationSize == 0 || variable < 0 || variable >= m.variables {
		return 0, ErrDimension
	}
	return m.transactional(func(t *ExtraPermutationZDD) (int, error) { return t.permute(root, variable) })
}
func (m *ExtraPermutationZDD) PermutationProduct(a, b int) (int, error) {
	if m.permutationSize == 0 {
		return 0, ErrDimension
	}
	return m.binary(a, b, func(t *ExtraPermutationZDD, x, y int) (int, error) { return t.permutationProduct(x, y) })
}

func (m *ExtraPermutationZDD) BuildPermutation(permutation []int) (int, error) {
	if len(permutation) != m.permutationSize {
		return 0, ErrDimension
	}
	transpositions, err := ExtraPermutationToTranspositions(permutation)
	if err != nil {
		return 0, err
	}
	variables := make([]int, len(transpositions))
	for index, pair := range transpositions {
		variables[index], err = m.TranspositionVariable(pair.First, pair.Second)
		if err != nil {
			return 0, err
		}
	}
	return m.BuildSet(variables)
}
func (m *ExtraPermutationZDD) Permutations(root, maxPermutations int) ([][]int, error) {
	if m.permutationSize == 0 {
		return nil, ErrDimension
	}
	subsets, err := m.Subsets(root, maxPermutations, m.options.MaxNodes*m.variables)
	if err != nil {
		return nil, err
	}
	result := make([][]int, len(subsets))
	for index, subset := range subsets {
		pairs := make([]ExtraTransposition, len(subset))
		for j, variable := range subset {
			pairs[j] = ExtraTransposition{m.variableFirst[variable], m.variableSecond[variable]}
		}
		result[index], err = ExtraTranspositionsToPermutation(pairs, m.permutationSize)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

type ExtraZDDGiaAnalysis struct {
	Roots       []int
	PathCount   *big.Int
	SharedNodes int
	Stats       ExtraPermutationZDDStats
}

// AnalyzeExtraZDDGia preserves Abc_ZddGiaTest's complement-insensitive fanin
// object analysis and returns its aggregate instead of printing it.
func AnalyzeExtraZDDGia(graph *GiaManager, options ExtraPermutationZDDOptions) (ExtraZDDGiaAnalysis, error) {
	fail := func(err error) (ExtraZDDGiaAnalysis, error) { return ExtraZDDGiaAnalysis{}, err }
	if graph == nil {
		return fail(ErrDimension)
	}
	if err := graph.Check(); err != nil {
		return fail(err)
	}
	manager, err := NewExtraPermutationZDD(graph.ObjectCount(), options)
	if err != nil {
		return fail(err)
	}
	values := make([]int, graph.ObjectCount())
	for _, id := range graph.cis {
		values[id] = int(id) + 2
	}
	roots := make([]int, 0, graph.AndCount())
	for id, object := range graph.objects {
		if !object.IsAnd() {
			continue
		}
		left, leftErr := object.Fanin0(GiaID(id))
		if leftErr != nil {
			return fail(leftErr)
		}
		right, rightErr := object.Fanin1(GiaID(id))
		if rightErr != nil {
			return fail(rightErr)
		}
		root, productErr := manager.DotMinimalProduct6(values[left.ID()], values[right.ID()])
		if productErr != nil {
			return fail(productErr)
		}
		variable, _ := manager.Variable(id)
		root, err = manager.Union(root, variable)
		if err != nil {
			return fail(err)
		}
		values[id] = root
		roots = append(roots, root)
	}
	total := new(big.Int)
	for _, root := range roots {
		count, countErr := manager.CountPaths(root)
		if countErr != nil {
			return fail(countErr)
		}
		total.Add(total, count)
	}
	shared, err := manager.CountSharedNodes(roots)
	if err != nil {
		return fail(err)
	}
	return ExtraZDDGiaAnalysis{Roots: roots, PathCount: total, SharedNodes: shared, Stats: manager.Stats()}, nil
}

type ExtraCubeZDDOptions struct {
	MaxDepth int
	ZDD      ExtraPermutationZDDOptions
}

type ExtraCubeZDDEnumeration struct {
	LayerTotals []int
	NodeTotals  []int
	UsedNodes   []int
	Converged   bool
	Stats       ExtraPermutationZDDStats
}

// EnumerateExtraCubeStatesZDD is the bounded, returned-result form of the
// source's 2x2x2 permutation-ZDD experiment.
func EnumerateExtraCubeStatesZDD(options ExtraCubeZDDOptions) (ExtraCubeZDDEnumeration, error) {
	fail := func(result ExtraCubeZDDEnumeration, err error) (ExtraCubeZDDEnumeration, error) { return result, err }
	if options.MaxDepth < 0 {
		return fail(ExtraCubeZDDEnumeration{}, ErrDimension)
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 5
	}
	if options.ZDD.MaxNodes == 0 {
		options.ZDD.MaxNodes = 1 << 22
	}
	manager, err := NewExtraPermutationZDD(24*23/2, options.ZDD)
	if err != nil {
		return fail(ExtraCubeZDDEnumeration{}, err)
	}
	if err = manager.ConfigurePermutations(24); err != nil {
		return fail(ExtraCubeZDDEnumeration{}, err)
	}
	turns := 1
	for _, permutation := range ExtraCubeGenerators() {
		values := make([]int, len(permutation))
		for index, value := range permutation {
			values[index] = int(value)
		}
		turn, buildErr := manager.BuildPermutation(values)
		if buildErr != nil {
			return fail(ExtraCubeZDDEnumeration{}, buildErr)
		}
		turns, err = manager.Union(turns, turn)
		if err != nil {
			return fail(ExtraCubeZDDEnumeration{}, err)
		}
	}
	result := ExtraCubeZDDEnumeration{LayerTotals: []int{1}, NodeTotals: []int{0}, UsedNodes: []int{2}}
	count, _ := manager.CountPaths(turns)
	nodes, _ := manager.CountNodes(turns)
	if !count.IsInt64() {
		return fail(result, ErrLimit)
	}
	result.LayerTotals = append(result.LayerTotals, int(count.Int64()))
	result.NodeTotals = append(result.NodeTotals, nodes)
	result.UsedNodes = append(result.UsedNodes, manager.NodeCount())
	all := turns
	for depth := 2; depth <= options.MaxDepth; depth++ {
		previous := all
		all, err = manager.PermutationProduct(all, turns)
		if err != nil {
			result.Stats = manager.Stats()
			return fail(result, err)
		}
		count, _ = manager.CountPaths(all)
		nodes, _ = manager.CountNodes(all)
		if !count.IsInt64() {
			result.Stats = manager.Stats()
			return fail(result, ErrLimit)
		}
		result.LayerTotals = append(result.LayerTotals, int(count.Int64()))
		result.NodeTotals = append(result.NodeTotals, nodes)
		result.UsedNodes = append(result.UsedNodes, manager.NodeCount())
		if all == previous {
			result.Converged = true
			break
		}
	}
	result.Stats = manager.Stats()
	if !result.Converged {
		return fail(result, ErrLimit)
	}
	return result, nil
}
