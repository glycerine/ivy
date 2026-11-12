package rpo

/**CFile****************************************************************

  FileName    [rpo.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [RPO]

  Synopsis    [Performs read polarity once factorization.]

  Author      [Mayler G. A. Martins / Vinicius Callegaro]
  
  Affiliation [UFRGS]

  Date        [Ver. 1.0. Started - May 08, 2013.]

  Revision    [$Id: rpo.c,v 1.00 2013/05/08 00:00:00 mgamartins Exp $]

 ***********************************************************************/

//#include <stdio.h>

//#include "literal.h"
//#include "rpo.h"
//#include "bool/kit/kit.h"
//#include "misc/util/abc_global.h"
//#include "misc/vec/vec.h"


//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


   

type Rpo_Man_t struct {
    target  *uint32;
    nVars int32;

    literals  **Literal_t;
    nLits int32;
    nLitsMax int32;

    lci *Rpo_LCI_Edge_t;
    nLCIElems int32;

    thresholdMax int32;

}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Check if two literals are AND-grouped]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Rpo_CheckANDGroup(lit1 *Literal_t, lit2 *Literal_t, nVars int32) int32 {
    var notLit1Func *uint32  = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    var notLit2Func *uint32  = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    var and1 *uint32;
    var and2 *uint32;
    var isZero int32;

    Kit_TruthNot(notLit1Func, lit1.function, nVars);
    Kit_TruthNot(notLit2Func, lit2.function, nVars);
    and1 = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    and2 = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    Kit_TruthAnd(and1, lit1.transition, notLit2Func, nVars);
    isZero = Kit_TruthIsConst0(and1, nVars);
    if isZero != 0 {
        Kit_TruthAnd(and2, lit2.transition, notLit1Func, nVars);
        isZero = Kit_TruthIsConst0(and2, nVars);
    }
    ABC_FREE(notLit1Func);
    ABC_FREE(notLit2Func);
    ABC_FREE(and1);
    ABC_FREE(and2);
    return isZero;
}

/**Function*************************************************************

  Synopsis    [Check if two literals are AND-grouped]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Rpo_CheckORGroup(lit1 *Literal_t, lit2 *Literal_t, nVars int32) int32 {
    var and1 *uint32  = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    var and2 *uint32  = ABC_ALLOC(unsigned, Kit_TruthWordNum(nVars));
    var isZero int32;
    Kit_TruthAnd(and1, lit1.transition, lit2.function, nVars);
    isZero = Kit_TruthIsConst0(and1, nVars);
    if isZero != 0 {
        Kit_TruthAnd(and2, lit2.transition, lit1.function, nVars);
        isZero = Kit_TruthIsConst0(and2, nVars);
    }
    ABC_FREE(and1);
    ABC_FREE(and2);
    return isZero;
}

func Rpo_CreateEdge(op Operator_t, i int32, j int32, vertexDegree *int32) *Rpo_LCI_Edge_t {
    var edge *Rpo_LCI_Edge_t  = new(Rpo_LCI_Edge_t);
    edge.connectionType = op;
    edge.idx1 = i;
    edge.idx2 = j;
    edge.visited = 0;
    vertexDegree[i]++;
    vertexDegree[j]++;
    return edge;
}

func Rpo_computeMinEdgeCost(edges **Rpo_LCI_Edge_t, edgeCount int32, vertexDegree *int32) int32 {
    var minCostIndex int32  = -1;
    var minVertexIndex int32  = -1;
    unsigned int minCost = ~0;
    var edge *Rpo_LCI_Edge_t;
    unsigned int edgeCost;
    var minVertex int32;
    var i int32;
    for i = 0; i < edgeCount; ++i {
        edge = edges[i];
        if !edge.visited {
            edgeCost = vertexDegree[edge.idx1] + vertexDegree[edge.idx2];
            if edge.idx1 < edge.idx2 {
                minVertex = edge.idx1;
            } else {
                minVertex = edge.idx2;
            }
            if edgeCost < minCost {
                minCost = edgeCost;
                minCostIndex = i;
                minVertexIndex = minVertex;
            } else if (edgeCost == minCost) && minVertex < minVertexIndex {
                minCost = edgeCost;
                minCostIndex = i;
                minVertexIndex = minVertex;
            }
        }
    }
    return minCostIndex;
}

func Rpo_PrintEdge(edge *Rpo_LCI_Edge_t) {
    Abc_Print(-2, "Edge (%d,%d)/%d\n", edge.idx1, edge.idx2, edge.connectionType);
}

/**Function*************************************************************

  Synopsis    [Test]

  Description []
               
  SideEffects []

  SeeAlso     []

 ***********************************************************************/
