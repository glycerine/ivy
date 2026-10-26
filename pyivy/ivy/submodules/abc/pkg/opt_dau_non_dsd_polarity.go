package abc

// Shared-minterm polarity variants from opt/dau/dauNonDsd.c:Dau_DecPerform6.

import "slices"

type DauSetPolarityOptions struct {
	MaxVariants, MaxWords int
	Stop                  func() bool
}

func cloneDauSetDecomposition(source DauSetDecomposition) DauSetDecomposition {
	return DauSetDecomposition{
		Found:           source.Found,
		Variables:       source.Variables,
		Composition:     source.Composition.Clone(),
		Decomposition:   source.Decomposition.Clone(),
		FreeVariables:   slices.Clone(source.FreeVariables),
		UniqueVariables: slices.Clone(source.UniqueVariables),
		SharedVariables: slices.Clone(source.SharedVariables),
	}
}

// PolarityVariants returns the source family obtained by complementing the
// selector independently for shared-variable minterms. Phase bit zero is fixed
// to zero, eliminating the duplicate family caused by globally complementing
// the selector. Discovery supports at most three shared variables, avoiding
// the source's overflow-prone shifts for larger counts.
func (result DauSetDecomposition) PolarityVariants(options DauSetPolarityOptions) ([]DauSetDecomposition, error) {
	if options.MaxVariants < 0 || options.MaxWords < 0 {
		return nil, ErrDimension
	}
	if !result.Found || result.Composition == nil || result.Decomposition == nil || len(result.SharedVariables) > 3 {
		return nil, ErrDimension
	}
	if _, err := result.Recompose(options.MaxWords, options.Stop); err != nil {
		return nil, err
	}
	count := 1 << uint((1<<uint(len(result.SharedVariables)))-1)
	if options.MaxVariants != 0 && count > options.MaxVariants {
		return nil, ErrLimit
	}
	variants := make([]DauSetDecomposition, 0, count)
	phaseLimit := 1 << uint(1<<uint(len(result.SharedVariables)))
	for phaseMask := 0; phaseMask < phaseLimit; phaseMask += 2 {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		variant := cloneDauSetDecomposition(result)
		decomposition, err := NewConstantTruthTable(result.Decomposition.variables, false, options.MaxWords)
		if err != nil {
			return nil, err
		}
		for assignment := 0; assignment < 1<<uint(result.Decomposition.variables); assignment++ {
			sharedBits := assignment >> uint(len(result.UniqueVariables))
			value, err := result.Decomposition.Get(uint64(assignment))
			if err != nil {
				return nil, err
			}
			if phaseMask&(1<<uint(sharedBits)) != 0 {
				value = !value
			}
			if value {
				if err := decomposition.Set(uint64(assignment), true); err != nil {
					return nil, err
				}
			}
		}
		composition, err := NewConstantTruthTable(result.Composition.variables, false, options.MaxWords)
		if err != nil {
			return nil, err
		}
		freeCount := len(result.FreeVariables)
		sharedCount := len(result.SharedVariables)
		selectorPosition := freeCount + sharedCount
		for assignment := 0; assignment < 1<<uint(result.Composition.variables); assignment++ {
			sharedBits := (assignment >> uint(freeCount)) & ((1 << uint(sharedCount)) - 1)
			sourceAssignment := assignment
			if phaseMask&(1<<uint(sharedBits)) != 0 {
				sourceAssignment ^= 1 << uint(selectorPosition)
			}
			value, err := result.Composition.Get(uint64(sourceAssignment))
			if err != nil {
				return nil, err
			}
			if value {
				if err := composition.Set(uint64(assignment), true); err != nil {
					return nil, err
				}
			}
		}
		variant.Decomposition = decomposition
		variant.Composition = composition
		variants = append(variants, variant)
	}
	return variants, nil
}
