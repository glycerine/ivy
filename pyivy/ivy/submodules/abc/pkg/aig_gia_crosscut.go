package abc

// Cross-cut traversal from src/aig/gia/giaUtil.c:Gia_ManDfsForCrossCut
// and Gia_ManCrossCut. Iterative traversal avoids recursion depth limits.

// DepthFirstCrossCutOrder visits output cones in postorder, preserving packed
// fanin order. Constant-driven COs are skipped as in C. Go also handles internal
// constants, buffers and native XOR/mux edges explicitly; C assumes plain AIGs.
func (m *GiaManager) DepthFirstCrossCutOrder(reverseOutputs bool, maxObjects int) ([]GiaID, error) {
	order, _, err := m.crossCutTraversal(reverseOutputs, maxObjects)
	return order, err
}

func (m *GiaManager) crossCutFanins(id GiaID) (children [3]GiaID, n int, err error) {
	obj := m.objects[id]
	if !obj.IsAnd() && !obj.IsCO() {
		return children, 0, nil
	}
	a, err := obj.Fanin0(id)
	if err != nil {
		return children, 0, err
	}
	children[0] = a.ID()
	n = 1
	if obj.IsAnd() && !obj.IsBuffer() {
		b, err := obj.Fanin1(id)
		if err != nil {
			return children, 0, err
		}
		children[n] = b.ID()
		n++
	}
	if c, yes := m.MuxControl(id); yes {
		children[n] = c.ID()
		n++
	}
	return children, n, nil
}
func (m *GiaManager) crossCutTraversal(reverse bool, maxObjects int) ([]GiaID, []uint64, error) {
	if maxObjects < 0 {
		return nil, nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, nil, err
	}
	if len(m.objects) > maxObjects {
		return nil, nil, ErrLimit
	}
	seen := make([]bool, len(m.objects))
	refs := make([]uint64, len(m.objects))
	order := make([]GiaID, 0)
	type visit struct {
		id   GiaID
		exit bool
	}
	stack := make([]visit, 0)
	for index := range m.cos {
		at := index
		if reverse {
			at = len(m.cos) - 1 - index
		}
		root := m.cos[at]
		driver, err := m.objects[root].Fanin0(root)
		if err != nil {
			return nil, nil, err
		}
		if driver.ID() == 0 {
			continue
		}
		stack = append(stack, visit{id: root})
		for len(stack) > 0 {
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if item.exit {
				order = append(order, item.id)
				continue
			}
			if seen[item.id] {
				continue
			}
			seen[item.id] = true
			children, n, err := m.crossCutFanins(item.id)
			if err != nil {
				return nil, nil, err
			}
			for _, child := range children[:n] {
				refs[child]++
			}
			stack = append(stack, visit{id: item.id, exit: true})
			for i := n - 1; i >= 0; i-- {
				stack = append(stack, visit{id: children[i]})
			}
		}
	}
	return order, refs, nil
}

// CrossCut measures the peak number of live values during the corresponding
// DFS order. As in ABC, peak size is sampled before consuming a node's fanins.
// Graph storage and scratch values are never modified.
func (m *GiaManager) CrossCut(reverseOutputs bool, maxObjects int) (int, error) {
	order, refs, err := m.crossCutTraversal(reverseOutputs, maxObjects)
	if err != nil {
		return 0, err
	}
	current, maximum := 0, 0
	for _, id := range order {
		if refs[id] != 0 {
			current++
		}
		maximum = max(maximum, current)
		children, n, err := m.crossCutFanins(id)
		if err != nil {
			return 0, err
		}
		for _, child := range children[:n] {
			if refs[child] == 0 {
				return 0, ErrInvariant
			}
			refs[child]--
			if refs[child] == 0 {
				current--
			}
		}
	}
	if current != 0 {
		return 0, ErrInvariant
	}
	return maximum, nil
}
