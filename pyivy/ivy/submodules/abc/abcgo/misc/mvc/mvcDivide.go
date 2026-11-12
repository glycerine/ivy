package mvc

// Algebraic division groups dividend cubes by their literals outside the
// divisor's support, then matches every divisor cube inside each group.
func Mvc_CoverDivideInternal(c, d *Mvc_Cover_t) (q, r *Mvc_Cover_t) {
	if c.NBits != d.NBits {
		panic("mvc: incompatible covers")
	}
	if d.Cubes.Items == 0 {
		panic("mvc: empty divisor")
	}
	if d.Cubes.Items == 1 {
		return Mvc_CoverDivideByCube(c, d)
	}
	q, r = Mvc_CoverClone(c), Mvc_CoverClone(c)
	mask := Mvc_CubeAlloc(c)
	Mvc_CoverSupport(d, mask)
	Mvc_CoverList2Array(c)
	seen := make([]bool, len(c.Array))
	for i, first := range c.Array {
		if seen[i] {
			continue
		}
		var group []*Mvc_Cube_t
		for j, x := range c.Array {
			if !seen[j] && Mvc_CubeBitEqualOutsideMask(first, x, mask) {
				seen[j] = true
				group = append(group, x)
			}
		}
		used := make([]bool, len(group))
		success := true
		for div := d.Cubes.Head; div != nil; div = div.Next {
			match := -1
			for j, x := range group {
				if !used[j] && Mvc_CubeBitEqualUnderMask(x, div, mask) {
					match = j
					break
				}
			}
			if match < 0 {
				success = false
				break
			}
			used[match] = true
		}
		if success {
			x := Mvc_CubeDup(q, first)
			Mvc_CubeBitSharp(x, x, mask)
			Mvc_CoverAddCubeTail(q, x)
		}
		for j, x := range group {
			if !success || !used[j] {
				Mvc_CoverAddDupCubeTail(r, x)
			}
		}
	}
	return
}
func Mvc_CoverDivide(c, d *Mvc_Cover_t) (*Mvc_Cover_t, *Mvc_Cover_t) {
	if c.Cubes.Items < d.Cubes.Items || !Mvc_CoverCheckSuppContainment(c, d) {
		return nil, nil
	}
	return Mvc_CoverDivideInternal(c, d)
}
func Mvc_CoverDivideByCube(c, d *Mvc_Cover_t) (q, r *Mvc_Cover_t) {
	if d.Cubes.Items != 1 {
		panic("mvc: expected one divisor cube")
	}
	q, r = Mvc_CoverClone(c), Mvc_CoverClone(c)
	div := d.Cubes.Head
	for x := c.Cubes.Head; x != nil; x = x.Next {
		if !Mvc_CubeBitNotImpl(div, x) {
			y := Mvc_CubeDup(q, x)
			Mvc_CubeBitSharp(y, y, div)
			Mvc_CoverAddCubeTail(q, y)
		} else {
			Mvc_CoverAddDupCubeTail(r, x)
		}
	}
	return
}
func Mvc_CoverDivideByLiteral(c, d *Mvc_Cover_t) (*Mvc_Cover_t, *Mvc_Cover_t) {
	return Mvc_CoverDivideByCube(c, d)
}
func Mvc_CoverDivideByLiteralQuo(c *Mvc_Cover_t, bit int) {
	var prev *Mvc_Cube_t
	for x := c.Cubes.Head; x != nil; {
		next := x.Next
		if !Mvc_CubeBitValue(x, bit) {
			Mvc_CoverDeleteCube(c, prev, x)
		} else {
			Mvc_CubeBitRemove(x, bit)
			prev = x
		}
		x = next
	}
}
func Mvc_CoverVerifyDivision(c, d, q, r *Mvc_Cover_t) bool {
	p := Mvc_CoverAlgebraicMultiply(d, q)
	Mvc_CoverCopyAndAppendCubes(p, r)
	return Mvc_CoverAlgebraicEqual(c, p) && Mvc_CoverAlgebraicEqual(p, c)
}
