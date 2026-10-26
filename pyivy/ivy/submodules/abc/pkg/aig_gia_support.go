package abc

// CI support collection from src/aig/gia/giaDfs.c:Gia_ManCollectCis.

// CollectInputs returns reachable CI object IDs in first-encounter DFS order.
// It stops at register outputs as well as primary inputs, skips constants,
// accepts CO wrappers and follows native mux controls. It does not compute
// functional support after Boolean simplification. maxObjects zero selects
// one million objects. Manager traversal scratch is not modified.
func (m *GiaManager) CollectInputs(roots []GiaID, maxObjects int) ([]GiaID, error) {
	if maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > maxObjects || len(roots) > maxObjects {
		return nil, ErrLimit
	}
	for _, root := range roots {
		if uint64(root) >= uint64(len(m.objects)) {
			return nil, ErrInvalidLiteral
		}
	}
	seen := make([]bool, len(m.objects))
	seen[0] = true
	result := make([]GiaID, 0)
	stack := make([]GiaID, 0)
	for _, root := range roots {
		stack = append(stack, root)
		for len(stack) > 0 {
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[id] {
				continue
			}
			seen[id] = true
			if m.objects[id].IsCI() {
				result = append(result, id)
				continue
			}
			children, n, err := m.crossCutFanins(id)
			if err != nil {
				return nil, err
			}
			for i := n - 1; i >= 0; i-- {
				stack = append(stack, children[i])
			}
		}
	}
	return result, nil
}

// AndObjects ports Gia_ManCollectAndsAll. It returns every AND-like object in
// ID order, including dangling buffers and native XOR/mux nodes.
func (m *GiaManager) AndObjects(maxObjects int) ([]GiaID, error) {
	if maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > maxObjects {
		return nil, ErrLimit
	}
	result := make([]GiaID, 0)
	for i, obj := range m.objects {
		if obj.IsAnd() {
			result = append(result, GiaID(i))
		}
	}
	return result, nil
}
