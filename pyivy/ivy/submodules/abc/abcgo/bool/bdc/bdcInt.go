package bdc

import "math/bits"

type Bdc_Isf_t struct {
	Supp, Uniq uint32
	On, Off    []uint32
}
type Bdc_Man_t struct {
	Pars                                                           Bdc_Par_t
	nVars, nNodesMax, nNodesNew                                    int32
	nodes                                                          []*Bdc_Fun_t
	root                                                           *Bdc_Fun_t
	table                                                          map[uint32][]*Bdc_Fun_t
	ol, or, al, ar                                                 Bdc_Isf_t
	calls, numNodes, numMuxes, numAnds, numOrs, numWeaks, numReuse int32
}

func truthWords(n int32) int32 {
	if n <= 5 {
		return 1
	}
	return 1 << uint(n-5)
}
func bdcFlag(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func Bdc_FunNew(p *Bdc_Man_t) *Bdc_Fun_t {
	if len(p.nodes) >= 512 || p.nNodesNew >= p.nNodesMax {
		return nil
	}
	f := &Bdc_Fun_t{id: int32(len(p.nodes))}
	p.nodes = append(p.nodes, f)
	p.nNodesNew++
	return f
}
func Bdc_FunWithId(p *Bdc_Man_t, id int32) *Bdc_Fun_t { return p.nodes[id] }
func Bdc_FunId(_ *Bdc_Man_t, f *Bdc_Fun_t) int32      { return Bdc_Regular(f).id }
func Bdc_IsfStart(p *Bdc_Man_t, f *Bdc_Isf_t) {
	*f = Bdc_Isf_t{On: make([]uint32, truthWords(p.nVars)), Off: make([]uint32, truthWords(p.nVars))}
}
func Bdc_IsfClean(f *Bdc_Isf_t)   { f.Supp, f.Uniq = 0, 0 }
func Bdc_IsfCopy(a, b *Bdc_Isf_t) { *a, *b = *b, *a }
func Bdc_IsfNot(f *Bdc_Isf_t)     { f.On, f.Off = f.Off, f.On }
func ttZero(t []uint32) bool {
	for _, w := range t {
		if w != 0 {
			return false
		}
	}
	return true
}
func ttFull(t []uint32) bool {
	for _, w := range t {
		if w != ^uint32(0) {
			return false
		}
	}
	return true
}
func ttDisjoint(a, b []uint32) bool {
	for i := range a {
		if a[i]&b[i] != 0 {
			return false
		}
	}
	return true
}
func ttDisjoint3(a, b, c []uint32) bool {
	for i := range a {
		if a[i]&b[i]&c[i] != 0 {
			return false
		}
	}
	return true
}
func ttImpl(a, b []uint32) bool {
	for i := range a {
		if a[i]&^b[i] != 0 {
			return false
		}
	}
	return true
}
func ttBinary(a, b []uint32, op byte) []uint32 {
	r := make([]uint32, len(a))
	for i := range a {
		switch op {
		case '&':
			r[i] = a[i] & b[i]
		case '|':
			r[i] = a[i] | b[i]
		case '^':
			r[i] = a[i] ^ b[i]
		case '-':
			r[i] = a[i] &^ b[i]
		}
	}
	return r
}
func ttNot(a []uint32) []uint32 {
	r := make([]uint32, len(a))
	for i, w := range a {
		r[i] = ^w
	}
	return r
}
func ttCof(a []uint32, v int32, one bool) []uint32 {
	r := make([]uint32, len(a))
	if v < 5 {
		mask := [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}[v]
		shift := uint(1 << uint(v))
		for i, w := range a {
			if one {
				w &= mask
				r[i] = w | (w >> shift)
			} else {
				w &^= mask
				r[i] = w | (w << shift)
			}
		}
		return r
	}
	bit := 1 << uint(v-5)
	for i := range a {
		index := i &^ bit
		if one {
			index |= bit
		}
		r[i] = a[index]
	}
	return r
}
func ttQuant(a []uint32, v int32, forall bool) []uint32 {
	op := byte('|')
	if forall {
		op = '&'
	}
	return ttBinary(ttCof(a, v, false), ttCof(a, v, true), op)
}
func ttExistSet(a []uint32, set uint32) []uint32 {
	r := append([]uint32(nil), a...)
	for set != 0 {
		v := bits.TrailingZeros32(set)
		r = ttQuant(r, int32(v), false)
		set &= set - 1
	}
	return r
}
func ttHasVar(t []uint32, v int32) bool {
	if v < 5 {
		mask := [5]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}[v]
		for _, w := range t {
			if w&^mask != (w>>uint(1<<uint(v)))&^mask {
				return true
			}
		}
		return false
	}
	bit := 1 << uint(v-5)
	for i, w := range t {
		if w != t[i^bit] {
			return true
		}
	}
	return false
}
func ttSupport(a []uint32, n int32) uint32 {
	var supp uint32
	for v := int32(0); v < n; v++ {
		if ttHasVar(a, v) {
			supp |= 1 << uint(v)
		}
	}
	return supp
}
func ttCount(a []uint32) int32 {
	n := 0
	for _, w := range a {
		n += bits.OnesCount32(w)
	}
	return int32(n)
}
func funTruth(f *Bdc_Fun_t) []uint32 {
	base := Bdc_Regular(f)
	if Bdc_IsComplement(f) != 0 {
		return ttNot(base.Func)
	}
	return base.Func
}
