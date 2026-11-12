package dec

// Edge identifies a node, optionally complemented. Leaves occupy the first
// LeafCount node slots, followed by AND nodes in topological order.
type Edge struct {
	Node       int
	Complement bool
}

func (e Edge) Not() Edge { e.Complement = !e.Complement; return e }
func (e Edge) Literal() uint32 {
	if e.Node < 0 || e.Node >= 1<<30 {
		panic("dec: edge index out of range")
	}
	v := uint32(e.Node) << 1
	if e.Complement {
		v |= 1
	}
	return v
}
func EdgeFromLiteral(v uint32) Edge { return Edge{int(v >> 1), v&1 != 0} }

type Node struct{ Left, Right Edge }
type Graph struct {
	LeafCount int
	Nodes     []Node
	Root      Edge
	Constant  bool
}

func NewGraph(leaves int) *Graph {
	if leaves < 0 {
		panic("dec: negative leaf count")
	}
	return &Graph{LeafCount: leaves, Nodes: make([]Node, leaves)}
}
func NewConstant(value bool) *Graph { return &Graph{Constant: true, Root: Edge{Complement: !value}} }
func (g *Graph) valid(e Edge) bool  { return e.Node >= 0 && e.Node < len(g.Nodes) }
func (g *Graph) And(a, b Edge) Edge {
	if g.Constant || !g.valid(a) || !g.valid(b) {
		panic("dec: invalid AND operand")
	}
	n := Edge{Node: len(g.Nodes)}
	g.Nodes = append(g.Nodes, Node{a, b})
	return n
}
func (g *Graph) Or(a, b Edge) Edge  { return g.And(a.Not(), b.Not()).Not() }
func (g *Graph) Xor(a, b Edge) Edge { return g.Or(g.And(a, b.Not()), g.And(a.Not(), b)) }
func (g *Graph) Mux(control, then, otherwise Edge) Edge {
	return g.Or(g.And(control, then), g.And(control.Not(), otherwise))
}
func (g *Graph) SetRoot(root Edge) {
	if !g.Constant && !g.valid(root) {
		panic("dec: invalid root")
	}
	g.Root = root
}
func (g *Graph) Complement() { g.Root = g.Root.Not() }

// Evaluate uses a separate scratch slice, so evaluating a graph does not mutate
// nodes and concurrent readers can safely share an immutable graph.
func (g *Graph) Evaluate(inputs []bool) bool {
	if g.Constant {
		return !g.Root.Complement
	}
	if len(inputs) != g.LeafCount || !g.valid(g.Root) {
		panic("dec: invalid evaluation")
	}
	values := make([]bool, len(g.Nodes))
	copy(values, inputs)
	value := func(e Edge) bool {
		v := values[e.Node]
		if e.Complement {
			return !v
		}
		return v
	}
	for i := g.LeafCount; i < len(g.Nodes); i++ {
		n := g.Nodes[i]
		if n.Left.Node < 0 || n.Right.Node < 0 || n.Left.Node >= i || n.Right.Node >= i {
			panic("dec: graph is not topological")
		}
		values[i] = value(n.Left) && value(n.Right)
	}
	return value(g.Root)
}

// Truth32 reproduces Dec_GraphDeriveTruth's 32-bit truth convention. Small
// functions repeat across the word, with variable zero alternating each bit.
func (g *Graph) Truth32() uint32 {
	if g.LeafCount > 5 {
		panic("dec: Truth32 requires at most five leaves")
	}
	inputs := make([]bool, g.LeafCount)
	var result uint32
	for assignment := 0; assignment < 32; assignment++ {
		for i := range inputs {
			inputs[i] = assignment&(1<<i) != 0
		}
		if g.Evaluate(inputs) {
			result |= 1 << assignment
		}
	}
	return result
}
