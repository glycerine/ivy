package extra

// General combinatorial, truth-table and bit-matrix utilities. All returned
// tables are owned by the caller; the C implementation's mutable static caches
// are replaced by local values or immutable lookup data.
import (
	"bufio"
	"cmp"
	"fmt"
	"math"
	"math/bits"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/glycerine/abcgo/misc/extra/npn"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Extra_Base2LogDouble(n float64) int32 {
	if n <= 0 || math.IsNaN(n) || math.IsInf(n, 0) {
		panic("invalid logarithm argument")
	}
	v := math.Log2(n)
	return int32(math.Ceil(v))
}
func Extra_Power2(degree int32) float64 {
	if degree < 0 {
		panic("negative power")
	}
	return math.Ldexp(1, int(degree))
}
func Extra_Power3(n int32) int32 {
	result := int32(1)
	for i := int32(0); i < n; i++ {
		result *= 3
	}
	return result
}
func Extra_NumCombinations(k, n int32) int32 {
	result := int64(1)
	if k < 0 || n < 0 || k > n {
		return 0
	}
	for i := int32(1); i <= min(k, n-k); i++ {
		result = result * int64(n-i+1) / int64(i)
		if result > math.MaxInt32 {
			panic("combination count overflow")
		}
	}
	return int32(result)
}
func Extra_DeriveRadixCode(number, radix, digits int32) []int32 {
	if number < 0 || radix < 2 || digits < 0 || digits >= 100 {
		panic("invalid radix code")
	}
	code := make([]int32, digits)
	for i := range code {
		code[i] = number % radix
		number /= radix
	}
	if number != 0 {
		panic("radix code does not fit")
	}
	return code
}
func Extra_CountOnes(data []byte, n int32) int32 {
	var count int32
	for _, v := range data[:n] {
		count += int32(bits.OnesCount8(v))
	}
	return count
}
func Extra_Factorial(n int32) int32 {
	if n < 0 || n > 12 {
		panic("factorial outside int32 range")
	}
	result := int32(1)
	for i := int32(2); i <= n; i++ {
		result *= i
	}
	return result
}
func Extra_Permutations(n int32) [][]byte {
	if n < 0 || n > 10 {
		panic("invalid materialized permutation width")
	}
	count := Extra_Factorial(n)
	result := make([][]byte, count)
	for i := range result {
		result[i] = make([]byte, n)
	}
	array := make([]byte, n)
	for i := range array {
		array[i] = byte(i)
	}
	Extra_Permutations_rec(result, count, n, array)
	return result
}
func Extra_Permutations_rec(result [][]byte, factorial, n int32, array []byte) {
	if n == 0 {
		return
	}
	if n == 1 {
		result[0][0] = array[0]
		return
	}
	next := factorial / n
	for current := int32(0); current < n; current++ {
		array[current], array[n-1] = array[n-1], array[current]
		part := result[(n-1-current)*next:]
		for k := int32(0); k < next; k++ {
			part[k][n-1] = array[n-1]
		}
		Extra_Permutations_rec(part, next, n-1, array)
		array[current], array[n-1] = array[n-1], array[current]
	}
}
func miscCheckPermutation(perm []byte, n int32) {
	if n < 0 || n > 31 || len(perm) < int(n) {
		panic("invalid truth permutation size")
	}
	var seen uint32
	for _, v := range perm[:n] {
		if int32(v) >= n || seen>>v&1 != 0 {
			panic("invalid truth permutation")
		}
		seen |= 1 << v
	}
}
func Extra_TruthPermute_int(minterms []int32, count int32, perm []byte, n int32, out []int32) {
	miscCheckPermutation(perm, n)
	input := slices.Clone(minterms[:count])
	for i, m := range input {
		var value int32
		for v, dest := range perm[:n] {
			if m>>v&1 != 0 {
				value |= 1 << dest
			}
		}
		out[i] = value
	}
}
func Extra_TruthPermute(truth uint32, perm []byte, n, reverse int32) uint32 {
	if n < 0 || n > 5 {
		panic("word truth requires 0..5 variables")
	}
	miscCheckPermutation(perm, n)
	var result uint32
	for m := uint32(0); m < 1<<n; m++ {
		var dest uint32
		for v, p := range perm[:n] {
			dest |= (m >> v & 1) << p
		}
		a, b := m, dest
		if reverse != 0 {
			a, b = b, a
		}
		result |= (truth >> a & 1) << b
	}
	return result
}
func Extra_TruthPolarize(truth uint32, phase, n int32) uint32 {
	if n < 0 || n > 5 || phase < 0 || phase >= 1<<n {
		panic("invalid word truth phase")
	}
	var result uint32
	for m := uint32(0); m < 32; m++ {
		result |= (truth >> (m ^ uint32(phase)) & 1) << m
	}
	return result
}
func miscTruthMask(n int32) uint32 {
	if n < 0 || n > 5 {
		panic("invalid word truth width")
	}
	return ^uint32(0) >> uint(32-(1<<n))
}
func Extra_TruthCanonN(truth uint32, n int32) uint32 {
	miscTruthMask(n)
	best := ^uint32(0)
	for phase := int32(0); phase < 1<<n; phase++ {
		best = min(best, Extra_TruthPolarize(truth, phase, n))
	}
	return best
}
func Extra_TruthCanonNN(truth uint32, n int32) uint32 {
	return min(Extra_TruthCanonN(truth, n), Extra_TruthCanonN(^truth&miscTruthMask(n), n))
}
func miscTruthCanon(truth uint32, n int32, input, output bool) uint32 {
	mask := miscTruthMask(n)
	best := ^uint32(0)
	phases := int32(1)
	if input {
		phases = 1 << n
	}
	permutations := Extra_Permutations(n)
	for phase := int32(0); phase < phases; phase++ {
		for _, perm := range permutations {
			best = min(best, Extra_TruthPermute(Extra_TruthPolarize(truth, phase, n), perm, n, 0))
			if output {
				best = min(best, Extra_TruthPermute(Extra_TruthPolarize(^truth&mask, phase, n), perm, n, 0))
			}
		}
	}
	return best
}
func Extra_TruthCanonP(t uint32, n int32) uint32             { return miscTruthCanon(t, n, false, false) }
func Extra_TruthCanonNP(t uint32, n int32) uint32            { return miscTruthCanon(t, n, true, false) }
func Extra_TruthCanonNPN(t uint32, n int32) uint32           { return miscTruthCanon(t, n, true, true) }
func Extra_Truth4VarNPN() ([]uint16, []byte, []byte, []byte) { return npn.Extra_Truth4VarNPN() }
func miscNClasses(n, limit int32) ([]uint32, [][]byte, []byte) {
	if limit < 1 || limit > 1<<n {
		panic("invalid phase limit")
	}
	count := 1 << uint(1<<n)
	canons := make([]uint32, count)
	phases := make([][]byte, count)
	counters := make([]byte, count)
	assigned := make([]bool, count)
	for truth := uint32(0); int(truth) < count; truth++ {
		if assigned[truth] {
			continue
		}
		for phase := int32(0); phase < 1<<n; phase++ {
			value := Extra_TruthPolarize(truth, phase, n)
			if !assigned[value] {
				assigned[value] = true
				canons[value] = truth
			}
			if int32(len(phases[value])) < limit {
				phases[value] = append(phases[value], byte(phase))
				counters[value]++
			}
		}
	}
	return canons, phases, counters
}
func Extra_Truth3VarN() ([]uint32, [][]byte, []byte) {
	c, p, n := miscNClasses(3, 8)
	for i := range c {
		c[i] *= 0x01010101
	}
	return c, p, n
}
func Extra_Truth4VarN(limit int32) ([]uint16, [][]byte, []byte) {
	c, p, n := miscNClasses(4, limit)
	result := make([]uint16, len(c))
	for i, v := range c {
		result[i] = uint16(v)
	}
	return result, p, n
}
func Extra_ArrayAlloc(cols, rows, size int32) [][]byte {
	if cols <= 0 || rows <= 0 || size <= 0 || int64(cols) > int64(math.MaxInt)/(int64(rows)*int64(size)) {
		panic("invalid array dimensions")
	}
	data := make([]byte, int(cols)*int(rows)*int(size))
	result := make([][]byte, cols)
	stride := int(rows) * int(size)
	for i := range result {
		result[i] = data[i*stride : (i+1)*stride : (i+1)*stride]
	}
	return result
}

