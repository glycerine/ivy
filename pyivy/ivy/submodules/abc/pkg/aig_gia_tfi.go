package abc

// Transitive-fanin collection from src/aig/gia/giaDfs.c:Gia_ManCollectTfi.

// CollectTransitiveFanins returns each reachable AND-like node once, in fanin
// postorder, including gate roots. CIs/constants are boundaries and omitted;
// CO roots are accepted as wrappers and omitted. Native mux controls are
// traversed explicitly. roots are object IDs, not literals. The iterative
// traversal owns its visited state. maxObjects zero selects one million.
func (m *GiaManager) CollectTransitiveFanins(roots []GiaID, maxObjects int) ([]GiaID, error) {
	return m.CollectConeAnds(roots, nil, maxObjects)
}

// CollectConeAnds ports Gia_ManCollectAnds with explicit leaf boundaries.
// Leaves are excluded and their fanins are not traversed. Every call owns
// fresh visitation state, replacing C's caller-managed traversal-ID contract.
func (m *GiaManager) CollectConeAnds(roots, leaves []GiaID, maxObjects int) ([]GiaID, error) {
	return m.collectConeNodes(roots, leaves, maxObjects, false)
}

// CollectNodesAndInputs ports Gia_ManCollectNodesCis: combined gate/CI
// postorder, excluding constants and CO wrappers, with each object once.
func (m *GiaManager) CollectNodesAndInputs(roots []GiaID, maxObjects int) ([]GiaID, error) {
	return m.collectConeNodes(roots, nil, maxObjects, true)
}

func (m *GiaManager) collectConeNodes(roots, leaves []GiaID, maxObjects int, includeInputs bool) ([]GiaID, error) {
	if maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > maxObjects || len(roots) > maxObjects || len(leaves) > maxObjects {
		return nil, ErrLimit
	}
	for _, id := range roots {
		if uint64(id) >= uint64(len(m.objects)) {
			return nil, ErrInvalidLiteral
		}
	}
	for _, id := range leaves {
		if uint64(id) >= uint64(len(m.objects)) {
			return nil, ErrInvalidLiteral
		}
	}
	seen := make([]bool, len(m.objects))
	seen[0] = true
	for _, id := range leaves {
		seen[id] = true
	}
	result := make([]GiaID, 0)
	type visit struct {
		id   GiaID
		exit bool
	}
	stack := make([]visit, 0)
	for _, root := range roots {
		stack = append(stack, visit{id: root})
		for len(stack) > 0 {
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if item.exit {
				if m.objects[item.id].IsAnd() || (includeInputs && m.objects[item.id].IsCI()) {
					result = append(result, item.id)
				}
				continue
			}
			if seen[item.id] {
				continue
			}
			seen[item.id] = true
			children, n, err := m.crossCutFanins(item.id)
			if err != nil {
				return nil, err
			}
			stack = append(stack, visit{id: item.id, exit: true})
			for i := n - 1; i >= 0; i-- {
				stack = append(stack, visit{id: children[i]})
			}
		}
	}
	return result, nil
}
