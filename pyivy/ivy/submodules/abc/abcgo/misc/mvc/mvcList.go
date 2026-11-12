package mvc

func Mvc_ListAddCubeHead(l *Mvc_List_t, c *Mvc_Cube_t) {
	c.Next = l.Head
	l.Head = c
	if l.Tail == nil {
		l.Tail = c
	}
	l.Items++
}
func Mvc_ListAddCubeTail(l *Mvc_List_t, c *Mvc_Cube_t) {
	c.Next = nil
	if l.Tail == nil {
		l.Head = c
	} else {
		l.Tail.Next = c
	}
	l.Tail = c
	l.Items++
}
func Mvc_ListDeleteCube(l *Mvc_List_t, prev, c *Mvc_Cube_t) {
	if prev == nil {
		l.Head = c.Next
	} else {
		prev.Next = c.Next
	}
	if l.Tail == c {
		l.Tail = prev
	}
	l.Items--
	c.Next = nil
}
func Mvc_CoverAddCubeHead(c *Mvc_Cover_t, x *Mvc_Cube_t)      { Mvc_ListAddCubeHead(&c.Cubes, x) }
func Mvc_CoverAddCubeTail(c *Mvc_Cover_t, x *Mvc_Cube_t)      { Mvc_ListAddCubeTail(&c.Cubes, x) }
func Mvc_CoverDeleteCube(c *Mvc_Cover_t, prev, x *Mvc_Cube_t) { Mvc_ListDeleteCube(&c.Cubes, prev, x) }
func Mvc_CoverAddDupCubeHead(c *Mvc_Cover_t, x *Mvc_Cube_t) {
	Mvc_CoverAddCubeHead(c, Mvc_CubeDup(c, x))
}
func Mvc_CoverAddDupCubeTail(c *Mvc_Cover_t, x *Mvc_Cube_t) {
	Mvc_CoverAddCubeTail(c, Mvc_CubeDup(c, x))
}
func Mvc_CoverAddLiteralsOfCube(c *Mvc_Cover_t, x *Mvc_Cube_t) {
	Mvc_CoverAllocateArrayLits(c)
	for i := 0; i < c.NBits; i++ {
		if Mvc_CubeBitValue(x, i) {
			c.Lits[i]++
		}
	}
}
func Mvc_CoverDeleteLiteralsOfCube(c *Mvc_Cover_t, x *Mvc_Cube_t) {
	for i := 0; i < c.NBits; i++ {
		if Mvc_CubeBitValue(x, i) {
			c.Lits[i]--
		}
	}
}
func Mvc_CoverList2Array(c *Mvc_Cover_t) {
	c.Array = c.Array[:0]
	for x := c.Cubes.Head; x != nil; x = x.Next {
		c.Array = append(c.Array, x)
	}
}
func Mvc_CoverArray2List(c *Mvc_Cover_t) {
	c.Cubes = Mvc_List_t{}
	for _, x := range c.Array {
		Mvc_CoverAddCubeTail(c, x)
	}
}
func Mvc_ListGetTailFromHead(x *Mvc_Cube_t) *Mvc_Cube_t {
	if x == nil {
		return nil
	}
	for x.Next != nil {
		x = x.Next
	}
	return x
}

func Mvc_ListAddCubeHead_(l *Mvc_List_t, c *Mvc_Cube_t)     { Mvc_ListAddCubeHead(l, c) }
func Mvc_ListAddCubeTail_(l *Mvc_List_t, c *Mvc_Cube_t)     { Mvc_ListAddCubeTail(l, c) }
func Mvc_ListDeleteCube_(l *Mvc_List_t, p, c *Mvc_Cube_t)   { Mvc_ListDeleteCube(l, p, c) }
func Mvc_CoverAddCubeHead_(c *Mvc_Cover_t, x *Mvc_Cube_t)   { Mvc_CoverAddCubeHead(c, x) }
func Mvc_CoverAddCubeTail_(c *Mvc_Cover_t, x *Mvc_Cube_t)   { Mvc_CoverAddCubeTail(c, x) }
func Mvc_CoverDeleteCube_(c *Mvc_Cover_t, p, x *Mvc_Cube_t) { Mvc_CoverDeleteCube(c, p, x) }
