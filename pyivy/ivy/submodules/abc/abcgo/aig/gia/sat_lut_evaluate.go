package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Sbl_ManEvaluateMappingEdge(p *Sbl_Man_t, delay int32) int32 {
	started := time.Now()
	defer func() { p.timeTime += time.Since(started) }()
	vec.Vec_IntClear(p.vPath)
	Sbl_ManGetCurrentMapping(p)
	actual := Gia_ManEvalWindow(p.pGia, p.vLeaves, p.vAnds, p.vWindow, p.vPolar, 1)
	if actual <= delay {
		return 1
	}
	for i := 0; i < vec.Vec_WecSize(p.vWindow); i++ {
		if vec.Vec_IntSize(vec.Vec_WecEntry(p.vWindow, i)) > 0 {
			vec.Vec_IntPush(p.vPath, int32(i)*2+1)
		}
	}
	return 0
}
func Sbl_ManEvaluateMapping(p *Sbl_Man_t, delay int32) int32 {
	if p.pGia.VEdge1 != nil {
		return Sbl_ManEvaluateMappingEdge(p, delay)
	}
	started := time.Now()
	defer func() { p.timeTime += time.Since(started) }()
	vec.Vec_IntClear(p.vPath)
	Sbl_ManCreateTiming(p, delay)
	Sbl_ManGetCurrentMapping(p)
	for i, id := range vec.Vec_IntArray(p.vAnds) {
		vec.Vec_IntWriteEntry(p.vArrs, int(id), Sbl_ManComputeDelay(p, id, vec.Vec_WecEntry(p.vWindow, i)))
	}
	var bad int32 = -1
	for _, id := range vec.Vec_IntArray(p.vRoots) {
		if vec.Vec_IntEntry(p.vArrs, int(id)) > vec.Vec_IntEntry(p.vReqs, int(id)) {
			bad = id
			break
		}
	}
	if bad < 0 {
		return 1
	}
	index := vec.Vec_IntFind(p.vAnds, bad)
	if index < 0 {
		panic("gia: critical root not in SAT window")
	}
	for {
		vec.Vec_IntPush(p.vPath, 2*int32(index)+1)
		bad = Sbl_ManCriticalFanin(p, bad, vec.Vec_WecEntry(p.vWindow, index))
		if bad <= 0 {
			panic("gia: missing critical SAT window fanin")
		}
		index = vec.Vec_IntFind(p.vAnds, bad)
		if index < 0 {
			break
		}
	}
	return 0
}
