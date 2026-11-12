package abc

import (
	"fmt"
	"strconv"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [abcHieNew.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [New hierarchy manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: abcHieNew.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/util/utilNam.h"
//#include "misc/extra/extra.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const AU_MAX_FANINS = 0x1FFFFFFF
 
 type Au_Type_t int32
const ( 
    AU_OBJ_NONE Au_Type_t = iota
           // 0: non-existent object
    AU_OBJ_CONST0
         // 1: constant node
    AU_OBJ_PI
             // 2: primary input
    AU_OBJ_PO
             // 3: primary output
    AU_OBJ_FAN
            // 4: box output
    AU_OBJ_FLOP
           // 5: flip-flop
    AU_OBJ_BOX
            // 6: box
    AU_OBJ_NODE
           // 7: logic node
    AU_OBJ_VOID            // 8: placeholder

)

 


type Au_Man_t =   Au_Man_t_;
type Au_Ntk_t =   Au_Ntk_t_;
     

type Au_Obj_t struct { // 16 bytes

    Func uint32; /* C bitfield: 30 bits; review packing and truncation */       // functionality
    Value uint32; /* C bitfield: 2 bits; review packing and truncation */       // node value
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */       // object type
    nFanins uint32; /* C bitfield: 29 bits; review packing and truncation */       // fanin count (related to AU_MAX_FANIN_NUM)
    Fanins [2]int32;          // fanin literals
}

type Au_Ntk_t_ struct {
    pName                  string;              // model name
    pMan              *Au_Man_t;               // model manager
    Id                    int32;                 // model ID
    // objects
    vPis              Vec_Int_t;               // primary inputs (CI id -> handle)
    vPos              Vec_Int_t;               // primary outputs (CI id -> handle)
    vObjs              Vec_Int_t;              // internal nodes (obj id -> handle)
    nObjs [AU_OBJ_VOID]int32; // counter of objects of each type
    // memory for objects
    vChunks             *Vec_Ptr_t;            // memory pages
    vPages              Vec_Ptr_t;             // memory pages
    iHandle                    int32;            // currently available ID
    nObjsAlloc                    int32;         // the total number of objects allocated
    nObjsUsed                    int32;          // the number of useful entries
    // object attributes
    nTravIds                    int32;           // counter of traversal IDs
    vTravIds              Vec_Int_t;           // trav IDs of the objects
    vCopies              Vec_Int_t;            // object copies
    // structural hashing
    nHTable                    int32;            // hash table size
    pHTable                   *int32;            // hash table
    pConst0              *Au_Obj_t;            // constant node
    // statistics
    fMark                    int32;
    nBoxes                 float64;
    nNodes                 float64;
    nPorts                 float64;
    nNodeAnds                 float64;
    nNodeXors                 float64;
    nNodeMuxs                 float64;
}

type Au_Man_t_ struct {
    pName                  string;              // the name of the library
    vNtks              Vec_Ptr_t;              // the array of modules
    pFuncs             *Abc_Nam_t;             // hashing functions into integers
    nRefs                    int32;              // reference counter
    // statistics
    nGiaObjMax                    int32;         // max number of GIA objects
    nPortsC0                 float64;           // const ports
    nPortsC1                 float64;           // const ports
    nPortsNC                 float64;           // non-const ports
}


func Au_Var2Lit(Var int32, fCompl int32) int32 { return Var + Var + fCompl;                       }
func Au_Lit2Var(Lit int32) int32 { return Lit >> 1;                                 }
func Au_LitIsCompl(Lit int32) int32 { return Lit & 1;                                  }
func Au_LitNot(Lit int32) int32 { return Lit ^ 1;                                  }
func Au_LitNotCond(Lit int32, c int32) int32 { return Lit ^ int32((c > 0));                       }
func Au_LitRegular(Lit int32) int32 { return Lit & ~01;                                }

func Au_Regular(p *Au_Obj_t) *Au_Obj_t { return (Au_Obj_t *)((ABC_PTRUINT_T)(p) & ~01);   }
func Au_Not(p *Au_Obj_t) *Au_Obj_t { return (Au_Obj_t *)((ABC_PTRUINT_T)(p) ^  01);   }
func Au_NotCond(p *Au_Obj_t, c int32) *Au_Obj_t { return (Au_Obj_t *)((ABC_PTRUINT_T)(p) ^ (c));   }
func Au_IsComplement(p *Au_Obj_t) int32 { return int32(((ABC_PTRUINT_T)(p) & 01));           }
 
func Au_UtilStrsav(s string) string { if s {
    // inlining of macro ABC_ALLOC
    return strcpy(((char *) malloc(sizeof(char) * (size_t)(len(s)+1))), s);
} else {
    return nil;
}             }

func Au_ManName(p *Au_Man_t) string { return p.pName;                                                       }
func Au_ManNtkNum(p *Au_Man_t) int32 { return Vec_PtrSize(&p.vNtks) - 1;                                     }
func Au_ManNtk(p *Au_Man_t, i int32) *Au_Ntk_t { return (Au_Ntk_t *)Vec_PtrEntry(&p.vNtks, i);                         }
func Au_ManNtkRoot(p *Au_Man_t) *Au_Ntk_t { return Au_ManNtk( p, 1 );                                              }

func Au_NtkName(p *Au_Ntk_t) string { return p.pName;                                                       }
func Au_NtkMan(p *Au_Ntk_t) *Au_Man_t { return p.pMan;                                                        }
func Au_NtkPiNum(p *Au_Ntk_t) int32 { return p.nObjs[AU_OBJ_PI];                                            } 
func Au_NtkPoNum(p *Au_Ntk_t) int32 { return p.nObjs[AU_OBJ_PO];                                            } 
func Au_NtkFanNum(p *Au_Ntk_t) int32 { return p.nObjs[AU_OBJ_FAN];                                           } 
func Au_NtkFlopNum(p *Au_Ntk_t) int32 { return p.nObjs[AU_OBJ_FLOP];                                          } 
func Au_NtkBoxNum(p *Au_Ntk_t) int32 { return p.nObjs[AU_OBJ_BOX];                                           } 
func Au_NtkNodeNum(p *Au_Ntk_t) int32 { return p.nObjs[AU_OBJ_NODE];                                          } 
func Au_NtkObjNumMax(p *Au_Ntk_t) int32 { return (Vec_PtrSize(&p.vPages) - 1) * (1 << 12) + p.iHandle;         } 
func Au_NtkObjNum(p *Au_Ntk_t) int32 { return Vec_IntSize(&p.vObjs);                                         } 
func Au_NtkObj(p *Au_Ntk_t, h int32) *Au_Obj_t { return (Au_Obj_t *)p.vPages.pArray[h >> 12] + (h & 0xFFF);            }

func Au_NtkPi(p *Au_Ntk_t, i int32) *Au_Obj_t { return Au_NtkObj(p, Vec_IntEntry(&p.vPis, i));                        }
func Au_NtkPo(p *Au_Ntk_t, i int32) *Au_Obj_t { return Au_NtkObj(p, Vec_IntEntry(&p.vPos, i));                        }
func Au_NtkObjI(p *Au_Ntk_t, i int32) *Au_Obj_t { return Au_NtkObj(p, Vec_IntEntry(&p.vObjs, i));                       }

func Au_ObjIsNone(p *Au_Obj_t) int32 { return p.Type == AU_OBJ_NONE;                                         } 
func Au_ObjIsConst0(p *Au_Obj_t) int32 { return p.Type == AU_OBJ_CONST0;                                       } 
func Au_ObjIsPi(p *Au_Obj_t) int32 { return p.Type == AU_OBJ_PI;                                           } 
func Au_ObjIsPo(p *Au_Obj_t) int32 { return p.Type == AU_OBJ_PO;                                           } 
func Au_ObjIsFan(p *Au_Obj_t) int32 { return p.Type == AU_OBJ_FAN;                                          } 
func Au_ObjIsFlop(p *Au_Obj_t) int32 { return p.Type == AU_OBJ_FLOP;                                         } 
func Au_ObjIsBox(p *Au_Obj_t) int32 { return p.Type == AU_OBJ_BOX;                                          } 
func Au_ObjIsNode(p *Au_Obj_t) int32 { return p.Type == AU_OBJ_NODE;                                         }
func Au_ObjIsTerm(p *Au_Obj_t) int32 { return p.Type >= AU_OBJ_PI && p.Type <= AU_OBJ_FLOP;                 } 

func Au_ObjBase(p *Au_Obj_t) string { return (char *)p - ((ABC_PTRINT_T)p & 0x3FF);                          } 
func Au_ObjNtk(p *Au_Obj_t) *Au_Ntk_t { return ((Au_Ntk_t **)Au_ObjBase(p))[0];                                } 
func Au_ObjId(p *Au_Obj_t) int32 { return ((int *)Au_ObjBase(p))[2] | (((ABC_PTRINT_T)p & 0x3FF) >> 4);   }
func Au_ObjPioNum(p *Au_Obj_t) int32 { assert((Au_ObjIsTerm(p) != 0), "Au_ObjIsTerm(p)"); return p.Fanins[p.nFanins];                 }
func Au_ObjFunc(p *Au_Obj_t) int32 { return p.Func;                                                        }
func Au_ObjModel(p *Au_Obj_t) *Au_Ntk_t { assert((Au_ObjIsFan(p) != 0)||(Au_ObjIsBox(p) != 0), "Au_ObjIsFan(p)||Au_ObjIsBox(p)"); return Au_ManNtk(Au_NtkMan(Au_ObjNtk(p)), p.Func); }

func Au_ObjFaninNum(p *Au_Obj_t) int32 { return p.nFanins;                                                     }
func Au_ObjFaninId(p *Au_Obj_t, i int32) int32 { assert(i >= 0 && i < int32(p.nFanins) && p.Fanins[i], "i >= 0 && i < int32(p.nFanins) && p.Fanins[i]"); return Au_Lit2Var(p.Fanins[i]);     }
func Au_ObjFaninId0(p *Au_Obj_t) int32 { return Au_ObjFaninId(p, 0);                                                                 }
func Au_ObjFaninId1(p *Au_Obj_t) int32 { return Au_ObjFaninId(p, 1);                                                                 }
func Au_ObjFaninId2(p *Au_Obj_t) int32 { return Au_ObjFaninId(p, 2);                                                                 }
func Au_ObjFanin(p *Au_Obj_t, i int32) *Au_Obj_t { return Au_NtkObj(Au_ObjNtk(p), Au_ObjFaninId(p, i));                                        }
func Au_ObjFanin0(p *Au_Obj_t) *Au_Obj_t { return Au_ObjFanin( p, 0 );                                                                 }
func Au_ObjFanin1(p *Au_Obj_t) *Au_Obj_t { return Au_ObjFanin( p, 1 );                                                                 }
func Au_ObjFanin2(p *Au_Obj_t) *Au_Obj_t { return Au_ObjFanin( p, 2 );                                                                 }
func Au_ObjFaninC(p *Au_Obj_t, i int32) int32 { assert(i >= 0 && i < int32(p.nFanins) && p.Fanins[i], "i >= 0 && i < int32(p.nFanins) && p.Fanins[i]"); return Au_LitIsCompl(p.Fanins[i]);  }
func Au_ObjFaninC0(p *Au_Obj_t) int32 { return Au_ObjFaninC(p, 0);                                                                  }
func Au_ObjFaninC1(p *Au_Obj_t) int32 { return Au_ObjFaninC(p, 1);                                                                  }
func Au_ObjFaninC2(p *Au_Obj_t) int32 { return Au_ObjFaninC(p, 2);                                                                  }
func Au_ObjFaninLit(p *Au_Obj_t, i int32) int32 { assert(i >= 0 && i < int32(p.nFanins) && p.Fanins[i], "i >= 0 && i < int32(p.nFanins) && p.Fanins[i]"); return p.Fanins[i];                 }
func Au_ObjSetFanin(p *Au_Obj_t, i int32, f int32) { assert(f > 0 && p.Fanins[i] == 0, "f > 0 && p.Fanins[i] == 0"); p.Fanins[i] = Au_Var2Lit(f,0);                         }
func Au_ObjSetFaninLit(p *Au_Obj_t, i int32, f int32) { assert(f >= 0 && p.Fanins[i] == 0, "f >= 0 && p.Fanins[i] == 0"); p.Fanins[i] = f;                                      }

func Au_BoxFanoutNum(p *Au_Obj_t) int32 { assert((Au_ObjIsBox(p) != 0), "Au_ObjIsBox(p)"); return p.Fanins[p.nFanins];                                       }
func Au_BoxFanoutId(p *Au_Obj_t, i int32) int32 { assert(i >= 0 && i < Au_BoxFanoutNum(p), "i >= 0 && i < Au_BoxFanoutNum(p)"); return p.Fanins[p.nFanins+1+i];                 }
func Au_BoxFanout(p *Au_Obj_t, i int32) *Au_Obj_t { return Au_NtkObj(Au_ObjNtk(p), Au_BoxFanoutId(p, i));                                       }

func Au_ObjCopy(p *Au_Obj_t) int32 { return Vec_IntEntry( &Au_ObjNtk(p).vCopies, Au_ObjId(p) );                                 }
func Au_ObjSetCopy(p *Au_Obj_t, c int32) { Vec_IntWriteEntry( &Au_ObjNtk(p).vCopies, Au_ObjId(p), c );                                }

func Au_ObjFanout(p *Au_Obj_t, i int32) int32 { assert(p.Type == AU_OBJ_BOX && i >= 0 && i < p.Fanins[p.nFanins] && p.Fanins[i], "p.Type == AU_OBJ_BOX && i >= 0 && i < p.Fanins[p.nFanins] && p.Fanins[i]"); return p.Fanins[p.nFanins + 1 + i];             }
func Au_ObjSetFanout(p *Au_Obj_t, i int32, f int32) { assert(p.Type == AU_OBJ_BOX && i >= 0 && i < p.Fanins[p.nFanins] && p.Fanins[i] == 0 && f > 0, "p.Type == AU_OBJ_BOX && i >= 0 && i < p.Fanins[p.nFanins] && p.Fanins[i] == 0 && f > 0"); p.Fanins[p.nFanins + 1 + i] = f;  }

func Au_NtkIncrementTravId(p *Au_Ntk_t) { if p.vTravIds.pArray == nil { Vec_IntFill(&p.vTravIds, Au_NtkObjNumMax(p)+500, 0); }
 p.nTravIds++; assert(p.nTravIds < (1<<30), "p.nTravIds < (1<<30)");  }
func Au_ObjSetTravIdCurrent(p *Au_Obj_t) { Vec_IntSetEntry(&Au_ObjNtk(p).vTravIds, Au_ObjId(p), Au_ObjNtk(p).nTravIds );                        }
func Au_ObjSetTravIdPrevious(p *Au_Obj_t) { Vec_IntSetEntry(&Au_ObjNtk(p).vTravIds, Au_ObjId(p), Au_ObjNtk(p).nTravIds-1 );                      }
func Au_ObjIsTravIdCurrent(p *Au_Obj_t) int32 { return (Vec_IntGetEntry(&Au_ObjNtk(p).vTravIds, Au_ObjId(p)) == Au_ObjNtk(p).nTravIds);              }
func Au_ObjIsTravIdPrevious(p *Au_Obj_t) int32 { return (Vec_IntGetEntry(&Au_ObjNtk(p).vTravIds, Au_ObjId(p)) == Au_ObjNtk(p).nTravIds-1);            }
func Au_ObjSetTravIdCurrentId(p *Au_Ntk_t, Id int32) { Vec_IntSetEntry(&p.vTravIds, Id, p.nTravIds );                                                       }
func Au_ObjIsTravIdCurrentId(p *Au_Ntk_t, Id int32) int32 { return (Vec_IntGetEntry(&p.vTravIds, Id) == p.nTravIds);                                             }

// Original macro:
// #define Au_ManForEachNtk( p, pNtk, i )           \
//     for ( i = 1; (i < Vec_PtrSize(&p->vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++ ) 

// Original macro:
// #define Au_ManForEachNtkReverse( p, pNtk, i )    \
//     for ( i = Vec_PtrSize(&p->vNtks) - 1;(i>=1) && (((pNtk) = Au_ManNtk(p, i)), 1); i-- ) 


// Original macro:
// #define Au_ObjForEachFaninId( pObj, hFanin, i )  \
//     for ( i = 0; (i < Au_ObjFaninNum(pObj))     && (((hFanin) = Au_ObjFaninId(pObj, i)), 1); i++ )  

// Original macro:
// #define Au_BoxForEachFanoutId( pObj, hFanout, i) \
//     for ( i = 0; (i < Au_BoxFanoutNum(pObj))    && (((hFanout) = Au_BoxFanoutId(pObj, i)), 1); i++ )


// Original macro:
// #define Au_ObjForEachFanin( pObj, pFanin, i )    \
//     for ( i = 0; (i < Au_ObjFaninNum(pObj))     && (((pFanin) = Au_ObjFanin(pObj, i)), 1); i++ )  

// Original macro:
// #define Au_BoxForEachFanout( pObj, pFanout, i)   \
//     for ( i = 0; (i < Au_BoxFanoutNum(pObj))    && (((pFanout) = Au_BoxFanout(pObj, i)), 1); i++ )


// Original macro:
// #define Au_NtkForEachPi( p, pObj, i )            \
//     for ( i = 0; (i < Vec_IntSize(&p->vPis))    && (((pObj) = Au_NtkPi(p, i)), 1); i++ )

// Original macro:
// #define Au_NtkForEachPo( p, pObj, i )            \
//     for ( i = 0; (i < Vec_IntSize(&p->vPos))    && (((pObj) = Au_NtkPo(p, i)), 1); i++ )

// Original macro:
// #define Au_NtkForEachObj( p, pObj, i )           \
//     for ( i = 0; (i < Vec_IntSize(&p->vObjs))   && (((pObj) = Au_NtkObjI(p, i)), 1); i++ )

// Original macro:
// #define Au_NtkForEachNode( p, pObj, i )           \
//     for ( i = 0; (i < Vec_IntSize(&p->vObjs))   && (((pObj) = Au_NtkObjI(p, i)), 1); i++ ) if ( !Au_ObjIsNode(pObj) ) {} else

// Original macro:
// #define Au_NtkForEachBox( p, pObj, i )           \
//     for ( i = 0; (i < Vec_IntSize(&p->vObjs))   && (((pObj) = Au_NtkObjI(p, i)), 1); i++ ) if ( !Au_ObjIsBox(pObj) ) {} else




// extern void Au_ManAddNtk( Au_Man_t * pMan, Au_Ntk_t * p );

// extern void Au_ManFree( Au_Man_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Working with models.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkAlloc(pMan *Au_Man_t, pName string) *Au_Ntk_t {
    
    // inlining of macro ABC_CALLOC
    p := &Au_Ntk_t{};
    p.pName = Au_UtilStrsav( pName );
    p.vChunks = Vec_PtrAlloc( 111 );
    Vec_IntGrow( &p.vPis,  111 );
    Vec_IntGrow( &p.vPos,  111 );
    Vec_IntGrow( &p.vObjs, 1111 );
    Vec_PtrGrow( &p.vPages, 11 );
    Au_ManAddNtk( pMan, p );
    return p;
}
func Au_NtkFree(p *Au_Ntk_t) {
    Au_ManFree( p.pMan );
    Vec_PtrFreeFree( p.vChunks );
    p.vCopies.pArray = nil;
    p.vPages.pArray = nil;
    p.vObjs.pArray = nil;
    p.vPis.pArray = nil;
    p.vPos.pArray = nil;
    p.pHTable = nil;
    p.pName = nil;
    p = nil;
}
func Au_NtkMemUsage(p *Au_Ntk_t) int32 {
    var Mem int32  = sizeof(Au_Ntk_t);
    Mem += 4 * p.vPis.nCap;
    Mem += 4 * p.vPos.nCap;
    Mem += 4 * p.vObjs.nCap;
    Mem += 16 * p.nObjsAlloc;
    return Mem;
}
func Au_NtkPrintStats(p *Au_Ntk_t) {
    fmt.Printf("%-30s:", Au_NtkName(p));
    fmt.Printf(" i/o =%6d/%6d", Au_NtkPiNum(p), Au_NtkPoNum(p));
    if Au_NtkFlopNum(p) != 0 {
        fmt.Printf("  lat =%5d", Au_NtkFlopNum(p)); }
    fmt.Printf("  nd =%6d", Au_NtkNodeNum(p));
//    if ( Au_NtkBoxNum(p) )
        fmt.Printf("  box =%5d", Au_NtkBoxNum(p));
    fmt.Printf("  obj =%7d", Au_NtkObjNum(p));
//    printf( "  max =%7d",    Au_NtkObjNumMax(p) );
//    printf( "  use =%7d",    p->nObjsUsed );
    fmt.Printf(" %5.1f %%", 100.0 * (Au_NtkObjNumMax(p) - Au_NtkObjNum(p)) / Au_NtkObjNumMax(p));
    fmt.Printf(" %6.1f MB", 1.0 * Au_NtkMemUsage(p) / (1 << 20));
    fmt.Printf(" %5.1f %%", 100.0 * (p.nObjsAlloc - p.nObjsUsed) / p.nObjsAlloc);
    fmt.Printf("\n");
}
func Au_NtkCleanCopy(p *Au_Ntk_t) {
    Vec_IntFill( &p.vCopies, Au_NtkObjNumMax(p), -1 );
}
func Au_NtkNodeNumFunc(p *Au_Ntk_t, Func int32) int32 {
    var pObj *Au_Obj_t;
    var i, Counter int32;
    if p.pMan && p.pMan.pFuncs {
        return 0; }
    // inlining of macro Au_NtkForEachNode
    for  i = 0; (i < Vec_IntSize(&p.vObjs))   && (((pObj) = Au_NtkObjI(p, i)), 1); i++  { if !(Au_ObjIsNode(pObj) != 0) {} else {
        Counter += (pObj.Func == uint32(Func));
//        printf( "%d ", pObj->Func );
    }
    }
//    printf( "\n" );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Working with manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_ManAlloc(pName string) *Au_Man_t {
    
    // inlining of macro ABC_CALLOC
    p := &Au_Man_t{};
    p.pName = Au_UtilStrsav( pName );
    Vec_PtrGrow( &p.vNtks,  111 );
    Vec_PtrPush( &p.vNtks, nil );
    return p;
}
func Au_ManFree(p *Au_Man_t) {
    assert(p.nRefs > 0, "p.nRefs > 0");
    if --p.nRefs > 0 {
        return; }
    if p.pFuncs {
        Abc_NamStop( p.pFuncs ); }
    p.vNtks.pArray = nil;
    p.pName = nil;
    p = nil;
}
func Au_ManDelete(p *Au_Man_t) {
    var pNtk *Au_Ntk_t;
    var i int32;
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  { 
        Au_NtkFree( pNtk );
    }
}
func Au_ManFindNtk(p *Au_Man_t, pName string) int32 {
    var pNtk *Au_Ntk_t;
    var i int32;
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  { 
        if !(int32(strings.Compare(Au_NtkName(pNtk), pName)) != 0) {
            return i; }
    }
    return -1;
}
func Au_ManFindNtkP(p *Au_Man_t, pName string) *Au_Ntk_t {
    var iNtk int32  = Au_ManFindNtk( p, pName );
    if iNtk == -1 {
        return nil; }
    return Au_ManNtk( p, iNtk );
}
func Au_ManAddNtk(pMan *Au_Man_t, p *Au_Ntk_t) {
    assert(Au_ManFindNtk(pMan, Au_NtkName(p)) == -1, "Au_ManFindNtk(pMan, Au_NtkName(p)) == -1");
    p.pMan = pMan; pMan.nRefs++;
    p.Id = Vec_PtrSize( &pMan.vNtks );
    Vec_PtrPush( &pMan.vNtks, p );
}
func Au_ManMemUsage(p *Au_Man_t) int32 {
    var pNtk *Au_Ntk_t;
    var i, Mem int32;
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  { 
        Mem += 16 * pNtk.nObjsAlloc;
    }
    return Mem;
}
func Au_ManMemUsageUseful(p *Au_Man_t) int32 {
    var pNtk *Au_Ntk_t;
    var i, Mem int32;
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  { 
        Mem += 16 * pNtk.nObjsUsed;
    }
    return Mem;
}
func Au_ManPrintStats(p *Au_Man_t) {
    var pNtk *Au_Ntk_t;
    var i int32;
    if Vec_PtrSize(&p.vNtks) > 2 {
        fmt.Printf("Design %-13s\n", Au_ManName(p)); }
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  { 
        Au_NtkPrintStats( pNtk );
    }
    fmt.Printf("Different functions = %d. ", p.pFuncs ? Abc_NamObjNumMax(p.pFuncs) : 0);
    fmt.Printf("Memory = %.1f MB", 1.0 * Au_ManMemUsage(p) / (1 << 20));
    fmt.Printf(" %5.1f %%", 100.0 * (Au_ManMemUsage(p) - Au_ManMemUsageUseful(p)) / Au_ManMemUsage(p));
    fmt.Printf("\n");
//    if ( p->pFuncs )
//        Abc_NamPrint( p->pFuncs );
}

func Au_ManReorderModels_rec(pNtk *Au_Ntk_t, vOrder *Vec_Int_t) {
    var pBoxModel *Au_Ntk_t;
    var pObj *Au_Obj_t;
    var k int32;
    if pNtk.fMark {
        return; }
    pNtk.fMark = 1;
    // inlining of macro Au_NtkForEachBox
    for  k = 0; (k < Vec_IntSize(&pNtk.vObjs))   && (((pObj) = Au_NtkObjI(pNtk, k)), 1); k++  { if !(Au_ObjIsBox(pObj) != 0) {} else {
        pBoxModel = Au_ObjModel(pObj);
        if pBoxModel == nil || pBoxModel == pNtk {
            continue; }
        Au_ManReorderModels_rec( pBoxModel, vOrder );
    }
    }
    Vec_IntPush( vOrder, pNtk.Id );
}
func Au_ManReorderModels(p *Au_Man_t, pRoot *Au_Ntk_t) {
    var vNtksNew *Vec_Ptr_t;
    var vOrder, vTemp *Vec_Int_t;
    var pNtk, pBoxModel *Au_Ntk_t;
    var pBox, pFan *Au_Obj_t;
    var i, k, j, Entry int32;
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  { 
        pNtk.fMark = 0;
    }
    // collect networks in the DFS order
    vOrder = Vec_IntAlloc( Au_ManNtkNum(p)+1 );
    Vec_IntPush( vOrder, 0 );
    Au_ManReorderModels_rec( pRoot, vOrder );
    assert(Vec_IntEntryLast(vOrder) == pRoot.Id, "Vec_IntEntryLast(vOrder) == pRoot.Id");
    // add unconnected ones
    Vec_IntPop( vOrder );
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  { 
        if pNtk.fMark == 0 {
            Vec_IntPush( vOrder, pNtk.Id ); }
    }
    Vec_IntPush( vOrder, pRoot.Id );
    assert(Vec_IntSize(vOrder) == Au_ManNtkNum(p)+1, "Vec_IntSize(vOrder) == Au_ManNtkNum(p)+1");
    // reverse order
    vOrder.nSize--;
    vOrder.pArray++;
    Vec_IntReverseOrder( vOrder ); 
    vOrder.pArray--;
    vOrder.nSize++;
    // compute new order
    vNtksNew = Vec_PtrAlloc( Au_ManNtkNum(p)+1 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vOrder)) && (((Entry) = Vec_IntEntry(vOrder, i)), 1); i++  {
        Vec_PtrPush( vNtksNew, Au_ManNtk(p, Entry) );
    }
    // invert order
    assert(Vec_IntEntry(vOrder, 1) == pRoot.Id, "Vec_IntEntry(vOrder, 1) == pRoot.Id");
    vOrder = Vec_IntInvert( vTemp = vOrder, 0 );
    Vec_IntFree( vTemp );
    assert(Vec_IntEntry(vOrder, 1) == pRoot.Id, "Vec_IntEntry(vOrder, 1) == pRoot.Id");
    // update model numbers
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  {
        pNtk.Id = Vec_IntEntry( vOrder, pNtk.Id );
        // inlining of macro Au_NtkForEachBox
        for  k = 0; (k < Vec_IntSize(&pNtk.vObjs))   && (((pBox) = Au_NtkObjI(pNtk, k)), 1); k++  { if !(Au_ObjIsBox(pBox) != 0) {} else {
            pBox.Func = Vec_IntEntry( vOrder, pBox.Func );
            assert(pBox.Func > 0, "pBox.Func > 0");
            // inlining of macro Au_BoxForEachFanout
            for  j = 0; (j < Au_BoxFanoutNum(pBox))    && (((pFan) = Au_BoxFanout(pBox, j)), 1); j++  {
                pFan.Func = pBox.Func;
            }
        }
        }
    }
    // update
    p.vNtks.pArray = nil;
    p.vNtks.pArray = vNtksNew.pArray;
    vNtksNew.pArray = nil;
    Vec_PtrFree( vNtksNew );
    // verify
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&p.vNtks))   && (((pNtk) = Au_ManNtk(p, i)), 1); i++  { 
        // inlining of macro Au_NtkForEachBox
        for  k = 0; (k < Vec_IntSize(&pNtk.vObjs))   && (((pBox) = Au_NtkObjI(pNtk, k)), 1); k++  { if !(Au_ObjIsBox(pBox) != 0) {} else {
            pBoxModel = Au_ObjModel(pBox);
            if pBoxModel == nil || pBoxModel == pNtk {
                continue; }
            assert(!pBox.Func || pBox.Func >= uint32(pNtk.Id), "!pBox.Func || pBox.Func >= uint32(pNtk.Id)");
            assert(Au_ObjFaninNum(pBox) == Au_NtkPiNum(pBoxModel), "Au_ObjFaninNum(pBox) == Au_NtkPiNum(pBoxModel)");
            assert(Au_BoxFanoutNum(pBox) == Au_NtkPoNum(pBoxModel), "Au_BoxFanoutNum(pBox) == Au_NtkPoNum(pBoxModel)");
        }
        }
    }
    Vec_IntFree( vOrder );
}
func Au_ManCountThings(p *Au_Man_t) {
    var pNtk, pBoxModel *Au_Ntk_t;
    var pBox *Au_Obj_t;
    var i, k int32;//, clk = Abc_Clock();
    // inlining of macro Au_ManForEachNtkReverse
    for  i = Vec_PtrSize(&p.vNtks) - 1;(i>=1) && (((pNtk) = Au_ManNtk(p, i)), 1); i--  {
        pNtk.nBoxes = Au_NtkBoxNum(pNtk);
        pNtk.nNodes = Au_NtkNodeNum(pNtk);
        pNtk.nPorts = Au_NtkPiNum(pNtk) + Au_NtkPoNum(pNtk);
        pNtk.nNodeAnds = Au_NtkNodeNumFunc( pNtk, 1 );
        pNtk.nNodeXors = Au_NtkNodeNumFunc( pNtk, 2 );
        pNtk.nNodeMuxs = Au_NtkNodeNumFunc( pNtk, 3 );
//        assert( pNtk->nNodes == pNtk->nNodeAnds + pNtk->nNodeXors + pNtk->nNodeMuxs );
//        printf( "adding %.0f nodes of model %s\n", pNtk->nNodes, Au_NtkName(pNtk) );
        // inlining of macro Au_NtkForEachBox
        for  k = 0; (k < Vec_IntSize(&pNtk.vObjs))   && (((pBox) = Au_NtkObjI(pNtk, k)), 1); k++  { if !(Au_ObjIsBox(pBox) != 0) {} else {
            pBoxModel = Au_ObjModel(pBox);
            if pBoxModel == nil || pBoxModel == pNtk {
                continue; }
            assert(Au_ObjFaninNum(pBox) == Au_NtkPiNum(pBoxModel), "Au_ObjFaninNum(pBox) == Au_NtkPiNum(pBoxModel)");
            assert(Au_BoxFanoutNum(pBox) == Au_NtkPoNum(pBoxModel), "Au_BoxFanoutNum(pBox) == Au_NtkPoNum(pBoxModel)");
            assert(pBoxModel.Id > pNtk.Id, "pBoxModel.Id > pNtk.Id");
            assert(pBoxModel.nPorts > 0, "pBoxModel.nPorts > 0");
            pNtk.nBoxes += pBoxModel.nBoxes;
            pNtk.nNodes += pBoxModel.nNodes;
            pNtk.nPorts += pBoxModel.nPorts;
            pNtk.nNodeAnds += pBoxModel.nNodeAnds;
            pNtk.nNodeXors += pBoxModel.nNodeXors;
            pNtk.nNodeMuxs += pBoxModel.nNodeMuxs;
//            printf( "    adding %.0f nodes of model %s\n", pBoxModel->nNodes, Au_NtkName(pBoxModel) );
        }
        }
//        printf( "total %.0f nodes in model %s\n", pNtk->nNodes, Au_NtkName(pNtk) );
    }
    pNtk = Au_ManNtkRoot(p);
    fmt.Printf("Total nodes = %15.0f. Total instances = %15.0f. Total ports = %15.0f.\n", //    printf( "Total nodes = %.2e. Total instances = %.2e. Total ports = %.2e.\n", 
            pNtk.nNodes, pNtk.nBoxes, pNtk.nPorts);
//    printf( "Total ANDs  = %15.0f. Total XORs      = %15.0f. Total MUXes = %15.0f.\n", 
//    printf( "Total ANDs  = %.2e. Total XORs      = %.2e. Total MUXes = %.2e.  ", 
//        pNtk->nNodeAnds, pNtk->nNodeXors, pNtk->nNodeMuxs );
    fmt.Printf("Total ANDs  = %15.0f.\n", pNtk.nNodeAnds);
    fmt.Printf("Total XORs  = %15.0f.\n", pNtk.nNodeXors);
    fmt.Printf("Total MUXes = %15.0f.\n", pNtk.nNodeMuxs);
//    Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
}

