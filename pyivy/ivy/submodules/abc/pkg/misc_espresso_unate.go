package abc

// Unate cover mapping, complementation and intersection from src/misc/espresso/unate.c.

type EspressoUnateOptions struct {
	MaxSets, MaxOperations int
	Stop                   func() bool
}

type espressoUnateWork struct {
	options    EspressoUnateOptions
	operations int
}

func newEspressoUnateWork(options EspressoUnateOptions) (*espressoUnateWork, error) {
	if options.MaxSets < 0 || options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	if options.MaxSets == 0 {
		options.MaxSets = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 16 << 20
	}
	return &espressoUnateWork{options: options}, nil
}

func (work *espressoUnateWork) step() error {
	work.operations++
	if work.operations > work.options.MaxOperations {
		return ErrLimit
	}
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func espressoUnateSubset(first, second *EspressoSet, work *espressoUnateWork) (bool, error) {
	for index, word := range first.words {
		if err := work.step(); err != nil {
			return false, err
		}
		if word&^second.words[index] != 0 {
			return false, nil
		}
	}
	return true, nil
}

func espressoUnateMinimal(sets []*EspressoSet, work *espressoUnateWork) ([]*EspressoSet, error) {
	keep := make([]bool, len(sets))
	for index := range keep {
		keep[index] = true
	}
	for current := range sets {
		if !keep[current] {
			continue
		}
		for other := range sets {
			if current == other || !keep[current] {
				continue
			}
			subset, err := espressoUnateSubset(sets[other], sets[current], work)
			if err != nil {
				return nil, err
			}
			if !subset {
				continue
			}
			equal, _ := sets[other].Equal(sets[current])
			if !equal || other < current {
				keep[current] = false
			}
		}
	}
	result := make([]*EspressoSet, 0, len(sets))
	for index, set := range sets {
		if keep[index] {
			result = append(result, set)
		}
	}
	return result, nil
}

func espressoUnateMaximal(sets []*EspressoSet, work *espressoUnateWork) ([]*EspressoSet, error) {
	keep := make([]bool, len(sets))
	for index := range keep {
		keep[index] = true
	}
	for current := range sets {
		if !keep[current] {
			continue
		}
		for other := range sets {
			if current == other || !keep[current] {
				continue
			}
			subset, err := espressoUnateSubset(sets[current], sets[other], work)
			if err != nil {
				return nil, err
			}
			if !subset {
				continue
			}
			equal, _ := sets[other].Equal(sets[current])
			if !equal || other < current {
				keep[current] = false
			}
		}
	}
	result := make([]*EspressoSet, 0, len(sets))
	for index, set := range sets {
		if keep[index] {
			result = append(result, set)
		}
	}
	return result, nil
}

// UnateComplement returns all inclusion-minimal sets that intersect every row
// of family. This is the concrete minimal-transversal meaning of unate_compl.
func UnateComplement(family *EspressoSetFamily, options EspressoUnateOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoUnateWork(options)
	if err != nil {
		return nil, err
	}
	if err = family.check(); err != nil {
		return nil, err
	}
	if len(family.sets) > work.options.MaxSets {
		return nil, ErrLimit
	}
	empty, _ := NewEspressoSet(family.bits)
	candidates := []*EspressoSet{empty}
	for _, row := range family.sets {
		rowEmpty, _ := row.Empty()
		if rowEmpty {
			return NewEspressoSetFamily(0, family.bits)
		}
		next := make([]*EspressoSet, 0, len(candidates))
		for _, candidate := range candidates {
			intersects := false
			for index, word := range candidate.words {
				if err = work.step(); err != nil {
					return nil, err
				}
				if word&row.words[index] != 0 {
					intersects = true
					break
				}
			}
			if intersects {
				clone, _ := candidate.Clone()
				next = append(next, clone)
				continue
			}
			for part := 0; part < family.bits; part++ {
				if err = work.step(); err != nil {
					return nil, err
				}
				present, _ := row.Has(part)
				if !present {
					continue
				}
				clone, _ := candidate.Clone()
				_ = clone.Set(part, true)
				next = append(next, clone)
				if len(next) > work.options.MaxSets {
					return nil, ErrLimit
				}
			}
		}
		candidates, err = espressoUnateMinimal(next, work)
		if err != nil {
			return nil, err
		}
	}
	result, _ := NewEspressoSetFamily(len(candidates), family.bits)
	for _, candidate := range candidates {
		_ = result.Add(candidate)
	}
	return result, nil
}

// ExactMinimumCovers returns the source's family of all inclusion-minimal
// column covers. Despite its historical name, exact_minimum_cover enumerates
// minimal rather than only minimum-cardinality covers.
func ExactMinimumCovers(family *EspressoSetFamily, options EspressoUnateOptions) (*EspressoSetFamily, error) {
	return UnateComplement(family, options)
}

// IntersectUnateFamilies forms nonempty pairwise intersections and retains
// inclusion-maximal results. largestOnly additionally keeps maximum cardinality.
func IntersectUnateFamilies(first, second *EspressoSetFamily, largestOnly bool, options EspressoUnateOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoUnateWork(options)
	if err != nil {
		return nil, err
	}
	if err = first.check(); err != nil {
		return nil, err
	}
	if err = second.check(); err != nil {
		return nil, err
	}
	if first.bits != second.bits {
		return nil, ErrDimension
	}
	if len(first.sets) > work.options.MaxSets || len(second.sets) > work.options.MaxSets {
		return nil, ErrLimit
	}
	sets := make([]*EspressoSet, 0)
	maximum := -1
	for _, left := range first.sets {
		for _, right := range second.sets {
			if err = work.step(); err != nil {
				return nil, err
			}
			intersection, nonempty, _ := left.IntersectionNonempty(right)
			if !nonempty {
				continue
			}
			cardinality, _ := intersection.Cardinality()
			if largestOnly && cardinality < maximum {
				continue
			}
			if largestOnly && cardinality > maximum {
				sets = sets[:0]
				maximum = cardinality
			}
			sets = append(sets, intersection)
			if len(sets) > work.options.MaxSets {
				return nil, ErrLimit
			}
		}
	}
	sets, err = espressoUnateMaximal(sets, work)
	if err != nil {
		return nil, err
	}
	result, _ := NewEspressoSetFamily(len(sets), first.bits)
	for _, set := range sets {
		_ = result.Add(set)
	}
	return result, nil
}

type EspressoUnateMapping struct {
	Layout         *EspressoCubeLayout
	Family         *EspressoSetFamily
	ActiveParts    []int
	UnateVariables []int
	PartZeros      []int
}

// MapCoverToUnate maps a unate positional cover to its missing-part matrix and
// returns all metadata needed for an owned inverse mapping.
func (layout *EspressoCubeLayout) MapCoverToUnate(cover *EspressoSetFamily, options EspressoUnateOptions) (EspressoUnateMapping, error) {
	work, err := newEspressoUnateWork(options)
	if err != nil {
		return EspressoUnateMapping{}, err
	}
	list, err := NewEspressoCubeList(layout, cover)
	if err != nil {
		return EspressoUnateMapping{}, err
	}
	counts, err := list.Analyze(EspressoCofactorOptions{MaxCubes: work.options.MaxSets, MaxChecks: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return EspressoUnateMapping{}, err
	}
	if counts.VariablesActive != counts.VariablesUnate {
		return EspressoUnateMapping{}, ErrInvariant
	}
	activeParts := make([]int, 0, counts.VariablesUnate)
	for part, zeros := range counts.PartZeros {
		if zeros > 0 {
			activeParts = append(activeParts, part)
		}
	}
	unateVariables := make([]int, 0, counts.VariablesUnate)
	for variable, unate := range counts.IsUnate {
		if unate {
			unateVariables = append(unateVariables, variable)
		}
	}
	if len(activeParts) != len(unateVariables) {
		return EspressoUnateMapping{}, ErrInvariant
	}
	mapped, _ := NewEspressoSetFamily(len(cover.sets), len(activeParts))
	for _, cube := range cover.sets {
		row, _ := NewEspressoSet(len(activeParts))
		for column, part := range activeParts {
			if err = work.step(); err != nil {
				return EspressoUnateMapping{}, err
			}
			present, _ := cube.Has(part)
			if !present {
				_ = row.Set(column, true)
			}
		}
		_ = mapped.Add(row)
	}
	layoutClone, _ := layout.Clone()
	return EspressoUnateMapping{Layout: layoutClone, Family: mapped, ActiveParts: activeParts, UnateVariables: unateVariables, PartZeros: slicesCloneInts(counts.PartZeros)}, nil
}

// ToCover maps a unate result family back to positional cubes.
func (mapping EspressoUnateMapping) ToCover(family *EspressoSetFamily, options EspressoUnateOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoUnateWork(options)
	if err != nil {
		return nil, err
	}
	if err = mapping.Layout.check(); err != nil {
		return nil, err
	}
	if err = family.check(); err != nil {
		return nil, err
	}
	if family.bits != len(mapping.ActiveParts) || len(mapping.ActiveParts) != len(mapping.UnateVariables) || len(mapping.PartZeros) != mapping.Layout.bitCount {
		return nil, ErrDimension
	}
	if len(family.sets) > work.options.MaxSets {
		return nil, ErrLimit
	}
	result, _ := NewEspressoSetFamily(len(family.sets), mapping.Layout.bitCount)
	for _, row := range family.sets {
		cube, _ := mapping.Layout.FullCube()
		for column, variable := range mapping.UnateVariables {
			selected, _ := row.Has(column)
			if !selected {
				continue
			}
			start := mapping.Layout.firstParts[variable]
			for part := start; part < start+mapping.Layout.partSizes[variable]; part++ {
				if err = work.step(); err != nil {
					return nil, err
				}
				if mapping.PartZeros[part] == 0 {
					_ = cube.Set(part, false)
				}
			}
		}
		_ = result.Add(cube)
	}
	return result, nil
}
