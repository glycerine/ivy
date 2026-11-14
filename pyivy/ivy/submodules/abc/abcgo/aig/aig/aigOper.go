package aig

import "fmt"

// Port of aigOper.c. Pointer-tagged Aig_Obj_t values are represented by Edge;
// callers pass slices in place of C pointer/vector ranges.

func Aig_IthVar(p *Graph, i int32) Edge {
	if i < 0 {
		panic("aig: negative variable")
	}
	for p.InputCount() <= int(i) {
		p.NewInput()
	}
	return p.inputs[i]
}
func Aig_And(p *Graph, a, b Edge) Edge        { return p.AND(a, b) }
func Aig_Or(p *Graph, a, b Edge) Edge         { return p.OR(a, b) }
func Aig_Exor(p *Graph, a, b Edge) Edge       { return p.XOR(a, b) }
func Aig_Mux2(p *Graph, c, yes, no Edge) Edge { return p.Mux(c, yes, no) }
func Aig_Mux(p *Graph, c, yes, no Edge) Edge  { return p.Mux(c, yes, no) }
func Aig_Maj(p *Graph, a, b, c Edge) Edge     { return p.OR(p.OR(p.AND(a, b), p.AND(a, c)), p.AND(b, c)) }

// Aig_Oper preserves ABC's numeric AND/EXOR operation codes (5 and 6).
func Aig_Oper(p *Graph, a, b Edge, kind int32) Edge {
	switch kind {
	case 5:
		return p.AND(a, b)
	case 6:
		return p.XOR(a, b)
	default:
		panic("aig: unsupported operation")
	}
}
func Aig_Multi_rec(p *Graph, objects []Edge, n int32, kind int32) Edge {
	if n <= 0 || int(n) > len(objects) {
		panic("aig: invalid operand range")
	}
	if n == 1 {
		return objects[0]
	}
	half := n / 2
	return Aig_Oper(p, Aig_Multi_rec(p, objects[:half], half, kind), Aig_Multi_rec(p, objects[half:n], n-half, kind), kind)
}
func Aig_Multi(p *Graph, objects []Edge, n int32, kind int32) Edge {
	if kind != 5 && kind != 6 {
		panic("aig: unsupported operation")
	}
	return Aig_Multi_rec(p, objects, n, kind)
}

// Aig_Miter overwrites the first half of pairs with XNOR terms, as C does.
// A Go slice's length cannot be shortened for the caller; only pairs[:len(pairs)/2]
// contains the compacted terms after the call.
func Aig_Miter(p *Graph, pairs []Edge) Edge {
	if len(pairs) == 0 || len(pairs)%2 != 0 {
		panic("aig: invalid miter pairs")
	}
	for i := 0; i < len(pairs); i += 2 {
		pairs[i/2] = p.XOR(pairs[i], pairs[i+1]).Not()
	}
	return Aig_Multi_rec(p, pairs, int32(len(pairs)/2), 5).Not()
}

// Aig_MiterTwo replaces the first vector's entries with XNOR terms.
func Aig_MiterTwo(p *Graph, a, b []Edge) Edge {
	if len(a) == 0 || len(a) != len(b) {
		panic("aig: invalid miter vectors")
	}
	for i := range a {
		a[i] = p.XOR(a[i], b[i]).Not()
	}
	return Aig_Multi_rec(p, a, int32(len(a)), 5).Not()
}
func Aig_CreateAnd(p *Graph, n int32) Edge {
	out := p.Const1()
	for i := int32(0); i < n; i++ {
		out = p.AND(out, Aig_IthVar(p, i))
	}
	return out
}
func Aig_CreateOr(p *Graph, n int32) Edge {
	out := p.Const0()
	for i := int32(0); i < n; i++ {
		out = p.OR(out, Aig_IthVar(p, i))
	}
	return out
}
func Aig_CreateExor(p *Graph, n int32) Edge {
	out := p.Const0()
	for i := int32(0); i < n; i++ {
		out = p.XOR(out, Aig_IthVar(p, i))
	}
	return out
}

