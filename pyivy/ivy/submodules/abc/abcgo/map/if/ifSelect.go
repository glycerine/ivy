package if_

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func If_ObjConePrint_rec(man *If_Man_t, obj *If_Obj_t, visited *Vec_Ptr_t) {
	cut := If_ObjCutBest(obj)
	if If_CutDataInt(cut) != 0 {
		return
	}
	Vec_PtrPush(visited, cut)
	If_CutSetDataInt(cut, -1)
	if If_ObjIsCi(obj) != 0 {
		return
	}
	if obj.pEquiv != nil {
		If_ObjConePrint_rec(man, obj.pEquiv, visited)
	}
	If_ObjConePrint_rec(man, obj.pFanin0, visited)
	If_ObjConePrint_rec(man, obj.pFanin1, visited)
	choice := int32(0)
	if obj.pEquiv != nil {
		choice = obj.pEquiv.Id
	}
	fmt.Printf("%5d = %5d & %5d | %5d\n", obj.Id, obj.pFanin0.Id, obj.pFanin1.Id, choice)
}

func If_ObjConePrint(man *If_Man_t, obj *If_Obj_t) {
	Vec_PtrClear(man.vTemp)
	If_ObjConePrint_rec(man, obj, man.vTemp)
	for _, entry := range Vec_PtrArray(man.vTemp) {
		If_CutSetDataInt(entry.(*If_Cut_t), 0)
	}
	cut := If_ObjCutBest(obj)
	fmt.Print("Cut ")
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		fmt.Printf("%d ", If_CutLeaf(man, cut, i).Id)
	}
	fmt.Println()
}

func If_ManNodeShapeMap_rec(man *If_Man_t, obj *If_Obj_t, visited *Vec_Ptr_t, shape *Vec_Int_t) int32 {
	cut := If_ObjCutBest(obj)
	if value := If_CutDataInt(cut); value != 0 {
		return value
	}
	Vec_PtrPush(visited, cut)
	If_CutSetDataInt(cut, -1)
	if If_ObjIsCi(obj) != 0 {
		return -1
	}
	for alternative := obj; alternative != nil; alternative = alternative.pEquiv {
		if If_ManNodeShapeMap_rec(man, alternative.pFanin0, visited, shape) == -1 {
			continue
		}
		if If_ManNodeShapeMap_rec(man, alternative.pFanin1, visited, shape) == -1 {
			continue
		}
		vec.Vec_IntPush(shape, obj.Id)
		vec.Vec_IntPush(shape, alternative.Id)
		If_CutSetDataInt(cut, 1)
		break
	}
	return If_CutDataInt(cut)
}

func If_ManNodeShapeMap(man *If_Man_t, obj *If_Obj_t, shape *Vec_Int_t) int32 {
	cut := If_ObjCutBest(obj)
	assert(cut.nLeaves > 1, "cut has multiple leaves")
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		leaf := If_CutLeaf(man, cut, i)
		assert(If_CutDataInt(If_ObjCutBest(leaf)) == 0, "leaf scratch is clear")
		If_CutSetDataInt(If_ObjCutBest(leaf), 1)
	}
	vec.Vec_IntClear(shape)
	Vec_PtrClear(man.vTemp)
	result := If_ManNodeShapeMap_rec(man, obj, man.vTemp, shape)
	if result != -1 {
		for i := int32(0); i < int32(cut.nLeaves); i++ {
			leaf := If_CutLeaf(man, cut, i)
			If_CutSetDataInt(If_ObjCutBest(leaf), 0)
		}
	}
	for _, entry := range Vec_PtrArray(man.vTemp) {
		If_CutSetDataInt(entry.(*If_Cut_t), 0)
	}
	if result == -1 {
		return 0
	}
	return 1
}

func If_WordCountOnes(x uint32) int32 {
	x = (x & 0x55555555) + ((x >> 1) & 0x55555555)
	x = (x & 0x33333333) + ((x >> 2) & 0x33333333)
	x = (x & 0x0f0f0f0f) + ((x >> 4) & 0x0f0f0f0f)
	x = (x & 0x00ff00ff) + ((x >> 8) & 0x00ff00ff)
	return int32((x & 0xffff) + (x >> 16))
}

func If_ManNodeShapeMap2_rec(man *If_Man_t, obj *If_Obj_t, visited *Vec_Ptr_t, shape *Vec_Int_t) int32 {
	cut := If_ObjCutBest(obj)
	if value := If_CutDataInt(cut); value != 0 {
		return value
	}
	Vec_PtrPush(visited, cut)
	If_CutSetDataInt(cut, -1)
	if If_ObjIsCi(obj) != 0 {
		return -1
	}
	var best *If_Obj_t
	var bestMask int32
	for alternative := obj; alternative != nil; alternative = alternative.pEquiv {
		a := If_ManNodeShapeMap2_rec(man, alternative.pFanin0, visited, shape)
		if a == -1 {
			continue
		}
		b := If_ManNodeShapeMap2_rec(man, alternative.pFanin1, visited, shape)
		if b == -1 {
			continue
		}
		combined := a | b
		if best == nil || If_WordCountOnes(uint32(combined)) > If_WordCountOnes(uint32(bestMask)) {
			best, bestMask = alternative, combined
		}
	}
	if best != nil {
		vec.Vec_IntPush(shape, obj.Id)
		vec.Vec_IntPush(shape, best.Id)
		If_CutSetDataInt(cut, bestMask)
	}
	return If_CutDataInt(cut)
}

