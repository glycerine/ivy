package abc

// Owned structural DSD manager core from opt/dau/dauTree.c.

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"slices"
	"strings"

	"github.com/glycerine/insdict"
)

type DauTreeLiteral uint32

func (literal DauTreeLiteral) ID() int             { return int(literal >> 1) }
func (literal DauTreeLiteral) Complemented() bool  { return literal&1 != 0 }
func (literal DauTreeLiteral) Not() DauTreeLiteral { return literal ^ 1 }

type DauTreeNode struct {
	Kind    DauDSDNodeKind
	Support int
	Fanins  []DauTreeLiteral
	Truth   []uint64
}

type DauTreeMergeResult struct {
	Literal DauTreeLiteral
	Mapping []DauTreeLiteral
}

type DauTreeSharedVariable struct {
	SecondIndex int
	First       DauTreeLiteral
}

type DauTreeSharedMap struct {
	First, Second DauTreeLiteral
	Variables     []DauTreeSharedVariable
}

func dauTreePrimeDigits(words []uint64, variables int) string {
	bits := 1 << uint(variables)
	digits := make([]byte, max(1, bits/4))
	for digit := range digits {
		value := 0
		for bit := 0; bit < 4; bit++ {
			index := 4*digit + bit
			if words[index>>6]&(1<<uint(index&63)) != 0 {
				value |= 1 << uint(bit)
			}
		}
		if value < 10 {
			digits[len(digits)-1-digit] = byte('0' + value)
		} else {
			digits[len(digits)-1-digit] = byte('A' + value - 10)
		}
	}
	return string(digits)
}

// Format returns Dss_ManPrint_rec's formula payload without its numeric stdout
// prefix. The leaf mapping is consumed in structural traversal order.
func (manager *DauTreeManager) Format(literal DauTreeLiteral, permutation []DauTreeLiteral, maxBytes int) (string, error) {
	if maxBytes < 0 {
		return "", ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = DauMaxFormulaBytes
	}
	if err := manager.checkLiteral(literal); err != nil {
		return "", err
	}
	if len(permutation) != manager.nodes[literal.ID()].Support {
		return "", ErrDimension
	}
	for _, mapped := range permutation {
		if mapped.ID() < 0 || mapped.ID() >= 26 {
			return "", ErrDimension
		}
	}
	position := 0
	var render func(DauTreeLiteral) (string, error)
	render = func(current DauTreeLiteral) (string, error) {
		if manager.stop != nil && manager.stop() {
			return "", ErrCancelled
		}
		node := manager.nodes[current.ID()]
		prefix := ""
		if current.Complemented() {
			prefix = "!"
		}
		if node.Kind == DauDSDConstant {
			if current.Complemented() {
				return "1", nil
			}
			return "0", nil
		}
		if node.Kind == DauDSDVariable {
			mapped := permutation[position]
			position++
			if current.Complemented() != mapped.Complemented() {
				prefix = "!"
			} else {
				prefix = ""
			}
			return prefix + string(rune('a'+mapped.ID())), nil
		}
		open, close := byte(0), byte(0)
		switch node.Kind {
		case DauDSDAnd:
			open, close = '(', ')'
		case DauDSDXor:
			open, close = '[', ']'
		case DauDSDMux:
			open, close = '<', '>'
		case DauDSDPrime:
			open, close = '{', '}'
		default:
			return "", ErrInvariant
		}
		var body strings.Builder
		if node.Kind == DauDSDPrime {
			body.WriteString(dauTreePrimeDigits(node.Truth, len(node.Fanins)))
		}
		body.WriteByte(open)
		for _, child := range node.Fanins {
			text, err := render(child)
			if err != nil {
				return "", err
			}
			if len(prefix)+body.Len()+len(text)+1 > maxBytes {
				return "", ErrLimit
			}
			body.WriteString(text)
		}
		body.WriteByte(close)
		result := prefix + body.String()
		if len(result) > maxBytes {
			return "", ErrLimit
		}
		return result, nil
	}
	result, err := render(literal)
	if err != nil {
		return "", err
	}
	if position != len(permutation) {
		return "", ErrInvariant
	}
	return result, nil
}

func (manager *DauTreeManager) ContainsPrime(literal DauTreeLiteral) (bool, error) {
	if err := manager.checkLiteral(literal); err != nil {
		return false, err
	}
	seen := make([]bool, len(manager.nodes))
	var visit func(int) (bool, error)
	visit = func(id int) (bool, error) {
		if manager.stop != nil && manager.stop() {
			return false, ErrCancelled
		}
		if seen[id] {
			return false, nil
		}
		seen[id] = true
		node := manager.nodes[id]
		if node.Kind == DauDSDPrime {
			return true, nil
		}
		for _, child := range node.Fanins {
			found, err := visit(child.ID())
			if err != nil || found {
				return found, err
			}
		}
		return false, nil
	}
	return visit(literal.ID())
}

