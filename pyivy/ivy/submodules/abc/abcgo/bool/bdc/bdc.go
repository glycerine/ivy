// Package bdc implements truth-table based AND/OR bi-decomposition with
// incompletely specified functions, cached divisors and Shannon fallback.
package bdc

type Bdc_Par_t struct{ NVarsMax, FVerbose, FVeryVerbose int32 }
type Bdc_Type_t = int32

const (
	BDC_TYPE_NONE int32 = iota
	BDC_TYPE_CONST1
	BDC_TYPE_PI
	BDC_TYPE_AND
	BDC_TYPE_OR
	BDC_TYPE_XOR
	BDC_TYPE_MUX
	BDC_TYPE_OTHER
)
const BDC_SCALE int32 = 1000

type Bdc_Fun_t struct {
	Type              int32
	Fan0, Fan1        *Bdc_Fun_t
	Supp              uint32
	Func              []uint32
	Copy              any
	CopyInt           int32
	id                int32
	regular, negative *Bdc_Fun_t
}

// Complemented references use Go objects, never tagged/unsafe pointers.
func Bdc_IsComplement(f *Bdc_Fun_t) int32 {
	if f != nil && f.regular != nil {
		return 1
	}
	return 0
}
func Bdc_Regular(f *Bdc_Fun_t) *Bdc_Fun_t {
	if f != nil && f.regular != nil {
		return f.regular
	}
	return f
}
func Bdc_Not(f *Bdc_Fun_t) *Bdc_Fun_t {
	if f == nil {
		panic("bdc: complement of nil")
	}
	if f.regular != nil {
		return f.regular
	}
	if f.negative == nil {
		f.negative = &Bdc_Fun_t{regular: f}
	}
	return f.negative
}
func Bdc_NotCond(f *Bdc_Fun_t, c int32) *Bdc_Fun_t {
	if c != 0 {
		return Bdc_Not(f)
	}
	return f
}
func Bdc_FuncFanin0(f *Bdc_Fun_t) *Bdc_Fun_t { return Bdc_Regular(f).Fan0 }
func Bdc_FuncFanin1(f *Bdc_Fun_t) *Bdc_Fun_t { return Bdc_Regular(f).Fan1 }
func Bdc_FuncCopy(f *Bdc_Fun_t) any          { return Bdc_Regular(f).Copy }
func Bdc_FuncCopyInt(f *Bdc_Fun_t) int32     { return Bdc_Regular(f).CopyInt }
func Bdc_FuncSetCopy(f *Bdc_Fun_t, c any)    { node := Bdc_Regular(f); node.Copy = c; node.CopyInt = 0 }
func Bdc_FuncSetCopyInt(f *Bdc_Fun_t, c int32) {
	node := Bdc_Regular(f)
	node.CopyInt = c
	node.Copy = nil
}
func Bdc_FunObjCopy(f *Bdc_Fun_t) int32    { return Bdc_FuncCopyInt(f) ^ Bdc_IsComplement(f) }
func Bdc_FunFanin0Copy(f *Bdc_Fun_t) int32 { return Bdc_FunObjCopy(Bdc_FuncFanin0(f)) }
func Bdc_FunFanin1Copy(f *Bdc_Fun_t) int32 { return Bdc_FunObjCopy(Bdc_FuncFanin1(f)) }
