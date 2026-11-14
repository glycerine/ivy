package gia

// SetGateClass updates the native gate-level abstraction map. Positive classes
// include an object; zero leaves it as an abstract boundary when referenced.
func (p *Gia_Man_t) SetGateClass(id, value int32) {
	if id < 0 || id >= p.nObjs || value < 0 {
		panic("gia: invalid gate abstraction class")
	}
	if p.VGateClasses == nil {
		p.VGateClasses = Vec_IntStart(int(p.nObjs))
	}
	if Vec_IntSize(p.VGateClasses) != int(p.nObjs) {
		panic("gia: abstraction map size mismatch")
	}
	Vec_IntWriteEntry(p.VGateClasses, int(id), value)
}

// AbstractionPPIs follows Gia_GlaCollectAssigned/Gia_ManGlaCollect: included
// gates and flops contribute their immediate fanins, then ordinary PIs are
// removed. Results are unique object IDs in increasing order, not CO drivers.
func (p *Gia_Man_t) AbstractionPPIs() []int32 {
	if Gia_ManPoNum(p) != 1 || p.VGateClasses == nil || Vec_IntSize(p.VGateClasses) != int(p.nObjs) {
		panic("gia: invalid gate abstraction interface")
	}
	assigned := make([]bool, p.nObjs)
	for id, class := range Vec_IntArray(p.VGateClasses) {
		if class == 0 {
			continue
		}
		if class < 0 {
			panic("gia: negative gate abstraction class")
		}
		obj := p.pObjs[id]
		assigned[id] = true
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			for _, fanin := range giaObjectFanins(p, obj) {
				assigned[fanin] = true
			}
		case Gia_ObjIsRo(p, obj) != 0:
			assigned[Gia_ObjFaninId0p(p, Gia_ObjRoToRi(p, obj))] = true
		case Gia_ObjIsConst0(obj) != 0:
		default:
			panic("gia: abstraction class must name constant, register or gate")
		}
	}
	var ppis []int32
	for id, present := range assigned {
		if present && Gia_ObjIsPi(p, p.pObjs[id]) == 0 && Vec_IntEntry(p.VGateClasses, id) == 0 {
			ppis = append(ppis, int32(id))
		}
	}
	return ppis
}
