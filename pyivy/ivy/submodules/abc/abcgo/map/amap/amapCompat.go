package amap

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Vec_Vec_t = vec.Vec_Vec_t

func Vec_PtrAlloc(n int32) *Vec_Ptr_t              { return vec.Vec_PtrAlloc(int(n)) }
func Vec_PtrSize(v *Vec_Ptr_t) int32               { return int32(vec.Vec_PtrSize(v)) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any       { return vec.Vec_PtrEntry(v, int(i)) }
func Vec_PtrWriteEntry(v *Vec_Ptr_t, i int, x any) { vec.Vec_PtrWriteEntry(v, i, x) }
func Vec_PtrArray(v *Vec_Ptr_t) []any              { return vec.Vec_PtrArray(v) }
func Vec_PtrPush(v *Vec_Ptr_t, x any)              { vec.Vec_PtrPush(v, x) }
func Vec_PtrPop(v *Vec_Ptr_t) any                  { return vec.Vec_PtrPop(v) }
func Vec_PtrClear(v *Vec_Ptr_t)                    { vec.Vec_PtrClear(v) }
func Vec_PtrFree(v *Vec_Ptr_t)                     { vec.Vec_PtrFree(v) }
func Vec_PtrFreeP(v **Vec_Ptr_t)                   { vec.Vec_PtrFreeP(v) }
func Vec_PtrDup(v *Vec_Ptr_t) *Vec_Ptr_t           { return vec.Vec_PtrDup(v) }
func Vec_PtrFind(v *Vec_Ptr_t, x any) int32        { return int32(vec.Vec_PtrFind(v, x)) }

func Vec_IntAlloc(n int32) *Vec_Int_t               { return vec.Vec_IntAlloc(int(n)) }
func Vec_IntSize(v *Vec_Int_t) int32                { return int32(vec.Vec_IntSize(v)) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32      { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntArray(v *Vec_Int_t) []int32             { return vec.Vec_IntArray(v) }
func Vec_IntPush(v *Vec_Int_t, x int32)             { vec.Vec_IntPush(v, x) }
func Vec_IntPop(v *Vec_Int_t) int32                 { return vec.Vec_IntPop(v) }
func Vec_IntClear(v *Vec_Int_t)                     { vec.Vec_IntClear(v) }
func Vec_IntFree(v *Vec_Int_t)                      { vec.Vec_IntFree(v) }
func Vec_IntFreeP(v **Vec_Int_t)                    { vec.Vec_IntFreeP(v) }
func Vec_IntDup(v *Vec_Int_t) *Vec_Int_t            { return vec.Vec_IntDup(v) }
func Vec_IntWriteEntry(v *Vec_Int_t, i, x int32)    { vec.Vec_IntWriteEntry(v, int(i), x) }
func Vec_IntStartNatural(n int32) *Vec_Int_t        { return vec.Vec_IntStartNatural(int(n)) }
func Vec_IntGrow(v *Vec_Int_t, n int32)             { vec.Vec_IntGrow(v, int(n)) }
func Vec_IntShrink(v *Vec_Int_t, n int32)           { vec.Vec_IntShrink(v, int(n)) }
func Vec_IntPushUnique(v *Vec_Int_t, x int32) int32 { return int32(vec.Vec_IntPushUnique(v, x)) }
func Vec_IntPushOrder(v *Vec_Int_t, x int32)        { vec.Vec_IntPushOrder(v, x) }

func Vec_StrAlloc(n int32) *Vec_Str_t        { return vec.Vec_StrAlloc(int(n)) }
func Vec_StrArray(v *Vec_Str_t) []byte       { return vec.Vec_StrArray(v) }
func Vec_StrPush(v *Vec_Str_t, x byte)       { vec.Vec_StrPush(v, x) }
func Vec_StrPrintStr(v *Vec_Str_t, s string) { vec.Vec_StrPrintStr(v, s) }

func Vec_VecFree(v *Vec_Vec_t) { vec.Vec_VecFree(v) }

func Aig_MmFixedStart(size, maximum int32) *Aig_MmFixed_t  { return aig.Aig_MmFixedStart(size, maximum) }
func Aig_MmFixedStop(p *Aig_MmFixed_t, verbose int32)      { aig.Aig_MmFixedStop(p, verbose) }
func Aig_MmFixedEntryFetch(p *Aig_MmFixed_t) []byte        { return aig.Aig_MmFixedEntryFetch(p) }
func Aig_MmFlexStart() *Aig_MmFlex_t                       { return aig.Aig_MmFlexStart() }
func Aig_MmFlexStop(p *Aig_MmFlex_t, verbose int32)        { aig.Aig_MmFlexStop(p, verbose) }
func Aig_MmFlexRestart(p *Aig_MmFlex_t)                    { aig.Aig_MmFlexRestart(p) }
func Aig_MmFlexEntryFetch(p *Aig_MmFlex_t, n int32) []byte { return aig.Aig_MmFlexEntryFetch(p, n) }

func Abc_Clock() int64                             { return util.Abc_Clock() }
func Abc_AbsFloat(x float32) float32               { return util.Abc_AbsFloat(x) }
func Abc_Lit2Var(x int32) int32                    { return util.Abc_Lit2Var(x) }
func Abc_LitIsCompl(x int32) int32                 { return util.Abc_LitIsCompl(x) }
func Abc_LitNot(x int32) int32                     { return util.Abc_LitNot(x) }
func Abc_LitRegular(x int32) int32                 { return util.Abc_LitRegular(x) }
func Abc_MaxInt(a, b int32) int32                  { return util.Abc_MaxInt(a, b) }
func Abc_MinInt(a, b int32) int32                  { return util.Abc_MinInt(a, b) }
func Abc_TruthWordNum(n int32) int32               { return util.Abc_TruthWordNum(n) }
func Abc_UtilStrsav(s string) string               { return util.Abc_UtilStrsav(s) }
func Abc_Var2Lit(v, c int32) int32                 { return util.Abc_Var2Lit(v, c) }
func Abc_Print(level int32, f string, args ...any) { util.Abc_Print(level, f, args...) }

func Abc_FrameReadLibGen2() any { return abcmain.Abc_FrameReadLibGen2() }
