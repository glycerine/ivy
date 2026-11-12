package extra

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [extraUtilProgress.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [extra]

  Synopsis    [Progress bar.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: extraUtilProgress.c,v 1.0 2003/02/01 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <stdio.h>
//#include "extra.h"
//#include "base/main/main.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type ProgressBarStruct struct {
    nItemsNext              int32;   // the number of items for the next update of the progress bar
    nItemsTotal              int32;  // the total number of items
    posTotal              int32;     // the total number of positions
    posCur              int32;       // the current position
    pFile            *os.File;        // the output stream 
}

// static void Extra_ProgressBarShow( ProgressBar * p, char * pString );

// static void Extra_ProgressBarClean( ProgressBar * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts the progress bar.]

  Description [The first parameter is the output stream (pFile), where
  the progress is printed. The current printing position should be the
  first one on the given line. The second parameters is the total
  number of items that correspond to 100% position of the progress bar.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ProgressBarStart(pFile *os.File, nItemsTotal int32) *ProgressBar {
    var p *ProgressBar;
    if !Abc_FrameShowProgress(Abc_FrameGetGlobalFrame()) { return nil; }
    // inlining of macro ABC_ALLOC
    p = ((ProgressBar *) malloc(sizeof(ProgressBar) * (size_t)(1)));
    memset( p, 0, sizeof(ProgressBar) );
    p.pFile       = pFile;
    p.nItemsTotal = nItemsTotal;
    p.posTotal    = 78;
    p.posCur      = 1;
    p.nItemsNext  = int32(((7.0+p.posCur)*p.nItemsTotal/p.posTotal));
    Extra_ProgressBarShow( p, nil );
    return p;
}

/**Function*************************************************************

  Synopsis    [Updates the progress bar.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ProgressBarUpdate_int(p *ProgressBar, nItemsCur int32, pString string) {
    if p == nil { return; }
    if nItemsCur < p.nItemsNext {
        return; }
    if nItemsCur >= p.nItemsTotal {
        p.posCur = 78;
        p.nItemsNext = 0x7FFFFFFF;
    
    } else {
        p.posCur += 7;
        p.nItemsNext = int32(((7.0+p.posCur)*p.nItemsTotal/p.posTotal));
    }
    Extra_ProgressBarShow( p, pString );
}


/**Function*************************************************************

  Synopsis    [Stops the progress bar.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ProgressBarStop(p *ProgressBar) {
    if p == nil { return; }
    Extra_ProgressBarClean( p );
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prints the progress bar of the given size.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ProgressBarShow(p *ProgressBar, pString string) {
    var i int32;
    if p == nil { 
        return; }
    if Abc_FrameIsBatchMode() {
        return; }
    if pString {
        fmt.Fprintf(p.pFile, "%s ", pString); }
    for  i = (pString? len(pString) + 1 : 0); i < p.posCur; i++  {
        fmt.Fprintf(p.pFile, "-");
    }
    if i == p.posCur {
        fmt.Fprintf(p.pFile, ">"); }
    for  i++  ; i <= p.posTotal; i++  {
        fmt.Fprintf(p.pFile, " ");
    }
    fmt.Fprintf(p.pFile, "\r");
    fflush( os.Stdout );
}

/**Function*************************************************************

  Synopsis    [Cleans the progress bar before quitting.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Extra_ProgressBarClean(p *ProgressBar) {
    var i int32;
    if p == nil { 
        return; }
    if Abc_FrameIsBatchMode() {
        return; }
    for  i = 0; i <= p.posTotal; i++  {
        fmt.Fprintf(p.pFile, " ");
    }
    fmt.Fprintf(p.pFile, "\r");
    fflush( os.Stdout );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

