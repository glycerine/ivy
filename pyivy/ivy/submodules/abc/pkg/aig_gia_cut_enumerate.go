package abc

// Batch cut computation from aig/gia/giaCut.c: Gia_CutMergeOrder, dominance
// checks, cost insertion, truth composition and Gia_StoComputeCuts helpers.

import "math/bits"

type GiaCut struct {
	Leaves     []GiaID // strictly increasing
	Signature  uint64
	Function   int // interned truth ID*2 + output phase; -1 when not computed
	TreeLeaves int // leaves referenced once in the complete source graph
}
type GiaCutStatistics struct {
	Pairs, Merges, Evaluations, Cuts, SaturatedNodes uint64
}
type GiaCutEnumerationOptions struct {
	CutSize, CutLimit                   int  // zero: 6 and 25; supported ranges 2..8 and 2..50
	ComputeTruth, MinimizeSupport       bool // support reduction requires truth computation
	MaxObjects, MaxStoredCuts, MaxPairs int  // zero: one million each
	MaxTruthEntries, MaxTruthWords      int  // zero: one million; allocated paged word budget
	Stop                                func() bool
}

// GiaCutEnumeration owns per-object cut lists in native ranked order, followed
// by a unit cut when C retains one. CO lists are empty. Truth IDs are stable
// insertion IDs, including truths of candidates later dropped by ranking.
// Composed functional cuts can have correlated internal leaf signals:
// stored truth matches all realizable source assignments, and can
// differ from treating a structural boundary's leaves as independent inputs.
// Source graph and its scratch remain unchanged. Mutation of returned cut
// slices cannot affect the graph; TruthTable returns independent truth storage.
type GiaCutEnumeration struct {
	CutSize, CutLimit int
	Cuts              [][]GiaCut
	Statistics        GiaCutStatistics
	truths            *WordMemory
}

