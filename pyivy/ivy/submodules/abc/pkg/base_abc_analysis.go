package abc

// Derived from ABC src/base/abc/{abcCheck.c,abcDfs.c,abcRefs.c}.

import (
	"fmt"
	"slices"
	"sort"
)

// ABCAnalysisOptions bounds standalone network analysis.
type ABCAnalysisOptions struct {
	MaxVisited int
	Stop       func() bool
}

func (o ABCAnalysisOptions) normalized(n *ABCNetwork) ABCAnalysisOptions {
	if o.MaxVisited <= 0 {
		o.MaxVisited = 1 << 22
		if n != nil && n.limits.MaxObjects > 0 && n.limits.MaxObjects < o.MaxVisited {
			o.MaxVisited = n.limits.MaxObjects
		}
	}
	return o
}

type abcAnalysisBudget struct {
	options ABCAnalysisOptions
	seen    int
}

func (b *abcAnalysisBudget) tick() error {
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	b.seen++
	if b.seen > b.options.MaxVisited {
		return ErrLimit
	}
	return nil
}

func abcAnalysisInternal(t ABCObjectType) bool {
	return t == ABCObjectNode || abcIsBox(t)
}

func abcAnalysisBoundary(t ABCObjectType) bool {
	return abcIsCI(t) || t == ABCObjectConst1
}

func abcAnalysisCycleBoundary(t ABCObjectType) bool {
	return abcAnalysisBoundary(t) || abcIsBox(t)
}

// ABCReferenceCounts returns source-shaped fanout reference multiplicities.
func ABCReferenceCounts(network *ABCNetwork) ([]int, error) {
	if err := network.Validate(); err != nil {
		return nil, err
	}
	counts := make([]int, len(network.objects))
	for id, object := range network.objects {
		if object.Alive {
			counts[id] = len(object.Fanouts)
		}
	}
	return counts, nil
}

// CheckABCNetwork performs the independently useful structural portion of
// Abc_NtkCheck and strengthens it with explicit cycle detection.
func CheckABCNetwork(network *ABCNetwork, options ABCAnalysisOptions) error {
	if err := network.Validate(); err != nil {
		return err
	}
	budget := abcAnalysisBudget{options: options.normalized(network)}
	for id, object := range network.objects {
		if !object.Alive {
			continue
		}
		if err := budget.tick(); err != nil {
			return err
		}
		switch object.Type {
		case ABCObjectPI, ABCObjectConst1:
			if len(object.Fanins) != 0 {
				return fmt.Errorf("object %d is a source with fanins: %w", id, ErrInvariant)
			}
		case ABCObjectPO, ABCObjectBI, ABCObjectNet:
			if len(object.Fanins) != 1 {
				return fmt.Errorf("object %d requires one driver: %w", id, ErrInvariant)
			}
		case ABCObjectBO:
			if len(object.Fanins) > 1 {
				return fmt.Errorf("BO %d has multiple drivers: %w", id, ErrInvariant)
			}
		case ABCObjectLatch:
			if len(object.Fanins) != 1 || len(object.Fanouts) != 1 || object.Init < ABCInitZero || object.Init > ABCInitDC {
				return fmt.Errorf("latch %d shape/init: %w", id, ErrInvariant)
			}
			input := &network.objects[object.Fanins[0].ID]
			output := &network.objects[object.Fanouts[0]]
			if input.Type != ABCObjectBI || len(input.Fanins) != 1 || len(input.Fanouts) != 1 ||
				output.Type != ABCObjectBO || len(output.Fanins) != 1 || output.Fanins[0].ID != id {
				return fmt.Errorf("latch %d terminal shape: %w", id, ErrInvariant)
			}
		}
	}
	acyclic, _, err := ABCNetworkAcyclic(network, options)
	if err != nil {
		return err
	}
	if !acyclic {
		return fmt.Errorf("combinational cycle: %w", ErrInvariant)
	}
	return nil
}

type abcDFSFrame struct {
	id   int
	next int
}

func abcForwardRoots(network *ABCNetwork, roots []int) ([]int, error) {
	result := make([]int, 0, len(roots))
	for _, id := range roots {
		object, err := network.live(id)
		if err != nil {
			return nil, err
		}
		if abcIsCO(object.Type) {
			for _, edge := range object.Fanins {
				result = append(result, edge.ID)
			}
			continue
		}
		result = append(result, id)
	}
	return result, nil
}

