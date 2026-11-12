package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaUnate.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Computation of structural unateness.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaUnate.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Flips the first bit in all entries of the vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Vec_IntFlopBit(p *Vec_Int_t) *Vec_Int_t {
    var i int32;
    for  i = 0; i < p.nSize; i++  {
        if i+1 < p.nSize && Abc_Lit2Var(p.pArray[i]) == Abc_Lit2Var(p.pArray[i+1]) { 
            i++; // skip variable appearing as both pos and neg literal
        } else {
            p.pArray[i] ^= 1; }
    }
    return p;
}

/**Function*************************************************************

  Synopsis    [Compute unateness for all outputs in terms of inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckUnateVec(p *Gia_Man_t, vCiIds *Vec_Int_t, vCoIds *Vec_Int_t) *Vec_Wec_t {
    var vCiVec *Vec_Int_t;
    if vCiIds != nil {
        vCiVec = Vec_IntDup(vCiIds);
    } else {
        vCiVec = Vec_IntStartNatural(Gia_ManCiNum(p));
    }
    var vCoVec *Vec_Int_t;
    if vCoIds != nil {
        vCoVec = Vec_IntDup(vCoIds);
    } else {
        vCoVec = Vec_IntStartNatural(Gia_ManCoNum(p));
    }
    var vUnatesCo *Vec_Wec_t  = Vec_WecStart( Vec_IntSize(vCoVec) );
    var vUnates *Vec_Wec_t    = Vec_WecStart( Gia_ManObjNum(p) );
    var vUnate0, vUnate1 *Vec_Int_t;
    var pObj *Gia_Obj_t; var i, CioId int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCiVec)) && (((CioId) = Vec_IntEntry(vCiVec, i)), 1); i++  {
        pObj = Gia_ManCi( p, CioId );
        Vec_IntPush( Vec_WecEntry(vUnates, Gia_ObjId(p, pObj)), Abc_Var2Lit(CioId, 0) );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        vUnate0 = Vec_WecEntry(vUnates, Gia_ObjFaninId0(pObj, i));
        vUnate1 = Vec_WecEntry(vUnates, Gia_ObjFaninId1(pObj, i));
        if Gia_ObjFaninC0(pObj) {
            vUnate0 = Vec_IntFlopBit(vUnate0);
        } else {
            vUnate0 = vUnate0;
        }
        if Gia_ObjFaninC1(pObj) {
            vUnate1 = Vec_IntFlopBit(vUnate1);
        } else {
            vUnate1 = vUnate1;
        }
        Vec_IntTwoMerge2( vUnate0, vUnate1, Vec_WecEntry(vUnates, i) ); 
        if Gia_ObjFaninC0(pObj) {
            vUnate0 = Vec_IntFlopBit(vUnate0);
        } else {
            vUnate0 = vUnate0;
        }
        if Gia_ObjFaninC1(pObj) {
            vUnate1 = Vec_IntFlopBit(vUnate1);
        } else {
            vUnate1 = vUnate1;
        }
    }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCoVec)) && (((CioId) = Vec_IntEntry(vCoVec, i)), 1); i++  {
        pObj    = Gia_ManCo( p, CioId );
        vUnate0 = Vec_WecEntry(vUnates, Gia_ObjFaninId0p(p, pObj));
        if Gia_ObjFaninC0(pObj) {
            vUnate0 = Vec_IntFlopBit(vUnate0);
        } else {
            vUnate0 = vUnate0;
        }
        Vec_IntAppend( Vec_WecEntry(vUnatesCo, i), vUnate0 );
        if Gia_ObjFaninC0(pObj) {
            vUnate0 = Vec_IntFlopBit(vUnate0);
        } else {
            vUnate0 = vUnate0;
        }
    }
    Vec_WecFree( vUnates );
    Vec_IntFree( vCiVec );
    Vec_IntFree( vCoVec );
    return vUnatesCo;
}

/**Function*************************************************************

  Synopsis    [Checks unateness one function in one variable.]

  Description [Returns 0 if Co is not unate in Ci.
               Returns 1 if Co is neg-unate in Ci.
               Returns 2 if Co is pos-unate in Ci.
               Returns 3 if Co does not depend on Ci.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckUnate_rec(p *Gia_Man_t, iObj int32) int32 {
    var pObj *Gia_Obj_t;
    var Res0, Res1 int32;
    if p.nTravIds - p.pTravIds[iObj] <= 3 {
        return p.nTravIds - p.pTravIds[iObj]; }
    pObj = Gia_ManObj( p, iObj );
    p.pTravIds[iObj] = p.nTravIds - 3;
    if Gia_ObjIsCi(pObj) {
        return 3; }
    Res0 = Gia_ManCheckUnate_rec( p, Gia_ObjFaninId0(pObj, iObj) );
    Res1 = Gia_ManCheckUnate_rec( p, Gia_ObjFaninId1(pObj, iObj) );
    if (Res0 == 1 || Res0 == 2) && Gia_ObjFaninC0(pObj) {
        Res0 = Res0 ^ 3;
    } else {
        Res0 = Res0;
    }
    if (Res1 == 1 || Res1 == 2) && Gia_ObjFaninC1(pObj) {
        Res1 = Res1 ^ 3;
    } else {
        Res1 = Res1;
    }
    p.pTravIds[iObj] = p.nTravIds - (Res0 & Res1);
    assert((Res0 & Res1) <= 3, "(Res0 & Res1) <= 3");
    return p.nTravIds - p.pTravIds[iObj];
}
func Gia_ManCheckUnate(p *Gia_Man_t, iCiId int32, iCoId int32) int32 {
    var Res int32;
    var CiObjId int32  = Gia_ObjId(p, Gia_ManCi(p, iCiId));
    var CoObjId int32  = Gia_ObjId(p, Gia_ManCo(p, iCoId));
    var pCoObj *Gia_Obj_t  = Gia_ManCo(p, iCoId);
    Gia_ManIncrementTravId( p ); // Co does not depend on Ci.
    Gia_ManIncrementTravId( p ); // Co is pos-unate in Ci
    Gia_ObjSetTravIdCurrentId( p, CiObjId );
    Gia_ManIncrementTravId( p ); // Co is neg-unate in Ci
    Gia_ManIncrementTravId( p ); // Co is not unate in Ci
    Res = Gia_ManCheckUnate_rec( p, Gia_ObjFaninId0(pCoObj, CoObjId) );
    if (Res == 1 || Res == 2) && Gia_ObjFaninC0(pCoObj) {
        return Res ^ 3;
    } else {
        return Res;
    }
}

/**Function*************************************************************

  Synopsis    [Testing procedure for all pairs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckUnateVecTest(p *Gia_Man_t, fVerbose int32) {
    var clk abctime  = Abc_Clock();
    var vUnates *Vec_Wec_t  = Gia_ManCheckUnateVec( p, nil, nil );
    var i int32;
    var o int32;
    var Var int32;
    var nVars int32 = Gia_ManCiNum(p);
    var nUnate, nNonUnate int32;
    // inlining of macro ABC_CALLOC
    var pBuffer string  = ((char *) calloc((size_t)(nVars+1), sizeof(char)));
    if fVerbose != 0 {
        fmt.Printf("Inputs  : ");
        for  i = 0; i < nVars; i++  {
            fmt.Printf("%d", i % 10);
        }
        fmt.Printf("\n");
    }
    for  o = 0; o < Gia_ManCoNum(p); o++  {
        var vUnate *Vec_Int_t  = Vec_WecEntry( vUnates, o );
        memset( pBuffer, ' ', (size_t)nVars );
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vUnate)) && (((Var) = Vec_IntEntry(vUnate, i)), 1); i++  {
            if i+1 < Vec_IntSize(vUnate) && Abc_Lit2Var(Var) == Abc_Lit2Var(Vec_IntEntry(vUnate, i+1)) { // both lits are present
                pBuffer[Abc_Lit2Var(Var)] = '.', i++, nNonUnate++; // does not depend on this var
            } else {
                pBuffer[Abc_Lit2Var(Var)] = Abc_LitIsCompl(Var) ? 'n' : 'p', nUnate++; }
        }
        if fVerbose != 0 {
            fmt.Printf("Out%4d : %s\n", o, pBuffer); }
    }
    // inlining of macro ABC_FREE
    pBuffer = "";
    // print stats
    fmt.Printf("Ins/Outs = %4d/%4d.  Total supp = %5d.  Total unate = %5d.\n", Gia_ManCiNum(p), Gia_ManCoNum(p), nUnate+nNonUnate, nUnate);
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Total time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    //Vec_WecPrint( vUnates, 0 );
    Vec_WecFree( vUnates );
}

/**Function*************************************************************

  Synopsis    [Testing procedure for one pair.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckUnateTest(p *Gia_Man_t, fComputeAll int32, fVerbose int32) {
    if fComputeAll != 0 {
        Gia_ManCheckUnateVecTest( p, fVerbose );
    } else {
        var clk abctime  = Abc_Clock();
        var i int32;
        var o int32;
        var nVars int32 = Gia_ManCiNum(p);
        var nUnate, nNonUnate int32;
        // inlining of macro ABC_CALLOC
        var pBuffer string  = ((char *) calloc((size_t)(nVars+1), sizeof(char)));
        if fVerbose != 0 {
            fmt.Printf("Inputs  : ");
            for  i = 0; i < nVars; i++  {
                fmt.Printf("%d", i % 10);
            }
            fmt.Printf("\n");
        }
        for  o = 0; o < Gia_ManCoNum(p); o++  {
            for  i = 0; i < nVars; i++  {
                var Res int32  = Gia_ManCheckUnate( p, i, o );
                if Res == 3 {       pBuffer[i] = ' ';
                } else if Res == 2 {  pBuffer[i] = 'p', nUnate++;
                } else if Res == 1 {  pBuffer[i] = 'n', nUnate++;
                } else if Res == 0 {  pBuffer[i] = '.', nNonUnate++;
                } else { assert(false, "0"); }
            }
            if fVerbose != 0 {
                fmt.Printf("Out%4d : %s\n", o, pBuffer); }
        }
        // inlining of macro ABC_FREE
        pBuffer = "";
        // print stats
        fmt.Printf("Ins/Outs = %4d/%4d.  Total supp = %5d.  Total unate = %5d.\n", Gia_ManCiNum(p), Gia_ManCoNum(p), nUnate+nNonUnate, nUnate);
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Total time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

