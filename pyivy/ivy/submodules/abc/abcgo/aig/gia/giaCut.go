package gia

// Stand-alone priority cut enumeration, adapted from giaCut.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"time"
)

const (
	GIA_MAX_CUTSIZE  = 8
	GIA_MAX_CUTNUM   = 51
	GIA_MAX_TT_WORDS = 4
	GIA_CUT_NO_LEAF  = 15
)

type Gia_Cut_t struct {
	Sign                 uint64
	iFunc, Cost, CostLev int32
	nTreeLeaves, nLeaves uint32
	pLeaves              [GIA_MAX_CUTSIZE]int32
}
type Gia_Sto_t struct {
	nCutSize, nCutNum, fCutMin, fTruthMin, fVerbose int32
	pGia                                            *Gia_Man_t
	vRefs                                           *vec.Vec_Int_t
	vCuts                                           *vec.Vec_Wec_t
	vTtMem                                          *vec.Vec_Mem_t
	pCuts                                           [3][GIA_MAX_CUTNUM]Gia_Cut_t
	ppCuts                                          [GIA_MAX_CUTNUM]*Gia_Cut_t
	nCutsR, Pivot, iCutBest, nCutsOver              int32
	CutCount                                        [4]float64
	clkStart                                        time.Time
}

func Gia_CutTruth(p *Gia_Sto_t, cut *Gia_Cut_t) []uint64 {
	return vec.Vec_MemReadEntry(p.vTtMem, int(cut.iFunc>>1))
}
func Gia_CutGetSign(cut *Gia_Cut_t) uint64 {
	var sign uint64
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		sign |= 1 << uint(id&63)
	}
	return sign
}
func Gia_CutCheck(base, cut *Gia_Cut_t) int32 {
	i, j := 0, 0
	for i < int(base.nLeaves) && j < int(cut.nLeaves) {
		if base.pLeaves[i] == cut.pLeaves[j] {
			j++
		}
		i++
	}
	return giaBool(j == int(cut.nLeaves))
}
func Gia_CutSetCheckArray(cuts []*Gia_Cut_t, n int32) int32 {
	if n <= 0 || int(n) > len(cuts) {
		panic("gia: invalid cut set")
	}
	for i, cut := range cuts[:n] {
		if cut.nLeaves > GIA_MAX_CUTSIZE || cut.Sign != Gia_CutGetSign(cut) {
			panic("gia: invalid cut signature")
		}
		for j := 1; j < int(cut.nLeaves); j++ {
			if cut.pLeaves[j-1] >= cut.pLeaves[j] {
				panic("gia: invalid cut ordering")
			}
		}
		for j, other := range cuts[:n] {
			if i != j && Gia_CutCheck(cut, other) != 0 {
				panic("gia: contained cut")
			}
		}
	}
	return 1
}
func Gia_CutMergeOrder(a, b, out *Gia_Cut_t, limit int32) int32 {
	if limit < 0 || limit > GIA_MAX_CUTSIZE {
		panic("gia: invalid cut size")
	}
	i, j, k := 0, 0, 0
	for i < int(a.nLeaves) || j < int(b.nLeaves) {
		if k == int(limit) {
			return 0
		}
		var value int32
		if j == int(b.nLeaves) || i < int(a.nLeaves) && a.pLeaves[i] < b.pLeaves[j] {
			value = a.pLeaves[i]
			i++
		} else if i == int(a.nLeaves) || b.pLeaves[j] < a.pLeaves[i] {
			value = b.pLeaves[j]
			j++
		} else {
			value = a.pLeaves[i]
			i++
			j++
		}
		out.pLeaves[k] = value
		k++
	}
	out.nLeaves = uint32(k)
	out.iFunc = -1
	out.Sign = a.Sign | b.Sign
	return 1
}
func Gia_CutMergeOrder2(a, b, out *Gia_Cut_t, limit int32) int32 {
	return Gia_CutMergeOrder(a, b, out, limit)
}
func Gia_CutSetCutIsContainedOrder(base, cut *Gia_Cut_t) int32 { return Gia_CutCheck(base, cut) }
func Gia_CutSetLastCutIsContained(cuts []*Gia_Cut_t, n int32) int32 {
	last := cuts[n]
	for _, cut := range cuts[:n] {
		if cut.nLeaves <= last.nLeaves && cut.Sign&last.Sign == cut.Sign && Gia_CutCheck(last, cut) != 0 {
			return 1
		}
	}
	return 0
}
func Gia_CutCompare(a, b *Gia_Cut_t) int32 {
	if a.nTreeLeaves < b.nTreeLeaves {
		return -1
	}
	if a.nTreeLeaves > b.nTreeLeaves {
		return 1
	}
	if a.nLeaves < b.nLeaves {
		return -1
	}
	if a.nLeaves > b.nLeaves {
		return 1
	}
	return 0
}
func Gia_CutSetLastCutContains(cuts []*Gia_Cut_t, n int32) int32 {
	last := cuts[n]
	for i := int32(0); i < n; i++ {
		cut := cuts[i]
		if last.nLeaves < cut.nLeaves && last.Sign&cut.Sign == last.Sign && Gia_CutCheck(cut, last) != 0 {
			cut.nLeaves = GIA_CUT_NO_LEAF
		}
	}
	k := int32(0)
	for i := int32(0); i <= n; i++ {
		if cuts[i].nLeaves == GIA_CUT_NO_LEAF {
			continue
		}
		if k < i {
			cuts[k], cuts[i] = cuts[i], cuts[k]
		}
		k++
	}
	return k - 1
}

