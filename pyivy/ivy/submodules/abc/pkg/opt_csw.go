package abc

// Functional port of src/opt/csw/{csw.h,cswInt.h,cswCore.c,cswCut.c,
// cswMan.c,cswTable.c}; csw_.c has no active functions.

import (
	"math/bits"
	"slices"
)

type CswOptions struct {
	CutsMax    int
	LeavesMax  int
	MaxObjects int
	MaxWork    int
	MaxBytes   int
	Stop       func() bool
}

type CswStats struct {
	NodesTried, NodesRewritten                  int
	TrivialConstants, TrivialLiterals, Trivial2 int
	CutsStored, CutPairs, TruthAssignments      int
}

type CswResult struct {
	Graph *AigManager
	Stats CswStats
}

type cswCut struct {
	next   *cswCut
	cost   int
	sign   uint32
	node   AigID
	leaves []AigID
	truth  []uint64
}

type cswManager struct {
	source, result *AigManager
	options        CswOptions
	stats          CswStats
	equivalent     []AigLit
	cuts           [][]*cswCut
	references     []uint64
	table          []*cswCut
	work, bytes    int
}

func normalizeCswOptions(options CswOptions) (CswOptions, error) {
	if options.CutsMax < 0 || options.LeavesMax < 0 || options.MaxObjects < 0 || options.MaxWork < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.CutsMax == 0 {
		options.CutsMax = 8
	}
	if options.LeavesMax == 0 {
		options.LeavesMax = 6
	}
	if options.CutsMax < 2 || options.LeavesMax < 1 || options.LeavesMax > 16 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 24
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 28
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 1 << 28
	}
	return options, nil
}

func (manager *cswManager) step(amount int) error {
	if amount < 0 || manager.work > manager.options.MaxWork-amount {
		return ErrLimit
	}
	if manager.options.Stop != nil && manager.options.Stop() {
		return ErrCancelled
	}
	manager.work += amount
	return nil
}

func (manager *cswManager) reserveTruth(entries int) error {
	words := max(1, (entries+63)/64)
	if words < 0 {
		return ErrLimit
	}
	return manager.reserveBytes(words * 8)
}

func (manager *cswManager) reserveBytes(bytes int) error {
	if bytes < 0 || manager.bytes > manager.options.MaxBytes-bytes {
		return ErrLimit
	}
	manager.bytes += bytes
	return nil
}

func cswCutSign(id AigID) uint32 { return 1 << (uint32(id) & 31) }

func cswTruthGet(truth []uint64, assignment int) bool {
	return truth[assignment>>6]&(uint64(1)<<uint(assignment&63)) != 0
}

func cswTruthSet(truth []uint64, assignment int) {
	truth[assignment>>6] |= uint64(1) << uint(assignment&63)
}

func cswTruthEqual(first, second []uint64, variables int, opposite bool) bool {
	entries := 1 << uint(variables)
	for assignment := 0; assignment < entries; assignment++ {
		if cswTruthGet(first, assignment) == cswTruthGet(second, assignment) == opposite {
			return false
		}
	}
	return true
}

func cswTruthSupport(truth []uint64, variables int) uint32 {
	var support uint32
	entries := 1 << uint(variables)
	for variable := 0; variable < variables; variable++ {
		bit := 1 << uint(variable)
		for assignment := 0; assignment < entries; assignment++ {
			if assignment&bit == 0 && cswTruthGet(truth, assignment) != cswTruthGet(truth, assignment|bit) {
				support |= 1 << uint(variable)
				break
			}
		}
	}
	return support
}

func cswMergeLeaves(first, second []AigID, maximum int) ([]AigID, bool) {
	merged := make([]AigID, 0, min(maximum, len(first)+len(second)))
	left, right := 0, 0
	for left < len(first) || right < len(second) {
		var id AigID
		switch {
		case right == len(second) || left < len(first) && first[left] < second[right]:
			id = first[left]
			left++
		case left == len(first) || second[right] < first[left]:
			id = second[right]
			right++
		default:
			id = first[left]
			left++
			right++
		}
		if len(merged) == maximum {
			return nil, false
		}
		merged = append(merged, id)
	}
	return merged, true
}

