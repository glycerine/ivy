package abc

// Essential-prime extraction and sharp consensus from src/misc/espresso/essen.c.

type EspressoEssentialOptions struct {
	MaxCubes, MaxOperations int
	Stop                    func() bool
}

type espressoEssentialWork struct {
	options    EspressoEssentialOptions
	operations int
	generated  int
}

func newEspressoEssentialWork(options EspressoEssentialOptions) (*espressoEssentialWork, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 16 << 20
	}
	return &espressoEssentialWork{options: options}, nil
}

func (work *espressoEssentialWork) consume(count int) error {
	if count < 0 || count > work.options.MaxOperations-work.operations {
		return ErrLimit
	}
	work.operations += count
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *espressoEssentialWork) add(count int) error {
	if count < 0 || count > work.options.MaxCubes-work.generated {
		return ErrLimit
	}
	work.generated += count
	return nil
}

func (layout *EspressoCubeLayout) essentialFamily(cubes []*EspressoSet, work *espressoEssentialWork) (*EspressoSetFamily, error) {
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
	return result, nil
}

func (layout *EspressoCubeLayout) consensusSharpDistanceZero(first, second *EspressoSet, work *espressoEssentialWork) ([]*EspressoSet, error) {
	if err := layout.checkSet(first); err != nil {
		return nil, err
	}
	if err := layout.checkSet(second); err != nil {
		return nil, err
	}
	implies, _ := first.Implies(second)
	if implies {
		return nil, nil
	}
	difference, _ := first.Difference(second)
	intersection, _ := first.Intersection(second)
	cubes := make([]*EspressoSet, 0)
	for variable := layout.binaryVariables; variable < layout.VariableCount(); variable++ {
		if err := work.consume(layout.partSizes[variable]); err != nil {
			return nil, err
		}
		start, size := layout.firstParts[variable], layout.partSizes[variable]
		hasDifference := false
		for part := start; part < start+size; part++ {
			if difference.words[part>>5]&(uint32(1)<<uint(part&31)) != 0 {
				hasDifference = true
				break
			}
		}
		if hasDifference {
			mask, _ := layout.VariableMask(variable)
			cube, _ := second.Merge(intersection, mask)
			cubes = append(cubes, cube)
		}
	}
	if len(cubes) == 0 && layout.binaryVariables > 0 {
		cubes = append(cubes, intersection)
	}
	return cubes, nil
}

func (layout *EspressoCubeLayout) ConsensusSharpDistanceZero(first, second *EspressoSet, options EspressoEssentialOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoEssentialWork(options)
	if err != nil {
		return nil, err
	}
	cubes, err := layout.consensusSharpDistanceZero(first, second, work)
	if err != nil {
		return nil, err
	}
	return layout.essentialFamily(cubes, work)
}

