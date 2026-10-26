package abc

// Last-gasp and super-gasp cover improvement from src/misc/espresso/gasp.c.

import "slices"

type EspressoGaspOptions struct {
	SparseVariables                  []bool
	ExactIrredundant                 bool
	MaxCubes, MaxOperations, MaxRows int
	MaxDepth                         int
	Stop                             func() bool
}

func normalizeEspressoGaspOptions(layout *EspressoCubeLayout, options EspressoGaspOptions) (EspressoGaspOptions, error) {
	if err := layout.check(); err != nil {
		return options, err
	}
	if options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxRows < 0 || options.MaxDepth < 0 {
		return options, ErrDimension
	}
	if len(options.SparseVariables) != 0 && len(options.SparseVariables) != layout.VariableCount() {
		return options, ErrDimension
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

func (options EspressoGaspOptions) reduceOptions() EspressoReduceOptions {
	return EspressoReduceOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop}
}

func (options EspressoGaspOptions) expandOptions() EspressoExpandOptions {
	return EspressoExpandOptions{SparseVariables: options.SparseVariables, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop}
}

func (options EspressoGaspOptions) irredundantOptions() EspressoIrredundantOptions {
	return EspressoIrredundantOptions{Exact: options.ExactIrredundant, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, Stop: options.Stop}
}

func (layout *EspressoCubeLayout) validateGaspFamilies(families ...*EspressoSetFamily) error {
	if err := layout.check(); err != nil {
		return err
	}
	for _, family := range families {
		if err := family.check(); err != nil {
			return err
		}
		if family.bits != layout.bitCount {
			return ErrDimension
		}
	}
	return nil
}

// ReduceGasp independently maximally reduces each ON cube against the
// original ON and DC covers. Unchanged cubes are marked prime; changed cubes
// are nonprime. Unlike ReduceCover, reductions do not replace earlier rows.
func (layout *EspressoCubeLayout) ReduceGasp(on, dontCare *EspressoSetFamily, options EspressoGaspOptions) (*EspressoSetFamily, error) {
	if err := layout.validateGaspFamilies(on, dontCare); err != nil {
		return nil, err
	}
	options, err := normalizeEspressoGaspOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if len(on.sets)+len(dontCare.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	work, err := newEspressoReduceWork(options.reduceOptions())
	if err != nil {
		return nil, err
	}
	result, _ := NewEspressoSetFamily(len(on.sets), layout.bitCount)
	for index, target := range on.sets {
		other, _ := NewEspressoSetFamily(len(on.sets)-1, layout.bitCount)
		for current, cube := range on.sets {
			if current != index {
				_ = other.Add(cube)
			}
		}
		reduced, reduceErr := layout.reduceCubeAgainst(target, other, dontCare, work)
		if reduceErr != nil {
			return nil, reduceErr
		}
		empty, _ := reduced.Empty()
		if empty {
			return nil, ErrInvariant
		}
		equal, _ := reduced.Equal(target)
		reduced.Rank = target.Rank
		reduced.Flags = target.Flags | EspressoSetActive
		if equal {
			reduced.Flags |= EspressoSetPrime
		} else {
			reduced.Flags &^= EspressoSetPrime
		}
		_ = result.Add(reduced)
	}
	return result, nil
}

func espressoGaspUnion(first, second *EspressoSet) *EspressoSet {
	result, _ := first.Clone()
	for index, word := range second.words {
		result.words[index] |= word
	}
	result.Flags &^= EspressoSetPrime | EspressoSetCovered | EspressoSetNonessential
	result.Flags |= EspressoSetActive
	return result
}

func (layout *EspressoCubeLayout) gaspCandidateForPair(reduced, dontCare, original *EspressoSetFamily, first, second int, primes *EspressoSetFamily, work *espressoSharpWork) (*EspressoSet, bool, error) {
	modified, _ := original.Clone()
	replacement, _ := reduced.sets[first].Clone()
	modified.sets[first] = replacement

	other, _ := NewEspressoSetFamily(len(modified.sets)-1, layout.bitCount)
	for index, cube := range modified.sets {
		if index != second {
			_ = other.Add(cube)
		}
	}
	essential, err := layout.reduceCubeAgainst(modified.sets[second], other, dontCare, work)
	if err != nil {
		return nil, false, err
	}
	empty, _ := essential.Empty()
	if empty {
		return nil, false, nil
	}
	seed := espressoGaspUnion(reduced.sets[first], essential)
	for _, prime := range primes.sets {
		if err := work.step(); err != nil {
			return nil, false, err
		}
		contained, _ := seed.Implies(prime)
		if contained {
			return seed, true, nil
		}
	}
	return nil, false, nil
}

func (layout *EspressoCubeLayout) expandGaspCube(reduced, dontCare, original *EspressoSetFamily, index int, options EspressoGaspOptions, primes *EspressoSetFamily, work *espressoSharpWork) (*EspressoSetFamily, error) {
	result, _ := NewEspressoSetFamily(len(reduced.sets), layout.bitCount)
	for second, cube := range reduced.sets {
		if second == index || cube.Flags&EspressoSetPrime != 0 {
			continue
		}
		candidate, feasible, err := layout.gaspCandidateForPair(reduced, dontCare, original, index, second, primes, work)
		if err != nil {
			return nil, err
		}
		if feasible {
			if len(result.sets) >= options.MaxCubes {
				return nil, ErrLimit
			}
			_ = result.Add(candidate)
		}
	}
	return result, nil
}

// ExpandGaspCube returns the nonprime candidate cubes produced while trying
// to cover other independently reduced rows with an expansion of one row.
func (layout *EspressoCubeLayout) ExpandGaspCube(reduced, dontCare, off, original *EspressoSetFamily, index int, options EspressoGaspOptions) (*EspressoSetFamily, error) {
	if err := layout.validateGaspFamilies(reduced, dontCare, off, original); err != nil {
		return nil, err
	}
	options, err := normalizeEspressoGaspOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if len(reduced.sets) != len(original.sets) || index < 0 || index >= len(reduced.sets) {
		return nil, ErrDimension
	}
	if len(reduced.sets)+len(dontCare.sets)+len(off.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	primes, err := layout.expandPrimeUniverse(off, options.expandOptions())
	if err != nil {
		return nil, err
	}
	work, err := newEspressoReduceWork(options.reduceOptions())
	if err != nil {
		return nil, err
	}
	return layout.expandGaspCube(reduced, dontCare, original, index, options, primes, work)
}

// ExpandGasp generates all pair candidates, removes exact duplicates and
// expands the survivors into prime implicants against the OFF cover.
func (layout *EspressoCubeLayout) ExpandGasp(reduced, dontCare, off, original *EspressoSetFamily, options EspressoGaspOptions) (*EspressoSetFamily, error) {
	if err := layout.validateGaspFamilies(reduced, dontCare, off, original); err != nil {
		return nil, err
	}
	options, err := normalizeEspressoGaspOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if len(reduced.sets) != len(original.sets) {
		return nil, ErrDimension
	}
	if len(reduced.sets)+len(dontCare.sets)+len(off.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	primes, err := layout.expandPrimeUniverse(off, options.expandOptions())
	if err != nil {
		return nil, err
	}
	work, err := newEspressoReduceWork(options.reduceOptions())
	if err != nil {
		return nil, err
	}
	all, _ := NewEspressoSetFamily(len(reduced.sets), layout.bitCount)
	for index := range reduced.sets {
		candidates, candidateErr := layout.expandGaspCube(reduced, dontCare, original, index, options, primes, work)
		if candidateErr != nil {
			return nil, candidateErr
		}
		for _, candidate := range candidates.sets {
			if len(all.sets) >= options.MaxCubes {
				return nil, ErrLimit
			}
			_ = all.Add(candidate)
		}
	}
	if len(all.sets) == 0 {
		return all, nil
	}
	unique, err := all.Deduplicate(EspressoContainOptions{MaxSets: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	return layout.ExpandCover(unique, off, options.expandOptions())
}

// IrredundantGasp adds candidate primes to the ON cover and extracts an
// irredundant owned subset.
func (layout *EspressoCubeLayout) IrredundantGasp(on, dontCare, candidates *EspressoSetFamily, options EspressoGaspOptions) (*EspressoSetFamily, error) {
	if err := layout.validateGaspFamilies(on, dontCare, candidates); err != nil {
		return nil, err
	}
	options, err := normalizeEspressoGaspOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if len(on.sets)+len(dontCare.sets)+len(candidates.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	joined, err := JoinEspressoSetFamilies(on, candidates)
	if err != nil {
		return nil, err
	}
	marked, err := layout.Irredundant(joined, dontCare, options.irredundantOptions())
	if err != nil {
		return nil, err
	}
	return marked.Cover, nil
}

func (layout *EspressoCubeLayout) validateGaspPartition(on, dontCare, off *EspressoSetFamily, options EspressoGaspOptions) error {
	if err := layout.validateGaspFamilies(on, dontCare, off); err != nil {
		return err
	}
	if len(on.sets)+len(dontCare.sets)+len(off.sets) > options.MaxCubes {
		return ErrLimit
	}
	operations := 0
	// Minimized covers may intentionally absorb DC space. Only the active ON
	// and blocking OFF covers must be disjoint; DC may overlap either polarity.
	for _, pair := range [][2]*EspressoSetFamily{{on, off}} {
		for _, first := range pair[0].sets {
			for _, second := range pair[1].sets {
				operations += max(1, layout.VariableCount())
				if operations > options.MaxOperations {
					return ErrLimit
				}
				if options.Stop != nil && options.Stop() {
					return ErrCancelled
				}
				intersects, _ := layout.DistanceZero(first, second)
				if intersects {
					return ErrInvariant
				}
			}
		}
	}
	return nil
}

// LastGasp performs independent reduction, pair-candidate expansion and
// irredundant selection over the original cover plus new primes.
func (layout *EspressoCubeLayout) LastGasp(on, dontCare, off *EspressoSetFamily, options EspressoGaspOptions) (*EspressoSetFamily, error) {
	options, err := normalizeEspressoGaspOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if err = layout.validateGaspPartition(on, dontCare, off, options); err != nil {
		return nil, err
	}
	reduced, err := layout.ReduceGasp(on, dontCare, options)
	if err != nil {
		return nil, err
	}
	candidates, err := layout.ExpandGasp(reduced, dontCare, off, on, options)
	if err != nil {
		return nil, err
	}
	return layout.IrredundantGasp(on, dontCare, candidates, options)
}

// SuperGasp selects an irredundant subset from the original ON cover and all
// prime implicants containing an independently reduced cube.
func (layout *EspressoCubeLayout) SuperGasp(on, dontCare, off *EspressoSetFamily, options EspressoGaspOptions) (*EspressoSetFamily, error) {
	options, err := normalizeEspressoGaspOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if err = layout.validateGaspPartition(on, dontCare, off, options); err != nil {
		return nil, err
	}
	reduced, err := layout.ReduceGasp(on, dontCare, options)
	if err != nil {
		return nil, err
	}
	primes, err := layout.AllPrimes(reduced, off, options.expandOptions())
	if err != nil {
		return nil, err
	}
	joined, err := JoinEspressoSetFamilies(on, primes)
	if err != nil {
		return nil, err
	}
	unique, err := joined.Deduplicate(EspressoContainOptions{MaxSets: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	marked, err := layout.Irredundant(unique, dontCare, options.irredundantOptions())
	if err != nil {
		return nil, err
	}
	return marked.Cover, nil
}
