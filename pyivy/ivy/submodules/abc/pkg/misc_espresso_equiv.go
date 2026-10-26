package abc

// Output-function equivalence analysis from src/misc/espresso/equiv.c.

import "fmt"

type EspressoOutputEquivalence struct {
	First, Second           int
	FirstLabel, SecondLabel string
	Inverted                bool
}

func (layout *EspressoCubeLayout) equivalentCovers(first, second *EspressoSetFamily, work *espressoSharpWork) (bool, error) {
	firstOutside, err := layout.coverDifference(first, second, work)
	if err != nil {
		return false, err
	}
	if len(firstOutside.sets) != 0 {
		return false, nil
	}
	secondOutside, err := layout.coverDifference(second, first, work)
	if err != nil {
		return false, err
	}
	return len(secondOutside.sets) == 0, nil
}

// EquivalentCovers reports semantic equality of two covers. Cube order,
// duplication and representation do not affect the result.
func (layout *EspressoCubeLayout) EquivalentCovers(first, second *EspressoSetFamily, options EspressoVerifyOptions) (bool, error) {
	work, err := newEspressoVerifyWork(options)
	if err != nil {
		return false, err
	}
	if err = layout.check(); err != nil {
		return false, err
	}
	return layout.equivalentCovers(first, second, work)
}

func (layout *EspressoCubeLayout) complementaryCovers(first, second *EspressoSetFamily, work *espressoSharpWork) (bool, error) {
	intersection, err := layout.intersectFamilies(first, second, work)
	if err != nil {
		return false, err
	}
	if len(intersection.sets) != 0 {
		return false, nil
	}
	union, err := JoinEspressoSetFamilies(first, second)
	if err != nil {
		return false, err
	}
	full, _ := layout.FullCube()
	missing, err := layout.sharpCubeByCover(full, union, work)
	if err != nil {
		return false, err
	}
	return len(missing.sets) == 0, nil
}

// cofactorOutput selects rows containing one output part and makes the output
// variable unrestricted, matching cof_output's observable cover semantics.
func (pla *EspressoPLA) cofactorOutput(part int, work *espressoSharpWork) (*EspressoSetFamily, error) {
	output := pla.Layout.OutputVariable()
	if output < 0 {
		return nil, ErrDimension
	}
	first := pla.Layout.firstParts[output]
	if part < 0 || part >= pla.Layout.partSizes[output] {
		return nil, ErrDimension
	}
	result, _ := NewEspressoSetFamily(len(pla.Off.sets), pla.Layout.bitCount)
	for _, cube := range pla.Off.sets {
		if err := work.step(); err != nil {
			return nil, err
		}
		selected, _ := cube.Has(first + part)
		if !selected {
			continue
		}
		if err := work.addCube(); err != nil {
			return nil, err
		}
		clone, _ := cube.Clone()
		for offset := 0; offset < pla.Layout.partSizes[output]; offset++ {
			if err := work.step(); err != nil {
				return nil, err
			}
			_ = clone.Set(first+offset, true)
		}
		clone.Flags &^= EspressoSetPrime
		result.sets = append(result.sets, clone)
		if clone.Flags&EspressoSetActive != 0 {
			result.activeCount++
		}
	}
	return result, nil
}

func (pla *EspressoPLA) outputLabel(output, part int) string {
	index := pla.Layout.firstParts[output] + part
	if len(pla.Labels) != 0 && pla.Labels[index] != "" {
		return pla.Labels[index]
	}
	return fmt.Sprintf("v%d.%d", output, part)
}

// FindEquivalentOutputs returns output pairs that implement the same function
// or each other's complement. Results retain the source's lexicographic pair
// order and prefer direct equivalence when both predicates coincide.
func (pla *EspressoPLA) FindEquivalentOutputs(options EspressoVerifyOptions) ([]EspressoOutputEquivalence, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	output := pla.Layout.OutputVariable()
	if output < 0 {
		return nil, ErrDimension
	}
	work, err := newEspressoVerifyWork(options)
	if err != nil {
		return nil, err
	}
	count := pla.Layout.partSizes[output]
	off := make([]*EspressoSetFamily, count)
	for part := range off {
		off[part], err = pla.cofactorOutput(part, work)
		if err != nil {
			return nil, err
		}
	}
	result := make([]EspressoOutputEquivalence, 0)
	for first := 0; first < count-1; first++ {
		for second := first + 1; second < count; second++ {
			equal, err := pla.Layout.equivalentCovers(off[first], off[second], work)
			if err != nil {
				return nil, err
			}
			inverted := false
			if !equal {
				inverted, err = pla.Layout.complementaryCovers(off[first], off[second], work)
				if err != nil {
					return nil, err
				}
			}
			if equal || inverted {
				result = append(result, EspressoOutputEquivalence{
					First: first, Second: second, FirstLabel: pla.outputLabel(output, first),
					SecondLabel: pla.outputLabel(output, second), Inverted: inverted,
				})
			}
		}
	}
	return result, nil
}
