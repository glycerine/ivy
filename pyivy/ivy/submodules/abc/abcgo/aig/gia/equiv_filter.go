package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManEquivTransform(p *Gia_Man_t, verbose int32) {
	if p.PReprs == nil || p.PNexts == nil {
		panic("gia: missing equivalence classes")
	}
	for _, obj := range p.pObjs {
		if (Gia_ObjIsCi(obj) != 0 || Gia_ObjIsAnd(obj) != 0) && Gia_ObjColors(p, obj.id) == 0 {
			panic("gia: uncolored candidate")
		}
	}
	var totalClasses, removedClasses, totalLits, removedLits int32
	for head := p.nObjs - 1; head > 0; head-- {
		if Gia_ObjIsHead(p, head) == 0 {
			continue
		}
		totalClasses++
		var keep []int32
		for id := head; id != 0; {
			next := p.PNexts[id]
			totalLits++
			colors := Gia_ObjColors(p, id)
			if colors == 0 {
				panic("gia: uncolored class member")
			}
			if colors != 3 {
				keep = append(keep, id)
			} else {
				removedLits++
			}
			p.PReprs[id].fFailed = 0
			p.PReprs[id].fProved = 0
			Gia_ObjUnsetRepr(p, id)
			p.PNexts[id] = 0
			id = next
		}
		if len(keep) < 2 {
			removedClasses++
			continue
		}
		for i := 1; i < len(keep); i++ {
			Gia_ObjSetRepr(p, keep[i], keep[0])
			p.PNexts[keep[i-1]] = keep[i]
		}
	}
	if verbose != 0 {
		fmt.Printf("Removed classes = %6d (out of %6d). Removed literals = %6d (out of %6d).\n", removedClasses, totalClasses, removedLits, totalLits)
	}
}
func Gia_ManEquivMark(p *Gia_Man_t, filename string, skipSome, verbose int32) {
	total := Gia_ManEquivCountLitsAll(p)
	if total == 0 {
		fmt.Println("Gia_ManEquivMark(): Current AIG does not have equivalences.")
		return
	}
	miter, err := Gia_AigerRead(filename, 0, 0, 0)
	if err != nil {
		fmt.Printf("Gia_ManEquivMark(): Input file %s could not be read: %v\n", filename, err)
		return
	}
	defer Gia_ManStop(miter)
	var trace *Vec_Int_t
	added := total
	if skipSome != 0 {
		trace = vec.Vec_IntAlloc(100)
		q := Gia_ManSpecReduceTrace(p, trace, nil)
		Gia_ManStop(q)
		if int32(vec.Vec_IntSize(trace)) != total {
			panic("gia: invalid equivalence trace")
		}
		added = 0
		for _, x := range vec.Vec_IntArray(trace) {
			if x != 0 {
				added++
			}
		}
	}
	if Gia_ManPoNum(miter) != Gia_ManPoNum(p)+added {
		fmt.Printf("Gia_ManEquivMark(): Incorrect output count: %d, expected %d.\n", Gia_ManPoNum(miter), Gia_ManPoNum(p)+added)
		return
	}
	var index, output, proved int32
	for id := int32(0); id < p.nObjs; id++ {
		if Gia_ObjRepr(p, id) == GIA_VOID {
			continue
		}
		include := true
		if trace != nil {
			include = vec.Vec_IntEntry(trace, int(index)) != 0
		}
		index++
		if !include {
			continue
		}
		obj := Gia_ManPo(miter, Gia_ManPoNum(p)+output)
		output++
		if Gia_ObjFaninLit0p(miter, obj) == 0 {
			Gia_ObjSetProved(p, id)
			proved++
		}
	}
	if verbose != 0 {
		fmt.Printf("Set %d equivalences as proved.\n", proved)
	}
}
func Gia_ManEquivFilter(p *Gia_Man_t, outputs *Vec_Int_t, verbose int32) {
	if p.PReprs == nil || p.PNexts == nil {
		fmt.Println("Gia_ManEquivFilter(): Equivalence classes are not defined.")
		return
	}
	if outputs == nil {
		fmt.Println("Gia_ManEquivFilter(): Array of disproved POs is not available.")
		return
	}
	if vec.Vec_IntSize(outputs) == 0 {
		return
	}
	mapping, trace := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	srm := Gia_ManSpecReduceTrace(p, trace, mapping)
	if Gia_ManPoNum(srm) != Gia_ManPoNum(p)+int32(vec.Vec_IntSize(mapping)) {
		Gia_ManStop(srm)
		panic("gia: speculative output count mismatch")
	}
	Gia_ManStop(srm)
	if verbose != 0 {
		fmt.Printf("Design POs = %d. SRM POs = %d. Spec POs = %d. Disproved POs = %d.\n", Gia_ManPoNum(p), Gia_ManPoNum(p)+int32(vec.Vec_IntSize(mapping)), vec.Vec_IntSize(mapping), vec.Vec_IntSize(outputs))
	}
	vec.Vec_IntSort(outputs, 0)
	previous := int32(-1)
	for _, id := range vec.Vec_IntArray(outputs) {
		if id < 0 || id >= Gia_ManPoNum(p)+int32(vec.Vec_IntSize(mapping)) {
			fmt.Printf("Gia_ManEquivFilter(): Output index %d is out of range.\n", id)
			return
		}
		if id == previous {
			fmt.Printf("Gia_ManEquivFilter(): Duplicate output index %d.\n", id)
			return
		}
		previous = id
		if id < Gia_ManPoNum(p) {
			fmt.Printf("Gia_ManEquivFilter(): One of the original POs (%d) have failed.\n", id)
		}
	}
	for _, id := range vec.Vec_IntArray(outputs) {
		if id >= Gia_ManPoNum(p) {
			Gia_ObjUnsetRepr(p, vec.Vec_IntEntry(mapping, int(id-Gia_ManPoNum(p))))
		}
	}
	p.PNexts = Gia_ManDeriveNexts(p)
}
func Gia_ManEquivFilterTest(p *Gia_Man_t) {
	outputs := vec.Vec_IntAlloc(10)
	for i := int32(0); i < 10; i++ {
		id := Gia_ManPoNum(p) + 2*i + 2
		vec.Vec_IntPush(outputs, id)
		fmt.Printf("%d ", id)
	}
	fmt.Println()
	Gia_ManEquivFilter(p, outputs, 1)
}

