package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/sbd"
	"sort"
)

// abcSbdImport retains the source object-to-literal map even when strashing
// collapses a node or changes its phase. LUT truth tables use that same map.
func abcSbdImport(p *Gia_Man_t) (*sbd.Network, error) {
	g := aig.NewGraph()
	copies := make([]aig.Edge, gia.Gia_ManObjNum(p))
	copies[0] = g.Const0()
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		a, b, c, kind := p.Gate(id)
		edge := func(lit int32) aig.Edge { return copies[lit>>1] ^ aig.Edge(lit&1) }
		switch kind {
		case 1:
			copies[id] = g.NewInput()
		case 2:
			copies[id] = g.AND(edge(a), edge(b))
		case 3:
			copies[id] = g.XOR(edge(a), edge(b))
		case 4:
			copies[id] = g.Mux(edge(a), edge(b), edge(c))
		case 5:
			copies[id] = edge(a)
		case 6:
			copies[id] = g.AddOutput(edge(a))
		default:
			return nil, fmt.Errorf("sbd: unsupported GIA object %d", id)
		}
	}
	aig.Aig_ManSetRegNum(g, gia.Gia_ManRegNum(p))
	aig.Aig_ManSetConstrNum(g, gia.Gia_ManConstrNum(p))
	g.Name, g.Spec = p.Identity()
	n := sbd.NewNetwork(g)
	n.BoxLevelInputs = make(map[int32][]int32)
	if gia.Gia_ManHasMapping(p) != 0 {
		n.LUTs = make(map[int32]sbd.LUTCut)
		for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
			leaves, truth, ok := p.LUT(id)
			if !ok {
				continue
			}
			root := copies[id]
			if g.Node(root).Kind != aig.And && g.Node(root).Kind != aig.ExclusiveOr {
				continue
			}
			mapped := make([]aig.Edge, len(leaves))
			for i, leaf := range leaves {
				mapped[i] = copies[leaf]
			}
			n.LUTs[int32(root.NodeID())] = abcSbdCut(mapped, truth, root.Complemented())
		}
	}
	timing, _ := p.BoxTiming()
	if timing != nil {
		for box := 0; box < tim.Tim_ManBoxNum(timing); box++ {
			table := tim.Tim_ManBoxDelayTable(timing, box)
			inputs := tim.Tim_ManBoxInputNum(timing, box)
			for pin := 0; pin < tim.Tim_ManBoxOutputNum(timing, box); pin++ {
				ci := tim.Tim_ManBoxOutputFirst(timing, box) + pin
				id := int32(g.Inputs()[ci].NodeID())
				n.BoxLevelInputs[id] = []int32{}
				for k := 0; k < inputs; k++ {
					co := int32(g.Outputs()[tim.Tim_ManBoxInputFirst(timing, box)+k].NodeID())
					n.BoxInputs[id] = append(n.BoxInputs[id], co)
					if table == nil || table[3+pin*inputs+k] != -tim.TIM_ETERNITY {
						n.BoxLevelInputs[id] = append(n.BoxLevelInputs[id], co)
					}
				}
			}
		}
	}
	return n, nil
}

func abcSbdCut(edges []aig.Edge, truth []uint64, complement bool) sbd.LUTCut {
	leaves := []int32{}
	positions := map[int]int{}
	for _, edge := range edges {
		if edge.NodeID() == 0 {
			continue
		}
		if _, ok := positions[edge.NodeID()]; !ok {
			positions[edge.NodeID()] = len(leaves)
			leaves = append(leaves, int32(edge.NodeID()))
		}
	}
	values := make([]uint64, max(1, (1<<len(leaves))/64))
	for assignment := 0; assignment < len(values)*64; assignment++ {
		index := 0
		for i, edge := range edges {
			value := true
			if edge.NodeID() != 0 {
				value = assignment&(1<<positions[edge.NodeID()]) != 0
			}
			if edge.Complemented() {
				value = !value
			}
			if value {
				index |= 1 << i
			}
		}
		value := truth[index/64]&(uint64(1)<<uint(index%64)) != 0
		if value != complement {
			values[assignment/64] |= uint64(1) << uint(assignment%64)
		}
	}
	return sbd.LUTCut{Leaves: leaves, Truth: values}
}

func abcSbdPerform(p *Gia_Man_t, parameters *sbd.Parameters) (*Gia_Man_t, error) {
	input, err := abcSbdImport(p)
	if err != nil {
		return nil, err
	}
	defer aig.Aig_ManStop(input.Graph)
	result := sbd.Sbd_NtkPerform(input, parameters)
	if result == nil {
		return nil, fmt.Errorf("sbd: optimization failed")
	}
	defer aig.Aig_ManStop(result.Graph)
	out := gia.Gia_ManFromAigSimple(result.Graph)
	if parameters.Mapping {
		data := make([]int32, gia.Gia_ManObjNum(out))
		ids := make([]int, 0, len(result.LUTs))
		for id := range result.LUTs {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		for _, id := range ids {
			root := int32(aig.Aig_ObjCopy(result.Graph, aig.Edge(id)<<1)) >> 1
			if root == 0 || gia.Gia_ObjIsAnd(gia.Gia_ManObj(out, root)) == 0 {
				continue
			}
			cut := result.LUTs[int32(id)]
			leaves := []int32{}
			seen := map[int32]bool{}
			for _, leaf := range cut.Leaves {
				mapped := int32(aig.Aig_ObjCopy(result.Graph, aig.Edge(leaf)<<1)) >> 1
				if !seen[mapped] {
					seen[mapped] = true
					leaves = append(leaves, mapped)
				}
			}
			data[root] = int32(len(data))
			data = append(data, int32(len(leaves)))
			data = append(data, leaves...)
			data = append(data, root)
		}
		if err := out.SetMappedData(vec.Vec_IntAllocArray(data, len(data)), nil, nil, ""); err != nil {
			gia.Gia_ManStop(out)
			return nil, err
		}
	}
	timing, extra := p.BoxTiming()
	out.SetBoxTiming(timing, extra)
	p.SetBoxTiming(nil, nil)
	return out, nil
}
