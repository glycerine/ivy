package abc

// Algebraic cover division from src/misc/mvc/mvcDivide.c.

func mvcNewDivisionResults(cover *MvcCover) (*MvcCover, *MvcCover, error) {
	quotient, err := cover.Clone()
	if err != nil {
		return nil, nil, err
	}
	remainder, err := cover.Clone()
	if err != nil {
		_ = quotient.Close()
		return nil, nil, err
	}
	return quotient, remainder, nil
}

func mvcCloseDivisionResults(quotient, remainder *MvcCover) {
	if remainder != nil {
		_ = remainder.Close()
	}
	if quotient != nil {
		_ = quotient.Close()
	}
}

func mvcAppendClone(destination *MvcCover, source *MvcCube) error {
	clone, err := destination.cloneCube(source)
	if err != nil {
		return err
	}
	if err = destination.AddTail(clone); err != nil {
		_ = clone.Close()
		return err
	}
	return nil
}

// Divide performs native algebraic division. A non-dividable pair returns
// nil, nil, nil. Multi-cube division preserves native behavior by sorting both
// input covers in place before producing ordered quotient and remainder covers.
func (cover *MvcCover) Divide(divisor *MvcCover) (*MvcCover, *MvcCover, error) {
	if err := mvcCheckCoverPair(cover, divisor); err != nil {
		return nil, nil, err
	}
	if len(divisor.cubes) == 0 {
		return nil, nil, ErrInvariant
	}
	if len(cover.cubes) < len(divisor.cubes) {
		return nil, nil, nil
	}
	contains, err := cover.SupportContains(divisor)
	if err != nil {
		return nil, nil, err
	}
	if !contains {
		return nil, nil, nil
	}
	return cover.DivideInternal(divisor)
}

// DivideInternal performs division without the outer count/support rejection.
func (cover *MvcCover) DivideInternal(divisor *MvcCover) (*MvcCover, *MvcCover, error) {
	if err := mvcCheckCoverPair(cover, divisor); err != nil {
		return nil, nil, err
	}
	if len(divisor.cubes) == 0 {
		return nil, nil, ErrInvariant
	}
	if len(divisor.cubes) == 1 {
		oneLiteral, err := divisor.IsOneLiteral()
		if err != nil {
			return nil, nil, err
		}
		if oneLiteral {
			return cover.DivideByLiteral(divisor)
		}
		return cover.DivideByCube(divisor)
	}

	mask, err := divisor.Support()
	if err != nil {
		return nil, nil, err
	}
	if err = divisor.Sort(MvcOrderInteger, nil); err != nil {
		_ = mask.Close()
		return nil, nil, err
	}
	if err = cover.Sort(MvcOrderOutsideAndUnderMask, mask); err != nil {
		_ = mask.Close()
		return nil, nil, err
	}
	quotient, remainder, err := mvcNewDivisionResults(cover)
	if err != nil {
		_ = mask.Close()
		return nil, nil, err
	}

	groups := []int{0}
	for index := 1; index < len(cover.cubes); index++ {
		equal := true
		for word := range mask.words {
			if cover.cubes[index-1].words[word]&^mask.words[word] != cover.cubes[index].words[word]&^mask.words[word] {
				equal = false
				break
			}
		}
		if !equal {
			groups = append(groups, index)
		}
	}
	groups = append(groups, len(cover.cubes))

	for group := 0; group+1 < len(groups); group++ {
		start, end := groups[group], groups[group+1]
		groupSize := end - start
		if groupSize < len(divisor.cubes) {
			for _, cube := range cover.cubes[start:end] {
				if err = mvcAppendClone(remainder, cube); err != nil {
					mvcCloseDivisionResults(quotient, remainder)
					_ = mask.Close()
					return nil, nil, err
				}
			}
			continue
		}

		matched := make([]bool, groupSize)
		coverIndex, divisorIndex := 0, 0
		lastMatch := -1
		failed := false
		for {
			comparison, compareErr := cover.cubes[start+coverIndex].CompareIntUnderMask(divisor.cubes[divisorIndex], mask)
			if compareErr != nil {
				mvcCloseDivisionResults(quotient, remainder)
				_ = mask.Close()
				return nil, nil, compareErr
			}
			if comparison < 0 {
				coverIndex++
				if groupSize-coverIndex < len(divisor.cubes)-divisorIndex {
					failed = true
					break
				}
				continue
			}
			if comparison > 0 {
				failed = true
				break
			}
			matched[coverIndex] = true
			lastMatch = coverIndex
			divisorIndex++
			if divisorIndex == len(divisor.cubes) {
				break
			}
			coverIndex++
		}

		if failed {
			for _, cube := range cover.cubes[start:end] {
				if err = mvcAppendClone(remainder, cube); err != nil {
					mvcCloseDivisionResults(quotient, remainder)
					_ = mask.Close()
					return nil, nil, err
				}
			}
			continue
		}
		for offset, cube := range cover.cubes[start:end] {
			if !matched[offset] {
				if err = mvcAppendClone(remainder, cube); err != nil {
					mvcCloseDivisionResults(quotient, remainder)
					_ = mask.Close()
					return nil, nil, err
				}
			}
		}
		cube, cubeErr := quotient.NewCube()
		if cubeErr != nil {
			mvcCloseDivisionResults(quotient, remainder)
			_ = mask.Close()
			return nil, nil, cubeErr
		}
		for word := range cube.words {
			cube.words[word] = cover.cubes[start+lastMatch].words[word] &^ mask.words[word]
		}
		if cubeErr = quotient.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			mvcCloseDivisionResults(quotient, remainder)
			_ = mask.Close()
			return nil, nil, cubeErr
		}
	}
	if err = mask.Close(); err != nil {
		mvcCloseDivisionResults(quotient, remainder)
		return nil, nil, err
	}
	return quotient, remainder, nil
}

