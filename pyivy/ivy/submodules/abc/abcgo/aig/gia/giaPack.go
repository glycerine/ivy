package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

// Gia_ManLutCollect2 collects reachable LUTs in reverse breadth-first order
// from output drivers, placing each LUT after its fanin LUTs.
func Gia_ManLutCollect2(p *Gia_Man_t) *Vec_Int_t {
	order := vec.Vec_IntAlloc(int(Gia_ManLutNum(p)))
	seen := make([]bool, int(p.nObjs))
	add := func(id int32) {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) == 0 || seen[id] {
			return
		}
		if Gia_ObjIsLut(p, id) == 0 {
			panic("gia: reachable AND is not a LUT")
		}
		seen[id] = true
		vec.Vec_IntPush(order, id)
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		add(Gia_ObjFaninId0(Gia_ManObj(p, id), id))
	}
	for i := 0; i < vec.Vec_IntSize(order); i++ {
		id := vec.Vec_IntEntry(order, i)
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			add(fanin)
		}
	}
	vec.Vec_IntReverseOrder(order)
	return order
}

// Gia_ManLutCollect orders LUTs by decreasing distance from the outputs.
func Gia_ManLutCollect(p *Gia_Man_t) *Vec_Int_t {
	ids := make([]int32, 0, int(Gia_ManLutNum(p)))
	for id := int32(0); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			ids = append(ids, id)
		}
	}
	distance := make([]int32, int(p.nObjs))
	for _, id := range vec.Vec_IntArray(p.vCos) {
		distance[Gia_ObjFaninId0(Gia_ManObj(p, id), id)] = 1
	}
	for i := len(ids) - 1; i >= 0; i-- {
		id := ids[i]
		for _, fanin := range Gia_ObjLutFanins(p, id) {
			distance[fanin] = max(distance[fanin], distance[id]+1)
		}
	}
	sort.SliceStable(ids, func(i, j int) bool { return distance[ids[i]] > distance[ids[j]] })
	order := vec.Vec_IntAlloc(len(ids))
	for _, id := range ids {
		vec.Vec_IntPush(order, id)
	}
	return order
}

// Gia_ManLutPacking groups LUTs into blocks and records the resulting packing.
func Gia_ManLutPacking(p *Gia_Man_t, blockSize, delayRoute, delayDir, verbose int32) {
	if blockSize <= 0 {
		panic("gia: invalid LUT block size")
	}
	_ = verbose
	order := vec.Vec_IntArray(Gia_ManLutCollect(p))
	delays := make([]int32, int(p.nObjs))
	blocks := make([]int, int(p.nObjs))
	sizes := make([]int, 0)
	maxDelay := int32(0)
	type arrival struct{ id, delay int32 }
	for _, id := range order {
		fanins := Gia_ObjLutFanins(p, id)
		if len(fanins) > 32 {
			panic("gia: LUT has more than 32 inputs")
		}
		arrivals := make([]arrival, len(fanins))
		for i, f := range fanins {
			arrivals[i] = arrival{f, delays[f]}
		}
		sort.SliceStable(arrivals, func(i, j int) bool { return arrivals[i].delay > arrivals[j].delay })
		first, second := int32(0), int32(0)
		if len(arrivals) > 0 {
			first = arrivals[0].delay
		}
		if len(arrivals) > 1 {
			second = arrivals[1].delay
		}
		var delay int32
		if len(arrivals) > 0 && first > 0 && first > second && sizes[blocks[arrivals[0].id]] < int(blockSize) {
			block := blocks[arrivals[0].id]
			blocks[id] = block
			sizes[block]++
			delay = first + delayDir
		} else {
			blocks[id] = len(sizes)
			sizes = append(sizes, 1)
			delay = first + delayRoute
		}
		for k := 1; k < len(arrivals); k++ {
			delay = max(delay, arrivals[k].delay+delayRoute)
		}
		delays[id] = delay
		maxDelay = max(maxDelay, delay)
	}
	packing := vec.Vec_IntAlloc(1 + len(sizes) + len(order))
	vec.Vec_IntPush(packing, int32(len(sizes)))
	starts := make([]int, len(sizes))
	for i, size := range sizes {
		vec.Vec_IntPush(packing, int32(size))
		starts[i] = vec.Vec_IntSize(packing)
		for j := 0; j < size; j++ {
			vec.Vec_IntPush(packing, -1)
		}
	}
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		block := blocks[id]
		vec.Vec_IntWriteEntry(packing, starts[block], id)
		starts[block]++
	}
	p.VPacking = packing
	fmt.Printf("Global delay = %d.\n", maxDelay)
}
