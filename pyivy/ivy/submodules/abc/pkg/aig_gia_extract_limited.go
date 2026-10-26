package abc

// Bounded gate cones from src/aig/gia/giaDup.c:
// Gia_ManDupAndConesLimit and Gia_ManDupAndConesLimit2.

type GiaCutExtraction struct {
	Graph *GiaManager
	// Source object IDs supplying each new CI, in first-encounter order.
	Inputs []GiaID
}

// ExtractLevelLimitedCones stops at gates whose forward level is below minimum
// and at all CIs. Constants remain constants. The result is combinational even
// for sequential source graphs; register outputs are boundary inputs.
func (m *GiaManager) ExtractLevelLimitedCones(roots []GiaID, minimum uint32, options GiaCofactorOptions) (GiaCutExtraction, error) {
	return m.extractLimitedCones(roots, minimum, 0, false, options)
}

// ExtractDepthLimitedCones expands at most depth fanin steps on first visits.
// As in C, memoization is by object ID: a shared node keeps its first expansion,
// so root/DFS order can affect the cut and later paths can reuse deeper logic.
func (m *GiaManager) ExtractDepthLimitedCones(roots []GiaID, depth int, options GiaCofactorOptions) (GiaCutExtraction, error) {
	if depth < 0 {
		return GiaCutExtraction{}, ErrDimension
	}
	return m.extractLimitedCones(roots, 0, depth, true, options)
}
func (m *GiaManager) extractLimitedCones(roots []GiaID, minimum uint32, depth int, byDepth bool, options GiaCofactorOptions) (GiaCutExtraction, error) {
	var result GiaCutExtraction
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
	var levels []uint32
	var err error
	if !byDepth {
		levels, _, err = m.Levels(options.MaxMappingEntries)
		if err != nil {
			return result, err
		}
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return result, err
	}
	target.Name, target.Spec = m.Name, m.Spec
	mapping := make([]GiaLit, m.ObjectCount())
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	type visit struct {
		id    GiaID
		depth int
		exit  bool
	}
	stack := make([]visit, 0)
	for _, root := range roots {
		stack = append(stack, visit{id: root, depth: depth})
		for len(stack) > 0 {
			if stopped() {
				return GiaCutExtraction{}, ErrCancelled
			}
			item := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if mapping[item.id].Valid() {
				continue
			}
			if item.exit {
				mapping[item.id], err = appendGiaCopiedGate(m, target, item.id, mapping)
				if err != nil {
					return GiaCutExtraction{}, err
				}
				continue
			}
			boundary := !m.objects[item.id].IsAnd()
			if byDepth {
				boundary = boundary || item.depth == 0
			} else {
				boundary = boundary || levels[item.id] < minimum
			}
			if boundary {
				mapping[item.id], err = target.AppendCI()
				if err != nil {
					return GiaCutExtraction{}, err
				}
				result.Inputs = append(result.Inputs, item.id)
				continue
			}
			children, n, err := m.crossCutFanins(item.id)
			if err != nil {
				return GiaCutExtraction{}, err
			}
			nextDepth := item.depth
			if byDepth {
				nextDepth--
			}
			stack = append(stack, visit{id: item.id, depth: item.depth, exit: true})
			for i := n - 1; i >= 0; i-- {
				stack = append(stack, visit{id: children[i], depth: nextDepth})
			}
		}
	}
	for _, root := range roots {
		if stopped() {
			return GiaCutExtraction{}, ErrCancelled
		}
		if _, err = target.AppendCO(mapping[root]); err != nil {
			return GiaCutExtraction{}, err
		}
	}
	if err = target.Check(); err != nil {
		return GiaCutExtraction{}, err
	}
	result.Graph = target
	return result, nil
}