// SharedMap preserves Dss_ManSharedMap's mapping from second-function leaf
// positions into first-function local literals.
func (manager *DauTreeManager) SharedMap(first, second DauTreeLiteral, firstMapping, secondMapping []DauTreeLiteral, sharedMask uint32) (DauTreeSharedMap, error) {
	if err := manager.checkLiteral(first); err != nil {
		return DauTreeSharedMap{}, err
	}
	if err := manager.checkLiteral(second); err != nil {
		return DauTreeSharedMap{}, err
	}
	if len(firstMapping) != manager.nodes[first.ID()].Support || len(secondMapping) != manager.nodes[second.ID()].Support {
		return DauTreeSharedMap{}, ErrDimension
	}
	globalToLocal := make([]DauTreeLiteral, manager.variables)
	present := make([]bool, manager.variables)
	for local, global := range firstMapping {
		if global.ID() >= manager.variables || present[global.ID()] {
			return DauTreeSharedMap{}, ErrInvariant
		}
		present[global.ID()] = true
		globalToLocal[global.ID()] = DauTreeLiteral(local << 1)
		if global.Complemented() {
			globalToLocal[global.ID()] = globalToLocal[global.ID()].Not()
		}
	}
	result := DauTreeSharedMap{First: first, Second: second}
	for secondIndex, global := range secondMapping {
		if global.ID() >= manager.variables {
			return DauTreeSharedMap{}, ErrDimension
		}
		if sharedMask&(1<<uint(global.ID())) == 0 {
			continue
		}
		if !present[global.ID()] {
			return DauTreeSharedMap{}, ErrInvariant
		}
		local := globalToLocal[global.ID()]
		if global.Complemented() {
			local = local.Not()
		}
		result.Variables = append(result.Variables, DauTreeSharedVariable{SecondIndex: secondIndex, First: local})
	}
	return result, nil
}

func (manager *DauTreeManager) IsTransparent(literal DauTreeLiteral) (bool, error) {
	if err := manager.checkLiteral(literal); err != nil {
		return false, err
	}
	var visit func(int) (bool, error)
	visit = func(id int) (bool, error) {
		if manager.stop != nil && manager.stop() {
			return false, ErrCancelled
		}
		node := manager.nodes[id]
		switch node.Kind {
		case DauDSDConstant, DauDSDAnd, DauDSDPrime:
			return false, nil
		case DauDSDVariable:
			return true, nil
		case DauDSDXor:
			for _, fanin := range node.Fanins {
				transparent, err := visit(fanin.ID())
				if err != nil || transparent {
					return transparent, err
				}
			}
			return false, nil
		case DauDSDMux:
			if len(node.Fanins) != 3 {
				return false, ErrInvariant
			}
			first, err := visit(node.Fanins[1].ID())
			if err != nil || !first {
				return false, err
			}
			return visit(node.Fanins[2].ID())
		default:
			return false, ErrInvariant
		}
	}
	return visit(literal.ID())
}

// MergeAnd ports Dss_ManMerge's constant, disjoint-support and shared-support
// paths. Overlapping functions are evaluated in a phase-correct local support,
// generally decomposed, and atomically imported into the structural manager.
func (manager *DauTreeManager) MergeAnd(first, second DauTreeLiteral, firstMapping, secondMapping []DauTreeLiteral, sharedMask uint32, maxLeaves int) (DauTreeMergeResult, error) {
	if manager == nil || maxLeaves < 0 {
		return DauTreeMergeResult{}, ErrDimension
	}
	if err := manager.checkLiteral(first); err != nil {
		return DauTreeMergeResult{}, err
	}
	if err := manager.checkLiteral(second); err != nil {
		return DauTreeMergeResult{}, err
	}
	if first > second {
		return DauTreeMergeResult{}, ErrInvariant
	}
	if len(firstMapping) != manager.nodes[first.ID()].Support || len(secondMapping) != manager.nodes[second.ID()].Support {
		return DauTreeMergeResult{}, ErrDimension
	}
	allowedMask := uint32(0)
	if manager.variables != 0 {
		allowedMask = (uint32(1) << uint(manager.variables)) - 1
	}
	if sharedMask&^allowedMask != 0 {
		return DauTreeMergeResult{}, ErrDimension
	}
	firstMask, secondMask := uint32(0), uint32(0)
	for _, mapped := range firstMapping {
		if mapped.ID() >= manager.variables {
			return DauTreeMergeResult{}, ErrDimension
		}
		bit := uint32(1) << uint(mapped.ID())
		if firstMask&bit != 0 {
			return DauTreeMergeResult{}, ErrInvariant
		}
		firstMask |= bit
	}
	for _, mapped := range secondMapping {
		if mapped.ID() >= manager.variables {
			return DauTreeMergeResult{}, ErrDimension
		}
		bit := uint32(1) << uint(mapped.ID())
		if secondMask&bit != 0 {
			return DauTreeMergeResult{}, ErrInvariant
		}
		secondMask |= bit
	}
	if firstMask&secondMask != sharedMask {
		return DauTreeMergeResult{}, ErrInvariant
	}
	if first == 0 || second == 0 {
		return DauTreeMergeResult{Literal: 0}, nil
	}
	if first == 1 {
		return DauTreeMergeResult{Literal: second, Mapping: slices.Clone(secondMapping)}, nil
	}
	if second == 1 {
		return DauTreeMergeResult{Literal: first, Mapping: slices.Clone(firstMapping)}, nil
	}
	if len(firstMapping)+len(secondMapping) > maxLeaves+bits.OnesCount32(sharedMask) {
		return DauTreeMergeResult{}, ErrLimit
	}
	if sharedMask != 0 {
		shared, firstLocal, secondLocal, sources, err := manager.sharedMergeContext(first, second, firstMapping, secondMapping, sharedMask)
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		cacheKey := dauTreeCacheKey(shared)
		if manager.cache != nil {
			if cached, found := manager.cache.Get2(cacheKey); found {
				manager.cacheHits[1]++
				return dauTreeTranslateSources(cached, sources)
			}
			manager.cacheMisses[1]++
			if manager.cache.Len() >= manager.maxCacheEntries {
				return DauTreeMergeResult{}, ErrLimit
			}
		}
		firstTruth, err := manager.Truth(first, firstLocal, DauMaxTruthWords, manager.stop)
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		secondTruth, err := manager.Truth(second, secondLocal, DauMaxTruthWords, manager.stop)
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		keep := make([]int, len(sources))
		for index := range keep {
			keep[index] = index
		}
		firstTruth, err = firstTruth.ProjectVariables(keep)
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		secondTruth, err = secondTruth.ProjectVariables(keep)
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		conjunction, err := firstTruth.And(secondTruth, false, false)
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		decomposition, err := DauDSDDecompose(conjunction, DauDSDDecomposeOptions{
			MaxBytes: DauMaxFormulaBytes, MaxNodes: manager.maxNodes, MaxDepth: 128,
			MaxWords: DauMaxTruthWords, Stop: manager.stop,
		})
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		if manager.nonDecomposableLimit != 0 && decomposition.LargestPrime > manager.nonDecomposableLimit {
			return DauTreeMergeResult{}, ErrLimit
		}
		working := manager.clone()
		literal, localMapping, err := working.ImportFormula(decomposition.Formula, DauDSDOptions{
			MaxBytes: DauMaxFormulaBytes, MaxNodes: manager.maxNodes, MaxDepth: 128, MaxVariables: manager.variables, Stop: manager.stop,
		})
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		localResult := DauTreeMergeResult{Literal: literal, Mapping: localMapping}
		result, err := dauTreeTranslateSources(localResult, sources)
		if err != nil {
			return DauTreeMergeResult{}, err
		}
		if len(result.Mapping) > maxLeaves {
			return DauTreeMergeResult{}, ErrLimit
		}
		manager.nodes, manager.index = working.nodes, working.index
		if manager.cache != nil {
			manager.cache.Set(cacheKey, cloneDauTreeMergeResult(localResult))
			manager.cacheEntries[1]++
		}
		return result, nil
	}
	cacheKey := dauTreeCacheKey(DauTreeSharedMap{First: first, Second: second})
	if manager.cache != nil {
		if cached, found := manager.cache.Get2(cacheKey); found {
			manager.cacheHits[0]++
			return manager.translateMergeMapping(cached, firstMapping, secondMapping)
		}
		manager.cacheMisses[0]++
		if manager.cache.Len() >= manager.maxCacheEntries {
			return DauTreeMergeResult{}, ErrLimit
		}
	}
	working := manager.clone()
	literal, local, err := working.AndWithPermutation([]DauTreeLiteral{first, second})
	if err != nil {
		return DauTreeMergeResult{}, err
	}
	localResult := DauTreeMergeResult{Literal: literal, Mapping: local}
	result, err := manager.translateMergeMapping(localResult, firstMapping, secondMapping)
	if err != nil {
		return DauTreeMergeResult{}, err
	}
	manager.nodes, manager.index = working.nodes, working.index
	if manager.cache != nil {
		manager.cache.Set(cacheKey, cloneDauTreeMergeResult(localResult))
		manager.cacheEntries[0]++
	}
	return result, nil
}

