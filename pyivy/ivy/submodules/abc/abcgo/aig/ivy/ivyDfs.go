// DFS and timing-level collection ported from ABC ivyDfs.c.
package ivy

func Ivy_ManDfs_rec(p *Ivy_Man_t, n *Ivy_Obj_t, nodes *[]int32) {
	n = Ivy_Regular(n)
	if n == nil || n.owner != p || n.recycled {
		panic("ivy: invalid DFS object")
	}
	if n.fMarkA != 0 {
		return
	}
	n.fMarkA = 1
	if Ivy_ObjIsConst1(n) != 0 || Ivy_ObjIsCi(n) != 0 {
		if p.pHaig == nil && n.pEquiv != nil && Ivy_Regular(n.pEquiv).owner == p {
			Ivy_ManDfs_rec(p, Ivy_Regular(n.pEquiv), nodes)
		}
		return
	}
	if Ivy_ObjIsNode(n) == 0 && Ivy_ObjIsBuf(n) == 0 {
		panic("ivy: invalid DFS gate")
	}
	Ivy_ManDfs_rec(p, Ivy_ObjFanin0(n), nodes)
	if Ivy_ObjIsNode(n) != 0 {
		Ivy_ManDfs_rec(p, Ivy_ObjFanin1(n), nodes)
	}
	if p.pHaig == nil && n.pEquiv != nil && Ivy_Regular(n.pEquiv).owner == p {
		Ivy_ManDfs_rec(p, Ivy_Regular(n.pEquiv), nodes)
	}
	*nodes = append(*nodes, n.Id)
}
func ivyDfsClear(p *Ivy_Man_t) {
	for _, n := range p.vObjs {
		if n != nil {
			n.fMarkA = 0
		}
	}
}
func ivyDfsCheckMarks(p *Ivy_Man_t) {
	for _, n := range p.vObjs {
		if n != nil && (n.fMarkA != 0 || n.fMarkB != 0) {
			panic("ivy: DFS requires unmarked graph")
		}
	}
}
func Ivy_ManDfs(p *Ivy_Man_t) []int32 {
	if Ivy_ManLatchNum(p) != 0 {
		panic("ivy: combinational DFS on sequential graph")
	}
	ivyDfsCheckMarks(p)
	defer ivyDfsClear(p)
	nodes := make([]int32, 0, Ivy_ManNodeNum(p))
	for _, n := range p.vPos {
		Ivy_ManDfs_rec(p, Ivy_ObjFanin0(n), &nodes)
	}
	if len(nodes) != int(Ivy_ManNodeNum(p)+Ivy_ManBufNum(p)) {
		panic("ivy: DFS graph contains dangling gates")
	}
	return nodes
}
func Ivy_ManDfsSeq(p *Ivy_Man_t, latches *[]int32) []int32 {
	ivyDfsCheckMarks(p)
	defer ivyDfsClear(p)
	nodes := make([]int32, 0, Ivy_ManNodeNum(p))
	var regs []int32
	for _, n := range p.vObjs {
		if n != nil && Ivy_ObjIsLatch(n) != 0 {
			regs = append(regs, n.Id)
		}
	}
	for _, n := range p.vPos {
		Ivy_ManDfs_rec(p, Ivy_ObjFanin0(n), &nodes)
	}
	for _, id := range regs {
		Ivy_ManDfs_rec(p, Ivy_ObjFanin0(p.vObjs[id]), &nodes)
	}
	if latches != nil {
		*latches = regs
	}
	return nodes
}
func Ivy_ManCollectCone_rec(n *Ivy_Obj_t, cone *[]*Ivy_Obj_t) {
	n = Ivy_Regular(n)
	if n.fMarkA != 0 {
		return
	}
	if Ivy_ObjIsNode(n) == 0 && Ivy_ObjIsBuf(n) == 0 {
		panic("ivy: cone frontier does not cut every input")
	}
	for _, old := range *cone {
		if old == n {
			return
		}
	}
	Ivy_ManCollectCone_rec(Ivy_ObjFanin0(n), cone)
	if Ivy_ObjIsNode(n) != 0 {
		Ivy_ManCollectCone_rec(Ivy_ObjFanin1(n), cone)
	}
	*cone = append(*cone, n)
}
func Ivy_ManCollectCone(n *Ivy_Obj_t, front []*Ivy_Obj_t, cone *[]*Ivy_Obj_t) {
	if Ivy_IsComplement(n) != 0 || Ivy_ObjIsNode(n) == 0 {
		panic("ivy: invalid cone root")
	}
	marks := map[*Ivy_Obj_t]int32{}
	for _, edge := range front {
		regular := Ivy_Regular(edge)
		if _, seen := marks[regular]; !seen {
			marks[regular] = regular.fMarkA
		}
		regular.fMarkA = 1
	}
	defer func() {
		for regular, mark := range marks {
			regular.fMarkA = mark
		}
	}()
	if n.fMarkA != 0 {
		panic("ivy: cone root belongs to frontier")
	}
	*cone = (*cone)[:0]
	Ivy_ManCollectCone_rec(n, cone)
}
func Ivy_ManLevelize(p *Ivy_Man_t) [][]*Ivy_Obj_t {
	var levels [][]*Ivy_Obj_t
	for _, n := range p.vObjs {
		if n == nil {
			continue
		}
		if Ivy_ObjIsBuf(n) != 0 {
			panic("ivy: levelize requires no buffers")
		}
		if Ivy_ObjIsNode(n) == 0 {
			continue
		}
		for len(levels) <= int(n.Level) {
			levels = append(levels, nil)
		}
		levels[n.Level] = append(levels[n.Level], n)
	}
	return levels
}
func Ivy_ManRequiredLevels(p *Ivy_Man_t) []int32 {
	if p.vRequired != nil {
		panic("ivy: required levels already assigned")
	}
	required := make([]int32, len(p.vObjs)+1)
	levels := Ivy_ManLevelize(p)
	for i := len(levels) - 1; i >= 0; i-- {
		for _, n := range levels[i] {
			level := required[n.Id] + 1 + Ivy_ObjIsExor(n)
			for _, fanin := range []*Ivy_Obj_t{Ivy_ObjFanin0(n), Ivy_ObjFanin1(n)} {
				required[fanin.Id] = max(required[fanin.Id], level)
			}
		}
	}
	maximum := Ivy_ManLevels(p)
	for _, n := range p.vObjs {
		if n != nil {
			required[n.Id] = maximum - required[n.Id]
		}
	}
	p.vRequired = required
	return required
}
func Ivy_ManSetLevels_rec(n *Ivy_Obj_t, history int32) int32 {
	n = Ivy_Regular(n)
	if n.fMarkA != 0 {
		return n.Level
	}
	n.fMarkA = 1
	if Ivy_ObjIsConst1(n) != 0 || Ivy_ObjIsCi(n) != 0 {
		n.Level = 0
		return 0
	}
	if Ivy_ObjIsNode(n) == 0 && Ivy_ObjIsBuf(n) == 0 {
		panic("ivy: invalid level gate")
	}
	Ivy_ManSetLevels_rec(Ivy_ObjFanin0(n), history)
	if Ivy_ObjIsNode(n) != 0 {
		Ivy_ManSetLevels_rec(Ivy_ObjFanin1(n), history)
		n.Level = Ivy_ObjLevelNew(n)
	} else {
		n.Level = 1 + Ivy_ObjFanin0(n).Level
	}
	if history != 0 && n.pEquiv != nil && Ivy_Regular(n.pEquiv).owner == n.owner && n.nRefs > 0 {
		members := []*Ivy_Obj_t{n}
		seen := map[*Ivy_Obj_t]bool{n: true}
		level := n.Level
		for other := n.pEquiv; other != n; other = Ivy_Regular(other.pEquiv) {
			if other == nil || seen[other] {
				panic("ivy: malformed choice ring")
			}
			seen[other] = true
			members = append(members, other)
			level = max(level, Ivy_ManSetLevels_rec(other, history))
		}
		for _, member := range members {
			member.Level = level
		}
	}
	return n.Level
}
func Ivy_ManSetLevels(p *Ivy_Man_t, history int32) int32 {
	if Ivy_ManIsAcyclic(p) == 0 {
		panic("ivy: cannot level cyclic graph")
	}
	for _, n := range p.vObjs {
		if n != nil {
			n.Level = 0
		}
	}
	defer ivyDfsClear(p)
	var maximum int32
	for _, n := range p.vObjs {
		if n == nil {
			continue
		}
		if Ivy_ObjIsCo(n) != 0 {
			maximum = max(maximum, Ivy_ManSetLevels_rec(Ivy_ObjFanin0(n), history))
		} else if (Ivy_ObjIsNode(n) != 0 || Ivy_ObjIsBuf(n) != 0) && n.nRefs == 0 {
			maximum = max(maximum, Ivy_ManSetLevels_rec(n, history))
		}
	}
	return maximum
}

func Ivy_ManLevels(p *Ivy_Man_t) int32 {
	var level int32
	for _, n := range p.vPos {
		level = max(level, Ivy_ObjFanin0(n).Level)
	}
	return level
}
