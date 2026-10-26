package abc

// Sequential retiming algorithms from src/opt/ret. Registers are represented
// explicitly on directed edges, in source-to-sink order. This representation
// makes the local forward/backward transformations and their initial-state
// correspondence concrete without borrowing mutable ABC objects.

import "slices"

type RetimingDirection uint8

const (
	RetimingForward RetimingDirection = iota
	RetimingBackward
)

type RetimingNode struct {
	PrimaryInput bool
	Delay        int
	// Truth is indexed by the input-edge order; bit i is input i.
	Truth []bool
}

type RetimingEdge struct {
	From, To int
	// Initial is ordered from the source toward the sink.
	Initial []bool
}

type RetimingLimits struct {
	MaxNodes, MaxEdges, MaxRegisters, MaxMoves, MaxIterations, MaxInitialAssignments int
	Stop                                                                             func() bool
}

type RetimingNetwork struct {
	nodes   []RetimingNode
	edges   []RetimingEdge
	outputs []int
	limits  RetimingLimits
}

func normalizeRetimingLimits(limits RetimingLimits) (RetimingLimits, error) {
	if limits.MaxNodes < 0 || limits.MaxEdges < 0 || limits.MaxRegisters < 0 || limits.MaxMoves < 0 || limits.MaxIterations < 0 || limits.MaxInitialAssignments < 0 {
		return limits, ErrDimension
	}
	if limits.MaxNodes == 0 {
		limits.MaxNodes = 1 << 20
	}
	if limits.MaxEdges == 0 {
		limits.MaxEdges = 1 << 22
	}
	if limits.MaxRegisters == 0 {
		limits.MaxRegisters = 1 << 28
	}
	if limits.MaxMoves == 0 {
		limits.MaxMoves = 10000
	}
	if limits.MaxIterations == 0 {
		limits.MaxIterations = 1000
	}
	if limits.MaxInitialAssignments == 0 {
		limits.MaxInitialAssignments = 1 << 20
	}
	return limits, nil
}

func NewRetimingNetwork(nodes []RetimingNode, edges []RetimingEdge, outputs []int, limits RetimingLimits) (*RetimingNetwork, error) {
	var err error
	limits, err = normalizeRetimingLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(nodes) > limits.MaxNodes || len(edges) > limits.MaxEdges {
		return nil, ErrLimit
	}
	network := &RetimingNetwork{nodes: make([]RetimingNode, len(nodes)), edges: make([]RetimingEdge, len(edges)), outputs: slices.Clone(outputs), limits: limits}
	for i, node := range nodes {
		network.nodes[i] = node
		network.nodes[i].Truth = slices.Clone(node.Truth)
	}
	for i, edge := range edges {
		network.edges[i] = edge
		network.edges[i].Initial = slices.Clone(edge.Initial)
	}
	if err = network.check(); err != nil {
		return nil, err
	}
	return network, nil
}

func (network *RetimingNetwork) check() error {
	if network == nil || len(network.nodes) > network.limits.MaxNodes || len(network.edges) > network.limits.MaxEdges {
		return ErrInvariant
	}
	inputs := make([]int, len(network.nodes))
	registers := 0
	for i, node := range network.nodes {
		if node.Delay < 0 {
			return ErrDimension
		}
		if network.limits.Stop != nil && i&255 == 0 && network.limits.Stop() {
			return ErrCancelled
		}
	}
	for _, edge := range network.edges {
		if edge.From < 0 || edge.From >= len(network.nodes) || edge.To < 0 || edge.To >= len(network.nodes) {
			return ErrDimension
		}
		inputs[edge.To]++
		if registers > network.limits.MaxRegisters-len(edge.Initial) {
			return ErrLimit
		}
		registers += len(edge.Initial)
	}
	for i, node := range network.nodes {
		if node.PrimaryInput {
			if inputs[i] != 0 || len(node.Truth) != 0 {
				return ErrInvariant
			}
			continue
		}
		if inputs[i] >= 63 || len(node.Truth) != 1<<uint(inputs[i]) {
			return ErrDimension
		}
	}
	for _, output := range network.outputs {
		if output < 0 || output >= len(network.nodes) {
			return ErrDimension
		}
	}
	_, err := network.combinationalOrder()
	return err
}

