package abc

// Decomposition graph storage and inline operations from src/bool/deco/deco.h.

import "math"

const (
	decoEdgeNodeBits  = 30
	decoEdgeNodeLimit = 1 << decoEdgeNodeBits
)

// DecoEdge is the typed replacement for Dec_Edge_t. Node is a stable index;
// Complement is the source bitfield fCompl.
type DecoEdge struct {
	Node       int
	Complement bool
}

// NewDecoEdge implements Dec_EdgeCreate with checked bitfield bounds.
func NewDecoEdge(node int, complement bool) (DecoEdge, error) {
	if node < 0 || node >= decoEdgeNodeLimit {
		return DecoEdge{}, ErrDimension
	}
	return DecoEdge{Node: node, Complement: complement}, nil
}

// Uint32 implements both Dec_EdgeToInt and the bitfield-layout helper
// Dec_EdgeToInt_. Go gives the representation an explicit portable layout.
func (edge DecoEdge) Uint32() (uint32, error) {
	if edge.Node < 0 || edge.Node >= decoEdgeNodeLimit {
		return 0, ErrDimension
	}
	value := uint32(edge.Node) << 1
	if edge.Complement {
		value |= 1
	}
	return value, nil
}

// DecoEdgeFromUint32 implements Dec_IntToEdge and Dec_IntToEdge_. The source
// bitfields occupy 31 bits; accepting bit 31 would silently truncate Node.
func DecoEdgeFromUint32(value uint32) (DecoEdge, error) {
	if value >= 1<<31 {
		return DecoEdge{}, ErrDimension
	}
	return DecoEdge{Node: int(value >> 1), Complement: value&1 != 0}, nil
}

// Not returns the same edge in the opposite polarity.
func (edge DecoEdge) Not() DecoEdge {
	edge.Complement = !edge.Complement
	return edge
}

// DecoNode is the owned Dec_Node_t representation. The source's iFunc/pFunc
// union is retained as one machine word and exposed through typed accessors.
type DecoNode struct {
	Edge0, Edge1                 DecoEdge
	functionWord                 uintptr
	Level                        uint16
	NodeOr, OriginalComplement0  bool
	OriginalComplement1          bool
	Latches0, Latches1, Latches2 uint8
}

// FunctionLiteral returns the signed 32-bit iFunc interpretation of the
// source union.
func (node DecoNode) FunctionLiteral() int {
	return int(int32(uint32(node.functionWord)))
}

// SetFunctionLiteral stores the signed 32-bit iFunc interpretation of the
// source union. Nodes are zero-initialized just as in Dec_GraphAppendNode, so
// the upper portion of a wider pointer word remains zero.
func (node *DecoNode) SetFunctionLiteral(literal int) error {
	if node == nil {
		return ErrInvariant
	}
	if literal < -1<<31 || literal > 1<<31-1 {
		return ErrDimension
	}
	node.functionWord = uintptr(uint32(int32(literal)))
	return nil
}

// FunctionToken returns the pFunc interpretation of the source union as an
// opaque non-owning token.
func (node DecoNode) FunctionToken() uintptr { return node.functionWord }

// SetFunctionToken stores the pFunc interpretation of the source union as an
// opaque non-owning token.
func (node *DecoNode) SetFunctionToken(token uintptr) error {
	if node == nil {
		return ErrInvariant
	}
	node.functionWord = token
	return nil
}

// DecoManagerState is the owned storage represented by Dec_Man_t. The header
// defines no manager algorithms; concrete factoring code may fill these
// arrays without a generic allocator or callback abstraction.
type DecoManagerState struct {
	MVC                  *MvcManager
	Cubes, Literals      []int
	Canons               []uint16
	Phases, Permutations []int8
	ClassMap             []byte
}

// DecoGraphLimits bounds retained graph storage. Zero limits select package
// defaults. Stop is polled before allocations and mutations.
type DecoGraphLimits struct {
	MaxLeaves int
	MaxNodes  int
	Stop      func() bool
}

func normalizeDecoGraphLimits(limits DecoGraphLimits) (DecoGraphLimits, error) {
	if limits.MaxLeaves < 0 || limits.MaxNodes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxLeaves == 0 {
		limits.MaxLeaves = 1 << 20
	}
	if limits.MaxNodes == 0 {
		limits.MaxNodes = 1 << 22
	}
	if limits.MaxLeaves > limits.MaxNodes {
		limits.MaxLeaves = limits.MaxNodes
	}
	return limits, nil
}

// DecoGraph is an owned Dec_Graph_t. Nodes retains the source distinction
// between the leading leaves and appended internal nodes.
type DecoGraph struct {
	constant bool
	leaves   int
	nodes    []DecoNode
	root     DecoEdge
	limits   DecoGraphLimits
	closed   bool
}

