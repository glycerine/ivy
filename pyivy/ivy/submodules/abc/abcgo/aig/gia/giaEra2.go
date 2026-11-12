package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [gia.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "giaAig.h"

//ABC_NAMESPACE_IMPL_START

/*
    Limitations of this package:
    - no more than (1<<31)-1 state cubes and internal nodes
    - no more than MAX_VARS_NUM state variables
    - no more than MAX_CALL_NUM transitions from a state
    - cube list rebalancing happens when cube count reaches MAX_CUBE_NUM
*/

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

#define MAX_CALL_NUM    (1000000)  // the max number of recursive calls
#define MAX_ITEM_NUM      (1<<20)  // the number of items on a page
#define MAX_PAGE_NUM      (1<<11)  // the max number of memory pages
#define MAX_VARS_NUM      (1<<14)  // the max number of state vars allowed
const MAX_CUBE_NUM =          63   // the max number of cubes before rebalancing

// pointer to the tree node or state cube
   
type Gia_PtrAre_t struct {
    nItem uint32; /* C bitfield: 20 bits; review packing and truncation */  // item number (related to MAX_ITEM_NUM)
    nPage uint32; /* C bitfield: 11 bits; review packing and truncation */  // page number (related to MAX_PAGE_NUM)
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */  // user mark
}

typedef union Gia_PtrAreInt_t_ Gia_PtrAreInt_t;
union Gia_PtrAreInt_t_
{
    Gia_PtrAre_t   iGia;
    unsigned       iInt;
};

// tree nodes
   
type Gia_ObjAre_t struct {
    iVar uint32; /* C bitfield: 14 bits; review packing and truncation */  // variable     (related to MAX_VARS_NUM)
    nStas0 uint32; /* C bitfield: 6 bits; review packing and truncation */  // cube counter (related to MAX_CUBE_NUM)
    nStas1 uint32; /* C bitfield: 6 bits; review packing and truncation */  // cube counter (related to MAX_CUBE_NUM)
    nStas2 uint32; /* C bitfield: 6 bits; review packing and truncation */  // cube counter (related to MAX_CUBE_NUM)
    F [3]Gia_PtrAre_t;           // branches
}

// state cube
   
type Gia_StaAre_t struct {
    iPrev   Gia_PtrAre_t;          // previous state
    iNext   Gia_PtrAre_t;          // next cube in the list
    pData [0]uint32;       // state bits
}

// explicit state reachability manager
   
type Gia_ManAre_t struct {
    pAig     *Gia_Man_t;           // user's AIG manager
    pNew     *Gia_Man_t;           // temporary AIG manager
    ppObjs     **uint32;         // storage for objects (MAX_PAGE_NUM pages)
    ppStas     **uint32;         // storage for states  (MAX_PAGE_NUM pages)
//    unsigned *     pfUseless;      // to label useless cubes
//    int            nUselessAlloc;  // the number of useless alloced
    // internal flags
    fMiter            int32;         // stops when a bug is discovered
    fStopped            int32;       // set high when reachability is stopped
    fTree            int32;          // working in the tree mode
    // internal parametesr
    nWords            int32;         // the size of bit info in words
    nSize            int32;          // the size of state structure in words
    nObjPages            int32;      // the number of pages used for objects
    nStaPages            int32;      // the number of pages used for states
    nObjs            int32;          // the number of objects 
    nStas            int32;          // the number of states  
    iStaCur            int32;        // the next state to be explored
    Root   Gia_PtrAre_t;           // root of the tree
    vCiTfos     *Vec_Vec_t;        // storage for nodes in the CI TFOs
    vCiLits     *Vec_Vec_t;        // storage for literals of these nodes
    vCubesA     *Vec_Int_t;        // checked cubes
    vCubesB     *Vec_Int_t;        // unchecked cubes
    // deriving counter-example
    pSat          any;           // SAT solver
    vSatNumCis     *Vec_Int_t;     // SAT variables for CIs
    vSatNumCos     *Vec_Int_t;     // SAT variables for COs 
    vCofVars     *Vec_Int_t;       // variables used to cofactor
    vAssumps     *Vec_Int_t;       // temporary storage for assumptions
    pTarget  *Gia_StaAre_t;        // state that needs to be reached
    iOutFail            int32;       // the number of the failed output
    // statistics
    nChecks            int32;        // the number of timea cube was checked
    nEquals            int32;        // total number of equal
    nCompares            int32;      // the number of compares
    nRecCalls            int32;      // the number of rec calls
    nDisjs            int32;         // the number of disjoint cube pairs
    nDisjs2            int32;        // the number of disjoint cube pairs
    nDisjs3            int32;        // the number of disjoint cube pairs
    // time
    timeAig            int32;        // AIG cofactoring time
    timeCube            int32;       // cube checking time 
} 

func Gia_Int2Ptr(n uint32) Gia_PtrAre_t { var g Gia_PtrAreInt_t; g.iInt = n; return g.iGia;            }
func Gia_Ptr2Int(n Gia_PtrAre_t) uint32 { var g Gia_PtrAreInt_t; g.iGia = n; return g.iInt & 0x7fffffff; }

func Gia_ObjHasBranch0(q *Gia_ObjAre_t) int32 { return !q.nStas0 && (q.F[0].nPage || q.F[0].nItem);   }
func Gia_ObjHasBranch1(q *Gia_ObjAre_t) int32 { return !q.nStas1 && (q.F[1].nPage || q.F[1].nItem);   }
func Gia_ObjHasBranch2(q *Gia_ObjAre_t) int32 { return !q.nStas2 && (q.F[2].nPage || q.F[2].nItem);   }

func Gia_ManAreObj(p *Gia_ManAre_t, n Gia_PtrAre_t) *Gia_ObjAre_t { return (Gia_ObjAre_t *)(p.ppObjs[n.nPage] + (n.nItem << 2));      }
func Gia_ManAreSta(p *Gia_ManAre_t, n Gia_PtrAre_t) *Gia_StaAre_t { return (Gia_StaAre_t *)(p.ppStas[n.nPage] +  n.nItem * p.nSize); }
func Gia_ManAreObjInt(p *Gia_ManAre_t, n int32) *Gia_ObjAre_t { return Gia_ManAreObj( p, Gia_Int2Ptr(n) );               }
func Gia_ManAreStaInt(p *Gia_ManAre_t, n int32) *Gia_StaAre_t { return Gia_ManAreSta( p, Gia_Int2Ptr(n) );               }
func Gia_ManAreObjLast(p *Gia_ManAre_t) *Gia_ObjAre_t { return Gia_ManAreObjInt( p, p.nObjs-1 );                }
func Gia_ManAreStaLast(p *Gia_ManAre_t) *Gia_StaAre_t { return Gia_ManAreStaInt( p, p.nStas-1 );                }

func Gia_ObjNextObj0(p *Gia_ManAre_t, q *Gia_ObjAre_t) *Gia_ObjAre_t { return Gia_ManAreObj( p, q.F[0] );                      }
func Gia_ObjNextObj1(p *Gia_ManAre_t, q *Gia_ObjAre_t) *Gia_ObjAre_t { return Gia_ManAreObj( p, q.F[1] );                      }
func Gia_ObjNextObj2(p *Gia_ManAre_t, q *Gia_ObjAre_t) *Gia_ObjAre_t { return Gia_ManAreObj( p, q.F[2] );                      }

func Gia_StaHasValue0(p *Gia_StaAre_t, iReg int32) int32 { return Abc_InfoHasBit( p.pData,  iReg << 1 );           }
func Gia_StaHasValue1(p *Gia_StaAre_t, iReg int32) int32 { return Abc_InfoHasBit( p.pData, (iReg << 1) + 1 );      }

func Gia_StaSetValue0(p *Gia_StaAre_t, iReg int32) { Abc_InfoSetBit( p.pData,  iReg << 1 );                  }
func Gia_StaSetValue1(p *Gia_StaAre_t, iReg int32) { Abc_InfoSetBit( p.pData, (iReg << 1) + 1 );             }

func Gia_StaPrev(p *Gia_ManAre_t, pS *Gia_StaAre_t) *Gia_StaAre_t { return Gia_ManAreSta(p, pS.iPrev);                      }
func Gia_StaNext(p *Gia_ManAre_t, pS *Gia_StaAre_t) *Gia_StaAre_t { return Gia_ManAreSta(p, pS.iNext);                      }
func Gia_StaIsGood(p *Gia_ManAre_t, pS *Gia_StaAre_t) int32 { return ((unsigned *)pS) != p.ppStas[0];                         }

func Gia_StaSetUnused(pS *Gia_StaAre_t) { pS.iPrev.fMark = 1;                                     }
func Gia_StaIsUnused(pS *Gia_StaAre_t) int32 { return  pS.iPrev.fMark;                                 }
func Gia_StaIsUsed(pS *Gia_StaAre_t) int32 { return !pS.iPrev.fMark;                                 }

// Original macro:
// #define Gia_ManAreForEachCubeList( p, pList, pCube )                         \
//     for ( pCube = pList; Gia_StaIsGood(p, pCube); pCube = Gia_StaNext(p, pCube) )

// Original macro:
// #define Gia_ManAreForEachCubeList2( p, iList, pCube, iCube )                 \
//     for ( iCube = Gia_Ptr2Int(iList), pCube = Gia_ManAreSta(p, iList);       \
//           Gia_StaIsGood(p, pCube);                                           \
//           iCube = Gia_Ptr2Int(pCube->iNext), pCube = Gia_StaNext(p, pCube) )

// Original macro:
// #define Gia_ManAreForEachCubeStore( p, pCube, i )                            \
//     for ( i = 1; i < p->nStas && (pCube = Gia_ManAreStaInt(p, i)); i++ )

// Original macro:
// #define Gia_ManAreForEachCubeVec( vVec, p, pCube, i )                        \
//     for ( i = 0; i < Vec_IntSize(vVec) && (pCube = Gia_ManAreStaInt(p, Vec_IntEntry(vVec,i))); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Count state minterms contained in a cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountMintermsInCube(pCube *Gia_StaAre_t, nVars int32, pStore *uint32) {
    var Mint, Mask uint32;
    var i int32;
    var m int32;
    var nMints int32;
    var nDashes int32;
    var Dashes [32]int32;
    // count the number of dashes
    for  i = 0; i < nVars; i++  {
        if Gia_StaHasValue0( pCube, i ) != 0 {
            continue; }
        if Gia_StaHasValue1( pCube, i ) != 0 {
            Mask |= (1 << i);
        } else {
            Dashes[nDashes++] = i; }
    }
    // fill in the miterms
    nMints = (1 << nDashes);
    for  m = 0; m < nMints; m++  {
        Mint = Mask;
        for  i = 0; i < nVars; i++  {
            if m & (1 << i) {
                Mint |= (1 << Dashes[i]); }
        }
        Abc_InfoSetBit( pStore, Mint );
    }
}

