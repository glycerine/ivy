package abc

// Packing records, statistics and transfer from aig/gia/giaIf.c:
// Gia_ManPrintPackingStats and Gia_ManTransferPacking.

import "slices"

// GiaLUTPacking owns [group count, group size, LUT IDs..., ...]. Native order
// and repeated IDs are preserved. The mapping is immutable and identifies the
// object arena used by these records. Some mapping paths create one-to-three
// member clusters; giaPack.c also permits caller-selected wider blocks.
type GiaLUTPacking struct {
	mapping *GiaLUTMapping
	packed  []int
}

type GiaLUTPackingStatistics struct {
	Groups, LUTs, MaxSize int
	BySize                []int
}

func packingLimit(limit int) (int, error) {
	if limit < 0 {
		return 0, ErrDimension
	}
	if limit == 0 {
		limit = 1 << 20
	}
	return min(limit, int(^uint32(0)>>1)), nil
}
func NewGiaLUTPacking(mapping *GiaLUTMapping, groups [][]GiaID, maxEntries int) (*GiaLUTPacking, error) {
	if mapping == nil || mapping.objects < 1 {
		return nil, ErrDimension
	}
	limit, err := packingLimit(maxEntries)
	if err != nil {
		return nil, err
	}
	total := 1
	for _, group := range groups {
		if len(group) < 1 {
			return nil, ErrDimension
		}
		total, err = checkedSize(total, 1, len(group)+1, limit)
		if err != nil {
			return nil, err
		}
	}
	packed := make([]int, 1, total)
	packed[0] = len(groups)
	for _, group := range groups {
		packed = append(packed, len(group))
		for _, id := range group {
			if uint64(id) >= uint64(mapping.objects) || mapping.packed[id] == 0 {
				return nil, ErrDimension
			}
			packed = append(packed, int(id))
		}
	}
	return &GiaLUTPacking{mapping: mapping, packed: packed}, nil
}
func GiaLUTPackingFromPacked(mapping *GiaLUTMapping, data []int, maxEntries int) (*GiaLUTPacking, error) {
	if mapping == nil || mapping.objects < 1 || len(data) == 0 {
		return nil, ErrDimension
	}
	limit, err := packingLimit(maxEntries)
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, ErrLimit
	}
	if data[0] < 0 || data[0] > (len(data)-1)/2 {
		return nil, ErrDimension
	}
	position := 1
	for group := 0; group < data[0]; group++ {
		if position >= len(data) {
			return nil, ErrDimension
		}
		count := data[position]
		position++
		if count < 1 || count > len(data)-position {
			return nil, ErrDimension
		}
		for _, id := range data[position : position+count] {
			if id < 0 || id >= mapping.objects || mapping.packed[id] == 0 {
				return nil, ErrDimension
			}
		}
		position += count
	}
	if position != len(data) {
		return nil, ErrDimension
	}
	return &GiaLUTPacking{mapping: mapping, packed: slices.Clone(data)}, nil
}
func (p *GiaLUTPacking) Packed() []int {
	if p == nil {
		return nil
	}
	return slices.Clone(p.packed)
}
func (p *GiaLUTPacking) Groups() [][]GiaID {
	if p == nil || len(p.packed) == 0 {
		return nil
	}
	groups := make([][]GiaID, 0, p.packed[0])
	for position := 1; position < len(p.packed); {
		count := p.packed[position]
		position++
		group := make([]GiaID, count)
		for i := range group {
			group[i] = GiaID(p.packed[position+i])
		}
		position += count
		groups = append(groups, group)
	}
	return groups
}
func (p *GiaLUTPacking) Statistics() GiaLUTPackingStatistics {
	r := GiaLUTPackingStatistics{MaxSize: -1, BySize: make([]int, 4)}
	if p == nil {
		return r
	}
	for position := 1; position < len(p.packed); {
		count := p.packed[position]
		position += count + 1
		if len(r.BySize) <= count {
			r.BySize = append(r.BySize, make([]int, count-len(r.BySize)+1)...)
		}
		r.Groups++
		r.LUTs += count
		r.BySize[count]++
		r.MaxSize = max(r.MaxSize, count)
	}
	return r
}

// Transfer preserves groups and slot order while discarding correspondence
// phases, as in C. Every packed source and target ID must have a LUT record.
// Dropped packed LUTs reject instead of indexing C's invalid sentinel. Target
// aliases and repeated members are retained. Unreferenced correspondence
// entries may be dropped. Bounds/cancellation expose no partially built pack.
func (p *GiaLUTPacking) Transfer(target *GiaLUTMapping, correspondence []GiaLit, options GiaLUTTransferOptions) (*GiaLUTPacking, error) {
	if p == nil || p.mapping == nil || len(p.packed) == 0 || target == nil || target.objects < 1 || len(correspondence) != p.mapping.objects {
		return nil, ErrDimension
	}
	limit, err := packingLimit(options.MaxEntries)
	if err != nil {
		return nil, err
	}
	if len(p.packed) > limit {
		return nil, ErrLimit
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	data := slices.Clone(p.packed)
	for position := 1; position < len(data); {
		if stopped() {
			return nil, ErrCancelled
		}
		count := data[position]
		position++
		for i := 0; i < count; i++ {
			if stopped() {
				return nil, ErrCancelled
			}
			lit := correspondence[data[position+i]]
			if lit == InvalidGiaLit || !lit.Valid() || uint64(lit.ID()) >= uint64(target.objects) {
				return nil, ErrInvalidLiteral
			}
			if target.packed[lit.ID()] == 0 {
				return nil, ErrDimension
			}
			data[position+i] = int(lit.ID())
		}
		position += count
	}
	if stopped() {
		return nil, ErrCancelled
	}
	return &GiaLUTPacking{mapping: target, packed: data}, nil
}
