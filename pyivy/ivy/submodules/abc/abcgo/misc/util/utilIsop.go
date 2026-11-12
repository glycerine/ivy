// ISOP/ESOP cover computation, ported from ABC utilIsop.c (Alan Mishchenko).
package util

import (
	"fmt"
	"math/bits"
	"sync"
)

const ABC_ISOP_MAX_VAR int32 = 16
const ABC_ISOP_MAX_WORD = 1 << (ABC_ISOP_MAX_VAR - 6)
const ABC_ISOP_MAX_CUBE int32 = 0xffff

type FUNC_ISOP = func([]word, []word, []word, word, []int32) word

func Abc_Cube2Cost(n int32) word { return word(uint32(n)) << 32 }
func Abc_CostCubes(c word) int32 { return int32(c >> 32) }
func Abc_CostLits(c word) int32  { return int32(uint32(c)) }
func isopWords(n int32) int {
	if n < 0 || n > ABC_ISOP_MAX_VAR {
		panic("ISOP variable count outside 0..16")
	}
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func isopMask(v int32) word { return s_Truths6[v] }

func isopCofactor(t word, v int32, positive bool) word {
	mask := isopMask(v)
	shift := uint(1 << uint(v))
	if positive {
		x := t & mask
		return x | x>>shift
	}
	x := t &^ mask
	return x | x<<shift
}
func isopHasVar(t word, v int32) bool { return isopCofactor(t, v, false) != isopCofactor(t, v, true) }
func isopCoverTail(cover []int32, n int32) []int32 {
	if cover == nil {
		return nil
	}
	return cover[n:]
}
func Abc_IsopAddLits(cover []int32, c0, c1 word, v int32) {
	if cover == nil {
		return
	}
	n0, n1 := Abc_CostCubes(c0), Abc_CostCubes(c1)
	for i := int32(0); i < n0; i++ {
		cover[i] = int32(uint32(cover[i]) | 1<<uint(2*v))
	}
	for i := int32(0); i < n1; i++ {
		cover[n0+i] = int32(uint32(cover[n0+i]) | 1<<uint(2*v+1))
	}
}

type isopSolution struct {
	cost  word
	truth []word
	cubes []int32
}

func isopStore(s isopSolution, result []word, cover []int32, limit word) word {
	if s.cost >= limit {
		return limit
	}
	if len(result) < len(s.truth) {
		panic("short ISOP result buffer")
	}
	copy(result, s.truth)
	if cover != nil {
		if len(cover) < len(s.cubes) {
			panic("short ISOP cover buffer")
		}
		copy(cover, s.cubes)
	}
	return s.cost
}
func isopScalar(on, dc word, n int32, limit word, collect bool) isopSolution {
	if on&^dc != 0 {
		panic("ISOP onset is not contained in onset/don't-care set")
	}
	if on == 0 {
		return isopSolution{truth: []word{0}}
	}
	if dc == ^word(0) {
		s := isopSolution{cost: Abc_Cube2Cost(1), truth: []word{^word(0)}}
		if collect {
			s.cubes = []int32{0}
		}
		return s
	}
	v := n - 1
	for v >= 0 && !isopHasVar(on, v) && !isopHasVar(dc, v) {
		v--
	}
	if v < 0 {
		panic("ISOP truth table has inconsistent support")
	}
	on0, on1 := isopCofactor(on, v, false), isopCofactor(on, v, true)
	dc0, dc1 := isopCofactor(dc, v, false), isopCofactor(dc, v, true)
	a := isopScalar(on0&^dc1, dc0, v, limit, collect)
	if a.cost >= limit {
		return isopSolution{cost: limit}
	}
	b := isopScalar(on1&^dc0, dc1, v, limit, collect)
	if a.cost+b.cost >= limit {
		return isopSolution{cost: limit}
	}
	c := isopScalar((on0&^a.truth[0])|(on1&^b.truth[0]), dc0&dc1, v, limit, collect)
	if a.cost+b.cost+c.cost >= limit {
		return isopSolution{cost: limit}
	}
	result := isopSolution{cost: a.cost + b.cost + c.cost + word(Abc_CostCubes(a.cost)+Abc_CostCubes(b.cost)), truth: []word{c.truth[0] | (a.truth[0] &^ isopMask(v)) | (b.truth[0] & isopMask(v))}}
	if collect {
		result.cubes = append(result.cubes, a.cubes...)
		result.cubes = append(result.cubes, b.cubes...)
		Abc_IsopAddLits(result.cubes, a.cost, b.cost, v)
		result.cubes = append(result.cubes, c.cubes...)
	}
	return result
}
func isopRecur(on, dc []word, n int32, limit word, collect bool) isopSolution {
	count := isopWords(n)
	if len(on) < count || len(dc) < count {
		panic("short ISOP input")
	}
	if n <= 6 {
		return isopScalar(on[0], dc[0], n, limit, collect)
	}
	for i := 0; i < count; i++ {
		if on[i]&^dc[i] != 0 {
			panic("ISOP onset outside allowed set")
		}
	}
	half := count / 2
	// Removing an unused top variable is the original Abc_IsopCheck shortcut.
	equal := true
	for i := 0; i < half; i++ {
		if on[i] != on[half+i] || dc[i] != dc[half+i] {
			equal = false
			break
		}
	}
	if equal {
		s := isopRecur(on[:half], dc[:half], n-1, limit, collect)
		if s.cost < limit {
			s.truth = append(s.truth, s.truth...)
		}
		return s
	}
	aOn, bOn := make([]word, half), make([]word, half)
	for i := 0; i < half; i++ {
		aOn[i] = on[i] &^ dc[half+i]
		bOn[i] = on[half+i] &^ dc[i]
	}
	a := isopRecur(aOn, dc[:half], n-1, limit, collect)
	if a.cost >= limit {
		return isopSolution{cost: limit}
	}
	b := isopRecur(bOn, dc[half:count], n-1, limit, collect)
	if a.cost+b.cost >= limit {
		return isopSolution{cost: limit}
	}
	middle, middleDc := make([]word, half), make([]word, half)
	for i := 0; i < half; i++ {
		middle[i] = (on[i] &^ a.truth[i]) | (on[half+i] &^ b.truth[i])
		middleDc[i] = dc[i] & dc[half+i]
	}
	c := isopRecur(middle, middleDc, n-1, limit, collect)
	if a.cost+b.cost+c.cost >= limit {
		return isopSolution{cost: limit}
	}
	result := isopSolution{cost: a.cost + b.cost + c.cost + word(Abc_CostCubes(a.cost)+Abc_CostCubes(b.cost)), truth: make([]word, count)}
	for i := 0; i < half; i++ {
		result.truth[i] = c.truth[i] | a.truth[i]
		result.truth[half+i] = c.truth[i] | b.truth[i]
	}
	if collect {
		result.cubes = append(result.cubes, a.cubes...)
		result.cubes = append(result.cubes, b.cubes...)
		Abc_IsopAddLits(result.cubes, a.cost, b.cost, n-1)
		result.cubes = append(result.cubes, c.cubes...)
	}
	return result
}
func Abc_Isop6Cover(on, dc word, result []word, n int32, limit word, cover []int32) word {
	if n < 0 || n > 6 {
		panic("scalar ISOP supports at most six variables")
	}
	return isopStore(isopScalar(on, dc, n, limit, cover != nil), result, cover, limit)
}
func Abc_IsopCheck(on, dc, result []word, n int32, limit word, cover []int32) word {
	return isopStore(isopRecur(on, dc, n, limit, cover != nil), result, cover, limit)
}
func Abc_Isop7Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 7, limit, cover)
}
func Abc_Isop8Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 8, limit, cover)
}
func Abc_Isop9Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 9, limit, cover)
}
func Abc_Isop10Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 10, limit, cover)
}
func Abc_Isop11Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 11, limit, cover)
}
func Abc_Isop12Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 12, limit, cover)
}
func Abc_Isop13Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 13, limit, cover)
}
func Abc_Isop14Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 14, limit, cover)
}
func Abc_Isop15Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 15, limit, cover)
}
func Abc_Isop16Cover(on, dc, result []word, limit word, cover []int32) word {
	return Abc_IsopCheck(on, dc, result, 16, limit, cover)
}

