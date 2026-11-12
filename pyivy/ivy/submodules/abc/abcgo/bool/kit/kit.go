package kit

/**CFile****************************************************************

  FileName    [kit.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Computation kit.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - Dec 6, 2006.]

  Revision    [$Id: kit.h,v 1.00 2006/12/06 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "cloud.h"

#ifdef ABC_USE_CUDD
#include "bdd/extrab/extraBdd.h"
#endif

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

   
type Kit_Sop_t struct {
    nLits               int32;          // the number of literals
    nCubes               int32;         // the number of cubes
    pCubes         *uint32;         // the storage for cubes
}

   
type Kit_Edge_t struct {
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */   // the complemented bit
    Node uint32; /* C bitfield: 30 bits; review packing and truncation */   // the decomposition node pointed by the edge
}

   
type Kit_Node_t struct {
    eEdge0        Kit_Edge_t;          // the left child of the node
    eEdge1        Kit_Edge_t;          // the right child of the node
    // other info
    union { int       iFunc;           // the function of the node (BDD or AIG)
    void *            pFunc; };        // the function of the node (BDD or AIG)
    Level uint32; /* C bitfield: 14 bits; review packing and truncation */   // the level of this node in the global AIG
    // printing info 
    fNodeOr uint32; /* C bitfield: 1 bits; review packing and truncation */   // marks the original OR node
    fCompl0 uint32; /* C bitfield: 1 bits; review packing and truncation */   // marks the original complemented edge
    fCompl1 uint32; /* C bitfield: 1 bits; review packing and truncation */   // marks the original complemented edge
    // latch info
    nLat0 uint32; /* C bitfield: 5 bits; review packing and truncation */   // the number of latches on the first edge
    nLat1 uint32; /* C bitfield: 5 bits; review packing and truncation */   // the number of latches on the second edge
    nLat2 uint32; /* C bitfield: 5 bits; review packing and truncation */   // the number of latches on the output edge
}

   
type Kit_Graph_t struct {
    fConst               int32;          // marks the constant 1 graph
    nLeaves               int32;         // the number of leaves
    nSize               int32;           // the number of nodes (including the leaves) 
    nCap               int32;            // the number of allocated nodes
    pNodes       *Kit_Node_t;          // the array of leaves and internal nodes
    eRoot        Kit_Edge_t;           // the pointer to the topmost node
}


// DSD node types
 type Kit_Dsd_t int32
const ( 
    KIT_DSD_NONE Kit_Dsd_t  = 0
  // 0: unknown
    KIT_DSD_CONST1 Kit_Dsd_t = KIT_DSD_NONE + 1
     // 1: constant 1
    KIT_DSD_VAR Kit_Dsd_t = KIT_DSD_CONST1 + 1
        // 2: elementary variable
    KIT_DSD_AND Kit_Dsd_t = KIT_DSD_VAR + 1
        // 3: multi-input AND
    KIT_DSD_XOR Kit_Dsd_t = KIT_DSD_AND + 1
        // 4: multi-input XOR
    KIT_DSD_PRIME Kit_Dsd_t = KIT_DSD_XOR + 1       // 5: arbitrary function of 3+ variables

)

 

// DSD node
   
type Kit_DsdObj_t struct { 
    Id uint32; /* C bitfield: 6 bits; review packing and truncation */  // the number of this node
    Type uint32; /* C bitfield: 3 bits; review packing and truncation */  // none, const, var, AND, XOR, MUX, PRIME
    fMark uint32; /* C bitfield: 1 bits; review packing and truncation */  // finished checking output
    Offset uint32; /* C bitfield: 8 bits; review packing and truncation */  // offset to the truth table
    nRefs uint32; /* C bitfield: 8 bits; review packing and truncation */  // offset to the truth table
    nFans uint32; /* C bitfield: 6 bits; review packing and truncation */  // the number of fanins of this node
    unsigned short pFans[0];        // the fanin literals
}

// DSD network
   
type Kit_DsdNtk_t struct {
    unsigned short nVars;           // at most 16 (perhaps 18?)
    unsigned short nNodesAlloc;     // the number of allocated nodes (at most nVars)
    unsigned short nNodes;          // the number of nodes
    unsigned short Root;            // the root of the tree
    pMem      *uint32;            // memory for the truth tables (memory manager?)
    pSupps      *uint32;          // supports of the nodes
    pNodes **Kit_DsdObj_t;          // the nodes
}

// DSD manager
   
type Kit_DsdMan_t struct {
    nVars            int32;           // the maximum number of variables
    nWords            int32;          // the number of words in TTs
    vTtElems     *Vec_Ptr_t;        // elementary truth tables
    vTtNodes     *Vec_Ptr_t;        // the node truth tables
    // BDD representation
    dd  *CloudManager;              // BDD package
    vTtBdds     *Vec_Ptr_t;         // the node truth tables
    vNodes     *Vec_Int_t;          // temporary array for BDD nodes
}
 