func (network *RetimingNetwork) clone() *RetimingNetwork {
	copy := &RetimingNetwork{nodes: make([]RetimingNode, len(network.nodes)), edges: make([]RetimingEdge, len(network.edges)), outputs: slices.Clone(network.outputs), limits: network.limits}
	for i, node := range network.nodes {
		copy.nodes[i] = node
		copy.nodes[i].Truth = slices.Clone(node.Truth)
	}
	for i, edge := range network.edges {
		copy.edges[i] = edge
		copy.edges[i].Initial = slices.Clone(edge.Initial)
	}
	return copy
}

func (network *RetimingNetwork) Nodes() []RetimingNode {
	if network == nil {
		return nil
	}
	result := make([]RetimingNode, len(network.nodes))
	for i, node := range network.nodes {
		result[i] = node
		result[i].Truth = slices.Clone(node.Truth)
	}
	return result
}

func (network *RetimingNetwork) Edges() []RetimingEdge {
	if network == nil {
		return nil
	}
	result := make([]RetimingEdge, len(network.edges))
	for i, edge := range network.edges {
		result[i] = edge
		result[i].Initial = slices.Clone(edge.Initial)
	}
	return result
}

func (network *RetimingNetwork) Outputs() []int {
	if network == nil {
		return nil
	}
	return slices.Clone(network.outputs)
}

func (network *RetimingNetwork) RegisterCount() int {
	if network == nil {
		return 0
	}
	total := 0
	for _, edge := range network.edges {
		total += len(edge.Initial)
	}
	return total
}

func (network *RetimingNetwork) fanins(node int) []int {
	result := make([]int, 0)
	for i, edge := range network.edges {
		if edge.To == node {
			result = append(result, i)
		}
	}
	return result
}

func (network *RetimingNetwork) fanouts(node int) []int {
	result := make([]int, 0)
	for i, edge := range network.edges {
		if edge.From == node {
			result = append(result, i)
		}
	}
	return result
}

func (network *RetimingNetwork) combinationalOrder() ([]int, error) {
	indegree := make([]int, len(network.nodes))
	out := make([][]int, len(network.nodes))
	for _, edge := range network.edges {
		if len(edge.Initial) == 0 {
			indegree[edge.To]++
			out[edge.From] = append(out[edge.From], edge.To)
		}
	}
	queue := make([]int, 0, len(network.nodes))
	for node, degree := range indegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}
	order := make([]int, 0, len(network.nodes))
	for len(queue) != 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)
		for _, next := range out[node] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if len(order) != len(network.nodes) {
		return nil, ErrInvariant
	}
	return order, nil
}

func evalRetimingTruth(truth []bool, values []bool) bool {
	index := 0
	for i, value := range values {
		if value {
			index |= 1 << uint(i)
		}
	}
	return truth[index]
}

// Step evaluates one frame and returns an independently owned next state.
func (network *RetimingNetwork) Step(primary []bool) ([]bool, *RetimingNetwork, error) {
	if err := network.check(); err != nil {
		return nil, nil, err
	}
	order, _ := network.combinationalOrder()
	values := make([]bool, len(network.nodes))
	pi := 0
	for _, node := range order {
		if network.nodes[node].PrimaryInput {
			if pi >= len(primary) {
				return nil, nil, ErrDimension
			}
			values[node] = primary[pi]
			pi++
			continue
		}
		args := make([]bool, 0)
		for _, edgeIndex := range network.fanins(node) {
			edge := network.edges[edgeIndex]
			if len(edge.Initial) == 0 {
				args = append(args, values[edge.From])
			} else {
				args = append(args, edge.Initial[len(edge.Initial)-1])
			}
		}
		values[node] = evalRetimingTruth(network.nodes[node].Truth, args)
	}
	if pi != len(primary) {
		return nil, nil, ErrDimension
	}
	outputs := make([]bool, len(network.outputs))
	for i, node := range network.outputs {
		outputs[i] = values[node]
	}
	next := network.clone()
	for i, edge := range network.edges {
		if len(edge.Initial) == 0 {
			continue
		}
		copy(next.edges[i].Initial[1:], edge.Initial[:len(edge.Initial)-1])
		next.edges[i].Initial[0] = values[edge.From]
	}
	return outputs, next, nil
}

