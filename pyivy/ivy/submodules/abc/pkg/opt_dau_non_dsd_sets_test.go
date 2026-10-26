package abc

// Go-only unit and independent semantic fuzz checks for non-DSD set discovery.

import (
	"errors"
	"math/bits"
	"slices"
	"strings"
	"testing"
)

func dauTruthFromPredicate(t *testing.T, variables int, predicate func(uint64) bool) *TruthTable {
	t.Helper()
	table, err := NewConstantTruthTable(variables, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	for assignment := uint64(0); assignment < 1<<uint(variables); assignment++ {
		if predicate(assignment) {
			if err := table.Set(assignment, true); err != nil {
				t.Fatal(err)
			}
		}
	}
	return table
}

func TestDauDecompositionSetEncodingAndFormatting(t *testing.T) {
	set, err := ParseDauDecompositionSet("Bcd", 6)
	if err != nil {
		t.Fatal(err)
	}
	stats, err := set.Stats(6)
	if err != nil || stats != (DauDecompositionSetStats{Unique: 2, Shared: 1, Free: 3}) {
		t.Fatal(stats, err)
	}
	roles := []DauDecompositionRole{DauDecompositionFree, DauDecompositionShared, DauDecompositionUnique, DauDecompositionUnique, DauDecompositionFree, DauDecompositionFree}
	for variable, want := range roles {
		got, err := set.Role(variable)
		if err != nil || got != want {
			t.Fatal(variable, got, err)
		}
	}
	line, err := FormatDauDecompositionSet(set, 6)
	if err != nil || !strings.HasPrefix(line, "S = 1  D = 3  C = 5   x=Bcd y=xaBef") {
		t.Fatal(line, err)
	}
	listing, err := FormatDauDecompositionSets([]DauDecompositionSet{set}, 6)
	if err != nil || !strings.HasPrefix(listing, "The 6-variable set family contains 1 sets:\n"+line+"\n") {
		t.Fatal(listing, err)
	}
	for _, text := range []string{"bb", "A", "a?"} {
		if _, err := ParseDauDecompositionSet(text, 0); err == nil {
			t.Fatal(text)
		}
	}
	if _, err := (DauDecompositionSet(2)).Stats(1); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestFindDauDecompositionSetsAndAndMixed(t *testing.T) {
	and := dauTruthFromPredicate(t, 4, func(assignment uint64) bool { return bits.OnesCount64(assignment) == 4 })
	sets, err := FindDauDecompositionSets(and, DauDecompositionSetOptions{})
	if err != nil || len(sets) != 10 {
		t.Fatal(sets, err)
	}
	want := make([]DauDecompositionSet, 0, 10)
	for size := 2; size < 4; size++ {
		for mask := 0; mask < 16; mask++ {
			if bits.OnesCount(uint(mask)) != size {
				continue
			}
			var set uint32
			for variable := 0; variable < 4; variable++ {
				if mask&(1<<uint(variable)) != 0 {
					set |= uint32(DauDecompositionUnique) << uint(2*variable)
				}
			}
			want = append(want, DauDecompositionSet(set))
		}
	}
	if !slices.Equal(sets, want) {
		t.Fatal(sets, want)
	}

	mixedTruth := dauTruthFromPredicate(t, 4, func(assignment uint64) bool {
		a := assignment&1 != 0
		b := assignment&2 != 0
		c := assignment&4 != 0
		d := assignment&8 != 0
		selector := d
		if b {
			selector = c
		}
		if selector {
			return a != b
		}
		return b
	})
	sets, err = FindDauDecompositionSets(mixedTruth, DauDecompositionSetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	mixed, _ := ParseDauDecompositionSet("Bcd", 4)
	if !slices.Contains(sets, mixed) {
		t.Fatal("missing mixed set", sets)
	}
}

func TestFindDauDecompositionSetsErrorsAndOwnership(t *testing.T) {
	table := dauTruthFromPredicate(t, 4, func(assignment uint64) bool { return assignment&15 == 15 })
	if result, err := FindDauDecompositionSets(table, DauDecompositionSetOptions{MaxSets: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := FindDauDecompositionSets(table, DauDecompositionSetOptions{MaxChecks: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := FindDauDecompositionSets(table, DauDecompositionSetOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	unused := dauTruthFromPredicate(t, 4, func(assignment uint64) bool { return assignment&7 == 7 })
	if result, err := FindDauDecompositionSets(unused, DauDecompositionSetOptions{}); result != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(result, err)
	}
	sets, err := FindDauDecompositionSets(table, DauDecompositionSetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	sets[0] = 0
	again, err := FindDauDecompositionSets(table, DauDecompositionSetOptions{})
	if err != nil || again[0] == 0 {
		t.Fatal(again, err)
	}
}

func modelDauSetWorks(table *TruthTable, set DauDecompositionSet) bool {
	free, unique, shared := []int{}, []int{}, []int{}
	for variable := 0; variable < table.variables; variable++ {
		role, _ := set.Role(variable)
		switch role {
		case DauDecompositionFree:
			free = append(free, variable)
		case DauDecompositionUnique:
			unique = append(unique, variable)
		case DauDecompositionShared:
			shared = append(shared, variable)
		}
	}
	for sharedBits := 0; sharedBits < 1<<uint(len(shared)); sharedBits++ {
		cofactors := make(map[string]struct{})
		for uniqueBits := 0; uniqueBits < 1<<uint(len(unique)); uniqueBits++ {
			values := make([]byte, 1<<uint(len(free)))
			for freeBits := range values {
				assignment := uint64(0)
				for i, variable := range shared {
					assignment |= uint64((sharedBits>>uint(i))&1) << uint(variable)
				}
				for i, variable := range unique {
					assignment |= uint64((uniqueBits>>uint(i))&1) << uint(variable)
				}
				for i, variable := range free {
					assignment |= uint64((freeBits>>uint(i))&1) << uint(variable)
				}
				if value, _ := table.Get(assignment); value {
					values[freeBits] = 1
				}
			}
			cofactors[string(values)] = struct{}{}
			if len(cofactors) > 2 {
				return false
			}
		}
	}
	return true
}

func FuzzFindDauDecompositionSets(f *testing.F) {
	f.Add(uint8(4), []byte{0x96, 0x69})
	f.Add(uint8(5), []byte{0x31, 0x72, 0x55, 0xaa})
	f.Fuzz(func(t *testing.T, width uint8, data []byte) {
		variables := 3 + int(width%4)
		if len(data) == 0 {
			return
		}
		table, err := NewConstantTruthTable(variables, false, 0)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			value := data[assignment%len(data)]&(1<<uint(assignment&7)) != 0
			if value {
				_ = table.Set(uint64(assignment), true)
			}
		}
		support, err := table.Support()
		if err != nil || len(support) != variables {
			return
		}
		sets, err := FindDauDecompositionSets(table, DauDecompositionSetOptions{MaxChecks: 1 << 24})
		if err != nil {
			t.Fatal(err)
		}
		for _, set := range sets {
			if !modelDauSetWorks(table, set) {
				t.Fatal("invalid set", set)
			}
		}
		for boundSize := 2; boundSize < variables; boundSize++ {
			for uniqueMask := 0; uniqueMask < 1<<uint(variables); uniqueMask++ {
				if bits.OnesCount(uint(uniqueMask)) != boundSize {
					continue
				}
				for sharedMask := uniqueMask; ; sharedMask = (sharedMask - 1) & uniqueMask {
					sharedCount := bits.OnesCount(uint(sharedMask))
					if sharedCount <= 3 && sharedCount <= boundSize-2 {
						var candidate uint32
						for variable := 0; variable < variables; variable++ {
							if uniqueMask&(1<<uint(variable)) != 0 {
								role := uint32(DauDecompositionUnique)
								if sharedMask&(1<<uint(variable)) != 0 {
									role = uint32(DauDecompositionShared)
								}
								candidate |= role << uint(2*variable)
							}
						}
						set := DauDecompositionSet(candidate)
						if modelDauSetWorks(table, set) && !slices.Contains(sets, set) && !dauDecompositionSetContained(sets, set) {
							t.Fatal("uncovered decomposable set", set, sets)
						}
					}
					if sharedMask == 0 {
						break
					}
				}
			}
		}
	})
}
