package dch

import (
	"container/heap"
	"sort"

	aig "github.com/glycerine/abcgo/aig/aig"
)

type tfoEntry struct{ id, rank int }
type tfoQueue []tfoEntry

func (q tfoQueue) Len() int           { return len(q) }
func (q tfoQueue) Less(i, j int) bool { return q[i].rank < q[j].rank }
func (q tfoQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *tfoQueue) Push(value any)    { *q = append(*q, value.(tfoEntry)) }
func (q *tfoQueue) Pop() any {
	last := len(*q) - 1
	value := (*q)[last]
	*q = (*q)[:last]
	return value
}

// Resimulate only changed TFO values, processing fanouts in topological order.
// Controlling inputs stop propagation when a node's value does not change.
func (m *Manager) simulateTFO(inputs []bool) map[int]bool {
	changed, queued := map[int]bool{}, map[int]bool{}
	pending := &tfoQueue{}
	enqueueFanouts := func(id int) {
		for _, fanout := range m.fanouts[id] {
			if !queued[fanout] {
				queued[fanout] = true
				heap.Push(pending, tfoEntry{fanout, m.ranks[fanout]})
			}
		}
	}
	for i, input := range m.graph.Inputs() {
		id := input.NodeID()
		if m.lastValues[id] != inputs[i] {
			m.lastValues[id] = inputs[i]
			changed[id] = true
			enqueueFanouts(id)
		}
	}
	value := func(edge aig.Edge) bool { return m.lastValues[edge.NodeID()] != edge.Complemented() }
	for pending.Len() > 0 {
		entry := heap.Pop(pending).(tfoEntry)
		delete(queued, entry.id)
		n := m.graph.Node(aig.Edge(entry.id) << 1)
		next := value(n.Fanins[0])
		switch n.Kind {
		case aig.And:
			next = next && value(n.Fanins[1])
		case aig.ExclusiveOr:
			next = next != value(n.Fanins[1])
		}
		if next != m.lastValues[entry.id] {
			m.lastValues[entry.id] = next
			changed[entry.id] = true
			enqueueFanouts(entry.id)
		}
	}
	return changed
}

// Refine candidates with an actual SAT counterexample. Classes are partitioned
// by the full counterexample history, so unchanged classes need no splitting.
func (m *Manager) refineCounterexample(inputs []bool, target ...aig.Edge) {
	var tfoCandidates map[int]bool
	var tfoClasses map[int]bool
	if m.params.SimulateTFO && len(target) == 2 {
		tfoCandidates, tfoClasses = map[int]bool{}, map[int]bool{}
		seen := map[int]bool{0: true}
		var visit func(int)
		visit = func(id int) {
			if seen[id] {
				return
			}
			seen[id] = true
			for _, fanout := range m.fanouts[id] {
				visit(fanout)
			}
			repr := m.classes.representative(aig.Edge(id << 1))
			if repr == aig.InvalidEdge {
				return
			}
			if repr.NodeID() == 0 {
				tfoCandidates[id] = true
			} else {
				tfoClasses[repr.NodeID()] = true
			}
		}
		visit(target[0].NodeID())
		visit(target[1].NodeID())
	}
	var values []bool
	var changed map[int]bool
	if len(target) == 2 {
		// The C resimulator fixes CIs in the queried cones to the SAT model,
		// then gives each newly encountered CI in the other selected cones a
		// fresh random value, visiting roots in increasing object-ID order.
		if m.random == nil {
			m.random = aig.NewAigRandomSource()
		}
		values = make([]bool, len(m.graph.Nodes()))
		visited := make([]bool, len(values))
		values[0], visited[0] = true, true
		inputIndex := make(map[int]int, m.graph.InputCount())
		for i, input := range m.graph.Inputs() {
			inputIndex[input.NodeID()] = i
		}
		var simulate func(aig.Edge, bool) bool
		simulate = func(edge aig.Edge, solved bool) bool {
			id := edge.NodeID()
			if !visited[id] {
				visited[id] = true
				n := m.graph.Node(aig.Aig_Regular(edge))
				if n.Kind == aig.Input {
					if solved {
						values[id] = inputs[inputIndex[id]]
					} else {
						values[id] = m.random.Uint32()&1 != 0
					}
				} else if n.Kind == aig.And {
					a := simulate(n.Fanins[0], solved)
					b := simulate(n.Fanins[1], solved)
					values[id] = a && b
				}
			}
			return values[id] != edge.Complemented()
		}
		simulate(target[0], true)
		simulate(target[1], true)
		var roots, classes []int
		for _, group := range m.classes.groups {
			head := group[0].NodeID()
			if head == 0 {
				for _, node := range group[1:] {
					id := node.NodeID()
					if tfoCandidates != nil && tfoCandidates[id] || !m.params.SimulateTFO && target[1].NodeID() == 0 && id >= target[0].NodeID() && id < target[0].NodeID()+500 {
						roots = append(roots, id)
					}
				}
			} else if tfoClasses != nil && tfoClasses[head] || !m.params.SimulateTFO && head == target[1].NodeID() {
				classes = append(classes, head)
			}
		}
		sort.Ints(roots)
		sort.Ints(classes)
		for _, id := range roots {
			simulate(aig.Edge(id<<1), false)
		}
		for _, head := range classes {
			group := m.classes.groups[m.classes.byNode[head]]
			for _, node := range group {
				simulate(node, false)
			}
		}
		m.lastValues = values
	} else if m.params.SimulateTFO {
		changed = m.simulateTFO(inputs)
		values = m.lastValues
	} else {
		values = simulateAssignment(m.graph, m.order, inputs)
		m.lastValues = values
	}
	refined := [][]aig.Edge{}
	for _, group := range m.classes.groups {
		if tfoClasses != nil && group[0].NodeID() != 0 && !tfoClasses[group[0].NodeID()] {
			refined = append(refined, group)
			continue
		}
		if !m.params.SimulateTFO && len(target) == 2 && group[0] != target[1] {
			refined = append(refined, group)
			continue
		}
		needsSplit := !m.params.SimulateTFO || len(target) == 2
		if !needsSplit {
			for _, node := range group {
				if changed[node.NodeID()] {
					needsSplit = true
					break
				}
			}
		}
		if !needsSplit {
			refined = append(refined, group)
			continue
		}
		zero, one := []aig.Edge{}, []aig.Edge{}
		for _, node := range group {
			if tfoCandidates != nil && group[0].NodeID() == 0 && !tfoCandidates[node.NodeID()] {
				zero = append(zero, node)
				continue
			}
			if !m.params.SimulateTFO && len(target) == 2 && target[1].NodeID() == 0 && (node.NodeID() < target[0].NodeID() || node.NodeID() >= target[0].NodeID()+500) {
				zero = append(zero, node)
				continue
			}
			if values[node.NodeID()] != aig.Aig_ObjPhase(m.graph, node) {
				one = append(one, node)
			} else {
				zero = append(zero, node)
			}
		}
		if len(zero) > 0 {
			refined = append(refined, zero)
		}
		if len(one) > 0 {
			refined = append(refined, one)
		}
	}
	m.classes.groups = refined
	m.classes.index()
}
