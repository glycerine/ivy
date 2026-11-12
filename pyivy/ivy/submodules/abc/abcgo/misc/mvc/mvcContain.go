package mvc

func Mvc_CoverContain(c *Mvc_Cover_t) bool {
	before := c.Cubes.Items
	Mvc_CoverSetCubeSizes(c)
	Mvc_CoverSort(c, nil, Mvc_CubeCompareSizeAndInt)
	Mvc_CoverRemoveDuplicates(c)
	Mvc_CoverRemoveContained(c)
	return before != c.Cubes.Items
}
func Mvc_CoverRemoveDuplicates(c *Mvc_Cover_t) {
	var prev *Mvc_Cube_t
	for x := c.Cubes.Head; x != nil; {
		next := x.Next
		duplicate := false
		for y := c.Cubes.Head; y != x; y = y.Next {
			if Mvc_CubeBitEqual(x, y) {
				duplicate = true
				break
			}
		}
		if duplicate {
			Mvc_CoverDeleteCube(c, prev, x)
		} else {
			prev = x
		}
		x = next
	}
}
func Mvc_CoverRemoveContained(c *Mvc_Cover_t) {
	var prev *Mvc_Cube_t
	for x := c.Cubes.Head; x != nil; {
		next := x.Next
		contained := false
		for y := c.Cubes.Head; y != nil; y = y.Next {
			if y != x && !Mvc_CubeBitNotImpl(x, y) {
				contained = true
				break
			}
		}
		if contained {
			Mvc_CoverDeleteCube(c, prev, x)
		} else {
			prev = x
		}
		x = next
	}
}
