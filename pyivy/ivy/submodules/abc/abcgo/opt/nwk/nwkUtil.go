package nwk

import (
	"fmt"
	"math"
)

import "os"

/**CFile****************************************************************

  FileName    [nwkUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <math.h>
//#include "nwk.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Increments the current traversal ID of the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManIncrementTravId(pNtk *Nwk_Man_t) {
    var pObj *Nwk_Obj_t;
    var i int32;
    if pNtk.nTravIds >= (1<<26)-1 {
        pNtk.nTravIds = 0;
        Nwk_ManForEachObj( pNtk, pObj, i )
            pObj.TravId = 0;
    }
    pNtk.nTravIds++;
}

/**Function*************************************************************

  Synopsis    [Reads the maximum number of fanins of a node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGetFaninMax(pNtk *Nwk_Man_t) int32 {
    var pNode *Nwk_Obj_t;
    var i, nFaninsMax int32;
    Nwk_ManForEachNode( pNtk, pNode, i )
    {
        if nFaninsMax < Nwk_ObjFaninNum(pNode) {
            nFaninsMax = Nwk_ObjFaninNum(pNode); }
    }
    return nFaninsMax;
}

/**Function*************************************************************

  Synopsis    [Reads the total number of all fanins.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGetTotalFanins(pNtk *Nwk_Man_t) int32 {
    var pNode *Nwk_Obj_t;
    var i, nFanins int32;
    Nwk_ManForEachNode( pNtk, pNode, i )
        nFanins += Nwk_ObjFaninNum(pNode);
    return nFanins;
}


/**Function*************************************************************

  Synopsis    [Returns the number of true PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPiNum(pNtk *Nwk_Man_t) int32 {
    var pNode *Nwk_Obj_t;
    var i, Counter int32;
    Nwk_ManForEachCi( pNtk, pNode, i )
        Counter += Nwk_ObjIsPi( pNode );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Returns the number of true POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPoNum(pNtk *Nwk_Man_t) int32 {
    var pNode *Nwk_Obj_t;
    var i, Counter int32;
    Nwk_ManForEachCo( pNtk, pNode, i )
        Counter += Nwk_ObjIsPo( pNode );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Reads the number of AIG nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManGetAigNodeNum(pNtk *Nwk_Man_t) int32 {
    var pNode *Nwk_Obj_t;
    var i, nNodes int32;
    Nwk_ManForEachNode( pNtk, pNode, i )
    {
        if pNode.pFunc == nil {
            fmt.Printf("Nwk_ManGetAigNodeNum(): Local AIG of node %d is not assigned.\n", pNode.Id);
            continue;
        }
        if Nwk_ObjFaninNum(pNode) < 2 {
            continue; }
        nNodes += Hop_DagSize( pNode.pFunc );
    }
    return nNodes;
}

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in increasing order of levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodeCompareLevelsIncrease(pp1 **Nwk_Obj_t, pp2 **Nwk_Obj_t) int32 {
    var Diff int32  = (*pp1).Level - (*pp2).Level;
    if Diff < 0 {
        return -1; }
    if Diff > 0 { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Procedure used for sorting the nodes in decreasing order of levels.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_NodeCompareLevelsDecrease(pp1 **Nwk_Obj_t, pp2 **Nwk_Obj_t) int32 {
    var Diff int32  = (*pp1).Level - (*pp2).Level;
    if Diff > 0 {
        return -1; }
    if Diff < 0 { 
        return 1; }
    return 0; 
}

/**Function*************************************************************

  Synopsis    [Prints the objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ObjPrint(pObj *Nwk_Obj_t) {
    var pNext *Nwk_Obj_t;
    var i int32;
    fmt.Printf("ObjId = %5d.  ", pObj.Id);
    if Nwk_ObjIsPi(pObj) {
        fmt.Printf("PI"); }
    if Nwk_ObjIsPo(pObj) {
        fmt.Printf("PO"); }
    if Nwk_ObjIsNode(pObj) {
        fmt.Printf("Node"); }
    fmt.Printf("   Fanins = ");
    Nwk_ObjForEachFanin( pObj, pNext, i )
        fmt.Printf("%d ", pNext.Id);
    fmt.Printf("   Fanouts = ");
    Nwk_ObjForEachFanout( pObj, pNext, i )
        fmt.Printf("%d ", pNext.Id);
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Dumps the BLIF file for the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManDumpBlif(pNtk *Nwk_Man_t, pFileName string, vPiNames *Vec_Ptr_t, vPoNames *Vec_Ptr_t) {
    var pFile *os.File;
    var vNodes *Vec_Ptr_t;
    var vTruth *Vec_Int_t;
    var vCover *Vec_Int_t;
    var pObj, pFanin *Nwk_Obj_t;
    var pMem *Aig_MmFlex_t;
    var pSop string  = nil;
    var pTruth *uint32;
    var i, k, nDigits int32;
    if Nwk_ManPoNum(pNtk) == 0 {
        fmt.Printf("Nwk_ManDumpBlif(): Network does not have POs.\n");
        return;
    }
    // collect nodes in the DFS order
    nDigits = Abc_Base10Log( Nwk_ManObjNumMax(pNtk) );
    // write the file 
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, "# BLIF file written by procedure Nwk_ManDumpBlif()\n");
//    fprintf( pFile, "# http://www.eecs.berkeley.edu/~alanmi/abc/\n" );
    fmt.Fprintf(pFile, ".model %s\n", pNtk.pName);
    // write PIs
    fmt.Fprintf(pFile, ".inputs");
    Nwk_ManForEachCi( pNtk, pObj, i )
        if ( vPiNames )
            fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPiNames, i));
        else
            fmt.Fprintf(pFile, " n%0*d", nDigits, pObj.Id);
    fmt.Fprintf(pFile, "\n");
    // write POs
    fmt.Fprintf(pFile, ".outputs");
    Nwk_ManForEachCo( pNtk, pObj, i )
        if ( vPoNames )
            fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPoNames, i));
        else
            fmt.Fprintf(pFile, " n%0*d", nDigits, pObj.Id);
    fmt.Fprintf(pFile, "\n");
    // write nodes
    pMem = Aig_MmFlexStart();
    vTruth = Vec_IntAlloc( 1 << 16 );
    vCover = Vec_IntAlloc( 1 << 16 );
    vNodes = Nwk_ManDfs( pNtk );
    Vec_PtrForEachEntry( Nwk_Obj_t *, vNodes, pObj, i )
    {
        if !Nwk_ObjIsNode(pObj) {
            continue; }
        // derive SOP for the AIG
        pTruth = Hop_ManConvertAigToTruth( pNtk.pManHop, Hop_Regular(pObj.pFunc), Nwk_ObjFaninNum(pObj), vTruth, 0 );
        if Hop_IsComplement(pObj.pFunc) {
            Kit_TruthNot( pTruth, pTruth, Nwk_ObjFaninNum(pObj) ); }
        pSop = Kit_PlaFromTruth( pMem, pTruth, Nwk_ObjFaninNum(pObj), vCover );
        // write the node
        fmt.Fprintf(pFile, ".names");
        if !Kit_TruthIsConst0(pTruth, Nwk_ObjFaninNum(pObj)) && !Kit_TruthIsConst1(pTruth, Nwk_ObjFaninNum(pObj)) {
            Nwk_ObjForEachFanin( pObj, pFanin, k )
                if ( vPiNames && Nwk_ObjIsPi(pFanin) )
                    fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPiNames, Nwk_ObjPioNum(pFanin)));
                else
                    fmt.Fprintf(pFile, " n%0*d", nDigits, pFanin.Id);
        }
        fmt.Fprintf(pFile, " n%0*d\n", nDigits, pObj.Id);
        // write the function
        fmt.Fprintf(pFile, "%s", pSop);
    }
    Vec_IntFree( vCover );
    Vec_IntFree( vTruth );
    Vec_PtrFree( vNodes );
    Aig_MmFlexStop( pMem, 0 );
    // write POs
    Nwk_ManForEachCo( pNtk, pObj, i )
    {
        fmt.Fprintf(pFile, ".names");
        if (vPiNames != nil) && Nwk_ObjIsPi(Nwk_ObjFanin0(pObj)) {
            fmt.Fprintf(pFile, " %s", (char*)Vec_PtrEntry(vPiNames, Nwk_ObjPioNum(Nwk_ObjFanin0(pObj))));
        } else {
            fmt.Fprintf(pFile, " n%0*d", nDigits, Nwk_ObjFanin0(pObj).Id); }
        if vPoNames != nil {
            fmt.Fprintf(pFile, " %s\n", (char*)Vec_PtrEntry(vPoNames, Nwk_ObjPioNum(pObj)));
        } else {
            fmt.Fprintf(pFile, " n%0*d\n", nDigits, pObj.Id); }
        fmt.Fprintf(pFile, "%d 1\n", !pObj.fInvert);
    }
    fmt.Fprintf(pFile, ".end\n\n");
    _ = (pFile).Close();
}

/**Function*************************************************************

  Synopsis    [Prints the distribution of fanins/fanouts in the network.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPrintFanioNew(pNtk *Nwk_Man_t) {
    var Buffer string // was [100]byte in C

    var pNode *Nwk_Obj_t;
    var vFanins, vFanouts *Vec_Int_t;
    var nFanins, nFanouts, nFaninsMax, nFanoutsMax, nFaninsAll, nFanoutsAll int32;
    var i, k, nSizeMax int32;

    // determine the largest fanin and fanout
    nFaninsMax = nFanoutsMax = 0;
    nFaninsAll = nFanoutsAll = 0;
    Nwk_ManForEachNode( pNtk, pNode, i )
    {
        nFanins  = Nwk_ObjFaninNum(pNode);
        nFanouts = Nwk_ObjFanoutNum(pNode);
        nFaninsAll  += nFanins;
        nFanoutsAll += nFanouts;
        nFaninsMax   = Abc_MaxInt( nFaninsMax, nFanins );
        nFanoutsMax  = Abc_MaxInt( nFanoutsMax, nFanouts );
    }

    // allocate storage for fanin/fanout numbers
    nSizeMax = Abc_MaxInt( 10 * (Abc_Base10Log(nFaninsMax) + 1), 10 * (Abc_Base10Log(nFanoutsMax) + 1) );
    vFanins  = Vec_IntStart( nSizeMax );
    vFanouts = Vec_IntStart( nSizeMax );

    // count the number of fanins and fanouts
    Nwk_ManForEachNode( pNtk, pNode, i )
    {
        nFanins  = Nwk_ObjFaninNum(pNode);
        nFanouts = Nwk_ObjFanoutNum(pNode);
//        nFanouts = Nwk_NodeMffcSize(pNode);

        if nFanins < 10 {
            Vec_IntAddToEntry( vFanins, nFanins, 1 );
        } else if nFanins < 100 {
            Vec_IntAddToEntry( vFanins, 10 + nFanins/10, 1 );
        } else if nFanins < 1000 {
            Vec_IntAddToEntry( vFanins, 20 + nFanins/100, 1 );
        } else if nFanins < 10000 {
            Vec_IntAddToEntry( vFanins, 30 + nFanins/1000, 1 );
        } else if nFanins < 100000 {
            Vec_IntAddToEntry( vFanins, 40 + nFanins/10000, 1 );
        } else if nFanins < 1000000 {
            Vec_IntAddToEntry( vFanins, 50 + nFanins/100000, 1 );
        } else if nFanins < 10000000 {
            Vec_IntAddToEntry( vFanins, 60 + nFanins/1000000, 1 ); }

        if nFanouts < 10 {
            Vec_IntAddToEntry( vFanouts, nFanouts, 1 );
        } else if nFanouts < 100 {
            Vec_IntAddToEntry( vFanouts, 10 + nFanouts/10, 1 );
        } else if nFanouts < 1000 {
            Vec_IntAddToEntry( vFanouts, 20 + nFanouts/100, 1 );
        } else if nFanouts < 10000 {
            Vec_IntAddToEntry( vFanouts, 30 + nFanouts/1000, 1 );
        } else if nFanouts < 100000 {
            Vec_IntAddToEntry( vFanouts, 40 + nFanouts/10000, 1 );
        } else if nFanouts < 1000000 {
            Vec_IntAddToEntry( vFanouts, 50 + nFanouts/100000, 1 );
        } else if nFanouts < 10000000 {
            Vec_IntAddToEntry( vFanouts, 60 + nFanouts/1000000, 1 ); }
    }

    fmt.Printf("The distribution of fanins and fanouts in the network:\n");
    fmt.Printf("         Number   Nodes with fanin  Nodes with fanout\n");
    for  k = 0; k < nSizeMax; k++  {
        if vFanins.pArray[k] == 0 && vFanouts.pArray[k] == 0 {
            continue; }
        if k < 10 {
            fmt.Printf("%15d : ", k);
        } else {
            Buffer = fmt.Sprintf("%d - %d", int32(math.Pow(float64(float64(10)), float64(k/10))) * (k%10), int32(math.Pow(float64(float64(10)), float64(k/10))) * (k%10+1) - 1); 
            printf( "%15s : ", Buffer );
        }
        if vFanins.pArray[k] == 0 {
            fmt.Printf("              ");
        } else {
            fmt.Printf("%12d  ", vFanins.pArray[k]); }
        fmt.Printf("    ");
        if vFanouts.pArray[k] == 0 {
            fmt.Printf("              ");
        } else {
            fmt.Printf("%12d  ", vFanouts.pArray[k]); }
        fmt.Printf("\n");
    }
    Vec_IntFree( vFanins );
    Vec_IntFree( vFanouts );

    fmt.Printf("Fanins: Max = %d. Ave = %.2f.  Fanouts: Max = %d. Ave =  %.2f.\n", nFaninsMax, 1.0*nFaninsAll/Nwk_ManNodeNum(pNtk), nFanoutsMax, 1.0*nFanoutsAll/Nwk_ManNodeNum(pNtk));
}

/**Function*************************************************************

  Synopsis    [Cleans the temporary marks of the nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCleanMarks(pMan *Nwk_Man_t) {
    var pObj *Nwk_Obj_t;
    var i int32;
    Nwk_ManForEachObj( pMan, pObj, i )
        pObj.MarkA = pObj.MarkB = 0;
}

/**Function*************************************************************

  Synopsis    [Minimizes the support of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManMinimumBaseNode(pObj *Nwk_Obj_t, vTruth *Vec_Int_t, fVerbose int32) int32 {
    var pTruth *uint32;
    var pFanin, pObjNew *Nwk_Obj_t;
    var pNtk *Nwk_Man_t  = pObj.pMan;
    var uSupp, nSuppSize, k, Counter int32;
    pTruth = Hop_ManConvertAigToTruth( pNtk.pManHop, Hop_Regular(pObj.pFunc), Nwk_ObjFaninNum(pObj), vTruth, 0 );
    nSuppSize = Kit_TruthSupportSize(pTruth, Nwk_ObjFaninNum(pObj));
    if nSuppSize == Nwk_ObjFaninNum(pObj) {
        return 0; }
    Counter++;
    uSupp = Kit_TruthSupport( pTruth, Nwk_ObjFaninNum(pObj) );
    // create new node with the given support
    pObjNew = Nwk_ManCreateNode( pNtk, nSuppSize, Nwk_ObjFanoutNum(pObj) );
    Nwk_ObjForEachFanin( pObj, pFanin, k )
        if ( uSupp & (1 << k) )
            Nwk_ObjAddFanin( pObjNew, pFanin );
    pObjNew.pFunc = Hop_Remap( pNtk.pManHop, pObj.pFunc, uSupp, Nwk_ObjFaninNum(pObj) );
    if fVerbose != 0 {
        fmt.Printf("Reducing node %d fanins from %d to %d.\n", pObj.Id, Nwk_ObjFaninNum(pObj), Nwk_ObjFaninNum(pObjNew)); }
    Nwk_ObjReplace( pObj, pObjNew );
    return 1;
}

/**Function*************************************************************

  Synopsis    [Minimizes the support of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManMinimumBaseInt(pNtk *Nwk_Man_t, fVerbose int32) int32 {
    var vTruth *Vec_Int_t;
    var pObj *Nwk_Obj_t;
    var i, Counter int32;
    vTruth = Vec_IntAlloc( 1 << 16 );
    Nwk_ManForEachNode( pNtk, pObj, i )
        Counter += Nwk_ManMinimumBaseNode( pObj, vTruth, fVerbose );
    if (fVerbose != 0) && (Counter != 0) {
        fmt.Printf("Support minimization reduced support of %d nodes.\n", Counter); }
    Vec_IntFree( vTruth );
    return Counter;
}

/**Function*************************************************************

  Synopsis    [Minimizes the support of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManMinimumBaseRec(pNtk *Nwk_Man_t, fVerbose int32) {
    var i int32;
    var clk abctime  = Abc_Clock();
    for  i = 0; (Nwk_ManMinimumBaseInt( pNtk, fVerbose ) != 0); i++  {}
    ABC_PRT( "Minbase", Abc_Clock() - clk );
}

/**Function*************************************************************

  Synopsis    [Minimizes the support of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManMinimumBase(pNtk *Nwk_Man_t, fVerbose int32) {
    var vTruth *Vec_Int_t;
    var pObj *Nwk_Obj_t;
    var i, Counter int32;
    vTruth = Vec_IntAlloc( 1 << 16 );
    Nwk_ManForEachNode( pNtk, pObj, i )
        Counter += Nwk_ManMinimumBaseNode( pObj, vTruth, fVerbose );
    if (fVerbose != 0) && (Counter != 0) {
        fmt.Printf("Support minimization reduced support of %d nodes.\n", Counter); }
    Vec_IntFree( vTruth );
}

/**Function*************************************************************

  Synopsis    [Minimizes the support of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManRemoveDupFaninsNode(pObj *Nwk_Obj_t, iFan0 int32, iFan1 int32, vTruth *Vec_Int_t) {
    var pManHop *Hop_Man_t  = pObj.pMan.pManHop;
//    Nwk_Obj_t * pFanin0 = pObj->pFanio[iFan0];
//    Nwk_Obj_t * pFanin1 = pObj->pFanio[iFan1];
    assert(pObj.pFanio[iFan0] == pObj.pFanio[iFan1], "pObj.pFanio[iFan0] == pObj.pFanio[iFan1]");
    pObj.pFunc = Hop_Compose( pManHop, pObj.pFunc, Hop_IthVar(pManHop,iFan0), iFan1 );
    Nwk_ManMinimumBaseNode( pObj, vTruth, 0 );
}

/**Function*************************************************************

  Synopsis    [Minimizes the support of all nodes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManRemoveDupFanins(pNtk *Nwk_Man_t, fVerbose int32) {
    var vTruth *Vec_Int_t;
    var pObj *Nwk_Obj_t;
    var i, k, m, fFound int32;
    // check if the nodes have duplicated fanins
    vTruth = Vec_IntAlloc( 1 << 16 );
    Nwk_ManForEachNode( pNtk, pObj, i )
    {
        fFound = 0;
        for  k = 0; k < pObj.nFanins; k++  {
            for  m = k + 1; m < pObj.nFanins; m++  {
                if pObj.pFanio[k] == pObj.pFanio[m] {
                    if fVerbose != 0 {
                        fmt.Printf("Removing duplicated fanins of node %d (fanins %d and %d).\n", pObj.Id, pObj.pFanio[k].Id, pObj.pFanio[m].Id); }
                    Nwk_ManRemoveDupFaninsNode( pObj, k, m, vTruth );
                    fFound = 1;
                    break;
                }
            }
            if fFound != 0 {
                break; }
        }
    }
    Vec_IntFree( vTruth );
//    Nwk_ManMinimumBase( pNtk, fVerbose );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

