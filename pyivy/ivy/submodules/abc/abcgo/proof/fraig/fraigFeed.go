package fraig

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/sat/msat"
	"math/bits"
)

func fraigBit(a []uint32, i int32) bool { return a[i/32]&(uint32(1)<<uint(i%32)) != 0 }
func fraigSetBit(a []uint32, i int32)   { a[i/32] |= uint32(1) << uint(i%32) }
func Fraig_FeedBackInit(p *Fraig_Man_t) {
	p.vCones = Fraig_NodeVecAlloc(500)
	p.vPatsReal = msat.Msat_IntVecAlloc(1000)
	p.pSimsReal = fraigMemFixedEntryWords(p.mmSims, p.nWordsDyna)
	p.pSimsTemp = fraigMemFixedEntryWords(p.mmSims, p.nWordsDyna)
	p.pSimsDiff = fraigMemFixedEntryWords(p.mmSims, p.nWordsDyna)
}
func Fraig_FeedBack(p *Fraig_Man_t, model []int, vars *msat.Msat_IntVec_t, old, node *Fraig_Node_t) {
	start := util.Abc_Clock()
	n := Fraig_FeedBackPrepare(p, model, vars)
	words := Fraig_FeedBackInsert(p, n)
	assert(p.iWordStart+words <= p.nWordsDyna, "feedback fits")
	for _, n := range p.vNodes.pArray[1:] {
		if Fraig_NodeIsAnd(n) != 0 {
			Fraig_NodeSimulate(n, p.iWordStart, p.iWordStart+words, 0)
		}
	}
	if p.fDoSparse != 0 {
		Fraig_TableRehashF0(p, 0)
	}
	if p.fChoicing == 0 {
		Fraig_FeedBackVerify(p, old, node)
	}
	if p.iWordStart+words == p.nWordsDyna {
		p.iWordStart = Fraig_FeedBackCompress(p)
	} else {
		p.iWordStart += words
	}
	p.timeFeed += util.Abc_Clock() - start
}
func Fraig_FeedBackPrepare(p *Fraig_Man_t, model []int, vars *msat.Msat_IntVec_t) int32 {
	for _, node := range p.vInputs.pArray {
		node.fFeedUse = 0
	}
	var count int32
	for _, index := range msat.Msat_IntVecReadArray(vars) {
		node := p.vNodes.pArray[index]
		if Fraig_NodeIsVar(node) == 0 {
			continue
		}
		node.fFeedUse = 1
		if model[index]&1 == 0 {
			node.fFeedVal = 1
		} else {
			node.fFeedVal = 0
		}
		count++
	}
	return count
}
func Fraig_FeedBackInsert(p *Fraig_Man_t, nVars int32) int32 {
	words := (nVars + 32) / 32
	if words > p.nWordsDyna-p.iWordStart {
		words = p.nWordsDyna - p.iWordStart
	}
	limit := words*32 - 2
	msat.Msat_IntVecPush(p.vPatsReal, int(p.iWordStart*32))
	fraigSetBit(p.pSimsReal, p.iWordStart*32)
	flip := int32(1)
	for _, node := range p.vInputs.pArray {
		for w := p.iWordStart; w < p.iWordStart+words; w++ {
			if node.fFeedUse == 0 {
				node.puSimD[w] = aig.Aig_ManRandom(0)
			} else if node.fFeedVal != 0 {
				node.puSimD[w] = ^uint32(0)
			} else {
				node.puSimD[w] = 0
			}
		}
		if node.fFeedUse != 0 && flip < limit {
			index := p.iWordStart*32 + flip
			node.puSimD[index/32] ^= uint32(1) << uint(index%32)
			flip++
		}
		node.fFeedUse = 0
		for w := p.iWordStart; w < p.iWordStart+words; w++ {
			node.uHashD ^= node.puSimD[w] * fraigSimulationPrime(w)
		}
	}
	return words
}
func Fraig_FeedBackVerify(p *Fraig_Man_t, old, node *Fraig_Node_t) {
	i := msat.Msat_IntVecReadEntry(p.vPatsReal, msat.Msat_IntVecReadSize(p.vPatsReal)-1)
	_, _ = fraigBit(old.puSimD, int32(i)), fraigBit(node.puSimD, int32(i))
}
func Fraig_FeedBackCompress(p *Fraig_Man_t) int32 {
	checks := p.nBTLimit == -1
	if checks {
		Fraig_FeedBackCheckTable(p)
		if p.fDoSparse != 0 {
			Fraig_FeedBackCheckTableF0(p)
		}
	}
	Fraig_FeedBackCovering(p, p.vPatsReal)
	pats := msat.Msat_IntVecReadArray(p.vPatsReal)
	count := int32(len(pats))
	p.iWordStart = (p.iPatsPerm + count + 31) / 32
	for _, node := range p.vInputs.pArray {
		sims := node.puSimD
		for w := p.iWordPerm; w < p.iWordStart; w++ {
			p.pSimsTemp[w] = 0
		}
		for i, pat := range pats {
			if fraigBit(sims, int32(pat)) {
				target := p.iPatsPerm + int32(i)
				if target < p.iWordPerm*32 {
					sims[target/32] |= uint32(1) << uint(target%32)
				} else {
					p.pSimsTemp[target/32] |= uint32(1) << uint(target%32)
				}
			}
		}
		for w := p.iWordPerm; w < p.iWordStart; w++ {
			sims[w] = p.pSimsTemp[w]
		}
		node.uHashD = 0
		for w := int32(0); w < p.iWordStart; w++ {
			node.uHashD ^= sims[w] * fraigSimulationPrime(w)
		}
	}
	p.iWordPerm = p.iWordStart
	p.iPatsPerm += count
	assert(p.iWordPerm == (p.iPatsPerm+31)/32, "permanent pattern count")
	for _, node := range p.vNodes.pArray[1:] {
		if Fraig_NodeIsAnd(node) != 0 {
			node.uHashD = 0
			Fraig_NodeSimulate(node, 0, p.iWordPerm, 0)
		}
	}
	if checks {
		Fraig_FeedBackCheckTable(p)
	}
	if p.fDoSparse != 0 {
		Fraig_TableRehashF0(p, 0)
		if checks {
			Fraig_FeedBackCheckTableF0(p)
		}
	}
	if p.iWordPerm+FRAIG_WORDS_STORE > p.nWordsDyna {
		Fraig_ReallocateSimulationInfo(p)
	}
	msat.Msat_IntVecClear(p.vPatsReal)
	clear(p.pSimsReal)
	for w := int32(0); w < p.iWordPerm; w++ {
		p.pSimsReal[w] = ^uint32(0)
	}
	if p.iPatsPerm%32 > 0 {
		p.pSimsReal[p.iWordPerm-1] = (uint32(1) << uint(p.iPatsPerm%32)) - 1
	}
	return p.iWordStart
}
func Fraig_FeedBackCovering(p *Fraig_Man_t, pats *msat.Msat_IntVec_t) {
	columns := Fraig_FeedBackCoveringStart(p)
	hits := make([]int32, len(columns))
	for i, col := range columns {
		hits[i] = Fraig_BitStringCountOnes(col, p.iWordStart)
	}
	for {
		i := Fraig_GetSmallestColumn(hits, int32(len(hits)))
		if i < 0 {
			break
		}
		pat := Fraig_GetHittingPattern(columns[i], p.iWordStart)
		Fraig_CancelCoveredColumns(columns, hits, pat)
		msat.Msat_IntVecPush(pats, int(pat))
	}
	for _, col := range columns {
		fraigMemFixedRecycleWords(p.mmSims, col)
	}
}
func Fraig_FeedBackCoveringStart(p *Fraig_Man_t) [][]uint32 {
	var columns [][]uint32
	collect := func(head *Fraig_Node_t, limit bool) {
		var nodes []*Fraig_Node_t
		for e := head; e != nil; e = e.pNextD {
			nodes = append(nodes, e)
		}
		if len(nodes) < 2 || limit && len(nodes) > 20 {
			return
		}
		for i, a := range nodes {
			for _, b := range nodes[i+1:] {
				if limit && Fraig_CompareSimInfoUnderMask(a, b, p.iWordStart, 0, p.pSimsReal) == 0 {
					continue
				}
				col := fraigMemFixedEntryWords(p.mmSims, p.nWordsRand+p.nWordsDyna)[:p.iWordStart]
				for w := int32(0); w < p.iWordStart; w++ {
					col[w] = (a.puSimD[w] ^ b.puSimD[w]) & ^p.pSimsReal[w]
				}
				columns = append(columns, col)
			}
		}
	}
	for _, head := range p.pTableF.pBins {
		for e := head; e != nil; e = e.pNextF {
			collect(e, true)
		}
	}
	if p.fDoSparse == 0 {
		return columns
	}
	for _, head := range p.pTableF0.pBins {
		for e := head; e != nil; e = e.pNextF {
			e.uHashD = 0
			for w := int32(0); w < p.iWordStart; w++ {
				e.uHashD ^= (e.puSimD[w] & p.pSimsReal[w]) * fraigSimulationPrime(w)
			}
		}
	}
	Fraig_TableRehashF0(p, 1)
	for _, head := range p.pTableF0.pBins {
		for e := head; e != nil; e = e.pNextF {
			collect(e, false)
		}
	}
	return columns
}
func Fraig_GetSmallestColumn(hits []int32, n int32) int32 {
	best, smallest := int32(-1), int32(1000000)
	for i, h := range hits[:n] {
		if h == 1 {
			return int32(i)
		}
		if h != 0 && h < smallest {
			best, smallest = int32(i), h
		}
	}
	return best
}
func Fraig_GetHittingPattern(sims []uint32, n int32) int32 {
	for i, w := range sims[:n] {
		if w != 0 {
			return int32(i*32 + bits.TrailingZeros32(w))
		}
	}
	return -1
}
func Fraig_CancelCoveredColumns(cols [][]uint32, hits []int32, pat int32) {
	for i, col := range cols {
		if fraigBit(col, pat) {
			hits[i] = 0
		}
	}
}
func Fraig_FeedBackCheckTable(p *Fraig_Man_t) {
	for _, head := range p.pTableF.pBins {
		for e := head; e != nil; e = e.pNextF {
			for a := e; a != nil; a = a.pNextD {
				for b := a.pNextD; b != nil; b = b.pNextD {
					if Fraig_CompareSimInfo(a, b, p.iWordStart, 0) != 0 {
						fmt.Printf("Nodes %d and %d have the same D simulation info.\n", a.Num, b.Num)
					}
				}
			}
		}
	}
}
func Fraig_FeedBackCheckTableF0(p *Fraig_Man_t) {
	for _, head := range p.pTableF0.pBins {
		for a := head; a != nil; a = a.pNextF {
			for b := a.pNextF; b != nil; b = b.pNextF {
				if Fraig_CompareSimInfo(a, b, p.iWordStart, 0) != 0 {
					fmt.Printf("Nodes %d and %d have the same D simulation info.\n", a.Num, b.Num)
				}
			}
		}
	}
}
func Fraig_ReallocateSimulationInfo(p *Fraig_Man_t) {
	p.nWordsDyna *= 2
	oldMem := p.mmSims
	p.mmSims = Fraig_MemFixedStart(4 * (p.nWordsRand + p.nWordsDyna))
	fraigNodeSimulationStorage(p, p.pConst1)
	for _, node := range p.vInputs.pArray {
		oldR, oldD := node.puSimR, node.puSimD
		fraigNodeSimulationStorage(p, node)
		copy(node.puSimR, oldR)
		copy(node.puSimD, oldD[:p.iWordStart])
	}
	Fraig_MemFixedStop(oldMem, 0)
	for _, node := range p.vNodes.pArray[1:] {
		if Fraig_NodeIsAnd(node) == 0 {
			continue
		}
		fraigNodeSimulationStorage(p, node)
		old := node.uHashR
		node.uHashR = 0
		Fraig_NodeSimulate(node, 0, p.nWordsRand, 1)
		assert(old == node.uHashR, "random signature preserved")
		old = node.uHashD
		node.uHashD = 0
		Fraig_NodeSimulate(node, 0, p.iWordStart, 0)
		assert(old == node.uHashD, "dynamic signature preserved")
	}
	p.pSimsReal = fraigMemFixedEntryWords(p.mmSims, p.nWordsDyna)
	p.pSimsTemp = fraigMemFixedEntryWords(p.mmSims, p.nWordsDyna)
	p.pSimsDiff = fraigMemFixedEntryWords(p.mmSims, p.nWordsDyna)
}
func Fraig_ManAllocCounterExample(p *Fraig_Man_t) []int32 {
	return make([]int32, len(p.vInputs.pArray))
}
func Fraig_ManSimulateBitNode_rec(p *Fraig_Man_t, node *Fraig_Node_t) int32 {
	if Fraig_NodeIsTravIdCurrent(p, node) != 0 {
		return int32(node.fMark3)
	}
	Fraig_NodeSetTravIdCurrent(p, node)
	a := Fraig_ManSimulateBitNode_rec(p, Fraig_Regular(node.p1))
	b := Fraig_ManSimulateBitNode_rec(p, Fraig_Regular(node.p2))
	if Fraig_IsComplement(node.p1) {
		a ^= 1
	}
	if Fraig_IsComplement(node.p2) {
		b ^= 1
	}
	node.fMark3 = uint32(a & b)
	return int32(node.fMark3)
}
func Fraig_ManSimulateBitNode(p *Fraig_Man_t, node *Fraig_Node_t, model []int32) int32 {
	Fraig_ManIncrementTravId(p)
	for i, n := range p.vInputs.pArray {
		Fraig_NodeSetTravIdCurrent(p, n)
		n.fMark3 = uint32(model[i]) & 1
	}
	n := Fraig_ManSimulateBitNode_rec(p, Fraig_Regular(node))
	if Fraig_IsComplement(node) {
		n ^= 1
	}
	return n
}
func Fraig_ManSaveCounterExample(p *Fraig_Man_t, node *Fraig_Node_t) []int32 {
	model := Fraig_ManAllocCounterExample(p)
	compl := fraigBool(!Fraig_IsComplement(node))
	pattern := Fraig_FindFirstDiff(p.pConst1, Fraig_Regular(node), compl, p.nWordsRand, 1)
	random := true
	if pattern < 0 {
		pattern = Fraig_FindFirstDiff(p.pConst1, Fraig_Regular(node), compl, p.iWordStart, 0)
		random = false
	}
	if pattern < 0 {
		return nil
	}
	for i, n := range p.vInputs.pArray {
		words := n.puSimD
		if random {
			words = n.puSimR
		}
		if fraigBit(words, pattern) {
			model[i] = 1
		}
	}
	assert(Fraig_ManSimulateBitNode(p, node, model) != 0, "counterexample satisfies node")
	return model
}
