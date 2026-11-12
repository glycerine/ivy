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
func Gia_ManMemory(p *Gia_Man_t) float64 {
    var Memory float64  = sizeof(Gia_Man_t);
    Memory += sizeof(Gia_Obj_t) * Gia_ManObjNum(p);
    Memory += sizeof(int) * Gia_ManCiNum(p);
    Memory += sizeof(int) * Gia_ManCoNum(p);
    Memory += sizeof(int) * Vec_IntSize(&p.vHTable);
    Memory += sizeof(int) * Gia_ManObjNum(p) * (p.Refs != nil);
    Memory += Vec_IntMemory( p.vLevels );
    Memory += Vec_IntMemory( p.vCellMapping );
    Memory += Vec_IntMemory( &p.vCopies );
    Memory += Vec_FltMemory( p.vInArrs );
    Memory += Vec_FltMemory( p.vOutReqs );
    Memory += Vec_PtrMemory( p.vNamesIn );
    Memory += Vec_PtrMemory( p.vNamesOut );
    return Memory;
}

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
func Gia_ManPrintClasses_old(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i int32;
    if p.vFlopClasses == nil {
        return; }
    Gia_ManForEachRo( p, pObj, i )
        Abc_Print( 1, "%d", Vec_IntEntry(p.vFlopClasses, i) );
    Abc_Print( 1, "\n" );

    {
        var pTemp *Gia_Man_t;
        pTemp = Gia_ManDupFlopClass( p, 1 );
        Gia_AigerWrite( pTemp, "dom1.aig", 0, 0, 0 );
        Gia_ManStop( pTemp );
        pTemp = Gia_ManDupFlopClass( p, 2 );
        Gia_AigerWrite( pTemp, "dom2.aig", 0, 0, 0 );
        Gia_ManStop( pTemp );
    }
}

/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintPlacement(p *Gia_Man_t) {
    var i, nFixed, nUndef int32;
    if p.pPlacement == nil {
        return; }
    for  i = 0; i < Gia_ManObjNum(p); i++  {
        nFixed += p.pPlacement[i].fFixed;
        nUndef += p.pPlacement[i].fUndef;
    }
    Abc_Print( 1, "Placement:  Objects = %8d.  Fixed = %8d.  Undef = %8d.\n", Gia_ManObjNum(p), nFixed, nUndef );
}


