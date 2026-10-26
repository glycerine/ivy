package abc

// Go-only assignment models for src/misc/espresso/verify.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestEspressoVerifyCoversDifferences(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	original := espressoMapFamily(t, 4, 0b01|0b11<<2)
	current := espressoMapFamily(t, 4, 0b01|0b01<<2)
	dontCare := espressoMapFamily(t, 4, 0b01|0b10<<2)
	verification, err := layout.VerifyCovers(current, original, dontCare, EspressoVerifyOptions{})
	if err != nil || !verification.Equivalent || len(espressoFamilyMasks(t, verification.CurrentOutside)) != 0 || len(espressoFamilyMasks(t, verification.OriginalOutside)) != 0 {
		t.Fatal(verification, err)
	}
	extra := espressoMapFamily(t, 4, 0b10|0b10<<2)
	verification, err = layout.VerifyCovers(extra, original, dontCare, EspressoVerifyOptions{})
	if err != nil || verification.Equivalent || espressoFamilyTruth(t, layout, verification.CurrentOutside) == 0 {
		t.Fatal(verification, err)
	}
}

func TestEspressoPLAPermutationVerificationAndOwnership(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	empty := espressoMapFamily(t, 4)
	sourceOn := espressoMapFamily(t, 4, 0b01|0b10<<2)
	source, err := NewEspressoPLA(layout, sourceOn, empty, empty, []string{"a0", "a1", "b0", "b1"})
	if err != nil {
		t.Fatal(err)
	}
	referenceOn := espressoMapFamily(t, 4, 0b10|0b01<<2)
	reference, _ := NewEspressoPLA(layout, referenceOn, empty, empty, []string{"b0", "b1", "a0", "a1"})
	permuted, err := source.PermuteTo(reference)
	if err != nil || !slices.Equal(espressoFamilyMasks(t, permuted.On), []uint64{0b10 | 0b01<<2}) || !slices.Equal(permuted.Labels, reference.Labels) {
		t.Fatal(permuted, err)
	}
	verification, err := source.Verify(reference, EspressoVerifyOptions{})
	if err != nil || !verification.Equivalent {
		t.Fatal(verification, err)
	}
	permuted.Labels[0] = "changed"
	if source.Labels[0] != "a0" || reference.Labels[0] != "b0" || !slices.Equal(espressoFamilyMasks(t, source.On), []uint64{0b01 | 0b10<<2}) {
		t.Fatal("PLA ownership violated")
	}
}

func TestEspressoPLAConsistencyStructuredWitnesses(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	on := espressoMapFamily(t, 4, 0b01|0b11<<2)
	dontCare := espressoMapFamily(t, 4, 0b10|0b01<<2)
	off := espressoMapFamily(t, 4, 0b10|0b10<<2)
	pla, _ := NewEspressoPLA(layout, on, dontCare, off, nil)
	consistent, err := pla.CheckConsistency(EspressoVerifyOptions{})
	if err != nil || !consistent.Consistent {
		t.Fatal(consistent, err)
	}
	overlapOff := espressoMapFamily(t, 4, 0b01|0b01<<2)
	bad, _ := NewEspressoPLA(layout, on, dontCare, overlapOff, nil)
	inconsistent, err := bad.CheckConsistency(EspressoVerifyOptions{})
	if err != nil || inconsistent.Consistent || espressoFamilyTruth(t, layout, inconsistent.OnOff) == 0 || espressoFamilyTruth(t, layout, inconsistent.Unspecified) == 0 {
		t.Fatal(inconsistent, err)
	}
}

