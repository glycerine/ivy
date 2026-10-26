package abc

// Non-DSD composition/decomposition construction from opt/dau/dauNonDsd.c.

import "slices"

type DauSetDecompositionOptions struct {
	MaxWords  int
	MaxChecks uint64
	Stop      func() bool
}

type DauSetDecomposition struct {
	Found                                           bool
	Variables                                       int
	Composition, Decomposition                      *TruthTable
	FreeVariables, UniqueVariables, SharedVariables []int
}

func normalizeDauSetDecompositionOptions(options DauSetDecompositionOptions) (DauSetDecompositionOptions, error) {
	if options.MaxWords < 0 {
		return options, ErrDimension
	}
	if options.MaxWords == 0 {
		options.MaxWords = 1 << 20
	}
	if options.MaxChecks == 0 {
		options.MaxChecks = 1 << 32
	}
	return options, nil
}

func dauSetVariables(set DauDecompositionSet, variables int) (free, unique, shared []int, err error) {
	if variables < 0 || variables > 16 {
		return nil, nil, nil, ErrDimension
	}
	for variable := 0; variable < variables; variable++ {
		role, roleErr := set.Role(variable)
		if roleErr != nil {
			return nil, nil, nil, roleErr
		}
		switch role {
		case DauDecompositionFree:
			free = append(free, variable)
		case DauDecompositionUnique:
			unique = append(unique, variable)
		case DauDecompositionShared:
			shared = append(shared, variable)
		}
	}
	if variables < 16 && uint32(set)>>uint(2*variables) != 0 {
		return nil, nil, nil, ErrDimension
	}
	return free, unique, shared, nil
}

// DecomposeDauVariableSet constructs F(free,shared,unique) as
// C(free,shared,D(unique,shared)). Cofactor zero is the first cofactor in
// ascending unique assignment order for each shared assignment, matching the
// source's old checker used by Dau_DecDecomposeSet.
func DecomposeDauVariableSet(table *TruthTable, set DauDecompositionSet, options DauSetDecompositionOptions) (DauSetDecomposition, error) {
	if err := table.Check(); err != nil {
		return DauSetDecomposition{}, err
	}
	if table.variables < 3 || table.variables > 16 {
		return DauSetDecomposition{}, ErrDimension
	}
	o, err := normalizeDauSetDecompositionOptions(options)
	if err != nil {
		return DauSetDecomposition{}, err
	}
	free, unique, shared, err := dauSetVariables(set, table.variables)
	if err != nil {
		return DauSetDecomposition{}, err
	}
	if len(free) == 0 || len(unique) < 2 || len(shared) > 6 {
		return DauSetDecomposition{}, ErrDimension
	}
	compositionVariables := len(free) + len(shared) + 1
	decompositionVariables := len(unique) + len(shared)
	composition, err := NewConstantTruthTable(compositionVariables, false, o.MaxWords)
	if err != nil {
		return DauSetDecomposition{}, err
	}
	decomposition, err := NewConstantTruthTable(decompositionVariables, false, o.MaxWords)
	if err != nil {
		return DauSetDecomposition{}, err
	}
	budget := &dauDecompositionCheckBudget{maximum: o.MaxChecks, stop: o.Stop}
	freeAssignments := 1 << uint(len(free))
	for sharedBits := 0; sharedBits < 1<<uint(len(shared)); sharedBits++ {
		var first, second []bool
		for uniqueBits := 0; uniqueBits < 1<<uint(len(unique)); uniqueBits++ {
			cofactor := make([]bool, freeAssignments)
			for freeBits := 0; freeBits < freeAssignments; freeBits++ {
				if err := budget.take(); err != nil {
					return DauSetDecomposition{}, err
				}
				assignment := uint64(0)
				for i, variable := range free {
					assignment |= uint64((freeBits>>uint(i))&1) << uint(variable)
				}
				for i, variable := range unique {
					assignment |= uint64((uniqueBits>>uint(i))&1) << uint(variable)
				}
				for i, variable := range shared {
					assignment |= uint64((sharedBits>>uint(i))&1) << uint(variable)
				}
				cofactor[freeBits], err = table.Get(assignment)
				if err != nil {
					return DauSetDecomposition{}, err
				}
			}
			class := 0
			if first == nil {
				first = cofactor
			} else if !slices.Equal(first, cofactor) {
				class = 1
				if second == nil {
					second = cofactor
				} else if !slices.Equal(second, cofactor) {
					return DauSetDecomposition{Found: false}, nil
				}
			}
			if class == 1 {
				assignment := uint64(uniqueBits | sharedBits<<uint(len(unique)))
				if err := decomposition.Set(assignment, true); err != nil {
					return DauSetDecomposition{}, err
				}
			}
		}
		if second == nil {
			second = first
		}
		for selector, cofactor := range [][]bool{first, second} {
			for freeBits, value := range cofactor {
				if !value {
					continue
				}
				assignment := uint64(freeBits | sharedBits<<uint(len(free)) | selector<<uint(len(free)+len(shared)))
				if err := composition.Set(assignment, true); err != nil {
					return DauSetDecomposition{}, err
				}
			}
		}
	}
	return DauSetDecomposition{
		Found:           true,
		Variables:       table.variables,
		Composition:     composition,
		Decomposition:   decomposition,
		FreeVariables:   slices.Clone(free),
		UniqueVariables: slices.Clone(unique),
		SharedVariables: slices.Clone(shared),
	}, nil
}

