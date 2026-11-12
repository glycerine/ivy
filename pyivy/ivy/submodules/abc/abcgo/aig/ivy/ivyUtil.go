package ivy

import (
	"fmt"
)

func ivyManCollectCutRec(p *Ivy_Man_t, node *Ivy_Obj_t, nodes *Vec_Int_t, oldMarks map[*Ivy_Obj_t]int32) {
	node = Ivy_Regular(node)
	if node == nil || node.owner != p {
		panic("ivy: cut collection reached an invalid node")
	}
	if node.fMarkA != 0 {
		return
	}
	oldMarks[node] = node.fMarkA
	node.fMarkA = 1
	assert(Ivy_ObjIsNode(node) != 0, "cut interiors must be AND/XOR nodes")
	ivyManCollectCutRec(p, Ivy_ObjFanin0(node), nodes, oldMarks)
	ivyManCollectCutRec(p, Ivy_ObjFanin1(node), nodes, oldMarks)
	Vec_IntPush(nodes, node.Id)
}

// Ivy_ManCollectCut returns leaf IDs followed by the topologically ordered
// internal nodes. It restores every mark even when an invalid cut is rejected.
func Ivy_ManCollectCut(p *Ivy_Man_t, root *Ivy_Obj_t, leaves, nodes *Vec_Int_t) {
	if p == nil || nodes == nil || leaves == nil {
		panic("ivy: nil cut argument")
	}
	Vec_IntClear(nodes)
	marked := make(map[*Ivy_Obj_t]int32, Vec_IntSize(leaves))
	for i := int32(0); i < Vec_IntSize(leaves); i++ {
		id := Vec_IntEntry(leaves, i)
		leaf := Ivy_ManObj(p, id)
		if _, seen := marked[leaf]; !seen {
			marked[leaf] = leaf.fMarkA
		}
		Vec_IntPush(nodes, id)
		leaf.fMarkA = 1
	}
	defer func() {
		for node, mark := range marked {
			node.fMarkA = mark
		}
	}()
	ivyManCollectCutRec(p, root, nodes, marked)
}

func ivyTruthStore(objNum int32, truth []uint32) []uint32 {
	start := int(objNum) * 8
	if objNum < 0 || start+8 > len(truth) {
		panic("ivy: truth table slot out of bounds")
	}
	return truth[start : start+8]
}

func Ivy_ManCutTruthOne(_ *Ivy_Man_t, node *Ivy_Obj_t, truth []uint32) {
	out := ivyTruthStore(node.TravId, truth)
	a := ivyTruthStore(Ivy_ObjFanin0(node).TravId, truth)
	b := ivyTruthStore(Ivy_ObjFanin1(node).TravId, truth)
	for i := range out {
		switch {
		case Ivy_ObjIsExor(node) != 0:
			out[i] = a[i] ^ b[i]
		case Ivy_ObjFaninC0(node) == 0 && Ivy_ObjFaninC1(node) == 0:
			out[i] = a[i] & b[i]
		case Ivy_ObjFaninC0(node) == 0:
			out[i] = a[i] &^ b[i]
		case Ivy_ObjFaninC1(node) == 0:
			out[i] = ^a[i] & b[i]
		default:
			out[i] = ^a[i] &^ b[i]
		}
	}
}

