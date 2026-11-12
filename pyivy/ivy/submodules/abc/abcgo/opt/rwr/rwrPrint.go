package rwr

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [rwrCut.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [Cut computation.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwrCut.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////
 
/**Function*************************************************************

  Synopsis    [Adds one node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_Trav2_rec(p *Rwr_Man_t, pNode *Rwr_Node_t, pVolume *int32) {
    if pNode.fUsed || pNode.TravId == p.nTravIds {
        return; }
    pNode.TravId = p.nTravIds;
    (*pVolume)++;
    Rwr_Trav2_rec( p, Rwr_Regular(pNode.p0), pVolume );
    Rwr_Trav2_rec( p, Rwr_Regular(pNode.p1), pVolume );
}

/**Function*************************************************************

  Synopsis    [Adds the node to the end of the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_GetBushVolume(p *Rwr_Man_t, Entry int32, pVolume *int32, pnFuncs *int32) {
    var pNode *Rwr_Node_t;
    var Volume int32;
    var nFuncs int32;
    Rwr_ManIncTravId( p );
    for  pNode = p.pTable[Entry]; (pNode != nil); pNode = pNode.pNext  {
        if pNode.uTruth != p.puCanons[pNode.uTruth] {
            continue; }
        nFuncs++;
        Rwr_Trav2_rec( p, pNode, &Volume );
    }
    *pVolume = Volume;
    *pnFuncs = nFuncs;
}

/**Function*************************************************************

  Synopsis    [Adds the node to the end of the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_GetBushSumOfVolumes(p *Rwr_Man_t, Entry int32) int32 {
    var pNode *Rwr_Node_t;
    var Volume, VolumeTotal int32;
    for  pNode = p.pTable[Entry]; (pNode != nil); pNode = pNode.pNext  {
        if pNode.uTruth != p.puCanons[pNode.uTruth] {
            continue; }
        Volume = 0;
        Rwr_ManIncTravId( p );
        Rwr_Trav2_rec( p, pNode, &Volume );
        VolumeTotal += Volume;
    }
    return VolumeTotal;
}

/**Function*************************************************************

  Synopsis    [Prints one rwr node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_NodePrint_rec(pFile *os.File, pNode *Rwr_Node_t) {
    assert(!Rwr_IsComplement(pNode), "!Rwr_IsComplement(pNode)");

    if pNode.Id == 0 {
        fmt.Fprintf(pFile, "Const1");
        return;
    }

    if pNode.Id < 5 {
        fmt.Fprintf(pFile, "%c", 'a' + pNode.Id - 1);
        return;
    }

    if Rwr_IsComplement(pNode.p0) {
        if Rwr_Regular(pNode.p0).Id < 5 {
            Rwr_NodePrint_rec( pFile, Rwr_Regular(pNode.p0) );
            fmt.Fprintf(pFile, "\'");
        
        } else {
            fmt.Fprintf(pFile, "(");
            Rwr_NodePrint_rec( pFile, Rwr_Regular(pNode.p0) );
            fmt.Fprintf(pFile, ")\'");
        }
    
    } else {
        if Rwr_Regular(pNode.p0).Id < 5 {
            Rwr_NodePrint_rec( pFile, Rwr_Regular(pNode.p0) );
        
        } else {
            fmt.Fprintf(pFile, "(");
            Rwr_NodePrint_rec( pFile, Rwr_Regular(pNode.p0) );
            fmt.Fprintf(pFile, ")");
        }
    }

    if pNode.fExor {
        fmt.Fprintf(pFile, "+"); }

    if Rwr_IsComplement(pNode.p1) {
        if Rwr_Regular(pNode.p1).Id < 5 {
            Rwr_NodePrint_rec( pFile, Rwr_Regular(pNode.p1) );
            fmt.Fprintf(pFile, "\'");
        
        } else {
            fmt.Fprintf(pFile, "(");
            Rwr_NodePrint_rec( pFile, Rwr_Regular(pNode.p1) );
            fmt.Fprintf(pFile, ")\'");
        }
    
    } else {
        if Rwr_Regular(pNode.p1).Id < 5 {
            Rwr_NodePrint_rec( pFile, Rwr_Regular(pNode.p1) );
        
        } else {
            fmt.Fprintf(pFile, "(");
            Rwr_NodePrint_rec( pFile, Rwr_Regular(pNode.p1) );
            fmt.Fprintf(pFile, ")");
        }
    }
}

/**Function*************************************************************

  Synopsis    [Prints one rwr node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_NodePrint(pFile *os.File, p *Rwr_Man_t, pNode *Rwr_Node_t) {
    var uTruth uint32;
    fmt.Fprintf(pFile, "%5d : ", pNode.Id);
    uTruth = pNode.uTruth;
    Extra_PrintHex( pFile, &uTruth, 4 );
    fmt.Fprintf(pFile, " tt=");
    Extra_PrintBinary( pFile, &uTruth, 16 );
//    fprintf( pFile, " cn=", pNode->Id );
//    uTruth = p->puCanons[pNode->uTruth];
//    Extra_PrintBinary( pFile, &uTruth, 16 );
    fmt.Fprintf(pFile, " lev=%d", pNode.Level);
    fmt.Fprintf(pFile, " vol=%d", pNode.Volume);
    fmt.Fprintf(pFile, "  ");
    Rwr_NodePrint_rec( pFile, pNode );
    fmt.Fprintf(pFile, "\n");
}

/**Function*************************************************************

  Synopsis    [Prints one rwr node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManPrint(p *Rwr_Man_t) {
    var pFile *os.File;
    var pNode *Rwr_Node_t;
    var uTruth uint32;
    var Limit, Counter, Volume, nFuncs, i int32;
    pFile = os.OpenFile("graph_lib.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    Counter = 0;
    Limit = (1 << 16);
    for  i = 0; i < Limit; i++  {
        if p.pTable[i] == nil {
            continue; }
        if i != p.puCanons[i] {
            continue; }
        fmt.Fprintf(pFile, "\nClass %3d. Func %6d.  ", p.pMap[i], Counter++);
        Rwr_GetBushVolume( p, i, &Volume, &nFuncs );
        fmt.Fprintf(pFile, "Roots = %3d. Vol = %3d. Sum = %3d.  ", nFuncs, Volume, Rwr_GetBushSumOfVolumes(p, i));
        uTruth = i;
        Extra_PrintBinary( pFile, &uTruth, 16 );
        fmt.Fprintf(pFile, "\n");
        for  pNode = p.pTable[i]; (pNode != nil); pNode = pNode.pNext  {
            if pNode.uTruth == p.puCanons[pNode.uTruth] {
                Rwr_NodePrint( pFile, p, pNode ); }
        }
    }
    _ = (pFile).Close();
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

