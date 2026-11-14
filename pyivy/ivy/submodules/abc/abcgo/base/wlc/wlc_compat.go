package wlc

import (
	"fmt"
	aig "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	abc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
	ifpkg "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/acec"
	pdr "github.com/glycerine/abcgo/proof/pdr"
	"github.com/glycerine/abcgo/proof/ssw"
	"github.com/glycerine/abcgo/sat/bmc"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"io"
	"unsafe"
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

func Gia_ManCi(p *Gia_Man_t, i int32) *Gia_Obj_t            { return gia.Gia_ManCi(p, i) }
func Gia_ManCoNum(p *Gia_Man_t) int32                       { return gia.Gia_ManCoNum(p) }
func Gia_ManSetRegNum(p *Gia_Man_t, n int32)                { gia.Gia_ManSetRegNum(p, n) }
func Gia_ManHashDualMiter(p *Gia_Man_t, v *Vec_Int_t) int32 { return gia.Gia_ManHashDualMiter(p, v) }
func Gia_ManTransformMiter(p *Gia_Man_t) *Gia_Man_t         { return gia.Gia_ManTransformMiter(p) }
func Gia_ManCleanup(p *Gia_Man_t) *Gia_Man_t                { return gia.Gia_ManCleanup(p) }
func Gia_ManDupRemapLiterals(v *Vec_Int_t, p *Gia_Man_t)    { gia.Gia_ManDupRemapLiterals(v, p) }
func Gia_ManDupZeroUndc(p *Gia_Man_t, inits string, saveNames, simple, verbose int32) *Gia_Man_t {
	return gia.Gia_ManDupZeroUndc(p, inits, saveNames, simple, verbose)
}
func Gia_ManPoNum(p *Gia_Man_t) int32              { return gia.Gia_ManPoNum(p) }
func Gia_ManFromAigSimple(p *Aig_Man_t) *Gia_Man_t { return gia.Gia_ManFromAigSimple(p) }
func Gia_ManSetConstrNum(p *Gia_Man_t, n int32)    { gia.Gia_ManSetConstrNum(p, n) }
func Aig_ManSetConstrNum(p *Aig_Man_t, n int32)    { aig.Aig_ManSetConstrNum(p, n) }
func Gia_AigerWrite(p *Gia_Man_t, name string, symbols, compact, newline int32) error {
	return gia.Gia_AigerWrite(p, name, symbols, compact, newline)
}
func Abc_CexFreeP(p **Abc_Cex_t) { util.Abc_CexFreeP(p) }
func Saig_ManDupFoldConstrsFunc(p *Aig_Man_t, complement, verbose int32) *Aig_Man_t {
	return saig.Saig_ManDupFoldConstrsFunc(p, complement, verbose)
}
func Gia_ManDupNormalize(p *Gia_Man_t, hash int32) *Gia_Man_t {
	return gia.Gia_ManDupNormalize(p, hash)
}
func Gia_ManTransferTiming(dst, src *Gia_Man_t)             { gia.Gia_ManTransferTiming(dst, src) }
func Gia_ManSetBoxData(p, extra *Gia_Man_t, timing any)     { gia.Gia_ManSetBoxData(p, extra, timing) }
func Gia_ManSetInputNames(p *Gia_Man_t, names *Vec_Ptr_t)   { gia.Gia_ManSetInputNames(p, names) }
func Gia_ManSetOutputNames(p *Gia_Man_t, names *Vec_Ptr_t)  { gia.Gia_ManSetOutputNames(p, names) }
func Gia_ManSetRegClasses(p *Gia_Man_t, classes *Vec_Int_t) { gia.Gia_ManSetRegClasses(p, classes) }
func Gia_ManSimWordNum(p *Gia_Man_t) int32                  { return gia.Gia_ManSimWordNum(p) }
func Gia_ManSimData(p *Gia_Man_t) []uint64                  { return gia.Gia_ManSimData(p) }
func Gia_ManSimObjectData(p *Gia_Man_t, id int32) []uint64  { return gia.Gia_ManSimObjectData(p, id) }
func Gia_ManSetSimData(p *Gia_Man_t, data *Vec_Wrd_t, words int32) {
	gia.Gia_ManSetSimData(p, data, words)
}
func Gia_ManClearSimData(p *Gia_Man_t)                    { gia.Gia_ManClearSimData(p) }
func Abc_FrameUpdateGia(f *Abc_Frame_t, value *Gia_Man_t) { f.SetResource("pGia", value) }

type Mem_Flex_t = mem.Mem_Flex_t

func Mem_FlexStart() *Mem_Flex_t                { return mem.Mem_FlexStart() }
func Mem_FlexStop(p *Mem_Flex_t, verbose int32) { mem.Mem_FlexStop(p, verbose) }
func Mem_FlexReadMemUsage(p *Mem_Flex_t) int32  { return mem.Mem_FlexReadMemUsage(p) }
func Vec_PtrFreeP(p **Vec_Ptr_t)                { vec.Vec_PtrFreeP(p) }

type Abc_Frame_t = abcmain.Frame

func Abc_FrameGetGlobalFrame() *Abc_Frame_t        { return abcmain.Abc_FrameGetGlobalFrame() }
func Abc_FrameInputNdr(f *Abc_Frame_t, data any)   { abcmain.Abc_FrameInputNdr(f, data) }
func Abc_ShowFile(name string)                     { abc.Abc_ShowFile(name) }
func Abc_FrameReadProbStatus(f *Abc_Frame_t) int32 { return abcmain.Abc_FrameReadProbStatus(f) }
func Abc_FrameReadCex(f *Abc_Frame_t) *Abc_Cex_t {
	cex, _ := abcmain.Abc_FrameReadCex(f).(*Abc_Cex_t)
	return cex
}
func Abc_FrameSetBoxes(boxes []int32)                      { abcmain.Abc_FrameSetBoxes(boxes) }
func Cmd_CommandExecute(f *Abc_Frame_t, line string) int32 { return cmd.Cmd_CommandExecute(f, line) }

type Abc_Ntk_t = abc.Network
type Abc_Obj_t = abc.Object
type Aig_Man_t = aig.Graph
type abctime = int64
type word = uint64
type Cnf_Dat_t = cnf.Cnf_Dat_t
type Aig_Obj_t = aig.Edge
type Tim_Man_t = tim.Tim_Man_t
type If_LibBox_t = ifpkg.If_LibBox_t
type If_Box_t = ifpkg.If_Box_t
type Pdr_Man_t = pdr.Pdr_Man_t
type Pdr_Par_t = pdr.Pdr_Par_t
type Ssw_Pars_t = ssw.Ssw_Pars_t

var globalUtilOptind int32

func Extra_UtilGetoptReset() { extra.Extra_UtilGetoptReset(); globalUtilOptind = 0 }
func Extra_UtilGetopt(argc int32, argv []string, spec string) int32 {
	extra.Extra_UtilGetoptSetIndex(globalUtilOptind)
	c := extra.Extra_UtilGetopt(argc, argv, spec)
	globalUtilOptind = extra.Extra_UtilGetoptIndex()
	return c
}
func Extra_FileGetSimilarName(wrong, s1, s2, s3, s4, s5 string) string {
	return extra.Extra_FileGetSimilarName(wrong, s1, s2, s3, s4, s5)
}
func Extra_FileNameExtension(name string) string { return extra.Extra_FileNameExtension(name) }
func Extra_FileNameGeneric(name string) string   { return extra.Extra_FileNameGeneric(name) }
func Extra_FileDesignName(name string) string    { return extra.Extra_FileDesignName(name) }
func Extra_TimeStamp() string                    { return extra.Extra_TimeStamp() }
func Extra_FileCheck(name string) int32          { return extra.Extra_FileCheck(name) }
func Extra_FileReadContents(name string) string {
	contents, _ := extra.Extra_FileReadContents(name)
	return contents
}
func Extra_FileNameGenericAppend(name, suffix string) string {
	return extra.Extra_FileNameGenericAppend(name, suffix)
}

const (
	ABC_NTK_LOGIC = abc.ABC_NTK_LOGIC
	ABC_FUNC_SOP  = abc.ABC_FUNC_SOP
	ABC_INFINITY  = util.ABC_INFINITY
)

var (
	Pdr_InvCounts       = pdr.Pdr_InvCounts
	Pdr_InvPrintStr     = pdr.Pdr_InvPrintStr
	Pdr_InvPrint        = pdr.Pdr_InvPrint
	Pdr_InvCheck        = pdr.Pdr_InvCheck
	Pdr_InvMinimize     = pdr.Pdr_InvMinimize
	Pdr_InvMinimizeLits = pdr.Pdr_InvMinimizeLits
)

func Abc_FrameReadNtk(f *Abc_Frame_t) *Abc_Ntk_t {
	n, _ := abcmain.Abc_FrameReadNtk(f).(*Abc_Ntk_t)
	return n
}
func Abc_FrameReplaceCurrentNetwork(f *Abc_Frame_t, n *Abc_Ntk_t) {
	abcmain.Abc_FrameReplaceCurrentNetwork(f, n)
}

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
func Abc_NtkPi(n *Abc_Ntk_t, i int32) *Abc_Obj_t    { return abc.Abc_NtkPi(n, i) }
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
	Gia_ManRegNum             = gia.Gia_ManRegNum
	Gia_ManPiNum              = gia.Gia_ManPiNum
	Gia_ManObj                = gia.Gia_ManObj
	Gia_ManObjNum             = gia.Gia_ManObjNum
	Gia_ManCo                 = gia.Gia_ManCo
	Gia_ManRi                 = gia.Gia_ManRi
	Gia_ManRo                 = gia.Gia_ManRo
	Gia_ManCleanMark0         = gia.Gia_ManCleanMark0
	Gia_ObjId                 = gia.Gia_ObjId
	Gia_ManRandomW            = gia.Gia_ManRandomW
	Gia_ManRandom             = gia.Gia_ManRandom
	Gia_ManStop               = gia.Gia_ManStop
	Gia_ObjFaninC0            = gia.Gia_ObjFaninC0
	Gia_ObjFaninC1            = gia.Gia_ObjFaninC1
	Gia_ObjFaninId0           = gia.Gia_ObjFaninId0
	Gia_ObjFaninId1           = gia.Gia_ObjFaninId1
	Gia_ObjFanin0             = gia.Gia_ObjFanin0
	Gia_ObjFanin1             = gia.Gia_ObjFanin1
	Gia_ObjMark0              = gia.Gia_ObjMark0
	Gia_ObjSetMark0           = gia.Gia_ObjSetMark0
	Gia_ObjIsAnd              = gia.Gia_ObjIsAnd
	Gia_ObjIsCo               = gia.Gia_ObjIsCo
	Gia_ObjIsPi               = gia.Gia_ObjIsPi
	Gia_ObjIsCi               = gia.Gia_ObjIsCi
	Gia_ObjSetTravIdCurrentId = gia.Gia_ObjSetTravIdCurrentId
	Gia_ObjSetTravIdCurrent   = gia.Gia_ObjSetTravIdCurrent
	Gia_ObjIsTravIdCurrent    = gia.Gia_ObjIsTravIdCurrent
	Gia_ManIncrementTravId    = gia.Gia_ManIncrementTravId
	Gia_ObjIsRo               = gia.Gia_ObjIsRo
	Gia_ObjRoToRiId           = gia.Gia_ObjRoToRiId
)

