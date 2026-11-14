package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func giaEdgeFanins(p *Gia_Man_t, id int32) []int32 {
	var fanins []int32
	if p.VMapping != nil && Gia_ObjIsLut(p, id) != 0 {
		fanins = Gia_ObjLutFanins(p, id)
	} else if p.vMapping2 != nil && Gia_ObjIsLut2(p, id) != 0 {
		fanins = vec.Vec_IntArray(Gia_ObjLutFanins2(p, id))
	} else {
		panic("gia: edge delay requires a mapped LUT")
	}
	if len(fanins) > 4 {
		panic("gia: edge delay supports at most four LUT fanins")
	}
	return fanins
}
func giaIsEdgeLut(p *Gia_Man_t, id int32) bool {
	return p.VMapping != nil && Gia_ObjIsLut(p, id) != 0 || p.VMapping == nil && p.vMapping2 != nil && Gia_ObjIsLut2(p, id) != 0
}
func Gia_ObjEvalEdgeDelay(p *Gia_Man_t, id int32, delays *Vec_Int_t) int32 {
	var result int32
	for _, fanin := range giaEdgeFanins(p, id) {
		increment := int32(10)
		if Gia_ObjHaveEdge(p, id, fanin) != 0 {
			increment = 2
		}
		result = max(result, vec.Vec_IntEntry(delays, int(fanin))+increment)
	}
	return result
}
func Gia_ObjComputeEdgeDelay(p *Gia_Man_t, id int32, delays, first, second *Vec_Int_t, useTwo int32) int32 {
	vec.Vec_IntWriteEntry(first, int(id), 0)
	vec.Vec_IntWriteEntry(second, int(id), 0)
	var highest, runner, count int32
	max1, max2 := int32(-1), int32(-1)
	for _, fanin := range giaEdgeFanins(p, id) {
		delay := vec.Vec_IntEntry(delays, int(fanin)) + 10
		if delay > highest {
			runner, highest, max1, count = highest, delay, fanin, 1
		} else if delay == highest {
			max2 = fanin
			count++
			if useTwo == 0 {
				runner = highest
			}
		} else {
			runner = max(runner, delay)
		}
	}
	if count == 0 {
		panic("gia: mapped LUT has no fanins")
	}
	if highest > 10 && count == 1 && Gia_ObjEdgeCount(max1, first, second) <= 1 {
		Gia_ObjEdgeAdd(max1, id, first, second)
		Gia_ObjEdgeAdd(id, max1, first, second)
		highest = max(runner, highest-8)
	} else if highest > 10 && useTwo != 0 && count == 2 && Gia_ObjEdgeCount(max1, first, second) <= 1 && Gia_ObjEdgeCount(max2, first, second) <= 1 {
		Gia_ObjEdgeAdd(max1, id, first, second)
		Gia_ObjEdgeAdd(max2, id, first, second)
		Gia_ObjEdgeAdd(id, max1, first, second)
		Gia_ObjEdgeAdd(id, max2, first, second)
		highest = max(runner, highest-8)
	}
	vec.Vec_IntWriteEntry(delays, int(id), highest)
	return highest
}
func giaWalkEdgeTiming(p *Gia_Man_t, compute bool, useTwo int32) int32 {
	if p.VMapping == nil && p.vMapping2 == nil {
		panic("gia: missing LUT mapping")
	}
	var manager *tim.Tim_Man_t
	if p.PManTime != nil {
		var ok bool
		manager, ok = p.PManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
	}
	if manager != nil && tim.Tim_ManBoxNum(manager) > 0 {
		nodes := Gia_ManOrderWithBoxes(p)
		if nodes == nil {
			panic("gia: invalid timing-box order")
		}
		tim.Tim_ManIncrementTravId(manager)
		for _, id := range vec.Vec_IntArray(nodes) {
			obj := p.pObjs[id]
			switch {
			case Gia_ObjIsAnd(obj) != 0:
				if giaIsEdgeLut(p, id) {
					if compute {
						Gia_ObjComputeEdgeDelay(p, id, p.vEdgeDelay, p.VEdge1, p.VEdge2, useTwo)
					} else {
						vec.Vec_IntWriteEntry(p.vEdgeDelay, int(id), Gia_ObjEvalEdgeDelay(p, id, p.vEdgeDelay))
					}
				}
			case Gia_ObjIsCi(obj) != 0:
				vec.Vec_IntWriteEntry(p.vEdgeDelay, int(id), int32(tim.Tim_ManGetCiArrival(manager, int(Gia_ObjCioId(obj)))))
			case Gia_ObjIsCo(obj) != 0:
				tim.Tim_ManSetCoArrival(manager, int(Gia_ObjCioId(obj)), float32(vec.Vec_IntEntry(p.vEdgeDelay, int(Gia_ObjFaninId0(obj, id)))))
			case id != 0:
				panic("gia: invalid timing node")
			}
		}
	} else {
		for id := int32(0); id < p.nObjs; id++ {
			if !giaIsEdgeLut(p, id) {
				continue
			}
			if compute {
				Gia_ObjComputeEdgeDelay(p, id, p.vEdgeDelay, p.VEdge1, p.VEdge2, useTwo)
			} else {
				vec.Vec_IntWriteEntry(p.vEdgeDelay, int(id), Gia_ObjEvalEdgeDelay(p, id, p.vEdgeDelay))
			}
		}
	}
	var result int32
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		result = max(result, vec.Vec_IntEntry(p.vEdgeDelay, int(Gia_ObjFaninId0(obj, obj.id))))
	}
	return result
}
func Gia_ManEvalEdgeDelay(p *Gia_Man_t) int32 {
	if p.VEdge1 == nil || p.VEdge2 == nil {
		panic("gia: missing assigned edges")
	}
	p.vEdgeDelay = vec.Vec_IntStart(int(p.nObjs))
	return giaWalkEdgeTiming(p, false, 0)
}
func Gia_ManEvalEdgeCount(p *Gia_Man_t) int32 {
	return int32((vec.Vec_IntCountPositive(p.VEdge1) + vec.Vec_IntCountPositive(p.VEdge2)) / 2)
}
func Gia_ManComputeEdgeDelay(p *Gia_Man_t, useTwo int32) int32 {
	p.vEdgeDelay = vec.Vec_IntStart(int(p.nObjs))
	p.VEdge1 = vec.Vec_IntStart(int(p.nObjs))
	p.VEdge2 = vec.Vec_IntStart(int(p.nObjs))
	return giaWalkEdgeTiming(p, true, useTwo)
}

