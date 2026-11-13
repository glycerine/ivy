package acec

// Signature parsing and polynomial construction from acecPo.c.
import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_ParseSignatureMono(text string, level *vec.Vec_Int_t) {
	sign := int32(1)
	if len(text) > 0 && text[0] == '-' {
		sign = -1
	}
	text = strings.TrimLeft(text, "+- (")
	var coefficient int32
	found := false
	for _, field := range strings.Split(text, "*") {
		if field == "" {
			panic("invalid signature monomial")
		}
		switch field[0] {
		case 'i', 'o':
			n, err := strconv.ParseInt(field[1:], 10, 32)
			if err != nil || n < 0 {
				panic("invalid signature variable")
			}
			v := int32(n)
			if field[0] == 'i' {
				v = -1 - v
			}
			vec.Vec_IntPush(level, v)
		default:
			n, err := strconv.ParseInt(field, 10, 32)
			if err != nil || n < 0 || found {
				panic("invalid signature coefficient")
			}
			coefficient = int32(n) + 1
			found = true
		}
	}
	if !found {
		panic("missing signature coefficient")
	}
	vec.Vec_IntPush(level, sign*coefficient)
}
func Acec_ParseSignatureOne(text string) *vec.Vec_Wec_t {
	if len(text) < 2 || text[0] != '(' || text[len(text)-1] != ')' {
		return nil
	}
	result := vec.Vec_WecAlloc(10)
	body := text[1 : len(text)-1]
	for len(body) > 0 {
		end := 1
		for end < len(body) && body[end] != '+' && body[end] != '-' {
			end++
		}
		Acec_ParseSignatureMono(body[:end], vec.Vec_WecPushLevel(result))
		body = body[end:]
	}
	return result
}
func Acec_ParseDistribute(a, b, add *vec.Vec_Wec_t) *vec.Vec_Wec_t {
	out := vec.Vec_WecAlloc(10)
	for i := 0; i < vec.Vec_WecSize(a); i++ {
		x := vec.Vec_IntArray(vec.Vec_WecEntry(a, i))
		for j := 0; j < vec.Vec_WecSize(b); j++ {
			y := vec.Vec_IntArray(vec.Vec_WecEntry(b, j))
			row := vec.Vec_WecPushLevel(out)
			vec.Vec_IntPushArray(row, x, len(x)-1)
			vec.Vec_IntPushArray(row, y, len(y)-1)
			left, right := x[len(x)-1], y[len(y)-1]
			coefficient := abs32(left) + abs32(right) - 1
			if (left < 0) != (right < 0) {
				coefficient = -coefficient
			}
			vec.Vec_IntPush(row, coefficient)
		}
	}
	if add != nil {
		for i := 0; i < vec.Vec_WecSize(add); i++ {
			x := vec.Vec_IntArray(vec.Vec_WecEntry(add, i))
			vec.Vec_IntPushArray(vec.Vec_WecPushLevel(out), x, len(x))
		}
	}
	return out
}
func Acec_ParseSignature(text string) *vec.Vec_Wec_t {
	if !strings.HasPrefix(text, "(") {
		return Acec_ParseSignatureOne("(" + text + ")")
	}
	end := strings.IndexByte(text, ')')
	if end < 0 {
		return nil
	}
	first := Acec_ParseSignatureOne(text[:end+1])
	rest := text[end+1:]
	if rest == "" {
		return first
	}
	if !strings.HasPrefix(rest, "*(") {
		return nil
	}
	rest = rest[1:]
	end = strings.IndexByte(rest, ')')
	if end < 0 {
		return nil
	}
	second := Acec_ParseSignatureOne(rest[:end+1])
	rest = rest[end+1:]
	if rest == "" {
		return Acec_ParseDistribute(first, second, nil)
	}
	if !strings.HasPrefix(rest, "+(") {
		return nil
	}
	rest = rest[1:]
	end = strings.IndexByte(rest, ')')
	if end < 0 {
		return nil
	}
	return Acec_ParseDistribute(first, second, Acec_ParseSignatureOne(rest[:end+1]))
}
func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
func Acec_PrintSignature(monos *vec.Vec_Wec_t) {
	fmt.Printf("Output signature with %d monomials:\n", vec.Vec_WecSize(monos))
	for i := 0; i < vec.Vec_WecSize(monos); i++ {
		row := vec.Vec_IntArray(vec.Vec_WecEntry(monos, i))
		c := row[len(row)-1]
		sign := "+"
		if c < 0 {
			sign = "-"
		}
		fmt.Printf("  %s2^%d", sign, abs32(c)-1)
		for _, v := range row[:len(row)-1] {
			if v < 0 {
				fmt.Printf(" * i%d", -v-1)
			} else {
				fmt.Printf(" * o%d", v)
			}
		}
		fmt.Println()
	}
}
func Acec_ParseSignatureTest() {
	Acec_PrintSignature(Acec_ParseSignature("(4*o1+2*o2+1*o3)*(4*i4+2*i5+1*i6)+(4*o4+2*o5+1*o6)"))
}

