package sbd

import (
	"cmp"
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"
	"slices"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

const SBD_MAX_CUTSIZE = 10
const SBD_MAX_CUTNUM = 501
const SBD_MAX_TT_WORDS = 16
const SBD_CUT_NO_LEAF = 15

type Sbd_Cut_t struct {
	Sign                                                                uint64
	iFunc, Cost, CostLev, nTreeLeaves, nSlowLeaves, nTopLeaves, nLeaves int32
	pLeaves                                                             [SBD_MAX_CUTSIZE]int32
}
type Sbd_Sto_t_ struct {
	nLutSize, nCutSize, nCutNum, fCutMin, fVerbose          int32
	pGia                                                    *Network
	vMirrors                                                *vec.Vec_Int_t
	vDelays, vLevels, vRefs                                 *vec.Vec_Int_t
	cuts                                                    [][]Sbd_Cut_t
	truths                                                  [][]uint64
	truthIndex                                              map[string]int32
	ppCuts                                                  []*Sbd_Cut_t
	nCutsR, Pivot, iCutBest, nCutsSpec, nCutsOver, DelayMin int32
	CutCount                                                [4]float64
}

// Alias is colocated with implementation so explicit-file tests are useful
// while the larger package is being ported.
type Sbd_Sto_t = Sbd_Sto_t_

func Sbd_CutGetSign(cut *Sbd_Cut_t) uint64 {
	var sign uint64
	for _, leaf := range cut.pLeaves[:cut.nLeaves] {
		sign |= uint64(1) << uint(leaf&63)
	}
	return sign
}
func Sbd_CutCheck(base, cut *Sbd_Cut_t) int32 {
	for _, leaf := range cut.pLeaves[:cut.nLeaves] {
		if !slices.Contains(base.pLeaves[:base.nLeaves], leaf) {
			return 0
		}
	}
	return 1
}
func Sbd_CutSetCutIsContainedOrder(base, cut *Sbd_Cut_t) int32 {
	if base.nLeaves < cut.nLeaves {
		return 0
	}
	return Sbd_CutCheck(base, cut)
}
func Sbd_CutSetCheckArray(cuts []*Sbd_Cut_t, count int32) int32 {
	if count < 1 {
		panic("empty cut set")
	}
	for i, c := range cuts[:count] {
		if c.nLeaves > 10 || c.Sign != Sbd_CutGetSign(c) {
			panic("invalid cut signature")
		}
		for k := int32(1); k < c.nLeaves; k++ {
			if c.pLeaves[k-1] >= c.pLeaves[k] {
				panic("unordered cut")
			}
		}
		for j, d := range cuts[:count] {
			if i != j && Sbd_CutCheck(c, d) != 0 {
				panic("dominated cut")
			}
		}
	}
	return 1
}
func Sbd_CutMergeOrder(a, b, out *Sbd_Cut_t, limit int32) int32 {
	i, j, n := int32(0), int32(0), int32(0)
	var leaves [10]int32
	for i < a.nLeaves || j < b.nLeaves {
		if n == limit {
			return 0
		}
		var leaf int32
		switch {
		case i == a.nLeaves:
			leaf = b.pLeaves[j]
			j++
		case j == b.nLeaves:
			leaf = a.pLeaves[i]
			i++
		case a.pLeaves[i] < b.pLeaves[j]:
			leaf = a.pLeaves[i]
			i++
		case a.pLeaves[i] > b.pLeaves[j]:
			leaf = b.pLeaves[j]
			j++
		default:
			leaf = a.pLeaves[i]
			i++
			j++
		}
		leaves[n] = leaf
		n++
	}
	out.pLeaves = leaves
	out.nLeaves = n
	out.iFunc = -1
	out.Sign = a.Sign | b.Sign
	return 1
}
func Sbd_CutMergeOrder2(a, b, out *Sbd_Cut_t, limit int32) int32 {
	return Sbd_CutMergeOrder(a, b, out, limit)
}
func Sbd_CutSetLastCutIsContained(cuts []*Sbd_Cut_t, count int32) int32 {
	last := cuts[count]
	for _, cut := range cuts[:count] {
		if cut.nLeaves <= last.nLeaves && cut.Sign&last.Sign == cut.Sign && Sbd_CutCheck(last, cut) != 0 {
			return 1
		}
	}
	return 0
}
func Sbd_CutCompare(a, b *Sbd_Cut_t) int {
	if a.nLeaves <= 4 && b.nLeaves <= 4 {
		if v := cmp.Compare(a.nLeaves, b.nLeaves); v != 0 {
			return v
		}
	} else {
		if a.nLeaves <= 4 {
			return -1
		}
		if b.nLeaves <= 4 {
			return 1
		}
		if v := cmp.Compare(a.nTreeLeaves, b.nTreeLeaves); v != 0 {
			return v
		}
	}
	if v := cmp.Compare(a.Cost, b.Cost); v != 0 {
		return v
	}
	if v := cmp.Compare(a.CostLev, b.CostLev); v != 0 {
		return v
	}
	return cmp.Compare(a.nLeaves, b.nLeaves)
}
func Sbd_CutCompare2(a, b *Sbd_Cut_t) int {
	if v := cmp.Compare(a.nSlowLeaves, b.nSlowLeaves); v != 0 {
		return v
	}
	if v := cmp.Compare(a.nTreeLeaves, b.nTreeLeaves); v != 0 {
		return v
	}
	if v := cmp.Compare(a.Cost, b.Cost); v != 0 {
		return v
	}
	if v := cmp.Compare(a.CostLev, b.CostLev); v != 0 {
		return v
	}
	return cmp.Compare(a.nLeaves, b.nLeaves)
}
func Sbd_CutSetLastCutContains(cuts []*Sbd_Cut_t, count int32) int32 {
	last := cuts[count]
	at := int32(0)
	for i := int32(0); i <= count; i++ {
		c := cuts[i]
		if i != count && last.nLeaves < c.nLeaves && last.Sign&c.Sign == last.Sign && Sbd_CutCheck(c, last) != 0 {
			c.nLeaves = SBD_CUT_NO_LEAF
			continue
		}
		cuts[at], cuts[i] = cuts[i], cuts[at]
		at++
	}
	return at - 1
}
func Sbd_CutSetSortByCost(cuts []*Sbd_Cut_t, count int32) {
	for i := count; i > 0; i-- {
		if Sbd_CutCompare(cuts[i-1], cuts[i]) < 0 {
			return
		}
		cuts[i-1], cuts[i] = cuts[i], cuts[i-1]
	}
}
func Sbd_CutSetAddCut(cuts []*Sbd_Cut_t, count, limit int32) int32 {
	if count == 0 {
		return 1
	}
	count = Sbd_CutSetLastCutContains(cuts, count)
	Sbd_CutSetSortByCost(cuts, count)
	return min(count+1, limit-1)
}
func sbdTruthKey(truth []uint64) string {
	key := make([]byte, len(truth)*8)
	for i, v := range truth {
		binary.LittleEndian.PutUint64(key[8*i:], v)
	}
	return string(key)
}
func (p *Sbd_Sto_t_) intern(truth []uint64) int32 {
	key := sbdTruthKey(truth)
	if index, ok := p.truthIndex[key]; ok {
		return index
	}
	index := int32(len(p.truths))
	p.truthIndex[key] = index
	p.truths = append(p.truths, slices.Clone(truth))
	return index
}
func Sbd_CutTruth(p *Sbd_Sto_t, cut *Sbd_Cut_t) []uint64 { return p.truths[cut.iFunc>>1] }
func Sbd_CutComputeTruth(p *Sbd_Sto_t, a, b *Sbd_Cut_t, complA, complB int32, result *Sbd_Cut_t, xor int32) int32 {
	old := result.nLeaves
	nWords := int32(1)
	if p.nCutSize > 6 {
		nWords <<= p.nCutSize - 6
	}
	ta, tb := slices.Clone(Sbd_CutTruth(p, a)), slices.Clone(Sbd_CutTruth(p, b))
	if a.iFunc&1^complA != 0 {
		for i := range ta {
			ta[i] = ^ta[i]
		}
	}
	if b.iFunc&1^complB != 0 {
		for i := range tb {
			tb[i] = ^tb[i]
		}
	}
	util.Abc_TtExpand(ta, p.nCutSize, a.pLeaves[:], a.nLeaves, result.pLeaves[:], result.nLeaves)
	util.Abc_TtExpand(tb, p.nCutSize, b.pLeaves[:], b.nLeaves, result.pLeaves[:], result.nLeaves)
	truth := make([]uint64, nWords)
	for i := range truth {
		if xor != 0 {
			truth[i] = ta[i] ^ tb[i]
		} else {
			truth[i] = ta[i] & tb[i]
		}
	}
	compl := int32(truth[0] & 1)
	if compl != 0 {
		for i := range truth {
			truth[i] = ^truth[i]
		}
	}
	result.nLeaves = util.Abc_TtMinBase(truth, result.pLeaves[:], result.nLeaves, p.nCutSize)
	result.iFunc = 2*p.intern(truth) + compl
	if result.nLeaves < old {
		return 1
	}
	return 0
}
func Sbd_CutComputeTruth6(p *Sbd_Sto_t, a, b *Sbd_Cut_t, ca, cb int32, r *Sbd_Cut_t, x int32) int32 {
	return Sbd_CutComputeTruth(p, a, b, ca, cb, r, x)
}
func Sbd_CutCountBits(v uint64) int32 { return int32(bits.OnesCount64(v)) }
func Sbd_CutCost(p *Sbd_Sto_t, c *Sbd_Cut_t) int32 {
	var result int32
	for _, v := range c.pLeaves[:c.nLeaves] {
		result += vec.Vec_IntEntry(p.vDelays, int(v))
	}
	return result
}
func Sbd_CutCostLev(p *Sbd_Sto_t, c *Sbd_Cut_t) int32 {
	var result int32
	for _, v := range c.pLeaves[:c.nLeaves] {
		result += vec.Vec_IntEntry(p.vLevels, int(v))
	}
	return result
}
func Sbd_CutTreeLeaves(p *Sbd_Sto_t, c *Sbd_Cut_t) int32 {
	var result int32
	for _, v := range c.pLeaves[:c.nLeaves] {
		if vec.Vec_IntEntry(p.vRefs, int(v)) == 1 {
			result++
		}
	}
	return result
}
func Sbd_CutSlowLeaves(p *Sbd_Sto_t, id int32, c *Sbd_Cut_t) int32 {
	var result int32
	delay := vec.Vec_IntEntry(p.vDelays, int(id))
	for _, v := range c.pLeaves[:c.nLeaves] {
		if vec.Vec_IntEntry(p.vDelays, int(v))-delay >= -1 {
			result++
		}
	}
	return result
}
func Sbd_CutTopLeaves(p *Sbd_Sto_t, id int32, c *Sbd_Cut_t) int32 {
	var result int32
	delay := vec.Vec_IntEntry(p.vDelays, int(id))
	for _, v := range c.pLeaves[:c.nLeaves] {
		if vec.Vec_IntEntry(p.vDelays, int(v))-delay == -2 {
			result++
		}
	}
	return result
}
func Sbd_CutAddUnit(p *Sbd_Sto_t, id int32) {
	c := Sbd_Cut_t{nLeaves: 1, iFunc: 2}
	c.pLeaves[0] = id
	c.Sign = Sbd_CutGetSign(&c)
	p.cuts[id] = append(p.cuts[id], c)
}
func Sbd_CutAddZero(p *Sbd_Sto_t, id int32) {
	if len(p.cuts[id]) != 0 {
		panic("constant cuts already initialized")
	}
	p.cuts[id] = []Sbd_Cut_t{{iFunc: 0}}
}
func (p *Sbd_Sto_t_) prepare(id int32) []Sbd_Cut_t {
	cuts := slices.Clone(p.cuts[id])
	for i := range cuts {
		c := &cuts[i]
		c.Sign = Sbd_CutGetSign(c)
		c.Cost = Sbd_CutCost(p, c)
		c.CostLev = Sbd_CutCostLev(p, c)
		c.nTreeLeaves = Sbd_CutTreeLeaves(p, c)
		c.nSlowLeaves = Sbd_CutSlowLeaves(p, id, c)
		c.nTopLeaves = Sbd_CutTopLeaves(p, id, c)
	}
	return cuts
}
func Sbd_StoComputeDelay(p *Sbd_Sto_t, id int32, cuts []*Sbd_Cut_t, count int32) {
	minimum := int32(math.MaxInt32)
	p.iCutBest = -1
	for i, c := range cuts[:count] {
		if c.nLeaves > p.nLutSize {
			continue
		}
		delay := int32(0)
		for _, v := range c.pLeaves[:c.nLeaves] {
			delay = max(delay, vec.Vec_IntEntry(p.vDelays, int(v)))
		}
		if delay < minimum || delay == minimum && p.iCutBest >= 0 && cuts[p.iCutBest].nLeaves > c.nLeaves {
			minimum = delay
			p.iCutBest = int32(i)
		}
	}
	if p.iCutBest < 0 {
		panic("no LUT-sized cut")
	}
	if count > 1 || cuts[0].nLeaves > 1 {
		minimum++
	}
	vec.Vec_IntWriteEntry(p.vDelays, int(id), minimum)
	p.DelayMin = max(p.DelayMin, minimum)
}
func Sbd_StoComputeSpec(p *Sbd_Sto_t, id int32, cuts []*Sbd_Cut_t, count int32) {
	for _, c := range cuts[:count] {
		c.nTopLeaves = Sbd_CutTopLeaves(p, id, c)
		c.nSlowLeaves = Sbd_CutSlowLeaves(p, id, c)
		if c.nSlowLeaves == 0 {
			p.nCutsSpec++
		}
	}
}
func Sbd_CutPrint(p *Sbd_Sto_t, id int32, c *Sbd_Cut_t) {
	if c == nil {
		fmt.Println("No cut.")
		return
	}
	fmt.Printf("%d {%v} Cost=%d CostL=%d Tree=%d Slow=%d Top=%d\n", c.nLeaves, c.pLeaves[:c.nLeaves], c.Cost, c.CostLev, c.nTreeLeaves, c.nSlowLeaves, c.nTopLeaves)
}
func Sbd_StoMergeCuts(p *Sbd_Sto_t, id int32) {
	node := p.pGia.object(id)
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("cut merge requires binary gate")
	}
	resolve := func(edge aig.Edge) int32 {
		lit := sourceLiteral(edge)
		if p.vMirrors != nil {
			if replacement := vec.Vec_IntEntry(p.vMirrors, edge.NodeID()); replacement >= 0 {
				lit = replacement ^ (lit & 1)
			}
		}
		return lit
	}
	a, b := resolve(node.Fanins[0]), resolve(node.Fanins[1])
	left, right := p.prepare(a>>1), p.prepare(b>>1)
	storage := make([]Sbd_Cut_t, p.nCutNum)
	p.ppCuts = make([]*Sbd_Cut_t, p.nCutNum)
	for i := range storage {
		p.ppCuts[i] = &storage[i]
	}
	count := int32(0)
	p.CutCount[0] += float64(len(left) * len(right))
	for i := range left {
		for j := range right {
			ca, cb := &left[i], &right[j]
			if ca.nLeaves+cb.nLeaves > p.nCutSize && bits.OnesCount64(ca.Sign|cb.Sign) > int(p.nCutSize) {
				continue
			}
			p.CutCount[1]++
			result := p.ppCuts[count]
			if Sbd_CutMergeOrder(ca, cb, result, p.nCutSize) == 0 || Sbd_CutSetLastCutIsContained(p.ppCuts, count) != 0 {
				continue
			}
			p.CutCount[2]++
			if p.fCutMin != 0 {
				isXor := int32(0)
				if node.Kind == aig.ExclusiveOr {
					isXor = 1
				}
				if Sbd_CutComputeTruth(p, ca, cb, a&1, b&1, result, isXor) != 0 {
					result.Sign = Sbd_CutGetSign(result)
				}
			}
			result.Cost = Sbd_CutCost(p, result)
			result.CostLev = Sbd_CutCostLev(p, result)
			result.nTreeLeaves = Sbd_CutTreeLeaves(p, result)
			count = Sbd_CutSetAddCut(p.ppCuts, count, p.nCutNum)
		}
	}
	Sbd_StoComputeDelay(p, id, p.ppCuts, count)
	Sbd_StoComputeSpec(p, id, p.ppCuts, count)
	p.CutCount[3] += float64(count)
	if count == p.nCutNum-1 {
		p.nCutsOver++
	}
	p.nCutsR, p.Pivot = count, id
	Sbd_CutSetCheckArray(p.ppCuts, count)
	p.cuts[id] = p.cuts[id][:0]
	for _, c := range p.ppCuts[:count] {
		p.cuts[id] = append(p.cuts[id], *c)
	}
	if count > 1 || p.ppCuts[0].nLeaves > 1 {
		Sbd_CutAddUnit(p, id)
	}
}
func Sbd_StoAlloc(network *Network, mirrors *vec.Vec_Int_t, lutSize, cutSize, cutNum, minimize, verbose int32) *Sbd_Sto_t {
	if lutSize < 1 || lutSize > cutSize || cutSize < 2 || cutSize > 10 || cutNum < 2 || cutNum >= 501 {
		panic("invalid SBD cut dimensions")
	}
	count := len(network.Graph.Nodes())
	p := &Sbd_Sto_t{pGia: network, vMirrors: mirrors, nLutSize: lutSize, nCutSize: cutSize, nCutNum: cutNum, fCutMin: minimize, fVerbose: verbose, vDelays: vec.Vec_IntStart(count), vLevels: vec.Vec_IntStart(count), vRefs: vec.Vec_IntAlloc(count), cuts: make([][]Sbd_Cut_t, count), truthIndex: map[string]int32{}}
	words := 1
	if cutSize > 6 {
		words <<= cutSize - 6
	}
	p.intern(make([]uint64, words))
	variable := make([]uint64, words)
	for i := range variable {
		variable[i] = 0xaaaaaaaaaaaaaaaa
	}
	p.intern(variable)
	return p
}
func Sbd_StoFree(p *Sbd_Sto_t) {
	if p != nil {
		*p = Sbd_Sto_t{}
	}
}
func Sbd_StoComputeCutsObj(p *Sbd_Sto_t, id, delay, level int32) {
	if int(id) == len(p.cuts) {
		p.cuts = append(p.cuts, nil)
		vec.Vec_IntPush(p.vDelays, delay)
		vec.Vec_IntPush(p.vLevels, level)
	} else {
		vec.Vec_IntWriteEntry(p.vDelays, int(id), delay)
		vec.Vec_IntWriteEntry(p.vLevels, int(id), level)
	}
}
func Sbd_StoComputeCutsConst0(p *Sbd_Sto_t, id int32) {
	Sbd_StoComputeCutsObj(p, id, 0, 0)
	Sbd_CutAddZero(p, id)
}
func Sbd_StoComputeCutsCi(p *Sbd_Sto_t, id, delay, level int32) {
	Sbd_StoComputeCutsObj(p, id, delay, level)
	Sbd_CutAddUnit(p, id)
}
func Sbd_StoComputeCutsNode(p *Sbd_Sto_t, id int32) int32 {
	node := p.pGia.object(id)
	level := 1 + max(vec.Vec_IntEntry(p.vLevels, node.Fanins[0].NodeID()), vec.Vec_IntEntry(p.vLevels, node.Fanins[1].NodeID()))
	Sbd_StoComputeCutsObj(p, id, -1, level)
	Sbd_StoMergeCuts(p, id)
	return vec.Vec_IntEntry(p.vDelays, int(id))
}
func Sbd_StoSaveBestDelayCut(p *Sbd_Sto_t, id int32, out []int32) {
	if id != p.Pivot {
		panic("cut pivot mismatch")
	}
	c := p.ppCuts[p.iCutBest]
	out[0] = c.nLeaves
	copy(out[1:], c.pLeaves[:c.nLeaves])
}
func Sbd_StoObjRefs(p *Sbd_Sto_t, id int32) int32 { return vec.Vec_IntEntry(p.vRefs, int(id)) }
func Sbd_StoRefObj(p *Sbd_Sto_t, id, mirror int32) {
	if int(id) != vec.Vec_IntSize(p.vRefs) || mirror >= id {
		panic("nonsequential cut references")
	}
	vec.Vec_IntPush(p.vRefs, 0)
	if mirror > 0 {
		vec.Vec_IntWriteEntry(p.vRefs, int(id), vec.Vec_IntEntry(p.vRefs, int(mirror)))
		vec.Vec_IntWriteEntry(p.vRefs, int(mirror), 1)
	}
	node := p.pGia.object(id)
	fanins := 0
	switch node.Kind {
	case aig.And, aig.ExclusiveOr:
		fanins = 2
	case aig.Output, aig.Buffer:
		fanins = 1
	}
	for _, edge := range node.Fanins[:fanins] {
		fan := int32(edge.NodeID())
		if p.vMirrors != nil {
			if lit := vec.Vec_IntEntry(p.vMirrors, int(fan)); lit >= 0 {
				fan = lit >> 1
			}
		}
		vec.Vec_IntAddToEntry(p.vRefs, int(fan), 1)
	}
}

