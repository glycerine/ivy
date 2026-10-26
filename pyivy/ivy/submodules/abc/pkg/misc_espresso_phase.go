package abc

// Output-phase assignment and double-phase covers from src/misc/espresso/opo.c.

import "slices"

type EspressoPhaseStrategy uint8

const (
	EspressoPhaseHeuristic EspressoPhaseStrategy = iota
	EspressoPhaseExact
)

type EspressoPhaseOptions struct {
	Strategy                EspressoPhaseStrategy
	Repeated                bool
	MaxPhases               int
	MaxCubes, MaxOperations int
	Stop                    func() bool
	Heuristic               EspressoMinimizeOptions
	Exact                   EspressoExactOptions
}

type EspressoPhaseResult struct {
	PLA           *EspressoPLA
	PositivePhase []bool
	Evaluated     int
	Cost          EspressoCoverCost
}

func normalizeEspressoPhaseOptions(options EspressoPhaseOptions) (EspressoPhaseOptions, error) {
	if options.Strategy > EspressoPhaseExact || options.MaxPhases < 0 || options.MaxCubes < 0 || options.MaxOperations < 0 {
		return options, ErrDimension
	}
	if options.MaxPhases == 0 {
		options.MaxPhases = 1 << 20
	}
	if options.MaxCubes == 0 {
		options.MaxCubes = 1 << 20
	}
	if options.MaxOperations == 0 {
		options.MaxOperations = 16 << 20
	}
	if options.Heuristic.Stop == nil {
		options.Heuristic.Stop = options.Stop
	}
	if options.Heuristic.MaxCubes == 0 {
		options.Heuristic.MaxCubes = options.MaxCubes
	}
	if options.Heuristic.MaxOperations == 0 {
		options.Heuristic.MaxOperations = options.MaxOperations
	}
	if options.Exact.Stop == nil {
		options.Exact.Stop = options.Stop
	}
	if options.Exact.MaxCubes == 0 {
		options.Exact.MaxCubes = options.MaxCubes
	}
	if options.Exact.MaxOperations == 0 {
		options.Exact.MaxOperations = options.MaxOperations
	}
	return options, nil
}

func espressoPhaseCopyPrefix(source *EspressoSet, bits int, end int) *EspressoSet {
	result, _ := NewEspressoSet(bits)
	for part := 0; part < end; part++ {
		present, _ := source.Has(part)
		if present {
			_ = result.Set(part, true)
		}
	}
	result.Flags, result.Rank = source.Flags, source.Rank
	return result
}