func TestEspressoVerifyErrorsLimitsAndCancellation(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	cover := espressoMapFamily(t, 4, 0b1111, 0b0101)
	empty := espressoMapFamily(t, 4)
	if _, err := layout.VerifyCovers(cover, cover, empty, EspressoVerifyOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := layout.VerifyCovers(cover, cover, empty, EspressoVerifyOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if _, err := NewEspressoPLA(layout, cover, empty, empty, []string{"short"}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	first, _ := NewEspressoPLA(layout, cover, empty, empty, []string{"a", "a", "b", "c"})
	second, _ := NewEspressoPLA(layout, cover, empty, empty, []string{"a", "b", "c", "d"})
	if _, err := first.PermuteTo(second); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	withoutLabels, _ := NewEspressoPLA(layout, cover, empty, empty, nil)
	if _, err := withoutLabels.Verify(withoutLabels, EspressoVerifyOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoVerifyAndConsistency(f *testing.F) {
	f.Add([]byte{2, 2, 1, 1, 0x55, 0xaa})
	f.Add([]byte{4, 5, 3, 2, 0xff, 0x01})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 5 {
			return
		}
		variables := int(data[0]%4) + 1
		layout, err := NewEspressoCubeLayout(slices.Repeat([]int{2}, variables), variables)
		if err != nil {
			t.Fatal(err)
		}
		makeCover := func(count, offset int) ([]uint64, *EspressoSetFamily) {
			masks := make([]uint64, count)
			for cube := range masks {
				for variable := 0; variable < variables; variable++ {
					choice := data[(offset+cube*variables+variable)%len(data)] % 3
					masks[cube] |= uint64(choice+1) << uint(2*variable)
				}
			}
			return masks, espressoMapFamily(t, layout.BitCount(), masks...)
		}
		currentMasks, current := makeCover(int(data[1]%6), 4)
		originalMasks, original := makeCover(int(data[2]%6), 5+len(currentMasks)*variables)
		dontCareMasks, dontCare := makeCover(int(data[3]%4), 6+(len(currentMasks)+len(originalMasks))*variables)
		verification, err := layout.VerifyCovers(current, original, dontCare, EspressoVerifyOptions{})
		if err != nil {
			t.Fatal(err)
		}
		currentTruth := espressoFamilyTruth(t, layout, current)
		originalTruth := espressoFamilyTruth(t, layout, original)
		dontCareTruth := espressoFamilyTruth(t, layout, dontCare)
		wantEquivalent := currentTruth&^(originalTruth|dontCareTruth) == 0 && originalTruth&^(currentTruth|dontCareTruth) == 0
		if verification.Equivalent != wantEquivalent || espressoFamilyTruth(t, layout, verification.CurrentOutside) != currentTruth&^(originalTruth|dontCareTruth) || espressoFamilyTruth(t, layout, verification.OriginalOutside) != originalTruth&^(currentTruth|dontCareTruth) {
			t.Fatal(verification.Equivalent, wantEquivalent, currentTruth, originalTruth, dontCareTruth)
		}
		offMasks, off := makeCover(int(data[4]%5), 7+(len(currentMasks)+len(originalMasks)+len(dontCareMasks))*variables)
		_ = offMasks
		pla, err := NewEspressoPLA(layout, current, dontCare, off, nil)
		if err != nil {
			t.Fatal(err)
		}
		consistency, err := pla.CheckConsistency(EspressoVerifyOptions{})
		if err != nil {
			t.Fatal(err)
		}
		offTruth := espressoFamilyTruth(t, layout, off)
		totalAssignments := 1 << uint(variables)
		universe := uint64(1)<<uint(totalAssignments) - 1
		wantConsistent := currentTruth&dontCareTruth == 0 && currentTruth&offTruth == 0 && dontCareTruth&offTruth == 0 && currentTruth|dontCareTruth|offTruth == universe
		if consistency.Consistent != wantConsistent || espressoFamilyTruth(t, layout, consistency.OnDontCare) != currentTruth&dontCareTruth || espressoFamilyTruth(t, layout, consistency.OnOff) != currentTruth&offTruth || espressoFamilyTruth(t, layout, consistency.DontCareOff) != dontCareTruth&offTruth || espressoFamilyTruth(t, layout, consistency.Unspecified) != universe&^(currentTruth|dontCareTruth|offTruth) {
			t.Fatal(consistency.Consistent, wantConsistent, currentTruth, dontCareTruth, offTruth)
		}
	})
}
