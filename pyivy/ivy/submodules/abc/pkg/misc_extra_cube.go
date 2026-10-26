package abc

// Two-by-two-by-two cube-state enumeration from src/misc/extra/extraUtilCube.c.

import (
	"fmt"
	"strings"
)

type ExtraCubeState [24]uint8

func ExtraCubeIdentity() ExtraCubeState {
	var state ExtraCubeState
	for i := range state {
		state[i] = uint8(i)
	}
	return state
}

// PermuteExtraCubeState computes result[i] = state[permutation[i]].
func PermuteExtraCubeState(state, permutation ExtraCubeState) (ExtraCubeState, error) {
	var result ExtraCubeState
	seenState, seenPermutation := [24]bool{}, [24]bool{}
	for _, value := range state {
		if value >= 24 || seenState[value] {
			return result, ErrDimension
		}
		seenState[value] = true
	}
	for i, source := range permutation {
		if source >= 24 || seenPermutation[source] {
			return result, ErrDimension
		}
		seenPermutation[source] = true
		result[i] = state[source]
	}
	return result, nil
}

func extraCubeQuarterTurns() [3]ExtraCubeState {
	swaps := [3][9][2]uint8{
		{{3, 5}, {3, 17}, {3, 15}, {1, 6}, {1, 16}, {1, 14}, {2, 4}, {2, 18}, {2, 13}},
		{{2, 14}, {2, 24}, {2, 12}, {3, 13}, {3, 23}, {3, 10}, {1, 15}, {1, 22}, {1, 11}},
		{{1, 10}, {1, 7}, {1, 4}, {3, 12}, {3, 9}, {3, 6}, {2, 11}, {2, 8}, {2, 5}},
	}
	result := [3]ExtraCubeState{}
	for axis := range result {
		result[axis] = ExtraCubeIdentity()
		for _, pair := range swaps[axis] {
			first, second := pair[0]-1, pair[1]-1
			result[axis][first], result[axis][second] = result[axis][second], result[axis][first]
		}
	}
	return result
}

// ExtraCubeGenerators returns powers one, two and three grouped by power,
// matching pFirst[0..8] in the source traversal.
func ExtraCubeGenerators() [9]ExtraCubeState {
	quarter := extraCubeQuarterTurns()
	result := [9]ExtraCubeState{}
	for axis := 0; axis < 3; axis++ {
		result[axis] = quarter[axis]
		result[3+axis], _ = PermuteExtraCubeState(result[axis], quarter[axis])
		result[6+axis], _ = PermuteExtraCubeState(result[3+axis], quarter[axis])
	}
	return result
}

var extraCubeCorner = [24]int8{0, 0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5, 5, -1, -1, -1, 6, 6, 6}
var extraCubeOrientation = [24]int8{0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 1, 2, -1, -1, -1, 0, 1, 2}

// ExtraCubeSignature packs the corner identity and orientation visible at the
// seven source positions into seven five-bit fields.
func ExtraCubeSignature(state ExtraCubeState) (uint64, error) {
	positions := [...]int{0, 3, 6, 9, 12, 15, 21}
	var signature uint64
	for cube, position := range positions {
		sticker := state[position]
		if sticker >= 24 || extraCubeCorner[sticker] < 0 || extraCubeOrientation[sticker] < 0 {
			return 0, ErrDimension
		}
		value := uint64(extraCubeCorner[sticker])<<2 | uint64(extraCubeOrientation[sticker])
		signature ^= value << uint(5*cube)
	}
	return signature, nil
}

func FormatExtraCubeState(state ExtraCubeState) string {
	var builder strings.Builder
	for _, value := range state {
		fmt.Fprintf(&builder, " %2d", value)
	}
	builder.WriteByte('\n')
	return builder.String()
}

type ExtraCubeEnumerationOptions struct {
	MaxStates, MaxDepth int
	SignatureOnly       bool
	ReturnStates        bool
	Stop                func() bool
}

type ExtraCubeEnumeration struct {
	LayerTotals []int
	States      []ExtraCubeState
	Converged   bool
}

func EnumerateExtraCubeStates(options ExtraCubeEnumerationOptions) (ExtraCubeEnumeration, error) {
	fail := func(result ExtraCubeEnumeration, e error) (ExtraCubeEnumeration, error) { return result, e }
	if options.MaxStates < 0 || options.MaxDepth < 0 {
		return fail(ExtraCubeEnumeration{}, ErrDimension)
	}
	if options.MaxStates == 0 {
		options.MaxStates = 1 << 22
	}
	if options.MaxDepth == 0 {
		options.MaxDepth = 100
	}
	states := make([]ExtraCubeState, 0, min(options.MaxStates, 1<<16))
	fullSeen := map[ExtraCubeState]struct{}{}
	signatureSeen := map[uint64]struct{}{}
	add := func(state ExtraCubeState) (bool, error) {
		if options.Stop != nil && options.Stop() {
			return false, ErrCancelled
		}
		if options.SignatureOnly {
			signature, e := ExtraCubeSignature(state)
			if e != nil {
				return false, e
			}
			if _, found := signatureSeen[signature]; found {
				return false, nil
			}
			if len(states) >= options.MaxStates {
				return false, ErrLimit
			}
			signatureSeen[signature] = struct{}{}
		} else {
			if _, found := fullSeen[state]; found {
				return false, nil
			}
			if len(states) >= options.MaxStates {
				return false, ErrLimit
			}
			fullSeen[state] = struct{}{}
		}
		states = append(states, state)
		return true, nil
	}
	identity := ExtraCubeIdentity()
	if _, e := add(identity); e != nil {
		return fail(ExtraCubeEnumeration{}, e)
	}
	result := ExtraCubeEnumeration{LayerTotals: []int{1}}
	generators := ExtraCubeGenerators()
	// The source seeds powers 1,2,3 for each axis before switching to the
	// pFirst[0..8] generator order used by all later frontiers.
	for axis := 0; axis < 3; axis++ {
		for power := 0; power < 3; power++ {
			if _, e := add(generators[3*power+axis]); e != nil {
				if options.ReturnStates {
					result.States = append([]ExtraCubeState(nil), states...)
				}
				return fail(result, e)
			}
		}
	}
	result.LayerTotals = append(result.LayerTotals, len(states))
	begin, end := 1, len(states)
	for depth := 2; depth <= options.MaxDepth; depth++ {
		if begin == end {
			result.Converged = true
			break
		}
		for index := begin; index < end; index++ {
			for _, generator := range generators {
				state, e := PermuteExtraCubeState(states[index], generator)
				if e != nil {
					return fail(result, e)
				}
				if _, e = add(state); e != nil {
					result.LayerTotals = append(result.LayerTotals, len(states))
					if options.ReturnStates {
						result.States = append([]ExtraCubeState(nil), states...)
					}
					return fail(result, e)
				}
			}
		}
		begin, end = end, len(states)
		result.LayerTotals = append(result.LayerTotals, end)
	}
	if begin == end {
		result.Converged = true
	}
	if options.ReturnStates {
		result.States = append([]ExtraCubeState(nil), states...)
	}
	if !result.Converged {
		return fail(result, ErrLimit)
	}
	return result, nil
}
