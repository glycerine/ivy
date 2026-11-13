package rwr

// Rewrite-library construction from src/opt/rwr/rwrLib.c.
import "fmt"

func rwrAddTail(head **Rwr_Node_t, node *Rwr_Node_t) {
	for *head != nil {
		head = &(*head).pNext
	}
	*head = node
}
func rwrNodeTruth(a, b Rwr_NodeRef, xor bool) uint16 {
	x, y := a.node.uTruth, b.node.uTruth
	if xor {
		return x ^ y
	}
	if a.complemented {
		x = ^x
	}
	if b.complemented {
		y = ^y
	}
	return x & y
}
func Rwr_ManAddVar(p *Rwr_Man_t, truth uint32, precompute int32) *Rwr_Node_t {
	node := &Rwr_Node_t{Id: int32(len(p.vForest)), uTruth: uint16(truth), fUsed: true}
	p.vForest = append(p.vForest, node)
	if precompute != 0 {
		rwrAddTail(&p.pTable[node.uTruth], node)
	}
	return node
}
func Rwr_ManAddNode(p *Rwr_Man_t, a, b Rwr_NodeRef, fExor, level, volume int32) *Rwr_Node_t {
	p.nConsidered++
	node := &Rwr_Node_t{Id: int32(len(p.vForest)), uTruth: rwrNodeTruth(a, b, fExor != 0), Level: uint32(level) & 63, Volume: uint32(volume) & 255, fExor: fExor != 0, p0: a, p1: b}
	p.vForest = append(p.vForest, node)
	if node.uTruth != p.puCanons[node.uTruth] {
		return node
	}
	p.nAdded++
	if p.pTable[node.uTruth] == nil {
		p.nClasses++
	}
	rwrAddTail(&p.pTable[node.uTruth], node)
	return node
}
func Rwr_ManTryNode(p *Rwr_Man_t, a, b Rwr_NodeRef, fExor, level, volume int32) *Rwr_Node_t {
	p.nConsidered++
	truth := rwrNodeTruth(a, b, fExor != 0)
	if level > 2 && !p.pPractical[p.puCanons[truth]] {
		return nil
	}
	place := &p.pTable[truth]
	for *place != nil {
		old := *place
		if old.Level <= uint32(level) && old.Volume < uint32(volume) {
			return nil
		}
		place = &old.pNext
	}
	if p.pTable[truth] == nil && p.puCanons[truth] == truth {
		p.nClasses++
	}
	node := &Rwr_Node_t{Id: int32(len(p.vForest)), uTruth: truth, Level: uint32(level) & 63, Volume: uint32(volume) & 255, fExor: fExor != 0, p0: a, p1: b}
	p.vForest = append(p.vForest, node)
	*place = node
	return node
}
func Rwr_MarkUsed_rec(p *Rwr_Man_t, node *Rwr_Node_t) {
	if node.fUsed || node.TravId == p.nTravIds {
		return
	}
	node.TravId = p.nTravIds
	node.fUsed = true
	Rwr_MarkUsed_rec(p, node.p0.node)
	Rwr_MarkUsed_rec(p, node.p1.node)
}
func Rwr_Trav_rec(p *Rwr_Man_t, node *Rwr_Node_t, volume *int32) {
	if node.fUsed || node.TravId == p.nTravIds {
		return
	}
	node.TravId = p.nTravIds
	*volume++
	if node.fExor {
		*volume++
	}
	Rwr_Trav_rec(p, node.p0.node, volume)
	Rwr_Trav_rec(p, node.p1.node, volume)
}
func Rwr_ManNodeVolume(p *Rwr_Man_t, a, b *Rwr_Node_t) int32 {
	var volume int32
	Rwr_ManIncTravId(p)
	Rwr_Trav_rec(p, a, &volume)
	Rwr_Trav_rec(p, b, &volume)
	return volume
}
func Rwr_ManPrecompute(p *Rwr_Man_t) {
	oldLevel := int32(-1)
search:
	for i := 1; i < len(p.vForest); i++ {
		for k := 1; k < len(p.vForest); k++ {
			a, b := p.vForest[i], p.vForest[k]
			if oldLevel < int32(a.Level) {
				oldLevel = int32(a.Level)
				fmt.Printf("Starting level %d  (at %d nodes).\n", oldLevel+1, i)
				fmt.Printf("Considered = %5d M.   Found = %8d.   Classes = %6d.   Trying %7d.\n", p.nConsidered/1000000, len(p.vForest), p.nClasses, i)
			}
			if k == i || a.Level+b.Level > 5 {
				break
			}
			level := int32(1 + max(a.Level, b.Level))
			volume := 1 + Rwr_ManNodeVolume(p, a, b)
			left, right := Rwr_NodeRef{node: a}, Rwr_NodeRef{node: b}
			Rwr_ManTryNode(p, left, right, 0, level, volume)
			Rwr_ManTryNode(p, Rwr_Not(left), right, 0, level, volume)
			Rwr_ManTryNode(p, left, Rwr_Not(right), 0, level, volume)
			Rwr_ManTryNode(p, Rwr_Not(left), Rwr_Not(right), 0, level, volume)
			Rwr_ManTryNode(p, left, right, 1, level, volume+1)
			if p.nConsidered%50000000 == 0 {
				fmt.Printf("Considered = %5d M.   Found = %8d.   Classes = %6d.   Trying %7d.\n", p.nConsidered/1000000, len(p.vForest), p.nClasses, i)
			}
			if len(p.vForest) == RWR_LIMIT+5 {
				break search
			}
		}
	}
	Rwr_ManIncTravId(p)
	canonical := 0
	for _, node := range p.vForest[5:] {
		if node.uTruth == p.puCanons[node.uTruth] {
			Rwr_MarkUsed_rec(p, node)
			canonical++
		}
	}
	kept := p.vForest[:5]
	for _, node := range p.vForest[5:] {
		if node.fUsed {
			node.Id = int32(len(kept))
			kept = append(kept, node)
		}
	}
	p.vForest = kept
	fmt.Printf("Total canonical = %4d. Total used = %5d.\n", canonical, len(kept))
}
