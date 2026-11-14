package gia

import "fmt"

// Gia_ManFrontFindNext reserves the next available frontier slot.
func Gia_ManFrontFindNext(front []byte, nFront, iFront int32) int32 {
	if nFront <= 0 || int(nFront) > len(front) || iFront < 0 || iFront >= nFront {
		panic("gia: invalid frontier")
	}
	start := iFront
	for front[iFront] != 0 {
		iFront = (iFront + 1) % nFront
		if iFront == start {
			panic("gia: frontier is full")
		}
	}
	front[iFront] = 1
	return iFront
}

// Gia_ManFrontTransform restores backward object differences from frontier slots.
func Gia_ManFrontTransform(p *Gia_Man_t) {
	if p.nFront <= 0 {
		panic("gia: graph has no frontier")
	}
	frontToID := make([]int32, p.nFront)
	for i := range frontToID {
		frontToID[i] = -1
	}
	for i, obj := range p.pObjs {
		id := int32(i)
		if Gia_ObjIsCo(obj) != 0 {
			if obj.Value != GIA_NONE {
				panic("gia: invalid frontier output")
			}
			obj.iDiff0 = uint32(id - frontToID[obj.iDiff0])
		} else if Gia_ObjIsAnd(obj) != 0 {
			obj.iDiff0 = uint32(id - frontToID[obj.iDiff0])
			obj.iDiff1 = uint32(id - frontToID[obj.iDiff1])
			frontToID[obj.Value] = id
		} else {
			frontToID[obj.Value] = id
		}
		obj.Value = 0
	}
}

// Gia_ManCrossCutSimple counts the maximum number of simultaneously live nodes.
// Like the C implementation, it uses and clears object Value fields.
func Gia_ManCrossCutSimple(p *Gia_Man_t) int32 {
	Gia_ManCreateValueRefs(p)
	var current, maximum int32
	for _, obj := range p.pObjs {
		if obj.Value != 0 {
			current++
		}
		if current > maximum {
			maximum = current
		}
		if Gia_ObjIsAnd(obj) != 0 {
			for _, fanin := range []*Gia_Obj_t{Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)} {
				fanin.Value--
				if fanin.Value == 0 {
					current--
				}
			}
		} else if Gia_ObjIsCo(obj) != 0 {
			fanin := Gia_ObjFanin0(obj)
			fanin.Value--
			if fanin.Value == 0 {
				current--
			}
		}
	}
	return maximum
}

// Gia_ManFront duplicates p with fanin differences represented as frontier slots.
// Call Gia_ManFrontTransform before using normal graph traversal on the result.
func Gia_ManFront(p *Gia_Man_t) *Gia_Man_t {
	maxCut := Gia_ManCrossCutSimple(p)
	Gia_ManCreateValueRefs(p)
	result := Gia_ManStart(p.nObjs)
	result.pName, result.PSpec = p.pName, p.PSpec
	result.nFront = 1 + int32(float32(maxCut)*1.1)
	front := make([]byte, result.nFront)
	position := Gia_ManFrontFindNext(front, result.nFront, 0)
	result.pObjs[0].Value = uint32(position)
	if p.pObjs[0].Value == 0 {
		front[position] = 0
	}
	for id := int32(1); id < p.nObjs; id++ {
		obj := p.pObjs[id]
		if Gia_ObjIsCi(obj) != 0 {
			lit := Gia_ManAppendCi(result)
			out := result.pObjs[lit>>1]
			position = Gia_ManFrontFindNext(front, result.nFront, position)
			out.Value = uint32(position)
			if obj.Value == 0 {
				front[position] = 0
			}
			continue
		}
		fanin0 := p.pObjs[Gia_ObjFaninId0(obj, id)]
		slot0 := result.pObjs[fanin0.id].Value
		if Gia_ObjIsCo(obj) != 0 {
			lit := Gia_ManAppendCo(result, 0)
			out := result.pObjs[lit>>1]
			out.Value = GIA_NONE
			out.iDiff0, out.fCompl0 = slot0, obj.fCompl0
			fanin0.Value--
			if fanin0.Value == 0 {
				front[slot0] = 0
			}
			continue
		}
		fanin1 := p.pObjs[Gia_ObjFaninId1(obj, id)]
		slot1 := result.pObjs[fanin1.id].Value
		out := Gia_ManAppendObj(result)
		out.iDiff0, out.fCompl0 = slot0, obj.fCompl0
		out.iDiff1, out.fCompl1 = slot1, obj.fCompl1
		position = Gia_ManFrontFindNext(front, result.nFront, position)
		out.Value = uint32(position)
		fanin0.Value--
		if fanin0.Value == 0 {
			front[slot0] = 0
		}
		fanin1.Value--
		if fanin1.Value == 0 {
			front[slot1] = 0
		}
		if obj.Value == 0 {
			front[position] = 0
		}
	}
	Gia_ManSetRegNum(result, Gia_ManRegNum(p))
	return result
}

// Gia_ManFrontTest checks that frontier conversion preserves graph fields.
func Gia_ManFrontTest(p *Gia_Man_t) {
	result := Gia_ManFront(p)
	Gia_ManFrontTransform(result)
	for i, a := range p.pObjs {
		b := result.pObjs[i]
		if a.iDiff0 != b.iDiff0 || a.iDiff1 != b.iDiff1 || a.fCompl0 != b.fCompl0 || a.fCompl1 != b.fCompl1 || a.fTerm != b.fTerm || a.Value != b.Value {
			fmt.Println("Verification failed.")
			return
		}
	}
	fmt.Println("Verification successful.")
}
