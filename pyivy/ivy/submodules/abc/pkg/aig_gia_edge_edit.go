package abc

// Directed slot edits from aig/gia/giaEdge.c:
// Gia_ObjEdgeAdd, Gia_ObjEdgeRemove and Gia_ObjEdgeClean.

type GiaEdgeEditKind uint8

const (
	GiaEdgeInsert GiaEdgeEditKind = iota
	GiaEdgeRemove
	GiaEdgeClear
)

type GiaEdgeEdit struct {
	Object, Neighbor GiaID
	Kind             GiaEdgeEditKind
}

// WithEdits applies ordered directed edits to an owned copy. Reciprocal rows
// are not changed. Insert retains duplicate/zero-sentinel/overflow behavior;
// skipped insertions accumulate in Skipped. Remove compacts the second slot
// when the first matches, then clears the second, as in C. Removing a missing
// neighbor rejects C's assertion precondition. Clear ignores Neighbor.
// MaxPairs bounds edit count here; limits/cancellation/errors expose no result
// and leave the original assignment unchanged.
func (e *GiaEdgeAssignment) WithEdits(edits []GiaEdgeEdit, options GiaEdgeOptions) (*GiaEdgeAssignment, error) {
	if e == nil || len(e.rows) == 0 {
		return nil, ErrDimension
	}
	result, limit, err := newGiaEdges(len(e.rows), options)
	if err != nil {
		return nil, err
	}
	if len(edits) > limit || len(edits) > int(^uint(0)>>1)-e.skipped {
		return nil, ErrLimit
	}
	copy(result.rows, e.rows)
	result.skipped = e.skipped
	for _, edit := range edits {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if uint64(edit.Object) >= uint64(len(e.rows)) {
			return nil, ErrDimension
		}
		if edit.Kind != GiaEdgeClear && uint64(edit.Neighbor) >= uint64(len(e.rows)) {
			return nil, ErrDimension
		}
		switch edit.Kind {
		case GiaEdgeInsert:
			result.add(edit.Object, edit.Neighbor)
		case GiaEdgeRemove:
			row := &result.rows[edit.Object]
			if row[0] != edit.Neighbor && row[1] != edit.Neighbor {
				return nil, ErrInvariant
			}
			if row[0] == edit.Neighbor {
				row[0] = row[1]
			}
			row[1] = 0
		case GiaEdgeClear:
			result.rows[edit.Object] = [2]GiaID{}
		default:
			return nil, ErrDimension
		}
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	return result, nil
}
