package abci

import (
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/vec"
)

// resubWords returns a simulation vector stored in an ABC object or vector slot.
// The manager owns these buffers; this routine only writes their contents.
func resubWords(value any) []uint32 {
	words, ok := value.([]uint32)
	if !ok {
		panic("resub: expected simulation word slice")
	}
	return words
}

func resubComplement(value any) bool { return abc.Abc_ObjIsComplement(value) != 0 }

func resubRegular(value any) *Abc_Obj_t { return abc.Abc_ObjRegular(value) }

func resubPhase(value any) int32 {
	phase := int32(0)
	if resubRegular(value).Phase {
		phase = 1
	}
	if resubComplement(value) {
		phase ^= 1
	}
	return phase
}

func resubSetGraphLeaf(g *dec.Dec_Graph_t, i int32, value any) {
	dec.Dec_GraphSetNodeFunc(g, i, resubRegular(value).Signal())
}

func resubFinishGraph(root *Abc_Obj_t, g *dec.Dec_Graph_t, edge dec.Dec_Edge_t) *dec.Dec_Graph_t {
	dec.Dec_GraphSetRoot(g, edge)
	if root.Phase {
		dec.Dec_GraphComplement(g)
	}
	return g
}

// Abc_ManResubSimulate follows the C topological divisor order. It associates
// each object with its manager-owned word slice and normalizes every divisor
// to phase zero while retaining the source fPhase result.
func Abc_ManResubSimulate(vDivs *Vec_Ptr_t, nLeaves int32, vSims *Vec_Ptr_t, nLeavesMax int32, nWords int32) {
	assert(Vec_PtrSize(vDivs)-nLeaves <= Vec_PtrSize(vSims)-nLeavesMax, "Vec_PtrSize(vDivs) - nLeaves <= Vec_PtrSize(vSims) - nLeavesMax")
	for i := int32(0); i < Vec_PtrSize(vDivs); i++ {
		obj := Vec_PtrEntry(vDivs, i).(*Abc_Obj_t)
		if i < nLeaves {
			obj.CopyData = Vec_PtrEntry(vSims, i)
			continue
		}
		obj.CopyData = Vec_PtrEntry(vSims, i-nLeaves+nLeavesMax)
		out := resubWords(obj.CopyData)
		in0 := resubWords(abc.Abc_ObjCopyData(abc.Abc_ObjFanin0(obj)))
		in1 := resubWords(abc.Abc_ObjCopyData(abc.Abc_ObjFanin1(obj)))
		for k := int32(0); k < nWords; k++ {
			a, b := in0[k], in1[k]
			if abc.Abc_ObjFaninC0(obj) != 0 {
				a = ^a
			}
			if abc.Abc_ObjFaninC1(obj) != 0 {
				b = ^b
			}
			out[k] = a & b
		}
	}
	for i := int32(0); i < Vec_PtrSize(vDivs); i++ {
		obj := Vec_PtrEntry(vDivs, i).(*Abc_Obj_t)
		words := resubWords(obj.CopyData)
		obj.Phase = words[0]&1 != 0
		if obj.Phase {
			for k := int32(0); k < nWords; k++ {
				words[k] = ^words[k]
			}
		}
	}
}

func Abc_ManResubQuit0(pRoot *Abc_Obj_t, pObj any) *dec.Dec_Graph_t {
	g := dec.Dec_GraphCreate(1)
	resubSetGraphLeaf(g, 0, pObj)
	return resubFinishGraph(pRoot, g, dec.Dec_EdgeCreate(0, resubBoolInt(resubRegular(pObj).Phase)))
}