func Gia_CutSetSortByCost(cuts []*Gia_Cut_t, n int32) {
	for i := n; i > 0; i-- {
		if Gia_CutCompare(cuts[i-1], cuts[i]) < 0 {
			return
		}
		cuts[i-1], cuts[i] = cuts[i], cuts[i-1]
	}
}
func Gia_CutSetAddCut(cuts []*Gia_Cut_t, n, limit int32) int32 {
	if n == 0 {
		return 1
	}
	n = Gia_CutSetLastCutContains(cuts, n)
	Gia_CutSetSortByCost(cuts, n)
	return min(n+1, limit-1)
}
func Gia_CutComputeTruth6(p *Gia_Sto_t, a, b *Gia_Cut_t, complA, complB int32, out *Gia_Cut_t, xor int32) int32 {
	return Gia_CutComputeTruth(p, a, b, complA, complB, out, xor)
}
func Gia_CutComputeTruth(p *Gia_Sto_t, a, b *Gia_Cut_t, complA, complB int32, out *Gia_Cut_t, xor int32) int32 {
	words := max(1, 1<<uint(max(0, int(p.nCutSize)-6)))
	ta := append([]uint64(nil), Gia_CutTruth(p, a)...)
	tb := append([]uint64(nil), Gia_CutTruth(p, b)...)
	if (a.iFunc&1)^complA != 0 {
		for i := range ta {
			ta[i] = ^ta[i]
		}
	}
	if (b.iFunc&1)^complB != 0 {
		for i := range tb {
			tb[i] = ^tb[i]
		}
	}
	util.Abc_TtExpand(ta, max(6, p.nCutSize), a.pLeaves[:], int32(a.nLeaves), out.pLeaves[:], int32(out.nLeaves))
	util.Abc_TtExpand(tb, max(6, p.nCutSize), b.pLeaves[:], int32(b.nLeaves), out.pLeaves[:], int32(out.nLeaves))
	result := make([]uint64, words)
	for i := range result {
		if xor != 0 {
			result[i] = ta[i] ^ tb[i]
		} else {
			result[i] = ta[i] & tb[i]
		}
	}
	complement := int32(result[0] & 1)
	if complement != 0 {
		for i := range result {
			result[i] = ^result[i]
		}
	}
	old := out.nLeaves
	if p.fTruthMin != 0 {
		out.nLeaves = uint32(util.Abc_TtMinBase(result, out.pLeaves[:], int32(out.nLeaves), max(6, p.nCutSize)))
	}
	id := vec.Vec_MemHashInsert(p.vTtMem, result)
	out.iFunc = int32(id<<1) | complement
	return giaBool(out.nLeaves < old)
}
func Gia_CutCountBits(value uint64) int32 { return int32(bits.OnesCount64(value)) }
func Gia_CutAddUnit(p *Gia_Sto_t, id int32) {
	list := vec.Vec_WecEntry(p.vCuts, int(id))
	if vec.Vec_IntSize(list) == 0 {
		vec.Vec_IntPush(list, 1)
	} else {
		vec.Vec_IntAddToEntry(list, 0, 1)
	}
	vec.Vec_IntPush(list, 1)
	vec.Vec_IntPush(list, id)
	vec.Vec_IntPush(list, 2)
}
func Gia_CutAddZero(p *Gia_Sto_t, id int32) {
	list := vec.Vec_WecEntry(p.vCuts, int(id))
	if vec.Vec_IntSize(list) != 0 {
		panic("gia: duplicate zero cut")
	}
	vec.Vec_IntPush(list, 1)
	vec.Vec_IntPush(list, 0)
	vec.Vec_IntPush(list, 0)
}
func Gia_CutTreeLeaves(p *Gia_Sto_t, cut *Gia_Cut_t) int32 {
	var n int32
	for _, id := range cut.pLeaves[:cut.nLeaves] {
		if vec.Vec_IntEntry(p.vRefs, int(id)) == 1 {
			n++
		}
	}
	return n
}
func Gia_StoPrepareSet(p *Gia_Sto_t, id, index int32) int32 {
	data := vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, int(id)))
	if len(data) == 0 {
		panic("gia: missing input cut set")
	}
	count := data[0]
	pos := 1
	for i := int32(0); i < count; i++ {
		cut := &p.pCuts[index][i]
		n := data[pos]
		pos++
		cut.nLeaves = uint32(n)
		copy(cut.pLeaves[:], data[pos:pos+int(n)])
		pos += int(n)
		cut.iFunc = data[pos]
		pos++
		cut.Sign = Gia_CutGetSign(cut)
		cut.nTreeLeaves = uint32(Gia_CutTreeLeaves(p, cut))
	}
	return count
}
func Gia_StoInitResult(p *Gia_Sto_t) {
	for i := range p.ppCuts {
		p.ppCuts[i] = &p.pCuts[2][i]
	}
}
func Gia_StoStoreResult(p *Gia_Sto_t, id int32, cuts []*Gia_Cut_t, n int32) {
	list := vec.Vec_WecEntry(p.vCuts, int(id))
	vec.Vec_IntPush(list, n)
	for _, cut := range cuts[:n] {
		vec.Vec_IntPush(list, int32(cut.nLeaves))
		for _, leaf := range cut.pLeaves[:cut.nLeaves] {
			vec.Vec_IntPush(list, leaf)
		}
		vec.Vec_IntPush(list, cut.iFunc)
	}
}
func Gia_CutPrint(p *Gia_Sto_t, id int32, cut *Gia_Cut_t) {
	if cut == nil {
		fmt.Println("No cut.")
		return
	}
	fmt.Printf("%d  %v  Cost = %3d  CostL = %3d  Tree = %d\n", cut.nLeaves, cut.pLeaves[:cut.nLeaves], cut.Cost, cut.CostLev, cut.nTreeLeaves)
}

