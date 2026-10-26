package abc

// Quick-divisor extraction from src/misc/mvc/mvcDivisor.c.

// Divisor returns an independently owned level-zero quick divisor, or nil
// when the cover has at most one cube or no repeated literal.
func (cover *MvcCover) Divisor() (*MvcCover, error) {
	if err := mvcCheckOwnedCubes(cover); err != nil {
		return nil, err
	}
	if len(cover.cubes) <= 1 {
		return nil, nil
	}
	literal, err := cover.AnyRepeatedLiteral(nil)
	if err != nil {
		return nil, err
	}
	if literal < 0 {
		return nil, nil
	}
	kernel, err := cover.Duplicate()
	if err != nil {
		return nil, err
	}
	for {
		literal, err = kernel.WorstRepeatedLiteral(nil)
		if err != nil {
			_ = kernel.Close()
			return nil, err
		}
		if literal < 0 {
			return kernel, nil
		}
		if err = kernel.DivideByLiteralQuotient(literal); err != nil {
			_ = kernel.Close()
			return nil, err
		}
		if err = kernel.MakeCubeFree(); err != nil {
			_ = kernel.Close()
			return nil, err
		}
	}
}
