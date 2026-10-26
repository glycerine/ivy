package abc

// Stand-alone models for selected invariants from src/opt/fret. These types do
// not port that package's MinReg manager, augmenting-path DFS, cut rewrite, or
// main loop. The concrete AIG operation below instead exposes the separately
// ported forward-frontier retimer. General ABC-network, mapped-gate,
// backward-initialization, and timing-manager modes are rejected rather than
// approximated.

import "slices"

const FlowRetimeMaximumDistance = 30000

type FlowRetimeMark uint16

const (
	FlowRetimeVisitedEdge FlowRetimeMark = 1 << iota
	FlowRetimeVisitedReverse
	FlowRetimeFlow
	FlowRetimeCrossBoundary
	FlowRetimeBlock
	FlowRetimeInitialZero
	FlowRetimeInitialOne
	FlowRetimeConservative
	FlowRetimeBiasNode
)

const (
	FlowRetimeVisited        = FlowRetimeVisitedEdge | FlowRetimeVisitedReverse
	FlowRetimeInitialCare    = FlowRetimeInitialZero | FlowRetimeInitialOne
	FlowRetimeBlockOrConserv = FlowRetimeBlock | FlowRetimeConservative
)

type FlowRetimeDirection uint8

const (
	FlowRetimeForward FlowRetimeDirection = iota
	FlowRetimeBackward
	FlowRetimeBoth
)

type FlowRetimeAIGOptions struct {
	Direction             FlowRetimeDirection
	MaxMoves              int
	MaxObjects            int
	MaxDelay              int
	ComputeInitialState   bool
	GuaranteeInitialState bool
	BlockConstants        bool
	ConservativeTiming    bool
	Stop                  func() bool
}

type FlowRetimeAIGResult struct {
	Graph *AigManager
	Moves int
}

// FlowRetimeAIG exposes the compatible forward AIG subset through the
// independently validated frontier retimer. This is not the FRET manager's
// distance-histogram flow algorithm. AigManager has no mapped/SOP nodes or
// latch-initialization metadata, so modes requiring those data return
// ErrUnsupported atomically.
func FlowRetimeAIG(source *AigManager, options FlowRetimeAIGOptions) (FlowRetimeAIGResult, error) {
	if source == nil || options.Direction > FlowRetimeBoth || options.MaxMoves < 0 || options.MaxObjects < 0 || options.MaxDelay < 0 {
		return FlowRetimeAIGResult{}, ErrDimension
	}
	if options.Direction != FlowRetimeForward || options.ComputeInitialState || options.GuaranteeInitialState || options.MaxDelay != 0 || options.BlockConstants || options.ConservativeTiming {
		return FlowRetimeAIGResult{}, ErrUnsupported
	}
	result, err := source.RetimeFrontier(AigRetimeOptions{MaxSteps: options.MaxMoves, MaxObjects: options.MaxObjects, Stop: options.Stop})
	if err != nil {
		return FlowRetimeAIGResult{}, err
	}
	return FlowRetimeAIGResult{Graph: result.Graph, Moves: result.Steps}, nil
}

type FlowRetimeNodeLag struct {
	Node int
	Lag  int
}

// FlowRetimeEdge is an edge-register form of the source lag equation. Delay is
// charged when this edge is combinational (Registers == 0).
type FlowRetimeEdge struct {
	From, To  int
	Registers int
	Delay     int
}

type FlowRetimeLagLimits struct {
	MaxNodes, MaxEdges, MaxRegisters int
	Stop                             func() bool
}

type FlowRetimeLagNetwork struct {
	nodes  int
	edges  []FlowRetimeEdge
	limits FlowRetimeLagLimits
}

