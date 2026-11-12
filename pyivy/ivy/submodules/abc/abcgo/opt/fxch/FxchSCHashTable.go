// Port of ABC FXCH, Bruno Schmitt, UFRGS (2016); original FxchSCHashTable.c.
package fxch

import (
	"encoding/binary"
	"fmt"
	"math/bits"
	"slices"
	"unsafe"
)

func MurmurHash3_x86_32(key []byte, seed uint32) uint32 {
	h := seed
	for i := 0; i+4 <= len(key); i += 4 {
		k := binary.LittleEndian.Uint32(key[i:])
		k *= 0xcc9e2d51
		k = bits.RotateLeft32(k, 15)
		k *= 0x1b873593
		h ^= k
		h = bits.RotateLeft32(h, 13)
		h = h*5 + 0xe6546b64
	}
	tail := key[len(key)&^3:]
	var k uint32
	for i, v := range tail {
		k ^= uint32(v) << uint(8*i)
	}
	if len(tail) > 0 {
		k *= 0xcc9e2d51
		k = bits.RotateLeft32(k, 15)
		k *= 0x1b873593
		h ^= k
	}
	h ^= uint32(len(key))
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	return h ^ (h >> 16)
}
func Fxch_SCHashTableCreate(p *Fxch_Man_t, n int32) *Fxch_SCHashTable_t {
	return &Fxch_SCHashTable_t{man: p, bins: make(map[uint32][]Fxch_SubCube_t)}
}
func Fxch_SCHashTableDelete(t *Fxch_SCHashTable_t) { t.bins = nil }
func subcube(c []int32, s *Fxch_SubCube_t) []int32 {
	a := make([]int32, 0, len(c))
	for i, l := range c {
		if i > 0 && (uint32(i) == s.ILit0 || s.ILit1 > 0 && uint32(i) == s.ILit1) {
			continue
		}
		a = append(a, l)
	}
	return a
}
func intersectionCount(a, b map[int32]bool) int32 {
	var n int32
	for k := range a {
		if b[k] {
			n++
		}
	}
	return n
}
func Fxch_SCHashTableEntryCompare(t *Fxch_SCHashTable_t, c Cubes, a, b *Fxch_SubCube_t) int32 {
	x, y := c[a.ICube], c[b.ICube]
	if len(x) == 0 || len(y) == 0 || x[0] != y[0] || a.Id != b.Id || intersectionCount(t.man.outputs[a.ICube], t.man.outputs[b.ICube]) == 0 {
		return 0
	}
	if a.ILit1 > 0 && b.ILit1 > 0 {
		if x[a.ILit0] == y[b.ILit0] || x[a.ILit0] == y[b.ILit1] || x[a.ILit1] == y[b.ILit0] || x[a.ILit1] == y[b.ILit1] {
			return 0
		}
	}
	return boolInt(slices.Equal(subcube(x, a), subcube(y, b)))
}
func Fxch_SCHashTableInsert(t *Fxch_SCHashTable_t, c Cubes, id, cube, l0, l1 uint32, update byte) int32 {
	s := Fxch_SubCube_t{id, cube, l0, l1}
	var pairs int32
	for _, old := range t.bins[id] {
		if (old.ILit1 > 0) != (l1 > 0) || Fxch_SCHashTableEntryCompare(t, c, &old, &s) == 0 {
			continue
		}
		if old.ILit0 == 0 || l0 == 0 {
			continue
		}
		base := Fxch_DivCreate(t.man, &old, &s)
		if base < 0 {
			continue
		}
		n := intersectionCount(t.man.outputs[old.ICube], t.man.outputs[cube])
		var d int32
		for i := int32(0); i < n; i++ {
			d = Fxch_DivAdd(t.man, int32(update), 0, base)
		}
		t.man.divisors[d].pairs = append(t.man.divisors[d].pairs, [2]int32{int32(old.ICube), int32(cube)})
		pairs++
	}
	t.bins[id] = append(t.bins[id], s)
	return pairs
}
func Fxch_SCHashTableRemove(t *Fxch_SCHashTable_t, c Cubes, id, cube, l0, l1 uint32, update byte) int32 {
	bin := t.bins[id]
	at := -1
	for i, s := range bin {
		if s.ICube == cube && s.ILit0 == l0 && s.ILit1 == l1 {
			at = i
			break
		}
	}
	if at < 0 {
		return 0
	}
	s := bin[at]
	var pairs int32
	for i, old := range bin {
		if i == at || (old.ILit1 > 0) != (l1 > 0) || old.ILit0 == 0 || l0 == 0 || Fxch_SCHashTableEntryCompare(t, c, &old, &s) == 0 {
			continue
		}
		base := Fxch_DivCreate(t.man, &old, &s)
		if base < 0 {
			continue
		}
		n := intersectionCount(t.man.outputs[old.ICube], t.man.outputs[cube])
		var d int32
		for j := int32(0); j < n; j++ {
			d = Fxch_DivRemove(t.man, int32(update), 0, base)
		}
		a := t.man.divisors[d].pairs
		out := a[:0]
		for _, pair := range a {
			if pair != [2]int32{int32(old.ICube), int32(cube)} && pair != [2]int32{int32(cube), int32(old.ICube)} {
				out = append(out, pair)
			}
		}
		t.man.divisors[d].pairs = out
		pairs++
	}
	t.bins[id] = append(bin[:at], bin[at+1:]...)
	return pairs
}
func Fxch_SCHashTableMemory(t *Fxch_SCHashTable_t) uint64 {
	n := uint64(unsafe.Sizeof(*t))
	for _, bin := range t.bins {
		n += uint64(cap(bin)) * uint64(unsafe.Sizeof(Fxch_SubCube_t{}))
	}
	return n
}
func Fxch_SCHashTablePrint(t *Fxch_SCHashTable_t) {
	n := 0
	for _, b := range t.bins {
		n += len(b)
	}
	fmt.Fprintf(t.man.writer(), "SubCube entries %d memory %d bytes\n", n, Fxch_SCHashTableMemory(t))
}
