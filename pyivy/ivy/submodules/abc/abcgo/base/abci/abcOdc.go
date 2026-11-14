package abci

import (
	"fmt"
	"time"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/res"
)

const ABC_DC_MAX_NODES = 1 << 15

type Odc_Lit_t = uint16

type Odc_Obj_t struct {
	iFan0, iFan1, iNext Odc_Lit_t
	TravId              uint16
	uData, uMask        uint32
}

type Odc_Man_t_ struct {
	nVarsMax, nLevels, fVerbose, fVeryVerbose, nPercCutoff int32
	pNode                                                  *Abc_Obj_t
	vLeaves, vRoots, vBranches                             *Vec_Ptr_t
	nPis, nObjs, nObjsAlloc                                int32
	pObjs                                                  []*Odc_Obj_t
	iRoot                                                  Odc_Lit_t
	nTravIds                                               uint16
	pTable                                                 []Odc_Lit_t
	nTableSize                                             int32
	vUsedSpots                                             *Vec_Int_t
	nBits, nWords                                          int32
	vTruths, vTruthsElem                                   *Vec_Ptr_t
	puTruth                                                []uint32
	nWins, nWinsEmpty, nSimsEmpty, nQuantsOver             int32
	nWinsFinish, nTotalDcs                                 int32
	timeClean, timeWin, timeMiter, timeSim                 int64
	timeQuant, timeTruth, timeTotal, timeAbort             int64
	rng                                                    *odcRand
}

type Odc_Man_t = Odc_Man_t_

type odcRand struct {
	state [31]uint32
	index uint64
}

func newOdcRand(seed uint32) *odcRand {
	if seed == 0 {
		seed = 1
	}
	seq := make([]uint32, 344)
	seq[0] = seed
	for i := 1; i < 31; i++ {
		seq[i] = uint32((uint64(16807) * uint64(seq[i-1])) % 2147483647)
	}
	seq[31], seq[32], seq[33] = seq[0], seq[1], seq[2]
	for i := 34; i < len(seq); i++ {
		seq[i] = seq[i-31] + seq[i-3]
	}
	r := &odcRand{index: 344}
	for i := 313; i < 344; i++ {
		r.state[i%31] = seq[i]
	}
	return r
}
func (r *odcRand) next() uint32 {
	i := r.index
	v := r.state[(i-31)%31] + r.state[(i-3)%31]
	r.state[i%31] = v
	r.index++
	return v >> 1
}