func Abc_ManResubQuit1(pRoot *Abc_Obj_t, pObj0, pObj1 any, fOrGate int32) *dec.Dec_Graph_t {
	assert(resubRegular(pObj0) != resubRegular(pObj1), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1)")
	g := dec.Dec_GraphCreate(2)
	resubSetGraphLeaf(g, 0, pObj0)
	resubSetGraphLeaf(g, 1, pObj1)
	e0, e1 := dec.Dec_EdgeCreate(0, resubPhase(pObj0)), dec.Dec_EdgeCreate(1, resubPhase(pObj1))
	var root dec.Dec_Edge_t
	if fOrGate != 0 {
		root = dec.Dec_GraphAddNodeOr(g, e0, e1)
	} else {
		root = dec.Dec_GraphAddNodeAnd(g, e0, e1)
	}
	return resubFinishGraph(pRoot, g, root)
}

func Abc_ManResubQuit21(pRoot *Abc_Obj_t, pObj0, pObj1, pObj2 any, fOrGate int32) *dec.Dec_Graph_t {
	assert(resubRegular(pObj0) != resubRegular(pObj1), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1)")
	assert(resubRegular(pObj0) != resubRegular(pObj2), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj2)")
	assert(resubRegular(pObj1) != resubRegular(pObj2), "Abc_ObjRegular(pObj1) != Abc_ObjRegular(pObj2)")
	g := dec.Dec_GraphCreate(3)
	resubSetGraphLeaf(g, 0, pObj0)
	resubSetGraphLeaf(g, 1, pObj1)
	resubSetGraphLeaf(g, 2, pObj2)
	e0, e1, e2 := dec.Dec_EdgeCreate(0, resubPhase(pObj0)), dec.Dec_EdgeCreate(1, resubPhase(pObj1)), dec.Dec_EdgeCreate(2, resubPhase(pObj2))
	var root dec.Dec_Edge_t
	if fOrGate != 0 {
		root = dec.Dec_GraphAddNodeOr(g, e0, e1)
		root = dec.Dec_GraphAddNodeOr(g, e2, root)
	} else {
		root = dec.Dec_GraphAddNodeAnd(g, e0, e1)
		root = dec.Dec_GraphAddNodeAnd(g, e2, root)
	}
	return resubFinishGraph(pRoot, g, root)
}

func Abc_ManResubQuit2(pRoot *Abc_Obj_t, pObj0, pObj1, pObj2 any, fOrGate int32) *dec.Dec_Graph_t {
	assert(resubRegular(pObj0) != resubRegular(pObj1), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1)")
	assert(resubRegular(pObj0) != resubRegular(pObj2), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj2)")
	assert(resubRegular(pObj1) != resubRegular(pObj2), "Abc_ObjRegular(pObj1) != Abc_ObjRegular(pObj2)")
	g := dec.Dec_GraphCreate(3)
	resubSetGraphLeaf(g, 0, pObj0)
	resubSetGraphLeaf(g, 1, pObj1)
	resubSetGraphLeaf(g, 2, pObj2)
	e0 := dec.Dec_EdgeCreate(0, resubPhase(pObj0))
	var e1, e2, prev dec.Dec_Edge_t
	if resubComplement(pObj1) && resubComplement(pObj2) {
		e1 = dec.Dec_EdgeCreate(1, resubBoolInt(resubRegular(pObj1).Phase))
		e2 = dec.Dec_EdgeCreate(2, resubBoolInt(resubRegular(pObj2).Phase))
		prev = dec.Dec_GraphAddNodeOr(g, e1, e2)
	} else {
		e1, e2 = dec.Dec_EdgeCreate(1, resubPhase(pObj1)), dec.Dec_EdgeCreate(2, resubPhase(pObj2))
		prev = dec.Dec_GraphAddNodeAnd(g, e1, e2)
	}
	var root dec.Dec_Edge_t
	if fOrGate != 0 {
		root = dec.Dec_GraphAddNodeOr(g, e0, prev)
	} else {
		root = dec.Dec_GraphAddNodeAnd(g, e0, prev)
	}
	return resubFinishGraph(pRoot, g, root)
}

