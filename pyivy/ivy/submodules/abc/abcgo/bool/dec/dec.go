package dec

// C-compatible entry points for decomposition graphs. Node pointers remain
// stable when the graph grows.
type Dec_Edge_t struct {
	fCompl uint32
	Node   uint32
}
type Dec_Node_t struct {
	pFunc          any
	eEdge0, eEdge1 Dec_Edge_t
	index, iFunc   int32
	Level          uint32
	printing       uint8 // Original OR and input complement bits, in that order.
	latches        [3]uint8
}

const decBlockSize = 32

type Dec_Graph_t struct {
	fConst  int32
	nLeaves int32
	size    int32
	first   []Dec_Node_t
	rest    [][]Dec_Node_t
	eRoot   Dec_Edge_t
}
type Dec_Man_t struct {
	puCanons              []uint16
	pPhases, pPerms, pMap []byte
}

func decBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Dec_EdgeCreate(node, compl int32) Dec_Edge_t {
	return Dec_Edge_t{uint32(compl) & 1, uint32(node)}.normalized()
}
func (e Dec_Edge_t) normalized() Dec_Edge_t {
	e.Node &= 1<<30 - 1
	return e
}
func Dec_EdgeToInt(e Dec_Edge_t) uint32      { return e.normalized().Node<<1 | e.fCompl }
func Dec_IntToEdge(e uint32) Dec_Edge_t      { return Dec_EdgeCreate(int32(e>>1), int32(e&1)) }
func Dec_EdgeIsComplement(e Dec_Edge_t) bool { return e.fCompl != 0 }
func Dec_EdgeNot(e Dec_Edge_t) Dec_Edge_t    { e.fCompl ^= 1; return e.normalized() }

// The historical union variant uses native bitfield layout: complement first.
func Dec_EdgeToInt_(e Dec_Edge_t) uint32 { return Dec_EdgeToInt(e) }
func Dec_IntToEdge_(e uint32) Dec_Edge_t { return Dec_IntToEdge(e) }
func Dec_GraphCreate(leaves int32) *Dec_Graph_t {
	if leaves < 0 || leaves > 1<<30 {
		panic("dec: leaf count outside edge index range")
	}
	// Most two-leaf graphs have one gate, and four-leaf rewrite candidates commonly have seven.
	leavesInBlock := min(leaves, decBlockSize)
	firstCap := min(decBlockSize, max(leavesInBlock, 4*leavesInBlock-5))
	return newDecGraph(leaves, leaves, firstCap)
}

// Dec_GraphCreateSized creates all leaf and internal nodes of a graph whose
// final topology is known. Separate blocks keep pointers stable on later
// appends and bound the number of siblings retained by a saved node.
func Dec_GraphCreateSized(leaves, internal int32) *Dec_Graph_t {
	if leaves < 0 || internal < 0 || int64(leaves)+int64(internal) > 1<<30 {
		panic("dec: graph size outside edge index range")
	}
	size := leaves + internal
	firstCap := size
	if size > decBlockSize {
		// Put the remainder first so the following blocks are full.
		firstCap = 1 + (size-1)%decBlockSize
	}
	return newDecGraph(leaves, size, firstCap)
}

func newDecGraph(leaves, size, firstCap int32) *Dec_Graph_t {
	g := &Dec_Graph_t{nLeaves: leaves, size: size, first: make([]Dec_Node_t, min(size, firstCap), firstCap)}
	for i := range g.first {
		g.first[i].index = int32(i)
	}
	remaining := int(size) - len(g.first)
	if remaining > 0 {
		g.rest = make([][]Dec_Node_t, 0, (remaining+decBlockSize-1)/decBlockSize)
		for offset := len(g.first); offset < int(size); {
			block := make([]Dec_Node_t, min(decBlockSize, int(size)-offset), decBlockSize)
			for i := range block {
				block[i].index = int32(offset + i)
			}
			g.rest = append(g.rest, block)
			offset += len(block)
		}
	}
	return g
}

func (g *Dec_Graph_t) node(index uint32) *Dec_Node_t {
	if index < uint32(len(g.first)) {
		return &g.first[index]
	}
	offset := index - uint32(len(g.first))
	return &g.rest[offset/decBlockSize][offset%decBlockSize]
}

