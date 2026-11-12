package mvc

func Mvc_CoverDivisor(c *Mvc_Cover_t) *Mvc_Cover_t {
	if c.Cubes.Items <= 1 || Mvc_CoverAnyLiteral(c, nil) < 0 {
		return nil
	}
	r := Mvc_CoverDup(c)
	Mvc_CoverDivisorZeroKernel(r)
	return r
}
func Mvc_CoverDivisorZeroKernel(c *Mvc_Cover_t) {
	for {
		bit := Mvc_CoverWorstLiteral(c, nil)
		if bit < 0 {
			return
		}
		Mvc_CoverDivideByLiteralQuo(c, bit)
		Mvc_CoverMakeCubeFree(c)
	}
}
