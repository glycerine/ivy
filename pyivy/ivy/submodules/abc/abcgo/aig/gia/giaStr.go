package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaStr.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [AIG structuring.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaStr.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "misc/util/utilNam.h"
//#include "misc/vec/vecWec.h"
//#include "misc/tim/tim.h"

//ABC_NAMESPACE_IMPL_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Str_NtkManForEachObj( p, pObj )  \
//     for ( pObj = p->pObjs; Str_ObjId(p, pObj) < p->nObjs; pObj++ )

// Original macro:
// #define Str_NtkManForEachObjVec( vVec, p, pObj, i )  \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Str_NtkObj(p, Vec_IntEntry(vVec,i))); i++ )


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Logic network manipulation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/









/**Function*************************************************************

  Synopsis    [Constructs a normalized AIG without structural hashing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupMuxesNoHash(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj, pFan0, pFan1, pFanC *Gia_Obj_t;
    var i, iLit0, iLit1, fCompl int32;
    assert(p.pMuxes == nil, "p.pMuxes == nil");
    // inlining of macro ABC_FREE
    p.Refs = nil;
    Gia_ManCreateRefs( p ); 
    // discount nodes with one fanout pointed to by MUX type
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjIsMuxType(pObj) {
            continue; }
        Gia_ObjRefDec(p, Gia_ObjFanin0(pObj));
        Gia_ObjRefDec(p, Gia_ObjFanin1(pObj));
    }
    }
    // start the new manager
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName  = Abc_UtilStrsav( p.pName );
    pNew.pSpec  = Abc_UtilStrsav( p.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    Gia_ManFillValue(p);
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        pObj.Value = Gia_ManAppendCi( pNew );
    }
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if !Gia_ObjRefNumId(p, i) {
            continue; }
        if !Gia_ObjIsMuxType(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjRecognizeExor(pObj, &pFan0, &pFan1) {
            iLit0 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0));
            iLit1 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1));
            fCompl = Abc_LitIsCompl(iLit0) ^ Abc_LitIsCompl(iLit1);
            pObj.Value = fCompl ^ Gia_ManAppendXorReal( pNew, Abc_LitRegular(iLit0), Abc_LitRegular(iLit1) );
        
        } else {
            pFanC = Gia_ObjRecognizeMux( pObj, &pFan1, &pFan0 );
            iLit0 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0));
            iLit1 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1));
            if iLit0 == iLit1 {
                pObj.Value = iLit0;
            } else if Abc_Lit2Var(iLit0) == Abc_Lit2Var(iLit1) {
                iLit1 = Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFanC));
                fCompl = Abc_LitIsCompl(iLit0) ^ Abc_LitIsCompl(iLit1);
                pObj.Value = fCompl ^ Gia_ManAppendXorReal( pNew, Abc_LitRegular(iLit0), Abc_LitRegular(iLit1) );
            
            } else {
                pObj.Value = Gia_ManAppendMuxReal( pNew, Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFanC)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan1)), Gia_ObjLitCopy(p, Gia_ObjToLit(p, pFan0)) ); }
        }
    }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    assert(!Gia_ManHasDangling(pNew), "!Gia_ManHasDangling(pNew)");
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Constructs AIG ordered for balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/













/**Function*************************************************************

  Synopsis    [Simplify multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Collect multi-input AND/XOR.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/




/**Function*************************************************************

  Synopsis    [Constructs AIG ordered for balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



 
/**Function*************************************************************

  Synopsis    [Delay computation]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/






/**Function*************************************************************

  Synopsis    [Transposing 64-bit matrix.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


/**Function*************************************************************

  Synopsis    [Perform affinity computation.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





/**Function*************************************************************

  Synopsis    [Count 1s.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/


func Str_PrintState(pCost *int32, pSuper *int32, pMatrix *word, nSize int32) {
    var i int32;
    for  i = 0; i < nSize; i++  {
        fmt.Printf("%6d : ", i);
        fmt.Printf("%6d : ", Abc_Lit2Var(pSuper[i]));
        fmt.Printf("%3d   ", pCost[i] >> 4);
        fmt.Printf("%3d   ", pCost[i] & 15);
//        Extra_PrintBinary( stdout, pMatrix+i, 64 ), printf( "\n" );
    }
    fmt.Printf("\n");
}


/**Function*************************************************************

  Synopsis    [Perform balancing.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/





func Str_NtkBalanceMux(pNew *Gia_Man_t, p *Str_Ntk_t, pObj *Str_Obj_t, vDelay *Vec_Int_t, nLutSize int32, nGroups int32, nMuxes int32, fRecursive int32, fOptArea int32, fVerbose int32) {
    extern int Str_MuxRestructure( Gia_Man_t * pNew, Str_Ntk_t * pNtk, int iMux, int nMuxes, Vec_Int_t * vDelay, int nLutSize, int fRecursive, int fOptArea, int fVerbose );
    var n int32;
    var m int32;
    var iRes int32;
    var fUseRestruct int32 = 1;
    if fUseRestruct != 0 {
        for  n = 0; n < nGroups; n++  {            
            iRes = Str_MuxRestructure( pNew, p, Str_ObjId(p, pObj), nMuxes, vDelay, nLutSize, fRecursive, fOptArea, fVerbose );
            if iRes == -1 {
                for  m = 0; m < nMuxes; m, pObj = m+1, pObj+1  {
                    pObj.iCopy = Gia_ManHashMuxReal( pNew, Str_ObjFaninCopy(p, pObj, 2), Str_ObjFaninCopy(p, pObj, 1), Str_ObjFaninCopy(p, pObj, 0) );
                    Str_ObjDelay( pNew, Abc_Lit2Var(pObj.iCopy), nLutSize, vDelay );
                }
            
            } else {
                pObj += nMuxes - 1;
                pObj.iCopy = iRes;
                pObj++;
            }
        }
    
    } else {
        for  n = 0; n < nGroups * nMuxes; n, pObj = n+1, pObj+1  {
            pObj.iCopy = Gia_ManHashMuxReal( pNew, Str_ObjFaninCopy(p, pObj, 2), Str_ObjFaninCopy(p, pObj, 1), Str_ObjFaninCopy(p, pObj, 0) );
            Str_ObjDelay( pNew, Abc_Lit2Var(pObj.iCopy), nLutSize, vDelay );
        }
    }
}
func Str_NtkBalance(pGia *Gia_Man_t, p *Str_Ntk_t, nLutSize int32, fUseMuxes int32, fRecursive int32, fOptArea int32, fVerbose int32) *Gia_Man_t {
    var pNew, pTemp *Gia_Man_t;
    var vDelay *Vec_Int_t;
    var pObj *Str_Obj_t; 
    var nGroups, nMuxes, CioId int32;
    var arrTime, Delay int32;
    assert(nLutSize < 16, "nLutSize < 16");
    assert(pGia.pMuxes == nil, "pGia.pMuxes == nil");
    pNew = Gia_ManStart( Gia_ManObjNum(pGia) );
    pNew.pName = Abc_UtilStrsav( pGia.pName );
    pNew.pSpec = Abc_UtilStrsav( pGia.pSpec );
    // inlining of macro ABC_CALLOC
    pNew.pMuxes = ((unsigned *) calloc((size_t)(pNew.nObjsAlloc), sizeof(unsigned)));
    Vec_IntFill( &pNew.vCopies, pNew.nObjsAlloc, -1 );
    if pNew.vSuper == nil {
        pNew.vSuper = Vec_IntAlloc( 1000 ); }
    if pNew.vStore == nil {
        pNew.vStore = Vec_IntAlloc( 1000 ); }
    vDelay = Vec_IntStart( 2*pNew.nObjsAlloc );
    Gia_ManHashStart( pNew );
    if pGia.pManTime != nil { // Tim_Man with unit delay 16
    
        Tim_ManInitPiArrivalAll( (Tim_Man_t *)pGia.pManTime, 17 );
        Tim_ManIncrementTravId( (Tim_Man_t *)pGia.pManTime );
    }
    // inlining of macro Str_NtkManForEachObj
    for  pObj = p.pObjs; Str_ObjId(p, pObj) < p.nObjs; pObj++  {
        if pObj.Type == STR_PI {
            pObj.iCopy = Gia_ManAppendCi( pNew );
            arrTime = 17;
            if pGia.pManTime != nil {
                CioId = Gia_ObjCioId( Gia_ManObj(pNew, Abc_Lit2Var(pObj.iCopy)) );
                arrTime = int32(Tim_ManGetCiArrival( (Tim_Man_t *)pGia.pManTime, CioId ));
            }
            Vec_IntWriteEntry( vDelay, Abc_Lit2Var(pObj.iCopy), arrTime );
        
        } else if pObj.Type == STR_AND || pObj.Type == STR_XOR {
            Str_NtkBalanceMulti( pNew, p, pObj, vDelay, nLutSize );
        } else if pObj.Type == STR_MUX && pObj.iTop >= 0 && (fUseMuxes != 0) {
            Str_ObjReadGroup( p, pObj, &nGroups, &nMuxes );
            assert(nGroups * nMuxes >= 2, "nGroups * nMuxes >= 2");
            Str_NtkBalanceMux( pNew, p, pObj, vDelay, nLutSize, nGroups, nMuxes, fRecursive, fOptArea, fVerbose );
            pObj += nGroups * nMuxes - 1;
        
        } else if pObj.Type == STR_MUX {
            pObj.iCopy = Gia_ManHashMuxReal( pNew, Str_ObjFaninCopy(p, pObj, 2), Str_ObjFaninCopy(p, pObj, 1), Str_ObjFaninCopy(p, pObj, 0) );
            Str_ObjDelay( pNew, Abc_Lit2Var(pObj.iCopy), nLutSize, vDelay );
        
        } else if pObj.Type == STR_PO {
            pObj.iCopy = Gia_ManAppendCo( pNew, Str_ObjFaninCopy(p, pObj, 0) );
            arrTime = Vec_IntEntry(vDelay, Abc_Lit2Var(Str_ObjFaninCopy(p, pObj, 0)) );
            Delay = Abc_MaxInt( Delay, arrTime );
            if pGia.pManTime != nil {
                CioId = Gia_ObjCioId( Gia_ManObj(pNew, Abc_Lit2Var(pObj.iCopy)) );
                Tim_ManSetCoArrival( (Tim_Man_t *)pGia.pManTime, CioId, float32(arrTime) );
            }
        
        } else if pObj.Type == STR_CONST0 {
            pObj.iCopy = 0, Vec_IntWriteEntry(vDelay, 0, 17);
        } else { assert(false, "0"); }
    }
    if fVerbose != 0 {
        fmt.Printf("Max delay = %d.  Old objs = %d.  New objs = %d.\n", Delay >> 4, Gia_ManObjNum(pGia), Gia_ManObjNum(pNew)); }
    Vec_IntFree( vDelay );
    // inlining of macro ABC_FREE
    pNew.vCopies.pArray = nil;
    Gia_ManHashStop( pNew );
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(pGia) );
    pNew = Gia_ManDupNoMuxes( pTemp = pNew, 0 );
    Gia_ManStop( pTemp );
//    if ( pGia->pManTime != NULL )
//        pNew->pManTime = Tim_ManDup( (Tim_Man_t *)pGia->pManTime, 0 );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Test normalization procedure.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManLutBalance(p *Gia_Man_t, nLutSize int32, fUseMuxes int32, fRecursive int32, fOptArea int32, fVerbose int32) *Gia_Man_t {
    var pNtk *Str_Ntk_t;
    var pNew *Gia_Man_t;
    var clk abctime  = Abc_Clock();
    if p.pManTime && Tim_ManBoxNum((Tim_Man_t*)p.pManTime) && Gia_ManIsNormalized(p) {
        var pTimOld *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
        p.pManTime = Tim_ManDup( pTimOld, 16 );
        pNew = Gia_ManDupUnnormalize( p );
        if pNew == nil {
            return nil; }
        Gia_ManTransferTiming( pNew, p );
        p = pNew;
        // optimize
        pNtk = Str_ManNormalize( p );
        pNew = Str_NtkBalance( p, pNtk, nLutSize, fUseMuxes, fRecursive, fOptArea, fVerbose );
        Gia_ManTransferTiming( pNew, p );
        Gia_ManStop( p );
        // normalize
        pNew = Gia_ManDupNormalize( p = pNew, 0 );
        Gia_ManTransferTiming( pNew, p );
        Gia_ManStop( p );
        // cleanup
        Tim_ManStop( (Tim_Man_t *)pNew.pManTime );
        pNew.pManTime = pTimOld;
        assert(Gia_ManIsNormalized(pNew), "Gia_ManIsNormalized(pNew)");
    
    } else {
        pNtk = Str_ManNormalize( p );
    //    Str_NtkPrintGroups( pNtk );
        pNew = Str_NtkBalance( p, pNtk, nLutSize, fUseMuxes, fRecursive, fOptArea, fVerbose );
        Gia_ManTransferTiming( pNew, p );
    }
    if fVerbose != 0 {
        Str_NtkPs( pNtk, Abc_Clock() - clk ); }
    Str_NtkDelete( pNtk );   
    return pNew;
}





/**Function*************************************************************

  Synopsis    [Perform MUX restructuring.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
    
type Str_Edg_t struct {
    Fan            int32;      // fanin ID
    fCompl            int32;   // fanin complement
    FanDel            int32;   // fanin delay
    Copy            int32;     // fanin copy
}

    // 64 bytes
type Str_Mux_t struct {
    Id            int32;       // node ID
    Delay            int32;    // node delay
    Copy            int32;     // node copy
    nLutSize            int32; // LUT size
    Edge [3]Str_Edg_t;  // fanins
}

func Str_MuxFanin(pMux *Str_Mux_t, i int32) *Str_Mux_t { return pMux - pMux.Id + pMux.Edge[i].Fan;                         }
func Str_MuxHasFanin(pMux *Str_Mux_t, i int32) int32 { return pMux.Edge[i].Fan > 0 && Str_MuxFanin(pMux, i).Copy != -2;  }

func Str_MuxDelayPrint_rec(pMux *Str_Mux_t, i int32) {
    var fShowDelay int32  = 1;
    var pFanin *Str_Mux_t;
    if pMux.Edge[i].Fan <= 0 {
        fmt.Printf("%d", -pMux.Edge[i].Fan);
        if fShowDelay != 0 {
            fmt.Printf("{%d}", pMux.Edge[i].FanDel); }
        return;
    }
    pFanin = Str_MuxFanin( pMux, i );
    fmt.Printf("[ ");
    if pFanin.Edge[0].fCompl {
        fmt.Printf("!"); }
    Str_MuxDelayPrint_rec( pFanin, 0 );
    fmt.Printf("|");
    if pFanin.Edge[1].fCompl {
        fmt.Printf("!"); }
    Str_MuxDelayPrint_rec( pFanin, 1 );
    fmt.Printf("(");
    if pFanin.Edge[2].fCompl {
        fmt.Printf("!"); }
    Str_MuxDelayPrint_rec( pFanin, 2 );
    fmt.Printf(")");
    fmt.Printf(" ]");
}
func Str_MuxDelayEdge_rec(pMux *Str_Mux_t, i int32) int32 {
    if pMux.Edge[i].Fan > 0 {
        var pFanin *Str_Mux_t  = Str_MuxFanin( pMux, i );
        Str_MuxDelayEdge_rec( pFanin, 0 );
        Str_MuxDelayEdge_rec( pFanin, 1 );
        pMux.Edge[i].FanDel = Str_Delay3( pFanin.Edge[0].FanDel, pFanin.Edge[1].FanDel, pFanin.Edge[2].FanDel, pFanin.nLutSize );
    }
    return pMux.Edge[i].FanDel;
}
func Str_MuxCreate(pTree *Str_Mux_t, pNtk *Str_Ntk_t, iMux int32, nMuxes int32, vDelay *Vec_Int_t, nLutSize int32) {
    var pObj *Str_Obj_t;
    var pMux *Str_Mux_t;
    var i, k, nPis int32;
    assert(nMuxes >= 2, "nMuxes >= 2");
    memset( pTree, 0, sizeof(Str_Mux_t) * (nMuxes + 1) );
    pTree.nLutSize = nLutSize;
    pTree.Edge[0].Fan = 1;
    for  i = 1; i <= nMuxes; i++  {
        pMux = pTree + i;
        pMux.Id = i;
        pMux.nLutSize = nLutSize;
        pMux.Delay = pMux.Copy = -1;
        // assign fanins
        pObj = Str_NtkObj( pNtk, iMux + nMuxes - i );
        assert(pObj.Type == STR_MUX, "pObj.Type == STR_MUX");
        for  k = 0; k < 3; k++  {
            pMux.Edge[k].fCompl = Str_ObjFaninC(pNtk, pObj, k);
            if Str_ObjFaninId(pNtk, pObj, k) >= iMux {
                pMux.Edge[k].Fan = iMux + nMuxes - Str_ObjFaninId(pNtk, pObj, k);
            } else {
                pMux.Edge[k].Fan = -nPis++; // count external inputs, including controls
                pMux.Edge[k].Copy = Str_ObjFanin(pNtk, pObj, k).iCopy;
                pMux.Edge[k].FanDel = Vec_IntEntry( vDelay, Abc_Lit2Var(pMux.Edge[k].Copy) );
            }
        }
    }
}
func Str_MuxToGia_rec(pNew *Gia_Man_t, pMux *Str_Mux_t, i int32, vDelay *Vec_Int_t) int32 {
    if pMux.Edge[i].Fan > 0 {
        var pFanin *Str_Mux_t  = Str_MuxFanin( pMux, i );
        var iLit0 int32  = Str_MuxToGia_rec( pNew, pFanin, 0, vDelay );
        var iLit1 int32  = Str_MuxToGia_rec( pNew, pFanin, 1, vDelay );
        assert(pFanin.Edge[2].Fan <= 0, "pFanin.Edge[2].Fan <= 0");
        assert(pFanin.Edge[2].fCompl == 0, "pFanin.Edge[2].fCompl == 0");
        pMux.Edge[i].Copy = Gia_ManHashMuxReal( pNew, pFanin.Edge[2].Copy, iLit1, iLit0 );
        Str_ObjDelay( pNew, Abc_Lit2Var(pMux.Edge[i].Copy), pFanin.nLutSize, vDelay );
    }
    return Abc_LitNotCond( pMux.Edge[i].Copy, pMux.Edge[i].fCompl );
}
func Str_MuxChangeOnce(pTree *Str_Mux_t, pPath *int32, i int32, k int32, pBackup *Str_Mux_t, pNew *Gia_Man_t, vDelay *Vec_Int_t) {
    var pSpots [3]*Str_Mux_t;
    var pInds [3]int32;
    var MidFan int32;
    var MidCom int32;
    var MidDel int32;
    var MidCop int32;
    var c int32;
    var iRes, iCond, fCompl int32;
    // save backup
    assert(i + 1 < k, "i + 1 < k");
    if pBackup != nil {
        pBackup[0] = pTree[ Abc_Lit2Var(pPath[k])  ];
        pBackup[1] = pTree[ Abc_Lit2Var(pPath[i+1])];
        pBackup[2] = pTree[ Abc_Lit2Var(pPath[i])  ];
    }
    // perform changes
    pSpots[0] = pTree + Abc_Lit2Var(pPath[k]); 
    pSpots[1] = pTree + Abc_Lit2Var(pPath[i+1]); 
    pSpots[2] = pTree + Abc_Lit2Var(pPath[i]); 
    pInds[0] = Abc_LitIsCompl(pPath[k]);
    pInds[1] = Abc_LitIsCompl(pPath[i+1]);
    pInds[2] = Abc_LitIsCompl(pPath[i]);
    // check
    assert(pSpots[0].Edge[pInds[0]].Fan > 0, "pSpots[0].Edge[pInds[0]].Fan > 0");
    assert(pSpots[1].Edge[pInds[1]].Fan > 0, "pSpots[1].Edge[pInds[1]].Fan > 0");
    // collect complement
    fCompl = 0;
    for  c = i+1; c < k; c++  {
        fCompl ^= pTree[Abc_Lit2Var(pPath[c])].Edge[Abc_LitIsCompl(pPath[c])].fCompl;
    }
    // remember bottom side
    MidFan = pSpots[2].Edge[!pInds[2]].Fan;
    MidCom = pSpots[2].Edge[!pInds[2]].fCompl;
    MidDel = pSpots[2].Edge[!pInds[2]].FanDel;
    MidCop = pSpots[2].Edge[!pInds[2]].Copy;
    // update bottom
    pSpots[2].Edge[!pInds[2]].Fan    = pSpots[0].Edge[pInds[0]].Fan;
    pSpots[2].Edge[!pInds[2]].fCompl = 0;
    // update top 
    pSpots[0].Edge[pInds[0]].Fan     = pSpots[2].Id;
    // update middle
    pSpots[1].Edge[pInds[1]].Fan     = MidFan;
    pSpots[1].Edge[pInds[1]].fCompl ^= MidCom;
    pSpots[1].Edge[pInds[1]].FanDel  = MidDel;
    pSpots[1].Edge[pInds[1]].Copy    = MidCop;
    // update delay of the control
    for  c = i + 1; c < k; c++  {
        pSpots[2].Edge[2].FanDel = Str_Delay2( pSpots[2].Edge[2].FanDel, pTree[Abc_Lit2Var(pPath[c])].Edge[2].FanDel, pTree.nLutSize );
    }
    if pNew == nil {
        return; }
    // create AND gates
    iRes = 1;
    for  c = i; c < k; c++  {
        assert(pTree[Abc_Lit2Var(pPath[c])].Edge[2].fCompl == 0, "pTree[Abc_Lit2Var(pPath[c])].Edge[2].fCompl == 0");
        iCond = pTree[Abc_Lit2Var(pPath[c])].Edge[2].Copy;
        iCond = Abc_LitNotCond( iCond, !Abc_LitIsCompl(pPath[c]) );
        iRes  = Gia_ManHashAnd( pNew, iRes, iCond );
        Str_ObjDelay( pNew, Abc_Lit2Var(iRes), pTree.nLutSize, vDelay );
    }
    // complement the condition
    pSpots[2].Edge[2].Copy = Abc_LitNotCond( iRes, !Abc_LitIsCompl(pPath[i]) );
    // complement the path
    pSpots[2].Edge[pInds[2]].fCompl ^= fCompl;
}
func Str_MuxChangeUndo(pTree *Str_Mux_t, pPath *int32, i int32, k int32, pBackup *Str_Mux_t) {
    pTree[ Abc_Lit2Var(pPath[k])  ] = pBackup[0]; 
    pTree[ Abc_Lit2Var(pPath[i+1])] = pBackup[1]; 
    pTree[ Abc_Lit2Var(pPath[i])  ] = pBackup[2]; 
}
func Str_MuxFindPathEdge_rec(pMux *Str_Mux_t, i int32, pPath *int32, pnLength *int32) int32 {
    extern int Str_MuxFindPath_rec( Str_Mux_t * pMux, int * pPath, int * pnLength );
    if pMux.Edge[i].Fan > 0 && !(Str_MuxFindPath_rec(Str_MuxFanin(pMux, i), pPath, pnLength) != 0) {
        return 0; }
    pPath[ (*pnLength)++ ] = Abc_Var2Lit(pMux.Id, i);
    return 1;
}
func Str_MuxFindPath_rec(pMux *Str_Mux_t, pPath *int32, pnLength *int32) int32 {
    var i int32;
    var DelayMax int32 = Abc_MaxInt( pMux.Edge[0].FanDel, Abc_MaxInt(pMux.Edge[1].FanDel, pMux.Edge[2].FanDel) );
    for  i = 0; i < 2; i++  {
        if pMux.Edge[i].FanDel == DelayMax {
            return Str_MuxFindPathEdge_rec( pMux, i, pPath, pnLength ); }
    }
    if pMux.Edge[2].FanDel == DelayMax {
        return 0; }
    assert(false, "0");
    return -1;
}
// return node whose both branches are non-trivial
func Str_MuxFindBranching(pRoot *Str_Mux_t, i int32) *Str_Mux_t {
    var pMux *Str_Mux_t;
    if pRoot.Edge[i].Fan <= 0 {
        return nil; }
    pMux = Str_MuxFanin( pRoot, i );
    for {
        if pMux.Edge[0].Fan <= 0 && pMux.Edge[1].Fan <= 0 {
            return nil; }
        if pMux.Edge[0].Fan > 0 && pMux.Edge[1].Fan > 0 {
            return pMux; }
        if pMux.Edge[0].Fan > 0 {
            pMux = Str_MuxFanin( pMux, 0 ); }
        if pMux.Edge[1].Fan > 0 {
            pMux = Str_MuxFanin( pMux, 1 ); }
    }
    assert(false, "0");
    return nil;
}
func Str_MuxTryOnce(pNew *Gia_Man_t, pNtk *Str_Ntk_t, pTree *Str_Mux_t, pRoot *Str_Mux_t, Edge int32, vDelay *Vec_Int_t, fVerbose int32) int32 {
    var pPath [MAX_TREE]int32;
    var pBackup [3]Str_Mux_t;
    var Delay int32;
    var DelayBest int32 = Str_MuxDelayEdge_rec( pRoot, Edge );
    var DelayInit int32 = DelayBest;
    var i int32;
    var k int32;
    var nLength int32;
    var ForkBest int32 = -1;
    var nChecks int32;
    var RetValue int32  = Str_MuxFindPathEdge_rec( pRoot, Edge, pPath, &nLength );
    if RetValue == 0 {
        return 0; }
    if fVerbose != 0 {
        fmt.Printf("Trying node %d with path of length %d.\n", pRoot.Id, nLength); }
    for  i = 0; i < nLength; i++  {
    for  k = i+2; k < nLength; k++  {
        Str_MuxChangeOnce( pTree, pPath, i, k, pBackup, nil, nil );
        Delay = Str_MuxDelayEdge_rec( pRoot, Edge );
        Str_MuxChangeUndo( pTree, pPath, i, k, pBackup );
        if DelayBest > Delay || (ForkBest > 0 && DelayBest == Delay) {
            DelayBest = Delay;
ForkBest = (i << 16) | k; }
        if fVerbose != 0 {
            fmt.Printf("%2d %2d -> %3d (%3d)\n", i, k, Delay, DelayBest); }
        nChecks++;
    }
    }
    if ForkBest == -1 {
        if fVerbose != 0 {
            fmt.Printf("Did not find!\n"); }
        return 0;
    }
//    Str_MuxDelayPrint_rec( pRoot, Edge ); printf( "\n" );
    Str_MuxChangeOnce( pTree, pPath, ForkBest >> 16, ForkBest & 0xFFFF, nil, pNew, vDelay );
//    Str_MuxDelayPrint_rec( pRoot, Edge ); printf( "\n" );
    if fVerbose != 0 {
        fmt.Printf("Node %6d (%3d %3d) : Checks = %d. Delay: %d -> %d.\n", pRoot.Id, ForkBest >> 16, ForkBest & 0xFFFF, nChecks, DelayInit, DelayBest); }
    if fVerbose != 0 {
        fmt.Printf("\n"); }
    return 1;
}
func Str_MuxRestruct_rec(pNew *Gia_Man_t, pNtk *Str_Ntk_t, pTree *Str_Mux_t, pRoot *Str_Mux_t, Edge int32, vDelay *Vec_Int_t, fVerbose int32) int32 {
    var fChanges int32;
    var pMux *Str_Mux_t  = Str_MuxFindBranching( pRoot, Edge );
    if pMux != nil {
        fChanges |= Str_MuxRestruct_rec( pNew, pNtk, pTree, pMux, 0, vDelay, fVerbose ); }
    if pMux != nil {
        fChanges |= Str_MuxRestruct_rec( pNew, pNtk, pTree, pMux, 1, vDelay, fVerbose ); }
    fChanges |= Str_MuxTryOnce( pNew, pNtk, pTree, pRoot, Edge, vDelay, fVerbose );
    return fChanges;
}
func Str_MuxRestructure2(pNew *Gia_Man_t, pNtk *Str_Ntk_t, iMux int32, nMuxes int32, vDelay *Vec_Int_t, nLutSize int32, fVerbose int32) int32 {
    var Limit int32  = MAX_TREE;
    var pTree [MAX_TREE]Str_Mux_t;
    var Delay, Delay2, fChanges int32;
    if nMuxes >= Limit {
        return -1; }
    assert(nMuxes < Limit, "nMuxes < Limit");
    Str_MuxCreate( pTree, pNtk, iMux, nMuxes, vDelay, nLutSize );
    Delay = Str_MuxDelayEdge_rec( pTree, 0 );
    for {     
        if !(Str_MuxRestruct_rec(pNew, pNtk, pTree, pTree, 0, vDelay, fVerbose) != 0) {
            break; }
        fChanges = 1;
    }
    if !(fChanges != 0) {
        return -1; }
    Delay2 = Str_MuxDelayEdge_rec( pTree, 0 );
//    printf( "Improved delay for tree %d with %d MUXes (%d -> %d).\n", iMux, nMuxes, Delay, Delay2 );
    pNtk.DelayGain += Delay - Delay2;
    return Str_MuxToGia_rec( pNew, pTree, 0, vDelay );
}
func Str_MuxRestructure1(pNew *Gia_Man_t, pNtk *Str_Ntk_t, iMux int32, nMuxes int32, vDelay *Vec_Int_t, nLutSize int32, fVerbose int32) int32 {
    var Limit int32  = MAX_TREE;
    var pTree [MAX_TREE]Str_Mux_t;
    var Delay, Delay2, fChanges int32;
    if nMuxes >= Limit {
        return -1; }
    assert(nMuxes < Limit, "nMuxes < Limit");
    Str_MuxCreate( pTree, pNtk, iMux, nMuxes, vDelay, nLutSize );
    Delay = Str_MuxDelayEdge_rec( pTree, 0 );
    for {     
        if !(Str_MuxTryOnce(pNew, pNtk, pTree, pTree, 0, vDelay, fVerbose) != 0) {
            break; }
        fChanges = 1;
    }
    if !(fChanges != 0) {
        return -1; }
    Delay2 = Str_MuxDelayEdge_rec( pTree, 0 );
//    printf( "Improved delay for tree %d with %d MUXes (%d -> %d).\n", iMux, nMuxes, Delay, Delay2 );
    pNtk.DelayGain += Delay - Delay2;
    return Str_MuxToGia_rec( pNew, pTree, 0, vDelay );
}
func Str_MuxRestructure(pNew *Gia_Man_t, pNtk *Str_Ntk_t, iMux int32, nMuxes int32, vDelay *Vec_Int_t, nLutSize int32, fRecursive int32, fOptArea int32, fVerbose int32) int32 {
    extern int Str_MuxRestructureArea( Gia_Man_t * pNew, Str_Ntk_t * pNtk, int iMux, int nMuxes, Vec_Int_t * vDelay, int nLutSize, int fVerbose );
    if fOptArea != 0 {
        if nMuxes < 2 {
            return Str_MuxRestructure1( pNew, pNtk, iMux, nMuxes, vDelay, nLutSize, fVerbose ); }
        return Str_MuxRestructureArea( pNew, pNtk, iMux, nMuxes, vDelay, nLutSize, fVerbose );
    }
    if fRecursive != 0 {
        return Str_MuxRestructure2( pNew, pNtk, iMux, nMuxes, vDelay, nLutSize, fVerbose ); }
    return Str_MuxRestructure1( pNew, pNtk, iMux, nMuxes, vDelay, nLutSize, fVerbose );
}

/**Function*************************************************************

  Synopsis    [Perform MUX restructuring for area.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Str_MuxRestructAreaThree(pNew *Gia_Man_t, pMux *Str_Mux_t, vDelay *Vec_Int_t, fVerbose int32) int32 {
    var iRes int32;
    var pFanin0 *Str_Mux_t  = Str_MuxFanin( pMux, 0 );
    var pFanin1 *Str_Mux_t  = Str_MuxFanin( pMux, 1 );
    assert(pMux.Copy == -1, "pMux.Copy == -1");
    pMux.Copy = -2;
    if pFanin0.Edge[2].Copy == pFanin1.Edge[2].Copy {
        return 0; }
    iRes = Gia_ManHashMuxReal( pNew, pMux.Edge[2].Copy, pFanin1.Edge[2].Copy, pFanin0.Edge[2].Copy );
    Str_ObjDelay( pNew, Abc_Lit2Var(iRes), pMux.nLutSize, vDelay );
    pFanin0.Edge[2].Copy = pFanin1.Edge[2].Copy = iRes;
//    printf( "Created triple\n" );
    return 0;
}
func Str_MuxRestructArea_rec(pNew *Gia_Man_t, pTree *Str_Mux_t, pRoot *Str_Mux_t, i int32, vDelay *Vec_Int_t, fVerbose int32) int32 {
    var Path [4]int32;
    var fSkipMoving int32  = 1;
    var pMux, pFanin0, pFanin1 *Str_Mux_t;
    var nMuxes0, nMuxes1 int32;
    if pRoot.Edge[i].Fan <= 0 {
        return 0; }
    pMux    = Str_MuxFanin( pRoot, i );
    nMuxes0 = Str_MuxRestructArea_rec( pNew, pTree, pMux, 0, vDelay, fVerbose );
    nMuxes1 = Str_MuxRestructArea_rec( pNew, pTree, pMux, 1, vDelay, fVerbose );
    if nMuxes0 + nMuxes1 < 2 {
        return 1 + nMuxes0 + nMuxes1; }
    if nMuxes0 + nMuxes1 == 2 {
        if nMuxes0 == 2 || nMuxes1 == 2 {
            pFanin0 = Str_MuxFanin( pMux, int32((nMuxes1 == 2)) );
            assert(Str_MuxHasFanin(pFanin0, 0) != Str_MuxHasFanin(pFanin0, 1), "Str_MuxHasFanin(pFanin0, 0) != Str_MuxHasFanin(pFanin0, 1)");
            Path[2] = Abc_Var2Lit(pRoot.Id, i);
            Path[1] = Abc_Var2Lit(pMux.Id, int32((nMuxes1 == 2)) );
            Path[0] = Abc_Var2Lit(pFanin0.Id, Str_MuxHasFanin(pFanin0, 1));
            Str_MuxChangeOnce( pTree, Path, 0, 2, nil, pNew, vDelay );
        }
        Str_MuxRestructAreaThree( pNew, Str_MuxFanin(pRoot, i), vDelay, fVerbose );
        return 0;
    }
    assert(nMuxes0 + nMuxes1 == 3 || nMuxes0 + nMuxes1 == 4, "nMuxes0 + nMuxes1 == 3 || nMuxes0 + nMuxes1 == 4");
    assert(nMuxes0 == 2 || nMuxes1 == 2, "nMuxes0 == 2 || nMuxes1 == 2");
    if fSkipMoving != 0 {
        Str_MuxRestructAreaThree( pNew, pMux, vDelay, fVerbose );
        return 0;
    }
    if nMuxes0 == 2 {
        pFanin0 = Str_MuxFanin( pMux, 0 );
        assert(Str_MuxHasFanin(pFanin0, 0) != Str_MuxHasFanin(pFanin0, 1), "Str_MuxHasFanin(pFanin0, 0) != Str_MuxHasFanin(pFanin0, 1)");
        Path[3] = Abc_Var2Lit(pRoot.Id, i);
        Path[2] = Abc_Var2Lit(pMux.Id, 0 );
        Path[1] = Abc_Var2Lit(pFanin0.Id, Str_MuxHasFanin(pFanin0, 1));
        pFanin1 = Str_MuxFanin( pFanin0, Str_MuxHasFanin(pFanin0, 1) );
        assert(!(Str_MuxHasFanin(pFanin1, 0) != 0) && !(Str_MuxHasFanin(pFanin1, 1) != 0), "!Str_MuxHasFanin(pFanin1, 0) && !Str_MuxHasFanin(pFanin1, 1)");
        Path[0] = Abc_Var2Lit(pFanin1.Id, 0);
        Str_MuxChangeOnce( pTree, Path, 0, 3, nil, pNew, vDelay );
    }
    if nMuxes1 == 2 {
        pFanin0 = Str_MuxFanin( pMux, 1 );
        assert(Str_MuxHasFanin(pFanin0, 0) != Str_MuxHasFanin(pFanin0, 1), "Str_MuxHasFanin(pFanin0, 0) != Str_MuxHasFanin(pFanin0, 1)");
        Path[3] = Abc_Var2Lit(pRoot.Id, i);
        Path[2] = Abc_Var2Lit(pMux.Id, 1 );
        Path[1] = Abc_Var2Lit(pFanin0.Id, Str_MuxHasFanin(pFanin0, 1));
        pFanin1 = Str_MuxFanin( pFanin0, Str_MuxHasFanin(pFanin0, 1) );
        assert(!(Str_MuxHasFanin(pFanin1, 0) != 0) && !(Str_MuxHasFanin(pFanin1, 1) != 0), "!Str_MuxHasFanin(pFanin1, 0) && !Str_MuxHasFanin(pFanin1, 1)");
        Path[0] = Abc_Var2Lit(pFanin1.Id, 0);
        Str_MuxChangeOnce( pTree, Path, 0, 3, nil, pNew, vDelay );
    }
    Str_MuxRestructAreaThree( pNew, pMux, vDelay, fVerbose );
    return nMuxes0 + nMuxes1 - 2;
}
func Str_MuxRestructureArea(pNew *Gia_Man_t, pNtk *Str_Ntk_t, iMux int32, nMuxes int32, vDelay *Vec_Int_t, nLutSize int32, fVerbose int32) int32 {
    var Limit int32  = MAX_TREE;
    var pTree [MAX_TREE]Str_Mux_t;
    var Result int32;
    if nMuxes >= Limit {
        return -1; }
    assert(nMuxes < Limit, "nMuxes < Limit");
    Str_MuxCreate( pTree, pNtk, iMux, nMuxes, vDelay, nLutSize );
    Result = Str_MuxRestructArea_rec( pNew, pTree, pTree, 0, vDelay, fVerbose );
    assert(Result >= 0 && Result <= 2, "Result >= 0 && Result <= 2");
    return Str_MuxToGia_rec( pNew, pTree, 0, vDelay );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