func Dec_GraphCreateConst0() *Dec_Graph_t {
	return &Dec_Graph_t{fConst: 1, eRoot: Dec_Edge_t{fCompl: 1}}
}
func Dec_GraphCreateConst1() *Dec_Graph_t { return &Dec_Graph_t{fConst: 1} }
func Dec_GraphCreateLeaf(i, n, c int32) *Dec_Graph_t {
	g := Dec_GraphCreate(n)
	if i < 0 || i >= n {
		panic("dec: leaf index")
	}
	g.eRoot = Dec_EdgeCreate(i, c)
	return g
}
func Dec_GraphFree(g *Dec_Graph_t) {
	if g != nil {
		g.first, g.rest, g.size = nil, nil, 0
	}
}
func Dec_GraphIsConst(g *Dec_Graph_t) int32             { return g.fConst }
func Dec_GraphIsConst0(g *Dec_Graph_t) int32            { return decBool(g.fConst != 0 && g.eRoot.fCompl != 0) }
func Dec_GraphIsConst1(g *Dec_Graph_t) int32            { return decBool(g.fConst != 0 && g.eRoot.fCompl == 0) }
func Dec_GraphIsComplement(g *Dec_Graph_t) int32        { return int32(g.eRoot.fCompl) }
func Dec_GraphComplement(g *Dec_Graph_t)                { g.eRoot.fCompl ^= 1 }
func Dec_GraphLeaveNum(g *Dec_Graph_t) int32            { return g.nLeaves }
func Dec_GraphNodeNum(g *Dec_Graph_t) int32             { return g.size - g.nLeaves }
func Dec_GraphNode(g *Dec_Graph_t, i int32) *Dec_Node_t { return g.node(uint32(i)) }
func Dec_GraphNodeLast(g *Dec_Graph_t) *Dec_Node_t      { return g.node(uint32(g.size - 1)) }

// Dec_GraphSetLeafFunc assigns the caller-owned signal represented by a graph
// leaf. Payloads remain opaque to this package and are consumed by network
// adapters during count/build operations.
func Dec_GraphSetLeafFunc(g *Dec_Graph_t, leaf int32, value any) {
	if g == nil || leaf < 0 || leaf >= g.nLeaves {
		panic("dec: leaf function index")
	}
	g.node(uint32(leaf)).pFunc = value
}

func Dec_GraphLeafFunc(g *Dec_Graph_t, leaf int32) any {
	if g == nil || leaf < 0 || leaf >= g.nLeaves {
		panic("dec: leaf function index")
	}
	return g.node(uint32(leaf)).pFunc
}

// Dec_GraphSetNodeFunc and Dec_GraphNodeFunc expose opaque node payloads for
// clients that incrementally evaluate decomposition candidates.
func Dec_GraphSetNodeFunc(g *Dec_Graph_t, index int32, value any) {
	if g == nil || index < 0 || index >= g.size {
		panic("dec: node function index")
	}
	g.node(uint32(index)).pFunc = value
}

func Dec_GraphNodeFunc(g *Dec_Graph_t, index int32) any {
	if g == nil || index < 0 || index >= g.size {
		panic("dec: node function index")
	}
	return g.node(uint32(index)).pFunc
}
func Dec_GraphNodeInt(g *Dec_Graph_t, n *Dec_Node_t) int32 {
	if n != nil && n.index >= 0 && n.index < g.size && g.node(uint32(n.index)) == n {
		return n.index
	}
	for i := range g.size {
		if g.node(uint32(i)) == n {
			return i
		}
	}
	panic("dec: foreign node")
}
func Dec_GraphIsVar(g *Dec_Graph_t) int32 { return decBool(g.eRoot.Node < uint32(g.nLeaves)) }
func Dec_GraphNodeIsVar(g *Dec_Graph_t, n *Dec_Node_t) int32 {
	return decBool(Dec_GraphNodeInt(g, n) < g.nLeaves)
}
func Dec_GraphVar(g *Dec_Graph_t) *Dec_Node_t { return Dec_GraphNode(g, Dec_GraphVarInt(g)) }
func Dec_GraphVarInt(g *Dec_Graph_t) int32 {
	if Dec_GraphIsVar(g) == 0 {
		panic("dec: root is not a variable")
	}
	return int32(g.eRoot.Node)
}
func Dec_GraphSetRoot(g *Dec_Graph_t, e Dec_Edge_t) { g.eRoot = e.normalized() }
func Dec_GraphAppendNode(g *Dec_Graph_t) *Dec_Node_t {
	if g.size >= 1<<30 {
		panic("dec: graph exceeds edge index range")
	}
	var n *Dec_Node_t
	if len(g.first) < cap(g.first) {
		g.first = g.first[:len(g.first)+1]
		n = &g.first[len(g.first)-1]
	} else {
		last := len(g.rest) - 1
		if last < 0 || len(g.rest[last]) == decBlockSize {
			g.rest = append(g.rest, make([]Dec_Node_t, 0, decBlockSize))
			last++
		}
		block := g.rest[last]
		block = block[:len(block)+1]
		g.rest[last] = block
		n = &block[len(block)-1]
	}
	n.index = g.size
	g.size++
	return n
}
func Dec_GraphAddNodeAnd(g *Dec_Graph_t, a, b Dec_Edge_t) Dec_Edge_t {
	n := Dec_GraphAppendNode(g)
	n.SetEdges(a, b)
	n.printing = uint8(a.fCompl<<1 | b.fCompl<<2)
	return Dec_EdgeCreate(g.size-1, 0)
}
func Dec_GraphAddNodeOr(g *Dec_Graph_t, a, b Dec_Edge_t) Dec_Edge_t {
	e := Dec_GraphAddNodeAnd(g, a, b)
	n := Dec_GraphNodeLast(g)
	n.printing |= 1
	n.eEdge0.fCompl ^= 1
	n.eEdge1.fCompl ^= 1
	e.fCompl = 1
	return e
}
func Dec_GraphAddNodeXor(g *Dec_Graph_t, a, b Dec_Edge_t, kind int32) Dec_Edge_t {
	if kind == 0 {
		na := a
		na.fCompl ^= 1
		nb := b
		nb.fCompl ^= 1
		return Dec_GraphAddNodeOr(g, Dec_GraphAddNodeAnd(g, na, b), Dec_GraphAddNodeAnd(g, a, nb))
	}
	na, nb := a, b
	na.fCompl ^= 1
	nb.fCompl ^= 1
	e := Dec_GraphAddNodeOr(g, Dec_GraphAddNodeAnd(g, a, b), Dec_GraphAddNodeAnd(g, na, nb))
	e.fCompl ^= 1
	return e
}
func Dec_GraphAddNodeMux(g *Dec_Graph_t, c, t, e Dec_Edge_t, kind int32) Dec_Edge_t {
	if kind != 0 {
		t.fCompl ^= 1
		e.fCompl ^= 1
	}
	nc := c
	nc.fCompl ^= 1
	r := Dec_GraphAddNodeOr(g, Dec_GraphAddNodeAnd(g, c, t), Dec_GraphAddNodeAnd(g, nc, e))
	if kind != 0 {
		r.fCompl ^= 1
	}
	return r
}