func If_ManNodeShapeMap2(man *If_Man_t, obj *If_Obj_t, shape *Vec_Int_t) int32 {
	cut := If_ObjCutBest(obj)
	assert(cut.nLeaves > 1, "cut has multiple leaves")
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		leaf := If_CutLeaf(man, cut, i)
		If_CutSetDataInt(If_ObjCutBest(leaf), 1<<uint(i))
	}
	vec.Vec_IntClear(shape)
	Vec_PtrClear(man.vTemp)
	result := If_ManNodeShapeMap2_rec(man, obj, man.vTemp, shape)
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		If_CutSetDataInt(If_ObjCutBest(If_CutLeaf(man, cut, i)), 0)
	}
	for _, entry := range Vec_PtrArray(man.vTemp) {
		If_CutSetDataInt(entry.(*If_Cut_t), 0)
	}
	if result == -1 {
		return 0
	}
	return 1
}

func If_ManConeCollect_rec(man *If_Man_t, obj *If_Obj_t, visited, cone *Vec_Ptr_t) int32 {
	cut := If_ObjCutBest(obj)
	if value := If_CutDataInt(cut); value != 0 {
		return value
	}
	Vec_PtrPush(visited, cut)
	If_CutSetDataInt(cut, -1)
	if If_ObjIsCi(obj) != 0 {
		return -1
	}
	var nodeAdded, rootAdded bool
	for alternative := obj; alternative != nil; alternative = alternative.pEquiv {
		if If_ManConeCollect_rec(man, alternative.pFanin0, visited, cone) == -1 {
			continue
		}
		if If_ManConeCollect_rec(man, alternative.pFanin1, visited, cone) == -1 {
			continue
		}
		nodeAdded = true
		If_CutSetDataInt(cut, 1)
		Vec_PtrPush(cone, alternative)
		if alternative == obj {
			rootAdded = true
		}
	}
	if nodeAdded && !rootAdded {
		Vec_PtrPush(cone, obj)
	}
	return If_CutDataInt(cut)
}

func If_ManConeCollect(man *If_Man_t, obj *If_Obj_t, cut *If_Cut_t) *Vec_Ptr_t {
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		leaf := If_CutLeaf(man, cut, i)
		assert(If_CutDataInt(If_ObjCutBest(leaf)) == 0, "leaf scratch is clear")
		If_CutSetDataInt(If_ObjCutBest(leaf), 1)
	}
	cone := Vec_PtrAlloc(100)
	Vec_PtrClear(man.vTemp)
	assert(If_ManConeCollect_rec(man, obj, man.vTemp, cone) != 0, "cone collection succeeds")
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		If_CutSetDataInt(If_ObjCutBest(If_CutLeaf(man, cut, i)), 0)
	}
	for _, entry := range Vec_PtrArray(man.vTemp) {
		If_CutSetDataInt(entry.(*If_Cut_t), 0)
	}
	return cone
}

func If_ObjSatVar(obj *If_Obj_t) int32       { return If_CutDataInt(If_ObjCutBest(obj)) }
func If_ObjSetSatVar(obj *If_Obj_t, v int32) { If_CutSetDataInt(If_ObjCutBest(obj), v) }

func satLit(v int32, negative bool) bsat.Literal {
	if negative {
		return bsat.Negative(v)
	}
	return bsat.Positive(v)
}
func satSolverAddChoice(s *bsat.Solver2, out int32, vars []int32) {
	clause := make([]bsat.Literal, 0, len(vars)+1)
	for _, v := range vars {
		clause = append(clause, satLit(v, false))
	}
	clause = append(clause, satLit(out, true))
	assert(s.AddClause(clause...), "choice clause accepted")
	for _, v := range vars {
		assert(s.AddClause(satLit(out, false), satLit(v, true)), "choice implication accepted")
	}
	for i := range vars {
		for j := i + 1; j < len(vars); j++ {
			assert(s.AddClause(satLit(vars[i], true), satLit(vars[j], true)), "choice exclusivity accepted")
		}
	}
}

func ifSatModelValue(s *bsat.Solver2, v int32) int32 {
	value, ok := s.Value(v)
	if !ok {
		return 0
	}
	if value {
		return 1
	}
	return 0
}

