package abc

// Recursive prime generation by consensus from src/misc/espresso/primes.c.

type EspressoPrimeOptions struct {
	MaxCubes, MaxOperations, MaxDepth int
	Stop                              func() bool
}

type espressoPrimeWork struct {
	options    EspressoPrimeOptions
	operations int
	generated  int
}

func newEspressoPrimeWork(options EspressoPrimeOptions) (*espressoPrimeWork, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxDepth < 0 {
		return nil, ErrDimension
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
	return &espressoPrimeWork{options: options}, nil
}

func (work *espressoPrimeWork) consume(count int) error {
	if count < 0 || count > work.options.MaxOperations-work.operations {
		return ErrLimit
	}
	work.operations += count
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *espressoPrimeWork) add(count int) error {
	if count < 0 || count > work.options.MaxCubes-work.generated {
		return ErrLimit
	}
	work.generated += count
	return nil
}

func (layout *EspressoCubeLayout) primeCubeEmpty(cube *EspressoSet) bool {
	for variable := range layout.partSizes {
		if !layout.variableIntersection(cube, cube, variable) {
			return true
		}
	}
	return false
}

func (layout *EspressoCubeLayout) primeFamily(cubes []*EspressoSet, contain bool, work *espressoPrimeWork) (*EspressoSetFamily, error) {
	if err := work.add(len(cubes)); err != nil {
		return nil, err
	}
	result, _ := NewEspressoSetFamily(len(cubes), layout.bitCount)
	for _, cube := range cubes {
		clone, _ := cube.Clone()
		result.sets = append(result.sets, clone)
		if clone.Flags&EspressoSetActive != 0 {
			result.activeCount++
		}
	}
	if !contain || len(cubes) < 2 {
		return result, nil
	}
	if len(cubes)-1 > work.options.MaxOperations/len(cubes) {
		return nil, ErrLimit
	}
	if err := work.consume(len(cubes) * (len(cubes) - 1) / 2); err != nil {
		return nil, err
	}
	return result.Contain(EspressoContainOptions{MaxSets: work.options.MaxCubes, MaxChecks: work.options.MaxOperations, Stop: work.options.Stop})
}

func (layout *EspressoCubeLayout) primeListClone(list *EspressoCubeList, cofactor *EspressoSet) *EspressoCubeList {
	result := &EspressoCubeList{layout: layout, cofactor: cofactor, cubes: make([]*EspressoSet, len(list.cubes))}
	for index, cube := range list.cubes {
		result.cubes[index], _ = cube.Clone()
	}
	return result
}

func (layout *EspressoCubeLayout) primeAndCofactor(family *EspressoSetFamily, cofactor *EspressoSet, work *espressoPrimeWork) (*EspressoSetFamily, error) {
	if err := work.consume(len(family.sets) * max(1, layout.bitCount)); err != nil {
		return nil, err
	}
	cubes := make([]*EspressoSet, 0, len(family.sets))
	for _, source := range family.sets {
		cube, _ := source.Intersection(cofactor)
		if layout.primeCubeEmpty(cube) {
			continue
		}
		cube.Flags |= EspressoSetActive
		cubes = append(cubes, cube)
	}
	return layout.primeFamily(cubes, false, work)
}

func (layout *EspressoCubeLayout) primeMerge(left, right *EspressoSetFamily, leftCofactor, rightCofactor *EspressoSet, work *espressoPrimeWork) (*EspressoSetFamily, error) {
	left, err := layout.primeAndCofactor(left, leftCofactor, work)
	if err != nil {
		return nil, err
	}
	right, err = layout.primeAndCofactor(right, rightCofactor, work)
	if err != nil {
		return nil, err
	}
	cubes := make([]*EspressoSet, 0, len(left.sets)+len(right.sets))
	cubes = append(cubes, left.sets...)
	cubes = append(cubes, right.sets...)
	for _, leftCube := range left.sets {
		for _, rightCube := range right.sets {
			if err = work.consume(layout.VariableCount()); err != nil {
				return nil, err
			}
			distance, _ := layout.DistanceAtMostOne(leftCube, rightCube)
			if distance == 1 {
				if err = work.consume(layout.VariableCount()); err != nil {
					return nil, err
				}
				consensus, consensusErr := layout.Consensus(leftCube, rightCube)
				if consensusErr != nil {
					return nil, consensusErr
				}
				cubes = append(cubes, consensus)
			}
		}
	}
	return layout.primeFamily(cubes, true, work)
}

func (layout *EspressoCubeLayout) primeConsensus(list *EspressoCubeList, depth int, work *espressoPrimeWork) (*EspressoSetFamily, error) {
	if depth > work.options.MaxDepth {
		return nil, ErrLimit
	}
	if err := work.consume(1); err != nil {
		return nil, err
	}
	if err := list.check(); err != nil {
		return nil, err
	}
	if len(list.cubes) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	if err := work.consume(2 * len(list.cubes) * max(1, len(list.cofactor.words))); err != nil {
		return nil, err
	}
	if len(list.cubes) == 0 {
		return layout.primeFamily(nil, false, work)
	}
	if len(list.cubes) == 1 {
		cube, _ := list.cubes[0].Union(list.cofactor)
		return layout.primeFamily([]*EspressoSet{cube}, false, work)
	}
	for _, cube := range list.cubes {
		full, err := layout.FullRow(cube, list.cofactor)
		if err != nil {
			return nil, err
		}
		if full {
			cube, _ := layout.FullCube()
			return layout.primeFamily([]*EspressoSet{cube}, false, work)
		}
	}
	ceiling, _ := list.cofactor.Clone()
	for _, cube := range list.cubes {
		for word := range ceiling.words {
			ceiling.words[word] |= cube.words[word]
		}
	}
	ceilingFull, _ := ceiling.Full()
	if !ceilingFull {
		full, _ := layout.FullCube()
		missing, _ := full.Difference(ceiling)
		factoredMask, _ := list.cofactor.Union(missing)
		factored := layout.primeListClone(list, factoredMask)
		result, err := layout.primeConsensus(factored, depth+1, work)
		if err != nil {
			return nil, err
		}
		for _, cube := range result.sets {
			for word := range cube.words {
				cube.words[word] &= ceiling.words[word]
			}
		}
		return result, nil
	}
	if err := work.consume(len(list.cubes) * max(1, layout.bitCount)); err != nil {
		return nil, err
	}
	counts, err := list.Analyze(EspressoCofactorOptions{MaxCubes: work.options.MaxCubes, MaxChecks: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return nil, err
	}
	if counts.VariablesActive == 1 {
		cube, _ := layout.FullCube()
		return layout.primeFamily([]*EspressoSet{cube}, false, work)
	}
	if counts.VariablesUnate == counts.VariablesActive {
		family, unlistErr := list.Unlist(EspressoCofactorOptions{MaxCubes: work.options.MaxCubes, Stop: work.options.Stop})
		if unlistErr != nil {
			return nil, unlistErr
		}
		return layout.primeFamily(family.sets, true, work)
	}
	variable, leftCofactor, rightCofactor, err := list.BinateSplit(counts)
	if err != nil {
		return nil, err
	}
	if err = work.consume(max(1, layout.bitCount)); err != nil {
		return nil, err
	}
	if err = work.consume(2 * len(list.cubes) * max(1, layout.VariableCount())); err != nil {
		return nil, err
	}
	cofactorOptions := EspressoCofactorOptions{MaxCubes: work.options.MaxCubes, MaxChecks: work.options.MaxOperations, Stop: work.options.Stop}
	leftList, err := list.SingleVariableCofactor(leftCofactor, variable, -1, cofactorOptions)
	if err != nil {
		return nil, err
	}
	rightList, err := list.SingleVariableCofactor(rightCofactor, variable, -1, cofactorOptions)
	if err != nil {
		return nil, err
	}
	left, err := layout.primeConsensus(leftList, depth+1, work)
	if err != nil {
		return nil, err
	}
	right, err := layout.primeConsensus(rightList, depth+1, work)
	if err != nil {
		return nil, err
	}
	return layout.primeMerge(left, right, leftCofactor, rightCofactor, work)
}

// PrimeConsensus generates the prime implicants of the cube-list function.
func (layout *EspressoCubeLayout) PrimeConsensus(list *EspressoCubeList, options EspressoPrimeOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoPrimeWork(options)
	if err != nil {
		return nil, err
	}
	if err = layout.check(); err != nil {
		return nil, err
	}
	if err = list.check(); err != nil {
		return nil, err
	}
	if list.layout.bitCount != layout.bitCount || list.layout.binaryVariables != layout.binaryVariables || len(list.layout.partSizes) != len(layout.partSizes) {
		return nil, ErrDimension
	}
	for variable := range layout.partSizes {
		if list.layout.partSizes[variable] != layout.partSizes[variable] {
			return nil, ErrDimension
		}
	}
	cofactor, _ := list.cofactor.Clone()
	return layout.primeConsensus(layout.primeListClone(list, cofactor), 0, work)
}
