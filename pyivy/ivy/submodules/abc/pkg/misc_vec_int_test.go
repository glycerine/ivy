package abc

// Go-only tests for src/misc/vec/vecInt.h, vec.h, and module.make.

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func requireIntEntries(t testing.TB, vector *IntVector, want []int) {
	t.Helper()
	got, err := vector.Entries()
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("entries %v want %v: %v", got, want, err)
	}
}

func TestIntVectorConstructionCapacityOwnershipAndDormantStorage(t *testing.T) {
	var zeroValue IntVector
	if err := zeroValue.Push(3); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, &zeroValue, []int{3})

	ordinary, err := NewIntVector(1, IntVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if capacity, _ := ordinary.Capacity(); capacity != 16 || ordinary.MemoryBytes() != 80 {
		t.Fatal(capacity, ordinary.MemoryBytes())
	}
	exact, _ := NewExactIntVector(1, IntVectorLimits{})
	if capacity, _ := exact.Capacity(); capacity != 1 {
		t.Fatal(capacity)
	}
	zero, _ := NewZeroIntVector(3, IntVectorLimits{})
	requireIntEntries(t, zero, []int{0, 0, 0})
	full, _ := NewFullIntVector(3, IntVectorLimits{})
	requireIntEntries(t, full, []int{-1, -1, -1})
	ranged, _ := NewRangeIntVector(-2, 5, IntVectorLimits{})
	requireIntEntries(t, ranged, []int{-2, -1, 0, 1, 2})
	natural, _ := NewNaturalIntVector(4, IntVectorLimits{})
	requireIntEntries(t, natural, []int{0, 1, 2, 3})
	input := []int{4, 5, 6}
	owned, _ := NewIntVectorFrom(input, IntVectorLimits{})
	input[0] = 99
	requireIntEntries(t, owned, []int{4, 5, 6})
	clone, err := owned.Clone()
	if err != nil {
		t.Fatal(err)
	}
	_ = owned.Write(0, 8)
	requireIntEntries(t, clone, []int{4, 5, 6})
	if err = owned.Shrink(1); err != nil {
		t.Fatal(err)
	}
	if err = owned.Resize(3); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, owned, []int{8, 5, 6})
	moved, err := owned.Move()
	if err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, moved, []int{8, 5, 6})
	requireIntEntries(t, owned, nil)
	prefixed, err := moved.ReleaseLengthPrefixed()
	if err != nil || !slices.Equal(prefixed, []int{4, 8, 5, 6}) {
		t.Fatal(prefixed, err)
	}
	released, err := moved.Release()
	if err != nil || len(released) != 3 || !slices.Equal(released, []int{8, 5, 6}) {
		t.Fatal(released, err)
	}
}

