package core

// A line-by-line port of ABC's DAG-aware rewriting: Dar_ManRewrite
// (darCore.c), the 4-input cut computation (darCut.c) and library
// evaluation/construction (Dar_LibPrepare .. Dar_LibBuildBest in darLib.c).
// The data layout mirrors the C structures so that every decision (cut
// replacement by value, dominance filtering, subgraph priorities, gain and
// level tie-breaks) is taken in the same order as in ABC.

import (
	"math/bits"
	"sync"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
)

// darCut is Dar_Cut_t.
type darCut struct {
	sign    uint32
	truth   uint32 // 16 bits used
	value   int32
	used    bool
	nLeaves int32
	leaves  [4]int32
}

// darLibObj is Dar_LibObj_t.
type darLibObj struct {
	fan0, fan1     int32
	compl0, compl1 bool
	phase, term    bool
	num            int32
}

// darLibDat is Dar_LibDat_t.
type darLibDat struct {
	fn     aig.Edge // InvalidEdge is NULL
	level  int32
	travID int32
	prob   float32
	mffc   bool
}

// darLibBase is the read-only library after Dar_LibSetup.
type darLibBase struct {
	objs   []darLibObj
	subgr  [222][]int32 // outputs by class
	prios  [222][]int32
	phases []uint8
	perms  []uint8
	class  []uint8
	perms4 [][4]int
}

// darLibPrepared is the library state after Dar_LibPrepare(nSubgraphs).
type darLibPrepared struct {
	subgr0    [222][]int32
	nodes0    [222][]int32
	nodes0Max int32
	objNum    []int32 // Num of every object after preparation
}

var darNativeLib struct {
	once     sync.Once
	base     *darLibBase
	mu       sync.Mutex
	prepared map[int]*darLibPrepared
}

func darLibBaseGet() *darLibBase {
	darNativeLib.once.Do(func() { darNativeLib.base = darLibRead() })
	return darNativeLib.base
}

// darLibRead is Dar_LibRead followed by Dar_LibSetup.
func darLibRead() *darLibBase {
	nodes, outs, prios := libraryNodeData[:], libraryOutputData[:], libraryPriorityData[:]
	p := &darLibBase{objs: make([]darLibObj, 4, len(nodes)/2+4)}
	table := Dar_Truth4VarNPN()
	p.phases, p.perms, p.class = table.phase[:], table.permutation[:], table.class[:]
	p.perms4 = table.permutations
	for i, t := range [4]int32{0xAAAA, 0xCCCC, 0xF0F0, 0xFF00} {
		p.objs[i] = darLibObj{term: true, num: t}
	}
	for i := 0; i < len(nodes); i += 2 {
		id0, id1 := int32(nodes[i]>>1), int32(nodes[i+1]>>1)
		c0, c1 := nodes[i]&1 != 0, nodes[i+1]&1 != 0
		f0, f1 := p.objs[id0], p.objs[id1]
		t0, t1 := f0.num, f1.num
		if c0 {
			t0 = ^t0
		}
		if c1 {
			t1 = ^t1
		}
		p.objs = append(p.objs, darLibObj{fan0: id0, fan1: id1, compl0: c0, compl1: c1,
			phase: (c0 != f0.phase) && (c1 != f1.phase), num: 0xFFFF & t0 & t1})
	}
	objTruth := func(o darLibObj) int32 { return min(o.num, 0xFFFF&^o.num) }
	for _, out := range outs {
		class := p.class[objTruth(p.objs[out])]
		p.subgr[class] = append(p.subgr[class], int32(out))
	}
	counter := 0
	for i := range p.subgr {
		p.prios[i] = make([]int32, len(p.subgr[i]))
		for k := range p.prios[i] {
			p.prios[i][k] = int32(prios[counter])
			counter++
		}
	}
	// Dar_LibSetup ends with the elementary objects numbered 0..3; the node
	// lists by class it also builds are only used for training.
	for i := range p.objs {
		p.objs[i].num = 0xff
	}
	for i := 0; i < 4; i++ {
		p.objs[i].num = int32(i)
	}
	return p
}

