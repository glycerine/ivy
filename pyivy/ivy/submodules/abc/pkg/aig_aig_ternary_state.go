package abc

// Sequential ternary-state simulation from src/aig/aig/aigTsim.c.

import "slices"

var aigTernaryHashPrimes = [...]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
	1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
	2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
	2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
	3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
	3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
	4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
	4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
	6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
	6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
	7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
	8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

// AigTernaryStateHash preserves Aig_TsiStateHash's uint32 overflow and prime
// cycle. It does not mutate words.
func AigTernaryStateHash(words []uint32, tableSize int) (int, error) {
	if tableSize <= 0 {
		return 0, ErrDimension
	}
	var hash uint32
	for index, word := range words {
		hash ^= word * aigTernaryHashPrimes[index&127]
	}
	return int(hash % uint32(tableSize)), nil
}

type AigTernaryStateOptions struct {
	MaxRounds     int // zero: native 1,000 rounds
	MaxObjects    int // zero: one million allocated IDs
	MaxStateWords int // zero: one million uint32 words across retained states
	Stop          func() bool
}

type AigTernaryConstants struct {
	Rounds    int
	States    int
	Constants int
	Aggregate []uint32
	Registers []AigTernary
	Mapping   []AigLit
}

func aigTernaryRead(values []AigTernary, literal AigLit) AigTernary {
	value := values[literal.ID()]
	if literal.Complemented() {
		return value.Not()
	}
	return value
}

func aigTernaryPack(values []AigTernary, registers []AigID) []uint32 {
	state := make([]uint32, (len(registers)+15)/16)
	for index, id := range registers {
		value := uint32(values[id])
		bit := 2 * index
		state[bit>>5] |= value << uint(bit&31)
	}
	return state
}

func aigTernaryPackedValue(state []uint32, index int) AigTernary {
	bit := 2 * index
	return AigTernary((state[bit>>5] >> uint(bit&31)) & 3)
}

// SequentialTernaryConstants implements Aig_ManTernarySimulate without
// borrowing nCuts scratch or printing. Mapping is nil when no register is
// constant across the reached ternary states, matching native's null result.
func (graph *AigManager) SequentialTernaryConstants(options AigTernaryStateOptions) (AigTernaryConstants, error) {
	result := AigTernaryConstants{}
	if graph == nil || options.MaxRounds < 0 || options.MaxObjects < 0 || options.MaxStateWords < 0 {
		return result, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return result, err
	}
	maximumObjects := options.MaxObjects
	if maximumObjects == 0 {
		maximumObjects = 1 << 20
	}
	if len(graph.objects) > maximumObjects {
		return result, ErrLimit
	}
	maximumRounds := options.MaxRounds
	if maximumRounds == 0 {
		maximumRounds = 1000
	}
	maximumWords := options.MaxStateWords
	if maximumWords == 0 {
		maximumWords = 1 << 20
	}
	wordsPerState := (graph.registers + 15) / 16
	for _, object := range graph.objects {
		if object.Type == AigBuffer || object.Type == AigXOR {
			return result, ErrUnsupported
		}
	}
	visit := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	primaryInputs := len(graph.cis) - graph.registers
	primaryOutputs := len(graph.cos) - graph.registers
	registers := graph.cis[primaryInputs:]
	latchInputs := graph.cos[primaryOutputs:]
	values := make([]AigTernary, len(graph.objects))
	values[0] = AigTernaryOne
	for _, id := range graph.cis[:primaryInputs] {
		values[id] = AigTernaryUnknown
	}
	for _, id := range registers {
		values[id] = AigTernaryZero
	}
	states := make([][]uint32, 0)
	buckets := make([][]int, 503)
	for round := 0; round < maximumRounds; round++ {
		if err := visit(); err != nil {
			return AigTernaryConstants{}, err
		}
		if uint64(wordsPerState)*uint64(len(states)+1) > uint64(maximumWords) {
			return AigTernaryConstants{}, ErrLimit
		}
		state := aigTernaryPack(values, registers)
		states = append(states, state)
		hash, _ := AigTernaryStateHash(state, len(buckets))
		repeated := false
		for _, stateIndex := range buckets[hash] {
			if slices.Equal(states[stateIndex], state) {
				repeated = true
				break
			}
		}
		if repeated {
			result.Rounds = round
			result.States = len(states)
			for _, retained := range states {
				if result.Aggregate == nil {
					result.Aggregate = slices.Clone(retained)
					continue
				}
				for index := range retained {
					result.Aggregate[index] |= retained[index]
				}
			}
			result.Registers = make([]AigTernary, graph.registers)
			result.Mapping = make([]AigLit, 0, len(graph.cis))
			for _, id := range graph.cis[:primaryInputs] {
				result.Mapping = append(result.Mapping, AigLit(id)<<1)
			}
			for index, id := range registers {
				value := aigTernaryPackedValue(result.Aggregate, index)
				result.Registers[index] = value
				switch value {
				case AigTernaryZero:
					result.Mapping = append(result.Mapping, AigConst0)
					result.Constants++
				case AigTernaryOne:
					result.Mapping = append(result.Mapping, AigConst1)
					result.Constants++
				case AigTernaryUnknown:
					result.Mapping = append(result.Mapping, AigLit(id)<<1)
				default:
					return AigTernaryConstants{}, ErrInvariant
				}
			}
			if result.Constants == 0 {
				result.Mapping = nil
			}
			return result, nil
		}
		buckets[hash] = append(buckets[hash], len(states)-1)
		for id, object := range graph.objects {
			if err := visit(); err != nil {
				return AigTernaryConstants{}, err
			}
			if object.Type == AigAND {
				values[id] = aigTernaryAnd(aigTernaryRead(values, object.Fanin0), aigTernaryRead(values, object.Fanin1))
			}
		}
		nextValues := make([]AigTernary, len(latchInputs))
		for index, output := range latchInputs {
			nextValues[index] = aigTernaryRead(values, graph.objects[output].Fanin0)
		}
		for index, next := range nextValues {
			id := registers[index]
			if round < 300 {
				values[id] = next
			} else if values[id] != next {
				values[id] = AigTernaryUnknown
			}
		}
	}
	return AigTernaryConstants{}, ErrLimit
}
