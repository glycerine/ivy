package extra

import (
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/glycerine/abcgo/misc/vec"
)

func Abc_SuppCountOnes(v uint32) int32   { return int32(bits.OnesCount32(v)) }
func Abc_SuppCountOnes64(v uint64) int32 { return int32(bits.OnesCount64(v)) }
func Abc_SuppGen(m, n int32) *vec.Vec_Wrd_t {
	if n < 0 || n > 30 || m < 0 || m > n {
		panic("invalid fixed-weight support dimensions")
	}
	result := vec.Vec_WrdAlloc(1000)
	if m == 0 {
		vec.Vec_WrdPush(result, 0)
		return result
	}
	// Gosper's successor generates the same numeric order as the source's full
	// 2^n scan while visiting only the requested fixed-weight vectors.
	for value := uint64(1)<<m - 1; value < uint64(1)<<n; {
		vec.Vec_WrdPush(result, value)
		low := value & -value
		next := value + low
		value = next | ((value^next)>>2)/low
	}
	return result
}
func suppDimensions(n int32) {
	if n < 1 || n > 63 {
		panic("support dimension requires 1..63 variables")
	}
}
func suppVerify(p *vec.Vec_Wrd_t, matrix []uint64, n, rows int32) int {
	if n < 1 || n > 64 || rows < 0 || rows > 64 {
		panic("invalid support verification dimensions")
	}
	frequency := map[uint64]int{}
	collisions := 0
	for _, value := range vec.Vec_WrdArray(p) {
		var encoded uint64
		for row, mask := range matrix[:rows] {
			encoded |= uint64(bits.OnesCount64(value&mask)&1) << uint(row)
		}
		collisions += frequency[encoded]
		frequency[encoded]++
	}
	return collisions
}
func Abc_SuppVerify(p *vec.Vec_Wrd_t, matrix []uint64, n, rows int32) {
	if count := suppVerify(p, matrix, n, rows); count != 0 {
		fmt.Printf("The total of %d pairs fail verification.\n", count)
	} else {
		fmt.Println("Verification successful.")
	}
}
func Abc_SuppGenPairs(p *vec.Vec_Wrd_t, n int32) *vec.Vec_Wrd_t {
	suppDimensions(n)
	result := vec.Vec_WrdAlloc(1000)
	seen := map[uint64]bool{}
	data := vec.Vec_WrdArray(p)
	for i, a := range data {
		for _, b := range data[i+1:] {
			value := a ^ b
			if value>>n != 0 {
				panic("support vector exceeds dimension")
			}
			if !seen[value] {
				seen[value] = true
				vec.Vec_WrdPush(result, value)
			}
		}
	}
	return result
}
func Abc_SuppGenPairs2(ones, n int32) *vec.Vec_Wrd_t {
	if n < 1 || n > 30 || ones < 0 {
		panic("invalid support pair dimensions")
	}
	result := vec.Vec_WrdAlloc(1000)
	for value := uint64(0); value < uint64(1)<<n; value++ {
		count := bits.OnesCount64(value)
		if count > 0 && count%2 == 0 && count <= 2*int(ones) {
			vec.Vec_WrdPush(result, value)
		}
	}
	return result
}
func Abc_SuppPrintMask(mask uint64, n int32) {
	for i := int32(0); i < n; i++ {
		fmt.Printf("%d", mask>>i&1)
	}
	fmt.Println()
}
func Abc_SuppGenProfile(p *vec.Vec_Wrd_t, n int32, counts []int32) {
	suppDimensions(n)
	for _, value := range vec.Vec_WrdArray(p) {
		marker := value >> n & 1
		for k := int32(0); k < n; k++ {
			counts[k] += int32(value>>k&1 ^ marker)
		}
	}
}
func Abc_SuppPrintProfile(p *vec.Vec_Wrd_t, n int32) {
	counts := make([]int32, n)
	Abc_SuppGenProfile(p, n, counts)
	for k, count := range counts {
		fmt.Printf("%2d : %6d  %6.2f %%\n", k, count, 100*float64(count)/float64(vec.Vec_WrdSize(p)))
	}
}
func Abc_SuppGenFindBest(p *vec.Vec_Wrd_t, n int32) (int32, int32) {
	counts := make([]int32, n)
	Abc_SuppGenProfile(p, n, counts)
	best := int32(0)
	for k := int32(1); k < n; k++ {
		if counts[best] < counts[k] {
			best = k
		}
	}
	return best, counts[best]
}
func Abc_SuppGenSelectVar(p *vec.Vec_Wrd_t, n, v int32) {
	suppDimensions(n)
	if v < 0 || v >= n {
		panic("invalid support variable")
	}
	for i, value := range vec.Vec_WrdArray(p) {
		if value>>v&1 != 0 {
			vec.Vec_WrdArray(p)[i] ^= uint64(1) << n
		}
	}
}
func Abc_SuppGenFilter(p *vec.Vec_Wrd_t, n int32) {
	at := 0
	for _, value := range vec.Vec_WrdArray(p) {
		if value>>n&1 == 0 {
			vec.Vec_WrdWriteEntry(p, at, value)
			at++
		}
	}
	vec.Vec_WrdShrink(p, at)
}
func Abc_SuppFindOne(p *vec.Vec_Wrd_t, n int32) uint64 {
	mask := uint64(0)
	previous := int32(-1)
	for {
		v, merit := Abc_SuppGenFindBest(p, n)
		if previous >= merit {
			break
		}
		previous = merit
		Abc_SuppGenSelectVar(p, n, v)
		mask ^= uint64(1) << v
	}
	return mask
}
func Abc_SuppMinimize(matrix []uint64, p *vec.Vec_Wrd_t, n, verbose int32) int32 {
	suppDimensions(n)
	row := int32(0)
	for vec.Vec_WrdSize(p) > 0 {
		before := vec.Vec_WrdSize(p)
		matrix[row] = Abc_SuppFindOne(p, n)
		Abc_SuppGenFilter(p, n)
		if vec.Vec_WrdSize(p) == before {
			panic("zero difference vector cannot be separated")
		}
		if verbose != 0 {
			fmt.Printf("%2d : %6d  ", row, vec.Vec_WrdSize(p))
			Abc_SuppPrintMask(matrix[row], n)
		}
		row++
	}
	return row
}
func Abc_SuppTest(ones, n, simple, check, verbose int32) {
	started := time.Now()
	vectors := Abc_SuppGen(ones, n)
	pairs := Abc_SuppGenPairs(vectors, n)
	if simple != 0 {
		pairs = Abc_SuppGenPairs2(ones, n)
	}
	fmt.Printf("M = %2d N = %2d K = %6d Distinct = %8d Time = %v\n", ones, n, vec.Vec_WrdSize(vectors), vec.Vec_WrdSize(pairs), time.Since(started))
	var matrix [64]uint64
	rows := Abc_SuppMinimize(matrix[:], pairs, n, verbose)
	fmt.Printf("Solution with %d variables found.\n", rows)
	if check != 0 {
		Abc_SuppVerify(vectors, matrix[:], n, rows)
	}
}
func Abc_SuppReadMin(name string) (*vec.Vec_Wrd_t, int32, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, 0, err
	}
	text := string(data)
	start := strings.Index(text, "INPUT F-COVER")
	if start < 0 {
		return nil, 0, fmt.Errorf("missing INPUT F-COVER")
	}
	tokens := strings.FieldsFunc(text[start+len("INPUT F-COVER"):], func(r rune) bool { return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == ',' })
	if len(tokens) == 0 {
		return nil, 0, fmt.Errorf("missing cube count")
	}
	count, e := strconv.Atoi(tokens[0])
	if e != nil || count < 1 || count > 1000000 {
		return nil, 0, fmt.Errorf("invalid cube count")
	}
	result := vec.Vec_WrdAlloc(count)
	var value uint64
	vars, width := int32(0), int32(-1)
	finish := func() error {
		if width == -1 {
			width = vars
		}
		if vars != width {
			return fmt.Errorf("inconsistent cube width")
		}
		vec.Vec_WrdPush(result, value)
		value = 0
		vars = 0
		return nil
	}
	for _, token := range tokens[1:] {
		if strings.HasPrefix(token, "INPUT") {
			break
		}
		if len(token) > 2 {
			if err = finish(); err != nil {
				return nil, 0, err
			}
			continue
		}
		if vars >= 64 || token != "10" && token != "01" {
			return nil, 0, fmt.Errorf("invalid MIN literal %q", token)
		}
		if token == "10" {
			value |= uint64(1) << vars
		}
		vars++
	}
	if vars != 0 {
		if err = finish(); err != nil {
			return nil, 0, err
		}
	}
	if vec.Vec_WrdSize(result) != count {
		return nil, 0, fmt.Errorf("expected %d cubes, got %d", count, vec.Vec_WrdSize(result))
	}
	return result, width, nil
}
func Abc_SuppDiffMatrix(cubes *vec.Vec_Wrd_t) *vec.Vec_Wrd_t {
	result := vec.Vec_WrdAlloc(1000)
	data := vec.Vec_WrdArray(cubes)
	for i, a := range data {
		for _, b := range data[i+1:] {
			vec.Vec_WrdPush(result, a^b)
		}
	}
	return result
}
func Abc_SuppFindVar(singles, doubles *vec.Vec_Wec_t, n int32) int32 {
	return suppFindVar(singles, doubles, n, nil)
}
func suppFindVar(singles, doubles *vec.Vec_Wec_t, n int32, removed []bool) int32 {
	best := int32(-1)
	for v := int32(0); v < n; v++ {
		if removed != nil && removed[v] || vec.Vec_IntSize(vec.Vec_WecEntry(singles, int(v))) != 0 {
			continue
		}
		if best == -1 || vec.Vec_IntSize(vec.Vec_WecEntry(doubles, int(v))) < vec.Vec_IntSize(vec.Vec_WecEntry(doubles, int(best))) {
			best = v
		}
	}
	return best
}
func Abc_SuppRemove(p *vec.Vec_Wrd_t, counts []int32, singles, doubles *vec.Vec_Wec_t, v, n int32) {
	if vec.Vec_IntSize(vec.Vec_WecEntry(singles, int(v))) != 0 {
		panic("cannot remove essential support variable")
	}
	vec.Vec_IntClear(vec.Vec_WecEntry(doubles, int(v)))
	for row, value := range vec.Vec_WrdArray(p) {
		if value>>v&1 == 0 {
			continue
		}
		value &^= uint64(1) << v
		vec.Vec_WrdArray(p)[row] = value
		counts[row]--
		if counts[row] == 1 {
			for k := int32(0); k < n; k++ {
				if value>>k&1 != 0 {
					vec.Vec_IntRemove(vec.Vec_WecEntry(doubles, int(k)), int32(row))
					vec.Vec_WecPush(singles, int(k), int32(row))
				}
			}
		} else if counts[row] == 2 {
			for k := int32(0); k < n; k++ {
				if value>>k&1 != 0 {
					vec.Vec_WecPush(doubles, int(k), int32(row))
				}
			}
		}
	}
}
func Abc_SuppProfile(singles, doubles *vec.Vec_Wec_t, n int32) {
	for v := int32(0); v < n; v++ {
		fmt.Printf("%2d : S = %3d D = %3d\n", v, vec.Vec_IntSize(vec.Vec_WecEntry(singles, int(v))), vec.Vec_IntSize(vec.Vec_WecEntry(doubles, int(v))))
	}
}
func Abc_SuppSolve(p *vec.Vec_Wrd_t, n int32) int32 {
	if n < 1 || n > 64 {
		panic("invalid difference matrix dimension")
	}
	copyMatrix := vec.Vec_WrdAlloc(vec.Vec_WrdSize(p))
	for _, v := range vec.Vec_WrdArray(p) {
		if v == 0 || n < 64 && v>>n != 0 {
			panic("invalid difference matrix row")
		}
		vec.Vec_WrdPush(copyMatrix, v)
	}
	singles, doubles := vec.Vec_WecStart(int(n)), vec.Vec_WecStart(int(n))
	counts := make([]int32, vec.Vec_WrdSize(p))
	for row, value := range vec.Vec_WrdArray(p) {
		counts[row] = int32(bits.OnesCount64(value))
		if counts[row] <= 2 {
			target := doubles
			if counts[row] == 1 {
				target = singles
			}
			for v := int32(0); v < n; v++ {
				if value>>v&1 != 0 {
					vec.Vec_WecPush(target, int(v), int32(row))
				}
			}
		}
	}
	removed := make([]bool, n)
	remaining := n
	for {
		v := suppFindVar(singles, doubles, n, removed)
		if v == -1 {
			break
		}
		Abc_SuppRemove(copyMatrix, counts, singles, doubles, v, n)
		removed[v] = true
		remaining--
	}
	return remaining
}
func Abc_SuppReadMinTest(name string) error {
	cubes, n, err := Abc_SuppReadMin(name)
	if err != nil {
		return err
	}
	rows := Abc_SuppDiffMatrix(cubes)
	fmt.Printf("Solution with %d variables found.\n", Abc_SuppSolve(rows, n))
	return nil
}