func Abc_ManResubQuit3(pRoot *Abc_Obj_t, pObj0, pObj1, pObj2, pObj3 any, fOrGate int32) *dec.Dec_Graph_t {
	assert(resubRegular(pObj0) != resubRegular(pObj1), "Abc_ObjRegular(pObj0) != Abc_ObjRegular(pObj1)")
	assert(resubRegular(pObj2) != resubRegular(pObj3), "Abc_ObjRegular(pObj2) != Abc_ObjRegular(pObj3)")
	g := dec.Dec_GraphCreate(4)
	for i, value := range []any{pObj0, pObj1, pObj2, pObj3} {
		resubSetGraphLeaf(g, int32(i), value)
	}
	var e0, e1, e2, e3 dec.Dec_Edge_t
	var left, right dec.Dec_Edge_t
	if resubComplement(pObj0) && resubComplement(pObj1) {
		e0 = dec.Dec_EdgeCreate(0, resubBoolInt(resubRegular(pObj0).Phase))
		e1 = dec.Dec_EdgeCreate(1, resubBoolInt(resubRegular(pObj1).Phase))
		left = dec.Dec_GraphAddNodeOr(g, e0, e1)
		if resubComplement(pObj2) && resubComplement(pObj3) {
			e2 = dec.Dec_EdgeCreate(2, resubBoolInt(resubRegular(pObj2).Phase))
			e3 = dec.Dec_EdgeCreate(3, resubBoolInt(resubRegular(pObj3).Phase))
			right = dec.Dec_GraphAddNodeOr(g, e2, e3)
		} else {
			e2 = dec.Dec_EdgeCreate(2, resubPhase(pObj2))
			e3 = dec.Dec_EdgeCreate(3, resubPhase(pObj3))
			right = dec.Dec_GraphAddNodeAnd(g, e2, e3)
		}
	} else {
		e0 = dec.Dec_EdgeCreate(0, resubPhase(pObj0))
		e1 = dec.Dec_EdgeCreate(1, resubPhase(pObj1))
		left = dec.Dec_GraphAddNodeAnd(g, e0, e1)
		if resubComplement(pObj2) && resubComplement(pObj3) {
			e2 = dec.Dec_EdgeCreate(2, resubBoolInt(resubRegular(pObj2).Phase))
			e3 = dec.Dec_EdgeCreate(3, resubBoolInt(resubRegular(pObj3).Phase))
			right = dec.Dec_GraphAddNodeOr(g, e2, e3)
		} else {
			e2 = dec.Dec_EdgeCreate(2, resubPhase(pObj2))
			e3 = dec.Dec_EdgeCreate(3, resubPhase(pObj3))
			right = dec.Dec_GraphAddNodeAnd(g, e2, e3)
		}
	}
	var root dec.Dec_Edge_t
	if fOrGate != 0 {
		root = dec.Dec_GraphAddNodeOr(g, left, right)
	} else {
		root = dec.Dec_GraphAddNodeAnd(g, left, right)
	}
	return resubFinishGraph(pRoot, g, root)
}

func resubBoolInt(value bool) int32 {
	if value {
		return 1
	}
	return 0
}

