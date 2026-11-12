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
/**Function*************************************************************

  Synopsis    [Computes AIG with boxes.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
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

