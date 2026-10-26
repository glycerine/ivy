package abc

// Go-only symbolic mapping models for src/misc/espresso/hack.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoSymbolicOutputSelected(pla *EspressoPLA, assignment, outputPart int) bool {
	for _, cube := range pla.On.sets {
		matches := true
		for variable := 0; variable < pla.Layout.binaryVariables; variable++ {
			value := assignment >> uint(variable) & 1
			present, _ := cube.Has(pla.Layout.firstParts[variable] + value)
			if !present {
				matches = false
				break
			}
		}
		if matches {
			present, _ := cube.Has(pla.Layout.firstParts[pla.Layout.OutputVariable()] + outputPart)
			if present {
				return true
			}
		}
	}
	return false
}

func TestEspressoSymbolicInputMappingLayoutLabelsAndTruth(t *testing.T) {
	truth := uint64(0b10110110)
	base := espressoPLAFromOutputTruths(t, 3, []uint64{truth})
	labels := []string{"a.bar", "a", "b.bar", "b", "c.bar", "c", "y"}
	pla, _ := NewEspressoPLA(base.Layout, base.On, base.DontCare, base.Off, labels)
	before := espressoFamilyMasks(t, pla.On)
	directive := EspressoSymbolicDirective{
		References: []EspressoSymbolicReference{{Variable: 0, Part: 1}, {Variable: 2, Part: 0}},
		Labels:     []string{"S0", "S1"},
	}
	mapped, err := pla.ApplySymbolicInputs([]EspressoSymbolicDirective{directive}, EspressoSymbolicOptions{})
	if err != nil || !slices.Equal(mapped.Layout.partSizes, []int{2, 4, 1}) || mapped.Layout.binaryVariables != 1 {
		t.Fatal(mapped.Layout.partSizes, mapped.Layout.binaryVariables, err)
	}
	if !slices.Equal(mapped.Labels, []string{"b.bar", "b", "S0", "S1", "X2", "X3", "y"}) {
		t.Fatal(mapped.Labels)
	}
	for assignment := 0; assignment < 8; assignment++ {
		b := assignment >> 1 & 1
		joint := (assignment&1)<<1 | (assignment >> 2 & 1)
		newAssignment := b + 2*joint
		got := espressoFamilyTruth(t, mapped.Layout, mapped.On)>>uint(newAssignment)&1 != 0
		want := truth>>uint(assignment)&1 != 0
		if got != want {
			t.Fatal(assignment, newAssignment, got, want)
		}
	}
	if !slices.Equal(espressoFamilyMasks(t, pla.On), before) {
		t.Fatal("symbolic input mapping mutated source")
	}
}

