package gia

import (
	"fmt"
	"math/bits"
	"sort"
)

const (
	BAL_LEAF_MAX = 6
	BAL_CUT_MAX  = 8
	BAL_SUPER    = 50
	BAL_NO_LEAF  = 31
	BAL_NO_FUNC  = 134217727
)

type Bal_Cut_t struct {
	Sign   uint64
	Delay  int32
	Func   int32
	Leaves []int32
}
type Bal_Man_t struct {
	Source, Result  *Gia_Man_t
	LutSize, CutNum int32
	Verbose         bool
	Costs           []int32
	Cuts            [][]Bal_Cut_t
}

func Bal_ManAlloc(src, dst *Gia_Man_t, lutSize, cutNum, verbose int32) *Bal_Man_t {
	if lutSize < 2 || lutSize > BAL_LEAF_MAX || cutNum < 2 || cutNum > BAL_CUT_MAX {
		panic("gia: LUT balancing needs LUT size 2..6 and cut count 2..8")
	}
	return &Bal_Man_t{Source: src, Result: dst, LutSize: lutSize, CutNum: cutNum, Verbose: verbose != 0, Costs: make([]int32, dst.ObjectCount()), Cuts: make([][]Bal_Cut_t, dst.ObjectCount())}
}
func Bal_ManFree(p *Bal_Man_t) {
	if p != nil {
		*p = Bal_Man_t{}
	}
}
func Bal_ObjCost(p *Bal_Man_t, id int32) int32   { return p.Costs[id] }
func Bal_LitCost(p *Bal_Man_t, lit int32) int32  { return Bal_ObjCost(p, lit>>1) }
func Bal_ObjDelay(p *Bal_Man_t, id int32) int32  { return Bal_ObjCost(p, id) >> 4 }
func Bal_LitDelay(p *Bal_Man_t, lit int32) int32 { return Bal_ObjDelay(p, lit>>1) }
func Bal_CutCountBits(sign uint64) int32         { return int32(bits.OnesCount64(sign)) }
func Bal_CutGetSign(leaves []int32) uint64 {
	var sign uint64
	for _, leaf := range leaves {
		sign |= 1 << uint(leaf&63)
	}
	return sign
}
func Bal_CutCreateUnit(c *Bal_Cut_t, id, delay int32) int32 {
	*c = Bal_Cut_t{Sign: 1 << uint(id&63), Delay: delay, Func: 2, Leaves: []int32{id}}
	return 1
}
func Bal_ManPrepareSet(p *Bal_Man_t, id, index, unit int32) []Bal_Cut_t {
	if unit != 0 || p.Cuts[id] == nil {
		var c Bal_Cut_t
		Bal_CutCreateUnit(&c, id, Bal_ObjDelay(p, id)+1)
		return []Bal_Cut_t{c}
	}
	return p.Cuts[id]
}
func Bal_CutCheck(base, cut *Bal_Cut_t) bool {
	if len(base.Leaves) < len(cut.Leaves) {
		return false
	}
	i := 0
	for _, leaf := range base.Leaves {
		if i < len(cut.Leaves) && leaf == cut.Leaves[i] {
			i++
		}
	}
	return i == len(cut.Leaves)
}
func Bal_SetCutIsContainedOrder(base, cut *Bal_Cut_t) bool { return Bal_CutCheck(base, cut) }
func Bal_SetCheckArray(cuts []Bal_Cut_t) bool {
	if len(cuts) == 0 {
		return false
	}
	for i := range cuts {
		c := &cuts[i]
		if len(c.Leaves) > BAL_LEAF_MAX || c.Sign != Bal_CutGetSign(c.Leaves) {
			return false
		}
		for k := 1; k < len(c.Leaves); k++ {
			if c.Leaves[k-1] >= c.Leaves[k] {
				return false
			}
		}
		for j := range cuts {
			if i != j && Bal_CutCheck(c, &cuts[j]) {
				return false
			}
		}
	}
	return true
}
func Bal_CutMergeOrder(a, b, out *Bal_Cut_t, limit int32) bool {
	leaves := make([]int32, 0, len(a.Leaves)+len(b.Leaves))
	i, j := 0, 0
	for i < len(a.Leaves) || j < len(b.Leaves) {
		var leaf int32
		if j == len(b.Leaves) || i < len(a.Leaves) && a.Leaves[i] < b.Leaves[j] {
			leaf = a.Leaves[i]
			i++
		} else if i == len(a.Leaves) || b.Leaves[j] < a.Leaves[i] {
			leaf = b.Leaves[j]
			j++
		} else {
			leaf = a.Leaves[i]
			i++
			j++
		}
		leaves = append(leaves, leaf)
		if len(leaves) > int(limit) {
			return false
		}
	}
	*out = Bal_Cut_t{a.Sign | b.Sign, max(a.Delay, b.Delay), BAL_NO_FUNC, leaves}
	return true
}
func Bal_CutMergeOrderMux(a, b, c, out *Bal_Cut_t, limit int32) bool {
	var temp Bal_Cut_t
	return Bal_CutMergeOrder(a, b, &temp, limit) && Bal_CutMergeOrder(&temp, c, out, limit)
}
func Bal_CutCompareArea(a, b *Bal_Cut_t) int32 {
	if a.Delay < b.Delay {
		return -1
	}
	if a.Delay > b.Delay {
		return 1
	}
	if len(a.Leaves) < len(b.Leaves) {
		return -1
	}
	if len(a.Leaves) > len(b.Leaves) {
		return 1
	}
	return 0
}

