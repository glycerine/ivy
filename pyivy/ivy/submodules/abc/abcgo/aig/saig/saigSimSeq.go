package saig

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [saigSimSeq.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Sequential AIG package.]

  Synopsis    [Fast sequential AIG simulator.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: saigSimSeq.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "saig.h"
//#include "proof/ssw/ssw.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// combinational simulation manager
   
type Raig_Man_t struct {
    // parameters
    pAig      *Aig_Man_t;         // the AIG to be used for simulation
    nWords             int32;       // the number of words to simulate
    // AIG representation
    nPis             int32;         // the number of primary inputs
    nPos             int32;         // the number of primary outputs
    nCis             int32;         // the number of combinational inputs
    nCos             int32;         // the number of combinational outputs
    nNodes             int32;       // the number of internal nodes
    nObjs             int32;        // nCis + nNodes + nCos + 2
    pFans0            *int32;       // fanin0 for all objects
    pFans1            *int32;       // fanin1 for all objects
    vCis2Ids      *Vec_Int_t;     // mapping of CIs into their PI ids
    vLos      *Vec_Int_t;         // register outputs
    vLis      *Vec_Int_t;         // register inputs
    // simulation info
    pRefs            *int32;        // reference counter for each node
    pSims       *uint32;        // simlulation information for each node
    // recycable memory
    pMems       *uint32;        // allocated simulaton memory
    nWordsAlloc             int32;  // the number of allocated entries
    nMems             int32;        // the number of used entries  
    nMemsMax             int32;     // the max number of used entries 
    MemFree             int32;      // next free entry
}

func Raig_Var2Lit(Var int32, fCompl int32) int32 { return Var + Var + fCompl; }
func Raig_Lit2Var(Lit int32) int32 { return Lit >> 1;           }
func Raig_LitIsCompl(Lit int32) int32 { return Lit & 1;            }
func Raig_LitNot(Lit int32) int32 { return Lit ^ 1;            }
func Raig_LitNotCond(Lit int32, c int32) int32 { return Lit ^ int32((c > 0)); }
func Raig_LitRegular(Lit int32) int32 { return Lit & ~01;          }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Find the PO corresponding to the PO driver.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManFindPo(pAig *Aig_Man_t, iNode int32) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // inlining of macro Saig_ManForEachPo
    // inlining of macro Vec_PtrForEachEntryStop
    for  i = 0; (i < Saig_ManPoNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        if pObj.iData == iNode {
            return i; }
    }
    return -1;
}

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManCreate_rec(p *Raig_Man_t, pObj *Aig_Obj_t) int32 {
    var iFan0, iFan1 int32;
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if pObj.iData {
        return pObj.iData; }
    assert(!Aig_ObjIsConst1(pObj), "!Aig_ObjIsConst1(pObj)");
    if Aig_ObjIsNode(pObj) {
        iFan0 = Raig_ManCreate_rec( p, Aig_ObjFanin0(pObj) );
        iFan0 = (iFan0 << 1) | Aig_ObjFaninC0(pObj);
        iFan1 = Raig_ManCreate_rec( p, Aig_ObjFanin1(pObj) );
        iFan1 = (iFan1 << 1) | Aig_ObjFaninC1(pObj);
    
    } else if Aig_ObjIsCo(pObj) {
        iFan0 = Raig_ManCreate_rec( p, Aig_ObjFanin0(pObj) );
        iFan0 = (iFan0 << 1) | Aig_ObjFaninC0(pObj);
        iFan1 = 0;
    
    } else {
        iFan0 = iFan1 = 0;
        Vec_IntPush( p.vCis2Ids, Aig_ObjCioId(pObj) );
    }
    p.pFans0[p.nObjs] = iFan0;
    p.pFans1[p.nObjs] = iFan1;
    p.pRefs[p.nObjs] = Aig_ObjRefs(pObj);
    return pObj.iData = p.nObjs++;
}

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManCreate(pAig *Aig_Man_t) *Raig_Man_t {
    var p *Raig_Man_t;
    var pObj *Aig_Obj_t;
    var i, nObjs int32;
    Aig_ManCleanData( pAig );
    // inlining of macro ABC_ALLOC
    p = (Raig_Man_t *)((Raig_Man_t *) malloc(sizeof(Raig_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Raig_Man_t) );
    p.pAig = pAig;
    p.nPis = Saig_ManPiNum(pAig);
    p.nPos = Saig_ManPoNum(pAig);
    p.nCis = Aig_ManCiNum(pAig);
    p.nCos = Aig_ManCoNum(pAig);
    p.nNodes = Aig_ManNodeNum(pAig);
    nObjs = p.nCis + p.nCos + p.nNodes + 2;
    // inlining of macro ABC_ALLOC
    p.pFans0 = ((int *) malloc(sizeof(int) * (size_t)(nObjs)));
    // inlining of macro ABC_ALLOC
    p.pFans1 = ((int *) malloc(sizeof(int) * (size_t)(nObjs)));
    // inlining of macro ABC_ALLOC
    p.pRefs = ((int *) malloc(sizeof(int) * (size_t)(nObjs)));
    // inlining of macro ABC_CALLOC
    p.pSims = ((unsigned *) calloc((size_t)(nObjs), sizeof(unsigned)));
    p.vCis2Ids = Vec_IntAlloc( Aig_ManCiNum(pAig) );
    // add objects (0=unused; 1=const1)
    p.nObjs = 2;
    pObj = Aig_ManConst1( pAig );
    pObj.iData = 1;
    // inlining of macro Aig_ManForEachCi
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCis)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i)), 1); i++  {
        if Aig_ObjRefs(pObj) == 0 {
            Raig_ManCreate_rec( p, pObj ); }
    }
    // inlining of macro Aig_ManForEachCo
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(pAig.vCos)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i)), 1); i++  {
        Raig_ManCreate_rec( p, pObj );
    }
    assert(Vec_IntSize(p.vCis2Ids) == Aig_ManCiNum(pAig), "Vec_IntSize(p.vCis2Ids) == Aig_ManCiNum(pAig)");
    assert(p.nObjs == nObjs, "p.nObjs == nObjs");
    // collect flop outputs
    p.vLos = Vec_IntAlloc( Aig_ManRegNum(pAig) );
    // inlining of macro Saig_ManForEachLo
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCis, i+Saig_ManPiNum(pAig))), 1); i++  {
        Vec_IntPush( p.vLos, pObj.iData );
    }
    // collect flop inputs
    p.vLis = Vec_IntAlloc( Aig_ManRegNum(pAig) );
    // inlining of macro Saig_ManForEachLi
    for  i = 0; (i < Saig_ManRegNum(pAig)) && (((pObj) = (Aig_Obj_t *)Vec_PtrEntry(pAig.vCos, i+Saig_ManPoNum(pAig))), 1); i++  {
        Vec_IntPush( p.vLis, pObj.iData );
        assert(p.pRefs[ pObj.iData ] == 0, "p.pRefs[ pObj.iData ] == 0");
        p.pRefs[ pObj.iData ]++;
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Creates fast simulation manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManDelete(p *Raig_Man_t) {
    Vec_IntFree( p.vCis2Ids );
    Vec_IntFree( p.vLos );
    Vec_IntFree( p.vLis );
    // inlining of macro ABC_FREE
    ((p.pFans0) ? (free((char *) (p.pFans0)), (p.pFans0) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pFans1) ? (free((char *) (p.pFans1)), (p.pFans1) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pRefs) ? (free((char *) (p.pRefs)), (p.pRefs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSims) ? (free((char *) (p.pSims)), (p.pSims) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMems) ? (free((char *) (p.pMems)), (p.pMems) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [References simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManSimRef(p *Raig_Man_t, i int32) *uint32 {
    var pSim *uint32;
    assert(i > 1, "i > 1");
    assert(p.pSims[i] == 0, "p.pSims[i] == 0");
    if p.MemFree == 0 {
        var pPlace *uint32;
        var Ent uint32;
        if p.nWordsAlloc == 0 {
            assert(p.pMems == nil, "p.pMems == nil");
            p.nWordsAlloc = (1<<17); // -> 1Mb
            p.nMems = 1;
        }
        p.nWordsAlloc *= 2;
        // inlining of macro ABC_REALLOC
        
        if p.pMems {
            p.pMems = ((unsigned *) realloc((char *)(p.pMems), sizeof(unsigned) * (size_t)(p.nWordsAlloc)));
        } else {
            p.pMems = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(p.nWordsAlloc)));
        }
        memset( p.pMems, 0xff, sizeof(unsigned) * (p.nWords + 1) );
        pPlace = (unsigned *)&p.MemFree;
        for  Ent = p.nMems * (p.nWords + 1); 
              Ent + p.nWords + 1 < uint32(p.nWordsAlloc); 
              Ent += p.nWords + 1  {
            *pPlace = Ent;
            pPlace = p.pMems + Ent;
        }
        *pPlace = 0;
    }
    p.pSims[i] = p.MemFree;
    pSim = p.pMems + p.MemFree;
    p.MemFree = pSim[0];
    pSim[0] = p.pRefs[i];
    p.nMems++;
    if p.nMemsMax < p.nMems {
        p.nMemsMax = p.nMems; }
    return pSim;
}

/**Function*************************************************************

  Synopsis    [Dereference simulaton info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManSimDeref(p *Raig_Man_t, i int32) *uint32 {
    var pSim *uint32;
    assert((i != 0), "i");
    if i == 1 { // const 1
        return p.pMems; }
    assert(p.pSims[i] > 0, "p.pSims[i] > 0");
    pSim = p.pMems + p.pSims[i];
    if --pSim[0] == 0 {
        pSim[0] = p.MemFree;
        p.MemFree = p.pSims[i];
        p.pSims[i] = 0;
        p.nMems--;
    }
    return pSim;
}

/**Function*************************************************************

  Synopsis    [Simulates one round.]

  Description [Returns the number of PO entry if failed; 0 otherwise.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManSimulateRound(p *Raig_Man_t, fMiter int32, fFirst int32, piPat *int32) int32 { 
    var pRes0, pRes1, pRes *uint32;
    var i, w, nCis, nCos, iFan0, iFan1, iPioNum int32;
    // nove the values to the register outputs
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCis2Ids)) && (((iPioNum) = Vec_IntEntry(p.vCis2Ids, i)), 1); i++  {
        if iPioNum < p.nPis {
            continue; }
        pRes = Raig_ManSimRef( p, Vec_IntEntry(p.vLos, iPioNum-p.nPis) );
        if fFirst != 0 {
            memset( pRes + 1, 0, sizeof(unsigned) * p.nWords );
        } else {
            pRes0 = Raig_ManSimDeref( p, Vec_IntEntry(p.vLis, iPioNum-p.nPis) );
            for  w = 1; w <= p.nWords; w++  {
                pRes[w] = pRes0[w];
            }
        }
        // handle unused PIs
        if pRes[0] == 0 {
            pRes[0] = 1;
            Raig_ManSimDeref( p, Vec_IntEntry(p.vLos, iPioNum-p.nPis) );
        }
    }
    // simulate the logic
    nCis = nCos = 0;
    for  i = 2; i < p.nObjs; i++  {
        if p.pFans0[i] == 0 { // ci always has zero first fanin
        
            iPioNum = Vec_IntEntry( p.vCis2Ids, nCis );
            if iPioNum < p.nPis {
                pRes = Raig_ManSimRef( p, i );
                for  w = 1; w <= p.nWords; w++  {
                    pRes[w] = Aig_ManRandom( 0 );
                }
                // handle unused PIs
                if pRes[0] == 0 {
                    pRes[0] = 1;
                    Raig_ManSimDeref( p, i );
                }
            
            } else {
                assert(Vec_IntEntry(p.vLos, iPioNum-p.nPis) == i, "Vec_IntEntry(p.vLos, iPioNum-p.nPis) == i"); }
            nCis++;
            continue;
        }
        if p.pFans1[i] == 0 { // co always has non-zero 1st fanin and zero 2nd fanin
        
            pRes0 = Raig_ManSimDeref( p, Raig_Lit2Var(p.pFans0[i]) );
            if nCos < p.nPos && (fMiter != 0) {
                var Const uint32;
                if Raig_LitIsCompl(p.pFans0[i]) != 0 {
                    Const = ~0;
                } else {
                    Const = 0;
                }
                for  w = 1; w <= p.nWords; w++  {
                    if pRes0[w] != Const {
                        *piPat = 32*(w-1) + Aig_WordFindFirstBit( pRes0[w] ^ Const );
                        return i;
                    }
                }
            
            } else {
                pRes = Raig_ManSimRef( p, i );
                assert(pRes[0] == 1, "pRes[0] == 1");
                if Raig_LitIsCompl(p.pFans0[i]) != 0 {
                    for  w = 1; w <= p.nWords; w++  {
                        pRes[w] = ~pRes0[w];
                    }
                } else {
                    for  w = 1; w <= p.nWords; w++  {
                        pRes[w] = pRes0[w];
                    } }
            }
            nCos++;
            continue;
        }
        pRes  = Raig_ManSimRef( p, i );
        assert(pRes[0] > 0, "pRes[0] > 0");
        iFan0 = p.pFans0[i];
        iFan1 = p.pFans1[i];
        pRes0 = Raig_ManSimDeref( p, Raig_Lit2Var(p.pFans0[i]) );
        pRes1 = Raig_ManSimDeref( p, Raig_Lit2Var(p.pFans1[i]) );
        if (Raig_LitIsCompl(iFan0) != 0) && (Raig_LitIsCompl(iFan1) != 0) {
            for  w = 1; w <= p.nWords; w++  {
                pRes[w] = ~(pRes0[w] | pRes1[w]);
            }
        } else if (Raig_LitIsCompl(iFan0) != 0) && !(Raig_LitIsCompl(iFan1) != 0) {
            for  w = 1; w <= p.nWords; w++  {
                pRes[w] = ~pRes0[w] & pRes1[w];
            }
        } else if !(Raig_LitIsCompl(iFan0) != 0) && (Raig_LitIsCompl(iFan1) != 0) {
            for  w = 1; w <= p.nWords; w++  {
                pRes[w] = pRes0[w] & ~pRes1[w];
            }
        } else if !(Raig_LitIsCompl(iFan0) != 0) && !(Raig_LitIsCompl(iFan1) != 0) {
            for  w = 1; w <= p.nWords; w++  {
                pRes[w] = pRes0[w] & pRes1[w];
            } }
    }
    assert(nCis == p.nCis, "nCis == p.nCis");
    assert(nCos == p.nCos, "nCos == p.nCos");
    assert(p.nMems == 1 + Vec_IntSize(p.vLis), "p.nMems == 1 + Vec_IntSize(p.vLis)");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns the counter-example.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManGenerateCounter(pAig *Aig_Man_t, iFrame int32, iOut int32, nWords int32, iPat int32, vCis2Ids *Vec_Int_t) *Abc_Cex_t {
    var p *Abc_Cex_t;
    var pData *uint32;
    var f, i, w, iPioId, Counter int32;
    p = Abc_CexAlloc( Aig_ManRegNum(pAig), Saig_ManPiNum(pAig), iFrame+1 );
    p.iFrame = iFrame;
    p.iPo = iOut;
    // fill in the binary data
    Aig_ManRandom( 1 );
    Counter = p.nRegs;
    // inlining of macro ABC_ALLOC
    pData = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nWords)));
    for  f = 0; f <= iFrame; f++, Counter += p.nPis  {
    for  i = 0; i < Aig_ManCiNum(pAig); i++  {
        iPioId = Vec_IntEntry( vCis2Ids, i );
        if iPioId >= p.nPis {
            continue; }
        for  w = 0; w < nWords; w++  {
            pData[w] = Aig_ManRandom( 0 );
        }
        if Abc_InfoHasBit( pData, iPat ) {
            Abc_InfoSetBit( p.pData, Counter + iPioId ); }
    }
    }
    // inlining of macro ABC_FREE
    ((pData) ? (free((char *) (pData)), (pData) = 0) : 0);
    return p;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if the bug is detected, 0 otherwise.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Raig_ManSimulate(pAig *Aig_Man_t, nWords int32, nIters int32, TimeLimit int32, fMiter int32, fVerbose int32) int32 {
    var p *Raig_Man_t;
    var Status Sec_MtrStatus_t;
    var i, iPat, RetValue int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    assert(Aig_ManRegNum(pAig) > 0, "Aig_ManRegNum(pAig) > 0");
    Status = Sec_MiterStatus( pAig );
    if Status.nSat > 0 {
        fmt.Printf("Miter is trivially satisfiable (output %d).\n", Status.iOut);
        return 1;
    }
    if Status.nUndec == 0 {
        fmt.Printf("Miter is trivially unsatisfiable.\n");
        return 0;
    }
    Aig_ManRandom( 1 );
    p = Raig_ManCreate( pAig );
    p.nWords = nWords;
    // iterate through objects
    for  i = 0; i < nIters; i++  {
        clk = Abc_Clock();
        RetValue = Raig_ManSimulateRound( p, fMiter, i==0, &iPat );
        if fVerbose != 0 {
            fmt.Printf("Frame %4d out of %4d and timeout %3d sec. ", i+1, nIters, TimeLimit);
            fmt.Printf("Time = %7.2f sec\r", (1.0*Abc_Clock()-clkTotal)/int64(time.Second));
        }
        if RetValue > 0 {
            var iOut int32  = Raig_ManFindPo(p.pAig, RetValue);
            assert(pAig.pSeqModel == nil, "pAig.pSeqModel == nil");
            pAig.pSeqModel = Raig_ManGenerateCounter( pAig, i, iOut, nWords, iPat, p.vCis2Ids );
            if fVerbose != 0 {
            fmt.Printf("Miter is satisfiable after simulation (output %d).\n", iOut); }
            break;
        }
        if (Abc_Clock() - clk)/int64(time.Second) >= TimeLimit {
            fmt.Printf("No bug detected after %d frames with time limit %d seconds.\n", i+1, TimeLimit);
            break;
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Maxcut = %8d.  AigMem = %7.2f MB.  SimMem = %7.2f MB.  ", p.nMemsMax, 1.0*(p.nObjs * 16)/(1<<20), 1.0*(p.nMemsMax * 4 * (nWords+1))/(1<<20));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clkTotal)))/(float64(int64(time.Second))));;
    }
    Raig_ManDelete( p );
    return RetValue > 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