func TestIntVectorAccessFillGrowthAndMutation(t *testing.T) {
	vector, _ := NewExactIntVector(0, IntVectorLimits{MaxEntries: 128})
	if err := vector.FillTwo(5, 2, 3); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, []int{2, 3, 2, 3, 2})
	if value, err := vector.AddTo(1, 4); err != nil || value != 7 {
		t.Fatal(value, err)
	}
	_ = vector.UpdateMax(0, 9)
	_ = vector.UpdateMax(0, 1)
	_ = vector.UpdateMin(2, -2)
	_ = vector.UpdateMin(2, 8)
	requireIntEntries(t, vector, []int{9, 7, -2, 3, 2})
	if value, err := vector.Get(7, 0); err != nil || value != 0 {
		t.Fatal(value, err)
	}
	if err := vector.SetSparse(6, 11, -1); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, []int{9, 7, -2, 3, 2, 0, 11, 0})
	if err := vector.Clear(); err != nil || vector.size != 0 {
		t.Fatal(err)
	}
	if err := vector.Resize(8); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, []int{9, 7, -2, 3, 2, 0, 11, 0})
	if err := vector.FillNatural(4); err != nil {
		t.Fatal(err)
	}
	if err := vector.Push(4, 5, 6, 7); err != nil {
		t.Fatal(err)
	}
	if err := vector.Prepend(-1); err != nil {
		t.Fatal(err)
	}
	if err := vector.Insert(3, 99); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, []int{-1, 0, 1, 99, 2, 3, 4, 5, 6, 7})
	if found, _ := vector.Find(99); found != 3 {
		t.Fatal(found)
	}
	removed, err := vector.Remove(99, false)
	if err != nil || !removed {
		t.Fatal(removed, err)
	}
	if err = vector.Drop(0); err != nil {
		t.Fatal(err)
	}
	if popped, err := vector.Pop(); err != nil || popped != 7 {
		t.Fatal(popped, err)
	}
	requireIntEntries(t, vector, []int{0, 1, 2, 3, 4, 5, 6})
	if err = vector.Reverse(); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, []int{6, 5, 4, 3, 2, 1, 0})
	if maxValue, _ := vector.MinMax(true); maxValue != 6 {
		t.Fatal(maxValue)
	}
	if minValue, _ := vector.MinMax(false); minValue != 0 {
		t.Fatal(minValue)
	}

	fetch, _ := NewExactIntVector(4, IntVectorLimits{})
	_ = fetch.Push(1, 2)
	if start, err := fetch.Fetch(0); err != nil || start != -1 {
		t.Fatal(start, err)
	}
	if start, err := fetch.Fetch(2); err != nil || start != 2 {
		t.Fatal(start, err)
	}
	requireIntEntries(t, fetch, []int{1, 2, 0, 0})
	before, _ := fetch.Entries()
	if _, err := fetch.Fetch(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	requireIntEntries(t, fetch, before)
	dormant, _ := NewExactIntVector(2, IntVectorLimits{})
	_ = dormant.Push(5, 6)
	_ = dormant.Shrink(1)
	if start, err := dormant.Fetch(1); err != nil || start != 1 {
		t.Fatal(start, err)
	}
	requireIntEntries(t, dormant, []int{5, 6})
}

