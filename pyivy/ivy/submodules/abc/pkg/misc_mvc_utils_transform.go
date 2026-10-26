package abc

// Cover transforms from src/misc/mvc/mvcUtils.c.

import "math/bits"

// CubePairDifferences returns XOR population for each i<j cube pair in list
// order. Counts retain native uint8 truncation.
func (cover *MvcCover) CubePairDifferences() ([]uint8, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	count := len(cover.cubes) * (len(cover.cubes) - 1) / 2
	differences := make([]uint8, 0, count)
	for first := 0; first < len(cover.cubes); first++ {
		for second := first + 1; second < len(cover.cubes); second++ {
			population := 0
			for word := range cover.cubes[first].words {
				population += bits.OnesCount32(cover.cubes[first].words[word] ^ cover.cubes[second].words[word])
			}
			differences = append(differences, uint8(population))
		}
	}
	return differences, nil
}

// Remap creates a cover whose output bit i copies mapping[i] from the source;
// -1 leaves that output column filled with ones.
func (cover *MvcCover) Remap(mapping []int) (*MvcCover, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	for _, source := range mapping {
		if source < -1 || source >= cover.bits {
			return nil, ErrDimension
		}
	}
	result, err := NewMvcCover(cover.manager, len(mapping))
	if err != nil {
		return nil, err
	}
	for _, sourceCube := range cover.cubes {
		cube, cubeErr := result.NewCube()
		if cubeErr != nil {
			_ = result.Close()
			return nil, cubeErr
		}
		for word := range cube.words {
			cube.words[word] = ^uint32(0)
		}
		cube.cleanUnused()
		for destination, source := range mapping {
			if source < 0 {
				continue
			}
			value := sourceCube.words[source>>5]&(uint32(1)<<uint(source&31)) != 0
			mask := uint32(1) << uint(destination&31)
			if value {
				cube.words[destination>>5] |= mask
			} else {
				cube.words[destination>>5] &^= mask
			}
		}
		if cubeErr = result.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = result.Close()
			return nil, cubeErr
		}
	}
	return result, nil
}

// Invert complements every used bit of every cube in place.
func (cover *MvcCover) Invert() error {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return err
	}
	for _, cube := range cover.cubes {
		for word := range cube.words {
			cube.words[word] = ^cube.words[word]
		}
		cube.cleanUnused()
	}
	return nil
}

// RemoveDontCareLiterals returns a deep copy with each 11 adjacent pair
// removed, preserving source order and storage.
func (cover *MvcCover) RemoveDontCareLiterals() (*MvcCover, error) {
	result, err := cover.Duplicate()
	if err != nil {
		return nil, err
	}
	for _, cube := range result.cubes {
		if err = cube.RemoveDontCares(); err != nil {
			_ = result.Close()
			return nil, err
		}
	}
	return result, nil
}

// Cofactor returns cubes containing value, with other inserted into each copy.
func (cover *MvcCover) Cofactor(value, other int) (*MvcCover, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	if value < 0 || value >= cover.bits || other < 0 || other >= cover.bits {
		return nil, ErrDimension
	}
	result, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	for _, source := range cover.cubes {
		if source.words[value>>5]&(uint32(1)<<uint(value&31)) == 0 {
			continue
		}
		cube, cubeErr := result.cloneCube(source)
		if cubeErr != nil {
			_ = result.Close()
			return nil, cubeErr
		}
		cube.words[other>>5] |= uint32(1) << uint(other&31)
		if cubeErr = result.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = result.Close()
			return nil, cubeErr
		}
	}
	return result, nil
}

