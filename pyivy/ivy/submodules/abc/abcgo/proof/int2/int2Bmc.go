package int2

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
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


/**Function*************************************************************

  Synopsis    [Trasnforms AIG to transition into the init state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManDupInit(p *Gia_Man_t, fVerbose int32) *Gia_Man_t {
	assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0")
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
			lit := gia.Gia_ManHashAnd(pNew, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj))
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
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Returns 1 if AIG has transition into init state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManCheckInit(p *Gia_Man_t) int32 {
    var pSat *sat_solver;
    var pCnf *Cnf_Dat_t;
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var vLits *Vec_Int_t;
    var i, Lit, RetValue int32;
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    pNew = Jf_ManDeriveCnf( p, 0 );  
    pCnf = (Cnf_Dat_t *)pNew.pData; pNew.pData = nil;
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    if pSat != nil {
        vLits = Vec_IntAlloc( Gia_ManRegNum(p) );
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(pNew)) && ((pObj) = Gia_ManCo(pNew, Gia_ManPoNum(pNew)+i)); i++  {
            Lit = pCnf.pVarNums[ Gia_ObjId(pNew, Gia_ObjFanin0(pObj)) ];
            Lit = Abc_Var2Lit( Lit, Gia_ObjFaninC0(pObj) );
            Vec_IntPush( vLits, Abc_LitNot(Lit) );
        }
        if sat_solver_solve( pSat, Vec_IntArray(vLits), Vec_IntArray(vLits) + Vec_IntSize(vLits), 0, 0, 0, 0 ) == l_True {
            RetValue = 1; }
        Vec_IntFree( vLits );
        sat_solver_delete( pSat );
    }
    Cnf_DataFree( pCnf );
    Gia_ManStop( pNew );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Creates the BMC instance in the SAT solver.]

  Description [The PIs are mapped in the natural order. The flop inputs
  are the last Gia_ManRegNum(p) variables of resulting SAT solver.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
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
				value := gia.Gia_ManHashAnd(pFrames, gia.Gia_ObjFanin0Copy(obj), gia.Gia_ObjFanin1Copy(obj))
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
    return pFrames;
}
func Int2_ManSetupBmcSolver(p *Gia_Man_t, nFrames int32) *sat_solver {
    var pFrames, pTemp *Gia_Man_t;
    var pCnf *Cnf_Dat_t;
    var pSat *sat_solver;
    // unfold for the given number of timeframes
    pFrames = Int2_ManFrameInit( p, nFrames, 1 );
    assert(Gia_ManRegNum(pFrames) == 0, "Gia_ManRegNum(pFrames) == 0");
    // derive CNF for the timeframes
    pFrames = Jf_ManDeriveCnf( pTemp = pFrames, 0 );  Gia_ManStop( pTemp );
    pCnf = (Cnf_Dat_t *)pFrames.pData; pFrames.pData = nil;
    // create SAT solver
    pSat = (sat_solver *)Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    if pSat != nil {
        var pObj *Gia_Obj_t;
        var i int32;
        var nVars int32 = sat_solver_nvars( pSat );
        sat_solver_setnvars( pSat, nVars + Gia_ManPoNum(pFrames) );
        // add clauses for the POs
        // inlining of macro Gia_ManForEachCo
        for  i = 0; (i < Vec_IntSize(pFrames.vCos)) && ((pObj) = Gia_ManCo(pFrames, i)); i++  {
            sat_solver_add_buffer( pSat, nVars + i, pCnf.pVarNums[Gia_ObjId(pFrames, Gia_ObjFanin0(pObj))], Gia_ObjFaninC0(pObj) );
        }
    }
    Cnf_DataFree( pCnf );
    Gia_ManStop( pFrames );
    return pSat;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Int2_ManCheckFrames(p *Int2_Man_t, iFrame int32, iObj int32) int32 {
    var vMapFrame *Vec_Int_t  = (Vec_Int_t *)Vec_PtrEntry(p.vMapFrames, iFrame);
    return Vec_IntEntry(vMapFrame, iObj);
}
func Int2_ManWriteFrames(p *Int2_Man_t, iFrame int32, iObj int32, iRes int32) {
    var vMapFrame *Vec_Int_t  = (Vec_Int_t *)Vec_PtrEntry(p.vMapFrames, iFrame);
    assert(Vec_IntEntry(vMapFrame, iObj) == -1, "Vec_IntEntry(vMapFrame, iObj) == -1");
    Vec_IntWriteEntry( vMapFrame, iObj, iRes );
}
func Int2_ManCreateFrames(p *Int2_Man_t, iFrame int32, vPrefCos *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i, Entry, iLit int32;
    // create storage room for unfolded IDs
    for  i = Vec_PtrSize(p.vMapFrames); i <= iFrame; i++  {
        Vec_PtrPush( p.vMapFrames, Vec_IntStartFull( Gia_ManObjNum(p.pGia) ) );
    }
    assert(Vec_PtrSize(p.vMapFrames) == iFrame + 1, "Vec_PtrSize(p.vMapFrames) == iFrame + 1");
    // create constant 0 node
    if f == 0 {
        iLit = 1;
        Int2_ManWriteFrames( p, iFrame, iObj, 0 );
        sat_solver_addclause( p.pGiaPref, &iLit, &iLit + 1 );
    }
    // start the stack
    Vec_IntClear( p.vStack );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vPrefCos)) && (((Entry) = Vec_IntEntry(vPrefCos, i)), 1); i++  {
        pObj = Gia_ManCo( p.pGia, Entry );
        Vec_IntPush( p.vStack, iFrame );
        Vec_IntPush( p.vStack, Gia_ObjId(p.pGia, pObj) );
    }
    // construct unfolded AIG
    for  Vec_IntSize(p.vStack) > 0  {
        var iObj int32    = Vec_IntPop(p.vStack);
        var iFrame int32  = Vec_IntPop(p.vStack);
        if Int2_ManCheckFrames(p, iFrame, iObj) >= 0 {
            continue; }
        pObj = Gia_ManObj( p.pGia, iObj );
        if Gia_ObjIsPi(p.pGia, pObj) {
            Int2_ManWriteFrames( p, iFrame, iObj, Gia_ManAppendCi(p.pFrames) );
        } else if iFrame == 0 && Gia_ObjIsRo(p.pGia, iObj) {
            Int2_ManWriteFrames( p, iFrame, iObj, 0 );
        } else if Gia_ObjIsRo(p.pGia, iObj) {
            var iObjF int32  = Gia_ObjId( p.pGia, Gia_ObjRoToRi(p.pGia, pObj) );
            var iLit int32  = Int2_ManCheckFrames( p, iFrame-1, iObjF );
            if iLit >= 0 {
                Int2_ManWriteFrames( p, iFrame, iObj, iLit );
            } else {
                Vec_IntPush( p.vStack, iFrame );
                Vec_IntPush( p.vStack, iObj );
                Vec_IntPush( p.vStack, iFrame-1 );
                Vec_IntPush( p.vStack, iObjF );
            }
        
        } else if Gia_ObjIsCo(pObj) {
            var iObjF int32  = Gia_ObjFaninId0(p.pGia, iObj);
            var iLit int32  = Int2_ManCheckFrames( p, iFrame, iObjF );
            if iLit >= 0 {
                Int2_ManWriteFrames( p, iFrame, iObj, Abc_LitNotCond(iLit, Gia_ObjFaninC0(pObj)) );
            } else {
                Vec_IntPush( p.vStack, iFrame );
                Vec_IntPush( p.vStack, iObj );
                Vec_IntPush( p.vStack, iFrame );
                Vec_IntPush( p.vStack, iObjF );
            }
        
        } else if Gia_ObjIsAnd(pObj) {
            var iObjF0 int32  = Gia_ObjFaninId0(p.pGia, iObj);
            var iLit0 int32  = Int2_ManCheckFrames( p, iFrame, iObjF0 );
            var iObjF1 int32  = Gia_ObjFaninId1(p.pGia, iObj);
            var iLit1 int32  = Int2_ManCheckFrames( p, iFrame, iObjF1 );
            if iLit0 >= 0 && iLit1 >= 0 {
                Entry = Gia_ManObjNum(pFrames);
                iLit = Gia_ManHashAnd(pFrames, iLit0, iLit1);
                Int2_ManWriteFrames( p, iFrame, iObj, iLit );
                if Entry < Gia_ManObjNum(pFrames) {
                    assert(!Abc_LitIsCompl(iLit), "!Abc_LitIsCompl(iLit)");
                    sat_solver_add_and( p.pGiaPref, Abc_Lit2Var(iLit), Abc_Lit2Var(iLit0), Abc_Lit2Var(iLit1), Abc_LitIsCompl(iLit0), Abc_LitIsCompl(iLit1), 0 ); 
                }
            
            } else {
                Vec_IntPush( p.vStack, iFrame );
                Vec_IntPush( p.vStack, iObj );
                if iLit0 < 0 {
                    Vec_IntPush( p.vStack, iFrame );
                    Vec_IntPush( p.vStack, iObjF0 );
                }
                if iLit1 < 0 {
                    Vec_IntPush( p.vStack, iFrame );
                    Vec_IntPush( p.vStack, iObjF1 );
                }
            }
        
        } else { assert(false, "0"); }
    }
}
func Int2_ManCheckBmc(p *Int2_Man_t, vCube *Vec_Int_t) int32 {
    var status int32;
    if vCube == nil {
        var pObj *Gia_Obj_t;
        var i, iLit int32;
        // inlining of macro Gia_ManForEachPo
        for  i = 0; (i < Gia_ManPoNum(p.pGia)) && ((pObj) = Gia_ManCo(p.pGia, i)); i++  {
            iLit = Int2_ManCheckFrames( p, 0, Gia_ObjId(p.pGia, pObj) );
            if iLit == 0 {
                continue; }
            if iLit == 1 {
                return 0; }
            status = sat_solver_solve( p.pSatPref, &iLit, &iLit + 1, 0, 0, 0, 0 );
            if status == l_False {
                continue; }
            if status == l_True {
                return 0; }
            return -1;
        }
        return 1;
    }
    status = sat_solver_solve( p.pSatPref, Vec_IntArray(vCube), Vec_IntArray(vCube) + Vec_IntSize(vCube), 0, 0, 0, 0 );
    if status == l_False {
        return 1; }
    if status == l_True {
        return 0; }
    return -1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

