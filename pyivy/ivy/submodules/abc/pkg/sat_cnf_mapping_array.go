package abc

// Packed mapping from sat/cnf/cnfWrite.c:Cnf_ManWriteCnfMapping.

type CNFMappingArrayOptions struct {
	MaxObjects, MaxEntries int // zero: 1M objects, 8M packed entries
	Stop                   func() bool
}

// WriteCNFMappingArray4 writes the native object-offset table followed by
// five-entry records in supplied cut order: stretched 32-bit truth and four
// leaf IDs padded with -1. Zero offsets indicate unmapped objects. The truth
// entry is signed int32, preserving the native bit pattern. Cut functions are
// trusted; output order and cone membership are the caller's responsibility.
func WriteCNFMappingArray4(g *AigManager, cuts []AigCNFCut4, o CNFMappingArrayOptions) ([]int32, error) {
	if g == nil || o.MaxObjects < 0 || o.MaxEntries < 0 {
		return nil, ErrDimension
	}
	if o.MaxObjects == 0 {
		o.MaxObjects = 1 << 20
	}
	if o.MaxEntries == 0 {
		o.MaxEntries = 8 << 20
	}
	if len(g.objects) > o.MaxObjects {
		return nil, ErrLimit
	}
	check := func() error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		return nil
	}
	if e := check(); e != nil {
		return nil, e
	}
	size, e := checkedSize(len(g.objects), 5, len(cuts), min(o.MaxEntries, int(^uint32(0)>>1)))
	if e != nil {
		return nil, e
	}
	if e = g.Check(); e != nil {
		return nil, e
	}
	result := make([]int32, size)
	offset := len(g.objects)
	for _, cut := range cuts {
		if e := check(); e != nil {
			return nil, e
		}
		if uint64(cut.Root) >= uint64(len(g.objects)) || g.objects[cut.Root].Type != AigAND || result[cut.Root] != 0 || len(cut.Leaves) > 4 {
			return nil, ErrDimension
		}
		mask := (1 << uint(len(cut.Leaves))) - 1
		for a := 0; a < 16; a++ {
			if cut.Truth>>uint(a)&1 != cut.Truth>>uint(a&mask)&1 {
				return nil, ErrDimension
			}
		}
		for _, leaf := range cut.Leaves {
			if leaf >= cut.Root {
				return nil, ErrDimension
			}
			typ := g.objects[leaf].Type
			if typ != AigAND && typ != AigCI && typ != AigConstant1 {
				return nil, ErrDimension
			}
		}
		result[cut.Root] = int32(offset)
		result[offset] = int32(uint32(cut.Truth) * 0x10001)
		for i := 0; i < 4; i++ {
			result[offset+1+i] = -1
		}
		for i, id := range cut.Leaves {
			result[offset+1+i] = int32(id)
		}
		offset += 5
	}
	if e := check(); e != nil {
		return nil, e
	}
	return result, nil
}

// DeriveAigCNFMappingArray4 matches native Cnf_DeriveMappingArray's postorder
// scan with default ten-slot minimized DAR cuts. Source graph remains unchanged.
func DeriveAigCNFMappingArray4(g *AigManager, cuts DARCutOptions, o CNFMappingArrayOptions) ([]int32, error) {
	cutStop, arrayStop := cuts.Stop, o.Stop
	stop := func() bool { return cutStop != nil && cutStop() || arrayStop != nil && arrayStop() }
	cuts.Stop = stop
	o.Stop = stop
	derived, e := ComputeDARCuts4(g, cuts)
	if e != nil {
		return nil, e
	}
	mapping, e := SelectCNFAreaFlow4(derived.Graph, derived.Candidates, CNFAreaFlowOptions{MaxObjects: o.MaxObjects, Stop: stop})
	if e != nil {
		return nil, e
	}
	scan, e := ScanCNFMapping4(derived.Graph, mapping.Cuts, CNFMappingScanOptions{MaxObjects: o.MaxObjects, Stop: stop})
	if e != nil {
		return nil, e
	}
	return WriteCNFMappingArray4(derived.Graph, scan.Cuts, o)
}
