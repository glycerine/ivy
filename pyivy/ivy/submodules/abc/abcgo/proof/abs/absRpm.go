package abs

// Structural reparameterization ported from absRpm.c by Alan Mishchenko,
// UC Berkeley (2005). The source graph's marks and values are scratch state.
import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ObjDom(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t) int32         { return gia.Gia_ObjDom(p, obj) }
func Gia_ObjSetDom(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, dom int32) { gia.Gia_ObjSetDom(p, obj, dom) }
func Gia_ManAddDom(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, dom int32) { gia.Gia_ManAddDom(p, obj, dom) }

// Gia_ManComputeDoms derives nearest single-node dominators, stopping at marked
// reparameterization inputs and, when available, unreferenced gates.
func Gia_ManComputeDoms(p *gia.Gia_Man_t) {
	gia.Gia_ManResetDoms(p)
	for id := gia.Gia_ManObjNum(p) - 1; id > 0; id-- {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsCi(obj) != 0 || gia.Gia_ObjMark1(obj) != 0 {
			continue
		}
		if p.Refs != nil && gia.Gia_ObjIsAnd(obj) != 0 && gia.Gia_ObjRefNum(p, obj) == 0 {
			continue
		}
		if gia.Gia_ObjIsCo(obj) != 0 {
			Gia_ObjSetDom(p, obj, id)
			Gia_ManAddDom(p, gia.Gia_ObjFanin0(obj), id)
			continue
		}
		if gia.Gia_ObjIsAnd(obj) == 0 {
			panic("abs: expected AND in dominator traversal")
		}
		a, b, c, kind := p.Gate(id)
		Gia_ManAddDom(p, gia.Gia_ManObj(p, a>>1), id)
		if kind == 2 || kind == 3 || kind == 4 {
			Gia_ManAddDom(p, gia.Gia_ManObj(p, b>>1), id)
		}
		if kind == 4 {
			Gia_ManAddDom(p, gia.Gia_ManObj(p, c>>1), id)
		}
	}
}

func Gia_ManCreateSupps(p *gia.Gia_Man_t, verbose int32) *vec.Vec_Wec_t {
	start := time.Now()
	supports := vec.Vec_WecStart(int(gia.Gia_ManObjNum(p)))
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))
		vec.Vec_IntPush(vec.Vec_WecEntry(supports, int(id)), i)
	}
	scratch := vec.Vec_IntAlloc(16)
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b, c, kind := p.Gate(id)
		if kind == 5 {
			vec.Vec_IntTwoMerge2(vec.Vec_WecEntry(supports, int(a>>1)), vec.Vec_WecEntry(supports, 0), vec.Vec_WecEntry(supports, int(id)))
		} else if kind == 4 {
			vec.Vec_IntTwoMerge2(vec.Vec_WecEntry(supports, int(a>>1)), vec.Vec_WecEntry(supports, int(b>>1)), scratch)
			vec.Vec_IntTwoMerge2(scratch, vec.Vec_WecEntry(supports, int(c>>1)), vec.Vec_WecEntry(supports, int(id)))
		} else {
			vec.Vec_IntTwoMerge2(vec.Vec_WecEntry(supports, int(a>>1)), vec.Vec_WecEntry(supports, int(b>>1)), vec.Vec_WecEntry(supports, int(id)))
		}
	}
	if verbose != 0 {
		fmt.Printf("Support computation: %v\n", time.Since(start))
	}
	return supports
}

func Gia_ManDomTest(p *gia.Gia_Man_t) {
	supports := Gia_ManCreateSupps(p, 1)
	dominated := vec.Vec_WecStart(int(gia.Gia_ManObjNum(p)))
	doms, supp := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	gia.Gia_ManCreateRefs(p)
	Gia_ManComputeDoms(p)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		id := Gia_ObjDom(p, gia.Gia_ManCi(p, i))
		for id >= 0 && gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, id)) != 0 {
			vec.Vec_IntPush(vec.Vec_WecEntry(dominated, int(id)), i)
			next := Gia_ObjDom(p, gia.Gia_ManObj(p, id))
			if next == id {
				break
			}
			id = next
		}
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, id)) != 0 && vec.Vec_IntEqual(vec.Vec_WecEntry(supports, int(id)), vec.Vec_WecEntry(dominated, int(id))) != 0 {
			vec.Vec_IntPush(doms, id)
		}
	}
	for _, id := range vec.Vec_IntArray(doms) {
		gia.Gia_ObjRefInc(p, gia.Gia_ManObj(p, id))
	}
	for _, id := range vec.Vec_IntArray(doms) {
		size := gia.Gia_NodeMffcSizeSupp(p, gia.Gia_ManObj(p, id), supp)
		fmt.Printf("%d(%d:%d) ", id, vec.Vec_IntSize(supp), size)
	}
	fmt.Println()
	for _, id := range vec.Vec_IntArray(doms) {
		gia.Gia_ObjRefDec(p, gia.Gia_ManObj(p, id))
	}
}

