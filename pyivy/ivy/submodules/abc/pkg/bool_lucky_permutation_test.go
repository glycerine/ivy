package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthPermutationScheduleNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_permutation_schedule.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables    int
		Swaps, Flips []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 9 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		schedule, e := NewTruthPermutationSchedule(row.Variables, 0)
		if e != nil || schedule.Variables() != row.Variables || !slices.Equal(schedule.AdjacentSwaps(), row.Swaps) || !slices.Equal(schedule.GrayFlips(), row.Flips) {
			t.Fatal(row.Variables, e)
		}
		checkTruthPermutationWalk(t, schedule)
	}
}
func checkTruthPermutationWalk(t *testing.T, schedule *TruthPermutationSchedule) {
	t.Helper()
	n := schedule.Variables()
	labels := make([]byte, n)
	for i := range labels {
		labels[i] = byte(i)
	}
	seen := map[string]bool{string(labels): true}
	for _, index := range schedule.AdjacentSwaps() {
		if index < 0 || index >= n-1 {
			t.Fatal("swap index", n, index)
		}
		labels[index], labels[index+1] = labels[index+1], labels[index]
		if seen[string(labels)] {
			t.Fatal("repeated permutation", n, labels)
		}
		seen[string(labels)] = true
	}
	factorial := 1
	for i := 2; i <= n; i++ {
		factorial *= i
	}
	if len(seen) != factorial {
		t.Fatal("incomplete permutations", n, len(seen), factorial)
	}
	phases := make([]bool, 1<<uint(n))
	phases[0] = true
	mask := 0
	for step, index := range schedule.GrayFlips() {
		if index < 0 || index >= n {
			t.Fatal("phase index", n, index)
		}
		mask ^= 1 << uint(index)
		expected := (step + 1) ^ ((step + 1) >> 1)
		if mask != expected || phases[mask] {
			t.Fatal("Gray step", n, step, mask, expected)
		}
		phases[mask] = true
	}
	for mask, visited := range phases {
		if !visited {
			t.Fatal("missing phase", n, mask)
		}
	}
}
func TestTruthPermutationScheduleLimitsAndOwnership(t *testing.T) {
	for _, tc := range []struct {
		variables, budget int
		want              error
	}{
		{-1, 0, ErrDimension}, {37, 0, ErrDimension}, {3, -1, ErrDimension},
		{3, 11, ErrLimit}, {12, 0, ErrLimit}, {36, int(^uint(0) >> 1), ErrLimit},
	} {
		if r, e := NewTruthPermutationSchedule(tc.variables, tc.budget); r != nil || !errors.Is(e, tc.want) {
			t.Fatal(tc, r, e)
		}
	}
	schedule, e := NewTruthPermutationSchedule(3, 12)
	if e != nil {
		t.Fatal(e)
	}
	swaps, flips := schedule.AdjacentSwaps(), schedule.GrayFlips()
	swaps[0] = 99
	flips[0] = 99
	if schedule.AdjacentSwaps()[0] == 99 || schedule.GrayFlips()[0] == 99 {
		t.Fatal("accessor aliases")
	}
	if (*TruthPermutationSchedule)(nil).Variables() != 0 || (*TruthPermutationSchedule)(nil).AdjacentSwaps() != nil || (*TruthPermutationSchedule)(nil).GrayFlips() != nil {
		t.Fatal("nil accessors")
	}
	checkTruthPermutationWalk(t, &TruthPermutationSchedule{})
}
func FuzzTruthPermutationSchedule(f *testing.F) {
	f.Add(uint8(3), uint16(12))
	f.Add(uint8(7), uint16(8191))
	f.Add(uint8(0), uint16(1))
	f.Fuzz(func(t *testing.T, width uint8, budget uint16) {
		variables := int(width % 11)
		limit := 1 + int(budget%8192)
		factorial := uint64(1)
		for i := 2; i <= variables; i++ {
			factorial *= uint64(i)
		}
		count := factorial - 1 + (uint64(1) << uint(variables)) - 1
		schedule, e := NewTruthPermutationSchedule(variables, limit)
		if count > uint64(limit) {
			if schedule != nil || !errors.Is(e, ErrLimit) {
				t.Fatal("missing limit", variables, limit, e)
			}
			return
		}
		if e != nil || len(schedule.swaps)+len(schedule.flips) != int(count) {
			t.Fatal(variables, limit, e)
		}
		checkTruthPermutationWalk(t, schedule)
	})
}