func Gia_ManFilterEquivsUsingLatches(p *Gia_Man_t, flopsOnly, flopsWith, useDrivers int32) {
	if (flopsOnly != 0) == (flopsWith != 0) {
		panic("gia: choose one latch filtering mode")
	}
	selected := make([]bool, p.nObjs)
	for i := int32(0); i < p.nRegs; i++ {
		if useDrivers != 0 {
			obj := Gia_ManRi(p, i)
			selected[Gia_ObjFaninId0(obj, obj.id)] = true
		} else {
			selected[Gia_ManRo(p, i).id] = true
		}
	}
	var before, after int32
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsConst(p, id) != 0 {
			before++
			if !selected[id] {
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
		var members, keep []int32
		found := false
		for id := head; id != 0; id = p.PNexts[id] {
			members = append(members, id)
			if selected[id] {
				found = true
				keep = append(keep, id)
			}
		}
		before += int32(len(members) - 1)
		if flopsWith != 0 && found {
			after += int32(len(members) - 1)
			continue
		}
		for _, id := range members {
			Gia_ObjUnsetRepr(p, id)
			p.PNexts[id] = 0
		}
		if flopsOnly != 0 && len(keep) > 1 {
			for i := 1; i < len(keep); i++ {
				Gia_ObjSetRepr(p, keep[i], keep[0])
				p.PNexts[keep[i-1]] = keep[i]
				after++
			}
		}
	}
	fmt.Printf("The number of literals: Before = %d. After = %d.\n", before, after)
}
func Gia_ManEquivImprove(p *Gia_Man_t) {
	if p.PReprs == nil || p.PNexts == nil {
		panic("gia: missing equivalence classes")
	}
	Gia_ManLevelNum(p)
	Gia_ManCreateRefs(p)
	for head := int32(1); head < p.nObjs; head++ {
		if Gia_ObjIsHead(p, head) == 0 {
			continue
		}
		best, bestLevel, bestMffc := int32(-1), int32(embInfinity), int32(embInfinity)
		var members []int32
		for id := head; id != 0; id = p.PNexts[id] {
			obj := Gia_ManObj(p, id)
			level, mffc := Gia_ObjLevel(p, obj), Gia_NodeMffcSize(p, obj)
			if level < bestLevel || level == bestLevel && mffc < bestMffc {
				best, bestLevel, bestMffc = id, level, mffc
			}
			members = append(members, id)
		}
		if best == head {
			continue
		}
		Gia_ObjUnsetRepr(p, best)
		Gia_ObjSetProved(p, head)
		Gia_ObjUnsetProved(p, best)
		for _, id := range members {
			if id != best {
				p.PReprs[id].iRepr = uint32(best)
			}
		}
	}
	p.PNexts = nil
}