/**Function*************************************************************

  Synopsis    [Count state minterms contains in the used cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCountMinterms(p *Gia_ManAre_t) int32 {
    var pCube *Gia_StaAre_t;
    var pMemory *uint32;
    var i, nMemSize, Counter int32;
    if Gia_ManRegNum(p.pAig) > 30 {
        return -1; }
    nMemSize = Abc_BitWordNum( 1 << Gia_ManRegNum(p.pAig) );
    // inlining of macro ABC_CALLOC
    pMemory  = ((unsigned *) calloc((size_t)(nMemSize), sizeof(unsigned)));
    // inlining of macro Gia_ManAreForEachCubeStore
    for  i = 1; i < p.nStas && (pCube = Gia_ManAreStaInt(p, i)); i++  {
        if Gia_StaIsUsed(pCube) != 0 {
            Gia_ManCountMintermsInCube( pCube, Gia_ManRegNum(p.pAig), pMemory ); }
    }
    for  i = 0; i < nMemSize; i++  {
        Counter += Gia_WordCountOnes( pMemory[i] );
    }
    // inlining of macro ABC_FREE
    ((pMemory) ? (free((char *) (pMemory)), (pMemory) = 0) : 0);
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Derives the TFO of one CI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDeriveCiTfo_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vRes *Vec_Int_t) int32 {
    if Gia_ObjIsCi(pObj) {
        return pObj.fMark0; }
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return pObj.fMark0; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManDeriveCiTfo_rec( p, Gia_ObjFanin0(pObj), vRes );
    Gia_ManDeriveCiTfo_rec( p, Gia_ObjFanin1(pObj), vRes );
    pObj.fMark0 = Gia_ObjFanin0(pObj).fMark0 | Gia_ObjFanin1(pObj).fMark0;
    if pObj.fMark0 {
        Vec_IntPush( vRes, Gia_ObjId(p, pObj) ); }
    return pObj.fMark0;
}


/**Function*************************************************************

  Synopsis    [Derives the TFO of one CI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDeriveCiTfoOne(p *Gia_Man_t, pPivot *Gia_Obj_t) *Vec_Int_t {
    var vRes *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    assert(pPivot.fMark0 == 0, "pPivot.fMark0 == 0");
    pPivot.fMark0 = 1;
    vRes = Vec_IntAlloc( 100 );
    Vec_IntPush( vRes, Gia_ObjId(p, pPivot) );
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        Gia_ManDeriveCiTfo_rec( p, Gia_ObjFanin0(pObj), vRes );
        if Gia_ObjFanin0(pObj).fMark0 {
            Vec_IntPush( vRes, Gia_ObjId(p, pObj) ); }
    }
    pPivot.fMark0 = 0;
    return vRes;
}

/**Function*************************************************************

  Synopsis    [Derives the TFO of each CI.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDeriveCiTfo(p *Gia_Man_t) *Vec_Vec_t {
    var vRes *Vec_Ptr_t;
    var pPivot *Gia_Obj_t;
    var i int32;
    Gia_ManCleanMark0( p );
    Gia_ManIncrementTravId( p );
    vRes = Vec_PtrAlloc( Gia_ManCiNum(p) );
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pPivot) = Gia_ManCi(p, i)); i++  {
        Vec_PtrPush( vRes, Gia_ManDeriveCiTfoOne(p, pPivot) );
    }
    Gia_ManCleanMark0( p );
    return (Vec_Vec_t *)vRes;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if states are equal.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_StaAreEqual(p1 *Gia_StaAre_t, p2 *Gia_StaAre_t, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if p1.pData[w] != p2.pData[w] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if states are disjoint.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_StaAreDisjoint(p1 *Gia_StaAre_t, p2 *Gia_StaAre_t, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if ((p1.pData[w] ^ p2.pData[w]) >> 1) & (p1.pData[w] ^ p2.pData[w]) & 0x55555555 {
            return 1; }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Returns 1 if cube p1 contains cube p2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_StaAreContain(p1 *Gia_StaAre_t, p2 *Gia_StaAre_t, nWords int32) int32 {
    var w int32;
    for  w = 0; w < nWords; w++  {
        if (p1.pData[w] | p2.pData[w]) != p2.pData[w] {
            return 0; }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Returns the number of dashes in p1 that are non-dashes in p2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_StaAreDashNum(p1 *Gia_StaAre_t, p2 *Gia_StaAre_t, nWords int32) int32 {
    var w, Counter int32;
    for  w = 0; w < nWords; w++  {
        Counter += Gia_WordCountOnes( (~(p1.pData[w] ^ (p1.pData[w] >> 1))) & (p2.pData[w] ^ (p2.pData[w] >> 1)) & 0x55555555 );
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns the number of a variable for sharping the cube.]

  Description [Counts the number of variables that have dash in p1 and 
  non-dash in p2. If there is exactly one such variable, returns its index.
  Otherwise returns -1.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_StaAreSharpVar(p1 *Gia_StaAre_t, p2 *Gia_StaAre_t, nWords int32) int32 {
    var Word uint32;
    var w int32;
    var iVar int32 = -1;
    for  w = 0; w < nWords; w++  {
        Word = (~(p1.pData[w] ^ (p1.pData[w] >> 1))) & (p2.pData[w] ^ (p2.pData[w] >> 1)) & 0x55555555;
        if Word == 0 {
            continue; }
        if !Gia_WordHasOneBit(Word) {
            return -1; }
        // has exactly one bit
        if iVar >= 0 {
            return -1; }
        // the first variable of this type
        iVar = 16 * w + Gia_WordFindFirstBit( Word ) / 2;
    }
    return iVar;
}

/**Function*************************************************************

  Synopsis    [Returns the number of a variable for merging the cubes.]

  Description [If there is exactly one such variable, returns its index.
  Otherwise returns -1.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_StaAreDisjointVar(p1 *Gia_StaAre_t, p2 *Gia_StaAre_t, nWords int32) int32 {
    var Word uint32;
    var w int32;
    var iVar int32 = -1;
    for  w = 0; w < nWords; w++  {
        Word = (p1.pData[w] ^ p2.pData[w]) & ((p1.pData[w] ^ p2.pData[w]) >> 1) & 0x55555555;
        if Word == 0 {
            continue; }
        if !Gia_WordHasOneBit(Word) {
            return -1; }
        // has exactly one bit
        if iVar >= 0 {
            return -1; }
        // the first variable of this type
        iVar = 16 * w + Gia_WordFindFirstBit( Word ) / 2;
    }
    return iVar;
}

/**Function*************************************************************

  Synopsis    [Creates reachability manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCreate(pAig *Gia_Man_t) *Gia_ManAre_t {
    
    assert(sizeof(Gia_ObjAre_t) == 16, "sizeof(Gia_ObjAre_t) == 16");
    // inlining of macro ABC_CALLOC
    p := &Gia_ManAre_t{};
    p.pAig       = pAig;
    p.nWords     = Abc_BitWordNum( 2 * Gia_ManRegNum(pAig) );
    p.nSize      = sizeof(Gia_StaAre_t)/4 + p.nWords;
    // inlining of macro ABC_CALLOC
    p.ppObjs     = ((unsigned * *) calloc((size_t)(MAX_PAGE_NUM), sizeof(unsigned *)));
    // inlining of macro ABC_CALLOC
    p.ppStas     = ((unsigned * *) calloc((size_t)(MAX_PAGE_NUM), sizeof(unsigned *)));
    p.vCiTfos    = Gia_ManDeriveCiTfo( pAig );
    p.vCiLits    = Vec_VecDupInt( p.vCiTfos );
    p.vCubesA    = Vec_IntAlloc( 100 );
    p.vCubesB    = Vec_IntAlloc( 100 );
    p.iOutFail   = -1;
    return p;
}

/**Function*************************************************************

  Synopsis    [Deletes reachability manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreFree(p *Gia_ManAre_t) {
    var i int32;
    Gia_ManStop( p.pAig );
    if p.pNew {
        Gia_ManStop( p.pNew ); }
    Vec_IntFree( p.vCubesA );
    Vec_IntFree( p.vCubesB );
    Vec_VecFree( p.vCiTfos );
    Vec_VecFree( p.vCiLits );
    for  i = 0; i < p.nObjPages; i++  {
        // inlining of macro ABC_FREE
        ((p.ppObjs[i]) ? (free((char *) (p.ppObjs[i])), (p.ppObjs[i]) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((p.ppObjs) ? (free((char *) (p.ppObjs)), (p.ppObjs) = 0) : 0);
    for  i = 0; i < p.nStaPages; i++  {
        // inlining of macro ABC_FREE
        ((p.ppStas[i]) ? (free((char *) (p.ppStas[i])), (p.ppStas[i]) = 0) : 0);
    }
    // inlining of macro ABC_FREE
    ((p.ppStas) ? (free((char *) (p.ppStas)), (p.ppStas) = 0) : 0);
//    ABC_FREE( p->pfUseless );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Returns new object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCreateObj(p *Gia_ManAre_t) *Gia_ObjAre_t {
    if p.nObjs == p.nObjPages * MAX_ITEM_NUM {
        if p.nObjPages == MAX_PAGE_NUM {
            fmt.Printf("ERA manager has run out of memory after allocating 2B internal nodes.\n");
            return nil;
        }
        // inlining of macro ABC_CALLOC
        p.ppObjs[p.nObjPages++] = ((unsigned *) calloc((size_t)(MAX_ITEM_NUM * 4), sizeof(unsigned)));
        if p.nObjs == 0 {
            p.nObjs = 1; }
    }
    return Gia_ManAreObjInt( p, p.nObjs++ );
}

/**Function*************************************************************

  Synopsis    [Returns new state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCreateSta(p *Gia_ManAre_t) *Gia_StaAre_t {
    if p.nStas == p.nStaPages * MAX_ITEM_NUM {
        if p.nStaPages == MAX_PAGE_NUM {
            fmt.Printf("ERA manager has run out of memory after allocating 2B state cubes.\n");
            return nil;
        }
        if p.ppStas[p.nStaPages] == nil {
            // inlining of macro ABC_CALLOC
            p.ppStas[p.nStaPages] = ((unsigned *) calloc((size_t)(MAX_ITEM_NUM * p.nSize), sizeof(unsigned))); }
        p.nStaPages++;
        if p.nStas == 0 {
            p.nStas = 1;
//            p->nUselessAlloc = (1 << 18);
//            p->pfUseless = ABC_CALLOC( unsigned, p->nUselessAlloc );
        }
//        if ( p->nStas == p->nUselessAlloc * 32 )
//        {
//            p->nUselessAlloc *= 2;
//            p->pfUseless = ABC_REALLOC( unsigned, p->pfUseless, p->nUselessAlloc );
//            memset( p->pfUseless + p->nUselessAlloc/2, 0, sizeof(unsigned) * p->nUselessAlloc/2 );
//        }
    }
    return Gia_ManAreStaInt( p, p.nStas++ );
}

/**Function*************************************************************

  Synopsis    [Recycles new state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreRycycleSta(p *Gia_ManAre_t, pSta *Gia_StaAre_t) {
    memset( pSta, 0, p.nSize << 2 );
    if pSta == Gia_ManAreStaLast(p) {
        p.nStas--;
        if p.nStas == (p.nStaPages-1) * MAX_ITEM_NUM {
            p.nStaPages--; }
    
    } else {
//        Gia_StaSetUnused( pSta );
    }
}

/**Function*************************************************************

  Synopsis    [Creates new state state from the latch values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCreateStaNew(p *Gia_ManAre_t) *Gia_StaAre_t {
    var pSta *Gia_StaAre_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pSta = Gia_ManAreCreateSta( p );
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, Gia_ManPoNum(p.pAig)+i)); i++  {
        if pObj.Value == 0 {
            Gia_StaSetValue0( pSta, i );
        } else if pObj.Value == 1 {
            Gia_StaSetValue1( pSta, i ); }
    }
    return pSta;
}

/**Function*************************************************************

  Synopsis    [Creates new state state with latch init values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCreateStaInit(p *Gia_ManAre_t) *Gia_StaAre_t {
    var pObj *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, Gia_ManPoNum(p.pAig)+i)); i++  {
        pObj.Value = 0;
    }
    return Gia_ManAreCreateStaNew( p );
}


/**Function*************************************************************

  Synopsis    [Prints the state cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManArePrintCube(p *Gia_ManAre_t, pSta *Gia_StaAre_t) {
    var pObj *Gia_Obj_t;
    var i, Count0, Count1, Count2 int32;
    fmt.Printf("%4d %4d :  ", p.iStaCur, p.nStas-1);
    fmt.Printf("Prev %4d   ", Gia_Ptr2Int(pSta.iPrev));
    fmt.Printf("%p   ", pSta);
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, Gia_ManPoNum(p.pAig)+i)); i++  {
        if Gia_StaHasValue0(pSta, i) != 0 {
            fmt.Printf("0");
Count0++;
        } else if Gia_StaHasValue1(pSta, i) != 0 {
            fmt.Printf("1");
Count1++;
        } else {
            fmt.Printf("-");
Count2++; }
    }
    fmt.Printf("  0 =%3d", Count0);
    fmt.Printf("  1 =%3d", Count1);
    fmt.Printf("  - =%3d", Count2);
    fmt.Printf("\n");
}
 
/**Function*************************************************************

  Synopsis    [Counts the depth of state transitions leading ot this state.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreDepth(p *Gia_ManAre_t, iState int32) int32 {
    var pSta *Gia_StaAre_t;
    var Counter int32;
    for  pSta = Gia_ManAreStaInt(p, iState); (Gia_StaIsGood(p, pSta) != 0); pSta = Gia_StaPrev(p, pSta)  {
        Counter++;
    }
    return Counter;
} 

/**Function*************************************************************

  Synopsis    [Counts the number of cubes in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreListCountListUsed(p *Gia_ManAre_t, Root Gia_PtrAre_t) int32 {
    var pCube *Gia_StaAre_t;
    var Counter int32;
    // inlining of macro Gia_ManAreForEachCubeList
    for  pCube = Gia_ManAreSta(p, Root); (Gia_StaIsGood(p, pCube) != 0); pCube = Gia_StaNext(p, pCube)  {
        Counter += Gia_StaIsUsed(pCube);
    }
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Counts the number of used cubes in the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreListCountUsed_rec(p *Gia_ManAre_t, Root Gia_PtrAre_t, fTree int32) int32 {
    var pObj *Gia_ObjAre_t;
    if !(fTree != 0) {
        return Gia_ManAreListCountListUsed( p, Root ); }
    pObj = Gia_ManAreObj(p, Root);
    return Gia_ManAreListCountUsed_rec( p, pObj.F[0], Gia_ObjHasBranch0(pObj) ) +
           Gia_ManAreListCountUsed_rec( p, pObj.F[1], Gia_ObjHasBranch1(pObj) ) +
           Gia_ManAreListCountUsed_rec( p, pObj.F[2], Gia_ObjHasBranch2(pObj) );
}

/**Function*************************************************************

  Synopsis    [Counts the number of used cubes in the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreListCountUsed(p *Gia_ManAre_t) int32 {
    return Gia_ManAreListCountUsed_rec( p, p.Root, p.fTree );
}


/**Function*************************************************************

  Synopsis    [Prints used cubes in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManArePrintListUsed(p *Gia_ManAre_t, Root Gia_PtrAre_t) int32 {
    var pCube *Gia_StaAre_t;
    // inlining of macro Gia_ManAreForEachCubeList
    for  pCube = Gia_ManAreSta(p, Root); (Gia_StaIsGood(p, pCube) != 0); pCube = Gia_StaNext(p, pCube)  {
        if Gia_StaIsUsed(pCube) != 0 {
            Gia_ManArePrintCube( p, pCube ); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Prints used cubes in the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManArePrintUsed_rec(p *Gia_ManAre_t, Root Gia_PtrAre_t, fTree int32) int32 {
    var pObj *Gia_ObjAre_t;
    if !(fTree != 0) {
        return Gia_ManArePrintListUsed( p, Root ); }
    pObj = Gia_ManAreObj(p, Root);
    return Gia_ManArePrintUsed_rec( p, pObj.F[0], Gia_ObjHasBranch0(pObj) ) +
           Gia_ManArePrintUsed_rec( p, pObj.F[1], Gia_ObjHasBranch1(pObj) ) +
           Gia_ManArePrintUsed_rec( p, pObj.F[2], Gia_ObjHasBranch2(pObj) );
}

/**Function*************************************************************

  Synopsis    [Prints used cubes in the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManArePrintUsed(p *Gia_ManAre_t) int32 {
    return Gia_ManArePrintUsed_rec( p, p.Root, p.fTree );
}


/**Function*************************************************************

  Synopsis    [Best var has max weight.]

  Description [Weight is defined as the number of 0/1-lits minus the 
  absolute value of the diff between the number of 0-lits and 1-lits.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreFindBestVar(p *Gia_ManAre_t, List Gia_PtrAre_t) int32 {
    var pCube *Gia_StaAre_t;
    var Count0, Count1, Count2 int32;
    var iVarThis int32;
    var iVarBest int32 = -1;
    var WeightThis int32;
    var WeightBest int32 = -1;
    for  iVarThis = 0; iVarThis < Gia_ManRegNum(p.pAig); iVarThis++  {
        Count0 = Count1 = Count2 = 0;
        // inlining of macro Gia_ManAreForEachCubeList
        for  pCube = Gia_ManAreSta(p, List); (Gia_StaIsGood(p, pCube) != 0); pCube = Gia_StaNext(p, pCube)  {
            if Gia_StaIsUnused(pCube) != 0 {
                continue; }
            if Gia_StaHasValue0(pCube, iVarThis) != 0 {
                Count0++;
            } else if Gia_StaHasValue1(pCube, iVarThis) != 0 {
                Count1++;
            } else {
                Count2++; }
        }
//        printf( "%4d : %5d  %5d  %5d   Weight = %5d\n", iVarThis, Count0, Count1, Count2, 
//            Count0 + Count1 - (Count0 > Count1 ? Count0 - Count1 : Count1 - Count0) );
        if (!Count0 && !Count1) || (!Count0 && !(Count2 != 0)) || (!Count1 && !(Count2 != 0)) {
            continue; }
        WeightThis = Count0 + Count1 - (Count0 > Count1 ? Count0 - Count1 : Count1 - Count0);
        if WeightBest < WeightThis {
            WeightBest = WeightThis;
            iVarBest = iVarThis;
        }
    }
    if iVarBest == -1 {
        Gia_ManArePrintListUsed( p, List );
        fmt.Printf("Error: Best variable not found!!!\n");
    }
    assert(iVarBest != -1, "iVarBest != -1");
    return iVarBest;
}
 
/**Function*************************************************************

  Synopsis    [Rebalances the tree when cubes exceed the limit.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreRebalance(p *Gia_ManAre_t, pRoot *Gia_PtrAre_t) {
    var pNode *Gia_ObjAre_t;
    var pCube *Gia_StaAre_t;
    var iCube, iNext Gia_PtrAre_t;
    assert(pRoot.nItem || pRoot.nPage, "pRoot.nItem || pRoot.nPage");
    pNode = Gia_ManAreCreateObj( p );
    pNode.iVar = Gia_ManAreFindBestVar( p, *pRoot );
    for  iCube = *pRoot, pCube = Gia_ManAreSta(p, iCube), iNext = pCube.iNext; 
          (Gia_StaIsGood(p, pCube) != 0); 
          iCube = iNext,  pCube = Gia_ManAreSta(p, iCube), iNext = pCube.iNext  {
        if Gia_StaIsUnused(pCube) != 0 {
            continue; }
        if Gia_StaHasValue0(pCube, pNode.iVar) != 0 {
            pCube.iNext = pNode.F[0], pNode.F[0] = iCube, pNode.nStas0++;
        } else if Gia_StaHasValue1(pCube, pNode.iVar) != 0 {
            pCube.iNext = pNode.F[1], pNode.F[1] = iCube, pNode.nStas1++;
        } else {
            pCube.iNext = pNode.F[2], pNode.F[2] = iCube, pNode.nStas2++; }
    }
    *pRoot = Gia_Int2Ptr(p.nObjs - 1);
    assert(pNode == Gia_ManAreObj(p, *pRoot), "pNode == Gia_ManAreObj(p, *pRoot)");
    p.fTree = 1;
}
 
/**Function*************************************************************

  Synopsis    [Compresses the list by removing unused cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCompress(p *Gia_ManAre_t, pRoot *Gia_PtrAre_t) {
    var pCube *Gia_StaAre_t;
    var iList Gia_PtrAre_t  = *pRoot;
    var iCube, iNext Gia_PtrAre_t;
    assert(pRoot.nItem || pRoot.nPage, "pRoot.nItem || pRoot.nPage");
    pRoot.nItem = 0; 
    pRoot.nPage = 0;
    for  iCube = iList, pCube = Gia_ManAreSta(p, iCube), iNext = pCube.iNext; 
          (Gia_StaIsGood(p, pCube) != 0); 
          iCube = iNext, pCube = Gia_ManAreSta(p, iCube), iNext = pCube.iNext  {
        if Gia_StaIsUnused(pCube) != 0 {
            continue; }
        pCube.iNext = *pRoot;
        *pRoot = iCube;
    }
}


/**Function*************************************************************

  Synopsis    [Checks if the state exists in the list.]

  Description [The state may be sharped.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCubeCheckList(p *Gia_ManAre_t, pRoot *Gia_PtrAre_t, pSta *Gia_StaAre_t) int32 {
    var fVerbose int32;
    var pCube *Gia_StaAre_t;
    var iVar int32;
if fVerbose != 0 {
fmt.Printf("Trying cube: ");
Gia_ManArePrintCube( p, pSta );
}
    // inlining of macro Gia_ManAreForEachCubeList
    for  pCube = Gia_ManAreSta(p, *pRoot); (Gia_StaIsGood(p, pCube) != 0); pCube = Gia_StaNext(p, pCube)  {
        p.nChecks++;
        if Gia_StaIsUnused( pCube ) != 0 {
            continue; }
        if Gia_StaAreDisjoint( pSta, pCube, p.nWords ) != 0 {
            continue; }
        if Gia_StaAreContain( pCube, pSta, p.nWords ) != 0 {
if fVerbose != 0 {
fmt.Printf("Contained in ");
Gia_ManArePrintCube( p, pCube );
}
            Gia_ManAreRycycleSta( p, pSta );
            return 0;
        }
        if Gia_StaAreContain( pSta, pCube, p.nWords ) != 0 {
if fVerbose != 0 {
fmt.Printf("Contains     ");
Gia_ManArePrintCube( p, pCube );
}
            Gia_StaSetUnused( pCube );
            continue;
        }
        iVar = Gia_StaAreSharpVar( pSta, pCube, p.nWords );
        if iVar == -1 {
            continue; }
if fVerbose != 0 {
fmt.Printf("Sharped by   ");
Gia_ManArePrintCube( p, pCube );
Gia_ManArePrintCube( p, pSta );
}
//        printf( "%d  %d\n", Gia_StaAreDashNum( pSta, pCube, p->nWords ), Gia_StaAreSharpVar( pSta, pCube, p->nWords ) );
        assert(!(Gia_StaHasValue0(pSta, iVar) != 0) && !(Gia_StaHasValue1(pSta, iVar) != 0), "!Gia_StaHasValue0(pSta, iVar) && !Gia_StaHasValue1(pSta, iVar)");
        assert(Gia_StaHasValue0(pCube, iVar) ^  Gia_StaHasValue1(pCube, iVar), "Gia_StaHasValue0(pCube, iVar) ^  Gia_StaHasValue1(pCube, iVar)");
        if Gia_StaHasValue0(pCube, iVar) != 0 {
            Gia_StaSetValue1( pSta, iVar );
        } else {
            Gia_StaSetValue0( pSta, iVar ); }
//        return Gia_ManAreCubeCheckList( p, pRoot, pSta );
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Adds new state to the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCubeAddToList(p *Gia_ManAre_t, pRoot *Gia_PtrAre_t, pSta *Gia_StaAre_t) {
    var fVerbose int32;
    pSta.iNext = *pRoot;
    *pRoot = Gia_Int2Ptr( p.nStas - 1 );
    assert(pSta == Gia_ManAreSta(p, *pRoot), "pSta == Gia_ManAreSta(p, *pRoot)");
if fVerbose != 0 {
fmt.Printf("Adding cube: ");
Gia_ManArePrintCube( p, pSta );
//printf( "\n" );
}
}
 
/**Function*************************************************************

  Synopsis    [Checks if the cube like this exists in the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCubeCheckTree_rec(p *Gia_ManAre_t, pObj *Gia_ObjAre_t, pSta *Gia_StaAre_t) int32 {
    var RetValue int32;
    if Gia_StaHasValue0(pSta, pObj.iVar) != 0 {
        if Gia_ObjHasBranch0(pObj) != 0 {
            RetValue = Gia_ManAreCubeCheckTree_rec( p, Gia_ObjNextObj0(p, pObj), pSta );
        } else {
            RetValue = Gia_ManAreCubeCheckList( p, pObj.F, pSta ); }
        if RetValue == 0 {
            return 0; }
    
    } else if Gia_StaHasValue1(pSta, pObj.iVar) != 0 {
        if Gia_ObjHasBranch1(pObj) != 0 {
            RetValue = Gia_ManAreCubeCheckTree_rec( p, Gia_ObjNextObj1(p, pObj), pSta );
        } else {
            RetValue = Gia_ManAreCubeCheckList( p, pObj.F + 1, pSta ); }
        if RetValue == 0 {
            return 0; }
    }
    if Gia_ObjHasBranch2(pObj) != 0 {
        return Gia_ManAreCubeCheckTree_rec( p, Gia_ObjNextObj2(p, pObj), pSta ); }
    return Gia_ManAreCubeCheckList( p, pObj.F + 2, pSta );
}

/**Function*************************************************************

  Synopsis    [Adds new cube to the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCubeAddToTree_rec(p *Gia_ManAre_t, pObj *Gia_ObjAre_t, pSta *Gia_StaAre_t) {
    if Gia_StaHasValue0(pSta, pObj.iVar) != 0 {
        if Gia_ObjHasBranch0(pObj) != 0 {
            Gia_ManAreCubeAddToTree_rec( p, Gia_ObjNextObj0(p, pObj), pSta );
        } else {
            Gia_ManAreCubeAddToList( p, pObj.F, pSta );
            if ++pObj.nStas0 == MAX_CUBE_NUM {
                pObj.nStas0 = Gia_ManAreListCountListUsed( p, pObj.F[0] );
                if pObj.nStas0 < MAX_CUBE_NUM/2 {
                    Gia_ManAreCompress( p, pObj.F );
                } else {
                    Gia_ManAreRebalance( p, pObj.F );
                    pObj.nStas0 = 0;
                }
            }
        }
    
    } else if Gia_StaHasValue1(pSta, pObj.iVar) != 0 {
        if Gia_ObjHasBranch1(pObj) != 0 {
            Gia_ManAreCubeAddToTree_rec( p, Gia_ObjNextObj1(p, pObj), pSta );
        } else {
            Gia_ManAreCubeAddToList( p, pObj.F+1, pSta );
            if ++pObj.nStas1 == MAX_CUBE_NUM {
                pObj.nStas1 = Gia_ManAreListCountListUsed( p, pObj.F[1] );
                if pObj.nStas1 < MAX_CUBE_NUM/2 {
                    Gia_ManAreCompress( p, pObj.F+1 );
                } else {
                    Gia_ManAreRebalance( p, pObj.F+1 );
                    pObj.nStas1 = 0;
                }
            }
        }
    
    } else {
        if Gia_ObjHasBranch2(pObj) != 0 {
            Gia_ManAreCubeAddToTree_rec( p, Gia_ObjNextObj2(p, pObj), pSta );
        } else {
            Gia_ManAreCubeAddToList( p, pObj.F+2, pSta );
            if ++pObj.nStas2 == MAX_CUBE_NUM {
                pObj.nStas2 = Gia_ManAreListCountListUsed( p, pObj.F[2] );
                if pObj.nStas2 < MAX_CUBE_NUM/2 {
                    Gia_ManAreCompress( p, pObj.F+2 );
                } else {
                    Gia_ManAreRebalance( p, pObj.F+2 );
                    pObj.nStas2 = 0;
                }
            }
        }
    }
}



/**Function*************************************************************

  Synopsis    [Collects overlapping cubes in the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCubeCollectList(p *Gia_ManAre_t, pRoot *Gia_PtrAre_t, pSta *Gia_StaAre_t) int32 {
    var pCube *Gia_StaAre_t;
    var iCube int32;
    // inlining of macro Gia_ManAreForEachCubeList2
    for  iCube = Gia_Ptr2Int(*pRoot), pCube = Gia_ManAreSta(p, *pRoot);                 (Gia_StaIsGood(p, pCube) != 0);                                                     iCube = Gia_Ptr2Int(pCube.iNext), pCube = Gia_StaNext(p, pCube)  {
        if Gia_StaIsUnused( pCube ) != 0 {
            continue; }
        if Gia_StaAreDisjoint( pSta, pCube, p.nWords ) != 0 {
/*
            int iVar;
            p->nDisjs++;
            iVar = Gia_StaAreDisjointVar( pSta, pCube, p->nWords );
            if ( iVar >= 0 )
            {
                p->nDisjs2++;
                if ( iCube > p->iStaCur )
                    p->nDisjs3++;
            }
*/
            continue;
        }
