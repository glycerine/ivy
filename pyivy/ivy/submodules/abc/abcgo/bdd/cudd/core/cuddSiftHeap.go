package core

import "sort"

// siftHeap is a private, mutable copy of the BDD heap on which variable
// sifting runs exactly as CUDD's does on its unique subtables
// (cuddReorder.c): cuddSwapInPlace touches only the two levels it swaps,
// keeps dead nodes of lower levels until the end (they still count as
// keys), and the objective is keys minus isolated projection functions.
// Only the final order is committed to the live manager.
type siftHeap struct {
	n        int
	nodes    []siftNode // nodes[0] is the constant one
	perm     []int      // level of each variable index
	invperm  []int      // variable index at each level
	levels   []siftLevel
	keys     int
	isolated int
	varNode  []int32 // projection node of each variable index
	interact []bool  // n*n, symmetric
	swaps    int     // ddTotalNumberSwapping
	roots    []siftEdge
	next     []int // symmetry groups by level (symmetric sifting)
	consts   int   // constants.keys: C's keys (and objective) include them
	free     []int32
	xnodes   []int32 // scratch: the x level's node list during a swap
	xmap     siftTab
}

// siftEdge is a node id shifted left once, with the complement in bit 0.
type siftEdge = int32

type siftNode struct {
	index int32 // variable index (-1 for the constant)
	t, e  siftEdge
	ref   int32
	alive bool
}

// siftLevel lists the nodes of one level. Lookups by (then, else) are
// needed only for the level being rebuilt in a swap (siftHeap.xmap).
type siftLevel struct {
	nodes []int32
}

func siftReg(e siftEdge) int32         { return e >> 1 }
func siftCompl(e siftEdge) bool        { return e&1 != 0 }
func siftNot(e siftEdge) siftEdge      { return e ^ 1 }
func siftMk(id int32, c bool) siftEdge { return id<<1 | siftEdge(boolToInt(c)) }

func (h *siftHeap) varRefIsolated(index int) bool {
	return h.nodes[h.varNode[index]].ref == 1
}

func (h *siftHeap) testInteract(a, b int) bool { return h.interact[a*h.n+b] }

func (h *siftHeap) inc(e siftEdge) { h.nodes[siftReg(e)].ref++ }
func (h *siftHeap) dec(e siftEdge) {
	if r := &h.nodes[siftReg(e)].ref; *r > 0 {
		*r--
	}
}

func (h *siftHeap) indexOf(e siftEdge) int32 { return h.nodes[siftReg(e)].index }

func (h *siftHeap) cost() int { return h.keys + h.consts - h.isolated }

// newSiftNode adds a node with one reference (cuddDynamicAllocNode).
func (h *siftHeap) newSiftNode(index int32, t, e siftEdge) int32 {
	if k := len(h.free); k > 0 {
		id := h.free[k-1]
		h.free = h.free[:k-1]
		h.nodes[id] = siftNode{index: index, t: t, e: e, ref: 1, alive: true}
		return id
	}
	h.nodes = append(h.nodes, siftNode{index: index, t: t, e: e, ref: 1, alive: true})
	return int32(len(h.nodes) - 1)
}

