package abc

// Port of src/aig/gia/gia.h and giaMan.c: compact objects and construction.

import (
	"fmt"
	"math"
	"slices"
)

type GiaID uint32
type GiaLit uint32

const (
	GiaNone       uint32 = 0x1fffffff
	MaxGiaID             = GiaID(GiaNone)
	InvalidGiaID         = GiaID(math.MaxUint32)
	InvalidGiaLit        = GiaLit(math.MaxUint32)
	GiaConst0            = GiaLit(0)
	GiaConst1            = GiaLit(1)
	giaComplement        = uint32(1) << 29
	giaTerm              = uint32(1) << 31
)

func MakeGiaLit(id GiaID, complement bool) (GiaLit, error) {
	if id > MaxGiaID {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	l := GiaLit(id) << 1
	if complement {
		l |= 1
	}
	return l, nil
}
func (l GiaLit) Valid() bool        { return l <= GiaLit(GiaNone<<1|1) }
func (l GiaLit) ID() GiaID          { return GiaID(l >> 1) }
func (l GiaLit) Complemented() bool { return l&1 != 0 }
func (l GiaLit) Not() GiaLit        { return l ^ 1 }
func (l GiaLit) NotIf(condition bool) GiaLit {
	if condition {
		return l ^ 1
	}
	return l
}
func (l GiaLit) Regular() GiaLit { return l &^ 1 }

// GiaObject retains Gia_Obj_t's three 32-bit words without pointers or Go
// padding. Fanins are backward differences; Diff1 is a terminal ordinal on
// CIs/COs. All returned objects are snapshots, never mutable arena pointers.
type GiaObject struct{ word0, word1, value uint32 }

func (o GiaObject) Diff0() uint32     { return o.word0 & GiaNone }
func (o GiaObject) Diff1() uint32     { return o.word1 & GiaNone }
func (o GiaObject) Complement0() bool { return o.word0&giaComplement != 0 }
func (o GiaObject) Complement1() bool { return o.word1&giaComplement != 0 }
func (o GiaObject) IsTerminal() bool  { return o.word0&giaTerm != 0 }
func (o GiaObject) IsCI() bool        { return o.IsTerminal() && o.Diff0() == GiaNone }
func (o GiaObject) IsCO() bool        { return o.IsTerminal() && o.Diff0() != GiaNone }
func (o GiaObject) IsConstant0() bool { return o.Diff0() == GiaNone && o.Diff1() == GiaNone }

// IsAnd follows Gia_ObjIsAnd: buffers, native XORs and native muxes also
// satisfy this predicate. IsBuffer/IsXor and Manager.MuxControl refine it.
func (o GiaObject) IsAnd() bool      { return !o.IsTerminal() && o.Diff0() != GiaNone }
func (o GiaObject) IsBuffer() bool   { return o.IsAnd() && o.Diff0() == o.Diff1() }
func (o GiaObject) IsXor() bool      { return o.IsAnd() && o.Diff0() < o.Diff1() }
func (o GiaObject) Phase() bool      { return o.word1&giaTerm != 0 }
func (o GiaObject) Mark0() bool      { return o.word0&(uint32(1)<<30) != 0 }
func (o GiaObject) Mark1() bool      { return o.word1&(uint32(1)<<30) != 0 }
func (o GiaObject) Value() uint32    { return o.value }
func (o GiaObject) Words() [3]uint32 { return [3]uint32{o.word0, o.word1, o.value} }
func (o GiaObject) TerminalIndex() (int, error) {
	if !o.IsTerminal() {
		return 0, ErrDimension
	}
	return int(o.Diff1()), nil
}
func (o GiaObject) Fanin0(id GiaID) (GiaLit, error) {
	if (!o.IsAnd() && !o.IsCO()) || o.Diff0() == 0 || o.Diff0() == GiaNone || uint32(id) < o.Diff0() {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	return MakeGiaLit(id-GiaID(o.Diff0()), o.Complement0())
}
func (o GiaObject) Fanin1(id GiaID) (GiaLit, error) {
	if !o.IsAnd() || o.Diff1() == 0 || o.Diff1() == GiaNone || uint32(id) < o.Diff1() {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	return MakeGiaLit(id-GiaID(o.Diff1()), o.Complement1())
}

type giaHashKey struct {
	a, b, c GiaLit
	kind    uint8
}

// GiaManager owns compact topological objects and optional hashing/mux data.
// It is single-owner and must not be copied. Graph IDs, hash insertion and
// iteration come from slices; the lookup map is never enumerated.
type GiaManager struct {
	objectLimit             int
	Name, Spec              string
	objects                 []GiaObject
	cis, cos                []GiaID
	registers, constraints  int
	xors, buffers, muxCount int
	muxes                   []GiaLit // nil until the first native mux; zero means no mux
	hash                    map[giaHashKey]GiaID
}

func NewGiaManager() *GiaManager {
	m, _ := NewGiaManagerWithLimit(int(MaxGiaID) + 1)
	return m
}

// NewGiaManagerWithLimit bounds the complete object arena, including constant
// zero and terminals. Failed appends return ErrLimit without terminating.
func NewGiaManagerWithLimit(maxObjects int) (*GiaManager, error) {
	if maxObjects < 1 {
		return nil, ErrDimension
	}
	if uint64(maxObjects) > uint64(MaxGiaID)+1 {
		return nil, ErrLimit
	}
	return &GiaManager{objectLimit: maxObjects, objects: []GiaObject{{word0: GiaNone, word1: GiaNone}}}, nil
}

func (m *GiaManager) ObjectCount() int        { return len(m.objects) }
func (m *GiaManager) InputCount() int         { return len(m.cis) }
func (m *GiaManager) OutputCount() int        { return len(m.cos) }
func (m *GiaManager) PrimaryInputCount() int  { return len(m.cis) - m.registers }
func (m *GiaManager) PrimaryOutputCount() int { return len(m.cos) - m.registers }
func (m *GiaManager) RegisterCount() int      { return m.registers }
func (m *GiaManager) ConstraintCount() int    { return m.constraints }

// AndCount follows Gia_ManAndNum, including buffers and native XOR/mux nodes.
func (m *GiaManager) AndCount() int {
	if len(m.objects) == 0 {
		return 0
	}
	return len(m.objects) - len(m.cis) - len(m.cos) - 1
}
func (m *GiaManager) XorCount() int    { return m.xors }
func (m *GiaManager) BufferCount() int { return m.buffers }
func (m *GiaManager) MuxCount() int    { return m.muxCount }
func (m *GiaManager) Inputs() []GiaID  { return slices.Clone(m.cis) }
func (m *GiaManager) Outputs() []GiaID { return slices.Clone(m.cos) }
func (m *GiaManager) Hashing() bool    { return m.hash != nil }
func (m *GiaManager) Object(id GiaID) (GiaObject, error) {
	if uint64(id) >= uint64(len(m.objects)) {
		return GiaObject{}, ErrInvalidLiteral
	}
	return m.objects[id], nil
}
func (m *GiaManager) MuxControl(id GiaID) (GiaLit, bool) {
	if uint64(id) >= uint64(len(m.muxes)) || m.muxes[id] == 0 {
		return InvalidGiaLit, false
	}
	return m.muxes[id], true
}
func (m *GiaManager) validSignal(l GiaLit) bool {
	return l.Valid() && uint64(l.ID()) < uint64(len(m.objects)) && !m.objects[l.ID()].IsCO()
}
func (m *GiaManager) room() error {
	if len(m.objects) == 0 {
		return fmt.Errorf("%w: uninitialized GIA manager", ErrInvariant)
	}
	if uint64(len(m.objects)) > uint64(MaxGiaID) || len(m.objects) >= m.objectLimit {
		return ErrLimit
	}
	return nil
}
func (m *GiaManager) push(obj GiaObject, control GiaLit) GiaLit {
	id := GiaID(len(m.objects))
	m.objects = append(m.objects, obj)
	if m.muxes != nil {
		m.muxes = append(m.muxes, control)
	}
	return GiaLit(id) << 1
}
func (m *GiaManager) AppendCI() (GiaLit, error) {
	if err := m.room(); err != nil {
		return InvalidGiaLit, err
	}
	l := m.push(GiaObject{word0: GiaNone | giaTerm, word1: uint32(len(m.cis))}, 0)
	m.cis = append(m.cis, l.ID())
	return l, nil
}
func giaDifference(id GiaID, l GiaLit) (uint32, error) {
	if id <= l.ID() || uint32(id-l.ID()) >= GiaNone {
		return 0, ErrLimit
	}
	value := uint32(id - l.ID())
	if l.Complemented() {
		value |= giaComplement
	}
	return value, nil
}
func (m *GiaManager) AppendCO(driver GiaLit) (GiaLit, error) {
	if !m.validSignal(driver) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	if err := m.room(); err != nil {
		return InvalidGiaLit, err
	}
	first, err := giaDifference(GiaID(len(m.objects)), driver)
	if err != nil {
		return InvalidGiaLit, err
	}
	l := m.push(GiaObject{word0: first | giaTerm, word1: uint32(len(m.cos))}, 0)
	m.cos = append(m.cos, l.ID())
	return l, nil
}

// AppendAnd is the raw Gia_ManAppendAnd operation: no constant simplification
// or duplicate elimination. Distinct regular fanins are required. If hashing
// is active, its lookup retains the first node with a duplicate structure.
func (m *GiaManager) AppendAnd(a, b GiaLit) (GiaLit, error)     { return m.appendGate(a, b, 0, 0) }
func (m *GiaManager) AppendXorReal(a, b GiaLit) (GiaLit, error) { return m.appendGate(a, b, 0, 1) }
func (m *GiaManager) AppendBuffer(driver GiaLit) (GiaLit, error) {
	return m.appendGate(driver, driver, 0, 3)
}
func (m *GiaManager) AppendMuxReal(control, whenTrue, whenFalse GiaLit) (GiaLit, error) {
	return m.appendGate(whenFalse, whenTrue, control, 2)
}
func (m *GiaManager) appendGate(a, b, control GiaLit, kind uint8) (GiaLit, error) {
	if !m.validSignal(a) || !m.validSignal(b) || (kind == 2 && !m.validSignal(control)) {
		return InvalidGiaLit, ErrInvalidLiteral
	}
	if kind != 3 && a.ID() == b.ID() {
		return InvalidGiaLit, ErrDimension
	}
	if kind == 2 && (control.ID() == 0 || control.ID() == a.ID() || control.ID() == b.ID()) {
		return InvalidGiaLit, ErrDimension
	}
	if err := m.room(); err != nil {
		return InvalidGiaLit, err
	}
	if kind == 1 {
		if a.ID() < b.ID() {
			a, b = b, a
		}
	} else if a > b {
		a, b = b, a
		if kind == 2 {
			control = control.Not()
		}
	}
	id := GiaID(len(m.objects))
	first, err := giaDifference(id, a)
	if err != nil {
		return InvalidGiaLit, err
	}
	second, err := giaDifference(id, b)
	if err != nil {
		return InvalidGiaLit, err
	}
	if kind == 2 && m.muxes == nil {
		m.muxes = make([]GiaLit, len(m.objects))
	}
	l := m.push(GiaObject{word0: first, word1: second}, control)
	switch kind {
	case 1:
		m.xors++
	case 2:
		m.muxCount++
	case 3:
		m.buffers++
	}
	if m.hash != nil {
		key := m.objectKey(l.ID())
		if _, exists := m.hash[key]; !exists {
			m.hash[key] = l.ID()
		}
	}
	return l, nil
}
func (m *GiaManager) SetRegisterCount(registers int) error {
	if registers < 0 || registers > len(m.cis) || registers > len(m.cos) || m.constraints > len(m.cos)-registers {
		return ErrDimension
	}
	m.registers = registers
	return nil
}
func (m *GiaManager) SetConstraintCount(constraints int) error {
	if constraints < 0 || constraints > m.PrimaryOutputCount() {
		return ErrDimension
	}
	m.constraints = constraints
	return nil
}

// PatchCODriver preserves the existing CO's ID and ordinal. Its driver must
// precede the CO; accepting a later object would violate compact differences.
func (m *GiaManager) PatchCODriver(output int, driver GiaLit) error {
	if output < 0 || output >= len(m.cos) {
		return ErrDimension
	}
	if !m.validSignal(driver) {
		return ErrInvalidLiteral
	}
	id := m.cos[output]
	if driver.ID() >= id {
		return ErrInvalidLiteral
	}
	first, err := giaDifference(id, driver)
	if err != nil {
		return err
	}
	m.objects[id].word0 = (m.objects[id].word0 &^ (GiaNone | giaComplement)) | first
	return nil
}
func (m *GiaManager) FlipCODriver(output int) error {
	if output < 0 || output >= len(m.cos) {
		return ErrDimension
	}
	m.objects[m.cos[output]].word0 ^= giaComplement
	return nil
}
func (m *GiaManager) RegisterInputForOutput(id GiaID) (GiaID, error) {
	obj, err := m.Object(id)
	if err != nil {
		return InvalidGiaID, err
	}
	if !obj.IsCI() || int(obj.Diff1()) < m.PrimaryInputCount() {
		return InvalidGiaID, ErrDimension
	}
	return m.cos[m.PrimaryOutputCount()+int(obj.Diff1())-m.PrimaryInputCount()], nil
}
func (m *GiaManager) RegisterOutputForInput(id GiaID) (GiaID, error) {
	obj, err := m.Object(id)
	if err != nil {
		return InvalidGiaID, err
	}
	if !obj.IsCO() || int(obj.Diff1()) < m.PrimaryOutputCount() {
		return InvalidGiaID, ErrDimension
	}
	return m.cis[m.PrimaryInputCount()+int(obj.Diff1())-m.PrimaryOutputCount()], nil
}

// Clone preserves compact IDs and independently copies all owned storage.
func (m *GiaManager) Clone() *GiaManager {
	d := *m
	d.objects = slices.Clone(m.objects)
	d.cis = slices.Clone(m.cis)
	d.cos = slices.Clone(m.cos)
	d.muxes = slices.Clone(m.muxes)
	if m.hash != nil {
		d.hash = nil
		_ = d.HashStart()
	}
	return &d
}
func (m *GiaManager) objectKey(id GiaID) giaHashKey {
	o := m.objects[id]
	a, _ := o.Fanin0(id)
	b, _ := o.Fanin1(id)
	key := giaHashKey{a: a, b: b}
	if o.IsBuffer() {
		key.kind = 3
	} else if o.IsXor() {
		key.kind = 1
	} else if c, yes := m.MuxControl(id); yes {
		key.kind = 2
		key.c = c
	}
	return key
}
