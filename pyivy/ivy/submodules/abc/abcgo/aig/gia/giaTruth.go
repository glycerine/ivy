package gia

// Truth-table computation and Shannon synthesis, ported from giaTruth.c.
// Tables use little-endian 64-bit words; unused variables repeat their pattern.
import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau/canon"
	"time"
)

// Keep the sentinel compatible with Gia_ObjResetNumId in gia.h.
const giaTruthUnnumbered = -1_000_000_000

var s_Truth6 = [6]word{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func giaTruthWordCount64(n int32) int {
	if n < 0 || n > 30 {
		panic("GIA truth variable count out of range")
	}
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func giaTruthVariable(n int32, v int) []word {
	r := make([]word, giaTruthWordCount64(n))
	for i := range r {
		if v < 6 {
			r[i] = s_Truth6[v]
		} else if i&(1<<uint(v-6)) != 0 {
			r[i] = ^word(0)
		}
	}
	return r
}
func Gla_ObjTruthElem(p *Gia_Man_t, i int32) []word {
	return vec.Vec_PtrEntry(p.vTtInputs, int(i)).([]word)
}
func Gla_ObjTruthNodeId(p *Gia_Man_t, id int32) []word {
	off := int(id * p.nTtWords)
	return vec.Vec_WrdArray(p.vTtMemory)[off : off+int(p.nTtWords)]
}
func Gla_ObjTruthNode(p *Gia_Man_t, obj *Gia_Obj_t) []word {
	return Gla_ObjTruthNodeId(p, vec.Vec_IntEntry(p.vTtNums, int(obj.id)))
}
func Gla_ObjTruthFree1(p *Gia_Man_t) []word {
	a := vec.Vec_WrdArray(p.vTtMemory)
	return a[len(a)-int(p.nTtWords):]
}
func Gla_ObjTruthFree2(p *Gia_Man_t) []word {
	a := vec.Vec_WrdArray(p.vTtMemory)
	n := int(p.nTtWords)
	return a[len(a)-2*n : len(a)-n]
}
func Gla_ObjTruthConst0(p *Gia_Man_t, dst []word) []word { clear(dst[:p.nTtWords]); return dst }
func Gla_ObjTruthDup(p *Gia_Man_t, dst, src []word, c int32) []word {
	for i := 0; i < int(p.nTtWords); i++ {
		dst[i] = src[i]
		if c != 0 {
			dst[i] = ^dst[i]
		}
	}
	return dst
}

// A cut evaluator owns its memoized tables. A CI outside the cut is an error.
// It also supports XOR, buffer and explicit MUX objects used by modern GIA.
type giaTruthEval struct {
	p      *Gia_Man_t
	memo   map[int32][]word
	leaves map[int32]bool
}

func giaNewTruthEval(p *Gia_Man_t, n int32, leaves []int32) *giaTruthEval {
	e := &giaTruthEval{p: p, memo: map[int32][]word{0: make([]word, giaTruthWordCount64(n))}, leaves: map[int32]bool{}}
	for i, id := range leaves {
		if _, ok := e.memo[id]; !ok {
			e.memo[id] = giaTruthVariable(n, i)
			e.leaves[id] = true
		}
	}
	return e
}
func (e *giaTruthEval) lit(l int32) []word {
	r := e.node(l >> 1)
	if l&1 == 0 {
		return r
	}
	out := make([]word, len(r))
	for i, x := range r {
		out[i] = ^x
	}
	return out
}
func (e *giaTruthEval) node(id int32) []word {
	if r, ok := e.memo[id]; ok {
		return r
	}
	o := Gia_ManObj(e.p, id)
	if Gia_ObjIsCi(o) != 0 {
		panic("GIA truth cut does not cover a CI")
	}
	a := e.lit(Gia_ObjFaninLit0(o, id))
	r := make([]word, len(a))
	if Gia_ObjIsCo(o) != 0 || Gia_ObjIsBuf(o) != 0 {
		copy(r, a)
	} else {
		b := e.lit(Gia_ObjFaninLit1(o, id))
		var c []word
		if int(id) < len(e.p.PMuxes) && e.p.PMuxes[id] != 0 {
			c = e.lit(int32(e.p.PMuxes[id]))
		}
		for i := range r {
			if c != nil {
				r[i] = c[i]&b[i] | ^c[i]&a[i]
			} else if Gia_ObjIsXor(o) != 0 {
				r[i] = a[i] ^ b[i]
			} else {
				r[i] = a[i] & b[i]
			}
		}
	}
	e.memo[id] = r
	return r
}

// giaTruthCis borrows CI IDs; callers only read them during evaluation.
func giaTruthCis(p *Gia_Man_t) []int32 { return vec.Vec_IntArray(p.vCis) }
func Gia_LutComputeTruth6Simple_rec(p *Gia_Man_t, id int32) word {
	leaves := giaTruthCis(p)
	if len(leaves) > 6 {
		leaves = leaves[:6]
	}
	return giaNewTruthEval(p, 6, leaves).node(id)[0]
}
func Gia_LutComputeTruth6Simple(p *Gia_Man_t, iPo int32) word {
	return Gia_LutComputeTruth6Simple_rec(p, Gia_ManPo(p, iPo).id)
}
func Gia_LutComputeTruth6Map_rec(p *Gia_Man_t, id int32, m *Vec_Int_t) word {
	e := giaNewTruthEval(p, 6, nil)
	for i, ci := range giaTruthCis(p) {
		v := vec.Vec_IntEntry(m, i)
		if v >= 0 && v < 6 {
			e.memo[ci] = []word{s_Truth6[v]}
		}
	}
	return e.node(id)[0]
}
func Gia_LutComputeTruth6Map(p *Gia_Man_t, iPo int32, m *Vec_Int_t) word {
	return Gia_LutComputeTruth6Map_rec(p, Gia_ManPo(p, iPo).id, m)
}
func Abc_Tt5HasVar(t uint32, v int32) int32 {
	if Abc_Tt5Cofactor0(t, v) != Abc_Tt5Cofactor1(t, v) {
		return 1
	}
	return 0
}
func Abc_Tt5Cofactor0(t uint32, v int32) uint32 {
	if v < 0 || v >= 5 {
		panic("five-input cofactor index out of range")
	}
	x := t &^ uint32(s_Truth6[v])
	return x | x<<uint(1<<uint(v))
}
func Abc_Tt5Cofactor1(t uint32, v int32) uint32 {
	if v < 0 || v >= 5 {
		panic("five-input cofactor index out of range")
	}
	x := t & uint32(s_Truth6[v])
	return x | x>>uint(1<<uint(v))
}
func giaTruthCof(t word, v int32, one bool) word {
	m := s_Truth6[v]
	s := uint(1 << uint(v))
	if one {
		x := t & m
		return x | x>>s
	}
	x := t &^ m
	return x | x<<s
}
func giaTruthAnd(p *Gia_Man_t, a, b int32, hash bool) int32 {
	if a == 0 || b == 0 || a == b^1 {
		return 0
	}
	if a == 1 {
		return b
	}
	if b == 1 || a == b {
		return a
	}
	return p.AndLiteral(a, b, hash)
}
func Gia_Truth6ToGia(p *Gia_Man_t, lits []int32, n int32, t word, hash int32) int32 {
	if t == 0 {
		return 0
	}
	if t == ^word(0) {
		return 1
	}
	if n < 1 || n > 6 || int(n) > len(lits) {
		panic("invalid Shannon synthesis input")
	}
	v := n - 1
	for v >= 0 && giaTruthCof(t, v, false) == giaTruthCof(t, v, true) {
		v--
	}
	if v < 0 {
		panic("non-normalized truth table")
	}
	a := Gia_Truth6ToGia(p, lits, v, giaTruthCof(t, v, false), hash)
	b := Gia_Truth6ToGia(p, lits, v, giaTruthCof(t, v, true), hash)
	x := giaTruthAnd(p, lits[v], b, hash != 0)
	y := giaTruthAnd(p, lits[v]^1, a, hash != 0)
	return giaTruthAnd(p, x^1, y^1, hash != 0) ^ 1
}
func Gia_Truth5ToGia(p *Gia_Man_t, lits []int32, n int32, t uint32, hash int32) int32 {
	if n > 5 {
		panic("five-input synthesis input too wide")
	}
	return Gia_Truth6ToGia(p, lits, n, word(t)|word(t)<<32, hash)
}
func giaTruthLutLeaves(p *Gia_Man_t, id int32) []int32 {
	a := vec.Vec_IntArray(p.VMapping)
	if id < 0 || int(id) >= len(a) || a[id] <= 0 {
		panic("GIA object has no LUT mapping")
	}
	off := int(a[id])
	n := int(a[off])
	if n > 6 {
		panic("LUT has more than six inputs")
	}
	return a[off+1 : off+1+n]
}
func giaTruthSeed6(p *Gia_Man_t, leaves []int32, temp *Vec_Wrd_t) {
	if len(leaves) > 6 {
		panic("cut has more than six inputs")
	}
	vec.Vec_WrdFillExtra(temp, int(p.nObjs), 0)
	Gia_ManIncrementTravId(p)
	Gia_ObjSetTravIdCurrentId(p, 0)
	vec.Vec_WrdWriteEntry(temp, 0, 0)
	for i, id := range leaves {
		vec.Vec_WrdWriteEntry(temp, int(id), s_Truth6[i])
		Gia_ObjSetTravIdCurrentId(p, id)
	}
}

// giaTruthWordNode evaluates one object after its fanins have been supplied.
func giaTruthWordNode(p *Gia_Man_t, o *Gia_Obj_t, read func(int32) word) word {
	if o.id == 0 {
		return 0
	}
	if Gia_ObjIsCi(o) != 0 {
		panic("GIA truth cut does not cover a CI")
	}
	a := read(Gia_ObjFaninId0(o, o.id))
	if o.fCompl0 != 0 {
		a = ^a
	}
	if Gia_ObjIsCo(o) != 0 || Gia_ObjIsBuf(o) != 0 {
		return a
	}
	b := read(Gia_ObjFaninId1(o, o.id))
	if o.fCompl1 != 0 {
		b = ^b
	}
	if int(o.id) < len(p.PMuxes) && p.PMuxes[o.id] != 0 {
		l := int32(p.PMuxes[o.id])
		c := read(l >> 1)
		if l&1 != 0 {
			c = ^c
		}
		return c&b | ^c&a
	}
	if Gia_ObjIsXor(o) != 0 {
		return a ^ b
	}
	return a & b
}
func Gia_LutComputeTruth6_rec(p *Gia_Man_t, id int32, temp *Vec_Wrd_t) word {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return vec.Vec_WrdEntry(temp, int(id))
	}
	r := giaTruthWordNode(p, p.pObjs[id], func(f int32) word { return Gia_LutComputeTruth6_rec(p, f, temp) })
	vec.Vec_WrdWriteEntry(temp, int(id), r)
	Gia_ObjSetTravIdCurrentId(p, id)
	return r
}
func Gia_LutComputeTruth6(p *Gia_Man_t, id int32, temp *Vec_Wrd_t) word {
	giaTruthSeed6(p, giaTruthLutLeaves(p, id), temp)
	return Gia_LutComputeTruth6_rec(p, id, temp)
}
func Gia_ObjComputeTruthTable6Lut_rec(p *Gia_Man_t, id int32, temp *Vec_Wrd_t) {
	Gia_LutComputeTruth6_rec(p, id, temp)
}
func Gia_ObjComputeTruthTable6Lut(p *Gia_Man_t, id int32, temp *Vec_Wrd_t) word {
	return Gia_LutComputeTruth6(p, id, temp)
}
func Gia_ObjComputeTruth6(p *Gia_Man_t, id int32, supp *Vec_Int_t, temp *Vec_Wrd_t) word {
	giaTruthSeed6(p, vec.Vec_IntArray(supp), temp)
	return Gia_LutComputeTruth6_rec(p, id, temp)
}
func Gia_ObjComputeTruth6CisSupport_rec(p *Gia_Man_t, id int32, supp *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	o := Gia_ManObj(p, id)
	if Gia_ObjIsCi(o) != 0 {
		vec.Vec_IntPushOrder(supp, id)
		return
	}
	for _, f := range giaObjectFanins(p, o) {
		Gia_ObjComputeTruth6CisSupport_rec(p, f, supp)
	}
}
func Gia_ObjComputeTruth6Cis(p *Gia_Man_t, lit int32, supp *Vec_Int_t, temp *Vec_Wrd_t) word {
	vec.Vec_IntClear(supp)
	Gia_ManIncrementTravId(p)
	Gia_ObjComputeTruth6CisSupport_rec(p, lit>>1, supp)
	if vec.Vec_IntSize(supp) > 6 {
		return 0
	}
	r := Gia_ObjComputeTruth6(p, lit>>1, supp, temp)
	if lit&1 != 0 {
		r = ^r
	}
	return r
}
func Gia_ObjComputeTruthTable6_rec(p *Gia_Man_t, o *Gia_Obj_t, temp *Vec_Wrd_t) {
	if Gia_ObjIsTravIdCurrentId(p, o.id) != 0 {
		return
	}
	r := giaTruthWordNode(p, o, func(id int32) word {
		f := p.pObjs[id]
		Gia_ObjComputeTruthTable6_rec(p, f, temp)
		return vec.Vec_WrdEntry(temp, int(f.Value))
	})
	o.Value = uint32(vec.Vec_WrdSize(temp))
	vec.Vec_WrdPush(temp, r)
	Gia_ObjSetTravIdCurrentId(p, o.id)
}
func Gia_ObjComputeTruthTable6(p *Gia_Man_t, o *Gia_Obj_t, supp *Vec_Int_t, temp *Vec_Wrd_t) word {
	if vec.Vec_IntSize(supp) > 6 {
		panic("cut too large")
	}
	vec.Vec_WrdClear(temp)
	Gia_ManIncrementTravId(p)
	for i, id := range vec.Vec_IntArray(supp) {
		p.pObjs[id].Value = uint32(i)
		vec.Vec_WrdPush(temp, s_Truth6[i])
		Gia_ObjSetTravIdCurrentId(p, id)
	}
	if Gia_ObjIsTravIdCurrentId(p, 0) == 0 {
		p.pObjs[0].Value = uint32(vec.Vec_WrdSize(temp))
		vec.Vec_WrdPush(temp, 0)
		Gia_ObjSetTravIdCurrentId(p, 0)
	}
	Gia_ObjComputeTruthTable6_rec(p, o, temp)
	return vec.Vec_WrdEntry(temp, int(o.Value))
}

func Gia_ObjComputeTruthTableStart(p *Gia_Man_t, n int32) {
	if p.vTtMemory != nil {
		panic("truth workspace already active")
	}
	p.nTtVars = n
	p.nTtWords = int32(giaTruthWordCount64(n))
	p.vTtNums = vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntFill(p.vTtNums, int(p.nObjs), giaTruthUnnumbered)
	p.vTtNodes = vec.Vec_IntAlloc(256)
	p.vTtInputs = vec.Vec_PtrAlloc(max(6, int(n)))
	for i := 0; i < max(6, int(n)); i++ {
		vec.Vec_PtrPush(p.vTtInputs, giaTruthVariable(n, i))
	}
	p.vTtMemory = vec.Vec_WrdStart(int(p.nTtWords) * 64)
}
func Gia_ObjComputeTruthTableStop(p *Gia_Man_t) {
	p.nTtVars = 0
	p.nTtWords = 0
	p.vTtNums = nil
	p.vTtNodes = nil
	p.vTtInputs = nil
	p.vTtMemory = nil
}
func Gia_ObjCollectInternal_rec(p *Gia_Man_t, o *Gia_Obj_t) {
	if Gia_ObjIsAnd(o) == 0 || o.fMark0 != 0 {
		return
	}
	o.fMark0 = 1
	for _, id := range giaObjectFanins(p, o) {
		Gia_ObjCollectInternal_rec(p, p.pObjs[id])
	}
	vec.Vec_IntFillExtra(p.vTtNums, int(p.nObjs), giaTruthUnnumbered)
	vec.Vec_IntWriteEntry(p.vTtNums, int(o.id), int32(vec.Vec_IntSize(p.vTtNodes)))
	vec.Vec_IntPush(p.vTtNodes, o.id)
}
func Gia_ObjCollectInternal(p *Gia_Man_t, o *Gia_Obj_t) {
	vec.Vec_IntClear(p.vTtNodes)
	Gia_ObjCollectInternal_rec(p, o)
}
func Gia_ObjCollectInternalCut_rec(p *Gia_Man_t, id int32) {
	if id == 0 || vec.Vec_IntEntry(p.vTtNums, int(id)) > giaTruthUnnumbered {
		return
	}
	o := p.pObjs[id]
	if Gia_ObjIsCi(o) != 0 {
		panic("cut does not cover CI")
	}
	for _, f := range giaObjectFanins(p, o) {
		Gia_ObjCollectInternalCut_rec(p, f)
	}
	vec.Vec_IntWriteEntry(p.vTtNums, int(id), int32(vec.Vec_IntSize(p.vTtNodes)))
	vec.Vec_IntPush(p.vTtNodes, id)
}
func Gia_ObjCollectInternalCut(p *Gia_Man_t, id int32, leaves *Vec_Int_t) {
	vec.Vec_IntFillExtra(p.vTtNums, int(p.nObjs), giaTruthUnnumbered)
	for i, l := range vec.Vec_IntArray(leaves) {
		if l == id {
			panic("cut root is also a leaf")
		}
		if vec.Vec_IntEntry(p.vTtNums, int(l)) <= giaTruthUnnumbered {
			vec.Vec_IntWriteEntry(p.vTtNums, int(l), -int32(i))
		}
	}
	vec.Vec_IntClear(p.vTtNodes)
	vec.Vec_IntPush(p.vTtNodes, -1)
	Gia_ObjCollectInternalCut_rec(p, id)
}
func giaTruthSave(p *Gia_Man_t, e *giaTruthEval, root int32, cut bool) []word {
	n := int(p.nTtWords)
	vec.Vec_IntFillExtra(p.vTtNums, int(p.nObjs), giaTruthUnnumbered)
	vec.Vec_IntClear(p.vTtNodes)
	if cut {
		vec.Vec_IntPush(p.vTtNodes, -1)
	}
	seen := map[int32]bool{}
	var collect func(int32)
	collect = func(id int32) {
		if id == 0 || seen[id] || e.leaves[id] {
			return
		}
		seen[id] = true
		o := p.pObjs[id]
		for _, f := range giaObjectFanins(p, o) {
			collect(f)
		}
		if Gia_ObjIsAnd(o) != 0 {
			vec.Vec_IntWriteEntry(p.vTtNums, int(id), int32(vec.Vec_IntSize(p.vTtNodes)))
			vec.Vec_IntPush(p.vTtNodes, id)
		}
	}
	collect(root)
	vec.Vec_WrdFillExtra(p.vTtMemory, (vec.Vec_IntSize(p.vTtNodes)+2)*n, 0)
	for _, id := range vec.Vec_IntArray(p.vTtNodes) {
		if id >= 0 {
			copy(Gla_ObjTruthNode(p, p.pObjs[id]), e.memo[id])
			p.pObjs[id].fMark0 = 0
		}
	}
	copy(Gla_ObjTruthFree2(p), e.memo[root])
	if cut && Gia_ObjIsAnd(p.pObjs[root]) != 0 {
		return Gla_ObjTruthNode(p, p.pObjs[root])
	}
	return Gla_ObjTruthFree2(p)
}

// Returned tables borrow manager storage until the next computation or Stop.
func Gia_ObjComputeTruthTable(p *Gia_Man_t, o *Gia_Obj_t) []word {
	n := Gia_ManPiNum(p)
	if p.vTtMemory == nil {
		Gia_ObjComputeTruthTableStart(p, n)
	} else if p.nTtVars != n {
		panic("primary input count changed since workspace creation")
	}
	leaves := giaTruthCis(p)[:n]
	e := giaNewTruthEval(p, n, leaves)
	e.node(o.id)
	r := giaTruthSave(p, e, o.id, false)
	return r
}
func Gia_ObjComputeTruthTableCut(p *Gia_Man_t, o *Gia_Obj_t, leaves *Vec_Int_t) []word {
	if p.vTtMemory == nil || vec.Vec_IntSize(leaves) > int(p.nTtVars) {
		panic("invalid truth workspace or cut width")
	}
	for _, l := range vec.Vec_IntArray(leaves) {
		if l == o.id {
			panic("cut root is also a leaf")
		}
	}
	e := giaNewTruthEval(p, p.nTtVars, vec.Vec_IntArray(leaves))
	e.node(o.id)
	r := giaTruthSave(p, e, o.id, true)
	vec.Vec_IntFill(p.vTtNums, vec.Vec_IntSize(p.vTtNums), giaTruthUnnumbered)
	return r
}
func Gia_ObjComputeTruthTableTest(p *Gia_Man_t) {
	start := time.Now()
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		Gia_ObjComputeTruthTable(p, Gia_ManPo(p, i))
	}
	fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
}
func Gia_Truth6ToGiaTest(p *Gia_Man_t) {
	count := 0
	n := Gia_ManCoNum(p)
	supp := vec.Vec_IntAlloc(6)
	temp := vec.Vec_WrdStart(int(p.nObjs))
	for i := int32(0); i < n; i++ {
		o := Gia_ManCo(p, i)
		t := Gia_ObjComputeTruth6Cis(p, Gia_ObjFaninLit0(o, o.id), supp, temp)
		if vec.Vec_IntSize(supp) > 5 {
			continue
		}
		lits := append([]int32(nil), vec.Vec_IntArray(supp)...)
		for k := range lits {
			lits[k] <<= 1
		}
		l := Gia_Truth5ToGia(p, lits, int32(len(lits)), uint32(t), 1)
		Gia_ManAppendCo(p, l)
		count++
	}
	fmt.Printf("Finished processing %d outputs.\n", count)
}

