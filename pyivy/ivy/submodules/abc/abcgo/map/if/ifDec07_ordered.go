package if_

import "math/bits"

// Ordered truth-word permutations and configuration derivation from ifDec07.c.
// Keep the source's evolving variable order and disjoint-before-shared policy.
var dec07SwapMasks = [5][3]uint64{
	{0x9999999999999999, 0x2222222222222222, 0x4444444444444444},
	{0xc3c3c3c3c3c3c3c3, 0x0c0c0c0c0c0c0c0c, 0x3030303030303030},
	{0xf00ff00ff00ff00f, 0x00f000f000f000f0, 0x0f000f000f000f00},
	{0xff0000ffff0000ff, 0x0000ff000000ff00, 0x00ff000000ff0000},
	{0xffff00000000ffff, 0x00000000ffff0000, 0x0000ffff00000000},
}

func dec07Swap6(t uint64, v int) uint64 {
	m := dec07SwapMasks[v]
	return t&m[0] | (t&m[1])<<uint(1<<v) | (t&m[2])>>uint(1<<v)
}
func dec07Move6(t uint64, v, target int, p2v, v2p *[7]int) uint64 {
	assert(v2p[v] >= target, "ordered decomposition variable moves left")
	for v2p[v] != target {
		a, b := v2p[v]-1, v2p[v]
		t = dec07Swap6(t, a)
		v2p[p2v[a]]++
		v2p[p2v[b]]--
		p2v[a], p2v[b] = p2v[b], p2v[a]
	}
	return t
}
func dec07Move7(t *[2]uint64, v, target int, p2v, v2p *[7]int) {
	assert(v2p[v] >= target, "ordered seven-input variable moves left")
	for v2p[v] != target {
		a, b := v2p[v]-1, v2p[v]
		if a == 5 {
			upper := t[0] >> 32
			t[0] = t[0]&0xffffffff | (t[1]&0xffffffff)<<32
			t[1] = t[1]&0xffffffff00000000 | upper
		} else {
			t[0] = dec07Swap6(t[0], a)
			t[1] = dec07Swap6(t[1], a)
		}
		v2p[p2v[a]]++
		v2p[p2v[b]]--
		p2v[a], p2v[b] = p2v[b], p2v[a]
	}
}
func dec07Permutations() (p2v, v2p [7]int) {
	for i := range p2v {
		p2v[i] = i
		v2p[i] = i
	}
	return
}
func dec07Count6(t uint64) int {
	var mask uint16
	for i := uint(0); i < 16; i++ {
		mask |= 1 << uint(t>>(4*i)&15)
	}
	return bits.OnesCount16(mask)
}
func dec07Count7(t [2]uint64) int {
	first := byte(t[0])
	second := -1
	for i := 1; i < 16; i++ {
		value := byte(t[i/8] >> uint(8*(i%8)))
		if value == first {
			continue
		}
		if second < 0 {
			second = int(value)
		} else if int(value) != second {
			return 3
		}
	}
	// Source requires complete support and asserts a second cofactor exists.
	// Retain safe Go handling of reduced-support inputs outside that precondition.
	if second < 0 {
		return 1
	}
	return 2
}
func dec07Derive(t [2]uint64, width uint) (mask, zero, one uint64) {
	lowMask := uint64(1)<<width - 1
	zero = t[0] & lowMask
	one = zero
	for i := uint(1); i < 16; i++ {
		offset := i * width
		value := t[offset/64] >> (offset % 64) & lowMask
		if value != zero {
			one = value
			mask |= 1 << i
		}
	}
	return
}
func dec07Cofactor6(t uint64, v int, one bool) uint64 {
	mask := truth6Variable(int32(v))
	if one {
		return t&mask | (t&mask)>>uint(1<<v)
	}
	return t&^mask | (t&^mask)<<uint(1<<v)
}
func dec07Disjoint6(t uint64, p2v [7]int) uint64 {
	z, c0, c1 := dec07Derive([2]uint64{t}, 4)
	for i := 0; i < 4; i++ {
		z |= uint64(p2v[i+2]) << uint(16+4*i)
	}
	upper := c1<<4 | c0
	z |= upper<<32 | upper<<40
	return z | uint64(p2v[0])<<48 | uint64(p2v[1])<<52 | uint64(7)<<56
}
func dec07Shared6(t uint64, shared int, p2v, v2p [7]int) uint64 {
	for i := shared; i < 5; i++ {
		t = dec07Swap6(t, i)
		v2p[p2v[i]]++
		v2p[p2v[i+1]]--
		p2v[i], p2v[i+1] = p2v[i+1], p2v[i]
	}
	c0, c1 := dec07Cofactor6(t, 5, false), dec07Cofactor6(t, 5, true)
	h0, lo0, hi0 := dec07Derive([2]uint64{c0}, 4)
	h1, lo1, hi1 := dec07Derive([2]uint64{c1}, 4)
	z := (h1&255)<<8 | h0&255
	for i := 0; i < 4; i++ {
		z |= uint64(p2v[i+2]) << uint(16+4*i)
	}
	z |= (hi0<<4|lo0)<<32 | (hi1<<4|lo1)<<40
	return z | uint64(p2v[0])<<48 | uint64(p2v[1])<<52 | uint64(7)<<56 | uint64(p2v[5])<<60
}
func If_Dec6Perform(t uint64, derive int32) uint64 {
	p2v, v2p := dec07Permutations()
	var result uint64
	for v := 0; v < 6; v++ {
		for u := v + 1; u < 6; u++ {
			t = dec07Move6(t, v, 0, &p2v, &v2p)
			t = dec07Move6(t, u, 1, &p2v, &v2p)
			count := dec07Count6(t)
			if count <= 2 {
				if derive == 0 {
					return 1
				}
				return dec07Disjoint6(t, p2v)
			}
			if result == 0 && (count == 3 || count == 4) {
				for x := 2; x < 6; x++ {
					if dec07Count6(dec07Cofactor6(t, x, false)) <= 2 && dec07Count6(dec07Cofactor6(t, x, true)) <= 2 {
						if derive == 0 {
							result = 1
						} else {
							result = dec07Shared6(t, x, p2v, v2p)
						}
						break
					}
				}
			}
		}
	}
	return result
}
func If_Dec7Perform(input []uint64, derive int32) uint64 {
	decCheck(input, 7)
	t := [2]uint64{input[0], input[1]}
	p2v, v2p := dec07Permutations()
	for v := 0; v < 7; v++ {
		for u := v + 1; u < 7; u++ {
			for y := u + 1; y < 7; y++ {
				dec07Move7(&t, v, 0, &p2v, &v2p)
				dec07Move7(&t, u, 1, &p2v, &v2p)
				dec07Move7(&t, y, 2, &p2v, &v2p)
				if dec07Count7(t) <= 2 {
					if derive == 0 {
						return 1
					}
					z, c0, c1 := dec07Derive(t, 8)
					for i := 0; i < 4; i++ {
						z |= uint64(p2v[i+3]) << uint(16+4*i)
					}
					z |= (c1<<8 | c0) << 32
					for i := 0; i < 3; i++ {
						z |= uint64(p2v[i]) << uint(48+4*i)
					}
					return z | uint64(7)<<60
				}
			}
		}
	}
	return 0
}

