package minimize

import (
	"testing"
)

func bruteMinimumCost(variables int, function uint64) (int, int) {
	all := (uint64(1) << variables) - 1
	type candidate struct {
		cube Cube
		set  uint64
	}
	var candidates []candidate
	for mask := uint64(0); mask <= all; mask++ {
		for value := uint64(0); value <= all; value++ {
			if value&^mask != 0 {
				continue
			}
			var set uint64
			valid := true
			for m := uint64(0); m <= all; m++ {
				if (m^value)&mask != 0 {
					continue
				}
				if function&(1<<m) == 0 {
					valid = false
					break
				}
				set |= 1 << m
			}
			if valid && set != 0 {
				candidates = append(candidates, candidate{Cube{Mask: mask, Bits: value}, set})
			}
		}
	}
	bestLits, bestTerms := 1<<30, 1<<30
	var search func(covered uint64, lits, terms int)
	search = func(covered uint64, lits, terms int) {
		if lits > bestLits || lits == bestLits && terms >= bestTerms {
			return
		}
		if covered == function {
			bestLits, bestTerms = lits, terms
			return
		}
		uncovered := function &^ covered
		pivot := uint64(1) << uint64(0)
		for pivot&uncovered == 0 {
			pivot <<= 1
		}
		for _, c := range candidates {
			if c.set&pivot != 0 {
				search(covered|c.set, lits+variablesOnes(c.cube.Mask), terms+1)
			}
		}
	}
	if function == 0 {
		return 0, 0
	}
	search(0, 0, 0)
	return bestLits, bestTerms
}

func variablesOnes(value uint64) int {
	count := 0
	for value != 0 {
		value &= value - 1
		count++
	}
	return count
}

func truth(cubes []Cube, minterm uint64) bool {
	for _, cube := range cubes {
		if cube.Matches(minterm) {
			return true
		}
	}
	return false
}

func TestMinimizeEveryThreeVariableFunction(t *testing.T) {
	for function := uint64(0); function < 1<<8; function++ {
		on := make([]uint64, 0, 8)
		for m := uint64(0); m < 8; m++ {
			if function&(1<<m) != 0 {
				on = append(on, m)
			}
		}
		cubes, err := Minimize(3, on, nil)
		if err != nil {
			t.Fatalf("function %#02x: Minimize: %v", function, err)
		}
		for m := uint64(0); m < 8; m++ {
			want := function&(1<<m) != 0
			if got := truth(cubes, m); got != want {
				t.Fatalf("function %#02x at minterm %d: got %v, want %v; cubes=%v", function, m, got, want, cubes)
			}
		}
		wantLits, wantTerms := bruteMinimumCost(3, function)
		gotLits, gotTerms := 0, len(cubes)
		for _, cube := range cubes {
			gotLits += variablesOnes(cube.Mask)
		}
		if gotLits != wantLits || gotTerms != wantTerms {
			t.Fatalf("function %#02x: cover cost (%d literals,%d cubes), want (%d,%d)", function, gotLits, gotTerms, wantLits, wantTerms)
		}
	}
}

func TestMinimizeUsesDontCareAndFindsMinimum(t *testing.T) {
	// f(a,b,c) is a for all values except minterm 1, which is unspecified.
	cubes, err := Minimize(3, []uint64{4, 5, 6, 7}, []uint64{1})
	if err != nil {
		t.Fatal(err)
	}
	if len(cubes) != 1 || cubes[0] != (Cube{Mask: 4, Bits: 4}) {
		t.Fatalf("got cover %v; want single cube a", cubes)
	}
	for _, m := range []uint64{4, 5, 6, 7} {
		if !truth(cubes, m) {
			t.Errorf("cover misses required minterm %d", m)
		}
	}
	for _, m := range []uint64{0, 2, 3} {
		if truth(cubes, m) {
			t.Errorf("cover includes off-set minterm %d", m)
		}
	}
}

func TestMinimizeHasStableTieBreak(t *testing.T) {
	on := []uint64{0, 1, 2, 5, 6, 7}
	want, err := Minimize(3, on, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		got, err := Minimize(3, on, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(want) {
			t.Fatalf("run %d changed cover size: %v vs %v", i, got, want)
		}
		for j := range got {
			if got[j] != want[j] {
				t.Fatalf("run %d changed cover: %v vs %v", i, got, want)
			}
		}
	}
}

func TestMinimizeConstantAndInvalidInput(t *testing.T) {
	cubes, err := Minimize(0, []uint64{0}, nil)
	if err != nil || len(cubes) != 1 || cubes[0] != (Cube{}) {
		t.Fatalf("constant one: cubes=%v err=%v", cubes, err)
	}
	cubes, err = Minimize(3, nil, nil)
	if err != nil || len(cubes) != 0 {
		t.Fatalf("constant zero: cubes=%v err=%v", cubes, err)
	}
	if _, err := Minimize(3, []uint64{8}, nil); err == nil {
		t.Fatal("expected out-of-range minterm error")
	}
	if _, err := Minimize(17, nil, nil); err == nil {
		t.Fatal("expected variable limit error")
	}
}