func Rpo_Factorize(target *uint32, nVars int32, nThreshold int32, verbose int32) *Literal_t {

    var nLitCap int32  = nVars * 2;
    var nLit int32;
    var w int32;
    var vecLit **Literal_t;
    var lit *Literal_t;
    var result *Literal_t;
    var thresholdCount int32;

    if Kit_TruthIsConst0(target, nVars) {
        return Lit_CreateLiteralConst(target, nVars, 0);
    } else if Kit_TruthIsConst1(target, nVars) {
        return Lit_CreateLiteralConst(target, nVars, 1);
    }

    //    if (nThreshold == -1) {
    //        nThreshold = nLitCap + nVars;
    //    }
    if verbose != 0 {
        Abc_Print(-2, "Target: ");
        Lit_PrintTT(target, nVars);
        Abc_Print(-2, "\n");
    }
    vecLit = ABC_ALLOC(Literal_t*, nLitCap);

    for w = nVars - 1; w >= 0; w-- {
        lit = Lit_Alloc(target, nVars, w, '+');
        if lit != nil {
            vecLit[nLit] = lit;
            nLit++;
        }
        lit = Lit_Alloc(target, nVars, w, '-');
        if lit != nil {
            vecLit[nLit] = lit;
            nLit++;
        }
    }
    if verbose != 0 {
        Abc_Print(-2, "Allocated %d literal clusters\n", nLit);
    }

    result = Rpo_Recursion(target, vecLit, nLit, nLit, nVars, &thresholdCount, nThreshold, verbose);

    //freeing the memory
    for w = 0; w < nLit; ++w {
        Lit_Free(vecLit[w]);
    }
    ABC_FREE(vecLit);

    return result;

}

