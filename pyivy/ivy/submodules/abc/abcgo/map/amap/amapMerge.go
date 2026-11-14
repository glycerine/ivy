package amap

import (
	"fmt"
	"time"
)

// Cuts are Go-owned values linked from each Amap node. Temporary cuts are
// bucketed by their resulting library node/literal in ppCutsTemp.
func Amap_ManSetupPis(p *Amap_Man_t) {
	for _, item := range Vec_PtrArray(p.vPis) {
		obj, ok := item.(*Amap_Obj_t)
		if !ok || obj == nil {
			continue
		}
		cut := &Amap_Cut_t{nFans: 1, Fans: []int32{Abc_Var2Lit(int32(obj.Id), 0)}}
		obj.pData, obj.nCuts, obj.EstRefs = cut, 1, 1
	}
}

func Amap_ManCutStore(p *Amap_Man_t, cut *Amap_Cut_t, complement int32) *Amap_Cut_t {
	if cut == nil {
		return nil
	}
	copyCut := &Amap_Cut_t{iMat: cut.iMat, fInv: cut.fInv ^ uint32(complement), nFans: cut.nFans, Fans: append([]int32(nil), cut.Fans...)}
	key := Abc_Var2Lit(int32(copyCut.iMat), int32(copyCut.fInv))
	if key < 0 || int(key) >= len(p.ppCutsTemp) {
		return nil
	}
	if p.ppCutsTemp[key] == nil {
		Vec_IntPushOrder(p.vTemp, key)
	}
	copyCut.pNext = p.ppCutsTemp[key]
	p.ppCutsTemp[key] = copyCut
	return copyCut
}

func Amap_ManCutCreate(p *Amap_Man_t, a, b *Amap_Cut_t, result int32) *Amap_Cut_t {
	return amapStoreCombinedCut(p, []*Amap_Cut_t{a, b}, result)
}

func Amap_ManCutCreate3(p *Amap_Man_t, a, b, c *Amap_Cut_t, result int32) *Amap_Cut_t {
	return amapStoreCombinedCut(p, []*Amap_Cut_t{a, b, c}, result)
}

func amapStoreCombinedCut(p *Amap_Man_t, parts []*Amap_Cut_t, result int32) *Amap_Cut_t {
	cut := &Amap_Cut_t{iMat: uint32(result)}
	for _, part := range parts {
		if part == nil {
			return nil
		}
		cut.Fans = append(cut.Fans, part.Fans...)
	}
	cut.nFans = uint32(len(cut.Fans))
	if result < 0 || int(result) >= len(p.ppCutsTemp) {
		return nil
	}
	if p.ppCutsTemp[result] == nil {
		Vec_IntPushOrder(p.vTemp, result)
	}
	cut.pNext = p.ppCutsTemp[result]
	p.ppCutsTemp[result] = cut
	return cut
}

func Amap_ManCutSaveStored(p *Amap_Man_t, node *Amap_Obj_t) {
	if node == nil {
		return
	}
	limit := int(p.pPars.NCutsMax) - 1
	if limit < 1 {
		limit = 1
	}
	cuts := []*Amap_Cut_t{{nFans: 1, Fans: []int32{Abc_Var2Lit(int32(node.Id), 0)}}}
	for _, key := range Vec_IntArray(p.vTemp) {
		if key < 0 || int(key) >= len(p.ppCutsTemp) {
			continue
		}
		for cut := p.ppCutsTemp[key]; cut != nil; cut = cut.pNext {
			cuts = append(cuts, &Amap_Cut_t{iMat: cut.iMat, fInv: cut.fInv, nFans: cut.nFans, Fans: append([]int32(nil), cut.Fans...)})
		}
		p.ppCutsTemp[key] = nil
	}
	if len(cuts) > 1 {
		// Preserve the source's stable key ordering: the trivial cut first,
		// then buckets in increasing result-literal order.
		for i := 1; i < len(cuts); i++ {
			for j := i; j > 1 && cuts[j].iMat < cuts[j-1].iMat; j-- {
				cuts[j], cuts[j-1] = cuts[j-1], cuts[j]
			}
		}
	}
	if len(cuts) > limit {
		cuts = cuts[:limit]
	}
	for i := range cuts {
		cuts[i].pNext = nil
		if i+1 < len(cuts) {
			cuts[i].pNext = cuts[i+1]
		}
	}
	node.pData, node.nCuts = cuts[0], uint32(len(cuts))
	p.nBytesUsed += 8 // two-word trivial cut header and fanin
	for _, cut := range cuts[1:] {
		p.nBytesUsed += int32(4 * (len(cut.Fans) + 1))
	}
	Vec_IntClear(p.vTemp)
}

