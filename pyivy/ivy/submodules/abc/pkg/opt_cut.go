package abc

// Ported from src/opt/cut/cut.h, cutInt.h, cutList.h, cutApi.c, cutCut.c,
// cutExpand.c, cutMan.c, cutMerge.c, cutNode.c, and cutTruth.c.

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	KCutMinLeaves = 3
	KCutMaxLeaves = 12
)

var (
	ErrKCutClosed    = errors.New("cut manager is closed")
	ErrKCutCancelled = errors.New("cut computation cancelled")
	ErrKCutBounds    = errors.New("cut index or size is out of bounds")
	ErrKCutInvalid   = errors.New("invalid cut")
)

// KCutOptions is the owned equivalent of Cut_Params_t. The callback is polled
// at deterministic merge boundaries and must not mutate the manager.
type KCutOptions struct {
	MaxLeaves    int
	KeepMax      int
	MaxNodes     int
	ComputeTruth bool
	Filter       bool
	Drop         bool
	Record       bool
	GlobalOnly   bool
	LocalOnly    bool
	Cancelled    func() bool
}

// KCut stores sorted node IDs and, when requested, a packed truth table whose
// bit i is the value for the little-endian assignment i over Leaves.
type KCut struct {
	Leaves     []int
	Truth      []uint64
	Signature  uint32
	Number0    int
	Number1    int
	Simulated  bool
	Complement bool
}

// KCutStats is an owned replacement for the source's process output.
type KCutStats struct {
	Nodes            int
	NodesWithoutCuts int
	CutsCreated      int
	CutsFiltered     int
	NodesLimited     int
	CutsFreed        int
	PeakStoredCuts   int
	DAGNodes         int
}

// KCutRecord preserves the source parent-cut recording without packed 16-bit
// truncation.
type KCutRecord struct {
	Node    int
	Parent0 int
	Parent1 int
}

// KCutCanonical is the owned result of Cut_TruthNCanonicize. Permutation maps
// canonical variable positions to original positions; Phase complements the
// corresponding original variables before evaluation.
type KCutCanonical struct {
	Truth         uint64
	Permutation   []int
	Phase         uint8
	OutputNegated bool
}

// KCutManager owns every cut and never exposes its internal slices.
type KCutManager struct {
	opts    KCutOptions
	cuts    [][]KCut
	fanouts []int
	global  []bool
	records []KCutRecord
	delays  []int
	best    []*KCut
	stats   KCutStats
	closed  bool
}

func NewKCutManager(opts KCutOptions) (*KCutManager, error) {
	if opts.MaxLeaves < KCutMinLeaves || opts.MaxLeaves > KCutMaxLeaves || opts.KeepMax <= 0 || opts.MaxNodes <= 0 {
		return nil, ErrKCutBounds
	}
	if opts.GlobalOnly && opts.LocalOnly {
		return nil, ErrKCutInvalid
	}
	return &KCutManager{
		opts:    opts,
		cuts:    make([][]KCut, opts.MaxNodes),
		fanouts: make([]int, opts.MaxNodes),
		global:  make([]bool, opts.MaxNodes),
		delays:  make([]int, opts.MaxNodes),
		best:    make([]*KCut, opts.MaxNodes),
	}, nil
}

func (m *KCutManager) check() error {
	if m == nil || m.closed {
		return ErrKCutClosed
	}
	if m.opts.Cancelled != nil && m.opts.Cancelled() {
		return ErrKCutCancelled
	}
	return nil
}

func (m *KCutManager) checkNode(node int) error {
	if err := m.check(); err != nil {
		return err
	}
	if node < 0 || node >= len(m.cuts) {
		return ErrKCutBounds
	}
	return nil
}

func (m *KCutManager) Close() {
	if m == nil || m.closed {
		return
	}
	for i := range m.cuts {
		m.stats.CutsFreed += len(m.cuts[i])
		m.cuts[i] = nil
		m.best[i] = nil
	}
	m.records = nil
	m.closed = true
}

