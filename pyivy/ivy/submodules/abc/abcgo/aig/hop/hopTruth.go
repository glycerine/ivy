package hop

/**CFile****************************************************************

  FileName    [hopTruth.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Minimalistic And-Inverter Graph package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - May 11, 2006.]

  Revision    [$Id: hopTruth.c,v 1.00 2006/05/11 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "hop.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

func Hop_ManTruthWordNum(nVars int32) int32 { if nVars <= 5 {
    return 1;
} else {
    return (1 << (nVars - 5));
} }

func Hop_ManTruthCopy(pOut *uint32, pIn *uint32, nVars int32) {
    var w int32;
    for  w = Hop_ManTruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn[w];
    }
}
func Hop_ManTruthClear(pOut *uint32, nVars int32) {
    var w int32;
    for  w = Hop_ManTruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = 0;
    }
}
func Hop_ManTruthFill(pOut *uint32, nVars int32) {
    var w int32;
    for  w = Hop_ManTruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = ~uint32(0);
    }
}
func Hop_ManTruthNot(pOut *uint32, pIn *uint32, nVars int32) {
    var w int32;
    for  w = Hop_ManTruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = ~pIn[w];
    }
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////


/**Function*************************************************************

  Synopsis    [Construct BDDs and mark AIG nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManConvertAigToTruth_rec1(pObj *Hop_Obj_t) int32 {
    var Counter int32;
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || Hop_ObjIsMarkA(pObj) {
        return 0; }
    Counter += Hop_ManConvertAigToTruth_rec1( Hop_ObjFanin0(pObj) ); 
    Counter += Hop_ManConvertAigToTruth_rec1( Hop_ObjFanin1(pObj) );
    assert(!Hop_ObjIsMarkA(pObj), "!Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjSetMarkA( pObj );
    return Counter + 1;
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManConvertAigToTruth_rec2(pObj *Hop_Obj_t, vTruth *Vec_Int_t, nWords int32) *uint32 {
    var pTruth, pTruth0, pTruth1 *uint32;
    var i int32;
    assert(!Hop_IsComplement(pObj), "!Hop_IsComplement(pObj)");
    if !Hop_ObjIsNode(pObj) || !Hop_ObjIsMarkA(pObj) {
        return (unsigned *)pObj.pData; }
    // compute the truth tables of the fanins
    pTruth0 = Hop_ManConvertAigToTruth_rec2( Hop_ObjFanin0(pObj), vTruth, nWords );
    pTruth1 = Hop_ManConvertAigToTruth_rec2( Hop_ObjFanin1(pObj), vTruth, nWords );
    // creat the truth table of the node
    pTruth  = Vec_IntFetch( vTruth, nWords );
    if Hop_ObjIsExor(pObj) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = pTruth0[i] ^ pTruth1[i];
        }
    } else if !Hop_ObjFaninC0(pObj) && !Hop_ObjFaninC1(pObj) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = pTruth0[i] & pTruth1[i];
        }
    } else if !Hop_ObjFaninC0(pObj) && Hop_ObjFaninC1(pObj) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = pTruth0[i] & ~pTruth1[i];
        }
    } else if Hop_ObjFaninC0(pObj) && !Hop_ObjFaninC1(pObj) {
        for  i = 0; i < nWords; i++  {
            pTruth[i] = ~pTruth0[i] & pTruth1[i];
        }
    } else { // if ( Hop_ObjFaninC0(pObj) && Hop_ObjFaninC1(pObj) )
        for  i = 0; i < nWords; i++  {
            pTruth[i] = ~pTruth0[i] & ~pTruth1[i];
        } }
    assert(Hop_ObjIsMarkA(pObj), "Hop_ObjIsMarkA(pObj)"); // loop detection
    Hop_ObjClearMarkA( pObj );
    pObj.pData = pTruth;
    return pTruth;
}

/**Function*************************************************************

  Synopsis    [Computes truth table of the node.]

  Description [Assumes that the structural support is no more than 8 inputs.
  Uses array vTruth to store temporary truth tables. The returned pointer should 
  be used immediately.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Hop_ManConvertAigToTruth(p *Hop_Man_t, pRoot *Hop_Obj_t, nVars int32, vTruth *Vec_Int_t, fMsbFirst int32) *uint32 {
    static unsigned uTruths[8][8] = { // elementary truth tables
        { 0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA,0xAAAAAAAA },
        { 0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC,0xCCCCCCCC },
        { 0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0,0xF0F0F0F0 },
        { 0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00,0xFF00FF00 },
        { 0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000,0xFFFF0000 }, 
        { 0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF,0x00000000,0xFFFFFFFF }, 
        { 0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF,0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF }, 
        { 0x00000000,0x00000000,0x00000000,0x00000000,0xFFFFFFFF,0xFFFFFFFF,0xFFFFFFFF,0xFFFFFFFF } 
    };
    var pObj *Hop_Obj_t;
    var pTruth, pTruth2 *uint32;
    var i, nWords, nNodes int32;
    var vTtElems *Vec_Ptr_t;

    // if the number of variables is more than 8, allocate truth tables
    if nVars > 8 {
        vTtElems = Vec_PtrAllocTruthTables( nVars );
    } else {
        vTtElems = nil; }

    // clear the data fields and set marks
    nNodes = Hop_ManConvertAigToTruth_rec1( Hop_Regular(pRoot) );
    // prepare memory
    nWords = Hop_TruthWordNum( nVars );
    Vec_IntClear( vTruth );
    Vec_IntGrow( vTruth, nWords * (nNodes+1) );
    pTruth = Vec_IntFetch( vTruth, nWords );
    // check the case of a constant
    if Hop_ObjIsConst1( Hop_Regular(pRoot) ) {
        assert(nNodes == 0, "nNodes == 0");
        if Hop_IsComplement(pRoot) {
            Hop_ManTruthClear( pTruth, nVars );
        } else {
            Hop_ManTruthFill( pTruth, nVars ); }
        return pTruth;
    }
    // set elementary truth tables at the leaves
    assert(nVars <= Hop_ManPiNum(p), "nVars <= Hop_ManPiNum(p)");
//    assert( Hop_ManPiNum(p) <= 8 ); 
    if fMsbFirst != 0 {
//        Hop_ManForEachPi( p, pObj, i )
        for  i = 0; i < nVars; i++  {
            pObj = Hop_ManPi( p, i );
            if vTtElems != nil {
                pObj.pData = Vec_PtrEntry(vTtElems, nVars-1-i);
            } else {               
                pObj.pData = (void *)uTruths[nVars-1-i]; }
        }
    
    } else {
//        Hop_ManForEachPi( p, pObj, i )
        for  i = 0; i < nVars; i++  {
            pObj = Hop_ManPi( p, i );
            if vTtElems != nil {
                pObj.pData = Vec_PtrEntry(vTtElems, i);
            } else {               
                pObj.pData = (void *)uTruths[i]; }
        }
    }
    // clear the marks and compute the truth table
    pTruth2 = Hop_ManConvertAigToTruth_rec2( Hop_Regular(pRoot), vTruth, nWords );
    // copy the result
    Hop_ManTruthCopy( pTruth, pTruth2, nVars );
    if Hop_IsComplement(pRoot) {
        Hop_ManTruthNot( pTruth, pTruth, nVars ); }
    if vTtElems != nil {
        Vec_PtrFree( vTtElems ); }
    return pTruth;
}


/**Function*************************************************************

  Synopsis    [Compute truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var Truth [8]word = [8]word{
    ABC_CONST(0xAAAAAAAAAAAAAAAA),
    ABC_CONST(0xCCCCCCCCCCCCCCCC),
    ABC_CONST(0xF0F0F0F0F0F0F0F0),
    ABC_CONST(0xFF00FF00FF00FF00),
    ABC_CONST(0xFFFF0000FFFF0000),
    ABC_CONST(0xFFFFFFFF00000000),
    ABC_CONST(0x0000000000000000),
    ABC_CONST(0xFFFFFFFFFFFFFFFF),
};
func Hop_ManComputeTruth6_rec(p *Hop_Man_t, pObj *Hop_Obj_t) word {
    var Truth0, Truth1 word;
    if Hop_ObjIsPi(pObj) {
        return Truth[pObj.iData]; }
    assert(Hop_ObjIsNode(pObj), "Hop_ObjIsNode(pObj)");
    Truth0 = Hop_ManComputeTruth6_rec( p, Hop_ObjFanin0(pObj) );
    Truth1 = Hop_ManComputeTruth6_rec( p, Hop_ObjFanin1(pObj) );
    if Hop_ObjFaninC0(pObj) {
        Truth0 = ~Truth0;
    } else {
        Truth0 = Truth0;
    }
    if Hop_ObjFaninC1(pObj) {
        Truth1 = ~Truth1;
    } else {
        Truth1 = Truth1;
    }
    return Truth0 & Truth1;
}
func Hop_ManComputeTruth6(p *Hop_Man_t, pObj *Hop_Obj_t, nVars int32) word {
    var Truth word;
    var i int32;
    if Hop_ObjIsConst1( Hop_Regular(pObj) ) {
        if Hop_IsComplement(pObj) {
            return 0;
        } else {
            return ~(word)0;
        } }
    for  i = 0; i < nVars; i++  {
        Hop_ManPi( p, i ).iData = i;
    }
    Truth = Hop_ManComputeTruth6_rec( p, Hop_Regular(pObj) );
    if Hop_IsComplement(pObj) {
        return ~Truth;
    } else {
        return Truth;
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