func Au_NtkCompareNames(p1 **Au_Ntk_t, p2 **Au_Ntk_t) int32 {
    return int32(strings.Compare(Au_NtkName(*p1), Au_NtkName(*p2)));
}
func Au_ManPrintBoxInfo(pNtk *Au_Ntk_t) {
    var vMods *Vec_Ptr_t;
    var pModel, pBoxModel *Au_Ntk_t;
    var pObj *Au_Obj_t;
    var vCounts *Vec_Int_t;
    var i, k, Num int32;
    if pNtk.pMan == nil {
        fmt.Printf("There is no hierarchy information.\n");
        return;
    }
    vMods = &pNtk.pMan.vNtks;

/*
    vMods->nSize--;
    vMods->pArray++;
    // sort models by name
    Vec_PtrSort( vMods, (int(*)())Au_NtkCompareNames );
    // swap the first model
    Num = Vec_PtrFind( vMods, pNtk );
    assert( Num >= 0 && Num < Vec_PtrSize(vMods) );
    pBoxModel = (Au_Ntk_t *)Vec_PtrEntry(vMods, 0);
    Vec_PtrWriteEntry(vMods, 0, (Au_Ntk_t *)Vec_PtrEntry(vMods, Num) );
    Vec_PtrWriteEntry(vMods, Num, pBoxModel );
    vMods->pArray--;
    vMods->nSize++;
*/

//    Vec_PtrForEachEntry( Au_Ntk_t *, vMods, pModel, i )
//        printf( "%s\n", Au_NtkName(pModel) );

    // print models
    vCounts = Vec_IntStart( Vec_PtrSize(vMods) );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vMods)) && (((pModel) = (Au_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
        if Au_NtkBoxNum(pModel) == 0 {
            continue; }
        Vec_IntFill( vCounts, Vec_IntSize(vCounts), 0 );
        // inlining of macro Au_NtkForEachBox
        for  k = 0; (k < Vec_IntSize(&pModel.vObjs))   && (((pObj) = Au_NtkObjI(pModel, k)), 1); k++  { if !(Au_ObjIsBox(pObj) != 0) {} else {
            pBoxModel = Au_ObjModel(pObj);
            if pBoxModel == nil || pBoxModel == pModel {
                continue; }
            Num = Vec_PtrFind( vMods, pBoxModel );
            assert(Num >= 0 && Num < Vec_PtrSize(vMods), "Num >= 0 && Num < Vec_PtrSize(vMods)");
            Vec_IntAddToEntry( vCounts, Num, 1 );
        }
        }

//        Au_NtkPrintStats( pModel, 0, 0, 0, 0, 0, 0, 0 );
        fmt.Printf("MODULE  ");
        fmt.Printf("%-30s : ", Au_NtkName(pModel));
        fmt.Printf("PI=%6d ", Au_NtkPiNum(pModel));
        fmt.Printf("PO=%6d ", Au_NtkPoNum(pModel));
        fmt.Printf("BB=%6d ", Au_NtkBoxNum(pModel));
        fmt.Printf("ND=%6d ", Au_NtkNodeNum(pModel)); // sans constants
//        printf( "Lev=%5d ", Au_NtkLevel(pModel) );
        fmt.Printf("\n");

        // inlining of macro Vec_IntForEachEntry
        for  k = 0; (k < Vec_IntSize(vCounts)) && (((Num) = Vec_IntEntry(vCounts, k)), 1); k++  {
            if Num != 0 {
                fmt.Printf("%15d : %s\n", Num, Au_NtkName((Au_Ntk_t *)Vec_PtrEntry(vMods, k))); }
        }
    }
    Vec_IntFree( vCounts );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vMods)) && (((pModel) = (Au_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
        if Au_NtkBoxNum(pModel) != 0 {
            continue; }
        fmt.Printf("MODULE  ");
        fmt.Printf("%-30s : ", Au_NtkName(pModel));
        fmt.Printf("PI=%6d ", Au_NtkPiNum(pModel));
        fmt.Printf("PO=%6d ", Au_NtkPoNum(pModel));
        fmt.Printf("BB=%6d ", Au_NtkBoxNum(pModel));
        fmt.Printf("ND=%6d ", Au_NtkNodeNum(pModel));
//        printf( "Lev=%5d ", Au_NtkLevel(pModel) );
        fmt.Printf("\n");
    }
}
func Au_NtkCompareSign(p1 **Au_Ntk_t, p2 **Au_Ntk_t) int32 {
    if Au_NtkPiNum(*p1) - Au_NtkPiNum(*p2) != 0 {
        return Au_NtkPiNum(*p1) - Au_NtkPiNum(*p2);
    } else {
        return Au_NtkPoNum(*p1) - Au_NtkPoNum(*p2); }
}
func Au_ManPrintBoxInfoSorted(pNtk *Au_Ntk_t) {
    var vMods, vModsNew *Vec_Ptr_t;
    var pModel *Au_Ntk_t;
    var i int32;
    if pNtk.pMan == nil {
        fmt.Printf("There is no hierarchy information.\n");
        return;
    }
    vMods = &pNtk.pMan.vNtks;

    vMods.nSize--;
    vMods.pArray++;
    vModsNew = Vec_PtrDup( vMods );
    vMods.pArray--;
    vMods.nSize++;

    Vec_PtrSort( vModsNew, (int(*)())Au_NtkCompareSign );
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vModsNew)) && (((pModel) = (Au_Ntk_t *)Vec_PtrEntry(vModsNew, i)), 1); i++  {
        fmt.Printf("MODULE  ");
        fmt.Printf("%-30s : ", Au_NtkName(pModel));
        fmt.Printf("PI=%6d ", Au_NtkPiNum(pModel));
        fmt.Printf("PO=%6d ", Au_NtkPoNum(pModel));
        fmt.Printf("BB=%6d ", Au_NtkBoxNum(pModel));
        fmt.Printf("ND=%6d ", Au_NtkNodeNum(pModel));
        fmt.Printf("\n");
    }
    Vec_PtrFree( vModsNew );
}

