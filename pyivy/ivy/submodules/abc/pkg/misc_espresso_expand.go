package abc

// Prime expansion from src/misc/espresso/expand.c.

import "slices"

type EspressoExpandOptions struct {
	Nonsparse                         bool
	SparseVariables                   []bool
	MaxCubes, MaxOperations, MaxDepth int
	Stop                              func() bool
}

type EspressoExpansion struct {
	Cube, Overexpanded *EspressoSet
	CoveredIndices     []int
	PrimeCandidates    int
	Nonessential       bool
}

func normalizeEspressoExpandOptions(layout *EspressoCubeLayout, options EspressoExpandOptions) (EspressoExpandOptions, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxDepth < 0 {
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

func (layout *EspressoCubeLayout) expandPrimeUniverse(off *EspressoSetFamily, options EspressoExpandOptions) (*EspressoSetFamily, error) {
	allowed, err := layout.ComplementCover(off, EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	list, err := NewEspressoCubeList(layout, allowed)
	if err != nil {
		return nil, err
	}
	primes, err := layout.PrimeConsensus(list, EspressoPrimeOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
	if err != nil {
		return nil, err
	}
	for _, prime := range primes.sets {
		prime.Flags |= EspressoSetPrime
		prime.Flags &^= EspressoSetCovered | EspressoSetNonessential
	}
	return primes, nil
}

func (layout *EspressoCubeLayout) restrictSparseExpansion(prime, source *EspressoSet, options EspressoExpandOptions) *EspressoSet {
	result, _ := prime.Clone()
	if !options.Nonsparse {
		return result
	}
	for variable, sparse := range options.SparseVariables {
		if !sparse {
			continue
		}
		start := layout.firstParts[variable]
		for part := start; part < start+layout.partSizes[variable]; part++ {
			present, _ := source.Has(part)
			_ = result.Set(part, present)
		}
	}
	return result
}

func (layout *EspressoCubeLayout) expansionCandidates(source *EspressoSet, primes *EspressoSetFamily, options EspressoExpandOptions) (*EspressoSetFamily, error) {
	if err := layout.checkSet(source); err != nil {
		return nil, err
	}
	candidates, _ := NewEspressoSetFamily(len(primes.sets), layout.bitCount)
	operations := 0
	for _, prime := range primes.sets {
		operations += max(1, len(prime.words))
		if operations > options.MaxOperations {
			return nil, ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		contained, _ := source.Implies(prime)
		if !contained {
			continue
		}
		candidate := layout.restrictSparseExpansion(prime, source, options)
		candidate.Flags |= EspressoSetPrime
		_ = candidates.Add(candidate)
	}
	if len(candidates.sets) == 0 {
		return nil, ErrInvariant
	}
	return candidates.Contain(EspressoContainOptions{MaxSets: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
}

func (layout *EspressoCubeLayout) chooseExpansion(source *EspressoSet, sourceIndex int, on, primes *EspressoSetFamily, eligible []bool, options EspressoExpandOptions) (EspressoExpansion, error) {
	candidates, err := layout.expansionCandidates(source, primes, options)
	if err != nil {
		return EspressoExpansion{}, err
	}
	overexpanded, _ := source.Clone()
	for _, candidate := range candidates.sets {
		for word, value := range candidate.words {
			overexpanded.words[word] |= value
		}
	}
	best, bestCovered, bestWeight := -1, -1, 0
	bestIndices := []int(nil)
	operations := 0
	for candidateIndex, candidate := range candidates.sets {
		covered := make([]int, 0)
		for index, cube := range on.sets {
			operations += max(1, len(cube.words))
			if operations > options.MaxOperations {
				return EspressoExpansion{}, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return EspressoExpansion{}, ErrCancelled
			}
			if index == sourceIndex || eligible != nil && !eligible[index] {
				continue
			}
			implied, _ := cube.Implies(candidate)
			if implied {
				covered = append(covered, index)
			}
		}
		weight := layout.exactLiteralWeight(candidate, options.SparseVariables)
		if len(covered) > bestCovered || len(covered) == bestCovered && (best < 0 || weight < bestWeight) {
			best, bestCovered, bestWeight, bestIndices = candidateIndex, len(covered), weight, covered
		}
	}
	chosen, _ := candidates.sets[best].Clone()
	chosen.Flags |= EspressoSetPrime
	equalOver, _ := chosen.Equal(overexpanded)
	return EspressoExpansion{Cube: chosen, Overexpanded: overexpanded, CoveredIndices: bestIndices, PrimeCandidates: len(candidates.sets), Nonessential: len(bestIndices) == 0 && !equalOver}, nil
}

// ExpandCube returns a deterministic prime (or dense-variable prime in
// nonsparse mode) containing source and avoiding the OFF cover.
func (layout *EspressoCubeLayout) ExpandCube(source *EspressoSet, on, off *EspressoSetFamily, options EspressoExpandOptions) (EspressoExpansion, error) {
	if err := layout.check(); err != nil {
		return EspressoExpansion{}, err
	}
	for _, family := range []*EspressoSetFamily{on, off} {
		if err := family.check(); err != nil {
			return EspressoExpansion{}, err
		}
		if family.bits != layout.bitCount {
			return EspressoExpansion{}, ErrDimension
		}
	}
	if err := layout.checkSet(source); err != nil {
		return EspressoExpansion{}, err
	}
	options, err := normalizeEspressoExpandOptions(layout, options)
	if err != nil {
		return EspressoExpansion{}, err
	}
	if len(on.sets)+len(off.sets) > options.MaxCubes {
		return EspressoExpansion{}, ErrLimit
	}
	for _, blocker := range off.sets {
		intersects, _ := layout.DistanceZero(source, blocker)
		if intersects {
			return EspressoExpansion{}, ErrInvariant
		}
	}
	primes, err := layout.expandPrimeUniverse(off, options)
	if err != nil {
		return EspressoExpansion{}, err
	}
	sourceIndex := -1
	for index, cube := range on.sets {
		equal, _ := cube.Equal(source)
		if equal {
			sourceIndex = index
			break
		}
	}
	return layout.chooseExpansion(source, sourceIndex, on, primes, nil, options)
}

func (layout *EspressoCubeLayout) expandOrder(on *EspressoSetFamily, options EspressoExpandOptions) ([]*EspressoSet, error) {
	ordered := make([]*EspressoSet, len(on.sets))
	counts := make([]int, layout.bitCount)
	operations := 0
	for index, cube := range on.sets {
		ordered[index], _ = cube.Clone()
		for part := 0; part < layout.bitCount; part++ {
			operations++
			if operations > options.MaxOperations {
				return nil, ErrLimit
			}
			present, _ := cube.Has(part)
			if present {
				counts[part]++
			}
		}
	}
	scores := make(map[*EspressoSet]int, len(ordered))
	for _, cube := range ordered {
		for part, count := range counts {
			present, _ := cube.Has(part)
			if present {
				scores[cube] += count
			}
		}
	}
	slices.SortStableFunc(ordered, func(first, second *EspressoSet) int {
		if scores[first] != scores[second] {
			return scores[first] - scores[second]
		}
		comparison, _ := EspressoSetCompareLexical(first, second)
		return -comparison
	})
	return ordered, nil
}

// ExpandCover expands nonprime cubes in source order after the source's
// ascending column-inner-product ordering and removes cubes covered by an
// earlier expansion.
func (layout *EspressoCubeLayout) ExpandCover(on, off *EspressoSetFamily, options EspressoExpandOptions) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	for _, family := range []*EspressoSetFamily{on, off} {
		if err := family.check(); err != nil {
			return nil, err
		}
		if family.bits != layout.bitCount {
			return nil, ErrDimension
		}
	}
	options, err := normalizeEspressoExpandOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if len(on.sets)+len(off.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	for _, cube := range on.sets {
		for _, blocker := range off.sets {
			intersects, _ := layout.DistanceZero(cube, blocker)
			if intersects {
				return nil, ErrInvariant
			}
		}
	}
	primes, err := layout.expandPrimeUniverse(off, options)
	if err != nil {
		return nil, err
	}
	ordered, err := layout.expandOrder(on, options)
	if err != nil {
		return nil, err
	}
	covered := make([]bool, len(ordered))
	for index, cube := range ordered {
		cube.Flags &^= EspressoSetCovered | EspressoSetNonessential
		if covered[index] || cube.Flags&EspressoSetPrime != 0 {
			continue
		}
		eligible := make([]bool, len(ordered))
		for current, candidate := range ordered {
			eligible[current] = current != index && !covered[current] && candidate.Flags&EspressoSetPrime == 0
		}
		expansion, expandErr := layout.chooseExpansion(cube, index, &EspressoSetFamily{bits: layout.bitCount, sets: ordered}, primes, eligible, options)
		if expandErr != nil {
			return nil, expandErr
		}
		expansion.Cube.Rank = cube.Rank
		expansion.Cube.Flags |= cube.Flags | EspressoSetPrime
		if expansion.Nonessential {
			expansion.Cube.Flags |= EspressoSetNonessential
		}
		ordered[index] = expansion.Cube
		for _, current := range expansion.CoveredIndices {
			covered[current] = true
		}
	}
	result, _ := NewEspressoSetFamily(len(ordered), layout.bitCount)
	for index, cube := range ordered {
		if covered[index] {
			continue
		}
		cube.Flags &^= EspressoSetCovered
		cube.Flags |= EspressoSetActive
		_ = result.Add(cube)
	}
	return result, nil
}

func (layout *EspressoCubeLayout) FindAllPrimesCoveringCube(source *EspressoSet, off *EspressoSetFamily, options EspressoExpandOptions) (*EspressoSetFamily, error) {
	options, err := normalizeEspressoExpandOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if err = layout.checkSet(source); err != nil {
		return nil, err
	}
	primes, err := layout.expandPrimeUniverse(off, options)
	if err != nil {
		return nil, err
	}
	options.Nonsparse = false
	return layout.expansionCandidates(source, primes, options)
}

func (layout *EspressoCubeLayout) AllPrimes(on, off *EspressoSetFamily, options EspressoExpandOptions) (*EspressoSetFamily, error) {
	options, err := normalizeEspressoExpandOptions(layout, options)
	if err != nil {
		return nil, err
	}
	if err = on.check(); err != nil {
		return nil, err
	}
	if on.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if err = off.check(); err != nil {
		return nil, err
	}
	if off.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(on.sets)+len(off.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	primes, err := layout.expandPrimeUniverse(off, options)
	if err != nil {
		return nil, err
	}
	result, _ := NewEspressoSetFamily(len(on.sets), layout.bitCount)
	for _, cube := range on.sets {
		if cube.Flags&EspressoSetPrime != 0 {
			if len(result.sets) >= options.MaxCubes {
				return nil, ErrLimit
			}
			_ = result.Add(cube)
			continue
		}
		candidates, candidateErr := layout.expansionCandidates(cube, primes, EspressoExpandOptions{SparseVariables: options.SparseVariables, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
		if candidateErr != nil {
			return nil, candidateErr
		}
		for _, candidate := range candidates.sets {
			if len(result.sets) >= options.MaxCubes {
				return nil, ErrLimit
			}
			_ = result.Add(candidate)
		}
	}
	return result, nil
}