type polynomialTerm struct{ vars, coef []int32 }
type polynomial struct {
	terms []*polynomialTerm
	byKey map[string]int
}

func newPolynomial() *polynomial {
	p := &polynomial{byKey: make(map[string]int)}
	p.add(nil, nil)
	return p
}
func insertUnique(values []int32, v int32) []int32 {
	i := sort.Search(len(values), func(i int) bool { return values[i] >= v })
	if i < len(values) && values[i] == v {
		return values
	}
	values = append(values, 0)
	copy(values[i+1:], values[i:])
	values[i] = v
	return values
}
func mergeCoefficient(values []int32, n int32) []int32 {
	if n == 0 {
		panic("zero encoded coefficient")
	}
	for i, old := range values {
		switch {
		case old == n:
			values = append(values[:i], values[i+1:]...)
			if n > 0 {
				return mergeCoefficient(values, n+1)
			}
			return mergeCoefficient(values, n-1)
		case abs32(old) == abs32(n):
			return append(values[:i], values[i+1:]...)
		case old+n == 1 || old+n == -1:
			values = append(values[:i], values[i+1:]...)
			v := abs32(old)
			if abs32(n) < v {
				v = abs32(n)
			}
			if old+n < 0 {
				v = -v
			}
			return mergeCoefficient(values, v)
		}
	}
	values = append(values, n)
	sort.Slice(values, func(i, j int) bool { return abs32(values[i]) > abs32(values[j]) })
	return values
}
func (p *polynomial) add(vars, coef []int32) {
	key := fmt.Sprint(vars)
	i, ok := p.byKey[key]
	if !ok {
		i = len(p.terms)
		p.byKey[key] = i
		p.terms = append(p.terms, &polynomialTerm{vars: append([]int32(nil), vars...)})
	}
	t := p.terms[i]
	for _, v := range coef {
		t.coef = mergeCoefficient(t.coef, v)
	}
}
func (p *polynomial) replace(id int32, positive, negative int32, literalMode bool) {
	count := len(p.terms)
	for i := 0; i < count; i++ {
		term := p.terms[i]
		if len(term.coef) == 0 {
			continue
		}
		pos := sort.Search(len(term.vars), func(i int) bool { return term.vars[i] >= id })
		if pos == len(term.vars) || term.vars[pos] != id {
			continue
		}
		vars := append([]int32(nil), term.vars...)
		vars = append(vars[:pos], vars[pos+1:]...)
		coef := append([]int32(nil), term.coef...)
		term.coef = nil
		if literalMode {
			if positive < 0 {
				p.add(vars, coef)
				neg := negCoefficients(coef)
				p.add(insertUnique(append([]int32(nil), vars...), negative), neg)
			} else {
				next := insertUnique(vars, positive)
				next = insertUnique(next, negative)
				p.add(next, coef)
			}
			continue
		}

		first, second := positive>>1, negative>>1
		switch {
		case positive&1 == 0 && negative&1 == 0:
			p.add(insertUnique(insertUnique(vars, first), second), coef)
		case positive&1 != 0 && negative&1 == 0:
			p.add(insertUnique(append([]int32(nil), vars...), second), coef)
			p.add(insertUnique(insertUnique(vars, first), second), negCoefficients(coef))
		case positive&1 == 0 && negative&1 != 0:
			p.add(insertUnique(append([]int32(nil), vars...), first), coef)
			p.add(insertUnique(insertUnique(vars, first), second), negCoefficients(coef))
		default:
			p.add(vars, coef)
			p.add(insertUnique(append([]int32(nil), vars...), first), negCoefficients(coef))
			p.add(insertUnique(append([]int32(nil), vars...), second), negCoefficients(coef))
			p.add(insertUnique(insertUnique(vars, first), second), coef)
		}
	}
}

