package cba

// Positive fons are network objects; negative fons encode constant IDs.
// Zero is the undriven sentinel.
func Cba_FonIsReal(f int32) int32 {
	if f > 0 {
		return 1
	}
	return 0
}
func Cba_FonIsConst(f int32) int32 {
	if f < 0 {
		return 1
	}
	return 0
}
func Cba_FonConst(f int32) int32 {
	if f >= 0 {
		panic("cba: expected constant fon")
	}
	return -f - 1
}
func Cba_FonFromConst(id int32) int32 {
	if id < 0 {
		panic("cba: expected nonnegative constant ID")
	}
	return -id - 1
}
