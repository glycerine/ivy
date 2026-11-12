// Port of ABC FXCH, Bruno Schmitt, UFRGS (2016); original FxchDiv.c.
package fxch

import (
	"encoding/binary"
	"fmt"
	"sort"
)

func Fxch_DivNormalize(l []int32) int32 {
	if len(l) != 4 {
		panic("FXCH MUX divisor must contain four literals")
	}
	var a, b, x, y int32
	found := false
	for _, ij := range [][2]int{{0, 1}, {1, 2}, {2, 3}} {
		i, j := ij[0], ij[1]
		if l[i]&1 == l[j]&1 || l[i]>>2 != l[j]>>2 {
			continue
		}
		var other []int
		for k := 0; k < 4; k++ {
			if k != i && k != j {
				other = append(other, k)
			}
		}
		u, v := other[0], other[1]
		if l[u]&1 == l[v]&1 {
			return -1
		}
		a, b = l[i]>>1, l[j]>>1
		if l[i]&1 == l[u]&1 {
			x, y = l[u]>>1, l[v]>>1
		} else {
			x, y = l[v]>>1, l[u]>>1
		}
		found = true
		break
	}
	if !found {
		return -1
	}
	if a&1 != 0 {
		a, b = b, a
		x, y = y, x
	}
	ret := int32(0)
	if x&1 != 0 {
		x ^= 1
		y ^= 1
		ret = 1
	}
	l[0], l[1], l[2], l[3] = a*2, b*2+1, x*2, y*2+1
	return ret
}
func Fxch_DivCreate(p *Fxch_Man_t, a, b *Fxch_SubCube_t) int32 {
	p.CubeFree = nil
	c, d := p.Cubes[a.ICube], p.Cubes[b.ICube]
	if a.ILit0 == 0 || b.ILit0 == 0 {
		return -1
	}
	x, y := c[a.ILit0], d[b.ILit0]
	if a.ILit1 == 0 && b.ILit1 == 0 {
		p.CubeFree = []int32{x, y}
	} else if a.ILit1 > 0 && b.ILit1 > 0 {
		u, v := c[a.ILit1], d[b.ILit1]
		if x < y {
			p.CubeFree = []int32{x * 2, y*2 + 1, u * 2, v*2 + 1}
		} else {
			p.CubeFree = []int32{y * 2, x*2 + 1, v * 2, u*2 + 1}
		}
		if Fxch_DivNormalize(p.CubeFree) < 0 {
			return -1
		}
	} else {
		if a.ILit1 > 0 {
			z := c[a.ILit1]
			p.CubeFree = []int32{y}
			if x == y^1 {
				p.CubeFree = append(p.CubeFree, z)
			} else if z == y^1 {
				p.CubeFree = append(p.CubeFree, x)
			}
		} else {
			z := d[b.ILit1]
			p.CubeFree = []int32{x}
			if y == x^1 {
				p.CubeFree = append(p.CubeFree, z)
			} else if z == x^1 {
				p.CubeFree = append(p.CubeFree, y)
			}
		}
	}
	if len(p.CubeFree) == 0 || len(p.CubeFree)%2 != 0 {
		return -1
	}
	if len(p.CubeFree) == 2 {
		sort.Slice(p.CubeFree, func(i, j int) bool { return p.CubeFree[i] < p.CubeFree[j] })
		p.CubeFree[0] *= 2
		p.CubeFree[1] = p.CubeFree[1]*2 + 1
	}
	n := len(c)
	if len(d) < n {
		n = len(d)
	}
	return int32(n - len(p.CubeFree)/2 - 1)
}
func divKey(a []int32) string {
	b := make([]byte, len(a)*4)
	for i, v := range a {
		binary.LittleEndian.PutUint32(b[i*4:], uint32(v))
	}
	return string(b)
}
func Fxch_DivAdd(p *Fxch_Man_t, update, single, base int32) int32 {
	key := divKey(p.CubeFree)
	id, ok := p.divIndex[key]
	if !ok {
		id = int32(len(p.divisors))
		p.divIndex[key] = id
		penalty := 0.0009
		if single != 0 {
			penalty = 0.001
		}
		weight := float32(-float64(len(p.CubeFree)) + 0.9 - penalty*float64(Fxch_ManComputeLevelDiv(p, p.CubeFree)))
		p.divisors = append(p.divisors, divisor{literals: copyInts(p.CubeFree), weight: weight})
	}
	if single != 0 {
		p.divisors[id].weight++
	} else {
		p.divisors[id].weight += float32(base + int32(len(p.CubeFree)) - 1)
	}
	return id
}
func Fxch_DivRemove(p *Fxch_Man_t, update, single, base int32) int32 {
	id, ok := p.divIndex[divKey(p.CubeFree)]
	if !ok {
		panic("removing absent FXCH divisor")
	}
	if single != 0 {
		p.divisors[id].weight--
	} else {
		p.divisors[id].weight -= float32(base + int32(len(p.CubeFree)) - 1)
	}
	return id
}
func Fxch_DivSepareteCubes(div []int32, a, b *[]int32) {
	for _, l := range div {
		if l&1 == 0 {
			*a = append(*a, l>>1)
		} else {
			*b = append(*b, l>>1)
		}
	}
	if len(div) == 4 {
		sort.Slice((*a)[1:], func(i, j int) bool { return (*a)[i+1] < (*a)[j+1] })
		sort.Slice((*b)[1:], func(i, j int) bool { return (*b)[i+1] < (*b)[j+1] })
	}
}
func removeLit(a *[]int32, l int32) int32 {
	for i, v := range *a {
		if i > 0 && v == l {
			*a = append((*a)[:i], (*a)[i+1:]...)
			return 1
		}
	}
	return 0
}
func Fxch_DivRemoveLits(a, b *[]int32, div []int32, compl *int32) int32 {
	var pos, neg, count int32
	for _, l := range div {
		if (l>>1)&1 != 0 {
			neg += removeLit(a, l>>1)
		} else {
			pos += removeLit(a, l>>1)
		}
	}
	for _, l := range div {
		count += removeLit(b, l>>1)
	}
	if len(div) == 2 {
		for _, l := range div {
			removeLit(a, (l>>1)^1)
			removeLit(b, (l>>1)^1)
		}
	}
	ret := pos + neg + count
	if len(div) == 4 {
		x, y, u, v := div[0]>>1, div[1]>>1, div[2]>>1, div[3]>>1
		if x == y^1 && u == v^1 && neg == 1 {
			*compl = 1
		}
		if x == y^1 && ret == 2 {
			*compl = 1
			for i, l := range div {
				lit := l >> 1
				if i > 1 {
					lit ^= 1
				}
				ret += removeLit(a, lit)
				ret += removeLit(b, lit)
			}
		}
	}
	return ret
}
func Fxch_DivIsNotConstant1(d []int32) int32 { return boolInt(len(d) != 2 || (d[0]>>1) != (d[1]>>1)^1) }
func Fxch_DivPrint(p *Fxch_Man_t, id int32) {
	fmt.Fprintf(p.writer(), "Div %7d : Weight %12.5f %v\n", id, p.divisors[id].weight, p.divisors[id].literals)
}
