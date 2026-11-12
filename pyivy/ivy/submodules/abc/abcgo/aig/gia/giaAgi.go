package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaAig.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    []

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaAig.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

#define AGI_PI ABC_CONST(0xFFFFFFFF00000000) 
#define AGI_RO ABC_CONST(0xFFFFFFFE00000000) 
#define AGI_PO ABC_CONST(0xFFFFFFFD00000000) 
#define AGI_RI ABC_CONST(0xFFFFFFFC00000000) 
#define AGI_C0 ABC_CONST(0xFFFFFFFBFFFFFFFA) 
#define AGI_M0 ABC_CONST(0x00000000FFFFFFFF) 
#define AGI_M1 ABC_CONST(0xFFFFFFFF00000000) 

   
type Agi_Man_t struct {
    pName              string;       // name of the AIG
    pSpec              string;       // name of the input file
    nCap                int32;        // number of objects
    nObjs                int32;       // number of objects
    nNodes                int32;      // number of objects
    nRegs                int32;       // number of registers
    nTravIds           uint32;    // number of objects
    vCis          Vec_Int_t;        // comb inputs
    vCos          Vec_Int_t;        // comb outputs
    pObjs              *word;       // objects
    pThird          *uint32;      // third input
    pTravIds          *uint32;    // traversal IDs
    pNext          *uint32;       // next values
    pTable          *uint32;      // hash table
    pCopy          *uint32;       // hash table
}

func Agi_ManObjNum(p *Agi_Man_t) int32 { return p.nObjs;                                        } 
func Agi_ManCiNum(p *Agi_Man_t) int32 { return Vec_IntSize( &p.vCis );                         } 
func Agi_ManCoNum(p *Agi_Man_t) int32 { return Vec_IntSize( &p.vCos );                         } 
func Agi_ManNodeNum(p *Agi_Man_t) int32 { return p.nNodes;                                       } 

func Agi_ObjLit0(p *Agi_Man_t, i int32) uint32 { return uint32((p.pObjs[i]));                         } 
func Agi_ObjLit1(p *Agi_Man_t, i int32) uint32 { return uint32((p.pObjs[i] >> 32));                   } 
func Agi_ObjLit2(p *Agi_Man_t, i int32) uint32 { return p.pThird[i];                                    } 
func Agi_ObjVar0(p *Agi_Man_t, i int32) int32 { return Agi_ObjLit0(p, i) >> 1;                          } 
func Agi_ObjVar1(p *Agi_Man_t, i int32) int32 { return Agi_ObjLit1(p, i) >> 1;                          } 
func Agi_ObjVar2(p *Agi_Man_t, i int32) int32 { return Agi_ObjLit2(p, i) >> 1;                          } 
func Agi_ObjSetLit0(p *Agi_Man_t, i int32, l uint32) { p.pObjs[i] = (p.pObjs[i] & AGI_M1) | (word)l;         } 
func Agi_ObjSetLit1(p *Agi_Man_t, i int32, l uint32) { p.pObjs[i] = (p.pObjs[i] & AGI_M0) | ((word)l << 32); } 
func Agi_ObjSetLit2(p *Agi_Man_t, i int32, l uint32) { p.pThird[i] = l;                                       }  

func Agi_ObjIsC0(p *Agi_Man_t, i int32) int32 { return (i == 0);                                        } 
func Agi_ObjIsPi(p *Agi_Man_t, i int32) int32 { return (p.pObjs[i] & AGI_PI) == AGI_PI;                } 
func Agi_ObjIsRo(p *Agi_Man_t, i int32) int32 { return (p.pObjs[i] & AGI_PI) == AGI_RO;                } 
func Agi_ObjIsPo(p *Agi_Man_t, i int32) int32 { return (p.pObjs[i] & AGI_PI) == AGI_PO;                } 
func Agi_ObjIsRi(p *Agi_Man_t, i int32) int32 { return (p.pObjs[i] & AGI_PI) == AGI_RI;                } 
func Agi_ObjIsCi(p *Agi_Man_t, i int32) int32 { return (p.pObjs[i] & AGI_RO) == AGI_RO;                } 
func Agi_ObjIsCo(p *Agi_Man_t, i int32) int32 { return (p.pObjs[i] & AGI_PO) == AGI_PO;                } 
func Agi_ObjIsNode(p *Agi_Man_t, i int32) int32 { return p.pObjs[i] < AGI_C0;                            } 
func Agi_ObjIsBuf(p *Agi_Man_t, i int32) int32 { return Agi_ObjLit0(p, i) == Agi_ObjLit1(p, i);          } 
func Agi_ObjIsAnd(p *Agi_Man_t, i int32) int32 { return Agi_ObjIsNode(p, i) && Agi_ObjLit0(p, i) < Agi_ObjLit1(p, i);  } 
func Agi_ObjIsXor(p *Agi_Man_t, i int32) int32 { return Agi_ObjIsNode(p, i) && Agi_ObjLit0(p, i) > Agi_ObjLit1(p, i);  } 
func Agi_ObjIsMux(p *Agi_Man_t, i int32) int32 { return Agi_ObjIsAnd(p, i) && ~Agi_ObjLit2(p, i);        }
func Agi_ObjIsMaj(p *Agi_Man_t, i int32) int32 { return Agi_ObjIsXor(p, i) && ~Agi_ObjLit2(p, i);        } 

