package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManUpdateMapping(p *Gia_Man_t, nodes *Vec_Int_t, window *Vec_Wec_t) {
	for i, id := range vec.Vec_IntArray(nodes) {
		a, b := vec.Vec_WecEntry(p.vMapping2, int(id)), vec.Vec_WecEntry(window, i)
		*a, *b = *b, *a
	}
}
func Gia_ManEvalWindowInc(p *Gia_Man_t, leaves, nodes *Vec_Int_t, window *Vec_Wec_t, temp *Vec_Int_t, useTwo int32) int32 {
	if vec.Vec_IntSize(nodes) != vec.Vec_WecSize(window) {
		panic("gia: window size mismatch")
	}
	Gia_ManUpdateMapping(p, nodes, window)
	defer Gia_ManUpdateMapping(p, nodes, window)
	Gia_ManCollectTfo(p, leaves, temp)
	vec.Vec_IntReverseOrder(temp)
	var result int32
	for _, id := range vec.Vec_IntArray(temp) {
		if Gia_ObjIsLut(p, id) != 0 {
			result = max(result, Gia_ObjComputeEdgeDelay(p, id, p.vEdgeDelay, p.VEdge1, p.VEdge2, useTwo))
		}
	}
	return result
}
func Gia_ManEvalWindow(p *Gia_Man_t, leaves, nodes *Vec_Int_t, window *Vec_Wec_t, temp *Vec_Int_t, useTwo int32) int32 {
	if vec.Vec_IntSize(nodes) != vec.Vec_WecSize(window) {
		panic("gia: window size mismatch")
	}
	Gia_ManUpdateMapping(p, nodes, window)
	defer Gia_ManUpdateMapping(p, nodes, window)
	return Gia_ManComputeEdgeDelay(p, useTwo)
}
func Edg_ManToMapping(p *Gia_Man_t) {
	if p.VMapping == nil {
		panic("gia: missing LUT mapping")
	}
	p.vMapping2 = vec.Vec_WecStart(int(p.nObjs))
	p.vFanouts2 = vec.Vec_WecStart(int(p.nObjs))
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		for _, fanin := range giaEdgeFanins(p, id) {
			vec.Vec_WecPush(p.vMapping2, int(id), fanin)
			vec.Vec_WecPush(p.vFanouts2, int(fanin), id)
		}
	}
}
func Edg_ObjEvalEdgeDelay(p *Gia_Man_t, id int32, delays *Vec_Int_t) int32 {
	var result int32
	for _, fanin := range vec.Vec_IntArray(Gia_ObjLutFanins2(p, id)) {
		delay := vec.Vec_IntEntry(delays, int(fanin))
		if Gia_ObjHaveEdge(p, id, fanin) == 0 {
			delay++
		}
		result = max(result, delay)
	}
	return result
}
func Edg_ManEvalEdgeDelay(p *Gia_Man_t) int32 {
	if p.VEdge1 == nil || p.VEdge2 == nil {
		panic("gia: missing assigned edges")
	}
	p.vEdgeDelay = vec.Vec_IntStart(int(p.nObjs))
	var result int32
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut2(p, id) == 0 {
			continue
		}
		delay := Edg_ObjEvalEdgeDelay(p, id, p.vEdgeDelay)
		vec.Vec_IntWriteEntry(p.vEdgeDelay, int(id), delay)
		result = max(result, delay)
	}
	return result
}
func Edg_ObjEvalEdgeDelayR(p *Gia_Man_t, id int32, delays *Vec_Int_t) int32 {
	var result int32
	for _, fanout := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vFanouts2, int(id))) {
		delay := vec.Vec_IntEntry(delays, int(fanout))
		if Gia_ObjHaveEdge(p, id, fanout) == 0 {
			delay++
		}
		result = max(result, delay)
	}
	return result
}
func Edg_ManEvalEdgeDelayR(p *Gia_Man_t) int32 {
	if p.VEdge1 == nil || p.VEdge2 == nil {
		panic("gia: missing assigned edges")
	}
	p.vEdgeDelayR = vec.Vec_IntStart(int(p.nObjs))
	var result int32
	for id := p.nObjs - 1; id > 0; id-- {
		if Gia_ObjIsLut2(p, id) == 0 {
			continue
		}
		delay := Edg_ObjEvalEdgeDelayR(p, id, p.vEdgeDelayR)
		vec.Vec_IntWriteEntry(p.vEdgeDelayR, int(id), delay)
		result = max(result, delay)
	}
	return result
}
func Edg_ManCollectCritEdges(p *Gia_Man_t, edges *Vec_Wec_t, delayMax int32) {
	if p.VEdge1 == nil || p.VEdge2 == nil {
		panic("gia: missing assigned edges")
	}
	vec.Vec_WecClear(edges)
	vec.Vec_WecInit(edges, int(delayMax+1))
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut2(p, id) == 0 {
			continue
		}
		forward, reverse := vec.Vec_IntEntry(p.vEdgeDelay, int(id)), vec.Vec_IntEntry(p.vEdgeDelayR, int(id))
		if forward+reverse > delayMax {
			panic("gia: inconsistent critical path delays")
		}
		if forward+reverse == delayMax {
			vec.Vec_WecPush(edges, int(forward), id)
		}
	}
	for i := 1; i < vec.Vec_WecSize(edges); i++ {
		if vec.Vec_IntSize(vec.Vec_WecEntry(edges, i)) == 0 {
			panic("gia: empty critical path level")
		}
	}
}

