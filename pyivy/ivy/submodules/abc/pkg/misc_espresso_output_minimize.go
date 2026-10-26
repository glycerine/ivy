package abc

// Per-output minimization from src/misc/espresso/cvrm.c.

type EspressoOutputMinimizeStrategy uint8

const (
	EspressoOutputHeuristic EspressoOutputMinimizeStrategy = iota
	EspressoOutputExact
)

type EspressoOutputMinimizeOptions struct {
	Strategy       EspressoOutputMinimizeStrategy
	ChoosePolarity bool
	Manipulation   EspressoCoverManipulationOptions
	Heuristic      EspressoMinimizeOptions
	Exact          EspressoExactOptions
}

type EspressoOutputMinimizeResult struct {
	PLA           *EspressoPLA
	PositivePhase []bool
	CubeCounts    []int
}

func (layout *EspressoCubeLayout) minimizeOneOutput(on, dontCare, off *EspressoSetFamily, options EspressoOutputMinimizeOptions) (*EspressoSetFamily, error) {
	switch options.Strategy {
	case EspressoOutputHeuristic:
		heuristic := options.Heuristic
		heuristic.SkipMakeSparse = true
		result, err := layout.MinimizeEspressoCover(on, dontCare, off, heuristic)
		if err != nil {
			return nil, err
		}
		return result.Cover, nil
	case EspressoOutputExact:
		result, err := layout.MinimizeExact(on, dontCare, off, true, options.Exact)
		if err != nil {
			return nil, err
		}
		return result.Cover, nil
	default:
		return nil, ErrDimension
	}
}

// MinimizeOutputs independently minimizes every output function. When
// ChoosePolarity is set, both polarities are minimized and the smaller cover
// is retained; ties preserve positive polarity. The returned phase slice is
// true for positive and false for complemented outputs.
func (pla *EspressoPLA) MinimizeOutputs(options EspressoOutputMinimizeOptions) (EspressoOutputMinimizeResult, error) {
	if err := pla.check(); err != nil {
		return EspressoOutputMinimizeResult{}, err
	}
	if options.Strategy > EspressoOutputExact {
		return EspressoOutputMinimizeResult{}, ErrDimension
	}
	manipulation, err := normalizeEspressoCoverManipulationOptions(options.Manipulation)
	if err != nil {
		return EspressoOutputMinimizeResult{}, err
	}
	options.Manipulation = manipulation
	if options.Heuristic.Stop == nil {
		options.Heuristic.Stop = manipulation.Stop
	}
	if options.Exact.Stop == nil {
		options.Exact.Stop = manipulation.Stop
	}
	output := pla.Layout.OutputVariable()
	if output < 0 {
		return EspressoOutputMinimizeResult{}, ErrDimension
	}
	count := pla.Layout.partSizes[output]
	combined, _ := NewEspressoSetFamily(len(pla.On.sets), pla.Layout.bitCount)
	phase := make([]bool, count)
	cubeCounts := make([]int, count)
	for part := 0; part < count; part++ {
		if manipulation.Stop != nil && manipulation.Stop() {
			return EspressoOutputMinimizeResult{}, ErrCancelled
		}
		on, err := pla.Layout.CofactorOutput(pla.On, part, manipulation)
		if err != nil {
			return EspressoOutputMinimizeResult{}, err
		}
		dontCare, err := pla.Layout.CofactorOutput(pla.DontCare, part, manipulation)
		if err != nil {
			return EspressoOutputMinimizeResult{}, err
		}
		off, err := pla.Layout.CofactorOutput(pla.Off, part, manipulation)
		if err != nil {
			return EspressoOutputMinimizeResult{}, err
		}
		positive, err := pla.Layout.minimizeOneOutput(on, dontCare, off, options)
		if err != nil {
			return EspressoOutputMinimizeResult{}, err
		}
		chosen := positive
		phase[part] = true
		if options.ChoosePolarity {
			negative, negativeErr := pla.Layout.minimizeOneOutput(off, dontCare, positive, options)
			if negativeErr != nil {
				return EspressoOutputMinimizeResult{}, negativeErr
			}
			if len(negative.sets) < len(positive.sets) {
				chosen, phase[part] = negative, false
			}
		}
		chosen, err = pla.Layout.UncofactorOutput(chosen, part, manipulation)
		if err != nil {
			return EspressoOutputMinimizeResult{}, err
		}
		cubeCounts[part] = len(chosen.sets)
		if len(combined.sets)+len(chosen.sets) > manipulation.MaxCubes {
			return EspressoOutputMinimizeResult{}, ErrLimit
		}
		for _, cube := range chosen.sets {
			_ = combined.Add(cube)
		}
	}
	result, err := NewEspressoPLA(pla.Layout, combined, pla.DontCare, pla.Off, pla.Labels)
	if err != nil {
		return EspressoOutputMinimizeResult{}, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	return EspressoOutputMinimizeResult{PLA: result, PositivePhase: phase, CubeCounts: cubeCounts}, nil
}
