package abc

// Cover reshaping, ordering and output cofactors from src/misc/espresso/cvrm.c.

type EspressoCoverManipulationOptions struct {
	MaxCubes, MaxOperations int
	Stop                    func() bool
}

func normalizeEspressoCoverManipulationOptions(options EspressoCoverManipulationOptions) (EspressoCoverManipulationOptions, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 {
		return options, ErrDimension
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 16 << 20
	}
	return options, nil
}

type espressoCoverManipulationWork struct {
	options    EspressoCoverManipulationOptions
	operations int
	generated  int
}

func newEspressoCoverManipulationWork(options EspressoCoverManipulationOptions) (*espressoCoverManipulationWork, error) {
	options, err := normalizeEspressoCoverManipulationOptions(options)
	if err != nil {
		return nil, err
	}
	return &espressoCoverManipulationWork{options: options}, nil
}

func (work *espressoCoverManipulationWork) step(count int) error {
	if count < 0 || count > work.options.MaxOperations-work.operations {
		return ErrLimit
	}
	work.operations += count
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (work *espressoCoverManipulationWork) add(count int) error {
	if count < 0 || count > work.options.MaxCubes-work.generated {
		return ErrLimit
	}
	work.generated += count
	return nil
}

// UnravelRange expands every coordinate choice in the inclusive variable
// range into its Cartesian product while retaining untouched variables.
func (layout *EspressoCubeLayout) UnravelRange(family *EspressoSetFamily, first, last int, options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	if err := family.check(); err != nil {
		return nil, err
	}
	if family.bits != layout.bitCount || first < 0 || last < first || last >= layout.VariableCount() {
		return nil, ErrDimension
	}
	work, err := newEspressoCoverManipulationWork(options)
	if err != nil {
		return nil, err
	}
	if len(family.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	result, _ := NewEspressoSetFamily(len(family.sets), layout.bitCount)
	for _, source := range family.sets {
		rows := []*EspressoSet{source}
		for variable := first; variable <= last; variable++ {
			start, size := layout.firstParts[variable], layout.partSizes[variable]
			parts := make([]int, 0, size)
			for part := start; part < start+size; part++ {
				if err := work.step(1); err != nil {
					return nil, err
				}
				present, _ := source.Has(part)
				if present {
					parts = append(parts, part)
				}
			}
			if len(parts) < 2 {
				continue
			}
			if len(rows) > (work.options.MaxCubes-work.generated)/len(parts) {
				return nil, ErrLimit
			}
			expanded := make([]*EspressoSet, 0, len(rows)*len(parts))
			for _, row := range rows {
				for _, chosen := range parts {
					if err := work.step(size); err != nil {
						return nil, err
					}
					clone, _ := row.Clone()
					for part := start; part < start+size; part++ {
						_ = clone.Set(part, part == chosen)
					}
					expanded = append(expanded, clone)
				}
			}
			rows = expanded
		}
		if err := work.add(len(rows)); err != nil {
			return nil, err
		}
		for _, row := range rows {
			_ = result.Add(row)
		}
	}
	return result, nil
}

func (layout *EspressoCubeLayout) Unravel(family *EspressoSetFamily, first int, options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	return layout.UnravelRange(family, first, layout.VariableCount()-1, options)
}

func espressoOwnedFamily(bitCount int, sets []*EspressoSet) *EspressoSetFamily {
	result, _ := NewEspressoSetFamily(len(sets), bitCount)
	for _, set := range sets {
		_ = result.Add(set)
	}
	return result
}

func espressoManipulationClones(family *EspressoSetFamily, options EspressoCoverManipulationOptions) ([]*EspressoSet, error) {
	options, err := normalizeEspressoCoverManipulationOptions(options)
	if err != nil {
		return nil, err
	}
	if err = family.check(); err != nil {
		return nil, err
	}
	if len(family.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	result := make([]*EspressoSet, len(family.sets))
	for index, set := range family.sets {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		result[index], _ = set.Clone()
	}
	return result, nil
}

type espressoCoverSortKind uint8

const (
	espressoSortLexical espressoCoverSortKind = iota
	espressoSortDescending
	espressoSortAscending
)

func espressoStableSortSets(sets []*EspressoSet, kind espressoCoverSortKind, options EspressoCoverManipulationOptions) error {
	options, err := normalizeEspressoCoverManipulationOptions(options)
	if err != nil {
		return err
	}
	operations := 0
	compare := func(first, second *EspressoSet) int {
		switch kind {
		case espressoSortDescending:
			result, _ := EspressoSetCompareDescending(first, second)
			return result
		case espressoSortAscending:
			result, _ := EspressoSetCompareAscending(first, second)
			return result
		default:
			result, _ := EspressoSetCompareLexical(first, second)
			return result
		}
	}
	for index := 1; index < len(sets); index++ {
		value, position := sets[index], index
		for position > 0 {
			operations++
			if operations > options.MaxOperations {
				return ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return ErrCancelled
			}
			if compare(sets[position-1], value) <= 0 {
				break
			}
			sets[position] = sets[position-1]
			position--
		}
		sets[position] = value
	}
	return nil
}

func (family *EspressoSetFamily) SortLexical(options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	sets, err := espressoManipulationClones(family, options)
	if err != nil {
		return nil, err
	}
	if err = espressoStableSortSets(sets, espressoSortLexical, options); err != nil {
		return nil, err
	}
	return espressoOwnedFamily(family.bits, sets), nil
}

func (family *EspressoSetFamily) SortBySize(options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	sets, err := espressoManipulationClones(family, options)
	if err != nil {
		return nil, err
	}
	for _, set := range sets {
		set.Rank, _ = set.Cardinality()
	}
	if err = espressoStableSortSets(sets, espressoSortDescending, options); err != nil {
		return nil, err
	}
	return espressoOwnedFamily(family.bits, sets), nil
}

// SortByColumnWeight implements mini_sort. descending selects the reduction
// order; ascending selects the expansion order.
func (family *EspressoSetFamily) SortByColumnWeight(descending bool, options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	sets, err := espressoManipulationClones(family, options)
	if err != nil {
		return nil, err
	}
	options, _ = normalizeEspressoCoverManipulationOptions(options)
	counts := make([]int, family.bits)
	operations := 0
	for _, set := range sets {
		for part := 0; part < family.bits; part++ {
			operations++
			if operations > options.MaxOperations {
				return nil, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			present, _ := set.Has(part)
			if present {
				counts[part]++
			}
		}
	}
	for _, set := range sets {
		set.Rank = 0
		for part, count := range counts {
			operations++
			if operations > options.MaxOperations {
				return nil, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			present, _ := set.Has(part)
			if present {
				set.Rank += count
			}
		}
	}
	sortOptions := options
	sortOptions.MaxOperations -= operations
	kind := espressoSortAscending
	if descending {
		kind = espressoSortDescending
	}
	if sortOptions.MaxOperations <= 0 && len(sets) > 1 {
		return nil, ErrLimit
	}
	if err = espressoStableSortSets(sets, kind, sortOptions); err != nil {
		return nil, err
	}
	return espressoOwnedFamily(family.bits, sets), nil
}

func (layout *EspressoCubeLayout) SortForReduction(family *EspressoSetFamily, options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	if err := family.check(); err != nil || family.bits != layout.bitCount {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	options, err := normalizeEspressoCoverManipulationOptions(options)
	if err != nil {
		return nil, err
	}
	if len(family.sets) > options.MaxCubes {
		return nil, ErrLimit
	}
	sets, err := espressoManipulationClones(family, options)
	if err != nil || len(sets) < 2 {
		if err != nil {
			return nil, err
		}
		return espressoOwnedFamily(layout.bitCount, sets), nil
	}
	largest := sets[0]
	largestSize, _ := largest.Cardinality()
	for _, cube := range sets[1:] {
		size, _ := cube.Cardinality()
		if size > largestSize {
			largest, largestSize = cube, size
		}
	}
	operations := 0
	for _, cube := range sets {
		operations += max(1, layout.VariableCount())
		if operations > options.MaxOperations {
			return nil, ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		distance, _ := layout.Distance(largest, cube)
		size, _ := cube.Cardinality()
		cube.Rank = (layout.VariableCount()-distance)<<7 + min(size, 127)
	}
	sortOptions := options
	sortOptions.MaxOperations -= operations
	if sortOptions.MaxOperations <= 0 {
		return nil, ErrLimit
	}
	if err = espressoStableSortSets(sets, espressoSortDescending, sortOptions); err != nil {
		return nil, err
	}
	return espressoOwnedFamily(layout.bitCount, sets), nil
}

// ShuffleCover replaces process-global random state with a caller seed.
func (family *EspressoSetFamily) ShuffleCover(seed uint64, options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	sets, err := espressoManipulationClones(family, options)
	if err != nil {
		return nil, err
	}
	options, _ = normalizeEspressoCoverManipulationOptions(options)
	if seed == 0 {
		seed = 0x9e3779b97f4a7c15
	}
	next := func() uint64 {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		return seed
	}
	operations := 0
	for index := len(sets) - 1; index > 0; index-- {
		operations++
		if operations > options.MaxOperations {
			return nil, ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		other := int(next() % uint64(index+1))
		sets[index], sets[other] = sets[other], sets[index]
	}
	return espressoOwnedFamily(family.bits, sets), nil
}

type EspressoCubeListPartition struct {
	Left, Right               *EspressoCubeList
	LeftIndices, RightIndices []int
	Partitioned               bool
}

// PartitionComponents extracts the first connected component under ccommon.
func (list *EspressoCubeList) PartitionComponents(options EspressoCoverManipulationOptions) (EspressoCubeListPartition, error) {
	if err := list.check(); err != nil {
		return EspressoCubeListPartition{}, err
	}
	work, err := newEspressoCoverManipulationWork(options)
	if err != nil {
		return EspressoCubeListPartition{}, err
	}
	if len(list.cubes) > work.options.MaxCubes {
		return EspressoCubeListPartition{}, ErrLimit
	}
	if len(list.cubes) == 0 {
		return EspressoCubeListPartition{}, nil
	}
	selected := make([]bool, len(list.cubes))
	selected[0] = true
	seed, _ := list.cubes[0].Clone()
	for changed := true; changed; {
		changed = false
		for index, cube := range list.cubes {
			if selected[index] {
				continue
			}
			if err := work.step(max(1, list.layout.VariableCount())); err != nil {
				return EspressoCubeListPartition{}, err
			}
			common, err := list.layout.CommonActive(cube, seed, list.cofactor)
			if err != nil {
				return EspressoCubeListPartition{}, err
			}
			if common {
				seed, _ = seed.Intersection(cube)
				selected[index], changed = true, true
			}
		}
	}
	leftIndices, rightIndices := make([]int, 0), make([]int, 0)
	for index, yes := range selected {
		if yes {
			leftIndices = append(leftIndices, index)
		} else {
			rightIndices = append(rightIndices, index)
		}
	}
	if len(rightIndices) == 0 {
		return EspressoCubeListPartition{LeftIndices: leftIndices}, nil
	}
	makeList := func(indices []int) *EspressoCubeList {
		cofactor, _ := list.cofactor.Clone()
		result := &EspressoCubeList{layout: list.layout, cofactor: cofactor, cubes: make([]*EspressoSet, 0, len(indices))}
		for _, index := range indices {
			clone, _ := list.cubes[index].Clone()
			result.cubes = append(result.cubes, clone)
		}
		return result
	}
	return EspressoCubeListPartition{Left: makeList(leftIndices), Right: makeList(rightIndices), LeftIndices: leftIndices, RightIndices: rightIndices, Partitioned: true}, nil
}

// CofactorOutput selects rows containing an output part and makes the entire
// output field unrestricted. part is relative to the output variable.
func (layout *EspressoCubeLayout) CofactorOutput(family *EspressoSetFamily, part int, options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	if err := family.check(); err != nil || family.bits != layout.bitCount {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	output := layout.OutputVariable()
	if output < 0 || part < 0 || part >= layout.partSizes[output] {
		return nil, ErrDimension
	}
	work, err := newEspressoCoverManipulationWork(options)
	if err != nil {
		return nil, err
	}
	start, size := layout.firstParts[output], layout.partSizes[output]
	result, _ := NewEspressoSetFamily(len(family.sets), layout.bitCount)
	for _, cube := range family.sets {
		if err := work.step(size + 1); err != nil {
			return nil, err
		}
		selected, _ := cube.Has(start + part)
		if !selected {
			continue
		}
		if err := work.add(1); err != nil {
			return nil, err
		}
		clone, _ := cube.Clone()
		for current := start; current < start+size; current++ {
			_ = clone.Set(current, true)
		}
		clone.Flags &^= EspressoSetPrime
		_ = result.Add(clone)
	}
	return result, nil
}

// UncofactorOutput restricts every row to one relative output part.
func (layout *EspressoCubeLayout) UncofactorOutput(family *EspressoSetFamily, part int, options EspressoCoverManipulationOptions) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	if err := family.check(); err != nil || family.bits != layout.bitCount {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	output := layout.OutputVariable()
	if output < 0 || part < 0 || part >= layout.partSizes[output] {
		return nil, ErrDimension
	}
	work, err := newEspressoCoverManipulationWork(options)
	if err != nil {
		return nil, err
	}
	start, size := layout.firstParts[output], layout.partSizes[output]
	result, _ := NewEspressoSetFamily(len(family.sets), layout.bitCount)
	for _, cube := range family.sets {
		if err := work.step(size); err != nil {
			return nil, err
		}
		if err := work.add(1); err != nil {
			return nil, err
		}
		clone, _ := cube.Clone()
		for current := start; current < start+size; current++ {
			_ = clone.Set(current, current == start+part)
		}
		_ = result.Add(clone)
	}
	return result, nil
}
