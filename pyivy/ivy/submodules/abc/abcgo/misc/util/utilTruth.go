// Truth table operations ported from src/misc/util/utilTruth.h.
package util

import (
	"math/bits"
)

var s_Truths6 = [6]word{0xAAAAAAAAAAAAAAAA, 0xCCCCCCCCCCCCCCCC, 0xF0F0F0F0F0F0F0F0, 0xFF00FF00FF00FF00, 0xFFFF0000FFFF0000, 0xFFFFFFFF00000000}
var s_Truths6Neg = func() (r [6]word) {
	for i, x := range s_Truths6 {
		r[i] = ^x
	}
	return
}()
var s_TruthXors = [6]word{0, 0x6666666666666666, 0x6969696969696969, 0x6996699669966996, 0x6996966969969669, 0x6996966996696996}
var s_PPMasks = func() (r [5][6][3]word) {
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 6; j++ {
			a, b := s_Truths6[i], s_Truths6[j]
			r[i][j] = [3]word{^(a ^ b), a &^ b, ^a & b}
		}
	}
	return
}()
var s_PMasks = func() (r [5][3]word) {
	for i := range r {
		r[i] = s_PPMasks[i][i+1]
	}
	return
}()
var Abc_TtBitCount8 = func() (r [256]int32) {
	for i := range r {
		r[i] = int32(bits.OnesCount8(uint8(i)))
	}
	return
}()

