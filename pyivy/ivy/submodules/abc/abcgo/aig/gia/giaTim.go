package gia

import "fmt"

/**CFile****************************************************************

  FileName    [giaTim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Procedures with hierarchy/timing manager.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaTim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "giaAig.h"
//#include "misc/tim/tim.h"
//#include "misc/extra/extra.h"
//#include "proof/cec/cec.h"
//#include "proof/fra/fra.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Returns the number of boxes in the AIG with boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Returns one if this is a seq AIG with non-trivial boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/

/**Function*************************************************************

  Synopsis    [Reorders flops for sequential AIGs with boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupUnshuffleInputs(p *Gia_Man_t) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i, nCIs, nAll, nPis int32;
    // sanity checks
    assert((Gia_ManIsNormalized(p) != 0), "Gia_ManIsNormalized(p)");
    assert((Gia_ManIsSeqWithBoxes(p) != 0), "Gia_ManIsSeqWithBoxes(p)");
    Gia_ManFillValue( p );
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // change input order
    // desired reorder:   PIs + NewCIs + FOs
    // current CI order:  PIs + FOs + NewCIs
    nCIs = Tim_ManPiNum( (Tim_Man_t *)p.pManTime );
    nAll = Tim_ManCiNum( (Tim_Man_t *)p.pManTime );
    nPis = nCIs - Gia_ManRegNum(p);
    assert(nAll == Gia_ManCiNum(p), "nAll == Gia_ManCiNum(p)");
    assert(nPis > 0, "nPis > 0");
    // copy PIs first
    for  i = 0; i < nPis; i++  {
        Gia_ManCi(p, i).Value = Gia_ManAppendCi(pNew);
    }
    // copy flops second
    for  i = nAll - Gia_ManRegNum(p); i < nAll; i++  {
        Gia_ManCi(p, i).Value = Gia_ManAppendCi(pNew);
    }
    // copy new CIs last
    for  i = nPis; i < nAll - Gia_ManRegNum(p); i++  {
        Gia_ManCi(p, i).Value = Gia_ManAppendCi(pNew);
    }
    fmt.Printf("Warning: Unshuffled CI order to be correct AIG with boxes.\n");
    // other things
    // inlining of macro Gia_ManForEachAnd
    for  i = 0; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {      if !Gia_ObjIsAnd(pObj) {} else {
        pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) ); }
    }
    // inlining of macro Gia_ManForEachCo
    for  i = 0; (i < Vec_IntSize(p.vCos)) && ((pObj) = Gia_ManCo(p, i)); i++  {
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    pNew.nConstrs = p.nConstrs;
    assert((Gia_ManIsNormalized(pNew) != 0), "Gia_ManIsNormalized(pNew)");
    Gia_ManDupRemapEquiv( pNew, p );
    return pNew;
}


/**Function*************************************************************

  Synopsis    [Find the ordering of AIG objects.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Duplicates AIG according to the timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Remaps the AIG from the old manager into the new manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManCleanupRemap(p *Gia_Man_t, pGia *Gia_Man_t) {
    var pObj, pObjGia *Gia_Obj_t;
    var i, iPrev int32;
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        iPrev = Gia_ObjValue(pObj);
        if iPrev == ~0 {
            continue; }
        pObjGia = Gia_ManObj( pGia, Abc_Lit2Var(iPrev) );
        if pObjGia.Value == ~0 {
            Gia_ObjSetValue( pObj, pObjGia.Value ); 
        } else {
            Gia_ObjSetValue( pObj, Abc_LitNotCond(pObjGia.Value, Abc_LitIsCompl(iPrev)) ); }
    }
}

/**Function*************************************************************

  Synopsis    [Computes level with boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/



/**Function*************************************************************

  Synopsis    [Computes level with boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Update hierarchy/timing manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
/**Function*************************************************************

  Synopsis    [Duplicates AIG while moving the last CIs to be after PIs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupMoveLast(p *Gia_Man_t, iInsert int32, nItems int32) *Gia_Man_t {
    var pNew *Gia_Man_t;
    var pObj *Gia_Obj_t;
    var i int32;
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    Gia_ManConst0(p).Value = 0;
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if i < iInsert {
            pObj.Value = Gia_ManAppendCi( pNew ); }
    }
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if i >= Gia_ManCiNum(p) - nItems {
            pObj.Value = Gia_ManAppendCi( pNew ); }
    }
    // inlining of macro Gia_ManForEachCi
    for  i = 0; (i < Vec_IntSize(p.vCis)) && ((pObj) = Gia_ManCi(p, i)); i++  {
        if i >= iInsert && i < Gia_ManCiNum(p) - nItems {
            pObj.Value = Gia_ManAppendCi( pNew ); }
    }
    // inlining of macro Gia_ManForEachObj1
    for  i = 1; (i < p.nObjs) && ((pObj) = Gia_ManObj(p, i)); i++  {
        if Gia_ObjIsCi(pObj) {
            continue; }
        if Gia_ObjIsAnd(pObj) {
            pObj.Value = Gia_ManAppendAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
        } else if Gia_ObjIsCo(pObj) {
            pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
        } else { assert(false, "0"); }
    }
    Gia_ManSetRegNum( pNew, Gia_ManRegNum(p) );
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Computes AIG with boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManDupCollapse_rec(p *Gia_Man_t, pObj *Gia_Obj_t, pNew *Gia_Man_t) {
    if Gia_ObjIsTravIdCurrent(p, pObj) {
        return; }
    Gia_ObjSetTravIdCurrent(p, pObj);
    assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)");
    if Gia_ObjSibl(p, Gia_ObjId(p, pObj)) {
        Gia_ManDupCollapse_rec( p, Gia_ObjSiblObj(p, Gia_ObjId(p, pObj)), pNew ); }
    Gia_ManDupCollapse_rec( p, Gia_ObjFanin0(pObj), pNew );
    Gia_ManDupCollapse_rec( p, Gia_ObjFanin1(pObj), pNew );
//    assert( !~pObj->Value );
    pObj.Value = Gia_ManHashAnd( pNew, Gia_ObjFanin0Copy(pObj), Gia_ObjFanin1Copy(pObj) );
    if Gia_ObjSibl(p, Gia_ObjId(p, pObj)) {
        pNew.pSibls[Abc_Lit2Var(pObj.Value)] = Abc_Lit2Var(Gia_ObjSiblObj(p, Gia_ObjId(p, pObj)).Value); }        
}
func Gia_ManDupCollapse(p *Gia_Man_t, pBoxes *Gia_Man_t, vBoxPres *Vec_Int_t, fSeq int32) *Gia_Man_t {
    // this procedure assumes that sequential AIG with boxes is unshuffled to have valid boxes
    var pManTime *Tim_Man_t  = (Tim_Man_t *)p.pManTime;
    var pNew, pTemp *Gia_Man_t;
    var pObj, pObjBox *Gia_Obj_t;
    var i, k, curCi, curCo, nBBins, nBBouts, nNewPis int32;
    assert(!(fSeq != 0) || p.vRegClasses, "!fSeq || p.vRegClasses");
    //assert( Gia_ManRegNum(p) == 0 );
    assert(Gia_ManCiNum(p) == Tim_ManPiNum(pManTime) + Gia_ManCoNum(pBoxes), "Gia_ManCiNum(p) == Tim_ManPiNum(pManTime) + Gia_ManCoNum(pBoxes)");
    pNew = Gia_ManStart( Gia_ManObjNum(p) );
    pNew.pName = Abc_UtilStrsav( p.pName );
    pNew.pSpec = Abc_UtilStrsav( p.pSpec );
    if Gia_ManHasChoices(p) {
        // inlining of macro ABC_CALLOC
        pNew.pSibls = ((int *) calloc((size_t)(Gia_ManObjNum(p)), sizeof(int))); }
    Gia_ManHashAlloc( pNew );
    // copy const and real PIs
    Gia_ManFillValue( p );
    Gia_ManConst0(p).Value = 0;
    Gia_ManIncrementTravId( p );
    Gia_ObjSetTravIdCurrent( p, Gia_ManConst0(p) );
    for  i = 0; i < Tim_ManPiNum(pManTime); i++  {
        pObj = Gia_ManCi( p, i );
        pObj.Value = Gia_ManAppendCi(pNew);
        Gia_ObjSetTravIdCurrent( p, pObj );
    }
    // create logic for each box
    curCi = Tim_ManPiNum(pManTime);
    curCo = 0;
    for  i = 0; i < Tim_ManBoxNum(pManTime); i++  {
        // clean boxes
        Gia_ManIncrementTravId( pBoxes );
        Gia_ObjSetTravIdCurrent( pBoxes, Gia_ManConst0(pBoxes) );
        Gia_ManConst0(pBoxes).Value = 0;
        // add internal nodes
        //printf( "%d ", Tim_ManBoxIsBlack(pManTime, i) );
        if Tim_ManBoxIsBlack(pManTime, i) {
            var fSkip int32  = (vBoxPres != nil && !Vec_IntEntry(vBoxPres, i));
            for  k = 0; k < Tim_ManBoxInputNum(pManTime, i); k++  {
                pObj = Gia_ManCo( p, curCo + k );
                Gia_ManDupCollapse_rec( p, Gia_ObjFanin0(pObj), pNew );
                if fSkip != 0 {
                    pObj.Value = -1;
                } else {
                    pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
                }
                nBBouts++;
            }
            for  k = 0; k < Tim_ManBoxOutputNum(pManTime, i); k++  {
                pObj = Gia_ManCi( p, curCi + k );
                if fSkip != 0 {
                    pObj.Value = 0;
                } else {
                    pObj.Value = Gia_ManAppendCi(pNew);
                }
                Gia_ObjSetTravIdCurrent( p, pObj );
                nBBins++;
                nNewPis += !fSkip;
            }
        
        } else {
            for  k = 0; k < Tim_ManBoxInputNum(pManTime, i); k++  {
                // build logic
                pObj = Gia_ManCo( p, curCo + k );
                Gia_ManDupCollapse_rec( p, Gia_ObjFanin0(pObj), pNew );
                // transfer to the PI
                pObjBox = Gia_ManCi( pBoxes, k );
                pObjBox.Value = Gia_ObjFanin0Copy(pObj);
                Gia_ObjSetTravIdCurrent( pBoxes, pObjBox );
            }
            for  k = 0; k < Tim_ManBoxOutputNum(pManTime, i); k++  {
                // build logic
                pObjBox = Gia_ManCo( pBoxes, curCi - Tim_ManPiNum(pManTime) + k );
                Gia_ManDupCollapse_rec( pBoxes, Gia_ObjFanin0(pObjBox), pNew );
                // transfer to the PI
                pObj = Gia_ManCi( p, curCi + k );
                pObj.Value = Gia_ObjFanin0Copy(pObjBox);
                Gia_ObjSetTravIdCurrent( p, pObj );
            }
        }
        curCo += Tim_ManBoxInputNum(pManTime, i);
        curCi += Tim_ManBoxOutputNum(pManTime, i);
    }
    //printf( "\n" );
    // add remaining nodes
    for  i = Tim_ManCoNum(pManTime) - Tim_ManPoNum(pManTime); i < Tim_ManCoNum(pManTime); i++  {
        pObj = Gia_ManCo( p, i );
        Gia_ManDupCollapse_rec( p, Gia_ObjFanin0(pObj), pNew );
        pObj.Value = Gia_ManAppendCo( pNew, Gia_ObjFanin0Copy(pObj) );
    }
    curCo += Tim_ManPoNum(pManTime);
    // verify counts
    assert(curCi == Gia_ManCiNum(p), "curCi == Gia_ManCiNum(p)");
    assert(curCo == Gia_ManCoNum(p), "curCo == Gia_ManCoNum(p)");
    Gia_ManSetRegNum( pNew, (fSeq && p.vRegClasses) ? Vec_IntSize(p.vRegClasses) : Gia_ManRegNum(p) );
    Gia_ManHashStop( pNew );
    pNew = Gia_ManCleanup( pTemp = pNew );
    Gia_ManCleanupRemap( p, pTemp );
    Gia_ManStop( pTemp );
    if nNewPis != 0 {
        pNew = Gia_ManDupMoveLast( pTemp = pNew, Tim_ManPiNum(pManTime)-Gia_ManRegNum(pNew), nNewPis );
        Gia_ManCleanupRemap( p, pTemp );
        Gia_ManStop( pTemp );
    }
/*
    printf( "%d = %d - %d    Diff = %d\n", 
        Tim_ManPoNum(pManTime),   Gia_ManCoNum(pNew),  nBBouts, 
        Tim_ManPoNum(pManTime) - (Gia_ManCoNum(pNew) - nBBouts) );

    printf( "%d = %d - %d    Diff = %d\n\n", 
        Tim_ManPiNum(pManTime),   Gia_ManCiNum(pNew),  nBBins, 
        Tim_ManPiNum(pManTime) - (Gia_ManCiNum(pNew) - nBBins) );
*/
    assert(vBoxPres != nil || Tim_ManPoNum(pManTime) == Gia_ManCoNum(pNew) - nBBouts, "vBoxPres != nil || Tim_ManPoNum(pManTime) == Gia_ManCoNum(pNew) - nBBouts");
    assert(vBoxPres != nil || Tim_ManPiNum(pManTime) == Gia_ManCiNum(pNew) - nBBins, "vBoxPres != nil || Tim_ManPiNum(pManTime) == Gia_ManCiNum(pNew) - nBBins");
    // implement initial state if given
    if (fSeq != 0) && p.vRegInits && Vec_IntSum(p.vRegInits) {
        // inlining of macro ABC_ALLOC
        var pInit string  = ((char *) malloc(sizeof(char) * (size_t)(Vec_IntSize(p.vRegInits) + 1)));
        var pObj *Gia_Obj_t;
        var i int32;
        assert(Vec_IntSize(p.vRegInits) == Gia_ManRegNum(pNew), "Vec_IntSize(p.vRegInits) == Gia_ManRegNum(pNew)");
        // inlining of macro Gia_ManForEachRo
        for  i = 0; (i < Gia_ManRegNum(pNew)) && ((pObj) = Gia_ManCi(pNew, Gia_ManPiNum(pNew)+i)); i++  {
            if Vec_IntEntry(p.vRegInits, i) == 0 {
                pInit[i] = '0';
            } else if Vec_IntEntry(p.vRegInits, i) == 1 {
                pInit[i] = '1';
            } else { 
                pInit[i] = 'X'; }
        }
        pInit[i] = 0;
        pNew = Gia_ManDupZeroUndc( pTemp = pNew, pInit, 0, 0, 1 );
        pNew.nConstrs = pTemp.nConstrs; pTemp.nConstrs = 0;
        Gia_ManStop( pTemp );
        // inlining of macro ABC_FREE
        pInit = "";
    }
    return pNew;
}

