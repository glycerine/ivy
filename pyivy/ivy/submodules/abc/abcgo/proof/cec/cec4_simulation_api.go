package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

type cec4SimulationState struct {
	sim         *cec3Simulation
	initialized bool
}

func Cec4_ManSimAlloc(p *Cec4_Man_t, words int32) {
	if words <= 0 {
		panic("cec4: invalid simulation word count")
	}
	sim := newCec3Simulation(p.pAig.ObjectCount(), int(words))
	p.simState = &cec4SimulationState{sim: sim}
	p.packer = &cec4PatternPacker{values: sim, reserved: newCec3Simulation(p.pAig.ObjectCount(), int(words))}
	p.pattern = 0
}
func Cec4_ObjSim(p *Cec4_Man_t, id int32) []uint64 { return p.simState.sim.object(id) }
func Cec4_ObjSimEqual(p *Cec4_Man_t, a, b int32) int32 {
	if p.simState.sim.equivalent(a, b) {
		return 1
	}
	return 0
}
func Cec4_RefineOneClassIter(p *Cec4_Man_t, repr int32) {
	cec4RefineClass(p.simState.sim, p.pAig, repr)
}
func Cec4_RefineOneClass(p *Cec4_Man_t, nodes *vec.Vec_Int_t) {
	g := p.pAig
	entries := vec.Vec_IntArray(nodes)
	bins := make(map[int32]int32)
	var order []int32
	for i := len(entries) - 1; i >= 0; i-- {
		id := entries[i]
		key := Cec4_ManSimHashKey(Cec4_ObjSim(p, id), int32(p.simState.sim.words), p.nTableSize)
		head, ok := bins[key]
		if !ok {
			head = -1
			order = append(order, key)
		}
		gia.Gia_ObjSetNext(g, id, head)
		bins[key] = id
	}
	for _, key := range order {
		head := bins[key]
		for id := gia.Gia_ObjNext(g, head); id > 0; id = gia.Gia_ObjNext(g, id) {
			gia.Gia_ObjSetRepr(g, id, head)
		}
		if gia.Gia_ObjNext(g, head) > 0 {
			Cec4_RefineOneClassIter(p, head)
		}
	}
	if p.vRefBins != nil {
		vec.Vec_IntClear(p.vRefBins)
	}
}
func Cec4_RefineClasses(p *Cec4_Man_t, _ *vec.Vec_Int_t) {
	if vec.Vec_IntSize(p.vRefClasses) == 0 {
		return
	}
	g := p.pAig
	if vec.Vec_IntSize(p.vRefNodes) > 0 {
		Cec4_RefineOneClass(p, p.vRefNodes)
	} else {
		for _, head := range vec.Vec_IntArray(p.vRefClasses) {
			gia.Gia_ObjUnsetColor(g, head, 0)
			vec.Vec_IntClear(p.vRefNodes)
			vec.Vec_IntPush(p.vRefNodes, head)
			for id := gia.Gia_ObjNext(g, head); id > 0; id = gia.Gia_ObjNext(g, id) {
				vec.Vec_IntPush(p.vRefNodes, id)
			}
			for _, id := range vec.Vec_IntArray(p.vRefNodes) {
				gia.Gia_ObjUnsetRepr(g, id)
				gia.Gia_ObjSetNext(g, id, -1)
			}
			Cec4_RefineOneClass(p, p.vRefNodes)
		}
	}
	vec.Vec_IntClear(p.vRefClasses)
	vec.Vec_IntClear(p.vRefNodes)
}
func Cec4_RefineInit(p *Cec4_Man_t) {
	g := p.pAig
	g.StartRepresentatives()
	count := gia.Gia_ManObjNum(g)
	p.nTableSize = util.Abc_PrimeCudd(uint32(count))
	p.vRefNodes = vec.Vec_IntAlloc(int(count))
	p.vRefBins = vec.Vec_IntAlloc(int(count) / 2)
	p.vRefClasses = vec.Vec_IntAlloc(int(count) / 2)
	for id := int32(0); id < count; id++ {
		gia.Gia_ObjSetNext(g, id, -1)
		obj := gia.Gia_ManObj(g, id)
		if gia.Gia_ObjIsCo(obj) == 0 && (p.pPars.NLevelMax == 0 || gia.Gia_ObjLevel(g, obj) <= p.pPars.NLevelMax) {
			vec.Vec_IntPush(p.vRefNodes, id)
		}
	}
	vec.Vec_IntPush(p.vRefClasses, 0)
	p.simState.initialized = true
}
func Cec4_ObjSimSetInputBit(p *Cec4_Man_t, id, value int32) {
	p.simState.sim.setInputBit(id, p.pattern, value != 0)
}
func Cec4_ObjSimGetInputBit(p *Cec4_Man_t, id int32) int32 {
	return int32((Cec4_ObjSim(p, id)[p.pattern/64] >> uint(p.pattern%64)) & 1)
}
func Cec4_ObjSimRo(p *Cec4_Man_t, id int32) {
	copy(Cec4_ObjSim(p, id), Cec4_ObjSim(p, gia.Gia_ObjRoToRiId(p.pAig, id)))
}
func Cec4_ObjSimCo(p *Cec4_Man_t, id int32)  { cec4SimulateNode(p.simState.sim, p.pAig, id) }
func Cec4_ObjSimAnd(p *Cec4_Man_t, id int32) { cec4SimulateNode(p.simState.sim, p.pAig, id) }
func Cec4_ObjSimXor(p *Cec4_Man_t, id int32) { cec4SimulateNode(p.simState.sim, p.pAig, id) }
func Cec4_ObjSimCi(p *Cec4_Man_t, id int32) {
	words := Cec4_ObjSim(p, id)
	for i := range words {
		words[i] = gia.Gia_ManRandomW(0)
	}
	words[0] <<= 1
}
func Cec4_ObjClearSimCi(p *Cec4_Man_t, id int32) { clear(Cec4_ObjSim(p, id)) }
func Cec4_ManSimulateCis(p *Cec4_Man_t) {
	for i := int32(0); i < gia.Gia_ManCiNum(p.pAig); i++ {
		Cec4_ObjSimCi(p, gia.Gia_ManCiIdToId(p.pAig, i))
	}
	p.pattern = 0
}
func Cec4_ManClearCis(p *Cec4_Man_t) {
	for i := int32(0); i < gia.Gia_ManCiNum(p.pAig); i++ {
		Cec4_ObjClearSimCi(p, gia.Gia_ManCiIdToId(p.pAig, i))
	}
	p.pattern = 0
}
func Cec4_ManDeriveCex(p *Cec4_Man_t, out, pattern int32) *util.Abc_Cex_t {
	cex := util.Abc_CexAlloc(0, gia.Gia_ManCiNum(p.pAig), 1)
	cex.IPo = out
	if pattern < 0 {
		return cex
	}
	for i := int32(0); i < gia.Gia_ManCiNum(p.pAig); i++ {
		id := gia.Gia_ManCiIdToId(p.pAig, i)
		if Cec4_ObjSim(p, id)[pattern/64]&(uint64(1)<<uint(pattern%64)) != 0 {
			util.Abc_InfoSetBit(cex.PData, i)
		}
	}
	return cex
}
func Cec4_ManSimulateCos(p *Cec4_Man_t) int32 {
	g := p.pAig
	for i := int32(0); i < gia.Gia_ManCoNum(g); i++ {
		id := gia.Gia_ObjId(g, gia.Gia_ManCo(g, i))
		Cec4_ObjSimCo(p, id)
		if Cec4_ObjSimEqual(p, id, 0) != 0 {
			continue
		}
		for w, v := range Cec4_ObjSim(p, id) {
			if v != 0 {
				cex := Cec4_ManDeriveCex(p, i, int32(w*64+bits.TrailingZeros64(v)))
				if err := g.SetCecCombinationalWitness(cex); err != nil {
					panic(err)
				}
				return 0
			}
		}
	}
	return 1
}
func Cec4_ManSimulate(p *Cec4_Man_t) {
	unpin := cecPinClockThread()
	defer unpin()
	start := cecCPUClock()
	p.nSimulates++
	if !p.simState.initialized {
		Cec4_RefineInit(p)
	}
	g := p.pAig
	for id := int32(1); id < gia.Gia_ManObjNum(g); id++ {
		_, _, _, kind := g.Gate(id)
		if kind < 2 || kind == 6 {
			continue
		}
		repr := gia.Gia_ObjRepr(g, id)
		cec4SimulateNode(p.simState.sim, g, id)
		if repr == gia.GIA_VOID || gia.Gia_ObjColor(g, repr, 0) != 0 || Cec4_ObjSimEqual(p, repr, id) != 0 {
			continue
		}
		gia.Gia_ObjSetColor(g, repr, 0)
		vec.Vec_IntPush(p.vRefClasses, repr)
	}
	p.timeSim += (cecCPUClock() - start)
	start = cecCPUClock()
	Cec4_RefineClasses(p, p.vRefClasses)
	p.timeRefine += (cecCPUClock() - start)
}
func Cec4_ManSimulate_rec(p *Cec4_Man_t, id int32) {
	if id == 0 || vec.Vec_IntEntry(p.vCexStamps, int(id)) == int32(p.pattern) {
		return
	}
	vec.Vec_IntWriteEntry(p.vCexStamps, int(id), int32(p.pattern))
	a, b, c, kind := p.pAig.Gate(id)
	if kind == 1 {
		return
	}
	if kind < 2 {
		panic("cec4: invalid simulation cone")
	}
	Cec4_ManSimulate_rec(p, a>>1)
	if kind >= 2 && kind <= 4 {
		Cec4_ManSimulate_rec(p, b>>1)
	}
	if kind == 4 {
		Cec4_ManSimulate_rec(p, c>>1)
	}
	cec4SimulateNode(p.simState.sim, p.pAig, id)
}
