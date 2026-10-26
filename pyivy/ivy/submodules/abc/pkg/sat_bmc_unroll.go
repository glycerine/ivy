package abc

// Port of src/sat/bmc/bmcUnroll.c: Unr_ManUnrollSimple. Input graphs are
// immutable here; explicit mapping storage replaces borrowed Value fields.

import "slices"

// GiaUnrollOptions bounds storage before allocation. Initial nil means zero
// initialization, matching Unr_ManUnrollSimple. Explicit one/unknown values
// extend that operation for AIGER models without silently changing their reset.
// Zero MaxObjects/MaxMappingEntries select 1M objects and 4M mapping entries.
type GiaUnrollOptions struct {
	Frames, MaxObjects, MaxMappingEntries int
	Initial                               []LatchInit
	Stop                                  func() bool
}

// GiaUnrolling is a combinational graph of all primary outputs at each frame.
// Unknown initial registers are CIs preceding the frame-major primary inputs.
// Output order is frame-major, including each frame's positive constraints.
// The graph's global constraint count is zero: constraints are interleaved and
// must be selected through ConstraintOutputs, never interpreted as properties.
type GiaUnrolling struct {
	graph                                               *GiaManager
	frames, sourceObjects, inputs, outputs, constraints int
	mapping                                             []GiaLit
	frameInputs                                         []GiaLit
	frameOutputs                                        []GiaID
	initial                                             []GiaLit
}

func (u *GiaUnrolling) Graph() *GiaManager        { return u.graph }
func (u *GiaUnrolling) Frames() int               { return u.frames }
func (u *GiaUnrolling) InitialLiterals() []GiaLit { return slices.Clone(u.initial) }

