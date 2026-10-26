package abc

// Single-cube containment removal from src/misc/mvc/mvcContain.c.

import (
	"math/bits"
	"slices"
)

func mvcCubePopulation(cube *MvcCube) int {
	count := 0
	for _, word := range cube.words {
		count += bits.OnesCount32(word)
	}
	return count
}

func mvcCubeSubset(first, second *MvcCube) bool {
	for index := range first.words {
		if first.words[index]&^second.words[index] != 0 {
			return false
		}
	}
	return true
}

func (cover *MvcCover) removeAndClose(index int) error {
	cube := cover.cubes[index]
	copy(cover.cubes[index:], cover.cubes[index+1:])
	cover.cubes[len(cover.cubes)-1] = nil
	cover.cubes = cover.cubes[:len(cover.cubes)-1]
	cube.listed = false
	return cube.Close()
}

func (cover *MvcCover) removeDuplicateCubes() error {
	for index := 1; index < len(cover.cubes); {
		if slices.Equal(cover.cubes[index-1].words, cover.cubes[index].words) {
			if err := cover.removeAndClose(index); err != nil {
				return err
			}
			continue
		}
		index++
	}
	return nil
}

func (cover *MvcCover) removeContainedCubes() error {
	for groupStart := 0; groupStart < len(cover.cubes); {
		population := mvcCubePopulation(cover.cubes[groupStart])
		groupEnd := groupStart
		for groupEnd+1 < len(cover.cubes) && mvcCubePopulation(cover.cubes[groupEnd+1]) == population {
			groupEnd++
		}
		if groupEnd+1 == len(cover.cubes) {
			return nil
		}
		for large := groupStart; large <= groupEnd; large++ {
			for candidate := groupEnd + 1; candidate < len(cover.cubes); {
				if mvcCubeSubset(cover.cubes[candidate], cover.cubes[large]) {
					if err := cover.removeAndClose(candidate); err != nil {
						return err
					}
					continue
				}
				candidate++
			}
		}
		groupStart = groupEnd + 1
	}
	return nil
}

// RemoveContained removes duplicates and every cube contained by a larger
// cube. It returns whether the cover changed and leaves native sorted order.
func (cover *MvcCover) RemoveContained() (bool, error) {
	if err := cover.check(); err != nil {
		return false, err
	}
	original := len(cover.cubes)
	if original < 2 {
		return false, nil
	}
	if err := cover.Sort(MvcOrderSizeAndInteger, nil); err != nil {
		return false, err
	}
	if err := cover.removeDuplicateCubes(); err != nil {
		return false, err
	}
	if len(cover.cubes) > 1 {
		if err := cover.removeContainedCubes(); err != nil {
			return false, err
		}
	}
	return len(cover.cubes) != original, nil
}
