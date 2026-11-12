package nwk

import (
	"fmt"
	"os"
	"strings"
)

/**CFile****************************************************************

  FileName    [nwkMan.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Logic network representation.]
 
  Synopsis    [Network manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: nwkMan.c,v 1.1 2008/10/10 14:09:30 mjarvin Exp $]

***********************************************************************/

//#include "nwk.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Allocates the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManAlloc() *Nwk_Man_t {
    var p *Nwk_Man_t;
    // inlining of macro ABC_ALLOC
    p = ((Nwk_Man_t *) malloc(sizeof(Nwk_Man_t) * (size_t)(1)));
    memset( p, 0, sizeof(Nwk_Man_t) );
    p.vCis = Vec_PtrAlloc( 1000 );
    p.vCos = Vec_PtrAlloc( 1000 );
    p.vObjs = Vec_PtrAlloc( 1000 );
    p.vTemp = Vec_PtrAlloc( 1000 );
    p.nFanioPlus = 2;
    p.pMemObjs = Aig_MmFlexStart();
    p.pManHop = Hop_ManStart();
    return p;
}

/**Function*************************************************************

  Synopsis    [Deallocates the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManFree(p *Nwk_Man_t) {
//    printf( "The number of realloced nodes = %d.\n", p->nRealloced );
    // inlining of macro ABC_FREE
    if p.pName {    ((p.pName) ? (free((char *) (p.pName)), (p.pName) = 0) : 0); }
    // inlining of macro ABC_FREE
    if p.pSpec {    ((p.pSpec) ? (free((char *) (p.pSpec)), (p.pSpec) = 0) : 0); }
    if p.vCis {     Vec_PtrFree( p.vCis ); }
    if p.vCos {     Vec_PtrFree( p.vCos ); }
    if p.vObjs {    Vec_PtrFree( p.vObjs ); }
    if p.vTemp {    Vec_PtrFree( p.vTemp ); }
    if p.pManTime { Tim_ManStop( p.pManTime ); }
    if p.pMemObjs { Aig_MmFlexStop( p.pMemObjs, 0 ); }
    if p.pManHop {  Hop_ManStop( p.pManHop ); }
    // inlining of macro ABC_FREE
    ((p) ? (free((char *) (p)), (p) = 0) : 0);
}

/**Function*************************************************************

  Synopsis    [Prints stats of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPrintLutSizes(p *Nwk_Man_t, pLutLib *If_LibLut_t) {
    var pObj *Nwk_Obj_t;
    var i int32;
    var Counters [256]int32 ;
    // inlining of macro Nwk_ManForEachNode
    for  i = 0; (i < Vec_PtrSize(p.vObjs)) && (((pObj) = (Nwk_Obj_t *)Vec_PtrEntry(p.vObjs, i)), 1); i++  {         if (pObj) == nil || !Nwk_ObjIsNode(pObj) {} else {
        Counters[Nwk_ObjFaninNum(pObj)]++; }
    }
    fmt.Printf("LUTs by size: ");
    for  i = 0; i <= pLutLib.LutMax; i++  {
        fmt.Printf("%d:%d ", i, Counters[i]);
    }
}

/**Function*************************************************************

  Synopsis    [If the network is best, saves it in "best.blif" and returns 1.]

  Description [If the networks are incomparable, saves the new network, 
  returns its parameters in the internal parameter structure, and returns 1.
  If the new network is not a logic network, quits without saving and returns 0.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManCompareAndSaveBest(pNtk *Nwk_Man_t, pNtl any) int32 {
//    extern void Ntl_WriteBlifLogic( Nwk_Man_t * pNtk, void * pNtl, char * pFileName );
    extern void Nwk_ManDumpBlif( Nwk_Man_t * pNtk, char * pFileName, Vec_Ptr_t * vPiNames, Vec_Ptr_t * vPoNames );
    static struct ParStruct {
        var pName string;  // name of the best saved network
        var Depth int32;  // depth of the best saved network
        var Flops int32;  // flops in the best saved network 
        var Nodes int32;  // nodes in the best saved network
        var nPis int32;   // the number of primary inputs
        var nPos int32;   // the number of primary outputs
    }
 ParsNew, ParsBest = { 0 };
    // free storage for the name
    if pNtk == nil {
        // inlining of macro ABC_FREE
        ((ParsBest.pName) ? (free((char *) (ParsBest.pName)), (ParsBest.pName) = 0) : 0);
        return 0;
    }
    // get the parameters
    ParsNew.Depth = Nwk_ManLevel( pNtk );
    ParsNew.Flops = Nwk_ManLatchNum( pNtk );
    ParsNew.Nodes = Nwk_ManNodeNum( pNtk );
    ParsNew.nPis  = Nwk_ManPiNum( pNtk );
    ParsNew.nPos  = Nwk_ManPoNum( pNtk );
    // reset the parameters if the network has the same name
    if ParsBest.pName == nil ||
          (int32(strings.Compare(ParsBest.pName, pNtk.pName)) != 0) ||
          ParsBest.Depth >  ParsNew.Depth ||
         (ParsBest.Depth == ParsNew.Depth && ParsBest.Flops >  ParsNew.Flops) ||
         (ParsBest.Depth == ParsNew.Depth && ParsBest.Flops == ParsNew.Flops && ParsBest.Nodes >  ParsNew.Nodes) {
        // inlining of macro ABC_FREE
        ((ParsBest.pName) ? (free((char *) (ParsBest.pName)), (ParsBest.pName) = 0) : 0);
        ParsBest.pName = Abc_UtilStrsav( pNtk.pName );
        ParsBest.Depth = ParsNew.Depth; 
        ParsBest.Flops = ParsNew.Flops; 
        ParsBest.Nodes = ParsNew.Nodes; 
        ParsBest.nPis  = ParsNew.nPis; 
        ParsBest.nPos  = ParsNew.nPos;
        // write the network
//        Ntl_WriteBlifLogic( pNtk, pNtl, "best.blif" );
//        Nwk_ManDumpBlif( pNtk, "best_map.blif", NULL, NULL );
        return 1;
    }
    return 0;
}

/**Function*************************************************************

  Synopsis    []

  Description []

  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_FileNameGeneric(FileName string) string {
    var pDot, pRes string;
    pRes = Abc_UtilStrsav( FileName );
    if pDot = strrchr( pRes, '.' ) {
        *pDot = 0; }
    return pRes;
}

/**Function*************************************************************

  Synopsis    [Marks nodes for power-optimization.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwl_ManComputeTotalSwitching(pNtk *Nwk_Man_t) float32 {
    extern Vec_Int_t * Saig_ManComputeSwitchProbs( Aig_Man_t * p, int nFrames, int nPref, int fProbOne );
    var vSwitching *Vec_Int_t;
    var pSwitching *float32;
    var pAig *Aig_Man_t;
    var pObjAig *Aig_Obj_t;
    var pObjAbc *Nwk_Obj_t;
    var Result float32  = float32(0);
    var i int32;
    // strash the network
    // map network into an AIG
    pAig = Nwk_ManStrash( pNtk );
    vSwitching = Saig_ManComputeSwitchProbs( pAig, 48, 16, 0 );
    pSwitching = (float *)vSwitching.pArray;
    // inlining of macro Nwk_ManForEachObj
    for  i = 0; (i < Vec_PtrSize(pNtk.vObjs)) && (((pObjAbc) = (Nwk_Obj_t *)Vec_PtrEntry(pNtk.vObjs, i)), 1); i++  {         if pObjAbc == nil {} else {
        if pObjAig = Aig_Regular((Aig_Obj_t *)pObjAbc.pCopy) {
            Result += Nwk_ObjFanoutNum(pObjAbc) * pSwitching[pObjAig.Id]; }
    }
    }
    Vec_IntFree( vSwitching );
    Aig_ManStop( pAig );
    return Result;
}
 
/**Function*************************************************************

  Synopsis    [Prints stats of the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Nwk_ManPrintStats(pNtk *Nwk_Man_t, pLutLib *If_LibLut_t, fSaveBest int32, fDumpResult int32, fPower int32, pNtl *Ntl_Man_t) {
//    extern int Ntl_ManLatchNum( Ntl_Man_t * p );
//    extern void Ntl_ManWriteBlifLogic( Nwk_Man_t * pNtk, void * pNtl, char * pFileName );
    if fSaveBest != 0 {
        Nwk_ManCompareAndSaveBest( pNtk, pNtl ); }
    if fDumpResult != 0 {
        var Buffer string // was [1000]byte in C

        char * pNameGen = pNtk.pSpec? Nwk_FileNameGeneric( pNtk.pSpec ) : "nameless_";
        Buffer = fmt.Sprintf("%s_dump.blif", pNameGen);
//        Ntl_ManWriteBlifLogic( pNtk, pNtl, Buffer );
//        sprintf( Buffer, "%s_dump_map.blif", pNameGen );
//        Nwk_ManDumpBlif( pNtk, Buffer, NULL, NULL );
        // inlining of macro ABC_FREE
        if pNtk.pSpec { ((pNameGen) ? (free((char *) (pNameGen)), (pNameGen) = 0) : 0); }
    }

    pNtk.pLutLib = pLutLib;
    fmt.Printf("%-15s : ", pNtk.pName);
    fmt.Printf("pi = %5d  ", Nwk_ManPiNum(pNtk));
    fmt.Printf("po = %5d  ", Nwk_ManPoNum(pNtk));
    fmt.Printf("ci = %5d  ", Nwk_ManCiNum(pNtk));
    fmt.Printf("co = %5d  ", Nwk_ManCoNum(pNtk));
//    printf( "lat = %5d  ",   Ntl_ManLatchNum(pNtl) );
    fmt.Printf("node = %5d  ", Nwk_ManNodeNum(pNtk));
    fmt.Printf("edge = %5d  ", Nwk_ManGetTotalFanins(pNtk));
    fmt.Printf("aig = %6d  ", Nwk_ManGetAigNodeNum(pNtk));
    fmt.Printf("lev = %3d  ", Nwk_ManLevel(pNtk));
//    printf( "lev2 = %3d  ",  Nwk_ManLevelBackup(pNtk) );
    fmt.Printf("delay = %5.2f  ", Nwk_ManDelayTraceLut(pNtk));
    if fPower != 0 {
        fmt.Printf("power = %7.2f   ", Nwl_ManComputeTotalSwitching(pNtk)); }
    Nwk_ManPrintLutSizes( pNtk, pLutLib );
    fmt.Printf("\n");
//    Nwk_ManDelayTracePrint( pNtk, pLutLib );
    fflush( os.Stdout );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

