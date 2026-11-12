package mvc

func Mvc_CubeCompareInt(a, b, m *Mvc_Cube_t) int { return compareWords(a, b, m, 0) }
func compareWords(a, b, m *Mvc_Cube_t, mode int) int {
	for i := len(a.Data) - 1; i >= 0; i-- {
		x, y := a.Data[i], b.Data[i]
		if mode == 1 {
			x &= m.Data[i]
			y &= m.Data[i]
		}
		if mode == 2 {
			x &^= m.Data[i]
			y &^= m.Data[i]
		}
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return 0
}
func Mvc_CubeCompareSizeAndInt(a, b, m *Mvc_Cube_t) int {
	if a.Ones > b.Ones {
		return -1
	}
	if a.Ones < b.Ones {
		return 1
	}
	return Mvc_CubeCompareInt(a, b, m)
}
func Mvc_CubeCompareIntUnderMask(a, b, m *Mvc_Cube_t) int   { return compareWords(a, b, m, 1) }
func Mvc_CubeCompareIntOutsideMask(a, b, m *Mvc_Cube_t) int { return compareWords(a, b, m, 2) }
func Mvc_CubeCompareIntOutsideAndUnderMask(a, b, m *Mvc_Cube_t) int {
	if c := compareWords(a, b, m, 2); c != 0 {
		return c
	}
	return compareWords(a, b, m, 1)
}
