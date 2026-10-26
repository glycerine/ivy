package abc

// Ported from ABC src/aig/aig/aigCuts.c and aigFact.c.

import (
	"errors"
	"fmt"
	"math/bits"
	"slices"
	"strings"
)

var ErrAigCutClosed = errors.New("AIG cut manager is closed")

// AigCutOptions bounds priority-cut enumeration. Zero work and word limits
// select conservative defaults. Cuts include the source's unit cut.
type AigCutOptions struct {
	MaximumCuts   int
	MaximumLeaves int
	ComputeTruth  bool
	MaximumWork   int
	MaximumWords  int
	Stop          func() bool
}

// AigKCut is an owned K-feasible cut. Leaves are sorted object IDs. Truth, when
// requested, is compact over Leaves in the same order.
type AigKCut struct {
	Node      AigID
	Leaves    []AigID
	Cost      int
	Signature uint32
	Truth     *TruthTable
}

func (c AigKCut) clone() AigKCut {
	c.Leaves = slices.Clone(c.Leaves)
	if c.Truth != nil {
		c.Truth = c.Truth.Clone()
	}
	return c
}

// AigCutManager owns a frozen graph snapshot and its priority cuts.
type AigCutManager struct {
	graph  *AigManager
	limit  int
	truth  bool
	cuts   [][]AigKCut
	closed bool
}

type aigCutBudget struct {
	left int
	stop func() bool
}

func (b *aigCutBudget) step() error {
	if b.stop != nil && b.stop() {
		return ErrCancelled
	}
	if b.left == 0 {
		return ErrLimit
	}
	b.left--
	return nil
}

func aigCutSignature(leaves []AigID) uint32 {
	var signature uint32
	for _, leaf := range leaves {
		signature |= uint32(1) << (uint32(leaf) & 31)
	}
	return signature
}

func aigCutMergeLeaves(first, second []AigID, maximum int) ([]AigID, bool) {
	merged := make([]AigID, 0, min(maximum, len(first)+len(second)))
	i, j := 0, 0
	for i < len(first) || j < len(second) {
		var next AigID
		switch {
		case j == len(second), i < len(first) && first[i] < second[j]:
			next = first[i]
			i++
		case i == len(first), second[j] < first[i]:
			next = second[j]
			j++
		default:
			next = first[i]
			i++
			j++
		}
		if len(merged) == maximum {
			return nil, false
		}
		merged = append(merged, next)
	}
	return merged, true
}

func aigCutContained(subset, superset []AigID) bool {
	i, j := 0, 0
	for i < len(subset) && j < len(superset) {
		switch {
		case subset[i] == superset[j]:
			i++
			j++
		case subset[i] > superset[j]:
			j++
		default:
			return false
		}
	}
	return i == len(subset)
}

func aigCutTruth(cut, first, second AigKCut, complementFirst, complementSecond bool, maximumWords int) (*TruthTable, error) {
	positions := func(leaves []AigID) []int {
		mapping := make([]int, len(leaves))
		for i, leaf := range leaves {
			mapping[i] = slices.Index(cut.Leaves, leaf)
		}
		return mapping
	}
	left, err := first.Truth.ExpandVariables(len(cut.Leaves), positions(first.Leaves))
	if err != nil {
		return nil, err
	}
	right, err := second.Truth.ExpandVariables(len(cut.Leaves), positions(second.Leaves))
	if err != nil {
		return nil, err
	}
	if len(left.words)+len(right.words) > maximumWords {
		return nil, ErrLimit
	}
	return left.And(right, complementFirst, complementSecond)
}

func aigCutCost(graph *AigManager, leaves []AigID) int {
	cost := uint64(0)
	for _, leaf := range leaves {
		cost += uint64(graph.objects[leaf].RefCount)
	}
	cost = cost * 1000 / uint64(len(leaves))
	if cost > uint64(^uint(0)>>1) {
		return int(^uint(0) >> 1)
	}
	return int(cost)
}

