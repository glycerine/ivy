package st

import (
	"fmt"
	"reflect"
)

func New(compare CompareFunc, hash HashFunc) *Table { return st__init_table(compare, hash) }
func st__init_table(c CompareFunc, h HashFunc) *Table {
	return st__init_table_with_params(c, h, 11, 5, 2, 0)
}
func st__init_table_with_params(c CompareFunc, h HashFunc, size, density int, growth float64, reorder int) *Table {
	if size < 1 {
		size = 1
	}
	if density < 1 {
		density = 1
	}
	if growth <= 1 {
		growth = 2
	}
	if c == nil {
		c = st__ptrcmp
	}
	if h == nil {
		h = st__ptrhash
	}
	return &Table{compare: c, hash: h, bins: make([]*Entry, size), density: density, growth: growth, reorder: reorder != 0}
}
func (t *Table) bucket(k any) int {
	n := t.hash(k, len(t.bins)) % len(t.bins)
	if n < 0 {
		n += len(t.bins)
	}
	return n
}
func (t *Table) find(k any) **Entry {
	b := t.bucket(k)
	p := &t.bins[b]
	for *p != nil {
		if t.compare((*p).Key, k) == 0 {
			if t.reorder && p != &t.bins[b] {
				e := *p
				*p = e.next
				e.next = t.bins[b]
				t.bins[b] = e
				return &t.bins[b]
			}
			return p
		}
		p = &(*p).next
	}
	return p
}
func (t *Table) grow() {
	old := t.bins
	n := int(float64(len(old)) * t.growth)
	if n <= len(old) {
		n = len(old) + 1
	}
	if n%2 == 0 {
		n++
	}
	t.bins = make([]*Entry, n)
	for _, head := range old {
		for e := head; e != nil; {
			next := e.next
			b := t.bucket(e.Key)
			e.next = t.bins[b]
			t.bins[b] = e
			e = next
		}
	}
}
func (t *Table) Count() int { return t.count }
func (t *Table) Lookup(k any) (any, bool) {
	e := *t.find(k)
	if e == nil {
		return nil, false
	}
	return e.Value, true
}
func (t *Table) Insert(k, v any) bool {
	e := *t.find(k)
	if e != nil {
		e.Value = v
		return true
	}
	t.AddDirect(k, v)
	return false
}
func (t *Table) AddDirect(k, v any) {
	if t.count/len(t.bins) >= t.density {
		t.grow()
	}
	b := t.bucket(k)
	t.bins[b] = &Entry{k, v, t.bins[b]}
	t.count++
}
func (t *Table) Delete(k any) (any, any, bool) {
	p := t.find(k)
	if *p == nil {
		return nil, nil, false
	}
	e := *p
	*p = e.next
	t.count--
	return e.Key, e.Value, true
}
func (t *Table) Clear() {
	for i := range t.bins {
		t.bins[i] = nil
	}
	t.count = 0
}
func (t *Table) Copy() *Table {
	r := st__init_table_with_params(t.compare, t.hash, len(t.bins), t.density, t.growth, 0)
	r.reorder = t.reorder
	for i, e := range t.bins {
		p := &r.bins[i]
		for ; e != nil; e = e.next {
			*p = &Entry{Key: e.Key, Value: e.Value}
			p = &(*p).next
			r.count++
		}
	}
	return r
}
func (t *Table) ForEach(fn func(any, any) VisitResult) bool {
	for i := range t.bins {
		p := &t.bins[i]
		for *p != nil {
			e := *p
			switch fn(e.Key, e.Value) {
			case Stop:
				return false
			case Delete:
				*p = e.next
				t.count--
			default:
				p = &e.next
			}
		}
	}
	return true
}
func (t *Table) Generator() *Generator { return &Generator{table: t} }
func (g *Generator) Next() (any, any, bool) {
	for g.entry == nil && g.index < len(g.table.bins) {
		g.entry = g.table.bins[g.index]
		g.index++
	}
	if g.entry == nil {
		return nil, nil, false
	}
	e := g.entry
	g.entry = e.next
	return e.Key, e.Value, true
}
func st__free_table(t *Table) { t.Clear() }
func st__lookup(t *Table, k any, out *any) int {
	v, ok := t.Lookup(k)
	if ok && out != nil {
		*out = v
	}
	return truth(ok)
}
func st__lookup_int(t *Table, k any, out *int) int {
	v, ok := t.Lookup(k)
	if ok && out != nil {
		*out = int(identity(v))
	}
	return truth(ok)
}
func st__insert(t *Table, k, v any) int     { return truth(t.Insert(k, v)) }
func st__add_direct(t *Table, k, v any) int { t.AddDirect(k, v); return 1 }
func st__find(t *Table, k any, out **any) int {
	e := *t.find(k)
	if e == nil {
		return 0
	}
	if out != nil {
		*out = &e.Value
	}
	return 1
}
func st__find_or_add(t *Table, k any, out **any) int {
	e := *t.find(k)
	old := e != nil
	if !old {
		t.AddDirect(k, nil)
		e = *t.find(k)
	}
	if out != nil {
		*out = &e.Value
	}
	return truth(old)
}
func st__copy(t *Table) *Table { return t.Copy() }
func st__delete(t *Table, k *any, out *any) int {
	key, v, ok := t.Delete(*k)
	if ok {
		*k = key
		if out != nil {
			*out = v
		}
	}
	return truth(ok)
}
func st__delete_int(t *Table, k *int, out *any) int {
	key := any(*k)
	ok := st__delete(t, &key, out)
	if ok != 0 {
		*k = int(identity(key))
	}
	return ok
}
func st__foreach(t *Table, f st__PFSR, arg any) int {
	return truth(t.ForEach(func(k, v any) VisitResult { return f(k, v, arg) }))
}
func st__init_gen(t *Table) *Generator { return t.Generator() }
func st__gen(g *Generator, k, v *any) int {
	key, value, ok := g.Next()
	if ok {
		if k != nil {
			*k = key
		}
		if v != nil {
			*v = value
		}
	}
	return truth(ok)
}
func st__gen_int(g *Generator, k *any, v *int) int {
	var value any
	ok := st__gen(g, k, &value)
	if ok != 0 && v != nil {
		*v = int(identity(value))
	}
	return ok
}
func st__free_gen(g *Generator)         { g.entry = nil; g.index = len(g.table.bins) }
func st__is_member(t *Table, k any) int { return st__lookup(t, k, nil) }
func st__count(t *Table) int            { return t.Count() }
func truth(b bool) int {
	if b {
		return 1
	}
	return 0
}
func st__strcmp(a, b any) int {
	x, y := a.(string), b.(string)
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}
func st__strhash(k any, n int) int {
	var h uint64
	for _, c := range []byte(k.(string)) {
		h = h*997 + uint64(c)
	}
	return int(h % uint64(n))
}
func identity(k any) uint64 {
	if k == nil {
		return 0
	}
	v := reflect.ValueOf(k)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint()
	case reflect.Pointer, reflect.Chan, reflect.UnsafePointer:
		return uint64(v.Pointer())
	}
	panic(fmt.Sprintf("st: unsupported identity key %T", k))
}
func st__numhash(k any, n int) int { return int(identity(k) % uint64(n)) }
func st__ptrhash(k any, n int) int { return int((identity(k) >> 2) % uint64(n)) }
func st__numcmp(a, b any) int      { return truth(identity(a) != identity(b)) }
func st__ptrcmp(a, b any) int      { return st__numcmp(a, b) }

// StringCompare and StringHash select string-content keys rather than pointer
// identity. Numeric and pointer comparators treat the key as an opaque identity.
func StringCompare(a, b any) int   { return st__strcmp(a, b) }
func StringHash(k any, n int) int  { return st__strhash(k, n) }
func NumericCompare(a, b any) int  { return st__numcmp(a, b) }
func NumericHash(k any, n int) int { return st__numhash(k, n) }
func PointerCompare(a, b any) int  { return st__ptrcmp(a, b) }
func PointerHash(k any, n int) int { return st__ptrhash(k, n) }

// Find returns a stable value slot until its entry is removed or the table is
// cleared. Rehashing preserves slot addresses, as in the original C table.
func (t *Table) Find(k any) (*any, bool) {
	e := *t.find(k)
	if e == nil {
		return nil, false
	}
	return &e.Value, true
}
func (t *Table) FindOrAdd(k any) (*any, bool) {
	if slot, ok := t.Find(k); ok {
		return slot, true
	}
	t.AddDirect(k, nil)
	slot, _ := t.Find(k)
	return slot, false
}
