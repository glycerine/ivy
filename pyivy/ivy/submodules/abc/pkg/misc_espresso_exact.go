package abc

// Exact prime-cover minimization from src/misc/espresso/exact.c.

import "slices"

type EspressoExactOptions struct {
	SparseVariables                   []bool
	MaxCubes, MaxOperations, MaxDepth int
	MaxSearchNodes                    int
	Stop                              func() bool
}

type EspressoExactResult struct {
	Cover, Primes *EspressoSetFamily
	Selection     []int
	Cost          int
	Exact         bool
	Weighted      bool
}

func normalizeEspressoExactOptions(layout *EspressoCubeLayout, options EspressoExactOptions) (EspressoExactOptions, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxDepth < 0 || options.MaxSearchNodes < 0 {
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
	if options.MaxSearchNodes == 0 {
		options.MaxSearchNodes = 1 << 20
	}
	if len(options.SparseVariables) == 0 {
		options.SparseVariables, _ = layout.DefaultSparseVariables()
	} else {
		options.SparseVariables = slices.Clone(options.SparseVariables)
	}
	return options, nil
}

func (layout *EspressoCubeLayout) exactLiteralWeight(cube *EspressoSet, sparse []bool) int {
	weight := 0
	for variable, size := range layout.partSizes {
		start := layout.firstParts[variable]
		present := 0
		for part := start; part < start+size; part++ {
			if cube.words[part>>5]&(uint32(1)<<uint(part&31)) != 0 {
				present++
			}
		}
		if variable == layout.OutputVariable() || sparse[variable] {
			weight += present
		} else {
			weight += size - present
		}
	}
	return weight
}

func (layout *EspressoCubeLayout) exactSelectedCover(primes *EspressoSetFamily, selected []bool) *EspressoSetFamily {
	count := 0
	for _, yes := range selected {
		if yes {
			count++
		}
	}
	result, _ := NewEspressoSetFamily(count, layout.bitCount)
	for index, yes := range selected {
		if yes {
			_ = result.Add(primes.sets[index])
		}
	}
	return result
}

func (layout *EspressoCubeLayout) exactCoversOn(on, dontCare, primes *EspressoSetFamily, selected []bool, work *espressoSharpWork) (bool, error) {
	chosen := layout.exactSelectedCover(primes, selected)
	allowed, err := JoinEspressoSetFamilies(chosen, dontCare)
	if err != nil {
		return false, err
	}
	remainder, err := layout.coverDifference(on, allowed, work)
	if err != nil {
		return false, err
	}
	return len(remainder.sets) == 0, nil
}

func espressoExactObjective(selected []bool, weights []int, weighted bool) (primary, secondary int) {
	for index, yes := range selected {
		if !yes {
			continue
		}
		secondary += weights[index]
		primary++
	}
	if weighted {
		return secondary, primary
	}
	return primary, secondary
}

func espressoExactBetter(primary, secondary, bestPrimary, bestSecondary int) bool {
	return primary < bestPrimary || primary == bestPrimary && secondary < bestSecondary
}

func (layout *EspressoCubeLayout) minimizeExact(on, dontCare, off *EspressoSetFamily, exactCover, weighted bool, options EspressoExactOptions) (EspressoExactResult, error) {
	if err := layout.check(); err != nil {
		return EspressoExactResult{}, err
	}
	for _, family := range []*EspressoSetFamily{on, dontCare, off} {
		if err := family.check(); err != nil {
			return EspressoExactResult{}, err
		}
		if family.bits != layout.bitCount {
			return EspressoExactResult{}, ErrDimension
		}
	}
	options, err := normalizeEspressoExactOptions(layout, options)
	if err != nil {
		return EspressoExactResult{}, err
	}
	if len(on.sets)+len(dontCare.sets)+len(off.sets) > options.MaxCubes {
		return EspressoExactResult{}, ErrLimit
	}
	preOperations := 0
	for _, pair := range [][2]*EspressoSetFamily{{on, dontCare}, {on, off}, {dontCare, off}} {
		for _, first := range pair[0].sets {
			for _, second := range pair[1].sets {
				preOperations += max(1, layout.VariableCount())
				if preOperations > options.MaxOperations {
					return EspressoExactResult{}, ErrLimit
				}
				if options.Stop != nil && options.Stop() {
					return EspressoExactResult{}, ErrCancelled
				}
				intersects, _ := layout.DistanceZero(first, second)
				if intersects {
					return EspressoExactResult{}, ErrInvariant
				}
			}
		}
	}
	allowed, err := JoinEspressoSetFamilies(on, dontCare)
	if err != nil {
		return EspressoExactResult{}, err
	}
	list, err := NewEspressoCubeList(layout, allowed)
	if err != nil {
		return EspressoExactResult{}, err
	}
	if options.MaxOperations-preOperations <= 0 {
		return EspressoExactResult{}, ErrLimit
	}
	primeWork, err := newEspressoPrimeWork(EspressoPrimeOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations - preOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
	if err != nil {
		return EspressoExactResult{}, err
	}
	cofactor, _ := list.cofactor.Clone()
	generated, err := layout.primeConsensus(layout.primeListClone(list, cofactor), 0, primeWork)
	if err != nil {
		return EspressoExactResult{}, err
	}
	remainingCubes := options.MaxCubes - primeWork.generated
	remainingOperations := options.MaxOperations - preOperations - primeWork.operations
	if remainingCubes <= 0 || remainingOperations <= 0 {
		return EspressoExactResult{}, ErrLimit
	}
	sharpWork, err := newEspressoSharpWork(EspressoSharpOptions{MaxCubes: remainingCubes, MaxOperations: remainingOperations, Stop: options.Stop})
	if err != nil {
		return EspressoExactResult{}, err
	}
	primes, _ := NewEspressoSetFamily(len(generated.sets), layout.bitCount)
	for _, prime := range generated.sets {
		care, careErr := layout.sharpCubeByCover(prime, dontCare, sharpWork)
		if careErr != nil {
			return EspressoExactResult{}, careErr
		}
		if len(care.sets) == 0 {
			continue
		}
		clone, _ := prime.Clone()
		clone.Flags |= EspressoSetPrime
		clone.Flags &^= EspressoSetActive
		_ = primes.Add(clone)
	}
	weights := make([]int, len(primes.sets))
	for index, prime := range primes.sets {
		weights[index] = layout.exactLiteralWeight(prime, options.SparseVariables)
	}
	selected := make([]bool, len(primes.sets))
	for index := range selected {
		selected[index] = true
	}
	covers, err := layout.exactCoversOn(on, dontCare, primes, selected, sharpWork)
	if err != nil {
		return EspressoExactResult{}, err
	}
	if !covers {
		return EspressoExactResult{}, ErrInvariant
	}
	// Deterministic irredundant upper bound, also the heuristic result.
	for index := range selected {
		selected[index] = false
		covers, err = layout.exactCoversOn(on, dontCare, primes, selected, sharpWork)
		if err != nil {
			return EspressoExactResult{}, err
		}
		if !covers {
			selected[index] = true
		}
	}
	best := slices.Clone(selected)
	bestPrimary, bestSecondary := espressoExactObjective(best, weights, weighted)
	if exactCover {
		current := make([]bool, len(primes.sets))
		nodes := 0
		var search func(int, int, int) error
		search = func(index, primary, secondary int) error {
			nodes++
			if nodes > options.MaxSearchNodes {
				return ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return ErrCancelled
			}
			objectivePrimary, objectiveSecondary := primary, secondary
			if weighted {
				objectivePrimary, objectiveSecondary = secondary, primary
			}
			if objectivePrimary > bestPrimary || objectivePrimary == bestPrimary && objectiveSecondary >= bestSecondary {
				return nil
			}
			covered, coverErr := layout.exactCoversOn(on, dontCare, primes, current, sharpWork)
			if coverErr != nil {
				return coverErr
			}
			if covered {
				best = slices.Clone(current)
				bestPrimary, bestSecondary = objectivePrimary, objectiveSecondary
				return nil
			}
			if index == len(current) {
				return nil
			}
			current[index] = true
			if err := search(index+1, primary+1, secondary+weights[index]); err != nil {
				return err
			}
			current[index] = false
			return search(index+1, primary, secondary)
		}
		if err = search(0, 0, 0); err != nil {
			return EspressoExactResult{}, err
		}
	}
	cover := layout.exactSelectedCover(primes, best)
	selection := make([]int, 0, len(cover.sets))
	for index, yes := range best {
		if yes {
			selection = append(selection, index)
		}
	}
	primesClone, _ := primes.Clone()
	return EspressoExactResult{Cover: cover, Primes: primesClone, Selection: selection, Cost: bestPrimary, Exact: exactCover, Weighted: weighted}, nil
}

func (layout *EspressoCubeLayout) MinimizeExact(on, dontCare, off *EspressoSetFamily, exactCover bool, options EspressoExactOptions) (EspressoExactResult, error) {
	return layout.minimizeExact(on, dontCare, off, exactCover, false, options)
}

func (layout *EspressoCubeLayout) MinimizeExactLiterals(on, dontCare, off *EspressoSetFamily, exactCover bool, options EspressoExactOptions) (EspressoExactResult, error) {
	return layout.minimizeExact(on, dontCare, off, exactCover, true, options)
}
