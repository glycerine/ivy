package dsd

/**CFile****************************************************************

  FileName    [dsdApi.c]

  PackageName [DSD: Disjoint-support decomposition package.]

  Synopsis    [Implementation of API functions.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 8.0. Started - September 22, 2003.]

  Revision    [$Id: dsdApi.c,v 1.0 2002/22/09 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "dsdInt.h"

//ABC_NAMESPACE_IMPL_START

 
////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [APIs of the DSD node.]

  Description [The node's type can be retrieved by calling
  Dsd_NodeReadType(). The type is one of the following: constant 1 node, 
  the buffer (or the elementary variable), OR gate, EXOR gate, or 
  PRIME function (a non-DSD-decomposable function with more than two 
  inputs). The return value of Dsd_NodeReadFunc() is the global function 
  of the DSD node. The return value of Dsd_NodeReadSupp() is the support 
  of the global function of the DSD node. The array of DSD nodes
  returned by Dsd_NodeReadDecs() is the array of decomposition nodes for 
  the formal inputs of the given node. The number of decomposition entries 
  returned by Dsd_NodeReadDecsNum() is the number of formal inputs. 
  The mark is explained below.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_NodeReadType(p *Dsd_Node_t) Dsd_Type_t { return p.Type;     } 
func Dsd_NodeReadFunc(p *Dsd_Node_t) *DdNode { return p.G;        } 
func Dsd_NodeReadSupp(p *Dsd_Node_t) *DdNode { return p.S;        } 
func Dsd_NodeReadDecs(p *Dsd_Node_t) **Dsd_Node_t { return p.pDecs;    } 
func Dsd_NodeReadDec(p *Dsd_Node_t, i int32) *Dsd_Node_t { return p.pDecs[i]; } 
func Dsd_NodeReadDecsNum(p *Dsd_Node_t) int32 { return p.nDecs;    } 
func Dsd_NodeReadMark(p *Dsd_Node_t) int32 { return p.Mark;     } 

/**Function*************************************************************

  Synopsis    [APIs of the DSD node.]

  Description [This API allows the user to set the integer mark in the
  given DSD node. The mark is guaranteed to persist as long as the
  calls to the decomposition are not performed. In any case, the mark 
  is useful to associate the node with some temporary information, such 
  as its number in the DFS ordered list of the DSD nodes or its number in 
  the BLIF file that it being written.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_NodeSetMark(p *Dsd_Node_t, Mark int32) { p.Mark = Mark;     } 

/**Function*************************************************************

  Synopsis    [APIs of the DSD manager.]

  Description [Allows the use to get hold of an individual leave of
  the DSD tree (Dsd_ManagerReadInput) or an individual root of the 
  decomposition tree (Dsd_ManagerReadRoot). The root may have the 
  complemented attribute.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dsd_ManagerReadRoot(pMan *Dsd_Manager_t, i int32) *Dsd_Node_t { return pMan.pRoots[i];  } 
func Dsd_ManagerReadInput(pMan *Dsd_Manager_t, i int32) *Dsd_Node_t { return pMan.pInputs[i]; } 
func Dsd_ManagerReadConst1(pMan *Dsd_Manager_t) *Dsd_Node_t { return pMan.pConst1;    } 
func Dsd_ManagerReadDd(pMan *Dsd_Manager_t) *DdManager { return pMan.dd;         } 

////////////////////////////////////////////////////////////////////////
///                           END OF FILE                            ///
////////////////////////////////////////////////////////////////////////
//ABC_NAMESPACE_IMPL_END

