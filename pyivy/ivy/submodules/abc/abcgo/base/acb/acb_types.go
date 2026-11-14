package acb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	miniaig "github.com/glycerine/abcgo/aig/miniaig"
	baseabc "github.com/glycerine/abcgo/base/abc"
	basemain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"github.com/glycerine/abcgo/sat/satoko"
	"os"
	"unsafe"
)

var s_Truths6 = [6]word{0xAAAAAAAAAAAAAAAA, 0xCCCCCCCCCCCCCCCC, 0xF0F0F0F0F0F0F0F0, 0xFF00FF00FF00FF00, 0xFFFF0000FFFF0000, 0xFFFFFFFF00000000}

// These names are the vector and name-manager types used by ABC's acb.h.
// Keep the underlying Go implementations shared with their owning packages.
type (
	Aig_Man_t     = aig.Graph
	Abc_Ntk_t     = baseabc.Abc_Ntk_t
	Abc_Obj_t     = baseabc.Abc_Obj_t
	Cnf_Dat_t     = cnf.Cnf_Dat_t
	Gia_Obj_t     = gia.Gia_Obj_t
	satoko_t      = satoko.Solver
	Gia_Man_t     = gia.Gia_Man_t
	Vec_Bit_t     = vec.Vec_Bit_t
	word          = uint64
	Vec_Int_t     = vec.Vec_Int_t
	Vec_Str_t     = vec.Vec_Str_t
	Vec_Wrd_t     = vec.Vec_Wrd_t
	Vec_Flt_t     = vec.Vec_Flt_t
	Vec_Wec_t     = vec.Vec_Wec_t
	Vec_Ptr_t     = vec.Vec_Ptr_t
	Vec_Que_t     = vec.Vec_Que_t
	Hash_IntMan_t = vec.Hash_IntMan_t
	Abc_Nam_t     = util.Abc_Nam_t
	Mem_Flex_t    = mem.Mem_Flex_t
	Ndr_Data_t    = miniaig.Ndr_Data_t
	Mio_Library_t = mio.Mio_Library_t
)

// Acb_Aig_t is opaque in acb.h; the package only stores pointers to it.
type Acb_Aig_t_ struct{}

type Acb_ObjType_t int32

