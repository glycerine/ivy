package abc

// Fast sequential-AIG structural signatures from src/aig/saig/saigIsoFast.c.

import (
	"slices"
	"sort"
)

type SAIGIsoFastOptions struct {
	MaxObjects int
	MaxVisits  int
	MaxRoots   int
	MaxEntries int
	MaxWork    int
	Stop       func() bool
}

type SAIGIsoFastOutput struct {
	Output    int
	Signature []uint32
}

type SAIGIsoFastResult struct {
	Outputs []SAIGIsoFastOutput
	Classes [][]int
	Unique  int
}

type saigIsoFastBudget struct {
	options SAIGIsoFastOptions
	work    int
	visits  int
}

func normalizeSAIGIsoFastOptions(options SAIGIsoFastOptions) (SAIGIsoFastOptions, error) {
	if options.MaxObjects < 0 || options.MaxVisits < 0 || options.MaxRoots < 0 || options.MaxEntries < 0 || options.MaxWork < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 8 << 20
	}
	if options.MaxRoots == 0 {
		options.MaxRoots = 1 << 20
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 32 << 20
	}
	return options, nil
}

func (budget *saigIsoFastBudget) step(amount int) error {
	if amount < 0 || budget.work > budget.options.MaxWork-amount {
		return ErrLimit
	}
	budget.work += amount
	if budget.options.Stop != nil && budget.options.Stop() {
		return ErrCancelled
	}
	return nil
}

func (budget *saigIsoFastBudget) visit() error {
	if budget.visits == budget.options.MaxVisits {
		return ErrLimit
	}
	budget.visits++
	return budget.step(1)
}

type saigIsoFastData struct {
	faninNegative  uint8
	fanoutNegative uint8
	fanoutPositive uint8
	firstLevel     uint8
	secondLevel    uint8
	level          uint8
	visited        bool
}

func (data saigIsoFastData) packed() uint16 {
	return uint16(data.faninNegative&7) |
		uint16(data.fanoutNegative&3)<<3 |
		uint16(data.fanoutPositive&3)<<5 |
		uint16(data.firstLevel&7)<<7 |
		uint16(data.secondLevel&7)<<10 |
		uint16(data.level&7)<<13
}

func (data *saigIsoFastData) addFanout(complemented bool) {
	if complemented {
		data.fanoutNegative = (data.fanoutNegative + 1) & 3
	} else {
		data.fanoutPositive = (data.fanoutPositive + 1) & 3
	}
}

type saigIsoFastFrame struct {
	id         AigID
	complement bool
	exit       bool
}

func saigIsoFastSignature(graph *AigManager, outputIndex int, ciOrdinal []int, budget *saigIsoFastBudget) ([]uint32, error) {
	data := make([]saigIsoFastData, graph.ObjectCount())
	visited := make([]AigID, 0)
	roots := []AigID{graph.cos[outputIndex]}
	for rootIndex := 0; rootIndex < len(roots); rootIndex++ {
		if err := budget.step(1); err != nil {
			return nil, err
		}
		root := graph.objects[roots[rootIndex]]
		literal := root.Fanin0
		if literal.ID() == 0 {
			continue
		}
		stack := []saigIsoFastFrame{{id: literal.ID(), complement: literal.Complemented()}}
		for len(stack) != 0 {
			frame := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if frame.exit {
				visited = append(visited, frame.id)
				continue
			}
			if err := budget.visit(); err != nil {
				return nil, err
			}
			entry := &data[frame.id]
			if entry.visited {
				entry.addFanout(frame.complement)
				continue
			}
			entry.visited = true
			entry.addFanout(frame.complement)
			object := graph.objects[frame.id]
			entry.level = uint8(object.Level) & 7
			stack = append(stack, saigIsoFastFrame{id: frame.id, exit: true})
			switch object.Type {
			case AigAND:
				negative0, negative1 := object.Fanin0.Complemented(), object.Fanin1.Complemented()
				if negative0 {
					entry.faninNegative++
				}
				if negative1 {
					entry.faninNegative++
				}
				level0 := object.Level - graph.objects[object.Fanin0.ID()].Level
				level1 := object.Level - graph.objects[object.Fanin1.ID()].Level
				if negative0 && !negative1 || negative0 == negative1 && graph.objects[object.Fanin1.ID()].Level < graph.objects[object.Fanin0.ID()].Level {
					level0, level1 = level1, level0
				}
				entry.firstLevel, entry.secondLevel = uint8(level0)&7, uint8(level1)&7
				stack = append(stack,
					saigIsoFastFrame{id: object.Fanin1.ID(), complement: object.Fanin1.Complemented()},
					saigIsoFastFrame{id: object.Fanin0.ID(), complement: object.Fanin0.Complemented()},
				)
			case AigCI:
				ordinal := ciOrdinal[frame.id]
				if ordinal < 0 {
					return nil, ErrInvariant
				}
				if ordinal >= graph.PrimaryInputCount() {
					entry.firstLevel = 1
					register := ordinal - graph.PrimaryInputCount()
					if len(roots) == budget.options.MaxRoots {
						return nil, ErrLimit
					}
					roots = append(roots, graph.cos[graph.PrimaryOutputCount()+register])
				}
			default:
				return nil, ErrUnsupported
			}
		}
	}
	if len(visited) > budget.options.MaxEntries {
		return nil, ErrLimit
	}
	counters := make([]int, 1<<16)
	places := make([]int, 0)
	for _, id := range visited {
		value := int(data[id].packed())
		if counters[value] == 0 {
			places = append(places, value)
		}
		counters[value]++
	}
	if len(places) > budget.options.MaxEntries {
		return nil, ErrLimit
	}
	frequencies := make([]int, len(places))
	for index, place := range places {
		frequencies[index] = counters[place]
	}
	permutation := IntCostPermutation(frequencies)
	result := make([]uint32, 0, len(places))
	for index := len(permutation) - 1; index >= 0; index-- {
		position := permutation[index]
		result = append(result, uint32(frequencies[position])<<16|uint32(places[position]))
	}
	return result, nil
}

