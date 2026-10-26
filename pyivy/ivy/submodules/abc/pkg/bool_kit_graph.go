package abc

// Concrete decomposition graph core from src/bool/kit/kitGraph.c.

import "sync"

type KitEdge struct {
	Node       int
	Complement bool
}

func (edge KitEdge) Not() KitEdge {
	edge.Complement = !edge.Complement
	return edge
}

type kitGraphNode struct {
	first, second                 KitEdge
	originalOr                    bool
	originalFirst, originalSecond bool
}

type KitGraphOptions struct {
	MaxNodes int
	MaxWork  int
	MaxBytes int
	Stop     func() bool
}

type KitGraph struct {
	mu       sync.RWMutex
	constant bool
	leaves   int
	maxNodes int
	maxBytes int
	nodes    []kitGraphNode
	root     KitEdge
}

func normalizeKitGraphOptions(options KitGraphOptions) (KitGraphOptions, error) {
	if options.MaxNodes < 0 || options.MaxWork < 0 || options.MaxBytes < 0 {
		return options, ErrDimension
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxWork == 0 {
		options.MaxWork = 1 << 20
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	return options, nil
}

func NewKitGraph(leaves int, options KitGraphOptions) (*KitGraph, error) {
	options, err := normalizeKitGraphOptions(options)
	if err != nil {
		return nil, err
	}
	if leaves < 0 || leaves > options.MaxNodes || leaves > options.MaxBytes/48 {
		return nil, ErrDimension
	}
	return &KitGraph{leaves: leaves, maxNodes: options.MaxNodes, maxBytes: options.MaxBytes, nodes: make([]kitGraphNode, leaves)}, nil
}

func NewKitConstantGraph(value bool) *KitGraph {
	return &KitGraph{constant: true, maxNodes: 1, maxBytes: 48, root: KitEdge{Complement: !value}}
}

func NewKitLeafGraph(leaf, leaves int, complement bool, options KitGraphOptions) (*KitGraph, error) {
	if leaf < 0 || leaf >= leaves {
		return nil, ErrDimension
	}
	graph, err := NewKitGraph(leaves, options)
	if err != nil {
		return nil, err
	}
	graph.root = KitEdge{Node: leaf, Complement: complement}
	return graph, nil
}

func (graph *KitGraph) validEdgeLocked(edge KitEdge) bool {
	return graph != nil && !graph.constant && edge.Node >= 0 && edge.Node < len(graph.nodes)
}

func (graph *KitGraph) Leaf(leaf int, complement bool) (KitEdge, error) {
	if graph == nil {
		return KitEdge{}, ErrDimension
	}
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	if graph.constant || leaf < 0 || leaf >= graph.leaves {
		return KitEdge{}, ErrDimension
	}
	return KitEdge{Node: leaf, Complement: complement}, nil
}

func (graph *KitGraph) SetRoot(root KitEdge) error {
	if graph == nil {
		return ErrDimension
	}
	graph.mu.Lock()
	defer graph.mu.Unlock()
	if !graph.validEdgeLocked(root) {
		return ErrInvalidLiteral
	}
	graph.root = root
	return nil
}

func (graph *KitGraph) preflightLocked(add int, edges ...KitEdge) error {
	if graph.constant || add < 0 || len(graph.nodes) > graph.maxNodes-add || len(graph.nodes) > graph.maxBytes/48-add {
		return ErrLimit
	}
	for _, edge := range edges {
		if !graph.validEdgeLocked(edge) {
			return ErrInvalidLiteral
		}
	}
	return nil
}

func (graph *KitGraph) addAndLocked(first, second KitEdge) KitEdge {
	node := kitGraphNode{first: first, second: second, originalFirst: first.Complement, originalSecond: second.Complement}
	graph.nodes = append(graph.nodes, node)
	return KitEdge{Node: len(graph.nodes) - 1}
}

func (graph *KitGraph) addOrLocked(first, second KitEdge) KitEdge {
	node := kitGraphNode{first: first.Not(), second: second.Not(), originalOr: true, originalFirst: first.Complement, originalSecond: second.Complement}
	graph.nodes = append(graph.nodes, node)
	return KitEdge{Node: len(graph.nodes) - 1, Complement: true}
}

func (graph *KitGraph) AddAnd(first, second KitEdge) (KitEdge, error) {
	if graph == nil {
		return KitEdge{}, ErrDimension
	}
	graph.mu.Lock()
	defer graph.mu.Unlock()
	if err := graph.preflightLocked(1, first, second); err != nil {
		return KitEdge{}, err
	}
	return graph.addAndLocked(first, second), nil
}

func (graph *KitGraph) AddOr(first, second KitEdge) (KitEdge, error) {
	if graph == nil {
		return KitEdge{}, ErrDimension
	}
	graph.mu.Lock()
	defer graph.mu.Unlock()
	if err := graph.preflightLocked(1, first, second); err != nil {
		return KitEdge{}, err
	}
	return graph.addOrLocked(first, second), nil
}

func (graph *KitGraph) AddXor(first, second KitEdge, form int) (KitEdge, error) {
	if graph == nil || (form != 0 && form != 1) {
		return KitEdge{}, ErrDimension
	}
	graph.mu.Lock()
	defer graph.mu.Unlock()
	if err := graph.preflightLocked(3, first, second); err != nil {
		return KitEdge{}, err
	}
	if form == 0 {
		left := graph.addAndLocked(first.Not(), second)
		right := graph.addAndLocked(first, second.Not())
		return graph.addOrLocked(left, right), nil
	}
	left := graph.addAndLocked(first, second)
	right := graph.addAndLocked(first.Not(), second.Not())
	return graph.addOrLocked(left, right).Not(), nil
}

func (graph *KitGraph) AddMux(control, thenEdge, elseEdge KitEdge, form int) (KitEdge, error) {
	if graph == nil || (form != 0 && form != 1) {
		return KitEdge{}, ErrDimension
	}
	graph.mu.Lock()
	defer graph.mu.Unlock()
	if err := graph.preflightLocked(3, control, thenEdge, elseEdge); err != nil {
		return KitEdge{}, err
	}
	if form == 0 {
		left := graph.addAndLocked(control, thenEdge)
		right := graph.addAndLocked(control.Not(), elseEdge)
		return graph.addOrLocked(left, right), nil
	}
	left := graph.addAndLocked(control, thenEdge.Not())
	right := graph.addAndLocked(control.Not(), elseEdge.Not())
	return graph.addOrLocked(left, right).Not(), nil
}

func (graph *KitGraph) NodeCount() int {
	if graph == nil {
		return 0
	}
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	return len(graph.nodes)
}

type kitGraphWork struct {
	options KitGraphOptions
	used    int
}

func (work *kitGraphWork) take() error {
	if work.options.Stop != nil && work.options.Stop() {
		return ErrCancelled
	}
	if work.used >= work.options.MaxWork {
		return ErrLimit
	}
	work.used++
	return nil
}

func (graph *KitGraph) truth32Locked(options KitGraphOptions) (uint32, error) {
	if graph.constant {
		if graph.root.Complement {
			return 0, nil
		}
		return ^uint32(0), nil
	}
	if graph.leaves < 0 || graph.leaves > 5 || len(graph.nodes) < graph.leaves || len(graph.nodes) > graph.maxNodes || !graph.validEdgeLocked(graph.root) {
		return 0, ErrInvariant
	}
	if len(graph.nodes) > options.MaxNodes || len(graph.nodes) > options.MaxBytes/4 {
		return 0, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return 0, ErrCancelled
	}
	patterns := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	values := make([]uint32, len(graph.nodes))
	copy(values, patterns[:graph.leaves])
	work := kitGraphWork{options: options}
	for index := graph.leaves; index < len(graph.nodes); index++ {
		if err := work.take(); err != nil {
			return 0, err
		}
		node := graph.nodes[index]
		if node.first.Node < 0 || node.first.Node >= index || node.second.Node < 0 || node.second.Node >= index {
			return 0, ErrInvariant
		}
		first, second := values[node.first.Node], values[node.second.Node]
		if node.first.Complement {
			first = ^first
		}
		if node.second.Complement {
			second = ^second
		}
		values[index] = first & second
	}
	result := values[graph.root.Node]
	if graph.root.Complement {
		result = ^result
	}
	return result, nil
}

func (graph *KitGraph) Truth32(options KitGraphOptions) (uint32, error) {
	if graph == nil {
		return 0, ErrDimension
	}
	options, err := normalizeKitGraphOptions(options)
	if err != nil {
		return 0, err
	}
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	return graph.truth32Locked(options)
}

func (graph *KitGraph) levelLocked(node int, memo []int, seen []bool, work *kitGraphWork) (int, error) {
	if err := work.take(); err != nil {
		return 0, err
	}
	if node < graph.leaves {
		return 0, nil
	}
	if seen[node] {
		return memo[node], nil
	}
	entry := graph.nodes[node]
	first, err := graph.levelLocked(entry.first.Node, memo, seen, work)
	if err != nil {
		return 0, err
	}
	second, err := graph.levelLocked(entry.second.Node, memo, seen, work)
	if err != nil {
		return 0, err
	}
	if second > first {
		first = second
	}
	memo[node], seen[node] = first+1, true
	return memo[node], nil
}

func (graph *KitGraph) Level(options KitGraphOptions) (int, error) {
	if graph == nil {
		return 0, ErrDimension
	}
	options, err := normalizeKitGraphOptions(options)
	if err != nil {
		return 0, err
	}
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	if graph.constant {
		return 0, nil
	}
	if !graph.validEdgeLocked(graph.root) {
		return 0, ErrInvalidLiteral
	}
	if len(graph.nodes) > options.MaxNodes || len(graph.nodes) > options.MaxBytes/16 {
		return 0, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return 0, ErrCancelled
	}
	return graph.levelLocked(graph.root.Node, make([]int, len(graph.nodes)), make([]bool, len(graph.nodes)), &kitGraphWork{options: options})
}

func (graph *KitGraph) leafDepthLocked(node, leaf int, memo []int, seen []bool, work *kitGraphWork) (int, error) {
	if err := work.take(); err != nil {
		return 0, err
	}
	if node < graph.leaves {
		if node == leaf {
			return 0, nil
		}
		return -100, nil
	}
	if seen[node] {
		return memo[node], nil
	}
	entry := graph.nodes[node]
	first, err := graph.leafDepthLocked(entry.first.Node, leaf, memo, seen, work)
	if err != nil {
		return 0, err
	}
	second, err := graph.leafDepthLocked(entry.second.Node, leaf, memo, seen, work)
	if err != nil {
		return 0, err
	}
	if second > first {
		first = second
	}
	if first != -100 {
		first++
	}
	memo[node], seen[node] = first, true
	return first, nil
}

func (graph *KitGraph) LeafDepth(leaf int, options KitGraphOptions) (int, error) {
	if graph == nil {
		return 0, ErrDimension
	}
	options, err := normalizeKitGraphOptions(options)
	if err != nil {
		return 0, err
	}
	graph.mu.RLock()
	defer graph.mu.RUnlock()
	if graph.constant || leaf < 0 || leaf >= graph.leaves || !graph.validEdgeLocked(graph.root) {
		return 0, ErrDimension
	}
	if len(graph.nodes) > options.MaxNodes || len(graph.nodes) > options.MaxBytes/16 {
		return 0, ErrLimit
	}
	if options.Stop != nil && options.Stop() {
		return 0, ErrCancelled
	}
	return graph.leafDepthLocked(graph.root.Node, leaf, make([]int, len(graph.nodes)), make([]bool, len(graph.nodes)), &kitGraphWork{options: options})
}
