package fx

import (
	"fmt"
	"slices"
)

// ApplyDivisor rewrites the selected divisor's occurrences and appends its
// defining cubes. nVars is the first available node ID. Inputs are copied so
// that invalid or overlapping occurrences leave the caller's data unchanged.
func ApplyDivisor(cubes [][]int32, divisor []int32, nVars int32) ([][]int32, bool, error) {
	if nVars <= 0 || len(divisor) < 2 {
		return nil, false, fmt.Errorf("fx: invalid divisor or node count")
	}
	pivotA, pivotB := DivisorPivots(divisor)
	if pivotA < 0 || pivotB < 0 {
		return nil, false, fmt.Errorf("fx: divisor has no two pivots")
	}
	has := func(cube []int32, lit int32) bool {
		_, ok := slices.BinarySearch(cube[1:], lit)
		return ok
	}
	var singles, partA, partB []int
	for i, cube := range cubes {
		if len(cube) == 0 {
			return nil, false, fmt.Errorf("fx: empty cube")
		}
		if len(divisor) == 2 && has(cube, pivotA^1) && has(cube, pivotB^1) {
			singles = append(singles, i)
			continue
		}
		if has(cube, pivotA) {
			partA = append(partA, i)
		}
		if has(cube, pivotB) {
			partB = append(partB, i)
		}
	}
	pairs, err := FindCommonPairs(cubes, partA, partB, divisor)
	if err != nil {
		return nil, false, err
	}
	if len(singles) == 0 && len(pairs) == 0 {
		return nil, false, fmt.Errorf("fx: divisor has no occurrences")
	}
	used := make(map[int]bool)
	for _, pair := range pairs {
		if used[pair.First] || used[pair.Second] {
			return nil, false, fmt.Errorf("fx: overlapping divisor occurrences")
		}
		used[pair.First], used[pair.Second] = true, true
	}
	result := make([][]int32, len(cubes))
	for i := range cubes {
		result[i] = slices.Clone(cubes[i])
	}
	simplify := len(divisor) == 2 && pivotA>>1 == pivotB>>1
	if simplify && len(singles) != 0 {
		return nil, false, fmt.Errorf("fx: contradictory single cube")
	}
	if !simplify {
		if len(divisor) == 2 {
			result = append(result, []int32{nVars, pivotA ^ 1, pivotB ^ 1})
			slices.Sort(result[len(result)-1][1:])
		} else {
			a, b := AddDivisorLiterals([]int32{nVars}, []int32{nVars}, divisor)
			result = append(result, a, b)
		}
	}
	for _, i := range singles {
		var removed int
		result[i], removed = RemoveDivisorLiterals(result[i], divisor, false)
		if removed != 2 {
			return nil, false, fmt.Errorf("fx: invalid single-cube occurrence")
		}
		result[i] = append(result[i], nVars<<1)
	}
	for _, pair := range pairs {
		var countA, countB int
		result[pair.First], countA = RemoveDivisorLiterals(result[pair.First], divisor, pair.Complement)
		result[pair.Second], countB = RemoveDivisorLiterals(result[pair.Second], divisor, pair.Complement)
		if countA+countB < len(divisor) || countA+countB > len(divisor)+1 {
			return nil, false, fmt.Errorf("fx: invalid paired-cube occurrence")
		}
		if !simplify {
			literal := nVars << 1
			if len(divisor) == 2 || pair.Complement {
				literal |= 1
			}
			result[pair.First] = append(result[pair.First], literal)
		}
		result[pair.Second] = nil
	}
	result = slices.DeleteFunc(result, func(c []int32) bool { return len(c) == 0 })
	return result, !simplify, nil
}
