package dau

// Bound-set decomposition from dauNonDsd.c. A set uses two bits per variable:
// zero=free, one=unique bound, three=shared. Cofactor classes are discovered in
// the original Gray-code order so generated decomposition phases stay stable.
import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau/canon"
	"math/bits"
	"strings"
)

func dauDecStretch(t uint64, n int32) uint64 {
	if n >= 6 {
		return t
	}
	size := uint(1) << uint(n)
	t &= (uint64(1) << size) - 1
	for s := size; s < 64; s *= 2 {
		t |= t << s
	}
	return t
}
func dauDecSchedule(n int32) []int32 {
	if n < 1 || n > 16 {
		panic("invalid DAU Gray schedule width")
	}
	a := make([]int32, 1<<uint(n))
	for i := 0; i < len(a)-1; i++ {
		a[i] = int32(bits.TrailingZeros(uint(i + 1)))
	}
	a[len(a)-1] = n - 1
	return a
}
func dauDecBoundMap(bound, shared, mask int32) ([16]int32, [16]int32) {
	var uniquePos, sharedPos [16]int32
	s, b := int32(0), int32(0)
	for v := int32(0); v < bound; v++ {
		if mask&(1<<uint(v)) != 0 {
			uniquePos[v] = -1
			sharedPos[v] = s
			s++
		} else {
			sharedPos[v] = -1
			uniquePos[v] = b
			b++
		}
	}
	if s != shared {
		panic("DAU shared-set mask count mismatch")
	}
	return uniquePos, sharedPos
}
func dauDecValidate(n, free, bound, shared int32, schedule []int32) {
	if n < 2 || n > 16 || n != free+bound || free < 1 || bound < 1 || shared < 0 || shared > 6 || shared > bound || len(schedule) < 1<<uint(bound) {
		panic("invalid DAU bound-set dimensions")
	}
}
func Dau_DecCheckSetTop5(p []uint64, n, free, bound, shared, mask int32, schedule []int32, dec, comp []uint64) int32 {
	dauDecValidate(n, free, bound, shared, schedule)
	if free > 5 {
		panic("DAU small cofactor exceeds five variables")
	}
	var cofs [2][64]uint64
	for s := 0; s < 1<<uint(shared); s++ {
		cofs[0][s] = ^uint64(0)
		cofs[1][s] = ^uint64(0)
	}
	uPos, sPos := dauDecBoundMap(bound, shared, mask)
	shift := 6 - free
	wordMask := int32(1<<uint(shift)) - 1
	truthMask := (uint64(1) << uint(1<<uint(free))) - 1
	mint, b, s := int32(0), int32(0), int32(0)
	for m := 0; m < 1<<uint(bound); m++ {
		value := (p[mint>>uint(shift)] >> uint((mint&wordMask)<<uint(free))) & truthMask
		if cofs[0][s] == ^uint64(0) || cofs[0][s] == value {
			cofs[0][s] = value
		} else if cofs[1][s] == ^uint64(0) || cofs[1][s] == value {
			cofs[1][s] = value
			if dec != nil {
				index := s*(1<<uint(bound-shared)) + b
				dec[index>>6] |= 1 << uint(index&63)
			}
		} else {
			return 0
		}
		v := schedule[m]
		mint ^= 1 << uint(v)
		if mask&(1<<uint(v)) != 0 {
			s ^= 1 << uint(sPos[v])
		} else {
			b ^= 1 << uint(uPos[v])
		}
	}
	if comp != nil {
		for s := int32(0); s < 1<<uint(shared); s++ {
			comp[s>>uint(shift)] |= cofs[0][s] << uint((s&wordMask)<<uint(free))
			other := cofs[1][s]
			if other == ^uint64(0) {
				other = cofs[0][s]
			}
			index := s + (1 << uint(shared))
			comp[index>>uint(shift)] |= other << uint((index&wordMask)<<uint(free))
		}
		if free+shared+1 < 6 {
			comp[0] = dauDecStretch(comp[0], free+shared+1)
		}
	}
	if dec != nil && bound < 6 {
		dec[0] = dauDecStretch(dec[0], bound)
	}
	return 1
}
func Dau_DecCheckSetTop6(p []uint64, n, free, bound, shared, mask int32, schedule []int32, dec, comp []uint64) int32 {
	dauDecValidate(n, free, bound, shared, schedule)
	if free < 6 {
		panic("DAU large cofactor needs at least six variables")
	}
	var cofs [2][64][]uint64
	uPos, sPos := dauDecBoundMap(bound, shared, mask)
	words := dauCountWords(free)
	mint, b, s := int32(0), int32(0), int32(0)
	for m := 0; m < 1<<uint(bound); m++ {
		value := p[int(mint)*words : (int(mint)+1)*words]
		if cofs[0][s] == nil || canon.CompareReverse(cofs[0][s], value) == 0 {
			cofs[0][s] = value
		} else if cofs[1][s] == nil || canon.CompareReverse(cofs[1][s], value) == 0 {
			cofs[1][s] = value
			if dec != nil {
				index := s*(1<<uint(bound-shared)) + b
				dec[index>>6] |= 1 << uint(index&63)
			}
		} else {
			return 0
		}
		v := schedule[m]
		mint ^= 1 << uint(v)
		if mask&(1<<uint(v)) != 0 {
			s ^= 1 << uint(sPos[v])
		} else {
			b ^= 1 << uint(uPos[v])
		}
	}
	if comp != nil {
		for s := 0; s < 1<<uint(shared); s++ {
			copy(comp[s*words:], cofs[0][s])
			other := cofs[1][s]
			if other == nil {
				other = cofs[0][s]
			}
			copy(comp[(s+(1<<uint(shared)))*words:], other)
		}
	}
	if dec != nil && bound < 6 {
		dec[0] = dauDecStretch(dec[0], bound)
	}
	return 1
}
func Dau_DecCheckSetTop(p []uint64, n, free, bound, shared, mask int32, schedule []int32, dec, comp []uint64) int32 {
	if free < 6 {
		return Dau_DecCheckSetTop5(p, n, free, bound, shared, mask, schedule, dec, comp)
	}
	return Dau_DecCheckSetTop6(p, n, free, bound, shared, mask, schedule, dec, comp)
}
func Dau_DecGetMinterm(p []uint64, g, n, mask int32) {
	m, c := int32(0), int32(0)
	for v := int32(0); v < n; v++ {
		if mask&(1<<uint(v)) == 0 {
			if g&(1<<uint(v)) != 0 {
				m |= 1 << uint(c)
			}
			c++
		}
	}
	if c < 2 {
		panic("DAU bound set needs at least two unique variables")
	}
	p[m>>6] |= 1 << uint(m&63)
}
func dauDecCofactor(p []uint64, free, g int32) []uint64 {
	if free < 6 {
		shift := 6 - free
		mask := int32(1<<uint(shift)) - 1
		value := p[g>>uint(shift)] >> uint((g&mask)<<uint(free))
		return []uint64{dauDecStretch(value, free)}
	}
	words := dauCountWords(free)
	return p[int(g)*words : (int(g)+1)*words]
}
func Dau_DecCheckSetAny(p []uint64, n, free, mask, value int32, cof0, cof1, dec []uint64) int32 {
	if free < 1 || free > n-2 {
		panic("invalid DAU free set")
	}
	var a, b []uint64
	if dec != nil {
		dec[0] = 0
	}
	for g := int32(0); g < 1<<uint(n-free); g++ {
		if g&mask != value {
			continue
		}
		cof := dauDecCofactor(p, free, g)
		if a == nil {
			a = cof
		} else if canon.CompareReverse(a, cof) == 0 {
			continue
		} else if b == nil || canon.CompareReverse(b, cof) == 0 {
			if b == nil {
				b = cof
			}
			if dec != nil {
				Dau_DecGetMinterm(dec, g, n-free, mask)
			}
		} else {
			return 0
		}
	}
	if cof0 != nil {
		if a == nil {
			panic("DAU shared minterm absent")
		}
		if b == nil {
			b = a
		}
		copy(cof0, a)
		copy(cof1, b)
	}
	return 1
}
func Dau_DecCheckSet5(p []uint64, n, free, mask, value int32, cof0, cof1, dec []uint64) int32 {
	if free > 5 {
		panic("DAU small free set exceeds five")
	}
	return Dau_DecCheckSetAny(p, n, free, mask, value, cof0, cof1, dec)
}
func Dau_DecCheckSet6(p []uint64, n, free, mask, value int32, cof0, cof1, dec []uint64) int32 {
	if free < 6 {
		panic("DAU large free set is below six")
	}
	return Dau_DecCheckSetAny(p, n, free, mask, value, cof0, cof1, dec)
}
func Dau_DecCheckSetTopOld(p []uint64, n, free, bound, shared, mask int32, cof0, cof1, dec [][]uint64) int32 {
	if n < 3 || n > 16 || n != free+bound || free < 1 || bound < 2 || shared < 0 || shared > bound-2 {
		panic("invalid DAU old decomposition dimensions")
	}
	row := func(a [][]uint64, i int) []uint64 {
		if a == nil {
			return nil
		}
		return a[i]
	}
	for s := 0; s < 1<<uint(shared); s++ {
		value, next := int32(0), uint(0)
		for v := int32(0); v < bound; v++ {
			if mask&(1<<uint(v)) != 0 {
				if s&(1<<next) != 0 {
					value |= 1 << uint(v)
				}
				next++
			}
		}
		if next != uint(shared) {
			panic("DAU old shared mask mismatch")
		}
		if Dau_DecCheckSetAny(p, n, free, mask, value, row(cof0, s), row(cof1, s), row(dec, s)) == 0 {
			return 0
		}
	}
	return 1
}
func Dau_DecCreateSet(vars []int32, size, shared int32) uint32 {
	var set uint32
	for i, v := range vars[:size] {
		if v < 0 || v >= 16 {
			panic("DAU set variable exceeds fifteen")
		}
		set |= 1 << uint(2*v)
		if shared&(1<<uint(i)) != 0 {
			set |= 1 << uint(2*v+1)
		}
	}
	return set
}
func Dau_DecSetHas01(mask uint32) int32 { return int32(mask & (^mask >> 1) & 0x55555555) }
func Dau_DecSetIsContained(sets *vec.Vec_Int_t, new uint32) int32 {
	for _, x := range vec.Vec_IntArray(sets) {
		old := uint32(x)
		if old&^new == 0 && Dau_DecSetHas01(^old&new) == 0 {
			return 1
		}
	}
	return 0
}
func Dau_DecSortSet(set uint32, n int32, unique, shared, free *int32) {
	u, s, f := int32(0), int32(0), int32(0)
	for v := int32(0); v < n; v++ {
		switch set >> uint(2*v) & 3 {
		case 0:
			f++
		case 1:
			u++
		case 3:
			s++
		default:
			panic("invalid DAU two-bit set encoding")
		}
	}
	if unique != nil {
		*unique = u
	}
	if shared != nil {
		*shared = s
	}
	if free != nil {
		*free = f
	}
}
func Dau_DecPrintSet(set uint32, n, newLine int32) {
	var u, s, f int32
	Dau_DecSortSet(set, n, &u, &s, &f)
	var a, b strings.Builder
	b.WriteByte('x')
	for v := int32(0); v < n; v++ {
		switch set >> uint(2*v) & 3 {
		case 0:
			b.WriteByte(byte('a' + v))
		case 1:
			a.WriteByte(byte('a' + v))
		case 3:
			a.WriteByte(byte('A' + v))
			b.WriteByte(byte('A' + v))
		}
	}
	fmt.Printf("S =%2d  D =%2d  C =%2d   x=%s y=%s", s, u+s, s+f+1, a.String(), b.String())
	for count := a.Len() + b.Len() - 1; count < 15; count++ {
		fmt.Print(" ")
	}
	if newLine != 0 {
		fmt.Println()
	}
}
func Dau_DecReadSet(text string) uint32 {
	var set uint32
	for _, c := range text {
		if c >= 'a' && c <= 'p' {
			set |= 1 << uint(2*(c-'a'))
		} else if c >= 'A' && c <= 'P' {
			set |= 3 << uint(2*(c-'A'))
		} else {
			break
		}
	}
	return set
}
func Dau_DecPrintSets(sets *vec.Vec_Int_t, n int32) {
	fmt.Printf("The %d-variable set family contains %d sets:\n", n, vec.Vec_IntSize(sets))
	for _, s := range vec.Vec_IntArray(sets) {
		Dau_DecPrintSet(uint32(s), n, 1)
	}
	fmt.Println()
}
func dauDecMoveVar(p []uint64, v2p, p2v []int32, v, pos int32) {
	old := v2p[v]
	if old == pos {
		return
	}
	canon.SwapVariables(p, int(old), int(pos))
	v2p[p2v[old]], v2p[p2v[pos]] = pos, old
	p2v[old], p2v[pos] = p2v[pos], p2v[old]
}
func Dau_DecMoveFreeToLSB(p []uint64, n int32, v2p, p2v []int32, mask, size int32) {
	next := int32(0)
	for v := int32(0); v < n; v++ {
		if mask&(1<<uint(v)) == 0 {
			dauDecMoveVar(p, v2p, p2v, v, next)
			next++
		}
	}
	if next != n-size {
		panic("DAU bound-set size mismatch")
	}
}
func Dau_DecFindSets_int(input []uint64, n int32, schedules [16][]int32) *vec.Vec_Int_t {
	if n < 0 || n > 16 {
		panic("DAU set search supports zero through sixteen variables")
	}
	p := append([]uint64(nil), input[:dauCountWords(n)]...)
	v2p, p2v, vars := make([]int32, n), make([]int32, n), make([]int32, n)
	for v := int32(0); v < n; v++ {
		if !dauHasVariable(p, int(v)) {
			panic("DAU set search expects full support")
		}
		v2p[v] = v
		p2v[v] = v
	}
	sets := vec.Vec_IntAlloc(32)
	for size := int32(2); size < n; size++ {
		for mask := int32(0); mask < 1<<uint(n); mask++ {
			if int32(bits.OnesCount32(uint32(mask))) != size {
				continue
			}
			Dau_DecMoveFreeToLSB(p, n, v2p, p2v, mask, size)
			copy(vars, p2v[n-size:])
			if Dau_DecCheckSetTop(p, n, n-size, size, 0, 0, schedules[size], nil, nil) != 0 {
				vec.Vec_IntPush(sets, int32(Dau_DecCreateSet(vars, size, 0)))
				continue
			}
			if size == 2 {
				continue
			}
			for shared := int32(1); shared <= min(3, size-2); shared++ {
				for sMask := int32(0); sMask < 1<<uint(size); sMask++ {
					if int32(bits.OnesCount32(uint32(sMask))) != shared {
						continue
					}
					set := Dau_DecCreateSet(vars, size, sMask)
					if Dau_DecSetIsContained(sets, set) != 0 {
						continue
					}
					if Dau_DecCheckSetTop(p, n, n-size, size, shared, sMask, schedules[size], nil, nil) != 0 {
						vec.Vec_IntPush(sets, int32(set))
					}
				}
			}
		}
	}
	return sets
}
func Dau_DecFindSets(input []uint64, n int32) *vec.Vec_Int_t {
	var schedules [16][]int32
	for v := int32(2); v < n; v++ {
		schedules[v] = dauDecSchedule(v)
	}
	return Dau_DecFindSets_int(input, n, schedules)
}
func Dau_DecVarReplace(text string, perm []int32, n int32) string {
	b := []byte(text)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			v := int32(c - 'a')
			if v >= n {
				panic("DAU DSD variable exceeds permutation")
			}
			b[i] = byte('a' + perm[v])
		}
	}
	return string(b)
}