func Gia_ManTestDoms2(p *gia.Gia_Man_t) {
	start := time.Now()
	Gia_ManComputeDoms(p)
	fmt.Printf("Time: %v\n", time.Since(start))
	gia.Gia_ManCleanMark1(p)
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ObjSetMark1(gia.Gia_ManPi(p, i), 1)
	}
	nodes := vec.Vec_IntAlloc(100)
	gia.Gia_ManCreateRefs(p)
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		obj := gia.Gia_ManPi(p, i)
		id := Gia_ObjDom(p, obj)
		if id < 0 || id == gia.Gia_ObjId(p, obj) {
			continue
		}
		dom := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsCo(dom) != 0 {
			if gia.Gia_ObjFanin0(dom) != obj {
				panic("abs: output dominator has wrong fanin")
			}
			continue
		}
		Abs_GiaObjDeref_rec(p, dom)
		Abs_ManSupport2(p, dom, nodes)
		Abs_GiaObjRef_rec(p, dom)
		if vec.Vec_IntFind(nodes, gia.Gia_ObjId(p, obj)) < 0 {
			fmt.Println("FAILURE.")
		}
	}
	gia.Gia_ManCleanMark1(p)
}

func Gia_ManCollectDoms(p *gia.Gia_Man_t) *vec.Vec_Int_t {
	nodes := vec.Vec_IntAlloc(100)
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjMark1(obj) == 0 || p.Refs != nil && gia.Gia_ObjRefNum(p, obj) == 0 {
			continue
		}
		dom := Gia_ObjDom(p, obj)
		if dom < 0 || dom == id {
			continue
		}
		for dom >= 0 && gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, dom)) != 0 {
			vec.Vec_IntPush(nodes, dom)
			next := Gia_ObjDom(p, gia.Gia_ManObj(p, dom))
			if next == dom {
				break
			}
			dom = next
		}
	}
	vec.Vec_IntUniqify(nodes)
	return nodes
}
func Gia_ManComputePiDoms(p *gia.Gia_Man_t) *vec.Vec_Int_t {
	Gia_ManComputeDoms(p)
	return Gia_ManCollectDoms(p)
}
func Gia_ManTestDoms(p *gia.Gia_Man_t) {
	gia.Gia_ManCleanMark1(p)
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ObjSetMark1(gia.Gia_ManPi(p, i), 1)
	}
	Gia_ManComputePiDoms(p)
	gia.Gia_ManCleanMark1(p)
}
func Gia_ManCountFanoutlessFlops(p *gia.Gia_Man_t) {
	gia.Gia_ManCreateRefs(p)
	count := 0
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		if gia.Gia_ObjRefNum(p, gia.Gia_ManRo(p, i)) == 0 {
			count++
		}
	}
	fmt.Printf("Fanoutless flops = %d.\n", count)
	p.Refs = nil
}
func Gia_ManCountPisNodes_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, pis, ands *vec.Vec_Int_t) {
	if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	if gia.Gia_ObjMark1(obj) != 0 {
		vec.Vec_IntPush(pis, gia.Gia_ObjId(p, obj))
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: unmarked non-AND in cone")
	}
	Gia_ManCountPisNodes_rec(p, gia.Gia_ObjFanin0(obj), pis, ands)
	Gia_ManCountPisNodes_rec(p, gia.Gia_ObjFanin1(obj), pis, ands)
	vec.Vec_IntPush(ands, gia.Gia_ObjId(p, obj))
}
func Gia_ManCountPisNodes(p *gia.Gia_Man_t, pis, ands *vec.Vec_Int_t) {
	gia.Gia_ManIncrementTravId(p)
	gia.Gia_ObjSetTravIdCurrent(p, gia.Gia_ManConst0(p))
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		gia.Gia_ObjSetTravIdCurrent(p, gia.Gia_ManRo(p, i))
	}
	vec.Vec_IntClear(pis)
	vec.Vec_IntClear(ands)
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		Gia_ManCountPisNodes_rec(p, gia.Gia_ObjFanin0(gia.Gia_ManCo(p, i)), pis, ands)
	}
}