// DivideByCube divides by a one-cube divisor using full-width containment.
func (cover *MvcCover) DivideByCube(divisor *MvcCover) (*MvcCover, *MvcCover, error) {
	if err := mvcCheckCoverPair(cover, divisor); err != nil {
		return nil, nil, err
	}
	if len(divisor.cubes) != 1 {
		return nil, nil, ErrInvariant
	}
	quotient, remainder, err := mvcNewDivisionResults(cover)
	if err != nil {
		return nil, nil, err
	}
	divisorCube := divisor.cubes[0]
	for _, source := range cover.cubes {
		contains := true
		for word := range source.words {
			if divisorCube.words[word]&^source.words[word] != 0 {
				contains = false
				break
			}
		}
		if !contains {
			if err = mvcAppendClone(remainder, source); err != nil {
				mvcCloseDivisionResults(quotient, remainder)
				return nil, nil, err
			}
			continue
		}
		cube, cubeErr := quotient.NewCube()
		if cubeErr != nil {
			mvcCloseDivisionResults(quotient, remainder)
			return nil, nil, cubeErr
		}
		for word := range cube.words {
			cube.words[word] = source.words[word] &^ divisorCube.words[word]
		}
		if cubeErr = quotient.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			mvcCloseDivisionResults(quotient, remainder)
			return nil, nil, cubeErr
		}
	}
	return quotient, remainder, nil
}

// DivideByLiteral divides by a one-cube, exactly-one-literal divisor.
func (cover *MvcCover) DivideByLiteral(divisor *MvcCover) (*MvcCover, *MvcCover, error) {
	if err := mvcCheckCoverPair(cover, divisor); err != nil {
		return nil, nil, err
	}
	if len(divisor.cubes) != 1 || mvcCubePopulation(divisor.cubes[0]) != 1 {
		return nil, nil, ErrInvariant
	}
	literal, err := divisor.FirstCubeFirstLiteral()
	if err != nil || literal < 0 {
		if err == nil {
			err = ErrInvariant
		}
		return nil, nil, err
	}
	quotient, remainder, err := mvcNewDivisionResults(cover)
	if err != nil {
		return nil, nil, err
	}
	for _, source := range cover.cubes {
		destination := remainder
		if source.words[literal>>5]&(uint32(1)<<uint(literal&31)) != 0 {
			destination = quotient
		}
		cube, cubeErr := destination.cloneCube(source)
		if cubeErr != nil {
			mvcCloseDivisionResults(quotient, remainder)
			return nil, nil, cubeErr
		}
		if destination == quotient {
			cube.words[literal>>5] &^= uint32(1) << uint(literal&31)
		}
		if cubeErr = destination.AddTail(cube); cubeErr != nil {
			_ = cube.Close()
			mvcCloseDivisionResults(quotient, remainder)
			return nil, nil, cubeErr
		}
	}
	return quotient, remainder, nil
}

// DivideByLiteralQuotient reduces cover in place to its quotient by literal.
func (cover *MvcCover) DivideByLiteralQuotient(literal int) error {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return err
	}
	if literal < 0 || literal >= cover.bits {
		return ErrDimension
	}
	survivors := make([]*MvcCube, 0, len(cover.cubes))
	removed := make([]*MvcCube, 0, len(cover.cubes))
	mask := uint32(1) << uint(literal&31)
	for _, cube := range cover.cubes {
		if cube.words[literal>>5]&mask == 0 {
			removed = append(removed, cube)
		} else {
			survivors = append(survivors, cube)
		}
	}
	for _, cube := range survivors {
		cube.words[literal>>5] &^= mask
	}
	for _, cube := range removed {
		cube.listed = false
		if err := cube.Close(); err != nil {
			return err
		}
	}
	cover.cubes = survivors
	return nil
}
