package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaResub2.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaResub2.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilTruth.h"
//#include "misc/vec/vecHsh.h"
//#include "opt/dau/dau.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////
 
   


// extern void Abc_ResubPrepareManager( int nWords );

// extern int Abc_ResubComputeFunction( void ** ppDivs, int nDivs, int nWords, int nLimit, int nDivsMax, int iChoice, int fUseXor, int fDebug, int fVerbose, int ** ppArray );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




// marks MFFC and returns its size


// collects divisors and maps them into nodes
// assumes MFFC is already marked


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_ResubPrintDivs(ppDivs *any, nDivs int32) {
    var pDivs **word  = (word **)ppDivs;
    var i int32;
    for  i = 0; i < nDivs; i++  {
        fmt.Printf("Div %2d : ", i);
        Dau_DsdPrintFromTruth( pDivs[i], 6 );
    }
}




/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Creating a window with support composed of primary inputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
// returns the number of nodes added to the window when is iPivot is added
// the window nodes in vNodes are labeled with the current traversal ID
// the new node iPivot and its fanout are temporarily labeled and then unlabeled
func Gia_WinTryAddingNode(p *Gia_Man_t, iPivot int32, iPivot2 int32, vLevels *Vec_Wec_t, vNodes *Vec_Int_t) int32 {
    var vLevel *Vec_Int_t; 
    var pObj, pFanout *Gia_Obj_t;
    var k, i, f, Count int32;
    // precondition:  the levelized structure is empty
    assert(Vec_WecSizeSize(vLevels) == 0, "Vec_WecSizeSize(vLevels) == 0");
    // precondition:  the new object to be added (iPivot) is not in the window
    assert(!Gia_ObjIsTravIdCurrentId(p, iPivot), "!Gia_ObjIsTravIdCurrentId(p, iPivot)");
    // add the object to the window and to the levelized structure
    Gia_ObjSetTravIdCurrentId( p, iPivot );
    Vec_WecPush( vLevels, Gia_ObjLevelId(p, iPivot), iPivot );
    // the same about the second node if it is given
    if iPivot2 != -1 {
        // precondition:  the new object to be added (iPivot2) is not in the window
        assert(!Gia_ObjIsTravIdCurrentId(p, iPivot2), "!Gia_ObjIsTravIdCurrentId(p, iPivot2)");
        // add the object to the window and to the levelized structure
        Gia_ObjSetTravIdCurrentId( p, iPivot2 );
        Vec_WecPush( vLevels, Gia_ObjLevelId(p, iPivot2), iPivot2 );
    }
    // iterate through all objects and explore their fanouts
    // inlining of macro Vec_WecForEachLevel
    for  k = 0; (k < Vec_WecSize(vLevels)) && (((vLevel) = Vec_WecEntry(vLevels, k)), 1); k++  {
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vLevel)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLevel,i))); i++  {
            // inlining of macro Gia_ObjForEachFanoutStatic
            for  f = 0; (f < Gia_ObjFanoutNum(p, pObj))   && (((pFanout) = Gia_ObjFanout(p, pObj, f)), 1); f++  {
                if f == 5 { // explore first 5 fanouts of the node
                    break; }
                if Gia_ObjIsAnd(pFanout) && // internal node
                    !Gia_ObjIsTravIdCurrent(p, pFanout) && // not in the window
                     Gia_ObjIsTravIdCurrent(p, Gia_ObjFanin0(pFanout)) && // but fanins are
                     Gia_ObjIsTravIdCurrent(p, Gia_ObjFanin1(pFanout)) {  // in the window
                
                    // add fanout to the window and to the levelized structure
                    Gia_ObjSetTravIdCurrent( p, pFanout );
                    Vec_WecPush( vLevels, Gia_ObjLevel(p, pFanout), Gia_ObjId(p, pFanout) );
                    // count the number of nodes in the structure
                    Count++;
                }
            }
        }
    }
    // iterate through the nodes in the levelized structure
    // inlining of macro Vec_WecForEachLevel
    for  k = 0; (k < Vec_WecSize(vLevels)) && (((vLevel) = Vec_WecEntry(vLevels, k)), 1); k++  {
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vLevel)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLevel,i))); i++  {
            if vNodes == nil { // it was a test run - unmark the node
                Gia_ObjSetTravIdPrevious( p, pObj );
            } else { // it was a real run - permanently add to the node to the window
                Vec_IntPush( vNodes, Gia_ObjId(p, pObj) ); }
        }
        // clean the levelized structure
        Vec_IntClear( vLevel );
    }
    // return the number of nodes to be added
    return Count;
}
// find the first PI to add based on the fanout count
func Gia_WinAddCiWithMaxFanouts(p *Gia_Man_t) int32 {
    var i int32;
    var Id int32;
    var nMaxFan int32 = -1;
    var iMaxFan int32 = -1;
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        if nMaxFan < Gia_ObjFanoutNumId(p, Id) {
            nMaxFan = Gia_ObjFanoutNumId(p, Id);
            iMaxFan = Id;
        }
    }
    assert(iMaxFan >= 0, "iMaxFan >= 0");
    return iMaxFan;
}
// find the next PI to add based on how many nodes will be added to the window
func Gia_WinAddCiWithMaxDivisors(p *Gia_Man_t, vLevels *Vec_Wec_t) int32 {
    var i int32;
    var Id int32;
    var nCurFan int32;
    var nMaxFan int32 = -1;
    var iMaxFan int32 = -1;
    // inlining of macro Gia_ManForEachCiId
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++  {
        if Gia_ObjIsTravIdCurrentId( p, Id ) {
            continue; }
        nCurFan = Gia_WinTryAddingNode( p, Id, -1, vLevels, nil );
        if nMaxFan < nCurFan {
            nMaxFan = nCurFan;
            iMaxFan = Id;
        }
    }
    assert(iMaxFan >= 0, "iMaxFan >= 0");
    return iMaxFan;
}
// check if the node has unmarked fanouts
func Gia_WinNodeHasUnmarkedFanouts(p *Gia_Man_t, iPivot int32) int32 {
    var f, iFan int32;
    // inlining of macro Gia_ObjForEachFanoutStaticId
    for  f = 0; (f < Gia_ObjFanoutNumId(p, iPivot))   && (((iFan) = Gia_ObjFanoutId(p, iPivot, f)), 1); f++  {
        if !Gia_ObjIsTravIdCurrentId(p, iFan) {
            return 1; }
    }
    return 0;
}
// this is a translation procedure, which converts the array of node IDs (vObjs) 
// into the internal represnetation for the resub engine, which is returned
// (this procedure is not needed when we simply construct a window)
func Gia_RsbCiTranslate(p *Gia_Man_t, vObjs *Vec_Int_t, vMap *Vec_Int_t) *Vec_Int_t {
    var i, iObj, Lit0, Lit1, Fan0, Fan1 int32;
    var vNodes *Vec_Int_t  = Vec_IntAlloc( 100 );
    assert(Vec_IntSize(vMap) == Gia_ManObjNum(p), "Vec_IntSize(vMap) == Gia_ManObjNum(p)");
    Vec_IntPushTwo( vNodes, 0, 0 ); // const0 node
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        var pObj *Gia_Obj_t  = Gia_ManObj(p, iObj);
        assert(Gia_ObjIsTravIdCurrentId(p, iObj), "Gia_ObjIsTravIdCurrentId(p, iObj)");
        if Gia_ObjIsCi(pObj) {
            Fan0 = 0;
        } else {
            Fan0 = Vec_IntEntry(vMap, Gia_ObjFaninId0(pObj, iObj));
        }
        if Gia_ObjIsCi(pObj) {
            Fan1 = 0;
        } else {
            Fan1 = Vec_IntEntry(vMap, Gia_ObjFaninId1(pObj, iObj));
        }
        if Gia_ObjIsCi(pObj) {
            Lit0 = 0;
        } else {
            Lit0 = Abc_LitNotCond( Fan0, Gia_ObjFaninC0(pObj) );
        }
        if Gia_ObjIsCi(pObj) {
            Lit1 = 0;
        } else {
            Lit1 = Abc_LitNotCond( Fan1, Gia_ObjFaninC1(pObj) );
        }
        Vec_IntWriteEntry( vMap, iObj, Vec_IntSize(vNodes) );
        Vec_IntPushTwo( vNodes, Lit0, Lit1 );
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vObjs)) && (((iObj) = Vec_IntEntry(vObjs, i)), 1); i++  {
        if Gia_WinNodeHasUnmarkedFanouts( p, iObj ) != 0 {
            Vec_IntPushTwo( vNodes, Vec_IntEntry(vMap, iObj), Vec_IntEntry(vMap, iObj) ); }
    }
    return vNodes;
}
// construct a high-volume window support by the given number (nPis) of primary inputs
func Gia_RsbCiWindow(p *Gia_Man_t, nPis int32) *Vec_Int_t {
    var vRes *Vec_Int_t;  var i, iMaxFan int32;
    var vNodes *Vec_Int_t   = Vec_IntAlloc( 100 );
    var vMap *Vec_Int_t     = Vec_IntStartFull( Gia_ManObjNum(p) );
    var vLevels *Vec_Wec_t  = Vec_WecStart( Gia_ManLevelNum(p)+1 );
    Gia_ManStaticFanoutStart( p );
    Gia_ManIncrementTravId(p);
    // add the first one
    iMaxFan = Gia_WinAddCiWithMaxFanouts( p );
    Gia_ObjSetTravIdCurrentId( p, iMaxFan );
    Vec_IntPush( vNodes, iMaxFan );
    // add remaining ones
    for  i = 1; i < nPis; i++  {
        iMaxFan = Gia_WinAddCiWithMaxDivisors( p, vLevels );
        Gia_WinTryAddingNode( p, iMaxFan, -1, vLevels, vNodes );
    }
    Vec_IntSort( vNodes, 0 );
    vRes = Gia_RsbCiTranslate( p, vNodes, vMap );
    Gia_ManStaticFanoutStop( p );
    Vec_WecFree( vLevels );
    Vec_IntFree( vMap );
//Vec_IntPrint( vNodes );
    Vec_IntFree( vNodes );
    return vRes;
}
func Gia_RsbCiWindowTest(p *Gia_Man_t) {
    var vWin *Vec_Int_t  = Gia_RsbCiWindow( p, 6 );
    //Vec_IntPrint( vWin );
    Vec_IntFree( vWin );
}




