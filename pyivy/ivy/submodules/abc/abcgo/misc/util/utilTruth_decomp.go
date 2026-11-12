// LUT evaluation, AND bi-decomposition and symmetry algorithms from utilTruth.h.
package util

import (
	"fmt"
	"math/bits"
)

func Abc_TtEvalLut6(ins []word, lut word, n int32) word {
	var result word
	for k := int32(0); k < 1<<n; k++ {
		if lut>>k&1 == 0 {
			continue
		}
		cube := ^word(0)
		for i := int32(0); i < n; i++ {
			if k>>i&1 != 0 {
				cube &= ins[i]
			} else {
				cube &= ^ins[i]
			}
		}
		result |= cube
	}
	return result
}
func Abc_TtEvalLut5(ins []uint32, lut, n int32) uint32 {
	in := make([]word, len(ins))
	for i, x := range ins {
		in[i] = word(x)
	}
	return uint32(Abc_TtEvalLut6(in, word(uint32(lut)), n))
}
func Abc_TtEvalLut4(ins []int32, lut, n int32) int32 {
	in := make([]word, len(ins))
	for i, x := range ins {
		in[i] = word(uint32(x))
	}
	return int32(Abc_TtEvalLut6(in, word(uint32(lut)), n) & Abc_Tt6Mask(1<<n))
}
func Abc_TtComputeGraph(t []word, v, n int32, graph []int32) {
	words := Abc_TtWordNum(n)
	c0, c1 := make([]word, words), make([]word, words)
	c00, c01, c10, c11 := make([]word, words), make([]word, words), make([]word, words), make([]word, words)
	graph[v] |= 1 << v
	if v == n-1 {
		return
	}
	if v >= n {
		panic("truth graph variable")
	}
	Abc_TtCofactor0p(c0, t, words, v)
	Abc_TtCofactor1p(c1, t, words, v)
	for i := v + 1; i < n; i++ {
		Abc_TtCofactor0p(c00, c0, words, i)
		Abc_TtCofactor1p(c01, c0, words, i)
		Abc_TtCofactor0p(c10, c1, words, i)
		Abc_TtCofactor1p(c11, c1, words, i)
		for w := int32(0); w < words; w++ {
			x := c00[w] ^ c01[w] ^ c10[w] ^ c11[w]
			a := (c00[w] & c01[w]) | (c10[w] & c11[w])
			if x&a != 0 {
				graph[v] |= 1 << i
				graph[i] |= 1 << v
			} else if x&^a != 0 {
				graph[v] |= 1 << (16 + i)
				graph[i] |= 1 << (16 + v)
			}
		}
	}
}
func Abc_TtPrintVarSet(mask, n int32) {
	for i := int32(0); i < n; i++ {
		if mask>>i&1 != 0 {
			fmt.Print("1")
		} else {
			fmt.Print(".")
		}
	}
}
func Abc_TtPrintBiDec(t []word, n int32) {
	if n > 12 {
		panic("truth bidecomposition variable limit")
	}
	graph := make([]int32, n)
	for v := int32(0); v < n; v++ {
		Abc_TtComputeGraph(t, v, n, graph)
		Abc_TtPrintVarSet(graph[v], n)
		fmt.Print("    ")
		Abc_TtPrintVarSet(graph[v]>>16, n)
		fmt.Println()
	}
}
func Abc_TtVerifyBiDec(t []word, n, this, that, limit int32, wThis, wThat word) int32 {
	// Map packed component inputs directly into the original variable positions.
	for assignment := int32(0); assignment < 1<<n; assignment++ {
		var a, b, ai, bi int32
		for v := int32(0); v < n; v++ {
			if this>>v&1 != 0 {
				a |= ((assignment >> v) & 1) << ai
				ai++
			}
			if that>>v&1 != 0 {
				b |= ((assignment >> v) & 1) << bi
				bi++
			}
		}
		value := int32((wThis >> a) & (wThat >> b) & 1)
		if Abc_TtGetBit(t, assignment) != value {
			return 0
		}
	}
	return 1
}
func Abc_TtExist(t []word, v, words int32) {
	a, b := make([]word, words), make([]word, words)
	Abc_TtCofactor0p(a, t, words, v)
	Abc_TtCofactor1p(b, t, words, v)
	Abc_TtOr(t, a, b, words)
}
func Abc_TtCheckBiDec(t []word, n, this, that int32) int32 {
	words := Abc_TtWordNum(n)
	temp := [2][]word{append([]word(nil), t[:words]...), append([]word(nil), t[:words]...)}
	masks := [2]int32{this &^ that, that &^ this}
	for c := 0; c < 2; c++ {
		for v := int32(0); v < n; v++ {
			if masks[c]>>v&1 != 0 {
				Abc_TtExist(temp[c], v, words)
			}
		}
	}
	for w := int32(0); w < words; w++ {
		if ^t[w]&temp[0][w]&temp[1][w] != 0 {
			return 0
		}
	}
	return 1
}
func Abc_TtDeriveBiDecOne(t []word, n, this int32) word {
	// Existentially quantify all variables outside the selected support, then
	// encode the remaining variables in their original increasing order.
	vars := make([]int32, 0, n)
	for v := int32(0); v < n; v++ {
		if this>>v&1 != 0 {
			vars = append(vars, v)
		}
	}
	if len(vars) > 6 {
		panic("truth component exceeds one word")
	}
	var result word
	for assignment := int32(0); assignment < 1<<n; assignment++ {
		if Abc_TtGetBit(t, assignment) == 0 {
			continue
		}
		var packed int32
		for i, v := range vars {
			packed |= ((assignment >> v) & 1) << i
		}
		result |= 1 << packed
	}
	if len(vars) == 6 {
		return result
	}
	period := int32(1) << len(vars)
	for bit := period; bit < 64; bit++ {
		result |= ((result >> (bit % period)) & 1) << bit
	}
	return result
}
func Abc_TtDeriveBiDec(t []word, n, this, that, limit int32, outThis, outThat []word) {
	if Abc_TtBitCount16(this) > limit || Abc_TtBitCount16(that) > limit {
		panic("truth support limit")
	}
	outThis[0] = Abc_TtDeriveBiDecOne(t, n, this)
	outThat[0] = Abc_TtDeriveBiDecOne(t, n, that)
	if Abc_TtVerifyBiDec(t, n, this, that, limit, outThis[0], outThat[0]) == 0 {
		fmt.Println("Bi-decomposition verification failed.")
	}
}
func Abc_TtCheckBiDecSimple(t []word, n, limit int32) int32 {
	words := Abc_TtWordNum(n)
	a, b := make([]word, words), make([]word, words)
	var result, count int32
	for v := int32(0); v < n; v++ {
		Abc_TtCofactor0p(a, t, words, v)
		Abc_TtCofactor1p(b, t, words, v)
		if Abc_TtIsConst0(a, words) == 0 && Abc_TtIsConst0(b, words) == 0 {
			continue
		}
		count++
		result |= 1 << v
		if count >= n-limit {
			return ((result ^ int32(Abc_Tt6Mask(n))) << 16) | result
		}
	}
	return 0
}
func Abc_TtProcessBiDecInt(t []word, n, limit int32) int32 {
	if !(limit < n && n <= 2*limit && n <= 12 && 2 <= limit && limit <= 6) {
		panic("truth bidecomposition limits")
	}
	if r := Abc_TtCheckBiDecSimple(t, n, limit); r != 0 {
		return r
	}
	graph := make([]int32, n)
	var shared int32
	for v := int32(0); v < n; v++ {
		Abc_TtComputeGraph(t, v, n, graph)
		support := Abc_TtBitCount16(graph[v] & 65535)
		if support > limit {
			shared++
			if shared > 2*limit-n {
				return 0
			}
		} else if n-support <= limit {
			this := graph[v] & 65535
			that := this ^ int32(Abc_Tt6Mask(n))
			g := that
			for i := int32(0); i < n; i++ {
				if that>>i&1 != 0 {
					g |= graph[i] & 65535
				}
			}
			if Abc_TtBitCount16(g) > limit {
				continue
			}
			if Abc_TtCheckBiDec(t, n, this, g) != 0 {
				return (g << 16) | this
			}
		}
	}
	return 0
}
func Abc_TtProcessBiDec(t []word, n, limit int32) int32 {
	words := Abc_TtWordNum(n)
	tmp := append([]word(nil), t[:words]...)
	if r := Abc_TtProcessBiDecInt(tmp, n, limit); r != 0 {
		return r
	}
	Abc_TtNot(tmp, words)
	if r := Abc_TtProcessBiDecInt(tmp, n, limit); r != 0 {
		return r | (1 << 30)
	}
	return 0
}
func Abc_TtProcessBiDecTest(t []word, n, limit int32) {
	r := Abc_TtProcessBiDec(t, n, limit)
	if r == 0 {
		return
	}
	tmp := append([]word(nil), t...)
	if r&(1<<30) != 0 {
		Abc_TtNot(tmp, Abc_TtWordNum(n))
	}
	a, b := make([]word, 1), make([]word, 1)
	Abc_TtDeriveBiDec(tmp, n, r&65535, (r>>16)&65535, limit, a, b)
}
func Abc_TtProcessBiDecExperiment() {
	t := []word{(s_Truths6[0] | s_Truths6[1]) & (s_Truths6[1] | s_Truths6[2])}
	Abc_TtPrintBiDec(t, 3)
	r := Abc_TtProcessBiDec(t, 3, 2)
	a, b := make([]word, 1), make([]word, 1)
	Abc_TtDeriveBiDec(t, 3, r&65535, r>>16, 2, a, b)
}
func Abc_Tt4Equal3(a, b, c, d int32) int32 {
	if a == b && a == c {
		return 3
	}
	if a == b && a == d {
		return 2
	}
	if a == d && a == c {
		return 1
	}
	if d == b && d == c {
		return 0
	}
	return -1
}
func Abc_Tt4Check2(t, i, j int32, f, r []int32) int32 {
	c0 := t & r[j]
	c1 := (t & f[j]) >> (1 << j)
	return Abc_Tt4Equal3(c0&r[i], (c0&f[i])>>(1<<i), c1&r[i], (c1&f[i])>>(1<<i))
}
func Abc_Tt4CheckTwoLevel(t int32) int32 {
	f := []int32{0xAAAA, 0xCCCC, 0xF0F0, 0xFF00}
	r := []int32{0x5555, 0x3333, 0x0F0F, 0x00FF}
	for k, pair := range [][4]int32{{0, 1, 2, 3}, {0, 2, 1, 3}, {0, 3, 1, 2}} {
		a := Abc_Tt4Check2(t, pair[0], pair[1], f, r)
		if a < 0 {
			continue
		}
		b := Abc_Tt4Check2(t, pair[2], pair[3], f, r)
		if b >= 0 {
			return int32(k+1)<<4 | b<<2 | a
		}
	}
	return -1
}
func Abc_Tt4CountOnes(t int32) int32 { return int32(bits.OnesCount16(uint16(t))) }
func Abc_Tt4FirstBit(t int32) int32 {
	if t == 0 {
		return -1
	}
	return int32(bits.TrailingZeros16(uint16(t)))
}
func Abc_Tt4Check(t int32) int32 {
	neg := 65535 &^ t
	if t == 0x6996 || neg == 0x6996 || t&(t-1) == 0 || neg&(neg-1) == 0 {
		return 1
	}
	n := Abc_Tt4CountOnes(t)
	return ttBool((n == 7 && Abc_Tt4CheckTwoLevel(t) > 0) || (n == 9 && Abc_Tt4CheckTwoLevel(neg) > 0))
}
func Abc_Tt6VarsAreSymmetric(t word, i, j int32) int32 {
	if i >= j {
		panic("truth variable order")
	}
	m := s_PPMasks[i][j]
	return ttBool((t&m[1])<<((1<<j)-(1<<i)) == t&m[2])
}
func Abc_Tt6VarsAreAntiSymmetric(t word, i, j int32) int32 {
	if i >= j {
		panic("truth variable order")
	}
	m := s_PPMasks[i][j]
	return ttBool((t&(m[1]>>(1<<i)))<<((1<<j)+(1<<i)) == t&(m[2]<<(1<<i)))
}
func Abc_TtVarsAreSymmetric(t []word, n, i, j int32, c0, c1 []word) int32 {
	words := Abc_TtWordNum(n)
	Abc_TtCofactor0p(c0, t, words, i)
	Abc_TtCofactor1p(c1, t, words, i)
	Abc_TtCofactor1(c0, words, j)
	Abc_TtCofactor0(c1, words, j)
	return Abc_TtEqual(c0, c1, words)
}
func Abc_TtVarsAreAntiSymmetric(t []word, n, i, j int32, c0, c1 []word) int32 {
	words := Abc_TtWordNum(n)
	Abc_TtCofactor0p(c0, t, words, i)
	Abc_TtCofactor1p(c1, t, words, i)
	Abc_TtCofactor0(c0, words, j)
	Abc_TtCofactor1(c1, words, j)
	return Abc_TtEqual(c0, c1, words)
}
func Abc_TtIsFullySymmetric(t []word, n int32) int32 {
	var polar, seen int32
	for m := int32(0); m < 1<<n; m++ {
		count := bits.OnesCount32(uint32(m))
		value := Abc_TtGetBit(t, m)
		if seen>>count&1 != 0 {
			if value != (polar>>count)&1 {
				return -1
			}
		} else {
			seen |= 1 << count
			if value != 0 {
				polar |= 1 << count
			}
		}
	}
	return polar
}
func Abc_TtGenFullySymmetric(t []word, n, polar int32) {
	Abc_TtClear(t, Abc_TtWordNum(n))
	for m := int32(0); m < 1<<n; m++ {
		if polar>>bits.OnesCount32(uint32(m))&1 != 0 {
			Abc_TtSetBit(t, m)
		}
	}
}
func Abc_TtTestFullySymmetric() {
	t := make([]word, 4)
	Abc_TtGenFullySymmetric(t, 8, 271)
	if Abc_TtIsFullySymmetric(t, 8) != 271 {
		panic("truth symmetry self-test")
	}
}
func Abc_TtSymFunGenerate(ones string, n int32) []word {
	if len(ones) != int(n+1) {
		panic("truth symmetry input length")
	}
	t := make([]word, Abc_TtWordNum(n))
	for m := int32(0); m < 1<<n; m++ {
		if ones[bits.OnesCount32(uint32(m))] == '1' {
			Abc_TtXorBit(t, m)
		}
	}
	return t
}

