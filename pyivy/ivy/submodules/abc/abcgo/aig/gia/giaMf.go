package gia

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [giaMf.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Cut computation.]

  Author      [Alan Mishchenko]`
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMf.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/vec/vecMem.h"
//#include "misc/util/utilTruth.h"
//#include "misc/extra/extra.h"
//#include "sat/cnf/cnf.h"
//#include "opt/dau/dau.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Mf_SetForEachCut( pList, pCut, i )      for ( i = 0, pCut = pList + 1; i < pList[0]; i++, pCut += Mf_CutSize(pCut) + 1 )

// Original macro:
// #define Mf_ObjForEachCut( pCuts, i, nCuts )     for ( i = 0, i < nCuts; i++ )


// extern int Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computing truth tables of useful DSD classes of 6-functions.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var s_nCalls int32;
var s_vTtMem *Vec_Mem_t = nil;
func Mf_ManTruthCanonicize(t *word, nVars int32) int32 {
    var Temp word;
    var Best word = *t;
    var r, i, Config int32;
    for  r = 0; r < 1; r++  {
        if Best > (Temp = ~Best) {
            Best = Temp;
Config ^= (1 << nVars); }
        for  i = 0; i < nVars; i++  {
            if Best > (Temp = Abc_Tt6Flip(Best, i)) {
                Best = Temp;
Config ^= (1 << i); }
        }
    }
    *t = Best;
    if s_vTtMem == nil {
        s_vTtMem = Vec_MemAllocForTT( 6, 0 ); }
    Vec_MemHashInsert( s_vTtMem, t );
    s_nCalls++;
    return Config;
}
func Mf_ManTruthQuit() {
    if s_vTtMem == nil {
        return; }
    fmt.Printf("TT = %d (%.2f %%)\n", Vec_MemEntryNum(s_vTtMem), 100.0 * Vec_MemEntryNum(s_vTtMem) / s_nCalls);
    Vec_MemHashFree( s_vTtMem );
    Vec_MemFree( s_vTtMem );
    s_vTtMem = nil;
    s_nCalls = 0;
}

func Mf_ManTruthCollect(Limit int32) *Vec_Wrd_t {
    extern Vec_Wrd_t * Mpm_ManGetTruthWithCnf( int Limit );
    var pPerm *int32  = Extra_PermSchedule( 6 );
    var pComp *int32  = Extra_GreyCodeSchedule( 6 );
    var vTruths *Vec_Wrd_t  = Mpm_ManGetTruthWithCnf( Limit );
    var vResult *Vec_Wrd_t  = Vec_WrdAlloc( 1 << 20 );
    var uTruth, tCur, tTemp1, tTemp2 word;
    var i, p, c, k int32;
    Vec_WrdForEachEntry( vTruths, uTruth, k )
    {
        for  i = 0; i < 2; i++  {
            if i != 0 {
                tCur = ~uTruth;
            } else {
                tCur = uTruth;
            }
            tTemp1 = tCur;
            for  p = 0; p < 720; p++  {
                tTemp2 = tCur;
                for  c = 0; c < 64; c++  {
                    tCur = Abc_Tt6Flip( tCur, pComp[c] );
                    Vec_WrdPush( vResult, tCur );
                }
                assert(tTemp2 == tCur, "tTemp2 == tCur");
                tCur = Abc_Tt6SwapAdjacent( tCur, pPerm[p] );
            }
            assert(tTemp1 == tCur, "tTemp1 == tCur");
        }
    }
    pPerm = nil;
    pComp = nil;
    fmt.Printf("Original = %d.  ", Vec_WrdSize(vTruths));
    Vec_WrdFree( vTruths );
    fmt.Printf("Total = %d.  ", Vec_WrdSize(vResult));
    vTruths = Vec_WrdUniqifyHash( vResult, 1 );
    Vec_WrdFree( vResult );
    fmt.Printf("Unique = %d.  ", Vec_WrdSize(vTruths));
    Vec_WrdForEachEntry( vTruths, uTruth, k )
    {
        Mf_ManTruthCanonicize( &uTruth, 6 );
        Vec_WrdWriteEntry( vTruths, k, uTruth );
    }
    vResult = Vec_WrdUniqifyHash( vTruths, 1 );
    Vec_WrdFree( vTruths );
    fmt.Printf("Unique = %d.  \n", Vec_WrdSize(vResult));
    return vResult;
}
func Mf_ManTruthCount() int32 {
    var vTruths *Vec_Wrd_t  = Mf_ManTruthCollect( 10 );
    var RetValue int32  = Vec_WrdSize( vTruths );
    Vec_WrdFree( vTruths );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Collect truth tables used by the mapper.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mf_ManProfileTruths(p *Mf_Man_t) {
    var vCounts *Vec_Int_t;
    var i int32;
    var Entry int32;
    var pCut *int32;
    var Counter int32;
    vCounts = Vec_IntStart( Vec_IntSize(&p.vCnfSizes) );
    Gia_ManForEachAndId( p.pGia, i )
    {
        if !(Mf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Mf_ObjCutBest( p, i );
        Vec_IntAddToEntry( vCounts, Abc_Lit2Var(Mf_CutFunc(pCut)), 1 );
    }
    Vec_IntForEachEntry( vCounts, Entry, i )
    {
        if Entry == 0 {
            continue; }
        fmt.Printf("%6d : ", Counter++);
        fmt.Printf("%6d : ", i);
        fmt.Printf("Occur = %4d  ", Entry); 
        fmt.Printf("CNF size = %2d  ", Vec_IntEntry(&p.vCnfSizes, i));
        Dau_DsdPrintFromTruth( Vec_MemReadEntry(p.vTtMem, i), p.pPars.nLutSize );
    }
    Vec_IntFree( vCounts );
}

/**Function*************************************************************

  Synopsis    [Derives CNFs for each function used in the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Derives CNF for the AIG using the mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mf_ManDeriveCnf(p *Mf_Man_t, fCnfObjIds int32, fAddOrCla int32) *Cnf_Dat_t {
    var pCnf *Cnf_Dat_t; 
    var pObj *Gia_Obj_t;
    var Id int32;
    var DriId int32;
    var nVars int32;
    var nClas int32;
    var nLits int32;
    var iVar int32 = 1;
    var iCla int32;
    var iLit int32;
    var vCnfs *Vec_Int_t  = Mf_ManDeriveCnfs( p, &nVars, &nClas, &nLits );
    var vCnfIds *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p.pGia) );
    var pFanins [16]int32;
    var pCut *int32;
    var pCnfIds *int32 = Vec_IntArray( vCnfIds );
    var i int32;
    var k int32;
    var c int32;
    var iFunc int32;
    var nCubes int32;
    var pCubes *int32;
    var fComplLast int32;
    nVars++;  // zero-ID to remain unused
    if fAddOrCla != 0 {
        nClas++;
        nLits += Gia_ManCoNum(p.pGia);
    }
    // create CNF IDs
    if fCnfObjIds != 0 {
        iVar += 1 + Gia_ManCiNum(p.pGia) + Gia_ManCoNum(p.pGia);
        Gia_ManForEachCoId( p.pGia, Id, i )
            Vec_IntWriteEntry( vCnfIds, Id, Id );
        Gia_ManForEachAndReverseId( p.pGia, Id )
            if ( Mf_ObjMapRefNum(p, Id) )
                Vec_IntWriteEntry( vCnfIds, Id, Id );
iVar++;
        Vec_IntWriteEntry( vCnfIds, 0, 0 );
        Gia_ManForEachCiId( p.pGia, Id, i )
            Vec_IntWriteEntry( vCnfIds, Id, Id );
        assert(iVar == nVars, "iVar == nVars");
    
    } else {
        Gia_ManForEachCoId( p.pGia, Id, i )
            Vec_IntWriteEntry( vCnfIds, Id, iVar++ );
        Gia_ManForEachAndReverseId( p.pGia, Id )
            if ( Mf_ObjMapRefNum(p, Id) )
                Vec_IntWriteEntry( vCnfIds, Id, iVar++ );
        Vec_IntWriteEntry( vCnfIds, 0, iVar++ );
        Gia_ManForEachCiId( p.pGia, Id, i )
            Vec_IntWriteEntry( vCnfIds, Id, iVar++ );
        assert(iVar == nVars, "iVar == nVars");
    }
    // generate CNF
    pCnf = new(Cnf_Dat_t);
    pCnf.pMan        = (Aig_Man_t *)p.pGia;
    pCnf.nVars       = nVars;
    pCnf.nLiterals   = nLits;
    pCnf.nClauses    = nClas;
    pCnf.pClauses    = ABC_ALLOC( int *, nClas+1 );
    pCnf.pClauses[0] = ABC_ALLOC( int, nLits );
    // add last clause
    if fAddOrCla != 0 {
        pCnf.pClauses[iCla++] = pCnf.pClauses[0] + iLit;
        Gia_ManForEachCoId( p.pGia, Id, i )
            pCnf.pClauses[0][iLit++] = Abc_Var2Lit(pCnfIds[Id], 0);
    }
    if p.pPars.fCnfMapping {
        pCnf.vMapping = Vec_IntStart( nVars ); }
    // add clauses for the COs
    Gia_ManForEachCo( p.pGia, pObj, i )
    {
        Id = Gia_ObjId( p.pGia, pObj );
        DriId = Gia_ObjFaninId0( pObj, Id );

        pCnf.pClauses[iCla++] = pCnf.pClauses[0] + iLit;
        pCnf.pClauses[0][iLit++] = Abc_Var2Lit(pCnfIds[Id], 0);
        pCnf.pClauses[0][iLit++] = Abc_Var2Lit(pCnfIds[DriId], !Gia_ObjFaninC0(pObj));

        pCnf.pClauses[iCla++] = pCnf.pClauses[0] + iLit;
        pCnf.pClauses[0][iLit++] = Abc_Var2Lit(pCnfIds[Id], 1);
        pCnf.pClauses[0][iLit++] = Abc_Var2Lit(pCnfIds[DriId], Gia_ObjFaninC0(pObj));
        // generate mapping
        if pCnf.vMapping {
            Vec_IntWriteEntry( pCnf.vMapping, pCnfIds[Id], Vec_IntSize(pCnf.vMapping) );
            Vec_IntPush( pCnf.vMapping, 1 );
            Vec_IntPush( pCnf.vMapping, pCnfIds[DriId] );
            Vec_IntPush( pCnf.vMapping, Gia_ObjFaninC0(pObj) ? 0x55555555 : 0xAAAAAAAA );
        }
    }
    // add clauses for the mapping
    Gia_ManForEachAndReverseId( p.pGia, Id )
    {
        if !(Mf_ObjMapRefNum(p, Id) != 0) {
            continue; }
        pCut = Mf_ObjCutBest( p, Id );
        iFunc = Abc_Lit2Var( Mf_CutFunc(pCut) );
        fComplLast = Abc_LitIsCompl( Mf_CutFunc(pCut) );
        if iFunc == 0 { // constant cut
        
            pCnf.pClauses[iCla++] = pCnf.pClauses[0] + iLit;
            pCnf.pClauses[0][iLit++] = Abc_Var2Lit(pCnfIds[Id], !fComplLast);
            assert(pCnf.vMapping == nil, "pCnf.vMapping == nil"); // bug fix does not handle generated mapping
            continue;
        }
        for  k = 0; k < Mf_CutSize(pCut); k++  {
            pFanins[k] = pCnfIds[pCut[k+1]];
        }
        pFanins[k++] = pCnfIds[Id];
        // get clauses
        pCubes = Vec_IntEntryP( vCnfs, Vec_IntEntry(vCnfs, iFunc) );
        nCubes = *pCubes++;
        for  c = 0; c < nCubes; c++  {
            pCnf.pClauses[iCla++] = pCnf.pClauses[0] + iLit;
            k = Mf_CutSize(pCut);
            assert((Mf_CubeLit(pCubes[c], k) != 0), "Mf_CubeLit(pCubes[c], k)");
            pCnf.pClauses[0][iLit++] = Abc_Var2Lit( pFanins[k], (Mf_CubeLit(pCubes[c], k) == 2) ^ fComplLast );
            for  k = 0; k < Mf_CutSize(pCut); k++  {
                if Mf_CubeLit(pCubes[c], k) != 0 {
                    pCnf.pClauses[0][iLit++] = Abc_Var2Lit( pFanins[k], Mf_CubeLit(pCubes[c], k) == 2 ); }
            }
        }
        // generate mapping
        if pCnf.vMapping {
            var pTruth [4]word;
            var pTruthP *word = Vec_MemReadEntry(p.vTtMem, iFunc);
            assert(p.pPars.nLutSize <= 8, "p.pPars.nLutSize <= 8");
            Abc_TtCopy( pTruth, pTruthP, Abc_Truth6WordNum(p.pPars.nLutSize), Abc_LitIsCompl(iFunc) );
            assert(pCnfIds[Id] >= 0 && pCnfIds[Id] < nVars, "pCnfIds[Id] >= 0 && pCnfIds[Id] < nVars");
            Vec_IntWriteEntry( pCnf.vMapping, pCnfIds[Id], Vec_IntSize(pCnf.vMapping) );
            Vec_IntPush( pCnf.vMapping, Mf_CutSize(pCut) );
            for  k = 0; k < Mf_CutSize(pCut); k++  {
                Vec_IntPush( pCnf.vMapping, pCnfIds[pCut[k+1]] );
            }
            Vec_IntPush( pCnf.vMapping, uint32(pTruth[0]) );
            if Mf_CutSize(pCut) >= 6 {
                Vec_IntPush( pCnf.vMapping, uint32((pTruth[0] >> 32)) );
                if Mf_CutSize(pCut) >= 7 {
                    Vec_IntPush( pCnf.vMapping, uint32((pTruth[1])) );
                    Vec_IntPush( pCnf.vMapping, uint32((pTruth[1] >> 32)) );
                }
                if Mf_CutSize(pCut) >= 8 {
                    Vec_IntPush( pCnf.vMapping, uint32((pTruth[2])) );
                    Vec_IntPush( pCnf.vMapping, uint32((pTruth[2] >> 32)) );
                    Vec_IntPush( pCnf.vMapping, uint32((pTruth[3])) );
                    Vec_IntPush( pCnf.vMapping, uint32((pTruth[3] >> 32)) );
                }
            }
        }
    }
    // constant clause
    pCnf.pClauses[iCla++] = pCnf.pClauses[0] + iLit;
    pCnf.pClauses[0][iLit++] = Abc_Var2Lit(pCnfIds[0], 1);
    assert(iCla == nClas, "iCla == nClas");
    assert(iLit == nLits, "iLit == nLits");
    // add closing pointer
    pCnf.pClauses[iCla++] = pCnf.pClauses[0] + iLit;
    // cleanup
    Vec_IntFree( vCnfs );
    // create mapping of objects into their clauses
    if fCnfObjIds != 0 {
        pCnf.pObj2Clause = ABC_FALLOC( int, Gia_ManObjNum(p.pGia) );
        pCnf.pObj2Count  = ABC_FALLOC( int, Gia_ManObjNum(p.pGia) );
        for  i = 0; i < pCnf.nClauses; i++  {
            Id = Abc_Lit2Var(pCnf.pClauses[i][0]);
            if pCnf.pObj2Clause[Id] == -1 {
                pCnf.pObj2Clause[Id] = i;
                pCnf.pObj2Count[Id] = 1;
            
            } else {
                assert(pCnf.pObj2Count[Id] > 0, "pCnf.pObj2Count[Id] > 0");
                pCnf.pObj2Count[Id]++;
            }
        }
    
    } else {
        if p.pGia != p.pGia0 { // diff managers - create map for CIs/COs
        
            pCnf.pVarNums = ABC_FALLOC( int, Gia_ManObjNum(p.pGia0) );
            Gia_ManForEachCiId( p.pGia0, Id, i )
                pCnf.pVarNums[Id] = pCnfIds[Gia_ManCiIdToId(p.pGia, i)];
            Gia_ManForEachCoId( p.pGia0, Id, i )
                pCnf.pVarNums[Id] = pCnfIds[Gia_ManCoIdToId(p.pGia, i)];
/*
            // transform polarity of the internal nodes
            Gia_ManSetPhase( p->pGia );
            Gia_ManForEachCo( p->pGia, pObj, i )
                pObj->fPhase = 0;
            for ( i = 0; i < pCnf->nLiterals; i++ )
                if ( Gia_ManObj(p->pGia, Abc_Lit2Var(pCnf->pClauses[0][i]))->fPhase )
                    pCnf->pClauses[0][i] = Abc_LitNot( pCnf->pClauses[0][i] );
*/
        
        } else {
            pCnf.pVarNums = Vec_IntReleaseArray(vCnfIds); }
    }
    Vec_IntFree( vCnfIds );
    return pCnf;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



func Mf_CutPrint(p *Mf_Man_t, pCut *Mf_Cut_t) {
    var i int32;
    var nDigits int32 = Abc_Base10Log(Gia_ManObjNum(p.pGia)); 
    fmt.Printf("%d  {", pCut.nLeaves);
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        fmt.Printf(" %*d", nDigits, pCut.pLeaves[i]);
    }
    for  ; i < int32(p.pPars.nLutSize); i++  {
        fmt.Printf(" %*s", nDigits, " ");
    }
    fmt.Printf("  }   D = %4d  A = %9.4f  F = %6d  ", pCut.Delay, pCut.Flow, pCut.iFunc);
    if p.vTtMem {
        if p.pPars.fGenCnf {
            fmt.Printf("CNF = %2d  ", Vec_IntEntry(&p.vCnfSizes, Abc_Lit2Var(pCut.iFunc))); }
        if p.pPars.fGenLit {
            fmt.Printf("Lit = %2d  ", Vec_IntEntry(&p.vCnfSizes, Abc_Lit2Var(pCut.iFunc))); }
        Dau_DsdPrintFromTruth( Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut.iFunc)), pCut.nLeaves );
    
    } else {
        fmt.Printf("\n"); }
}




/**Function*************************************************************

  Synopsis    [Check correctness of cuts.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Mf_SetCheckArray(ppCuts **Mf_Cut_t, nCuts int32) int32 {
    var pCut0, pCut1 *Mf_Cut_t; 
    var i, k, m, n, Value int32;
    assert(nCuts > 0, "nCuts > 0");
    for  i = 0; i < nCuts; i++  {
        pCut0 = ppCuts[i];
        assert(pCut0.nLeaves <= MF_LEAF_MAX, "pCut0.nLeaves <= MF_LEAF_MAX");
        assert(pCut0.Sign == Mf_CutGetSign(pCut0.pLeaves, pCut0.nLeaves), "pCut0.Sign == Mf_CutGetSign(pCut0.pLeaves, pCut0.nLeaves)");
        // check duplicates
        for  m = 0; m < int32(pCut0.nLeaves); m++  {
        for  n = m + 1; n < int32(pCut0.nLeaves); n++  {
            assert(pCut0.pLeaves[m] < pCut0.pLeaves[n], "pCut0.pLeaves[m] < pCut0.pLeaves[n]");
        }
        }
        // check pairs
        for  k = 0; k < nCuts; k++  {
            pCut1 = ppCuts[k];
            if pCut0 == pCut1 {
                continue; }
            // check containments
            Value = Mf_CutCheck( pCut0, pCut1 );
            assert(Value == 0, "Value == 0");
        }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/











 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    [Flow and area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
















/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mf_ManPrintFanoutProfile(p *Mf_Man_t, vFanCounts *Vec_Int_t) {
    var pGia *Gia_Man_t  = p.pGia0;
    var i int32;
    var Count int32;
    var nMax int32 = Vec_IntFindMax( vFanCounts );
    var vCounts *Vec_Int_t  = Vec_IntStart( nMax + 1 );
    Vec_IntForEachEntry( vFanCounts, Count, i )
        if ( Count && Gia_ObjIsAnd(Gia_ManObj(pGia, i)) ) 
            Vec_IntAddToEntry( vCounts, Count, 1 );
    fmt.Printf("\nFanout distribution for internal nodes:\n");
    Vec_IntForEachEntry( vCounts, Count, i )
        if ( Count ) fmt.Printf("Fanout = %5d : Nodes = %5d.\n", i, Count);
    fmt.Printf("Total nodes with fanout = %d. Max fanout = %d.\n\n", Vec_IntCountPositive(vCounts), nMax);
    Vec_IntFree( vCounts );
}
func Mf_ManPrintMfccStats(p *Mf_Man_t, iObj int32) int32 {
    var pGia *Gia_Man_t  = p.pGia0;
    var Area int32;
    fmt.Printf("%5d : Level = %5d  Refs = %5d  Mffc = %5d\n", iObj, Gia_ObjLevelId(pGia, iObj), Mf_ObjMapRefNum(p, iObj), (Area = Mf_CutAreaMffc(p, iObj)));
    return Area;
}
func Mf_ManOptimizationOne(p *Mf_Man_t, iObj int32) {
    var pGia *Gia_Man_t  = p.pGia0;
    var ppCuts [32]*int32;
    var nCuts int32;
    var iFanout, i, nAreaSum, nAreaBest int32;
    // skip pivots whose MFFC fanouts are pointed to by COs
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        if ( Gia_ObjIsCo(Gia_ManObj(pGia, iFanout)) )
            return;
    // the pivot is used in the mapping as well as all of its fanouts
    assert(Mf_ObjMapRefNum(p, iObj) > 1, "Mf_ObjMapRefNum(p, iObj) > 1");
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        assert(Mf_ObjMapRefNum(p, iFanout) > 0, "Mf_ObjMapRefNum(p, iFanout) > 0");
    // print this pivot and its fanouts
    fmt.Printf("\nPivot node = %d\n", iObj);
    fmt.Printf("Pivot ");
Mf_ManPrintMfccStats( p, iObj );
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        fmt.Printf("Node  ");
nAreaSum += Mf_ManPrintMfccStats( p, iFanout );
    // calculate the shared MFFC
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        Mf_ObjMapRefInc( p, iFanout );
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        ppCuts[nCuts++] = Mf_ObjCutBest( p, iFanout );
    nAreaBest = Mf_CutAreaRefed2Multi( p, iObj, ppCuts, nCuts );
    Gia_ObjForEachFanoutStaticId( pGia, iObj, iFanout, i )
        Mf_ObjMapRefDec( p, iFanout );
    fmt.Printf("Sum of MFFC sizes = %d\n", nAreaSum);
    fmt.Printf("Shared MFFC size  = %d\n", nAreaBest);
}
func Mf_ManOptimization(p *Mf_Man_t) {
    var nOutMax int32  = 3;
    var pGia *Gia_Man_t  = p.pGia0;
    var i int32;
    var Count int32;
    var nNodes int32 = Mf_ManMappingFromMapping( p );
    Gia_ManLevelNum( pGia );
    Gia_ManStaticMappingFanoutStart( pGia );
    Mf_ManPrintFanoutProfile( p, pGia.vFanoutNums );
    fmt.Printf("\nIndividual logic cones for mapping with %d nodes:\n", nNodes);
    Vec_IntForEachEntry( pGia.vFanoutNums, Count, i )
        if ( Count >= 2 && Count <= nOutMax && Gia_ObjIsAnd(Gia_ManObj(pGia, i)) )
            Mf_ManOptimizationOne( p, i );
    fmt.Printf("\nFinished printing individual logic cones.\n");
    Gia_ManStaticFanoutStop( pGia );
    Vec_IntFreeP( &pGia.vMapping );
}

/**Function*************************************************************

  Synopsis    [Technology mappping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Mf_ManPerformMapping(pGia *Gia_Man_t, pPars *Jf_Par_t) *Gia_Man_t {
    var p *Mf_Man_t;
    var pNew, pCls *Gia_Man_t;
    if pPars.fGenCnf || pPars.fGenLit {
        pPars.fCutMin = 1; }
    if Gia_ManHasChoices(pGia) {
        pPars.fCutMin = 1, pPars.fCoarsen = 0; } 
    if pPars.fCoarsen {
        pCls = Gia_ManDupMuxes(pGia, pPars.nCoarseLimit);
    } else {
        pCls = pGia;
    }
    p = Mf_ManAlloc( pCls, pPars );
    p.pGia0 = pGia;
    if pPars.fVerbose && pPars.fCoarsen {
        fmt.Printf("Initial ");  Gia_ManPrintMuxStats( pGia );  fmt.Printf("\n");
        fmt.Printf("Derived ");  Gia_ManPrintMuxStats( pCls );  fmt.Printf("\n");
    }
    Mf_ManPrintInit( p );
    Mf_ManComputeCuts( p );
    for  p.Iter = 1; p.Iter < p.pPars.nRounds; p.Iter++  {
        Mf_ManComputeMapping( p );
    }
    p.fUseEla = 1;
    for  ; p.Iter < p.pPars.nRounds + pPars.nRoundsEla; p.Iter++  {
        Mf_ManComputeMapping( p );
    }
    //Mf_ManOptimization( p );
    if pPars.fVeryVerbose && pPars.fCutMin {
        Vec_MemDumpTruthTables( p.vTtMem, Gia_ManName(p.pGia), pPars.nLutSize ); }
    if pPars.fCutMin {
        pNew = Mf_ManDeriveMappingGia( p );
    } else if pPars.fCoarsen {
        pNew = Mf_ManDeriveMappingCoarse( p );
    } else {
        pNew = Mf_ManDeriveMapping( p ); }
    if p.pPars.fGenCnf {
        pGia.pData = Mf_ManDeriveCnf( p, p.pPars.fCnfObjIds, p.pPars.fAddOrCla ); }
    //if ( p->pPars->fGenCnf || p->pPars->fGenLit )
    //    Mf_ManProfileTruths( p );
    Gia_ManMappingVerify( pNew );
    Mf_ManPrintQuit( p, pNew );
    Mf_ManFree( p );
    if pCls != pGia {
        Gia_ManStop( pCls ); }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [CNF generation]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

