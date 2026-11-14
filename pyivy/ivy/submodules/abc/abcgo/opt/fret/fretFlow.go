package fret

import (
	abc "github.com/glycerine/abcgo/base/abc"
	vec "github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [fretFlow.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Flow-based retiming package.]

  Synopsis    [Max-flow computation.]

  Author      [Aaron Hurst]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2008.]

  Revision    [$Id: fretFlow.c,v 1.00 2008/01/01 00:00:00 ahurst Exp $]

***********************************************************************/

//#include "fretime.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void dfsfast_e_retreat( abc.Abc_Obj_t *pObj );

// static void dfsfast_r_retreat( abc.Abc_Obj_t *pObj );

// Original macro:
// #define FDIST(xn, xe, yn, ye) (FDATA(xn)->xe##_dist == (FDATA(yn)->ye##_dist + 1))

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Fast DFS.]

  Description [Uses sink-distance-histogram heuristic.  May not find all
               flow paths: this occurs in a small number of cases where
               the flow predecessor points to a non-adjacent node and
               the distance ordering is perturbed.]

  SideEffects []

  SeeAlso     []

***********************************************************************/

func dfsfast_preorder(pNtk *abc.Abc_Ntk_t) {
	nObjs := abc.Abc_NtkObjNumMax(pNtk)
	queueNodes := vec.Vec_PtrAlloc(int(nObjs))
	queueEnds := vec.Vec_IntAlloc(int(nObjs))
	defer vec.Vec_PtrFree(queueNodes)
	defer vec.Vec_IntFree(queueEnds)
	vec.Vec_IntFill(pManMR.vSinkDistHist, vec.Vec_IntSize(pManMR.vSinkDistHist), 0)

	if pManMR.maxDelay != 0 {
		for id := int32(0); id < nObjs; id++ {
			obj := abc.Abc_NtkObj(pNtk, id)
			if obj == nil {
				continue
			}
			outgoing := FTIMEEDGES(obj)
			for j := 0; j < vec.Vec_PtrSize(outgoing); j++ {
				next, ok := vec.Vec_PtrEntry(outgoing, j).(*abc.Abc_Obj_t)
				if !ok || next == nil {
					panic("retiming timing edge contains a non-object")
				}
				reverse := flowData(next).vNodes
				if reverse == nil {
					reverse = vec.Vec_PtrAlloc(2)
					flowData(next).vNodes = reverse
				}
				vec.Vec_PtrPush(reverse, obj)
			}
		}
	}

	for id := int32(0); id < nObjs; id++ {
		obj := abc.Abc_NtkObj(pNtk, id)
		if obj == nil {
			continue
		}
		blocked := flowData(obj).mark&BLOCK_OR_CONS&pManMR.constraintMask != 0
		if abc.Abc_ObjIsPo(obj) != 0 || abc.Abc_ObjIsLatch(obj) != 0 || (pManMR.fIsForward != 0 && blocked) {
			vec.Vec_PtrPush(queueNodes, obj)
			vec.Vec_IntPush(queueEnds, 'r')
			flowData(obj).r_dist = 1
		} else if abc.Abc_ObjIsPi(obj) != 0 || (pManMR.fIsForward == 0 && blocked) {
			vec.Vec_PtrPush(queueNodes, obj)
			vec.Vec_IntPush(queueEnds, 'e')
			flowData(obj).e_dist = 1
		}
	}

	for qpos := 0; qpos < vec.Vec_PtrSize(queueNodes); qpos++ {
		obj, ok := vec.Vec_PtrEntry(queueNodes, qpos).(*abc.Abc_Obj_t)
		if !ok || obj == nil {
			panic("retiming flow queue contains a non-object")
		}
		end := vec.Vec_IntEntry(queueEnds, qpos)
		data := flowData(obj)
		if end == 'r' {
			d := data.r_dist
			if pManMR.fIsForward != 0 {
				for i := int32(0); i < abc.Abc_ObjFaninNum(obj); i++ {
					next := abc.Abc_ObjFanin(obj, i)
					if flowData(next).e_dist == 0 {
						flowData(next).e_dist = d + 1
						vec.Vec_PtrPush(queueNodes, next)
						vec.Vec_IntPush(queueEnds, 'e')
					}
				}
			} else {
				for i := int32(0); i < abc.Abc_ObjFanoutNum(obj); i++ {
					next := abc.Abc_ObjFanout(obj, i)
					if flowData(next).e_dist == 0 {
						flowData(next).e_dist = d + 1
						vec.Vec_PtrPush(queueNodes, next)
						vec.Vec_IntPush(queueEnds, 'e')
					}
				}
			}
			if d == 1 {
				continue
			}
			if pManMR.fIsForward != 0 {
				for i := int32(0); i < abc.Abc_ObjFanoutNum(obj); i++ {
					next := abc.Abc_ObjFanout(obj, i)
					if flowData(next).r_dist == 0 && abc.Abc_ObjIsLatch(next) == 0 {
						flowData(next).r_dist = d + 1
						vec.Vec_PtrPush(queueNodes, next)
						vec.Vec_IntPush(queueEnds, 'r')
					}
				}
				if pManMR.maxDelay != 0 && data.vNodes != nil {
					for i := 0; i < vec.Vec_PtrSize(data.vNodes); i++ {
						next, ok := vec.Vec_PtrEntry(data.vNodes, i).(*abc.Abc_Obj_t)
						if !ok || next == nil {
							panic("retiming reverse timing edge contains a non-object")
						}
						if flowData(next).r_dist == 0 {
							flowData(next).r_dist = d + 1
							vec.Vec_PtrPush(queueNodes, next)
							vec.Vec_IntPush(queueEnds, 'r')
						}
					}
				}
			}
			continue
		}

		if abc.Abc_ObjIsLatch(obj) != 0 {
			continue
		}
		d := data.e_dist
		if data.r_dist == 0 {
			data.r_dist = d + 1
			vec.Vec_PtrPush(queueNodes, obj)
			vec.Vec_IntPush(queueEnds, 'r')
		}
		if pManMR.fIsForward == 0 {
			for i := int32(0); i < abc.Abc_ObjFaninNum(obj); i++ {
				next := abc.Abc_ObjFanin(obj, i)
				if flowData(next).e_dist == 0 && abc.Abc_ObjIsLatch(next) == 0 {
					flowData(next).e_dist = d + 1
					vec.Vec_PtrPush(queueNodes, next)
					vec.Vec_IntPush(queueEnds, 'e')
				}
			}
			if pManMR.maxDelay != 0 && data.vNodes != nil {
				for i := 0; i < vec.Vec_PtrSize(data.vNodes); i++ {
					next, ok := vec.Vec_PtrEntry(data.vNodes, i).(*abc.Abc_Obj_t)
					if !ok || next == nil {
						panic("retiming reverse timing edge contains a non-object")
					}
					if flowData(next).e_dist == 0 {
						flowData(next).e_dist = d + 1
						vec.Vec_PtrPush(queueNodes, next)
						vec.Vec_IntPush(queueEnds, 'e')
					}
				}
			}
		}
	}

	if pManMR.maxDelay != 0 {
		for id := int32(0); id < nObjs; id++ {
			if obj := abc.Abc_NtkObj(pNtk, id); obj != nil {
				flowData(obj).vNodes = nil
			}
		}
	}
	for id := int32(0); id < nObjs; id++ {
		obj := abc.Abc_NtkObj(pNtk, id)
		if obj == nil {
			continue
		}
		data := flowData(obj)
		vec.Vec_IntAddToEntry(pManMR.vSinkDistHist, int(data.r_dist), 1)
		vec.Vec_IntAddToEntry(pManMR.vSinkDistHist, int(data.e_dist), 1)
	}
}

