package fpga

import "sort"

// Assign late signals to fast pins; truth-table leaf order stays unchanged.
func Fpga_TimeCutComputeArrival(p *Fpga_Man_t, c *Fpga_Cut_t) float32 {
	if len(c.Leaves) == 0 {
		c.Arrival = 0
		return 0
	}
	k := len(c.Leaves)
	delays := p.Library.Delays[k]
	c.PinOrder = make([]int, k)
	for i := range c.PinOrder {
		c.PinOrder[i] = i
	}
	pins := make([]int, k)
	for i := range pins {
		pins[i] = i
	}
	if len(delays) > 1 {
		sort.SliceStable(c.PinOrder, func(i, j int) bool {
			return c.Leaves[c.PinOrder[i]].Best.Arrival > c.Leaves[c.PinOrder[j]].Best.Arrival
		})
		sort.SliceStable(pins, func(i, j int) bool { return delays[pins[i]] < delays[pins[j]] })
	}
	order := make([]int, k)
	arrival := float32(-FPGA_FLOAT_LARGE)
	for rank, leafIndex := range c.PinOrder {
		pin := pins[rank]
		order[pin] = leafIndex
		d := delays[0]
		if len(delays) > 1 {
			d = delays[pin]
		}
		v := c.Leaves[leafIndex].Best.Arrival + d
		if v > arrival {
			arrival = v
		}
	}
	c.PinOrder = order
	c.Arrival = arrival
	return arrival
}
func Fpga_TimeCutComputeArrival_rec(p *Fpga_Man_t, c *Fpga_Cut_t) float32 {
	seen := map[*Fpga_Node_t]bool{}
	var update func(*Fpga_Cut_t)
	update = func(c *Fpga_Cut_t) {
		for _, leaf := range c.Leaves {
			if leaf.Refs == 0 && Fpga_NodeIsAnd(leaf) != 0 && !seen[leaf] {
				seen[leaf] = true
				update(leaf.Best)
			}
		}
		Fpga_TimeCutComputeArrival(p, c)
	}
	update(c)
	return c.Arrival
}
func timingOutputs(p *Fpga_Man_t) []*Fpga_Node_t {
	if p.LatchPaths && p.Latches > 0 {
		return p.Outputs[len(p.Outputs)-int(p.Latches):]
	}
	return p.Outputs
}
func Fpga_TimeComputeArrivalMax(p *Fpga_Man_t) float32 {
	result := float32(-FPGA_FLOAT_LARGE)
	found := false
	for _, edge := range timingOutputs(p) {
		n := Fpga_Regular(edge)
		if n != nil && Fpga_NodeIsConst(n) == 0 && n.Best != nil && n.Best.Arrival > result {
			found = true
			result = n.Best.Arrival
		}
	}
	if !found {
		return 0
	}
	return result
}
func Fpga_TimeComputeRequiredGlobal(p *Fpga_Man_t, first int32) {
	required := Fpga_TimeComputeArrivalMax(p)
	if p.DelayTarget > required {
		required = p.DelayTarget
	}
	Fpga_TimeComputeRequired(p, required)
}
func Fpga_TimeComputeRequired(p *Fpga_Man_t, required float32) {
	p.RequiredGlobal = required
	for _, n := range p.Nodes {
		n.Required = FPGA_FLOAT_LARGE
	}
	p.Const1.Required = FPGA_FLOAT_LARGE
	for _, edge := range timingOutputs(p) {
		n := Fpga_Regular(edge)
		n.Required = required
	}
	Fpga_TimePropagateRequired(p, p.Mapping)
}
func Fpga_TimePropagateRequired(p *Fpga_Man_t, nodes *Fpga_NodeVec_t) {
	for _, n := range nodes.Nodes {
		c := n.Best
		delays := p.Library.Delays[len(c.Leaves)]
		for pin, leafIndex := range c.PinOrder {
			leaf := c.Leaves[leafIndex]
			d := delays[0]
			if len(delays) > 1 {
				d = delays[pin]
			}
			r := n.Required - d
			if r < leaf.Required {
				leaf.Required = r
			}
		}
	}
}
func Fpga_TimePropagateArrival(p *Fpga_Man_t) {
	for i, n := range p.Inputs {
		n.Best.Arrival = p.InputArrivals[i]
	}
	for _, n := range p.Ands.Nodes {
		if n.Best != nil && Fpga_NodeIsAnd(n) != 0 {
			Fpga_TimeCutComputeArrival(p, n.Best)
		}
	}
}
