package abc

// Set-family containment and distance merging from src/misc/espresso/contain.c.

import "slices"

type EspressoContainOptions struct {
	MaxSets, MaxChecks int
	Stop               func() bool
}

func normalizeEspressoContainOptions(options EspressoContainOptions) (EspressoContainOptions, error) {
	if options.MaxSets < 0 || options.MaxChecks < 0 {
		return options, ErrDimension
	}
	if options.MaxSets == 0 {
		options.MaxSets = 1 << 20
	}
	if options.MaxChecks == 0 {
		options.MaxChecks = 16 << 20
	}
	return options, nil
}

type espressoContainItem struct {
	set      *EspressoSet
	original int
}

func espressoContainItems(family *EspressoSetFamily, options EspressoContainOptions) ([]espressoContainItem, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	if len(family.sets) > options.MaxSets {
		return nil, ErrLimit
	}
	items := make([]espressoContainItem, len(family.sets))
	for index, source := range family.sets {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		clone, _ := source.Clone()
		clone.Rank, _ = clone.Cardinality()
		items[index] = espressoContainItem{set: clone, original: index}
	}
	return items, nil
}

func espressoSortContainItems(items []espressoContainItem, descending bool) {
	slices.SortStableFunc(items, func(first, second espressoContainItem) int {
		var comparison int
		if descending {
			comparison, _ = EspressoSetCompareDescending(first.set, second.set)
		} else {
			comparison, _ = EspressoSetCompareAscending(first.set, second.set)
		}
		return comparison
	})
}

func espressoDeduplicateItems(items []espressoContainItem) []espressoContainItem {
	if len(items) == 0 {
		return items
	}
	result := items[:1]
	for _, item := range items[1:] {
		equal, _ := result[len(result)-1].set.Equal(item.set)
		if !equal {
			result = append(result, item)
		}
	}
	return result
}

func espressoItemsFamily(bitCount int, items []espressoContainItem) (*EspressoSetFamily, error) {
	result, err := NewEspressoSetFamily(len(items), bitCount)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		result.sets = append(result.sets, item.set)
		if item.set.Flags&EspressoSetActive != 0 {
			result.activeCount++
		}
	}
	return result, nil
}

func espressoContain(family *EspressoSetFamily, reverse bool, options EspressoContainOptions) (*EspressoSetFamily, []int, error) {
	options, err := normalizeEspressoContainOptions(options)
	if err != nil {
		return nil, nil, err
	}
	items, err := espressoContainItems(family, options)
	if err != nil {
		return nil, nil, err
	}
	espressoSortContainItems(items, !reverse)
	items = espressoDeduplicateItems(items)
	result := make([]espressoContainItem, 0, len(items))
	checks := 0
	for _, candidate := range items {
		contained := false
		for _, retained := range result {
			if candidate.set.Rank == retained.set.Rank {
				continue
			}
			checks++
			if checks > options.MaxChecks {
				return nil, nil, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return nil, nil, ErrCancelled
			}
			var implies bool
			if reverse {
				implies, _ = retained.set.Implies(candidate.set)
			} else {
				implies, _ = candidate.set.Implies(retained.set)
			}
			if implies {
				contained = true
				break
			}
		}
		if !contained {
			result = append(result, candidate)
		}
	}
	output, err := espressoItemsFamily(family.bits, result)
	if err != nil {
		return nil, nil, err
	}
	indices := make([]int, len(result))
	for index, item := range result {
		indices[index] = item.original
	}
	return output, indices, nil
}

// Contain removes sets contained by a strictly larger set and returns the
// survivors in source descending order of cardinality and packed words.
func (family *EspressoSetFamily) Contain(options EspressoContainOptions) (*EspressoSetFamily, error) {
	result, _, err := espressoContain(family, false, options)
	return result, err
}

// ReverseContain removes sets that contain a strictly smaller set.
func (family *EspressoSetFamily) ReverseContain(options EspressoContainOptions) (*EspressoSetFamily, error) {
	result, _, err := espressoContain(family, true, options)
	return result, err
}

func (family *EspressoSetFamily) IndexedContain(rowIndices []int, options EspressoContainOptions) (*EspressoSetFamily, []int, error) {
	if err := family.check(); err != nil {
		return nil, nil, err
	}
	if len(rowIndices) != len(family.sets) {
		return nil, nil, ErrDimension
	}
	result, original, err := espressoContain(family, false, options)
	if err != nil {
		return nil, nil, err
	}
	indices := make([]int, len(original))
	for index, source := range original {
		indices[index] = rowIndices[source]
	}
	return result, indices, nil
}

func (family *EspressoSetFamily) Deduplicate(options EspressoContainOptions) (*EspressoSetFamily, error) {
	options, err := normalizeEspressoContainOptions(options)
	if err != nil {
		return nil, err
	}
	items, err := espressoContainItems(family, options)
	if err != nil {
		return nil, err
	}
	espressoSortContainItems(items, true)
	items = espressoDeduplicateItems(items)
	return espressoItemsFamily(family.bits, items)
}

func ContainedUnionEspressoSetFamilies(first, second *EspressoSetFamily, options EspressoContainOptions) (*EspressoSetFamily, error) {
	joined, err := JoinEspressoSetFamilies(first, second)
	if err != nil {
		return nil, err
	}
	result, err := joined.Contain(options)
	_ = joined.Close()
	return result, err
}

func (family *EspressoSetFamily) DistanceMerge(mask *EspressoSet, options EspressoContainOptions) (*EspressoSetFamily, error) {
	options, err := normalizeEspressoContainOptions(options)
	if err != nil {
		return nil, err
	}
	if err = family.check(); err != nil {
		return nil, err
	}
	if err = mask.check(); err != nil || mask.bits != family.bits {
		return nil, ErrDimension
	}
	items, err := espressoContainItems(family, options)
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(items, func(first, second espressoContainItem) int {
		comparison, _ := EspressoSetCompareDistanceOne(first.set, second.set, mask)
		return comparison
	})
	if len(items) == 0 {
		return espressoItemsFamily(family.bits, items)
	}
	result := make([]espressoContainItem, 0, len(items))
	current := items[0]
	for _, item := range items[1:] {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		comparison, _ := EspressoSetCompareDistanceOne(current.set, item.set, mask)
		if comparison == 0 {
			for index, word := range item.set.words {
				current.set.words[index] |= word
			}
			continue
		}
		result = append(result, current)
		current = item
	}
	result = append(result, current)
	return espressoItemsFamily(family.bits, result)
}

func (family *EspressoSetFamily) DistanceOneMerge(layout *EspressoCubeLayout, variable int, options EspressoContainOptions) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	if family == nil || family.bits != layout.bitCount || variable < 0 || variable >= len(layout.partSizes) {
		return nil, ErrDimension
	}
	mask, _ := NewEspressoSet(layout.bitCount)
	start := layout.firstParts[variable]
	for part := start; part < start+layout.partSizes[variable]; part++ {
		_ = mask.Set(part, true)
	}
	return family.DistanceMerge(mask, options)
}