func dfsfast_e(pObj *abc.Abc_Obj_t, pPred *abc.Abc_Obj_t) int32 {
	if pManMR.fSinkDistTerminate != 0 {
		return 0
	}
	data := flowData(pObj)
	if data.mark&BLOCK_OR_CONS&pManMR.constraintMask != 0 || abc.Abc_ObjIsPi(pObj) != 0 {
		assert(pPred != nil, "pPred")
		assert(pManMR.fIsForward == 0, "!pManMR.fIsForward")
		return 1
	}

	FSET(pObj, VISITED_E)
	found := false
	if pManMR.fIsForward != 0 {
		for i := int32(0); i < abc.Abc_ObjFanoutNum(pObj); i++ {
			next := abc.Abc_ObjFanout(pObj, i)
			if FTEST(next, VISITED_R) == 0 &&
				flowData(pObj).e_dist == flowData(next).r_dist+1 &&
				dfsfast_r(next, pPred) != 0 {
				found = true
				break
			}
		}
	} else {
		for i := int32(0); i < abc.Abc_ObjFaninNum(pObj); i++ {
			next := abc.Abc_ObjFanin(pObj, i)
			if FTEST(next, VISITED_R) == 0 &&
				flowData(pObj).e_dist == flowData(next).r_dist+1 &&
				dfsfast_r(next, pPred) != 0 {
				found = true
				break
			}
		}
	}
	if found {
		FUNSET(pObj, VISITED_E)
		return 1
	}
	if abc.Abc_ObjIsLatch(pObj) != 0 {
		FUNSET(pObj, VISITED_E)
		dfsfast_e_retreat(pObj)
		return 0
	}

	if pManMR.fIsForward == 0 {
		for i := int32(0); i < abc.Abc_ObjFanoutNum(pObj); i++ {
			next := abc.Abc_ObjFanout(pObj, i)
			if FTEST(next, VISITED_E) == 0 &&
				flowData(pObj).e_dist == flowData(next).e_dist+1 &&
				dfsfast_e(next, pPred) != 0 {
				FUNSET(pObj, VISITED_E)
				return 1
			}
		}
		if pManMR.maxDelay != 0 {
			for i := 0; i < vec.Vec_PtrSize(FTIMEEDGES(pObj)); i++ {
				next, ok := vec.Vec_PtrEntry(FTIMEEDGES(pObj), i).(*abc.Abc_Obj_t)
				if !ok || next == nil {
					panic("retiming timing edge contains a non-object")
				}
				if FTEST(next, VISITED_E) == 0 &&
					flowData(pObj).e_dist == flowData(next).e_dist+1 &&
					dfsfast_e(next, pPred) != 0 {
					FUNSET(pObj, VISITED_E)
					return 1
				}
			}
		}
	}

	if FTEST(pObj, FLOW) != 0 &&
		FTEST(pObj, VISITED_R) == 0 &&
		flowData(pObj).e_dist == flowData(pObj).r_dist+1 &&
		dfsfast_r(pObj, FGETPRED(pObj)) != 0 {
		FUNSET(pObj, FLOW)
		FSETPRED(pObj, nil)
		found = true
	}
	FUNSET(pObj, VISITED_E)
	if found {
		return 1
	}
	dfsfast_e_retreat(pObj)
	return 0
}