// SetLeafSignal binds a graph input to an adapter-owned network signal.
func (g *Dec_Graph_t) SetLeafSignal(index int, signal any) {
	if index < 0 || index >= int(g.nLeaves) {
		panic("dec: leaf index")
	}
	g.node(uint32(index)).pFunc = signal
}
func (g *Dec_Graph_t) LeafCount() int { return int(g.nLeaves) }

// Latches returns the two input-edge counts followed by the output-edge count.
func (n *Dec_Node_t) Latches() [3]uint32 {
	return [3]uint32{uint32(n.latches[0]), uint32(n.latches[1]), uint32(n.latches[2])}
}
func (n *Dec_Node_t) SetLatches(counts [3]uint32) {
	n.latches = [3]uint8{uint8(counts[0] & 31), uint8(counts[1] & 31), uint8(counts[2] & 31)}
}

// PrintingFlags returns original OR, input-zero complement, input-one complement.
func (n *Dec_Node_t) PrintingFlags() [3]uint32 {
	flags := uint32(n.printing)
	return [3]uint32{flags & 1, flags >> 1 & 1, flags >> 2 & 1}
}
func (n *Dec_Node_t) SetPrintingFlags(flags [3]uint32) {
	n.printing = uint8((flags[0] & 1) | (flags[1]&1)<<1 | (flags[2]&1)<<2)
}

// SetLevel preserves the original node's unsigned fourteen-bit field semantics.
func (n *Dec_Node_t) SetLevel(level uint32) { n.Level = level & ((1 << 14) - 1) }

// Edges returns the ordered input edges of a decomposition node.
func (n *Dec_Node_t) Edges() (Dec_Edge_t, Dec_Edge_t) {
	if n == nil {
		panic("dec: nil node")
	}
	return n.eEdge0, n.eEdge1
}

// SetEdges assigns the ordered input edges of a decomposition node.
func (n *Dec_Node_t) SetEdges(a, b Dec_Edge_t) {
	if n == nil {
		panic("dec: nil node")
	}
	n.eEdge0, n.eEdge1 = a.normalized(), b.normalized()
}

// RootEdge returns the graph's output edge.
func Dec_GraphRootEdge(g *Dec_Graph_t) Dec_Edge_t {
	if g == nil {
		panic("dec: nil graph")
	}
	return g.eRoot
}

// Complemented reports whether this edge is complemented.
func (e Dec_Edge_t) Complemented() bool { return e.fCompl != 0 }

// Complement toggles the complemented-edge bit.
func (e *Dec_Edge_t) Complement() { *e = Dec_EdgeNot(*e) }