func Extra_FileDesignName(name string) string     { return extra.Extra_FileDesignName(name) }
func Abc_UtilStrsav(s string) string              { return util.Abc_UtilStrsav(s) }
func Abc_NamStart(n, average int32) *Abc_Nam_t    { return util.Abc_NamStart(n, average) }
func Abc_NamStrFind(p *Abc_Nam_t, s string) int32 { return util.Abc_NamStrFind(p, s) }
func Abc_NamStr(p *Abc_Nam_t, id int32) string    { return util.Abc_NamStr(p, id) }
func Abc_NamObjNumMax(p *Abc_Nam_t) int32         { return util.Abc_NamObjNumMax(p) }
func Hash_IntManStart(n int32) *Hash_IntMan_t     { return vec.Hash_IntManStart(int(n)) }
func boolInt(v bool, yes, no int32) int32 {
	if v {
		return yes
	}
	return no
}
func Abc_Lit2Var(l int32) int32    { return util.Abc_Lit2Var(l) }
func Abc_LitIsCompl(l int32) int32 { return util.Abc_LitIsCompl(l) }
func Abc_Var2Lit(v, c int32) int32 { return util.Abc_Var2Lit(v, c) }
func Abc_MaxInt(a, b int32) int32  { return util.Abc_MaxInt(a, b) }
func Abc_NamStrFindOrAdd(p *Abc_Nam_t, s string, found *int32) int32 {
	return util.Abc_NamStrFindOrAdd(p, s, found)
}
func Abc_NamRef(p *Abc_Nam_t) *Abc_Nam_t              { return util.Abc_NamRef(p) }
func Abc_NamDeref(p *Abc_Nam_t)                       { util.Abc_NamDeref(p) }
func Abc_NamMemUsed(p *Abc_Nam_t) int32               { return util.Abc_NamMemUsed(p) }
func Abc_NtkSpec(n *Abc_Ntk_t) string                 { return baseabc.Abc_NtkSpec(n) }
func Abc_NtkName(n *Abc_Ntk_t) string                 { return baseabc.Abc_NtkName(n) }
func Abc_NtkCiNum(n *Abc_Ntk_t) int32                 { return baseabc.Abc_NtkCiNum(n) }
func Abc_NtkCoNum(n *Abc_Ntk_t) int32                 { return baseabc.Abc_NtkCoNum(n) }
func Abc_NtkObjNum(n *Abc_Ntk_t) int32                { return baseabc.Abc_NtkObjNum(n) }
func Abc_NtkObjNumMax(n *Abc_Ntk_t) int32             { return baseabc.Abc_NtkObjNumMax(n) }
func Abc_NtkObj(n *Abc_Ntk_t, i int32) *Abc_Obj_t     { return baseabc.Abc_NtkObj(n, i) }
func Abc_NtkCi(n *Abc_Ntk_t, i int32) *Abc_Obj_t      { return baseabc.Abc_NtkCi(n, i) }
func Abc_NtkCo(n *Abc_Ntk_t, i int32) *Abc_Obj_t      { return baseabc.Abc_NtkCo(n, i) }
func Abc_NtkLatchNum(n *Abc_Ntk_t) int32              { return baseabc.Abc_NtkLatchNum(n) }
func Abc_NtkIsSopLogic(n *Abc_Ntk_t) int32            { return baseabc.Abc_NtkIsSopLogic(n) }
func Abc_ObjId(o *Abc_Obj_t) int32                    { return baseabc.Abc_ObjId(o) }
func Abc_ObjIsNode(o *Abc_Obj_t) int32                { return baseabc.Abc_ObjIsNode(o) }
func Abc_ObjIsCi(o *Abc_Obj_t) int32                  { return baseabc.Abc_ObjIsCi(o) }
func Abc_ObjIsCo(o *Abc_Obj_t) int32                  { return baseabc.Abc_ObjIsCo(o) }
func Abc_ObjFaninNum(o *Abc_Obj_t) int32              { return baseabc.Abc_ObjFaninNum(o) }
func Abc_ObjFanin(o *Abc_Obj_t, i int32) *Abc_Obj_t   { return baseabc.Abc_ObjFanin(o, i) }
func Abc_ObjFanin0(o *Abc_Obj_t) *Abc_Obj_t           { return baseabc.Abc_ObjFanin0(o) }
func Abc_ObjName(o *Abc_Obj_t) string                 { return baseabc.Abc_ObjName(o) }
func Abc_NtkDfs(n *Abc_Ntk_t, all int32) []*Abc_Obj_t { return baseabc.Abc_NtkDfs(n, all) }
func Abc_NtkDelete(n *Abc_Ntk_t)                      { baseabc.Abc_NtkDelete(n) }
func Abc_NtkCheck(n *Abc_Ntk_t) int32                 { return baseabc.Abc_NtkCheck(n) }
func Abc_SopToTruth(s string, n int32) uint64         { return baseabc.Abc_SopToTruth(s, n) }
func Abc_SopGetVarNum(s string) int32                 { return baseabc.Abc_SopGetVarNum(s) }
func Abc_NtkStartFrom(n *Abc_Ntk_t, kind baseabc.Abc_NtkType_t, fn baseabc.Abc_NtkFunc_t) *Abc_Ntk_t {
	return baseabc.Abc_NtkStartFrom(n, kind, fn)
}
func Abc_NtkFuncManager(n *Abc_Ntk_t) *mem.Mem_Flex_t {
	manager, _ := n.FunctionManager.(*mem.Mem_Flex_t)
	return manager
}

const (
	ABC_NTK_LOGIC = baseabc.ABC_NTK_LOGIC
	ABC_FUNC_SOP  = baseabc.ABC_FUNC_SOP
	NDR_OBJECT    = miniaig.NDR_OBJECT
	NDR_TARGET    = miniaig.NDR_TARGET
	NDR_INPUT     = miniaig.NDR_INPUT
	NDR_OUTPUT    = miniaig.NDR_OUTPUT
	NDR_OPERTYPE  = miniaig.NDR_OPERTYPE
)