// ComputeAigCuts implements Aig_ComputeCuts, including source-order slot
// replacement, signature screening, dominance filtering and optional truth.
// It never attaches scratch state to or mutates graph.
func ComputeAigCuts(graph *AigManager, options AigCutOptions) (*AigCutManager, error) {
	if graph == nil || options.MaximumCuts < 2 || options.MaximumLeaves < 1 || options.MaximumLeaves > 16 || options.MaximumWork < 0 || options.MaximumWords < 0 {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	work := options.MaximumWork
	if work == 0 {
		work = 1 << 22
	}
	words := options.MaximumWords
	if words == 0 {
		words = 1 << 20
	}
	if options.ComputeTruth {
		if _, err := TruthTableWordCount(options.MaximumLeaves); err != nil {
			return nil, err
		}
	}
	frozen := graph.Clone()
	manager := &AigCutManager{graph: frozen, limit: options.MaximumLeaves, truth: options.ComputeTruth, cuts: make([][]AigKCut, frozen.ObjectCount())}
	budget := aigCutBudget{left: work, stop: options.Stop}
	usedWords := 0
	unit := func(id AigID) (AigKCut, error) {
		cut := AigKCut{Node: id, Leaves: []AigID{id}, Signature: aigCutSignature([]AigID{id})}
		if options.ComputeTruth {
			truth, err := NewVariableTruthTable(1, 0, words)
			if err != nil {
				return AigKCut{}, err
			}
			cut.Truth = truth
		}
		return cut, nil
	}
	for _, id := range frozen.cis {
		if err := budget.step(); err != nil {
			return nil, err
		}
		cut, err := unit(id)
		if err != nil {
			return nil, err
		}
		manager.cuts[id] = []AigKCut{cut}
		if cut.Truth != nil {
			usedWords += len(cut.Truth.words)
			if usedWords > words {
				return nil, ErrLimit
			}
		}
	}
	for raw, object := range frozen.objects {
		id := AigID(raw)
		if object.Type == AigNone || object.Type == AigConstant1 || object.Type == AigCI || object.Type == AigCO {
			continue
		}
		if object.Type != AigAND {
			return nil, ErrUnsupported
		}
		if err := budget.step(); err != nil {
			return nil, err
		}
		slots := make([]*AigKCut, options.MaximumCuts)
		trivial, err := unit(id)
		if err != nil {
			return nil, err
		}
		slots[0] = &trivial
		leftCuts, rightCuts := manager.cuts[object.Fanin0.ID()], manager.cuts[object.Fanin1.ID()]
		if len(leftCuts) == 0 || len(rightCuts) == 0 {
			return nil, ErrInvariant
		}
		for _, left := range leftCuts {
			for _, right := range rightCuts {
				if err := budget.step(); err != nil {
					return nil, err
				}
				if bits.OnesCount32(left.Signature|right.Signature) > options.MaximumLeaves {
					continue
				}
				free := -1
				for i := range slots {
					if slots[i] == nil {
						free = i
						break
					}
				}
				if free < 0 {
					free = 0
					for i := 1; i < len(slots); i++ {
						if slots[free].Cost < slots[i].Cost {
							free = i
						}
					}
					slots[free] = nil
				}
				leaves, ok := aigCutMergeLeaves(left.Leaves, right.Leaves, options.MaximumLeaves)
				if !ok {
					continue
				}
				candidate := AigKCut{Node: id, Leaves: leaves, Signature: left.Signature | right.Signature}
				discard := false
				for i, current := range slots {
					if current == nil || len(current.Leaves) < 2 {
						continue
					}
					if len(current.Leaves) > len(candidate.Leaves) {
						if current.Signature&candidate.Signature == candidate.Signature && aigCutContained(candidate.Leaves, current.Leaves) {
							slots[i] = nil
						}
					} else if current.Signature&candidate.Signature == current.Signature && aigCutContained(current.Leaves, candidate.Leaves) {
						discard = true
						break
					}
				}
				if discard {
					continue
				}
				if options.ComputeTruth {
					candidate.Truth, err = aigCutTruth(candidate, left, right, object.Fanin0.Complemented(), object.Fanin1.Complemented(), words)
					if err != nil {
						return nil, err
					}
				}
				candidate.Cost = aigCutCost(frozen, candidate.Leaves)
				slots[free] = &candidate
			}
		}
		nodeWords := 0
		for _, cut := range slots {
			if cut != nil {
				manager.cuts[id] = append(manager.cuts[id], cut.clone())
				if cut.Truth != nil {
					nodeWords += len(cut.Truth.words)
				}
			}
		}
		if nodeWords > words-usedWords {
			return nil, ErrLimit
		}
		usedWords += nodeWords
	}
	return manager, nil
}

func (m *AigCutManager) Cuts(id AigID) ([]AigKCut, error) {
	if m == nil || m.closed {
		return nil, ErrAigCutClosed
	}
	if uint64(id) >= uint64(len(m.cuts)) {
		return nil, ErrInvalidLiteral
	}
	result := make([]AigKCut, len(m.cuts[id]))
	for i := range result {
		result[i] = m.cuts[id][i].clone()
	}
	return result, nil
}

func (m *AigCutManager) Count() (total, maximumSize int, err error) {
	if m == nil || m.closed {
		return 0, 0, ErrAigCutClosed
	}
	for id, object := range m.graph.objects {
		if object.Type != AigAND {
			continue
		}
		for _, cut := range m.cuts[id] {
			total++
			if len(cut.Leaves) == m.limit {
				maximumSize++
			}
		}
	}
	return total, maximumSize, nil
}

// ReciprocalCost implements the alternate Aig_CutFindCost2 metric. Zero
// reference leaves contribute an infinite reciprocal and therefore cost zero,
// matching floating-point evaluation of the source expression.
func (m *AigCutManager) ReciprocalCost(cut AigKCut) (float32, error) {
	if m == nil || m.closed {
		return 0, ErrAigCutClosed
	}
	if len(cut.Leaves) == 0 {
		return 0, ErrDimension
	}
	var sum float32
	for _, id := range cut.Leaves {
		if uint64(id) >= uint64(len(m.graph.objects)) || m.graph.objects[id].Type == AigNone {
			return 0, ErrInvalidLiteral
		}
		refs := m.graph.objects[id].RefCount
		if refs == 0 {
			return 0, nil
		}
		sum += 1 / float32(refs)
	}
	return 1 / sum, nil
}

func (m *AigCutManager) Format(id AigID) (string, error) {
	cuts, err := m.Cuts(id)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	for _, cut := range cuts {
		result.WriteByte('{')
		for _, leaf := range cut.Leaves {
			fmt.Fprintf(&result, " %d", leaf)
		}
		result.WriteString(" }\n")
	}
	return result.String(), nil
}

func (m *AigCutManager) Close() error {
	if m == nil || m.closed {
		return ErrAigCutClosed
	}
	m.graph, m.cuts, m.closed = nil, nil, true
	return nil
}

// MinimizeSupport implements Aig_CutSupportMinimize on an owned cut.
func (m *AigCutManager) MinimizeSupport(cut AigKCut) (AigKCut, error) {
	if m == nil || m.closed {
		return AigKCut{}, ErrAigCutClosed
	}
	if cut.Truth == nil || len(cut.Leaves) != cut.Truth.Variables() {
		return AigKCut{}, ErrDimension
	}
	reduction, err := cut.Truth.CompactSupport()
	if err != nil {
		return AigKCut{}, err
	}
	leaves := make([]AigID, len(reduction.SourceVariables))
	for i, source := range reduction.SourceVariables {
		leaves[i] = cut.Leaves[source]
	}
	cut.Leaves, cut.Truth = leaves, reduction.Table
	cut.Signature = aigCutSignature(leaves)
	return cut, nil
}

// AigFactorOptions bounds implication, overlap and reconstruction traversals.
type AigFactorOptions struct {
	MaximumWork int
	Stop        func() bool
}

// AigFactorResult contains an independently owned graph. Literal belongs to
// Graph when Found is true. The input graph is never changed.
type AigFactorResult struct {
	Graph        *AigManager
	Literal      AigLit
	Found        bool
	Implications []AigLit
}

func newAigFactorBudget(graph *AigManager, options AigFactorOptions) (*aigCutBudget, error) {
	if graph == nil || options.MaximumWork < 0 {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	work := options.MaximumWork
	if work == 0 {
		work = 1 << 22
	}
	return &aigCutBudget{left: work, stop: options.Stop}, nil
}

func aigImplications(graph *AigManager, root AigLit, budget *aigCutBudget) ([]AigLit, error) {
	if !graph.validSignal(root) || root.ID() == 0 {
		return nil, ErrInvalidLiteral
	}
	result := make([]AigLit, 0, 8)
	seen := make(map[AigLit]struct{})
	stack := []AigLit{root}
	for len(stack) > 0 {
		if err := budget.step(); err != nil {
			return nil, err
		}
		literal := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		object := graph.objects[literal.ID()]
		if literal.Complemented() || object.Type == AigCI {
			if _, exists := seen[literal]; !exists {
				seen[literal] = struct{}{}
				result = append(result, literal)
			}
			continue
		}
		if object.Type != AigAND {
			return nil, ErrUnsupported
		}
		stack = append(stack, object.Fanin1, object.Fanin0)
	}
	return result, nil
}

// AigImplications implements the source's backward implication/supergate
// boundary collection and returns owned literals in depth-first order.
func (graph *AigManager) AigImplications(root AigLit, options AigFactorOptions) ([]AigLit, error) {
	budget, err := newAigFactorBudget(graph, options)
	if err != nil {
		return nil, err
	}
	return aigImplications(graph, root, budget)
}

// AigConeOverlaps reports whether root's transitive fanin contains the regular
// object of an implication literal.
func (graph *AigManager) AigConeOverlaps(root AigLit, implications []AigLit, options AigFactorOptions) (bool, error) {
	budget, err := newAigFactorBudget(graph, options)
	if err != nil {
		return false, err
	}
	if !graph.validSignal(root) || root.ID() == 0 {
		return false, ErrInvalidLiteral
	}
	marked := make([]bool, len(graph.objects))
	for _, literal := range implications {
		if !graph.validSignal(literal) || literal.ID() == 0 {
			return false, ErrInvalidLiteral
		}
		marked[literal.ID()] = true
	}
	return aigConeOverlap(graph, root.ID(), marked, budget)
}

func aigConeOverlap(graph *AigManager, root AigID, marked []bool, budget *aigCutBudget) (bool, error) {
	seen := make([]bool, len(graph.objects))
	stack := []AigID{root}
	for len(stack) > 0 {
		if err := budget.step(); err != nil {
			return false, err
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if marked[id] {
			return true, nil
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		object := graph.objects[id]
		if object.Type == AigCI {
			continue
		}
		if object.Type != AigAND {
			return false, ErrUnsupported
		}
		stack = append(stack, object.Fanin1.ID(), object.Fanin0.ID())
	}
	return false, nil
}

func aigReconstruct(graph *AigManager, root AigLit, replacements map[AigID]AigLit, budget *aigCutBudget) (AigLit, error) {
	mapping := make([]AigLit, len(graph.objects))
	known := make([]bool, len(graph.objects))
	mapping[0], known[0] = AigConst1, true
	for id, literal := range replacements {
		mapping[id], known[id] = literal, true
	}
	for _, id := range graph.cis {
		if !known[id] {
			mapping[id], known[id] = AigLit(id)<<1, true
		}
	}
	type frame struct {
		id    AigID
		ready bool
	}
	stack := []frame{{id: root.ID()}}
	for len(stack) > 0 {
		if err := budget.step(); err != nil {
			return InvalidAigLit, err
		}
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if known[item.id] {
			continue
		}
		object := graph.objects[item.id]
		if object.Type != AigAND {
			return InvalidAigLit, ErrUnsupported
		}
		if !item.ready {
			stack = append(stack, frame{id: item.id, ready: true})
			if !known[object.Fanin1.ID()] {
				stack = append(stack, frame{id: object.Fanin1.ID()})
			}
			if !known[object.Fanin0.ID()] {
				stack = append(stack, frame{id: object.Fanin0.ID()})
			}
			continue
		}
		left := mapping[object.Fanin0.ID()].NotIf(object.Fanin0.Complemented())
		right := mapping[object.Fanin1.ID()].NotIf(object.Fanin1.Complemented())
		literal, err := graph.And(left, right)
		if err != nil {
			return InvalidAigLit, err
		}
		mapping[item.id], known[item.id] = literal, true
	}
	return mapping[root.ID()].NotIf(root.Complemented()), nil
}

// DeriveAigCone replaces each implication literal by true and reconstructs
// root in an independent graph. Thus a positive implication fixes its regular
// object to one, while a complemented implication fixes it to zero.
func (graph *AigManager) DeriveAigCone(root AigLit, implications []AigLit, options AigFactorOptions) (AigFactorResult, error) {
	budget, err := newAigFactorBudget(graph, options)
	if err != nil {
		return AigFactorResult{}, err
	}
	if !graph.validSignal(root) || root.ID() == 0 {
		return AigFactorResult{}, ErrInvalidLiteral
	}
	replacements := make(map[AigID]AigLit, len(implications))
	for _, literal := range implications {
		if !graph.validSignal(literal) || literal.ID() == 0 {
			return AigFactorResult{}, ErrInvalidLiteral
		}
		value := AigConst1.NotIf(literal.Complemented())
		if previous, exists := replacements[literal.ID()]; exists && previous != value {
			return AigFactorResult{}, ErrInvariant
		}
		replacements[literal.ID()] = value
	}
	clone := graph.Clone()
	literal, err := aigReconstruct(clone, root, replacements, budget)
	if err != nil {
		return AigFactorResult{}, err
	}
	return AigFactorResult{Graph: clone, Literal: literal, Found: true, Implications: slices.Clone(implications)}, nil
}

// FactorAlgebraic implements Aig_ManFactorAlgebraic for two output indices.
func (graph *AigManager) FactorAlgebraic(outputA, outputB int, value bool, options AigFactorOptions) (AigFactorResult, error) {
	budget, err := newAigFactorBudget(graph, options)
	if err != nil {
		return AigFactorResult{}, err
	}
	if outputA < 0 || outputA >= len(graph.cos) || outputB < 0 || outputB >= len(graph.cos) {
		return AigFactorResult{}, ErrDimension
	}
	driverA := graph.objects[graph.cos[outputA]].Fanin0
	driverB := graph.objects[graph.cos[outputB]].Fanin0
	copyGraph := graph.Clone()
	result := AigFactorResult{Graph: copyGraph, Literal: InvalidAigLit}
	if driverA.ID() == 0 || driverB.ID() == 0 || graph.objects[driverB.ID()].Type == AigCI {
		return result, nil
	}
	impliedRoot := driverA.NotIf(!value)
	implications, err := aigImplications(graph, impliedRoot, budget)
	if err != nil {
		return AigFactorResult{}, err
	}
	result.Implications = slices.Clone(implications)
	marked := make([]bool, len(graph.objects))
	for _, literal := range implications {
		marked[literal.ID()] = true
	}
	overlap, err := aigConeOverlap(graph, driverB.ID(), marked, budget)
	if err != nil {
		return AigFactorResult{}, err
	}
	if !overlap {
		return result, nil
	}
	replacements := make(map[AigID]AigLit, len(implications))
	for _, literal := range implications {
		replacements[literal.ID()] = AigConst1.NotIf(literal.Complemented())
	}
	literal, err := aigReconstruct(copyGraph, driverB, replacements, budget)
	if err != nil {
		return AigFactorResult{}, err
	}
	result.Literal, result.Found = literal, true
	return result, nil
}

// AigSupportMinStatus preserves the source's three non-success dispositions.
type AigSupportMinStatus int

const (
	AigSupportMinimized    AigSupportMinStatus = 0
	AigSupportNotOR        AigSupportMinStatus = -1
	AigSupportNoCone       AigSupportMinStatus = -2
	AigSupportNotReducible AigSupportMinStatus = -3
)

type AigSupportMinResult struct {
	Graph   *AigManager
	Literal AigLit
	Status  AigSupportMinStatus
}

func aigCollectORSuper(graph *AigManager, condition AigLit, budget *aigCutBudget) ([]AigLit, error) {
	object := graph.objects[condition.ID()]
	stack := []AigLit{object.Fanin1, object.Fanin0}
	result := make([]AigLit, 0, 4)
	seen := make(map[AigLit]struct{})
	for len(stack) > 0 {
		if err := budget.step(); err != nil {
			return nil, err
		}
		literal := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		child := graph.objects[literal.ID()]
		if literal.Complemented() || child.Type == AigCI {
			term := literal.Not()
			if _, exists := seen[term]; !exists {
				seen[term] = struct{}{}
				result = append(result, term)
			}
			continue
		}
		if child.Type != AigAND {
			return nil, ErrUnsupported
		}
		stack = append(stack, child.Fanin1, child.Fanin0)
	}
	return result, nil
}

func aigLimitedCone(graph *AigManager, root AigID, boundaries []AigLit, budget *aigCutBudget) ([]AigID, []AigID, bool, error) {
	boundary := make([]bool, len(graph.objects))
	for _, literal := range boundaries {
		boundary[literal.ID()] = true
	}
	reaches := make([]bool, len(graph.objects))
	state := make([]uint8, len(graph.objects))
	nodes := make([]AigID, 0)
	type frame struct {
		id    AigID
		ready bool
	}
	stack := []frame{{id: root}}
	for len(stack) > 0 {
		if err := budget.step(); err != nil {
			return nil, nil, false, err
		}
		item := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if state[item.id] == 2 {
			continue
		}
		if boundary[item.id] {
			reaches[item.id], state[item.id] = true, 2
			continue
		}
		object := graph.objects[item.id]
		if object.Type == AigCI || object.Type == AigConstant1 {
			state[item.id] = 2
			continue
		}
		if object.Type != AigAND {
			return nil, nil, false, ErrUnsupported
		}
		if !item.ready {
			state[item.id] = 1
			stack = append(stack, frame{id: item.id, ready: true}, frame{id: object.Fanin1.ID()}, frame{id: object.Fanin0.ID()})
			continue
		}
		state[item.id] = 2
		if reaches[object.Fanin0.ID()] || reaches[object.Fanin1.ID()] {
			reaches[item.id] = true
			nodes = append(nodes, item.id)
		}
	}
	if !reaches[root] {
		return nil, nil, false, nil
	}
	inside := make([]bool, len(graph.objects))
	for _, id := range nodes {
		inside[id] = true
	}
	support, supportSeen := make([]AigID, 0), make([]bool, len(graph.objects))
	for _, id := range nodes {
		object := graph.objects[id]
		for _, fanin := range []AigID{object.Fanin0.ID(), object.Fanin1.ID()} {
			if !inside[fanin] && !supportSeen[fanin] {
				supportSeen[fanin] = true
				support = append(support, fanin)
			}
		}
	}
	return nodes, support, true, nil
}

func truthEqualWithCare(first, second, care *TruthTable) (bool, error) {
	difference, err := first.Xor(second, false)
	if err != nil {
		return false, err
	}
	masked, err := difference.And(care, false, false)
	if err != nil {
		return false, err
	}
	return masked.IsConstant(false), nil
}

// MinimizeSupportUnderCondition implements Aig_ManSupportMinimization. The
// condition must be an AIG OR (a complemented regular AND). Success returns a
// graph clone containing the reconstructed result; failures still return an
// independent graph and the source status without partial mutations.
func (graph *AigManager) MinimizeSupportUnderCondition(condition, function AigLit, options AigFactorOptions) (AigSupportMinResult, error) {
	budget, err := newAigFactorBudget(graph, options)
	if err != nil {
		return AigSupportMinResult{}, err
	}
	result := AigSupportMinResult{Graph: graph.Clone(), Literal: InvalidAigLit}
	if !graph.validSignal(condition) || !graph.validSignal(function) {
		return AigSupportMinResult{}, ErrInvalidLiteral
	}
	if !condition.Complemented() || condition.ID() == 0 || graph.objects[condition.ID()].Type == AigCI {
		result.Status = AigSupportNotOR
		return result, nil
	}
	if graph.objects[condition.ID()].Type != AigAND || graph.objects[function.ID()].Type != AigAND {
		result.Status = AigSupportNoCone
		return result, nil
	}
	orGate, err := aigCollectORSuper(graph, condition, budget)
	if err != nil {
		return AigSupportMinResult{}, err
	}
	if len(orGate) > 16 {
		result.Status = AigSupportNoCone
		return result, nil
	}
	nodes, support, overlap, err := aigLimitedCone(graph, function.ID(), orGate, budget)
	if err != nil {
		return AigSupportMinResult{}, err
	}
	if !overlap || len(support) > 16 {
		result.Status = AigSupportNoCone
		return result, nil
	}
	supportPosition := make(map[AigID]int, len(support))
	for i, id := range support {
		supportPosition[id] = i
	}
	for _, literal := range orGate {
		if _, exists := supportPosition[literal.ID()]; !exists {
			result.Status = AigSupportNotReducible
			return result, nil
		}
	}
	truthOptions := AigCutTruthOptions{MaxObjects: len(nodes) + 1, MaxMappingEntries: graph.ObjectCount(), MaxWords: 1 << 16, Stop: options.Stop}
	functionTruth, err := graph.CutTruth(function.Regular(), support, truthOptions)
	if err != nil {
		return AigSupportMinResult{}, err
	}
	care, err := NewConstantTruthTable(len(support), false, 1<<16)
	if err != nil {
		return AigSupportMinResult{}, err
	}
	for _, literal := range orGate {
		term, e := NewVariableTruthTable(len(support), supportPosition[literal.ID()], 1<<16)
		if e != nil {
			return AigSupportMinResult{}, e
		}
		if literal.Complemented() {
			term = term.Not()
		}
		care, e = care.Or(term)
		if e != nil {
			return AigSupportMinResult{}, e
		}
	}
	replacements := make(map[AigID]AigLit, len(support))
	for _, id := range support {
		replacements[id] = AigLit(id) << 1
	}
	reduced := 0
	for i, id := range support {
		if err := budget.step(); err != nil {
			return AigSupportMinResult{}, err
		}
		f0, e := functionTruth.Table.Cofactor(i, false)
		if e != nil {
			return AigSupportMinResult{}, e
		}
		f1, e := functionTruth.Table.Cofactor(i, true)
		if e != nil {
			return AigSupportMinResult{}, e
		}
		c0, e := care.Cofactor(i, false)
		if e != nil {
			return AigSupportMinResult{}, e
		}
		c1, e := care.Cofactor(i, true)
		if e != nil {
			return AigSupportMinResult{}, e
		}
		if equal, e := truthEqualWithCare(f0, f1, c1); e != nil {
			return AigSupportMinResult{}, e
		} else if equal {
			replacements[id] = AigConst0
			reduced++
		} else if equal, e = truthEqualWithCare(f0, f1, c0); e != nil {
			return AigSupportMinResult{}, e
		} else if equal {
			replacements[id] = AigConst1
			reduced++
		}
	}
	if reduced == 0 {
		result.Status = AigSupportNotReducible
		return result, nil
	}
	literal, err := aigReconstruct(result.Graph, function, replacements, budget)
	if err != nil {
		return AigSupportMinResult{}, err
	}
	result.Literal, result.Status = literal, AigSupportMinimized
	return result, nil
}