// Retain at most CutNum-1 non-dominated cuts. New cuts precede exact cost
// ties, matching the source insertion order and subsequent pair choices.
func balLutAddCut(cuts []Bal_Cut_t, c Bal_Cut_t, capacity int32) []Bal_Cut_t {
	for i := range cuts {
		if cuts[i].Sign&c.Sign == cuts[i].Sign && Bal_CutCheck(&c, &cuts[i]) {
			return cuts
		}
	}
	kept := cuts[:0]
	for i := range cuts {
		if len(c.Leaves) < len(cuts[i].Leaves) && c.Sign&cuts[i].Sign == c.Sign && Bal_CutCheck(&cuts[i], &c) {
			continue
		}
		kept = append(kept, cuts[i])
	}
	index := len(kept)
	kept = append(kept, c)
	for index > 0 && Bal_CutCompareArea(&kept[index-1], &kept[index]) >= 0 {
		kept[index-1], kept[index] = kept[index], kept[index-1]
		index--
	}
	if len(kept) >= int(capacity) {
		kept = kept[:capacity-1]
	}
	return kept
}
func Bal_ManDeriveCuts(p *Bal_Man_t, fan0, fan1, fan2, comp0, comp1, comp2, unit0, unit1, unit2, isXor, target, save int32) int32 {
	a := Bal_ManPrepareSet(p, fan0, 0, unit0)
	b := Bal_ManPrepareSet(p, fan1, 1, unit1)
	var sets []Bal_Cut_t
	if fan2 > 0 {
		sets = Bal_ManPrepareSet(p, fan2, 2, unit2)
	}
	var result []Bal_Cut_t
	add := func(c Bal_Cut_t) {
		if c.Delay != target {
			panic("gia: LUT cut delay disagrees with target")
		}
		result = balLutAddCut(result, c, p.CutNum)
	}
	for i := range a {
		for j := range b {
			if fan2 > 0 {
				for k := range sets {
					if Bal_CutCountBits(a[i].Sign|b[j].Sign|sets[k].Sign) > p.LutSize {
						continue
					}
					var c Bal_Cut_t
					if Bal_CutMergeOrderMux(&a[i], &b[j], &sets[k], &c, p.LutSize) {
						add(c)
					}
				}
			} else {
				if Bal_CutCountBits(a[i].Sign|b[j].Sign) > p.LutSize {
					continue
				}
				var c Bal_Cut_t
				if Bal_CutMergeOrder(&a[i], &b[j], &c, p.LutSize) {
					add(c)
				}
			}
		}
	}
	if len(result) == 0 {
		return -1
	}
	if !Bal_SetCheckArray(result) {
		panic("gia: invalid LUT cut antichain")
	}
	cost := result[0].Delay<<4 | int32(len(result[0].Leaves))
	if save != 0 {
		p.Cuts = append(p.Cuts, result)
		p.Costs = append(p.Costs, cost)
		if len(p.Cuts) != p.Result.ObjectCount() {
			panic("gia: LUT cut/object indexing mismatch")
		}
	}
	return cost
}
func Bal_ManSetGateLevel(p *Bal_Man_t, old *Gia_Obj_t, lit int32) int32 {
	id := lit >> 1
	if int(id) < len(p.Cuts) {
		return -1
	}
	a, b, c, kind := p.Result.Gate(id)
	fan0, fan1, fan2 := a>>1, b>>1, int32(0)
	if kind == 4 {
		fan0, fan1, fan2 = c>>1, b>>1, a>>1
	}
	delay := max(Bal_ObjDelay(p, fan0), Bal_ObjDelay(p, fan1), Bal_ObjDelay(p, fan2))
	unit := func(id int32) int32 {
		if Bal_ObjDelay(p, id) != delay {
			return 1
		}
		return 0
	}
	if delay > 0 {
		cost := Bal_ManDeriveCuts(p, fan0, fan1, fan2, 0, 0, 0, unit(fan0), unit(fan1), unit(fan2), 0, delay, 1)
		if cost >= 0 {
			return cost
		}
	}
	cost := Bal_ManDeriveCuts(p, fan0, fan1, fan2, 0, 0, 0, 1, 1, 1, 0, delay+1, 1)
	if cost < 0 {
		panic("gia: gate does not fit LUT")
	}
	return cost
}
func Bal_ManEvalTwo(p *Bal_Man_t, a, b, c, isXor int32) int32 {
	fan0, fan1, fan2 := a>>1, b>>1, c>>1
	delay := max(Bal_ObjDelay(p, fan0), Bal_ObjDelay(p, fan1), Bal_ObjDelay(p, fan2))
	if delay == 0 {
		return -1
	}
	unit := func(id int32) int32 {
		if Bal_ObjDelay(p, id) != delay {
			return 1
		}
		return 0
	}
	return Bal_ManDeriveCuts(p, fan0, fan1, fan2, 0, 0, 0, unit(fan0), unit(fan1), unit(fan2), isXor, delay, 0)
}
func balLutSortCosts(p *Bal_Man_t, lits []int32) {
	for i := 0; i < len(lits)-1; i++ {
		best := i
		for j := i + 1; j < len(lits); j++ {
			if Bal_LitCost(p, lits[j]) > Bal_LitCost(p, lits[best]) {
				best = j
			}
		}
		lits[i], lits[best] = lits[best], lits[i]
	}
}
func balLutInsertCost(p *Bal_Man_t, lits []int32, lit int32) []int32 {
	lits = append(lits, lit)
	for i := len(lits) - 1; i > 0; i-- {
		if Bal_LitCost(p, lits[i]) <= Bal_LitCost(p, lits[i-1]) {
			break
		}
		lits[i], lits[i-1] = lits[i-1], lits[i]
	}
	return lits
}
func Bal_ManFindBestPair(p *Bal_Man_t, lits []int32, isXor int32) (int, int) {
	end := len(lits) - 1
	begin := end
	for begin > 0 && Bal_LitDelay(p, lits[begin-1]) == Bal_LitDelay(p, lits[end]) {
		begin--
	}
	bestCost := int32(1 << 30)
	bestI, bestK := -1, -1
	for k := begin - 1; k >= 0; k-- {
		for i := end; i >= begin; i-- {
			cost := Bal_ManEvalTwo(p, lits[i], lits[k], 0, isXor)
			if cost < 0 {
				continue
			}
			if cost == Bal_LitCost(p, lits[k]) {
				return k, i
			}
			if cost < bestCost {
				bestCost, bestI, bestK = cost, i, k
			}
		}
	}
	if bestK >= 0 && bestCost>>4 == Bal_LitDelay(p, lits[bestK]) {
		return bestK, bestI
	}
	bestCost = 1 << 30
	for i := begin; i <= end; i++ {
		for k := i + 1; k <= end; k++ {
			cost := Bal_ManEvalTwo(p, lits[i], lits[k], 0, isXor)
			if cost < 0 {
				continue
			}
			if cost == max(Bal_LitCost(p, lits[i]), Bal_LitCost(p, lits[k])) {
				return k, i
			}
			if cost < bestCost {
				bestCost, bestI, bestK = cost, i, k
			}
		}
	}
	if bestCost < 1<<30 {
		return bestK, bestI
	}
	return end, end - 1
}

