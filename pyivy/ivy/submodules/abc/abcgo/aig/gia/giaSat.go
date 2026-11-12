package gia

import (
	"fmt"
	"time"
)

/**CFile****************************************************************

  FileName    [giaSat.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [New constraint-propagation procedures.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSat.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

const GIA_LIMIT = 10


   
type Gia_ManSat_t struct {
    pMem     *Aig_MmFlex_t;
}

   
type Gia_ObjSat1_t struct {
    nFans              byte;
    nOffset              byte;
    PathsH              byte;
    PathsV              byte;
}

   
type Gia_ObjSat2_t struct {
    fTerm uint32; /* C bitfield: 1 bits; review packing and truncation */
    iLit uint32; /* C bitfield: 31 bits; review packing and truncation */
}

   
type Gia_ObjSat_t struct {
    union {
        Gia_ObjSat1_t Obj1;
        Gia_ObjSat2_t Obj2;
    };
}


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatStart() *Gia_ManSat_t {
    
    // inlining of macro ABC_CALLOC
    p := &Gia_ManSat_t{};
    p.pMem = Aig_MmFlexStart();
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatStop(p *Gia_ManSat_t) {
    Aig_MmFlexStop( p.pMem, 0 );
    // inlining of macro ABC_FREE
    p = nil;
}


/**Function*************************************************************

  Synopsis    [Collects the supergate rooted at this ]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatPartCollectSuper(p *Gia_Man_t, pObj *Gia_Obj_t, pLits *int32, pnLits *int32) {
    var pFanin *Gia_Obj_t;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
    pFanin = Gia_ObjFanin0(pObj);
    if pFanin.fMark0 || Gia_ObjFaninC0(pObj) {
        pLits[(*pnLits)++] = Gia_Var2Lit(Gia_ObjId(p, pFanin), Gia_ObjFaninC0(pObj));
    } else {
        Gia_ManSatPartCollectSuper(p, pFanin, pLits, pnLits); }
    pFanin = Gia_ObjFanin1(pObj);
    if pFanin.fMark0 || Gia_ObjFaninC1(pObj) {
        pLits[(*pnLits)++] = Gia_Var2Lit(Gia_ObjId(p, pFanin), Gia_ObjFaninC1(pObj));
    } else {
        Gia_ManSatPartCollectSuper(p, pFanin, pLits, pnLits); }
}

/**Function*************************************************************

  Synopsis    [Returns the number of words used.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatPartCreate_rec(p *Gia_Man_t, pObj *Gia_Obj_t, pObjPlace *int32, pStore *int32) int32 {
    var pFanin *Gia_Obj_t;
    var i int32;
    var nWordsUsed int32;
    var nSuperSize int32;
    var Super [2*GIA_LIMIT]int32;
    // make sure this is a valid node
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
    // collect inputs to the supergate
    Gia_ManSatPartCollectSuper( p, pObj, Super, &nSuperSize );
    assert(nSuperSize <= 2*GIA_LIMIT, "nSuperSize <= 2*GIA_LIMIT");
    // create the root entry
    *pObjPlace = 0;
    ((Gia_ObjSat1_t *)pObjPlace).nFans   = Gia_Var2Lit( nSuperSize, 0 );
    ((Gia_ObjSat1_t *)pObjPlace).nOffset = pStore - pObjPlace;
    nWordsUsed = nSuperSize;
    for  i = 0; i < nSuperSize; i++  {
        pFanin = Gia_ManObj( p, Gia_Lit2Var(Super[i]) );
        if pFanin.fMark0 {
            ((Gia_ObjSat2_t *)(pStore + i)).fTerm = 1;
            ((Gia_ObjSat2_t *)(pStore + i)).iLit  = Super[i];
        
        } else {
            assert(Gia_LitIsCompl(Super[i]), "Gia_LitIsCompl(Super[i])");
            nWordsUsed += Gia_ManSatPartCreate_rec( p, pFanin, pStore + i, pStore + nWordsUsed );
        }
    }
    return nWordsUsed;
}

/**Function*************************************************************

  Synopsis    [Creates part and returns the number of words used.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatPartCreate(p *Gia_Man_t, pObj *Gia_Obj_t, pStore *int32) int32 {
    return 1 + Gia_ManSatPartCreate_rec( p, pObj, pStore, pStore + 1 );
}


/**Function*************************************************************

  Synopsis    [Count the number of internal nodes in the leaf-DAG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatPartCountClauses(p *Gia_Man_t, pObj *Gia_Obj_t, pnOnset *int32, pnOffset *int32) {
    var pFanin *Gia_Obj_t;
    var nOnset0, nOnset1, nOffset0, nOffset1 int32;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    pFanin = Gia_ObjFanin0(pObj);
    if pFanin.fMark0 {
        nOnset0 = 1;
nOffset0 = 1;
    } else {
        Gia_ManSatPartCountClauses(p, pFanin, &nOnset0, &nOffset0);
        if Gia_ObjFaninC0(pObj) {
            var Temp int32  = nOnset0;
            nOnset0 = nOffset0;
            nOffset0 = Temp;
        }
    }
    pFanin = Gia_ObjFanin1(pObj);
    if pFanin.fMark0 {
        nOnset1 = 1;
nOffset1 = 1;
    } else {
        Gia_ManSatPartCountClauses(p, pFanin, &nOnset1, &nOffset1);
        if Gia_ObjFaninC1(pObj) {
            var Temp int32  = nOnset1;
            nOnset1 = nOffset1;
            nOffset1 = Temp;
        }
    }
    *pnOnset  = nOnset0  * nOnset1;
    *pnOffset = nOffset0 + nOffset1;
}

/**Function*************************************************************

  Synopsis    [Count the number of internal nodes in the leaf-DAG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatPartCount(p *Gia_Man_t, pObj *Gia_Obj_t, pnLeaves *int32, pnNodes *int32) int32 {
    var pFanin *Gia_Obj_t;
    var Level0, Level1 int32;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
    (*pnNodes)++;
    pFanin = Gia_ObjFanin0(pObj);
    if pFanin.fMark0 {
        (*pnLeaves)++;
    } else {
        Level0 = Gia_ManSatPartCount(p, pFanin, pnLeaves, pnNodes) + Gia_ObjFaninC0(pObj); }
    pFanin = Gia_ObjFanin1(pObj);
    if pFanin.fMark0 {
        (*pnLeaves)++;
    } else {
        Level1 = Gia_ManSatPartCount(p, pFanin, pnLeaves, pnNodes) + Gia_ObjFaninC1(pObj); }
    return Abc_MaxInt( Level0, Level1 );
}

/**Function*************************************************************

  Synopsis    [Count the number of internal nodes in the leaf-DAG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatPartCountNodes(p *Gia_Man_t, pObj *Gia_Obj_t) int32 {
    var pFanin *Gia_Obj_t;
    var nNodes0, nNodes1 int32;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
    pFanin = Gia_ObjFanin0(pObj);
    if !(pFanin.fMark0) {
        nNodes0 = Gia_ManSatPartCountNodes(p, pFanin); }
    pFanin = Gia_ObjFanin1(pObj);
    if !(pFanin.fMark0) {
        nNodes1 = Gia_ManSatPartCountNodes(p, pFanin); }
    return nNodes0 + nNodes1 + 1;
}

/**Function*************************************************************

  Synopsis    [Count the number of internal nodes in the leaf-DAG.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatPartPrint(p *Gia_Man_t, pObj *Gia_Obj_t, Step int32) {
    var pFanin *Gia_Obj_t;
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    assert(pObj.fMark0 == 0, "pObj.fMark0 == 0");
    pFanin = Gia_ObjFanin0(pObj);
    if pFanin.fMark0 {
        fmt.Printf("%s%d", Gia_ObjFaninC0(pObj)?"!":"", Gia_ObjId(p,pFanin));
    } else {
        if Gia_ObjFaninC0(pObj) {
            fmt.Printf("("); }
        Gia_ManSatPartPrint(p, pFanin, Step + Gia_ObjFaninC0(pObj));
        if Gia_ObjFaninC0(pObj) {
            fmt.Printf(")"); }
    }
    fmt.Printf("%s", (Step & 1)? " + " : "*");
    pFanin = Gia_ObjFanin1(pObj);
    if pFanin.fMark0 {
        fmt.Printf("%s%d", Gia_ObjFaninC1(pObj)?"!":"", Gia_ObjId(p,pFanin));
    } else {
        if Gia_ObjFaninC1(pObj) {
            fmt.Printf("("); }
        Gia_ManSatPartPrint(p, pFanin, Step + Gia_ObjFaninC1(pObj));
        if Gia_ObjFaninC1(pObj) {
            fmt.Printf(")"); }
    }
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSatExperiment(p *Gia_Man_t) {
    var nStored int32;
    var Storage [1000]int32;
    var pStart *int32; 
    var pMan *Gia_ManSat_t;
    var pObj *Gia_Obj_t;
    var i int32;
    var nLevels int32;
    var nLeaves int32;
    var nNodes int32;
    var nCount [2*GIA_LIMIT+2]int32 ;
    var nCountAll int32;
    var Num0, Num1 int32;
    var clk int64  = time.Now().UnixNano();
    var nWords, nWords2 int32;
    pMan = Gia_ManSatStart();
    // mark the nodes to become roots of leaf-DAGs
    Gia_ManCreateValueRefs( p );
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.fMark0 = 0;
        if Gia_ObjIsCo(pObj) {
            Gia_ObjFanin0(pObj).fMark0 = 1;
        } else if Gia_ObjIsCi(pObj) {
            pObj.fMark0 = 1;
        } else if Gia_ObjIsAnd(pObj) {
            if pObj.Value > 1 || Gia_ManSatPartCountNodes(p,pObj) >= GIA_LIMIT {
                pObj.fMark0 = 1; }
        }
        pObj.Value = 0;
    }
    // compute the sizes of leaf-DAGs
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        if pObj.fMark0 == 0 {
            continue; }
        pObj.fMark0 = 0;

        nCountAll++;
        nStored = Gia_ManSatPartCreate( p, pObj, Storage );
        nWords2 += nStored;
        assert(nStored < 500, "nStored < 500");
        pStart = (int *)Aig_MmFlexEntryFetch( pMan.pMem, sizeof(int) * nStored );
        memcpy( pStart, Storage, sizeof(int) * nStored );

        nLeaves = nNodes = 0;
        nLevels = 1+Gia_ManSatPartCount( p, pObj, &nLeaves, &nNodes );
        nWords += nLeaves + nNodes;
        if nNodes <= 2*GIA_LIMIT {
            nCount[nNodes]++;
        } else {
            nCount[2*GIA_LIMIT+1]++; }
//        if ( nNodes > 10 && i % 100 == 0 )
//        if ( nNodes > 5 )
        if false {
            Gia_ManSatPartCountClauses( p, pObj, &Num0, &Num1 );
            fmt.Printf("%8d :  And = %3d.  Lev = %2d. Clauses = %3d. (%3d + %3d).\n", i, nNodes, nLevels, Num0+Num1, Num0, Num1);
            Gia_ManSatPartPrint( p, pObj, 0 );
            fmt.Printf("\n");
        }

        pObj.fMark0 = 1;
    }
    }
    fmt.Printf("\n");
    // inlining of macro Gia_ManForEachObj
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        pObj.fMark0 = 0;
    }
    Gia_ManSatStop( pMan );
    for  i = 0; i < 2*GIA_LIMIT+2; i++  {
        fmt.Printf("%2d=%6d  %7.2f %%  %7.2f %%\n", i, nCount[i], 100.0*nCount[i]/nCountAll, 100.0*i*nCount[i]/Gia_ManAndNum(p));
    }
    // inlining of macro ABC_PRMn
    Abc_Print(1, "%s =", ("MemoryEst"));
Abc_Print(1, "%10.3f MB  ",    1.0*(float64((4*nWords)))/(1<<20));;
    // inlining of macro ABC_PRMn
    Abc_Print(1, "%s =", ("MemoryReal"));
Abc_Print(1, "%10.3f MB  ",    1.0*(float64((4*nWords2)))/(1<<20));;
    fmt.Printf("%5.2f bpn  ", 4.0*nWords2/Gia_ManObjNum(p));
    // inlining of macro ABC_PRT
    Abc_Print(1, "%s =", ("Time"));
Abc_Print(1, "%9.2f sec\n", 1.0*(float64((time.Now().UnixNano() - clk)))/(float64(int64(time.Second))));;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

