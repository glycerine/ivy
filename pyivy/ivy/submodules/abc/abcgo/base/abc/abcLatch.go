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
    vControls, vFlopClasses, counts := abcClassifyLatchControls(pNtk)
    CountN, Count2, Count1, Count0 = int32(counts.notMux), int32(counts.other), int32(counts.opposite), int32(counts.enabled)
    if Count1 {
        fmt.Printf("Opposite phase enable is present in %d flops (out of %d).\n", Count1, Abc_NtkLatchNum(pNtk)); }
    if fVerbose != 0 {
    fmt.Printf("CountN = %4d. Count2 = %4d. Count1 = %4d. Count0 = %4d. Ctrls = %d.\n", CountN, Count2, Count1, Count0, Vec_PtrSize(vControls)); }

    // Add control outputs before retiming.
    abcExposeLatchControls(pNtk, vControls, fVerbose)

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
////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

