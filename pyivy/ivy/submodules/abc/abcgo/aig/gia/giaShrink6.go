package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaShrink6.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Implementation of DAG-aware unmapping for 6-input cuts.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaShrink6.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "bool/bdc/bdc.h"
//#include "bool/rsb/rsb.h"
//#include "misc/util/utilTruth.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

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

// fanout structure
   
type Shr_Fan_t struct {
    iFan                int32;           // fanout ID
    Next                int32;           // next structure
}

// operation manager
   
type Shr_Man_t struct {
    pGia         *Gia_Man_t;           // user's AIG
    pNew         *Gia_Man_t;           // constructed AIG
    nDivMax                int32;        // max number of divisors
    nNewSize                int32;       // max growth size
    // dynamic fanout (can only grow)
    vFanMem         *Vec_Wrd_t;        // fanout memory
    vObj2Fan         *Vec_Int_t;       // fanout
    pFanTemp         *Shr_Fan_t;       // temporary fanout
    // divisors
    vDivs         *Vec_Int_t;          // divisors
    vPrio         *Vec_Int_t;          // priority queue
    vDivResub         *Vec_Int_t;      // resubstitution
    vLeaves         *Vec_Int_t;        // cut leaves
    // truth tables
    vTruths         *Vec_Wrd_t;        // truth tables
    vDivTruths         *Vec_Wrd_t;     // truth tables
    // bidecomposition
    pManRsb         *Rsb_Man_t;
    pManDec         *Bdc_Man_t;       
    Pars          Bdc_Par_t;
    // statistics 
}

// Original macro:
// #define Shr_ObjForEachFanout( p, iNode, iFan )                      \
//     for ( iFan = Shr_ManFanIterStart(p, iNode); iFan; iFan = Shr_ManFanIterNext(p) )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Shr_ManAlloc(pGia *Gia_Man_t) *Shr_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Shr_Man_t{};
    p.nDivMax     = 64;
    p.nNewSize    = 2 * Gia_ManObjNum(pGia);
    p.pGia        = pGia;
    p.vFanMem     = Vec_WrdAlloc( 1000 );   Vec_WrdPush(p.vFanMem, -1); 
    p.vObj2Fan    = Vec_IntStart( p.nNewSize );
    p.vDivs       = Vec_IntAlloc( 1000 );
    p.vPrio       = Vec_IntAlloc( 1000 );
    p.vTruths     = Vec_WrdStart( p.nNewSize );
    p.vDivTruths  = Vec_WrdAlloc( 100 );
    p.vDivResub   = Vec_IntAlloc( 6 );
    p.vLeaves     = Vec_IntAlloc( 6 );
    // start new manager
    p.pNew        = Gia_ManStart( p.nNewSize );
    p.pNew.pName = Abc_UtilStrsav( pGia.pName );
    p.pNew.pSpec = Abc_UtilStrsav( pGia.pSpec );
    Gia_ManHashAlloc( p.pNew );
    Gia_ManCleanLevels( p.pNew, p.nNewSize );
    // allocate traversal IDs
    p.pNew.nObjs = p.nNewSize;
    Gia_ManIncrementTravId( p.pNew );
    p.pNew.nObjs = 1;
    // start decompostion
    p.Pars.nVarsMax = 6;
    p.Pars.fVerbose = 0;
    p.pManDec = Bdc_ManAlloc( &p.Pars );
    p.pManRsb = Rsb_ManAlloc( 6, p.nDivMax, 4, 1 );
    return p;
}
func Shr_ManFree(p *Shr_Man_t) *Gia_Man_t {
    // prepare the manager
    var pTemp *Gia_Man_t;
    Gia_ManHashStop( p.pNew );
    Vec_IntFreeP( &p.pNew.vLevels );
    if Gia_ManHasDangling(p.pNew) {
        p.pNew = Gia_ManCleanup( pTemp = p.pNew );
        if Gia_ManAndNum(p.pNew) != Gia_ManAndNum(pTemp) {
            fmt.Printf("Node reduction after sweep %6d -> %6d.\n", Gia_ManAndNum(pTemp), Gia_ManAndNum(p.pNew)); }
        Gia_ManStop( pTemp );
    }
    Gia_ManSetRegNum( p.pNew, Gia_ManRegNum(p.pGia) );
    pTemp = p.pNew; p.pNew = nil;
    // free data structures
    Rsb_ManFree( p.pManRsb );
    Bdc_ManFree( p.pManDec );
    Gia_ManStopP( &p.pNew );
    Vec_WrdFree( p.vFanMem );
    Vec_IntFree( p.vObj2Fan );
    Vec_IntFree( p.vDivs );
    Vec_IntFree( p.vPrio );
    Vec_WrdFree( p.vTruths );
    Vec_WrdFree( p.vDivTruths );
    Vec_IntFree( p.vDivResub );
    Vec_IntFree( p.vLeaves );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
    return pTemp;
}


