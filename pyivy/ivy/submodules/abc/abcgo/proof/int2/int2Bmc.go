package int2

import "fmt"

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
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var i, iCtrl int32;
    assert(Gia_ManRegNum(p) > 0, "Gia_ManRegNum(p) > 0");
    pNew = Gia_ManStart( 10000 );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if i == Gia_ManPiNum(p) {
            iCtrl = Gia_ManAppendCi( pNew ); }
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    Gia_ManHashAlloc( pNew );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    // inlining of macro Gia_ManForEachRiRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Gia_ManAppendCo( pNew, Gia_ManHashMux( pNew, iCtrl, pObjRo.Value, Gia_ObjFanin0Copy(pObjRi) ) );
    }
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    // remove dangling
    pNew = Gia_ManCleanup( pTemp = pNew );
    if fVerbose != 0 {
        fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", Gia_ManAndNum(pTemp), Gia_ManAndNum(pNew)); }
    Gia_ManStop( pTemp );
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
    var pFrames, pTemp *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, f int32;
    pFrames = Gia_ManStart( 10000 );
    pFrames.pName = Abc_UtilStrsav( p.pName );
    pFrames.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // perform structural hashing
    Gia_ManHashAlloc( pFrames );
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.Value = Gia_ManAppendCi( pFrames );
        }
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            if f != 0 {
                pObj.Value = Gia_ObjRoToRi(p, pObj).Value;
            } else {
                pObj.Value = 0;
            }
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.Value = Gia_ManHashAnd( pFrames, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        }
    }
    Gia_ManHashStop( pFrames );
    // create flop inputs
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.Value = Gia_ManAppendCo( pFrames, Gia_ObjFanin0Copy(pObj) );
    }
    // remove dangling
    pFrames = Gia_ManCleanup( pTemp = pFrames );
    if fVerbose != 0 {
        fmt.Printf("Before cleanup = %d nodes. After cleanup = %d nodes.\n", Gia_ManAndNum(pTemp), Gia_ManAndNum(pFrames)); }
    Gia_ManStop( pTemp );
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

