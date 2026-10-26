package abc

// Sparse-variable cleanup from src/misc/espresso/sparse.c.

import "slices"

type EspressoMakeSparseOptions struct {
	SparseVariables                                 []bool
	ForceIrredundant                                bool
	ExactIrredundant                                bool
	MaxIterations, MaxCubes, MaxOperations, MaxRows int
	MaxDepth                                        int
	Stop                                            func() bool
}

func normalizeEspressoMakeSparseOptions(layout *EspressoCubeLayout, options EspressoMakeSparseOptions) (EspressoMakeSparseOptions, error) {
	if options.MaxIterations < 0 || options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxRows < 0 || options.MaxDepth < 0 {
		return options, ErrDimension
	}
	if len(options.SparseVariables) != 0 && len(options.SparseVariables) != layout.VariableCount() {
		return options, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 64
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 16 << 20
	}
	if options.MaxRows == 0 {
		options.MaxRows = 1 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 1 << 16
	}
	if len(options.SparseVariables) == 0 {
		options.SparseVariables, _ = layout.DefaultSparseVariables()
	} else {
		options.SparseVariables = slices.Clone(options.SparseVariables)
	}
	return options, nil
}

func (layout *EspressoCubeLayout) sparsePartCofactor(source *EspressoSetFamily, variable, part int) (*EspressoSetFamily, []int, error) {
	result, _ := NewEspressoSetFamily(len(source.sets), layout.bitCount)
	mapping := make([]int, 0, len(source.sets))
	start := layout.firstParts[variable]
	for index, cube := range source.sets {
		present, _ := cube.Has(part)
		if !present {
			continue
		}
		clone, _ := cube.Clone()
		for current := start; current < start+layout.partSizes[variable]; current++ {
			_ = clone.Set(current, current == part)
		}
		_ = result.Add(clone)
		mapping = append(mapping, index)
	}
	return result, mapping, nil
}

// MVReduce performs the source's per-part sparse reduction. A part is removed
// from an original cube when its single-part cofactor row is redundant.
func (layout *EspressoCubeLayout) MVReduce(on, dontCare *EspressoSetFamily, options EspressoMakeSparseOptions) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	for _, family := range []*EspressoSetFamily{on, dontCare} {
		if err := family.check(); err != nil {
			return nil, err
		}
		if family.bits != layout.bitCount {
			return nil, ErrDimension
		}
	}
	options, err := normalizeEspressoMakeSparseOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if len(on.sets)+len(dontCare.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	result, _ := on.Clone()
	for variable, sparse := range options.SparseVariables {
		if !sparse {
			continue
		}
		start := layout.firstParts[variable]
		for part := start; part < start+layout.partSizes[variable]; part++ {
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			cofactoredOn, mapping, err := layout.sparsePartCofactor(result, variable, part)
			if err != nil {
				return nil, err
			}
			cofactoredDC, _, err := layout.sparsePartCofactor(dontCare, variable, part)
			if err != nil {
				return nil, err
			}
			marked, err := layout.MarkIrredundant(cofactoredOn, cofactoredDC, EspressoIrredundantOptions{
				Exact: options.ExactIrredundant, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, Stop: options.Stop,
			})
			if err != nil {
				return nil, err
			}
			for cofactorIndex, sourceIndex := range mapping {
				if marked.Marked.sets[cofactorIndex].Flags&EspressoSetActive != 0 {
					continue
				}
				cube := result.sets[sourceIndex]
				if variable == layout.VariableCount()-1 || !layout.variableFull(cube, variable) {
					_ = cube.Set(part, false)
				}
				cube.Flags &^= EspressoSetPrime
			}
		}
	}
	filtered, _ := NewEspressoSetFamily(len(result.sets), layout.bitCount)
	for _, cube := range result.sets {
		keep := true
		for variable, sparse := range options.SparseVariables {
			if sparse && !layout.variableIntersection(cube, cube, variable) {
				keep = false
				break
			}
		}
		if keep {
			cube.Flags |= EspressoSetActive
			_ = filtered.Add(cube)
		}
	}
	return filtered, nil
}

// MakeSparse alternates sparse-variable reduction and dense-only expansion.
// Without ForceIrredundant it performs the source's single pass; otherwise it
// repeats while literal cost strictly changes, within MaxIterations.
func (layout *EspressoCubeLayout) MakeSparse(on, dontCare, off *EspressoSetFamily, options EspressoMakeSparseOptions) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	for _, family := range []*EspressoSetFamily{on, dontCare, off} {
		if err := family.check(); err != nil {
			return nil, err
		}
		if family.bits != layout.bitCount {
			return nil, ErrDimension
		}
	}
	options, err := normalizeEspressoMakeSparseOptions(layout, options)
	if err != nil {
		return nil, err
	}
	current, _ := on.Clone()
	costOptions := EspressoCoverCostOptions{SparseVariables: options.SparseVariables, MaxCubes: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop}
	best, err := layout.CoverCost(current, costOptions)
	if err != nil {
		return nil, err
	}
	for iteration := 0; iteration < options.MaxIterations; iteration++ {
		reduced, err := layout.MVReduce(current, dontCare, options)
		if err != nil {
			return nil, err
		}
		cost, err := layout.CoverCost(reduced, costOptions)
		if err != nil {
			return nil, err
		}
		current = reduced
		if cost.Total == best.Total {
			return current, nil
		}
		best = cost
		expanded, err := layout.ExpandCover(current, off, EspressoExpandOptions{
			Nonsparse: true, SparseVariables: options.SparseVariables, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop,
		})
		if err != nil {
			return nil, err
		}
		cost, err = layout.CoverCost(expanded, costOptions)
		if err != nil {
			return nil, err
		}
		current = expanded
		if cost.Total == best.Total || !options.ForceIrredundant {
			return current, nil
		}
		best = cost
	}
	return nil, ErrLimit
}
