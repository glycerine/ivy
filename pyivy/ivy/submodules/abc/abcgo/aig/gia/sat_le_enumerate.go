package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
)

// sleList returns views of a node's non-unit cuts and, when requested, its unit cut.
func sleList(cuts *Vec_Int_t, id int32, unit bool) [][]int32 {
	data := vec.Vec_IntArray(cuts)
	pos := int(data[id])
	count := int(data[pos])
	pos++
	if unit {
		count++
	}
	result := make([][]int32, 0, count)
	for i := 0; i < count; i++ {
		n := int(Sle_CutSize(data[pos:]))
		result = append(result, data[pos:pos+n+1])
		pos += n + 1
	}
	return result
}
func Sle_ManCutMerge(p *Gia_Man_t, id int32, cuts, temp *Vec_Int_t, limit int32) int32 {
	obj := Gia_ManObj(p, id)
	a := sleList(cuts, Gia_ObjFaninId0(obj, id), true)
	b := sleList(cuts, Gia_ObjFaninId1(obj, id), true)
	vec.Vec_IntFill(temp, 1, 0)
	candidate := make([]int32, limit+1)
	for _, x := range a {
		for _, y := range b {
			if Sle_CutSize(x)+Sle_CutSize(y) > limit && Sle_CutCountBits(uint32(Sle_CutSign(x)|Sle_CutSign(y))) > limit {
				continue
			}
			if Sle_CutMergeOrder(x, y, candidate, limit) == 0 || Sle_SetLastCutIsContained(temp, candidate) != 0 {
				continue
			}
			Sle_SetAddCut(temp, candidate)
		}
	}
	offset := int32(vec.Vec_IntSize(cuts))
	vec.Vec_IntWriteEntry(cuts, int(id), offset)
	vec.Vec_IntPush(cuts, 0)
	count := int32(0)
	sleEachCut(temp, func(c []int32) bool {
		if Sle_CutIsUsed(c) != 0 {
			vec.Vec_IntPushArray(cuts, c, len(c))
			count++
		}
		return false
	})
	vec.Vec_IntPush(cuts, Sle_CutSetSizeSign(1, 1<<uint(id%28)))
	vec.Vec_IntPush(cuts, id)
	vec.Vec_IntWriteEntry(cuts, int(offset), count)
	return count
}
func Sle_ManComputeCuts(p *Gia_Man_t, limit, verbose int32) *Vec_Int_t {
	if limit < 1 || limit > 6 {
		panic("gia: SAT LUT size must be 1..6")
	}
	cuts := vec.Vec_IntStart(int(p.nObjs))
	temp := vec.Vec_IntAlloc(1000)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		id := Gia_ObjId(p, Gia_ManCi(p, i))
		vec.Vec_IntWriteEntry(cuts, int(id), int32(vec.Vec_IntSize(cuts)))
		vec.Vec_IntPush(cuts, 0)
		vec.Vec_IntPush(cuts, Sle_CutSetSizeSign(1, 1<<uint(id%28)))
		vec.Vec_IntPush(cuts, id)
	}
	var count int32
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) != 0 {
			count += Sle_ManCutMerge(p, id, cuts, temp, limit)
		}
	}
	if verbose != 0 {
		fmt.Printf("Nodes = %d. Cuts = %d. Ints = %d.\n", Gia_ManAndNum(p), count, vec.Vec_IntSize(cuts))
	}
	return cuts
}
func Sle_ManComputeDelayCut(_ *Gia_Man_t, cut []int32, times *Vec_Int_t) int32 {
	var delay int32
	for _, id := range Sle_CutLeaves(cut) {
		delay = max(delay, vec.Vec_IntEntry(times, int(id)))
	}
	return delay + 1
}
func Sle_ManComputeDelayOne(p *Gia_Man_t, id int32, cuts, times *Vec_Int_t) int32 {
	delay := int32(math.MaxInt32)
	for _, cut := range sleList(cuts, id, false) {
		delay = min(delay, Sle_ManComputeDelayCut(p, cut, times))
	}
	vec.Vec_IntWriteEntry(times, int(id), delay)
	return delay
}
func Sle_ManComputeDelay(p *Gia_Man_t, cuts *Vec_Int_t) int32 {
	times := vec.Vec_IntStart(int(p.nObjs))
	var delay int32
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsAnd(Gia_ManObj(p, id)) != 0 {
			delay = max(delay, Sle_ManComputeDelayOne(p, id, cuts, times))
		}
	}
	return delay
}
