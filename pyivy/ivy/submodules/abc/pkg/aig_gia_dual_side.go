package abc

// Side traversal/order from src/aig/gia/giaDup.c:Gia_ManCollectOneSide_rec,
// Gia_ManCollectOneSide and Gia_ManTransformDualOutput.

// CollectOutputSide returns unique AND-like nodes feeding even (side=0) or
// odd (side=1) POs, in DFS postorder. Shared gates appear once per call, and
// RI cones do not participate. Extended mux controls/buffers are traversed.
func (m *GiaManager) CollectOutputSide(side, maxObjects int) ([]GiaID, error) {
	if m == nil || side < 0 || side > 1 || maxObjects < 0 {
		return nil, ErrDimension
	}
	if maxObjects == 0 {
		maxObjects = 1 << 20
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	if len(m.objects) > maxObjects {
		return nil, ErrLimit
	}
	roots := make([]GiaID, 0, (m.PrimaryOutputCount()+1)/2)
	for index := side; index < m.PrimaryOutputCount(); index += 2 {
		roots = append(roots, m.cos[index])
	}
	return m.CollectTransitiveFanins(roots, maxObjects)
}

// GiaDualSideOrder reports whether adjacent output pairs were swapped. Node
// counts describe original side traversals, including gates shared by both.
type GiaDualSideOrder struct {
	Graph               *GiaManager
	Swapped             bool
	EvenNodes, OddNodes int
}

// OrderDualOutputSides hashes the smaller side first, swapping every adjacent
// PO pair only when the even side has more nodes. Ties retain order. Only
// combinational graphs with even PO count are accepted. All POs participate;
// constraint roles are cleared as in C. Shared logic is hashed, extended nodes
// lowered, dangling logic cleaned up, and all CIs/name/spec retained.
func (m *GiaManager) OrderDualOutputSides(options GiaCofactorOptions) (GiaDualSideOrder, error) {
	fail := func(e error) (GiaDualSideOrder, error) { return GiaDualSideOrder{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if m.registers != 0 || len(m.cos)%2 != 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	even, err := m.CollectOutputSide(0, options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	odd, err := m.CollectOutputSide(1, options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	swapped := len(even) > len(odd)
	first, second := even, odd
	if swapped {
		first, second = odd, even
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return fail(err)
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	if err = g.HashStart(); err != nil {
		return fail(err)
	}
	for _, id := range m.cis {
		if stopped() {
			return fail(ErrCancelled)
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return fail(err)
		}
	}
	for _, nodes := range [][]GiaID{first, second} {
		for _, id := range nodes {
			if stopped() {
				return fail(ErrCancelled)
			}
			if mapping[id].Valid() {
				continue
			}
			mapping[id], err = hashGiaCopiedGate(m, g, id, mapping)
			if err != nil {
				return fail(err)
			}
		}
	}
	for index := range m.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		source := index
		if swapped {
			source ^= 1
		}
		id := m.cos[source]
		driver, e := m.objects[id].Fanin0(id)
		if e != nil {
			return fail(e)
		}
		l, e := mappedGiaSignal(mapping, driver)
		if e != nil {
			return fail(e)
		}
		if _, err = g.AppendCO(l); err != nil {
			return fail(err)
		}
	}
	g.HashStop()
	if stopped() {
		return fail(ErrCancelled)
	}
	cleaned, _, err := g.Cleanup()
	if err != nil {
		return fail(err)
	}
	return GiaDualSideOrder{cleaned, swapped, len(even), len(odd)}, nil
}
