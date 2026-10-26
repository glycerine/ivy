package abc

// Sequential reachability cleanup from src/aig/aig/aigScl.c.

import "slices"

type AigSequentialCleanupOptions struct {
	MaxObjects int // zero: one million objects
	Stop       func() bool
}

type AigSequentialCleanup struct {
	Graph          *AigManager
	Registers      []int
	RemovedObjects int
}

func aigSequentialReachable(graph *AigManager, options AigSequentialCleanupOptions) ([]bool, []int, error) {
	if graph == nil || options.MaxObjects < 0 {
		return nil, nil, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return nil, nil, err
	}
	maximum := options.MaxObjects
	if maximum == 0 {
		maximum = 1 << 20
	}
	if len(graph.objects) > maximum {
		return nil, nil, ErrLimit
	}
	if graph.equivs != nil {
		return nil, nil, ErrUnsupported
	}
	visit := func() error {
		if options.Stop != nil && options.Stop() {
			return ErrCancelled
		}
		return nil
	}
	marked := make([]bool, len(graph.objects))
	marked[0] = true
	primaryInputs := len(graph.cis) - graph.registers
	primaryOutputs := len(graph.cos) - graph.registers
	registerIndex := make([]int, len(graph.objects))
	for id := range registerIndex {
		registerIndex[id] = -1
	}
	for index, id := range graph.cis[primaryInputs:] {
		registerIndex[id] = index
	}
	for _, id := range graph.cis[:primaryInputs] {
		marked[id] = true
	}
	stack := slices.Clone(graph.cos[:primaryOutputs])
	for len(stack) != 0 {
		if err := visit(); err != nil {
			return nil, nil, err
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if marked[id] {
			continue
		}
		marked[id] = true
		object := graph.objects[id]
		switch object.Type {
		case AigConstant1:
		case AigCI:
			index := registerIndex[id]
			if index < 0 || index >= graph.registers {
				return nil, nil, ErrInvariant
			}
			stack = append(stack, graph.cos[primaryOutputs+index])
		case AigCO, AigBuffer:
			stack = append(stack, object.Fanin0.ID())
		case AigAND:
			stack = append(stack, object.Fanin1.ID(), object.Fanin0.ID())
		default:
			return nil, nil, ErrInvariant
		}
	}
	registers := make([]int, 0, graph.registers)
	for index, id := range graph.cis[primaryInputs:] {
		if marked[id] {
			registers = append(registers, index)
		}
	}
	return marked, registers, nil
}

// SequentialCleanup implements Aig_ManSeqCleanup as an owned atomic
// transformation. Primary terminals remain in order; retained registers keep
// source order and RemovedObjects counts swept internal objects only.
func (graph *AigManager) SequentialCleanup(options AigSequentialCleanupOptions) (AigSequentialCleanup, error) {
	marked, registers, err := aigSequentialReachable(graph, options)
	if err != nil {
		return AigSequentialCleanup{}, err
	}
	work := graph.Clone()
	primaryInputs := len(work.cis) - work.registers
	primaryOutputs := len(work.cos) - work.registers
	kept := make([]bool, work.registers)
	for _, index := range registers {
		kept[index] = true
	}
	newOutputs := slices.Clone(work.cos[:primaryOutputs])
	for index, output := range work.cos[primaryOutputs:] {
		if options.Stop != nil && options.Stop() {
			return AigSequentialCleanup{}, ErrCancelled
		}
		if kept[index] {
			newOutputs = append(newOutputs, output)
			continue
		}
		driver := work.objects[output].Fanin0.ID()
		if work.objects[driver].RefCount == 0 {
			return AigSequentialCleanup{}, ErrInvariant
		}
		work.objects[driver].RefCount--
		work.objects[output] = AigObject{}
	}
	work.cos = newOutputs
	work.registers = len(registers)
	removed := work.Cleanup()
	newInputs := slices.Clone(work.cis[:primaryInputs])
	for index, input := range work.cis[primaryInputs:] {
		if kept[index] {
			newInputs = append(newInputs, input)
			continue
		}
		if work.objects[input].RefCount != 0 || marked[input] {
			return AigSequentialCleanup{}, ErrInvariant
		}
		work.objects[input] = AigObject{}
	}
	work.cis = newInputs
	if err = work.Check(); err != nil {
		return AigSequentialCleanup{}, err
	}
	return AigSequentialCleanup{Graph: work, Registers: slices.Clone(registers), RemovedObjects: removed}, nil
}

// SequentialCleanupBasic implements Aig_ManSeqCleanupBasic. It retains every
// register slot, patches unreachable latch inputs to constant zero and sweeps
// logic made dangling by those patches.
func (graph *AigManager) SequentialCleanupBasic(options AigSequentialCleanupOptions) (AigSequentialCleanup, error) {
	marked, registers, err := aigSequentialReachable(graph, options)
	if err != nil {
		return AigSequentialCleanup{}, err
	}
	work := graph.Clone()
	primaryOutputs := len(work.cos) - work.registers
	for _, output := range work.cos[primaryOutputs:] {
		if marked[output] {
			continue
		}
		_, patchErr := work.PatchOutput(output, AigConst0, AigObjectOptions{MaxObjects: options.MaxObjects, Stop: options.Stop})
		if patchErr != nil {
			return AigSequentialCleanup{}, patchErr
		}
	}
	// Native returns only the final cleanup count; recursive deletion performed
	// while patching unused latch inputs is intentionally not included.
	removed := work.Cleanup()
	if err = work.Check(); err != nil {
		return AigSequentialCleanup{}, err
	}
	return AigSequentialCleanup{Graph: work, Registers: slices.Clone(registers), RemovedObjects: removed}, nil
}