// darLibPrepare is Dar_LibPrepare, cached per subgraph limit.
func darLibPrepare(nSubgraphs int) *darLibPrepared {
	darNativeLib.mu.Lock()
	defer darNativeLib.mu.Unlock()
	if p, ok := darNativeLib.prepared[nSubgraphs]; ok {
		return p
	}
	base := darLibBaseGet()
	r := &darLibPrepared{}
	for i := range base.subgr {
		n := min(len(base.subgr[i]), nSubgraphs)
		if i == 1 { // special classes
			n = len(base.subgr[i])
		}
		r.subgr0[i] = make([]int32, n)
		for k := range n {
			r.subgr0[i][k] = base.subgr[i][base.prios[i][k]]
		}
	}
	num := make([]int32, len(base.objs))
	for i := range num {
		num[i] = 0xff
	}
	var setup func(id int32, class int32)
	setup = func(id int32, class int32) {
		o := &base.objs[id]
		if o.term || num[id] == class {
			return
		}
		num[id] = class
		setup(o.fan0, class)
		setup(o.fan1, class)
		r.nodes0[class] = append(r.nodes0[class], id)
	}
	for i := range r.subgr0 {
		for _, out := range r.subgr0[i] {
			setup(out, int32(i))
		}
		r.nodes0Max = max(r.nodes0Max, int32(len(r.nodes0[i])))
	}
	for i := 0; i < 4; i++ {
		num[i] = int32(i)
	}
	r.objNum = num
	if darNativeLib.prepared == nil {
		darNativeLib.prepared = map[int]*darLibPrepared{}
	}
	darNativeLib.prepared[nSubgraphs] = r
	return r
}

// darRewriter holds Dar_Man_t's rewriting state for one pass.
type darRewriter struct {
	m        *Manager
	g        *aig.Graph
	base     *darLibBase
	prep     *darLibPrepared
	objNum   []int32 // per-run copy of the library objects' Num fields
	datas    []darLibDat
	cutStart []int32 // first cut of each object in cutMem, or -1
	cutMem   []darCut
	cutNodes []int32
	nCutsMax int32

	leavesBest           [4]aig.Edge
	outBest, gainBest    int32
	levelBest, classBest int32
}

func (r *darRewriter) cutsOf(id int32) []darCut {
	if int(id) >= len(r.cutStart) || r.cutStart[id] < 0 {
		return nil
	}
	s := r.cutStart[id]
	return r.cutMem[s : s+r.nCutsMax]
}

func (r *darRewriter) setNoCuts(id int32) {
	if int(id) < len(r.cutStart) {
		r.cutStart[id] = -1
	}
}

func darObjCutSign(id int32) uint32 { return 1 << (uint32(id) & 31) }

// darCutFindValue is Dar_CutFindValue.
func (r *darRewriter) darCutFindValue(c *darCut) int32 {
	var value, ones int32
	for i := int32(0); i < c.nLeaves; i++ {
		id := c.leaves[i]
		if !r.g.ValidEdge(aig.Edge(id) << 1) {
			return 0
		}
		refs := aig.Aig_ObjRefs(r.g, aig.Edge(id)<<1)
		value += refs
		if refs == 1 {
			ones++
		}
	}
	if c.nLeaves < 2 {
		return 1001
	}
	if value > 1000 {
		value = 1000
	}
	if ones > 3 {
		value = 5 - ones
	}
	return value
}

// darCutFindFree is Dar_CutFindFree.
func darCutFindFree(cuts []darCut) *darCut {
	var cutMax *darCut
	for i := range cuts {
		c := &cuts[i]
		if !c.used {
			return c
		}
		if c.nLeaves < 3 {
			continue
		}
		if cutMax == nil || cutMax.value > c.value {
			cutMax = c
		}
	}
	if cutMax == nil {
		for i := range cuts {
			c := &cuts[i]
			if c.nLeaves < 2 {
				continue
			}
			if cutMax == nil || cutMax.value > c.value {
				cutMax = c
			}
		}
	}
	if cutMax == nil {
		for i := range cuts {
			c := &cuts[i]
			if cutMax == nil || cutMax.value > c.value {
				cutMax = c
			}
		}
	}
	cutMax.used = false
	return cutMax
}

// darCutCheckDominance is Dar_CutCheckDominance.
func darCutCheckDominance(dom, c *darCut) bool {
	for i := int32(0); i < dom.nLeaves; i++ {
		k := int32(0)
		for ; k < c.nLeaves; k++ {
			if dom.leaves[i] == c.leaves[k] {
				break
			}
		}
		if k == c.nLeaves {
			return false
		}
	}
	return true
}

