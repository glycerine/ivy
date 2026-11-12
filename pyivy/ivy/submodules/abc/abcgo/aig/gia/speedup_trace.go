package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

func Gia_ManDelayTraceLut(p *Gia_Man_t) float32 {
	var timing *tim.Tim_Man_t
	if p.pManTime != nil {
		var ok bool
		timing, ok = p.pManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
	}
	Gia_ManTimeStart(p)
	Gia_ManLevelNum(p)
	if timing != nil {
		tim.Tim_ManIncrementTravId(timing)
	}
	for _, obj := range p.pObjs {
		id := obj.id
		if Gia_ObjIsCi(obj) == 0 && Gia_ObjIsCo(obj) == 0 && Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		arrival := Gia_ObjComputeArrival(p, id, 1)
		if Gia_ObjIsCi(obj) != 0 && timing != nil {
			arrival = tim.Tim_ManGetCiArrival(timing, int(Gia_ObjCioId(obj)))
		}
		if Gia_ObjIsCo(obj) != 0 && timing != nil {
			tim.Tim_ManSetCoArrival(timing, int(Gia_ObjCioId(obj)), arrival)
		}
		Gia_ObjSetTimeArrival(p, id, arrival)
	}
	arrival := float32(math.Inf(-1))
	for _, id := range vec.Vec_IntArray(p.vCos) {
		obj := Gia_ManObj(p, id)
		current := Gia_ObjTimeArrival(p, Gia_ObjFaninId0(obj, id))
		Gia_ObjSetTimeArrival(p, id, current)
		arrival = max(arrival, current)
	}
	if timing != nil {
		tim.Tim_ManIncrementTravId(timing)
		tim.Tim_ManInitPoRequiredAll(timing, arrival)
	} else {
		for _, id := range vec.Vec_IntArray(p.vCos) {
			Gia_ObjSetTimeRequired(p, id, arrival)
		}
	}
	for _, id := range vec.Vec_IntArray(Gia_ManOrderReverse(p)) {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsLut(p, id) != 0 {
			Gia_ObjPropagateRequired(p, id, 1)
		} else if Gia_ObjIsCi(obj) != 0 {
			if timing != nil {
				tim.Tim_ManSetCiRequired(timing, int(Gia_ObjCioId(obj)), Gia_ObjTimeRequired(p, id))
			}
		} else if Gia_ObjIsCo(obj) != 0 {
			if timing != nil {
				Gia_ObjSetTimeRequired(p, id, tim.Tim_ManGetCoRequired(timing, int(Gia_ObjCioId(obj))))
			}
			fan := Gia_ObjFaninId0(obj, id)
			if Gia_ObjTimeRequired(p, fan) > Gia_ObjTimeRequired(p, id) {
				Gia_ObjSetTimeRequired(p, fan, Gia_ObjTimeRequired(p, id))
			}
		}
		slack := Gia_ObjTimeRequired(p, id) - Gia_ObjTimeArrival(p, id)
		if slack < -0.01 {
			panic("gia: negative LUT timing slack")
		}
		Gia_ObjSetTimeSlack(p, id, max(slack, 0))
	}
	return arrival
}
func Gia_LutDelayTraceTCEdges(p *Gia_Man_t, id int32, delta float32) uint32 {
	fans := Gia_ObjLutFanins(p, id)
	if len(fans) > 32 {
		panic("gia: too many LUT pins")
	}
	order, delays := giaLutPinAssignments(p, id, 1)
	required := Gia_ObjTimeRequired(p, id)
	var mask uint32
	for k, index := range order {
		if required < Gia_ObjTimeArrival(p, fans[index])+delays[k]+delta {
			mask |= uint32(1) << uint(index)
		}
	}
	return mask
}
