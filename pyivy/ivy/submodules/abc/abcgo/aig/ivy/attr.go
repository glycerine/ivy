// Attribute management ported from ABC attr.h (Alan Mishchenko).
package ivy

import "reflect"

// Attr_Man_t keeps typed integers or Go values, never pointer bits disguised as
// integers. Slot cells keep FetchSpot addresses stable when the table grows.
// Managed attributes are zeroed byte slices owned by Go's garbage collector.
type Attr_Man_t struct {
	integers    bool
	managed     bool
	attrSize    int32
	ints        []*int32
	pointers    []*any
	context     any
	freeManager func(any)
	freeObject  func(any, any)
	stopped     bool
}

func Attr_ManAlloc(size, managed int32) *Attr_Man_t {
	if size < 0 || managed != 0 && size == 0 {
		panic("invalid attribute size")
	}
	return &Attr_Man_t{attrSize: size, managed: managed != 0}
}
func Attr_ManStartInt(n int32) *Attr_Man_t {
	if n < 0 {
		panic("negative attribute count")
	}
	return &Attr_Man_t{integers: true, attrSize: 4, ints: make([]*int32, n)}
}
func Attr_ManStartPtr(n int32) *Attr_Man_t {
	if n < 0 {
		panic("negative attribute count")
	}
	return &Attr_Man_t{pointers: make([]*any, n)}
}
func Attr_ManStartPtrMem(n, size int32) *Attr_Man_t {
	if n < 0 {
		panic("negative attribute count")
	}
	p := Attr_ManAlloc(size, 1)
	p.pointers = make([]*any, n)
	for i := range p.pointers {
		value := any(make([]byte, size))
		p.pointers[i] = &value
	}
	return p
}
func (p *Attr_Man_t) check(integer bool, id int32) {
	if p == nil || p.stopped {
		panic("attribute manager is closed")
	}
	if p.integers != integer {
		panic("attribute kind mismatch")
	}
	if id < 0 {
		panic("negative attribute ID")
	}
}

// SetCleanup configures the callbacks formerly stored as C function pointers.
func (p *Attr_Man_t) SetCleanup(context any, manager func(any), object func(any, any)) {
	if p == nil || p.stopped {
		panic("attribute manager is closed")
	}
	p.context = context
	p.freeManager = manager
	p.freeObject = object
}
func attrNonNil(value any) bool {
	if value == nil {
		return false
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return !v.IsNil()
	}
	return true
}
func Attr_ManStop(p *Attr_Man_t) {
	if p == nil || p.stopped {
		return
	}
	p.stopped = true
	ints, pointers := p.ints, p.pointers
	context, manager, object := p.context, p.freeManager, p.freeObject
	p.ints = nil
	p.pointers = nil
	p.context = nil
	p.freeManager = nil
	p.freeObject = nil
	if manager != nil && attrNonNil(context) {
		defer manager(context)
	}
	if object == nil {
		return
	}
	if p.integers {
		for _, slot := range ints {
			if slot != nil && *slot != 0 {
				object(context, *slot)
			}
		}
	} else {
		for _, slot := range pointers {
			if slot != nil && attrNonNil(*slot) {
				object(context, *slot)
			}
		}
	}
}
func Attr_ManReadAttrInt(p *Attr_Man_t, id int32) int32 {
	p.check(true, id)
	if int64(id) >= int64(len(p.ints)) || p.ints[id] == nil {
		return 0
	}
	return *p.ints[id]
}
func Attr_ManReadAttrPtr(p *Attr_Man_t, id int32) any {
	p.check(false, id)
	if int64(id) >= int64(len(p.pointers)) || p.pointers[id] == nil {
		return nil
	}
	return *p.pointers[id]
}
func Attr_ManWriteAttrInt(p *Attr_Man_t, id, value int32) {
	p.check(true, id)
	if int64(id) >= int64(len(p.ints)) {
		panic("attribute ID not allocated")
	}
	if p.ints[id] == nil {
		p.ints[id] = new(int32)
	}
	*p.ints[id] = value
}
func Attr_ManWriteAttrPtr(p *Attr_Man_t, id int32, value any) {
	p.check(false, id)
	if p.managed {
		panic("cannot replace managed attribute")
	}
	if int64(id) >= int64(len(p.pointers)) {
		panic("attribute ID not allocated")
	}
	if p.pointers[id] == nil {
		p.pointers[id] = new(any)
	}
	if !attrNonNil(value) {
		value = nil
	}
	*p.pointers[id] = value
}
func attrGrowSize(old int, id int32) int {
	n := 2 * old
	if n == 0 {
		n = 1024
	}
	if int64(n) <= int64(id) {
		n = int(id) + 1
	}
	return n
}
func Attr_ManFetchSpotInt(p *Attr_Man_t, id int32) *int32 {
	p.check(true, id)
	if int64(id) >= int64(len(p.ints)) {
		n := attrGrowSize(len(p.ints), id)
		p.ints = append(p.ints, make([]*int32, n-len(p.ints))...)
	}
	if p.ints[id] == nil {
		p.ints[id] = new(int32)
	}
	return p.ints[id]
}
func Attr_ManFetchSpotPtr(p *Attr_Man_t, id int32) *any {
	p.check(false, id)
	if int64(id) >= int64(len(p.pointers)) {
		n := attrGrowSize(len(p.pointers), id)
		p.pointers = append(p.pointers, make([]*any, n-len(p.pointers))...)
	}
	if p.pointers[id] == nil {
		p.pointers[id] = new(any)
	}
	if p.managed && *p.pointers[id] == nil {
		*p.pointers[id] = make([]byte, p.attrSize)
	}
	return p.pointers[id]
}
func Attr_ManFetchAttrInt(p *Attr_Man_t, id int32) int32 { return *Attr_ManFetchSpotInt(p, id) }
func Attr_ManFetchAttrPtr(p *Attr_Man_t, id int32) any   { return *Attr_ManFetchSpotPtr(p, id) }
func Attr_ManSetAttrInt(p *Attr_Man_t, id, value int32)  { *Attr_ManFetchSpotInt(p, id) = value }
func Attr_ManSetAttrPtr(p *Attr_Man_t, id int32, value any) {
	p.check(false, id)
	if p.managed {
		panic("cannot replace managed attribute")
	}
	if !attrNonNil(value) {
		value = nil
	}
	*Attr_ManFetchSpotPtr(p, id) = value
}