func Abs_GiaObjDeref_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t) int32 {
	if gia.Gia_ObjMark1(obj) != 0 || gia.Gia_ObjIsRo(p, obj) != 0 {
		return 0
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND while dereferencing")
	}
	count := int32(1)
	for _, fanin := range []*gia.Gia_Obj_t{gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)} {
		if gia.Gia_ObjRefNum(p, fanin) <= 0 {
			panic("abs: invalid reference count")
		}
		if gia.Gia_ObjRefDec(p, fanin) == 0 {
			count += Abs_GiaObjDeref_rec(p, fanin)
		}
	}
	return count
}
func Abs_GiaObjRef_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t) int32 {
	if gia.Gia_ObjMark1(obj) != 0 || gia.Gia_ObjIsRo(p, obj) != 0 {
		return 0
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND while restoring references")
	}
	count := int32(1)
	for _, fanin := range []*gia.Gia_Obj_t{gia.Gia_ObjFanin0(obj), gia.Gia_ObjFanin1(obj)} {
		if gia.Gia_ObjRefInc(p, fanin) == 0 {
			count += Abs_GiaObjRef_rec(p, fanin)
		}
	}
	return count
}
func Abs_GiaSortNodes(p *gia.Gia_Man_t, supp *vec.Vec_Int_t) int32 {
	ids := append([]int32(nil), vec.Vec_IntArray(supp)...)
	vec.Vec_IntClear(supp)
	for _, id := range ids {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjRefNum(p, obj) == 0 && gia.Gia_ObjIsRo(p, obj) == 0 {
			if gia.Gia_ObjMark1(obj) == 0 {
				panic("abs: removable support leaf is not marked")
			}
			vec.Vec_IntPush(supp, id)
		}
	}
	count := int32(vec.Vec_IntSize(supp))
	for _, id := range ids {
		obj := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjRefNum(p, obj) != 0 || gia.Gia_ObjIsRo(p, obj) != 0 {
			vec.Vec_IntPush(supp, id)
		}
	}
	return count
}
func Abs_ManSupport1_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, supp *vec.Vec_Int_t) {
	if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	if gia.Gia_ObjMark1(obj) != 0 || gia.Gia_ObjIsRo(p, obj) != 0 {
		vec.Vec_IntPush(supp, gia.Gia_ObjId(p, obj))
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND in support traversal")
	}
	Abs_ManSupport1_rec(p, gia.Gia_ObjFanin0(obj), supp)
	Abs_ManSupport1_rec(p, gia.Gia_ObjFanin1(obj), supp)
}
func Abs_ManSupport1(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, supp *vec.Vec_Int_t) int32 {
	vec.Vec_IntClear(supp)
	gia.Gia_ManIncrementTravId(p)
	Abs_ManSupport1_rec(p, obj, supp)
	return int32(vec.Vec_IntSize(supp))
}
func Abs_ManSupport2_rec(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, supp *vec.Vec_Int_t) {
	if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	if gia.Gia_ObjMark1(obj) != 0 || gia.Gia_ObjIsRo(p, obj) != 0 || gia.Gia_ObjRefNum(p, obj) > 0 {
		vec.Vec_IntPush(supp, gia.Gia_ObjId(p, obj))
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND in MFFC support")
	}
	Abs_ManSupport2_rec(p, gia.Gia_ObjFanin0(obj), supp)
	Abs_ManSupport2_rec(p, gia.Gia_ObjFanin1(obj), supp)
}
func Abs_ManSupport2(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, supp *vec.Vec_Int_t) int32 {
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("abs: expected AND for MFFC support")
	}
	vec.Vec_IntClear(supp)
	gia.Gia_ManIncrementTravId(p)
	Abs_ManSupport2_rec(p, gia.Gia_ObjFanin0(obj), supp)
	Abs_ManSupport2_rec(p, gia.Gia_ObjFanin1(obj), supp)
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	return int32(vec.Vec_IntSize(supp))
}
func Abs_ManSupport3(p *gia.Gia_Man_t, obj *gia.Gia_Obj_t, supp *vec.Vec_Int_t) int32 {
	Abs_ManSupport2(p, obj, supp)
	removable := int(Abs_GiaSortNodes(p, supp))
	if removable == 0 {
		panic("abs: no removable support leaves")
	}
	for {
		changed := false
		for i := removable; i < vec.Vec_IntSize(supp); i++ {
			node := gia.Gia_ManObj(p, vec.Vec_IntEntry(supp, i))
			if gia.Gia_ObjIsAnd(node) == 0 {
				continue
			}
			if gia.Gia_ObjMark1(node) != 0 || gia.Gia_ObjRefNum(p, node) <= 0 {
				panic("abs: invalid extended support")
			}
			a, b := gia.Gia_ObjFanin0(node), gia.Gia_ObjFanin1(node)
			aSeen, bSeen := gia.Gia_ObjIsTravIdCurrent(p, a) != 0, gia.Gia_ObjIsTravIdCurrent(p, b) != 0
			if !aSeen && !bSeen {
				continue
			}
			vec.Vec_IntRemove(supp, gia.Gia_ObjId(p, node))
			if !aSeen {
				vec.Vec_IntPush(supp, gia.Gia_ObjId(p, a))
				gia.Gia_ObjSetTravIdCurrent(p, a)
			}
			if !bSeen {
				vec.Vec_IntPush(supp, gia.Gia_ObjId(p, b))
				gia.Gia_ObjSetTravIdCurrent(p, b)
			}
			changed = true
			break
		}
		if !changed {
			break
		}
	}
	return int32(vec.Vec_IntSize(supp))
}

