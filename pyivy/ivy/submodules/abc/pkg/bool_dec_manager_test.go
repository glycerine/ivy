package abc

// Stand-alone lifecycle and exhaustive NPN tests for decMan.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestDecManagerNPN4Exhaustive(t *testing.T) {
	manager, err := NewDecManager()
	if err != nil {
		t.Fatal(err)
	}
	classes, err := EnumerateDauNPNClasses(4, DauNPNClassOptions{})
	if err != nil {
		t.Fatal(err)
	}
	permutations := decPermutations(4)
	seenClasses := make(map[uint8]bool)
	for truth := 0; truth < 1<<16; truth++ {
		canonical, phase, permutation, class, queryErr := manager.NPN4(uint16(truth))
		if queryErr != nil || int(permutation) >= len(permutations) || phase > 31 {
			t.Fatal(truth, canonical, phase, permutation, class, queryErr)
		}
		independent, canonicalErr := classes.Canonical(uint64(truth))
		if canonicalErr != nil || uint64(canonical) != independent {
			t.Fatal("canonical", truth, canonical, independent, canonicalErr)
		}
		reconstructed := decTransformTruth4(canonical, int(phase&15), permutations[permutation])
		if phase&16 != 0 {
			reconstructed = ^reconstructed
		}
		if reconstructed != uint16(truth) {
			t.Fatal("transform", truth, canonical, phase, permutation, reconstructed)
		}
		seenClasses[class] = true
	}
	if len(seenClasses) != 222 || !seenClasses[0] || !seenClasses[221] {
		t.Fatal(len(seenClasses), seenClasses[0], seenClasses[221])
	}
	if _, phase, _, class, err := manager.NPN4(0xffff); err != nil || phase != 16 || class != 0 {
		t.Fatal(phase, class, err)
	}
	if canonical, _, _, class, err := manager.NPN4(1); err != nil || canonical != 1 || class != 1 {
		t.Fatal(canonical, class, err)
	}
	// Preserve Extra_Truth4VarNPN's observable sparse-map quirk: the output
	// complement of a canonical representative is not assigned a class number.
	if canonical, _, _, class, err := manager.NPN4(0xfffe); err != nil || canonical != 1 || class != 0 {
		t.Fatal(canonical, class, err)
	}
}

func TestDecManagerOwnershipAndClose(t *testing.T) {
	first, err := NewDecManager()
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewDecManager()
	if err != nil {
		t.Fatal(err)
	}
	canonical, phase, permutation, classes, err := first.NPN4Tables()
	if err != nil || len(canonical) != 1<<16 || len(phase) != 1<<16 || len(permutation) != 1<<16 || len(classes) != 1<<16 {
		t.Fatal(len(canonical), len(phase), len(permutation), len(classes), err)
	}
	canonical[1], phase[1], permutation[1], classes[1] = 0xffff, 0xff, 0xff, 0xff
	wantCanonical, wantPhase, wantPermutation, wantClass, err := second.NPN4(1)
	if err != nil || wantCanonical == canonical[1] || wantPhase == phase[1] || wantPermutation == permutation[1] || wantClass == classes[1] {
		t.Fatal(wantCanonical, wantPhase, wantPermutation, wantClass, err)
	}
	secondCanonical, secondPhase, secondPermutation, secondClasses, err := second.NPN4Tables()
	if err != nil {
		t.Fatal(err)
	}
	if slices.Equal(canonical, secondCanonical) || slices.Equal(phase, secondPhase) || slices.Equal(permutation, secondPermutation) || slices.Equal(classes, secondClasses) {
		t.Fatal("mutated snapshots unexpectedly equal")
	}
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, err = first.NPN4(0); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = first.Close(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	if err = (*DecManager)(nil).Close(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err = second.Close(); err != nil {
		t.Fatal(err)
	}
}

func FuzzDecManagerNPN4(f *testing.F) {
	for _, seed := range []uint16{0, 1, 0x6996, 0x8000, 0xffff} {
		f.Add(seed, uint8(seed), uint8(seed>>8), seed&1 != 0)
	}
	manager, err := NewDecManager()
	if err != nil {
		f.Fatal(err)
	}
	permutations := decPermutations(4)
	f.Fuzz(func(t *testing.T, truth uint16, phase, permutation uint8, complement bool) {
		permutation %= uint8(len(permutations))
		phase &= 15
		transformed := decTransformTruth4(truth, int(phase), permutations[permutation])
		if complement {
			transformed = ^transformed
		}
		first, firstPhase, firstPermutation, _, firstErr := manager.NPN4(truth)
		second, _, _, _, secondErr := manager.NPN4(transformed)
		if firstErr != nil || secondErr != nil || first != second {
			t.Fatal(truth, transformed, first, second, firstErr, secondErr)
		}
		rebuilt := decTransformTruth4(first, int(firstPhase&15), permutations[firstPermutation])
		if firstPhase&16 != 0 {
			rebuilt = ^rebuilt
		}
		if rebuilt != truth {
			t.Fatal(truth, first, firstPhase, firstPermutation, rebuilt)
		}
	})
}
