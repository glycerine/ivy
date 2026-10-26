package abc

// Binary-variable pairing and pairing search from src/misc/espresso/pair.c.

import (
	"math"
	"slices"
	"strings"
)

type EspressoVariablePair struct {
	First, Second int
}

type EspressoPairing struct {
	Pairs []EspressoVariablePair
}

func (pairing EspressoPairing) Clone() EspressoPairing {
	return EspressoPairing{Pairs: slices.Clone(pairing.Pairs)}
}

func (pairing EspressoPairing) Format() string {
	var result strings.Builder
	result.WriteString("pair is")
	for _, pair := range pairing.Pairs {
		result.WriteString(" (")
		result.WriteString(itoa(pair.First + 1))
		result.WriteByte(' ')
		result.WriteString(itoa(pair.Second + 1))
		result.WriteByte(')')
	}
	return result.String()
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buffer [24]byte
	position := len(buffer)
	for value > 0 {
		position--
		buffer[position] = byte('0' + value%10)
		value /= 10
	}
	return string(buffer[position:])
}

type EspressoPairCostStrategy uint8

const (
	EspressoPairWeak EspressoPairCostStrategy = iota
	EspressoPairStrong
	EspressoPairFull
	EspressoPairExact
)

type EspressoPairOptimizationStrategy uint8

const (
	EspressoPairOptimizeHeuristic EspressoPairOptimizationStrategy = iota
	EspressoPairOptimizeExact
	EspressoPairOptimizePhase
)

type EspressoPairOptions struct {
	MaxPairings                      int
	MaxCubes, MaxOperations, MaxRows int
	MaxDepth, MaxSearchNodes         int
	Stop                             func() bool
	Heuristic                        EspressoMinimizeOptions
	Exact                            EspressoExactOptions
	Phase                            EspressoPhaseOptions
}

type EspressoPairResult struct {
	PLA             *EspressoPLA
	Pairing         EspressoPairing
	PositivePhase   []bool
	SparseVariables []bool
	Cost            int
	Evaluated       int
}