// Recompose rebuilds the original-variable truth table from an owned
// decomposition result. It is also an independent public verification path.
func (result DauSetDecomposition) Recompose(maxWords int, stop func() bool) (*TruthTable, error) {
	if !result.Found || result.Composition == nil || result.Decomposition == nil || result.Variables < 3 || result.Variables > 16 {
		return nil, ErrDimension
	}
	if err := result.Composition.Check(); err != nil {
		return nil, err
	}
	if err := result.Decomposition.Check(); err != nil {
		return nil, err
	}
	if len(result.FreeVariables)+len(result.UniqueVariables)+len(result.SharedVariables) != result.Variables ||
		result.Composition.variables != len(result.FreeVariables)+len(result.SharedVariables)+1 ||
		result.Decomposition.variables != len(result.UniqueVariables)+len(result.SharedVariables) {
		return nil, ErrInvariant
	}
	seen := make([]bool, result.Variables)
	for _, group := range [][]int{result.FreeVariables, result.UniqueVariables, result.SharedVariables} {
		for _, variable := range group {
			if variable < 0 || variable >= result.Variables || seen[variable] {
				return nil, ErrInvariant
			}
			seen[variable] = true
		}
	}
	table, err := NewConstantTruthTable(result.Variables, false, maxWords)
	if err != nil {
		return nil, err
	}
	for assignment := uint64(0); assignment < 1<<uint(result.Variables); assignment++ {
		if stop != nil && stop() {
			return nil, ErrCancelled
		}
		decompositionAssignment := uint64(0)
		for i, variable := range result.UniqueVariables {
			decompositionAssignment |= ((assignment >> uint(variable)) & 1) << uint(i)
		}
		for i, variable := range result.SharedVariables {
			decompositionAssignment |= ((assignment >> uint(variable)) & 1) << uint(len(result.UniqueVariables)+i)
		}
		selector, err := result.Decomposition.Get(decompositionAssignment)
		if err != nil {
			return nil, err
		}
		compositionAssignment := uint64(0)
		for i, variable := range result.FreeVariables {
			compositionAssignment |= ((assignment >> uint(variable)) & 1) << uint(i)
		}
		for i, variable := range result.SharedVariables {
			compositionAssignment |= ((assignment >> uint(variable)) & 1) << uint(len(result.FreeVariables)+i)
		}
		if selector {
			compositionAssignment |= 1 << uint(len(result.FreeVariables)+len(result.SharedVariables))
		}
		value, err := result.Composition.Get(compositionAssignment)
		if err != nil {
			return nil, err
		}
		if value {
			if err := table.Set(assignment, true); err != nil {
				return nil, err
			}
		}
	}
	return table, nil
}
