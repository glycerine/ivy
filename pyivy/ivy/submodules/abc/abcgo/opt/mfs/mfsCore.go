package mfs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/opt/sim"
	"math/bits"
	"sort"
)

func gateCount(g *aig.Graph) int {
	c := 0
	for _, e := range g.LiveEdges() {
		k := g.Node(e).Kind
		if k == aig.And || k == aig.ExclusiveOr {
			c++
		}
	}
	return c
}
func decompose(truth, care []uint32, vars int) ([]uint32, *aig.Graph) {
	manager := bdc.Bdc_ManAlloc(&bdc.Bdc_Par_t{NVarsMax: int32(max(3, vars))})
	defer bdc.Bdc_ManFree(manager)
	if bdc.Bdc_ManDecompose(manager, truth, care, int32(vars), nil, 100000) < 0 {
		return nil, nil
	}
	g := aig.NewGraph()
	mapped := map[*bdc.Bdc_Fun_t]aig.Edge{bdc.Bdc_ManFunc(manager, 0): g.Const1()}
	for i := 0; i < vars; i++ {
		mapped[bdc.Bdc_ManFunc(manager, int32(i+1))] = g.NewInput()
	}
	var rec func(*bdc.Bdc_Fun_t) aig.Edge
	rec = func(f *bdc.Bdc_Fun_t) aig.Edge {
		regular := bdc.Bdc_Regular(f)
		e, ok := mapped[regular]
		if !ok {
			a, b := rec(regular.Fan0), rec(regular.Fan1)
			switch regular.Type {
			case bdc.BDC_TYPE_AND:
				e = g.AND(a, b)
			case bdc.BDC_TYPE_OR:
				e = g.OR(a, b)
			case bdc.BDC_TYPE_XOR:
				e = g.XOR(a, b)
			default:
				panic("mfs: unsupported BDC gate")
			}
			mapped[regular] = e
		}
		if bdc.Bdc_IsComplement(f) != 0 {
			e = e.Not()
		}
		return e
	}
	root := bdc.Bdc_ManRoot(manager)
	g.AddOutput(rec(root))
	result := append([]uint32(nil), bdc.Bdc_Regular(root).Func...)
	if bdc.Bdc_IsComplement(root) != 0 {
		for i := range result {
			result[i] = ^result[i]
		}
	}
	return normalizeTruth(result, vars), g
}
func Abc_NtkMfsNode(p *Mfs_Man_t, target int) bool {
	if !Abc_WinNode(p, target) {
		return false
	}
	if !Abc_NtkMfsSolveSat(p, target) {
		return false
	}
	v := &p.Network.Nodes[target]
	truthInput, careInput := v.Truth, p.Care
	if p.Params.Power {
		prob := p.probabilities[target]
		if 2*prob*(1-prob) >= 0.4 {
			truthInput = append([]uint32(nil), truthInput...)
			for i := range truthInput {
				if prob > 0.5 {
					truthInput[i] |= ^careInput[i]
				} else {
					truthInput[i] &= careInput[i]
				}
			}
			careInput = nil
		}
	}
	truth, local := decompose(truthInput, careInput, len(v.Fanins))
	if truth == nil {
		return false
	}
	old := v.local
	if old == nil {
		old = truthGraph(v.Truth, len(v.Fanins))
	}
	gain := gateCount(old) - gateCount(local)
	if gain >= 0 {
		v.Truth, v.local = truth, local
		p.Stats.Decomposed++
		p.Stats.NodesGained += gain
	}
	return true
}
func Abc_NtkMfsResub(p *Mfs_Man_t, target int) bool {
	if !Abc_WinNode(p, target) {
		return false
	}
	Abc_MfsCreateSolverResub(p, nil, false)
	changed := false
	if p.Params.Power {
		changed = Abc_NtkMfsEdgePower(p, target)
	} else if p.Params.SwapEdge {
		return Abc_NtkMfsEdgeSwapEval(p, target) > 0
	} else {
		changed = Abc_NtkMfsResubNode(p, target)
		if p.Params.MoreEffort {
			if changed {
				if !Abc_WinNode(p, target) {
					return true
				}
				Abc_MfsCreateSolverResub(p, nil, false)
			}
			changed = Abc_NtkMfsResubNode2(p, target) || changed
		}
	}
	return changed
}
func initialize(pars *Mfs_Par_t, n *Network) (*Mfs_Man_t, error) {
	if pars != nil {
		if err := validateParams(*pars); err != nil {
			return nil, err
		}
	}
	p := Mfs_ManAlloc(pars)
	p.Network = n
	if err := p.refresh(); err != nil {
		return nil, err
	}
	p.Stats.NodesBefore, p.Stats.EdgesBefore = counts(n)
	if p.Params.Power {
		p.probabilities = estimateProbabilities(n, !p.Params.Resub)
	}
	return p, nil
}

