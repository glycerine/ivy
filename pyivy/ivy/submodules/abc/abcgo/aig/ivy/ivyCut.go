// Reconvergence-driven and bounded cut enumeration ported from ABC ivyCut.c.
package ivy

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"time"
)

func Ivy_NodeCutHashValue(id int32) int32 { return int32(uint32(1) << uint32(id%31)) }
func Ivy_NodeGetLeafCostOne(p *Ivy_Man_t, leaf int32, inside []int32) int32 {
	if leaf <= 0 {
		panic("ivy: invalid sequential leaf")
	}
	n := Ivy_ManObj(p, Ivy_LeafId(leaf))
	latches := Ivy_LeafLat(leaf) + Ivy_ObjIsLatch(n)
	if Ivy_ObjIsPi(n) != 0 || Ivy_ObjIsConst1(n) != 0 || latches > 15 {
		return 999
	}
	cost := int32(0)
	for _, edge := range []*Ivy_Obj_t{n.pFanin0, n.pFanin1} {
		if edge != nil {
			next := Ivy_LeafCreate(Ivy_Regular(edge).Id, latches)
			if next != 0 && !slices.Contains(inside, next) {
				cost++
			}
		}
	}
	return cost
}
func Ivy_ManSeqFindCut_int(p *Ivy_Man_t, front, inside *[]int32, limit int32) int32 {
	best := int32(99)
	var candidates []int32
	for _, leaf := range *front {
		cost := Ivy_NodeGetLeafCostOne(p, leaf, *inside)
		if cost < best {
			best = cost
			candidates = []int32{leaf}
		} else if cost == best {
			candidates = append(candidates, leaf)
		}
		if best <= 1 {
			break
		}
	}
	if best == 99 || int32(len(*front))-1+best > limit {
		return 0
	}
	leaf := candidates[rand.IntN(len(candidates))]
	index := slices.Index(*front, leaf)
	*front = slices.Delete(*front, index, index+1)
	n := Ivy_ManObj(p, Ivy_LeafId(leaf))
	latches := Ivy_LeafLat(leaf) + Ivy_ObjIsLatch(n)
	for _, edge := range []*Ivy_Obj_t{n.pFanin0, n.pFanin1} {
		if edge != nil {
			next := Ivy_LeafCreate(Ivy_Regular(edge).Id, latches)
			if next != 0 && !slices.Contains(*inside, next) {
				*front = append(*front, next)
				*inside = append(*inside, next)
			}
		}
	}
	return 1
}
func Ivy_ManSeqFindCut(p *Ivy_Man_t, root *Ivy_Obj_t, front, inside *[]int32, size int32) {
	if root == nil || Ivy_IsComplement(root) != 0 || Ivy_ObjIsNode(root) == 0 || size < 2 {
		panic("ivy: invalid sequential cut request")
	}
	a, b := Ivy_LeafCreate(Ivy_ObjFaninId0(root), 0), Ivy_LeafCreate(Ivy_ObjFaninId1(root), 0)
	*front = append((*front)[:0], a, b)
	*inside = append((*inside)[:0], Ivy_LeafCreate(root.Id, 0), a, b)
	for Ivy_ManSeqFindCut_int(p, front, inside, size) != 0 {
	}
}
func ivyCutPushUnique(nodes *[]*Ivy_Obj_t, n *Ivy_Obj_t) {
	if !slices.Contains(*nodes, n) {
		*nodes = append(*nodes, n)
	}
}
func Ivy_ManFindBoolCut_rec(p *Ivy_Man_t, n *Ivy_Obj_t, leaves, volume *[]*Ivy_Obj_t, pivot *Ivy_Obj_t) int32 {
	n = Ivy_Regular(n)
	if n == pivot {
		ivyCutPushUnique(leaves, n)
		ivyCutPushUnique(volume, n)
		return 1
	}
	if n.fMarkA != 0 || Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
		return 0
	}
	a := Ivy_ManFindBoolCut_rec(p, Ivy_ObjFanin0(n), leaves, volume, pivot)
	if Ivy_ObjIsBuf(n) != 0 {
		if a != 0 {
			ivyCutPushUnique(volume, n)
		}
		return a
	}
	b := Ivy_ManFindBoolCut_rec(p, Ivy_ObjFanin1(n), leaves, volume, pivot)
	if a == 0 && b == 0 {
		return 0
	}
	if a == 0 {
		ivyCutPushUnique(leaves, Ivy_ObjFanin0(n))
		ivyCutPushUnique(volume, Ivy_ObjFanin0(n))
	}
	if b == 0 {
		ivyCutPushUnique(leaves, Ivy_ObjFanin1(n))
		ivyCutPushUnique(volume, Ivy_ObjFanin1(n))
	}
	ivyCutPushUnique(volume, n)
	return 1
}
func Ivy_ManFindBoolCutCost(n *Ivy_Obj_t) int32 {
	if n.fMarkA == 0 {
		panic("ivy: Boolean cut cost outside construction zone")
	}
	if Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
		return 999
	}
	cost := ivyBool(Ivy_ObjFanin0(n).fMarkA == 0)
	if Ivy_ObjIsNode(n) != 0 {
		cost += ivyBool(Ivy_ObjFanin1(n).fMarkA == 0)
	}
	return cost
}
func Ivy_ManFindBoolCut(p *Ivy_Man_t, root *Ivy_Obj_t, front, volume, leaves *[]*Ivy_Obj_t) int32 {
	if root == nil || Ivy_IsComplement(root) != 0 || Ivy_ObjIsNode(root) == 0 {
		panic("ivy: invalid Boolean cut root")
	}
	ivyDfsCheckMarks(p)
	// The C routine leaked marks when no pivot was found. Always restore them.
	defer func() {
		for _, n := range p.vObjs {
			if n != nil {
				n.fMarkA, n.fMarkB = 0, 0
			}
		}
	}()
	*front = (*front)[:0]
	*volume = (*volume)[:0]
	*leaves = (*leaves)[:0]
	a, b := Ivy_ObjFanin0(root), Ivy_ObjFanin1(root)
	var control *Ivy_Obj_t
	if Ivy_ObjIsMuxType(root) != 0 {
		control = Ivy_ObjRecognizeMux(root, &a, &b)
		a, b = Ivy_Regular(a), Ivy_Regular(b)
	}
	a.fMarkA = 1
	b.fMarkB = 1
	*front = append(*front, a, b)
	*volume = append(*volume, a, b)
	var pivot *Ivy_Obj_t
	limit := max(root.Level-10, 1)
	for level := root.Level - 1; level >= limit && pivot == nil; level-- {
		for pivot == nil {
			index := -1
			for i, n := range *front {
				if n.Level == level && Ivy_ObjIsCi(n) == 0 && Ivy_ObjIsConst1(n) == 0 {
					index = i
					break
				}
			}
			if index < 0 {
				break
			}
			n := (*front)[index]
			*front = slices.Delete(*front, index, index+1)
			for _, edge := range []*Ivy_Obj_t{n.pFanin0, n.pFanin1} {
				if edge == nil {
					continue
				}
				fanin := Ivy_Regular(edge)
				if fanin.fMarkA == 0 && fanin.fMarkB == 0 {
					*front = append(*front, fanin)
					*volume = append(*volume, fanin)
				}
				fanin.fMarkA |= n.fMarkA
				fanin.fMarkB |= n.fMarkB
			}
			for _, edge := range []*Ivy_Obj_t{n.pFanin0, n.pFanin1} {
				if edge != nil {
					fanin := Ivy_Regular(edge)
					if fanin.fMarkA != 0 && fanin.fMarkB != 0 {
						pivot = fanin
						break
					}
				}
			}
		}
	}
	if pivot == nil {
		return 0
	}
	if control != nil && control.fMarkA == 0 && control.fMarkB == 0 {
		*front = append(*front, control)
	}
	for _, n := range *volume {
		n.fMarkA, n.fMarkB = 0, 0
	}
	for _, n := range *front {
		n.fMarkA = 1
	}
	*volume = (*volume)[:0]
	if Ivy_ManFindBoolCut_rec(p, root, leaves, volume, pivot) == 0 {
		panic("ivy: Boolean pivot unreachable")
	}
	for _, n := range *front {
		n.fMarkA = 0
	}
	for _, n := range *volume {
		n.fMarkA = 1
	}
	for {
		index := -1
		for i, n := range *leaves {
			if Ivy_ManFindBoolCutCost(n) < 2 {
				index = i
				break
			}
		}
		if index < 0 {
			break
		}
		n := (*leaves)[index]
		*leaves = slices.Delete(*leaves, index, index+1)
		for _, edge := range []*Ivy_Obj_t{n.pFanin0, n.pFanin1} {
			if edge != nil {
				fanin := Ivy_Regular(edge)
				if fanin.fMarkA == 0 {
					fanin.fMarkA = 1
					*volume = append(*volume, fanin)
					*leaves = append(*leaves, fanin)
				}
			}
		}
	}
	return 1
}
func Ivy_ManTestCutsBool(p *Ivy_Man_t) {
	var front, volume, leaves []*Ivy_Obj_t
	for _, n := range p.vObjs {
		if n != nil && Ivy_ObjIsNode(n) != 0 {
			if Ivy_ObjIsMuxType(n) != 0 {
				fmt.Print("m")
				continue
			}
			if Ivy_ManFindBoolCut(p, n, &front, &volume, &leaves) == 0 {
				fmt.Print("- ")
			} else {
				fmt.Printf("%d ", len(leaves))
			}
		}
	}
	fmt.Println()
}

