package abc

// Ternary sequential simulation from src/opt/fsim/fsimTsim.c.

import (
	"fmt"
	"slices"
	"strings"
)

// FsimTernaryOptions bounds retained state exploration. Zero limits select
// conservative defaults.
type FsimTernaryOptions struct {
	MaxStates     int
	MaxStateWords int
	Stop          func() bool
}

// FsimTernaryStateHash preserves Fsim_ManTerStateHash's uint32 overflow and
// 128-prime cycle.
func FsimTernaryStateHash(words []uint32, tableSize int) (int, error) {
	return AigTernaryStateHash(words, tableSize)
}

// FormatFsimTernaryState returns the exact state line emitted by the native
// diagnostic without writing to process-global stdout.
func FormatFsimTernaryState(state []AigTernary) (string, error) {
	var text strings.Builder
	zeros, ones, unknowns := 0, 0, 0
	text.Grow(len(state) + 32)
	for _, value := range state {
		switch value {
		case AigTernaryZero:
			text.WriteByte('0')
			zeros++
		case AigTernaryOne:
			text.WriteByte('1')
			ones++
		case AigTernaryUnknown:
			text.WriteByte('x')
			unknowns++
		default:
			return "", ErrInvariant
		}
	}
	fmt.Fprintf(&text, " (0=%5d, 1=%5d, x=%5d)\n", zeros, ones, unknowns)
	return text.String(), nil
}

// FsimTernaryStates returns the zero-initialized sequential ternary trajectory,
// including the final repeated state as in Fsim_ManTerSimulate. Primary inputs
// are unknown in every frame. Returned states own independent storage.
func (graph *AigManager) FsimTernaryStates(options FsimTernaryOptions) ([][]AigTernary, error) {
	if graph == nil || options.MaxStates < 0 || options.MaxStateWords < 0 {
		return nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, err
	}
	if graph.registers == 0 {
		return nil, ErrDimension
	}
	for _, object := range graph.objects {
		if object.Type == AigBuffer || object.Type == AigXOR {
			return nil, ErrUnsupported
		}
	}
	maximumStates := options.MaxStates
	if maximumStates == 0 {
		maximumStates = 1 << 20
	}
	maximumWords := options.MaxStateWords
	if maximumWords == 0 {
		maximumWords = 1 << 20
	}
	wordsPerState := (graph.registers + 15) / 16
	primaryInputs := graph.PrimaryInputCount()
	primaryOutputs := graph.PrimaryOutputCount()
	registerIDs := graph.cis[primaryInputs:]
	latchInputs := graph.cos[primaryOutputs:]
	state := make([]AigTernary, graph.registers)
	for index := range state {
		state[index] = AigTernaryZero
	}
	states := make([][]AigTernary, 0)
	packedStates := make([][]uint32, 0)
	buckets := make([][]int, 503)
	visit := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	for {
		if err := visit(); err != nil {
			return nil, err
		}
		if len(states) >= maximumStates || uint64(wordsPerState)*uint64(len(states)+1) > uint64(maximumWords) {
			return nil, ErrLimit
		}
		owned := slices.Clone(state)
		packed := packFsimTernaryState(owned)
		states = append(states, owned)
		packedStates = append(packedStates, packed)
		hash, _ := FsimTernaryStateHash(packed, len(buckets))
		repeated := false
		for _, index := range buckets[hash] {
			if slices.Equal(packedStates[index], packed) {
				repeated = true
				break
			}
		}
		if repeated {
			return states, nil
		}
		buckets[hash] = append(buckets[hash], len(states)-1)

		values := make([]AigTernary, len(graph.objects))
		values[0] = AigTernaryOne
		for _, id := range graph.cis[:primaryInputs] {
			values[id] = AigTernaryUnknown
		}
		for index, id := range registerIDs {
			values[id] = state[index]
		}
		for id, object := range graph.objects {
			if err := visit(); err != nil {
				return nil, err
			}
			if object.Type == AigAND {
				values[id] = aigTernaryAnd(aigTernaryRead(values, object.Fanin0), aigTernaryRead(values, object.Fanin1))
			}
		}
		next := make([]AigTernary, graph.registers)
		for index, output := range latchInputs {
			next[index] = aigTernaryRead(values, graph.objects[output].Fanin0)
		}
		state = next
	}
}

func packFsimTernaryState(state []AigTernary) []uint32 {
	packed := make([]uint32, (len(state)+15)/16)
	for index, value := range state {
		packed[index>>4] |= uint32(value) << uint((index&15)<<1)
	}
	return packed
}