func Dau_DecDecomposeSet(input []uint64, n int32, set uint32, comp, dec []uint64, permC, permD []int32, nC, nD, nS *int32) int32 {
	if n < 3 || n > 16 {
		panic("invalid DAU scalar decomposition width")
	}
	var unique, shared, free []int32
	for v := int32(0); v < n; v++ {
		switch set >> uint(2*v) & 3 {
		case 0:
			free = append(free, v)
		case 1:
			unique = append(unique, v)
		case 3:
			shared = append(shared, v)
		default:
			panic("invalid DAU two-bit set")
		}
	}
	f, s, u := int32(len(free)), int32(len(shared)), int32(len(unique))
	if f+s+1 > 6 || u+s > 6 || u < 2 || f < 1 {
		panic("DAU scalar decomposition exceeds six-input component limits")
	}
	p := append([]uint64(nil), input[:dauCountWords(n)]...)
	v2p, p2v := make([]int32, n), make([]int32, n)
	for v := range v2p {
		v2p[v] = int32(v)
		p2v[v] = int32(v)
	}
	order := append(append(append([]int32(nil), free...), shared...), unique...)
	for pos, v := range order {
		dauDecMoveVar(p, v2p, p2v, v, int32(pos))
	}
	count := 1 << uint(s)
	cof0, cof1, rows := make([][]uint64, count), make([][]uint64, count), make([][]uint64, count)
	for i := 0; i < count; i++ {
		cof0[i] = make([]uint64, 1)
		cof1[i] = make([]uint64, 1)
		rows[i] = make([]uint64, 1)
	}
	if Dau_DecCheckSetTopOld(p, n, f, s+u, s, int32(1<<uint(s))-1, cof0, cof1, rows) == 0 {
		return 0
	}
	comp[0], dec[0] = 0, 0
	for i := 0; i < count; i++ {
		cof := cof1[i][0]&dauTruthVariables[f+s] | cof0[i][0]&^dauTruthVariables[f+s]
		rows[i][0] = dauDecStretch(rows[i][0], u)
		mintC, mintD := ^uint64(0), ^uint64(0)
		for v := int32(0); v < s; v++ {
			a, b := dauTruthVariables[f+v], dauTruthVariables[u+v]
			if i&(1<<uint(v)) == 0 {
				a, b = ^a, ^b
			}
			mintC &= a
			mintD &= b
		}
		comp[0] |= mintC & cof
		dec[0] |= mintD & rows[i][0]
	}
	if permC != nil {
		copy(permC, free)
		copy(permC[len(free):], shared)
		permC[len(free)+len(shared)] = n
	}
	if permD != nil {
		copy(permD, unique)
		copy(permD[len(unique):], shared)
	}
	if nC != nil {
		*nC = f + s + 1
	}
	if nD != nil {
		*nD = u + s
	}
	if nS != nil {
		*nS = s
	}
	return 1
}
func Dau_DecVerify(input []uint64, n int32, dsdC, dsdD string) int32 {
	if n < 0 || n >= 16 {
		panic("DAU DSD verification needs fewer than sixteen original variables")
	}
	c, d := Dau_DsdToTruth(dsdC, n+1), Dau_DsdToTruth(dsdD, n)
	words := dauCountWords(n)
	good := true
	for i := 0; i < words; i++ {
		var a, b uint64
		if n >= 6 {
			a, b = c[i], c[i+words]
		} else {
			a = dauDecStretch(c[0], n)
			b = dauDecStretch(c[0]>>uint(1<<uint(n)), n)
		}
		value := d[i]&b | ^d[i]&a
		if input[i] != value {
			good = false
		}
	}
	if !good {
		fmt.Print("      Verification failed")
	}
	fmt.Println()
	return dauBool(good)
}
func dauDecComponentStrings(comp, dec []uint64, nC, nD int32, permC, permD []int32) (string, string) {
	var c, d []byte
	Dau_DsdDecompose(comp, nC, 0, 1, &c)
	Dau_DsdDecompose(dec, nD, 0, 1, &d)
	return Dau_DecVarReplace(string(c), permC, nC), Dau_DecVarReplace(string(d), permD, nD)
}
func Dau_DecPerform6(input []uint64, n int32, set uint32) int32 {
	comp, dec0 := make([]uint64, 1), make([]uint64, 1)
	permC, permD := make([]int32, 16), make([]int32, 16)
	var nc, nd, ns int32
	if Dau_DecDecomposeSet(input, n, set, comp, dec0, permC, permD, &nc, &nd, &ns) == 0 {
		fmt.Println("  Decomposition does not exist")
		return 0
	}
	nu, nf := nd-ns, nc-ns-1
	c0, c1 := dauCofactorWord(comp[0], nf+ns, false), dauCofactorWord(comp[0], nf+ns, true)
	counter := 0
	for i := uint32(0); i < 1<<uint(1<<uint(ns)); i++ {
		if i&1 != 0 {
			continue
		}
		fc, fd := uint64(0), uint64(0)
		for m := int32(0); m < 1<<uint(ns); m++ {
			if i&(1<<uint(m)) == 0 {
				continue
			}
			mc, md := ^uint64(0), ^uint64(0)
			for v := int32(0); v < ns; v++ {
				a, b := dauTruthVariables[nf+v], dauTruthVariables[nu+v]
				if m&(1<<uint(v)) == 0 {
					a, b = ^a, ^b
				}
				mc &= a
				md &= b
			}
			fc |= mc
			fd |= md
		}
		x := dauTruthVariables[nf+ns]
		tc := (^x & ((c0 &^ fc) | (c1 & fc))) | (x & ((c1 &^ fc) | (c0 & fc)))
		td := dec0[0] ^ fd
		dc, dd := dauDecComponentStrings([]uint64{tc}, []uint64{td}, nc, nd, permC, permD)
		fmt.Printf("%3d : %24s  %24s  ", counter, dd, dc)
		counter++
		if Dau_DecVerify(input, n, dc, dd) == 0 {
			return 0
		}
	}
	return 1
}
func Dau_DecPerform(input []uint64, n int32, set uint32) int32 {
	if n < 3 || n >= 16 {
		panic("DAU decomposition diagnostic supports three through fifteen inputs")
	}
	p := append([]uint64(nil), input[:dauCountWords(n)]...)
	var unique, shared, free int32
	Dau_DecSortSet(set, n, &unique, &shared, &free)
	v2p, p2v := make([]int32, n), make([]int32, n)
	for v := range v2p {
		v2p[v] = int32(v)
		p2v[v] = int32(v)
	}
	var pc, pd []int32
	pos := int32(0)
	for _, kind := range []uint32{0, 3, 1} {
		for v := int32(0); v < n; v++ {
			if set>>uint(2*v)&3 != kind {
				continue
			}
			dauDecMoveVar(p, v2p, p2v, v, pos)
			pos++
			if kind == 1 {
				pd = append(pd, v)
			} else {
				pc = append(pc, v)
			}
		}
	}
	pd = append(pd, pc[free:free+shared]...)
	pc = append(pc, n)
	nc, nd := free+shared+1, unique+shared
	comp, dec := make([]uint64, dauCountWords(nc)), make([]uint64, dauCountWords(nd))
	if Dau_DecCheckSetTop(p, n, free, unique+shared, shared, int32(1<<uint(shared))-1, dauDecSchedule(unique+shared), dec, comp) == 0 {
		fmt.Println("  Decomposition does not exist")
		return 0
	}
	dc, dd := dauDecComponentStrings(comp, dec, nc, nd, pc, pd)
	fmt.Printf("     %3d : %24s  %24s  ", 0, dd, dc)
	return Dau_DecVerify(input, n, dc, dd)
}
func Dau_DecTrySets(input []uint64, n, verbose int32) {
	sets := Dau_DecFindSets(input, n)
	if verbose == 0 {
		return
	}
	Dau_DsdPrintFromTruth(input, n)
	fmt.Printf("This %d-variable function has %d decomposable variable sets:\n", n, vec.Vec_IntSize(sets))
	for i, s := range vec.Vec_IntArray(sets) {
		fmt.Printf("Set %4d : ", i)
		if n > 6 {
			Dau_DecPrintSet(uint32(s), n, 0)
			Dau_DecPerform(input, n, uint32(s))
		} else {
			Dau_DecPrintSet(uint32(s), n, 1)
			Dau_DecPerform6(input, n, uint32(s))
		}
	}
}
func dauDecExample() uint64 {
	v := dauTruthVariables
	a := (^v[1] & v[2]) | (v[1] & v[3])
	b := (^v[1] & v[4]) | (v[1] & v[5])
	return ^v[0]&a | v[0]&b
}
func Dau_DecFindSetsTest2() { Dau_DecPrintSets(Dau_DecFindSets([]uint64{dauDecExample()}, 5), 5) }
func Dau_DecFindSetsTest3() { Dau_DecPerform6([]uint64{dauDecExample()}, 6, Dau_DecReadSet("Bcd")) }
func Dau_DecFindSetsTest()  { Dau_DecTrySets([]uint64{0x000030f00d0d3fff}, 6, 1) }
