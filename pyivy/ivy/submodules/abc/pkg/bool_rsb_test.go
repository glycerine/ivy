package abc

// Go-only tests for src/bool/rsb/rsbDec6.c, rsb.h, and rsbInt.h.

import (
	"errors"
	"math/bits"
	"slices"
	"testing"
)

func rsbTestWord(t testing.TB, variables int, logical uint64) uint64 {
	t.Helper()
	word, err := NormalizeSmallTruthWord(logical, variables)
	if err != nil {
		t.Fatal(err)
	}
	return word
}

func rsbModelPattern(c uint64, functions []uint64) uint32 {
	var result uint32
	for assignment := 0; assignment < 64; assignment++ {
		if c&(uint64(1)<<uint(assignment)) == 0 {
			continue
		}
		pattern := 0
		for i, function := range functions {
			if function&(uint64(1)<<uint(assignment)) != 0 {
				pattern |= 1 << uint(i)
			}
		}
		result |= 1 << uint(pattern)
	}
	return result
}

func rsbModelDecomposition(truth uint64, divisors []uint64, maximum int) ([]int, uint32, bool) {
	for size := 1; size <= maximum; size++ {
		for subset := uint64(1); subset < uint64(1)<<uint(len(divisors)); subset++ {
			if bits.OnesCount64(subset) != size {
				continue
			}
			selected := make([]int, 0, size)
			for index := len(divisors) - 1; index >= 0; index-- {
				if subset&(uint64(1)<<uint(index)) != 0 {
					selected = append(selected, index)
				}
			}
			var zero, one uint32
			for assignment := 0; assignment < 64; assignment++ {
				pattern := 0
				for bit, index := range selected {
					if rsbTruthBit(divisors[index], assignment) {
						pattern |= 1 << uint(bit)
					}
				}
				if rsbTruthBit(truth, assignment) {
					one |= 1 << uint(pattern)
				} else {
					zero |= 1 << uint(pattern)
				}
			}
			if zero&one == 0 {
				return selected, zero | one<<uint(1<<uint(size)), true
			}
		}
	}
	return nil, 0, false
}

func TestRsbDecTryDefinitions(t *testing.T) {
	functions := []uint64{
		truthVariablePattern(0), truthVariablePattern(1), truthVariablePattern(2),
		truthVariablePattern(3), truthVariablePattern(4),
	}
	for _, c := range []uint64{0, 1, 0x5555555555555555, 0x0123456789abcdef, ^uint64(0)} {
		got := []uint32{
			rsbDecTry0(c),
			rsbDecTry1(c, functions[0]),
			rsbDecTry2(c, functions[0], functions[1]),
			rsbDecTry3(c, functions[0], functions[1], functions[2]),
			rsbDecTry4(c, functions[0], functions[1], functions[2], functions[3]),
			rsbDecTry5(c, functions[0], functions[1], functions[2], functions[3], functions[4]),
		}
		for count := range got {
			if want := rsbModelPattern(c, functions[:count]); got[count] != want {
				t.Fatalf("count=%d c=%x got=%x want=%x", count, c, got[count], want)
			}
		}
	}
}

func TestRsbDecompositionCheckAndCounterexample(t *testing.T) {
	a, b := truthVariablePattern(0), truthVariablePattern(1)
	xor := a ^ b
	check, err := CheckRsbDecomposition6(6, xor, []uint64{a, b}, []int{0}, RsbResubOptions{})
	if err != nil || check.Pattern != 0 || check.Counterexample == nil {
		t.Fatalf("conflict: %+v %v", check, err)
	}
	zero, one := check.Counterexample.ZeroAssignment, check.Counterexample.OneAssignment
	if rsbTruthBit(a, zero) != rsbTruthBit(a, one) || rsbTruthBit(xor, zero) || !rsbTruthBit(xor, one) {
		t.Fatalf("bad counterexample %d/%d", zero, one)
	}
	if zero != 0 || one != 2 {
		t.Fatalf("counterexample ordering %d/%d", zero, one)
	}
	check, err = CheckRsbDecomposition6(6, xor, []uint64{a, b}, []int{1, 0}, RsbResubOptions{})
	if err != nil || check.Counterexample != nil || check.Pattern != 0x69 {
		t.Fatalf("decomposition: %+v %v", check, err)
	}
	for _, selected := range [][]int{nil, {0, 0}, {-1}, {2}, {0, 1, 0, 1, 0}} {
		if _, err = CheckRsbDecomposition6(6, xor, []uint64{a, b}, selected, RsbResubOptions{}); !errors.Is(err, ErrDimension) {
			t.Fatalf("selected=%v err=%v", selected, err)
		}
	}
}

