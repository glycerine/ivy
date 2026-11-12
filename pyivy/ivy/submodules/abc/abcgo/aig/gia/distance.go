package gia

import "fmt"

func Gia_ManRingAdd(p *Gia_Man_t, id int32, result, distances *Vec_Int_t, distance int32) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	Vec_IntWriteEntry(distances, int(id), distance)
	Vec_IntPush(result, id)
}
func Gia_ManCollectRing(p *Gia_Man_t, start, result, distances *Vec_Int_t) {
	for _, id := range Vec_IntArray(start) {
		weight := Vec_IntEntry(distances, int(id))
		if weight <= 0 {
			panic("gia: uninitialized ring weight")
		}
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) != 0 {
			for _, fanin := range giaObjectFanins(p, obj) {
				increment := int32(1)
				if Gia_ObjIsBuf(Gia_ManObj(p, fanin)) != 0 {
					increment = 0
				}
				Gia_ManRingAdd(p, fanin, result, distances, weight+increment)
			}
		}
		for i := int32(0); i < Gia_ObjFanoutNumId(p, id); i++ {
			fanout := Gia_ObjFanoutId(p, id, i)
			increment := int32(1)
			if Gia_ObjIsBuf(Gia_ManObj(p, fanout)) != 0 {
				increment = 0
			}
			Gia_ManRingAdd(p, fanout, result, distances, weight+increment)
		}
	}
}

// ComputeDistanceInt preserves source first-visit ring distances, with buffers
// assigned zero entry weight. It is not a general shortest-path relaxation API.
func Gia_ManComputeDistanceInt(p *Gia_Man_t, target int32, objects *Vec_Int_t, verbose int32) *Vec_Int_t {
	start, next := Vec_IntAlloc(100), Vec_IntAlloc(100)
	distances := Vec_IntStart(len(p.pObjs))
	Gia_ManIncrementTravId(p)
	if objects != nil {
		for _, id := range Vec_IntArray(objects) {
			Gia_ObjSetTravIdCurrentId(p, id)
			Vec_IntWriteEntry(distances, int(id), 1)
			Vec_IntPush(start, id)
		}
	} else {
		Gia_ObjSetTravIdCurrentId(p, target)
		Vec_IntWriteEntry(distances, int(target), 1)
		Vec_IntPush(start, target)
	}
	for ring := 0; ; ring++ {
		if verbose != 0 {
			visited := 0
			for _, distance := range Vec_IntArray(distances) {
				if distance != 0 {
					visited++
				}
			}
			fmt.Printf("Ring %2d : %6d\n", ring, visited)
		}
		Gia_ManCollectRing(p, start, next, distances)
		if Vec_IntSize(next) == 0 {
			break
		}
		Vec_IntClear(start)
		start, next = next, start
	}
	return distances
}
func Gia_ManComputeDistance(p *Gia_Man_t, target int32, objects *Vec_Int_t, verbose int32) *Vec_Int_t {
	if p.vFanoutNums == nil {
		Gia_ManStaticFanoutStart(p)
		defer Gia_ManStaticFanoutStop(p)
	}
	return Gia_ManComputeDistanceInt(p, target, objects, verbose)
}
