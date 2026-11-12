package mvc

import "math/bits"

func Mvc_CoverSupport(c *Mvc_Cover_t, out *Mvc_Cube_t) {
	Mvc_CubeBitClean(out)
	for x := c.Cubes.Head; x != nil; x = x.Next {
		Mvc_CubeBitOr(out, out, x)
	}
}
func Mvc_CoverSupportAnd(c *Mvc_Cover_t, out *Mvc_Cube_t) {
	Mvc_CubeBitFill(out)
	for x := c.Cubes.Head; x != nil; x = x.Next {
		Mvc_CubeBitAnd(out, out, x)
	}
}
func Mvc_CoverSupportSizeBinary(c *Mvc_Cover_t) int {
	x := Mvc_CubeAlloc(c)
	Mvc_CoverSupportAnd(c, x)
	n := 0
	for i := 0; i < c.NBits/2; i++ {
		if Mvc_CubeVarValue(x, i) != 3 {
			n++
		}
	}
	return n
}
func Mvc_CoverSupportVarBelongs(c *Mvc_Cover_t, i int) bool {
	x := Mvc_CubeAlloc(c)
	Mvc_CoverSupportAnd(c, x)
	return Mvc_CubeVarValue(x, i) != 3
}
func Mvc_CoverCommonCube(c *Mvc_Cover_t, out *Mvc_Cube_t) { Mvc_CoverSupportAnd(c, out) }
func Mvc_CoverIsCubeFree(c *Mvc_Cover_t) bool {
	Mvc_CoverAllocateMask(c)
	Mvc_CoverCommonCube(c, c.Mask)
	return Mvc_CubeBitEmpty(c.Mask)
}
func Mvc_CoverMakeCubeFree(c *Mvc_Cover_t) {
	Mvc_CoverAllocateMask(c)
	Mvc_CoverCommonCube(c, c.Mask)
	for x := c.Cubes.Head; x != nil; x = x.Next {
		Mvc_CubeBitSharp(x, x, c.Mask)
	}
}
func Mvc_CoverCommonCubeCover(c *Mvc_Cover_t) *Mvc_Cover_t {
	r := Mvc_CoverClone(c)
	x := Mvc_CubeAlloc(r)
	Mvc_CoverCommonCube(c, x)
	Mvc_CoverAddCubeTail(r, x)
	return r
}
func Mvc_CoverCheckSuppContainment(a, b *Mvc_Cover_t) bool {
	x, y := Mvc_CubeAlloc(a), Mvc_CubeAlloc(b)
	Mvc_CoverSupport(a, x)
	Mvc_CoverSupport(b, y)
	return !Mvc_CubeBitNotImpl(y, x)
}
func Mvc_CoverGetCubeSize(c *Mvc_Cube_t) int {
	n := 0
	for _, w := range c.Data {
		n += bits.OnesCount32(w)
	}
	return n
}
func Mvc_CoverSetCubeSizes(c *Mvc_Cover_t) bool {
	for x := c.Cubes.Head; x != nil; x = x.Next {
		x.Ones = Mvc_CoverGetCubeSize(x)
	}
	return true
}
func Mvc_CoverCountCubePairDiffs(c *Mvc_Cover_t, diffs []byte) int {
	n := 0
	for x := c.Cubes.Head; x != nil; x = x.Next {
		for y := x.Next; y != nil; y = y.Next {
			k := 0
			for i, w := range x.Data {
				k += bits.OnesCount32(w ^ y.Data[i])
			}
			diffs[n] = byte(k)
			n++
		}
	}
	return 1
}
func Mvc_CoverRemap(c *Mvc_Cover_t, columns []int, n int) *Mvc_Cover_t {
	r := Mvc_CoverAlloc(c.Mem, n)
	for x := c.Cubes.Head; x != nil; x = x.Next {
		z := Mvc_CubeAlloc(r)
		Mvc_CubeBitFill(z)
		for i, old := range columns[:n] {
			if old >= 0 && !Mvc_CubeBitValue(x, old) {
				Mvc_CubeBitRemove(z, i)
			}
		}
		Mvc_CoverAddCubeTail(r, z)
	}
	return r
}
func Mvc_CoverCopyColumn(a, b *Mvc_Cover_t, old, new int) {
	for x, y := a.Cubes.Head, b.Cubes.Head; x != nil; x, y = x.Next, y.Next {
		if Mvc_CubeBitValue(x, old) {
			Mvc_CubeBitInsert(y, new)
		} else {
			Mvc_CubeBitRemove(y, new)
		}
	}
}
func Mvc_CoverInverse(c *Mvc_Cover_t) {
	for x := c.Cubes.Head; x != nil; x = x.Next {
		Mvc_CubeBitNot(x)
	}
}
func Mvc_CoverRemoveDontCareLits(c *Mvc_Cover_t) *Mvc_Cover_t {
	r := Mvc_CoverDup(c)
	for x := r.Cubes.Head; x != nil; x = x.Next {
		Mvc_CubeBitRemoveDcs(x)
	}
	return r
}
func Mvc_CoverCofactor(c *Mvc_Cover_t, value, other int) *Mvc_Cover_t {
	r := Mvc_CoverClone(c)
	for x := c.Cubes.Head; x != nil; x = x.Next {
		if Mvc_CubeBitValue(x, value) {
			y := Mvc_CubeDup(r, x)
			Mvc_CubeBitInsert(y, other)
			Mvc_CoverAddCubeTail(r, y)
		}
	}
	return r
}
func setBit(c *Mvc_Cube_t, i int, value bool) {
	if value {
		Mvc_CubeBitInsert(c, i)
	} else {
		Mvc_CubeBitRemove(c, i)
	}
}
func Mvc_CoverFlipVar(c *Mvc_Cover_t, a, b int) *Mvc_Cover_t {
	r := Mvc_CoverDup(c)
	for x := r.Cubes.Head; x != nil; x = x.Next {
		va, vb := Mvc_CubeBitValue(x, a), Mvc_CubeBitValue(x, b)
		setBit(x, a, vb)
		setBit(x, b, va)
	}
	return r
}
func Mvc_CoverUnivQuantify(c *Mvc_Cover_t, a0, a1, b0, b1 int) *Mvc_Cover_t {
	r := Mvc_CoverClone(c)
	for x := c.Cubes.Head; x != nil; x = x.Next {
		va0, va1, vb0, vb1 := Mvc_CubeBitValue(x, a0), Mvc_CubeBitValue(x, a1), Mvc_CubeBitValue(x, b0), Mvc_CubeBitValue(x, b1)
		if va0 != vb0 && va1 != vb1 {
			continue
		}
		z := Mvc_CubeDup(r, x)
		setBit(z, a0, va0 && vb0)
		setBit(z, a1, va1 && vb1)
		setBit(z, b0, true)
		setBit(z, b1, true)
		Mvc_CoverAddCubeTail(r, z)
	}
	return r
}
func Mvc_CoverTranspose(c *Mvc_Cover_t) *Mvc_Cover_t {
	r := Mvc_CoverAlloc(c.Mem, c.Cubes.Items)
	for bit := 0; bit < c.NBits; bit++ {
		z := Mvc_CubeAlloc(r)
		i := 0
		for x := c.Cubes.Head; x != nil; x = x.Next {
			if Mvc_CubeBitValue(x, bit) {
				Mvc_CubeBitInsert(z, i)
			}
			i++
		}
		Mvc_CoverAddCubeTail(r, z)
	}
	return r
}
func Mvc_UtilsCheckUnusedZeros(c *Mvc_Cover_t) bool {
	for x := c.Cubes.Head; x != nil; x = x.Next {
		if x.Data[len(x.Data)-1]&^cubeMask(x) != 0 {
			return false
		}
	}
	return true
}
func Mvc_CoverCopyAndAppendCubes(a, b *Mvc_Cover_t) {
	for x := b.Cubes.Head; x != nil; x = x.Next {
		Mvc_CoverAddDupCubeTail(a, x)
	}
}
func Mvc_CoverRemoveCubes(c *Mvc_Cover_t) { Mvc_CoverMakeEmpty(c) }