// These helpers are LUT-specific: the AIG-balancing family has distinct
// supergate limits and pairing rules and keeps its original Gia_* helpers.
func giaLutSimplify(lits []int32, isXor bool) []int32 {
	sort.Slice(lits, func(i, j int) bool { return lits[i] < lits[j] })
	out := lits[:0]
	if isXor {
		phase := int32(0)
		for _, l := range lits {
			if l == 0 {
				continue
			}
			if l == 1 {
				phase ^= 1
				continue
			}
			if len(out) > 0 && out[len(out)-1] == l {
				out = out[:len(out)-1]
			} else {
				out = append(out, l)
			}
		}
		if len(out) == 0 {
			return append(out, phase)
		}
		out[0] ^= phase
		return out
	}
	for _, l := range lits {
		if l == 0 {
			return []int32{0}
		}
		if l == 1 {
			continue
		}
		if len(out) > 0 && out[len(out)-1]>>1 == l>>1 {
			if out[len(out)-1] != l {
				return []int32{0}
			}
			continue
		}
		out = append(out, l)
	}
	if len(out) == 0 {
		return []int32{1}
	}
	return out
}
func giaLutCollect(src *Gia_Man_t, id int32) []int32 {
	a, b, _, kind := src.Gate(id)
	isXor := kind == 3
	var lits []int32
	var collect func(int32)
	collect = func(lit int32) {
		x, y, _, k := src.Gate(lit >> 1)
		flatten := k == 2 && !isXor || k == 3 && isXor
		if lit&1 != 0 || !flatten || Gia_ObjRefNumId(src, lit>>1) > 3 || len(lits) > BAL_SUPER {
			lits = append(lits, lit)
			return
		}
		collect(x)
		collect(y)
	}
	collect(a)
	collect(b)
	return giaLutSimplify(lits, isXor)
}
func giaLutBalanceGate(p *Bal_Man_t, lits []int32, isXor bool) int32 {
	if len(lits) == 0 {
		panic("gia: empty supergate")
	}
	lits = append([]int32(nil), lits...)
	if len(lits) > 2 {
		balLutSortCosts(p, lits)
	}
	xorFlag := int32(0)
	if isXor {
		xorFlag = 1
	}
	for len(lits) > 1 {
		a, b := len(lits)-1, len(lits)-2
		if len(lits) > 2 {
			a, b = Bal_ManFindBestPair(p, lits, xorFlag)
		}
		x, y := lits[a], lits[b]
		if a < b {
			a, b = b, a
		}
		lits = append(lits[:a], lits[a+1:]...)
		lits = append(lits[:b], lits[b+1:]...)
		var result int32
		if isXor {
			result = Gia_ManHashXorReal(p.Result, x, y)
		} else {
			result = Gia_ManHashAnd(p.Result, x, y)
		}
		Bal_ManSetGateLevel(p, nil, result)
		lits = balLutInsertCost(p, lits, result)
	}
	return lits[0]
}
func giaLutBalanceInt(src *Gia_Man_t, lutSize, cutNum, verbose int32) *Gia_Man_t {
	src.Refs = nil
	Gia_ManCreateRefs(src)
	dst := Gia_ManStart(3 * src.nObjs / 2)
	dst.CopyIdentityFrom(src)
	dst.EnableMuxes()
	copies := make([]int32, src.nObjs)
	for i := range copies {
		copies[i] = -1
	}
	copies[0] = 0
	for _, lit := range src.InputLiterals() {
		copies[lit>>1] = Gia_ManAppendCi(dst)
	}
	p := Bal_ManAlloc(src, dst, lutSize, cutNum, verbose)
	Gia_ManHashStart(dst)
	var visit func(int32) int32
	visit = func(lit int32) int32 {
		id := lit >> 1
		if copies[id] >= 0 {
			return copies[id] ^ (lit & 1)
		}
		a, b, c, kind := src.Gate(id)
		var result int32
		switch kind {
		case 4:
			control, yes, no := visit(a), visit(b), visit(c)
			result = Gia_ManHashMuxReal(dst, control, yes, no)
			Bal_ManSetGateLevel(p, nil, result)
		case 2, 3:
			lits := giaLutCollect(src, id)
			for i, l := range lits {
				lits[i] = visit(l)
			}
			result = giaLutBalanceGate(p, lits, kind == 3)
		case 5, 6:
			result = visit(a)
		default:
			panic("gia: invalid LUT balancing node")
		}
		copies[id] = result
		return result ^ (lit & 1)
	}
	outs := src.OutputLiterals()
	mapped := make([]int32, len(outs))
	for i, lit := range outs {
		mapped[i] = visit(lit)
	}
	delay := int32(0)
	for _, lit := range mapped {
		delay = max(delay, Bal_LitDelay(p, lit))
		Gia_ManAppendCo(dst, lit)
	}
	fmt.Printf("Best delay = %d\n", delay)
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, src.nRegs)
	dst.nConstrs = src.nConstrs
	if err := dst.CopyIONamesFrom(src, nil, nil); err != nil {
		panic(err)
	}
	Bal_ManFree(p)
	return Gia_ManCleanup(dst)
}
func Gia_ManBalanceLut(src *Gia_Man_t, lutSize, cutNum, verbose int32) *Gia_Man_t {
	if lutSize < 2 || lutSize > BAL_LEAF_MAX || cutNum < 2 || cutNum > BAL_CUT_MAX {
		panic("gia: LUT balancing needs LUT size 2..6 and cut count 2..8")
	}
	coarse := Gia_ManDupMuxes(src, 2)
	balanced := giaLutBalanceInt(coarse, lutSize, cutNum, verbose)
	result := Gia_ManDupNoMuxes(balanced, 0)
	result.nConstrs = src.nConstrs
	if err := result.CopyIONamesFrom(src, nil, nil); err != nil {
		panic(err)
	}
	return result
}
