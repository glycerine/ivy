package mvc

func Mvc_CoverReadWordNum(c *Mvc_Cover_t) int            { return c.NWords }
func Mvc_CoverReadBitNum(c *Mvc_Cover_t) int             { return c.NBits }
func Mvc_CoverReadCubeNum(c *Mvc_Cover_t) int            { return c.Cubes.Items }
func Mvc_CoverReadCubeHead(c *Mvc_Cover_t) *Mvc_Cube_t   { return c.Cubes.Head }
func Mvc_CoverReadCubeTail(c *Mvc_Cover_t) *Mvc_Cube_t   { return c.Cubes.Tail }
func Mvc_CoverReadCubeList(c *Mvc_Cover_t) *Mvc_List_t   { return &c.Cubes }
func Mvc_ListReadCubeNum(l *Mvc_List_t) int              { return l.Items }
func Mvc_ListReadCubeHead(l *Mvc_List_t) *Mvc_Cube_t     { return l.Head }
func Mvc_ListReadCubeTail(l *Mvc_List_t) *Mvc_Cube_t     { return l.Tail }
func Mvc_CoverSetCubeNum(c *Mvc_Cover_t, n int)          { c.Cubes.Items = n }
func Mvc_CoverSetCubeHead(c *Mvc_Cover_t, x *Mvc_Cube_t) { c.Cubes.Head = x }
func Mvc_CoverSetCubeTail(c *Mvc_Cover_t, x *Mvc_Cube_t) { c.Cubes.Tail = x }
func Mvc_CoverSetCubeList(c *Mvc_Cover_t, l *Mvc_List_t) { c.Cubes = *l }
func Mvc_CoverIsEmpty(c *Mvc_Cover_t) bool               { return c.Cubes.Items == 0 }
func Mvc_CoverIsTautology(c *Mvc_Cover_t) bool {
	if c.Cubes.Items != 1 {
		return false
	}
	x := c.Cubes.Head
	for i := 0; i < c.NBits; i++ {
		if !Mvc_CubeBitValue(x, i) {
			return false
		}
	}
	return true
}
func Mvc_CoverIsBinaryBuffer(c *Mvc_Cover_t) bool {
	return c.NBits == 2 && c.Cubes.Items == 1 && !Mvc_CubeBitValue(c.Cubes.Head, 0) && Mvc_CubeBitValue(c.Cubes.Head, 1)
}
func Mvc_CoverMakeEmpty(c *Mvc_Cover_t) { c.Cubes = Mvc_List_t{} }
func Mvc_CoverMakeTautology(c *Mvc_Cover_t) {
	Mvc_CoverMakeEmpty(c)
	x := Mvc_CubeAlloc(c)
	Mvc_CubeBitFill(x)
	Mvc_CoverAddCubeTail(c, x)
}
func Mvc_CoverCreateEmpty(c *Mvc_Cover_t) *Mvc_Cover_t { return Mvc_CoverClone(c) }
func Mvc_CoverCreateTautology(c *Mvc_Cover_t) *Mvc_Cover_t {
	r := Mvc_CoverClone(c)
	Mvc_CoverMakeTautology(r)
	return r
}
