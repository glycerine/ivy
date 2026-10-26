package abc

// Frontier representation from src/aig/gia/giaFront.c.

import "slices"

type GiaFrontierOptions struct {
	MaxObjects, MaxFrontier int
	Stop                    func() bool
}

type GiaFrontierKind uint8

const (
	GiaFrontierConstant GiaFrontierKind = iota
	GiaFrontierCI
	GiaFrontierCO
	GiaFrontierAND
	GiaFrontierXOR
)

type GiaFrontierObject struct {
	Kind                     GiaFrontierKind
	Slot, Fanin0, Fanin1     int
	Complement0, Complement1 bool
}

// GiaFrontierGraph stores fanins as reusable frontier slots. A slot belongs to
// the most recently created live object assigned to it; COs do not own slots.
type GiaFrontierGraph struct {
	Name, Spec             string
	Width, Peak            int
	Registers, Constraints int
	objects                []GiaFrontierObject
}

func checkGiaFrontierOptions(options GiaFrontierOptions) (GiaFrontierOptions, error) {
	if options.MaxObjects < 0 || options.MaxFrontier < 0 {
		return options, ErrDimension
	}
	if options.MaxObjects == 0 {
		options.MaxObjects = 1 << 20
	}
	if options.MaxFrontier == 0 {
		options.MaxFrontier = 1 << 20
	}
	return options, nil
}

func nextGiaFrontierSlot(used []bool, start int) (int, error) {
	if len(used) == 0 || start < 0 || start >= len(used) {
		return 0, ErrDimension
	}
	for checked := 0; checked < len(used); checked++ {
		candidate := (start + checked) % len(used)
		if !used[candidate] {
			used[candidate] = true
			return candidate, nil
		}
	}
	return 0, ErrInvariant
}

func giaFrontierReferences(manager *GiaManager, options GiaFrontierOptions) ([]int, error) {
	references := make([]int, manager.ObjectCount())
	for id, object := range manager.objects {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		if object.IsCO() {
			fanin, err := object.Fanin0(GiaID(id))
			if err != nil {
				return nil, err
			}
			references[fanin.ID()]++
		} else if object.IsAnd() {
			if object.IsBuffer() {
				return nil, ErrUnsupported
			}
			if _, yes := manager.MuxControl(GiaID(id)); yes {
				return nil, ErrUnsupported
			}
			first, err := object.Fanin0(GiaID(id))
			if err != nil {
				return nil, err
			}
			second, err := object.Fanin1(GiaID(id))
			if err != nil {
				return nil, err
			}
			references[first.ID()]++
			references[second.ID()]++
		}
	}
	return references, nil
}

// FrontierCrossCut returns the source-order peak number of simultaneously
// live object values. It owns its reference counts and leaves manager scratch
// untouched.
func (manager *GiaManager) FrontierCrossCut(options GiaFrontierOptions) (int, error) {
	if manager == nil {
		return 0, ErrDimension
	}
	options, err := checkGiaFrontierOptions(options)
	if err != nil {
		return 0, err
	}
	if err = manager.Check(); err != nil {
		return 0, err
	}
	if manager.ObjectCount() > options.MaxObjects {
		return 0, ErrLimit
	}
	references, err := giaFrontierReferences(manager, options)
	if err != nil {
		return 0, err
	}
	current, peak := 0, 0
	consume := func(id GiaID) error {
		if references[id] <= 0 {
			return ErrInvariant
		}
		references[id]--
		if references[id] == 0 {
			current--
		}
		return nil
	}
	for id, object := range manager.objects {
		if options.Stop != nil && options.Stop() {
			return 0, ErrCancelled
		}
		if references[id] > 0 {
			current++
			if current > peak {
				peak = current
			}
		}
		if object.IsCO() {
			fanin, _ := object.Fanin0(GiaID(id))
			if err = consume(fanin.ID()); err != nil {
				return 0, err
			}
		} else if object.IsAnd() {
			first, _ := object.Fanin0(GiaID(id))
			second, _ := object.Fanin1(GiaID(id))
			if err = consume(first.ID()); err != nil {
				return 0, err
			}
			if err = consume(second.ID()); err != nil {
				return 0, err
			}
		}
	}
	if current != 0 {
		return 0, ErrInvariant
	}
	return peak, nil
}

