package gia

import "github.com/glycerine/abcgo/misc/vec"

func Spl_ManFindOne(p *Spl_Man_t) int32 {
	g := p.pGia
	nodes := vec.Vec_IntArray(p.vNodes)
	for _, id := range nodes {
		for _, fan := range vec.Vec_IntArray(Gia_ObjLutFanins2(g, id)) {
			Gia_ObjLutRefDecId(g, fan)
		}
	}
	defer func() {
		for _, id := range nodes {
			for _, fan := range vec.Vec_IntArray(Gia_ObjLutFanins2(g, id)) {
				Gia_ObjLutRefIncId(g, fan)
			}
		}
	}()
	external := func(appendCandidates bool) int32 {
		for _, id := range nodes {
			refs := Gia_ObjLutRefNumId(g, id)
			if refs <= 0 || refs >= 5 {
				continue
			}
			n := Spl_ManLutFanouts(g, id, p.vFanouts, p.vMarksNo, p.vMarksCIO)
			if refs == 1 && n == 1 {
				return vec.Vec_IntEntry(p.vFanouts, 0)
			}
			if appendCandidates {
				for _, fan := range vec.Vec_IntArray(p.vFanouts) {
					vec.Vec_IntPush(p.vCands, fan)
				}
			}
		}
		return 0
	}
	if p.fReverse != 0 && len(nodes)%2 != 0 {
		if id := external(false); id != 0 {
			return id
		}
	}
	collect := func(unreferenced bool) {
		vec.Vec_IntClear(p.vCands)
		vec.Vec_IntClear(p.vInputs)
		for _, id := range nodes {
			for _, fan := range vec.Vec_IntArray(Gia_ObjLutFanins2(g, id)) {
				if vec.Vec_BitEntry(p.vMarksNo, int(fan)) != 0 || vec.Vec_BitEntry(p.vMarksCIO, int(fan)) != 0 || (unreferenced && Gia_ObjLutRefNumId(g, fan) != 0) {
					continue
				}
				vec.Vec_IntPush(p.vCands, fan)
				vec.Vec_IntPush(p.vInputs, fan)
			}
		}
	}
	collect(true)
	if id := Spl_ManFindGoodCand(p); id != 0 {
		return id
	}
	collect(false)
	if id := external(true); id != 0 {
		return id
	}
	if id := Spl_ManFindGoodCand(p); id != 0 {
		return id
	}
	if vec.Vec_IntSize(p.vCands) > 0 {
		return vec.Vec_IntEntry(p.vCands, 0)
	}
	return 0
}
func Spl_ManLutMffcSize(g *Gia_Man_t, id int32, temp *Vec_Int_t, marked *Vec_Bit_t) int32 {
	if Gia_ObjIsLut2(g, id) == 0 {
		panic("gia: expected mapped LUT")
	}
	Gia_ManIncrementTravId(g)
	Gia_ManCollectAnds(g, []int32{id}, 1, temp, Gia_ObjLutFanins2(g, id))
	data := vec.Vec_IntArray(temp)
	n := 0
	for _, x := range data {
		if vec.Vec_BitEntry(marked, int(x)) == 0 {
			data[n] = x
			n++
		}
	}
	vec.Vec_IntShrink(temp, n)
	return int32(n)
}
func Spl_ManAddNode(p *Spl_Man_t, id int32, candidates *Vec_Int_t) {
	vec.Vec_IntPush(p.vNodes, id)
	vec.Vec_BitWriteEntry(p.vMarksNo, int(id), 1)
	for _, x := range vec.Vec_IntArray(candidates) {
		vec.Vec_IntPush(p.vAnds, x)
		vec.Vec_BitWriteEntry(p.vMarksAnd, int(x), 1)
	}
}
func Spl_ManComputeOne(p *Spl_Man_t, pivot int32) int32 {
	if Gia_ObjIsLut2(p.pGia, pivot) == 0 {
		panic("gia: window pivot is not mapped")
	}
	for _, id := range vec.Vec_IntArray(p.vNodes) {
		vec.Vec_BitWriteEntry(p.vMarksNo, int(id), 0)
	}
	for _, id := range vec.Vec_IntArray(p.vAnds) {
		vec.Vec_BitWriteEntry(p.vMarksAnd, int(id), 0)
	}
	vec.Vec_IntClear(p.vNodes)
	vec.Vec_IntClear(p.vAnds)
	Spl_ManLutMffcSize(p.pGia, pivot, p.vCands, p.vMarksAnd)
	Spl_ManAddNode(p, pivot, p.vCands)
	if int32(vec.Vec_IntSize(p.vAnds)) > p.Limit {
		return 0
	}
	for {
		id := Spl_ManFindOne(p)
		if id == 0 {
			break
		}
		n := Spl_ManLutMffcSize(p.pGia, id, p.vCands, p.vMarksAnd)
		if int32(vec.Vec_IntSize(p.vAnds))+n > p.Limit {
			break
		}
		Spl_ManAddNode(p, id, p.vCands)
	}
	vec.Vec_IntSort(p.vNodes, 0)
	vec.Vec_IntSort(p.vAnds, 0)
	Spl_ManWinFindLeavesRoots(p)
	vec.Vec_IntSort(p.vLeaves, 0)
	vec.Vec_IntSort(p.vRoots, 0)
	return 1
}
func Gia_ManComputeOneWin(g *Gia_Man_t, pivot int32, roots, nodes, leaves, ands **Vec_Int_t) int32 {
	p, ok := g.pSatlutWinman.(*Spl_Man_t)
	if !ok || p == nil {
		panic("gia: missing window manager")
	}
	if pivot == -1 {
		Spl_ManStop(p)
		g.pSatlutWinman = nil
		return 0
	}
	if Spl_ManComputeOne(p, pivot) == 0 {
		*roots, *nodes, *leaves, *ands = nil, nil, nil, nil
		return 0
	}
	*roots, *nodes, *leaves, *ands = p.vRoots, p.vNodes, p.vLeaves, p.vAnds
	return int32(vec.Vec_IntSize(p.vAnds))
}
func Gia_ManComputeOneWinStart(g *Gia_Man_t, limit, reverse int32) {
	if g.pSatlutWinman != nil {
		panic("gia: window manager already exists")
	}
	g.pSatlutWinman = Spl_ManAlloc(g, limit, reverse)
}
