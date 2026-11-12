package wln

import "fmt"

/**CFile****************************************************************

  FileName    [wlnRetime.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Word-level network.]

  Synopsis    [Retiming.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - September 23, 2018.]

  Revision    [$Id: wlnRetime.c,v 1.00 2018/09/23 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "wln.h"
//#include "misc/vec/vecHsh.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Wln_Ret_t struct {
    pNtk        *Wln_Ntk_t;         // static netlist
    vFanins         Vec_Int_t;      // fanins and edge places
    vFanouts         Vec_Int_t;     // fanouts and edge places
    vEdgeLinks         Vec_Int_t;   // edge links
    vFfClasses         Vec_Int_t;   // flop classes
    vNodeDelays         Vec_Int_t;  // object delays
    vPathDelays         Vec_Int_t;  // delays from sources to sinks
    vSources         Vec_Int_t;     // critical sources
    vSinks         Vec_Int_t;       // critical sinks
    vFront         Vec_Int_t;       // retiming frontier
    vMoves         Vec_Int_t;       // retiming moves (paired with delay)
    nClasses               int32;     // the number of flop classes
    DelayMax               int32;     // critical delay at any time
}

func Wln_RetFanins(p *Wln_Ret_t, i int32) *int32 { return Vec_IntEntryP( &p.vFanins,  Vec_IntEntry(&p.vFanins, i) );  }
func Wln_RetFanouts(p *Wln_Ret_t, i int32) *int32 { return Vec_IntEntryP( &p.vFanouts, Vec_IntEntry(&p.vFanouts, i) ); }

// Original macro:
// #define Wln_RetForEachFanin( p, iObj, iFanin, pLink, i )    \
//     for ( i = 0; (i < Wln_ObjFaninNum(p->pNtk, iObj)) &&    \
//          (((iFanin) = Wln_RetFanins(p, iObj)[2*i]), 1) &&   \
//            ((pLink) = (Wln_RetFanins(p, iObj)+2*i+1)); i++ ) if ( !iFanin || (!Wln_ObjFaninNum(p->pNtk, iFanin) && !Wln_ObjIsCi(p->pNtk, iFanin)) )  {} else


// Original macro:
// #define Wln_RetForEachFanout( p, iObj, iFanout, pLink, i )  \
//     for ( i = 0; (i < Wln_ObjRefs(p->pNtk, iObj)) &&        \
//         (((iFanout) = Wln_RetFanouts(p, iObj)[2*i]), 1) &&  \
//            ((pLink) = Vec_IntEntryP(&p->vFanins, Wln_RetFanouts(p, iObj)[2*i+1])); i++ ) if ( !iFanout ) {} else


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Printing procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_RetPrintObj(p *Wln_Ret_t, iObj int32) {
    var k int32;
    var iFanin int32;
    var Type int32 = Wln_ObjType(p.pNtk, iObj);
    var pLink *int32;
    fmt.Printf("Obj %6d : Type = %6s  NameId = %5d  InstId = %5d  Fanins = %d : ", iObj, Abc_OperName(Type), Wln_ObjNameId(p.pNtk, iObj), Wln_ObjInstId(p.pNtk, iObj), Wln_ObjFaninNum(p.pNtk, iObj));
    // inlining of macro Wln_RetForEachFanin
    for  k = 0; (k < Wln_ObjFaninNum(p.pNtk, iObj)) &&             (((iFanin) = Wln_RetFanins(p, iObj)[2*k]), 1) &&              ((pLink) = (Wln_RetFanins(p, iObj)+2*k+1)); k++  { if !(iFanin != 0) || (!Wln_ObjFaninNum(p.pNtk, iFanin) && !Wln_ObjIsCi(p.pNtk, iFanin))  {} else {
        fmt.Printf("%5d ", iFanin);
        if !pLink[0] {
            continue; }
        fmt.Printf("(%d : %d %d) ", pLink[0], Vec_IntEntry(&p.vEdgeLinks, pLink[0]), Vec_IntEntry(&p.vEdgeLinks, pLink[0]+1));
    }
    }
    fmt.Printf("\n");
}
func Wln_RetPrint(p *Wln_Ret_t, fVerbose int32) {
    var iObj, nCount int32;
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p.pNtk); iObj++  {
        if Wln_ObjInstId(p.pNtk, iObj) > 1 {
            nCount++; }
    }
    fmt.Printf("Total number of objects = %d.  Objects with non-trivial delay = %d.\n", Wln_NtkObjNum(p.pNtk), nCount);
    if !(fVerbose != 0) {
        var nPrints int32;
        var nLimit int32 = 5;
        fmt.Printf("The following %d objects have non-trivial delays:\n", nLimit);
        // inlining of macro Wln_NtkForEachObj
        for  iObj = 1; iObj < Wln_NtkObjNum(p.pNtk); iObj++  {
            if Wln_ObjInstId(p.pNtk, iObj) <= 1 {
                continue; }
            Wln_RetPrintObj( p, iObj );
            if nPrints++; nPrints == nLimit {
                break; }
        }
        return;
    }
    fmt.Printf("Printing %d objects of network \"%s\":\n", Wln_NtkObjNum(p.pNtk), p.pNtk.pName);
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p.pNtk); iObj++  {
        Wln_RetPrintObj( p, iObj );
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Retiming manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_RetComputeFfClasses(pNtk *Wln_Ntk_t, vClasses *Vec_Int_t) int32 {
    var i, k, iObj, nClasses int32;
    var p *Hsh_VecMan_t  = Hsh_VecManStart( 10 );
    var vFlop *Vec_Int_t  = Vec_IntAlloc( 6 ); 
    Vec_IntFill( vClasses, Wln_NtkObjNum(pNtk), -1 );
    // inlining of macro Wln_NtkForEachFf
    for  i = 0; (i < Wln_NtkFfNum(pNtk)) && (((iObj) = Wln_NtkFf(pNtk, i)), 1); i++  {
        Vec_IntClear( vFlop );
        for  k = 1; k <= 6; k++  {
            Vec_IntPush( vFlop, Wln_ObjFanin(pNtk, iObj, k) );
        }
        Vec_IntWriteEntry( vClasses, iObj, Hsh_VecManAdd(p, vFlop) );
    }
    nClasses = Hsh_VecSize( p );
    Hsh_VecManStop( p );
    Vec_IntFree( vFlop );
    fmt.Printf("Detected %d flops and %d flop classes.\n", Wln_NtkFfNum(pNtk), nClasses);
    return nClasses;
}
func Wln_RetAlloc(pNtk *Wln_Ntk_t) *Wln_Ret_t {
     var k int32;
var iObj int32;
var iFanin int32;
var fFirst int32 = 1;
    var vRefsCopy *Vec_Int_t  = Vec_IntAlloc(0);
    // inlining of macro ABC_CALLOC
    p := &Wln_Ret_t{};
    p.pNtk = pNtk;
    Wln_NtkCreateRefs( pNtk );
    // print objects without fanout
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(pNtk); iObj++  {
        if Wln_ObjRefs(pNtk, iObj) == 0 && !Wln_ObjIsCio(pNtk, iObj) {
            if fFirst != 0 {
                fFirst = 0;
                fmt.Printf("Objects without fanout:\n");
            }
            Wln_ObjPrint(pNtk, iObj);
        }
    }
    // start fanin/fanout maps
    Wln_NtkStartFaninMap( pNtk, &p.vFanins, 2 );
    Wln_NtkStartFanoutMap( pNtk, &p.vFanouts, &pNtk.vRefs, 2 );
    // inlining of macro ABC_SWAP
    { var t Vec_Int_t  = *vRefsCopy; *vRefsCopy = pNtk.vRefs; pNtk.vRefs = t; };
    Wln_NtkCleanRefs( pNtk );
    Vec_IntGrow( &p.vEdgeLinks, 10*Wln_NtkFfNum(pNtk) );
    Vec_IntPushTwo( &p.vEdgeLinks, -1, -1 );
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(pNtk); iObj++  {
        // inlining of macro Wln_ObjForEachFanin
        for  k = 0; (k < Wln_ObjFaninNum(pNtk, iObj)) && (((iFanin) = Wln_ObjFanin(pNtk, iObj, k)), 1); k++  { if !(iFanin != 0) {} else {
            var pFanins *int32   = Wln_RetFanins( p, iObj );
            var pFanouts *int32  = Wln_RetFanouts( p, iFanin );
            var Index int32       = Wln_ObjRefsInc( pNtk, iFanin );
            pFanins[2*k+0]      = iFanin;
            pFanins[2*k+1]      = Wln_ObjIsFf(pNtk, iFanin) ? Vec_IntSize(&p.vEdgeLinks) : 0;
            pFanouts[2*Index+0] = iObj;
            pFanouts[2*Index+1] = Vec_IntEntry(&p.vFanins, iObj) + 2*k + 1;
            if Wln_ObjIsFf(pNtk, iFanin) {
                Vec_IntPushTwo( &p.vEdgeLinks, 0, iFanin ); }
        }
        }
    }
    // double-check the current number of fanouts added
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(pNtk); iObj++  {
        assert(Wln_ObjRefs(pNtk, iObj) == Vec_IntEntry(vRefsCopy, iObj), "Wln_ObjRefs(pNtk, iObj) == Vec_IntEntry(vRefsCopy, iObj)");
    }
    Vec_IntFree( vRefsCopy );
    // other data
    p.nClasses = Wln_RetComputeFfClasses( pNtk, &p.vFfClasses );
    //ABC_SWAP( Vec_Int_t, p->vNodeDelays, pNtk->vInstIds );
    Vec_IntAppend( &p.vNodeDelays, &pNtk.vInstIds );
    Vec_IntGrow( &p.vSources, 1000 );
    Vec_IntGrow( &p.vSinks, 1000 );
    Vec_IntGrow( &p.vFront, 1000 );
    Vec_IntGrow( &p.vMoves, 1000 );
    return p;
}
func Wln_RetFree(p *Wln_Ret_t) {
    p.vFanins.pArray = nil;
    p.vFanouts.pArray = nil;
    p.vEdgeLinks.pArray = nil;
    p.vFfClasses.pArray = nil;
    p.vNodeDelays.pArray = nil;
    p.vPathDelays.pArray = nil;
    p.vSources.pArray = nil;
    p.vSinks.pArray = nil;
    p.vFront.pArray = nil;
    p.vMoves.pArray = nil;
    p = nil;
}
func Wln_RetMemUsage(p *Wln_Ret_t) int32 {
    var Mem int32  = sizeof(Wln_Ret_t);
    Mem += 4 * p.vFanins.nCap;
    Mem += 4 * p.vFanouts.nCap;
    Mem += 4 * p.vEdgeLinks.nCap;
    Mem += 4 * p.vFfClasses.nCap;
    Mem += 4 * p.vNodeDelays.nCap;
    Mem += 4 * p.vPathDelays.nCap;
    Mem += 4 * p.vSources.nCap;
    Mem += 4 * p.vSinks.nCap;
    Mem += 4 * p.vFront.nCap;
    Mem += 4 * p.vMoves.nCap;
    return Mem;
}

/**Function*************************************************************

  Synopsis    [Delay propagation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_RetMarkChanges_rec(p *Wln_Ret_t, iObj int32) {
    var k int32;
    var iFanout int32;
    var pLink *int32;
    var pDelay *int32 = Vec_IntEntryP( &p.vPathDelays, iObj );
    if *pDelay < 0 {
        return; }
    *pDelay = -1;
    // inlining of macro Wln_RetForEachFanout
    for  k = 0; (k < Wln_ObjRefs(p.pNtk, iObj)) &&                (((iFanout) = Wln_RetFanouts(p, iObj)[2*k]), 1) &&             ((pLink) = Vec_IntEntryP(&p.vFanins, Wln_RetFanouts(p, iObj)[2*k+1])); k++  { if !(iFanout != 0) {} else
        if !pLink[0] {
            Wln_RetMarkChanges_rec( p, iFanout ); }
    }
}
func Wln_RetMarkChanges(p *Wln_Ret_t, vFront *Vec_Int_t) {
    var i, iObj int32;
    if vFront != nil {
        // inlining of macro Vec_IntForEachEntry
        for  i = 0; (i < Vec_IntSize(vFront)) && (((iObj) = Vec_IntEntry(vFront, i)), 1); i++  {
            Wln_RetMarkChanges_rec( p, iObj );
        }
    
    } else {
        Vec_IntFill( &p.vPathDelays, Wln_NtkObjNum(p.pNtk), -1 );
        // inlining of macro Wln_NtkForEachCi
        for  i = 0; (i < Wln_NtkCiNum(p.pNtk)) && (((iObj) = Wln_NtkCi(p.pNtk, i)), 1); i++  {
            Vec_IntWriteEntry( &p.vPathDelays, iObj, 0 );
        }
    }
}
func Wln_RetPropDelay_rec(p *Wln_Ret_t, iObj int32) int32 {
    var k int32;
    var iFanin int32;
    var pLink *int32;
    var pDelay *int32 = Vec_IntEntryP( &p.vPathDelays, iObj );
    if *pDelay >= 0 {
        return *pDelay; }
    *pDelay = 0;
    // inlining of macro Wln_RetForEachFanin
    for  k = 0; (k < Wln_ObjFaninNum(p.pNtk, iObj)) &&             (((iFanin) = Wln_RetFanins(p, iObj)[2*k]), 1) &&              ((pLink) = (Wln_RetFanins(p, iObj)+2*k+1)); k++  { if !(iFanin != 0) || (!Wln_ObjFaninNum(p.pNtk, iFanin) && !Wln_ObjIsCi(p.pNtk, iFanin))  {} else {
        if Wln_ObjIsFf(p.pNtk, iObj) && k > 0 {
            continue; }
        if pLink[0] {
            *pDelay = Abc_MaxInt(*pDelay, 0);
        } else {
            *pDelay = Abc_MaxInt(*pDelay, Wln_RetPropDelay_rec(p, iFanin)); }
    }
    }
    *pDelay += Vec_IntEntry( &p.vNodeDelays, iObj );
    return *pDelay;
}
func Wln_RetPropDelay(p *Wln_Ret_t) int32 {
    var iObj, DelayMax int32;
    Vec_IntClear( &p.vSinks );
    // inlining of macro Wln_NtkForEachObj
    for  iObj = 1; iObj < Wln_NtkObjNum(p.pNtk); iObj++  {
        if !Wln_ObjIsCio(p.pNtk, iObj) {
            var Delay int32  = Wln_RetPropDelay_rec(p, iObj);
            if DelayMax == Delay {
                Vec_IntPush( &p.vSinks, iObj );
            } else if DelayMax < Delay {
                DelayMax = Delay;
                Vec_IntFill( &p.vSinks, 1, iObj );
            }
        }
    }

//    Vec_IntForEachEntry( &p->vPathDelays, iObj, i )
//        printf( "Obj = %d.  Delay = %d.\n", i, iObj );
//    printf( "\n" );

//    printf( "Sinks: " );
//    Vec_IntPrint( &p->vSinks );

    return DelayMax;
}

func Wln_RetFindSources_rec(p *Wln_Ret_t, iObj int32) {
    var k int32;
    var iFanin int32;
    var pLink *int32;
    var FaninDelay int32;
    if Wln_ObjIsCi(p.pNtk, iObj) || Wln_ObjCheckTravId(p.pNtk, iObj) {
        return; }
    FaninDelay = Vec_IntEntry( &p.vPathDelays, iObj ) - Vec_IntEntry( &p.vNodeDelays, iObj );
    // inlining of macro Wln_RetForEachFanin
    for  k = 0; (k < Wln_ObjFaninNum(p.pNtk, iObj)) &&             (((iFanin) = Wln_RetFanins(p, iObj)[2*k]), 1) &&              ((pLink) = (Wln_RetFanins(p, iObj)+2*k+1)); k++  { if !(iFanin != 0) || (!Wln_ObjFaninNum(p.pNtk, iFanin) && !Wln_ObjIsCi(p.pNtk, iFanin))  {} else
        if !pLink[0] && Vec_IntEntry(&p.vPathDelays, iFanin) == FaninDelay {
            Wln_RetFindSources_rec( p, iFanin ); }
    }
    if FaninDelay == 0 {
        Vec_IntPush( &p.vSources, iObj ); }
}
func Wln_RetFindSources(p *Wln_Ret_t) {
    var i, iObj int32;
    Vec_IntClear( &p.vSources );
    Wln_NtkIncrementTravId( p.pNtk );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(&p.vSinks)) && (((iObj) = Vec_IntEntry(&p.vSinks, i)), 1); i++  {
        Wln_RetFindSources_rec( p, iObj );
    }

//    printf( "Sources: " );
//    Vec_IntPrint( &p->vSources );
}

/**Function*************************************************************

  Synopsis    [Retimability check.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_RetHeadToTail(p *Wln_Ret_t, pHead *int32) *int32 {
    var pLink *int32;
    assert(pHead[0], "pHead[0]");
    pLink = Vec_IntEntryP( &p.vEdgeLinks, pHead[0] );
    if pLink[0] == 0 {
        return pHead; }
    return Wln_RetHeadToTail( p, pLink );
}

func Wln_RetCheckForwardOne(p *Wln_Ret_t, iObj int32) int32 {
    var k int32;
    var iFanin int32;
    var pLink *int32;
    var iFlop int32;
    var Class int32 = -1;
    // inlining of macro Wln_RetForEachFanin
    for  k = 0; (k < Wln_ObjFaninNum(p.pNtk, iObj)) &&             (((iFanin) = Wln_RetFanins(p, iObj)[2*k]), 1) &&              ((pLink) = (Wln_RetFanins(p, iObj)+2*k+1)); k++  { if !(iFanin != 0) || (!Wln_ObjFaninNum(p.pNtk, iFanin) && !Wln_ObjIsCi(p.pNtk, iFanin))  {} else {
        if Wln_ObjIsFf(p.pNtk, iObj) && k > 0 {
            continue; }
        if !pLink[0] {
            return 0; }
        iFlop = Vec_IntEntry( &p.vEdgeLinks, pLink[0] + 1 );
        assert(Wln_ObjIsFf( p.pNtk, iFlop ), "Wln_ObjIsFf( p.pNtk, iFlop )");
        if Class == -1 {
            Class = Vec_IntEntry( &p.vFfClasses, iFlop );
        } else if Class != Vec_IntEntry( &p.vFfClasses, iFlop ) {
            return 0; }
    }
    }
    return 1;
}
func Wln_RetCheckForward(p *Wln_Ret_t, vSet *Vec_Int_t) int32 {
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSet)) && (((iObj) = Vec_IntEntry(vSet, i)), 1); i++  {
        if !(Wln_RetCheckForwardOne( p, iObj ) != 0) {
            return 0; }
    }
    return 1;
}

func Wln_RetCheckBackwardOne(p *Wln_Ret_t, iObj int32) int32 {
    var k int32;
    var iFanin int32;
    var pLink *int32;
    var iFlop int32;
    var Class int32 = -1;
    if Wln_ObjRefs(p.pNtk, iObj) == 0 {
        return 0; }
    // inlining of macro Wln_RetForEachFanout
    for  k = 0; (k < Wln_ObjRefs(p.pNtk, iObj)) &&                (((iFanin) = Wln_RetFanouts(p, iObj)[2*k]), 1) &&             ((pLink) = Vec_IntEntryP(&p.vFanins, Wln_RetFanouts(p, iObj)[2*k+1])); k++  { if !(iFanin != 0) {} else {
        if !pLink[0] {
            return 0; }
        pLink = Wln_RetHeadToTail( p, pLink );
        iFlop = Vec_IntEntry( &p.vEdgeLinks, pLink[0] + 1 );
        assert(Wln_ObjIsFf( p.pNtk, iFlop ), "Wln_ObjIsFf( p.pNtk, iFlop )");
        if Class == -1 {
            Class = Vec_IntEntry( &p.vFfClasses, iFlop );
        } else if Class != Vec_IntEntry( &p.vFfClasses, iFlop ) {
            return 0; }
    }
    }
    return 1;
}
func Wln_RetCheckBackward(p *Wln_Ret_t, vSet *Vec_Int_t) int32 {
    var i, iObj int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSet)) && (((iObj) = Vec_IntEntry(vSet, i)), 1); i++  {
        if !(Wln_RetCheckBackwardOne( p, iObj ) != 0) {
            return 0; }
    }
    return 1;
}


/**Function*************************************************************

  Synopsis    [Moving flops.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_RetRemoveOneFanin(p *Wln_Ret_t, iObj int32) int32 {
    var k int32;
    var iFanin int32;
    var pLink *int32;
    var iFlop int32;
    var iFlop1 int32 = -1;
    var pFanins *int32   = Wln_RetFanins( p, iObj );
    // inlining of macro Wln_RetForEachFanin
    for  k = 0; (k < Wln_ObjFaninNum(p.pNtk, iObj)) &&             (((iFanin) = Wln_RetFanins(p, iObj)[2*k]), 1) &&              ((pLink) = (Wln_RetFanins(p, iObj)+2*k+1)); k++  { if !(iFanin != 0) || (!Wln_ObjFaninNum(p.pNtk, iFanin) && !Wln_ObjIsCi(p.pNtk, iFanin))  {} else {
        if Wln_ObjIsFf(p.pNtk, iObj) && k > 0 {
            continue; }
        assert(pLink[0], "pLink[0]");
        iFlop = Vec_IntEntry( &p.vEdgeLinks, pLink[0] + 1 );
        pFanins[2*k+1] = Vec_IntEntry( &p.vEdgeLinks, pLink[0] );
        assert(Wln_ObjIsFf( p.pNtk, iFlop ), "Wln_ObjIsFf( p.pNtk, iFlop )");
        if iFlop1 == -1 {
            iFlop1 = iFlop; }
    }
    }
    return iFlop1;
}
func Wln_RetRemoveOneFanout(p *Wln_Ret_t, iObj int32) int32 {
    var k int32;
    var iFanin int32;
    var pLink *int32;
    var iFlop int32;
    var iFlop1 int32 = -1;
    //int * pFanins  = Wln_RetFanins( p, iObj );
    // inlining of macro Wln_RetForEachFanout
    for  k = 0; (k < Wln_ObjRefs(p.pNtk, iObj)) &&                (((iFanin) = Wln_RetFanouts(p, iObj)[2*k]), 1) &&             ((pLink) = Vec_IntEntryP(&p.vFanins, Wln_RetFanouts(p, iObj)[2*k+1])); k++  { if !(iFanin != 0) {} else {
        assert(pLink[0], "pLink[0]");
        pLink = Wln_RetHeadToTail( p, pLink );
        iFlop = Vec_IntEntry( &p.vEdgeLinks, pLink[0] + 1 );
        pLink[0] = 0;
        assert(Wln_ObjIsFf( p.pNtk, iFlop ), "Wln_ObjIsFf( p.pNtk, iFlop )");
        if iFlop1 == -1 {
            iFlop1 = iFlop; }
    }
    }
    return iFlop1;
}
func Wln_RetInsertOneFanin(p *Wln_Ret_t, iObj int32, iFlop int32) {
    var k int32;
    var iHead int32;
    var iFanin int32;
    var pLink *int32;
    var pFanins *int32   = Wln_RetFanins( p, iObj );
    assert(Wln_ObjIsFf( p.pNtk, iFlop ), "Wln_ObjIsFf( p.pNtk, iFlop )");
    // inlining of macro Wln_RetForEachFanin
    for  k = 0; (k < Wln_ObjFaninNum(p.pNtk, iObj)) &&             (((iFanin) = Wln_RetFanins(p, iObj)[2*k]), 1) &&              ((pLink) = (Wln_RetFanins(p, iObj)+2*k+1)); k++  { if !(iFanin != 0) || (!Wln_ObjFaninNum(p.pNtk, iFanin) && !Wln_ObjIsCi(p.pNtk, iFanin))  {} else {
        if Wln_ObjIsFf(p.pNtk, iObj) && k > 0 {
            continue; }
        iHead = pFanins[2*k+1];
        pFanins[2*k+1] = Vec_IntSize(&p.vEdgeLinks);
        Vec_IntPushTwo( &p.vEdgeLinks, iHead, iFlop );
    }
    }
}
func Wln_RetInsertOneFanout(p *Wln_Ret_t, iObj int32, iFlop int32) {
    var k int32;
    var iFanin int32;
    var pLink *int32;
    assert(Wln_ObjIsFf( p.pNtk, iFlop ), "Wln_ObjIsFf( p.pNtk, iFlop )");
    // inlining of macro Wln_RetForEachFanout
    for  k = 0; (k < Wln_ObjRefs(p.pNtk, iObj)) &&                (((iFanin) = Wln_RetFanouts(p, iObj)[2*k]), 1) &&             ((pLink) = Vec_IntEntryP(&p.vFanins, Wln_RetFanouts(p, iObj)[2*k+1])); k++  { if !(iFanin != 0) {} else {
        if pLink[0] {
            pLink = Wln_RetHeadToTail( p, pLink ); }
        //assert( pLink[0] == 0 );
        pLink[0] = Vec_IntSize(&p.vEdgeLinks);
        Vec_IntPushTwo( &p.vEdgeLinks, 0, iFlop );
    }
    }
}
func Wln_RetRetimeForward(p *Wln_Ret_t, vSet *Vec_Int_t) {
    var i, iObj, iFlop int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSet)) && (((iObj) = Vec_IntEntry(vSet, i)), 1); i++  {
        iFlop = Wln_RetRemoveOneFanin( p, iObj );
        if iFlop == -1 {
            continue; }
        Wln_RetInsertOneFanout( p, iObj, iFlop );
    }
}
func Wln_RetRetimeBackward(p *Wln_Ret_t, vSet *Vec_Int_t) {
    var i, iObj, iFlop int32;
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSet)) && (((iObj) = Vec_IntEntry(vSet, i)), 1); i++  {
        iFlop = Wln_RetRemoveOneFanout( p, iObj );
        if iFlop == -1 {
            continue; }
        Wln_RetInsertOneFanin( p, iObj, iFlop );
    }
}
func Wln_RetAddToMoves(p *Wln_Ret_t, vSet *Vec_Int_t, Delay int32, fForward int32, nMoves int32, fSkipSimple int32, fVerbose int32) {
    var i, iObj int32;
    if vSet == nil {
        fmt.Printf("Move %4d : Recording initial state     (delay = %6d)\n", nMoves, Delay);
        Vec_IntPushTwo( &p.vMoves, Delay, 0 );
        return;
    }
    fmt.Printf("Move %4d : Recording %s retiming (delay = %6d) :", nMoves, fForward ? "forward " : "backward", Delay);
    Vec_IntPush( &p.vMoves, Delay );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vSet)) && (((iObj) = Vec_IntEntry(vSet, i)), 1); i++  {
        var NameId int32  = Vec_IntEntry( &p.pNtk.vNameIds, iObj );
        if (fSkipSimple != 0) && (Wln_ObjIsFf(p.pNtk, iObj) || Wln_ObjType(p.pNtk, iObj) == ABC_OPER_SLICE || Wln_ObjType(p.pNtk, iObj) == ABC_OPER_CONCAT) {
            continue; }
        Vec_IntPush( &p.vMoves, fForward ? -NameId : NameId );
        if fVerbose != 0 {
            fmt.Printf(" %d (NameID = %d)  ", fForward ? -iObj : iObj, fForward ? -NameId : NameId); }
    }
    Vec_IntPush( &p.vMoves, 0 );
    if !(fVerbose != 0) {
        fmt.Printf(" %3d retimed objects", Vec_IntSize(vSet)); }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Retiming computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Wln_NtkRetimeCreateDelayInfo(pNtk *Wln_Ntk_t) {
    var i, iObj int32;
//    if ( Wln_NtkHasInstId(pNtk) )
//        Vec_IntErase( &pNtk->vInstIds );
    if Wln_NtkHasInstId(pNtk) {
        fmt.Printf("Using delays given by the user in the input file.\n");
        // inlining of macro Wln_NtkForEachObj
        for  iObj = 1; iObj < Wln_NtkObjNum(pNtk); iObj++  {
            if !Wln_ObjIsCio(pNtk, iObj) && !Wln_ObjIsConst(pNtk, iObj) && Wln_ObjInstId(pNtk, iObj) == 0 {
                fmt.Printf("Warning: Object %d of type %s has zero delay. Retiming will not work correctly.\n", iObj, Abc_OperName(Wln_ObjType(pNtk, iObj))); }
        }
    
    } else {
        fmt.Printf("The design has no delay information.\n");
        Wln_NtkCleanInstId(pNtk);
        // inlining of macro Wln_NtkForEachObj
        for  iObj = 1; iObj < Wln_NtkObjNum(pNtk); iObj++  {
            if Wln_ObjIsFf(pNtk, iObj) || Wln_ObjType(pNtk, iObj) == ABC_OPER_SLICE || Wln_ObjType(pNtk, iObj) == ABC_OPER_CONCAT {
                Wln_ObjSetInstId( pNtk, iObj, 1 );
            } else if !Wln_ObjIsCio(pNtk, iObj) && Wln_ObjFaninNum(pNtk, iObj) > 0 {
                Wln_ObjSetInstId( pNtk, iObj, 10 ); }
        }
        // inlining of macro Wln_NtkForEachCo
        for  i = 0; (i < Wln_NtkCoNum(pNtk)) && (((iObj) = Wln_NtkCo(pNtk, i)), 1); i++  {
            if Wln_ObjType(pNtk, Wln_ObjFanin0(pNtk, iObj)) != ABC_OPER_LUT {
                Wln_ObjSetInstId( pNtk, Wln_ObjFanin0(pNtk, iObj), 1 ); }
        }
        fmt.Printf("Assuming default delays: 10 units for most nodes and 1 unit for bit-slice, concat, and buffers driving COs.\n");
    }
}
func Wln_NtkRetime(pNtk *Wln_Ntk_t, fSkipSimple int32, fVerbose int32) *Vec_Int_t {
    var p *Wln_Ret_t  = Wln_RetAlloc( pNtk );
    var vSources *Vec_Int_t  = &p.vSources;
    var vSinks *Vec_Int_t    = &p.vSinks;
    var vFront *Vec_Int_t    = &p.vFront;
    var vMoves *Vec_Int_t    = Vec_IntAlloc(0);
    var nMoves, fPrevFwd, fPrevBwd, nCountIncrease int32;
    var DelayInit, DelayBest, nChange int32;
    Wln_RetPrint( p, fVerbose );
    Wln_RetMarkChanges( p, nil );
    p.DelayMax = DelayInit = DelayBest = Wln_RetPropDelay( p );
    Wln_RetFindSources( p );
    Wln_RetAddToMoves( p, nil, p.DelayMax, 0, nMoves, fSkipSimple, fVerbose );
    for  Vec_IntSize(vSources) || Vec_IntSize(vSinks)  {
        var DelayMaxPrev int32  = p.DelayMax;
        var fForward int32   = Vec_IntSize(vSources) && Wln_RetCheckForward( p, vSources );
        var fBackward int32  = Vec_IntSize(vSinks)   && Wln_RetCheckBackward( p, vSinks );
        Vec_IntSort( vSources, 0 );
        Vec_IntSort( vSinks, 0 );

        if !(fForward != 0) && !(fBackward != 0) {
            fmt.Printf("Cannot retime forward and backward.\n");
            break;
        }
        if Vec_IntTwoCountCommon(vSources, vSinks) {
            fmt.Printf("Cannot reduce delay by retiming.\n");
            break;
        }
        nMoves++;
        Vec_IntClear( vFront );
        if (fPrevFwd && (fForward != 0)) || (!(fPrevBwd && (fBackward != 0)) && (((fForward != 0) && !(fBackward != 0)) || ((fForward != 0) && (fBackward != 0) && Vec_IntSize(vSources) < Vec_IntSize(vSinks)))) {
            Vec_IntAppend( vFront, vSources );
            Wln_RetMarkChanges( p, vFront );
            Wln_RetRetimeForward( p, vFront );
            p.DelayMax = Wln_RetPropDelay( p );
            fForward = 1;
fBackward = 0;
            fPrevFwd = 1;
        
        } else {
            Vec_IntAppend( vFront, vSinks );
            Wln_RetRetimeBackward( p, vFront );
            Wln_RetMarkChanges( p, vFront );
            p.DelayMax = Wln_RetPropDelay( p );
            fForward = 0;
fBackward = 1;
            fPrevBwd = 1;
        }
        DelayBest = Abc_MinInt( DelayBest, p.DelayMax );
        //Wln_RetPrint( p );
        if fVerbose != 0 {
            fmt.Printf("\n"); }
        Wln_RetAddToMoves( p, vFront, p.DelayMax, fForward, nMoves, fSkipSimple, fVerbose );
        if fVerbose != 0 {
            fmt.Printf("Sinks: ");
            Vec_IntPrint( &p.vSinks );
            fmt.Printf("Sources: ");
            Vec_IntPrint( &p.vSources );
        }
        if p.DelayMax >= DelayMaxPrev {
            nCountIncrease++;
        } else {
            if nCountIncrease > 0 {
                nChange++; }
            nCountIncrease = 0;
        }
        if nCountIncrease > 3 {
            break; }
        if nChange > 5 {
            break; }
        Wln_RetFindSources( p );
        if 2*Vec_IntSize(&p.vEdgeLinks) > Vec_IntCap(&p.vEdgeLinks) {
            Vec_IntGrow( &p.vEdgeLinks, 4*Vec_IntSize(&p.vEdgeLinks) ); }
    }
    // inlining of macro ABC_SWAP
    { var t Vec_Int_t  = *vMoves; *vMoves = p.vMoves; p.vMoves = t; };
    Wln_RetFree( p );
    if fVerbose != 0 {
        fmt.Printf("\nThe resulting moves recorded in terms of name IDs of the NDR nodes:\n");
        Vec_IntPrint( vMoves );
    
    } else {
        fmt.Printf("Retiming instruction contains %d moves and %d total retimed objects.\n", nMoves, Vec_IntSize(vMoves)-2*nMoves-2);
        fmt.Printf("Initial delay = %d.  The best delay achieved = %d.  Improvement = %d. (%6.2f %%)\n", DelayInit, DelayBest, DelayInit - DelayBest, 100.0 * (DelayInit - DelayBest) / DelayInit);
    }
    return vMoves;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

