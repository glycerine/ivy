// Timeframe unrolling ported from ABC ivyMan.c.
package ivy

// Ivy_ManFrames unrolls a combinational next-state network whose last inputs
// and outputs represent registers. Mapping rows use len(source.vObjs) slots.
func Ivy_ManFrames(source *Ivy_Man_t, registers, frames, initialize int32, mapping *[]*Ivy_Obj_t) *Ivy_Man_t {
	if Ivy_ManLatchNum(source) != 0 || registers < 0 || registers > int32(len(source.vPis)) || registers > int32(len(source.vPos)) || frames <= 0 {
		panic("ivy: invalid timeframe request")
	}
	nodes := Ivy_ManDfs(source)
	inputs, outputs := len(source.vPis)-int(registers), len(source.vPos)-int(registers)
	stride := len(source.vObjs)
	if uint64(stride)*uint64(frames) > uint64(^uint(0)>>1) {
		panic("ivy: timeframe mapping too large")
	}
	result := Ivy_ManStart()
	state := make([]*Ivy_Obj_t, registers)
	for i := range state {
		if initialize != 0 {
			state[i] = Ivy_ManConst0(result)
		} else {
			state[i] = Ivy_ObjCreatePi(result)
		}
	}
	saved := make([]*Ivy_Obj_t, stride*int(frames))
	for frame := 0; frame < int(frames); frame++ {
		source.pConst1.pEquiv = result.pConst1
		for _, n := range source.vPis[:inputs] {
			n.pEquiv = Ivy_ObjCreatePi(result)
		}
		for i, n := range source.vPis[inputs:] {
			n.pEquiv = state[i]
		}
		for _, id := range nodes {
			n := source.vObjs[id]
			if Ivy_ObjIsBuf(n) != 0 {
				n.pEquiv = Ivy_ObjChild0Equiv(n)
			} else {
				n.pEquiv = Ivy_Oper(result, Ivy_ObjChild0Equiv(n), Ivy_ObjChild1Equiv(n), n.Type)
			}
		}
		for _, n := range source.vPos[:outputs] {
			n.pEquiv = Ivy_ObjCreatePo(result, Ivy_ObjChild0Equiv(n))
		}
		for i, n := range source.vPos[outputs:] {
			state[i] = Ivy_ObjChild0Equiv(n)
			n.pEquiv = state[i]
		}
		for i, n := range source.vObjs {
			if n != nil {
				saved[frame*stride+i] = n.pEquiv
			}
		}
	}
	if initialize == 0 {
		for _, n := range state {
			Ivy_ObjCreatePo(result, n)
		}
	}
	Ivy_ManCleanup(result)
	// Cleanup may delete mapped dangling cones; do not expose recycled nodes.
	for i, edge := range saved {
		if edge != nil && Ivy_Regular(edge).recycled {
			saved[i] = nil
		}
	}
	if mapping != nil {
		*mapping = saved
	}
	if Ivy_ManCheck(result) == 0 {
		panic("ivy: unrolled graph failed validation")
	}
	return result
}