//        p->nCompares++;
//        p->nEquals += Gia_StaAreEqual( pSta, pCube, p->nWords );
        if iCube <= p.iStaCur {
            Vec_IntPush( p.vCubesA, iCube );
        } else {
            Vec_IntPush( p.vCubesB, iCube ); }
    }
    return 1;
}

/**Function*************************************************************

  Synopsis    [Collects overlapping cubes in the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCubeCollectTree_rec(p *Gia_ManAre_t, pObj *Gia_ObjAre_t, pSta *Gia_StaAre_t) int32 {
    var RetValue int32;
    if Gia_StaHasValue0(pSta, pObj.iVar) != 0 {
        if Gia_ObjHasBranch0(pObj) != 0 {
            RetValue = Gia_ManAreCubeCollectTree_rec( p, Gia_ObjNextObj0(p, pObj), pSta );
        } else {
            RetValue = Gia_ManAreCubeCollectList( p, pObj.F, pSta ); }
        if RetValue == 0 {
            return 0; }
    
    } else if Gia_StaHasValue1(pSta, pObj.iVar) != 0 {
        if Gia_ObjHasBranch1(pObj) != 0 {
            RetValue = Gia_ManAreCubeCollectTree_rec( p, Gia_ObjNextObj1(p, pObj), pSta );
        } else {
            RetValue = Gia_ManAreCubeCollectList( p, pObj.F + 1, pSta ); }
        if RetValue == 0 {
            return 0; }
    }
    if Gia_ObjHasBranch2(pObj) != 0 {
        return Gia_ManAreCubeCollectTree_rec( p, Gia_ObjNextObj2(p, pObj), pSta ); }
    return Gia_ManAreCubeCollectList( p, pObj.F + 2, pSta );
}
 
/**Function*************************************************************

  Synopsis    [Checks if the cube like this exists in the tree.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCubeCheckTree(p *Gia_ManAre_t, pSta *Gia_StaAre_t) int32 {
    var pCube *Gia_StaAre_t;
    var i, iVar int32;
    assert(p.fTree, "p.fTree");
    Vec_IntClear( p.vCubesA );
    Vec_IntClear( p.vCubesB );
    Gia_ManAreCubeCollectTree_rec( p, Gia_ManAreObj(p, p.Root), pSta );
//    if ( p->nStas > 3000 )
//        printf( "%d %d  \n", Vec_IntSize(p->vCubesA), Vec_IntSize(p->vCubesB) );
//    Vec_IntSort( p->vCubesA, 0 );
//    Vec_IntSort( p->vCubesB, 0 );
    // inlining of macro Gia_ManAreForEachCubeVec
    for  i = 0; i < Vec_IntSize(p.vCubesA) && (pCube = Gia_ManAreStaInt(p, Vec_IntEntry(p.vCubesA,i))); i++  {
        if Gia_StaIsUnused( pCube ) != 0 {
            continue; }
        if Gia_StaAreDisjoint( pSta, pCube, p.nWords ) != 0 {
            continue; }
        if Gia_StaAreContain( pCube, pSta, p.nWords ) != 0 {
            Gia_ManAreRycycleSta( p, pSta );
            return 0;
        }
        if Gia_StaAreContain( pSta, pCube, p.nWords ) != 0 {
            Gia_StaSetUnused( pCube );
            continue;
        }
        iVar = Gia_StaAreSharpVar( pSta, pCube, p.nWords );
        if iVar == -1 {
            continue; }
        assert(!(Gia_StaHasValue0(pSta, iVar) != 0) && !(Gia_StaHasValue1(pSta, iVar) != 0), "!Gia_StaHasValue0(pSta, iVar) && !Gia_StaHasValue1(pSta, iVar)");
        assert(Gia_StaHasValue0(pCube, iVar) ^  Gia_StaHasValue1(pCube, iVar), "Gia_StaHasValue0(pCube, iVar) ^  Gia_StaHasValue1(pCube, iVar)");
        if Gia_StaHasValue0(pCube, iVar) != 0 {
            Gia_StaSetValue1( pSta, iVar );
        } else {
            Gia_StaSetValue0( pSta, iVar ); }
        return Gia_ManAreCubeCheckTree( p, pSta );
    }
    // inlining of macro Gia_ManAreForEachCubeVec
    for  i = 0; i < Vec_IntSize(p.vCubesB) && (pCube = Gia_ManAreStaInt(p, Vec_IntEntry(p.vCubesB,i))); i++  {
        if Gia_StaIsUnused( pCube ) != 0 {
            continue; }
        if Gia_StaAreDisjoint( pSta, pCube, p.nWords ) != 0 {
            continue; }
        if Gia_StaAreContain( pCube, pSta, p.nWords ) != 0 {
            Gia_ManAreRycycleSta( p, pSta );
            return 0;
        }
        if Gia_StaAreContain( pSta, pCube, p.nWords ) != 0 {
            Gia_StaSetUnused( pCube );
            continue;
        }
        iVar = Gia_StaAreSharpVar( pSta, pCube, p.nWords );
        if iVar == -1 {
            continue; }
        assert(!(Gia_StaHasValue0(pSta, iVar) != 0) && !(Gia_StaHasValue1(pSta, iVar) != 0), "!Gia_StaHasValue0(pSta, iVar) && !Gia_StaHasValue1(pSta, iVar)");
        assert(Gia_StaHasValue0(pCube, iVar) ^  Gia_StaHasValue1(pCube, iVar), "Gia_StaHasValue0(pCube, iVar) ^  Gia_StaHasValue1(pCube, iVar)");
        if Gia_StaHasValue0(pCube, iVar) != 0 {
            Gia_StaSetValue1( pSta, iVar );
        } else {
            Gia_StaSetValue0( pSta, iVar ); }
        return Gia_ManAreCubeCheckTree( p, pSta );
    }
/*
    if ( p->nStas > 3000 )
    {
printf( "Trying cube:       " );
Gia_ManArePrintCube( p, pSta );
    Gia_ManAreForEachCubeVec( p->vCubesA, p, pCube, i )
    {
printf( "aaaaaaaaaaaa %5d ", Vec_IntEntry(p->vCubesA,i) );
Gia_ManArePrintCube( p, pCube );
    }
    Gia_ManAreForEachCubeVec( p->vCubesB, p, pCube, i )
    {
printf( "bbbbbbbbbbbb %5d ", Vec_IntEntry(p->vCubesB,i) );
Gia_ManArePrintCube( p, pCube );
    }
    }
*/
    return 1;
}

