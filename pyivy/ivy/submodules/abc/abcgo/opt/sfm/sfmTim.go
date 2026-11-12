package sfm

import "sort"

// Sfm_Tim_t is the source block-delay timing engine. Transition zero is rise,
// transition one fall. Values use the same integer units as Gate.Pins.
type Sfm_Tim_t struct {
	Network           *Sfm_Ntk_t
	Arrival, Required [][2]int
	Delay, DeltaCrit  int
	Path              []int
}

const timingInfinity = int(1 << 30)

func Sfm_TimStart(n *Sfm_Ntk_t, delta int) *Sfm_Tim_t {
	if delta <= 0 {
		panic("sfm: critical timing delta must be positive")
	}
	p := &Sfm_Tim_t{Network: n, DeltaCrit: delta}
	Sfm_TimTrace(p)
	return p
}
func Sfm_TimStop(p *Sfm_Tim_t) {
	if p != nil {
		p.Arrival = nil
		p.Required = nil
		p.Path = nil
	}
}
func Sfm_TimArrId(p *Sfm_Tim_t, id int) *[2]int { return &p.Arrival[id] }
func Sfm_TimReqId(p *Sfm_Tim_t, id int) *[2]int { return &p.Required[id] }
func Sfm_TimArrMaxId(p *Sfm_Tim_t, id int) int  { return max(p.Arrival[id][0], p.Arrival[id][1]) }
func Sfm_TimArr(p *Sfm_Tim_t, id int) *[2]int   { return Sfm_TimArrId(p, id) }
func Sfm_TimReq(p *Sfm_Tim_t, id int) *[2]int   { return Sfm_TimReqId(p, id) }
func Sfm_TimArrMax(p *Sfm_Tim_t, id int) int    { return Sfm_TimArrMaxId(p, id) }
func Sfm_TimSetReq(p *Sfm_Tim_t, id, t int)     { p.Required[id] = [2]int{t, t} }
func Sfm_TimSlack(p *Sfm_Tim_t, id int) int {
	return min(p.Required[id][0]-p.Arrival[id][0], p.Required[id][1]-p.Arrival[id][1])
}
func Sfm_TimEdgeArrival(pin Pin, in [2]int, out *[2]int) {
	if pin.Phase != 1 {
		out[0] = max(out[0], in[0]+pin.Rise)
		out[1] = max(out[1], in[1]+pin.Fall)
	}
	if pin.Phase != 2 {
		out[0] = max(out[0], in[1]+pin.Rise)
		out[1] = max(out[1], in[0]+pin.Fall)
	}
}
func Sfm_TimGateArrival(g *Gate, in [][2]int) (out [2]int) {
	if g == nil || len(g.Pins) != len(in) || g.Inputs != len(in) {
		panic("sfm: invalid timed gate pin count")
	}
	for i, pin := range g.Pins {
		Sfm_TimEdgeArrival(pin, in[i], &out)
	}
	return
}
func Sfm_TimEdgeRequired(pin Pin, in *[2]int, out [2]int) {
	if pin.Phase != 1 {
		in[0] = min(in[0], out[0]-pin.Rise)
		in[1] = min(in[1], out[1]-pin.Fall)
	}
	if pin.Phase != 2 {
		in[0] = min(in[0], out[1]-pin.Rise)
		in[1] = min(in[1], out[0]-pin.Fall)
	}
}
func Sfm_TimGateRequired(g *Gate, in []*[2]int, out [2]int) {
	if g == nil || len(g.Pins) != len(in) || g.Inputs != len(in) {
		panic("sfm: invalid timed gate pin count")
	}
	for i, pin := range g.Pins {
		Sfm_TimEdgeRequired(pin, in[i], out)
	}
}
func Sfm_TimNodeArrival(p *Sfm_Tim_t, id int) {
	in := make([][2]int, len(p.Network.Fanins[id]))
	for i, f := range p.Network.Fanins[id] {
		in[i] = p.Arrival[f]
	}
	p.Arrival[id] = Sfm_TimGateArrival(p.Network.Mapping[id], in)
}
func Sfm_TimNodeRequired(p *Sfm_Tim_t, id int) {
	in := make([]*[2]int, len(p.Network.Fanins[id]))
	for i, f := range p.Network.Fanins[id] {
		in[i] = &p.Required[f]
	}
	Sfm_TimGateRequired(p.Network.Mapping[id], in, p.Required[id])
}
func Sfm_TimTrace(p *Sfm_Tim_t) int {
	n := p.Network
	order, err := n.topology()
	if err != nil {
		panic(err)
	}
	p.Arrival = make([][2]int, len(n.Fanins))
	p.Required = make([][2]int, len(n.Fanins))
	p.Delay = 0
	for _, id := range order {
		if Sfm_ObjIsNode(n, id) && !n.Deleted[id] {
			Sfm_TimNodeArrival(p, id)
		}
	}
	for id := len(n.Fanins) - n.Pos; id < len(n.Fanins); id++ {
		f := n.Fanins[id][0]
		p.Arrival[id] = p.Arrival[f]
		p.Delay = max(p.Delay, Sfm_TimArrMaxId(p, f))
	}
	for i := range p.Required {
		p.Required[i] = [2]int{timingInfinity, timingInfinity}
	}
	for id := len(n.Fanins) - n.Pos; id < len(n.Fanins); id++ {
		Sfm_TimSetReq(p, n.Fanins[id][0], p.Delay)
		Sfm_TimSetReq(p, id, p.Delay)
	}
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		if Sfm_ObjIsNode(n, id) && !n.Deleted[id] {
			Sfm_TimNodeRequired(p, id)
		}
	}
	return p.Delay
}
func Sfm_TimReadNtkDelay(p *Sfm_Tim_t) int         { return p.Delay }
func Sfm_TimReadObjDelay(p *Sfm_Tim_t, id int) int { return Sfm_TimArrMaxId(p, id) }
func Sfm_TimCriticalPath(p *Sfm_Tim_t, window int) int {
	if window < 0 || window > 100 {
		panic("sfm: timing window outside 0..100")
	}
	p.Path = nil
	seen := make([]bool, len(p.Network.Fanins))
	limit := p.Delay * window / 100
	var visit func(int)
	visit = func(id int) {
		if seen[id] || !Sfm_ObjIsNode(p.Network, id) || len(p.Network.Fanins[id]) == 0 {
			return
		}
		seen[id] = true
		for _, f := range p.Network.Fanins[id] {
			if Sfm_TimSlack(p, f) <= limit {
				visit(f)
			}
		}
		p.Path = append(p.Path, id)
	}
	for id := len(p.Network.Fanins) - p.Network.Pos; id < len(p.Network.Fanins); id++ {
		f := p.Network.Fanins[id][0]
		if Sfm_TimSlack(p, f) <= limit {
			visit(f)
		}
	}
	return len(p.Path)
}
func Sfm_TimUpdateTiming(p *Sfm_Tim_t, nodes []int) {
	if len(nodes) == 0 || len(nodes) > 2 {
		panic("sfm: timing update requires one or two nodes")
	}
	Sfm_TimTrace(p)
}
func Sfm_TimSortArrayByArrival(p *Sfm_Tim_t, nodes []int, pivot int) int {
	// Source packed sort orders low arrival first, then object ID.
	sort.Slice(nodes, func(i, j int) bool {
		a, b := Sfm_TimArrMaxId(p, nodes[i]), Sfm_TimArrMaxId(p, nodes[j])
		if a != b {
			return a < b
		}
		return nodes[i] < nodes[j]
	})
	for i, id := range nodes {
		if Sfm_TimArrMaxId(p, id)+p.DeltaCrit > Sfm_TimArrMaxId(p, pivot) {
			return i
		}
	}
	return -1
}
func Sfm_TimPriorityNodes(p *Sfm_Tim_t, window int) []int {
	Sfm_TimCriticalPath(p, window)
	out := make([]int, 0, len(p.Path))
	for _, id := range p.Path {
		if !p.Network.Fixed[id] {
			out = append(out, id)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return p.Network.Levels[out[i]] < p.Network.Levels[out[j]] })
	return out
}
func Sfm_TimNodeIsNonCritical(p *Sfm_Tim_t, pivot, node int) bool {
	return Sfm_TimArrMaxId(p, node)+p.DeltaCrit <= Sfm_TimArrMaxId(p, pivot)
}
func Sfm_TimEvalRemapping(p *Sfm_Tim_t, fanins, mapping []int, first *Gate, fans1 []int, second *Gate, fans2 []int) int {
	inputs := func(fans []int, previous *[2]int) [][2]int {
		in := make([][2]int, len(fans))
		for i, f := range fans {
			if f == 16 {
				if previous == nil {
					panic("sfm: invalid cascade reference")
				}
				in[i] = *previous
			} else {
				in[i] = p.Arrival[mapping[fanins[f]]]
			}
		}
		return in
	}
	a := Sfm_TimGateArrival(first, inputs(fans1, nil))
	if second != nil {
		a = Sfm_TimGateArrival(second, inputs(fans2, &a))
	}
	return max(a[0], a[1])
}
