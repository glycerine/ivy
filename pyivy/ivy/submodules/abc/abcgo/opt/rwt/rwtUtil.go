package rwt

import (
	"fmt"
	"time"
)

func Rwt_ListAddToTail(list **Rwt_Node_t, node *Rwt_Node_t) {
	if node == nil {
		return
	}
	for *list != nil {
		list = &(*list).pNext
	}
	*list = node
}

func rwtNewNode(p *Rwt_Man_t) *Rwt_Node_t {
	node := &Rwt_Node_t{Id: int32(len(p.vForest))}
	neg := &Rwt_Node_t{Id: node.Id, regular: node, fComplement: true}
	node.neg = neg
	neg.neg = node
	p.vForest = append(p.vForest, node)
	return node
}

func Rwt_ManAddVar(p *Rwt_Man_t, truth uint32, precompute int32) *Rwt_Node_t {
	node := rwtNewNode(p)
	node.uTruth = truth & 0xffff
	node.fUsed = true
	if precompute != 0 {
		Rwt_ListAddToTail(&p.pTable[node.uTruth], node)
	}
	return node
}

func Rwt_ManAddNode(p *Rwt_Man_t, p0, p1 *Rwt_Node_t, exor int32, level, volume int32) *Rwt_Node_t {
	if p == nil || p0 == nil || p1 == nil {
		return nil
	}
	p.nConsidered++
	var truth uint32
	if exor != 0 {
		truth = Rwt_Regular(p0).uTruth ^ Rwt_Regular(p1).uTruth
	} else {
		truth0, truth1 := Rwt_Regular(p0).uTruth, Rwt_Regular(p1).uTruth
		if Rwt_IsComplement(p0) != 0 {
			truth0 = ^truth0
		}
		if Rwt_IsComplement(p1) != 0 {
			truth1 = ^truth1
		}
		truth = truth0 & truth1 & 0xffff
	}
	node := rwtNewNode(p)
	node.uTruth = truth
	node.Level = uint32(level) & 0x3f
	node.Volume = uint32(volume) & 0xff
	node.fExor = exor != 0
	node.p0, node.p1 = p0, p1
	if truth != uint32(p.puCanons[truth]) {
		return node
	}
	p.nAdded++
	if p.pTable[truth] == nil {
		p.nClasses++
	}
	Rwt_ListAddToTail(&p.pTable[truth], node)
	return node
}

func Rwt_Trav_rec(p *Rwt_Man_t, node *Rwt_Node_t, volume *int32) {
	node = Rwt_Regular(node)
	if node == nil || node.fUsed || node.TravId == p.nTravIds {
		return
	}
	node.TravId = p.nTravIds
	*volume++
	if node.fExor {
		*volume++
	}
	Rwt_Trav_rec(p, node.p0, volume)
	Rwt_Trav_rec(p, node.p1, volume)
}

func Rwt_ManIncTravId(p *Rwt_Man_t) {
	if p.nTravIds >= 0 && uint32(p.nTravIds) < 0x8fffffff {
		p.nTravIds++
		return
	}
	for _, node := range p.vForest {
		node.TravId = 0
	}
	p.nTravIds = 1
}

func Rwt_ManNodeVolume(p *Rwt_Man_t, p0, p1 *Rwt_Node_t) int32 {
	volume := int32(0)
	Rwt_ManIncTravId(p)
	Rwt_Trav_rec(p, p0, &volume)
	Rwt_Trav_rec(p, p1, &volume)
	return volume
}

func Rwt_ManLoadFromArray(p *Rwt_Man_t, verbose int32) {
	start := time.Now()
	entries := 0
	for i := 0; i+1 < len(rwtSubgraphs); i += 2 {
		entry0, entry1 := uint32(rwtSubgraphs[i]), uint32(rwtSubgraphs[i+1])
		if entry0 == 0 && entry1 == 0 {
			break
		}
		exor := int32(entry0 & 1)
		entry0 >>= 1
		idx0, idx1 := int(entry0>>1), int(entry1>>1)
		if idx0 >= len(p.vForest) || idx1 >= len(p.vForest) {
			panic("invalid RWT subgraph table reference")
		}
		p0, p1 := p.vForest[idx0], p.vForest[idx1]
		level := 1 + max(int32(p0.Level), int32(p1.Level))
		volume := 1 + Rwt_ManNodeVolume(p, p0, p1)
		p0 = Rwt_NotCond(p0, int32(entry0&1))
		p1 = Rwt_NotCond(p1, int32(entry1&1))
		Rwt_ManAddNode(p, p0, p1, exor, level, volume+exor)
		entries++
	}
	if verbose != 0 {
		fmt.Printf("The number of classes = %d. Canonical nodes = %d.\n", p.nClasses, p.nAdded)
		fmt.Printf("The number of nodes loaded = %d. Loading = %s\n", entries, time.Since(start))
	}
}

func Rwt_ManGetPractical(p *Rwt_Man_t) []byte {
	practical := make([]byte, p.nFuncs)
	practical[0] = 1
	for _, class := range rwtPracticalClasses[1:] {
		if class == 0 {
			break
		}
		practical[class] = 1
	}
	return practical
}
