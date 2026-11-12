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
func Gia_ManPrintStatsMiter(p *Gia_Man_t, fVerbose int32) {
    var pObj *Gia_Obj_t;
    var vProb *Vec_Flt_t;
    var i, iObjId int32;
    Gia_ManLevelNum( p );
    Gia_ManCreateRefs( p );
    vProb = Gia_ManPrintOutputProb( p );
    fmt.Printf("Statistics for each outputs of the miter:\n");
    Gia_ManForEachPo( p, pObj, i )
    {
        iObjId = Gia_ObjId(p, pObj);
        fmt.Printf("%4d : ", i);
        fmt.Printf("Level = %5d  ", Gia_ObjLevelId(p, iObjId));
        fmt.Printf("Supp = %5d  ", Gia_ManSuppSize(p, &iObjId, 1));
        fmt.Printf("Cone = %5d  ", Gia_ManConeSize(p, &iObjId, 1));
        fmt.Printf("Mffc = %5d  ", Gia_NodeMffcSize(p, Gia_ObjFanin0(pObj)));
        fmt.Printf("Prob = %8.4f  ", Vec_FltEntry(vProb, iObjId));
        fmt.Printf("\n");
    }
    Vec_FltFree( vProb );
}

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
func Gia_ManPrintNpnClasses(p *Gia_Man_t) {
    extern char ** Kit_DsdNpn4ClassNames();
    var pNames *string  = Kit_DsdNpn4ClassNames();
    var vLeaves, vTruth, vVisited *Vec_Int_t;
    var pLutClass *int32;
    var ClassCounts [222]int32 ;
    var i, k, iFan, Class, OtherClasses, OtherClasses2, nTotal, Counter, Counter2 int32;
    var pTruth *uint32;
    assert(Gia_ManHasMapping(p), "Gia_ManHasMapping(p)");
    assert(Gia_ManLutSizeMax( p ) <= 4, "Gia_ManLutSizeMax( p ) <= 4");
    vLeaves   = Vec_IntAlloc( 100 );
    vVisited  = Vec_IntAlloc( 100 );
    vTruth    = Vec_IntAlloc( (1<<16) );
    pLutClass = ABC_CALLOC( int, Gia_ManObjNum(p) );
    Gia_ManCleanTruth( p );
    Gia_ManForEachLut( p, i )
    {
        if Gia_ObjLutSize(p,i) > 4 {
            continue; }
        Vec_IntClear( vLeaves );
        Gia_LutForEachFanin( p, i, iFan, k )
            Vec_IntPush( vLeaves, iFan );
        for  ; k < 4; k++  {
            Vec_IntPush( vLeaves, 0 );
        }
        pTruth = Gia_ManConvertAigToTruth( p, Gia_ManObj(p, i), vLeaves, vTruth, vVisited );
        Class = Dar_LibReturnClass( *pTruth );
        ClassCounts[ Class ]++;
        pLutClass[i] = Class;
    }
    Vec_IntFree( vLeaves );
    Vec_IntFree( vTruth );
    Vec_IntFree( vVisited );
    Vec_IntFreeP( &p.vTruths );
    nTotal = 0;
    for  i = 0; i < 222; i++  {
        nTotal += ClassCounts[i];
    }
    Abc_Print( 1, "NPN CLASS STATISTICS (for %d LUT4 present in the current mapping):\n", nTotal );
    OtherClasses = 0;
    for  i = k = 0; i < 222; i++  {
        if ClassCounts[i] == 0 {
            continue; }
//        if ( 100.0 * ClassCounts[i] / (nTotal+1) < 0.1 ) // do not show anything below 0.1 percent
//            continue;
        OtherClasses += ClassCounts[i];
        Abc_Print( 1, "%3d: Class %3d :  Count = %6d   (%7.2f %%)   %s\n", 
            ++k, i, ClassCounts[i], 100.0 * ClassCounts[i] / (nTotal+1), pNames[i] );
    }
    OtherClasses = nTotal - OtherClasses;
    Abc_Print( 1, "Other     :  Count = %6d   (%7.2f %%)\n", 
        OtherClasses, 100.0 * OtherClasses / (nTotal+1) );
    // count the number of LUTs that have MUX function and two fanins with MUX functions
    OtherClasses = OtherClasses2 = 0;
    p.Refs = nil;
    Gia_ManSetRefsMapped( p );
    Gia_ManForEachLut( p, i )
    {
        if pLutClass[i] != 109 {
            continue; }
        Counter = Counter2 = 0;
        Gia_LutForEachFanin( p, i, iFan, k )
        {
            Counter  += (pLutClass[iFan] == 109);
            Counter2 += (pLutClass[iFan] == 109) && (Gia_ObjRefNumId(p, iFan) == 1);
        }
        OtherClasses  += (Counter > 1);
        OtherClasses2 += (Counter2 > 1);
//            Abc_Print( 1, "%d -- ", pLutClass[i] );
//            Gia_LutForEachFanin( p, i, iFan, k )
//                Abc_Print( 1, "%d ", pLutClass[iFan] );
//            Abc_Print( 1, "\n" );
    }
    p.Refs = nil;
    Abc_Print( 1, "Approximate number of 4:1 MUX structures: All = %6d  (%7.2f %%)  MFFC = %6d  (%7.2f %%)\n", 
        OtherClasses,  100.0 * OtherClasses  / (nTotal+1),
        OtherClasses2, 100.0 * OtherClasses2 / (nTotal+1) );
    pLutClass = nil;
}


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
