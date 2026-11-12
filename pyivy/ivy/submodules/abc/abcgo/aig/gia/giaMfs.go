package gia

/**CFile****************************************************************

  FileName    [giaMfs.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Interface with the MFS package.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMfs.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "opt/sfm/sfm.h"
//#include "misc/tim/tim.h"
//#include "misc/util/utilTruth.h"

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
func Gia_ManExtractMfs(p *Gia_Man_t) *Sfm_Ntk_t {
    var uTruth word;
    var pTruth *word;
    var uTruths6 [6]word = [6]word{
        ABC_CONST(0xAAAAAAAAAAAAAAAA),
        ABC_CONST(0xCCCCCCCCCCCCCCCC),
        ABC_CONST(0xF0F0F0F0F0F0F0F0),
        ABC_CONST(0xFF00FF00FF00FF00),
        ABC_CONST(0xFFFF0000FFFF0000),
        ABC_CONST(0xFFFFFFFF00000000),
    };
    var pObj, pObjExtra *Gia_Obj_t;
    var vFanins *Vec_Wec_t; // mfs data
    var vFixed *Vec_Str_t;  // mfs data 
    var vEmpty *Vec_Str_t;  // mfs data
    var vTruths *Vec_Wrd_t; // mfs data
    var vArray *Vec_Int_t;
    var vStarts *Vec_Int_t;
    var vTruths2 *Vec_Wrd_t;
    var vLeaves *Vec_Int_t;
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    var nBoxes int32 = Gia_ManBoxNum(p);
    var nVars int32;
    var nRealPis int32;
    if nBoxes != 0 {
        nRealPis = Tim_ManPiNum(pManTime);
    } else {
        nRealPis = Gia_ManPiNum(p);
    }
    var nRealPos int32;
    if nBoxes != 0 {
        nRealPos = Tim_ManPoNum(pManTime);
    } else {
        nRealPos = Gia_ManPoNum(p);
    }
    var i, j, k, curCi, curCo, nBoxIns, nBoxOuts int32;
    var Id, iFan, nMfsVars, nBbIns, nBbOuts, Counter int32;
    var nLutSizeMax int32  = Gia_ManLutSizeMax( p );
    nLutSizeMax = Abc_MaxInt( nLutSizeMax, 6 );
    assert(nLutSizeMax < 16, "nLutSizeMax < 16");
    //assert( !p->pAigExtra || Gia_ManPiNum(p->pAigExtra) <= 6 );
    if pManTime != nil { Tim_ManBlackBoxIoNum( pManTime, &nBbIns, &nBbOuts ); }
    // skip PIs due to box outputs
    Counter += nBbOuts;
    // prepare storage
    nMfsVars = Gia_ManCiNum(p) + 1 + Gia_ManLutNum(p) + Gia_ManCoNum(p) + nBbIns + nBbOuts;
    vFanins  = Vec_WecStart( nMfsVars );
    vFixed   = Vec_StrStart( nMfsVars );
    vEmpty   = Vec_StrStart( nMfsVars );
    vTruths  = Vec_WrdStart( nMfsVars );
    vStarts  = Vec_IntStart( nMfsVars );
    vTruths2 = Vec_WrdAlloc( 10000 );
    // set internal PIs
    Gia_ManCleanCopyArray( p );
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        Gia_ObjSetCopyArray( p, Id, Counter++ );
    }
    // set constant node
    Vec_StrWriteEntry( vFixed, Counter, byte(1) );
    Vec_WrdWriteEntry( vTruths, Counter, (word)0 );
    Gia_ObjSetCopyArray( p, 0, Counter++ );
    // set internal LUTs
    vLeaves = Vec_IntAlloc( nLutSizeMax );
    Gia_ObjComputeTruthTableStart( p, nLutSizeMax );
    // inlining of macro Gia_ManForEachLut
    for  Id = 1; Id < Gia_ManObjNum(p); Id++  { if !Gia_ObjIsLut(p, Id) {} else {
        Vec_IntClear( vLeaves );
        vArray = Vec_WecEntry( vFanins, Counter );
        Vec_IntGrow( vArray, Gia_ObjLutSize(p, Id) );
        // inlining of macro Gia_LutForEachFanin
        for  k = 0; k < Gia_ObjLutSize(p,Id) && ((iFan = Gia_ObjLutFanins(p,Id)[k]),1); k++  {
            assert(Gia_ObjCopyArray(p, iFan) >= 0, "Gia_ObjCopyArray(p, iFan) >= 0");
            Vec_IntPush( vArray, Gia_ObjCopyArray(p, iFan) );
            Vec_IntPush( vLeaves, iFan );
        }
        assert(Vec_IntSize(vLeaves) < 16, "Vec_IntSize(vLeaves) < 16");
        assert(Vec_IntSize(vLeaves) == Gia_ObjLutSize(p, Id), "Vec_IntSize(vLeaves) == Gia_ObjLutSize(p, Id)");
//        uTruth = *Gia_ObjComputeTruthTableCut( p, Gia_ManObj(p, Id), vLeaves );
//        nVars = Abc_Tt6MinBase( &uTruth, Vec_IntArray(vArray), Vec_IntSize(vArray) );
        pTruth = Gia_ObjComputeTruthTableCut( p, Gia_ManObj(p, Id), vLeaves );
        nVars = Abc_TtMinBase( pTruth, Vec_IntArray(vArray), Vec_IntSize(vArray), Vec_IntSize(vLeaves) );
        Vec_IntShrink( vArray, nVars );
        if nVars <= 6 {
            Vec_WrdWriteEntry( vTruths, Counter, pTruth[0] );
        } else {
            var w int32;
            var nWords int32 = Abc_Truth6WordNum( nVars );
            Vec_IntWriteEntry( vStarts, Counter, Vec_WrdSize(vTruths2) );
            for  w = 0; w < nWords; w++  {
                Vec_WrdPush( vTruths2, pTruth[w] );
            }
        }
        if Gia_ObjLutIsMux(p, Id) {
            Vec_StrWriteEntry( vFixed, Counter, byte(1) );
            Vec_StrWriteEntry( vEmpty, Counter, byte(1) );
        }
        Gia_ObjSetCopyArray( p, Id, Counter++ );
    }
    }
    Gia_ObjComputeTruthTableStop( p );
    // set all POs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        iFan = Gia_ObjFaninId0p( p, pObj );
        assert(Gia_ObjCopyArray(p, iFan) >= 0, "Gia_ObjCopyArray(p, iFan) >= 0");
        vArray = Vec_WecEntry( vFanins, Counter );
        Vec_IntFill( vArray, 1, Gia_ObjCopyArray(p, iFan) );
        if i < Gia_ManCoNum(p) - nRealPos { // internal PO
        
            Vec_StrWriteEntry( vFixed, Counter, byte(1) );
            Vec_StrWriteEntry( vEmpty, Counter, byte(1) );
            if Gia_ObjFaninC0(pObj) {
                uTruth = ~uTruths6[0];
            } else {
                uTruth = uTruths6[0];
            }
            Vec_WrdWriteEntry( vTruths, Counter, uTruth );
        }
        Gia_ObjSetCopyArray( p, Gia_ObjId(p, pObj), Counter++ );
    }
    // skip POs due to box inputs
    Counter += nBbIns;
    assert(Counter == nMfsVars, "Counter == nMfsVars");
    // add functions of the boxes
    if p.pAigExtra {
        var iBbIn, iBbOut int32;
        assert(Gia_ManCiNum(p.pAigExtra) < 16, "Gia_ManCiNum(p.pAigExtra) < 16");
        Gia_ObjComputeTruthTableStart( p.pAigExtra, Gia_ManCiNum(p.pAigExtra) );
        curCi = nRealPis;
        curCo = 0;
        for  k = 0; k < nBoxes; k++  {
            nBoxIns = Tim_ManBoxInputNum( pManTime, k );
            nBoxOuts = Tim_ManBoxOutputNum( pManTime, k );
            // iterate through box outputs
            if !Tim_ManBoxIsBlack(pManTime, k) { //&& Tim_ManBoxInputNum(pManTime, k) <= 6 )
            
                // collect truth table leaves
                Vec_IntClear( vLeaves );
                for  i = 0; i < nBoxIns; i++  {
                    Vec_IntPush( vLeaves, Gia_ObjId(p.pAigExtra, Gia_ManCi(p.pAigExtra, i)) );
                }
                for  j = 0; j < nBoxOuts; j++  {
                    // CI corresponding to the box outputs
                    pObj = Gia_ManCi( p, curCi + j );
                    Counter = Gia_ObjCopyArray( p, Gia_ObjId(p, pObj) );
                    // add box inputs (POs of the AIG) as fanins
                    vArray = Vec_WecEntry( vFanins, Counter );
                    Vec_IntGrow( vArray, nBoxIns );
                    for  i = 0; i < nBoxIns; i++  {
                        iFan = Gia_ObjId( p, Gia_ManCo(p, curCo + i) );
                        assert(Gia_ObjCopyArray(p, iFan) >= 0, "Gia_ObjCopyArray(p, iFan) >= 0");
                        Vec_IntPush( vArray, Gia_ObjCopyArray(p, iFan) );
                    }
                    Vec_StrWriteEntry( vFixed, Counter, byte(1) );
                    // box output in the extra manager
                    pObjExtra = Gia_ManCo( p.pAigExtra, curCi - nRealPis + j );
                    // compute truth table
                    pTruth = nil;
                    if Gia_ObjFaninId0p(p.pAigExtra, pObjExtra) == 0 {
                        uTruth = 0;
                        if Gia_ObjFaninC0(pObjExtra) {
                            uTruth = ~uTruth;
                        } else {
                            uTruth = uTruth;
                        }
                        pTruth = &uTruth;
                    
                    } else if Gia_ObjIsCi(Gia_ObjFanin0(pObjExtra)) {
                        uTruth = uTruths6[Gia_ObjCioId(Gia_ObjFanin0(pObjExtra))];
                        if Gia_ObjFaninC0(pObjExtra) {
                            uTruth = ~uTruth;
                        } else {
                            uTruth = uTruth;
                        }
                        pTruth = &uTruth;
                    
                    } else {
                        pTruth = Gia_ObjComputeTruthTableCut( p.pAigExtra, Gia_ObjFanin0(pObjExtra), vLeaves );
                        if Gia_ObjFaninC0(pObjExtra) {
                            Abc_TtNot( pTruth, Abc_Truth6WordNum(Vec_IntSize(vLeaves)) ); }
                    }
                    //uTruth = Gia_ObjFaninC0(pObjExtra) ? ~uTruth : uTruth;
                    //Dau_DsdPrintFromTruth( &uTruth, Vec_IntSize(vArray) );
                    //nVars = Abc_Tt6MinBase( &uTruth, Vec_IntArray(vArray), Vec_IntSize(vArray) );
                    nVars = Abc_TtMinBase( pTruth, Vec_IntArray(vArray), Vec_IntSize(vArray), Vec_IntSize(vLeaves) );
                    Vec_IntShrink( vArray, nVars );
                    if nVars <= 6 {
                        Vec_WrdWriteEntry( vTruths, Counter, pTruth[0] );
                    } else {
                        var w int32;
                        var nWords int32 = Abc_Truth6WordNum( nVars );
                        Vec_IntWriteEntry( vStarts, Counter, Vec_WrdSize(vTruths2) );
                        for  w = 0; w < nWords; w++  {
                            Vec_WrdPush( vTruths2, pTruth[w] );
                        }
                    }
                }
            
            } else { // create buffers for black box inputs and outputs
            
                for  j = 0; j < nBoxOuts; j++  {
                    // CI corresponding to the box outputs
                    pObj = Gia_ManCi( p, curCi + j );
                    Counter = Gia_ObjCopyArray( p, Gia_ObjId(p, pObj) );
                    // connect it with the special primary input (iBbOut)
                    vArray = Vec_WecEntry( vFanins, Counter );
                    assert(Vec_IntSize(vArray) == 0, "Vec_IntSize(vArray) == 0");
                    Vec_IntFill( vArray, 1, iBbOut++ );
                    Vec_StrWriteEntry( vFixed, Counter, byte(1) );
                    Vec_StrWriteEntry( vEmpty, Counter, byte(1) );
                    Vec_WrdWriteEntry( vTruths, Counter, uTruths6[0] );
                }
                for  i = 0; i < nBoxIns; i++  {
                    // CO corresponding to the box inputs
                    pObj = Gia_ManCo( p, curCo + i );
                    Counter = Gia_ObjCopyArray( p, Gia_ObjId(p, pObj) );
                    // connect it with the special primary output (iBbIn)
                    vArray = Vec_WecEntry( vFanins, nMfsVars - nBbIns + iBbIn++ );
                    assert(Vec_IntSize(vArray) == 0, "Vec_IntSize(vArray) == 0");
                    Vec_IntFill( vArray, 1, Counter );
                }
            }
            // set internal POs pointing directly to internal PIs as no-delay
            for  i = 0; i < nBoxIns; i++  {
                pObj = Gia_ManCo( p, curCo + i );
                if !Gia_ObjIsCi( Gia_ObjFanin0(pObj) ) {
                    continue; }
                Counter = Gia_ObjCopyArray( p, Gia_ObjFaninId0p(p, pObj) );
                Vec_StrWriteEntry( vEmpty, Counter, byte(1) );
            }
            curCo += nBoxIns;
            curCi += nBoxOuts;
        }
        curCo += nRealPos;
        Gia_ObjComputeTruthTableStop( p.pAigExtra );
        // verify counts
        assert(curCi == Gia_ManCiNum(p), "curCi == Gia_ManCiNum(p)");
        assert(curCo == Gia_ManCoNum(p), "curCo == Gia_ManCoNum(p)");
        assert(curCi - nRealPis == Gia_ManCoNum(p.pAigExtra), "curCi - nRealPis == Gia_ManCoNum(p.pAigExtra)");
        assert(iBbIn  == nBbIns, "iBbIn  == nBbIns");
        assert(iBbOut == nBbOuts, "iBbOut == nBbOuts");
    }
    // finalize 
    Vec_IntFree( vLeaves );
    return Sfm_NtkConstruct( vFanins, nBbOuts + nRealPis, nRealPos + nBbIns, vFixed, vEmpty, vTruths, vStarts, vTruths2 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManInsertMfs(p *Gia_Man_t, pNtk *Sfm_Ntk_t, fAllBoxes int32) *Gia_Man_t {
    extern int Gia_ManFromIfLogicCreateLut( Gia_Man_t * pNew, word * pRes, Vec_Int_t * vLeaves, Vec_Int_t * vCover, Vec_Int_t * vMapping, Vec_Int_t * vMapping2 );
    var pNew *Gia_Man_t;  var pObj *Gia_Obj_t;
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    var nBoxes int32    = Gia_ManBoxNum(p);
    var nRealPis int32;
    if nBoxes != 0 {
        nRealPis = Tim_ManPiNum(pManTime);
    } else {
        nRealPis = Gia_ManPiNum(p);
    }
    var nRealPos int32;
    if nBoxes != 0 {
        nRealPos = Tim_ManPoNum(pManTime);
    } else {
        nRealPos = Gia_ManPoNum(p);
    }
    var i, k, Id, curCi, curCo, nBoxIns, nBoxOuts, iLitNew, iMfsId, iGroup, Fanin int32;
    var nMfsNodes int32;
    var pTruth *word;
    var uTruthVar word = ABC_CONST(0xAAAAAAAAAAAAAAAA);
    var vGroups *Vec_Wec_t  = Vec_WecStart( nBoxes );
    var vMfs2Gia, vMfs2Old *Vec_Int_t;
    var vGroupMap *Vec_Int_t;
    var vMfsTopo, vCover, vBoxesLeft *Vec_Int_t;
    var vArray, vLeaves *Vec_Int_t;
    var vMapping, vMapping2 *Vec_Int_t;
    var nBbIns, nBbOuts int32;
    if pManTime != nil { Tim_ManBlackBoxIoNum( pManTime, &nBbIns, &nBbOuts ); }
    nMfsNodes = 1 + Gia_ManCiNum(p) + Gia_ManLutNum(p) + Gia_ManCoNum(p) + nBbIns + nBbOuts;
    vMfs2Gia  = Vec_IntStartFull( nMfsNodes );
    vMfs2Old  = Vec_IntStartFull( nMfsNodes );
    vGroupMap = Vec_IntStartFull( nMfsNodes );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjCopyArray(p, i) > 0 {
            Vec_IntWriteEntry( vMfs2Old, Gia_ObjCopyArray(p, i), i ); }
    }
    // collect nodes
    curCi = nRealPis;
    curCo = 0;
    for  i = 0; i < nBoxes; i++  {
        nBoxIns = Tim_ManBoxInputNum( pManTime, i );
        nBoxOuts = Tim_ManBoxOutputNum( pManTime, i );
        vArray = Vec_WecEntry( vGroups, i );
        for  k = 0; k < nBoxIns; k++  {
            pObj = Gia_ManCo( p, curCo + k );
            iMfsId = Gia_ObjCopyArray( p, Gia_ObjId(p, pObj) );
            assert(iMfsId > 0, "iMfsId > 0");
            Vec_IntPush( vArray, iMfsId );
            Vec_IntWriteEntry( vGroupMap, iMfsId, Abc_Var2Lit(i,0) );
        }
        for  k = 0; k < nBoxOuts; k++  {
            pObj = Gia_ManCi( p, curCi + k );
            iMfsId = Gia_ObjCopyArray( p, Gia_ObjId(p, pObj) );
            assert(iMfsId > 0, "iMfsId > 0");
            Vec_IntPush( vArray, iMfsId );
            Vec_IntWriteEntry( vGroupMap, iMfsId, Abc_Var2Lit(i,1) );
        }
        curCo += nBoxIns;
        curCi += nBoxOuts;
    }
    curCo += nRealPos;
    assert(curCi == Gia_ManCiNum(p), "curCi == Gia_ManCiNum(p)");
    assert(curCo == Gia_ManCoNum(p), "curCo == Gia_ManCoNum(p)");

    // collect nodes in the given order
    vBoxesLeft = Vec_IntAlloc( nBoxes );
    vMfsTopo = Sfm_NtkDfs( pNtk, vGroups, vGroupMap, vBoxesLeft, fAllBoxes );
    assert(Vec_IntSize(vBoxesLeft) <= nBoxes, "Vec_IntSize(vBoxesLeft) <= nBoxes");
    assert(Vec_IntSize(vMfsTopo) > 0, "Vec_IntSize(vMfsTopo) > 0");

    // start new GIA
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );

    // start mapping
    vMapping  = Vec_IntStart( Gia_ManObjNum(p) );
    vMapping2 = Vec_IntStart( 1 );
    // create const LUT
    Vec_IntWriteEntry( vMapping, 0, Vec_IntSize(vMapping2) );
    Vec_IntPush( vMapping2, 0 );
    Vec_IntPush( vMapping2, 0 );

    // map constant
    Vec_IntWriteEntry( vMfs2Gia, Gia_ObjCopyArray(p, 0), 0 );
    // map primary inputs
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        if i < nRealPis {
            Vec_IntWriteEntry( vMfs2Gia, Gia_ObjCopyArray(p, Id), Gia_ManAppendCi(pNew) ); }
    }
    // map internal nodes
    vLeaves = Vec_IntAlloc( 6 );
    vCover = Vec_IntAlloc( 1 << 16 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vMfsTopo)) && (((iMfsId) = Vec_IntEntry(vMfsTopo, i)), 1); i++  {
        pTruth = Sfm_NodeReadTruth( pNtk, iMfsId );
        iGroup = Vec_IntEntry( vGroupMap, iMfsId );
        vArray = Sfm_NodeReadFanins( pNtk, iMfsId ); // belongs to pNtk
        if Vec_IntSize(vArray) == 1 && Vec_IntEntry(vArray,0) < nBbOuts { // skip unreal inputs
        
            // create CI for the output of black box
            assert(Abc_LitIsCompl(iGroup), "Abc_LitIsCompl(iGroup)");
            iLitNew = Gia_ManAppendCi( pNew );
            Vec_IntWriteEntry( vMfs2Gia, iMfsId, iLitNew );
            continue;
        }
        Vec_IntClear( vLeaves );
        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vArray)) && (((Fanin) = Vec_IntEntry(vArray, k)), 1); k++  {
            iLitNew = Vec_IntEntry( vMfs2Gia, Fanin );  assert(iLitNew >= 0, "iLitNew >= 0");
            Vec_IntPush( vLeaves, iLitNew );            
        }
        if iGroup == -1 { // internal node
        
            assert(Sfm_NodeReadUsed(pNtk, iMfsId), "Sfm_NodeReadUsed(pNtk, iMfsId)");
            if Gia_ObjLutIsMux(p, Vec_IntEntry(vMfs2Old, iMfsId)) {
                var MapSize int32  = Vec_IntSize(vMapping2);
                var nVarsNew int32;
                Abc_TtSimplify( pTruth, Vec_IntArray(vLeaves), Vec_IntSize(vLeaves), &nVarsNew );
                Vec_IntShrink( vLeaves, nVarsNew );
                iLitNew = Gia_ManFromIfLogicCreateLut( pNew, pTruth, vLeaves, vCover, vMapping, vMapping2 );
                if MapSize < Vec_IntSize(vMapping2) {
                    assert(Vec_IntEntryLast(vMapping2) == Abc_Lit2Var(iLitNew), "Vec_IntEntryLast(vMapping2) == Abc_Lit2Var(iLitNew)");
                    Vec_IntWriteEntry(vMapping2, Vec_IntSize(vMapping2)-1, -Abc_Lit2Var(iLitNew) );
                }
            
            } else {
                iLitNew = Gia_ManFromIfLogicCreateLut( pNew, pTruth, vLeaves, vCover, vMapping, vMapping2 ); }
        
        } else if Abc_LitIsCompl(iGroup) { // internal CI
        
            //Dau_DsdPrintFromTruth( pTruth, Vec_IntSize(vLeaves) );
            iLitNew = Gia_ManAppendCi( pNew );
        
        } else { // internal CO
        
            assert(pTruth[0] == uTruthVar || pTruth[0] == ~uTruthVar, "pTruth[0] == uTruthVar || pTruth[0] == ~uTruthVar");
            iLitNew = Gia_ManAppendCo( pNew, Abc_LitNotCond(Vec_IntEntry(vLeaves, 0), pTruth[0] == ~uTruthVar) );
            //printf("Group = %d. po = %d\n", iGroup>>1, iMfsId );
        }
        Vec_IntWriteEntry( vMfs2Gia, iMfsId, iLitNew );
    }
    Vec_IntFree( vCover );
    Vec_IntFree( vLeaves );

    // map primary outputs
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        if i < Gia_ManCoNum(p) - nRealPos { // internal COs
        
            iMfsId = Gia_ObjCopyArray( p, Gia_ObjId(p, pObj) );
            iGroup = Vec_IntEntry( vGroupMap, iMfsId );
            if Vec_IntFind(vMfsTopo, iGroup) >= 0 {
                iLitNew = Vec_IntEntry( vMfs2Gia, iMfsId );
                if iLitNew < 0 {
                    continue; }
                assert(iLitNew >= 0, "iLitNew >= 0");
            }
            continue;
        }
        iLitNew = Vec_IntEntry( vMfs2Gia, Gia_ObjCopyArray(p, Gia_ObjFaninId0p(p, pObj)) );
        assert(iLitNew >= 0, "iLitNew >= 0");
        Gia_ManAppendCo( pNew, Abc_LitNotCond(iLitNew, Gia_ObjFaninC0(pObj)) );
    }

    // finish mapping 
    if Vec_IntSize(vMapping) > Gia_ManObjNum(pNew) {
        Vec_IntShrink( vMapping, Gia_ManObjNum(pNew) );
    } else {
        Vec_IntFillExtra( vMapping, Gia_ManObjNum(pNew), 0 ); }
    assert(Vec_IntSize(vMapping) == Gia_ManObjNum(pNew), "Vec_IntSize(vMapping) == Gia_ManObjNum(pNew)");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vMapping)) && (((iLitNew) = Vec_IntEntry(vMapping, i)), 1); i++  {
        if iLitNew > 0 {
            Vec_IntAddToEntry( vMapping, i, Gia_ManObjNum(pNew) ); }
    }
    Vec_IntAppend( vMapping, vMapping2 );
    Vec_IntFree( vMapping2 );
    assert(pNew.vMapping == nil, "pNew.vMapping == nil");
    pNew.vMapping = vMapping;

    // create new timing manager and extra AIG
    if pManTime != nil {
        pNew.pManTime = Gia_ManUpdateTimMan2( p, vBoxesLeft, 0 ); }
    // update extra STG
    if p.pAigExtra {
        pNew.pAigExtra = Gia_ManUpdateExtraAig2( p.pManTime, p.pAigExtra, vBoxesLeft ); }
    // duplicated flops
    if p.vRegClasses {
        pNew.vRegClasses = Vec_IntDup( p.vRegClasses ); }
    // duplicated initial state
    if p.vRegInits {
        pNew.vRegInits = Vec_IntDup( p.vRegInits ); }
    pNew.nAnd2Delay = p.nAnd2Delay;

    // cleanup
    Vec_WecFree( vGroups );
    Vec_IntFree( vMfsTopo );
    Vec_IntFree( vGroupMap );
    Vec_IntFree( vMfs2Gia );
    Vec_IntFree( vMfs2Old );
    Vec_IntFree( vBoxesLeft );
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPerformMfs(p *Gia_Man_t, pPars *Sfm_Par_t) *Gia_Man_t {
    var pNtk *Sfm_Ntk_t;
    var pNew *Gia_Man_t;
    var nFaninMax, nNodes int32;
    assert(Gia_ManRegNum(p) == 0, "Gia_ManRegNum(p) == 0");
    assert(p.vMapping != nil, "p.vMapping != nil");
    if p.pManTime != nil && p.pAigExtra == nil {
        Abc_Print( 1, "Timing manager is given but there is no GIA of boxes.\n" );
        return nil;
    }
    if p.pManTime != nil && p.pAigExtra != nil && Gia_ManCiNum(p.pAigExtra) > 15 {
        Abc_Print( 1, "Currently \"&mfs\" cannot process the network containing white-boxes with more than 15 inputs.\n" );
        return nil;
    }
    // count fanouts
    nFaninMax = Gia_ManLutSizeMax( p );
    if nFaninMax > 15 {
        Abc_Print( 1, "Currently \"&mfs\" cannot process the network containing nodes with more than 15 fanins.\n" );
        return nil;
    }
    // collect information
    pNtk = Gia_ManExtractMfs( p );
    // perform optimization
    nNodes = Sfm_NtkPerform( pNtk, pPars );
    if nNodes == 0 {
        if p.pManTime {
        Abc_Print( 1, "The network is not changed by \"&mfs\".\n" ); }
        pNew = Gia_ManDup( p );
        pNew.vMapping = Vec_IntDup( p.vMapping );
        Gia_ManTransferTiming( pNew, p );
    
    } else {
        pNew = Gia_ManInsertMfs( p, pNtk, pPars.fAllBoxes );
        if pPars.fVerbose {
            Abc_Print( 1, "The network has %d nodes changed by \"&mfs\".\n", nNodes ); }
        // check integrity
        //Gia_ManCheckIntegrityWithBoxes( pNew );
    }
    Sfm_NtkFree( pNtk );
    return pNew;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