func Unm_ManCheckTest2() {
	for a := int32(0); a < 6; a++ {
		for b := int32(0); b < 6; b++ {
			if a == b {
				continue
			}
			for phase := int32(0); phase < 5; phase++ {
				x, y := ttCompl(s_Truths6[a], phase&1), ttCompl(s_Truths6[b], phase>>1&1)
				t := x & y
				if phase == 4 {
					t = ^(x ^ y)
				}
				out := make([]word, 1)
				r := Abc_TtCheckDsdAnd(t, a, b, out)
				if r < 0 {
					fmt.Println("No decomposition")
					continue
				}
				x, y = ttCompl(s_Truths6[a], r&1), ttCompl(s_Truths6[b], r>>1&1)
				g := x & y
				if r == 4 {
					g = x ^ y
				}
				actual := (^g & Abc_Tt6Cofactor0(out[0], a)) | (g & Abc_Tt6Cofactor1(out[0], a))
				if actual == t {
					fmt.Println("Verification succeeded.")
				} else {
					fmt.Println("Verification failed.")
				}
			}
		}
	}
}
func Unm_ManCheckTest() {
	for c := int32(0); c < 6; c++ {
		for a := int32(0); a < 6; a++ {
			for b := int32(0); b < 6; b++ {
				if a == b || a == c || b == c {
					continue
				}
				for phase := int32(0); phase < 8; phase++ {
					ctrl, x, y := ttCompl(s_Truths6[c], phase&1), ttCompl(s_Truths6[a], phase>>1&1), ttCompl(s_Truths6[b], phase>>2&1)
					t := (^ctrl & x) | (ctrl & y)
					out := make([]word, 1)
					r := Abc_TtCheckDsdMux(t, c, out)
					if r < 0 {
						fmt.Println("No decomposition")
						continue
					}
					lo, hi := r&65535, (r>>16)&65535
					x = ttCompl(s_Truths6[lo>>1], lo&1)
					y = ttCompl(s_Truths6[hi>>1], hi&1)
					g := (^s_Truths6[c] & x) | (s_Truths6[c] & y)
					actual := (^g & Abc_Tt6Cofactor0(out[0], c)) | (g & Abc_Tt6Cofactor1(out[0], c))
					if actual == t {
						fmt.Println("Verification succeeded.")
					} else {
						fmt.Println("Verification failed.")
					}
				}
			}
		}
	}
}
