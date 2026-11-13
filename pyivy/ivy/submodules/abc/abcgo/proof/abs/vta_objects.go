package abs

// Vta_Obj_t records one graph object at one timeframe. Its stable ID replaces
// the source's pointer subtraction into a reallocating C array.
type Vta_Obj_t struct {
	iObj, iFrame, iNext         int32
	Prio, Value, fAdded, fVisit uint32
	id                          int32
}

// vtaObjectTable preserves ABC's chained-hash lookup and stable numeric IDs.
// Individual objects are allocated separately so Go slice growth cannot
// invalidate pointers held during recursive traversal.
type vtaObjectTable struct {
	objects []*Vta_Obj_t
	bins    []int32
}

func newVtaObjectTable(capacity int) *vtaObjectTable {
	if capacity < 2 {
		capacity = 2
	}
	t := &vtaObjectTable{objects: make([]*Vta_Obj_t, 1, capacity), bins: make([]int32, 2*capacity+1)}
	return t
}

func (t *vtaObjectTable) object(id int32) *Vta_Obj_t {
	if id < 0 || int(id) >= len(t.objects) {
		panic("VTA object ID out of range")
	}
	return t.objects[id]
}

func vtaHash(object, frame int32, bins int) int {
	if bins <= 0 {
		panic("invalid VTA hash size")
	}
	sum := uint32(object) + uint32(frame)
	return int((sum * (sum + 1)) % uint32(bins))
}

func (t *vtaObjectTable) lookup(object, frame int32) *int32 {
	place := &t.bins[vtaHash(object, frame, len(t.bins))]
	for *place != 0 {
		current := t.object(*place)
		if current.iObj == object && current.iFrame == frame {
			break
		}
		place = &current.iNext
	}
	return place
}

func (t *vtaObjectTable) find(object, frame int32) *Vta_Obj_t {
	return t.object(*t.lookup(object, frame))
}

func (t *vtaObjectTable) findOrAdd(object, frame int32) *Vta_Obj_t {
	if object < 0 || frame < -1 {
		panic("invalid VTA object or frame")
	}
	if len(t.objects) == cap(t.objects) {
		t.bins = make([]int32, 2*len(t.bins)+1)
		for id := 1; id < len(t.objects); id++ {
			current := t.objects[id]
			if current.iNext < 0 {
				continue
			}
			current.iNext = 0
			*t.lookup(current.iObj, current.iFrame) = int32(id)
		}
	}
	place := t.lookup(object, frame)
	if *place != 0 {
		return t.object(*place)
	}
	id := int32(len(t.objects))
	current := &Vta_Obj_t{iObj: object, iFrame: frame, id: id}
	*place = id
	t.objects = append(t.objects, current)
	return current
}

func (t *vtaObjectTable) delete(object, frame int32) {
	place := t.lookup(object, frame)
	if *place == 0 {
		panic("missing VTA object")
	}
	current := t.object(*place)
	*place = current.iNext
	current.iNext = -1
}

func (t *vtaObjectTable) rollback(first int32) {
	if first < 1 || int(first) > len(t.objects) {
		panic("invalid VTA rollback")
	}
	for _, object := range t.objects[first:] {
		if object.iNext >= 0 {
			t.delete(object.iObj, object.iFrame)
		}
	}
	clear(t.objects[first:])
	t.objects = t.objects[:first]
}
