package ssw

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [sswPart.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Inductive prover with constraints.]

  Synopsis    [Partitioned signal correspondence.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 1, 2008.]

  Revision    [$Id: sswPart.c,v 1.00 2008/09/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "sswInt.h"
//#include "aig/ioa/ioa.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Performs partitioned sequential SAT sweeping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Ssw_SignalCorrespondencePart(pAig *Aig_Man_t, pPars *Ssw_Pars_t) *Aig_Man_t {
    var fPrintParts int32;
    var Buffer string // was [100]byte in C

    var pTemp, pNew *Aig_Man_t;
    var vResult *Vec_Ptr_t;
    var vPart *Vec_Int_t;
    var pMapBack *int32;
    var i, nCountPis, nCountRegs int32;
    var nClasses, nPartSize, fVerbose int32;
    var clk abctime  = Abc_Clock();
    if pPars.fConstrs {
        Abc_Print( 1, "Cannot use partitioned computation with constraints.\n" );
        return nil;
    }
    // save parameters
    nPartSize = pPars.nPartSize; pPars.nPartSize = 0;
    fVerbose  = pPars.fVerbose;  pPars.fVerbose  = 0;
    // generate partitions
    if pAig.vClockDoms {
        // divide large clock domains into separate partitions
        vResult = Vec_PtrAlloc( 100 );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize((Vec_Ptr_t *)pAig.vClockDoms)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry((Vec_Ptr_t *)pAig.vClockDoms, i)), 1); i++  {
            if nPartSize && Vec_IntSize(vPart) > nPartSize {
                Aig_ManPartDivide( vResult, vPart, nPartSize, pPars.nOverSize );
            } else {
                Vec_PtrPush( vResult, Vec_IntDup(vPart) ); }
        }
    
    } else {
        vResult = Aig_ManRegPartitionSimple( pAig, nPartSize, pPars.nOverSize ); }
//    vResult = Aig_ManPartitionSmartRegisters( pAig, nPartSize, 0 ); 
//    vResult = Aig_ManRegPartitionSmart( pAig, nPartSize );
    if fPrintParts != 0 {
        // print partitions
        Abc_Print( 1, "Simple partitioning. %d partitions are saved:\n", Vec_PtrSize(vResult) );
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vResult)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vResult, i)), 1); i++  {
//            extern void Ioa_WriteAiger( Aig_Man_t * pMan, char * pFileName, int fWriteSymbols, int fCompact );
            Buffer = fmt.Sprintf("part%03d.aig", i);
            pTemp = Aig_ManRegCreatePart( pAig, vPart, &nCountPis, &nCountRegs, nil );
            Ioa_WriteAiger( pTemp, Buffer, 0, 0 );
            Abc_Print( 1, "part%03d.aig : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d.\n",
                i, Vec_IntSize(vPart), Aig_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Aig_ManNodeNum(pTemp) );
            Aig_ManStop( pTemp );
        }
    }

    // perform SSW with partitions
    Aig_ManReprStart( pAig, Aig_ManObjNumMax(pAig) );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vResult)) && (((vPart) = (Vec_Int_t *)Vec_PtrEntry(vResult, i)), 1); i++  {
        pTemp = Aig_ManRegCreatePart( pAig, vPart, &nCountPis, &nCountRegs, &pMapBack );
        Aig_ManSetRegNum( pTemp, pTemp.nRegs );
        // create the projection of 1-hot registers
        if pAig.vOnehots {
            pTemp.vOnehots = Aig_ManRegProjectOnehots( pAig, pTemp, pAig.vOnehots, fVerbose ); }
        // run SSW
        if nCountPis>0 {
            pNew = Ssw_SignalCorrespondence( pTemp, pPars );
            nClasses = Aig_TransferMappedClasses( pAig, pTemp, pMapBack );
            if fVerbose != 0 {
                Abc_Print( 1, "%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. It = %3d. Cl = %5d.\n",
                    i, Vec_IntSize(vPart), Aig_ManCiNum(pTemp)-Vec_IntSize(vPart), nCountPis, nCountRegs, Aig_ManNodeNum(pTemp), pPars.nIters, nClasses ); }
            Aig_ManStop( pNew );
        }
        Aig_ManStop( pTemp );
        // inlining of macro ABC_FREE
        ((pMapBack) ? (free((char *) (pMapBack)), (pMapBack) = 0) : 0);
    }
    // remap the AIG
    pNew = Aig_ManDupRepr( pAig, 0 );
    Aig_ManSeqCleanup( pNew );
//    Aig_ManPrintStats( pAig );
//    Aig_ManPrintStats( pNew );
    Vec_VecFree( (Vec_Vec_t *)vResult );
    pPars.nPartSize = nPartSize;
    pPars.fVerbose = fVerbose;
    if fVerbose != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
    return pNew;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