// replaceXor expands a Boolean XOR as a+b-2ab, accounting for input
// complements in object mode. Source acecPo.c left this case empty.
func (p *polynomial) replaceXor(id, left, right int32, literalMode bool) {
	count := len(p.terms)
	for i := 0; i < count; i++ {
		term := p.terms[i]
		if len(term.coef) == 0 {
			continue
		}
		at := sort.Search(len(term.vars), func(j int) bool { return term.vars[j] >= id })
		if at == len(term.vars) || term.vars[at] != id {
			continue
		}
		vars := append([]int32(nil), term.vars...)
		vars = append(vars[:at], vars[at+1:]...)
		coef := append([]int32(nil), term.coef...)
		term.coef = nil
		if !literalMode {
			if (left^right)&1 != 0 {
				p.add(vars, coef)
				coef = negCoefficients(coef)
			}
			left >>= 1
			right >>= 1
		}
		p.add(insertUnique(append([]int32(nil), vars...), left), coef)
		p.add(insertUnique(append([]int32(nil), vars...), right), coef)
		twice := make([]int32, len(coef))
		for j, c := range coef {
			if c > 0 {
				twice[j] = -(c + 1)
			} else {
				twice[j] = -c + 1
			}
		}
		both := insertUnique(append([]int32(nil), vars...), left)
		both = insertUnique(both, right)
		p.add(both, twice)
	}
}

