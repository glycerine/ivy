package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"time"
)

// Balance ordered supergates by level, preferring already hashed pairs.
func Gia_ManFindSharedNode(p *Gia_Man_t, super *Vec_Int_t, first int32) int32 {
	last := Vec_IntEntryLast(super)
	level := Gia_ObjLevelId(p, last>>1)
	for i := Vec_IntSize(super) - 1; i >= 0; i-- {
		candidate := Vec_IntEntry(super, i)
		if Gia_ObjLevelId(p, candidate>>1) != level {
			break
		}
		if first>>1 != candidate>>1 && Gia_ManHashLookupInt(p, first, candidate) == 0 {
			continue
		}
		Vec_IntWriteEntry(super, i, last)
		Vec_IntWriteEntry(super, Vec_IntSize(super)-1, candidate)
		break
	}
	return Vec_IntPop(super)
}
func Gia_ManPrepareLastTwo(p *Gia_Man_t, super *Vec_Int_t) {
	n := Vec_IntSize(super)
	if n == 2 {
		return
	}
	if n < 2 {
		panic("gia: short supergate")
	}
	level := Gia_ObjLevelId(p, Vec_IntEntry(super, n-2)>>1)
	stop := n - 3
	for ; stop >= 0; stop-- {
		if Gia_ObjLevelId(p, Vec_IntEntry(super, stop)>>1) != level {
			break
		}
	}
	if stop == n-3 {
		return
	}
	stop = max(stop, n-9)
	array := Vec_IntArray(super)
	for i := n - 1; i > stop; i-- {
		for k := i - 1; k > stop; k-- {
			a, b := array[i], array[k]
			if a>>1 != b>>1 && Gia_ManHashLookupInt(p, a, b) == 0 {
				continue
			}
			array[i], array[n-1] = array[n-1], array[i]
			array[k], array[n-2] = array[n-2], array[k]
		}
	}
}
func Gia_ManCreateGate(p *Gia_Man_t, obj *Gia_Obj_t, super *Vec_Int_t) {
	a, b := Vec_IntPop(super), Vec_IntPop(super)
	var literal int32
	if Gia_ObjIsXor(obj) == 0 {
		literal = Gia_ManHashAnd(p, a, b)
	} else if p.PMuxes != nil {
		literal = Gia_ManHashXorReal(p, a, b)
	} else {
		literal = Gia_ManHashXor(p, a, b)
	}
	Vec_IntPush(super, literal)
	Gia_ObjSetGateLevel(p, Gia_ManObj(p, literal>>1))
	for i := Vec_IntSize(super) - 1; i > 0; i-- {
		a, b := Vec_IntEntry(super, i), Vec_IntEntry(super, i-1)
		if Gia_ObjLevelId(p, a>>1) <= Gia_ObjLevelId(p, b>>1) {
			break
		}
		Vec_IntWriteEntry(super, i, b)
		Vec_IntWriteEntry(super, i-1, a)
	}
}
func Gia_ManBalanceGate(p *Gia_Man_t, obj *Gia_Obj_t, super *Vec_Int_t, literals []int32, n int32) int32 {
	if n < 1 || int(n) > len(literals) || Gia_ObjIsBuf(obj) != 0 {
		panic("gia: invalid balance supergate")
	}
	// Copy before clearing: callers may hand us a borrowed vector span.
	literals = append([]int32(nil), literals[:n]...)
	Vec_IntClear(super)
	if n > 2 {
		costs, order := make([]int32, n), make([]int32, n)
		for i, lit := range literals {
			costs[i] = Gia_ObjLevelId(p, lit>>1)
		}
		util.Abc_QuickSortCostData(costs, n, 1, make([]uint64, n), order)
		for _, i := range order {
			Vec_IntPush(super, literals[i])
		}
	} else {
		for _, lit := range literals {
			Vec_IntPush(super, lit)
		}
	}
	for Vec_IntSize(super) > 1 {
		if Gia_ObjIsXor(obj) == 0 {
			Gia_ManPrepareLastTwo(p, super)
		}
		Gia_ManCreateGate(p, obj, super)
	}
	return Vec_IntEntry(super, 0)
}
func Gia_ManBalance_rec(dst, src *Gia_Man_t, obj *Gia_Obj_t, strict int32) {
	if obj.Value != ^uint32(0) {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 {
		panic("gia: invalid balance recursion")
	}
	if Gia_ObjIsMux(src, obj) != 0 {
		for _, fanin := range []*Gia_Obj_t{Gia_ObjFanin0(obj), Gia_ObjFanin1(obj), Gia_ObjFanin2(src, obj)} {
			Gia_ManBalance_rec(dst, src, fanin, strict)
		}
		obj.Value = uint32(Gia_ManHashMuxReal(dst, Gia_ObjFanin2Copy(src, obj), Gia_ObjFanin1Copy(obj), Gia_ObjFanin0Copy(obj)))
	} else {
		Gia_ManSuperCollect(src, obj, strict)
		literals := append([]int32(nil), Vec_IntArray(src.vSuper)...)
		for i, lit := range literals {
			fanin := Gia_ManObj(src, lit>>1)
			Gia_ManBalance_rec(dst, src, fanin, strict)
			literals[i] = int32(fanin.Value) ^ (lit & 1)
		}
		obj.Value = uint32(Gia_ManBalanceGate(dst, obj, src.vSuper, literals, int32(len(literals))))
	}
	Gia_ObjSetGateLevel(dst, Gia_ManObj(dst, int32(obj.Value)>>1))
}
func giaBalanceStart(src *Gia_Man_t) *Gia_Man_t {
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	dst.EnableMuxes()
	Gia_ManCleanLevels(dst, src.nObjs)
	Gia_ManFillValue(src)
	src.pObjs[0].Value = 0
	for _, id := range Vec_IntArray(src.vCis) {
		src.pObjs[id].Value = uint32(Gia_ManAppendCi(dst))
	}
	return dst
}
func giaBalanceFinish(dst, src *Gia_Man_t) *Gia_Man_t {
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, src.nRegs)
	giaCopyInterfaceMetadata(dst, src)
	cleaned := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return cleaned
}
func giaBalanceInputs(src, dst *Gia_Man_t) {
	delay := src.And2Delay
	if delay == 0 {
		delay = 1
	}
	for i, id := range Vec_IntArray(dst.vCis) {
		var level int32
		if src.vCiArrs != nil {
			level = Vec_IntEntry(src.vCiArrs, i) / delay
		} else if src.VInArrs != nil {
			level = int32(Vec_FltEntry(src.VInArrs, i) / float32(delay))
		}
		Gia_ObjSetLevel(dst, dst.pObjs[id], level)
	}
}
func Gia_ManBalanceInt(src *Gia_Man_t, strict int32) *Gia_Man_t {
	src.Refs = nil
	Gia_ManCreateRefs(src)
	dst := giaBalanceStart(src)
	giaBalanceInputs(src, dst)
	Gia_ManHashStart(dst)
	for _, obj := range src.pObjs {
		if Gia_ObjIsBuf(obj) != 0 {
			Gia_ManBalance_rec(dst, src, Gia_ObjFanin0(obj), strict)
			obj.Value = uint32(Gia_ManAppendBuf(dst, Gia_ObjFanin0Copy(obj)))
			Gia_ObjSetGateLevel(dst, Gia_ManObj(dst, int32(obj.Value)>>1))
		}
	}
	for _, id := range Vec_IntArray(src.vCos) {
		obj := src.pObjs[id]
		Gia_ManBalance_rec(dst, src, Gia_ObjFanin0(obj), strict)
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	return giaBalanceFinish(dst, src)
}
func Gia_ManBalance(src *Gia_Man_t, simple, strict, verbose int32) *Gia_Man_t {
	var first *Gia_Man_t
	if simple != 0 {
		first = Gia_ManDup(src)
	} else {
		first = Gia_ManDupMuxes(src, 2)
	}
	Gia_ManTransferTiming(first, src)
	middle := Gia_ManBalanceInt(first, strict)
	Gia_ManTransferTiming(middle, first)
	Gia_ManStop(first)
	result := Gia_ManDupNoMuxes(middle, 0)
	Gia_ManTransferTiming(result, middle)
	Gia_ManStop(middle)
	if verbose != 0 {
		fmt.Printf("Balanced AIG: %d objects, %d levels\n", result.nObjs, Gia_ManLevelNum(result))
	}
	return result
}

// DAM stores size-prefixed literal/root lists. Slice offsets, not pointers,
// remain valid as either backing store grows during divisor extraction.
type Dam_Man_t struct {
	pGia                                     *Gia_Man_t
	vNod2Set, vDiv2Nod, vSetStore, vNodStore []int32
	vCounts                                  []float32
	vNodLevR, vDivLevR                       []int32
	vQue                                     *vec.Vec_Que_t
	vHash                                    *vec.Hash_IntMan_t
	clkStart                                 time.Time
	nLevelMax, nDivs, nAnds, nGain, nGainX   int32
}

func Dam_ObjHand(p *Dam_Man_t, id int32) int32 {
	if int(id) < len(p.vNod2Set) {
		return p.vNod2Set[id]
	}
	return 0
}
func Dam_ObjSet(p *Dam_Man_t, id int32) []int32 {
	if h := Dam_ObjHand(p, id); h != 0 {
		return p.vSetStore[h:]
	}
	return nil
}
func Dam_DivHand(p *Dam_Man_t, id int32) int32 {
	if int(id) < len(p.vDiv2Nod) {
		return p.vDiv2Nod[id]
	}
	return 0
}
func Dam_DivSet(p *Dam_Man_t, id int32) []int32 {
	if h := Dam_DivHand(p, id); h > 0 {
		return p.vNodStore[h:]
	}
	return nil
}
func Dam_ManAlloc(p *Gia_Man_t) *Dam_Man_t { return &Dam_Man_t{pGia: p, clkStart: time.Now()} }
func Dam_ManFree(p *Dam_Man_t) {
	if p != nil {
		*p = Dam_Man_t{}
	}
}
func Dam_ManCollectSets_rec(p *Dam_Man_t, id int32) {
	if id == 0 || Dam_ObjHand(p, id) != 0 {
		return
	}
	obj := p.pGia.pObjs[id]
	if Gia_ObjIsCi(obj) != 0 {
		return
	}
	if Gia_ObjIsBuf(obj) != 0 {
		Dam_ManCollectSets_rec(p, Gia_ObjFaninId0(obj, id))
		return
	}
	if Gia_ObjIsMux(p.pGia, obj) != 0 {
		if obj.fMark0 != 0 {
			return
		}
		obj.fMark0 = 1
		Dam_ManCollectSets_rec(p, Gia_ObjFaninId0(obj, id))
		Dam_ManCollectSets_rec(p, Gia_ObjFaninId1(obj, id))
		Dam_ManCollectSets_rec(p, Gia_ObjFaninId2(p.pGia, id))
		p.nAnds += 3
		return
	}
	Gia_ManSuperCollect(p.pGia, obj, 0)
	literals := append([]int32(nil), Vec_IntArray(p.pGia.vSuper)...)
	p.vNod2Set[id] = int32(len(p.vSetStore))
	p.vSetStore = append(p.vSetStore, int32(len(literals)))
	p.vSetStore = append(p.vSetStore, literals...)
	p.nAnds += (1 + 2*Gia_ObjIsXor(obj)) * int32(len(literals)-1)
	for _, lit := range literals {
		Dam_ManCollectSets_rec(p, lit>>1)
	}
}
func Dam_ManCollectSets(p *Dam_Man_t) {
	p.pGia.Refs = nil
	Gia_ManCreateRefs(p.pGia)
	p.vNod2Set = make([]int32, p.pGia.nObjs)
	p.vSetStore = []int32{-1}
	for _, id := range Vec_IntArray(p.pGia.vCos) {
		Dam_ManCollectSets_rec(p, Gia_ObjFaninId0(p.pGia.pObjs[id], id))
	}
	p.pGia.Refs = nil
	for _, obj := range p.pGia.pObjs {
		obj.fMark0 = 0
	}
}
func Dam_ManDivSlack(p *Dam_Man_t, a, b, reverse int32) int32 {
	return min(p.nLevelMax-reverse-max(Gia_ObjLevelId(p.pGia, a>>1), Gia_ObjLevelId(p.pGia, b>>1))-1-giaBool(a > b), 100)
}
func Dam_ManCreateMultiRefs(p *Dam_Man_t, andRefs, xorRefs **Vec_Int_t) {
	*andRefs = Vec_IntStart(2 * int(p.pGia.nObjs))
	*xorRefs = Vec_IntStart(int(p.pGia.nObjs))
	for _, obj := range p.pGia.pObjs {
		set := Dam_ObjSet(p, obj.id)
		if set == nil {
			continue
		}
		for _, lit := range set[1 : 1+set[0]] {
			if Gia_ObjIsXor(obj) != 0 {
				if lit&1 != 0 {
					panic("gia: complemented XOR divisor")
				}
				Vec_IntAddToEntry(*xorRefs, int(lit>>1), 1)
			} else {
				Vec_IntAddToEntry(*andRefs, int(lit), 1)
			}
		}
	}
}
func damPair(a, b int32, xor bool) (int32, int32) {
	if (a > b) != xor {
		return b, a
	}
	return a, b
}
func damGrow(values []int32, size int) []int32 {
	if len(values) < size {
		values = append(values, make([]int32, size-len(values))...)
	}
	return values
}
func Dam_ManCreatePairs(p *Dam_Man_t, verbose int32) {
	Dam_ManCollectSets(p)
	var andRefs, xorRefs *Vec_Int_t
	Dam_ManCreateMultiRefs(p, &andRefs, &xorRefs)
	hash := vec.Hash_IntManStart(int(p.pGia.nObjs) / 2)
	reverse := make([]int32, 1)
	type occurrence struct{ divisor, node int32 }
	var occurrences []occurrence
	for _, obj := range p.pGia.pObjs {
		set := Dam_ObjSet(p, obj.id)
		if set == nil {
			continue
		}
		xor := Gia_ObjIsXor(obj) != 0
		var candidates []int32
		for _, lit := range set[1 : 1+set[0]] {
			if xor && Vec_IntEntry(xorRefs, int(lit>>1)) > 1 || !xor && Vec_IntEntry(andRefs, int(lit)) > 1 {
				candidates = append(candidates, lit)
			}
		}
		for i, a := range candidates {
			for _, b := range candidates[i+1:] {
				a, b := damPair(a, b, xor)
				div := vec.Hash_Int2ManInsert(hash, a, b, 0)
				vec.Hash_Int2ObjInc(hash, div)
				occurrences = append(occurrences, occurrence{div, obj.id})
				reverse = damGrow(reverse, int(div)+1)
				reverse[div] = max(reverse[div], p.vNodLevR[obj.id])
			}
		}
	}
	p.vHash = vec.Hash_IntManStart(vec.Hash_IntManEntryNum(hash))
	p.vCounts = []float32{float32(util.ABC_INFINITY)}
	p.vDiv2Nod = []int32{util.ABC_INFINITY}
	p.vNodStore = []int32{-1}
	p.vQue = vec.Vec_QueAlloc(vec.Hash_IntManEntryNum(hash))
	vec.Vec_QueSetPriority(p.vQue, &p.vCounts)
	remap := make([]int32, vec.Hash_IntManEntryNum(hash)+1)
	for id := 1; id < len(remap); id++ {
		refs := vec.Hash_IntObjData2(hash, int32(id))
		if refs < 2 {
			continue
		}
		a, b := vec.Hash_IntObjData0(hash, int32(id)), vec.Hash_IntObjData1(hash, int32(id))
		div := vec.Hash_Int2ManInsert(p.vHash, a, b, 0)
		remap[id] = div
		p.vCounts = append(p.vCounts, float32(float64(refs)+0.005*float64(Dam_ManDivSlack(p, a, b, reverse[id]))))
		vec.Vec_QuePush(p.vQue, int(div))
		p.vDiv2Nod = append(p.vDiv2Nod, int32(len(p.vNodStore)))
		p.vNodStore = append(p.vNodStore, make([]int32, refs+1)...)
	}
	for _, occ := range occurrences {
		if id := remap[occ.divisor]; id != 0 {
			set := Dam_DivSet(p, id)
			set[0]++
			set[set[0]] = occ.node
		}
	}
	p.vDivLevR = make([]int32, len(p.vCounts)*2)
	if verbose != 0 {
		fmt.Printf("DAM: %d candidate pairs, %d shared divisors\n", vec.Hash_IntManEntryNum(hash), len(p.vCounts)-1)
	}
}

func Dam_ManMultiAig_rec(manager *Dam_Man_t, dst, src *Gia_Man_t, obj *Gia_Obj_t) {
	if obj.Value != ^uint32(0) {
		return
	}
	set := Dam_ObjSet(manager, obj.id)
	if set == nil {
		a, b, c, kind := src.Gate(obj.id)
		Dam_ManMultiAig_rec(manager, dst, src, Gia_ObjFanin0(obj))
		Dam_ManMultiAig_rec(manager, dst, src, Gia_ObjFanin1(obj))
		switch kind {
		case 4:
			Dam_ManMultiAig_rec(manager, dst, src, Gia_ObjFanin2(src, obj))
			obj.Value = uint32(Gia_ManHashMuxReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c)))
		case 3:
			obj.Value = uint32(Gia_ManHashXorReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b)))
		case 2:
			obj.Value = uint32(Gia_ManHashAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b)))
		default:
			panic("gia: invalid DAM reconstruction gate")
		}
	} else {
		literals := set[1 : 1+set[0]]
		for i, lit := range literals {
			fanin := src.pObjs[lit>>1]
			Dam_ManMultiAig_rec(manager, dst, src, fanin)
			literals[i] = int32(fanin.Value) ^ (lit & 1)
		}
		obj.Value = uint32(Gia_ManBalanceGate(dst, obj, src.vSuper, literals, int32(len(literals))))
	}
	Gia_ObjSetGateLevel(dst, dst.pObjs[obj.Value>>1])
}
func Dam_ManMultiAig(p *Dam_Man_t) *Gia_Man_t {
	src := p.pGia
	dst := giaBalanceStart(src)
	for i, id := range Vec_IntArray(src.vCis) {
		Gia_ObjSetLevel(dst, Gia_ManCi(dst, int32(i)), Gia_ObjLevelId(src, id))
	}
	Gia_ManHashStart(dst)
	for _, obj := range src.pObjs {
		if Gia_ObjIsBuf(obj) != 0 {
			Dam_ManMultiAig_rec(p, dst, src, Gia_ObjFanin0(obj))
			obj.Value = uint32(Gia_ManAppendBuf(dst, Gia_ObjFanin0Copy(obj)))
			Gia_ObjSetGateLevel(dst, dst.pObjs[obj.Value>>1])
		}
	}
	for _, id := range Vec_IntArray(src.vCos) {
		obj := src.pObjs[id]
		Dam_ManMultiAig_rec(p, dst, src, Gia_ObjFanin0(obj))
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	return giaBalanceFinish(dst, src)
}
func Dam_PrintDiv(p *Dam_Man_t, id int32) {
	if id == 0 {
		fmt.Printf("Final statistics after extracting %d divisors: ", p.nDivs)
	} else {
		a, b := vec.Hash_IntObjData0(p.vHash, id), vec.Hash_IntObjData1(p.vHash, id)
		fmt.Printf("Div %d: (%d,%d), weight %.2f; ", id, a, b, p.vCounts[id])
	}
	fmt.Printf("Divs=%d Ands=%d Time=%s\n", vec.Hash_IntManEntryNum(p.vHash), p.nAnds-p.nGain, time.Since(p.clkStart))
}
func Dam_PrintQue(p *Dam_Man_t) {
	for id := 1; id <= vec.Hash_IntManEntryNum(p.vHash); id++ {
		fmt.Printf("Div %d: (%d,%d), weight %.2f\n", id, vec.Hash_IntObjData0(p.vHash, int32(id)), vec.Hash_IntObjData1(p.vHash, int32(id)), p.vCounts[id])
	}
}
func Dam_ManUpdateNode(p *Dam_Man_t, id, a, b, replacement int32, divisors *Vec_Int_t) int32 {
	set := Dam_ObjSet(p, id)
	if set == nil {
		return 0
	}
	foundA, foundB := false, false
	for _, lit := range set[1 : 1+set[0]] {
		foundA = foundA || lit == a
		foundB = foundB || lit == b
	}
	if !foundA || !foundB {
		return 0
	}
	Vec_IntPush(divisors, -id)
	count := 1
	for _, lit := range set[1 : 1+set[0]] {
		if lit == a || lit == b {
			continue
		}
		set[count] = lit
		count++
		present := 0
		for index, other := range []int32{a, b} {
			x, y := damPair(lit, other, a > b)
			div := *vec.Hash_Int2ManLookup(p.vHash, x, y)
			if div > 0 {
				p.vCounts[div]--
				if vec.Vec_QueIsMember(p.vQue, int(div)) != 0 {
					vec.Vec_QueUpdate(p.vQue, int(div))
					present |= 1 << uint(index)
				}
			}
		}
		if present != 3 {
			continue
		}
		x, y := damPair(lit, replacement, a > b)
		div := vec.Hash_Int2ManInsert(p.vHash, x, y, 0)
		vec.Hash_Int2ObjInc(p.vHash, div)
		Vec_IntPush(divisors, div)
		p.vDivLevR = damGrow(p.vDivLevR, int(div)+1)
		p.vDivLevR[div] = max(p.vDivLevR[div], p.vNodLevR[id])
	}
	set[count] = replacement
	set[0] = int32(count)
	return 1
}
func Dam_ManUpdate(p *Dam_Man_t, divisor int32) {
	a, b := vec.Hash_IntObjData0(p.vHash, divisor), vec.Hash_IntObjData1(p.vHash, divisor)
	xor := a > b
	var replacement int32
	if xor {
		replacement = Gia_ManAppendXorReal(p.pGia, a, b)
	} else {
		replacement = Gia_ManAppendAnd(p.pGia, a, b)
	}
	Gia_ObjSetGateLevel(p.pGia, p.pGia.pObjs[replacement>>1])
	// Appending node lists below may reallocate storage; snapshot the current list.
	set := Dam_DivSet(p, divisor)
	nodes := append([]int32(nil), set[1:1+set[0]]...)
	first := vec.Hash_IntManEntryNum(p.vHash) + 1
	divisors := p.pGia.vSuper
	Vec_IntClear(divisors)
	var present int32
	for _, id := range nodes {
		present += Dam_ManUpdateNode(p, id, a, b, replacement, divisors)
	}
	stop := vec.Hash_IntManEntryNum(p.vHash) + 1
	p.vCounts = append(p.vCounts, make([]float32, stop-len(p.vCounts))...)
	for len(p.vDiv2Nod) < stop {
		p.vDiv2Nod = append(p.vDiv2Nod, -1)
	}
	for id := first; id < stop; id++ {
		refs := vec.Hash_IntObjData2(p.vHash, int32(id))
		if refs < 2 {
			continue
		}
		x, y := vec.Hash_IntObjData0(p.vHash, int32(id)), vec.Hash_IntObjData1(p.vHash, int32(id))
		p.vCounts[id] = float32(float64(refs) + 0.001*float64(Dam_ManDivSlack(p, x, y, p.vDivLevR[id])))
		vec.Vec_QuePush(p.vQue, id)
		p.vDiv2Nod[id] = int32(len(p.vNodStore))
		p.vNodStore = append(p.vNodStore, make([]int32, refs+1)...)
	}
	var node int32
	for _, div := range Vec_IntArray(divisors) {
		if div < 0 {
			node = -div
			continue
		}
		if p.vDiv2Nod[div] < 0 {
			continue
		}
		set := Dam_DivSet(p, div)
		set[0]++
		set[set[0]] = node
	}
	p.vCounts[divisor] = 0
	weight := int32(1)
	if xor {
		weight = 3
	}
	p.nGain += weight * (present - 1)
	if xor {
		p.nGainX += 3 * (present - 1)
	}
	p.nDivs++
}
func Dam_ManAreaBalanceInt(src *Gia_Man_t, inputs *Vec_Int_t, limit, verbose, veryVerbose int32) *Gia_Man_t {
	p := Dam_ManAlloc(src)
	p.nLevelMax = Gia_ManSetLevels(src, inputs)
	p.vNodLevR = append([]int32(nil), Vec_IntArray(Gia_ManReverseLevel(src))...)
	Dam_ManCreatePairs(p, verbose)
	for i := int32(0); i < limit && vec.Vec_QueSize(p.vQue) > 0 && vec.Vec_QueTopPriority(p.vQue) >= 2; i++ {
		divisor := int32(vec.Vec_QuePop(p.vQue))
		if veryVerbose != 0 {
			Dam_PrintDiv(p, divisor)
		}
		Dam_ManUpdate(p, divisor)
	}
	if verbose != 0 || veryVerbose != 0 {
		Dam_PrintDiv(p, 0)
	}
	result := Dam_ManMultiAig(p)
	Dam_ManFree(p)
	return result
}
func Gia_ManAreaBalance(src *Gia_Man_t, simple, limit, verbose, veryVerbose int32) *Gia_Man_t {
	// Set source input arrivals before recording the levels used after coarsening.
	if src.vCiArrs != nil || src.VInArrs != nil {
		Gia_ManCleanLevels(src, src.nObjs)
		giaBalanceInputs(src, src)
	}
	if src.PManTime != nil && src.vLevels == nil {
		Gia_ManLevelWithBoxes(src)
	}
	inputs := Gia_ManGetCiLevels(src)
	var initial *Gia_Man_t
	if Gia_ManHasMapping(src) != 0 {
		graph := dau.Dsm_ManDeriveGia(src, 0)
		initial = Gia_ManFromAig(graph)
	} else {
		initial = Gia_ManDup(src)
	}
	Gia_ManTransferTiming(initial, src)
	var coarse *Gia_Man_t
	if simple != 0 {
		coarse = Gia_ManDup(initial)
	} else {
		coarse = Gia_ManDupMuxes(initial, 2)
	}
	Gia_ManTransferTiming(coarse, initial)
	Gia_ManStop(initial)
	middle := Dam_ManAreaBalanceInt(coarse, inputs, limit, verbose, veryVerbose)
	Gia_ManTransferTiming(middle, coarse)
	Gia_ManStop(coarse)
	result := Gia_ManDupNoMuxes(middle, 0)
	Gia_ManTransferTiming(result, middle)
	Gia_ManStop(middle)
	if Gia_ManIsNormalized(result) == 0 {
		old := result
		result = Gia_ManDupNormalize(old, 0)
		Gia_ManTransferTiming(result, old)
		Gia_ManStop(old)
	}
	return result
}