func dfsfast_r(pObj *abc.Abc_Obj_t, pPred *abc.Abc_Obj_t) int32 {
	if pManMR.fSinkDistTerminate != 0 {
		return 0
	}
	if abc.Abc_ObjIsLatch(pObj) != 0 ||
		(pManMR.fIsForward != 0 && abc.Abc_ObjIsPo(pObj) != 0) ||
		(pManMR.fIsForward != 0 && flowData(pObj).mark&BLOCK_OR_CONS&pManMR.constraintMask != 0) {
		assert(pPred != nil, "pPred")
		return 1
	}

	FSET(pObj, VISITED_R)
	if FTEST(pObj, FLOW) != 0 {
		oldPred := FGETPRED(pObj)
		if oldPred != nil &&
			FTEST(oldPred, VISITED_E) == 0 &&
			flowData(pObj).r_dist == flowData(oldPred).e_dist+1 &&
			dfsfast_e(oldPred, oldPred) != 0 {
			FSETPRED(pObj, pPred)
			FUNSET(pObj, VISITED_R)
			return 1
		}
	} else if FTEST(pObj, VISITED_E) == 0 &&
		flowData(pObj).r_dist == flowData(pObj).e_dist+1 &&
		dfsfast_e(pObj, pObj) != 0 {
		FSET(pObj, FLOW)
		FSETPRED(pObj, pPred)
		FUNSET(pObj, VISITED_R)
		return 1
	}

	if pManMR.fIsForward != 0 {
		for i := int32(0); i < abc.Abc_ObjFaninNum(pObj); i++ {
			next := abc.Abc_ObjFanin(pObj, i)
			if FTEST(next, VISITED_R) == 0 &&
				flowData(pObj).r_dist == flowData(next).r_dist+1 &&
				abc.Abc_ObjIsLatch(next) == 0 &&
				dfsfast_r(next, pPred) != 0 {
				FUNSET(pObj, VISITED_R)
				return 1
			}
		}
		if pManMR.maxDelay != 0 {
			edges := FTIMEEDGES(pObj)
			for i := 0; i < vec.Vec_PtrSize(edges); i++ {
				next, ok := vec.Vec_PtrEntry(edges, i).(*abc.Abc_Obj_t)
				if !ok || next == nil {
					panic("retiming timing edge contains a non-object")
				}
				if FTEST(next, VISITED_R) == 0 &&
					flowData(pObj).r_dist == flowData(next).r_dist+1 &&
					dfsfast_r(next, pPred) != 0 {
					FUNSET(pObj, VISITED_R)
					return 1
				}
			}
		}
	}

	FUNSET(pObj, VISITED_R)
	dfsfast_r_retreat(pObj)
	return 0
}

