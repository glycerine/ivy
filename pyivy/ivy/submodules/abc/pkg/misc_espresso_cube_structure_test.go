package abc

// Go-only model and fuzz tests for src/misc/espresso/cubestr.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestEspressoCubeStructureDerivedDataAndIndependentLayouts(t *testing.T) {
	layout, err := NewEspressoCubeLayout([]int{2, 2, 3, 35}, 2)
	if err != nil || layout.BinaryVariableCount() != 2 || layout.MultipleValuedVariableCount() != 2 || layout.OutputVariable() != 3 || layout.BitCount() != 42 {
		t.Fatal(layout, err)
	}
	wantRanges := [][2]int{{0, 2}, {2, 4}, {4, 7}, {7, 42}}
	for variable, want := range wantRanges {
		first, end, rangeErr := layout.VariableRange(variable)
		firstWord, lastWord, wordErr := layout.VariableWordRange(variable)
		if rangeErr != nil || wordErr != nil || first != want[0] || end != want[1] || firstWord != want[0]>>5 || lastWord != (want[1]-1)>>5 {
			t.Fatal(variable, first, end, firstWord, lastWord, rangeErr, wordErr)
		}
		mask, maskErr := layout.VariableMask(variable)
		elements, elementsErr := mask.Elements()
		wantElements := make([]int, end-first)
		for index := range wantElements {
			wantElements[index] = first + index
		}
		if maskErr != nil || elementsErr != nil || !slices.Equal(elements, wantElements) {
			t.Fatal(variable, elements, wantElements, maskErr, elementsErr)
		}
	}
	binary, _ := layout.BinaryMask()
	multiple, _ := layout.MultipleValuedMask()
	full, _ := layout.FullCube()
	union, _ := binary.Union(multiple)
	disjoint, _ := binary.Disjoint(multiple)
	equal, _ := union.Equal(full)
	empty, _ := layout.EmptyCube()
	emptyValue, _ := empty.Empty()
	sparse, _ := layout.DefaultSparseVariables()
	word, inputMask, err := layout.BinaryInputWordMask()
	if err != nil || !disjoint || !equal || !emptyValue || !slices.Equal(sparse, []bool{false, false, true, true}) || word != 0 || inputMask != 0b0101 {
		t.Fatal(disjoint, equal, emptyValue, sparse, word, inputMask, err)
	}

	other, _ := NewEspressoCubeLayout([]int{2, 5}, 1)
	clone, err := layout.Clone()
	if err != nil || clone == layout || !slices.Equal(clone.PartSizes(), layout.PartSizes()) || other.BitCount() != 7 || layout.BitCount() != 42 {
		t.Fatal(clone, other, err)
	}
	clone.partSizes[2] = 4
	if layout.PartSizes()[2] != 3 {
		t.Fatal("clone aliases source")
	}
}

func TestEspressoCubeStructureSourceSymbolsAndErrors(t *testing.T) {
	layout, err := NewEspressoCubeLayoutFromSource([]int{99, -7, -3}, 1)
	if err != nil || !slices.Equal(layout.PartSizes(), []int{2, 7, 3}) || layout.BitCount() != 12 {
		t.Fatal(layout, err)
	}
	symbolic, err := layout.SymbolicVariables()
	if err != nil || !slices.Equal(symbolic, []bool{false, true, true}) {
		t.Fatal(symbolic, err)
	}
	symbolic[1] = false
	again, _ := layout.SymbolicVariables()
	if !again[1] {
		t.Fatal("symbolic snapshot aliases layout")
	}
	noBinary, _ := NewEspressoCubeLayout([]int{3}, 0)
	word, mask, err := noBinary.BinaryInputWordMask()
	if err != nil || word != -1 || mask != 0 || noBinary.OutputVariable() != 0 {
		t.Fatal(word, mask, noBinary.OutputVariable(), err)
	}
	pureBinary, _ := NewEspressoCubeLayout([]int{2, 2}, 2)
	if pureBinary.OutputVariable() != -1 {
		t.Fatal(pureBinary.OutputVariable())
	}
	if _, err = NewEspressoCubeLayoutFromSource([]int{2, 0}, 1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, _, err = layout.VariableRange(3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzEspressoCubeStructure(f *testing.F) {
	f.Add([]byte{2, 3, 4, 5, 6})
	f.Add([]byte{0, 1, 0xff})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) < 2 {
			return
		}
		variables := int(data[0]%12) + 1
		binary := int(data[1] % byte(variables+1))
		sourceSizes := make([]int, variables)
		for variable := range sourceSizes {
			value := int(data[(2+variable)%len(data)]%8) + 1
			if variable >= binary && data[(3+variable)%len(data)]&1 != 0 {
				value = -value
			}
			sourceSizes[variable] = value
		}
		layout, err := NewEspressoCubeLayoutFromSource(sourceSizes, binary)
		if err != nil {
			t.Fatal(err)
		}
		position := 0
		combined, _ := NewEspressoSet(layout.BitCount())
		for variable := 0; variable < variables; variable++ {
			first, end, rangeErr := layout.VariableRange(variable)
			mask, maskErr := layout.VariableMask(variable)
			if rangeErr != nil || maskErr != nil || first != position || end-first != layout.partSizes[variable] {
				t.Fatal(variable, first, end, position, rangeErr, maskErr)
			}
			overlap, _ := combined.Disjoint(mask)
			if !overlap {
				t.Fatal("variable masks overlap", variable)
			}
			combined, _ = combined.Union(mask)
			position = end
		}
		full, _ := layout.FullCube()
		equal, _ := combined.Equal(full)
		binaryMask, _ := layout.BinaryMask()
		mvMask, _ := layout.MultipleValuedMask()
		disjoint, _ := binaryMask.Disjoint(mvMask)
		joined, _ := binaryMask.Union(mvMask)
		joinedFull, _ := joined.Equal(full)
		if position != layout.BitCount() || !equal || !disjoint || !joinedFull {
			t.Fatal(position, layout.BitCount(), equal, disjoint, joinedFull)
		}
	})
}
