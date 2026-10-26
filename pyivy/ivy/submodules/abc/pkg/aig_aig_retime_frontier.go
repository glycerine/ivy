package abc

// Forward-frontier retiming from src/aig/aig/aigRetF.c.

import "slices"

type AigRetimeOptions struct {
	MaxSteps   int // zero: one million retiming steps
	MaxObjects int // zero: one million intermediate objects
	Stop       func() bool
}

type AigRetimeResult struct {
	Graph *AigManager
	Steps int
}

type aigRetimeNet struct {
	objects        []AigObject
	cis, cos       []AigID
	registers      int
	primaryInputs  int
	primaryOutputs int
	markA, markB   []bool
	pair           []AigID
}

func newAigRetimeNet(graph *AigManager) *aigRetimeNet {
	count := len(graph.objects)
	return &aigRetimeNet{
		objects: slices.Clone(graph.objects), cis: slices.Clone(graph.cis), cos: slices.Clone(graph.cos),
		registers: graph.registers, primaryInputs: len(graph.cis) - graph.registers,
		primaryOutputs: len(graph.cos) - graph.registers, markA: make([]bool, count),
		markB: make([]bool, count), pair: make([]AigID, count),
	}
}

func (net *aigRetimeNet) grow(object AigObject, maximum int) (AigID, error) {
	if len(net.objects) == maximum {
		return InvalidAigID, ErrLimit
	}
	id := AigID(len(net.objects))
	net.objects = append(net.objects, object)
	net.markA = append(net.markA, false)
	net.markB = append(net.markB, false)
	net.pair = append(net.pair, InvalidAigID)
	return id, nil
}

func (net *aigRetimeNet) reachMark(root AigID, traversed []bool) (bool, error) {
	type frame struct {
		id    AigID
		stage uint8
	}
	stack := []frame{{id: root}}
	last := false
	for len(stack) != 0 {
		top := &stack[len(stack)-1]
		object := net.objects[top.id]
		switch top.stage {
		case 0:
			if net.markB[top.id] {
				last = true
				stack = stack[:len(stack)-1]
				continue
			}
			if object.Type == AigCI || object.Type == AigConstant1 {
				last = false
				stack = stack[:len(stack)-1]
				continue
			}
			if traversed[top.id] {
				last = net.markB[top.id]
				stack = stack[:len(stack)-1]
				continue
			}
			if object.Type != AigCO && object.Type != AigBuffer && object.Type != AigAND {
				return false, ErrInvariant
			}
			traversed[top.id] = true
			top.stage = 1
			stack = append(stack, frame{id: object.Fanin0.ID()})
		case 1:
			if last {
				net.markB[top.id] = true
				stack = stack[:len(stack)-1]
				continue
			}
			if object.Type != AigAND {
				last = false
				stack = stack[:len(stack)-1]
				continue
			}
			top.stage = 2
			stack = append(stack, frame{id: object.Fanin1.ID()})
		case 2:
			if last {
				net.markB[top.id] = true
			}
			last = net.markB[top.id]
			stack = stack[:len(stack)-1]
		}
	}
	return last, nil
}

func (net *aigRetimeNet) markRetimable(stop func() bool) error {
	for _, id := range net.cis[:net.primaryInputs] {
		net.markB[id] = true
	}
	for index := 0; index < net.registers; index++ {
		lo := net.cis[net.primaryInputs+index]
		li := net.cos[net.primaryOutputs+index]
		net.pair[lo], net.pair[li] = li, lo
	}
	for {
		changed := false
		traversed := make([]bool, len(net.objects))
		for index, id := range net.cos {
			if stop != nil && stop() {
				return ErrCancelled
			}
			if net.markB[id] {
				continue
			}
			reached, err := net.reachMark(id, traversed)
			if err != nil {
				return err
			}
			if reached && index >= net.primaryOutputs {
				net.markB[net.cis[net.primaryInputs+index-net.primaryOutputs]] = true
				changed = true
			}
		}
		if !changed {
			return nil
		}
	}
}

func (net *aigRetimeNet) real(literal AigLit) (AigLit, error) {
	for steps := 0; ; steps++ {
		if uint64(literal.ID()) >= uint64(len(net.objects)) || steps > len(net.objects) {
			return InvalidAigLit, ErrInvariant
		}
		object := net.objects[literal.ID()]
		if object.Type != AigBuffer {
			return literal, nil
		}
		literal = object.Fanin0.NotIf(literal.Complemented())
	}
}

