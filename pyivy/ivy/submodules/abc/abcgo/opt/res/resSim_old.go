package res

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [resSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Resynthesis package.]

  Synopsis    [Simulation engine.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resSim.c,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "resInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocate simulation engine.]

  Description []
               
  SideEffects []

  SeeAlso     []
 
***********************************************************************/
func Res_SimAlloc(nWords int32) *Res_Sim_t {
    var p *Res_Sim_t;
    p = ALLOC( Res_Sim_t, 1 );
    memset( p, 0, sizeof(Res_Sim_t) );
    // simulation parameters
    p.nWords    = nWords;
    p.nPats     = 8 * sizeof(unsigned) * p.nWords;
    p.nWordsOut = p.nPats * p.nWords;
    p.nPatsOut  = p.nPats * p.nPats;
    // simulation info
    p.vPats     = Vec_PtrAllocSimInfo( 1024, p.nWords );
    p.vPats0    = Vec_PtrAllocSimInfo( 128,  p.nWords );
    p.vPats1    = Vec_PtrAllocSimInfo( 128,  p.nWords );
    p.vOuts     = Vec_PtrAllocSimInfo( 128,  p.nWordsOut );
    // resub candidates
    p.vCands    = Vec_VecStart( 16 );
    return p;
}

/**Function*************************************************************

  Synopsis    [Allocate simulation engine.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimAdjust(p *Res_Sim_t, pAig *Abc_Ntk_t) {
    srand( 0xABC );

    assert(Abc_NtkIsStrash(pAig), "Abc_NtkIsStrash(pAig)");
    p.pAig = pAig;
    if Vec_PtrSize(p.vPats) < Abc_NtkObjNumMax(pAig)+1 {
        Vec_PtrFree( p.vPats );
        p.vPats = Vec_PtrAllocSimInfo( Abc_NtkObjNumMax(pAig)+1, p.nWords );
    }
    if Vec_PtrSize(p.vPats0) < Abc_NtkPiNum(pAig) {
        Vec_PtrFree( p.vPats0 );
        p.vPats0 = Vec_PtrAllocSimInfo( Abc_NtkPiNum(pAig), p.nWords );
    }
    if Vec_PtrSize(p.vPats1) < Abc_NtkPiNum(pAig) {
        Vec_PtrFree( p.vPats1 );
        p.vPats1 = Vec_PtrAllocSimInfo( Abc_NtkPiNum(pAig), p.nWords );
    }
    if Vec_PtrSize(p.vOuts) < Abc_NtkPoNum(pAig) {
        Vec_PtrFree( p.vOuts );
        p.vOuts = Vec_PtrAllocSimInfo( Abc_NtkPoNum(pAig), p.nWordsOut );
    }
    // clean storage info for patterns
    Abc_InfoClear( Vec_PtrEntry(p.vPats0,0), p.nWords * Abc_NtkPiNum(pAig) );
    Abc_InfoClear( Vec_PtrEntry(p.vPats1,0), p.nWords * Abc_NtkPiNum(pAig) );
    p.nPats0 = 0;
    p.nPats1 = 0;
}

/**Function*************************************************************

  Synopsis    [Free simulation engine.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimFree(p *Res_Sim_t) {
    Vec_PtrFree( p.vPats );
    Vec_PtrFree( p.vPats0 );
    Vec_PtrFree( p.vPats1 );
    Vec_PtrFree( p.vOuts );
    Vec_VecFree( p.vCands );
    free( p );
}


/**Function*************************************************************

  Synopsis    [Sets random PI simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimSetRandom(p *Res_Sim_t) {
    var pObj *Abc_Obj_t;
    var pInfo *uint32;
    var i int32;
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
        pInfo = Vec_PtrEntry( p.vPats, pObj.Id );
        Abc_InfoRandom( pInfo, p.nWords );
    }
}

/**Function*************************************************************

  Synopsis    [Sets given PI simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimSetGiven(p *Res_Sim_t, vInfo *Vec_Ptr_t) {
    var pObj *Abc_Obj_t;
    var pInfo, pInfo2 *uint32;
    var i, w int32;
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
        pInfo = Vec_PtrEntry( p.vPats, pObj.Id );
        pInfo2 = Vec_PtrEntry( vInfo, i );
        for  w = 0; w < p.nWords; w++  {
            pInfo[w] = pInfo2[w];
        }
    }
}

/**Function*************************************************************

  Synopsis    [Simulates one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimPerformOne(pNode *Abc_Obj_t, vSimInfo *Vec_Ptr_t, nSimWords int32) {
    var pInfo, pInfo1, pInfo2 *uint32;
    var k, fComp1, fComp2 int32;
    // simulate the internal nodes
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)");
    pInfo  = Vec_PtrEntry(vSimInfo, pNode.Id);
    pInfo1 = Vec_PtrEntry(vSimInfo, Abc_ObjFaninId0(pNode));
    pInfo2 = Vec_PtrEntry(vSimInfo, Abc_ObjFaninId1(pNode));
    fComp1 = Abc_ObjFaninC0(pNode);
    fComp2 = Abc_ObjFaninC1(pNode);
    if (fComp1 != 0) && (fComp2 != 0) {
        for  k = 0; k < nSimWords; k++  {
            pInfo[k] = ~pInfo1[k] & ~pInfo2[k];
        }
    } else if (fComp1 != 0) && !(fComp2 != 0) {
        for  k = 0; k < nSimWords; k++  {
            pInfo[k] = ~pInfo1[k] &  pInfo2[k];
        }
    } else if !(fComp1 != 0) && (fComp2 != 0) {
        for  k = 0; k < nSimWords; k++  {
            pInfo[k] =  pInfo1[k] & ~pInfo2[k];
        }
    } else { // if ( fComp1 && fComp2 )
        for  k = 0; k < nSimWords; k++  {
            pInfo[k] =  pInfo1[k] &  pInfo2[k];
        } }
}

/**Function*************************************************************

  Synopsis    [Simulates one CO node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimTransferOne(pNode *Abc_Obj_t, vSimInfo *Vec_Ptr_t, nSimWords int32) {
    var pInfo, pInfo1 *uint32;
    var k, fComp1 int32;
    // simulate the internal nodes
    assert(Abc_ObjIsCo(pNode), "Abc_ObjIsCo(pNode)");
    pInfo  = Vec_PtrEntry(vSimInfo, pNode.Id);
    pInfo1 = Vec_PtrEntry(vSimInfo, Abc_ObjFaninId0(pNode));
    fComp1 = Abc_ObjFaninC0(pNode);
    if fComp1 != 0 {
        for  k = 0; k < nSimWords; k++  {
            pInfo[k] = ~pInfo1[k];
        }
    } else { 
        for  k = 0; k < nSimWords; k++  {
            pInfo[k] =  pInfo1[k];
        } }
}

/**Function*************************************************************

  Synopsis    [Performs one round of simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimPerformRound(p *Res_Sim_t) {
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_InfoFill( Vec_PtrEntry(p.vPats,0), p.nWords );
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((p.pAig).vObjs)) && (((pObj) = Abc_NtkObj(p.pAig, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Res_SimPerformOne( pObj, p.vPats, p.nWords ); }
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(p.pAig)) && (((pObj) = Abc_NtkPo(p.pAig, i)), 1); i++  {
        Res_SimTransferOne( pObj, p.vPats, p.nWords );
    }
}

/**Function*************************************************************

  Synopsis    [Processes simulation patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimProcessPats(p *Res_Sim_t) {
    var pObj *Abc_Obj_t;
    var pInfoCare, pInfoNode *uint32;
    var i, j, nDcs int32;
    pInfoCare = Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 0).Id );
    pInfoNode = Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 1).Id );
    for  i = 0; i < p.nPats; i++  {
        // skip don't-care patterns
        if !Abc_InfoHasBit(pInfoCare, i) {
            nDcs++;
            continue;
        }
        // separate offset and onset patterns
        if !Abc_InfoHasBit(pInfoNode, i) {
            if p.nPats0 >= p.nPats {
                continue; }
            // inlining of macro Abc_NtkForEachPi
            for  j = 0; (j < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, j)), 1); j++  {
                if Abc_InfoHasBit( Vec_PtrEntry(p.vPats, pObj.Id), i ) {
                    Abc_InfoSetBit( Vec_PtrEntry(p.vPats0, j), p.nPats0 ); }
            }
            p.nPats0++;
        
        } else {
            if p.nPats1 >= p.nPats {
                continue; }
            // inlining of macro Abc_NtkForEachPi
            for  j = 0; (j < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, j)), 1); j++  {
                if Abc_InfoHasBit( Vec_PtrEntry(p.vPats, pObj.Id), i ) {
                    Abc_InfoSetBit( Vec_PtrEntry(p.vPats1, j), p.nPats1 ); }
            }
            p.nPats1++;
        }
    }
}

/**Function*************************************************************

  Synopsis    [Pads the extra space with duplicated simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimPadSimInfo(vPats *Vec_Ptr_t, nPats int32, nWords int32) {
    var pInfo *uint32;
    var i, w, iWords int32;
    assert(nPats > 0 && nPats < nWords * 8 * int32( sizeof(unsigned)), "nPats > 0 && nPats < nWords * 8 * int32( sizeof(unsigned))");
    // pad the first word
    if nPats < 8 * sizeof(unsigned) {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(vPats)) && (((pInfo) = (unsigned *)Vec_PtrEntry(vPats, i)), 1); i++  {
            if pInfo[0] & 1 {
                pInfo[0] |= ((~0) << nPats); }
        }
        nPats = 8 * sizeof(unsigned);
    }
    // pad the empty words
    iWords = nPats / (8 * sizeof(unsigned));
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vPats)) && (((pInfo) = (unsigned *)Vec_PtrEntry(vPats, i)), 1); i++  {
        for  w = iWords; w < nWords; w++  {
            pInfo[w] = pInfo[0];
        }
    }
}

/**Function*************************************************************

  Synopsis    [Duplicates the simulation info to fill the space.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimDeriveInfoReplicate(p *Res_Sim_t) {
    var pInfo, pInfo2 *uint32;
    var pObj *Abc_Obj_t;
    var i, j, w int32;
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(p.pAig)) && (((pObj) = Abc_NtkPo(p.pAig, i)), 1); i++  {
        pInfo = Vec_PtrEntry( p.vPats, pObj.Id );
        pInfo2 = Vec_PtrEntry( p.vOuts, i );
        for  j = 0; j < p.nPats; j++  {
            for  w = 0; w < p.nWords; w++  {
                *pInfo2++ = pInfo[w];
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Complement the simulation info if necessary.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimDeriveInfoComplement(p *Res_Sim_t) {
    var pInfo, pInfo2 *uint32;
    var pObj *Abc_Obj_t;
    var i, j, w int32;
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(p.pAig)) && (((pObj) = Abc_NtkPo(p.pAig, i)), 1); i++  {
        pInfo = Vec_PtrEntry( p.vPats, pObj.Id );
        pInfo2 = Vec_PtrEntry( p.vOuts, i );
        for  j = 0; j < p.nPats; j++, pInfo2 += p.nWords  {
            if Abc_InfoHasBit( pInfo, j ) {
                for  w = 0; w < p.nWords; w++  {
                    pInfo2[w] = ~pInfo2[w];
                } }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Free simulation engine.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimReportOne(p *Res_Sim_t) {
    var pInfoCare, pInfoNode *uint32;
    var i, nDcs, nOnes, nZeros int32;
    pInfoCare = Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 0).Id );
    pInfoNode = Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 1).Id );
    nDcs = nOnes = nZeros = 0;
    for  i = 0; i < p.nPats; i++  {
        // skip don't-care patterns
        if !Abc_InfoHasBit(pInfoCare, i) {
            nDcs++;
            continue;
        }
        // separate offset and onset patterns
        if !Abc_InfoHasBit(pInfoNode, i) {
            nZeros++;
        } else {
            nOnes++; }
    }
    fmt.Printf("On = %3d (%7.2f %%)  ", nOnes, 100.0*nOnes/p.nPats);
    fmt.Printf("Off = %3d (%7.2f %%)  ", nZeros, 100.0*nZeros/p.nPats);
    fmt.Printf("Dc = %3d (%7.2f %%)  ", nDcs, 100.0*nDcs/p.nPats);
    fmt.Printf("P0 = %3d ", p.nPats0);
    fmt.Printf("P1 = %3d ", p.nPats1);
    if p.nPats0 < 4 || p.nPats1 < 4 {
        fmt.Printf("*"); }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Prints output patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimPrintOutPatterns(p *Res_Sim_t, pAig *Abc_Ntk_t) {
    var pObj *Abc_Obj_t;
    var pInfo2 *uint32;
    var i int32;
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pAig)) && (((pObj) = Abc_NtkPo(pAig, i)), 1); i++  {
        pInfo2 = Vec_PtrEntry( p.vOuts, i );
        Extra_PrintBinary( os.Stdout, pInfo2, p.nPatsOut );
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Prepares simulation info for candidate filtering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimPrepare(p *Res_Sim_t, pAig *Abc_Ntk_t, nTruePis int32, fVerbose int32) int32 {
    var Limit int32;
    // prepare the manager
    Res_SimAdjust( p, pAig );
    // collect 0/1 simulation info
    for  Limit = 0; Limit < 10; Limit++  {
        Res_SimSetRandom( p );
        Res_SimPerformRound( p );
        Res_SimProcessPats( p );
        if !(p.nPats0 < p.nPats || p.nPats1 < p.nPats) {
            break; }
    }
//    printf( "%d   ", Limit );
    // report the last set of patterns
//    Res_SimReportOne( p );
//    printf( "\n" );
    // quit if there is not enough
//    if ( p->nPats0 < 4 || p->nPats1 < 4 )
    if p.nPats0 < 4 || p.nPats1 < 4 {
//        Res_SimReportOne( p );
        return 0;
    }
    // create bit-matrix info
    if p.nPats0 < p.nPats {
        Res_SimPadSimInfo( p.vPats0, p.nPats0, p.nWords ); }
    if p.nPats1 < p.nPats {
        Res_SimPadSimInfo( p.vPats1, p.nPats1, p.nWords ); }
    // resimulate 0-patterns
    Res_SimSetGiven( p, p.vPats0 );
    Res_SimPerformRound( p );
    Res_SimDeriveInfoReplicate( p );
    // resimulate 1-patterns
    Res_SimSetGiven( p, p.vPats1 );
    Res_SimPerformRound( p );
    Res_SimDeriveInfoComplement( p );
    // print output patterns
//    Res_SimPrintOutPatterns( p, pAig );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