// darCutFilter is Dar_CutFilter.
func darCutFilter(cuts []darCut, c *darCut) bool {
	for i := range cuts {
		t := &cuts[i]
		if !t.used || t == c {
			continue
		}
		if t.nLeaves > c.nLeaves {
			if t.sign&c.sign != c.sign {
				continue
			}
			if darCutCheckDominance(c, t) {
				t.used = false
			}
		} else {
			if t.sign&c.sign != t.sign {
				continue
			}
			if darCutCheckDominance(t, c) {
				c.used = false
				return true
			}
		}
	}
	return false
}

// darCutMergeOrdered is Dar_CutMergeOrdered.
func darCutMergeOrdered(c, c0, c1 *darCut) bool {
	if c0.nLeaves == 4 && c1.nLeaves == 4 {
		if c0.sign != c1.sign {
			return false
		}
		for i := 0; i < 4; i++ {
			if c0.leaves[i] != c1.leaves[i] {
				return false
			}
		}
		c.leaves, c.nLeaves = c0.leaves, c0.nLeaves
		return true
	}
	if c0.nLeaves == 4 {
		if c0.sign&c1.sign != c1.sign {
			return false
		}
		for i := int32(0); i < c1.nLeaves; i++ {
			k := c0.nLeaves - 1
			for ; k >= 0; k-- {
				if c0.leaves[k] == c1.leaves[i] {
					break
				}
			}
			if k == -1 {
				return false
			}
		}
		c.leaves, c.nLeaves = c0.leaves, c0.nLeaves
		return true
	}
	i, k := int32(0), int32(0)
	var n int32
	for n = 0; n < 4; n++ {
		if k == c1.nLeaves {
			if i == c0.nLeaves {
				c.nLeaves = n
				return true
			}
			c.leaves[n] = c0.leaves[i]
			i++
			continue
		}
		if i == c0.nLeaves {
			c.leaves[n] = c1.leaves[k]
			k++
			continue
		}
		if c0.leaves[i] < c1.leaves[k] {
			c.leaves[n] = c0.leaves[i]
			i++
			continue
		}
		if c0.leaves[i] > c1.leaves[k] {
			c.leaves[n] = c1.leaves[k]
			k++
			continue
		}
		c.leaves[n] = c0.leaves[i]
		i++
		k++
	}
	if i < c0.nLeaves || k < c1.nLeaves {
		return false
	}
	c.nLeaves = n
	return true
}

// darCutMerge is Dar_CutMerge.
func darCutMerge(c, c0, c1 *darCut) bool {
	if c0.nLeaves <= c1.nLeaves {
		if !darCutMergeOrdered(c, c1, c0) {
			return false
		}
	} else if !darCutMergeOrdered(c, c0, c1) {
		return false
	}
	c.sign = c0.sign | c1.sign
	c.used = true
	return true
}

func darCutTruthPhase(c, c1 *darCut) uint32 {
	var phase uint32
	k := int32(0)
	for i := int32(0); i < c.nLeaves; i++ {
		if k == c1.nLeaves {
			break
		}
		if c.leaves[i] < c1.leaves[k] {
			continue
		}
		phase |= 1 << uint(i)
		k++
	}
	return phase
}

func darCutTruthSwapAdjacentVars(t uint32, v int) uint32 {
	switch v {
	case 0:
		return t&0x99999999 | (t&0x22222222)<<1 | (t&0x44444444)>>1
	case 1:
		return t&0xC3C3C3C3 | (t&0x0C0C0C0C)<<2 | (t&0x30303030)>>2
	default:
		return t&0xF00FF00F | (t&0x00F000F0)<<4 | (t&0x0F000F00)>>4
	}
}

func darCutTruthStretch(t uint32, nVars int32, phase uint32) uint32 {
	v := int(nVars - 1)
	for i := 3; i >= 0; i-- {
		if phase&(1<<uint(i)) != 0 {
			for k := v; k < i; k++ {
				t = darCutTruthSwapAdjacentVars(t, k)
			}
			v--
		}
	}
	return t
}

func darCutTruthShrink(t uint32, phase uint32) uint32 {
	v := 0
	for i := 0; i < 4; i++ {
		if phase&(1<<uint(i)) != 0 {
			for k := i - 1; k >= v; k-- {
				t = darCutTruthSwapAdjacentVars(t, k)
			}
			v++
		}
	}
	return t
}

