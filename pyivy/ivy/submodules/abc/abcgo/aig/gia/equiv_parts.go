package gia

import "fmt"

// giaMarkMiterParts reconstructs the same sequential miter and marks its
// vertices according to the source graphs' remapped object copies.
func giaMarkMiterParts(p *Gia_Man_t, nameA, nameB string) bool {
	if p.PReprs == nil || p.PNexts == nil {
		fmt.Println("Equivalences are not defined.")
		return false
	}
	a, err := Gia_AigerRead(nameA, 0, 0, 0)
	if err != nil {
		fmt.Printf("Cannot read first file %s: %v\n", nameA, err)
		return false
	}
	defer Gia_ManStop(a)
	b, err := Gia_AigerRead(nameB, 0, 0, 0)
	if err != nil {
		fmt.Printf("Cannot read second file %s: %v\n", nameB, err)
		return false
	}
	defer Gia_ManStop(b)
	m := Gia_ManMiter(a, b, 0, 0, 1, 0, 0)
	if m == nil {
		fmt.Println("Cannot create sequential miter.")
		return false
	}
	defer Gia_ManStop(m)
	if p.nObjs != m.nObjs {
		fmt.Println("The number of objects is different.")
		return false
	}
	for i, x := range p.pObjs {
		y := m.pObjs[i]
		if x.iDiff0 != y.iDiff0 || x.iDiff1 != y.iDiff1 || x.fCompl0 != y.fCompl0 || x.fCompl1 != y.fCompl1 || x.fTerm != y.fTerm {
			fmt.Println("The AIG structure of the miter does not match.")
			return false
		}
	}
	Gia_ManCleanMark0(p)
	Gia_ManCleanMark1(p)
	for _, obj := range a.pObjs {
		if lit := Gia_ObjValue(obj); lit >= 0 {
			p.pObjs[lit>>1].fMark0 = 1
		}
	}
	for _, obj := range b.pObjs {
		if lit := Gia_ObjValue(obj); lit >= 0 {
			p.pObjs[lit>>1].fMark1 = 1
		}
	}
	return true
}
func Gia_ManFilterEquivsForSpeculation(p *Gia_Man_t, nameA, nameB string, latchA, latchB int32) int32 {
	if !giaMarkMiterParts(p, nameA, nameB) {
		return 0
	}
	defer Gia_ManCleanMark0(p)
	defer Gia_ManCleanMark1(p)
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsConst(p, id) != 0 {
			Gia_ObjUnsetRepr(p, id)
		}
	}
	var count int32
	for head := int32(1); head < p.nObjs; head++ {
		if Gia_ObjIsHead(p, head) == 0 {
			continue
		}
		a, b := int32(-1), int32(-1)
		for id := head; id != 0; {
			next := p.PNexts[id]
			obj := p.pObjs[id]
			if a < 0 && obj.fMark0 != 0 && obj.fMark1 == 0 && (latchA == 0 || Gia_ObjIsRo(p, obj) != 0) {
				a = id
			}
			if b < 0 && obj.fMark1 != 0 && obj.fMark0 == 0 && (latchB == 0 || Gia_ObjIsRo(p, obj) != 0) {
				b = id
			}
			Gia_ObjUnsetRepr(p, id)
			p.PNexts[id] = 0
			id = next
		}
		if a > 0 && b > 0 {
			a, b = min(a, b), max(a, b)
			p.PNexts[a] = b
			Gia_ObjSetRepr(p, b, a)
			count++
		}
	}
	fmt.Printf("The number of two-node classes after filtering = %d.\n", count)
	return 1
}
func Gia_ManFilterEquivsUsingParts(p *Gia_Man_t, nameA, nameB string) int32 {
	if !giaMarkMiterParts(p, nameA, nameB) {
		return 0
	}
	defer Gia_ManCleanMark0(p)
	defer Gia_ManCleanMark1(p)
	var before, after int32
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsConst(p, id) != 0 {
			before++
			obj := p.pObjs[id]
			if obj.fMark0 != 0 && obj.fMark1 != 0 {
				Gia_ObjUnsetRepr(p, id)
			} else {
				after++
			}
		}
	}
	for head := int32(1); head < p.nObjs; head++ {
		if Gia_ObjIsHead(p, head) == 0 {
			continue
		}
		var keep []int32
		seenA, seenB := false, false
		count := int32(0)
		for id := head; id != 0; {
			next := p.PNexts[id]
			obj := p.pObjs[id]
			if obj.fMark0 != 0 && obj.fMark1 == 0 {
				seenA = true
				keep = append(keep, id)
			}
			if obj.fMark1 != 0 && obj.fMark0 == 0 {
				seenB = true
				keep = append(keep, id)
			}
			Gia_ObjUnsetRepr(p, id)
			p.PNexts[id] = 0
			count++
			id = next
		}
		before += count - 1
		if seenA && seenB && len(keep) > 1 {
			for i := 1; i < len(keep); i++ {
				Gia_ObjSetRepr(p, keep[i], keep[0])
				p.PNexts[keep[i-1]] = keep[i]
				after++
			}
		}
	}
	fmt.Printf("The number of literals: Before = %d. After = %d.\n", before, after)
	return 1
}