func Kit_DsdObjOffset(nFans int32) uint32 { return (nFans >> 1) + ((nFans & 1) > 0);                    }
func Kit_DsdObjTruth(pObj *Kit_DsdObj_t) *uint32 { if pObj.Type == KIT_DSD_PRIME {
    return (unsigned *)pObj.pFans + pObj.Offset;
} else {
    return nil;
} }
func Kit_DsdNtkObjNum(pNtk *Kit_DsdNtk_t) int32 { return pNtk.nVars + pNtk.nNodes; }
func Kit_DsdNtkObj(pNtk *Kit_DsdNtk_t, Id int32) *Kit_DsdObj_t { assert(Id >= 0 && Id < pNtk.nVars + pNtk.nNodes, "Id >= 0 && Id < pNtk.nVars + pNtk.nNodes"); if Id < pNtk.nVars {
    return nil;
} else {
    return pNtk.pNodes[Id - pNtk.nVars];
} }
func Kit_DsdNtkRoot(pNtk *Kit_DsdNtk_t) *Kit_DsdObj_t { return Kit_DsdNtkObj( pNtk, Abc_Lit2Var(pNtk.Root) );                      }
func Kit_DsdLitIsLeaf(pNtk *Kit_DsdNtk_t, Lit int32) int32 { var Id int32  = Abc_Lit2Var(Lit); assert(Id >= 0 && Id < pNtk.nVars + pNtk.nNodes, "Id >= 0 && Id < pNtk.nVars + pNtk.nNodes"); return Id < pNtk.nVars; }
func Kit_DsdLitSupport(pNtk *Kit_DsdNtk_t, Lit int32) uint32 { var Id int32  = Abc_Lit2Var(Lit); assert(Id >= 0 && Id < pNtk.nVars + pNtk.nNodes, "Id >= 0 && Id < pNtk.nVars + pNtk.nNodes"); if pNtk.pSupps {
    if Id < pNtk.nVars {
        return (1 << Id);
    } else {
        return pNtk.pSupps[Id - pNtk.nVars];
    }
} else {
    return 0;
} }

// Original macro:
// #define Kit_DsdNtkForEachObj( pNtk, pObj, i )                                      \
//     for ( i = 0; (i < (pNtk)->nNodes) && ((pObj) = (pNtk)->pNodes[i]); i++ )

// Original macro:
// #define Kit_DsdObjForEachFanin( pNtk, pObj, iLit, i )                              \
//     for ( i = 0; (i < (int)(pObj)->nFans) && ((iLit) = (pObj)->pFans[i], 1); i++ )

// Original macro:
// #define Kit_DsdObjForEachFaninReverse( pNtk, pObj, iLit, i )                       \
//     for ( i = (int)(pObj)->nFans - 1; (i >= 0) && ((iLit) = (pObj)->pFans[i], 1); i-- )


// Original macro:
// #define Kit_PlaForEachCube( pSop, nFanins, pCube )                \
//     for ( pCube = (pSop); *pCube; pCube += (nFanins) + 3 )

// Original macro:
// #define Kit_PlaCubeForEachVar( pCube, Value, i )                     \
//     for ( i = 0; (pCube[i] != ' ') && (Value = pCube[i]); i++ )           


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define KIT_MIN(a,b)       (((a) < (b))? (a) : (b))

// Original macro:
// #define KIT_MAX(a,b)       (((a) > (b))? (a) : (b))

#define KIT_INFINITY       (100000000)

func Kit_CubeHasLit(uCube uint32, i int32) int32 { return(uCube &  uint32((1<<i))) > 0;  }
func Kit_CubeSetLit(uCube uint32, i int32) uint32 { return uCube |  uint32((1<<i));       }
func Kit_CubeXorLit(uCube uint32, i int32) uint32 { return uCube ^  uint32((1<<i));       }
func Kit_CubeRemLit(uCube uint32, i int32) uint32 { return uCube & ~uint32((1<<i));       }

func Kit_CubeContains(uLarge uint32, uSmall uint32) int32 { return (uLarge & uSmall) == uSmall;     }
func Kit_CubeSharp(uCube uint32, uMask uint32) uint32 { return uCube & ~uMask;                  }
func Kit_CubeMask(nVar int32) uint32 { return (~uint32(0)) >> (32-nVar);     }

func Kit_CubeIsMarked(uCube uint32) int32 { return Kit_CubeHasLit( uCube, 31 );     }
func Kit_CubeMark(uCube uint32) uint32 { return Kit_CubeSetLit( uCube, 31 );     }
func Kit_CubeUnmark(uCube uint32) uint32 { return Kit_CubeRemLit( uCube, 31 );     }

func Kit_SopCubeNum(cSop *Kit_Sop_t) int32 { return cSop.nCubes;                    }
func Kit_SopCube(cSop *Kit_Sop_t, i int32) uint32 { return cSop.pCubes[i];                 }
func Kit_SopShrink(cSop *Kit_Sop_t, nCubesNew int32) { cSop.nCubes = nCubesNew;               }
func Kit_SopPushCube(cSop *Kit_Sop_t, uCube uint32) { cSop.pCubes[cSop.nCubes++] = uCube;   }
func Kit_SopWriteCube(cSop *Kit_Sop_t, uCube uint32, i int32) { cSop.pCubes[i] = uCube;                }

