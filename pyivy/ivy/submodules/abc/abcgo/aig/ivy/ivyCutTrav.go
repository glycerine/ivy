// Local-volume bit-cut traversal ported from ABC ivyCutTrav.c.
package ivy

import (
	"fmt"
	"math/bits"
	"slices"
	"sort"
	"time"
)

// Bit cuts are slices into Go-owned arena blocks. Arena growth cannot invalidate
// previously returned cuts, unlike pointers into a reallocating C vector.
func Ivy_NodeCutElementary(arena *[]uint32, words, id int32) []uint32 {
	if words <= 0 || id < 0 || int64(id) >= 32*int64(words) {
		panic("ivy: invalid elementary bit cut")
	}
	start := len(*arena)
	*arena = append(*arena, make([]uint32, words)...)
	cut := (*arena)[start : start+int(words)]
	cut[id/32] |= 1 << uint32(id%32)
	return cut
}
func Ivy_CompareNodesByLevel(a, b *Ivy_Obj_t) int {
	if a.Level < b.Level {
		return -1
	}
	if a.Level > b.Level {
		return 1
	}
	return 0
}
func ivyCutTraversalCheck(root *Ivy_Obj_t) {
	if root == nil || Ivy_IsComplement(root) != 0 || root.owner == nil || root.recycled || Ivy_ManIsAcyclic_rec(root.owner, root) == 0 {
		panic("ivy: invalid or cyclic cut volume")
	}
}
func ivyCutVolumeMark(root *Ivy_Obj_t, depth int32) {
	if depth <= 0 || Ivy_ObjIsCi(root) != 0 || Ivy_ObjIsConst1(root) != 0 {
		return
	}
	if root.pFanin0 != nil {
		ivyCutVolumeMark(Ivy_ObjFanin0(root), depth-1)
	}
	if root.pFanin1 != nil {
		ivyCutVolumeMark(Ivy_ObjFanin1(root), depth-1)
	}
	root.fMarkA = 1
}
func Ivy_NodeComputeVolumeTrav1_rec(root *Ivy_Obj_t, depth int32) {
	ivyCutTraversalCheck(root)
	ivyCutVolumeMark(root, depth)
}
func ivyCutVolumeCollect(root *Ivy_Obj_t, nodes *[]*Ivy_Obj_t, seen map[*Ivy_Obj_t]bool) {
	if root == nil || root.fMarkA == 0 || seen[root] {
		return
	}
	seen[root] = true
	if root.pFanin0 != nil {
		ivyCutVolumeCollect(Ivy_ObjFanin0(root), nodes, seen)
	}
	if root.pFanin1 != nil {
		ivyCutVolumeCollect(Ivy_ObjFanin1(root), nodes, seen)
	}
	*nodes = append(*nodes, root)
}
func Ivy_NodeComputeVolumeTrav2_rec(root *Ivy_Obj_t, nodes *[]*Ivy_Obj_t) {
	ivyCutTraversalCheck(root)
	seen := map[*Ivy_Obj_t]bool{}
	for _, n := range *nodes {
		seen[n] = true
	}
	ivyCutVolumeCollect(root, nodes, seen)
}

