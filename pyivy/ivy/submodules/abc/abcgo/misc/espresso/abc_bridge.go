package espresso

// AbcMinimize matches ABC's standard single-output Espresso setup with an
// empty don't-care set and an offset derived from the onset.
func AbcMinimize(onset, dontCare *Family) *Family {
	if onset == nil {
		return nil
	}
	newDC := dontCare == nil
	if newDC {
		dontCare = sf_new(1, cube.size)
	}
	dontCare.wsize, dontCare.sf_size = onset.wsize, onset.sf_size
	var offset pcover
	if dontCare.sf_size == 0 || dontCare.count == 0 {
		offset = complement(cube1list(onset))
	} else {
		offset = complement(cube2list(onset, dontCare))
	}
	skip_make_sparse = true
	onset = espresso(onset, dontCare, offset)
	sf_free(offset)
	if newDC {
		sf_free(dontCare)
	}
	return onset
}