func normalizeEspressoPairOptions(options EspressoPairOptions) (EspressoPairOptions, error) {
	if options.MaxPairings < 0 || options.MaxCubes < 0 || options.MaxOperations < 0 || options.MaxRows < 0 || options.MaxDepth < 0 || options.MaxSearchNodes < 0 {
		return options, ErrDimension
	}
	if options.MaxPairings == 0 {
		options.MaxPairings = 1 << 20
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
	if options.MaxSearchNodes == 0 {
		options.MaxSearchNodes = 1 << 20
	}
	if options.Heuristic.MaxCubes == 0 {
		options.Heuristic.MaxCubes = options.MaxCubes
	}
	if options.Heuristic.MaxOperations == 0 {
		options.Heuristic.MaxOperations = options.MaxOperations
	}
	if options.Heuristic.MaxRows == 0 {
		options.Heuristic.MaxRows = options.MaxRows
	}
	if options.Heuristic.MaxDepth == 0 {
		options.Heuristic.MaxDepth = options.MaxDepth
	}
	if options.Heuristic.Stop == nil {
		options.Heuristic.Stop = options.Stop
	}
	if options.Exact.MaxCubes == 0 {
		options.Exact.MaxCubes = options.MaxCubes
	}
	if options.Exact.MaxOperations == 0 {
		options.Exact.MaxOperations = options.MaxOperations
	}
	if options.Exact.MaxDepth == 0 {
		options.Exact.MaxDepth = options.MaxDepth
	}
	if options.Exact.MaxSearchNodes == 0 {
		options.Exact.MaxSearchNodes = options.MaxSearchNodes
	}
	if options.Exact.Stop == nil {
		options.Exact.Stop = options.Stop
	}
	if options.Phase.MaxCubes == 0 {
		options.Phase.MaxCubes = options.MaxCubes
	}
	if options.Phase.MaxOperations == 0 {
		options.Phase.MaxOperations = options.MaxOperations
	}
	if options.Phase.MaxPhases == 0 {
		options.Phase.MaxPhases = options.MaxPairings
	}
	if options.Phase.Stop == nil {
		options.Phase.Stop = options.Stop
	}
	return options, nil
}

func (layout *EspressoCubeLayout) validatePairing(pairing EspressoPairing) ([]bool, error) {
	if err := layout.check(); err != nil {
		return nil, err
	}
	paired := make([]bool, layout.binaryVariables)
	for _, pair := range pairing.Pairs {
		if pair.First < 0 || pair.Second < 0 || pair.First >= layout.binaryVariables || pair.Second >= layout.binaryVariables || pair.First == pair.Second || paired[pair.First] || paired[pair.Second] {
			return nil, ErrDimension
		}
		paired[pair.First], paired[pair.Second] = true, true
	}
	return paired, nil
}

func espressoPairTransformSet(oldLayout, newLayout *EspressoCubeLayout, source *EspressoSet, pairing EspressoPairing, paired []bool, work *espressoCoverManipulationWork) (*EspressoSet, error) {
	result, _ := NewEspressoSet(newLayout.bitCount)
	result.Flags, result.Rank = source.Flags, source.Rank
	newVariable := 0
	for variable := 0; variable < oldLayout.binaryVariables; variable++ {
		if paired[variable] {
			continue
		}
		if err := work.step(2); err != nil {
			return nil, err
		}
		for value := 0; value < 2; value++ {
			present, _ := source.Has(oldLayout.firstParts[variable] + value)
			if present {
				_ = result.Set(newLayout.firstParts[newVariable]+value, true)
			}
		}
		newVariable++
	}
	for pairIndex, pair := range pairing.Pairs {
		start := newLayout.firstParts[newLayout.binaryVariables+pairIndex]
		for firstValue := 0; firstValue < 2; firstValue++ {
			firstPresent, _ := source.Has(oldLayout.firstParts[pair.First] + firstValue)
			for secondValue := 0; secondValue < 2; secondValue++ {
				if err := work.step(1); err != nil {
					return nil, err
				}
				secondPresent, _ := source.Has(oldLayout.firstParts[pair.Second] + secondValue)
				if firstPresent && secondPresent {
					_ = result.Set(start+3-(2*firstValue+secondValue), true)
				}
			}
		}
	}
	newMVStart := newLayout.binaryVariables + len(pairing.Pairs)
	for offset, variable := range oldLayout.partSizes[oldLayout.binaryVariables:] {
		oldVariable := oldLayout.binaryVariables + offset
		newStart := newLayout.firstParts[newMVStart+offset]
		for part := 0; part < variable; part++ {
			if err := work.step(1); err != nil {
				return nil, err
			}
			present, _ := source.Has(oldLayout.firstParts[oldVariable] + part)
			if present {
				_ = result.Set(newStart+part, true)
			}
		}
	}
	return result, nil
}

func espressoPairLabels(pla *EspressoPLA, layout *EspressoCubeLayout, pairing EspressoPairing, paired []bool) []string {
	if len(pla.Labels) == 0 {
		return nil
	}
	labels := make([]string, layout.bitCount)
	newVariable := 0
	for variable := 0; variable < pla.Layout.binaryVariables; variable++ {
		if paired[variable] {
			continue
		}
		copy(labels[layout.firstParts[newVariable]:layout.firstParts[newVariable]+2], pla.Labels[pla.Layout.firstParts[variable]:pla.Layout.firstParts[variable]+2])
		newVariable++
	}
	for index, pair := range pairing.Pairs {
		start := layout.firstParts[layout.binaryVariables+index]
		first := pla.Layout.firstParts[pair.First]
		second := pla.Layout.firstParts[pair.Second]
		labels[start] = pla.Labels[first] + "+" + pla.Labels[second]
		labels[start+1] = pla.Labels[first] + "+" + pla.Labels[second+1]
		labels[start+2] = pla.Labels[first+1] + "+" + pla.Labels[second]
		labels[start+3] = pla.Labels[first+1] + "+" + pla.Labels[second+1]
	}
	newMV := layout.binaryVariables + len(pairing.Pairs)
	for offset, size := range pla.Layout.partSizes[pla.Layout.binaryVariables:] {
		oldVariable := pla.Layout.binaryVariables + offset
		copy(labels[layout.firstParts[newMV+offset]:layout.firstParts[newMV+offset]+size], pla.Labels[pla.Layout.firstParts[oldVariable]:pla.Layout.firstParts[oldVariable]+size])
	}
	return labels
}

// PairVariables replaces disjoint binary pairs with four-valued variables.
func (pla *EspressoPLA) PairVariables(pairing EspressoPairing, adjustLabels bool, options EspressoPairOptions) (EspressoPairResult, error) {
	if err := pla.check(); err != nil {
		return EspressoPairResult{}, err
	}
	options, err := normalizeEspressoPairOptions(options)
	if err != nil {
		return EspressoPairResult{}, err
	}
	paired, err := pla.Layout.validatePairing(pairing)
	if err != nil {
		return EspressoPairResult{}, err
	}
	newBinary := pla.Layout.binaryVariables - 2*len(pairing.Pairs)
	partSizes := make([]int, 0, pla.Layout.VariableCount()-len(pairing.Pairs))
	for variable := 0; variable < pla.Layout.binaryVariables; variable++ {
		if !paired[variable] {
			partSizes = append(partSizes, 2)
		}
	}
	for range pairing.Pairs {
		partSizes = append(partSizes, 4)
	}
	partSizes = append(partSizes, pla.Layout.partSizes[pla.Layout.binaryVariables:]...)
	layout, err := NewEspressoCubeLayout(partSizes, newBinary)
	if err != nil {
		return EspressoPairResult{}, err
	}
	for offset := range pla.Layout.partSizes[pla.Layout.binaryVariables:] {
		layout.symbolic[newBinary+len(pairing.Pairs)+offset] = pla.Layout.symbolic[pla.Layout.binaryVariables+offset]
	}
	work, _ := newEspressoCoverManipulationWork(EspressoCoverManipulationOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
	transform := func(source *EspressoSetFamily) (*EspressoSetFamily, error) {
		result, _ := NewEspressoSetFamily(len(source.sets), layout.bitCount)
		for _, cube := range source.sets {
			if err := work.add(1); err != nil {
				return nil, err
			}
			mapped, err := espressoPairTransformSet(pla.Layout, layout, cube, pairing, paired, work)
			if err != nil {
				return nil, err
			}
			_ = result.Add(mapped)
		}
		return result, nil
	}
	on, err := transform(pla.On)
	if err != nil {
		return EspressoPairResult{}, err
	}
	dontCare, err := transform(pla.DontCare)
	if err != nil {
		return EspressoPairResult{}, err
	}
	off, err := transform(pla.Off)
	if err != nil {
		return EspressoPairResult{}, err
	}
	labels := []string(nil)
	if adjustLabels {
		labels = espressoPairLabels(pla, layout, pairing, paired)
	}
	result, err := NewEspressoPLA(layout, on, dontCare, off, labels)
	if err != nil {
		return EspressoPairResult{}, err
	}
	result.Filename, result.Type = pla.Filename, pla.Type
	sparse, _ := layout.DefaultSparseVariables()
	for variable := newBinary; variable < newBinary+len(pairing.Pairs); variable++ {
		sparse[variable] = false
	}
	return EspressoPairResult{PLA: result, Pairing: pairing.Clone(), SparseVariables: sparse}, nil
}

// EnumeratePairings returns every maximum-cardinality disjoint pairing in the
// source recursion order. For odd n, exactly one variable remains unpaired.
func EnumeratePairings(variables int, options EspressoPairOptions) ([]EspressoPairing, error) {
	options, err := normalizeEspressoPairOptions(options)
	if err != nil || variables < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	available := make([]bool, variables)
	for index := range available {
		available[index] = true
	}
	current := EspressoPairing{}
	result := make([]EspressoPairing, 0)
	var recur func(int) error
	recur = func(left int) error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		if left < 2 {
			if len(result) >= options.MaxPairings {
				return ErrLimit
			}
			result = append(result, current.Clone())
			return nil
		}
		first := 0
		for first < variables && !available[first] {
			first++
		}
		available[first] = false
		for second := first + 1; second < variables; second++ {
			if !available[second] {
				continue
			}
			available[second] = false
			current.Pairs = append(current.Pairs, EspressoVariablePair{First: first, Second: second})
			if err := recur(left - 2); err != nil {
				return err
			}
			current.Pairs = current.Pairs[:len(current.Pairs)-1]
			available[second] = true
		}
		if left%2 == 1 {
			if err := recur(left - 1); err != nil {
				return err
			}
		}
		available[first] = true
		return nil
	}
	if err := recur(variables); err != nil {
		return nil, err
	}
	return result, nil
}

func pairingCost(costs [][]int, pairing EspressoPairing) (int, error) {
	total := 0
	for _, pair := range pairing.Pairs {
		if pair.First < 0 || pair.Second < 0 || pair.First >= len(costs) || pair.Second >= len(costs) || len(costs[pair.First]) != len(costs) {
			return 0, ErrDimension
		}
		value := costs[pair.First][pair.Second]
		if value > 0 && total > math.MaxInt-value || value < 0 && total < math.MinInt-value {
			return 0, ErrLimit
		}
		total += value
	}
	return total, nil
}

func BestPairing(costs [][]int, exhaustive bool, options EspressoPairOptions) (EspressoPairing, int, error) {
	options, err := normalizeEspressoPairOptions(options)
	if err != nil {
		return EspressoPairing{}, 0, err
	}
	for _, row := range costs {
		if len(row) != len(costs) {
			return EspressoPairing{}, 0, ErrDimension
		}
	}
	if exhaustive {
		pairings, err := EnumeratePairings(len(costs), options)
		if err != nil {
			return EspressoPairing{}, 0, err
		}
		bestCost := math.MinInt
		best := EspressoPairing{}
		for _, pairing := range pairings {
			cost, costErr := pairingCost(costs, pairing)
			if costErr != nil {
				return EspressoPairing{}, 0, costErr
			}
			if cost > bestCost {
				best, bestCost = pairing.Clone(), cost
			}
		}
		return best, bestCost, nil
	}
	available := make([]bool, len(costs))
	for index := range available {
		available[index] = true
	}
	result, total := EspressoPairing{}, 0
	for left := len(costs); left >= 2; left -= 2 {
		bestFirst, bestSecond, bestCost := -1, -1, 0
		for first := range costs {
			if !available[first] {
				continue
			}
			for second := first + 1; second < len(costs); second++ {
				if available[second] && (bestFirst < 0 || costs[first][second] > bestCost) {
					bestFirst, bestSecond, bestCost = first, second, costs[first][second]
				}
			}
		}
		if bestFirst < 0 {
			break
		}
		available[bestFirst], available[bestSecond] = false, false
		result.Pairs = append(result.Pairs, EspressoVariablePair{First: bestFirst, Second: bestSecond})
		if bestCost > 0 && total > math.MaxInt-bestCost || bestCost < 0 && total < math.MinInt-bestCost {
			return EspressoPairing{}, 0, ErrLimit
		}
		total += bestCost
	}
	return result, total, nil
}

func (pla *EspressoPLA) weakPairCost(first, second int, options EspressoPairOptions) (int, error) {
	merged, _ := pla.On.Clone()
	for _, cube := range merged.sets {
		for _, variable := range []int{first, second} {
			for value := 0; value < 2; value++ {
				_ = cube.Set(pla.Layout.firstParts[variable]+value, true)
			}
		}
	}
	unique, err := merged.Deduplicate(EspressoContainOptions{MaxSets: options.MaxCubes, MaxChecks: options.MaxOperations, Stop: options.Stop})
	if err != nil {
		return 0, err
	}
	return len(pla.On.sets) - len(unique.sets), nil
}

func (pla *EspressoPLA) evaluatePair(first, second int, strategy EspressoPairCostStrategy, options EspressoPairOptions) (int, error) {
	if strategy == EspressoPairWeak {
		return pla.weakPairCost(first, second, options)
	}
	paired, err := pla.PairVariables(EspressoPairing{Pairs: []EspressoVariablePair{{First: first, Second: second}}}, false, options)
	if err != nil {
		return 0, err
	}
	before := len(pla.On.sets)
	switch strategy {
	case EspressoPairStrong:
		reduced, err := paired.PLA.Layout.ReduceCover(paired.PLA.On, paired.PLA.DontCare, EspressoReduceOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, Stop: options.Stop})
		if err != nil {
			return 0, err
		}
		expanded, err := paired.PLA.Layout.ExpandCover(reduced, paired.PLA.Off, EspressoExpandOptions{SparseVariables: paired.SparseVariables, MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxDepth: options.MaxDepth, Stop: options.Stop})
		if err != nil {
			return 0, err
		}
		irred, err := paired.PLA.Layout.Irredundant(expanded, paired.PLA.DontCare, EspressoIrredundantOptions{MaxCubes: options.MaxCubes, MaxOperations: options.MaxOperations, MaxRows: options.MaxRows, Stop: options.Stop})
		if err != nil {
			return 0, err
		}
		return before - len(irred.Cover.sets), nil
	case EspressoPairFull:
		heuristic := options.Heuristic
		heuristic.SparseVariables = slices.Clone(paired.SparseVariables)
		result, err := paired.PLA.Layout.MinimizeEspressoCover(paired.PLA.On, paired.PLA.DontCare, paired.PLA.Off, heuristic)
		if err != nil {
			return 0, err
		}
		return before - len(result.Cover.sets), nil
	case EspressoPairExact:
		exact := options.Exact
		exact.SparseVariables = slices.Clone(paired.SparseVariables)
		result, err := paired.PLA.Layout.MinimizeExact(paired.PLA.On, paired.PLA.DontCare, paired.PLA.Off, true, exact)
		if err != nil {
			return 0, err
		}
		return before - len(result.Cover.sets), nil
	default:
		return 0, ErrDimension
	}
}

