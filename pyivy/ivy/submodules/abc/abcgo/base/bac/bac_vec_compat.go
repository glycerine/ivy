package bac

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/base/cmd"
	abcio "github.com/glycerine/abcgo/base/io"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// These adapters keep the translated BAC API's int32 indices while routing
// storage operations through vec's public slice-backed implementation.
func Vec_IntPush(v *vec.Vec_Int_t, x int32)          { vec.Vec_IntPush(v, x) }
func Vec_IntPushTwo(v *vec.Vec_Int_t, a, b int32)    { vec.Vec_IntPushTwo(v, a, b) }
func Vec_IntWriteEntry(v *vec.Vec_Int_t, i, x int32) { vec.Vec_IntWriteEntry(v, int(i), x) }
func Vec_IntSetEntry(v *vec.Vec_Int_t, i, x int32)   { vec.Vec_IntSetEntry(v, int(i), x) }
func Vec_IntFill(v *vec.Vec_Int_t, n, x int32)       { vec.Vec_IntFill(v, int(n), x) }
func Vec_IntErase(v *vec.Vec_Int_t)                  { vec.Vec_IntErase(v) }
func Vec_IntGrow(v *vec.Vec_Int_t, n int32)          { vec.Vec_IntGrow(v, int(n)) }
func Vec_IntEntry(v *vec.Vec_Int_t, i int32) int32   { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntEntryLast(v *vec.Vec_Int_t) int32        { return vec.Vec_IntEntryLast(v) }
func Vec_IntAppend(v, w *vec.Vec_Int_t)              { vec.Vec_IntAppend(v, w) }
func Vec_IntSize(v *vec.Vec_Int_t) int32             { return int32(vec.Vec_IntSize(v)) }
func Vec_IntCap(v *vec.Vec_Int_t) int32              { return int32(vec.Vec_IntCap(v)) }
func Vec_IntMemory(v *vec.Vec_Int_t) float64         { return vec.Vec_IntMemory(v) }
func Vec_IntPrint(v *vec.Vec_Int_t)                  { vec.Vec_IntPrint(v) }

func Vec_StrPush(v *vec.Vec_Str_t, x byte)                { vec.Vec_StrPush(v, x) }
func Vec_StrWriteEntry(v *vec.Vec_Str_t, i int32, x byte) { vec.Vec_StrWriteEntry(v, int(i), x) }
func Vec_StrGrow(v *vec.Vec_Str_t, n int32)               { vec.Vec_StrGrow(v, int(n)) }
func Vec_StrErase(v *vec.Vec_Str_t)                       { vec.Vec_StrErase(v) }
func Vec_StrArray(v *vec.Vec_Str_t) []byte                { return vec.Vec_StrArray(v) }
func Vec_StrSize(v *vec.Vec_Str_t) int32                  { return int32(vec.Vec_StrSize(v)) }
func Vec_StrCap(v *vec.Vec_Str_t) int32                   { return int32(vec.Vec_StrCap(v)) }
func Vec_StrMemory(v *vec.Vec_Str_t) float64              { return vec.Vec_StrMemory(v) }
func Vec_StrEntry(v *vec.Vec_Str_t, i int32) byte         { return vec.Vec_StrEntry(v, int(i)) }
func Vec_StrCountEntryLit(v *vec.Vec_Str_t, x byte) int32 {
	return int32(vec.Vec_StrCountEntryLit(v, x))
}
func Vec_StrCountLargerLit(v *vec.Vec_Str_t, x byte) int32 {
	return int32(vec.Vec_StrCountLargerLit(v, x))
}
func Vec_StrCountSmallerLit(v *vec.Vec_Str_t, x byte) int32 {
	return int32(vec.Vec_StrCountSmallerLit(v, x))
}
func Abc_Lit2Att2(x int32) int32    { return util.Abc_Lit2Att2(x) }
func Abc_Base10Log(n int32) int32   { return util.Abc_Base10Log(uint32(n)) }
func Abc_Lit2Var2(x int32) int32    { return util.Abc_Lit2Var2(x) }
func Abc_Lit2Var(x int32) int32     { return util.Abc_Lit2Var(x) }
func Abc_LitIsCompl(x int32) int32  { return util.Abc_LitIsCompl(x) }
func Abc_LitNot(x int32) int32      { return util.Abc_LitNot(x) }
func Abc_Var2Lit(v, c int32) int32  { return util.Abc_Var2Lit(v, c) }
func Abc_Var2Lit2(v, a int32) int32 { return util.Abc_Var2Lit2(v, a) }
func Abc_NamStrFindOrAdd(p *util.Abc_Nam_t, s string, found *int32) int32 {
	return util.Abc_NamStrFindOrAdd(p, s, found)
}

func Abc_NamStrFindOrAddLim(p *util.Abc_Nam_t, s string, limit int, found *int32) int32 {
	return util.Abc_NamStrFindOrAddLim(p, s, limit, found)
}

func Vec_IntClear(p *vec.Vec_Int_t) { vec.Vec_IntClear(p) }

type Vec_Int_t = vec.Vec_Int_t

func Vec_StrClear(p *vec.Vec_Str_t)          { vec.Vec_StrClear(p) }
func Vec_StrAlloc(n int32) *vec.Vec_Str_t    { return vec.Vec_StrAlloc(int(n)) }
func Vec_StrFreeP(p **vec.Vec_Str_t)         { vec.Vec_StrFreeP(p) }
func Vec_StrEntryLast(p *vec.Vec_Str_t) byte { return vec.Vec_StrEntryLast(p) }

type Vec_Str_t = vec.Vec_Str_t

func Vec_StrPrintStr(p *vec.Vec_Str_t, s string) { vec.Vec_StrPrintStr(p, s) }
func Vec_StrPrintNum(p *vec.Vec_Str_t, n int32)  { vec.Vec_StrPrintNum(p, n) }
func Vec_PtrSize(p *vec.Vec_Ptr_t) int32         { return int32(vec.Vec_PtrSize(p)) }
func Vec_PtrEntry(p *vec.Vec_Ptr_t, i int32) any { return vec.Vec_PtrEntry(p, int(i)) }

type abctime = int64

func Abc_Clock() int64 { return util.Abc_Clock() }
func Abc_PrintTime(level int32, label string, elapsed int64) {
	util.Abc_PrintTime(level, label, elapsed)
}
func Abc_NamStart(n, average int32) *util.Abc_Nam_t { return util.Abc_NamStart(n, average) }
func Abc_NamRef(p *util.Abc_Nam_t) *util.Abc_Nam_t  { return util.Abc_NamRef(p) }
func Abc_NamDeref(p *util.Abc_Nam_t)                { util.Abc_NamDeref(p) }
func Abc_NamMemUsed(p *util.Abc_Nam_t) int32        { return util.Abc_NamMemUsed(p) }
func Abc_UtilStrsav(s string) string                { return util.Abc_UtilStrsav(s) }
func Extra_FileDesignName(s string) string          { return extra.Extra_FileDesignName(s) }
func Extra_UtilGetoptReset()                        { extra.Extra_UtilGetoptReset() }
func Extra_UtilGetopt(argc int32, argv []string, spec string) int32 {
	return extra.Extra_UtilGetopt(argc, argv, spec)
}
func Extra_UtilGetoptArg() string      { return extra.Extra_UtilGetoptArg() }
func Extra_UtilGetoptIndex() int32     { return extra.Extra_UtilGetoptIndex() }
func Extra_UtilGetoptSetIndex(i int32) { extra.Extra_UtilGetoptSetIndex(i) }
func Extra_FileGetSimilarName(a, b, c, d, e, f string) string {
	return extra.Extra_FileGetSimilarName(a, b, c, d, e, f)
}
func Extra_FileNameExtension(s string) string { return extra.Extra_FileNameExtension(s) }
func Extra_FileNameGenericAppend(s, suffix string) string {
	return extra.Extra_FileNameGenericAppend(s, suffix)
}
func Extra_TimeStamp() string                           { return extra.Extra_TimeStamp() }
func Abc_Print(level int32, format string, args ...any) { util.Abc_Print(level, format, args...) }
func Io_ReadFileType(s string) abcio.Io_FileType_t      { return abcio.Io_ReadFileType(s) }
func Io_ReadNetlist(s string, typ abcio.Io_FileType_t, check int32) *abc.Network {
	return abcio.Io_ReadNetlist(s, typ, check)
}
func Abc_NtkDelete(n *abc.Network) { abc.Abc_NtkDelete(n) }
func Cmd_CommandAdd(f *abcmain.Abc_Frame_t, group, name string, handler cmd.Cmd_CommandFuncType, changes int32) {
	cmd.Cmd_CommandAdd(f, group, name, handler, changes)
}
func Abc_NamObjNumMax(p *util.Abc_Nam_t) int32      { return util.Abc_NamObjNumMax(p) }
func Abc_NamStr(p *util.Abc_Nam_t, id int32) string { return util.Abc_NamStr(p, id) }