/**Function*************************************************************

  Synopsis    [Return initial window for the given node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ObjFaninId(pObj *Gia_Obj_t, iObj int32, n int32) int32 { if n != 0 {
    return Gia_ObjFaninId1(pObj, iObj);
} else {
    return Gia_ObjFaninId0(pObj, iObj);
}  }

func Gia_ObjTravIsTopTwo(p *Gia_Man_t, iNodeA int32) int32 { return (p.pTravIds[iNodeA] >= p.nTravIds - 1);       }
func Gia_ObjTravIsSame(p *Gia_Man_t, iNodeA int32, iNodeB int32) int32 { return (p.pTravIds[iNodeA] == p.pTravIds[iNodeB]);   }
func Gia_ObjTravSetSame(p *Gia_Man_t, iNodeA int32, iNodeB int32) { p.pTravIds[iNodeA] = p.pTravIds[iNodeB];             }

// collect nodes on the path from the meeting point to the root node, excluding the meeting point
func Gia_RsbWindowGather(p *Gia_Man_t, vPaths *Vec_Int_t, iNode int32, vVisited *Vec_Int_t) {
    var iPrev int32;
    if iNode == 0 {
        return; }
    Vec_IntPush( vVisited, iNode );
    iPrev = Vec_IntEntry( vPaths, iNode );
    if iPrev == 0 {
        return; }
    assert((Gia_ObjTravIsSame(p, iPrev, iNode) != 0), "Gia_ObjTravIsSame(p, iPrev, iNode)");
    Gia_RsbWindowGather( p, vPaths, iPrev, vVisited );
}
// explore the frontier of nodes in the breadth-first traversal
func Gia_RsbWindowExplore(p *Gia_Man_t, vVisited *Vec_Int_t, iStart int32, vPaths *Vec_Int_t, piMeet *int32, piNode *int32) int32 {
    var i int32;
    var n int32;
    var iObj int32;
    var iLimit int32 = Vec_IntSize( vVisited );
    *piMeet = *piNode = 0;
    // inlining of macro Vec_IntForEachEntryStartStop
    for  i = iStart; (i < iLimit) && (((iObj) = Vec_IntEntry(vVisited, i)), 1); i++  {
        var pObj *Gia_Obj_t  = Gia_ManObj( p, iObj );
        if !Gia_ObjIsAnd(pObj) {
            continue; }
        for  n = 0; n < 2; n++  {
            var iFan int32  = Gia_ObjFaninId( pObj, iObj, n );
            // if the node was visited on the paths to both fanins, collect it
            if (Gia_ObjTravIsTopTwo(p, iObj) != 0) && (Gia_ObjTravIsTopTwo(p, iFan) != 0) && !(Gia_ObjTravIsSame(p, iObj, iFan) != 0) {
                *piMeet = iFan;
                *piNode = iObj;
                return 1;
            }
            // if the node was visited already on this path, skip it
            if Gia_ObjTravIsTopTwo(p, iFan) != 0 {
                assert((Gia_ObjTravIsSame(p, iObj, iFan) != 0), "Gia_ObjTravIsSame(p, iObj, iFan)");
                continue;
            }
            // label the node as visited
            Gia_ObjTravSetSame( p, iFan, iObj );
            Vec_IntWriteEntry( vPaths, iFan, iObj ); 
            Vec_IntPush( vVisited, iFan );
        }
    }
    return 0;
}
func Gia_RsbWindowInit(p *Gia_Man_t, vPaths *Vec_Int_t, iPivot int32, nIter int32) *Vec_Int_t {
    var vVisited *Vec_Int_t  = Vec_IntAlloc( 100 );
    var pPivot *Gia_Obj_t  = Gia_ManObj( p, iPivot ); 
    var i, n, iStart int32;
    assert(Gia_ObjIsAnd(pPivot), "Gia_ObjIsAnd(pPivot)");
    // start paths for both fanins of the pivot node
    for  n = 0; n < 2; n++  {
        var iFan int32  = Gia_ObjFaninId( pPivot, iPivot, n );
        Gia_ManIncrementTravId(p);
        Vec_IntPush( vVisited, iFan );
        Vec_IntWriteEntry( vPaths, iFan, 0 ); 
        Gia_ObjSetTravIdCurrentId( p, iFan );
    }
    // perform several iterations of breadth-first search
    for  i = 0; i < nIter; i++  {
        var iMeet int32;
        var iNode int32;
        var iNext int32 = Vec_IntSize(vVisited);
        if Gia_RsbWindowExplore( p, vVisited, iStart, vPaths, &iMeet, &iNode ) != 0 {
            // found the shared path
            if Gia_ObjIsTravIdCurrentId(p, iMeet) {
                assert(Gia_ObjIsTravIdPreviousId(p, iNode), "Gia_ObjIsTravIdPreviousId(p, iNode)");
            } else if Gia_ObjIsTravIdPreviousId(p, iMeet) {
                assert(Gia_ObjIsTravIdCurrentId(p, iNode), "Gia_ObjIsTravIdCurrentId(p, iNode)");
            } else { assert(false, "0"); }
            // collect the initial window
            Vec_IntClear( vVisited );
            Gia_RsbWindowGather( p, vPaths, Vec_IntEntry(vPaths, iMeet), vVisited );
            Gia_RsbWindowGather( p, vPaths, iNode, vVisited );
            Vec_IntPush( vVisited, iPivot );
            break;
        }
        iStart = iNext;
    }
    // if no meeting point is found, make sure to return NULL
    if i == nIter {
        Vec_IntFreeP( &vVisited ); }
    return vVisited;
}
func Gia_RsbCreateWindowInputs(p *Gia_Man_t, vWin *Vec_Int_t) *Vec_Int_t {
    var vInputs *Vec_Int_t  = Vec_IntAlloc(10);
    var pObj *Gia_Obj_t; var i, n, iObj int32;
    Gia_ManIncrementTravId(p);
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vWin)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vWin,i))); i++  {
        Gia_ObjSetTravIdCurrent(p, pObj);
    }
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vWin)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vWin,i))); i++  {
        assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
        for  n = 0; n < 2; n++  {
            var iFan int32;
            if n != 0 {
                iFan = Gia_ObjFaninId1p(p, pObj);
            } else {
                iFan = Gia_ObjFaninId0p(p, pObj);
            }
            if !Gia_ObjIsTravIdCurrentId(p, iFan) {
                Vec_IntPushUnique( vInputs, iFan ); }
        }
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vInputs)) && (((iObj) = Vec_IntEntry(vInputs, i)), 1); i++  {
        Gia_ObjSetTravIdCurrentId( p, iObj );
        Vec_IntPush( vWin, iObj );
    }
    return vInputs;
}

/**Function*************************************************************

  Synopsis    [Grow window for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_RsbAddSideInputs(p *Gia_Man_t, vLevels *Vec_Wec_t, vWin *Vec_Int_t) {
    var vLevel *Vec_Int_t; 
    var pObj, pFanout *Gia_Obj_t;
    var k, i, f, iObj int32;
    // precondition: the levelized structure is empty
    assert(Vec_WecSizeSize(vLevels) == 0, "Vec_WecSizeSize(vLevels) == 0");
    // precondition: window nodes are labeled with the current ID
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vWin)) && (((iObj) = Vec_IntEntry(vWin, i)), 1); i++  {
        assert(Gia_ObjIsTravIdCurrentId(p, iObj), "Gia_ObjIsTravIdCurrentId(p, iObj)");
        Vec_WecPush( vLevels, Gia_ObjLevelId(p, iObj), iObj );
    }
    // iterate through all objects and explore their fanouts
    // inlining of macro Vec_WecForEachLevel
    for  k = 0; (k < Vec_WecSize(vLevels)) && (((vLevel) = Vec_WecEntry(vLevels, k)), 1); k++  {
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vLevel)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLevel,i))); i++  {
            // inlining of macro Gia_ObjForEachFanoutStatic
            for  f = 0; (f < Gia_ObjFanoutNum(p, pObj))   && (((pFanout) = Gia_ObjFanout(p, pObj, f)), 1); f++  {
                if f == 5 { // explore first 5 fanouts of the node
                    break; }
                if Gia_ObjIsAnd(pFanout) && // internal node
                    !Gia_ObjIsTravIdCurrent(p, pFanout) && // not in the window
                     Gia_ObjIsTravIdCurrent(p, Gia_ObjFanin0(pFanout)) && // but fanins are
                     Gia_ObjIsTravIdCurrent(p, Gia_ObjFanin1(pFanout)) {  // in the window
                
                    // add fanout to the window and to the levelized structure
                    Gia_ObjSetTravIdCurrent( p, pFanout );
                    Vec_WecPush( vLevels, Gia_ObjLevel(p, pFanout), Gia_ObjId(p, pFanout) );
                    Vec_IntPush( vWin, Gia_ObjId(p, pFanout) );
                }
            }
        }
    }
    // iterate through the nodes in the levelized structure
    // inlining of macro Vec_WecForEachLevel
    for  k = 0; (k < Vec_WecSize(vLevels)) && (((vLevel) = Vec_WecEntry(vLevels, k)), 1); k++  {
        Vec_IntClear( vLevel );
    }
}
// expland inputs until saturation while adding the side-fanouts
func Gia_RsbExpandInputs(p *Gia_Man_t, vLevels *Vec_Wec_t, vWin *Vec_Int_t, vInputs *Vec_Int_t) {
    var pObj *Gia_Obj_t; 
    var i int32;
    var n int32;
    var iFans [2]int32;
    var fChange int32 = 1;
    for  fChange != 0  {
        fChange = 0;
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vInputs)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vInputs,i))); i++  {
            if !Gia_ObjIsAnd(pObj) {
                continue; }
            iFans[0] = Gia_ObjFaninId0p(p, pObj);
            iFans[1] = Gia_ObjFaninId1p(p, pObj);
            if !Gia_ObjIsTravIdCurrentId(p, iFans[0]) && !Gia_ObjIsTravIdCurrentId(p, iFans[1]) {
                continue; }
            Vec_IntRemove( vInputs, Gia_ObjId(p, pObj) );
            assert(Vec_IntFind(vWin, Gia_ObjId(p, pObj)) >= 0, "Vec_IntFind(vWin, Gia_ObjId(p, pObj)) >= 0");
            for  n = 0; n < 2; n++  {
                if Gia_ObjIsTravIdCurrentId(p, iFans[n]) {
                    continue; }
                assert(Vec_IntFind(vInputs, iFans[n]) == -1, "Vec_IntFind(vInputs, iFans[n]) == -1");
                Vec_IntPush( vInputs, iFans[n] );
                Gia_WinTryAddingNode( p, iFans[n], -1, vLevels, vWin );
                assert(Gia_ObjIsTravIdCurrentId(p, iFans[n]), "Gia_ObjIsTravIdCurrentId(p, iFans[n])");
            }
            fChange = 1;
        }
    }
}
// select the best input to expand, based on its contribution to the window size
func Gia_RsbSelectOneInput(p *Gia_Man_t, vLevels *Vec_Wec_t, vIns *Vec_Int_t) int32 {
    var i int32;
    var iNode int32;
    var WeightThis int32;
    var WeightBest int32 = -1;
    var pObj *Gia_Obj_t; 
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vIns)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vIns,i))); i++  {
        if Gia_ObjIsAnd(pObj) {
            var iFan0 int32  = Gia_ObjFaninId0p(p, pObj);
            var iFan1 int32  = Gia_ObjFaninId1p(p, pObj);
            assert(!Gia_ObjIsTravIdCurrentId(p, iFan0) && !Gia_ObjIsTravIdCurrentId(p, iFan1), "!Gia_ObjIsTravIdCurrentId(p, iFan0) && !Gia_ObjIsTravIdCurrentId(p, iFan1)");
            WeightThis = Gia_WinTryAddingNode( p, iFan0, iFan1, vLevels, nil );
            if WeightBest < WeightThis {
                WeightBest = WeightThis;
                iNode = Gia_ObjId(p, pObj);
            }
        }
    }
    return iNode;
}
// grow the initial window as long as it fits the input count limit
func Gia_RsbWindowGrow(p *Gia_Man_t, vLevels *Vec_Wec_t, vWin *Vec_Int_t, vIns *Vec_Int_t, nInputsMax int32) {
    var iNode int32;
    Gia_RsbAddSideInputs( p, vLevels, vWin );
    Gia_RsbExpandInputs( p, vLevels, vWin, vIns );
    for  Vec_IntSize(vIns) < nInputsMax && (iNode = Gia_RsbSelectOneInput(p, vLevels, vIns))  {
        var iFan0 int32  = Gia_ObjFaninId0p(p, Gia_ManObj(p, iNode));
        var iFan1 int32  = Gia_ObjFaninId1p(p, Gia_ManObj(p, iNode));
        assert(!Gia_ObjIsTravIdCurrentId(p, iFan0) && !Gia_ObjIsTravIdCurrentId(p, iFan1), "!Gia_ObjIsTravIdCurrentId(p, iFan0) && !Gia_ObjIsTravIdCurrentId(p, iFan1)");
        Gia_WinTryAddingNode( p, iFan0, iFan1, vLevels, vWin );
        assert(Gia_ObjIsTravIdCurrentId(p, iFan0) &&  Gia_ObjIsTravIdCurrentId(p, iFan1), "Gia_ObjIsTravIdCurrentId(p, iFan0) &&  Gia_ObjIsTravIdCurrentId(p, iFan1)");
        Vec_IntRemove( vIns, iNode );
        Vec_IntPushTwo( vIns, iFan0, iFan1 );
        Gia_RsbExpandInputs( p, vLevels, vWin, vIns );
    }
}

/**Function*************************************************************

  Synopsis    [Grow window for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_WinCreateFromCut_rec(p *Gia_Man_t, iObj int32, vWin *Vec_Int_t) {
    var pObj *Gia_Obj_t;
    if Gia_ObjIsTravIdCurrentId(p, iObj) {
        return; }
    Gia_ObjSetTravIdCurrentId(p, iObj);
    pObj = Gia_ManObj( p, iObj );
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_WinCreateFromCut_rec( p, Gia_ObjFaninId0(pObj, iObj), vWin );
    Gia_WinCreateFromCut_rec( p, Gia_ObjFaninId1(pObj, iObj), vWin );
    Vec_IntPush( vWin, iObj );
}
// uses levelized structure (vLevels) to collect in array vWin divisors supported by the cut (vIn)
func Gia_WinCreateFromCut(p *Gia_Man_t, iPivot int32, vIn *Vec_Int_t, vLevels *Vec_Wec_t, vWin *Vec_Int_t) {
    var vLevel *Vec_Int_t; 
    var pObj, pFanout *Gia_Obj_t;
    var k, i, f, iObj, Level int32;
    var vUsed *Vec_Int_t  = Vec_IntAlloc( 100 );
    // precondition:  the levelized structure is empty
    assert(Vec_WecSizeSize(vLevels) == 0, "Vec_WecSizeSize(vLevels) == 0");
    // clean the resulting array
    Vec_IntClear( vWin );
    // collect leaves
    Gia_ManIncrementTravId( p );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vIn)) && (((iObj) = Vec_IntEntry(vIn, i)), 1); i++  {
        Gia_ObjSetTravIdCurrentId( p, iObj );
        Vec_IntPush( vWin, iObj );
    }
    // collect internal cone
    Gia_WinCreateFromCut_rec( p, iPivot, vWin );
    // add nodes to the levelized structure
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vWin)) && (((iObj) = Vec_IntEntry(vWin, i)), 1); i++  {
        Vec_WecPush( vLevels, Gia_ObjLevelId(p, iObj), iObj );
        Vec_IntPushUniqueOrder( vUsed, Gia_ObjLevelId(p, iObj) );
    }
    // iterate through all objects and explore their fanouts
    //Vec_WecForEachLevel( vLevels, vLevel, k )
    // inlining of macro Vec_IntForEachEntry
    for  k = 0; (k < Vec_IntSize(vUsed)) && (((Level) = Vec_IntEntry(vUsed, k)), 1); k++  {
        vLevel = Vec_WecEntry( vLevels, Level );
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vLevel)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vLevel,i))); i++  {
            // inlining of macro Gia_ObjForEachFanoutStatic
            for  f = 0; (f < Gia_ObjFanoutNum(p, pObj))   && (((pFanout) = Gia_ObjFanout(p, pObj, f)), 1); f++  {
                if f == 5 { // explore first 5 fanouts of the node
                    break; }
                if Gia_ObjIsAnd(pFanout) && // internal node
                    !Gia_ObjIsTravIdCurrent(p, pFanout) && // not in the window
                     Gia_ObjIsTravIdCurrent(p, Gia_ObjFanin0(pFanout)) && // but fanins are
                     Gia_ObjIsTravIdCurrent(p, Gia_ObjFanin1(pFanout)) {  // in the window
                
                    // add fanout to the window and to the levelized structure
                    Gia_ObjSetTravIdCurrent( p, pFanout );
                    Vec_WecPush( vLevels, Gia_ObjLevel(p, pFanout), Gia_ObjId(p, pFanout) );
                    Vec_IntPush( vWin, Gia_ObjId(p, pFanout) );
                    Vec_IntPushUniqueOrder( vUsed, Gia_ObjLevel(p, pFanout) );
                }
            }
        }
        Vec_IntClear( vLevel );
    }
    Vec_IntSort( vWin, 0 );
    Vec_IntFree( vUsed );
}
// update the cut until both fanins of AND nodes are not in the cut
func Gia_RsbExpandCut(p *Gia_Man_t, vIns *Vec_Int_t) int32 {
    var fOnlyPis int32;
    var fChange int32 = 1;
    var nSize int32 = Vec_IntSize(vIns);
    for  fChange != 0  {
        var pObj *Gia_Obj_t;
        var i, iFan0, iFan1, fHave0, fHave1 int32;
        fOnlyPis = 1;
        fChange = 0;
        // check if some nodes can be expanded without increasing cut size
        // inlining of macro Gia_ManForEachObjVec
        for  i = 0; (i < Vec_IntSize(vIns)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vIns,i))); i++  {
            assert(Gia_ObjIsTravIdCurrent(p, pObj), "Gia_ObjIsTravIdCurrent(p, pObj)");
            if !Gia_ObjIsAnd(pObj) {
                continue; }
            fOnlyPis = 0;
            iFan0 = Gia_ObjFaninId0p(p, pObj);
            iFan1 = Gia_ObjFaninId1p(p, pObj);
            fHave0 = Gia_ObjIsTravIdCurrentId(p, iFan0);
            fHave1 = Gia_ObjIsTravIdCurrentId(p, iFan1);
            if !fHave0 && !(fHave1 != 0) {
                continue; }
            // can expand because one of the fanins is already in the cut
            // remove current cut node
            Vec_IntDrop( vIns, i );
            // add missing fanin
            if !fHave0 {
                Vec_IntPush( vIns, iFan0 );
                Gia_ObjSetTravIdCurrentId( p, iFan0 );
            }
            if !(fHave1 != 0) {
                Vec_IntPush( vIns, iFan1 );
                Gia_ObjSetTravIdCurrentId( p, iFan1 );
            }
            fChange = 1;
            break;
        }
    }
    assert(Vec_IntSize(vIns) <= nSize, "Vec_IntSize(vIns) <= nSize");
    return fOnlyPis;
}
func Gia_RsbFindFaninAdd(iFan int32, pFanins []int32, pFaninCounts []int32, nFanins int32) int32 {
    var i int32;
    for  i = 0; i < nFanins; i++  {
        if pFanins[i] == iFan {
            break; }
    }
    pFanins[i] = iFan;
    pFaninCounts[i]++;
    return nFanins + (i == nFanins);
}                        
func Gia_RsbFindFaninToAddToCut(p *Gia_Man_t, vIns *Vec_Int_t) int32 {
    var pObj *Gia_Obj_t;
    var nFanins int32;
    var pFanins [64]int32 ;
    var pFaninCounts [64]int32 ;
    var i int32;
    var iFan0 int32;
    var iFan1 int32;
    var iFanMax int32 = -1;
    var CountMax int32;
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vIns)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vIns,i))); i++  {
        if !Gia_ObjIsAnd(pObj) {
            continue; }
        iFan0 = Gia_ObjFaninId0p(p, pObj);
        iFan1 = Gia_ObjFaninId1p(p, pObj);
        assert(!Gia_ObjIsTravIdCurrentId(p, iFan0), "!Gia_ObjIsTravIdCurrentId(p, iFan0)");
        assert(!Gia_ObjIsTravIdCurrentId(p, iFan1), "!Gia_ObjIsTravIdCurrentId(p, iFan1)");
        nFanins = Gia_RsbFindFaninAdd( iFan0, pFanins, pFaninCounts, nFanins );
        nFanins = Gia_RsbFindFaninAdd( iFan1, pFanins, pFaninCounts, nFanins );
        assert(nFanins < 64, "nFanins < 64");
    }
    // find fanin with the highest count
    for  i = 0; i < nFanins; i++  {
//        if ( CountMax < pFaninCounts[i] )
        if CountMax < pFaninCounts[i] || (CountMax == pFaninCounts[i] && (Gia_ObjFanoutNumId(p, iFanMax) < Gia_ObjFanoutNumId(p, pFanins[i]))) {
            CountMax = pFaninCounts[i];
            iFanMax  = pFanins[i];
        }
    }
    return iFanMax;
}
// precondition: nodes in vWin and in vIns are marked with the current ID
func Gia_RsbWindowGrow2(p *Gia_Man_t, iObj int32, vLevels *Vec_Wec_t, vWin *Vec_Int_t, vIns *Vec_Int_t, nInputsMax int32) {
    // window will be recomputed later
    Vec_IntClear( vWin );
    // expand the cut without increasing its cost
    if !(Gia_RsbExpandCut( p, vIns ) != 0) {
        // save it as the best cut
        var vBest *Vec_Int_t;
        if Vec_IntSize(vIns) <= nInputsMax {
            vBest = Vec_IntDup(vIns);
        } else {
            vBest = nil;
        }
        var fOnlyPis, Iter int32;
        // iterate expansion until  
        // (1) the cut cannot be expanded because all leaves are PIs
        // (2) the cut size exceeded the limit for 5 consecutive iterations
        for  !(fOnlyPis != 0) && (Vec_IntSize(vIns) <= nInputsMax || Iter < 5)  {
            var iFanBest int32  = Gia_RsbFindFaninToAddToCut( p, vIns );
            Vec_IntPush( vIns, iFanBest );
            Gia_ObjSetTravIdCurrentId( p, iFanBest );
            fOnlyPis = Gia_RsbExpandCut( p, vIns );
            if Vec_IntSize(vIns) > nInputsMax {
                Iter++;
            } else {
                Iter = 0; }                
            if Vec_IntSize(vIns) <= nInputsMax && (!(vBest != nil) || Vec_IntSize(vBest) <= Vec_IntSize(vIns)) {
                if vBest != nil {
                    Vec_IntClear(vBest);
                } else {
                    vBest = Vec_IntAlloc( 10 ); }
                Vec_IntAppend( vBest, vIns );
            }
        }
        if vBest != nil {
            Vec_IntClear( vIns );
            Vec_IntAppend( vIns, vBest );
            Vec_IntFree( vBest );
        
        } else {
            assert(Vec_IntSize(vIns) > nInputsMax, "Vec_IntSize(vIns) > nInputsMax"); }
    }
    if Vec_IntSize(vIns) <= nInputsMax {
        Vec_IntSort( vIns, 0 );
        Gia_WinCreateFromCut( p, iObj, vIns, vLevels, vWin );
    }
}

/**Function*************************************************************

  Synopsis    [Create window for the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_RsbWindowCompute(p *Gia_Man_t, iObj int32, nInputsMax int32, nLevelsMax int32, vLevels *Vec_Wec_t, vPaths *Vec_Int_t, pvWin **Vec_Int_t, pvIns **Vec_Int_t) int32 {
    var vWin, vIns *Vec_Int_t;
    *pvWin = *pvIns = nil;
    vWin = Gia_RsbWindowInit( p, vPaths, iObj, nLevelsMax );
    if vWin == nil {
        return 0; }
    vIns = Gia_RsbCreateWindowInputs( p, vWin ); 
    // vWin and vIns are labeled with the current trav ID
    //Vec_IntPrint( vWin );    
    //Vec_IntPrint( vIns );    
    if Vec_IntSize(vIns) <= nInputsMax + 3 { // consider windows, which initially has a larger input space
        Gia_RsbWindowGrow2( p, iObj, vLevels, vWin, vIns, nInputsMax ); }
    if Vec_IntSize(vIns) <= nInputsMax {
        Vec_IntSort( vWin, 0 );
        Vec_IntSort( vIns, 0 );
        *pvWin = vWin;
        *pvIns = vIns;
        return 1;
    
    } else {
        Vec_IntFree( vWin );
        Vec_IntFree( vIns );
        return 0;
    }
}

/**Function*************************************************************

  Synopsis    [Derive GIA from the window]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Naive truth-table-based verification.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Enumerate windows of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_RsbEnumerateWindows(p *Gia_Man_t, nInputsMax int32, nLevelsMax int32) {
    var fVerbose int32;
    var fUseHash int32;
    var i, nWins, nWinSize, nInsSize, nOutSize, nNodeGain int32;
    var vLevels *Vec_Wec_t  = Vec_WecStart( Gia_ManLevelNum(p)+1 );
    var vPaths *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var vRefs *Vec_Int_t  = Vec_IntStart( Gia_ManObjNum(p) );
    var pHash *Hsh_VecMan_t  = Hsh_VecManStart( 1000 );
    var pObj *Gia_Obj_t;
    var pIn, pOut *Gia_Man_t;
    var clk abctime  = Abc_Clock();
    Gia_ManStaticFanoutStart( p );
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        var vWin, vIns, vOuts *Vec_Int_t;
        if !(Gia_RsbWindowCompute( p, i, nInputsMax, nLevelsMax, vLevels, vPaths, &vWin, &vIns ) != 0) {
            continue; }
        vOuts = Gia_RsbFindOutputs( p, vWin, vIns, vRefs );
        nWins++;
        nWinSize += Vec_IntSize(vWin);
        nInsSize += Vec_IntSize(vIns);
        nOutSize += Vec_IntSize(vOuts);
   

        if fVerbose != 0 {
            fmt.Printf("\n\nObj %d\n", i);
            Vec_IntPrint( vWin );    
            Vec_IntPrint( vIns );    
            Vec_IntPrint( vOuts );    
            fmt.Printf("\n");
        
        } else if Vec_IntSize(vWin) > 1000 {
            fmt.Printf("Obj %d.   Window: Ins = %d. Ands = %d. Outs = %d.\n", i, Vec_IntSize(vIns), Vec_IntSize(vWin)-Vec_IntSize(vIns), Vec_IntSize(vOuts)); }

        if fUseHash != 0 {
            var nEntries int32  = Hsh_VecSize(pHash);
            Hsh_VecManAdd( pHash, vWin );
            if nEntries == Hsh_VecSize(pHash) {
                Vec_IntFree( vWin );
                Vec_IntFree( vIns );
                Vec_IntFree( vOuts );
                continue;
            }
        }

        pIn = Gia_RsbDeriveGiaFromWindows( p, vWin, vIns, vOuts );
        pOut = Gia_ManResub2Test( pIn );
        //pOut = Gia_ManDup( pIn );
        if !(Gia_ManVerifyTwoTruths( pIn, pOut ) != 0) {
            Gia_ManPrint( pIn );
            Gia_ManPrint( pOut );
            pOut = pOut;
        }

        nNodeGain += Gia_ManAndNum(pIn) - Gia_ManAndNum(pOut);
        Gia_ManStop( pIn );
        Gia_ManStop( pOut );

        Vec_IntFree( vWin );
        Vec_IntFree( vIns );
        Vec_IntFree( vOuts );
    }
    }
    Gia_ManStaticFanoutStop( p );
    Vec_WecFree( vLevels );
    Vec_IntFree( vPaths );
    Vec_IntFree( vRefs );
    fmt.Printf("Computed windows for %d nodes (out of %d). Unique = %d. Ave inputs = %.2f. Ave outputs = %.2f. Ave volume = %.2f.  Gain = %d. ", nWins, Gia_ManAndNum(p), Hsh_VecSize(pHash), 1.0*nInsSize/Abc_MaxInt(1,nWins), 1.0*nOutSize/Abc_MaxInt(1,nWins), 1.0*nWinSize/Abc_MaxInt(1,nWins), nNodeGain);
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    Hsh_VecManStop( pHash );
}

/**Function*************************************************************

  Synopsis    [Apply k-resub to one AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_RsbTryOneWindow(p *Gia_Man_t) *Gia_Man_t {
    return Gia_ManResub2Test( p );
}

/**Function*************************************************************

  Synopsis    [Apply k-resub to one AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_RsbTestArray() {
    var Array [1000]int32 = [1000]int32{ 
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 6, 3, 7, 15, 17, 8, 19, 
        5, 20, 5, 12, 8, 24, 4, 12, 9, 28, 27, 31, 23, 32, 4, 13, 8, 36, 5, 
        13, 18, 40, 9, 18, 5, 44, 19, 36, 9, 48, 47, 51, 10, 18, 40, 54, 8, 
        56, 25, 37, 44, 61, 59, 63, 8, 28, 8, 18, 25, 68, 66, 70, 64, 73, 11, 
        19, 8, 13, 76, 78, 10, 19, 40, 82, 9, 84, 81, 87, 20, 61, 19, 28, 30, 
        92, 91, 95, 88, 96, 74, 98, 9, 40, 49, 103, 27, 104, 10, 107, 8, 40, 
        9, 24, 111, 113, 11, 115, 109, 117, 11, 66, 51, 121, 118, 122, 18, 36, 
        18, 110, 93, 127, 10, 131, 129, 133, 11, 38, 32, 137, 103, 138, 19, 141, 
        134, 143, 28, 76, 9, 146, 11, 110, 19, 150, 149, 153, 87, 95, 9, 19, 10, 
        159, 61, 160, 18, 30, 61, 158, 9, 12, 25, 169, 19, 171, 111, 173, 10, 175, 
        167, 177, 18, 102, 4, 20, 18, 171, 183, 185, 11, 187, 181, 189, 178, 190, 
        24, 44, 11, 194, 8, 54, 4, 198, 197, 201, 45, 49, 10, 39, 9, 126, 73, 209, 
        11, 211, 54, 168, 213, 215, 43, 167, 67, 218, 10, 221, 26, 54, 18, 18, 34, 
        34, 38, 38, 40, 40, 42, 42, 52, 52, 100, 100, 124, 124, 126, 126, 144, 144, 
        148, 148, 154, 154, 156, 156, 162, 162, 164, 164, 192, 192, 70, 70, 202, 
        202, 204, 204, 206, 206, 216, 216, 222, 222, 224, 224,
    };
    var i, iFan0, iFan1, nResubs int32;
    var pRes *int32;
    // create the internal array
    var vArray *Vec_Int_t  = Vec_IntAlloc( 100 );
    for  i = 0; i < 50 || Array[i] > 0; i++  {
        Vec_IntPush( vArray, Array[i] );
    }
    Vec_IntPrint( vArray );
    // check the nodes
    fmt.Printf("Constant0 and primary inputs:\n");
    // inlining of macro Vec_IntForEachEntryDouble
    for  i = 0; (i+1 < Vec_IntSize(vArray)) && (((iFan0) = Vec_IntEntry(vArray, i)), 1) && (((iFan1) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
        if iFan0 != iFan1 {
            break; }
        fmt.Printf("%2d = %c%2d & %c%2d;\n", i, Abc_LitIsCompl(iFan0) ? '!' : ' ', Abc_Lit2Var(iFan0), Abc_LitIsCompl(iFan1) ? '!' : ' ', Abc_Lit2Var(iFan1));
    }
    fmt.Printf("Primary outputs:\n");
    // inlining of macro Vec_IntForEachEntryDoubleStart
    for  i = 14; (i+1 < Vec_IntSize(vArray)) && (((iFan0) = Vec_IntEntry(vArray, i)), 1) && (((iFan1) = Vec_IntEntry(vArray, i+1)), 1); i += 2  {
        if iFan0 != iFan1 {
            continue; }
        fmt.Printf("%2d = %c%2d & %c%2d;\n", i, Abc_LitIsCompl(iFan0) ? '!' : ' ', Abc_Lit2Var(iFan0), Abc_LitIsCompl(iFan1) ? '!' : ' ', Abc_Lit2Var(iFan1));
    }
    // run the resub
    Abc_ResubPrepareManager( 1 );
    Abc_ResubComputeWindow( Vec_IntArray(vArray), Vec_IntSize(vArray)/2, 10, -1, 0, 0, 1, 1, &pRes, &nResubs );
    Abc_ResubPrepareManager( 0 );
    Vec_IntFree( vArray );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

