package abc

// Cover sharp, disjoint sharp and intersection from src/misc/espresso/sharp.c.

type EspressoSharpOptions struct {
	MaxCubes, MaxOperations int
	Stop                    func() bool
}

type espressoSharpWork struct {
	options    EspressoSharpOptions
	operations int
	generated  int
}

func newEspressoSharpWork(options EspressoSharpOptions) (*espressoSharpWork, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 16 << 20
	}
	return &espressoSharpWork{options: options}, nil
}

func (work *espressoSharpWork) step() error {
	work.operations++
	if work.operations > work.options.MaxOperations {
		return ErrLimit
	}
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *espressoSharpWork) addCube() error {
	work.generated++
	if work.generated > work.options.MaxCubes {
		return ErrLimit
	}
	return nil
}

func (layout *EspressoCubeLayout) cubeEmpty(cube *EspressoSet, work *espressoSharpWork) (bool, error) {
	for variable, size := range layout.partSizes {
		present := false
		start := layout.firstParts[variable]
		for part := start; part < start+size; part++ {
			if err := work.step(); err != nil {
				return false, err
			}
			if cube.words[part>>5]&(uint32(1)<<uint(part&31)) != 0 {
				present = true
				break
			}
		}
		if !present {
			return true, nil
		}
	}
	return false, nil
}

func (layout *EspressoCubeLayout) sharpFamily(cubes []*EspressoSet, contain bool, work *espressoSharpWork) (*EspressoSetFamily, error) {
	result, err := NewEspressoSetFamily(len(cubes), layout.bitCount)
	if err != nil {
		return nil, err
	}
	for _, cube := range cubes {
		if err = work.step(); err != nil {
			return nil, err
		}
		if err = work.addCube(); err != nil {
			return nil, err
		}
		clone, _ := cube.Clone()
		result.sets = append(result.sets, clone)
		if clone.Flags&EspressoSetActive != 0 {
			result.activeCount++
		}
	}
	if !contain || len(result.sets) < 2 {
		return result, nil
	}
	contained, err := result.Contain(EspressoContainOptions{MaxSets: work.options.MaxCubes, MaxChecks: work.options.MaxOperations, Stop: work.options.Stop})
	if err != nil {
		return nil, err
	}
	return contained, nil
}

func (layout *EspressoCubeLayout) sharpCubes(first, second *EspressoSet, disjoint bool, work *espressoSharpWork) ([]*EspressoSet, error) {
	if err := layout.checkSet(first); err != nil {
		return nil, err
	}
	if err := layout.checkSet(second); err != nil {
		return nil, err
	}
	firstEmpty, err := layout.cubeEmpty(first, work)
	if err != nil {
		return nil, err
	}
	if firstEmpty {
		return nil, nil
	}
	secondEmpty, err := layout.cubeEmpty(second, work)
	if err != nil {
		return nil, err
	}
	if secondEmpty {
		clone, _ := first.Clone()
		return []*EspressoSet{clone}, nil
	}
	intersects, _ := layout.DistanceZero(first, second)
	if !intersects {
		clone, _ := first.Clone()
		return []*EspressoSet{clone}, nil
	}
	difference, _ := first.Difference(second)
	intersection, _ := first.Intersection(second)
	result := make([]*EspressoSet, 0, layout.VariableCount())
	priorDifference, _ := NewEspressoSet(layout.bitCount)
	for variable, size := range layout.partSizes {
		start := layout.firstParts[variable]
		hasDifference := false
		for part := start; part < start+size; part++ {
			if err := work.step(); err != nil {
				return nil, err
			}
			present, _ := difference.Has(part)
			if present {
				hasDifference = true
				break
			}
		}
		if !hasDifference {
			continue
		}
		cube, _ := NewEspressoSet(layout.bitCount)
		for part := start; part < start+size; part++ {
			if err := work.step(); err != nil {
				return nil, err
			}
			present, _ := difference.Has(part)
			if present {
				_ = cube.Set(part, true)
			}
		}
		if disjoint {
			for part := 0; part < layout.bitCount; part++ {
				if err := work.step(); err != nil {
					return nil, err
				}
				prior, _ := priorDifference.Has(part)
				if prior {
					common, _ := intersection.Has(part)
					if common {
						_ = cube.Set(part, true)
					}
				}
			}
		}
		for other, otherSize := range layout.partSizes {
			otherStart := layout.firstParts[other]
			if other == variable {
				continue
			}
			prior := false
			if disjoint {
				for part := otherStart; part < otherStart+otherSize; part++ {
					if err := work.step(); err != nil {
						return nil, err
					}
					priorPart, _ := priorDifference.Has(part)
					if priorPart {
						prior = true
						break
					}
				}
			}
			if disjoint && prior {
				continue
			}
			for part := otherStart; part < otherStart+otherSize; part++ {
				if err := work.step(); err != nil {
					return nil, err
				}
				present, _ := first.Has(part)
				if present {
					_ = cube.Set(part, true)
				}
			}
		}
		result = append(result, cube)
		if disjoint {
			for part := start; part < start+size; part++ {
				if err := work.step(); err != nil {
					return nil, err
				}
				_ = priorDifference.Set(part, true)
			}
		}
	}
	return result, nil
}