func (m *KCutManager) Options() (KCutOptions, error) {
	if err := m.check(); err != nil {
		return KCutOptions{}, err
	}
	return m.opts, nil
}

func (m *KCutManager) Stats() (KCutStats, error) {
	if err := m.check(); err != nil {
		return KCutStats{}, err
	}
	return m.stats, nil
}

func (m *KCutManager) MaxLeaves() (int, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	return m.opts.MaxLeaves, nil
}

func (m *KCutManager) IncrementDAGNodes() error {
	if err := m.check(); err != nil {
		return err
	}
	m.stats.DAGNodes++
	return nil
}

func (m *KCutManager) Records() ([]KCutRecord, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	return append([]KCutRecord(nil), m.records...), nil
}

func (m *KCutManager) SetFanoutCounts(counts []int) error {
	if err := m.check(); err != nil {
		return err
	}
	if len(counts) > len(m.fanouts) {
		return ErrKCutBounds
	}
	for _, n := range counts {
		if n < 0 {
			return ErrKCutInvalid
		}
	}
	clear(m.fanouts)
	copy(m.fanouts, counts)
	return nil
}

func (m *KCutManager) SetGlobalNodes(attrs []bool) error {
	if err := m.check(); err != nil {
		return err
	}
	if len(attrs) > len(m.global) {
		return ErrKCutBounds
	}
	clear(m.global)
	copy(m.global, attrs)
	return nil
}

func kCutSignature(leaves []int) uint32 {
	var sign uint32
	for _, leaf := range leaves {
		sign |= uint32(1) << uint(leaf%31)
	}
	return sign
}

func kCutClone(c KCut) KCut {
	c.Leaves = append([]int(nil), c.Leaves...)
	c.Truth = append([]uint64(nil), c.Truth...)
	return c
}

func cloneKCuts(cuts []KCut) []KCut {
	res := make([]KCut, len(cuts))
	for i := range cuts {
		res[i] = kCutClone(cuts[i])
	}
	return res
}

func (m *KCutManager) Cuts(node int) ([]KCut, error) {
	if err := m.checkNode(node); err != nil {
		return nil, err
	}
	return cloneKCuts(m.cuts[node]), nil
}

func kCutTruthWords(vars int) int {
	bits := 1 << uint(vars)
	return (bits + 63) >> 6
}

func kCutTruthBit(words []uint64, assignment int) bool {
	return words[assignment>>6]&(uint64(1)<<uint(assignment&63)) != 0
}

func kCutSetTruthBit(words []uint64, assignment int) {
	words[assignment>>6] |= uint64(1) << uint(assignment&63)
}

func kCutVariableTruth() []uint64 { return []uint64{2} }

func (m *KCutManager) makeTrivial(node int) KCut {
	c := KCut{Leaves: []int{node}, Signature: kCutSignature([]int{node})}
	if m.opts.ComputeTruth {
		c.Truth = kCutVariableTruth()
	}
	return c
}

func (m *KCutManager) SetTrivial(node int) error {
	if err := m.checkNode(node); err != nil {
		return err
	}
	if len(m.cuts[node]) != 0 {
		return ErrKCutInvalid
	}
	m.cuts[node] = []KCut{m.makeTrivial(node)}
	m.stats.CutsCreated++
	m.updatePeak()
	return nil
}

func (m *KCutManager) FreeNode(node int) error {
	if err := m.checkNode(node); err != nil {
		return err
	}
	m.stats.CutsFreed += len(m.cuts[node])
	m.cuts[node] = nil
	m.best[node] = nil
	return nil
}

func (m *KCutManager) DropFanout(node int) error {
	if err := m.checkNode(node); err != nil {
		return err
	}
	if !m.opts.Drop || m.fanouts[node] <= 0 {
		return ErrKCutInvalid
	}
	m.fanouts[node]--
	if m.fanouts[node] == 0 {
		return m.FreeNode(node)
	}
	return nil
}

