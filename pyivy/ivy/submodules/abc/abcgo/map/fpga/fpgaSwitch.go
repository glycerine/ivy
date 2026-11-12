package fpga

func Fpga_CutGetSwitchDerefed(p *Fpga_Man_t, n *Fpga_Node_t, c *Fpga_Cut_t) float32 {
	cost := Fpga_CutRefSwitch(p, n, c, 0)
	Fpga_CutDerefSwitch(p, n, c, 0)
	return cost
}
func Fpga_CutRefSwitch(p *Fpga_Man_t, n *Fpga_Node_t, c *Fpga_Cut_t, f int32) float32 {
	cost := n.Switching
	for _, leaf := range c.Leaves {
		leaf.Refs++
		if leaf.Refs == 1 {
			if Fpga_NodeIsAnd(leaf) != 0 {
				cost += Fpga_CutRefSwitch(p, leaf, leaf.Best, f)
			} else {
				cost += leaf.Switching
			}
		}
	}
	return cost
}
func Fpga_CutDerefSwitch(p *Fpga_Man_t, n *Fpga_Node_t, c *Fpga_Cut_t, f int32) float32 {
	cost := n.Switching
	for _, leaf := range c.Leaves {
		if leaf.Refs <= 0 {
			panic("fpga: reference underflow")
		}
		leaf.Refs--
		if leaf.Refs == 0 {
			if Fpga_NodeIsAnd(leaf) != 0 {
				cost += Fpga_CutDerefSwitch(p, leaf, leaf.Best, f)
			} else {
				cost += leaf.Switching
			}
		}
	}
	return cost
}
func Fpga_MappingGetSwitching(p *Fpga_Man_t, nodes *Fpga_NodeVec_t) float32 {
	var total float32
	for _, n := range nodes.Nodes {
		total += n.Switching
	}
	for _, e := range p.Outputs {
		if Fpga_NodeIsVar(e) != 0 {
			total += Fpga_Regular(e).Switching
		}
	}
	return total
}