func cswProjectAssignment(merged, leaves []AigID, assignment int) int {
	projected, leaf := 0, 0
	for position, id := range merged {
		if leaf < len(leaves) && id == leaves[leaf] {
			if assignment&(1<<uint(position)) != 0 {
				projected |= 1 << uint(leaf)
			}
			leaf++
		}
	}
	return projected
}

func (manager *cswManager) mergeTruth(leaves []AigID, first, second *cswCut, complementFirst, complementSecond bool) ([]uint64, error) {
	entries := 1 << uint(len(leaves))
	if err := manager.reserveTruth(entries); err != nil {
		return nil, err
	}
	truth := make([]uint64, max(1, (entries+63)/64))
	for assignment := 0; assignment < entries; assignment++ {
		if assignment&1023 == 0 {
			if err := manager.step(min(1024, entries-assignment)); err != nil {
				return nil, err
			}
		}
		left := cswTruthGet(first.truth, cswProjectAssignment(leaves, first.leaves, assignment)) != complementFirst
		right := cswTruthGet(second.truth, cswProjectAssignment(leaves, second.leaves, assignment)) != complementSecond
		if left && right {
			cswTruthSet(truth, assignment)
		}
	}
	manager.stats.TruthAssignments += entries
	return truth, nil
}

func cswDominates(first, second []AigID) bool {
	left, right := 0, 0
	for left < len(first) && right < len(second) {
		switch {
		case first[left] == second[right]:
			left++
			right++
		case first[left] > second[right]:
			right++
		default:
			return false
		}
	}
	return left == len(first)
}

func (manager *cswManager) filter(slots []*cswCut, candidate *cswCut) bool {
	for _, current := range slots {
		if current == candidate || len(current.leaves) < 2 {
			continue
		}
		if len(current.leaves) > len(candidate.leaves) {
			if current.sign&candidate.sign == candidate.sign && cswDominates(candidate.leaves, current.leaves) {
				current.leaves = nil
				current.truth = nil
			}
		} else if current.sign&candidate.sign == current.sign && cswDominates(current.leaves, candidate.leaves) {
			candidate.leaves = nil
			candidate.truth = nil
			return true
		}
	}
	return false
}

func cswFindFree(slots []*cswCut) *cswCut {
	var worst *cswCut
	for _, cut := range slots {
		if len(cut.leaves) == 0 {
			return cut
		}
		if worst == nil || worst.cost < cut.cost {
			worst = cut
		}
	}
	worst.leaves = nil
	worst.truth = nil
	return worst
}

func cswAllocateMergedCut(slots []*cswCut, first, second *cswCut, maximum int) (*cswCut, []AigID, bool) {
	// C's lossy signature test precedes Csw_CutFindFree. A pair rejected
	// here must not evict a stored cut.
	if bits.OnesCount32(first.sign|second.sign) > maximum {
		return nil, nil, false
	}
	// Csw_CutFindFree precedes the exact merge. Consequently an infeasible
	// merge can evict the current worst cut; later pairs observe this.
	candidate := cswFindFree(slots)
	leaves, okay := cswMergeLeaves(first.leaves, second.leaves, maximum)
	return candidate, leaves, okay
}

var cswHashPrimes = [...]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459, 1499, 1559, 1607, 1657, 1709, 1759,
	1823, 1877, 1933, 1997, 2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543, 2609, 2663,
}

func cswHash(cut *cswCut) uint32 {
	var hash uint32
	for index, id := range cut.leaves {
		hash ^= uint32(id) * cswHashPrimes[index]
	}
	return hash
}

func (manager *cswManager) insert(cut *cswCut) {
	bucket := cswHash(cut) % uint32(len(manager.table))
	cut.next = manager.table[bucket]
	manager.table[bucket] = cut
	manager.stats.CutsStored++
}