var isopElementsOnce sync.Once
var isopElements [][]word

func Abc_IsopTtElems() [][]word {
	isopElementsOnce.Do(func() {
		isopElements = make([][]word, ABC_ISOP_MAX_VAR+1)
		for v := int32(0); v <= ABC_ISOP_MAX_VAR; v++ {
			isopElements[v] = make([]word, ABC_ISOP_MAX_WORD)
			for i := range isopElements[v] {
				if v < 6 {
					isopElements[v][i] = isopMask(v)
				} else if i&(1<<uint(v-6)) != 0 {
					isopElements[v][i] = ^word(0)
				}
			}
		}
	})
	return isopElements
}
func isopCubeWord(cube uint32, vCount int32, index int) word {
	value := ^word(0)
	for v := int32(0); v < vCount; v++ {
		literal := (cube >> uint(2*v)) & 3
		if literal == 3 {
			panic("cube contains both polarities")
		}
		if literal == 0 {
			continue
		}
		if v < 6 {
			mask := isopMask(v)
			if literal == 1 {
				mask = ^mask
			}
			value &= mask
		} else {
			positive := index&(1<<uint(v-6)) != 0
			if positive != (literal == 2) {
				return 0
			}
		}
	}
	return value
}
func Abc_IsopBuildTruth(cover []int32, n int32, result []word, xor, compl int32) {
	words := isopWords(n)
	if len(result) < words {
		panic("short truth result")
	}
	clear(result[:words])
	for _, cube := range cover {
		for i := 0; i < words; i++ {
			value := isopCubeWord(uint32(cube), n, i)
			if xor != 0 {
				result[i] ^= value
			} else {
				result[i] |= value
			}
		}
	}
	if compl != 0 {
		for i := 0; i < words; i++ {
			result[i] = ^result[i]
		}
	}
}
func Abc_IsopVerify(function []word, n int32, result []word, cover []int32, xor, compl int32) {
	Abc_IsopBuildTruth(cover, n, result, xor, compl)
	for i := 0; i < isopWords(n); i++ {
		if function[i] != result[i] {
			fmt.Println("Verification failed.")
			return
		}
	}
}
func Abc_Isop(function []word, n, limit int32, cover *[]int32, tryBoth int32) int32 {
	if limit < 0 {
		panic("negative cube limit")
	}
	isopWords(n)
	costLimit := Abc_Cube2Cost(limit)
	normal := isopRecur(function, function, n, costLimit, true)
	best := normal
	complement := int32(0)
	if tryBoth != 0 {
		neg := make([]word, isopWords(n))
		for i := range neg {
			neg[i] = ^function[i]
		}
		other := isopRecur(neg, neg, n, normal.cost, true)
		if other.cost < best.cost {
			best = other
			complement = 1
		}
	}
	if best.cost >= costLimit {
		return -1
	}
	if cover != nil {
		*cover = append((*cover)[:0], best.cubes...)
	}
	return complement
}
func Abc_IsopCnf(function []word, n, limit int32, cover []int32) int32 {
	if n >= 16 {
		panic("CNF cover needs an extra literal variable; at most 15 inputs")
	}
	costLimit := Abc_Cube2Cost(limit)
	a := isopRecur(function, function, n, costLimit, cover != nil)
	if a.cost >= costLimit {
		return int32(costLimit)
	}
	neg := make([]word, isopWords(n))
	for i := range neg {
		neg[i] = ^function[i]
	}
	b := isopRecur(neg, neg, n, costLimit, cover != nil)
	if a.cost+b.cost >= costLimit {
		return int32(costLimit)
	}
	if cover != nil {
		combined := append(a.cubes, b.cubes...)
		Abc_IsopAddLits(combined, a.cost, b.cost, n)
		if len(cover) < len(combined) {
			panic("short CNF cover")
		}
		copy(cover, combined)
	}
	return Abc_CostCubes(a.cost) + Abc_CostCubes(b.cost)
}
func Abc_IsopCountLits(cover []int32, n int32) int32 {
	var count int32
	for _, cube := range cover {
		for v := int32(0); v < n; v++ {
			literal := (uint32(cube) >> uint(2*v)) & 3
			if literal == 3 {
				panic("invalid cover literal")
			}
			if literal != 0 {
				count++
			}
		}
	}
	return count
}
func Abc_IsopPrintCover(cover []int32, n, compl int32) {
	if len(cover) == 0 || (len(cover) == 1 && cover[0] == 0) {
		fmt.Printf("Constant %d\n", len(cover))
		return
	}
	for _, cube := range cover {
		for v := int32(0); v < n; v++ {
			switch uint32(cube) >> uint(2*v) & 3 {
			case 0:
				fmt.Print("-")
			case 1:
				fmt.Print("0")
			case 2:
				fmt.Print("1")
			default:
				panic("invalid cube")
			}
		}
		fmt.Printf(" %d\n", 1-compl)
	}
}
func Abc_IsopPrint(function []word, n int32, cover *[]int32, both int32) {
	compl := Abc_Isop(function, n, ABC_ISOP_MAX_CUBE, cover, both)
	if cover != nil {
		Abc_IsopPrintCover(*cover, n, compl)
	}
}

