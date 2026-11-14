package cut

import "fmt"

// NetworkNode is a structural view supplied by a network adapter. No ABC frame
// or opaque C object pointers are needed by cut enumeration.
type NetworkNode struct {
	ID                                  int32
	FanoutAdjust                        int32
	Fanins                              [2]int32
	Compl                               [2]bool
	Latches                             [2]int32
	Input, Output, Constant, MuxControl bool
	Value, XOR, Buffer                  bool
	Choices                             []int32
}
type Network interface{ CutNodes() []NetworkNode }
type NodeAttributeProvider interface{ CutNodeAttributes() []int32 }

func networkInventory(network Network) (map[int32]NetworkNode, []int32) {
	nodes := network.CutNodes()
	maxID := int32(0)
	lookup := make(map[int32]NetworkNode, len(nodes))
	for _, node := range nodes {
		if node.ID < 0 {
			panic("cut: negative network ID")
		}
		if _, duplicate := lookup[node.ID]; duplicate {
			panic("cut: duplicate network ID")
		}
		lookup[node.ID] = node
		maxID = max(maxID, node.ID)
	}
	counts := make([]int32, maxID+1)
	for _, node := range nodes {
		pins := 2
		if node.Input || node.Constant {
			pins = 0
		}
		if node.Output || node.Buffer {
			pins = 1
		}
		for pin := 0; pin < pins; pin++ {
			fanin := node.Fanins[pin]
			if _, ok := lookup[fanin]; !ok {
				panic("cut: missing network fanin")
			}
			counts[fanin]++
		}
	}
	for _, node := range nodes {
		counts[node.ID] += node.FanoutAdjust
		if counts[node.ID] < 0 {
			panic("cut: negative adjusted fanout count")
		}
	}
	return lookup, counts
}
func Abc_NtkGetNodeAttributes(network Network) []int32 {
	if provider, ok := network.(NodeAttributeProvider); ok {
		return append([]int32(nil), provider.CutNodeAttributes()...)
	}
	nodes, counts := networkInventory(network)
	attrs := make([]int32, len(counts))
	for id, node := range nodes {
		if counts[id] > 1 && !node.MuxControl {
			attrs[id] = 1
		}
	}
	return attrs
}
func Abc_NtkCuts(network Network, params *Cut_Params_t) *Cut_Man_t {
	p := Cut_ManStart(params)
	p.network, p.networkCounts = networkInventory(network)
	if params.FGlobal != 0 || params.FLocal != 0 {
		p.attrs = Abc_NtkGetNodeAttributes(network)
	}
	if params.FDrop != 0 {
		p.fanCounts = append([]int32(nil), p.networkCounts...)
	}
	for _, node := range network.CutNodes() {
		if node.Input && p.networkCounts[node.ID] > 0 {
			Cut_NodeSetTriv(p, node.ID)
		}
	}
	active := map[int32]bool{}
	done := map[int32]bool{}
	var visit func(int32)
	visit = func(id int32) {
		if done[id] {
			return
		}
		node, ok := p.network[id]
		if !ok {
			panic("cut: missing network node")
		}
		if active[id] {
			panic("cut: combinational network cycle")
		}
		active[id] = true
		if node.Input {
			if p.newCuts[id] == nil {
				Cut_NodeSetTriv(p, id)
			}
		} else if node.Constant {
			cutSetConstant(p, node.ID, node.Value)
		} else if node.Output {
			visit(node.Fanins[0])
			if params.FDrop != 0 {
				Cut_NodeTryDroppingCuts(p, node.Fanins[0])
			}
		} else if node.Buffer {
			visit(node.Fanins[0])
			Cut_NodeComputeBuffer(p, node.ID, node.Fanins[0], cutBool(node.Compl[0]))
			if params.FDrop != 0 {
				Cut_NodeTryDroppingCuts(p, node.Fanins[0])
			}
		} else {
			visit(node.Fanins[0])
			visit(node.Fanins[1])
			Abc_NodeGetCuts(p, node, params.FDag, params.FTree)
			if len(node.Choices) > 0 {
				ids := append([]int32{id}, node.Choices...)
				for _, choice := range node.Choices {
					visit(choice)
				}
				Cut_NodeUnionCuts(p, ids)
			}
			if params.FDrop != 0 {
				Cut_NodeTryDroppingCuts(p, node.Fanins[0])
				Cut_NodeTryDroppingCuts(p, node.Fanins[1])
			}
		}
		active[id] = false
		done[id] = true
	}
	for _, node := range network.CutNodes() {
		if node.Output {
			visit(node.ID)
		}
	}
	for _, node := range network.CutNodes() {
		if !done[node.ID] && !node.Output {
			visit(node.ID)
		}
	}
	return p
}
func Abc_NtkCutsOracle(network Network, p *Cut_Oracle_t) {
	nodes, fanoutCounts := networkInventory(network)
	if Cut_OracleReadDrop(p) != 0 {
		Cut_OracleSetFanoutCounts(p, fanoutCounts)
	}
	done := map[int32]bool{}
	active := map[int32]bool{}
	var visit func(int32)
	visit = func(id int32) {
		if done[id] {
			return
		}
		if active[id] {
			panic("cut: oracle network cycle")
		}
		node := nodes[id]
		active[id] = true
		if node.Input {
			Cut_OracleNodeSetTriv(p, id)
		} else if node.Constant {
			cutSetConstant(p.manager, id, node.Value)
		} else if !node.Output {
			for _, step := range p.records[id].Steps {
				if step.Proof.Kind != 0 {
					visit(step.Proof.Sources[0])
					if step.Proof.Kind != 3 {
						visit(step.Proof.Sources[1])
					}
				}
			}
			visit(node.Fanins[0])
			if !node.Buffer {
				visit(node.Fanins[1])
			}
			Cut_OracleComputeCuts(p, id, node.Fanins[0], node.Fanins[1], cutBool(node.Compl[0]), cutBool(node.Compl[1]))
			if Cut_OracleReadDrop(p) != 0 {
				Cut_OracleTryDroppingCuts(p, node.Fanins[0])
				if !node.Buffer {
					Cut_OracleTryDroppingCuts(p, node.Fanins[1])
				}
			}
		} else if node.Output {
			visit(node.Fanins[0])
			if Cut_OracleReadDrop(p) != 0 {
				Cut_OracleTryDroppingCuts(p, node.Fanins[0])
			}
		}
		done[id] = true
		active[id] = false
	}
	for _, node := range network.CutNodes() {
		visit(node.ID)
	}
}
func cutBool(value bool) int32 {
	if value {
		return 1
	}
	return 0
}
func Abc_NodeGetCutsRecursive(p *Cut_Man_t, node NetworkNode, dag, tree int32) *Cut_Cut_t {
	if list := p.newCuts[node.ID]; list != nil {
		return list
	}
	if node.Input {
		Cut_NodeSetTriv(p, node.ID)
		return p.newCuts[node.ID]
	}
	for _, id := range node.Fanins {
		child, ok := p.network[id]
		if !ok {
			panic("cut: missing recursive fanin")
		}
		Abc_NodeGetCutsRecursive(p, child, dag, tree)
	}
	return Abc_NodeGetCuts(p, node, dag, tree)
}
func Abc_NodeGetCuts(p *Cut_Man_t, node NetworkNode, dag, tree int32) *Cut_Cut_t {
	dagNode := p.networkCounts[node.ID] > 1 && !node.MuxControl
	trivial := cutBool(dagNode || dag == 0)
	if dagNode {
		Cut_ManIncrementDagNodes(p)
	}
	var code int32
	for pin, id := range node.Fanins {
		if tree != 0 && p.networkCounts[id] > 1 && !p.network[id].MuxControl {
			code |= 1 << uint(pin)
		}
	}
	if p.params.FLocal != 0 {
		trivial = cutBool(p.attrs[node.ID] == 0)
		for pin, id := range node.Fanins {
			if p.attrs[id] != 0 {
				code |= 1 << uint(pin)
			}
		}
	}
	if node.XOR {
		return Cut_NodeComputeCutsXor(p, node.ID, node.Fanins[0], node.Fanins[1], cutBool(node.Compl[0]), cutBool(node.Compl[1]), trivial, code)
	}
	return Cut_NodeComputeCuts(p, node.ID, node.Fanins[0], node.Fanins[1], cutBool(node.Compl[0]), cutBool(node.Compl[1]), trivial, code)
}
func Abc_NodeGetCutsSeq(p *Cut_Man_t, node NetworkNode, trivial int32) {
	Cut_NodeComputeCutsSeq(p, node.ID, node.Fanins[0], node.Fanins[1], cutBool(node.Compl[0]), cutBool(node.Compl[1]), node.Latches[0], node.Latches[1], trivial, -1)
}
func Abc_NodeReadCuts(p *Cut_Man_t, node NetworkNode) *Cut_Cut_t {
	return Cut_NodeReadCutsNew(p, node.ID)
}
func Abc_NodeFreeCuts(p *Cut_Man_t, node NetworkNode) { Cut_NodeFreeCuts(p, node.ID) }
func Abc_NtkPrintCuts(p *Cut_Man_t, network Network, seq int32) {
	for _, node := range network.CutNodes() {
		if node.Output {
			continue
		}
		fmt.Printf("Node %d:\n", node.ID)
		Cut_CutPrintList(p.newCuts[node.ID], seq)
	}
}
func Abc_NtkPrintCuts_(p *Cut_Man_t, network Network, seq int32) { Abc_NtkPrintCuts(p, network, seq) }