func Abc_TtBitCount16(i int32) int32 { return Abc_TtBitCount8[i&255] + Abc_TtBitCount8[(i>>8)&255] }
func ttCompl(x word, c int32) word {
	if c != 0 {
		return ^x
	}
	return x
}
func ttBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Abc_TtGetBit(p []word, i int32) int32    { k := i << 0; return int32(p[k>>6]>>(k&63)) & 1 }
func Abc_TtSetBit(p []word, i int32)          { k := i << 0; p[k>>6] |= word(1) << (k & 63) }
func Abc_TtXorBit(p []word, i int32)          { k := i << 0; p[k>>6] ^= word(1) << (k & 63) }
func Abc_TtGetQua(p []word, i int32) int32    { k := i << 1; return int32(p[k>>6]>>(k&63)) & 3 }
func Abc_TtSetQua(p []word, i int32, d int32) { k := i << 1; p[k>>6] |= word(d) << (k & 63) }
func Abc_TtXorQua(p []word, i int32, d int32) { k := i << 1; p[k>>6] ^= word(d) << (k & 63) }
func Abc_TtGetHex(p []word, i int32) int32    { k := i << 2; return int32(p[k>>6]>>(k&63)) & 15 }
func Abc_TtSetHex(p []word, i int32, d int32) { k := i << 2; p[k>>6] |= word(d) << (k & 63) }
func Abc_TtXorHex(p []word, i int32, d int32) { k := i << 2; p[k>>6] ^= word(d) << (k & 63) }
func Abc_TtGet256(p []word, i int32) int32    { k := i << 3; return int32(p[k>>6]>>(k&63)) & 255 }
func Abc_TtSet256(p []word, i int32, d int32) { k := i << 3; p[k>>6] |= word(d) << (k & 63) }
func Abc_TtXor256(p []word, i int32, d int32) { k := i << 3; p[k>>6] ^= word(d) << (k & 63) }
func Abc_TtWordNum(n int32) int32 {
	if n <= 6 {
		return 1
	}
	return 1 << (n - 6)
}
func Abc_TtByteNum(n int32) int32 {
	if n <= 3 {
		return 1
	}
	return 1 << (n - 3)
}
func Abc_TtHexDigitNum(n int32) int32 {
	if n <= 2 {
		return 1
	}
	return 1 << (n - 2)
}
func Abc_Tt6Mask(n int32) word {
	if n < 0 || n > 64 {
		panic("truth mask size")
	}
	if n == 64 {
		return ^word(0)
	}
	return (word(1) << n) - 1
}
func Abc_TtMask(p []word, nWords, nBits int32) {
	if nBits < 0 || nBits > nWords*64 {
		panic("truth mask size")
	}
	for i := int32(0); i < nWords; i++ {
		n := nBits - i*64
		if n < 0 {
			n = 0
		}
		if n > 64 {
			n = 64
		}
		p[i] = Abc_Tt6Mask(n)
	}
}
func Abc_TtVec(out []word, n int32, x word) {
	for i := int32(0); i < n; i++ {
		out[i] = x
	}
}
func Abc_TtConst(out []word, n, c int32) { Abc_TtVec(out, n, ttCompl(0, c)) }
func Abc_TtClear(out []word, n int32)    { Abc_TtVec(out, n, 0) }
func Abc_TtFill(out []word, n int32)     { Abc_TtVec(out, n, ^word(0)) }
func Abc_TtUnit(out []word, n, c int32)  { Abc_TtVec(out, n, ttCompl(s_Truths6[0], c)) }
func Abc_TtNot(out []word, n int32) {
	for i := int32(0); i < n; i++ {
		out[i] = ^out[i]
	}
}
func Abc_TtCopy(out, in []word, n, c int32) {
	for i := int32(0); i < n; i++ {
		out[i] = ttCompl(in[i], c)
	}
}
func Abc_TtAnd(out, a, b []word, n int32, c int32) {
	for i := int32(0); i < n; i++ {
		out[i] = ttCompl(a[i]&b[i], c)
	}
}
func Abc_TtAndSharp(out, a, b []word, n int32, c int32) {
	for i := int32(0); i < n; i++ {
		out[i] = a[i] & ttCompl(b[i], c)
	}
}
func Abc_TtSharp(out, a, b []word, n int32) {
	for i := int32(0); i < n; i++ {
		out[i] = a[i] &^ b[i]
	}
}
func Abc_TtOr(out, a, b []word, n int32) {
	for i := int32(0); i < n; i++ {
		out[i] = a[i] | b[i]
	}
}
func Abc_TtOrXor(out, a, b []word, n int32) {
	for i := int32(0); i < n; i++ {
		out[i] = out[i] | (a[i] ^ b[i])
	}
}
func Abc_TtOrAnd(out, a, b []word, n int32) {
	for i := int32(0); i < n; i++ {
		out[i] = out[i] | (a[i] & b[i])
	}
}
func Abc_TtXor(out, a, b []word, n int32, c int32) {
	for i := int32(0); i < n; i++ {
		out[i] = ttCompl(a[i]^b[i], c)
	}
}
func Abc_TtAndCompl(out, a []word, ca int32, b []word, cb, n int32) {
	for i := int32(0); i < n; i++ {
		out[i] = ttCompl(a[i], ca) & ttCompl(b[i], cb)
	}
}
func Abc_TtXorMask(out, a, b, mask []word, n, c int32) {
	for i := int32(0); i < n; i++ {
		out[i] = (a[i] ^ b[i]) & ttCompl(mask[i], c)
	}
}
func Abc_TtMux(out, ctrl, a, b []word, n int32) {
	for i := int32(0); i < n; i++ {
		out[i] = (ctrl[i] & a[i]) | (^ctrl[i] & b[i])
	}
}
func Abc_TtMaj(out, a, b, c []word, n int32) {
	for i := int32(0); i < n; i++ {
		out[i] = (a[i] & b[i]) | (a[i] & c[i]) | (b[i] & c[i])
	}
}
func Abc_TtIntersect(a, b []word, n, c int32) int32 {
	for i := int32(0); i < n; i++ {
		if ttCompl(a[i], c)&b[i] != 0 {
			return 1
		}
	}
	return 0
}
func Abc_TtIntersectOne(out []word, c int32, a []word, ca, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if ttCompl(out[i], c)&ttCompl(a[i], ca) != 0 {
			return 1
		}
	}
	return 0
}
func Abc_TtIntersectTwo(out []word, c int32, a []word, ca int32, b []word, cb, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if ttCompl(out[i], c)&ttCompl(a[i], ca)&ttCompl(b[i], cb) != 0 {
			return 1
		}
	}
	return 0
}
func Abc_TtIntersectXor(out []word, c int32, a, b []word, cab, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if ttCompl(out[i], c)&ttCompl(a[i]^b[i], cab) != 0 {
			return 1
		}
	}
	return 0
}
func Abc_TtEqual(a, b []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if a[i] != b[i] {
			return 0
		}
	}
	return 1
}
func Abc_TtOpposite(a, b []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if a[i] != ^b[i] {
			return 0
		}
	}
	return 1
}
func Abc_TtImply(a, b []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if a[i]&^b[i] != 0 {
			return 0
		}
	}
	return 1
}
func Abc_TtCompare(a, b []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
func Abc_TtCompareRev(a, b []word, n int32) int32 {
	for i := n - 1; i >= 0; i-- {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
func Abc_TtIsConst0(a []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if a[i] != 0 {
			return 0
		}
	}
	return 1
}
func Abc_TtConst0(a []word, n int32) { Abc_TtVec(a, n, 0) }
func Abc_TtIsConst1(a []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if a[i] != ^word(0) {
			return 0
		}
	}
	return 1
}
func Abc_TtConst1(a []word, n int32) { Abc_TtVec(a, n, ^word(0)) }
func Abc_TtIthVar(out []word, v, nVars int32) {
	for i := int32(0); i < Abc_TtWordNum(nVars); i++ {
		if v < 6 {
			out[i] = s_Truths6[v]
		} else if (i>>(v-6))&1 != 0 {
			out[i] = ^word(0)
		} else {
			out[i] = 0
		}
	}
}
func Abc_TtIsAndCompl(out []word, c int32, a []word, ca int32, b []word, cb int32, care []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if ttCompl(out[i], c)&care[i] != (ttCompl(a[i], ca) & ttCompl(b[i], cb) & care[i]) {
			return 0
		}
	}
	return 1
}
func Abc_TtIsXorCompl(out []word, c int32, a, b, care []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if ttCompl(out[i], c)&care[i] != (a[i]^b[i])&care[i] {
			return 0
		}
	}
	return 1
}
func ttCompareCofs(t []word, n, v int32, reverse bool) int32 {
	for k := int32(0); k < n; k++ {
		i := k
		if reverse {
			i = n - 1 - k
		}
		var a, b word
		if v < 6 {
			a = t[i] & s_Truths6Neg[v]
			b = (t[i] >> (1 << v)) & s_Truths6Neg[v]
		} else {
			step := int32(1) << (v - 6)
			if i&step != 0 {
				continue
			}
			a = t[i]
			b = t[i+step]
		}
		if a < b {
			return -1
		}
		if a > b {
			return 1
		}
	}
	return 0
}
func Abc_TtCompare1VarCofs(t []word, n, v int32) int32    { return ttCompareCofs(t, n, v, false) }
func Abc_TtCompare1VarCofsRev(t []word, n, v int32) int32 { return ttCompareCofs(t, n, v, true) }
func Abc_TtElemInit(out [][]word, n int32) {
	for i := int32(0); i < n; i++ {
		Abc_TtIthVar(out[i], i, n)
	}
}
func Abc_TtElemInit2(out []word, n int32) {
	words := Abc_TtWordNum(n)
	for i := int32(0); i < n; i++ {
		Abc_TtIthVar(out[i*words:], i, n)
	}
}

