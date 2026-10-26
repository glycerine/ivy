package abc

// Functional-library truth enumeration from opt/dau/dauNpn2.c.

import (
	"slices"

	"github.com/glycerine/insdict"
)

type DttEnumerationOptions struct {
	Variables, MaxFrontier int
	Delay, Multiplicity    bool
	MaxFunctions           int
	MaxCandidates          uint64
	MaxOrbitTransforms     uint64
	Stop                   func() bool
}

type DttFunction struct {
	Truth, Canonical     uint64
	Frontier, Nodes      int
	Fanin0, Fanin1       int
	Type                 uint8
	LeftTransform        DttNPNTransform
	LeftOutputComplement bool
}

type DttLevel struct {
	Frontier, Start, End int
	Candidates, Pairs    uint64
}

type DttClass struct {
	Canonical                 uint64
	Frontier, Implementations int
	Multiplicity              uint64
}

type DttEnumeration struct {
	variables int
	functions []DttFunction
	levels    []DttLevel
	classes   []DttClass
}

func (enumeration *DttEnumeration) Variables() int {
	if enumeration == nil {
		return 0
	}
	return enumeration.variables
}

func (enumeration *DttEnumeration) Functions() []DttFunction {
	if enumeration == nil {
		return nil
	}
	return slices.Clone(enumeration.functions)
}

func (enumeration *DttEnumeration) Levels() []DttLevel {
	if enumeration == nil {
		return nil
	}
	return slices.Clone(enumeration.levels)
}

func (enumeration *DttEnumeration) Classes() []DttClass {
	if enumeration == nil {
		return nil
	}
	return slices.Clone(enumeration.classes)
}

type dttOrbitWalker struct {
	variables  int
	mask       uint64
	top        uint64
	phaseFlips []int
	permSwaps  []int
	remaining  uint64
	limited    bool
	stop       func() bool
}

type dttOrbitEntry struct {
	Truth            uint64
	Transform        DttNPNTransform
	OutputComplement bool
}

func newDttOrbitWalker(variables int, maxTransforms uint64, stop func() bool) (*dttOrbitWalker, error) {
	if variables < 2 || variables > 5 {
		return nil, ErrDimension
	}
	schedule, err := NewTruthPermutationSchedule(variables, 0)
	if err != nil {
		return nil, err
	}
	phaseFlips := append(schedule.GrayFlips(), variables-1)
	permSwaps := schedule.AdjacentSwaps()
	labels := make([]int, variables)
	for index := range labels {
		labels[index] = index
	}
	for _, swap := range permSwaps {
		labels[swap], labels[swap+1] = labels[swap+1], labels[swap]
	}
	final := -1
	for swap := 0; swap+1 < variables; swap++ {
		labels[swap], labels[swap+1] = labels[swap+1], labels[swap]
		identity := true
		for index, label := range labels {
			identity = identity && index == label
		}
		labels[swap], labels[swap+1] = labels[swap+1], labels[swap]
		if identity {
			final = swap
			break
		}
	}
	if final < 0 {
		return nil, ErrInvariant
	}
	permSwaps = append(permSwaps, final)
	walker := &dttOrbitWalker{
		variables:  variables,
		mask:       dauTruthMask(variables),
		top:        uint64(1) << uint((1<<uint(variables))-1),
		phaseFlips: phaseFlips,
		permSwaps:  permSwaps,
		remaining:  maxTransforms,
		limited:    maxTransforms != 0,
		stop:       stop,
	}
	return walker, nil
}

func dttFlipTruth(truth uint64, variables, variable int) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		source := assignment ^ (1 << uint(variable))
		if truth&(1<<uint(source)) != 0 {
			result |= 1 << uint(assignment)
		}
	}
	return result
}

