package dar

import "fmt"

/**CFile****************************************************************

  FileName    [darLib.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [Library of AIG subgraphs used for rewriting.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: darLib.c,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "darInt.h"
//#include "aig/gia/gia.h"
//#include "dar.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Dar_Lib_t =            Dar_Lib_t_;
type Dar_LibObj_t =         Dar_LibObj_t_;
type Dar_LibDat_t =         Dar_LibDat_t_;

type Dar_LibObj_t_ struct { // library object (2 words)

    Fan0 uint32; /* C bitfield: 16 bits; review packing and truncation */  // the first fanin
    Fan1 uint32; /* C bitfield: 16 bits; review packing and truncation */  // the second fanin
    fCompl0 uint32; /* C bitfield: 1 bits; review packing and truncation */  // the first compl attribute
    fCompl1 uint32; /* C bitfield: 1 bits; review packing and truncation */  // the second compl attribute
    fPhase uint32; /* C bitfield: 1 bits; review packing and truncation */  // the phase of the node
    fTerm uint32; /* C bitfield: 1 bits; review packing and truncation */  // indicates a PI
    Num uint32; /* C bitfield: 28 bits; review packing and truncation */  // internal use
}

type Dar_LibDat_t_ struct { // library object data

    union { 
    Aig_Obj_t *      pFunc;         // the corresponding AIG node if it exists
    int              iGunc; };      // the corresponding AIG node if it exists
    Level              int32;         // level of this node after it is constructured
    TravId              int32;        // traversal ID of the library object data
    dProb            float32;         // probability of the node being 1
    unsigned char    fMffc;         // set to one if node is part of MFFC
    unsigned char    nLats[3];      // the number of latches on the input/output stem
}

type Dar_Lib_t_ struct { // library 

    // objects
    pObjs    *Dar_LibObj_t;         // the set of library objects
    nObjs              int32;         // the number of objects used
    iObj              int32;          // the current object
    // structures by class
    nSubgr [222]int32;   // the number of subgraphs by class
    pSubgr [222]*int32;   // the subgraphs for each class
    pSubgrMem             *int32;     // memory for subgraph pointers
    nSubgrTotal              int32;   // the total number of subgraph
    // structure priorities
    pPriosMem             *int32;     // memory for priority of structures
    pPrios [222]*int32;   // pointers to the priority numbers
    // structure places in the priorities
    pPlaceMem             *int32;     // memory for places of structures in the priority lists
    pPlace [222]*int32;   // pointers to the places numbers
    // structure scores
    pScoreMem             *int32;     // memory for scores of structures
    pScore [222]*int32;   // pointers to the scores numbers
    // nodes by class
    nNodes [222]int32;   // the number of nodes by class
    pNodes [222]*int32;   // the nodes for each class
    pNodesMem             *int32;     // memory for nodes pointers
    nNodesTotal              int32;   // the total number of nodes
    // prepared library
    nSubgraphs              int32;
    nNodes0Max              int32;
    // nodes by class
    nNodes0 [222]int32;   // the number of nodes by class
    pNodes0 [222]*int32;   // the nodes for each class
    pNodes0Mem             *int32;     // memory for nodes pointers
    nNodes0Total              int32;   // the total number of nodes
    // structures by class
    nSubgr0 [222]int32;   // the number of subgraphs by class
    pSubgr0 [222]*int32;   // the subgraphs for each class
    pSubgr0Mem             *int32;     // memory for subgraph pointers
    nSubgr0Total              int32;   // the total number of subgraph
    // object data
    pDatas    *Dar_LibDat_t;
    nDatas              int32;
    // information about NPN classes
    pPerms4           *string;
    unsigned short * puCanons; 
    pPhases            string; 
    pPerms            string; 
    unsigned char *  pMap;
}

var s_DarLib *Dar_Lib_t = nil;

func Dar_LibObj(p *Dar_Lib_t, Id int32) *Dar_LibObj_t { return p.pObjs + Id; }
func Dar_LibObjTruth(pObj *Dar_LibObj_t) int32 { if pObj.Num < (0xFFFF & ~pObj.Num) {
    return pObj.Num;
} else {
    return (0xFFFF & ~pObj.Num);
} }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibAlloc(nObjs int32) *Dar_Lib_t {
    var uTruths [4]uint32 = [4]uint32{ 0xAAAA, 0xCCCC, 0xF0F0, 0xFF00 };
    var p *Dar_Lib_t;
    var i int32;//, clk = Abc_Clock();
    // inlining of macro ABC_ALLOC
    p = ((Dar_Lib_t *) malloc(sizeof(Dar_Lib_t) * (size_t)(1)));
    memset( p, 0, sizeof(Dar_Lib_t) );
    // allocate objects
    p.nObjs = nObjs;
    // inlining of macro ABC_ALLOC
    p.pObjs = ((Dar_LibObj_t *) malloc(sizeof(Dar_LibObj_t) * (size_t)(nObjs)));
    memset( p.pObjs, 0, sizeof(Dar_LibObj_t) * nObjs );
    // allocate canonical data
    p.pPerms4 = Dar_Permutations( 4 );
    Dar_Truth4VarNPN( &p.puCanons, &p.pPhases, &p.pPerms, &p.pMap );
    // start the elementary objects
    p.iObj = 4;
    for  i = 0; i < 4; i++  {
        p.pObjs[i].fTerm = 1;
        p.pObjs[i].Num = uTruths[i];
    }
//    ABC_PRT( "Library start", Abc_Clock() - clk );
    return p;
}