func (net *aigRetimeNet) and(left, right AigLit, maximum int) (AigLit, error) {
	switch {
	case left == right:
		return left, nil
	case left == right.Not(), left == AigConst0, right == AigConst0:
		return AigConst0, nil
	case left == AigConst1:
		return right, nil
	case right == AigConst1:
		return left, nil
	}
	if left.ID() > right.ID() {
		left, right = right, left
	}
	id, err := net.grow(AigObject{Type: AigAND, Fanin0: left, Fanin1: right}, maximum)
	if err != nil {
		return InvalidAigLit, err
	}
	return AigLit(id) << 1, nil
}

func (net *aigRetimeNet) retime(options AigRetimeOptions) (int, error) {
	for _, id := range net.cis[net.primaryInputs:] {
		net.markA[id] = true
	}
	steps := 0
	for changed := true; changed; {
		changed = false
		for id := 1; id < len(net.objects); id++ {
			if options.Stop != nil && options.Stop() {
				return 0, ErrCancelled
			}
			object := net.objects[id]
			if object.Type != AigAND || !net.markB[id] {
				continue
			}
			left, err := net.real(object.Fanin0)
			if err != nil {
				return 0, err
			}
			right, err := net.real(object.Fanin1)
			if err != nil {
				return 0, err
			}
			if !net.markA[left.ID()] || !net.markA[right.ID()] {
				continue
			}
			complement := left.Complemented() && right.Complemented()
			leftInput, rightInput := net.pair[left.ID()], net.pair[right.ID()]
			if leftInput == InvalidAigID || rightInput == InvalidAigID || net.objects[leftInput].Type != AigCO || net.objects[rightInput].Type != AigCO {
				return 0, ErrInvariant
			}
			leftDriver := net.objects[leftInput].Fanin0.NotIf(left.Complemented())
			rightDriver := net.objects[rightInput].Fanin0.NotIf(right.Complemented())
			newNode, err := net.and(leftDriver, rightDriver, options.MaxObjects)
			if err != nil {
				return 0, err
			}
			net.markB[newNode.ID()] = true
			newLO, err := net.grow(AigObject{Type: AigCI, Fanin0: InvalidAigLit, Fanin1: InvalidAigLit}, options.MaxObjects)
			if err != nil {
				return 0, err
			}
			net.markA[newLO] = true
			newLI, err := net.grow(AigObject{Type: AigCO, Fanin0: newNode.NotIf(complement), Fanin1: InvalidAigLit}, options.MaxObjects)
			if err != nil {
				return 0, err
			}
			net.cis = append(net.cis, newLO)
			net.cos = append(net.cos, newLI)
			net.pair[newLO], net.pair[newLI] = newLI, newLO
			net.registers++
			net.objects[id] = AigObject{Type: AigBuffer, Fanin0: (AigLit(newLO) << 1).NotIf(complement), Fanin1: InvalidAigLit}
			steps++
			changed = true
			if steps == options.MaxSteps {
				return steps, nil
			}
		}
	}
	return steps, nil
}

