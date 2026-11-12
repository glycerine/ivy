package mvc

func Mvc_CoverBooleanOr(a, b *Mvc_Cover_t) *Mvc_Cover_t {
	r := Mvc_CoverDup(a)
	Mvc_CoverCopyAndAppendCubes(r, b)
	return r
}
