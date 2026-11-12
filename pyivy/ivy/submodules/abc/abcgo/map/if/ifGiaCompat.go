package if_

import "github.com/glycerine/abcgo/aig/gia"

// These adapters preserve the source package's GIA vocabulary while routing
// the algorithms through the owning GIA package's public operations.
func Gia_ManStart(n int32) *Gia_Man_t                  { return gia.Gia_ManStart(n) }
func Gia_ManStop(p *Gia_Man_t)                         { gia.Gia_ManStop(p) }
func Gia_ManCleanup(p *Gia_Man_t) *Gia_Man_t           { return gia.Gia_ManCleanup(p) }
func Gia_ManObjNum(p *Gia_Man_t) int32                 { return gia.Gia_ManObjNum(p) }
func Gia_ManCiNum(p *Gia_Man_t) int32                  { return gia.Gia_ManCiNum(p) }
func Gia_ManCoNum(p *Gia_Man_t) int32                  { return gia.Gia_ManCoNum(p) }
func Gia_ManPiNum(p *Gia_Man_t) int32                  { return gia.Gia_ManPiNum(p) }
func Gia_ManPoNum(p *Gia_Man_t) int32                  { return gia.Gia_ManPoNum(p) }
func Gia_ManRegNum(p *Gia_Man_t) int32                 { return gia.Gia_ManRegNum(p) }
func Gia_ManSetRegNum(p *Gia_Man_t, n int32)           { gia.Gia_ManSetRegNum(p, n) }
func Gia_ManCi(p *Gia_Man_t, i int32) *Gia_Obj_t       { return gia.Gia_ManCi(p, i) }
func Gia_ManCo(p *Gia_Man_t, i int32) *Gia_Obj_t       { return gia.Gia_ManCo(p, i) }
func Gia_ManPo(p *Gia_Man_t, i int32) *Gia_Obj_t       { return gia.Gia_ManPo(p, i) }
func Gia_ManObj(p *Gia_Man_t, i int32) *Gia_Obj_t      { return gia.Gia_ManObj(p, i) }
func Gia_ManConst0(p *Gia_Man_t) *Gia_Obj_t            { return gia.Gia_ManConst0(p) }
func Gia_ManAppendCi(p *Gia_Man_t) int32               { return gia.Gia_ManAppendCi(p) }
func Gia_ManAppendCo(p *Gia_Man_t, lit int32) int32    { return gia.Gia_ManAppendCo(p, lit) }
func Gia_ManAppendBuf(p *Gia_Man_t, lit int32) int32   { return gia.Gia_ManAppendBuf(p, lit) }
func Gia_ManHashAlloc(p *Gia_Man_t)                    { gia.Gia_ManHashAlloc(p) }
func Gia_ManHashStart(p *Gia_Man_t)                    { gia.Gia_ManHashStart(p) }
func Gia_ManHashStop(p *Gia_Man_t)                     { gia.Gia_ManHashStop(p) }
func Gia_ManHashAnd(p *Gia_Man_t, a, b int32) int32    { return gia.Gia_ManHashAnd(p, a, b) }
func Gia_ManHashXor(p *Gia_Man_t, a, b int32) int32    { return gia.Gia_ManHashXor(p, a, b) }
func Gia_ManHashMux(p *Gia_Man_t, c, t, e int32) int32 { return gia.Gia_ManHashMux(p, c, t, e) }
func Gia_ManFillValue(p *Gia_Man_t)                    { gia.Gia_ManFillValue(p) }
func Gia_ManHasMapping(p *Gia_Man_t) int32             { return gia.Gia_ManHasMapping(p) }
func Gia_ObjIsAnd(p *Gia_Obj_t) int32                  { return gia.Gia_ObjIsAnd(p) }
func Gia_ObjIsBuf(p *Gia_Obj_t) int32                  { return gia.Gia_ObjIsBuf(p) }
func Gia_ObjIsLut(p *Gia_Man_t, id int32) int32        { return gia.Gia_ObjIsLut(p, id) }
func Gia_ObjLutFanins(p *Gia_Man_t, id int32) []int32  { return gia.Gia_ObjLutFanins(p, id) }
func Gia_ObjFanin0Copy(p *Gia_Obj_t) int32             { return gia.Gia_ObjFanin0Copy(p) }
func Gia_ObjFanin1Copy(p *Gia_Obj_t) int32             { return gia.Gia_ObjFanin1Copy(p) }
func Gia_ObjId(p *Gia_Man_t, obj *Gia_Obj_t) int32     { return gia.Gia_ObjId(p, obj) }
