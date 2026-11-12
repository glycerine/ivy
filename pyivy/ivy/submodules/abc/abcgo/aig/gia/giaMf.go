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
func Mf_CutPrintOne(pCut *int32) {
    var i int32; 
    fmt.Printf("%d {", Mf_CutSize(pCut));
    for  i = 1; i <= Mf_CutSize(pCut); i++  {
        fmt.Printf(" %d", pCut[i]);
    }
    fmt.Printf(" }\n");
}
func Mf_CubeLit(Cube int32, iVar int32) int32 { return (Cube >> (iVar << 1)) & 3; }
func Mf_ManCountLits(pCnf *int32, nCubes int32, nVars int32) int32 {
    var i int32;
    var k int32;
    var nLits int32 = nCubes;
    for  i = 0; i < nCubes; i++  {
        for  k = 0; k < nVars; k++  {
            if Mf_CubeLit(pCnf[i], k) != 0 {
                nLits++; }
        }
    }
    return nLits;
}
func Mf_ManDeriveCnfs(p *Mf_Man_t, pnVars *int32, pnClas *int32, pnLits *int32) *Vec_Int_t {
    var i int32;
    var k int32;
    var iFunc int32;
    var nCubes int32;
    var nLits int32;
    var pCut *int32;
    var pCnf [512]int32;
    var vLits *Vec_Int_t  = Vec_IntStart( Vec_IntSize(&p.vCnfSizes) );
    var vCnfs *Vec_Int_t  = Vec_IntAlloc( 3 * Vec_IntSize(&p.vCnfSizes) );
    Vec_IntFill( vCnfs, Vec_IntSize(&p.vCnfSizes), -1 );
    assert(p.pPars.nLutSize <= 8, "p.pPars.nLutSize <= 8");
    // constant/buffer
    for  iFunc = 0; iFunc < 2; iFunc++  {
        if p.pPars.nLutSize <= 6 {
            nCubes = Abc_Tt6Cnf( *Vec_MemReadEntry(p.vTtMem, iFunc), iFunc, pCnf );
        } else {
            nCubes = Abc_Tt8Cnf( Vec_MemReadEntry(p.vTtMem, iFunc), iFunc, pCnf ); }
        nLits = Mf_ManCountLits( pCnf, nCubes, iFunc );
        Vec_IntWriteEntry( vLits, iFunc, nLits );
        Vec_IntWriteEntry( vCnfs, iFunc, Vec_IntSize(vCnfs) );
        Vec_IntPush( vCnfs, nCubes );
        for  k = 0; k < nCubes; k++  {
            Vec_IntPush( vCnfs, pCnf[k] );
        }
    }
    // other functions
    *pnVars = 1 + Gia_ManCiNum(p.pGia) + Gia_ManCoNum(p.pGia);
    *pnClas = 1 + 2 * Gia_ManCoNum(p.pGia);
    *pnLits = 1 + 4 * Gia_ManCoNum(p.pGia);
    Gia_ManForEachAndId( p.pGia, i )
    {
        if !(Mf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Mf_ObjCutBest( p, i );
        //Mf_CutPrintOne( pCut );
        iFunc = Abc_Lit2Var( Mf_CutFunc(pCut) );
        if Vec_IntEntry(vCnfs, iFunc) == -1 {
            if p.pPars.nLutSize <= 6 {
                nCubes = Abc_Tt6Cnf( *Vec_MemReadEntry(p.vTtMem, iFunc), Mf_CutSize(pCut), pCnf );
            } else {
                nCubes = Abc_Tt8Cnf( Vec_MemReadEntry(p.vTtMem, iFunc), Mf_CutSize(pCut), pCnf ); }
            assert(nCubes == Vec_IntEntry(&p.vCnfSizes, iFunc), "nCubes == Vec_IntEntry(&p.vCnfSizes, iFunc)");
            nLits = Mf_ManCountLits( pCnf, nCubes, Mf_CutSize(pCut) );
            // save CNF
            Vec_IntWriteEntry( vLits, iFunc, nLits );
            Vec_IntWriteEntry( vCnfs, iFunc, Vec_IntSize(vCnfs) );
            Vec_IntPush( vCnfs, nCubes );
            for  k = 0; k < nCubes; k++  {
                Vec_IntPush( vCnfs, pCnf[k] );
            }
        }
        *pnVars += 1;
        *pnClas += Vec_IntEntry(&p.vCnfSizes, iFunc);
        *pnLits += Vec_IntEntry(vLits, iFunc);
    }
    Vec_IntFree( vLits );
    return vCnfs;
}

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
func Mf_CutComputeTruth6(p *Mf_Man_t, pCut0 *Mf_Cut_t, pCut1 *Mf_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Mf_Cut_t, fIsXor int32) int32 {
//    extern int Mf_ManTruthCanonicize( word * t, int nVars );
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var fCompl int32; var t word;
    var t0 word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut0.iFunc));
    var t1 word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut1.iFunc));
    if Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 { t0 = ~t0; }
    if Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 { t1 = ~t1; }
    t0 = Abc_Tt6Expand( t0, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t1 = Abc_Tt6Expand( t1, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    if fIsXor != 0 {
        t = t0 ^ t1;
    } else {
        t = t0 & t1;
    }
    if fCompl = int32((t & 1)) { t = ~t; }
    if !p.pPars.fCnfObjIds {
    pCutR.nLeaves = Abc_Tt6MinBase( &t, pCutR.pLeaves, pCutR.nLeaves ); }
    assert(int32((t & 1)) == 0, "int32((t & 1)) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, &t);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    if (p.pPars.fGenCnf || p.pPars.fGenLit) && truthId == Vec_IntSize(&p.vCnfSizes) {
        Vec_IntPush( &p.vCnfSizes, p.pPars.fGenCnf ? Abc_Tt6CnfSize(t, pCutR.nLeaves) : Kit_TruthLitNum((unsigned *)&t, pCutR.nLeaves, &p.vCnfMem) ); }
//    p->nCutMux += Mf_ManTtIsMux( t );
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
//    Mf_ManTruthCanonicize( &t, pCutR->nLeaves );
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Mf_CutComputeTruth(p *Mf_Man_t, pCut0 *Mf_Cut_t, pCut1 *Mf_Cut_t, fCompl0 int32, fCompl1 int32, pCutR *Mf_Cut_t, fIsXor int32) int32 {
    if p.pPars.nLutSize <= 6 {
        return Mf_CutComputeTruth6( p, pCut0, pCut1, fCompl0, fCompl1, pCutR, fIsXor ); }
    {
    var uTruth, uTruth0, uTruth1 [MF_TT_WORDS]word;
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var LutSize int32 = p.pPars.nLutSize;
    var fCompl int32;
    var nWords int32      = Abc_Truth6WordNum(LutSize);
    var pTruth0 *word  = Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut0.iFunc));
    var pTruth1 *word  = Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut1.iFunc));
    Abc_TtCopy( uTruth0, pTruth0, nWords, Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 );
    Abc_TtCopy( uTruth1, pTruth1, nWords, Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 );
    Abc_TtExpand( uTruth0, LutSize, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtExpand( uTruth1, LutSize, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    if fIsXor != 0 {
        Abc_TtXor( uTruth, uTruth0, uTruth1, nWords, (fCompl = int32(((uTruth0[0] ^ uTruth1[0]) & 1))) );
    } else {
        Abc_TtAnd( uTruth, uTruth0, uTruth1, nWords, (fCompl = int32(((uTruth0[0] & uTruth1[0]) & 1))) ); }
    pCutR.nLeaves = Abc_TtMinBase( uTruth, pCutR.pLeaves, pCutR.nLeaves, LutSize );
    assert((uTruth[0] & 1) == 0, "(uTruth[0] & 1) == 0");
//Kit_DsdPrintFromTruth( uTruth, pCutR->nLeaves ), printf("\n" ), printf("\n" );
    truthId        = Vec_MemHashInsert(p.vTtMem, uTruth);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    if (p.pPars.fGenCnf || p.pPars.fGenLit) && truthId == Vec_IntSize(&p.vCnfSizes) && LutSize <= 8 {
        Vec_IntPush( &p.vCnfSizes, p.pPars.fGenCnf ? Abc_Tt8CnfSize(uTruth, pCutR.nLeaves) : Kit_TruthLitNum((unsigned *)uTruth, pCutR.nLeaves, &p.vCnfMem) ); }
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
    }
}
func Mf_CutComputeTruthMux6(p *Mf_Man_t, pCut0 *Mf_Cut_t, pCut1 *Mf_Cut_t, pCutC *Mf_Cut_t, fCompl0 int32, fCompl1 int32, fComplC int32, pCutR *Mf_Cut_t) int32 {
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var fCompl int32; var t word;
    var t0 word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut0.iFunc));
    var t1 word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut1.iFunc));
    var tC word  = *Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCutC.iFunc));
    if Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 { t0 = ~t0; }
    if Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 { t1 = ~t1; }
    if Abc_LitIsCompl(pCutC.iFunc) ^ fComplC { tC = ~tC; }
    t0 = Abc_Tt6Expand( t0, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t1 = Abc_Tt6Expand( t1, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    tC = Abc_Tt6Expand( tC, pCutC.pLeaves, pCutC.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    t = (tC & t1) | (~tC & t0);
    if fCompl = int32((t & 1)) { t = ~t; }
    pCutR.nLeaves = Abc_Tt6MinBase( &t, pCutR.pLeaves, pCutR.nLeaves );
    assert(int32((t & 1)) == 0, "int32((t & 1)) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, &t);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    if (p.pPars.fGenCnf || p.pPars.fGenLit) && truthId == Vec_IntSize(&p.vCnfSizes) {
        Vec_IntPush( &p.vCnfSizes, p.pPars.fGenCnf ? Abc_Tt6CnfSize(t, pCutR.nLeaves) : Kit_TruthLitNum((unsigned *)&t, pCutR.nLeaves, &p.vCnfMem) ); }
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
}
func Mf_CutComputeTruthMux(p *Mf_Man_t, pCut0 *Mf_Cut_t, pCut1 *Mf_Cut_t, pCutC *Mf_Cut_t, fCompl0 int32, fCompl1 int32, fComplC int32, pCutR *Mf_Cut_t) int32 {
    if p.pPars.nLutSize <= 6 {
        return Mf_CutComputeTruthMux6( p, pCut0, pCut1, pCutC, fCompl0, fCompl1, fComplC, pCutR ); }
    {
    var uTruth, uTruth0, uTruth1, uTruthC [MF_TT_WORDS]word;
    var nOldSupp int32 = pCutR.nLeaves;
    var truthId int32;
    var LutSize int32 = p.pPars.nLutSize;
    var fCompl int32;
    var nWords int32      = Abc_Truth6WordNum(LutSize);
    var pTruth0 *word  = Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut0.iFunc));
    var pTruth1 *word  = Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCut1.iFunc));
    var pTruthC *word  = Vec_MemReadEntry(p.vTtMem, Abc_Lit2Var(pCutC.iFunc));
    Abc_TtCopy( uTruth0, pTruth0, nWords, Abc_LitIsCompl(pCut0.iFunc) ^ fCompl0 );
    Abc_TtCopy( uTruth1, pTruth1, nWords, Abc_LitIsCompl(pCut1.iFunc) ^ fCompl1 );
    Abc_TtCopy( uTruthC, pTruthC, nWords, Abc_LitIsCompl(pCutC.iFunc) ^ fComplC );
    Abc_TtExpand( uTruth0, LutSize, pCut0.pLeaves, pCut0.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtExpand( uTruth1, LutSize, pCut1.pLeaves, pCut1.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtExpand( uTruthC, LutSize, pCutC.pLeaves, pCutC.nLeaves, pCutR.pLeaves, pCutR.nLeaves );
    Abc_TtMux( uTruth, uTruthC, uTruth1, uTruth0, nWords );
    fCompl         = int32((uTruth[0] & 1));
    if fCompl != 0 { Abc_TtNot( uTruth, nWords ); }
    pCutR.nLeaves = Abc_TtMinBase( uTruth, pCutR.pLeaves, pCutR.nLeaves, LutSize );
    assert((uTruth[0] & 1) == 0, "(uTruth[0] & 1) == 0");
    truthId        = Vec_MemHashInsert(p.vTtMem, uTruth);
    pCutR.iFunc   = Abc_Var2Lit( truthId, fCompl );
    if (p.pPars.fGenCnf || p.pPars.fGenLit) && truthId == Vec_IntSize(&p.vCnfSizes) && LutSize <= 8 {
        Vec_IntPush( &p.vCnfSizes, p.pPars.fGenCnf ? Abc_Tt8CnfSize(uTruth, pCutR.nLeaves) : Kit_TruthLitNum((unsigned *)uTruth, pCutR.nLeaves, &p.vCnfMem) ); }
    assert(int32(pCutR.nLeaves) <= nOldSupp, "int32(pCutR.nLeaves) <= nOldSupp");
    return int32(pCutR.nLeaves) < nOldSupp;
    }
}


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