func darCutTruth(c, c0, c1 *darCut, compl0, compl1 bool) uint32 {
	t0, t1 := c0.truth, c1.truth
	if compl0 {
		t0 = ^t0
	}
	if compl1 {
		t1 = ^t1
	}
	t0 = darCutTruthStretch(t0, c0.nLeaves, darCutTruthPhase(c, c0))
	t1 = darCutTruthStretch(t1, c1.nLeaves, darCutTruthPhase(c, c1))
	return t0 & t1
}

// darCutSuppMinimize is Dar_CutSuppMinimize.
func darCutSuppMinimize(c *darCut) bool {
	masks := [4][2]uint32{{0x5555, 0xAAAA}, {0x3333, 0xCCCC}, {0x0F0F, 0xF0F0}, {0x00FF, 0xFF00}}
	t := 0xFFFF & c.truth
	var phase uint32
	n := c.nLeaves
	for i := int32(0); i < c.nLeaves; i++ {
		if t&masks[i][0] == (t&masks[i][1])>>(1<<uint(i)) {
			n--
		} else {
			phase |= 1 << uint(i)
		}
	}
	if n == c.nLeaves {
		return false
	}
	c.truth = 0xFFFF & darCutTruthShrink(t, phase)
	c.sign = 0
	k := 0
	for i := int32(0); i < c.nLeaves; i++ {
		if phase&(1<<uint(i)) == 0 {
			continue
		}
		c.leaves[k] = c.leaves[i]
		c.sign |= darObjCutSign(c.leaves[i])
		k++
	}
	c.nLeaves = n
	return true
}

// darObjPrepareCuts is Dar_ObjPrepareCuts.
func (r *darRewriter) darObjPrepareCuts(id int32) []darCut {
	for int(id) >= len(r.cutStart) {
		r.cutStart = append(r.cutStart, -1)
	}
	start := int32(len(r.cutMem))
	r.cutMem = append(r.cutMem, make([]darCut, r.nCutsMax)...)
	r.cutStart[id] = start
	r.cutNodes = append(r.cutNodes, id)
	cuts := r.cutMem[start : start+r.nCutsMax]
	c := &cuts[0]
	c.used = true
	if r.g.Kind(aig.Edge(id)<<1) == aig.Constant {
		c.nLeaves, c.sign, c.truth = 0, 0, 0xFFFF
	} else {
		c.nLeaves, c.leaves[0], c.sign, c.truth = 1, id, darObjCutSign(id), 0xAAAA
	}
	c.value = r.darCutFindValue(c)
	return cuts
}

// darManCutsRestart is Dar_ManCutsRestart.
func (r *darRewriter) darManCutsRestart() {
	for i := range r.cutStart {
		r.cutStart[i] = -1
	}
	r.cutNodes = r.cutNodes[:0]
	r.cutMem = r.cutMem[:0]
	r.darObjPrepareCuts(0)
}

// darObjComputeCuts is Dar_ObjComputeCuts (fSkipTtMin = 0).
func (r *darRewriter) darObjComputeCuts(id int32) []darCut {
	g := r.g
	e := aig.Edge(id) << 1
	f0, f1 := aig.Aig_ObjReal_rec(g, aig.Aig_ObjChild0(g, e)), aig.Aig_ObjReal_rec(g, aig.Aig_ObjChild1(g, e))
	cutSet := r.darObjPrepareCuts(id)
	cuts0, cuts1 := r.cutsOf(int32(f0.NodeID())), r.cutsOf(int32(f1.NodeID()))
	for i := range cuts0 {
		c0 := &cuts0[i]
		if !c0.used {
			continue
		}
		for k := range cuts1 {
			c1 := &cuts1[k]
			if !c1.used {
				continue
			}
			r.m.Stats.CutsAll++
			if bits.OnesCount32(c0.sign|c1.sign) > 4 {
				continue
			}
			c := darCutFindFree(cutSet)
			if !darCutMerge(c, c0, c1) {
				continue
			}
			r.m.Stats.CutsTried++
			if darCutFilter(cutSet, c) {
				continue
			}
			c.truth = 0xFFFF & darCutTruth(c, c0, c1, f0.Complemented(), f1.Complemented())
			if darCutSuppMinimize(c) {
				darCutFilter(cutSet, c)
			}
			c.value = r.darCutFindValue(c)
			if c.value == 0 {
				r.m.Stats.CutsSkipped++
				c.used = false
			} else if c.nLeaves < 2 {
				return cutSet
			}
		}
	}
	// count the nontrivial cuts (the trivial one is discounted)
	for k := range cutSet {
		if cutSet[k].used {
			r.m.Stats.CutsUsed++
		}
	}
	r.m.Stats.CutsUsed--
	return cutSet
}

