package abc

// Main cover-minimization loop from src/misc/espresso/espresso.c.

import "slices"

type EspressoMinimizeOptions struct {
	RecomputeOnset, UnwrapOnset, RemoveEssential bool
	SingleExpand, UseSuperGasp, SkipMakeSparse   bool
	ForceIrredundant, ExactIrredundant           bool
	UseRandomOrder                               bool
	RandomSeed                                   uint64
	ReduceOrder                                  EspressoReduceOrder
	SparseVariables                              []bool
	MaxIterations                                int
	MaxCubes, MaxOperations, MaxRows, MaxDepth   int
	Stop                                         func() bool
}

type EspressoMinimizeResult struct {
	Cover      *EspressoSetFamily
	Cost       EspressoCoverCost
	Iterations int
	Unwrapped  bool
	Fallback   bool
}

func normalizeEspressoMinimizeOptions(layout *EspressoCubeLayout, options EspressoMinimizeOptions) (EspressoMinimizeOptions, error) {
	if err := layout.check(); err != nil {
		return options, err
	}
	if options.ReduceOrder > EspressoReduceColumnOrder || options.MaxIterations < 0 || options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxRows < 0 || options.MaxDepth < 0 {
		return options, ErrDimension
	}
	if len(options.SparseVariables) != 0 && len(options.SparseVariables) != layout.VariableCount() {
		return options, ErrDimension
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 64
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 16 << 20
	}
	if options.MaxRows == 0 {
		options.MaxRows = 1 << 20
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 1 << 16
	}
	if len(options.SparseVariables) == 0 {
		options.SparseVariables, _ = layout.DefaultSparseVariables()
	} else {
		options.SparseVariables = slices.Clone(options.SparseVariables)
	}
	return options, nil
}

func (options EspressoMinimizeOptions) expandOptions() EspressoExpandOptions {
	return EspressoExpandOptions{SparseVariables: options.SparseVariables, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop}
}

func (options EspressoMinimizeOptions) reduceOptions() EspressoReduceOptions {
	return EspressoReduceOptions{Order: options.ReduceOrder, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop}
}

func (options EspressoMinimizeOptions) irredundantOptions() EspressoIrredundantOptions {
	return EspressoIrredundantOptions{Exact: options.ExactIrredundant, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, Stop: options.Stop}
}

func (options EspressoMinimizeOptions) gaspOptions() EspressoGaspOptions {
	return EspressoGaspOptions{SparseVariables: options.SparseVariables, ExactIrredundant: options.ExactIrredundant, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, MaxDepth: options.MaxDepth, Stop: options.Stop}
}

func (options EspressoMinimizeOptions) sparseOptions() EspressoMakeSparseOptions {
	return EspressoMakeSparseOptions{SparseVariables: options.SparseVariables, ForceIrredundant: options.ForceIrredundant, ExactIrredundant: options.ExactIrredundant, MaxIterations: options.MaxIterations, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, MaxDepth: options.MaxDepth, Stop: options.Stop}
}

func (options EspressoMinimizeOptions) costOptions() EspressoCoverCostOptions {
	return EspressoCoverCostOptions{SparseVariables: options.SparseVariables, MaxCubes: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop}
}

