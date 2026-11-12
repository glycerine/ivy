package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaSplit.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Structural AIG splitting.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSplit.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Spl_Man_t struct {
    // input data
    pGia       *Gia_Man_t;       // user AIG with nodes marked
    iObj              int32;       // object ID
    Limit              int32;      // limit on AIG nodes
    fReverse              int32;   // enable reverse search
    // intermediate
    vMarksCIO       *Vec_Bit_t;  // CI/CO marks
    vMarksIn       *Vec_Bit_t;   // input marks
    vMarksNo       *Vec_Bit_t;   // node marks
    vMarksAnd       *Vec_Bit_t;  // AND node marks
    vRoots       *Vec_Int_t;     // nodes pointing to Nodes
    vNodes       *Vec_Int_t;     // nodes used in the window
    vLeaves       *Vec_Int_t;    // nodes pointed by Nodes
    vAnds       *Vec_Int_t;      // AND nodes used in the window
    // temporary 
    vFanouts       *Vec_Int_t;   // fanouts of the node
    vCands       *Vec_Int_t;     // candidate nodes
    vInputs       *Vec_Int_t;    // non-trivial inputs
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Transforming to the internal LUT representation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManToWecMapping(p *Gia_Man_t) *Vec_Wec_t {
    var vMapping *Vec_Wec_t  = Vec_WecStart( Gia_ManObjNum(p) );
    var Obj, Fanin, k int32;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    // inlining of macro Gia_ManForEachLut
    for  Obj = 1; Obj < Gia_ManObjNum(p); Obj++  { if !Gia_ObjIsLut(p, Obj) {} else {
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,Obj) && ((Fanin = Gia_ObjLutFanins(p,Obj)[k]),1); k++  {
            Vec_WecPush( vMapping, Obj, Fanin );
        } }
    }
    return vMapping;
}
func Spl_ManFromWecMapping(p *Gia_Man_t, vMap *Vec_Wec_t) *Vec_Int_t {
    var vMapping, vVec *Vec_Int_t; var i, k, Obj int32;
    assert(Gia_ManHasMapping2(p), "Gia_ManHasMapping2(p)");
    vMapping = Vec_IntAlloc( Gia_ManObjNum(p) + Vec_WecSizeSize(vMap) + 2*Vec_WecSizeUsed(vMap) );
    Vec_IntFill( vMapping, Gia_ManObjNum(p), 0 );
    // inlining of macro Vec_WecForEachLevel
    for  i = 0; (i < Vec_WecSize(vMap)) && (((vVec) = Vec_WecEntry(vMap, i)), 1); i++  {
        if Vec_IntSize(vVec) > 0 {
            Vec_IntWriteEntry( vMapping, i, Vec_IntSize(vMapping) );
            Vec_IntPush( vMapping, Vec_IntSize(vVec) );
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vVec)) && (((Obj) = Vec_IntEntry(vVec, k)), 1); k++  {
                Vec_IntPush( vMapping, Obj );
            }
            Vec_IntPush( vMapping, i );
        }
    }
    assert(Vec_IntSize(vMapping) < 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping), "Vec_IntSize(vMapping) < 16 || Vec_IntSize(vMapping) == Vec_IntCap(vMapping)");
    return vMapping;
}


