package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Gia_ObjSimCollect(p *Gia_SimRsbMan_t) {
	vec.Vec_IntClear(p.vFanins2)
	if vec.Vec_IntSize(p.vFanins) == 0 {
		panic("gia: empty candidate frontier")
	}
	g := p.pGia
	for _, id := range vec.Vec_IntArray(p.vFanins) {
		obj := Gia_ManObj(g, id)
		if Gia_ObjIsAnd(obj) != 0 {
			for _, fan := range []int32{Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)} {
				if Gia_ObjIsTravIdCurrentId(g, fan) == 0 {
					vec.Vec_IntPush(p.vFanins2, fan)
				}
			}
		}
		for i := int32(0); i < Gia_ObjFanoutNumId(g, id); i++ {
			fan := Gia_ObjFanoutId(g, id, i)
			if Gia_ObjIsAnd(Gia_ManObj(g, fan)) != 0 && Gia_ObjIsTravIdCurrentId(g, fan) == 0 {
				vec.Vec_IntPush(p.vFanins2, fan)
			}
		}
	}
}
func Gia_ObjSimCands(p *Gia_SimRsbMan_t, id, limit int32) *Vec_Int_t {
	if id <= 0 || Gia_ObjIsAnd(Gia_ManObj(p.pGia, id)) == 0 || limit < 0 {
		panic("gia: invalid resubstitution candidate request")
	}
	vec.Vec_IntClear(p.vCands)
	vec.Vec_IntFill(p.vFanins, 1, id)
	for vec.Vec_IntSize(p.vFanins) > 0 && vec.Vec_IntSize(p.vCands) < int(limit) {
		for _, v := range vec.Vec_IntArray(p.vFanins) {
			Gia_ObjSetTravIdCurrentId(p.pGia, v)
		}
		Gia_ObjSimCollect(p)
		vec.Vec_IntAppend(p.vCands, p.vFanins2)
		p.vFanins, p.vFanins2 = p.vFanins2, p.vFanins
	}
	if vec.Vec_IntSize(p.vCands) > int(limit) {
		vec.Vec_IntShrink(p.vCands, int(limit))
	}
	return p.vCands
}
func Gia_ObjSimRsb(p *Gia_SimRsbMan_t, id, limit, verbose int32, buffers, inverters *int32) int32 {
	tfo := Gia_SimRsbTfo(p, id, -1)
	care := Gia_SimRsbCareSet(p, id, tfo)
	data := vec.Vec_WrdArray(p.vSimsObj)
	target := data[id*p.nWords : (id+1)*p.nWords]
	cands := Gia_ObjSimCands(p, id, limit)
	for w := int32(0); w < p.nWords; w++ {
		p.pFunc[0][w] = care[w] &^ target[w]
		p.pFunc[1][w] = care[w] & target[w]
	}
	var found int32
	for _, cand := range vec.Vec_IntArray(cands) {
		div := data[cand*p.nWords : (cand+1)*p.nWords]
		buf, inv := true, true
		for w := int32(0); w < p.nWords; w++ {
			if div[w]&p.pFunc[0][w] != 0 || ^div[w]&p.pFunc[1][w] != 0 {
				buf = false
			}
			if ^div[w]&p.pFunc[0][w] != 0 || div[w]&p.pFunc[1][w] != 0 {
				inv = false
			}
		}
		if buf {
			*buffers++
			found = 1
			if verbose != 0 {
				fmt.Printf("Level %3d : %d = buf(%d)\n", Gia_ObjLevelId(p.pGia, id), id, cand)
			}
		}
		if inv {
			*inverters++
			found = 1
			if verbose != 0 {
				fmt.Printf("Level %3d : %d = inv(%d)\n", Gia_ObjLevelId(p.pGia, id), id, cand)
			}
		}
	}
	return found
}
func Gia_ManSimRsb(g *Gia_Man_t, limit, verbose int32) int32 {
	start := time.Now()
	p := Gia_SimRsbAlloc(g)
	defer Gia_SimRsbFree(p)
	Gia_ManLevelNum(g)
	var count, buffers, inverters int32
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			count += Gia_ObjSimRsb(p, obj.id, limit, verbose, &buffers, &inverters)
		}
	}
	percentage := float64(0)
	if Gia_ManAndNum(g) != 0 {
		percentage = 100 * float64(count) / float64(Gia_ManAndNum(g))
	}
	fmt.Printf("Can resubstitute %d nodes (%.2f %% out of %d) (Bufs = %d Invs = %d)  Time: %s\n", count, percentage, Gia_ManAndNum(g), buffers, inverters, time.Since(start))
	return count
}
