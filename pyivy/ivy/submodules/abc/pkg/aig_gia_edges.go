package abc

// Two-slot edge assignments from aig/gia/giaEdge.c: Gia_ObjEdgeCount/Add,
// Gia_ManEdgeFromArray/EdgeToArray and Gia_ManConvertPackingToEdges.

type GiaEdgeOptions struct {
	MaxObjects int // arena entries; zero: one million
	MaxPairs   int // attempted undirected insertions; zero: one million
	Stop       func() bool
}

// GiaEdgeAssignment owns native two-slot rows. Zero denotes an empty slot.
// Duplicates consume slots; overflow skips each directed insertion separately,
// so rows can be asymmetric. Skipped reports directed overflow attempts.
type GiaEdgeAssignment struct {
	rows    [][2]GiaID
	skipped int
}

func (e *GiaEdgeAssignment) Rows() [][2]GiaID {
	if e == nil {
		return nil
	}
	return append([][2]GiaID(nil), e.rows...)
}
func (e *GiaEdgeAssignment) Skipped() int {
	if e == nil {
		return 0
	}
	return e.skipped
}
func (e *GiaEdgeAssignment) Count(id GiaID) (int, error) {
	if e == nil || uint64(id) >= uint64(len(e.rows)) {
		return 0, ErrDimension
	}
	count := 0
	for _, v := range e.rows[id] {
		if v > 0 {
			count++
		}
	}
	return count, nil
}

// HasEdge follows native slot equality, including zero matching an empty slot.
func (e *GiaEdgeAssignment) HasEdge(id, next GiaID) (bool, error) {
	if e == nil || uint64(id) >= uint64(len(e.rows)) || uint64(next) >= uint64(len(e.rows)) {
		return false, ErrDimension
	}
	return e.rows[id][0] == next || e.rows[id][1] == next, nil
}

// Pairs exports nonzero lower neighbors in object/slot order. It does not
// repair asymmetry or deduplicate, matching Gia_ManEdgeToArray.
func (e *GiaEdgeAssignment) Pairs() [][2]GiaID {
	if e == nil {
		return nil
	}
	pairs := make([][2]GiaID, 0)
	for id, row := range e.rows {
		for _, neighbor := range row {
			if neighbor != 0 && uint64(neighbor) < uint64(id) {
				pairs = append(pairs, [2]GiaID{neighbor, GiaID(id)})
			}
		}
	}
	return pairs
}
func (e *GiaEdgeAssignment) add(id, next GiaID) {
	if e.rows[id][0] == 0 {
		e.rows[id][0] = next
	} else if e.rows[id][1] == 0 {
		e.rows[id][1] = next
	} else {
		e.skipped++
	}
}
func newGiaEdges(objects int, options GiaEdgeOptions) (*GiaEdgeAssignment, int, error) {
	if objects < 1 || options.MaxObjects < 0 || options.MaxPairs < 0 {
		return nil, 0, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxPairs == 0 {
		options.MaxPairs = 1 << 20
	}
	if objects > options.MaxObjects {
		return nil, 0, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, 0, ErrCancelled
	}
	return &GiaEdgeAssignment{rows: make([][2]GiaID, objects)}, options.MaxPairs, nil
}
func GiaEdgesFromPairs(objects int, pairs [][2]GiaID, options GiaEdgeOptions) (*GiaEdgeAssignment, error) {
	e, limit, err := newGiaEdges(objects, options)
	if err != nil {
		return nil, err
	}
	if len(pairs) > limit {
		return nil, ErrLimit
	}
	for _, pair := range pairs {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if pair[0] >= pair[1] || uint64(pair[1]) >= uint64(objects) {
			return nil, ErrDimension
		}
		e.add(pair[0], pair[1])
		e.add(pair[1], pair[0])
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return e, nil
}

// Edges converts two-member groups into one edge, and three-member groups
// into first-to-third and second-to-third edges, preserving native insertion
// direction/order. One-member groups add nothing. Source packing is unchanged.
func (p *GiaLUTPacking) Edges(options GiaEdgeOptions) (*GiaEdgeAssignment, error) {
	if p == nil || p.mapping == nil || len(p.packed) == 0 {
		return nil, ErrDimension
	}
	e, limit, err := newGiaEdges(p.mapping.objects, options)
	if err != nil {
		return nil, err
	}
	pairs := 0
	for position := 1; position < len(p.packed); {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		count := p.packed[position]
		position++
		if count > 3 {
			return nil, ErrUnsupported
		}
		nodes := p.packed[position : position+count]
		position += count
		needed := count - 1
		if needed > limit-pairs {
			return nil, ErrLimit
		}
		pairs += needed
		if count == 2 {
			e.add(GiaID(nodes[0]), GiaID(nodes[1]))
			e.add(GiaID(nodes[1]), GiaID(nodes[0]))
		} else if count == 3 {
			e.add(GiaID(nodes[0]), GiaID(nodes[2]))
			e.add(GiaID(nodes[2]), GiaID(nodes[0]))
			e.add(GiaID(nodes[1]), GiaID(nodes[2]))
			e.add(GiaID(nodes[2]), GiaID(nodes[1]))
		}
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return e, nil
}