// PairingCosts estimates every binary pair by the selected source strategy.
func (pla *EspressoPLA) PairingCosts(strategy EspressoPairCostStrategy, options EspressoPairOptions) ([][]int, error) {
	if err := pla.check(); err != nil {
		return nil, err
	}
	if strategy > EspressoPairExact {
		return nil, ErrDimension
	}
	options, err := normalizeEspressoPairOptions(options)
	if err != nil {
		return nil, err
	}
	count := pla.Layout.binaryVariables
	costs := make([][]int, count)
	for index := range costs {
		costs[index] = make([]int, count)
	}
	evaluated := 0
	for first := 0; first < count-1; first++ {
		for second := first + 1; second < count; second++ {
			evaluated++
			if evaluated > options.MaxPairings {
				return nil, ErrLimit
			}
			cost, err := pla.evaluatePair(first, second, strategy, options)
			if err != nil {
				return nil, err
			}
			costs[first][second] = cost
		}
	}
	return costs, nil
}

// FindOptimalPairing estimates individual edges, chooses the best matching,
// applies it and runs the full heuristic minimizer.
func (pla *EspressoPLA) FindOptimalPairing(strategy EspressoPairCostStrategy, options EspressoPairOptions) (EspressoPairResult, error) {
	options, err := normalizeEspressoPairOptions(options)
	if err != nil {
		return EspressoPairResult{}, err
	}
	costs, err := pla.PairingCosts(strategy, options)
	if err != nil {
		return EspressoPairResult{}, err
	}
	pairing, cost, err := BestPairing(costs, len(costs) <= 14, options)
	if err != nil {
		return EspressoPairResult{}, err
	}
	paired, err := pla.PairVariables(pairing, true, options)
	if err != nil {
		return EspressoPairResult{}, err
	}
	heuristic := options.Heuristic
	heuristic.SparseVariables = slices.Clone(paired.SparseVariables)
	minimized, err := paired.PLA.Layout.MinimizeEspressoCover(paired.PLA.On, paired.PLA.DontCare, paired.PLA.Off, heuristic)
	if err != nil {
		return EspressoPairResult{}, err
	}
	paired.PLA.On = minimized.Cover
	paired.Cost = cost
	paired.Evaluated = len(costs) * (len(costs) - 1) / 2
	return paired, nil
}