func TestIntVectorOrderingCountsTransformsAndFormatting(t *testing.T) {
	vector, _ := NewIntVectorFrom([]int{3, -1, 3, 0, 2, 0}, IntVectorLimits{})
	if total, err := vector.Sum(false); err != nil || total != 7 {
		t.Fatal(total, err)
	}
	if total, _ := vector.Sum(true); total != 8 {
		t.Fatal(total)
	}
	if count, _ := vector.Count(3); count != 2 {
		t.Fatal(count)
	}
	if count, _ := vector.CountLarger(0); count != 3 {
		t.Fatal(count)
	}
	if count, _ := vector.CountSmaller(0); count != 1 {
		t.Fatal(count)
	}
	if count, _ := vector.CountPositive(); count != 3 {
		t.Fatal(count)
	}
	if count, _ := vector.CountZero(); count != 2 {
		t.Fatal(count)
	}
	if duplicates, err := vector.CountDuplicates(); err != nil || duplicates != 2 {
		t.Fatal(duplicates, err)
	}
	if unique, _ := vector.Unique(); unique {
		t.Fatal("duplicates should be found")
	}
	if err := vector.Sort(false); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, []int{-1, 0, 0, 2, 3, 3})
	if sorted, _ := vector.IsSorted(false); !sorted {
		t.Fatal("not sorted")
	}
	if duplicates, err := vector.Uniqify(); err != nil || duplicates != 2 {
		t.Fatal(duplicates, err)
	}
	requireIntEntries(t, vector, []int{-1, 0, 2, 3})
	if err := vector.SortUnsigned(); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, []int{0, 2, 3, -1})

	pairs, _ := NewIntVectorFrom([]int{2, 3, 1, 9, 2, 3, 1, 8}, IntVectorLimits{})
	if duplicates, err := pairs.UniqifyPairs(); err != nil || duplicates != 1 {
		t.Fatal(duplicates, err)
	}
	requireIntEntries(t, pairs, []int{1, 9, 1, 8, 2, 3})
	// Equal first keys deliberately preserve input order where qsort did not.

	invertSource, _ := NewIntVectorFrom([]int{2, -1, 0}, IntVectorLimits{})
	inverted, err := invertSource.Invert(-1)
	if err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, inverted, []int{2, -1, 0})
	positiveFill, _ := NewIntVectorFrom([]int{2, 5}, IntVectorLimits{})
	positiveInverted, err := positiveFill.Invert(5)
	if err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, positiveInverted, []int{5, 5, 0, 5, 5, 5})
	if capacity, _ := positiveInverted.Capacity(); capacity != 6 {
		t.Fatalf("invert capacity %d, want exact source growth 6", capacity)
	}
	condensed, _ := invertSource.Condense(-1)
	requireIntEntries(t, condensed, []int{2, 0})
	if capacity, _ := condensed.Capacity(); capacity != 16 {
		t.Fatalf("condense capacity %d, want source reservation 16", capacity)
	}

	parity, _ := NewIntVectorFrom([]int{0, 1, 2, 3, 4, 5}, IntVectorLimits{})
	if err := parity.RemoveParity(true); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, parity, []int{0, 2, 4})
	parity, _ = NewIntVectorFrom([]int{0, 1, 2, 3, 4, 5}, IntVectorLimits{})
	_ = parity.RemoveParity(false)
	requireIntEntries(t, parity, []int{1, 3, 5})

	textVector, _ := NewIntVectorFrom([]int{1, 0, -2}, IntVectorLimits{})
	if text, _ := textVector.Format(); text != "Vector has 3 entries: { 1 0 -2 }\n" {
		t.Fatalf("%q", text)
	}
	if text, _ := textVector.BinaryText(); text != "101" {
		t.Fatalf("%q", text)
	}
}

