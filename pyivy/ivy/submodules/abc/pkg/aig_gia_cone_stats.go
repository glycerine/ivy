package abc

// Cone/support cardinalities from src/aig/gia/giaDfs.c:
// Gia_ManSuppSize, Gia_ManSuppSizeOne and Gia_ManConeSize.

type GiaConeStatistics struct{ Inputs, Gates int }

// ConeStatistics counts distinct structural CI support and AND-like nodes in
// the union of root cones. Constants/CO wrappers are not counted. Registers
// are CI boundaries, and native mux controls participate. This single iterative
// pass avoids materializing traversal output lists. Zero cap selects one million.
func (m *GiaManager) ConeStatistics(roots []GiaID, maxObjects int) (GiaConeStatistics, error) {
	var result GiaConeStatistics
	if maxObjects < 0 {
		return result, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if err := m.Check(); err != nil {
		return result, err
	}
	if len(m.objects) > maxObjects || len(roots) > maxObjects {
		return result, ErrLimit
	}
	for _, root := range roots {
		if uint64(root) >= uint64(len(m.objects)) {
			return result, ErrInvalidLiteral
		}
	}
	seen := make([]bool, len(m.objects))
	seen[0] = true
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
				result.Inputs++
				continue
			}
			if m.objects[id].IsAnd() {
				result.Gates++
			}
			children, n, err := m.crossCutFanins(id)
			if err != nil {
				return GiaConeStatistics{}, err
			}
			for i := n - 1; i >= 0; i-- {
				stack = append(stack, children[i])
			}
		}
	}
	return result, nil
}

func (m *GiaManager) SupportSize(roots []GiaID, maxObjects int) (int, error) {
	result, err := m.ConeStatistics(roots, maxObjects)
	return result.Inputs, err
}
func (m *GiaManager) ConeSize(roots []GiaID, maxObjects int) (int, error) {
	result, err := m.ConeStatistics(roots, maxObjects)
	return result.Gates, err
}
