package wlc

import (
	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	abc "github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	pdr "github.com/glycerine/abcgo/proof/pdr"
)

type Vec_Int_t = vec.Vec_Int_t
type Vec_Str_t = vec.Vec_Str_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Wrd_t = vec.Vec_Wrd_t
type Vec_Wec_t = vec.Vec_Wec_t
type Vec_Bit_t = vec.Vec_Bit_t
type Vec_Vec_t = vec.Vec_Vec_t
type Vec_Mem_t = vec.Vec_Mem_t
type Abc_Cex_t = util.Abc_Cex_t
type Gia_Man_t = gia.Gia_Man_t
type Gia_Obj_t = gia.Gia_Obj_t
type Mem_Flex_t = mem.Mem_Flex_t
type Abc_Frame_t = abcmain.Frame
type Abc_Ntk_t = abc.Network
type Abc_Obj_t = abc.Object
type Aig_Man_t = aig.Graph
type abctime = int64
type word = uint64

const (
	ABC_NTK_LOGIC = abc.ABC_NTK_LOGIC
	ABC_FUNC_SOP  = abc.ABC_FUNC_SOP
)

var (
	Pdr_InvCounts   = pdr.Pdr_InvCounts
	Pdr_InvPrintStr = pdr.Pdr_InvPrintStr
)

func Abc_NtkAlloc(kind abc.Abc_NtkType_t, fun abc.Abc_NtkFunc_t, memory int32) *Abc_Ntk_t {
	return abc.Abc_NtkAlloc(kind, fun, memory)
}
func Abc_NtkDelete(n *Abc_Ntk_t)                    { abc.Abc_NtkDelete(n) }
func Abc_NtkCreatePi(n *Abc_Ntk_t) *Abc_Obj_t       { return abc.Abc_NtkCreatePi(n) }
func Abc_NtkCreatePo(n *Abc_Ntk_t) *Abc_Obj_t       { return abc.Abc_NtkCreatePo(n) }
func Abc_NtkCreateNode(n *Abc_Ntk_t) *Abc_Obj_t     { return abc.Abc_NtkCreateNode(n) }
func Abc_NtkPiNum(n *Abc_Ntk_t) int32               { return abc.Abc_NtkPiNum(n) }
func Abc_NtkPoNum(n *Abc_Ntk_t) int32               { return abc.Abc_NtkPoNum(n) }
func Abc_NtkNodeNum(n *Abc_Ntk_t) int32             { return abc.Abc_NtkNodeNum(n) }
func Abc_NtkCo(n *Abc_Ntk_t, i int32) *Abc_Obj_t    { return abc.Abc_NtkCo(n, i) }
func Abc_ObjFanin0(o *Abc_Obj_t) *Abc_Obj_t         { return abc.Abc_ObjFanin0(o) }
func Abc_ObjFanin(o *Abc_Obj_t, i int32) *Abc_Obj_t { return abc.Abc_ObjFanin(o, i) }
func Abc_ObjFaninNum(o *Abc_Obj_t) int32            { return abc.Abc_ObjFaninNum(o) }
func Abc_ObjIsCi(o *Abc_Obj_t) int32                { return abc.Abc_ObjIsCi(o) }
func Abc_ObjName(o *Abc_Obj_t) string               { return abc.Abc_ObjName(o) }
func Abc_ObjAssignName(o *Abc_Obj_t, name, suffix string) string {
	return abc.Abc_ObjAssignName(o, name, suffix)
}
func Abc_ObjAddFanin(o *Abc_Obj_t, fanin any)        { abc.Abc_ObjAddFanin(o, fanin) }
func Abc_SopGetVarNum(s string) int32                { return abc.Abc_SopGetVarNum(s) }
func Abc_SopGetCubeNum(s string) int32               { return abc.Abc_SopGetCubeNum(s) }
func Abc_SopRegister(m *Mem_Flex_t, s string) string { return abc.Abc_SopRegister(m, s) }

var (
	Gia_ManObj      = gia.Gia_ManObj
	Gia_ManObjNum   = gia.Gia_ManObjNum
	Gia_ManRandomW  = gia.Gia_ManRandomW
	Gia_ManStop     = gia.Gia_ManStop
	Gia_ObjFaninC0  = gia.Gia_ObjFaninC0
	Gia_ObjFaninC1  = gia.Gia_ObjFaninC1
	Gia_ObjFaninId0 = gia.Gia_ObjFaninId0
	Gia_ObjFaninId1 = gia.Gia_ObjFaninId1
	Gia_ObjIsAnd    = gia.Gia_ObjIsAnd
	Gia_ObjIsCo     = gia.Gia_ObjIsCo
	Gia_ObjIsPi     = gia.Gia_ObjIsPi
	Gia_ObjIsRo     = gia.Gia_ObjIsRo
	Gia_ObjRoToRiId = gia.Gia_ObjRoToRiId
)

