package abc

// Port of src/aig/gia/giaAig.c, giaScl.c and giaDup.c: conversion and cleanup.
// Explicit mapping slices replace writes to the input graph's scratch data.

// GiaFromAig converts every live AIG object, including dangling logic and
// buffers. The object map contains new literals for old regular objects;
// complement it to map a complemented old signal. Tombstones map to invalid.
// CI/CO order, register partition, constraints, Name and Spec are preserved.
func GiaFromAig(source *AigManager) (*GiaManager, []GiaLit, error) {
	if source == nil {
		return nil, nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, nil, err
	}
	target := NewGiaManager()
	target.Name = source.Name
	target.Spec = source.Spec
	mapping := make([]GiaLit, len(source.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst1
	copyLiteral := func(l AigLit) GiaLit { return mapping[l.ID()].NotIf(l.Complemented()) }
	for i, obj := range source.objects {
		if i == 0 || obj.Type == AigNone {
			continue
		}
		var l GiaLit
		var err error
		switch obj.Type {
		case AigCI:
			l, err = target.AppendCI()
		case AigCO:
			l, err = target.AppendCO(copyLiteral(obj.Fanin0))
		case AigBuffer:
			l, err = target.AppendBuffer(copyLiteral(obj.Fanin0))
		case AigAND:
			l, err = target.AppendAnd(copyLiteral(obj.Fanin0), copyLiteral(obj.Fanin1))
		default:
			return nil, nil, ErrUnsupported
		}
		if err != nil {
			return nil, nil, err
		}
		mapping[i] = l
	}
	if err := target.SetRegisterCount(source.registers); err != nil {
		return nil, nil, err
	}
	if err := target.SetConstraintCount(source.constraints); err != nil {
		return nil, nil, err
	}
	if err := target.Check(); err != nil {
		return nil, nil, err
	}
	return target, mapping, nil
}

// ToAig lowers native XOR/mux nodes and structurally hashes ANDs. The explicit
// old-object-to-new-literal map handles merging without assuming equal IDs.
// Every object is converted; use Cleanup first to discard dangling logic.
func (source *GiaManager) ToAig() (*AigManager, []AigLit, error) {
	if source == nil {
		return nil, nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, nil, err
	}
	target := NewAigManager()
	target.Name = source.Name
	target.Spec = source.Spec
	mapping := make([]AigLit, len(source.objects))
	mapping[0] = AigConst0
	copyLiteral := func(l GiaLit) AigLit { return mapping[l.ID()].NotIf(l.Complemented()) }
	for i, obj := range source.objects {
		if i == 0 {
			continue
		}
		var l AigLit
		var err error
		if obj.IsCI() {
			l, err = target.CreateCI()
		} else {
			a, _ := obj.Fanin0(GiaID(i))
			left := copyLiteral(a)
			if obj.IsCO() {
				var id AigID
				id, err = target.CreateCO(left)
				l = AigLit(id) << 1
			} else if obj.IsBuffer() {
				l, err = target.CreateBuffer(left)
			} else {
				b, _ := obj.Fanin1(GiaID(i))
				right := copyLiteral(b)
				if c, yes := source.MuxControl(GiaID(i)); yes {
					l, err = target.Mux(copyLiteral(c), right, left)
				} else if obj.IsXor() {
					l, err = target.Xor(left, right)
				} else {
					l, err = target.And(left, right)
				}
			}
		}
		if err != nil {
			return nil, nil, err
		}
		mapping[i] = l
	}
	if err := target.SetRegisterCount(source.registers); err != nil {
		return nil, nil, err
	}
	if err := target.SetConstraintCount(source.constraints); err != nil {
		return nil, nil, err
	}
	if err := target.Check(); err != nil {
		return nil, nil, err
	}
	return target, mapping, nil
}

// Cleanup implements Gia_ManCleanup/Gia_ManDupMarked. It returns a new dense
// graph and an old-object-to-new-literal map; removed logic maps to invalid.
// As in C, all CIs, COs and buffers are retained, including dangling buffers
// and their cones. Native XOR/mux structure is preserved without rehashing.
// The source's marks, phases and Value scratch fields are never modified.
func (source *GiaManager) Cleanup() (*GiaManager, []GiaLit, error) {
	return source.cleanupOutputs(0)
}

// CleanupOutputs implements Gia_ManCleanupOutputs. It removes the first skip
// combinational outputs but deliberately retains logic reached only by them:
// the source marks through every output before marking the output prefix for
// omission. At least one output remains. Sequential graphs and requests that
// split the constraint suffix are rejected instead of relying on assertions.
func (source *GiaManager) CleanupOutputs(skip int) (*GiaManager, []GiaLit, error) {
	if source == nil || skip < 0 {
		return nil, nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, nil, err
	}
	if source.registers != 0 || skip >= len(source.cos) || skip > len(source.cos)-source.constraints {
		return nil, nil, ErrDimension
	}
	return source.cleanupOutputs(skip)
}

func (source *GiaManager) cleanupOutputs(skip int) (*GiaManager, []GiaLit, error) {
	if source == nil {
		return nil, nil, ErrDimension
	}
	if err := source.Check(); err != nil {
		return nil, nil, err
	}
	used := make([]bool, len(source.objects))
	used[0] = true
	stack := make([]GiaID, 0, len(source.cos)+source.buffers)
	for i, obj := range source.objects {
		if obj.IsCI() {
			used[i] = true
		}
		if obj.IsCO() || obj.IsBuffer() {
			stack = append(stack, GiaID(i))
		}
	}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if used[id] {
			continue
		}
		used[id] = true
		obj := source.objects[id]
		if obj.IsCI() || id == 0 {
			continue
		}
		a, _ := obj.Fanin0(id)
		stack = append(stack, a.ID())
		if obj.IsAnd() && !obj.IsBuffer() {
			b, _ := obj.Fanin1(id)
			stack = append(stack, b.ID())
			if c, yes := source.MuxControl(id); yes {
				stack = append(stack, c.ID())
			}
		}
	}
	target := NewGiaManager()
	target.objectLimit = source.objectLimit
	target.Name = source.Name
	target.Spec = source.Spec
	mapping := make([]GiaLit, len(source.objects))
	for i := range mapping {
		mapping[i] = InvalidGiaLit
	}
	mapping[0] = GiaConst0
	copyLiteral := func(l GiaLit) GiaLit { return mapping[l.ID()].NotIf(l.Complemented()) }
	for i, obj := range source.objects {
		if i == 0 || !used[i] {
			continue
		}
		if obj.IsCO() && int(obj.Diff1()) < skip {
			continue
		}
		var l GiaLit
		var err error
		if obj.IsCI() {
			l, err = target.AppendCI()
		} else {
			a, _ := obj.Fanin0(GiaID(i))
			left := copyLiteral(a)
			if obj.IsCO() {
				l, err = target.AppendCO(left)
			} else if obj.IsBuffer() {
				l, err = target.AppendBuffer(left)
			} else {
				b, _ := obj.Fanin1(GiaID(i))
				right := copyLiteral(b)
				if c, yes := source.MuxControl(GiaID(i)); yes {
					l, err = target.AppendMuxReal(copyLiteral(c), right, left)
				} else if obj.IsXor() {
					l, err = target.AppendXorReal(left, right)
				} else {
					l, err = target.AppendAnd(left, right)
				}
			}
		}
		if err != nil {
			return nil, nil, err
		}
		mapping[i] = l
	}
	if err := target.SetRegisterCount(source.registers); err != nil {
		return nil, nil, err
	}
	if err := target.SetConstraintCount(source.constraints); err != nil {
		return nil, nil, err
	}
	if err := target.Check(); err != nil {
		return nil, nil, err
	}
	return target, mapping, nil
}
