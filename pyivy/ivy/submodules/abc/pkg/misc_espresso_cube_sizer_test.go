package abc

// Go-only unit and fuzz tests for src/misc/espresso/cubehack.c.

import (
	"errors"
	"slices"
	"sync"
	"testing"
)

func TestEspressoCubeSizerLifecycleAndDefaults(t *testing.T) {
	wantFlags := EspressoMinimizeFlags{RemoveEssential: true, ForceIrredundant: true, UnwrapOnset: true}
	if got := DefaultEspressoMinimizeFlags(); got != wantFlags {
		t.Fatal(got)
	}
	sizer := NewEspressoCubeSizer()
	if current, err := sizer.Current(); err != nil || current != nil {
		t.Fatal(current, err)
	}
	for _, inputs := range []int{0, 7, 100, 101, 7} {
		layout, err := sizer.Define(inputs)
		if err != nil || layout.BinaryVariableCount() != inputs || layout.VariableCount() != inputs || layout.BitCount() != 2*inputs || layout.OutputVariable() != -1 || !slices.Equal(layout.PartSizes(), slices.Repeat([]int{2}, inputs)) {
			t.Fatal(inputs, layout, err)
		}
		current, currentErr := sizer.Current()
		if currentErr != nil || current == layout || current.BitCount() != layout.BitCount() {
			t.Fatal(current, currentErr)
		}
		if inputs > 0 {
			layout.partSizes[0] = 9
			if current.PartSizes()[0] != 2 {
				t.Fatal("returned layout aliases cache")
			}
		}
	}
	if err := sizer.Undefine(); err != nil {
		t.Fatal(err)
	}
	if current, err := sizer.Current(); err != nil || current != nil {
		t.Fatal(current, err)
	}
}

func TestEspressoCubeSizerConcurrentOwnership(t *testing.T) {
	sizer := NewEspressoCubeSizer()
	var wait sync.WaitGroup
	errorsFound := make(chan error, 32)
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func(worker int) {
			defer wait.Done()
			for iteration := 0; iteration < 32; iteration++ {
				layout, err := sizer.Define((worker + iteration) % 17)
				if err != nil {
					errorsFound <- err
					return
				}
				if validateErr := layout.Validate(); validateErr != nil {
					errorsFound <- validateErr
					return
				}
				if iteration%11 == 0 {
					if err = sizer.Undefine(); err != nil {
						errorsFound <- err
						return
					}
				}
			}
		}(worker)
	}
	wait.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Fatal(err)
	}
}

func TestEspressoCubeSizerErrors(t *testing.T) {
	sizer := NewEspressoCubeSizer()
	if layout, err := sizer.CautiousDefine(-1); layout != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(layout, err)
	}
	if layout, err := sizer.Define(espressoSetMaximumBits/2 + 1); layout != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(layout, err)
	}
	var nilSizer *EspressoCubeSizer
	if _, err := nilSizer.Define(1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if err := nilSizer.Undefine(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoCubeSizer(f *testing.F) {
	f.Add([]byte{0, 7, 100, 101, 3})
	f.Add([]byte{255, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		sizer := NewEspressoCubeSizer()
		for index, value := range data {
			if value&0x80 != 0 {
				if err := sizer.Undefine(); err != nil {
					t.Fatal(err)
				}
				continue
			}
			inputs := int(value)
			var layout *EspressoCubeLayout
			var err error
			if index&1 == 0 {
				layout, err = sizer.Define(inputs)
			} else {
				layout, err = sizer.CautiousDefine(inputs)
			}
			if err != nil || layout.BitCount() != 2*inputs || layout.VariableCount() != inputs || layout.BinaryVariableCount() != inputs {
				t.Fatal(inputs, layout, err)
			}
			binary, binaryErr := layout.BinaryMask()
			full, fullErr := layout.FullCube()
			equal, equalErr := binary.Equal(full)
			if binaryErr != nil || fullErr != nil || equalErr != nil || !equal {
				t.Fatal(binaryErr, fullErr, equalErr, equal)
			}
		}
		current, err := sizer.Current()
		if err != nil {
			t.Fatal(err)
		}
		if current != nil && (current.VariableCount() != current.BinaryVariableCount() || current.BitCount() != 2*current.VariableCount()) {
			t.Fatal(current)
		}
	})
}
