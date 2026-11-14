package gia

import "fmt"

func Cec4_ManMarkIndependentClasses_rec(p *Gia_Man_t, id int32) int32 {
	if id <= 0 {
		panic("gia: independent-class traversal expects nonconstant object")
	}
	if Gia_ObjIsTravIdPreviousId(p, id) != 0 {
		return 0
	}
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return 1
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		return 1
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: independent-class traversal expects AND")
	}
	if Cec4_ManMarkIndependentClasses_rec(p, Gia_ObjFaninId0(obj, id)) != 0 && Cec4_ManMarkIndependentClasses_rec(p, Gia_ObjFaninId1(obj, id)) != 0 {
		return 1
	}
	Gia_ObjSetTravIdPreviousId(p, id)
	return 0
}
func Cec4_ManMarkIndependentClasses(p, q *Gia_Man_t) int32 {
	Gia_ManCleanMark01(p)
	var found int32
	for head := int32(1); head < p.nObjs; head++ {
		if Gia_ObjIsHead(p, head) == 0 {
			continue
		}
		Gia_ManIncrementTravId(q)
		Gia_ManIncrementTravId(q)
		id := Gia_ObjValue(Gia_ManObj(p, head)) >> 1
		if Cec4_ManMarkIndependentClasses_rec(q, id) == 0 {
			panic("gia: dependent class head")
		}
		Gia_ObjSetTravIdPreviousId(q, id)
		p.PReprs[head].fColorA = 1
		for member := p.PNexts[head]; member != 0; member = p.PNexts[member] {
			if Gia_ObjRepr(p, member) != head {
				panic("gia: invalid equivalence class")
			}
			id = Gia_ObjValue(Gia_ManObj(p, member)) >> 1
			if Cec4_ManMarkIndependentClasses_rec(q, id) != 0 {
				p.PReprs[member].fColorA = 1
				found = 1
			}
			Gia_ObjSetTravIdPreviousId(q, id)
		}
	}
	return found
}
func Gia_ManTransferTest(p *Gia_Man_t) {
	oldReprs, oldNexts := p.PReprs, p.PNexts
	q := Gia_ManChangeOrder(p)
	Gia_ManTransferEquivs(p, q)
	p.PReprs = nil
	p.PNexts = nil
	for _, obj := range p.pObjs {
		lit := Gia_ObjValue(obj)
		if lit&1 != 0 {
			panic("gia: complemented reordered copy")
		}
		Gia_ObjSetValue(Gia_ManObj(q, lit>>1), obj.id<<1)
	}
	Gia_ManTransferEquivs(q, p)
	reprsMatch, nextsMatch := len(p.PReprs) == len(oldReprs), len(p.PNexts) == len(oldNexts)
	if reprsMatch {
		for i := range oldReprs {
			oldReprs[i].fProved = 0
			if p.PReprs[i] != oldReprs[i] {
				reprsMatch = false
			}
		}
	}
	if nextsMatch {
		for i := range oldNexts {
			if p.PNexts[i] != oldNexts[i] {
				nextsMatch = false
			}
		}
	}
	if reprsMatch {
		fmt.Println("Verification of reprs succeeded.")
	} else {
		fmt.Println("Verification of reprs failed.")
	}
	if nextsMatch {
		fmt.Println("Verification of nexts succeeded.")
	} else {
		fmt.Println("Verification of nexts failed.")
	}
	Gia_ManStop(q)
}
