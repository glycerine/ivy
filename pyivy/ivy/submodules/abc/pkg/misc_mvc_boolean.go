package abc

// Active Boolean cover operation from src/misc/mvc/mvcOpBool.c.

// BooleanOr returns an independent deep-copy concatenation of cover followed
// by other. Native Mvc_CoverBooleanOr does not perform containment or remove
// duplicate cubes.
func (cover *MvcCover) BooleanOr(other *MvcCover) (*MvcCover, error) {
	if err := mvcCheckCoverPair(cover, other); err != nil {
		return nil, err
	}
	result, err := cover.Clone()
	if err != nil {
		return nil, err
	}
	for _, source := range cover.cubes {
		if _, err = result.AddCloneTail(source); err != nil {
			_ = result.Close()
			return nil, err
		}
	}
	for _, source := range other.cubes {
		if _, err = result.AddCloneTail(source); err != nil {
			_ = result.Close()
			return nil, err
		}
	}
	return result, nil
}