func Vec_IntSize(v *Vec_Int_t) int32                    { return int32(vec.Vec_IntSize(v)) }
func Vec_IntCap(v *Vec_Int_t) int32                     { return int32(vec.Vec_IntCap(v)) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32          { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntEntryLast(v *Vec_Int_t) int32               { return vec.Vec_IntEntryLast(v) }
func Vec_IntPush(v *Vec_Int_t, x int32)                 { vec.Vec_IntPush(v, x) }
func Vec_IntPushTwo(v *Vec_Int_t, x, y int32)           { vec.Vec_IntPushTwo(v, x, y) }
func Vec_IntPop(v *Vec_Int_t) int32                     { return vec.Vec_IntPop(v) }
func Vec_IntGrow(v *Vec_Int_t, n int32)                 { vec.Vec_IntGrow(v, int(n)) }
func Vec_IntErase(v *Vec_Int_t)                         { vec.Vec_IntErase(v) }
func Vec_IntFree(v *Vec_Int_t)                          { vec.Vec_IntFree(v) }
func Vec_IntClear(v *Vec_Int_t)                         { vec.Vec_IntClear(v) }
func Vec_IntFill(v *Vec_Int_t, n, value int32)          { vec.Vec_IntFill(v, int(n), value) }
func Vec_IntWriteEntry(v *Vec_Int_t, i, x int32)        { vec.Vec_IntWriteEntry(v, int(i), x) }
func Vec_IntArray(v *Vec_Int_t) []int32                 { return vec.Vec_IntArray(v) }
func Vec_IntStart(n int32) *Vec_Int_t                   { return vec.Vec_IntStart(int(n)) }
func Vec_IntStartNatural(n int32) *Vec_Int_t            { return vec.Vec_IntStartNatural(int(n)) }
func Vec_IntAppend(a, b *Vec_Int_t)                     { vec.Vec_IntAppend(a, b) }
func Vec_IntFind(a *Vec_Int_t, x int32) int32           { return vec.Vec_IntFind(a, x) }
func Vec_IntPushUnique(a *Vec_Int_t, x int32) int32     { return vec.Vec_IntPushUnique(a, x) }
func Vec_IntPushArray(a *Vec_Int_t, p []int32, n int32) { vec.Vec_IntPushArray(a, p, int(n)) }
func Vec_StrArray(v *Vec_Str_t) []byte                  { return vec.Vec_StrArray(v) }
func Vec_StrSize(v *Vec_Str_t) int                      { return vec.Vec_StrSize(v) }
func Vec_StrPush(v *Vec_Str_t, c byte)                  { vec.Vec_StrPush(v, c) }
func Vec_StrFree(v *Vec_Str_t)                          { vec.Vec_StrFree(v) }
func Vec_StrFreeP(v **Vec_Str_t)                        { vec.Vec_StrFreeP(v) }
func Vec_PtrAlloc(n int32) *Vec_Ptr_t                   { return vec.Vec_PtrAlloc(int(n)) }
func Vec_PtrPush(v *Vec_Ptr_t, x any)                   { vec.Vec_PtrPush(v, x) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any            { return vec.Vec_PtrEntry(v, int(i)) }
func Vec_PtrSize(v *Vec_Ptr_t) int32                    { return int32(vec.Vec_PtrSize(v)) }
func Vec_PtrFree(v *Vec_Ptr_t)                          { vec.Vec_PtrFree(v) }
func Vec_WecSize(v *Vec_Wec_t) int32                    { return int32(vec.Vec_WecSize(v)) }
func Vec_WecCap(v *Vec_Wec_t) int32                     { return int32(vec.Vec_WecCap(v)) }
func Vec_WecEntry(v *Vec_Wec_t, i int32) *Vec_Int_t     { return vec.Vec_WecEntry(v, int(i)) }
func Vec_WecPushLevel(v *Vec_Wec_t)                     { vec.Vec_WecPushLevel(v) }
func Vec_WecGrow(v *Vec_Wec_t, n int32)                 { vec.Vec_WecGrow(v, int(n)) }
func Vec_WecErase(v *Vec_Wec_t)                         { vec.Vec_WecErase(v) }
func Vec_WrdStart(n int32) *Vec_Wrd_t                   { return vec.Vec_WrdStart(int(n)) }
func Vec_WrdFreeP(v **Vec_Wrd_t)                        { vec.Vec_WrdFreeP(v) }
func Vec_WrdEntryP(v *Vec_Wrd_t, i int) *word           { return vec.Vec_WrdEntryP(v, i) }
func Abc_Lit2Var(lit int32) int32                       { return util.Abc_Lit2Var(lit) }
func Abc_LitIsCompl(lit int32) int32                    { return util.Abc_LitIsCompl(lit) }
func Abc_Var2Lit(v, c int32) int32                      { return util.Abc_Var2Lit(v, c) }

func boolToInt32(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