// swap is cuddSwapInPlace(x, x+1); it returns keys - isolated.
func (h *siftHeap) swap(x int) int {
	y := x + 1
	h.swaps++
	xindex, yindex := h.invperm[x], h.invperm[y]
	xl, yl := &h.levels[x], &h.levels[y]
	oldx, oldy := len(xl.nodes), len(yl.nodes)
	newxkeys, newykeys := oldx, oldy
	if h.testInteract(xindex, yindex) {
		// rebuild level x in its own storage: the old node list is copied
		// aside and the map cleared
		h.xnodes = append(h.xnodes[:0], xl.nodes...)
		h.xmap.reset(2 * len(h.xnodes))
		newX := siftLevel{nodes: xl.nodes[:0]}
		xmap := &h.xmap
		newxkeys = 0
		isolated := -(boolToInt(h.varRefIsolated(xindex)) + boolToInt(h.varRefIsolated(yindex)))
		var g []int32
		for _, f := range h.xnodes {
			nf := &h.nodes[f]
			if int(h.indexOf(nf.t)) != yindex && int(h.indexOf(nf.e)) != yindex {
				newxkeys++
				newX.nodes = append(newX.nodes, f)
				xmap.put(siftKey(nf.t, nf.e), f)
			} else {
				nf.index = int32(yindex)
				g = append(g, f)
			}
		}
		// C keeps g as a LIFO list
		for k := len(g) - 1; k >= 0; k-- {
			f := g[k]
			f1 := h.nodes[f].t
			f11, f10 := f1, f1
			if int(h.indexOf(f1)) == yindex {
				f11, f10 = h.nodes[siftReg(f1)].t, h.nodes[siftReg(f1)].e
			}
			f0 := h.nodes[f].e
			comple := siftCompl(f0)
			f0r := siftMk(siftReg(f0), false)
			f01, f00 := f0r, f0r
			if int(h.indexOf(f0r)) == yindex {
				f01, f00 = h.nodes[siftReg(f0r)].t, h.nodes[siftReg(f0r)].e
			}
			if comple {
				f01, f00 = siftNot(f01), siftNot(f00)
			}
			h.dec(f1)
			var newf1 siftEdge
			if f11 == f01 {
				newf1 = f11
				h.inc(newf1)
			} else if id, ok := xmap.get(siftKey(f11, f01)); ok {
				h.nodes[id].ref++
				newf1 = siftMk(id, false)
			} else {
				id := h.newSiftNode(int32(xindex), f11, f01)
				newxkeys++
				newX.nodes = append(newX.nodes, id)
				xmap.put(siftKey(f11, f01), id)
				h.inc(f11)
				h.inc(f01)
				newf1 = siftMk(id, false)
			}
			h.nodes[f].t = newf1
			h.dec(f0)
			var newf0 siftEdge
			if f10 == f00 {
				newf0 = f00
				h.inc(newf0)
			} else {
				newcomplement := siftCompl(f10)
				if newcomplement {
					f10, f00 = siftNot(f10), siftNot(f00)
				}
				if id, ok := xmap.get(siftKey(f10, f00)); ok {
					h.nodes[id].ref++
					newf0 = siftMk(id, false)
				} else {
					id := h.newSiftNode(int32(xindex), f10, f00)
					newxkeys++
					newX.nodes = append(newX.nodes, id)
					xmap.put(siftKey(f10, f00), id)
					h.inc(f10)
					h.inc(f00)
					newf0 = siftMk(id, false)
				}
				if newcomplement {
					newf0 = siftNot(newf0)
				}
			}
			h.nodes[f].e = newf0
			newykeys++
			yl.nodes = append(yl.nodes, f)
		}
		// free the dead nodes of the y list (their children stay)
		kept := yl.nodes[:0]
		for _, f := range yl.nodes {
			nf := &h.nodes[f]
			if nf.ref == 0 {
				h.dec(nf.t)
				h.dec(nf.e)
				nf.alive = false
				h.free = append(h.free, f)
				newykeys--
				continue
			}
			kept = append(kept, f)
		}
		yl.nodes = kept
		isolated += boolToInt(h.varRefIsolated(xindex)) + boolToInt(h.varRefIsolated(yindex))
		h.isolated += isolated
		*xl = newX
	}
	h.levels[x], h.levels[y] = h.levels[y], h.levels[x]
	h.perm[xindex], h.perm[yindex] = y, x
	h.invperm[x], h.invperm[y] = yindex, xindex
	h.keys += newxkeys + newykeys - oldx - oldy
	return h.cost()
}

type siftMove struct{ x, y, size int }

func (h *siftHeap) levelKeys(l int) int { return len(h.levels[l].nodes) }

// siftUp is ddSiftingUp; moves are returned newest first.
func (h *siftHeap) siftUp(y, xLow int, dd *DdManager) []siftMove {
	var moves []siftMove
	yindex := h.invperm[y]
	limitSize := h.cost()
	L := limitSize
	for x := xLow + 1; x < y; x++ {
		xindex := h.invperm[x]
		if h.testInteract(xindex, yindex) {
			L -= h.levelKeys(x) - boolToInt(h.varRefIsolated(xindex))
		}
	}
	L -= h.levelKeys(y) - boolToInt(h.varRefIsolated(yindex))
	x := y - 1
	for x >= xLow && L <= limitSize {
		xindex := h.invperm[x]
		size := h.swap(x)
		if h.testInteract(xindex, yindex) {
			L += h.levelKeys(y) - boolToInt(h.varRefIsolated(xindex))
		}
		moves = append([]siftMove{{x, y, size}}, moves...)
		if float64(size) > float64(limitSize)*dd.maxGrowth {
			break
		}
		if size < limitSize {
			limitSize = size
		}
		y = x
		x = y - 1
	}
	return moves
}