func (manager *cswManager) lookup(cut *cswCut) AigLit {
	bucket := cswHash(cut) % uint32(len(manager.table))
	for entry := manager.table[bucket]; entry != nil; entry = entry.next {
		if entry.sign != cut.sign || !slices.Equal(entry.leaves, cut.leaves) {
			continue
		}
		if cswTruthEqual(entry.truth, cut.truth, len(cut.leaves), false) {
			return AigLit(entry.node) << 1
		}
		if cswTruthEqual(entry.truth, cut.truth, len(cut.leaves), true) {
			return (AigLit(entry.node) << 1).Not()
		}
	}
	return InvalidAigLit
}

func (manager *cswManager) ensureObject(id AigID) error {
	if int(id) >= manager.options.MaxObjects {
		return ErrLimit
	}
	for len(manager.cuts) <= int(id) {
		manager.cuts = append(manager.cuts, nil)
		manager.references = append(manager.references, 0)
	}
	return nil
}

func (manager *cswManager) prepare(id AigID, trivial bool) error {
	if err := manager.ensureObject(id); err != nil {
		return err
	}
	if manager.options.CutsMax > int(^uint(0)>>1)/64 || manager.reserveBytes(manager.options.CutsMax*64) != nil {
		return ErrLimit
	}
	slots := make([]*cswCut, manager.options.CutsMax)
	for index := range slots {
		slots[index] = &cswCut{node: id}
	}
	manager.cuts[id] = slots
	if trivial {
		if err := manager.reserveTruth(2); err != nil {
			return err
		}
		slots[0].leaves = []AigID{id}
		slots[0].sign = cswCutSign(id)
		slots[0].truth = []uint64{2}
	}
	return nil
}

func (manager *cswManager) cutCost(cut *cswCut) (int, error) {
	var cost uint64
	for _, leaf := range cut.leaves {
		if int(leaf) >= len(manager.references) || cost > uint64(^uint(0)>>1)-manager.references[leaf] {
			return 0, ErrLimit
		}
		cost += manager.references[leaf]
	}
	return int(cost) * 100 / len(cut.leaves), nil
}

// cutCost2 is the unused alternate cost metric compiled in cswCut.c.
func (manager *cswManager) cutCost2(cut *cswCut) (float32, error) {
	if len(cut.leaves) == 0 {
		return 0, ErrDimension
	}
	var reciprocal float32
	for _, leaf := range cut.leaves {
		if int(leaf) >= len(manager.result.objects) || manager.result.objects[leaf].RefCount == 0 {
			return 0, ErrInvariant
		}
		reciprocal += 1 / float32(manager.result.objects[leaf].RefCount)
	}
	return 1 / reciprocal, nil
}

func (manager *cswManager) tableCount() int {
	count := 0
	for _, head := range manager.table {
		for entry := head; entry != nil; entry = entry.next {
			count++
		}
	}
	return count
}

func cswMinimize(cut *cswCut) {
	support := cswTruthSupport(cut.truth, len(cut.leaves))
	if support == uint32(1<<uint(len(cut.leaves)))-1 {
		return
	}
	leaves := make([]AigID, 0, len(cut.leaves))
	positions := make([]int, 0, len(cut.leaves))
	for index, leaf := range cut.leaves {
		if support&(1<<uint(index)) != 0 {
			leaves = append(leaves, leaf)
			positions = append(positions, index)
		}
	}
	truth := make([]uint64, max(1, (1<<uint(len(leaves))+63)/64))
	for assignment := 0; assignment < 1<<uint(len(leaves)); assignment++ {
		old := 0
		for index, position := range positions {
			if assignment&(1<<uint(index)) != 0 {
				old |= 1 << uint(position)
			}
		}
		if cswTruthGet(cut.truth, old) {
			cswTruthSet(truth, assignment)
		}
	}
	cut.leaves, cut.truth = leaves, truth
}