// Ivy_NodeComputeVolume preserves the original six-level neighborhood policy.
// The historical nodeLimit argument is advisory; Volume2 is the size-driven variant.
func Ivy_NodeComputeVolume(root *Ivy_Obj_t, nodeLimit int32, nodes, front *[]*Ivy_Obj_t) {
	ivyCutTraversalCheck(root)
	if Ivy_ObjIsNode(root) == 0 && Ivy_ObjIsBuf(root) == 0 {
		panic("ivy: cut volume requires a gate")
	}
	ivyCutVolumeMark(root, 6)
	var inside []*Ivy_Obj_t
	ivyCutVolumeCollect(root, &inside, map[*Ivy_Obj_t]bool{})
	*nodes = (*nodes)[:0]
	for _, n := range inside {
		for _, edge := range []*Ivy_Obj_t{n.pFanin0, n.pFanin1} {
			if edge != nil {
				fanin := Ivy_Regular(edge)
				if fanin.fMarkA == 0 {
					fanin.fMarkA = 1
					*nodes = append(*nodes, fanin)
				}
			}
		}
	}
	*front = append((*front)[:0], (*nodes)...)
	*nodes = append(*nodes, inside...)
	for i, n := range *nodes {
		n.fMarkA = 0
		n.TravId = int32(i)
	}
}
func Ivy_NodeComputeVolume2(root *Ivy_Obj_t, nodeLimit int32, nodes, front *[]*Ivy_Obj_t) {
	ivyCutTraversalCheck(root)
	if Ivy_ObjIsNode(root) == 0 || nodeLimit < 1 {
		panic("ivy: invalid size-driven cut volume")
	}
	*nodes = append((*nodes)[:0], root)
	*front = append((*front)[:0], root)
	seen := map[*Ivy_Obj_t]bool{root: true}
	// Expand highest-level frontier gates, preserving reverse-frontier tie order.
	for {
		pivot := -1
		for i := len(*front) - 1; i >= 0; i-- {
			n := (*front)[i]
			if Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
				continue
			}
			if pivot < 0 || n.Level > (*front)[pivot].Level {
				pivot = i
			}
		}
		if pivot < 0 {
			break
		}
		n := (*front)[pivot]
		*front = slices.Delete(*front, pivot, pivot+1)
		for _, edge := range []*Ivy_Obj_t{n.pFanin0, n.pFanin1} {
			if edge == nil {
				continue
			}
			fanin := Ivy_Regular(edge)
			if !seen[fanin] {
				seen[fanin] = true
				*nodes = append(*nodes, fanin)
				*front = append(*front, fanin)
			}
		}
		if len(*nodes) >= int(nodeLimit) {
			break
		}
	}
	sort.SliceStable(*nodes, func(i, j int) bool { return Ivy_CompareNodesByLevel((*nodes)[i], (*nodes)[j]) < 0 })
	for i, n := range *nodes {
		n.TravId = int32(i)
	}
}
func Extra_TruthOrWords(out, a, b []uint32, words int32) {
	for i := int32(0); i < words; i++ {
		out[i] = a[i] | b[i]
	}
}
func Extra_TruthIsImplyWords(a, b []uint32, words int32) int32 {
	for i := int32(0); i < words; i++ {
		if a[i]&^b[i] != 0 {
			return 0
		}
	}
	return 1
}
func Ivy_NodeFindCutsMerge(a, b [][]uint32, cuts *[][]uint32, leaves, words int32, arena *[]uint32) {
	if leaves < 1 || leaves > IVY_CUT_INPUT || words < 1 {
		panic("ivy: invalid bit-cut merge bounds")
	}
	for _, left := range a {
		for _, right := range b {
			count := 0
			for w := int32(0); w < words; w++ {
				count += bits.OnesCount32(left[w] | right[w])
			}
			if count > int(leaves) {
				continue
			}
			candidate := make([]uint32, words)
			Extra_TruthOrWords(candidate, left, right, words)
			rejected := false
			for _, prior := range *cuts {
				if Extra_TruthIsImplyWords(prior, candidate, words) != 0 {
					rejected = true
					break
				}
			}
			if rejected {
				continue
			}
			keep := (*cuts)[:0]
			for _, prior := range *cuts {
				if Extra_TruthIsImplyWords(candidate, prior, words) == 0 {
					keep = append(keep, prior)
				}
			}
			start := len(*arena)
			*arena = append(*arena, candidate...)
			keep = append(keep, (*arena)[start:start+int(words)])
			*cuts = keep
		}
	}
}
func Ivy_NodeFindCutsTravAll(p *Ivy_Man_t, root *Ivy_Obj_t, leaves, nodeLimit int32, nodes, front *[]*Ivy_Obj_t, arena *[]uint32, bitCuts *[][][]uint32) *Ivy_Store_t {
	if root == nil || root.owner != p || leaves < 1 || leaves > IVY_CUT_INPUT {
		panic("ivy: invalid traversal cut request")
	}
	Ivy_NodeComputeVolume(root, nodeLimit-1, nodes, front)
	// Traversal indexes are scratch state; clear even if an invalid cut fails.
	defer func() {
		for _, n := range *nodes {
			n.TravId = 0
			n.fMarkA = 0
		}
	}()
	*arena = (*arena)[:0]
	*bitCuts = make([][][]uint32, len(*nodes))
	words := int32((len(*nodes) + 31) / 32)
	for _, n := range *front {
		(*bitCuts)[n.TravId] = append((*bitCuts)[n.TravId], Ivy_NodeCutElementary(arena, words, n.TravId))
	}
	for _, n := range *nodes {
		row := &(*bitCuts)[n.TravId]
		if len(*row) != 0 {
			continue
		}
		*row = append(*row, Ivy_NodeCutElementary(arena, words, n.TravId))
		a := (*bitCuts)[Ivy_ObjFanin0(n).TravId]
		if Ivy_ObjIsBuf(n) != 0 {
			for _, cut := range a {
				*row = append(*row, cut)
			}
			continue
		}
		b := (*bitCuts)[Ivy_ObjFanin1(n).TravId]
		Ivy_NodeFindCutsMerge(a, b, row, leaves, words, arena)
	}
	store := &Ivy_Store_t{nCutsMax: IVY_CUT_LIMIT}
	for _, bitCut := range (*bitCuts)[root.TravId] {
		cut := Ivy_Cut_t{nSizeMax: int16(leaves)}
		for index, n := range *nodes {
			if bitCut[index/32]&(1<<uint(index%32)) != 0 {
				if int(cut.nSize) >= int(leaves) {
					panic("ivy: oversized traversal cut")
				}
				cut.pArray[cut.nSize] = n.Id
				cut.nSize++
			}
		}
		slices.Sort(cut.pArray[:cut.nSize])
		Ivy_NodeCutHash(&cut)
		store.pCuts[store.nCuts] = cut
		store.nCuts++
		if store.nCuts == store.nCutsMax {
			break
		}
	}
	return store
}
func Ivy_ManTestCutsTravAll(p *Ivy_Man_t) {
	var nodes, front []*Ivy_Obj_t
	var arena []uint32
	var cuts [][][]uint32
	start := time.Now()
	total, count, saturated := int32(0), int32(0), int32(0)
	for _, n := range p.vObjs {
		if n == nil || Ivy_ObjIsNode(n) == 0 {
			continue
		}
		store := Ivy_NodeFindCutsTravAll(p, n, 4, 60, &nodes, &front, &arena, &cuts)
		total += store.nCuts - 1
		count++
		saturated += ivyBool(store.nCuts == IVY_CUT_LIMIT)
	}
	fmt.Printf("Total cuts = %6d. Trivial = %6d.   Nodes = %6d. Satur = %6d. Time = %s\n", total, Ivy_ManPiNum(p)+Ivy_ManNodeNum(p), count, saturated, time.Since(start))
}
