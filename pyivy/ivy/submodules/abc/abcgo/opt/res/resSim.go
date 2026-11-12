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
    // inlining of macro ABC_ALLOC
    p = ((Res_Sim_t *) malloc(sizeof(Res_Sim_t) * (size_t)(1)));
    memset( p, 0, sizeof(Res_Sim_t) );
    // simulation parameters
    p.nWords    = nWords;
    p.nPats     = p.nWords * 8 * sizeof(unsigned);
    p.nWordsIn  = p.nPats;
    p.nBytesIn  = p.nPats * sizeof(unsigned);
    p.nPatsIn   = p.nPats * 8 * sizeof(unsigned);
    p.nWordsOut = p.nPats * p.nWords;
    p.nPatsOut  = p.nPats * p.nPats;
    // simulation info
    p.vPats     = Vec_PtrAllocSimInfo( 1024, p.nWordsIn );
    p.vPats0    = Vec_PtrAllocSimInfo( 128, p.nWords );
    p.vPats1    = Vec_PtrAllocSimInfo( 128, p.nWords );
    p.vOuts     = Vec_PtrAllocSimInfo( 128, p.nWordsOut );
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
func Res_SimAdjust(p *Res_Sim_t, pAig *Abc_Ntk_t, nTruePis int32) {
    srand( 0xABC );

    assert(Abc_NtkIsStrash(pAig), "Abc_NtkIsStrash(pAig)");
    p.pAig = pAig;
    p.nTruePis = nTruePis;
    if Vec_PtrSize(p.vPats) < Abc_NtkObjNumMax(pAig)+1 {
        Vec_PtrFree( p.vPats );
        p.vPats = Vec_PtrAllocSimInfo( Abc_NtkObjNumMax(pAig)+1, p.nWordsIn );
    }
    if Vec_PtrSize(p.vPats0) < nTruePis {
        Vec_PtrFree( p.vPats0 );
        p.vPats0 = Vec_PtrAllocSimInfo( nTruePis, p.nWords );
    }
    if Vec_PtrSize(p.vPats1) < nTruePis {
        Vec_PtrFree( p.vPats1 );
        p.vPats1 = Vec_PtrAllocSimInfo( nTruePis, p.nWords );
    }
    if Vec_PtrSize(p.vOuts) < Abc_NtkPoNum(pAig) {
        Vec_PtrFree( p.vOuts );
        p.vOuts = Vec_PtrAllocSimInfo( Abc_NtkPoNum(pAig), p.nWordsOut );
    }
    // clean storage info for patterns
    Abc_InfoClear( (unsigned *)Vec_PtrEntry(p.vPats0,0), p.nWords * nTruePis );
    Abc_InfoClear( (unsigned *)Vec_PtrEntry(p.vPats1,0), p.nWords * nTruePis );
    p.nPats0 = 0;
    p.nPats1 = 0;
    p.fConst0 = 0;
    p.fConst1 = 0;
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
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}


/**Function*************************************************************

  Synopsis    [Sets random PI simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_InfoRandomBytes(p *uint32, nWords int32) { 
    var i, Num int32; 
    for  i = nWords - 1; i >= 0; i--  { 
        Num = rand(); 
        p[i] = (Num & 1)? 0xff : 0;
        p[i] = (p[i] << 8) | ((Num & 2)? 0xff : 0);
        p[i] = (p[i] << 8) | ((Num & 4)? 0xff : 0);
        p[i] = (p[i] << 8) | ((Num & 8)? 0xff : 0);
    } 
//    Extra_PrintBinary( stdout, p, 32 ); printf( "\n" );
} 

/**Function*************************************************************

  Synopsis    [Sets random PI simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimSetRandomBytes(p *Res_Sim_t) {
    var pObj *Abc_Obj_t;
    var pInfo *uint32;
    var i int32;
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
        pInfo = (unsigned *)Vec_PtrEntry( p.vPats, pObj.Id );
        if i < p.nTruePis {
            Abc_InfoRandomBytes( pInfo, p.nWordsIn );
        } else {
            Abc_InfoRandom( pInfo, p.nWordsIn ); }
    }
/*
    // double-check that all are byte-patterns
    Abc_NtkForEachPi( p->pAig, pObj, i )
    {
        if ( i == p->nTruePis )
            break;
        pInfoC = (unsigned char *)Vec_PtrEntry( p->vPats, pObj->Id );
        for ( k = 0; k < p->nBytesIn; k++ )
            assert( pInfoC[k] == 0 || pInfoC[k] == 0xff );
    }
*/
}