/**Function*************************************************************

  Synopsis    [Fanout manipulation.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Shr_ManAddFanout(p *Shr_Man_t, iFanin int32, iFanout int32) {
    union {
        var sFan Shr_Fan_t;
        var sWord word;
    }
 FanStr;
    FanStr.sFan.iFan = iFanout;
    FanStr.sFan.Next = Vec_IntEntry(p.vObj2Fan, iFanin);
    Vec_IntWriteEntry( p.vObj2Fan, iFanin, Vec_WrdSize(p.vFanMem) );
    Vec_WrdPush(p.vFanMem, FanStr.sWord );
}
func Shr_ManFanIterStart(p *Shr_Man_t, iNode int32) int32 {
    if Vec_IntEntry(p.vObj2Fan, iNode) == 0 {
        return 0; }
    p.pFanTemp = (Shr_Fan_t *)Vec_WrdEntryP( p.vFanMem, Vec_IntEntry(p.vObj2Fan, iNode) );
    return p.pFanTemp.iFan;
}
func Shr_ManFanIterNext(p *Shr_Man_t) int32 {
    if p.pFanTemp.Next == 0 {
        return 0; }
    p.pFanTemp = (Shr_Fan_t *)Vec_WrdEntryP( p.vFanMem, p.pFanTemp.Next );
    return p.pFanTemp.iFan;
}

/**Function*************************************************************

  Synopsis    [Collect divisors.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Shr_ManDivPushOrderByLevel(p *Shr_Man_t, iDiv int32) int32 {
    var iPlace int32;
    var pArray *int32;
    Vec_IntPush( p.vPrio, iDiv );
    if Vec_IntSize(p.vPrio) == 1 {
        return 0; } 
    pArray = Vec_IntArray(p.vPrio);
    for  iPlace = Vec_IntSize(p.vPrio) - 1; iPlace > 0; iPlace--  {
        if Gia_ObjLevel(p.pNew, Gia_ManObj(p.pNew, pArray[iPlace-1])) > 
             Gia_ObjLevel(p.pNew, Gia_ManObj(p.pNew, pArray[iPlace])) {
             // inlining of macro ABC_SWAP
              var t int32  = pArray[iPlace-1]; pArray[iPlace-1] = pArray[iPlace]; pArray[iPlace] = t; 
        } else {
            break; }
    }
    return iPlace; // the place of the new divisor
}
func Shr_ManCollectDivisors(p *Shr_Man_t, vLeaves *Vec_Int_t, Limit int32, nFanoutMax int32) int32 {
    var pFan *Gia_Obj_t;
    var i, c, iDiv, iFan, iPlace int32;
    assert(Limit > 6, "Limit > 6");
    Vec_IntClear( p.vDivs );
    Vec_IntClear( p.vPrio );
    Gia_ManIncrementTravId( p.pNew );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iDiv) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        Vec_IntPush( p.vDivs, iDiv );
        Shr_ManDivPushOrderByLevel( p, iDiv );
        Gia_ObjSetTravIdCurrentId( p.pNew, iDiv );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vPrio)) && (((iDiv) = Vec_IntEntry(p.vPrio, i)), 1); i++  {
        c = 0;
        assert(Gia_ObjIsTravIdCurrentId(p.pNew, iDiv), "Gia_ObjIsTravIdCurrentId(p.pNew, iDiv)");
        // inlining of macro Shr_ObjForEachFanout
        for  iFan = Shr_ManFanIterStart(p, iDiv); (iFan != 0); iFan = Shr_ManFanIterNext(p)  {
            if c++ == nFanoutMax {
                break; }
            if Gia_ObjIsTravIdCurrentId(p.pNew, iFan) {
                continue; }
            pFan = Gia_ManObj( p.pNew, iFan );
            assert(Gia_ObjIsAnd(pFan), "Gia_ObjIsAnd(pFan)");
            assert(Gia_ObjLevel(p.pNew, Gia_ManObj(p.pNew, iDiv)) < Gia_ObjLevel(p.pNew, pFan), "Gia_ObjLevel(p.pNew, Gia_ManObj(p.pNew, iDiv)) < Gia_ObjLevel(p.pNew, pFan)");
            if !Gia_ObjIsTravIdCurrentId(p.pNew, Gia_ObjFaninId0(pFan, iFan)) ||
                 !Gia_ObjIsTravIdCurrentId(p.pNew, Gia_ObjFaninId1(pFan, iFan)) {
                continue; }
            Vec_IntPush( p.vDivs, iFan );
            Gia_ObjSetTravIdCurrentId( p.pNew, iFan );
            iPlace = Shr_ManDivPushOrderByLevel( p, iFan );
            assert(i < iPlace, "i < iPlace");
            if Vec_IntSize(p.vDivs) == Limit {
                return Vec_IntSize(p.vDivs); }
        }
    }
    return Vec_IntSize(p.vDivs);
}

/**Function*************************************************************

  Synopsis    [Resynthesizes nodes using bi-decomposition.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Shr_ObjPerformBidec(p *Shr_Man_t, pManDec *Bdc_Man_t, pNew *Gia_Man_t, vLeafLits *Vec_Int_t, uTruth1 word, uTruthC word) int32 {
    var pFunc *Bdc_Fun_t;
    var pObj *Gia_Obj_t;
    var i, iVar, iLit, nNodes, iLast int32;
    var nVars int32  = Vec_IntSize(vLeafLits);
    assert(uTruth1 != 0 && uTruthC != 0, "uTruth1 != 0 && uTruthC != 0");
    Bdc_ManDecompose( pManDec, (unsigned *)&uTruth1, (unsigned *)&uTruthC, nVars, nil, 1000 );
    Bdc_FuncSetCopyInt( Bdc_ManFunc(pManDec, 0), 1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeafLits)) && (((iVar) = Vec_IntEntry(vLeafLits, i)), 1); i++  {
        Bdc_FuncSetCopyInt( Bdc_ManFunc(pManDec, i+1), Abc_Var2Lit(iVar, 0) );
    }
    nNodes = Bdc_ManNodeNum( pManDec );
    for  i = nVars + 1; i < nNodes; i++  {
        pFunc = Bdc_ManFunc( pManDec, i );
        iLast = Gia_ManObjNum(pNew);
        iLit  = Gia_ManHashAnd( pNew, Bdc_FunFanin0Copy(pFunc), Bdc_FunFanin1Copy(pFunc) );
        Bdc_FuncSetCopyInt( pFunc, iLit );
        if iLast == Gia_ManObjNum(pNew) {
            continue; }
        assert(iLast + 1 == Gia_ManObjNum(pNew), "iLast + 1 == Gia_ManObjNum(pNew)");
        pObj = Gia_ManObj(pNew, Abc_Lit2Var(iLit));
        Gia_ObjSetAndLevel( pNew, pObj );
        Shr_ManAddFanout( p, Gia_ObjFaninId0p(pNew, pObj), Gia_ObjId(pNew, pObj) );
        Shr_ManAddFanout( p, Gia_ObjFaninId1p(pNew, pObj), Gia_ObjId(pNew, pObj) );
        assert(Gia_ManObjNum(pNew) < p.nNewSize, "Gia_ManObjNum(pNew) < p.nNewSize");
    }
    return Bdc_FunObjCopy( Bdc_ManRoot(pManDec) );
}


/**Function*************************************************************

  Synopsis    [Compute truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Shr_ManComputeTruth6_rec(p *Gia_Man_t, iNode int32, vTruths *Vec_Wrd_t) word {
    var pObj *Gia_Obj_t;
    var Truth0, Truth1 word;
    if Gia_ObjIsTravIdCurrentId(p, iNode) {
        return Vec_WrdEntry(vTruths, iNode); }
    Gia_ObjSetTravIdCurrentId(p, iNode);
    pObj = Gia_ManObj( p, iNode );
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Truth0 = Shr_ManComputeTruth6_rec( p, Gia_ObjFaninId0p(p, pObj), vTruths );
    Truth1 = Shr_ManComputeTruth6_rec( p, Gia_ObjFaninId1p(p, pObj), vTruths );
    if Gia_ObjFaninC0(pObj) {
        Truth0 = ~Truth0; }
    if Gia_ObjFaninC1(pObj) {
        Truth1 = ~Truth1; }
    Vec_WrdWriteEntry( vTruths, iNode, Truth0 & Truth1 );
    return Truth0 & Truth1;
}
func Shr_ManComputeTruth6(p *Gia_Man_t, pObj *Gia_Obj_t, vLeaves *Vec_Int_t, vTruths *Vec_Wrd_t) word {
    var i, iLeaf int32;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManIncrementTravId( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeaves)) && (((iLeaf) = Vec_IntEntry(vLeaves, i)), 1); i++  {
        Gia_ObjSetTravIdCurrentId( p, iLeaf );
        Vec_WrdWriteEntry( vTruths, iLeaf, Truth[i] );
    }
    return Shr_ManComputeTruth6_rec( p, Gia_ObjId(p, pObj), vTruths );
}

/**Function*************************************************************

  Synopsis    [Compute truth table.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Shr_ManComputeTruths(p *Gia_Man_t, nVars int32, vDivs *Vec_Int_t, vDivTruths *Vec_Wrd_t, vTruths *Vec_Wrd_t) {
    var pObj *Gia_Obj_t;
    var Truth0, Truth1 word;//, Truthh;
    var i, iDiv int32;
    Vec_WrdClear( vDivTruths );
    // inlining of macro Vec_IntForEachEntryStop
    for  i = 0; (i < nVars) && (((iDiv) = Vec_IntEntry(vDivs, i)), 1); i++  {
        Vec_WrdWriteEntry( vTruths, iDiv, Truth[i] );
        Vec_WrdPush( vDivTruths, Truth[i] );
    }
    // inlining of macro Vec_IntForEachEntryStart
    for  i = nVars; (i < Vec_IntSize(vDivs)) && (((iDiv) = Vec_IntEntry(vDivs, i)), 1); i++  {
        pObj = Gia_ManObj( p, iDiv );
        Truth0 = Vec_WrdEntry( vTruths, Gia_ObjFaninId0(pObj, iDiv) );
        Truth1 = Vec_WrdEntry( vTruths, Gia_ObjFaninId1(pObj, iDiv) );
        if Gia_ObjFaninC0(pObj) {
            Truth0 = ~Truth0; }
        if Gia_ObjFaninC1(pObj) {
            Truth1 = ~Truth1; }
        Vec_WrdWriteEntry( vTruths, iDiv, Truth0 & Truth1 );
        Vec_WrdPush( vDivTruths, Truth0 & Truth1 );

//        Truthh = Truth0 & Truth1;
//        Abc_TtPrintBinary( &Truthh, nVars );   //printf( "\n" );
//        Kit_DsdPrintFromTruth( &Truthh, nVars );    printf( "\n" );
    }
//    printf( "\n" );
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManMapShrink6(p *Gia_Man_t, nFanoutMax int32, fKeepLevel int32, fVerbose int32) *Gia_Man_t {
    var pMan *Shr_Man_t;
    var pObj, pFanin *Gia_Obj_t;
    var uTruth, uTruth0, uTruth1 word;
    var i, k, nDivs, iNode int32;
    var RetValue, Counter1, Counter2 int32;
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    var timeFanout abctime;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    pMan = Shr_ManAlloc( p );
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            pObj.Value = Gia_ManAppendCi( pMan.pNew );
            if p.vLevels {
                Gia_ObjSetLevel( pMan.pNew, Gia_ObjFromLit(pMan.pNew, Gia_ObjValue(pObj)), Gia_ObjLevel(p, pObj) ); }
        
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pMan.pNew, Gia_ObjFanin0Copy(pObj) );
        
        } else if Gia_ObjIsLut(p, i) {
            // collect leaves of this gate  
            Vec_IntClear( pMan.vLeaves );
            // inlining of macro Gia_LutForEachFanin
            for  k = 0; k < Gia_ObjLutSize(p,i) && ((iNode = Gia_ObjLutFanins(p,i)[k]),1); k++  {
                Vec_IntPush( pMan.vLeaves, iNode );
            }
            assert(Vec_IntSize(pMan.vLeaves) <= 6, "Vec_IntSize(pMan.vLeaves) <= 6");
            // compute truth table 
            uTruth = Shr_ManComputeTruth6( pMan.pGia, pObj, pMan.vLeaves, pMan.vTruths );
            assert(pObj.Value == ~0, "pObj.Value == ~0");
            if uTruth == 0 || ~uTruth == 0 {
                pObj.Value = Abc_LitNotCond( 0, ~uTruth == 0 );
            } else {
                // inlining of macro Gia_ManForEachObjVec
                for  k = 0; (k < Vec_IntSize(pMan.vLeaves)) && ((pFanin) = Gia_ManObj(p, Vec_IntEntry(pMan.vLeaves,k))); k++  {
                    if uTruth == Truth[k] || ~uTruth == Truth[k] {
                        pObj.Value = Abc_LitNotCond( pFanin.Value, ~uTruth == Truth[k] ); }
                } }
            if pObj.Value != ~0 {
                continue; }
            // translate into new nodes
            // inlining of macro Gia_ManForEachObjVec
            for  k = 0; (k < Vec_IntSize(pMan.vLeaves)) && ((pFanin) = Gia_ManObj(p, Vec_IntEntry(pMan.vLeaves,k))); k++  {
                if Abc_LitIsCompl(pFanin.Value) {
                    uTruth = ((uTruth & Truth[k]) >> (1 << k)) | ((uTruth & ~Truth[k]) << (1 << k)); }
                Vec_IntWriteEntry( pMan.vLeaves, k, Abc_Lit2Var(pFanin.Value) );
            }
            // compute divisors
            clk2 = Abc_Clock();
            nDivs = Shr_ManCollectDivisors( pMan, pMan.vLeaves, pMan.nDivMax, nFanoutMax );
            assert(nDivs <= pMan.nDivMax, "nDivs <= pMan.nDivMax");
            timeFanout += Abc_Clock() - clk2;
            // compute truth tables
            Shr_ManComputeTruths( pMan.pNew, Vec_IntSize(pMan.vLeaves), pMan.vDivs, pMan.vDivTruths, pMan.vTruths );
            // perform resubstitution
            RetValue = Rsb_ManPerformResub6( pMan.pManRsb, Vec_IntSize(pMan.vLeaves), uTruth, pMan.vDivTruths, &uTruth0, &uTruth1, 0 );
            if RetValue { // resub exists
            
                var vResult *Vec_Int_t  = Rsb_ManGetFanins(pMan.pManRsb);
                Vec_IntClear( pMan.vDivResub );
                // inlining of macro Vec_IntForEachEntry
                for  k = 0; (k < Vec_IntSize(vResult)) && (((iNode) = Vec_IntEntry(vResult, k)), 1); k++  {
                    Vec_IntPush( pMan.vDivResub, Vec_IntEntry(pMan.vDivs, iNode) );
                }
                pObj.Value = Shr_ObjPerformBidec( pMan, pMan.pManDec, pMan.pNew, pMan.vDivResub, uTruth1, uTruth0 | uTruth1 );
                Counter1++;
            
            } else {
                pObj.Value = Shr_ObjPerformBidec( pMan, pMan.pManDec, pMan.pNew, pMan.vLeaves, uTruth, ~(word)0 );
                Counter2++;
            }
        }
    }
    if fVerbose != 0 {
        fmt.Printf("Performed %d resubs and %d decomps.  ", Counter1, Counter2);
        fmt.Printf("Gain in AIG nodes = %d.  ", Gia_ManObjNum(p)-Gia_ManObjNum(pMan.pNew));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Runtime"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
//        ABC_PRT( "Divisors", timeFanout );        
    }
    return Shr_ManFree( pMan );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

