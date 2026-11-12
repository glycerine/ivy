package gia

// Literal packs an object index and its complemented-edge bit as in ABC GIA.
// Object zero is false; its complemented literal is true.
type Literal uint32

func (l Literal) Object() uint32     { return uint32(l) >> 1 }
func (l Literal) Complemented() bool { return l&1 != 0 }
func (l Literal) Not() Literal       { return l ^ 1 }

type graphKind uint8

const (
	graphConstant graphKind = iota
	graphInput
	graphAnd
)

type graphNode struct {
	kind graphKind
	a, b Literal
}
type andKey struct{ a, b Literal }

// Graph is a Go representation of GIA's combinational AND/inverter core.
// Nodes are topologically ordered; edge inversion lives in literals, not nodes.
type Graph struct {
	nodes   []graphNode
	inputs  []Literal
	outputs []Literal
	hash    map[andKey]Literal
}

func NewGraph() *Graph {
	return &Graph{nodes: []graphNode{{kind: graphConstant}}, hash: make(map[andKey]Literal)}
}
func (g *Graph) check(l Literal) {
	if int(l.Object()) >= len(g.nodes) {
		panic("gia: literal references nonexistent object")
	}
}
func (g *Graph) NewInput() Literal {
	l := Literal(len(g.nodes) << 1)
	g.nodes = append(g.nodes, graphNode{kind: graphInput})
	g.inputs = append(g.inputs, l)
	return l
}

// And performs the constant, idempotence and opposite-input reductions used
// by structurally hashed AIGs, then canonicalizes the unordered fanin pair.
func (g *Graph) And(a, b Literal) Literal {
	g.check(a)
	g.check(b)
	if a == 0 || b == 0 {
		return 0
	}
	if a == 1 {
		return b
	}
	if b == 1 {
		return a
	}
	if a == b {
		return a
	}
	if a == b.Not() {
		return 0
	}
	if a > b {
		a, b = b, a
	}
	key := andKey{a, b}
	if l, ok := g.hash[key]; ok {
		return l
	}
	l := Literal(len(g.nodes) << 1)
	g.nodes = append(g.nodes, graphNode{kind: graphAnd, a: a, b: b})
	g.hash[key] = l
	return l
}
func (g *Graph) Or(a, b Literal) Literal { return g.And(a.Not(), b.Not()).Not() }
func (g *Graph) Xor(a, b Literal) Literal {
	return g.Or(g.And(a, b.Not()), g.And(a.Not(), b))
}
func (g *Graph) Mux(control, yes, no Literal) Literal {
	return g.Or(g.And(control, yes), g.And(control.Not(), no))
}
func (g *Graph) AddOutput(l Literal) { g.check(l); g.outputs = append(g.outputs, l) }
func (g *Graph) ObjectCount() int    { return len(g.nodes) }
func (g *Graph) InputCount() int     { return len(g.inputs) }
func (g *Graph) OutputCount() int    { return len(g.outputs) }

// Evaluate computes outputs in their insertion order. The returned slice is
// independent of the graph and the caller's input slice.
func (g *Graph) Evaluate(inputs []bool) []bool {
	if len(inputs) != len(g.inputs) {
		panic("gia: input count mismatch")
	}
	values := make([]bool, len(g.nodes))
	for i, l := range g.inputs {
		values[l.Object()] = inputs[i]
	}
	value := func(l Literal) bool { return values[l.Object()] != l.Complemented() }
	for i, n := range g.nodes {
		if n.kind == graphAnd {
			values[i] = value(n.a) && value(n.b)
		}
	}
	outputs := make([]bool, len(g.outputs))
	for i, l := range g.outputs {
		outputs[i] = value(l)
	}
	return outputs
}

// Levels returns the longest AND path from any input for each output.
func (g *Graph) Levels() []int {
	levels := make([]int, len(g.nodes))
	for i, n := range g.nodes {
		if n.kind == graphAnd {
			a, b := levels[n.a.Object()], levels[n.b.Object()]
			if b > a {
				a = b
			}
			levels[i] = a + 1
		}
	}
	outputs := make([]int, len(g.outputs))
	for i, l := range g.outputs {
		outputs[i] = levels[l.Object()]
	}
	return outputs
}