func normalizeFlowRetimeLagLimits(limits FlowRetimeLagLimits) (FlowRetimeLagLimits, error) {
	if limits.MaxNodes < 0 || limits.MaxEdges < 0 || limits.MaxRegisters < 0 {
		return limits, ErrDimension
	}
	if limits.MaxNodes == 0 {
		limits.MaxNodes = 1 << 20
	}
	if limits.MaxEdges == 0 {
		limits.MaxEdges = 1 << 22
	}
	if limits.MaxRegisters == 0 {
		limits.MaxRegisters = 1 << 30
	}
	return limits, nil
}

func NewFlowRetimeLagNetwork(nodes int, edges []FlowRetimeEdge, limits FlowRetimeLagLimits) (*FlowRetimeLagNetwork, error) {
	var err error
	limits, err = normalizeFlowRetimeLagLimits(limits)
	if err != nil || nodes < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if nodes > limits.MaxNodes || len(edges) > limits.MaxEdges {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	network := &FlowRetimeLagNetwork{nodes: nodes, edges: slices.Clone(edges), limits: limits}
	if err = network.check(); err != nil {
		return nil, err
	}
	return network, nil
}

func (network *FlowRetimeLagNetwork) check() error {
	if network == nil || network.nodes < 0 || network.nodes > network.limits.MaxNodes || len(network.edges) > network.limits.MaxEdges {
		return ErrInvariant
	}
	total := 0
	for index, edge := range network.edges {
		if network.limits.Stop != nil && index&255 == 0 && network.limits.Stop() {
			return ErrCancelled
		}
		if edge.From < 0 || edge.From >= network.nodes || edge.To < 0 || edge.To >= network.nodes || edge.Registers < 0 || edge.Delay < 0 || edge.Delay > FlowRetimeMaximumDistance {
			return ErrDimension
		}
		if total > network.limits.MaxRegisters-edge.Registers {
			return ErrLimit
		}
		total += edge.Registers
	}
	return nil
}

func (network *FlowRetimeLagNetwork) Nodes() int {
	if network == nil {
		return 0
	}
	return network.nodes
}

func (network *FlowRetimeLagNetwork) Edges() []FlowRetimeEdge {
	if network == nil {
		return nil
	}
	return slices.Clone(network.edges)
}

func (network *FlowRetimeLagNetwork) RegisterCount() (int, error) {
	if err := network.check(); err != nil {
		return 0, err
	}
	total := 0
	for _, edge := range network.edges {
		total += edge.Registers
	}
	return total, nil
}

// Retimed applies the standard retiming equation
// w'(u,v)=w(u,v)+r(v)-r(u). FRET stores and updates node lags but does not use
// this edge-network representation. Negative edge register counts reject the
// whole operation.
func (network *FlowRetimeLagNetwork) Retimed(lags []int) (*FlowRetimeLagNetwork, error) {
	if err := network.check(); err != nil {
		return nil, err
	}
	if len(lags) != network.nodes {
		return nil, ErrDimension
	}
	edges := make([]FlowRetimeEdge, len(network.edges))
	total := 0
	for index, edge := range network.edges {
		if network.limits.Stop != nil && index&255 == 0 && network.limits.Stop() {
			return nil, ErrCancelled
		}
		to, from := lags[edge.To], lags[edge.From]
		maximum := int(^uint(0) >> 1)
		minimum := -maximum - 1
		if from > 0 && to < minimum+from || from < 0 && to > maximum+from {
			return nil, ErrLimit
		}
		delta := to - from
		if delta < -edge.Registers || delta > network.limits.MaxRegisters-edge.Registers {
			return nil, ErrDimension
		}
		edge.Registers += delta
		if total > network.limits.MaxRegisters-edge.Registers {
			return nil, ErrLimit
		}
		total += edge.Registers
		edges[index] = edge
	}
	return &FlowRetimeLagNetwork{nodes: network.nodes, edges: edges, limits: network.limits}, nil
}

// PathRegisters returns the register count along an explicit edge-index path.
func (network *FlowRetimeLagNetwork) PathRegisters(path []int) (int, error) {
	if err := network.check(); err != nil {
		return 0, err
	}
	if len(path) == 0 {
		return 0, nil
	}
	total := 0
	previousTo := -1
	for offset, edgeIndex := range path {
		if network.limits.Stop != nil && offset&255 == 0 && network.limits.Stop() {
			return 0, ErrCancelled
		}
		if edgeIndex < 0 || edgeIndex >= len(network.edges) {
			return 0, ErrDimension
		}
		edge := network.edges[edgeIndex]
		if offset != 0 && edge.From != previousTo {
			return 0, ErrInvariant
		}
		if total > network.limits.MaxRegisters-edge.Registers {
			return 0, ErrLimit
		}
		total += edge.Registers
		previousTo = edge.To
	}
	return total, nil
}

// MaximumCombinationalDelay computes the longest delay in the zero-register
// edge DAG. A zero-register cycle is rejected instead of recursing forever.
func (network *FlowRetimeLagNetwork) MaximumCombinationalDelay() (int, error) {
	if err := network.check(); err != nil {
		return 0, err
	}
	indegree := make([]int, network.nodes)
	out := make([][]int, network.nodes)
	for index, edge := range network.edges {
		if edge.Registers == 0 {
			indegree[edge.To]++
			out[edge.From] = append(out[edge.From], index)
		}
	}
	queue := make([]int, 0, network.nodes)
	for node, degree := range indegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}
	distance := make([]int, network.nodes)
	seen, maximum := 0, 0
	for len(queue) != 0 {
		if network.limits.Stop != nil && seen&255 == 0 && network.limits.Stop() {
			return 0, ErrCancelled
		}
		node := queue[0]
		queue = queue[1:]
		seen++
		for _, edgeIndex := range out[node] {
			edge := network.edges[edgeIndex]
			if distance[node] > FlowRetimeMaximumDistance-edge.Delay {
				return 0, ErrLimit
			}
			candidate := distance[node] + edge.Delay
			if candidate > distance[edge.To] {
				distance[edge.To] = candidate
			}
			if candidate > maximum {
				maximum = candidate
			}
			indegree[edge.To]--
			if indegree[edge.To] == 0 {
				queue = append(queue, edge.To)
			}
		}
	}
	if seen != network.nodes {
		return 0, ErrInvariant
	}
	return maximum, nil
}