// EncodeGiaFrontier replaces object-ID differences with cyclically allocated
// live slots while retaining one typed record per source object.
func (manager *GiaManager) EncodeGiaFrontier(options GiaFrontierOptions) (*GiaFrontierGraph, error) {
	if manager == nil {
		return nil, ErrDimension
	}
	options, err := checkGiaFrontierOptions(options)
	if err != nil {
		return nil, err
	}
	if err = manager.Check(); err != nil {
		return nil, err
	}
	if manager.ObjectCount() > options.MaxObjects {
		return nil, ErrLimit
	}
	references, err := giaFrontierReferences(manager, options)
	if err != nil {
		return nil, err
	}
	peak, err := manager.FrontierCrossCut(options)
	if err != nil {
		return nil, err
	}
	width := 1 + int(float32(1.1)*float32(peak))
	if width > options.MaxFrontier {
		return nil, ErrLimit
	}
	result := &GiaFrontierGraph{
		Name: manager.Name, Spec: manager.Spec, Width: width, Peak: peak,
		Registers: manager.registers, Constraints: manager.constraints,
		objects: make([]GiaFrontierObject, manager.ObjectCount()),
	}
	used := make([]bool, width)
	objectSlots := make([]int, manager.ObjectCount())
	for index := range objectSlots {
		objectSlots[index] = -1
	}
	next := 0
	allocate := func(id int) error {
		slot, allocateErr := nextGiaFrontierSlot(used, next)
		if allocateErr != nil {
			return allocateErr
		}
		next = slot
		objectSlots[id] = slot
		result.objects[id].Slot = slot
		if references[id] == 0 {
			used[slot] = false
		}
		return nil
	}
	if err = allocate(0); err != nil {
		return nil, err
	}
	result.objects[0].Kind = GiaFrontierConstant
	consume := func(id GiaID) error {
		if references[id] <= 0 || objectSlots[id] < 0 {
			return ErrInvariant
		}
		references[id]--
		if references[id] == 0 {
			used[objectSlots[id]] = false
		}
		return nil
	}
	for id := 1; id < manager.ObjectCount(); id++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		object := manager.objects[id]
		record := &result.objects[id]
		record.Slot, record.Fanin0, record.Fanin1 = -1, -1, -1
		if object.IsCI() {
			record.Kind = GiaFrontierCI
			if err = allocate(id); err != nil {
				return nil, err
			}
			continue
		}
		first, faninErr := object.Fanin0(GiaID(id))
		if faninErr != nil || objectSlots[first.ID()] < 0 {
			return nil, ErrInvariant
		}
		record.Fanin0, record.Complement0 = objectSlots[first.ID()], first.Complemented()
		if object.IsCO() {
			record.Kind = GiaFrontierCO
			if err = consume(first.ID()); err != nil {
				return nil, err
			}
			continue
		}
		second, faninErr := object.Fanin1(GiaID(id))
		if faninErr != nil || objectSlots[second.ID()] < 0 {
			return nil, ErrInvariant
		}
		record.Fanin1, record.Complement1 = objectSlots[second.ID()], second.Complemented()
		if object.IsXor() {
			record.Kind = GiaFrontierXOR
		} else {
			record.Kind = GiaFrontierAND
		}
		if err = allocate(id); err != nil {
			return nil, err
		}
		if err = consume(first.ID()); err != nil {
			return nil, err
		}
		if err = consume(second.ID()); err != nil {
			return nil, err
		}
	}
	if slices.Contains(used, true) {
		return nil, ErrInvariant
	}
	return result, nil
}