/**Function*************************************************************

  Synopsis    [Creating manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManAlloc(pGia *Gia_Man_t, Limit int32, fReverse int32) *Spl_Man_t {
    var i, iObj int32;
    // inlining of macro ABC_CALLOC
    var p *Spl_Man_t  = &Spl_Man_t{};
    p.pGia       = pGia;
    p.Limit      = Limit;
    p.fReverse   = fReverse;
    // intermediate
    p.vMarksCIO  = Vec_BitStart( Gia_ManObjNum(pGia) );
    p.vMarksIn   = Vec_BitStart( Gia_ManObjNum(pGia) );
    p.vMarksNo   = Vec_BitStart( Gia_ManObjNum(pGia) );
    p.vMarksAnd  = Vec_BitStart( Gia_ManObjNum(pGia) );
    p.vRoots     = Vec_IntAlloc( 100 );
    p.vNodes     = Vec_IntAlloc( 100 );
    p.vLeaves    = Vec_IntAlloc( 100 );
    p.vAnds      = Vec_IntAlloc( 100 );
    // temporary 
    p.vFanouts   = Vec_IntAlloc( 100 );
    p.vCands     = Vec_IntAlloc( 100 );
    p.vInputs    = Vec_IntAlloc( 100 );
    // mark CIs/COs
    Vec_BitWriteEntry( p.vMarksCIO, 0, 1 );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(pGia.vCis)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCi(pGia, i))); i++  {
        Vec_BitWriteEntry( p.vMarksCIO, iObj, 1 );
    }
    // inlining of macro Gia_ManForEachCoId
    for  i = 0; (i < Vec_IntSize(pGia.vCos)) && ((iObj) = Gia_ObjId(pGia, Gia_ManCo(pGia, i))); i++  {
        Vec_BitWriteEntry( p.vMarksCIO, iObj, 1 );
    }
    // mapping
    // inlining of macro ABC_FREE
    pGia.Refs = nil;
    Gia_ManCreateRefs( pGia );
    Gia_ManSetLutRefs( pGia );
    assert(Gia_ManHasMapping(pGia), "Gia_ManHasMapping(pGia)");
    assert(!Gia_ManHasMapping2(pGia), "!Gia_ManHasMapping2(pGia)");
    pGia.vMapping2 = Spl_ManToWecMapping( pGia );
    Vec_IntFreeP( &pGia.vMapping );
    // fanout
    Gia_ManStaticFanoutStart( pGia );
    return p;
}
func Spl_ManStop(p *Spl_Man_t) {
    // fanout
    Gia_ManStaticFanoutStop( p.pGia );
    // mapping
    assert(!Gia_ManHasMapping(p.pGia), "!Gia_ManHasMapping(p.pGia)");
    assert(Gia_ManHasMapping2(p.pGia), "Gia_ManHasMapping2(p.pGia)");
    p.pGia.vMapping = Spl_ManFromWecMapping( p.pGia, p.pGia.vMapping2 );
    Vec_WecFreeP( &p.pGia.vMapping2 );
    // intermediate
    Vec_BitFree( p.vMarksCIO );
    Vec_BitFree( p.vMarksIn );
    Vec_BitFree( p.vMarksNo );
    Vec_BitFree( p.vMarksAnd );
    Vec_IntFree( p.vRoots );
    Vec_IntFree( p.vNodes );
    Vec_IntFree( p.vLeaves );
    Vec_IntFree( p.vAnds );
    // temporary 
    Vec_IntFree( p.vFanouts );
    Vec_IntFree( p.vCands );
    Vec_IntFree( p.vInputs );
    // inlining of macro ABC_FREE
    p = nil;
}

/**Function*************************************************************

  Synopsis    [Takes Nodes and Marks. Returns Leaves and Roots.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManWinFindLeavesRoots(p *Spl_Man_t) {
    var vVec *Vec_Int_t;
    var iObj, iFan, i, k int32;
    // collect leaves
/*
    Vec_IntClear( p->vLeaves );
    Gia_ManForEachLut2Vec( p->vNodes, p->pGia, vVec, iObj, i )
    {
        assert( Vec_BitEntry(p->vMarksNo, iObj) );
        Vec_IntForEachEntry( vVec, iFan, k )
            if ( !Vec_BitEntry(p->vMarksNo, iFan) )
            {
                Vec_BitWriteEntry(p->vMarksNo, iFan, 1);
                Vec_IntPush( p->vLeaves, iFan );
            }
    }
    Vec_IntForEachEntry( p->vLeaves, iFan, i )
        Vec_BitWriteEntry(p->vMarksNo, iFan, 0);
*/
    Vec_IntClear( p.vLeaves );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && (((iObj) = Vec_IntEntry(p.vAnds, i)), 1); i++  {
        var pObj *Gia_Obj_t  = Gia_ManObj( p.pGia, iObj );
        assert(Vec_BitEntry(p.vMarksAnd, iObj), "Vec_BitEntry(p.vMarksAnd, iObj)");
        iFan = Gia_ObjFaninId0( pObj, iObj );
        if !Vec_BitEntry(p.vMarksAnd, iFan) {
            assert(Gia_ObjIsLut2(p.pGia, iFan) || Vec_BitEntry(p.vMarksCIO, iFan), "Gia_ObjIsLut2(p.pGia, iFan) || Vec_BitEntry(p.vMarksCIO, iFan)");
            Vec_BitWriteEntry(p.vMarksAnd, iFan, 1);
            Vec_IntPush( p.vLeaves, iFan );
        }
        iFan = Gia_ObjFaninId1( pObj, iObj );
        if !Vec_BitEntry(p.vMarksAnd, iFan) {
            assert(Gia_ObjIsLut2(p.pGia, iFan) || Vec_BitEntry(p.vMarksCIO, iFan), "Gia_ObjIsLut2(p.pGia, iFan) || Vec_BitEntry(p.vMarksCIO, iFan)");
            Vec_BitWriteEntry(p.vMarksAnd, iFan, 1);
            Vec_IntPush( p.vLeaves, iFan );
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vLeaves)) && (((iFan) = Vec_IntEntry(p.vLeaves, i)), 1); i++  {
        Vec_BitWriteEntry(p.vMarksAnd, iFan, 0);
    }

    // collect roots
    Vec_IntClear( p.vRoots );
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            Gia_ObjLutRefDecId( p.pGia, iFan );
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && (((iObj) = Vec_IntEntry(p.vAnds, i)), 1); i++  {
        if Gia_ObjLutRefNumId(p.pGia, iObj) {
            Vec_IntPush( p.vRoots, iObj ); }
    }
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            Gia_ObjLutRefIncId( p.pGia, iFan );
        }
    }
}




