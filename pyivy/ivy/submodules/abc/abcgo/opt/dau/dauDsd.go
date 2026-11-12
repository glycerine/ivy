// Truth-table disjoint-support decomposition and its expression language.
package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"io"
	"math/bits"
	"math/rand/v2"
	"sort"
	"strings"
)

type dsExpr struct {
	kind     byte
	neg      bool
	variable int
	hex      string
	children []*dsExpr
	function *dsExpr
}

func (e *dsExpr) text() string {
	var b strings.Builder
	if e.neg {
		b.WriteByte('!')
	}
	switch e.kind {
	case 'v':
		b.WriteByte(byte('a' + e.variable))
	case '0', '1':
		b.WriteByte(e.kind)
	case '@':
		b.WriteString(e.function.text())
		b.WriteByte('{')
		for _, c := range e.children {
			b.WriteString(c.text())
		}
		b.WriteByte('}')
	case '{':
		b.WriteString(e.hex)
		b.WriteByte('{')
		for _, c := range e.children {
			b.WriteString(c.text())
		}
		b.WriteByte('}')
	default:
		b.WriteByte(e.kind)
		for _, c := range e.children {
			b.WriteString(c.text())
		}
		b.WriteByte(map[byte]byte{'(': ')', '[': ']', '<': '>'}[e.kind])
	}
	return b.String()
}
func dsParse(text string) *dsExpr {
	position := 0
	var parse func() *dsExpr
	parse = func() *dsExpr {
		e := &dsExpr{}
		for position < len(text) && text[position] == '!' {
			e.neg = !e.neg
			position++
		}
		if position == len(text) {
			panic("dau: missing DSD operand")
		}
		c := text[position]
		if c >= 'a' && c <= 'z' {
			e.kind = 'v'
			e.variable = int(c - 'a')
			position++
			return e
		}
		if (c == '0' || c == '1') && position+1 == len(text) {
			e.kind = c
			position++
			return e
		}
		start := position
		for position < len(text) && ((text[position] >= '0' && text[position] <= '9') || (text[position] >= 'A' && text[position] <= 'F')) {
			position++
		}
		e.hex = text[start:position]
		if position == len(text) {
			panic("dau: prime truth has no argument list")
		}
		e.kind = text[position]
		close, ok := map[byte]byte{'(': ')', '[': ']', '<': '>', '{': '}'}[e.kind]
		if !ok {
			panic("dau: invalid DSD operator")
		}
		position++
		for position < len(text) && text[position] != close {
			e.children = append(e.children, parse())
		}
		if position == len(text) {
			panic("dau: unterminated DSD expression")
		}
		position++
		if e.kind == '<' && len(e.children) != 3 {
			panic("dau: MUX needs three arguments")
		}
		if e.kind == '<' && position < len(text) && text[position] == '{' {
			neg := e.neg
			e.neg = false
			composed := &dsExpr{kind: '@', neg: neg, function: e}
			position++
			for position < len(text) && text[position] != '}' {
				composed.children = append(composed.children, parse())
			}
			if position == len(text) {
				panic("dau: unterminated composed MUX")
			}
			position++
			return composed
		}
		return e
	}
	result := parse()
	if position != len(text) {
		panic("dau: trailing DSD text")
	}
	return result
}
func dsCheckVars(n int32) {
	if n < 0 || n > 16 {
		panic("dau: truth table supports zero through sixteen variables")
	}
}
func dsWords(n int32) int {
	dsCheckVars(n)
	if n <= 6 {
		return 1
	}
	return 1 << uint(n-6)
}
func dsBit(t []uint64, index int) bool { return t[index>>6]>>uint(index&63)&1 != 0 }
func dsSet(t []uint64, index int)      { t[index>>6] |= uint64(1) << uint(index&63) }
func dsElem(n int32, v int) []uint64 {
	result := make([]uint64, dsWords(n))
	for i := 0; i < 64*len(result); i++ {
		if i&(1<<uint(v)) != 0 {
			dsSet(result, i)
		}
	}
	return result
}
func dsEval(e *dsExpr, assignment int, arguments []bool) bool {
	value := false
	switch e.kind {
	case '0':
	case '1':
		value = true
	case 'v':
		if arguments != nil {
			if e.variable >= len(arguments) {
				panic("dau: prime variable exceeds fanin count")
			}
			value = arguments[e.variable]
		} else {
			value = assignment&(1<<uint(e.variable)) != 0
		}
	case '(':
		value = true
		for _, c := range e.children {
			value = value && dsEval(c, assignment, arguments)
		}
	case '[':
		for _, c := range e.children {
			value = value != dsEval(c, assignment, arguments)
		}
	case '<':
		if dsEval(e.children[0], assignment, arguments) {
			value = dsEval(e.children[1], assignment, arguments)
		} else {
			value = dsEval(e.children[2], assignment, arguments)
		}
	case '@':
		inputs := make([]bool, len(e.children))
		for i, c := range e.children {
			inputs[i] = dsEval(c, assignment, arguments)
		}
		value = dsEval(e.function, 0, inputs)
	case '{':
		if e.hex == "" {
			panic("dau: prime expression needs inline hexadecimal truth")
		}
		index := 0
		for i, c := range e.children {
			if dsEval(c, assignment, arguments) {
				index |= 1 << uint(i)
			}
		}
		digit := len(e.hex) - 1 - index/4
		if digit < 0 {
			panic("dau: prime truth shorter than fanin domain")
		}
		character := e.hex[digit]
		nibble := int(character - '0')
		if character >= 'A' {
			nibble = int(character-'A') + 10
		}
		value = nibble&(1<<uint(index%4)) != 0
	}
	return value != e.neg
}
func Dau_DsdComputeMatches(text string) []int32 {
	matches := make([]int32, len(text))
	stack := []int{}
	for i, c := range []byte(text) {
		switch c {
		case '(', '[', '<', '{':
			stack = append(stack, i)
		case ')', ']', '>', '}':
			if len(stack) == 0 {
				panic("dau: unmatched DSD closing delimiter")
			}
			opening := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if map[byte]byte{'(': ')', '[': ']', '<': '>', '{': '}'}[text[opening]] != c {
				panic("dau: mismatched DSD delimiter")
			}
			matches[opening] = int32(i)
		}
	}
	if len(stack) != 0 {
		panic("dau: unmatched DSD opening delimiter")
	}
	return matches
}
func Dau_DsdFindVarNum(text string) int32 {
	maximum := byte('a')
	for _, c := range []byte(text) {
		if c >= 'a' && c <= 'z' && c > maximum {
			maximum = c
		}
	}
	return int32(maximum - 'a' + 1)
}
func Dau_DsdGenRandPerm(permutation []int32, n int32) {
	if n < 0 || int(n) > len(permutation) {
		panic("dau: permutation size")
	}
	for i := int32(0); i < n; i++ {
		permutation[i] = i
	}
	for i := int32(0); i < n; i++ {
		j := rand.Int32N(n)
		permutation[i], permutation[j] = permutation[j], permutation[i]
	}
}
func Dau_DsdPermute(text string) string {
	n := Dau_DsdFindVarNum(text)
	permutation := make([]int32, n)
	Dau_DsdGenRandPerm(permutation, n)
	result := []byte(text)
	for i, c := range result {
		if c >= 'a' && c < byte('a'+n) {
			result[i] = byte('a' + permutation[c-'a'])
		}
	}
	return string(result)
}
func dsLetters(e *dsExpr) string {
	var b strings.Builder
	for _, c := range e.text() {
		if c >= 'a' && c <= 'z' {
			b.WriteRune(c)
		}
	}
	return b.String()
}
func Dau_DsdNormalize(text string) string {
	if text == "0" || text == "1" {
		return text
	}
	root := dsParse(text)
	var normalize func(*dsExpr)
	normalize = func(e *dsExpr) {
		for _, c := range e.children {
			normalize(c)
		}

		if e.kind == '(' || e.kind == '[' {
			sort.SliceStable(e.children, func(i, j int) bool { return dsLetters(e.children[i]) < dsLetters(e.children[j]) })
		}
	}
	normalize(root)
	return root.text()
}
func Dau_DsdCountAnds(text string) int32 {
	var count func(*dsExpr) int32
	count = func(e *dsExpr) int32 {
		total := int32(0)
		for _, c := range e.children {
			total += count(c)
		}
		switch e.kind {
		case '(':
			total += int32(len(e.children) - 1)
		case '[':
			total += 3 * int32(len(e.children)-1)
		case '<', '{', '@':
			total += 3
		}
		return total
	}
	return count(dsParse(text))
}
func Dau_DsdToTruth(text string, n int32) []uint64 {
	dsCheckVars(n)
	root := dsParse(text)
	var validate func(*dsExpr, int)
	validate = func(e *dsExpr, variables int) {
		if e.kind == 'v' && e.variable >= variables {
			panic("dau: variable exceeds truth-table domain")
		}
		for _, c := range e.children {
			validate(c, variables)
		}
		if e.function != nil {
			validate(e.function, len(e.children))
		}
		if e.kind == '{' {
			expected := 1 << uint(max(len(e.children)-2, 0))
			if len(e.hex) != expected {
				panic("dau: prime hexadecimal width does not match fanin count")
			}
		}
	}
	validate(root, int(n))
	result := make([]uint64, dsWords(n))
	// Small tables retain ABC's repeated 64-bit representation.
	for index := 0; index < 64*len(result); index++ {
		if dsEval(root, index, nil) {
			dsSet(result, index)
		}
	}
	return result
}
func Dau_Dsd6ToTruth(text string) uint64 { return Dau_DsdToTruth(text, 6)[0] }
func dsFlatten(e *dsExpr) *dsExpr {
	for i, c := range e.children {
		e.children[i] = dsFlatten(c)
	}

	if e.kind == '(' || e.kind == '[' {
		children := []*dsExpr{}
		for _, c := range e.children {
			if c.kind == e.kind && (!c.neg || e.kind == '[') {
				if c.neg && len(c.children) > 0 {
					c.children[0].neg = !c.children[0].neg
				}
				children = append(children, c.children...)
			} else {
				children = append(children, c)
			}
		}
		e.children = children
	}
	return e
}

