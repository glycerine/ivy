package mvc

func Mvc_CoverAlloc(mem *Mvc_Manager_t, bits int) *Mvc_Cover_t {
	if bits < 0 {
		panic("mvc: negative bit count")
	}
	return &Mvc_Cover_t{NBits: bits, NWords: (bits + 31) / 32, NUnused: (32 - bits%32) % 32, Mem: mem}
}
func Mvc_CoverClone(c *Mvc_Cover_t) *Mvc_Cover_t { return Mvc_CoverAlloc(c.Mem, c.NBits) }
func Mvc_CoverDup(c *Mvc_Cover_t) *Mvc_Cover_t {
	r := Mvc_CoverClone(c)
	for x := c.Cubes.Head; x != nil; x = x.Next {
		Mvc_CoverAddCubeTail(r, Mvc_CubeDup(r, x))
	}
	return r
}
func Mvc_CoverFree(c *Mvc_Cover_t) {
	if c != nil {
		c.Cubes = Mvc_List_t{}
		c.Array = nil
		c.Lits = nil
		c.Mask = nil
	}
}
func Mvc_CoverCreateConst(mem *Mvc_Manager_t, bits, phase int) *Mvc_Cover_t {
	c := Mvc_CoverAlloc(mem, bits)
	if phase != 0 {
		Mvc_CoverMakeTautology(c)
	}
	return c
}
func Mvc_CoverAllocateMask(c *Mvc_Cover_t) {
	if c.Mask == nil {
		c.Mask = Mvc_CubeAlloc(c)
	}
}
func Mvc_CoverAllocateArrayLits(c *Mvc_Cover_t) {
	if len(c.Lits) != c.NBits {
		c.Lits = make([]int, c.NBits)
	}
}
func Mvc_CoverAllocateArrayCubes(c *Mvc_Cover_t)  { c.Array = make([]*Mvc_Cube_t, c.Cubes.Items) }
func Mvc_CoverDeallocateMask(c *Mvc_Cover_t)      { c.Mask = nil }
func Mvc_CoverDeallocateArrayLits(c *Mvc_Cover_t) { c.Lits = nil }
