package cec

import "fmt"

/**CFile****************************************************************

  FileName    [cecSimBack.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Combinational equivalence checking.]

  Synopsis    [Backward simulation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cecSimBack.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "cecInt.h"
//#include "aig/gia/giaAig.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ObjSatVerify(p *Gia_Man_t, pRoot *Gia_Obj_t, Value int32) int32 {
    var pObj *Gia_Obj_t;
    var i int32;
    var RetValue int32 = 1;
    fmt.Printf("Obj = %4d  Value = %d  ", Gia_ObjId(p, pRoot), Value);
    Gia_ObjTerSimSet0( Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if !Gia_ObjIsTravIdCurrent(p, pObj) {
            Gia_ObjTerSimSetX( pObj );
fmt.Printf("x");
        } else if pObj.fMark0 {
            Gia_ObjTerSimSet1( pObj );
fmt.Printf("1");
        } else {
            Gia_ObjTerSimSet0( pObj );
fmt.Printf("0"); }
    }
    fmt.Printf(" ");
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Gia_ObjTerSimAnd( pObj ); }
    }
    if Value ? Gia_ObjTerSimGet1(pRoot) : Gia_ObjTerSimGet0(pRoot) {
        fmt.Printf("Verification successful.\n");
    } else {
        fmt.Printf("Verification failed.\n");
RetValue = 0; }
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Return 1 if pObj can have Value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCheckSat2_rec(p *Gia_Man_t, pObj *Gia_Obj_t, Value word, vValues *Vec_Wrd_t) word {
    var pFan0, pFan1 *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return Value == int32(pObj.fMark0); }
    Gia_ObjSetTravIdCurrent(p, pObj);
    pObj.fMark0 = Value;
    if !Gia_ObjIsAnd(pObj) {
        return 1; }
    pFan0 = Gia_ObjFanin0(pObj);
    pFan1 = Gia_ObjFanin1(pObj);
    if Value != 0 {
        return Cec_ManCheckSat2_rec( p, pFan0, !Gia_ObjFaninC0(pObj), vValues ) &&
               Cec_ManCheckSat2_rec( p, pFan1, !Gia_ObjFaninC1(pObj), vValues );
    }
    if Gia_ObjIsTravIdCurrent(p, pFan0) { // already assigned
    
        if Gia_ObjFaninC0(pObj) == int32(pFan0.fMark0) { // justified
            return 1; }
        return Cec_ManCheckSat2_rec( p, pFan1, Gia_ObjFaninC1(pObj), vValues );        
    }
    if Gia_ObjIsTravIdCurrent(p, pFan1) { // already assigned
    
        if Gia_ObjFaninC1(pObj) == int32(pFan1.fMark0) { // justified
            return 1; }
        return Cec_ManCheckSat2_rec( p, pFan0, Gia_ObjFaninC0(pObj), vValues );        
    }
    return Cec_ManCheckSat2_rec( p, pFan0, Gia_ObjFaninC0(pObj), vValues );
}
func Cec_ManCheckSat2(p *Gia_Man_t, pObj *Gia_Obj_t, Value int32, vValues *Vec_Wrd_t) word {
    return 0;
}

/**Function*************************************************************

  Synopsis    [Return 1 if pObj can have Value.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Cec_ManCheckSat_rec(p *Gia_Man_t, pObj *Gia_Obj_t, Value int32) int32 {
    var pFan0, pFan1 *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return Value == int32(pObj.fMark0); }
    Gia_ObjSetTravIdCurrent(p, pObj);
    pObj.fMark0 = Value;
    if !Gia_ObjIsAnd(pObj) {
        return 1; }
    pFan0 = Gia_ObjFanin0(pObj);
    pFan1 = Gia_ObjFanin1(pObj);
    if Value != 0 {
        return Cec_ManCheckSat_rec( p, pFan0, !Gia_ObjFaninC0(pObj) ) &&
               Cec_ManCheckSat_rec( p, pFan1, !Gia_ObjFaninC1(pObj) );
    }
    if Gia_ObjIsTravIdCurrent(p, pFan0) { // already assigned
    
        if Gia_ObjFaninC0(pObj) == int32(pFan0.fMark0) { // justified
            return 1; }
        return Cec_ManCheckSat_rec( p, pFan1, Gia_ObjFaninC1(pObj) );        
    }
    if Gia_ObjIsTravIdCurrent(p, pFan1) { // already assigned
    
        if Gia_ObjFaninC1(pObj) == int32(pFan1.fMark0) { // justified
            return 1; }
        return Cec_ManCheckSat_rec( p, pFan0, Gia_ObjFaninC0(pObj) );        
    }
    return Cec_ManCheckSat_rec( p, pFan0, Gia_ObjFaninC0(pObj) );
}
func Cec_ManSimBack(p *Gia_Man_t) {
    var clk abctime  = Abc_Clock();
    var vValues *Vec_Wrd_t  = Vec_WrdStart( Gia_ManObjNum(p) );
    var pObj *Gia_Obj_t; 
    var i, Count int32;
    var Res word;
    Gia_ManSetPhase( p );
    //Gia_ManForEachAnd( p, pObj, i )
    //    printf( "%d", pObj->fPhase );
    //printf( "\n" );
    //return;

    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        Gia_ManIncrementTravId(p);
        //Gia_ManCleanMark0( p );
        Res = Cec_ManCheckSat_rec( p, pObj, !pObj.fPhase );
        //if ( Res )
        //    Cec_ObjSatVerify( p, pObj, !pObj->fPhase );

        //Res = Cec_ManCheckSat2_rec( p, pObj, !pObj->fPhase ? ~(word)0 : 0, vValues );
        //if ( Res )
        //    Cec_ObjSatVerify2( p, pObj, !pObj->fPhase, Res );

        Count += int32((Res > 0));
    }
    }
    Vec_WrdFree( vValues );
    fmt.Printf("Obj = %6d.  SAT = %6d.  ", Gia_ManAndNum(p), Count);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