// siftDown is ddSiftingDown; moves are returned newest first.
func (h *siftHeap) siftDown(x, xHigh int, dd *DdManager) []siftMove {
	var moves []siftMove
	xindex := h.invperm[x]
	size := h.cost()
	limitSize := size
	R := 0
	for y := xHigh; y > x; y-- {
		yindex := h.invperm[y]
		if h.testInteract(xindex, yindex) {
			R += h.levelKeys(y) - boolToInt(h.varRefIsolated(yindex))
		}
	}
	y := x + 1
	for y <= xHigh && size-R < limitSize {
		yindex := h.invperm[y]
		if h.testInteract(xindex, yindex) {
			R -= h.levelKeys(y) - boolToInt(h.varRefIsolated(yindex))
		}
		size = h.swap(x)
		moves = append([]siftMove{{x, y, size}}, moves...)
		if float64(size) > float64(limitSize)*dd.maxGrowth {
			break
		}
		if size < limitSize {
			limitSize = size
		}
		x = y
		y = x + 1
	}
	return moves
}

// siftBackward is ddSiftingBackward: back to the newest best position.
func (h *siftHeap) siftBackward(size int, moves []siftMove) {
	for _, m := range moves {
		if m.size < size {
			size = m.size
		}
	}
	for _, m := range moves {
		if m.size == size {
			return
		}
		h.swap(m.x)
	}
}

// siftAux is ddSiftingAux.
func (h *siftHeap) siftAux(x, xLow, xHigh int, dd *DdManager) {
	initial := h.cost()
	switch {
	case x == xLow:
		h.siftBackward(initial, h.siftDown(x, xHigh, dd))
	case x == xHigh:
		h.siftBackward(initial, h.siftUp(x, xLow, dd))
	case x-xLow > xHigh-x:
		down := h.siftDown(x, xHigh, dd)
		if len(down) > 0 {
			x = down[0].y
		}
		h.siftBackward(initial, h.siftUp(x, xLow, dd))
	default:
		up := h.siftUp(x, xLow, dd)
		if len(up) > 0 {
			x = up[0].x
		}
		h.siftBackward(initial, h.siftDown(x, xHigh, dd))
	}
}

// sifting is cuddSifting(lower, upper).
func (h *siftHeap) sifting(lower, upper int, dd *DdManager) {
	entry := make([]int, h.n)
	vars := make([]int, h.n)
	for i := 0; i < h.n; i++ {
		entry[i] = h.levelKeys(h.perm[i])
		vars[i] = i
	}
	// qsort with ddUniqueCompare (glibc's qsort is a stable merge sort)
	sort.SliceStable(vars, func(a, b int) bool { return entry[vars[a]] > entry[vars[b]] })
	for i := 0; i < min(dd.siftMaxVar, h.n); i++ {
		if h.swaps >= dd.siftMaxSwap {
			break
		}
		x := h.perm[vars[i]]
		if x < lower || x > upper {
			continue
		}
		h.siftAux(x, lower, upper, dd)
	}
}