func (r *GiaCutEnumeration) TruthCount() int {
	if r == nil || r.truths == nil {
		return 0
	}
	return r.truths.Len()
}
func (r *GiaCutEnumeration) TruthTable(cut GiaCut) (*TruthTable, error) {
	if r == nil || r.truths == nil {
		return nil, ErrInvariant
	}
	if r.CutSize < 2 || r.CutSize > 8 || cut.Function < 0 || len(cut.Leaves) > r.CutSize {
		return nil, ErrDimension
	}
	width, e := TruthTableWordCount(r.CutSize)
	if e != nil || width != r.truths.EntryWords() {
		return nil, ErrDimension
	}
	words, e := r.truths.Read(cut.Function / 2)
	if e != nil {
		return nil, e
	}
	count, e := TruthTableWordCount(len(cut.Leaves))
	if e != nil {
		return nil, e
	}
	if cut.Function&1 != 0 {
		for i := range words {
			words[i] = ^words[i]
		}
	}
	return TruthTableFromWords(len(cut.Leaves), words[:count], 4)
}
func giaCutSignature(leaves []GiaID) uint64 {
	var sign uint64
	for _, id := range leaves {
		sign |= uint64(1) << uint(id&63)
	}
	return sign
}
func giaCutContains(base, subset []GiaID) bool {
	i := 0
	for _, value := range subset {
		for i < len(base) && base[i] < value {
			i++
		}
		if i == len(base) || base[i] != value {
			return false
		}
		i++
	}
	return true
}
func giaCutMerge(a, b []GiaID, limit int) ([]GiaID, bool) {
	merged := make([]GiaID, 0, min(limit, len(a)+len(b)))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		if len(merged) == limit {
			return nil, false
		}
		if j == len(b) || i < len(a) && a[i] < b[j] {
			merged = append(merged, a[i])
			i++
		} else if i == len(a) || b[j] < a[i] {
			merged = append(merged, b[j])
			j++
		} else {
			merged = append(merged, a[i])
			i++
			j++
		}
	}
	return merged, true
}
func giaCutCompare(a, b GiaCut) int {
	if a.TreeLeaves < b.TreeLeaves {
		return -1
	}
	if a.TreeLeaves > b.TreeLeaves {
		return 1
	}
	if len(a.Leaves) < len(b.Leaves) {
		return -1
	}
	if len(a.Leaves) > len(b.Leaves) {
		return 1
	}
	return 0
}
func giaCutAdd(cuts []GiaCut, cut GiaCut, limit int) []GiaCut {
	kept := cuts[:0]
	for _, old := range cuts {
		if len(cut.Leaves) < len(old.Leaves) && cut.Signature&old.Signature == cut.Signature && giaCutContains(old.Leaves, cut.Leaves) {
			continue
		}
		kept = append(kept, old)
	}
	kept = append(kept, cut)
	// Native ties insert the newest cut before equal-cost existing cuts.
	for i := len(kept) - 1; i > 0; i-- {
		if giaCutCompare(kept[i-1], kept[i]) < 0 {
			break
		}
		kept[i-1], kept[i] = kept[i], kept[i-1]
	}
	if len(kept) > limit {
		kept = kept[:limit]
	}
	return kept
}
func newGiaCutTruthMemory(size, entries, words int) (*WordMemory, error) {
	if entries == 0 {
		entries = 1 << 20
	}
	if words == 0 {
		words = 1 << 20
	}
	width, e := TruthTableWordCount(size)
	if e != nil {
		return nil, e
	}
	if width > words {
		return nil, ErrLimit
	}
	pageLog := 0
	for pageLog < 12 && (1<<uint(pageLog+1)) <= entries && (1<<uint(pageLog+1)) <= words/width {
		pageLog++
	}
	memory, e := NewWordMemory(width, pageLog, WordMemoryOptions{MaxEntries: entries, MaxWords: words})
	if e != nil {
		return nil, e
	}
	if e = memory.EnableHash(); e != nil {
		return nil, e
	}
	entry := make([]uint64, width)
	if _, _, e = memory.Intern(entry); e != nil {
		return nil, e
	}
	for i := range entry {
		entry[i] = truthVariablePattern(0)
	}
	if _, _, e = memory.Intern(entry); e != nil {
		return nil, e
	}
	return memory, nil
}
func (r *GiaCutEnumeration) composeCut(a, b GiaCut, complementA, complementB, xor, minimize bool, leaves []GiaID) (GiaCut, bool, error) {
	left, e := r.TruthTable(a)
	if e != nil {
		return GiaCut{}, false, e
	}
	right, e := r.TruthTable(b)
	if e != nil {
		return GiaCut{}, false, e
	}
	expand := func(table *TruthTable, source []GiaID) (*TruthTable, error) {
		order := make([]int, len(source))
		j := 0
		for i, id := range source {
			for leaves[j] != id {
				j++
			}
			order[i] = j
		}
		return table.ExpandVariables(len(leaves), order)
	}
	left, e = expand(left, a.Leaves)
	if e != nil {
		return GiaCut{}, false, e
	}
	right, e = expand(right, b.Leaves)
	if e != nil {
		return GiaCut{}, false, e
	}
	if complementA {
		left = left.Not()
	}
	if complementB {
		right = right.Not()
	}
	var truth *TruthTable
	if xor {
		truth, e = left.Xor(right, false)
	} else {
		truth, e = left.And(right, false, false)
	}
	if e != nil {
		return GiaCut{}, false, e
	}
	phase := int(truth.words[0] & 1)
	if phase != 0 {
		truth = truth.Not()
	}
	oldSize := len(leaves)
	if minimize {
		compact, e := truth.CompactSupport()
		if e != nil {
			return GiaCut{}, false, e
		}
		reduced := make([]GiaID, len(compact.SourceVariables))
		for i, v := range compact.SourceVariables {
			reduced[i] = leaves[v]
		}
		leaves = reduced
		truth = compact.Table
	}
	stretched, e := truth.Stretch(r.CutSize)
	if e != nil {
		return GiaCut{}, false, e
	}
	id, _, e := r.truths.Intern(stretched.words)
	if e != nil {
		return GiaCut{}, false, e
	}
	function, e := checkedSize(phase, 2, id, int(^uint(0)>>1))
	if e != nil {
		return GiaCut{}, false, e
	}
	return GiaCut{Leaves: leaves, Signature: giaCutSignature(leaves), Function: function}, len(leaves) < oldSize, nil
}