func Kit_EdgeCreate(Node int32, fCompl int32) Kit_Edge_t { var eEdge Kit_Edge_t  = Kit_Edge_t{ uint32(fCompl), uint32(Node) }; return eEdge;  }
func Kit_EdgeToInt(eEdge Kit_Edge_t) uint32 { return (eEdge.Node << 1) | eEdge.fCompl;            }
func Kit_IntToEdge(Edge uint32) Kit_Edge_t { return Kit_EdgeCreate( Edge >> 1, Edge & 1 );       }
//static inline unsigned     Kit_EdgeToInt_( Kit_Edge_t eEdge )                             { return *(unsigned *)&eEdge;                         }
//static inline Kit_Edge_t   Kit_IntToEdge_( unsigned Edge )                                { return *(Kit_Edge_t *)&Edge;                        }
func Kit_EdgeToInt_(m Kit_Edge_t) uint32 { union { var x Kit_Edge_t; var y uint32; }
 v; v.x = m; return v.y;  }
func Kit_IntToEdge_(m uint32) Kit_Edge_t { union { var x Kit_Edge_t; var y uint32; }
 v; v.y = m; return v.x;  }

func Kit_GraphIsConst(pGraph *Kit_Graph_t) int32 { return pGraph.fConst;                              }
func Kit_GraphIsConst0(pGraph *Kit_Graph_t) int32 { return pGraph.fConst && pGraph.eRoot.fCompl;      }
func Kit_GraphIsConst1(pGraph *Kit_Graph_t) int32 { return pGraph.fConst && !pGraph.eRoot.fCompl;     }
func Kit_GraphIsComplement(pGraph *Kit_Graph_t) int32 { return pGraph.eRoot.fCompl;                        }
func Kit_GraphIsVar(pGraph *Kit_Graph_t) int32 { return pGraph.eRoot.Node < uint32(pGraph.nLeaves); }
func Kit_GraphComplement(pGraph *Kit_Graph_t) { pGraph.eRoot.fCompl ^= 1;                          }
func Kit_GraphSetRoot(pGraph *Kit_Graph_t, eRoot Kit_Edge_t) { pGraph.eRoot = eRoot;                              }
func Kit_GraphLeaveNum(pGraph *Kit_Graph_t) int32 { return pGraph.nLeaves;                             }
func Kit_GraphNodeNum(pGraph *Kit_Graph_t) int32 { return pGraph.nSize - pGraph.nLeaves;             }
func Kit_GraphNode(pGraph *Kit_Graph_t, i int32) *Kit_Node_t { return pGraph.pNodes + i;                          }
func Kit_GraphNodeLast(pGraph *Kit_Graph_t) *Kit_Node_t { return pGraph.pNodes + pGraph.nSize - 1;          }
func Kit_GraphNodeInt(pGraph *Kit_Graph_t, pNode *Kit_Node_t) int32 { return pNode - pGraph.pNodes;                      }
func Kit_GraphNodeIsVar(pGraph *Kit_Graph_t, pNode *Kit_Node_t) int32 { return Kit_GraphNodeInt(pGraph,pNode) < pGraph.nLeaves; }
func Kit_GraphVar(pGraph *Kit_Graph_t) *Kit_Node_t { assert((Kit_GraphIsVar( pGraph ) != 0), "Kit_GraphIsVar( pGraph )");    return Kit_GraphNode( pGraph, pGraph.eRoot.Node );      }
func Kit_GraphVarInt(pGraph *Kit_Graph_t) int32 { assert((Kit_GraphIsVar( pGraph ) != 0), "Kit_GraphIsVar( pGraph )");    return Kit_GraphNodeInt( pGraph, Kit_GraphVar(pGraph) ); }
func Kit_GraphNodeFanin0(pGraph *Kit_Graph_t, pNode *Kit_Node_t) *Kit_Node_t { if Kit_GraphNodeIsVar(pGraph, pNode) != 0 {
    return nil;
} else {
    return Kit_GraphNode(pGraph, pNode.eEdge0.Node);
}     }
func Kit_GraphNodeFanin1(pGraph *Kit_Graph_t, pNode *Kit_Node_t) *Kit_Node_t { if Kit_GraphNodeIsVar(pGraph, pNode) != 0 {
    return nil;
} else {
    return Kit_GraphNode(pGraph, pNode.eEdge1.Node);
}     }
func Kit_GraphRootLevel(pGraph *Kit_Graph_t) int32 { return Kit_GraphNode(pGraph, pGraph.eRoot.Node).Level;                                        }

func Kit_SuppIsMinBase(Supp int32) int32 { return (Supp & (Supp+1)) == 0;                      }