/**Function*************************************************************

  Synopsis    [Computes LUTs that are fanouts of the node outside of the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManLutFanouts_rec(p *Gia_Man_t, iObj int32, vFanouts *Vec_Int_t, vMarksNo *Vec_Bit_t, vMarksCIO *Vec_Bit_t) {
    var iFanout, i int32;
    if Vec_BitEntry(vMarksNo, iObj) || Vec_BitEntry(vMarksCIO, iObj) {
        return; }
    if Gia_ObjIsLut2(p, iObj) {
        Vec_BitWriteEntry( vMarksCIO, iObj, 1 );
        Vec_IntPush( vFanouts, iObj );
        return;
    }
    // inlining of macro Gia_ObjForEachFanoutStaticId
    for  i = 0; (i < Gia_ObjFanoutNumId(p, iObj))   && (((iFanout) = Gia_ObjFanoutId(p, iObj, i)), 1); i++  {
        Spl_ManLutFanouts_rec( p, iFanout, vFanouts, vMarksNo, vMarksCIO );
    }
}
func Spl_ManLutFanouts(p *Gia_Man_t, iObj int32, vFanouts *Vec_Int_t, vMarksNo *Vec_Bit_t, vMarksCIO *Vec_Bit_t) int32 {
    var i, iFanout int32;
    assert(Gia_ObjIsLut2(p, iObj), "Gia_ObjIsLut2(p, iObj)");
    Vec_IntClear( vFanouts );
    // inlining of macro Gia_ObjForEachFanoutStaticId
    for  i = 0; (i < Gia_ObjFanoutNumId(p, iObj))   && (((iFanout) = Gia_ObjFanoutId(p, iObj, i)), 1); i++  {
        Spl_ManLutFanouts_rec( p, iFanout, vFanouts, vMarksNo, vMarksCIO );
    }
    // clean up
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFanouts)) && (((iFanout) = Vec_IntEntry(vFanouts, i)), 1); i++  {
        Vec_BitWriteEntry( vMarksCIO, iFanout, 0 );
    }
    return Vec_IntSize(vFanouts);
}


/**Function*************************************************************

  Synopsis    [Returns the number of fanins beloning to the set.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManCountMarkedFanins(p *Gia_Man_t, iObj int32, vMarks *Vec_Bit_t) int32 {
    var i, iFan, Count int32;
    var vFanins *Vec_Int_t  = Gia_ObjLutFanins2(p, iObj);
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFanins)) && (((iFan) = Vec_IntEntry(vFanins, i)), 1); i++  {
        if Vec_BitEntry(vMarks, iFan) {
            Count++; }
    }
    return Count;
}
func Spl_ManFindGoodCand(p *Spl_Man_t) int32 {
    var i, iObj int32;
    var Res int32;
    var InCount int32;
    var InCountMax int32 = -1; 
    // mark leaves
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vInputs)) && (((iObj) = Vec_IntEntry(p.vInputs, i)), 1); i++  {
        Vec_BitWriteEntry( p.vMarksIn, iObj, 1 );
    }
    // find candidate with maximum input overlap
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vCands)) && (((iObj) = Vec_IntEntry(p.vCands, i)), 1); i++  {
        InCount = Spl_ManCountMarkedFanins( p.pGia, iObj, p.vMarksIn );
        if InCountMax < InCount {
            InCountMax = InCount;
            Res = iObj;
        }
    }
    // unmark leaves
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vInputs)) && (((iObj) = Vec_IntEntry(p.vInputs, i)), 1); i++  {
        Vec_BitWriteEntry( p.vMarksIn, iObj, 0 );
    }
    return Res;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManFindOne(p *Spl_Man_t) int32 {
    var vVec *Vec_Int_t;
    var nFanouts, iObj, iFan, i, k int32;
    var Res int32; 

    // deref
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            Gia_ObjLutRefDecId( p.pGia, iFan );
        }
    }

    // collect external nodes
    if p.fReverse && (Vec_IntSize(p.vNodes) & 1) {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(p.vNodes)) && (((iObj) = Vec_IntEntry(p.vNodes, i)), 1); i++  {
            if Gia_ObjLutRefNumId(p.pGia, iObj) == 0 {
                continue; }
            assert(Gia_ObjLutRefNumId(p.pGia, iObj) > 0, "Gia_ObjLutRefNumId(p.pGia, iObj) > 0");
            if Gia_ObjLutRefNumId(p.pGia, iObj) >= 5 {  // skip nodes with high fanout!
                continue; }
            nFanouts = Spl_ManLutFanouts( p.pGia, iObj, p.vFanouts, p.vMarksNo, p.vMarksCIO );
            if Gia_ObjLutRefNumId(p.pGia, iObj) == 1 && nFanouts == 1 {
                Res = Vec_IntEntry(p.vFanouts, 0);
                goto finish;
            }
            //Vec_IntAppend( p->vCands, p->vFanouts );
        }
    }

    // consider LUT inputs - get one that has no refs
    Vec_IntClear( p.vCands );
    Vec_IntClear( p.vInputs );
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            if !Vec_BitEntry(p.vMarksNo, iFan) && !Vec_BitEntry(p.vMarksCIO, iFan) && !Gia_ObjLutRefNumId(p.pGia, iFan) {
                Vec_IntPush( p.vCands, iFan );
                Vec_IntPush( p.vInputs, iFan );
            }
        }
    }
    Res = Spl_ManFindGoodCand( p );
    if Res != 0 {
        goto finish; }

    // collect candidates
    Vec_IntClear( p.vCands );
    Vec_IntClear( p.vInputs );
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            if !Vec_BitEntry(p.vMarksNo, iFan) && !Vec_BitEntry(p.vMarksCIO, iFan) {
                Vec_IntPush( p.vCands, iFan );
                Vec_IntPush( p.vInputs, iFan );
            }
        }
    }

    // all inputs have refs - collect external nodes
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vNodes)) && (((iObj) = Vec_IntEntry(p.vNodes, i)), 1); i++  {
        if Gia_ObjLutRefNumId(p.pGia, iObj) == 0 {
            continue; }
        assert(Gia_ObjLutRefNumId(p.pGia, iObj) > 0, "Gia_ObjLutRefNumId(p.pGia, iObj) > 0");
        if Gia_ObjLutRefNumId(p.pGia, iObj) >= 5 {  // skip nodes with high fanout!
            continue; }
        nFanouts = Spl_ManLutFanouts( p.pGia, iObj, p.vFanouts, p.vMarksNo, p.vMarksCIO );
        if Gia_ObjLutRefNumId(p.pGia, iObj) == 1 && nFanouts == 1 {
            Res = Vec_IntEntry(p.vFanouts, 0);
            goto finish;
        }
        Vec_IntAppend( p.vCands, p.vFanouts );
    }

    // choose among not-so-good ones
    Res = Spl_ManFindGoodCand( p );
    if Res != 0 {
        goto finish; }

    // get the first candidate
    if Res == 0 && Vec_IntSize(p.vCands) > 0 {
        Res = Vec_IntEntry( p.vCands, 0 ); }

finish:
    // ref
    // inlining of macro Gia_ManForEachLut2Vec
    for  i = 0; i < Vec_IntSize(p.vNodes) && (vVec = Vec_WecEntry(p.pGia.vMapping2, (iObj = Vec_IntEntry(p.vNodes, i)))); i++  {
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vVec)) && (((iFan) = Vec_IntEntry(vVec, k)), 1); k++  {
            Gia_ObjLutRefIncId( p.pGia, iFan );
        }
    }
    return Res;
}



/**Function*************************************************************

  Synopsis    [Computing window for one pivot node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManLutMffcSize(p *Gia_Man_t, iObj int32, vTemp *Vec_Int_t, vMarksAnd *Vec_Bit_t) int32 {
    var iTemp, i, k int32;
    assert(Gia_ObjIsLut2(p, iObj), "Gia_ObjIsLut2(p, iObj)");
//Vec_IntPrint( Gia_ObjLutFanins2(p, iObj) );
    Gia_ManIncrementTravId( p );
    Gia_ManCollectAnds( p, &iObj, 1, vTemp, Gia_ObjLutFanins2(p, iObj) );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vTemp)) && (((iTemp) = Vec_IntEntry(vTemp, i)), 1); i++  {
        if !Vec_BitEntry(vMarksAnd, iTemp) {
            Vec_IntWriteEntry( vTemp, k++, iTemp ); }
    }
    Vec_IntShrink( vTemp, k );
    return k;
}
func Spl_ManAddNode(p *Spl_Man_t, iPivot int32, vCands *Vec_Int_t) {
    var i, iObj int32;
    Vec_IntPush( p.vNodes, iPivot );
    Vec_BitWriteEntry( p.vMarksNo, iPivot, 1 );
    Vec_IntAppend( p.vAnds, vCands );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vCands)) && (((iObj) = Vec_IntEntry(vCands, i)), 1); i++  {
        Vec_BitWriteEntry( p.vMarksAnd, iObj, 1 );
    }
}
func Spl_ManComputeOne(p *Spl_Man_t, iPivot int32) int32 {
    var CountAdd, iObj, i int32;
    assert(Gia_ObjIsLut2(p.pGia, iPivot), "Gia_ObjIsLut2(p.pGia, iPivot)");
//Gia_ManPrintCone2( p->pGia, Gia_ManObj(p->pGia, iPivot) );
    // assume it was previously filled in
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vNodes)) && (((iObj) = Vec_IntEntry(p.vNodes, i)), 1); i++  {
        Vec_BitWriteEntry( p.vMarksNo, iObj, 0 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(p.vAnds)) && (((iObj) = Vec_IntEntry(p.vAnds, i)), 1); i++  {
        Vec_BitWriteEntry( p.vMarksAnd, iObj, 0 );
    }
    // double check that it is empty
    //Gia_ManForEachLut2( p->pGia, iObj )
    //    assert( !Vec_BitEntry(p->vMarksNo, iObj) );
    //Gia_ManForEachLut2( p->pGia, iObj )
    //    assert( !Vec_BitEntry(p->vMarksAnd, iObj) );
    // clean arrays
    Vec_IntClear( p.vNodes );
    Vec_IntClear( p.vAnds );
    // add root node
    Spl_ManLutMffcSize( p.pGia, iPivot, p.vCands, p.vMarksAnd );
    Spl_ManAddNode( p, iPivot, p.vCands );
    if Vec_IntSize(p.vAnds) > p.Limit {
        return 0; }
    //printf( "%d ", iPivot );
    // add one node at a time
    for  iObj = Spl_ManFindOne(p)  {
        assert(Gia_ObjIsLut2(p.pGia, iObj), "Gia_ObjIsLut2(p.pGia, iObj)");
        assert(!Vec_BitEntry(p.vMarksNo, iObj), "!Vec_BitEntry(p.vMarksNo, iObj)");
        CountAdd = Spl_ManLutMffcSize( p.pGia, iObj, p.vCands, p.vMarksAnd );
        if Vec_IntSize(p.vAnds) + CountAdd > p.Limit {
            break; }
        Spl_ManAddNode( p, iObj, p.vCands );
        //printf( "+%d ", iObj );
    }
    //printf( "\n" );
    Vec_IntSort( p.vNodes, 0 );
    Vec_IntSort( p.vAnds, 0 );
    Spl_ManWinFindLeavesRoots( p );
    Vec_IntSort( p.vLeaves, 0 );
    Vec_IntSort( p.vRoots, 0 );
    return 1;
}

/**Function*************************************************************

  Synopsis    [External procedures.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManComputeOneWin(pGia *Gia_Man_t, iPivot int32, pvRoots **Vec_Int_t, pvNodes **Vec_Int_t, pvLeaves **Vec_Int_t, pvAnds **Vec_Int_t) int32 {
    var p *Spl_Man_t  = (Spl_Man_t *)pGia.pSatlutWinman;
    assert(p != nil, "p != nil");
    if iPivot == -1 {
        Spl_ManStop( p );
        pGia.pSatlutWinman = nil;
        return 0;
    }
    if !(Spl_ManComputeOne( p, iPivot ) != 0) {
        *pvRoots  = nil;
        *pvNodes  = nil;
        *pvLeaves = nil;
        *pvAnds   = nil;
        return 0;
    }
    *pvRoots  = p.vRoots;
    *pvNodes  = p.vNodes;
    *pvLeaves = p.vLeaves;
    *pvAnds   = p.vAnds;
    // Vec_IntPrint( p->vNodes );
    return Vec_IntSize(p.vAnds);
}
func Gia_ManComputeOneWinStart(pGia *Gia_Man_t, nAnds int32, fReverse int32) {
    assert(pGia.pSatlutWinman == nil, "pGia.pSatlutWinman == nil");
    pGia.pSatlutWinman = Spl_ManAlloc( pGia, nAnds, fReverse );
}

/**Function*************************************************************

  Synopsis    [Testing procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Spl_ManComputeOneTest(pGia *Gia_Man_t) {
    var iLut, Count int32;
    Gia_ManComputeOneWinStart( pGia, 64, 0 );
    // inlining of macro Gia_ManForEachLut2
    for  iLut = 1; iLut < Gia_ManObjNum(pGia); iLut++  { if !Gia_ObjIsLut2(pGia, iLut) {} else {
        var vRoots, vNodes, vLeaves, vAnds *Vec_Int_t;
        Count = Gia_ManComputeOneWin( pGia, iLut, &vRoots, &vNodes, &vLeaves, &vAnds );
        fmt.Printf("Obj = %6d : Leaf = %2d.  Node = %2d.  Root = %2d.    AND = %3d.\n", iLut, Vec_IntSize(vLeaves), Vec_IntSize(vNodes), Vec_IntSize(vRoots), Count); 
    }
    }
    Gia_ManComputeOneWin( pGia, -1, nil, nil, nil, nil );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

