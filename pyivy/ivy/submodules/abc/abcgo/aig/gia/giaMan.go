package gia

import "fmt"

import "os"

/**CFile****************************************************************

  FileName    [giaMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Package manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaMan.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include <ctype.h>

//#include "gia.h"
//#include "misc/tim/tim.h"
//#include "proof/abs/abs.h"
//#include "opt/dar/dar.h"
//#include "misc/extra/extra.h"

#ifdef WIN32
#include <windows.h>
#endif

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// extern void Gia_ManDfsSlacksPrint( Gia_Man_t * p );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Creates AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Deletes AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Returns memory used in megabytes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Stops the AIG manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Duplicates AIG for unrolling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/*
void Gia_ManLogAigStats( Gia_Man_t * p, char * pDumpFile )
{
    FILE * pTable = fopen( pDumpFile, "wb" );
    fprintf( pTable, "Name = %24s     ", p->pName );
    fprintf( pTable, "In = %6d   ",      Gia_ManCiNum(p) );
    fprintf( pTable, "Out = %6d   ",     Gia_ManCoNum(p) );
    fprintf( pTable, "And = %8d   ",     Gia_ManAndNum(p) );
    fprintf( pTable, "Lev = %6d",        Gia_ManLevelNum(p) );
    fprintf( pTable, "\n" );
    fclose( pTable );
}
*/


/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Gia_ManPrintStats_nPiPo int32;
var spStatic_Gia_ManPrintStats_PrevSwiTotal float32;
func Gia_ManPrintStats(p *Gia_Man_t, pPars *Gps_Par_t) {
    extern float Gia_ManLevelAve( Gia_Man_t * p );
    var fHaveLevels int32  = p.vLevels != nil;
    if (pPars != nil) && pPars.fMiter {
        Gia_ManPrintStatsMiter( p, 0 );
        return;
    }
    if (pPars != nil) && pPars.fNoColor {
        if p.pName {
            Abc_Print( 1, "%-8s : ", p.pName ); }
    
    } else {
#ifdef WIN32
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 15 ); // bright
    if ( p->pName )
        Abc_Print( 1, "%-8s : ", p->pName );
    SetConsoleTextAttribute( GetStdHandle(STD_OUTPUT_HANDLE), 7 );  // normal
#else
    if ( p->pName )
        Abc_Print( 1, "%s%-8s%s : ", "\033[1;37m", p->pName, "\033[0m" );  // bright
#endif
    }
    Abc_Print( 1, "i/o =%7d/%7d", 
        Gia_ManPiNum(p) - Gia_ManBoxCiNum(p) - Gia_ManRegBoxNum(p), 
        Gia_ManPoNum(p) - Gia_ManBoxCoNum(p) - Gia_ManRegBoxNum(p) );
    if Gia_ManConstrNum(p) {
        Abc_Print( 1, "(c=%d)", Gia_ManConstrNum(p) ); }
    if Gia_ManRegNum(p) {
        Abc_Print( 1, "  ff =%7d", Gia_ManRegNum(p) ); }
    if Gia_ManRegBoxNum(p) {
        Abc_Print( 1, "  boxff =%d(%d)", Gia_ManRegBoxNum(p), Gia_ManClockDomainNum(p) ); }
    if (pPars != nil) && pPars.fNoColor {
        Abc_Print( 1, "  %s =%8d", p.pMuxes? "nod" : "and", Gia_ManAndNum(p) );
        Abc_Print( 1, "  lev =%5d", Gia_ManLevelNum(p) ); 
        Abc_Print( 1, " (%.2f)", Gia_ManLevelAve(p) ); 
    
    } else {
#ifdef WIN32
    {
    HANDLE hConsole = GetStdHandle(STD_OUTPUT_HANDLE);
    SetConsoleTextAttribute( hConsole, 11 ); // blue
    Abc_Print( 1, "  %s =%8d", p->pMuxes? "nod" : "and", Gia_ManAndNum(p) );
    SetConsoleTextAttribute( hConsole, 13 ); // magenta
    Abc_Print( 1, "  lev =%5d", Gia_ManLevelNum(p) ); 
    Abc_Print( 1, " (%.2f)", Gia_ManLevelAve(p) ); 
    SetConsoleTextAttribute( hConsole, 7 ); // normal
    }
#else
    Abc_Print( 1, "  %s%s =%8d%s",  "\033[1;36m", p->pMuxes? "nod" : "and", Gia_ManAndNum(p), "\033[0m" ); // blue
    Abc_Print( 1, "  %slev =%5d%s", "\033[1;35m", Gia_ManLevelNum(p), "\033[0m" ); // magenta
    Abc_Print( 1, " %s(%.2f)%s",    "\033[1;35m", Gia_ManLevelAve(p), "\033[0m" ); 
#endif
    }
    if !(fHaveLevels != 0) { 
        Vec_IntFreeP( &p.vLevels ); }
    if (pPars != nil) && pPars.fCut {
        Abc_Print( 1, "  cut = %d(%d)", Gia_ManCrossCut(p, 0), Gia_ManCrossCut(p, 1) ); }
    Abc_Print( 1, "  mem =%5.2f MB", Gia_ManMemory(p)/(1<<20) );
    if Gia_ManHasChoices(p) {
        Abc_Print( 1, "  ch =%5d", Gia_ManChoiceNum(p) ); }
    if p.pManTime {
        Abc_Print( 1, "  box = %d", Gia_ManNonRegBoxNum(p) ); }
    if p.pManTime {
        Abc_Print( 1, "  bb = %d", Gia_ManBlackBoxNum(p) ); }
    if Gia_ManBufNum(p) {
        Abc_Print( 1, "  buf = %d", Gia_ManBufNum(p) ); }
    if Gia_ManXorNum(p) && p.pMuxes == nil {
        Abc_Print( 1, "  xor = %d", Gia_ManXorNum(p) ); }
    if (pPars != nil) && pPars.fMuxXor {
        fmt.Printf("\nXOR/MUX ");
Gia_ManPrintMuxStats( p ); }
    if (pPars != nil) && pPars.fSwitch {
        
        
        var SwiTotal float32  = Gia_ManComputeSwitching( p, 48, 16, 0 );
        Abc_Print( 1, "  power =%8.1f", SwiTotal );
        if spStatic_Gia_ManPrintStats_PrevSwiTotal > 0 && spStatic_Gia_ManPrintStats_nPiPo == Gia_ManCiNum(p) + Gia_ManCoNum(p) {
            Abc_Print( 1, " %6.2f %%", 100.0*(spStatic_Gia_ManPrintStats_PrevSwiTotal-SwiTotal)/spStatic_Gia_ManPrintStats_PrevSwiTotal );
        } else if spStatic_Gia_ManPrintStats_PrevSwiTotal == 0 || spStatic_Gia_ManPrintStats_nPiPo != Gia_ManCiNum(p) + Gia_ManCoNum(p) {
            spStatic_Gia_ManPrintStats_PrevSwiTotal = SwiTotal;
spStatic_Gia_ManPrintStats_nPiPo = Gia_ManCiNum(p) + Gia_ManCoNum(p); }
    }