func Kit_BitWordNum(nBits int32) int32 { return nBits/(8*sizeof(unsigned)) + ((nBits%(8*sizeof(unsigned))) > 0); }
func Kit_TruthWordNum(nVars int32) int32 { if nVars <= 5 {
    return 1;
} else {
    return (1 << (nVars - 5));
}                             }
func Kit_BitMask(nBits int32) uint32 { assert(nBits <= 32, "nBits <= 32"); return ~((~uint32(0)) << nBits);               }

func Kit_TruthSetBit(p *uint32, Bit int32) { p[Bit>>5] |= (1<<(Bit & 31));               }
func Kit_TruthXorBit(p *uint32, Bit int32) { p[Bit>>5] ^= (1<<(Bit & 31));               }
func Kit_TruthHasBit(p *uint32, Bit int32) int32 { return (p[Bit>>5] & (1<<(Bit & 31))) > 0;   }

func Kit_WordFindFirstBit(uWord uint32) int32 {
    var i int32;
    for  i = 0; i < 32; i++  {
        if uWord & (1 << i) {
            return i; }
    }
    return -1;
}
func Kit_WordHasOneBit(uWord uint32) int32 {
    return (uWord & (uWord - 1)) == 0;
}
func Kit_WordCountOnes(uWord uint32) int32 {
    uWord = (uWord & 0x55555555) + ((uWord>>1) & 0x55555555);
    uWord = (uWord & 0x33333333) + ((uWord>>2) & 0x33333333);
    uWord = (uWord & 0x0F0F0F0F) + ((uWord>>4) & 0x0F0F0F0F);
    uWord = (uWord & 0x00FF00FF) + ((uWord>>8) & 0x00FF00FF);
    return  (uWord & 0x0000FFFF) + (uWord>>16);
}
func Kit_TruthCountOnes(pIn *uint32, nVars int32) int32 {
    var w, Counter int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        Counter += Kit_WordCountOnes(pIn[w]);
    }
    return Counter;
}
func Kit_TruthFindFirstBit(pIn *uint32, nVars int32) int32 {
    var w int32;
    for  w = 0; w < Kit_TruthWordNum(nVars); w++  {
        if pIn[w] {
            return 32*w + Kit_WordFindFirstBit(pIn[w]); }
    }
    return -1;
}
func Kit_TruthFindFirstZero(pIn *uint32, nVars int32) int32 {
    var w int32;
    for  w = 0; w < Kit_TruthWordNum(nVars); w++  {
        if ~pIn[w] {
            return 32*w + Kit_WordFindFirstBit(~pIn[w]); }
    }
    return -1;
}
func Kit_TruthIsEqual(pIn0 *uint32, pIn1 *uint32, nVars int32) int32 {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn0[w] != pIn1[w] {
            return 0; }
    }
    return 1;
}
func Kit_TruthIsEqualWithCare(pIn0 *uint32, pIn1 *uint32, pCare *uint32, nVars int32) int32 {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        if (pIn0[w] & pCare[w]) != (pIn1[w] & pCare[w]) {
            return 0; }
    }
    return 1;
}
func Kit_TruthIsOpposite(pIn0 *uint32, pIn1 *uint32, nVars int32) int32 {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn0[w] != ~pIn1[w] {
            return 0; }
    }
    return 1;
}
func Kit_TruthIsEqualWithPhase(pIn0 *uint32, pIn1 *uint32, nVars int32) int32 {
    var w int32;
    if (pIn0[0] & 1) == (pIn1[0] & 1) {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            if pIn0[w] != pIn1[w] {
                return 0; }
        }
    
    } else {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            if pIn0[w] != ~pIn1[w] {
                return 0; }
        }
    }
    return 1;
}
func Kit_TruthIsConst0(pIn *uint32, nVars int32) int32 {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn[w] {
            return 0; }
    }
    return 1;
}
func Kit_TruthIsConst1(pIn *uint32, nVars int32) int32 {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn[w] != ~uint32(0) {
            return 0; }
    }
    return 1;
}
func Kit_TruthIsImply(pIn1 *uint32, pIn2 *uint32, nVars int32) int32 {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn1[w] & ~pIn2[w] {
            return 0; }
    }
    return 1;
}
func Kit_TruthIsDisjoint(pIn1 *uint32, pIn2 *uint32, nVars int32) int32 {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn1[w] & pIn2[w] {
            return 0; }
    }
    return 1;
}
func Kit_TruthIsDisjoint3(pIn1 *uint32, pIn2 *uint32, pIn3 *uint32, nVars int32) int32 {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        if pIn1[w] & pIn2[w] & pIn3[w] {
            return 0; }
    }
    return 1;
}
func Kit_TruthCopy(pOut *uint32, pIn *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn[w];
    }
}
func Kit_TruthClear(pOut *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = 0;
    }
}
func Kit_TruthFill(pOut *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = ~uint32(0);
    }
}
func Kit_TruthNot(pOut *uint32, pIn *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = ~pIn[w];
    }
}
func Kit_TruthAnd(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn0[w] & pIn1[w];
    }
}
func Kit_TruthOr(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn0[w] | pIn1[w];
    }
}
func Kit_TruthXor(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn0[w] ^ pIn1[w];
    }
}
func Kit_TruthSharp(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = pIn0[w] & ~pIn1[w];
    }
}
func Kit_TruthNand(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = ~(pIn0[w] & pIn1[w]);
    }
}
func Kit_TruthAndPhase(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32, fCompl0 int32, fCompl1 int32) {
    var w int32;
    if fCompl0 && fCompl1 {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = ~(pIn0[w] | pIn1[w]);
        }
    
    } else if fCompl0 && !fCompl1 {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = ~pIn0[w] & pIn1[w];
        }
    
    } else if !fCompl0 && fCompl1 {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = pIn0[w] & ~pIn1[w];
        }
    
    } else { // if ( !fCompl0 && !fCompl1 )
    
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = pIn0[w] & pIn1[w];
        }
    }
}
func Kit_TruthOrPhase(pOut *uint32, pIn0 *uint32, pIn1 *uint32, nVars int32, fCompl0 int32, fCompl1 int32) {
    var w int32;
    if fCompl0 && fCompl1 {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = ~(pIn0[w] & pIn1[w]);
        }
    
    } else if fCompl0 && !fCompl1 {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = ~pIn0[w] | pIn1[w];
        }
    
    } else if !fCompl0 && fCompl1 {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = pIn0[w] | ~pIn1[w];
        }
    
    } else { // if ( !fCompl0 && !fCompl1 )
    
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = pIn0[w] | pIn1[w];
        }
    }
}
func Kit_TruthMux(pOut *uint32, pIn0 *uint32, pIn1 *uint32, pCtrl *uint32, nVars int32) {
    var w int32;
    for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
        pOut[w] = (pIn0[w] & ~pCtrl[w]) | (pIn1[w] & pCtrl[w]);
    }
}
func Kit_TruthMuxPhase(pOut *uint32, pIn0 *uint32, pIn1 *uint32, pCtrl *uint32, nVars int32, fComp0 int32) {
    var w int32;
    if fComp0 != 0 {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = (~pIn0[w] & ~pCtrl[w]) | (pIn1[w] & pCtrl[w]);
        }
    } else {
        for  w = Kit_TruthWordNum(nVars)-1; w >= 0; w--  {
            pOut[w] = (pIn0[w] & ~pCtrl[w]) | (pIn1[w] & pCtrl[w]);
        } }
}
func Kit_TruthIthVar(pTruth *uint32, nVars int32, iVar int32) {
    var Masks [5]uint32 = [5]uint32{ 0xAAAAAAAA, 0xCCCCCCCC, 0xF0F0F0F0, 0xFF00FF00, 0xFFFF0000 };
    var k int32;
    var nWords int32;
    if nVars <= 5 {
        nWords = 1;
    } else {
        nWords = (1 << (nVars - 5));
    }
    if iVar < 5 {
        for  k = 0; k < nWords; k++  {
            pTruth[k] = Masks[iVar];
        }
    
    } else {
        for  k = 0; k < nWords; k++  {
            if k & (1 << (iVar-5)) {
                pTruth[k] = ~uint32(0);
            } else {
                pTruth[k] = 0; }
        }
    }
}


