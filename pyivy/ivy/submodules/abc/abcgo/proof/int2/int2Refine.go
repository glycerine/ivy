package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [int2Refine.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 1, 2013.]

  Revision    [$Id: int2Refine.c,v 1.00 2013/12/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "int2Int.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Int2_ManJustify_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vSelect *Vec_Int_t) {
	if gia.Gia_ObjMark1(pObj) != 0 {
		return
	}
	gia.Gia_ObjSetMark1(pObj, 1)
	if gia.Gia_ManRegNum(p) > 0 && gia.Gia_ObjIsPi(p, pObj) != 0 {
		return
	}
	if gia.Gia_ObjIsCi(pObj) != 0 {
		vec.Vec_IntPush(vSelect, gia.Gia_ObjCioId(pObj))
		return
	}
	assert(gia.Gia_ObjIsAnd(pObj) != 0, "Gia_ObjIsAnd(pObj)")
	fanin0 := gia.Gia_ObjFanin0(pObj)
	fanin1 := gia.Gia_ObjFanin1(pObj)
	if gia.Gia_ObjMark0(pObj) == 1 {
		if gia.Gia_ObjValue(fanin0) < util.ABC_INFINITY {
			Int2_ManJustify_rec(p, fanin0, vSelect)
		}
		if gia.Gia_ObjValue(fanin1) < util.ABC_INFINITY {
			Int2_ManJustify_rec(p, fanin1, vSelect)
		}
		return
	}
	mark0 := gia.Gia_ObjMark0(fanin0)
	mark1 := gia.Gia_ObjMark0(fanin1)
	need0 := mark0 ^ gia.Gia_ObjFaninC0(pObj)
	need1 := mark1 ^ gia.Gia_ObjFaninC1(pObj)
	switch {
	case need0 == 0 && need1 == 0:
		if gia.Gia_ObjValue(fanin0) <= gia.Gia_ObjValue(fanin1) {
			if gia.Gia_ObjValue(fanin0) < util.ABC_INFINITY {
				Int2_ManJustify_rec(p, fanin0, vSelect)
			}
		} else if gia.Gia_ObjValue(fanin1) < util.ABC_INFINITY {
			Int2_ManJustify_rec(p, fanin1, vSelect)
		}
	case need0 == 0:
		if gia.Gia_ObjValue(fanin0) < util.ABC_INFINITY {
			Int2_ManJustify_rec(p, fanin0, vSelect)
		}
	case need1 == 0:
		if gia.Gia_ObjValue(fanin1) < util.ABC_INFINITY {
			Int2_ManJustify_rec(p, fanin1, vSelect)
		}
	default:
		assert(false, "AND assignment has no satisfying fanin")
	}
}

/*
*Function*************************************************************

	Synopsis    [Computes the reduced set of flop variables.]

	Description [Given is a single-output seq AIG manager and an assignment
	of its CIs.  Returned is a subset of flops that justifies the output.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Int2_ManRefineCube(p *Gia_Man_t, vAssign *Vec_Int_t, vPrio *Vec_Int_t) *Vec_Int_t {
	// set values and prios
	assert(vec.Vec_IntSize(vAssign) == vec.Vec_IntSize(vPrio), "assignment and priority sizes match")
	assert(vec.Vec_IntSize(vAssign) == int(gia.Gia_ManCiNum(p)), "assignment covers every CI")
	constant := gia.Gia_ManConst0(p)
	gia.Gia_ObjSetMark0(constant, 0)
	gia.Gia_ObjSetMark1(constant, 0)
	gia.Gia_ObjSetValue(constant, util.ABC_INFINITY)
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		obj := gia.Gia_ManCi(p, i)
		gia.Gia_ObjSetMark0(obj, vec.Vec_IntEntry(vAssign, int(i)))
		gia.Gia_ObjSetMark1(obj, 0)
		gia.Gia_ObjSetValue(obj, vec.Vec_IntEntry(vPrio, int(i)))
	}
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		obj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		fanin0 := gia.Gia_ObjFanin0(obj)
		fanin1 := gia.Gia_ObjFanin1(obj)
		value0 := gia.Gia_ObjMark0(fanin0) ^ gia.Gia_ObjFaninC0(obj)
		value1 := gia.Gia_ObjMark0(fanin1) ^ gia.Gia_ObjFaninC1(obj)
		gia.Gia_ObjSetMark0(obj, value0&value1)
		gia.Gia_ObjSetMark1(obj, 0)
		switch {
		case value0 == 1 && value1 == 1:
			gia.Gia_ObjSetValue(obj, util.Abc_MaxInt(gia.Gia_ObjValue(fanin0), gia.Gia_ObjValue(fanin1)))
		case value0 == 0 && value1 == 0:
			gia.Gia_ObjSetValue(obj, util.Abc_MinInt(gia.Gia_ObjValue(fanin0), gia.Gia_ObjValue(fanin1)))
		case value0 == 0:
			gia.Gia_ObjSetValue(obj, gia.Gia_ObjValue(fanin0))
		default:
			gia.Gia_ObjSetValue(obj, gia.Gia_ObjValue(fanin1))
		}
	}
	obj := gia.Gia_ManPo(p, 0)
	fanin := gia.Gia_ObjFanin0(obj)
	gia.Gia_ObjSetMark0(obj, gia.Gia_ObjMark0(fanin)^gia.Gia_ObjFaninC0(obj))
	gia.Gia_ObjSetMark1(obj, 0)
	gia.Gia_ObjSetValue(obj, gia.Gia_ObjValue(fanin))
	assert(gia.Gia_ObjMark0(obj) == 1, "output is asserted by the assignment")
	// A constant-true output is justified by the empty cube.
	if fanin == constant {
		return vec.Vec_IntAlloc(0)
	}
	assert(gia.Gia_ObjValue(obj) < util.ABC_INFINITY, "output has finite priority")
	// select subset
	vSubset := vec.Vec_IntAlloc(100)
	Int2_ManJustify_rec(p, fanin, vSubset)
	return vSubset
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
