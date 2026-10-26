package abc

// Go-only reconstruction and fuzz checks for contiguous bound-set decomposition.

import (
	"errors"
	"testing"
)

func truthFromBits(t *testing.T, variables int, bits uint64) *TruthTable {
	t.Helper()
	result, err := NewConstantTruthTable(variables, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < uint64(1)<<uint(variables); assignment++ {
		if bits>>assignment&1 != 0 {
			if err = result.Set(assignment, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return result
}

func TestDuaDSDRangeReconstruction(t *testing.T) {
	base := truthFromBits(t, 3, 0x69)
	reduced, err := NewConstantTruthTable(4, false, 1)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 16; assignment++ {
		low := assignment & 1
		selector := assignment >> 1 & 1
		high := assignment >> 2
		value, _ := base.Get(low | high<<1)
		if value != (selector != 0) {
			if err = reduced.Set(assignment, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	for mask := uint64(1); mask < 1<<8; mask++ {
		if mask&1 != 0 {
			continue
		}
		original, err := DuaComposeRange(reduced, 1, 4, mask, 0)
		if err != nil {
			t.Fatal(err)
		}
		decomposition, err := DuaDSDRange(original, 1, 4, DuaRangeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if !decomposition.Found || decomposition.Mask != mask {
			t.Fatalf("mask %08b => found=%t mask=%08b", mask, decomposition.Found, decomposition.Mask)
		}
		again, err := DuaComposeRange(decomposition.Reduced, 1, 4, decomposition.Mask, 0)
		if err != nil || !again.Equal(original) {
			t.Fatalf("mask %08b reconstruction: %v", mask, err)
		}
	}
	nondecomposable := truthFromBits(t, 5, 0x96e3a75c)
	if result, err := DuaDSDRange(nondecomposable, 1, 4, DuaRangeOptions{}); err != nil || result.Found {
		t.Fatalf("nondecomposable: %+v, %v", result, err)
	}
	original, err := DuaComposeRange(reduced, 1, 4, 0x6a, 0)
	if err != nil {
		t.Fatal(err)
	}
	third := original.Clone()
	value, _ := third.Get(6) // low=0, bound=3, high=0
	if err = third.Set(6, !value); err != nil {
		t.Fatal(err)
	}
	if result, err := DuaDSDRange(third, 1, 4, DuaRangeOptions{}); err != nil || result.Found {
		t.Fatalf("third cofactor: %+v, %v", result, err)
	}
	inconsistent := original.Clone()
	for low := uint64(0); low < 2; low++ {
		assignment := low | 1<<1 | 1<<4 // bound=1 in outer partition one
		value, _ := inconsistent.Get(assignment)
		if err = inconsistent.Set(assignment, !value); err != nil {
			t.Fatal(err)
		}
	}
	if result, err := DuaDSDRange(inconsistent, 1, 4, DuaRangeOptions{}); err != nil || result.Found {
		t.Fatalf("inconsistent mask: %+v, %v", result, err)
	}
}

func TestDuaDSDRangeZeroReconstruction(t *testing.T) {
	bound := truthFromBits(t, 4, 0x96e3)
	phase := truthFromBits(t, 3, 0x6a)
	original, err := DuaComposeRangeZero(bound, phase, 0)
	if err != nil {
		t.Fatal(err)
	}
	result, err := DuaDSDRangeZero(original, 4, DuaRangeOptions{})
	if err != nil || !result.Found {
		t.Fatalf("decompose: %+v, %v", result, err)
	}
	again, err := DuaComposeRangeZero(result.BoundFunction, result.Phase, 0)
	if err != nil || !again.Equal(original) {
		t.Fatalf("reconstruct: %v", err)
	}

	broken := original.Clone()
	value, _ := broken.Get(37)
	if err = broken.Set(37, !value); err != nil {
		t.Fatal(err)
	}
	if result, err := DuaDSDRangeZero(broken, 4, DuaRangeOptions{}); err != nil || result.Found {
		t.Fatalf("broken: %+v, %v", result, err)
	}
}

func TestDuaDSDRangeBoundsAndCancellation(t *testing.T) {
	table := truthFromBits(t, 6, 0x9ef7a8d9c7193a0f)
	for _, bounds := range [][2]int{{-1, 3}, {0, 2}, {0, 7}, {1, 8}} {
		if _, err := DuaDSDRange(table, bounds[0], bounds[1], DuaRangeOptions{}); !errors.Is(err, ErrDimension) {
			t.Errorf("bounds %v: %v", bounds, err)
		}
	}
	if result, err := DuaDSDRange(table, 0, 3, DuaRangeOptions{Stop: func() bool { return true }}); result.Found || !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel: %+v, %v", result, err)
	}
	if _, err := DuaDSDRange(table, 0, 3, DuaRangeOptions{MaxWords: -1}); !errors.Is(err, ErrDimension) {
		t.Fatalf("negative: %v", err)
	}
	for _, bound := range []int{2, 6, 7} {
		if _, err := DuaDSDRangeZero(table, bound, DuaRangeOptions{}); !errors.Is(err, ErrDimension) {
			t.Errorf("zero bound %d: %v", bound, err)
		}
	}
}

func FuzzDuaDSDRange(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 0x9ef7a8d9c7193a0f, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		// Make both selector cofactors differ for every high assignment.
		base := truthFromBits(t, 3, seed)
		reduced, err := NewConstantTruthTable(4, false, 1)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := uint64(0); assignment < 16; assignment++ {
			low := assignment & 1
			selector := assignment >> 1 & 1
			high := assignment >> 2
			value, _ := base.Get(low | high<<1)
			if value != (selector != 0) {
				if err = reduced.Set(assignment, true); err != nil {
					t.Fatal(err)
				}
			}
		}
		mask := seed >> 8 & 0xfe
		if mask == 0 {
			mask = 2
		}
		original, err := DuaComposeRange(reduced, 1, 4, mask, 0)
		if err != nil {
			t.Fatal(err)
		}
		result, err := DuaDSDRange(original, 1, 4, DuaRangeOptions{})
		if err != nil || !result.Found || result.Mask != mask {
			t.Fatalf("mask %02x: %+v, %v", mask, result, err)
		}
		again, err := DuaComposeRange(result.Reduced, 1, 4, result.Mask, 0)
		if err != nil || !again.Equal(original) {
			t.Fatalf("reconstruction: %v", err)
		}
	})
}

func FuzzDuaDSDRangeZero(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 0x123456789abcdef0, ^uint64(0)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, seed uint64) {
		bound := truthFromBits(t, 4, uint64(uint16(seed)))
		// Normalize phase zero so the recovered bound function has the same phase.
		phaseBits := seed >> 16 & 0xfe
		phase := truthFromBits(t, 3, phaseBits)
		original, err := DuaComposeRangeZero(bound, phase, 0)
		if err != nil {
			t.Fatal(err)
		}
		result, err := DuaDSDRangeZero(original, 4, DuaRangeOptions{})
		if err != nil || !result.Found || !result.BoundFunction.Equal(bound) || !result.Phase.Equal(phase) {
			t.Fatalf("decompose: %+v, %v", result, err)
		}
		again, err := DuaComposeRangeZero(result.BoundFunction, result.Phase, 0)
		if err != nil || !again.Equal(original) {
			t.Fatalf("reconstruction: %v", err)
		}
	})
}
