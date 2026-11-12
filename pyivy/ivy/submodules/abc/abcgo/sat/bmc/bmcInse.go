package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcInse.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcInse.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bmc.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satStore.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// inlining of macro ABC_ALLOC
func Gia_ParTestAlloc(p *Gia_Man_t, nWords int32) { assert(!p.pData, "!p.pData"); p.pData = (unsigned *)((word *) malloc(sizeof(word) * (size_t)(2*nWords*Gia_ManObjNum(p)))); p.iData = nWords; }
func Gia_ParTestFree(p *Gia_Man_t) { p.pData = nil; p.iData = 0;             }
func Gia_ParTestObj(p *Gia_Man_t, Id int32) *word { return (word *)p.pData + Id*(p.iData << 1);   }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInseInit(p *Gia_Man_t, vInit *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var pData1, pData0 *word;
    var i, k int32;
    // inlining of macro Gia_ManForEachRi
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+k)); k++  {
        pData0 = Gia_ParTestObj( p, Gia_ObjId(p, pObj) );
        pData1 = pData0 + p.iData;
        if Vec_IntEntry(vInit, k) == 0 { // 0
            for  i = 0; i < p.iData; i++  {
                pData0[i] = ~(word)0, pData1[i] = 0;
            }
        } else if Vec_IntEntry(vInit, k) == 1 { // 1
            for  i = 0; i < p.iData; i++  {
                pData0[i] = 0, pData1[i] = ~(word)0;
            }
        } else { // if ( Vec_IntEntry(vInit, k) > 1 ) // X
            for  i = 0; i < p.iData; i++  {
                pData0[i] = pData1[i] = 0;
            } }
    }
}
func Gia_ManInseSimulateObj(p *Gia_Man_t, Id int32) {
    var pObj *Gia_Obj_t  = Gia_ManObj( p, Id );
    var pData0, pDataA0, pDataB0 *word;
    var pData1, pDataA1, pDataB1 *word;
    var i int32;
    if Gia_ObjIsAnd(pObj) {
        pData0  = Gia_ParTestObj( p, Id );
        pData1  = pData0 + p.iData;
        if Gia_ObjFaninC0(pObj) {
            pDataA1 = Gia_ParTestObj( p, Gia_ObjFaninId0(pObj, Id) );
            pDataA0 = pDataA1 + p.iData;
            if Gia_ObjFaninC1(pObj) {
                pDataB1 = Gia_ParTestObj( p, Gia_ObjFaninId1(pObj, Id) );
                pDataB0 = pDataB1 + p.iData;
            
            } else {
                pDataB0 = Gia_ParTestObj( p, Gia_ObjFaninId1(pObj, Id) );
                pDataB1 = pDataB0 + p.iData;
            }
        
        } else {
            pDataA0 = Gia_ParTestObj( p, Gia_ObjFaninId0(pObj, Id) );
            pDataA1 = pDataA0 + p.iData;
            if Gia_ObjFaninC1(pObj) {
                pDataB1 = Gia_ParTestObj( p, Gia_ObjFaninId1(pObj, Id) );
                pDataB0 = pDataB1 + p.iData;
            
            } else {
                pDataB0 = Gia_ParTestObj( p, Gia_ObjFaninId1(pObj, Id) );
                pDataB1 = pDataB0 + p.iData;
            }
        }
        for  i = 0; i < p.iData; i++  {
            pData0[i] = pDataA0[i] | pDataB0[i], pData1[i] = pDataA1[i] & pDataB1[i];
        }
    
    } else if Gia_ObjIsCo(pObj) {
        pData0  = Gia_ParTestObj( p, Id );
        pData1  = pData0 + p.iData;
        if Gia_ObjFaninC0(pObj) {
            pDataA1 = Gia_ParTestObj( p, Gia_ObjFaninId0(pObj, Id) );
            pDataA0 = pDataA1 + p.iData;
        
        } else {
            pDataA0 = Gia_ParTestObj( p, Gia_ObjFaninId0(pObj, Id) );
            pDataA1 = pDataA0 + p.iData;
        }
        for  i = 0; i < p.iData; i++  {
            pData0[i] = pDataA0[i], pData1[i] = pDataA1[i];
        }
    
    } else if Gia_ObjIsCi(pObj) {
        if Gia_ObjIsPi(p, pObj) {
            pData0  = Gia_ParTestObj( p, Id );
            pData1  = pData0 + p.iData;
            for  i = 0; i < p.iData; i++  {
                pData0[i] = Gia_ManRandomW(0), pData1[i] = ~pData0[i];
            }
        
        } else {
            var Id2 int32  = Gia_ObjId(p, Gia_ObjRoToRi(p, pObj));
            pData0  = Gia_ParTestObj( p, Id );
            pData1  = pData0 + p.iData;
            pDataA0 = Gia_ParTestObj( p, Id2 );
            pDataA1 = pDataA0 + p.iData;
            for  i = 0; i < p.iData; i++  {
                pData0[i] = pDataA0[i], pData1[i] = pDataA1[i];
            }
        }
    
    } else if Gia_ObjIsConst0(pObj) {
        pData0  = Gia_ParTestObj( p, Id );
        pData1  = pData0 + p.iData;
        for  i = 0; i < p.iData; i++  {
            pData0[i] = ~(word)0, pData1[i] = 0;
        }
    
    } else { assert(false, "0"); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInseHighestScore(p *Gia_Man_t, pCost *int32) int32 {
    var pObj *Gia_Obj_t;
    var pData0, pData1 *word;
    var pCounts *int32;
    var CountBest int32;
    var i, k, b, nPats, iPat int32;
    nPats = 64 * p.iData;
    // inlining of macro ABC_CALLOC
    pCounts = ((int *) calloc((size_t)(nPats), sizeof(int)));
    // inlining of macro Gia_ManForEachRi
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+k)); k++  {
        pData0 = Gia_ParTestObj( p, Gia_ObjId(p, pObj) );
        pData1 = pData0 + p.iData;
        for  i = 0; i < p.iData; i++  {
            for  b = 0; b < 64; b++  {
                pCounts[64*i+b] += (((pData0[i] >> b) & 1) || ((pData1[i] >> b) & 1));
            }
        } // binary
    }
    iPat = 0;
    CountBest = pCounts[0]; 
    for  k = 1; k < nPats; k++  {
        if CountBest < pCounts[k] {
            CountBest = pCounts[k];
iPat = k; }
    }
    *pCost = Gia_ManRegNum(p) - CountBest; // ternary
    pCounts = nil;
    return iPat;
}
func Gia_ManInseFindStarting(p *Gia_Man_t, iPat int32, vInit *Vec_Int_t, vInputs *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var pData0, pData1 *word;
    var i, k int32;
    Vec_IntClear( vInit );
    // inlining of macro Gia_ManForEachRi
    for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+k)); k++  {
        pData0 = Gia_ParTestObj( p, Gia_ObjId(p, pObj) );
        pData1 = pData0 + p.iData;
        for  i = 0; i < p.iData; i++  {
            assert((pData0[i] & pData1[i]) == 0, "(pData0[i] & pData1[i]) == 0");
        }
        if Abc_InfoHasBit( (unsigned *)pData0, iPat ) {
            Vec_IntPush( vInit, 0 );
        } else if Abc_InfoHasBit( (unsigned *)pData1, iPat ) {
            Vec_IntPush( vInit, 1 );
        } else { 
            Vec_IntPush( vInit, 2 ); }
    }
    // inlining of macro Gia_ManForEachPi
    for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
        pData0 = Gia_ParTestObj( p, Gia_ObjId(p, pObj) );
        pData1 = pData0 + p.iData;
        for  i = 0; i < p.iData; i++  {
            assert((pData0[i] & pData1[i]) == 0, "(pData0[i] & pData1[i]) == 0");
        }
        if Abc_InfoHasBit( (unsigned *)pData0, iPat ) {
            Vec_IntPush( vInputs, 0 );
        } else if Abc_InfoHasBit( (unsigned *)pData1, iPat ) {
            Vec_IntPush( vInputs, 1 );
        } else { 
            Vec_IntPush( vInputs, 2 ); }
    }
}
func Gia_ManInseSimulate(p *Gia_Man_t, vInit0 *Vec_Int_t, vInputs *Vec_Int_t, vInit *Vec_Int_t) *Vec_Int_t {
    var vRes *Vec_Int_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var nFrames int32  = Vec_IntSize(vInputs) / Gia_ManPiNum(p);
    var i, f, iBit int32;
    assert(Vec_IntSize(vInputs) % Gia_ManPiNum(p) == 0, "Vec_IntSize(vInputs) % Gia_ManPiNum(p) == 0");
    assert(Vec_IntSize(vInit0) == Gia_ManRegNum(p), "Vec_IntSize(vInit0) == Gia_ManRegNum(p)");
    assert(Vec_IntSize(vInit) == Gia_ManRegNum(p), "Vec_IntSize(vInit) == Gia_ManRegNum(p)");
    Gia_ManConst0(p).fMark0 = 0;
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        pObj.fMark0 = Vec_IntEntry(vInit0, i);
    }
    for  f = 0; f < nFrames; f++  {
        // inlining of macro Gia_ManForEachPi
        for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
            pObj.fMark0 = Vec_IntEntry(vInputs, iBit++);
        }
        // inlining of macro Gia_ManForEachAnd
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
            pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & 
                           (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)); }
        }
        // inlining of macro Gia_ManForEachRi
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
            pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj);
        }
        // inlining of macro Gia_ManForEachRiRo
        for  i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
            pObjRo.fMark0 = pObjRi.fMark0;
        }
    }
    assert(iBit == Vec_IntSize(vInputs), "iBit == Vec_IntSize(vInputs)");
    vRes = Vec_IntAlloc( Gia_ManRegNum(p) );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        assert(Vec_IntEntry(vInit, i) == 2 || Vec_IntEntry(vInit, i) == int32(pObj.fMark0), "Vec_IntEntry(vInit, i) == 2 || Vec_IntEntry(vInit, i) == int32(pObj.fMark0)");
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        Vec_IntPush( vRes, pObj.fMark0 | (Vec_IntEntry(vInit, i) != 2 ? 4 : 0) );
    }
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.fMark0 = 0;
    }
    return vRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInsePerform(p *Gia_Man_t, vInit0 *Vec_Int_t, nFrames int32, nWords int32, fVerbose int32) *Vec_Int_t {
    var vRes, vInit, vInputs *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i, f, iPat, Cost, Cost0 int32;
    var clk abctime;
    var clkTotal abctime = Abc_Clock();
    Gia_ManRandomW( 1 );
    if fVerbose != 0 {
        fmt.Printf("Running with %d frames, %d words, and %sgiven init state.\n", nFrames, nWords, vInit0 ? "":"no "); }
    vInit = Vec_IntAlloc(0); Vec_IntFill( vInit, Gia_ManRegNum(p), 2 );
    vInputs = Vec_IntStart( Gia_ManPiNum(p) * nFrames );
    Gia_ParTestAlloc( p, nWords );
    Gia_ManInseInit( p, vInit );
    Cost0 = 0;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vInit)) && (((iPat) = Vec_IntEntry(vInit, i)), 1); i++  {
        Cost0 += ((iPat >> 1) & 1);
    }
    if fVerbose != 0 {
        fmt.Printf("Frame =%6d : Values =%6d (out of %6d)\n", 0, Cost0, Cost0); }
    for  f = 0; f < nFrames; f++  {
        clk = Abc_Clock();
        // inlining of macro Gia_ManForEachObj
        for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
            Gia_ManInseSimulateObj( p, i );
        }
        iPat = Gia_ManInseHighestScore( p, &Cost );
        Gia_ManInseFindStarting( p, iPat, vInit, vInputs );
        Gia_ManInseInit( p, vInit );
        if fVerbose != 0 {
            fmt.Printf("Frame =%6d : Values =%6d (out of %6d)   ", f+1, Cost, Cost0); }
        if fVerbose != 0 {
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk ); }
    }
    Gia_ParTestFree( p );
    vRes = Gia_ManInseSimulate( p, vInit0, vInputs, vInit );
    Vec_IntFreeP( &vInit );
    Vec_IntFreeP( &vInputs );
    fmt.Printf("After %d frames, found a sequence to produce %d x-values (out of %d).  ", f, Cost, Gia_ManRegNum(p));
    Abc_PrintTime( 1, "Total runtime", Abc_Clock() - clkTotal );
    return vRes;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInseTest(p *Gia_Man_t, vInit0 *Vec_Int_t, nFrames int32, nWords int32, nTimeOut int32, fSim int32, fVerbose int32) *Vec_Int_t {
    var vRes, vInit *Vec_Int_t;
    vInit = Vec_IntAlloc(0); Vec_IntFill( vInit, Gia_ManRegNum(p), 0 );
    vRes = Gia_ManInsePerform( p, vInit, nFrames, nWords, fVerbose );
    if vInit != vInit0 {
        Vec_IntFree( vInit ); }
    return vRes;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

