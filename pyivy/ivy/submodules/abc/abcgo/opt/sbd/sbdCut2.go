package sbd

// Port of src/opt/sbd/sbdCut2.c: the delay-oriented cut server.

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Sbd_Srv_t struct {
	nLutSize int32
	nCutSize int32
	nCutNum  int32
	fVerbose int32
	pGia     *Gia_Man_t // user's AIG manager (will be modified by adding nodes)
	vMirrors *Vec_Int_t // mirrors for each node
	vLutLevs *Vec_Int_t // delays for each node
	vLevs    *Vec_Int_t // levels for each node
	vRefs    *Vec_Int_t // refs for each node
	clkStart int64      // starting time
	vCut0    *Vec_Int_t // current cut
	vCut     *Vec_Int_t // current cut
	vCutTop  *Vec_Int_t // current cut
	vCutBot  *Vec_Int_t // current cut
}

func Sbd_ManCutServerStart(pGia *Gia_Man_t, vMirrors *Vec_Int_t,
	vLutLevs, vLevs, vRefs *Vec_Int_t,
	nLutSize, nCutSize, nCutNum, fVerbose int32) *Sbd_Srv_t {
	assert(nLutSize <= nCutSize, "nLutSize <= nCutSize")
	assert(nCutSize < SBD_CUT_NO_LEAF, "nCutSize < SBD_CUT_NO_LEAF")
	assert(nCutSize > 1 && nCutSize <= SBD_MAX_CUTSIZE, "nCutSize > 1 && nCutSize <= SBD_MAX_CUTSIZE")
	assert(nCutNum > 1 && nCutNum < SBD_MAX_CUTNUM, "nCutNum > 1 && nCutNum < SBD_MAX_CUTNUM")
	p := &Sbd_Srv_t{}
	p.clkStart = util.Abc_Clock()
	p.nLutSize = nLutSize
	p.nCutSize = nCutSize
	p.nCutNum = nCutNum
	p.fVerbose = fVerbose
	p.pGia = pGia
	p.vMirrors = vMirrors
	p.vLutLevs = vLutLevs
	p.vLevs = vLevs
	p.vRefs = vRefs
	p.vCut0 = vec.Vec_IntAlloc(100)
	p.vCut = vec.Vec_IntAlloc(100)
	p.vCutTop = vec.Vec_IntAlloc(100)
	p.vCutBot = vec.Vec_IntAlloc(100)
	return p
}
func Sbd_ManCutServerStop(p *Sbd_Srv_t) {}

