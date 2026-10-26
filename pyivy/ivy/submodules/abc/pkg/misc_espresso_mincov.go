package abc

// Weighted sparse minimum covering from src/misc/espresso/mincov.c.

import (
	"math"
	"math/bits"
	"slices"
	"sort"
)

type MvcMinimumCoverSearchOptions struct {
	MaxRows, MaxColumns, MaxPairChecks, MaxSearchNodes int
	Heuristic                                          bool
	Stop                                               func() bool
}

type MvcMinimumCoverSearchStats struct {
	Nodes, MaxDepth, LowerBound int
	Heuristic                   bool
}

type MvcMinimumCoverSearchResult struct {
	Columns []int
	Cost    int
	Stats   MvcMinimumCoverSearchStats
}

func normalizeMvcMinimumCoverSearchOptions(options MvcMinimumCoverSearchOptions) (MvcMinimumCoverSearchOptions, error) {
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

func mvcCoverWeights(columns int, weights []int) ([]int, error) {
	if weights != nil && len(weights) != columns {
		return nil, ErrDimension
	}
	result := make([]int, columns)
	for column := range result {
		result[column] = 1
		if weights != nil {
			result[column] = weights[column]
		}
		if result[column] < 0 {
			return nil, ErrDimension
		}
	}
	return result, nil
}

type mvcCoverCandidate struct {
	column, coverage int
}

func mvcRatioGreater(first, second mvcCoverCandidate, weights []int) bool {
	firstWeight, secondWeight := weights[first.column], weights[second.column]
	if firstWeight == 0 || secondWeight == 0 {
		if firstWeight != secondWeight {
			return firstWeight == 0
		}
		if first.coverage != second.coverage {
			return first.coverage > second.coverage
		}
		return first.column < second.column
	}
	firstHigh, firstLow := bits.Mul64(uint64(first.coverage), uint64(secondWeight))
	secondHigh, secondLow := bits.Mul64(uint64(second.coverage), uint64(firstWeight))
	if firstHigh != secondHigh {
		return firstHigh > secondHigh
	}
	if firstLow != secondLow {
		return firstLow > secondLow
	}
	return first.column < second.column
}

func mvcCoverCandidates(row []int, rows [][]int, weights []int) []mvcCoverCandidate {
	result := make([]mvcCoverCandidate, len(row))
	for index, column := range row {
		result[index].column = column
		for _, candidateRow := range rows {
			if mvcRowContains(candidateRow, column) {
				result[index].coverage++
			}
		}
	}
	sort.SliceStable(result, func(first, second int) bool {
		return mvcRatioGreater(result[first], result[second], weights)
	})
	return result
}

func mvcWeightedGreedyCover(rows [][]int, columns int, weights []int, stop func() bool) ([]int, int, error) {
	remaining := rows
	selected := make([]bool, columns)
	cost := 0
	for len(remaining) != 0 {
		if stop != nil && stop() {
			return nil, 0, ErrCancelled
		}
		present := make([]int, 0)
		seen := make([]bool, columns)
		for _, row := range remaining {
			for _, column := range row {
				if !seen[column] {
					seen[column] = true
					present = append(present, column)
				}
			}
		}
		candidates := mvcCoverCandidates(present, remaining, weights)
		if len(candidates) == 0 {
			return nil, 0, ErrInvariant
		}
		column := candidates[0].column
		if weights[column] > math.MaxInt-cost {
			return nil, 0, ErrLimit
		}
		cost += weights[column]
		selected[column] = true
		remaining = mvcFilterCoveredRows(remaining, column)
	}
	result := make([]int, 0)
	for column, yes := range selected {
		if yes {
			result = append(result, column)
		}
	}
	return result, cost, nil
}

func mvcWeightedIndependentBound(rows [][]int, columns int, weights []int) (int, error) {
	order := make([]int, len(rows))
	for index := range order {
		order[index] = index
	}
	sort.SliceStable(order, func(first, second int) bool { return len(rows[order[first]]) < len(rows[order[second]]) })
	used := make([]bool, columns)
	bound := 0
	for _, rowIndex := range order {
		row := rows[rowIndex]
		disjoint := true
		for _, column := range row {
			if used[column] {
				disjoint = false
				break
			}
		}
		if !disjoint {
			continue
		}
		minimum := math.MaxInt
		for _, column := range row {
			minimum = min(minimum, weights[column])
			used[column] = true
		}
		if minimum > math.MaxInt-bound {
			return 0, ErrLimit
		}
		bound += minimum
	}
	return bound, nil
}

func mvcVerifyColumnCover(rows [][]int, selected []bool) bool {
	for _, row := range rows {
		covered := false
		for _, column := range row {
			if selected[column] {
				covered = true
				break
			}
		}
		if !covered {
			return false
		}
	}
	return true
}

// MinimumWeightedColumnCover returns a deterministic exact minimum-cost cover.
// Heuristic mode returns the deterministic no-branch greedy upper bound. Both
// modes return owned sorted column IDs and never mutate the input cover.
func (cover *MvcCover) MinimumWeightedColumnCover(weights []int, options MvcMinimumCoverSearchOptions) (MvcMinimumCoverSearchResult, error) {
	options, err := normalizeMvcMinimumCoverSearchOptions(options)
	if err != nil {
		return MvcMinimumCoverSearchResult{}, err
	}
	actualWeights, err := mvcCoverWeights(cover.BitCount(), weights)
	if err != nil {
		return MvcMinimumCoverSearchResult{}, err
	}
	rows, err := mvcMinimumCoverRows(cover, MvcMinimumCoverOptions{
		MaxRows: options.MaxRows, MaxColumns: options.MaxColumns,
		MaxPairChecks: options.MaxPairChecks, MaxSearchNodes: options.MaxSearchNodes,
		Stop: options.Stop,
	})
	if err != nil {
		return MvcMinimumCoverSearchResult{}, err
	}
	greedy, greedyCost, err := mvcWeightedGreedyCover(rows, cover.bits, actualWeights, options.Stop)
	if err != nil {
		return MvcMinimumCoverSearchResult{}, err
	}
	lowerBound, err := mvcWeightedIndependentBound(rows, cover.bits, actualWeights)
	if err != nil {
		return MvcMinimumCoverSearchResult{}, err
	}
	result := MvcMinimumCoverSearchResult{
		Columns: greedy, Cost: greedyCost,
		Stats: MvcMinimumCoverSearchStats{LowerBound: lowerBound, Heuristic: options.Heuristic},
	}
	if options.Heuristic || len(rows) == 0 {
		return result, nil
	}
	selected := make([]bool, cover.bits)
	best := slices.Clone(greedy)
	bestCost := greedyCost
	var search func([][]int, int, int) error
	search = func(remaining [][]int, cost, depth int) error {
		result.Stats.Nodes++
		result.Stats.MaxDepth = max(result.Stats.MaxDepth, depth)
		if result.Stats.Nodes > options.MaxSearchNodes {
			return ErrLimit
		}
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if len(remaining) == 0 {
			if cost < bestCost {
				bestCost = cost
				best = best[:0]
				for column, yes := range selected {
					if yes {
						best = append(best, column)
					}
				}
			}
			return nil
		}
		bound, boundErr := mvcWeightedIndependentBound(remaining, cover.bits, actualWeights)
		if boundErr != nil {
			return boundErr
		}
		if bound > math.MaxInt-cost || cost+bound >= bestCost {
			return nil
		}
		choice := remaining[0]
		for _, row := range remaining[1:] {
			if len(row) < len(choice) {
				choice = row
			}
		}
		for _, candidate := range mvcCoverCandidates(choice, remaining, actualWeights) {
			column := candidate.column
			if actualWeights[column] > math.MaxInt-cost {
				return ErrLimit
			}
			selected[column] = true
			if searchErr := search(mvcFilterCoveredRows(remaining, column), cost+actualWeights[column], depth+1); searchErr != nil {
				selected[column] = false
				return searchErr
			}
			selected[column] = false
		}
		return nil
	}
	if err = search(rows, 0, 0); err != nil {
		return MvcMinimumCoverSearchResult{}, err
	}
	selected = make([]bool, cover.bits)
	for _, column := range best {
		selected[column] = true
	}
	if !mvcVerifyColumnCover(rows, selected) {
		return MvcMinimumCoverSearchResult{}, ErrInvariant
	}
	result.Columns, result.Cost = slices.Clone(best), bestCost
	return result, nil
}