func (pla *EspressoPLA) optimizePairedCandidate(pairing EspressoPairing, strategy EspressoPairOptimizationStrategy, options EspressoPairOptions) (EspressoPairResult, error) {
	paired, err := pla.PairVariables(pairing, true, options)
	if err != nil {
		return EspressoPairResult{}, err
	}
	switch strategy {
	case EspressoPairOptimizeHeuristic:
		heuristic := options.Heuristic
		heuristic.SparseVariables = slices.Clone(paired.SparseVariables)
		result, err := paired.PLA.Layout.MinimizeEspressoCover(paired.PLA.On, paired.PLA.DontCare, paired.PLA.Off, heuristic)
		if err != nil {
			return EspressoPairResult{}, err
		}
		paired.PLA.On = result.Cover
	case EspressoPairOptimizeExact:
		exact := options.Exact
		exact.SparseVariables = slices.Clone(paired.SparseVariables)
		result, err := paired.PLA.Layout.MinimizeExact(paired.PLA.On, paired.PLA.DontCare, paired.PLA.Off, true, exact)
		if err != nil {
			return EspressoPairResult{}, err
		}
		paired.PLA.On = result.Cover
	case EspressoPairOptimizePhase:
		result, err := paired.PLA.AssignOutputPhase(options.Phase)
		if err != nil {
			return EspressoPairResult{}, err
		}
		paired.PLA, paired.PositivePhase = result.PLA, result.PositivePhase
	default:
		return EspressoPairResult{}, ErrDimension
	}
	return paired, nil
}