/**Function*************************************************************

  Synopsis    [Frees the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibFree(p *Dar_Lib_t) {
    // inlining of macro ABC_FREE
    ((p.pObjs) ? (free((char *) (p.pObjs)), (p.pObjs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pDatas) ? (free((char *) (p.pDatas)), (p.pDatas) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pNodesMem) ? (free((char *) (p.pNodesMem)), (p.pNodesMem) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pNodes0Mem) ? (free((char *) (p.pNodes0Mem)), (p.pNodes0Mem) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSubgrMem) ? (free((char *) (p.pSubgrMem)), (p.pSubgrMem) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pSubgr0Mem) ? (free((char *) (p.pSubgr0Mem)), (p.pSubgr0Mem) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPriosMem) ? (free((char *) (p.pPriosMem)), (p.pPriosMem) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPlaceMem) ? (free((char *) (p.pPlaceMem)), (p.pPlaceMem) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pScoreMem) ? (free((char *) (p.pScoreMem)), (p.pScoreMem) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPerms4) ? (free((char *) (p.pPerms4)), (p.pPerms4) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.puCanons) ? (free((char *) (p.puCanons)), (p.puCanons) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPhases) ? (free((char *) (p.pPhases)), (p.pPhases) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pPerms) ? (free((char *) (p.pPerms)), (p.pPerms) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pMap) ? (free((char *) (p.pMap)), (p.pMap) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Returns canonical truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibReturnClass(uTruth uint32) int32 {
    return s_DarLib.pMap[uTruth & 0xffff];
}


/**Function*************************************************************

  Synopsis    [Returns canonical truth tables.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibReturnCanonicals(pCanons *uint32) {
    var Visits [222]int32 ;
    var i, k int32;
    // find canonical truth tables
    for  i = k = 0; i < (1<<16); i++  {
        if !Visits[s_DarLib.pMap[i]] {
            Visits[s_DarLib.pMap[i]] = 1;
            pCanons[k++] = ((i<<16) | i);
        }
    }
    assert(k == 222, "k == 222");
}

/**Function*************************************************************

  Synopsis    [Adds one AND to the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibAddNode(p *Dar_Lib_t, Id0 int32, Id1 int32, fCompl0 int32, fCompl1 int32) {
    var pFan0 *Dar_LibObj_t  = Dar_LibObj( p, Id0 );
    var pFan1 *Dar_LibObj_t  = Dar_LibObj( p, Id1 );
    var pObj *Dar_LibObj_t   = p.pObjs + p.iObj++;
    pObj.Fan0 = Id0;
    pObj.Fan1 = Id1;
    pObj.fCompl0 = fCompl0;
    pObj.fCompl1 = fCompl1;
    pObj.fPhase = (fCompl0 ^ pFan0.fPhase) & (fCompl1 ^ pFan1.fPhase);
    pObj.Num = 0xFFFF & (fCompl0? ~pFan0.Num : pFan0.Num) & (fCompl1? ~pFan1.Num : pFan1.Num);
}

/**Function*************************************************************

  Synopsis    [Adds one AND to the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibSetup_rec(p *Dar_Lib_t, pObj *Dar_LibObj_t, Class int32, fCollect int32) {
    if pObj.fTerm || int32(pObj.Num) == Class {
        return; }
    pObj.Num = Class;
    Dar_LibSetup_rec( p, Dar_LibObj(p, pObj.Fan0), Class, fCollect );
    Dar_LibSetup_rec( p, Dar_LibObj(p, pObj.Fan1), Class, fCollect );
    if fCollect != 0 {
        p.pNodes[Class][ p.nNodes[Class]++ ] = pObj-p.pObjs;
    } else {
        p.nNodes[Class]++; }
}

/**Function*************************************************************

  Synopsis    [Adds one AND to the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibSetup(p *Dar_Lib_t, vOuts *Vec_Int_t, vPrios *Vec_Int_t) {
    var fTraining int32;
    var pObj *Dar_LibObj_t;
    var nNodesTotal, uTruth, Class, Out, i, k int32;
    assert(p.iObj == p.nObjs, "p.iObj == p.nObjs");

    // count the number of representatives of each class
    for  i = 0; i < 222; i++  {
        p.nSubgr[i] = p.nNodes[i] = 0;
    }
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOuts)) && (((Out) = Vec_IntEntry(vOuts, i)), 1); i++  {
        pObj = Dar_LibObj( p, Out );
        uTruth = Dar_LibObjTruth( pObj );
        Class = p.pMap[uTruth];
        p.nSubgr[Class]++;
    }
    // allocate memory for the roots of each class
    // inlining of macro ABC_ALLOC
    p.pSubgrMem = ((int *) malloc(sizeof(int) * (size_t)(Vec_IntSize(vOuts))));
    // inlining of macro ABC_ALLOC
    p.pSubgr0Mem = ((int *) malloc(sizeof(int) * (size_t)(Vec_IntSize(vOuts))));
    p.nSubgrTotal = 0;
    for  i = 0; i < 222; i++  {
        p.pSubgr[i] = p.pSubgrMem + p.nSubgrTotal;
        p.pSubgr0[i] = p.pSubgr0Mem + p.nSubgrTotal;
        p.nSubgrTotal += p.nSubgr[i];
        p.nSubgr[i] = 0;
    }
    assert(p.nSubgrTotal == Vec_IntSize(vOuts), "p.nSubgrTotal == Vec_IntSize(vOuts)");
    // add the outputs to storage
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOuts)) && (((Out) = Vec_IntEntry(vOuts, i)), 1); i++  {
        pObj = Dar_LibObj( p, Out );
        uTruth = Dar_LibObjTruth( pObj );
        Class = p.pMap[uTruth];
        p.pSubgr[Class][ p.nSubgr[Class]++ ] = Out;
    }

    if fTraining != 0 {
        // allocate memory for the priority of roots of each class
        // inlining of macro ABC_ALLOC
        p.pPriosMem = ((int *) malloc(sizeof(int) * (size_t)(Vec_IntSize(vOuts))));
        p.nSubgrTotal = 0;
        for  i = 0; i < 222; i++  {
            p.pPrios[i] = p.pPriosMem + p.nSubgrTotal;
            p.nSubgrTotal += p.nSubgr[i];
            for  k = 0; k < p.nSubgr[i]; k++  {
                p.pPrios[i][k] = k;
            }

        }
        assert(p.nSubgrTotal == Vec_IntSize(vOuts), "p.nSubgrTotal == Vec_IntSize(vOuts)");

        // allocate memory for the priority of roots of each class
        // inlining of macro ABC_ALLOC
        p.pPlaceMem = ((int *) malloc(sizeof(int) * (size_t)(Vec_IntSize(vOuts))));
        p.nSubgrTotal = 0;
        for  i = 0; i < 222; i++  {
            p.pPlace[i] = p.pPlaceMem + p.nSubgrTotal;
            p.nSubgrTotal += p.nSubgr[i];
            for  k = 0; k < p.nSubgr[i]; k++  {
                p.pPlace[i][k] = k;
            }

        }
        assert(p.nSubgrTotal == Vec_IntSize(vOuts), "p.nSubgrTotal == Vec_IntSize(vOuts)");

        // allocate memory for the priority of roots of each class
        // inlining of macro ABC_ALLOC
        p.pScoreMem = ((int *) malloc(sizeof(int) * (size_t)(Vec_IntSize(vOuts))));
        p.nSubgrTotal = 0;
        for  i = 0; i < 222; i++  {
            p.pScore[i] = p.pScoreMem + p.nSubgrTotal;
            p.nSubgrTotal += p.nSubgr[i];
            for  k = 0; k < p.nSubgr[i]; k++  {
                p.pScore[i][k] = 0;
            }

        }
        assert(p.nSubgrTotal == Vec_IntSize(vOuts), "p.nSubgrTotal == Vec_IntSize(vOuts)");
    
    } else {
        var Counter int32;
        // allocate memory for the priority of roots of each class
        // inlining of macro ABC_ALLOC
        p.pPriosMem = ((int *) malloc(sizeof(int) * (size_t)(Vec_IntSize(vOuts))));
        p.nSubgrTotal = 0;
        for  i = 0; i < 222; i++  {
            p.pPrios[i] = p.pPriosMem + p.nSubgrTotal;
            p.nSubgrTotal += p.nSubgr[i];
            for  k = 0; k < p.nSubgr[i]; k++  {
                p.pPrios[i][k] = Vec_IntEntry(vPrios, Counter++);
            }

        }
        assert(p.nSubgrTotal == Vec_IntSize(vOuts), "p.nSubgrTotal == Vec_IntSize(vOuts)");
        assert(Counter == Vec_IntSize(vPrios), "Counter == Vec_IntSize(vPrios)");
    }

    // create traversal IDs
    for  i = 0; i < p.iObj; i++  {
        Dar_LibObj(p, i).Num = 0xff;
    }
    // count nodes in each class
    for  i = 0; i < 222; i++  {
        for  k = 0; k < p.nSubgr[i]; k++  {
            Dar_LibSetup_rec( p, Dar_LibObj(p, p.pSubgr[i][k]), i, 0 );
        }
    }
    // count the total number of nodes
    p.nNodesTotal = 0;
    for  i = 0; i < 222; i++  {
        p.nNodesTotal += p.nNodes[i];
    }
    // allocate memory for the nodes of each class
    // inlining of macro ABC_ALLOC
    p.pNodesMem = ((int *) malloc(sizeof(int) * (size_t)(p.nNodesTotal)));
    // inlining of macro ABC_ALLOC
    p.pNodes0Mem = ((int *) malloc(sizeof(int) * (size_t)(p.nNodesTotal)));
    p.nNodesTotal = 0;
    for  i = 0; i < 222; i++  {
        p.pNodes[i] = p.pNodesMem + p.nNodesTotal;
        p.pNodes0[i] = p.pNodes0Mem + p.nNodesTotal;
        p.nNodesTotal += p.nNodes[i];
        p.nNodes[i] = 0;
    }
    // create traversal IDs
    for  i = 0; i < p.iObj; i++  {
        Dar_LibObj(p, i).Num = 0xff;
    }
    // add the nodes to storage
    nNodesTotal = 0;
    for  i = 0; i < 222; i++  {
         for  k = 0; k < p.nSubgr[i]; k++  {
            Dar_LibSetup_rec( p, Dar_LibObj(p, p.pSubgr[i][k]), i, 1 );
         }
         nNodesTotal += p.nNodes[i];
//printf( "Class %3d : Subgraphs = %4d. Nodes = %5d.\n", i, p->nSubgr[i], p->nNodes[i] );
    }
    assert(nNodesTotal == p.nNodesTotal, "nNodesTotal == p.nNodesTotal");
     // prepare the number of the PI nodes
    for  i = 0; i < 4; i++  {
        Dar_LibObj(p, i).Num = i;
    }
}

/**Function*************************************************************

  Synopsis    [Starts the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibCreateData(p *Dar_Lib_t, nDatas int32) {
    if p.nDatas == nDatas {
        return; }
    // inlining of macro ABC_FREE
    ((p.pDatas) ? (free((char *) (p.pDatas)), (p.pDatas) = 0) : 0);
    // allocate datas
    p.nDatas = nDatas;
    // inlining of macro ABC_ALLOC
    p.pDatas = ((Dar_LibDat_t *) malloc(sizeof(Dar_LibDat_t) * (size_t)(nDatas)));
    memset( p.pDatas, 0, sizeof(Dar_LibDat_t) * nDatas );
}

/**Function*************************************************************

  Synopsis    [Adds one AND to the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibSetup0_rec(p *Dar_Lib_t, pObj *Dar_LibObj_t, Class int32, fCollect int32) {
    if pObj.fTerm || int32(pObj.Num) == Class {
        return; }
    pObj.Num = Class;
    Dar_LibSetup0_rec( p, Dar_LibObj(p, pObj.Fan0), Class, fCollect );
    Dar_LibSetup0_rec( p, Dar_LibObj(p, pObj.Fan1), Class, fCollect );
    if fCollect != 0 {
        p.pNodes0[Class][ p.nNodes0[Class]++ ] = pObj-p.pObjs;
    } else {
        p.nNodes0[Class]++; }
}

/**Function*************************************************************

  Synopsis    [Starts the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibPrepare(nSubgraphs int32) {
    var p *Dar_Lib_t  = s_DarLib;
    var i, k, nNodes0Total int32;
    if p.nSubgraphs == nSubgraphs {
        return; }

    // favor special classes:
    //  1 : F = (!d*!c*!b*!a)
    //  4 : F = (!d*!c*!(b*a))
    // 12 : F = (!d*!(c*!(!b*!a)))
    // 20 : F = (!d*!(c*b*a))

    // set the subgraph counters 
    p.nSubgr0Total = 0;
    for  i = 0; i < 222; i++  {
//        if ( i == 1 || i == 4 || i == 12 || i == 20 ) // special classes 
        if i == 1 { // special classes 
            p.nSubgr0[i] = p.nSubgr[i];
        } else {
            p.nSubgr0[i] = Abc_MinInt( p.nSubgr[i], nSubgraphs ); }
        p.nSubgr0Total += p.nSubgr0[i];
        for  k = 0; k < p.nSubgr0[i]; k++  {
            p.pSubgr0[i][k] = p.pSubgr[i][ p.pPrios[i][k] ];
        }
    }

    // count the number of nodes
    // clean node counters
    for  i = 0; i < 222; i++  {
        p.nNodes0[i] = 0;
    }
    // create traversal IDs
    for  i = 0; i < p.iObj; i++  {
        Dar_LibObj(p, i).Num = 0xff;
    }
    // count nodes in each class
    // count the total number of nodes and the largest class
    p.nNodes0Total = 0;
    p.nNodes0Max = 0;
    for  i = 0; i < 222; i++  {
        for  k = 0; k < p.nSubgr0[i]; k++  {
            Dar_LibSetup0_rec( p, Dar_LibObj(p, p.pSubgr0[i][k]), i, 0 );
        }
        p.nNodes0Total += p.nNodes0[i];
        p.nNodes0Max = Abc_MaxInt( p.nNodes0Max, p.nNodes0[i] );
    }

    // clean node counters
    for  i = 0; i < 222; i++  {
        p.nNodes0[i] = 0;
    }
    // create traversal IDs
    for  i = 0; i < p.iObj; i++  {
        Dar_LibObj(p, i).Num = 0xff;
    }
    // add the nodes to storage
    nNodes0Total = 0;
    for  i = 0; i < 222; i++  {
        for  k = 0; k < p.nSubgr0[i]; k++  {
            Dar_LibSetup0_rec( p, Dar_LibObj(p, p.pSubgr0[i][k]), i, 1 );
        }
         nNodes0Total += p.nNodes0[i];
    }
    assert(nNodes0Total == p.nNodes0Total, "nNodes0Total == p.nNodes0Total");
     // prepare the number of the PI nodes
    for  i = 0; i < 4; i++  {
        Dar_LibObj(p, i).Num = i;
    }

    // realloc the datas
    Dar_LibCreateData( p, p.nNodes0Max + 32 ); 
    // allocated more because Dar_LibBuildBest() sometimes requires more entries
}

/**Function*************************************************************

  Synopsis    [Reads library from array.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibRead() *Dar_Lib_t {
    var vObjs, vOuts, vPrios *Vec_Int_t;
    var p *Dar_Lib_t;
    var i int32;
    // read nodes and outputs
    vObjs  = Dar_LibReadNodes();
    vOuts  = Dar_LibReadOuts();
    vPrios = Dar_LibReadPrios();
    // create library
    p = Dar_LibAlloc( Vec_IntSize(vObjs)/2 + 4 );
    // create nodes
    for  i = 0; i < vObjs.nSize; i += 2  {
        Dar_LibAddNode( p, vObjs.pArray[i] >> 1, vObjs.pArray[i+1] >> 1,
            vObjs.pArray[i] & 1, vObjs.pArray[i+1] & 1 );
    }
    // create outputs
    Dar_LibSetup( p, vOuts, vPrios );
    Vec_IntFree( vObjs );
    Vec_IntFree( vOuts );
    Vec_IntFree( vPrios );
    return p;
}

/**Function*************************************************************

  Synopsis    [Starts the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibStart() {
//    abctime clk = Abc_Clock();
    if s_DarLib != nil {
        return; }
    assert(s_DarLib == nil, "s_DarLib == nil");
    s_DarLib = Dar_LibRead();
//    printf( "The 4-input library started with %d nodes and %d subgraphs. ", s_DarLib->nObjs - 4, s_DarLib->nSubgrTotal );
//    ABC_PRT( "Time", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    [Stops the library.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibStop() {
    assert(s_DarLib != nil, "s_DarLib != nil");
    Dar_LibFree( s_DarLib );
    s_DarLib = nil;
}

/**Function*************************************************************

  Synopsis    [Updates the score of the class and adjusts the priority of this class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibIncrementScore(Class int32, Out int32, Gain int32) {
    var pPrios *int32  = s_DarLib.pPrios[Class];  // pPrios[i] = Out
    var pPlace *int32  = s_DarLib.pPlace[Class];  // pPlace[Out] = i
    var pScore *int32  = s_DarLib.pScore[Class];  // score of Out
    var Out2 int32;
    assert(Class >= 0 && Class < 222, "Class >= 0 && Class < 222");
    assert(Out >= 0 && Out < s_DarLib.nSubgr[Class], "Out >= 0 && Out < s_DarLib.nSubgr[Class]");
    assert(pPlace[pPrios[Out]] == Out, "pPlace[pPrios[Out]] == Out");
    // increment the score
    pScore[Out] += Gain;
    // move the out in the order 
    for  pPlace[Out] > 0 && pScore[Out] > pScore[ pPrios[pPlace[Out]-1] ]  {
        // get the previous output in the priority list
        Out2 = pPrios[pPlace[Out]-1];
        // swap Out and Out2
        pPlace[Out]--;
        pPlace[Out2]++;
        pPrios[pPlace[Out]] = Out;
        pPrios[pPlace[Out2]] = Out2;
    }
}

/**Function*************************************************************

  Synopsis    [Prints out the priorities into the file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibDumpPriorities() {
    var i, k, Out, Out2, Counter, Printed int32;
    fmt.Printf("\nOutput priorities (total = %d):\n", s_DarLib.nSubgrTotal);
    for  i = 0; i < 222; i++  {
//        printf( "Class%d: ", i );
        for  k = 0; k < s_DarLib.nSubgr[i]; k++  {
            Out = s_DarLib.pPrios[i][k];
            if k == 0 {
                Out2 = Out;
            } else {
                Out2 = s_DarLib.pPrios[i][k-1];
            }
            assert(s_DarLib.pScore[i][Out2] >= s_DarLib.pScore[i][Out], "s_DarLib.pScore[i][Out2] >= s_DarLib.pScore[i][Out]");
//            printf( "%d(%d), ", Out, s_DarLib->pScore[i][Out] );
            fmt.Printf("%d, ", Out);
            Printed++;
            if Counter++; Counter == 15 {
                fmt.Printf("\n");
                Counter = 0;
            }
        }
    }
    fmt.Printf("\n");
    assert(Printed == s_DarLib.nSubgrTotal, "Printed == s_DarLib.nSubgrTotal");
}


/**Function*************************************************************

  Synopsis    [Matches the cut with its canonical form.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibCutMatch(p *Dar_Man_t, pCut *Dar_Cut_t) int32 {
    var pFanin *Aig_Obj_t;
    var uPhase uint32;
    var pPerm string;
    var i int32;
    assert(pCut.nLeaves == 4, "pCut.nLeaves == 4");
    // get the fanin permutation
    uPhase = s_DarLib.pPhases[pCut.uTruth];
    pPerm = s_DarLib.pPerms4[ int32(s_DarLib.pPerms[pCut.uTruth]) ];
    // collect fanins with the corresponding permutation/phase
    for  i = 0; i < int32(pCut.nLeaves); i++  {
        pFanin = Aig_ManObj( p.pAig, pCut.pLeaves[ int32(pPerm[i]) ] );
        if pFanin == nil {
            p.nCutsBad++;
            return 0;
        }
        pFanin = Aig_NotCond(pFanin, ((uPhase >> i) & 1) );
        s_DarLib.pDatas[i].pFunc = pFanin;
        s_DarLib.pDatas[i].Level = Aig_Regular(pFanin).Level;
        // copy the propability of node being one
        if p.pPars.fPower {
            var Prob float32  = Abc_Int2Float( Vec_IntEntry( p.pAig.vProbs, Aig_ObjId(Aig_Regular(pFanin)) ) );
            s_DarLib.pDatas[i].dProb = Aig_IsComplement(pFanin)? 1.0-Prob : Prob;
        }
    }
    p.nCutsGood++;
    return 1;
}



/**Function*************************************************************

  Synopsis    [Marks the MFFC of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibCutMarkMffc(p *Aig_Man_t, pRoot *Aig_Obj_t, nLeaves int32, pPower *float32) int32 {
    var i, nNodes int32;
    // mark the cut leaves
    for  i = 0; i < nLeaves; i++  {
        Aig_Regular(s_DarLib.pDatas[i].pFunc).nRefs++;
    }
    // label MFFC with current ID
    nNodes = Aig_NodeMffcLabel( p, pRoot, pPower );
    // unmark the cut leaves
    for  i = 0; i < nLeaves; i++  {
        Aig_Regular(s_DarLib.pDatas[i].pFunc).nRefs--;
    }
    return nNodes;
}

/**Function*************************************************************

  Synopsis    [Evaluates one cut.]

  Description [Returns the best gain.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibObjPrint_rec(pObj *Dar_LibObj_t) {
    if pObj.fTerm {
        fmt.Printf("%c", 'a' + int32((pObj - s_DarLib.pObjs)));
        return;
    }
    fmt.Printf("(");
    Dar_LibObjPrint_rec( Dar_LibObj(s_DarLib, pObj.Fan0) );
    if pObj.fCompl0 {
        fmt.Printf("\'"); }
    Dar_LibObjPrint_rec( Dar_LibObj(s_DarLib, pObj.Fan1) );
    if pObj.fCompl0 {
        fmt.Printf("\'"); }
    fmt.Printf(")");
}


/**Function*************************************************************

  Synopsis    [Assigns numbers to the nodes of one class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibEvalAssignNums(p *Dar_Man_t, Class int32, pRoot *Aig_Obj_t) {
    var pObj *Dar_LibObj_t;
    var pData, pData0, pData1 *Dar_LibDat_t;
    var pFanin0, pFanin1 *Aig_Obj_t;
    var i int32;
    for  i = 0; i < s_DarLib.nNodes0[Class]; i++  {
        // get one class node, assign its temporary number and set its data
        pObj = Dar_LibObj(s_DarLib, s_DarLib.pNodes0[Class][i]);
        pObj.Num = 4 + i;
        assert(int32(pObj.Num) < s_DarLib.nNodes0Max + 4, "int32(pObj.Num) < s_DarLib.nNodes0Max + 4");
        pData = s_DarLib.pDatas + pObj.Num;
        pData.fMffc = 0;
        pData.pFunc = nil;
        pData.TravId = 0xFFFF;

        // explore the fanins
        assert(int32(Dar_LibObj(s_DarLib, pObj.Fan0).Num) < s_DarLib.nNodes0Max + 4, "int32(Dar_LibObj(s_DarLib, pObj.Fan0).Num) < s_DarLib.nNodes0Max + 4");
        assert(int32(Dar_LibObj(s_DarLib, pObj.Fan1).Num) < s_DarLib.nNodes0Max + 4, "int32(Dar_LibObj(s_DarLib, pObj.Fan1).Num) < s_DarLib.nNodes0Max + 4");
        pData0 = s_DarLib.pDatas + Dar_LibObj(s_DarLib, pObj.Fan0).Num;
        pData1 = s_DarLib.pDatas + Dar_LibObj(s_DarLib, pObj.Fan1).Num;
        pData.Level = 1 + Abc_MaxInt(pData0.Level, pData1.Level);
        if pData0.pFunc == nil || pData1.pFunc == nil {
            continue; }
        pFanin0 = Aig_NotCond( pData0.pFunc, pObj.fCompl0 );
        pFanin1 = Aig_NotCond( pData1.pFunc, pObj.fCompl1 );
        if Aig_Regular(pFanin0) == pRoot || Aig_Regular(pFanin1) == pRoot {
            continue; }
        pData.pFunc = Aig_TableLookupTwo( p.pAig, pFanin0, pFanin1 );
        if pData.pFunc {
            // update the level to be more accurate
            pData.Level = Aig_Regular(pData.pFunc).Level;
            // mark the node if it is part of MFFC
            pData.fMffc = Aig_ObjIsTravIdCurrent(p.pAig, Aig_Regular(pData.pFunc));
            // assign the probability
            if p.pPars.fPower {
                var Prob float32  = Abc_Int2Float( Vec_IntEntry( p.pAig.vProbs, Aig_ObjId(Aig_Regular(pData.pFunc)) ) );
                if Aig_IsComplement(pData.pFunc) {
                    pData.dProb = 1.0-Prob;
                } else {
                    pData.dProb = Prob;
                }
            }
        }
    }
}

/**Function*************************************************************

  Synopsis    [Evaluates one cut.]

  Description [Returns the best gain.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibEval_rec(pObj *Dar_LibObj_t, Out int32, nNodesSaved int32, Required int32, pPower *float32) int32 {
    var pData *Dar_LibDat_t;
    var Power0, Power1 float32;
    var Area int32;
    if pPower != nil {
        *pPower = float32(0.0); }
    pData = s_DarLib.pDatas + pObj.Num;
    if pData.TravId == Out {
        return 0; }
    pData.TravId = Out;
    if pObj.fTerm {
        if pPower != nil {
            *pPower = pData.dProb; }
        return 0;
    }
    assert(pObj.Num > 3, "pObj.Num > 3");
    if pData.Level > Required {
        return 0xff; }
    if pData.pFunc && !pData.fMffc {
        if pPower != nil {
            *pPower = pData.dProb; }
        return 0;
    }
    // this is a new node - get a bound on the area of its branches
    nNodesSaved--;
    Area = Dar_LibEval_rec( Dar_LibObj(s_DarLib, pObj.Fan0), Out, nNodesSaved, Required+1, pPower? &Power0 : nil );
    if Area > nNodesSaved {
        return 0xff; }
    Area += Dar_LibEval_rec( Dar_LibObj(s_DarLib, pObj.Fan1), Out, nNodesSaved, Required+1, pPower? &Power1 : nil );
    if Area > nNodesSaved {
        return 0xff; }
    if pPower != nil {
        var pData0 *Dar_LibDat_t  = s_DarLib.pDatas + Dar_LibObj(s_DarLib, pObj.Fan0).Num;
        var pData1 *Dar_LibDat_t  = s_DarLib.pDatas + Dar_LibObj(s_DarLib, pObj.Fan1).Num;
        pData.dProb = (pObj.fCompl0? 1.0 - pData0.dProb : pData0.dProb)*
                       (pObj.fCompl1? 1.0 - pData1.dProb : pData1.dProb);
        *pPower = Power0 + 2.0 * pData0.dProb * (1.0 - pData0.dProb) +
                  Power1 + 2.0 * pData1.dProb * (1.0 - pData1.dProb);
    }
    return Area + 1;
}

/**Function*************************************************************

  Synopsis    [Evaluates one cut.]

  Description [Returns the best gain.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibEval(p *Dar_Man_t, pRoot *Aig_Obj_t, pCut *Dar_Cut_t, Required int32, pnMffcSize *int32) {
    var fTraining int32;
    var PowerSaved, PowerAdded float32;
    var pObj *Dar_LibObj_t;
    var Out, k, Class, nNodesSaved, nNodesAdded, nNodesGained int32;
    var clk abctime  = Abc_Clock();
    if pCut.nLeaves != 4 {
        return; }
    // check if the cut exits and assigns leaves and their levels
    if !(Dar_LibCutMatch(p, pCut) != 0) {
        return; }
    // mark MFFC of the node
    nNodesSaved = Dar_LibCutMarkMffc( p.pAig, pRoot, pCut.nLeaves, p.pPars.fPower? &PowerSaved : nil );
    // evaluate the cut
    Class = s_DarLib.pMap[pCut.uTruth];
    Dar_LibEvalAssignNums( p, Class, pRoot );
    // profile outputs by their savings
    p.nTotalSubgs += s_DarLib.nSubgr0[Class];
    p.ClassSubgs[Class] += s_DarLib.nSubgr0[Class];
    for  Out = 0; Out < s_DarLib.nSubgr0[Class]; Out++  {
        pObj = Dar_LibObj(s_DarLib, s_DarLib.pSubgr0[Class][Out]);
        if Aig_Regular(s_DarLib.pDatas[pObj.Num].pFunc) == pRoot {
            continue; }
        nNodesAdded = Dar_LibEval_rec( pObj, Out, nNodesSaved - !p.pPars.fUseZeros, Required, p.pPars.fPower? &PowerAdded : nil );
        nNodesGained = nNodesSaved - nNodesAdded;
        if p.pPars.fPower && PowerSaved < PowerAdded {
            continue; }
        if (fTraining != 0) && nNodesGained >= 0 {
            Dar_LibIncrementScore( Class, Out, nNodesGained + 1 ); }
        if nNodesGained < 0 || (nNodesGained == 0 && !p.pPars.fUseZeros) {
            continue; }
        if nNodesGained <  p.GainBest || 
            (nNodesGained == p.GainBest && s_DarLib.pDatas[pObj.Num].Level >= p.LevelBest) {
            continue; }
        // remember this possibility
        Vec_PtrClear( p.vLeavesBest );
        for  k = 0; k < int32(pCut.nLeaves); k++  {
            Vec_PtrPush( p.vLeavesBest, s_DarLib.pDatas[k].pFunc );
        }
        p.OutBest    = s_DarLib.pSubgr0[Class][Out];
        p.OutNumBest = Out;
        p.LevelBest  = s_DarLib.pDatas[pObj.Num].Level;
        p.GainBest   = nNodesGained;
        p.ClassBest  = Class;
        assert(p.LevelBest <= Required, "p.LevelBest <= Required");
        *pnMffcSize   = nNodesSaved;
    }
clk = Abc_Clock() - clk;
p.ClassTimes[Class] += clk;
p.timeEval += clk;
}

/**Function*************************************************************

  Synopsis    [Clears the fields of the nodes used in this cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibBuildClear_rec(pObj *Dar_LibObj_t, pCounter *int32) {
    if pObj.fTerm {
        return; }
    pObj.Num = (*pCounter)++;
    s_DarLib.pDatas[ pObj.Num ].pFunc = nil;
    Dar_LibBuildClear_rec( Dar_LibObj(s_DarLib, pObj.Fan0), pCounter );
    Dar_LibBuildClear_rec( Dar_LibObj(s_DarLib, pObj.Fan1), pCounter );
}

/**Function*************************************************************

  Synopsis    [Reconstructs the best cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibBuildBest_rec(p *Dar_Man_t, pObj *Dar_LibObj_t) *Aig_Obj_t {
    var pFanin0, pFanin1 *Aig_Obj_t;
    var pData *Dar_LibDat_t  = s_DarLib.pDatas + pObj.Num;
    if pData.pFunc {
        return pData.pFunc; }
    pFanin0 = Dar_LibBuildBest_rec( p, Dar_LibObj(s_DarLib, pObj.Fan0) );
    pFanin1 = Dar_LibBuildBest_rec( p, Dar_LibObj(s_DarLib, pObj.Fan1) );
    pFanin0 = Aig_NotCond( pFanin0, pObj.fCompl0 );
    pFanin1 = Aig_NotCond( pFanin1, pObj.fCompl1 );
    pData.pFunc = Aig_And( p.pAig, pFanin0, pFanin1 );
//    assert( pData->Level == (int)Aig_Regular(pData->pFunc)->Level );
    return pData.pFunc;
}

/**Function*************************************************************

  Synopsis    [Reconstructs the best cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibBuildBest(p *Dar_Man_t) *Aig_Obj_t {
    var i int32;
    var Counter int32 = 4;
    for  i = 0; i < Vec_PtrSize(p.vLeavesBest); i++  {
        s_DarLib.pDatas[i].pFunc = (Aig_Obj_t *)Vec_PtrEntry( p.vLeavesBest, i );
    }
    Dar_LibBuildClear_rec( Dar_LibObj(s_DarLib, p.OutBest), &Counter );
    return Dar_LibBuildBest_rec( p, Dar_LibObj(s_DarLib, p.OutBest) );
}






/**Function*************************************************************

  Synopsis    [Matches the cut with its canonical form.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar2_LibCutMatch(p *Gia_Man_t, vCutLits *Vec_Int_t, uTruth uint32) int32 {
    var uPhase uint32;
    var pPerm string;
    var i int32;
    assert(Vec_IntSize(vCutLits) == 4, "Vec_IntSize(vCutLits) == 4");
    // get the fanin permutation
    uPhase = s_DarLib.pPhases[uTruth];
    pPerm  = s_DarLib.pPerms4[ int32(s_DarLib.pPerms[uTruth]) ];
    // collect fanins with the corresponding permutation/phase
    for  i = 0; i < Vec_IntSize(vCutLits); i++  {
//        pFanin = Gia_ManObj( p, pCut->pLeaves[ (int)pPerm[i] ] );
//        pFanin = Gia_ManObj( p, Vec_IntEntry( vCutLits, (int)pPerm[i] ) );
//        pFanin = Gia_ObjFromLit( p, Vec_IntEntry( vCutLits, (int)pPerm[i] ) );
        s_DarLib.pDatas[i].iGunc = Abc_LitNotCond( Vec_IntEntry(vCutLits, int32(pPerm[i])), ((uPhase >> i) & 1) );
        s_DarLib.pDatas[i].Level = Gia_ObjLevel( p, Gia_Regular(Gia_ObjFromLit(p, s_DarLib.pDatas[i].iGunc)) );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Assigns numbers to the nodes of one class.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar2_LibEvalAssignNums(p *Gia_Man_t, Class int32) {
    var pObj *Dar_LibObj_t;
    var pData, pData0, pData1 *Dar_LibDat_t;
    var iFanin0, iFanin1, i, iLit int32;
    for  i = 0; i < s_DarLib.nNodes0[Class]; i++  {
        // get one class node, assign its temporary number and set its data
        pObj = Dar_LibObj(s_DarLib, s_DarLib.pNodes0[Class][i]);
        pObj.Num = 4 + i;
        assert(int32(pObj.Num) < s_DarLib.nNodes0Max + 4, "int32(pObj.Num) < s_DarLib.nNodes0Max + 4");
        pData = s_DarLib.pDatas + pObj.Num;
        pData.fMffc = 0;
        pData.iGunc = -1;
        pData.TravId = 0xFFFF;

        // explore the fanins
        assert(int32(Dar_LibObj(s_DarLib, pObj.Fan0).Num) < s_DarLib.nNodes0Max + 4, "int32(Dar_LibObj(s_DarLib, pObj.Fan0).Num) < s_DarLib.nNodes0Max + 4");
        assert(int32(Dar_LibObj(s_DarLib, pObj.Fan1).Num) < s_DarLib.nNodes0Max + 4, "int32(Dar_LibObj(s_DarLib, pObj.Fan1).Num) < s_DarLib.nNodes0Max + 4");
        pData0 = s_DarLib.pDatas + Dar_LibObj(s_DarLib, pObj.Fan0).Num;
        pData1 = s_DarLib.pDatas + Dar_LibObj(s_DarLib, pObj.Fan1).Num;
        pData.Level = 1 + Abc_MaxInt(pData0.Level, pData1.Level);
        if pData0.iGunc == -1 || pData1.iGunc == -1 {
            continue; }
        iFanin0 = Abc_LitNotCond( pData0.iGunc, pObj.fCompl0 );
        iFanin1 = Abc_LitNotCond( pData1.iGunc, pObj.fCompl1 );
        // compute the resulting literal
        if iFanin0 == 0 || iFanin1 == 0 || iFanin0 == Abc_LitNot(iFanin1) {
            iLit = 0;
        } else if iFanin0 == 1 || iFanin0 == iFanin1 {
            iLit = iFanin1;
        } else if iFanin1 == 1 {
            iLit = iFanin0;
        } else {
            iLit = Gia_ManHashLookup( p, Gia_ObjFromLit(p, iFanin0), Gia_ObjFromLit(p, iFanin1) );
            if iLit == 0 {
                iLit = -1; }
        }
        pData.iGunc = iLit;
        if pData.iGunc >= 0 {
            // update the level to be more accurate
            pData.Level = Gia_ObjLevel( p, Gia_Regular(Gia_ObjFromLit(p, pData.iGunc)) );
            // mark the node if it is part of MFFC
//            pData->fMffc = Gia_ObjIsTravIdCurrentArray(p, Gia_Regular(pData->pGunc));
        }
    }
}

/**Function*************************************************************

  Synopsis    [Evaluates one cut.]

  Description [Returns the best gain.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar2_LibEval_rec(pObj *Dar_LibObj_t, Out int32) int32 {
    var pData *Dar_LibDat_t;
    var Area int32;
    pData = s_DarLib.pDatas + pObj.Num;
    if pData.TravId == Out {
        return 0; }
    pData.TravId = Out;
    if pObj.fTerm {
        return 0; }
    assert(pObj.Num > 3, "pObj.Num > 3");
    if pData.iGunc >= 0 {//&& !pData->fMffc )
        return 0; }
    // this is a new node - get a bound on the area of its branches
//    nNodesSaved--;
    Area = Dar2_LibEval_rec( Dar_LibObj(s_DarLib, pObj.Fan0), Out );
//    if ( Area > nNodesSaved )
//        return 0xff;
    Area += Dar2_LibEval_rec( Dar_LibObj(s_DarLib, pObj.Fan1), Out );
//    if ( Area > nNodesSaved )
//        return 0xff;
    return Area + 1;
}

/**Function*************************************************************

  Synopsis    [Evaluates one cut.]

  Description [Returns the best gain.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar2_LibEval(p *Gia_Man_t, vCutLits *Vec_Int_t, uTruth uint32, fKeepLevel int32, vLeavesBest2 *Vec_Int_t) int32 {
    var p_OutBest int32     = -1;
    var p_OutNumBest int32  = -1;
    var p_LevelBest int32   =  1000000;
    var p_GainBest int32    = -1000000;
    var p_ClassBest int32   = -1;
//    int fTraining    =  0;
    var pObj *Dar_LibObj_t;
    var Out, k, Class, nNodesSaved, nNodesAdded, nNodesGained int32;
//    abctime clk = Abc_Clock();
    assert(Vec_IntSize(vCutLits) == 4, "Vec_IntSize(vCutLits) == 4");
    assert((uTruth >> 16) == 0, "(uTruth >> 16) == 0");
    // check if the cut exits and assigns leaves and their levels
    if !(Dar2_LibCutMatch(p, vCutLits, uTruth) != 0) {
        return -1; }
    // mark MFFC of the node
//    nNodesSaved = Dar2_LibCutMarkMffc( p->pAig, pRoot, pCut->nLeaves, p->pPars->fPower? &PowerSaved : NULL );
    nNodesSaved = 0;
    // evaluate the cut
    Class = s_DarLib.pMap[uTruth];
    Dar2_LibEvalAssignNums( p, Class );
    // profile outputs by their savings
//    p->nTotalSubgs += s_DarLib->nSubgr0[Class];
//    p->ClassSubgs[Class] += s_DarLib->nSubgr0[Class];
    for  Out = 0; Out < s_DarLib.nSubgr0[Class]; Out++  {
        pObj = Dar_LibObj(s_DarLib, s_DarLib.pSubgr0[Class][Out]);
//        nNodesAdded = Dar2_LibEval_rec( pObj, Out, nNodesSaved - !p->pPars->fUseZeros, Required, p->pPars->fPower? &PowerAdded : NULL );
        nNodesAdded = Dar2_LibEval_rec( pObj, Out );
        nNodesGained = nNodesSaved - nNodesAdded;
        if fKeepLevel != 0 {
            if s_DarLib.pDatas[pObj.Num].Level >  p_LevelBest || 
                (s_DarLib.pDatas[pObj.Num].Level == p_LevelBest && nNodesGained <= p_GainBest) {
                continue; }
        
        } else {
            if nNodesGained <  p_GainBest || 
                (nNodesGained == p_GainBest && s_DarLib.pDatas[pObj.Num].Level >= p_LevelBest) {
                continue; }
        }
        // remember this possibility
        Vec_IntClear( vLeavesBest2 );
        for  k = 0; k < Vec_IntSize(vCutLits); k++  {
            Vec_IntPush( vLeavesBest2, s_DarLib.pDatas[k].iGunc );
        }
        p_OutBest    = s_DarLib.pSubgr0[Class][Out];
        p_OutNumBest = Out;
        p_LevelBest  = s_DarLib.pDatas[pObj.Num].Level;
        p_GainBest   = nNodesGained;
        p_ClassBest  = Class;
//        assert( p_LevelBest <= Required );
    }
//clk = Abc_Clock() - clk;
//p->ClassTimes[Class] += clk;
//p->timeEval += clk;
    assert(p_OutBest != -1, "p_OutBest != -1");
    return p_OutBest;
}

/**Function*************************************************************

  Synopsis    [Clears the fields of the nodes used i this cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar2_LibBuildClear_rec(pObj *Dar_LibObj_t, pCounter *int32) {
    if pObj.fTerm {
        return; }
    pObj.Num = (*pCounter)++;
    s_DarLib.pDatas[ pObj.Num ].iGunc = -1;
    Dar2_LibBuildClear_rec( Dar_LibObj(s_DarLib, pObj.Fan0), pCounter );
    Dar2_LibBuildClear_rec( Dar_LibObj(s_DarLib, pObj.Fan1), pCounter );
}

/**Function*************************************************************

  Synopsis    [Reconstructs the best cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar2_LibBuildBest_rec(p *Gia_Man_t, pObj *Dar_LibObj_t) int32 {
    var pNode *Gia_Obj_t;
    var pData *Dar_LibDat_t;
    var iFanin0, iFanin1 int32;
    pData = s_DarLib.pDatas + pObj.Num;
    if pData.iGunc >= 0 {
        return pData.iGunc; }
    iFanin0 = Dar2_LibBuildBest_rec( p, Dar_LibObj(s_DarLib, pObj.Fan0) );
    iFanin1 = Dar2_LibBuildBest_rec( p, Dar_LibObj(s_DarLib, pObj.Fan1) );
    iFanin0 = Abc_LitNotCond( iFanin0, pObj.fCompl0 );
    iFanin1 = Abc_LitNotCond( iFanin1, pObj.fCompl1 );
    pData.iGunc = Gia_ManHashAnd( p, iFanin0, iFanin1 );
    pNode = Gia_ManObj( p, Abc_Lit2Var(pData.iGunc) );
    if Gia_ObjIsAnd( pNode ) {
        Gia_ObjSetAndLevel( p, pNode ); }
    Gia_ObjSetPhase( p, pNode );
    return pData.iGunc;
}

/**Function*************************************************************

  Synopsis    [Reconstructs the best cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar2_LibBuildBest(p *Gia_Man_t, vLeavesBest2 *Vec_Int_t, OutBest int32) int32 {
    var i int32;
    var iLeaf int32;
    var Counter int32 = 4;
    assert(Vec_IntSize(vLeavesBest2) == 4, "Vec_IntSize(vLeavesBest2) == 4");
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLeavesBest2)) && (((iLeaf) = Vec_IntEntry(vLeavesBest2, i)), 1); i++  {
        s_DarLib.pDatas[i].iGunc = iLeaf;
    }
    Dar2_LibBuildClear_rec( Dar_LibObj(s_DarLib, OutBest), &Counter );
    return Dar2_LibBuildBest_rec( p, Dar_LibObj(s_DarLib, OutBest) );
}

/**Function*************************************************************

  Synopsis    [Evaluate and build the new node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dar_LibEvalBuild(p *Gia_Man_t, vCutLits *Vec_Int_t, uTruth uint32, fKeepLevel int32, vLeavesBest2 *Vec_Int_t) int32 {
    var OutBest int32  = Dar2_LibEval( p, vCutLits, uTruth, fKeepLevel, vLeavesBest2 );
    return Dar2_LibBuildBest( p, vLeavesBest2, OutBest );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