func Abc_ManResubDivsS(p *Abc_ManRes_t, Required int32) {
	vec.Vec_PtrClear(p.vDivs1UP)
	vec.Vec_PtrClear(p.vDivs1UN)
	vec.Vec_PtrClear(p.vDivs1B)
	root := resubWords(p.pRoot.CopyData)
	for i := int32(0); i < p.nDivs; i++ {
		value := vec.Vec_PtrEntry(p.vDivs, int(i))
		obj := resubRegular(value)
		if int32(obj.Level) > Required-1 {
			continue
		}
		data := resubWords(obj.CopyData)
		positive, negative := true, true
		for w := int32(0); w < p.nWords; w++ {
			positive = positive && data[w]&^root[w]&p.pCareSet[w] == 0
			negative = negative && ^data[w]&root[w]&p.pCareSet[w] == 0
		}
		if positive {
			vec.Vec_PtrPush(p.vDivs1UP, value)
			continue
		}
		if complement := abc.Abc_ObjNot(value); !resubComplement(value) && resubContained(data, root, p.pCareSet, p.nWords, true) {
			vec.Vec_PtrPush(p.vDivs1UP, complement)
			continue
		}
		if negative {
			vec.Vec_PtrPush(p.vDivs1UN, value)
			continue
		}
		if complement := abc.Abc_ObjNot(value); !resubComplement(value) && resubContained(data, root, p.pCareSet, p.nWords, false) {
			vec.Vec_PtrPush(p.vDivs1UN, complement)
			continue
		}
		vec.Vec_PtrPush(p.vDivs1B, value)
	}
}

func resubContained(a, root, care []uint32, nWords int32, positive bool) bool {
	for w := int32(0); w < nWords; w++ {
		var bad uint32
		if positive {
			bad = ^a[w] & ^root[w] & care[w]
		} else {
			bad = a[w] & root[w] & care[w]
		}
		if bad != 0 {
			return false
		}
	}
	return true
}

func Abc_ManResubDivsD(p *Abc_ManRes_t, Required int32) {
	vec.Vec_PtrClear(p.vDivs2UP0)
	vec.Vec_PtrClear(p.vDivs2UP1)
	vec.Vec_PtrClear(p.vDivs2UN0)
	vec.Vec_PtrClear(p.vDivs2UN1)
	root := resubWords(p.pRoot.CopyData)
	for i := 0; i < vec.Vec_PtrSize(p.vDivs1B); i++ {
		v0 := vec.Vec_PtrEntry(p.vDivs1B, i)
		o0 := resubRegular(v0)
		if int32(o0.Level) > Required-2 {
			continue
		}
		d0 := resubWords(o0.CopyData)
		for j := i + 1; j < vec.Vec_PtrSize(p.vDivs1B); j++ {
			v1 := vec.Vec_PtrEntry(p.vDivs1B, j)
			o1 := resubRegular(v1)
			if int32(o1.Level) > Required-2 {
				continue
			}
			d1 := resubWords(o1.CopyData)
			if vec.Vec_PtrSize(p.vDivs2UP0) < ABC_RS_DIV2_MAX {
				for _, flags := range [][2]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
					if resubPairRelation(d0, d1, root, p.pCareSet, p.nWords, flags[0], flags[1], true) {
						pushResubPair(p.vDivs2UP0, p.vDivs2UP1, v0, v1, flags)
					}
				}
			}
			if vec.Vec_PtrSize(p.vDivs2UN0) < ABC_RS_DIV2_MAX {
				for _, flags := range [][2]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
					if resubPairRelation(d0, d1, root, p.pCareSet, p.nWords, flags[0], flags[1], false) {
						pushResubPair(p.vDivs2UN0, p.vDivs2UN1, v0, v1, flags)
					}
				}
			}
		}
	}
}

func pushResubPair(out0, out1 *Vec_Ptr_t, value0, value1 any, invert [2]bool) {
	if invert[0] {
		value0 = abc.Abc_ObjNot(value0)
	}
	if invert[1] {
		value1 = abc.Abc_ObjNot(value1)
	}
	vec.Vec_PtrPush(out0, value0)
	vec.Vec_PtrPush(out1, value1)
}

func resubPairRelation(a, b, root, care []uint32, nWords int32, invert0, invert1, positive bool) bool {
	for w := int32(0); w < nWords; w++ {
		x, y := a[w], b[w]
		if invert0 {
			x = ^x
		}
		if invert1 {
			y = ^y
		}
		got := x & y
		if invert0 && invert1 {
			got = a[w] | b[w]
		}
		if positive {
			if got&^root[w]&care[w] != 0 {
				return false
			}
		} else if ^got&root[w]&care[w] != 0 {
			return false
		}
	}
	return true
}