func dfsfast_e_retreat(pObj *abc.Abc_Obj_t) {
	oldDist := flowData(pObj).e_dist
	minDist := MAX_DIST
	consider := func(distance int32) {
		if distance != 0 && distance < minDist {
			minDist = distance
		}
	}

	if pManMR.fIsForward != 0 {
		for i := int32(0); i < abc.Abc_ObjFanoutNum(pObj); i++ {
			consider(flowData(abc.Abc_ObjFanout(pObj, i)).r_dist)
		}
	} else {
		for i := int32(0); i < abc.Abc_ObjFaninNum(pObj); i++ {
			consider(flowData(abc.Abc_ObjFanin(pObj, i)).r_dist)
		}
	}

	if abc.Abc_ObjIsLatch(pObj) == 0 {
		if FTEST(pObj, FLOW) != 0 {
			consider(flowData(pObj).r_dist)
		}
		if pManMR.fIsForward == 0 {
			for i := int32(0); i < abc.Abc_ObjFanoutNum(pObj); i++ {
				consider(flowData(abc.Abc_ObjFanout(pObj, i)).e_dist)
			}
			if pManMR.maxDelay != 0 {
				edges := FTIMEEDGES(pObj)
				for i := 0; i < vec.Vec_PtrSize(edges); i++ {
					next, ok := vec.Vec_PtrEntry(edges, i).(*abc.Abc_Obj_t)
					if !ok || next == nil {
						panic("retiming timing edge contains a non-object")
					}
					consider(flowData(next).e_dist)
				}
			}
		}
	}

	minDist++
	if minDist >= MAX_DIST {
		minDist = 0
	}
	flowData(pObj).e_dist = minDist
	updateSinkDistanceHistogram(oldDist, minDist)
}

func dfsfast_r_retreat(pObj *abc.Abc_Obj_t) {
	oldDist := flowData(pObj).r_dist
	minDist := MAX_DIST
	consider := func(distance int32) {
		if distance != 0 && distance < minDist {
			minDist = distance
		}
	}

	if FTEST(pObj, FLOW) != 0 {
		if pred := FGETPRED(pObj); pred != nil {
			consider(flowData(pred).e_dist)
		}
	} else {
		consider(flowData(pObj).e_dist)
	}

	if pManMR.fIsForward != 0 {
		for i := int32(0); i < abc.Abc_ObjFaninNum(pObj); i++ {
			next := abc.Abc_ObjFanin(pObj, i)
			if abc.Abc_ObjIsLatch(next) == 0 {
				consider(flowData(next).r_dist)
			}
		}
		if pManMR.maxDelay != 0 {
			edges := FTIMEEDGES(pObj)
			for i := 0; i < vec.Vec_PtrSize(edges); i++ {
				next, ok := vec.Vec_PtrEntry(edges, i).(*abc.Abc_Obj_t)
				if !ok || next == nil {
					panic("retiming timing edge contains a non-object")
				}
				consider(flowData(next).r_dist)
			}
		}
	}

	minDist++
	if minDist >= MAX_DIST {
		minDist = 0
	}
	flowData(pObj).r_dist = minDist
	updateSinkDistanceHistogram(oldDist, minDist)
}

