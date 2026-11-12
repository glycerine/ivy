package util

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"strings"
)

func ttFlag(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func ttWords(n int32) int32 {
	if n <= 6 {
		return 1
	}
	return 1 << (n - 6)
}
func ttMask(v int32) word {
	if v < 0 || v >= 6 {
		panic("truth: invalid small variable")
	}
	var m word
	for i := 0; i < 64; i++ {
		if i&(1<<v) != 0 {
			m |= word(1) << i
		}
	}
	return m
}
func ttBit(t []word, i int) bool { return t[i>>6]&(word(1)<<uint(i&63)) != 0 }
func ttPut(t []word, i int, v bool) {
	if v {
		t[i>>6] |= word(1) << uint(i&63)
	}
}
func ttMap(t []word, n int32, mapping func(int) int) {
	old := append([]word(nil), t[:n]...)
	clear(t[:n])
	for i := 0; i < int(n)*64; i++ {
		ttPut(t, i, ttBit(old, mapping(i)))
	}
}
func Abc_TtCofactor0(t []word, n, v int32) { ttMap(t, n, func(i int) int { return i &^ (1 << v) }) }
func Abc_TtCofactor1(t []word, n, v int32) { ttMap(t, n, func(i int) int { return i | (1 << v) }) }
func Abc_TtCheckEqualCofs(t []word, n, i, j, a, b int32) int32 {
	if i >= j || a < 0 || a >= b || b > 3 {
		panic("truth: invalid cofactor comparison")
	}
	for k := 0; k < int(n)*64; k++ {
		if k&(1<<i|1<<j) != 0 {
			continue
		}
		x := k | int(a&1)<<i | int(a>>1)<<j
		y := k | int(b&1)<<i | int(b>>1)<<j
		if ttBit(t, x) != ttBit(t, y) {
			return 0
		}
	}
	return 1
}
func Abc_Tt6Cof0IsConst0(t word, v int32) int32 { return ttFlag(t&^ttMask(v) == 0) }
func Abc_Tt6Cof0IsConst1(t word, v int32) int32 { return ttFlag(t&^ttMask(v) == ^ttMask(v)) }
func Abc_Tt6Cof1IsConst0(t word, v int32) int32 { return ttFlag(t&ttMask(v) == 0) }
func Abc_Tt6Cof1IsConst1(t word, v int32) int32 { return ttFlag(t&ttMask(v) == ttMask(v)) }
func Abc_Tt6CofsOpposite(t word, v int32) int32 {
	return ttFlag((^t &^ ttMask(v)) == ((t >> (1 << v)) &^ ttMask(v)))
}
func Abc_Tt6Cof0EqualCof1(a, b word, v int32) int32 {
	return ttFlag((a &^ ttMask(v)) == ((b >> (1 << v)) &^ ttMask(v)))
}
func Abc_Tt6Cof0EqualCof0(a, b word, v int32) int32 {
	return ttFlag((a &^ ttMask(v)) == (b &^ ttMask(v)))
}
func Abc_Tt6Cof1EqualCof1(a, b word, v int32) int32 {
	return ttFlag((a & ttMask(v)) == (b & ttMask(v)))
}
func Abc_TtTruthIsConst0(t []word, n int32) int32 {
	for _, w := range t[:n] {
		if w != 0 {
			return 0
		}
	}
	return 1
}
func Abc_TtTruthIsConst1(t []word, n int32) int32 {
	for _, w := range t[:n] {
		if w != ^word(0) {
			return 0
		}
	}
	return 1
}
func ttCofConst(t []word, n, v int32, phase, value bool) int32 {
	for i := 0; i < int(n)*64; i++ {
		if (i&(1<<v) != 0) == phase && ttBit(t, i) != value {
			return 0
		}
	}
	return 1
}
func Abc_TtCof0IsConst0(t []word, n, v int32) int32 { return ttCofConst(t, n, v, false, false) }
func Abc_TtCof0IsConst1(t []word, n, v int32) int32 { return ttCofConst(t, n, v, false, true) }
func Abc_TtCof1IsConst0(t []word, n, v int32) int32 { return ttCofConst(t, n, v, true, false) }
func Abc_TtCof1IsConst1(t []word, n, v int32) int32 { return ttCofConst(t, n, v, true, true) }
func Abc_TtCofsOpposite(t []word, n, v int32) int32 {
	for i := 0; i < int(n)*64; i++ {
		if i&(1<<v) == 0 && ttBit(t, i) == ttBit(t, i|(1<<v)) {
			return 0
		}
	}
	return 1
}
func Abc_TtStretch5(t []uint32, a, b int32) {
	if a > b {
		panic("truth: reversed stretch")
	}
	size := int32(1)
	if a > 5 {
		size = 1 << (a - 5)
	}
	n := int32(1)
	if b > 5 {
		n = 1 << (b - 5)
	}
	for i := size; i < n; i++ {
		t[i] = t[i%size]
	}
}
func Abc_TtStretch6(t []word, a, b int32) {
	if a > b {
		panic("truth: reversed stretch")
	}
	size, n := ttWords(a), ttWords(b)
	for i := size; i < n; i++ {
		t[i] = t[i%size]
	}
}
func Abc_Tt6Stretch(t word, n int32) word {
	if n < 0 || n > 6 {
		panic("truth: invalid stretch")
	}
	size := 1 << n
	var out word
	for i := 0; i < 64; i++ {
		if t&(word(1)<<uint(i%size)) != 0 {
			out |= word(1) << i
		}
	}
	return out
}
func Abc_TtSuppFindFirst(s int32) int32 {
	if s <= 0 {
		panic("truth: empty support")
	}
	return int32(bits.TrailingZeros32(uint32(s)))
}
func Abc_TtSuppOnlyOne(s int32) int32 { return ttFlag(s != 0 && s&(s-1) == 0) }
func Abc_TtSuppIsMinBase(s int32) int32 {
	if s <= 0 {
		panic("truth: empty support")
	}
	return ttFlag(s&(s+1) == 0)
}
func Abc_Tt6HasVar(t word, v int32) int32 {
	return ttFlag(((t >> (1 << v)) &^ ttMask(v)) != (t &^ ttMask(v)))
}

// This matches the original helper's exact full-word complement comparison.
func Abc_Tt6XorVar(t word, v int32) int32 {
	return ttFlag(((t >> (1 << v)) &^ ttMask(v)) == ^(t &^ ttMask(v)))
}
func Abc_TtHasVar(t []word, n, v int32) int32 {
	if v < 0 || v >= n {
		panic("truth: variable outside support")
	}
	for i := 0; i < int(ttWords(n))*64; i++ {
		if i&(1<<v) == 0 && ttBit(t, i) != ttBit(t, i|(1<<v)) {
			return 1
		}
	}
	return 0
}
func Abc_TtSupport(t []word, n int32) int32 {
	var s int32
	for v := int32(0); v < n; v++ {
		if Abc_TtHasVar(t, n, v) != 0 {
			s |= 1 << v
		}
	}
	return s
}
func Abc_TtSupportSize(t []word, n int32) int32 {
	return int32(bits.OnesCount32(uint32(Abc_TtSupport(t, n))))
}
func Abc_TtSupportAndSize(t []word, n int32, size *int32) int32 {
	s := Abc_TtSupport(t, n)
	*size = int32(bits.OnesCount32(uint32(s)))
	return s
}
func Abc_Tt6SupportAndSize(t word, n int32, size *int32) int32 {
	return Abc_TtSupportAndSize([]word{t}, n, size)
}
func Abc_TtCheckCondDep2(t []word, n, limit int32) int32 {
	if n <= limit+1 {
		return 0
	}
	for v := int32(0); v < n; v++ {
		a, b := append([]word(nil), t[:ttWords(n)]...), append([]word(nil), t[:ttWords(n)]...)
		Abc_TtCofactor0(a, ttWords(n), v)
		Abc_TtCofactor1(b, ttWords(n), v)
		if Abc_TtSupportSize(a, n) <= limit && Abc_TtSupportSize(b, n) <= limit {
			return v
		}
	}
	return n
}
func Abc_TtCheckCondDep(t []word, n, limit int32) int32 {
	if n > 13 {
		panic("truth: conditional support limited to 13 variables")
	}
	return Abc_TtCheckCondDep2(t, n, limit)
}
func Abc_TtOnlyOneOne(t word) int32 { return ttFlag(t != 0 && t&(t-1) == 0) }
func ttLowMask(n int32) word {
	if n == 6 {
		return ^word(0)
	}
	return (word(1) << uint(1<<n)) - 1
}
func Abc_Tt6IsAndType(t word, n int32) int32 { return Abc_TtOnlyOneOne(t & ttLowMask(n)) }
func Abc_Tt6IsOrType(t word, n int32) int32  { return Abc_TtOnlyOneOne(^t & ttLowMask(n)) }
func Abc_Tt6IsXorType(t word, n int32) int32 {
	phase := t & 1
	for i := 0; i < 1<<n; i++ {
		if ((t>>i)&1)^phase != word(bits.OnesCount(uint(i))&1) {
			return 0
		}
	}
	return 1
}
func Abc_Tt6Flip(t word, v int32) word {
	return ((t << (1 << v)) & ttMask(v)) | ((t & ttMask(v)) >> (1 << v))
}
func Abc_TtFlip(t []word, n, v int32) { ttMap(t, n, func(i int) int { return i ^ (1 << v) }) }
func Abc_Tt6Permute_rec(t word, perm []int32, n int32) word {
	var out word
	for i := 0; i < 64; i++ {
		index := i
		for v := int32(0); v < n; v++ {
			index &= ^(1 << v)
			index |= ((i >> perm[v]) & 1) << v
		}
		if t&(word(1)<<index) != 0 {
			out |= word(1) << i
		}
	}
	return out
}
func Abc_Tt6SwapVars(t word, i, j int32) word {
	data := []word{t}
	Abc_TtSwapVars(data, 6, i, j)
	return data[0]
}
func Abc_Tt6SwapAdjacent(t word, v int32) word { return Abc_Tt6SwapVars(t, v, v+1) }
func Abc_TtSwapVars(t []word, n, i, j int32) {
	if i < 0 || j < 0 || i >= n || j >= n {
		panic("truth: invalid swap")
	}
	if i == j {
		return
	}
	ttMap(t, ttWords(n), func(k int) int {
		if (k>>i)&1 != (k>>j)&1 {
			k ^= 1<<i | 1<<j
		}
		return k
	})
}
func Abc_TtSwapAdjacent(t []word, n, v int32) {
	if v < 5 {
		for i := int32(0); i < n; i++ {
			t[i] = Abc_Tt6SwapAdjacent(t[i], v)
		}
		return
	}
	ttMap(t, n, func(k int) int {
		if (k>>v)&1 != (k>>(v+1))&1 {
			k ^= 1<<v | 1<<(v+1)
		}
		return k
	})
}
func Abc_TtMoveVar(t []word, n int32, v2p, p2v []int32, v, p int32) {
	a := v2p[v]
	if a == p {
		return
	}
	Abc_TtSwapVars(t, n, a, p)
	v2p[p2v[a]], v2p[p2v[p]] = p, a
	p2v[a], p2v[p] = p2v[p], p2v[a]
}
func Abc_Tt6RemoveVar(t word, v int32) word {
	if Abc_Tt6HasVar(t, v) != 0 {
		panic("truth: removing supported variable")
	}
	for ; v < 5; v++ {
		t = Abc_Tt6SwapAdjacent(t, v)
	}
	return t
}
func Abc_TtShrink(t []word, n, all int32, phase uint32) {
	var dst int32
	for v := int32(0); v < all; v++ {
		if phase&(1<<v) != 0 {
			for k := v - 1; k >= dst; k-- {
				Abc_TtSwapAdjacent(t, ttWords(all), k)
			}
			dst++
		}
	}
	if dst != n {
		panic("truth: shrink support mismatch")
	}
}
func Abc_TtMinimumBase(t []word, supp []int32, n int32, size *int32) int32 {
	s := Abc_TtSupport(t, n)
	count := int32(0)
	for v := int32(0); v < n; v++ {
		if s&(1<<v) != 0 {
			if supp != nil {
				supp[count] = supp[v]
			}
			count++
		}
	}
	if size != nil {
		*size = count
	}
	if s == 0 || Abc_TtSuppIsMinBase(s) != 0 {
		return 0
	}
	Abc_TtShrink(t, count, n, uint32(s))
	return 1
}
func Abc_TtSimplify(t []word, lits []int32, n int32, size *int32) int32 {
	for v := int32(0); v < n; v++ {
		if lits[v] == 0 {
			Abc_TtCofactor0(t, ttWords(n), v)
		} else if lits[v] == 1 {
			Abc_TtCofactor1(t, ttWords(n), v)
		}
	}
	for v := int32(0); v < n; v++ {
		for u := v + 1; u < n; u++ {
			if lits[v]>>1 != lits[u]>>1 {
				continue
			}
			if n > 6 {
				panic("truth: duplicate literals require small truth")
			}
			var out word
			for i := 0; i < 64; i++ {
				source := i
				bit := (i >> v) & 1
				if lits[v] != lits[u] {
					bit ^= 1
				}
				source = source&^(1<<u) | bit<<u
				if t[0]&(word(1)<<source) != 0 {
					out |= word(1) << i
				}
			}
			t[0] = out
		}
	}
	return Abc_TtMinimumBase(t, lits, n, size)
}
func Abc_Tt6Expand(t word, old []int32, a int32, cut []int32, b int32) word {
	data := []word{t}
	Abc_TtExpand(data, 6, old, a, cut, b)
	return data[0]
}
func Abc_TtExpand(t []word, n int32, old []int32, a int32, cut []int32, b int32) {
	k := a - 1
	for i := b - 1; i >= 0 && k >= 0; i-- {
		if cut[i] > old[k] {
			continue
		}
		if cut[i] != old[k] {
			panic("truth: cut is not a superset")
		}
		if k < i {
			Abc_TtSwapVars(t, n, k, i)
		}
		k--
	}
	if k != -1 {
		panic("truth: cut is not a superset")
	}
}
func Abc_Tt6MinBase(t *word, vars []int32, n int32) int32 {
	data := []word{*t}
	count := Abc_TtMinBase(data, vars, n, 6)
	*t = data[0]
	return count
}
func Abc_TtMinBase(t []word, vars []int32, n, all int32) int32 {
	var k int32
	for i := int32(0); i < n; i++ {
		if Abc_TtHasVar(t, all, i) == 0 {
			continue
		}
		if k < i {
			if vars != nil {
				vars[k] = vars[i]
			}
			Abc_TtSwapVars(t, all, k, i)
		}
		k++
	}
	return k
}
func Abc_TtImplementNpnConfig(t []word, n int32, perm []byte, phase uint32) {
	for i := int32(0); i < ttWords(n); i++ {
		if phase&(1<<n) != 0 {
			t[i] = ^t[i]
		}
	}
	for i := int32(0); i < n; i++ {
		if phase&(1<<i) != 0 {
			Abc_TtFlip(t, ttWords(n), i)
		}
	}
	if perm != nil {
		for i := int32(0); i < n; i++ {
			k := i
			for ; k < n && perm[k] != byte(i); k++ {
			}
			if k == n {
				panic("truth: invalid permutation")
			}
			if k != i {
				Abc_TtSwapVars(t, n, i, k)
				perm[i], perm[k] = perm[k], perm[i]
			}
		}
	}
}

func Abc_TtIsHexDigit(c byte) int32 {
	return ttFlag(c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f')
}
func Abc_TtPrintDigit(d int32) byte {
	if d < 0 || d > 15 {
		panic("truth: invalid hex digit")
	}
	return "0123456789ABCDEF"[d]
}
func Abc_TtPrintDigitLower(d int32) byte {
	if d < 0 || d > 15 {
		panic("truth: invalid hex digit")
	}
	return "0123456789abcdef"[d]
}
func Abc_TtReadHexDigit(c byte) int32 {
	if c >= '0' && c <= '9' {
		return int32(c - '0')
	}
	if c >= 'A' && c <= 'F' {
		return int32(c-'A') + 10
	}
	if c >= 'a' && c <= 'f' {
		return int32(c-'a') + 10
	}
	panic("truth: invalid hex digit")
}
func Abc_TtPrintHex(t []word, n int32) {
	if n < 2 {
		panic("truth: hex requires two variables")
	}
	for _, w := range t[:ttWords(n)] {
		for k := 0; k < 16; k++ {
			fmt.Fprintf(os.Stdout, "%c", Abc_TtPrintDigit(int32(w>>uint(4*k)&15)))
		}
	}
	fmt.Fprintln(os.Stdout)
}
func Abc_TtPrintHexRev(out io.Writer, t []word, n int32) {
	size := int32(16)
	if n < 2 {
		panic("truth: hex requires two variables")
	}
	if n < 6 {
		size = 1 << (n - 2)
	}
	for i := ttWords(n) - 1; i >= 0; i-- {
		for k := size - 1; k >= 0; k-- {
			fmt.Fprintf(out, "%c", Abc_TtPrintDigit(int32(t[i]>>uint(4*k)&15)))
		}
	}
}
func Abc_TtPrintHexSpecial(t []word, n int32) {
	if n < 2 {
		panic("truth: hex requires two variables")
	}
	for i := ttWords(n) - 1; i >= 0; i-- {
		for k := 0; k < 16; k++ {
			fmt.Fprintf(os.Stdout, "%c", Abc_TtPrintDigit(int32(t[i]>>uint(4*k)&15)))
		}
	}
	fmt.Fprintln(os.Stdout)
}
func Abc_TtWriteHexRev(dst []byte, t []word, n int32) int32 {
	var out strings.Builder
	Abc_TtPrintHexRev(&out, t, n)
	s := out.String()
	if len(dst) < len(s) {
		panic("truth: hex output buffer too small")
	}
	copy(dst, s)
	return int32(len(s))
}
func Abc_TtPrintHexArrayRev(out io.Writer, t []word, n int32) {
	for i := n - 1; i >= 0; i-- {
		fmt.Fprintf(out, "%c", Abc_TtPrintDigitLower(int32(t[i/16]>>uint(4*(i%16))&15)))
	}
}
func Abc_TtReadHex(t []word, s string) int32 {
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	digits := 0
	for digits < len(s) && Abc_TtIsHexDigit(s[digits]) != 0 {
		digits++
	}
	if digits == 0 {
		panic("truth: no hex digits")
	}
	if digits == 1 {
		switch s[0] {
		case '0':
			t[0] = 0
			return 0
		case 'F':
			t[0] = ^word(0)
			return 0
		case '5':
			t[0] = ^ttMask(0)
			return 1
		case 'A':
			t[0] = ttMask(0)
			return 1
		}
	}
	n := int32(2 + bits.Len(uint(digits-1)))
	clear(t[:ttWords(n)])
	Abc_TtReadHexNumber(t, s[:digits])
	if n < 6 {
		t[0] = Abc_Tt6Stretch(t[0], n)
	}
	return n
}
func Abc_TtReadHexNumber(t []word, s string) int32 {
	digits := 0
	for digits < len(s) && Abc_TtIsHexDigit(s[digits]) != 0 {
		digits++
	}
	for k := 0; k < digits; k++ {
		shift := uint(4 * (k % 16))
		i := k / 16
		t[i] |= word(Abc_TtReadHexDigit(s[digits-1-k])) << shift
	}
	return int32(digits)
}
func Abc_TtPrintBinary(t []word, n int32) {
	if n < 2 {
		panic("truth: binary requires two variables")
	}
	limit := int32(64)
	if n < 6 {
		limit = 1 << n
	}
	for _, w := range t[:ttWords(n)] {
		for k := int32(0); k < limit; k++ {
			fmt.Fprint(os.Stdout, w>>uint(k)&1)
		}
	}
	fmt.Fprintln(os.Stdout)
}