func decoInitialCapacity(leaves int) (int, error) {
	if leaves > (math.MaxInt-50)/2 {
		return 0, ErrLimit
	}
	return 2*leaves + 50, nil
}

// NewDecoGraph implements Dec_GraphCreate.
func NewDecoGraph(leaves int, limits DecoGraphLimits) (*DecoGraph, error) {
	limits, err := normalizeDecoGraphLimits(limits)
	if err != nil || leaves < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if leaves > limits.MaxLeaves || leaves > limits.MaxNodes || leaves >= decoEdgeNodeLimit {
		return nil, ErrLimit
	}
	capacity, err := decoInitialCapacity(leaves)
	if err != nil {
		return nil, err
	}
	if capacity > limits.MaxNodes {
		capacity = limits.MaxNodes
	}
	if capacity < leaves {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &DecoGraph{leaves: leaves, nodes: make([]DecoNode, leaves, capacity), limits: limits}, nil
}

func newConstantDecoGraph(value bool, limits DecoGraphLimits) (*DecoGraph, error) {
	limits, err := normalizeDecoGraphLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &DecoGraph{constant: true, root: DecoEdge{Complement: !value}, limits: limits}, nil
}

func NewConstantDecoGraph(value bool, limits DecoGraphLimits) (*DecoGraph, error) {
	return newConstantDecoGraph(value, limits)
}

// NewLeafDecoGraph implements Dec_GraphCreateLeaf.
func NewLeafDecoGraph(leaf, leaves int, complement bool, limits DecoGraphLimits) (*DecoGraph, error) {
	if leaf < 0 || leaf >= leaves {
		return nil, ErrDimension
	}
	graph, err := NewDecoGraph(leaves, limits)
	if err != nil {
		return nil, err
	}
	graph.root = DecoEdge{Node: leaf, Complement: complement}
	return graph, nil
}

func (graph *DecoGraph) check() error {
	if graph == nil || graph.closed {
		return ErrInvariant
	}
	if graph.leaves < 0 || graph.leaves > len(graph.nodes) || len(graph.nodes) > graph.limits.MaxNodes || cap(graph.nodes) > graph.limits.MaxNodes {
		return ErrInvariant
	}
	if graph.constant {
		if graph.leaves != 0 || len(graph.nodes) != 0 {
			return ErrInvariant
		}
		return nil
	}
	for index := graph.leaves; index < len(graph.nodes); index++ {
		node := graph.nodes[index]
		if node.Edge0.Node < 0 || node.Edge0.Node >= index || node.Edge1.Node < 0 || node.Edge1.Node >= index || node.Level >= 1<<14 || node.Latches0 >= 1<<5 || node.Latches1 >= 1<<5 || node.Latches2 >= 1<<5 {
			return ErrInvariant
		}
	}
	return nil
}

func (graph *DecoGraph) stopped() bool {
	return graph.limits.Stop != nil && graph.limits.Stop()
}

func (graph *DecoGraph) IsConstant() bool { return graph != nil && !graph.closed && graph.constant }

func (graph *DecoGraph) ConstantValue() (bool, bool) {
	if !graph.IsConstant() {
		return false, false
	}
	return !graph.root.Complement, true
}

func (graph *DecoGraph) IsComplemented() (bool, error) {
	if err := graph.check(); err != nil {
		return false, err
	}
	return graph.root.Complement, nil
}

func (graph *DecoGraph) Complement() error {
	if err := graph.check(); err != nil {
		return err
	}
	if graph.stopped() {
		return ErrCancelled
	}
	graph.root.Complement = !graph.root.Complement
	return nil
}

func (graph *DecoGraph) LeafCount() (int, error) {
	if err := graph.check(); err != nil {
		return 0, err
	}
	return graph.leaves, nil
}

func (graph *DecoGraph) InternalCount() (int, error) {
	if err := graph.check(); err != nil {
		return 0, err
	}
	return len(graph.nodes) - graph.leaves, nil
}

func (graph *DecoGraph) Capacity() (int, error) {
	if err := graph.check(); err != nil {
		return 0, err
	}
	return cap(graph.nodes), nil
}

func (graph *DecoGraph) Root() (DecoEdge, error) {
	if err := graph.check(); err != nil {
		return DecoEdge{}, err
	}
	if !graph.constant && (graph.root.Node < 0 || graph.root.Node >= len(graph.nodes)) {
		return DecoEdge{}, ErrInvariant
	}
	return graph.root, nil
}

func (graph *DecoGraph) SetRoot(root DecoEdge) error {
	if err := graph.check(); err != nil {
		return err
	}
	if graph.constant || root.Node < 0 || root.Node >= len(graph.nodes) {
		return ErrDimension
	}
	if graph.stopped() {
		return ErrCancelled
	}
	graph.root = root
	return nil
}

func (graph *DecoGraph) Node(index int) (DecoNode, error) {
	if err := graph.check(); err != nil {
		return DecoNode{}, err
	}
	if index < 0 || index >= len(graph.nodes) {
		return DecoNode{}, ErrDimension
	}
	return graph.nodes[index], nil
}

// WriteNode replaces mutation through Dec_GraphNode's borrowed pointer while
// retaining stable indexes and graph invariants.
func (graph *DecoGraph) WriteNode(index int, node DecoNode) error {
	if err := graph.check(); err != nil {
		return err
	}
	if index < 0 || index >= len(graph.nodes) {
		return ErrDimension
	}
	if node.Level >= 1<<14 || node.Latches0 >= 1<<5 || node.Latches1 >= 1<<5 || node.Latches2 >= 1<<5 {
		return ErrDimension
	}
	if index >= graph.leaves && (node.Edge0.Node < 0 || node.Edge0.Node >= index || node.Edge1.Node < 0 || node.Edge1.Node >= index) {
		return ErrDimension
	}
	if graph.stopped() {
		return ErrCancelled
	}
	graph.nodes[index] = node
	return nil
}

func (graph *DecoGraph) LastNode() (DecoNode, int, error) {
	if err := graph.check(); err != nil {
		return DecoNode{}, -1, err
	}
	if len(graph.nodes) == 0 {
		return DecoNode{}, -1, ErrDimension
	}
	return graph.nodes[len(graph.nodes)-1], len(graph.nodes) - 1, nil
}

func (graph *DecoGraph) IsVariable() (bool, error) {
	if err := graph.check(); err != nil {
		return false, err
	}
	return !graph.constant && graph.root.Node >= 0 && graph.root.Node < graph.leaves, nil
}

func (graph *DecoGraph) NodeIsVariable(index int) (bool, error) {
	if _, err := graph.Node(index); err != nil {
		return false, err
	}
	return index < graph.leaves, nil
}

func (graph *DecoGraph) Variable() (int, error) {
	variable, err := graph.IsVariable()
	if err != nil {
		return 0, err
	}
	if !variable {
		return 0, ErrDimension
	}
	return graph.root.Node, nil
}

// Leaves and Internals replace the two pointer iterator macros with owned
// index/snapshot records in identical order.
type DecoNodeRecord struct {
	Index int
	Node  DecoNode
}

func (graph *DecoGraph) nodeRecords(begin, end int) ([]DecoNodeRecord, error) {
	if err := graph.check(); err != nil {
		return nil, err
	}
	if graph.stopped() {
		return nil, ErrCancelled
	}
	records := make([]DecoNodeRecord, 0, end-begin)
	for index := begin; index < end; index++ {
		if graph.stopped() {
			return nil, ErrCancelled
		}
		records = append(records, DecoNodeRecord{Index: index, Node: graph.nodes[index]})
	}
	return records, nil
}

func (graph *DecoGraph) Leaves() ([]DecoNodeRecord, error) {
	if err := graph.check(); err != nil {
		return nil, err
	}
	return graph.nodeRecords(0, graph.leaves)
}

func (graph *DecoGraph) Internals() ([]DecoNodeRecord, error) {
	if err := graph.check(); err != nil {
		return nil, err
	}
	return graph.nodeRecords(graph.leaves, len(graph.nodes))
}

func (graph *DecoGraph) validateEdge(edge DecoEdge) error {
	if edge.Node < 0 || edge.Node >= len(graph.nodes) || edge.Node >= decoEdgeNodeLimit {
		return ErrDimension
	}
	return nil
}

func (graph *DecoGraph) appendNode(node DecoNode) (DecoEdge, error) {
	if err := graph.check(); err != nil {
		return DecoEdge{}, err
	}
	if graph.constant {
		return DecoEdge{}, ErrDimension
	}
	if err := graph.validateEdge(node.Edge0); err != nil {
		return DecoEdge{}, err
	}
	if err := graph.validateEdge(node.Edge1); err != nil {
		return DecoEdge{}, err
	}
	if node.Level >= 1<<14 || node.Latches0 >= 1<<5 || node.Latches1 >= 1<<5 || node.Latches2 >= 1<<5 {
		return DecoEdge{}, ErrDimension
	}
	if len(graph.nodes) >= graph.limits.MaxNodes || len(graph.nodes) >= decoEdgeNodeLimit {
		return DecoEdge{}, ErrLimit
	}
	if graph.stopped() {
		return DecoEdge{}, ErrCancelled
	}
	if len(graph.nodes) == cap(graph.nodes) {
		capacity := cap(graph.nodes)
		if capacity == 0 {
			capacity = 1
		} else if capacity > graph.limits.MaxNodes/2 {
			capacity = graph.limits.MaxNodes
		} else {
			capacity *= 2
		}
		if capacity <= len(graph.nodes) || capacity > graph.limits.MaxNodes {
			return DecoEdge{}, ErrLimit
		}
		nodes := make([]DecoNode, len(graph.nodes), capacity)
		copy(nodes, graph.nodes)
		graph.nodes = nodes
	}
	index := len(graph.nodes)
	graph.nodes = append(graph.nodes, node)
	return DecoEdge{Node: index}, nil
}

func (graph *DecoGraph) AddAND(first, second DecoEdge) (DecoEdge, error) {
	node := DecoNode{Edge0: first, Edge1: second, OriginalComplement0: first.Complement, OriginalComplement1: second.Complement}
	return graph.appendNode(node)
}

func (graph *DecoGraph) AddOR(first, second DecoEdge) (DecoEdge, error) {
	node := DecoNode{
		Edge0: first.Not(), Edge1: second.Not(), NodeOr: true,
		OriginalComplement0: first.Complement, OriginalComplement1: second.Complement,
	}
	edge, err := graph.appendNode(node)
	if err != nil {
		return DecoEdge{}, err
	}
	edge.Complement = true
	return edge, nil
}

func (graph *DecoGraph) AddXOR(first, second DecoEdge, alternate bool) (DecoEdge, error) {
	// Build on a clone and commit only after all three source nodes succeed.
	work, err := graph.Clone()
	if err != nil {
		return DecoEdge{}, err
	}
	var left, right DecoEdge
	if !alternate {
		left, err = work.AddAND(first.Not(), second)
		if err == nil {
			right, err = work.AddAND(first, second.Not())
		}
	} else {
		left, err = work.AddAND(first, second)
		if err == nil {
			right, err = work.AddAND(first.Not(), second.Not())
		}
	}
	if err != nil {
		return DecoEdge{}, err
	}
	result, err := work.AddOR(left, right)
	if err != nil {
		return DecoEdge{}, err
	}
	if alternate {
		result = result.Not()
	}
	graph.nodes = work.nodes
	return result, nil
}

func (graph *DecoGraph) AddMUX(control, whenTrue, whenFalse DecoEdge, alternate bool) (DecoEdge, error) {
	work, err := graph.Clone()
	if err != nil {
		return DecoEdge{}, err
	}
	if alternate {
		whenTrue = whenTrue.Not()
		whenFalse = whenFalse.Not()
	}
	left, err := work.AddAND(control, whenTrue)
	if err != nil {
		return DecoEdge{}, err
	}
	right, err := work.AddAND(control.Not(), whenFalse)
	if err != nil {
		return DecoEdge{}, err
	}
	result, err := work.AddOR(left, right)
	if err != nil {
		return DecoEdge{}, err
	}
	if alternate {
		result = result.Not()
	}
	graph.nodes = work.nodes
	return result, nil
}

func (graph *DecoGraph) Clone() (*DecoGraph, error) {
	if err := graph.check(); err != nil {
		return nil, err
	}
	if graph.stopped() {
		return nil, ErrCancelled
	}
	nodes := make([]DecoNode, len(graph.nodes), cap(graph.nodes))
	copy(nodes, graph.nodes)
	return &DecoGraph{constant: graph.constant, leaves: graph.leaves, nodes: nodes, root: graph.root, limits: graph.limits}, nil
}

// Evaluate interprets the source's AND/inverter representation. OR/XOR/MUX
// builders are already lowered into this representation.
func (graph *DecoGraph) Evaluate(inputs []bool) (bool, error) {
	if err := graph.check(); err != nil {
		return false, err
	}
	if graph.stopped() {
		return false, ErrCancelled
	}
	if graph.constant {
		return !graph.root.Complement, nil
	}
	if len(inputs) != graph.leaves || graph.root.Node < 0 || graph.root.Node >= len(graph.nodes) {
		return false, ErrDimension
	}
	values := make([]bool, len(graph.nodes))
	copy(values, inputs)
	read := func(edge DecoEdge) bool {
		value := values[edge.Node]
		if edge.Complement {
			return !value
		}
		return value
	}
	for index := graph.leaves; index < len(graph.nodes); index++ {
		if graph.stopped() {
			return false, ErrCancelled
		}
		node := graph.nodes[index]
		values[index] = read(node.Edge0) && read(node.Edge1)
	}
	return read(graph.root), nil
}

func (graph *DecoGraph) Close() {
	if graph == nil || graph.closed {
		return
	}
	graph.nodes = nil
	graph.leaves = 0
	graph.closed = true
}