/**Function*************************************************************

  Synopsis    [Checks pairs of cofactors w.r.t. two variables.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Stretch truthtable to have more input variables.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reads hex truth table from a string.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks if there is a var whose both cofs have supp <= nSuppLim.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Detecting elementary functions.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

// moves one var (v) to the given position (p)

/**Function*************************************************************

  Synopsis    [Support minimization.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Cut minimization.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Implemeting given NPN config.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks unateness of a function.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computes ISOP for 6 variables or less.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computes CNF size.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Derives ISOP cover for the function.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computes CNF for the function.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Computes ISOP for 6 variables or less.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Check if the function is output-decomposable with the given var.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Check if the function is input-decomposable with the given pair.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Truth table evaluation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Checks existence of bi-decomposition.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

// detect simple case of decomposition with topmost AND gate

/**Function*************************************************************

  Synopsis    [Tests decomposition procedures.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Truth table checking procedure.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns symmetry profile of the function.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Generates truth table of a symmetric function.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

func Abc_Tt6Cofactor0(t word, v int32) word { x := t & s_Truths6Neg[v]; return x | (x << (1 << v)) }
func Abc_Tt6Cofactor1(t word, v int32) word { x := t & s_Truths6[v]; return x | (x >> (1 << v)) }
func Abc_TtCofactor0p(out, in []word, n, v int32) {
	tmp := append([]word(nil), in[:n]...)
	Abc_TtCofactor0(tmp, n, v)
	copy(out, tmp)
}
func Abc_TtCofactor1p(out, in []word, n, v int32) {
	tmp := append([]word(nil), in[:n]...)
	Abc_TtCofactor1(tmp, n, v)
	copy(out, tmp)
}
