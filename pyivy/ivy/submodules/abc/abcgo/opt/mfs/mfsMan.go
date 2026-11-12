package mfs

import "fmt"

func Mfs_ManAlloc(pars *Mfs_Par_t) *Mfs_Man_t {
	p := DefaultParams()
	if pars != nil {
		p = *pars
	}
	if err := validateParams(p); err != nil {
		panic(err)
	}
	return &Mfs_Man_t{Params: p, Target: -1}
}
func Mfs_ManClean(p *Mfs_Man_t) {
	p.Target = -1
	p.Roots = nil
	p.Support = nil
	p.Nodes = nil
	p.Divisors = nil
	p.Window = nil
	p.Values = nil
	p.Care = nil
	p.Solver = nil
	p.clauses = nil
	p.projection = nil
	p.candidates = nil
	p.witnesses = nil
}
func Mfs_ManStop(p *Mfs_Man_t) { *p = Mfs_Man_t{} }
func (p *Mfs_Man_t) refresh() error {
	order, err := p.Network.topology()
	if err != nil {
		return err
	}
	n := len(p.Network.Nodes)
	p.levels = make([]int, n)
	p.fanouts = make([][]int, n)
	p.outputDriver = make([]bool, n)
	for _, id := range order {
		for _, f := range p.Network.Nodes[id].Fanins {
			p.fanouts[f] = append(p.fanouts[f], id)
			p.levels[id] = max(p.levels[id], p.levels[f]+1)
		}
		p.Network.Nodes[id].Level = p.levels[id]
	}
	maxlevel := 0
	for _, e := range p.Network.Outputs {
		p.outputDriver[e.ID()] = true
		maxlevel = max(maxlevel, p.levels[e.ID()])
	}
	p.required = make([]int, n)
	for i := range p.required {
		p.required[i] = maxlevel + p.Params.GrowthLevel
	}
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		for _, f := range p.Network.Nodes[id].Fanins {
			p.required[f] = min(p.required[f], p.required[id]-1)
		}
	}
	return nil
}
func counts(n *Network) (nodes, edges int) {
	for id, v := range n.Nodes {
		if id != 0 && !v.Input && !v.Deleted {
			nodes++
			edges += len(v.Fanins)
		}
	}
	return
}
func Abc_NtkMfsPrintResubStats(p *Mfs_Man_t) {
	fmt.Printf("MFS: tried %d, removed %d, resubstituted %d, SAT %d, timeouts %d, nodes %d -> %d, edges %d -> %d\n", p.Stats.NodesTried, p.Stats.Removes, p.Stats.Resubs, p.Stats.SATCalls, p.Stats.Timeouts, p.Stats.NodesBefore, p.Stats.NodesAfter, p.Stats.EdgesBefore, p.Stats.EdgesAfter)
}

func validateParams(p Mfs_Par_t) error {
	if p.WinTfoLevs < 0 || p.FanoutsMax < 0 || p.DepthMax < 0 || p.WinMax < 0 || p.GrowthLevel < 0 || p.BTLimit < 0 {
		return fmt.Errorf("mfs: negative parameter")
	}
	return nil
}