// Abs_RpmPerformMark finds parameterizable dominators and marks them as the new
// inputs. It retains references until every candidate in an iteration is tested.
func Abs_RpmPerformMark(p *gia.Gia_Man_t, cutMax, verbose, veryVerbose int32) {
	gia.Gia_ManCreateRefs(p)
	gia.Gia_ManCleanMark1(p)
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ObjSetMark1(gia.Gia_ManPi(p, i), 1)
	}
	pis, ands, support := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	iteration := 0
	for changed := true; changed; iteration++ {
		changed = false
		doms := Gia_ManComputePiDoms(p)
		if verbose != 0 || veryVerbose != 0 {
			Gia_ManCountPisNodes(p, pis, ands)
			fmt.Printf("Iter %3d :  PI = %5d  And = %6d  Dom = %5d\n", iteration, vec.Vec_IntSize(pis), vec.Vec_IntSize(ands), vec.Vec_IntSize(doms))
		}
		for _, id := range vec.Vec_IntArray(doms) {
			obj := gia.Gia_ManObj(p, id)
			if gia.Gia_ObjMark1(obj) != 0 || gia.Gia_ObjRefNum(p, obj) <= 0 {
				panic("abs: invalid reparameterization candidate")
			}
			nodes := Abs_GiaObjDeref_rec(p, obj)
			if Abs_ManSupport2(p, obj, support) > cutMax {
				Abs_GiaObjRef_rec(p, obj)
				continue
			}
			removable := Abs_GiaSortNodes(p, support)
			if removable <= 0 || removable > cutMax {
				panic("abs: invalid removable support")
			}
			truth := gia.Gia_ObjComputeTruthTableCut(p, obj, support)
			if truth == nil {
				Abs_GiaObjRef_rec(p, obj)
				continue
			}
			hasConst := Abs_GiaCheckTruth(truth, int32(vec.Vec_IntSize(support)), removable) == 0
			if veryVerbose != 0 {
				fmt.Printf("Nodes =%3d Size =%3d Size0 =%3d %v", nodes, vec.Vec_IntSize(support), removable, hasConst)
				Abs_GiaCofPrint(truth, int32(vec.Vec_IntSize(support)), removable, boolToInt32(hasConst))
			}
			if hasConst {
				Abs_GiaObjRef_rec(p, obj)
				continue
			}
			gia.Gia_ObjSetMark1(obj, 1)
			changed = true
		}
	}
	if veryVerbose != 0 {
		Gia_ManCountPisNodes(p, pis, ands)
		fmt.Printf("Iter %3d :  PI = %5d  And = %6d\n", iteration, vec.Vec_IntSize(pis), vec.Vec_IntSize(ands))
	}
	p.Refs = nil
}
func boolToInt32(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

// Gia_ManDupRpm duplicates the graph with marked internal inputs exposed as PIs.
func Gia_ManDupRpm(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	pis, ands := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	Gia_ManCountPisNodes(p, pis, ands)
	gia.Gia_ManFillValue(p)
	q := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	gia.Gia_ManCopyNameSpec(q, p)
	gia.Gia_ManConst0(p).Value = 0
	for _, id := range vec.Vec_IntArray(pis) {
		gia.Gia_ManObj(p, id).Value = uint32(gia.Gia_ManAppendCi(q))
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		gia.Gia_ManRo(p, i).Value = uint32(gia.Gia_ManAppendCi(q))
	}
	for _, id := range vec.Vec_IntArray(ands) {
		obj := gia.Gia_ManObj(p, id)
		obj.Value = uint32(gia.Gia_ManAppendAnd(q, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj)))
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		gia.Gia_ManAppendCo(q, gia.Gia_ObjFanin0Copy(gia.Gia_ManCo(p, i)))
	}
	gia.Gia_ManSetRegNum(q, gia.Gia_ManRegNum(p))
	return q
}
func Abs_RpmPerform(p *gia.Gia_Man_t, cutMax, verbose, veryVerbose int32) *gia.Gia_Man_t {
	gia.Gia_ObjComputeTruthTableStart(p, cutMax)
	Abs_RpmPerformMark(p, cutMax, verbose, veryVerbose)
	gia.Gia_ObjComputeTruthTableStop(p)
	result := Gia_ManDupRpm(p)
	gia.Gia_ManCleanMark1(p)
	return result
}
