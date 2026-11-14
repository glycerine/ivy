package wln

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Vec_Int_t = vec.Vec_Int_t
type Vec_Str_t = vec.Vec_Str_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Hash_IntMan_t = vec.Hash_IntMan_t
type Abc_Nam_t = util.Abc_Nam_t

const (
	ABC_OPER_NONE         = miniaig.ABC_OPER_NONE
	ABC_OPER_CI           = miniaig.ABC_OPER_CI
	ABC_OPER_CO           = miniaig.ABC_OPER_CO
	ABC_OPER_FON          = miniaig.ABC_OPER_FON
	ABC_OPER_DFFRSE       = miniaig.ABC_OPER_DFFRSE
	ABC_OPER_CONST        = miniaig.ABC_OPER_CONST
	ABC_OPER_SLICE        = miniaig.ABC_OPER_SLICE
	ABC_OPER_SHIFT_ROTL   = miniaig.ABC_OPER_SHIFT_ROTL
	ABC_OPER_SHIFT_ROTR   = miniaig.ABC_OPER_SHIFT_ROTR
	ABC_OPER_TABLE        = miniaig.ABC_OPER_TABLE
	ABC_OPER_LUT          = miniaig.ABC_OPER_LUT
	ABC_OPER_RAMR         = miniaig.ABC_OPER_RAMR
	ABC_OPER_RAMW         = miniaig.ABC_OPER_RAMW
	ABC_OPER_CONCAT       = miniaig.ABC_OPER_CONCAT
	ABC_OPER_ARI_ADD      = miniaig.ABC_OPER_ARI_ADD
	ABC_OPER_ARI_ADDSUB   = miniaig.ABC_OPER_ARI_ADDSUB
	ABC_OPER_ARI_DIV      = miniaig.ABC_OPER_ARI_DIV
	ABC_OPER_ARI_MIN      = miniaig.ABC_OPER_ARI_MIN
	ABC_OPER_ARI_MOD      = miniaig.ABC_OPER_ARI_MOD
	ABC_OPER_ARI_MUL      = miniaig.ABC_OPER_ARI_MUL
	ABC_OPER_ARI_POW      = miniaig.ABC_OPER_ARI_POW
	ABC_OPER_ARI_REM      = miniaig.ABC_OPER_ARI_REM
	ABC_OPER_ARI_SMUL     = miniaig.ABC_OPER_ARI_SMUL
	ABC_OPER_ARI_SQRT     = miniaig.ABC_OPER_ARI_SQRT
	ABC_OPER_ARI_SQUARE   = miniaig.ABC_OPER_ARI_SQUARE
	ABC_OPER_ARI_SUB      = miniaig.ABC_OPER_ARI_SUB
	ABC_OPER_BIT_AND      = miniaig.ABC_OPER_BIT_AND
	ABC_OPER_BIT_BUF      = miniaig.ABC_OPER_BIT_BUF
	ABC_OPER_BIT_INV      = miniaig.ABC_OPER_BIT_INV
	ABC_OPER_BIT_MUX      = miniaig.ABC_OPER_BIT_MUX
	ABC_OPER_BIT_NAND     = miniaig.ABC_OPER_BIT_NAND
	ABC_OPER_BIT_NOR      = miniaig.ABC_OPER_BIT_NOR
	ABC_OPER_BIT_NXOR     = miniaig.ABC_OPER_BIT_NXOR
	ABC_OPER_BIT_OR       = miniaig.ABC_OPER_BIT_OR
	ABC_OPER_BIT_XOR      = miniaig.ABC_OPER_BIT_XOR
	ABC_OPER_COMP_EQU     = miniaig.ABC_OPER_COMP_EQU
	ABC_OPER_COMP_LESS    = miniaig.ABC_OPER_COMP_LESS
	ABC_OPER_COMP_LESSEQU = miniaig.ABC_OPER_COMP_LESSEQU
	ABC_OPER_COMP_MORE    = miniaig.ABC_OPER_COMP_MORE
	ABC_OPER_COMP_MOREEQU = miniaig.ABC_OPER_COMP_MOREEQU
	ABC_OPER_COMP_NOTEQU  = miniaig.ABC_OPER_COMP_NOTEQU
	ABC_OPER_DFF          = miniaig.ABC_OPER_DFF
	ABC_OPER_LOGIC_AND    = miniaig.ABC_OPER_LOGIC_AND
	ABC_OPER_LOGIC_IMPL   = miniaig.ABC_OPER_LOGIC_IMPL
	ABC_OPER_LOGIC_NOT    = miniaig.ABC_OPER_LOGIC_NOT
	ABC_OPER_LOGIC_OR     = miniaig.ABC_OPER_LOGIC_OR
	ABC_OPER_LOGIC_XOR    = miniaig.ABC_OPER_LOGIC_XOR
	ABC_OPER_RED_AND      = miniaig.ABC_OPER_RED_AND
	ABC_OPER_RED_NAND     = miniaig.ABC_OPER_RED_NAND
	ABC_OPER_RED_NOR      = miniaig.ABC_OPER_RED_NOR
	ABC_OPER_RED_NXOR     = miniaig.ABC_OPER_RED_NXOR
	ABC_OPER_RED_OR       = miniaig.ABC_OPER_RED_OR
	ABC_OPER_RED_XOR      = miniaig.ABC_OPER_RED_XOR
	ABC_OPER_SEL_DEC      = miniaig.ABC_OPER_SEL_DEC
	ABC_OPER_SEL_NMUX     = miniaig.ABC_OPER_SEL_NMUX
	ABC_OPER_SEL_SEL      = miniaig.ABC_OPER_SEL_SEL
	ABC_OPER_SHIFT_L      = miniaig.ABC_OPER_SHIFT_L
	ABC_OPER_SHIFT_LA     = miniaig.ABC_OPER_SHIFT_LA
	ABC_OPER_SHIFT_R      = miniaig.ABC_OPER_SHIFT_R
	ABC_OPER_SHIFT_RA     = miniaig.ABC_OPER_SHIFT_RA
	ABC_OPER_SIGNEXT      = miniaig.ABC_OPER_SIGNEXT
	ABC_OPER_ZEROPAD      = miniaig.ABC_OPER_ZEROPAD
	ABC_OPER_LAST         = miniaig.ABC_OPER_LAST
)