// newSiftHeap copies a canonical reorder heap (reorderBuild's) into a sift
// heap; it returns nil when the heap holds anything but BDD nodes.
func newSiftHeap(h *reorderHeap) *siftHeap {
	scratch := h.dd
	n := len(scratch.Vars)
	s := &siftHeap{consts: h.liveConstants, n: n, perm: append([]int(nil), scratch.Perm[:n]...), invperm: append([]int(nil), scratch.InvPerm[:n]...)}
	s.levels = make([]siftLevel, n)
	ids := map[*DdNode]int32{scratch.One: 0}
	s.nodes = append(s.nodes, siftNode{index: -1, alive: true})
	order := make([]*DdNode, 0, len(h.nodes))
	for node := range h.nodes {
		if Cudd_IsConstant(node) {
			continue // One, or ADD constants that BDD nodes never reference
		}
		if node.family != 0 {
			return nil
		}
		order = append(order, node)
	}
	sort.Slice(order, func(a, b int) bool { return order[a].ID < order[b].ID })
	for _, node := range order {
		ids[node] = int32(len(s.nodes))
		s.nodes = append(s.nodes, siftNode{index: int32(node.Index), alive: true})
	}
	foreign := false
	edge := func(x *DdNode) siftEdge {
		id, ok := ids[Cudd_Regular(x)]
		if !ok {
			foreign = true
		}
		return siftMk(id, Cudd_IsComplement(x))
	}
	s.varNode = make([]int32, n)
	for i, v := range scratch.Vars {
		s.varNode[i] = ids[v]
	}
	for _, node := range order {
		id := ids[node]
		nd := &s.nodes[id]
		nd.t, nd.e = edge(node.T), edge(node.E)
		nd.ref = int32(h.incoming[node]) + int32(h.external[node])
		if scratch.Vars[node.Index] == node {
			nd.ref++
		}
		l := &s.levels[s.perm[node.Index]]
		l.nodes = append(l.nodes, id)
		s.keys++
	}
	for _, r := range h.active {
		s.roots = append(s.roots, edge(r))
	}
	if foreign {
		return nil
	}
	for i := 0; i < n; i++ {
		if s.varRefIsolated(i) {
			s.isolated++
		}
	}
	// interaction: variables in the support of a common node
	words := (n + 63) / 64
	supp := make([][]uint64, len(s.nodes))
	var support func(id int32) []uint64
	support = func(id int32) []uint64 {
		if id == 0 {
			return nil
		}
		if supp[id] != nil {
			return supp[id]
		}
		b := make([]uint64, words)
		nd := s.nodes[id]
		b[nd.index/64] |= 1 << (uint(nd.index) % 64)
		for _, c := range [2]siftEdge{nd.t, nd.e} {
			for w, x := range support(siftReg(c)) {
				b[w] |= x
			}
		}
		supp[id] = b
		return b
	}
	parents := make([]int32, len(s.nodes))
	for _, node := range order {
		nd := s.nodes[ids[node]]
		parents[siftReg(nd.t)]++
		parents[siftReg(nd.e)]++
	}
	s.interact = make([]bool, n*n)
	var vars []int
	for _, node := range order {
		id := ids[node]
		if parents[id] != 0 {
			continue // its support lies within a parent's
		}
		vars = vars[:0]
		for w, x := range support(id) {
			for x != 0 {
				bit := trailingZeros64(x)
				vars = append(vars, w*64+bit)
				x &= x - 1
			}
		}
		for _, a := range vars {
			for _, b := range vars {
				s.interact[a*n+b] = true
			}
		}
	}
	return s
}

func trailingZeros64(x uint64) int {
	n := 0
	for x&1 == 0 {
		x >>= 1
		n++
	}
	return n
}

// liveCount counts the nodes reachable from referenced roots: nodes with
// references that are not only from other heap nodes.
func (s *siftHeap) liveCount(roots []siftEdge) int {
	seen := map[int32]bool{}
	var visit func(id int32)
	visit = func(id int32) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		visit(siftReg(s.nodes[id].t))
		visit(siftReg(s.nodes[id].e))
	}
	for _, r := range roots {
		visit(siftReg(r))
	}
	return len(seen)
}

// siftTab maps a (then, else) pair to a node id for the level being
// rebuilt: open addressing, cleared in O(1) by bumping a generation.
type siftTab struct {
	keys []uint64
	vals []int32
	gens []uint32
	gen  uint32
	n    int
}

func siftKey(t, e siftEdge) uint64 { return uint64(uint32(t))<<32 | uint64(uint32(e)) }

func siftHash(k uint64) uint64 {
	k *= 0x9E3779B97F4A7C15
	return k ^ k>>31
}

// reset empties the table and makes room for n entries.
func (t *siftTab) reset(n int) {
	size := 16
	for size < 2*n {
		size <<= 1
	}
	if size > len(t.keys) {
		t.keys, t.vals, t.gens = make([]uint64, size), make([]int32, size), make([]uint32, size)
		t.gen = 0
	}
	t.gen++
	if t.gen == 0 { // wrapped: really clear
		clear(t.gens)
		t.gen = 1
	}
	t.n = 0
}

func (t *siftTab) get(k uint64) (int32, bool) {
	mask := uint64(len(t.keys) - 1)
	for i := siftHash(k) & mask; ; i = (i + 1) & mask {
		if t.gens[i] != t.gen {
			return 0, false
		}
		if t.keys[i] == k {
			return t.vals[i], true
		}
	}
}

func (t *siftTab) put(k uint64, v int32) {
	if 2*(t.n+1) > len(t.keys) {
		keys, vals, gens, gen := t.keys, t.vals, t.gens, t.gen
		t.keys, t.vals, t.gens = make([]uint64, 2*len(keys)), make([]int32, 2*len(keys)), make([]uint32, 2*len(keys))
		t.gen, t.n = 1, 0
		for i := range keys {
			if gens[i] == gen {
				t.put(keys[i], vals[i])
			}
		}
	}
	mask := uint64(len(t.keys) - 1)
	for i := siftHash(k) & mask; ; i = (i + 1) & mask {
		if t.gens[i] != t.gen {
			t.keys[i], t.vals[i], t.gens[i] = k, v, t.gen
			t.n++
			return
		}
		if t.keys[i] == k {
			t.vals[i] = v
			return
		}
	}
}
