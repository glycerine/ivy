package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func Gia_ManSpeedup(p *Gia_Man_t, percentage, degree, verbose, veryVerbose int32) *Gia_Man_t {
	if Gia_ManHasMapping(p) == 0 {
		panic("gia: speedup requires a mapped graph")
	}
	hasLib := p.pLutLib != nil
	if !hasLib && p.pManTime != nil {
		old, ok := p.pManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
		duplicate := tim.Tim_ManDup(old, 1)
		p.pManTime = duplicate
		defer func() { tim.Tim_ManStop(duplicate); p.pManTime = old }()
	}
	arrival := Gia_ManDelayTraceLut(p)
	defer Gia_ManTimeStop(p)
	delta := float32(1)
	if hasLib {
		delta = arrival * float32(percentage) / 100
	}
	if verbose != 0 {
		model := "unit-delay"
		if hasLib {
			model = "LUT library"
		}
		fmt.Printf("Max delay = %.2f. Delta = %.2f. Using %s model.\n", arrival, delta, model)
	}
	edges := make([]uint32, p.nObjs)
	for id := int32(0); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 && Gia_ObjTimeSlack(p, id) < delta {
			edges[id] = Gia_LutDelayTraceTCEdges(p, id, delta)
		}
	}
	if verbose != 0 {
		var total, critical int
		for id := int32(0); id < p.nObjs; id++ {
			if Gia_ObjIsLut(p, id) == 0 {
				continue
			}
			for _, fan := range Gia_ObjLutFanins(p, id) {
				if Gia_ObjIsCi(Gia_ManObj(p, fan)) == 0 && Gia_ObjTimeSlack(p, fan) < delta {
					total++
				}
			}
			critical += bits.OnesCount32(edges[id])
		}
		ratio := float64(0)
		if total != 0 {
			ratio = float64(critical) / float64(total)
		}
		fmt.Printf("Edges: Total = %7d. 0-slack = %7d. Critical = %7d. Ratio = %4.2f\n", Gia_ManLutFaninCount(p), total, critical, ratio)
	}
	dst := Gia_ManDup(p)
	Gia_ManHashStart(dst)
	capacity := 1000 + 3*dst.nObjs
	dst.pNexts = make([]int32, capacity)
	dst.pReprs = make([]Gia_Rpr_t, capacity)
	for i := range dst.pReprs {
		dst.pReprs[i].iRepr = GIA_VOID
	}
	critical, leaves := vec.Vec_IntAlloc(16), vec.Vec_IntAlloc(16)
	var examined, workable int32
	for id := int32(0); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 || Gia_ObjTimeSlack(p, id) >= delta {
			continue
		}
		fans := Gia_ObjLutFanins(p, id)
		var nCritical int32
		for k, fan := range fans {
			if Gia_ObjIsCi(Gia_ManObj(p, fan)) == 0 && edges[id]&(1<<uint(k)) != 0 {
				nCritical++
			}
		}
		if veryVerbose == 0 && nCritical == 0 {
			continue
		}
		examined++
		vec.Vec_IntClear(critical)
		for k, fan := range fans {
			if Gia_ObjIsCi(Gia_ManObj(p, fan)) != 0 || edges[id]&(1<<uint(k)) == 0 {
				continue
			}
			for j, child := range Gia_ObjLutFanins(p, fan) {
				if edges[fan]&(1<<uint(j)) != 0 {
					vec.Vec_IntPushUnique(critical, child)
				}
			}
		}
		if vec.Vec_IntSize(critical) == 0 || int32(vec.Vec_IntSize(critical)) > degree {
			continue
		}
		workable++
		vec.Vec_IntClear(leaves)
		for _, fan := range fans {
			if Gia_ObjIsCi(Gia_ManObj(p, fan)) != 0 {
				vec.Vec_IntPushUnique(leaves, fan)
			} else {
				for _, child := range Gia_ObjLutFanins(p, fan) {
					vec.Vec_IntPushUnique(leaves, child)
				}
			}
		}
		if veryVerbose != 0 {
			fmt.Printf("%5d Node %5d : %d %2d %2d\n", examined, id, nCritical, vec.Vec_IntSize(critical), vec.Vec_IntSize(leaves))
		}
		list := vec.Vec_IntArray(critical)
		if len(list) > 1 && Gia_ObjTimeSlack(p, list[0]) < Gia_ObjTimeSlack(p, list[1]) {
			list[0], list[1] = list[1], list[0]
		}
		if len(list) > 2 {
			if Gia_ObjTimeSlack(p, list[1]) < Gia_ObjTimeSlack(p, list[2]) {
				list[1], list[2] = list[2], list[1]
			}
			if Gia_ObjTimeSlack(p, list[0]) < Gia_ObjTimeSlack(p, list[1]) {
				list[0], list[1] = list[1], list[0]
			}
		}
		Gia_ManSpeedupObj(dst, p, Gia_ManObj(p, id), leaves, critical)
		if dst.nObjs > capacity-100 {
			fmt.Println("Speedup stopped adding choices because there were too many to add.")
			break
		}
	}
	if verbose != 0 {
		ratio := float64(0)
		if examined != 0 {
			ratio = float64(workable) / float64(examined)
		}
		fmt.Printf("Nodes: Total = %7d. 0-slack = %7d. Workable = %7d. Ratio = %4.2f\n", Gia_ManLutNum(p), examined, workable, ratio)
	}
	result := Gia_ManEquivToChoices(dst, 1)
	Gia_ManStop(dst)
	return result
}