// Minimum-base compression retains the original order of supported variables.
func giaTruthMinBase(t []word, n int) ([]word, int) {
	used := []int{}
	for v := 0; v < n; v++ {
		has := false
		for a := 0; a < 1<<uint(n); a++ {
			b := a ^ (1 << uint(v))
			if (t[a>>6]>>uint(a&63))&1 != (t[b>>6]>>uint(b&63))&1 {
				has = true
				break
			}
		}
		if has {
			used = append(used, v)
		}
	}
	r := make([]word, giaTruthWordCount64(int32(len(used))))
	for a := 0; a < len(r)*64; a++ {
		old := 0
		for j, v := range used {
			if a&(1<<uint(j)) != 0 {
				old |= 1 << uint(v)
			}
		}
		if t[old>>6]&(1<<uint(old&63)) != 0 {
			r[a>>6] |= 1 << uint(a&63)
		}
	}
	return r, len(used)
}
func Gia_ManIsoNpnReduce(p *Gia_Man_t, equivs **Vec_Ptr_t, verbose int32) *Gia_Man_t {
	classes := [][]int32{}
	lookup := map[string]int{}
	supp := vec.Vec_IntAlloc(16)
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		o := Gia_ManPo(p, i)
		vec.Vec_IntClear(supp)
		Gia_ManIncrementTravId(p)
		Gia_ObjComputeTruth6CisSupport_rec(p, o.id, supp)
		leaves := vec.Vec_IntArray(supp)
		key := fmt.Sprintf("unique:%d", i)
		if len(leaves) <= 16 {
			t := giaNewTruthEval(p, int32(len(leaves)), leaves).node(o.id)
			t, n := giaTruthMinBase(t, len(leaves))
			if n < 2 {
				key = fmt.Sprint(n)
			} else {
				perm := make([]byte, n)
				canon.Canonicalize(t, int32(n), perm)
				b := make([]byte, 1+8*len(t))
				b[0] = byte(n)
				for k, w := range t {
					binary.LittleEndian.PutUint64(b[1+8*k:], w)
				}
				key = string(b)
			}
		}
		c, ok := lookup[key]
		if !ok {
			c = len(classes)
			lookup[key] = c
			classes = append(classes, nil)
		}
		classes[c] = append(classes[c], i)
	}
	out := Gia_ManStart(p.nObjs)
	ci := map[int32]int32{0: 0}
	out.pName, out.PSpec = p.pName, p.PSpec
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		ci[Gia_ManPi(p, i).id] = Gia_ManAppendCi(out)
	}
	// Follow register-input cones to a fixed point, as Gia_ManDupCones does.
	roots := make([]int32, 0, len(classes))
	for _, c := range classes {
		roots = append(roots, Gia_ManPo(p, c[0]).id)
	}
	seen := map[int32]bool{0: true}
	var collect func(int32)
	collect = func(id int32) {
		if seen[id] {
			return
		}
		seen[id] = true
		o := p.pObjs[id]
		if Gia_ObjIsCi(o) != 0 {
			idx := Gia_ObjCioId(o) - Gia_ManPiNum(p)
			if idx >= 0 {
				roots = append(roots, Gia_ManRi(p, idx).id)
				ci[id] = Gia_ManAppendCi(out)
			}
			return
		}
		for _, f := range giaObjectFanins(p, o) {
			collect(f)
		}
	}
	for i := 0; i < len(roots); i++ {
		collect(roots[i])
	}
	var clone func(int32) int32
	clone = func(id int32) int32 {
		if l, ok := ci[id]; ok {
			return l
		}
		o := p.pObjs[id]
		a := clone(Gia_ObjFaninId0(o, id)) ^ int32(o.fCompl0)
		var l int32
		if Gia_ObjIsCo(o) != 0 || Gia_ObjIsBuf(o) != 0 {
			l = a
		} else {
			b := clone(Gia_ObjFaninId1(o, id)) ^ int32(o.fCompl1)
			if int(id) < len(p.PMuxes) && p.PMuxes[id] != 0 {
				c := int32(p.PMuxes[id])
				control := clone(c>>1) ^ (c & 1)
				l = giaTruthAnd(out, giaTruthAnd(out, control, b, true)^1, giaTruthAnd(out, control^1, a, true)^1, true) ^ 1
			} else if Gia_ObjIsXor(o) != 0 {
				l = giaTruthAnd(out, giaTruthAnd(out, a, b^1, true)^1, giaTruthAnd(out, a^1, b, true)^1, true) ^ 1
			} else {
				l = giaTruthAnd(out, a, b, true)
			}
		}
		ci[id] = l
		return l
	}
	for _, root := range roots {
		Gia_ManAppendCo(out, clone(root))
	}
	out.nRegs = int32(len(roots) - len(classes))
	if equivs != nil {
		*equivs = vec.Vec_PtrAlloc(len(classes))
		for _, c := range classes {
			v := vec.Vec_IntAlloc(len(c))
			for _, i := range c {
				vec.Vec_IntPush(v, i)
			}
			vec.Vec_PtrPush(*equivs, v)
		}
	}
	if verbose != 0 {
		fmt.Printf("Nontrivial classes:\n")
		for _, c := range classes {
			if len(c) > 1 {
				fmt.Println(c)
			}
		}
	}
	return out
}
