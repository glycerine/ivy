package abc

// Go-only exact class-count and orbit checks for small NPN enumeration.

import (
	"errors"
	"slices"
	"testing"
)

func TestEnumerateDauNPNClassesKnownCounts(t *testing.T) {
	counts := []int{1, 2, 4, 14, 222}
	for variables, want := range counts {
		classes, err := EnumerateDauNPNClasses(variables, DauNPNClassOptions{})
		if err != nil || len(classes.Representatives()) != want || classes.Variables() != variables {
			t.Fatal(variables, want, classes, err)
		}
		representatives := classes.Representatives()
		if !slices.IsSorted(representatives) {
			t.Fatal("representatives are not sorted", variables)
		}
		for _, representative := range representatives {
			canonical, err := classes.Canonical(representative)
			if err != nil || canonical != representative {
				t.Fatal(variables, representative, canonical, err)
			}
		}
		if len(representatives) != 0 {
			representatives[0] = ^uint64(0)
			if again := classes.Representatives(); again[0] == ^uint64(0) {
				t.Fatal("representatives alias storage")
			}
		}
	}
}

func TestEnumerateDauNPNClassesLimits(t *testing.T) {
	if result, err := EnumerateDauNPNClasses(5, DauNPNClassOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	if result, err := EnumerateDauNPNClasses(4, DauNPNClassOptions{MaxClasses: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := EnumerateDauNPNClasses(4, DauNPNClassOptions{MaxTransforms: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := EnumerateDauNPNClasses(4, DauNPNClassOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	classes, err := EnumerateDauNPNClasses(3, DauNPNClassOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := classes.Canonical(1 << 8); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := (*DauNPNClasses)(nil).Canonical(0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzDauNPNCanonicalOrbit(f *testing.F) {
	f.Add(uint8(4), uint16(0x6996), uint8(3), uint8(5), true)
	f.Add(uint8(3), uint16(0x96), uint8(1), uint8(2), false)
	f.Fuzz(func(t *testing.T, width uint8, raw uint16, swapSeed, phaseSeed uint8, complement bool) {
		variables := int(width % 5)
		classes, err := EnumerateDauNPNClasses(variables, DauNPNClassOptions{})
		if err != nil {
			t.Fatal(err)
		}
		truthBits := 1 << uint(variables)
		mask := uint64(1)<<uint(truthBits) - 1
		truth := uint64(raw) & mask
		permutation := make([]int, variables)
		for i := range permutation {
			permutation[i] = i
		}
		for i := variables - 1; i > 0; i-- {
			other := int(swapSeed) % (i + 1)
			permutation[i], permutation[other] = permutation[other], permutation[i]
			swapSeed = swapSeed*33 + uint8(i)
		}
		phase := int(phaseSeed)
		transformed := transformDauNPNTruth(truth, variables, permutation, phase)
		if complement {
			transformed = (^transformed) & mask
		}
		first, err := classes.Canonical(truth)
		if err != nil {
			t.Fatal(err)
		}
		second, err := classes.Canonical(transformed)
		if err != nil || first != second {
			t.Fatal(variables, truth, transformed, permutation, phase, first, second, err)
		}
	})
}
