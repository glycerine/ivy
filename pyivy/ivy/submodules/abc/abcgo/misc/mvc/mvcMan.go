package mvc

import "sync"

var managerPool = sync.Pool{New: func() any { return &Mvc_Manager_t{} }}

// Mvc_ManagerStart returns an empty manager, reusing a stopped one's slabs.
func Mvc_ManagerStart() *Mvc_Manager_t { return managerPool.Get().(*Mvc_Manager_t) }

// Mvc_ManagerFree leaves the manager to the collector.
func Mvc_ManagerFree(_ *Mvc_Manager_t) {}

// Mvc_ManagerStop releases every cube and cover the manager handed out, as
// freeing ABC's manager does: none of them may be used afterwards.
func Mvc_ManagerStop(m *Mvc_Manager_t) {
	if m == nil {
		return
	}
	m.cubes, m.cubeSlab = nil, 0
	m.covers, m.coverSlab = nil, 0
	managerPool.Put(m)
}

const mvcSlabSize = 256

func (m *Mvc_Manager_t) newCube() *Mvc_Cube_t {
	if len(m.cubes) == 0 {
		if m.cubeSlab == len(m.cubeSlabs) {
			m.cubeSlabs = append(m.cubeSlabs, make([]Mvc_Cube_t, mvcSlabSize))
		}
		m.cubes = m.cubeSlabs[m.cubeSlab]
		m.cubeSlab++
	}
	r := &m.cubes[0]
	m.cubes = m.cubes[1:]
	*r = Mvc_Cube_t{} // a reused slab holds earlier cubes
	return r
}

func (m *Mvc_Manager_t) newCover() *Mvc_Cover_t {
	if len(m.covers) == 0 {
		if m.coverSlab == len(m.coverSlabs) {
			m.coverSlabs = append(m.coverSlabs, make([]Mvc_Cover_t, mvcSlabSize/4))
		}
		m.covers = m.coverSlabs[m.coverSlab]
		m.coverSlab++
	}
	r := &m.covers[0]
	m.covers = m.covers[1:]
	// keep the slices' storage (their contents belong to released cubes)
	array, lits := r.Array[:0], r.Lits[:0]
	*r = Mvc_Cover_t{Array: array, Lits: lits}
	return r
}
