package mvc

func Mvc_CoverAlgebraicMultiply(a, b *Mvc_Cover_t) *Mvc_Cover_t {
	if a.NBits != b.NBits {
		panic("mvc: incompatible covers")
	}
	r := Mvc_CoverClone(a)
	for x := a.Cubes.Head; x != nil; x = x.Next {
		for y := b.Cubes.Head; y != nil; y = y.Next {
			z := Mvc_CubeAlloc(r)
			Mvc_CubeBitOr(z, x, y)
			Mvc_CoverAddCubeTail(r, z)
		}
	}
	return r
}
func Mvc_CoverAlgebraicSubtract(a, b *Mvc_Cover_t) *Mvc_Cover_t {
	r := Mvc_CoverClone(a)
	for x := a.Cubes.Head; x != nil; x = x.Next {
		found := false
		for y := b.Cubes.Head; y != nil; y = y.Next {
			if Mvc_CubeBitEqual(x, y) {
				found = true
				break
			}
		}
		if !found {
			Mvc_CoverAddDupCubeTail(r, x)
		}
	}
	return r
}
func Mvc_CoverAlgebraicEqual(a, b *Mvc_Cover_t) bool {
	return Mvc_CoverIsEmpty(Mvc_CoverAlgebraicSubtract(a, b))
}
