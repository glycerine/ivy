package mfs

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [mfsSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [The good old minimization with complete don't-cares.]

  Synopsis    [Procedures to compute don't-cares using SAT.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: mfsSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "mfsInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Enumerates through the SAT assignments.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkMfsSolveSat_iter(p *Mfs_Man_t) int32 {
    var Lits [MFS_FANIN_MAX]int32;
    var RetValue, nBTLimit, iVar, b, Mint int32;
//    int nConfs = p->pSat->stats.conflicts;
    if p.nTotConfLim && p.nTotConfLim <= p.pSat.stats.conflicts {
        return -1; }
    if p.nTotConfLim {
        nBTLimit = p.nTotConfLim - p.pSat.stats.conflicts;
    } else {
        nBTLimit = 0;
    }
    RetValue = sat_solver_solve( p.pSat, nil, nil, (ABC_INT64_T)nBTLimit, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    assert(RetValue == l_Undef || RetValue == l_True || RetValue == l_False, "RetValue == l_Undef || RetValue == l_True || RetValue == l_False");
//printf( "%c", RetValue==l_Undef ? '?' : (RetValue==l_False ? '-' : '+') );
//printf( "%d ", p->pSat->stats.conflicts-nConfs );
//if ( RetValue==l_False )
//printf( "\n" );
    if RetValue == l_Undef {
        return -1; }
    if RetValue == l_False {
        return 0; }
    p.nCares++;
    // add SAT assignment to the solver
    Mint = 0;
    // inlining of macro Vec_IntForEachEntry
    for  b = 0; (b < Vec_IntSize(p.vProjVarsSat)) && (((iVar) = Vec_IntEntry(p.vProjVarsSat, b)), 1); b++  {
        Lits[b] = toLit( iVar );
        if sat_solver_var_value( p.pSat, iVar ) {
            Mint |= (1 << b);
            Lits[b] = lit_neg( Lits[b] );
        }
    }
    assert(!Abc_InfoHasBit(p.uCare, Mint), "!Abc_InfoHasBit(p.uCare, Mint)");
    Abc_InfoSetBit( p.uCare, Mint );
    // add the blocking clause
    RetValue = sat_solver_addclause( p.pSat, Lits, Lits + Vec_IntSize(p.vProjVarsSat) );
    if RetValue == 0 {
        return 0; }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Enumerates through the SAT assignments.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Abc_NtkMfsSolveSat(p *Mfs_Man_t, pNode *Abc_Obj_t) int32 {
    var pObjPo *Aig_Obj_t;
    var RetValue, i int32;
    // collect projection variables
    Vec_IntClear( p.vProjVarsSat );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = Aig_ManCoNum(p.pAigWin) - Abc_ObjFaninNum(pNode); (i < Vec_PtrSize(p.pAigWin.vCos)) && (((pObjPo) = (Aig_Obj_t *)Vec_PtrEntry(p.pAigWin.vCos, i)), 1); i++  {
        assert(p.pCnf.pVarNums[pObjPo.Id] >= 0, "p.pCnf.pVarNums[pObjPo.Id] >= 0");
        Vec_IntPush( p.vProjVarsSat, p.pCnf.pVarNums[pObjPo.Id] );
    }

    // prepare the truth table of care set
    p.nFanins = Vec_IntSize( p.vProjVarsSat );
    p.nWords = Abc_TruthWordNum( p.nFanins );
    memset( p.uCare, 0, sizeof(unsigned) * p.nWords );

    // iterate through the SAT assignments
    p.nCares = 0;
    p.nTotConfLim = p.pPars.nBTLimit;
    for  (RetValue = Abc_NtkMfsSolveSat_iter(p)) == 1  {; }
    if RetValue == -1 {
        return 0; }

    // write statistics
    p.nMintsCare  += p.nCares;
    p.nMintsTotal += (1<<p.nFanins);

    if p.pPars.fVeryVerbose {
        fmt.Printf("Node %4d : Care = %2d. Total = %2d.  ", pNode.Id, p.nCares, (1<<p.nFanins));
        Extra_PrintBinary( os.Stdout, p.uCare, (1<<p.nFanins) );
        fmt.Printf("\n");
    }

    // map the care 
    if p.nFanins > 4 {
        return 1; }
    if p.nFanins == 4 {
        p.uCare[0] = p.uCare[0] | (p.uCare[0] << 16); }
    if p.nFanins == 3 {
        p.uCare[0] = p.uCare[0] | (p.uCare[0] << 8) | (p.uCare[0] << 16) | (p.uCare[0] << 24); }
    if p.nFanins == 2 {
        p.uCare[0] = p.uCare[0] | (p.uCare[0] <<  4) | (p.uCare[0] <<  8) | (p.uCare[0] << 12) |
              (p.uCare[0] << 16) | (p.uCare[0] << 20) | (p.uCare[0] << 24) | (p.uCare[0] << 28); }
    assert(p.nFanins != 1, "p.nFanins != 1");
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds one-hotness constraints for the window inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Abc_NtkAddOneHotness(p *Mfs_Man_t) int32 {
    var pObj1, pObj2 *Aig_Obj_t;
    var i int32;
    var k int32;
    var Lits [2]int32;
    for  i = 0; i < Vec_PtrSize(p.pAigWin.vCis); i++  {
    for  k = i+1; k < Vec_PtrSize(p.pAigWin.vCis); k++  {
        pObj1 = Aig_ManCi( p.pAigWin, i );
        pObj2 = Aig_ManCi( p.pAigWin, k );
        Lits[0] = toLitCond( p.pCnf.pVarNums[pObj1.Id], 1 );
        Lits[1] = toLitCond( p.pCnf.pVarNums[pObj2.Id], 1 );
        if !sat_solver_addclause( p.pSat, Lits, Lits+2 ) {
            sat_solver_delete( p.pSat );
            p.pSat = nil;
            return 0;
        }
    }
    }
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

