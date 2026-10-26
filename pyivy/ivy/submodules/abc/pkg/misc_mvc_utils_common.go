package abc

// Common-cube and support utilities from src/misc/mvc/mvcUtils.c.

func (cover *MvcCover) intersectionCube() (*MvcCube, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	result, err := cover.NewCube()
	if err != nil {
		return nil, err
	}
	for index := range result.words {
		result.words[index] = ^uint32(0)
	}
	result.cleanUnused()
	for _, cube := range cover.cubes {
		for index := range result.words {
			result.words[index] &= cube.words[index]
		}
	}
	return result, nil
}

// SupportIntersection implements Mvc_CoverSupportAnd and returns an owned cube
// containing bits asserted by every cube. An empty cover yields all used bits.
func (cover *MvcCover) SupportIntersection() (*MvcCube, error) {
	return cover.intersectionCube()
}

// CommonCube implements Mvc_CoverCommonCube. Native SupportAnd and CommonCube
// are behaviorally identical but remain named separately in the Go surface.
func (cover *MvcCover) CommonCube() (*MvcCube, error) {
	return cover.intersectionCube()
}

// BinarySupportSize counts paired binary variables constrained by at least one
// cover cube. An odd final bit is ignored, matching the native routine.
func (cover *MvcCover) BinarySupportSize() (int, error) {
	common, err := cover.intersectionCube()
	if err != nil {
		return 0, err
	}
	count := cover.bits / 2
	for variable := 0; variable < cover.bits/2; variable++ {
		bit := variable * 2
		if (common.words[bit>>5]>>uint(bit&31))&3 == 3 {
			count--
		}
	}
	if err = common.Close(); err != nil {
		return 0, err
	}
	return count, nil
}

// BinarySupportHasVariable reports whether a paired binary variable is
// constrained by at least one cover cube.
func (cover *MvcCover) BinarySupportHasVariable(variable int) (bool, error) {
	if variable < 0 || variable >= cover.BitCount()/2 {
		return false, ErrDimension
	}
	common, err := cover.intersectionCube()
	if err != nil {
		return false, err
	}
	bit := variable * 2
	belongs := (common.words[bit>>5]>>uint(bit&31))&3 != 3
	if err = common.Close(); err != nil {
		return false, err
	}
	return belongs, nil
}

// IsCubeFree reports whether the intersection of all cubes is empty.
func (cover *MvcCover) IsCubeFree() (bool, error) {
	common, err := cover.intersectionCube()
	if err != nil {
		return false, err
	}
	empty := true
	for _, word := range common.words {
		if word != 0 {
			empty = false
			break
		}
	}
	if err = common.Close(); err != nil {
		return false, err
	}
	return empty, nil
}

// MakeCubeFree removes every common bit from every cube in place.
func (cover *MvcCover) MakeCubeFree() error {
	common, err := cover.intersectionCube()
	if err != nil {
		return err
	}
	for _, cube := range cover.cubes {
		for index := range cube.words {
			cube.words[index] &^= common.words[index]
		}
	}
	return common.Close()
}

// CommonCubeCover returns a new one-cube cover holding the common cube.
func (cover *MvcCover) CommonCubeCover() (*MvcCover, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	result, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	common, err := cover.intersectionCube()
	if err != nil {
		_ = result.Close()
		return nil, err
	}
	copyCube, err := result.cloneCube(common)
	closeErr := common.Close()
	if err != nil {
		_ = result.Close()
		return nil, err
	}
	if closeErr != nil {
		_ = copyCube.Close()
		_ = result.Close()
		return nil, closeErr
	}
	if err = result.AddTail(copyCube); err != nil {
		_ = copyCube.Close()
		_ = result.Close()
		return nil, err
	}
	return result, nil
}

// SupportContains reports whether cover's union support contains other's.
func (cover *MvcCover) SupportContains(other *MvcCover) (bool, error) {
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
	contains := true
	for index := range first.words {
		if second.words[index]&^first.words[index] != 0 {
			contains = false
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
	return contains, nil
}