/**Function*************************************************************

  Synopsis    [Duplicates AIG for unrolling.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintTents_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vObjs *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    Vec_IntPush( vObjs, Gia_ObjId(p, pObj) );
    if Gia_ObjIsCi(pObj) {
        return; }
    Gia_ManPrintTents_rec( p, Gia_ObjFanin0(pObj), vObjs );
    if Gia_ObjIsAnd(pObj) {
        Gia_ManPrintTents_rec( p, Gia_ObjFanin1(pObj), vObjs ); }
}
func Gia_ManPrintTents(p *Gia_Man_t) {
    var vObjs *Vec_Int_t;
    var pObj *Gia_Obj_t;
    var t, i, iObjId, nSizePrev, nSizeCurr int32;
    assert(Gia_ManPoNum(p) > 0, "Gia_ManPoNum(p) > 0");
    vObjs = Vec_IntAlloc( 100 );
    // save constant class
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    Vec_IntPush( vObjs, 0 );
    // create starting root
    nSizePrev = Vec_IntSize(vObjs);
    Gia_ManForEachPo( p, pObj, i )
        Gia_ManPrintTents_rec( p, pObj, vObjs );
    // build tents
    Abc_Print( 1, "Tents:  " );
    for  t = 1; nSizePrev < Vec_IntSize(vObjs); t++  {
        var nPis int32;
        nSizeCurr = Vec_IntSize(vObjs);
        Vec_IntForEachEntryStartStop( vObjs, iObjId, i, nSizePrev, nSizeCurr )
        {
            nPis += Gia_ObjIsPi(p, Gia_ManObj(p, iObjId));
            if Gia_ObjIsRo(p, Gia_ManObj(p, iObjId)) {
                Gia_ManPrintTents_rec( p, Gia_ObjRoToRi(p, Gia_ManObj(p, iObjId)), vObjs ); }
        }
        Abc_Print( 1, "%d=%d(%d)  ", t, nSizeCurr - nSizePrev, nPis );
        nSizePrev = nSizeCurr;
    }
    Abc_Print( 1, " Unused=%d\n", Gia_ManObjNum(p) - Vec_IntSize(vObjs) );
    Vec_IntFree( vObjs );
    // the remaining objects are PIs without fanout
//    Gia_ManForEachObj( p, pObj, i )
//        if ( !Gia_ObjIsTravIdCurrent(p, pObj) )
//            Gia_ObjPrint( p, pObj );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintInitClasses(vInits *Vec_Int_t) {
    var i, Value int32;
    var Counts [6]int32 ;
    Vec_IntForEachEntry( vInits, Value, i )
        Counts[Value]++;
    for  i = 0; i < 6; i++  {
        if Counts[i] {
            fmt.Printf("%d = %d  ", i, Counts[i]); }
    }
    fmt.Printf("  ");
    fmt.Printf("B = %d  ", Counts[0] + Counts[1]);
    fmt.Printf("X = %d  ", Counts[2] + Counts[3]);
    fmt.Printf("Q = %d\n", Counts[4] + Counts[5]);
    Vec_IntForEachEntry( vInits, Value, i )
    {
        Counts[Value]++;
        if Value == 0 {
            fmt.Printf("0");
        } else if Value == 1 {
            fmt.Printf("1");
        } else if Value == 2 {
            fmt.Printf("2");
        } else if Value == 3 {
            fmt.Printf("3");
        } else if Value == 4 {
            fmt.Printf("4");
        } else if Value == 5 {
            fmt.Printf("5");
        } else { assert(false, "0"); }
    }
    fmt.Printf("\n");
    
}

/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintChoiceStats(p *Gia_Man_t) {
    var pObj *Gia_Obj_t;
    var i, nEquivs, nChoices int32;
    Gia_ManMarkFanoutDrivers( p );
    Gia_ManForEachAnd( p, pObj, i )
    {
        if !Gia_ObjSibl(p, i) {
            continue; }
        nEquivs++;
        if pObj.fMark0 {
            nChoices++; }
        assert(!Gia_ObjSiblObj(p, i).fMark0, "!Gia_ObjSiblObj(p, i).fMark0");
        assert(Gia_ObjIsAnd(Gia_ObjSiblObj(p, i)), "Gia_ObjIsAnd(Gia_ObjSiblObj(p, i))");
    }
    Abc_Print( 1, "Choice stats: Equivs =%7d. Choices =%7d.\n", nEquivs, nChoices );
    Gia_ManCleanMark0( p );
}


/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintEdges(p *Gia_Man_t) int32 {
    fmt.Printf("Edges (Q=2)    :                ");
    fmt.Printf("edge =%8d  ", (Vec_IntCountPositive(p.vEdge1) + Vec_IntCountPositive(p.vEdge2))/2);
    fmt.Printf("lev =%5.1f", 0.1*Gia_ManEvalEdgeDelay(p));
    fmt.Printf("\n");
    return 0;
}

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
func Gia_ManLogAigStats(p *Gia_Man_t, pDumpFile string) {
    var pTable *os.File  = os.OpenFile(pDumpFile, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    fmt.Fprintf(pTable, "{\n");
    fmt.Fprintf(pTable, "    \"name\" : \"%s\",\n", p.pName);
    fmt.Fprintf(pTable, "    \"input\" : %d,\n", Gia_ManCiNum(p));
    fmt.Fprintf(pTable, "    \"output\" : %d,\n", Gia_ManCoNum(p));
    fmt.Fprintf(pTable, "    \"and\" : %d,\n", Gia_ManAndNum(p));
    fmt.Fprintf(pTable, "    \"level\" : %d\n", Gia_ManLevelNum(p));
    fmt.Fprintf(pTable, "}\n");
    _ = (pTable).Close();
}

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
func Gia_ManPrintStatsShort(p *Gia_Man_t) {
    Abc_Print( 1, "i/o =%7d/%7d  ", Gia_ManPiNum(p), Gia_ManPoNum(p) );
    Abc_Print( 1, "ff =%7d  ", Gia_ManRegNum(p) );
    Abc_Print( 1, "and =%8d  ", Gia_ManAndNum(p) );
    Abc_Print( 1, "lev =%5d  ", Gia_ManLevelNum(p) );
//    Abc_Print( 1, "mem =%5.2f MB", 12.0*Gia_ManObjNum(p)/(1<<20) );
    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    [Prints stats for the AIG.]

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManPrintMiterStatus(p *Gia_Man_t) {
    var pObj, pChild *Gia_Obj_t;
    var i int32;
    var nSat int32;
    var nUnsat int32;
    var nUndec int32;
    var iOut int32 = -1;
    Gia_ManForEachPo( p, pObj, i )
    {
        pChild = Gia_ObjChild0(pObj);
        // check if the output is constant 0
        if pChild == Gia_ManConst0(p) {
            nUnsat++;
        // check if the output is constant 1
        } else if pChild == Gia_ManConst1(p) {
            nSat++;
            if iOut == -1 {
                iOut = i; }
        
        // check if the output is a primary input
        } else if Gia_ObjIsPi(p, Gia_Regular(pChild)) {
            nSat++;
            if iOut == -1 {
                iOut = i; }
        
/*
        // check if the output is 1 for the 0000 pattern
        else if ( Gia_Regular(pChild)->fPhase != (unsigned)Gia_IsComplement(pChild) )
        {
            nSat++;
            if ( iOut == -1 )
                iOut = i;
        }
*/
        } else {
            nUndec++; }
    }
    Abc_Print( 1, "Outputs = %7d.  Unsat = %7d.  Sat = %7d.  Undec = %7d.\n",
        Gia_ManPoNum(p), nUnsat, nSat, nUndec );
}

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
func Gia_ManReportImprovement(p *Gia_Man_t, pNew *Gia_Man_t) {
    Abc_Print( 1, "REG: Beg = %5d. End = %5d. (R =%5.1f %%)  ",
        Gia_ManRegNum(p), Gia_ManRegNum(pNew),
        Gia_ManRegNum(p)? 100.0*(Gia_ManRegNum(p)-Gia_ManRegNum(pNew))/Gia_ManRegNum(p) : 0.0 );
    Abc_Print( 1, "AND: Beg = %6d. End = %6d. (R =%5.1f %%)",
        Gia_ManAndNum(p), Gia_ManAndNum(pNew),
        Gia_ManAndNum(p)? 100.0*(Gia_ManAndNum(p)-Gia_ManAndNum(pNew))/Gia_ManAndNum(p) : 0.0 );
    Abc_Print( 1, "\n" );
}

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
func Gia_ManDfsCollect_rec(p *Gia_Man_t, pObj *Gia_Obj_t, vObjs *Vec_Int_t) {
    if Gia_ObjIsTravIdCurrent( p, pObj ) {
        return; }
    Gia_ObjSetTravIdCurrent( p, pObj );
    if Gia_ObjIsCi(pObj) {
        var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
        if pManTime != nil {
            var i, iFirst, nTerms, iBox int32;
            iBox = Tim_ManBoxForCi( pManTime, Gia_ObjCioId(pObj) );
            if iBox >= 0 { // pObj is a box input
            
                // mark box outputs
                iFirst = Tim_ManBoxOutputFirst( pManTime, iBox );
                nTerms = Tim_ManBoxOutputNum( pManTime, iBox );
                for  i = 0; i < nTerms; i++  {
                    pObj = Gia_ManCi( p, iFirst + i );
                    Gia_ObjSetTravIdCurrent( p, pObj );
                }
                // traverse box inputs
                iFirst = Tim_ManBoxInputFirst( pManTime, iBox );
                nTerms = Tim_ManBoxInputNum( pManTime, iBox );
                for  i = 0; i < nTerms; i++  {
                    pObj = Gia_ManCo( p, iFirst + i );
                    Gia_ManDfsCollect_rec( p, pObj, vObjs );
                }
                // save the box
                Vec_IntPush( vObjs, -iBox-1 );
            }
        }
        return;
    
    } else if Gia_ObjIsCo(pObj) {
        Gia_ManDfsCollect_rec( p, Gia_ObjFanin0(pObj), vObjs );
    
    } else if Gia_ObjIsAnd(pObj) { 
        var iFan int32;
        var k int32;
        var iObj int32 = Gia_ObjId(p, pObj);
        if Gia_ManHasMapping(p) {
            assert(Gia_ObjIsLut(p, iObj), "Gia_ObjIsLut(p, iObj)");
            Gia_LutForEachFanin( p, iObj, iFan, k )
                Gia_ManDfsCollect_rec( p, Gia_ManObj(p, iFan), vObjs );
        
        } else {
            Gia_ManDfsCollect_rec( p, Gia_ObjFanin0(pObj), vObjs );
            Gia_ManDfsCollect_rec( p, Gia_ObjFanin1(pObj), vObjs );
        }
        // save the object
        Vec_IntPush( vObjs, iObj );
    
    } else if !Gia_ObjIsConst0(pObj) {
        assert(false, "0"); }
}
func Gia_ManDfsCollect(p *Gia_Man_t) *Vec_Int_t {
    var vObjs *Vec_Int_t  = Vec_IntAlloc( Gia_ManObjNum(p) );
    var pObj *Gia_Obj_t; var i int32;
    Gia_ManIncrementTravId( p );
    Gia_ManForEachCo( p, pObj, i )
        Gia_ManDfsCollect_rec( p, pObj, vObjs );
    Gia_ManForEachCi( p, pObj, i )
        Gia_ManDfsCollect_rec( p, pObj, vObjs );
    return vObjs;
} 