// darObjComputeCutsRec is Dar_ObjComputeCuts_rec.
func (r *darRewriter) darObjComputeCutsRec(e aig.Edge) []darCut {
	id := int32(e.NodeID())
	if cuts := r.cutsOf(id); cuts != nil {
		return cuts
	}
	switch r.g.Kind(e) {
	case aig.Input, aig.Constant:
		return r.darObjPrepareCuts(id)
	case aig.Buffer:
		return r.darObjComputeCutsRec(aig.Aig_ObjFanin0(r.g, e))
	}
	r.darObjComputeCutsRec(aig.Aig_ObjFanin0(r.g, e))
	r.darObjComputeCutsRec(aig.Aig_ObjFanin1(r.g, e))
	return r.darObjComputeCuts(id)
}

// darLibCutMatch is Dar_LibCutMatch.
func (r *darRewriter) darLibCutMatch(c *darCut) bool {
	phase := r.base.phases[c.truth]
	perm := r.base.perms4[r.base.perms[c.truth]]
	for i := 0; i < 4; i++ {
		id := c.leaves[perm[i]]
		e := aig.Edge(id) << 1
		if !r.g.ValidEdge(e) {
			r.m.Stats.CutsBad++
			return false
		}
		if (phase>>uint(i))&1 != 0 {
			e = e.Not()
		}
		r.datas[i].fn = e
		r.datas[i].level = aig.Aig_ObjLevel(r.g, e)
		if r.m.Params.Power {
			r.datas[i].prob = edgeProbability(r.m.probabilities, e)
		}
	}
	return true
}

// darLibCutMarkMffc is Dar_LibCutMarkMffc.
func (r *darRewriter) darLibCutMarkMffc(root aig.Edge, power *float32) int32 {
	for i := 0; i < 4; i++ {
		aig.Aig_ObjRef(r.g, aig.Aig_Regular(r.datas[i].fn))
	}
	var probs []float32
	if power != nil {
		probs = r.m.probabilities
	}
	n := aig.Aig_NodeMffcLabel(r.g, root, power, probs)
	for i := 0; i < 4; i++ {
		aig.Aig_ObjDeref(r.g, aig.Aig_Regular(r.datas[i].fn))
	}
	return n
}

// darLibEvalAssignNums is Dar_LibEvalAssignNums.
func (r *darRewriter) darLibEvalAssignNums(class int32, root aig.Edge) {
	g := r.g
	for i, id := range r.prep.nodes0[class] {
		o := &r.base.objs[id]
		r.objNum[id] = 4 + int32(i)
		d := &r.datas[r.objNum[id]]
		d.mffc, d.fn, d.travID = false, aig.InvalidEdge, 0xFFFF
		d0, d1 := &r.datas[r.objNum[o.fan0]], &r.datas[r.objNum[o.fan1]]
		d.level = 1 + max(d0.level, d1.level)
		if d0.fn == aig.InvalidEdge || d1.fn == aig.InvalidEdge {
			continue
		}
		x, y := d0.fn, d1.fn
		if o.compl0 {
			x = x.Not()
		}
		if o.compl1 {
			y = y.Not()
		}
		if x.NodeID() == root.NodeID() || y.NodeID() == root.NodeID() {
			continue
		}
		d.fn = aig.Aig_TableLookupTwo(g, x, y)
		if d.fn != aig.InvalidEdge {
			d.level = aig.Aig_ObjLevel(g, d.fn)
			d.mffc = aig.Aig_ObjIsTravIdCurrent(g, aig.Aig_Regular(d.fn))
			if r.m.Params.Power {
				d.prob = edgeProbability(r.m.probabilities, d.fn)
			}
		}
	}
}