func Amap_ManMergeCountCuts(p *Amap_Man_t, node *Amap_Obj_t) int32 {
	left, right := Amap_ObjFanin0(p, node), Amap_ObjFanin1(p, node)
	count := int32(1)
	for a := Amap_CutHead(left); a != nil; a = a.pNext {
		litA := amapCutLiteral(a, Amap_ObjFaninC0(node))
		for b := Amap_CutHead(right); b != nil; b = b.pNext {
			litB := amapCutLiteral(b, Amap_ObjFaninC1(node))
			if Amap_LibFindNode(p.pLib, litA, litB, amapBool(node.Type == AMAP_OBJ_XOR)) >= 0 {
				count++
			}
		}
	}
	return count
}

func Amap_ManPrintCuts(node *Amap_Obj_t) {
	if node == nil {
		return
	}
	fmt.Printf("NODE %5d : Type = %d  Cuts = %d\n", node.Id, node.Type, node.nCuts)
	for i, cut := 0, Amap_CutHead(node); cut != nil; i, cut = i+1, cut.pNext {
		fmt.Printf("%3d : Mat=%3d Inv=%d ", i, cut.iMat, cut.fInv)
		for _, fan := range cut.Fans {
			fmt.Printf("%d%c ", Abc_Lit2Var(fan), polarity(Abc_LitIsCompl(fan)))
		}
		fmt.Println()
	}
}

func Amap_ManMergeNodeChoice(p *Amap_Man_t, node *Amap_Obj_t) {
	for choice := node; choice != nil; choice = Amap_ObjChoice(p, choice) {
		for cut := Amap_CutHead(choice); cut != nil; cut = cut.pNext {
			if cut.iMat != 0 {
				Amap_ManCutStore(p, cut, int32(node.fPhase^choice.fPhase))
			}
		}
		choice.pData = nil
	}
	Amap_ManCutSaveStored(p, node)
}

func Amap_ManFindCut(_ *Amap_Obj_t, fanin *Amap_Obj_t, complement, value int32, out *Vec_Ptr_t) int32 {
	Vec_PtrClear(out)
	for cut := Amap_CutHead(fanin); cut != nil; cut = cut.pNext {
		if amapCutLiteral(cut, complement) == value {
			Vec_PtrPush(out, cut)
		}
	}
	return amapBool(Vec_PtrSize(out) == 0)
}

func Amap_ManMergeNodeCutsMux(p *Amap_Man_t, node *Amap_Obj_t) {
	left, middle, right := Amap_ObjFanin0(p, node), Amap_ObjFanin1(p, node), Amap_ObjFanin2(p, node)
	comps := [3]int32{Amap_ObjFaninC0(node), Amap_ObjFaninC1(node), Amap_ObjFaninC2(node)}
	for i := int32(0); i+3 < Vec_IntSize(p.pLib.vRules3); i += 4 {
		vals := [3]int32{Vec_IntEntry(p.pLib.vRules3, i), Vec_IntEntry(p.pLib.vRules3, i+1), Vec_IntEntry(p.pLib.vRules3, i+2)}
		if Amap_ManFindCut(node, left, comps[0], vals[0], p.vCuts0) != 0 || Amap_ManFindCut(node, middle, comps[1], vals[1], p.vCuts1) != 0 || Amap_ManFindCut(node, right, comps[2], vals[2], p.vCuts2) != 0 {
			continue
		}
		for _, ca := range Vec_PtrArray(p.vCuts0) {
			for _, cb := range Vec_PtrArray(p.vCuts1) {
				for _, cc := range Vec_PtrArray(p.vCuts2) {
					a, _ := ca.(*Amap_Cut_t)
					b, _ := cb.(*Amap_Cut_t)
					c, _ := cc.(*Amap_Cut_t)
					Amap_ManCutCreate3(p, amapComplementSingleton(a, comps[0]), amapComplementSingleton(b, comps[1]), amapComplementSingleton(c, comps[2]), Vec_IntEntry(p.pLib.vRules3, i+3))
				}
			}
		}
	}
	Amap_ManCutSaveStored(p, node)
	p.nCutsUsed += int32(node.nCuts)
	p.nCutsTried3 += int32(left.nCuts * middle.nCuts * right.nCuts)
}