/**Function*************************************************************

  Synopsis    [Processes the new cube.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreCubeProcess(p *Gia_ManAre_t, pSta *Gia_StaAre_t) int32 {
    var RetValue int32;
    p.nChecks = 0;
    if !p.fTree && p.nStas == MAX_CUBE_NUM {
        Gia_ManAreRebalance( p, &p.Root ); }
    if p.fTree {
//        RetValue = Gia_ManAreCubeCheckTree_rec( p, Gia_ManAreObj(p, p->Root), pSta );
        RetValue = Gia_ManAreCubeCheckTree( p, pSta );
        if RetValue != 0 {
            Gia_ManAreCubeAddToTree_rec( p, Gia_ManAreObj(p, p.Root), pSta ); }
    
    } else {
        RetValue = Gia_ManAreCubeCheckList( p, &p.Root, pSta );
        if RetValue != 0 {
            Gia_ManAreCubeAddToList( p, &p.Root, pSta ); }
    }
//    printf( "%d ", p->nChecks );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    [Returns the most used CI, or NULL if condition is met.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreMostUsedPi_rec(p *Gia_Man_t, pObj *Gia_Obj_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        pObj.Value++;
        return;
    }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    Gia_ManAreMostUsedPi_rec( p, Gia_ObjFanin0(pObj) );
    Gia_ManAreMostUsedPi_rec( p, Gia_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Returns the most used CI, or NULL if condition is met.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreMostUsedPi(p *Gia_ManAre_t) *Gia_Obj_t {
    var pObj *Gia_Obj_t;
    var pObjMax *Gia_Obj_t = nil;
    var i int32;
    // clean CI counters
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pNew.vCis)) && ((pObj) = Gia_ManCi(p.pNew, i)); i++  {
        pObj.Value = 0;
    }
    // traverse from each register output
    // inlining of macro Gia_ManForEachRi
    for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, Gia_ManPoNum(p.pAig)+i)); i++  {
        if pObj.Value <= 1 {
            continue; }
        Gia_ManIncrementTravId( p.pNew );
        Gia_ManAreMostUsedPi_rec( p.pNew, Gia_ManObj(p.pNew, Abc_Lit2Var(pObj.Value)) );
    }
    // check the CI counters
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.pNew.vCis)) && ((pObj) = Gia_ManCi(p.pNew, i)); i++  {
        if pObjMax == nil || pObjMax.Value < pObj.Value {
            pObjMax = pObj; }
    }
    // return the result
    if pObjMax.Value > 1 {
        return pObjMax;
    } else {
        return nil;
    }
}

/**Function*************************************************************

  Synopsis    [Counts maximum support of primary outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckPOs_rec(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return 0; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    if Gia_ObjIsCi(pObj) {
        return 1; }
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    return Gia_ManCheckPOs_rec( p, Gia_ObjFanin0(pObj) ) +
           Gia_ManCheckPOs_rec( p, Gia_ObjFanin1(pObj) );
}

/**Function*************************************************************

  Synopsis    [Counts maximum support of primary outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckPOs(p *Gia_ManAre_t) int32 {
    var pObj, pObjNew *Gia_Obj_t;
    var i, CountCur, CountMax int32;
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, i)); i++  {
        pObjNew = Gia_ManObj( p.pNew, Abc_Lit2Var(pObj.Value) );
        if Gia_ObjIsConst0(pObjNew) {
            CountCur = 0;
        } else {
            Gia_ManIncrementTravId( p.pNew );
            CountCur = Gia_ManCheckPOs_rec( p.pNew, pObjNew );
        }
        CountMax = Abc_MaxInt( CountMax, CountCur );
    }
    return CountMax;
}

/**Function*************************************************************

  Synopsis    [Returns the status of the primary outputs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCheckPOstatus(p *Gia_ManAre_t) int32 {
    var pObj, pObjNew *Gia_Obj_t;
    var i int32;
    // inlining of macro Gia_ManForEachPo
    for  i = 0; (i < Gia_ManPoNum(p.pAig)) && ((pObj) = Gia_ManCo(p.pAig, i)); i++  {
        pObjNew = Gia_ManObj( p.pNew, Abc_Lit2Var(pObj.Value) );
        if Gia_ObjIsConst0(pObjNew) {
            if Abc_LitIsCompl(pObj.Value) {
                p.iOutFail = i;
                return 1;
            }
        
        } else {
            p.iOutFail = i;
//            printf( "To fix later:  PO may be assertable.\n" );
            return 1;
        }
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Derives next state cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreDeriveNexts_rec(p *Gia_ManAre_t, Sta Gia_PtrAre_t) int32 {
    var pPivot *Gia_Obj_t;
    var vLits, vTfos *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var clk abctime;
    if ++p.nRecCalls == MAX_CALL_NUM {
        return 0; }
    if (pPivot = Gia_ManAreMostUsedPi(p)) == nil {
        var pNew *Gia_StaAre_t;
        clk = Abc_Clock();
        pNew = Gia_ManAreCreateStaNew( p );
        pNew.iPrev = Sta;
        p.fStopped = (p.fMiter && (Gia_ManCheckPOstatus(p) & 1));
        if p.fStopped {
            assert(p.pTarget == nil, "p.pTarget == nil");
            p.pTarget = pNew;
            return 1;
        }
        Gia_ManAreCubeProcess( p, pNew );
        p.timeCube += Abc_Clock() - clk;
        return p.fStopped;
    }
    // remember values in the cone and perform update
    vTfos = Vec_VecEntryInt( p.vCiTfos, Gia_ObjCioId(pPivot) );
    vLits = Vec_VecEntryInt( p.vCiLits, Gia_ObjCioId(pPivot) );
    assert(Vec_IntSize(vTfos) == Vec_IntSize(vLits), "Vec_IntSize(vTfos) == Vec_IntSize(vLits)");
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vTfos)) && ((pObj) = Gia_ManObj(p.pAig, Vec_IntEntry(vTfos,i))); i++  {
        Vec_IntWriteEntry( vLits, i, pObj.Value );
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( p.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        } else {
            assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
            pObj.Value = 0;
        }
    }
    if Gia_ManAreDeriveNexts_rec( p, Sta ) != 0 {
        return 1; }
    // compute different values
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vTfos)) && ((pObj) = Gia_ManObj(p.pAig, Vec_IntEntry(vTfos,i))); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManHashAnd( p.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ObjFanin0Copy(pObj);
        } else {
            assert(Gia_ObjIsCi(pObj), "Gia_ObjIsCi(pObj)");
            pObj.Value = 1;
        }
    }
    if Gia_ManAreDeriveNexts_rec( p, Sta ) != 0 {
        return 1; }
    // reset the original values
    // inlining of macro Gia_ManForEachObjVec
    for  i = 0; (i < Vec_IntSize(vTfos)) && ((pObj) = Gia_ManObj(p.pAig, Vec_IntEntry(vTfos,i))); i++  {
        pObj.Value = Vec_IntEntry( vLits, i );
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    [Derives next state cubes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreDeriveNexts(p *Gia_ManAre_t, Sta Gia_PtrAre_t) int32 {
    var pSta *Gia_StaAre_t;
    var pObj *Gia_Obj_t;
    var i, RetValue int32;
    var clk abctime  = Abc_Clock();
    pSta = Gia_ManAreSta( p, Sta );
    if Gia_StaIsUnused(pSta) != 0 {
        return 0; }
    // recycle the manager
    if p.pNew && Gia_ManObjNum(p.pNew) > 1000000 {
        Gia_ManStop( p.pNew );
        p.pNew = nil;
    }
    // allocate the manager
    if p.pNew == nil {
        p.pNew = Gia_ManStart( 10 * Gia_ManObjNum(p.pAig) );
        Gia_ManIncrementTravId( p.pNew );
        Gia_ManHashAlloc( p.pNew );
        Gia_ManConst0(p.pAig).Value = 0;
        // inlining of macro Gia_ManForEachCi
        for  i = 0; (i < Vec_IntSize(p.pAig.vCis)) && ((pObj) = Gia_ManCi(p.pAig, i)); i++  {
            pObj.Value = Gia_ManAppendCi(p.pNew);
        }
    }
    // inlining of macro Gia_ManForEachRo
    for  i = 0; (i < Gia_ManRegNum(p.pAig)) && ((pObj) = Gia_ManCi(p.pAig, Gia_ManPiNum(p.pAig)+i)); i++  {
        if Gia_StaHasValue0( pSta, i ) != 0 {
            pObj.Value = 0;
        } else if Gia_StaHasValue1( pSta, i ) != 0 {
            pObj.Value = 1;
        } else { // don't-care literal
            pObj.Value = Abc_Var2Lit( Gia_ObjId( p.pNew, Gia_ManCi(p.pNew, Gia_ObjCioId(pObj)) ), 0 ); }
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.pAig.nObjs) && ((pObj) = Gia_ManObj(p.pAig, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManHashAnd( p.pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.pAig.vCos)) && ((pObj) = Gia_ManCo(p.pAig, i)); i++  {
        pObj.Value = Gia_ObjFanin0Copy(pObj);
    }

    // perform case-splitting
    p.nRecCalls = 0;
    RetValue = Gia_ManAreDeriveNexts_rec( p, Sta );
    if p.nRecCalls >= MAX_CALL_NUM {
        fmt.Printf("Exceeded the limit on the number of transitions from a state cube (%d).\n", MAX_CALL_NUM);
        p.fStopped = 1;
    }
//    printf( "%d ", p->nRecCalls );
//printf( "%d ", Gia_ManObjNum(p->pNew) );
    p.timeAig += Abc_Clock() - clk;
    return RetValue;
}

 
/**Function*************************************************************

  Synopsis    [Prints the report]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManArePrintReport(p *Gia_ManAre_t, Time abctime, fFinal int32) {
    fmt.Printf("States =%10d. Reached =%10d. R = %5.3f. Depth =%6d. Mem =%9.2f MB.  ", p.iStaCur, p.nStas, 1.0*p.iStaCur/p.nStas, Gia_ManAreDepth(p, p.iStaCur), (sizeof(Gia_ManAre_t) + 4.0*Gia_ManRegNum(p.pAig) + 8.0*MAX_PAGE_NUM + 
             4.0*p.nStaPages*p.nSize*MAX_ITEM_NUM + 16.0*p.nObjPages*MAX_ITEM_NUM)/(1<<20));
    if fFinal != 0 {
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - Time)))/(float64(int64(time.Second))));;
    
    } else {
        // inlining of macro ABC_PRTr
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\r", 1.0*(float64((Abc_Clock() - Time)))/(float64(int64(time.Second))));;
    }
}
 
/**Function*************************************************************

  Synopsis    [Performs explicit reachability.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManArePerform(pAig *Gia_Man_t, nStatesMax int32, fMiter int32, fVerbose int32) int32 {
//    extern Gia_Man_t * Gia_ManCompress2( Gia_Man_t * p, int fUpdateLevel, int fVerbose );
    extern Abc_Cex_t * Gia_ManAreDeriveCex( Gia_ManAre_t * p, Gia_StaAre_t * pLast );
    var p *Gia_ManAre_t;
    var clk abctime  = Abc_Clock();
    var RetValue int32  = 1;
    if Gia_ManRegNum(pAig) > MAX_VARS_NUM {
        fmt.Printf("Currently can only handle circuit with up to %d registers.\n", MAX_VARS_NUM);
        return -1;
    }
    // inlining of macro ABC_FREE
    ((pAig.pCexSeq) ? (free((char *) (pAig.pCexSeq)), (pAig.pCexSeq) = 0) : 0);
//    p = Gia_ManAreCreate( Gia_ManCompress2(pAig, 0, 0) );
    p = Gia_ManAreCreate( Gia_ManDup(pAig) );
    p.fMiter = fMiter;
    Gia_ManAreCubeProcess( p, Gia_ManAreCreateStaInit(p) );
    for  p.iStaCur = 1; p.iStaCur < p.nStas; p.iStaCur++  {
//        printf( "Explored state %d. Total cubes %d.\n", p->iStaCur, p->nStas-1 );
        if (Gia_ManAreDeriveNexts( p, Gia_Int2Ptr(p.iStaCur) ) != 0) || p.nStas > nStatesMax {
            pAig.pCexSeq = Gia_ManAreDeriveCex( p, p.pTarget ); }
        if p.fStopped {
            RetValue = -1;
            break;
        }
        if fVerbose != 0 {//&& p->iStaCur % 5000 == 0 )
            Gia_ManArePrintReport( p, clk, 0 ); }
    }
    Gia_ManArePrintReport( p, clk, 1 );
    fmt.Printf("%s after finding %d state cubes (%d not contained) with depth %d.  ", p.fStopped ? "Stopped" : "Completed", p.nStas, Gia_ManAreListCountUsed(p), Gia_ManAreDepth(p, p.iStaCur-1));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    if pAig.pCexSeq != nil {
        Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.\n", 
            p.iStaCur, pAig.pName, Gia_ManAreDepth(p, p.iStaCur)-1 ); }
    if fVerbose != 0 {
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Cofactoring"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeAig - p.timeCube)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeAig - p.timeCube)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Containment"));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((p.timeCube)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((p.timeCube)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("Other      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((Abc_Clock() - clk - p.timeAig)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((Abc_Clock() - clk - p.timeAig)))/(float64((Abc_Clock() - clk))) : 0.0);;
        // inlining of macro ABC_PRTP
        Abc_Print(1, "%s =", ("TOTAL      "));
Abc_Print(1, "%9.2f sec (%6.2f %%)\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))), (float64((Abc_Clock() - clk)))? 100.0*(float64((Abc_Clock() - clk)))/(float64((Abc_Clock() - clk))) : 0.0);;
    }
    if Gia_ManRegNum(pAig) <= 30 {
        clk = Abc_Clock();
        fmt.Printf("The number of unique state minterms in computed state cubes is %d.   ", Gia_ManCountMinterms(p));
        // inlining of macro ABC_PRT
        Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
//    printf( "Compares = %d.  Equals = %d.  Disj = %d.  Disj2 = %d.  Disj3 = %d.\n", 
//        p->nCompares, p->nEquals, p->nDisjs, p->nDisjs2, p->nDisjs3 );
//    Gia_ManAreFindBestVar( p, Gia_ManAreSta(p, p->Root) );
//    Gia_ManArePrintUsed( p );
    Gia_ManAreFree( p );
    // verify
    if pAig.pCexSeq {
        if !Gia_ManVerifyCex( pAig, pAig.pCexSeq, 0 ) {
            fmt.Printf("Generated counter-example is INVALID.                       \n");
        } else {
            fmt.Printf("Generated counter-example verified correctly.               \n"); }
        return 0;
    }
    return RetValue;
}

//ABC_NAMESPACE_IMPL_END

//#include "giaAig.h"
//#include "sat/cnf/cnf.h"
//#include "sat/bsat/satSolver.h"

//ABC_NAMESPACE_IMPL_START


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreDeriveCexSatStart(p *Gia_ManAre_t) {
    var pAig2 *Aig_Man_t;
    var pCnf *Cnf_Dat_t;
    assert(p.pSat == nil, "p.pSat == nil");
    pAig2 = Gia_ManToAig( p.pAig, 0 );
    Aig_ManSetRegNum( pAig2, 0 );
    pCnf = Cnf_Derive( pAig2, Gia_ManCoNum(p.pAig) );
    p.pSat = Cnf_DataWriteIntoSolver( pCnf, 1, 0 );
    p.vSatNumCis = Cnf_DataCollectCiSatNums( pCnf, pAig2 );
    p.vSatNumCos = Cnf_DataCollectCoSatNums( pCnf, pAig2 );
    Cnf_DataFree( pCnf );
    Aig_ManStop( pAig2 );
    p.vAssumps = Vec_IntAlloc( 100 );
    p.vCofVars = Vec_IntAlloc( 100 );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreDeriveCexSatStop(p *Gia_ManAre_t) {
    assert(p.pSat != nil, "p.pSat != nil");
    assert(p.pTarget != nil, "p.pTarget != nil");
    sat_solver_delete( (sat_solver *)p.pSat );
    Vec_IntFree( p.vSatNumCis );
    Vec_IntFree( p.vSatNumCos );
    Vec_IntFree( p.vAssumps );
    Vec_IntFree( p.vCofVars );
    p.pTarget = nil;
    p.pSat = nil;
}

/**Function*************************************************************

  Synopsis    [Computes satisfying assignment in one timeframe.]

  Description [Returns the vector of integers represeting PIO ids
  of the primary inputs that should be 1 in the counter-example.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreDeriveCexSat(p *Gia_ManAre_t, pCur *Gia_StaAre_t, pNext *Gia_StaAre_t, iOutFailed int32) {
    var i, status int32;
    // make assuptions
    Vec_IntClear( p.vAssumps );
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if Gia_StaHasValue0(pCur, i) != 0 {
            Vec_IntPush( p.vAssumps, Abc_Var2Lit( Vec_IntEntry(p.vSatNumCis, Gia_ManPiNum(p.pAig)+i), 1 ) );
        } else if Gia_StaHasValue1(pCur, i) != 0 {
            Vec_IntPush( p.vAssumps, Abc_Var2Lit( Vec_IntEntry(p.vSatNumCis, Gia_ManPiNum(p.pAig)+i), 0 ) ); }
    }
    if pNext != nil {
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if Gia_StaHasValue0(pNext, i) != 0 {
            Vec_IntPush( p.vAssumps, Abc_Var2Lit( Vec_IntEntry(p.vSatNumCos, Gia_ManPoNum(p.pAig)+i), 1 ) );
        } else if Gia_StaHasValue1(pNext, i) != 0 {
            Vec_IntPush( p.vAssumps, Abc_Var2Lit( Vec_IntEntry(p.vSatNumCos, Gia_ManPoNum(p.pAig)+i), 0 ) ); }
    } }
    if iOutFailed >= 0 {
        assert(iOutFailed < Gia_ManPoNum(p.pAig), "iOutFailed < Gia_ManPoNum(p.pAig)");
        Vec_IntPush( p.vAssumps, Abc_Var2Lit( Vec_IntEntry(p.vSatNumCos, iOutFailed), 0 ) );
    }
    // solve SAT
    status = sat_solver_solve( (sat_solver *)p.pSat, (int *)Vec_IntArray(p.vAssumps), (int *)Vec_IntArray(p.vAssumps) + Vec_IntSize(p.vAssumps), 
        (ABC_INT64_T)1000000, (ABC_INT64_T)0, (ABC_INT64_T)0, (ABC_INT64_T)0 );
    if status != l_True {
        fmt.Printf("SAT problem is not satisfiable. Failure...\n");
        return;
    }
    assert(status == l_True, "status == l_True");
    // check the model
    Vec_IntClear( p.vCofVars );
    for  i = 0; i < Gia_ManPiNum(p.pAig); i++  {
        if sat_solver_var_value( (sat_solver *)p.pSat, Vec_IntEntry(p.vSatNumCis, i) ) {
            Vec_IntPush( p.vCofVars, i ); }
    }
    // write the current state
    for  i = 0; i < Gia_ManRegNum(p.pAig); i++  {
        if Gia_StaHasValue0(pCur, i) != 0 {
            assert(sat_solver_var_value( (sat_solver *)p.pSat, Vec_IntEntry(p.vSatNumCis, Gia_ManPiNum(p.pAig)+i) ) == 0, "sat_solver_var_value( (sat_solver *)p.pSat, Vec_IntEntry(p.vSatNumCis, Gia_ManPiNum(p.pAig)+i) ) == 0");
        } else if Gia_StaHasValue1(pCur, i) != 0 {
            assert(sat_solver_var_value( (sat_solver *)p.pSat, Vec_IntEntry(p.vSatNumCis, Gia_ManPiNum(p.pAig)+i) ) == 1, "sat_solver_var_value( (sat_solver *)p.pSat, Vec_IntEntry(p.vSatNumCis, Gia_ManPiNum(p.pAig)+i) ) == 1"); }
        // set don't-care value
        if sat_solver_var_value( (sat_solver *)p.pSat, Vec_IntEntry(p.vSatNumCis, Gia_ManPiNum(p.pAig)+i) ) == 0 {
            Gia_StaSetValue0( pCur, i );
        } else {
            Gia_StaSetValue1( pCur, i ); }
    }
}

/**Function*************************************************************

  Synopsis    [Returns the status of the cone.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManAreDeriveCex(p *Gia_ManAre_t, pLast *Gia_StaAre_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var vStates *Vec_Ptr_t;
    var pSta, pPrev *Gia_StaAre_t;
    var Var, i, v int32;
    assert(p.iOutFail >= 0, "p.iOutFail >= 0");
    Gia_ManAreDeriveCexSatStart( p );
    // compute the trace
    vStates = Vec_PtrAlloc( 1000 );
    for  pSta = pLast; (Gia_StaIsGood(p, pSta) != 0); pSta = Gia_StaPrev(p, pSta)  {
        if pSta != pLast {
            Vec_PtrPush( vStates, pSta ); }
    }
    assert(Vec_PtrSize(vStates) >= 1, "Vec_PtrSize(vStates) >= 1");
    // start the counter-example
    pCex = Abc_CexAlloc( Gia_ManRegNum(p.pAig), Gia_ManPiNum(p.pAig), Vec_PtrSize(vStates) );
    pCex.iFrame = Vec_PtrSize(vStates)-1;
    pCex.iPo = p.iOutFail;
    // compute states
    pPrev = nil;
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vStates)) && (((pSta) = (Gia_StaAre_t *)Vec_PtrEntry(vStates, i)), 1); i++  {
        Gia_ManAreDeriveCexSat( p, pSta, pPrev, (i == 0) ? p.iOutFail : -1 ); 
        pPrev = pSta;
        // create the counter-example
        // inlining of macro Vec_IntForEachEntry
        for  v = 0; (v < Vec_IntSize(p.vCofVars)) && (((Var) = Vec_IntEntry(p.vCofVars, v)), 1); v++  {
            assert(Var < Gia_ManPiNum(p.pAig), "Var < Gia_ManPiNum(p.pAig)");
            Abc_InfoSetBit( pCex.pData, Gia_ManRegNum(p.pAig) + (Vec_PtrSize(vStates)-1-i) * Gia_ManPiNum(p.pAig) + Var );
        }
    }
    // free temporary things
    Vec_PtrFree( vStates );
    Gia_ManAreDeriveCexSatStop( p );
    return pCex;
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