type DauTreeOptions struct {
	Variables, NonDecomposableLimit, MaxNodes int
	Stop                                      func() bool
}

type DauTreeManager struct {
	variables, nonDecomposableLimit, maxNodes int
	stop                                      func() bool
	nodes                                     []DauTreeNode
	index                                     *insdict.Dict[string, int]
	cache                                     *insdict.Dict[string, DauTreeMergeResult]
	maxCacheEntries                           int
	cacheHits, cacheMisses, cacheEntries      [2]int
}

type DauTreeCacheStats struct {
	UnsharedHits, SharedHits       int
	UnsharedMisses, SharedMisses   int
	UnsharedEntries, SharedEntries int
}

type DauTreeSummaryRow struct {
	ID, Support int
	Formula     string
}

type DauTreeSummary struct {
	Objects, PrimeObjects, PrimeStructures int
	Cache                                  DauTreeCacheStats
	Rows                                   []DauTreeSummaryRow
}

type DauTreePrimeTruth struct {
	Node  int
	Truth *TruthTable
}

func NewDauTreeManager(options DauTreeOptions) (*DauTreeManager, error) {
	if options.Variables < 0 || options.Variables > DauMaxVariables || options.NonDecomposableLimit < 0 || options.MaxNodes < 0 {
		return nil, ErrDimension
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxNodes < 2 {
		return nil, ErrLimit
	}
	manager := &DauTreeManager{
		variables: options.Variables, nonDecomposableLimit: options.NonDecomposableLimit,
		maxNodes: options.MaxNodes, stop: options.Stop, index: insdict.NewDict[string, int](),
	}
	manager.nodes = append(manager.nodes,
		DauTreeNode{Kind: DauDSDConstant},
		DauTreeNode{Kind: DauDSDVariable, Support: 1},
	)
	return manager, nil
}

func (manager *DauTreeManager) VariableCount() int {
	if manager == nil {
		return 0
	}
	return manager.variables
}

func (manager *DauTreeManager) NodeCount() int {
	if manager == nil {
		return 0
	}
	return len(manager.nodes)
}

func (manager *DauTreeManager) EnableCache(maxEntries int) error {
	if manager == nil || maxEntries < 0 {
		return ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	manager.cache = insdict.NewDict[string, DauTreeMergeResult]()
	manager.maxCacheEntries = maxEntries
	manager.cacheHits, manager.cacheMisses, manager.cacheEntries = [2]int{}, [2]int{}, [2]int{}
	return nil
}

func (manager *DauTreeManager) DisableCache() {
	if manager == nil {
		return
	}
	manager.cache = nil
	manager.maxCacheEntries = 0
}

func (manager *DauTreeManager) CacheStats() DauTreeCacheStats {
	if manager == nil {
		return DauTreeCacheStats{}
	}
	return DauTreeCacheStats{
		UnsharedHits: manager.cacheHits[0], SharedHits: manager.cacheHits[1],
		UnsharedMisses: manager.cacheMisses[0], SharedMisses: manager.cacheMisses[1],
		UnsharedEntries: manager.cacheEntries[0], SharedEntries: manager.cacheEntries[1],
	}
}

func (manager *DauTreeManager) Summary(maxRows, maxFormulaBytes int) (DauTreeSummary, error) {
	if manager == nil || maxRows < 0 || maxFormulaBytes < 0 {
		return DauTreeSummary{}, ErrDimension
	}
	if maxRows == 0 {
		maxRows = 50
	}
	result := DauTreeSummary{Objects: len(manager.nodes), Cache: manager.CacheStats()}
	result.Rows = make([]DauTreeSummaryRow, 0, min(maxRows, len(manager.nodes)))
	for id, node := range manager.nodes {
		if manager.stop != nil && manager.stop() {
			return DauTreeSummary{}, ErrCancelled
		}
		if node.Kind == DauDSDPrime {
			result.PrimeObjects++
		}
		contains, err := manager.ContainsPrime(DauTreeLiteral(id << 1))
		if err != nil {
			return DauTreeSummary{}, err
		}
		if contains {
			result.PrimeStructures++
		}
		if id >= maxRows {
			continue
		}
		mapping := make([]DauTreeLiteral, node.Support)
		for variable := range mapping {
			mapping[variable] = DauTreeLiteral(variable << 1)
		}
		formula, err := manager.Format(DauTreeLiteral(id<<1), mapping, maxFormulaBytes)
		if err != nil {
			return DauTreeSummary{}, err
		}
		result.Rows = append(result.Rows, DauTreeSummaryRow{ID: id, Support: node.Support, Formula: formula})
	}
	return result, nil
}

func FormatDauTreeSummary(summary DauTreeSummary, maxBytes int) (string, error) {
	if maxBytes < 0 {
		return "", ErrDimension
	}
	if maxBytes == 0 {
		maxBytes = 1 << 20
	}
	var result strings.Builder
	appendLine := func(line string) error {
		if result.Len() > maxBytes-len(line) {
			return ErrLimit
		}
		result.WriteString(line)
		return nil
	}
	for _, line := range []string{
		fmt.Sprintf("Total number of objects = %d\n", summary.Objects),
		fmt.Sprintf("Non-DSD objects = %d\n", summary.PrimeObjects),
		fmt.Sprintf("Non-DSD structures = %d\n", summary.PrimeStructures),
		fmt.Sprintf("Cache hits = %d %d\n", summary.Cache.UnsharedHits, summary.Cache.SharedHits),
		fmt.Sprintf("Cache misses = %d %d\n", summary.Cache.UnsharedMisses, summary.Cache.SharedMisses),
		fmt.Sprintf("Cache entries = %d %d\n", summary.Cache.UnsharedEntries, summary.Cache.SharedEntries),
	} {
		if err := appendLine(line); err != nil {
			return "", err
		}
	}
	for _, row := range summary.Rows {
		if err := appendLine(fmt.Sprintf("%6d : %2d %s\n", row.ID, row.Support, row.Formula)); err != nil {
			return "", err
		}
	}
	return result.String(), nil
}

func (manager *DauTreeManager) PrimeTruthTables() ([]DauTreePrimeTruth, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	result := make([]DauTreePrimeTruth, 0)
	for id, node := range manager.nodes {
		if manager.stop != nil && manager.stop() {
			return nil, ErrCancelled
		}
		if node.Kind != DauDSDPrime {
			continue
		}
		truth, err := TruthTableFromWords(len(node.Fanins), node.Truth, DauMaxTruthWords)
		if err != nil {
			return nil, err
		}
		truth, err = truth.Stretch(manager.variables)
		if err != nil {
			return nil, err
		}
		result = append(result, DauTreePrimeTruth{Node: id, Truth: truth})
	}
	return result, nil
}

func (manager *DauTreeManager) Nodes() []DauTreeNode {
	if manager == nil {
		return nil
	}
	result := make([]DauTreeNode, len(manager.nodes))
	for index, node := range manager.nodes {
		result[index] = DauTreeNode{Kind: node.Kind, Support: node.Support, Fanins: slices.Clone(node.Fanins), Truth: slices.Clone(node.Truth)}
	}
	return result
}

func (manager *DauTreeManager) clone() *DauTreeManager {
	clone := &DauTreeManager{
		variables: manager.variables, nonDecomposableLimit: manager.nonDecomposableLimit,
		maxNodes: manager.maxNodes, stop: manager.stop, index: insdict.NewDict[string, int](),
	}
	clone.nodes = manager.Nodes()
	for id := 2; id < len(clone.nodes); id++ {
		node := clone.nodes[id]
		clone.index.Set(dauTreeKey(node.Kind, node.Fanins, node.Truth), id)
	}
	return clone
}

func dauTreeCacheKey(shared DauTreeSharedMap) string {
	data := make([]byte, 8+8*len(shared.Variables))
	binary.LittleEndian.PutUint32(data, uint32(shared.First))
	binary.LittleEndian.PutUint32(data[4:], uint32(shared.Second))
	position := 8
	for _, variable := range shared.Variables {
		binary.LittleEndian.PutUint32(data[position:], uint32(variable.SecondIndex))
		binary.LittleEndian.PutUint32(data[position+4:], uint32(variable.First))
		position += 8
	}
	return string(data)
}

func cloneDauTreeMergeResult(result DauTreeMergeResult) DauTreeMergeResult {
	return DauTreeMergeResult{Literal: result.Literal, Mapping: slices.Clone(result.Mapping)}
}

func (manager *DauTreeManager) sharedMergeContext(first, second DauTreeLiteral, firstMapping, secondMapping []DauTreeLiteral, sharedMask uint32) (DauTreeSharedMap, []DauTreeLiteral, []DauTreeLiteral, []DauTreeLiteral, error) {
	shared, err := manager.SharedMap(first, second, firstMapping, secondMapping, sharedMask)
	if err != nil {
		return DauTreeSharedMap{}, nil, nil, nil, err
	}
	firstLocal := make([]DauTreeLiteral, len(firstMapping))
	for index := range firstLocal {
		firstLocal[index] = DauTreeLiteral(index << 1)
	}
	secondLocal := make([]DauTreeLiteral, len(secondMapping))
	sources := slices.Clone(firstMapping)
	sharedIndex := 0
	for index := range secondMapping {
		if sharedIndex < len(shared.Variables) && shared.Variables[sharedIndex].SecondIndex == index {
			secondLocal[index] = shared.Variables[sharedIndex].First
			sharedIndex++
			continue
		}
		secondLocal[index] = DauTreeLiteral(len(sources) << 1)
		sources = append(sources, secondMapping[index])
	}
	if sharedIndex != len(shared.Variables) {
		return DauTreeSharedMap{}, nil, nil, nil, ErrInvariant
	}
	return shared, firstLocal, secondLocal, sources, nil
}

func dauTreeTranslateSources(local DauTreeMergeResult, sources []DauTreeLiteral) (DauTreeMergeResult, error) {
	mapping := make([]DauTreeLiteral, len(local.Mapping))
	for index, source := range local.Mapping {
		if source.ID() < 0 || source.ID() >= len(sources) {
			return DauTreeMergeResult{}, ErrInvariant
		}
		mapping[index] = sources[source.ID()]
		if source.Complemented() {
			mapping[index] = mapping[index].Not()
		}
	}
	return DauTreeMergeResult{Literal: local.Literal, Mapping: mapping}, nil
}

func (manager *DauTreeManager) translateMergeMapping(local DauTreeMergeResult, firstMapping, secondMapping []DauTreeLiteral) (DauTreeMergeResult, error) {
	mapping := make([]DauTreeLiteral, len(local.Mapping))
	for index, source := range local.Mapping {
		localIndex := source.ID()
		var mapped DauTreeLiteral
		if localIndex < len(firstMapping) {
			mapped = firstMapping[localIndex]
		} else {
			localIndex -= len(firstMapping)
			if localIndex >= len(secondMapping) {
				return DauTreeMergeResult{}, ErrInvariant
			}
			mapped = secondMapping[localIndex]
		}
		if source.Complemented() {
			mapped = mapped.Not()
		}
		mapping[index] = mapped
	}
	return DauTreeMergeResult{Literal: local.Literal, Mapping: mapping}, nil
}

func (manager *DauTreeManager) Constant(value bool) DauTreeLiteral {
	if value {
		return 1
	}
	return 0
}

func (manager *DauTreeManager) Variable() DauTreeLiteral { return 2 }

func (manager *DauTreeManager) checkLiteral(literal DauTreeLiteral) error {
	if manager == nil || literal.ID() < 0 || literal.ID() >= len(manager.nodes) {
		return ErrDimension
	}
	return nil
}

func dauTreeKey(kind DauDSDNodeKind, fanins []DauTreeLiteral, truth []uint64) string {
	data := make([]byte, 1+4*len(fanins)+8*len(truth))
	data[0] = byte(kind)
	position := 1
	for _, fanin := range fanins {
		binary.LittleEndian.PutUint32(data[position:], uint32(fanin))
		position += 4
	}
	for _, word := range truth {
		binary.LittleEndian.PutUint64(data[position:], word)
		position += 8
	}
	return string(data)
}

func (manager *DauTreeManager) compare(first, second DauTreeLiteral) int {
	firstNode, secondNode := manager.nodes[first.ID()], manager.nodes[second.ID()]
	if firstNode.Kind != secondNode.Kind {
		if firstNode.Kind < secondNode.Kind {
			return -1
		}
		return 1
	}
	if firstNode.Kind < DauDSDAnd {
		return 0
	}
	if len(firstNode.Fanins) != len(secondNode.Fanins) {
		if len(firstNode.Fanins) < len(secondNode.Fanins) {
			return -1
		}
		return 1
	}
	for index := range firstNode.Fanins {
		if comparison := manager.compare(firstNode.Fanins[index], secondNode.Fanins[index]); comparison != 0 {
			return comparison
		}
	}
	if first.Complemented() != second.Complemented() {
		if !first.Complemented() {
			return -1
		}
		return 1
	}
	return 0
}

func (manager *DauTreeManager) sortFanins(fanins []DauTreeLiteral) {
	for first := 0; first+1 < len(fanins); first++ {
		best := first
		for next := first + 1; next < len(fanins); next++ {
			if manager.compare(fanins[best], fanins[next]) > 0 {
				best = next
			}
		}
		if best != first {
			fanins[first], fanins[best] = fanins[best], fanins[first]
		}
	}
}

func (manager *DauTreeManager) intern(kind DauDSDNodeKind, fanins []DauTreeLiteral, truth []uint64) (DauTreeLiteral, error) {
	if manager.stop != nil && manager.stop() {
		return 0, ErrCancelled
	}
	support := 0
	for _, fanin := range fanins {
		if err := manager.checkLiteral(fanin); err != nil {
			return 0, err
		}
		support += manager.nodes[fanin.ID()].Support
	}
	if support > manager.variables {
		return 0, ErrLimit
	}
	key := dauTreeKey(kind, fanins, truth)
	if id, found := manager.index.Get2(key); found {
		return DauTreeLiteral(id << 1), nil
	}
	if len(manager.nodes) >= manager.maxNodes {
		return 0, ErrLimit
	}
	id := len(manager.nodes)
	manager.nodes = append(manager.nodes, DauTreeNode{Kind: kind, Support: support, Fanins: slices.Clone(fanins), Truth: slices.Clone(truth)})
	manager.index.Set(key, id)
	return DauTreeLiteral(id << 1), nil
}

// Operation preserves Dss_ManOperation's no-permutation structural path.
func (manager *DauTreeManager) Operation(kind DauDSDNodeKind, inputs []DauTreeLiteral, prime *TruthTable) (DauTreeLiteral, error) {
	if manager == nil || len(inputs) > DauMaxVariables || len(inputs) == 0 && kind != DauDSDPrime {
		return 0, ErrDimension
	}
	for _, input := range inputs {
		if err := manager.checkLiteral(input); err != nil {
			return 0, err
		}
	}
	children := make([]DauTreeLiteral, 0, len(inputs))
	outputComplement := false
	switch kind {
	case DauDSDAnd:
		for _, input := range inputs {
			node := manager.nodes[input.ID()]
			if !input.Complemented() && node.Kind == DauDSDAnd {
				children = append(children, node.Fanins...)
			} else {
				children = append(children, input)
			}
		}
		manager.sortFanins(children)
	case DauDSDXor:
		for _, input := range inputs {
			if input.Complemented() {
				outputComplement = !outputComplement
			}
			node := manager.nodes[input.ID()]
			if node.Kind == DauDSDXor {
				children = append(children, node.Fanins...)
			} else {
				children = append(children, input&^1)
			}
		}
		manager.sortFanins(children)
	case DauDSDMux:
		if len(inputs) != 3 || prime != nil {
			return 0, ErrDimension
		}
		children = slices.Clone(inputs)
		if children[0].Complemented() {
			children[0] = children[0].Not()
			children[1], children[2] = children[2], children[1]
		}
		if children[1].Complemented() {
			children[1] = children[1].Not()
			children[2] = children[2].Not()
			outputComplement = true
		}
	case DauDSDPrime:
		if prime == nil || prime.variables != len(inputs) || prime.variables < 0 || prime.variables > DauMaxVariables {
			return 0, ErrDimension
		}
		if err := prime.Check(); err != nil {
			return 0, err
		}
		canonical, err := prime.CanonicalizeNPN(TruthCanonicalOptions{Stop: manager.stop})
		if err != nil {
			return 0, err
		}
		children = make([]DauTreeLiteral, len(inputs))
		for source, target := range canonical.Permutation {
			if target < 0 || target >= len(children) {
				return 0, ErrInvariant
			}
			children[target] = inputs[source]
			if canonical.InputPhases[source] {
				children[target] = children[target].Not()
			}
		}
		prime = canonical.Table
		outputComplement = canonical.OutputPhase
	default:
		return 0, ErrDimension
	}
	if len(children) > DauMaxVariables {
		return 0, ErrLimit
	}
	var truth []uint64
	if prime != nil {
		truth = prime.Words()
	}
	literal, err := manager.intern(kind, children, truth)
	if err != nil {
		return 0, err
	}
	if outputComplement {
		literal = literal.Not()
	}
	return literal, nil
}

type dauTreePermutationSegment struct {
	literal    DauTreeLiteral
	begin, end int
	complement bool
}

// AndWithPermutation preserves Dss_ManOperation's distinct pPerm path. The
// returned literals map the result's leaf order to the concatenated input
// supports; complemented variable leaves carry their phase in that mapping.
func (manager *DauTreeManager) AndWithPermutation(inputs []DauTreeLiteral) (DauTreeLiteral, []DauTreeLiteral, error) {
	if manager == nil || len(inputs) == 0 || len(inputs) > DauMaxVariables {
		return 0, nil, ErrDimension
	}
	segments := make([]dauTreePermutationSegment, 0, len(inputs))
	support := 0
	appendSegment := func(literal DauTreeLiteral) error {
		if err := manager.checkLiteral(literal); err != nil {
			return err
		}
		node := manager.nodes[literal.ID()]
		phase := node.Kind == DauDSDVariable && literal.Complemented()
		if phase {
			literal = literal.Not()
		}
		segments = append(segments, dauTreePermutationSegment{literal: literal, begin: support, end: support + node.Support, complement: phase})
		support += node.Support
		return nil
	}
	for _, input := range inputs {
		if err := manager.checkLiteral(input); err != nil {
			return 0, nil, err
		}
		node := manager.nodes[input.ID()]
		if !input.Complemented() && node.Kind == DauDSDAnd {
			for _, child := range node.Fanins {
				if err := appendSegment(child); err != nil {
					return 0, nil, err
				}
			}
		} else if err := appendSegment(input); err != nil {
			return 0, nil, err
		}
	}
	if support > manager.variables {
		return 0, nil, ErrLimit
	}
	for first := 0; first+1 < len(segments); first++ {
		best := first
		for next := first + 1; next < len(segments); next++ {
			if manager.compare(segments[best].literal, segments[next].literal) > 0 {
				best = next
			}
		}
		if best != first {
			segments[first], segments[best] = segments[best], segments[first]
		}
	}
	children := make([]DauTreeLiteral, len(segments))
	permutation := make([]DauTreeLiteral, 0, support)
	for index, segment := range segments {
		children[index] = segment.literal
		for variable := segment.begin; variable < segment.end; variable++ {
			literal := DauTreeLiteral(variable << 1)
			if segment.complement {
				literal = literal.Not()
			}
			permutation = append(permutation, literal)
		}
	}
	literal, err := manager.intern(DauDSDAnd, children, nil)
	if err != nil {
		return 0, nil, err
	}
	return literal, permutation, nil
}

func dauTreePrimeTable(bits []bool, variables int) (*TruthTable, error) {
	table, err := NewConstantTruthTable(variables, false, DauMaxTruthWords)
	if err != nil {
		return nil, err
	}
	for index, bit := range bits {
		if bit {
			table.words[index>>6] |= 1 << uint(index&63)
		}
	}
	if variables < 6 {
		logical := table.words[0] & dauTruthMask(variables)
		for shift := 1 << uint(variables); shift < 64; shift <<= 1 {
			logical |= logical << uint(shift)
		}
		table.words[0] = logical
	}
	return table, nil
}

// ImportFormula atomically replaces Dss_NtkCreate/Transform/Rebuild for the
// checked Go parser. It returns the structural literal and its depth-first leaf
// mapping. Mux-local binding lists are outside dauTree.c's active grammar.
func (manager *DauTreeManager) ImportFormula(text string, options DauDSDOptions) (DauTreeLiteral, []DauTreeLiteral, error) {
	if manager == nil {
		return 0, nil, ErrDimension
	}
	formula, err := ParseDauDSD(text, options)
	if err != nil {
		return 0, nil, err
	}
	if formula.variables > manager.variables {
		return 0, nil, ErrDimension
	}
	working := manager.clone()
	seenVariables := make([]bool, manager.variables)
	type importedNode struct {
		literal DauTreeLiteral
		mapping []DauTreeLiteral
	}
	type importedFanin struct {
		literal DauTreeLiteral
		mapping []DauTreeLiteral
	}
	expand := func(node importedNode) ([]importedFanin, error) {
		structure := working.nodes[node.literal.ID()]
		result := make([]importedFanin, len(structure.Fanins))
		position := 0
		for index, fanin := range structure.Fanins {
			support := working.nodes[fanin.ID()].Support
			if position+support > len(node.mapping) {
				return nil, ErrInvariant
			}
			result[index] = importedFanin{literal: fanin, mapping: slices.Clone(node.mapping[position : position+support])}
			position += support
		}
		if position != len(node.mapping) {
			return nil, ErrInvariant
		}
		return result, nil
	}
	sortFanins := func(fanins []importedFanin) {
		for first := 0; first+1 < len(fanins); first++ {
			best := first
			for next := first + 1; next < len(fanins); next++ {
				if working.compare(fanins[best].literal, fanins[next].literal) > 0 {
					best = next
				}
			}
			if best != first {
				fanins[first], fanins[best] = fanins[best], fanins[first]
			}
		}
	}
	var importNode func(int) (importedNode, error)
	importNode = func(index int) (importedNode, error) {
		if working.stop != nil && working.stop() {
			return importedNode{}, ErrCancelled
		}
		node := formula.nodes[index]
		if len(node.bindings) != 0 {
			return importedNode{}, ErrUnsupported
		}
		switch node.kind {
		case DauDSDConstant:
			value := node.value != 0
			if node.complement {
				value = !value
			}
			return importedNode{literal: working.Constant(value)}, nil
		case DauDSDVariable:
			if node.value >= len(seenVariables) || seenVariables[node.value] {
				return importedNode{}, ErrInvariant
			}
			seenVariables[node.value] = true
			mapped := DauTreeLiteral(node.value << 1)
			if node.complement {
				mapped = mapped.Not()
			}
			return importedNode{literal: working.Variable(), mapping: []DauTreeLiteral{mapped}}, nil
		}
		children := make([]importedNode, len(node.children))
		for childIndex, child := range node.children {
			children[childIndex], err = importNode(child)
			if err != nil {
				return importedNode{}, err
			}
		}
		fanins := make([]importedFanin, 0, len(children))
		var prime *TruthTable
		switch node.kind {
		case DauDSDAnd:
			for _, child := range children {
				if !child.literal.Complemented() && working.nodes[child.literal.ID()].Kind == DauDSDAnd {
					expanded, expandErr := expand(child)
					if expandErr != nil {
						return importedNode{}, expandErr
					}
					fanins = append(fanins, expanded...)
				} else {
					fanins = append(fanins, importedFanin{literal: child.literal, mapping: child.mapping})
				}
			}
			sortFanins(fanins)
		case DauDSDXor:
			for _, child := range children {
				if working.nodes[child.literal.ID()].Kind == DauDSDXor {
					expanded, expandErr := expand(child)
					if expandErr != nil {
						return importedNode{}, expandErr
					}
					fanins = append(fanins, expanded...)
				} else {
					fanins = append(fanins, importedFanin{literal: child.literal &^ 1, mapping: child.mapping})
				}
			}
			sortFanins(fanins)
		case DauDSDMux:
			if len(children) != 3 {
				return importedNode{}, ErrInvariant
			}
			for _, child := range children {
				fanins = append(fanins, importedFanin{literal: child.literal, mapping: child.mapping})
			}
			if fanins[0].literal.Complemented() {
				fanins[0].literal = fanins[0].literal.Not()
				fanins[1], fanins[2] = fanins[2], fanins[1]
			}
			if fanins[1].literal.Complemented() {
				fanins[1].literal = fanins[1].literal.Not()
				fanins[2].literal = fanins[2].literal.Not()
			}
		case DauDSDPrime:
			prime, err = dauTreePrimeTable(node.primeBits, len(children))
			if err != nil {
				return importedNode{}, err
			}
			canonical, canonicalErr := prime.CanonicalizeNPN(TruthCanonicalOptions{Stop: working.stop})
			if canonicalErr != nil {
				return importedNode{}, canonicalErr
			}
			fanins = make([]importedFanin, len(children))
			for source, target := range canonical.Permutation {
				fanins[target] = importedFanin{literal: children[source].literal, mapping: children[source].mapping}
				if canonical.InputPhases[source] {
					fanins[target].literal = fanins[target].literal.Not()
				}
			}
		default:
			return importedNode{}, ErrInvariant
		}
		childLiterals := make([]DauTreeLiteral, len(children))
		for index := range children {
			childLiterals[index] = children[index].literal
		}
		literal, err := working.Operation(node.kind, childLiterals, prime)
		if err != nil {
			return importedNode{}, err
		}
		structure := working.nodes[literal.ID()]
		if len(structure.Fanins) != len(fanins) {
			return importedNode{}, ErrInvariant
		}
		mapping := make([]DauTreeLiteral, 0, structure.Support)
		for index, fanin := range fanins {
			if structure.Fanins[index] != fanin.literal {
				return importedNode{}, ErrInvariant
			}
			mapping = append(mapping, fanin.mapping...)
		}
		if node.complement {
			literal = literal.Not()
		}
		return importedNode{literal: literal, mapping: mapping}, nil
	}
	result, err := importNode(formula.root)
	if err != nil {
		return 0, nil, err
	}
	if len(result.mapping) != working.nodes[result.literal.ID()].Support {
		return 0, nil, ErrInvariant
	}
	manager.nodes, manager.index = working.nodes, working.index
	return result.literal, result.mapping, nil
}

func (manager *DauTreeManager) Truth(literal DauTreeLiteral, permutation []DauTreeLiteral, maxWords int, stop func() bool) (*TruthTable, error) {
	if err := manager.checkLiteral(literal); err != nil {
		return nil, err
	}
	if len(permutation) != manager.nodes[literal.ID()].Support || len(permutation) > manager.variables {
		return nil, ErrDimension
	}
	for _, variable := range permutation {
		if variable.ID() < 0 || variable.ID() >= manager.variables {
			return nil, ErrDimension
		}
	}
	result, err := NewConstantTruthTable(manager.variables, false, maxWords)
	if err != nil {
		return nil, err
	}
	var evaluate func(DauTreeLiteral, uint64, *int) (bool, error)
	evaluate = func(current DauTreeLiteral, assignment uint64, position *int) (bool, error) {
		if stop != nil && stop() {
			return false, ErrCancelled
		}
		node := manager.nodes[current.ID()]
		value := false
		switch node.Kind {
		case DauDSDConstant:
		case DauDSDVariable:
			mapped := permutation[*position]
			*position++
			value = assignment&(1<<uint(mapped.ID())) != 0
			if mapped.Complemented() {
				value = !value
			}
		case DauDSDAnd, DauDSDXor:
			value = node.Kind == DauDSDAnd
			for _, child := range node.Fanins {
				childValue, err := evaluate(child, assignment, position)
				if err != nil {
					return false, err
				}
				if node.Kind == DauDSDAnd {
					value = value && childValue
				} else {
					value = value != childValue
				}
			}
		case DauDSDMux:
			values := [3]bool{}
			for index, child := range node.Fanins {
				childValue, err := evaluate(child, assignment, position)
				if err != nil {
					return false, err
				}
				values[index] = childValue
			}
			if values[0] {
				value = values[1]
			} else {
				value = values[2]
			}
		case DauDSDPrime:
			local := 0
			for index, child := range node.Fanins {
				childValue, err := evaluate(child, assignment, position)
				if err != nil {
					return false, err
				}
				if childValue {
					local |= 1 << uint(index)
				}
			}
			value = node.Truth[local>>6]&(1<<uint(local&63)) != 0
		default:
			return false, ErrInvariant
		}
		if current.Complemented() {
			value = !value
		}
		return value, nil
	}
	assignments := 1 << uint(manager.variables)
	for assignment := 0; assignment < assignments; assignment++ {
		position := 0
		value, err := evaluate(literal, uint64(assignment), &position)
		if err != nil {
			return nil, err
		}
		if position != len(permutation) {
			return nil, ErrInvariant
		}
		if value {
			result.words[assignment>>6] |= 1 << uint(assignment&63)
		}
	}
	if manager.variables < 6 {
		logical := result.words[0] & dauTruthMask(manager.variables)
		for shift := 1 << uint(manager.variables); shift < 64; shift <<= 1 {
			logical |= logical << uint(shift)
		}
		result.words[0] = logical
	}
	return result, nil
}