func resubEqual(got, root, care []uint32, nWords int32) bool {
	for w := int32(0); w < nWords; w++ {
		if (got[w]^root[w])&care[w] != 0 {
			return false
		}
	}
	return true
}

func resubLiteral(words []uint32, value any, word int) uint32 {
	if resubComplement(value) {
		return ^words[word]
	}
	return words[word]
}

func Abc_ManResubQuit(p *Abc_ManRes_t) *dec.Dec_Graph_t {
	root := resubWords(p.pRoot.CopyData)
	for w := int32(0); w < p.nWords; w++ {
		if root[w]&p.pCareSet[w] != 0 {
			return nil
		}
	}
	if p.pRoot.Phase {
		return dec.Dec_GraphCreateConst1()
	}
	return dec.Dec_GraphCreateConst0()
}

func Abc_ManResubDivs0(p *Abc_ManRes_t) *dec.Dec_Graph_t {
	root := resubWords(p.pRoot.CopyData)
	for i := int32(0); i < p.nDivs; i++ {
		value := vec.Vec_PtrEntry(p.vDivs, int(i))
		data := resubWords(resubRegular(value).CopyData)
		if resubEqual(data, root, p.pCareSet, p.nWords) {
			return Abc_ManResubQuit0(p.pRoot, value)
		}
	}
	return nil
}

func Abc_ManResubDivs1(p *Abc_ManRes_t, _ int32) *dec.Dec_Graph_t {
	root := resubWords(p.pRoot.CopyData)
	for i := 0; i < vec.Vec_PtrSize(p.vDivs1UP); i++ {
		a := vec.Vec_PtrEntry(p.vDivs1UP, i)
		wa := resubWords(resubRegular(a).CopyData)
		for j := i + 1; j < vec.Vec_PtrSize(p.vDivs1UP); j++ {
			b := vec.Vec_PtrEntry(p.vDivs1UP, j)
			wb := resubWords(resubRegular(b).CopyData)
			if resubPairEqual(wa, wb, a, b, root, p.pCareSet, p.nWords, true) {
				p.nUsedNode1Or++
				return Abc_ManResubQuit1(p.pRoot, a, b, 1)
			}
		}
	}
	for i := 0; i < vec.Vec_PtrSize(p.vDivs1UN); i++ {
		a := vec.Vec_PtrEntry(p.vDivs1UN, i)
		wa := resubWords(resubRegular(a).CopyData)
		for j := i + 1; j < vec.Vec_PtrSize(p.vDivs1UN); j++ {
			b := vec.Vec_PtrEntry(p.vDivs1UN, j)
			wb := resubWords(resubRegular(b).CopyData)
			// Preserve the source's sequential if statements: the last matching
			// complemented-pattern check determines the returned candidate.
			match := false
			if resubComplement(a) && resubComplement(b) {
				match = resubPairWordsEqual(wa, wb, root, p.pCareSet, p.nWords, 3, false)
			}
			if resubComplement(a) {
				match = resubPairWordsEqual(wa, wb, root, p.pCareSet, p.nWords, 2, false)
			}
			if resubComplement(b) {
				match = resubPairWordsEqual(wa, wb, root, p.pCareSet, p.nWords, 1, false)
			} else {
				match = resubPairWordsEqual(wa, wb, root, p.pCareSet, p.nWords, 0, false)
			}
			if match {
				p.nUsedNode1And++
				return Abc_ManResubQuit1(p.pRoot, a, b, 0)
			}
		}
	}
	return nil
}

func resubPairEqual(a, b []uint32, va, vb any, root, care []uint32, nWords int32, isOr bool) bool {
	flags := 0
	if resubComplement(va) {
		flags |= 2
	}
	if resubComplement(vb) {
		flags |= 1
	}
	return resubPairWordsEqual(a, b, root, care, nWords, flags, isOr)
}