func TestIntVectorOrderedSetsRecordsAndSelection(t *testing.T) {
	ordered, _ := NewIntVector(0, IntVectorLimits{})
	for _, value := range []int{3, 1, 2, 2} {
		if err := ordered.PushOrdered(value, false); err != nil {
			t.Fatal(err)
		}
	}
	requireIntEntries(t, ordered, []int{1, 2, 2, 3})
	duplicate, err := ordered.PushUniqueOrdered(2, false)
	if err != nil || !duplicate {
		t.Fatal(duplicate, err)
	}
	duplicate, err = ordered.PushUniqueOrdered(4, false)
	if err != nil || duplicate {
		t.Fatal(duplicate, err)
	}
	requireIntEntries(t, ordered, []int{1, 2, 2, 3, 4})

	costs, _ := NewIntVectorFrom([]int{9, 4, 7, 1}, IntVectorLimits{})
	byCost, _ := NewIntVector(0, IntVectorLimits{})
	for _, value := range []int{0, 1, 2, 3} {
		if err := byCost.PushOrderedByCost(value, costs); err != nil {
			t.Fatal(err)
		}
	}
	requireIntEntries(t, byCost, []int{3, 1, 2, 0})

	first, _ := NewIntVectorFrom([]int{1, 2, 2, 4, 8}, IntVectorLimits{})
	second, _ := NewIntVectorFrom([]int{2, 2, 3, 8}, IntVectorLimits{})
	common, err := CommonSorted(first, second, false)
	if err != nil || !slices.Equal(common, []int{2, 2, 8}) {
		t.Fatal(common, err)
	}
	merged, err := MergeSorted(first, second)
	if err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, merged, []int{1, 2, 2, 3, 4, 8})
	if capacity, _ := merged.Capacity(); capacity != 16 {
		t.Fatalf("merge capacity %d, want source reservation 16", capacity)
	}
	all, left, right, err := SplitSorted(first, second)
	if err != nil || !slices.Equal(all, []int{2, 2, 8}) || !slices.Equal(left, []int{1, 4}) || !slices.Equal(right, []int{3}) {
		t.Fatal(all, left, right, err)
	}
	removeFirst, _ := first.Clone()
	removeSecond, _ := second.Clone()
	removed, err := RemoveCommonSorted(removeFirst, removeSecond)
	if err != nil || !slices.Equal(removed, common) {
		t.Fatal(removed, err)
	}
	requireIntEntries(t, removeFirst, []int{1, 4})
	requireIntEntries(t, removeSecond, []int{3})

	records, _ := NewIntVectorFrom([]int{1, 2, 3, 1, 2, 3, 4, 5, 6}, IntVectorLimits{})
	unique, classes, err := records.IntVectorRecordClasses(3)
	if err != nil || unique != 2 || !slices.Equal(classes, []int{0, 0, 2}) {
		t.Fatal(unique, classes, err)
	}
	compact, err := records.UniqifyRecords(3)
	if err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, compact, []int{1, 2, 3, 4, 5, 6})
	repeated, _ := NewIntVectorFrom([]int{7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}, IntVectorLimits{})
	compactRepeated, err := repeated.UniqifyRecords(1)
	if err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, compactRepeated, []int{7})
	if capacity, _ := compactRepeated.Capacity(); capacity != 16 {
		t.Fatalf("record uniquing capacity %d, want unique-size reservation 16", capacity)
	}
	hash, err := IntVectorRecordHash([]int{1, -2, math.MaxInt32})
	if err != nil || hash == 0 {
		t.Fatal(hash, err)
	}
	if hash, err = IntVectorRecordHash([]int{1}); err != nil || hash != 1338126263 {
		t.Fatalf("record hash %d: %v", hash, err)
	}
	if hash2, hashErr := IntVectorRecordHash2([]int{1}); hashErr != nil || hash2 != 12582917 {
		t.Fatalf("record hash2 %d: %v", hash2, hashErr)
	}
	if hash2, hashErr := IntVectorRecordHash2([]int{-1}); hashErr != nil || hash2 != 885011404 {
		t.Fatalf("record hash2 -1 %d: %v", hash2, hashErr)
	}
	if _, err = IntVectorRecordHash([]int{math.MaxInt}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}

	values := []int{3, 1, 2}
	SelectSort(values, false)
	if !slices.Equal(values, []int{1, 2, 3}) {
		t.Fatal(values)
	}
	values, parallelCosts := []int{10, 20, 30}, []int{2, 1, 3}
	if err = SelectSortParallel(values, parallelCosts, false); err != nil || !slices.Equal(values, []int{20, 10, 30}) || !slices.Equal(parallelCosts, []int{1, 2, 3}) {
		t.Fatal(values, parallelCosts, err)
	}
}