/**Function*************************************************************

  Synopsis    [Verify XAIG against its spec.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManVerifyWithBoxes(pGia *Gia_Man_t, nBTLimit int32, nTimeLim int32, fSeq int32, fDumpFiles int32, fVerbose int32, pFileSpec string) int32 {
    var Status int32    = -1;
    var pSpec, pGia0, pGia1, pMiter *Gia_Man_t;
    var vBoxPres *Vec_Int_t  = nil;
    if pFileSpec == nil && pGia.pSpec == "" {
        fmt.Printf("Spec file is not given. Use standard flow.\n");
        return Status;
    }
    if (Gia_ManBoxNum(pGia) != 0) && pGia.pAigExtra == nil {
        fmt.Printf("Design has no box logic. Use standard flow.\n");
        return Status;
    }
    // read original AIG
    pSpec = Gia_AigerRead( pFileSpec ? pFileSpec : pGia.pSpec, 0, 0, 0 );
    if (Gia_ManBoxNum(pSpec) != 0) && pSpec.pAigExtra == nil {
        Gia_ManStop( pSpec );
        fmt.Printf("Spec has no box logic. Use standard flow.\n");
        return Status;
    }
    // prepare miter
    if pGia.pManTime == nil && pSpec.pManTime == nil {
        pGia0 = Gia_ManDup( pSpec );
        pGia1 = Gia_ManDup( pGia );
    
    } else {
        // if timing managers have different number of black boxes,
        // it is possible that some of the boxes are swept away
        if pSpec.pManTime && Tim_ManBlackBoxNum((Tim_Man_t *)pSpec.pManTime) > 0 && Gia_ManBoxNum(pGia) > 0 {
            // specification cannot have fewer boxes than implementation
            if Gia_ManBoxNum(pSpec) < Gia_ManBoxNum(pGia) {
                fmt.Printf("Spec has less boxes than the design. Cannot proceed.\n");
                return Status;
            }
            // to align the boxes, find what boxes of pSpec are dropped in pGia
            if Gia_ManBoxNum(pSpec) > Gia_ManBoxNum(pGia) {
                vBoxPres = Tim_ManAlignTwo( (Tim_Man_t *)pSpec.pManTime, (Tim_Man_t *)pGia.pManTime );
                if vBoxPres == nil {
                    fmt.Printf("Boxes of spec and design cannot be aligned. Cannot proceed.\n");
                    return Status;
                }
            }
        }
        // collapse two designs
        if Gia_ManBoxNum(pSpec) > 0 {
            pGia0 = Gia_ManDupCollapse( pSpec, pSpec.pAigExtra, vBoxPres, fSeq );
        } else {
            pGia0 = Gia_ManDup( pSpec ); }
        if Gia_ManBoxNum(pGia) > 0 {
            pGia1 = Gia_ManDupCollapse( pGia,  pGia.pAigExtra,  nil, fSeq  );
        } else {
            pGia1 = Gia_ManDup( pGia ); }
        Vec_IntFreeP( &vBoxPres );
    }
    if fDumpFiles != 0 {
        var pFileName0, pFileName1 [1000]byte;
        var pNameGeneric string  = Extra_FileNameGeneric( pFileSpec ? pFileSpec : pGia.pSpec );
        pFileName0 = fmt.Sprintf("%s_spec.aig", pNameGeneric);
        pFileName1 = fmt.Sprintf("%s_impl.aig", pNameGeneric);
        Gia_AigerWrite( pGia0, pFileName0, 0, 0, 0 );
        Gia_AigerWrite( pGia1, pFileName1, 0, 0, 0 );
        // inlining of macro ABC_FREE
        pNameGeneric = "";
        printf( "Dumped two parts of the miter into files \"%s\" and \"%s\".\n", pFileName0, pFileName1 );
    }
    // compute the miter
    if fSeq != 0 {
        pMiter = Gia_ManMiter( pGia0, pGia1, 0, 0, 1, 0, fVerbose );
        if pMiter != nil {
            var pMan *Aig_Man_t;
            var SecPar Fra_Sec_t;
            var pSecPar *Fra_Sec_t = &SecPar;
            Fra_SecSetDefaultParams( pSecPar );
            pSecPar.fRetimeFirst = 0;
            pSecPar.nBTLimit  = nBTLimit;
            pSecPar.TimeLimit = nTimeLim;
            pSecPar.fVerbose  = fVerbose;
            pMan = Gia_ManToAig( pMiter, 0 );
            Gia_ManStop( pMiter );
            Status = Fra_FraigSec( pMan, pSecPar, nil );
            Aig_ManStop( pMan );
        }
    
    } else {
        pMiter = Gia_ManMiter( pGia0, pGia1, 0, 1, 0, 0, fVerbose );
        if pMiter != nil {
            var ParsCec Cec_ParCec_t;
            var pPars *Cec_ParCec_t = &ParsCec;
            Cec_ManCecSetDefaultParams( pPars );
            pPars.nBTLimit  = nBTLimit;
            pPars.TimeLimit = nTimeLim;
            pPars.fVerbose  = fVerbose;
            Status = Cec_ManVerify( pMiter, pPars );
            if pPars.iOutFail >= 0 {
                Abc_Print( 1, "Verification failed for at least one output (%d).\n", pPars.iOutFail ); }
            Gia_ManStop( pMiter );
        }
    }
    Gia_ManStop( pGia0 );
    Gia_ManStop( pGia1 );
    Gia_ManStop( pSpec );
    return Status;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