func Mf_ObjMergeOrder(p *Mf_Man_t, iObj int32) {
    var pCuts0 [MF_CUT_MAX]Mf_Cut_t;
    var pCuts1 [MF_CUT_MAX]Mf_Cut_t;
    var pCuts [MF_CUT_MAX]Mf_Cut_t;
    var pCutsR [MF_CUT_MAX]*Mf_Cut_t;
    var pObj *Gia_Obj_t  = Gia_ManObj(p.pGia, iObj);
    var pBest *Mf_Obj_t  = Mf_ManObj(p, iObj);
    var nLutSize int32  = p.pPars.nLutSize;
    var nCutNum int32   = p.pPars.nCutNum;
    var nCuts0 int32    = Mf_ManPrepareCuts(pCuts0, p, Gia_ObjFaninId0(pObj, iObj), 1);
    var nCuts1 int32    = Mf_ManPrepareCuts(pCuts1, p, Gia_ObjFaninId1(pObj, iObj), 1);
    var fComp0 int32    = Gia_ObjFaninC0(pObj);
    var fComp1 int32    = Gia_ObjFaninC1(pObj);
    var iSibl int32     = Gia_ObjSibl(p.pGia, iObj);
    var pCut0 *Mf_Cut_t;
    var pCut1 *Mf_Cut_t;
    var pCut0Lim *Mf_Cut_t = pCuts0 + nCuts0;
    var pCut1Lim *Mf_Cut_t = pCuts1 + nCuts1;
    var i, nCutsR int32;
    for  i = 0; i < nCutNum; i++  {
        pCutsR[i] = pCuts + i;
    }
    if iSibl != 0 {
        var pCuts2 [MF_CUT_MAX]Mf_Cut_t;
        var pObjE *Gia_Obj_t  = Gia_ObjSiblObj(p.pGia, iObj);
        var fCompE int32  = Gia_ObjPhase(pObj) ^ Gia_ObjPhase(pObjE);
        var nCuts2 int32  = Mf_ManPrepareCuts(pCuts2, p, iSibl, 0);
        var pCut2 *Mf_Cut_t;
        var pCut2Lim *Mf_Cut_t = pCuts2 + nCuts2;
        for  pCut2 = pCuts2; pCut2 < pCut2Lim; pCut2++  {
            *pCutsR[nCutsR] = *pCut2;
            if pCutsR[nCutsR].iFunc >= 0 {
                pCutsR[nCutsR].iFunc = Abc_LitNotCond( pCutsR[nCutsR].iFunc, fCompE ); }
            Mf_CutParams( p, pCutsR[nCutsR], pBest.nFlowRefs );
            nCutsR = Mf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
    }
    if Gia_ObjIsMuxId(p.pGia, iObj) {
        var pCuts2 [MF_CUT_MAX]Mf_Cut_t;
        var nCuts2 int32   = Mf_ManPrepareCuts(pCuts2, p, Gia_ObjFaninId2(p.pGia, iObj), 1);
        var fComp2 int32   = Gia_ObjFaninC2(p.pGia, pObj);
        var pCut2 *Mf_Cut_t;
        var pCut2Lim *Mf_Cut_t = pCuts2 + nCuts2;
        p.CutCount[0] += nCuts0 * nCuts1 * nCuts2;
        for  pCut0 = pCuts0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCuts1; pCut1 < pCut1Lim; pCut1++  {
        for  pCut2 = pCuts2; pCut2 < pCut2Lim; pCut2++  {
            if Mf_CutCountBits(pCut0.Sign | pCut1.Sign | pCut2.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Mf_CutMergeOrderMux(pCut0, pCut1, pCut2, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Mf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if p.pPars.fCutMin && (Mf_CutComputeTruthMux(p, pCut0, pCut1, pCut2, fComp0, fComp1, fComp2, pCutsR[nCutsR]) != 0) {
                pCutsR[nCutsR].Sign = Mf_CutGetSign(pCutsR[nCutsR].pLeaves, pCutsR[nCutsR].nLeaves); }
            Mf_CutParams( p, pCutsR[nCutsR], pBest.nFlowRefs );
            nCutsR = Mf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        }
        }
    
    } else {
        var fIsXor int32  = Gia_ObjIsXor(pObj);
        p.CutCount[0] += nCuts0 * nCuts1;
        for  pCut0 = pCuts0; pCut0 < pCut0Lim; pCut0++  {
        for  pCut1 = pCuts1; pCut1 < pCut1Lim; pCut1++  {
            if int32((pCut0.nLeaves + pCut1.nLeaves)) > nLutSize && Mf_CutCountBits(pCut0.Sign | pCut1.Sign) > nLutSize {
                continue; }
            p.CutCount[1]++; 
            if !(Mf_CutMergeOrder(pCut0, pCut1, pCutsR[nCutsR], nLutSize) != 0) {
                continue; }
            if Mf_SetLastCutIsContained(pCutsR, nCutsR) != 0 {
                continue; }
            p.CutCount[2]++;
            if p.pPars.fCutMin && (Mf_CutComputeTruth(p, pCut0, pCut1, fComp0, fComp1, pCutsR[nCutsR], fIsXor) != 0) {
                pCutsR[nCutsR].Sign = Mf_CutGetSign(pCutsR[nCutsR].pLeaves, pCutsR[nCutsR].nLeaves); }
            Mf_CutParams( p, pCutsR[nCutsR], pBest.nFlowRefs );
            nCutsR = Mf_SetAddCut( pCutsR, nCutsR, nCutNum );
        }
        }
    }
    // debug printout
    if false {
//    if ( iObj % 1000 == 0 )
//    if ( iObj == 474 )
    
        fmt.Printf("*** Obj = %d  FlowRefs = %.2f  MapRefs = %2d\n", iObj, pBest.nFlowRefs, pBest.nMapRefs);
        for  i = 0; i < nCutsR; i++  {
            Mf_CutPrint( p, pCutsR[i] );
        }
        fmt.Printf("\n");
    } 
    // store the cutset
    pBest.Flow = pCutsR[0].Flow;
    pBest.Delay = pCutsR[0].Delay;
    pBest.iCutSet = Mf_ManSaveCuts( p, pCutsR, nCutsR );
    // verify
    assert(nCutsR > 0 && nCutsR < nCutNum, "nCutsR > 0 && nCutsR < nCutNum");
//    assert( Mf_SetCheckArray(pCutsR, nCutsR) );
    p.nCutCounts[pCutsR[0].nLeaves]++;
    p.CutCount[3] += nCutsR;
}
 

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mf_ManSetFlowRefs(p *Gia_Man_t, vRefs *Vec_Int_t) {
    var fDiscount int32  = 1;
    var pObj, pCtrl, pData0, pData1 *Gia_Obj_t; 
    var i, Id int32;
    Vec_IntFill( vRefs, Gia_ManObjNum(p), 0 );
    Gia_ManForEachAnd( p, pObj, i )
    {
        if Gia_ObjIsAnd(Gia_ObjFanin0(pObj)) {
            Vec_IntAddToEntry( vRefs, Gia_ObjFaninId0(pObj, i), 1 ); }
        if Gia_ObjIsAnd(Gia_ObjFanin1(pObj)) {
            Vec_IntAddToEntry( vRefs, Gia_ObjFaninId1(pObj, i), 1 ); }
        if p.pMuxes {
            if Gia_ObjIsMuxId(p, i) && Gia_ObjIsAnd(Gia_ObjFanin2(p, pObj)) {
                Vec_IntAddToEntry( vRefs, Gia_ObjFaninId2(p, i), 1 ); }
        
        } else if (fDiscount != 0) && Gia_ObjIsMuxType(pObj) { // discount XOR/MUX
        
            pCtrl  = Gia_Regular(Gia_ObjRecognizeMux(pObj, &pData1, &pData0));
            pData0 = Gia_Regular(pData0);
            pData1 = Gia_Regular(pData1);
            if Gia_ObjIsAnd(pCtrl) {
                Vec_IntAddToEntry( vRefs, Gia_ObjId(p, pCtrl), -1 ); }
            if pData0 == pData1 && Gia_ObjIsAnd(pData0) {
                Vec_IntAddToEntry( vRefs, Gia_ObjId(p, pData0), -1 ); }
        }
    }
    Gia_ManForEachCoDriverId( p, Id, i )
        if ( Gia_ObjIsAnd(Gia_ManObj(p, Id)) )
            Vec_IntAddToEntry( vRefs, Id, 1 );
    for  i = 0; i < Vec_IntSize(vRefs); i++  {
        Vec_IntUpdateEntry( vRefs, i, 1 );
    }
}
func Mf_ManSetMapRefs(p *Mf_Man_t) int32 {
    var Coef float32  = 1.0 / (1.0 + (p.Iter + 1) * (p.Iter + 1));
    var pCut *int32;
    var i int32;
    var k int32;
    var Id int32;
    // compute delay
    var Delay int32;
    Gia_ManForEachCoDriverId( p.pGia, Id, i )
        Delay = Abc_MaxInt( Delay, Mf_ManObj(p, Id).Delay );
    // check delay target
    if p.pPars.DelayTarget == -1 && p.pPars.nRelaxRatio {
        p.pPars.DelayTarget = int32((float32(Delay) * (100.0 + p.pPars.nRelaxRatio) / 100.0)); }
    if p.pPars.DelayTarget != -1 {
        if Delay < p.pPars.DelayTarget + 0.01 {
            Delay = p.pPars.DelayTarget;
        } else if p.pPars.nRelaxRatio == 0 {
            Abc_Print( 0, "Relaxing user-specified delay target from %d to %d.\n", p.pPars.DelayTarget, Delay ); }
    }
    p.pPars.Delay = Delay;
    // check references
//    Gia_ManForEachAndId( p->pGia, i )
//        assert( Mf_ManObj(p, i)->nMapRefs == 0 );
    // compute area and edges
    if !p.fUseEla {
        Gia_ManForEachCoDriverId( p.pGia, Id, i )
            Mf_ObjMapRefInc( p, Id ); }
    p.pPars.Area = p.pPars.Edge = p.pPars.Clause = 0;
    Gia_ManForEachAndReverseId( p.pGia, i )
    {
        if !(Mf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Mf_ObjCutBest( p, i );
        if !p.fUseEla {
            for  k = 1; k <= Mf_CutSize(pCut); k++  {
                Mf_ObjMapRefInc( p, pCut[k] );
            } }
        p.pPars.Edge += Mf_CutSize(pCut);
        p.pPars.Area++;
        if p.pPars.fGenCnf || p.pPars.fGenLit {
            p.pPars.Clause += Mf_CutArea(p, Mf_CutSize(pCut), Mf_CutFunc(pCut)); }
    }
    // blend references
    for  i = 0; i < Gia_ManObjNum(p.pGia); i++  {
        p.pLfObjs[i].nFlowRefs = Coef * p.pLfObjs[i].nFlowRefs + (1.0 - Coef) * Abc_MaxFloat(1, p.pLfObjs[i].nMapRefs);
    }
//        p->pLfObjs[i]. = 0.2 * p->pLfObjs[i]. + 0.8 * Abc_MaxFloat(1, p->pLfObjs[i].nMapRefs);
    return p.pPars.Area;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

func Mf_ManDeriveMappingCoarse(p *Mf_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pGia *Gia_Man_t = p.pGia;
    var pObj *Gia_Obj_t;
    var i int32;
    var k int32;
    var pCut *int32;
    assert(!p.pPars.fCutMin && pGia.pMuxes, "!p.pPars.fCutMin && pGia.pMuxes");
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(pGia) );
    pNew.pName = Abc_UtilStrsav( pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( pGia.pSpec );
    // map primary inputs
    Gia_ManConst0(pGia).Value = 0;
    Gia_ManForEachCi( pGia, pObj, i )
        pObj.Value = Gia_ManAppendCi( pNew );
    // start mapping
    pNew.vMapping = Vec_IntAlloc( Gia_ManObjNum(pGia) + 2*Gia_ManXorNum(pGia) + 2*Gia_ManMuxNum(pGia) + int32(p.pPars.Edge) + int32(p.pPars.Area) * 2 );
    Vec_IntFill( pNew.vMapping, Gia_ManObjNum(pGia) + 2*Gia_ManXorNum(pGia) + 2*Gia_ManMuxNum(pGia), 0 );
    // iterate through nodes used in the mapping
    Gia_ManForEachAnd( pGia, pObj, i )
    {
        if Gia_ObjIsMuxId(pGia, i) {
            pObj.Value = Gia_ManAppendMux( pNew, Gia_ObjFanin2Copy(pGia, pObj), Gia_ObjFanin1Copy(pObj), Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsXor(pObj) {
            pObj.Value = Gia_ManAppendXor( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else { 
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
        if !(Mf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Mf_ObjCutBest( p, i );
        Vec_IntWriteEntry( pNew.vMapping, Abc_Lit2Var(pObj.Value), Vec_IntSize(pNew.vMapping) );
        Vec_IntPush( pNew.vMapping, Mf_CutSize(pCut));
        for  k = 1; k <= Mf_CutSize(pCut); k++  {
            Vec_IntPush( pNew.vMapping, Abc_Lit2Var(Gia_ManObj(pGia, pCut[k]).Value) );
        }
        Vec_IntPush( pNew.vMapping, Abc_Lit2Var(pObj.Value) );
    }
    Gia_ManForEachCo( pGia, pObj, i )
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(pGia) );
    assert(Vec_IntCap(pNew.vMapping) == 16 || Vec_IntSize(pNew.vMapping) == Vec_IntCap(pNew.vMapping), "Vec_IntCap(pNew.vMapping) == 16 || Vec_IntSize(pNew.vMapping) == Vec_IntCap(pNew.vMapping)");
    return pNew;
}
func Mf_ManDeriveMappingGia(p *Mf_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t; 
    var pObj *Gia_Obj_t; 
    var vCopies *Vec_Int_t    = Vec_IntStartFull( Gia_ManObjNum(p.pGia) );
    var vMapping *Vec_Int_t   = Vec_IntStart( 2 * Gia_ManObjNum(p.pGia) + int32(p.pPars.Edge) + 2 * int32(p.pPars.Area) );
    var vMapping2 *Vec_Int_t  = Vec_IntStart( int32(p.pPars.Edge) + 2 * int32(p.pPars.Area) + 1000 );
    var vCover *Vec_Int_t     = Vec_IntAlloc( 1 << 16 );
    var vLeaves *Vec_Int_t    = Vec_IntAlloc( 16 );
    var i int32;
    var k int32;
    var Id int32;
    var iLit int32;
    var pCut *int32;
    var uTruth word;
    var pTruth *word = &uTruth;
    assert(p.pPars.fCutMin, "p.pPars.fCutMin");
    // create new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p.pGia) );
    pNew.pName = Abc_UtilStrsav( p.pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pGia.pSpec );
    // map primary inputs
    Vec_IntWriteEntry( vCopies, 0, 0 );
    Gia_ManForEachCiId( p.pGia, Id, i )
        Vec_IntWriteEntry( vCopies, Id, Gia_ManAppendCi(pNew) );
    // iterate through nodes used in the mapping
    Gia_ManForEachAnd( p.pGia, pObj, i )
    {
        if !(Mf_ObjMapRefNum(p, i) != 0) {
            continue; }
        pCut = Mf_ObjCutBest( p, i );
        if Mf_CutSize(pCut) == 0 {
            assert(Abc_Lit2Var(Mf_CutFunc(pCut)) == 0, "Abc_Lit2Var(Mf_CutFunc(pCut)) == 0");
            Vec_IntWriteEntry( vCopies, i, Mf_CutFunc(pCut) );
            continue;
        }
        if Mf_CutSize(pCut) == 1 {
            assert(Abc_Lit2Var(Mf_CutFunc(pCut)) == 1, "Abc_Lit2Var(Mf_CutFunc(pCut)) == 1");
            iLit = Vec_IntEntry( vCopies, pCut[1] );
            Vec_IntWriteEntry( vCopies, i, Abc_LitNotCond(iLit, Abc_LitIsCompl(Mf_CutFunc(pCut))) );
            continue;
        }
        Vec_IntClear( vLeaves );
        for  k = 1; k <= Mf_CutSize(pCut); k++  {
            Vec_IntPush( vLeaves, Vec_IntEntry(vCopies, pCut[k]) );
        }
        pTruth = Vec_MemReadEntry( p.vTtMem, Abc_Lit2Var(Mf_CutFunc(pCut)) );
        iLit = Kit_TruthToGia( pNew, (unsigned *)pTruth, Vec_IntSize(vLeaves), vCover, vLeaves, 0 );
        Vec_IntWriteEntry( vCopies, i, Abc_LitNotCond(iLit, Abc_LitIsCompl(Mf_CutFunc(pCut))) );
        // create mapping
        Vec_IntSetEntry( vMapping, Abc_Lit2Var(iLit), Vec_IntSize(vMapping2) );
        Vec_IntPush( vMapping2, Vec_IntSize(vLeaves) );
        Vec_IntForEachEntry( vLeaves, iLit, k )
            Vec_IntPush( vMapping2, Abc_Lit2Var(iLit) );
        Vec_IntPush( vMapping2, Abc_Lit2Var(Vec_IntEntry(vCopies, i)) );
    }
    Gia_ManForEachCo( p.pGia, pObj, i )
    {
        iLit = Vec_IntEntry( vCopies, Gia_ObjFaninId0p(p.pGia, pObj) );
        iLit = Gia_ManAppendCo( pNew, Abc_LitNotCond(iLit, Gia_ObjFaninC0(pObj)) );
    }
    Vec_IntFree( vCopies );
    Vec_IntFree( vCover );
    Vec_IntFree( vLeaves );
    // finish mapping 
    if Vec_IntSize(vMapping) > Gia_ManObjNum(pNew) {
        Vec_IntShrink( vMapping, Gia_ManObjNum(pNew) );
    } else {
        Vec_IntFillExtra( vMapping, Gia_ManObjNum(pNew), 0 ); }
    assert(Vec_IntSize(vMapping) == Gia_ManObjNum(pNew), "Vec_IntSize(vMapping) == Gia_ManObjNum(pNew)");
    Vec_IntForEachEntry( vMapping, iLit, i )
        if ( iLit > 0 )
            Vec_IntAddToEntry( vMapping, i, Gia_ManObjNum(pNew) );
    Vec_IntAppend( vMapping, vMapping2 );
    Vec_IntFree( vMapping2 );
    // attach mapping and packing
    assert(pNew.vMapping == nil, "pNew.vMapping == nil");
    pNew.vMapping = vMapping;
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p.pGia) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mf_ManAlloc(pGia *Gia_Man_t, pPars *Jf_Par_t) *Mf_Man_t {
    var p *Mf_Man_t;
    var vFlowRefs *Vec_Int_t;
    var i, Entry int32;
    assert(pPars.nCutNum > 1  && pPars.nCutNum <= MF_CUT_MAX, "pPars.nCutNum > 1  && pPars.nCutNum <= MF_CUT_MAX");
    assert(pPars.nLutSize > 1 && pPars.nLutSize <= MF_LEAF_MAX, "pPars.nLutSize > 1 && pPars.nLutSize <= MF_LEAF_MAX");
    pGia.Refs = nil;
    Vec_IntFreeP( &pGia.vMapping );
    if Gia_ManHasChoices(pGia) {
        Gia_ManSetPhase(pGia); }
    p = new(Mf_Man_t);
    p.clkStart  = Abc_Clock();
    p.pGia      = pGia;
    p.pPars     = pPars;
    if pPars.fCutMin {
        p.vTtMem = Vec_MemAllocForTT( pPars.nLutSize, 0 );
    } else {
        p.vTtMem = nil;
    }
    p.pLfObjs   = ABC_CALLOC( Mf_Obj_t, Gia_ManObjNum(pGia) );
    p.iCur      = 2;
    Vec_PtrGrow( &p.vPages, 256 );
    if pPars.fGenCnf || pPars.fGenLit {
        Vec_IntGrow( &p.vCnfSizes, 10000 );
        Vec_IntPush( &p.vCnfSizes, 1 );
        Vec_IntPush( &p.vCnfSizes, 2 );
        Vec_IntGrow( &p.vCnfMem, 10000 );
    }
    vFlowRefs = Vec_IntAlloc(0);
    Mf_ManSetFlowRefs( pGia, vFlowRefs );
    Vec_IntForEachEntry( vFlowRefs, Entry, i )
        p.pLfObjs[i].nFlowRefs = Entry;
    Vec_IntFree(vFlowRefs);
    return p;
}
func Mf_ManFree(p *Mf_Man_t) {
    assert(!p.pPars.fGenCnf || !p.pPars.fGenLit || Vec_IntSize(&p.vCnfSizes) == Vec_MemEntryNum(p.vTtMem), "!p.pPars.fGenCnf || !p.pPars.fGenLit || Vec_IntSize(&p.vCnfSizes) == Vec_MemEntryNum(p.vTtMem)");
    if p.pPars.fCutMin {
        Vec_MemHashFree( p.vTtMem ); }
    if p.pPars.fCutMin {
        Vec_MemFree( p.vTtMem ); }
    Vec_PtrFreeData( &p.vPages );
    p.vCnfSizes.pArray = nil;
    p.vCnfMem.pArray = nil;
    p.vPages.pArray = nil;
    p.vTemp.pArray = nil;
    p.pLfObjs = nil;
    p = nil;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Mf_ManSetDefaultPars(pPars *Jf_Par_t) {
    memset( pPars, 0, sizeof(Jf_Par_t) );
    pPars.nLutSize     =  6;
    pPars.nCutNum      =  8;
    pPars.nProcNum     =  0;
    pPars.nRounds      =  2;
    pPars.nRoundsEla   =  1;
    pPars.nRelaxRatio  =  0;
    pPars.nCoarseLimit =  3;
    pPars.nAreaTuner   =  1;
    pPars.nVerbLimit   =  5;
    pPars.DelayTarget  = -1;
    pPars.fAreaOnly    =  0;
    pPars.fOptEdge     =  1; 
    pPars.fCoarsen     =  1;
    pPars.fCutMin      =  0;
    pPars.fGenCnf      =  0;
    pPars.fGenLit      =  0;
    pPars.fPureAig     =  0;
    pPars.fVerbose     =  0;
    pPars.fVeryVerbose =  0;
    pPars.nLutSizeMax  =  MF_LEAF_MAX;
    pPars.nCutNumMax   =  MF_CUT_MAX;
}
func Mf_ManPrintStats(p *Mf_Man_t, pTitle string) {
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("%s :  ", pTitle);
    fmt.Printf("Level =%6d   ", int64(p.pPars.Delay));
    fmt.Printf("Area =%9d   ", int64(p.pPars.Area));
    fmt.Printf("Edge =%9d   ", int64(p.pPars.Edge));
    if p.pPars.fGenCnf {
        fmt.Printf("CNF =%9d   ", int64(p.pPars.Clause)); }
    if p.pPars.fGenLit {
        fmt.Printf("FFL =%9d   ", int64(p.pPars.Clause)); }
    Abc_PrintTime( 1, "Time", Abc_Clock() - p.clkStart );
    fflush( os.Stdout );
}
func Mf_ManPrintInit(p *Mf_Man_t) {
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("LutSize = %d  ", p.pPars.nLutSize);
    fmt.Printf("CutNum = %d  ", p.pPars.nCutNum);
    fmt.Printf("Iter = %d  ", p.pPars.nRounds + p.pPars.nRoundsEla);
    fmt.Printf("Edge = %d  ", p.pPars.fOptEdge);
    fmt.Printf("CutMin = %d  ", p.pPars.fCutMin);
    fmt.Printf("Coarse = %d  ", p.pPars.fCoarsen);
    fmt.Printf("CNF = %d  ", p.pPars.fGenCnf);
    fmt.Printf("FFL = %d  ", p.pPars.fGenLit);
    fmt.Printf("\n");
    fmt.Printf("Computing cuts...\r");
    fflush( os.Stdout );
}
func Mf_ManPrintQuit(p *Mf_Man_t, pNew *Gia_Man_t) {
    var MemGia float32    = Gia_ManMemory(p.pGia) / (1<<20);
    var MemMan float32    = 1.0 * sizeof(Mf_Obj_t) * Gia_ManObjNum(p.pGia) / (1<<20);
    var MemCuts float32   = 1.0 * sizeof(int) * (1 << 16) * Vec_PtrSize(&p.vPages) / (1<<20);
    var MemTt float32;
    if p.vTtMem {
        MemTt = Vec_MemMemory(p.vTtMem) / (1<<20);
    } else {
        MemTt = 0;
    }
    var MemMap float32    = Vec_IntMemory(pNew.vMapping) / (1<<20);
    if p.CutCount[0] == 0 {
        p.CutCount[0] = 1; }
    if !p.pPars.fVerbose {
        return; }
    fmt.Printf("CutPair = %.0f  ", p.CutCount[0]);
    fmt.Printf("Merge = %.0f (%.2f %%)  ", p.CutCount[1], 100.0*p.CutCount[1]/p.CutCount[0]);
    fmt.Printf("Eval = %.0f (%.2f %%)  ", p.CutCount[2], 100.0*p.CutCount[2]/p.CutCount[0]);
    fmt.Printf("Cut = %.0f (%.2f %%)  ", p.CutCount[3], 100.0*p.CutCount[3]/p.CutCount[0]);
    fmt.Printf("\n");
    fmt.Printf("Gia = %.2f MB  ", MemGia);
    fmt.Printf("Man = %.2f MB  ", MemMan); 
    fmt.Printf("Cut = %.2f MB   ", MemCuts);
    fmt.Printf("Map = %.2f MB  ", MemMap); 
    fmt.Printf("TT = %.2f MB  ", MemTt); 
    fmt.Printf("Total = %.2f MB", MemGia + MemMan + MemCuts + MemMap + MemTt); 
    fmt.Printf("\n");
    if true {
        var i int32;
        for  i = 0; i <= p.pPars.nLutSize; i++  {
            fmt.Printf("%d = %d  ", i, p.nCutCounts[i]);
        }
        if p.vTtMem {
            fmt.Printf("TT = %d (%.2f %%)  ", Vec_MemEntryNum(p.vTtMem), 100.0 * Vec_MemEntryNum(p.vTtMem) / p.CutCount[2]); }
        Abc_PrintTime( 1, "Time",    Abc_Clock() - p.clkStart );
    }
    fflush( os.Stdout );
}
func Mf_ManComputeCuts(p *Mf_Man_t) {
    var i int32;
    Gia_ManForEachAndId( p.pGia, i )
        Mf_ObjMergeOrder( p, i );
    Mf_ManSetMapRefs( p );
    Mf_ManPrintStats( p, (char *)(p.fUseEla ? "Ela  " : (p.Iter ? "Area " : "Delay")) );
}

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
func Mf_ManComputeMapping(p *Mf_Man_t) {
    var i int32;
    Gia_ManForEachAndId( p.pGia, i )
        Mf_ObjComputeBestCut( p, i );
    Mf_ManSetMapRefs( p );
    Mf_ManPrintStats( p, (char *)(p.fUseEla ? "Ela  " : (p.Iter ? "Area " : "Delay")) );
}
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