func Odc_PiNum(p *Odc_Man_t) int32       { return p.nPis }
func Odc_NodeNum(p *Odc_Man_t) int32     { return p.nObjs - p.nPis - 1 }
func Odc_ObjNum(p *Odc_Man_t) int32      { return p.nObjs }
func Odc_IsComplement(l Odc_Lit_t) int32 { return int32(l & 1) }
func Odc_Regular(l Odc_Lit_t) Odc_Lit_t  { return l &^ 1 }
func Odc_Not(l Odc_Lit_t) Odc_Lit_t      { return l ^ 1 }
func Odc_NotCond(l Odc_Lit_t, c int32) Odc_Lit_t {
	if c != 0 {
		return l ^ 1
	}
	return l
}
func Odc_Const0() Odc_Lit_t { return 1 }
func Odc_Const1() Odc_Lit_t { return 0 }
func Odc_Var(p *Odc_Man_t, i int32) Odc_Lit_t {
	assert(i >= 0 && i < p.nPis, "i >= 0 && i < p.nPis")
	return Odc_Lit_t((i + 1) << 1)
}
func Odc_IsConst(l Odc_Lit_t) int32 {
	if l < 2 {
		return 1
	}
	return 0
}
func Odc_IsTerm(p *Odc_Man_t, l Odc_Lit_t) int32 {
	if int32(l>>1) <= p.nPis {
		return 1
	}
	return 0
}
func Odc_ObjNew(p *Odc_Man_t) *Odc_Obj_t {
	assert(p.nObjs < p.nObjsAlloc, "ODC object capacity")
	o := &Odc_Obj_t{}
	p.pObjs = append(p.pObjs, o)
	p.nObjs++
	return o
}
func Odc_Obj2Lit(p *Odc_Man_t, o *Odc_Obj_t) Odc_Lit_t {
	for i, x := range p.pObjs {
		if x == o {
			return Odc_Lit_t(i << 1)
		}
	}
	panic("ODC object not in manager")
}
func Odc_Lit2Obj(p *Odc_Man_t, l Odc_Lit_t) *Odc_Obj_t {
	assert(l&1 == 0 && int32(l>>1) < p.nObjs, "invalid ODC literal")
	return p.pObjs[int(l>>1)]
}
func Odc_ObjChild0(o *Odc_Obj_t) Odc_Lit_t               { return o.iFan0 }
func Odc_ObjChild1(o *Odc_Obj_t) Odc_Lit_t               { return o.iFan1 }
func Odc_ObjFanin0(o *Odc_Obj_t) Odc_Lit_t               { return Odc_Regular(o.iFan0) }
func Odc_ObjFanin1(o *Odc_Obj_t) Odc_Lit_t               { return Odc_Regular(o.iFan1) }
func Odc_ObjFaninC0(o *Odc_Obj_t) int32                  { return Odc_IsComplement(o.iFan0) }
func Odc_ObjFaninC1(o *Odc_Obj_t) int32                  { return Odc_IsComplement(o.iFan1) }
func Odc_ManIncrementTravId(p *Odc_Man_t)                { p.nTravIds++ }
func Odc_ObjSetTravIdCurrent(p *Odc_Man_t, o *Odc_Obj_t) { o.TravId = p.nTravIds }
func Odc_ObjIsTravIdCurrent(p *Odc_Man_t, o *Odc_Obj_t) int32 {
	if o.TravId == p.nTravIds {
		return 1
	}
	return 0
}
func Odc_ObjTruth(p *Odc_Man_t, l Odc_Lit_t) []uint32 {
	assert(l&1 == 0, "truth literal must be regular")
	return vec.Vec_PtrEntry(p.vTruths, int(l>>1)).([]uint32)
}

func Abc_NtkDontCareAlloc(nVarsMax, nLevels, fVerbose, fVeryVerbose int32) *Odc_Man_t {
	assert(nVarsMax > 4 && nVarsMax < 16 && nLevels > 0 && nLevels < 10, "invalid ODC parameters")
	p := &Odc_Man_t_{nVarsMax: nVarsMax, nLevels: nLevels, fVerbose: fVerbose, fVeryVerbose: fVeryVerbose, nPercCutoff: 10, nObjsAlloc: ABC_DC_MAX_NODES, nPis: nVarsMax + 32, rng: newOdcRand(0xABC)}
	p.vRoots, p.vBranches = vec.Vec_PtrAlloc(128), vec.Vec_PtrAlloc(128)
	p.nObjs = 1 + p.nPis
	p.pObjs = make([]*Odc_Obj_t, p.nObjs)
	for i := range p.pObjs {
		p.pObjs[i] = &Odc_Obj_t{}
	}
	for i := int32(0); i < 32; i++ {
		p.pObjs[1+int(nVarsMax+i)].uMask = uint32(1) << uint(i)
	}
	p.nTableSize = p.nObjsAlloc/3 + 1
	p.pTable = make([]Odc_Lit_t, p.nTableSize)
	p.vUsedSpots = vec.Vec_IntAlloc(1000)
	p.nWords = util.Abc_TruthWordNum(nVarsMax)
	p.nBits = p.nWords * 32
	p.vTruths = vec.Vec_PtrAllocSimInfo(int(p.nObjsAlloc), int(p.nWords))
	p.vTruthsElem = vec.Vec_PtrAllocSimInfo(int(nVarsMax), int(p.nWords))
	res.Abc_InfoFill(vec.Vec_PtrEntry(p.vTruths, 0).([]uint32), p.nWords)
	for k := int32(0); k < nVarsMax; k++ {
		d := vec.Vec_PtrEntry(p.vTruthsElem, int(k)).([]uint32)
		for bit := int32(0); bit < p.nBits; bit++ {
			if bit&(1<<uint(k)) != 0 {
				d[bit>>5] |= uint32(1) << uint(bit&31)
			}
		}
	}
	for k := nVarsMax; k < p.nPis; k++ {
		d := Odc_ObjTruth(p, Odc_Var(p, k))
		for i := len(d) - 1; i >= 0; i-- {
			d[i] = (p.rng.next() << 24) ^ (p.rng.next() << 12) ^ p.rng.next()
		}
	}
	p.iRoot = 0xffff
	return p
}

