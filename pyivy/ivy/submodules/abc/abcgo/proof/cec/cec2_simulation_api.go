package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"time"
)

func cec2GraphSimulation(p *gia.Gia_Man_t) *cec3Simulation {
	return &cec3Simulation{words: int(p.SimulationWordCount()), data: gia.Gia_ManCecSimWords(p)}
}
func Cec2_ObjSim(p *gia.Gia_Man_t, id int32) []uint64 { return cec2GraphSimulation(p).object(id) }
func Cec2_ObjSimSetInputBit(p *gia.Gia_Man_t, id, value int32) {
	cec2GraphSimulation(p).setInputBit(id, int(gia.Gia_ManCecSimPattern(p)), value != 0)
}
func Cec2_ObjSimRo(p *gia.Gia_Man_t, id int32) {
	copy(Cec2_ObjSim(p, id), Cec2_ObjSim(p, gia.Gia_ObjRoToRiId(p, id)))
}
func Cec2_ObjSimCo(p *gia.Gia_Man_t, id int32)  { cec2GraphSimulation(p).simulateOutput(p, id) }
func Cec2_ObjSimAnd(p *gia.Gia_Man_t, id int32) { cec2GraphSimulation(p).simulateAnd(p, id) }
func Cec2_ObjSimEqual(p *gia.Gia_Man_t, a, b int32) int32 {
	if cec2GraphSimulation(p).equivalent(a, b) {
		return 1
	}
	return 0
}
func Cec2_ObjSimCi(p *gia.Gia_Man_t, id int32) {
	words := Cec2_ObjSim(p, id)
	for i := range words {
		words[i] = gia.Gia_ManRandomW(0)
	}
	words[0] <<= 1
}
func Cec2_ManSimulateCis(p *gia.Gia_Man_t) {
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		Cec2_ObjSimCi(p, gia.Gia_ManCiIdToId(p, i))
	}
	gia.Gia_ManCecSetSimPattern(p, 0)
}
func Cec2_ManDeriveCex(p *gia.Gia_Man_t, out, pattern int32) *util.Abc_Cex_t {
	cex := util.Abc_CexAlloc(0, gia.Gia_ManCiNum(p), 1)
	cex.IPo = out
	if pattern < 0 {
		return cex
	}
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		words := Cec2_ObjSim(p, gia.Gia_ManCiIdToId(p, i))
		if words[pattern/64]&(uint64(1)<<uint(pattern%64)) != 0 {
			util.Abc_InfoSetBit(cex.PData, i)
		}
	}
	return cex
}
func Cec2_ManSimulateCos(p *gia.Gia_Man_t) int32 {
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		id := gia.Gia_ObjId(p, gia.Gia_ManCo(p, i))
		Cec2_ObjSimCo(p, id)
		for w, word := range Cec2_ObjSim(p, id) {
			if word != 0 {
				cex := Cec2_ManDeriveCex(p, i, int32(w*64+bits.TrailingZeros64(word)))
				if err := p.SetCecCombinationalWitness(cex); err != nil {
					panic(err)
				}
				return 0
			}
		}
	}
	return 1
}
func Cec2_ManSaveCis(p *gia.Gia_Man_t) {
	for w := 0; w < int(p.SimulationWordCount()); w++ {
		for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
			gia.Gia_ManCecSaveCIWord(p, Cec2_ObjSim(p, gia.Gia_ManCiIdToId(p, i))[w])
		}
	}
}
func Cec2_ManSimAlloc(p *gia.Gia_Man_t, words int32) { gia.Gia_ManCecSimAlloc(p, words) }
func Cec2_ManSimHashKey(sim []uint64, words, tableSize int32) int32 {
	return Cec3_ManSimHashKey(sim, words, tableSize)
}
func Cec2_ManSimClassRefineOne(p *gia.Gia_Man_t, repr int32) {
	cec2GraphSimulation(p).refineClass(p, repr)
}
func Cec2_ManCreateClasses(p *gia.Gia_Man_t, man *Cec2_Man_t) {
	start := time.Now()
	cec2GraphSimulation(p).createClasses(p)
	man.timeRefine += time.Since(start).Nanoseconds()
}
func Cec2_ManSimulate(p *gia.Gia_Man_t, triples *vec.Vec_Int_t, man *Cec2_Man_t) int32 {
	start := time.Now()
	sim := cec2GraphSimulation(p)
	for id := int32(1); int(id) < p.ObjectCount(); id++ {
		_, _, _, kind := p.Gate(id)
		if kind == 2 {
			sim.simulateAnd(p, id)
		}
	}
	man.timeSim += time.Since(start).Nanoseconds()
	if !p.HasRepresentatives() {
		return 0
	}
	var count int32
	if triples != nil {
		for i := 0; i+2 < vec.Vec_IntSize(triples); i += 3 {
			a, b, entry := vec.Vec_IntEntry(triples, i), vec.Vec_IntEntry(triples, i+1), vec.Vec_IntEntry(triples, i+2)
			pattern := entry >> 1
			x := (sim.object(a)[pattern/64] >> uint(pattern%64)) & 1
			y := (sim.object(b)[pattern/64] >> uint(pattern%64)) & 1
			if (x ^ uint64(entry&1)) == y {
				count++
			}
		}
	}
	start = time.Now()
	for id := int32(0); int(id) < p.ObjectCount(); id++ {
		if gia.Gia_ObjIsHead(p, id) != 0 {
			sim.refineClass(p, id)
		}
	}
	man.timeRefine += time.Since(start).Nanoseconds()
	return count
}