func Agi_ManAppendObj(p *Agi_Man_t) int32 {
    assert(p.nObjs < p.nCap, "p.nObjs < p.nCap");
    return p.nObjs++; // return var
}
func Agi_ManAppendCi(p *Agi_Man_t) int32 {
    var iObj int32  = Agi_ManAppendObj( p );
    p.pObjs[iObj] = AGI_PI | (word)Vec_IntSize(&p.vCis);
    Vec_IntPush( &p.vCis, iObj );
    return Abc_Var2Lit( iObj, 0 ); // return lit
}
func Agi_ManAppendCo(p *Agi_Man_t, iLit0 int32) int32 {
    var iObj int32  = Agi_ManAppendObj( p );
    p.pObjs[iObj] = AGI_PO | (word)iLit0;
    Vec_IntPush( &p.vCos, iObj );
    return Abc_Var2Lit( iObj, 0 ); // return lit
}
func Agi_ManAppendAnd(p *Agi_Man_t, iLit0 int32, iLit1 int32) int32 {
    var iObj int32  = Agi_ManAppendObj( p );
    assert(iLit0 < iLit1, "iLit0 < iLit1");
    p.pObjs[iObj] = ((word)iLit1 << 32) | (word)iLit0;
    p.nNodes++;
    return Abc_Var2Lit( iObj, 0 ); // return lit
}

// Original macro:
// #define Agi_ManForEachCi( p, iCi, i )   Vec_IntForEachEntry( &p->vCis, iCi, i )

// Original macro:
// #define Agi_ManForEachCo( p, iCo, i )   Vec_IntForEachEntry( &p->vCos, iCo, i )

// Original macro:
// #define Agi_ManForEachObj( p, i )       for ( i = 0; i < Agi_ManObjNum(p); i++ )

// Original macro:
// #define Agi_ManForEachObj1( p, i )      for ( i = 1; i < Agi_ManObjNum(p); i++ )

// Original macro:
// #define Agi_ManForEachNode( p, i )      for ( i = 1; i < Agi_ManObjNum(p); i++ ) if ( !Agi_ObjIsNode(p, i) ) {} else


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Agi_ManAlloc(nCap int32) *Agi_Man_t {
    
    nCap = Abc_MaxInt( nCap, 16 );
    // inlining of macro ABC_CALLOC
    p := &Agi_Man_t{};
    p.nCap     = nCap;
    // inlining of macro ABC_CALLOC
    p.pObjs    = ((word *) calloc((size_t)(nCap), sizeof(word)));
    // inlining of macro ABC_CALLOC
    p.pTravIds = ((unsigned *) calloc((size_t)(nCap), sizeof(unsigned)));
    p.pObjs[0] = AGI_C0;
    p.nObjs    = 1;
    return p;
}
func Agi_ManFree(p *Agi_Man_t) {
    // inlining of macro ABC_FREE
    ((p.pObjs) ? (free((char *) (p.pObjs)), (p.pObjs) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.pTravIds) ? (free((char *) (p.pTravIds)), (p.pTravIds) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCis.pArray) ? (free((char *) (p.vCis.pArray)), (p.vCis.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p.vCos.pArray) ? (free((char *) (p.vCos.pArray)), (p.vCos.pArray) = 0) : 0);
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Agi_ManFromGia(p *Gia_Man_t) *Agi_Man_t {
    var pNew *Agi_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Agi_ManAlloc( Gia_ManObjNum(p) );
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Agi_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Agi_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else if Gia_ObjIsCi(pObj) {
            pObj.Value = Agi_ManAppendCi( pNew );
        } else { assert(false, "0"); }
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Agi_ManSuppSize_rec(p *Agi_Man_t, i int32) int32 {
    if p.pTravIds[i] == p.nTravIds {
        return 0; }
    p.pTravIds[i] = p.nTravIds;
    if Agi_ObjIsCi(p, i) != 0 {
        return 1; }
    assert((Agi_ObjIsAnd(p, i) != 0), "Agi_ObjIsAnd(p, i)");
    return Agi_ManSuppSize_rec( p, Agi_ObjVar0(p, i) ) +  Agi_ManSuppSize_rec( p, Agi_ObjVar1(p, i) );
}
func Agi_ManSuppSizeOne(p *Agi_Man_t, i int32) int32 {
    p.nTravIds++;
    return Agi_ManSuppSize_rec( p, i );
}
func Agi_ManSuppSizeTest(p *Agi_Man_t) int32 {
    var clk abctime  = Abc_Clock();
    var i, Counter int32;
    // inlining of macro Agi_ManForEachNode
    for  i = 1; i < Agi_ManObjNum(p); i++  { if !(Agi_ObjIsNode(p, i) != 0) {} else {
        Counter += (Agi_ManSuppSizeOne(p, i) <= 16); }
    }
    fmt.Printf("Nodes with small support %d (out of %d)\n", Counter, Agi_ManNodeNum(p));
    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    return Counter;

}
func Agi_ManTest(pGia *Gia_Man_t) {
    extern int Gia_ManSuppSizeTest( Gia_Man_t * p );
    var p *Agi_Man_t;
    Gia_ManSuppSizeTest( pGia );
    p = Agi_ManFromGia( pGia );
    Agi_ManSuppSizeTest( p );
    Agi_ManFree( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

