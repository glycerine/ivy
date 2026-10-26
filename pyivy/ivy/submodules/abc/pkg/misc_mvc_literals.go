package abc

// Literal counting and selection from src/misc/mvc/mvcLits.c.

func (cover *MvcCover) checkLiteralMask(mask *MvcCube) error {
	if err := cover.check(); err != nil {
		return err
	}
	if mask == nil {
		return nil
	}
	if err := mask.check(); err != nil {
		return err
	}
	if mask.cover.bits != cover.bits || len(mask.words) != cover.words {
		return ErrDimension
	}
	return nil
}

// LiteralFrequencies returns an owned bit-indexed occurrence count.
func (cover *MvcCover) LiteralFrequencies() ([]int, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	frequencies := make([]int, cover.bits)
	for _, cube := range cover.cubes {
		for literal := 0; literal < cover.bits; literal++ {
			if cube.words[literal>>5]&(uint32(1)<<uint(literal&31)) != 0 {
				frequencies[literal]++
			}
		}
	}
	return frequencies, nil
}

// AnyRepeatedLiteral returns the highest selected literal occurring in at
// least two cubes, or -1 when none exists.
func (cover *MvcCover) AnyRepeatedLiteral(mask *MvcCube) (int, error) {
	if err := cover.checkLiteralMask(mask); err != nil {
		return -1, err
	}
	frequencies, err := cover.LiteralFrequencies()
	if err != nil {
		return -1, err
	}
	for literal := cover.bits - 1; literal >= 0; literal-- {
		if mask != nil && mask.words[literal>>5]&(uint32(1)<<uint(literal&31)) == 0 {
			continue
		}
		if frequencies[literal] > 1 {
			return literal, nil
		}
	}
	return -1, nil
}

// BestRepeatedLiteral returns the lowest-indexed selected literal having the
// greatest occurrence count, provided that count is at least two.
func (cover *MvcCover) BestRepeatedLiteral(mask *MvcCube) (int, error) {
	if err := cover.checkLiteralMask(mask); err != nil {
		return -1, err
	}
	frequencies, err := cover.LiteralFrequencies()
	if err != nil {
		return -1, err
	}
	best, bestCount := -1, -1
	for literal, count := range frequencies {
		if mask != nil && mask.words[literal>>5]&(uint32(1)<<uint(literal&31)) == 0 {
			continue
		}
		if bestCount < count {
			best, bestCount = literal, count
		}
	}
	if bestCount > 1 {
		return best, nil
	}
	return -1, nil
}

// WorstRepeatedLiteral returns the lowest-indexed selected literal having the
// smallest occurrence count of at least two.
func (cover *MvcCover) WorstRepeatedLiteral(mask *MvcCube) (int, error) {
	if err := cover.checkLiteralMask(mask); err != nil {
		return -1, err
	}
	frequencies, err := cover.LiteralFrequencies()
	if err != nil {
		return -1, err
	}
	worst, worstCount := -1, int(^uint(0)>>1)
	for literal, count := range frequencies {
		if mask != nil && mask.words[literal>>5]&(uint32(1)<<uint(literal&31)) == 0 {
			continue
		}
		if count >= 2 && count < worstCount {
			worst, worstCount = literal, count
		}
	}
	return worst, nil
}

// BestLiteralCover returns a one-cube cover containing the best repeated
// literal selected by simple's only cube.
func (cover *MvcCover) BestLiteralCover(simple *MvcCover) (*MvcCover, error) {
	if err := cover.check(); err != nil {
		return nil, err
	}
	if err := simple.check(); err != nil {
		return nil, err
	}
	if cover.bits != simple.bits {
		return nil, ErrDimension
	}
	if len(simple.cubes) != 1 {
		return nil, ErrInvariant
	}
	literal, err := cover.BestRepeatedLiteral(simple.cubes[0])
	if err != nil {
		return nil, err
	}
	if literal < 0 {
		return nil, ErrInvariant
	}
	result, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	cube, err := result.NewCube()
	if err != nil {
		_ = result.Close()
		return nil, err
	}
	if err = cube.SetBit(literal, true); err != nil {
		_ = cube.Close()
		_ = result.Close()
		return nil, err
	}
	if err = result.AddTail(cube); err != nil {
		_ = cube.Close()
		_ = result.Close()
		return nil, err
	}
	return result, nil
}

// FirstCubeFirstLiteral returns the lowest asserted literal in the first cube,
// or -1 for an empty cover or an all-zero first cube.
func (cover *MvcCover) FirstCubeFirstLiteral() (int, error) {
	if err := cover.check(); err != nil {
		return -1, err
	}
	if len(cover.cubes) == 0 {
		return -1, nil
	}
	for literal := 0; literal < cover.bits; literal++ {
		if cover.cubes[0].words[literal>>5]&(uint32(1)<<uint(literal&31)) != 0 {
			return literal, nil
		}
	}
	return -1, nil
}

// LiteralCount returns the total number of asserted bits in all cubes.
func (cover *MvcCover) LiteralCount() (int, error) {
	if err := cover.check(); err != nil {
		return 0, err
	}
	total := 0
	for _, cube := range cover.cubes {
		total += mvcCubePopulation(cube)
	}
	return total, nil
}

// IsOneLiteral preserves Mvc_CoverIsOneLiteral: exactly one cube must exist,
// and that cube may contain zero or one asserted literal.
func (cover *MvcCover) IsOneLiteral() (bool, error) {
	if err := cover.check(); err != nil {
		return false, err
	}
	if len(cover.cubes) != 1 {
		return false, nil
	}
	return mvcCubePopulation(cover.cubes[0]) <= 1, nil
}
