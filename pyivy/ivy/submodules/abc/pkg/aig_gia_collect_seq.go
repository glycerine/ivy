package abc

// Sequential cone collection from src/aig/gia/giaDfs.c:Gia_ManCollectSeq.

// CollectSequential returns constant zero followed by postorder object IDs
// reachable from selected primary-output indices and their transitive register
// input cones. Register inputs discovered at CIs are queued after the initial
// outputs, matching ABC's dynamic root order. nil selects no outputs, not all.
// Native mux controls participate explicitly. The traversal is iterative and
// does not mutate manager traversal IDs. maxObjects zero selects one million.
func (m *GiaManager) CollectSequential(outputs []int, maxObjects int) ([]GiaID, error) {
	if maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if len(m.objects) > maxObjects || len(outputs) > maxObjects {
		return nil, ErrLimit
	}
	for _, output := range outputs {
		if output < 0 || output >= m.PrimaryOutputCount() {
			return nil, ErrDimension
		}
	}
	roots := make([]GiaID, 0, len(outputs))
	for _, output := range outputs {
		roots = append(roots, m.cos[output])
	}
	seen := make([]bool, len(m.objects))
	seen[0] = true
	result := []GiaID{0}
	type visit struct {
		id   GiaID
		exit bool
	}
	stack := make([]visit, 0)
	for rootIndex := 0; rootIndex < len(roots); rootIndex++ {
		stack = append(stack, visit{id: roots[rootIndex]})
		for len(stack) > 0 {
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if item.exit {
				result = append(result, item.id)
				continue
			}
			if seen[item.id] {
				continue
			}
			seen[item.id] = true
			obj := m.objects[item.id]
			if obj.IsCI() && int(obj.Diff1()) >= m.PrimaryInputCount() {
				ri, err := m.RegisterInputForOutput(item.id)
				if err != nil {
					return nil, err
				}
				roots = append(roots, ri)
			}
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
