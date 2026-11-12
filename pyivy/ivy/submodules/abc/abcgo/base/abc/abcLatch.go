package abc

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [abcLatch.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Procedures working with latches.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcLatch.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "abc.h"

/**Function*************************************************************

  Synopsis    [Checks if latches form self-loop.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkConverLatchNamesIntoNumbers(pNtk *Abc_Ntk_t) *Vec_Ptr_t {
    var vResult, vNames *Vec_Ptr_t;
    var vNumbers *Vec_Int_t;
    var pObj *Abc_Obj_t;
    var pName string;
    var i, k, Num int32;
    if pNtk.vOnehots == nil {
        return nil; }
    // set register numbers
    Abc_NtkForEachLatch( pNtk, pObj, i )
        pObj.pNext = (Abc_Obj_t *)(ABC_PTRINT_T)i;
    // add the numbers
    vResult = Vec_PtrAlloc( Vec_PtrSize(pNtk.vOnehots) );
    Vec_PtrForEachEntry( Vec_Ptr_t *, pNtk.vOnehots, vNames, i )
    {
        vNumbers = Vec_IntAlloc( Vec_PtrSize(vNames) );
        Vec_PtrForEachEntry( char *, vNames, pName, k )
        {
            Num = Nm_ManFindIdByName( pNtk.pManName, pName, ABC_OBJ_BO );
            if Num < 0 {
                continue; }
            pObj = Abc_NtkObj( pNtk, Num );
            if Abc_ObjFaninNum(pObj) != 1 || !Abc_ObjIsLatch(Abc_ObjFanin0(pObj)) {
                continue; }
            Vec_IntPush( vNumbers, int32((ABC_PTRINT_T))pObj.pNext );
        }
        if Vec_IntSize( vNumbers ) > 1 {
            Vec_PtrPush( vResult, vNumbers );
fmt.Printf("Converted %d one-hot registers.\n", Vec_IntSize(vNumbers));
        
        } else {
            Vec_IntFree( vNumbers ); }
    }
    // clean the numbers
    Abc_NtkForEachLatch( pNtk, pObj, i )
        pObj.pNext = nil;
    return vResult;
}


/**Function*************************************************************

  Synopsis    [Converts registers with DC values into additional PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
//ABC_NAMESPACE_IMPL_END

//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    [Performs retiming with classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeWithClassesAig(pMan *Aig_Man_t, vClasses *Vec_Int_t, pvClasses **Vec_Int_t, fVerbose int32) *Aig_Man_t {
    var pManNew *Aig_Man_t;
    var pGia, pGiaNew *Gia_Man_t;
    pGia = Gia_ManFromAigSimple( pMan );
    assert(Gia_ManRegNum(pGia) == Vec_IntSize(vClasses), "Gia_ManRegNum(pGia) == Vec_IntSize(vClasses)");
    pGia.vFlopClasses = vClasses;
    pGiaNew = Gia_ManRetimeForward( pGia, 10, fVerbose );
    *pvClasses = pGiaNew.vFlopClasses; 
    pGiaNew.vFlopClasses = nil;
    pManNew = Gia_ManToAig( pGiaNew, 0 );
    Gia_ManStop( pGiaNew );
    Gia_ManStop( pGia );
    return pManNew;
}

/**Function*************************************************************

  Synopsis    [Performs retiming with classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkRetimeWithClassesNtk(pNtk *Abc_Ntk_t, vClasses *Vec_Int_t, pvClasses **Vec_Int_t, fVerbose int32) *Abc_Ntk_t {
    extern Aig_Man_t * Abc_NtkToDar( Abc_Ntk_t * pNtk, int fExors, int fRegisters );
    extern Abc_Ntk_t * Abc_NtkFromDarSeqSweep( Abc_Ntk_t * pNtkOld, Aig_Man_t * pMan );
    var pNtkAig, pNtkAigRet, pNtkRes *Abc_Ntk_t;
    var pMan, pManNew *Aig_Man_t;
    pNtkAig    = Abc_NtkStrash( pNtk, 0, 1, 0 );
    pMan       = Abc_NtkToDar( pNtkAig, 0, 1 );
    pManNew    = Abc_NtkRetimeWithClassesAig( pMan, vClasses, pvClasses, fVerbose );
    pNtkAigRet = Abc_NtkFromDarSeqSweep( pNtkAig, pManNew );
    pNtkRes    = Abc_NtkToLogic( pNtkAigRet );
    Abc_NtkDelete( pNtkAigRet );
    Abc_NtkDelete( pNtkAig );
    Aig_ManStop( pManNew );
    Aig_ManStop( pMan );
    return pNtkRes;
}

/**Function*************************************************************

  Synopsis    [Returns self-loops back into the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkTransformBack(pNtkOld *Abc_Ntk_t, pNtkNew *Abc_Ntk_t, vControls *Vec_Ptr_t, vClasses *Vec_Int_t) {
    var pObj, pNodeNew, pCtrl, pDriver *Abc_Obj_t;
    var i, Class int32;
    assert(Abc_NtkPoNum(pNtkOld) == Abc_NtkPoNum(pNtkNew), "Abc_NtkPoNum(pNtkOld) == Abc_NtkPoNum(pNtkNew)");
    // match the POs of the old into new
    Abc_NtkForEachPo( pNtkOld, pObj, i )
        pObj.pCopy = Abc_NtkPo( pNtkNew, i );
    // remap the flops
    Vec_PtrForEachEntry( Abc_Obj_t *, vControls, pObj, i )
    {
        assert(Abc_ObjIsPo(pObj) && pObj.pNtk == pNtkOld, "Abc_ObjIsPo(pObj) && pObj.pNtk == pNtkOld");
        Vec_PtrWriteEntry( vControls, i, pObj.pCopy );
    }
    // create self-loops
    assert(Abc_NtkLatchNum(pNtkNew) == Vec_IntSize(vClasses), "Abc_NtkLatchNum(pNtkNew) == Vec_IntSize(vClasses)");
    Abc_NtkForEachLatch( pNtkNew, pObj, i )
    {
        Class = Vec_IntEntry( vClasses, i );
        if Class == -1 {
            continue; }
        pDriver = Abc_ObjFanin0(Abc_ObjFanin0(pObj));
        pCtrl = (Abc_Obj_t *)Vec_PtrEntry( vControls, Class );
        pCtrl = Abc_ObjFanin0( pCtrl );
        pNodeNew = Abc_NtkCreateNode( pNtkNew );
        Abc_ObjAddFanin( pNodeNew, pCtrl );
        Abc_ObjAddFanin( pNodeNew, pDriver );
        Abc_ObjAddFanin( pNodeNew, Abc_ObjFanout0(pObj) );
        Abc_ObjSetData( pNodeNew, Abc_SopRegister((Mem_Flex_t *)pNtkNew.pManFunc, "0-1 1\n11- 1\n") );
        Abc_ObjPatchFanin( Abc_ObjFanin0(pObj), pDriver, pNodeNew );
    }
    // remove the useless POs
    Vec_PtrForEachEntry( Abc_Obj_t *, vControls, pObj, i )
        Abc_NtkDeleteObj( pObj );
}

/**Function*************************************************************

  Synopsis    [Classify flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkCRetime(pNtk *Abc_Ntk_t, fVerbose int32) *Abc_Ntk_t {
    var pNtkNew *Abc_Ntk_t;
    var vControls *Vec_Ptr_t;
    var vFlopClasses, vFlopClassesNew *Vec_Int_t;
    var pObj, pDriver, pFlopOut, pObjPo *Abc_Obj_t;
    var i, iFlop, CountN, Count2, Count1, Count0 int32;

    // duplicate the AIG
    pNtk = Abc_NtkDup( pNtk );

    // update registers
    vControls    = Vec_PtrAlloc( 100 );
    vFlopClasses = Vec_IntAlloc( 100 );
    Abc_NtkForEachLatch( pNtk, pObj, i )
    {
        pFlopOut = Abc_ObjFanout0(pObj);
        pDriver = Abc_ObjFanin0( Abc_ObjFanin0(pObj) );
        if Abc_ObjFaninNum(pDriver) != 3 {
            Vec_IntPush( vFlopClasses, -1 );
            CountN++;
            continue;
        }
        if Abc_ObjFanin(pDriver, 1) != pFlopOut && Abc_ObjFanin(pDriver, 2) != pFlopOut {
            Vec_IntPush( vFlopClasses, -1 );
            Count2++;
            continue;
        }
        if Abc_ObjFanin(pDriver, 1) == pFlopOut {
            Vec_IntPush( vFlopClasses, -1 );
            Count1++;
            continue;
        }
        assert(Abc_ObjFanin(pDriver, 2) == pFlopOut, "Abc_ObjFanin(pDriver, 2) == pFlopOut");
        Count0++;
        Vec_PtrPushUnique( vControls, Abc_ObjFanin0(pDriver) );
        // set the flop class
        iFlop = Vec_PtrFind( vControls, Abc_ObjFanin0(pDriver) );
        Vec_IntPush( vFlopClasses, iFlop );
        // update
        Abc_ObjPatchFanin( Abc_ObjFanin0(pObj), pDriver, Abc_ObjFanin(pDriver, 1) );
    }
    if Count1 {
        fmt.Printf("Opposite phase enable is present in %d flops (out of %d).\n", Count1, Abc_NtkLatchNum(pNtk)); }
    if fVerbose != 0 {
    fmt.Printf("CountN = %4d. Count2 = %4d. Count1 = %4d. Count0 = %4d. Ctrls = %d.\n", CountN, Count2, Count1, Count0, Vec_PtrSize(vControls)); }

    // add the controls to the list of POs
    Vec_PtrForEachEntry( Abc_Obj_t *, vControls, pObj, i )
    {
        pObjPo = Abc_NtkCreatePo( pNtk );
        Abc_ObjAddFanin( pObjPo, pObj );
        Abc_ObjAssignName( pObjPo, Abc_ObjName(pObjPo), nil );
        Vec_PtrWriteEntry( vControls, i, pObjPo );
    }
    Abc_NtkOrderCisCos( pNtk );
    Abc_NtkCleanup( pNtk, fVerbose );

    // performs retiming with classes
    pNtkNew = Abc_NtkRetimeWithClassesNtk( pNtk, vFlopClasses, &vFlopClassesNew, fVerbose );
    Abc_NtkTransformBack( pNtk, pNtkNew, vControls, vFlopClassesNew );
//    assert( Abc_NtkPoNum(pNtkNew) == Abc_NtkPoNum(pNtk) );
    Abc_NtkDelete( pNtk );

    Vec_PtrFree( vControls );
//    Vec_IntFree( vFlopClasses );
    Vec_IntFree( vFlopClassesNew );
    return pNtkNew;
}

/**Function*************************************************************

  Synopsis    [Resimulates CEX and return the ID of the PO that failed.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkVerifyCex(pNtk *Abc_Ntk_t, p *Abc_Cex_t) int32 {
    var pObj *Abc_Obj_t;
    var RetValue, i, k, iBit int32;
    assert(Abc_NtkIsStrash(pNtk), "Abc_NtkIsStrash(pNtk)");
    assert(p.nPis  == Abc_NtkPiNum(pNtk), "p.nPis  == Abc_NtkPiNum(pNtk)");
//    assert( p->nRegs == Abc_NtkLatchNum(pNtk) );
    Abc_NtkCleanMarkC( pNtk );
    Abc_AigConst1(pNtk).fMarkC = 1;
    // initialize flops
    Abc_NtkForEachLatch( pNtk, pObj, i )
        Abc_ObjFanout0(pObj).fMarkC = Abc_InfoHasBit(p.pData, iBit++);
    // simulate timeframes
    iBit = p.nRegs;
    for  i = 0; i <= p.iFrame; i++  {
        Abc_NtkForEachPi( pNtk, pObj, k )
            pObj.fMarkC = Abc_InfoHasBit(p.pData, iBit++);
        Abc_NtkForEachNode( pNtk, pObj, k )
            pObj.fMarkC = (Abc_ObjFanin0(pObj).fMarkC ^ Abc_ObjFaninC0(pObj)) & 
                           (Abc_ObjFanin1(pObj).fMarkC ^ Abc_ObjFaninC1(pObj));
        Abc_NtkForEachCo( pNtk, pObj, k )
            pObj.fMarkC = Abc_ObjFanin0(pObj).fMarkC ^ Abc_ObjFaninC0(pObj);
        Abc_NtkForEachLatch( pNtk, pObj, k )
            Abc_ObjFanout0(pObj).fMarkC = Abc_ObjFanin0(pObj).fMarkC;
    }
    assert(iBit == p.nBits, "iBit == p.nBits");
    // figure out the number of failed output
    RetValue = -1;
    Abc_NtkForEachPo( pNtk, pObj, i )
    {
        if pObj.fMarkC {
            RetValue = i;
            break;
        }
    }
    Abc_NtkCleanMarkC( pNtk );
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