// Abc_NtkSeqCuts activates the sequential iteration described in the original
// (commented-out) network driver. Public node-level frontier APIs remain usable
// directly; this driver recomputes from the accumulated cut set for simplicity.
func Abc_NtkSeqCuts(network Network, params *Cut_Params_t) *Cut_Man_t {
	params.FSeq = 1
	p := Cut_ManStart(params)
	p.network, p.networkCounts = networkInventory(network)
	for _, node := range network.CutNodes() {
		if !node.Output {
			if node.Constant {
				cutSetConstant(p, node.ID, node.Value)
				p.oldCuts[node.ID] = p.newCuts[node.ID]
				delete(p.newCuts, node.ID)
			} else {
				p.oldCuts[node.ID] = Cut_CutCreateTriv(p, node.ID)
			}
		}
	}
	for iteration := int32(0); iteration < 10; iteration++ {
		next := map[int32]*Cut_Cut_t{}
		changed := false
		for _, node := range network.CutNodes() {
			if node.Input || node.Output || node.Constant {
				continue
			}
			oldCount := Cut_CutCountList(p.oldCuts[node.ID])
			if oldCount >= params.NKeepMax {
				continue
			}
			fanin1, latch1, phase1 := node.Fanins[1], node.Latches[1], node.Compl[1]
			if node.Buffer {
				fanin1, latch1, phase1 = node.Fanins[0], node.Latches[0], node.Compl[0]
			}
			a, b := Cut_CutDupList(p, p.oldCuts[node.Fanins[0]]), Cut_CutDupList(p, p.oldCuts[fanin1])
			Cut_NodeShiftCutLeaves(a, node.Latches[0])
			Cut_NodeShiftCutLeaves(b, latch1)
			var list Cut_List_t
			Cut_ListStart(&list)
			p.budget = params.NKeepMax - oldCount
			p.xor = node.XOR && !node.Buffer
			Cut_NodeDoComputeCuts(p, &list, node.ID, cutBool(node.Compl[0]), cutBool(phase1), a, b, 0, 0)
			p.xor = false
			Cut_CutRecycleList(p, a)
			Cut_CutRecycleList(p, b)
			var head *Cut_Cut_t
			tail := &head
			for c := Cut_ListFinish(&list); c != nil; {
				after := c.Next
				c.Next = nil
				if Cut_CutFilterOld(p, p.oldCuts[node.ID], c) == 0 {
					*tail = c
					tail = &c.Next
				}
				c = after
			}
			if head != nil {
				next[node.ID] = head
				changed = true
			}
		}
		p.budget = -1
		for node, cuts := range next {
			p.oldCuts[node] = Cut_CutMergeLists(p.oldCuts[node], cuts)
		}
		p.stats.SeqIterations = iteration + 1
		if !changed {
			p.stats.SeqConverged = true
			break
		}
	}
	for _, node := range network.CutNodes() {
		Cut_NodeOldTransferToNew(p, node.ID)
	}
	return p
}

// Constants retain a positive-size cut, as required by the source list buckets,
// while their truth and simulation represent the constant rather than its ID.
func cutSetConstant(p *Cut_Man_t, id int32, value bool) {
	c := Cut_CutCreateTriv(p, id)
	c.FSimul = value
	for i := range c.truth {
		c.truth[i] = 0
		if value {
			c.truth[i] = ^uint32(0)
		}
	}
	p.newCuts[id] = c
}
