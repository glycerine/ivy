package rsb

import (
	"fmt"
	"math/bits"
	"os"

	"github.com/glycerine/abcgo/misc/util"
)

func Rsb_DecTry0(c uint64) uint32 {
	if c != 0 {
		return 1
	}
	return 0
}
func decTry(c uint64, functions []uint64) uint32 {
	if len(functions) == 0 {
		return Rsb_DecTry0(c)
	}
	last := functions[len(functions)-1]
	return decTry(c&last, functions[:len(functions)-1])<<uint(1<<uint(len(functions)-1)) | decTry(c&^last, functions[:len(functions)-1])
}
func Rsb_DecTry1(c, f1 uint64) uint32                 { return decTry(c, []uint64{f1}) }
func Rsb_DecTry2(c, f1, f2 uint64) uint32             { return decTry(c, []uint64{f1, f2}) }
func Rsb_DecTry3(c, f1, f2, f3 uint64) uint32         { return decTry(c, []uint64{f1, f2, f3}) }
func Rsb_DecTry4(c, f1, f2, f3, f4 uint64) uint32     { return decTry(c, []uint64{f1, f2, f3, f4}) }
func Rsb_DecTry5(c, f1, f2, f3, f4, f5 uint64) uint32 { return decTry(c, []uint64{f1, f2, f3, f4, f5}) }
func Rsb_DecTryCex(g []uint64, a, b int32) int32 {
	if rsbBit(g, a) == rsbBit(g, b) {
		return 1
	}
	return 0
}
func Rsb_DecVerifyCex(f []uint64, g [][]uint64, n int32, a, b int32) bool {
	if a < 0 || b < 0 || Rsb_DecTryCex(f, a, b) != 0 {
		return false
	}
	for _, truth := range g[:n] {
		if Rsb_DecTryCex(truth, a, b) == 0 {
			return false
		}
	}
	return true
}
func Rsb_DecRecordCex(g [][]uint64, n int32, a, b int32, cexes []uint64, count int32) {
	if count < 0 || count >= 64 {
		panic("rsb: counterexample budget exceeded")
	}
	for i, truth := range g[:n] {
		if Rsb_DecTryCex(truth, a, b) != 0 {
			cexes[i] |= 1 << uint(count)
		}
	}
}
func Rsb_DecCofactor(g [][]uint64, n, w, mint int32) uint64 {
	cof := ^uint64(0)
	for i, truth := range g[:n] {
		if mint&(1<<uint(i)) != 0 {
			cof &= truth[w]
		} else {
			cof &^= truth[w]
		}
	}
	return cof
}

// A pattern packs reachable output=0 divisor assignments in its low half and
// output=1 assignments in its high half. An overlap supplies a genuine pair of
// minterms that the selected divisors cannot distinguish.
func Rsb_DecCheck(nVars int32, f []uint64, g [][]uint64, nGs int32, patterns []uint32, cexA, cexB *int32) uint32 {
	if nGs <= 0 || nGs > 4 || len(g) < int(nGs) {
		panic("rsb: decomposition needs 1..4 divisors")
	}
	words := rsbWords(nVars)
	if len(f) < words || len(patterns) < words {
		panic("rsb: short truth/pattern buffer")
	}
	for _, truth := range g[:nGs] {
		if len(truth) < words {
			panic("rsb: short divisor truth")
		}
	}
	shift := uint(1 << uint(nGs))
	mask := uint32(1)<<shift - 1
	var total uint32
	var functions [5]uint64
	for w := 0; w < words; w++ {
		for i := int32(0); i < nGs; i++ {
			functions[i] = g[i][w]
		}
		functions[nGs] = f[w]
		patterns[w] = decTry(^uint64(0), functions[:nGs+1])
		overlap := (patterns[w] >> shift) & patterns[w] & mask
		if overlap != 0 {
			mint := int32(bits.TrailingZeros32(overlap))
			cof := Rsb_DecCofactor(g, nGs, int32(w), mint)
			*cexA = int32(w*64 + bits.TrailingZeros64(cof&^f[w]))
			*cexB = int32(w*64 + bits.TrailingZeros64(cof&f[w]))
			return 0
		}
		total |= patterns[w]
		if w == 0 || (total>>shift)&total&mask == 0 {
			continue
		}
		for k := 0; k < w; k++ {
			overlap = ((patterns[k] | patterns[w]) >> shift) & (patterns[k] | patterns[w]) & mask
			if overlap == 0 {
				continue
			}
			mint := int32(bits.TrailingZeros32(overlap))
			a, b := Rsb_DecCofactor(g, nGs, int32(k), mint), Rsb_DecCofactor(g, nGs, int32(w), mint)
			if a&^f[k] != 0 && b&f[w] != 0 {
				*cexA = int32(k*64 + bits.TrailingZeros64(a&^f[k]))
				*cexB = int32(w*64 + bits.TrailingZeros64(b&f[w]))
			} else {
				*cexA = int32(k*64 + bits.TrailingZeros64(a&f[k]))
				*cexB = int32(w*64 + bits.TrailingZeros64(b&^f[w]))
			}
			return 0
		}
	}
	return total
}
func firstLast(t []uint64, value bool) (int32, int32) {
	first, last := int32(-1), int32(-1)
	for i, w := range t {
		if !value {
			w = ^w
		}
		if w == 0 {
			continue
		}
		if first < 0 {
			first = int32(i*64 + bits.TrailingZeros64(w))
		}
		last = int32(i*64 + 63 - bits.LeadingZeros64(w))
	}
	return first, last
}
func Rsb_DecInitCexes(nVars int32, f []uint64, g [][]uint64, nGs, nGsAll int32, cexes []uint64, tries *[]int32) int32 {
	on0, on1 := firstLast(f[:rsbWords(nVars)], true)
	off0, off1 := firstLast(f[:rsbWords(nVars)], false)
	if on0 < 0 || off0 < 0 {
		return 0
	}
	for i, pair := range [][2]int32{{on0, off0}, {on0, off1}, {on1, off0}, {on1, off1}} {
		Rsb_DecRecordCex(g, nGsAll, pair[0], pair[1], cexes, int32(i))
		if tries != nil {
			*tries = append(*tries, -1)
		}
	}
	return 4
}