// FlipVariable returns a deep copy with an adjacent binary value pair swapped.
func (cover *MvcCover) FlipVariable(value0, value1 int) (*MvcCover, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	if value0 < 0 || value1 != value0+1 || value1 >= cover.bits {
		return nil, ErrDimension
	}
	for _, cube := range cover.cubes {
		if cube.words[value0>>5]&(uint32(1)<<uint(value0&31)) == 0 && cube.words[value1>>5]&(uint32(1)<<uint(value1&31)) == 0 {
			return nil, ErrInvariant
		}
	}
	result, err := cover.Duplicate()
	if err != nil {
		return nil, err
	}
	for _, cube := range result.cubes {
		first := cube.words[value0>>5]&(uint32(1)<<uint(value0&31)) != 0
		second := cube.words[value1>>5]&(uint32(1)<<uint(value1&31)) != 0
		_ = cube.SetBit(value0, second)
		_ = cube.SetBit(value1, first)
	}
	return result, nil
}

// UniversalQuantify equates adjacent binary variables A and B, drops
// incompatible cubes, intersects their A values and makes B independent.
func (cover *MvcCover) UniversalQuantify(a0, a1, b0, b1 int) (*MvcCover, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	if a0 < 0 || a1 != a0+1 || b0 < 0 || b1 != b0+1 || a1 >= cover.bits || b1 >= cover.bits {
		return nil, ErrDimension
	}
	for _, cube := range cover.cubes {
		aValue0 := cube.words[a0>>5]&(uint32(1)<<uint(a0&31)) != 0
		aValue1 := cube.words[a1>>5]&(uint32(1)<<uint(a1&31)) != 0
		bValue0 := cube.words[b0>>5]&(uint32(1)<<uint(b0&31)) != 0
		bValue1 := cube.words[b1>>5]&(uint32(1)<<uint(b1&31)) != 0
		if !aValue0 && !aValue1 || !bValue0 && !bValue1 {
			return nil, ErrInvariant
		}
	}
	result, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	for _, source := range cover.cubes {
		aValue0 := source.words[a0>>5]&(uint32(1)<<uint(a0&31)) != 0
		aValue1 := source.words[a1>>5]&(uint32(1)<<uint(a1&31)) != 0
		bValue0 := source.words[b0>>5]&(uint32(1)<<uint(b0&31)) != 0
		bValue1 := source.words[b1>>5]&(uint32(1)<<uint(b1&31)) != 0
		if aValue0 != bValue0 && aValue1 != bValue1 {
			continue
		}
		cube, cubeErr := result.cloneCube(source)
		if cubeErr != nil {
			_ = result.Close()
			return nil, cubeErr
		}
		_ = cube.SetBit(a0, aValue0 && bValue0)
		_ = cube.SetBit(a1, aValue1 && bValue1)
		_ = cube.SetBit(b0, true)
		_ = cube.SetBit(b1, true)
		if cubeErr = result.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = result.Close()
			return nil, cubeErr
		}
	}
	return result, nil
}

// Transpose returns the bit-by-cube matrix transpose in native order.
func (cover *MvcCover) Transpose() (*MvcCover, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	result, err := NewMvcCover(cover.manager, len(cover.cubes))
	if err != nil {
		return nil, err
	}
	for bit := 0; bit < cover.bits; bit++ {
		cube, cubeErr := result.NewCube()
		if cubeErr != nil {
			_ = result.Close()
			return nil, cubeErr
		}
		for cubeIndex, source := range cover.cubes {
			if source.words[bit>>5]&(uint32(1)<<uint(bit&31)) != 0 {
				cube.words[cubeIndex>>5] |= uint32(1) << uint(cubeIndex&31)
			}
		}
		if cubeErr = result.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = result.Close()
			return nil, cubeErr
		}
	}
	return result, nil
}

// UnusedBitsZero reports whether every packed tail is canonical.
func (cover *MvcCover) UnusedBitsZero() (bool, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return false, err
	}
	if cover.unused == 0 || cover.words == 0 {
		return true, nil
	}
	dirtyMask := ^(^uint32(0) >> uint(cover.unused))
	for _, cube := range cover.cubes {
		if cube.words[len(cube.words)-1]&dirtyMask != 0 {
			return false, nil
		}
	}
	return true, nil
}
