package abc

// MFFC size/support from src/aig/gia/giaMffc.c:Gia_NodeMffcSizeSupp.

type GiaMffc struct {
	Size     int
	Boundary []GiaID
	// Deltas are remaining references minus original references at each leaf,
	// matching C's vSuppRefs after restoring the graph (nonpositive values).
	BoundaryReferenceDeltas []int64
}

// Mffc computes the maximum fanout-free cone rooted at an AND-like object.
// Boundary order is first-encounter fanin DFS. Owned counters eliminate C's
// temporary manager mutations and reference restoration. Native mux controls,
// buffers and constants use their actual edges. Zero cap selects one million.
func (m *GiaManager) Mffc(root GiaID, maxObjects int) (GiaMffc, error) {
	var result GiaMffc
	refs, err := m.ReferenceCounts(maxObjects)
	if err != nil {
		return result, err
	}
	if uint64(root) >= uint64(len(m.objects)) || !m.objects[root].IsAnd() {
		return result, ErrInvalidLiteral
	}
	original := append([]uint64(nil), refs...)
	stack := []GiaID{root}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		result.Size++
		children, n, err := m.crossCutFanins(id)
		if err != nil {
			return GiaMffc{}, err
		}
		for _, child := range children[:n] {
			if refs[child] == 0 {
				return GiaMffc{}, ErrInvariant
			}
			refs[child]--
			if refs[child] == 0 && m.objects[child].IsAnd() {
				stack = append(stack, child)
			}
		}
	}
	seen := make([]bool, len(m.objects))
	children, n, err := m.crossCutFanins(root)
	if err != nil {
		return GiaMffc{}, err
	}
	for i := n - 1; i >= 0; i-- {
		stack = append(stack, children[i])
	}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[id] {
			continue
		}
		seen[id] = true
		if refs[id] != 0 || !m.objects[id].IsAnd() {
			result.Boundary = append(result.Boundary, id)
			result.BoundaryReferenceDeltas = append(result.BoundaryReferenceDeltas, int64(refs[id])-int64(original[id]))
			continue
		}
		children, n, err := m.crossCutFanins(id)
		if err != nil {
			return GiaMffc{}, err
		}
		for i := n - 1; i >= 0; i-- {
			stack = append(stack, children[i])
		}
	}
	return result, nil
}
