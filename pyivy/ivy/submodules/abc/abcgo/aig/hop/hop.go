// Ported from ABC hop.h, Alan Mishchenko, 2006.
// Package hop implements structurally hashed And-Inverter Graphs.
package hop

type Hop_Edge_t = int32
type Hop_Type_t int32

const (
	AIG_NONE Hop_Type_t = iota
	AIG_CONST1
	AIG_PI
	AIG_PO
	AIG_AND
	AIG_EXOR
	AIG_VOID
)

type Hop_Obj_t = Hop_Obj_t_
type Hop_Man_t = Hop_Man_t_

// A complemented edge is a stable proxy for its regular node. No address bits
// are modified, so pointers remain visible to Go's garbage collector.
type Hop_Obj_t_ struct {
	pData                             any
	iData                             int32
	pNext                             *Hop_Obj_t
	pNextData                         any
	PioNum                            int32
	pFanin0, pFanin1                  *Hop_Obj_t
	Type                              Hop_Type_t
	fPhase, fMarkA, fMarkB, nRefs, Id int32
	regular, neg                      *Hop_Obj_t
	travId                            int32
}
type Hop_Man_t_ struct {
	vPis, vPos, vObjs               *Vec_Ptr_t
	pConst1                         *Hop_Obj_t
	Ghost                           Hop_Obj_t
	nObjs                           [AIG_VOID]int32
	nCreated, nDeleted              int32
	pTable                          []*Hop_Obj_t
	nTableSize                      int32
	pData                           any
	nTravIds, fRefCount, fCatchExor int32
	nodePages                       [][]Hop_Obj_t
	pListFree                       *Hop_Obj_t
	time1, time2                    int64
}

func hopBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Hop_BitWordNum(n int32) int32 { return (n + 31) >> 5 }
func Hop_TruthWordNum(n int32) int32 {
	if n <= 5 {
		return 1
	}
	return 1 << (n - 5)
}
func Hop_InfoHasBit(p []uint32, i int32) int32 { return hopBool(p[i>>5]&(1<<uint(i&31)) != 0) }
func Hop_InfoSetBit(p []uint32, i int32)       { p[i>>5] |= 1 << uint(i&31) }
func Hop_InfoXorBit(p []uint32, i int32)       { p[i>>5] ^= 1 << uint(i&31) }
func Hop_Base2Log(n uint32) int32 {
	if n < 2 {
		return int32(n)
	}
	var r int32
	for n--; n != 0; n >>= 1 {
		r++
	}
	return r
}
func Hop_Base10Log(n uint32) int32 {
	if n < 2 {
		return int32(n)
	}
	var r int32
	for n--; n != 0; n /= 10 {
		r++
	}
	return r
}
func Hop_Regular(p *Hop_Obj_t) *Hop_Obj_t {
	if p != nil && p.regular != nil {
		return p.regular
	}
	return p
}
func Hop_Not(p *Hop_Obj_t) *Hop_Obj_t {
	if p == nil {
		return nil
	}
	if p.regular != nil {
		return p.regular
	}
	if p.neg == nil {
		p.neg = &Hop_Obj_t{regular: p}
	}
	return p.neg
}
func Hop_NotCond(p *Hop_Obj_t, c int32) *Hop_Obj_t {
	if c != 0 {
		return Hop_Not(p)
	}
	return p
}
func Hop_IsComplement(p *Hop_Obj_t) int32        { return hopBool(p != nil && p.regular != nil) }
func Hop_ManConst0(p *Hop_Man_t) *Hop_Obj_t      { return Hop_Not(p.pConst1) }
func Hop_ManConst1(p *Hop_Man_t) *Hop_Obj_t      { return p.pConst1 }
func Hop_ManGhost(p *Hop_Man_t) *Hop_Obj_t       { return &p.Ghost }
func Hop_ManPi(p *Hop_Man_t, i int32) *Hop_Obj_t { return Vec_PtrEntry(p.vPis, i).(*Hop_Obj_t) }
func Hop_ManPo(p *Hop_Man_t, i int32) *Hop_Obj_t { return Vec_PtrEntry(p.vPos, i).(*Hop_Obj_t) }
func Hop_ManObj(p *Hop_Man_t, i int32) *Hop_Obj_t {
	if p.vObjs == nil {
		return nil
	}
	return Vec_PtrEntry(p.vObjs, i).(*Hop_Obj_t)
}
func Hop_EdgeCreate(id, c int32) Hop_Edge_t            { return id<<1 | c }
func Hop_EdgeId(e Hop_Edge_t) int32                    { return e >> 1 }
func Hop_EdgeIsComplement(e Hop_Edge_t) int32          { return e & 1 }
func Hop_EdgeRegular(e Hop_Edge_t) Hop_Edge_t          { return e &^ 1 }
func Hop_EdgeNot(e Hop_Edge_t) Hop_Edge_t              { return e ^ 1 }
func Hop_EdgeNotCond(e Hop_Edge_t, c int32) Hop_Edge_t { return e ^ c }
func Hop_ManPiNum(p *Hop_Man_t) int32                  { return p.nObjs[AIG_PI] }
func Hop_ManPoNum(p *Hop_Man_t) int32                  { return p.nObjs[AIG_PO] }
func Hop_ManAndNum(p *Hop_Man_t) int32                 { return p.nObjs[AIG_AND] }
func Hop_ManExorNum(p *Hop_Man_t) int32                { return p.nObjs[AIG_EXOR] }
func Hop_ManNodeNum(p *Hop_Man_t) int32                { return p.nObjs[AIG_AND] + p.nObjs[AIG_EXOR] }
func Hop_ManObjNum(p *Hop_Man_t) int32                 { return p.nCreated - p.nDeleted }
func Hop_ManObjects(p *Hop_Man_t) []*Hop_Obj_t {
	if p == nil || p.vObjs == nil {
		return nil
	}
	objects := make([]*Hop_Obj_t, 0, Vec_PtrSize(p.vObjs))
	for _, item := range Vec_PtrArray(p.vObjs) {
		if o, ok := item.(*Hop_Obj_t); ok && o.Type != AIG_NONE {
			objects = append(objects, o)
		}
	}
	return objects
}
func Hop_ManStartWithObjects(capacity int) *Hop_Man_t {
	p := Hop_ManStart()
	p.vObjs = Vec_PtrAlloc(int32(capacity))
	Vec_PtrPush(p.vObjs, p.pConst1)
	return p
}
func Hop_ObjReadNext(o *Hop_Obj_t) *Hop_Obj_t   { return Hop_Regular(o).pNext }
func Hop_ObjSetNext(o, next *Hop_Obj_t)         { Hop_Regular(o).pNext = next }
func Hop_ObjReadNextData(o *Hop_Obj_t) any      { return Hop_Regular(o).pNextData }
func Hop_ObjSetNextData(o *Hop_Obj_t, next any) { Hop_Regular(o).pNextData = next }
func Hop_ManGetCost(p *Hop_Man_t) int32         { return Hop_ManAndNum(p) + 3*Hop_ManExorNum(p) }
func Hop_ObjType(p *Hop_Obj_t) Hop_Type_t       { return p.Type }
func Hop_ObjIsNone(p *Hop_Obj_t) int32          { return hopBool(p.Type == AIG_NONE) }
func Hop_ObjIsConst1(p *Hop_Obj_t) int32        { return hopBool(p.Type == AIG_CONST1) }
func Hop_ObjIsPi(p *Hop_Obj_t) int32            { return hopBool(p.Type == AIG_PI) }
func Hop_ObjIsPo(p *Hop_Obj_t) int32            { return hopBool(p.Type == AIG_PO) }
func Hop_ObjIsAnd(p *Hop_Obj_t) int32           { return hopBool(p.Type == AIG_AND) }
func Hop_ObjIsExor(p *Hop_Obj_t) int32          { return hopBool(p.Type == AIG_EXOR) }
func Hop_ObjIsNode(p *Hop_Obj_t) int32          { return hopBool(p.Type == AIG_AND || p.Type == AIG_EXOR) }
func Hop_ObjIsTerm(p *Hop_Obj_t) int32 {
	return hopBool(p.Type == AIG_PI || p.Type == AIG_PO || p.Type == AIG_CONST1)
}
func Hop_ObjIsHash(p *Hop_Obj_t) int32                        { return hopBool(p.Type == AIG_AND || p.Type == AIG_EXOR) }
func Hop_ObjIsMarkA(p *Hop_Obj_t) int32                       { return hopBool(p.fMarkA != 0) }
func Hop_ObjSetMarkA(p *Hop_Obj_t)                            { p.fMarkA = 1 }
func Hop_ObjClearMarkA(p *Hop_Obj_t)                          { p.fMarkA = 0 }
func Hop_ObjSetTravId(p *Hop_Obj_t, id int32)                 { p.travId = id }
func Hop_ObjSetTravIdCurrent(p *Hop_Man_t, o *Hop_Obj_t)      { o.travId = p.nTravIds }
func Hop_ObjSetTravIdPrevious(p *Hop_Man_t, o *Hop_Obj_t)     { o.travId = p.nTravIds - 1 }
func Hop_ObjIsTravIdCurrent(p *Hop_Man_t, o *Hop_Obj_t) int32 { return hopBool(o.travId == p.nTravIds) }
func Hop_ObjIsTravIdPrevious(p *Hop_Man_t, o *Hop_Obj_t) int32 {
	return hopBool(o.travId == p.nTravIds-1)
}
func Hop_ObjTravId(o *Hop_Obj_t) int32      { return o.travId }
func Hop_ObjPhase(o *Hop_Obj_t) int32       { return o.fPhase }
func Hop_ObjRefs(o *Hop_Obj_t) int32        { return o.nRefs }
func Hop_ObjRef(o *Hop_Obj_t)               { o.nRefs++ }
func Hop_ObjDeref(o *Hop_Obj_t)             { assert(o.nRefs > 0, "positive reference count"); o.nRefs-- }
func Hop_ObjClearRef(o *Hop_Obj_t)          { o.nRefs = 0 }
func Hop_ObjFaninC0(o *Hop_Obj_t) int32     { return Hop_IsComplement(o.pFanin0) }
func Hop_ObjFaninC1(o *Hop_Obj_t) int32     { return Hop_IsComplement(o.pFanin1) }
func Hop_ObjFanin0(o *Hop_Obj_t) *Hop_Obj_t { return Hop_Regular(o.pFanin0) }
func Hop_ObjFanin1(o *Hop_Obj_t) *Hop_Obj_t { return Hop_Regular(o.pFanin1) }
func Hop_ObjChild0(o *Hop_Obj_t) *Hop_Obj_t { return o.pFanin0 }
func Hop_ObjChild1(o *Hop_Obj_t) *Hop_Obj_t { return o.pFanin1 }
func Hop_ObjChild0Copy(o *Hop_Obj_t) *Hop_Obj_t {
	f := Hop_ObjFanin0(o)
	if f == nil {
		return nil
	}
	v, _ := f.pData.(*Hop_Obj_t)
	return Hop_NotCond(v, Hop_ObjFaninC0(o))
}
func Hop_ObjChild1Copy(o *Hop_Obj_t) *Hop_Obj_t {
	f := Hop_ObjFanin1(o)
	if f == nil {
		return nil
	}
	v, _ := f.pData.(*Hop_Obj_t)
	return Hop_NotCond(v, Hop_ObjFaninC1(o))
}
func Hop_ObjChild0CopyI(o *Hop_Obj_t) int32 {
	f := Hop_ObjFanin0(o)
	if f == nil {
		return -1
	}
	return f.iData ^ Hop_ObjFaninC0(o)
}
func Hop_ObjChild1CopyI(o *Hop_Obj_t) int32 {
	f := Hop_ObjFanin1(o)
	if f == nil {
		return -1
	}
	return f.iData ^ Hop_ObjFaninC1(o)
}
func Hop_ObjLevel(o *Hop_Obj_t) int32 { return o.nRefs }
func Hop_ObjLevelNew(o *Hop_Obj_t) int32 {
	a, b := Hop_ObjFanin0(o).nRefs, Hop_ObjFanin1(o).nRefs
	if b > a {
		a = b
	}
	return 1 + Hop_ObjIsExor(o) + a
}
func Hop_ObjPhaseCompl(o *Hop_Obj_t) int32 { return Hop_Regular(o).fPhase ^ Hop_IsComplement(o) }
func Hop_ObjClean(o *Hop_Obj_t)            { *o = Hop_Obj_t{} }
func Hop_ObjWhatFanin(o, f *Hop_Obj_t) int32 {
	if Hop_ObjFanin0(o) == f {
		return 0
	}
	if Hop_ObjFanin1(o) == f {
		return 1
	}
	panic("object is not a fanin")
}
func Hop_ObjFanoutC(o, f *Hop_Obj_t) int32 {
	if Hop_ObjFanin0(f) == o {
		return Hop_ObjFaninC0(f)
	}
	if Hop_ObjFanin1(f) == o {
		return Hop_ObjFaninC1(f)
	}
	panic("object is not a fanin")
}
func Hop_ObjCreateGhost(p *Hop_Man_t, a, b *Hop_Obj_t, t Hop_Type_t) *Hop_Obj_t {
	if b != nil && Hop_Regular(a).Id > Hop_Regular(b).Id {
		a, b = b, a
	}
	p.Ghost = Hop_Obj_t{Type: t, pFanin0: a, pFanin1: b}
	return &p.Ghost
}
func Hop_ManFetchMemory(p *Hop_Man_t) *Hop_Obj_t {
	if p.pListFree == nil {
		Hop_ManAddMemory(p)
	}
	o := p.pListFree
	p.pListFree = o.pNext
	*o = Hop_Obj_t{Id: p.nCreated}
	p.nCreated++
	if p.vObjs != nil {
		Vec_PtrPush(p.vObjs, o)
	}
	return o
}
func Hop_ManRecycleMemory(p *Hop_Man_t, o *Hop_Obj_t) {
	o.Type = AIG_NONE
	o.pNext = p.pListFree
	p.pListFree = o
}
