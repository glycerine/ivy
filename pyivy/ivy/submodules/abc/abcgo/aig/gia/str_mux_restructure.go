package gia

import (
	"fmt"
)

func Str_MuxChangeOnce(tree []Str_Mux_t, path []int32, i, k int, backup []Str_Mux_t, dst *Gia_Man_t, delays *Vec_Int_t) {
	if i+1 >= k || k >= len(path) {
		panic("gia: invalid mux change path")
	}
	ids := [3]int32{path[k] >> 1, path[i+1] >> 1, path[i] >> 1}
	inds := [3]int32{path[k] & 1, path[i+1] & 1, path[i] & 1}
	if backup != nil {
		if len(backup) < 3 {
			panic("gia: short mux backup")
		}
		for n, id := range ids {
			backup[n] = tree[id]
		}
	}
	top, middle, bottom := &tree[ids[0]], &tree[ids[1]], &tree[ids[2]]
	if top.Edge[inds[0]].Fan <= 0 || middle.Edge[inds[1]].Fan <= 0 {
		panic("gia: invalid mux change")
	}
	var phase int32
	for c := i + 1; c < k; c++ {
		phase ^= tree[path[c]>>1].Edge[path[c]&1].fCompl
	}
	other := inds[2] ^ 1
	old := bottom.Edge[other]
	bottom.Edge[other].Fan = top.Edge[inds[0]].Fan
	bottom.Edge[other].fCompl = 0
	top.Edge[inds[0]].Fan = bottom.Id
	middle.Edge[inds[1]].Fan = old.Fan
	middle.Edge[inds[1]].fCompl ^= old.fCompl
	middle.Edge[inds[1]].FanDel = old.FanDel
	middle.Edge[inds[1]].Copy = old.Copy
	for c := i + 1; c < k; c++ {
		bottom.Edge[2].FanDel = Str_Delay2(bottom.Edge[2].FanDel, tree[path[c]>>1].Edge[2].FanDel, tree[0].nLutSize)
	}
	if dst == nil {
		return
	}
	result := int32(1)
	for c := i; c < k; c++ {
		node := &tree[path[c]>>1]
		if node.Edge[2].fCompl != 0 {
			panic("gia: complemented mux control")
		}
		condition := node.Edge[2].Copy
		if path[c]&1 == 0 {
			condition ^= 1
		}
		result = Gia_ManHashAnd(dst, result, condition)
		Str_ObjDelay(dst, result>>1, tree[0].nLutSize, delays)
	}
	if path[i]&1 == 0 {
		result ^= 1
	}
	bottom.Edge[2].Copy = result
	bottom.Edge[inds[2]].fCompl ^= phase
}
func Str_MuxChangeUndo(tree []Str_Mux_t, path []int32, i, k int, backup []Str_Mux_t) {
	tree[path[k]>>1] = backup[0]
	tree[path[i+1]>>1] = backup[1]
	tree[path[i]>>1] = backup[2]
}
func Str_MuxTryOnce(dst *Gia_Man_t, ntk *Str_Ntk_t, tree []Str_Mux_t, root *Str_Mux_t, edge int32, delays *Vec_Int_t, verbose int32) int32 {
	best := Str_MuxDelayEdge_rec(root, edge)
	initial := best
	path := make([]int32, 0, 32)
	if Str_MuxFindPathEdge_rec(root, edge, &path) == 0 {
		return 0
	}
	bestI, bestK, checks := -1, -1, 0
	var backup [3]Str_Mux_t
	for i := 0; i < len(path); i++ {
		for k := i + 2; k < len(path); k++ {
			Str_MuxChangeOnce(tree, path, i, k, backup[:], nil, nil)
			delay := Str_MuxDelayEdge_rec(root, edge)
			Str_MuxChangeUndo(tree, path, i, k, backup[:])
			if best > delay || bestI >= 0 && best == delay {
				best, bestI, bestK = delay, i, k
			}
			checks++
		}
	}
	if bestI < 0 {
		return 0
	}
	Str_MuxChangeOnce(tree, path, bestI, bestK, nil, dst, delays)
	if verbose != 0 {
		fmt.Printf("Node %6d (%3d %3d) : Checks = %d. Delay: %d -> %d.\n", root.Id, bestI, bestK, checks, initial, best)
	}
	return 1
}
func Str_MuxRestruct_rec(dst *Gia_Man_t, ntk *Str_Ntk_t, tree []Str_Mux_t, root *Str_Mux_t, edge int32, delays *Vec_Int_t, verbose int32) int32 {
	var changes int32
	if branch := Str_MuxFindBranching(root, edge); branch != nil {
		changes |= Str_MuxRestruct_rec(dst, ntk, tree, branch, 0, delays, verbose)
		changes |= Str_MuxRestruct_rec(dst, ntk, tree, branch, 1, delays, verbose)
	}
	return changes | Str_MuxTryOnce(dst, ntk, tree, root, edge, delays, verbose)
}
func strMuxRestructure(dst *Gia_Man_t, ntk *Str_Ntk_t, start, count int32, delays *Vec_Int_t, lutSize, verbose int32, recursive bool) int32 {
	if count >= MAX_TREE || count < 2 {
		return -1
	}
	tree := make([]Str_Mux_t, count+1)
	Str_MuxCreate(tree, ntk, start, count, delays, lutSize)
	initial := Str_MuxDelayEdge_rec(&tree[0], 0)
	changed := false
	for {
		var next int32
		if recursive {
			next = Str_MuxRestruct_rec(dst, ntk, tree, &tree[0], 0, delays, verbose)
		} else {
			next = Str_MuxTryOnce(dst, ntk, tree, &tree[0], 0, delays, verbose)
		}
		if next == 0 {
			break
		}
		changed = true
	}
	if !changed {
		return -1
	}
	ntk.DelayGain += initial - Str_MuxDelayEdge_rec(&tree[0], 0)
	return Str_MuxToGia_rec(dst, &tree[0], 0, delays)
}
func Str_MuxRestructure1(dst *Gia_Man_t, ntk *Str_Ntk_t, start, count int32, delays *Vec_Int_t, lutSize, verbose int32) int32 {
	return strMuxRestructure(dst, ntk, start, count, delays, lutSize, verbose, false)
}
func Str_MuxRestructure2(dst *Gia_Man_t, ntk *Str_Ntk_t, start, count int32, delays *Vec_Int_t, lutSize, verbose int32) int32 {
	return strMuxRestructure(dst, ntk, start, count, delays, lutSize, verbose, true)
}
