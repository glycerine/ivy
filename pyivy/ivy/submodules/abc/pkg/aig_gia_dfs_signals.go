package abc

// Signal-root DFS copying from src/aig/gia/giaDup.c:
// Gia_ManDupDfs3_rec/DfsNode and Gia_ManDupDfsLitArray (without repr metadata).

// GiaEncounterCone records source CI ordinals in first-encounter target order,
// including original register outputs. The extracted graph is combinational.
type GiaEncounterCone struct {
	Graph  *GiaManager
	Inputs []int
}

// ExtractEncounterOrderedCone raw-copies a phased signal's reachable gates and
// CIs in DFS encounter/postorder. Unlike CI-trimmed original-order extraction,
// new CIs appear when first visited, possibly between gate objects. Constants
// and CI roots extend C's AND-only entry point. Native kinds/name/spec survive;
// constraint/register roles are not inherited. Source scratch stays untouched.
func (m *GiaManager) ExtractEncounterOrderedCone(root GiaLit, options GiaCofactorOptions) (GiaEncounterCone, error) {
	fail := func(e error) (GiaEncounterCone, error) { return GiaEncounterCone{}, e }
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
	if !m.validSignal(root) {
		return fail(ErrInvalidLiteral)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	nodes, err := m.CollectNodesAndInputs([]GiaID{root.ID()}, options.MaxMappingEntries)
	if err != nil {
		return fail(err)
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
	inputs := make([]int, 0)
	for _, id := range nodes {
		if stopped() {
			return fail(ErrCancelled)
		}
		obj := m.objects[id]
		if obj.IsCI() {
			index, e := obj.TerminalIndex()
			if e != nil {
				return fail(e)
			}
			inputs = append(inputs, index)
			mapping[id], err = g.AppendCI()
		} else {
			mapping[id], err = appendGiaCopiedGate(m, g, id, mapping)
		}
		if err != nil {
			return fail(err)
		}
	}
	driver, err := mappedGiaSignal(mapping, root)
	if err != nil {
		return fail(err)
	}
	if _, err = g.AppendCO(driver); err != nil {
		return fail(err)
	}
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaEncounterCone{g, inputs}, nil
}

// ExtractSignalsRetainingInputs creates phased signal outputs in caller order,
// retaining every CI in source order. Raw gates are copied once in DFS order
// as each output is visited. Duplicated outputs and constants are allowed.
// Register outputs become ordinary CIs; no constraint/register roles survive.
// Native gate kinds and name/spec are retained without source scratch mutation.
func (m *GiaManager) ExtractSignalsRetainingInputs(signals []GiaLit, options GiaCofactorOptions) (*GiaManager, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if e := m.Check(); e != nil {
		return nil, e
	}
	if len(m.objects) > options.MaxMappingEntries || len(signals) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	for _, l := range signals {
		if !m.validSignal(l) {
			return nil, ErrInvalidLiteral
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	g, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	g.Name, g.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, len(m.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, id := range m.cis {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for _, root := range signals {
		if stopped() {
			return nil, ErrCancelled
		}
		nodes, e := m.CollectTransitiveFanins([]GiaID{root.ID()}, options.MaxMappingEntries)
		if e != nil {
			return nil, e
		}
		for _, id := range nodes {
			if mapping[id].Valid() {
				continue
			}
			if stopped() {
				return nil, ErrCancelled
			}
			mapping[id], err = appendGiaCopiedGate(m, g, id, mapping)
			if err != nil {
				return nil, err
			}
		}
		driver, e := mappedGiaSignal(mapping, root)
		if e != nil {
			return nil, e
		}
		if _, err = g.AppendCO(driver); err != nil {
			return nil, err
		}
	}
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}
