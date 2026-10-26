package abc

// Go-only structural checks for DSD merge preparation and purity analysis.

import (
	"errors"
	"slices"
	"testing"
)

func mustParseDauMerge(t *testing.T, text string) *DauDSDFormula {
	t.Helper()
	formula, err := ParseDauDSD(text, DauDSDOptions{})
	if err != nil {
		t.Fatal(text, err)
	}
	return formula
}

func TestDauDSDMergeRemapPresenceAndMaps(t *testing.T) {
	formula := mustParseDauMerge(t, "<abc>{d(ef)g}")
	remapped, err := dauDSDMergeRemap(formula, []int{6, 5, 4, 2, 3, 0, 1}, nil)
	if err != nil || remapped.String() != "<abc>{c(da)b}" || formula.String() != "<abc>{d(ef)g}" {
		t.Fatal(remapped, err, formula.String())
	}
	present, err := dauDSDMergePresence(remapped, 4, nil)
	if err != nil || !slices.Equal(present, []bool{true, true, true, true}) {
		t.Fatal(present, err)
	}
	first := []bool{true, false, true, true, false, true}
	second := []bool{false, true, true, false, false, true}
	shared, oldToNew, newToOld, err := dauDSDMergeMaps(first, second)
	if err != nil || shared != 2 || !slices.Equal(oldToNew, []int{2, 3, 0, 4, -1, 1}) || !slices.Equal(newToOld, []int{2, 5, 0, 1, 3}) {
		t.Fatal(shared, oldToNew, newToOld, err)
	}
	if _, err := dauDSDMergeRemap(formula, []int{0}, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := dauDSDMergePresence(formula, 3, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, _, err := dauDSDMergeMaps([]bool{true}, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestDauDSDMergeStatuses(t *testing.T) {
	tests := []struct {
		text       string
		shared     int
		rootStatus int
	}{
		{"(ab)", 2, 0},
		{"(ab)", 1, 1},
		{"(ab)", 0, 3},
		{"(a(bc)(de))", 1, 2},
		{"<abc>", 1, 2},
		{"96{a(bc)d}", 1, 2},
		{"<abc>{d(ef)g}", 4, 2},
	}
	for _, test := range tests {
		formula := mustParseDauMerge(t, test.text)
		statuses, err := dauDSDMergeStatuses(formula, test.shared, nil)
		if err != nil || statuses[formula.root] != test.rootStatus {
			t.Fatal(test, statuses, err)
		}
		if test.text == "<abc>{d(ef)g}" {
			for _, local := range formula.nodes[formula.root].children {
				if statuses[local] != -1 {
					t.Fatal("local template was classified", statuses)
				}
			}
		}
	}
	formula := mustParseDauMerge(t, "(ab)")
	if _, err := dauDSDMergeStatuses(formula, 3, nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := dauDSDMergeStatuses(formula, 1, func() bool { return true }); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
}

func FuzzDauDSDMergeRemap(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4}, uint8(1))
	f.Add([]byte{9, 8, 7, 6, 5, 4}, uint8(3))
	f.Fuzz(func(t *testing.T, data []byte, shift uint8) {
		if len(data) > 128 {
			data = data[:128]
		}
		formula, err := ParseDauDSD(fuzzDauFormula(data), DauDSDOptions{})
		if err != nil || formula.VariableCount() == 0 || formula.VariableCount() > 10 {
			return
		}
		variables := formula.VariableCount()
		mapping := make([]int, variables)
		for i := range mapping {
			mapping[i] = (i + int(shift)%variables) % variables
		}
		remapped, err := dauDSDMergeRemap(formula, mapping, nil)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			originalAssignment := uint64(0)
			for source, target := range mapping {
				if assignment&(1<<uint(target)) != 0 {
					originalAssignment |= 1 << uint(source)
				}
			}
			before, err := formula.eval(formula.root, originalAssignment, nil)
			if err != nil {
				t.Fatal(err)
			}
			after, err := remapped.eval(remapped.root, uint64(assignment), nil)
			if err != nil || before != after {
				t.Fatal("semantic remap mismatch", formula.String(), remapped.String(), mapping, assignment, err)
			}
		}
	})
}
