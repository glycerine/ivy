package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Cec3 uses the same packed simulation and phase-normalized classes as Cec2.
func Cec3_ObjSim(p *gia.Gia_Man_t, id int32) []uint64        { return Cec2_ObjSim(p, id) }
func Cec3_ObjSimSetInputBit(p *gia.Gia_Man_t, id, bit int32) { Cec2_ObjSimSetInputBit(p, id, bit) }
func Cec3_ObjSimRo(p *gia.Gia_Man_t, id int32)               { Cec2_ObjSimRo(p, id) }
func Cec3_ObjSimCo(p *gia.Gia_Man_t, id int32)               { Cec2_ObjSimCo(p, id) }
func Cec3_ObjSimAnd(p *gia.Gia_Man_t, id int32)              { Cec2_ObjSimAnd(p, id) }
func Cec3_ObjSimEqual(p *gia.Gia_Man_t, a, b int32) int32    { return Cec2_ObjSimEqual(p, a, b) }
func Cec3_ObjSimCi(p *gia.Gia_Man_t, id int32)               { Cec2_ObjSimCi(p, id) }
func Cec3_ManSimulateCis(p *gia.Gia_Man_t)                   { Cec2_ManSimulateCis(p) }
func Cec3_ManDeriveCex(p *gia.Gia_Man_t, out, pattern int32) *util.Abc_Cex_t {
	return Cec2_ManDeriveCex(p, out, pattern)
}
func Cec3_ManSimulateCos(p *gia.Gia_Man_t) int32             { return Cec2_ManSimulateCos(p) }
func Cec3_ManSaveCis(p *gia.Gia_Man_t)                       { Cec2_ManSaveCis(p) }
func Cec3_ManSimAlloc(p *gia.Gia_Man_t, words int32)         { Cec2_ManSimAlloc(p, words) }
func Cec3_ManSimClassRefineOne(p *gia.Gia_Man_t, repr int32) { Cec2_ManSimClassRefineOne(p, repr) }
func Cec3_ManCreateClasses(p *gia.Gia_Man_t, man *Cec3_Man_t) {
	var cec2 Cec2_Man_t
	Cec2_ManCreateClasses(p, &cec2)
	man.timeRefine += cec2.timeRefine
}
func Cec3_ManSimulate(p *gia.Gia_Man_t, triples *vec.Vec_Int_t, man *Cec3_Man_t) int32 {
	var cec2 Cec2_Man_t
	result := Cec2_ManSimulate(p, triples, &cec2)
	man.timeSim += cec2.timeSim
	man.timeRefine += cec2.timeRefine
	return result
}