////////////////////////////////////////////////////////////////////////
///                           ITERATORS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Kit_SopForEachCube( cSop, uCube, i )                                      \
//     for ( i = 0; (i < Kit_SopCubeNum(cSop)) && ((uCube) = Kit_SopCube(cSop, i)); i++ )

// Original macro:
// #define Kit_CubeForEachLiteral( uCube, Lit, nLits, i )                            \
//     for ( i = 0; (i < (nLits)) && ((Lit) = Kit_CubeHasLit(uCube, i)); i++ )


// Original macro:
// #define Kit_GraphForEachLeaf( pGraph, pLeaf, i )                                              \
//     for ( i = 0; (i < (pGraph)->nLeaves) && (((pLeaf) = Kit_GraphNode(pGraph, i)), 1); i++ )

// Original macro:
// #define Kit_GraphForEachNode( pGraph, pAnd, i )                                               \
//     for ( i = (pGraph)->nLeaves; (i < (pGraph)->nSize) && (((pAnd) = Kit_GraphNode(pGraph, i)), 1); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== kitBdd.c ==========================================================*/
#ifdef ABC_USE_CUDD
extern DdNode *        Kit_SopToBdd( DdManager * dd, Kit_Sop_t * cSop, int nVars );
extern DdNode *        Kit_GraphToBdd( DdManager * dd, Kit_Graph_t * pGraph );
extern DdNode *        Kit_TruthToBdd( DdManager * dd, unsigned * pTruth, int nVars, int fMSBonTop );
#endif
/*=== kitCloud.c ==========================================================*/
// extern CloudNode *     Kit_TruthToCloud( CloudManager * dd, unsigned * pTruth, int nVars );