func (net *aigRetimeNet) normalize(source *AigManager, stop func() bool) (*AigManager, error) {
	registerIndex := make([]int, len(net.objects))
	for id := range registerIndex {
		registerIndex[id] = -1
	}
	for index, id := range net.cis[net.primaryInputs:] {
		registerIndex[id] = index
	}
	marked := make([]bool, len(net.objects))
	marked[0] = true
	for _, id := range net.cis[:net.primaryInputs] {
		marked[id] = true
	}
	stack := slices.Clone(net.cos[:net.primaryOutputs])
	for len(stack) != 0 {
		if stop != nil && stop() {
			return nil, ErrCancelled
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if marked[id] {
			continue
		}
		marked[id] = true
		object := net.objects[id]
		switch object.Type {
		case AigCI:
			index := registerIndex[id]
			if index < 0 {
				return nil, ErrInvariant
			}
			stack = append(stack, net.cos[net.primaryOutputs+index])
		case AigCO, AigBuffer:
			stack = append(stack, object.Fanin0.ID())
		case AigAND:
			stack = append(stack, object.Fanin1.ID(), object.Fanin0.ID())
		case AigConstant1:
		default:
			return nil, ErrInvariant
		}
	}
	kept := make([]bool, net.registers)
	for index, id := range net.cis[net.primaryInputs:] {
		kept[index] = marked[id]
	}
	target := NewAigManager()
	target.Name, target.Spec = source.Name, source.Spec
	mapping := make([]AigLit, len(net.objects))
	for id := range mapping {
		mapping[id] = InvalidAigLit
	}
	mapping[0] = AigConst1
	for index, id := range net.cis {
		if index >= net.primaryInputs && !kept[index-net.primaryInputs] {
			continue
		}
		literal, err := target.CreateCI()
		if err != nil {
			return nil, err
		}
		mapping[id] = literal
	}
	copyRoot := func(root AigID) error {
		type event struct {
			id   AigID
			exit bool
		}
		state := make([]uint8, len(net.objects))
		work := []event{{id: root}}
		for len(work) != 0 {
			if stop != nil && stop() {
				return ErrCancelled
			}
			e := work[len(work)-1]
			work = work[:len(work)-1]
			if mapping[e.id].Valid() {
				continue
			}
			object := net.objects[e.id]
			if e.exit {
				var result AigLit
				var err error
				switch object.Type {
				case AigBuffer:
					result = mapping[object.Fanin0.ID()].NotIf(object.Fanin0.Complemented())
				case AigAND:
					left := mapping[object.Fanin0.ID()].NotIf(object.Fanin0.Complemented())
					right := mapping[object.Fanin1.ID()].NotIf(object.Fanin1.Complemented())
					result, err = target.And(left, right)
				default:
					return ErrInvariant
				}
				if err != nil {
					return err
				}
				mapping[e.id] = result
				state[e.id] = 2
				continue
			}
			if state[e.id] == 1 {
				return ErrInvariant
			}
			if object.Type != AigBuffer && object.Type != AigAND {
				return ErrInvariant
			}
			state[e.id] = 1
			work = append(work, event{id: e.id, exit: true})
			if object.Type == AigAND && !mapping[object.Fanin1.ID()].Valid() {
				work = append(work, event{id: object.Fanin1.ID()})
			}
			if !mapping[object.Fanin0.ID()].Valid() {
				work = append(work, event{id: object.Fanin0.ID()})
			}
		}
		return nil
	}
	for index, id := range net.cos {
		if index >= net.primaryOutputs && !kept[index-net.primaryOutputs] {
			continue
		}
		driver := net.objects[id].Fanin0
		if err := copyRoot(driver.ID()); err != nil {
			return nil, err
		}
		if _, err := target.CreateCO(mapping[driver.ID()].NotIf(driver.Complemented())); err != nil {
			return nil, err
		}
	}
	retained := 0
	for _, yes := range kept {
		if yes {
			retained++
		}
	}
	if err := target.SetRegisterCount(retained); err != nil {
		return nil, err
	}
	if err := target.SetConstraintCount(source.constraints); err != nil {
		return nil, err
	}
	if err := target.Check(); err != nil {
		return nil, err
	}
	return target, nil
}

// RetimeFrontier implements the intended Aig_ManRetimeFrontier transform. It
// avoids native's forward-buffer/ordered-duplication crash by normalizing the
// private intermediate netlist with dependency-driven iterative duplication.
func (graph *AigManager) RetimeFrontier(options AigRetimeOptions) (AigRetimeResult, error) {
	if graph == nil || options.MaxSteps < 0 || options.MaxObjects < 0 {
		return AigRetimeResult{}, ErrDimension
	}
	if err := graph.Check(); err != nil {
		return AigRetimeResult{}, err
	}
	if graph.registers == 0 || graph.equivs != nil {
		return AigRetimeResult{}, ErrUnsupported
	}
	if options.MaxSteps == 0 {
		options.MaxSteps = 1 << 20
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if len(graph.objects) > options.MaxObjects {
		return AigRetimeResult{}, ErrLimit
	}
	net := newAigRetimeNet(graph)
	if err := net.markRetimable(options.Stop); err != nil {
		return AigRetimeResult{}, err
	}
	steps, err := net.retime(options)
	if err != nil {
		return AigRetimeResult{}, err
	}
	result, err := net.normalize(graph, options.Stop)
	if err != nil {
		return AigRetimeResult{}, err
	}
	return AigRetimeResult{Graph: result, Steps: steps}, nil
}