func Gia_ManInputNames(p *Gia_Man_t) []string            { return p.InputNames() }
func Gia_ManStart(n int32) *Gia_Man_t                    { return gia.Gia_ManStart(n) }
func Gia_ManCiNum(p *Gia_Man_t) int32                    { return gia.Gia_ManCiNum(p) }
func Gia_ManPi(p *Gia_Man_t, i int32) *Gia_Obj_t         { return gia.Gia_ManPi(p, i) }
func Gia_ManHashAlloc(p *Gia_Man_t)                      { gia.Gia_ManHashAlloc(p) }
func Gia_ManAppendCi(p *Gia_Man_t) int32                 { return gia.Gia_ManAppendCi(p) }
func Gia_ManAppendCo(p *Gia_Man_t, lit int32) int32      { return gia.Gia_ManAppendCo(p, lit) }
func Gia_Obj2Lit(p *Gia_Man_t, obj *Gia_Obj_t) int32     { return gia.Gia_Obj2Lit(p, obj) }
func Gia_ManSetNameSpec(p *Gia_Man_t, name, spec string) { gia.Gia_ManSetNameSpec(p, name, spec) }
func Gia_ManSetSimple(p *Gia_Man_t, simple int32)        { gia.Gia_ManSetSimple(p, simple) }
func Abc_UtilStrsav(s string) string                     { return util.Abc_UtilStrsav(s) }
func Abc_FrameSetLibBox(box *If_LibBox_t)                { abcmain.Abc_FrameSetLibBox(box) }
func Abc_FrameReadLibBox() *If_LibBox_t {
	box, _ := abcmain.Abc_FrameReadLibBox().(*If_LibBox_t)
	return box
}
func If_LibBoxFree(box *If_LibBox_t)         { ifpkg.If_LibBoxFree(box) }
func Tim_ManStart(cis, cos int32) *Tim_Man_t { return tim.Tim_ManStart(int(cis), int(cos)) }
func Tim_ManCiNum(p *Tim_Man_t) int32        { return int32(tim.Tim_ManCiNum(p)) }
func Tim_ManCoNum(p *Tim_Man_t) int32        { return int32(tim.Tim_ManCoNum(p)) }
func Tim_ManCreateBox(p *Tim_Man_t, fi, ni, fo, no, table, black int32) {
	tim.Tim_ManCreateBox(p, int(fi), int(ni), int(fo), int(no), int(table), int(black))
}
func Tim_ManSetDelayTables(p *Tim_Man_t, tables *Vec_Ptr_t) {
	if tables == nil {
		tim.Tim_ManSetDelayTables(p, nil)
		return
	}
	converted := make([][]float32, Vec_PtrSize(tables))
	for i := range converted {
		converted[i] = Vec_PtrEntry(tables, int32(i)).([]float32)
	}
	tim.Tim_ManSetDelayTables(p, converted)
}
func If_LibBoxStart() *If_LibBox_t                     { return ifpkg.If_LibBoxStart() }
func If_LibBoxNum(p *If_LibBox_t) int32                { return ifpkg.If_LibBoxNum(p) }
func If_LibBoxAdd(p *If_LibBox_t, box *If_Box_t) error { return ifpkg.If_LibBoxAdd(p, box) }
func If_BoxStart(name string, id, inputs, outputs, seq, black, outer int32) *If_Box_t {
	return ifpkg.If_BoxStart(name, id, inputs, outputs, seq, black, outer)
}
func Vec_IntAlloc(n int32) *Vec_Int_t    { return vec.Vec_IntAlloc(int(n)) }
func Vec_IntDup(v *Vec_Int_t) *Vec_Int_t { return vec.Vec_IntDup(v) }
func Vec_IntAddToEntry(v *Vec_Int_t, i, add int32) int32 {
	return vec.Vec_IntAddToEntry(v, int(i), add)
}
func Abc_NamStart(a, b int32) *Abc_Nam_t { return util.Abc_NamStart(a, b) }
func Abc_NamStop(p *Abc_Nam_t)           { util.Abc_NamDeref(p) }
func Abc_NamStrFindOrAdd(p *Abc_Nam_t, s string, found *int32) int32 {
	return util.Abc_NamStrFindOrAdd(p, s, found)
}
func Abc_NamStrFind(p *Abc_Nam_t, s string) int32 { return util.Abc_NamStrFind(p, s) }
func Abc_NamStr(p *Abc_Nam_t, id int32) string    { return util.Abc_NamStr(p, id) }
func Abc_NamObjNumMax(p *Abc_Nam_t) int32         { return util.Abc_NamObjNumMax(p) }
func Abc_NamMemUsed(p *Abc_Nam_t) int32           { return util.Abc_NamMemUsed(p) }
func Abc_TtPrintHexArrayRev(out io.Writer, t *int32, n int32) {
	words32 := unsafe.Slice(t, int((n*4+31)/32))
	words64 := make([]word, (len(words32)+1)/2)
	for i, x := range words32 {
		words64[i/2] |= word(uint32(x)) << (32 * (i % 2))
	}
	util.Abc_TtPrintHexArrayRev(out, words64, n)
}