//    Abc_Print( 1, "obj =%5d  ", Gia_ManObjNum(p) );
    Abc_Print( 1, "\n" );

//    Gia_ManSatExperiment( p );
    if p.pReprs && p.pNexts {
        Gia_ManEquivPrintClasses( p, 0, 0.0 ); }
    if Gia_ManHasMapping(p) && (pPars == nil || !pPars.fSkipMap) {
        Gia_ManPrintMappingStats( p, pPars ? pPars.pDumpFile : nil );
    } else if (pPars != nil) && pPars.pDumpFile {
        Gia_ManLogAigStats( p, pPars.pDumpFile ); }
    if (pPars != nil) && pPars.fNpn && Gia_ManHasMapping(p) && Gia_ManLutSizeMax(p) <= 4 {
        Gia_ManPrintNpnClasses( p ); }
    if p.vPacking {
        Gia_ManPrintPackingStats( p ); }
    if p.vEdge1 {
        Gia_ManPrintEdges( p ); }
    if (pPars != nil) && pPars.fLutProf && Gia_ManHasMapping(p) {
        Gia_ManPrintLutStats( p ); }
    if p.pPlacement {
        Gia_ManPrintPlacement( p ); }
//    if ( p->pManTime )
//        Tim_ManPrintStats( (Tim_Man_t *)p->pManTime, p->nAnd2Delay );
    Gia_ManPrintFlopClasses( p );
    Gia_ManPrintGateClasses( p );
    Gia_ManPrintObjClasses( p );
//    if ( p->vRegClasses )
//    {
//        printf( "The design has %d flops with the following class info: ", Vec_IntSize(p->vRegClasses) );
//        Vec_IntPrint( p->vRegClasses );
//    }
    if p.vInitClasses {
        Gia_ManPrintInitClasses( p.vInitClasses ); }
    // check integrity of boxes
    Gia_ManCheckIntegrityWithBoxes( p );
/*
    if ( Gia_ManRegBoxNum(p) )
    {
        int i, Limit = Vec_IntFindMax(p->vRegClasses);
        for ( i = 1; i <= Limit; i++ )
            printf( "%d ", Vec_IntCountEntry(p->vRegClasses, i) );
        printf( "\n" );
    }
*/
    if (pPars != nil) && pPars.fTents {
/*
        int k, Entry, Prev = 1;
        Vec_Int_t * vLimit = Vec_IntAlloc( 1000 );
        Gia_Man_t * pNew = Gia_ManUnrollDup( p, vLimit );
        Abc_Print( 1, "Tents:  " );
        Vec_IntForEachEntryStart( vLimit, Entry, k, 1 )
            Abc_Print( 1, "%d=%d  ", k, Entry-Prev ), Prev = Entry;
        Abc_Print( 1, " Unused=%d.", Gia_ManObjNum(p) - Gia_ManObjNum(pNew) );
        Abc_Print( 1, "\n" );
        Vec_IntFree( vLimit );
        Gia_ManStop( pNew );
*/
        Gia_ManPrintTents( p );
    }
    if (pPars != nil) && pPars.fSlacks {
        Gia_ManDfsSlacksPrint( p ); }
}

/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Statistics of the miter.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Reports the reduction of the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Prints NPN class statistics.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Collects internal nodes and boxes in the DFS order.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

 

/**Function*************************************************************

  Synopsis    [Compute arrival/required times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/







/**Function*************************************************************

  Synopsis    [Compute arrival/required times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


var spStatic_Gia_ObjGetDumpName_pBuffer string // was [10000]byte in C





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
