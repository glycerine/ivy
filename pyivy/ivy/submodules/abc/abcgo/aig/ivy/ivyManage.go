// Manager duplication and sequential cleanup ported from ABC ivyMan.c.
package ivy

import "fmt"

func Ivy_ManDup(p *Ivy_Man_t) *Ivy_Man_t {
	var latches []int32
	nodes := Ivy_ManDfsSeq(p, &latches)
	result := Ivy_ManStartFrom(p)
	for _, id := range latches {
		p.vObjs[id].pEquiv = Ivy_ObjCreatePi(result)
	}
	for _, id := range nodes {
		n := p.vObjs[id]
		if Ivy_ObjIsBuf(n) != 0 {
			n.pEquiv = Ivy_ObjChild0Equiv(n)
		} else {
			n.pEquiv = Ivy_Oper(result, Ivy_ObjChild0Equiv(n), Ivy_ObjChild1Equiv(n), n.Type)
		}
	}
	for _, n := range p.vPos {
		Ivy_ObjCreatePo(result, Ivy_ObjChild0Equiv(n))
	}
	for _, id := range latches {
		old := p.vObjs[id]
		node := old.pEquiv
		if Ivy_ObjFaninC0(old) != 0 {
			panic("ivy: latch input not normalized")
		}
		node.Type, node.Init = IVY_LATCH, old.Init
		Ivy_ObjConnect(result, node, Ivy_ObjChild0Equiv(old), nil)
	}
	result.vPis = result.vPis[:len(p.vPis)]
	result.nObjs[IVY_PI] -= int32(len(latches))
	result.nObjs[IVY_LATCH] += int32(len(latches))
	if Ivy_ManCheck(result) == 0 {
		panic("ivy: duplicate graph failed validation")
	}
	return result
}

func Ivy_ManCleanupSeq_rec(n *Ivy_Obj_t) {
	n = Ivy_Regular(n)
	if n.fMarkA != 0 {
		return
	}
	n.fMarkA = 1
	if n.pFanin0 != nil {
		Ivy_ManCleanupSeq_rec(Ivy_ObjFanin0(n))
	}
	if n.pFanin1 != nil {
		Ivy_ManCleanupSeq_rec(Ivy_ObjFanin1(n))
	}
}
func Ivy_ManCleanupSeq(p *Ivy_Man_t) int32 {
	ivyDfsCheckMarks(p)
	p.pConst1.fMarkA = 1
	for _, n := range p.vPis {
		n.fMarkA = 1
	}
	for _, n := range p.vPos {
		Ivy_ManCleanupSeq_rec(n)
	}
	var dead []*Ivy_Obj_t
	for _, n := range p.vObjs {
		if n == nil {
			continue
		}
		if n.fMarkA != 0 {
			n.fMarkA = 0
		} else {
			dead = append(dead, n)
		}
	}
	// First disconnect the entire dead component, including feedback cycles.
	for _, n := range dead {
		Ivy_ObjDisconnect(p, n)
	}
	for _, n := range dead {
		if n.nRefs != 0 {
			panic("ivy: unreachable component still referenced")
		}
		p.nObjs[n.Type]--
		p.nDeleted++
		if n.Type == IVY_BUF {
			p.vBufs = ivyRemoveObject(p.vBufs, n)
		}
		p.vObjs[n.Id] = nil
		Ivy_ManRecycleMemory(p, n)
	}
	return int32(len(dead))
}
func Ivy_ManLatchIsSelfFeed_rec(n, root *Ivy_Obj_t) int32 {
	seen := map[*Ivy_Obj_t]bool{}
	for n != nil {
		n = Ivy_Regular(n)
		if Ivy_ObjIsLatch(n) == 0 && Ivy_ObjIsBuf(n) == 0 {
			return 0
		}
		if n == root {
			return 1
		}
		if seen[n] {
			return 0
		}
		seen[n] = true
		n = Ivy_ObjFanin0(n)
	}
	return 0
}
func Ivy_ManLatchIsSelfFeed(n *Ivy_Obj_t) int32 {
	if n == nil || Ivy_ObjIsLatch(n) == 0 {
		return 0
	}
	return Ivy_ManLatchIsSelfFeed_rec(Ivy_ObjFanin0(n), n)
}
func Ivy_ManPrintStats(p *Ivy_Man_t) {
	level := int32(-1)
	if Ivy_ManLatchNum(p) == 0 {
		level = Ivy_ManLevels(p)
	}
	fmt.Printf("PI/PO = %d/%d A = %7d. L = %5d. MaxID = %7d. Lev = %3d.\n", Ivy_ManPiNum(p), Ivy_ManPoNum(p), Ivy_ManAndNum(p), Ivy_ManLatchNum(p), Ivy_ManObjIdMax(p), level)
}

