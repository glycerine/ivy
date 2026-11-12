package extra

// Small complemented-edge ROBDD package. Literal zero is false, one is true,
// and regular nodes have even literals. Every stored else edge is regular.
import (
	"fmt"
	"io"
	"math/bits"
	"os"

	"github.com/glycerine/abcgo/misc/vec"
)

type abcBddNode struct {
	variable, then, otherwise int32
	mark                      bool
}
type abcBddKey struct{ variable, then, otherwise int32 }
type abcBddCache struct{ a, b, result int32 }
type Abc_BddMan struct {
	nVars, nObjs, nObjsAlloc    int32
	nodes                       []abcBddNode
	unique                      map[abcBddKey]int32
	cache                       map[uint32]abcBddCache
	nUniqueMask, nCacheMask     uint32
	nCacheLookups, nCacheMisses uint64
}

func Abc_BddIthVar(i int32) int32 {
	if i < 0 || i >= 254 {
		panic("invalid BDD input")
	}
	return (i + 1) << 1
}
func Abc_BddHash(a, b, c int32) uint32 {
	return 12582917*uint32(a) + 4256249*uint32(b) + 741457*uint32(c)
}
func bddNode(p *Abc_BddMan, lit int32) *abcBddNode {
	if lit < 0 || lit>>1 >= p.nObjs {
		panic("invalid BDD literal")
	}
	return &p.nodes[lit>>1]
}
func Abc_BddVar(p *Abc_BddMan, lit int32) int32  { return bddNode(p, lit).variable }
func Abc_BddThen(p *Abc_BddMan, lit int32) int32 { return bddNode(p, lit).then ^ (lit & 1) }
func Abc_BddElse(p *Abc_BddMan, lit int32) int32 { return bddNode(p, lit).otherwise ^ (lit & 1) }
func Abc_BddMark(p *Abc_BddMan, lit int32) int32 {
	if bddNode(p, lit).mark {
		return 1
	}
	return 0
}
func Abc_BddSetMark(p *Abc_BddMan, lit, mark int32) { bddNode(p, lit).mark = mark != 0 }
func Abc_BddUniqueCreateInt(p *Abc_BddMan, v, a, b int32) int32 {
	if b&1 != 0 {
		panic("complemented stored BDD else edge")
	}
	key := abcBddKey{v, a, b}
	if lit, ok := p.unique[key]; ok {
		return lit
	}
	if p.nObjs == p.nObjsAlloc {
		panic("BDD node limit exceeded")
	}
	lit := p.nObjs << 1
	p.unique[key] = lit
	p.nodes = append(p.nodes, abcBddNode{variable: v, then: a, otherwise: b})
	p.nObjs++
	return lit
}
func Abc_BddUniqueCreate(p *Abc_BddMan, v, a, b int32) int32 {
	if v < 0 || v >= p.nVars || v >= Abc_BddVar(p, a) || v >= Abc_BddVar(p, b) {
		panic("unordered BDD node")
	}
	if a == b {
		return b
	}
	if b&1 == 0 {
		return Abc_BddUniqueCreateInt(p, v, a, b)
	}
	return Abc_BddUniqueCreateInt(p, v, a^1, b^1) ^ 1
}
func Abc_BddCacheLookup(p *Abc_BddMan, a, b int32) int32 {
	p.nCacheLookups++
	entry, ok := p.cache[Abc_BddHash(0, a, b)&p.nCacheMask]
	if ok && entry.a == a && entry.b == b {
		return entry.result
	}
	return -1
}
func Abc_BddCacheInsert(p *Abc_BddMan, a, b, result int32) int32 {
	if result < 0 {
		panic("negative BDD cache result")
	}
	p.cache[Abc_BddHash(0, a, b)&p.nCacheMask] = abcBddCache{a, b, result}
	p.nCacheMisses++
	return result
}
func Abc_BddManAlloc(vars, limit int32) *Abc_BddMan {
	if vars < 0 || vars >= 255 || limit <= vars || limit > 1<<30 {
		panic("invalid BDD manager dimensions")
	}
	mask := uint32(1)<<uint(bits.Len32(uint32(limit-1))) - 1
	p := &Abc_BddMan{nVars: vars, nObjs: 1, nObjsAlloc: limit, nUniqueMask: mask, nCacheMask: mask, nodes: []abcBddNode{{variable: 255}}, unique: map[abcBddKey]int32{}, cache: map[uint32]abcBddCache{}}
	for i := int32(0); i < vars; i++ {
		Abc_BddUniqueCreate(p, i, 1, 0)
	}
	return p
}
func Abc_BddManFree(p *Abc_BddMan) {
	if p == nil || p.nodes == nil {
		return
	}
	fmt.Printf("BDD stats: Var = %d  Obj = %d  Alloc = %d  Hit = %d  Miss = %d\n", p.nVars, p.nObjs, p.nObjsAlloc, p.nCacheLookups-p.nCacheMisses, p.nCacheMisses)
	*p = Abc_BddMan{}
}
func Abc_BddAnd(p *Abc_BddMan, a, b int32) int32 {
	bddNode(p, a)
	bddNode(p, b)
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 {
		return b
	}
	if b == 1 || a == b {
		return a
	}
	if a == b^1 {
		return 0
	}
	if a > b {
		a, b = b, a
	}
	if result := Abc_BddCacheLookup(p, a, b); result >= 0 {
		return result
	}
	va, vb := Abc_BddVar(p, a), Abc_BddVar(p, b)
	var low, high int32
	switch {
	case va < vb:
		low = Abc_BddAnd(p, Abc_BddElse(p, a), b)
		high = Abc_BddAnd(p, Abc_BddThen(p, a), b)
	case va > vb:
		low = Abc_BddAnd(p, a, Abc_BddElse(p, b))
		high = Abc_BddAnd(p, a, Abc_BddThen(p, b))
	default:
		low = Abc_BddAnd(p, Abc_BddElse(p, a), Abc_BddElse(p, b))
		high = Abc_BddAnd(p, Abc_BddThen(p, a), Abc_BddThen(p, b))
	}
	result := Abc_BddUniqueCreate(p, min(va, vb), high, low)
	return Abc_BddCacheInsert(p, a, b, result)
}
func Abc_BddOr(p *Abc_BddMan, a, b int32) int32 { return Abc_BddAnd(p, a^1, b^1) ^ 1 }
func Abc_BddCount_rec(p *Abc_BddMan, lit int32) int32 {
	node := bddNode(p, lit)
	if lit < 2 || node.mark {
		return 0
	}
	node.mark = true
	return 1 + Abc_BddCount_rec(p, Abc_BddElse(p, lit)) + Abc_BddCount_rec(p, Abc_BddThen(p, lit))
}
func Abc_BddUnmark_rec(p *Abc_BddMan, lit int32) {
	node := bddNode(p, lit)
	if lit < 2 || !node.mark {
		return
	}
	node.mark = false
	Abc_BddUnmark_rec(p, Abc_BddElse(p, lit))
	Abc_BddUnmark_rec(p, Abc_BddThen(p, lit))
}
func Abc_BddCountNodes(p *Abc_BddMan, lit int32) int32 {
	count := Abc_BddCount_rec(p, lit)
	Abc_BddUnmark_rec(p, lit)
	return count
}
func Abc_BddCountNodesArray(p *Abc_BddMan, roots *vec.Vec_Int_t) int32 {
	count := int32(0)
	for i := 0; i < vec.Vec_IntSize(roots); i++ {
		count += Abc_BddCount_rec(p, vec.Vec_IntEntry(roots, i))
	}
	for i := 0; i < vec.Vec_IntSize(roots); i++ {
		Abc_BddUnmark_rec(p, vec.Vec_IntEntry(roots, i))
	}
	return count
}
func Abc_BddCountNodesArray2(p *Abc_BddMan, roots *vec.Vec_Int_t) int32 {
	count := int32(0)
	for i := 0; i < vec.Vec_IntSize(roots); i++ {
		count += Abc_BddCountNodes(p, vec.Vec_IntEntry(roots, i))
	}
	return count
}
func bddPrint(out io.Writer, p *Abc_BddMan, lit int32, path []int32) {
	if lit == 0 {
		return
	}
	if lit == 1 {
		for i, value := range path[:p.nVars] {
			if value == 0 || value == 1 {
				sign := '-'
				if value == 1 {
					sign = '+'
				}
				fmt.Fprintf(out, "%c%d", sign, i)
			}
		}
		fmt.Fprint(out, " ")
		return
	}
	v := Abc_BddVar(p, lit)
	path[v] = 0
	bddPrint(out, p, Abc_BddElse(p, lit), path)
	path[v] = 1
	bddPrint(out, p, Abc_BddThen(p, lit), path)
	path[v] = -1
}
func Abc_BddPrint_rec(p *Abc_BddMan, lit int32, path []int32) { bddPrint(os.Stdout, p, lit, path) }
func Abc_BddPrint(p *Abc_BddMan, lit int32) {
	path := make([]int32, p.nVars)
	for i := range path {
		path[i] = -1
	}
	fmt.Printf("BDD %d = ", lit)
	Abc_BddPrint_rec(p, lit, path)
	fmt.Println()
}
func Abc_BddPrintTest(p *Abc_BddMan) {
	a := Abc_BddAnd(p, Abc_BddIthVar(0), Abc_BddIthVar(1))
	b := Abc_BddAnd(p, Abc_BddIthVar(2), Abc_BddIthVar(3))
	f := Abc_BddOr(p, a, b)
	Abc_BddPrint(p, f)
	fmt.Printf("Nodes = %d\n", Abc_BddCountNodes(p, f))
}