func Abc_EsopAddLits(cover []int32, a, b, c, max word, v int32) int32 {
	na, nb, nc := Abc_CostCubes(a), Abc_CostCubes(b), Abc_CostCubes(c)
	if max == a {
		if cover != nil {
			copy(cover, cover[na:na+nb])
			for i := int32(0); i < nc; i++ {
				cover[nb+i] = int32(uint32(cover[na+nb+i]) | 1<<uint(2*v))
			}
		}
		return nc
	}
	if max == b {
		if cover != nil {
			for i := int32(0); i < nc; i++ {
				cover[na+i] = int32(uint32(cover[na+nb+i]) | 1<<uint(2*v+1))
			}
		}
		return nc
	}
	if cover != nil {
		Abc_IsopAddLits(cover, a, b, v)
	}
	return na + nb
}

type esopSolution struct {
	cost  word
	cubes []int32
}

func esopCombine(a, b, c esopSolution, v int32, limit word, collect bool) esopSolution {
	max := a.cost
	if b.cost > max {
		max = b.cost
	}
	if c.cost > max {
		max = c.cost
	}
	cost := a.cost + b.cost + c.cost - max
	if cost >= limit {
		return esopSolution{cost: limit}
	}
	var added int32
	var cubes []int32
	switch max {
	case a.cost:
		added = Abc_CostCubes(c.cost)
		if collect {
			cubes = append(cubes, b.cubes...)
			for _, cube := range c.cubes {
				cubes = append(cubes, int32(uint32(cube)|1<<uint(2*v)))
			}
		}
	case b.cost:
		added = Abc_CostCubes(c.cost)
		if collect {
			cubes = append(cubes, a.cubes...)
			for _, cube := range c.cubes {
				cubes = append(cubes, int32(uint32(cube)|1<<uint(2*v+1)))
			}
		}
	default:
		added = Abc_CostCubes(a.cost) + Abc_CostCubes(b.cost)
		if collect {
			cubes = append(cubes, a.cubes...)
			cubes = append(cubes, b.cubes...)
			Abc_IsopAddLits(cubes, a.cost, b.cost, v)
		}
	}
	return esopSolution{cost: cost + word(added), cubes: cubes}
}
func esopScalar(t word, n int32, limit word, collect bool) esopSolution {
	if t == 0 {
		return esopSolution{}
	}
	if t == ^word(0) {
		s := esopSolution{cost: Abc_Cube2Cost(1)}
		if collect {
			s.cubes = []int32{0}
		}
		return s
	}
	v := n - 1
	for v >= 0 && !isopHasVar(t, v) {
		v--
	}
	if v < 0 {
		panic("ESOP support mismatch")
	}
	c0, c1 := isopCofactor(t, v, false), isopCofactor(t, v, true)
	a := esopScalar(c0, v, limit, collect)
	if a.cost >= limit {
		return esopSolution{cost: limit}
	}
	b := esopScalar(c1, v, limit, collect)
	if b.cost >= limit {
		return esopSolution{cost: limit}
	}
	c := esopScalar(c0^c1, v, limit, collect)
	if c.cost >= limit {
		return esopSolution{cost: limit}
	}
	return esopCombine(a, b, c, v, limit, collect)
}
func esopRecur(t []word, n int32, limit word, collect bool) esopSolution {
	count := isopWords(n)
	if len(t) < count {
		panic("short ESOP truth")
	}
	if n <= 6 {
		return esopScalar(t[0], n, limit, collect)
	}
	half := count / 2
	equal := true
	for i := 0; i < half; i++ {
		if t[i] != t[i+half] {
			equal = false
			break
		}
	}
	if equal {
		return esopRecur(t[:half], n-1, limit, collect)
	}
	a := esopRecur(t[:half], n-1, limit, collect)
	if a.cost >= limit {
		return esopSolution{cost: limit}
	}
	b := esopRecur(t[half:count], n-1, limit, collect)
	if b.cost >= limit {
		return esopSolution{cost: limit}
	}
	xor := make([]word, half)
	for i := range xor {
		xor[i] = t[i] ^ t[half+i]
	}
	c := esopRecur(xor, n-1, limit, collect)
	if c.cost >= limit {
		return esopSolution{cost: limit}
	}
	return esopCombine(a, b, c, n-1, limit, collect)
}
func esopStore(s esopSolution, cover []int32, limit word) word {
	if s.cost >= limit {
		return limit
	}
	if cover != nil {
		if len(cover) < len(s.cubes) {
			panic("short ESOP cover buffer")
		}
		copy(cover, s.cubes)
	}
	return s.cost
}
func Abc_Esop6Cover(t word, n int32, limit word, cover []int32) word {
	if n < 0 || n > 6 {
		panic("scalar ESOP supports <=6 variables")
	}
	return esopStore(esopScalar(t, n, limit, cover != nil), cover, limit)
}
func Abc_EsopCover(t []word, n int32, limit word, cover []int32) word {
	return Abc_EsopCheck(t, n, limit, cover)
}
func Abc_EsopCheck(t []word, n int32, limit word, cover []int32) word {
	return esopStore(esopRecur(t, n, limit, cover != nil), cover, limit)
}

