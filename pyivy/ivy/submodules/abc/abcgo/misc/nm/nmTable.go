package nm

import "fmt"

func Nm_ManTableLookupId(p *Nm_Man_t, id int32) *Nm_Entry_t { return p.byID[id] }
func Nm_ManTableLookupName(p *Nm_Man_t, name string, typ int32) *Nm_Entry_t {
	first := p.byName[name]
	if first == nil {
		return nil
	}
	if typ == -1 || first.Type == uint32(typ) {
		return first
	}
	for e := first.pNameSake; e != nil && e != first; e = e.pNameSake {
		if e.Type == uint32(typ) {
			return e
		}
	}
	return nil
}
func Nm_ManTableAdd(p *Nm_Man_t, e *Nm_Entry_t) int32 {
	if p.byID[int32(e.ObjId)] != nil {
		return 0
	}
	if p.byID == nil {
		p.byID = make(map[int32]*Entry)
	}
	if p.byName == nil {
		p.byName = make(map[string]*Entry)
	}
	p.byID[int32(e.ObjId)] = e
	if other := p.byName[e.Name]; other != nil {
		e.pNameSake = other.pNameSake
		if e.pNameSake == nil {
			e.pNameSake = other
		}
		other.pNameSake = e
	} else {
		p.byName[e.Name] = e
	}
	return 1
}
func Nm_ManTableDelete(p *Nm_Man_t, id int32) int32 {
	e := p.byID[id]
	if e == nil {
		return 0
	}
	delete(p.byID, id)
	anchor := p.byName[e.Name]
	if e.pNameSake == nil {
		delete(p.byName, e.Name)
		return 1
	}
	prev := e
	for prev.pNameSake != e {
		prev = prev.pNameSake
	}
	if e.pNameSake == prev {
		prev.pNameSake = nil
	} else {
		prev.pNameSake = e.pNameSake
	}
	if anchor == e {
		p.byName[e.Name] = prev
	}
	e.pNameSake = nil
	return 1
}
func Nm_ManProfile(p *Nm_Man_t) {
	fmt.Printf("Name manager: %d entries, %d distinct names\n", len(p.byID), len(p.byName))
}

// Nm_ManResize rebuilds maps to release excess capacity after deletions.
func Nm_ManResize(p *Nm_Man_t) {
	ids := make(map[int32]*Entry, len(p.byID))
	names := make(map[string]*Entry, len(p.byName))
	for k, v := range p.byID {
		ids[k] = v
	}
	for k, v := range p.byName {
		names[k] = v
	}
	p.byID, p.byName = ids, names
}