func Abc_NtkDontCareClear(p *Odc_Man_t) {
	clk := util.Abc_Clock()
	if vec.Vec_IntSize(p.vUsedSpots) > int(p.nTableSize/3) {
		clear(p.pTable)
	} else {
		for i := 0; i < vec.Vec_IntSize(p.vUsedSpots); i++ {
			p.pTable[vec.Vec_IntEntry(p.vUsedSpots, i)] = 0
		}
	}
	vec.Vec_IntClear(p.vUsedSpots)
	p.nObjs = 1 + p.nPis
	p.pObjs = p.pObjs[:p.nObjs]
	p.iRoot = 0xffff
	p.timeClean += util.Abc_Clock() - clk
}
func Abc_NtkDontCareFree(p *Odc_Man_t) {
	if p == nil {
		return
	}
	if p.fVerbose != 0 {
		fmt.Printf("Wins = %5d. Empty = %5d. SimsEmpty = %5d. QuantOver = %5d. WinsFinish = %5d.\n", p.nWins, p.nWinsEmpty, p.nSimsEmpty, p.nQuantsOver, p.nWinsFinish)
		fmt.Printf("Ave DCs per window = %6.2f %%. Ave DCs per finished window = %6.2f %%.\n", float64(p.nTotalDcs)/float64(p.nWins), float64(p.nTotalDcs)/float64(p.nWinsFinish))
		for _, x := range []struct {
			name string
			t    int64
		}{{"Cleaning", p.timeClean}, {"Windowing", p.timeWin}, {"Miter", p.timeMiter}, {"Simulation", p.timeSim}, {"Quantifying", p.timeQuant}, {"Truth table", p.timeTruth}, {"TOTAL", p.timeTotal}, {"Aborted", p.timeAbort}} {
			fmt.Printf("%-12s %9.2f sec\n", x.name, float64(x.t)/float64(time.Second))
		}
	}
	vec.Vec_PtrFree(p.vRoots)
	vec.Vec_PtrFree(p.vBranches)
	vec.Vec_PtrFree(p.vTruths)
	vec.Vec_PtrFree(p.vTruthsElem)
	vec.Vec_IntFree(p.vUsedSpots)
}