// newGiaCutStore prepares C's batch path for ordinary AND/native XOR graphs.
// Buffers/native muxes reject as in C's merge preconditions. Sorted union,
// signature filtering, pre-evaluation dominance, support compaction, stable
// truth IDs, cost ordering and unit-cut policy retain native behavior. The
// complete source graph determines leaf reference costs. Limits/cancellation
// return no partial enumeration; optional truth data remains owned by result.
func newGiaCutStore(m *GiaManager, options GiaCutEnumerationOptions) (*GiaCutStore, error) {
	if m == nil {
		return nil, ErrDimension
	}
	if options.CutSize == 0 {
		options.CutSize = 6
	}
	if options.CutLimit == 0 {
		options.CutLimit = 25
	}
	if options.CutSize < 2 || options.CutSize > 8 || options.CutLimit < 2 || options.CutLimit > 50 || options.MaxObjects < 0 || options.MaxStoredCuts < 0 || options.MaxPairs < 0 || options.MaxTruthEntries < 0 || options.MaxTruthWords < 0 {
		return nil, ErrDimension
	}
	maxObjects, maxStored, maxPairs := options.MaxObjects, options.MaxStoredCuts, options.MaxPairs
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if maxStored == 0 {
		maxStored = 1 << 20
	}
	if maxPairs == 0 {
		maxPairs = 1 << 20
	}
	if len(m.objects) > maxObjects {
		return nil, ErrLimit
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	references := make([]uint64, len(m.objects))
	for id, object := range m.objects {
		if object.IsBuffer() {
			return nil, ErrUnsupported
		}
		if _, mux := m.MuxControl(GiaID(id)); mux {
			return nil, ErrUnsupported
		}
		if object.IsAnd() || object.IsCO() {
			a, _ := object.Fanin0(GiaID(id))
			references[a.ID()]++
		}
		if object.IsAnd() {
			b, _ := object.Fanin1(GiaID(id))
			references[b.ID()]++
		}
	}
	result := &GiaCutEnumeration{CutSize: options.CutSize, CutLimit: options.CutLimit, Cuts: make([][]GiaCut, len(m.objects))}
	var e error
	if options.ComputeTruth {
		result.truths, e = newGiaCutTruthMemory(options.CutSize, options.MaxTruthEntries, options.MaxTruthWords)
		if e != nil {
			return nil, e
		}
	}
	store := &GiaCutStore{source: m, options: options, result: result, references: references, maxStored: maxStored, maxPairs: maxPairs, next: 1}
	if e = store.store(0, []GiaCut{{Leaves: []GiaID{}, Function: 0}}); e != nil {
		return nil, e
	}
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		if e = store.store(id, []GiaCut{store.unit(id)}); e != nil {
			return nil, e
		}
	}
	return store, nil
}

// GiaCutStore computes one AND/XOR node per Advance call. It owns a fixed
// graph snapshot, complete reference counts, cut lists and stable truth IDs.
// C's Gia_StoAlloc/RefObj/ComputeCuts* lifecycle is represented by construction,
// Advance and Finish. A store is single-owner and must not be used concurrently.
// Any Advance failure closes the store; no partially interned state can escape.
// Finish transfers the completed result and closes the store.
type GiaCutStore struct {
	source                            *GiaManager
	options                           GiaCutEnumerationOptions
	result                            *GiaCutEnumeration
	references                        []uint64
	stored, maxStored, maxPairs, next int
	failure                           error
}

// NewGiaCutStore snapshots a validated graph. Later source edits do not alter
// the reference costs, cuts or truth IDs of this computation. All CI and
// constant cuts are initialized in native batch order before any gate cuts.
func NewGiaCutStore(source *GiaManager, options GiaCutEnumerationOptions) (*GiaCutStore, error) {
	s, err := newGiaCutStore(source, options)
	if err != nil {
		return nil, err
	}
	s.source = source.Clone()
	return s, nil
}

func (s *GiaCutStore) store(id GiaID, cuts []GiaCut) error {
	count, err := checkedSize(s.stored, 1, len(cuts), s.maxStored)
	if err != nil {
		return err
	}
	s.stored = count
	s.result.Cuts[id] = cuts
	return nil
}
func (s *GiaCutStore) unit(id GiaID) GiaCut {
	tree := 0
	if s.references[id] == 1 {
		tree = 1
	}
	return GiaCut{Leaves: []GiaID{id}, Signature: uint64(1) << uint(id&63), Function: 2, TreeLeaves: tree}
}
func (s *GiaCutStore) fail(err error) (bool, error) {
	s.failure = err
	s.source, s.result, s.references = nil, nil, nil
	s.options = GiaCutEnumerationOptions{}
	return false, err
}