func Abc_OperName(typ int32) string { return miniaig.Abc_OperName(int(typ)) }

func Vec_IntSize(v *Vec_Int_t) int32                    { return int32(vec.Vec_IntSize(v)) }
func Vec_IntCap(v *Vec_Int_t) int32                     { return int32(vec.Vec_IntCap(v)) }
func Vec_IntEntry(v *Vec_Int_t, i int32) int32          { return vec.Vec_IntEntry(v, int(i)) }
func Vec_IntEntryP(v *Vec_Int_t, i int32) *int32        { return vec.Vec_IntEntryP(v, int(i)) }
func Vec_IntWriteEntry(v *Vec_Int_t, i, value int32)    { vec.Vec_IntWriteEntry(v, int(i), value) }
func Vec_IntFill(v *Vec_Int_t, n, value int32)          { vec.Vec_IntFill(v, int(n), value) }
func Vec_StrEntry(v *Vec_Str_t, i int32) byte           { return vec.Vec_StrEntry(v, int(i)) }
func Vec_StrSetEntry(v *Vec_Str_t, i int32, value byte) { vec.Vec_StrSetEntry(v, int(i), value) }
func Hash_IntObjData0(v *Hash_IntMan_t, i int32) int32  { return vec.Hash_IntObjData0(v, i) }
func Hash_IntObjData1(v *Hash_IntMan_t, i int32) int32  { return vec.Hash_IntObjData1(v, i) }
func Abc_AbsInt(x int32) int32                          { return util.Abc_AbsInt(x) }
func Abc_Var2Lit(v, c int32) int32                      { return util.Abc_Var2Lit(v, c) }
func Abc_Base10Log(x uint32) int32                      { return int32(util.Abc_Base10Log(x)) }

func wlnIntArray(v *Vec_Int_t) []int32 { return vec.Vec_IntArray(v) }

type Hsh_VecMan_t = vec.Hsh_VecMan_t

func Vec_IntAlloc(n int32) *Vec_Int_t                   { return vec.Vec_IntAlloc(int(n)) }
func Vec_IntStart(n int32) *Vec_Int_t                   { return vec.Vec_IntStart(int(n)) }
func Vec_IntAppend(a, b *Vec_Int_t)                     { vec.Vec_IntAppend(a, b) }
func Vec_IntArray(a *Vec_Int_t) []int32                 { return vec.Vec_IntArray(a) }
func Vec_IntClear(a *Vec_Int_t)                         { vec.Vec_IntClear(a) }
func Vec_IntCountPositive(a *Vec_Int_t) int32           { return int32(vec.Vec_IntCountPositive(a)) }
func Vec_IntErase(a *Vec_Int_t)                         { vec.Vec_IntErase(a) }
func Vec_IntFindMax(a *Vec_Int_t) int32                 { return vec.Vec_IntFindMax(a) }
func Vec_IntFree(a *Vec_Int_t)                          { vec.Vec_IntFree(a) }
func Vec_IntGrow(a *Vec_Int_t, n int32)                 { vec.Vec_IntGrow(a, int(n)) }
func Vec_IntInvert(a *Vec_Int_t, fill int32) *Vec_Int_t { return vec.Vec_IntInvert(a, fill) }
func Vec_IntPrint(a *Vec_Int_t)                         { vec.Vec_IntPrint(a) }
func Vec_IntPush(a *Vec_Int_t, x int32)                 { vec.Vec_IntPush(a, x) }
func Vec_IntPushTwo(a *Vec_Int_t, x, y int32)           { vec.Vec_IntPushTwo(a, x, y) }
func Vec_IntSort(a *Vec_Int_t, reverse int32)           { vec.Vec_IntSort(a, int(reverse)) }
func Vec_IntSum(a *Vec_Int_t) int32                     { return vec.Vec_IntSum(a) }
func Vec_IntTwoCountCommon(a, b *Vec_Int_t) int32       { return int32(vec.Vec_IntTwoCountCommon(a, b)) }
func Hsh_VecManStart(n int32) *Hsh_VecMan_t             { return vec.Hsh_VecManStart(int(n)) }
func Hsh_VecManStop(p *Hsh_VecMan_t)                    { vec.Hsh_VecManStop(p) }
func Hsh_VecManAdd(p *Hsh_VecMan_t, v *Vec_Int_t) int32 { return vec.Hsh_VecManAdd(p, v) }
func Hsh_VecSize(p *Hsh_VecMan_t) int32                 { return int32(vec.Hsh_VecSize(p)) }