// Edg_ObjImprove assigns all unassigned critical fanins or fanouts when the
// endpoint capacity permits; an existing critical edge prevents that choice.
func Edg_ObjImprove(p *Gia_Man_t, id, limit, delayMax, verbose int32) int32 {
	var criticalIn, criticalOut []int32
	var edgeIn, edgeOut, noncriticalIn, noncriticalOut, edges int32
	for _, next := range vec.Vec_IntArray(Gia_ObjLutFanins2(p, id)) {
		if Gia_ObjIsAnd(Gia_ManObj(p, next)) == 0 {
			continue
		}
		delay := vec.Vec_IntEntry(p.vEdgeDelayR, int(id)) + vec.Vec_IntEntry(p.vEdgeDelay, int(next))
		if Gia_ObjHaveEdge(p, id, next) != 0 {
			edges++
			if delay > delayMax {
				panic("gia: inconsistent critical path")
			}
			if delay == delayMax {
				edgeIn++
			} else {
				noncriticalIn++
			}
		} else {
			if delay+1 > delayMax {
				panic("gia: inconsistent critical path")
			}
			if delay+1 == delayMax {
				criticalIn = append(criticalIn, next)
			}
		}
	}
	for _, next := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vFanouts2, int(id))) {
		if Gia_ObjIsAnd(Gia_ManObj(p, next)) == 0 {
			panic("gia: non-AND LUT fanout")
		}
		delay := vec.Vec_IntEntry(p.vEdgeDelay, int(id)) + vec.Vec_IntEntry(p.vEdgeDelayR, int(next))
		if Gia_ObjHaveEdge(p, id, next) != 0 {
			edges++
			if delay > delayMax {
				panic("gia: inconsistent critical path")
			}
			if delay == delayMax {
				edgeOut++
			} else {
				noncriticalOut++
			}
		} else {
			if delay+1 > delayMax {
				panic("gia: inconsistent critical path")
			}
			if delay+1 == delayMax {
				criticalOut = append(criticalOut, next)
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("%8d : Edges = %d  Fanins (all %d EC %d ENC %d C %d)  Fanouts (all %d EC %d ENC %d C %d)  ", id, edges, Gia_ObjLutSize2(p, id), edgeIn, noncriticalIn, len(criticalIn), Gia_ObjLutFanoutNum2(p, id), edgeOut, noncriticalOut, len(criticalOut))
	}
	if edges > limit {
		panic("gia: edge limit exceeded")
	}
	if edges == limit {
		if verbose != 0 {
			fmt.Println("Full")
		}
		return 0
	}
	assign := func(candidates []int32, existing int32) bool {
		if existing != 0 || len(candidates) == 0 || int32(len(candidates)) > limit-edges {
			return false
		}
		for _, next := range candidates {
			if Gia_ObjEdgeCount(next, p.VEdge1, p.VEdge2) >= limit {
				return false
			}
		}
		for _, next := range candidates {
			if Gia_ObjEdgeAdd(id, next, p.VEdge1, p.VEdge2) != 0 || Gia_ObjEdgeAdd(next, id, p.VEdge1, p.VEdge2) != 0 {
				panic("gia: edge capacity exceeded")
			}
		}
		return true
	}
	if assign(criticalIn, edgeIn) || assign(criticalOut, edgeOut) {
		if verbose != 0 {
			fmt.Println("Fixed critical edges")
		}
		return 1
	}
	if verbose != 0 {
		fmt.Println("Cannot fix")
	}
	return 0
}

func Edg_ManAssignEdgeNew(p *Gia_Man_t, limit, verbose int32) int32 {
	if limit < 0 || limit > 2 {
		panic("gia: edge limit must be between zero and two")
	}
	if verbose != 0 {
		fmt.Printf("Running edge assignment with E = %d.\n", limit)
	}
	Edg_ManToMapping(p)
	p.VEdge1 = vec.Vec_IntStart(int(p.nObjs))
	p.VEdge2 = vec.Vec_IntStart(int(p.nObjs))
	levels := vec.Vec_WecStart(0)
	var best1, best2 *Vec_Int_t
	best, last := int32(^uint32(0)>>1), -1
	var delay int32
	for i := 0; i < 10000 && i <= last+50; i++ {
		delay = Edg_ManEvalEdgeDelay(p)
		reverse := Edg_ManEvalEdgeDelayR(p)
		if delay != reverse+1 {
			panic("gia: inconsistent forward and reverse edge delays")
		}
		if delay < best {
			best = delay
			last = i
			best1 = vec.Vec_IntDup(p.VEdge1)
			best2 = vec.Vec_IntDup(p.VEdge2)
		}
		if verbose != 0 {
			fmt.Printf("\nIter %4d : Delay = %4d\n", i, delay)
		}
		Edg_ManCollectCritEdges(p, levels, delay)
		vec.Vec_WecSort(levels, 0)
		improved := false
		for k := 0; k < vec.Vec_WecSize(levels) && !improved; k++ {
			for _, id := range vec.Vec_IntArray(vec.Vec_WecEntry(levels, k)) {
				if Edg_ObjImprove(p, id, limit, delay, verbose) != 0 {
					improved = true
					break
				}
			}
		}
		if !improved {
			break
		}
	}
	p.VEdge1, p.VEdge2 = best1, best2
	return delay
}
