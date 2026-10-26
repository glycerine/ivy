package abc

// Sequential cone extraction from src/aig/gia/giaDup.c:Gia_ManDupCones.

type GiaConeExtraction struct {
	Graph *GiaManager
	// Source indices for each retained PI/register, in new terminal order.
	// Use Registers to project latch initialization; graph managers do not store it.
	PrimaryInputs, Registers []int
}

// ExtractOutputCones copies selected POs and their transitive register-input
// cones. Optional PI trimming keeps reachable PIs in original order. Registers
// follow first-discovery order, matching C's dynamic root queue. Constraint
// outputs must remain a suffix. Owned index maps expose terminal reordering;
// source graph/scratch remain unchanged. Extended node kinds are preserved.
func (m *GiaManager) ExtractOutputCones(outputs []int, trimInputs bool, options GiaCofactorOptions) (GiaConeExtraction, error) {
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
	if m.ObjectCount() > options.MaxMappingEntries || len(outputs) > options.MaxMappingEntries {
		return result, ErrLimit
	}
	constraints := 0
	inConstraints := false
	for _, output := range outputs {
		if output < 0 || output >= m.PrimaryOutputCount() {
			return result, ErrDimension
		}
		if output >= m.PrimaryOutputCount()-m.constraints {
			inConstraints = true
			constraints++
		} else if inConstraints {
			return result, ErrDimension
		}
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return result, ErrCancelled
	}
	order, err := m.CollectSequential(outputs, options.MaxMappingEntries)
	if err != nil {
		return result, err
	}
	if stopped() {
		return result, ErrCancelled
	}
	usedPi := make([]bool, m.PrimaryInputCount())
	nodes := make([]GiaID, 0)
	for _, id := range order {
		obj := m.objects[id]
		if obj.IsAnd() {
			nodes = append(nodes, id)
		}
		if obj.IsCI() {
			index := int(obj.Diff1())
			if index < m.PrimaryInputCount() {
				usedPi[index] = true
			} else {
				result.Registers = append(result.Registers, index-m.PrimaryInputCount())
			}
		}
	}
	for index, used := range usedPi {
		if used || !trimInputs {
			result.PrimaryInputs = append(result.PrimaryInputs, index)
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
	for _, index := range result.PrimaryInputs {
		if stopped() {
			return GiaConeExtraction{}, ErrCancelled
		}
		id := m.cis[index]
		mapping[id], err = target.AppendCI()
		if err != nil {
			return GiaConeExtraction{}, err
		}
	}
	for _, index := range result.Registers {
		if stopped() {
			return GiaConeExtraction{}, ErrCancelled
		}
		id := m.cis[m.PrimaryInputCount()+index]
		mapping[id], err = target.AppendCI()
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
	roots := make([]GiaID, 0, len(outputs)+len(result.Registers))
	for _, output := range outputs {
		roots = append(roots, m.cos[output])
	}
	for _, reg := range result.Registers {
		roots = append(roots, m.cos[m.PrimaryOutputCount()+reg])
	}
	for _, id := range roots {
		if stopped() {
			return GiaConeExtraction{}, ErrCancelled
		}
		driver, err := m.objects[id].Fanin0(id)
		if err != nil {
			return GiaConeExtraction{}, err
		}
		copied, err := mappedGiaSignal(mapping, driver)
		if err != nil {
			return GiaConeExtraction{}, err
		}
		if _, err = target.AppendCO(copied); err != nil {
			return GiaConeExtraction{}, err
		}
	}
	if err = target.SetRegisterCount(len(result.Registers)); err != nil {
		return GiaConeExtraction{}, err
	}
	if err = target.SetConstraintCount(constraints); err != nil {
		return GiaConeExtraction{}, err
	}
	if err = target.Check(); err != nil {
		return GiaConeExtraction{}, err
	}
	result.Graph = target
	return result, nil
}
