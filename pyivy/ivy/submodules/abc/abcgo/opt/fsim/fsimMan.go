package fsim

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

func Fsim_ManCreate(g *aig.Graph) (*Manager, error) {
	if g == nil {
		return nil, fmt.Errorf("fsim: nil graph")
	}
	if err := g.Check(); err != nil {
		return nil, err
	}
	regs := int(aig.Aig_ManRegNum(g))
	p := &Manager{Graph: g, PIs: g.InputCount() - regs, POs: g.OutputCount() - regs, CIs: g.InputCount(), COs: g.OutputCount(), sourceIDs: map[int]int{0: 1}}
	p.ops = []Fsim_Obj_t{{}, {Node: 1, Kind: aig.Constant}}
	inputs := g.Inputs()
	positions := make(map[int]int)
	for i, e := range inputs {
		positions[e.NodeID()] = i
	}
	active := make(map[int]bool)
	var visit func(aig.Edge) int
	visit = func(e aig.Edge) int {
		id := e.NodeID()
		n := g.Node(e)
		if n.Kind == aig.Output {
			return visit(n.Fanins[0]) ^ int(e&1)
		}
		if copy, ok := p.sourceIDs[id]; ok {
			return copy*2 + int(e&1)
		}
		if active[id] {
			panic("fsim: cyclic source graph")
		}
		active[id] = true
		defer delete(active, id)
		o := Fsim_Obj_t{Kind: n.Kind, Source: aig.Edge(id * 2)}
		switch n.Kind {
		case aig.Input:
			o.Index = positions[id]
			p.CIOrder = append(p.CIOrder, o.Index)
		case aig.And, aig.ExclusiveOr:
			o.Fan0, o.Fan1 = visit(n.Fanins[0]), visit(n.Fanins[1])
			if o.Fan0 < o.Fan1 {
				o.Fan0, o.Fan1 = o.Fan1, o.Fan0
			}
		case aig.Buffer:
			o.Fan0 = visit(n.Fanins[0])
		default:
			panic("fsim: unsupported source kind")
		}
		o.Node = len(p.ops)
		p.ops = append(p.ops, o)
		p.sourceIDs[id] = o.Node
		return o.Node*2 + int(e&1)
	}
	// Preserve the source ordering: unused CIs first, then output-root DFS.
	for _, e := range inputs {
		if g.Node(e).Refs == 0 {
			visit(e)
		}
	}
	for i, e := range g.Outputs() {
		driver := g.Node(e).Fanins[0]
		literal := visit(driver)
		p.ops = append(p.ops, Fsim_Obj_t{Node: len(p.ops), Fan0: literal, Kind: aig.Output, Index: i, Source: e})
		p.sourceIDs[e.NodeID()] = len(p.ops) - 1
	}
	for _, e := range g.LiveEdges() {
		if g.Node(e).Kind != aig.Output {
			visit(e)
		}
	}
	p.references = make([]int, len(p.ops))
	for _, o := range p.ops[2:] {
		for _, f := range fsimFanins(o) {
			p.references[f>>1]++
		}
	}
	refs := append([]int(nil), p.references...)
	live := 0
	if refs[1] > 0 {
		live = 1
	}
	p.CrossCut = live
	peakTransient := live
	for _, o := range p.ops[2:] {
		if o.Kind != aig.Output {
			peakTransient = max(peakTransient, live+1)
		}
		for _, f := range fsimFanins(o) {
			refs[f>>1]--
			if refs[f>>1] == 0 {
				live--
			}
		}
		if o.Kind != aig.Output && p.references[o.Node] > 0 {
			live++
		}
		p.CrossCut = max(p.CrossCut, live)
	}
	if live != 0 {
		return nil, fmt.Errorf("fsim: schedule reference imbalance")
	}
	p.Frontier = max(2, peakTransient+1, 1+int(1.1*float64(p.CrossCut)))
	return p, nil
}
func Fsim_ManCreate_rec(p *Manager, root aig.Edge) (int, error) {
	if !p.Graph.ValidEdge(root) {
		return 0, fmt.Errorf("fsim: invalid source edge")
	}
	id, ok := p.sourceIDs[root.NodeID()]
	if !ok {
		return 0, fmt.Errorf("fsim: source edge not in schedule")
	}
	return id, nil
}
func Fsim_ManDelete(p *Manager) {
	if p != nil {
		*p = Manager{}
	}
}
func Fsim_ManTest(g *aig.Graph) error {
	p, err := Fsim_ManCreate(g)
	if err != nil {
		return err
	}
	defer Fsim_ManDelete(p)
	if err = Fsim_ManFront(p, false); err != nil {
		return err
	}
	return Fsim_ManVerifyFront(p)
}