// ObjectLiteral maps a regular source object in a frame. COs map to driver
// signals, as in C's Value field, rather than to the unrolled CO objects.
// Logic discarded by final cleanup returns ErrInvalidLiteral.
func (u *GiaUnrolling) ObjectLiteral(frame int, id GiaID) (GiaLit, error) {
	if frame < 0 || frame >= u.frames || uint64(id) >= uint64(u.sourceObjects) {
		return InvalidGiaLit, ErrDimension
	}
	l := u.mapping[frame*u.sourceObjects+int(id)]
	if l == InvalidGiaLit {
		return l, ErrInvalidLiteral
	}
	return l, nil
}
func (u *GiaUnrolling) FrameInputs(frame int) ([]GiaLit, error) {
	if frame < 0 || frame >= u.frames {
		return nil, ErrDimension
	}
	return slices.Clone(u.frameInputs[frame*u.inputs : (frame+1)*u.inputs]), nil
}
func (u *GiaUnrolling) FrameOutputs(frame int) ([]GiaID, error) {
	if frame < 0 || frame >= u.frames {
		return nil, ErrDimension
	}
	return slices.Clone(u.frameOutputs[frame*u.outputs : (frame+1)*u.outputs]), nil
}
func (u *GiaUnrolling) PropertyOutputs(frame int) ([]GiaID, error) {
	all, err := u.FrameOutputs(frame)
	if err != nil {
		return nil, err
	}
	return all[:u.outputs-u.constraints], nil
}
func (u *GiaUnrolling) ConstraintOutputs(frame int) ([]GiaID, error) {
	all, err := u.FrameOutputs(frame)
	if err != nil {
		return nil, err
	}
	return all[u.outputs-u.constraints:], nil
}
func (m *GiaManager) Unroll(frames int) (*GiaUnrolling, error) {
	return m.UnrollWithOptions(GiaUnrollOptions{Frames: frames})
}
func (m *GiaManager) UnrollWithOptions(options GiaUnrollOptions) (*GiaUnrolling, error) {
	if m == nil {
		return nil, ErrDimension
	}
	if err := m.Check(); err != nil {
		return nil, err
	}
	if options.Frames < 1 || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxMappingEntries == 0 {
		options.MaxMappingEntries = 4 << 20
	}
	initial := options.Initial
	if initial == nil {
		initial = make([]LatchInit, m.registers)
	}
	if len(initial) != m.registers {
		return nil, ErrDimension
	}
	for _, value := range initial {
		if value > LatchInitUnknown {
			return nil, ErrDimension
		}
	}
	if options.Stop != nil && options.Stop() {
		return nil, ErrCancelled
	}
	entries, err := checkedSize(0, options.Frames, len(m.objects), options.MaxMappingEntries)
	if err != nil {
		return nil, err
	}
	target, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	target.Name = m.Name
	target.Spec = m.Spec
	if err := target.HashStart(); err != nil {
		return nil, err
	}
	u := &GiaUnrolling{graph: target, frames: options.Frames, sourceObjects: len(m.objects), inputs: m.PrimaryInputCount(), outputs: m.PrimaryOutputCount(), constraints: m.constraints, mapping: make([]GiaLit, entries), initial: make([]GiaLit, m.registers)}
	for i := range u.mapping {
		u.mapping[i] = InvalidGiaLit
	}
	state := make([]GiaLit, m.registers)
	for i, value := range initial {
		switch value {
		case LatchInitZero:
			state[i] = GiaConst0
		case LatchInitOne:
			state[i] = GiaConst1
		case LatchInitUnknown:
			state[i], err = target.AppendCI()
			if err != nil {
				return nil, err
			}
		}
		u.initial[i] = state[i]
	}
	for frame := 0; frame < options.Frames; frame++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		row := u.mapping[frame*len(m.objects) : (frame+1)*len(m.objects)]
		row[0] = GiaConst0
		for _, id := range m.cis[:u.inputs] {
			l, err := target.AppendCI()
			if err != nil {
				return nil, err
			}
			row[id] = l
			u.frameInputs = append(u.frameInputs, l)
		}
		for i, id := range m.cis[u.inputs:] {
			row[id] = state[i]
		}
		copyLiteral := func(l GiaLit) GiaLit { return row[l.ID()].NotIf(l.Complemented()) }
		for i, obj := range m.objects {
			if options.Stop != nil && options.Stop() {
				return nil, ErrCancelled
			}
			if !obj.IsAnd() {
				continue
			}
			a, _ := obj.Fanin0(GiaID(i))
			left := copyLiteral(a)
			if obj.IsBuffer() {
				row[i] = left
				continue
			}
			b, _ := obj.Fanin1(GiaID(i))
			right := copyLiteral(b)
			var l GiaLit
			var err error
			if c, yes := m.MuxControl(GiaID(i)); yes {
				l, err = target.HashMux(copyLiteral(c), right, left)
			} else if obj.IsXor() {
				l, err = target.HashXor(left, right)
			} else {
				l, err = target.HashAnd(left, right)
			}
			if err != nil {
				return nil, err
			}
			row[i] = l
		}
		for _, id := range m.cos {
			driver, _ := m.objects[id].Fanin0(id)
			row[id] = copyLiteral(driver)
		}
		for _, id := range m.cos[:u.outputs] {
			l, err := target.AppendCO(row[id])
			if err != nil {
				return nil, err
			}
			u.frameOutputs = append(u.frameOutputs, l.ID())
		}
		// All register inputs must be saved before any register-output mapping is
		// reused for the next frame: simultaneous clocked update, never in place.
		for i, id := range m.cos[u.outputs:] {
			state[i] = row[id]
		}
	}
	target.HashStop()
	cleaned, remap, err := target.Cleanup()
	if err != nil {
		return nil, err
	}
	u.graph = cleaned
	remapLiteral := func(l GiaLit) GiaLit {
		if l == InvalidGiaLit || remap[l.ID()] == InvalidGiaLit {
			return InvalidGiaLit
		}
		return remap[l.ID()].NotIf(l.Complemented())
	}
	for i, l := range u.mapping {
		u.mapping[i] = remapLiteral(l)
	}
	for i, l := range u.frameInputs {
		u.frameInputs[i] = remapLiteral(l)
	}
	for i, l := range u.initial {
		u.initial[i] = remapLiteral(l)
	}
	for i, id := range u.frameOutputs {
		u.frameOutputs[i] = remap[id].ID()
	}
	return u, nil
}