func dttSwapTruth(truth uint64, variables, variable int) uint64 {
	var result uint64
	for assignment := 0; assignment < 1<<uint(variables); assignment++ {
		first := (assignment >> uint(variable)) & 1
		second := (assignment >> uint(variable+1)) & 1
		source := assignment
		if first != second {
			source ^= 1<<uint(variable) | 1<<uint(variable+1)
		}
		if truth&(1<<uint(source)) != 0 {
			result |= 1 << uint(assignment)
		}
	}
	return result
}

func (walker *dttOrbitWalker) normalize(truth uint64) uint64 {
	truth &= walker.mask
	if truth&walker.top != 0 {
		truth = (^truth) & walker.mask
	}
	return truth
}

func (walker *dttOrbitWalker) orbit(truth uint64) ([]uint64, error) {
	entries, err := walker.entries(truth)
	if err != nil {
		return nil, err
	}
	result := make([]uint64, len(entries))
	for index, entry := range entries {
		result[index] = entry.Truth
	}
	return result, nil
}

func (walker *dttOrbitWalker) entries(truth uint64) ([]dttOrbitEntry, error) {
	truth &= walker.mask
	current := truth
	transform, err := IdentityDttNPNTransform(walker.variables)
	if err != nil {
		return nil, err
	}
	seen := insdict.NewDict[uint64, bool]()
	result := make([]dttOrbitEntry, 0, len(walker.permSwaps)*len(walker.phaseFlips))
	for _, swap := range walker.permSwaps {
		for _, flip := range walker.phaseFlips {
			if walker.stop != nil && walker.stop() {
				return nil, ErrCancelled
			}
			if walker.limited && walker.remaining == 0 {
				return nil, ErrLimit
			}
			if walker.limited {
				walker.remaining--
			}
			normalized := walker.normalize(current)
			if _, found := seen.Get2(normalized); !found {
				seen.Set(normalized, true)
				result = append(result, dttOrbitEntry{Truth: normalized, Transform: transform, OutputComplement: current&walker.top != 0})
			}
			current = dttFlipTruth(current, walker.variables, flip)
			transform.phases[flip] = !transform.phases[flip]
		}
		current = dttSwapTruth(current, walker.variables, swap)
		transform.permutation[swap], transform.permutation[swap+1] = transform.permutation[swap+1], transform.permutation[swap]
	}
	if current != truth {
		return nil, ErrInvariant
	}
	return result, nil
}

func dttCandidate(kind uint8, left, right, mask uint64) uint64 {
	switch kind {
	case 0:
		return left & right
	case 1:
		return (^left) & right & mask
	case 2:
		return left & (^right) & mask
	case 3:
		return left | right
	case 4:
		return left ^ right
	}
	return 0
}