func Abc_TtIntersect2(a, b []word, n int32) int32 {
	for i := int32(0); i < n; i++ {
		if a[i]&b[i] != 0 {
			return 1
		}
	}
	return 0
}
func isopCheckCubeShift(offset, cube []word, n, v int32, positiveToNegative bool) int32 {
	if v < 6 {
		shift := uint(1 << uint(v))
		for i := int32(0); i < n; i++ {
			moved := cube[i] << shift
			if positiveToNegative {
				moved = cube[i] >> shift
			}
			if offset[i]&moved != 0 {
				return 0
			}
		}
		return 1
	}
	step := int32(1) << uint(v-6)
	for base := int32(0); base < n; base += 2 * step {
		for i := int32(0); i < step; i++ {
			a, b := base+i, base+step+i
			if !positiveToNegative {
				a, b = b, a
			}
			if offset[a]&cube[b] != 0 {
				return 0
			}
		}
	}
	return 1
}
func Abc_TtCheckWithCubePos2Neg(t, c []word, n, v int32) int32 {
	return isopCheckCubeShift(t, c, n, v, true)
}
func Abc_TtCheckWithCubeNeg2Pos(t, c []word, n, v int32) int32 {
	return isopCheckCubeShift(t, c, n, v, false)
}
func isopExpandCube(t []word, n, v int32, positiveToNegative bool) {
	if v < 6 {
		shift := uint(1 << uint(v))
		for i := int32(0); i < n; i++ {
			if positiveToNegative {
				t[i] |= t[i] >> shift
			} else {
				t[i] |= t[i] << shift
			}
		}
		return
	}
	step := int32(1) << uint(v-6)
	for base := int32(0); base < n; base += 2 * step {
		for i := int32(0); i < step; i++ {
			a, b := base+i, base+step+i
			if !positiveToNegative {
				a, b = b, a
			}
			t[a] |= t[b]
		}
	}
}
func Abc_TtExpandCubePos2Neg(t []word, n, v int32) { isopExpandCube(t, n, v, true) }
func Abc_TtExpandCubeNeg2Pos(t []word, n, v int32) { isopExpandCube(t, n, v, false) }