func (network *RetimingNetwork) NodeEnabled(node int, direction RetimingDirection) (bool, error) {
	if err := network.check(); err != nil {
		return false, err
	}
	if node < 0 || node >= len(network.nodes) || network.nodes[node].PrimaryInput || direction > RetimingBackward {
		return false, ErrDimension
	}
	ins, outs := network.fanins(node), network.fanouts(node)
	if len(outs) == 0 || direction == RetimingForward && len(ins) == 0 {
		return false, nil
	}
	edges := ins
	if direction == RetimingBackward {
		edges = outs
	}
	for _, edge := range edges {
		if len(network.edges[edge].Initial) == 0 {
			return false, nil
		}
	}
	return true, nil
}

// Move performs the exact local transformations in retIncrem.c. Backward
// initialization is accepted only when all removed output registers agree and
// a concrete local preimage exists; ambiguous preimages use lexical order.
func (network *RetimingNetwork) Move(node int, direction RetimingDirection) (*RetimingNetwork, error) {
	enabled, err := network.NodeEnabled(node, direction)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, ErrInvariant
	}
	result := network.clone()
	ins, outs := result.fanins(node), result.fanouts(node)
	if direction == RetimingForward {
		args := make([]bool, len(ins))
		for i, edgeIndex := range ins {
			values := result.edges[edgeIndex].Initial
			args[i] = values[len(values)-1]
			result.edges[edgeIndex].Initial = slices.Clone(values[:len(values)-1])
		}
		initial := evalRetimingTruth(result.nodes[node].Truth, args)
		for _, edgeIndex := range outs {
			result.edges[edgeIndex].Initial = append([]bool{initial}, result.edges[edgeIndex].Initial...)
		}
	} else {
		target := result.edges[outs[0]].Initial[0]
		for _, edgeIndex := range outs {
			values := result.edges[edgeIndex].Initial
			if values[0] != target {
				return nil, ErrUnsupported
			}
		}
		if len(ins) >= 63 {
			return nil, ErrLimit
		}
		assignment := -1
		args := make([]bool, len(ins))
		assignmentCount := 1 << uint(len(ins))
		if assignmentCount > result.limits.MaxInitialAssignments {
			return nil, ErrLimit
		}
		for candidate := 0; candidate < assignmentCount; candidate++ {
			if candidate&255 == 0 && result.limits.Stop != nil && result.limits.Stop() {
				return nil, ErrCancelled
			}
			for i := range args {
				args[i] = candidate&(1<<uint(i)) != 0
			}
			if evalRetimingTruth(result.nodes[node].Truth, args) == target {
				assignment = candidate
				break
			}
		}
		if assignment < 0 {
			return nil, ErrInvariant
		}
		for _, edgeIndex := range outs {
			result.edges[edgeIndex].Initial = slices.Clone(result.edges[edgeIndex].Initial[1:])
		}
		for i, edgeIndex := range ins {
			result.edges[edgeIndex].Initial = append(result.edges[edgeIndex].Initial, assignment&(1<<uint(i)) != 0)
		}
	}
	if result.RegisterCount() > result.limits.MaxRegisters {
		return nil, ErrLimit
	}
	if err = result.check(); err != nil {
		return nil, err
	}
	return result, nil
}

type RetimingRunResult struct {
	Network *RetimingNetwork
	Moves   int
}

type RetimingDelayResult struct {
	Network   *RetimingNetwork
	Moves     int
	Delay     int
	BestRound int
	LastRound int
}