func Abc_NtkCreateNode(n *Abc_Ntk_t) *Abc_Obj_t { return baseabc.Abc_NtkCreateNode(n) }
func Abc_ObjAddFanin(o, f *Abc_Obj_t)           { baseabc.Abc_ObjAddFanin(o, f) }
func Abc_SopCreateFromTruthIsop(m *mem.Mem_Flex_t, n int32, truth []uint64, cover *Vec_Int_t) string {
	return baseabc.Abc_SopCreateFromTruthIsop(m, n, truth, cover)
}
func Vec_IntStart(n int32) *Vec_Int_t             { return vec.Vec_IntStart(int(n)) }
func Ndr_DataCiNum(p *Ndr_Data_t, i int32) int32  { return int32(miniaig.Ndr_DataCiNum(p, int(i))) }
func Ndr_DataCoNum(p *Ndr_Data_t, i int32) int32  { return int32(miniaig.Ndr_DataCoNum(p, int(i))) }
func Ndr_DataObjNum(p *Ndr_Data_t, i int32) int32 { return int32(miniaig.Ndr_DataObjNum(p, int(i))) }
func Ndr_DataEnd(p *Ndr_Data_t, i int32) int32    { return int32(miniaig.Ndr_DataEnd(p, int(i))) }
func Ndr_DataSize(p *Ndr_Data_t, i int32) int32   { return int32(miniaig.Ndr_DataSize(p, int(i))) }
func Ndr_DataType(p *Ndr_Data_t, i int32) int32   { return int32(miniaig.Ndr_DataType(p, int(i))) }
func Ndr_DataEntry(p *Ndr_Data_t, i int32) int32  { return int32(miniaig.Ndr_DataEntry(p, int(i))) }
func Ndr_ObjIsType(p *Ndr_Data_t, obj int32, oper Acb_ObjType_t) bool {
	return miniaig.Ndr_ObjIsType(p, int(obj), int(oper))
}
func Ndr_ObjReadBody(p *Ndr_Data_t, obj, typ int32) int32 {
	return int32(miniaig.Ndr_ObjReadBody(p, int(obj), int(typ)))
}
func Ndr_ObjReadArray(p *Ndr_Data_t, obj, typ int32, out *[]int32) int32 {
	var a []int
	n := miniaig.Ndr_ObjReadArray(p, int(obj), int(typ), &a)
	b := make([]int32, len(a))
	for i, v := range a {
		b[i] = int32(v)
	}
	*out = b
	return int32(n)
}
func truthWords(p *word, fanins int32) []uint64 {
	words := (1<<uint(fanins) + 63) / 64
	return unsafe.Slice((*uint64)(unsafe.Pointer(p)), words)
}
func Mio_LibraryReadBuffer(s string, extended int32, exclude map[string]bool, verbose int32) *Mio_Library_t {
	return mio.Mio_LibraryReadBuffer(s, extended, exclude, verbose)
}
func Mio_LibrarySetName(p *Mio_Library_t, s string) { mio.Mio_LibrarySetName(p, s) }
func Mio_UpdateGenlib(p *Mio_Library_t)             { mio.Mio_UpdateGenlib(p) }
func Ndr_Create(name int32) *Ndr_Data_t             { return miniaig.Ndr_Create(int(name)) }
func Ndr_AddModule(p *Ndr_Data_t, name int32) int32 {
	return int32(miniaig.Ndr_AddModule(p, int(name)))
}
func Ndr_DataResize(p *Ndr_Data_t, add int32)    { miniaig.Ndr_DataResize(p, int(add)) }
func Ndr_DataPush(p *Ndr_Data_t, typ, val int32) { miniaig.Ndr_DataPush(p, int(typ), int(val)) }
func Ndr_DataAddTo(p *Ndr_Data_t, i, add int32)  { miniaig.Ndr_DataAddTo(p, int(i), int(add)) }
func Ndr_AddObject(p *Ndr_Data_t, module int32, typ int32, name, left, right, signed, nIn int32, in *int32, nOut int32, out *int32, function any) {
	ins := make([]int, int(nIn))
	for i, v := range unsafe.Slice(in, int(nIn)) {
		ins[i] = int(v)
	}
	outs := make([]int, int(nOut))
	for i, v := range unsafe.Slice(out, int(nOut)) {
		outs[i] = int(v)
	}
	miniaig.Ndr_AddObject(p, int(module), int(typ), int(name), int(left), int(right), int(signed), int(nIn), ins, int(nOut), outs, function)
}
func Ndr_Delete(p *Ndr_Data_t) { miniaig.Ndr_Delete(p) }
func Extra_FileReadContents(name string) string {
	s, err := extra.Extra_FileReadContents(name)
	if err != nil {
		return ""
	}
	return s
}

