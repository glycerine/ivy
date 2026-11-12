package abc

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"strconv"
	"strings"
)

// SOP strings are immutable. Constructors allocate Go-owned storage; the
// source arena argument remains for callers sharing the original interface.
func Abc_SopRegister(_ *mem.Mem_Flex_t, s string) string { return strings.Clone(s) }
func Abc_SopStart(_ *mem.Mem_Flex_t, cubes, variables int32) string {
	if cubes < 0 || variables < 0 {
		panic("abc: negative SOP dimensions")
	}
	return strings.Repeat(strings.Repeat("-", int(variables))+" 1\n", int(cubes))
}
func Abc_SopCreateConst1(m *mem.Mem_Flex_t) string { return Abc_SopRegister(m, " 1\n") }
func Abc_SopCreateConst0(m *mem.Mem_Flex_t) string { return Abc_SopRegister(m, " 0\n") }
func Abc_SopCreateAnd2(m *mem.Mem_Flex_t, a, b int32) string {
	return Abc_SopCreateAnd(m, 2, []int32{a, b})
}
func sopSingle(variables int32, compl []int32, positive bool, tag byte) string {
	if variables < 0 || compl != nil && len(compl) < int(variables) {
		panic("abc: invalid SOP variable count")
	}
	row := []byte(Abc_SopStart(nil, 1, variables))
	for i := int32(0); i < variables; i++ {
		c := int32(0)
		if compl != nil {
			c = compl[i]
		}
		if c < 0 || c > 1 {
			panic("abc: complement flag is not binary")
		}
		row[i] = '0' + byte(c)
		if positive {
			row[i] = '1' - byte(c)
		}
	}
	row[variables+1] = tag
	return string(row)
}
func Abc_SopCreateAnd(m *mem.Mem_Flex_t, n int32, c []int32) string {
	return sopSingle(n, c, true, '1')
}
func Abc_SopCreateNand(m *mem.Mem_Flex_t, n int32) string { return sopSingle(n, nil, true, '0') }
func Abc_SopCreateOr(m *mem.Mem_Flex_t, n int32, c []int32) string {
	return sopSingle(n, c, false, '0')
}
func Abc_SopCreateNor(m *mem.Mem_Flex_t, n int32) string { return sopSingle(n, nil, false, '1') }
func Abc_SopCreateOrMultiCube(m *mem.Mem_Flex_t, n int32, c []int32) string {
	row := []byte(Abc_SopStart(m, n, n))
	if c != nil && len(c) < int(n) {
		panic("abc: missing complement flags")
	}
	for i := int32(0); i < n; i++ {
		row[i*(n+3)+i] = '1'
		if c != nil {
			row[i*(n+3)+i] -= byte(c[i])
		}
	}
	return string(row)
}
func Abc_SopCreateXor(m *mem.Mem_Flex_t, n int32) string {
	if n != 2 {
		panic("abc: ordinary XOR SOP needs two variables")
	}
	return Abc_SopRegister(m, "01 1\n10 1\n")
}
func Abc_SopCreateXorSpecial(m *mem.Mem_Flex_t, n int32) string { return sopSingle(n, nil, true, 'x') }
func Abc_SopCreateNxor(m *mem.Mem_Flex_t, n int32) string {
	if n != 2 {
		panic("abc: ordinary NXOR SOP needs two variables")
	}
	return Abc_SopRegister(m, "11 1\n00 1\n")
}
func Abc_SopCreateMux(m *mem.Mem_Flex_t) string { return Abc_SopRegister(m, "11- 1\n0-1 1\n") }
func Abc_SopCreateInv(m *mem.Mem_Flex_t) string { return Abc_SopRegister(m, "0 1\n") }
func Abc_SopCreateBuf(m *mem.Mem_Flex_t) string { return Abc_SopRegister(m, "1 1\n") }
func Abc_SopGetCubeNum(s string) int32          { return int32(strings.Count(s, "\n")) }
func Abc_SopGetLitNum(s string) int32 {
	count := int32(0)
	for _, c := range []byte(s) {
		if c == '0' || c == '1' {
			count++
		} else if c == '\n' {
			count--
		}
	}
	return count
}
func Abc_SopGetVarNum(s string) int32 {
	i := strings.IndexByte(s, '\n')
	if i < 2 {
		return -1
	}
	return int32(i - 2)
}
func Abc_SopGetPhase(s string) int32 {
	n := Abc_SopGetVarNum(s)
	if n < 0 {
		panic("abc: malformed SOP")
	}
	switch s[n+1] {
	case '0', 'n':
		return 0
	case '1', 'x':
		return 1
	}
	panic("abc: invalid SOP phase")
}
func Abc_SopGetIthCareLit(s string, index int32) int32 {
	n := Abc_SopGetVarNum(s)
	if index < 0 || index >= n {
		panic("abc: variable outside SOP")
	}
	for i := 0; i < len(s); i += int(n + 3) {
		if s[i+int(index)] != '-' {
			return int32(s[i+int(index)] - '0')
		}
	}
	return -1
}