func TestIntVectorTraversalAppendRemapErrorsAndLifecycle(t *testing.T) {
	vector, _ := NewIntVectorFrom([]int{0, 1, 2, 3, 4, 5}, IntVectorLimits{})
	allEntries, err := vector.EntriesRange(0, 6, false)
	if err != nil || len(allEntries) != 6 || allEntries[5] != (IntVectorIndexedEntry{5, 5}) {
		t.Fatal(allEntries, err)
	}
	startEntries, err := vector.EntriesRange(2, 6, false)
	if err != nil || len(startEntries) != 4 || startEntries[0].Index != 2 {
		t.Fatal(startEntries, err)
	}
	stopEntries, err := vector.EntriesRange(0, 3, false)
	if err != nil || len(stopEntries) != 3 || stopEntries[2].Value != 2 {
		t.Fatal(stopEntries, err)
	}
	entries, err := vector.EntriesRange(1, 5, false)
	if err != nil || entries[0] != (IntVectorIndexedEntry{1, 1}) || entries[3] != (IntVectorIndexedEntry{4, 4}) {
		t.Fatal(entries, err)
	}
	reverse, _ := vector.EntriesRange(1, 5, true)
	if reverse[0].Value != 4 || reverse[3].Value != 1 {
		t.Fatal(reverse)
	}
	fullReverse, _ := vector.EntriesRange(0, 6, true)
	if len(fullReverse) != 6 || fullReverse[0] != (IntVectorIndexedEntry{5, 5}) {
		t.Fatal(fullReverse)
	}
	zipAll, err := ZipIntVectors(vector, vector, 0)
	if err != nil || len(zipAll) != 6 || zipAll[0] != (IntVectorPair{0, 0, 0}) {
		t.Fatal(zipAll, err)
	}
	zipped, err := ZipIntVectors(vector, vector, 2)
	if err != nil || len(zipped) != 4 || zipped[0] != (IntVectorPair{2, 2, 2}) {
		t.Fatal(zipped, err)
	}
	pairs, err := vector.Groups(2, 1)
	if err != nil || len(pairs) != 2 || pairs[0] != (IntVectorTriple{Index: 1, First: 1, Second: 2}) {
		t.Fatal(pairs, err)
	}
	doubles, err := vector.Groups(2, 0)
	if err != nil || len(doubles) != 3 || doubles[2] != (IntVectorTriple{Index: 4, First: 4, Second: 5}) {
		t.Fatal(doubles, err)
	}
	triples, err := vector.Groups(3, 0)
	if err != nil || len(triples) != 2 || triples[1] != (IntVectorTriple{Index: 3, First: 3, Second: 4, Third: 5}) {
		t.Fatal(triples, err)
	}
	thisNext, _ := vector.ThisNext()
	if !slices.Equal(thisNext, []IntVectorPair{{0, 0, 1}, {2, 1, 3}, {4, 3, 5}}) {
		t.Fatal(thisNext)
	}
	indexes, _ := NewIntVectorFrom([]int{5, 0, 2}, IntVectorLimits{})
	selected, err := EntriesByIndex(vector, indexes)
	if err != nil || !slices.Equal(selected, []int{5, 0, 2}) {
		t.Fatal(selected, err)
	}
	appended, _ := NewIntVectorFrom([]int{9}, IntVectorLimits{})
	if err = appended.Append(vector); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, appended, []int{9, 0, 1, 2, 3, 4, 5})
	if err = appended.AppendSkip(vector, 2); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, appended, []int{9, 0, 1, 2, 3, 4, 5, 0, 1, 3, 4, 5})
	if err = appended.ReplaceWithSigned(vector, true); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, appended, []int{0, -1, -2, -3, -4, -5})

	oldToNew, _ := NewIntVectorFrom([]int{0, 3, 1, -1}, IntVectorLimits{})
	old, _ := NewIntVectorFrom([]int{0, 7, 8, 9}, IntVectorLimits{})
	remapped, err := RemapIntVector(oldToNew, old, 4, IntVectorLimits{})
	if err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, remapped, []int{0, 8, 0, 7})

	if _, err = NewIntVector(-1, IntVectorLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	limited, _ := NewExactIntVector(2, IntVectorLimits{MaxEntries: 2})
	_ = limited.Push(1, 2)
	before, _ := limited.Entries()
	if err = limited.Push(3); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	requireIntEntries(t, limited, before)
	cancelled := false
	cancellable, _ := NewIntVectorFrom([]int{2, 1}, IntVectorLimits{Stop: func() bool { return cancelled }})
	cancelled = true
	before, _ = cancellable.Entries()
	if err = cancellable.Sort(false); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	requireIntEntries(t, cancellable, before)
	if err = cancellable.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = cancellable.Size(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestIntVectorRemainingOperationsAndBoundaries(t *testing.T) {
	vector, err := NewExactIntVector(0, IntVectorLimits{MaxEntries: 64})
	if err != nil {
		t.Fatal(err)
	}
	if err = vector.Grow(5); err != nil {
		t.Fatal(err)
	}
	if capacity, _ := vector.Capacity(); capacity != 5 {
		t.Fatal(capacity)
	}
	if err = vector.FillExtra(3, 7); err != nil {
		t.Fatal(err)
	}
	if first, _ := vector.Entry(0); first != 7 {
		t.Fatal(first)
	}
	if last, _ := vector.Last(); last != 7 {
		t.Fatal(last)
	}
	clone, _ := vector.Clone()
	if equal, err := vector.Equal(clone); err != nil || !equal {
		t.Fatal(equal, err)
	}
	if count, err := vector.CountUnique(); err != nil || count != 1 {
		t.Fatal(count, err)
	}
	short, _ := NewIntVectorFrom([]int{7}, IntVectorLimits{})
	if count, err := short.CountCommon(vector); err != nil || count != 3 {
		t.Fatal(count, err)
	}
	tieFirst, _ := NewIntVectorFrom([]int{1, 1, 1}, IntVectorLimits{})
	tieSecond, _ := NewIntVectorFrom([]int{1, 2, 3}, IntVectorLimits{})
	if count, err := tieFirst.CountCommon(tieSecond); err != nil || count != 3 {
		t.Fatalf("equal-size source multiplicity %d: %v", count, err)
	}
	if count, err := tieSecond.CountCommon(tieFirst); err != nil || count != 1 {
		t.Fatalf("equal-size reversed source multiplicity %d: %v", count, err)
	}
	negative, _ := NewIntVectorFrom([]int{-1}, IntVectorLimits{})
	if _, err = negative.CountCommon(vector); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}

	records, _ := NewIntVectorFrom([]int{2, 9, 1, 8, 2, 7, 1, 6}, IntVectorLimits{})
	if err = records.SortRecords(2, false); err != nil {
		t.Fatal(err)
	}
	recordEntries, _ := records.Entries()
	if recordEntries[0] != 1 || recordEntries[2] != 1 || recordEntries[4] != 2 || recordEntries[6] != 2 {
		t.Fatal(recordEntries)
	}
	hash1, err := IntVectorRecordHash2([]int{0, 1, -1, math.MinInt32})
	if err != nil {
		t.Fatal(err)
	}
	hash2, _ := IntVectorRecordHash2([]int{0, 1, -1, math.MinInt32})
	if hash1 != hash2 {
		t.Fatal(hash1, hash2)
	}
	debugHash, steps, err := IntVectorRecordHashDebug([]int{0, 1, -1, math.MinInt32}, 15)
	if err != nil || debugHash != hash1 || len(steps) != 16 || steps[len(steps)-1].Key != debugHash {
		t.Fatal(debugHash, len(steps), err)
	}
	profile, err := records.IntVectorRecordProfile(2)
	if err != nil || len(profile) != 4 || profile[0]+profile[1]+profile[2]+profile[3] != 4 {
		t.Fatal(profile, err)
	}

	removeFrom, _ := NewIntVectorFrom([]int{1, 2, 2, 3, 5}, IntVectorLimits{})
	removeThese, _ := NewIntVectorFrom([]int{2, 4}, IntVectorLimits{})
	if err = RemoveSorted(removeFrom, removeThese); err != nil {
		t.Fatal(err)
	}
	requireIntEntries(t, removeFrom, []int{1, 2, 3, 5})
	costs, _ := NewIntVectorFrom([]int{40, 10, 30, 20}, IntVectorLimits{})
	indexes := []int{0, 1, 2, 3}
	if err = SelectSortByIndexedCost(indexes, costs, false); err != nil || !slices.Equal(indexes, []int{1, 3, 2, 0}) {
		t.Fatal(indexes, err)
	}
	byCost, _ := NewIntVectorFrom([]int{2, 0}, IntVectorLimits{})
	duplicate, err := byCost.PushUniqueOrderedByCost(2, costs)
	if err != nil || !duplicate {
		t.Fatal(duplicate, err)
	}
	duplicate, err = byCost.PushUniqueOrderedByCost(1, costs)
	if err != nil || duplicate {
		t.Fatal(duplicate, err)
	}
	requireIntEntries(t, byCost, []int{1, 2, 0})

	if comparison, err := CompareIntVectors(nil, vector); err != nil || comparison >= 0 {
		t.Fatal(comparison, err)
	}
	if comparison, err := CompareIntVectors(vector, clone); err != nil || comparison != 0 {
		t.Fatal(comparison, err)
	}
	bytewiseHigh, _ := NewIntVectorFrom([]int{256}, IntVectorLimits{})
	bytewiseLow, _ := NewIntVectorFrom([]int{1}, IntVectorLimits{})
	if comparison, err := CompareIntVectors(bytewiseHigh, bytewiseLow); err != nil || comparison >= 0 {
		t.Fatalf("little-endian memcmp ordering = %d: %v", comparison, err)
	}
	var nilVector *IntVector
	if _, err = nilVector.UniqifyPairs(); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil pair uniquing: %v", err)
	}
	if _, err = NewRangeIntVector(math.MaxInt, 2, IntVectorLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = clone.Erase(); err != nil {
		t.Fatal(err)
	}
	if size, _ := clone.Size(); size != 0 || clone.MemoryBytes() != 16 {
		t.Fatal(size, clone.MemoryBytes())
	}
}

func TestIntVectorLimitsCancellationAndAtomicity(t *testing.T) {
	limited, err := NewExactIntVector(3, IntVectorLimits{MaxEntries: 3})
	if err != nil {
		t.Fatal(err)
	}
	if err = limited.Push(1, 2, 3); err != nil {
		t.Fatal(err)
	}
	before, _ := limited.Entries()
	if err = limited.FillExtra(4, 9); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	requireIntEntries(t, limited, before)
	if err = limited.Insert(1, 8); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	requireIntEntries(t, limited, before)
	if err = limited.SetSparse(3, 8, -1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	requireIntEntries(t, limited, before)

	cancelled := false
	vector, err := NewExactIntVector(4, IntVectorLimits{MaxEntries: 16, Stop: func() bool { return cancelled }})
	if err != nil {
		t.Fatal(err)
	}
	if err = vector.Push(4, 3, 2, 1); err != nil {
		t.Fatal(err)
	}
	before, _ = vector.Entries()
	capacity, _ := vector.Capacity()
	cancelled = true
	operations := []func() error{
		func() error { return vector.Grow(8) },
		func() error { return vector.Resize(2) },
		func() error { return vector.Fill(2, 0) },
		func() error { return vector.FillExtra(5, 0) },
		func() error { return vector.Push(0) },
		func() error { return vector.Drop(0) },
		func() error { return vector.Reverse() },
		func() error { return vector.Erase() },
	}
	for i, operation := range operations {
		if err = operation(); !errors.Is(err, ErrCancelled) {
			t.Fatalf("operation %d: %v", i, err)
		}
		requireIntEntries(t, vector, before)
		if gotCapacity, _ := vector.Capacity(); gotCapacity != capacity {
			t.Fatalf("operation %d changed capacity to %d", i, gotCapacity)
		}
	}
	if _, err = vector.Pop(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, before)
	if _, err = vector.Release(); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	requireIntEntries(t, vector, before)
}

type intVectorModel struct {
	values []int
	limit  int
}

func reverseIntModel(values []int) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func sortIntModel(values []int) {
	for i := 1; i < len(values); i++ {
		value := values[i]
		j := i
		for j > 0 && values[j-1] > value {
			values[j] = values[j-1]
			j--
		}
		values[j] = value
	}
}

func containsIntModel(values []int, value int) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func equalIntModel(first, second []int) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i] != second[i] {
			return false
		}
	}
	return true
}