// ABCForwardDFS returns source-order fanin postorder, excluding terminals and
// transparent nets. With collectAll it appends dangling logic in object order.
func ABCForwardDFS(network *ABCNetwork, roots []int, collectAll bool, options ABCAnalysisOptions) ([]int, error) {
	if err := network.Validate(); err != nil {
		return nil, err
	}
	if roots == nil {
		roots = network.COs()
	}
	starts, err := abcForwardRoots(network, roots)
	if err != nil {
		return nil, err
	}
	if collectAll {
		for id, object := range network.objects {
			if object.Alive && abcAnalysisInternal(object.Type) {
				starts = append(starts, id)
			}
		}
	}
	state := make([]uint8, len(network.objects))
	result := make([]int, 0)
	budget := abcAnalysisBudget{options: options.normalized(network)}
	for _, start := range starts {
		if state[start] == 2 || abcAnalysisBoundary(network.objects[start].Type) {
			continue
		}
		stack := []abcDFSFrame{{id: start}}
		for len(stack) != 0 {
			frame := &stack[len(stack)-1]
			object := &network.objects[frame.id]
			if state[frame.id] == 0 {
				if err := budget.tick(); err != nil {
					return nil, err
				}
				state[frame.id] = 1
			}
			if frame.next < len(object.Fanins) {
				child := object.Fanins[frame.next].ID
				frame.next++
				if abcAnalysisBoundary(network.objects[child].Type) || state[child] == 2 {
					continue
				}
				if state[child] == 1 {
					return nil, fmt.Errorf("cycle through object %d: %w", child, ErrInvariant)
				}
				stack = append(stack, abcDFSFrame{id: child})
				continue
			}
			state[frame.id] = 2
			if abcAnalysisInternal(object.Type) {
				result = append(result, frame.id)
			}
			stack = stack[:len(stack)-1]
		}
	}
	return result, nil
}

// ABCReverseDFS returns fanout-first postorder from the given roots (all CIs
// when nil), excluding CIs and COs.
func ABCReverseDFS(network *ABCNetwork, roots []int, options ABCAnalysisOptions) ([]int, error) {
	if err := network.Validate(); err != nil {
		return nil, err
	}
	defaultRoots := roots == nil
	if defaultRoots {
		roots = network.CIs()
	}
	state := make([]uint8, len(network.objects))
	result := make([]int, 0)
	budget := abcAnalysisBudget{options: options.normalized(network)}
	for _, root := range roots {
		if _, err := network.live(root); err != nil {
			return nil, err
		}
		stack := []abcDFSFrame{{id: root}}
		for len(stack) != 0 {
			frame := &stack[len(stack)-1]
			object := &network.objects[frame.id]
			if state[frame.id] == 0 {
				if err := budget.tick(); err != nil {
					return nil, err
				}
				state[frame.id] = 1
			}
			if frame.next < len(object.Fanouts) {
				child := object.Fanouts[frame.next]
				frame.next++
				if abcIsCO(network.objects[child].Type) || state[child] == 2 {
					continue
				}
				if state[child] == 1 {
					return nil, fmt.Errorf("cycle through object %d: %w", child, ErrInvariant)
				}
				stack = append(stack, abcDFSFrame{id: child})
				continue
			}
			state[frame.id] = 2
			if abcAnalysisInternal(object.Type) {
				result = append(result, frame.id)
			}
			stack = stack[:len(stack)-1]
		}
	}
	// The source reverse traversal appends SOP constant nodes after the
	// CI-rooted traversal. Transparent nets are never returned.
	if defaultRoots && network.Type != ABCNetworkStrash {
		for id, object := range network.objects {
			if object.Alive && object.Type == ABCObjectNode && len(object.Fanins) == 0 &&
				(object.Function == " 0\n" || object.Function == " 1\n") {
				result = append(result, id)
			}
		}
	}
	return result, nil
}

func abcAnalysisNetDriver(network *ABCNetwork, id int) (int, error) {
	for steps := 0; network.objects[id].Type == ABCObjectNet; steps++ {
		if steps >= len(network.objects) || len(network.objects[id].Fanins) != 1 {
			return -1, fmt.Errorf("transparent net %d: %w", id, ErrInvariant)
		}
		id = network.objects[id].Fanins[0].ID
	}
	return id, nil
}

