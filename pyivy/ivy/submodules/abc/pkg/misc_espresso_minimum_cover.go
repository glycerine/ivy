package abc

// Deterministic minimum-column cover from src/misc/espresso/sminterf.c.

import (
	"math/bits"
	"sort"
)

type MvcMinimumCoverOptions struct {
	MaxRows, MaxColumns, MaxPairChecks, MaxSearchNodes int
	Stop                                               func() bool
}

func normalizeMvcMinimumCoverOptions(options MvcMinimumCoverOptions) (MvcMinimumCoverOptions, error) {
	if options.MaxRows < 0 || options.MaxColumns < 0 || options.MaxPairChecks < 0 || options.MaxSearchNodes < 0 {
		return options, ErrDimension
	}
	if options.MaxRows == 0 {
		options.MaxRows = 1 << 20
	}
	if options.MaxColumns == 0 {
		options.MaxColumns = 1 << 20
	}
	if options.MaxPairChecks == 0 {
		options.MaxPairChecks = 4 << 20
	}
	if options.MaxSearchNodes == 0 {
		options.MaxSearchNodes = 1 << 20
	}
	return options, nil
}

func mvcMinimumCoverRows(cover *MvcCover, options MvcMinimumCoverOptions) ([][]int, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	if len(cover.cubes) > options.MaxRows || cover.bits > options.MaxColumns {
		return nil, ErrLimit
	}
	rows := make([][]int, len(cover.cubes))
	for row, cube := range cover.cubes {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		for wordIndex, word := range cube.words {
			for word != 0 {
				bit := bits.TrailingZeros32(word)
				column := wordIndex*32 + bit
				if column < cover.bits {
					rows[row] = append(rows[row], column)
				}
				word &^= uint32(1) << uint(bit)
			}
		}
		if len(rows[row]) == 0 {
			return nil, ErrInvariant
		}
	}
	return mvcRemoveDominatedCoverRows(rows, options)
}

func mvcColumnsSubset(first, second []int) bool {
	left, right := 0, 0
	for left < len(first) && right < len(second) {
		if first[left] == second[right] {
			left++
			right++
		} else if first[left] > second[right] {
			right++
		} else {
			return false
		}
	}
	return left == len(first)
}

func mvcRemoveDominatedCoverRows(rows [][]int, options MvcMinimumCoverOptions) ([][]int, error) {
	redundant := make([]bool, len(rows))
	checks := 0
	for first := range rows {
		for second := range rows {
			checks++
			if checks > options.MaxPairChecks {
				return nil, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			if first == second || redundant[second] || len(rows[first]) > len(rows[second]) {
				continue
			}
			if mvcColumnsSubset(rows[first], rows[second]) && (len(rows[first]) < len(rows[second]) || first < second) {
				redundant[second] = true
			}
		}
	}
	result := make([][]int, 0, len(rows))
	for index, row := range rows {
		if !redundant[index] {
			result = append(result, row)
		}
	}
	return result, nil
}

func mvcRowContains(row []int, column int) bool {
	position := sort.SearchInts(row, column)
	return position < len(row) && row[position] == column
}

func mvcFilterCoveredRows(rows [][]int, column int) [][]int {
	result := make([][]int, 0, len(rows))
	for _, row := range rows {
		if !mvcRowContains(row, column) {
			result = append(result, row)
		}
	}
	return result
}

func mvcGreedyColumnCover(rows [][]int, columns int) []int {
	selected := make([]bool, columns)
	remaining := rows
	for len(remaining) != 0 {
		counts := make([]int, columns)
		for _, row := range remaining {
			for _, column := range row {
				counts[column]++
			}
		}
		best := 0
		for column := 1; column < columns; column++ {
			if counts[column] > counts[best] {
				best = column
			}
		}
		selected[best] = true
		remaining = mvcFilterCoveredRows(remaining, best)
	}
	result := make([]int, 0)
	for column, yes := range selected {
		if yes {
			result = append(result, column)
		}
	}
	return result
}

func mvcDisjointRowLowerBound(rows [][]int, columns int) int {
	order := make([]int, len(rows))
	for index := range order {
		order[index] = index
	}
	sort.SliceStable(order, func(left, right int) bool { return len(rows[order[left]]) < len(rows[order[right]]) })
	used := make([]bool, columns)
	bound := 0
	for _, index := range order {
		disjoint := true
		for _, column := range rows[index] {
			if used[column] {
				disjoint = false
				break
			}
		}
		if !disjoint {
			continue
		}
		bound++
		for _, column := range rows[index] {
			used[column] = true
		}
	}
	return bound
}

// MinimumColumnCover returns a minimum-cardinality set of bit positions that
// intersects every cube. The native wrapper requested the sparse solver's
// non-branching heuristic; Go performs bounded exact branch-and-bound so the
// method's minimum-cover contract is independently checkable and deterministic.
func (cover *MvcCover) MinimumColumnCover(options MvcMinimumCoverOptions) (*MvcCube, error) {
	options, err := normalizeMvcMinimumCoverOptions(options)
	if err != nil {
		return nil, err
	}
	rows, err := mvcMinimumCoverRows(cover, options)
	if err != nil {
		return nil, err
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	best := mvcGreedyColumnCover(rows, cover.bits)
	selected := make([]bool, cover.bits)
	searchNodes := 0
	var search func([][]int, int) error
	search = func(remaining [][]int, count int) error {
		searchNodes++
		if searchNodes > options.MaxSearchNodes {
			return ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if len(remaining) == 0 {
			if count < len(best) {
				best = best[:0]
				for column, yes := range selected {
					if yes {
						best = append(best, column)
					}
				}
			}
			return nil
		}
		if count+mvcDisjointRowLowerBound(remaining, cover.bits) >= len(best) {
			return nil
		}
		choice := remaining[0]
		for _, row := range remaining[1:] {
			if len(row) < len(choice) {
				choice = row
			}
		}
		type candidate struct{ column, coverage int }
		candidates := make([]candidate, len(choice))
		for index, column := range choice {
			candidates[index].column = column
			for _, row := range remaining {
				if mvcRowContains(row, column) {
					candidates[index].coverage++
				}
			}
		}
		sort.SliceStable(candidates, func(left, right int) bool {
			if candidates[left].coverage != candidates[right].coverage {
				return candidates[left].coverage > candidates[right].coverage
			}
			return candidates[left].column < candidates[right].column
		})
		for _, candidate := range candidates {
			selected[candidate.column] = true
			if err := search(mvcFilterCoveredRows(remaining, candidate.column), count+1); err != nil {
				selected[candidate.column] = false
				return err
			}
			selected[candidate.column] = false
		}
		return nil
	}
	if len(rows) != 0 {
		if err = search(rows, 0); err != nil {
			return nil, err
		}
	}
	result, err := cover.NewCube()
	if err != nil {
		return nil, err
	}
	for _, column := range best {
		result.words[column>>5] |= uint32(1) << uint(column&31)
	}
	return result, nil
}