var ivyElementaryTruth = [8][8]uint32{
	{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
	{0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC},
	{0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0},
	{0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00},
	{0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000, 0xFFFF0000},
	{0x00000000, 0xFFFFFFFF, 0x00000000, 0xFFFFFFFF, 0x00000000, 0xFFFFFFFF, 0x00000000, 0xFFFFFFFF},
	{0x00000000, 0x00000000, 0xFFFFFFFF, 0xFFFFFFFF, 0x00000000, 0x00000000, 0xFFFFFFFF, 0xFFFFFFFF},
	{0x00000000, 0x00000000, 0x00000000, 0x00000000, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
}

// Ivy_ManCutTruth computes the truth of root over leaves in the order given.
// The returned word slice is owned by this call and is safe to retain.
func Ivy_ManCutTruth(p *Ivy_Man_t, root *Ivy_Obj_t, leaves *Vec_Int_t, nodes *Vec_Int_t, scratch *Vec_Int_t) []uint32 {
	if Vec_IntSize(leaves) > int32(len(ivyElementaryTruth)) {
		panic("ivy: cut truth supports at most eight leaves")
	}
	Ivy_ManCollectCut(p, root, leaves, nodes)
	for i := int32(0); i < Vec_IntSize(nodes); i++ {
		Ivy_ManObj(p, Vec_IntEntry(nodes, i)).TravId = i
	}
	truth := make([]uint32, 8*Vec_IntSize(nodes))
	for i := int32(0); i < Vec_IntSize(leaves); i++ {
		copy(ivyTruthStore(i, truth), ivyElementaryTruth[i][:])
	}
	for i := Vec_IntSize(leaves); i < Vec_IntSize(nodes); i++ {
		Ivy_ManCutTruthOne(p, Ivy_ManObj(p, Vec_IntEntry(nodes, i)), truth)
	}
	if scratch != nil {
		Vec_IntClear(scratch)
		Vec_IntGrow(scratch, int32(len(truth)))
		for _, word := range truth {
			Vec_IntPush(scratch, int32(word))
		}
	}
	rootIndex := Ivy_Regular(root).TravId
	return ivyTruthStore(rootIndex, truth)
}

func Ivy_ManLatches(p *Ivy_Man_t) *Vec_Int_t {
	latches := Vec_IntAlloc(Ivy_ManLatchNum(p))
	for _, node := range p.vObjs {
		if node != nil && Ivy_ObjIsLatch(node) != 0 {
			Vec_IntPush(latches, node.Id)
		}
	}
	return latches
}

// Ivy_ObjRefDeref adjusts the reference counts of the MFFC rooted at node and
// returns the number of internal logic nodes. fLabel stamps visited nodes.
func Ivy_ObjRefDeref(p *Ivy_Man_t, node *Ivy_Obj_t, reference, label int32) int32 {
	node = Ivy_Regular(node)
	if label != 0 {
		Ivy_ObjSetTravIdCurrent(p, node)
	}
	if Ivy_ObjIsPi(node) != 0 {
		return 0
	}
	assert(Ivy_ObjIsNode(node) != 0 || Ivy_ObjIsBuf(node) != 0 || Ivy_ObjIsLatch(node) != 0,
		"reference walk expects an internal node")
	fanins := [2]*Ivy_Obj_t{Ivy_ObjFanin0(node), Ivy_ObjFanin1(node)}
	count := Ivy_ObjIsNode(node)
	for _, fanin := range fanins {
		if fanin == nil {
			continue
		}
		if reference != 0 {
			if fanin.nRefs == 0 {
				count += Ivy_ObjRefDeref(p, fanin, reference, label)
			}
			fanin.nRefs++
			continue
		}
		assert(fanin.nRefs > 0, "fanin.nRefs > 0")
		fanin.nRefs--
		if fanin.nRefs == 0 {
			count += Ivy_ObjRefDeref(p, fanin, reference, label)
		}
	}
	return count
}

func Ivy_ObjMffcLabel(p *Ivy_Man_t, node *Ivy_Obj_t) int32 {
	assert(Ivy_IsComplement(node) == 0 && Ivy_ObjIsNode(node) != 0, "regular logic node required")
	removed := Ivy_ObjRefDeref(p, node, 0, 1)
	restored := Ivy_ObjRefDeref(p, node, 1, 0)
	assert(removed == restored && removed > 0, "MFFC dereference/reference counts must match")
	return removed
}

func Ivy_ObjPrintVerbose(p *Ivy_Man_t, node *Ivy_Obj_t, haig int32) {
	assert(Ivy_IsComplement(node) == 0, "regular node required")
	fmt.Printf("Node %5d : ", Ivy_ObjId(node))
	switch {
	case Ivy_ObjIsConst1(node) != 0:
		fmt.Print("constant 1")
	case Ivy_ObjIsPi(node) != 0:
		fmt.Print("PI")
	case Ivy_ObjIsPo(node) != 0:
		fmt.Print("PO")
	case Ivy_ObjIsLatch(node) != 0:
		fmt.Printf("latch (%d%s)", Ivy_ObjFaninId0(node), ivyPolarity(Ivy_ObjFaninC0(node)))
	case Ivy_ObjIsBuf(node) != 0:
		fmt.Printf("buffer (%d%s)", Ivy_ObjFaninId0(node), ivyPolarity(Ivy_ObjFaninC0(node)))
	case Ivy_ObjIsNode(node) != 0:
		fmt.Printf("%s( %5d%s, %5d%s )", ivyLogicName(node), Ivy_ObjFaninId0(node), ivyPolarity(Ivy_ObjFaninC0(node)), Ivy_ObjFaninId1(node), ivyPolarity(Ivy_ObjFaninC1(node)))
	default:
		fmt.Printf("type %d", node.Type)
	}
	fmt.Printf(" (refs = %3d)", Ivy_ObjRefs(node))
	if p != nil && p.fFanout != 0 {
		var fanouts []*Ivy_Obj_t
		Ivy_ObjCollectFanouts(p, node, &fanouts)
		if len(fanouts) > 0 {
			fmt.Print("\nFanouts:")
			for _, fanout := range fanouts {
				fmt.Printf("\n    Node %5d : %s", Ivy_ObjId(fanout), ivyLogicName(fanout))
			}
		}
	}
	if haig == 0 {
		if node.pEquiv == nil {
			fmt.Print(" HAIG node not given")
		} else {
			fmt.Printf(" HAIG node = %d%s", Ivy_ObjId(node.pEquiv), ivyPolarity(Ivy_IsComplement(node.pEquiv)))
		}
		return
	}
	if node.pEquiv == nil {
		return
	}
	if Ivy_ObjRefs(node) > 0 {
		fmt.Printf("  { %5d", node.Id)
		seen := map[*Ivy_Obj_t]bool{node: true}
		for edge := node.pEquiv; edge != node; edge = Ivy_Regular(edge.pEquiv) {
			member := Ivy_Regular(edge)
			if member == nil || seen[member] {
				break
			}
			seen[member] = true
			fmt.Printf(" %5d%s", member.Id, ivyPolarity(Ivy_IsComplement(member.pEquiv)))
		}
		fmt.Print(" }")
		return
	}
	for candidate := Ivy_Regular(node.pEquiv); candidate != nil; candidate = Ivy_Regular(candidate.pEquiv) {
		if Ivy_ObjRefs(candidate) > 0 {
			fmt.Printf("  class of %d", candidate.Id)
			return
		}
	}
	fmt.Print("  choice class has no referenced representative")
}

func ivyLogicName(node *Ivy_Obj_t) string {
	switch {
	case Ivy_ObjIsPi(node) != 0:
		return "PI"
	case Ivy_ObjIsPo(node) != 0:
		return "PO"
	case Ivy_ObjIsLatch(node) != 0:
		return "latch"
	case Ivy_ObjIsBuf(node) != 0:
		return "buffer"
	case Ivy_ObjIsExor(node) != 0:
		return "EXOR"
	case Ivy_ObjIsAnd(node) != 0:
		return "AND"
	default:
		return fmt.Sprintf("type %d", Ivy_ObjType(node))
	}
}

func ivyPolarity(complement int32) string {
	if complement != 0 {
		return "'"
	}
	return " "
}

func Ivy_ManPrintVerbose(p *Ivy_Man_t, haig int32) {
	fmt.Print("PIs:")
	for _, node := range p.vPis {
		fmt.Printf(" %d", node.Id)
	}
	fmt.Print("\nPOs:")
	for _, node := range p.vPos {
		fmt.Printf(" %d", node.Id)
	}
	fmt.Print("\nLatches:")
	for _, node := range p.vObjs {
		if node != nil && Ivy_ObjIsLatch(node) != 0 {
			fmt.Printf(" %d=%d%s", node.Id, Ivy_ObjFaninId0(node), ivyPolarity(Ivy_ObjFaninC0(node)))
		}
	}
	fmt.Println()
	for _, id := range Ivy_ManDfsSeq(p, nil) {
		Ivy_ObjPrintVerbose(p, Ivy_ManObj(p, id), haig)
		fmt.Println()
	}
	fmt.Println()
}

func Ivy_CutTruthPrint2(_ *Ivy_Man_t, cut *Ivy_Cut_t, truth uint32) int32 {
	fmt.Print("Trying cut : {")
	for i := int32(0); i < int32(cut.nSize); i++ {
		fmt.Printf(" %6d(%d)", Ivy_LeafId(cut.pArray[i]), Ivy_LeafLat(cut.pArray[i]))
	}
	fmt.Printf(" }   %016b\n", truth)
	return 0
}

func Ivy_CutTruthPrint(p *Ivy_Man_t, cut *Ivy_Cut_t, _ uint32) int32 {
	var latches, present int32
	for i := int32(0); i < int32(cut.nSize); i++ {
		latches += Ivy_LeafLat(cut.pArray[i])
	}
	if p.fFanout != 0 {
		for i := int32(0); i < int32(cut.nSize); i++ {
			obj := Ivy_ManObj(p, Ivy_LeafId(cut.pArray[i]))
			var fanouts []*Ivy_Obj_t
			Ivy_ObjCollectFanouts(p, obj, &fanouts)
			for _, fanout := range fanouts {
				if Ivy_ObjIsLatch(fanout) != 0 {
					present++
					break
				}
			}
		}
	}
	missing := latches - present
	if missing > 0 {
		return 2 * missing
	}
	return 0
}