func resubPairWordsEqual(a, b, root, care []uint32, nWords int32, flags int, isOr bool) bool {
	for w := int32(0); w < nWords; w++ {
		x, y := a[w], b[w]
		if flags&2 != 0 {
			x = ^x
		}
		if flags&1 != 0 {
			y = ^y
		}
		got := x & y
		if isOr {
			got = x | y
		}
		if (got^root[w])&care[w] != 0 {
			return false
		}
	}
	return true
}

func Abc_ManResubDivs12(p *Abc_ManRes_t, Required int32) *dec.Dec_Graph_t {
	root := resubWords(p.pRoot.CopyData)
	for _, list := range []*Vec_Ptr_t{p.vDivs1UP, p.vDivs1UN} {
		isOr := list == p.vDivs1UP
		for i := 0; i < vec.Vec_PtrSize(list); i++ {
			a := vec.Vec_PtrEntry(list, i)
			wa := resubWords(resubRegular(a).CopyData)
			for j := i + 1; j < vec.Vec_PtrSize(list); j++ {
				b := vec.Vec_PtrEntry(list, j)
				wb := resubWords(resubRegular(b).CopyData)
				for k := j + 1; k < vec.Vec_PtrSize(list); k++ {
					c := vec.Vec_PtrEntry(list, k)
					wc := resubWords(resubRegular(c).CopyData)
					if !resubTripleEqual(wa, wb, wc, a, b, c, root, p.pCareSet, p.nWords, isOr) {
						continue
					}
					la, lb, lc := int32(resubRegular(a).Level), int32(resubRegular(b).Level), int32(resubRegular(c).Level)
					maxLevel := resubMax32(la, resubMax32(lb, lc))
					assert(maxLevel <= Required-1, "LevelMax <= Required - 1")
					var maxValue, min0, min1 any
					if la == maxLevel {
						maxValue, min0, min1 = a, b, c
					}
					if lb == maxLevel {
						if maxValue != nil {
							continue
						}
						maxValue, min0, min1 = b, a, c
					}
					if lc == maxLevel {
						if maxValue != nil {
							continue
						}
						maxValue, min0, min1 = c, a, b
					}
					if isOr {
						p.nUsedNode2Or++
					} else {
						p.nUsedNode2And++
					}
					assert(min0 != nil && min1 != nil && maxValue != nil, "unique maximum level operand")
					return Abc_ManResubQuit21(p.pRoot, min0, min1, maxValue, resubBoolInt(isOr))
				}
			}
		}
	}
	return nil
}

func resubTripleEqual(a, b, c []uint32, va, vb, vc any, root, care []uint32, nWords int32, isOr bool) bool {
	for w := int32(0); w < nWords; w++ {
		x, y, z := resubLiteral(a, va, int(w)), resubLiteral(b, vb, int(w)), resubLiteral(c, vc, int(w))
		got := x & y & z
		if isOr {
			got = x | y | z
		}
		if (got^root[w])&care[w] != 0 {
			return false
		}
	}
	return true
}

func resubMax32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func Abc_ManResubDivs2(p *Abc_ManRes_t, _ int32) *dec.Dec_Graph_t {
	root := resubWords(p.pRoot.CopyData)
	if graph := resubDivs2Pass(p, p.vDivs1UP, p.vDivs2UP0, p.vDivs2UP1, root, true); graph != nil {
		return graph
	}
	return resubDivs2Pass(p, p.vDivs1UN, p.vDivs2UN0, p.vDivs2UN1, root, false)
}

