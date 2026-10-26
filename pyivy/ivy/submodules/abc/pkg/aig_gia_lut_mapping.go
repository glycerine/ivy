package abc

// Packed mapping access from aig/gia/gia.h; LUT statistics/levels from
// aig/gia/giaIf.c. Boundary evaluation uses aig/gia/giaTruth.c.

import "slices"

type GiaLUTRecord struct {
	Root   GiaID
	Fanins []GiaID // ordered, may repeat; each precedes Root
	MuxID  int     // signed native footer; negative denotes a mux LUT
}

// GiaLUTMapping owns a native offset table followed by variable-length records.
// The layout is [fanin count, fanins..., signed mux ID] at each nonzero offset.
// Records may include constant object zero, which native LUT iteration skips.
// Storage is immutable; accessors return owned copies.
type GiaLUTMapping struct {
	objects int
	packed  []int
}

// NewGiaLUTMapping preserves record insertion order in packed storage. LUT
// iteration is always in ascending object order, as in C. Limits count all
// packed integers, including the object-indexed header; zero means one million.
func NewGiaLUTMapping(objects int, records []GiaLUTRecord, maxEntries int) (*GiaLUTMapping, error) {
	if objects < 1 || maxEntries < 0 {
		return nil, ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if objects > maxEntries || objects > int(^uint32(0)>>1) {
		return nil, ErrLimit
	}
	total := objects
	for _, record := range records {
		var err error
		total, err = checkedSize(total, 1, len(record.Fanins)+2, min(maxEntries, int(^uint32(0)>>1)))
		if err != nil {
			return nil, err
		}
	}
	data := make([]int, objects, total)
	for _, record := range records {
		if uint64(record.Root) >= uint64(objects) || data[record.Root] != 0 {
			return nil, ErrDimension
		}
		if record.MuxID <= -objects || record.MuxID >= objects {
			return nil, ErrDimension
		}
		if record.Root == 0 && len(record.Fanins) != 0 {
			return nil, ErrDimension
		}
		data[record.Root] = len(data)
		data = append(data, len(record.Fanins))
		for _, id := range record.Fanins {
			if id >= record.Root {
				return nil, ErrDimension
			}
			data = append(data, int(id))
		}
		data = append(data, record.MuxID)
	}
	return &GiaLUTMapping{objects: objects, packed: data}, nil
}

// GiaLUTMappingFromPacked validates offsets, sizes, fanins, footers and
// disjoint referenced records before copying. Unreferenced trailing records
// are permitted because native mapping construction can replace root records.
func GiaLUTMappingFromPacked(objects int, data []int, maxEntries int) (*GiaLUTMapping, error) {
	if objects < 1 || maxEntries < 0 || len(data) < objects {
		return nil, ErrDimension
	}
	if maxEntries == 0 {
		maxEntries = 1 << 20
	}
	if len(data) > maxEntries || len(data) > int(^uint32(0)>>1) {
		return nil, ErrLimit
	}
	for _, value := range data {
		if int64(value) < -2147483648 || int64(value) > 2147483647 {
			return nil, ErrDimension
		}
	}
	occupied := make([]bool, len(data))
	for root, offset := range data[:objects] {
		if offset == 0 {
			continue
		}
		if offset < objects || offset >= len(data) {
			return nil, ErrDimension
		}
		count := data[offset]
		if count < 0 || count > len(data)-offset-2 {
			return nil, ErrDimension
		}
		end := offset + count + 2
		for i := offset; i < end; i++ {
			if occupied[i] {
				return nil, ErrDimension
			}
			occupied[i] = true
		}
		if root == 0 && count != 0 {
			return nil, ErrDimension
		}
		for _, fanin := range data[offset+1 : end-1] {
			if fanin < 0 || fanin >= root {
				return nil, ErrDimension
			}
		}
		footer := data[end-1]
		if footer <= -objects || footer >= objects {
			return nil, ErrDimension
		}
	}
	return &GiaLUTMapping{objects: objects, packed: slices.Clone(data)}, nil
}
func (m *GiaLUTMapping) Packed() []int {
	if m == nil {
		return nil
	}
	return slices.Clone(m.packed)
}
func (m *GiaLUTMapping) Record(root GiaID) (GiaLUTRecord, bool) {
	if m == nil || uint64(root) >= uint64(m.objects) || m.packed[root] == 0 {
		return GiaLUTRecord{}, false
	}
	offset := m.packed[root]
	count := m.packed[offset]
	r := GiaLUTRecord{Root: root, Fanins: make([]GiaID, count), MuxID: m.packed[offset+count+1]}
	for i := range r.Fanins {
		r.Fanins[i] = GiaID(m.packed[offset+1+i])
	}
	return r, true
}
func (m *GiaLUTMapping) Roots() []GiaID {
	if m == nil {
		return nil
	}
	roots := make([]GiaID, 0)
	for id := 1; id < m.objects; id++ {
		if m.packed[id] != 0 {
			roots = append(roots, GiaID(id))
		}
	}
	return roots
}

type GiaLUTStatistics struct{ LUTs, Fanins, MaxSize int }

func (m *GiaLUTMapping) Statistics() GiaLUTStatistics {
	r := GiaLUTStatistics{MaxSize: -1}
	if m == nil {
		return r
	}
	for id := 1; id < m.objects; id++ {
		if offset := m.packed[id]; offset != 0 {
			count := m.packed[offset]
			r.LUTs++
			r.Fanins += count
			r.MaxSize = max(r.MaxSize, count)
		}
	}
	return r
}
func (mapping *GiaLUTMapping) checkGraph(m *GiaManager) error {
	if mapping == nil || m == nil || mapping.objects != len(m.objects) {
		return ErrDimension
	}
	if err := m.Check(); err != nil {
		return err
	}
	for _, id := range mapping.Roots() {
		if !m.objects[id].IsAnd() {
			return ErrDimension
		}
	}
	return nil
}

// Levels matches Gia_ManLutLevel, including zero levels for unmapped objects,
// CO driver levels and the maximum over COs rather than dangling LUTs.
func (mapping *GiaLUTMapping) Levels(m *GiaManager) ([]int, int, error) {
	if err := mapping.checkGraph(m); err != nil {
		return nil, 0, err
	}
	levels := make([]int, mapping.objects)
	for _, id := range mapping.Roots() {
		r, _ := mapping.Record(id)
		level := 0
		for _, fanin := range r.Fanins {
			level = max(level, levels[fanin])
		}
		levels[id] = level + 1
	}
	maximum := 0
	for _, id := range m.cos {
		lit, err := m.objects[id].Fanin0(id)
		if err != nil {
			return nil, 0, err
		}
		levels[id] = levels[lit.ID()]
		maximum = max(maximum, levels[id])
	}
	return levels, maximum, nil
}

// Truth6 evaluates the stored ordered LUT boundary, including native last-slot
// handling for repeated fanins. Mapping identity is its object-ID arena; graph
// edits after mapping construction require callers to rebuild the mapping.
func (mapping *GiaLUTMapping) Truth6(m *GiaManager, root GiaID, options GiaCutTruthOptions) (uint64, error) {
	if err := mapping.checkGraph(m); err != nil {
		return 0, err
	}
	record, ok := mapping.Record(root)
	if !ok {
		return 0, ErrDimension
	}
	return m.BoundaryTruth6(GiaLit(root)<<1, record.Fanins, options)
}