/**Function*************************************************************

  Synopsis    [Sets random PI simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimSetDerivedBytes(p *Res_Sim_t, fUseWalk int32) {
    var vPatsSource [2]*Vec_Ptr_t;
    var nPatsSource [2]int32;
    var pObj *Abc_Obj_t;
    unsigned char * pInfo;
    var i, k, z, s, nPats int32;

    // set several random patterns
    assert(p.nBytesIn % 32 == 0, "p.nBytesIn % 32 == 0");
    nPats = p.nBytesIn/8;
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
        if i == p.nTruePis {
            break; }
        Abc_InfoRandomBytes( (unsigned *)Vec_PtrEntry(p.vPats, pObj.Id), nPats/4 );
    }

    // set special patterns
    if fUseWalk != 0 {
        for  z = 0; z < 2; z++  {
            // set the zero pattern
            // inlining of macro Abc_NtkForEachPi
            for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
                if i == p.nTruePis {
                    break; }
                pInfo = (unsigned char *)Vec_PtrEntry( p.vPats, pObj.Id );
                pInfo[nPats] = z ? 0xff : 0;
            }
            if nPats++; nPats == p.nBytesIn {
                return; }
            // set the walking zero pattern
            for  k = 0; k < p.nTruePis; k++  {
                // inlining of macro Abc_NtkForEachPi
                for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
                    if i == p.nTruePis {
                        break; }
                    pInfo = (unsigned char *)Vec_PtrEntry( p.vPats, pObj.Id );
                    pInfo[nPats] = ((i == k) ^ z) ? 0xff : 0;
                }
                if nPats++; nPats == p.nBytesIn {
                    return; }
            }
        }
    }

    // decide what patterns to set first
    if p.nPats0 < p.nPats1 {
        nPatsSource[0] = p.nPats0;
        vPatsSource[0] = p.vPats0;
        nPatsSource[1] = p.nPats1;
        vPatsSource[1] = p.vPats1;
    
    } else {
        nPatsSource[0] = p.nPats1;
        vPatsSource[0] = p.vPats1;
        nPatsSource[1] = p.nPats0;
        vPatsSource[1] = p.vPats0;
    }
    for  z = 0; z < 2; z++  {
        for  s = nPatsSource[z] - 1; s >= 0; s--  {
//            if ( s == 0 )
//            printf( "Patterns:\n" );
            // set the given source pattern
            for  k = 0; k < p.nTruePis; k++  {
                // inlining of macro Abc_NtkForEachPi
                for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
                    if i == p.nTruePis {
                        break; }
                    pInfo = (unsigned char *)Vec_PtrEntry( p.vPats, pObj.Id );
                    if (i == k) ^ Abc_InfoHasBit( (unsigned *)Vec_PtrEntry(vPatsSource[z], i), s ) {
                        pInfo[nPats] = 0xff;
//                        if ( s == 0 )
//                        printf( "1" );
                    
                    } else {
                        pInfo[nPats] = 0;
//                        if ( s == 0 )
//                        printf( "0" );
                    }
                }
//                if ( s == 0 )
//                printf( "\n" );
                if nPats++; nPats == p.nBytesIn {
                    return; }
            }
        }
    }
    // clean the rest
    for  z = nPats; z < p.nBytesIn; z++  {
        // inlining of macro Abc_NtkForEachPi
        for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
            if i == p.nTruePis {
                break; }
            pInfo = (unsigned char *)Vec_PtrEntry( p.vPats, pObj.Id );
            memset( pInfo + nPats, 0, (size_t)(p.nBytesIn - nPats) );
        }
    }
/*
    // double-check that all are byte-patterns
    Abc_NtkForEachPi( p->pAig, pObj, i )
    {
        if ( i == p->nTruePis )
            break;
        pInfo = (unsigned char *)Vec_PtrEntry( p->vPats, pObj->Id );
        for ( k = 0; k < p->nBytesIn; k++ )
            assert( pInfo[k] == 0 || pInfo[k] == 0xff );
    }
*/
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
        if i == p.nTruePis {
            break; }
        pInfo = (unsigned *)Vec_PtrEntry( p.vPats, pObj.Id );
        pInfo2 = (unsigned *)Vec_PtrEntry( vInfo, i );
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
    pInfo  = (unsigned *)Vec_PtrEntry(vSimInfo, pNode.Id);
    pInfo1 = (unsigned *)Vec_PtrEntry(vSimInfo, Abc_ObjFaninId0(pNode));
    pInfo2 = (unsigned *)Vec_PtrEntry(vSimInfo, Abc_ObjFaninId1(pNode));
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
    pInfo  = (unsigned *)Vec_PtrEntry(vSimInfo, pNode.Id);
    pInfo1 = (unsigned *)Vec_PtrEntry(vSimInfo, Abc_ObjFaninId0(pNode));
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
func Res_SimPerformRound(p *Res_Sim_t, nWords int32) {
    var pObj *Abc_Obj_t;
    var i int32;
    Abc_InfoFill( (unsigned *)Vec_PtrEntry(p.vPats,0), nWords );
    // inlining of macro Abc_AigForEachAnd
    for  i = 0; (i < Vec_PtrSize((p.pAig).vObjs)) && (((pObj) = Abc_NtkObj(p.pAig, i)), 1); i++  {           if (pObj) == nil || !Abc_AigNodeIsAnd(pObj) {} else {
        Res_SimPerformOne( pObj, p.vPats, nWords ); }
    }
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(p.pAig)) && (((pObj) = Abc_NtkPo(p.pAig, i)), 1); i++  {
        Res_SimTransferOne( pObj, p.vPats, nWords );
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
        pInfo = (unsigned *)Vec_PtrEntry( p.vPats, pObj.Id );
        pInfo2 = (unsigned *)Vec_PtrEntry( p.vOuts, i );
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
        pInfo = (unsigned *)Vec_PtrEntry( p.vPats, pObj.Id );
        pInfo2 = (unsigned *)Vec_PtrEntry( p.vOuts, i );
        for  j = 0; j < p.nPats; j++, pInfo2 += p.nWords  {
            if Abc_InfoHasBit( pInfo, j ) {
                for  w = 0; w < p.nWords; w++  {
                    pInfo2[w] = ~pInfo2[w];
                } }
        }
    }
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
        pInfo2 = (unsigned *)Vec_PtrEntry( p.vOuts, i );
        Extra_PrintBinary( os.Stdout, pInfo2, p.nPatsOut );
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Prints output patterns.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimPrintNodePatterns(p *Res_Sim_t, pAig *Abc_Ntk_t) {
    var pInfo *uint32;
    pInfo = (unsigned *)Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 1).Id );
    Extra_PrintBinary( os.Stdout, pInfo, p.nPats );
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Counts the number of patters of different type.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimCountResults(p *Res_Sim_t, pnDcs *int32, pnOnes *int32, pnZeros *int32, fVerbose int32) {
    unsigned char * pInfoCare, * pInfoNode;
    var i, nTotal int32;
    pInfoCare = (unsigned char *)Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 0).Id );
    pInfoNode = (unsigned char *)Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 1).Id );
    for  i = 0; i < p.nBytesIn; i++  {
        if !pInfoCare[i] {
            (*pnDcs)++;
        } else if !pInfoNode[i] {
            (*pnZeros)++;
        } else {
            (*pnOnes)++; }
    }
    nTotal += *pnDcs;
    nTotal += *pnZeros;
    nTotal += *pnOnes;
    if fVerbose != 0 {
        fmt.Printf("Dc = %7.2f %%  ", 100.0*(*pnDcs)  /nTotal);
        fmt.Printf("On = %7.2f %%  ", 100.0*(*pnOnes) /nTotal);
        fmt.Printf("Off = %7.2f %%  ", 100.0*(*pnZeros)/nTotal);
    }
}