func TestEspressoSymbolicOutputMappingOneHotAndLabels(t *testing.T) {
	y0, y1 := uint64(0b1010), uint64(0b1100)
	base := espressoPLAFromOutputTruths(t, 2, []uint64{y0, y1})
	labels := []string{"a.bar", "a", "b.bar", "b", "y0", "y1"}
	pla, _ := NewEspressoPLA(base.Layout, base.On, base.DontCare, base.Off, labels)
	directive := EspressoSymbolicDirective{
		References: []EspressoSymbolicReference{{Variable: 2, Part: 0}, {Variable: 2, Part: 1}},
		Labels:     []string{"N0", "N1", "N2", "N3"},
	}
	mapped, err := pla.ApplySymbolicOutputs([]EspressoSymbolicDirective{directive}, EspressoSymbolicOptions{})
	if err != nil || !slices.Equal(mapped.Layout.partSizes, []int{2, 2, 4}) || len(mapped.Off.sets) != 0 || !slices.Equal(mapped.Labels, []string{"a.bar", "a", "b.bar", "b", "N0", "N1", "N2", "N3"}) {
		t.Fatal(mapped.Layout.partSizes, espressoFamilyMasks(t, mapped.Off), mapped.Labels, err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		value := int(y0>>uint(assignment)&1)<<1 | int(y1>>uint(assignment)&1)
		for part := 0; part < 4; part++ {
			got := espressoSymbolicOutputSelected(mapped, assignment, part)
			// find_inputs intersects the decoded region with F, so the all-zero
			// output point has no row when every original output is mapped.
			if got != (value != 0 && part == value) {
				t.Fatal(assignment, value, part, got, espressoFamilyMasks(t, mapped.On))
			}
		}
	}
}

func TestEspressoSymbolicOutputKeepsUnmappedAndRecomputesOn(t *testing.T) {
	base := espressoPLAFromOutputTruths(t, 1, []uint64{0b10, 0b01, 0b11})
	dcCube := espressoOutputCube(t, base.Layout.bitCount, 0, 2)
	dc := espressoOutputFamily(t, base.Layout.bitCount, dcCube)
	pla, _ := NewEspressoPLA(base.Layout, base.On, dc, base.Off, nil)
	directive := EspressoSymbolicDirective{References: []EspressoSymbolicReference{{Variable: 1, Part: 0}}, Labels: []string{"zero", "one"}}
	mapped, err := pla.ApplySymbolicOutputs([]EspressoSymbolicDirective{directive}, EspressoSymbolicOptions{})
	if err != nil || !slices.Equal(mapped.Layout.partSizes, []int{2, 4}) || len(mapped.DontCare.sets) == 0 {
		t.Fatal(mapped.Layout.partSizes, espressoFamilyMasks(t, mapped.DontCare), err)
	}
	// New one-hot parts precede the two untouched old outputs.
	if espressoSymbolicOutputSelected(mapped, 1, 1) != true || espressoSymbolicOutputSelected(mapped, 0, 0) != true {
		t.Fatal(espressoFamilyMasks(t, mapped.On))
	}
}

func TestEspressoSymbolicErrorsLimitsAndCancellation(t *testing.T) {
	pla := espressoPLAFromOutputTruths(t, 2, []uint64{0b1010, 0b1100})
	badInput := []EspressoSymbolicDirective{{References: []EspressoSymbolicReference{{Variable: 2}}}}
	if result, err := pla.ApplySymbolicInputs(badInput, EspressoSymbolicOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	badOutput := []EspressoSymbolicDirective{{References: []EspressoSymbolicReference{{Part: 2}}}}
	if result, err := pla.ApplySymbolicOutputs(badOutput, EspressoSymbolicOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	deep := []EspressoSymbolicDirective{{References: []EspressoSymbolicReference{{Variable: 0}, {Variable: 1}}}}
	if result, err := pla.ApplySymbolicInputs(deep, EspressoSymbolicOptions{MaxDepth: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := pla.ApplySymbolicInputs(deep, EspressoSymbolicOptions{MaxOperations: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := pla.ApplySymbolicOutputs([]EspressoSymbolicDirective{{References: []EspressoSymbolicReference{{Part: 0}}}}, EspressoSymbolicOptions{MaxCubes: 1}); result != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := pla.ApplySymbolicInputs(deep, EspressoSymbolicOptions{Stop: func() bool { return true }}); result != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	invalidCube, _ := NewEspressoSet(pla.Layout.bitCount)
	_ = invalidCube.Set(2, true)
	_ = invalidCube.Set(4, true)
	invalidFamily := espressoOutputFamily(t, pla.Layout.bitCount, invalidCube)
	invalidPLA, _ := NewEspressoPLA(pla.Layout, invalidFamily, pla.DontCare, pla.Off, nil)
	if result, err := invalidPLA.ApplySymbolicInputs([]EspressoSymbolicDirective{{References: []EspressoSymbolicReference{{Variable: 0}}}}, EspressoSymbolicOptions{}); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
}

func FuzzEspressoSymbolicInputs(f *testing.F) {
	f.Add(byte(0x96), byte(0x69), byte(0))
	f.Add(byte(0xff), byte(0x00), byte(1))
	f.Fuzz(func(t *testing.T, onByte, offByte, orderByte byte) {
		onTruth := uint64(onByte)
		offTruth := uint64(offByte)
		base := espressoPLAFromOutputTruths(t, 3, []uint64{onTruth})
		// Replace OFF independently so both cover transformations are checked.
		offPLA := espressoPLAFromOutputTruths(t, 3, []uint64{offTruth})
		pla, _ := NewEspressoPLA(base.Layout, base.On, base.DontCare, offPLA.On, nil)
		first, second := 0, 2
		if orderByte&1 != 0 {
			first, second = second, first
		}
		directive := EspressoSymbolicDirective{References: []EspressoSymbolicReference{{Variable: first}, {Variable: second}}}
		mapped, err := pla.ApplySymbolicInputs([]EspressoSymbolicDirective{directive}, EspressoSymbolicOptions{MaxCubes: 1 << 12, MaxOperations: 1 << 16})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 8; assignment++ {
			unpaired := assignment >> 1 & 1
			firstValue := assignment >> uint(first) & 1
			secondValue := assignment >> uint(second) & 1
			newAssignment := unpaired + 2*(firstValue*2+secondValue)
			for _, item := range []struct {
				family *EspressoSetFamily
				truth  uint64
			}{{mapped.On, onTruth}, {mapped.Off, offTruth}} {
				got := espressoFamilyTruth(t, mapped.Layout, item.family)>>uint(newAssignment)&1 != 0
				want := item.truth>>uint(assignment)&1 != 0
				if got != want {
					t.Fatal(assignment, newAssignment, first, second, got, want)
				}
			}
		}
	})
}

func FuzzEspressoSymbolicOutputs(f *testing.F) {
	f.Add(byte(0xa), byte(0xc), byte(0))
	f.Add(byte(0xf), byte(0x0), byte(1))
	f.Fuzz(func(t *testing.T, firstTruthByte, secondTruthByte, reverseByte byte) {
		firstTruth := uint64(firstTruthByte & 0xf)
		secondTruth := uint64(secondTruthByte & 0xf)
		pla := espressoPLAFromOutputTruths(t, 2, []uint64{firstTruth, secondTruth})
		refs := []EspressoSymbolicReference{{Part: 0}, {Part: 1}}
		if reverseByte&1 != 0 {
			refs[0], refs[1] = refs[1], refs[0]
		}
		mapped, err := pla.ApplySymbolicOutputs([]EspressoSymbolicDirective{{References: refs}}, EspressoSymbolicOptions{MaxCubes: 1 << 14, MaxOperations: 1 << 18})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 4; assignment++ {
			values := []int{int(firstTruth >> uint(assignment) & 1), int(secondTruth >> uint(assignment) & 1)}
			value := values[refs[0].Part]*2 + values[refs[1].Part]
			for part := 0; part < 4; part++ {
				if got := espressoSymbolicOutputSelected(mapped, assignment, part); got != (value != 0 && part == value) {
					t.Fatal(firstTruth, secondTruth, assignment, refs, value, part, got)
				}
			}
		}
	})
}
