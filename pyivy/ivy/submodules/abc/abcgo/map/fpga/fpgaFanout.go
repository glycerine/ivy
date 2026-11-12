package fpga

func Fpga_NodeAddFaninFanout(fanin, node *Fpga_Node_t) {
	n := Fpga_Regular(fanin)
	n.fanouts = append(n.fanouts, Fpga_Regular(node))
}
func Fpga_NodeRemoveFaninFanout(fanin, node *Fpga_Node_t) {
	n := Fpga_Regular(fanin)
	for i, v := range n.fanouts {
		if v == Fpga_Regular(node) {
			n.fanouts = append(n.fanouts[:i], n.fanouts[i+1:]...)
			return
		}
	}
	panic("fpga: fanout not present")
}
func Fpga_NodeGetFanoutNum(n *Fpga_Node_t) int32              { return int32(len(Fpga_Regular(n).fanouts)) }
func Fpga_NodeGetFanout(n *Fpga_Node_t, i int32) *Fpga_Node_t { return Fpga_Regular(n).fanouts[i] }
