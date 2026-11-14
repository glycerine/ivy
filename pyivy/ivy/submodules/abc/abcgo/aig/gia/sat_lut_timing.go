package gia

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func Sbl_ManComputeDelay(p *Sbl_Man_t, _ int32, fanins *Vec_Int_t) int32 {
	var delay int32
	for _, id := range vec.Vec_IntArray(fanins) {
		delay = max(delay, vec.Vec_IntEntry(p.vArrs, int(id))+1)
	}
	return delay
}
func Sbl_ManCriticalFanin(p *Sbl_Man_t, id int32, fanins *Vec_Int_t) int32 {
	delay := vec.Vec_IntEntry(p.vArrs, int(id))
	for _, fan := range vec.Vec_IntArray(fanins) {
		if vec.Vec_IntEntry(p.vArrs, int(fan))+1 == delay {
			return fan
		}
	}
	return -1
}
func Sbl_ManCreateTiming(p *Sbl_Man_t, start int32) int32 {
	g := p.pGia
	vec.Vec_IntFill(p.vArrs, int(g.nObjs), 0)
	var timing *tim.Tim_Man_t
	if g.PManTime != nil {
		var ok bool
		timing, ok = g.PManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: invalid timing manager")
		}
		if tim.Tim_ManBoxNum(timing) == 0 {
			timing = nil
		}
	}
	delayMax := start
	update := func(id int32) {
		if Gia_ObjIsLut2(g, id) != 0 {
			delay := Sbl_ManComputeDelay(p, id, Gia_ObjLutFanins2(g, id))
			vec.Vec_IntWriteEntry(p.vArrs, int(id), delay)
			delayMax = max(delayMax, delay)
		}
	}
	var order []int32
	if timing != nil {
		order = vec.Vec_IntArray(Gia_ManOrderWithBoxes(g))
		tim.Tim_ManIncrementTravId(timing)
		for _, id := range order {
			obj := Gia_ManObj(g, id)
			switch {
			case Gia_ObjIsAnd(obj) != 0:
				update(id)
			case Gia_ObjIsCi(obj) != 0:
				vec.Vec_IntWriteEntry(p.vArrs, int(id), int32(tim.Tim_ManGetCiArrival(timing, int(Gia_ObjCioId(obj)))))
			case Gia_ObjIsCo(obj) != 0:
				tim.Tim_ManSetCoArrival(timing, int(Gia_ObjCioId(obj)), float32(vec.Vec_IntEntry(p.vArrs, int(Gia_ObjFaninId0(obj, id)))))
			}
		}
	} else {
		for id := int32(1); id < g.nObjs; id++ {
			update(id)
		}
	}
	vec.Vec_IntFill(p.vReqs, int(g.nObjs), util.ABC_INFINITY)
	lower := func(id, value int32) {
		if value < vec.Vec_IntEntry(p.vReqs, int(id)) {
			vec.Vec_IntWriteEntry(p.vReqs, int(id), value)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		lower(Gia_ObjFaninId0p(g, Gia_ManCo(g, i)), delayMax)
	}
	propagate := func(id int32) {
		if Gia_ObjIsLut2(g, id) != 0 {
			required := vec.Vec_IntEntry(p.vReqs, int(id)) - 1
			for _, fan := range vec.Vec_IntArray(Gia_ObjLutFanins2(g, id)) {
				lower(fan, required)
			}
		}
	}
	if timing != nil {
		tim.Tim_ManIncrementTravId(timing)
		tim.Tim_ManInitPoRequiredAll(timing, float32(delayMax))
		for i := len(order) - 1; i >= 0; i-- {
			id := order[i]
			obj := Gia_ManObj(g, id)
			switch {
			case Gia_ObjIsAnd(obj) != 0:
				propagate(id)
			case Gia_ObjIsCi(obj) != 0:
				tim.Tim_ManSetCiRequired(timing, int(Gia_ObjCioId(obj)), float32(vec.Vec_IntEntry(p.vReqs, int(id))))
			case Gia_ObjIsCo(obj) != 0:
				vec.Vec_IntWriteEntry(p.vReqs, int(Gia_ObjFaninId0(obj, id)), int32(tim.Tim_ManGetCoRequired(timing, int(Gia_ObjCioId(obj)))))
			}
		}
	} else {
		for id := g.nObjs - 1; id > 0; id-- {
			propagate(id)
		}
	}
	return delayMax
}
