package exor

func AllocateCover(m *Manager, nCubes int) int {
	if nCubes < 0 {
		panic("exor: negative cube allocation")
	}
	m.free = make([]*Cube, nCubes)
	for i := range m.free {
		m.free[i] = &Cube{Inputs: make([]uint32, (m.Inputs+15)/16), Outputs: make([]uint32, (m.Outputs+31)/32)}
	}
	return nCubes * (4 * ((m.Inputs+15)/16 + (m.Outputs+31)/32))
}
func DelocateCover(m *Manager) { m.Cubes = nil; m.free = nil; m.pairs = nil; m.linkGroups = nil }
func AddToFreeCubes(m *Manager, c *Cube) {
	for _, x := range m.free {
		if x == c {
			panic("exor: cube freed twice")
		}
	}
	if c.active {
		panic("exor: free of active cube")
	}
	m.free = append(m.free, c)
}
func GetFreeCube(m *Manager) *Cube {
	var c *Cube
	if len(m.free) == 0 {
		c = &Cube{Inputs: make([]uint32, (m.Inputs+15)/16), Outputs: make([]uint32, (m.Outputs+31)/32)}
	} else {
		c = m.free[len(m.free)-1]
		m.free = m.free[:len(m.free)-1]
		clear(c.Inputs)
		clear(c.Outputs)
	}
	c.ID = m.nextID
	m.nextID++
	return c
}