func (network *RetimingNetwork) Incremental(direction RetimingDirection, oneStep bool) (RetimingRunResult, error) {
	if err := network.check(); err != nil || direction > RetimingBackward {
		if err != nil {
			return RetimingRunResult{}, err
		}
		return RetimingRunResult{}, ErrDimension
	}
	result, moves := network.clone(), 0
	for {
		changed := false
		for node := range result.nodes {
			if result.limits.Stop != nil && result.limits.Stop() {
				return RetimingRunResult{}, ErrCancelled
			}
			enabled, err := result.NodeEnabled(node, direction)
			if err == ErrDimension {
				continue
			}
			if err != nil {
				return RetimingRunResult{}, err
			}
			if !enabled {
				continue
			}
			if moves >= result.limits.MaxMoves {
				return RetimingRunResult{}, ErrLimit
			}
			result, err = result.Move(node, direction)
			if err != nil {
				return RetimingRunResult{}, err
			}
			moves++
			changed = true
			if oneStep {
				return RetimingRunResult{Network: result, Moves: moves}, nil
			}
		}
		if !changed {
			return RetimingRunResult{Network: result, Moves: moves}, nil
		}
	}
}

// MinDelay follows retDelay.c: analyze critical boundary nodes, move every
// enabled critical node in ID order, remember the best round, and stop after
// the configured iteration bound, the requested delay, or 20 stale rounds.
func (network *RetimingNetwork) MinDelay(direction RetimingDirection, delayLimit int, oneStep bool) (RetimingDelayResult, error) {
	if err := network.check(); err != nil || direction > RetimingBackward || delayLimit < 0 {
		if err != nil {
			return RetimingDelayResult{}, err
		}
		return RetimingDelayResult{}, ErrDimension
	}
	current := network.clone()
	best := current.clone()
	bestDelay := int(^uint(0) >> 1)
	bestRound, moves := 0, 0
	lastRound := 0
	iterationLimit := current.limits.MaxIterations
	if oneStep {
		iterationLimit = 1
	}
	for round := 0; round <= iterationLimit; round++ {
		if current.limits.Stop != nil && current.limits.Stop() {
			return RetimingDelayResult{}, ErrCancelled
		}
		delay, critical, err := current.Timing(direction)
		if err != nil {
			return RetimingDelayResult{}, err
		}
		lastRound = round
		if delay < bestDelay {
			bestDelay, bestRound, best = delay, round, current.clone()
		}
		if round == iterationLimit || round-bestRound > 20 || delayLimit > 0 && delay <= delayLimit {
			break
		}
		changed := false
		for _, node := range critical {
			enabled, enabledErr := current.NodeEnabled(node, direction)
			if enabledErr != nil || !enabled {
				continue
			}
			if moves >= current.limits.MaxMoves {
				return RetimingDelayResult{}, ErrLimit
			}
			current, err = current.Move(node, direction)
			if err != nil {
				return RetimingDelayResult{}, err
			}
			moves++
			changed = true
		}
		if !changed {
			break
		}
	}
	return RetimingDelayResult{Network: best, Moves: moves, Delay: bestDelay, BestRound: bestRound, LastRound: lastRound}, nil
}

// Timing performs the unit/general-delay combinational analysis used by
// retDelay.c and returns deterministic critical boundary nodes.
func (network *RetimingNetwork) Timing(direction RetimingDirection) (int, []int, error) {
	if err := network.check(); err != nil || direction > RetimingBackward {
		if err != nil {
			return 0, nil, err
		}
		return 0, nil, ErrDimension
	}
	order, _ := network.combinationalOrder()
	if direction == RetimingForward {
		slices.Reverse(order)
	}
	level := make([]int, len(network.nodes))
	maximum := 0
	maximumInt := int(^uint(0) >> 1)
	for _, node := range order {
		best := 0
		edges := network.fanouts(node)
		if direction == RetimingBackward {
			edges = network.fanins(node)
		}
		for _, edgeIndex := range edges {
			edge := network.edges[edgeIndex]
			if len(edge.Initial) != 0 {
				continue
			}
			next := edge.To
			if direction == RetimingBackward {
				next = edge.From
			}
			if level[next] > best {
				best = level[next]
			}
		}
		if best > maximumInt-network.nodes[node].Delay {
			return 0, nil, ErrLimit
		}
		level[node] = best + network.nodes[node].Delay
		if level[node] > maximum {
			maximum = level[node]
		}
	}
	critical := make([]int, 0)
	for node := range network.nodes {
		if network.nodes[node].PrimaryInput {
			continue
		}
		boundary := false
		edges := network.fanins(node)
		if direction == RetimingBackward {
			edges = network.fanouts(node)
		}
		for _, edge := range edges {
			candidate := network.edges[edge]
			if len(candidate.Initial) != 0 || direction == RetimingForward && network.nodes[candidate.From].PrimaryInput || direction == RetimingBackward && slices.Contains(network.outputs, candidate.To) {
				boundary = true
			}
		}
		if boundary && level[node] == maximum {
			critical = append(critical, node)
		}
	}
	return maximum, critical, nil
}

