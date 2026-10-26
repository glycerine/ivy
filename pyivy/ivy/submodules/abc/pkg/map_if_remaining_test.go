package abc

// Standalone tests for bounded IF cut, truth, LUT-library, and timing behavior.

import (
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
)

func ifTruthFromFunction(t testing.TB, leaves []int, fn func(int) bool) *IfMappingCut {
	t.Helper()
	truth := make([]uint64, ifTruthWords(len(leaves)))
	for assignment := 0; assignment < 1<<uint(len(leaves)); assignment++ {
		if fn(assignment) {
			truth[assignment>>6] |= uint64(1) << uint(assignment&63)
		}
	}
	cut, err := NewIfMappingCut(leaves, truth, IfCutCost{}, IfMapLimits{})
	if err != nil {
		t.Fatal(err)
	}
	return cut
}

func TestIfMappingCutMergeTruthExhaustive(t *testing.T) {
	for leftTruth := 0; leftTruth < 16; leftTruth++ {
		for rightTruth := 0; rightTruth < 16; rightTruth++ {
			left := ifTruthFromFunction(t, []int{0, 2}, func(a int) bool { return leftTruth&(1<<uint(a)) != 0 })
			right := ifTruthFromFunction(t, []int{1, 2}, func(a int) bool { return rightTruth&(1<<uint(a)) != 0 })
			for flags := 0; flags < 4; flags++ {
				got, err := MergeIfMappingCuts(left, right, flags&1 != 0, flags&2 != 0, IfCutCost{}, IfMapLimits{MaxWork: 1000, MaxBytes: 1000})
				if err != nil || !slices.Equal(got.Leaves(), []int{0, 1, 2}) {
					t.Fatal(leftTruth, rightTruth, flags, err)
				}
				for assignment := 0; assignment < 8; assignment++ {
					la := (assignment & 1) | ((assignment >> 1) & 2)
					ra := ((assignment >> 1) & 1) | ((assignment >> 1) & 2)
					want := (leftTruth&(1<<uint(la)) != 0) != (flags&1 != 0)
					want = want && ((rightTruth&(1<<uint(ra)) != 0) != (flags&2 != 0))
					if got.value(assignment) != want {
						t.Fatalf("left=%x right=%x flags=%d assignment=%d", leftTruth, rightTruth, flags, assignment)
					}
				}
			}
		}
	}
}

