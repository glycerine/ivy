package vec

/**CFile****************************************************************

  FileName    [vecAtt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resizable arrays.]

  Synopsis    [Array of user-specified attiributes.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: vecAtt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

// various attributes
type Vec_AttrType_t int32

const (
	VEC_ATTR_NONE Vec_AttrType_t = 0
	// 0
	VEC_ATTR_COPY Vec_AttrType_t = VEC_ATTR_NONE + 1
	// 1
	VEC_ATTR_LOCAL_AIG Vec_AttrType_t = VEC_ATTR_COPY + 1
	// 2
	VEC_ATTR_LOCAL_SOP Vec_AttrType_t = VEC_ATTR_LOCAL_AIG + 1
	// 3
	VEC_ATTR_LOCAL_BDD Vec_AttrType_t = VEC_ATTR_LOCAL_SOP + 1
	// 4
	VEC_ATTR_GLOBAL_AIG Vec_AttrType_t = VEC_ATTR_LOCAL_BDD + 1
	// 5
	VEC_ATTR_GLOBAL_SOP Vec_AttrType_t = VEC_ATTR_GLOBAL_AIG + 1
	// 6
	VEC_ATTR_GLOBAL_BDD Vec_AttrType_t = VEC_ATTR_GLOBAL_SOP + 1
	// 7
	VEC_ATTR_LEVEL Vec_AttrType_t = VEC_ATTR_GLOBAL_BDD + 1
	// 8
	VEC_ATTR_LEVEL_REV Vec_AttrType_t = VEC_ATTR_LEVEL + 1
	// 9
	VEC_ATTR_RETIME_LAG Vec_AttrType_t = VEC_ATTR_LEVEL_REV + 1
	// 10
	VEC_ATTR_FRAIG Vec_AttrType_t = VEC_ATTR_RETIME_LAG + 1
	// 11
	VEC_ATTR_MVVAR Vec_AttrType_t = VEC_ATTR_FRAIG + 1
	// 12
	VEC_ATTR_DATA1 Vec_AttrType_t = VEC_ATTR_MVVAR + 1
	// 13
	VEC_ATTR_DATA2 Vec_AttrType_t = VEC_ATTR_DATA1 + 1
	// 14
	VEC_ATTR_TOTAL_NUM Vec_AttrType_t = VEC_ATTR_DATA2 + 1 // 15

)

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// Vec_Att_t stores optional attributes and their ownership callbacks.
type Vec_Att_t struct {
	pArrayPtr     []any
	pMan          any
	pFuncFreeMan  func(any)
	pFuncStartObj func(any) any
	pFuncFreeObj  func(any, any)
}

func Vec_AttAlloc(nSize int, manager any, freeManager func(any), start func(any) any, freeObject func(any, any)) *Vec_Att_t {
	if nSize == 0 {
		nSize = 16
	}
	return &Vec_Att_t{pArrayPtr: make([]any, nSize), pMan: manager, pFuncFreeMan: freeManager, pFuncStartObj: start, pFuncFreeObj: freeObject}
}

// Vec_AttFree releases owned attributes and optionally their manager. Otherwise
// ownership of the manager is returned to the caller.
func Vec_AttFree(p *Vec_Att_t, freeManager int) any {
	if p == nil {
		return nil
	}
	Vec_AttClear(p)
	manager := p.pMan
	if !ptrNil(manager) && freeManager != 0 {
		p.pFuncFreeMan(manager)
	}
	p.pArrayPtr = nil
	p.pMan = nil

	if freeManager != 0 {
		return nil
	}
	return manager
}
func Vec_AttClear(p *Vec_Att_t) {
	if p.pFuncFreeObj != nil {
		for _, entry := range p.pArrayPtr {
			if !ptrNil(entry) {
				p.pFuncFreeObj(p.pMan, entry)
			}
		}
	}
	clear(p.pArrayPtr)
}
func Vec_AttFreeEntry(p *Vec_Att_t, i int) {
	if i >= len(p.pArrayPtr) {
		return
	}
	if !ptrNil(p.pMan) && !ptrNil(p.pArrayPtr[i]) && p.pFuncFreeObj != nil {
		p.pFuncFreeObj(p.pMan, p.pArrayPtr[i])
	}
	p.pArrayPtr[i] = nil
}
func Vec_AttGrow(p *Vec_Att_t, minimum int) {
	if len(p.pArrayPtr) >= minimum {
		return
	}
	p.pArrayPtr = append(p.pArrayPtr, make([]any, minimum-len(p.pArrayPtr))...)

}
func Vec_AttWriteEntry(p *Vec_Att_t, i int, entry any) {
	if p.pArrayPtr == nil || p.pFuncStartObj != nil {
		panic("Vec_AttWriteEntry requires initialized unmanaged attributes")
	}
	if i >= len(p.pArrayPtr) {
		Vec_AttGrow(p, max(2*len(p.pArrayPtr), i+10))
	}
	p.pArrayPtr[i] = entry
}
func Vec_AttEntry(p *Vec_Att_t, i int) any {
	if p.pArrayPtr == nil {
		panic("Vec_AttEntry requires initialized attributes")
	}
	if i >= len(p.pArrayPtr) {
		Vec_AttGrow(p, max(2*len(p.pArrayPtr), i+10))
	}
	if ptrNil(p.pArrayPtr[i]) && p.pFuncStartObj != nil {
		p.pArrayPtr[i] = p.pFuncStartObj(p.pMan)
	}
	return p.pArrayPtr[i]
}
func Vec_AttMan(p *Vec_Att_t) any     { return p.pMan }
func Vec_AttArray(p *Vec_Att_t) []any { return p.pArrayPtr }