// extern unsigned *      Kit_CloudToTruth( Vec_Int_t * vNodes, int nVars, Vec_Ptr_t * vStore, int fInv );

// extern int             Kit_CreateCloud( CloudManager * dd, CloudNode * pFunc, Vec_Int_t * vNodes );

// extern int             Kit_CreateCloudFromTruth( CloudManager * dd, unsigned * pTruth, int nVars, Vec_Int_t * vNodes );

// extern unsigned *      Kit_TruthCompose( CloudManager * dd, unsigned * pTruth, int nVars, unsigned ** pInputs, int nVarsAll, Vec_Ptr_t * vStore, Vec_Int_t * vNodes );

// extern void            Kit_TruthCofSupports( Vec_Int_t * vBddDir, Vec_Int_t * vBddInv, int nVars, Vec_Int_t * vMemory, unsigned * puSupps );

/*=== kitDsd.c ==========================================================*/
// extern Kit_DsdMan_t *  Kit_DsdManAlloc( int nVars, int nNodes );

// extern void            Kit_DsdManFree( Kit_DsdMan_t * p );

// extern Kit_DsdNtk_t *  Kit_DsdDeriveNtk( unsigned * pTruth, int nVars, int nLutSize );

// extern unsigned *      Kit_DsdTruthCompute( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk );

// extern void            Kit_DsdTruth( Kit_DsdNtk_t * pNtk, unsigned * pTruthRes );

// extern void            Kit_DsdTruthPartial( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, unsigned * pTruthRes, unsigned uSupp );

// extern void            Kit_DsdTruthPartialTwo( Kit_DsdMan_t * p, Kit_DsdNtk_t * pNtk, unsigned uSupp, int iVar, unsigned * pTruthCo, unsigned * pTruthDec );

// extern void            Kit_DsdPrint( FILE * pFile, Kit_DsdNtk_t * pNtk );

// extern void            Kit_DsdPrintExpanded( Kit_DsdNtk_t * pNtk );

// extern void            Kit_DsdPrintFromTruth( unsigned * pTruth, int nVars );

// extern void            Kit_DsdPrintFromTruth2( FILE * pFile, unsigned * pTruth, int nVars );

// extern void            Kit_DsdWriteFromTruth( char * pBuffer, unsigned * pTruth, int nVars );

// extern Kit_DsdNtk_t *  Kit_DsdDecompose( unsigned * pTruth, int nVars );

// extern Kit_DsdNtk_t *  Kit_DsdDecomposeExpand( unsigned * pTruth, int nVars );

// extern Kit_DsdNtk_t *  Kit_DsdDecomposeMux( unsigned * pTruth, int nVars, int nDecMux );

// extern void            Kit_DsdVerify( Kit_DsdNtk_t * pNtk, unsigned * pTruth, int nVars );

// extern void            Kit_DsdNtkFree( Kit_DsdNtk_t * pNtk );

// extern int             Kit_DsdNonDsdSizeMax( Kit_DsdNtk_t * pNtk );

// extern Kit_DsdObj_t *  Kit_DsdNonDsdPrimeMax( Kit_DsdNtk_t * pNtk );

// extern unsigned        Kit_DsdNonDsdSupports( Kit_DsdNtk_t * pNtk );

// extern int             Kit_DsdCountAigNodes( Kit_DsdNtk_t * pNtk );

// extern unsigned        Kit_DsdGetSupports( Kit_DsdNtk_t * p );

// extern Kit_DsdNtk_t *  Kit_DsdExpand( Kit_DsdNtk_t * p );

// extern Kit_DsdNtk_t *  Kit_DsdShrink( Kit_DsdNtk_t * p, int pPrios[] );

// extern void            Kit_DsdRotate( Kit_DsdNtk_t * p, int pFreqs[] );

// extern int             Kit_DsdCofactoring( unsigned * pTruth, int nVars, int * pCofVars, int nLimit, int fVerbose );

/*=== kitFactor.c ==========================================================*/
// extern Kit_Graph_t *   Kit_SopFactor( Vec_Int_t * vCover, int fCompl, int nVars, Vec_Int_t * vMemory );

/*=== kitGraph.c ==========================================================*/
// extern Kit_Graph_t *   Kit_GraphCreate( int nLeaves );
   
// extern Kit_Graph_t *   Kit_GraphCreateConst0();
   
// extern Kit_Graph_t *   Kit_GraphCreateConst1();
   
// extern Kit_Graph_t *   Kit_GraphCreateLeaf( int iLeaf, int nLeaves, int fCompl );
   
// extern void            Kit_GraphFree( Kit_Graph_t * pGraph );
   
// extern Kit_Node_t *    Kit_GraphAppendNode( Kit_Graph_t * pGraph );
   
// extern Kit_Edge_t      Kit_GraphAddNodeAnd( Kit_Graph_t * pGraph, Kit_Edge_t eEdge0, Kit_Edge_t eEdge1 );