// Gia_ObjComputeEdgeDelay2 assigns edges while propagating delay backwards.
func Gia_ObjComputeEdgeDelay2(p *Gia_Man_t, id int32, delays, edges1, edges2, first, second, counts *Vec_Int_t) int32 {
	vec.Vec_IntWriteEntry(edges1, int(id), 0)
	vec.Vec_IntWriteEntry(edges2, int(id), 0)
	delay := vec.Vec_IntEntry(delays, int(id))
	count := vec.Vec_IntEntry(counts, int(id))
	if delay > 0 && (count == 1 || count == 2) {
		a := vec.Vec_IntEntry(first, int(id))
		b := vec.Vec_IntEntry(second, int(id))
		if Gia_ObjEdgeCount(a, edges1, edges2) <= 1 && (count == 1 || Gia_ObjEdgeCount(b, edges1, edges2) <= 1) {
			Gia_ObjEdgeAdd(a, id, edges1, edges2)
			Gia_ObjEdgeAdd(id, a, edges1, edges2)
			if count == 2 {
				Gia_ObjEdgeAdd(b, id, edges1, edges2)
				Gia_ObjEdgeAdd(id, b, edges1, edges2)
			}
			delay--
		}
	}
	vec.Vec_IntWriteEntry(delays, int(id), delay)
	for _, fanin := range giaEdgeFanins(p, id) {
		old := vec.Vec_IntEntry(delays, int(fanin))
		if old < delay+1 {
			vec.Vec_IntWriteEntry(delays, int(fanin), delay+1)
			vec.Vec_IntWriteEntry(first, int(fanin), id)
			vec.Vec_IntWriteEntry(counts, int(fanin), 1)
		} else if old == delay+1 {
			vec.Vec_IntWriteEntry(second, int(fanin), id)
			vec.Vec_IntWriteEntry(counts, int(fanin), vec.Vec_IntEntry(counts, int(fanin))+1)
		}
	}
	return delay
}

func Gia_ManComputeEdgeDelay2(p *Gia_Man_t) int32 {
	if p.PManTime != nil || p.VMapping == nil && p.vMapping2 == nil {
		panic("gia: reverse edge assignment requires mapping without timing boxes")
	}
	p.vEdgeDelay = vec.Vec_IntStart(int(p.nObjs))
	p.VEdge1 = vec.Vec_IntStart(int(p.nObjs))
	p.VEdge2 = vec.Vec_IntStart(int(p.nObjs))
	first := vec.Vec_IntStart(int(p.nObjs))
	second := vec.Vec_IntStart(int(p.nObjs))
	counts := vec.Vec_IntStart(int(p.nObjs))
	for id := p.nObjs - 1; id >= 0; id-- {
		if giaIsEdgeLut(p, id) {
			Gia_ObjComputeEdgeDelay2(p, id, p.vEdgeDelay, p.VEdge1, p.VEdge2, first, second, counts)
		}
	}
	var result int32
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		result = max(result, vec.Vec_IntEntry(p.vEdgeDelay, int(Gia_ManCi(p, i).id)))
	}
	return result
}