func Vec_IntSize(v *Vec_Int_t) int32           { return int32(vec.Vec_IntSize(v)) }
func Vec_IntCap(v *Vec_Int_t) int32            { return int32(vec.Vec_IntCap(v)) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32 { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntEntryLast(v *Vec_Int_t) int32      { return vec.Vec_IntEntryLast(v) }
func Vec_IntPush(v *Vec_Int_t, x int32)        { vec.Vec_IntPush(v, x) }
func Vec_IntPushTwo(v *Vec_Int_t, x, y int32)  { vec.Vec_IntPushTwo(v, x, y) }
func Vec_IntFillTwo(v *Vec_Int_t, n, x, y int32) {
	vec.Vec_IntClear(v)
	for i := int32(0); i < n; i++ {
		if i&1 == 0 {
			vec.Vec_IntPush(v, x)
		} else {
			vec.Vec_IntPush(v, y)
		}
	}
}
func Vec_IntPop(v *Vec_Int_t) int32                     { return vec.Vec_IntPop(v) }
func Vec_IntGrow(v *Vec_Int_t, n int32)                 { vec.Vec_IntGrow(v, int(n)) }
func Vec_IntUpdateEntry(v *Vec_Int_t, i int32, x int32) { vec.Vec_IntUpdateEntry(v, int(i), x) }
func Vec_IntFindMax(v *Vec_Int_t) int32                 { return vec.Vec_IntFindMax(v) }
func Abc_MinInt(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func Abc_Base10Log(n uint32) int32                      { return util.Abc_Base10Log(n) }
func Abc_AbsInt(n int32) int32                          { return util.Abc_AbsInt(n) }
func Abc_Base16Log(n uint32) int32                      { return util.Abc_Base16Log(n) }
func Vec_WrdAddToEntry(v *Vec_Wrd_t, i int32, x uint64) { vec.Vec_WrdAddToEntry(v, int(i), x) }
func Vec_WrdPush(v *Vec_Wrd_t, x uint64)                { vec.Vec_WrdPush(v, x) }
func Vec_WrdArray(v *Vec_Wrd_t) []uint64                { return vec.Vec_WrdArray(v) }
func Vec_WrdAlloc(n int32) *Vec_Wrd_t                   { return vec.Vec_WrdAlloc(int(n)) }
func Vec_WrdReverseOrder(v *Vec_Wrd_t)                  { vec.Vec_WrdReverseOrder(v) }
func Vec_PtrStart(n int32) *Vec_Ptr_t                   { return vec.Vec_PtrStart(int(n)) }
func Vec_VecFree(v *Vec_Vec_t)                          { vec.Vec_VecFree(v) }
func Vec_IntErase(v *Vec_Int_t)                         { vec.Vec_IntErase(v) }
func Vec_IntFree(v *Vec_Int_t)                          { vec.Vec_IntFree(v) }
func Vec_IntFreeP(v **Vec_Int_t)                        { vec.Vec_IntFreeP(v) }
func Vec_IntClear(v *Vec_Int_t)                         { vec.Vec_IntClear(v) }
func Vec_IntFill(v *Vec_Int_t, n, value int32)          { vec.Vec_IntFill(v, int(n), value) }
func Vec_IntFillExtra(v *Vec_Int_t, n, value int32)     { vec.Vec_IntFillExtra(v, int(n), value) }
func Vec_IntWriteEntry(v *Vec_Int_t, i, x int32)        { vec.Vec_IntWriteEntry(v, int(i), x) }
func Vec_IntArray(v *Vec_Int_t) []int32                 { return vec.Vec_IntArray(v) }
func Vec_IntReleaseNewArray(v *Vec_Int_t) []int32       { return vec.Vec_IntReleaseNewArray(v) }
func Vec_IntStart(n int32) *Vec_Int_t                   { return vec.Vec_IntStart(int(n)) }
func Vec_IntStartFull(n int32) *Vec_Int_t               { return vec.Vec_IntStartFull(int(n)) }
func Vec_IntStartNatural(n int32) *Vec_Int_t            { return vec.Vec_IntStartNatural(int(n)) }
func Vec_IntAppend(a, b *Vec_Int_t)                     { vec.Vec_IntAppend(a, b) }
func Vec_IntFind(a *Vec_Int_t, x int32) int32           { return int32(vec.Vec_IntFind(a, x)) }
func Vec_IntSort(a *Vec_Int_t, reverse int32)           { vec.Vec_IntSort(a, int(reverse)) }
func Vec_IntReverseOrder(a *Vec_Int_t)                  { vec.Vec_IntReverseOrder(a) }
func Vec_IntInsert(a *Vec_Int_t, i, x int32)            { vec.Vec_IntInsert(a, int(i), x) }
func Vec_IntPrint(a *Vec_Int_t)                         { vec.Vec_IntPrint(a) }
func Vec_IntShrink(a *Vec_Int_t, n int32)               { vec.Vec_IntShrink(a, int(n)) }
func Vec_IntPushUnique(a *Vec_Int_t, x int32) int32     { return int32(vec.Vec_IntPushUnique(a, x)) }
func Vec_IntPushUniqueOrder(a *Vec_Int_t, x int32) int32 {
	return int32(vec.Vec_IntPushUniqueOrder(a, x))
}
func Vec_IntPushOrder(a *Vec_Int_t, x int32)             { vec.Vec_IntPushOrder(a, x) }
func Vec_IntPushArray(a *Vec_Int_t, p []int32, n int32)  { vec.Vec_IntPushArray(a, p, int(n)) }
func Vec_StrArray(v *Vec_Str_t) []byte                   { return vec.Vec_StrArray(v) }
func Vec_StrSize(v *Vec_Str_t) int                       { return vec.Vec_StrSize(v) }
func Vec_StrPush(v *Vec_Str_t, c byte)                   { vec.Vec_StrPush(v, c) }
func Vec_StrAlloc(n int32) *Vec_Str_t                    { return vec.Vec_StrAlloc(int(n)) }
func Vec_StrFree(v *Vec_Str_t)                           { vec.Vec_StrFree(v) }
func Vec_StrFreeP(v **Vec_Str_t)                         { vec.Vec_StrFreeP(v) }
func Vec_PtrAlloc(n int32) *Vec_Ptr_t                    { return vec.Vec_PtrAlloc(int(n)) }
func Vec_PtrPush(v *Vec_Ptr_t, x any)                    { vec.Vec_PtrPush(v, x) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int32) any             { return vec.Vec_PtrEntry(v, int(i)) }
func Vec_PtrWriteEntry(v *Vec_Ptr_t, i int32, value any) { vec.Vec_PtrWriteEntry(v, int(i), value) }
func Vec_PtrSize(v *Vec_Ptr_t) int32                     { return int32(vec.Vec_PtrSize(v)) }
func Vec_PtrFree(v *Vec_Ptr_t)                           { vec.Vec_PtrFree(v) }
func Vec_PtrFreeFree(v *Vec_Ptr_t)                       { vec.Vec_PtrFreeFree(v) }
func Cmd_CommandAdd(f *Abc_Frame_t, group, name string, handler cmd.Cmd_CommandFuncType, changes int32) {
	cmd.Cmd_CommandAdd(f, group, name, handler, changes)
}
func Vec_PtrSort(v *Vec_Ptr_t, compare func(*any, *any) int) { vec.Vec_PtrSort(v, compare) }
func Vec_WecSize(v *Vec_Wec_t) int32                         { return int32(vec.Vec_WecSize(v)) }
func Vec_WecStart(n int32) *Vec_Wec_t                        { return vec.Vec_WecStart(int(n)) }
func Vec_WecPush(v *Vec_Wec_t, level, x int32)               { vec.Vec_WecPush(v, int(level), x) }
func Vec_WecFree(v *Vec_Wec_t)                               { vec.Vec_WecFree(v) }
func Vec_WecSizeSize(v *Vec_Wec_t) int32                     { return int32(vec.Vec_WecSizeSize(v)) }
func Vec_VecSize(v *Vec_Vec_t) int32                         { return int32(vec.Vec_VecSize(v)) }
func Vec_VecEntry(v *Vec_Vec_t, i int32) *Vec_Int_t          { return Vec_PtrEntry(v, i).(*Vec_Int_t) }
func Vec_WecCap(v *Vec_Wec_t) int32                          { return int32(vec.Vec_WecCap(v)) }
func Vec_WecEntry(v *Vec_Wec_t, i int32) *Vec_Int_t          { return vec.Vec_WecEntry(v, int(i)) }
func Vec_WecAlloc(n int32) *Vec_Wec_t                        { return vec.Vec_WecAlloc(int(n)) }
func Vec_WecEntryLast(v *Vec_Wec_t) *Vec_Int_t               { return vec.Vec_WecEntryLast(v) }
func Vec_WecPushLevel(v *Vec_Wec_t)                          { vec.Vec_WecPushLevel(v) }
func Vec_WecGrow(v *Vec_Wec_t, n int32)                      { vec.Vec_WecGrow(v, int(n)) }
func Vec_BitStart(n int32) *Vec_Bit_t                        { return vec.Vec_BitStart(int(n)) }
func Vec_BitStartFull(n int32) *Vec_Bit_t                    { return vec.Vec_BitStartFull(int(n)) }
func Vec_BitFree(v *Vec_Bit_t)                               { vec.Vec_BitFree(v) }
func Vec_BitReset(v *Vec_Bit_t)                              { vec.Vec_BitReset(v) }
func Vec_BitCount(v *Vec_Bit_t) int32                        { return int32(vec.Vec_BitCount(v)) }
func Vec_BitEntry(v *Vec_Bit_t, i int32) bool                { return vec.Vec_BitEntry(v, int(i)) != 0 }
func Vec_BitWriteEntry(v *Vec_Bit_t, i, x int32)             { vec.Vec_BitWriteEntry(v, int(i), int(x)) }
func Vec_WecErase(v *Vec_Wec_t)                              { vec.Vec_WecErase(v) }
func Vec_WrdStart(n int32) *Vec_Wrd_t                        { return vec.Vec_WrdStart(int(n)) }
func Vec_WrdFreeP(v **Vec_Wrd_t)                             { vec.Vec_WrdFreeP(v) }
func Vec_WrdFree(v *Vec_Wrd_t)                               { vec.Vec_WrdFree(v) }
func Vec_WrdEntryP(v *Vec_Wrd_t, i int) *word                { return vec.Vec_WrdEntryP(v, i) }
func Vec_WrdEntry(v *Vec_Wrd_t, i int32) word                { return vec.Vec_WrdEntry(v, int(i)) }
func Vec_WrdWriteEntry(v *Vec_Wrd_t, i int32, x word)        { vec.Vec_WrdWriteEntry(v, int(i), x) }
func Vec_WrdStartFull(n int32) *Vec_Wrd_t                    { return vec.Vec_WrdStartFull(int(n)) }
func Vec_WrdSize(v *Vec_Wrd_t) int32                         { return int32(vec.Vec_WrdSize(v)) }
func Abc_Lit2Var(lit int32) int32                            { return util.Abc_Lit2Var(lit) }
func Abc_LitIsCompl(lit int32) int32                         { return util.Abc_LitIsCompl(lit) }
func Abc_Var2Lit(v, c int32) int32                           { return util.Abc_Var2Lit(v, c) }

func boolToInt32(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func Abc_Print(_ int32, format string, args ...any) { fmt.Printf(format, args...) }
func ABC_PRTP(name string, t, total int64) {
	percent := 0.0
	if total != 0 {
		percent = 100 * float64(t) / float64(total)
	}
	fmt.Printf("%s = %9.2f sec (%6.2f %%)\n", name, float64(t)/1e9, percent)
}
func Abc_Clock() int64                                   { return util.Abc_Clock() }
func Abc_PrintTime(level int32, s string, elapsed int64) { util.Abc_PrintTime(level, s, elapsed) }
func Abc_MaxInt(a, b int32) int32                        { return util.Abc_MaxInt(a, b) }
func Abc_TtGetBit(p *word, i int32) int32                { return util.Abc_TtGetBit(unsafe.Slice(p, 1), i) }
func Abc_InfoHasBit(p []uint32, i int32) bool            { return util.Abc_InfoHasBit(p, i) }
func Abc_InfoSetBit(p []uint32, i int32)                 { util.Abc_InfoSetBit(p, i) }
func Abc_CexFree(p *Abc_Cex_t)                           { util.Abc_CexFree(p) }
func Bmc_CexCareMinimizeAig(p *Gia_Man_t, n int32, c *Abc_Cex_t, tries, check, verbose int32) *Abc_Cex_t {
	return bmc.Bmc_CexCareMinimizeAig(p, n, c, tries, check, verbose)
}

func Gia_ManToAigSimple(p *Gia_Man_t) *Aig_Man_t { return gia.Gia_ManToAigSimple(p) }
func Aig_ManAndNum(p *Aig_Man_t) int32           { return aig.Aig_ManAndNum(p) }
func Aig_ManDupSimple(p *Aig_Man_t) *Aig_Man_t   { return aig.Aig_ManDupSimple(p) }
func Ssw_ManSetDefaultParams(p *Ssw_Pars_t)      { ssw.Ssw_ManSetDefaultParams(p) }
func Ssw_SignalCorrespondence(p *Aig_Man_t, pars *Ssw_Pars_t) *Aig_Man_t {
	return ssw.Ssw_SignalCorrespondence(p, pars)
}
func Pdr_ManStart(p *Aig_Man_t, pars *Pdr_Par_t, priorities *Vec_Int_t) *Pdr_Man_t {
	return pdr.Pdr_ManStart(p, pars, priorities)
}
func Pdr_ManStop(p *Pdr_Man_t)                  { pdr.Pdr_ManStop(p) }
func IPdr_ManCheckCombUnsat(p *Pdr_Man_t) int32 { return pdr.IPdr_ManCheckCombUnsat(p) }
func IPdr_ManRebuildClauses(p *Pdr_Man_t, clauses *Vec_Vec_t) int32 {
	return pdr.IPdr_ManRebuildClauses(p, clauses)
}
func IPdr_ManRestoreClauses(p *Pdr_Man_t, clauses *Vec_Vec_t, mapping *Vec_Int_t) int32 {
	return pdr.IPdr_ManRestoreClauses(p, clauses, mapping)
}
func IPdr_ManSolveInt(p *Pdr_Man_t, check, push int32) int32 {
	return pdr.IPdr_ManSolveInt(p, check, push)
}
func IPdr_ManSaveClauses(p *Pdr_Man_t, drop int32) *Vec_Vec_t {
	return pdr.IPdr_ManSaveClauses(p, drop)
}
func Pdr_ManSetDefaultParams(p *Pdr_Par_t)                 { pdr.Pdr_ManSetDefaultParams(p) }
func Pdr_ManSolve(graph *Aig_Man_t, pars *Pdr_Par_t) int32 { return pdr.Pdr_ManSolve(graph, pars) }
func Gia_ManPermuteInputs(p *Gia_Man_t, a, b int32) *Gia_Man_t {
	return gia.Gia_ManPermuteInputs(p, a, b)
}
func Gia_ManTransformMiter2(p *Gia_Man_t) *Gia_Man_t  { return gia.Gia_ManTransformMiter2(p) }
func Gia_ManPrintStats(p *Gia_Man_t, _ any)           { gia.Gia_ManPrintStats(p, nil) }
func Gia_ManAndNum(p *Gia_Man_t) int32                { return gia.Gia_ManAndNum(p) }
func Gia_ManCleanLevels(p *Gia_Man_t, n int32)        { gia.Gia_ManCleanLevels(p, n) }
func Gia_ObjSetAndLevel(p *Gia_Man_t, obj *Gia_Obj_t) { gia.Gia_ObjSetAndLevel(p, obj) }
func Gia_ObjLevelId(p *Gia_Man_t, id int32) int32     { return gia.Gia_ObjLevelId(p, id) }
func Gia_ObjComputeTruth6Cis(p *Gia_Man_t, lit int32, support *Vec_Int_t, temp *Vec_Wrd_t) word {
	return gia.Gia_ObjComputeTruth6Cis(p, lit, support, temp)
}
func Extra_PrintHex(out io.Writer, truth *word, n int32) {
	extra.Extra_PrintHex(out, []uint32{uint32(*truth), uint32(*truth >> 32)}, n)
}
func Extra_PrintBinary(out io.Writer, value *word, bits int32) {
	extra.Extra_PrintBinary(out, []uint32{uint32(*value), uint32(*value >> 32)}, bits)
}
func Gia_ManHashAnd(p *Gia_Man_t, a, b int32) int32        { return gia.Gia_ManHashAnd(p, a, b) }
func Gia_ManHashOr(p *Gia_Man_t, a, b int32) int32         { return gia.Gia_ManHashOr(p, a, b) }
func Gia_ManHashMux(p *Gia_Man_t, c, t, e int32) int32     { return gia.Gia_ManHashMux(p, c, t, e) }
func Gia_ManHashAndMulti(p *Gia_Man_t, v *Vec_Int_t) int32 { return gia.Gia_ManHashAndMulti(p, v) }
func Gia_ManHashXor(p *Gia_Man_t, a, b int32) int32        { return gia.Gia_ManHashXor(p, a, b) }
func Abc_LitNot(l int32) int32                             { return util.Abc_LitNot(l) }
func Abc_LitNotCond(l, c int32) int32                      { return util.Abc_LitNotCond(l, c) }
func Abc_Base2Log(n int32) int32                           { return util.Abc_Base2Log(uint32(n)) }
func Vec_IntEntryP(v *Vec_Int_t, i int32) *int32           { return vec.Vec_IntEntryP(v, int(i)) }
func Vec_IntReleaseArray(v *Vec_Int_t) []int32             { return vec.Vec_IntReleaseArray(v) }
func Hash_Int2ManInsert(h *vec.Hash_IntMan_t, a, b, data int32) int32 {
	return vec.Hash_Int2ManInsert(h, a, b, data)
}
func ptrSlice(p *int32, n int32) []int32 {
	if p == nil || n <= 0 {
		return nil
	}
	return unsafe.Slice(p, int(n))
}
func Aig_ManCoNum(p *Aig_Man_t) int32             { return aig.Aig_ManCoNum(p) }
func Aig_ManCi(p *Aig_Man_t, i int32) Aig_Obj_t   { return aig.Aig_ManCi(p, i) }
func Aig_ManStop(p *Aig_Man_t)                    { aig.Aig_ManStop(p) }
func Cnf_Derive(p *Aig_Man_t, n int32) *Cnf_Dat_t { return cnf.Cnf_Derive(p, n) }
func Cnf_ManFree()                                { cnf.Cnf_ManFree() }
func toLitCond(v, c int32) bsat.Literal {
	if c != 0 {
		return bsat.Negative(v)
	}
	return bsat.Positive(v)
}
func lit_neg(l int32) int32                                       { return l ^ 1 }
func lit_var(l int32) int32                                       { return l >> 1 }
func sat_solver_new() *sat_solver                                 { return bsat.NewSolver() }
func sat_solver_setnvars(s *sat_solver, n int32)                  { s.SetNumVars(n) }
func sat_solver_set_runid(s *sat_solver, id int32)                { s.SetStopFunc(id, nil) }
func sat_solver_set_stop_func(s *sat_solver, f func(int32) int32) { s.SetStopFunc(0, f) }
func sat_solver_addclause(s *sat_solver, literals []int32) int32 {
	ls := make([]bsat.Literal, len(literals))
	for i, v := range literals {
		ls[i] = bsat.Literal(v)
	}
	if s.AddClause(ls...) {
		return 1
	}
	return 0
}
func sat_solver_solve(s *sat_solver, assumptions []int32, conflicts, propagations, globalConflicts, globalPropagations int64) int32 {
	ls := make([]bsat.Literal, len(assumptions))
	for i, v := range assumptions {
		ls[i] = bsat.Literal(v)
	}
	return int32(s.SolveWithResourceLimits(ls, conflicts, propagations, globalConflicts, globalPropagations))
}
func sat_solver_delete(s *sat_solver) { s.Close() }

func Vec_MemAlloc(words, logPage int32) *Vec_Mem_t       { return vec.Vec_MemAlloc(int(words), int(logPage)) }
func Vec_MemFreeP(v **Vec_Mem_t)                         { vec.Vec_MemFreeP(v) }
func Vec_MemEntrySize(v *Vec_Mem_t) int32                { return int32(vec.Vec_MemEntrySize(v)) }
func Vec_MemEntryNum(v *Vec_Mem_t) int32                 { return int32(vec.Vec_MemEntryNum(v)) }
func Vec_MemHashAlloc(v *Vec_Mem_t, n int32)             { vec.Vec_MemHashAlloc(v, int(n)) }
func Vec_MemHashFree(v *Vec_Mem_t)                       { vec.Vec_MemHashFree(v) }
func Vec_MemHashInsert(v *Vec_Mem_t, data []word) int32  { return int32(vec.Vec_MemHashInsert(v, data)) }
func Vec_MemHashLookup(v *Vec_Mem_t, data []word) *int32 { return vec.Vec_MemHashLookup(v, data) }
func Abc_TtNot(data []word, n int32)                     { util.Abc_TtNot(data, n) }
func Abc_Tt6Mask(n int32) word                           { return util.Abc_Tt6Mask(n) }
func Sdb_StoComputeCutsDetect(p *Gia_Man_t) *Vec_Int_t   { return acec.Sdb_StoComputeCutsDetect(p) }