func dfsplain_e(pObj *abc.Abc_Obj_t, pPred *abc.Abc_Obj_t) int32 {
	if flowData(pObj).mark&BLOCK_OR_CONS&pManMR.constraintMask != 0 || abc.Abc_ObjIsPi(pObj) != 0 {
		assert(pPred != nil, "pPred")
		assert(pManMR.fIsForward == 0, "!pManMR.fIsForward")
		return 1
	}

	FSET(pObj, VISITED_E)
	if pManMR.fIsForward != 0 {
		for i := int32(0); i < abc.Abc_ObjFanoutNum(pObj); i++ {
			next := abc.Abc_ObjFanout(pObj, i)
			if FTEST(next, VISITED_R) == 0 && dfsplain_r(next, pPred) != 0 {
				return 1
			}
		}
	} else {
		for i := int32(0); i < abc.Abc_ObjFaninNum(pObj); i++ {
			next := abc.Abc_ObjFanin(pObj, i)
			if FTEST(next, VISITED_R) == 0 && dfsplain_r(next, pPred) != 0 {
				return 1
			}
		}
	}

	if abc.Abc_ObjIsLatch(pObj) != 0 {
		return 0
	}
	if pManMR.fIsForward == 0 {
		for i := int32(0); i < abc.Abc_ObjFanoutNum(pObj); i++ {
			next := abc.Abc_ObjFanout(pObj, i)
			if FTEST(next, VISITED_E) == 0 && dfsplain_e(next, pPred) != 0 {
				return 1
			}
		}
		if pManMR.maxDelay != 0 {
			edges := FTIMEEDGES(pObj)
			for i := 0; i < vec.Vec_PtrSize(edges); i++ {
				next, ok := vec.Vec_PtrEntry(edges, i).(*abc.Abc_Obj_t)
				if !ok || next == nil {
					panic("retiming timing edge contains a non-object")
				}
				if FTEST(next, VISITED_E) == 0 && dfsplain_e(next, pPred) != 0 {
					return 1
				}
			}
		}
	}

	if FTEST(pObj, FLOW) != 0 &&
		FTEST(pObj, VISITED_R) == 0 &&
		dfsplain_r(pObj, FGETPRED(pObj)) != 0 {
		FUNSET(pObj, FLOW)
		FSETPRED(pObj, nil)
		return 1
	}
	return 0
}

func dfsplain_r(pObj *abc.Abc_Obj_t, pPred *abc.Abc_Obj_t) int32 {
	if abc.Abc_ObjIsLatch(pObj) != 0 ||
		(pManMR.fIsForward != 0 && abc.Abc_ObjIsPo(pObj) != 0) ||
		(pManMR.fIsForward != 0 && flowData(pObj).mark&BLOCK_OR_CONS&pManMR.constraintMask != 0) {
		assert(pPred != nil, "pPred")
		return 1
	}

	FSET(pObj, VISITED_R)
	if FTEST(pObj, FLOW) != 0 {
		oldPred := FGETPRED(pObj)
		if oldPred != nil &&
			FTEST(oldPred, VISITED_E) == 0 &&
			dfsplain_e(oldPred, oldPred) != 0 {
			FSETPRED(pObj, pPred)
			return 1
		}
	} else if FTEST(pObj, VISITED_E) == 0 && dfsplain_e(pObj, pObj) != 0 {
		FSET(pObj, FLOW)
		FSETPRED(pObj, pPred)
		return 1
	}

	if pManMR.fIsForward != 0 {
		for i := int32(0); i < abc.Abc_ObjFaninNum(pObj); i++ {
			next := abc.Abc_ObjFanin(pObj, i)
			if FTEST(next, VISITED_R) == 0 &&
				abc.Abc_ObjIsLatch(next) == 0 &&
				dfsplain_r(next, pPred) != 0 {
				return 1
			}
		}
		if pManMR.maxDelay != 0 {
			edges := FTIMEEDGES(pObj)
			for i := 0; i < vec.Vec_PtrSize(edges); i++ {
				next, ok := vec.Vec_PtrEntry(edges, i).(*abc.Abc_Obj_t)
				if !ok || next == nil {
					panic("retiming timing edge contains a non-object")
				}
				if FTEST(next, VISITED_R) == 0 && dfsplain_r(next, pPred) != 0 {
					return 1
				}
			}
		}
	}
	return 0
}

//ABC_NAMESPACE_IMPL_END
