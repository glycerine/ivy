package abc

// Explicit-support cone extraction/reinsertion from src/aig/gia/giaDup.c:
// Gia_ManDupConeSupp_rec/ConeSupp and Gia_ManDupConeBack_rec/Back/BackObjs.

// ExtractSignalWithInputs copies a phased source signal into a combinational
// graph with CIs in inputIndices order (source CI ordinals, including ROs).
// Indices must be unique and cover every reached CI; extra unused inputs stay.
// Missing support is an error, not a constant substitution. Constants and CI
// roots are supported safely. Raw extended gate kinds and name/spec survive;
// arbitrary root outputs inherit no constraint/register roles.
func (m *GiaManager) ExtractSignalWithInputs(root GiaLit, inputIndices []int, options GiaCofactorOptions) (*GiaManager, error) {
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
	if !m.validSignal(root) {
		return nil, ErrInvalidLiteral
	}
	if len(m.objects) > options.MaxMappingEntries {
		return nil, ErrLimit
	}
	if len(inputIndices) > len(m.cis) {
		return nil, ErrDimension
	}
	seen := make([]bool, len(m.cis))
	for _, index := range inputIndices {
		if index < 0 || index >= len(m.cis) || seen[index] {
			return nil, ErrDimension
		}
		seen[index] = true
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return nil, ErrCancelled
	}
	nodes, err := m.CollectTransitiveFanins([]GiaID{root.ID()}, options.MaxMappingEntries)
	if err != nil {
		return nil, err
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
	for _, index := range inputIndices {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[m.cis[index]], err = g.AppendCI()
		if err != nil {
			return nil, err
		}
	}
	for _, id := range nodes {
		if stopped() {
			return nil, ErrCancelled
		}
		mapping[id], err = appendGiaCopiedGate(m, g, id, mapping)
		if err != nil {
			return nil, err
		}
	}
	driver, err := mappedGiaSignal(mapping, root)
	if err != nil {
		return nil, err
	}
	if _, err = g.AppendCO(driver); err != nil {
		return nil, err
	}
	if err = g.Check(); err != nil {
		return nil, err
	}
	return g, nil
}

// WithInsertedCone clones the target and hashes a single-output combinational
// cone into it, binding each cone CI to a phased target signal. Existing target
// IDs, outputs, register/constraint roles and name/spec remain. The returned
// literal names the inserted function; no CO is appended/rewired. New logic can
// therefore be dangling until a caller uses it. Aliased/constant bindings are
// allowed. Both source graphs remain unchanged, including on failure.
func (m *GiaManager) WithInsertedCone(cone *GiaManager, bindings []GiaLit, options GiaCofactorOptions) (*GiaManager, GiaLit, error) {
	fail := func(e error) (*GiaManager, GiaLit, error) { return nil, InvalidGiaLit, e }
	if m == nil || cone == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	for _, source := range []*GiaManager{m, cone} {
		if e := source.Check(); e != nil {
			return fail(e)
		}
	}
	if cone.registers != 0 || cone.constraints != 0 || len(cone.cos) != 1 || len(bindings) != len(cone.cis) {
		return fail(ErrDimension)
	}
	if _, e := checkedSize(len(m.objects), 1, len(cone.objects), options.MaxMappingEntries); e != nil {
		return fail(e)
	}
	if len(m.objects) > options.MaxObjects {
		return fail(ErrLimit)
	}
	for _, l := range bindings {
		if !m.validSignal(l) {
			return fail(ErrInvalidLiteral)
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	nodes, err := cone.CollectTransitiveFanins(cone.cos, options.MaxMappingEntries)
	if err != nil {
		return fail(err)
	}
	g := m.Clone()
	g.objectLimit = options.MaxObjects
	hadHash := g.hash != nil
	if !hadHash {
		if err = g.HashStart(); err != nil {
			return fail(err)
		}
	}
	mapping := make([]GiaLit, len(cone.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for i, id := range cone.cis {
		mapping[id] = bindings[i]
	}
	for _, id := range nodes {
		if stopped() {
			return fail(ErrCancelled)
		}
		mapping[id], err = hashGiaCopiedGate(cone, g, id, mapping)
		if err != nil {
			return fail(err)
		}
	}
	co := cone.cos[0]
	driver, err := cone.objects[co].Fanin0(co)
	if err != nil {
		return fail(err)
	}
	result, err := mappedGiaSignal(mapping, driver)
	if err != nil {
		return fail(err)
	}
	if !hadHash {
		g.HashStop()
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return g, result, nil
}
