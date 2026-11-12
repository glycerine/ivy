package io

import (
	"fmt"
	"math"
	"strings"
)

import "os"

/**CFile****************************************************************

  FileName    [ioWriteBook.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Command processing package.]

  Synopsis    [Procedures to write Bookshelf files.]

  Author      [Myungchul Kim]
  
  Affiliation [U of Michigan]

  Date        [Ver. 1.0. Started - October 25, 2008.]

  Revision    [$Id: ioWriteBook.c,v 1.00 2005/11/10 00:00:00 mckima Exp $]

***********************************************************************/

//#include <math.h>

//#include "base/main/main.h"
//#include "map/mio/mio.h"
//#include "ioAbc.h"

//ABC_NAMESPACE_IMPL_START

const NODES =       0
const PL =       1
const coreHeight = 1
const termWidth =  1
const termHeight = 1

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static unsigned Io_NtkWriteNodes( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWritePiPoNodes( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteLatchNode( FILE * pFile, Abc_Obj_t * pLatch, int NodesOrPl );

// static unsigned Io_NtkWriteIntNode( FILE * pFile, Abc_Obj_t * pNode, int NodesOrPl );

// static unsigned Io_NtkWriteNodeGate( FILE * pFile, Abc_Obj_t * pNode );

// static void Io_NtkWriteNets( FILE * pFile, Abc_Ntk_t * pNtk );

// static void Io_NtkWriteIntNet( FILE * pFile, Abc_Obj_t * pNode );

// static void Io_NtkBuildLayout( FILE * pFile1, FILE *pFile2, Abc_Ntk_t * pNtk, double aspectRatio, double whiteSpace, unsigned coreCellArea );

// static void Io_NtkWriteScl( FILE * pFile, unsigned numCoreRows, double layoutWidth );

// static void Io_NtkWritePl( FILE * pFile, Abc_Ntk_t * pNtk, unsigned numTerms, double layoutHeight, double layoutWidth );

// static Vec_Ptr_t * Io_NtkOrderingPads( Abc_Ntk_t * pNtk, Vec_Ptr_t * vTerms );
 
// static Abc_Obj_t * Io_NtkBfsPads( Abc_Ntk_t * pNtk, Abc_Obj_t * pCurrEntry, unsigned numTerms, int * pOrdered );

// static int Abc_NodeIsNand2( Abc_Obj_t * pNode );

// static int Abc_NodeIsNor2( Abc_Obj_t * pNode );

// static int Abc_NodeIsAnd2( Abc_Obj_t * pNode );

// static int Abc_NodeIsOr2( Abc_Obj_t * pNode );

// static int Abc_NodeIsXor2( Abc_Obj_t * pNode );

// static int Abc_NodeIsXnor2( Abc_Obj_t * pNode );


func Abc_Rint(x float64) float64 { return float64(int32(x));  }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Write the network into a Bookshelf file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBookLogic(pNtk *Abc_Ntk_t, FileName string) {
    var pNtkTemp *Abc_Ntk_t;
    // derive the netlist
    pNtkTemp = Abc_NtkToNetlist(pNtk);
    if pNtkTemp == nil {
        fmt.Fprintf(os.Stdout, "Writing BOOK has failed.\n");
        return;
    }
    Io_WriteBook( pNtkTemp, FileName );
    Abc_NtkDelete( pNtkTemp );
}

/**Function*************************************************************

  Synopsis    [Write the network into a BOOK file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_WriteBook(pNtk *Abc_Ntk_t, FileName string) {

    var pFileNodes, pFileNets, pFileAux *os.File;
    var pFileScl, pFilePl, pFileWts *os.File;
    // inlining of macro ABC_CALLOC
    var FileExt string  = ((char *) calloc((size_t)(len(FileName)+7), sizeof(char)));
    var coreCellArea uint32;
    var pExdc, pNtkTemp *Abc_Ntk_t;
    var i int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // start writing the files
    FileExt = FileName;
    pFileNodes = os.OpenFile(strcat(FileExt,".nodes"), os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    FileExt = FileName;
    pFileNets  = os.OpenFile(strcat(FileExt,".nets"), os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    FileExt = FileName;
    pFileAux   = os.OpenFile(strcat(FileExt,".aux"), os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);

        // write the aux file
    if (pFileNodes == nil) || (pFileNets == nil) || (pFileAux == nil) {
        _ = (pFileAux).Close();
        fmt.Fprintf(os.Stdout, "Io_WriteBook(): Cannot open the output files.\n");
        return;
    }
    fmt.Fprintf(pFileAux, "RowBasedPlacement : %s.nodes %s.nets %s.scl %s.pl %s.wts", FileName, FileName, FileName, FileName, FileName);
    _ = (pFileAux).Close();

    // write the master network
    coreCellArea+=Io_NtkWriteNodes( pFileNodes, pNtk );
    Io_NtkWriteNets( pFileNets, pNtk );

    // write EXDC network if it exists
    pExdc = Abc_NtkExdc( pNtk );
    if pExdc {
    coreCellArea+=Io_NtkWriteNodes( pFileNodes, pNtk );
        Io_NtkWriteNets( pFileNets, pNtk );
    }

    // make sure there is no logic hierarchy
    assert(Abc_NtkWhiteboxNum(pNtk) == 0, "Abc_NtkWhiteboxNum(pNtk) == 0");

    // write the hierarchy if present
    if Abc_NtkBlackboxNum(pNtk) > 0 {
        // inlining of macro Vec_PtrForEachEntry
        for  i = 0; (i < Vec_PtrSize(pNtk.pDesign.vModules)) && (((pNtkTemp) = (Abc_Ntk_t *)Vec_PtrEntry(pNtk.pDesign.vModules, i)), 1); i++  {
            if pNtkTemp == pNtk {
                continue; }
            coreCellArea+=Io_NtkWriteNodes( pFileNodes, pNtkTemp );
            Io_NtkWriteNets( pFileNets, pNtkTemp );
        }
    }
    _ = (pFileNodes).Close();
    _ = (pFileNets).Close();

    FileExt = FileName;
    pFileScl = os.OpenFile(strcat(FileExt,".scl"), os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    FileExt = FileName;
    pFilePl  = os.OpenFile(strcat(FileExt,".pl"), os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    FileExt = FileName;
    pFileWts   = os.OpenFile(strcat(FileExt,".wts"), os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    // inlining of macro ABC_FREE
    ((FileExt) ? (free((char *) (FileExt)), (FileExt) = 0) : 0);

    Io_NtkBuildLayout( pFileScl, pFilePl, pNtk, 1.0, 10, coreCellArea );
    _ = (pFileScl).Close();
    _ = (pFilePl).Close();
    _ = (pFileWts).Close();
}

/**Function*************************************************************

  Synopsis    [Write the network into a BOOK file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteNodes(pFile *os.File, pNtk *Abc_Ntk_t) uint32 {
    var pProgress *ProgressBar;
    var pLatch, pNode *Abc_Obj_t;
    var numTerms, numNodes, coreCellArea uint32;
    var i int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // write the forehead
    numTerms=Abc_NtkPiNum(pNtk)+Abc_NtkPoNum(pNtk);
    numNodes=numTerms+Abc_NtkNodeNum(pNtk)+Abc_NtkLatchNum(pNtk);
    fmt.Printf("NumNodes : %d\t", numNodes);
    fmt.Printf("NumTerminals : %d\n", numTerms);
    fmt.Fprintf(pFile, "UCLA    nodes    1.0\n");
    fmt.Fprintf(pFile, "NumNodes : %d\n", numNodes);
    fmt.Fprintf(pFile, "NumTerminals : %d\n", numTerms);
    // write the PI/POs
    Io_NtkWritePiPoNodes( pFile, pNtk );
    // write the latches
    if !Abc_NtkIsComb(pNtk) {
        // inlining of macro Abc_NtkForEachLatch
        for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
            Io_NtkWriteLatchNode( pFile, pLatch, NODES );
            coreCellArea+=6*coreHeight;
        }
        }
    }
    // write each internal node
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkNodeNum(pNtk) );
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        coreCellArea+=Io_NtkWriteIntNode( pFile, pNode, NODES );
    }
    }
    Extra_ProgressBarStop( pProgress );
    return coreCellArea;
}

/**Function*************************************************************

  Synopsis    [Writes the primary input nodes into a file]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWritePiPoNodes(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pTerm, pNet *Abc_Obj_t;
    var i int32;

    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
        pNet = Abc_ObjFanout0(pTerm);
    fmt.Fprintf(pFile, "i%s_input\t", Abc_ObjName(pNet)); 
    fmt.Fprintf(pFile, "terminal ");
    fmt.Fprintf(pFile, " %d %d\n", termWidth, termHeight);
    }
    
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
    pNet = Abc_ObjFanin0(pTerm);
    fmt.Fprintf(pFile, "o%s_output\t", Abc_ObjName(pNet)); 
    fmt.Fprintf(pFile, "terminal ");
    fmt.Fprintf(pFile, " %d %d\n", termWidth, termHeight);
    }
}

/**Function*************************************************************

  Synopsis    [Write the latch nodes into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteLatchNode(pFile *os.File, pLatch *Abc_Obj_t, NodesOrPl int32) {
    var pNetLi, pNetLo *Abc_Obj_t;
    
    pNetLi = Abc_ObjFanin0( Abc_ObjFanin0(pLatch) );
    pNetLo = Abc_ObjFanout0( Abc_ObjFanout0(pLatch) );
    /// write the latch line
    fmt.Fprintf(pFile, "%s_%s_latch\t", Abc_ObjName(pNetLi), Abc_ObjName(pNetLo));
    if NodesOrPl == NODES {
        fmt.Fprintf(pFile, " %d %d\n", 6, 1); }
}

/**Function*************************************************************

  Synopsis    [Write the internal node into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteIntNode(pFile *os.File, pNode *Abc_Obj_t, NodesOrPl int32) uint32 {
    var sizex uint32;
    var sizey uint32 = coreHeight;
    var isize uint32;
    //double nx, ny, xstep, ystep;    
    var pNeti, pNeto *Abc_Obj_t;
    var i int32;
        
    // write the network after mapping 
    if Abc_NtkHasMapping(pNode.pNtk) {
        sizex=Io_NtkWriteNodeGate( pFile, pNode );
    } else {
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNode)) && (((pNeti) = Abc_ObjFanin(pNode, i)), 1); i++  {
        fmt.Fprintf(pFile, "%s_", Abc_ObjName(pNeti));
    }
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNode)) && (((pNeto) = Abc_ObjFanout(pNode, i)), 1); i++  {
        fmt.Fprintf(pFile, "%s_", Abc_ObjName(pNeto));
    }    
    fmt.Fprintf(pFile, "name\t");
    
    if NodesOrPl == NODES {
        isize=Abc_ObjFaninNum(pNode);
        if Abc_NodeIsConst0(pNode) || Abc_NodeIsConst1(pNode) { 
            sizex=0;
        } else if Abc_NodeIsInv(pNode) {
            sizex=1;
        } else if Abc_NodeIsBuf(pNode) {
            sizex=2;
        } else {
            assert(Abc_NtkHasSop(pNode.pNtk), "Abc_NtkHasSop(pNode.pNtk)");
            if (Abc_NodeIsNand2(pNode) != 0) || (Abc_NodeIsNor2(pNode) != 0) {
            sizex=2;
        } else if (Abc_NodeIsAnd2(pNode) != 0) || (Abc_NodeIsOr2(pNode) != 0) {
            sizex=3;
        } else if (Abc_NodeIsXor2(pNode) != 0) || (Abc_NodeIsXnor2(pNode) != 0) {
            sizex=5;
        } else {
            assert(isize > 2, "isize > 2");
            sizex=isize+Abc_SopGetCubeNum((char *)pNode.pData);
        }
        }
    }
    }
    if NodesOrPl == NODES {
    fmt.Fprintf(pFile, " %d %d\n", sizex, sizey);

    // Equally place pins. Size pins needs /  isize+#output+1
    isize= isize + Abc_ObjFanoutNum(pNode) + 1;
    }
    return sizex*sizey;
    /*
    xstep = sizex / isize;
    ystep = sizey / isize;
    nx= -0.5 * sizex;
    ny= -0.5 * sizey;

    Abc_ObjForEachFanin( pNode, pFanin, i )
    {
            nx+= xstep;
            ny+= ystep;
            if (fabs(nx) < 0.001)
                    nx= 0;
            if (fabs(ny) < 0.001)
                    ny= 0;
    }
    Abc_ObjForEachFanout( pNode, pFanout, i )
    {
            nx+= xstep;
            ny+= ystep;
            if (fabs(nx) < 0.001)
                    nx= 0;
            if (fabs(ny) < 0.001)
                    ny= 0;
    }
    */
}

/**Function*************************************************************

  Synopsis    [Writes the internal node after tech mapping.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteNodeGate(pFile *os.File, pNode *Abc_Obj_t) uint32 {
    var pGate *Mio_Gate_t  = (Mio_Gate_t *)pNode.pData;
    var pGatePin *Mio_Pin_t;
    var i int32;
    // write the node gate
    for  pGatePin = Mio_GateReadPins(pGate), i = 0; (pGatePin != nil); pGatePin = Mio_PinReadNext(pGatePin), i++  {
        fmt.Fprintf(pFile, "%s_", Abc_ObjName( Abc_ObjFanin(pNode,i) ));
    }
    assert(i == Abc_ObjFaninNum(pNode), "i == Abc_ObjFaninNum(pNode)");
    fmt.Fprintf(pFile, "%s_%s\t", Abc_ObjName( Abc_ObjFanout0(pNode) ), Mio_GateReadName(pGate));
    return Mio_GateReadArea(pGate);
}

/**Function*************************************************************

  Synopsis    [Write the nets into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteNets(pFile *os.File, pNtk *Abc_Ntk_t) {
    var pProgress *ProgressBar;
    var pNet *Abc_Obj_t;
    var numPin uint32;
    var i int32;

    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    // write the head
    // inlining of macro Abc_NtkForEachNet
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNet) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pNet) == nil || !Abc_ObjIsNet(pNet) {} else {
    numPin+=Abc_ObjFaninNum(pNet)+Abc_ObjFanoutNum(pNet); }
    }
    fmt.Printf("NumNets  : %d\t", Abc_NtkNetNum(pNtk));
    fmt.Printf("NumPins      : %d\n\n", numPin);
    fmt.Fprintf(pFile, "UCLA    nets    1.0\n");
    fmt.Fprintf(pFile, "NumNets : %d\n", Abc_NtkNetNum(pNtk));
    fmt.Fprintf(pFile, "NumPins : %d\n", numPin);

    // write nets
    pProgress = Extra_ProgressBarStart( os.Stdout, Abc_NtkNetNum(pNtk) );
    // inlining of macro Abc_NtkForEachNet
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNet) = Abc_NtkObj(pNtk, i)), 1); i++  {            if (pNet) == nil || !Abc_ObjIsNet(pNet) {} else {
        Extra_ProgressBarUpdate( pProgress, i, nil );
        Io_NtkWriteIntNet( pFile, pNet );
    }
    }
    Extra_ProgressBarStop( pProgress );
}

/**Function*************************************************************

  Synopsis    [Write the nets into a file.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteIntNet(pFile *os.File, pNet *Abc_Obj_t) {
    var pFanin, pFanout *Abc_Obj_t;
    var pNeti, pNeto *Abc_Obj_t;
    var pNetLi, pNetLo, pLatch *Abc_Obj_t;
    var i, j int32;
    var NetDegree int32 =Abc_ObjFaninNum(pNet)+Abc_ObjFanoutNum(pNet);

    fmt.Fprintf(pFile, "NetDegree\t:\t\t%d\t\t%s\n", NetDegree, Abc_ObjName(Abc_ObjFanin0(pNet)));

    pFanin=Abc_ObjFanin0(pNet);
    if Abc_ObjIsPi(pFanin) {
    fmt.Fprintf(pFile, "i%s_input I\n", Abc_ObjName(pNet));
    } else {
    if !Abc_NtkIsComb(pNet.pNtk) && Abc_ObjFaninNum(pFanin) && Abc_ObjIsLatch(Abc_ObjFanin0(pFanin)) {
        pLatch=Abc_ObjFanin0(pFanin);
        pNetLi=Abc_ObjFanin0(Abc_ObjFanin0(pLatch));
        pNetLo=Abc_ObjFanout0(Abc_ObjFanout0(pLatch));
        fmt.Fprintf(pFile, "%s_%s_latch I : ", Abc_ObjName(pNetLi), Abc_ObjName(pNetLo));
    
    } else {
        // inlining of macro Abc_ObjForEachFanin
        for  j = 0; (j < Abc_ObjFaninNum(pFanin)) && (((pNeti) = Abc_ObjFanin(pFanin, j)), 1); j++  {
        fmt.Fprintf(pFile, "%s_", Abc_ObjName(pNeti));
        }
        // inlining of macro Abc_ObjForEachFanout
        for  j = 0; (j < Abc_ObjFanoutNum(pFanin)) && (((pNeto) = Abc_ObjFanout(pFanin, j)), 1); j++  {
        fmt.Fprintf(pFile, "%s_", Abc_ObjName(pNeto));
        }    
        if Abc_NtkHasMapping(pNet.pNtk) {
        fmt.Fprintf(pFile, "%s : ", Mio_GateReadName((Mio_Gate_t *)pFanin.pData)); 
        } else {
        fmt.Fprintf(pFile, "name I : "); }                                
    }
    // offsets are simlply 0.00 0.00 at the moment
    fmt.Fprintf(pFile, "%.2f %.2f\n", .0, .0);
    }

    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNet)) && (((pFanout) = Abc_ObjFanout(pNet, i)), 1); i++  {
    if Abc_ObjIsPo(pFanout) {
        fmt.Fprintf(pFile, "o%s_output O\n", Abc_ObjName(pNet));
    } else {
        if !Abc_NtkIsComb(pNet.pNtk) && Abc_ObjFanoutNum(pFanout) && Abc_ObjIsLatch( Abc_ObjFanout0(pFanout) ) {
        pLatch=Abc_ObjFanout0(pFanout);
        pNetLi=Abc_ObjFanin0(Abc_ObjFanin0(pLatch));
        pNetLo=Abc_ObjFanout0(Abc_ObjFanout0(pLatch));
        fmt.Fprintf(pFile, "%s_%s_latch O : ", Abc_ObjName(pNetLi), Abc_ObjName(pNetLo));
        
        } else {
        // inlining of macro Abc_ObjForEachFanin
        for  j = 0; (j < Abc_ObjFaninNum(pFanout)) && (((pNeti) = Abc_ObjFanin(pFanout, j)), 1); j++  {
        fmt.Fprintf(pFile, "%s_", Abc_ObjName(pNeti));
        }
        // inlining of macro Abc_ObjForEachFanout
        for  j = 0; (j < Abc_ObjFanoutNum(pFanout)) && (((pNeto) = Abc_ObjFanout(pFanout, j)), 1); j++  {
        fmt.Fprintf(pFile, "%s_", Abc_ObjName(pNeto));
        }    
        if Abc_NtkHasMapping(pNet.pNtk) {
            fmt.Fprintf(pFile, "%s : ", Mio_GateReadName((Mio_Gate_t *)pFanout.pData)); 
        } else {
            fmt.Fprintf(pFile, "name O : "); }
        }
    // offsets are simlply 0.00 0.00 at the moment
    fmt.Fprintf(pFile, "%.2f %.2f\n", .0, .0);
    }
    }
}

/**Function*************************************************************

  Synopsis    [Write the network into a BOOK file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkBuildLayout(pFileScl *os.File, pFilePl *os.File, pNtk *Abc_Ntk_t, aspectRatio float64, whiteSpace float64, coreCellArea uint32) {
    var numCoreCells uint32 =Abc_NtkNodeNum(pNtk)+Abc_NtkLatchNum(pNtk);
    var targetLayoutArea float64  = coreCellArea/(1.0-(whiteSpace/100.0));
    var numCoreRows uint32;
    if aspectRatio>0.0 {
        numCoreRows = (Abc_Rint(math.Sqrt(float64(targetLayoutArea/aspectRatio))/coreHeight));
    } else {
        numCoreRows = 0;
    }
    var numTerms uint32 =Abc_NtkPiNum(pNtk)+Abc_NtkPoNum(pNtk);
    var totalWidth uint32 =coreCellArea/coreHeight;
    var layoutHeight float64  = numCoreRows * coreHeight;
    var layoutWidth float64  = Abc_Rint(targetLayoutArea/layoutHeight);
    var actualLayoutArea float64  = layoutWidth * layoutHeight;

    fmt.Printf("Core cell height(==site height) is %d\n", coreHeight);
    fmt.Printf("Total core cell width is %d giving an ave width of %f\n", totalWidth, float64((totalWidth/numCoreCells)));
    fmt.Printf("Target Dimensions:\n");
    fmt.Printf("  Area  :   %f\n", targetLayoutArea);
    fmt.Printf("  WS%%   :   %f\n", whiteSpace);
    fmt.Printf("  AR    :   %f\n", aspectRatio);
    fmt.Printf("Actual Dimensions:\n");
    fmt.Printf("  Width :   %f\n", layoutWidth);
    fmt.Printf("  Height:   %f (%d rows)\n", layoutHeight, numCoreRows);
    fmt.Printf("  Area  :   %f\n", actualLayoutArea);
    fmt.Printf("  WS%%   :   %f\n", 100*(actualLayoutArea-coreCellArea)/actualLayoutArea);
    fmt.Printf("  AR    :   %f\n\n", layoutWidth/layoutHeight);    

    Io_NtkWriteScl( pFileScl, numCoreRows, layoutWidth ); 
    Io_NtkWritePl( pFilePl, pNtk, numTerms, layoutHeight, layoutWidth );
}

/**Function*************************************************************

  Synopsis    [Write the network into a BOOK file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWriteScl(pFile *os.File, numCoreRows uint32, layoutWidth float64) {
    var origin_y int32;
    var rowOrients [2]string = [2]string{"N", "FS"};
    var symmetry byte ='Y';
    var sitewidth float64 =1.0;
    var spacing float64 =1.0;
        
    var rowId uint32;
    // write the forehead
    fmt.Fprintf(pFile, "UCLA    scl    1.0\n\n");
    fmt.Fprintf(pFile, "Numrows : %d\n\n", numCoreRows);

    for  rowId=0 ; rowId<numCoreRows ; rowId++, origin_y += coreHeight  {
    fmt.Fprintf(pFile, "CoreRow Horizontal\n");
    fmt.Fprintf(pFile, " Coordinate   : \t%d\n", origin_y);
    fmt.Fprintf(pFile, " Height       : \t%d\n", coreHeight);
    fmt.Fprintf(pFile, " Sitewidth    : \t%d\n", uint32(sitewidth));
    fmt.Fprintf(pFile, " Sitespacing  : \t%d\n", uint32(spacing));
    fmt.Fprintf(pFile, " Siteorient   : \t%s\n", rowOrients[rowId%2]);
    //if( coreRow[i].site.symmetry.rot90 || coreRow[i].site.symmetry.y || coreRow[i].site.symmetry.x )
    fmt.Fprintf(pFile, " Sitesymmetry : \t%c\n", symmetry);
    //else fprintf( pFile, "Sitesymmetry            : \t\t\t1\n" );
    fmt.Fprintf(pFile, " SubrowOrigin : \t%d Numsites :    \t%d\n", 0, uint32(layoutWidth));
    fmt.Fprintf(pFile, "End\n");
    }
}

/**Function*************************************************************

  Synopsis    [Write the network into a BOOK file with the given name.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkWritePl(pFile *os.File, pNtk *Abc_Ntk_t, numTerms uint32, layoutWidth float64, layoutHeight float64) {
    var pTerm, pLatch, pNode *Abc_Obj_t;
    var vTerms *Vec_Ptr_t  = Vec_PtrAlloc ( numTerms ); 
    var vOrderedTerms *Vec_Ptr_t  = Vec_PtrAlloc ( numTerms ); 
    var layoutPerim float64  = 2*layoutWidth + 2*layoutHeight;
    var nextLoc_x, nextLoc_y float64;
    var delta float64;
    var termsOnTop, termsOnBottom, termsOnLeft, termsOnRight uint32;
    var t uint32;
    var i int32;

    termsOnTop = termsOnBottom = uint32((Abc_Rint(numTerms*(layoutWidth/layoutPerim))));
    termsOnLeft = numTerms - (termsOnTop+termsOnBottom);
    termsOnRight = uint32((math.Ceil(float64(termsOnLeft/2.0))));
    termsOnLeft -= termsOnRight;

    // inlining of macro Abc_NtkForEachPi
    for  i = 0; (i < Abc_NtkPiNum(pNtk)) && (((pTerm) = Abc_NtkPi(pNtk, i)), 1); i++  {
    Vec_PtrPush( vTerms, pTerm );
    }    
    // inlining of macro Abc_NtkForEachPo
    for  i = 0; (i < Abc_NtkPoNum(pNtk)) && (((pTerm) = Abc_NtkPo(pNtk, i)), 1); i++  {
    Vec_PtrPush( vTerms, pTerm );
    }
    // Ordering Pads
    vOrderedTerms=Io_NtkOrderingPads( pNtk, vTerms );
    assert(termsOnTop+termsOnBottom+termsOnLeft+termsOnRight == uint32(Vec_PtrSize(vOrderedTerms)), "termsOnTop+termsOnBottom+termsOnLeft+termsOnRight == uint32(Vec_PtrSize(vOrderedTerms))");

    fmt.Printf("Done constructing layout region\n");
    fmt.Printf("Terminals: %d\n", numTerms);
    fmt.Printf("  Top:     %d\n", termsOnTop);
    fmt.Printf("  Bottom:  %d\n", termsOnBottom);
    fmt.Printf("  Left:    %d\n", termsOnLeft);
    fmt.Printf("  Right:   %d\n", termsOnRight);

    fmt.Fprintf(pFile, "UCLA    pl    1.0\n\n");

    nextLoc_x = math.Floor(float64(.0));
    nextLoc_y = math.Ceil(float64(layoutHeight + 2*coreHeight));
    delta   = layoutWidth / termsOnTop;
    for t = 0; t < termsOnTop; t++ {
    pTerm = (Abc_Obj_t *)Vec_PtrEntry( vOrderedTerms, t );
    if Abc_ObjIsPi(pTerm) {
        fmt.Fprintf(pFile, "i%s_input\t\t", Abc_ObjName(Abc_ObjFanout0(pTerm))); 
    } else {
        fmt.Fprintf(pFile, "o%s_output\t\t", Abc_ObjName(Abc_ObjFanin0(pTerm))); } 
    if (t != 0) && Abc_Rint(nextLoc_x) < Abc_Rint(nextLoc_x-delta)+termWidth {
        nextLoc_x++; }
    fmt.Fprintf(pFile, "%d\t\t%d\t: %s /FIXED\n", int32(Abc_Rint(nextLoc_x)), int32(Abc_Rint(nextLoc_y)), "FS");    
    nextLoc_x += delta;
    }

    nextLoc_x = math.Floor(float64(.0));
    nextLoc_y = math.Floor(float64(.0 - 2*coreHeight - termHeight));
    delta   = layoutWidth / termsOnBottom;
    for ;t < termsOnTop+termsOnBottom; t++ {
    pTerm = (Abc_Obj_t *)Vec_PtrEntry( vOrderedTerms, t );
    if Abc_ObjIsPi(pTerm) {
        fmt.Fprintf(pFile, "i%s_input\t\t", Abc_ObjName(Abc_ObjFanout0(pTerm))); 
    } else {
        fmt.Fprintf(pFile, "o%s_output\t\t", Abc_ObjName(Abc_ObjFanin0(pTerm))); } 
    if t!=termsOnTop && Abc_Rint(nextLoc_x) < Abc_Rint(nextLoc_x-delta)+termWidth {
        nextLoc_x++; }
    fmt.Fprintf(pFile, "%d\t\t%d\t: %s /FIXED\n", int32(Abc_Rint(nextLoc_x)), int32(Abc_Rint(nextLoc_y)), "N");    
    nextLoc_x     += delta;
    }

    nextLoc_x = math.Floor(float64(.0-2*coreHeight-termWidth));
    nextLoc_y = math.Floor(float64(.0));
    delta   = layoutHeight / termsOnLeft;
    for ;t < termsOnTop+termsOnBottom+termsOnLeft; t++ {
    pTerm = (Abc_Obj_t *)Vec_PtrEntry( vOrderedTerms, t );
    if Abc_ObjIsPi(pTerm) {
        fmt.Fprintf(pFile, "i%s_input\t\t", Abc_ObjName(Abc_ObjFanout0(pTerm))); 
    } else {
        fmt.Fprintf(pFile, "o%s_output\t\t", Abc_ObjName(Abc_ObjFanin0(pTerm))); } 
    if Abc_Rint(nextLoc_y) < Abc_Rint(nextLoc_y-delta)+termHeight {
        nextLoc_y++; }
    fmt.Fprintf(pFile, "%d\t\t%d\t: %s /FIXED\n", int32(Abc_Rint(nextLoc_x)), int32(Abc_Rint(nextLoc_y)), "E");    
    nextLoc_y     += delta;
    }

    nextLoc_x = math.Ceil(float64(layoutWidth+2*coreHeight));
    nextLoc_y = math.Floor(float64(.0));
    delta   = layoutHeight / termsOnRight;
    for ;t < termsOnTop+termsOnBottom+termsOnLeft+termsOnRight; t++ {
           pTerm = (Abc_Obj_t *)Vec_PtrEntry( vOrderedTerms, t );
    if Abc_ObjIsPi(pTerm) {
        fmt.Fprintf(pFile, "i%s_input\t\t", Abc_ObjName(Abc_ObjFanout0(pTerm))); 
    } else {
        fmt.Fprintf(pFile, "o%s_output\t\t", Abc_ObjName(Abc_ObjFanin0(pTerm))); } 
    if Abc_Rint(nextLoc_y) < Abc_Rint(nextLoc_y-delta)+termHeight {
        nextLoc_y++; }
    fmt.Fprintf(pFile, "%d\t\t%d\t: %s /FIXED\n", int32(Abc_Rint(nextLoc_x)), int32(Abc_Rint(nextLoc_y)), "FW");    
    nextLoc_y     += delta;
    }

    if !Abc_NtkIsComb(pNtk) {
    // inlining of macro Abc_NtkForEachLatch
    for  i = 0; (i < Vec_PtrSize((pNtk).vBoxes)) && (((pLatch) = Abc_NtkBox(pNtk, i)), 1); i++  {           if !Abc_ObjIsLatch(pLatch) {} else {
        Io_NtkWriteLatchNode( pFile, pLatch, PL );
        fmt.Fprintf(pFile, "\t%d\t\t%d\t: %s\n", 0, 0, "N");    
    }
    }
    }
    
    // inlining of macro Abc_NtkForEachNode
    for  i = 0; (i < Vec_PtrSize((pNtk).vObjs)) && (((pNode) = Abc_NtkObj(pNtk, i)), 1); i++  {           if (pNode) == nil || !Abc_ObjIsNode(pNode) {} else {
    Io_NtkWriteIntNode( pFile, pNode, PL );
    fmt.Fprintf(pFile, "\t%d\t\t%d\t: %s\n", 0, 0, "N");    
    }
    }
}

/**Function*************************************************************

  Synopsis    [Returns the closest I/O to a given I/O.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkOrderingPads(pNtk *Abc_Ntk_t, vTerms *Vec_Ptr_t) *Vec_Ptr_t { 
    var pProgress *ProgressBar;
    var numTerms uint32 =Vec_PtrSize(vTerms); 
    var termIdx, termCount uint32;
    // inlining of macro ABC_ALLOC
    var pOrdered *int32  = ((int *) malloc(sizeof(int) * (size_t)(numTerms))); 
    var newNeighbor int32 =1;
    var vOrderedTerms *Vec_Ptr_t  = Vec_PtrAlloc ( numTerms );
    var pNeighbor *Abc_Obj_t = nil;
    var pNextTerm *Abc_Obj_t;     
    var i uint32; 

    for  i=0 ; i<numTerms ; i++  {
    pOrdered[i]=0;
    } 
   
    pNextTerm = (Abc_Obj_t *)Vec_PtrEntry(vTerms, termIdx++);
    pProgress = Extra_ProgressBarStart( os.Stdout, numTerms );
    for  termCount < numTerms && termIdx < numTerms  {
    if pOrdered[Abc_ObjId(pNextTerm)] && !(newNeighbor != 0) {
        pNextTerm = (Abc_Obj_t *)Vec_PtrEntry( vTerms, termIdx++ );
        continue;
    }
    if !Vec_PtrPushUnique( vOrderedTerms, pNextTerm ) {
        pOrdered[Abc_ObjId(pNextTerm)]=1;
        termCount++; 
    }
    pNeighbor=Io_NtkBfsPads( pNtk, pNextTerm, numTerms, pOrdered );
    if newNeighbor=!Vec_PtrPushUnique( vOrderedTerms, pNeighbor ) {
       pOrdered[Abc_ObjId(pNeighbor)]=1;
       termCount++;
       pNextTerm=pNeighbor;
    
    } else if termIdx < numTerms {
        pNextTerm = (Abc_Obj_t *)Vec_PtrEntry( vTerms, termIdx++ ); }

    Extra_ProgressBarUpdate( pProgress, termCount, nil );
    }
    Extra_ProgressBarStop( pProgress );
    assert(termCount==numTerms, "termCount==numTerms");
    return vOrderedTerms;
}

/**Function*************************************************************

  Synopsis    [Returns the closest I/O to a given I/O.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Io_NtkBfsPads(pNtk *Abc_Ntk_t, pTerm *Abc_Obj_t, numTerms uint32, pOrdered *int32) *Abc_Obj_t {
    var vNeighbors *Vec_Ptr_t  = Vec_PtrAlloc ( numTerms ); 
    var pNet, pNode, pNeighbor *Abc_Obj_t;
    var foundNeighbor int32;
    var i int32;

    assert(Abc_ObjIsPi(pTerm) || Abc_ObjIsPo(pTerm), "Abc_ObjIsPi(pTerm) || Abc_ObjIsPo(pTerm)");
    Abc_NtkIncrementTravId ( pNtk );
    Abc_NodeSetTravIdCurrent( pTerm );
    if Abc_ObjIsPi(pTerm) {
    pNet = Abc_ObjFanout0(pTerm);
    // inlining of macro Abc_ObjForEachFanout
    for  i = 0; (i < Abc_ObjFanoutNum(pNet)) && (((pNode) = Abc_ObjFanout(pNet, i)), 1); i++  { 
        Vec_PtrPush( vNeighbors, pNode );
    }
    
    } else {
    pNet = Abc_ObjFanin0(pTerm);
    // inlining of macro Abc_ObjForEachFanin
    for  i = 0; (i < Abc_ObjFaninNum(pNet)) && (((pNode) = Abc_ObjFanin(pNet, i)), 1); i++  { 
        Vec_PtrPush( vNeighbors, pNode );
    }   
    }

    for  Vec_PtrSize(vNeighbors) >0  {
    pNeighbor = (Abc_Obj_t *)Vec_PtrEntry( vNeighbors, 0 );
    assert(Abc_ObjIsNode(pNeighbor) || Abc_ObjIsTerm(pNeighbor), "Abc_ObjIsNode(pNeighbor) || Abc_ObjIsTerm(pNeighbor)");
    Vec_PtrRemove( vNeighbors, pNeighbor );

    if Abc_NodeIsTravIdCurrent( pNeighbor ) {
        continue; }
    Abc_NodeSetTravIdCurrent( pNeighbor );

    if ((Abc_ObjIsPi(pNeighbor) || Abc_ObjIsPo(pNeighbor))) && !pOrdered[Abc_ObjId(pNeighbor)] {
        foundNeighbor=1;
        break;
    }
    if Abc_ObjFanoutNum( pNeighbor ) {   
        pNet=Abc_ObjFanout0( pNeighbor );
        if !Abc_NtkIsComb(pNtk) && Abc_ObjIsLatch(pNet) {
        pNet=Abc_ObjFanout0( Abc_ObjFanout0(pNet) ); }
        // inlining of macro Abc_ObjForEachFanout
        for  i = 0; (i < Abc_ObjFanoutNum(pNet)) && (((pNode) = Abc_ObjFanout(pNet, i)), 1); i++  {
        if !Abc_NodeIsTravIdCurrent(pNode) { 
            Vec_PtrPush( vNeighbors, pNode ); }
        } 
    }
    if Abc_ObjFaninNum( pNeighbor ) {
        if !Abc_NtkIsComb(pNtk) && Abc_ObjIsLatch(Abc_ObjFanin0(pNeighbor)) {
        pNeighbor=Abc_ObjFanin0( Abc_ObjFanin0(pNeighbor) ); }
        // inlining of macro Abc_ObjForEachFanin
        for  i = 0; (i < Abc_ObjFaninNum(pNeighbor)) && (((pNet) = Abc_ObjFanin(pNeighbor, i)), 1); i++  {
        if !Abc_NodeIsTravIdCurrent(pNode=Abc_ObjFanin0(pNet)) { 
            Vec_PtrPush( vNeighbors, pNode ); }
        }
    }
    }
    if foundNeighbor != 0 {
        return pNeighbor;
    } else {
        return pTerm;
    }
} 

/**Function*************************************************************

  Synopsis    [Test is the node is nand2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeIsNand2(pNode *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)"); 
    if Abc_ObjFaninNum(pNode) != 2 {
        return 0; }
    if Abc_NtkHasSop(pNtk) {
    return ( !int32(strings.Compare(((char *)pNode.pData), "-0 1\n0- 1\n")) || 
         !int32(strings.Compare(((char *)pNode.pData), "0- 1\n-0 1\n")) || 
         !int32(strings.Compare(((char *)pNode.pData), "11 0\n")) ); }
    if Abc_NtkHasMapping(pNtk) {
        return pNode.pData == (void *)Mio_LibraryReadNand2((Mio_Library_t *)Abc_FrameReadLibGen()); }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Test is the node is nand2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeIsNor2(pNode *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)"); 
    if Abc_ObjFaninNum(pNode) != 2 {
        return 0; }
    if Abc_NtkHasSop(pNtk) {
        return ( !int32(strings.Compare(((char *)pNode.pData), "00 1\n")) ); }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Test is the node is and2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeIsAnd2(pNode *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)"); 
    if Abc_ObjFaninNum(pNode) != 2 {
        return 0; }
    if Abc_NtkHasSop(pNtk) {
    return Abc_SopIsAndType(((char *)pNode.pData)); }
    if Abc_NtkHasMapping(pNtk) {
        return pNode.pData == (void *)Mio_LibraryReadAnd2((Mio_Library_t *)Abc_FrameReadLibGen()); }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Test is the node is or2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeIsOr2(pNode *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)"); 
    if Abc_ObjFaninNum(pNode) != 2 {
        return 0; }
    if Abc_NtkHasSop(pNtk) {
    return ( Abc_SopIsOrType(((char *)pNode.pData))   || 
         !int32(strings.Compare(((char *)pNode.pData), "01 0\n")) ||
         !int32(strings.Compare(((char *)pNode.pData), "10 0\n")) ||
         !int32(strings.Compare(((char *)pNode.pData), "00 0\n")) ); }
         //off-sets, too
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Test is the node is xor2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeIsXor2(pNode *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)"); 
    if Abc_ObjFaninNum(pNode) != 2 {
        return 0; }
    if Abc_NtkHasSop(pNtk) {
    return ( !int32(strings.Compare(((char *)pNode.pData), "01 1\n10 1\n")) || !int32(strings.Compare(((char *)pNode.pData), "10 1\n01 1\n"))  ); }
    assert(false, "0");
    return 0;
}

/**Function*************************************************************

  Synopsis    [Test is the node is xnor2.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Abc_NodeIsXnor2(pNode *Abc_Obj_t) int32 {
    var pNtk *Abc_Ntk_t  = pNode.pNtk;
    assert(Abc_NtkIsNetlist(pNtk), "Abc_NtkIsNetlist(pNtk)");
    assert(Abc_ObjIsNode(pNode), "Abc_ObjIsNode(pNode)"); 
    if Abc_ObjFaninNum(pNode) != 2 {
        return 0; }
    if Abc_NtkHasSop(pNtk) {
    return ( !int32(strings.Compare(((char *)pNode.pData), "11 1\n00 1\n")) || !int32(strings.Compare(((char *)pNode.pData), "00 1\n11 1\n")) ); }
    assert(false, "0");
    return 0;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

