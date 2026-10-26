package abc

import (
	"errors"
	"math"
	"slices"
	"sort"
	"testing"
)

func msatSortTestClauses(activities []float32) []*MsatClause {
	clauses := make([]*MsatClause, len(activities))
	for index, activity := range activities {
		clauses[index] = &MsatClause{number: index, activity: activity}
	}
	return clauses
}

func msatSortedActivities(clauses []*MsatClause) []float32 {
	result := make([]float32, len(clauses))
	for index, clause := range clauses {
		result[index] = clause.activity
	}
	return result
}

func msatClauseNumbers(clauses []*MsatClause) []int {
	result := make([]int, len(clauses))
	for index, clause := range clauses {
		result[index] = clause.number
	}
	return result
}

func TestMsatClauseSortSelectionAndQuickDeterminism(t *testing.T) {
	for size := 0; size <= 80; size++ {
		activities := make([]float32, size)
		for index := range activities {
			activities[index] = float32((index*37+size*11)%17 - 8)
		}
		clauses := msatSortTestClauses(activities)
		second := slices.Clone(clauses)
		if err := SortMsatClausesByActivity(clauses); err != nil {
			t.Fatal(size, err)
		}
		if err := SortMsatClausesByActivity(second); err != nil || !slices.Equal(clauses, second) {
			t.Fatal("nondeterministic", size, err, msatClauseNumbers(clauses), msatClauseNumbers(second))
		}
		want := slices.Clone(activities)
		slices.Sort(want)
		if !slices.Equal(msatSortedActivities(clauses), want) {
			t.Fatal(size, msatSortedActivities(clauses), want)
		}
		ids := msatClauseNumbers(clauses)
		slices.Sort(ids)
		for index, id := range ids {
			if id != index {
				t.Fatal("lost identity", size, ids)
			}
		}
	}
	if got := msatSortIRand(msatClauseSortSeed, 100); got != 38 {
		t.Fatal("native pivot recurrence", got)
	}
	if compare := msatClauseActivityCompare(&MsatClause{activity: -1}, &MsatClause{activity: 2}); compare != -1 {
		t.Fatal(compare)
	}
}

func TestMsatClauseVectorSortAndAtomicErrors(t *testing.T) {
	vector, _ := NewMsatClauseVector(0, 32)
	clauses := msatSortTestClauses([]float32{3, -1, 2, 2, 0})
	for _, clause := range clauses {
		vector.Push(clause)
	}
	if err := vector.SortByActivity(); err != nil || !slices.Equal(msatSortedActivities(vector.Values()), []float32{-1, 0, 2, 2, 3}) {
		t.Fatal(err, msatSortedActivities(vector.Values()))
	}
	withNil := slices.Clone(clauses)
	withNil[2] = nil
	before := slices.Clone(withNil)
	if err := SortMsatClausesByActivity(withNil); !errors.Is(err, ErrDimension) || !slices.Equal(withNil, before) {
		t.Fatal(err, withNil, before)
	}
	closed := &MsatClause{closed: true}
	if err := SortMsatClausesByActivity([]*MsatClause{closed}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	bad, _ := NewMsatClauseVector(0, 16)
	bad.Push(nil)
	if err := bad.SortByActivity(); !errors.Is(err, ErrDimension) || bad.Values()[0] != nil {
		t.Fatal(err, bad.Values())
	}
	bad.Close()
	if err := bad.SortByActivity(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestMsatClauseSortNaNAndEqualPermutation(t *testing.T) {
	activities := []float32{1, float32(math.NaN()), 1, -1, float32(math.Inf(1)), float32(math.Inf(-1)), 1}
	clauses := msatSortTestClauses(activities)
	if err := SortMsatClausesByActivity(clauses); err != nil {
		t.Fatal(err)
	}
	ids := msatClauseNumbers(clauses)
	slices.Sort(ids)
	if !slices.Equal(ids, []int{0, 1, 2, 3, 4, 5, 6}) {
		t.Fatal(ids)
	}
	equal := msatSortTestClauses(make([]float32, 20))
	if err := SortMsatClausesByActivity(equal); err != nil {
		t.Fatal(err)
	}
	if got, want := msatClauseNumbers(equal), []int{19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}; !slices.Equal(got, want) {
		t.Fatal(got, want)
	}
}

func FuzzMsatClauseSort(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {3, 1, 2}, {255, 0, 127, 127, 1}, {9, 8, 7, 6, 5, 4, 3, 2, 1, 0}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 256 {
			return
		}
		activities := make([]float32, len(raw))
		for index, value := range raw {
			activities[index] = float32(int8(value))
		}
		clauses := msatSortTestClauses(activities)
		again := slices.Clone(clauses)
		if err := SortMsatClausesByActivity(clauses); err != nil {
			t.Fatal(err)
		}
		if err := SortMsatClausesByActivity(again); err != nil || !slices.Equal(clauses, again) {
			t.Fatal("determinism", err)
		}
		want := slices.Clone(activities)
		sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
		if !slices.Equal(msatSortedActivities(clauses), want) {
			t.Fatal(msatSortedActivities(clauses), want)
		}
		ids := msatClauseNumbers(clauses)
		slices.Sort(ids)
		for index, id := range ids {
			if id != index {
				t.Fatal("permutation", ids)
			}
		}
	})
}