func Ivy_ManPropagateBuffers(p *Ivy_Man_t, update int32) int32 {
	if p.fFanout == 0 {
		panic("ivy: buffer propagation requires fanouts")
	}
	limit := max(int32(1), Ivy_ManNodeNum(p)) * 100
	var steps int32
	for len(p.vBufs) > 0 {
		n := p.vBufs[len(p.vBufs)-1]
		if n == nil || n.recycled || Ivy_ObjIsBuf(n) == 0 {
			p.vBufs = p.vBufs[:len(p.vBufs)-1]
			continue
		}
		if n.nRefs == 0 {
			Ivy_ObjDelete_rec(p, n, 1)
			continue
		}
		seen := map[*Ivy_Obj_t]bool{}
		for Ivy_ObjIsBuf(n) != 0 {
			if seen[n] {
				panic("ivy: cyclic buffer chain")
			}
			seen[n] = true
			n = Ivy_ObjReadFirstFanout(p, n)
			if n == nil {
				panic("ivy: referenced buffer has no fanout")
			}
		}
		if Ivy_ManLatchIsSelfFeed(n) != 0 {
			p.vBufs = p.vBufs[:len(p.vBufs)-1]
			continue
		}
		Ivy_NodeFixBufferFanins(p, n, update)
		steps++
		if steps > limit {
			break
		}
	}
	return steps
}
func Ivy_ManMakeSeq(p *Ivy_Man_t, registers int32, inits []int32) {
	if registers == 0 {
		return
	}
	if registers < 0 || registers >= int32(len(p.vPis)) || registers >= int32(len(p.vPos)) || len(p.vBufs) != 0 || (inits != nil && len(inits) < int(registers)) {
		panic("ivy: invalid sequential conversion")
	}
	for i := int32(0); inits != nil && i < registers; i++ {
		if Ivy_Init_t(inits[i]) < IVY_INIT_0 || Ivy_Init_t(inits[i]) > IVY_INIT_DC {
			panic("ivy: invalid latch initialization")
		}
	}
	if p.fFanout == 0 {
		Ivy_ManStartFanout(p)
	}
	piStart, poStart := len(p.vPis)-int(registers), len(p.vPos)-int(registers)
	for i := 0; i < int(registers); i++ {
		init := IVY_INIT_0
		if inits != nil {
			init = Ivy_Init_t(inits[i])
		}
		output := p.vPos[poStart+i]
		latch := Ivy_Latch(p, Ivy_ObjChild0(output), init)
		Ivy_ObjDisconnect(p, output)
		p.vObjs[output.Id] = nil
		Ivy_ManRecycleMemory(p, output)
		input := p.vPis[piStart+i]
		input.Type = IVY_BUF
		Ivy_ObjConnect(p, input, latch, nil)
		p.vBufs = append(p.vBufs, input)
	}
	clear(p.vPis[piStart:])
	clear(p.vPos[poStart:])
	p.vPis = p.vPis[:piStart]
	p.vPos = p.vPos[:poStart]
	p.nObjs[IVY_PI] -= registers
	p.nObjs[IVY_PO] -= registers
	p.nObjs[IVY_BUF] += registers
	p.nDeleted += registers
	Ivy_ManCleanup(p)
	Ivy_ManCleanupSeq(p)
	Ivy_ManPropagateBuffers(p, 0)
	Ivy_ManResetLevels(p)
	if Ivy_ManCheck(p) == 0 {
		panic("ivy: sequential conversion failed validation")
	}
}