func TestRsbSourceExampleAndOrder(t *testing.T) {
	a, b := truthVariablePattern(0), truthVariablePattern(1)
	c, d := truthVariablePattern(2), truthVariablePattern(3)
	e, f := truthVariablePattern(4), truthVariablePattern(5)
	ab, cd, ef := a&b, c&d, e&f
	divisors := []uint64{a, b, c, d, e, f, ab, cd, ef}
	manager, err := NewRsbManager(RsbManagerConfig{LeafMaximum: 6, DivisorMaximum: 64, DecompositionMaximum: 4, Verbose: true})
	if err != nil {
		t.Fatal(err)
	}
	result, err := manager.PerformResub6(6, ab|cd|ef, divisors, RsbResubOptions{})
	if err != nil || !result.Found || !slices.Equal(result.Fanins, []int{8, 7, 6}) {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if valid, verifyErr := VerifyRsbDecomposition(6, ab|cd|ef, divisors, result); verifyErr != nil || !valid {
		t.Fatalf("verify=%v err=%v", valid, verifyErr)
	}
	if result.Format() != "fanins { 8 7 6 } offset=101010101010101 onset=fefefefefefefefe" {
		t.Fatalf("format %q", result.Format())
	}
	if !slices.Equal(manager.Fanins(), result.Fanins) || manager.Statistics().IntermediateTryEntries == 0 {
		t.Fatal("manager scratch was not committed")
	}
	copyResult := result.Clone()
	copyResult.Fanins[0] = 0
	if manager.Fanins()[0] != 8 || result.Fanins[0] != 8 {
		t.Fatal("result aliases manager")
	}

	ordered, _ := NewRsbManager(RsbManagerConfig{LeafMaximum: 6, DivisorMaximum: 4, DecompositionMaximum: 4})
	orderedResult, err := ordered.PerformResub6(6, a, []uint64{b, a, a ^ b}, RsbResubOptions{})
	if err != nil || !slices.Equal(orderedResult.Fanins, []int{1}) {
		t.Fatalf("source order %+v %v", orderedResult, err)
	}
}

func TestRsbAllThreeVariableFunctions(t *testing.T) {
	divisors := []uint64{truthVariablePattern(0), truthVariablePattern(1), truthVariablePattern(2)}
	for logical := 0; logical < 256; logical++ {
		truth := rsbTestWord(t, 3, uint64(logical))
		manager, err := NewRsbManager(RsbManagerConfig{LeafMaximum: 3, DivisorMaximum: 3, DecompositionMaximum: 3})
		if err != nil {
			t.Fatal(err)
		}
		result, err := manager.PerformResub6(3, truth, divisors, RsbResubOptions{})
		if err != nil || !result.Found {
			t.Fatalf("truth=%02x result=%+v err=%v", logical, result, err)
		}
		valid, err := VerifyRsbDecomposition(3, truth, divisors, result)
		if err != nil || !valid {
			t.Fatalf("truth=%02x valid=%v err=%v", logical, valid, err)
		}
	}
}

func TestRsbLimitsCancellationAndAtomicity(t *testing.T) {
	a, b := truthVariablePattern(0), truthVariablePattern(1)
	if _, err := CheckRsbDecomposition6(6, a, []uint64{a}, []int{0}, RsbResubOptions{MaxOperations: 63}); !errors.Is(err, ErrLimit) {
		t.Fatalf("packed-word operation limit: %v", err)
	}
	if check, err := CheckRsbDecomposition6(6, a, []uint64{a}, []int{0}, RsbResubOptions{MaxOperations: 64}); err != nil || check.Counterexample != nil {
		t.Fatalf("packed-word operation accounting: %+v %v", check, err)
	}
	manager, _ := NewRsbManager(RsbManagerConfig{LeafMaximum: 6, DivisorMaximum: 4, DecompositionMaximum: 2})
	good, err := manager.PerformResub6(6, a^b, []uint64{a, b}, RsbResubOptions{})
	if err != nil || !good.Found {
		t.Fatal(good, err)
	}
	beforeFanins, beforeStats := manager.Fanins(), manager.Statistics()
	for _, test := range []struct {
		variables int
		truth     uint64
		divisors  []uint64
		options   RsbResubOptions
		want      error
	}{
		{7, a, []uint64{a}, RsbResubOptions{}, ErrDimension},
		{6, a, []uint64{a}, RsbResubOptions{MaxOperations: -1}, ErrDimension},
		{6, a ^ b, []uint64{a, b}, RsbResubOptions{MaxOperations: 1}, ErrLimit},
		{6, a, []uint64{a}, RsbResubOptions{Stop: func() bool { return true }}, ErrCancelled},
		{2, uint64(2), []uint64{rsbTestWord(t, 2, 2)}, RsbResubOptions{}, ErrInvariant},
	} {
		result, gotErr := manager.PerformResub6(test.variables, test.truth, test.divisors, test.options)
		if gotErr == nil || !errors.Is(gotErr, test.want) || result.Found || result.Fanins != nil || result.Counterexamples != nil || result.Tries != nil {
			t.Fatalf("test=%+v result=%+v err=%v", test, result, gotErr)
		}
		if !slices.Equal(manager.Fanins(), beforeFanins) || manager.Statistics() != beforeStats {
			t.Fatal("failed search mutated manager")
		}
	}
	manager.Close()
	if _, err = manager.PerformResub6(6, a, []uint64{a}, RsbResubOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestRsbIndependentModelCampaign(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		variables := 2 + int(state%5)
		used := 1 + int((state>>8)%uint64(min(4, variables)))
		logicalBits := 1 << uint(used)
		logical := (state >> 16) & (uint64(1)<<uint(logicalBits) - 1)
		truth := rsbTestWord(t, used, logical)
		truth, _ = NormalizeSmallTruthWord(truth, variables)
		divisors := make([]uint64, variables)
		for i := range divisors {
			divisors[i] = truthVariablePattern(i)
		}
		manager, err := NewRsbManager(RsbManagerConfig{LeafMaximum: 6, DivisorMaximum: 6, DecompositionMaximum: 4})
		if err != nil {
			t.Fatal(err)
		}
		result, err := manager.PerformResub6(variables, truth, divisors, RsbResubOptions{})
		wantFanins, wantPattern, wantFound := rsbModelDecomposition(truth, divisors, 4)
		if err != nil || result.Found != wantFound || !slices.Equal(result.Fanins, wantFanins) || result.Pattern != wantPattern {
			t.Fatalf("execution=%d vars=%d used=%d truth=%x result=%+v wantFanins=%v wantPattern=%x wantFound=%v err=%v", execution, variables, used, truth, result, wantFanins, wantPattern, wantFound, err)
		}
		valid, err := VerifyRsbDecomposition(variables, truth, divisors, result)
		if err != nil || !valid {
			t.Fatalf("execution=%d valid=%v err=%v", execution, valid, err)
		}
	}
}

func FuzzRsbResub6(f *testing.F) {
	f.Add(uint64(0), uint8(3))
	f.Add(uint64(0x96), uint8(3))
	f.Add(^uint64(0), uint8(6))
	f.Fuzz(func(t *testing.T, raw uint64, rawVariables uint8) {
		variables := 1 + int(rawVariables%6)
		truth, err := NormalizeSmallTruthWord(raw, variables)
		if err != nil {
			t.Fatal(err)
		}
		divisors := make([]uint64, variables)
		for i := range divisors {
			divisors[i] = truthVariablePattern(i)
		}
		manager, err := NewRsbManager(RsbManagerConfig{LeafMaximum: 6, DivisorMaximum: 6, DecompositionMaximum: 4})
		if err != nil {
			t.Fatal(err)
		}
		result, err := manager.PerformResub6(variables, truth, divisors, RsbResubOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantFanins, wantPattern, wantFound := rsbModelDecomposition(truth, divisors, 4)
		if result.Found != wantFound || !slices.Equal(result.Fanins, wantFanins) || result.Pattern != wantPattern {
			t.Fatalf("result=%+v wantFanins=%v wantPattern=%x wantFound=%v", result, wantFanins, wantPattern, wantFound)
		}
		if result.Found {
			valid, verifyErr := VerifyRsbDecomposition(variables, truth, divisors, result)
			if verifyErr != nil || !valid {
				t.Fatal(valid, verifyErr)
			}
		}
	})
}