func Au_NtkCheckRecursive(pNtk *Au_Ntk_t) int32 {
    var vMods *Vec_Ptr_t;
    var pModel *Au_Ntk_t;
    var pObj *Au_Obj_t;
    var i, k, RetValue int32;

    if pNtk.pMan == nil {
        fmt.Printf("There is no hierarchy information.\n");
        return RetValue;
    }

    vMods = &pNtk.pMan.vNtks;
    // inlining of macro Vec_PtrForEachEntryStart
    for  i = 1; (i < Vec_PtrSize(vMods)) && (((pModel) = (Au_Ntk_t *)Vec_PtrEntry(vMods, i)), 1); i++  {
        // inlining of macro Au_NtkForEachObj
        for  k = 0; (k < Vec_IntSize(&pModel.vObjs))   && (((pObj) = Au_NtkObjI(pModel, k)), 1); k++  {
            if (Au_ObjIsBox(pObj) != 0) && Au_ObjModel(pObj) == pModel {
                fmt.Printf("WARNING: Model \"%s\" contains a recursive definition.\n", Au_NtkName(pModel));
                RetValue = 1;
                break;
            }
        }
    }
    return RetValue;
}

// count the number of support variables
func Au_ObjSuppSize_rec(p *Au_Ntk_t, Id int32) int32 {
    var pObj *Au_Obj_t;
    var i, iFanin, Counter int32;
    if Au_ObjIsTravIdCurrentId(p, Id) != 0 {
        return 0; }
    Au_ObjSetTravIdCurrentId(p, Id);
    pObj = Au_NtkObj( p, Id );
    if Au_ObjIsPi(pObj) != 0 {
        return 1; }
    assert((Au_ObjIsNode(pObj) != 0) || (Au_ObjIsBox(pObj) != 0) || (Au_ObjIsFan(pObj) != 0), "Au_ObjIsNode(pObj) || Au_ObjIsBox(pObj) || Au_ObjIsFan(pObj)");
    // inlining of macro Au_ObjForEachFaninId
    for  i = 0; (i < Au_ObjFaninNum(pObj))     && (((iFanin) = Au_ObjFaninId(pObj, i)), 1); i++  {  
        Counter += Au_ObjSuppSize_rec( p, iFanin );
    }
    return Counter;
}
func Au_ObjSuppSize(pObj *Au_Obj_t) int32 {
    var p *Au_Ntk_t  = Au_ObjNtk(pObj);
    Au_NtkIncrementTravId( p );
    return Au_ObjSuppSize_rec( p, Au_ObjId(pObj) );
}
/*
// this version is 50% slower than above
int Au_ObjSuppSize_rec( Au_Obj_t * pObj )
{
    Au_Obj_t * pFanin;
    int i, Counter = 0;
    if ( Au_ObjIsTravIdCurrent(pObj) )
        return 0;
    Au_ObjSetTravIdCurrent(pObj);
    if ( Au_ObjIsPi(pObj) )
        return 1;
    assert( Au_ObjIsNode(pObj) || Au_ObjIsBox(pObj) || Au_ObjIsFan(pObj) );
    Au_ObjForEachFanin( pObj, pFanin, i )
        Counter += Au_ObjSuppSize_rec( pFanin );
    return Counter;
}
int Au_ObjSuppSize( Au_Obj_t * pObj )
{
    Au_NtkIncrementTravId( Au_ObjNtk(pObj) );
    return Au_ObjSuppSize_rec( pObj );
}
*/
func Au_NtkSuppSizeTest(p *Au_Ntk_t) int32 {
    var pObj *Au_Obj_t;
    var i, Counter int32;
    // inlining of macro Au_NtkForEachObj
    for  i = 0; (i < Vec_IntSize(&p.vObjs))   && (((pObj) = Au_NtkObjI(p, i)), 1); i++  {
        if Au_ObjIsNode(pObj) != 0 {
            Counter += (Au_ObjSuppSize(pObj) <= 16); }
    }
    fmt.Printf("Nodes with small support %d (out of %d)\n", Counter, Au_NtkNodeNum(p));
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns memory for the next object.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkInsertHeader(p *Au_Ntk_t) {
    var pMem *Au_Obj_t  = (Au_Obj_t *)Vec_PtrEntryLast( &p.vPages );
    assert((((ABC_PTRINT_T)(pMem + p.iHandle) & 0x3FF) >> 4) == 0, "(((ABC_PTRINT_T)(pMem + p.iHandle) & 0x3FF) >> 4) == 0");
    ((Au_Ntk_t **)(pMem + p.iHandle))[0] = p;
    ((int *)(pMem + p.iHandle))[2] = ((Vec_PtrSize(&p.vPages) - 1) << 12) | (p.iHandle & 0xFC0);
    p.iHandle++;
}
func Au_NtkAllocObj(p *Au_Ntk_t, nFanins int32, Type int32) int32 {
    var pMem, pObj, pTemp *Au_Obj_t;
    var nObjInt int32  = ((2+nFanins) >> 2) + (((2+nFanins) & 3) > 0);
    var Id int32;
    var nObjIntReal int32 = nObjInt;
    if nObjInt > 63 {
        nObjInt = 63 + 64 * (((nObjInt-63) >> 6) + (((nObjInt-63) & 63) > 0)); }
    if Vec_PtrSize(&p.vPages) == 0 || p.iHandle + nObjInt > (1 << 12) {
        if nObjInt + 64 > (1 << 12) {
            // inlining of macro ABC_CALLOC
            pMem = ((Au_Obj_t *) calloc((size_t)(nObjInt + 64), sizeof(Au_Obj_t)));
p.nObjsAlloc += nObjInt + 64;
        } else {
            // inlining of macro ABC_CALLOC
            pMem = ((Au_Obj_t *) calloc((size_t)((1 << 12) + 64), sizeof(Au_Obj_t)));
p.nObjsAlloc += (1 << 12) + 64; }
        Vec_PtrPush( p.vChunks, pMem );
        if (ABC_PTRINT_T)pMem & 0xF {
            pMem = (Au_Obj_t *)((char *)pMem + 16 - ((ABC_PTRINT_T)pMem & 0xF)); }
        assert(((ABC_PTRINT_T)pMem & 0xF) == 0, "((ABC_PTRINT_T)pMem & 0xF) == 0");
        p.iHandle = (((ABC_PTRINT_T)pMem & 0x3FF) >> 4);
        if p.iHandle {
            pMem += 64 - (p.iHandle & 63);
            p.iHandle = 0; 
        }
        Vec_PtrPush( &p.vPages, pMem );
        Au_NtkInsertHeader( p );
    
    } else {
        pMem = (Au_Obj_t *)Vec_PtrEntryLast( &p.vPages );
        if (p.iHandle & 63) == 0 || nObjInt > (64 - (p.iHandle & 63)) {
            if p.iHandle & 63 {
                p.iHandle += 64 - (p.iHandle & 63); } 
            Au_NtkInsertHeader( p );
        }
        if p.iHandle + nObjInt > (1 << 12) {
            return Au_NtkAllocObj( p, nFanins, Type ); }
    }
    pObj = pMem + p.iHandle;
    assert(*((int *)pObj) == 0, "*((int *)pObj) == 0");
    pObj.nFanins = nFanins;
    p.nObjs[pObj.Type = Type]++;
    if Type == AU_OBJ_PI {
        Au_ObjSetFaninLit( pObj, 0, Vec_IntSize(&p.vPis) );
        Vec_IntPush( &p.vPis, Au_ObjId(pObj) );
    
    } else if Type == AU_OBJ_PO {
        Au_ObjSetFaninLit( pObj, 1, Vec_IntSize(&p.vPos) );
        Vec_IntPush( &p.vPos, Au_ObjId(pObj) );
    }
    p.iHandle += nObjInt;
    p.nObjsUsed += nObjIntReal;

    Id = Au_ObjId(pObj);
    Vec_IntPush( &p.vObjs, Id );
    pTemp = Au_NtkObj( p, Id );
    assert(pTemp == pObj, "pTemp == pObj");
    return Id;
}
func Au_NtkCreateConst0(pNtk *Au_Ntk_t) int32 {
    return Au_NtkAllocObj( pNtk, 0, AU_OBJ_CONST0 );
}
func Au_NtkCreatePi(pNtk *Au_Ntk_t) int32 {
    return Au_NtkAllocObj( pNtk, 0, AU_OBJ_PI );
}
func Au_NtkCreatePo(pNtk *Au_Ntk_t, iFanin int32) int32 {
    var Id int32  = Au_NtkAllocObj( pNtk, 1, AU_OBJ_PO );
    if iFanin != 0 {
        Au_ObjSetFaninLit( Au_NtkObj(pNtk, Id), 0, iFanin ); }
    return Id;
}
func Au_NtkCreateFan(pNtk *Au_Ntk_t, iFanin int32, iFanout int32, iModel int32) int32 {
    var Id int32  = Au_NtkAllocObj( pNtk, 1, AU_OBJ_FAN );
    var p *Au_Obj_t  = Au_NtkObj( pNtk, Id );
    if iFanin != 0 {
        Au_ObjSetFaninLit( p, 0, iFanin ); }
    Au_ObjSetFaninLit( p, 1, iFanout );
    p.Func = iModel;
    return Id;
}
func Au_NtkCreateNode(pNtk *Au_Ntk_t, vFanins *Vec_Int_t, iFunc int32) int32 {
    var i, iFanin int32;
    var Id int32  = Au_NtkAllocObj( pNtk, Vec_IntSize(vFanins), AU_OBJ_NODE );
    var p *Au_Obj_t  = Au_NtkObj( pNtk, Id );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFanins)) && (((iFanin) = Vec_IntEntry(vFanins, i)), 1); i++  {
        Au_ObjSetFaninLit( p, i, iFanin );
    }
    p.Func = iFunc;
    return Id;
}
func Au_NtkCreateBox(pNtk *Au_Ntk_t, vFanins *Vec_Int_t, nFanouts int32, iModel int32) int32 {
    var i int32;
    var iFanin int32;
    var nFanins int32 = Vec_IntSize(vFanins);
    var Id int32  = Au_NtkAllocObj( pNtk, nFanins + 1 + nFanouts, AU_OBJ_BOX );
    var p *Au_Obj_t  = Au_NtkObj( pNtk, Id );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vFanins)) && (((iFanin) = Vec_IntEntry(vFanins, i)), 1); i++  {
        Au_ObjSetFaninLit( p, i, iFanin );
    }
    Au_ObjSetFaninLit( p, nFanins, nFanouts );
    for  i = 0; i < nFanouts; i++  {
        Au_ObjSetFaninLit( p, nFanins + 1 + i, Au_NtkCreateFan(pNtk, Au_Var2Lit(Id,0), i, iModel) );
    }
    p.nFanins = nFanins;
    p.Func = iModel;
    assert(iModel > 0, "iModel > 0");
    return Id;
}