func mergeKCutLeaves(a, b []int, limit int) ([]int, bool) {
	res := make([]int, 0, min(limit, len(a)+len(b)))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		var v int
		switch {
		case j == len(b) || i < len(a) && a[i] < b[j]:
			v = a[i]
			i++
		case i == len(a) || b[j] < a[i]:
			v = b[j]
			j++
		default:
			v = a[i]
			i++
			j++
		}
		res = append(res, v)
		if len(res) > limit {
			return nil, false
		}
	}
	return res, true
}

// MergeKCuts is the checked, allocation-owning form of all five source merge
// variants. The variants differ only in their scratch strategy.
func MergeKCuts(a, b KCut, limit int) (KCut, bool) {
	if limit < 1 || !kCutLeavesValid(a.Leaves) || !kCutLeavesValid(b.Leaves) {
		return KCut{}, false
	}
	leaves, ok := mergeKCutLeaves(a.Leaves, b.Leaves, limit)
	if !ok {
		return KCut{}, false
	}
	return KCut{Leaves: leaves, Signature: kCutSignature(leaves)}, true
}

func kCutLeavesValid(leaves []int) bool {
	if len(leaves) == 0 {
		return false
	}
	for i, leaf := range leaves {
		if leaf < 0 || i > 0 && leaves[i-1] >= leaf {
			return false
		}
	}
	return true
}

func kCutSubset(a, b []int) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			i++
			j++
		} else if a[i] > b[j] {
			j++
		} else {
			return false
		}
	}
	return i == len(a)
}

func kCutLess(a, b KCut) bool {
	if len(a.Leaves) != len(b.Leaves) {
		return len(a.Leaves) < len(b.Leaves)
	}
	for i := range a.Leaves {
		if a.Leaves[i] != b.Leaves[i] {
			return a.Leaves[i] < b.Leaves[i]
		}
	}
	return false
}

func kCutEvalOnMerged(c KCut, merged []int, assignment int) bool {
	local := 0
	for i, leaf := range c.Leaves {
		pos := sort.SearchInts(merged, leaf)
		if assignment&(1<<uint(pos)) != 0 {
			local |= 1 << uint(i)
		}
	}
	return kCutTruthBit(c.Truth, local)
}

func computeKCutAND(dst *KCut, a, b KCut, complementA, complementB bool) error {
	if len(a.Truth) != kCutTruthWords(len(a.Leaves)) || len(b.Truth) != kCutTruthWords(len(b.Leaves)) {
		return ErrKCutInvalid
	}
	dst.Truth = make([]uint64, kCutTruthWords(len(dst.Leaves)))
	for assignment := 0; assignment < 1<<uint(len(dst.Leaves)); assignment++ {
		av := kCutEvalOnMerged(a, dst.Leaves, assignment) != complementA
		bv := kCutEvalOnMerged(b, dst.Leaves, assignment) != complementB
		value := av && bv
		if dst.Complement {
			value = !value
		}
		if value {
			kCutSetTruthBit(dst.Truth, assignment)
		}
	}
	return nil
}

func (m *KCutManager) acceptCut(dst []KCut, candidate KCut) ([]KCut, bool) {
	if m.opts.GlobalOnly && len(candidate.Leaves) > 1 {
		found := false
		for _, leaf := range candidate.Leaves {
			if m.global[leaf] {
				found = true
				break
			}
		}
		if !found {
			m.stats.CutsFiltered++
			return dst, false
		}
	}
	if !m.opts.Filter {
		return append(dst, candidate), true
	}
	for _, old := range dst {
		if old.Signature&candidate.Signature == old.Signature && kCutSubset(old.Leaves, candidate.Leaves) {
			m.stats.CutsFiltered++
			return dst, false
		}
	}
	write := 0
	for i := range dst {
		if candidate.Signature&dst[i].Signature == candidate.Signature && kCutSubset(candidate.Leaves, dst[i].Leaves) {
			m.stats.CutsFiltered++
			continue
		}
		dst[write] = dst[i]
		write++
	}
	dst = dst[:write]
	return append(dst, candidate), true
}