func Sbd_ManCutIsTopo_rec(p *Gia_Man_t, vMirrors *Vec_Int_t, iObj int32) int32 {
	if vec.Vec_IntEntry(vMirrors, int(iObj)) >= 0 {
		iObj = Abc_Lit2Var(vec.Vec_IntEntry(vMirrors, int(iObj)))
	}
	if iObj == 0 || Gia_ObjIsTravIdCurrentId(p, iObj) {
		return 1
	}
	Gia_ObjSetTravIdCurrentId(p, iObj)
	pObj := Gia_ManObj(p, iObj)
	if Gia_ObjIsCi(pObj) {
		return 0
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	Ret0 := Sbd_ManCutIsTopo_rec(p, vMirrors, Gia_ObjFaninId0(pObj, iObj))
	Ret1 := Sbd_ManCutIsTopo_rec(p, vMirrors, Gia_ObjFaninId1(pObj, iObj))
	return b2i(Ret0 != 0 && Ret1 != 0)
}
func Sbd_ManCutIsTopo(p *Gia_Man_t, vMirrors, vCut *Vec_Int_t, iObj int32) int32 {
	Gia_ManIncrementTravId(p)
	for _, Entry := range vec.Vec_IntArray(vCut) {
		Gia_ObjSetTravIdCurrentId(p, Entry)
	}
	RetValue := Sbd_ManCutIsTopo_rec(p, vMirrors, iObj)
	if RetValue == 0 {
		fmt.Printf("Cut of node %d is not tological\n", iObj)
	}
	assert(RetValue != 0, "RetValue")
	return RetValue
}

func Sbd_ManCutExpandOne(p *Gia_Man_t, vMirrors, vLutLevs, vCut *Vec_Int_t, iThis, iObj int32) int32 {
	LutLev := vec.Vec_IntEntry(vLutLevs, int(iObj))
	pObj := Gia_ManObj(p, iObj)
	if Gia_ObjIsCi(pObj) {
		return 0
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	Lit0m := vec.Vec_IntEntry(vMirrors, int(Gia_ObjFaninId0(pObj, iObj)))
	Lit1m := vec.Vec_IntEntry(vMirrors, int(Gia_ObjFaninId1(pObj, iObj)))
	Fan0 := Gia_ObjFaninId0(pObj, iObj)
	if Lit0m >= 0 {
		Fan0 = Abc_Lit2Var(Lit0m)
	}
	Fan1 := Gia_ObjFaninId1(pObj, iObj)
	if Lit1m >= 0 {
		Fan1 = Abc_Lit2Var(Lit1m)
	}
	iPlace0 := vec.Vec_IntFind(vCut, Fan0)
	iPlace1 := vec.Vec_IntFind(vCut, Fan1)
	if iPlace0 == -1 && iPlace1 == -1 {
		return 0
	}
	if vec.Vec_IntEntry(vLutLevs, int(Fan0)) > LutLev || vec.Vec_IntEntry(vLutLevs, int(Fan1)) > LutLev {
		return 0
	}
	vec.Vec_IntDrop(vCut, int(iThis))
	if iPlace0 == -1 && Fan0 != 0 {
		vec.Vec_IntPushOrder(vCut, Fan0)
	}
	if iPlace1 == -1 && Fan1 != 0 {
		vec.Vec_IntPushOrder(vCut, Fan1)
	}
	return 1
}
func Vec_IntIsOrdered(vCut *Vec_Int_t) {
	Prev := vec.Vec_IntEntry(vCut, 0)
	for i := 1; i < vec.Vec_IntSize(vCut); i++ {
		Entry := vec.Vec_IntEntry(vCut, i)
		assert(Prev < Entry, "Prev < Entry")
		Prev = Entry
	}
}
func Sbd_ManCutExpand(p *Gia_Man_t, vMirrors, vLutLevs, vCut *Vec_Int_t) {
	var i int
	for {
		for i = 0; i < vec.Vec_IntSize(vCut); i++ {
			if Sbd_ManCutExpandOne(p, vMirrors, vLutLevs, vCut, int32(i), vec.Vec_IntEntry(vCut, i)) != 0 {
				break
			}
		}
		if !(i < vec.Vec_IntSize(vCut)) {
			break
		}
	}
}
func Sbd_ManCutReload(vMirrors, vLutLevs *Vec_Int_t, LevStop int32, vCut, vCutTop, vCutBot *Vec_Int_t) {
	vec.Vec_IntClear(vCutTop)
	vec.Vec_IntClear(vCutBot)
	for _, Entry := range vec.Vec_IntArray(vCut) {
		assert(Entry != 0, "Entry")
		assert(vec.Vec_IntEntry(vMirrors, int(Entry)) == -1, "Vec_IntEntry(vMirrors, Entry) == -1")
		assert(vec.Vec_IntEntry(vLutLevs, int(Entry)) <= LevStop, "Vec_IntEntry(vLutLevs, Entry) <= LevStop")
		if vec.Vec_IntEntry(vLutLevs, int(Entry)) == LevStop {
			vec.Vec_IntPush(vCutTop, Entry)
		} else {
			vec.Vec_IntPush(vCutBot, Entry)
		}
	}
	Vec_IntIsOrdered(vCut)
}
func Sbd_ManCutCollect_rec(p *Gia_Man_t, vMirrors *Vec_Int_t, iObj, LevStop int32, vLutLevs, vCut *Vec_Int_t) int32 {
	if vec.Vec_IntEntry(vMirrors, int(iObj)) >= 0 {
		iObj = Abc_Lit2Var(vec.Vec_IntEntry(vMirrors, int(iObj)))
	}
	if iObj == 0 || Gia_ObjIsTravIdCurrentId(p, iObj) {
		return 1
	}
	Gia_ObjSetTravIdCurrentId(p, iObj)
	pObj := Gia_ManObj(p, iObj)
	if Gia_ObjIsCi(pObj) || vec.Vec_IntEntry(vLutLevs, int(iObj)) <= LevStop {
		vec.Vec_IntPush(vCut, iObj)
		return b2i(vec.Vec_IntEntry(vLutLevs, int(iObj)) <= LevStop)
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	Ret0 := Sbd_ManCutCollect_rec(p, vMirrors, Gia_ObjFaninId0(pObj, iObj), LevStop, vLutLevs, vCut)
	Ret1 := Sbd_ManCutCollect_rec(p, vMirrors, Gia_ObjFaninId1(pObj, iObj), LevStop, vLutLevs, vCut)
	return b2i(Ret0 != 0 && Ret1 != 0)
}

func Sbd_ManCutReduceTop(p *Gia_Man_t, vMirrors *Vec_Int_t, iObj int32, vLutLevs, vCut, vCutTop *Vec_Int_t, nCutSize int32) int32 {
	LevStop := vec.Vec_IntEntry(vLutLevs, int(iObj)) - 2
	Vec_IntIsOrdered(vCut)
	for i := vec.Vec_IntSize(vCutTop) - 1; i >= 0; i-- {
		Entry := vec.Vec_IntEntry(vCutTop, i)
		pObj := Gia_ManObj(p, Entry)
		if Gia_ObjIsCi(pObj) {
			continue
		}
		assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
		assert(vec.Vec_IntEntry(vLutLevs, int(Entry)) == LevStop, "Vec_IntEntry(vLutLevs, Entry) == LevStop")
		Lit0m := vec.Vec_IntEntry(vMirrors, int(Gia_ObjFaninId0(pObj, Entry)))
		Lit1m := vec.Vec_IntEntry(vMirrors, int(Gia_ObjFaninId1(pObj, Entry)))
		Fan0 := Gia_ObjFaninId0(pObj, Entry)
		if Lit0m >= 0 {
			Fan0 = Abc_Lit2Var(Lit0m)
		}
		Fan1 := Gia_ObjFaninId1(pObj, Entry)
		if Lit1m >= 0 {
			Fan1 = Abc_Lit2Var(Lit1m)
		}
		if vec.Vec_IntEntry(vLutLevs, int(Fan0)) > LevStop || vec.Vec_IntEntry(vLutLevs, int(Fan1)) > LevStop {
			continue
		}
		if vec.Vec_IntEntry(vLutLevs, int(Fan0)) == LevStop && vec.Vec_IntEntry(vLutLevs, int(Fan1)) == LevStop {
			continue
		}
		vec.Vec_IntRemove(vCut, Entry)
		if Fan0 != 0 {
			vec.Vec_IntPushUniqueOrder(vCut, Fan0)
		}
		if Fan1 != 0 {
			vec.Vec_IntPushUniqueOrder(vCut, Fan1)
		}
		return 1
	}
	return 0
}

// sbdSrvTake copies the cut into pLeaves (memcpy) and returns its size.
func sbdSrvTake(p *Sbd_Srv_t, pLeaves []int32) int32 {
	copy(pLeaves, vec.Vec_IntArray(p.vCut))
	return int32(vec.Vec_IntSize(p.vCut))
}

func Sbd_ManCutServerFirst(p *Sbd_Srv_t, iObj int32, pLeaves []int32) int32 {
	LevStop := vec.Vec_IntEntry(p.vLutLevs, int(iObj)) - 2
	vec.Vec_IntClear(p.vCut)
	Gia_ManIncrementTravId(p.pGia)
	RetValue := Sbd_ManCutCollect_rec(p.pGia, p.vMirrors, iObj, LevStop, p.vLutLevs, p.vCut)
	if RetValue == 0 { // cannot build delay-improving cut
		return -1
	}
	// check if the current cut is good
	vec.Vec_IntSort(p.vCut, 0)
	// try to expand the cut
	Sbd_ManCutExpand(p.pGia, p.vMirrors, p.vLutLevs, p.vCut)
	Sbd_ManCutReload(p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot)
	if int32(vec.Vec_IntSize(p.vCut)) <= p.nCutSize && int32(vec.Vec_IntSize(p.vCutTop)) <= p.nLutSize-1 {
		return sbdSrvTake(p, pLeaves)
	}
	// try to reduce the topmost
	vec.Vec_IntClear(p.vCut0)
	vec.Vec_IntAppend(p.vCut0, p.vCut)
	for try := 0; try < 4; try++ {
		if !(int32(vec.Vec_IntSize(p.vCut)) < p.nCutSize && Sbd_ManCutReduceTop(p.pGia, p.vMirrors, iObj, p.vLutLevs, p.vCut, p.vCutTop, p.nCutSize) != 0) {
			break
		}
		Sbd_ManCutExpand(p.pGia, p.vMirrors, p.vLutLevs, p.vCut)
		Sbd_ManCutReload(p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot)
		assert(int32(vec.Vec_IntSize(p.vCut)) <= p.nCutSize, "Vec_IntSize(p->vCut) <= p->nCutSize")
		if int32(vec.Vec_IntSize(p.vCutTop)) <= p.nLutSize-1 {
			return sbdSrvTake(p, pLeaves)
		}
	}
	// recompute the cut
	vec.Vec_IntClear(p.vCut)
	Gia_ManIncrementTravId(p.pGia)
	RetValue = Sbd_ManCutCollect_rec(p.pGia, p.vMirrors, iObj, LevStop-1, p.vLutLevs, p.vCut)
	if RetValue == 0 { // cannot build delay-improving cut
		return -1
	}
	// check if the current cut is good
	vec.Vec_IntSort(p.vCut, 0)
	// try to expand the cut
	Sbd_ManCutExpand(p.pGia, p.vMirrors, p.vLutLevs, p.vCut)
	Sbd_ManCutReload(p.vMirrors, p.vLutLevs, LevStop, p.vCut, p.vCutTop, p.vCutBot)
	if int32(vec.Vec_IntSize(p.vCut)) <= p.nCutSize && int32(vec.Vec_IntSize(p.vCutTop)) <= p.nLutSize-1 {
		return sbdSrvTake(p, pLeaves)
	}
	return -1
}
