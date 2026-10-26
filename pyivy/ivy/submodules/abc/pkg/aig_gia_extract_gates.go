package abc

// Combinational gate-root cones from src/aig/gia/giaDup.c:Gia_ManDupAndCones.

// ExtractGateCones makes one output per regular gate/CI/constant root, with
// optional PI trimming in original order. This C routine's register creation
// is disabled, so sequential graphs are rejected explicitly. Source constraint
// roles do not attach to arbitrary roots; the result has no constraints. Raw
// extended node kinds are preserved, with source PI indices returned.
func (m *GiaManager) ExtractGateCones(roots []GiaID, trimInputs bool, options GiaCofactorOptions) (GiaConeExtraction, error) {
	var result GiaConeExtraction
	if options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return result, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 1 << 20
	}
	if err := m.Check(); err != nil {
		return result, err
	}
	if m.registers != 0 {
		return result, ErrDimension
	}
	if m.ObjectCount() > options.MaxMappingEntries || len(roots) > options.MaxMappingEntries {
		return result, ErrLimit
	}
	for _, id := range roots {
		if uint64(id) >= uint64(m.ObjectCount()) || m.objects[id].IsCO() {
			return result, ErrInvalidLiteral
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return result, ErrCancelled
	}
	nodes, err := m.CollectConeAnds(roots, nil, options.MaxMappingEntries)
	if err != nil {
		return result, err
	}
	support, err := m.CollectInputs(roots, options.MaxMappingEntries)
	if err != nil {
		return result, err
	}
	if stopped() {
		return result, ErrCancelled
	}
	used := make([]bool, len(m.cis))
	for _, id := range support {
		used[m.objects[id].Diff1()] = true
	}
	for i, present := range used {
		if present || !trimInputs {
			result.PrimaryInputs = append(result.PrimaryInputs, i)
		}
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return GiaConeExtraction{}, err
	}
	target.Name, target.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, m.ObjectCount())
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	for _, i := range result.PrimaryInputs {
		if stopped() {
			return GiaConeExtraction{}, ErrCancelled
		}
		mapping[m.cis[i]], err = target.AppendCI()
		if err != nil {
			return GiaConeExtraction{}, err
		}
	}
	for _, id := range nodes {
		if stopped() {
			return GiaConeExtraction{}, ErrCancelled
		}
		mapping[id], err = appendGiaCopiedGate(m, target, id, mapping)
		if err != nil {
			return GiaConeExtraction{}, err
		}
	}
	for _, id := range roots {
		if stopped() {
			return GiaConeExtraction{}, ErrCancelled
		}
		if !mapping[id].Valid() {
			return GiaConeExtraction{}, ErrInvariant
		}
		if _, err = target.AppendCO(mapping[id]); err != nil {
			return GiaConeExtraction{}, err
		}
	}
	if err = target.Check(); err != nil {
		return GiaConeExtraction{}, err
	}
	result.Graph = target
	return result, nil
}
