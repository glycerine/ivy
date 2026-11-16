package if_

import (
	"fmt"
	"math"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Iff_Man_t holds delay and clustering choices for a mapped GIA.
type Iff_Man_t struct {
	graph           *Gia_Man_t
	library         *If_LibLut_t
	lutSize, degree int32
	times           []float32
	matches         [4][]int32
}

func Gia_ManIffStart(graph *Gia_Man_t) *Iff_Man_t {
	p := &Iff_Man_t{graph: graph, times: make([]float32, Gia_ManObjNum(graph))}
	for i := range p.times {
		p.times[i] = math.Float32frombits(0xffffffff) // Vec_FltStartFull byte fill.
	}
	for _, k := range []int{2, 3} {
		p.matches[k] = make([]int32, len(p.times))
		for i := range p.matches[k] {
			p.matches[k][i] = -1
		}
	}
	return p
}
func Gia_ManIffStop(p *Iff_Man_t) {
	if p != nil {
		*p = Iff_Man_t{}
	}
}
func Iff_ObjTimeId(p *Iff_Man_t, id int32) float32           { return p.times[id] }
func Iff_ObjTime(p *Iff_Man_t, obj *Gia_Obj_t) float32       { return p.times[gia.Gia_ObjId(p.graph, obj)] }
func Iff_ObjSetTimeId(p *Iff_Man_t, id int32, value float32) { p.times[id] = value }
func Iff_ObjSetTime(p *Iff_Man_t, obj *Gia_Obj_t, value float32) {
	p.times[gia.Gia_ObjId(p.graph, obj)] = value
}
func Iff_ObjMatchId(p *Iff_Man_t, id, kind int32) int32 { return p.matches[kind][id] }
func Iff_ObjMatch(p *Iff_Man_t, obj *Gia_Obj_t, kind int32) int32 {
	return Iff_ObjMatchId(p, gia.Gia_ObjId(p.graph, obj), kind)
}
func Iff_ObjSetMatchId(p *Iff_Man_t, id, kind, match int32) { p.matches[kind][id] = match }
func Iff_ObjSetMatch(p *Iff_Man_t, obj *Gia_Obj_t, kind, match int32) {
	Iff_ObjSetMatchId(p, gia.Gia_ObjId(p.graph, obj), kind, match)
}

// iffExpandedFanins visits the boundary of a cluster, omitting packed roots.
func iffExpandedFanins(graph *Gia_Man_t, root, skip2, skip3 int32, visit func(int32)) {
	for _, id := range gia.Gia_ObjLutFanins(graph, root) {
		if id != skip2 && id != skip3 {
			visit(id)
		}
	}
	if skip2 >= 0 {
		for _, id := range gia.Gia_ObjLutFanins(graph, skip2) {
			if id != skip3 {
				visit(id)
			}
		}
	}
	if skip3 >= 0 {
		for _, id := range gia.Gia_ObjLutFanins(graph, skip3) {
			if id != skip2 {
				visit(id)
			}
		}
	}
}
func Gia_IffObjCount(graph *Gia_Man_t, root, skip2, skip3 int32) int32 {
	gia.Gia_ManIncrementTravId(graph)
	var count int32
	iffExpandedFanins(graph, root, skip2, skip3, func(id int32) {
		if gia.Gia_ObjIsTravIdCurrentId(graph, id) == 0 {
			gia.Gia_ObjSetTravIdCurrentId(graph, id)
			count++
		}
	})
	return count
}
func Gia_IffObjTimeOne(p *Iff_Man_t, root, skip2, skip3 int32) float32 {
	delay := -float32(util.ABC_INFINITY)
	visit := func(id int32) {
		if delay < p.times[id] {
			delay = p.times[id]
		}
	}
	for _, id := range gia.Gia_ObjLutFanins(p.graph, root) {
		if id != skip2 && id != skip3 {
			visit(id)
		}
	}
	if skip2 == -1 {
		return delay
	}
	for _, id := range gia.Gia_ObjLutFanins(p.graph, skip2) {
		if id != skip3 {
			visit(id)
		}
	}
	if skip3 == -1 {
		return delay
	}
	for _, id := range gia.Gia_ObjLutFanins(p.graph, skip3) {
		if id != skip2 {
			visit(id)
		}
	}
	assert(delay >= 0, "nonnegative triple boundary arrival")
	return delay
}
func Gia_IffObjTimeTwo(p *Iff_Man_t, root int32, fanin *int32, minimum float32) float32 {
	*fanin = -1
	for _, id := range gia.Gia_ObjLutFanins(p.graph, root) {
		if gia.Gia_ObjIsCi(Gia_ManObj(p.graph, id)) != 0 {
			continue
		}
		value := Gia_IffObjTimeOne(p, root, id, -1)
		size := Gia_IffObjCount(p.graph, root, id, -1)
		assert(size <= p.library.LutMax, "cluster fits LUT library")
		value += p.library.pLutDelays[size][0]
		if value < minimum {
			minimum, *fanin = value, id
		}
	}
	return minimum
}
func Gia_IffObjTimeThree(p *Iff_Man_t, root int32, first, second *int32, minimum float32) float32 {
	*first, *second = -1, -1
	for _, a := range gia.Gia_ObjLutFanins(p.graph, root) {
		if gia.Gia_ObjIsCi(Gia_ManObj(p.graph, a)) != 0 {
			continue
		}
		for _, b := range gia.Gia_ObjLutFanins(p.graph, root) {
			if a == b || gia.Gia_ObjIsCi(Gia_ManObj(p.graph, b)) != 0 {
				continue
			}
			value := Gia_IffObjTimeOne(p, root, a, b)
			size := Gia_IffObjCount(p.graph, root, a, b)
			assert(size <= p.library.LutMax, "cluster fits LUT library")
			value += p.library.pLutDelays[size][0]
			if value < minimum {
				minimum, *first, *second = value, a, b
			}
		}
	}
	return minimum
}

func Gia_ManIffPerform(graph *Gia_Man_t, library *If_LibLut_t, timing *Tim_Man_t, lutSize, degree int32) *Iff_Man_t {
	assert(degree == 2 || degree == 3, "cluster degree is two or three")
	p := Gia_ManIffStart(graph)
	p.library, p.lutSize, p.degree = library, lutSize, degree
	p.times[0] = 0
	tim.Tim_ManIncrementTravId(timing)
	var all, pairs, triples int32
	maximum := -float32(util.ABC_INFINITY)
	for id := int32(1); id < Gia_ManObjNum(graph); id++ {
		obj := Gia_ManObj(graph, id)
		switch {
		case gia.Gia_ObjIsAnd(obj) != 0:
			if gia.Gia_ObjIsLut(graph, id) == 0 {
				continue
			}
			all++
			size := gia.Gia_ObjLutSize(graph, id)
			assert(size <= library.LutMax, "LUT fits library")
			one := Gia_IffObjTimeOne(p, id, -1, -1) + library.pLutDelays[size][0]
			var fanin, first, second int32
			two := Gia_IffObjTimeTwo(p, id, &fanin, one)
			if degree == 2 {
				p.times[id] = two
				if two < one {
					p.matches[2][id] = fanin
					pairs++
				}
				continue
			}
			three := Gia_IffObjTimeThree(p, id, &first, &second, two)
			p.times[id] = three
			if three == one {
				continue
			}
			if three == two {
				p.matches[2][id] = fanin
				pairs++
			} else {
				assert(three < two, "triple strictly improves pair delay")
				p.matches[2][id], p.matches[3][id] = first, second
				triples++
			}
		case gia.Gia_ObjIsCi(obj) != 0:
			p.times[id] = tim.Tim_ManGetCiArrival(timing, int(gia.Gia_ObjCioId(obj)))
		case gia.Gia_ObjIsCo(obj) != 0:
			p.times[id] = p.times[gia.Gia_ObjFaninId0p(graph, obj)]
			tim.Tim_ManSetCoArrival(timing, int(gia.Gia_ObjCioId(obj)), p.times[id])
			maximum = util.Abc_MaxFloat(maximum, p.times[id])
		default:
			panic("if: invalid IFF object")
		}
	}
	fmt.Printf("Max delay = %.2f.  Count1 = %d.  Count2 = %d.  Count3 = %d.\n", maximum, all-pairs-triples, pairs, triples)
	return p
}
func Gia_ManIffSelect(p *Iff_Man_t) *Vec_Int_t {
	packing := vec.Vec_IntAlloc(int(Gia_ManObjNum(p.graph)))
	Vec_IntPush(packing, 0)
	gia.Gia_ManIncrementTravId(p.graph)
	gia.Gia_ObjSetTravIdCurrentId(p.graph, 0)
	for i := int32(0); i < Gia_ManCiNum(p.graph); i++ {
		gia.Gia_ObjSetTravIdCurrent(p.graph, Gia_ManCi(p.graph, i))
	}
	var visit func(int32)
	visit = func(id int32) {
		if gia.Gia_ObjIsTravIdCurrentId(p.graph, id) != 0 {
			return
		}
		gia.Gia_ObjSetTravIdCurrentId(p.graph, id)
		assert(gia.Gia_ObjIsLut(p.graph, id) != 0, "selected node is a LUT")
		a, b := p.matches[2][id], p.matches[3][id]
		if a == -1 {
			assert(b == -1, "valid single packing")
		} else if b == -1 {
			assert(a > 0, "positive packed root")
		} else {
			assert(a > 0 && b > 0, "positive packed roots")
		}
		iffExpandedFanins(p.graph, id, a, b, visit)
		switch {
		case a == -1:
			Vec_IntPush(packing, 1)
		case b == -1:
			Vec_IntPush(packing, 2)
			Vec_IntPush(packing, a)
		default:
			Vec_IntPush(packing, 3)
			Vec_IntPush(packing, a)
			Vec_IntPush(packing, b)
		}
		Vec_IntPush(packing, id)
		vec.Vec_IntAddToEntry(packing, 0, 1)
	}
	for i := int32(0); i < Gia_ManCoNum(p.graph); i++ {
		visit(gia.Gia_ObjFaninId0p(p.graph, Gia_ManCo(p.graph, i)))
	}
	return packing
}
func Gia_ManIffTest(graph *Gia_Man_t, library *If_LibLut_t, verbose int32) {
	size, degree := gia.Gia_ManLutSizeMax(graph), int32(-1)
	switch {
	case size <= 4:
		size = 4
		if library.LutMax == 7 {
			degree = 2
		} else if library.LutMax == 10 {
			degree = 3
		}
	case size <= 6:
		size = 6
		if library.LutMax == 11 {
			degree = 2
		} else if library.LutMax == 16 {
			degree = 3
		}
	default:
		fmt.Println("The LUT size is more than 6.")
		return
	}
	if degree < 0 {
		if size == 4 {
			fmt.Println("LUT library for packing 4-LUTs should have 7 or 10 inputs.")
		} else {
			fmt.Println("LUT library for packing 6-LUTs should have 11 or 16 inputs.")
		}
		return
	}
	if verbose != 0 {
		fmt.Printf("Performing %d-clustering with %d-LUTs:\n", degree, size)
	}
	_, _, _, timing := graph.MappingInputs()
	if timing == nil {
		timing = tim.Tim_ManStart(int(Gia_ManCiNum(graph)), int(Gia_ManCoNum(graph)))
		defer tim.Tim_ManStop(timing)
	}
	p := Gia_ManIffPerform(graph, library, timing, size, degree)
	defer Gia_ManIffStop(p)
	if err := graph.ReplacePacking(Gia_ManIffSelect(p)); err != nil {
		panic(err)
	}
	if verbose != 0 {
		gia.Gia_ManPrintPackingStats(graph)
	}
}
