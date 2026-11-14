package cba

import (
	gia "github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/cmd"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec"
)

// These are the package-local names used by cba.h, corresponding to the
// opaque C types declared by utilNam.h and the misc/vec headers.
type Abc_Nam_t = util.Abc_Nam_t
type Hash_IntMan_t = vec.Hash_IntMan_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Str_t = vec.Vec_Str_t
type Vec_Wrd_t = vec.Vec_Wrd_t
type Vec_Flt_t = vec.Vec_Flt_t
type Vec_Wec_t = vec.Vec_Wec_t
type Vec_Bit_t = vec.Vec_Bit_t
type Vec_Vec_t = vec.Vec_Vec_t
type Vec_Que_t = vec.Vec_Que_t

// CBA's C declarations use the unqualified Gia names; these aliases and API
// references preserve the existing Gia manager implementation.
type Gia_Man_t = gia.Gia_Man_t
type Gia_Obj_t = gia.Gia_Obj_t
type word = uint64
type Abc_Frame_t = abcmain.Frame
type Cec_ParCec_t = cec.Cec_ParCec_t
type Mio_Library_t = mio.Mio_Library_t
type Mio_Gate_t = mio.Mio_Gate_t
type abctime = int64

var (
	Gia_ManAndNum           = gia.Gia_ManAndNum
	Gia_ManAppendCi         = gia.Gia_ManAppendCi
	Gia_ManAppendCo         = gia.Gia_ManAppendCo
	Gia_ManCleanup          = gia.Gia_ManCleanup
	Gia_ManDupRemapLiterals = gia.Gia_ManDupRemapLiterals
	Gia_ManDupZeroUndc      = gia.Gia_ManDupZeroUndc
	Gia_ManHashAlloc        = gia.Gia_ManHashAlloc
	Gia_ManHashAnd          = gia.Gia_ManHashAnd
	Gia_ManHashMux          = gia.Gia_ManHashMux
	Gia_ManHashOr           = gia.Gia_ManHashOr
	Gia_ManHashXor          = gia.Gia_ManHashXor
	Gia_ManSetRegNum        = gia.Gia_ManSetRegNum
	Gia_ManStart            = gia.Gia_ManStart
	Gia_ManMiter            = gia.Gia_ManMiter
	Gia_AigerWrite          = gia.Gia_AigerWrite
	Gia_ManObj              = gia.Gia_ManObj
	Gia_ManName             = gia.Gia_ManName
	Gia_ManSpec             = gia.Gia_ManSpec
	Gia_ManCiNum            = gia.Gia_ManCiNum
	Gia_ManCoNum            = gia.Gia_ManCoNum
	Gia_ManObjNum           = gia.Gia_ManObjNum
	Gia_ManCi               = gia.Gia_ManCi
	Gia_ManCo               = gia.Gia_ManCo
	Gia_ManInputNames       = gia.Gia_ManInputNames
	Gia_ManOutputNames      = gia.Gia_ManOutputNames
	Gia_ObjId               = gia.Gia_ObjId
	Gia_ObjFaninLit0        = gia.Gia_ObjFaninLit0
	Gia_ObjFaninLit1        = gia.Gia_ObjFaninLit1
	Gia_ObjFaninLit2        = gia.Gia_ObjFaninLit2
	Gia_ObjIsMux            = gia.Gia_ObjIsMux
	Gia_ObjIsAnd            = gia.Gia_ObjIsAnd
	Gia_ObjIsXor            = gia.Gia_ObjIsXor
	Gia_ManSetNameSpec      = gia.Gia_ManSetNameSpec
	Gia_ManStop             = gia.Gia_ManStop
)