// Optimize applies the source node-order resubstitution or level-order BDC pass.
// Statistics are returned by value. Invalid input is rejected before mutation.
func Optimize(n *Network, pars *Mfs_Par_t) (Stats, error) {
	p, err := initialize(pars, n)
	if err != nil {
		return Stats{}, err
	}
	order, _ := n.topology()
	if !p.Params.Resub { // topology has nondecreasing dependencies; group explicitly by levels.
		sort.SliceStable(order, func(i, j int) bool { return p.levels[order[i]] < p.levels[order[j]] })
	} else {
		order = order[:0]
		for id := range n.Nodes {
			order = append(order, id)
		}
	}
	limit := MFS_FANIN_MAX
	if p.Params.Resub {
		limit = 8
	}
	if p.Params.Resub && p.Params.Power {
		Abc_NtkMfsPowerResub(p)
	} else {
		for _, id := range order {
			v := n.Nodes[id]
			if id == 0 || v.Input || v.Deleted || len(v.Fanins) < 2 || len(v.Fanins) > limit || (p.Params.DepthMax > 0 && p.levels[id] > p.Params.DepthMax) {
				continue
			}
			if p.Params.Resub {
				Abc_NtkMfsResub(p, id)
			} else {
				Abc_NtkMfsNode(p, id)
			}
		}
	}
	p.Stats.NodesAfter, p.Stats.EdgesAfter = counts(n)
	if p.Params.Verbose {
		Abc_NtkMfsPrintResubStats(p)
	}
	return p.Stats, n.Check()
}
func Abc_NtkMfs(n *Network, pars *Mfs_Par_t) bool {
	_, err := Optimize(n, pars)
	if err != nil {
		if pars != nil && pars.Verbose {
			fmt.Println(err)
		}
		return false
	}
	return true
}

// estimateProbabilities follows the source 48-frame/16-prefix, ten-word
// sequential protocol. Transition words pack previous/current 16-bit lanes.
func estimateProbabilities(n *Network, one bool) []float32 {
	copy := *n
	copy.Outputs = nil
	var ids []int
	for id, v := range n.Nodes {
		if !v.Deleted {
			ids = append(ids, id)
			copy.Outputs = append(copy.Outputs, Literal(id<<1))
		}
	}
	if n.Registers > 0 {
		copy.Outputs = append(copy.Outputs, n.Outputs[len(n.Outputs)-n.Registers:]...)
	}
	g, err := copy.ToAIG()
	if err != nil {
		panic(err)
	}
	p := sim.Sim_ManStart(g, true)
	p.Words = 10
	p.Sim0 = sim.Sim_UtilInfoAlloc(int(aig.Aig_ManObjNumMax(g)), 10, true)
	cis, cos := g.Inputs(), g.Outputs()
	pis := len(cis) - n.Registers
	rng := aig.NewAigRandomSource()
	rng.Uint32()
	for _, ci := range cis[:pis] {
		row := p.Sim0[ci.NodeID()]
		for w := 9; w >= 0; w-- {
			row[w] = rng.Uint32()
		}
	}
	counts := make([]uint64, len(ids))
	for frame := 0; frame < 48; frame++ {
		sim.Sim_UtilSimulate(p, false)
		if frame >= 16 {
			for i, out := range cos[:len(ids)] {
				for _, word := range p.Sim0[out.NodeID()] {
					if one {
						counts[i] += uint64(bits.OnesCount32(word))
					} else {
						counts[i] += uint64(2 * bits.OnesCount32((word^(word>>16))&0xffff))
					}
				}
			}
		}
		if frame == 47 {
			break
		}
		for r, ci := range cis[pis:] {
			dst, src := p.Sim0[ci.NodeID()], p.Sim0[cos[len(ids)+r].NodeID()]
			for w := 9; w >= 0; w-- {
				if one {
					dst[w] = src[w]
				} else {
					dst[w] = (dst[w] << 16) | (src[w] & 0xffff)
				}
			}
		}
		for _, ci := range cis[:pis] {
			row := p.Sim0[ci.NodeID()]
			if one {
				for w := 9; w >= 0; w-- {
					row[w] = rng.Uint32()
				}
			} else {
				mask := rng.Uint32()
				for w := 9; w >= 0; w-- {
					row[w] = (row[w] << 16) | ((row[w] ^ mask) & 0xffff)
				}
			}
		}
	}
	result := make([]float32, len(n.Nodes))
	for i, id := range ids {
		result[id] = float32(counts[i]) / float32(32*10*32)
	}
	return result
}
func Abc_NtkMfsTotalSwitching(n *Network) float64 {
	p := estimateProbabilities(n, false)
	sum := 0.0
	for id, v := range n.Nodes {
		if id != 0 && !v.Input && !v.Deleted {
			sum += float64(p[id])
		}
	}
	return sum
}
