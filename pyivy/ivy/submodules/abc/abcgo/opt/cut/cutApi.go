package cut

func Cut_NodeReadCutsNew(p *Cut_Man_t, node int32) *Cut_Cut_t  { return p.newCuts[node] }
func Cut_NodeReadCutsOld(p *Cut_Man_t, node int32) *Cut_Cut_t  { return p.oldCuts[node] }
func Cut_NodeReadCutsTemp(p *Cut_Man_t, node int32) *Cut_Cut_t { return p.tempCuts[node] }
func Cut_NodeWriteCutsNew(p *Cut_Man_t, node int32, list *Cut_Cut_t) {
	p.check()
	p.newCuts[node] = list
}
func Cut_NodeWriteCutsOld(p *Cut_Man_t, node int32, list *Cut_Cut_t) {
	p.check()
	p.oldCuts[node] = list
}
func Cut_NodeWriteCutsTemp(p *Cut_Man_t, node int32, list *Cut_Cut_t) {
	p.check()
	p.tempCuts[node] = list
}
func Cut_NodeSetTriv(p *Cut_Man_t, node int32) {
	p.check()
	if p.newCuts[node] != nil {
		panic("cut: node already initialized")
	}
	p.newCuts[node] = Cut_CutCreateTriv(p, node)
}
func Cut_NodeTryDroppingCuts(p *Cut_Man_t, node int32) {
	p.check()
	if node < 0 || int(node) >= len(p.fanCounts) || p.fanCounts[node] <= 0 {
		panic("cut: invalid fanout count")
	}
	p.fanCounts[node]--
	if p.fanCounts[node] == 0 {
		Cut_NodeFreeCuts(p, node)
	}
}
func Cut_NodeFreeCuts(p *Cut_Man_t, node int32) {
	Cut_CutRecycleList(p, p.newCuts[node])
	delete(p.newCuts, node)
}
