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
	return graphTruth32(g, len(g.Nodes), g.LeafCount, g.Root, g.Constant)
}

func graphTruth32[G *Graph | *Dec_Graph_t](g G, size, leaves int, root Edge, constant bool) uint32 {
	if leaves < 0 || leaves > 5 {
		panic("dec: Truth32 requires zero to five leaves")
	}
	if constant {
		if root.Complement {
			return 0
		}
		return ^uint32(0)
	}
	if leaves > size || uint(root.Node) >= uint(size) {
		panic("dec: invalid evaluation")
	}
	// Keep the small graphs in the rewrite library on the stack.
	var local [64]uint32
	words := local[:]
	if size > len(words) {
		words = make([]uint32, size)
	}
	variables := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	copy(words, variables[:leaves])
	switch graph := any(g).(type) {
	case *Graph:
		for i := leaves; i < size; i++ {
			inputs := words[:i]
			node := graph.Nodes[i]
			words[i] = edgeTruth(inputs, node.Left) & edgeTruth(inputs, node.Right)
		}
	case *Dec_Graph_t:
		i := leaves
		for ; i < len(graph.first); i++ {
			words[i] = andTruthWord(&graph.first[i], words[:i])
		}
		for ; i < size; i++ {
			words[i] = andTruthWord(graph.node(uint32(i)), words[:i])
		}
	}
	return edgeTruth(words, root)
}

func andTruthWord(n *Dec_Node_t, inputs []uint32) uint32 {
	a, b := n.eEdge0, n.eEdge1
	return (inputs[a.Node] ^ -a.fCompl) & (inputs[b.Node] ^ -b.fCompl)
}

func edgeTruth(words []uint32, e Edge) uint32 {
	value := words[e.Node]
	if e.Complement {
		return ^value
	}
	return value
}