// extern Kit_Edge_t      Kit_GraphAddNodeOr( Kit_Graph_t * pGraph, Kit_Edge_t eEdge0, Kit_Edge_t eEdge1 );

// extern Kit_Edge_t      Kit_GraphAddNodeXor( Kit_Graph_t * pGraph, Kit_Edge_t eEdge0, Kit_Edge_t eEdge1, int Type );

// extern Kit_Edge_t      Kit_GraphAddNodeMux( Kit_Graph_t * pGraph, Kit_Edge_t eEdgeC, Kit_Edge_t eEdgeT, Kit_Edge_t eEdgeE, int Type );

// extern unsigned        Kit_GraphToTruth( Kit_Graph_t * pGraph );

// extern Kit_Graph_t *   Kit_TruthToGraph( unsigned * pTruth, int nVars, Vec_Int_t * vMemory );

// extern Kit_Graph_t *   Kit_TruthToGraph2( unsigned * pTruth0, unsigned * pTruth1, int nVars, Vec_Int_t * vMemory );

// extern int             Kit_GraphLeafDepth_rec( Kit_Graph_t * pGraph, Kit_Node_t * pNode, Kit_Node_t * pLeaf );

// extern int             Kit_TruthLitNum( unsigned * pTruth, int nVars, Vec_Int_t * vMemory );

/*=== kitHop.c ==========================================================*/
//extern int             Kit_TruthToGia( Gia_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory, Vec_Int_t * vLeaves, int fHash );
//extern Hop_Obj_t *     Kit_GraphToHop( Hop_Man_t * pMan, Kit_Graph_t * pGraph );
//extern Hop_Obj_t *     Kit_TruthToHop( Hop_Man_t * pMan, unsigned * pTruth, int nVars, Vec_Int_t * vMemory );
//extern Hop_Obj_t *     Kit_CoverToHop( Hop_Man_t * pMan, Vec_Int_t * vCover, int nVars, Vec_Int_t * vMemory );
/*=== kitIsop.c ==========================================================*/
// extern int             Kit_TruthIsop( unsigned * puTruth, int nVars, Vec_Int_t * vMemory, int fTryBoth );

// extern int             Kit_TruthIsop2( unsigned * puTruth0, unsigned * puTruth1, int nVars, Vec_Int_t * vMemory, int fTryBoth, int fReturnTt );

// extern void            Kit_TruthIsopPrint( unsigned * puTruth, int nVars, Vec_Int_t * vMemory, int fTryBoth );

// extern void            Kit_TruthIsopPrintCover( Vec_Int_t * vCover, int nVars, int fCompl );

/*=== kitPla.c ==========================================================*/
// extern int             Kit_PlaIsConst0( char * pSop );

// extern int             Kit_PlaIsConst1( char * pSop );

// extern int             Kit_PlaIsBuf( char * pSop );

// extern int             Kit_PlaIsInv( char * pSop );

// extern int             Kit_PlaGetVarNum( char * pSop );

// extern int             Kit_PlaGetCubeNum( char * pSop );

// extern int             Kit_PlaIsComplement( char * pSop );

// extern void            Kit_PlaComplement( char * pSop );

// extern char *          Kit_PlaStart( void * p, int nCubes, int nVars );

// extern char *          Kit_PlaCreateFromIsop( void * p, int nVars, Vec_Int_t * vCover );

// extern void            Kit_PlaToIsop( char * pSop, Vec_Int_t * vCover );

// extern char *          Kit_PlaStoreSop( void * p, char * pSop );

// extern char *          Kit_PlaFromTruth( void * p, unsigned * pTruth, int nVars, Vec_Int_t * vCover );

// extern char *          Kit_PlaFromTruthNew( unsigned * pTruth, int nVars, Vec_Int_t * vCover, Vec_Str_t * vStr );

// extern ABC_UINT64_T    Kit_PlaToTruth6( char * pSop, int nVars );

// extern void            Kit_PlaToTruth( char * pSop, int nVars, Vec_Ptr_t * vVars, unsigned * pTemp, unsigned * pTruth );

/*=== kitSop.c ==========================================================*/
// extern void            Kit_SopCreate( Kit_Sop_t * cResult, Vec_Int_t * vInput, int nVars, Vec_Int_t * vMemory );

// extern void            Kit_SopCreateInverse( Kit_Sop_t * cResult, Vec_Int_t * vInput, int nVars, Vec_Int_t * vMemory );

// extern void            Kit_SopDup( Kit_Sop_t * cResult, Kit_Sop_t * cSop, Vec_Int_t * vMemory );

// extern void            Kit_SopDivideByLiteralQuo( Kit_Sop_t * cSop, int iLit );

// extern void            Kit_SopDivideByCube( Kit_Sop_t * cSop, Kit_Sop_t * cDiv, Kit_Sop_t * vQuo, Kit_Sop_t * vRem, Vec_Int_t * vMemory );

// extern void            Kit_SopDivideInternal( Kit_Sop_t * cSop, Kit_Sop_t * cDiv, Kit_Sop_t * vQuo, Kit_Sop_t * vRem, Vec_Int_t * vMemory );

