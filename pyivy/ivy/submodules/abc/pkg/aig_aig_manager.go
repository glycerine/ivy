package abc

// Derived from ABC src/aig/aig/{aig.h,aigMan.c,aigObj.c,aigCheck.c}.
// Original algorithms by Alan Mishchenko.

import (
	"fmt"
	"slices"
)

// AigType uses the Aig_Type_t values. The initial native representation
// supports constant, CI, CO, buffer and AND objects. XORs are lowered to ANDs.
type AigType uint8

const (
	AigNone AigType = iota
	AigConstant1
	AigCI
	AigCO
	AigBuffer
	AigAND
	AigXOR
	AigVoid
)

// AigObject is a value snapshot. Editing it cannot mutate its manager.
// Fanin1 is InvalidAigLit for unary objects; both fanins are invalid on CIs
// and constant one. RefCount counts graph edges, including CO edges.
type AigObject struct {
	Type           AigType
	Fanin0, Fanin1 AigLit
	RefCount       uint32
	Level          uint32
	Phase          bool
}

type aigHashKey struct{ a, b AigLit }

// AigManager is an append-only, structurally hashed AIG with stable object
// IDs, including tombstones for cleaned nodes. CI order is PIs then latch
// outputs; CO order is POs then latch inputs. A manager must not be copied.
// The zero value is not initialized; construct it with NewAigManager.
type AigManager struct {
	Name, Spec  string
	objects     []AigObject
	cis, cos    []AigID
	equivs      []AigID
	registers   int
	constraints int
	// This map is never enumerated. IDs and iteration order come from slices.
	hash map[aigHashKey]AigID
}

func NewAigManager() *AigManager {
	return &AigManager{objects: []AigObject{{Type: AigConstant1, Fanin0: InvalidAigLit, Fanin1: InvalidAigLit, Phase: true}}, hash: make(map[aigHashKey]AigID)}
}

// ObjectCount is the allocated ID span, including permanent tombstones.
// It corresponds to Aig_ManObjNumMax, not the number of live objects.
func (m *AigManager) ObjectCount() int        { return len(m.objects) }
func (m *AigManager) InputCount() int         { return len(m.cis) }
func (m *AigManager) OutputCount() int        { return len(m.cos) }
func (m *AigManager) RegisterCount() int      { return m.registers }
func (m *AigManager) ConstraintCount() int    { return m.constraints }
func (m *AigManager) PrimaryInputCount() int  { return len(m.cis) - m.registers }
func (m *AigManager) PrimaryOutputCount() int { return len(m.cos) - m.registers }
func (m *AigManager) AndCount() int           { return len(m.hash) }
func (m *AigManager) Inputs() []AigID         { return slices.Clone(m.cis) }
func (m *AigManager) Outputs() []AigID        { return slices.Clone(m.cos) }
func (m *AigManager) Object(id AigID) (AigObject, error) {
	if uint64(id) >= uint64(len(m.objects)) || m.objects[id].Type == AigNone {
		return AigObject{}, ErrInvalidLiteral
	}
	return m.objects[id], nil
}
func (m *AigManager) validSignal(l AigLit) bool {
	if !l.Valid() || uint64(l.ID()) >= uint64(len(m.objects)) {
		return false
	}
	switch m.objects[l.ID()].Type {
	case AigConstant1, AigCI, AigBuffer, AigAND:
		return true
	}
	return false
}
func (m *AigManager) phase(l AigLit) bool { return m.objects[l.ID()].Phase != l.Complemented() }
func (m *AigManager) room() error {
	if len(m.objects) == 0 || m.hash == nil {
		return fmt.Errorf("%w: uninitialized AIG manager", ErrInvariant)
	}
	if uint64(len(m.objects)) > uint64(MaxLiteralID) {
		return ErrLimit
	}
	return nil
}
func (m *AigManager) appendObject(obj AigObject) AigID {
	id := AigID(len(m.objects))
	m.objects = append(m.objects, obj)
	if m.equivs != nil {
		m.equivs = append(m.equivs, InvalidAigID)
	}
	return id
}

// CreateCI implements Aig_ObjCreateCi. Partition the final CI/CO lists with
// SetRegisterCount after constructing all terminals.
func (m *AigManager) CreateCI() (AigLit, error) {
	if err := m.room(); err != nil {
		return InvalidAigLit, err
	}
	id := m.appendObject(AigObject{Type: AigCI, Fanin0: InvalidAigLit, Fanin1: InvalidAigLit})
	m.cis = append(m.cis, id)
	return AigLit(id) << 1, nil
}