// ComputeAND computes the K-feasible cuts of one AND node. treeMask bit zero
// restricts the first fanin to its trivial cut; bit one does the same to the
// second fanin, matching Cut_NodeDoComputeCuts.
func (m *KCutManager) ComputeAND(node, node0, node1 int, complement0, complement1, includeTrivial bool, treeMask uint8) ([]KCut, error) {
	if err := m.checkNode(node); err != nil {
		return nil, err
	}
	if err := m.checkNode(node0); err != nil {
		return nil, err
	}
	if err := m.checkNode(node1); err != nil {
		return nil, err
	}
	if node == node0 || node == node1 || len(m.cuts[node0]) == 0 || len(m.cuts[node1]) == 0 {
		return nil, ErrKCutInvalid
	}
	left, right := m.cuts[node0], m.cuts[node1]
	if treeMask&1 != 0 {
		left = left[:1]
	}
	if treeMask&2 != 0 {
		right = right[:1]
	}
	statsBefore, recordCount := m.stats, len(m.records)
	rollback := func(err error) ([]KCut, error) {
		m.stats = statsBefore
		m.records = m.records[:recordCount]
		return nil, err
	}
	result := make([]KCut, 0, min(m.opts.KeepMax, len(left)*len(right)+1))
	if includeTrivial {
		result = append(result, m.makeTrivial(node))
		if len(result) == m.opts.KeepMax {
			m.stats.NodesLimited++
			goto done
		}
	}
	if m.opts.LocalOnly && treeMask != 0 {
		goto done
	}
	for i := range left {
		for j := range right {
			if err := m.check(); err != nil {
				return rollback(err)
			}
			candidate, ok := MergeKCuts(left[i], right[j], m.opts.MaxLeaves)
			if !ok || len(candidate.Leaves) == 1 {
				continue
			}
			candidate.Number0, candidate.Number1 = i, j
			candidate.Simulated = (left[i].Simulated != complement0) && (right[j].Simulated != complement1)
			if m.opts.ComputeTruth {
				if err := computeKCutAND(&candidate, left[i], right[j], complement0, complement1); err != nil {
					return rollback(err)
				}
			}
			var accepted bool
			result, accepted = m.acceptCut(result, candidate)
			if accepted && m.opts.Record {
				m.records = append(m.records, KCutRecord{Node: node, Parent0: i, Parent1: j})
			}
			if len(result) >= m.opts.KeepMax {
				m.stats.NodesLimited++
				goto done
			}
		}
	}
done:
	sort.SliceStable(result, func(i, j int) bool { return kCutLess(result[i], result[j]) })
	if len(result) == 0 {
		m.stats.NodesWithoutCuts++
	}
	oldCount := len(m.cuts[node])
	m.cuts[node] = cloneKCuts(result)
	m.stats.CutsFreed += oldCount
	m.stats.CutsCreated += len(result)
	m.stats.Nodes++
	m.updatePeak()
	return cloneKCuts(result), nil
}

// UnionChoices combines equivalent-node cut sets, keeping the first node's
// trivial cut and filtering all nontrivial cuts as one deterministic list.
func (m *KCutManager) UnionChoices(nodes []int) ([]KCut, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, ErrKCutInvalid
	}
	for _, node := range nodes {
		if err := m.checkNode(node); err != nil || len(m.cuts[node]) == 0 {
			if err != nil {
				return nil, err
			}
			return nil, ErrKCutInvalid
		}
	}
	statsBefore := m.stats
	rollback := func(err error) ([]KCut, error) {
		m.stats = statsBefore
		return nil, err
	}
	result := []KCut{kCutClone(m.cuts[nodes[0]][0])}
	topSim := result[0].Simulated
	for _, node := range nodes {
		for i := 1; i < len(m.cuts[node]); i++ {
			if err := m.check(); err != nil {
				return rollback(err)
			}
			candidate := kCutClone(m.cuts[node][i])
			candidate.Complement = topSim != candidate.Simulated
			result, _ = m.acceptCut(result, candidate)
			if len(result) >= m.opts.KeepMax {
				m.stats.NodesLimited++
				goto done
			}
		}
	}
