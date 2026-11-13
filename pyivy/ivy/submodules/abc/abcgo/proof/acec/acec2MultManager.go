package acec

// Incremental cut computation from src/proof/acec/acec2Mult.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Sdb_StoAlloc(g *Gia_Man_t, size, num, min, truthMin, verbose int32) *Sdb_Sto_t {
	if size >= SDB_CUT_NO_LEAF || size <= 1 || size > SDB_MAX_CUTSIZE || num <= 1 || num >= SDB_MAX_CUTNUM {
		panic("acec: invalid cut parameters")
	}
	p := &Sdb_Sto_t{clkStart: time.Now(), nCutSize: size, nCutNum: num, fCutMin: min, fTruthMin: truthMin, fVerbose: verbose, pGia: g, vRefs: vec.Vec_IntAlloc(int(gia.Gia_ManObjNum(g))), vCuts: vec.Vec_WecStart(int(gia.Gia_ManObjNum(g)))}
	if min != 0 {
		p.vTtMem = vec.Vec_MemAllocForTT(int(size), 0)
	}
	return p
}
func Sdb_StoFree(p *Sdb_Sto_t) {
	vec.Vec_IntFree(p.vRefs)
	vec.Vec_WecFree(p.vCuts)
	if p.fCutMin != 0 {
		vec.Vec_MemHashFree(p.vTtMem)
		vec.Vec_MemFree(p.vTtMem)
	}
}
func Sdb_StoComputeCutsConst0(p *Sdb_Sto_t, id int32) { Sdb_CutAddZero(p, id) }
func Sdb_StoComputeCutsCi(p *Sdb_Sto_t, id int32)     { Sdb_CutAddUnit(p, id) }
func Sdb_StoComputeCutsNode(p *Sdb_Sto_t, id int32)   { Sdb_StoMergeCuts(p, id) }
func Sdb_StoRefObj(p *Sdb_Sto_t, id int32) {
	o := gia.Gia_ManObj(p.pGia, id)
	if int(id) != vec.Vec_IntSize(p.vRefs) {
		panic("acec: out-of-order reference preparation")
	}
	vec.Vec_IntPush(p.vRefs, 0)
	if gia.Gia_ObjIsAnd(o) != 0 {
		vec.Vec_IntAddToEntry(p.vRefs, int(gia.Gia_ObjFaninId0(o, id)), 1)
		vec.Vec_IntAddToEntry(p.vRefs, int(gia.Gia_ObjFaninId1(o, id)), 1)
	} else if gia.Gia_ObjIsCo(o) != 0 {
		vec.Vec_IntAddToEntry(p.vRefs, int(gia.Gia_ObjFaninId0(o, id)), 1)
	}
}
func Sdb_StoComputeCutsDetect(g *Gia_Man_t) *Vec_Int_t {

	p := Sdb_StoAlloc(g, 5, 20, 1, 0, 1)
	for id := int32(0); id < gia.Gia_ManObjNum(g); id++ {
		Sdb_StoRefObj(p, id)
	}
	Sdb_StoComputeCutsConst0(p, 0)
	for i := int32(0); i < gia.Gia_ManCiNum(g); i++ {
		Sdb_StoComputeCutsCi(p, gia.Gia_ObjId(g, gia.Gia_ManCi(g, i)))
	}
	for id := int32(1); id < gia.Gia_ManObjNum(g); id++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(g, id)) != 0 {
			Sdb_StoComputeCutsNode(p, id)
		}
	}
	if p.fVerbose != 0 {
		fmt.Printf("Running cut computation with CutSize = %d  CutNum = %d:\n", p.nCutSize, p.nCutNum)
		fmt.Printf("CutPair = %.0f  Merge = %.0f (%.2f %%)  Eval = %.0f (%.2f %%)  Cut = %.0f (%.2f %%)  Cut/Node = %.2f\nOver = %4d  Time = %v\n", p.CutCount[0], p.CutCount[1], 100*p.CutCount[1]/p.CutCount[0], p.CutCount[2], 100*p.CutCount[2]/p.CutCount[0], p.CutCount[3], 100*p.CutCount[3]/p.CutCount[0], p.CutCount[3]/float64(gia.Gia_ManAndNum(g)), p.nCutsOver, time.Since(p.clkStart))
	}
	result := Sdb_StoIterCuts(p)
	Sdb_StoFree(p)
	return result
}
func Sdb_StoComputeCutsTest(g *Gia_Man_t) { vec.Vec_IntFree(Sdb_StoComputeCutsDetect(g)) }