// The subtracting-cubes variant follows the original greedy expansion strategy:
// choose the first uncovered minterm, enlarge against the offset, subtract the
// resulting cube from the onset and repeat. Offset tests use complete words.
func Abc_IsopNew(on, dc, result []word, n int32, limit word, cover []int32) word {
	words := isopWords(n)
	if len(on) < words || len(dc) < words || len(result) < words {
		panic("short greedy ISOP truth")
	}
	remaining := append([]word(nil), on[:words]...)
	offset := make([]word, words)
	for i := range offset {
		if on[i]&^dc[i] != 0 {
			panic("onset outside allowed set")
		}
		offset[i] = ^dc[i]
	}
	clear(result[:words])
	var cost word
	count := 0
	cube := make([]word, words)
	for {
		mint := -1
		for i, w := range remaining {
			if w != 0 {
				mint = i*64 + bits.TrailingZeros64(w)
				break
			}
		}
		if mint < 0 {
			break
		}
		var encoded uint32
		for v := int32(0); v < n; v++ {
			encoded |= 1 << uint(2*v+int32((mint>>uint(v))&1))
		}
		for i := range cube {
			cube[i] = isopCubeWord(encoded, n, i)
		}
		for v := int32(0); v < n; v++ {
			positive := encoded>>uint(2*v)&3 == 2
			if isopCheckCubeShift(offset, cube, int32(words), v, positive) != 0 {
				isopExpandCube(cube, int32(words), v, positive)
				encoded &^= 3 << uint(2*v)
			}
		}
		literals := bits.OnesCount32(encoded)
		cost += Abc_Cube2Cost(1) + word(literals)
		if cost >= limit {
			return limit
		}
		if cover != nil {
			if count >= len(cover) {
				panic("short greedy ISOP cover")
			}
			cover[count] = int32(encoded)
		}
		count++
		for i := range cube {
			result[i] |= cube[i]
			remaining[i] &^= cube[i]
		}
	}
	return cost
}
func Abc_IsopTestNew() {
	t := (isopMask(0) & isopMask(1)) ^ (isopMask(2) & isopMask(3))
	cover := make([]int32, 1000)
	cost := Abc_EsopCheck([]word{t}, 4, Abc_Cube2Cost(ABC_ISOP_MAX_CUBE), cover)
	cover = cover[:Abc_CostCubes(cost)]
	fmt.Printf("Cubes = %d.  Lits = %d.\n", Abc_CostCubes(cost), Abc_CostLits(cost))
	Abc_IsopPrintCover(cover, 4, 0)
	Abc_IsopVerify([]word{t}, 4, make([]word, 1), cover, 1, 0)
}
func Abc_IsopTest(function []word, n int32, cover *[]int32) int32 {
	// Retain the original diagnostic's three enabled computations. The input
	// remains unchanged; the final cover is the preferred-polarity result.
	result := make([]word, isopWords(n))
	limit := Abc_Cube2Cost(ABC_ISOP_MAX_CUBE)
	Abc_IsopCheck(function, function, result, n, limit, nil)
	neg := make([]word, len(result))
	for i := range neg {
		neg[i] = ^function[i]
	}
	Abc_IsopCheck(neg, neg, result, n, limit, nil)
	Abc_Isop(function, n, ABC_ISOP_MAX_CUBE, cover, 1)
	return 1
}
