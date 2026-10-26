package abc

// Choice-chain traversal and levels from aig/aig/aigDfs.c.

import "slices"

// AigChoiceLevels is an owned choice-aware level assignment. Levels is indexed
// by object ID and Maximum is the largest CO or dangling box-output CI level.
type AigChoiceLevels struct {
	Levels  []uint32
	Maximum uint32
}

// SetChoiceChains atomically replaces all AIG equivalence chains. Every chain
// is ordered representative first, followed by alternatives. Structural and
// chain validity are checked; functional equivalence is the caller's proof
// obligation, matching native pEquivs metadata.
func (g *AigManager) SetChoiceChains(chains [][]AigID, o AigTraversalOptions) error {
	checked, err := checkAigTraversal(g, o)
	if err != nil {
		return err
	}
	if len(chains) == 0 {
		g.equivs = nil
		return nil
	}
	visit := aigTraversalVisit(checked)
	next := make([]AigID, len(g.objects))
	for i := range next {
		next[i] = InvalidAigID
	}
	used := make([]bool, len(g.objects))
	for _, chain := range chains {
		if len(chain) < 2 {
			return ErrDimension
		}
		for i, id := range chain {
			if err = visit(); err != nil {
				return err
			}
			if uint64(id) >= uint64(len(g.objects)) || g.objects[id].Type != AigAND {
				return ErrInvalidLiteral
			}
			if used[id] {
				return ErrDimension
			}
			used[id] = true
			if i+1 < len(chain) {
				next[id] = chain[i+1]
			}
		}
	}
	work := g.Clone()
	work.equivs = next
	if err = work.Check(); err != nil {
		return err
	}
	*g = *work
	return nil
}

// ChoiceChains returns independent representative-first chains in ascending
// representative object order.
func (g *AigManager) ChoiceChains() [][]AigID {
	if g == nil || len(g.equivs) == 0 {
		return nil
	}
	incoming := make([]bool, len(g.equivs))
	for _, next := range g.equivs {
		if next != InvalidAigID {
			incoming[next] = true
		}
	}
	var result [][]AigID
	for head, next := range g.equivs {
		if next == InvalidAigID || incoming[head] {
			continue
		}
		chain := []AigID{AigID(head)}
		for id := next; id != InvalidAigID; id = g.equivs[id] {
			chain = append(chain, id)
		}
		result = append(result, chain)
	}
	return result
}

func (g *AigManager) choiceNext(id AigID) AigID {
	if len(g.equivs) == 0 {
		return InvalidAigID
	}
	return g.equivs[id]
}

func validateAigTimingBoxes(g *AigManager, timing *TimManager) error {
	if g == nil || timing == nil || len(g.cis) != len(timing.cis) || len(g.cos) != len(timing.cos) {
		return ErrDimension
	}
	for boxID, box := range timing.boxes {
		if box.ID != boxID || box.FirstInput < 0 || box.Inputs < 1 || box.FirstInput > len(g.cos) || box.Inputs > len(g.cos)-box.FirstInput || box.FirstOutput < 0 || box.Outputs < 1 || box.FirstOutput > len(g.cis) || box.Outputs > len(g.cis)-box.FirstOutput {
			return ErrInvariant
		}
		for position := 0; position < box.Inputs; position++ {
			terminal := timing.cos[box.FirstInput+position]
			if terminal.Box != boxID || terminal.Position != position {
				return ErrInvariant
			}
		}
		for position := 0; position < box.Outputs; position++ {
			terminal := timing.cis[box.FirstOutput+position]
			if terminal.Box != boxID || terminal.Position != position {
				return ErrInvariant
			}
		}
	}
	for index, terminal := range timing.cis {
		if terminal.Box < 0 {
			continue
		}
		if terminal.Box >= len(timing.boxes) {
			return ErrInvariant
		}
		box := timing.boxes[terminal.Box]
		if terminal.Position < 0 || terminal.Position >= box.Outputs || box.FirstOutput+terminal.Position != index {
			return ErrInvariant
		}
	}
	for index, terminal := range timing.cos {
		if terminal.Box < 0 {
			continue
		}
		if terminal.Box >= len(timing.boxes) {
			return ErrInvariant
		}
		box := timing.boxes[terminal.Box]
		if terminal.Position < 0 || terminal.Position >= box.Inputs || box.FirstInput+terminal.Position != index {
			return ErrInvariant
		}
	}
	return nil
}

// VerifyTopologicalOrderWithBoxes extends Aig_ManVerifyTopoOrder to timing-box
// dependencies. Each box-input CO must precede every output CI of that box in
// object order. Timing traversal state is not modified.
func (g *AigManager) VerifyTopologicalOrderWithBoxes(timing *TimManager, o AigTraversalOptions) error {
	checked, err := checkAigTraversal(g, o)
	if err != nil {
		return err
	}
	if err = validateAigTimingBoxes(g, timing); err != nil {
		return err
	}
	ciIndex := make([]int, len(g.objects))
	for i := range ciIndex {
		ciIndex[i] = -1
	}
	for i, id := range g.cis {
		ciIndex[id] = i
	}
	seen := make([]bool, len(g.objects))
	visit := aigTraversalVisit(checked)
	for id, object := range g.objects {
		if err = visit(); err != nil {
			return err
		}
		if object.Type == AigCI {
			boxID := timing.BoxForCI(ciIndex[id])
			if boxID >= 0 {
				box := timing.boxes[boxID]
				for i := 0; i < box.Inputs; i++ {
					if err = visit(); err != nil {
						return err
					}
					if !seen[g.cos[box.FirstInput+i]] {
						return ErrInvariant
					}
				}
			}
		}
		seen[id] = true
	}
	return nil
}