type Dau_Dsd_t struct {
	nVarsInit, nVarsUsed, nConsts, nSizeNonDec int32
	uConstMask                                 uint32
	fSplitPrime, fWriteTruth                   int32
	pVarLevels                                 []int32
	definitions                                []string
	cache                                      map[[2]int32]int32
	output                                     strings.Builder
}

func dsInitialize(p *Dau_Dsd_t, n int32) {
	p.nVarsInit = n
	p.nVarsUsed = n
	p.nConsts = 0
	p.nSizeNonDec = 0
	p.uConstMask = 0
	p.output.Reset()
	p.cache = make(map[[2]int32]int32)
	p.definitions = make([]string, n)
	for i := range p.definitions {
		p.definitions[i] = string(byte('a' + i))
	}
}
func (p *Dau_Dsd_t) variable(id int32) string {
	var b strings.Builder
	for _, c := range []byte(p.definitions[id]) {
		if int32(c) >= 'a'+p.nVarsInit && int32(c) < 'a'+p.nVarsUsed {
			b.WriteString(p.variable(int32(c - 'a')))
		} else {
			b.WriteByte(c)
		}
	}
	return b.String()
}
func (p *Dau_Dsd_t) writeVar(id int32, invert bool) {
	if invert {
		p.output.WriteByte('!')
	}
	p.output.WriteString(p.variable(id))
}
func (p *Dau_Dsd_t) addDefinition(text string) int32 {
	id := p.nVarsUsed
	p.nVarsUsed++
	p.definitions = append(p.definitions, text)
	return id
}
func (p *Dau_Dsd_t) translate(text string, variables []int32, n int32) {
	for _, c := range []byte(text) {
		if c >= 'a' && int32(c-'a') < n {
			p.writeVar(variables[c-'a'], false)
		} else {
			p.output.WriteByte(c)
		}
	}
}
func Dau_DsdLevelVar(p *Dau_Dsd_t, id int32) int32 {
	maximum := int32(0)
	for _, c := range []byte(p.definitions[id]) {
		if c < 'a' || int32(c) >= 'a'+p.nVarsUsed {
			continue
		}
		level := int32(0)
		if int32(c) >= 'a'+p.nVarsInit {
			level = 1 + Dau_DsdLevelVar(p, int32(c-'a'))
		} else if p.pVarLevels != nil {
			level = p.pVarLevels[c-'a']
		}
		maximum = max(maximum, level)
	}
	return maximum
}
func dsCof(t []uint64, n, v int32, value bool) []uint64 {
	result := append([]uint64(nil), t[:dsWords(n)]...)
	if value {
		util.Abc_TtCofactor1(result, int32(len(result)), v)
	} else {
		util.Abc_TtCofactor0(result, int32(len(result)), v)
	}
	return result
}
func dsEqual(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func dsConstant(t []uint64, value bool) bool {
	expected := uint64(0)
	if value {
		expected = ^uint64(0)
	}
	for _, word := range t {
		if word != expected {
			return false
		}
	}
	return true
}
func dsComplement(t []uint64) {
	for i := range t {
		t[i] = ^t[i]
	}
}
func dsMux(t, a, b []uint64, n, v int32) {
	ctrl := dsElem(n, int(v))
	for i := range ctrl {
		t[i] = ctrl[i]&a[i] | ^ctrl[i]&b[i]
	}
}
func dsFind(variables []int32, n, id int32) int32 {
	for i := int32(0); i < n; i++ {
		if variables[i] == id {
			return i
		}
	}
	panic("dau: lost decomposition variable")
}
func dsSingleOne(p *Dau_Dsd_t, t []uint64, variables []int32, n, v int32) bool {
	c0, c1 := dsCof(t, n, v, false), dsCof(t, n, v, true)
	prefix := ""
	var rest []uint64
	switch {
	case dsConstant(c0, true):
		prefix = "!("
		rest = c1
		dsComplement(rest)
	case dsConstant(c0, false):
		prefix = "("
		rest = c1
	case dsConstant(c1, true):
		prefix = "!(!"
		rest = c0
		dsComplement(rest)
	case dsConstant(c1, false):
		prefix = "(!"
		rest = c0
	default:
		opposite := true
		for i := range c0 {
			opposite = opposite && c0[i] == ^c1[i]
		}
		if !opposite {
			return false
		}
		prefix = "["
		rest = c0
		p.uConstMask |= 1 << uint(p.nConsts)
	}
	p.output.WriteString(prefix)
	p.nConsts++
	p.writeVar(variables[v], false)
	copy(t, rest)
	variables[v] = variables[n-1]
	util.Abc_TtSwapVars(t, n, v, n-1)
	return true
}
func dsSingle(p *Dau_Dsd_t, t []uint64, variables []int32, n int32) int32 {
	for n > 1 {
		changed := false
		for v := n - 1; v >= 0; v-- {
			if dsSingleOne(p, t, variables, n, v) {
				n--
				changed = true
				break
			}
		}
		if !changed {
			break
		}
	}
	if n == 1 {
		p.writeVar(variables[0], t[0]&1 != 0)
		return 0
	}
	return n
}
func dsSupport(p *Dau_Dsd_t, t []uint64, variables []int32, n, v, u int32) int32 {
	key := [2]int32{variables[v], variables[u]}
	if status := p.cache[key]; status != 0 {
		return status
	}
	c0, c1 := dsCof(t, n, v, false), dsCof(t, n, v, true)
	status := util.Abc_TtHasVar(c0, n, u) | util.Abc_TtHasVar(c1, n, u)<<1
	if status == 0 {
		panic("dau: unsupported variable in reduced base")
	}
	p.cache[key] = status
	return status
}
func dsDoubleOne(p *Dau_Dsd_t, t []uint64, variables []int32, n, v, u int32) int32 {
	status := dsSupport(p, t, variables, n, v, u)
	c0, c1 := dsCof(t, n, v, false), dsCof(t, n, v, true)
	c00, c01 := dsCof(c0, n, u, false), dsCof(c0, n, u, true)
	c10, c11 := dsCof(c1, n, u, false), dsCof(c1, n, u, true)
	definition := ""
	var low, high []uint64
	a, b := byte('a'+variables[v]), byte('a'+variables[u])
	switch status {
	case 3:
		if dsEqual(c00, c11) && dsEqual(c01, c10) {
			definition = fmt.Sprintf("[%c%c]", a, b)
			low, high = c00, c01
		}
	case 2:
		if dsEqual(c00, c10) {
			definition = fmt.Sprintf("(%c%c)", a, b)
			low, high = c00, c11
		} else if dsEqual(c00, c11) {
			definition = fmt.Sprintf("(%c!%c)", a, b)
			low, high = c00, c10
		}
	case 1:
		if dsEqual(c00, c11) {
			definition = fmt.Sprintf("(!%c%c)", a, b)
			low, high = c00, c01
		} else if dsEqual(c01, c11) {
			definition = fmt.Sprintf("(!%c!%c)", a, b)
			low, high = c11, c00
		}
	}
	if definition == "" {
		return n
	}
	dsMux(t, high, low, n, u)
	variables[u] = p.addDefinition(definition)
	variables[v] = variables[n-1]
	util.Abc_TtSwapVars(t, n, v, n-1)
	n--
	if dsSingleOne(p, t, variables, n, u) {
		n = dsSingle(p, t, variables, n-1)
	}
	return n
}
func dsDouble(p *Dau_Dsd_t, t []uint64, variables []int32, n int32) int32 {
	for n > 0 {
		changed := false
		for v := n - 1; v > 0 && !changed; v-- {
			for u := v - 1; u >= 0; u-- {
				if p.cache[[2]int32{variables[v], variables[u]}] != 0 {
					continue
				}
				next := dsDoubleOne(p, t, variables, n, v, u)
				if next == 0 {
					return 0
				}
				if next < n {
					n = next
					changed = true
					break
				}
			}
		}
		if !changed {
			return n
		}
	}
	return n
}
func dsOuter(p *Dau_Dsd_t, t []uint64, variables []int32, n, v int32) int32 {
	variables[v], variables[n-1] = variables[n-1], variables[v]
	util.Abc_TtSwapVars(t, n, v, n-1)
	c0, c1 := dsCof(t, n, n-1, false), dsCof(t, n, n-1, true)
	p.output.WriteByte('<')
	p.writeVar(variables[n-1], false)
	child := &Dau_Dsd_t{fWriteTruth: p.fWriteTruth}
	dsDecomposeInternal(child, c1, n-1)
	p.translate(child.output.String(), variables, n-1)
	p.nSizeNonDec = child.nSizeNonDec
	if child.nSizeNonDec > 0 {
		copy(t, c1[:dsWords(child.nSizeNonDec)])
	}
	dsDecomposeInternal(child, c0, n-1)
	p.translate(child.output.String(), variables, n-1)
	p.output.WriteByte('>')
	p.nSizeNonDec = max(p.nSizeNonDec, child.nSizeNonDec)
	if child.nSizeNonDec > 0 {
		copy(t, c0[:dsWords(child.nSizeNonDec)])
	}
	return 0
}
func dsInner(p *Dau_Dsd_t, t []uint64, variables []int32, n, v int32, support uint32) int32 {
	i0 := int32(bits.TrailingZeros32(support&(^support>>1)&0x55555555) / 2)
	i1 := int32(bits.TrailingZeros32(^support&(support>>1)&0x55555555) / 2)
	c0, c1 := dsCof(t, n, v, false), dsCof(t, n, v, true)
	c00, c01 := dsCof(c0, n, i0, false), dsCof(c0, n, i0, true)
	c10, c11 := dsCof(c1, n, i1, false), dsCof(c1, n, i1, true)
	equal0, equal1 := dsEqual(c00, c10) && dsEqual(c01, c11), dsEqual(c00, c11) && dsEqual(c01, c10)
	if !equal0 && !equal1 {
		return n
	}
	id0 := variables[i0]
	dsMux(t, c11, c10, n, v)
	inversion := ""
	if equal1 {
		inversion = "!"
	}
	variables[v] = p.addDefinition(fmt.Sprintf("<%c%c%s%c>", byte('a'+variables[v]), byte('a'+variables[i1]), inversion, byte('a'+variables[i0])))
	variables[i1], variables[n-1] = variables[n-1], variables[i1]
	util.Abc_TtSwapVars(t, n, i1, n-1)
	n--
	i0 = dsFind(variables, n, id0)
	variables[i0], variables[n-1] = variables[n-1], variables[i0]
	util.Abc_TtSwapVars(t, n, i0, n-1)
	n--
	v = dsFind(variables, n, p.nVarsUsed-1)
	if dsSingleOne(p, t, variables, n, v) {
		n = dsSingle(p, t, variables, n-1)
	}
	return n
}
func dsTriple(p *Dau_Dsd_t, t []uint64, variables []int32, n int32) int32 {
	for {
		changed := false
		for v := n - 1; v >= 0; v-- {
			support := uint32(0)
			for u := int32(0); u < n; u++ {
				if u != v {
					support |= uint32(dsSupport(p, t, variables, n, v, u)) << uint(2*u)
				}
			}
			if support&(support>>1)&0x55555555 == 0 {
				return dsOuter(p, t, variables, n, v)
			}
			unique0, unique1 := support&(^support>>1)&0x55555555, ^support&(support>>1)&0x55555555
			if bits.OnesCount32(unique0) == 1 && bits.OnesCount32(unique1) == 1 {
				next := dsInner(p, t, variables, n, v, support)
				if next == n {
					continue
				}
				if next == 0 {
					return 0
				}
				n = dsDouble(p, t, variables, next)
				if n == 0 {
					return 0
				}
				changed = true
				break
			}
		}
		if !changed {
			return n
		}
	}
}
func dsHex(t []uint64, n int32) string {
	digits := 1 << uint(max(n-2, 0))
	result := make([]byte, digits)
	for i := range result {
		nibble := (t[i/16] >> uint(4*(i%16))) & 15
		result[digits-1-i] = "0123456789ABCDEF"[nibble]
	}
	return string(result)
}
func dsPrime(p *Dau_Dsd_t, t []uint64, variables []int32, n int32) int32 {
	status := int32(2)
	if p.fSplitPrime != 0 {
		variable := Dau_DsdCheck1Step(p, t, n, p.pVarLevels)
		if variable == -2 {
			p.output.WriteString(dsHex(t, n))
		} else if variable >= 0 {
			p.output.WriteByte('<')
			p.writeVar(variable, false)
			for _, value := range []bool{true, false} {
				cofactor := dsCof(t, n, variable, value)
				var result []byte
				Dau_DsdDecompose(cofactor, n, 0, p.fWriteTruth, &result)
				p.output.Write(result)
			}
			p.output.WriteByte('>')
			status = 1
		} else {
			panic("dau: prime block unexpectedly decomposes")
		}
	} else if p.fWriteTruth != 0 {
		p.output.WriteString(dsHex(t, n))
	}
	p.output.WriteByte('{')
	for v := int32(0); v < n; v++ {
		p.writeVar(variables[v], false)
	}
	p.output.WriteByte('}')
	p.nSizeNonDec = n
	return status
}
func Dau_DsdMinBase(t []uint64, n int32, variables []int32) int32 {
	for v := int32(0); v < n; v++ {
		variables[v] = v
	}
	for v := n - 1; v >= 0; v-- {
		if util.Abc_TtHasVar(t, n, v) != 0 {
			continue
		}
		util.Abc_TtSwapVars(t, n, v, n-1)
		n--
		variables[v] = variables[n]
	}
	return n
}
func dsDecomposeInternal(p *Dau_Dsd_t, t []uint64, initial int32) int32 {
	dsInitialize(p, initial)
	variables := make([]int32, initial)
	n := Dau_DsdMinBase(t, initial, variables)
	if n == 0 {
		if t[0]&1 != 0 {
			p.output.WriteByte('1')
		} else {
			p.output.WriteByte('0')
		}
		return 0
	}
	status := int32(0)
	if n == 1 {
		p.writeVar(variables[0], t[0]&1 != 0)
	} else {
		n = dsSingle(p, t, variables, n)
		if n > 0 {
			n = dsDouble(p, t, variables, n)
		}
		if n > 0 {
			n = dsTriple(p, t, variables, n)
		}
		if n > 0 {
			status = dsPrime(p, t, variables, n)
		}
	}
	for i := p.nConsts - 1; i >= 0; i-- {
		if p.uConstMask&(1<<uint(i)) != 0 {
			p.output.WriteByte(']')
		} else {
			p.output.WriteByte(')')
		}
	}
	return status
}
func Dau_DsdDecompose(truth []uint64, n, split, write int32, result *[]byte) int32 {
	return Dau_DsdDecomposeLevel(truth, n, split, write, result, nil)
}
func Dau_DsdDecomposeLevel(truth []uint64, n, split, write int32, result *[]byte, levels []int32) int32 {
	if len(levels) == 0 {
		levels = nil
	}
	dsCheckVars(n)
	if len(truth) < dsWords(n) {
		panic("dau: truth table too short")
	}
	if levels != nil && len(levels) < int(n) {
		panic("dau: level vector too short")
	}
	p := &Dau_Dsd_t{fSplitPrime: split, fWriteTruth: write, pVarLevels: levels}
	status := dsDecomposeInternal(p, truth, n)
	text := p.output.String()
	if len(text) > 1 {
		text = dsFlatten(dsParse(text)).text()
	}
	if result != nil {
		*result = append((*result)[:0], text...)
	}
	if split != 0 && status == 2 {
		return -1
	}
	return p.nSizeNonDec
}
func DecomposeDSD(truth []uint64, n int32) (string, int32) {
	copyTruth := append([]uint64(nil), truth...)
	var result []byte
	prime := Dau_DsdDecompose(copyTruth, n, 0, 1, &result)
	return string(result), prime
}
func Dau_DsdCheck1Step(manager any, truth []uint64, n int32, levels []int32) int32 {
	probe := append([]uint64(nil), truth...)
	if Dau_DsdDecompose(probe, n, 0, 0, nil) == 0 {
		return -1
	}
	order := make([]int32, n)
	for i := range order {
		order[i] = int32(i)
	}
	if levels != nil {
		p, ok := manager.(*Dau_Dsd_t)
		if !ok {
			panic("dau: level-aware split needs decomposition manager")
		}
		for i := 0; i < len(order)-1; i++ {
			best := i
			for j := i + 1; j < len(order); j++ {
				if Dau_DsdLevelVar(p, order[j]) > Dau_DsdLevelVar(p, order[best]) {
					best = j
				}
			}
			order[i], order[best] = order[best], order[i]
		}
	}
	best, cost := int32(-2), int32(1<<30)
	for _, v := range order {
		sum := int32(0)
		decomposable := true
		for _, value := range []bool{false, true} {
			cofactor := dsCof(truth, n, v, value)
			sum += util.Abc_TtSupportSize(cofactor, n)
			if Dau_DsdDecompose(cofactor, n, 0, 0, nil) != 0 {
				decomposable = false
			}
		}
		if decomposable && sum < cost {
			best, cost = v, sum
		}
	}
	return best
}
func Dau_DsdPerform(t uint64) string {
	var perform func(uint64, []int32) string
	perform = func(truth uint64, variables []int32) string {
		if truth == 0 {
			return "0"
		}
		if truth == ^uint64(0) {
			return "1"
		}
		active := []int32{}
		for _, v := range variables {
			if util.Abc_Tt6HasVar(truth, v) != 0 {
				active = append(active, v)
			}
		}
		if len(active) == 1 {
			prefix := ""
			if truth&1 != 0 {
				prefix = "!"
			}
			return prefix + string(byte('a'+active[0]))
		}
		c0, c1 := make([]uint64, len(active)), make([]uint64, len(active))
		for i, v := range active {
			c0[i], c1[i] = util.Abc_Tt6Cofactor0(truth, v), util.Abc_Tt6Cofactor1(truth, v)
			name := string(byte('a' + v))
			switch {
			case c0[i] == 0:
				return "(" + name + perform(c1[i], active) + ")"
			case c0[i] == ^uint64(0):
				return "!(" + name + perform(^c1[i], active) + ")"
			case c1[i] == 0:
				return "(!" + name + perform(c0[i], active) + ")"
			case c1[i] == ^uint64(0):
				return "!(!" + name + perform(^c0[i], active) + ")"
			case c0[i] == ^c1[i]:
				return "[" + name + perform(c0[i], active) + "]"
			}
		}
		for v := range active {
			for u := v + 1; u < len(active); u++ {
				a, b := byte('a'+active[v]), byte('a'+active[u])
				f0, f1 := util.Abc_Tt6Cofactor0(c0[v], active[u]), util.Abc_Tt6Cofactor1(c0[v], active[u])
				f2, f3 := util.Abc_Tt6Cofactor0(c1[v], active[u]), util.Abc_Tt6Cofactor1(c1[v], active[u])
				definition := ""
				low, high := uint64(0), uint64(0)
				switch {
				case f0 == f1 && f0 == f2:
					definition = fmt.Sprintf("(%c%c)", a, b)
					low, high = f0, f3
				case f0 == f1 && f0 == f3:
					definition = fmt.Sprintf("(%c!%c)", a, b)
					low, high = f0, f2
				case f0 == f2 && f0 == f3:
					definition = fmt.Sprintf("(!%c%c)", a, b)
					low, high = f0, f1
				case f1 == f2 && f1 == f3:
					definition = fmt.Sprintf("(!%c!%c)", a, b)
					low, high = f1, f0
				case f0 == f3 && f1 == f2:
					definition = fmt.Sprintf("[%c%c]", a, b)
					low, high = f0, f1
				}
				if definition != "" {
					mask := dsElem(6, int(active[u]))[0]
					rest := perform(mask&high|^mask&low, active)
					return strings.ReplaceAll(rest, string(b), definition)
				}
			}
		}
		best, cost := 0, 10
		for v := range active {
			count := 0
			for u := range active {
				if u != v && util.Abc_Tt6HasVar(c0[v], active[u]) != 0 && util.Abc_Tt6HasVar(c1[v], active[u]) != 0 {
					count++
				}
			}
			if count < cost {
				best, cost = v, count
			}
			if count == 0 {
				break
			}
		}
		return "<" + string(byte('a'+active[best])) + perform(c1[best], active) + perform(c0[best], active) + ">"
	}
	return dsFlatten(dsParse(perform(t, []int32{0, 1, 2, 3, 4, 5}))).text()
}
func Dau_DsdPrintFromTruthFile(out io.Writer, truth []uint64, n int32) {
	text, _ := DecomposeDSD(truth, n)
	fmt.Fprintln(out, text)
}

func Dau_DsdTruthCompose_rec(function []uint64, fanins [][]uint64, result []uint64, nVars, nWords int32) {
	dsCheckVars(nVars)
	if len(function) < dsWords(nVars) || nWords < 0 || int(nWords) > len(result) || int(nVars) > len(fanins) {
		panic("dau: invalid truth composition storage")
	}
	for _, fanin := range fanins[:nVars] {
		if int(nWords) > len(fanin) {
			panic("dau: short composition fanin")
		}
	}
	composed := make([]uint64, nWords)
	for bit := 0; bit < 64*int(nWords); bit++ {
		index := 0
		for v := int32(0); v < nVars; v++ {
			if dsBit(fanins[v], bit) {
				index |= 1 << uint(v)
			}
		}
		if dsBit(function, index) {
			dsSet(composed, bit)
		}
	}
	copy(result, composed)
}
func Dau_Dsd6TruthCompose_rec(function uint64, fanins []uint64, nVars int32) uint64 {
	if nVars < 0 || nVars > 6 || int(nVars) > len(fanins) {
		panic("dau: invalid six-variable composition")
	}
	arguments := make([][]uint64, nVars)
	for i := range arguments {
		arguments[i] = []uint64{fanins[i]}
	}
	result := []uint64{0}
	Dau_DsdTruthCompose_rec([]uint64{function}, arguments, result, nVars, 1)
	return result[0]
}
func Dau_DsdPrintFromTruth(truth []uint64, n int32) {
	text, _ := DecomposeDSD(truth, n)
	fmt.Println(text)
}
func Dau_DsdPrintFromTruth2(truth []uint64, n int32) {
	text, _ := DecomposeDSD(truth, n)
	fmt.Print(text)
}
