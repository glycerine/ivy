// Core portable primitives from ABC abc_global.h.
package util

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

type ABC_PTRDIFF_T = int
type ABC_PTRINT_T = int
type ABC_PTRUINT_T = uintptr
type ABC_INT64_T = int64
type ABC_UINT64_T = uint64
type word = uint64
type Word = uint64
type iword = int64
type abctime = int64
type AbcTime = int64

const ABC_INFINITY = 1000000000
const CLOCKS_PER_SEC = int64(time.Second)

type Abc_VerbLevel = int32

const (
	ABC_PROMPT   = -2
	ABC_ERROR    = -1
	ABC_WARNING  = 0
	ABC_STANDARD = 1
	ABC_VERBOSE  = 2
)

func Abc_AbsInt(a int32) int32 {
	if a < 0 {
		return -a
	}
	return a
}
func Abc_MaxInt(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
func Abc_MinInt(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func Abc_MaxWord(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
func Abc_MinWord(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
func Abc_AbsFloat(a float32) float32 {
	if a < 0 {
		return -a
	}
	return a
}
func Abc_MaxFloat(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
func Abc_MinFloat(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
func Abc_AbsDouble(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}
func Abc_MaxDouble(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
func Abc_MinDouble(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func Abc_Float2Int(v float32) int32 { return int32(math.Float32bits(v)) }
func Abc_Int2Float(v int32) float32 { return math.Float32frombits(uint32(v)) }
func Abc_Dbl2Word(v float64) uint64 { return math.Float64bits(v) }
func Abc_Word2Dbl(v uint64) float64 { return math.Float64frombits(v) }
func abcBaseLog(n, base uint32) int32 {
	if n < 2 {
		return int32(n)
	}
	var r int32
	for n--; n != 0; n /= base {
		r++
	}
	return r
}
func Abc_Base2Log(n uint32) int32    { return abcBaseLog(n, 2) }
func Abc_Base10Log(n uint32) int32   { return abcBaseLog(n, 10) }
func Abc_Base16Log(n uint32) int32   { return abcBaseLog(n, 16) }
func Abc_UtilStrsav(s string) string { return strings.Clone(s) }
func boolToInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func Abc_BitByteNum(n int32) int32  { return (n >> 3) + boolToInt(n&7 != 0) }
func Abc_BitWordNum(n int32) int32  { return (n >> 5) + boolToInt(n&31 != 0) }
func Abc_Bit6WordNum(n int32) int32 { return (n >> 6) + boolToInt(n&63 != 0) }
func Abc_TruthByteNum(n int32) int32 {
	if n <= 3 {
		return 1
	}
	return 1 << (n - 3)
}
func Abc_TruthWordNum(n int32) int32 {
	if n <= 5 {
		return 1
	}
	return 1 << (n - 5)
}
func Abc_Truth6WordNum(n int32) int32 {
	if n <= 6 {
		return 1
	}
	return 1 << (n - 6)
}
func Abc_InfoHasBit(p []uint32, i int32) bool { return p[i>>5]&(1<<uint32(i&31)) != 0 }
func Abc_InfoSetBit(p []uint32, i int32)      { p[i>>5] |= 1 << uint32(i&31) }
func Abc_InfoXorBit(p []uint32, i int32)      { p[i>>5] ^= 1 << uint32(i&31) }
func Abc_InfoMask(n int32) uint32             { return ^uint32(0) >> uint32(32-n) }
func Abc_Var2Lit(v, c int32) int32            { assert(v >= 0 && c>>1 == 0, "literal range"); return v*2 + c }
func Abc_Lit2Var(l int32) int32               { assert(l >= 0, "literal nonnegative"); return l >> 1 }
func Abc_LitIsCompl(l int32) int32            { assert(l >= 0, "literal nonnegative"); return l & 1 }
func Abc_LitNot(l int32) int32                { assert(l >= 0, "literal nonnegative"); return l ^ 1 }
func Abc_LitNotCond(l, c int32) int32 {
	assert(l >= 0, "literal nonnegative")
	if c > 0 {
		return l ^ 1
	}
	return l
}
func Abc_LitRegular(l int32) int32          { assert(l >= 0, "literal nonnegative"); return l &^ 1 }
func Abc_Lit2LitV(m []int32, l int32) int32 { return Abc_Var2Lit(m[Abc_Lit2Var(l)], Abc_LitIsCompl(l)) }
func Abc_Lit2LitL(m []int32, l int32) int32 {
	return Abc_LitNotCond(m[Abc_Lit2Var(l)], Abc_LitIsCompl(l))
}

// AbcPointerValue represents the integer payloads historically smuggled through
// C void pointers. It is an ID/tagged value, never a Go heap address.
type AbcPointerValue uint64

func Abc_Ptr2Int(p AbcPointerValue) int32  { return int32(p) }
func Abc_Int2Ptr(i int32) AbcPointerValue  { return AbcPointerValue(uint64(int64(i))) }
func Abc_Ptr2Wrd(p AbcPointerValue) uint64 { return uint64(p) }
func Abc_Wrd2Ptr(i uint64) AbcPointerValue { return AbcPointerValue(i) }
func Abc_Var2Lit2(v, a int32) int32        { assert(a>>2 == 0, "attribute range"); return v<<2 | a }
func Abc_Lit2Var2(l int32) int32           { assert(l >= 0, "literal nonnegative"); return l >> 2 }
func Abc_Lit2Att2(l int32) int32           { assert(l >= 0, "literal nonnegative"); return l & 3 }
func Abc_Var2Lit3(v, a int32) int32        { assert(a>>3 == 0, "attribute range"); return v<<3 | a }
func Abc_Lit2Var3(l int32) int32           { assert(l >= 0, "literal nonnegative"); return l >> 3 }
func Abc_Lit2Att3(l int32) int32           { assert(l >= 0, "literal nonnegative"); return l & 7 }
func Abc_Var2Lit4(v, a int32) int32        { assert(a>>4 == 0, "attribute range"); return v<<4 | a }
func Abc_Lit2Var4(l int32) int32           { assert(l >= 0, "literal nonnegative"); return l >> 4 }
func Abc_Lit2Att4(l int32) int32           { assert(l >= 0, "literal nonnegative"); return l & 15 }

var abcClockStart = time.Now()

// Abc_Clock reports monotonic elapsed nanoseconds. Unlike the platform-specific
// C thread CPU clock, this portable clock includes time spent blocked.
func Abc_Clock() int64 { return time.Since(abcClockStart).Nanoseconds() }

// PrintBridge lets the command layer route output into ABC's bridge protocol.
var PrintBridge func(string)

func Abc_Print(level int32, format string, args ...any) {
	prefix := ""
	if level == ABC_ERROR {
		prefix = "Error: "
	} else if level == ABC_WARNING {
		prefix = "Warning: "
	}
	text := prefix + fmt.Sprintf(format, args...)
	if PrintBridge != nil {
		PrintBridge(text)
	} else {
		fmt.Print(text)
	}
}
func Abc_PrintInt(i int32) {
	Abc_Print(1, "  ")
	v3, v6 := float64(i)/1000, float64(i)/1000000
	switch {
	case i > -1000 && i < 1000:
		Abc_Print(1, " %4d", i)
	case v3 > -9.995 && v3 < 9.995:
		Abc_Print(1, "%4.2fk", v3)
	case v3 > -99.95 && v3 < 99.95:
		Abc_Print(1, "%4.1fk", v3)
	case v3 > -999.5 && v3 < 999.5:
		Abc_Print(1, "%4.0fk", v3)
	case v6 > -9.995 && v6 < 9.995:
		Abc_Print(1, "%4.2fm", v6)
	case v6 > -99.95 && v6 < 99.95:
		Abc_Print(1, "%4.1fm", v6)
	case v6 > -999.5 && v6 < 999.5:
		Abc_Print(1, "%4.0fm", v6)
	}
}
func Abc_PrintTime(level int32, s string, t int64) {
	Abc_Print(level, "%s =%9.2f sec\n", s, float64(t)/float64(time.Second))
}
func Abc_PrintTimeP(level int32, s string, t, total int64) {
	var percent float64
	if total != 0 {
		percent = 100 * float64(t) / float64(total)
	}
	Abc_Print(level, "%s =%9.2f sec (%6.2f %%)\n", s, float64(t)/float64(time.Second), percent)
}
func Abc_PrintMemoryP(level int32, s string, mem, total int32) {
	var percent float64
	if total != 0 {
		percent = 100 * float64(mem) / float64(total)
	}
	Abc_Print(level, "%s =%10.3f MB (%6.2f %%)\n", s, float64(mem)/(1<<20), percent)
}

// Match Cudd's odd-prime selection, including returning one for inputs <=1.
func Abc_PrimeCudd(p uint32) int32 {
	for {
		if p&1 != 0 {
			prime := true
			for i := uint32(3); uint64(i)*uint64(i) <= uint64(p); i += 2 {
				if p%i == 0 {
					prime = false
					break
				}
			}
			if prime {
				return int32(p)
			}
		}
		p++
	}
}
func Abc_FileReadContents(name string, size *int32) []byte {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil
	}
	if size != nil {
		*size = int32(len(data))
	}
	return append(data, make([]byte, 32)...)
}
func Abc_ReverseOrder(a []int32, n int32) {
	a = a[:n]
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}
