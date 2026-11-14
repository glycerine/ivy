package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func giaDfsTiming(p *Gia_Man_t) *tim.Tim_Man_t {
	if p.PManTime == nil {
		return nil
	}
	timing, ok := p.PManTime.(*tim.Tim_Man_t)
	if !ok {
		panic("gia: invalid timing manager")
	}
	return timing
}
func Gia_ManDfsCollect_rec(p *Gia_Man_t, obj *Gia_Obj_t, objects *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsCi(obj) != 0 {
		tm := giaDfsTiming(p)
		if tm != nil {
			box := tim.Tim_ManBoxForCi(tm, int(Gia_ObjCioId(obj)))
			if box >= 0 {
				first := tim.Tim_ManBoxOutputFirst(tm, box)
				for i := 0; i < tim.Tim_ManBoxOutputNum(tm, box); i++ {
					Gia_ObjSetTravIdCurrent(p, Gia_ManCi(p, int32(first+i)))
				}
				first = tim.Tim_ManBoxInputFirst(tm, box)
				for i := 0; i < tim.Tim_ManBoxInputNum(tm, box); i++ {
					Gia_ManDfsCollect_rec(p, Gia_ManCo(p, int32(first+i)), objects)
				}
				vec.Vec_IntPush(objects, int32(-box-1))
			}
		}
		return
	}
	if Gia_ObjIsCo(obj) != 0 {
		Gia_ManDfsCollect_rec(p, Gia_ObjFanin0(obj), objects)
		return
	}
	if Gia_ObjIsAnd(obj) != 0 {
		if Gia_ManHasMapping(p) != 0 {
			if Gia_ObjIsLut(p, obj.id) == 0 {
				panic("gia: missing LUT mapping")
			}
			for _, id := range Gia_ObjLutFanins(p, obj.id) {
				Gia_ManDfsCollect_rec(p, Gia_ManObj(p, id), objects)
			}
		} else {
			Gia_ManDfsCollect_rec(p, Gia_ObjFanin0(obj), objects)
			Gia_ManDfsCollect_rec(p, Gia_ObjFanin1(obj), objects)
		}
		vec.Vec_IntPush(objects, obj.id)
		return
	}
	if Gia_ObjIsConst0(obj) == 0 {
		panic("gia: invalid DFS object")
	}
}
func Gia_ManDfsCollect(p *Gia_Man_t) *Vec_Int_t {
	objects := vec.Vec_IntAlloc(int(p.nObjs))
	Gia_ManIncrementTravId(p)
	for _, id := range vec.Vec_IntArray(p.vCos) {
		Gia_ManDfsCollect_rec(p, Gia_ManObj(p, id), objects)
	}
	for _, id := range vec.Vec_IntArray(p.vCis) {
		Gia_ManDfsCollect_rec(p, Gia_ManObj(p, id), objects)
	}
	return objects
}
func Gia_ManDfsArrivals(p *Gia_Man_t, objects *Vec_Int_t) *Vec_Int_t {
	tm := giaDfsTiming(p)
	times := vec.Vec_IntStartFull(int(p.nObjs))
	vec.Vec_IntWriteEntry(times, 0, 0)
	if tm != nil {
		tim.Tim_ManIncrementTravId(tm)
	}
	for i, id := range vec.Vec_IntArray(p.vCis) {
		if tm == nil {
			vec.Vec_IntWriteEntry(times, int(id), 0)
		} else if i < tim.Tim_ManPiNum(tm) {
			vec.Vec_IntWriteEntry(times, int(id), int32(tim.Tim_ManGetCiArrival(tm, i)))
		}
	}
	for _, entry := range vec.Vec_IntArray(objects) {
		if entry < 0 {
			if tm == nil {
				panic("gia: timing box missing manager")
			}
			box := int(-entry - 1)
			first := tim.Tim_ManBoxInputFirst(tm, box)
			for i := 0; i < tim.Tim_ManBoxInputNum(tm, box); i++ {
				obj := Gia_ManCo(p, int32(first+i))
				value := vec.Vec_IntEntry(times, int(Gia_ObjFanin0(obj).id))
				if value < 0 {
					panic("gia: missing box input arrival")
				}
				tim.Tim_ManSetCoArrival(tm, int(Gia_ObjCioId(obj)), float32(value))
			}
			first = tim.Tim_ManBoxOutputFirst(tm, box)
			for i := 0; i < tim.Tim_ManBoxOutputNum(tm, box); i++ {
				obj := Gia_ManCi(p, int32(first+i))
				value := int32(tim.Tim_ManGetCiArrival(tm, int(Gia_ObjCioId(obj))))
				if value < 0 {
					panic("gia: invalid box output arrival")
				}
				vec.Vec_IntWriteEntry(times, int(obj.id), value)
			}
			continue
		}
		if entry == 0 {
			panic("gia: invalid DFS entry")
		}
		var arrival int32
		obj := Gia_ManObj(p, entry)
		if Gia_ManHasMapping(p) != 0 {
			if Gia_ObjIsLut(p, entry) == 0 {
				panic("gia: missing LUT mapping")
			}
			for _, id := range Gia_ObjLutFanins(p, entry) {
				value := vec.Vec_IntEntry(times, int(id))
				if value < 0 {
					panic("gia: missing LUT arrival")
				}
				arrival = max(arrival, value)
			}
		} else {
			a, b := vec.Vec_IntEntry(times, int(Gia_ObjFaninId0(obj, entry))), vec.Vec_IntEntry(times, int(Gia_ObjFaninId1(obj, entry)))
			if a < 0 || b < 0 {
				panic("gia: missing fanin arrival")
			}
			arrival = max(a, b)
		}
		vec.Vec_IntWriteEntry(times, int(entry), arrival+10)
	}
	return times
}
func Gia_ManDfsUpdateRequired(times *Vec_Int_t, id, required int32) {
	value := vec.Vec_IntEntryP(times, int(id))
	if *value == -1 || *value > required {
		*value = required
	}
}
func Gia_ManDfsRequireds(p *Gia_Man_t, objects *Vec_Int_t, required int32) *Vec_Int_t {
	tm := giaDfsTiming(p)
	times := vec.Vec_IntStartFull(int(p.nObjs))
	vec.Vec_IntWriteEntry(times, 0, 0)
	if tm != nil {
		tim.Tim_ManIncrementTravId(tm)
	}
	limit := 0
	if tm != nil {
		limit = int(Gia_ManCoNum(p)) - tim.Tim_ManPoNum(tm)
	}
	for i, id := range vec.Vec_IntArray(p.vCos) {
		if i < limit {
			continue
		}
		obj := Gia_ManObj(p, id)
		if tm != nil {
			tim.Tim_ManSetCoRequired(tm, i, float32(required))
		}
		Gia_ManDfsUpdateRequired(times, Gia_ObjFanin0(obj).id, required)
	}
	data := vec.Vec_IntArray(objects)
	for j := len(data) - 1; j >= 0; j-- {
		entry := data[j]
		if entry < 0 {
			if tm == nil {
				panic("gia: timing box missing manager")
			}
			box := int(-entry - 1)
			first := tim.Tim_ManBoxOutputFirst(tm, box)
			for i := 0; i < tim.Tim_ManBoxOutputNum(tm, box); i++ {
				obj := Gia_ManCi(p, int32(first+i))
				r := vec.Vec_IntEntry(times, int(obj.id))
				if r == -1 {
					r = required
				}
				if r < 0 {
					panic("gia: invalid box required time")
				}
				tim.Tim_ManSetCiRequired(tm, int(Gia_ObjCioId(obj)), float32(r))
			}
			first = tim.Tim_ManBoxInputFirst(tm, box)
			for i := 0; i < tim.Tim_ManBoxInputNum(tm, box); i++ {
				obj := Gia_ManCo(p, int32(first+i))
				r := int32(tim.Tim_ManGetCoRequired(tm, int(Gia_ObjCioId(obj))))
				if r < 0 {
					panic("gia: invalid box required time")
				}
				Gia_ManDfsUpdateRequired(times, Gia_ObjFanin0(obj).id, r)
			}
			continue
		}
		if entry == 0 {
			panic("gia: invalid DFS entry")
		}
		r := vec.Vec_IntEntry(times, int(entry)) - 10
		if r < 0 {
			panic("gia: negative node required time")
		}
		if Gia_ManHasMapping(p) != 0 {
			if Gia_ObjIsLut(p, entry) == 0 {
				panic("gia: missing LUT mapping")
			}
			for _, id := range Gia_ObjLutFanins(p, entry) {
				Gia_ManDfsUpdateRequired(times, id, r)
			}
		} else {
			obj := Gia_ManObj(p, entry)
			Gia_ManDfsUpdateRequired(times, Gia_ObjFaninId0(obj, entry), r)
			Gia_ManDfsUpdateRequired(times, Gia_ObjFaninId1(obj, entry), r)
		}
	}
	return times
}
func Gia_ManDfsSlacks(p *Gia_Man_t) *Vec_Int_t {
	slack := vec.Vec_IntStartFull(int(p.nObjs))
	objects := Gia_ManDfsCollect(p)
	if vec.Vec_IntSize(objects) == 0 {
		return slack
	}
	arrivals := Gia_ManDfsArrivals(p, objects)
	required := Gia_ManDfsRequireds(p, objects, vec.Vec_IntFindMax(arrivals))
	for id, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ManHasMapping(p) != 0 && Gia_ObjIsLut(p, int32(id)) == 0 {
			continue
		}
		a, r := vec.Vec_IntEntry(arrivals, id), vec.Vec_IntEntry(required, id)
		if a > r {
			panic("gia: negative timing slack")
		}
		vec.Vec_IntWriteEntry(slack, id, r-a)
	}
	return slack
}
func Gia_ManDfsSlacksPrint(p *Gia_Man_t) {
	slacks := Gia_ManDfsSlacks(p)
	if vec.Vec_IntSize(slacks) == 0 {
		fmt.Println("Network contains no internal objects.")
		return
	}
	var maximum int32
	for _, v := range vec.Vec_IntArray(slacks) {
		if v >= 0 {
			maximum = max(maximum, v/10)
		}
	}
	counts := make([]int32, maximum+1)
	var total int32
	for _, v := range vec.Vec_IntArray(slacks) {
		if v >= 0 {
			counts[v/10]++
			total++
		}
	}
	if total == 0 {
		panic("gia: no timed nodes")
	}
	for i, n := range counts {
		fmt.Printf("Slack range %3d = [%4d, %4d)   Nodes = %5d  (%6.2f %%) \n", i, 10*i, 10*(i+1), n, 100*float64(n)/float64(total))
	}
}