// EnumerateDttFunctions preserves Dtt_EnumerateLf's source frontier, orbit and
// five-operation order. All state and statistics are owned by the result.
func EnumerateDttFunctions(options DttEnumerationOptions) (*DttEnumeration, error) {
	if options.Variables < 2 || options.Variables > 5 || options.MaxFrontier < 0 || options.MaxFrontier > 31 || options.MaxFunctions < 0 {
		return nil, ErrDimension
	}
	maxFunctions := options.MaxFunctions
	if maxFunctions == 0 {
		maxFunctions = 1 << 20
	}
	maxCandidates := options.MaxCandidates
	if maxCandidates == 0 {
		maxCandidates = ^uint64(0)
	}
	walker, err := newDttOrbitWalker(options.Variables, options.MaxOrbitTransforms, options.Stop)
	if err != nil {
		return nil, err
	}
	result := &DttEnumeration{variables: options.Variables}
	frontiers := make([][]int, options.MaxFrontier+1)
	classIndex := insdict.NewDict[uint64, int]()
	addInitial := func(truth uint64) error {
		orbit, err := walker.orbit(truth)
		if err != nil {
			return err
		}
		canonical := slices.Min(orbit)
		if _, found := classIndex.Get2(canonical); found {
			return ErrInvariant
		}
		classIndex.Set(canonical, len(result.classes))
		result.classes = append(result.classes, DttClass{Canonical: canonical, Frontier: 0, Implementations: 1})
		identity, _ := IdentityDttNPNTransform(options.Variables)
		result.functions = append(result.functions, DttFunction{Truth: truth & walker.mask, Canonical: canonical, Frontier: 0, Fanin0: -1, Fanin1: -1, LeftTransform: identity})
		return nil
	}
	if err = addInitial(0); err != nil {
		return nil, err
	}
	variable := truthVariablePattern(0) & walker.mask
	if err = addInitial(variable); err != nil {
		return nil, err
	}
	frontiers[0] = append(frontiers[0], 1)
	result.levels = append(result.levels, DttLevel{Frontier: 0, Start: 0, End: 2})
	candidates := uint64(0)
	pairs := uint64(0)
	for frontier := 1; frontier <= options.MaxFrontier; frontier++ {
		start := len(result.functions)
		for leftFrontier, rightFrontier := 0, frontier-1; leftFrontier < frontier && leftFrontier <= rightFrontier; leftFrontier++ {
			leftEntries := frontiers[leftFrontier]
			rightEntries := frontiers[rightFrontier]
			for leftPosition, leftIndex := range leftEntries {
				leftOrbit, err := walker.entries(result.functions[leftIndex].Truth)
				if err != nil {
					return nil, err
				}
				rightStart := 0
				if leftFrontier == rightFrontier {
					rightStart = leftPosition
				}
				for _, rightIndex := range rightEntries[rightStart:] {
					pairVisited := insdict.NewDict[uint64, bool]()
					for _, transformedLeft := range leftOrbit {
						batch := make([]uint64, 0, 5)
						for kind := uint8(0); kind < 5; kind++ {
							if options.Stop != nil && options.Stop() {
								return nil, ErrCancelled
							}
							if candidates >= maxCandidates {
								return nil, ErrLimit
							}
							candidates++
							truth := dttCandidate(kind, transformedLeft.Truth, result.functions[rightIndex].Truth, walker.mask)
							orbit, err := walker.orbit(truth)
							if err != nil {
								return nil, err
							}
							canonical := slices.Min(orbit)
							batch = append(batch, canonical)
							class, found := classIndex.Get2(canonical)
							if found {
								record := result.classes[class]
								if record.Frontier < frontier || !options.Multiplicity {
									continue
								}
								if record.Frontier > frontier {
									return nil, ErrInvariant
								}
								if _, visited := pairVisited.Get2(canonical); visited {
									continue
								}
							}
							if len(result.functions) >= maxFunctions {
								return nil, ErrLimit
							}
							nodes := result.functions[leftIndex].Nodes + result.functions[rightIndex].Nodes + 1
							functionIndex := len(result.functions)
							result.functions = append(result.functions, DttFunction{Truth: truth, Canonical: canonical, Frontier: frontier, Nodes: nodes, Fanin0: leftIndex, Fanin1: rightIndex, Type: kind, LeftTransform: transformedLeft.Transform, LeftOutputComplement: transformedLeft.OutputComplement})
							frontiers[frontier] = append(frontiers[frontier], functionIndex)
							if !found {
								classIndex.Set(canonical, len(result.classes))
								result.classes = append(result.classes, DttClass{Canonical: canonical, Frontier: frontier, Implementations: 1})
							} else {
								result.classes[class].Implementations++
							}
						}
						for _, canonical := range batch {
							class, found := classIndex.Get2(canonical)
							if !found || result.classes[class].Frontier != frontier {
								continue
							}
							pairVisited.Set(canonical, true)
						}
					}
					for canonical := range pairVisited.All() {
						class, _ := classIndex.Get2(canonical)
						result.classes[class].Multiplicity++
					}
					pairs++
				}
			}
			if options.Delay {
				rightFrontier = frontier - 1
			} else {
				rightFrontier--
			}
		}
		result.levels = append(result.levels, DttLevel{Frontier: frontier, Start: start, End: len(result.functions), Candidates: candidates, Pairs: pairs})
		if len(frontiers[frontier]) == 0 {
			break
		}
	}
	return result, nil
}