func (manager *cswManager) twoVariable(cut *cswCut) (AigLit, error) {
	cswMinimize(cut)
	if len(cut.leaves) != 2 {
		return InvalidAigLit, ErrInvariant
	}
	truth := 0
	for assignment := 0; assignment < 4; assignment++ {
		if cswTruthGet(cut.truth, assignment) {
			truth |= 1 << uint(assignment)
		}
	}
	complement := truth == 14 || truth == 13 || truth == 11 || truth == 7
	if complement {
		truth = (^truth) & 15
	}
	left, right := AigLit(cut.leaves[0])<<1, AigLit(cut.leaves[1])<<1
	var first, second AigLit
	switch truth {
	case 1:
		first, second = left.Not(), right.Not()
	case 2:
		first, second = left, right.Not()
	case 4:
		first, second = left.Not(), right
	case 8:
		first, second = left, right
	default:
		return InvalidAigLit, nil
	}
	result, err := manager.result.And(first, second)
	if err != nil {
		return InvalidAigLit, err
	}
	if err = manager.ensureObject(result.ID()); err != nil {
		return InvalidAigLit, err
	}
	return result.NotIf(complement), nil
}

func (manager *cswManager) sweepObject(id AigID, trivial bool) (AigLit, error) {
	if int(id) >= len(manager.result.objects) || manager.result.objects[id].Type != AigAND {
		return AigLit(id) << 1, nil
	}
	if int(id) < len(manager.cuts) && manager.cuts[id] != nil {
		return AigLit(id) << 1, nil
	}
	if err := manager.prepare(id, trivial); err != nil {
		return InvalidAigLit, err
	}
	object := manager.result.objects[id]
	firstSlots, secondSlots := manager.cuts[object.Fanin0.ID()], manager.cuts[object.Fanin1.ID()]
	if firstSlots == nil || secondSlots == nil {
		return InvalidAigLit, ErrInvariant
	}
	for _, first := range firstSlots {
		if len(first.leaves) == 0 {
			continue
		}
		for _, second := range secondSlots {
			if len(second.leaves) == 0 {
				continue
			}
			if err := manager.step(1); err != nil {
				return InvalidAigLit, err
			}
			manager.stats.CutPairs++
			candidate, leaves, okay := cswAllocateMergedCut(manager.cuts[id], first, second, manager.options.LeavesMax)
			if !okay {
				continue
			}
			candidate.node = id
			candidate.leaves = leaves
			candidate.sign = first.sign | second.sign
			if manager.filter(manager.cuts[id], candidate) {
				continue
			}
			truth, err := manager.mergeTruth(leaves, first, second, object.Fanin0.Complemented(), object.Fanin1.Complemented())
			if err != nil {
				return InvalidAigLit, err
			}
			candidate.truth = truth
			support := cswTruthSupport(truth, len(leaves))
			variables := 0
			for value := support; value != 0; value &= value - 1 {
				variables++
			}
			switch variables {
			case 0:
				manager.stats.TrivialConstants++
				return AigConst1.NotIf(!cswTruthGet(truth, 0)), nil
			case 1:
				manager.stats.TrivialLiterals++
				position := 0
				for support&(1<<uint(position)) == 0 {
					position++
				}
				return (AigLit(leaves[position]) << 1).NotIf(cswTruthGet(truth, 0)), nil
			case 2:
				if len(leaves) > 2 {
					replacement, err := manager.twoVariable(candidate)
					if err != nil {
						return InvalidAigLit, err
					}
					if replacement != InvalidAigLit {
						manager.stats.Trivial2++
						return replacement, nil
					}
				}
			}
			if len(candidate.leaves) > 2 {
				if replacement := manager.lookup(candidate); replacement != InvalidAigLit {
					manager.stats.NodesRewritten++
					return replacement, nil
				}
			}
			candidate.cost, err = manager.cutCost(candidate)
			if err != nil {
				return InvalidAigLit, err
			}
			if candidate.cost <= 0 {
				candidate.cost = 1
			}
		}
	}
	manager.stats.NodesTried++
	for _, cut := range manager.cuts[id] {
		if len(cut.leaves) > 2 {
			manager.insert(cut)
		}
	}
	return AigLit(id) << 1, nil
}

