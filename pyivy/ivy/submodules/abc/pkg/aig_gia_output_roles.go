package abc

// Explicit PO role conversion from src/aig/gia/giaDup.c:Gia_ManDupWithConstraints.

// GiaOutputRoleConversion maps target COs to source CO indices. Complemented
// records inverted constraint-violation signals; RIs remain uncomplemented.
type GiaOutputRoleConversion struct {
	Graph        *GiaManager
	Sources      []int
	Complemented []bool
}

// ConvertOutputRoles reinterprets every source PO: type 0 retains an ordinary
// property, type 1 complements a violation into a positive constraint, and any
// other type omits that PO, matching C. New order is stable type-0 POs, stable
// type-1 constraints, original RIs. Existing constraint roles are superseded.
// CI/register/reset order, all raw gates (including dangling logic), extended
// node kinds and name/spec survive. Returned projections make phase changes
// explicit; source scratch is never changed.
func (m *GiaManager) ConvertOutputRoles(types []int, options GiaCofactorOptions) (GiaOutputRoleConversion, error) {
	fail := func(e error) (GiaOutputRoleConversion, error) { return GiaOutputRoleConversion{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if len(types) != m.PrimaryOutputCount() {
		return fail(ErrDimension)
	}
	limit := options.MaxMappingEntries
	if limit == 0 {
		limit = 1 << 20
	}
	if len(m.objects) > limit {
		return fail(ErrLimit)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	indices := make([]int, 0, len(m.cos))
	phases := make([]bool, 0, len(m.cos))
	constraints := 0
	for role := 0; role < 2; role++ {
		for index, kind := range types {
			if kind != role {
				continue
			}
			indices = append(indices, index)
			phases = append(phases, role == 1)
			if role == 1 {
				constraints++
			}
		}
	}
	for index := m.PrimaryOutputCount(); index < len(m.cos); index++ {
		indices = append(indices, index)
		phases = append(phases, false)
	}
	cos := make([]GiaID, len(indices))
	for i, index := range indices {
		cos[i] = m.cos[index]
	}
	gates := make([]GiaID, 0, m.AndCount())
	for i, obj := range m.objects {
		if obj.IsAnd() {
			gates = append(gates, GiaID(i))
		}
	}
	g, _, err := m.DuplicateFromLists(m.cis, gates, cos, m.registers, options)
	if err != nil {
		return fail(err)
	}
	for index, complement := range phases {
		if !complement {
			continue
		}
		if stopped() {
			return fail(ErrCancelled)
		}
		id := g.cos[index]
		driver, e := g.objects[id].Fanin0(id)
		if e != nil {
			return fail(e)
		}
		if err = g.PatchCODriver(index, driver.Not()); err != nil {
			return fail(err)
		}
	}
	g.constraints = constraints
	if err = g.Check(); err != nil {
		return fail(err)
	}
	return GiaOutputRoleConversion{g, indices, phases}, nil
}
