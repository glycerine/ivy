package abc

// Terminal-preserving DFS copies from src/aig/gia/giaDup.c:
// Gia_ManDupDfs_rec/Gia_ManDupDfs and Gia_ManDupDfsOnePo.

// DuplicateDFS retains all CIs/COs in original order, copying only reachable
// gates in output-root DFS postorder. Raw extended node kinds, register and
// constraint roles/name/spec survive. The owned source-object map marks
// unreachable gates invalid and maps source COs to target CO object literals.
// Source marks/Value/hash state are not mutated. Counterexample/choice metadata
// outside GiaManager's graph representation is not copied by this operation.
func (m *GiaManager) DuplicateDFS(options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	return m.duplicateSelectedDFS(-1, options)
}

// SelectPropertyWithState keeps one source PO and ALL original register inputs,
// even state unrelated to that property. This differs from sequential-closure
// cone extraction. CI/register/reset order stays unchanged, and the chosen PO
// becomes ordinary (source constraint roles are cleared, matching C). A source
// map exposes retained/omitted objects. Negative property indices are invalid.
func (m *GiaManager) SelectPropertyWithState(index int, options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if index < 0 {
		return nil, nil, ErrDimension
	}
	return m.duplicateSelectedDFS(index, options)
}
func (m *GiaManager) duplicateSelectedDFS(property int, options GiaCofactorOptions) (*GiaManager, []GiaLit, error) {
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, nil, ErrDimension
	}
	if e := m.Check(); e != nil {
		return nil, nil, e
	}
	if property >= m.PrimaryOutputCount() {
		return nil, nil, ErrDimension
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return nil, nil, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return nil, nil, ErrCancelled
	}
	roots := m.cos
	if property >= 0 {
		roots = make([]GiaID, 0, 1+m.registers)
		roots = append(roots, m.cos[property])
		roots = append(roots, m.cos[m.PrimaryOutputCount():]...)
	}
	nodes, err := m.CollectTransitiveFanins(roots, limit)
	if err != nil {
		return nil, nil, err
	}
	g, mapping, err := m.DuplicateFromLists(m.cis, nodes, roots, m.registers, options)
	if err != nil {
		return nil, nil, err
	}
	if property < 0 {
		g.constraints = m.constraints
	}
	if err = g.Check(); err != nil {
		return nil, nil, err
	}
	return g, mapping, nil
}