type RetimingCut struct {
	Flow  int
	Nodes []int
}

type retResidual struct {
	to, reverse, capacity int
}

// MaxFlowCut implements the directed node-capacitated max-flow used by
// retFlow.c. Register boundaries are flow sources; primary outputs (forward)
// or primary inputs (backward) are terminals. Registered edges are not
// traversed as combinational arcs.
func (network *RetimingNetwork) MaxFlowCut(direction RetimingDirection) (RetimingCut, error) {
	if err := network.check(); err != nil || direction > RetimingBackward {
		if err != nil {
			return RetimingCut{}, err
		}
		return RetimingCut{}, ErrDimension
	}
	n := len(network.nodes)
	sources, terminals := make([]bool, n), make([]bool, n)
	for _, edge := range network.edges {
		if len(edge.Initial) == 0 {
			continue
		}
		if direction == RetimingForward {
			sources[edge.To] = true
		} else {
			sources[edge.From] = true
		}
	}
	if direction == RetimingForward {
		for _, output := range network.outputs {
			terminals[output] = true
		}
	} else {
		for node := range network.nodes {
			if network.nodes[node].PrimaryInput {
				terminals[node] = true
			}
		}
	}
	const infinity = int(^uint(0)>>1) / 4
	sourceID, sinkID := 2*n, 2*n+1
	graph := make([][]retResidual, 2*n+2)
	add := func(from, to, capacity int) {
		forward := retResidual{to: to, reverse: len(graph[to]), capacity: capacity}
		reverse := retResidual{to: from, reverse: len(graph[from])}
		graph[from] = append(graph[from], forward)
		graph[to] = append(graph[to], reverse)
	}
	for node := 0; node < n; node++ {
		capacity := 1
		add(2*node, 2*node+1, capacity)
		if sources[node] {
			add(sourceID, 2*node, infinity)
		}
		if terminals[node] {
			add(2*node+1, sinkID, infinity)
		}
	}
	for _, edge := range network.edges {
		if len(edge.Initial) != 0 {
			continue
		}
		from, to := edge.From, edge.To
		if direction == RetimingBackward {
			from, to = to, from
		}
		add(2*from+1, 2*to, infinity)
	}
	flow := 0
	for flow < network.limits.MaxRegisters {
		if network.limits.Stop != nil && network.limits.Stop() {
			return RetimingCut{}, ErrCancelled
		}
		level := make([]int, len(graph))
		for i := range level {
			level[i] = -1
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
		for {
			pushed := send(sourceID, network.limits.MaxRegisters-flow)
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
	if reachable[sinkID] {
		return RetimingCut{}, ErrLimit
	}
	cut := make([]int, 0, flow)
	for node := 0; node < n; node++ {
		if reachable[2*node] && !reachable[2*node+1] {
			cut = append(cut, node)
		}
	}
	return RetimingCut{Flow: flow, Nodes: cut}, nil
}

// ApplyLags applies w'(u,v)=w(u,v)+r(v)-r(u). Initial-value placement is
// realized as deterministic local moves, which preserves sequential behavior.
func (network *RetimingNetwork) ApplyLags(lags []int) (RetimingRunResult, error) {
	if err := network.check(); err != nil {
		return RetimingRunResult{}, err
	}
	if len(lags) != len(network.nodes) {
		return RetimingRunResult{}, ErrDimension
	}
	remaining := slices.Clone(lags)
	result, moves := network.clone(), 0
	for {
		changed := false
		for node, lag := range remaining {
			if lag == 0 {
				continue
			}
			direction := RetimingBackward
			if lag < 0 {
				direction = RetimingForward
			}
			enabled, err := result.NodeEnabled(node, direction)
			if err != nil || !enabled {
				continue
			}
			if moves >= result.limits.MaxMoves {
				return RetimingRunResult{}, ErrLimit
			}
			result, err = result.Move(node, direction)
			if err != nil {
				return RetimingRunResult{}, err
			}
			if direction == RetimingForward {
				remaining[node]++
			} else {
				remaining[node]--
			}
			moves++
			changed = true
		}
		if !changed {
			break
		}
	}
	for _, lag := range remaining {
		if lag != 0 {
			return RetimingRunResult{}, ErrInvariant
		}
	}
	return RetimingRunResult{Network: result, Moves: moves}, nil
}

type RetimingLValueResult struct {
	Period int
	Lags   []int
}

func (network *RetimingNetwork) periodFeasible(period int) (bool, []int, error) {
	if period <= 0 {
		return false, nil, ErrDimension
	}
	const negativeInfinity = -1 << 50
	value := make([]int, len(network.nodes))
	for node := range value {
		if network.nodes[node].PrimaryInput || len(network.fanins(node)) == 0 {
			value[node] = 0
		} else {
			value[node] = negativeInfinity
		}
	}
	order, _ := network.combinationalOrder()
	converged := false
	for iteration := 0; iteration < network.limits.MaxIterations; iteration++ {
		if network.limits.Stop != nil && network.limits.Stop() {
			return false, nil, ErrCancelled
		}
		changed := false
		for _, node := range order {
			if network.nodes[node].PrimaryInput {
				continue
			}
			best := negativeInfinity
			for _, edgeIndex := range network.fanins(node) {
				edge := network.edges[edgeIndex]
				candidate := negativeInfinity
				if value[edge.From] > negativeInfinity/2 {
					if len(edge.Initial) > int(^uint(0)>>1)/period {
						return false, nil, ErrLimit
					}
					product := len(edge.Initial) * period
					if value[edge.From] >= negativeInfinity+product {
						candidate = value[edge.From] - product
					}
				}
				if candidate > best {
					best = candidate
				}
			}
			if best > int(^uint(0)>>1)-network.nodes[node].Delay {
				return false, nil, ErrLimit
			}
			best += network.nodes[node].Delay
			if best > value[node] {
				value[node] = best
				changed = true
			}
		}
		if !changed {
			converged = true
			break
		}
		for _, output := range network.outputs {
			if value[output] > period {
				return false, value, nil
			}
		}
	}
	return converged, value, nil
}

// LValue computes Pan-style feasible l-values and binary-searches the minimum
// integral clock period, following retLvalue.c.
func (network *RetimingNetwork) LValue() (RetimingLValueResult, error) {
	maximum, _, err := network.Timing(RetimingForward)
	if err != nil {
		return RetimingLValueResult{}, err
	}
	if maximum == 0 {
		return RetimingLValueResult{Period: 0, Lags: make([]int, len(network.nodes))}, nil
	}
	feasible, _, err := network.periodFeasible(maximum)
	if err != nil {
		return RetimingLValueResult{}, err
	}
	if !feasible {
		return RetimingLValueResult{}, ErrLimit
	}
	low, high := 0, maximum
	for low+1 < high {
		middle := low + (high-low)/2
		feasible, _, err = network.periodFeasible(middle)
		if err != nil {
			return RetimingLValueResult{}, err
		}
		if feasible {
			high = middle
		} else {
			low = middle
		}
	}
	_, values, err := network.periodFeasible(high)
	if err != nil {
		return RetimingLValueResult{}, err
	}
	lags := make([]int, len(values))
	for node, value := range values {
		if network.nodes[node].PrimaryInput {
			continue
		}
		if value <= -1<<49 {
			continue
		}
		// Overflow-free exact integer form of Abc_NodeComputeLag.
		lags[node] = value / high
		if value < 0 && value%high != 0 {
			lags[node]--
		}
		if value%high == 0 {
			lags[node]--
		}
	}
	return RetimingLValueResult{Period: high, Lags: lags}, nil
}
