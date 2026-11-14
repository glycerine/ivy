package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Kf_ManCreateFaninCounts(g *Gia_Man_t) *Vec_Int_t {
	counts := vec.Vec_IntStart(int(g.nObjs))
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			count := 2 - Gia_ObjIsCi(Gia_ObjFanin0(obj)) - Gia_ObjIsCi(Gia_ObjFanin1(obj))
			vec.Vec_IntWriteEntry(counts, int(obj.id), count)
		}
	}
	return counts
}
func kfSaveNode(p *Kf_Man_t, s *Kf_Set_t, id int32) {
	Kf_ManSaveResults(s.ppCuts[:], s.nCuts, s.pCutBest, p.vTemp)
	vec.Vec_IntWriteEntry(&p.vTime, int(id), s.pCutBest.Delay+1)
	vec.Vec_FltWriteEntry(&p.vArea, int(id), (s.pCutBest.Area+1)/Kf_ObjRefs(p, id))
	if s.pCutBest.nLeaves > 1 {
		Kf_ManStoreAddUnit(p.vTemp, id, Kf_ObjTime(p, id), Kf_ObjArea(p, id))
	}
	Kf_ObjSetCuts(p, id, p.vTemp)
}

type kfJob struct {
	id   int32
	a, b []int32
}
type kfDone struct{ id, worker int32 }

// Kf_ManComputeCuts runs ready nodes on workers; the coordinator alone
// mutates the manager's storage and releases dependent nodes after completion.
func Kf_ManComputeCuts(p *Kf_Man_t) {
	workers := int(p.pPars.NProcNum)
	if workers <= 0 {
		panic("KF requires workers")
	}
	g := p.pGia
	counts := vec.Vec_IntArray(Kf_ManCreateFaninCounts(g))
	fanouts := make([][]int32, g.nObjs)
	var ready []int32
	total := 0
	for i := len(g.pObjs) - 1; i >= 0; i-- {
		obj := g.pObjs[i]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		total++
		if counts[i] == 0 {
			ready = append(ready, obj.id)
		}
		for _, fid := range []int32{Gia_ObjFanin0(obj).id, Gia_ObjFanin1(obj).id} {
			if Gia_ObjIsAnd(Gia_ManObj(g, fid)) != 0 {
				fanouts[fid] = append(fanouts[fid], obj.id)
			}
		}
	}
	jobs := make([]chan kfJob, workers)
	done := make(chan kfDone, workers)
	for i := 0; i < workers; i++ {
		jobs[i] = make(chan kfJob)
		worker := int32(i)
		go func() {
			for job := range jobs[worker] {
				Kf_SetMergeOrder(&p.pSett[worker], job.a, job.b, p.pPars.FAreaOnly, p.pPars.FCutMin)
				done <- kfDone{job.id, worker}
			}
		}()
	}
	defer func() {
		for _, ch := range jobs {
			close(ch)
		}
	}()
	idle := make([]int32, workers)
	for i := range idle {
		idle[i] = int32(i)
	}
	running, finished := 0, 0
	for finished < total {
		for len(ready) > 0 && len(idle) > 0 {
			id := ready[len(ready)-1]
			ready = ready[:len(ready)-1]
			worker := idle[len(idle)-1]
			idle = idle[:len(idle)-1]
			jobs[worker] <- kfJob{id, Kf_ObjCuts0(p, id), Kf_ObjCuts1(p, id)}
			running++
		}
		if running == 0 {
			panic("KF unresolved fanin dependencies")
		}
		result := <-done
		running--
		finished++
		kfSaveNode(p, &p.pSett[result.worker], result.id)
		idle = append(idle, result.worker)
		for _, next := range fanouts[result.id] {
			counts[next]--
			if counts[next] < 0 {
				panic("KF invalid fanin count")
			}
			if counts[next] == 0 {
				ready = append(ready, next)
			}
		}
	}
}
func Kf_ManPrintStats(p *Kf_Man_t, title string) {
	if p.pPars.FVerbose != 0 {
		fmt.Printf("%s : Level =%6d   Area =%9d   Edge =%9d   Time = %s\n", title, p.pPars.Delay, p.pPars.Area, p.pPars.Edge, time.Since(p.clkStart))
	}
}
func Kf_ManComputeMapping(p *Kf_Man_t) {
	g := p.pGia
	for _, id := range vec.Vec_IntArray(g.vCis) {
		Kf_ManStoreStart(p.vTemp, 0)
		Kf_ManStoreAddUnit(p.vTemp, id, 0, 0)
		Kf_ObjSetCuts(p, id, p.vTemp)
	}
	if p.pPars.NProcNum > 0 {
		Kf_ManComputeCuts(p)
	} else {
		for _, obj := range g.pObjs {
			if Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			id := obj.id
			set := &p.pSett[0]
			a, b := Kf_ObjCuts0(p, id), Kf_ObjCuts1(p, id)
			if p.pPars.FCutHashing != 0 {
				Kf_SetMerge(set, a, b, p.pPars.FAreaOnly, p.pPars.FCutMin)
			} else if p.pPars.FCutSimple != 0 {
				Kf_SetMergeSimple(set, a, b, p.pPars.FAreaOnly, p.pPars.FCutMin)
			} else {
				Kf_SetMergeOrder(set, a, b, p.pPars.FAreaOnly, p.pPars.FCutMin)
			}
			kfSaveNode(p, set, id)
		}
	}
	Kf_ManComputeRefs(p)
	Kf_ManPrintStats(p, "Start")
}
func Kf_ManDerive(p *Kf_Man_t) *Gia_Man_t {
	if p.pPars.FCutMin != 0 {
		panic("KF minimized mapping not implemented by original")
	}
	g := p.pGia
	mapping := vec.Vec_IntStart(int(g.nObjs))
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsBuf(obj) != 0 || Gia_ObjRefNum(g, obj) == 0 {
			continue
		}
		cut := Kf_ObjCutBest(p, obj.id)
		vec.Vec_IntWriteEntry(mapping, int(obj.id), int32(vec.Vec_IntSize(mapping)))
		vec.Vec_IntPush(mapping, Kf_CutSize(cut))
		for k := int32(1); k <= Kf_CutSize(cut); k++ {
			vec.Vec_IntPush(mapping, Kf_CutLeaf(cut, k))
		}
		vec.Vec_IntPush(mapping, obj.id)
	}
	g.vMapping = mapping
	return g
}
func Kf_ManSetDefaultPars(p *Jf_Par_t) {
	*p = Jf_Par_t{NLutSize: 6, NCutNum: 8, NRounds: 1, NVerbLimit: 5, DelayTarget: -1, FOptEdge: 1, NLutSizeMax: KF_LEAF_MAX, NCutNumMax: KF_CUT_MAX, NProcNumMax: KF_PROC_MAX}
}
func Kf_ManPerformMapping(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	p := Kf_ManAlloc(g, pars)
	defer Kf_ManFree(p)
	Kf_ManComputeMapping(p)
	return Kf_ManDerive(p)
}
