package bmc

import "fmt"

/**CFile****************************************************************

  FileName    [bmcCexMin2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [SAT-based bounded model checking.]

  Synopsis    [CEX minimization.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: bmcCexMin2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "aig/gia/gia.h"
//#include "bmc.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Abc_InfoGet2Bits(vData *Vec_Int_t, nWords int32, iFrame int32, iObj int32) int32 {
    var pInfo *uint32  = (unsigned *)Vec_IntEntryP( vData, nWords * iFrame );
    return 3 & (pInfo[iObj >> 4] >> ((iObj & 15) << 1));
}
func Abc_InfoSet2Bits(vData *Vec_Int_t, nWords int32, iFrame int32, iObj int32, Value int32) {
    var pInfo *uint32  = (unsigned *)Vec_IntEntryP( vData, nWords * iFrame );
    Value ^= (3 & (pInfo[iObj >> 4] >> ((iObj & 15) << 1)));
    pInfo[iObj >> 4] ^= (Value << ((iObj & 15) << 1));
}
func Abc_InfoAdd2Bits(vData *Vec_Int_t, nWords int32, iFrame int32, iObj int32, Value int32) {
    var pInfo *uint32  = (unsigned *)Vec_IntEntryP( vData, nWords * iFrame );
    pInfo[iObj >> 4] |= (Value << ((iObj & 15) << 1));
}

func Gia_ManGetTwo(p *Gia_Man_t, iFrame int32, pObj *Gia_Obj_t) int32 { return Abc_InfoGet2Bits( p.vTruths, p.nTtWords, iFrame, Gia_ObjId(p, pObj) ); }
func Gia_ManSetTwo(p *Gia_Man_t, iFrame int32, pObj *Gia_Obj_t, Value int32) { Abc_InfoSet2Bits( p.vTruths, p.nTtWords, iFrame, Gia_ObjId(p, pObj), Value ); }
func Gia_ManAddTwo(p *Gia_Man_t, iFrame int32, pObj *Gia_Obj_t, Value int32) { Abc_InfoAdd2Bits( p.vTruths, p.nTtWords, iFrame, Gia_ObjId(p, pObj), Value ); }

/*
    For CEX minimization, all terminals (const0, PI, RO in frame0) have equal priority.
    For abstraction refinement, all terminals, except PPis, have higher priority.
*/

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Annotates the unrolling with CEX value/priority.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAnnotateUnrolling(p *Gia_Man_t, pCex *Abc_Cex_t, fJustMax int32) int32 {
    var pObj, pObjRi, pObjRo *Gia_Obj_t;
    var RetValue, f, k, Value, Value0, Value1, iBit int32;
    // start storage for internal info
    assert(p.vTruths == nil, "p.vTruths == nil");
    p.nTtWords = Abc_BitWordNum( 2 * Gia_ManObjNum(p) );
    p.vTruths  = Vec_IntStart( (pCex.iFrame + 1) * p.nTtWords );
    // assign values to all objects (const0 and RO in frame0 are assigned 0)
    Gia_ManCleanMark0(p);
    for  f, iBit = 0, pCex.nRegs; f <= pCex.iFrame; f++  {
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            if pObj.fMark0 = Abc_InfoHasBit(pCex.pData, iBit++) {
                Gia_ManAddTwo( p, f, pObj, 1 ); }
        }
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else
            if pObj.fMark0 = (Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj)) & (Gia_ObjFanin1(pObj).fMark0 ^ Gia_ObjFaninC1(pObj)) {
                Gia_ManAddTwo( p, f, pObj, 1 ); }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            if pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 ^ Gia_ObjFaninC0(pObj) {
                Gia_ManAddTwo( p, f, pObj, 1 ); }
        }
        if f == pCex.iFrame {
            break; }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            if pObjRo.fMark0 = pObjRi.fMark0 {
                Gia_ManAddTwo( p, f+1, pObjRo, 1 ); }
        }
    }
    assert(iBit == pCex.nBits, "iBit == pCex.nBits");
    // check the output value
    RetValue = Gia_ManPo(p, pCex.iPo).fMark0;
    if RetValue != 1 {
        fmt.Printf("Counter-example is invalid.\n"); }
    // assign justification to nodes
    Gia_ManCleanMark0(p);
    pObj = Gia_ManPo(p, pCex.iPo);
    pObj.fMark0 = 1;
    Gia_ManAddTwo( p, pCex.iFrame, pObj, 2 );
    for  f = pCex.iFrame; f >= 0; f--  {
        // transfer to CO drivers
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            if Gia_ObjFanin0(pObj).fMark0 = pObj.fMark0 {
                pObj.fMark0 = 0;
                Gia_ManAddTwo( p, f, Gia_ObjFanin0(pObj), 2 );
            }
        }
        // compute justification
        // inlining of macro Gia_ManForEachAndReverse
        for  k = p.nObjs - 1; (k > 0) && ((pObj) = Gia_ManObj(p, k)); k--  {  if !Gia_ObjIsAnd(pObj) {} else {
            if !pObj.fMark0 { 
                continue; }
            pObj.fMark0 = 0;
            Value = (1 & Gia_ManGetTwo(p, f, pObj));
            Value0 = (1 & Gia_ManGetTwo(p, f, Gia_ObjFanin0(pObj))) ^ Gia_ObjFaninC0(pObj);
            Value1 = (1 & Gia_ManGetTwo(p, f, Gia_ObjFanin1(pObj))) ^ Gia_ObjFaninC1(pObj);
            if Value0 == Value1 {
                assert(Value == (Value0 & Value1), "Value == (Value0 & Value1)");
                if (fJustMax != 0) || Value == 1 {
                    Gia_ObjFanin0(pObj).fMark0 = Gia_ObjFanin1(pObj).fMark0 = 1;
                    Gia_ManAddTwo( p, f, Gia_ObjFanin0(pObj), 2 );
                    Gia_ManAddTwo( p, f, Gia_ObjFanin1(pObj), 2 );
                
                } else {
                    Gia_ObjFanin0(pObj).fMark0 = 1;
                    Gia_ManAddTwo( p, f, Gia_ObjFanin0(pObj), 2 );
                }
            
            } else if Value0 == 0 {
                assert(Value == 0, "Value == 0");
                Gia_ObjFanin0(pObj).fMark0 = 1;
                Gia_ManAddTwo( p, f, Gia_ObjFanin0(pObj), 2 );
            
            } else if Value1 == 0 {
                assert(Value == 0, "Value == 0");
                Gia_ObjFanin1(pObj).fMark0 = 1;
                Gia_ManAddTwo( p, f, Gia_ObjFanin1(pObj), 2 );
            
            } else { assert(false, "0"); }
        }
        }
        if f == 0 {
            break; }
        // transfer to RIs
        // inlining of macro Gia_ManForEachPi
        for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
            pObj.fMark0 = 0;
        }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            if pObjRi.fMark0 = pObjRo.fMark0 {
                pObjRo.fMark0 = 0;
                Gia_ManAddTwo( p, f-1, pObjRi, 2 );
            }
        }
    }
    Gia_ManCleanMark0(p);
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Computing AIG characterizing all justifying assignments.]

  Description [Used in CEX minimization.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCreateUnate(p *Gia_Man_t, pCex *Abc_Cex_t, iFrame int32, nRealPis int32, fUseAllObjects int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjRo, pObjRi *Gia_Obj_t;
    var f, k, Value int32;
    assert(iFrame >= 0 && iFrame <= pCex.iFrame, "iFrame >= 0 && iFrame <= pCex.iFrame");
    pNew = Gia_ManStart( 1000 );
    pNew.pName = Abc_UtilStrsav( "unate" );
    Gia_ManCleanValue( p );
    // set flop outputs
    if nRealPis < 0 { // CEX min
    
        // inlining of macro Gia_ManForEachRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            if fUseAllObjects != 0 {
                var Value int32  = Gia_ManAppendCi(pNew);
                if Gia_ManGetTwo(p, iFrame, pObj) >> 1 { // in the path
                    pObj.Value = Value; }
            
            } else if Gia_ManGetTwo(p, iFrame, pObj) >> 1 { // in the path
                pObj.Value = Gia_ManAppendCi(pNew); }
        }
    
    } else {
        // inlining of macro Gia_ManForEachRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObj.Value = (Gia_ManGetTwo(p, iFrame, pObj) >> 1);
        }
    }
    Gia_ManHashAlloc( pNew );
    for  f = iFrame; f <= pCex.iFrame; f++  {
/*
        printf( "  F%03d ", f );
        Gia_ManForEachRo( p, pObj, k )
            printf( "%d", pObj->Value > 0 );
        printf( "\n" );
*/
        // set const0 to const1 if present
        pObj = Gia_ManConst0(p);
        pObj.Value = (Gia_ManGetTwo(p, f, pObj) >> 1);
        // set primary inputs 
        if nRealPis < 0 { // CEX min
        
            // inlining of macro Gia_ManForEachPi
            for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
                pObj.Value = (Gia_ManGetTwo(p, f, pObj) >> 1);
            }
        
        } else {
            // inlining of macro Gia_ManForEachPi
            for  k = 0; (k < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, k)); k++  {
                pObj.Value = (Gia_ManGetTwo(p, f, pObj) >> 1);
                if k >= nRealPis {
                    if fUseAllObjects != 0 {
                        var Value int32  = Gia_ManAppendCi(pNew);
                        if Gia_ManGetTwo(p, f, pObj) >> 1 { // in the path
                            pObj.Value = Value; }
                    
                    } else if Gia_ManGetTwo(p, f, pObj) >> 1 { // in the path
                        pObj.Value = Gia_ManAppendCi(pNew); }
                }
            }
        }
        // traverse internal nodes
        // inlining of macro Gia_ManForEachAnd
        for  k = 0; (k < p.nObjs) && ((pObj) = Gia_ManObj(p, k)); k++  {      if !Gia_ObjIsAnd(pObj) {} else { 
            pObj.Value = 0;
            Value = Gia_ManGetTwo(p, f, pObj);
            if !(Value >> 1) { // not in the path
                continue; }
            if Gia_ObjFanin0(pObj).Value && Gia_ObjFanin1(pObj).Value {
                if 1 & Gia_ManGetTwo(p, f, pObj) { // value 1
                
                    if Gia_ObjFanin0(pObj).Value > 1 && Gia_ObjFanin1(pObj).Value > 1 {
                        pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
                    } else if Gia_ObjFanin0(pObj).Value > 1 {
                        pObj.Value = Gia_ObjFanin0(pObj).Value;
                    } else if Gia_ObjFanin1(pObj).Value > 1 {
                        pObj.Value = Gia_ObjFanin1(pObj).Value;
                    } else { 
                        pObj.Value = 1; }
                
                } else { // value 0
                
                    if Gia_ObjFanin0(pObj).Value > 1 && Gia_ObjFanin1(pObj).Value > 1 {
                        pObj.Value = Gia_ManHashOr( pNew, Gia_ObjFanin0(pObj).Value, Gia_ObjFanin1(pObj).Value );
                    } else if Gia_ObjFanin0(pObj).Value > 1 {
                        pObj.Value = Gia_ObjFanin0(pObj).Value;
                    } else if Gia_ObjFanin1(pObj).Value > 1 {
                        pObj.Value = Gia_ObjFanin1(pObj).Value;
                    } else { 
                        pObj.Value = 1; }
                }
            
            } else if Gia_ObjFanin0(pObj).Value {
                pObj.Value = Gia_ObjFanin0(pObj).Value;
            } else if Gia_ObjFanin1(pObj).Value {
                pObj.Value = Gia_ObjFanin1(pObj).Value;
            } else { assert(false, "0"); }
        }
        }
        // inlining of macro Gia_ManForEachCo
        for  k = 0; (k < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, k)); k++  {
            pObj.Value = Gia_ObjFanin0(pObj).Value;
        }
        if f == pCex.iFrame {
            break; }
        // inlining of macro Gia_ManForEachRiRo
        for  k = 0; (k < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+k)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+k)); k++  {
            pObjRo.Value = pObjRi.Value;
        }
    }
    Gia_ManHashStop( pNew );
    // create primary output
    pObj = Gia_ManPo(p, pCex.iPo);
    assert((Gia_ManGetTwo(p, pCex.iFrame, pObj) >> 1), "(Gia_ManGetTwo(p, pCex.iFrame, pObj) >> 1)");
    assert(pObj.Value, "pObj.Value");
    Gia_ManAppendCo( pNew, pObj.Value );
    // cleanup
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManStop( pTemp );
    return pNew;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCexMin(p *Gia_Man_t, pCex *Abc_Cex_t, iFrameStart int32, nRealPis int32, fJustMax int32, fUseAll int32, fVerbose int32) *Abc_Cex_t {
    var pNew *Gia_Man_t;
    var f, RetValue int32;
    // check CEX
    assert(pCex.nPis == Gia_ManPiNum(p), "pCex.nPis == Gia_ManPiNum(p)");
//    assert( pCex->nRegs == Gia_ManRegNum(p) );
    assert(pCex.iPo < Gia_ManPoNum(p), "pCex.iPo < Gia_ManPoNum(p)");
    // check frames
    assert(iFrameStart >= 0 && iFrameStart <= pCex.iFrame, "iFrameStart >= 0 && iFrameStart <= pCex.iFrame");
    // check primary inputs
    assert(nRealPis < Gia_ManPiNum(p), "nRealPis < Gia_ManPiNum(p)");
    // prepare
    RetValue = Gia_ManAnnotateUnrolling( p, pCex, fJustMax );
    if nRealPis >= 0 { // refinement
    
        pNew = Gia_ManCreateUnate( p, pCex, iFrameStart, nRealPis, fUseAll );
        fmt.Printf("%3d : ", iFrameStart);
        Gia_ManPrintStats( pNew, nil );
        if fVerbose != 0 {
            Gia_AigerWrite( pNew, "temp.aig", 0, 0, 0 ); }
        Gia_ManStop( pNew );
    
    } else { // CEX min
    
        for  f = pCex.iFrame; f >= iFrameStart; f--  {
            pNew = Gia_ManCreateUnate( p, pCex, f, -1, fUseAll );
            fmt.Printf("%3d : ", f);
            Gia_ManPrintStats( pNew, nil );
            if fVerbose != 0 {
                Gia_AigerWrite( pNew, "temp.aig", 0, 0, 0 ); }
            Gia_ManStop( pNew );
        }
    }
    Vec_IntFreeP( &p.vTruths );
    p.nTtWords = 0;
    return nil;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