type FlowRetimeArc struct {
	From, To int
}

type FlowRetimeCutOptions struct {
	MaxNodes, MaxArcs, MaxFlow int
	Stop                       func() bool
}

type FlowRetimeCutResult struct {
	Flow int
	Cut  []int
}

type flowRetimeResidualEdge struct {
	to, reverse, capacity int
}

// MinimumFlowRetimeCut computes a deterministic minimum vertex cut on an owned
// directed topology. It is a stand-alone model of the cut problem, not a port
// of FRET's distance-histogram and plain augmenting-path routines. Blocked
// vertices cannot carry flow. Source and sink vertices have infinite split
// capacity unless blocked.
func MinimumFlowRetimeCut(nodes int, arcs []FlowRetimeArc, sources, sinks []int, blocked []bool, options FlowRetimeCutOptions) (FlowRetimeCutResult, error) {
	if nodes < 0 || options.MaxNodes < 0 || options.MaxArcs < 0 || options.MaxFlow < 0 || len(blocked) != 0 && len(blocked) != nodes {
		return FlowRetimeCutResult{}, ErrDimension
	}
	if options.MaxNodes == 0 {
		options.MaxNodes = 1 << 20
	}
	if options.MaxArcs == 0 {
		options.MaxArcs = 1 << 22
	}
	if options.MaxFlow == 0 {
		options.MaxFlow = 1 << 20
	}
	if nodes > options.MaxNodes || len(arcs) > options.MaxArcs {
		return FlowRetimeCutResult{}, ErrLimit
	}
	isSource, isSink := make([]bool, nodes), make([]bool, nodes)
	for _, node := range sources {
		if node < 0 || node >= nodes {
			return FlowRetimeCutResult{}, ErrDimension
		}
		isSource[node] = true
	}
	for _, node := range sinks {
		if node < 0 || node >= nodes {
			return FlowRetimeCutResult{}, ErrDimension
		}
		isSink[node] = true
	}
	const infinity = int(^uint(0)>>1) / 4
	sourceID, sinkID := 2*nodes, 2*nodes+1
	graph := make([][]flowRetimeResidualEdge, 2*nodes+2)
	add := func(from, to, capacity int) {
		forward := flowRetimeResidualEdge{to: to, reverse: len(graph[to]), capacity: capacity}
		reverse := flowRetimeResidualEdge{to: from, reverse: len(graph[from]), capacity: 0}
		graph[from] = append(graph[from], forward)
		graph[to] = append(graph[to], reverse)
	}
	for node := 0; node < nodes; node++ {
		capacity := 1
		if isSource[node] || isSink[node] {
			capacity = infinity
		}
		if len(blocked) != 0 && blocked[node] {
			capacity = 0
		}
		add(2*node, 2*node+1, capacity)
		if isSource[node] {
			add(sourceID, 2*node, infinity)
		}
		if isSink[node] {
			add(2*node+1, sinkID, infinity)
		}
	}
	for index, arc := range arcs {
		if options.Stop != nil && index&255 == 0 && options.Stop() {
			return FlowRetimeCutResult{}, ErrCancelled
		}
		if arc.From < 0 || arc.From >= nodes || arc.To < 0 || arc.To >= nodes {
			return FlowRetimeCutResult{}, ErrDimension
		}
		add(2*arc.From+1, 2*arc.To, infinity)
	}
	flow := 0
	for flow < options.MaxFlow {
		if options.Stop != nil && options.Stop() {
			return FlowRetimeCutResult{}, ErrCancelled
		}
		level := make([]int, len(graph))
		for index := range level {
			level[index] = -1
		}
		level[sourceID] = 0
		queue := []int{sourceID}
		for len(queue) != 0 {
			from := queue[0]
			queue = queue[1:]
			for _, edge := range graph[from] {
				if edge.capacity > 0 && level[edge.to] < 0 {
					level[edge.to] = level[from] + 1
					queue = append(queue, edge.to)
				}
			}
		}
		if level[sinkID] < 0 {
			break
		}
		next := make([]int, len(graph))
		var send func(int, int) int
		send = func(from, amount int) int {
			if from == sinkID {
				return amount
			}
			for next[from] < len(graph[from]) {
				index := next[from]
				edge := &graph[from][index]
				if edge.capacity > 0 && level[edge.to] == level[from]+1 {
					pushed := send(edge.to, min(amount, edge.capacity))
					if pushed != 0 {
						edge.capacity -= pushed
						graph[edge.to][edge.reverse].capacity += pushed
						return pushed
					}
				}
				next[from]++
			}
			return 0
		}
		for flow < options.MaxFlow {
			pushed := send(sourceID, options.MaxFlow-flow)
			if pushed == 0 {
				break
			}
			flow += pushed
		}
	}
	reachable := make([]bool, len(graph))
	reachable[sourceID] = true
	queue := []int{sourceID}
	for len(queue) != 0 {
		from := queue[0]
		queue = queue[1:]
		for _, edge := range graph[from] {
			if edge.capacity > 0 && !reachable[edge.to] {
				reachable[edge.to] = true
				queue = append(queue, edge.to)
			}
		}
	}
	if flow == options.MaxFlow && reachable[sinkID] {
		return FlowRetimeCutResult{}, ErrLimit
	}
	cut := make([]int, 0, flow)
	for node := 0; node < nodes; node++ {
		if !isSource[node] && !isSink[node] && (len(blocked) == 0 || !blocked[node]) && reachable[2*node] && !reachable[2*node+1] {
			cut = append(cut, node)
		}
	}
	return FlowRetimeCutResult{Flow: flow, Cut: cut}, nil
}
