package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"sort"
)

func ifLeaves(c *If_Cut_t) []int32 {
	if int(c.nLeaves) > len(c.pLeaves) {
		panic("if: cut leaf count exceeds storage")
	}
	return c.pLeaves[:c.nLeaves]
}
func If_CutVerifyCut(base, c *If_Cut_t) int32 {
	for _, leaf := range ifLeaves(c) {
		found := false
		for _, other := range ifLeaves(base) {
			found = found || leaf == other
		}
		if !found {
			return 0
		}
	}
	return 1
}
func If_CutCheckDominance(dom, c *If_Cut_t) int32 {
	if dom.nLeaves > c.nLeaves {
		return 0
	}
	return If_CutVerifyCut(c, dom)
}
func If_CutVerifyCuts(set *If_Set_t, ordered int32) int32 {
	if set.nCuts <= 0 || int(set.nCuts) > len(set.ppCuts) {
		return 0
	}
	for i, c := range set.ppCuts[:set.nCuts] {
		if c.uSign != If_ObjCutSignCompute(c) {
			return 0
		}
		leaves := ifLeaves(c)
		for j, leaf := range leaves {
			for k := j + 1; k < len(leaves); k++ {
				if leaf == leaves[k] || ordered != 0 && leaf >= leaves[k] {
					return 0
				}
			}
		}
		for j := i + 1; j < int(set.nCuts); j++ {
			d := set.ppCuts[j]
			if If_CutCheckDominance(c, d) != 0 || If_CutCheckDominance(d, c) != 0 {
				return 0
			}
		}
	}
	return 1
}
func If_CutFilter(set *If_Set_t, c *If_Cut_t, saveFirst int32) int32 {
	if set.nCuts < 0 || int(set.nCuts) >= len(set.ppCuts) || set.ppCuts[set.nCuts] != c {
		panic("if: filter candidate is outside scratch slot")
	}
	for i := 0; i < int(set.nCuts); i++ {
		other := set.ppCuts[i]
		if other.nLeaves > c.nLeaves {
			if i == 0 && (set.nCuts > 1 && set.ppCuts[1].fUseless != 0 || saveFirst != 0 && set.nCuts == 1) {
				continue
			}
			if other.uSign&c.uSign == c.uSign && If_CutCheckDominance(c, other) != 0 {
				old := set.nCuts
				copy(set.ppCuts[i:old], set.ppCuts[i+1:old+1])
				set.ppCuts[old] = other
				set.nCuts--
				i--
			}
		} else if other.uSign&c.uSign == other.uSign && If_CutCheckDominance(other, c) != 0 {
			return 1
		}
	}
	return 0
}
func ifMerge(p *If_Man_t, a, b, c *If_Cut_t, ordered, permutations bool) int32 {
	limit := int(a.nLimit)
	var leaves []int32
	var perm0, perm1, shared [IF_MAX_LUTSIZE]int32
	sharedMask := uint32(0)
	nShared := 0
	if ordered {
		i, j := 0, 0
		x, y := ifLeaves(a), ifLeaves(b)
		for i < len(x) || j < len(y) {
			if len(leaves) == limit {
				return 0
			}
			position := int32(len(leaves))
			switch {
			case j == len(y) || i < len(x) && x[i] < y[j]:
				leaves = append(leaves, x[i])
				perm0[i] = position
				i++
			case i == len(x) || y[j] < x[i]:
				leaves = append(leaves, y[j])
				perm1[j] = position
				j++
			default:
				leaves = append(leaves, x[i])
				perm0[i], perm1[j] = position, position
				shared[nShared] = position
				nShared++
				sharedMask |= 1 << uint(position)
				i++
				j++
			}
		}
	} else {
		leaves = append(leaves, ifLeaves(a)...)
		for i, leaf := range ifLeaves(b) {
			position := -1
			for j, x := range leaves {
				if leaf == x {
					position = j
					break
				}
			}
			if position < 0 {
				if len(leaves) == limit {
					return 0
				}
				position = len(leaves)
				leaves = append(leaves, leaf)
			}
			perm1[i] = int32(position)
		}
	}
	c.pLeaves = append(c.pLeaves[:0], leaves...)
	c.nLeaves = uint32(len(leaves))
	c.uSign = a.uSign | b.uSign
	if permutations {
		p.pPerm[0], p.pPerm[1], p.pPerm[2] = perm0, perm1, shared
		p.uSharedMask = sharedMask
	} else if !ordered {
		p.pPerm[1] = perm1
	}
	return 1
}
func If_CutMergeOrdered_(p *If_Man_t, a, b, c *If_Cut_t) int32 {
	return ifMerge(p, a, b, c, true, true)
}
func If_CutMergeOrdered(p *If_Man_t, a, b, c *If_Cut_t) int32 {
	return ifMerge(p, a, b, c, true, false)
}
func If_CutMerge(p *If_Man_t, a, b, c *If_Cut_t) int32 { return ifMerge(p, a, b, c, false, false) }
func ifCompare(p *If_Man_t, a, b *If_Cut_t, fields string) int32 {
	for _, field := range fields {
		var x, y, eps float32
		switch field {
		case 'd':
			x, y, eps = a.Delay, b.Delay, p.fEpsilon
		case 'a':
			x, y, eps = a.Area, b.Area, p.fEpsilon
		case 'e':
			x, y, eps = a.Edge, b.Edge, p.fEpsilon
		case 'p':
			x, y, eps = a.Power, b.Power, p.fEpsilon
		case 'n':
			x, y = float32(a.nLeaves), float32(b.nLeaves)
		case 'u':
			x, y = float32(a.fUseless), float32(b.fUseless)
		}
		if x < y-eps {
			return -1
		}
		if x > y+eps {
			return 1
		}
	}
	return 0
}
func If_CutCompareDelay(p *If_Man_t, a, b **If_Cut_t) int32    { return ifCompare(p, *a, *b, "dna") }
func If_CutCompareDelayOld(p *If_Man_t, a, b **If_Cut_t) int32 { return ifCompare(p, *a, *b, "dan") }
func If_CutCompareArea(p *If_Man_t, a, b **If_Cut_t) int32     { return ifCompare(p, *a, *b, "and") }
func If_ManSortCompare(p *If_Man_t, a, b *If_Cut_t) int32 {
	fields := ""
	if p.pPars.FPower != 0 {
		switch p.SortMode {
		case 0:
			fields = "dnape"
		case 1:
			fields = "apend"
		case 2:
			fields = "dpean"
		}
	} else {
		switch p.SortMode {
		case 0:
			fields = "dnaepu"
		case 1:
			fields = "aepndu"
		case 2:
			fields = "duaepn"
		}
	}
	if fields == "" {
		panic("if: unknown cut sort mode")
	}
	return ifCompare(p, a, b, fields)
}
func If_ManSortCompare_old(p *If_Man_t, a, b *If_Cut_t) int32 {
	if p.SortMode == 1 {
		return ifCompare(p, a, b, "and")
	}
	if p.SortMode == 0 {
		return ifCompare(p, a, b, "dna")
	}
	return ifCompare(p, a, b, "dan")
}
func If_CutSort(p *If_Man_t, set *If_Set_t, c *If_Cut_t) {
	n := int(set.nCuts)
	if n >= len(set.ppCuts) || set.ppCuts[n] != c || set.nCuts > set.nCutsMax {
		panic("if: invalid candidate cut slot")
	}
	if n == 0 {
		set.nCuts++
		return
	}
	pars := p.pPars
	special := pars.FUseDsd != 0 || pars.pFuncCell2 != nil || pars.FUseBat != 0 || pars.PLutStruct != "" || pars.FUserRecLib != 0 || pars.FUserSesLib != 0 || pars.FEnableCheck07 != 0 || pars.FUseCofVars != 0 || pars.FUseAndVars != 0 || pars.FUse34Spec != 0 || pars.FUseDsdTune != 0 || pars.FEnableCheck75 != 0 || pars.FEnableCheck75u != 0
	if c.fUseless == 0 && special && (set.ppCuts[0].fUseless != 0 || If_ManSortCompare(p, set.ppCuts[0], c) == 1) {
		first := set.ppCuts[0]
		set.ppCuts[0], set.ppCuts[n] = c, first
		If_CutSort(p, set, first)
		return
	}
	for i := n - 1; i >= 0; i-- {
		if If_ManSortCompare(p, set.ppCuts[i], c) <= 0 || i == 0 && set.ppCuts[0].fUseless == 0 && c.fUseless != 0 {
			break
		}
		set.ppCuts[i+1] = set.ppCuts[i]
		set.ppCuts[i] = c
	}
	if set.nCuts < set.nCutsMax {
		set.nCuts++
	}
}
func If_CutOrder(c *If_Cut_t) {
	leaves := ifLeaves(c)
	sort.Slice(leaves, func(i, j int) bool { return leaves[i] < leaves[j] })
	for i := 1; i < len(leaves); i++ {
		if leaves[i] == leaves[i-1] {
			panic("if: duplicate cut leaf")
		}
	}
}
func If_CutCheck(c *If_Cut_t) int32 {
	if c.nLeaves > c.nLimit || int(c.nLeaves) > len(c.pLeaves) {
		return 0
	}
	for i := 1; i < int(c.nLeaves); i++ {
		if c.pLeaves[i-1] >= c.pLeaves[i] {
			return 0
		}
	}
	return 1
}
func If_CutPrint(c *If_Cut_t) {
	fmt.Print("{")
	for i, leaf := range ifLeaves(c) {
		prefix := ""
		if c.uMaskFunc >= 0 && uint32(c.uMaskFunc)&(1<<uint(i)) != 0 {
			prefix = "!"
		}
		fmt.Printf(" %s%d", prefix, leaf)
	}
	fmt.Println(" }")
}
func If_CutPrintTiming(p *If_Man_t, c *If_Cut_t) {
	fmt.Print("{")
	for _, id := range ifLeaves(c) {
		o := If_ManObj(p, id)
		fmt.Printf(" %d(%.2f/%.2f)", id, o.CutBest.Delay, o.Required)
	}
	fmt.Println(" }")
}
func If_CutLift(c *If_Cut_t) {
	for i, leaf := range ifLeaves(c) {
		if leaf&255 == 255 {
			panic("if: cut latch offset overflow")
		}
		c.pLeaves[i]++
	}
}
func ifSwitching(p *If_Man_t, id int32) float32 {
	return math.Float32frombits(uint32(vec.Vec_IntEntry(p.vSwitching, int(id))))
}
func ifCutBaseCost(p *If_Man_t, c *If_Cut_t, metric byte) float32 {
	switch metric {
	case 'a':
		return If_CutLutArea(p, c)
	case 'e':
		return float32(c.nLeaves)
	case 'p':
		var cost float32
		for _, id := range ifLeaves(c) {
			cost += ifSwitching(p, id)
		}
		return cost
	}
	panic("if: invalid cut metric")
}
func ifCutFlow(p *If_Man_t, c *If_Cut_t, metric byte) float32 {
	flow := ifCutBaseCost(p, c, metric)
	for _, id := range ifLeaves(c) {
		leaf := If_ManObj(p, id)
		var add float32
		switch metric {
		case 'a':
			add = leaf.CutBest.Area
		case 'e':
			add = leaf.CutBest.Edge
		case 'p':
			add = leaf.CutBest.Power
		}
		if leaf.nRefs != 0 && If_ObjIsConst1(leaf) == 0 {
			if leaf.EstRefs <= p.fEpsilon {
				panic("if: invalid estimated fanout count")
			}
			add /= leaf.EstRefs
		}
		flow += add
		if metric != 'p' && flow > 1e32 {
			flow = 1e32
		}
	}
	return flow
}
func If_CutAreaFlow(p *If_Man_t, c *If_Cut_t) float32                  { return ifCutFlow(p, c, 'a') }
func If_CutEdgeFlow(p *If_Man_t, c *If_Cut_t) float32                  { return ifCutFlow(p, c, 'e') }
func If_CutPowerFlow(p *If_Man_t, c *If_Cut_t, root *If_Obj_t) float32 { return ifCutFlow(p, c, 'p') }
func If_CutAverageRefs(p *If_Man_t, c *If_Cut_t) float32 {
	var total int32
	for _, id := range ifLeaves(c) {
		total += If_ManObj(p, id).nRefs
	}
	if c.nLeaves == 0 {
		return 0
	}
	return float32(total) / float32(c.nLeaves)
}