func odcEntry(p *Odc_Man_t, v *Vec_Ptr_t, i int32) *Abc_Obj_t {
	x := vec.Vec_PtrEntry(v, int(i))
	o, _ := x.(*Abc_Obj_t)
	if o == nil {
		o, _ = x.(*baseabc.Abc_Obj_t)
	}
	return o
}
func Abc_NtkDontCareWinSweepLeafTfo_rec(o *Abc_Obj_t, limit int32, pivot *Abc_Obj_t) {
	if baseabc.Abc_ObjIsCo(o) != 0 || int32(o.Level) > limit || o == pivot || baseabc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return
	}
	baseabc.Abc_NodeSetTravIdCurrent(o)
	if baseabc.Abc_ObjFanoutNum(o) > 100 {
		return
	}
	for i := int32(0); i < baseabc.Abc_ObjFanoutNum(o); i++ {
		Abc_NtkDontCareWinSweepLeafTfo_rec(baseabc.Abc_ObjFanout(o, i), limit, pivot)
	}
}
func Abc_NtkDontCareWinSweepLeafTfo(p *Odc_Man_t) {
	baseabc.Abc_NtkIncrementTravId(p.pNode.Network())
	for i := int32(0); i < int32(vec.Vec_PtrSize(p.vLeaves)); i++ {
		Abc_NtkDontCareWinSweepLeafTfo_rec(odcEntry(p, p.vLeaves, i), int32(p.pNode.Level)+p.nLevels, p.pNode)
	}
}
func Abc_NtkDontCareWinCollectRoots_rec(o *Abc_Obj_t, roots *Vec_Ptr_t) {
	assert(baseabc.Abc_ObjIsNode(o) != 0 && baseabc.Abc_NodeIsTravIdCurrent(o) != 0, "ODC root traversal")
	for i := int32(0); i < baseabc.Abc_ObjFanoutNum(o); i++ {
		if baseabc.Abc_NodeIsTravIdCurrent(baseabc.Abc_ObjFanout(o, i)) == 0 {
			vec.Vec_PtrPushUnique(roots, o)
			return
		}
	}
	for i := int32(0); i < baseabc.Abc_ObjFanoutNum(o); i++ {
		Abc_NtkDontCareWinCollectRoots_rec(baseabc.Abc_ObjFanout(o, i), roots)
	}
}
func Abc_NtkDontCareWinCollectRoots(p *Odc_Man_t) {
	assert(baseabc.Abc_NodeIsTravIdCurrent(p.pNode) == 0, "pivot traversal state")
	baseabc.Abc_NodeSetTravIdCurrent(p.pNode)
	vec.Vec_PtrClear(p.vRoots)
	Abc_NtkDontCareWinCollectRoots_rec(p.pNode, p.vRoots)
}
func Abc_NtkDontCareWinAddMissing_rec(p *Odc_Man_t, o *Abc_Obj_t) int32 {
	if baseabc.Abc_NodeIsTravIdCurrent(o) != 0 {
		return 1
	}
	if baseabc.Abc_NodeIsTravIdPrevious(o) == 0 || baseabc.Abc_ObjIsCi(o) != 0 {
		baseabc.Abc_NodeSetTravIdCurrent(o)
		vec.Vec_PtrPush(p.vBranches, o)
		if vec.Vec_PtrSize(p.vBranches) <= 32 {
			return 1
		}
		return 0
	}
	for i := int32(0); i < baseabc.Abc_ObjFaninNum(o); i++ {
		if Abc_NtkDontCareWinAddMissing_rec(p, baseabc.Abc_ObjFanin(o, i)) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkDontCareWinAddMissing(p *Odc_Man_t) int32 {
	baseabc.Abc_NtkIncrementTravId(p.pNode.Network())
	for i := int32(0); i < int32(vec.Vec_PtrSize(p.vLeaves)); i++ {
		baseabc.Abc_NodeSetTravIdCurrent(odcEntry(p, p.vLeaves, i))
	}
	vec.Vec_PtrClear(p.vBranches)
	for i := int32(0); i < int32(vec.Vec_PtrSize(p.vRoots)); i++ {
		if Abc_NtkDontCareWinAddMissing_rec(p, odcEntry(p, p.vRoots, i)) == 0 {
			return 0
		}
	}
	return 1
}
func Abc_NtkDontCareWindow(p *Odc_Man_t) int32 {
	Abc_NtkDontCareWinSweepLeafTfo(p)
	Abc_NtkDontCareWinCollectRoots(p)
	if vec.Vec_PtrSize(p.vRoots) == 1 && odcEntry(p, p.vRoots, 0) == p.pNode {
		return 0
	}
	return Abc_NtkDontCareWinAddMissing(p)
}

func Odc_HashKey(a, b Odc_Lit_t, size int32) uint32 {
	return uint32((uint32(Odc_Regular(a))*7937 ^ uint32(Odc_Regular(b))*2971 ^ uint32(Odc_IsComplement(a))*911 ^ uint32(Odc_IsComplement(b))*353) % uint32(size))
}
func Odc_HashLookup(p *Odc_Man_t, a, b Odc_Lit_t) *Odc_Lit_t {
	assert(a < b, "canonical ODC fanins")
	h := int(Odc_HashKey(a, b, p.nTableSize))
	if p.pTable[h] == 0 {
		vec.Vec_IntPush(p.vUsedSpots, int32(h))
	}
	link := &p.pTable[h]
	for *link != 0 {
		o := Odc_Lit2Obj(p, *link)
		if o.iFan0 == a && o.iFan1 == b {
			return link
		}
		link = &o.iNext
	}
	return link
}
func Odc_And(p *Odc_Man_t, a, b Odc_Lit_t) Odc_Lit_t {
	if a == b {
		return a
	}
	if a == Odc_Not(b) {
		return Odc_Const0()
	}
	if Odc_Regular(a) == 0 {
		if a == 0 {
			return b
		}
		return 1
	}
	if Odc_Regular(b) == 0 {
		if b == 0 {
			return a
		}
		return 1
	}
	if a > b {
		a, b = b, a
	}
	entry := Odc_HashLookup(p, a, b)
	if *entry != 0 {
		return *entry
	}
	o := Odc_ObjNew(p)
	o.iFan0 = a
	o.iFan1 = b
	o.uMask = Odc_Lit2Obj(p, Odc_Regular(a)).uMask | Odc_Lit2Obj(p, Odc_Regular(b)).uMask
	*entry = Odc_Obj2Lit(p, o)
	return *entry
}
func Odc_Or(p *Odc_Man_t, a, b Odc_Lit_t) Odc_Lit_t {
	return Odc_Not(Odc_And(p, Odc_Not(a), Odc_Not(b)))
}
func Odc_Xor(p *Odc_Man_t, a, b Odc_Lit_t) Odc_Lit_t {
	return Odc_Or(p, Odc_And(p, a, Odc_Not(b)), Odc_And(p, Odc_Not(a), b))
}

func odcTransferRec(p *Odc_Man_t, node, pivot *Abc_Obj_t, memo map[*Abc_Obj_t]uint32) uint32 {
	if v, ok := memo[node]; ok {
		return v
	}
	if node == pivot {
		return uint32(Odc_Const1())<<16 | uint32(Odc_Const0())
	}
	a := odcTransferRec(p, baseabc.Abc_ObjFanin0(node), pivot, memo)
	b := odcTransferRec(p, baseabc.Abc_ObjFanin1(node), pivot, memo)
	lo := Odc_And(p, Odc_NotCond(Odc_Lit_t(a), baseabc.Abc_ObjFaninC0(node)), Odc_NotCond(Odc_Lit_t(b), baseabc.Abc_ObjFaninC1(node)))
	hi := Odc_And(p, Odc_NotCond(Odc_Lit_t(a>>16), baseabc.Abc_ObjFaninC0(node)), Odc_NotCond(Odc_Lit_t(b>>16), baseabc.Abc_ObjFaninC1(node)))
	v := uint32(hi)<<16 | uint32(lo)
	memo[node] = v
	return v
}
func Abc_NtkDontCareTransfer_rec(p *Odc_Man_t, node, pivot *Abc_Obj_t) uint32 {
	return odcTransferRec(p, node, pivot, map[*Abc_Obj_t]uint32{})
}
func Abc_NtkDontCareTransfer(p *Odc_Man_t) int32 {
	memo := map[*Abc_Obj_t]uint32{}
	for i := int32(0); i < int32(vec.Vec_PtrSize(p.vLeaves)); i++ {
		memo[odcEntry(p, p.vLeaves, i)] = uint32(Odc_Var(p, i))<<16 | uint32(Odc_Var(p, i))
	}
	for i := int32(0); i < int32(vec.Vec_PtrSize(p.vBranches)); i++ {
		l := Odc_Var(p, i+p.nVarsMax)
		memo[odcEntry(p, p.vBranches, i)] = uint32(l)<<16 | uint32(l)
	}
	p.iRoot = Odc_Const0()
	for i := int32(0); i < int32(vec.Vec_PtrSize(p.vRoots)); i++ {
		d := odcTransferRec(p, odcEntry(p, p.vRoots, i), p.pNode, memo)
		p.iRoot = Odc_Or(p, p.iRoot, Odc_Xor(p, Odc_Lit_t(d), Odc_Lit_t(d>>16)))
	}
	return 1
}

func Abc_NtkDontCareCofactors_rec(p *Odc_Man_t, l Odc_Lit_t, mask uint32) uint32 {
	assert(l&1 == 0, "regular cofactors literal")
	o := Odc_Lit2Obj(p, l)
	if Odc_ObjIsTravIdCurrent(p, o) != 0 {
		return o.uData
	}
	Odc_ObjSetTravIdCurrent(p, o)
	if o.uMask&mask == 0 {
		o.uData = uint32(l)<<16 | uint32(l)
		return o.uData
	}
	if o.uMask == mask && Odc_IsTerm(p, l) != 0 {
		o.uData = uint32(Odc_Const1())<<16 | uint32(Odc_Const0())
		return o.uData
	}
	a := Abc_NtkDontCareCofactors_rec(p, Odc_ObjFanin0(o), mask)
	b := Abc_NtkDontCareCofactors_rec(p, Odc_ObjFanin1(o), mask)
	lo := Odc_And(p, Odc_NotCond(Odc_Lit_t(a), Odc_ObjFaninC0(o)), Odc_NotCond(Odc_Lit_t(b), Odc_ObjFaninC1(o)))
	hi := Odc_And(p, Odc_NotCond(Odc_Lit_t(a>>16), Odc_ObjFaninC0(o)), Odc_NotCond(Odc_Lit_t(b>>16), Odc_ObjFaninC1(o)))
	o.uData = uint32(hi)<<16 | uint32(lo)
	return o.uData
}
func Abc_NtkDontCareQuantify(p *Odc_Man_t) int32 {
	assert(p.iRoot < 0xffff && vec.Vec_PtrSize(p.vBranches) <= 32, "ODC quantification bounds")
	for i := int32(0); i < int32(vec.Vec_PtrSize(p.vBranches)); i++ {
		Odc_ManIncrementTravId(p)
		d := Abc_NtkDontCareCofactors_rec(p, Odc_Regular(p.iRoot), uint32(1)<<uint(i))
		a := Odc_NotCond(Odc_Lit_t(d), Odc_IsComplement(p.iRoot))
		b := Odc_NotCond(Odc_Lit_t(d>>16), Odc_IsComplement(p.iRoot))
		p.iRoot = Odc_Or(p, a, b)
		if Odc_ObjNum(p) > ABC_DC_MAX_NODES/2 {
			return 0
		}
	}
	return 1
}

func Abc_NtkDontCareSimulateSetElem2(p *Odc_Man_t) {
	for k := int32(0); k < p.nVarsMax; k++ {
		d := Odc_ObjTruth(p, Odc_Var(p, k))
		clear(d)
		for bit := int32(0); bit < p.nBits; bit++ {
			if bit&(1<<uint(k)) != 0 {
				d[bit>>5] |= uint32(1) << uint(bit&31)
			}
		}
	}
}
func Abc_NtkDontCareSimulateSetElem(p *Odc_Man_t) {
	for k := int32(0); k < p.nVarsMax; k++ {
		copy(Odc_ObjTruth(p, Odc_Var(p, k)), vec.Vec_PtrEntry(p.vTruthsElem, int(k)).([]uint32))
	}
}
func Abc_NtkDontCareSimulateSetRand(p *Odc_Man_t) {
	for w := int32(0); w < p.nWords; w++ {
		n := p.rng.next()
		for k := int32(0); k < p.nVarsMax; k++ {
			d := Odc_ObjTruth(p, Odc_Var(p, k))
			if n&(uint32(1)<<uint(k)) != 0 {
				d[w] = ^uint32(0)
			} else {
				d[w] = 0
			}
		}
	}
}
func Abc_NtkDontCareCountMintsWord(p *Odc_Man_t, d []uint32) int32 {
	var n int32
	for w := int32(0); w < p.nWords; w++ {
		if d[w] != 0 {
			n++
		}
	}
	return n
}
func Abc_NtkDontCareTruthOne(p *Odc_Man_t, l Odc_Lit_t) {
	o := Odc_Lit2Obj(p, l)
	dst, a, b := Odc_ObjTruth(p, l), Odc_ObjTruth(p, Odc_ObjFanin0(o)), Odc_ObjTruth(p, Odc_ObjFanin1(o))
	for i := int32(0); i < p.nWords; i++ {
		x, y := a[i], b[i]
		if Odc_ObjFaninC0(o) != 0 {
			x = ^x
		}
		if Odc_ObjFaninC1(o) != 0 {
			y = ^y
		}
		dst[i] = x & y
	}
}
func Abc_NtkDontCareSimulate_rec(p *Odc_Man_t, l Odc_Lit_t) {
	if Odc_IsTerm(p, l) != 0 {
		return
	}
	o := Odc_Lit2Obj(p, l)
	if Odc_ObjIsTravIdCurrent(p, o) != 0 {
		return
	}
	Odc_ObjSetTravIdCurrent(p, o)
	Abc_NtkDontCareSimulate_rec(p, Odc_ObjFanin0(o))
	Abc_NtkDontCareSimulate_rec(p, Odc_ObjFanin1(o))
	Abc_NtkDontCareTruthOne(p, l)
}
func Abc_NtkDontCareSimulate(p *Odc_Man_t, d []uint32) int32 {
	Odc_ManIncrementTravId(p)
	l := Odc_Regular(p.iRoot)
	Abc_NtkDontCareSimulate_rec(p, l)
	copy(d, Odc_ObjTruth(p, l))
	if Odc_IsComplement(p.iRoot) != 0 {
		for i := range d[:p.nWords] {
			d[i] = ^d[i]
		}
	}
	return extra.Extra_TruthCountOnes(d, p.nVarsMax)
}
func Abc_NtkDontCareSimulateBefore(p *Odc_Man_t, d []uint32) int32 {
	const iters = 2
	rounds := p.nBits / p.nWords
	var c int32
	for i := 0; i < iters; i++ {
		Abc_NtkDontCareSimulateSetRand(p)
		Abc_NtkDontCareSimulate(p, d)
		c += Abc_NtkDontCareCountMintsWord(p, d)
	}
	return c * rounds / iters
}

func Abc_NtkDontCareCompute(p *Odc_Man_t, node *Abc_Obj_t, leaves *Vec_Ptr_t, d []uint32) int32 {
	start := util.Abc_Clock()
	p.nWins++
	assert(baseabc.Abc_ObjIsComplement(node) == 0 && baseabc.Abc_ObjIsNode(node) != 0 && int32(vec.Vec_PtrSize(leaves)) <= p.nVarsMax, "ODC compute arguments")
	p.vLeaves, p.pNode = leaves, node
	clk := util.Abc_Clock()
	ret := Abc_NtkDontCareWindow(p)
	p.timeWin += util.Abc_Clock() - clk
	if ret == 0 {
		p.timeAbort += util.Abc_Clock() - start
		res.Abc_InfoFill(d, p.nWords)
		p.nWinsEmpty++
		return 0
	}
	if p.fVeryVerbose != 0 {
		fmt.Printf(" %5d : Leaf = %2d Root = %2d Bran = %2d | ", node.ID, vec.Vec_PtrSize(leaves), vec.Vec_PtrSize(p.vRoots), vec.Vec_PtrSize(p.vBranches))
	}
	clk = util.Abc_Clock()
	Abc_NtkDontCareTransfer(p)
	p.timeMiter += util.Abc_Clock() - clk
	clk = util.Abc_Clock()
	nm := Abc_NtkDontCareSimulateBefore(p, d)
	p.timeSim += util.Abc_Clock() - clk
	if p.fVeryVerbose != 0 {
		fmt.Printf("AIG = %5d %6.2f %% ", Odc_NodeNum(p), 100*float64(p.nBits-nm)/float64(p.nBits))
	}
	if 100*float64(p.nBits-nm)/float64(p.nBits) < float64(p.nPercCutoff) {
		p.timeAbort += util.Abc_Clock() - start
		if p.fVeryVerbose != 0 {
			fmt.Println("Simulation cutoff.")
		}
		res.Abc_InfoFill(d, p.nWords)
		p.nSimsEmpty++
		return 0
	}
	clk = util.Abc_Clock()
	ret = Abc_NtkDontCareQuantify(p)
	p.timeQuant += util.Abc_Clock() - clk
	if ret == 0 {
		p.timeAbort += util.Abc_Clock() - start
		if p.fVeryVerbose != 0 {
			fmt.Println("=== Overflow! ===")
		}
		res.Abc_InfoFill(d, p.nWords)
		p.nQuantsOver++
		return 0
	}
	clk = util.Abc_Clock()
	Abc_NtkDontCareSimulateSetElem(p)
	nm = Abc_NtkDontCareSimulate(p, d)
	p.timeTruth += util.Abc_Clock() - clk
	if p.fVeryVerbose != 0 {
		fmt.Printf("AIG = %5d %6.2f %%\n", Odc_NodeNum(p), 100*float64(p.nBits-nm)/float64(p.nBits))
	}
	p.timeTotal += util.Abc_Clock() - start
	p.nWinsFinish++
	p.nTotalDcs += int32(100 * float64(p.nBits-nm) / float64(p.nBits))
	return nm
}
