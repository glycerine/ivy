package abc

// MFS fanout-window root collection from src/opt/mfs/mfsWin.c.

type ABCMFSWindowOptions struct {
	MaxVisited int // distinct reached nodes; zero: one million
	MaxRoots   int // retained roots; zero: one million
	Stop       func() bool
}

// ABCMFSWindowRoots returns the ordered fanout-window roots reached from start.
// levels is the precomputed object-ID-indexed level array consumed by the C
// routine. A reached node is a root when its fanout multiplicity exceeds
// fanoutLimit, when it drives a combinational output, or when any fanout lies
// above levels[start]+maxTFOLevels. Otherwise its node fanouts are explored in
// source order. The network and levels are not mutated.
func ABCMFSWindowRoots(network *ABCNetwork, start int, levels []int, maxTFOLevels, fanoutLimit int, options ABCMFSWindowOptions) ([]int, error) {
	if network == nil || maxTFOLevels < 0 || fanoutLimit < 0 || options.MaxVisited < 0 || options.MaxRoots < 0 {
		return nil, ErrDimension
	}
	if err := network.Validate(); err != nil {
		return nil, err
	}
	if len(levels) != len(network.objects) || start < 0 || start >= len(network.objects) ||
		!network.objects[start].Alive || network.objects[start].Type != ABCObjectNode {
		return nil, ErrDimension
	}
	for id, object := range network.objects {
		if object.Alive && levels[id] < 0 {
			return nil, ErrDimension
		}
	}
	maxVisited, maxRoots := options.MaxVisited, options.MaxRoots
	if maxVisited == 0 {
		maxVisited = 1 << 20
	}
	if maxRoots == 0 {
		maxRoots = 1 << 20
	}
	if levels[start] > int(^uint(0)>>1)-maxTFOLevels {
		return nil, ErrLimit
	}
	levelMaximum := levels[start] + maxTFOLevels
	visited := make([]bool, len(network.objects))
	stack := []int{start}
	roots := make([]int, 0)
	seen := 0
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	for len(stack) > 0 {
		if stopped() {
			return nil, ErrCancelled
		}
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[id] {
			continue
		}
		visited[id] = true
		seen++
		if seen > maxVisited {
			return nil, ErrLimit
		}
		object := &network.objects[id]
		if object.Type != ABCObjectNode {
			return nil, ErrInvariant
		}
		isRoot := len(object.Fanouts) > fanoutLimit
		for _, fanout := range object.Fanouts {
			if stopped() {
				return nil, ErrCancelled
			}
			other := &network.objects[fanout]
			if abcIsCO(other.Type) || levels[fanout] > levelMaximum {
				isRoot = true
			}
		}
		if isRoot {
			if len(roots) >= maxRoots {
				return nil, ErrLimit
			}
			roots = append(roots, id)
			continue
		}
		for index := len(object.Fanouts) - 1; index >= 0; index-- {
			fanout := object.Fanouts[index]
			if network.objects[fanout].Type != ABCObjectNode {
				return nil, ErrInvariant
			}
			stack = append(stack, fanout)
		}
	}
	if len(roots) == 0 {
		return nil, ErrInvariant
	}
	return roots, nil
}
