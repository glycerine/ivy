package dch

import (
	"container/heap"

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
	var values []bool
	var changed map[int]bool
	if m.params.SimulateTFO {
		changed = m.simulateTFO(inputs)
		values = m.lastValues
	} else {
		values = simulateAssignment(m.graph, m.order, inputs)
		m.lastValues = values
	}
	refined := [][]aig.Edge{}
	for _, group := range m.classes.groups {
		if !m.params.SimulateTFO && len(target) == 2 && group[0] != target[1] {
			refined = append(refined, group)
			continue
		}
		needsSplit := !m.params.SimulateTFO
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
