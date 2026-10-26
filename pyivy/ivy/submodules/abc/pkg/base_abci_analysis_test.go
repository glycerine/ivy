package abc

// Standalone tests for abcMeasure.c, abcQuant.c, and abcUnate.c semantics.

import (
	"errors"
	"math/bits"
	"slices"
	"sync"
	"testing"
)

func abciTruth(t testing.TB, n int, fn func(int) bool) AbciTruthTable {
	t.Helper()
	words := make([]uint64, ((1<<uint(n))+63)/64)
	for x := 0; x < 1<<uint(n); x++ {
		if fn(x) {
			words[x>>6] |= 1 << uint(x&63)
		}
	}
	table, err := NewAbciTruthTable(n, words)
	if err != nil {
		t.Fatal(err)
	}
	return table
}

func TestAbciTruthMeasuresExhaustiveSmall(t *testing.T) {
	for n := 0; n <= 3; n++ {
		functions := 1 << uint(1<<uint(n))
		for truth := 0; truth < functions; truth++ {
			table := abciTruth(t, n, func(x int) bool { return truth&(1<<uint(x)) != 0 })
			got, err := AnalyzeAbciTruth(table, AbciAnalysisOptions{})
			if err != nil {
				t.Fatal(n, truth, err)
			}
			ones := 0
			for x := 0; x < 1<<uint(n); x++ {
				if truth&(1<<uint(x)) != 0 {
					ones++
				}
			}
			if got.Ones != ones || got.OnSetOffSetPairs != uint64(ones*((1<<uint(n))-ones)) {
				t.Fatal(n, truth, got)
			}
			for i := 0; i < n; i++ {
				want := 0
				for x := 0; x < 1<<uint(n); x++ {
					if (truth&(1<<uint(x)) != 0) != (truth&(1<<uint(x^(1<<uint(i)))) != 0) {
						want++
					}
				}
				if got.VariableDifferences[i] != want {
					t.Fatal(n, truth, i, got.VariableDifferences[i], want)
				}
				for k := 0; k < n; k++ {
					if i == k {
						continue
					}
					pairWant := 0
					mask := 1<<uint(i) | 1<<uint(k)
					for x := 0; x < 1<<uint(n); x++ {
						if x&mask != 0 {
							continue
						}
						f00 := truth&(1<<uint(x)) != 0
						f10 := truth&(1<<uint(x|(1<<uint(i)))) != 0
						f01 := truth&(1<<uint(x|(1<<uint(k)))) != 0
						f11 := truth&(1<<uint(x|mask)) != 0
						if f00 != f11 {
							pairWant += 4
						}
						if f01 != f10 {
							pairWant += 4
						}
					}
					if got.PairDifferences[i][k] != pairWant {
						t.Fatal(n, truth, i, k, got.PairDifferences[i][k], pairWant)
					}
				}
			}
		}
	}
}

