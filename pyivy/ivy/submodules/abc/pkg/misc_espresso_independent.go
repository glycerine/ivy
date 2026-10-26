package abc

// Maximal independent-row lower bound from src/misc/espresso/indep.c.

import (
	"math"
	"math/bits"
	"slices"
	"sort"
)

type MvcIndependentRowsOptions struct {
	MaxRows, MaxColumns         int
	MaxPairChecks, MaxAdjacency int
	MaxVisits                   int
	Stop                        func() bool
}

type MvcIndependentRows struct {
	Rows           []int
	SelectionOrder []int
	Cost           int
}

func normalizeMvcIndependentRowsOptions(options MvcIndependentRowsOptions) (MvcIndependentRowsOptions, error) {
	if options.MaxRows < 0 || options.MaxColumns < 0 || options.MaxPairChecks < 0 || options.MaxAdjacency < 0 || options.MaxVisits < 0 {
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
	if options.MaxAdjacency == 0 {
		options.MaxAdjacency = 16 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 16 << 20
	}
	return options, nil
}

func mvcCubesIntersect(first, second *MvcCube) bool {
	for index, word := range first.words {
		if word&second.words[index] != 0 {
			return true
		}
	}
	return false
}

func mvcIndependentRowWeight(cube *MvcCube, weights []int) (int, error) {
	minimum, found := math.MaxInt, false
	for wordIndex, word := range cube.words {
		for word != 0 {
			bit := bits.TrailingZeros32(word)
			column := wordIndex*32 + bit
			weight := 1
			if weights != nil {
				weight = weights[column]
			}
			if weight < 0 {
				return 0, ErrDimension
			}
			minimum = min(minimum, weight)
			found = true
			word &^= uint32(1) << uint(bit)
		}
	}
	if !found {
		return 0, ErrInvariant
	}
	return minimum, nil
}

// MaximalIndependentRows implements the source's greedy row-intersection
// lower bound. Rows is sorted as in the native sparse solution row;
// SelectionOrder records the otherwise implicit greedy sequence.
func (cover *MvcCover) MaximalIndependentRows(weights []int, options MvcIndependentRowsOptions) (MvcIndependentRows, error) {
	options, err := normalizeMvcIndependentRowsOptions(options)
	if err != nil {
		return MvcIndependentRows{}, err
	}
	if err = mvcCheckOwnedCubes(cover); err != nil {
		return MvcIndependentRows{}, err
	}
	if len(cover.cubes) > options.MaxRows || cover.bits > options.MaxColumns {
		return MvcIndependentRows{}, ErrLimit
	}
	if weights != nil && len(weights) != cover.bits {
		return MvcIndependentRows{}, ErrDimension
	}
	rowWeights := make([]int, len(cover.cubes))
	for row, cube := range cover.cubes {
		if options.Stop != nil && options.Stop() {
			return MvcIndependentRows{}, ErrCancelled
		}
		rowWeights[row], err = mvcIndependentRowWeight(cube, weights)
		if err != nil {
			return MvcIndependentRows{}, err
		}
	}
	adjacency := make([][]int, len(cover.cubes))
	entries, checks := 0, 0
	for first := range cover.cubes {
		adjacency[first] = append(adjacency[first], first)
		entries++
		if entries > options.MaxAdjacency {
			return MvcIndependentRows{}, ErrLimit
		}
		for second := first + 1; second < len(cover.cubes); second++ {
			checks++
			if checks > options.MaxPairChecks {
				return MvcIndependentRows{}, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return MvcIndependentRows{}, ErrCancelled
			}
			if !mvcCubesIntersect(cover.cubes[first], cover.cubes[second]) {
				continue
			}
			entries += 2
			if entries > options.MaxAdjacency {
				return MvcIndependentRows{}, ErrLimit
			}
			adjacency[first] = append(adjacency[first], second)
			adjacency[second] = append(adjacency[second], first)
		}
	}
	active := make([]bool, len(cover.cubes))
	for row := range active {
		active[row] = true
	}
	result := MvcIndependentRows{}
	remaining, visits := len(active), 0
	for remaining != 0 {
		best, bestDegree := -1, 0
		for row, yes := range active {
			if !yes {
				continue
			}
			degree := 0
			for _, neighbor := range adjacency[row] {
				visits++
				if visits > options.MaxVisits {
					return MvcIndependentRows{}, ErrLimit
				}
				if options.Stop != nil && options.Stop() {
					return MvcIndependentRows{}, ErrCancelled
				}
				if active[neighbor] {
					degree++
				}
			}
			if best < 0 || degree < bestDegree {
				best, bestDegree = row, degree
			}
		}
		if best < 0 {
			return MvcIndependentRows{}, ErrInvariant
		}
		if rowWeights[best] > math.MaxInt-result.Cost {
			return MvcIndependentRows{}, ErrLimit
		}
		result.Cost += rowWeights[best]
		result.SelectionOrder = append(result.SelectionOrder, best)
		for _, neighbor := range adjacency[best] {
			visits++
			if visits > options.MaxVisits {
				return MvcIndependentRows{}, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return MvcIndependentRows{}, ErrCancelled
			}
			if active[neighbor] {
				active[neighbor] = false
				remaining--
			}
		}
		if options.Stop != nil && options.Stop() {
			return MvcIndependentRows{}, ErrCancelled
		}
	}
	result.Rows = slices.Clone(result.SelectionOrder)
	sort.Ints(result.Rows)
	return result, nil
}
