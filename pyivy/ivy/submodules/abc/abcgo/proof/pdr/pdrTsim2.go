package pdr

import (
	"fmt"
	"os"
)

/**CFile****************************************************************

  FileName    [pdrTsim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Property driven reachability.]

  Synopsis    [Improved ternary simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 20, 2010.]

  Revision    [$Id: pdrTsim.c,v 1.00 2010/11/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "pdrInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Txs_Man_t_ struct {
    pGia  *Gia_Man_t;      // user's AIG
    vPrio  *Vec_Int_t;     // priority of each flop
    vCiObjs  *Vec_Int_t;   // cone leaves (CI obj IDs)
    vCoObjs  *Vec_Int_t;   // cone roots (CO obj IDs)
    vCiVals  *Vec_Int_t;   // cone leaf values (0/1 CI values)
    vCoVals  *Vec_Int_t;   // cone root values (0/1 CO values)
    vNodes  *Vec_Int_t;    // cone nodes (node obj IDs)
    vTemp  *Vec_Int_t;     // cone nodes (node obj IDs)
    vPiLits  *Vec_Int_t;   // resulting array of PI literals
    vFfLits  *Vec_Int_t;   // resulting array of flop literals
    pMan  *Pdr_Man_t;      // calling manager
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Start and stop the ternary simulation engine.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs_ManStart(pMan *Pdr_Man_t, pAig *Aig_Man_t, vPrio *Vec_Int_t) *Txs_Man_t {
    var p *Txs_Man_t;
//    Aig_Obj_t * pObj;
//    int i;
    assert(Vec_IntSize(vPrio) == Aig_ManRegNum(pAig), "Vec_IntSize(vPrio) == Aig_ManRegNum(pAig)");
    p = new(Txs_Man_t);
    p.pGia    = Gia_ManFromAigSimple(pAig); // user's AIG
//    Aig_ManForEachObj( pAig, pObj, i )
//        assert( i == 0 || pObj->iData == Abc_Var2Lit(i, 0) );
    p.vPrio   = vPrio;                // priority of each flop
    p.vCiObjs = Vec_IntAlloc( 100 );  // cone leaves (CI obj IDs)
    p.vCoObjs = Vec_IntAlloc( 100 );  // cone roots (CO obj IDs)
    p.vCiVals = Vec_IntAlloc( 100 );  // cone leaf values (0/1 CI values)
    p.vCoVals = Vec_IntAlloc( 100 );  // cone root values (0/1 CO values)
    p.vNodes  = Vec_IntAlloc( 100 );  // cone nodes (node obj IDs)
    p.vTemp   = Vec_IntAlloc( 100 );  // cone nodes (node obj IDs)
    p.vPiLits = Vec_IntAlloc( 100 );  // resulting array of PI literals
    p.vFfLits = Vec_IntAlloc( 100 );  // resulting array of flop literals
    p.pMan    = pMan;                 // calling manager
    return p;
}
func Txs_ManStop(p *Txs_Man_t) {
    Gia_ManStop( p.pGia );
    Vec_IntFree( p.vCiObjs );
    Vec_IntFree( p.vCoObjs );
    Vec_IntFree( p.vCiVals );
    Vec_IntFree( p.vCoVals );
    Vec_IntFree( p.vNodes );
    Vec_IntFree( p.vTemp );
    Vec_IntFree( p.vPiLits );
    Vec_IntFree( p.vFfLits );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Marks the TFI cone and collects CIs and nodes.]

  Description [For this procedure to work Value should not be ~0 
  at the beginning.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs_ManCollectCone_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vCiObjs *Vec_Int_t, vNodes *Vec_Int_t) {
    if !~pObj.Value {
        return; }
    pObj.Value = ~0;
    if Gia_ObjIsCi(pObj) {
        Vec_IntPush( vCiObjs, Gia_ObjId(p, pObj) );
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Txs_ManCollectCone_rec( p, Gia_ObjFanin0(pObj), vCiObjs, vNodes );
    Txs_ManCollectCone_rec( p, Gia_ObjFanin1(pObj), vCiObjs, vNodes );
    Vec_IntPush( vNodes, Gia_ObjId(p, pObj) );
}
func Txs_ManCollectCone(p *Gia_Man_t, vCoObjs *Vec_Int_t, vCiObjs *Vec_Int_t, vNodes *Vec_Int_t) {
    var pObj *Gia_Obj_t; var i int32;
//    printf( "Collecting cones for clause with %d literals.\n", Vec_IntSize(vCoObjs) );
    Vec_IntClear( vCiObjs );
    Vec_IntClear( vNodes );
    Gia_ManConst0(p).Value = ~0;
    Gia_ManForEachObjVec( vCoObjs, p, pObj, i )
        Txs_ManCollectCone_rec( p, Gia_ObjFanin0(pObj), vCiObjs, vNodes );
}

/**Function*************************************************************

  Synopsis    [Propagate values and assign priority.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs_ManForwardPass(p *Gia_Man_t, vPrio *Vec_Int_t, vCiObjs *Vec_Int_t, vCiVals *Vec_Int_t, vNodes *Vec_Int_t, vCoObjs *Vec_Int_t, vCoVals *Vec_Int_t) {
    var pObj, pFan0, pFan1 *Gia_Obj_t; 
    var i, value0, value1 int32;
    pObj = Gia_ManConst0(p);
    pObj.fMark0 = 0;
    pObj.fMark1 = 0;
    Gia_ManForEachObjVec( vCiObjs, p, pObj, i )
    {
        pObj.fMark0 = Vec_IntEntry(vCiVals, i);
        pObj.fMark1 = 0;
        if Gia_ObjIsPi(p, pObj) {
            pObj.Value = 0x7FFFFFFF;
        } else {
            pObj.Value = Vec_IntEntry(vPrio, Gia_ObjCioId(pObj)-Gia_ManPiNum(p));
        }
        assert(~pObj.Value, "~pObj.Value");
    }
    Gia_ManForEachObjVec( vNodes, p, pObj, i )
    {
        pFan0 = Gia_ObjFanin0(pObj);
        pFan1 = Gia_ObjFanin1(pObj);
        value0 = pFan0.fMark0 ^ Gia_ObjFaninC0(pObj);
        value1 = pFan1.fMark0 ^ Gia_ObjFaninC1(pObj);
        pObj.fMark0 = value0 && value1;
        pObj.fMark1 = 0;
        if pObj.fMark0 {
            pObj.Value = Abc_MinInt( pFan0.Value, pFan1.Value );
        } else if value0 != 0 {
            pObj.Value = pFan1.Value;
        } else if value1 != 0 {
            pObj.Value = pFan0.Value;
        } else { // if ( value0 == 0 && value1 == 0 )
            pObj.Value = Abc_MaxInt( pFan0.Value, pFan1.Value ); }
        assert(~pObj.Value, "~pObj.Value");
    }
    Gia_ManForEachObjVec( vCoObjs, p, pObj, i )
    {
        pFan0 = Gia_ObjFanin0(pObj);
        pObj.fMark0 = (pFan0.fMark0 ^ Gia_ObjFaninC0(pObj));
        pFan0.fMark1 = 1;
        assert(int32(pObj.fMark0) == Vec_IntEntry(vCoVals, i), "int32(pObj.fMark0) == Vec_IntEntry(vCoVals, i)");
    }
}

/**Function*************************************************************

  Synopsis    [Propagate requirements and collect results.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs_ObjIsJust(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    var pFan0 *Gia_Obj_t  = Gia_ObjFanin0(pObj);
    var pFan1 *Gia_Obj_t  = Gia_ObjFanin1(pObj);
    var value0 int32  = pFan0.fMark0 ^ Gia_ObjFaninC0(pObj);
    var value1 int32  = pFan1.fMark0 ^ Gia_ObjFaninC1(pObj);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if pObj.fMark0 {
        return pFan0.fMark1 && pFan1.fMark1; }
    assert(!pObj.fMark0, "!pObj.fMark0");
    assert(!(value0 != 0) || !(value1 != 0), "!value0 || !value1");
    if value0 != 0 {
        return pFan1.fMark1 || Gia_ObjIsPi(p, pFan1); }
    if value1 != 0 {
        return pFan0.fMark1 || Gia_ObjIsPi(p, pFan0); }
    assert(!(value0 != 0) && !(value1 != 0), "!value0 && !value1");
    return pFan0.fMark1 || pFan1.fMark1 || Gia_ObjIsPi(p, pFan0) || Gia_ObjIsPi(p, pFan1);
}
func Txs_ManBackwardPass(p *Gia_Man_t, vCiObjs *Vec_Int_t, vNodes *Vec_Int_t, vPiLits *Vec_Int_t, vFfLits *Vec_Int_t) {
    var pObj, pFan0, pFan1 *Gia_Obj_t; var i, value0, value1 int32;
    Gia_ManForEachObjVecReverse( vNodes, p, pObj, i )
    {
        if !pObj.fMark1 {
            continue; }
        pObj.fMark1 = 0;
        pFan0 = Gia_ObjFanin0(pObj);
        pFan1 = Gia_ObjFanin1(pObj);
        if pObj.fMark0 {
            pFan0.fMark1 = 1;
            pFan1.fMark1 = 1;
            continue;
        }
        value0 = pFan0.fMark0 ^ Gia_ObjFaninC0(pObj);
        value1 = pFan1.fMark0 ^ Gia_ObjFaninC1(pObj);
        assert(!(value0 != 0) || !(value1 != 0), "!value0 || !value1");
        if value0 != 0 {
            pFan1.fMark1 = 1;
        } else if value1 != 0 {
            pFan0.fMark1 = 1;
        } else { // if ( !value0 && !value1 )
        
            if pFan0.fMark1 || pFan1.fMark1 {
                continue; }
            if Gia_ObjIsPi(p, pFan0) {
                pFan0.fMark1 = 1;
            } else if Gia_ObjIsPi(p, pFan1) {
                pFan1.fMark1 = 1;
            } else if Gia_ObjIsAnd(pFan0) && (Txs_ObjIsJust(p, pFan0) != 0) {
                pFan0.fMark1 = 1;
            } else if Gia_ObjIsAnd(pFan1) && (Txs_ObjIsJust(p, pFan1) != 0) {
                pFan1.fMark1 = 1;
            } else {
                if pFan0.Value >= pFan1.Value {
                    pFan0.fMark1 = 1;
                } else {
                    pFan1.fMark1 = 1; }
            }
        }
    }
    Vec_IntClear( vPiLits );
    Vec_IntClear( vFfLits );
    Gia_ManForEachObjVec( vCiObjs, p, pObj, i )
    {
        if !pObj.fMark1 {
            continue; }
        if Gia_ObjIsPi(p, pObj) {
            Vec_IntPush( vPiLits, Abc_Var2Lit(Gia_ObjCioId(pObj),                 !pObj.fMark0) );
        } else {
            Vec_IntPush( vFfLits, Abc_Var2Lit(Gia_ObjCioId(pObj)-Gia_ManPiNum(p), !pObj.fMark0) ); }
    }
    assert(Vec_IntSize(vFfLits) > 0, "Vec_IntSize(vFfLits) > 0");
}

/**Function*************************************************************

  Synopsis    [Collects justification path.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs_ManSelectJustPath(p *Gia_Man_t, vNodes *Vec_Int_t, vCoObjs *Vec_Int_t, vRes *Vec_Int_t) {
    var pObj, pFan0, pFan1 *Gia_Obj_t; 
    var i, value0, value1 int32;
    // mark CO drivers
    Gia_ManForEachObjVec( vCoObjs, p, pObj, i )
        Gia_ObjFanin0(pObj).fMark1 = 1;
    // collect just paths
    Vec_IntClear( vRes );
    Gia_ManForEachObjVecReverse( vNodes, p, pObj, i )
    {
        if !pObj.fMark1 {
            continue; }
        pObj.fMark1 = 0;
        Vec_IntPush( vRes, Gia_ObjId(p, pObj) );
        pFan0 = Gia_ObjFanin0(pObj);
        pFan1 = Gia_ObjFanin1(pObj);
        if pObj.fMark0 {
            pFan0.fMark1 = 1;
            pFan1.fMark1 = 1;
            continue;
        }
        value0 = pFan0.fMark0 ^ Gia_ObjFaninC0(pObj);
        value1 = pFan1.fMark0 ^ Gia_ObjFaninC1(pObj);
        assert(!(value0 != 0) || !(value1 != 0), "!value0 || !value1");
        if value0 != 0 {
            pFan1.fMark1 = 1;
        } else if value1 != 0 {
            pFan0.fMark1 = 1;
        } else { // if ( !value0 && !value1 )
        
            pFan0.fMark1 = 1;
            pFan1.fMark1 = 1;
        }
    }
    Vec_IntReverseOrder( vRes );
}
func Txs_ManCollectJustPis(p *Gia_Man_t, vCiObjs *Vec_Int_t, vPiLits *Vec_Int_t) {
    var pObj *Gia_Obj_t; var i int32;
    Vec_IntClear( vPiLits );
    Gia_ManForEachObjVec( vCiObjs, p, pObj, i )
        if ( pObj.fMark1 && Gia_ObjIsPi(p, pObj) )
            Vec_IntPush( vPiLits, Abc_Var2Lit(Gia_ObjCioId(pObj), !pObj.fMark0) );
}
func Txs_ManInitPrio(p *Gia_Man_t, vCiObjs *Vec_Int_t) {
    var pObj *Gia_Obj_t; var i int32;
    Gia_ManConst0(p).Value = 0x7FFFFFFF;
    Gia_ManForEachObjVec( vCiObjs, p, pObj, i )
        pObj.Value = Gia_ObjIsPi(p, pObj) ? 0x7FFFFFFF : Gia_ObjCioId(pObj) - Gia_ManPiNum(p);
}
func Txs_ManPropagatePrio(p *Gia_Man_t, vNodes *Vec_Int_t, vPrio *Vec_Int_t) {
    var pObj, pFan0, pFan1 *Gia_Obj_t;
    var i, value0, value1 int32;
    Gia_ManForEachObjVec( vNodes, p, pObj, i )
    {
        pFan0 = Gia_ObjFanin0(pObj);
        pFan1 = Gia_ObjFanin1(pObj);
        if pObj.fMark0 {
//            pObj->Value = Abc_MinInt( pFan0->Value, pFan1->Value );
            if pFan0.Value == 0x7FFFFFFF {
                pObj.Value = pFan1.Value;
            } else if pFan1.Value == 0x7FFFFFFF {
                pObj.Value = pFan0.Value;
            } else if Vec_IntEntry(vPrio, pFan0.Value) < Vec_IntEntry(vPrio, pFan1.Value) {
                pObj.Value = pFan0.Value;
            } else {
                pObj.Value = pFan1.Value; }
            continue;
        }
        value0 = pFan0.fMark0 ^ Gia_ObjFaninC0(pObj);
        value1 = pFan1.fMark0 ^ Gia_ObjFaninC1(pObj);
        assert(!(value0 != 0) || !(value1 != 0), "!value0 || !value1");
        if value0 != 0 {
            pObj.Value = pFan1.Value;
        } else if value1 != 0 {
            pObj.Value = pFan0.Value;
        } else { // if ( value0 == 0 && value1 == 0 )
        
//            pObj->Value = Abc_MaxInt( pFan0->Value, pFan1->Value );
            if pFan0.Value == 0x7FFFFFFF || pFan1.Value == 0x7FFFFFFF {
                pObj.Value = 0x7FFFFFFF;
            } else if Vec_IntEntry(vPrio, pFan0.Value) >= Vec_IntEntry(vPrio, pFan1.Value) {
                pObj.Value = pFan0.Value;
            } else {
                pObj.Value = pFan1.Value; }
        }
        assert(~pObj.Value, "~pObj.Value");
    }
}
func Txs_ManFindMinId(p *Gia_Man_t, vCoObjs *Vec_Int_t, vPrio *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t; var i int32;
var iMinId int32 = -1;
    Gia_ManForEachObjVec( vCoObjs, p, pObj, i )
        if Gia_ObjFanin0(pObj).Value != 0x7FFFFFFF {
            if iMinId == -1 || Vec_IntEntry(vPrio, iMinId) > Vec_IntEntry(vPrio, Gia_ObjFanin0(pObj).Value) {
                iMinId = Gia_ObjFanin0(pObj).Value; }
        }
    return iMinId;
}
func Txs_ManFindCiReduction(p *Gia_Man_t, vPrio *Vec_Int_t, vCiObjs *Vec_Int_t, vNodes *Vec_Int_t, vCoObjs *Vec_Int_t, vPiLits *Vec_Int_t, vFfLits *Vec_Int_t, vTemp *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var iPrioCi int32;
    // propagate PI influence
    Txs_ManSelectJustPath( p, vNodes, vCoObjs, vTemp );
    Txs_ManCollectJustPis( p, vCiObjs, vPiLits );
//    printf( "%d -> %d  ", Vec_IntSize(vNodes), Vec_IntSize(vTemp) );
    // iteratively detect and remove smallest IDs
    Vec_IntClear( vFfLits );
    Txs_ManInitPrio( p, vCiObjs );
    for {
        Txs_ManPropagatePrio( p, vTemp, vPrio );
        iPrioCi = Txs_ManFindMinId( p, vCoObjs, vPrio );
        if iPrioCi == -1 {
            break; }
        pObj = Gia_ManCi( p, Gia_ManPiNum(p)+iPrioCi );
        Vec_IntPush( vFfLits, Abc_Var2Lit(iPrioCi, !pObj.fMark0) );
        pObj.Value = 0x7FFFFFFF;
    }
}
func Txs_ManPrintFlopLits(vFfLits *Vec_Int_t, vPrio *Vec_Int_t) {
    var i, Entry int32;
    fmt.Printf("%3d : ", Vec_IntSize(vFfLits));
    Vec_IntForEachEntry( vFfLits, Entry, i )
        fmt.Printf("%s%d(%d) ", Abc_LitIsCompl(Entry)? "+":"-", Abc_Lit2Var(Entry), Vec_IntEntry(vPrio, Abc_Lit2Var(Entry)));
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Verify the result.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Txs_ManVerify(p *Gia_Man_t, vCiObjs *Vec_Int_t, vNodes *Vec_Int_t, vPiLits *Vec_Int_t, vFfLits *Vec_Int_t, vCoObjs *Vec_Int_t, vCoVals *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    var i, iLit int32;
    Gia_ObjTerSimSet0( Gia_ManConst0(p) );
    Gia_ManForEachObjVec( vCiObjs, p, pObj, i )
        Gia_ObjTerSimSetX( pObj );
    Vec_IntForEachEntry( vPiLits, iLit, i )
    {
        pObj = Gia_ManPi( p, Abc_Lit2Var(iLit) );
        assert(Gia_ObjTerSimGetX(pObj), "Gia_ObjTerSimGetX(pObj)");
        if Abc_LitIsCompl(iLit) {
            Gia_ObjTerSimSet0( pObj );
        } else {
            Gia_ObjTerSimSet1( pObj ); }
    }
    Vec_IntForEachEntry( vFfLits, iLit, i )
    {
        pObj = Gia_ManCi( p, Gia_ManPiNum(p) + Abc_Lit2Var(iLit) );
        assert(Gia_ObjTerSimGetX(pObj), "Gia_ObjTerSimGetX(pObj)");
        if Abc_LitIsCompl(iLit) {
            Gia_ObjTerSimSet0( pObj );
        } else {
            Gia_ObjTerSimSet1( pObj ); }
    }
    Gia_ManForEachObjVec( vNodes, p, pObj, i )
        Gia_ObjTerSimAnd( pObj );
    Gia_ManForEachObjVec( vCoObjs, p, pObj, i )
    {
        Gia_ObjTerSimCo( pObj );
        if Vec_IntEntry(vCoVals, i) {
            assert(Gia_ObjTerSimGet1(pObj), "Gia_ObjTerSimGet1(pObj)");
        } else {
            assert(Gia_ObjTerSimGet0(pObj), "Gia_ObjTerSimGet0(pObj)"); }
    }
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
func Txs_ManTernarySim(p *Txs_Man_t, k int32, pCube *Pdr_Set_t) *Pdr_Set_t {
    var fTryNew int32  = 1;
    var pRes *Pdr_Set_t;
    var pObj *Gia_Obj_t;
    // collect CO objects
    Vec_IntClear( p.vCoObjs );
    if pCube == nil { // the target is the property output
    
        pObj = Gia_ManCo(p.pGia, p.pMan.iOutCur);
        Vec_IntPush( p.vCoObjs, Gia_ObjId(p.pGia, pObj) );
    
    } else { // the target is the cube
    
        var i int32;
        for  i = 0; i < pCube.nLits; i++  {
            if pCube.Lits[i] == -1 {
                continue; }
            pObj = Gia_ManCo(p.pGia, Gia_ManPoNum(p.pGia) + Abc_Lit2Var(pCube.Lits[i]));
            Vec_IntPush( p.vCoObjs, Gia_ObjId(p.pGia, pObj) );
        }
    }
if false {
Abc_Print( 1, "Trying to justify cube " );
if pCube != nil {
    Pdr_SetPrint( os.Stdout, pCube, Gia_ManRegNum(p.pGia), nil );
} else {
    Abc_Print( 1, "<prop=fail>" ); }
Abc_Print( 1, " in frame %d.\n", k );
}

    // collect CI objects
    Txs_ManCollectCone( p.pGia, p.vCoObjs, p.vCiObjs, p.vNodes );
    // collect values
    Pdr_ManCollectValues( p.pMan, k, p.vCiObjs, p.vCiVals );
    Pdr_ManCollectValues( p.pMan, k, p.vCoObjs, p.vCoVals );

    // perform two passes
    Txs_ManForwardPass( p.pGia, p.vPrio, p.vCiObjs, p.vCiVals, p.vNodes, p.vCoObjs, p.vCoVals );
    if fTryNew != 0 {
        Txs_ManFindCiReduction( p.pGia, p.vPrio, p.vCiObjs, p.vNodes, p.vCoObjs, p.vPiLits, p.vFfLits, p.vTemp );
    } else {
        Txs_ManBackwardPass( p.pGia, p.vCiObjs, p.vNodes, p.vPiLits, p.vFfLits ); }
    Txs_ManVerify( p.pGia, p.vCiObjs, p.vNodes, p.vPiLits, p.vFfLits, p.vCoObjs, p.vCoVals );

    // derive the final set
    pRes = Pdr_SetCreate( p.vFfLits, p.vPiLits );
    //ZH: Disabled assertion because this invariant doesn't hold with down
    //because of the join operation which can bring in initial states
    //assert( k == 0 || !Pdr_SetIsInit(pRes, -1) );
    return pRes;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
