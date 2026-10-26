package abc

// Reverse edge selection from aig/gia/giaEdge.c:
// Gia_ObjComputeEdgeDelay2 and Gia_ManComputeEdgeDelay2 (plain mapping).

// GiaReverseEdgeTiming owns reverse unit distances. Maximum scans CIs, not
// outputs; every mapped LUT participates, including dangling logic. Selected
// edges remove one unit of distance. This differs from forward 2/10 timing.
type GiaReverseEdgeTiming struct {
	Distances      []int
	Maximum, Edges int
}

// SelectEdgesReverse processes LUTs in descending object order, retaining
// native critical-fanout ties, duplicate slots and capacity/insertion order.
// It requires no timing boxes and supports plain packed mappings. Private
// scratch preserves source graph/mapping state; bounded fanin work and
// cancellation return no partial rows or distances.
func (mapping *GiaLUTMapping) SelectEdgesReverse(m *GiaManager, options GiaLUTAnalysisOptions) (*GiaEdgeAssignment, GiaReverseEdgeTiming, error) {
	fail := func(err error) (*GiaEdgeAssignment, GiaReverseEdgeTiming, error) {
		return nil, GiaReverseEdgeTiming{}, err
	}
	if mapping == nil || m == nil || options.MaxObjects < 0 || options.MaxVisits < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxVisits == 0 {
		options.MaxVisits = 1 << 20
	}
	if len(m.objects) > options.MaxObjects {
		return fail(ErrLimit)
	}
	if err := mapping.checkGraph(m); err != nil {
		return fail(err)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	n := len(m.objects)
	r := GiaReverseEdgeTiming{Distances: make([]int, n)}
	edges := &GiaEdgeAssignment{rows: make([][2]GiaID, n)}
	first, second := make([]GiaID, n), make([]GiaID, n)
	ties := make([]int, n)
	visits := 0
	roots := mapping.Roots()
	for index := len(roots) - 1; index >= 0; index-- {
		if stopped() {
			return fail(ErrCancelled)
		}
		root := roots[index]
		edges.rows[root] = [2]GiaID{}
		distance := r.Distances[root]
		capacity := func(id GiaID) bool { count, _ := edges.Count(id); return count <= 1 }
		if distance != 0 {
			if ties[root] == 1 && capacity(first[root]) {
				edges.add(first[root], root)
				edges.add(root, first[root])
				distance--
			} else if ties[root] == 2 && capacity(first[root]) && capacity(second[root]) {
				edges.add(first[root], root)
				edges.add(second[root], root)
				edges.add(root, first[root])
				edges.add(root, second[root])
				distance--
			}
		}
		r.Distances[root] = distance
		next, err := checkedSize(distance, 1, 1, int(^uint(0)>>1))
		if err != nil {
			return fail(err)
		}
		offset := mapping.packed[root]
		count := mapping.packed[offset]
		for _, fanin := range mapping.packed[offset+1 : offset+count+1] {
			if stopped() {
				return fail(ErrCancelled)
			}
			if visits == options.MaxVisits {
				return fail(ErrLimit)
			}
			visits++
			if r.Distances[fanin] < next {
				r.Distances[fanin] = next
				first[fanin] = root
				ties[fanin] = 1
			} else if r.Distances[fanin] == next {
				second[fanin] = root
				ties[fanin]++
			}
		}
	}
	for _, id := range m.cis {
		if stopped() {
			return fail(ErrCancelled)
		}
		r.Maximum = max(r.Maximum, r.Distances[id])
	}
	positive := 0
	for _, row := range edges.rows {
		for _, id := range row {
			if id > 0 {
				positive++
			}
		}
	}
	r.Edges = positive / 2
	if stopped() {
		return fail(ErrCancelled)
	}
	return edges, r, nil
}