// Reference journals make invalid cone traversals atomic while preserving the
// source shared-DAG zero-to-one/one-to-zero area accounting.
func ifCutChangeRefs(p *If_Man_t, c *If_Cut_t, delta int32, metric byte) float32 {
	changed := make(map[*If_Obj_t]int32)
	active := make(map[*If_Cut_t]bool)
	ok := false
	defer func() {
		if !ok {
			for o, n := range changed {
				o.nRefs = n
			}
		}
	}()
	var visit func(*If_Cut_t) float32
	visit = func(c *If_Cut_t) float32 {
		if active[c] {
			panic("if: cyclic mapped cut")
		}
		active[c] = true
		defer delete(active, c)
		cost := ifCutBaseCost(p, c, metric)
		for _, id := range ifLeaves(c) {
			leaf := If_ManObj(p, id)
			if leaf.nRefs < 0 || delta < 0 && leaf.nRefs == 0 {
				panic("if: reference underflow")
			}
			if _, seen := changed[leaf]; !seen {
				changed[leaf] = leaf.nRefs
			}
			old := leaf.nRefs
			leaf.nRefs += delta
			if If_ObjIsAnd(leaf) != 0 && (delta > 0 && old == 0 || delta < 0 && leaf.nRefs == 0) {
				cost += visit(&leaf.CutBest)
			}
		}
		return cost
	}
	result := visit(c)
	ok = true
	return result
}
func If_CutAreaRef(p *If_Man_t, c *If_Cut_t) float32   { return ifCutChangeRefs(p, c, 1, 'a') }
func If_CutAreaDeref(p *If_Man_t, c *If_Cut_t) float32 { return ifCutChangeRefs(p, c, -1, 'a') }
func If_CutAreaDerefed(p *If_Man_t, c *If_Cut_t) float32 {
	if c.nLeaves < 2 {
		return 0
	}
	first := ifCutChangeRefs(p, c, 1, 'a')
	second := ifCutChangeRefs(p, c, -1, 'a')
	if float32(math.Abs(float64(first-second))) > 3*p.fEpsilon {
		panic("if: inconsistent reversible cut accounting")
	}
	return second
}
func If_CutAreaRefed(p *If_Man_t, c *If_Cut_t) float32 {
	if c.nLeaves < 2 {
		return 0
	}
	first := ifCutChangeRefs(p, c, -1, 'a')
	second := ifCutChangeRefs(p, c, 1, 'a')
	if float32(math.Abs(float64(first-second))) > 3*p.fEpsilon {
		panic("if: inconsistent reversible cut accounting")
	}
	return second
}
func If_CutEdgeRef(p *If_Man_t, c *If_Cut_t) float32   { return ifCutChangeRefs(p, c, 1, 'e') }
func If_CutEdgeDeref(p *If_Man_t, c *If_Cut_t) float32 { return ifCutChangeRefs(p, c, -1, 'e') }
func If_CutEdgeDerefed(p *If_Man_t, c *If_Cut_t) float32 {
	if c.nLeaves < 2 {
		return float32(c.nLeaves)
	}
	first := ifCutChangeRefs(p, c, 1, 'e')
	second := ifCutChangeRefs(p, c, -1, 'e')
	if float32(math.Abs(float64(first-second))) > 3*p.fEpsilon {
		panic("if: inconsistent reversible cut accounting")
	}
	return second
}
func If_CutEdgeRefed(p *If_Man_t, c *If_Cut_t) float32 {
	if c.nLeaves < 2 {
		return float32(c.nLeaves)
	}
	first := ifCutChangeRefs(p, c, -1, 'e')
	second := ifCutChangeRefs(p, c, 1, 'e')
	if float32(math.Abs(float64(first-second))) > 3*p.fEpsilon {
		panic("if: inconsistent reversible cut accounting")
	}
	return second
}
func If_CutPowerRef(p *If_Man_t, c *If_Cut_t, root *If_Obj_t) float32 {
	return ifCutChangeRefs(p, c, 1, 'p')
}
func If_CutPowerDeref(p *If_Man_t, c *If_Cut_t, root *If_Obj_t) float32 {
	return ifCutChangeRefs(p, c, -1, 'p')
}
func If_CutPowerDerefed(p *If_Man_t, c *If_Cut_t, root *If_Obj_t) float32 {
	if c.nLeaves < 2 {
		return 0
	}
	first := ifCutChangeRefs(p, c, 1, 'p')
	second := ifCutChangeRefs(p, c, -1, 'p')
	if float32(math.Abs(float64(first-second))) > 3*p.fEpsilon {
		panic("if: inconsistent reversible cut accounting")
	}
	return second
}
func If_CutPowerRefed(p *If_Man_t, c *If_Cut_t, root *If_Obj_t) float32 {
	if c.nLeaves < 2 {
		return 0
	}
	first := ifCutChangeRefs(p, c, -1, 'p')
	second := ifCutChangeRefs(p, c, 1, 'p')
	if float32(math.Abs(float64(first-second))) > 3*p.fEpsilon {
		panic("if: inconsistent reversible cut accounting")
	}
	return second
}
func If_CutGetCutMinLevel(p *If_Man_t, c *If_Cut_t) int32 {
	result := int32(IF_INFINITY)
	for _, id := range ifLeaves(c) {
		result = min(result, int32(If_ManObj(p, id).Level))
	}
	return result
}
func If_CutGetCone_rec(p *If_Man_t, o *If_Obj_t, c *If_Cut_t) int32 {
	memo := make(map[*If_Obj_t]bool)
	active := make(map[*If_Obj_t]bool)
	boundary := make(map[int32]bool)
	for _, id := range ifLeaves(c) {
		boundary[id] = true
	}
	var visit func(*If_Obj_t) bool
	visit = func(o *If_Obj_t) bool {
		if o == nil {
			return false
		}
		o = If_Regular(o)
		if boundary[o.Id] {
			return true
		}
		if active[o] {
			panic("if: cycle in choice cone")
		}
		if value, seen := memo[o]; seen {
			return value
		}
		if If_ObjIsCi(o) != 0 || If_ObjIsConst1(o) != 0 {
			return false
		}
		active[o] = true
		found := false
		choices := make(map[*If_Obj_t]bool)
		for alternative := o; alternative != nil; alternative = alternative.pEquiv {
			if choices[alternative] {
				panic("if: cycle in equivalence chain")
			}
			choices[alternative] = true
			if visit(alternative.pFanin0) && visit(alternative.pFanin1) {
				found = true
				break
			}
		}
		delete(active, o)
		memo[o] = found
		return found
	}
	if visit(o) {
		return 1
	}
	return 0
}
func If_CutGetCones(p *If_Man_t) int32 {
	missing := 0
	for _, raw := range vec.Vec_PtrArray(p.vObjs) {
		o := raw.(*If_Obj_t)
		if If_ObjIsAnd(o) != 0 && o.nRefs != 0 && If_CutGetCone_rec(p, o, &o.CutBest) == 0 {
			missing++
		}
	}
	fmt.Printf("Could not find boundary for %d nodes.\n", missing)
	return 1
}
func If_CutFoundFanins_rec(o *If_Obj_t, leaves *Vec_Int_t) {
	seen := make(map[*If_Obj_t]bool)
	var visit func(*If_Obj_t)
	visit = func(o *If_Obj_t) {
		if o == nil {
			return
		}
		o = If_Regular(o)
		if seen[o] {
			return
		}
		seen[o] = true
		if o.nRefs != 0 || If_ObjIsCi(o) != 0 || If_ObjIsConst1(o) != 0 {
			vec.Vec_IntPushUnique(leaves, o.Id)
			return
		}
		visit(o.pFanin0)
		visit(o.pFanin1)
	}
	visit(o)
}
func If_CutCountTotalFanins(p *If_Man_t) int32 {
	total, incremental := 0, 0
	leaves := vec.Vec_IntAlloc(100)
	for _, raw := range vec.Vec_PtrArray(p.vObjs) {
		o := raw.(*If_Obj_t)
		if If_ObjIsAnd(o) != 0 && o.nRefs != 0 {
			total += int(o.CutBest.nLeaves)
			vec.Vec_IntClear(leaves)
			If_CutFoundFanins_rec(o.pFanin0, leaves)
			If_CutFoundFanins_rec(o.pFanin1, leaves)
			incremental += vec.Vec_IntSize(leaves)
		}
	}
	fmt.Printf("Total cut inputs = %d. Total fanins incremental = %d.\n", total, incremental)
	return 1
}
