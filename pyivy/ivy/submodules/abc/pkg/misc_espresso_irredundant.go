package abc

// Irredundant-cover selection and tautology checks from src/misc/espresso/irred.c.

import "slices"

type EspressoIrredundantOptions struct {
	Exact                            bool
	MaxCubes, MaxOperations, MaxRows int
	Stop                             func() bool
}

func normalizeEspressoIrredundantOptions(options EspressoIrredundantOptions) (EspressoIrredundantOptions, error) {
	if options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxRows < 0 {
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
	return options, nil
}

type EspressoIrredundantSplit struct {
	Essential, TotallyRedundant, PartiallyRedundant *EspressoSetFamily
	EssentialIndices, TotallyRedundantIndices       []int
	PartiallyRedundantIndices                       []int
}

type EspressoIrredundantTable struct {
	Rows [][]int
}

type EspressoIrredundantResult struct {
	Marked, Cover *EspressoSetFamily
	Split         EspressoIrredundantSplit
	Table         EspressoIrredundantTable
	Selection     []int
}

func (layout *EspressoCubeLayout) irredCubeCovered(cube *EspressoSet, cover *EspressoSetFamily, work *espressoSharpWork) (bool, error) {
	remainder, err := layout.sharpCubeByCover(cube, cover, work)
	if err != nil {
		return false, err
	}
	return len(remainder.sets) == 0, nil
}

// CubeCoveredByCover implements cube_is_covered without a borrowed cubelist.
func (layout *EspressoCubeLayout) CubeCoveredByCover(cover *EspressoSetFamily, cube *EspressoSet, options EspressoIrredundantOptions) (bool, error) {
	options, err := normalizeEspressoIrredundantOptions(options)
	if err != nil {
		return false, err
	}
	work, _ := newEspressoSharpWork(EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	if err = layout.checkSet(cube); err != nil {
		return false, err
	}
	if err = cover.check(); err != nil || cover.bits != layout.bitCount {
		if err != nil {
			return false, err
		}
		return false, ErrDimension
	}
	return layout.irredCubeCovered(cube, cover, work)
}

// CoverIsTautology implements tautology and its recursive special cases as an
// exact universe-minus-cover check through the validated complement engine.
func (layout *EspressoCubeLayout) CoverIsTautology(cover *EspressoSetFamily, options EspressoIrredundantOptions) (bool, error) {
	options, err := normalizeEspressoIrredundantOptions(options)
	if err != nil {
		return false, err
	}
	work, _ := newEspressoSharpWork(EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	complement, err := layout.complementCover(cover, work)
	if err != nil {
		return false, err
	}
	return len(complement.sets) == 0, nil
}

func espressoIrredFamily(layout *EspressoCubeLayout, source *EspressoSetFamily, indices []int) *EspressoSetFamily {
	result, _ := NewEspressoSetFamily(len(indices), layout.bitCount)
	for _, index := range indices {
		clone, _ := source.sets[index].Clone()
		clone.Rank = index
		_ = result.Add(clone)
	}
	return result
}

func (layout *EspressoCubeLayout) splitIrredundant(on, dontCare *EspressoSetFamily, work *espressoSharpWork) (EspressoIrredundantSplit, error) {
	if err := layout.check(); err != nil {
		return EspressoIrredundantSplit{}, err
	}
	for _, family := range []*EspressoSetFamily{on, dontCare} {
		if err := family.check(); err != nil {
			return EspressoIrredundantSplit{}, err
		}
		if family.bits != layout.bitCount {
			return EspressoIrredundantSplit{}, ErrDimension
		}
	}
	if len(on.sets)+len(dontCare.sets) > work.options.MaxCubes {
		return EspressoIrredundantSplit{}, ErrLimit
	}
	essentialIndices, redundant := make([]int, 0), make([]int, 0)
	for index, cube := range on.sets {
		other, _ := NewEspressoSetFamily(len(on.sets)+len(dontCare.sets)-1, layout.bitCount)
		for current, candidate := range on.sets {
			if current != index {
				_ = other.Add(candidate)
			}
		}
		for _, candidate := range dontCare.sets {
			_ = other.Add(candidate)
		}
		covered, err := layout.irredCubeCovered(cube, other, work)
		if err != nil {
			return EspressoIrredundantSplit{}, err
		}
		if covered {
			redundant = append(redundant, index)
		} else {
			essentialIndices = append(essentialIndices, index)
		}
	}
	essential := espressoIrredFamily(layout, on, essentialIndices)
	essentialAndDC, _ := JoinEspressoSetFamilies(essential, dontCare)
	totalIndices, partialIndices := make([]int, 0), make([]int, 0)
	for _, index := range redundant {
		covered, err := layout.irredCubeCovered(on.sets[index], essentialAndDC, work)
		if err != nil {
			return EspressoIrredundantSplit{}, err
		}
		if covered {
			totalIndices = append(totalIndices, index)
		} else {
			partialIndices = append(partialIndices, index)
		}
	}
	return EspressoIrredundantSplit{
		Essential: essential, TotallyRedundant: espressoIrredFamily(layout, on, totalIndices), PartiallyRedundant: espressoIrredFamily(layout, on, partialIndices),
		EssentialIndices: essentialIndices, TotallyRedundantIndices: totalIndices, PartiallyRedundantIndices: partialIndices,
	}, nil
}

func (layout *EspressoCubeLayout) SplitIrredundant(on, dontCare *EspressoSetFamily, options EspressoIrredundantOptions) (EspressoIrredundantSplit, error) {
	options, err := normalizeEspressoIrredundantOptions(options)
	if err != nil {
		return EspressoIrredundantSplit{}, err
	}
	work, _ := newEspressoSharpWork(EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	return layout.splitIrredundant(on, dontCare, work)
}

type espressoIrredCell struct {
	cover   *EspressoSetFamily
	members []int
}

func espressoIrredAppendRow(rows *[][]int, row []int, maximum int) error {
	for _, existing := range *rows {
		if slices.Equal(existing, row) {
			return nil
		}
	}
	if len(*rows) >= maximum {
		return ErrLimit
	}
	*rows = append(*rows, slices.Clone(row))
	return nil
}

func (layout *EspressoCubeLayout) deriveIrredundantTable(dontCare *EspressoSetFamily, split EspressoIrredundantSplit, work *espressoSharpWork, maxRows int) (EspressoIrredundantTable, error) {
	if err := layout.check(); err != nil {
		return EspressoIrredundantTable{}, err
	}
	if err := dontCare.check(); err != nil {
		return EspressoIrredundantTable{}, err
	}
	if dontCare.bits != layout.bitCount {
		return EspressoIrredundantTable{}, ErrDimension
	}
	for _, item := range []struct {
		family  *EspressoSetFamily
		indices []int
	}{
		{split.Essential, split.EssentialIndices},
		{split.TotallyRedundant, split.TotallyRedundantIndices},
		{split.PartiallyRedundant, split.PartiallyRedundantIndices},
	} {
		if err := item.family.check(); err != nil {
			return EspressoIrredundantTable{}, err
		}
		if item.family.bits != layout.bitCount || len(item.family.sets) != len(item.indices) {
			return EspressoIrredundantTable{}, ErrInvariant
		}
		for _, index := range item.indices {
			if index < 0 {
				return EspressoIrredundantTable{}, ErrInvariant
			}
		}
	}
	fixed, err := JoinEspressoSetFamilies(dontCare, split.Essential)
	if err != nil {
		return EspressoIrredundantTable{}, err
	}
	target, err := layout.coverDifference(split.PartiallyRedundant, fixed, work)
	if err != nil {
		return EspressoIrredundantTable{}, err
	}
	if len(target.sets) == 0 {
		return EspressoIrredundantTable{}, nil
	}
	cells := []espressoIrredCell{{cover: target}}
	for candidate, cube := range split.PartiallyRedundant.sets {
		next := make([]espressoIrredCell, 0, 2*len(cells))
		for _, cell := range cells {
			inside, insideErr := layout.intersectFamilies(cell.cover, espressoIrredFamily(layout, split.PartiallyRedundant, []int{candidate}), work)
			if insideErr != nil {
				return EspressoIrredundantTable{}, insideErr
			}
			outside, outsideErr := layout.disjointSharpFamilyByCube(cell.cover, cube, work, true)
			if outsideErr != nil {
				return EspressoIrredundantTable{}, outsideErr
			}
			if len(inside.sets) != 0 {
				members := append(slices.Clone(cell.members), candidate)
				next = append(next, espressoIrredCell{cover: inside, members: members})
			}
			if len(outside.sets) != 0 {
				next = append(next, espressoIrredCell{cover: outside, members: slices.Clone(cell.members)})
			}
			if len(next) > maxRows {
				return EspressoIrredundantTable{}, ErrLimit
			}
		}
		cells = next
	}
	rows := make([][]int, 0, len(cells))
	for _, cell := range cells {
		if len(cell.members) == 0 {
			return EspressoIrredundantTable{}, ErrInvariant
		}
		row := make([]int, len(cell.members))
		for index, candidate := range cell.members {
			row[index] = split.PartiallyRedundantIndices[candidate]
		}
		if err = espressoIrredAppendRow(&rows, row, maxRows); err != nil {
			return EspressoIrredundantTable{}, err
		}
	}
	return EspressoIrredundantTable{Rows: rows}, nil
}

func (layout *EspressoCubeLayout) DeriveIrredundantTable(dontCare *EspressoSetFamily, split EspressoIrredundantSplit, options EspressoIrredundantOptions) (EspressoIrredundantTable, error) {
	options, err := normalizeEspressoIrredundantOptions(options)
	if err != nil {
		return EspressoIrredundantTable{}, err
	}
	work, _ := newEspressoSharpWork(EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	return layout.deriveIrredundantTable(dontCare, split, work, options.MaxRows)
}

func espressoIrredSelection(table EspressoIrredundantTable, partialIndices []int, exact bool, options EspressoIrredundantOptions, remainingOperations int) ([]int, error) {
	positions := make(map[int]int, len(partialIndices))
	for position, index := range partialIndices {
		positions[index] = position
	}
	selected := make([]bool, len(partialIndices))
	if len(table.Rows) == 0 {
		return nil, nil
	}
	operations := 0
	step := func() error {
		operations++
		if operations > remainingOperations {
			return ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	if exact {
		family, _ := NewEspressoSetFamily(len(table.Rows), len(partialIndices))
		for _, sourceRow := range table.Rows {
			row, _ := NewEspressoSet(len(partialIndices))
			for _, index := range sourceRow {
				position, exists := positions[index]
				if !exists {
					return nil, ErrInvariant
				}
				_ = row.Set(position, true)
			}
			_ = family.Add(row)
		}
		covers, err := UnateComplement(family, EspressoUnateOptions{MaxSets: options.MaxRows, MaxOperations: remainingOperations, Stop: options.Stop})
		if err != nil {
			return nil, err
		}
		best := -1
		for index, cover := range covers.sets {
			cardinality, _ := cover.Cardinality()
			if best < 0 {
				best = index
				continue
			}
			bestCardinality, _ := covers.sets[best].Cardinality()
			if cardinality < bestCardinality {
				best = index
			}
		}
		if best >= 0 {
			for position := range selected {
				selected[position], _ = covers.sets[best].Has(position)
			}
		}
	} else {
		remaining := make([]bool, len(table.Rows))
		for index := range remaining {
			remaining[index] = true
		}
		for {
			if err := step(); err != nil {
				return nil, err
			}
			left := 0
			for _, yes := range remaining {
				if yes {
					left++
				}
			}
			if left == 0 {
				break
			}
			best, bestCount := -1, -1
			for position, sourceIndex := range partialIndices {
				count := 0
				for row, indices := range table.Rows {
					if err := step(); err != nil {
						return nil, err
					}
					if remaining[row] && slices.Contains(indices, sourceIndex) {
						count++
					}
				}
				if count > bestCount {
					best, bestCount = position, count
				}
			}
			if best < 0 || bestCount <= 0 {
				return nil, ErrInvariant
			}
			selected[best] = true
			for row, indices := range table.Rows {
				if slices.Contains(indices, partialIndices[best]) {
					remaining[row] = false
				}
			}
		}
		for position := len(selected) - 1; position >= 0; position-- {
			if !selected[position] {
				continue
			}
			selected[position] = false
			valid := true
			for _, row := range table.Rows {
				covered := false
				for current, yes := range selected {
					if err := step(); err != nil {
						return nil, err
					}
					if yes && slices.Contains(row, partialIndices[current]) {
						covered = true
						break
					}
				}
				if !covered {
					valid = false
					break
				}
			}
			if !valid {
				selected[position] = true
			}
		}
	}
	result := make([]int, 0)
	for position, yes := range selected {
		if yes {
			result = append(result, partialIndices[position])
		}
	}
	return result, nil
}

// MarkIrredundant returns both the complete marked source-order cover and its
// active subset. Essential rows also receive RelEssential.
func (layout *EspressoCubeLayout) MarkIrredundant(on, dontCare *EspressoSetFamily, options EspressoIrredundantOptions) (EspressoIrredundantResult, error) {
	options, err := normalizeEspressoIrredundantOptions(options)
	if err != nil {
		return EspressoIrredundantResult{}, err
	}
	work, _ := newEspressoSharpWork(EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	split, err := layout.splitIrredundant(on, dontCare, work)
	if err != nil {
		return EspressoIrredundantResult{}, err
	}
	table, err := layout.deriveIrredundantTable(dontCare, split, work, options.MaxRows)
	if err != nil {
		return EspressoIrredundantResult{}, err
	}
	remainingOperations := options.MaxOperations - work.operations
	if remainingOperations <= 0 && len(table.Rows) != 0 {
		return EspressoIrredundantResult{}, ErrLimit
	}
	selection, err := espressoIrredSelection(table, split.PartiallyRedundantIndices, options.Exact, options, remainingOperations)
	if err != nil {
		return EspressoIrredundantResult{}, err
	}
	active := make([]bool, len(on.sets))
	essential := make([]bool, len(on.sets))
	for _, index := range split.EssentialIndices {
		active[index], essential[index] = true, true
	}
	for _, index := range selection {
		active[index] = true
	}
	marked, _ := on.Clone()
	cover, _ := NewEspressoSetFamily(len(on.sets), layout.bitCount)
	for index, cube := range marked.sets {
		cube.Rank = index
		cube.Flags &^= EspressoSetActive | EspressoSetRelEssential
		if essential[index] {
			cube.Flags |= EspressoSetRelEssential
		}
		if active[index] {
			cube.Flags |= EspressoSetActive
			_ = cover.Add(cube)
		}
	}
	// Add computed Active counts after direct flag mutation of the clone.
	marked.activeCount = len(split.EssentialIndices) + len(selection)
	return EspressoIrredundantResult{Marked: marked, Cover: cover, Split: split, Table: table, Selection: selection}, nil
}

func (layout *EspressoCubeLayout) Irredundant(on, dontCare *EspressoSetFamily, options EspressoIrredundantOptions) (EspressoIrredundantResult, error) {
	return layout.MarkIrredundant(on, dontCare, options)
}
