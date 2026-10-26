package abc

// Port of src/aig/aig/aigMffc.c using owned reference-count transactions.

import "slices"

type AigMFFCOptions struct {
	MaxVisits int // zero: 1M
	Stop      func() bool
}

type AigMFFC struct {
	Size    int
	Support []AigID
	Members []bool
	Power   float32
}

type aigMFFCBudget struct {
	options AigMFFCOptions
	visits  int
}

func newAigMFFCBudget(g *AigManager, root AigID, options AigMFFCOptions) (*aigMFFCBudget, error) {
	if g == nil || options.MaxVisits < 0 {
		return nil, ErrDimension
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	if uint64(root) >= uint64(len(g.objects)) || g.objects[root].Type == AigNone {
		return nil, ErrInvalidLiteral
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	return &aigMFFCBudget{options: options}, nil
}

func (b *aigMFFCBudget) step() error {
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	if b.visits == b.options.MaxVisits {
		return ErrLimit
	}
	b.visits++
	return nil
}

type aigMFFCFrame struct {
	id    AigID
	stage uint8
	count int
	power float32
}

// aigMFFCAddPower mirrors C's compound assignment to float. The unsuffixed
// constants in aigMffc.c promote every probability expression to double;
// assignment rounds the result back to float after each fanin.
func aigMFFCAddPower(current, nested, probability float32) float32 {
	p := float64(probability)
	return float32(float64(current) + float64(nested) + 2*p*(1-p))
}

func aigMFFCDeref(g *AigManager, root AigID, levelMinimum uint32, references []uint32, probabilities []float32, budget *aigMFFCBudget) (int, float32, error) {
	stack := []aigMFFCFrame{{id: root}}
	lastCount := 0
	lastPower := float32(0)
	for len(stack) != 0 {
		frame := &stack[len(stack)-1]
		object := g.objects[frame.id]
		switch frame.stage {
		case 0:
			if err := budget.step(); err != nil {
				return 0, 0, err
			}
			if object.Type == AigCI {
				lastCount, lastPower = 0, 0
				stack = stack[:len(stack)-1]
				continue
			}
			if object.Type != AigAND && object.Type != AigBuffer {
				return 0, 0, ErrUnsupported
			}
			fanin := object.Fanin0.ID()
			if references[fanin] == 0 {
				return 0, 0, ErrInvariant
			}
			references[fanin]--
			frame.stage = 1
			if references[fanin] == 0 && (levelMinimum == 0 || g.objects[fanin].Level > levelMinimum) {
				stack = append(stack, aigMFFCFrame{id: fanin})
				continue
			}
			lastCount, lastPower = 0, 0
		case 1:
			fanin := object.Fanin0.ID()
			frame.count = lastCount
			if probabilities != nil {
				frame.power = aigMFFCAddPower(0, lastPower, probabilities[fanin])
			}
			if object.Type == AigBuffer {
				lastCount, lastPower = frame.count, frame.power
				stack = stack[:len(stack)-1]
				continue
			}
			fanin = object.Fanin1.ID()
			if references[fanin] == 0 {
				return 0, 0, ErrInvariant
			}
			references[fanin]--
			frame.stage = 2
			if references[fanin] == 0 && (levelMinimum == 0 || g.objects[fanin].Level > levelMinimum) {
				stack = append(stack, aigMFFCFrame{id: fanin})
				continue
			}
			lastCount, lastPower = 0, 0
		case 2:
			fanin := object.Fanin1.ID()
			lastCount = frame.count + lastCount + 1
			if probabilities != nil {
				lastPower = aigMFFCAddPower(frame.power, lastPower, probabilities[fanin])
			} else {
				lastPower = 0
			}
			stack = stack[:len(stack)-1]
		}
	}
	return lastCount, lastPower, nil
}

func aigMFFCRef(g *AigManager, root AigID, levelMinimum uint32, references []uint32, members []bool, budget *aigMFFCBudget) (int, error) {
	stack := []aigMFFCFrame{{id: root}}
	lastCount := 0
	for len(stack) != 0 {
		frame := &stack[len(stack)-1]
		object := g.objects[frame.id]
		switch frame.stage {
		case 0:
			if err := budget.step(); err != nil {
				return 0, err
			}
			if object.Type == AigCI {
				lastCount = 0
				stack = stack[:len(stack)-1]
				continue
			}
			if object.Type != AigAND && object.Type != AigBuffer {
				return 0, ErrUnsupported
			}
			if members != nil {
				members[frame.id] = true
			}
			fanin := object.Fanin0.ID()
			previous := references[fanin]
			if previous == ^uint32(0) {
				return 0, ErrLimit
			}
			references[fanin]++
			frame.stage = 1
			if previous == 0 && (levelMinimum == 0 || g.objects[fanin].Level > levelMinimum) {
				stack = append(stack, aigMFFCFrame{id: fanin})
				continue
			}
			lastCount = 0
		case 1:
			frame.count = lastCount
			if object.Type == AigBuffer {
				lastCount = frame.count
				stack = stack[:len(stack)-1]
				continue
			}
			fanin := object.Fanin1.ID()
			previous := references[fanin]
			if previous == ^uint32(0) {
				return 0, ErrLimit
			}
			references[fanin]++
			frame.stage = 2
			if previous == 0 && (levelMinimum == 0 || g.objects[fanin].Level > levelMinimum) {
				stack = append(stack, aigMFFCFrame{id: fanin})
				continue
			}
			lastCount = 0
		case 2:
			lastCount = frame.count + lastCount + 1
			stack = stack[:len(stack)-1]
		}
	}
	return lastCount, nil
}

func aigMFFCSupport(g *AigManager, root AigID, levelMinimum uint32, references []uint32, skip AigID, budget *aigMFFCBudget) ([]AigID, error) {
	type item struct {
		id  AigID
		top bool
	}
	seen := make([]bool, len(g.objects))
	stack := []item{{id: root, top: true}}
	var support []AigID
	for len(stack) != 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[current.id] {
			continue
		}
		if err := budget.step(); err != nil {
			return nil, err
		}
		seen[current.id] = true
		object := g.objects[current.id]
		if !current.top && current.id != skip && (object.Type == AigCI || references[current.id] > 0 || object.Level <= levelMinimum) {
			support = append(support, current.id)
			continue
		}
		if object.Type != AigAND {
			return nil, ErrUnsupported
		}
		stack = append(stack, item{id: object.Fanin1.ID()}, item{id: object.Fanin0.ID()})
	}
	return support, nil
}

func aigMFFCReferences(g *AigManager) []uint32 {
	references := make([]uint32, len(g.objects))
	for id, object := range g.objects {
		references[id] = object.RefCount
	}
	return references
}

// MFFCSupport returns the depth-limited fanout-free cone size and its ordered
// boundary. A CI root has size zero and is its own boundary.
func (g *AigManager) MFFCSupport(root AigID, levelMinimum int, options AigMFFCOptions) (AigMFFC, error) {
	var result AigMFFC
	if levelMinimum < 0 {
		return result, ErrDimension
	}
	budget, err := newAigMFFCBudget(g, root, options)
	if err != nil {
		return result, err
	}
	if g.objects[root].Type == AigCI {
		result.Support = []AigID{root}
		return result, nil
	}
	if g.objects[root].Type != AigAND {
		return result, ErrUnsupported
	}
	references := aigMFFCReferences(g)
	result.Size, _, err = aigMFFCDeref(g, root, uint32(levelMinimum), references, nil, budget)
	if err != nil {
		return AigMFFC{}, err
	}
	result.Support, err = aigMFFCSupport(g, root, uint32(levelMinimum), references, InvalidAigID, budget)
	if err != nil {
		return AigMFFC{}, err
	}
	restored, err := aigMFFCRef(g, root, uint32(levelMinimum), references, nil, budget)
	if err != nil {
		return AigMFFC{}, err
	}
	if restored != result.Size || !slices.Equal(references, aigMFFCReferences(g)) || result.Size == 0 {
		return AigMFFC{}, ErrInvariant
	}
	return result, nil
}

// MFFCLabel returns the complete native dereference/reference membership and,
// when probabilities are supplied, the switching-power sum.
func (g *AigManager) MFFCLabel(root AigID, probabilities []float32, options AigMFFCOptions) (AigMFFC, error) {
	var result AigMFFC
	budget, err := newAigMFFCBudget(g, root, options)
	if err != nil {
		return result, err
	}
	if g.objects[root].Type != AigAND || (probabilities != nil && len(probabilities) != len(g.objects)) {
		return result, ErrDimension
	}
	references := aigMFFCReferences(g)
	result.Size, result.Power, err = aigMFFCDeref(g, root, 0, references, probabilities, budget)
	if err != nil {
		return AigMFFC{}, err
	}
	result.Members = make([]bool, len(g.objects))
	restored, err := aigMFFCRef(g, root, 0, references, result.Members, budget)
	if err != nil {
		return AigMFFC{}, err
	}
	if restored != result.Size || !slices.Equal(references, aigMFFCReferences(g)) || result.Size == 0 {
		return AigMFFC{}, ErrInvariant
	}
	return result, nil
}

// MFFCLabelCut temporarily protects every leaf reference, preserving duplicate
// leaves, and returns the labeled cone.
func (g *AigManager) MFFCLabelCut(root AigID, leaves []AigID, options AigMFFCOptions) (AigMFFC, error) {
	var result AigMFFC
	budget, err := newAigMFFCBudget(g, root, options)
	if err != nil {
		return result, err
	}
	if g.objects[root].Type != AigAND {
		return result, ErrDimension
	}
	references := aigMFFCReferences(g)
	for _, leaf := range leaves {
		if uint64(leaf) >= uint64(len(g.objects)) || g.objects[leaf].Type == AigNone || references[leaf] == ^uint32(0) {
			return AigMFFC{}, ErrInvalidLiteral
		}
		references[leaf]++
	}
	result.Size, _, err = aigMFFCDeref(g, root, 0, references, nil, budget)
	if err != nil {
		return AigMFFC{}, err
	}
	result.Members = make([]bool, len(g.objects))
	restored, err := aigMFFCRef(g, root, 0, references, result.Members, budget)
	if err != nil {
		return AigMFFC{}, err
	}
	for _, leaf := range leaves {
		if references[leaf] == 0 {
			return AigMFFC{}, ErrInvariant
		}
		references[leaf]--
	}
	if restored != result.Size || !slices.Equal(references, aigMFFCReferences(g)) || result.Size == 0 {
		return AigMFFC{}, ErrInvariant
	}
	return result, nil
}

// ExtendMFFCCut expands the first cheapest maximum-level leaf and returns the
// new ordered boundary. Exists is false when all leaves have level zero.
func (g *AigManager) ExtendMFFCCut(root AigID, leaves []AigID, options AigMFFCOptions) (result []AigID, exists bool, err error) {
	budget, err := newAigMFFCBudget(g, root, options)
	if err != nil {
		return nil, false, err
	}
	if g.objects[root].Type != AigAND {
		return nil, false, ErrDimension
	}
	levelMaximum := uint32(0)
	for _, leaf := range leaves {
		if uint64(leaf) >= uint64(len(g.objects)) || g.objects[leaf].Type == AigNone {
			return nil, false, ErrInvalidLiteral
		}
		levelMaximum = max(levelMaximum, g.objects[leaf].Level)
	}
	if levelMaximum == 0 {
		return nil, false, nil
	}
	references := aigMFFCReferences(g)
	originalSize, _, err := aigMFFCDeref(g, root, 0, references, nil, budget)
	if err != nil {
		return nil, false, err
	}
	best, bestSize := InvalidAigID, int(^uint(0)>>1)
	for _, leaf := range leaves {
		if g.objects[leaf].Level != levelMaximum {
			continue
		}
		current, _, err := aigMFFCDeref(g, leaf, 0, references, nil, budget)
		if err != nil {
			return nil, false, err
		}
		if current < bestSize {
			best, bestSize = leaf, current
		}
		restored, err := aigMFFCRef(g, leaf, 0, references, nil, budget)
		if err != nil || restored != current {
			return nil, false, ErrInvariant
		}
	}
	if best == InvalidAigID || g.objects[best].Type != AigAND {
		return nil, false, ErrInvariant
	}
	removed, _, err := aigMFFCDeref(g, best, 0, references, nil, budget)
	if err != nil {
		return nil, false, err
	}
	result, err = aigMFFCSupport(g, root, 0, references, best, budget)
	if err != nil {
		return nil, false, err
	}
	restored, err := aigMFFCRef(g, best, 0, references, nil, budget)
	if err != nil || restored != removed {
		return nil, false, ErrInvariant
	}
	restored, err = aigMFFCRef(g, root, 0, references, nil, budget)
	if err != nil || restored != originalSize || !slices.Equal(references, aigMFFCReferences(g)) {
		return nil, false, ErrInvariant
	}
	return result, true, nil
}
