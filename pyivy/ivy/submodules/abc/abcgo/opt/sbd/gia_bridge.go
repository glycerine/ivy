package sbd

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

type Gia_Man_t = gia.Gia_Man_t
type Gia_Obj_t = gia.Gia_Obj_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Wec_t = vec.Vec_Wec_t
type Vec_Bit_t = vec.Vec_Bit_t

const ABC_INFINITY = 1000000000

func Gia_ManObj(p *Gia_Man_t, i int32) *Gia_Obj_t       { return gia.Gia_ManObj(p, i) }
func Gia_ManObjNum(p *Gia_Man_t) int32                  { return gia.Gia_ManObjNum(p) }
func Gia_ObjId(p *Gia_Man_t, o *Gia_Obj_t) int32        { return gia.Gia_ObjId(p, o) }
func Gia_ObjFaninId0(o *Gia_Obj_t, id int32) int32      { return gia.Gia_ObjFaninId0(o, id) }
func Gia_ObjFaninId1(o *Gia_Obj_t, id int32) int32      { return gia.Gia_ObjFaninId1(o, id) }
func Gia_ObjFaninC0(o *Gia_Obj_t) int32                 { return gia.Gia_ObjFaninC0(o) }
func Gia_ObjFaninC1(o *Gia_Obj_t) int32                 { return gia.Gia_ObjFaninC1(o) }
func Gia_ObjFanin0(o *Gia_Obj_t) *Gia_Obj_t             { return gia.Gia_ObjFanin0(o) }
func Gia_ObjFanin1(o *Gia_Obj_t) *Gia_Obj_t             { return gia.Gia_ObjFanin1(o) }
func Gia_ObjFaninId0p(p *Gia_Man_t, o *Gia_Obj_t) int32 { return gia.Gia_ObjFaninId0p(p, o) }
func Gia_ObjIsAnd(o *Gia_Obj_t) bool                    { return gia.Gia_ObjIsAnd(o) != 0 }
func Gia_ObjIsXor(o *Gia_Obj_t) bool                    { return gia.Gia_ObjIsXor(o) != 0 }
func Gia_ObjIsCi(o *Gia_Obj_t) bool                     { return gia.Gia_ObjIsCi(o) != 0 }
func Gia_ObjIsCo(o *Gia_Obj_t) bool                     { return gia.Gia_ObjIsCo(o) != 0 }
func Gia_ObjIsBuf(o *Gia_Obj_t) bool                    { return gia.Gia_ObjIsBuf(o) != 0 }
func Gia_ObjIsConst0(o *Gia_Obj_t) bool                 { return gia.Gia_ObjIsConst0(o) != 0 }
func Gia_ObjIsMux(p *Gia_Man_t, o *Gia_Obj_t) bool      { return gia.Gia_ObjIsMux(p, o) != 0 }
func Gia_ObjMark0(o *Gia_Obj_t) bool                    { return gia.Gia_ObjMark0(o) != 0 }
func Gia_ObjMark1(o *Gia_Obj_t) bool                    { return gia.Gia_ObjMark1(o) != 0 }
func Gia_ObjSetMark0(o *Gia_Obj_t, v int32)             { gia.Gia_ObjSetMark0(o, v) }
func Gia_ObjSetMark1(o *Gia_Obj_t, v int32)             { gia.Gia_ObjSetMark1(o, v) }
func Gia_ObjValue(o *Gia_Obj_t) int32                   { return gia.Gia_ObjValue(o) }
func Gia_ObjSetValue(o *Gia_Obj_t, v int32)             { gia.Gia_ObjSetValue(o, v) }
func Gia_ObjCioId(o *Gia_Obj_t) int32                   { return gia.Gia_ObjCioId(o) }
func Gia_ObjLevelId(p *Gia_Man_t, id int32) int32       { return gia.Gia_ObjLevelId(p, id) }
func Gia_ObjRefNumId(p *Gia_Man_t, id int32) int32      { return gia.Gia_ObjRefNumId(p, id) }
func Gia_ManIncrementTravId(p *Gia_Man_t)               { gia.Gia_ManIncrementTravId(p) }
func Gia_ObjIsTravIdCurrentId(p *Gia_Man_t, id int32) bool {
	return gia.Gia_ObjIsTravIdCurrentId(p, id) != 0
}
func Gia_ObjSetTravIdCurrentId(p *Gia_Man_t, id int32) { gia.Gia_ObjSetTravIdCurrentId(p, id) }
func Gia_ManRandomW(reset int32) uint64                { return gia.Gia_ManRandomW(reset) }
func Gia_ManCiNum(p *Gia_Man_t) int32                  { return gia.Gia_ManCiNum(p) }
func Gia_ManCoNum(p *Gia_Man_t) int32                  { return gia.Gia_ManCoNum(p) }
func Gia_ManCi(p *Gia_Man_t, i int32) *Gia_Obj_t       { return gia.Gia_ManCi(p, i) }
func Gia_ManCo(p *Gia_Man_t, i int32) *Gia_Obj_t       { return gia.Gia_ManCo(p, i) }
func Gia_ManCiIdToId(p *Gia_Man_t, i int32) int32      { return gia.Gia_ManCiIdToId(p, i) }

// Gia_ManForEachCoDriverId
func giaCoDriverIds(p *Gia_Man_t) []int32 {
	n := Gia_ManCoNum(p)
	res := make([]int32, n)
	for i := int32(0); i < n; i++ {
		res[i] = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))
	}
	return res
}

func gia_ManAndNum(p *Gia_Man_t) int32 { return gia.Gia_ManAndNum(p) }