// Shared ABC utility and naming APIs used by the translated CBA sources.
var (
	Abc_NamStart                = util.Abc_NamStart
	Abc_NamStop                 = util.Abc_NamStop
	Abc_NamRef                  = util.Abc_NamRef
	Abc_NamDeref                = util.Abc_NamDeref
	Abc_NamStr                  = util.Abc_NamStr
	Abc_NamStrFind              = util.Abc_NamStrFind
	Abc_NamStrFindOrAdd         = util.Abc_NamStrFindOrAdd
	Abc_NamObjNumMax            = util.Abc_NamObjNumMax
	Abc_NamBuffer               = util.Abc_NamBuffer
	Abc_NamMemUsed              = util.Abc_NamMemUsed
	Abc_UtilStrsav              = util.Abc_UtilStrsav
	Abc_Ptr2Int                 = util.Abc_Ptr2Int
	Abc_Int2Ptr                 = util.Abc_Int2Ptr
	Abc_Lit2Var                 = util.Abc_Lit2Var
	Abc_LitIsCompl              = util.Abc_LitIsCompl
	Abc_LitNot                  = util.Abc_LitNot
	Abc_LitNotCond              = util.Abc_LitNotCond
	Abc_MaxInt                  = util.Abc_MaxInt
	Abc_Var2Lit                 = util.Abc_Var2Lit
	Abc_Lit2Var2                = util.Abc_Lit2Var2
	Abc_Lit2Att2                = util.Abc_Lit2Att2
	Abc_Var2Lit2                = util.Abc_Var2Lit2
	Extra_FileDesignName        = extra.Extra_FileDesignName
	Extra_UtilGetoptReset       = extra.Extra_UtilGetoptReset
	Extra_UtilGetopt            = extra.Extra_UtilGetopt
	Extra_UtilGetoptIndex       = extra.Extra_UtilGetoptIndex
	Extra_UtilGetoptSetIndex    = extra.Extra_UtilGetoptSetIndex
	Extra_FileGetSimilarName    = extra.Extra_FileGetSimilarName
	Extra_FileNameExtension     = extra.Extra_FileNameExtension
	Extra_FileNameGenericAppend = extra.Extra_FileNameGenericAppend
	Extra_TimeStamp             = extra.Extra_TimeStamp
	Abc_FrameUpdateGia          = abcmain.Abc_FrameUpdateGia
	Abc_FrameReadNtk            = abcmain.Abc_FrameReadNtk
	Abc_FrameReadGia            = abcmain.Abc_FrameReadGia
	Abc_FrameSetStatus          = abcmain.Abc_FrameSetStatus
	Abc_FrameReadFlag           = abcmain.Abc_FrameReadFlag
	Abc_FrameReadLibGen         = abcmain.Abc_FrameReadLibGen
	Cec_ManCecSetDefaultParams  = cec.Cec_ManCecSetDefaultParams
	Cec_ManVerify               = cec.Cec_ManVerify
	Abc_Print                   = util.Abc_Print
	Abc_MinInt                  = util.Abc_MinInt
	Abc_Base2Log                = util.Abc_Base2Log
	Abc_Clock                   = util.Abc_Clock
	Abc_PrintTime               = util.Abc_PrintTime
	Mio_LibraryReadGateByName   = mio.Mio_LibraryReadGateByName
	Mio_GateReadPinName         = mio.Mio_GateReadPinName
	Mio_GateReadOutName         = mio.Mio_GateReadOutName
)

func boolToInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func Cec_ParCecSetVerbose(p *Cec_ParCec_t, value int32) { cec.Cec_ParCecSetVerbose(p, value) }
func Cec_ParCecVerbose(p *Cec_ParCec_t) int32           { return cec.Cec_ParCecVerbose(p) }

func Hash_IntManStart(n int32) *Hash_IntMan_t { return vec.Hash_IntManStart(int(n)) }

func Cmd_CommandAdd(f *Abc_Frame_t, group, name string, handler func(*Abc_Frame_t, int32, []string) int32, changes int32) {
	cmd.Cmd_CommandAdd(f, group, name, cmd.Cmd_CommandFuncType(handler), changes)
}

func Abc_NamStrFindOrAddLim(p *Abc_Nam_t, s string, limit int, found *int32) int32 {
	return util.Abc_NamStrFindOrAddLim(p, s, limit, found)
}
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