/*
 * 0/1 would denote false/true respectively.
 * Signals would be even numbers, and negation would be handled by xor with 1.
 * The output signal for each gate or subckt could be implicitly generated just use the next signal number.
 * For ranges, we could use "start:cnt" to denote the sequence "start, start+2, ..., start + 2*(cnt- 1)".
    - "cnt" seems more intuitive when signals are restricted to even numbers.
 * We'd have subckts and specialized gates .and, .xor, and .mux.

Here is a small example:

.model test
.inputs 3 # Inputs 2 4 6
.subckt and3 3 1 2:3 # 8 is implicit output
.outputs 1 8
.end

.model and3
.inputs 3 # Inputs 2 4 6
.and 2 4 # 8 output
.and 6 8 # 10 output
.outputs 1 10
.end
*/

/**Function*************************************************************

  Synopsis    [Reads one entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkRemapNum(vNum2Obj *Vec_Int_t, Num int32) int32 {
    return Au_Var2Lit(Vec_IntEntry(vNum2Obj, Au_Lit2Var(Num)), Au_LitIsCompl(Num));
}
/**Function*************************************************************

  Synopsis    [Reads one entry.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkParseCBlifNum(vFanins *Vec_Int_t, pToken string, vNum2Obj *Vec_Int_t) {
    var pCur string;
    var Num1, Num2, i int32;
    assert(pToken[0] >= '0' && pToken[0] <= '9', "pToken[0] >= '0' && pToken[0] <= '9'");
    Num1 = strconv.Atoi(pToken);
    for  pCur = pToken; *pCur; pCur++  {
        if *pCur == ':' {
            Num2 = strconv.Atoi(pCur+1);
            for  i = 0; i < Num2; i++  {
                Vec_IntPush( vFanins, Au_NtkRemapNum(vNum2Obj, Num1 + i + i) );
            }
            return;
        
        } else if *pCur == '*' {
            Num2 = strconv.Atoi(pCur+1);
            for  i = 0; i < Num2; i++  {
                Vec_IntPush( vFanins, Au_NtkRemapNum(vNum2Obj, Num1) );
            }
            return;
        }
    }
    assert(*pCur == 0, "*pCur == 0");
    Vec_IntPush( vFanins, Au_NtkRemapNum(vNum2Obj, Num1) );
}

/**Function*************************************************************

  Synopsis    [Parses CBLIF file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkParseCBlif(pFileName string) *Au_Ntk_t {
    var pFile *os.File;
    var pMan *Au_Man_t;
    var pRoot *Au_Ntk_t  = nil;
    var pBox, pFan *Au_Obj_t;
    var pBuffer, pCur string;
    var vLines, vNum2Obj, vFanins *Vec_Int_t;
    var i, k, j, Id, nInputs, nOutputs int32;
    var Line, Num, Func int32;
    // read the file
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Cannot open file \"%s\".\n", pFileName);
        return nil;
    }
    pBuffer = Extra_FileRead( pFile );
    _ = (pFile).Close();
    // split into lines
    vLines = Vec_IntAlloc( 1000 );
    Vec_IntPush( vLines, 0 );
    for  pCur = pBuffer; *pCur; pCur++  {
        if *pCur == '\n' {
            *pCur = 0;
            Vec_IntPush( vLines, pCur - pBuffer + 1 );
        }
    }
    // start the manager
    pMan = Au_ManAlloc( pFileName );
    // parse the lines
    vNum2Obj = Vec_IntAlloc( 1000 );
    vFanins = Vec_IntAlloc( 1000 );
    // inlining of macro Vec_IntForEachEntry
    for  i = 0; (i < Vec_IntSize(vLines)) && (((Line) = Vec_IntEntry(vLines, i)), 1); i++  {
        pCur = strtok( pBuffer + Line, " \t\r" );
        if pCur == nil || *pCur == '#' {
            continue; }
        if *pCur != '.' {
            fmt.Printf("Cannot read directive in line %d: \"%s\".\n", i, pBuffer + Line);
            continue;
        }
        Vec_IntClear( vFanins );
        if !(int32(strings.Compare(pCur, ".and")) != 0) {
            for  k = 0; k < 2; k++  {
                pCur = strtok( nil, " \t\r" );
                Num  = strconv.Atoi(pCur);
                Vec_IntPush( vFanins, Au_NtkRemapNum(vNum2Obj, Num) );
            }
            Id = Au_NtkCreateNode( pRoot, vFanins, 1 );
            Vec_IntPush( vNum2Obj, Id );
        
        } else if !(int32(strings.Compare(pCur, ".xor")) != 0) {
            for  k = 0; k < 2; k++  {
                pCur = strtok( nil, " \t\r" );
                Num  = strconv.Atoi(pCur);
                Vec_IntPush( vFanins, Au_NtkRemapNum(vNum2Obj, Num) );
            }
            Id = Au_NtkCreateNode( pRoot, vFanins, 2 );
            Vec_IntPush( vNum2Obj, Id );
        
        } else if !(int32(strings.Compare(pCur, ".mux")) != 0) {
            for  k = 0; k < 3; k++  {
                pCur = strtok( nil, " \t\r" );
                Num  = strconv.Atoi(pCur);
                Vec_IntPush( vFanins, Au_NtkRemapNum(vNum2Obj, Num) );
            }
            Id = Au_NtkCreateNode( pRoot, vFanins, 3 );
            Vec_IntPush( vNum2Obj, Id );
        
        } else if !(int32(strings.Compare(pCur, ".subckt")) != 0) {
            pCur = strtok( nil, " \t\r" );
            Func = pCur - pBuffer;
            pCur = strtok( nil, " \t\r" );
            nInputs = strconv.Atoi(pCur);
            pCur = strtok( nil, " \t\r" );
            nOutputs = strconv.Atoi(pCur);
            for {
                pCur = strtok( nil, " \t\r" );
                if pCur == nil || *pCur == '#' {
                    break; }
                Au_NtkParseCBlifNum( vFanins, pCur, vNum2Obj );
            }
            assert(Vec_IntSize(vFanins) == nInputs, "Vec_IntSize(vFanins) == nInputs");
            Id = Au_NtkCreateBox( pRoot, vFanins, nOutputs, Func );
            pBox = Au_NtkObj( pRoot, Id );
            // inlining of macro Au_BoxForEachFanoutId
            for  k = 0; (k < Au_BoxFanoutNum(pBox))    && (((Num) = Au_BoxFanoutId(pBox, k)), 1); k++  {
                Vec_IntPush( vNum2Obj, Num );
            }
        
        } else if !(int32(strings.Compare(pCur, ".model")) != 0) {
            pCur  = strtok( nil, " \t\r" );
            pRoot = Au_NtkAlloc( pMan, pCur );
            Id    = Au_NtkCreateConst0( pRoot );
            Vec_IntClear( vNum2Obj );
            Vec_IntPush( vNum2Obj, Id );
        
        } else if !(int32(strings.Compare(pCur, ".inputs")) != 0) {
            pCur = strtok( nil, " \t\r" );
            Num  = strconv.Atoi(pCur);
            for  k = 0; k < Num; k++  {
                Vec_IntPush( vNum2Obj, Au_NtkCreatePi(pRoot) );
            }
        
        } else if !(int32(strings.Compare(pCur, ".outputs")) != 0) {
            pCur = strtok( nil, " \t\r" );
            nOutputs = strconv.Atoi(pCur);
            for {
                pCur = strtok( nil, " \t\r" );
                if pCur == nil || *pCur == '#' {
                    break; } 
                Au_NtkParseCBlifNum( vFanins, pCur, vNum2Obj );
            }
            assert(Vec_IntSize(vFanins) == nOutputs, "Vec_IntSize(vFanins) == nOutputs");
            // inlining of macro Vec_IntForEachEntry
            for  k = 0; (k < Vec_IntSize(vFanins)) && (((Num) = Vec_IntEntry(vFanins, k)), 1); k++  {
                Au_NtkCreatePo( pRoot, Num );
            }
        
        } else if int32(strings.Compare(pCur, ".end")) != 0 {
            fmt.Printf("Unknown directive in line %d: \"%s\".\n", i, pBuffer + Line); }
    }
    Vec_IntFree( vFanins );
    Vec_IntFree( vNum2Obj );
    Vec_IntFree( vLines );
    // set pointers to models
    // inlining of macro Au_ManForEachNtk
    for  i = 1; (i < Vec_PtrSize(&pMan.vNtks))   && (((pRoot) = Au_ManNtk(pMan, i)), 1); i++  { 
        // inlining of macro Au_NtkForEachBox
        for  k = 0; (k < Vec_IntSize(&pRoot.vObjs))   && (((pBox) = Au_NtkObjI(pRoot, k)), 1); k++  { if !(Au_ObjIsBox(pBox) != 0) {} else {
            pBox.Func = Au_ManFindNtk( pMan, pBuffer + pBox.Func );
            assert(pBox.Func > 0, "pBox.Func > 0");
            // inlining of macro Au_BoxForEachFanout
            for  j = 0; (j < Au_BoxFanoutNum(pBox))    && (((pFan) = Au_BoxFanout(pBox, j)), 1); j++  {
                pFan.Func = pBox.Func;
            }
        }
        }
    }
    pBuffer = nil;
    // order models in topological order
    pRoot = Au_ManNtkRoot( pMan );
    Au_ManReorderModels( pMan, pRoot );
    return pRoot;
}

//ABC_NAMESPACE_IMPL_END

//#include "abc.h"
//#include "aig/gia/gia.h"

//ABC_NAMESPACE_IMPL_START

// extern Vec_Ptr_t * Abc_NtkDfsBoxes( Abc_Ntk_t * pNtk );

// extern int Abc_NtkDeriveFlatGiaSop( Gia_Man_t * pGia, int * gFanins, char * pSop );

// extern int Abc_NtkCheckRecursive( Abc_Ntk_t * pNtk );


/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkDeriveFlatGia_rec(pGia *Gia_Man_t, p *Au_Ntk_t) { 
    var pObj, pTerm *Au_Obj_t;
    var i, k, Lit int32;
    // inlining of macro Au_NtkForEachPi
    for  i = 0; (i < Vec_IntSize(&p.vPis))    && (((pTerm) = Au_NtkPi(p, i)), 1); i++  {
        assert(Au_ObjCopy(pTerm) >= 0, "Au_ObjCopy(pTerm) >= 0");
    }
    if int32(strings.Compare(Au_NtkName(p), "ref_egcd")) == 0 {
        fmt.Printf("Replacing one instance of recursive model \"%s\" by a black box.\n", "ref_egcd");
        // inlining of macro Au_NtkForEachPo
        for  i = 0; (i < Vec_IntSize(&p.vPos))    && (((pTerm) = Au_NtkPo(p, i)), 1); i++  {
            Au_ObjSetCopy( pTerm, Gia_ManAppendCi(pGia) );
        }
        return;
    }
    // inlining of macro Au_NtkForEachObj
    for  i = 0; (i < Vec_IntSize(&p.vObjs))   && (((pObj) = Au_NtkObjI(p, i)), 1); i++  {
        if Au_ObjIsNode(pObj) != 0 {
            if p.pMan.pFuncs {
                var gFanins [16]int32;
                var pSop string  = Abc_NamStr( p.pMan.pFuncs, pObj.Func );
                assert(Au_ObjFaninNum(pObj) <= 16, "Au_ObjFaninNum(pObj) <= 16");
                assert(Au_ObjFaninNum(pObj) == Abc_SopGetVarNum(pSop), "Au_ObjFaninNum(pObj) == Abc_SopGetVarNum(pSop)");
                // inlining of macro Au_ObjForEachFanin
                for  k = 0; (k < Au_ObjFaninNum(pObj))     && (((pTerm) = Au_ObjFanin(pObj, k)), 1); k++  {
                    gFanins[k] = Au_ObjCopy(pTerm);
                    assert(gFanins[k] >= 0, "gFanins[k] >= 0");
                }
                Lit = Abc_NtkDeriveFlatGiaSop( pGia, gFanins, pSop );
            
            } else {
                var Lit0, Lit1, Lit2 int32;
                assert(pObj.Func >= 1 && pObj.Func <= 3, "pObj.Func >= 1 && pObj.Func <= 3");
                Lit0 = Abc_LitNotCond( Au_ObjCopy(Au_ObjFanin0(pObj)), Au_ObjFaninC0(pObj) );
                Lit1 = Abc_LitNotCond( Au_ObjCopy(Au_ObjFanin1(pObj)), Au_ObjFaninC1(pObj) );
                if pObj.Func == 1 {
                    Lit = Gia_ManHashAnd( pGia, Lit0, Lit1 );
                } else if pObj.Func == 2 {
                    Lit = Gia_ManHashXor( pGia, Lit0, Lit1 );
                } else if pObj.Func == 3 {
                    Lit2 = Abc_LitNotCond( Au_ObjCopy(Au_ObjFanin2(pObj)), Au_ObjFaninC2(pObj) );
                    Lit = Gia_ManHashMux( pGia, Lit0, Lit1, Lit2 );
                
                } else { assert(false, "0"); } 
            } 
            assert(Lit >= 0, "Lit >= 0");
            Au_ObjSetCopy( pObj, Lit );
        
        } else if Au_ObjIsBox(pObj) != 0 {
            var pModel *Au_Ntk_t  = Au_ObjModel(pObj);
            Au_NtkCleanCopy( pModel );
            // check the match between the number of actual and formal parameters
            assert(Au_ObjFaninNum(pObj) == Au_NtkPiNum(pModel), "Au_ObjFaninNum(pObj) == Au_NtkPiNum(pModel)");
            assert(Au_BoxFanoutNum(pObj) == Au_NtkPoNum(pModel), "Au_BoxFanoutNum(pObj) == Au_NtkPoNum(pModel)");
            // assign PIs
            // inlining of macro Au_ObjForEachFanin
            for  k = 0; (k < Au_ObjFaninNum(pObj))     && (((pTerm) = Au_ObjFanin(pObj, k)), 1); k++  {  
                Au_ObjSetCopy( Au_NtkPi(pModel, k), Au_ObjCopy(pTerm) );
            }
            // call recursively
            Au_NtkDeriveFlatGia_rec( pGia, pModel );
            // assign POs
            // inlining of macro Au_BoxForEachFanout
            for  k = 0; (k < Au_BoxFanoutNum(pObj))    && (((pTerm) = Au_BoxFanout(pObj, k)), 1); k++  {
                Au_ObjSetCopy( pTerm, Au_ObjCopy(Au_NtkPo(pModel, k)) );
            }
        
        } else if Au_ObjIsConst0(pObj) != 0 {
            Au_ObjSetCopy( pObj, 0 ); }
            
    }
    // inlining of macro Au_NtkForEachPo
    for  i = 0; (i < Vec_IntSize(&p.vPos))    && (((pTerm) = Au_NtkPo(p, i)), 1); i++  {
        Lit = Abc_LitNotCond( Au_ObjCopy(Au_ObjFanin0(pTerm)), Au_ObjFaninC0(pTerm) );
        Au_ObjSetCopy( pTerm, Lit );
    }
    // inlining of macro Au_NtkForEachPo
    for  i = 0; (i < Vec_IntSize(&p.vPos))    && (((pTerm) = Au_NtkPo(p, i)), 1); i++  {
        assert(Au_ObjCopy(pTerm) >= 0, "Au_ObjCopy(pTerm) >= 0");
    }
//    p->pMan->nGiaObjMax = Abc_MaxInt( p->pMan->nGiaObjMax, Gia_ManObjNum(pGia) );
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkDeriveFlatGia(p *Au_Ntk_t) *Gia_Man_t {
    var pTemp *Gia_Man_t;
    var pGia *Gia_Man_t = nil;
    var pTerm *Au_Obj_t;
    var i int32;
    fmt.Printf("Collapsing model \"%s\"...\n", Au_NtkName(p));
    Au_NtkCleanCopy( p );
    // start the network
    pGia = Gia_ManStart( (1<<16) );
    pGia.pName = Abc_UtilStrsav( Au_NtkName(p) );
//    pGia->pSpec = Abc_UtilStrsav( Au_NtkSpec(p) );
    Gia_ManHashAlloc( pGia );
    Gia_ManFlipVerbose( pGia );
    // create PIs
    // inlining of macro Au_NtkForEachPi
    for  i = 0; (i < Vec_IntSize(&p.vPis))    && (((pTerm) = Au_NtkPi(p, i)), 1); i++  {
        Au_ObjSetCopy( pTerm, Gia_ManAppendCi(pGia) );
    }
    // recursively flatten hierarchy
    Au_NtkDeriveFlatGia_rec( pGia, p );
    // create POs
    // inlining of macro Au_NtkForEachPo
    for  i = 0; (i < Vec_IntSize(&p.vPos))    && (((pTerm) = Au_NtkPo(p, i)), 1); i++  {
        Gia_ManAppendCo( pGia, Au_ObjCopy(pTerm) );
    }
    // prepare return value
//    Gia_ManHashProfile( pGia );
    Gia_ManHashStop( pGia );
    Gia_ManSetRegNum( pGia, 0 );
    pGia = Gia_ManCleanup( pTemp = pGia );
    Gia_ManStop( pTemp );
    return pGia;
}


// ternary simulation
const AU_VAL0 =   1
const AU_VAL1 =   2
const AU_VALX =   3

func Au_ObjSetXsim(pObj *Au_Obj_t, Value int32) { pObj.Value = Value;  }
func Au_ObjGetXsim(pObj *Au_Obj_t) int32 { return pObj.Value;   }
func Au_XsimInv(Value int32) int32 { 
    if Value == AU_VAL0 {
        return AU_VAL1; }
    if Value == AU_VAL1 {
        return AU_VAL0; }
    assert(Value == AU_VALX, "Value == AU_VALX");       
    return AU_VALX;
}
func Au_XsimAnd(Value0 int32, Value1 int32) int32 { 
    if Value0 == AU_VAL0 || Value1 == AU_VAL0 {
        return AU_VAL0; }
    if Value0 == AU_VALX || Value1 == AU_VALX {
        return AU_VALX; }
    assert(Value0 == AU_VAL1 && Value1 == AU_VAL1, "Value0 == AU_VAL1 && Value1 == AU_VAL1");
    return AU_VAL1;
}
func Au_XsimXor(Value0 int32, Value1 int32) int32 { 
    if Value0 == AU_VALX || Value1 == AU_VALX {
        return AU_VALX; }
    if (Value0 == AU_VAL0) == (Value1 == AU_VAL0) {
        return AU_VAL0; }
    return AU_VAL1;
}
func Au_XsimMux(ValueC int32, Value1 int32, Value0 int32) int32 { 
    if ValueC == AU_VAL0 {
        return Value0; }
    if ValueC == AU_VAL1 {
        return Value1; }
    if Value0 == AU_VAL0 && Value1 == AU_VAL0 {
        return AU_VAL0; }
    if Value0 == AU_VAL1 && Value1 == AU_VAL1 {
        return AU_VAL1; }
    return AU_VALX;
}
func Au_ObjGetXsimFan0(pObj *Au_Obj_t) int32 { 
    var Value int32  = Au_ObjGetXsim( Au_ObjFanin0(pObj) );
    if Au_ObjFaninC0(pObj) != 0 {
        return Au_XsimInv(Value);
    } else {
        return Value;
    }
}
func Au_ObjGetXsimFan1(pObj *Au_Obj_t) int32 { 
    var Value int32  = Au_ObjGetXsim( Au_ObjFanin1(pObj) );
    if Au_ObjFaninC1(pObj) != 0 {
        return Au_XsimInv(Value);
    } else {
        return Value;
    }
}
func Au_ObjGetXsimFan2(pObj *Au_Obj_t) int32 { 
    var Value int32  = Au_ObjGetXsim( Au_ObjFanin2(pObj) );
    if Au_ObjFaninC2(pObj) != 0 {
        return Au_XsimInv(Value);
    } else {
        return Value;
    }
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkTerSimulate_rec(p *Au_Ntk_t) { 
    var pObj *Au_Obj_t = nil;
    var pTerm *Au_Obj_t;
    var i, k int32;
    // inlining of macro Au_NtkForEachPi
    for  i = 0; (i < Vec_IntSize(&p.vPis))    && (((pTerm) = Au_NtkPi(p, i)), 1); i++  {
        assert(Au_ObjGetXsim(pTerm) > 0, "Au_ObjGetXsim(pTerm) > 0");
        if Au_ObjGetXsim(pTerm) == AU_VALX {
            p.pMan.nPortsNC++;
        } else if Au_ObjGetXsim(pTerm) == AU_VAL0 {
            p.pMan.nPortsC0++;
        } else {
            p.pMan.nPortsC1++; }
    }
    if int32(strings.Compare(Au_NtkName(p), "ref_egcd")) == 0 {
        fmt.Printf("Replacing one instance of recursive model \"%s\" by a black box.\n", "ref_egcd");
        // inlining of macro Au_NtkForEachPo
        for  i = 0; (i < Vec_IntSize(&p.vPos))    && (((pTerm) = Au_NtkPo(p, i)), 1); i++  {
            Au_ObjSetXsim( pTerm, AU_VALX );
        }
        return;
    }
    // inlining of macro Au_NtkForEachObj
    for  i = 0; (i < Vec_IntSize(&p.vObjs))   && (((pObj) = Au_NtkObjI(p, i)), 1); i++  {
        if Au_ObjIsNode(pObj) != 0 {
            if pObj.Func == 1 {
                Au_ObjSetXsim( pObj, Au_XsimAnd(Au_ObjGetXsimFan0(pObj), Au_ObjGetXsimFan1(pObj)) );
            } else if pObj.Func == 2 {
                Au_ObjSetXsim( pObj, Au_XsimXor(Au_ObjGetXsimFan0(pObj), Au_ObjGetXsimFan1(pObj)) );
            } else if pObj.Func == 3 {
                Au_ObjSetXsim( pObj, Au_XsimMux(Au_ObjGetXsimFan0(pObj), Au_ObjGetXsimFan1(pObj), Au_ObjGetXsimFan2(pObj)) );
            } else { assert(false, "0"); }
        
        } else if Au_ObjIsBox(pObj) != 0 {
            var pModel *Au_Ntk_t  = Au_ObjModel(pObj);
            // check the match between the number of actual and formal parameters
            assert(Au_ObjFaninNum(pObj) == Au_NtkPiNum(pModel), "Au_ObjFaninNum(pObj) == Au_NtkPiNum(pModel)");
            assert(Au_BoxFanoutNum(pObj) == Au_NtkPoNum(pModel), "Au_BoxFanoutNum(pObj) == Au_NtkPoNum(pModel)");
            // assign PIs
            // inlining of macro Au_ObjForEachFanin
            for  k = 0; (k < Au_ObjFaninNum(pObj))     && (((pTerm) = Au_ObjFanin(pObj, k)), 1); k++  {  
                Au_ObjSetXsim( Au_NtkPi(pModel, k), Au_ObjGetXsim(pTerm) );
            }
            // call recursively
            Au_NtkTerSimulate_rec( pModel );
            // assign POs
            // inlining of macro Au_BoxForEachFanout
            for  k = 0; (k < Au_BoxFanoutNum(pObj))    && (((pTerm) = Au_BoxFanout(pObj, k)), 1); k++  {
                Au_ObjSetXsim( pTerm, Au_ObjGetXsim(Au_NtkPo(pModel, k)) );
            }
        
        } else if Au_ObjIsConst0(pObj) != 0 {
            Au_ObjSetXsim( pObj, AU_VAL0 ); }
            
    }
    // inlining of macro Au_NtkForEachPo
    for  i = 0; (i < Vec_IntSize(&p.vPos))    && (((pTerm) = Au_NtkPo(p, i)), 1); i++  {
        Au_ObjSetXsim( pTerm, Au_ObjGetXsimFan0(pObj) );
    }
    // inlining of macro Au_NtkForEachPo
    for  i = 0; (i < Vec_IntSize(&p.vPos))    && (((pTerm) = Au_NtkPo(p, i)), 1); i++  {
        assert(Au_ObjGetXsim(pTerm) > 0, "Au_ObjGetXsim(pTerm) > 0");
        if Au_ObjGetXsim(pTerm) == AU_VALX {
            p.pMan.nPortsNC++;
        } else if Au_ObjGetXsim(pTerm) == AU_VAL0 {
            p.pMan.nPortsC0++;
        } else {
            p.pMan.nPortsC1++; }
    }
}

/**Function*************************************************************

  Synopsis    [Flattens the logic hierarchy of the netlist.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkTerSimulate(p *Au_Ntk_t) {
    var pTerm *Au_Obj_t;
    var i int32;
    var Counter [2]int32 ;
    assert(p.pMan.pFuncs == nil, "p.pMan.pFuncs == nil");
    fmt.Printf("Collapsing model \"%s\"...\n", Au_NtkName(p));
    // create PIs
    // inlining of macro Au_NtkForEachPi
    for  i = 0; (i < Vec_IntSize(&p.vPis))    && (((pTerm) = Au_NtkPi(p, i)), 1); i++  {
        Au_ObjSetXsim( pTerm, AU_VALX );
    }
    // recursively flatten hierarchy
    p.pMan.nPortsC0 = 0;
    p.pMan.nPortsC1 = 0;
    p.pMan.nPortsNC = 0;
    Au_NtkTerSimulate_rec( p );
    // analyze outputs
    // inlining of macro Au_NtkForEachPo
    for  i = 0; (i < Vec_IntSize(&p.vPos))    && (((pTerm) = Au_NtkPo(p, i)), 1); i++  {
        if Au_ObjGetXsim(pTerm) == AU_VAL0 {
            Counter[0]++;
        } else if Au_ObjGetXsim(pTerm) == AU_VAL1 {
            Counter[1]++; }
    }
    // print results
    fmt.Printf("Const0 outputs =%15d. Const1 outputs =%15d.  Total outputs =%15d.\n", Counter[0], Counter[1], Au_NtkPoNum(p));
    fmt.Printf("Const0 ports =  %.0f. Const1  ports =  %.0f. Non-const ports=  %.0f.  Total ports =  %.0f.\n", p.pMan.nPortsC0, p.pMan.nPortsC1, p.pMan.nPortsNC, p.pMan.nPortsC0 + p.pMan.nPortsC1 + p.pMan.nPortsNC);
}


/**Function*************************************************************

  Synopsis    [Duplicates ABC network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Au_NtkDerive(pMan *Au_Man_t, pNtk *Abc_Ntk_t, vOrder *Vec_Ptr_t) *Au_Ntk_t {
    var p *Au_Ntk_t;
    var pAuObj *Au_Obj_t;
    var pObj, pTerm *Abc_Obj_t;
//    Vec_Ptr_t * vOrder;
    var vFanins *Vec_Int_t;
    var i, k, iFunc int32;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    Abc_NtkCleanCopy( pNtk );
    p = Au_NtkAlloc( pMan, Abc_NtkName(pNtk) );
    // copy PIs
    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        Abc_ObjFanout0(pTerm).iTemp = Au_NtkCreatePi(p);
    }
    // copy nodes and boxes
    vFanins = Vec_IntAlloc( 100 );
//    vOrder = Abc_NtkDfsBoxes( pNtk );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vOrder)) && (((pObj) = (Abc_Obj_t *)Vec_PtrEntry(vOrder, i)), 1); i++  {
        Vec_IntClear( vFanins );
        if Abc_ObjIsNode(pObj) {
            // inlining of macro Abc_ObjForEachFanin
            for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
                Vec_IntPush( vFanins, Au_Var2Lit(pTerm.iTemp, 0) );
            }
            iFunc = Abc_NamStrFindOrAdd( pMan.pFuncs, (char *)pObj.pData, nil );
            Abc_ObjFanout0(pObj).iTemp = Au_NtkCreateNode(p, vFanins, iFunc);
            continue;
        }
        assert(Abc_ObjIsBox(pObj), "Abc_ObjIsBox(pObj)");
        // inlining of macro Abc_ObjForEachFanin
        for  k = 0; (k < Abc_ObjFaninNum(pObj)) && (((pTerm) = Abc_ObjFanin(pObj, k)), 1); k++  {
            Vec_IntPush( vFanins, Au_Var2Lit(Abc_ObjFanin0(pTerm).iTemp, 0) );
        }
        pObj.iTemp = Au_NtkCreateBox(p, vFanins, Abc_ObjFanoutNum(pObj), ((Abc_Ntk_t *)pObj.pData).iStep );
        pAuObj = Au_NtkObj(p, pObj.iTemp);
        // inlining of macro Abc_ObjForEachFanout
        for  k = 0; (k < Abc_ObjFanoutNum(pObj)) && (((pTerm) = Abc_ObjFanout(pObj, k)), 1); k++  {
            Abc_ObjFanout0(pTerm).iTemp = Au_ObjFanout(pAuObj, k);
        }
    }
//    Vec_PtrFree( vOrder );
    Vec_IntFree( vFanins );
    // copy POs
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
        Au_NtkCreatePo( p, Au_Var2Lit(Abc_ObjFanin0(pTerm).iTemp, 0) );
    }
//    Au_NtkPrintStats( p );
    return p;
}

func Au_ManDeriveTest(pRoot *Abc_Ntk_t) *Gia_Man_t {
    extern Vec_Ptr_t * Abc_NtkCollectHie( Abc_Ntk_t * pNtk );

//    char * pModelName = NULL;
    var pModelName string  = "path_0_r_x_lhs";
    var pGia *Gia_Man_t  = nil;
    var vOrder, vModels *Vec_Ptr_t;
    var pMod *Abc_Ntk_t;
    var pMan *Au_Man_t;
    var pNtk *Au_Ntk_t  = nil;
    var clk1 abctime;
    var clk2 abctime;
    var clk3 abctime;
    var clk abctime = Abc_Clock();
    var i int32;

    clk1 = Abc_Clock();
    pMan = Au_ManAlloc( pRoot.pDesign ? pRoot.pDesign.pName : pRoot.pName );
    pMan.pFuncs = Abc_NamStart( 100, 16 );
    clk2 += Abc_Clock() - clk1;

    vModels = Abc_NtkCollectHie( pRoot );
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(vModels)) && (((pMod) = (Abc_Ntk_t *)Vec_PtrEntry(vModels, i)), 1); i++  {
        vOrder = Abc_NtkDfsBoxes( pMod );

        clk1 = Abc_Clock();
        pNtk = Au_NtkDerive( pMan, pMod, vOrder );
        pMod.iStep = pNtk.Id;
        pMod.pData = pNtk;
        clk2 += Abc_Clock() - clk1;

        Vec_PtrFree( vOrder );
    }
    Vec_PtrFree( vModels );
    // order models in topological order
    Au_ManReorderModels( pMan, pNtk );

    // print statistics
    Au_ManPrintStats( pMan );
    Au_ManCountThings( pNtk.pMan );

    // select network
    if pModelName {
        pNtk = Au_ManFindNtkP( pMan, pModelName );
        if pNtk == nil {
            fmt.Printf("Could not find module \"%s\".\n", pModelName); }
    }
    if pNtk == nil {
        pNtk = (Au_Ntk_t *)pRoot.pData; }

  
//    if ( !Abc_NtkCheckRecursive(pRoot) )
    {
        clk1 = Abc_Clock();
        pGia = Au_NtkDeriveFlatGia( pNtk );
        clk3 = Abc_Clock() - clk1;
//        printf( "GIA objects max = %d.\n", pMan->nGiaObjMax );
    }

//    clk1 = Abc_Clock();
//    Au_NtkSuppSizeTest( (Au_Ntk_t *)pRoot->pData );
//    clk4 = Abc_Clock() - clk1;

    clk1 = Abc_Clock();
    Au_ManDelete( pMan );
    clk2 += Abc_Clock() - clk1;
    
    Abc_PrintTime( 1, "Time all ", Abc_Clock() - clk );
    Abc_PrintTime( 1, "Time new ", clk2 );
    Abc_PrintTime( 1, "Time GIA ", clk3 );
//    Abc_PrintTime( 1, "Time supp", clk4 );
    return pGia;
}

/**Function*************************************************************

  Synopsis    [Performs hierarchical equivalence checking.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NtkHieCecTest2(pFileName string, pModelName string, fVerbose int32) *Gia_Man_t {
    var fSimulation int32;
    var pGia *Gia_Man_t  = nil;
    var pNtk *Au_Ntk_t;
    var pNtkClp *Au_Ntk_t = nil;
    var clk1 abctime;
    var clk abctime = Abc_Clock();

    // read hierarchical netlist
    pNtk = Au_NtkParseCBlif( pFileName );
    if pNtk == nil {
        fmt.Printf("Reading CBLIF file has failed.\n");
        return nil;
    }
    if pNtk.pMan == nil || pNtk.pMan.vNtks.pArray == nil {
        fmt.Printf("There is no hierarchy information.\n");
        Au_NtkFree( pNtk );
        return nil;
    }
    Abc_PrintTime( 1, "Reading file", Abc_Clock() - clk );

    if fVerbose != 0 {
        Au_ManPrintBoxInfo( pNtk );
//    Au_ManPrintBoxInfoSorted( pNtk );
        Au_ManPrintStats( pNtk.pMan );
    }
    Au_ManCountThings( pNtk.pMan );

    // select network
    if pModelName {
        pNtkClp = Au_ManFindNtkP( pNtk.pMan, pModelName ); }
    if pNtkClp == nil {
        pNtkClp = pNtk; }

    // check if the model is recursive
    Au_NtkCheckRecursive( pNtkClp );

    // collapse
    clk1 = Abc_Clock();
    if fSimulation != 0 {
        Au_NtkTerSimulate( pNtkClp );
        Abc_PrintTime( 1, "Time sim ", Abc_Clock() - clk1 );
    
    } else {
        pGia = Au_NtkDeriveFlatGia( pNtkClp );
        Abc_PrintTime( 1, "Time GIA ", Abc_Clock() - clk1 );
    }

    // delete
    Au_ManDelete( pNtk.pMan );
    Abc_PrintTime( 1, "Time all ", Abc_Clock() - clk );
    return pGia;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

