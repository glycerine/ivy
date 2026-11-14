// Package ivy implements ABC's sequential And-Inverter Graph representation.
// Ported from src/aig/ivy/ivy.h (Alan Mishchenko, UC Berkeley).
package ivy

// Integer edges retain the low complement bit without signed-shift overflow.
type Ivy_Edge_t uint32
type Ivy_Type_t int32

const (
	IVY_NONE Ivy_Type_t = iota
	IVY_PI
	IVY_PO
	IVY_ASSERT
	IVY_LATCH
	IVY_AND
	IVY_EXOR
	IVY_BUF
	IVY_VOID
)

type Ivy_Init_t int32

const (
	IVY_INIT_NONE Ivy_Init_t = iota
	IVY_INIT_0
	IVY_INIT_1
	IVY_INIT_DC
)
const (
	IVY_CUT_LIMIT = 256
	IVY_CUT_INPUT = 6
	IVY_LEAF_MASK = 255
	IVY_LEAF_BITS = 8
)

type Ivy_Obj_t = Ivy_Obj_t_
type Ivy_Man_t = Ivy_Man_t_
type Ivy_FraigParams_t = Ivy_FraigParams_t_

// Regular objects own state. A complemented pointer is a cached proxy whose
// regular field points at the object; no pointer bits are changed or concealed
// from Go's garbage collector. A manager is not safe for concurrent mutation.
type Ivy_Obj_t_ struct {
	fraig                                               *ivyFraigNodeState
	Id, TravId                                          int32
	Type                                                Ivy_Type_t
	fMarkA, fMarkB, fExFan, fPhase, fFailTfo            int32
	Init                                                Ivy_Init_t
	Level, nRefs                                        int32
	pFanin0, pFanin1                                    *Ivy_Obj_t
	pFanout, pNextFan0, pNextFan1, pPrevFan0, pPrevFan1 *Ivy_Obj_t
	pEquiv                                              *Ivy_Obj_t
	regular, neg                                        *Ivy_Obj_t
	owner                                               *Ivy_Man_t
	recycled                                            bool
}
type Ivy_Man_t_ struct {
	vPis, vPos, vBufs, vObjs        []*Ivy_Obj_t
	pConst1                         *Ivy_Obj_t
	Ghost                           Ivy_Obj_t
	nObjs                           [IVY_VOID]int32
	nCreated, nDeleted              int32
	pTable                          map[ivyHashKey]*Ivy_Obj_t
	fCatchExor, nTravIds, nLevelMax int32
	vRequired                       []int32
	fFanout                         int32
	pData, pCopy                    any
	pHaig                           *Ivy_Man_t
	nClassesSkip                    int32
	pages                           [][]Ivy_Obj_t
	freeNodes                       []*Ivy_Obj_t
	memoryStopped                   bool
	time1, time2                    int64
}
type Ivy_FraigParams_t_ struct {
	nSimWords                                                int32
	dSimSatur                                                float64
	fPatScores, MaxScore                                     int32
	dActConeRatio, dActConeBumpMax                           float64
	fProve, fVerbose, fDoSparse, nBTLimitNode, nBTLimitMiter int32
}

// Ivy_ManTakeData transfers the source C pData payload to the caller.
func Ivy_ManTakeData(manager *Ivy_Man_t) any {
	if manager == nil {
		return nil
	}
	data := manager.pData
	manager.pData = nil
	return data
}

// Ivy_ManReadHaig returns the sequential equivalence graph created by
// Ivy_ManHaigStart, matching the source pHaig field read by abcIvy.c.
func Ivy_ManReadHaig(manager *Ivy_Man_t) *Ivy_Man_t {
	if manager == nil {
		return nil
	}
	return manager.pHaig
}

type Ivy_Cut_t struct {
	nLatches        int32
	nSize, nSizeMax int16
	pArray          [IVY_CUT_INPUT]int32
	uHash           uint32
}
type Ivy_Store_t struct {
	nCuts, nCutsM, nCutsMax, fSatur int32
	pCuts                           [IVY_CUT_LIMIT]Ivy_Cut_t
}

