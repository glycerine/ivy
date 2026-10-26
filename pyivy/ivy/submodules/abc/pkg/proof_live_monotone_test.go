package abc

// Standalone tests for monotone.c and disjunctiveMonotone.c helpers.

import (
	"errors"
	"slices"
	"sync"
	"testing"
)

func TestLiveCompiledSourceInventory(t *testing.T) {
	want := []string{"liveness.c", "liveness_sim.c", "ltl_parser.c", "kliveness.c", "monotone.c", "disjunctiveMonotone.c", "arenaViolation.c", "kLiveConstraints.c", "combination.c"}
	got := LiveCompiledSourceInventory()
	if !slices.Equal(got, want) {
		t.Fatalf("inventory=%v", got)
	}
	got[0] = "changed"
	if LiveCompiledSourceInventory()[0] != "liveness.c" {
		t.Fatal("inventory aliases caller storage")
	}
}

func TestLiveFindMonotoneOutputs(t *testing.T) {
	names := []string{"safe", "x_pendingSignal_suffix", "hint_a", "ordinary", "pre_hint_b", "pendingSignal_second"}
	got, err := LiveFindMonotoneOutputs(names, LiveMonotoneOptions{})
	if err != nil || got.Pending != 1 || got.HintBegin != 2 || got.HintEnd != 4 || !slices.Equal(got.Hints, []int{2, 4}) {
		t.Fatalf("got=%+v err=%v", got, err)
	}
	empty, err := LiveFindMonotoneOutputs([]string{"x"}, LiveMonotoneOptions{})
	if err != nil || empty.Pending != -1 || empty.HintBegin != -1 || empty.HintEnd != -1 || len(empty.Hints) != 0 {
		t.Fatalf("empty=%+v err=%v", empty, err)
	}
}

func TestLiveMonotoneVectorHelpers(t *testing.T) {
	diff, err := LiveVectorDifference([]int{3, 1, 3, 4, 2}, []int{2, 3}, LiveMonotoneOptions{})
	if err != nil || !slices.Equal(diff, []int{1, 4}) {
		t.Fatalf("diff=%v err=%v", diff, err)
	}
	remaining, err := LiveRemainingMonotoneCandidates([]int{4, 1}, []int{1, 2, 4, 2, 5}, LiveMonotoneOptions{})
	if err != nil || !slices.Equal(remaining, []int{2, 2, 5}) {
		t.Fatalf("remaining=%v err=%v", remaining, err)
	}
	consequents, err := LiveConsequentCandidates([]int{0, 2}, []int{0, 1, 2, 3}, LiveMonotoneOptions{})
	if err != nil || !slices.Equal(consequents, []int{1, 3}) {
		t.Fatalf("consequents=%v err=%v", consequents, err)
	}
	values := []int{1, 2}
	got, present, err := LivePushUnique(values, 2, LiveMonotoneOptions{})
	if err != nil || !present || !slices.Equal(got, values) {
		t.Fatalf("existing=%v %v %v", got, present, err)
	}
	got[0] = 9
	if values[0] != 1 {
		t.Fatal("result aliases input")
	}
	got, present, err = LivePushUnique(values, 3, LiveMonotoneOptions{})
	if err != nil || present || !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("new=%v %v %v", got, present, err)
	}
}

func TestLiveMonotoneLimitsCancellationAtomicity(t *testing.T) {
	for _, tc := range []struct {
		options LiveMonotoneOptions
		want    error
	}{
		{LiveMonotoneOptions{MaxItems: -1}, ErrDimension},
		{LiveMonotoneOptions{MaxItems: 1}, ErrLimit},
		{LiveMonotoneOptions{MaxWork: 1}, ErrLimit},
		{LiveMonotoneOptions{MaxBytes: 1}, ErrLimit},
		{LiveMonotoneOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		got, err := LiveVectorDifference([]int{1, 2}, []int{2}, tc.options)
		if !errors.Is(err, tc.want) || got != nil {
			t.Fatalf("options=%+v got=%v err=%v", tc.options, got, err)
		}
	}
	if got, err := LiveVectorDifference([]int{-1}, nil, LiveMonotoneOptions{}); !errors.Is(err, ErrDimension) || got != nil {
		t.Fatalf("negative got=%v err=%v", got, err)
	}
	calls := 0
	got, err := LiveVectorDifference([]int{1, 2, 3, 4}, []int{8, 9}, LiveMonotoneOptions{Stop: func() bool {
		calls++
		return calls == 8
	}})
	if !errors.Is(err, ErrCancelled) || got != nil {
		t.Fatalf("mid cancel got=%v err=%v calls=%d", got, err, calls)
	}
	if got, err := LiveFindMonotoneOutputs([]string{"hint_a", "hint_b"}, LiveMonotoneOptions{MaxBytes: 8}); !errors.Is(err, ErrLimit) || got.Hints != nil {
		t.Fatalf("byte limit got=%+v err=%v", got, err)
	}
}

func TestLiveMonotoneConcurrentReaders(t *testing.T) {
	first := []int{0, 1, 2, 3, 4, 5}
	second := []int{1, 3, 5}
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				got, err := LiveVectorDifference(first, second, LiveMonotoneOptions{})
				if err != nil || !slices.Equal(got, []int{0, 2, 4}) {
					t.Errorf("got=%v err=%v", got, err)
					return
				}
			}
		}()
	}
	wg.Wait()
}

func FuzzLiveMonotoneHelpers(f *testing.F) {
	for _, seed := range []struct {
		first  []byte
		second []byte
	}{{nil, nil}, {[]byte{1, 2, 1, 3}, []byte{1}}, {[]byte{0, 255}, []byte{255}}} {
		f.Add(seed.first, seed.second)
	}
	f.Fuzz(func(t *testing.T, rawFirst, rawSecond []byte) {
		if len(rawFirst) > 32 || len(rawSecond) > 32 {
			t.Skip()
		}
		first := make([]int, len(rawFirst))
		second := make([]int, len(rawSecond))
		for i, value := range rawFirst {
			first[i] = int(value)
		}
		for i, value := range rawSecond {
			second[i] = int(value)
		}
		got, err := LiveVectorDifference(first, second, LiveMonotoneOptions{MaxItems: 64, MaxWork: 4096, MaxBytes: 1024})
		if err != nil {
			t.Fatal(err)
		}
		want := make([]int, 0, len(first))
		for _, value := range first {
			if !slices.Contains(second, value) {
				want = append(want, value)
			}
		}
		if !slices.Equal(got, want) {
			t.Fatalf("got=%v want=%v", got, want)
		}
	})
}