func runIntVectorModel(data []byte) error {
	vector, err := NewExactIntVector(0, IntVectorLimits{MaxEntries: 96})
	if err != nil {
		return err
	}
	model := intVectorModel{limit: 96}
	for step := 0; step+1 < len(data); step += 2 {
		op, raw := data[step]%10, int(int8(data[step+1]))
		switch op {
		case 0:
			if len(model.values) < model.limit {
				if err = vector.Push(raw); err != nil {
					return err
				}
				model.values = append(model.values, raw)
			}
		case 1:
			if len(model.values) != 0 {
				got, popErr := vector.Pop()
				want := model.values[len(model.values)-1]
				if popErr != nil || got != want {
					return ErrInvariant
				}
				model.values = model.values[:len(model.values)-1]
			}
		case 2:
			if len(model.values) != 0 {
				index := raw
				if index < 0 {
					index = -index
				}
				index %= len(model.values)
				if err = vector.Write(index, raw); err != nil {
					return err
				}
				model.values[index] = raw
			}
		case 3:
			if len(model.values) != 0 {
				index := raw
				if index < 0 {
					index = -index
				}
				index %= len(model.values)
				if err = vector.Drop(index); err != nil {
					return err
				}
				model.values = append(model.values[:index], model.values[index+1:]...)
			}
		case 4:
			if err = vector.Reverse(); err != nil {
				return err
			}
			reverseIntModel(model.values)
		case 5:
			if err = vector.Sort(false); err != nil {
				return err
			}
			sortIntModel(model.values)
		case 6:
			if len(model.values) < model.limit {
				index := raw
				if index < 0 {
					index = -index
				}
				index %= len(model.values) + 1
				if err = vector.Insert(index, raw); err != nil {
					return err
				}
				model.values = append(model.values, 0)
				copy(model.values[index+1:], model.values[index:])
				model.values[index] = raw
			}
		case 7:
			if len(model.values) != 0 {
				size := raw
				if size < 0 {
					size = -size
				}
				size %= len(model.values) + 1
				if err = vector.Shrink(size); err != nil {
					return err
				}
				model.values = model.values[:size]
			}
		case 8:
			duplicate := containsIntModel(model.values, raw)
			if !duplicate && len(model.values) == model.limit {
				before, entriesErr := vector.Entries()
				if entriesErr != nil {
					return entriesErr
				}
				_, pushErr := vector.PushUnique(raw)
				if !errors.Is(pushErr, ErrLimit) {
					return ErrInvariant
				}
				after, entriesErr := vector.Entries()
				if entriesErr != nil || !slices.Equal(before, after) {
					return ErrInvariant
				}
				break
			}
			got, pushErr := vector.PushUnique(raw)
			if pushErr != nil {
				return pushErr
			}
			if got != duplicate {
				return ErrInvariant
			}
			if !duplicate && len(model.values) < model.limit {
				model.values = append(model.values, raw)
			}
		case 9:
			if err = vector.Clear(); err != nil {
				return err
			}
			model.values = model.values[:0]
		}
		got, entriesErr := vector.Entries()
		if entriesErr != nil || !equalIntModel(got, model.values) {
			return ErrInvariant
		}
	}
	return nil
}

func TestIntVectorIndependentModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	data := make([]byte, 96)
	for execution := 0; execution < 5000; execution++ {
		for i := range data {
			state ^= state << 7
			state ^= state >> 9
			state ^= state << 8
			data[i] = byte(state)
		}
		if err := runIntVectorModel(data); err != nil {
			t.Fatalf("execution %d: %v", execution, err)
		}
	}
}

func FuzzIntVectorModel(f *testing.F) {
	f.Add([]byte{0, 1, 0, 2, 4, 0, 5, 0})
	f.Add([]byte{0, 255, 6, 3, 2, 4, 8, 4})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			return
		}
		if err := runIntVectorModel(data); err != nil {
			t.Fatal(err)
		}
	})
}
