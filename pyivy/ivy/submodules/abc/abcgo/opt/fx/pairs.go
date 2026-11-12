package fx

import (
	"fmt"
	"slices"
)

// CubePair identifies two cubes of the same owner that contain a divisor.
// Complement says that the normalized divisor's output phase is inverted.
type CubePair struct {
	First, Second int
	Complement    bool
}

// FindCommonPairs matches a divisor against two ordered lists of cube indices
// containing its respective pivot literals. It retains the source algorithm's
// Cartesian-product order within each owner group.
func FindCommonPairs(cubes [][]int32, first, second []int, divisor []int32) ([]CubePair, error) {
	owner := func(index int) (int32, error) {
		if index < 0 || index >= len(cubes) || len(cubes[index]) == 0 {
			return 0, fmt.Errorf("fx: invalid cube index %d", index)
		}
		return cubes[index][0], nil
	}
	var pairs []CubePair
	for i, j := 0, 0; i < len(first) && j < len(second); {
		ownerA, err := owner(first[i])
		if err != nil {
			return nil, err
		}
		ownerB, err := owner(second[j])
		if err != nil {
			return nil, err
		}
		if ownerA < ownerB {
			i++
			continue
		}
		if ownerA > ownerB {
			j++
			continue
		}
		endA, endB := i+1, j+1
		for endA < len(first) {
			next, problem := owner(first[endA])
			if problem != nil {
				return nil, problem
			}
			if next != ownerA {
				break
			}
			endA++
		}
		for endB < len(second) {
			next, problem := owner(second[endB])
			if problem != nil {
				return nil, problem
			}
			if next != ownerB {
				break
			}
			endB++
		}
		for _, a := range first[i:endA] {
			for _, b := range second[j:endB] {
				if a == b {
					continue
				}
				candidate, _, problem := CubeFreeDivisor(cubes[a], cubes[b])
				if problem != nil {
					return nil, problem
				}
				complement := len(candidate) == 4 && NormalizeDivisor(candidate) == 1
				if slices.Equal(candidate, divisor) {
					pairs = append(pairs, CubePair{a, b, complement})
				}
			}
		}
		i, j = endA, endB
	}
	return pairs, nil
}