// Enumerate singleton/pair/triple/quadruple candidates in exactly the original
// nested-loop order. A counterexample column filters every later candidate
// whose selected divisors all fail to distinguish that pair.
func Rsb_DecPerformInt(p *Rsb_Man_t, nVars int32, f []uint64, g [][]uint64, nGs, nGsAll, findAll int32) uint32 {
	if nVars > p.leafMax || nGs < 0 || nGs > nGsAll || int(nGsAll) > len(g) {
		panic("rsb: invalid search dimensions")
	}
	words := rsbWords(nVars)
	if len(f) < words {
		panic("rsb: short function")
	}
	for _, t := range g[:nGsAll] {
		if len(t) < words {
			panic("rsb: short divisor")
		}
	}
	p.cexes = append(p.cexes[:0], make([]uint64, nGsAll)...)
	p.patterns = append(p.patterns[:0], make([]uint32, words)...)
	p.fanins = p.fanins[:0]
	p.tries = p.tries[:0]
	count := int32(0)
	var result uint32
	var chosen []int32
	limit := p.decMax
	if limit <= 0 || limit > 4 {
		limit = 4
	}
	limit = min(limit, nGs)
	var enumerate func(int32, int32, []int32) bool
	enumerate = func(left, upper int32, indices []int32) bool {
		if left != 0 {
			for index := left - 1; index < upper; index++ {
				if enumerate(left-1, index, append(indices, index)) {
					return true
				}
			}
			return false
		}
		common := ^uint64(0)
		divisors := make([][]uint64, len(indices))
		for i, index := range indices {
			common &= p.cexes[index]
			divisors[i] = g[index]
		}
		if common != 0 {
			return false
		}
		a, b := int32(-1), int32(-1)
		pattern := Rsb_DecCheck(nVars, f, divisors, int32(len(divisors)), p.patterns, &a, &b)
		if pattern != 0 {
			if findAll != 0 {
				fmt.Printf("Decomposition %v pattern %08x\n", indices, pattern)
				return false
			}
			result = pattern
			chosen = append([]int32(nil), indices...)
			return true
		}
		if count == 64 {
			return true
		}
		if !Rsb_DecVerifyCex(f, divisors, int32(len(divisors)), a, b) {
			panic("rsb: invalid generated counterexample")
		}
		Rsb_DecRecordCex(g, nGsAll, a, b, p.cexes, count)
		count++
		if p.verbose {
			p.tries = append(p.tries, indices...)
			p.tries = append(p.tries, -1)
		}
		return false
	}
	for size := int32(1); size <= limit; size++ {
		if enumerate(size, nGs, nil) {
			break
		}
	}
	p.fanins = append(p.fanins, chosen...)
	return result
}
func Rsb_ManPerform(p *Rsb_Man_t, nVars int32, f []uint64, g [][]uint64, nGs, nGsAll, verbose int32) uint32 {
	if len(p.faninsOld) > 0 && len(p.faninsOld) <= 4 {
		divisors := make([][]uint64, len(p.faninsOld))
		for i, index := range p.faninsOld {
			if index < 0 || int(index) >= len(g) {
				panic("rsb: old fanin out of range")
			}
			divisors[i] = g[index]
		}
		a, b := int32(-1), int32(-1)
		if Rsb_DecCheck(nVars, f, divisors, int32(len(divisors)), make([]uint32, rsbWords(nVars)), &a, &b) == 0 {
			fmt.Printf("Verification FAILED for original decomposition %v.\n", p.faninsOld)
		}
	}
	return Rsb_DecPerformInt(p, nVars, f, g, nGs, nGsAll, 0)
}
func Rsb_ManPerformResub6(p *Rsb_Man_t, nVars int32, truth uint64, divisors []uint64, off, on *uint64, verbose int32) int32 {
	if nVars < 0 || nVars > 6 || len(divisors) >= 200 {
		panic("rsb: six-variable resubstitution dimensions invalid")
	}
	g := make([][]uint64, len(divisors))
	for i := range divisors {
		g[i] = divisors[i : i+1]
	}
	pattern := Rsb_DecPerformInt(p, nVars, []uint64{truth}, g, int32(len(g)), int32(len(g)), 0)
	if pattern == 0 {
		return 0
	}
	n := int32(len(p.fanins))
	*off = rsbStretch(uint64(pattern), n)
	*on = rsbStretch(uint64(pattern)>>uint(1<<uint(n)), n)
	if verbose != 0 {
		Rsb_DecPrintFunc(p, pattern, []uint64{truth}, g, int32(len(g)), nVars)
		Rsb_DecPrintTable(p.cexes, int32(len(g)), int32(len(g)), p.tries)
	}
	return 1
}
func Rsb_DecVerify(p *Rsb_Man_t, nVars int32, f []uint64, g [][]uint64, nGs int32, pattern uint32, temp, result []uint64) int32 {
	words := rsbWords(nVars)
	if len(temp) < words || len(result) < words {
		panic("rsb: short verification scratch")
	}
	clear(temp[:words])
	n := uint(len(p.fanins))
	function := pattern >> uint(1<<n)
	for a := 0; a < words*64; a++ {
		mint := uint(0)
		for i, index := range p.fanins {
			if index < 0 || index >= nGs {
				panic("rsb: fanin out of range")
			}
			mint |= uint(rsbBit(g[index], int32(a))) << uint(i)
		}
		if function&(1<<mint) != 0 {
			temp[a>>6] |= 1 << uint(a&63)
		}
	}
	for i, w := range temp[:words] {
		if w != f[i] {
			return 0
		}
	}
	return 1
}
func Rsb_DecPrintFunc(p *Rsb_Man_t, pattern uint32, f []uint64, g [][]uint64, nGs, nVars int32) {
	n := int32(len(p.fanins))
	off, on := rsbStretch(uint64(pattern), n), rsbStretch(uint64(pattern)>>uint(1<<uint(n)), n)
	count := int32(0)
	isop := util.Abc_Tt6Isop(on, ^off, n, &count)
	fmt.Printf("Offset : %016x\nOnset  : %016x\nResult : %016x\nFunc   : %x\n", off, on, isop, f)
	for i, t := range g[:nGs] {
		fmt.Printf("Div%3d : %x\n", i, t)
	}
	fmt.Printf("Solution : %v\n", p.fanins)
}
func Rsb_DecPrintTable(cexes []uint64, nGs, nGsAll int32, tries []int32) {
	if len(tries) == 0 {
		return
	}
	start := 0
	for c := 0; start < len(tries); c++ {
		end := start
		for end < len(tries) && tries[end] != -1 {
			end++
		}
		fmt.Fprintf(os.Stdout, "%v  ", tries[start:end])
		for i, w := range cexes[:nGsAll] {
			if int32(i) == nGs {
				fmt.Print("|")
			}
			if w&(1<<uint(c)) != 0 {
				fmt.Print(".")
			} else {
				fmt.Print("+")
			}
		}
		fmt.Printf("  %3d\n", c)
		start = end + 1
	}
	fmt.Print("Counts : ")
	for i, w := range cexes[:nGsAll] {
		if int32(i) == nGs {
			fmt.Print("|")
		}
		fmt.Printf(" %d", bits.OnesCount64(w))
	}
	fmt.Println()
}
func Rsb_ManPerformResub6Test() {
	x := [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}
	d := []uint64{x[0], x[1], x[2], x[3], x[4], x[5], x[0] & x[1], x[2] & x[3], x[4] & x[5]}
	p := Rsb_ManAlloc(6, 64, 4, 1)
	var off, on uint64
	Rsb_ManPerformResub6(p, 6, d[6]|d[7]|d[8], d, &off, &on, 1)
}
