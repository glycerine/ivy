package abc

// Mapping connectivity from aig/gia/giaIf.c:
// Gia_ManMappingVerify_rec and Gia_ManMappingVerify.

// Verify checks that every non-buffer gate reached through LUT fanins from
// CO drivers or any buffer driver has a mapping record. Dangling buffers are
// checked, while unrelated dangling gates may remain unmapped, as in C.
// Constants, CIs and buffers terminate each mapped dependency traversal.
// This checks mapped connectivity; Analyze/Truth6 separately check structural
// boundaries and functions. An incomplete mapping returns false, nil. Invalid
// graphs/options, resource limits and cancellation return an error. Source
// marks, traversal IDs and values remain unchanged; traversal is iterative.
func (mapping *GiaLUTMapping) Verify(m *GiaManager, options GiaLUTAnalysisOptions) (bool, error) {
	if m == nil || mapping == nil || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return false, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(m.objects) > options.MaxObjects {
		return false, ErrLimit
	}
	if err := mapping.checkGraph(m); err != nil {
		return false, err
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return false, ErrCancelled
	}
	roots := make([]GiaID, 0)
	for id, obj := range m.objects {
		if !obj.IsBuffer() {
			continue
		}
		driver, err := obj.Fanin0(GiaID(id))
		if err != nil {
			return false, err
		}
		roots = append(roots, driver.ID())
	}
	for _, id := range m.cos {
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return false, err
		}
		roots = append(roots, driver.ID())
	}
	seen := make([]bool, len(m.objects))
	stack := make([]GiaID, 0)
	visits := 0
	for _, root := range roots {
		stack = append(stack[:0], root)
		for len(stack) > 0 {
			if stopped() {
				return false, ErrCancelled
			}
			id := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if seen[id] {
				continue
			}
			seen[id] = true
			object := m.objects[id]
			if !object.IsAnd() || object.IsBuffer() {
				continue
			}
			if visits == options.MaxVisits {
				return false, ErrLimit
			}
			visits++
			offset := mapping.packed[id]
			if offset == 0 {
				return false, nil
			}
			count := mapping.packed[offset]
			for i := count - 1; i >= 0; i-- {
				fanin := GiaID(mapping.packed[offset+1+i])
				if !seen[fanin] {
					stack = append(stack, fanin)
				}
			}
		}
	}
	if stopped() {
		return false, ErrCancelled
	}
	return true, nil
}
