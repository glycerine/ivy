package mvc

func literalCounts(c *Mvc_Cover_t) []int {
	counts := make([]int, c.NBits)
	for x := c.Cubes.Head; x != nil; x = x.Next {
		for i := range counts {
			if Mvc_CubeBitValue(x, i) {
				counts[i]++
			}
		}
	}
	return counts
}
func Mvc_CoverAnyLiteral(c *Mvc_Cover_t, m *Mvc_Cube_t) int {
	counts := literalCounts(c)
	for i := len(counts) - 1; i >= 0; i-- {
		if counts[i] > 1 && (m == nil || Mvc_CubeBitValue(m, i)) {
			return i
		}
	}
	return -1
}
func Mvc_CoverBestLiteral(c *Mvc_Cover_t, m *Mvc_Cube_t) int {
	best, n := -1, 1
	for i, k := range literalCounts(c) {
		if k > n && (m == nil || Mvc_CubeBitValue(m, i)) {
			best, n = i, k
		}
	}
	return best
}
func Mvc_CoverWorstLiteral(c *Mvc_Cover_t, m *Mvc_Cube_t) int {
	best, n := -1, int(^uint(0)>>1)
	for i, k := range literalCounts(c) {
		if k >= 2 && k < n && (m == nil || Mvc_CubeBitValue(m, i)) {
			best, n = i, k
		}
	}
	return best
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
func Mvc_CoverCountLiterals(c *Mvc_Cover_t) int {
	n := 0
	for _, k := range literalCounts(c) {
		n += k
	}
	return n
}
func Mvc_CoverIsOneLiteral(c *Mvc_Cover_t) bool {
	return c.Cubes.Items == 1 && Mvc_CoverGetCubeSize(c.Cubes.Head) <= 1
}
