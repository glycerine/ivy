package int2

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

/**CFile****************************************************************

  FileName    [int2Bmc.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Interpolation engine.]

  Synopsis    [BMC used inside IMC.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 1, 2013.]

  Revision    [$Id: int2Bmc.c,v 1.00 2013/12/01 00:00:00 alanmi Exp $]

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

	Synopsis    [Trasnforms AIG to transition into the init state.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Int2_ManDupInit(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
	assert(gia.Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0")
	pNew := gia.Gia_ManStart(10000)
	pNew.CopyIdentityFrom(p)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)

	var iCtrl int32
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		obj := gia.Gia_ManCi(p, i)
		if i == gia.Gia_ManPiNum(p) {
			iCtrl = gia.Gia_ManAppendCi(pNew)
		}
		gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCi(pNew))
	}

	gia.Gia_ManHashAlloc(pNew)
	for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
		obj := gia.Gia_ManObj(p, i)
		if gia.Gia_ObjIsAnd(obj) != 0 {
			lit := int2HashCopyGate(pNew, p, i)
			gia.Gia_ObjSetValue(obj, lit)
		}
	}
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		obj := gia.Gia_ManPo(p, i)
		gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCo(pNew, gia.Gia_ObjFanin0Copy(obj)))
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		ri := gia.Gia_ManRi(p, i)
		ro := gia.Gia_ManRo(p, i)
		lit := gia.Gia_ManHashMux(pNew, iCtrl, gia.Gia_ObjValue(ro), gia.Gia_ObjFanin0Copy(ri))
		gia.Gia_ManAppendCo(pNew, lit)
	}
	gia.Gia_ManHashStop(pNew)
	gia.Gia_ManSetRegNum(pNew, gia.Gia_ManRegNum(p))
	// Remove dangling nodes after the initialization transition is added.
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	if fVerbose != 0 {
		fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", gia.Gia_ManAndNum(pTemp), gia.Gia_ManAndNum(pNew))
	}
	gia.Gia_ManStop(pTemp)
	return pNew
}

/*
*Function*************************************************************

	Synopsis    [Returns 1 if AIG has transition into init state.]

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Int2_ManCheckInit(p *Gia_Man_t) int32 {
	assert(gia.Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0")
	pNew := gia.Jf_ManDeriveCnf(p, 0)
	cnf, ok := pNew.TakeData().(*Cnf_Dat_t)
	if !ok || cnf == nil {
		gia.Gia_ManStop(pNew)
		return 0
	}
	solver := sat_solver_from_clauses(int32(cnf.Variables), cnf.Clauses)
	if solver == nil {
		gia.Gia_ManStop(pNew)
		return 0
	}
	assumptions := make([]int32, 0, gia.Gia_ManRegNum(p))
	for i := int32(0); i < gia.Gia_ManRegNum(pNew); i++ {
		obj := gia.Gia_ManRi(pNew, i)
		faninID := gia.Gia_ObjFaninId0p(pNew, obj)
		variable := cnf.VarNums[faninID]
		assumptions = append(assumptions, sat_literal(variable, gia.Gia_ObjFaninC0(obj))^1)
	}
	status := sat_solver_solve(solver, assumptions, 0, 0, 0, 0)
	gia.Gia_ManStop(pNew)
	if status == l_True {
		return 1
	}
	return 0
}

/*
*Function*************************************************************

	Synopsis    [Creates the BMC instance in the SAT solver.]

	Description [The PIs are mapped in the natural order. The flop inputs
	are the last Gia_ManRegNum(p) variables of resulting SAT solver.]

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Int2_ManFrameInit(p *Gia_Man_t, nFrames int32, fVerbose int32) *Gia_Man_t {
	pFrames := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	pFrames.CopyIdentityFrom(p)
	gia.Gia_ObjSetValue(gia.Gia_ManConst0(p), 0)
	gia.Gia_ManHashAlloc(pFrames)
	for frame := int32(0); frame < nFrames; frame++ {
		for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
			obj := gia.Gia_ManPi(p, i)
			gia.Gia_ObjSetValue(obj, gia.Gia_ManAppendCi(pFrames))
		}
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			ro := gia.Gia_ManRo(p, i)
			value := int32(0)
			if frame != 0 {
				value = gia.Gia_ObjValue(gia.Gia_ObjRoToRi(p, ro))
			}
			gia.Gia_ObjSetValue(ro, value)
		}
		for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
			obj := gia.Gia_ManObj(p, i)
			if gia.Gia_ObjIsAnd(obj) != 0 {
				value := int2HashCopyGate(pFrames, p, i)
				gia.Gia_ObjSetValue(obj, value)
			}
		}
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			ri := gia.Gia_ManRi(p, i)
			gia.Gia_ObjSetValue(ri, gia.Gia_ObjFanin0Copy(ri))
		}
	}
	gia.Gia_ManHashStop(pFrames)
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		ri := gia.Gia_ManRi(p, i)
		gia.Gia_ObjSetValue(ri, gia.Gia_ManAppendCo(pFrames, gia.Gia_ObjFanin0Copy(ri)))
	}
	pTemp := pFrames
	pFrames = gia.Gia_ManCleanup(pTemp)
	if fVerbose != 0 {
		fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", gia.Gia_ManAndNum(pTemp), gia.Gia_ManAndNum(pFrames))
	}
	gia.Gia_ManStop(pTemp)
	return pFrames
}
func Int2_ManSetupBmcSolver(p *Gia_Man_t, nFrames int32) *sat_solver {
	pFrames := Int2_ManFrameInit(p, nFrames, 1)
	assert(gia.Gia_ManRegNum(pFrames) == 0, "Gia_ManRegNum(pFrames) == 0")
	pNew := gia.Jf_ManDeriveCnf(pFrames, 0)
	gia.Gia_ManStop(pFrames)
	cnf, ok := pNew.TakeData().(*Cnf_Dat_t)
	if !ok || cnf == nil {
		gia.Gia_ManStop(pNew)
		return nil
	}
	solver := sat_solver_from_clauses(int32(cnf.Variables), cnf.Clauses)
	if solver != nil {
		firstOutputVar := sat_solver_nvars(solver)
		sat_solver_setnvars(solver, firstOutputVar+gia.Gia_ManCoNum(pNew))
		for i := int32(0); i < gia.Gia_ManCoNum(pNew); i++ {
			obj := gia.Gia_ManCo(pNew, i)
			faninID := gia.Gia_ObjFaninId0p(pNew, obj)
			faninVar := cnf.VarNums[faninID]
			if sat_solver_add_buffer(solver, firstOutputVar+i, faninVar, gia.Gia_ObjFaninC0(obj)) == 0 {
				solver = nil
				break
			}
		}
	}
	gia.Gia_ManStop(pNew)
	return solver
}

/*
*Function*************************************************************

	Synopsis    []

	Description []

	SideEffects []

	SeeAlso     []

**********************************************************************
*/
func Int2_ManCheckFrames(p *Int2_Man_t, iFrame int32, iObj int32) int32 {
	if p == nil || p.vMapFrames == nil || iFrame < 0 || iFrame >= int32(vec.Vec_PtrSize(p.vMapFrames)) {
		return -1
	}
	frame, ok := vec.Vec_PtrEntry(p.vMapFrames, int(iFrame)).(*vec.Vec_Int_t)
	if !ok || frame == nil || iObj < 0 || iObj >= int32(vec.Vec_IntSize(frame)) {
		return -1
	}
	return vec.Vec_IntEntry(frame, int(iObj))
}
func Int2_ManWriteFrames(p *Int2_Man_t, iFrame int32, iObj int32, iRes int32) {
	assert(p != nil && p.vMapFrames != nil && iFrame >= 0 && iFrame < int32(vec.Vec_PtrSize(p.vMapFrames)), "frame index is in range")
	frame, ok := vec.Vec_PtrEntry(p.vMapFrames, int(iFrame)).(*vec.Vec_Int_t)
	assert(ok && frame != nil && iObj >= 0 && iObj < int32(vec.Vec_IntSize(frame)), "object index is in range")
	assert(vec.Vec_IntEntry(frame, int(iObj)) == -1, "frame entry is unwritten")
	vec.Vec_IntWriteEntry(frame, int(iObj), iRes)
}
func Int2_ManCreateFrames(p *Int2_Man_t, iFrame int32, vPrefCos *Vec_Int_t) {
	assert(p != nil && p.pGia != nil && p.pGiaPref != nil && p.pSatPref != nil, "manager and prefix state exist")
	assert(iFrame >= 0, "frame index is nonnegative")
	for i := int32(vec.Vec_PtrSize(p.vMapFrames)); i <= iFrame; i++ {
		frame := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p.pGia)))
		vec.Vec_IntWriteEntry(frame, 0, 0)
		vec.Vec_PtrPush(p.vMapFrames, frame)
	}
	assert(int32(vec.Vec_PtrSize(p.vMapFrames)) == iFrame+1, "frame maps are contiguous")
	if iFrame == 0 {
		if sat_solver_addclause(p.pSatPref, []int32{sat_literal(0, 1)}) == 0 {
			panic("int2: constant-zero SAT unit is inconsistent")
		}
	}
	vec.Vec_IntClear(p.vStack)
	// Keep the SAT prefix synchronized with every AND created while lowering
	// native XOR and MUX nodes. Hashing may return an existing or constant lit.
	hashAnd := func(a, b int32) int32 {
		before := gia.Gia_ManObjNum(p.pGiaPref)
		lit := gia.Gia_ManHashAnd(p.pGiaPref, a, b)
		if gia.Gia_ManObjNum(p.pGiaPref) > before {
			assert(lit&1 == 0, "new AND literal is regular")
			if sat_solver_add_and(p.pSatPref, lit>>1, a>>1, b>>1, a&1, b&1, 0) == 0 {
				panic("int2: SAT solver rejected an AND constraint")
			}
		}
		return lit
	}
	for i := 0; i < vec.Vec_IntSize(vPrefCos); i++ {
		coID := vec.Vec_IntEntry(vPrefCos, i)
		co := gia.Gia_ManCo(p.pGia, coID)
		vec.Vec_IntPush(p.vStack, iFrame)
		vec.Vec_IntPush(p.vStack, gia.Gia_ObjId(p.pGia, co))
	}
	for vec.Vec_IntSize(p.vStack) > 0 {
		iObj := vec.Vec_IntPop(p.vStack)
		frame := vec.Vec_IntPop(p.vStack)
		if Int2_ManCheckFrames(p, frame, iObj) >= 0 {
			continue
		}
		obj := gia.Gia_ManObj(p.pGia, iObj)
		switch {
		case gia.Gia_ObjIsPi(p.pGia, obj) != 0:
			lit := gia.Gia_ManAppendCi(p.pGiaPref)
			sat_solver_setnvars(p.pSatPref, (lit>>1)+1)
			Int2_ManWriteFrames(p, frame, iObj, lit)
		case gia.Gia_ObjIsRo(p.pGia, obj) != 0 && frame == 0:
			Int2_ManWriteFrames(p, frame, iObj, 0)
		case gia.Gia_ObjIsRo(p.pGia, obj) != 0:
			previousID := gia.Gia_ObjId(p.pGia, gia.Gia_ObjRoToRi(p.pGia, obj))
			previousLit := Int2_ManCheckFrames(p, frame-1, previousID)
			if previousLit >= 0 {
				Int2_ManWriteFrames(p, frame, iObj, previousLit)
			} else {
				vec.Vec_IntPush(p.vStack, frame)
				vec.Vec_IntPush(p.vStack, iObj)
				vec.Vec_IntPush(p.vStack, frame-1)
				vec.Vec_IntPush(p.vStack, previousID)
			}
		case gia.Gia_ObjIsCo(obj) != 0:
			faninID := gia.Gia_ObjFaninId0p(p.pGia, obj)
			faninLit := Int2_ManCheckFrames(p, frame, faninID)
			if faninLit >= 0 {
				Int2_ManWriteFrames(p, frame, iObj, faninLit^gia.Gia_ObjFaninC0(obj))
			} else {
				vec.Vec_IntPush(p.vStack, frame)
				vec.Vec_IntPush(p.vStack, iObj)
				vec.Vec_IntPush(p.vStack, frame)
				vec.Vec_IntPush(p.vStack, faninID)
			}
		case gia.Gia_ObjIsAnd(obj) != 0:
			a, b, c, kind := p.pGia.Gate(iObj)
			fanins := []int32{a, b}
			if kind == 4 {
				fanins = append(fanins, c)
			}
			if kind == 5 {
				fanins = fanins[:1]
			}
			mapped := make([]int32, len(fanins))
			ready := true
			for i, fanin := range fanins {
				mapped[i] = Int2_ManCheckFrames(p, frame, fanin>>1)
				if mapped[i] < 0 {
					ready = false
				} else {
					mapped[i] ^= fanin & 1
				}
			}
			if !ready {
				vec.Vec_IntPush(p.vStack, frame)
				vec.Vec_IntPush(p.vStack, iObj)
				for i, fanin := range fanins {
					if mapped[i] < 0 {
						vec.Vec_IntPush(p.vStack, frame)
						vec.Vec_IntPush(p.vStack, fanin>>1)
					}
				}
				continue
			}
			var lit int32
			switch kind {
			case 2:
				lit = hashAnd(mapped[0], mapped[1])
			case 3:
				x := hashAnd(mapped[0], mapped[1]^1)
				y := hashAnd(mapped[0]^1, mapped[1])
				lit = hashAnd(x^1, y^1) ^ 1
			case 4:
				x := hashAnd(mapped[0], mapped[1])
				y := hashAnd(mapped[0]^1, mapped[2])
				lit = hashAnd(x^1, y^1) ^ 1
			case 5:
				lit = mapped[0]
			default:
				panic("int2: unsupported internal GIA object")
			}
			Int2_ManWriteFrames(p, frame, iObj, lit)
		default:
			panic("int2: unsupported GIA object in frame construction")
		}
	}
}

func Int2_ManCheckBmc(p *Int2_Man_t, vCube *Vec_Int_t) int32 {
	assert(p != nil && p.pGia != nil && p.pSatPref != nil, "BMC manager is initialized")
	if vCube == nil {
		for i := int32(0); i < gia.Gia_ManPoNum(p.pGia); i++ {
			obj := gia.Gia_ManPo(p.pGia, i)
			lit := Int2_ManCheckFrames(p, 0, gia.Gia_ObjId(p.pGia, obj))
			if lit == 0 {
				continue
			}
			if lit == 1 {
				return 0
			}
			status := sat_solver_solve(p.pSatPref, []int32{lit}, 0, 0, 0, 0)
			switch status {
			case l_False:
				continue
			case l_True:
				return 0
			default:
				return -1
			}
		}
		return 1
	}
	status := sat_solver_solve(p.pSatPref, vec.Vec_IntArray(vCube), 0, 0, 0, 0)
	if status == l_False {
		return 1
	}
	if status == l_True {
		return 0
	}
	return -1
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END