/**Function*************************************************************

  Synopsis    [Compute arrival/required times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDfsArrivals(p *Gia_Man_t, vObjs *Vec_Int_t) *Vec_Int_t {
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    var vTimes *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var pObj *Gia_Obj_t; var j, Entry, k, iFan int32;
    Vec_IntWriteEntry( vTimes, 0, 0 );
    if pManTime != nil {
        Tim_ManIncrementTravId( pManTime );
        Gia_ManForEachCi( p, pObj, j )
            if j < Tim_ManPiNum(pManTime) {
                var arrTime float32  = Tim_ManGetCiArrival( pManTime, j );
                Vec_IntWriteEntry( vTimes, Gia_ObjId(p, pObj), int32(arrTime) );
            }
    
    } else {
        Gia_ManForEachCi( p, pObj, j )
            Vec_IntWriteEntry( vTimes, Gia_ObjId(p, pObj), 0 );
    }
    Vec_IntForEachEntry( vObjs, Entry, j )
    {
        if Entry < 0 { // box
        
            var Time0 int32;
            var iFirst int32;
            var nTerms int32;
            var iBox int32 = -Entry-1;
            assert(iBox >= 0, "iBox >= 0");
            // set arrivals for box inputs
            iFirst = Tim_ManBoxInputFirst( pManTime, iBox );
            nTerms = Tim_ManBoxInputNum( pManTime, iBox );
            for  k = 0; k < nTerms; k++  {
                pObj  = Gia_ManCo( p, iFirst + k );
                Time0 = Vec_IntEntry( vTimes, Gia_ObjFaninId0p(p, pObj) );
                assert(Time0 >= 0, "Time0 >= 0");
                Tim_ManSetCoArrival( pManTime, Gia_ObjCioId(pObj), Time0 );
            }
            // derive arrivals for box outputs
            iFirst = Tim_ManBoxOutputFirst( pManTime, iBox );
            nTerms = Tim_ManBoxOutputNum( pManTime, iBox );
            for  k = 0; k < nTerms; k++  {
                pObj  = Gia_ManCi( p, iFirst + k );
                Time0 = Tim_ManGetCiArrival( pManTime, Gia_ObjCioId(pObj) );
                assert(Time0 >= 0, "Time0 >= 0");
                Vec_IntWriteEntry( vTimes, Gia_ObjId(p, pObj), Time0 );
            }
        
        } else if Entry > 0 { // node
        
            var Time0, Time1, TimeMax int32;
            if Gia_ManHasMapping(p) {
                assert(Gia_ObjIsLut(p, Entry), "Gia_ObjIsLut(p, Entry)");
                Gia_LutForEachFanin( p, Entry, iFan, k )
                {
                    Time0 = Vec_IntEntry( vTimes, iFan );
                    assert(Time0 >= 0, "Time0 >= 0");
                    TimeMax = Abc_MaxInt( TimeMax, Time0 );
                }
            
            } else {
                pObj  = Gia_ManObj( p, Entry );
                Time0 = Vec_IntEntry( vTimes, Gia_ObjFaninId0(pObj, Entry) );
                Time1 = Vec_IntEntry( vTimes, Gia_ObjFaninId1(pObj, Entry) );
                assert(Time0 >= 0 && Time1 >= 0, "Time0 >= 0 && Time1 >= 0");
                TimeMax = Abc_MaxInt( Time0, Time1 );
            }
            Vec_IntWriteEntry( vTimes, Entry, TimeMax + 10 );
        
        } else { assert(false, "0"); }
    }
    return vTimes;
}
func Gia_ManDfsUpdateRequired(vTimes *Vec_Int_t, iObj int32, Req int32) {
    var pTime *int32  = Vec_IntEntryP( vTimes, iObj );
    if *pTime == -1 || *pTime > Req {
        *pTime = Req; }
}
func Gia_ManDfsRequireds(p *Gia_Man_t, vObjs *Vec_Int_t, ReqTime int32) *Vec_Int_t {
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    var vTimes *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var pObj *Gia_Obj_t; 
    var j, Entry, k, iFan, Req int32;
    Vec_IntWriteEntry( vTimes, 0, 0 );
    if pManTime != nil {
        var nCoLimit int32  = Gia_ManCoNum(p) - Tim_ManPoNum(pManTime);
        Tim_ManIncrementTravId( pManTime );
        //Tim_ManInitPoRequiredAll( pManTime, (float)ReqTime );
        Gia_ManForEachCo( p, pObj, j )
            if j >= nCoLimit {
                Tim_ManSetCoRequired( pManTime, j, ReqTime );
                Gia_ManDfsUpdateRequired( vTimes, Gia_ObjFaninId0p(p, pObj), ReqTime );
            }
    
    } else {
        Gia_ManForEachCo( p, pObj, j )
            Gia_ManDfsUpdateRequired( vTimes, Gia_ObjFaninId0p(p, pObj), ReqTime );
    }
    Vec_IntForEachEntryReverse( vObjs, Entry, j )
    {
        if Entry < 0 { // box
        
            var iFirst int32;
            var nTerms int32;
            var iBox int32 = -Entry-1;
            assert(iBox >= 0, "iBox >= 0");
            // set requireds for box outputs
            iFirst = Tim_ManBoxOutputFirst( pManTime, iBox );
            nTerms = Tim_ManBoxOutputNum( pManTime, iBox );
            for  k = 0; k < nTerms; k++  {
                pObj = Gia_ManCi( p, iFirst + k );
                Req  = Vec_IntEntry( vTimes, Gia_ObjId(p, pObj) );
                if Req == -1 {
                    Req = ReqTime;
                } else {
                    Req = Req;
                } // dangling box output
                assert(Req >= 0, "Req >= 0");
                Tim_ManSetCiRequired( pManTime, Gia_ObjCioId(pObj), Req );
            }
            // derive requireds for box inputs
            iFirst = Tim_ManBoxInputFirst( pManTime, iBox );
            nTerms = Tim_ManBoxInputNum( pManTime, iBox );
            for  k = 0; k < nTerms; k++  {
                pObj = Gia_ManCo( p, iFirst + k );
                Req  = Tim_ManGetCoRequired( pManTime, Gia_ObjCioId(pObj) );
                assert(Req >= 0, "Req >= 0");
                Gia_ManDfsUpdateRequired( vTimes, Gia_ObjFaninId0p(p, pObj), Req );
            }
        
        } else if Entry > 0 { // node
        
            Req = Vec_IntEntry(vTimes, Entry) - 10;
            assert(Req >= 0, "Req >= 0");
            if Gia_ManHasMapping(p) {
                assert(Gia_ObjIsLut(p, Entry), "Gia_ObjIsLut(p, Entry)");
                Gia_LutForEachFanin( p, Entry, iFan, k )
                    Gia_ManDfsUpdateRequired( vTimes, iFan, Req );
            
            } else {
                pObj  = Gia_ManObj( p, Entry );
                Gia_ManDfsUpdateRequired( vTimes, Gia_ObjFaninId0(pObj, Entry), Req );
                Gia_ManDfsUpdateRequired( vTimes, Gia_ObjFaninId1(pObj, Entry), Req );
            }
        
        } else { assert(false, "0"); }
    }
    return vTimes;
}
func Gia_ManDfsSlacks(p *Gia_Man_t) *Vec_Int_t {
    var vSlack *Vec_Int_t  = Vec_IntStartFull( Gia_ManObjNum(p) );
    var vObjs *Vec_Int_t   = Gia_ManDfsCollect( p );
    if Vec_IntSize(vObjs) > 0 {
        var vArrs *Vec_Int_t   = Gia_ManDfsArrivals( p, vObjs );
        var Required int32        = Vec_IntFindMax( vArrs );
        var vReqs *Vec_Int_t   = Gia_ManDfsRequireds( p, vObjs, Required );
        var i int32;
        var Arr int32;
        var Req int32;
        var Arrivals int32 = ABC_INFINITY;
        Vec_IntForEachEntry( vReqs, Req, i )
            if ( Req != -1 )
                Arrivals = Abc_MinInt( Arrivals, Req );
        //if ( Arrivals != 0 )
        //    printf( "\nGlobal timing check has failed.\n\n" );
        //assert( Arrivals == 0 );
        Vec_IntForEachEntryTwo( vArrs, vReqs, Arr, Req, i )
        {
            if !Gia_ObjIsAnd(Gia_ManObj(p, i)) {
                continue; }
            if Gia_ManHasMapping(p) && !Gia_ObjIsLut(p, i) {
                continue; }
            assert(Arr <= Req, "Arr <= Req");
            Vec_IntWriteEntry( vSlack, i, Req - Arr );
        }
        Vec_IntFree( vArrs );
        Vec_IntFree( vReqs );
    }
    Vec_IntFree( vObjs );
    return vSlack;
}
func Gia_ManDfsSlacksPrint(p *Gia_Man_t) {
    var vCounts *Vec_Int_t;
    var vSlacks *Vec_Int_t = Gia_ManDfsSlacks( p );
    var i, Entry, nRange, nTotal int32;
    if Vec_IntSize(vSlacks) == 0 {
        fmt.Printf("Network contains no internal objects.\n");
        Vec_IntFree( vSlacks );
        return;
    }
    // compute slacks
    Vec_IntForEachEntry( vSlacks, Entry, i )
        if ( Entry != -1 )
            Vec_IntWriteEntry( vSlacks, i, Entry/10 );
    nRange = Vec_IntFindMax( vSlacks );
    // count items
    vCounts = Vec_IntStart( nRange + 1 );
    Vec_IntForEachEntry( vSlacks, Entry, i )
        if ( Entry != -1 )
            Vec_IntAddToEntry( vCounts, Entry, 1 );
    // print slack ranges
    nTotal = Vec_IntSum( vCounts );
    assert(nTotal > 0, "nTotal > 0");
    Vec_IntForEachEntry( vCounts, Entry, i )
    {
        fmt.Printf("Slack range %3d = ", i);
        fmt.Printf("[%4d, %4d)   ", 10*i, 10*(i+1));
        fmt.Printf("Nodes = %5d  ", Entry);
        fmt.Printf("(%6.2f %%) ", 100.0*Entry/nTotal);
        fmt.Printf("\n");
    }
    Vec_IntFree( vSlacks );
    Vec_IntFree( vCounts );
}


/**Function*************************************************************

  Synopsis    [Compute arrival/required times.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManGenUsed(p *Gia_Man_t, fBuf int32) *Vec_Bit_t {
    var pObj *Gia_Obj_t; var i int32;
    var vUsed *Vec_Bit_t  = Vec_BitStart( Gia_ManObjNum(p) );
    Gia_ManForEachAnd( p, pObj, i )
    {
        if fBuf != 0 {
            Vec_BitWriteEntry( vUsed, i, 1 ); }
        if Gia_ObjFaninC0(pObj) ^ fBuf {
            Vec_BitWriteEntry( vUsed, Gia_ObjFaninId0(pObj, i), 1 ); }
        if Gia_ObjFaninC1(pObj) ^ fBuf {
            Vec_BitWriteEntry( vUsed, Gia_ObjFaninId1(pObj, i), 1 ); }
    }
    Gia_ManForEachCo( p, pObj, i )
        if ( Gia_ObjFaninC0(pObj) ^ fBuf )
            Vec_BitWriteEntry( vUsed, Gia_ObjFaninId0p(p, pObj), 1 );
    Vec_BitWriteEntry( vUsed, 0, 0 ); // clean zero
    return vUsed;
}
func Gia_ManNameIsLegalInVerilog(pName string) int32 {
    // identifier ::= simple_identifier | escaped_identifier
    // simple_identifier ::= [a-zA-Z_][a-zA-Z0-9_$]
    // escaped_identifier ::= \ {Any_ASCII_character_except_white_space} white_space
    // white_space ::= space | tab | newline
    assert(pName != nil && *pName != '\0', "pName != nil && *pName != '\\0'");
    if *pName == '\\' {
        return 1; }
    if (*pName < 'a' || *pName > 'z') && (*pName < 'A' || *pName > 'Z') && *pName != '_' {
        return 0; }
    for  *(++pName)  {
        if (*pName < 'a' || *pName > 'z') && (*pName < 'A' || *pName > 'Z') && (*pName < '0' || *pName > '9') && *pName != '_' && *pName != '$' { 
            return 0; }
    }
    return 1;
}
var spStatic_Gia_ObjGetDumpName_pBuffer string // was [10000]byte in C

func Gia_ObjGetDumpName(vNames *Vec_Ptr_t, c byte, i int32, d int32) string {
    
    if vNames != nil {
        var pName string  = (char *)Vec_PtrEntry(vNames, i);
        if Gia_ManNameIsLegalInVerilog(pName) != 0 {
            spStatic_Gia_ObjGetDumpName_pBuffer = fmt.Sprintf("%s", pName);
        } else {
            spStatic_Gia_ObjGetDumpName_pBuffer = fmt.Sprintf("\\%s ", pName); }
    
    } else {
        spStatic_Gia_ObjGetDumpName_pBuffer = fmt.Sprintf("%c%0*d%c", c, d, i, c); }
    return spStatic_Gia_ObjGetDumpName_pBuffer;
}
func Gia_ManWriteNames(pFile *os.File, c byte, n int32, vNames *Vec_Ptr_t, Start int32, Skip int32, vObjs *Vec_Bit_t) {
    var Digits int32  = Abc_Base10Log( n );
    var Length int32 = Start;
    var i int32;
    var fFirst int32 = 1; 
    var pName string;
    for  i = 0; i < n; i++  {
        if vObjs && !Vec_BitEntry(vObjs, i) {
            continue; }
        pName = Gia_ObjGetDumpName( vNames, c, i, Digits );
        Length += len(pName) + 2;
        if Length > 60 {
            fmt.Fprintf(pFile, ",\n    ");
            Length = Skip;
            fFirst = 1;
        }
        fmt.Fprintf(pFile, "%s%s", fFirst ? "":", ", pName);
        fFirst = 0;
    }
}
func Gia_ManDumpVerilog(p *Gia_Man_t, pFileName string, vObjs *Vec_Int_t) {
    var pFile *os.File;
    var pObj *Gia_Obj_t;
    var vInvs, vUsed *Vec_Bit_t;
    var nDigits int32  = Abc_Base10Log( Gia_ManObjNum(p) );
    var nDigits2 int32  = Abc_Base10Log( Gia_ManPiNum(p) );
    var i, k, iObj int32;
    if Gia_ManRegNum(p) {
        fmt.Printf("Currently cannot write sequential AIG.\n");
        return;
    }
    pFile = os.OpenFile(pFileName, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0666);
    if pFile == nil {
        fmt.Printf("Cannot open output file \"%s\".\n", pFileName);
        return;
    }

    vInvs = Gia_ManGenUsed( p, 0 );
    vUsed = Gia_ManGenUsed( p, 1 );

    //fprintf( pFile, "// This Verilog file is written by ABC on %s\n\n", Extra_TimeStamp() );

    fmt.Fprintf(pFile, "module ");
    for  i = 0; i < int32(len(p.pName)); i++  {
        if isalpha(p.pName[i]) || isdigit(p.pName[i]) {
            fmt.Fprintf(pFile, "%c", p.pName[i]);
        } else {
            fmt.Fprintf(pFile, "_"); }
    }
    fmt.Fprintf(pFile, " (\n    ");
    Gia_ManWriteNames( pFile, 'x', Gia_ManPiNum(p), p.vNamesIn, 4, 4, nil );
    fmt.Fprintf(pFile, ",\n    ");

    Gia_ManWriteNames( pFile, 'z', Gia_ManPoNum(p), p.vNamesOut, 4, 4, nil );
    fmt.Fprintf(pFile, "\n  );\n\n");

    fmt.Fprintf(pFile, "  input ");
    Gia_ManWriteNames( pFile, 'x', Gia_ManPiNum(p), p.vNamesIn, 8, 4, nil );
    fmt.Fprintf(pFile, ";\n\n");

    fmt.Fprintf(pFile, "  output ");
    Gia_ManWriteNames( pFile, 'z', Gia_ManPoNum(p), p.vNamesOut, 9, 4, nil );
    fmt.Fprintf(pFile, ";\n\n");

    if Vec_BitCount(vUsed) {
        fmt.Fprintf(pFile, "  wire ");
        Gia_ManWriteNames( pFile, 'n', Gia_ManObjNum(p), nil, 7, 4, vUsed );
        fmt.Fprintf(pFile, ";\n\n");
    }

    if Vec_BitCount(vInvs) {
        fmt.Fprintf(pFile, "  wire ");
        Gia_ManWriteNames( pFile, 'i', Gia_ManObjNum(p), nil, 7, 4, vInvs );
        fmt.Fprintf(pFile, ";\n\n");
    }

    if vObjs {
        fmt.Fprintf(pFile, "  wire ");
        Vec_IntForEachEntry( vObjs, iObj, i )
            fmt.Fprintf(pFile, " t_%d%s", i, i==Vec_IntSize(vObjs)-1 ? "" : ","); 
        fmt.Fprintf(pFile, ";\n\n");
        Vec_IntForEachEntry( vObjs, iObj, i )
        {
            fmt.Fprintf(pFile, "  buf( %s,", Gia_ObjGetDumpName(nil, 'n', iObj, nDigits));
            fmt.Fprintf(pFile, " t_%d );\n", i);
        }
        fmt.Fprintf(pFile, "\n");
    }

    // input inverters
    Gia_ManForEachPi( p, pObj, i )
    {
        if Vec_BitEntry(vUsed, Gia_ObjId(p, pObj)) {
            fmt.Fprintf(pFile, "  buf( %s,", Gia_ObjGetDumpName(nil, 'n', Gia_ObjId(p, pObj), nDigits));
            fmt.Fprintf(pFile, " %s );\n", Gia_ObjGetDumpName(p.vNamesIn, 'x', i, nDigits2));
        }
        if Vec_BitEntry(vInvs, Gia_ObjId(p, pObj)) {
            fmt.Fprintf(pFile, "  not( %s,", Gia_ObjGetDumpName(nil, 'i', Gia_ObjId(p, pObj), nDigits));
            fmt.Fprintf(pFile, " %s );\n", Gia_ObjGetDumpName(p.vNamesIn, 'x', i, nDigits2));
        }
    }

    // internal nodes and their inverters
    fmt.Fprintf(pFile, "\n");
    Gia_ManForEachAnd( p, pObj, i )
    {
        var fSkip int32;
        if vObjs {
            Vec_IntForEachEntry( vObjs, iObj, k )
                if ( iObj == i )
                    break;
            if k < Vec_IntSize(vObjs) {
                fSkip = 1; }
        }
        if !(fSkip != 0) {
            fmt.Fprintf(pFile, "  and( %s,", Gia_ObjGetDumpName(nil, 'n', i, nDigits));
            fmt.Fprintf(pFile, " %s,", Gia_ObjGetDumpName(nil, byte((Gia_ObjFaninC0(pObj)? 'i':'n')), Gia_ObjFaninId0(pObj, i), nDigits));
            fmt.Fprintf(pFile, " %s );\n", Gia_ObjGetDumpName(nil, byte((Gia_ObjFaninC1(pObj)? 'i':'n')), Gia_ObjFaninId1(pObj, i), nDigits));
        }
        if Vec_BitEntry(vInvs, i) {
            fmt.Fprintf(pFile, "  not( %s,", Gia_ObjGetDumpName(nil, 'i', i, nDigits));
            fmt.Fprintf(pFile, " %s );\n", Gia_ObjGetDumpName(nil, 'n', i, nDigits));
        }
    }
    
    // output drivers
    fmt.Fprintf(pFile, "\n");
    nDigits2 = Abc_Base10Log( Gia_ManPoNum(p) );
    Gia_ManForEachPo( p, pObj, i )
    {
/*
        fprintf( pFile, "  assign %s = ", Gia_ObjGetDumpName(p->vNamesOut, 'z', i, nDigits2) );
        if ( Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) )
            fprintf( pFile, "1\'b%d;\n", Gia_ObjFaninC0(pObj) );
        else 
            fprintf( pFile, "%s;\n", Gia_ObjGetDumpName(NULL, (char)(Gia_ObjFaninC0(pObj)? 'i':'n'), Gia_ObjFaninId0p(p, pObj), nDigits) );
*/
        fmt.Fprintf(pFile, "  buf( %s, ", Gia_ObjGetDumpName(p.vNamesOut, 'z', i, nDigits2));
        if Gia_ObjIsConst0(Gia_ObjFanin0(pObj)) {
            fmt.Fprintf(pFile, "1\'b%d );\n", Gia_ObjFaninC0(pObj));
        } else { 
            fmt.Fprintf(pFile, "%s );\n", Gia_ObjGetDumpName(nil, byte((Gia_ObjFaninC0(pObj)? 'i':'n')), Gia_ObjFaninId0p(p, pObj), nDigits)); }
    }

    fmt.Fprintf(pFile, "\nendmodule\n\n");
    _ = (pFile).Close();

    Vec_BitFree( vInvs );
    Vec_BitFree( vUsed );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END