func (layout *EspressoCubeLayout) consensusSharp(cover *EspressoSetFamily, cube *EspressoSet, exclude int, work *espressoEssentialWork) (*EspressoSetFamily, error) {
	if err := layout.checkSet(cube); err != nil {
		return nil, err
	}
	if err := cover.check(); err != nil {
		return nil, err
	}
	if cover.bits != layout.bitCount || exclude < -1 || exclude >= len(cover.sets) {
		return nil, ErrDimension
	}
	if len(cover.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	result := make([]*EspressoSet, 0, 2*len(cover.sets))
	for index, candidate := range cover.sets {
		if index == exclude {
			continue
		}
		if err := work.consume(max(1, layout.bitCount)); err != nil {
			return nil, err
		}
		distance, _ := layout.DistanceAtMostOne(candidate, cube)
		switch distance {
		case 0:
			pieces, err := layout.consensusSharpDistanceZero(candidate, cube, work)
			if err != nil {
				return nil, err
			}
			result = append(result, pieces...)
		case 1:
			consensus, err := layout.Consensus(candidate, cube)
			if err != nil {
				return nil, err
			}
			result = append(result, consensus)
		}
	}
	return layout.essentialFamily(result, work)
}

// ConsensusSharp computes consensus(cover # cube, cube). exclude is -1 or a
// source row index to omit explicitly instead of relying on pointer identity.
func (layout *EspressoCubeLayout) ConsensusSharp(cover *EspressoSetFamily, cube *EspressoSet, exclude int, options EspressoEssentialOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoEssentialWork(options)
	if err != nil {
		return nil, err
	}
	return layout.consensusSharp(cover, cube, exclude, work)
}

func (layout *EspressoCubeLayout) cubeCoveredByCover(cube *EspressoSet, cover *EspressoSetFamily, work *espressoEssentialWork) (bool, error) {
	if err := work.consume(1); err != nil {
		return false, err
	}
	remainingOperations := work.options.MaxOperations - work.operations
	remainingCubes := work.options.MaxCubes - work.generated
	if remainingOperations <= 0 || remainingCubes <= 0 {
		return false, ErrLimit
	}
	sharpWork, err := newEspressoSharpWork(EspressoSharpOptions{MaxCubes: remainingCubes, MaxOperations: remainingOperations, Stop: work.options.Stop})
	if err != nil {
		return false, err
	}
	remainder, sharpErr := layout.sharpCubeByCover(cube, cover, sharpWork)
	work.operations += sharpWork.operations
	work.generated += sharpWork.generated
	if sharpErr != nil {
		return false, sharpErr
	}
	return len(remainder.sets) == 0, nil
}

func (layout *EspressoCubeLayout) essentialCube(onset, dontCare *EspressoSetFamily, cube *EspressoSet, work *espressoEssentialWork) (bool, error) {
	if err := onset.check(); err != nil {
		return false, err
	}
	if err := dontCare.check(); err != nil {
		return false, err
	}
	if err := layout.checkSet(cube); err != nil {
		return false, err
	}
	if onset.bits != layout.bitCount || dontCare.bits != layout.bitCount {
		return false, ErrDimension
	}
	if len(onset.sets)+len(dontCare.sets) > work.options.MaxCubes {
		return false, ErrLimit
	}
	joined, err := JoinEspressoSetFamilies(onset, dontCare)
	if err != nil {
		return false, err
	}
	consensus, err := layout.consensusSharp(joined, cube, -1, work)
	if err != nil {
		return false, err
	}
	coverage, err := JoinEspressoSetFamilies(consensus, dontCare)
	if err != nil {
		return false, err
	}
	covered, err := layout.cubeCoveredByCover(cube, coverage, work)
	return !covered, err
}

func (layout *EspressoCubeLayout) EssentialCube(onset, dontCare *EspressoSetFamily, cube *EspressoSet, options EspressoEssentialOptions) (bool, error) {
	work, err := newEspressoEssentialWork(options)
	if err != nil {
		return false, err
	}
	return layout.essentialCube(onset, dontCare, cube, work)
}

type EspressoEssentialResult struct {
	Essential, Remaining, DontCare *EspressoSetFamily
}

// ExtractEssentials returns owned replacements for essential's three mutated
// covers: selected essentials, remaining onset and augmented don't-care set.
func (layout *EspressoCubeLayout) ExtractEssentials(onset, dontCare *EspressoSetFamily, options EspressoEssentialOptions) (EspressoEssentialResult, error) {
	work, err := newEspressoEssentialWork(options)
	if err != nil {
		return EspressoEssentialResult{}, err
	}
	if err = layout.check(); err != nil {
		return EspressoEssentialResult{}, err
	}
	if err = onset.check(); err != nil {
		return EspressoEssentialResult{}, err
	}
	if err = dontCare.check(); err != nil {
		return EspressoEssentialResult{}, err
	}
	if onset.bits != layout.bitCount || dontCare.bits != layout.bitCount {
		return EspressoEssentialResult{}, ErrDimension
	}
	if len(onset.sets)+len(dontCare.sets) > work.options.MaxCubes {
		return EspressoEssentialResult{}, ErrLimit
	}
	remaining, _ := onset.Clone()
	_ = remaining.ActivateAll()
	essential, _ := NewEspressoSetFamily(0, layout.bitCount)
	for index, cube := range remaining.sets {
		if cube.Flags&EspressoSetNonessential != 0 || cube.Flags&EspressoSetRelEssential == 0 {
			continue
		}
		isEssential, essentialErr := layout.essentialCube(remaining, dontCare, cube, work)
		if essentialErr != nil {
			return EspressoEssentialResult{}, essentialErr
		}
		if !isEssential {
			continue
		}
		if err = work.add(1); err != nil {
			return EspressoEssentialResult{}, err
		}
		clone, _ := cube.Clone()
		essential.sets = append(essential.sets, clone)
		if clone.Flags&EspressoSetActive != 0 {
			essential.activeCount++
		}
		_ = remaining.SetActive(index, false)
	}
	_ = remaining.CompactActive()
	augmented, err := JoinEspressoSetFamilies(dontCare, essential)
	if err != nil {
		return EspressoEssentialResult{}, err
	}
	return EspressoEssentialResult{Essential: essential, Remaining: remaining, DontCare: augmented}, nil
}
