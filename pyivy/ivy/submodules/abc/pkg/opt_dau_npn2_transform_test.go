package abc

// Go-only assignment models for DTT packed NPN configurations and gate types.

import (
	"errors"
	"slices"
	"testing"
)

func TestDttNPNTransformComposition(t *testing.T) {
	first, err := NewDttNPNTransform([]int{1, 2, 0}, []bool{true, false, true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewDttNPNTransform([]int{2, 0, 1}, []bool{false, true, false})
	if err != nil {
		t.Fatal(err)
	}
	composed, err := ComposeDttNPNTransforms(first, second)
	if err != nil {
		t.Fatal(err)
	}
	truth := uint64(0x96)
	firstTruth, _ := first.ApplyTruth(truth)
	want, _ := second.ApplyTruth(firstTruth)
	got, err := composed.ApplyTruth(truth)
	if err != nil || got != want {
		t.Fatal(got, want, composed.Permutation(), composed.InputPhases(), err)
	}
	permutation, phases := composed.Permutation(), composed.InputPhases()
	permutation[0] = 99
	phases[0] = !phases[0]
	if slices.Equal(permutation, composed.Permutation()) || slices.Equal(phases, composed.InputPhases()) {
		t.Fatal("accessors alias transform")
	}
	for source := 0; source < 3; source++ {
		target, complement, err := composed.Variable(source)
		if err != nil {
			t.Fatal(err)
		}
		variable := truthVariablePattern(source) & 0xff
		actual, _ := composed.ApplyTruth(variable)
		want := truthVariablePattern(target) & 0xff
		if complement {
			want = (^want) & 0xff
		}
		if actual != want {
			t.Fatal(source, target, complement, actual, want)
		}
	}
}

func TestFindDttNPNTransform(t *testing.T) {
	source := uint64(0x96)
	transform, _ := NewDttNPNTransform([]int{2, 0, 1}, []bool{true, false, true})
	target, _ := transform.ApplyTruth(source)
	target = (^target) & 0xff
	found, output, err := FindDttNPNTransform(source, target, 3, DttTransformSearchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	actual, _ := found.ApplyTruth(source)
	if output {
		actual = (^actual) & 0xff
	}
	if actual != target {
		t.Fatal(found.Permutation(), found.InputPhases(), output, actual, target)
	}
	if found, output, err := FindDttNPNTransform(source, 0x81, 3, DttTransformSearchOptions{}); found.Variables() != 0 || output || !errors.Is(err, ErrUnsupported) {
		t.Fatal(found, output, err)
	}
	if found, output, err := FindDttNPNTransform(source, target, 3, DttTransformSearchOptions{MaxTransforms: 47}); found.Variables() != 0 || output || !errors.Is(err, ErrLimit) {
		t.Fatal(found, output, err)
	}
	if found, output, err := FindDttNPNTransform(source, target, 3, DttTransformSearchOptions{Stop: func() bool { return true }}); found.Variables() != 0 || output || !errors.Is(err, ErrCancelled) {
		t.Fatal(found, output, err)
	}
}

func TestProcessDttTypeAssignmentModel(t *testing.T) {
	mask := uint64(0xff)
	first, second := uint64(0xa6), uint64(0x3c)
	for kind := uint8(0); kind < 10; kind++ {
		for fanin := 1; fanin <= 3; fanin++ {
			processed, err := ProcessDttType(kind, fanin)
			if err != nil {
				t.Fatal(kind, fanin, err)
			}
			got, _ := ApplyDttType(processed, first, second, mask)
			wantFirst, wantSecond := first, second
			if fanin == 1 {
				wantFirst = (^wantFirst) & mask
			} else if fanin == 2 {
				wantSecond = (^wantSecond) & mask
			}
			want, _ := ApplyDttType(kind, wantFirst, wantSecond, mask)
			if fanin == 3 {
				want = (^want) & mask
			}
			if got != want {
				t.Fatal(kind, fanin, processed, got, want)
			}
		}
	}
	if _, err := ApplyDttType(10, first, second, mask); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzDttNPNTransform(f *testing.F) {
	f.Add(uint64(0x96), uint8(3), uint8(0x5), true)
	f.Add(uint64(0x6996), uint8(4), uint8(0xa), false)
	f.Fuzz(func(t *testing.T, truth uint64, width, phase uint8, output bool) {
		variables := 2 + int(width%4)
		permutation := make([]int, variables)
		phases := make([]bool, variables)
		for index := range permutation {
			permutation[index] = (index + 1) % variables
			phases[index] = phase&(1<<uint(index)) != 0
		}
		transform, err := NewDttNPNTransform(permutation, phases)
		if err != nil {
			t.Fatal(err)
		}
		target, _ := transform.ApplyTruth(truth)
		if output {
			target = (^target) & dauTruthMask(variables)
		}
		found, complement, err := FindDttNPNTransform(truth, target, variables, DttTransformSearchOptions{})
		if err != nil {
			t.Fatal(variables, truth, target, err)
		}
		actual, _ := found.ApplyTruth(truth)
		if complement {
			actual = (^actual) & dauTruthMask(variables)
		}
		if actual != target {
			t.Fatal(variables, truth, target, actual)
		}
	})
}
