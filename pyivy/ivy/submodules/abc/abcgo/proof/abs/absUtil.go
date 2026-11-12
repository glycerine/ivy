package abs

/**CFile****************************************************************

  FileName    [absUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Abstraction package.]

  Synopsis    [Interface to pthreads.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: absUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 
//#include "abs.h"

//ABC_NAMESPACE_IMPL_START 

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [This procedure sets default parameters.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abs_ParSetDefaults(p *Abs_Par_t) {
    memset( p, 0, sizeof(Abs_Par_t) );
    p.nFramesMax         =      0;   // maximum frames
    p.nFramesStart       =      0;   // starting frame 
    p.nFramesPast        =      4;   // overlap frames
    p.nConfLimit         =      0;   // conflict limit
    p.nLearnedMax        =   1000;   // max number of learned clauses
    p.nLearnedStart      =   1000;   // max number of learned clauses
    p.nLearnedDelta      =    200;   // max number of learned clauses
    p.nLearnedPerce      =     70;   // max number of learned clauses
    p.nTimeOut           =      0;   // timeout in seconds
    p.nRatioMin          =      0;   // stop when less than this % of object is abstracted
    p.nRatioMax          =     30;   // restart when more than this % of object is abstracted
    p.fUseTermVars       =      0;   // use terminal variables
    p.fUseRollback       =      0;   // use rollback to the starting number of frames
    p.fPropFanout        =      1;   // propagate fanouts during refinement
    p.fVerbose           =      0;   // verbose flag
    p.iFrame             =     -1;   // the number of frames covered 
    p.iFrameProved       =     -1;   // the number of frames proved
    p.nFramesNoChangeLim =      2;   // the number of frames without change to dump abstraction
}

/**Function*************************************************************

  Synopsis    [Converting VTA vector to GLA vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaConvertToGla(p *Gia_Man_t, vVta *Vec_Int_t) *Vec_Int_t {
    var pObj *Gia_Obj_t;
    var vGla *Vec_Int_t;
    var nObjMask int32;
    var nObjs int32 = Gia_ManObjNum(p);
    var i int32;
    var Entry int32;
    var nFrames int32 = Vec_IntEntry( vVta, 0 );
    assert(Vec_IntEntry(vVta, nFrames+1) == Vec_IntSize(vVta), "Vec_IntEntry(vVta, nFrames+1) == Vec_IntSize(vVta)");
    // get the bitmask
    nObjMask = (1 << Abc_Base2Log(nObjs)) - 1;
    assert(nObjs <= nObjMask, "nObjs <= nObjMask");
    // go through objects
    vGla = Vec_IntStart( nObjs );
    // inlining of macro Vec_IntForEachEntryStart
    for  i = nFrames+2; (i < Vec_IntSize(vVta)) && (((Entry) = Vec_IntEntry(vVta, i)), 1); i++  {
        pObj = Gia_ManObj( p, (Entry &  nObjMask) );
        assert(Gia_ObjIsRo(p, pObj) || Gia_ObjIsAnd(pObj) || Gia_ObjIsConst0(pObj), "Gia_ObjIsRo(p, pObj) || Gia_ObjIsAnd(pObj) || Gia_ObjIsConst0(pObj)");
        Vec_IntAddToEntry( vGla, (Entry &  nObjMask), 1 );
    }
    Vec_IntWriteEntry( vGla, 0, nFrames );
    return vGla;
}

/**Function*************************************************************

  Synopsis    [Converting GLA vector to VTA vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_VtaConvertFromGla(p *Gia_Man_t, vGla *Vec_Int_t, nFrames int32) *Vec_Int_t {
    var vVta *Vec_Int_t;
    var nObjBits int32;
    var nObjMask int32;
    var nObjs int32 = Gia_ManObjNum(p);
    var i, k, j, Entry, Counter, nGlaSize int32;
    //. get the GLA size
    nGlaSize = Vec_IntSum(vGla);
    // get the bitmask
    nObjBits = Abc_Base2Log(nObjs);
    nObjMask = (1 << Abc_Base2Log(nObjs)) - 1;
    assert(nObjs <= nObjMask, "nObjs <= nObjMask");
    // go through objects
    vVta = Vec_IntAlloc( 1000 );
    Vec_IntPush( vVta, nFrames );
    Counter = nFrames + 2;
    for  i = 0; i <= nFrames; i++, Counter += i * nGlaSize  {
        Vec_IntPush( vVta, Counter );
    }
    for  i = 0; i < nFrames; i++  {
        for  k = 0; k <= i; k++  {
            // inlining of macro Vec_IntForEachEntry
            for  j = 0; (j < Vec_IntSize(vGla)) && (((Entry) = Vec_IntEntry(vGla, j)), 1); j++  {
                if Entry != 0 { 
                    Vec_IntPush( vVta, (k << nObjBits) | j ); }
            }
        }
    }
    Counter = Vec_IntEntry(vVta, nFrames+1);
    assert(Vec_IntEntry(vVta, nFrames+1) == Vec_IntSize(vVta), "Vec_IntEntry(vVta, nFrames+1) == Vec_IntSize(vVta)");
    return vVta;
}

/**Function*************************************************************

  Synopsis    [Converting GLA vector to FLA vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_FlaConvertToGla_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vGla *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    Vec_IntWriteEntry( vGla, Gia_ObjId(p, pObj), 1 );
    if Gia_ObjIsRo(p, pObj) {
        return; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_FlaConvertToGla_rec( p, Gia_ObjFanin0(pObj), vGla );
    Gia_FlaConvertToGla_rec( p, Gia_ObjFanin1(pObj), vGla );
}

/**Function*************************************************************

  Synopsis    [Converting FLA vector to GLA vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_FlaConvertToGla(p *Gia_Man_t, vFla *Vec_Int_t) *Vec_Int_t {
    var vGla *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    // mark const0 and relevant CI objects
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p));
    // inlining of macro Gia_ManForEachPi
    for  i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        Gia_ObjSetTravIdCurrent(p, pObj);
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if !Vec_IntEntry(vFla, i) {
            Gia_ObjSetTravIdCurrent(p, pObj); }
    }
    // label all objects reachable from the PO and selected flops
    vGla = Vec_IntStart( Gia_ManObjNum(p) );
    Vec_IntWriteEntry( vGla, 0, 1 );
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_FlaConvertToGla_rec( p, Gia_ObjFanin0(pObj), vGla );
    }
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++  {
        if Vec_IntEntry(vFla, i) {
            Gia_FlaConvertToGla_rec( p, Gia_ObjFanin0(pObj), vGla ); }
    }
    return vGla;
}

/**Function*************************************************************

  Synopsis    [Converting GLA vector to FLA vector.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaConvertToFla(p *Gia_Man_t, vGla *Vec_Int_t) *Vec_Int_t {
    var vFla *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    vFla = Vec_IntStart( Gia_ManRegNum(p) );
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if Vec_IntEntry(vGla, Gia_ObjId(p, pObj)) {
            Vec_IntWriteEntry( vFla, i, 1 ); }
    }
    return vFla;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_GlaCountFlops(p *Gia_Man_t, vGla *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Count int32;
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++  {
        if Vec_IntEntry(vGla, Gia_ObjId(p, pObj)) {
            Count++; }
    }
    return Count;
}
func Gia_GlaCountNodes(p *Gia_Man_t, vGla *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t;
    var i, Count int32;
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else
        if Vec_IntEntry(vGla, Gia_ObjId(p, pObj)) {
            Count++; }
    }
    return Count;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