// Support expansion moves the first popcount(phase) input variables into phase
// positions; remaining variables retain their relative order.
func miscExpandPermutation(n int32, phase uint32) []byte {
	perm := make([]byte, n)
	one, zero := byte(0), byte(bits.OnesCount32(phase))
	for v := range perm {
		if phase>>v&1 != 0 {
			perm[v] = one
			one++
		} else {
			perm[v] = zero
			zero++
		}
	}
	return perm
}
func miscExpandWord(truth uint64, n int32, phase uint32) uint64 {
	if phase >= 1<<n {
		panic("invalid expanded support")
	}
	if phase&(phase+1) == 0 {
		return truth
	}
	// Preserve the source's elementary-variable shortcut and its precondition:
	// one-bit support denotes the positive elementary variable.
	if bits.OnesCount32(phase) == 1 {
		return miscTruth6[bits.TrailingZeros32(phase)]
	}
	perm := miscExpandPermutation(n, phase)
	switch n {
	case 4:
		perm = miscPerm4[phase][:]
	case 5:
		perm = miscPerm5[phase][:]
	case 6:
		perm = miscPerm6[phase][:]
	}
	var result uint64
	for m := uint32(0); m < 1<<n; m++ {
		var dest uint32
		for v, p := range perm {
			dest |= (m >> p & 1) << v
		}
		result |= (truth >> m & 1) << dest
	}
	return result
}
func Extra_TruthPerm4One(t uint32, phase int32) uint16 {
	if phase < 0 {
		panic("negative support")
	}
	return uint16(miscExpandWord(uint64(t), 4, uint32(phase)))
}
func Extra_TruthPerm5One(t uint32, phase int32) uint32 {
	if phase < 0 {
		panic("negative support")
	}
	return uint32(miscExpandWord(uint64(t), 5, uint32(phase)))
}
func Extra_TruthPerm6One(t []uint32, phase int32, out []uint32) {
	if phase < 0 {
		panic("negative support")
	}
	value := miscExpandWord(uint64(t[0])|uint64(t[1])<<32, 6, uint32(phase))
	out[0], out[1] = uint32(value), uint32(value>>32)
}
func Extra_TruthExpand(n, words int32, truth []uint32, phase uint32, out []uint32) {
	expected := int32(1)
	if n > 5 && n <= 8 {
		expected <<= n - 5
	}
	if n < 1 || n > 8 || words != expected || phase == 0 || phase >= 1<<n {
		panic("invalid expanded truth dimensions")
	}
	input := slices.Clone(truth[:words])
	clear(out[:words])
	perm := miscExpandPermutation(n, phase)
	if phase&(phase+1) == 0 {
		copy(out, input)
		return
	}
	if bits.OnesCount32(phase) == 1 {
		v := bits.TrailingZeros32(phase)
		for m := 0; m < 32*int(words); m++ {
			if m>>v&1 != 0 {
				out[m>>5] |= 1 << uint(m&31)
			}
		}
		return
	}
	for m := uint32(0); m < 1<<n; m++ {
		var dest uint32
		for v, p := range perm {
			dest |= (m >> p & 1) << v
		}
		if input[m>>5]>>(m&31)&1 != 0 {
			out[dest>>5] |= 1 << (dest & 31)
		}
	}
}
func Extra_TruthPerm43() [][]uint16 {
	result := make([][]uint16, 256)
	for i := range result {
		result[i] = make([]uint16, 16)
		for p := range result[i] {
			result[i][p] = Extra_TruthPerm4One(uint32(i)*0x101, int32(p))
		}
	}
	return result
}
func Extra_TruthPerm53() [][]uint32 {
	result := make([][]uint32, 256)
	for i := range result {
		result[i] = make([]uint32, 32)
		for p := range result[i] {
			result[i][p] = Extra_TruthPerm5One(uint32(i)*0x1010101, int32(p))
		}
	}
	return result
}
func Extra_TruthPerm54() [][]uint32 {
	result := make([][]uint32, 65536)
	for i := range result {
		result[i] = make([]uint32, 4)
		for p, phase := range []int32{23, 27, 29, 30} {
			result[i][p] = Extra_TruthPerm5One(uint32(i)*0x10001, phase)
		}
	}
	return result
}
func Extra_TruthPerm63() [][]uint64 {
	result := make([][]uint64, 256)
	for i := range result {
		result[i] = make([]uint64, 64)
		for p := range result[i] {
			result[i][p] = miscExpandWord(uint64(i)*0x0101010101010101, 6, uint32(p))
		}
	}
	return result
}
func Extra_Truths8() [][]uint32 {
	result := make([][]uint32, 8)
	for v := range result {
		result[v] = make([]uint32, 8)
		for bit := 0; bit < 256; bit++ {
			if bit>>v&1 != 0 {
				result[v][bit>>5] |= 1 << uint(bit&31)
			}
		}
	}
	return result
}
func Extra_BubbleSort(order, costs []int32, n, increasing int32) {
	for i := range order[:n] {
		order[i] = int32(i)
	}
	slices.SortStableFunc(order[:n], func(a, b int32) int {
		if increasing != 0 {
			return cmp.Compare(costs[a], costs[b])
		}
		return cmp.Compare(costs[b], costs[a])
	})
}
func Extra_TruthExpandGeneratePermTable() {
	fmt.Print("\nstatic char Cases[256] = {\n")
	for i := uint32(0); i < 256; i++ {
		value := -1
		if i&(i+1) == 0 {
			value = 0
		} else if bits.OnesCount32(i) == 1 {
			value = bits.TrailingZeros32(i)
		}
		separator := ","
		if i == 255 {
			separator = " "
		}
		fmt.Printf("    %d%s // %08b\n", value, separator, i)
	}
	fmt.Print("};\n\nstatic char Perms[256][8] = {\n")
	for i := uint32(0); i < 256; i++ {
		fmt.Print("    {")
		for v, p := range miscExpandPermutation(8, i) {
			if v != 0 {
				fmt.Print(",")
			}
			fmt.Printf(" %d", p)
		}
		separator := ","
		if i == 255 {
			separator = " "
		}
		fmt.Printf(" }%s // %08b\n", separator, i)
	}
	fmt.Println("};")
}
func Extra_GreyCodeSchedule(n int32) []int32 {
	if n < 1 || n > 24 {
		panic("invalid Gray code width")
	}
	result := make([]int32, 1<<n)
	at := 0
	for k := int32(0); k < n; k++ {
		result[at] = k
		at++
		for i := 0; i < (1<<k)-1; i++ {
			result[at] = result[i]
			at++
		}
	}
	result[at] = n - 1
	return result
}
func Extra_PermSchedule(n int32) []int32 {
	if n < 1 || n > 10 {
		panic("invalid permutation schedule width")
	}
	if n == 1 {
		return []int32{0}
	}
	if n == 2 {
		return []int32{0, 0}
	}
	result := make([]int32, 0, Extra_Factorial(n))
	previous := Extra_PermSchedule(n - 1)
	for k := int32(0); k < Extra_Factorial(n)/n/2; k++ {
		for i := n - 2; i >= 0; i-- {
			result = append(result, i)
		}
		result = append(result, previous[2*k]+1)
		for i := int32(0); i < n-1; i++ {
			result = append(result, i)
		}
		result = append(result, previous[2*k+1])
	}
	return result
}

