package rwr

import (
	"fmt"
	"time"
)

import "os"

/**CFile****************************************************************

  FileName    [rwrUtil.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting package.]

  Synopsis    [Various utilities.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: rwrUtil.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "rwr.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// precomputed data
// the following 135 practical NPN classes of 4-variable functions were computed
// by considering all 4-input cuts appearing in IWLS, MCNC, and ISCAS benchmarks




////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Writes data.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManWriteToArray(p *Rwr_Man_t) {
    var pFile *os.File;
    var pNode *Rwr_Node_t;
    var Entry0, Entry1 uint32;
    var i, nEntries int32;
    var clk abctime  = Abc_Clock();
    // prepare the buffer
    nEntries = p.vForest.nSize - 5;
    pFile = os.OpenFile("npn4_aig_array.txt", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pFile, "static unsigned short s_RwtAigSubgraphs[] = \n{");
    for  i = 0; i < nEntries; i++  {
        if i % 5 == 0 {
            fmt.Fprintf(pFile, "\n    "); }
        pNode = (Rwr_Node_t *)p.vForest.pArray[i+5];
        Entry0 = (Rwr_Regular(pNode.p0).Id << 1) | Rwr_IsComplement(pNode.p0);
        Entry1 = (Rwr_Regular(pNode.p1).Id << 1) | Rwr_IsComplement(pNode.p1);
        Entry0 = (Entry0 << 1) | pNode.fExor;
        Extra_PrintHex( pFile, &Entry0, 4 );
        fmt.Fprintf(pFile, ",");
        Extra_PrintHex( pFile, &Entry1, 4 );
        fmt.Fprintf(pFile, ", ");
    }
    if i % 5 == 0 {
        fmt.Fprintf(pFile, "\n    "); }
    Entry0 = 0;
    Extra_PrintHex( pFile, &Entry0, 4 );
    fmt.Fprintf(pFile, ",");
    Extra_PrintHex( pFile, &Entry0, 4 );
    fmt.Fprintf(pFile, " \n};\n");
    _ = (pFile).Close();
    // inlining of macro ABC_PRT
    fmt.Printf("The number of nodes saved = %d.   ", nEntries);  Abc_Print(1, "%s =", ("Saving"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Loads data.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManLoadFromArray(p *Rwr_Man_t, fVerbose int32) {
    unsigned short * pArray = s_RwtAigSubgraphs;
    var p0, p1 *Rwr_Node_t;
    var Entry0, Entry1 uint32;
    var Level, Volume, nEntries, fExor int32;
    var i int32;
    var clk abctime  = Abc_Clock();

    // reconstruct the forest
    for  i = 0; ; i++  {
        Entry0 = pArray[2*i + 0];
        Entry1 = pArray[2*i + 1];
        if Entry0 == 0 && Entry1 == 0 {
            break; }
        // get EXOR flag
        fExor = (Entry0 & 1);
        Entry0 >>= 1;
        // get the nodes
        p0 = (Rwr_Node_t *)p.vForest.pArray[Entry0 >> 1];
        p1 = (Rwr_Node_t *)p.vForest.pArray[Entry1 >> 1];
        // compute the level and volume of the new nodes
        Level  = 1 + Abc_MaxInt( p0.Level, p1.Level );
        Volume = 1 + Rwr_ManNodeVolume( p, p0, p1 );
        // set the complemented attributes
        p0 = Rwr_NotCond( p0, (Entry0 & 1) );
        p1 = Rwr_NotCond( p1, (Entry1 & 1) );
        // add the node
//        Rwr_ManTryNode( p, p0, p1, Level, Volume );
        Rwr_ManAddNode( p, p0, p1, fExor, Level, Volume + fExor );
    }
    nEntries = i - 1;
    if fVerbose != 0 {
        fmt.Printf("The number of classes = %d. Canonical nodes = %d.\n", p.nClasses, p.nAdded);
        // inlining of macro ABC_PRT
        fmt.Printf("The number of nodes loaded = %d.  ", nEntries);  Abc_Print(1, "%s =", ("Loading"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
    }
}


/**Function*************************************************************

  Synopsis    [Writes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManWriteToFile(p *Rwr_Man_t, pFileName string) {
    var pFile *os.File;
    var pNode *Rwr_Node_t;
    var pBuffer *uint32;
    var i, nEntries int32;
    var clk abctime  = Abc_Clock();
    // prepare the buffer
    nEntries = p.vForest.nSize - 5;
    // inlining of macro ABC_ALLOC
    pBuffer = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nEntries * 2)));
    for  i = 0; i < nEntries; i++  {
        pNode = (Rwr_Node_t *)p.vForest.pArray[i+5];
        pBuffer[2*i + 0] = (Rwr_Regular(pNode.p0).Id << 1) | Rwr_IsComplement(pNode.p0);
        pBuffer[2*i + 1] = (Rwr_Regular(pNode.p1).Id << 1) | Rwr_IsComplement(pNode.p1);
        // save EXOR flag
        pBuffer[2*i + 0] = (pBuffer[2*i + 0] << 1) | pNode.fExor;

    }
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fwrite( &nEntries, sizeof(int), 1, pFile );
    fwrite( pBuffer, sizeof(unsigned), nEntries * 2, pFile );
    pBuffer = nil;
    _ = (pFile).Close();
    // inlining of macro ABC_PRT
    fmt.Printf("The number of nodes saved = %d.   ", nEntries);  Abc_Print(1, "%s =", ("Saving"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}

/**Function*************************************************************

  Synopsis    [Loads data.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManLoadFromFile(p *Rwr_Man_t, pFileName string) {
    var pFile *os.File;
    var p0, p1 *Rwr_Node_t;
    var pBuffer *uint32;
    var Level, Volume, nEntries, fExor int32;
    var i int32;
    var clk abctime  = Abc_Clock();
    var RetValue int32;

    // load the data
    pFile = os.Open(pFileName);
    if pFile == nil {
        fmt.Printf("Rwr_ManLoadFromFile: Cannot open file \"%s\".\n", pFileName);
        return;
    }
    RetValue = fread( &nEntries, sizeof(int), 1, pFile );
    // inlining of macro ABC_ALLOC
    pBuffer = ((unsigned *) malloc(sizeof(unsigned) * (size_t)(nEntries * 2)));
    RetValue = fread( pBuffer, sizeof(unsigned), nEntries * 2, pFile );
    _ = (pFile).Close();
    // reconstruct the forest
    for  i = 0; i < nEntries; i++  {
        // get EXOR flag
        fExor = (pBuffer[2*i + 0] & 1);
        pBuffer[2*i + 0] = (pBuffer[2*i + 0] >> 1);
        // get the nodes
        p0 = (Rwr_Node_t *)p.vForest.pArray[pBuffer[2*i + 0] >> 1];
        p1 = (Rwr_Node_t *)p.vForest.pArray[pBuffer[2*i + 1] >> 1];
        // compute the level and volume of the new nodes
        Level  = 1 + Abc_MaxInt( p0.Level, p1.Level );
        Volume = 1 + Rwr_ManNodeVolume( p, p0, p1 );
        // set the complemented attributes
        p0 = Rwr_NotCond( p0, (pBuffer[2*i + 0] & 1) );
        p1 = Rwr_NotCond( p1, (pBuffer[2*i + 1] & 1) );
        // add the node
//        Rwr_ManTryNode( p, p0, p1, Level, Volume );
        Rwr_ManAddNode( p, p0, p1, fExor, Level, Volume + fExor );
    }
    pBuffer = nil;
    fmt.Printf("The number of classes = %d. Canonical nodes = %d.\n", p.nClasses, p.nAdded);
    // inlining of macro ABC_PRT
    fmt.Printf("The number of nodes loaded = %d.   ", nEntries);  Abc_Print(1, "%s =", ("Loading"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((Abc_Clock() - clk)))/(float64(int64(time.Second))));;
}


/**Function*************************************************************

  Synopsis    [Adds the node to the end of the list.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ListAddToTail(ppList **Rwr_Node_t, pNode *Rwr_Node_t) {
    var pTemp *Rwr_Node_t;
    // find the last one
    for  pTemp = *ppList; (pTemp != nil); pTemp = pTemp.pNext  {
        ppList = &pTemp.pNext;
    }
    // attach at the end
    *ppList = pNode;
}

/**Function*************************************************************

  Synopsis    [Create practical classes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Rwr_ManGetPractical(p *Rwr_Man_t) string {
    var pPractical string;
    var i int32;
    // inlining of macro ABC_ALLOC
    pPractical = ((char *) malloc(sizeof(char) * (size_t)(p.nFuncs)));
    memset( pPractical, 0, sizeof(char) * p.nFuncs );
    pPractical[0] = 1;
    for  i = 1; ; i++  {
        if s_RwrPracticalClasses[i] == 0 {
            break; }
        pPractical[ s_RwrPracticalClasses[i] ] = 1;
    }
    return pPractical;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_IMPL_END