func Gia_StoMergeCuts(p *Gia_Sto_t, id int32) {
	obj := Gia_ManObj(p.pGia, id)
	a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	if Gia_ObjIsBuf(obj) != 0 || Gia_ObjIsMux(p.pGia, obj) != 0 {
		panic("gia: cut enumeration requires non-mux AND/XOR gates")
	}
	n0, n1 := Gia_StoPrepareSet(p, a, 0), Gia_StoPrepareSet(p, b, 1)
	Gia_StoInitResult(p)
	p.CutCount[0] += float64(n0 * n1)
	var count int32
	for i := int32(0); i < n0; i++ {
		for j := int32(0); j < n1; j++ {
			x, y := &p.pCuts[0][i], &p.pCuts[1][j]
			candidate := p.ppCuts[count]
			if int32(x.nLeaves+y.nLeaves) > p.nCutSize && Gia_CutCountBits(x.Sign|y.Sign) > p.nCutSize {
				continue
			}
			p.CutCount[1]++
			if Gia_CutMergeOrder(x, y, candidate, p.nCutSize) == 0 || Gia_CutSetLastCutIsContained(p.ppCuts[:], count) != 0 {
				continue
			}
			p.CutCount[2]++
			if p.fCutMin != 0 && Gia_CutComputeTruth(p, x, y, Gia_ObjFaninC0(obj), Gia_ObjFaninC1(obj), candidate, Gia_ObjIsXor(obj)) != 0 {
				candidate.Sign = Gia_CutGetSign(candidate)
			}
			candidate.nTreeLeaves = uint32(Gia_CutTreeLeaves(p, candidate))
			count = Gia_CutSetAddCut(p.ppCuts[:], count, p.nCutNum)
		}
	}
	p.CutCount[3] += float64(count)
	if count == p.nCutNum-1 {
		p.nCutsOver++
	}
	p.nCutsR = count
	p.Pivot = id
	if count <= 0 || count >= p.nCutNum {
		panic("gia: invalid computed cut count")
	}
	Gia_CutSetCheckArray(p.ppCuts[:], count)
	Gia_StoStoreResult(p, id, p.ppCuts[:], count)
	if count > 1 || p.ppCuts[0].nLeaves > 1 {
		Gia_CutAddUnit(p, id)
	}
}
func Gia_StoAlloc(g *Gia_Man_t, size, num, cutMin, truthMin, verbose int32) *Gia_Sto_t {
	if size <= 1 || size > GIA_MAX_CUTSIZE || num <= 1 || num >= GIA_MAX_CUTNUM {
		panic("gia: invalid cut enumeration parameters")
	}
	p := &Gia_Sto_t{pGia: g, nCutSize: size, nCutNum: num, fCutMin: cutMin, fTruthMin: truthMin, fVerbose: verbose, clkStart: time.Now(), vRefs: vec.Vec_IntAlloc(int(g.nObjs)), vCuts: vec.Vec_WecStart(int(g.nObjs))}
	if cutMin != 0 {
		p.vTtMem = vec.Vec_MemAllocForTT(int(size), 0)
	}
	return p
}
func Gia_StoFree(p *Gia_Sto_t) {
	if p != nil {
		*p = Gia_Sto_t{}
	}
}
func Gia_StoComputeCutsConst0(p *Gia_Sto_t, id int32) { Gia_CutAddZero(p, id) }
func Gia_StoComputeCutsCi(p *Gia_Sto_t, id int32)     { Gia_CutAddUnit(p, id) }
func Gia_StoComputeCutsNode(p *Gia_Sto_t, id int32)   { Gia_StoMergeCuts(p, id) }
func Gia_StoRefObj(p *Gia_Sto_t, id int32) {
	if int(id) != vec.Vec_IntSize(p.vRefs) {
		panic("gia: reference objects must be appended in order")
	}
	vec.Vec_IntPush(p.vRefs, 0)
	for _, fanin := range giaObjectFanins(p.pGia, Gia_ManObj(p.pGia, id)) {
		vec.Vec_IntAddToEntry(p.vRefs, int(fanin), 1)
	}
}
func Gia_StoComputeCuts(g *Gia_Man_t) {
	p := Gia_StoAlloc(g, 6, 25, 1, 1, 1)
	defer Gia_StoFree(p)
	for id := int32(0); id < g.nObjs; id++ {
		Gia_StoRefObj(p, id)
	}
	Gia_StoComputeCutsConst0(p, 0)
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		Gia_StoComputeCutsCi(p, Gia_ManCi(g, i).id)
	}
	for id := int32(1); id < g.nObjs; id++ {
		if Gia_ObjIsAnd(g.pObjs[id]) != 0 {
			Gia_StoComputeCutsNode(p, id)
		}
	}
	if p.fVerbose != 0 {
		fmt.Printf("Running cut computation with CutSize = %d CutNum = %d\nCutPair = %.0f Merge = %.0f Eval = %.0f Cut = %.0f\nThe number of nodes with cut count over the limit = %d. Time: %s\n", p.nCutSize, p.nCutNum, p.CutCount[0], p.CutCount[1], p.CutCount[2], p.CutCount[3], p.nCutsOver, time.Since(p.clkStart))
	}
}
