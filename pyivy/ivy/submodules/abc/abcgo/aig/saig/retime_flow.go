package saig

// Vertex-capacitated maximum flow for the retiming cuts computed by ABC's
// Nwk_ManRetimeCutForward and Nwk_ManRetimeCutBackward in nwkFlow.c.
import "github.com/glycerine/abcgo/aig/aig"

// Saig_ManDeriveRetimingCut returns the minimum retiming cut without importing
// opt/nwk, which itself depends on saig. The returned edges are regular.
func Saig_ManDeriveRetimingCut(g *aig.Graph, forward bool) []aig.Edge {
	return saigRetimeMinCut(g, forward)
}

type retimeFlowEdge struct{ to, reverse, capacity int }
type retimeFlow struct {
	edges       [][]retimeFlowEdge
	level, next []int
}

func (f *retimeFlow) add(from, to, capacity int) {
	a := retimeFlowEdge{to: to, reverse: len(f.edges[to]), capacity: capacity}
	b := retimeFlowEdge{to: from, reverse: len(f.edges[from])}
	f.edges[from] = append(f.edges[from], a)
	f.edges[to] = append(f.edges[to], b)
}
func (f *retimeFlow) levels(source, sink int) bool {
	for i := range f.level {
		f.level[i] = -1
	}
	f.level[source] = 0
	q := []int{source}
	for len(q) > 0 {
		v := q[0]
		q = q[1:]
		for _, e := range f.edges[v] {
			if e.capacity > 0 && f.level[e.to] < 0 {
				f.level[e.to] = f.level[v] + 1
				q = append(q, e.to)
			}
		}
	}
	return f.level[sink] >= 0
}
func (f *retimeFlow) push(v, sink, available int) int {
	if v == sink {
		return available
	}
	for f.next[v] < len(f.edges[v]) {
		index := f.next[v]
		edge := f.edges[v][index]
		if edge.capacity > 0 && f.level[edge.to] == f.level[v]+1 {
			sent := f.push(edge.to, sink, min(available, edge.capacity))
			if sent > 0 {
				f.edges[v][index].capacity -= sent
				f.edges[edge.to][edge.reverse].capacity += sent
				return sent
			}
		}
		f.next[v]++
	}
	return 0
}

// saigRetimeMinCut returns regular AIG nodes on the minimum vertex cut. The
// forward cut separates latch outputs from primary-output/PI fanout cones;
// the backward cut separates latch-input drivers from primary-input/PO cones.
func saigRetimeMinCut(g *aig.Graph, forward bool) []aig.Edge {
	nodes, inputs, outputs := g.Nodes(), g.Inputs(), g.Outputs()
	regs := int(g.RegisterCount())
	pis, pos := len(inputs)-regs, len(outputs)-regs
	if regs < 0 || pis < 0 || pos < 0 {
		panic("saig: invalid retiming interface")
	}
	count := len(nodes)
	source, sink := 2*count, 2*count+1
	f := retimeFlow{edges: make([][]retimeFlowEdge, 2*count+2), level: make([]int, 2*count+2), next: make([]int, 2*count+2)}
	inf := count + 1
	fanin, fanout := make([][]int, count), make([][]int, count)
	for _, e := range g.LiveEdges() {
		id := e.NodeID()
		n := nodes[id]
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			fanin[id] = append(fanin[id], n.Fanins[0].NodeID(), n.Fanins[1].NodeID())
		case aig.Buffer, aig.Output:
			fanin[id] = append(fanin[id], n.Fanins[0].NodeID())
		}
		for _, x := range fanin[id] {
			fanout[x] = append(fanout[x], id)
		}
		cap := 1
		if n.Kind == aig.Output {
			cap = inf
		}
		f.add(2*id, 2*id+1, cap)
	}
	boundary := make([]bool, count)
	var mark func(int, [][]int)
	mark = func(id int, adj [][]int) {
		if boundary[id] {
			return
		}
		boundary[id] = true
		for _, next := range adj[id] {
			mark(next, adj)
		}
	}
	if forward {
		for _, e := range outputs {
			boundary[e.NodeID()] = true
		}
		for _, e := range inputs[:pis] {
			mark(e.NodeID(), fanout)
		}
		for _, e := range inputs[pis:] {
			f.add(source, 2*e.NodeID(), inf)
		}
	} else {
		boundary[0] = true
		for _, e := range inputs {
			boundary[e.NodeID()] = true
		}
		for _, e := range outputs[:pos] {
			mark(e.NodeID(), fanin)
		}
		for _, e := range outputs[pos:] {
			f.add(source, 2*nodes[e.NodeID()].Fanins[0].NodeID(), inf)
		}
	}
	for _, e := range g.LiveEdges() {
		id := e.NodeID()
		if boundary[id] {
			f.add(2*id+1, sink, inf)
		}
		adj := fanin[id]
		if forward {
			adj = fanout[id]
		}
		for _, next := range adj {
			f.add(2*id+1, 2*next, inf)
		}
	}
	for f.levels(source, sink) {
		clear(f.next)
		for {
			sent := f.push(source, sink, inf)
			if sent == 0 {
				break
			}
		}
	}
	reachable := make([]bool, len(f.edges))
	reachable[source] = true
	q := []int{source}
	for len(q) > 0 {
		v := q[0]
		q = q[1:]
		for _, e := range f.edges[v] {
			if e.capacity > 0 && !reachable[e.to] {
				reachable[e.to] = true
				q = append(q, e.to)
			}
		}
	}
	var cut []aig.Edge
	for _, edge := range g.LiveEdges() {
		id := edge.NodeID()
		if reachable[2*id] && !reachable[2*id+1] {
			cut = append(cut, edge)
		}
	}
	return cut
}
