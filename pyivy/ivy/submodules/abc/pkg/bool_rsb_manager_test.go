package abc

import (
	"errors"
	"slices"
	"testing"
)

func TestRsbManagerLifecycleOwnershipAndSizing(t *testing.T) {
	for leaves := 0; leaves <= RsbMaxLeaves; leaves++ {
		config := RsbManagerConfig{LeafMaximum: leaves, DivisorMaximum: RsbMaxDivisors, DecompositionMaximum: 4, Verbose: leaves&1 != 0}
		manager, err := NewRsbManager(config)
		if err != nil {
			t.Fatal(leaves, err)
		}
		words := 1
		if leaves > 6 {
			words = 1 << uint(leaves-6)
		}
		stats := manager.Statistics()
		if stats.Config != config || stats.CounterexampleCapacity != RsbMaxDivisors+150 || stats.PatternWordCapacity != words || stats.Fanins != 0 || stats.OriginalFanins != 0 {
			t.Fatal(leaves, stats)
		}
		if got, err := manager.Config(); err != nil || got != config {
			t.Fatal(got, err)
		}
		manager.Close()
		manager.Close()
		if manager.Fanins() != nil || manager.OriginalFanins() != nil || manager.Statistics() != (RsbManagerStats{}) {
			t.Fatal("closed manager exposes storage")
		}
		if _, err := manager.Config(); !errors.Is(err, ErrInvariant) {
			t.Fatal("closed config", err)
		}
	}
}

func TestRsbManagerFaninsAndErrors(t *testing.T) {
	manager, err := NewRsbManager(RsbManagerConfig{LeafMaximum: 6, DivisorMaximum: 8, DecompositionMaximum: 4})
	if err != nil {
		t.Fatal(err)
	}
	current := []int{7, 2, 0, 4}
	original := []int{1, 3, 5}
	if err = manager.SetFanins(current); err != nil {
		t.Fatal(err)
	}
	if err = manager.SetOriginalFanins(original); err != nil {
		t.Fatal(err)
	}
	current[0], original[0] = 0, 0
	if !slices.Equal(manager.Fanins(), []int{7, 2, 0, 4}) || !slices.Equal(manager.OriginalFanins(), []int{1, 3, 5}) {
		t.Fatal("setter storage aliases caller")
	}
	copyFanins := manager.Fanins()
	copyFanins[0] = 1
	if manager.Fanins()[0] != 7 {
		t.Fatal("accessor storage aliases caller")
	}
	before := manager.Statistics()
	for _, test := range []struct {
		values   []int
		original bool
		want     error
	}{
		{[]int{-1}, false, ErrDimension},
		{[]int{8}, true, ErrDimension},
		{make([]int, 9), false, ErrLimit},
	} {
		if test.original {
			err = manager.SetOriginalFanins(test.values)
		} else {
			err = manager.SetFanins(test.values)
		}
		if !errors.Is(err, test.want) || manager.Statistics() != before {
			t.Fatal(test, err, manager.Statistics(), before)
		}
	}
	for _, config := range []RsbManagerConfig{
		{LeafMaximum: -1}, {LeafMaximum: RsbMaxLeaves + 1},
		{DivisorMaximum: -1}, {DivisorMaximum: RsbMaxDivisors + 1},
		{DecompositionMaximum: -1},
	} {
		if manager, err := NewRsbManager(config); manager != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(config, manager, err)
		}
	}
	if (*RsbManager)(nil).SetFanins(nil) == nil {
		t.Fatal("nil manager")
	}
}

func FuzzRsbManagerFanins(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {7, 2, 0, 4}, {255, 254, 253}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > RsbMaxDivisors {
			return
		}
		manager, err := NewRsbManager(RsbManagerConfig{LeafMaximum: int(uint8(len(data))) % 21, DivisorMaximum: RsbMaxDivisors, DecompositionMaximum: 4})
		if err != nil {
			t.Fatal(err)
		}
		values := make([]int, len(data))
		for i, value := range data {
			values[i] = int(value) % RsbMaxDivisors
		}
		if err = manager.SetFanins(values); err != nil {
			t.Fatal(err)
		}
		got := manager.Fanins()
		if !slices.Equal(got, values) {
			t.Fatal(got, values)
		}
		if len(got) != 0 {
			got[0] ^= 1
			if slices.Equal(got, manager.Fanins()) {
				t.Fatal("returned alias")
			}
		}
	})
}