func Amap_ManMergeNodeCuts(p *Amap_Man_t, node *Amap_Obj_t) {
	left, right := Amap_ObjFanin0(p, node), Amap_ObjFanin1(p, node)
	if node.Type == AMAP_OBJ_MUX {
		Amap_ManMergeNodeCutsMux(p, node)
		return
	}
	rules := p.pLib.pRules
	if node.Type == AMAP_OBJ_XOR {
		rules = p.pLib.pRulesX
	}
	if node.pData != nil {
		panic("amap: cuts already attached while merging node")
	}
	for ca := Amap_CutHead(left); ca != nil; ca = ca.pNext {
		litA := amapCutLiteral(ca, Amap_ObjFaninC0(node))
		if int(litA) >= len(rules) || litA < 0 {
			continue
		}
		resultByFan := make(map[int32]int32)
		for _, packed := range rules[litA] {
			if packed == 0 {
				break
			}
			resultByFan[packed&0xffff] = packed >> 16
		}
		for cb := Amap_CutHead(right); cb != nil; cb = cb.pNext {
			litB := amapCutLiteral(cb, Amap_ObjFaninC1(node))
			result, ok := resultByFan[litB]
			if !ok {
				continue
			}
			first, second := amapComplementSingleton(ca, Amap_ObjFaninC0(node)), amapComplementSingleton(cb, Amap_ObjFaninC1(node))
			if litA >= litB {
				Amap_ManCutCreate(p, first, second, result)
			} else {
				Amap_ManCutCreate(p, second, first, result)
			}
		}
	}
	Amap_ManCutSaveStored(p, node)
	p.nCutsUsed += int32(node.nCuts)
	p.nCutsTried += int32(left.nCuts * right.nCuts)
	if node.fRepr != 0 {
		Amap_ManMergeNodeChoice(p, node)
	}
}

func Amap_ManMerge(p *Amap_Man_t) {
	started := time.Now()
	Amap_ManSetupPis(p)
	for _, item := range Vec_PtrArray(p.vObjs) {
		node, ok := item.(*Amap_Obj_t)
		if ok && node != nil && Amap_ObjIsNode(node) != 0 {
			Amap_ManMergeNodeCuts(p, node)
		}
	}
	if p.pPars.FVerbose != 0 {
		fmt.Printf("Internal AIG objects=%d cuts=%d; merge elapsed %.2fs\n", Amap_ManObjNum(p), p.nCutsUsed, time.Since(started).Seconds())
	}
}

func Amap_CutHead(node *Amap_Obj_t) *Amap_Cut_t {
	if node == nil {
		return nil
	}
	node = Amap_Regular(node)
	cut, _ := node.pData.(*Amap_Cut_t)
	return cut
}

func amapCutLiteral(cut *Amap_Cut_t, complement int32) int32 {
	if cut == nil || cut.iMat == 0 {
		return 0
	}
	return Abc_Var2Lit(int32(cut.iMat), int32(cut.fInv)^complement)
}

func amapComplementSingleton(cut *Amap_Cut_t, complement int32) *Amap_Cut_t {
	if cut == nil || cut.nFans != 1 || complement == 0 {
		return cut
	}
	copyCut := *cut
	copyCut.Fans = append([]int32(nil), cut.Fans...)
	copyCut.Fans[0] = Abc_LitNot(copyCut.Fans[0])
	return &copyCut
}