func (layout *EspressoCubeLayout) unravelOutput(on *EspressoSetFamily, options EspressoMinimizeOptions) (*EspressoSetFamily, error) {
	output := layout.OutputVariable()
	if output < 0 {
		return on.Clone()
	}
	start, size := layout.firstParts[output], layout.partSizes[output]
	result, _ := NewEspressoSetFamily(len(on.sets), layout.bitCount)
	generated, operations := 0, 0
	for _, cube := range on.sets {
		for part := start; part < start+size; part++ {
			operations++
			if operations > options.MaxOperations {
				return nil, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			present, _ := cube.Has(part)
			if !present {
				continue
			}
			generated++
			if generated > options.MaxCubes {
				return nil, ErrLimit
			}
			clone, _ := cube.Clone()
			for current := start; current < start+size; current++ {
				_ = clone.Set(current, current == part)
			}
			_ = result.Add(clone)
		}
	}
	return result.Contain(EspressoContainOptions{MaxSets: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
}

func (layout *EspressoCubeLayout) minimizeEspressoOnce(on, dontCare, off *EspressoSetFamily, options EspressoMinimizeOptions) (EspressoMinimizeResult, error) {
	current, _ := on.Clone()
	dc, _ := dontCare.Clone()
	unwrapped := false
	if options.RecomputeOnset {
		var err error
		current, err = layout.SimplifyCover(current, EspressoSharpOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
		if err != nil {
			return EspressoMinimizeResult{}, err
		}
	}
	initialCost, err := layout.CoverCost(current, options.costOptions())
	if err != nil {
		return EspressoMinimizeResult{}, err
	}
	output := layout.OutputVariable()
	if options.UnwrapOnset && output >= 0 && layout.partSizes[output] > 1 && initialCost.Output != initialCost.Cubes*layout.partSizes[output] && initialCost.Output < 5000 {
		current, err = layout.unravelOutput(current, options)
		if err != nil {
			return EspressoMinimizeResult{}, err
		}
		unwrapped = true
	}
	for _, cube := range current.sets {
		cube.Flags &^= EspressoSetPrime
	}
	current, err = layout.ExpandCover(current, off, options.expandOptions())
	if err != nil {
		return EspressoMinimizeResult{}, err
	}
	marked, err := layout.Irredundant(current, dc, options.irredundantOptions())
	if err != nil {
		return EspressoMinimizeResult{}, err
	}
	current = marked.Cover
	iterations := 0
	essential, _ := NewEspressoSetFamily(0, layout.bitCount)
	if !options.SingleExpand {
		if options.RemoveEssential {
			extracted, extractErr := layout.ExtractEssentials(current, dc, EspressoEssentialOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
			if extractErr != nil {
				return EspressoMinimizeResult{}, extractErr
			}
			essential, current, dc = extracted.Essential, extracted.Remaining, extracted.DontCare
		}
		cost, costErr := layout.CoverCost(current, options.costOptions())
		if costErr != nil {
			return EspressoMinimizeResult{}, costErr
		}
		for {
			for {
				if iterations >= options.MaxIterations {
					return EspressoMinimizeResult{}, ErrLimit
				}
				iterations++
				best := cost
				reductionInput := current
				if options.UseRandomOrder {
					reductionInput, err = current.ShuffleCover(options.RandomSeed+uint64(iterations-1), EspressoCoverManipulationOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
					if err != nil {
						return EspressoMinimizeResult{}, err
					}
				}
				current, err = layout.ReduceCover(reductionInput, dc, options.reduceOptions())
				if err != nil {
					return EspressoMinimizeResult{}, err
				}
				current, err = layout.ExpandCover(current, off, options.expandOptions())
				if err != nil {
					return EspressoMinimizeResult{}, err
				}
				marked, err = layout.Irredundant(current, dc, options.irredundantOptions())
				if err != nil {
					return EspressoMinimizeResult{}, err
				}
				current = marked.Cover
				cost, err = layout.CoverCost(current, options.costOptions())
				if err != nil {
					return EspressoMinimizeResult{}, err
				}
				if cost.Cubes >= best.Cubes {
					break
				}
			}
			best := cost
			if options.UseSuperGasp {
				current, err = layout.SuperGasp(current, dc, off, options.gaspOptions())
			} else {
				current, err = layout.LastGasp(current, dc, off, options.gaspOptions())
			}
			if err != nil {
				return EspressoMinimizeResult{}, err
			}
			cost, err = layout.CoverCost(current, options.costOptions())
			if err != nil {
				return EspressoMinimizeResult{}, err
			}
			if options.UseSuperGasp && cost.Cubes >= best.Cubes {
				break
			}
			if cost.Cubes > best.Cubes || cost.Cubes == best.Cubes && cost.Total >= best.Total {
				break
			}
		}
		current, err = JoinEspressoSetFamilies(current, essential)
		if err != nil {
			return EspressoMinimizeResult{}, err
		}
	}
	if !options.SkipMakeSparse {
		current, err = layout.MakeSparse(current, dontCare, off, options.sparseOptions())
		if err != nil {
			return EspressoMinimizeResult{}, err
		}
	}
	cost, err := layout.CoverCost(current, options.costOptions())
	if err != nil {
		return EspressoMinimizeResult{}, err
	}
	return EspressoMinimizeResult{Cover: current, Cost: cost, Iterations: iterations, Unwrapped: unwrapped}, nil
}

// MinimizeEspressoCover runs the complete source minimization strategy with
// explicit options replacing process-global controls. The caller's covers are
// never mutated. If output unraveling makes the cover larger, the source
// fallback is rerun without unraveling.
func (layout *EspressoCubeLayout) MinimizeEspressoCover(on, dontCare, off *EspressoSetFamily, options EspressoMinimizeOptions) (EspressoMinimizeResult, error) {
	if err := layout.validateGaspFamilies(on, dontCare, off); err != nil {
		return EspressoMinimizeResult{}, err
	}
	options, err := normalizeEspressoMinimizeOptions(layout, options)
	if err != nil {
		return EspressoMinimizeResult{}, err
	}
	if len(on.sets)+len(dontCare.sets)+len(off.sets) > options.MaxCubes {
		return EspressoMinimizeResult{}, ErrLimit
	}
	if err = layout.validateGaspPartition(on, dontCare, off, options.gaspOptions()); err != nil {
		return EspressoMinimizeResult{}, err
	}
	result, err := layout.minimizeEspressoOnce(on, dontCare, off, options)
	if err != nil {
		return EspressoMinimizeResult{}, err
	}
	if result.Unwrapped && len(result.Cover.sets) > len(on.sets) {
		options.UnwrapOnset = false
		fallback, fallbackErr := layout.minimizeEspressoOnce(on, dontCare, off, options)
		if fallbackErr != nil {
			return EspressoMinimizeResult{}, fallbackErr
		}
		fallback.Fallback = true
		return fallback, nil
	}
	if len(result.Cover.sets) > len(on.sets) {
		original, _ := on.Clone()
		cost, costErr := layout.CoverCost(original, options.costOptions())
		if costErr != nil {
			return EspressoMinimizeResult{}, costErr
		}
		return EspressoMinimizeResult{Cover: original, Cost: cost, Iterations: result.Iterations, Fallback: true}, nil
	}
	return result, nil
}