// The C mutators return a replacement string in Go; callers must save it.
func Abc_SopComplement(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c != '\n' {
			continue
		}
		switch b[i-1] {
		case '0':
			b[i-1] = '1'
		case '1':
			b[i-1] = '0'
		case 'x':
			b[i-1] = 'n'
		case 'n':
			b[i-1] = 'x'
		default:
			panic("abc: invalid SOP phase")
		}
	}
	return string(b)
}
func Abc_SopComplementVar(s string, index int32) string {
	n := Abc_SopGetVarNum(s)
	if index < 0 || index >= n {
		panic("abc: variable outside SOP")
	}
	b := []byte(s)
	for i := int(index); i < len(b); i += int(n + 3) {
		if b[i] == '0' {
			b[i] = '1'
		} else if b[i] == '1' {
			b[i] = '0'
		}
	}
	return string(b)
}
func Abc_SopIsComplement(s string) int32 { return 1 - Abc_SopGetPhase(s) }
func sopBool(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func Abc_SopIsConst0(s string) int32 { return sopBool(strings.HasPrefix(s, " 0")) }
func Abc_SopIsConst1(s string) int32 { return sopBool(strings.HasPrefix(s, " 1")) }
func Abc_SopIsBuf(s string) int32 {
	return sopBool(len(s) == 4 && (s[0] == '1' && s[2] == '1' || s[0] == '0' && s[2] == '0'))
}
func Abc_SopIsInv(s string) int32 {
	return sopBool(len(s) == 4 && (s[0] == '0' && s[2] == '1' || s[0] == '1' && s[2] == '0'))
}
func Abc_SopIsAndType(s string) int32 {
	n := Abc_SopGetVarNum(s)
	return sopBool(n >= 0 && Abc_SopGetCubeNum(s) == 1 && !strings.Contains(s[:max(n, 0)], "-") && s[n+1] == '1')
}
func Abc_SopIsOrType(s string) int32 {
	n := Abc_SopGetVarNum(s)
	if n < 0 || Abc_SopGetCubeNum(s) != n {
		return 0
	}
	for i := 0; i < len(s); i += int(n + 3) {
		if int(n)-strings.Count(s[i:i+int(n)], "-") != 1 {
			return 0
		}
	}
	return 1
}
func Abc_SopIsExorType(s string) int32 {
	n := Abc_SopGetVarNum(s)
	return sopBool(n >= 0 && (s[n+1] == 'x' || s[n+1] == 'n'))
}
func Abc_SopCheck(s string, n int32) int32 {
	if n < 0 {
		return 0
	}
	zero, one := false, false
	for offset := 0; offset < len(s); offset += int(n + 3) {
		if offset+int(n+3) > len(s) {
			return 0
		}
		row := s[offset : offset+int(n+3)]
		if row[n] != ' ' || row[n+2] != '\n' {
			return 0
		}
		for _, c := range row[:n] {
			if c != '0' && c != '1' && c != '-' {
				return 0
			}
		}
		switch row[n+1] {
		case '0':
			zero = true
		case '1':
			one = true
		case 'x', 'n':
		default:
			return 0
		}
	}
	return sopBool(!(zero && one))
}
func Abc_SopCreateFromTruth(m *mem.Mem_Flex_t, n int32, truth []uint32) string {
	if n < 0 || n > 30 || len(truth) < max(1, (1<<uint(n)+31)/32) {
		panic("abc: invalid truth table dimensions")
	}
	if n == 0 {
		if truth[0] != 0 {
			return Abc_SopCreateConst1(m)
		}
		return Abc_SopCreateConst0(m)
	}
	var b strings.Builder
	for mint := 0; mint < 1<<uint(n); mint++ {
		if truth[mint/32]&(1<<uint(mint%32)) == 0 {
			continue
		}
		for v := int32(0); v < n; v++ {
			b.WriteByte('0' + byte(mint>>uint(v)&1))
		}
		b.WriteString(" 1\n")
	}
	return b.String()
}
func Abc_SopCreateFromIsop(m *mem.Mem_Flex_t, n int32, cover *vec.Vec_Int_t) string {
	if n < 0 || n > 16 {
		panic("abc: invalid ISOP variable count")
	}
	b := []byte(Abc_SopStart(m, int32(vec.Vec_IntSize(cover)), n))
	for i, cube := range vec.Vec_IntArray(cover) {
		for v := int32(0); v < n; v++ {
			lit := uint32(cube) >> uint(2*v) & 3
			switch lit {
			case 0:
			case 1:
				b[i*int(n+3)+int(v)] = '0'
			case 2:
				b[i*int(n+3)+int(v)] = '1'
			default:
				panic("abc: contradictory ISOP literal")
			}
		}
	}
	return string(b)
}
func Abc_SopCreateFromTruthIsop(m *mem.Mem_Flex_t, n int32, truth []uint64, cover *vec.Vec_Int_t) string {
	if n < 0 || n >= 16 {
		panic("abc: ISOP supports fewer than 16 variables")
	}
	words := 1
	if n > 6 {
		words = 1 << uint(n-6)
	}
	if len(truth) < words {
		panic("abc: short truth table")
	}
	zero, one := true, true
	for _, w := range truth[:words] {
		zero = zero && w == 0
		one = one && w == ^uint64(0)
	}
	if zero {
		return Abc_SopCreateConst0(m)
	}
	if one {
		return Abc_SopCreateConst1(m)
	}
	tt := make([]uint32, 2*words)
	for i, w := range truth[:words] {
		tt[2*i], tt[2*i+1] = uint32(w), uint32(w>>32)
	}
	var cubes []uint32
	phase := kit.Kit_TruthIsop(tt, n, &cubes, 1)
	if phase < 0 {
		panic("abc: ISOP resource limit")
	}
	vec.Vec_IntClear(cover)
	for _, c := range cubes {
		vec.Vec_IntPush(cover, int32(c))
	}
	s := Abc_SopCreateFromIsop(m, n, cover)
	if phase != 0 {
		s = Abc_SopComplement(s)
	}
	return s
}
func Abc_SopToIsop(s string, cover *vec.Vec_Int_t) {
	n := Abc_SopGetVarNum(s)
	if n < 1 || n > 16 || Abc_SopCheck(s, n) == 0 {
		panic("abc: invalid SOP for ISOP conversion")
	}
	vec.Vec_IntClear(cover)
	for i := 0; i < len(s); i += int(n + 3) {
		var cube uint32
		for v := int32(0); v < n; v++ {
			if s[i+int(v)] == '0' {
				cube |= 1 << uint(2*v)
			} else if s[i+int(v)] == '1' {
				cube |= 2 << uint(2*v)
			}
		}
		vec.Vec_IntPush(cover, int32(cube))
	}
}

// Binary text follows the source's MSB-first variable order; hexadecimal text
// follows its LSB-first variable order. Their ordering difference is intentional.
func Abc_SopFromTruthBin(text string) string {
	n := len(text)
	if n == 0 || n&(n-1) != 0 {
		return ""
	}
	variables := bits.Len(uint(n)) - 1
	var b strings.Builder
	ones := 0
	for i, c := range []byte(text) {
		if c != '0' && c != '1' {
			return ""
		}
		if c == '0' {
			continue
		}
		ones++
		mint := n - 1 - i
		for v := variables - 1; v >= 0; v-- {
			b.WriteByte('0' + byte(mint>>uint(v)&1))
		}
		b.WriteString(" 1\n")
	}
	if ones == 0 || ones == n {
		return ""
	}
	return b.String()
}
func Abc_SopFromTruthHex(text string) string {
	n := len(text)
	if n == 0 || n&(n-1) != 0 {
		return ""
	}
	variables := bits.Len(uint(n)) - 1 + 2
	var b strings.Builder
	for i, c := range []byte(text) {
		value, e := strconv.ParseUint(string(c), 16, 4)
		if e != nil {
			return ""
		}
		for bit := 0; bit < 4; bit++ {
			if value&(1<<uint(bit)) == 0 {
				continue
			}
			mint := 4*(n-1-i) + bit
			for v := 0; v < variables; v++ {
				b.WriteByte('0' + byte(mint>>uint(v)&1))
			}
			b.WriteString(" 1\n")
		}
	}
	return b.String()
}
func Abc_SopEncoderPos(m *mem.Mem_Flex_t, value, values int32) string {
	if value < 0 || value >= values {
		panic("abc: invalid encoder value")
	}
	return Abc_SopRegister(m, fmt.Sprintf("d0\n%d 1\n", value))
}
func Abc_SopEncoderLog(m *mem.Mem_Flex_t, bit, values int32) string {
	if values < 2 || bit < 0 || int(bit) >= bits.Len32(uint32(values-1)) {
		panic("abc: invalid encoder bit")
	}
	var numbers []string
	for v := int32(0); v < values; v++ {
		if v&(1<<uint(bit)) != 0 {
			numbers = append(numbers, strconv.Itoa(int(v)))
		}
	}
	text := strings.Join(numbers, ",")
	if len(numbers) > 1 {
		text = "(" + text + ")"
	}
	return Abc_SopRegister(m, "d0\n"+text+" 1\n")
}
func Abc_SopDecoderPos(m *mem.Mem_Flex_t, values int32) string {
	if values < 2 {
		panic("abc: invalid decoder size")
	}
	var b strings.Builder
	for i := int32(0); i < values; i++ {
		for v := int32(0); v < values; v++ {
			if v == i {
				b.WriteString("1 ")
			} else {
				b.WriteString("- ")
			}
		}
		fmt.Fprintf(&b, "%d\n", i)
	}
	return Abc_SopRegister(m, b.String())
}
func Abc_SopDecoderLog(m *mem.Mem_Flex_t, values int32) string {
	if values < 2 {
		panic("abc: invalid decoder size")
	}
	n := bits.Len32(uint32(values - 1))
	var b strings.Builder
	for i := int32(0); i < values; i++ {
		for v := 0; v < n; v++ {
			b.WriteByte('0' + byte(i>>uint(v)&1))
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%d\n", i)
	}
	return Abc_SopRegister(m, b.String())
}
func sopTruth(s string, n int32, variables [][]uint64, cube, result []uint64) {
	if n < 0 || n > 16 || Abc_SopGetVarNum(s) != n || Abc_SopCheck(s, n) == 0 || s == "" {
		panic("abc: invalid SOP truth conversion")
	}
	words := 1
	if n > 6 {
		words = 1 << uint(n-6)
	}
	if len(result) < words || len(cube) < words || len(variables) < int(n) {
		panic("abc: short truth storage")
	}
	for _, v := range variables[:n] {
		if len(v) < words {
			panic("abc: short variable truth table")
		}
	}
	clear(result[:words])
	for offset := 0; offset < len(s); offset += int(n + 3) {
		for w := 0; w < words; w++ {
			cube[w] = ^uint64(0)
		}
		for v := int32(0); v < n; v++ {
			switch s[offset+int(v)] {
			case '0':
				for w := 0; w < words; w++ {
					cube[w] &= ^variables[v][w]
				}
			case '1':
				for w := 0; w < words; w++ {
					cube[w] &= variables[v][w]
				}
			}
		}
		for w := 0; w < words; w++ {
			result[w] |= cube[w]
		}
	}
	if Abc_SopIsComplement(s) != 0 {
		for w := 0; w < words; w++ {
			result[w] = ^result[w]
		}
	}
}
func sopVariables(n int32, words int) [][]uint64 {
	variables := make([][]uint64, n)
	for v := int32(0); v < n; v++ {
		variables[v] = make([]uint64, words)
		for bit := 0; bit < words*64; bit++ {
			if bit&(1<<uint(v)) != 0 {
				variables[v][bit/64] |= 1 << uint(bit%64)
			}
		}
	}
	return variables
}
func Abc_SopToTruth(s string, n int32) uint64 {
	if n > 6 {
		panic("abc: one-word truth supports at most six variables")
	}
	var cube, result [1]uint64
	sopTruth(s, n, sopVariables(n, 1), cube[:], result[:])
	return result[0]
}
func Abc_SopToTruth7(s string, n int32, result []uint64) {
	if n < 0 || n > 7 || len(result) < 2 {
		panic("abc: two-word truth supports at most seven variables")
	}
	if n <= 6 {
		result[0] = Abc_SopToTruth(s, n)
		result[1] = result[0]
		return
	}
	var cube [2]uint64
	sopTruth(s, n, sopVariables(n, 2), cube[:], result)
}
func Abc_SopToTruthBig(s string, n int32, variables [][]uint64, cube, result []uint64) {
	sopTruth(s, n, variables, cube, result)
}