// darLibEvalRec is Dar_LibEval_rec.
func (r *darRewriter) darLibEvalRec(id int32, out int32, saved, required int32, power *float32) int32 {
	if power != nil {
		*power = 0
	}
	o := &r.base.objs[id]
	d := &r.datas[r.objNum[id]]
	if d.travID == out {
		return 0
	}
	d.travID = out
	if o.term {
		if power != nil {
			*power = d.prob
		}
		return 0
	}
	if d.level > required {
		return 0xff
	}
	if d.fn != aig.InvalidEdge && !d.mffc {
		if power != nil {
			*power = d.prob
		}
		return 0
	}
	saved--
	var p0, p1 float32
	var pp0, pp1 *float32
	if power != nil {
		pp0, pp1 = &p0, &p1
	}
	area := r.darLibEvalRec(o.fan0, out, saved, required+1, pp0)
	if area > saved {
		return 0xff
	}
	area += r.darLibEvalRec(o.fan1, out, saved, required+1, pp1)
	if area > saved {
		return 0xff
	}
	if power != nil {
		d0, d1 := &r.datas[r.objNum[o.fan0]], &r.datas[r.objNum[o.fan1]]
		q0, q1 := d0.prob, d1.prob
		if o.compl0 {
			q0 = 1.0 - q0
		}
		if o.compl1 {
			q1 = 1.0 - q1
		}
		d.prob = q0 * q1
		*power = p0 + 2.0*d0.prob*(1.0-d0.prob) + p1 + 2.0*d1.prob*(1.0-d1.prob)
	}
	return area + 1
}

// darLibEval is Dar_LibEval.
func (r *darRewriter) darLibEval(root aig.Edge, c *darCut, required int32) {
	if c.nLeaves != 4 {
		return
	}
	if !r.darLibCutMatch(c) {
		return
	}
	r.m.Stats.Cuts++
	var powerSaved, powerAdded float32
	var pSaved, pAdded *float32
	if r.m.Params.Power {
		pSaved, pAdded = &powerSaved, &powerAdded
	}
	saved := r.darLibCutMarkMffc(root, pSaved)
	class := int32(r.base.class[c.truth])
	r.darLibEvalAssignNums(class, root)
	useZeros := int32(0)
	if !r.m.Params.UseZeros {
		useZeros = 1
	}
	for out, id := range r.prep.subgr0[class] {
		d := &r.datas[r.objNum[id]]
		if d.fn != aig.InvalidEdge && d.fn.NodeID() == root.NodeID() {
			continue
		}
		added := r.darLibEvalRec(id, int32(out), saved-useZeros, required, pAdded)
		gained := saved - added
		if r.m.Params.Power && powerSaved < powerAdded {
			r.m.Stats.PowerRejected++
			continue
		}
		if gained < 0 || gained == 0 && !r.m.Params.UseZeros {
			continue
		}
		if gained < r.gainBest || gained == r.gainBest && d.level >= r.levelBest {
			continue
		}
		for k := 0; k < 4; k++ {
			r.leavesBest[k] = r.datas[k].fn
		}
		r.outBest, r.levelBest, r.gainBest, r.classBest = id, d.level, gained, class
	}
}

// darLibBuildBest is Dar_LibBuildBest.
func (r *darRewriter) darLibBuildBest() aig.Edge {
	for i := 0; i < 4; i++ {
		r.datas[i].fn = r.leavesBest[i]
	}
	counter := int32(4)
	var clear func(id int32)
	clear = func(id int32) {
		o := &r.base.objs[id]
		if o.term {
			return
		}
		r.objNum[id] = counter
		counter++
		r.datas[r.objNum[id]].fn = aig.InvalidEdge
		clear(o.fan0)
		clear(o.fan1)
	}
	clear(r.outBest)
	var build func(id int32) aig.Edge
	build = func(id int32) aig.Edge {
		o := &r.base.objs[id]
		d := &r.datas[r.objNum[id]]
		if d.fn != aig.InvalidEdge {
			return d.fn
		}
		a, b := build(o.fan0), build(o.fan1)
		if o.compl0 {
			a = a.Not()
		}
		if o.compl1 {
			b = b.Not()
		}
		d.fn = r.g.AND(a, b)
		return d.fn
	}
	return build(r.outBest)
}