// DFSChoices implements Aig_ManDfsChoices: fanin-zero, fanin-one and then the
// equivalence successor are visited before each AND. Constants/CIs are omitted.
func (g *AigManager) DFSChoices(o AigTraversalOptions) ([]AigID, error) {
	checked, err := checkAigTraversal(g, o)
	if err != nil {
		return nil, err
	}
	if len(g.equivs) == 0 {
		return nil, ErrDimension
	}
	seen := make([]bool, len(g.objects))
	seen[0] = true
	for _, id := range g.cis {
		seen[id] = true
	}
	type event struct {
		id   AigID
		exit bool
	}
	visit := aigTraversalVisit(checked)
	var result []AigID
	for _, co := range g.cos {
		stack := []event{{id: g.objects[co].Fanin0.ID()}}
		for len(stack) > 0 {
			if err = visit(); err != nil {
				return nil, err
			}
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if current.exit {
				result = append(result, current.id)
				continue
			}
			if seen[current.id] {
				continue
			}
			object := g.objects[current.id]
			if object.Type != AigAND {
				return nil, ErrDimension
			}
			seen[current.id] = true
			stack = append(stack, event{id: current.id, exit: true})
			if next := g.choiceNext(current.id); next != InvalidAigID {
				stack = append(stack, event{id: next})
			}
			stack = append(stack, event{id: object.Fanin1.ID()}, event{id: object.Fanin0.ID()})
		}
	}
	return result, nil
}

// ComputeChoiceLevels implements the no-timing-box behavior of
// Aig_ManChoiceLevel without changing structural levels in the manager.
func (g *AigManager) ComputeChoiceLevels(o AigTraversalOptions) (*AigChoiceLevels, error) {
	return g.computeChoiceLevels(nil, o)
}

// ComputeChoiceLevelsWithBoxes adds the white-box dependencies of
// Aig_ManChoiceLevel: an output CI is one level beyond the maximum level of
// every input CO belonging to its box.
func (g *AigManager) ComputeChoiceLevelsWithBoxes(timing *TimManager, o AigTraversalOptions) (*AigChoiceLevels, error) {
	return g.computeChoiceLevels(timing, o)
}

func (g *AigManager) computeChoiceLevels(timing *TimManager, o AigTraversalOptions) (*AigChoiceLevels, error) {
	checked, err := checkAigTraversal(g, o)
	if err != nil {
		return nil, err
	}
	if timing != nil {
		if err = validateAigTimingBoxes(g, timing); err != nil {
			return nil, err
		}
	}
	ciIndex := make([]int, len(g.objects))
	for i := range ciIndex {
		ciIndex[i] = -1
	}
	for i, id := range g.cis {
		ciIndex[id] = i
	}
	type event struct {
		id   AigID
		exit bool
	}
	state := make([]uint8, len(g.objects))
	levels := make([]uint32, len(g.objects))
	visit := aigTraversalVisit(checked)
	process := func(root AigID) error {
		stack := []event{{id: root}}
		for len(stack) > 0 {
			if err := visit(); err != nil {
				return err
			}
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if current.exit {
				object := g.objects[current.id]
				switch object.Type {
				case AigConstant1:
				case AigCI:
					if timing != nil {
						if boxID := timing.BoxForCI(ciIndex[current.id]); boxID >= 0 {
							box := timing.boxes[boxID]
							level := uint32(0)
							for i := 0; i < box.Inputs; i++ {
								level = max(level, levels[g.cos[box.FirstInput+i]])
							}
							if level == ^uint32(0) {
								return ErrLimit
							}
							levels[current.id] = level + 1
						}
					}
				case AigCO:
					levels[current.id] = levels[object.Fanin0.ID()]
				case AigAND:
					level := max(levels[object.Fanin0.ID()], levels[object.Fanin1.ID()])
					if level == ^uint32(0) {
						return ErrLimit
					}
					level++
					if next := g.choiceNext(current.id); next != InvalidAigID {
						level = max(level, levels[next])
					}
					levels[current.id] = level
				default:
					return ErrDimension
				}
				state[current.id] = 2
				continue
			}
			if state[current.id] == 2 {
				continue
			}
			if state[current.id] == 1 {
				return ErrInvariant
			}
			state[current.id] = 1
			object := g.objects[current.id]
			stack = append(stack, event{id: current.id, exit: true})
			switch object.Type {
			case AigConstant1:
			case AigCI:
				if timing != nil {
					if boxID := timing.BoxForCI(ciIndex[current.id]); boxID >= 0 {
						box := timing.boxes[boxID]
						for i := box.Inputs - 1; i >= 0; i-- {
							stack = append(stack, event{id: g.cos[box.FirstInput+i]})
						}
					}
				}
			case AigCO:
				stack = append(stack, event{id: object.Fanin0.ID()})
			case AigAND:
				if next := g.choiceNext(current.id); next != InvalidAigID {
					stack = append(stack, event{id: next})
				}
				stack = append(stack, event{id: object.Fanin1.ID()}, event{id: object.Fanin0.ID()})
			default:
				return ErrDimension
			}
		}
		return nil
	}
	maximum := uint32(0)
	for _, id := range g.cos {
		if err = process(id); err != nil {
			return nil, err
		}
		maximum = max(maximum, levels[id])
	}
	for _, id := range g.cis {
		if err = process(id); err != nil {
			return nil, err
		}
		maximum = max(maximum, levels[id])
	}
	return &AigChoiceLevels{Levels: slices.Clone(levels), Maximum: maximum}, nil
}