// The original function returns unconditionally before its dereference code.
func Sbd_StoDerefObj(_ *Sbd_Sto_t, _ int32) {}
func Sbd_StoObjBestCut(p *Sbd_Sto_t, id, size int32, out []int32) int32 {
	if id != p.Pivot {
		panic("cut pivot mismatch")
	}
	var best *Sbd_Cut_t
	for _, c := range p.ppCuts[:p.nCutsR] {
		if size != 0 && c.nLeaves != size {
			continue
		}
		if c.nLeaves > p.nLutSize && c.nSlowLeaves <= 1 && c.nTopLeaves <= p.nLutSize-1 && (best == nil || Sbd_CutCompare2(best, c) > 0) {
			best = c
		}
	}
	if best == nil {
		return -1
	}
	copy(out, best.pLeaves[:best.nLeaves])
	return best.nLeaves
}
func Sbd_StoComputeCutsTest(network *Network) {
	p := Sbd_StoAlloc(network, nil, 4, 8, 100, 1, 1)
	defer Sbd_StoFree(p)
	for _, edge := range network.Graph.LiveEdges() {
		Sbd_StoRefObj(p, int32(edge.NodeID()), -1)
	}
	Sbd_StoComputeCutsConst0(p, 0)
	for _, edge := range network.Graph.Inputs() {
		Sbd_StoComputeCutsCi(p, int32(edge.NodeID()), 0, 0)
	}
	for _, edge := range network.Graph.LiveEdges() {
		kind := network.Graph.Node(edge).Kind
		if kind == aig.And || kind == aig.ExclusiveOr {
			Sbd_StoComputeCutsNode(p, int32(edge.NodeID()))
		}
	}
	fmt.Printf("CutPair=%.0f Merge=%.0f Eval=%.0f Cut=%.0f Lev=%d\n", p.CutCount[0], p.CutCount[1], p.CutCount[2], p.CutCount[3], p.DelayMin)
}
