package abc

// Port of src/aig/aig/aigWin.c using owned frontier marks.

import "slices"

type AigWindowOptions struct {
	MaxWork int // zero: 1M frontier examinations
	Stop    func() bool
}

type AigWindowCut struct {
	Leaves  []AigID
	Visited []AigID
}

type aigWindowBudget struct {
	options AigWindowOptions
	work    int
}

func newAigWindowBudget(options AigWindowOptions) (*aigWindowBudget, error) {
	if options.MaxWork < 0 {
		return nil, ErrDimension
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 20
	}
	return &aigWindowBudget{options: options}, nil
}

func (b *aigWindowBudget) step() error {
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	if b.work == b.options.MaxWork {
		return ErrLimit
	}
	b.work++
	return nil
}

// FindReconvergentCut implements Aig_ManFindCut. It repeatedly expands the
// cheapest frontier node, preferring the deeper node on equal cost. Returned
// slices are owned and the manager is unchanged.
func (g *AigManager) FindReconvergentCut(root AigID, sizeLimit, fanoutLimit int, options AigWindowOptions) (AigWindowCut, error) {
	var result AigWindowCut
	if g == nil || sizeLimit < 2 || fanoutLimit < 0 {
		return result, ErrDimension
	}
	if err := g.Check(); err != nil {
		return result, err
	}
	if uint64(root) >= uint64(len(g.objects)) || g.objects[root].Type == AigNone {
		return result, ErrInvalidLiteral
	}
	rootObject := g.objects[root]
	if rootObject.Type != AigAND {
		return result, ErrUnsupported
	}
	budget, err := newAigWindowBudget(options)
	if err != nil {
		return result, err
	}
	marked := make([]bool, len(g.objects))
	first, second := rootObject.Fanin0.ID(), rootObject.Fanin1.ID()
	result.Leaves = []AigID{first, second}
	result.Visited = []AigID{root, first, second}
	marked[root], marked[first], marked[second] = true, true, true

	for {
		bestIndex, bestCost := -1, 100
		for index, id := range result.Leaves {
			if err := budget.step(); err != nil {
				return AigWindowCut{}, err
			}
			object := g.objects[id]
			cost := 999
			switch object.Type {
			case AigCI:
			case AigAND:
				cost = 0
				if !marked[object.Fanin0.ID()] {
					cost++
				}
				if !marked[object.Fanin1.ID()] {
					cost++
				}
				if cost == 2 && int(object.RefCount) > fanoutLimit {
					cost = 999
				}
			default:
				return AigWindowCut{}, ErrUnsupported
			}
			if cost < bestCost || (cost == bestCost && bestIndex >= 0 && object.Level > g.objects[result.Leaves[bestIndex]].Level) {
				bestIndex, bestCost = index, cost
			}
			if bestCost == 0 {
				break
			}
		}
		if bestIndex < 0 || len(result.Leaves)-1+bestCost > sizeLimit {
			break
		}
		best := g.objects[result.Leaves[bestIndex]]
		result.Leaves = slices.Delete(result.Leaves, bestIndex, bestIndex+1)
		for _, fanin := range []AigID{best.Fanin0.ID(), best.Fanin1.ID()} {
			if marked[fanin] {
				continue
			}
			marked[fanin] = true
			result.Leaves = append(result.Leaves, fanin)
			result.Visited = append(result.Visited, fanin)
		}
	}
	return result, nil
}