func negCoefficients(in []int32) []int32 {
	out := make([]int32, len(in))
	for i, v := range in {
		out[i] = -v
	}
	return out
}
func (p *polynomial) result() *vec.Vec_Wec_t {
	var terms []*polynomialTerm
	for _, t := range p.terms {
		if len(t.coef) > 0 {
			terms = append(terms, t)
		}
	}
	sort.SliceStable(terms, func(i, j int) bool {
		a, b := terms[i], terms[j]
		if abs32(a.coef[0]) != abs32(b.coef[0]) {
			return abs32(a.coef[0]) < abs32(b.coef[0])
		}
		first := func(t *polynomialTerm) int32 {
			if len(t.vars) == 0 {
				return -1
			}
			return t.vars[0]
		}
		return first(a) < first(b)
	})
	out := vec.Vec_WecAlloc(2 * len(terms))
	for _, t := range terms {
		vec.Vec_IntPushArray(vec.Vec_WecPushLevel(out), t.coef, len(t.coef))
		vec.Vec_IntPushArray(vec.Vec_WecPushLevel(out), t.vars, len(t.vars))
	}
	return out
}
func Gia_PolynPrintMono(c, m *vec.Vec_Int_t, prev int32) {
	values := vec.Vec_IntArray(c)
	marker := " "
	if prev != abs32(values[0]) {
		marker = "|"
	}
	fmt.Print(marker, " ")
	for _, v := range values {
		sign := "+"
		if v < 0 {
			sign = "-"
		}
		fmt.Printf("%s2^%d", sign, abs32(v)-1)
	}
	for _, v := range vec.Vec_IntArray(m) {
		fmt.Printf(" * i%d", v-1)
	}
	fmt.Println()
}
func Gia_PolynPrint(p *vec.Vec_Wec_t) {
	fmt.Printf("Polynomial with %d monomials:\n", vec.Vec_WecSize(p)/2)
	prev := int32(-1)
	for i := 0; i < vec.Vec_WecSize(p); i += 2 {
		c := vec.Vec_WecEntry(p, i)
		Gia_PolynPrintMono(c, vec.Vec_WecEntry(p, i+1), prev)
		prev = abs32(vec.Vec_IntEntry(c, 0))
	}
}
func Gia_PolynPrintStats(p *vec.Vec_Wec_t) {
	counts := map[int32]int{}
	for i := 0; i < vec.Vec_WecSize(p); i += 2 {
		counts[vec.Vec_IntEntry(vec.Vec_WecEntry(p, i), 0)]++
	}
	fmt.Printf("Input signature with %d monomials:\n", vec.Vec_WecSize(p)/2)
	keys := make([]int, 0, len(counts))
	for k := range counts {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, k := range keys {
		sign := "+"
		if k < 0 {
			sign = "-"
		}
		fmt.Printf("  %s2^%d appears %d times\n", sign, abs32(int32(k))-1, counts[int32(k)])
	}
}

func Gia_PolynBuildNew2(graph *gia.Gia_Man_t, roots *vec.Vec_Int_t, extra int32, leaves, nodes *vec.Vec_Int_t, signed, verbose, veryVerbose int32) *vec.Vec_Wec_t {
	p := newPolynomial()
	rootLits := vec.Vec_IntArray(roots)
	for i, lit := range rootLits {
		rank := i
		if limit := len(rootLits) - int(extra); rank > limit {
			rank = limit
		}
		coefficient := int32(rank + 1)
		if signed != 0 && i == len(rootLits)-1-int(extra) {
			coefficient = -coefficient
		}
		p.add([]int32{lit}, []int32{coefficient})
	}
	values := vec.Vec_IntArray(nodes)
	for i := len(values) - 1; i >= 0; i-- {
		id := values[i]
		obj := gia.Gia_ManObj(graph, id)
		p.replace(2*id+1, -1, 2*id, true)
		if gia.Gia_ObjIsXor(obj) != 0 {
			p.replaceXor(2*id, gia.Gia_ObjFaninLit0(obj, id), gia.Gia_ObjFaninLit1(obj, id), true)
		} else {
			p.replace(2*id, gia.Gia_ObjFaninLit0(obj, id), gia.Gia_ObjFaninLit1(obj, id), true)
		}
	}
	for _, id := range vec.Vec_IntArray(leaves) {
		p.replace(2*id+1, -1, 2*id, true)
	}
	if verbose != 0 || veryVerbose != 0 {
		fmt.Printf("Polynomial construction: %d terms\n", len(p.terms))
	}
	return p.result()
}
func Gia_PolynBuildNew(graph *gia.Gia_Man_t, sign *vec.Vec_Wec_t, roots *vec.Vec_Int_t, extra int32, leaves, nodes *vec.Vec_Int_t, signed, verbose, veryVerbose int32) *vec.Vec_Wec_t {
	p := newPolynomial()
	rootLits := vec.Vec_IntArray(roots)
	if sign != nil {
		leafIDs := vec.Vec_IntArray(leaves)
		for i := 0; i < vec.Vec_WecSize(sign); i++ {
			row := vec.Vec_IntArray(vec.Vec_WecEntry(sign, i))
			vars := []int32{}
			output := int32(-1)
			coef := row[len(row)-1]
			for _, v := range row[:len(row)-1] {
				if v < 0 {
					index := -1 - v
					if int(index) >= len(leafIDs) {
						panic("signature input out of range")
					}
					vars = insertUnique(vars, leafIDs[index])
				} else {
					if output != -1 || int(v) >= len(rootLits) {
						panic("invalid signature output")
					}
					output = rootLits[v]
				}
			}
			if output == -1 {
				p.add(vars, []int32{coef})
			} else if output&1 == 0 {
				p.add(insertUnique(vars, output>>1), []int32{coef})
			} else {
				p.add(vars, []int32{coef})
				p.add(insertUnique(append([]int32(nil), vars...), output>>1), []int32{-coef})
			}
		}
	} else {
		for i, lit := range rootLits {
			rank := i
			if limit := len(rootLits) - int(extra); rank > limit {
				rank = limit
			}
			coef := int32(rank + 1)
			if signed != 0 && i >= len(rootLits)-int(extra)-1 {
				coef = -coef
			}
			if lit&1 != 0 {
				p.add(nil, []int32{coef})
				p.add([]int32{lit >> 1}, []int32{-coef})
			} else {
				p.add([]int32{lit >> 1}, []int32{coef})
			}
		}
	}
	values := vec.Vec_IntArray(nodes)
	for i := len(values) - 1; i >= 0; i-- {
		id := values[i]
		obj := gia.Gia_ManObj(graph, id)
		if gia.Gia_ObjIsXor(obj) != 0 {
			p.replaceXor(id, gia.Gia_ObjFaninLit0(obj, id), gia.Gia_ObjFaninLit1(obj, id), false)
		} else {
			p.replace(id, gia.Gia_ObjFaninLit0(obj, id), gia.Gia_ObjFaninLit1(obj, id), false)
		}
	}
	if verbose != 0 || veryVerbose != 0 {
		fmt.Printf("Polynomial construction: %d terms\n", len(p.terms))
	}
	return p.result()
}
func Gia_PolynBuild2Test(graph *gia.Gia_Man_t, signature string, extra, signed, verbose, veryVerbose int32) {
	var sign *vec.Vec_Wec_t
	if signature != "" {
		sign = Acec_ParseSignature(signature)
		if sign == nil {
			fmt.Println("Cannot parse the output signatures.")
			return
		}
		if verbose != 0 {
			Acec_PrintSignature(sign)
		}
	}
	roots, leaves, nodes := vec.Vec_IntAlloc(16), vec.Vec_IntAlloc(16), vec.Vec_IntAlloc(16)
	if extra == -1 {
		gia.Gia_ManLevelNum(graph)
		max, imax := int32(-1), int32(-1)
		for id := int32(0); id < gia.Gia_ManObjNum(graph); id++ {
			obj := gia.Gia_ManObj(graph, id)
			if gia.Gia_ObjIsCo(obj) != 0 {
				level := gia.Gia_ObjLevel(graph, obj)
				if level > max {
					max = level
					imax = int32(vec.Vec_IntSize(roots))
				}
				vec.Vec_IntPush(roots, gia.Gia_ObjFaninLit0(obj, id))
			}
		}
		extra = gia.Gia_ManCoNum(graph) - imax - 1
		vec.Vec_IntClear(roots)
	}
	for id := int32(0); id < gia.Gia_ManObjNum(graph); id++ {
		obj := gia.Gia_ManObj(graph, id)
		if gia.Gia_ObjIsCi(obj) != 0 {
			vec.Vec_IntPush(leaves, id)
		} else if gia.Gia_ObjIsAnd(obj) != 0 {
			vec.Vec_IntPush(nodes, id)
		} else if gia.Gia_ObjIsCo(obj) != 0 {
			vec.Vec_IntPush(roots, gia.Gia_ObjFaninLit0(obj, id))
		}
	}
	result := Gia_PolynBuildNew(graph, sign, roots, extra, leaves, nodes, signed, verbose, veryVerbose)
	if verbose != 0 || veryVerbose != 0 {
		Gia_PolynPrintStats(result)
	}
	if veryVerbose != 0 {
		Gia_PolynPrint(result)
	}
}
