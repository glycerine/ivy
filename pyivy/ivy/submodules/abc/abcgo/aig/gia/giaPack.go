package gia

import (
	"fmt"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
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
	// collect LUTs
	vLuts := vec.Vec_IntAlloc(int(Gia_ManAndNum(p)))
	for Id := int32(1); Id < p.nObjs; Id++ {
		if Gia_ObjIsLut(p, Id) != 0 {
			vec.Vec_IntPush(vLuts, Id)
		}
	}
	// propagate distance
	vDist := vec.Vec_IntStart(int(p.nObjs))
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		vec.Vec_IntWriteEntry(vDist, int(Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1)
	}
	for i := vec.Vec_IntSize(vLuts) - 1; i >= 0; i-- {
		Id := vec.Vec_IntEntry(vLuts, i)
		Dist := 1 + vec.Vec_IntEntry(vDist, int(Id))
		for _, iFan := range Gia_ObjLutFanins(p, Id) {
			vec.Vec_IntWriteEntry(vDist, int(iFan), max(vec.Vec_IntEntry(vDist, int(iFan)), Dist))
		}
	}
	// sort LUTs by distance
	k := 0
	for _, Id := range vec.Vec_IntArray(vLuts) {
		vec.Vec_IntWriteEntry(vDist, k, -vec.Vec_IntEntry(vDist, int(Id)))
		k++
	}
	vec.Vec_IntShrink(vDist, k)
	pPerm := abcutil.Abc_MergeSortCost(vec.Vec_IntArray(vDist), int32(vec.Vec_IntSize(vDist)))
	// collect
	vOrder := vec.Vec_IntAlloc(vec.Vec_IntSize(vLuts))
	for i := 0; i < vec.Vec_IntSize(vLuts); i++ {
		vec.Vec_IntPush(vOrder, vec.Vec_IntEntry(vLuts, int(pPerm[i])))
	}
	return vOrder
}

// Gia_ManLutPacking is the LUT packing algorithm of giaPack.c. Delays and
// Perm live across iterations as C's stack arrays do: a one-input LUT
// compares with the Delays[1] left by an earlier LUT.
func Gia_ManLutPacking(p *Gia_Man_t, nBlockSize, DelayRoute, DelayDir, fVerbose int32) {
	var Delays, Perm [32]int32
	var Delay, DelayMax int32
	vOrder := Gia_ManLutCollect(p)
	vDelay := vec.Vec_IntStart(int(p.nObjs))
	vBlock := vec.Vec_IntStart(int(p.nObjs))
	vBSize := vec.Vec_IntAlloc(2 * vec.Vec_IntSize(vOrder) / int(nBlockSize))
	// create blocks
	for _, Id := range vec.Vec_IntArray(vOrder) {
		fanins := Gia_ObjLutFanins(p, Id)
		nSize := len(fanins)
		for k, iFan := range fanins {
			Delays[k] = vec.Vec_IntEntry(vDelay, int(iFan))
			Perm[k] = iFan
		}
		vec.Vec_IntSelectSortCost2Reverse(Perm[:], nSize, Delays[:])
		// check if we can reduce delay by adding it to the same bin as the latest one
		iBlock := vec.Vec_IntEntry(vBlock, int(Perm[0]))
		if Delays[0] > 0 && Delays[0] > Delays[1] && vec.Vec_IntEntry(vBSize, int(iBlock)) < nBlockSize {
			Delay = Delays[0] + DelayDir
			vec.Vec_IntWriteEntry(vBlock, int(Id), iBlock)
			vec.Vec_IntAddToEntry(vBSize, int(iBlock), 1)
		} else { // clean new block
			Delay = Delays[0] + DelayRoute
			vec.Vec_IntWriteEntry(vBlock, int(Id), int32(vec.Vec_IntSize(vBSize)))
			vec.Vec_IntPush(vBSize, 1)
		}
		// calculate delay
		for k := 1; k < nSize; k++ {
			Delay = max(Delay, Delays[k]+DelayRoute)
		}
		vec.Vec_IntWriteEntry(vDelay, int(Id), Delay)
		DelayMax = max(DelayMax, Delay)
	}
	// create packing info
	vPacking := vec.Vec_IntAlloc(vec.Vec_IntSize(vBSize) + vec.Vec_IntSize(vOrder) + 1)
	vec.Vec_IntPush(vPacking, int32(vec.Vec_IntSize(vBSize)))
	// create starting places for each block
	vStarts := vec.Vec_IntAlloc(vec.Vec_IntSize(vBSize))
	for _, nSize := range vec.Vec_IntArray(vBSize) {
		vec.Vec_IntPush(vPacking, nSize)
		vec.Vec_IntPush(vStarts, int32(vec.Vec_IntSize(vPacking)))
		for j := int32(0); j < nSize; j++ {
			vec.Vec_IntPush(vPacking, -1)
		}
	}
	// collect LUTs from the block
	for i := vec.Vec_IntSize(vOrder) - 1; i >= 0; i-- {
		Id := vec.Vec_IntEntry(vOrder, i)
		Block := vec.Vec_IntEntry(vBlock, int(Id))
		Start := vec.Vec_IntEntry(vStarts, int(Block))
		vec.Vec_IntWriteEntry(vPacking, int(Start), Id)
		vec.Vec_IntAddToEntry(vStarts, int(Block), 1)
	}
	p.VPacking = vPacking
	fmt.Printf("Global delay = %d.\n", DelayMax)
}
