package mvc

import "sort"

func Mvc_CoverSort(c *Mvc_Cover_t, m *Mvc_Cube_t, compare func(*Mvc_Cube_t, *Mvc_Cube_t, *Mvc_Cube_t) int) {
	Mvc_CoverList2Array(c)
	sort.SliceStable(c.Array, func(i, j int) bool { return compare(c.Array[i], c.Array[j], m) < 0 })
	Mvc_CoverArray2List(c)
}