// Abc_BddGiaGraph supplies ABC object-ID literals, including complement bits,
// without importing gia (which may itself consume these helpers).
type Abc_BddGiaGraph interface {
	ObjectCount() int
	InputObjectIDs() []int32
	AndFaninLiterals(int32) (int32, int32, bool)
	OutputFaninLiterals() []int32
}

func bddGiaRoots(graph Abc_BddGiaGraph, limit int32) (*Abc_BddMan, *vec.Vec_Int_t) {
	inputs := graph.InputObjectIDs()
	p := Abc_BddManAlloc(int32(len(inputs)), limit)
	values := make([]int32, graph.ObjectCount())
	known := make([]bool, len(values))
	if len(values) == 0 {
		panic("missing AIG constant")
	}
	known[0] = true
	for i, id := range inputs {
		if id <= 0 || int(id) >= len(values) || known[id] {
			panic("invalid AIG input")
		}
		values[id] = Abc_BddIthVar(int32(i))
		known[id] = true
	}
	read := func(lit int32) int32 {
		if lit < 0 || int(lit>>1) >= len(values) || !known[lit>>1] {
			panic("non-topological AIG edge")
		}
		return values[lit>>1] ^ (lit & 1)
	}
	for id := int32(1); int(id) < len(values); id++ {
		a, b, ok := graph.AndFaninLiterals(id)
		if ok {
			values[id] = Abc_BddAnd(p, read(a), read(b))
			known[id] = true
		}
	}
	roots := vec.Vec_IntAlloc(len(graph.OutputFaninLiterals()))
	for _, lit := range graph.OutputFaninLiterals() {
		vec.Vec_IntPush(roots, read(lit))
	}
	return p, roots
}
func Abc_BddGiaTest(graph Abc_BddGiaGraph, verbose int32) {
	p, roots := bddGiaRoots(graph, 1<<20)
	defer Abc_BddManFree(p)
	if verbose != 0 {
		for i := 0; i < vec.Vec_IntSize(roots); i++ {
			Abc_BddPrint(p, vec.Vec_IntEntry(roots, i))
		}
	}
	fmt.Printf("Shared nodes = %d.\n", Abc_BddCountNodesArray2(p, roots))
}
func Abc_BddGiaTest2(_ Abc_BddGiaGraph, _ int32) {
	p := Abc_BddManAlloc(10, 100)
	Abc_BddPrintTest(p)
	Abc_BddManFree(p)
}
