package abc

// Cover complementation and semantic simplification from src/misc/espresso/compl.c.

func (layout *EspressoCubeLayout) complementCover(cover *EspressoSetFamily, work *espressoSharpWork) (*EspressoSetFamily, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	if err := cover.check(); err != nil {
		return nil, err
	}
	if cover.bits != layout.bitCount {
		return nil, ErrDimension
	}
	if len(cover.sets) > work.options.MaxCubes {
		return nil, ErrLimit
	}
	full, _ := layout.FullCube()
	return layout.sharpCubeByCover(full, cover, work)
}

// ComplementCover returns an owned cover for every assignment not represented
// by cover. The sharp engine consolidates compl.c's recursive split, merge,
// distance-one and lifting helpers into one deterministic operation.
func (layout *EspressoCubeLayout) ComplementCover(cover *EspressoSetFamily, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, err
	}
	return layout.complementCover(cover, work)
}

// SimplifyAndComplementCover returns a deterministic equivalent cover and its
// complement. Complementing the exact complement performs the source
// simp_comp contract without exposing its representation-only lifting state.
func (layout *EspressoCubeLayout) SimplifyAndComplementCover(cover *EspressoSetFamily, options EspressoSharpOptions) (simplified, complement *EspressoSetFamily, err error) {
	work, err := newEspressoSharpWork(options)
	if err != nil {
		return nil, nil, err
	}
	complement, err = layout.complementCover(cover, work)
	if err != nil {
		return nil, nil, err
	}
	simplified, err = layout.complementCover(complement, work)
	if err != nil {
		return nil, nil, err
	}
	return simplified, complement, nil
}

// SimplifyCover returns an owned semantic simplification of cover.
func (layout *EspressoCubeLayout) SimplifyCover(cover *EspressoSetFamily, options EspressoSharpOptions) (*EspressoSetFamily, error) {
	simplified, _, err := layout.SimplifyAndComplementCover(cover, options)
	return simplified, err
}
