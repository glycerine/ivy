package mf

func (m *Mapping) ObjectCount() int       { return m.Network.ObjectCount() }
func (m *Mapping) InputLiterals() []int32 { return append([]int32(nil), m.Network.InputLiterals()...) }
func (m *Mapping) OutputLiterals() []int32 {
	return append([]int32(nil), m.Network.OutputLiterals()...)
}
func (m *Mapping) Gate(id int32) (a, b, c int32, kind int) { return m.Network.Gate(id) }
func (m *Mapping) RegisterCount() int {
	if sequential, ok := m.Network.(interface{ RegisterCount() int }); ok {
		return sequential.RegisterCount()
	}
	return 0
}

// LUT returns an independently owned selected cut and its actual truth table.
// Leaves are object IDs; truth uses little-endian64 words and repeated unused
// variables. Function complement is applied, leaving interned tables unchanged.
func (m *Mapping) LUT(id int32) (leaves []int32, truth []uint64, ok bool) {
	if id < 0 || int(id) >= len(m.Cuts) || int(id) >= len(m.Refs) || m.Refs[id] == 0 || len(m.Cuts[id]) == 0 {
		return nil, nil, false
	}
	cut := m.Cuts[id][0]
	leaves = append([]int32(nil), cut.Leaves...)
	truth = append([]uint64(nil), m.Truths[cut.Function>>1]...)
	if cut.Function&1 != 0 {
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	return leaves, truth, true
}

func (m *Mapping) NetworkName() string {
	if named, ok := m.Network.(interface{ NetworkName() string }); ok {
		return named.NetworkName()
	}
	return ""
}
func (m *Mapping) NetworkSpec() string {
	if specified, ok := m.Network.(interface{ NetworkSpec() string }); ok {
		return specified.NetworkSpec()
	}
	return ""
}