func TestAbciActivityOrderOwnershipLimitsCancellation(t *testing.T) {
	table := abciTruth(t, 4, func(x int) bool {
		a, b, c := x&1 != 0, x&2 != 0, x&4 != 0
		return a != (b && c)
	})
	order, err := AbciVariableActivityOrder(table, AbciAnalysisOptions{})
	if err != nil || !slices.Equal(order, []int{3, 1, 2, 0}) {
		t.Fatal(order, err)
	}
	words := table.Words()
	words[0] = 0
	order2, err := AbciVariableActivityOrder(table, AbciAnalysisOptions{})
	if err != nil || !slices.Equal(order2, order) {
		t.Fatal("snapshot mutated table", order2, err)
	}
	for _, options := range []AbciAnalysisOptions{{MaxEvaluations: -1}, {MaxBytes: -1}, {MaxEvaluations: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		if got, err := AnalyzeAbciTruth(table, options); err == nil || got.VariableDifferences != nil {
			t.Fatal(got, err)
		}
	}
	if _, err := AnalyzeAbciTruth(table, AbciAnalysisOptions{MaxEvaluations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := AnalyzeAbciTruth(table, AbciAnalysisOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	needed := (4 + 4*4 + 4*3) * abciIntBytes
	if _, err := AnalyzeAbciTruth(table, AbciAnalysisOptions{MaxBytes: needed}); err != nil {
		t.Fatal("exact byte bound", err)
	}
	if got, err := AnalyzeAbciTruth(table, AbciAnalysisOptions{MaxBytes: needed - 1}); !errors.Is(err, ErrLimit) || got.VariableDifferences != nil {
		t.Fatal("byte preflight was not atomic", got, err)
	}
	polls := 0
	got, err := AnalyzeAbciTruth(table, AbciAnalysisOptions{Stop: func() bool {
		polls++
		return polls == 12
	}})
	if !errors.Is(err, ErrCancelled) || got.VariableDifferences != nil || polls != 12 {
		t.Fatal("mid-operation cancellation was not atomic", polls, got, err)
	}
}

func TestAbciBoundDecomposition(t *testing.T) {
	// Depend only on two of the six bound variables: each mask containing both
	// admits at most two distinct cofactors per group.
	table := abciTruth(t, 8, func(x int) bool {
		return ((x>>0)&1 != 0) != ((x>>3)&1 != 0) != ((x>>6)&1 != 0)
	})
	result, err := AbciBestBoundMask(table, AbciAnalysisOptions{})
	if err != nil || result.Size < 2 || result.Size > 4 || result.Mask>>6 != 0 || result.Evaluations == 0 {
		t.Fatal(result, err)
	}
	for _, options := range []AbciAnalysisOptions{{MaxEvaluations: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		if result, err := AbciBestBoundMask(table, options); err == nil || result.Mask != 0 {
			t.Fatal(result, err)
		}
	}
	if _, err := AbciBestBoundMask(abciTruth(t, 6, func(int) bool { return false }), AbciAnalysisOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	constant := abciTruth(t, 11, func(int) bool { return true })
	best, err := AbciBestBoundMask(constant, AbciAnalysisOptions{})
	if err != nil || best.Size != 4 || best.Mask != 0xf {
		t.Fatal("deterministic corrected bound-only search", best, err)
	}
	for seed := uint64(1); seed <= 64; seed++ {
		state := seed
		random := abciTruth(t, 7, func(int) bool {
			state ^= state << 13
			state ^= state >> 7
			state ^= state << 17
			return state&1 != 0
		})
		got, err := AbciBestBoundMask(random, AbciAnalysisOptions{})
		if err != nil {
			t.Fatal(seed, err)
		}
		wantSize := abciReferenceBestBoundSize(random)
		if got.Size != wantSize || got.Mask>>6 != 0 || bits.OnesCount32(got.Mask) != got.Size && got.Mask != 0 {
			t.Fatal(seed, got, wantSize)
		}
	}
}

func abciReferenceBestBoundSize(table AbciTruthTable) int {
	cofactors := make([]uint32, 64)
	cofBits := 1 << uint(table.Vars()-6)
	for i := range cofactors {
		for k := 0; k < cofBits; k++ {
			if table.value(i*cofBits + k) {
				cofactors[i] |= 1 << uint(k)
			}
		}
	}
	best := 1
	for mask := uint32(1); mask < 64; mask++ {
		size := bits.OnesCount32(mask)
		if size < 2 || size > 4 {
			continue
		}
		var values [64][2]uint32
		var counts [64]uint8
		valid := true
		for i, cof := range cofactors {
			group := i & int(mask)
			if counts[group] == 0 {
				values[group][0], counts[group] = cof, 1
			} else if cof != values[group][0] && (counts[group] == 1 || cof != values[group][1]) {
				if counts[group] == 2 {
					valid = false
					break
				}
				values[group][1], counts[group] = cof, 2
			}
		}
		if valid && size > best {
			best = size
		}
	}
	return best
}

func TestAbciTruthValidation(t *testing.T) {
	words := []uint64{1}
	table, err := NewAbciTruthTable(3, words)
	if err != nil {
		t.Fatal(err)
	}
	words[0] = ^uint64(0)
	if table.Words()[0] != 1 {
		t.Fatal("constructor retained caller storage")
	}
	snapshot := table.Words()
	snapshot[0] = 0
	if table.Words()[0] != 1 || table.Vars() != 3 {
		t.Fatal("snapshot retained table storage")
	}
	for _, tc := range []struct {
		n     int
		words []uint64
	}{{-1, nil}, {17, nil}, {3, nil}, {3, []uint64{1 << 9}}} {
		if _, err := NewAbciTruthTable(tc.n, tc.words); !errors.Is(err, ErrDimension) {
			t.Fatal(tc, err)
		}
	}
}

func TestAbciQuantificationAndUnatenessExhaustive(t *testing.T) {
	for n := 1; n <= 3; n++ {
		for truth := 0; truth < 1<<uint(1<<uint(n)); truth++ {
			table := abciTruth(t, n, func(x int) bool { return truth&(1<<uint(x)) != 0 })
			kinds, err := AbciTruthUnateness(table, AbciAnalysisOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for v := 0; v < n; v++ {
				exists, err := AbciQuantifyVariable(table, v, false, AbciAnalysisOptions{})
				if err != nil {
					t.Fatal(err)
				}
				forall, err := AbciQuantifyVariable(table, v, true, AbciAnalysisOptions{})
				if err != nil {
					t.Fatal(err)
				}
				rise, fall := false, false
				for x := 0; x < 1<<uint(n); x++ {
					a := truth&(1<<uint(x&^(1<<uint(v)))) != 0
					b := truth&(1<<uint(x|(1<<uint(v)))) != 0
					if exists.value(x) != (a || b) || forall.value(x) != (a && b) {
						t.Fatal(n, truth, v, x)
					}
					rise, fall = rise || !a && b, fall || a && !b
				}
				want := AbciIndependent
				if rise {
					want = AbciPositiveUnate
				}
				if fall {
					want = AbciNegativeUnate
				}
				if rise && fall {
					want = AbciBinate
				}
				if kinds[v] != want {
					t.Fatal(n, truth, v, kinds[v], want)
				}
			}
		}
	}
	table := abciTruth(t, 4, func(x int) bool { return x&1 != 0 })
	for _, options := range []AbciAnalysisOptions{{MaxEvaluations: 1}, {MaxBytes: 1}, {Stop: func() bool { return true }}} {
		if got, err := AbciQuantifyVariable(table, 0, false, options); err == nil || got.words != nil {
			t.Fatal(got, err)
		}
		if got, err := AbciTruthUnateness(table, options); err == nil || got != nil {
			t.Fatal(got, err)
		}
	}
}

func TestAbciTruthTableConcurrentReaders(t *testing.T) {
	table := abciTruth(t, 8, func(x int) bool { return x&3 == 1 || x&0x70 == 0x50 })
	const readers = 12
	var wg sync.WaitGroup
	errs := make(chan error, readers)
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(variable int) {
			defer wg.Done()
			if _, err := AnalyzeAbciTruth(table, AbciAnalysisOptions{}); err != nil {
				errs <- err
				return
			}
			quantified, err := AbciQuantifyVariable(table, variable%table.Vars(), variable&1 != 0, AbciAnalysisOptions{})
			if err != nil {
				errs <- err
				return
			}
			words := quantified.Words()
			words[0] ^= 1
			if _, err := AbciTruthUnateness(table, AbciAnalysisOptions{}); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}

func FuzzAbciTruthMeasures(f *testing.F) {
	for _, seed := range []struct {
		n     byte
		truth uint64
	}{{0, 0}, {1, 2}, {3, 0x96}, {6, 0x6996966996696996}} {
		f.Add(seed.n, seed.truth)
	}
	f.Fuzz(func(t *testing.T, rawN byte, truth uint64) {
		n := int(rawN % 7)
		table := abciTruth(t, n, func(x int) bool { return truth&(1<<uint(x)) != 0 })
		got, err := AnalyzeAbciTruth(table, AbciAnalysisOptions{MaxEvaluations: 100000})
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < n; i++ {
			if got.VariableDifferences[i]&1 != 0 || got.VariableDifferences[i] < 0 || got.VariableDifferences[i] > 1<<uint(n) {
				t.Fatal(got)
			}
			for k := 0; k < n; k++ {
				if got.PairDifferences[i][k] != got.PairDifferences[k][i] {
					t.Fatal(got.PairDifferences)
				}
			}
		}
		order, err := AbciVariableActivityOrder(table, AbciAnalysisOptions{MaxEvaluations: 100000})
		if err != nil || len(order) != n {
			t.Fatal(order, err)
		}
	})
}