done:
	sort.SliceStable(result, func(i, j int) bool { return kCutLess(result[i], result[j]) })
	root := nodes[0]
	m.cuts[root] = cloneKCuts(result)
	for _, node := range nodes[1:] {
		m.stats.CutsFreed += len(m.cuts[node])
		m.cuts[node] = nil
	}
	return cloneKCuts(result), nil
}

func (m *KCutManager) updatePeak() {
	total := 0
	for i := range m.cuts {
		total += len(m.cuts[i])
	}
	if total > m.stats.PeakStoredCuts {
		m.stats.PeakStoredCuts = total
	}
}

// VerifyCuts checks sorting, signatures, truth dimensions and the dominance
// antichain promised by filtered cut lists.
func VerifyKCuts(cuts []KCut, maxLeaves int, truth bool) error {
	for i := range cuts {
		if !kCutLeavesValid(cuts[i].Leaves) || len(cuts[i].Leaves) > maxLeaves || cuts[i].Signature != kCutSignature(cuts[i].Leaves) {
			return ErrKCutInvalid
		}
		if truth && len(cuts[i].Truth) != kCutTruthWords(len(cuts[i].Leaves)) {
			return ErrKCutInvalid
		}
		if i > 0 && kCutLess(cuts[i], cuts[i-1]) {
			return ErrKCutInvalid
		}
		for j := 0; j < i; j++ {
			if kCutSubset(cuts[j].Leaves, cuts[i].Leaves) {
				return ErrKCutInvalid
			}
		}
	}
	return nil
}

