package abc

// Owned cube lists, cofactoring and split analysis from src/misc/espresso/cofactor.c.

type EspressoCofactorOptions struct {
	MaxCubes, MaxChecks int
	Stop                func() bool
}

func normalizeEspressoCofactorOptions(options EspressoCofactorOptions) (EspressoCofactorOptions, error) {
	if options.MaxCubes < 0 || options.MaxChecks < 0 {
		return options, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxChecks == 0 {
		options.MaxChecks = 16 << 20
	}
	return options, nil
}

type EspressoCubeList struct {
	layout   *EspressoCubeLayout
	cofactor *EspressoSet
	cubes    []*EspressoSet
}

func NewEspressoCubeList(layout *EspressoCubeLayout, families ...*EspressoSetFamily) (*EspressoCubeList, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	if len(families) < 1 || len(families) > 3 {
		return nil, ErrDimension
	}
	total := 0
	for _, family := range families {
		if err := family.check(); err != nil {
			return nil, err
		}
		if family.bits != layout.bitCount {
			return nil, ErrDimension
		}
		if len(family.sets) > (1<<20)-total {
			return nil, ErrLimit
		}
		total += len(family.sets)
	}
	empty, _ := NewEspressoSet(layout.bitCount)
	result := &EspressoCubeList{layout: layout, cofactor: empty, cubes: make([]*EspressoSet, 0, total)}
	for _, family := range families {
		for _, cube := range family.sets {
			clone, _ := cube.Clone()
			result.cubes = append(result.cubes, clone)
		}
	}
	return result, nil
}

func (list *EspressoCubeList) check() error {
	if list == nil || list.layout == nil {
		return ErrDimension
	}
	if err := list.layout.check(); err != nil {
		return err
	}
	if err := list.cofactor.check(); err != nil || list.cofactor.bits != list.layout.bitCount {
		return ErrInvariant
	}
	for _, cube := range list.cubes {
		if err := cube.check(); err != nil || cube.bits != list.layout.bitCount {
			return ErrInvariant
		}
	}
	return nil
}

func (list *EspressoCubeList) Count() (int, error) {
	if err := list.check(); err != nil {
		return 0, err
	}
	return len(list.cubes), nil
}

func (list *EspressoCubeList) CofactorMask() (*EspressoSet, error) {
	if err := list.check(); err != nil {
		return nil, err
	}
	return list.cofactor.Clone()
}

func (list *EspressoCubeList) Cube(index int) (*EspressoSet, error) {
	if err := list.check(); err != nil {
		return nil, err
	}
	if index < 0 || index >= len(list.cubes) {
		return nil, ErrDimension
	}
	return list.cubes[index].Clone()
}

func (list *EspressoCubeList) cofactorWith(cube *EspressoSet, variable, exclude int, specialized bool, options EspressoCofactorOptions) (*EspressoCubeList, error) {
	options, err := normalizeEspressoCofactorOptions(options)
	if err != nil {
		return nil, err
	}
	if err = list.check(); err != nil {
		return nil, err
	}
	if err = list.layout.checkSet(cube); err != nil {
		return nil, err
	}
	if len(list.cubes) > options.MaxCubes || exclude < -1 || exclude >= len(list.cubes) {
		if exclude < -1 || exclude >= len(list.cubes) {
			return nil, ErrDimension
		}
		return nil, ErrLimit
	}
	if specialized {
		if variable < 0 || variable >= list.layout.VariableCount() {
			return nil, ErrDimension
		}
		for other := range list.layout.partSizes {
			if other != variable && !list.layout.variableFull(cube, other) {
				return nil, ErrDimension
			}
		}
	}
	full, _ := list.layout.FullCube()
	complement, _ := full.Difference(cube)
	newMask, _ := list.cofactor.Union(complement)
	result := &EspressoCubeList{layout: list.layout, cofactor: newMask, cubes: make([]*EspressoSet, 0, len(list.cubes))}
	checks := 0
	for index, candidate := range list.cubes {
		if index == exclude {
			continue
		}
		keep := false
		if specialized {
			keep = true
			for current := range list.layout.partSizes {
				checks++
				if checks > options.MaxChecks {
					return nil, ErrLimit
				}
				if options.Stop != nil && options.Stop() {
					return nil, ErrCancelled
				}
				// The C fast path assumes stored cubes are nonempty in every
				// other variable. Preserve cofactor semantics for malformed
				// products instead of silently retaining an empty cube.
				if !list.layout.variableIntersection(candidate, cube, current) {
					keep = false
					break
				}
			}
		} else {
			keep = true
			for current := range list.layout.partSizes {
				checks++
				if checks > options.MaxChecks {
					return nil, ErrLimit
				}
				if options.Stop != nil && options.Stop() {
					return nil, ErrCancelled
				}
				if !list.layout.variableIntersection(candidate, cube, current) {
					keep = false
					break
				}
			}
		}
		if keep {
			clone, _ := candidate.Clone()
			result.cubes = append(result.cubes, clone)
		}
	}
	return result, nil
}

// Cofactor applies an external cube. exclude is -1 or the source-list index
// whose pointer identity should be omitted, matching cofactor's p != c rule.
func (list *EspressoCubeList) Cofactor(cube *EspressoSet, exclude int, options EspressoCofactorOptions) (*EspressoCubeList, error) {
	return list.cofactorWith(cube, -1, exclude, false, options)
}

func (list *EspressoCubeList) CofactorAt(index int, options EspressoCofactorOptions) (*EspressoCubeList, error) {
	if err := list.check(); err != nil {
		return nil, err
	}
	if index < 0 || index >= len(list.cubes) {
		return nil, ErrDimension
	}
	return list.cofactorWith(list.cubes[index], -1, index, false, options)
}

// SingleVariableCofactor is scofactor and requires every other variable of the
// cofactor cube to be full, making the source fast-path assumption explicit.
func (list *EspressoCubeList) SingleVariableCofactor(cube *EspressoSet, variable, exclude int, options EspressoCofactorOptions) (*EspressoCubeList, error) {
	return list.cofactorWith(cube, variable, exclude, true, options)
}

func (list *EspressoCubeList) Unlist(options EspressoCofactorOptions) (*EspressoSetFamily, error) {
	options, err := normalizeEspressoCofactorOptions(options)
	if err != nil {
		return nil, err
	}
	if err = list.check(); err != nil {
		return nil, err
	}
	if len(list.cubes) > options.MaxCubes {
		return nil, ErrLimit
	}
	result, _ := NewEspressoSetFamily(len(list.cubes), list.layout.bitCount)
	for _, cube := range list.cubes {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		unlisted, _ := cube.Union(list.cofactor)
		result.sets = append(result.sets, unlisted)
	}
	return result, nil
}

type EspressoCubeCounts struct {
	PartZeros, VariableZeros, PartsActive []int
	IsUnate                               []bool
	VariablesActive, VariablesUnate, Best int
}

func (list *EspressoCubeList) Analyze(options EspressoCofactorOptions) (EspressoCubeCounts, error) {
	options, err := normalizeEspressoCofactorOptions(options)
	if err != nil {
		return EspressoCubeCounts{}, err
	}
	if err = list.check(); err != nil {
		return EspressoCubeCounts{}, err
	}
	if len(list.cubes) > options.MaxCubes {
		return EspressoCubeCounts{}, ErrLimit
	}
	result := EspressoCubeCounts{
		PartZeros: make([]int, list.layout.bitCount), VariableZeros: make([]int, list.layout.VariableCount()),
		PartsActive: make([]int, list.layout.VariableCount()), IsUnate: make([]bool, list.layout.VariableCount()), Best: -1,
	}
	checks := 0
	for _, cube := range list.cubes {
		for part := 0; part < list.layout.bitCount; part++ {
			checks++
			if checks > options.MaxChecks {
				return EspressoCubeCounts{}, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return EspressoCubeCounts{}, ErrCancelled
			}
			cubePresent, _ := cube.Has(part)
			cofactored, _ := list.cofactor.Has(part)
			if !cubePresent && !cofactored {
				result.PartZeros[part]++
			}
		}
	}
	mostActive, mostZero, mostBalanced := 0, 0, 32000
	for variable, size := range list.layout.partSizes {
		start := list.layout.firstParts[variable]
		active, maxActive := 0, 0
		for part := 0; part < size; part++ {
			zeros := result.PartZeros[start+part]
			result.VariableZeros[variable] += zeros
			if zeros > 0 {
				active++
			}
			if variable < list.layout.binaryVariables && zeros > maxActive {
				maxActive = zeros
			}
		}
		if variable >= list.layout.binaryVariables {
			// Preserve massive_count's executable MV balancing value.
			maxActive = active
		}
		if active > mostActive || active == mostActive && (result.VariableZeros[variable] > mostZero || result.VariableZeros[variable] == mostZero && maxActive < mostBalanced) {
			result.Best, mostActive, mostZero, mostBalanced = variable, active, result.VariableZeros[variable], maxActive
		}
		result.PartsActive[variable] = active
		result.IsUnate[variable] = active == 1
		if active > 0 {
			result.VariablesActive++
		}
		if active == 1 {
			result.VariablesUnate++
		}
	}
	return result, nil
}

func (list *EspressoCubeList) BinateSplit(counts EspressoCubeCounts) (int, *EspressoSet, *EspressoSet, error) {
	if err := list.check(); err != nil {
		return -1, nil, nil, err
	}
	if len(counts.PartZeros) != list.layout.bitCount || len(counts.VariableZeros) != list.layout.VariableCount() || len(counts.PartsActive) != list.layout.VariableCount() || len(counts.IsUnate) != list.layout.VariableCount() || counts.Best < 0 || counts.Best >= list.layout.VariableCount() {
		return -1, nil, nil, ErrDimension
	}
	variable := counts.Best
	left, _ := list.layout.FullCube()
	right, _ := list.layout.FullCube()
	start, size := list.layout.firstParts[variable], list.layout.partSizes[variable]
	for part := 0; part < size; part++ {
		_ = left.Set(start+part, false)
		_ = right.Set(start+part, false)
	}
	uncofactored := make([]int, 0, size)
	for part := 0; part < size; part++ {
		masked, _ := list.cofactor.Has(start + part)
		if !masked {
			uncofactored = append(uncofactored, start+part)
		}
	}
	half := len(uncofactored) / 2
	for index, part := range uncofactored {
		if index < half {
			_ = left.Set(part, true)
		} else {
			_ = right.Set(part, true)
		}
	}
	return variable, left, right, nil
}

// Simplify sorts by cube|cofactor and removes masked duplicates. It repairs
// simplify_cubelist's dormant off-by-two loop, which otherwise drops entries.
func (list *EspressoCubeList) Simplify(options EspressoCofactorOptions) (*EspressoCubeList, error) {
	options, err := normalizeEspressoCofactorOptions(options)
	if err != nil {
		return nil, err
	}
	if err = list.check(); err != nil {
		return nil, err
	}
	if len(list.cubes) > options.MaxCubes {
		return nil, ErrLimit
	}
	result := &EspressoCubeList{layout: list.layout, cubes: make([]*EspressoSet, len(list.cubes))}
	result.cofactor, _ = list.cofactor.Clone()
	for index, cube := range list.cubes {
		result.cubes[index], _ = cube.Clone()
	}
	checks := 0
	buffer := make([]*EspressoSet, len(result.cubes))
	for width := 1; width < len(result.cubes); {
		for first := 0; first < len(result.cubes); first += 2 * width {
			middle, last := min(first+width, len(result.cubes)), min(first+2*width, len(result.cubes))
			left, right, destination := first, middle, first
			for left < middle && right < last {
				checks++
				if checks > options.MaxChecks {
					return nil, ErrLimit
				}
				if options.Stop != nil && options.Stop() {
					return nil, ErrCancelled
				}
				comparison, _ := EspressoSetCompareDistanceOne(result.cubes[left], result.cubes[right], result.cofactor)
				if comparison <= 0 {
					buffer[destination], left = result.cubes[left], left+1
				} else {
					buffer[destination], right = result.cubes[right], right+1
				}
				destination++
			}
			for left < middle {
				buffer[destination], left, destination = result.cubes[left], left+1, destination+1
			}
			for right < last {
				buffer[destination], right, destination = result.cubes[right], right+1, destination+1
			}
		}
		result.cubes, buffer = buffer, result.cubes
		if width > len(result.cubes)/2 {
			break
		}
		width *= 2
	}
	unique := result.cubes[:0]
	for _, cube := range result.cubes {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if len(unique) != 0 {
			checks++
			if checks > options.MaxChecks {
				return nil, ErrLimit
			}
			comparison, _ := EspressoSetCompareDistanceOne(unique[len(unique)-1], cube, result.cofactor)
			if comparison == 0 {
				continue
			}
		}
		unique = append(unique, cube)
	}
	result.cubes = unique
	return result, nil
}