func saigIsoFastCompare(left, right []uint32) int {
	if len(left) < len(right) {
		return -1
	}
	if len(left) > len(right) {
		return 1
	}
	for index := range left {
		for shift := uint(0); shift < 32; shift += 8 {
			first := byte(left[index] >> shift)
			second := byte(right[index] >> shift)
			if first < second {
				return -1
			}
			if first > second {
				return 1
			}
		}
	}
	return 0
}

// SAIGDetectIsomorphicOutputsFast preserves the source's packed 16-bit node
// data, modulo-four fanout fields, latch traversal, ABC cost permutation and
// little-endian memcmp ordering. It returns the classes the C experiment
// prints and frees.
func SAIGDetectIsomorphicOutputsFast(graph *AigManager, options SAIGIsoFastOptions) (SAIGIsoFastResult, error) {
	options, err := normalizeSAIGIsoFastOptions(options)
	if err != nil {
		return SAIGIsoFastResult{}, err
	}
	if graph == nil {
		return SAIGIsoFastResult{}, ErrDimension
	}
	if err = graph.Check(); err != nil {
		return SAIGIsoFastResult{}, err
	}
	if graph.PrimaryOutputCount() == 0 {
		return SAIGIsoFastResult{}, ErrDimension
	}
	if graph.ObjectCount() > options.MaxObjects || graph.PrimaryOutputCount() > options.MaxEntries {
		return SAIGIsoFastResult{}, ErrLimit
	}
	for _, object := range graph.objects {
		if object.Type == AigBuffer {
			return SAIGIsoFastResult{}, ErrUnsupported
		}
	}
	budget := &saigIsoFastBudget{options: options}
	ciOrdinal := make([]int, graph.ObjectCount())
	for index := range ciOrdinal {
		ciOrdinal[index] = -1
	}
	for index, id := range graph.cis {
		ciOrdinal[id] = index
	}
	result := SAIGIsoFastResult{Outputs: make([]SAIGIsoFastOutput, graph.PrimaryOutputCount())}
	combined := make([][]uint32, graph.PrimaryOutputCount())
	for output := 0; output < graph.PrimaryOutputCount(); output++ {
		signature, signatureErr := saigIsoFastSignature(graph, output, ciOrdinal, budget)
		if signatureErr != nil {
			return SAIGIsoFastResult{}, signatureErr
		}
		result.Outputs[output] = SAIGIsoFastOutput{Output: output, Signature: slices.Clone(signature)}
		combined[output] = append(slices.Clone(signature), uint32(output))
	}
	sort.Slice(combined, func(i, j int) bool { return saigIsoFastCompare(combined[i], combined[j]) < 0 })
	for _, info := range combined {
		output := int(info[len(info)-1])
		signature := info[:len(info)-1]
		if len(result.Classes) == 0 {
			result.Classes = append(result.Classes, []int{output})
			continue
		}
		previousOutput := result.Classes[len(result.Classes)-1][0]
		if saigIsoFastCompare(result.Outputs[previousOutput].Signature, signature) != 0 {
			result.Classes = append(result.Classes, []int{output})
		} else {
			result.Classes[len(result.Classes)-1] = append(result.Classes[len(result.Classes)-1], output)
		}
	}
	for _, class := range result.Classes {
		if len(class) == 1 {
			result.Unique++
		}
	}
	return result, nil
}