func TestIfCutPermutationDelayAndLibrary(t *testing.T) {
	cut := ifTruthFromFunction(t, []int{10, 20, 30}, func(a int) bool { return a&1 != 0 || a&2 != 0 && a&4 != 0 })
	permuted, err := PermuteIfCutPins(cut, []float64{2, 7, 4}, IfMapLimits{MaxWork: 1000, MaxBytes: 1000})
	if err != nil || !slices.Equal(permuted.Leaves, []int{20, 30, 10}) || !slices.Equal(permuted.OldPins, []int{1, 2, 0}) {
		t.Fatal(permuted, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		old := ((assignment >> 2) & 1) | ((assignment & 1) << 1) | (((assignment >> 1) & 1) << 2)
		got := permuted.Truth[assignment>>6]&(1<<uint(assignment&63)) != 0
		if got != (cut.value(old) != permuted.Complemented) {
			t.Fatal(assignment, old)
		}
	}
	complemented := ifTruthFromFunction(t, []int{10, 20, 30}, func(a int) bool { return a == 0 || a == 7 })
	if !complemented.Complemented() {
		t.Fatal("expected normalized complemented cut")
	}
	permutedComplemented, err := PermuteIfCutPins(complemented, []float64{2, 7, 4}, IfMapLimits{MaxWork: 1000, MaxBytes: 1000})
	if err != nil || !permutedComplemented.Complemented || permutedComplemented.Truth[0]&1 != 0 {
		t.Fatal(permutedComplemented, err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		old := ((assignment >> 2) & 1) | ((assignment & 1) << 1) | (((assignment >> 1) & 1) << 2)
		stored := permutedComplemented.Truth[assignment>>6]&(1<<uint(assignment&63)) != 0
		if (stored != permutedComplemented.Complemented) != complemented.value(old) {
			t.Fatal("complemented phase lost", assignment, old)
		}
	}
	library, err := ParseIfLUTLibrary("# sized LUTs\n1 1.0 1.0\n2 1.2 0.5 1.5\n3 2.0 0.25 1.0 2.0\n", IfMapLimits{MaxWork: 100, MaxBytes: 1000})
	if err != nil || library.FastestPinDelay() != 0.25 || library.SlowestPinDelay() != 2 {
		t.Fatal(library, err)
	}
	delay, err := IfCutDelay(cut, []float64{2, 7, 4}, library, IfMapLimits{MaxWork: 100, MaxBytes: 1000})
	if err != nil || delay != math.Max(7.25, math.Max(5, 4)) {
		t.Fatal(delay, err)
	}
	entry, _ := library.Entry(3)
	entry.Delays[0] = 99
	entry2, _ := library.Entry(3)
	if entry2.Delays[0] == 99 {
		t.Fatal("library leaked delay storage")
	}
}

func TestIfCutCostsDominanceNormalizationAndOwnership(t *testing.T) {
	truth := []uint64{0xff}
	cut, err := NewIfMappingCut([]int{2, 5, 8}, truth, IfCutCost{Delay: 4, Area: 2}, IfMapLimits{})
	if err != nil || !cut.Complemented() || cut.truth[0]&1 != 0 || !slices.Equal(cut.Truth(), truth) {
		t.Fatal(cut, err)
	}
	truth[0] = 0
	leaves := cut.Leaves()
	leaves[0] = 99
	if cut.Truth()[0] != 0xff || cut.Leaves()[0] != 2 {
		t.Fatal("cut leaked caller or result storage")
	}
	small := ifTruthFromFunction(t, []int{2, 8}, func(a int) bool { return a != 0 })
	if !IfCutDominates(small, cut) || IfCutDominates(cut, small) {
		t.Fatal("dominance")
	}
	if signature := IfCutSignature(cut); signature != 1<<2|1<<5|1<<8 {
		t.Fatal(signature)
	}
	if got, err := NewIfMappingCut([]int{1}, []uint64{0}, IfCutCost{}, IfMapLimits{MaxBytes: 103}); got != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(got, err)
	}
	if got, err := NewIfMappingCut([]int{1}, []uint64{0}, IfCutCost{}, IfMapLimits{MaxBytes: 104}); err != nil || got == nil {
		t.Fatal(got, err)
	}
	other, _ := NewIfMappingCut([]int{1, 3}, []uint64{6}, IfCutCost{Delay: 4.01, Area: 1}, IfMapLimits{})
	if got, _ := CompareIfCutCost(cut, other, IfCutDelayFirst, 0.02); got <= 0 {
		t.Fatal(got)
	}
	if got, _ := CompareIfCutCost(cut, other, IfCutAreaFirst, 0.02); got <= 0 {
		t.Fatal(got)
	}
}

func TestIfMappingLimitsCancellationAndAtomicity(t *testing.T) {
	left := ifTruthFromFunction(t, []int{0, 2}, func(a int) bool { return a&1 != 0 })
	right := ifTruthFromFunction(t, []int{1, 3}, func(a int) bool { return a&2 != 0 })
	for _, tc := range []struct {
		limits IfMapLimits
		want   error
	}{
		{IfMapLimits{MaxLeaves: 3}, ErrLimit},
		{IfMapLimits{MaxWork: 2}, ErrLimit},
		{IfMapLimits{MaxBytes: 8}, ErrLimit},
		{IfMapLimits{Stop: func() bool { return true }}, ErrCancelled},
	} {
		got, err := MergeIfMappingCuts(left, right, false, false, IfCutCost{}, tc.limits)
		if got != nil || !errors.Is(err, tc.want) {
			t.Fatal(tc, got, err)
		}
	}
	calls := 0
	got, err := MergeIfMappingCuts(left, right, false, false, IfCutCost{}, IfMapLimits{Stop: func() bool { calls++; return calls > 5 }})
	if got != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(got, err)
	}
	for _, malformed := range []struct {
		leaves []int
		truth  []uint64
	}{{[]int{2, 1}, []uint64{0}}, {[]int{1, 1}, []uint64{0}}, {[]int{-1}, []uint64{0}}, {[]int{1}, nil}, {[]int{1}, []uint64{4}}} {
		if got, err = NewIfMappingCut(malformed.leaves, malformed.truth, IfCutCost{}, IfMapLimits{}); got != nil || err == nil {
			t.Fatal(malformed, got, err)
		}
	}
	for _, text := range []string{"", "2 1 1\n", "1 1\n", "1 x 1\n", "1 1 -1\n", "1 1 1\n2 1 2 1\n", "1 1 1\n2 1 1 2\n3 1 1\n"} {
		if library, parseErr := ParseIfLUTLibrary(text, IfMapLimits{}); library != nil || parseErr == nil {
			t.Fatal(text, library, parseErr)
		}
	}
	if got, err := MergeIfMappingCuts(left, right, false, false, IfCutCost{Delay: math.NaN()}, IfMapLimits{}); got != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(got, err)
	}
	if library, err := ParseIfLUTLibrary("# a\n# b\n1 1 1\n", IfMapLimits{MaxWork: 1}); library != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(library, err)
	}
}