// Close releases an unfinished store without producing a partial result.
// It is idempotent, including on a nil store or after Finish.
func (s *GiaCutStore) Close() {
	if s != nil && s.result != nil {
		s.fail(ErrInvariant)
	}
}

// Advance computes the next gate in object order. done is true when every
// gate has been processed. Repeated completed calls are harmless until Finish.
// Cancellation is checked even for an empty or already completed graph.
func (s *GiaCutStore) Advance() (done bool, err error) {
	if s == nil {
		return false, ErrDimension
	}
	if s.failure != nil {
		return false, s.failure
	}
	if s.result == nil {
		return false, ErrInvariant
	}
	if s.options.Stop != nil && s.options.Stop() {
		return s.fail(ErrCancelled)
	}
	for s.next < len(s.source.objects) && !s.source.objects[s.next].IsAnd() {
		s.next++
	}
	if s.next == len(s.source.objects) {
		return true, nil
	}
	if err := s.computeNode(s.next); err != nil {
		return s.fail(err)
	}
	s.next++
	for s.next < len(s.source.objects) && !s.source.objects[s.next].IsAnd() {
		s.next++
	}
	if s.options.Stop != nil && s.options.Stop() {
		return s.fail(ErrCancelled)
	}
	return s.next == len(s.source.objects), nil
}

// Finish computes any remaining gates and transfers the owned result once.
func (s *GiaCutStore) Finish() (*GiaCutEnumeration, error) {
	for {
		done, err := s.Advance()
		if err != nil {
			return nil, err
		}
		if done {
			break
		}
	}
	result := s.result
	s.source, s.result, s.references = nil, nil, nil
	s.options = GiaCutEnumerationOptions{}
	return result, nil
}

// EnumerateCuts computes the complete native batch path. The internal store
// borrows the graph only for this synchronous call; it never changes it.
func (m *GiaManager) EnumerateCuts(options GiaCutEnumerationOptions) (*GiaCutEnumeration, error) {
	s, err := newGiaCutStore(m, options)
	if err != nil {
		return nil, err
	}
	return s.Finish()
}

func (s *GiaCutStore) computeNode(id int) error {
	m, options, result, references, maxPairs := s.source, s.options, s.result, s.references, s.maxPairs
	object := m.objects[id]
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	var e error
	first, _ := object.Fanin0(GiaID(id))
	second, _ := object.Fanin1(GiaID(id))
	cuts := []GiaCut{}
	for _, a := range result.Cuts[first.ID()] {
		for _, b := range result.Cuts[second.ID()] {
			if stopped() {
				return ErrCancelled
			}
			if result.Statistics.Pairs >= uint64(maxPairs) {
				return ErrLimit
			}
			result.Statistics.Pairs++
			if len(a.Leaves)+len(b.Leaves) > options.CutSize && bits.OnesCount64(a.Signature|b.Signature) > options.CutSize {
				continue
			}
			result.Statistics.Merges++
			leaves, ok := giaCutMerge(a.Leaves, b.Leaves, options.CutSize)
			if !ok {
				continue
			}
			candidate := GiaCut{Leaves: leaves, Signature: a.Signature | b.Signature, Function: -1}
			contained := false
			for _, old := range cuts {
				if len(old.Leaves) <= len(leaves) && old.Signature&candidate.Signature == old.Signature && giaCutContains(leaves, old.Leaves) {
					contained = true
					break
				}
			}
			if contained {
				continue
			}
			result.Statistics.Evaluations++
			if options.ComputeTruth {
				candidate, _, e = result.composeCut(a, b, first.Complemented(), second.Complemented(), object.IsXor(), options.MinimizeSupport, leaves)
				if e != nil {
					return e
				}
			}
			for _, leaf := range candidate.Leaves {
				if references[leaf] == 1 {
					candidate.TreeLeaves++
				}
			}
			cuts = giaCutAdd(cuts, candidate, options.CutLimit-1)
		}
	}
	if len(cuts) == 0 {
		return ErrInvariant
	}
	result.Statistics.Cuts += uint64(len(cuts))
	if len(cuts) == options.CutLimit-1 {
		result.Statistics.SaturatedNodes++
	}
	if len(cuts) > 1 || len(cuts[0].Leaves) > 1 {
		cuts = append(cuts, s.unit(GiaID(id)))
	}
	if e = s.store(GiaID(id), cuts); e != nil {
		return e
	}
	return nil
}
