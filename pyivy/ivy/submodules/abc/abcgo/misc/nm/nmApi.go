package nm

import (
	"fmt"
	"sort"
)

func Nm_ManCreate(size int32) *Nm_Man_t {
	if size < 0 {
		size = 0
	}
	return &Manager{byID: make(map[int32]*Entry, int(size)), byName: make(map[string]*Entry, int(size))}
}
func Nm_ManFree(p *Nm_Man_t)             { clear(p.byID); clear(p.byName) }
func Nm_ManNumEntries(p *Nm_Man_t) int32 { return int32(len(p.byID)) }
func Nm_ManStoreIdName(p *Nm_Man_t, id, typ int32, name, suffix string) string {
	if p.byID[id] != nil {
		return ""
	}
	e := &Entry{ObjId: uint32(id), Type: uint32(typ), Name: name + suffix}
	Nm_ManTableAdd(p, e)
	return e.Name
}
func Nm_ManDeleteIdName(p *Nm_Man_t, id int32) { Nm_ManTableDelete(p, id) }
func Nm_ManCreateUniqueName(p *Nm_Man_t, id int32) string {
	if e := p.byID[id]; e != nil {
		return e.Name
	}
	name := fmt.Sprintf("n%d", id)
	for i := 1; p.byName[name] != nil; i++ {
		name = fmt.Sprintf("n%d_%d", id, i)
	}
	return name
}
func Nm_ManFindNameById(p *Nm_Man_t, id int32) string {
	if e := p.byID[id]; e != nil {
		return e.Name
	}
	return ""
}
func Nm_ManFindIdByName(p *Nm_Man_t, name string, typ int32) int32 {
	if e := Nm_ManTableLookupName(p, name, typ); e != nil {
		return int32(e.ObjId)
	}
	return -1
}
func Nm_ManFindIdByNameTwoTypes(p *Nm_Man_t, name string, a, b int32) int32 {
	id := Nm_ManFindIdByName(p, name, a)
	if id == -1 {
		return Nm_ManFindIdByName(p, name, b)
	}
	return id
}

// Nm_ManReturnNameIds returns every stored ID, sorted for reproducibility. The
// C implementation returned hash-bin heads only; Go has no hash-bin interface.
func Nm_ManReturnNameIds(p *Nm_Man_t) []int32 {
	ids := make([]int32, 0, len(p.byID))
	for id := range p.byID {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Lookup distinguishes a missing ID from an explicitly stored empty name.
func (p *Manager) Lookup(id int32) (string, bool) {
	e, ok := p.byID[id]
	if !ok {
		return "", false
	}
	return e.Name, true
}
