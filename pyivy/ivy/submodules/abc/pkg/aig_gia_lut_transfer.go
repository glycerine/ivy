package abc

// Mapping transfer from aig/gia/giaIf.c: Gia_ManTransferMapping.

type GiaLUTTransferOptions struct {
	MaxEntries int // complete target packed storage, including stale records; zero: one million
	Stop       func() bool
}

// Transfer moves mapping records through a source-object to target-literal
// correspondence. InvalidGiaLit drops an object. Dropped roots are skipped;
// dropped fanins are removed. Fanin/footer phases are discarded as in C, but
// complemented surviving LUT roots reject C's assertion precondition. A later
// source root replaces an earlier record at the same target root, retaining
// the earlier packed record as native stale storage. Object-zero records are
// skipped. A dropped footer of a surviving LUT rejects instead of producing
// C's invalid index. Both inputs remain unchanged and failures expose no result.
// The caller must establish the correspondence's functional correctness;
// this operation transfers IDs and does not prove graph equivalence.
func (mapping *GiaLUTMapping) Transfer(targetObjects int, correspondence []GiaLit, options GiaLUTTransferOptions) (*GiaLUTMapping, error) {
	if mapping == nil || targetObjects < 1 || len(correspondence) != mapping.objects || options.MaxEntries < 0 {
		return nil, ErrDimension
	}
	limit := options.MaxEntries
	if limit == 0 {
		limit = 1 << 20
	}
	limit = min(limit, int(^uint32(0)>>1))
	if targetObjects > limit {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	for _, lit := range correspondence {
		if lit != InvalidGiaLit && (!lit.Valid() || uint64(lit.ID()) >= uint64(targetObjects)) {
			return nil, ErrInvalidLiteral
		}
	}
	packed := make([]int, targetObjects)
	for _, root := range mapping.Roots() {
		if stopped() {
			return nil, ErrCancelled
		}
		target := correspondence[root]
		if target == InvalidGiaLit {
			continue
		}
		if target.Complemented() {
			return nil, ErrDimension
		}
		offset := mapping.packed[root]
		count := mapping.packed[offset]
		footer := mapping.packed[offset+count+1]
		footerID := footer
		if footerID < 0 {
			footerID = -footerID
		}
		mappedFooter := correspondence[footerID]
		if mappedFooter == InvalidGiaLit {
			return nil, ErrInvalidLiteral
		}
		kept := 0
		for _, fanin := range mapping.packed[offset+1 : offset+1+count] {
			if correspondence[fanin] != InvalidGiaLit {
				kept++
			}
		}
		if _, err := checkedSize(len(packed), 1, kept+2, limit); err != nil {
			return nil, err
		}
		packed[target.ID()] = len(packed)
		packed = append(packed, kept)
		for _, fanin := range mapping.packed[offset+1 : offset+1+count] {
			if stopped() {
				return nil, ErrCancelled
			}
			value := correspondence[fanin]
			if value == InvalidGiaLit {
				continue
			}
			if value.ID() >= target.ID() {
				return nil, ErrDimension
			}
			packed = append(packed, int(value.ID()))
		}
		footer = int(mappedFooter.ID())
		if mapping.packed[offset+count+1] < 0 {
			footer = -footer
		}
		packed = append(packed, footer)
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return &GiaLUTMapping{objects: targetObjects, packed: packed}, nil
}