func CswSweep(graph *AigManager, options CswOptions) (CswResult, error) {
	fail := func(err error) (CswResult, error) { return CswResult{}, err }
	options, err := normalizeCswOptions(options)
	if err != nil {
		return fail(err)
	}
	if graph == nil {
		return fail(ErrDimension)
	}
	if err = graph.Check(); err != nil {
		return fail(err)
	}
	if graph.ObjectCount() > options.MaxObjects {
		return fail(ErrLimit)
	}
	result := NewAigManager()
	result.Name, result.Spec = graph.Name, graph.Spec
	manager := &cswManager{source: graph, result: result, options: options, equivalent: make([]AigLit, graph.ObjectCount())}
	for index := range manager.equivalent {
		manager.equivalent[index] = InvalidAigLit
	}
	if graph.AndCount() > int(^uint(0)>>1)/options.CutsMax {
		return fail(ErrLimit)
	}
	tableTarget := max(1, graph.AndCount()*options.CutsMax/2)
	tableSize, primeErr := hshPrimeAtLeast(tableTarget, int(^uint(0)>>1), options.Stop)
	if primeErr != nil {
		return fail(primeErr)
	}
	if tableSize > int(^uint(0)>>1)/8 || manager.reserveBytes(tableSize*8) != nil {
		return fail(ErrLimit)
	}
	manager.table = make([]*cswCut, tableSize)
	manager.equivalent[0] = AigConst1
	for _, sourceID := range graph.cis {
		if err = manager.step(1); err != nil {
			return fail(err)
		}
		literal, createErr := result.CreateCI()
		if createErr != nil {
			return fail(createErr)
		}
		manager.equivalent[sourceID] = literal
		if err = manager.prepare(literal.ID(), true); err != nil {
			return fail(err)
		}
		manager.references[literal.ID()] += uint64(graph.objects[sourceID].RefCount)
	}
	for sourceID, object := range graph.objects {
		if err = manager.step(1); err != nil {
			return fail(err)
		}
		if object.Type == AigNone || object.Type == AigConstant1 || object.Type == AigCI || object.Type == AigCO {
			continue
		}
		if object.Type != AigAND || manager.equivalent[object.Fanin0.ID()] == InvalidAigLit || manager.equivalent[object.Fanin1.ID()] == InvalidAigLit {
			return fail(ErrUnsupported)
		}
		left := manager.equivalent[object.Fanin0.ID()].NotIf(object.Fanin0.Complemented())
		right := manager.equivalent[object.Fanin1.ID()].NotIf(object.Fanin1.Complemented())
		created, createErr := result.And(left, right)
		if createErr != nil {
			return fail(createErr)
		}
		if err = manager.ensureObject(created.ID()); err != nil {
			return fail(err)
		}
		replacement := created
		for {
			replacement, err = manager.sweepObject(replacement.ID(), object.RefCount > 1)
			if err != nil {
				return fail(err)
			}
			replacement = replacement.NotIf(created.Complemented())
			created = replacement
			if replacement.ID() == 0 || manager.cuts[replacement.ID()] != nil {
				break
			}
		}
		manager.equivalent[sourceID] = replacement
		if manager.references[replacement.ID()] > ^uint64(0)-uint64(object.RefCount) {
			return fail(ErrLimit)
		}
		manager.references[replacement.ID()] += uint64(object.RefCount)
	}
	for _, outputID := range graph.cos {
		object := graph.objects[outputID]
		if manager.equivalent[object.Fanin0.ID()] == InvalidAigLit {
			return fail(ErrInvariant)
		}
		driver := manager.equivalent[object.Fanin0.ID()].NotIf(object.Fanin0.Complemented())
		if _, err = result.CreateCO(driver); err != nil {
			return fail(err)
		}
	}
	if err = result.SetRegisterCount(graph.registers); err != nil {
		return fail(err)
	}
	if err = result.SetConstraintCount(graph.constraints); err != nil {
		return fail(err)
	}
	result.Cleanup()
	if err = result.Check(); err != nil {
		return fail(err)
	}
	return CswResult{Graph: result, Stats: manager.stats}, nil
}
