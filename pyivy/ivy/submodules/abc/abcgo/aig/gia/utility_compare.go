package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/util"
	"time"
)

func Gia_ManCollectObjs_rec(p *Gia_Man_t, id int32, objects *Vec_Int_t, limit int32) {
	if int32(Vec_IntSize(objects)) >= limit || Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) != 0 {
		Gia_ManCollectObjs_rec(p, Gia_ObjFaninId0p(p, obj), objects, limit)
		if int32(Vec_IntSize(objects)) >= limit {
			return
		}
		Gia_ManCollectObjs_rec(p, Gia_ObjFaninId1p(p, obj), objects, limit)
		if int32(Vec_IntSize(objects)) >= limit {
			return
		}
	}
	Vec_IntPush(objects, id)
}

// Gia_ManComputePoTruthTables retains ABC's batching benchmark; the original
// routine reports traversal statistics and returns no truth-table allocation.
func Gia_ManComputePoTruthTables(p *Gia_Man_t, maxBytes int32) []uint32 {
	vars := Gia_ManPiNum(p)
	words := util.Abc_TruthWordNum(vars)
	if words <= 0 || maxBytes < 4*words {
		panic("gia: insufficient truth-table buffer")
	}
	capacity := maxBytes / (4 * words)
	fmt.Printf("Var = %d. Words = %d. Truths = %d.\n", vars, words, capacity)
	start := time.Now()
	objects := Vec_IntAlloc(int(capacity))
	var total, rounds int32
	Gia_ManIncrementTravId(p)
	for i := int32(0); i < Gia_ManPoNum(p); i++ {
		Gia_ManCollectObjs_rec(p, Gia_ObjFaninId0p(p, Gia_ManPo(p, i)), objects, capacity)
		if int32(Vec_IntSize(objects)) == capacity {
			rounds++
			total += int32(Vec_IntSize(objects))
			Vec_IntClear(objects)
			Gia_ManIncrementTravId(p)
		}
	}
	total += int32(Vec_IntSize(objects))
	fmt.Printf("Rounds = %d. Objects = %d. Total = %d.   Time = %.2f sec\n", rounds, Gia_ManObjNum(p), total, time.Since(start).Seconds())
	return nil
}
func Gia_ManCompare(a, b *Gia_Man_t) int32 {
	if Gia_ManObjNum(a) != Gia_ManObjNum(b) {
		fmt.Println("AIGs have different number of objects.")
		return 0
	}
	Gia_ManCleanValue(a)
	Gia_ManCleanValue(b)
	for i, x := range a.pObjs {
		y := b.pObjs[i]
		if x.iDiff0 != y.iDiff0 || x.iDiff1 != y.iDiff1 || x.fCompl0 != y.fCompl0 || x.fCompl1 != y.fCompl1 || x.fMark0 != y.fMark0 || x.fMark1 != y.fMark1 || x.fTerm != y.fTerm || x.fPhase != y.fPhase {
			fmt.Printf("Objects %d are different.\n", i)
			return 0
		}
		if a.PReprs != nil && b.PReprs != nil && a.PReprs[i] != b.PReprs[i] {
			fmt.Printf("Representatives of objects %d are different.\n", i)
			return 0
		}
	}
	return 1
}