// Aig_TableLookupInt returns an existing AND without allocating. The boolean
// replaces the original C NULL sentinel, since edge zero is the constant one.
func Aig_TableLookupInt(p *Graph, a, b Edge) (Edge, bool) {
	if !p.valid(a) || !p.valid(b) {
		panic("aig: invalid lookup edge")
	}
	if a == b {
		return a, true
	}
	if a == 1 || b == 1 || a == b.Not() {
		return 1, true
	}
	if a == 0 {
		return b, true
	}
	if b == 0 {
		return a, true
	}
	if a > b {
		a, b = b, a
	}
	e, ok := p.hash[edgePair{a, b}]
	return e, ok
}
func Aig_ObjIsExorType(p *Graph, a, b Edge, fan0, fan1 *Edge) bool {
	if !p.valid(a) || !p.valid(b) || !a.Complemented() || !b.Complemented() {
		return false
	}
	x, y := p.nodes[a.NodeID()], p.nodes[b.NodeID()]
	if x.Kind != And || y.Kind != And {
		return false
	}
	if x.Fanins[0].NodeID() != y.Fanins[0].NodeID() || x.Fanins[1].NodeID() != y.Fanins[1].NodeID() {
		return false
	}
	if x.Fanins[0].Complemented() == y.Fanins[0].Complemented() || x.Fanins[1].Complemented() == y.Fanins[1].Complemented() {
		return false
	}
	*fan0, *fan1 = x.Fanins[0], x.Fanins[1]
	return true
}

// muxExperimentGraph builds the source's 2,000-MUX seeded experiment.
func muxExperimentGraph() *Graph {
	p := NewGraph()
	const nPIs, nNodes = 20, 2000
	for i := int32(0); i < nPIs; i++ {
		Aig_IthVar(p, i)
	}
	random := newMuxExperimentRandom()
	nodes := make([]Edge, 0, nNodes)
	operand := func(i int) Edge {
		var e Edge
		switch {
		case random()%10 == 0:
			e = p.Const0()
		case random()%10 == 0:
			e = p.Const1()
		case random()%3 == 0 || i < nPIs:
			e = Aig_IthVar(p, int32(random()%nPIs))
		default:
			e = nodes[random()%i]
		}
		return dupPhase(e, random()%2 == 0)
	}
	for i := 0; i < nNodes; i++ {
		control, yes, no := operand(i), operand(i), operand(i)
		nodes = append(nodes, Aig_Mux(p, control, yes, no))
	}
	for _, e := range nodes {
		p.AddOutput(e)
	}
	return p
}

// newMuxExperimentRandom reproduces Linux/glibc srand(321), privately so the
// experiment does not reset random streams used by other packages. C rand
// sequences on other platforms are implementation dependent.
func newMuxExperimentRandom() func() int {
	var state [31]uint32
	value := int64(321)
	state[0] = uint32(value)
	for i := 1; i < len(state); i++ {
		value = value * 16807 % 2147483647
		state[i] = uint32(value)
	}
	front, rear := 3, 0
	next := func() int {
		state[front] += state[rear]
		value := state[front] >> 1
		front, rear = (front+1)%31, (rear+1)%31
		return int(value)
	}
	for i := 0; i < 310; i++ {
		next()
	}
	return next
}

// Aig_MuxTest constructs and cleans up the original seeded MUX experiment,
// prints both object counts, and writes the resulting circuit to test1.blif.
func Aig_MuxTest() {
	p := muxExperimentGraph()
	fmt.Printf("Number of nodes = %6d.\n", Aig_ManObjNum(p))
	Aig_ManCleanup(p)
	fmt.Printf("Number of nodes = %6d.\n", Aig_ManObjNum(p))
	if err := Aig_ManDumpBlif(p, "test1.blif", nil, nil); err != nil {
		fmt.Println(err)
	}
	Aig_ManStop(p)
}
