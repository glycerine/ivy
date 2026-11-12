package dec

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [decPrint.c]

  PackageName [MVSIS 2.0: Multi-valued logic synthesis system.]

  Synopsis    [Procedures to print the decomposition graphs (factored forms).]

  Author      [MVSIS Group]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - February 1, 2003.]

  Revision    [$Id: decPrint.c,v 1.1 2003/05/22 19:20:05 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "dec.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void   Dec_GraphPrint_rec( FILE * pFile, Dec_Graph_t * pGraph, Dec_Node_t * pNode, int fCompl, char * pNamesIn[], int * pPos, int LitSizeMax );

// static int    Dec_GraphPrintGetLeafName( FILE * pFile, int iLeaf, int fCompl, char * pNamesIn[] );

// static void   Dec_GraphPrintUpdatePos( FILE * pFile, int * pPos, int LitSizeMax );

// static int    Dec_GraphPrintOutputName( FILE * pFile, char * pNameOut );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Prints the decomposition graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphPrint(pFile *os.File, pGraph *Dec_Graph_t, pNamesIn []string, pNameOut string) {
    var vNamesIn *Vec_Ptr_t  = nil;
    var LitSizeMax, LitSizeCur, Pos, i int32;

    // create the names if not given by the user
    if pNamesIn == nil {
        vNamesIn = Abc_NodeGetFakeNames( Dec_GraphLeaveNum(pGraph) );
        pNamesIn = (char **)vNamesIn.pArray;
    }
    if pNameOut == nil {
        pNameOut = "F"; }

    // get the size of the longest literal
    LitSizeMax = 0;
    for  i = 0; i < Dec_GraphLeaveNum(pGraph); i++  {
        LitSizeCur = len(pNamesIn[i]);
        if LitSizeMax < LitSizeCur {
            LitSizeMax = LitSizeCur; }
    }
    if LitSizeMax > 50 {
        LitSizeMax = 20; }

    // write the decomposition graph (factored form)
    if Dec_GraphIsConst(pGraph) { // constant
    
        Pos = Dec_GraphPrintOutputName( pFile, pNameOut );
        fmt.Fprintf(pFile, "Constant %d", !Dec_GraphIsComplement(pGraph));
    
    } else if Dec_GraphIsVar(pGraph) { // literal
    
        Pos = Dec_GraphPrintOutputName( pFile, pNameOut );
        Dec_GraphPrintGetLeafName( pFile, Dec_GraphVarInt(pGraph), Dec_GraphIsComplement(pGraph), pNamesIn );
    
    } else {
        Pos = Dec_GraphPrintOutputName( pFile, pNameOut );
        Dec_GraphPrint_rec( pFile, pGraph, Dec_GraphNodeLast(pGraph), Dec_GraphIsComplement(pGraph), pNamesIn, &Pos, LitSizeMax );
    }
    fmt.Fprintf(pFile, "\n");

    if vNamesIn != nil {
        Abc_NodeFreeNames( vNamesIn ); }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphPrint2_rec(pFile *os.File, pGraph *Dec_Graph_t, pNode *Dec_Node_t, fCompl int32, pNamesIn []string, pPos *int32, LitSizeMax int32) {
    var pNode0, pNode1 *Dec_Node_t;
    pNode0 = Dec_GraphNode(pGraph, pNode.eEdge0.Node);
    pNode1 = Dec_GraphNode(pGraph, pNode.eEdge1.Node);
    if Dec_GraphNodeIsVar(pGraph, pNode) { // FT_NODE_LEAF )
    
        (*pPos) += Dec_GraphPrintGetLeafName( pFile, Dec_GraphNodeInt(pGraph,pNode), fCompl, pNamesIn );
        return;
    }
    if !pNode.fNodeOr { // FT_NODE_AND )
    
        if !pNode0.fNodeOr { // != FT_NODE_OR )
            Dec_GraphPrint_rec( pFile, pGraph, pNode0, pNode.fCompl0, pNamesIn, pPos, LitSizeMax );
        } else {
            fmt.Fprintf(pFile, "(");
            (*pPos)++;
            Dec_GraphPrint_rec( pFile, pGraph, pNode0, pNode.fCompl0, pNamesIn, pPos, LitSizeMax );
            fmt.Fprintf(pFile, ")");
            (*pPos)++;
        }
        fmt.Fprintf(pFile, " ");
        (*pPos)++;

        Dec_GraphPrintUpdatePos( pFile, pPos, LitSizeMax );

        if !pNode1.fNodeOr { // != FT_NODE_OR )
            Dec_GraphPrint_rec( pFile, pGraph, pNode1, pNode.fCompl1, pNamesIn, pPos, LitSizeMax );
        } else {
            fmt.Fprintf(pFile, "(");
            (*pPos)++;
            Dec_GraphPrint_rec( pFile, pGraph, pNode1, pNode.fCompl1, pNamesIn, pPos, LitSizeMax );
            fmt.Fprintf(pFile, ")");
            (*pPos)++;
        }
        return;
    }
    if pNode.fNodeOr { // FT_NODE_OR )
    
        Dec_GraphPrint_rec( pFile, pGraph, pNode0, pNode.fCompl0, pNamesIn, pPos, LitSizeMax );
        fmt.Fprintf(pFile, " + ");
        (*pPos) += 3;

        Dec_GraphPrintUpdatePos( pFile, pPos, LitSizeMax );

        Dec_GraphPrint_rec( pFile, pGraph, pNode1, pNode.fCompl1, pNamesIn, pPos, LitSizeMax );
        return;
    }
    assert(false, "0");
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphPrint_rec(pFile *os.File, pGraph *Dec_Graph_t, pNode *Dec_Node_t, fCompl int32, pNamesIn []string, pPos *int32, LitSizeMax int32) {
    var pNode0, pNode1 *Dec_Node_t;
    var pNode00, pNode01, pNode10, pNode11 *Dec_Node_t;
    pNode0 = Dec_GraphNode(pGraph, pNode.eEdge0.Node);
    pNode1 = Dec_GraphNode(pGraph, pNode.eEdge1.Node);
    if Dec_GraphNodeIsVar(pGraph, pNode) { // FT_NODE_LEAF )
    
        (*pPos) += Dec_GraphPrintGetLeafName( pFile, Dec_GraphNodeInt(pGraph,pNode), fCompl, pNamesIn );
        return;
    }
    if !Dec_GraphNodeIsVar(pGraph, pNode0) && !Dec_GraphNodeIsVar(pGraph, pNode1) {
        pNode00 = Dec_GraphNode(pGraph, pNode0.eEdge0.Node);
        pNode01 = Dec_GraphNode(pGraph, pNode0.eEdge1.Node);
        pNode10 = Dec_GraphNode(pGraph, pNode1.eEdge0.Node);
        pNode11 = Dec_GraphNode(pGraph, pNode1.eEdge1.Node);
        if (pNode00 == pNode10 || pNode00 == pNode11) && (pNode01 == pNode10 || pNode01 == pNode11) {
            fmt.Fprintf(pFile, "(");
            (*pPos)++;
            Dec_GraphPrint_rec( pFile, pGraph, pNode00, pNode00.fCompl0, pNamesIn, pPos, LitSizeMax );
            fmt.Fprintf(pFile, " # ");
            (*pPos) += 3;
            Dec_GraphPrint_rec( pFile, pGraph, pNode01, pNode01.fCompl1, pNamesIn, pPos, LitSizeMax );
            fmt.Fprintf(pFile, ")");
            (*pPos)++;
            return;
        }
    }
    if fCompl != 0 {
        fmt.Fprintf(pFile, "(");
        (*pPos)++;
        Dec_GraphPrint_rec( pFile, pGraph, pNode0, !pNode.eEdge0.fCompl, pNamesIn, pPos, LitSizeMax );
        fmt.Fprintf(pFile, " + ");
        (*pPos) += 3;
        Dec_GraphPrint_rec( pFile, pGraph, pNode1, !pNode.eEdge1.fCompl, pNamesIn, pPos, LitSizeMax );
        fmt.Fprintf(pFile, ")");
        (*pPos)++;
    
    } else {
        fmt.Fprintf(pFile, "(");
        (*pPos)++;
        Dec_GraphPrint_rec( pFile, pGraph, pNode0, pNode.eEdge0.fCompl, pNamesIn, pPos, LitSizeMax );
        Dec_GraphPrint_rec( pFile, pGraph, pNode1, pNode.eEdge1.fCompl, pNamesIn, pPos, LitSizeMax );
        fmt.Fprintf(pFile, ")");
        (*pPos)++;
    }
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Dec_GraphPrintGetLeafName_Buffer string // was [100]byte in C

func Dec_GraphPrintGetLeafName(pFile *os.File, iLeaf int32, fCompl int32, pNamesIn []string) int32 {
    
    spStatic_Dec_GraphPrintGetLeafName_Buffer = fmt.Sprintf("%s%s", fCompl? "!" : "", pNamesIn[iLeaf]);
    fprintf( pFile, "%s", spStatic_Dec_GraphPrintGetLeafName_Buffer );
    return strlen( spStatic_Dec_GraphPrintGetLeafName_Buffer );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphPrintUpdatePos(pFile *os.File, pPos *int32, LitSizeMax int32) {
    var i int32;
    if *pPos + LitSizeMax < 77 {
        return; }
    fmt.Fprintf(pFile, "\n");
    for  i = 0; i < 10; i++  {
        fmt.Fprintf(pFile, " ");
    }
    *pPos = 10;
}

/**Function*************************************************************

  Synopsis    [Starts the printout for a decomposition graph.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Dec_GraphPrintOutputName(pFile *os.File, pNameOut string) int32 {
    if pNameOut == nil {
        return 0; }
    fmt.Fprintf(pFile, "%6s = ", pNameOut);
    return 10;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

