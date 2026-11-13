package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestRetimeMinimumCutBackwardSharedDriver(t *testing.T) {
	g := aig.NewGraph()
	x, y, q1, q2 := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	g.AddOutput(g.XOR(q1, q2))
	g.AddOutput(gate)
	g.AddOutput(gate)
	aig.Aig_ManSetRegNum(g, 2)
	cut := saigRetimeMinCut(g, false)
	if len(cut) != 1 || cut[0] != gate {
		t.Fatalf("backward cut=%v, want shared driver %v", cut, gate)
	}
}
func TestRetimeMinimumCutForwardSharedFanout(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	q1, q2 := g.NewInput(), g.NewInput()
	gate := g.AND(q1, q2)
	g.AddOutput(gate)
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 2)
	cut := saigRetimeMinCut(g, true)
	if len(cut) != 1 || cut[0] != gate {
		t.Fatalf("forward cut=%v, want shared fanout %v", cut, gate)
	}
}
func TestRetimeMinimumCutRespectsPrimaryInputBoundary(t *testing.T) {
	g := aig.NewGraph()
	input := g.NewInput()
	q1, q2 := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(input, q1))
	g.AddOutput(q2)
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 2)
	if cut := saigRetimeMinCut(g, true); len(cut) != 2 {
		t.Fatalf("crossed PI boundary: cut=%v", cut)
	}
}

func TestRetimeMinimumCutsMatchExhaustiveSearch(t *testing.T) {
	for shape := 0; shape < 16; shape++ {
		g := aig.NewGraph()
		x, y, a, b := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
		left := g.AND(a, x)
		right := g.AND(b, y)
		shared := g.AND(left, right)
		choices := []aig.Edge{x, y, a, b, left, right, shared, g.Const1()}
		g.AddOutput(choices[shape%len(choices)])
		g.AddOutput(choices[(shape+3)%len(choices)])
		g.AddOutput(choices[(shape+5)%len(choices)])
		aig.Aig_ManSetRegNum(g, 2)
		for _, forward := range []bool{false, true} {
			cut := saigRetimeMinCut(g, forward)
			if want := bruteRetimeCutSize(g, forward); len(cut) != want {
				t.Fatalf("shape=%d forward=%v got=%v want size %d", shape, forward, cut, want)
			}
		}
	}
}

func bruteRetimeCutSize(g *aig.Graph, forward bool) int {
	nodes, inputs, outputs := g.Nodes(), g.Inputs(), g.Outputs()
	regs := int(g.RegisterCount())
	pis, pos := len(inputs)-regs, len(outputs)-regs
	fanin, fanout := make([][]int, len(nodes)), make([][]int, len(nodes))
	var vertices []int
	for _, e := range g.LiveEdges() {
		id := e.NodeID()
		n := nodes[id]
		switch n.Kind {
		case aig.And, aig.ExclusiveOr:
			fanin[id] = []int{n.Fanins[0].NodeID(), n.Fanins[1].NodeID()}
		case aig.Buffer, aig.Output:
			fanin[id] = []int{n.Fanins[0].NodeID()}
		}
		for _, v := range fanin[id] {
			fanout[v] = append(fanout[v], id)
		}
		if n.Kind != aig.Output {
			vertices = append(vertices, id)
		}
	}
	sinks := make([]bool, len(nodes))
	var mark func(int, [][]int)
	mark = func(id int, adj [][]int) {
		if sinks[id] {
			return
		}
		sinks[id] = true
		for _, n := range adj[id] {
			mark(n, adj)
		}
	}
	var roots []int
	adj := fanin
	if forward {
		adj = fanout
		for _, e := range outputs {
			sinks[e.NodeID()] = true
		}
		for _, e := range inputs[:pis] {
			mark(e.NodeID(), adj)
		}
		for _, e := range inputs[pis:] {
			roots = append(roots, e.NodeID())
		}
	} else {
		sinks[0] = true
		for _, e := range inputs {
			sinks[e.NodeID()] = true
		}
		for _, e := range outputs[:pos] {
			mark(e.NodeID(), adj)
		}
		for _, e := range outputs[pos:] {
			roots = append(roots, nodes[e.NodeID()].Fanins[0].NodeID())
		}
	}
	best := len(vertices) + 1
	for mask := 0; mask < (1 << len(vertices)); mask++ {
		removed := make([]bool, len(nodes))
		size := 0
		for i, id := range vertices {
			if mask&(1<<i) != 0 {
				removed[id] = true
				size++
			}
		}
		if size >= best {
			continue
		}
		visited := make([]bool, len(nodes))
		queue := append([]int(nil), roots...)
		valid := true
		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]
			if removed[id] || visited[id] {
				continue
			}
			visited[id] = true
			if sinks[id] {
				valid = false
				break
			}
			queue = append(queue, adj[id]...)
		}
		if valid {
			best = size
		}
	}
	return best
}
