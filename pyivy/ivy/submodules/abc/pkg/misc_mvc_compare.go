package abc

// Cube comparisons from src/misc/mvc/mvcCompare.c.

import "math/bits"

func mvcCheckComparable(first, second *MvcCube) error {
	if err := first.check(); err != nil {
		return err
	}
	if err := second.check(); err != nil {
		return err
	}
	if first.cover.bits != second.cover.bits || len(first.words) != len(second.words) {
		return ErrDimension
	}
	return nil
}

func mvcCheckComparisonMask(first, second, mask *MvcCube) error {
	if err := mvcCheckComparable(first, second); err != nil {
		return err
	}
	if err := mask.check(); err != nil {
		return err
	}
	if first.cover.bits != mask.cover.bits || len(first.words) != len(mask.words) {
		return ErrDimension
	}
	return nil
}

func mvcCompareWords(first, second []uint32) int {
	for index := len(first) - 1; index >= 0; index-- {
		if first[index] < second[index] {
			return -1
		}
		if first[index] > second[index] {
			return 1
		}
	}
	return 0
}

// OneCount returns the number of asserted logical bits in cube.
func (cube *MvcCube) OneCount() (int, error) {
	if err := cube.check(); err != nil {
		return 0, err
	}
	count := 0
	for _, word := range cube.words {
		count += bits.OnesCount32(word)
	}
	return count, nil
}

// CompareInt compares packed cubes as unsigned multiword integers, with the
// highest-indexed word most significant.
func (cube *MvcCube) CompareInt(other *MvcCube) (int, error) {
	if err := mvcCheckComparable(cube, other); err != nil {
		return 0, err
	}
	return mvcCompareWords(cube.words, other.words), nil
}

// CompareSizeAndInt orders larger-population cubes first and uses CompareInt
// to break equal-population ties.
func (cube *MvcCube) CompareSizeAndInt(other *MvcCube) (int, error) {
	if err := mvcCheckComparable(cube, other); err != nil {
		return 0, err
	}
	firstCount, secondCount := 0, 0
	for index := range cube.words {
		firstCount += bits.OnesCount32(cube.words[index])
		secondCount += bits.OnesCount32(other.words[index])
	}
	if firstCount < secondCount {
		return 1, nil
	}
	if firstCount > secondCount {
		return -1, nil
	}
	return mvcCompareWords(cube.words, other.words), nil
}

// CompareIntUnderMask compares only bits selected by mask.
func (cube *MvcCube) CompareIntUnderMask(other, mask *MvcCube) (int, error) {
	if err := mvcCheckComparisonMask(cube, other, mask); err != nil {
		return 0, err
	}
	for index := len(cube.words) - 1; index >= 0; index-- {
		first := cube.words[index] & mask.words[index]
		second := other.words[index] & mask.words[index]
		if first < second {
			return -1, nil
		}
		if first > second {
			return 1, nil
		}
	}
	return 0, nil
}

// CompareIntOutsideMask compares only bits not selected by mask. It preserves
// the native OR-with-mask spelling, which is order-equivalent to clearing it.
func (cube *MvcCube) CompareIntOutsideMask(other, mask *MvcCube) (int, error) {
	if err := mvcCheckComparisonMask(cube, other, mask); err != nil {
		return 0, err
	}
	for index := len(cube.words) - 1; index >= 0; index-- {
		first := cube.words[index] | mask.words[index]
		second := other.words[index] | mask.words[index]
		if first < second {
			return -1, nil
		}
		if first > second {
			return 1, nil
		}
	}
	return 0, nil
}

// CompareIntOutsideAndUnderMask first compares unselected bits and then uses
// selected bits as a tie breaker. Unlike the native assertion, equality is a
// valid comparator result and returns zero.
func (cube *MvcCube) CompareIntOutsideAndUnderMask(other, mask *MvcCube) (int, error) {
	if err := mvcCheckComparisonMask(cube, other, mask); err != nil {
		return 0, err
	}
	for index := len(cube.words) - 1; index >= 0; index-- {
		first := cube.words[index] &^ mask.words[index]
		second := other.words[index] &^ mask.words[index]
		if first < second {
			return -1, nil
		}
		if first > second {
			return 1, nil
		}
	}
	for index := len(cube.words) - 1; index >= 0; index-- {
		first := cube.words[index] & mask.words[index]
		second := other.words[index] & mask.words[index]
		if first < second {
			return -1, nil
		}
		if first > second {
			return 1, nil
		}
	}
	return 0, nil
}