// ABCNetworkSupport returns CIs reached from roots in first-encounter order.
func ABCNetworkSupport(network *ABCNetwork, roots []int, options ABCAnalysisOptions) ([]int, error) {
	if err := network.Validate(); err != nil {
		return nil, err
	}
	if roots == nil {
		roots = network.COs()
	}
	starts, err := abcForwardRoots(network, roots)
	if err != nil {
		return nil, err
	}
	visited := make([]bool, len(network.objects))
	result := make([]int, 0)
	budget := abcAnalysisBudget{options: options.normalized(network)}
	stack := slices.Clone(starts)
	slices.Reverse(stack)
	for len(stack) != 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[id] {
			continue
		}
		visited[id] = true
		if err := budget.tick(); err != nil {
			return nil, err
		}
		object := &network.objects[id]
		if abcIsCI(object.Type) {
			result = append(result, id)
			continue
		}
		for index := len(object.Fanins) - 1; index >= 0; index-- {
			stack = append(stack, object.Fanins[index].ID)
		}
	}
	return result, nil
}

// ABCNetworkLevels returns forward levels by stable object ID and the maximum
// internal level. CIs and the distinguished constant are level zero.
func ABCNetworkLevels(network *ABCNetwork, options ABCAnalysisOptions) ([]int, int, error) {
	order, err := ABCForwardDFS(network, nil, true, options)
	if err != nil {
		return nil, 0, err
	}
	levels := make([]int, len(network.objects))
	maximum := 0
	for _, id := range order {
		level := 0
		for _, edge := range network.objects[id].Fanins {
			fanin, resolveErr := abcAnalysisNetDriver(network, edge.ID)
			if resolveErr != nil {
				return nil, 0, resolveErr
			}
			candidate := levels[fanin] + 1
			if candidate > level {
				level = candidate
			}
		}
		levels[id] = level
		if level > maximum {
			maximum = level
		}
	}
	for _, id := range network.cos {
		if len(network.objects[id].Fanins) == 1 {
			fanin, resolveErr := abcAnalysisNetDriver(network, network.objects[id].Fanins[0].ID)
			if resolveErr != nil {
				return nil, 0, resolveErr
			}
			levels[id] = levels[fanin]
		}
	}
	return levels, maximum, nil
}

// ABCNetworkAcyclic returns one closed cycle when the network is cyclic.
func ABCNetworkAcyclic(network *ABCNetwork, options ABCAnalysisOptions) (bool, []int, error) {
	if err := network.Validate(); err != nil {
		return false, nil, err
	}
	color := make([]uint8, len(network.objects))
	parent := make([]int, len(network.objects))
	for i := range parent {
		parent[i] = -1
	}
	budget := abcAnalysisBudget{options: options.normalized(network)}
	for start, object := range network.objects {
		if !object.Alive || color[start] != 0 || abcAnalysisCycleBoundary(object.Type) {
			continue
		}
		stack := []abcDFSFrame{{id: start}}
		for len(stack) != 0 {
			frame := &stack[len(stack)-1]
			if color[frame.id] == 0 {
				if err := budget.tick(); err != nil {
					return false, nil, err
				}
				color[frame.id] = 1
			}
			object := &network.objects[frame.id]
			if frame.next < len(object.Fanins) {
				child := object.Fanins[frame.next].ID
				frame.next++
				if abcAnalysisCycleBoundary(network.objects[child].Type) || color[child] == 2 {
					continue
				}
				if color[child] == 1 {
					cycle := []int{frame.id}
					for id := parent[frame.id]; id != child && id >= 0; id = parent[id] {
						cycle = append(cycle, id)
					}
					cycle = append(cycle, child)
					slices.Reverse(cycle)
					cycle = append(cycle, child)
					return false, cycle, nil
				}
				parent[child] = frame.id
				stack = append(stack, abcDFSFrame{id: child})
				continue
			}
			color[frame.id] = 2
			stack = stack[:len(stack)-1]
		}
	}
	return true, nil, nil
}

// ABCMFFCResult is an owned MFFC cone/support snapshot in fanin postorder.
type ABCMFFCResult struct {
	Size    int
	Cone    []int
	Support []int
}