// extern void            Kit_SopMakeCubeFree( Kit_Sop_t * cSop );

// extern int             Kit_SopIsCubeFree( Kit_Sop_t * cSop );

// extern void            Kit_SopCommonCubeCover( Kit_Sop_t * cResult, Kit_Sop_t * cSop, Vec_Int_t * vMemory );

// extern int             Kit_SopAnyLiteral( Kit_Sop_t * cSop, int nLits );

// extern int             Kit_SopDivisor( Kit_Sop_t * cResult, Kit_Sop_t * cSop, int nLits, Vec_Int_t * vMemory );

// extern void            Kit_SopBestLiteralCover( Kit_Sop_t * cResult, Kit_Sop_t * cSop, unsigned uCube, int nLits, Vec_Int_t * vMemory );

/*=== kitTruth.c ==========================================================*/
// extern void            Kit_TruthSwapAdjacentVars( unsigned * pOut, unsigned * pIn, int nVars, int Start );

// extern void            Kit_TruthStretch( unsigned * pOut, unsigned * pIn, int nVars, int nVarsAll, unsigned Phase, int fReturnIn );

// extern void            Kit_TruthPermute( unsigned * pOut, unsigned * pIn, int nVars, char * pPerm, int fReturnIn );

// extern void            Kit_TruthShrink( unsigned * pOut, unsigned * pIn, int nVars, int nVarsAll, unsigned Phase, int fReturnIn );

// extern int             Kit_TruthVarInSupport( unsigned * pTruth, int nVars, int iVar );

// extern int             Kit_TruthSupportSize( unsigned * pTruth, int nVars );

// extern unsigned        Kit_TruthSupport( unsigned * pTruth, int nVars );

// extern void            Kit_TruthCofactor0( unsigned * pTruth, int nVars, int iVar );

// extern void            Kit_TruthCofactor1( unsigned * pTruth, int nVars, int iVar );

// extern void            Kit_TruthCofactor0New( unsigned * pOut, unsigned * pIn, int nVars, int iVar );

// extern void            Kit_TruthCofactor1New( unsigned * pOut, unsigned * pIn, int nVars, int iVar );

// extern int             Kit_TruthVarIsVacuous( unsigned * pOnset, unsigned * pOffset, int nVars, int iVar );

// extern void            Kit_TruthExist( unsigned * pTruth, int nVars, int iVar );

// extern void            Kit_TruthExistNew( unsigned * pRes, unsigned * pTruth, int nVars, int iVar );

// extern void            Kit_TruthExistSet( unsigned * pRes, unsigned * pTruth, int nVars, unsigned uMask );

// extern void            Kit_TruthForall( unsigned * pTruth, int nVars, int iVar );

// extern void            Kit_TruthForallNew( unsigned * pRes, unsigned * pTruth, int nVars, int iVar );

// extern void            Kit_TruthForallSet( unsigned * pRes, unsigned * pTruth, int nVars, unsigned uMask );

// extern void            Kit_TruthUniqueNew( unsigned * pRes, unsigned * pTruth, int nVars, int iVar );

// extern void            Kit_TruthMuxVar( unsigned * pOut, unsigned * pCof0, unsigned * pCof1, int nVars, int iVar );

// extern void            Kit_TruthMuxVarPhase( unsigned * pOut, unsigned * pCof0, unsigned * pCof1, int nVars, int iVar, int fCompl0 );

// extern void            Kit_TruthChangePhase( unsigned * pTruth, int nVars, int iVar );

// extern int             Kit_TruthVarsSymm( unsigned * pTruth, int nVars, int iVar0, int iVar1, unsigned * pCof0, unsigned * pCof1 );

// extern int             Kit_TruthVarsAntiSymm( unsigned * pTruth, int nVars, int iVar0, int iVar1, unsigned * pCof0, unsigned * pCof1 );

// extern int             Kit_TruthMinCofSuppOverlap( unsigned * pTruth, int nVars, int * pVarMin );

// extern int             Kit_TruthBestCofVar( unsigned * pTruth, int nVars, unsigned * pCof0, unsigned * pCof1 );

// extern void            Kit_TruthCountOnesInCofs( unsigned * pTruth, int nVars, int * pStore );

// extern void            Kit_TruthCountOnesInCofs0( unsigned * pTruth, int nVars, int * pStore );

// extern void            Kit_TruthCountOnesInCofsSlow( unsigned * pTruth, int nVars, int * pStore, unsigned * pAux );

// extern unsigned        Kit_TruthHash( unsigned * pIn, int nWords );

// extern unsigned        Kit_TruthSemiCanonicize( unsigned * pInOut, unsigned * pAux, int nVars, char * pCanonPerm );

// extern char *          Kit_TruthDumpToFile( unsigned * pTruth, int nVars, int nFile );

// extern void            Kit_TruthPrintProfile( unsigned * pTruth, int nVars );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

