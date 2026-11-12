package pdr

import "os"

/**CFile****************************************************************

  FileName    [pdrTsim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Ternary simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrTsim.c,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pdrInt.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const PDR_ZER = 1
const PDR_ONE = 2
const PDR_UND = 3

func Pdr_ManSimInfoNot(Value int32) int32 {
    if Value == PDR_ZER {
        return PDR_ONE; }
    if Value == PDR_ONE {
        return PDR_ZER; }
    return PDR_UND;
}

func Pdr_ManSimInfoAnd(Value0 int32, Value1 int32) int32 {
    if Value0 == PDR_ZER || Value1 == PDR_ZER {
        return PDR_ZER; }
    if Value0 == PDR_ONE && Value1 == PDR_ONE {
        return PDR_ONE; }
    return PDR_UND;
}

func Pdr_ManSimInfoGet(p *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    return 3 & (p.pTerSimData[Aig_ObjId(pObj) >> 4] >> ((Aig_ObjId(pObj) & 15) << 1));
}

func Pdr_ManSimInfoSet(p *Aig_Man_t, pObj *Aig_Obj_t, Value int32) {
    assert(Value >= PDR_ZER && Value <= PDR_UND, "Value >= PDR_ZER && Value <= PDR_UND");
    Value ^= Pdr_ManSimInfoGet( p, pObj );
    p.pTerSimData[Aig_ObjId(pObj) >> 4] ^= (Value << ((Aig_ObjId(pObj) & 15) << 1));
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Marks the TFI cone and collects CIs and nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManCollectCone_rec(pAig *Aig_Man_t, pObj *Aig_Obj_t, vCiObjs *Vec_Int_t, vNodes *Vec_Int_t) {
    assert(!Aig_IsComplement(pObj), "!Aig_IsComplement(pObj)");
    if Aig_ObjIsTravIdCurrent(pAig, pObj) {
        return; }
    Aig_ObjSetTravIdCurrent(pAig, pObj);
    if Aig_ObjIsCi(pObj) {
        Vec_IntPush( vCiObjs, Aig_ObjId(pObj) );
        return;
    }
    Pdr_ManCollectCone_rec( pAig, Aig_ObjFanin0(pObj), vCiObjs, vNodes );
    if Aig_ObjIsCo(pObj) {
        return; }
    Pdr_ManCollectCone_rec( pAig, Aig_ObjFanin1(pObj), vCiObjs, vNodes );
    Vec_IntPush( vNodes, Aig_ObjId(pObj) );
}

/**Function*************************************************************

  Synopsis    [Marks the TFI cone and collects CIs and nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManCollectCone(pAig *Aig_Man_t, vCoObjs *Vec_Int_t, vCiObjs *Vec_Int_t, vNodes *Vec_Int_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    Vec_IntClear( vCiObjs );
    Vec_IntClear( vNodes );
    Aig_ManIncrementTravId( pAig );
    Aig_ObjSetTravIdCurrent( pAig, Aig_ManConst1(pAig) );
    Aig_ManForEachObjVec( vCoObjs, pAig, pObj, i )
        Pdr_ManCollectCone_rec( pAig, pObj, vCiObjs, vNodes );
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation for one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManExtendOneEval(pAig *Aig_Man_t, pObj *Aig_Obj_t) int32 {
    var Value0, Value1, Value int32;
    Value0 = Pdr_ManSimInfoGet( pAig, Aig_ObjFanin0(pObj) );
    if Aig_ObjFaninC0(pObj) {
        Value0 = Pdr_ManSimInfoNot( Value0 ); }
    if Aig_ObjIsCo(pObj) {
        Pdr_ManSimInfoSet( pAig, pObj, Value0 );
        return Value0;
    }
    assert(Aig_ObjIsNode(pObj), "Aig_ObjIsNode(pObj)");
    Value1 = Pdr_ManSimInfoGet( pAig, Aig_ObjFanin1(pObj) );
    if Aig_ObjFaninC1(pObj) {
        Value1 = Pdr_ManSimInfoNot( Value1 ); }
    Value = Pdr_ManSimInfoAnd( Value0, Value1 );
    Pdr_ManSimInfoSet( pAig, pObj, Value );
    return Value;
}

/**Function*************************************************************

  Synopsis    [Performs ternary simulation for one design.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManSimDataInit(pAig *Aig_Man_t, vCiObjs *Vec_Int_t, vCiVals *Vec_Int_t, vNodes *Vec_Int_t, vCoObjs *Vec_Int_t, vCoVals *Vec_Int_t, vCi2Rem *Vec_Int_t) int32 {
    var pObj *Aig_Obj_t;
    var i int32;
    // set the CI values
    Pdr_ManSimInfoSet( pAig, Aig_ManConst1(pAig), PDR_ONE );
    Aig_ManForEachObjVec( vCiObjs, pAig, pObj, i )
        Pdr_ManSimInfoSet( pAig, pObj, (Vec_IntEntry(vCiVals, i)?PDR_ONE:PDR_ZER) );
    // set the FOs to remove
    if vCi2Rem != nil {
    Aig_ManForEachObjVec( vCi2Rem, pAig, pObj, i )
        Pdr_ManSimInfoSet( pAig, pObj, PDR_UND ); }
    // perform ternary simulation
    Aig_ManForEachObjVec( vNodes, pAig, pObj, i )
        Pdr_ManExtendOneEval( pAig, pObj );
    // transfer results to the output
    Aig_ManForEachObjVec( vCoObjs, pAig, pObj, i )
        Pdr_ManExtendOneEval( pAig, pObj );
    // check the results
    Aig_ManForEachObjVec( vCoObjs, pAig, pObj, i )
        if ( Pdr_ManSimInfoGet( pAig, pObj ) != (Vec_IntEntry(vCoVals, i)?PDR_ONE:PDR_ZER) )
            return 0;
    return 1;
}

/**Function*************************************************************

  Synopsis    [Tries to assign ternary value to one of the CIs.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManExtendOne(pAig *Aig_Man_t, pObj *Aig_Obj_t, vUndo *Vec_Int_t, vVis *Vec_Int_t) int32 {
    var pFanout *Aig_Obj_t;
    var i int32;
    var k int32;
    var iFanout int32 = -1;
    var Value int32;
    var Value2 int32;
    assert(Saig_ObjIsLo(pAig, pObj), "Saig_ObjIsLo(pAig, pObj)");
    assert(Aig_ObjIsTravIdCurrent(pAig, pObj), "Aig_ObjIsTravIdCurrent(pAig, pObj)");
    // save original value
    Value = Pdr_ManSimInfoGet( pAig, pObj );
    assert(Value == PDR_ZER || Value == PDR_ONE, "Value == PDR_ZER || Value == PDR_ONE");
    Vec_IntPush( vUndo, Aig_ObjId(pObj) );
    Vec_IntPush( vUndo, Value );
    // update original value
    Pdr_ManSimInfoSet( pAig, pObj, PDR_UND );
    // traverse
    Vec_IntClear( vVis );
    Vec_IntPush( vVis, Aig_ObjId(pObj) );
    Aig_ManForEachObjVec( vVis, pAig, pObj, i )
    {
        Aig_ObjForEachFanout( pAig, pObj, pFanout, iFanout, k )
        {
            if !Aig_ObjIsTravIdCurrent(pAig, pFanout) {
                continue; }
            assert(Aig_ObjId(pObj) < Aig_ObjId(pFanout), "Aig_ObjId(pObj) < Aig_ObjId(pFanout)");
            Value = Pdr_ManSimInfoGet( pAig, pFanout );
            if Value == PDR_UND {
                continue; }
            Value2 = Pdr_ManExtendOneEval( pAig, pFanout );
            if Value2 == Value {
                continue; }
            assert(Value2 == PDR_UND, "Value2 == PDR_UND");
            Vec_IntPush( vUndo, Aig_ObjId(pFanout) );
            Vec_IntPush( vUndo, Value );
            if Aig_ObjIsCo(pFanout) {
                return 0; }
            assert(Aig_ObjIsNode(pFanout), "Aig_ObjIsNode(pFanout)");
            Vec_IntPushOrder( vVis, Aig_ObjId(pFanout) );
        }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Undoes the partial results of ternary simulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManExtendUndo(pAig *Aig_Man_t, vUndo *Vec_Int_t) {
    var pObj *Aig_Obj_t;
    var i, Value int32;
    Aig_ManForEachObjVec( vUndo, pAig, pObj, i )
    {
        Value  = Vec_IntEntry(vUndo, ++i);
        assert(Pdr_ManSimInfoGet(pAig, pObj) == PDR_UND, "Pdr_ManSimInfoGet(pAig, pObj) == PDR_UND");
        Pdr_ManSimInfoSet( pAig, pObj, Value );
    }
}

/**Function*************************************************************

  Synopsis    [Derives the resulting cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManDeriveResult(pAig *Aig_Man_t, vCiObjs *Vec_Int_t, vCiVals *Vec_Int_t, vCi2Rem *Vec_Int_t, vRes *Vec_Int_t, vPiLits *Vec_Int_t) {
    var pObj *Aig_Obj_t;
    var i, Lit int32;
    // mark removed flop outputs
    Aig_ManIncrementTravId( pAig );
    Aig_ManForEachObjVec( vCi2Rem, pAig, pObj, i )
    {
        assert(Saig_ObjIsLo( pAig, pObj ), "Saig_ObjIsLo( pAig, pObj )");
        Aig_ObjSetTravIdCurrent(pAig, pObj);
    }
    // collect flop outputs that are not marked
    Vec_IntClear( vRes );
    Vec_IntClear( vPiLits );
    Aig_ManForEachObjVec( vCiObjs, pAig, pObj, i )
    {
        if Saig_ObjIsPi(pAig, pObj) {
            Lit = Abc_Var2Lit( Aig_ObjCioId(pObj), (Vec_IntEntry(vCiVals, i) == 0) );
            Vec_IntPush( vPiLits, Lit );
            continue;
        }
        assert(Saig_ObjIsLo(pAig, pObj), "Saig_ObjIsLo(pAig, pObj)");
        if Aig_ObjIsTravIdCurrent(pAig, pObj) {
            continue; }
        Lit = Abc_Var2Lit( Aig_ObjCioId(pObj) - Saig_ManPiNum(pAig), (Vec_IntEntry(vCiVals, i) == 0) );
        Vec_IntPush( vRes, Lit );
    }
    if Vec_IntSize(vRes) == 0 {
        Vec_IntPush(vRes, 0); }
}

/**Function*************************************************************

  Synopsis    [Derives the resulting cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManPrintCex(pAig *Aig_Man_t, vCiObjs *Vec_Int_t, vCiVals *Vec_Int_t, vCi2Rem *Vec_Int_t) {
    var pObj *Aig_Obj_t;
    var i int32;
    var pBuff string  = ABC_ALLOC( char, Aig_ManCiNum(pAig)+1 );
    for  i = 0; i < Aig_ManCiNum(pAig); i++  {
        pBuff[i] = '-';
    }
    pBuff[i] = 0;
    Aig_ManForEachObjVec( vCiObjs, pAig, pObj, i )
        pBuff[Aig_ObjCioId(pObj)] = (Vec_IntEntry(vCiVals, i)? '1':'0');
    if vCi2Rem != nil {
    Aig_ManForEachObjVec( vCi2Rem, pAig, pObj, i )
        pBuff[Aig_ObjCioId(pObj)] = 'x'; }
    Abc_Print( 1, "%s\n", pBuff );
    ABC_FREE( pBuff );
}

/**Function*************************************************************

  Synopsis    [Shrinks values using ternary simulation.]

  Description [The cube contains the set of flop index literals which,
  when converted into a clause and applied to the combinational outputs, 
  led to a satisfiable SAT run in frame k (values stored in the SAT solver).
  If the cube is NULL, it is assumed that the first property output was
  asserted and failed.
  The resulting array is a set of flop index literals that asserts the COs.
  Priority contains 0 for i-th entry if the i-th FF is desirable to remove.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Pdr_ManTernarySim(p *Pdr_Man_t, k int32, pCube *Pdr_Set_t) *Pdr_Set_t {
    var pRes *Pdr_Set_t;
    var vPrio *Vec_Int_t    = p.vPrio;    // priority flops (flop indices)
    var vPiLits *Vec_Int_t  = p.vLits;    // array of literals (0/1 PI values)
    var vCiObjs *Vec_Int_t  = p.vCiObjs;  // cone leaves (CI obj IDs)
    var vCoObjs *Vec_Int_t  = p.vCoObjs;  // cone roots (CO obj IDs)
    var vCiVals *Vec_Int_t  = p.vCiVals;  // cone leaf values (0/1 CI values)
    var vCoVals *Vec_Int_t  = p.vCoVals;  // cone root values (0/1 CO values)
    var vNodes *Vec_Int_t   = p.vNodes;   // cone nodes (node obj IDs)
    var vUndo *Vec_Int_t    = p.vUndo;    // cone undos (node obj IDs)
    var vVisits *Vec_Int_t  = p.vVisits;  // intermediate (obj IDs)
    var vCi2Rem *Vec_Int_t  = p.vCi2Rem;  // CIs to be removed (CI obj IDs)
    var vRes *Vec_Int_t     = p.vRes;     // final result (flop literals)
    var pObj *Aig_Obj_t;
    var i, Entry, RetValue int32;
    //abctime clk = Abc_Clock();

    // collect CO objects
    Vec_IntClear( vCoObjs );
    if pCube == nil { // the target is the property output
    
//        Vec_IntPush( vCoObjs, Aig_ObjId(Aig_ManCo(p->pAig, (p->pPars->iOutput==-1)?0:p->pPars->iOutput)) );
        Vec_IntPush( vCoObjs, Aig_ObjId(Aig_ManCo(p.pAig, p.iOutCur)) );
    
    } else { // the target is the cube
    
        for  i = 0; i < pCube.nLits; i++  {
            if pCube.Lits[i] == -1 {
                continue; }
            pObj = Saig_ManLi(p.pAig, (pCube.Lits[i] >> 1));
            Vec_IntPush( vCoObjs, Aig_ObjId(pObj) );
        }
    }
if p.pPars.fVeryVerbose {
Abc_Print( 1, "Trying to justify cube " );
if pCube != nil {
    Pdr_SetPrint( os.Stdout, pCube, Aig_ManRegNum(p.pAig), nil );
} else {
    Abc_Print( 1, "<prop=fail>" ); }
Abc_Print( 1, " in frame %d.\n", k );
}

    // collect CI objects
    Pdr_ManCollectCone( p.pAig, vCoObjs, vCiObjs, vNodes );
    // collect values
    Pdr_ManCollectValues( p, k, vCiObjs, vCiVals );
    Pdr_ManCollectValues( p, k, vCoObjs, vCoVals );
    // simulate for the first time
if p.pPars.fVeryVerbose {
Pdr_ManPrintCex( p.pAig, vCiObjs, vCiVals, nil ); }
    RetValue = Pdr_ManSimDataInit( p.pAig, vCiObjs, vCiVals, vNodes, vCoObjs, vCoVals, nil );
    assert((RetValue != 0), "RetValue");

    // iteratively remove flops
    if p.pPars.fFlopPrio {
        // collect flops and sort them by priority
        Vec_IntClear( vRes );
        Aig_ManForEachObjVec( vCiObjs, p.pAig, pObj, i )
        {
            if !Saig_ObjIsLo( p.pAig, pObj ) {
                continue; }
            Entry = Aig_ObjCioId(pObj) - Saig_ManPiNum(p.pAig);
            Vec_IntPush( vRes, Entry );
        }
        Vec_IntSelectSortCost( Vec_IntArray(vRes), Vec_IntSize(vRes), vPrio );

        // try removing flops starting from low-priority to high-priority
        Vec_IntClear( vCi2Rem );
        Vec_IntForEachEntry( vRes, Entry, i )
        {
            pObj = Aig_ManCi( p.pAig, Saig_ManPiNum(p.pAig) + Entry );
            assert(Saig_ObjIsLo( p.pAig, pObj ), "Saig_ObjIsLo( p.pAig, pObj )");
            Vec_IntClear( vUndo );
            if Pdr_ManExtendOne( p.pAig, pObj, vUndo, vVisits ) != 0 {
                Vec_IntPush( vCi2Rem, Aig_ObjId(pObj) );
            } else {
                Pdr_ManExtendUndo( p.pAig, vUndo ); }
        }
    
    } else {
        // try removing low-priority flops first
        Vec_IntClear( vCi2Rem );
        Aig_ManForEachObjVec( vCiObjs, p.pAig, pObj, i )
        {
            if !Saig_ObjIsLo( p.pAig, pObj ) {
                continue; }
            Entry = Aig_ObjCioId(pObj) - Saig_ManPiNum(p.pAig);
            if Vec_IntEntry(vPrio, Entry) {
                continue; }
            Vec_IntClear( vUndo );
            if Pdr_ManExtendOne( p.pAig, pObj, vUndo, vVisits ) != 0 {
                Vec_IntPush( vCi2Rem, Aig_ObjId(pObj) );
            } else {
                Pdr_ManExtendUndo( p.pAig, vUndo ); }
        }
        // try removing high-priority flops next
        Aig_ManForEachObjVec( vCiObjs, p.pAig, pObj, i )
        {
            if !Saig_ObjIsLo( p.pAig, pObj ) {
                continue; }
            Entry = Aig_ObjCioId(pObj) - Saig_ManPiNum(p.pAig);
            if !Vec_IntEntry(vPrio, Entry) {
                continue; }
            Vec_IntClear( vUndo );
            if Pdr_ManExtendOne( p.pAig, pObj, vUndo, vVisits ) != 0 {
                Vec_IntPush( vCi2Rem, Aig_ObjId(pObj) );
            } else {
                Pdr_ManExtendUndo( p.pAig, vUndo ); }
        }
    }

if p.pPars.fVeryVerbose {
Pdr_ManPrintCex( p.pAig, vCiObjs, vCiVals, vCi2Rem ); }
    RetValue = Pdr_ManSimDataInit( p.pAig, vCiObjs, vCiVals, vNodes, vCoObjs, vCoVals, vCi2Rem );
    assert((RetValue != 0), "RetValue");

    // derive the set of resulting registers
    Pdr_ManDeriveResult( p.pAig, vCiObjs, vCiVals, vCi2Rem, vRes, vPiLits );
    assert(Vec_IntSize(vRes) > 0, "Vec_IntSize(vRes) > 0");
    //p->tTsim += Abc_Clock() - clk;

    // move abstracted literals from flops to inputs
    if p.pPars.fUseAbs && p.vAbsFlops {
        var i, iLit, k int32;
        Vec_IntForEachEntry( vRes, iLit, i )
        {
            if Vec_IntEntry(p.vAbsFlops, Abc_Lit2Var(iLit)) { // used flop
                Vec_IntWriteEntry( vRes, k++, iLit );
            } else {
                Vec_IntPush( vPiLits, 2*Saig_ManPiNum(p.pAig) + iLit ); }
        }
        Vec_IntShrink( vRes, k );
    }
    pRes = Pdr_SetCreate( vRes, vPiLits );
    //ZH: Disabled assertion because this invariant doesn't hold with down
    //because of the join operation which can bring in initial states
    //assert( k == 0 || !Pdr_SetIsInit(pRes, -1) );
    return pRes;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
