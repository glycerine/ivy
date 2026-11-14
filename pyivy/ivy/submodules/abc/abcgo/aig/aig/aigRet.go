package aig

import "fmt"

// Retiming stores delay and initial state on edges, independently of graph IDs.
func Rtm_ManAlloc(p *Graph) *Rtm_Man_t {
	return &Rtm_Man_t{source: p, sourceMap: make(map[int]*Rtm_Obj_t)}
}
func Rtm_ManFree(p *Rtm_Man_t) { *p = Rtm_Man_t{} }
func Rtm_ManLatchMax(p *Rtm_Man_t) int32 {
	var n int32
	for _, o := range p.vObjs {
		for _, e := range o.edges {
			n = max(n, e.nLats)
		}
	}
	return n
}
func Rtm_ObjAlloc(p *Rtm_Man_t, ni, no int32) *Rtm_Obj_t {
	o := &Rtm_Obj_t{Id: int32(len(p.vObjs)), Num: ni, Temp: no, Type: ni == 1 && no == 0, pCopy: InvalidEdge, Kind: And}
	p.vObjs = append(p.vObjs, o)
	return o
}
func Rtm_ObjAddFanin(o, in *Rtm_Obj_t, complement int32) {
	if len(o.fanins) >= 2 || int32(len(o.fanins)) >= o.Num || int32(len(in.fanouts)) >= in.Temp {
		panic("aig: retiming connection exceeds construction capacity")
	}
	e := &Rtm_Edg_t{}
	o.fanins = append(o.fanins, in)
	o.edges = append(o.edges, e)
	o.complements = append(o.complements, complement != 0)
	in.fanouts = append(in.fanouts, o)
	in.outEdges = append(in.outEdges, e)
}
func Rtm_ObjCheckRetimeFwd(o *Rtm_Obj_t) int32 {
	for _, e := range o.edges {
		if e.nLats == 0 {
			return 0
		}
	}
	return 1
}
func Rtm_ObjCheckRetimeBwd(o *Rtm_Obj_t) int32 {
	for _, e := range o.outEdges {
		if e.nLats == 0 {
			return 0
		}
	}
	return 1
}
func Rtm_ObjGetDegreeFwd(o *Rtm_Obj_t) int32 {
	var n int32
	for _, f := range o.fanins {
		n = max(n, f.Num)
	}
	return n + 1
}
func Rtm_ObjGetDegreeBwd(o *Rtm_Obj_t) int32 {
	var n int32
	for _, f := range o.fanouts {
		n = max(n, f.Num)
	}
	return n + 1
}
func Rtm_ObjRetimeFwd(p *Rtm_Man_t, o *Rtm_Obj_t) {
	if Rtm_ObjCheckRetimeFwd(o) == 0 {
		panic("aig: forward retiming without input latches")
	}
	v := RTM_VAL_ONE
	for i, e := range o.edges {
		x := Rtm_ObjRemFirst(p, e)
		if o.complements[i] {
			x = Rtm_InitNot(x)
		}
		if i == 0 {
			v = x
		} else if o.Kind == ExclusiveOr {
			if v == RTM_VAL_VOID || x == RTM_VAL_VOID {
				v = RTM_VAL_VOID
			} else if v == x {
				v = RTM_VAL_ZERO
			} else {
				v = RTM_VAL_ONE
			}
		} else {
			v = Rtm_InitAnd(v, x)
		}
	}
	for _, e := range o.outEdges {
		Rtm_ObjAddLast(p, e, v)
	}
}
func Rtm_ObjRetimeBwd(p *Rtm_Man_t, o *Rtm_Obj_t) {
	if Rtm_ObjCheckRetimeBwd(o) == 0 {
		panic("aig: backward retiming without output latches")
	}
	for _, e := range o.outEdges {
		Rtm_ObjRemLast(p, e)
	}
	for _, e := range o.edges {
		Rtm_ObjAddFirst(p, e, RTM_VAL_VOID)
	}
}
func Rtm_ObjMarkAutoFwd_rec(o *Rtm_Obj_t) {
	if o.fAuto {
		return
	}
	o.fAuto = true
	for _, f := range o.fanouts {
		Rtm_ObjMarkAutoFwd_rec(f)
	}
}
func Rtm_ObjMarkAutoBwd_rec(o *Rtm_Obj_t) {
	if o.fAuto {
		return
	}
	o.fAuto = true
	for _, f := range o.fanins {
		Rtm_ObjMarkAutoBwd_rec(f)
	}
}
func Rtm_ManMarkAutoFwd(p *Rtm_Man_t) int32 {
	for _, o := range p.vObjs {
		o.fAuto = false
	}
	Rtm_ObjMarkAutoFwd_rec(p.vObjs[0])
	for _, o := range p.vPis {
		Rtm_ObjMarkAutoFwd_rec(o)
	}
	var n int32
	for _, o := range p.vObjs {
		o.fAuto = !o.fAuto
		if o.fAuto {
			n++
		}
	}
	return n
}
func Rtm_ManMarkAutoBwd(p *Rtm_Man_t) int32 {
	for _, o := range p.vObjs {
		o.fAuto = false
	}
	p.vObjs[0].fAuto = true
	for _, o := range p.vPis {
		o.fAuto = true
	}
	for _, o := range p.vPos {
		Rtm_ObjMarkAutoBwd_rec(o)
	}
	var n int32
	for _, o := range p.vObjs {
		o.fAuto = !o.fAuto
		if o.fAuto {
			n++
		}
	}
	return n
}
func Rtm_ManFromAig(g *Graph) *Rtm_Man_t {
	if err := g.Check(); err != nil {
		panic(err)
	}
	p := Rtm_ManAlloc(g)
	npi := len(g.inputs) - int(g.nRegs)
	npo := len(g.outputs) - int(g.nRegs)
	allocate := func(e Edge, ni, no int32) *Rtm_Obj_t {
		o := Rtm_ObjAlloc(p, ni, no)
		o.Kind = g.node(e).Kind
		p.sourceMap[e.NodeID()] = o
		return o
	}
	allocate(g.Const1(), 0, int32(g.node(g.Const1()).Refs))
	for _, e := range g.inputs[:npi] {
		p.vPis = append(p.vPis, allocate(e, 0, int32(g.node(e).Refs)))
	}
	for _, e := range g.outputs[:npo] {
		p.vPos = append(p.vPos, allocate(e, 1, 0))
	}
	for _, e := range g.inputs[npi:] {
		p.latchObjs = append(p.latchObjs, allocate(e, 1, int32(g.node(e).Refs)))
	}
	for _, e := range g.outputs[npo:] {
		allocate(e, 1, 1)
	}
	for id, n := range g.nodes {
		if !g.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			allocate(Edge(id)<<1, int32(len(nodeFanins(n))), int32(n.Refs))
		}
	}
	connect := func(e, f Edge) {
		Rtm_ObjAddFanin(p.sourceMap[e.NodeID()], p.sourceMap[f.NodeID()], Aig_IsComplement(f))
	}
	for _, e := range g.outputs {
		connect(e, g.node(e).Fanins[0])
	}
	for i, e := range g.inputs[npi:] {
		connect(e, g.outputs[npo+i])
	}
	for id, n := range g.nodes {
		if !g.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr || n.Kind == Buffer) {
			for _, f := range nodeFanins(n) {
				connect(Edge(id)<<1, f)
			}
		}
	}
	return p
}
func Rtm_ManToAig_rec(g *Graph, p *Rtm_Man_t, o *Rtm_Obj_t, offsets []int32) Edge {
	if o.pCopy != InvalidEdge {
		return o.pCopy
	}
	if o.fMark {
		panic("aig: combinational cycle after retiming")
	}
	o.fMark = true
	defer func() { o.fMark = false }()
	v := g.Const1()
	for i, e := range o.edges {
		var f Edge
		if e.nLats > 0 {
			f = g.inputs[offsets[2*o.Id+int32(i)]+e.nLats-1]
			if Rtm_ObjGetFirst(p, e) == RTM_VAL_ONE {
				f = f.Not()
			}
		} else {
			f = Rtm_ManToAig_rec(g, p, o.fanins[i], offsets)
		}
		if o.complements[i] {
			f = f.Not()
		}
		if i == 0 {
			v = f
		} else if o.Kind == ExclusiveOr {
			v = g.NewXor(v, f)
		} else {
			v = g.AND(v, f)
		}
	}
	o.pCopy = v
	return v
}
func Rtm_ManToAig(p *Rtm_Man_t) *Graph {
	g := NewGraph()
	offsets := make([]int32, 2*len(p.vObjs))
	n := int32(len(p.vPis))
	for _, o := range p.vObjs {
		o.pCopy = InvalidEdge
		o.fMark = false
		for i, e := range o.edges {
			offsets[2*o.Id+int32(i)] = n
			n += e.nLats
		}
	}
	p.vObjs[0].pCopy = g.Const1()
	for _, o := range p.vPis {
		o.pCopy = g.NewInput()
	}
	for int32(len(g.inputs)) < n {
		g.NewInput()
	}
	for _, o := range p.vObjs {
		Rtm_ManToAig_rec(g, p, o, offsets)
	}
	for _, o := range p.vPos {
		g.AddOutput(o.pCopy)
	}
	for _, o := range p.vObjs {
		for i, e := range o.edges {
			v := o.fanins[i].pCopy
			for m := int32(0); m < e.nLats; m++ {
				one := Rtm_ObjGetOne(p, e, e.nLats-1-m) == RTM_VAL_ONE
				next := v
				if one {
					next = next.Not()
				}
				g.AddOutput(next)
				v = g.inputs[offsets[2*o.Id+int32(i)]+m]
				if one {
					v = v.Not()
				}
			}
		}
	}
	if p.source != nil {
		dupMetadata(g, p.source)
	}
	Aig_ManSetRegNum(g, n-int32(len(p.vPis)))
	Aig_ManCleanup(g)
	return g
}

