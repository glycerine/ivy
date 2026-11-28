package mvc

import "math/bits"

// literalCount counts the cubes of c that contain literal i (one column of
// mvcLits.c's loops).
func literalCount(c *Mvc_Cover_t, i int) int {
	word, bit := i>>5, uint32(1)<<uint(i&31)
	n := 0
	for x := c.Cubes.Head; x != nil; x = x.Next {
		if x.Data[word]&bit != 0 {
			n++
		}
	}
	return n
}

// Mvc_CoverAnyLiteral returns the last literal (of the mask) that occurs in
// more than one cube, as mvcLits.c (fUseFirst = 0): it stops at the second
// occurrence.
func Mvc_CoverAnyLiteral(c *Mvc_Cover_t, m *Mvc_Cube_t) int {
	for i := c.NBits - 1; i >= 0; i-- {
		if m != nil && !Mvc_CubeBitValue(m, i) {
			continue
		}
		word, bit := i>>5, uint32(1)<<uint(i&31)
		n := 0
		for x := c.Cubes.Head; x != nil; x = x.Next {
			if x.Data[word]&bit != 0 {
				n++
				if n > 1 {
					return i
				}
			}
		}
	}
	return -1
}
func Mvc_CoverBestLiteral(c *Mvc_Cover_t, m *Mvc_Cube_t) int {
	best, most := -1, -1
	for i := 0; i < c.NBits; i++ {
		if m != nil && !Mvc_CubeBitValue(m, i) {
			continue
		}
		if k := literalCount(c, i); most < k {
			best, most = i, k
		}
	}
	if most > 1 {
		return best
	}
	return -1
}
func Mvc_CoverWorstLiteral(c *Mvc_Cover_t, m *Mvc_Cube_t) int {
	worst, least := -1, 1000000
	for i := 0; i < c.NBits; i++ {
		if m != nil && !Mvc_CubeBitValue(m, i) {
			continue
		}
		// skip the literal that does not occur or occurs once
		if k := literalCount(c, i); k >= 2 && least > k {
			worst, least = i, k
		}
	}
	if least < 1000000 {
		return worst
	}
	return -1
}
func Mvc_CoverBestLiteralCover(c, s *Mvc_Cover_t) *Mvc_Cover_t {
	if s.Cubes.Items != 1 {
		panic("mvc: simple cover requires one cube")
	}
	i := Mvc_CoverBestLiteral(c, s.Cubes.Head)
	if i < 0 {
		panic("mvc: no repeated literal")
	}
	r := Mvc_CoverClone(c)
	x := Mvc_CubeAlloc(r)
	Mvc_CubeBitInsert(x, i)
	Mvc_CoverAddCubeTail(r, x)
	return r
}
func Mvc_CoverFirstCubeFirstLit(c *Mvc_Cover_t) int {
	if c.Cubes.Head != nil {
		for i := 0; i < c.NBits; i++ {
			if Mvc_CubeBitValue(c.Cubes.Head, i) {
				return i
			}
		}
	}
	return -1
}

// Mvc_CoverCountLiterals counts the literals of the first NBits of every
// cube, as mvcLits.c's per-literal loops do.
func Mvc_CoverCountLiterals(c *Mvc_Cover_t) int {
	words, rest := c.NBits>>5, c.NBits&31
	n := 0
	for x := c.Cubes.Head; x != nil; x = x.Next {
		for _, w := range x.Data[:words] {
			n += bits.OnesCount32(w)
		}
		if rest != 0 {
			n += bits.OnesCount32(x.Data[words] & (uint32(1)<<uint(rest) - 1))
		}
	}
	return n
}
func Mvc_CoverIsOneLiteral(c *Mvc_Cover_t) bool {
	return c.Cubes.Items == 1 && Mvc_CoverGetCubeSize(c.Cubes.Head) <= 1
}