func FormatKCut(c KCut, sequential bool) (string, error) {
	if !kCutLeavesValid(c.Leaves) {
		return "", ErrKCutInvalid
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d : {", len(c.Leaves))
	for _, leaf := range c.Leaves {
		if sequential {
			fmt.Fprintf(&b, " %d", leaf>>8)
			if latch := leaf & 0xff; latch != 0 {
				fmt.Fprintf(&b, "(%d)", latch)
			}
		} else {
			fmt.Fprintf(&b, " %d", leaf)
		}
	}
	b.WriteString(" }")
	return b.String(), nil
}

// ComposeKCutTruth substitutes tree for node in factor and evaluates the
// result over resultLeaves without mutating either input cut.
func ComposeKCutTruth(factor KCut, node int, tree KCut, resultLeaves []int) (KCut, error) {
	if !kCutLeavesValid(factor.Leaves) || !kCutLeavesValid(tree.Leaves) || !kCutLeavesValid(resultLeaves) ||
		len(factor.Leaves) > KCutMaxLeaves || len(tree.Leaves) > KCutMaxLeaves || len(resultLeaves) > KCutMaxLeaves ||
		len(factor.Truth) != kCutTruthWords(len(factor.Leaves)) || len(tree.Truth) != kCutTruthWords(len(tree.Leaves)) {
		return KCut{}, ErrKCutInvalid
	}
	nodePos := sort.SearchInts(factor.Leaves, node)
	if nodePos == len(factor.Leaves) || factor.Leaves[nodePos] != node {
		return KCut{}, ErrKCutInvalid
	}
	for _, leaf := range factor.Leaves {
		if leaf != node && !containsSortedInt(resultLeaves, leaf) {
			return KCut{}, ErrKCutInvalid
		}
	}
	for _, leaf := range tree.Leaves {
		if !containsSortedInt(resultLeaves, leaf) {
			return KCut{}, ErrKCutInvalid
		}
	}
	res := KCut{Leaves: append([]int(nil), resultLeaves...), Signature: kCutSignature(resultLeaves), Truth: make([]uint64, kCutTruthWords(len(resultLeaves)))}
	for assignment := 0; assignment < 1<<uint(len(resultLeaves)); assignment++ {
		treeValue := kCutEvalOnMerged(tree, resultLeaves, assignment)
		factorAssignment := 0
		for i, leaf := range factor.Leaves {
			value := treeValue
			if leaf != node {
				pos := sort.SearchInts(resultLeaves, leaf)
				value = assignment&(1<<uint(pos)) != 0
			}
			if value {
				factorAssignment |= 1 << uint(i)
			}
		}
		if kCutTruthBit(factor.Truth, factorAssignment) {
			kCutSetTruthBit(res.Truth, assignment)
		}
	}
	return res, nil
}

func containsSortedInt(values []int, value int) bool {
	i := sort.SearchInts(values, value)
	return i < len(values) && values[i] == value
}

// CanonicalizeKCutNPN computes the minimum direct/complemented truth over all
// input permutations and phases. The source fast table supports fewer than six
// configured variables; the Go form safely supports up to six actual leaves.
func CanonicalizeKCutNPN(c KCut) (direct, complemented KCutCanonical, err error) {
	n := len(c.Leaves)
	if n < 1 || n > 6 || len(c.Truth) != kCutTruthWords(n) {
		return KCutCanonical{}, KCutCanonical{}, ErrKCutInvalid
	}
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}
	first := true
	var visit func(int)
	visit = func(pos int) {
		if pos == n {
			for phase := 0; phase < 1<<uint(n); phase++ {
				var truth uint64
				for canonicalAssignment := 0; canonicalAssignment < 1<<uint(n); canonicalAssignment++ {
					originalAssignment := 0
					for canonicalVar, originalVar := range perm {
						bit := (canonicalAssignment >> uint(canonicalVar)) & 1
						bit ^= (phase >> uint(originalVar)) & 1
						originalAssignment |= bit << uint(originalVar)
					}
					if kCutTruthBit(c.Truth, originalAssignment) {
						truth |= uint64(1) << uint(canonicalAssignment)
					}
				}
				mask := ^uint64(0)
				if n < 6 {
					mask = uint64(1)<<uint(1<<uint(n)) - 1
				}
				negative := (^truth) & mask
				if first || truth < direct.Truth {
					direct = KCutCanonical{Truth: truth, Permutation: append([]int(nil), perm...), Phase: uint8(phase)}
				}
				if first || negative < complemented.Truth {
					complemented = KCutCanonical{Truth: negative, Permutation: append([]int(nil), perm...), Phase: uint8(phase), OutputNegated: true}
				}
				first = false
			}
			return
		}
		for i := pos; i < n; i++ {
			perm[pos], perm[i] = perm[i], perm[pos]
			visit(pos + 1)
			perm[pos], perm[i] = perm[i], perm[pos]
		}
	}
	visit(0)
	return direct, complemented, nil
}

// MapNode selects the minimum-delay nontrivial cut. It is deterministic on
// equal delays and records an owned best cut for MappingArea.
func (m *KCutManager) MapNode(node int) (int, error) {
	if err := m.checkNode(node); err != nil {
		return 0, err
	}
	bestDelay := int(^uint(0) >> 1)
	var best *KCut
	for i := range m.cuts[node] {
		if len(m.cuts[node][i].Leaves) == 1 {
			continue
		}
		delay := 0
		for _, leaf := range m.cuts[node][i].Leaves {
			if m.delays[leaf] > delay {
				delay = m.delays[leaf]
			}
		}
		delay++
		if delay < bestDelay {
			copyCut := kCutClone(m.cuts[node][i])
			best = &copyCut
			bestDelay = delay
		}
	}
	if best == nil {
		return 0, ErrKCutInvalid
	}
	m.delays[node] = bestDelay
	m.best[node] = best
	return bestDelay, nil
}

func (m *KCutManager) MappingArea(root int) (int, error) {
	if err := m.checkNode(root); err != nil {
		return 0, err
	}
	seen := make([]bool, len(m.best))
	var visit func(int) (int, error)
	visit = func(node int) (int, error) {
		if node < 0 || node >= len(m.best) {
			return 0, ErrKCutBounds
		}
		if seen[node] || m.best[node] == nil || len(m.best[node].Leaves) == 1 {
			return 0, nil
		}
		seen[node] = true
		area := 1
		for _, leaf := range m.best[node].Leaves {
			sub, err := visit(leaf)
			if err != nil {
				return 0, err
			}
			area += sub
		}
		return area, nil
	}
	return visit(root)
}