// Rtm_ManRetime follows ABC's edge-retiming driver. Forward mode preserves
// zero-reset behavior. Backward mode discards initial values, inserts VOID
// latches, then reconstructs them as zero without solving reset constraints.
// Consequently backward mode need not preserve the original zero-reset trace.
func Rtm_ManRetime(g *Graph, forward, steps, verbose int32) *Graph {
	if g.nRegs == 0 {
		return Aig_ManDupOrdered(g)
	}
	if len(g.ClockDomains) > 1 {
		panic("aig: edge retiming requires a single clock domain")
	}
	if len(g.ClockDomains) == 1 {
		// The edge algorithm cannot track different domains through a move.
		// A partial list must not silently absorb previously unassigned latches.
		seen := make([]bool, g.nRegs)
		for _, reg := range g.ClockDomains[0] {
			if reg < 0 || reg >= g.nRegs || seen[reg] {
				panic("aig: invalid edge-retiming clock-domain membership")
			}
			seen[reg] = true
		}
		if len(g.ClockDomains[0]) != len(seen) {
			panic("aig: edge retiming requires every register in its clock domain")
		}
	}
	p := Rtm_ManFromAig(g)
	defer Rtm_ManFree(p)
	for _, o := range p.latchObjs {
		v := RTM_VAL_ZERO
		if forward == 0 {
			v = RTM_VAL_VOID
		}
		Rtm_ObjAddFirst(p, o.edges[0], v)
	}
	for _, o := range p.vObjs {
		if int32(len(o.fanins)) != o.Num || int32(len(o.fanouts)) != o.Temp {
			panic("aig: incomplete retiming construction")
		}
		o.Num = 0
	}
	var queue []*Rtm_Obj_t
	if forward != 0 {
		Rtm_ManMarkAutoFwd(p)
		queue = append(queue, p.latchObjs...)
	} else {
		Rtm_ManMarkAutoBwd(p)
		for _, o := range p.latchObjs {
			queue = append(queue, o.fanins[0])
		}
	}
	for _, o := range queue {
		o.fMark = true
	}
	moves := 0
	for head := 0; head < len(queue); head++ {
		o := queue[head]
		o.fMark = false
		if o.fAuto {
			continue
		}
		if forward != 0 {
			Rtm_ObjRetimeFwd(p, o)
		} else {
			Rtm_ObjRetimeBwd(p, o)
		}
		moves++
		neighbors := o.fanouts
		if forward == 0 {
			neighbors = o.fanins
		}
		for _, f := range neighbors {
			if f.fMark || f.fAuto || f.Type || len(f.fanins) == 0 {
				continue
			}
			degree := Rtm_ObjGetDegreeFwd(f)
			eligible := Rtm_ObjCheckRetimeFwd(f)
			if forward == 0 {
				degree = Rtm_ObjGetDegreeBwd(f)
				eligible = Rtm_ObjCheckRetimeBwd(f)
			}
			if eligible != 0 && degree <= steps {
				f.Num = degree
				f.fMark = true
				queue = append(queue, f)
			}
		}
	}
	out := Rtm_ManToAig(p)
	if moves > 0 {
		out.RetimingChanged = true
		out.FlopNumbers = nil
		out.FlopReprs = nil
		// Edge retiming replaces register identities. Only one-hot groups
		// consisting entirely of unchanged true PIs retain their meaning.
		out.OneHotGroups = nil
		var inputGroups [][]int32
		for _, group := range out.OneHotInputGroups {
			keep := true
			for _, input := range group {
				if input < 0 || int(input) >= len(p.vPis) {
					keep = false
					break
				}
			}
			if keep {
				inputGroups = append(inputGroups, group)
			}
		}
		out.OneHotInputGroups = inputGroups
		if len(out.ClockDomains) == 1 {
			out.ClockDomains[0] = make([]int32, out.nRegs)
			for i := range out.ClockDomains[0] {
				out.ClockDomains[0][i] = int32(i)
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Retiming: %d moves, %d registers\n", moves, out.nRegs)
	}
	return Aig_ManReduceLaches(out, verbose)
}