// IthVar implements Aig_IthVar, appending missing CIs. The caller must respect
// the PI/latch-output partition when constructing sequential graphs.
func (m *AigManager) IthVar(index int) (AigLit, error) {
	if index < 0 {
		return InvalidAigLit, ErrDimension
	}
	if uint64(index) >= uint64(MaxLiteralID) {
		return InvalidAigLit, ErrLimit
	}
	for len(m.cis) <= index {
		if _, err := m.CreateCI(); err != nil {
			return InvalidAigLit, err
		}
	}
	return AigLit(m.cis[index]) << 1, nil
}
func (m *AigManager) CreateCO(driver AigLit) (AigID, error) {
	if !m.validSignal(driver) {
		return InvalidAigID, ErrInvalidLiteral
	}
	if err := m.room(); err != nil {
		return InvalidAigID, err
	}
	if m.objects[driver.ID()].RefCount == ^uint32(0) {
		return InvalidAigID, ErrLimit
	}
	obj := AigObject{Type: AigCO, Fanin0: driver, Fanin1: InvalidAigLit, Level: m.objects[driver.ID()].Level, Phase: m.phase(driver)}
	id := m.appendObject(obj)
	m.objects[driver.ID()].RefCount++
	m.cos = append(m.cos, id)
	return id, nil
}
func (m *AigManager) CreateBuffer(driver AigLit) (AigLit, error) {
	if !m.validSignal(driver) {
		return InvalidAigLit, ErrInvalidLiteral
	}
	if err := m.room(); err != nil {
		return InvalidAigLit, err
	}
	if m.objects[driver.ID()].RefCount == ^uint32(0) {
		return InvalidAigLit, ErrLimit
	}
	id := m.appendObject(AigObject{Type: AigBuffer, Fanin0: driver, Fanin1: InvalidAigLit, Level: m.objects[driver.ID()].Level, Phase: m.phase(driver)})
	m.objects[driver.ID()].RefCount++
	return AigLit(id) << 1, nil
}
func (m *AigManager) SetRegisterCount(registers int) error {
	if registers < 0 || registers > len(m.cis) || registers > len(m.cos) || m.constraints > len(m.cos)-registers {
		return ErrDimension
	}
	m.registers = registers
	return nil
}

// Cleanup implements Aig_ManCleanup by deleting unreferenced logic in reverse
// topological order. Removed IDs become permanent tombstones; live IDs stay
// valid and are never remapped or reused. CIs, COs and constant one survive.
// This also removes dangling buffers, which the initial C sweep omits.
func (m *AigManager) Cleanup() int {
	removed := 0
	choiceMember := make([]bool, len(m.objects))
	for id, next := range m.equivs {
		if next != InvalidAigID {
			choiceMember[id] = true
			if uint64(next) < uint64(len(choiceMember)) {
				choiceMember[next] = true
			}
		}
	}
	for i := len(m.objects) - 1; i > 0; i-- {
		obj := m.objects[i]
		if (obj.Type != AigAND && obj.Type != AigBuffer) || obj.RefCount != 0 || choiceMember[i] {
			continue
		}
		if obj.Type == AigAND {
			delete(m.hash, aigHashKey{obj.Fanin0, obj.Fanin1})
			m.objects[obj.Fanin1.ID()].RefCount--
		}
		m.objects[obj.Fanin0.ID()].RefCount--
		m.objects[i] = AigObject{}
		removed++
	}
	return removed
}

// Clone returns independent graph and hash storage with exactly the same IDs.
func (m *AigManager) Clone() *AigManager {
	d := &AigManager{Name: m.Name, Spec: m.Spec, registers: m.registers, constraints: m.constraints, objects: slices.Clone(m.objects), cis: slices.Clone(m.cis), cos: slices.Clone(m.cos), equivs: slices.Clone(m.equivs), hash: make(map[aigHashKey]AigID, len(m.hash))}
	for i, obj := range d.objects {
		if obj.Type == AigAND {
			d.hash[aigHashKey{obj.Fanin0, obj.Fanin1}] = AigID(i)
		}
	}
	return d
}

