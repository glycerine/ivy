package abc

// Espresso-II cube reduction from src/misc/espresso/reduce.c.

import "slices"

type EspressoReduceOrder uint8

const (
	EspressoReduceDistanceOrder EspressoReduceOrder = iota
	EspressoReduceColumnOrder
)

type EspressoReduceOptions struct {
	Order                   EspressoReduceOrder
	MaxCubes, MaxOperations int
	Stop                    func() bool
}

func newEspressoReduceWork(options EspressoReduceOptions) (*espressoSharpWork, error) {
	if options.Order > EspressoReduceColumnOrder {
		return nil, ErrDimension
	}
	return newEspressoSharpWork(EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
}

func (layout *EspressoCubeLayout) positionalSupercube(cover *EspressoSetFamily, work *espressoSharpWork) (*EspressoSet, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	if cover.bits != layout.bitCount {
		return nil, ErrDimension
	}
	result, _ := NewEspressoSet(layout.bitCount)
	for _, cube := range cover.sets {
		for index, word := range cube.words {
			if err := work.step(); err != nil {
				return nil, err
			}
			result.words[index] |= word
		}
	}
	return result, nil
}

func (layout *EspressoCubeLayout) smallestCubeContainingComplement(cover *EspressoSetFamily, work *espressoSharpWork) (*EspressoSet, error) {
	complement, err := layout.complementCover(cover, work)
	if err != nil {
		return nil, err
	}
	return layout.positionalSupercube(complement, work)
}

// SmallestCubeContainingComplement implements sccc while consolidating its
// recursive split, merge and special-case helpers into exact complementation.
func (layout *EspressoCubeLayout) SmallestCubeContainingComplement(cover *EspressoSetFamily, options EspressoReduceOptions) (*EspressoSet, error) {
	work, err := newEspressoReduceWork(options)
	if err != nil {
		return nil, err
	}
	if err = layout.check(); err != nil {
		return nil, err
	}
	return layout.smallestCubeContainingComplement(cover, work)
}

func (layout *EspressoCubeLayout) reduceCubeAgainst(target *EspressoSet, otherOn, dontCare *EspressoSetFamily, work *espressoSharpWork) (*EspressoSet, error) {
	if err := layout.checkSet(target); err != nil {
		return nil, err
	}
	allowed, err := JoinEspressoSetFamilies(otherOn, dontCare)
	if err != nil {
		return nil, err
	}
	if len(allowed.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	essential, err := layout.sharpCubeByCover(target, allowed, work)
	if err != nil {
		return nil, err
	}
	return layout.positionalSupercube(essential, work)
}

// ReduceCube returns the maximal source reduction of one ON cube against all
// other ON cubes and the DC cover. Inputs and the result are independently
// owned.
func (layout *EspressoCubeLayout) ReduceCube(on, dontCare *EspressoSetFamily, index int, options EspressoReduceOptions) (*EspressoSet, error) {
	work, err := newEspressoReduceWork(options)
	if err != nil {
		return nil, err
	}
	if err = layout.check(); err != nil {
		return nil, err
	}
	if err = on.check(); err != nil {
		return nil, err
	}
	if err = dontCare.check(); err != nil {
		return nil, err
	}
	if on.bits != layout.bitCount || dontCare.bits != layout.bitCount || index < 0 || index >= len(on.sets) {
		return nil, ErrDimension
	}
	if len(on.sets) > work.options.MaxCubes || len(dontCare.sets) > work.options.MaxCubes-len(on.sets)+1 {
		return nil, ErrLimit
	}
	other, _ := NewEspressoSetFamily(len(on.sets)-1, layout.bitCount)
	for current, cube := range on.sets {
		if current != index {
			_ = other.Add(cube)
		}
	}
	return layout.reduceCubeAgainst(on.sets[index], other, dontCare, work)
}

func (layout *EspressoCubeLayout) reduceOrder(on *EspressoSetFamily, order EspressoReduceOrder, work *espressoSharpWork) ([]*EspressoSet, error) {
	ordered := make([]*EspressoSet, len(on.sets))
	for index, cube := range on.sets {
		ordered[index], _ = cube.Clone()
	}
	if len(ordered) < 2 {
		return ordered, nil
	}
	scores := make(map[*EspressoSet]int, len(ordered))
	switch order {
	case EspressoReduceDistanceOrder:
		largest := ordered[0]
		largestSize, _ := largest.Cardinality()
		for _, cube := range ordered[1:] {
			size, _ := cube.Cardinality()
			if size > largestSize {
				largest, largestSize = cube, size
			}
		}
		for _, cube := range ordered {
			if err := work.step(); err != nil {
				return nil, err
			}
			distance, _ := layout.Distance(largest, cube)
			size, _ := cube.Cardinality()
			scores[cube] = (layout.VariableCount()-distance)<<7 + min(size, 127)
		}
	case EspressoReduceColumnOrder:
		counts := make([]int, layout.bitCount)
		for _, cube := range ordered {
			for part := 0; part < layout.bitCount; part++ {
				if err := work.step(); err != nil {
					return nil, err
				}
				present, _ := cube.Has(part)
				if present {
					counts[part]++
				}
			}
		}
		for _, cube := range ordered {
			for part := 0; part < layout.bitCount; part++ {
				if err := work.step(); err != nil {
					return nil, err
				}
				present, _ := cube.Has(part)
				if present {
					scores[cube] += counts[part]
				}
			}
		}
	}
	slices.SortStableFunc(ordered, func(first, second *EspressoSet) int {
		return scores[second] - scores[first]
	})
	return ordered, nil
}

// ReduceCover performs source-ordered in-place-equivalent reductions on an
// owned working copy and returns only nonempty cubes. Distance order is the
// source's sort_reduce strategy; column order is its mini_sort strategy.
func (layout *EspressoCubeLayout) ReduceCover(on, dontCare *EspressoSetFamily, options EspressoReduceOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoReduceWork(options)
	if err != nil {
		return nil, err
	}
	if err = layout.check(); err != nil {
		return nil, err
	}
	if err = on.check(); err != nil {
		return nil, err
	}
	if err = dontCare.check(); err != nil {
		return nil, err
	}
	if on.bits != layout.bitCount || dontCare.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(on.sets) > work.options.MaxCubes || len(dontCare.sets) > work.options.MaxCubes-len(on.sets)+1 {
		return nil, ErrLimit
	}
	ordered, err := layout.reduceOrder(on, options.Order, work)
	if err != nil {
		return nil, err
	}
	for index, target := range ordered {
		other, _ := NewEspressoSetFamily(len(ordered)-1, layout.bitCount)
		for current, cube := range ordered {
			if current != index {
				_ = other.Add(cube)
			}
		}
		reduced, err := layout.reduceCubeAgainst(target, other, dontCare, work)
		if err != nil {
			return nil, err
		}
		equal, _ := reduced.Equal(target)
		reduced.Rank = target.Rank
		reduced.Flags = target.Flags | EspressoSetActive
		if equal {
			reduced.Flags |= EspressoSetPrime
		} else {
			reduced.Flags &^= EspressoSetPrime
		}
		ordered[index] = reduced
	}
	result, _ := NewEspressoSetFamily(len(ordered), layout.bitCount)
	for _, cube := range ordered {
		empty, _ := cube.Empty()
		if !empty {
			_ = result.Add(cube)
		}
	}
	return result, nil
}