var miscTruth6 = [6]uint64{0xaaaaaaaaaaaaaaaa, 0xcccccccccccccccc, 0xf0f0f0f0f0f0f0f0, 0xff00ff00ff00ff00, 0xffff0000ffff0000, 0xffffffff00000000}

func Extra_Truth6SwapAdjacent(t uint64, v int32) uint64 {
	data := []uint64{t}
	util.Abc_TtSwapAdjacent(data, 1, v)
	return data[0]
}
func Extra_Truth6ChangePhase(t uint64, v int32) uint64 {
	if v < 0 || v > 5 {
		panic("invalid six-variable phase")
	}
	mask := miscTruth6[v]
	return (t&^mask)<<(1<<v) | (t&mask)>>(1<<v)
}
func Extra_Truth6MinimumExact(t uint64, comp, perm []int32) uint64 {
	best := ^uint64(0)
	for polarity := 0; polarity < 2; polarity++ {
		current := t
		if polarity != 0 {
			current = ^t
		}
		start := current
		for _, swap := range perm[:720] {
			before := current
			for _, flip := range comp[:64] {
				best = min(best, current)
				current = Extra_Truth6ChangePhase(current, flip)
			}
			if current != before {
				panic("phase schedule does not close")
			}
			current = Extra_Truth6SwapAdjacent(current, swap)
		}
		if current != start {
			panic("permutation schedule does not close")
		}
	}
	return best
}
func Extra_Truth6Ones(t uint64) int32 { return int32(bits.OnesCount64(t)) }
func Extra_Truth6MinimumRoundOne(t uint64, v int32) uint64 {
	if v < 0 || v > 4 {
		panic("invalid minimum round")
	}
	best := t
	for swapped := 0; swapped < 2; swapped++ {
		if swapped != 0 {
			t = Extra_Truth6SwapAdjacent(t, v)
		}
		for phase := 0; phase < 4; phase++ {
			current := t
			if phase&1 != 0 {
				current = Extra_Truth6ChangePhase(current, v)
			}
			if phase&2 != 0 {
				current = Extra_Truth6ChangePhase(current, v+1)
			}
			best = min(best, current)
		}
	}
	return best
}
func Extra_Truth6MinimumRoundMany(t uint64) uint64 {
	for round := 0; round < 10; round++ {
		before := t
		for v := int32(4); v >= 0; v-- {
			t = Extra_Truth6MinimumRoundOne(t, v)
		}
		if t == before {
			break
		}
	}
	return t
}
func Extra_Truth6MinimumHeuristic(t uint64) uint64 {
	count := bits.OnesCount64(t)
	if count < 32 {
		return Extra_Truth6MinimumRoundMany(t)
	}
	if count > 32 {
		return Extra_Truth6MinimumRoundMany(^t)
	}
	return min(Extra_Truth6MinimumRoundMany(t), Extra_Truth6MinimumRoundMany(^t))
}
func Extra_Truth6MinimumHeuristicTest() {
	Extra_Truth6MinimumRoundMany(0x5555555555555555 &^ (uint64(0x3333333333333333) & 0x0f0f0f0f0f0f0f0f))
}
func Extra_NpnRead(name string, count int32) ([]uint64, error) {
	if count < 0 {
		return nil, fmt.Errorf("negative function count")
	}
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	result := make([]uint64, 0, count)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.TrimPrefix(line, "0x")
		value, e := strconv.ParseUint(line, 16, 64)
		if e != nil {
			return nil, e
		}
		if len(result) == int(count) {
			return nil, fmt.Errorf("too many truth tables")
		}
		result = append(result, value)
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	if len(result) != int(count) {
		return nil, fmt.Errorf("expected %d truth tables, got %d", count, len(result))
	}
	return result, nil
}
func CompareWords(a, b uint64) int { return cmp.Compare(a, b) }
func Extra_NpnTest1()              { Extra_PermSchedule(5) }
func Extra_NpnTest2() {
	truth := uint64(0xa2222aaa08888000)
	best := Extra_Truth6MinimumExact(truth, Extra_GreyCodeSchedule(6), Extra_PermSchedule(6))
	fmt.Printf("%016x\n%016x\n", truth, best)
}
func Extra_NpnTest() error {
	started := time.Now()
	truths, err := Extra_NpnRead(`C:\_projects\abc\_TEST\allan\test.txt`, 10)
	if err != nil {
		return err
	}
	for i, t := range truths {
		truths[i] = Extra_Truth6MinimumExact(t, Extra_GreyCodeSchedule(6), Extra_PermSchedule(6))
	}
	fmt.Println("Finished deriving minimum form")
	for i, t := range truths {
		fmt.Printf("Line %d : %016x\n", i, t)
	}
	fmt.Printf("Time = %v\n", time.Since(started))
	return nil
}
func Extra_NtkPrintBin(t []uint64, n int32) {
	if n < 0 || n > 64 {
		panic("invalid binary width")
	}
	for bit := n - 1; bit >= 0; bit-- {
		fmt.Printf("%d", t[0]>>bit&1)
	}
}
func Extra_NtkPowerTest() {
	for i := uint64(0); i < 16; i++ {
		for j := uint64(0); j < 16; j++ {
			t := i
			for k := uint64(1); k < j; k++ {
				t *= i
			}
			fmt.Printf("%04b%04b %064b\n", i, j, t)
		}
	}
}
func Extra_Transpose64Simple(a, b []uint64) {
	input := slices.Clone(a[:64])
	clear(b[:64])
	for i, v := range input {
		for k := 0; k < 64; k++ {
			b[k] |= (v >> k & 1) << uint(63-i)
		}
	}
}
func Extra_Transpose32(a []uint32) {
	a = a[:32]
	mask := uint32(0x0000ffff)
	for shift := uint(16); shift != 0; {
		for k := 0; k < 32; k = ((k | int(shift)) + 1) &^ int(shift) {
			t := (a[k] ^ (a[k|int(shift)] >> shift)) & mask
			a[k] ^= t
			a[k|int(shift)] ^= t << shift
		}
		shift >>= 1
		mask ^= mask << shift
	}
}
func Extra_Transpose64(a []uint64) {
	a = a[:64]
	mask := uint64(0xffffffff)
	for shift := uint(32); shift != 0; {
		for k := 0; k < 64; k = (k + int(shift) + 1) &^ int(shift) {
			t := (a[k] ^ (a[k+int(shift)] >> shift)) & mask
			a[k] ^= t
			a[k+int(shift)] ^= t << shift
		}
		shift >>= 1
		mask ^= mask << shift
	}
}
func Extra_Transpose64p(a [][]uint64) {
	var temp [64]uint64
	for i := range temp {
		temp[i] = a[i][0]
	}
	Extra_Transpose64(temp[:])
	for i, v := range temp {
		a[i][0] = v
	}
}
func miscTranspose(input []uint64, inWords int32, output []uint64, outWords int32) {
	if inWords <= 0 || outWords <= 0 || int64(len(input)) != 64*int64(inWords)*int64(outWords) || len(input) != len(output) {
		panic("invalid bit matrix dimensions")
	}
	input = slices.Clone(input)
	clear(output)
	for row := 0; row < 64*int(outWords); row++ {
		for col := 0; col < 64*int(inWords); col++ {
			if input[row*int(inWords)+col/64]>>uint(col&63)&1 != 0 {
				output[col*int(outWords)+row/64] |= uint64(1) << uint(row&63)
			}
		}
	}
}
func Extra_BitMatrixTransposeSimple(in *vec.Vec_Wrd_t, inWords int32, out *vec.Vec_Wrd_t, outWords int32) {
	miscTranspose(vec.Vec_WrdArray(in), inWords, vec.Vec_WrdArray(out), outWords)
}
func Extra_BitMatrixTransposeP(in *vec.Vec_Wrd_t, inWords int32, out *vec.Vec_Wrd_t, outWords int32) {
	input := slices.Clone(vec.Vec_WrdArray(in))
	output := vec.Vec_WrdArray(out)
	if inWords <= 0 || outWords <= 0 || len(input) != len(output) || int64(len(input)) != 64*int64(inWords)*int64(outWords) {
		panic("invalid bit matrix dimensions")
	}
	for x := int32(0); x < outWords; x++ {
		for y := int32(0); y < inWords; y++ {
			var block [64]uint64
			for i := int32(0); i < 64; i++ {
				block[i] = input[(64*x+63-i)*inWords+y]
			}
			Extra_Transpose64(block[:])
			for i := int32(0); i < 64; i++ {
				output[(64*y+63-i)*outWords+x] = block[i]
			}
		}
	}
}
func Extra_BitMatrixTransposePP(in *vec.Vec_Ptr_t, inWords int32, out *vec.Vec_Wrd_t, outWords int32) {
	input := vec.Vec_PtrArray(in)
	flat := make([]uint64, 0, len(input)*int(inWords))
	for _, row := range input {
		words, ok := row.([]uint64)
		if !ok || len(words) != int(inWords) {
			panic("invalid pointer matrix row")
		}
		flat = append(flat, words...)
	}
	miscTranspose(flat, inWords, vec.Vec_WrdArray(out), outWords)
}
func Extra_BitMatrixShow(matrix *vec.Vec_Wrd_t, words int32) {
	if words <= 0 || vec.Vec_WrdSize(matrix)%int(words) != 0 {
		panic("invalid displayed matrix")
	}
	for row := 0; row < vec.Vec_WrdSize(matrix)/int(words); row++ {
		if row%64 == 0 {
			fmt.Println()
		}
		for _, v := range vec.Vec_WrdArray(matrix)[row*int(words) : (row+1)*int(words)] {
			fmt.Printf("%064b ", v)
		}
		fmt.Println()
	}
	fmt.Println()
}
func Extra_BitMatrixTransposeTest() {
	in := vec.Vec_WrdStart(64 * 100 * 200)
	for i := range vec.Vec_WrdArray(in) {
		vec.Vec_WrdArray(in)[i] = uint64(i) * 0x9e3779b97f4a7c15
	}
	a, b := vec.Vec_WrdStart(vec.Vec_WrdSize(in)), vec.Vec_WrdStart(vec.Vec_WrdSize(in))
	Extra_BitMatrixTransposeP(in, 100, a, 200)
	Extra_BitMatrixTransposeSimple(in, 100, b, 200)
	if slices.Equal(vec.Vec_WrdArray(a), vec.Vec_WrdArray(b)) {
		fmt.Println("Verification succeeded.")
	} else {
		fmt.Println("Verification failed.")
	}
}