func fileTell(f *os.File) int32   { pos, _ := f.Seek(0, 1); return int32(pos) }
func Abc_MinInt(a, b int32) int32 { return util.Abc_MinInt(a, b) }
func Abc_LitNot(l int32) int32    { return util.Abc_LitNot(l) }
func Abc_LitNotCond(l int32, c bool) int32 {
	x := int32(0)
	if c {
		x = 1
	}
	return util.Abc_LitNotCond(l, x)
}
func Gia_ManHashAnd(p *Gia_Man_t, a, b int32) int32       { return gia.Gia_ManHashAnd(p, a, b) }
func Gia_ManHashOr(p *Gia_Man_t, a, b int32) int32        { return gia.Gia_ManHashOr(p, a, b) }
func Gia_ManHashXor(p *Gia_Man_t, a, b int32) int32       { return gia.Gia_ManHashXor(p, a, b) }
func Gia_ManStart(n int32) *Gia_Man_t                     { return gia.Gia_ManStart(n) }
func Gia_ManHashAlloc(p *Gia_Man_t)                       { gia.Gia_ManHashAlloc(p) }
func Gia_ManAppendCi(p *Gia_Man_t) int32                  { return gia.Gia_ManAppendCi(p) }
func Gia_ManAppendCo(p *Gia_Man_t, l int32) int32         { return gia.Gia_ManAppendCo(p, l) }
func Gia_ManHashStop(p *Gia_Man_t)                        { gia.Gia_ManHashStop(p) }
func Gia_ManCleanup(p *Gia_Man_t) *Gia_Man_t              { return gia.Gia_ManCleanup(p) }
func Gia_ManStop(p *Gia_Man_t)                            { gia.Gia_ManStop(p) }
func Gia_ManFillValue(p *Gia_Man_t)                       { gia.Gia_ManFillValue(p) }
func Gia_ManObjNum(p *Gia_Man_t) int32                    { return gia.Gia_ManObjNum(p) }
func Gia_ManConst0(p *Gia_Man_t) *Gia_Obj_t               { return gia.Gia_ManConst0(p) }
func Gia_ManCi(p *Gia_Man_t, i int32) *Gia_Obj_t          { return gia.Gia_ManCi(p, i) }
func Gia_ManCo(p *Gia_Man_t, i int32) *Gia_Obj_t          { return gia.Gia_ManCo(p, i) }
func Gia_ManCiNum(p *Gia_Man_t) int32                     { return gia.Gia_ManCiNum(p) }
func Gia_ManCoNum(p *Gia_Man_t) int32                     { return gia.Gia_ManCoNum(p) }
func Gia_ManAndNum(p *Gia_Man_t) int32                    { return gia.Gia_ManAndNum(p) }
func Gia_ManDup(p *Gia_Man_t) *Gia_Man_t                  { return gia.Gia_ManDup(p) }
func Gia_ManPrintStats(p *Gia_Man_t, pars *gia.Gps_Par_t) { gia.Gia_ManPrintStats(p, pars) }
func Gia_ManAreaBalance(p *Gia_Man_t, simple, limit, verbose, veryVerbose int32) *Gia_Man_t {
	return gia.Gia_ManAreaBalance(p, simple, limit, verbose, veryVerbose)
}
func Gia_ManCompress2(p *Gia_Man_t, update, verbose int32) *Gia_Man_t {
	return gia.Gia_ManCompress2(p, update, verbose)
}
func Gia_ManAigSyn2(p *Gia_Man_t, old, coarsen, cutMin, relax, delayMin, verbose, veryVerbose int32) *Gia_Man_t {
	return gia.Gia_ManAigSyn2(p, old, coarsen, cutMin, relax, delayMin, verbose, veryVerbose)
}
func Gia_ManDupUniv(p *Gia_Man_t, variable int32) *Gia_Man_t { return gia.Gia_ManDupUniv(p, variable) }
func Gia_ManDupCofactorVar(p *Gia_Man_t, variable, value int32) *Gia_Man_t {
	return gia.Gia_ManDupCofactorVar(p, variable, value)
}
func Gia_ManDupMux(variable int32, positive, negative *Gia_Man_t) *Gia_Man_t {
	return gia.Gia_ManDupMux(variable, positive, negative)
}
func Gia_ManSetSimInputs(p *Gia_Man_t, inputs *Vec_Wrd_t) { gia.Gia_ManSetSimInputs(p, inputs) }
func Gia_ManSimPatSim(p *Gia_Man_t) *Vec_Wrd_t            { return gia.Gia_ManSimPatSim(p) }
func Gia_ManSetNameSpec(p *Gia_Man_t, name, spec string)  { gia.Gia_ManSetNameSpec(p, name, spec) }
func Gia_ManToAig(p *Gia_Man_t, choices int32) *Aig_Man_t { return gia.Gia_ManToAig(p, choices) }
func Aig_ManStop(p *Aig_Man_t)                            { aig.Aig_ManStop(p) }
func Gia_ManHashMux(p *Gia_Man_t, control, then, otherwise int32) int32 {
	return gia.Gia_ManHashMux(p, control, then, otherwise)
}
func Vec_WrdStartRandom(n int32) *Vec_Wrd_t { return vec.Vec_WrdStartRandom(int(n)) }
func Vec_WrdDup(p *Vec_Wrd_t) *Vec_Wrd_t    { return vec.Vec_WrdDup(p) }
func Abc_TtCountOnesVec(data *word, n int32) int32 {
	return util.Abc_TtCountOnesVec(unsafe.Slice(data, int(n)), n)
}
func Gia_ManDupRemovePis(p *Gia_Man_t, n int32) *Gia_Man_t { return gia.Gia_ManDupRemovePis(p, n) }
func Gia_ManInterOne(on, off *Gia_Man_t, verbose int32) *Gia_Man_t {
	if on == nil || off == nil || gia.Gia_ManCiNum(on) != gia.Gia_ManCiNum(off) || gia.Gia_ManCoNum(on) != 1 || gia.Gia_ManCoNum(off) != 1 {
		return nil
	}
	pManOn := gia.Gia_ManToAigSimple(on)
	if pManOn == nil {
		return nil
	}
	pManOff := gia.Gia_ManToAigSimple(off)
	if pManOff == nil {
		aig.Aig_ManStop(pManOn)
		return nil
	}
	pManAig := aig.Aig_ManInter(pManOn, pManOff, 0, verbose)
	aig.Aig_ManStop(pManOn)
	aig.Aig_ManStop(pManOff)
	if pManAig == nil {
		return nil
	}
	pNtkAig := gia.Gia_ManFromAigSimple(pManAig)
	aig.Aig_ManStop(pManAig)
	return pNtkAig
}
func Abc_GiaSynthesizeInter(p *Gia_Man_t) *Gia_Man_t { return baseabc.Abc_GiaSynthesizeInter(p) }
func Gia_AigerWrite(p *Gia_Man_t, name string, symbols, compact, newline int32) {
	_ = gia.Gia_AigerWrite(p, name, symbols, compact, newline)
}
func Mf_ManGenerateCnf(p *Gia_Man_t, lutSize, objectIDs, addOr, withMapping, verbose int32) *Cnf_Dat_t {
	generated := gia.Mf_ManGenerateCnf(p, lutSize, objectIDs, addOr, withMapping, verbose)
	if generated == nil {
		return nil
	}
	offsets := []int32{0}
	var literals []int32
	for _, clause := range generated.Clauses {
		literals = append(literals, clause...)
		offsets = append(offsets, int32(len(literals)))
	}
	data := new(Cnf_Dat_t)
	cnf.Cnf_DataSetPackedClauses(data, int32(generated.Variables), offsets, literals)
	data.VarNums = append([]int32(nil), generated.VarNums...)
	data.Mapping = append([]int32(nil), generated.Mapping...)
	data.Obj2Clause = append([]int32(nil), generated.Obj2Clause...)
	data.Obj2Count = append([]int32(nil), generated.Obj2Count...)
	return data
}
func Gia_ObjIsCi(p *Gia_Obj_t) int32                         { return gia.Gia_ObjIsCi(p) }
func Gia_ObjCioId(p *Gia_Obj_t) int32                        { return gia.Gia_ObjCioId(p) }
func Vec_StrPrintF(v *Vec_Str_t, format string, args ...any) { vec.Vec_StrPrintF(v, format, args...) }
func Vec_PtrStart(n int32) *Vec_Ptr_t                        { return vec.Vec_PtrStart(int(n)) }
func Vec_PtrAlloc(n int32) *Vec_Ptr_t                        { return vec.Vec_PtrAlloc(int(n)) }
func Abc_GiaSynthesize(graphs *Vec_Ptr_t, multi *Gia_Man_t) *Vec_Wec_t {
	return baseabc.Abc_GiaSynthesize(graphs, multi)
}
func Abc_SopSynthesize(sops *Vec_Ptr_t) *Vec_Wec_t { return baseabc.Abc_SopSynthesize(sops) }
func Abc_FrameReadSpecName() string                { return basemain.Abc_FrameReadSpecName() }
func Abc_FrameReadSignalNames() *Vec_Ptr_t {
	values := basemain.Abc_FrameReadSignalNames()
	out := vec.Vec_PtrAlloc(len(values))
	for _, value := range values {
		vec.Vec_PtrPush(out, value)
	}
	return out
}
func Vec_PtrFree(v *Vec_Ptr_t)     { vec.Vec_PtrFree(v) }
func Vec_PtrFreeFree(v *Vec_Ptr_t) { vec.Vec_PtrFreeFree(v) }
func Vec_WecFree(v *Vec_Wec_t)     { vec.Vec_WecFree(v) }
func strncmp(a, b string, n int32) int32 {
	limit := int(n)
	if limit > len(a) {
		limit = len(a)
	}
	if limit > len(b) {
		limit = len(b)
	}
	for i := 0; i < limit; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
func Gia_ObjIsAnd(p *Gia_Obj_t) int32              { return gia.Gia_ObjIsAnd(p) }
func Gia_ObjFanin0Copy(p *Gia_Obj_t) int32         { return gia.Gia_ObjFanin0Copy(p) }
func Gia_ObjFanin1Copy(p *Gia_Obj_t) int32         { return gia.Gia_ObjFanin1Copy(p) }
func Gia_ObjFanin0(p *Gia_Obj_t) *Gia_Obj_t        { return gia.Gia_ObjFanin0(p) }
func Gia_ObjFanin1(p *Gia_Obj_t) *Gia_Obj_t        { return gia.Gia_ObjFanin1(p) }
func Gia_ObjId(p *Gia_Man_t, obj *Gia_Obj_t) int32 { return gia.Gia_ObjId(p, obj) }
func Gia_ObjIsTravIdCurrent(p *Gia_Man_t, obj *Gia_Obj_t) bool {
	return gia.Gia_ObjIsTravIdCurrent(p, obj) != 0
}
func Gia_ObjSetTravIdCurrent(p *Gia_Man_t, obj *Gia_Obj_t) { gia.Gia_ObjSetTravIdCurrent(p, obj) }
func Gia_ManIncrementTravId(p *Gia_Man_t)                  { gia.Gia_ManIncrementTravId(p) }
func Gia_ManObj(p *Gia_Man_t, i int32) *Gia_Obj_t          { return gia.Gia_ManObj(p, i) }

const (
	l_True  int32 = 1
	l_False int32 = -1
	l_Undef int32 = 0
)

func sat_solver_solve(s *bsat.Solver, assumptions, end []int32, conflicts, propagations, globalConflicts, globalPropagations int64) int32 {
	return int32(s.SolveWithResourceLimits(assumptions, conflicts, propagations, globalConflicts, globalPropagations))
}
func sat_solver_var_value(s *bsat.Solver, v int32) int32 {
	value, ok := s.Value(v)
	if !ok {
		return 0
	}
	if value {
		return 1
	}
	return 0
}
func sat_solver_new() *bsat.Solver                { return bsat.NewSolver() }
func sat_solver_setnvars(s *bsat.Solver, n int32) { s.SetNumVars(n) }
func sat_solver_nvars(s *bsat.Solver) int32       { return s.NumVars() }
func sat_solver_delete(s *bsat.Solver)            { s.Delete() }
func sat_solver_addclause(s *bsat.Solver, literals []int32) bool {
	return s.AddClause(literals...)
}
func Cnf_DataLift(p *Cnf_Dat_t, n int32)           { cnf.Cnf_DataLift(p, n) }
func Cnf_DataVarNum(p *Cnf_Dat_t) int32            { return cnf.Cnf_DataVarNum(p) }
func Cnf_DataClauseNum(p *Cnf_Dat_t) int32         { return cnf.Cnf_DataClauseNum(p) }
func Cnf_DataClause(p *Cnf_Dat_t, i int32) []int32 { return cnf.Cnf_DataClause(p, i) }
func Cnf_DataClauses(p *Cnf_Dat_t) [][]int32       { return cnf.Cnf_DataClauses(p) }
func Cnf_DataSetManNil(p *Cnf_Dat_t)               { cnf.Cnf_DataSetManNil(p) }
func Cnf_DataFree(p *Cnf_Dat_t)                    { cnf.Cnf_DataFree(p) }
func Cnf_DataWriteIntoSolver(p *Cnf_Dat_t, frames, init int32) *sat_solver {
	return cnf.Cnf_DataWriteIntoSolver(p, frames, init)
}
func Cnf_DataSetPackedClauses(p *Cnf_Dat_t, variables int32, offsets, literals []int32) {
	cnf.Cnf_DataSetPackedClauses(p, variables, offsets, literals)
}
func Abc_Tt6Cofactor0(t word, v int32) word { return util.Abc_Tt6Cofactor0(t, v) }
func Abc_Tt6Cofactor1(t word, v int32) word { return util.Abc_Tt6Cofactor1(t, v) }
func Abc_Tt6RemoveVar(t word, v int32) word { return util.Abc_Tt6RemoveVar(t, v) }
func Abc_Tt6HasVar(t word, v int32) int32   { return util.Abc_Tt6HasVar(t, v) }
func Abc_Tt6Flip(t word, v int32) word      { return util.Abc_Tt6Flip(t, v) }
func Abc_Tt6CheckOutDec(t word, v int32, out *word) int32 {
	var result [1]word
	phase := util.Abc_Tt6CheckOutDec(t, v, result[:])
	if out != nil {
		*out = result[0]
	}
	return phase
}
func Abc_TtCheckDsdAnd(t word, i, j int32, out *word) int32 {
	var result [1]word
	kind := util.Abc_TtCheckDsdAnd(t, i, j, result[:])
	if out != nil {
		*out = result[0]
	}
	return kind
}
func Abc_Random(reset int32) uint32 { return util.Abc_Random(reset) }
func sat_solver_add_xor(s *bsat.Solver, out, a, b, complement int32) bool {
	return s.AddXor(out, a, b, complement != 0)
}
func sat_solver_set_runtime_limit(s *bsat.Solver, deadline int64) {
	s.SetRuntimeDeadline(util.Abc_Clock, deadline)
}
func sat_solver_minimize_assumptions(s *bsat.Solver, lits []int32, n, budget int32) int32 {
	if n < 0 || int(n) > len(lits) {
		panic("invalid assumption count")
	}
	return s.MinimizeAssumptions(lits[:n], budget)
}
func sat_solver_var_literal(s *bsat.Solver, v int32) int32 {
	value, ok := s.Value(v)
	if !ok {
		panic("SAT variable has no model value")
	}
	if value {
		return bsat.Positive(v)
	}
	return bsat.Negative(v)
}
func sat_solver_push(s *bsat.Solver, lit int32) int32 { return boolInt(s.Push(lit), 1, 0) }
func sat_solver_pop(s *bsat.Solver)                   { s.Pop() }
func Vec_IntFind(v *Vec_Int_t, value int32) int32     { return int32(vec.Vec_IntFind(v, value)) }
func Abc_PrintTime(level int32, label string, elapsed int64) {
	util.Abc_PrintTime(level, label, elapsed)
}
func sat_solver_final(s *bsat.Solver, out *[]int32) int32 {
	*out = s.UnsatCore()
	return int32(len(*out))
}
func toLitCond(v int32, complement any) int32 {
	switch c := complement.(type) {
	case int32:
		return Abc_Var2Lit(v, c)
	case int:
		return Abc_Var2Lit(v, int32(c))
	case bool:
		if c {
			return Abc_Var2Lit(v, 1)
		}
		return Abc_Var2Lit(v, 0)
	default:
		panic("invalid literal complement")
	}
}
func satoko_create() *satoko.Solver                   { return satoko.New() }
func satoko_setnvars(s *satoko.Solver, n int32)       { s.SetVariables(int(n)) }
func satoko_options(s *satoko.Solver) *satoko.Options { return s.Options() }
func satoko_add_clause(s *satoko.Solver, literals any, sizes ...int32) int32 {
	var ls []int32
	switch v := literals.(type) {
	case []int32:
		ls = v
	case [2]int32:
		ls = v[:]
	case [3]int32:
		ls = v[:]
	case *int32:
		ls = []int32{*v}
	default:
		panic("unsupported satoko clause")
	}
	if len(sizes) != 0 {
		ls = ls[:sizes[0]]
	}
	ints := make([]int, len(ls))
	for i, l := range ls {
		ints[i] = int(l)
	}
	if s.AddClause(ints) {
		return 1
	}
	return 0
}
func satoko_solve_assumptions(s *satoko.Solver, literals []int32, n int32) int32 {
	ints := make([]int, n)
	for i := range ints {
		ints[i] = int(literals[i])
	}
	return int32(s.SolveAssumptions(ints))
}
func satoko_destroy(s *satoko.Solver) { s.Destroy() }
func satoko_minimize_assumptions(s *satoko.Solver, literals []int32, n, limit int32) int32 {
	ints := make([]int, n)
	for i := range ints {
		ints[i] = int(literals[i])
	}
	kept := s.MinimizeAssumptionsSource(ints, int64(limit))
	for i := 0; i < kept; i++ {
		literals[i] = int32(ints[i])
	}
	return int32(kept)
}
func satoko_final_conflict(s *satoko.Solver, out *[]int32) int32 {
	ints := s.FinalConflict()
	*out = make([]int32, len(ints))
	for i, v := range ints {
		(*out)[i] = int32(v)
	}
	return int32(len(ints))
}
func Abc_TtGetBit(p any, i int32) int32 {
	switch v := p.(type) {
	case []word:
		return util.Abc_TtGetBit(v, i)
	case *word:
		return util.Abc_TtGetBit(unsafe.Slice(v, int(i>>6)+1), i)
	default:
		panic("unsupported truth pointer")
	}
}
func Abc_TtSetBit(p any, i int32) {
	switch v := p.(type) {
	case []word:
		util.Abc_TtSetBit(v, i)
	case *word:
		util.Abc_TtSetBit(unsafe.Slice(v, int(i>>6)+1), i)
	default:
		panic("unsupported truth pointer")
	}
}
func Abc_Bit6WordNum(n int32) int32 { return util.Abc_Bit6WordNum(n) }
func asWords(v any, n int32) []word {
	switch x := v.(type) {
	case []word:
		return x
	case *word:
		return unsafe.Slice(x, int(n))
	default:
		panic("unsupported truth words")
	}
}
func Abc_TtCountOnesVecMask(x, mask any, n, compl int32) int32 {
	return util.Abc_TtCountOnesVecMask(asWords(x, n), asWords(mask, n), n, compl)
}
func Abc_TtAndSharp(out, a, b any, n, c int32) {
	util.Abc_TtAndSharp(asWords(out, n), asWords(a, n), asWords(b, n), n, c)
}
func Abc_TtConst(out any, n, c int32) { util.Abc_TtConst(asWords(out, n), n, c) }
func Abc_TtOr(out, a, b any, n int32) {
	util.Abc_TtOr(asWords(out, n), asWords(a, n), asWords(b, n), n)
}
func Abc_TtEqual(a, b any, n int32) bool {
	return util.Abc_TtEqual(asWords(a, n), asWords(b, n), n) != 0
}
func Abc_Clock() int64 { return util.Abc_Clock() }

var Gia_ManAppendBuf = gia.Gia_ManAppendBuf
var Gia_ManAppendAnd2 = gia.Gia_ManAppendAnd2
var Gia_ManAppendOr2 = gia.Gia_ManAppendOr2
var Gia_ManAppendXor2 = gia.Gia_ManAppendXor2
var Gia_ManAppendXorReal2 = gia.Gia_ManAppendXorReal2
var Gia_ManUpdateCopy = gia.Gia_ManUpdateCopy
var Gia_ObjFaninId0 = gia.Gia_ObjFaninId0