// Check validates terminal ordering, topological fanins, levels, phases,
// reference counts and hash canonicality, returning an error on corruption.
func (m *AigManager) Check() error {
	fail := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s", ErrInvariant, fmt.Sprintf(format, args...))
	}
	if len(m.objects) == 0 || m.hash == nil {
		return fail("uninitialized AIG manager")
	}
	if m.registers < 0 || m.registers > len(m.cis) || m.registers > len(m.cos) || m.constraints < 0 || m.constraints > len(m.cos)-m.registers {
		return fail("register partition")
	}
	refs := make([]uint32, len(m.objects))
	ciIndex, coIndex, ands := 0, 0, 0
	for i, obj := range m.objects {
		if i == 0 {
			if obj.Type != AigConstant1 || !obj.Phase || obj.Level != 0 || obj.Fanin0 != InvalidAigLit || obj.Fanin1 != InvalidAigLit {
				return fail("constant one")
			}
			continue
		}
		if obj.Type == AigNone {
			continue
		}
		switch obj.Type {
		case AigCI:
			if obj.Fanin0 != InvalidAigLit || obj.Fanin1 != InvalidAigLit || obj.Phase || obj.Level != 0 {
				return fail("CI %d metadata", i)
			}
			if ciIndex >= len(m.cis) || m.cis[ciIndex] != AigID(i) {
				return fail("CI %d ordering", i)
			}
			ciIndex++
			continue
		case AigCO:
			if coIndex >= len(m.cos) || m.cos[coIndex] != AigID(i) {
				return fail("CO %d ordering", i)
			}
			coIndex++
		case AigBuffer, AigAND:
		default:
			return fail("unsupported object type %d at %d", obj.Type, i)
		}
		if !m.validSignal(obj.Fanin0) || int(obj.Fanin0.ID()) >= i {
			return fail("object %d first fanin", i)
		}
		refs[obj.Fanin0.ID()]++
		level := m.objects[obj.Fanin0.ID()].Level
		phase := m.phase(obj.Fanin0)
		if obj.Type == AigAND {
			if !m.validSignal(obj.Fanin1) || int(obj.Fanin1.ID()) >= i || obj.Fanin0.ID() >= obj.Fanin1.ID() || obj.Fanin0.ID() == 0 {
				return fail("AND %d second fanin/canonicality", i)
			}
			refs[obj.Fanin1.ID()]++
			level = 1 + max(level, m.objects[obj.Fanin1.ID()].Level)
			phase = phase && m.phase(obj.Fanin1)
			id, found := m.hash[aigHashKey{obj.Fanin0, obj.Fanin1}]
			if !found || id != AigID(i) {
				return fail("AND %d hash entry", i)
			}
			ands++
		} else if obj.Fanin1 != InvalidAigLit {
			return fail("unary object %d second fanin", i)
		}
		if obj.Level != level || obj.Phase != phase {
			return fail("object %d level/phase", i)
		}
	}
	if ciIndex != len(m.cis) || coIndex != len(m.cos) || ands != len(m.hash) {
		return fail("object counts")
	}
	if m.equivs != nil {
		if len(m.equivs) != len(m.objects) {
			return fail("choice array length")
		}
		incoming := make([]bool, len(m.objects))
		for id, next := range m.equivs {
			if next == InvalidAigID {
				continue
			}
			if m.objects[id].Type != AigAND || uint64(next) >= uint64(len(m.objects)) || m.objects[next].Type != AigAND || next == AigID(id) {
				return fail("choice link %d -> %d", id, next)
			}
			if incoming[next] {
				return fail("choice object %d has multiple predecessors", next)
			}
			incoming[next] = true
		}
		state := make([]uint8, len(m.objects))
		for start := range m.equivs {
			if state[start] != 0 {
				continue
			}
			var path []AigID
			for id := AigID(start); id != InvalidAigID; id = m.equivs[id] {
				if state[id] == 1 {
					return fail("choice cycle at %d", id)
				}
				if state[id] == 2 {
					break
				}
				state[id] = 1
				path = append(path, id)
			}
			for _, id := range path {
				state[id] = 2
			}
		}
	}
	for i, obj := range m.objects {
		if obj.RefCount != refs[i] {
			return fail("object %d references: got %d want %d", i, obj.RefCount, refs[i])
		}
	}
	return nil
}

// SetConstraintCount sets the number of final primary outputs representing
// constraints, matching Aig_Man_t.nConstrs. It does not invert their polarity.
func (m *AigManager) SetConstraintCount(constraints int) error {
	if constraints < 0 || constraints > m.PrimaryOutputCount() {
		return ErrDimension
	}
	m.constraints = constraints
	return nil
}
