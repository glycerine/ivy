package csw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
)

func Csw_CutFindCost(p *Csw_Man_t, c *Csw_Cut_t) int32 {
	if len(c.pFanins) == 0 {
		panic("csw: empty cut has no cost")
	}
	sum := int32(0)
	for _, id := range c.pFanins {
		sum += Csw_ObjRefs(p, aig.Edge(id)<<1)
	}
	return sum * 100 / int32(len(c.pFanins))
}
func Csw_CutFindCost2(p *Csw_Man_t, c *Csw_Cut_t) float32 {
	if len(c.pFanins) == 0 {
		panic("csw: empty cut has no cost")
	}
	sum := float32(0)
	for _, id := range c.pFanins {
		sum += 1 / float32(aig.Aig_ObjRefs(p.pManRes, aig.Edge(id)<<1))
	}
	return 1 / sum
}
func Csw_CutFindFree(p *Csw_Man_t, obj aig.Edge) *Csw_Cut_t {
	var worst *Csw_Cut_t
	for _, c := range Csw_ObjForEachCut(p, obj) {
		if len(c.pFanins) == 0 {
			return c
		}
		if worst == nil || worst.Cost < c.Cost {
			worst = c
		}
	}
	if worst == nil {
		panic("csw: node has no cut slots")
	}
	worst.pFanins = worst.pFanins[:0]
	return worst
}
func Cut_TruthPhase(merged, child *Csw_Cut_t) uint32 {
	var phase uint32
	k := 0
	for i, id := range merged.pFanins {
		if k == len(child.pFanins) {
			break
		}
		if id < child.pFanins[k] {
			continue
		}
		if id != child.pFanins[k] {
			panic("csw: merged cut omits a child leaf")
		}
		phase |= 1 << uint(i)
		k++
	}
	if k != len(child.pFanins) {
		panic("csw: incomplete child support")
	}
	return phase
}
func Csw_CutComputeTruth(p *Csw_Man_t, c, a, b *Csw_Cut_t, complA, complB int32) []uint32 {
	if complA != 0 {
		kit.Kit_TruthNot(p.puTemp[0], a.truth, p.nLeafMax)
	} else {
		kit.Kit_TruthCopy(p.puTemp[0], a.truth, p.nLeafMax)
	}
	kit.Kit_TruthStretch(p.puTemp[2], p.puTemp[0], int32(len(a.pFanins)), p.nLeafMax, Cut_TruthPhase(c, a), 0)
	if complB != 0 {
		kit.Kit_TruthNot(p.puTemp[1], b.truth, p.nLeafMax)
	} else {
		kit.Kit_TruthCopy(p.puTemp[1], b.truth, p.nLeafMax)
	}
	kit.Kit_TruthStretch(p.puTemp[3], p.puTemp[1], int32(len(b.pFanins)), p.nLeafMax, Cut_TruthPhase(c, b), 0)
	if p.pManRes.Node(aig.Edge(c.iNode)<<1).Kind == aig.ExclusiveOr {
		kit.Kit_TruthXor(c.truth, p.puTemp[2], p.puTemp[3], p.nLeafMax)
	} else {
		kit.Kit_TruthAnd(c.truth, p.puTemp[2], p.puTemp[3], p.nLeafMax)
	}
	return c.truth
}
func Csw_CutSupportMinimize(p *Csw_Man_t, c *Csw_Cut_t) int32 {
	support := kit.Kit_TruthSupport(c.truth, p.nLeafMax)
	count := bits.OnesCount32(support)
	if count == len(c.pFanins) {
		return int32(count)
	}
	if count > len(c.pFanins) || support>>uint(len(c.pFanins)) != 0 {
		panic("csw: truth exceeds cut support")
	}
	kit.Kit_TruthShrink(p.puTemp[0], c.truth, int32(count), p.nLeafMax, support, 1)
	write := 0
	for i, id := range c.pFanins {
		if support&(1<<uint(i)) != 0 {
			c.pFanins[write] = id
			write++
		}
	}
	c.pFanins = c.pFanins[:write]
	return int32(count)
}
func Csw_CutCheckDominance(a, b *Csw_Cut_t) int32 {
	k := 0
	for _, id := range a.pFanins {
		for k < len(b.pFanins) && b.pFanins[k] < id {
			k++
		}
		if k == len(b.pFanins) || b.pFanins[k] != id {
			return 0
		}
	}
	return 1
}
func Csw_CutFilter(p *Csw_Man_t, obj aig.Edge, c *Csw_Cut_t) int32 {
	for _, old := range Csw_ObjForEachCut(p, obj) {
		if old == c || len(old.pFanins) < 2 {
			continue
		}
		if len(old.pFanins) > len(c.pFanins) {
			if old.uSign&c.uSign == c.uSign && Csw_CutCheckDominance(c, old) != 0 {
				old.pFanins = old.pFanins[:0]
			}
		} else if old.uSign&c.uSign == old.uSign && Csw_CutCheckDominance(old, c) != 0 {
			c.pFanins = c.pFanins[:0]
			return 1
		}
	}
	return 0
}
func Csw_CutMergeOrdered(p *Csw_Man_t, a, b, c *Csw_Cut_t) int32 {
	c.pFanins = c.pFanins[:0]
	i, j := 0, 0
	for i < len(a.pFanins) || j < len(b.pFanins) {
		var id int32
		if j == len(b.pFanins) || i < len(a.pFanins) && a.pFanins[i] < b.pFanins[j] {
			id = a.pFanins[i]
			i++
		} else if i == len(a.pFanins) || a.pFanins[i] > b.pFanins[j] {
			id = b.pFanins[j]
			j++
		} else {
			id = a.pFanins[i]
			i++
			j++
		}
		if len(c.pFanins) == int(p.nLeafMax) {
			c.pFanins = c.pFanins[:0]
			return 0
		}
		c.pFanins = append(c.pFanins, id)
	}
	return 1
}
func Csw_CutMerge(p *Csw_Man_t, a, b, c *Csw_Cut_t) int32 {
	if Csw_CutMergeOrdered(p, a, b, c) == 0 {
		return 0
	}
	c.uSign = a.uSign | b.uSign
	return 1
}
func Csw_ObjTwoVarCut(p *Csw_Man_t, c *Csw_Cut_t) aig.Edge {
	if Csw_CutSupportMinimize(p, c) != 2 {
		panic("csw: two-variable resubstitution has different support")
	}
	a, b := aig.Edge(c.pFanins[0])<<1, aig.Edge(c.pFanins[1])<<1
	t := c.truth[0] & 15
	complement := false
	if t == 14 || t == 13 || t == 11 || t == 7 {
		t = (^t) & 15
		complement = true
	}
	var e aig.Edge
	switch t {
	case 1:
		e = p.pManRes.AND(a.Not(), b.Not())
	case 2:
		e = p.pManRes.AND(a, b.Not())
	case 4:
		e = p.pManRes.AND(a.Not(), b)
	case 8:
		e = p.pManRes.AND(a, b)
	default:
		return aig.InvalidEdge
	}
	if complement {
		e = e.Not()
	}
	return e
}
func Csw_ObjPrepareCuts(p *Csw_Man_t, obj aig.Edge, trivial int32) *Csw_Cut_t {
	slots := make([]Csw_Cut_t, p.nCutsMax)
	words := cswTruthWords(p.nLeafMax)
	leaves := make([]int32, int(p.nCutsMax*p.nLeafMax))
	truths := make([]uint32, int(p.nCutsMax)*words)
	for i := range slots {
		c := &slots[i]
		c.iNode = int32(obj.NodeID())
		start := i * int(p.nLeafMax)
		c.pFanins = leaves[start : start : start+int(p.nLeafMax)]
		c.truth = truths[i*words : (i+1)*words]
		if i+1 < len(slots) {
			c.nextRecord = &slots[i+1]
		}
	}
	head := &slots[0]
	Csw_ObjSetCuts(p, obj, head)
	if trivial != 0 {
		head.pFanins = append(head.pFanins, int32(obj.NodeID()))
		head.uSign = aig.Aig_ObjCutSign(uint32(obj.NodeID()))
		for i := range head.truth {
			head.truth[i] = 0xaaaaaaaa
		}
	}
	return head
}
func Csw_ObjSweep(p *Csw_Man_t, obj aig.Edge, trivial int32) aig.Edge {
	node := p.pManRes.Node(obj)
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		return obj
	}
	if Csw_ObjCuts(p, obj) != nil {
		return obj
	}
	Csw_ObjPrepareCuts(p, obj, trivial)
	for _, a := range Csw_ObjForEachCut(p, node.Fanins[0]) {
		if len(a.pFanins) == 0 {
			continue
		}
		for _, b := range Csw_ObjForEachCut(p, node.Fanins[1]) {
			if len(b.pFanins) == 0 || bits.OnesCount32(a.uSign|b.uSign) > int(p.nLeafMax) {
				continue
			}
			c := Csw_CutFindFree(p, obj)
			start := util.Abc_Clock()
			if Csw_CutMerge(p, a, b, c) == 0 || Csw_CutFilter(p, obj, c) != 0 {
				continue
			}
			truth := Csw_CutComputeTruth(p, c, a, b, int32(node.Fanins[0]&1), int32(node.Fanins[1]&1))
			fans := len(c.pFanins)
			support := kit.Kit_TruthSupport(truth, p.nLeafMax)
			count := bits.OnesCount32(support)
			p.timeCuts += util.Abc_Clock() - start
			if count == 0 {
				p.nNodesTriv0++
				if truth[0]&1 != 0 {
					return p.pManRes.Const1()
				}
				return p.pManRes.Const0()
			}
			if count == 1 {
				p.nNodesTriv1++
				v := bits.TrailingZeros32(support)
				e := aig.Edge(c.pFanins[v]) << 1
				if truth[0]&1 != 0 {
					e = e.Not()
				}
				return e
			}
			if count == 2 && fans > 2 {
				if e := Csw_ObjTwoVarCut(p, c); e != aig.InvalidEdge {
					p.nNodesTriv2++
					return e
				}
			}
			start = util.Abc_Clock()
			equivalent := aig.InvalidEdge
			if len(c.pFanins) > 2 {
				equivalent = Csw_TableCutLookup(p, c)
			}
			p.timeHash += util.Abc_Clock() - start
			if equivalent != aig.InvalidEdge {
				p.nNodesCuts++
				return equivalent
			}
			c.Cost = Csw_CutFindCost(p, c)
		}
	}
	p.nNodesTried++
	start := util.Abc_Clock()
	for _, c := range Csw_ObjForEachCut(p, obj) {
		if len(c.pFanins) > 2 {
			Csw_TableCutInsert(p, c)
		}
	}
	p.timeHash += util.Abc_Clock() - start
	return obj
}