func (layout *EspressoCubeLayout) SharpCubes(first, second *EspressoSet, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	cubes, err := layout.sharpCubes(first, second, false, work)
	if err != nil {
		return nil, err
	}
	return layout.sharpFamily(cubes, false, work)
}

func (layout *EspressoCubeLayout) DisjointSharpCubes(first, second *EspressoSet, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	cubes, err := layout.sharpCubes(first, second, true, work)
	if err != nil {
		return nil, err
	}
	return layout.sharpFamily(cubes, false, work)
}

func (layout *EspressoCubeLayout) intersectFamilies(first, second *EspressoSetFamily, work *espressoSharpWork) (*EspressoSetFamily, error) {
	if err := first.check(); err != nil {
		return nil, err
	}
	if err := second.check(); err != nil {
		return nil, err
	}
	if first.bits != layout.bitCount || second.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(first.sets) > work.options.MaxCubes || len(second.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	cubes := make([]*EspressoSet, 0)
	for _, left := range first.sets {
		leftEmpty, err := layout.cubeEmpty(left, work)
		if err != nil {
			return nil, err
		}
		if leftEmpty {
			continue
		}
		for _, right := range second.sets {
			rightEmpty, err := layout.cubeEmpty(right, work)
			if err != nil {
				return nil, err
			}
			if rightEmpty {
				continue
			}
			intersects, _ := layout.DistanceZero(left, right)
			if intersects {
				cube, _ := left.Intersection(right)
				cubes = append(cubes, cube)
			}
		}
	}
	return layout.sharpFamily(cubes, true, work)
}

func (layout *EspressoCubeLayout) IntersectCovers(first, second *EspressoSetFamily, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	if err = layout.check(); err != nil {
		return nil, err
	}
	return layout.intersectFamilies(first, second, work)
}

func (layout *EspressoCubeLayout) sharpCubeRange(cube *EspressoSet, cover *EspressoSetFamily, first, last int, work *espressoSharpWork) (*EspressoSetFamily, error) {
	if first == last {
		pieces, err := layout.sharpCubes(cube, cover.sets[first], false, work)
		if err != nil {
			return nil, err
		}
		return layout.sharpFamily(pieces, false, work)
	}
	middle := (first + last) / 2
	left, err := layout.sharpCubeRange(cube, cover, first, middle, work)
	if err != nil {
		return nil, err
	}
	right, err := layout.sharpCubeRange(cube, cover, middle+1, last, work)
	if err != nil {
		return nil, err
	}
	return layout.intersectFamilies(left, right, work)
}

func (layout *EspressoCubeLayout) sharpCubeByCover(cube *EspressoSet, cover *EspressoSetFamily, work *espressoSharpWork) (*EspressoSetFamily, error) {
	if err := layout.checkSet(cube); err != nil {
		return nil, err
	}
	if err := cover.check(); err != nil {
		return nil, err
	}
	if cover.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(cover.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	empty, err := layout.cubeEmpty(cube, work)
	if err != nil {
		return nil, err
	}
	if empty {
		return NewEspressoSetFamily(0, layout.bitCount)
	}
	if len(cover.sets) == 0 {
		return layout.sharpFamily([]*EspressoSet{cube}, false, work)
	}
	return layout.sharpCubeRange(cube, cover, 0, len(cover.sets)-1, work)
}

func (layout *EspressoCubeLayout) SharpCubeByCover(cube *EspressoSet, cover *EspressoSetFamily, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	return layout.sharpCubeByCover(cube, cover, work)
}

func (layout *EspressoCubeLayout) SharpCovers(first, second *EspressoSetFamily, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	if err = layout.check(); err != nil {
		return nil, err
	}
	if err = first.check(); err != nil {
		return nil, err
	}
	if err = second.check(); err != nil {
		return nil, err
	}
	if first.bits != layout.bitCount || second.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(first.sets) > work.options.MaxCubes || len(second.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	cubes := make([]*EspressoSet, 0)
	for _, cube := range first.sets {
		pieces, pieceErr := layout.sharpCubeByCover(cube, second, work)
		if pieceErr != nil {
			return nil, pieceErr
		}
		cubes = append(cubes, pieces.sets...)
	}
	return layout.sharpFamily(cubes, true, work)
}

func (layout *EspressoCubeLayout) disjointSharpFamilyByCube(family *EspressoSetFamily, cube *EspressoSet, work *espressoSharpWork, contain bool) (*EspressoSetFamily, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	if err := layout.checkSet(cube); err != nil {
		return nil, err
	}
	if family.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(family.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	pieces := make([]*EspressoSet, 0)
	for _, candidate := range family.sets {
		sharp, err := layout.sharpCubes(candidate, cube, true, work)
		if err != nil {
			return nil, err
		}
		pieces = append(pieces, sharp...)
	}
	return layout.sharpFamily(pieces, contain, work)
}

func (layout *EspressoCubeLayout) DisjointSharpFamilyByCube(family *EspressoSetFamily, cube *EspressoSet, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	return layout.disjointSharpFamilyByCube(family, cube, work, true)
}

func (layout *EspressoCubeLayout) disjointSharpCubeByCover(cube *EspressoSet, cover *EspressoSetFamily, work *espressoSharpWork) (*EspressoSetFamily, error) {
	if err := layout.checkSet(cube); err != nil {
		return nil, err
	}
	if err := cover.check(); err != nil {
		return nil, err
	}
	if cover.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(cover.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	current, err := layout.sharpFamily([]*EspressoSet{cube}, false, work)
	if err != nil {
		return nil, err
	}
	for _, remove := range cover.sets {
		current, err = layout.disjointSharpFamilyByCube(current, remove, work, false)
		if err != nil {
			return nil, err
		}
		if len(current.sets) == 0 {
			break
		}
	}
	return current, nil
}

func (layout *EspressoCubeLayout) DisjointSharpCubeByCover(cube *EspressoSet, cover *EspressoSetFamily, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	return layout.disjointSharpCubeByCover(cube, cover, work)
}

func (layout *EspressoCubeLayout) DisjointSharpCovers(first, second *EspressoSetFamily, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	if err = layout.check(); err != nil {
		return nil, err
	}
	if err = first.check(); err != nil {
		return nil, err
	}
	if err = second.check(); err != nil {
		return nil, err
	}
	if first.bits != layout.bitCount || second.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(first.sets) > work.options.MaxCubes || len(second.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	cubes := make([]*EspressoSet, 0)
	for _, cube := range first.sets {
		pieces, pieceErr := layout.disjointSharpCubeByCover(cube, second, work)
		if pieceErr != nil {
			return nil, pieceErr
		}
		cubes = append(cubes, pieces.sets...)
	}
	return layout.sharpFamily(cubes, true, work)
}

func (layout *EspressoCubeLayout) MakeDisjoint(cover *EspressoSetFamily, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	if err = layout.check(); err != nil {
		return nil, err
	}
	if err = cover.check(); err != nil {
		return nil, err
	}
	if cover.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(cover.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	result, _ := NewEspressoSetFamily(0, layout.bitCount)
	for _, cube := range cover.sets {
		pieces, pieceErr := layout.disjointSharpCubeByCover(cube, result, work)
		if pieceErr != nil {
			return nil, pieceErr
		}
		for _, piece := range pieces.sets {
			clone, _ := piece.Clone()
			result.sets = append(result.sets, clone)
			if clone.Flags&EspressoSetActive != 0 {
				result.activeCount++
			}
		}
	}
	return result, nil
}