// OutputPhaseSetup constructs the source double-phase characteristic PLA.
// Outputs at and after firstOutput are duplicated; earlier outputs stay fixed.
func (pla *EspressoPLA) OutputPhaseSetup(firstOutput int, options EspressoPhaseOptions) (*EspressoPLA, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	options, err := normalizeEspressoPhaseOptions(options)
	if err != nil {
		return nil, err
	}
	output := pla.Layout.OutputVariable()
	if output < 0 || firstOutput < 0 || firstOutput >= pla.Layout.partSizes[output] {
		return nil, ErrDimension
	}
	originalOutputs := pla.Layout.partSizes[output]
	offset := originalOutputs - firstOutput
	partSizes := slices.Clone(pla.Layout.partSizes)
	partSizes[output] += offset
	layout, err := NewEspressoCubeLayout(partSizes, pla.Layout.binaryVariables)
	if err != nil {
		return nil, err
	}
	copy(layout.symbolic, pla.Layout.symbolic)
	oldStart := pla.Layout.firstParts[output]
	firstPart := oldStart + firstOutput
	newOn, _ := NewEspressoSetFamily(len(pla.On.sets)+len(pla.Off.sets), layout.bitCount)
	newOff, _ := NewEspressoSetFamily(len(pla.On.sets)+len(pla.Off.sets), layout.bitCount)
	newDC, _ := NewEspressoSetFamily(len(pla.DontCare.sets), layout.bitCount)
	operations, generated := 0, 0
	step := func(count int) error {
		if count < 0 || count > options.MaxOperations-operations {
			return ErrLimit
		}
		operations += count
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	add := func(family *EspressoSetFamily, cube *EspressoSet) error {
		generated++
		if generated > options.MaxCubes {
			return ErrLimit
		}
		return family.Add(cube)
	}
	for _, cube := range pla.On.sets {
		if err := step(originalOutputs); err != nil {
			return nil, err
		}
		positive := espressoPhaseCopyPrefix(cube, layout.bitCount, firstPart)
		negative := espressoPhaseCopyPrefix(cube, layout.bitCount, oldStart)
		hasTail := false
		for part := firstPart; part < oldStart+originalOutputs; part++ {
			present, _ := cube.Has(part)
			if present {
				hasTail = true
				_ = positive.Set(part, true)
				_ = negative.Set(part+offset, true)
			}
		}
		if err := add(newOn, positive); err != nil {
			return nil, err
		}
		if hasTail {
			if err := add(newOff, negative); err != nil {
				return nil, err
			}
		}
	}
	for _, cube := range pla.Off.sets {
		if err := step(originalOutputs); err != nil {
			return nil, err
		}
		negative := espressoPhaseCopyPrefix(cube, layout.bitCount, oldStart)
		positive := espressoPhaseCopyPrefix(cube, layout.bitCount, firstPart)
		hasTail := false
		for part := firstPart; part < oldStart+originalOutputs; part++ {
			present, _ := cube.Has(part)
			if present {
				hasTail = true
				_ = negative.Set(part+offset, true)
				_ = positive.Set(part, true)
			}
		}
		if hasTail {
			if err := add(newOn, negative); err != nil {
				return nil, err
			}
		}
		if err := add(newOff, positive); err != nil {
			return nil, err
		}
	}
	for _, cube := range pla.DontCare.sets {
		if err := step(originalOutputs); err != nil {
			return nil, err
		}
		both := espressoPhaseCopyPrefix(cube, layout.bitCount, firstPart)
		for part := firstPart; part < oldStart+originalOutputs; part++ {
			present, _ := cube.Has(part)
			if present {
				_ = both.Set(part, true)
				_ = both.Set(part+offset, true)
			}
		}
		if err := add(newDC, both); err != nil {
			return nil, err
		}
	}
	labels := []string(nil)
	if len(pla.Labels) != 0 {
		labels = make([]string, layout.bitCount)
		copy(labels, pla.Labels)
		for part := firstOutput; part < originalOutputs; part++ {
			label := pla.Labels[oldStart+part]
			if label != "" {
				label += ".bar"
			}
			labels[oldStart+part+offset] = label
		}
	}
	result, err := NewEspressoPLA(layout, newOn, newDC, newOff, labels)
	if err != nil {
		return nil, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	return result, nil
}

func (pla *EspressoPLA) phasePiece(cube *EspressoSet, phase []bool, positive bool) (*EspressoSet, bool) {
	output := pla.Layout.OutputVariable()
	start := pla.Layout.firstParts[output]
	result := espressoPhaseCopyPrefix(cube, pla.Layout.bitCount, start)
	hasOutput := false
	for part, polarity := range phase {
		present, _ := cube.Has(start + part)
		if present && polarity == positive {
			hasOutput = true
			_ = result.Set(start+part, true)
		}
	}
	return result, hasOutput
}

// ApplyOutputPhase moves each selected negative-polarity output between ON
// and OFF while leaving DC unchanged.
func (pla *EspressoPLA) ApplyOutputPhase(phase []bool, options EspressoPhaseOptions) (*EspressoPLA, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	options, err := normalizeEspressoPhaseOptions(options)
	if err != nil {
		return nil, err
	}
	output := pla.Layout.OutputVariable()
	if output < 0 || len(phase) != pla.Layout.partSizes[output] {
		return nil, ErrDimension
	}
	on, _ := NewEspressoSetFamily(len(pla.On.sets)+len(pla.Off.sets), pla.Layout.bitCount)
	off, _ := NewEspressoSetFamily(len(pla.On.sets)+len(pla.Off.sets), pla.Layout.bitCount)
	operations, generated := 0, 0
	add := func(family *EspressoSetFamily, cube *EspressoSet) error {
		generated++
		if generated > options.MaxCubes {
			return ErrLimit
		}
		return family.Add(cube)
	}
	for _, source := range []struct {
		family   *EspressoSetFamily
		reversed bool
	}{{pla.On, false}, {pla.Off, true}} {
		for _, cube := range source.family.sets {
			operations += len(phase)
			if operations > options.MaxOperations {
				return nil, ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			positive, hasPositive := pla.phasePiece(cube, phase, true)
			negative, hasNegative := pla.phasePiece(cube, phase, false)
			if hasPositive {
				target := on
				if source.reversed {
					target = off
				}
				if err := add(target, positive); err != nil {
					return nil, err
				}
			}
			if hasNegative {
				target := off
				if source.reversed {
					target = on
				}
				if err := add(target, negative); err != nil {
					return nil, err
				}
			}
		}
	}
	result, err := NewEspressoPLA(pla.Layout, on, pla.DontCare, off, pla.Labels)
	if err != nil {
		return nil, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	return result, nil
}

func (pla *EspressoPLA) minimizePhase(options EspressoPhaseOptions) (*EspressoPLA, EspressoCoverCost, error) {
	var cover *EspressoSetFamily
	switch options.Strategy {
	case EspressoPhaseHeuristic:
		result, err := pla.Layout.MinimizeEspressoCover(pla.On, pla.DontCare, pla.Off, options.Heuristic)
		if err != nil {
			return nil, EspressoCoverCost{}, err
		}
		cover = result.Cover
	case EspressoPhaseExact:
		result, err := pla.Layout.MinimizeExact(pla.On, pla.DontCare, pla.Off, true, options.Exact)
		if err != nil {
			return nil, EspressoCoverCost{}, err
		}
		cover = result.Cover
	default:
		return nil, EspressoCoverCost{}, ErrDimension
	}
	result, err := NewEspressoPLA(pla.Layout, cover, pla.DontCare, pla.Off, pla.Labels)
	if err != nil {
		return nil, EspressoCoverCost{}, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	cost, err := pla.Layout.CoverCost(cover, EspressoCoverCostOptions{MaxCubes: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
	return result, cost, err
}

// SelectOutputPhasePrimes performs opo/opo_recur/opo_leaf without global
// recursion state. It maximizes excludable primes, equivalently minimizing the
// union of primes needed by one polarity of every remaining output.
func (pla *EspressoPLA) SelectOutputPhasePrimes(firstOutput int, options EspressoPhaseOptions) (EspressoPhaseResult, error) {
	if err := pla.check(); err != nil {
		return EspressoPhaseResult{}, err
	}
	options, err := normalizeEspressoPhaseOptions(options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	output := pla.Layout.OutputVariable()
	if output < 0 || firstOutput < 0 || firstOutput >= pla.Layout.partSizes[output] {
		return EspressoPhaseResult{}, ErrDimension
	}
	if len(pla.On.sets) > options.MaxCubes {
		return EspressoPhaseResult{}, ErrLimit
	}
	doubled := pla.Layout.partSizes[output]
	remaining := doubled - firstOutput
	if remaining%2 != 0 {
		return EspressoPhaseResult{}, ErrDimension
	}
	offset := remaining / 2
	originalOutputs := firstOutput + offset
	start := pla.Layout.firstParts[output]
	phase := make([]bool, originalOutputs)
	for index := range phase {
		phase[index] = true
	}
	bestPhase := slices.Clone(phase)
	bestKeep := make([]bool, len(pla.On.sets))
	bestCount, evaluated, operations := len(pla.On.sets)+1, 0, 0
	var enumerate func(int) error
	enumerate = func(index int) error {
		if index == originalOutputs {
			evaluated++
			if evaluated > options.MaxPhases {
				return ErrLimit
			}
			if options.Stop != nil && options.Stop() {
				return ErrCancelled
			}
			keep := make([]bool, len(pla.On.sets))
			count := 0
			for cubeIndex, cube := range pla.On.sets {
				needed := false
				for part := 0; part < originalOutputs; part++ {
					operations++
					if operations > options.MaxOperations {
						return ErrLimit
					}
					selected := start + part
					if part >= firstOutput && !phase[part] {
						selected += offset
					}
					present, _ := cube.Has(selected)
					if present {
						needed = true
						break
					}
				}
				if needed {
					keep[cubeIndex] = true
					count++
				}
			}
			if count < bestCount {
				bestCount, bestKeep, bestPhase = count, keep, slices.Clone(phase)
			}
			return nil
		}
		if index < firstOutput {
			phase[index] = true
			return enumerate(index + 1)
		}
		phase[index] = true
		if err := enumerate(index + 1); err != nil {
			return err
		}
		phase[index] = false
		return enumerate(index + 1)
	}
	if err := enumerate(0); err != nil {
		return EspressoPhaseResult{}, err
	}
	selected, _ := NewEspressoSetFamily(bestCount, pla.Layout.bitCount)
	for index, yes := range bestKeep {
		if yes {
			_ = selected.Add(pla.On.sets[index])
		}
	}
	result, err := NewEspressoPLA(pla.Layout, selected, pla.DontCare, pla.Off, pla.Labels)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	cost, err := pla.Layout.CoverCost(selected, EspressoCoverCostOptions{MaxCubes: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	return EspressoPhaseResult{PLA: result, PositivePhase: bestPhase, Evaluated: evaluated, Cost: cost}, nil
}

// FindOutputPhase constructs and minimizes the double-phase function, then
// selects the minimum prime union. prefix fixes outputs before firstOutput.
func (pla *EspressoPLA) FindOutputPhase(firstOutput int, prefix []bool, options EspressoPhaseOptions) (EspressoPhaseResult, error) {
	if err := pla.check(); err != nil {
		return EspressoPhaseResult{}, err
	}
	options, err := normalizeEspressoPhaseOptions(options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	output := pla.Layout.OutputVariable()
	if output < 0 || firstOutput < 0 || firstOutput >= pla.Layout.partSizes[output] || len(prefix) != 0 && len(prefix) != pla.Layout.partSizes[output] {
		return EspressoPhaseResult{}, ErrDimension
	}
	working := pla
	if len(prefix) != 0 {
		working, err = pla.ApplyOutputPhase(prefix, options)
		if err != nil {
			return EspressoPhaseResult{}, err
		}
	}
	doubled, err := working.OutputPhaseSetup(firstOutput, options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	minimized, _, err := doubled.minimizePhase(options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	selected, err := minimized.SelectOutputPhasePrimes(firstOutput, options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	if len(prefix) != 0 {
		copy(selected.PositivePhase[:firstOutput], prefix[:firstOutput])
	}
	return selected, nil
}

// AssignOutputPhase implements phase_assignment and its repeated variant,
// then applies and minimizes the chosen phase on the original PLA.
func (pla *EspressoPLA) AssignOutputPhase(options EspressoPhaseOptions) (EspressoPhaseResult, error) {
	if err := pla.check(); err != nil {
		return EspressoPhaseResult{}, err
	}
	options, err := normalizeEspressoPhaseOptions(options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	output := pla.Layout.OutputVariable()
	if output < 0 {
		return EspressoPhaseResult{}, ErrDimension
	}
	count := pla.Layout.partSizes[output]
	phase := make([]bool, count)
	for index := range phase {
		phase[index] = true
	}
	evaluated := 0
	if options.Repeated {
		for first := 0; first < count; first++ {
			found, err := pla.FindOutputPhase(first, phase, options)
			if err != nil {
				return EspressoPhaseResult{}, err
			}
			phase[first] = found.PositivePhase[first]
			evaluated += found.Evaluated
			if evaluated > options.MaxPhases {
				return EspressoPhaseResult{}, ErrLimit
			}
		}
	} else {
		found, err := pla.FindOutputPhase(0, nil, options)
		if err != nil {
			return EspressoPhaseResult{}, err
		}
		phase, evaluated = found.PositivePhase, found.Evaluated
	}
	phased, err := pla.ApplyOutputPhase(phase, options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	minimized, cost, err := phased.minimizePhase(options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	return EspressoPhaseResult{PLA: minimized, PositivePhase: slices.Clone(phase), Evaluated: evaluated, Cost: cost}, nil
}

// EnumerateOutputPhases implements opoall over an inclusive output range.
func (pla *EspressoPLA) EnumerateOutputPhases(firstOutput, lastOutput int, options EspressoPhaseOptions) (EspressoPhaseResult, error) {
	if err := pla.check(); err != nil {
		return EspressoPhaseResult{}, err
	}
	options, err := normalizeEspressoPhaseOptions(options)
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	output := pla.Layout.OutputVariable()
	if output < 0 || firstOutput < 0 || lastOutput < firstOutput || lastOutput >= pla.Layout.partSizes[output] {
		return EspressoPhaseResult{}, ErrDimension
	}
	phase := make([]bool, pla.Layout.partSizes[output])
	for index := range phase {
		phase[index] = true
	}
	best, _ := pla.Clone()
	bestPhase := slices.Clone(phase)
	bestCount, evaluated := len(best.On.sets), 0
	var enumerate func(int) error
	enumerate = func(index int) error {
		if index > lastOutput {
			evaluated++
			if evaluated > options.MaxPhases {
				return ErrLimit
			}
			phased, err := pla.ApplyOutputPhase(phase, options)
			if err != nil {
				return err
			}
			candidate, _, err := phased.minimizePhase(options)
			if err != nil {
				return err
			}
			if len(candidate.On.sets) < bestCount {
				best, bestCount, bestPhase = candidate, len(candidate.On.sets), slices.Clone(phase)
			}
			return nil
		}
		phase[index] = false
		if err := enumerate(index + 1); err != nil {
			return err
		}
		phase[index] = true
		return enumerate(index + 1)
	}
	if err := enumerate(firstOutput); err != nil {
		return EspressoPhaseResult{}, err
	}
	cost, err := best.Layout.CoverCost(best.On, EspressoCoverCostOptions{MaxCubes: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
	if err != nil {
		return EspressoPhaseResult{}, err
	}
	return EspressoPhaseResult{PLA: best, PositivePhase: bestPhase, Evaluated: evaluated, Cost: cost}, nil
}