func Ivy_NodeCutHash(c *Ivy_Cut_t) uint32 {
	c.uHash = 0
	for _, id := range c.pArray[:c.nSize] {
		c.uHash |= uint32(Ivy_NodeCutHashValue(id))
	}
	return c.uHash
}
func Ivy_NodeCutShrink(c *Ivy_Cut_t, id int32) {
	i := slices.Index(c.pArray[:c.nSize], id)
	if i < 0 {
		panic("ivy: missing cut leaf")
	}
	copy(c.pArray[i:], c.pArray[i+1:c.nSize])
	c.nSize--
}
func Ivy_NodeCutExtend(c *Ivy_Cut_t, id int32) int32 {
	ids := c.pArray[:c.nSize]
	index, found := slices.BinarySearch(ids, id)
	if found {
		return 1
	}
	if c.nSize >= c.nSizeMax || int(c.nSize) >= len(c.pArray) {
		return 0
	}
	copy(c.pArray[index+1:], c.pArray[index:c.nSize])
	c.pArray[index] = id
	c.nSize++
	return 1
}
func Ivy_NodeCutPrescreen(c *Ivy_Cut_t, a, b int32) int32 {
	return ivyBool(c.nSize < c.nSizeMax || slices.Contains(c.pArray[:c.nSize], a) || slices.Contains(c.pArray[:c.nSize], b) || a == b)
}
func Ivy_NodeCutDeriveNew(c, next *Ivy_Cut_t, old, a, b int32) int32 {
	result := *c
	Ivy_NodeCutShrink(&result, old)
	if Ivy_NodeCutExtend(&result, a) == 0 || Ivy_NodeCutExtend(&result, b) == 0 {
		return 0
	}
	Ivy_NodeCutHash(&result)
	*next = result
	return 1
}
func Ivy_NodeCutFindOrAdd(store *Ivy_Store_t, next *Ivy_Cut_t) int32 {
	for i := int32(0); i < store.nCuts; i++ {
		c := &store.pCuts[i]
		if c.uHash == next.uHash && slices.Equal(c.pArray[:c.nSize], next.pArray[:next.nSize]) {
			return 1
		}
	}
	if store.nCuts >= store.nCutsMax || int(store.nCuts) >= len(store.pCuts) {
		panic("ivy: cut store full")
	}
	store.pCuts[store.nCuts] = *next
	store.nCuts++
	return 0
}
func Ivy_CutCheckDominance(dom, c *Ivy_Cut_t) int32 {
	for _, id := range dom.pArray[:dom.nSize] {
		if !slices.Contains(c.pArray[:c.nSize], id) {
			return 0
		}
	}
	return 1
}
func Ivy_NodeCutFindOrAddFilter(store *Ivy_Store_t, next *Ivy_Cut_t) int32 {
	for i := int32(0); i < store.nCuts; i++ {
		c := &store.pCuts[i]
		if c.nSize == 0 {
			continue
		}
		if c.nSize <= next.nSize && c.uHash&next.uHash == c.uHash && Ivy_CutCheckDominance(c, next) != 0 {
			return 1
		}
	}
	// Only remove dominated entries after ensuring no existing cut rejects this one.
	for i := int32(0); i < store.nCuts; i++ {
		c := &store.pCuts[i]
		if c.nSize > next.nSize && c.uHash&next.uHash == next.uHash && Ivy_CutCheckDominance(next, c) != 0 {
			c.nSize = 0
		}
	}
	return Ivy_NodeCutFindOrAdd(store, next)
}
func Ivy_NodeCompactCuts(store *Ivy_Store_t) {
	k := int32(0)
	for i := int32(0); i < store.nCuts; i++ {
		if store.pCuts[i].nSize != 0 {
			store.pCuts[k] = store.pCuts[i]
			k++
		}
	}
	clear(store.pCuts[k:store.nCuts])
	store.nCuts = k
}
func Ivy_NodePrintCut(c *Ivy_Cut_t) {
	fmt.Printf("%d : {", c.nSize)
	for _, id := range c.pArray[:c.nSize] {
		fmt.Printf(" %d", id)
	}
	fmt.Println(" }")
}
func Ivy_NodePrintCuts(store *Ivy_Store_t) {
	if store.nCuts == 0 {
		return
	}
	fmt.Printf("Node %d\n", store.pCuts[0].pArray[0])
	for i := int32(0); i < store.nCuts; i++ {
		Ivy_NodePrintCut(&store.pCuts[i])
	}
}
func Ivy_ObjRealFanin(n *Ivy_Obj_t) *Ivy_Obj_t { return Ivy_Regular(Ivy_ObjReal(n)) }
func Ivy_NodeFindCutsAll(p *Ivy_Man_t, root *Ivy_Obj_t, leaves int32) *Ivy_Store_t {
	if root == nil || Ivy_IsComplement(root) != 0 || leaves < 1 || leaves > IVY_CUT_INPUT {
		panic("ivy: invalid cut enumeration request")
	}
	// Per-call ownership replaces C's shared static store.
	store := &Ivy_Store_t{nCutsMax: IVY_CUT_LIMIT}
	initial := Ivy_Cut_t{nSize: 1, nSizeMax: int16(leaves)}
	initial.pArray[0] = root.Id
	Ivy_NodeCutHash(&initial)
	Ivy_NodeCutFindOrAdd(store, &initial)
	for i := int32(0); i < store.nCuts; i++ {
		cut := store.pCuts[i]
		if cut.nSize == 0 {
			continue
		}
		for _, id := range cut.pArray[:cut.nSize] {
			n := Ivy_ManObj(p, id)
			if Ivy_ObjIsCi(n) != 0 || Ivy_ObjIsConst1(n) != 0 {
				continue
			}
			a := Ivy_ObjRealFanin(Ivy_ObjFanin0(n)).Id
			b := a
			if Ivy_ObjIsNode(n) != 0 {
				b = Ivy_ObjRealFanin(Ivy_ObjFanin1(n)).Id
			}
			if Ivy_NodeCutPrescreen(&cut, a, b) == 0 {
				continue
			}
			var next Ivy_Cut_t
			if Ivy_NodeCutDeriveNew(&cut, &next, id, a, b) != 0 {
				Ivy_NodeCutFindOrAddFilter(store, &next)
			}
			if store.nCuts == IVY_CUT_LIMIT {
				break
			}
		}
		if store.nCuts == IVY_CUT_LIMIT {
			break
		}
	}
	Ivy_NodeCompactCuts(store)
	return store
}
func Ivy_ManTestCutsAll(p *Ivy_Man_t) {
	start := time.Now()
	total, nodes, saturated := int32(0), int32(0), int32(0)
	for _, n := range p.vObjs {
		if n != nil && Ivy_ObjIsNode(n) != 0 {
			count := Ivy_NodeFindCutsAll(p, n, 5).nCuts
			total += count - 1
			nodes++
			saturated += ivyBool(count == IVY_CUT_LIMIT)
		}
	}
	fmt.Printf("Total cuts = %6d. Trivial = %6d.   Nodes = %6d. Satur = %6d. Time = %s\n", total, Ivy_ManPiNum(p)+Ivy_ManNodeNum(p), nodes, saturated, time.Since(start))
}