// OptimizePairing evaluates every complete pairing after actual minimization,
// replacing pair_all/minimize_pair and their process-global scratch.
func (pla *EspressoPLA) OptimizePairing(strategy EspressoPairOptimizationStrategy, options EspressoPairOptions) (EspressoPairResult, error) {
	if err := pla.check(); err != nil {
		return EspressoPairResult{}, err
	}
	if strategy > EspressoPairOptimizePhase {
		return EspressoPairResult{}, ErrDimension
	}
	options, err := normalizeEspressoPairOptions(options)
	if err != nil {
		return EspressoPairResult{}, err
	}
	pairings, err := EnumeratePairings(pla.Layout.binaryVariables, options)
	if err != nil {
		return EspressoPairResult{}, err
	}
	best := EspressoPairResult{}
	bestCount := int(^uint(0) >> 1)
	for index, pairing := range pairings {
		candidate, err := pla.optimizePairedCandidate(pairing, strategy, options)
		if err != nil {
			return EspressoPairResult{}, err
		}
		if len(candidate.PLA.On.sets) < bestCount {
			best, bestCount = candidate, len(candidate.PLA.On.sets)
		}
		best.Evaluated = index + 1
	}
	if best.PLA == nil {
		return EspressoPairResult{}, ErrInvariant
	}
	return best, nil
}