func ivyBool(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func Ivy_BitWordNum(n int32) int32 {
	if n < 0 {
		panic("negative bit count")
	}
	return n/32 + ivyBool(n%32 != 0)
}
func Ivy_TruthWordNum(n int32) int32 {
	if n < 0 || n > 35 {
		panic("invalid truth support")
	}
	if n <= 5 {
		return 1
	}
	return 1 << uint(n-5)
}
func Ivy_InfoHasBit(p []uint32, i int32) int32 {
	if i < 0 {
		panic("negative bit index")
	}
	return int32(p[i>>5] >> uint(i&31) & 1)
}
func Ivy_InfoSetBit(p []uint32, i int32) {
	if i < 0 {
		panic("negative bit index")
	}
	p[i>>5] |= 1 << uint(i&31)
}
func Ivy_InfoXorBit(p []uint32, i int32) {
	if i < 0 {
		panic("negative bit index")
	}
	p[i>>5] ^= 1 << uint(i&31)
}
func Ivy_Regular(p *Ivy_Obj_t) *Ivy_Obj_t {
	if p != nil && p.regular != nil {
		return p.regular
	}
	return p
}
func Ivy_Not(p *Ivy_Obj_t) *Ivy_Obj_t {
	if p == nil {
		return nil
	}
	if p.regular != nil {
		return p.regular
	}
	if p.neg == nil {
		p.neg = &Ivy_Obj_t{regular: p}
	}
	return p.neg
}
func Ivy_NotCond(p *Ivy_Obj_t, c int32) *Ivy_Obj_t {
	if c != 0 {
		return Ivy_Not(p)
	}
	return p
}
func Ivy_IsComplement(p *Ivy_Obj_t) int32         { return ivyBool(p != nil && p.regular != nil) }
func Ivy_ManConst0(p *Ivy_Man_t) *Ivy_Obj_t       { return Ivy_Not(p.pConst1) }
func Ivy_ManConst1(p *Ivy_Man_t) *Ivy_Obj_t       { return p.pConst1 }
func Ivy_ManGhost(p *Ivy_Man_t) *Ivy_Obj_t        { return &p.Ghost }
func Ivy_ManPi(p *Ivy_Man_t, i int32) *Ivy_Obj_t  { return p.vPis[i] }
func Ivy_ManPo(p *Ivy_Man_t, i int32) *Ivy_Obj_t  { return p.vPos[i] }
func Ivy_ManObj(p *Ivy_Man_t, i int32) *Ivy_Obj_t { return p.vObjs[i] }
func Ivy_EdgeCreate(id, c int32) Ivy_Edge_t {
	if id < 0 {
		panic("negative edge ID")
	}
	return Ivy_Edge_t(uint32(id)<<1 | uint32(ivyBool(c != 0)))
}
func Ivy_EdgeId(e Ivy_Edge_t) int32           { return int32(e >> 1) }
func Ivy_EdgeIsComplement(e Ivy_Edge_t) int32 { return int32(e & 1) }
func Ivy_EdgeRegular(e Ivy_Edge_t) Ivy_Edge_t { return e &^ 1 }
func Ivy_EdgeNot(e Ivy_Edge_t) Ivy_Edge_t     { return e ^ 1 }
func Ivy_EdgeNotCond(e Ivy_Edge_t, c int32) Ivy_Edge_t {
	if c != 0 {
		return e ^ 1
	}
	return e
}
func Ivy_EdgeFromNode(node *Ivy_Obj_t) Ivy_Edge_t {
	if node == nil {
		panic("nil edge node")
	}
	return Ivy_EdgeCreate(Ivy_Regular(node).Id, Ivy_IsComplement(node))
}
func Ivy_EdgeToNode(p *Ivy_Man_t, e Ivy_Edge_t) *Ivy_Obj_t {
	return Ivy_NotCond(Ivy_ManObj(p, Ivy_EdgeId(e)), Ivy_EdgeIsComplement(e))
}
func Ivy_LeafCreate(id, lat int32) int32 {
	if id < 0 || id > 0xffffff || lat < 0 || lat > 255 {
		panic("invalid sequential leaf")
	}
	return int32(uint32(id)<<8 | uint32(lat))
}
func Ivy_LeafId(leaf int32) int32          { return int32(uint32(leaf) >> 8) }
func Ivy_LeafLat(leaf int32) int32         { return leaf & 255 }
func Ivy_ManPiNum(p *Ivy_Man_t) int32      { return p.nObjs[IVY_PI] }
func Ivy_ManPoNum(p *Ivy_Man_t) int32      { return p.nObjs[IVY_PO] }
func Ivy_ManAssertNum(p *Ivy_Man_t) int32  { return p.nObjs[IVY_ASSERT] }
func Ivy_ManLatchNum(p *Ivy_Man_t) int32   { return p.nObjs[IVY_LATCH] }
func Ivy_ManAndNum(p *Ivy_Man_t) int32     { return p.nObjs[IVY_AND] }
func Ivy_ManExorNum(p *Ivy_Man_t) int32    { return p.nObjs[IVY_EXOR] }
func Ivy_ManBufNum(p *Ivy_Man_t) int32     { return p.nObjs[IVY_BUF] }
func Ivy_ManObjNum(p *Ivy_Man_t) int32     { return p.nCreated - p.nDeleted }
func Ivy_ManObjIdMax(p *Ivy_Man_t) int32   { return int32(len(p.vObjs)) - 1 }
func Ivy_ManNodeNum(p *Ivy_Man_t) int32    { return p.nObjs[IVY_AND] + p.nObjs[IVY_EXOR] }
func Ivy_ManHashObjNum(p *Ivy_Man_t) int32 { return Ivy_ManNodeNum(p) + p.nObjs[IVY_LATCH] }
func Ivy_ManGetCost(p *Ivy_Man_t) int32 {
	return p.nObjs[IVY_AND] + 3*p.nObjs[IVY_EXOR] + 8*p.nObjs[IVY_LATCH]
}
func Ivy_ObjType(o *Ivy_Obj_t) Ivy_Type_t { return Ivy_Regular(o).Type }
func Ivy_ObjInit(o *Ivy_Obj_t) Ivy_Init_t { return Ivy_Regular(o).Init }
func Ivy_ObjIsConst1(o *Ivy_Obj_t) int32  { return ivyBool(Ivy_Regular(o).Id == 0) }
func Ivy_ObjIsGhost(o *Ivy_Obj_t) int32   { return ivyBool(Ivy_Regular(o).Id < 0) }
func Ivy_ObjIsNone(o *Ivy_Obj_t) int32    { return ivyBool(Ivy_ObjType(o) == IVY_NONE) }
func Ivy_ObjIsPi(o *Ivy_Obj_t) int32      { t := Ivy_ObjType(o); return ivyBool(t == IVY_PI) }
func Ivy_ObjIsPo(o *Ivy_Obj_t) int32      { t := Ivy_ObjType(o); return ivyBool(t == IVY_PO) }
func Ivy_ObjIsCi(o *Ivy_Obj_t) int32 {
	t := Ivy_ObjType(o)
	return ivyBool(t == IVY_PI || t == IVY_LATCH)
}
func Ivy_ObjIsCo(o *Ivy_Obj_t) int32 {
	t := Ivy_ObjType(o)
	return ivyBool(t == IVY_PO || t == IVY_LATCH)
}
func Ivy_ObjIsAssert(o *Ivy_Obj_t) int32 { t := Ivy_ObjType(o); return ivyBool(t == IVY_ASSERT) }
func Ivy_ObjIsLatch(o *Ivy_Obj_t) int32  { t := Ivy_ObjType(o); return ivyBool(t == IVY_LATCH) }
func Ivy_ObjIsAnd(o *Ivy_Obj_t) int32    { t := Ivy_ObjType(o); return ivyBool(t == IVY_AND) }
func Ivy_ObjIsExor(o *Ivy_Obj_t) int32   { t := Ivy_ObjType(o); return ivyBool(t == IVY_EXOR) }
func Ivy_ObjIsBuf(o *Ivy_Obj_t) int32    { t := Ivy_ObjType(o); return ivyBool(t == IVY_BUF) }
func Ivy_ObjIsNode(o *Ivy_Obj_t) int32 {
	t := Ivy_ObjType(o)
	return ivyBool(t == IVY_AND || t == IVY_EXOR)
}
func Ivy_ObjIsTerm(o *Ivy_Obj_t) int32 {
	t := Ivy_ObjType(o)
	return ivyBool(t == IVY_PI || t == IVY_PO || t == IVY_ASSERT)
}
func Ivy_ObjIsHash(o *Ivy_Obj_t) int32 {
	t := Ivy_ObjType(o)
	return ivyBool(t == IVY_AND || t == IVY_EXOR || t == IVY_LATCH)
}
func Ivy_ObjIsOneFanin(o *Ivy_Obj_t) int32 {
	t := Ivy_ObjType(o)
	return ivyBool(t == IVY_PO || t == IVY_ASSERT || t == IVY_BUF || t == IVY_LATCH)
}
func Ivy_ObjIsMarkA(o *Ivy_Obj_t) int32                   { return Ivy_Regular(o).fMarkA }
func Ivy_ObjSetMarkA(o *Ivy_Obj_t)                        { Ivy_Regular(o).fMarkA = 1 }
func Ivy_ObjClearMarkA(o *Ivy_Obj_t)                      { Ivy_Regular(o).fMarkA = 0 }
func Ivy_ObjSetTravId(o *Ivy_Obj_t, id int32)             { Ivy_Regular(o).TravId = id }
func Ivy_ObjSetTravIdCurrent(p *Ivy_Man_t, o *Ivy_Obj_t)  { Ivy_ObjSetTravId(o, p.nTravIds) }
func Ivy_ObjSetTravIdPrevious(p *Ivy_Man_t, o *Ivy_Obj_t) { Ivy_ObjSetTravId(o, p.nTravIds-1) }
func Ivy_ObjIsTravIdCurrent(p *Ivy_Man_t, o *Ivy_Obj_t) int32 {
	return ivyBool(Ivy_Regular(o).TravId == p.nTravIds)
}
func Ivy_ObjIsTravIdPrevious(p *Ivy_Man_t, o *Ivy_Obj_t) int32 {
	return ivyBool(Ivy_Regular(o).TravId == p.nTravIds-1)
}
func Ivy_ObjId(o *Ivy_Obj_t) int32         { return Ivy_Regular(o).Id }
func Ivy_ObjTravId(o *Ivy_Obj_t) int32     { return Ivy_Regular(o).TravId }
func Ivy_ObjPhase(o *Ivy_Obj_t) int32      { return Ivy_Regular(o).fPhase }
func Ivy_ObjExorFanout(o *Ivy_Obj_t) int32 { return Ivy_Regular(o).fExFan }
func Ivy_ObjRefs(o *Ivy_Obj_t) int32       { return Ivy_Regular(o).nRefs }
func Ivy_ObjRefsInc(o *Ivy_Obj_t)          { Ivy_Regular(o).nRefs++ }
func Ivy_ObjRefsDec(o *Ivy_Obj_t) {
	o = Ivy_Regular(o)
	if o.nRefs <= 0 {
		panic("reference count underflow")
	}
	o.nRefs--
}
func Ivy_ObjFaninId0(o *Ivy_Obj_t) int32 {
	f := Ivy_ObjFanin0(o)
	if f == nil {
		return 0
	}
	return f.Id
}
func Ivy_ObjFaninId1(o *Ivy_Obj_t) int32 {
	f := Ivy_ObjFanin1(o)
	if f == nil {
		return 0
	}
	return f.Id
}
func Ivy_ObjFaninC0(o *Ivy_Obj_t) int32     { return Ivy_IsComplement(Ivy_Regular(o).pFanin0) }
func Ivy_ObjFaninC1(o *Ivy_Obj_t) int32     { return Ivy_IsComplement(Ivy_Regular(o).pFanin1) }
func Ivy_ObjFanin0(o *Ivy_Obj_t) *Ivy_Obj_t { return Ivy_Regular(Ivy_Regular(o).pFanin0) }
func Ivy_ObjFanin1(o *Ivy_Obj_t) *Ivy_Obj_t { return Ivy_Regular(Ivy_Regular(o).pFanin1) }
func Ivy_ObjChild0(o *Ivy_Obj_t) *Ivy_Obj_t { return Ivy_Regular(o).pFanin0 }
func Ivy_ObjChild1(o *Ivy_Obj_t) *Ivy_Obj_t { return Ivy_Regular(o).pFanin1 }
func Ivy_ObjChild0Equiv(o *Ivy_Obj_t) *Ivy_Obj_t {
	f := Ivy_ObjFanin0(o)
	if f == nil {
		return nil
	}
	return Ivy_NotCond(f.pEquiv, Ivy_ObjFaninC0(o))
}
func Ivy_ObjChild1Equiv(o *Ivy_Obj_t) *Ivy_Obj_t {
	f := Ivy_ObjFanin1(o)
	if f == nil {
		return nil
	}
	return Ivy_NotCond(f.pEquiv, Ivy_ObjFaninC1(o))
}
func Ivy_ObjEquiv(o *Ivy_Obj_t) *Ivy_Obj_t {
	if o == nil {
		return nil
	}
	regular := Ivy_Regular(o)
	if regular.pEquiv == nil {
		return nil
	}
	return Ivy_NotCond(regular.pEquiv, Ivy_IsComplement(o))
}
func Ivy_ObjLevel(o *Ivy_Obj_t) int32 { return Ivy_Regular(o).Level }
func Ivy_ObjLevelNew(o *Ivy_Obj_t) int32 {
	level := 1 + Ivy_ObjIsExor(o) + max(Ivy_ObjLevel(Ivy_ObjFanin0(o)), Ivy_ObjLevel(Ivy_ObjFanin1(o)))
	if level > 0x1fffff {
		panic("Ivy level exceeds representation bound")
	}
	return level
}
func Ivy_ObjFaninPhase(o *Ivy_Obj_t) int32 {
	if o == nil {
		return 0
	}
	return Ivy_ObjPhase(o) ^ Ivy_IsComplement(o)
}
func Ivy_ObjClean(o *Ivy_Obj_t) {
	o = Ivy_Regular(o)
	id, owner, neg, recycled := o.Id, o.owner, o.neg, o.recycled
	*o = Ivy_Obj_t{Id: id, owner: owner, neg: neg, recycled: recycled}
}
func Ivy_ObjOverwrite(base, data *Ivy_Obj_t) {
	base, data = Ivy_Regular(base), Ivy_Regular(data)
	if base.recycled {
		panic("cannot overwrite recycled node")
	}
	id, owner, neg := base.Id, base.owner, base.neg
	*base = *data
	base.Id, base.owner, base.neg, base.regular = id, owner, neg, nil
	base.recycled = false
}
func Ivy_ObjWhatFanin(o, f *Ivy_Obj_t) int32 {
	f = Ivy_Regular(f)
	if Ivy_ObjFanin0(o) == f {
		return 0
	}
	if Ivy_ObjFanin1(o) == f {
		return 1
	}
	panic("object does not have this fanin")
}
func Ivy_ObjFanoutC(o, fanout *Ivy_Obj_t) int32 {
	o = Ivy_Regular(o)
	if Ivy_ObjFanin0(fanout) == o {
		return Ivy_ObjFaninC0(fanout)
	}
	if Ivy_ObjFanin1(fanout) == o {
		return Ivy_ObjFaninC1(fanout)
	}
	panic("object does not drive this fanout")
}
func Ivy_ObjCreateGhost(p *Ivy_Man_t, a, b *Ivy_Obj_t, typ Ivy_Type_t, init Ivy_Init_t) *Ivy_Obj_t {
	if p == nil || p.memoryStopped || typ <= IVY_NONE || typ >= IVY_VOID {
		panic("invalid ghost manager/type")
	}
	for _, fanin := range []*Ivy_Obj_t{a, b} {
		if fanin != nil && (Ivy_Regular(fanin).owner != p || Ivy_Regular(fanin).recycled) {
			panic("ghost fanin is not live in this manager")
		}
	}
	if typ == IVY_PI {
		if a != nil || b != nil || init != IVY_INIT_NONE {
			panic("PI ghost has fanins/initial state")
		}
	} else if a == nil {
		panic("ghost requires first fanin")
	}
	if typ == IVY_AND || typ == IVY_EXOR {
		if b == nil {
			panic("binary ghost requires second fanin")
		}
	} else if b != nil {
		panic("unary ghost has second fanin")
	}
	if typ == IVY_LATCH {
		if init < IVY_INIT_0 || init > IVY_INIT_DC {
			panic("latch ghost needs initial state")
		}
	} else if init != IVY_INIT_NONE {
		panic("nonlatch ghost has initial state")
	}
	if typ == IVY_AND && (a == nil || Ivy_ObjIsConst1(a) != 0) {
		panic("AND ghost needs nonconstant fanin")
	}
	if b != nil && Ivy_ObjIsConst1(b) != 0 {
		panic("second ghost fanin must be nonconstant")
	}
	if typ != IVY_PI && a != nil && b != nil && Ivy_Regular(a) == Ivy_Regular(b) {
		panic("ghost fanins need distinct regular nodes")
	}
	if typ == IVY_LATCH && Ivy_IsComplement(a) != 0 {
		panic("latch ghost requires regular input")
	}
	ghost := Ivy_ManGhost(p)
	ghost.Id = -1
	ghost.Type = typ
	ghost.Init = init
	ghost.pFanin0, ghost.pFanin1 = a, b
	if b != nil && Ivy_ObjFaninId0(ghost) > Ivy_ObjFaninId1(ghost) {
		ghost.pFanin0, ghost.pFanin1 = b, a
	}
	return ghost
}
func Ivy_InitNotCond(init Ivy_Init_t, compl int32) Ivy_Init_t {
	if init < IVY_INIT_0 || init > IVY_INIT_DC {
		panic("missing latch initialization")
	}
	if compl == 0 || init == IVY_INIT_DC {
		return init
	}
	if init == IVY_INIT_0 {
		return IVY_INIT_1
	}
	if init == IVY_INIT_1 {
		return IVY_INIT_0
	}
	panic("invalid initialization")
}
func Ivy_InitAnd(a, b Ivy_Init_t) Ivy_Init_t {
	if a < IVY_INIT_0 || a > IVY_INIT_DC || b < IVY_INIT_0 || b > IVY_INIT_DC {
		panic("missing latch initialization")
	}
	if a == IVY_INIT_0 || b == IVY_INIT_0 {
		return IVY_INIT_0
	}
	if a == IVY_INIT_DC || b == IVY_INIT_DC {
		return IVY_INIT_DC
	}
	return IVY_INIT_1
}
func Ivy_InitExor(a, b Ivy_Init_t) Ivy_Init_t {
	if a < IVY_INIT_0 || a > IVY_INIT_DC || b < IVY_INIT_0 || b > IVY_INIT_DC {
		panic("missing latch initialization")
	}
	if a == IVY_INIT_DC || b == IVY_INIT_DC {
		return IVY_INIT_DC
	}
	if a != b {
		return IVY_INIT_1
	}
	return IVY_INIT_0
}
