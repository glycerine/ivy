package abc

// Static fanout indexing from src/aig/gia/giaFanout.c:Gia_ManStaticFanoutStart.

import "slices"

// GiaFanouts is an immutable graph snapshot with one entry per incoming edge.
// Repeated uses of a signal produce repeated fanout IDs, in object-ID order.
// It is not updated when its source graph changes.
type GiaFanouts struct {
	offsets []int
	targets []GiaID
}

func (f *GiaFanouts) Fanouts(id GiaID) ([]GiaID, error) {
	if f == nil || uint64(id)+1 >= uint64(len(f.offsets)) {
		return nil, ErrInvalidLiteral
	}
	return slices.Clone(f.targets[f.offsets[id]:f.offsets[int(id)+1]]), nil
}

// StaticFanouts returns owned fanout storage without changing manager scratch.
// Zero limits select one million objects and three million edges.
func (m *GiaManager) StaticFanouts(maxObjects, maxEdges int) (*GiaFanouts, error) {
	if maxEdges < 0 {
		return nil, ErrDimension
	}
	if maxEdges == 0 {
		maxEdges = 3 << 20
	}
	counts, err := m.ReferenceCounts(maxObjects)
	if err != nil {
		return nil, err
	}
	offsets := make([]int, len(counts)+1)
	for i, count := range counts {
		if count > uint64(maxEdges) {
			return nil, ErrLimit
		}
		offsets[i+1], err = checkedSize(offsets[i], 1, int(count), maxEdges)
		if err != nil {
			return nil, err
		}
	}
	targets := make([]GiaID, offsets[len(counts)])
	positions := slices.Clone(offsets[:len(counts)])
	for i := range m.objects {
		children, n, err := m.crossCutFanins(GiaID(i))
		if err != nil {
			return nil, err
		}
		for _, child := range children[:n] {
			targets[positions[child]] = GiaID(i)
			positions[child]++
		}
	}
	return &GiaFanouts{offsets: offsets, targets: targets}, nil
}

// CollectTransitiveFanouts ports Gia_ManCollectTfo using an owned static index
// and iterative fanout postorder. COs terminate traversal and are omitted.
// Only AND-like objects are returned. CI/constant roots are accepted as an
// extension to C's gate-root precondition. Native mux controls participate.
func (m *GiaManager) CollectTransitiveFanouts(roots []GiaID, maxObjects, maxEdges int) ([]GiaID, error) {
	fanouts, err := m.StaticFanouts(maxObjects, maxEdges)
	if err != nil {
		return nil, err
	}
	bound := maxObjects
	if bound == 0 {
		bound = 1 << 20
	}
	if len(roots) > bound {
		return nil, ErrLimit
	}
	for _, root := range roots {
		if uint64(root) >= uint64(len(m.objects)) {
			return nil, ErrInvalidLiteral
		}
	}
	seen := make([]bool, len(m.objects))
	order := make([]GiaID, 0)
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
				if m.objects[item.id].IsAnd() {
					order = append(order, item.id)
				}
				continue
			}
			if seen[item.id] {
				continue
			}
			seen[item.id] = true
			if m.objects[item.id].IsCO() {
				continue
			}
			stack = append(stack, visit{id: item.id, exit: true})
			edges := fanouts.targets[fanouts.offsets[item.id]:fanouts.offsets[int(item.id)+1]]
			for i := len(edges) - 1; i >= 0; i-- {
				stack = append(stack, visit{id: edges[i]})
			}
		}
	}
	return order, nil
}