// Pack the source's five-input C/D truth and pin fields, dropping unused shared
// variables to trailing sentinel6 slots. The root's bound-output pin is7.
func dec07Pack5(t uint64, first, second, x, y int) uint32 {
	has2 := dec07Cofactor6(t, 2, false) != dec07Cofactor6(t, 2, true)
	has3 := dec07Cofactor6(t, 3, false) != dec07Cofactor6(t, 3, true)
	a, b := 6, 6
	switch {
	case has2 && has3:
		a, b = x, y
	case has2:
		a = x
	case has3:
		t = dec07Swap6(t, 2)
		a = y
	}
	return uint32(t&0xffff) | uint32(first)<<16 | uint32(second)<<20 | uint32(a)<<24 | uint32(b)<<28
}
func dec07FiveGroupMask(t uint64, x, y, m int) uint32 {
	var mask uint32
	for i := 0; i < 16; i++ {
		if i>>x&1 == m&1 && i>>y&1 == m>>1&1 {
			mask |= 1 << uint(t>>uint(2*i)&3)
		}
	}
	return mask
}
func dec07FiveCofactors(t uint64, x, y int, p2v [7]int, derive int32) uint64 {
	for m := 0; m < 4; m++ {
		if bits.OnesCount32(dec07FiveGroupMask(t, x, y, m)) > 2 {
			return 0
		}
	}
	if derive == 0 {
		return 1
	}
	z1, z2 := -1, -1
	for v := 0; v < 4; v++ {
		if v != x && v != y {
			if z1 < 0 {
				z1 = v
			} else {
				z2 = v
			}
		}
	}
	f := [4]uint64{0, ^truth6Variable(0), truth6Variable(0), ^uint64(0)}
	var c, d [4]uint64
	for m := 0; m < 4; m++ {
		mask := dec07FiveGroupMask(t, x, y, m)
		b0 := bits.TrailingZeros32(mask)
		if bits.OnesCount32(mask) == 1 {
			c[m] = f[b0]
			d[m] = ^uint64(0)
			continue
		}
		b1 := bits.TrailingZeros32(mask &^ (1 << uint(b0)))
		c[m] = f[b1]&truth6Variable(1) | f[b0]&^truth6Variable(1)
		for i := 0; i < 16; i++ {
			if i>>x&1 == m&1 && i>>y&1 == m>>1&1 && int(t>>uint(2*i)&3) == b1 {
				d[m] |= 1 << uint((i>>z2&1)<<1|i>>z1&1)
			}
		}
		d[m] |= d[m] << 4
		d[m] |= d[m] << 8
		d[m] |= d[m] << 16
		d[m] |= d[m] << 32
	}
	mux := func(s, a, b uint64) uint64 { return s&a | ^s&b }
	ct := mux(truth6Variable(3), mux(truth6Variable(2), c[3], c[2]), mux(truth6Variable(2), c[1], c[0]))
	dt := mux(truth6Variable(3), mux(truth6Variable(2), d[3], d[2]), mux(truth6Variable(2), d[1], d[0]))
	lower := dec07Pack5(dt, p2v[z1+1], p2v[z2+1], p2v[x+1], p2v[y+1])
	upper := dec07Pack5(ct, p2v[0], 7, p2v[x+1], p2v[y+1])
	return uint64(lower) | uint64(upper)<<32
}
func If_Dec5Perform(t uint64, derive int32) uint64 {
	original := t
	p2v, v2p := dec07Permutations()
	for v := 0; v < 5; v++ {
		t = dec07Move6(t, v, 0, &p2v, &v2p)
		for x := 0; x < 4; x++ {
			for y := x + 1; y < 4; y++ {
				if z := dec07FiveCofactors(t, x, y, p2v, derive); z != 0 {
					if derive != 0 {
						assert(If_Dec6Truth(z) == original, "source five-input configuration verifies")
					}
					return z
				}
			}
		}
	}
	return 0
}
