package abc

// Algebraic cover operations from src/misc/mvc/mvcOpAlg.c and support from mvcUtils.c.

import "slices"

func mvcCheckOwnedCubes(cover *MvcCover) error {
	if err := cover.check(); err != nil {
		return err
	}
	for _, cube := range cover.cubes {
		if err := cube.check(); err != nil {
			return err
		}
		if cube.cover != cover || !cube.listed {
			return ErrInvariant
		}
	}
	return nil
}

func mvcCheckCoverPair(first, second *MvcCover) error {
	if err := mvcCheckOwnedCubes(first); err != nil {
		return err
	}
	if err := mvcCheckOwnedCubes(second); err != nil {
		return err
	}
	if first.bits != second.bits || first.words != second.words {
		return ErrDimension
	}
	return nil
}

// Support returns an owned cube containing the union of every cover cube.
func (cover *MvcCover) Support() (*MvcCube, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	support, err := cover.NewCube()
	if err != nil {
		return nil, err
	}
	for _, cube := range cover.cubes {
		for index := range support.words {
			support.words[index] |= cube.words[index]
		}
	}
	return support, nil
}

// SupportsDisjoint reports whether the two covers use no common bit.
func (cover *MvcCover) SupportsDisjoint(other *MvcCover) (bool, error) {
	if err := mvcCheckCoverPair(cover, other); err != nil {
		return false, err
	}
	first, err := cover.Support()
	if err != nil {
		return false, err
	}
	second, err := other.Support()
	if err != nil {
		_ = first.Close()
		return false, err
	}
	disjoint := true
	for index := range first.words {
		if first.words[index]&second.words[index] != 0 {
			disjoint = false
			break
		}
	}
	if err = second.Close(); err != nil {
		_ = first.Close()
		return false, err
	}
	if err = first.Close(); err != nil {
		return false, err
	}
	return disjoint, nil
}

// AlgebraicMultiply returns the Cartesian pairwise OR in left-major order.
// Native ABC only warns when supports overlap, so overlap remains accepted.
func (cover *MvcCover) AlgebraicMultiply(other *MvcCover) (*MvcCover, error) {
	if err := mvcCheckCoverPair(cover, other); err != nil {
		return nil, err
	}
	result, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	for _, first := range cover.cubes {
		for _, second := range other.cubes {
			cube, cubeErr := result.NewCube()
			if cubeErr != nil {
				_ = result.Close()
				return nil, cubeErr
			}
			for index := range cube.words {
				cube.words[index] = first.words[index] | second.words[index]
			}
			if cubeErr = result.AddTail(cube); cubeErr != nil {
				_ = cube.Close()
				_ = result.Close()
				return nil, cubeErr
			}
		}
	}
	return result, nil
}

// AlgebraicSubtract removes every left cube whose exact value occurs at least
// once in other, preserving order and multiplicity of all unmatched cubes.
func (cover *MvcCover) AlgebraicSubtract(other *MvcCover) (*MvcCover, error) {
	if err := mvcCheckCoverPair(cover, other); err != nil {
		return nil, err
	}
	result, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	for _, first := range cover.cubes {
		found := false
		for _, second := range other.cubes {
			if slices.Equal(first.words, second.words) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		cube, cubeErr := result.cloneCube(first)
		if cubeErr != nil {
			_ = result.Close()
			return nil, cubeErr
		}
		if cubeErr = result.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			_ = result.Close()
			return nil, cubeErr
		}
	}
	return result, nil
}

// AlgebraicEqual preserves Mvc_CoverAlgebraicEqual's directional semantics:
// every cube in cover must occur in other; extra cubes and multiplicity do not
// affect the result.
func (cover *MvcCover) AlgebraicEqual(other *MvcCover) (bool, error) {
	if err := mvcCheckCoverPair(cover, other); err != nil {
		return false, err
	}
	for _, first := range cover.cubes {
		found := false
		for _, second := range other.cubes {
			if slices.Equal(first.words, second.words) {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	return true, nil
}