// rewriteNative is Dar_ManRewrite.
func (m *Manager) rewriteNative() error {
	g := m.Graph
	base := darLibBaseGet()
	prep := darLibPrepare(m.Params.SubgraphsMax)
	r := &darRewriter{m: m, g: g, base: base, prep: prep,
		objNum: append([]int32(nil), prep.objNum...), nCutsMax: int32(m.Params.CutsMax)}
	r.datas = make([]darLibDat, prep.nodes0Max+32)
	aig.Aig_ManCleanup(g)
	if m.Params.Fanout {
		aig.Aig_ManFanoutStart(g)
		defer aig.Aig_ManFanoutStop(g)
	}
	if m.Params.UpdateLevel {
		aig.Aig_ManStartReverseLevels(g, 0)
		defer aig.Aig_ManStopReverseLevels(g)
	}
	// Dar_ManStart prepares the constant's cuts through Dar_ManCutsRestart
	// only when recycling; ABC computes them on demand otherwise. Neither
	// case reaches the constant in a strashed AIG.
	clkStart := time.Now()
	defer func() {
		m.Stats.TimeTotal = time.Since(clkStart)
		m.Stats.CutMemUsed = max(m.Stats.CutMemUsed, len(r.cutMem)*24>>20) // sizeof(Dar_Cut_t) = 24
	}()
	m.Stats.NodesInit = int(aig.Aig_ManNodeNum(g))
	nNodesOld := int32(aig.Aig_ManObjNumMax(g))
	counter := 0
	for i := int32(0); i < int32(aig.Aig_ManObjNumMax(g)); i++ {
		if m.Params.Context != nil && i&256 == 0 {
			if err := m.Params.Context.Err(); err != nil {
				return err
			}
		}
		root := aig.Edge(i) << 1
		if !g.ValidEdge(root) || g.Kind(root) != aig.And {
			continue
		}
		if i > nNodesOld {
			break
		}
		counter++
		if m.Params.Recycle && counter%50000 == 0 && int(aig.Aig_DagSize(g, root)) < len(r.cutNodes)/100 {
			r.darManCutsRestart()
		}
		m.Stats.Roots++
		m.Stats.NodesTried++
		clk := time.Now()
		r.setNoCuts(i)
		r.darObjComputeCutsRec(root)
		m.Stats.TimeCuts += time.Since(clk)
		cuts := r.cutsOf(i)
		var trivial *darCut
		for k := range cuts {
			c := &cuts[k]
			if !c.used {
				continue
			}
			if c.nLeaves == 0 || c.nLeaves == 1 && c.leaves[0] != i && g.ValidEdge(aig.Edge(c.leaves[0])<<1) {
				trivial = c
				break
			}
		}
		if trivial != nil {
			var repl aig.Edge
			if trivial.nLeaves == 0 {
				repl = g.Const1()
				if trivial.truth == 0 {
					repl = repl.Not()
				}
			} else {
				repl = aig.Edge(trivial.leaves[0]) << 1
				if trivial.truth == 0x5555 {
					repl = repl.Not()
				}
			}
			r.setNoCuts(i)
			aig.Aig_ObjReplace(g, root, repl, btoi32(m.Params.UpdateLevel))
			m.extendProbabilities() // buffer propagation can create nodes
			m.Stats.Replacements++
			continue
		}
		r.gainBest = -1
		required := int32(1<<30 - 1)
		if m.Params.UpdateLevel {
			required = aig.Aig_ObjRequiredLevel(g, root)
		}
		clk = time.Now()
		for k := range cuts {
			c := &cuts[k]
			if !c.used {
				continue
			}
			nLeavesOld := c.nLeaves
			if c.nLeaves == 3 {
				c.leaves[c.nLeaves] = 0
				c.nLeaves++
			}
			r.darLibEval(root, c, required)
			c.nLeaves = nLeavesOld
		}
		m.Stats.TimeEval += time.Since(clk)
		if !(r.gainBest > 0 || r.gainBest == 0 && m.Params.UseZeros) {
			continue
		}
		r.setNoCuts(i)
		repl := r.darLibBuildBest()
		m.extendProbabilities()
		if aig.Aig_ObjPhaseReal(g, repl) != g.Node(root).Phase {
			repl = repl.Not()
		}
		aig.Aig_ObjReplace(g, root, repl, btoi32(m.Params.UpdateLevel))
		m.extendProbabilities()
		m.Stats.Replacements++
		m.Stats.NodesSaved += int(r.gainBest)
	}
	aig.Aig_ManCheckPhase(g)
	return g.Check()
}

func btoi32(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
