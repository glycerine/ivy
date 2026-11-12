package mvc

func Mvc_CubeAlloc(c *Mvc_Cover_t) *Mvc_Cube_t {
	return &Mvc_Cube_t{Data: make([]uint32, max(1, Mvc_CoverReadWordNum(c))), Unused: (32 - c.NBits%32) % 32}
}
func Mvc_CubeDup(c *Mvc_Cover_t, x *Mvc_Cube_t) *Mvc_Cube_t {
	r := Mvc_CubeAlloc(c)
	copy(r.Data, x.Data)
	r.Prime = x.Prime
	r.Essen = x.Essen
	r.Ones = x.Ones
	return r
}

// Cube storage is Go-owned; releasing it needs no explicit allocator action.
func Mvc_CubeFree(_ *Mvc_Cover_t, _ *Mvc_Cube_t)   {}
func Mvc_CubeReadNext(c *Mvc_Cube_t) *Mvc_Cube_t   { return c.Next }
func Mvc_CubeReadNextP(c *Mvc_Cube_t) **Mvc_Cube_t { return &c.Next }
func Mvc_CubeReadLast(c *Mvc_Cube_t) int           { return len(c.Data) - 1 }
func Mvc_CubeReadSize(c *Mvc_Cube_t) int           { return c.Ones }
func Mvc_CubeSetNext(c, n *Mvc_Cube_t)             { c.Next = n }
func Mvc_CubeSetLast(c *Mvc_Cube_t, n int) {
	if n < 0 {
		panic("mvc: negative last word")
	}
	if n+1 <= cap(c.Data) {
		c.Data = c.Data[:n+1]
	} else {
		c.Data = append(c.Data, make([]uint32, n+1-len(c.Data))...)
	}
}
func Mvc_CubeSetSize(c *Mvc_Cube_t, n int) { c.Ones = n }
func Mvc_CubeWhichWord(bit int) int        { return bit >> 5 }
func Mvc_CubeWhichBit(bit int) int         { return bit & 31 }
func Mvc_CubeBitValue(c *Mvc_Cube_t, bit int) bool {
	return c.Data[bit>>5]&(uint32(1)<<uint(bit&31)) != 0
}
func Mvc_CubeBitInsert(c *Mvc_Cube_t, bit int)  { c.Data[bit>>5] |= uint32(1) << uint(bit&31) }
func Mvc_CubeBitRemove(c *Mvc_Cube_t, bit int)  { c.Data[bit>>5] &^= uint32(1) << uint(bit&31) }
func Mvc_CubeVarValue(c *Mvc_Cube_t, v int) int { return int(c.Data[v>>4] >> uint(2*(v&15)) & 3) }
func Mvc_CubeBitClean(c *Mvc_Cube_t)            { clear(c.Data) }
func cubeMask(c *Mvc_Cube_t) uint32 {
	if c.Unused == 0 {
		return BITS_FULL
	}
	return BITS_FULL >> uint(c.Unused)
}
func Mvc_CubeBitFill(c *Mvc_Cube_t) {
	for i := range c.Data {
		c.Data[i] = BITS_FULL
	}
	c.Data[len(c.Data)-1] &= cubeMask(c)
}
func Mvc_CubeBitNot(c *Mvc_Cube_t) {
	for i := range c.Data {
		c.Data[i] = ^c.Data[i]
	}
	c.Data[len(c.Data)-1] &= cubeMask(c)
}
func Mvc_CubeBitCopy(r, c *Mvc_Cube_t) { copy(r.Data, c.Data) }
func Mvc_CubeBitOr(r, a, b *Mvc_Cube_t) {
	for i := range r.Data {
		r.Data[i] = a.Data[i] | b.Data[i]
	}
}
func Mvc_CubeBitAnd(r, a, b *Mvc_Cube_t) {
	for i := range r.Data {
		r.Data[i] = a.Data[i] & b.Data[i]
	}
}
func Mvc_CubeBitExor(r, a, b *Mvc_Cube_t) {
	for i := range r.Data {
		r.Data[i] = a.Data[i] ^ b.Data[i]
	}
}
func Mvc_CubeBitSharp(r, a, b *Mvc_Cube_t) {
	for i := range r.Data {
		r.Data[i] = a.Data[i] &^ b.Data[i]
	}
}
func Mvc_CubeBitEmpty(c *Mvc_Cube_t) bool {
	for _, w := range c.Data {
		if w != 0 {
			return false
		}
	}
	return true
}
func Mvc_CubeBitEqual(a, b *Mvc_Cube_t) bool {
	for i := range a.Data {
		if a.Data[i] != b.Data[i] {
			return false
		}
	}
	return true
}

// The historical BitLess/More macros compare unsigned word values; subset
// implication is expressed separately by BitNotImpl.
func Mvc_CubeBitLess(a, b *Mvc_Cube_t) bool {
	if len(a.Data) == 1 {
		return a.Data[0] < b.Data[0]
	}
	if len(a.Data) == 2 {
		return a.Data[0] < b.Data[0] || a.Data[0] == b.Data[0] && a.Data[1] < b.Data[1]
	}
	for i := range a.Data {
		if a.Data[i] >= b.Data[i] {
			return false
		}
	}
	return true
}
func Mvc_CubeBitMore(a, b *Mvc_Cube_t) bool { return Mvc_CubeBitLess(b, a) }
func Mvc_CubeBitCleanUnused(c *Mvc_Cube_t)  { c.Data[len(c.Data)-1] &= cubeMask(c) }
func Mvc_Cube1Words(c *Mvc_Cube_t) bool     { return len(c.Data)-1 == 0 }
func Mvc_Cube2Words(c *Mvc_Cube_t) bool     { return len(c.Data)-1 == 1 }
func Mvc_CubeNWords(c *Mvc_Cube_t) bool     { return len(c.Data)-1 > 1 }
func Mvc_CubeBitNotImpl(a, b *Mvc_Cube_t) bool {
	for i := range a.Data {
		if a.Data[i]&^b.Data[i] != 0 {
			return true
		}
	}
	return false
}
func Mvc_CubeBitDisjoint(a, b *Mvc_Cube_t) bool {
	for i := range a.Data {
		if a.Data[i]&b.Data[i] != 0 {
			return false
		}
	}
	return true
}
func Mvc_CubeBitEqualUnderMask(a, b, m *Mvc_Cube_t) bool {
	for i := range a.Data {
		if (a.Data[i]^b.Data[i])&m.Data[i] != 0 {
			return false
		}
	}
	return true
}
func Mvc_CubeBitEqualOutsideMask(a, b, m *Mvc_Cube_t) bool {
	for i := range a.Data {
		if (a.Data[i]^b.Data[i])&^m.Data[i] != 0 {
			return false
		}
	}
	return true
}
func Mvc_CubeBitIntersectUnderMask(a, b, m *Mvc_Cube_t) bool {
	for i := range a.Data {
		if a.Data[i]&b.Data[i]&m.Data[i] != 0 {
			return true
		}
	}
	return false
}
func Mvc_CubeBitNotImplUnderMask(a, b, m *Mvc_Cube_t) bool {
	for i := range a.Data {
		if a.Data[i]&^b.Data[i]&m.Data[i] != 0 {
			return true
		}
	}
	return false
}
func Mvc_CubeBitRemoveDcs(c *Mvc_Cube_t) {
	for i, w := range c.Data {
		m := (w ^ (w >> 1)) & BITS_DISJOINT
		m |= m << 1
		c.Data[i] &= m
	}
}