func (frontier *GiaFrontierGraph) Objects() []GiaFrontierObject {
	if frontier == nil {
		return nil
	}
	return slices.Clone(frontier.objects)
}

// Decode reconstructs ordinary GIA object-ID differences from frontier slots.
func (frontier *GiaFrontierGraph) Decode(options GiaFrontierOptions) (*GiaManager, error) {
	if frontier == nil {
		return nil, ErrDimension
	}
	options, err := checkGiaFrontierOptions(options)
	if err != nil {
		return nil, err
	}
	if len(frontier.objects) == 0 || len(frontier.objects) > options.MaxObjects || frontier.Width < 1 || frontier.Width > options.MaxFrontier {
		return nil, ErrLimit
	}
	if frontier.objects[0].Kind != GiaFrontierConstant || frontier.objects[0].Slot < 0 || frontier.objects[0].Slot >= frontier.Width {
		return nil, ErrInvariant
	}
	manager, err := NewGiaManagerWithLimit(options.MaxObjects)
	if err != nil {
		return nil, err
	}
	manager.Name, manager.Spec = frontier.Name, frontier.Spec
	slots := make([]GiaLit, frontier.Width)
	valid := make([]bool, frontier.Width)
	slots[frontier.objects[0].Slot], valid[frontier.objects[0].Slot] = GiaConst0, true
	read := func(slot int, complement bool) (GiaLit, error) {
		if slot < 0 || slot >= len(slots) || !valid[slot] {
			return InvalidGiaLit, ErrInvariant
		}
		return slots[slot].NotIf(complement), nil
	}
	for id := 1; id < len(frontier.objects); id++ {
		if options.Stop != nil && options.Stop() {
			return nil, ErrCancelled
		}
		record := frontier.objects[id]
		var literal GiaLit
		switch record.Kind {
		case GiaFrontierCI:
			literal, err = manager.AppendCI()
		case GiaFrontierCO:
			var first GiaLit
			first, err = read(record.Fanin0, record.Complement0)
			if err == nil {
				literal, err = manager.AppendCO(first)
			}
		case GiaFrontierAND, GiaFrontierXOR:
			var first, second GiaLit
			first, err = read(record.Fanin0, record.Complement0)
			if err == nil {
				second, err = read(record.Fanin1, record.Complement1)
			}
			if err == nil && record.Kind == GiaFrontierAND {
				literal, err = manager.AppendAnd(first, second)
			} else if err == nil {
				literal, err = manager.AppendXorReal(first, second)
			}
		default:
			return nil, ErrInvariant
		}
		if err != nil {
			return nil, err
		}
		if literal.ID() != GiaID(id) {
			return nil, ErrInvariant
		}
		if record.Kind != GiaFrontierCO {
			if record.Slot < 0 || record.Slot >= len(slots) {
				return nil, ErrInvariant
			}
			slots[record.Slot], valid[record.Slot] = literal.Regular(), true
		}
	}
	if err = manager.SetRegisterCount(frontier.Registers); err != nil {
		return nil, err
	}
	if err = manager.SetConstraintCount(frontier.Constraints); err != nil {
		return nil, err
	}
	if err = manager.Check(); err != nil {
		return nil, err
	}
	return manager, nil
}

// FrontierRoundTrip performs the source test's encode/transform sequence and
// returns whether the exact compact object words and graph metadata survive.
func (manager *GiaManager) FrontierRoundTrip(options GiaFrontierOptions) (bool, error) {
	frontier, err := manager.EncodeGiaFrontier(options)
	if err != nil {
		return false, err
	}
	rebuilt, err := frontier.Decode(options)
	if err != nil {
		return false, err
	}
	return slices.Equal(manager.objects, rebuilt.objects) && manager.Name == rebuilt.Name && manager.Spec == rebuilt.Spec && manager.registers == rebuilt.registers && manager.constraints == rebuilt.constraints, nil
}