func TestIfCompiledInventoryAndConcurrentReaders(t *testing.T) {
	sources := IfCompiledSources()
	if len(sources) != 26 || sources[0] != "ifCom.c" || sources[len(sources)-1] != "ifUtil.c" {
		t.Fatal(sources)
	}
	sources[0] = "bad"
	if IfCompiledSources()[0] != "ifCom.c" {
		t.Fatal("inventory leaked storage")
	}
	cut := ifTruthFromFunction(t, []int{0, 1, 2}, func(a int) bool { return a == 3 || a == 7 })
	var wait sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for i := 0; i < 100; i++ {
				if !slices.Equal(cut.Leaves(), []int{0, 1, 2}) || cut.Truth()[0] != 0x88 || IfTruthSupportSize(cut) != 2 {
					t.Error("unstable concurrent snapshot")
					return
				}
			}
		}()
	}
	wait.Wait()
}

func FuzzIfMappingCutMerge(f *testing.F) {
	for _, seed := range []struct{ a, b byte }{{0, 0}, {6, 9}, {15, 15}, {8, 7}} {
		f.Add(seed.a, seed.b)
	}
	f.Fuzz(func(t *testing.T, rawLeft, rawRight byte) {
		leftTruth, rightTruth := int(rawLeft&15), int(rawRight&15)
		left := ifTruthFromFunction(t, []int{0, 2}, func(a int) bool { return leftTruth&(1<<uint(a)) != 0 })
		right := ifTruthFromFunction(t, []int{1, 2}, func(a int) bool { return rightTruth&(1<<uint(a)) != 0 })
		got, err := MergeIfMappingCuts(left, right, rawLeft&16 != 0, rawRight&16 != 0, IfCutCost{}, IfMapLimits{MaxWork: 1000, MaxBytes: 1000})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			la := (assignment & 1) | ((assignment >> 1) & 2)
			ra := ((assignment >> 1) & 1) | ((assignment >> 1) & 2)
			want := (leftTruth&(1<<uint(la)) != 0) != (rawLeft&16 != 0)
			want = want && ((rightTruth&(1<<uint(ra)) != 0) != (rawRight&16 != 0))
			if got.value(assignment) != want {
				t.Fatal(assignment, got.Truth(), want)
			}
		}
	})
}
