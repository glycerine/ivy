package abc

// Go-only assignment models for src/misc/espresso/equiv.c.

import (
	"errors"
	"slices"
	"testing"
)

func espressoOutputPLA(t testing.TB, variables, outputs int, truths []uint64, labels []string) *EspressoPLA {
	t.Helper()
	parts := append(slices.Repeat([]int{2}, variables), outputs)
	layout, err := NewEspressoCubeLayout(parts, variables)
	if err != nil {
		t.Fatal(err)
	}
	off, _ := NewEspressoSetFamily(1<<uint(variables), layout.BitCount())
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		cube, _ := NewEspressoSet(layout.BitCount())
		for variable := 0; variable < variables; variable++ {
			_ = cube.Set(2*variable+((assignment>>uint(variable))&1), true)
		}
		selected := false
		for output := 0; output < outputs; output++ {
			if truths[output]>>uint(assignment)&1 != 0 {
				_ = cube.Set(2*variables+output, true)
				selected = true
			}
		}
		if selected {
			if err := off.Add(cube); err != nil {
				t.Fatal(err)
			}
		}
	}
	empty, _ := NewEspressoSetFamily(0, layout.BitCount())
	pla, err := NewEspressoPLA(layout, empty, empty, off, labels)
	if err != nil {
		t.Fatal(err)
	}
	return pla
}

func TestEspressoEquivalentCoversSemanticAndOwned(t *testing.T) {
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	first := espressoMapFamily(t, 4, 0b01|0b11<<2, 0b10|0b01<<2)
	second := espressoMapFamily(t, 4, 0b01|0b10<<2, 0b11|0b01<<2)
	equal, err := layout.EquivalentCovers(first, second, EspressoVerifyOptions{})
	if err != nil || !equal {
		t.Fatal(equal, err)
	}
	other := espressoMapFamily(t, 4, 0b01|0b01<<2)
	equal, err = layout.EquivalentCovers(first, other, EspressoVerifyOptions{})
	if err != nil || equal {
		t.Fatal(equal, err)
	}
	if !slices.Equal(espressoFamilyMasks(t, first), []uint64{0b01 | 0b11<<2, 0b10 | 0b01<<2}) {
		t.Fatal("source cover mutated")
	}
}

func TestEspressoFindEquivalentOutputsOrderLabelsAndPolarity(t *testing.T) {
	labels := []string{"a.bar", "a", "b.bar", "b", "same0", "same1", "inverse"}
	pla := espressoOutputPLA(t, 2, 3, []uint64{0b0011, 0b0011, 0b1100}, labels)
	result, err := pla.FindEquivalentOutputs(EspressoVerifyOptions{})
	want := []EspressoOutputEquivalence{
		{First: 0, Second: 1, FirstLabel: "same0", SecondLabel: "same1"},
		{First: 0, Second: 2, FirstLabel: "same0", SecondLabel: "inverse", Inverted: true},
		{First: 1, Second: 2, FirstLabel: "same1", SecondLabel: "inverse", Inverted: true},
	}
	if err != nil || !slices.Equal(result, want) {
		t.Fatal(result, err)
	}
	result[0].FirstLabel = "changed"
	if pla.Labels[4] != "same0" {
		t.Fatal("result aliases PLA labels")
	}
	defaultLabels := espressoOutputPLA(t, 1, 2, []uint64{1, 1}, nil)
	defaultResult, err := defaultLabels.FindEquivalentOutputs(EspressoVerifyOptions{})
	if err != nil || len(defaultResult) != 1 || defaultResult[0].FirstLabel != "v1.0" || defaultResult[0].SecondLabel != "v1.1" {
		t.Fatal(defaultResult, err)
	}
}

func TestEspressoFindEquivalentOutputsNoneErrorsLimitsAndCancellation(t *testing.T) {
	pla := espressoOutputPLA(t, 2, 3, []uint64{0b0001, 0b0011, 0b0111}, nil)
	result, err := pla.FindEquivalentOutputs(EspressoVerifyOptions{})
	if err != nil || len(result) != 0 {
		t.Fatal(result, err)
	}
	if _, err = pla.FindEquivalentOutputs(EspressoVerifyOptions{MaxOperations: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = pla.FindEquivalentOutputs(EspressoVerifyOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	layout, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	empty := espressoMapFamily(t, 4)
	withoutOutput, _ := NewEspressoPLA(layout, empty, empty, empty, nil)
	if _, err = withoutOutput.FindEquivalentOutputs(EspressoVerifyOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	wrong, _ := NewEspressoCubeLayout([]int{2}, 1)
	if _, err = wrong.EquivalentCovers(empty, empty, EspressoVerifyOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoOutputEquivalence(f *testing.F) {
	f.Add([]byte{2, 3, 0x3c, 0xc3, 0xaa})
	f.Add([]byte{1, 4, 0, 3, 1, 2})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		variables := int(data[0]%3) + 1
		outputs := int(data[1]%4) + 2
		assignments := 1 << uint(variables)
		universe := uint64(1)<<uint(assignments) - 1
		truths := make([]uint64, outputs)
		for output := range truths {
			for assignment := 0; assignment < assignments; assignment++ {
				if data[(2+output*assignments+assignment)%len(data)]>>uint(assignment&7)&1 != 0 {
					truths[output] |= uint64(1) << uint(assignment)
				}
			}
		}
		pla := espressoOutputPLA(t, variables, outputs, truths, nil)
		got, err := pla.FindEquivalentOutputs(EspressoVerifyOptions{})
		if err != nil {
			t.Fatal(err)
		}
		want := make([]EspressoOutputEquivalence, 0)
		for first := 0; first < outputs-1; first++ {
			for second := first + 1; second < outputs; second++ {
				equal := truths[first] == truths[second]
				inverted := truths[first] == universe&^truths[second]
				if equal || inverted {
					want = append(want, EspressoOutputEquivalence{First: first, Second: second, FirstLabel: "v" + string(rune('0'+variables)) + "." + string(rune('0'+first)), SecondLabel: "v" + string(rune('0'+variables)) + "." + string(rune('0'+second)), Inverted: !equal && inverted})
				}
			}
		}
		if len(got) != len(want) {
			t.Fatal(truths, got, want)
		}
		for index := range got {
			if got[index].First != want[index].First || got[index].Second != want[index].Second || got[index].Inverted != want[index].Inverted {
				t.Fatal(truths, got, want)
			}
		}
	})
}