func If_ManNodeShape2_rec(s *bsat.Solver2, man *If_Man_t, obj *If_Obj_t, shape *Vec_Int_t) {
	assert(ifSatModelValue(s, If_ObjSatVar(obj)) == 1, "external variable selected")
	if obj.fMark != 0 {
		return
	}
	obj.fMark = 1
	chosen := obj
	for chosen != nil && ifSatModelValue(s, If_ObjSatVar(chosen)+1) != 1 {
		chosen = chosen.pEquiv
	}
	assert(chosen != nil, "a choice implementation is selected")
	If_ManNodeShape2_rec(s, man, chosen.pFanin0, shape)
	If_ManNodeShape2_rec(s, man, chosen.pFanin1, shape)
	vec.Vec_IntPush(shape, obj.Id)
	vec.Vec_IntPush(shape, chosen.Id)
}

func If_ManNodeShapeSat(man *If_Man_t, obj *If_Obj_t, shape *Vec_Int_t) int32 {
	cut := If_ObjCutBest(obj)
	assert(cut.nLeaves > 1, "SAT shape requires a nontrivial cut")
	cone := If_ManConeCollect(man, obj, cut)
	defer Vec_PtrFree(cone)
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		leaf := If_CutLeaf(man, cut, i)
		assert(If_ObjSatVar(leaf) == 0, "SAT scratch is clear")
		If_ObjSetSatVar(leaf, 2*(i+1))
	}
	for i, e := range Vec_PtrArray(cone) {
		n := e.(*If_Obj_t)
		assert(If_ObjSatVar(n) == 0, "SAT scratch is clear")
		If_ObjSetSatVar(n, 2*(int32(i)+1+int32(cut.nLeaves)))
	}
	solver := bsat.NewSolver2()
	for i := int32(0); i < 2*(int32(cut.nLeaves)+int32(Vec_PtrSize(cone))+1); i++ {
		solver.NewVariable()
	}
	for _, entry := range Vec_PtrArray(cone) {
		n := entry.(*If_Obj_t)
		assert(If_ObjIsAnd(n) != 0, "cone contains AND nodes")
		var alternatives []int32
		for a := n; a != nil; a = a.pEquiv {
			if If_ObjSatVar(a) != 0 {
				alternatives = append(alternatives, If_ObjSatVar(a)+1)
			}
		}
		assert(len(alternatives) > 0, "choice node has a candidate")
		satSolverAddChoice(solver, If_ObjSatVar(n), alternatives)
		for _, fanin := range []*If_Obj_t{n.pFanin0, n.pFanin1} {
			if If_ObjSatVar(fanin) > 0 {
				assert(solver.AddClause(satLit(If_ObjSatVar(n)+1, true), satLit(If_ObjSatVar(fanin), false)), "AND implication clause accepted")
			}
		}
	}
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		assert(solver.AddClause(satLit(If_ObjSatVar(If_CutLeaf(man, cut, i)), false)), "cut input unit accepted")
	}
	assert(solver.AddClause(satLit(If_ObjSatVar(obj), false)), "output unit accepted")
	assert(solver.Solve() == bsat.Satisfiable, "shape SAT instance has a model")
	vec.Vec_IntClear(shape)
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		leaf := If_CutLeaf(man, cut, i)
		assert(leaf.fMark == 0, "leaf mark clear")
		leaf.fMark = 1
	}
	If_ManNodeShape2_rec(solver, man, obj, shape)
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		leaf := If_CutLeaf(man, cut, i)
		If_ObjSetSatVar(leaf, 0)
		leaf.fMark = 0
	}
	for _, entry := range Vec_PtrArray(cone) {
		n := entry.(*If_Obj_t)
		If_ObjSetSatVar(n, 0)
		n.fMark = 0
	}
	return 1
}

func If_ManCheckShape(man *If_Man_t, obj *If_Obj_t, shape *Vec_Int_t) int32 {
	cut := If_ObjCutBest(obj)
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		assert(If_CutLeaf(man, cut, i).fMark == 0, "cut leaf mark clear")
	}
	ids := vec.Vec_IntArray(shape)
	for i := 0; i+1 < len(ids); i += 2 {
		n := If_ManObj(man, ids[i+1])
		n.pFanin0.fMark = 1
		n.pFanin1.fMark = 1
	}
	result := int32(1)
	for i := int32(0); i < int32(cut.nLeaves); i++ {
		leaf := If_CutLeaf(man, cut, i)
		if leaf.fMark == 0 {
			result = 0
		} else {
			leaf.fMark = 0
		}
	}
	for i := 0; i+1 < len(ids); i += 2 {
		n := If_ManObj(man, ids[i+1])
		n.pFanin0.fMark = 0
		n.pFanin1.fMark = 0
	}
	return result
}

func If_ManNodeShape(man *If_Man_t, obj *If_Obj_t, shape *Vec_Int_t, exact int32) int32 {
	if If_ManNodeShapeMap(man, obj, shape) != 0 && (exact == 0 || If_ManCheckShape(man, obj, shape) != 0) {
		return 1
	}
	if If_ManNodeShapeMap2(man, obj, shape) == 0 {
		return 0
	}
	return If_ManCheckShape(man, obj, shape)
}
