package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

// Gia_ObjCheckMffc_rec temporarily dereferences the cone and records each
// reference so the caller can restore it, even if the limit is exceeded.
func Gia_ObjCheckMffc_rec(p *Gia_Man_t, obj *Gia_Obj_t, limit int32, nodes *Vec_Int_t) int32 {
	if Gia_ObjIsCi(obj) != 0 {
		return 1
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid MFFC node")
	}
	fans := []int32{Gia_ObjFaninId0(obj, obj.id), Gia_ObjFaninId1(obj, obj.id)}
	if Gia_ObjIsMuxId(p, obj.id) != 0 {
		fans = append(fans, Gia_ObjFaninId2(p, obj.id))
	}
	for _, fanin := range fans {
		vec.Vec_IntPush(nodes, fanin)
		if Gia_ObjRefDecId(p, fanin) == 0 && (int32(vec.Vec_IntSize(nodes)) > limit || Gia_ObjCheckMffc_rec(p, Gia_ManObj(p, fanin), limit, nodes) == 0) {
			return 0
		}
	}
	return 1
}
func Gia_ObjCheckMffc(p *Gia_Man_t, root *Gia_Obj_t, limit int32, nodes, leaves, inners *Vec_Int_t) int32 {
	vec.Vec_IntClear(nodes)
	ok := Gia_ObjCheckMffc_rec(p, root, limit, nodes)
	if ok != 0 {
		vec.Vec_IntClear(leaves)
		vec.Vec_IntClear(inners)
		vec.Vec_IntSort(nodes, 0)
		for _, id := range vec.Vec_IntArray(nodes) {
			dst := inners
			if Gia_ObjRefNumId(p, id) > 0 || Gia_ObjIsCi(Gia_ManObj(p, id)) != 0 {
				dst = leaves
			}
			if vec.Vec_IntSize(dst) == 0 || vec.Vec_IntEntryLast(dst) != id {
				vec.Vec_IntPush(dst, id)
			}
		}
		vec.Vec_IntPush(inners, root.id)
	}
	for _, id := range vec.Vec_IntArray(nodes) {
		Gia_ObjRefIncId(p, id)
	}
	return ok
}
func Gia_ManComputeMffcs(p *Gia_Man_t, limitMin, limitMax, suppMax, ratioBest int32) *Vec_Wec_t {
	if p.PMuxes == nil || limitMin < 0 || limitMax < 0 || suppMax < 0 || ratioBest < 0 || ratioBest > 100 {
		panic("gia: invalid MFFC parameters")
	}
	nodes, leaves, inners := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	result := vec.Vec_WecAlloc(1000)
	Gia_ManCreateRefs(p)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjRefNum(p, obj) == 0 || Gia_ObjCheckMffc(p, obj, limitMax, nodes, leaves, inners) == 0 {
			continue
		}
		if int32(vec.Vec_IntSize(inners)) < limitMin || int32(vec.Vec_IntSize(leaves)) > suppMax || vec.Vec_IntSize(leaves) == 0 {
			continue
		}
		m := vec.Vec_WecPushLevel(result)
		vec.Vec_IntPush(m, obj.id)
		vec.Vec_IntPush(m, int32(vec.Vec_IntSize(leaves)))
		vec.Vec_IntPush(m, int32(vec.Vec_IntSize(inners)))
		for _, id := range vec.Vec_IntArray(leaves) {
			vec.Vec_IntPush(m, id)
		}
		vec.Vec_IntPush(m, int32(1000*vec.Vec_IntSize(inners)/vec.Vec_IntSize(leaves)))
	}
	vec.Vec_WecSortByLastInt(result, 1)
	for i := 0; i < vec.Vec_WecSize(result); i++ {
		vec.Vec_IntPop(vec.Vec_WecEntry(result, i))
	}
	vec.Vec_WecShrink(result, int(ratioBest)*vec.Vec_WecSize(result)/100)
	return result
}
func Gia_ManPrintDivStats(p *Gia_Man_t, mffcs, pivots *Vec_Wec_t) {
	count, total, empty := vec.Vec_WecSize(mffcs), 0, 0
	for i := 0; i < count; i++ {
		m := vec.Vec_WecEntry(mffcs, i)
		n := vec.Vec_IntSize(m) - 3 - int(vec.Vec_IntEntry(m, 1)) - int(vec.Vec_IntEntry(m, 2))
		if n == 0 {
			empty++
		}
		total += n
	}
	ands := max(1, Gia_ManAndNum(p))
	fmt.Printf("Collected %d (%.1f %%) MFFCs and %d (%.1f %%) have no divisors (div ave for others is %.2f).\n", count, 100*float64(count)/float64(ands), empty, 100*float64(empty)/float64(ands), float64(total)/float64(max(1, count-empty)))
	fmt.Printf("Using %.2f MB for MFFCs and %.2f MB for pivots.   ", float64(vec.Vec_WecMemory(mffcs))/(1<<20), float64(vec.Vec_WecMemory(pivots))/(1<<20))
}
func Gia_ManAddDivisors(p *Gia_Man_t, mffcs *Vec_Wec_t) {
	pivots := vec.Vec_WecStart(int(p.nObjs))
	mapping := make([]int, int(p.nObjs))
	for i := range mapping {
		mapping[i] = -1
	}
	for i := 0; i < vec.Vec_WecSize(mffcs); i++ {
		m := vec.Vec_WecEntry(mffcs, i)
		root := vec.Vec_IntEntry(m, 0)
		if vec.Vec_IntSize(m) != 3+int(vec.Vec_IntEntry(m, 1)) {
			panic("gia: invalid MFFC")
		}
		mapping[root] = i
		for _, leaf := range vec.Vec_IntArray(m)[3:] {
			vec.Vec_IntPush(vec.Vec_WecEntry(pivots, int(leaf)), root)
		}
	}
	for i := 0; i < vec.Vec_WecSize(pivots); i++ {
		vec.Vec_IntSort(vec.Vec_WecEntry(pivots, i), 0)
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		fans := []int32{Gia_ObjFaninId0(obj, obj.id), Gia_ObjFaninId1(obj, obj.id)}
		if Gia_ObjIsMuxId(p, obj.id) != 0 {
			fans = append(fans, Gia_ObjFaninId2(p, obj.id))
		}
		common := append([]int32(nil), vec.Vec_IntArray(vec.Vec_WecEntry(pivots, int(fans[0])))...)
		for _, fan := range fans[1:] {
			other := vec.Vec_IntArray(vec.Vec_WecEntry(pivots, int(fan)))
			merged := make([]int32, 0)
			for i, j := 0, 0; i < len(common) && j < len(other); {
				if common[i] < other[j] {
					i++
				} else if common[i] > other[j] {
					j++
				} else {
					merged = append(merged, common[i])
					i++
					j++
				}
			}
			common = merged
		}
		if len(common) == 0 {
			continue
		}
		target := vec.Vec_WecEntry(pivots, int(obj.id))
		all := append(append([]int32(nil), vec.Vec_IntArray(target)...), common...)
		sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
		vec.Vec_IntClear(target)
		for _, id := range all {
			if vec.Vec_IntSize(target) == 0 || vec.Vec_IntEntryLast(target) != id {
				vec.Vec_IntPush(target, id)
			}
		}
		for _, id := range common {
			if mapping[id] < 0 {
				panic("gia: missing MFFC pivot")
			}
			vec.Vec_IntPush(vec.Vec_WecEntry(mffcs, mapping[id]), obj.id)
		}
	}
	Gia_ManPrintDivStats(p, mffcs, pivots)
}