// ABCNodeMFFC computes source-style reference/dereference membership without
// mutating the network. leaves add one temporary external reference each.
func ABCNodeMFFC(network *ABCNetwork, root int, leaves []int, stopComplemented bool, options ABCAnalysisOptions) (ABCMFFCResult, error) {
	if err := network.Validate(); err != nil {
		return ABCMFFCResult{}, err
	}
	rootObject, err := network.live(root)
	if err != nil {
		return ABCMFFCResult{}, err
	}
	if rootObject.Type != ABCObjectNode {
		return ABCMFFCResult{}, fmt.Errorf("MFFC root: %w", ErrDimension)
	}
	if len(rootObject.Fanins) == 0 {
		return ABCMFFCResult{}, nil
	}
	refs := make([]int, len(network.objects))
	for id, object := range network.objects {
		if object.Alive {
			refs[id] = len(object.Fanouts)
		}
	}
	for _, leaf := range leaves {
		if _, err := network.live(leaf); err != nil {
			return ABCMFFCResult{}, err
		}
		refs[leaf]++
	}
	members := make([]bool, len(network.objects))
	stack := []int{root}
	budget := abcAnalysisBudget{options: options.normalized(network)}
	for len(stack) != 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if members[id] || abcAnalysisBoundary(network.objects[id].Type) {
			continue
		}
		if err := budget.tick(); err != nil {
			return ABCMFFCResult{}, err
		}
		members[id] = true
		for _, edge := range network.objects[id].Fanins {
			if stopComplemented && edge.Complemented {
				continue
			}
			if refs[edge.ID] <= 0 {
				return ABCMFFCResult{}, fmt.Errorf("reference underflow at %d: %w", edge.ID, ErrInvariant)
			}
			refs[edge.ID]--
			if refs[edge.ID] == 0 {
				stack = append(stack, edge.ID)
			}
		}
	}
	cone := make([]int, 0)
	support := make([]int, 0)
	supportSeen := make([]bool, len(network.objects))
	state := make([]uint8, len(network.objects))
	frames := []abcDFSFrame{{id: root}}
	for len(frames) != 0 {
		frame := &frames[len(frames)-1]
		object := &network.objects[frame.id]
		if state[frame.id] == 0 {
			if err := budget.tick(); err != nil {
				return ABCMFFCResult{}, err
			}
			state[frame.id] = 1
		}
		if frame.next < len(object.Fanins) {
			edge := object.Fanins[frame.next]
			frame.next++
			if stopComplemented && edge.Complemented {
				if !supportSeen[edge.ID] {
					supportSeen[edge.ID] = true
					support = append(support, edge.ID)
				}
				continue
			}
			if members[edge.ID] {
				if state[edge.ID] == 0 {
					frames = append(frames, abcDFSFrame{id: edge.ID})
				}
			} else if !supportSeen[edge.ID] {
				supportSeen[edge.ID] = true
				support = append(support, edge.ID)
			}
			continue
		}
		state[frame.id] = 2
		cone = append(cone, frame.id)
		frames = frames[:len(frames)-1]
	}
	return ABCMFFCResult{Size: len(cone), Cone: cone, Support: support}, nil
}

// ABCCompareSignals compares source-shaped name-sorted terminal names. When
// combinational is set, latch output names are compared as in Abc_NtkCompareBoxes.
func ABCCompareSignals(first, second *ABCNetwork, onlyPIs, combinational bool) (bool, error) {
	if err := CheckABCNetwork(first, ABCAnalysisOptions{}); err != nil {
		return false, err
	}
	if err := CheckABCNetwork(second, ABCAnalysisOptions{}); err != nil {
		return false, err
	}
	compare := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		firstNames := make([]string, len(a))
		secondNames := make([]string, len(b))
		for i := range a {
			firstNames[i] = first.objects[a[i]].Name
			secondNames[i] = second.objects[b[i]].Name
		}
		sort.Strings(firstNames)
		sort.Strings(secondNames)
		for i := range firstNames {
			if firstNames[i] != secondNames[i] {
				return false
			}
		}
		return true
	}
	if !compare(first.pis, second.pis) {
		return false, nil
	}
	if onlyPIs {
		return true, nil
	}
	if combinational {
		if len(first.boxes) != len(second.boxes) {
			return false, nil
		}
		firstOutputs := make([]int, len(first.boxes))
		secondOutputs := make([]int, len(second.boxes))
		for i := range first.boxes {
			a, b := &first.objects[first.boxes[i]], &second.objects[second.boxes[i]]
			if a.Type != ABCObjectLatch || b.Type != ABCObjectLatch || len(a.Fanouts) != 1 || len(b.Fanouts) != 1 {
				return false, fmt.Errorf("signal comparison requires latch boxes: %w", ErrDimension)
			}
			firstOutputs[i] = a.Fanouts[0]
			secondOutputs[i] = b.Fanouts[0]
		}
		if !compare(firstOutputs, secondOutputs) {
			return false, nil
		}
	}
	if !compare(first.pos, second.pos) {
		return false, nil
	}
	return true, nil
}

// ABCSortedSignalNames is the owned replacement for name-comparator helpers.
func ABCSortedSignalNames(network *ABCNetwork, ids []int) ([]string, error) {
	if err := network.Validate(); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		object, err := network.live(id)
		if err != nil {
			return nil, err
		}
		names = append(names, object.Name)
	}
	sort.Strings(names)
	return names, nil
}