func Rpo_Recursion(target *uint32, vecLit **Literal_t, nLit int32, nLitCount int32, nVars int32, thresholdCount *int32, thresholdMax int32, verbose int32) *Literal_t {
    var i, j, k int32;
    var copyResult *Literal_t;
    var vertexDegree *int32;
    var v int32;
    var edgeSize int32;
    var edges **Rpo_LCI_Edge_t;
    var edgeCount int32;
    var isAnd int32;
    var isOr int32;
    var edge *Rpo_LCI_Edge_t;
    var result *Literal_t  = nil;
    var edgeIndex int32;
    var minLitIndex int32;
    var maxLitIndex int32;
    var oldLit1 *Literal_t;
    var oldLit2 *Literal_t;
    var newLit *Literal_t;

    *thresholdCount = *thresholdCount + 1;
    if *thresholdCount == thresholdMax {
        return nil;
    }
    if verbose != 0 {
        Abc_Print(-2, "Entering recursion %d\n", *thresholdCount);
    }
    // verify if solution is the target or not
    if nLitCount == 1 {
        if verbose != 0 {
            Abc_Print(-2, "Checking solution: ");
        }
        for k = 0; k < nLit; ++k {
            if vecLit[k] != nil {
                if Kit_TruthIsEqual(target, vecLit[k].function, nVars) {
                    copyResult = Lit_Copy(vecLit[k], nVars);
                    if verbose != 0 {
                        Abc_Print(-2, "FOUND!\n", thresholdCount);
                    }
                    thresholdCount = 0; //??
                    return copyResult;
                }
            }
        }
        if verbose != 0 {
            Abc_Print(-2, "FAILED!\n", thresholdCount);
        }
        return nil;
    }

    vertexDegree = ABC_ALLOC(int, nLit);
    //    if(verbose) {
    //        Abc_Print(-2,"Allocating vertexDegree...\n");
    //    }
    for v = 0; v < nLit; v++ {
        vertexDegree[v] = 0;
    }
    // building edges
    edgeSize = (nLit * (nLit - 1)) / 2;
    edges = ABC_ALLOC(Rpo_LCI_Edge_t*, edgeSize);
    if verbose != 0 {
        Abc_Print(-2, "Creating Edges: \n");
    }

    for i = 0; i < nLit; ++i {
        if vecLit[i] == nil {
            continue;
        }
        for j = i; j < nLit; ++j {
            if vecLit[j] == nil {
                continue;
            }
            isAnd = Rpo_CheckANDGroup(vecLit[i], vecLit[j], nVars);
            isOr = Rpo_CheckORGroup(vecLit[i], vecLit[j], nVars);
            if isAnd != 0 {
                if verbose != 0 {
                    Abc_Print(-2, "Grouped: ");
                    Lit_PrintExp(vecLit[j]);
                    Abc_Print(-2, " AND ");
                    Lit_PrintExp(vecLit[i]);
                    Abc_Print(-2, "\n");
                }
                // add edge
                edge = Rpo_CreateEdge(LIT_AND, i, j, vertexDegree);
                edges[edgeCount++] = edge;
            }
            if isOr != 0 {
                if verbose != 0 {
                    Abc_Print(-2, "Grouped: ");
                    Lit_PrintExp(vecLit[j]);
                    Abc_Print(-2, " OR ");
                    Lit_PrintExp(vecLit[i]);
                    Abc_Print(-2, "\n");
                }
                // add edge
                edge = Rpo_CreateEdge(LIT_OR, i, j, vertexDegree);
                edges[edgeCount++] = edge;
            }
        }
    }
    if verbose != 0 {
        Abc_Print(-2, "%d edges created.\n", edgeCount);
    }


    //traverse the edges, grouping new Literal Clusters
    for {
        edgeIndex = Rpo_computeMinEdgeCost(edges, edgeCount, vertexDegree);
        if edgeIndex < 0 {
            if verbose != 0 {
                Abc_Print(-2, "There is no edges unvisited... Exiting recursion.\n");
                //exit(-1);
            }
            break;
            //return NULL; // the graph does not have unvisited edges
        }
        edge = edges[edgeIndex];
        edge.visited = 1;
        //Rpo_PrintEdge(edge);
        if edge.idx1 < edge.idx2 {
            minLitIndex = edge.idx1;
        } else {
            minLitIndex = edge.idx2;
        }
        if edge.idx1 > edge.idx2 {
            maxLitIndex = edge.idx1;
        } else {
            maxLitIndex = edge.idx2;
        }
        oldLit1 = vecLit[minLitIndex];
        oldLit2 = vecLit[maxLitIndex];
        newLit = Lit_GroupLiterals(oldLit1, oldLit2, (Operator_t)edge.connectionType, nVars);
        vecLit[minLitIndex] = newLit;
        vecLit[maxLitIndex] = nil;

        if verbose != 0 {
            Abc_Print(-2, "New Literal Cluster found: ");
            Lit_PrintExp(newLit);
            Abc_Print(-2, " -> ");
            Lit_PrintTT(newLit.function, nVars);
            Abc_Print(-2, "\n");
        }
        result = Rpo_Recursion(target, vecLit, nLit, (nLitCount - 1), nVars, thresholdCount, thresholdMax, verbose);
        //independent of result, free the newLit and restore the vector of Literal Clusters
        Lit_Free(newLit);
        vecLit[minLitIndex] = oldLit1;
        vecLit[maxLitIndex] = oldLit2;
        if *thresholdCount == thresholdMax {
            break;
        }
    
if !(result == nil) { break }
} 
    //freeing memory
    //    if(verbose) {
    //        Abc_Print(-2,"Freeing vertexDegree...\n");
    //    }
    ABC_FREE(vertexDegree);
    for i = 0; i < edgeCount; ++i {
        //Abc_Print(-2, "%p ", edges[i]);
        ABC_FREE(edges[i]);
    }
    ABC_FREE(edges);
    return result;
}

//ABC_NAMESPACE_IMPL_END