/**Function*************************************************************

  Synopsis    [Counts the number of patters of different type.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimCollectPatterns(p *Res_Sim_t, fVerbose int32) {
    var pObj *Abc_Obj_t;
    unsigned char * pInfoCare, * pInfoNode, * pInfo;
    var i, j int32;
    pInfoCare = (unsigned char *)Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 0).Id );
    pInfoNode = (unsigned char *)Vec_PtrEntry( p.vPats, Abc_NtkPo(p.pAig, 1).Id );
    for  i = 0; i < p.nBytesIn; i++  {
        // skip don't-care patterns
        if !pInfoCare[i] {
            continue; }
        // separate offset and onset patterns
        assert(pInfoNode[i] == 0 || pInfoNode[i] == 0xff, "pInfoNode[i] == 0 || pInfoNode[i] == 0xff");
        if !pInfoNode[i] {
            if p.nPats0 >= p.nPats {
                continue; }
            // inlining of macro Abc_NtkForEachPi
            for  j = 0; (j < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, j)), 1); j++  {
                if j == p.nTruePis {
                    break; }
                pInfo = (unsigned char *)Vec_PtrEntry( p.vPats, pObj.Id );
                assert(pInfo[i] == 0 || pInfo[i] == 0xff, "pInfo[i] == 0 || pInfo[i] == 0xff");
                if pInfo[i] {
                    Abc_InfoSetBit( (unsigned *)Vec_PtrEntry(p.vPats0, j), p.nPats0 ); }
            }
            p.nPats0++;
        
        } else {
            if p.nPats1 >= p.nPats {
                continue; }
            // inlining of macro Abc_NtkForEachPi
            for  j = 0; (j < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, j)), 1); j++  {
                if j == p.nTruePis {
                    break; }
                pInfo = (unsigned char *)Vec_PtrEntry( p.vPats, pObj.Id );
                assert(pInfo[i] == 0 || pInfo[i] == 0xff, "pInfo[i] == 0 || pInfo[i] == 0xff");
                if pInfo[i] {
                    Abc_InfoSetBit( (unsigned *)Vec_PtrEntry(p.vPats1, j), p.nPats1 ); }
            }
            p.nPats1++;
        }
        if p.nPats0 >= p.nPats && p.nPats1 >= p.nPats {
            break; }
    }
    if fVerbose != 0 {
        fmt.Printf("|  ");
        fmt.Printf("On = %3d  ", p.nPats1);
        fmt.Printf("Off = %3d  ", p.nPats0);
        fmt.Printf("\n");
    }
}

/**Function*************************************************************

  Synopsis    [Verifies the last pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimVerifyValue(p *Res_Sim_t, fOnSet int32) int32 {
    var pObj *Abc_Obj_t;
    var pInfo, pInfo2 *uint32;
    var i, value int32;
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(p.pAig)) && (((pObj) = Abc_NtkPi(p.pAig, i)), 1); i++  {
        if i == p.nTruePis {
            break; }
        if fOnSet != 0 {
            pInfo2 = (unsigned *)Vec_PtrEntry( p.vPats1, i );
            value = Abc_InfoHasBit( pInfo2, p.nPats1 - 1 );
        
        } else {
            pInfo2 = (unsigned *)Vec_PtrEntry( p.vPats0, i );
            value = Abc_InfoHasBit( pInfo2, p.nPats0 - 1 );
        }
        pInfo = (unsigned *)Vec_PtrEntry( p.vPats, pObj.Id );
        pInfo[0] = value ? ~0 : 0;
    }
    Res_SimPerformRound( p, 1 );
    pObj = Abc_NtkPo( p.pAig, 1 );
    pInfo = (unsigned *)Vec_PtrEntry( p.vPats, pObj.Id );
    assert(pInfo[0] == 0 || pInfo[0] == ~0, "pInfo[0] == 0 || pInfo[0] == ~0");
    return pInfo[0] > 0;
}

/**Function*************************************************************

  Synopsis    [Prepares simulation info for candidate filtering.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Res_SimPrepare(p *Res_Sim_t, pAig *Abc_Ntk_t, nTruePis int32, fVerbose int32) int32 {
    var i, nOnes, nZeros, nDcs int32;
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    // prepare the manager
    Res_SimAdjust( p, pAig, nTruePis );
    // estimate the number of patterns
    Res_SimSetRandomBytes( p );
    Res_SimPerformRound( p, p.nWordsIn );
    Res_SimCountResults( p, &nDcs, &nOnes, &nZeros, fVerbose );
    // collect the patterns
    Res_SimCollectPatterns( p, fVerbose );
    // add more patterns using constraint simulation
    if p.nPats0 < 8 {
        if !Res_SatSimulate( p, 16, 0 ) {
            return p.fConst0 || p.fConst1; }
//            return 0;
//        printf( "Value0 = %d\n", Res_SimVerifyValue( p, 0 ) );
    }
    if p.nPats1 < 8 {
        if !Res_SatSimulate( p, 16, 1 ) {
            return p.fConst0 || p.fConst1; }
//            return 0;
//        printf( "Value1 = %d\n", Res_SimVerifyValue( p, 1 ) );
    }
    // generate additional patterns
    for  i = 0; i < 2; i++  {
        if p.nPats0 > p.nPats*7/8 && p.nPats1 > p.nPats*7/8 {
            break; }
        Res_SimSetDerivedBytes( p, i==0 );
        Res_SimPerformRound( p, p.nWordsIn );
        Res_SimCountResults( p, &nDcs, &nOnes, &nZeros, fVerbose );
        Res_SimCollectPatterns( p, fVerbose );
    }
    // create bit-matrix info
    if p.nPats0 < p.nPats {
        Res_SimPadSimInfo( p.vPats0, p.nPats0, p.nWords ); }
    if p.nPats1 < p.nPats {
        Res_SimPadSimInfo( p.vPats1, p.nPats1, p.nWords ); }
    // resimulate 0-patterns
    Res_SimSetGiven( p, p.vPats0 );
    Res_SimPerformRound( p, p.nWords );
//Res_SimPrintNodePatterns( p, pAig );
    Res_SimDeriveInfoReplicate( p );
    // resimulate 1-patterns
    Res_SimSetGiven( p, p.vPats1 );
    Res_SimPerformRound( p, p.nWords );
//Res_SimPrintNodePatterns( p, pAig );
    Res_SimDeriveInfoComplement( p );
    // print output patterns
//    Res_SimPrintOutPatterns( p, pAig );
    return 1;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