func resubDivs2Pass(p *Abc_ManRes_t, singles, pair0, pair1 *Vec_Ptr_t, root []uint32, positive bool) *dec.Dec_Graph_t {
	for i := 0; i < vec.Vec_PtrSize(singles); i++ {
		a := vec.Vec_PtrEntry(singles, i)
		wa := resubWords(resubRegular(a).CopyData)
		for j := 0; j < vec.Vec_PtrSize(pair0); j++ {
			b, c := vec.Vec_PtrEntry(pair0, j), vec.Vec_PtrEntry(pair1, j)
			wb, wc := resubWords(resubRegular(b).CopyData), resubWords(resubRegular(c).CopyData)
			if !resubDivs2Equal(wa, wb, wc, a, b, c, root, p.pCareSet, p.nWords, positive) {
				continue
			}
			if positive {
				p.nUsedNode2OrAnd++
				return Abc_ManResubQuit2(p.pRoot, a, b, c, 1)
			}
			p.nUsedNode2AndOr++
			return Abc_ManResubQuit2(p.pRoot, a, b, c, 0)
		}
	}
	return nil
}

func resubDivs2Equal(a, b, c []uint32, va, vb, vc any, root, care []uint32, nWords int32, positive bool) bool {
	for w := int32(0); w < nWords; w++ {
		x, y, z := a[w], b[w], c[w]
		f0, f1, f2 := resubComplement(va), resubComplement(vb), resubComplement(vc)
		if f0 {
			x = ^x
		}
		var got uint32
		if positive {
			if f1 && f2 {
				got = x | (y | z)
			} else if f1 {
				got = x | (^y & z)
			} else if f2 {
				got = x | (y & ^z)
			} else {
				got = x | (y & z)
			}
		} else if f1 && f2 {
			got = x & (y | z)
		} else if f1 {
			got = x & (^y & z)
		} else if f2 {
			got = x & (y & ^z)
		} else {
			got = x & (y & z)
		}
		if (got^root[w])&care[w] != 0 {
			return false
		}
	}
	return true
}

func Abc_ManResubDivs3(p *Abc_ManRes_t, _ int32) *dec.Dec_Graph_t {
	for i := 0; i < vec.Vec_PtrSize(p.vDivs2UP0); i++ {
		a, b := vec.Vec_PtrEntry(p.vDivs2UP0, i), vec.Vec_PtrEntry(p.vDivs2UP1, i)
		wa, wb := resubWords(resubRegular(a).CopyData), resubWords(resubRegular(b).CopyData)
		for j := i + 1; j < vec.Vec_PtrSize(p.vDivs2UP0); j++ {
			c, d := vec.Vec_PtrEntry(p.vDivs2UP0, j), vec.Vec_PtrEntry(p.vDivs2UP1, j)
			wc, wd := resubWords(resubRegular(c).CopyData), resubWords(resubRegular(d).CopyData)
			if !resubDivs3Equal(wa, wb, wc, wd, a, b, c, d, p.pRoot.CopyData, p.pCareSet, p.nWords) {
				continue
			}
			p.nUsedNode3OrAnd++
			return Abc_ManResubQuit3(p.pRoot, a, b, c, d, 1)
		}
	}
	return nil
}

func resubDivs3Equal(a, b, c, d []uint32, va, vb, vc, vd any, rootValue any, care []uint32, nWords int32) bool {
	root := resubWords(rootValue)
	for w := int32(0); w < nWords; w++ {
		x0, x1 := a[w], b[w]
		y0, y1 := c[w], d[w]
		f0 := resubComplement(va)
		f1 := resubComplement(vb)
		f2 := resubComplement(vc)
		f3 := resubComplement(vd)
		if f0 {
			x0 = ^x0
		}
		if f1 {
			x1 = ^x1
		}
		if f2 {
			y0 = ^y0
		}
		if f3 {
			y1 = ^y1
		}
		left, right := x0&x1, y0&y1
		if f0 && f1 {
			left = a[w] | b[w]
		}
		if f2 && f3 {
			right = c[w] | d[w]
		}
		if ((left|right)^root[w])&care[w] != 0 {
			return false
		}
	}
	return true
}
