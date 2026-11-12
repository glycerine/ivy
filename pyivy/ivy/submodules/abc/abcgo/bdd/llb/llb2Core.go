package llb

import (
	"fmt"
	"math"
	"os"
	"time"
)

/**CFile****************************************************************

  FileName    [llb2Core.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [BDD based reachability.]

  Synopsis    [Core procedure.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: llb2Core.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "llbInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

   
type Llb_Img_t struct {
    pInit      *Aig_Man_t;          // AIG manager
    pAig      *Aig_Man_t;           // AIG manager
    pPars   *Gia_ParLlb_t;          // parameters

    dd      *DdManager;             // BDD manager
    ddG      *DdManager;            // BDD manager
    ddR      *DdManager;            // BDD manager
    vDdMans      *Vec_Ptr_t;        // BDD managers for each partition
    vRings      *Vec_Ptr_t;         // onion rings in ddR

    vDriRefs      *Vec_Int_t;       // driver references
    vVarsCs      *Vec_Int_t;        // cur state variables
    vVarsNs      *Vec_Int_t;        // next state variables

    vCs2Glo      *Vec_Int_t;        // cur state variables into global variables
    vNs2Glo      *Vec_Int_t;        // next state variables into global variables
    vGlo2Cs      *Vec_Int_t;        // global variables into cur state variables
    vGlo2Ns      *Vec_Int_t;        // global variables into next state variables
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Computes cube composed of given variables with given values.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreComputeCube(dd *DdManager, vVars *Vec_Int_t, fUseVarIndex int32, pValues string) *DdNode {
    var bRes, bVar, bTemp *DdNode;
    var i, iVar, Index int32;
    var TimeStop abctime;
    TimeStop = dd.TimeStop; dd.TimeStop = 0;
    bRes = Cudd_ReadOne( dd );   Cudd_Ref( bRes );
    Vec_IntForEachEntry( vVars, Index, i )
    {
        if fUseVarIndex != 0 {
            iVar = Index;
        } else {
            iVar = i;
        }
        // inlining of macro Cudd_NotCond
        bVar  = ((DdNode *)((ptrint)(Cudd_bddIthVar(dd, iVar)) ^ (int32((pValues == nil || pValues[i] != 1)))));
        bRes  = Cudd_bddAnd( dd, bTemp = bRes, bVar );  Cudd_Ref( bRes );
        Cudd_RecursiveDeref( dd, bTemp );
    }
    Cudd_Deref( bRes );
    dd.TimeStop = TimeStop;
    return bRes;
}

/**Function*************************************************************

  Synopsis    [Derives counter-example by backward reachability.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreDeriveCex(p *Llb_Img_t) *Abc_Cex_t {
    var pCex *Abc_Cex_t;
    var pObj *Aig_Obj_t;
    var vSupps, vQuant0, vQuant1 *Vec_Ptr_t;
    var bState *DdNode = nil;
    var bImage *DdNode;
    var bOneCube *DdNode;
    var bTemp *DdNode;
    var bRing *DdNode;
    var i, v, RetValue, nPiOffset int32;
    var pValues string  = ABC_ALLOC( char, Cudd_ReadSize(p.ddR) );
    assert(Vec_PtrSize(p.vRings) > 0, "Vec_PtrSize(p.vRings) > 0");

    p.dd.TimeStop  = 0;
    p.ddR.TimeStop = 0;

    // get supports and quantified variables
    Vec_PtrReverseOrder( p.vDdMans );
    vSupps = Llb_ImgSupports( p.pAig, p.vDdMans, p.vVarsNs, p.vVarsCs, 1, 0 );
    Llb_ImgSchedule( vSupps, &vQuant0, &vQuant1, 0 );
    Vec_VecFree( (Vec_Vec_t *)vSupps );
    Llb_ImgQuantifyReset( p.vDdMans );
//    Llb_ImgQuantifyFirst( p->pAig, p->vDdMans, vQuant0 );

    // allocate room for the counter-example
    pCex = Abc_CexAlloc( Saig_ManRegNum(p.pAig), Saig_ManPiNum(p.pAig), Vec_PtrSize(p.vRings) );
    pCex.iFrame = Vec_PtrSize(p.vRings) - 1;
    pCex.iPo = -1;

    // get the last cube
    bOneCube = Cudd_bddIntersect( p.ddR, (DdNode *)Vec_PtrEntryLast(p.vRings), p.ddR.bFunc );  Cudd_Ref( bOneCube );
    RetValue = Cudd_bddPickOneCube( p.ddR, bOneCube, pValues );
    Cudd_RecursiveDeref( p.ddR, bOneCube );
    assert((RetValue != 0), "RetValue");

    // write PIs of counter-example
    nPiOffset = Saig_ManRegNum(p.pAig) + Saig_ManPiNum(p.pAig) * (Vec_PtrSize(p.vRings) - 1);
    Saig_ManForEachPi( p.pAig, pObj, i )
        if ( pValues[Saig_ManRegNum(p.pAig)+i] == 1 )
            Abc_InfoSetBit( pCex.pData, nPiOffset + i );

    // write state in terms of NS variables
    if Vec_PtrSize(p.vRings) > 1 {
        bState = Llb_CoreComputeCube( p.dd, p.vVarsNs, 1, pValues );   Cudd_Ref( bState );
    }
    // perform backward analysis
    Vec_PtrForEachEntryReverse( DdNode *, p.vRings, bRing, v )
    { 
        if v == Vec_PtrSize(p.vRings) - 1 {
            continue; }
        // compute the next states
        bImage = Llb_ImgComputeImage( p.pAig, p.vDdMans, p.dd, bState, 
            vQuant0, vQuant1, p.vDriRefs, p.pPars.TimeTarget, 1, 0, 0 );
        assert(bImage != nil, "bImage != nil");
        Cudd_Ref( bImage );
        Cudd_RecursiveDeref( p.dd, bState );
//Extra_bddPrintSupport( p->dd, bImage ); printf( "\n" );

        // move reached states into ring manager
        bImage = Extra_TransferPermute( p.dd, p.ddR, bTemp = bImage, Vec_IntArray(p.vCs2Glo) );    Cudd_Ref( bImage );
        Cudd_RecursiveDeref( p.dd, bTemp );

        // intersect with the previous set
        bOneCube = Cudd_bddIntersect( p.ddR, bImage, bRing );                Cudd_Ref( bOneCube );
        Cudd_RecursiveDeref( p.ddR, bImage );

        // find any assignment of the BDD
        RetValue = Cudd_bddPickOneCube( p.ddR, bOneCube, pValues );
        Cudd_RecursiveDeref( p.ddR, bOneCube );
        assert((RetValue != 0), "RetValue");

        // write PIs of counter-example
        nPiOffset -= Saig_ManPiNum(p.pAig);
        Saig_ManForEachPi( p.pAig, pObj, i )
            if ( pValues[Saig_ManRegNum(p.pAig)+i] == 1 )
                Abc_InfoSetBit( pCex.pData, nPiOffset + i );

        // check that we get the init state
        if v == 0 {
            Saig_ManForEachLo( p.pAig, pObj, i )
                assert(pValues[i] == 0, "pValues[i] == 0");
            break;
        }

        // write state in terms of NS variables
        bState = Llb_CoreComputeCube( p.dd, p.vVarsNs, 1, pValues );   Cudd_Ref( bState );
    }
    assert(nPiOffset == Saig_ManRegNum(p.pAig), "nPiOffset == Saig_ManRegNum(p.pAig)");
    // update the output number
    RetValue = Saig_ManFindFailedPoCex( p.pInit, pCex );
    assert(RetValue >= 0 && RetValue < Saig_ManPoNum(p.pInit), "RetValue >= 0 && RetValue < Saig_ManPoNum(p.pInit)"); // invalid CEX!!!
    pCex.iPo = RetValue;
    // cleanup
    ABC_FREE( pValues );
    Vec_VecFree( (Vec_Vec_t *)vQuant0 );
    Vec_VecFree( (Vec_Vec_t *)vQuant1 );
    return pCex;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreReachability_int(p *Llb_Img_t, vQuant0 *Vec_Ptr_t, vQuant1 *Vec_Ptr_t) int32 {
    var pLoc2Glo *int32;
    if p.pPars.fBackward {
        pLoc2Glo = Vec_IntArray( p.vCs2Glo );
    } else {
        pLoc2Glo = Vec_IntArray( p.vNs2Glo );
    }
    var pLoc2GloR *int32;
    if p.pPars.fBackward {
        pLoc2GloR = Vec_IntArray( p.vNs2Glo );
    } else {
        pLoc2GloR = Vec_IntArray( p.vCs2Glo );
    }
    var pGlo2Loc *int32;
    if p.pPars.fBackward {
        pGlo2Loc = Vec_IntArray( p.vGlo2Ns );
    } else {
        pGlo2Loc = Vec_IntArray( p.vGlo2Cs );
    }
    var bCurrent, bReached, bNext, bTemp *DdNode;
    var clk2 abctime;
    var clk abctime = Abc_Clock();
    var nIters, nBddSize int32;//, iOutFail = -1;
/*
    // compute time to stop
    if ( p->pPars->TimeLimit )
        p->pPars->TimeTarget = Abc_Clock() + p->pPars->TimeLimit * CLOCKS_PER_SEC;
    else
        p->pPars->TimeTarget = 0;
*/

    if Abc_Clock() > p.pPars.TimeTarget {
        if !p.pPars.fSilent {
            fmt.Printf("Reached timeout (%d seconds) before image computation.\n", p.pPars.TimeLimit); }
        p.pPars.iFrame = -1;
        return -1;
    }

    // set the stop time parameter
    p.dd.TimeStop  = p.pPars.TimeTarget;
    p.ddG.TimeStop = p.pPars.TimeTarget;
    p.ddR.TimeStop = p.pPars.TimeTarget;

    // compute initial states
    if p.pPars.fBackward {
        // create init state in the global manager
        bTemp = Llb_BddComputeBad( p.pInit, p.ddR, p.pPars.TimeTarget );   
        if bTemp == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) while computing bad states.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = -1;
            return -1;
        }
        Cudd_Ref( bTemp );
        // create bad state in the ring manager
        p.ddR.bFunc = Llb_CoreComputeCube( p.ddR, p.vVarsCs, 0, nil );      Cudd_Ref( p.ddR.bFunc );
        bCurrent = Llb_BddQuantifyPis( p.pInit, p.ddR, bTemp );                Cudd_Ref( bCurrent );
        Cudd_RecursiveDeref( p.ddR, bTemp );
        bReached = Cudd_bddTransfer( p.ddR, p.ddG, bCurrent );                 Cudd_Ref( bReached );
        Cudd_RecursiveDeref( p.ddR, bCurrent );
        // move init state to the working manager
        bCurrent = Extra_TransferPermute( p.ddG, p.dd, bReached, pGlo2Loc );   
        if bCurrent == nil {
            Cudd_RecursiveDeref( p.ddG, bReached );
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during transfer 0.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = -1;
            return -1;
        }
        Cudd_Ref( bCurrent );
    
    } else {
        // create bad state in the ring manager
        p.ddR.bFunc = Llb_BddComputeBad( p.pInit, p.ddR, p.pPars.TimeTarget );  
        if p.ddR.bFunc == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) while computing bad states.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = -1;
            return -1;
        }
        Cudd_Ref( p.ddR.bFunc );
        // create init state in the working and global manager
        bCurrent = Llb_CoreComputeCube( p.dd,  p.vVarsCs, 1, nil );           Cudd_Ref( bCurrent );
        bReached = Llb_CoreComputeCube( p.ddG, p.vVarsCs, 0, nil );           Cudd_Ref( bReached );
//Extra_bddPrint( p->dd, bCurrent );  printf( "\n" );
//Extra_bddPrint( p->ddG, bReached );  printf( "\n" );
    }

    // compute onion rings
    for  nIters = 0; nIters < p.pPars.nIterMax; nIters++  { 
        clk2 = Abc_Clock();
        // check the runtime limit
        if p.pPars.TimeLimit && Abc_Clock() > p.pPars.TimeTarget {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.dd,  bCurrent );  bCurrent = nil;
            Cudd_RecursiveDeref( p.ddG, bReached );  bReached = nil;
            return -1;
        }

        // save the onion ring
        bTemp = Extra_TransferPermute( p.dd, p.ddR, bCurrent, pLoc2GloR );  
        if bTemp == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.dd,  bCurrent );  bCurrent = nil;
            Cudd_RecursiveDeref( p.ddG, bReached );  bReached = nil;
            return -1;
        }
        Cudd_Ref( bTemp );
        Vec_PtrPush( p.vRings, bTemp );

        // check it for bad states
        // inlining of macro Cudd_Not
        if !p.pPars.fSkipOutCheck && !Cudd_bddLeq( p.ddR, bTemp, ((DdNode *)((ptrint)(p.ddR.bFunc) ^ 01)) ) {
            assert(p.pInit.pSeqModel == nil, "p.pInit.pSeqModel == nil");
            if !p.pPars.fBackward {
                p.pInit.pSeqModel = Llb_CoreDeriveCex( p ); } 
            Cudd_RecursiveDeref( p.dd,  bCurrent );  bCurrent = nil;
            Cudd_RecursiveDeref( p.ddG, bReached );  bReached = nil;
            if !p.pPars.fSilent {
                if !p.pPars.fBackward {
                    Abc_Print( 1, "Output %d of miter \"%s\" was asserted in frame %d.  ", p.pInit.pSeqModel.iPo, p.pInit.pName, nIters );
                } else {
                    Abc_Print( 1, "Output ??? was asserted in frame %d (counter-example is not produced).  ", nIters ); }
                Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
            }
            p.pPars.iFrame = nIters - 1;
            return 0;
        }

        // compute the next states
        bNext = Llb_ImgComputeImage( p.pAig, p.vDdMans, p.dd, bCurrent, 
            vQuant0, vQuant1, p.vDriRefs, p.pPars.TimeTarget, 
            p.pPars.fBackward, p.pPars.fReorder, p.pPars.fVeryVerbose );
        if bNext == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.dd,  bCurrent );   bCurrent = nil;
            Cudd_RecursiveDeref( p.ddG, bReached );   bReached = nil;
            return -1;
        }
        Cudd_Ref( bNext );
        Cudd_RecursiveDeref( p.dd, bCurrent );        bCurrent = nil;
//Extra_bddPrintSupport( p->dd, bNext ); printf( "\n" );

        // remap these states into the global manager
//        bNext = Extra_TransferPermute( p->dd, p->ddG, bTemp = bNext, pLoc2Glo );    Cudd_Ref( bNext );
//        Cudd_RecursiveDeref( p->dd, bTemp );

//        bNext = Extra_TransferPermuteTime( p->dd, p->ddG, bTemp = bNext, pLoc2Glo, p->pPars->TimeTarget );    
        bNext = Extra_TransferPermute( p.dd, p.ddG, bTemp = bNext, pLoc2Glo );    
        if bNext == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 1.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.dd,  bTemp );  
            Cudd_RecursiveDeref( p.ddG, bReached );   bReached = nil;
            return -1;
        }
        Cudd_Ref( bNext );
        Cudd_RecursiveDeref( p.dd, bTemp );  

        nBddSize = Cudd_DagSize(bNext);
        // check if there are any new states
        if Cudd_bddLeq( p.ddG, bNext, bReached ) { // implication = no new states
        
            Cudd_RecursiveDeref( p.ddG,  bNext );     bNext = nil;
            break;
        }

        // get the new states
        // inlining of macro Cudd_Not
        bCurrent = Cudd_bddAnd( p.ddG, bNext, ((DdNode *)((ptrint)(bReached) ^ 01)) );                    
        if bCurrent == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 2.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.ddG, bNext );  
            Cudd_RecursiveDeref( p.ddG, bReached );   bReached = nil;
            return -1;
        }
        Cudd_Ref( bCurrent );

        // remap these states into the current state vars
//        bCurrent = Extra_TransferPermute( p->ddG, p->dd, bTemp = bCurrent, pGlo2Loc );   Cudd_Ref( bCurrent );
//        Cudd_RecursiveDeref( p->ddG, bTemp );

//        bCurrent = Extra_TransferPermuteTime( p->ddG, p->dd, bTemp = bCurrent, pGlo2Loc, p->pPars->TimeTarget );    
        bCurrent = Extra_TransferPermute( p.ddG, p.dd, bTemp = bCurrent, pGlo2Loc );    
        if bCurrent == nil {
            if !p.pPars.fSilent {
                fmt.Printf("Reached timeout (%d seconds) during image computation in transfer 2.\n", p.pPars.TimeLimit); }
            p.pPars.iFrame = nIters - 1;
            Cudd_RecursiveDeref( p.ddG, bTemp );  
            Cudd_RecursiveDeref( p.ddG, bReached );   bReached = nil;
            return -1;
        }
        Cudd_Ref( bCurrent );
        Cudd_RecursiveDeref( p.ddG, bTemp );

        // add to the reached states
        bReached = Cudd_bddOr( p.ddG, bTemp = bReached, bNext );                       Cudd_Ref( bReached );
        Cudd_RecursiveDeref( p.ddG, bTemp );
        Cudd_RecursiveDeref( p.ddG, bNext );
        bNext = nil;

        if p.pPars.fVeryVerbose {
            var nMints float64  = Cudd_CountMinterm(p.ddG, bReached, Saig_ManRegNum(p.pAig) );
//            Extra_bddPrint( p->ddG, bReached );printf( "\n" );
            fmt.Fprintf(os.Stdout, "        Reachable states = %.0f. (Ratio = %.4f %%)\n", nMints, 100.0*nMints/math.Pow(float64(2.0), float64(Saig_ManRegNum(p.pAig))));
            fflush( os.Stdout ); 
        }
        if p.pPars.fVerbose {
            fmt.Fprintf(os.Stdout, "F =%3d : ", nIters);
            fmt.Fprintf(os.Stdout, "Image =%6d  ", nBddSize);
            fmt.Fprintf(os.Stdout, "(%4d%4d)  ", Cudd_ReadReorderings(p.dd), Cudd_ReadGarbageCollections(p.dd));
            fmt.Fprintf(os.Stdout, "Reach =%6d  ", Cudd_DagSize(bReached));
            fmt.Fprintf(os.Stdout, "(%4d%4d)  ", Cudd_ReadReorderings(p.ddG), Cudd_ReadGarbageCollections(p.ddG));
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk2 );
        }

        // check timeframe limit
        if nIters == p.pPars.nIterMax - 1 {
            if !p.pPars.fSilent {
                fmt.Printf("Reached limit on the number of timeframes (%d).\n", p.pPars.nIterMax); }
            p.pPars.iFrame = nIters;
            Cudd_RecursiveDeref( p.dd,  bCurrent );  bCurrent = nil;
            Cudd_RecursiveDeref( p.ddG, bReached );  bReached = nil;
            return -1;
        }
    }
    if bReached == nil {
        p.pPars.iFrame = nIters - 1;
        return 0; // reachable
    }
    if bCurrent {
        Cudd_RecursiveDeref( p.dd, bCurrent ); }
    // report the stats
    if p.pPars.fVerbose {
        var nMints float64  = Cudd_CountMinterm(p.ddG, bReached, Saig_ManRegNum(p.pAig) );
        if nIters >= p.pPars.nIterMax {
            fmt.Fprintf(os.Stdout, "Reachability analysis is stopped after %d frames.\n", nIters);
        } else {
            fmt.Fprintf(os.Stdout, "Reachability analysis completed after %d frames.\n", nIters); }
        fmt.Fprintf(os.Stdout, "Reachable states = %.0f. (Ratio = %.4f %%)\n", nMints, 100.0*nMints/math.Pow(float64(2.0), float64(Saig_ManRegNum(p.pAig))));
        fflush( os.Stdout ); 
    }
    if p.pPars.fDumpReached {
        Llb_ManDumpReached( p.ddG, bReached, p.pAig.pName, "reached.blif" );
        fmt.Printf("Reached states with %d BDD nodes are dumpted into file \"reached.blif\".\n", Cudd_DagSize(bReached));
    }
    Cudd_RecursiveDeref( p.ddG, bReached );
    if nIters >= p.pPars.nIterMax {
        if !p.pPars.fSilent {
            fmt.Printf("Verified only for states reachable in %d frames.  ", nIters);
            Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
        }
        p.pPars.iFrame = p.pPars.nIterMax;
        return -1; // undecided
    }
    if !p.pPars.fSilent {
        fmt.Printf("The miter is proved unreachable after %d iterations.  ", nIters);
        Abc_PrintTime( 1, "Time", Abc_Clock() - clk );
    }
    p.pPars.iFrame = nIters - 1;
    return 1; // unreachable
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreReachability(p *Llb_Img_t) int32 {
    var vSupps, vQuant0, vQuant1 *Vec_Ptr_t;
    var RetValue int32;
    // get supports and quantified variables
    if p.pPars.fBackward {
        Vec_PtrReverseOrder( p.vDdMans );
        vSupps = Llb_ImgSupports( p.pAig, p.vDdMans, p.vVarsNs, p.vVarsCs, 0, p.pPars.fVeryVerbose );
    
    } else {
        vSupps = Llb_ImgSupports( p.pAig, p.vDdMans, p.vVarsCs, p.vVarsNs, 0, p.pPars.fVeryVerbose ); }
    Llb_ImgSchedule( vSupps, &vQuant0, &vQuant1, p.pPars.fVeryVerbose );
    Vec_VecFree( (Vec_Vec_t *)vSupps );
    // remove variables
    Llb_ImgQuantifyFirst( p.pAig, p.vDdMans, vQuant0, p.pPars.fVeryVerbose );
    // perform reachability
    RetValue = Llb_CoreReachability_int( p, vQuant0, vQuant1 );
    Vec_VecFree( (Vec_Vec_t *)vQuant0 );
    Vec_VecFree( (Vec_Vec_t *)vQuant1 );
    return RetValue;
}


/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreConstructAll(p *Aig_Man_t, vResult *Vec_Ptr_t, vVarsNs *Vec_Int_t, TimeTarget abctime) *Vec_Ptr_t {
    var dd *DdManager;
    var vDdMans *Vec_Ptr_t;
    var vLower *Vec_Ptr_t;
    var vUpper *Vec_Ptr_t = nil;
    var i int32;
    vDdMans = Vec_PtrStart( Vec_PtrSize(vResult) );
    Vec_PtrForEachEntryReverse( Vec_Ptr_t *, vResult, vLower, i )
    {
        if i < Vec_PtrSize(vResult) - 1 {
            dd = Llb_ImgPartition( p, vLower, vUpper, TimeTarget );
        } else {
            dd = Llb_DriverLastPartition( p, vVarsNs, TimeTarget ); }
        if dd == nil {
            Vec_PtrForEachEntry( DdManager *, vDdMans, dd, i )
            {
                if dd == nil {
                    continue; }
                if dd.bFunc {
                    Cudd_RecursiveDeref( dd, dd.bFunc ); }
                Extra_StopManager( dd );
            }
            Vec_PtrFree( vDdMans );
            return nil;
        }
        Vec_PtrWriteEntry( vDdMans, i, dd );
        vUpper = vLower;
    }
    return vDdMans;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreSetVarMaps(p *Llb_Img_t) {
    var pObj *Aig_Obj_t;
    var i, iVarCs, iVarNs int32;
    assert(p.vVarsCs != nil, "p.vVarsCs != nil");
    assert(p.vVarsNs != nil, "p.vVarsNs != nil");
    assert(p.vCs2Glo == nil, "p.vCs2Glo == nil");
    assert(p.vNs2Glo == nil, "p.vNs2Glo == nil");
    assert(p.vGlo2Cs == nil, "p.vGlo2Cs == nil");
    assert(p.vGlo2Ns == nil, "p.vGlo2Ns == nil");
    p.vCs2Glo = Vec_IntStartFull( Aig_ManObjNumMax(p.pAig) );
    p.vNs2Glo = Vec_IntStartFull( Aig_ManObjNumMax(p.pAig) );
    p.vGlo2Cs = Vec_IntStartFull( Aig_ManRegNum(p.pAig) );
    p.vGlo2Ns = Vec_IntStartFull( Aig_ManRegNum(p.pAig) );
    for  i = 0; i < Aig_ManRegNum(p.pAig); i++  {
        iVarCs = Vec_IntEntry( p.vVarsCs, i );
        iVarNs = Vec_IntEntry( p.vVarsNs, i );
        assert(iVarCs >= 0 && iVarCs < Aig_ManObjNumMax(p.pAig), "iVarCs >= 0 && iVarCs < Aig_ManObjNumMax(p.pAig)");
        assert(iVarNs >= 0 && iVarNs < Aig_ManObjNumMax(p.pAig), "iVarNs >= 0 && iVarNs < Aig_ManObjNumMax(p.pAig)");
        Vec_IntWriteEntry( p.vCs2Glo, iVarCs, i );
        Vec_IntWriteEntry( p.vNs2Glo, iVarNs, i );
        Vec_IntWriteEntry( p.vGlo2Cs, i, iVarCs );
        Vec_IntWriteEntry( p.vGlo2Ns, i, iVarNs );
    }
    // add mapping of the PIs
    Saig_ManForEachPi( p.pAig, pObj, i )
        Vec_IntWriteEntry( p.vCs2Glo, Aig_ObjId(pObj), Aig_ManRegNum(p.pAig)+i );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreStart(pInit *Aig_Man_t, pAig *Aig_Man_t, pPars *Gia_ParLlb_t) *Llb_Img_t {
    var p *Llb_Img_t;
    p = new(Llb_Img_t);
    p.pInit = pInit;
    p.pAig  = pAig;
    p.pPars = pPars;
    p.dd    = Cudd_Init( Aig_ManObjNumMax(pAig), 0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    p.ddG   = Cudd_Init( Aig_ManRegNum(pAig),    0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    p.ddR   = Cudd_Init( Aig_ManCiNum(pAig),     0, CUDD_UNIQUE_SLOTS, CUDD_CACHE_SLOTS, 0 );
    Cudd_AutodynEnable( p.dd,  CUDD_REORDER_SYMM_SIFT );
    Cudd_AutodynEnable( p.ddG, CUDD_REORDER_SYMM_SIFT );
    Cudd_AutodynEnable( p.ddR, CUDD_REORDER_SYMM_SIFT );
    p.vRings = Vec_PtrAlloc( 100 );
    p.vDriRefs = Llb_DriverCountRefs( pAig );
    p.vVarsCs  = Llb_DriverCollectCs( pAig );
    p.vVarsNs  = Llb_DriverCollectNs( pAig, p.vDriRefs );
    Llb_CoreSetVarMaps( p );
    return p;
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreStop(p *Llb_Img_t) {
    var dd *DdManager;
    var bTemp *DdNode;
    var i int32;
    if ( p.vDdMans )
    Vec_PtrForEachEntry( DdManager *, p.vDdMans, dd, i )
    {
        if dd.bFunc {
            Cudd_RecursiveDeref( dd, dd.bFunc ); }
        if dd.bFunc2 {
            Cudd_RecursiveDeref( dd, dd.bFunc2 ); }
        Extra_StopManager( dd );
    }
    Vec_PtrFreeP( &p.vDdMans );
    if p.ddR.bFunc {
        Cudd_RecursiveDeref( p.ddR, p.ddR.bFunc ); }
    Vec_PtrForEachEntry( DdNode *, p.vRings, bTemp, i )
        Cudd_RecursiveDeref( p.ddR, bTemp );
    Vec_PtrFree( p.vRings );
    Extra_StopManager( p.dd );
    Extra_StopManager( p.ddG );
    Extra_StopManager( p.ddR );
    Vec_IntFreeP( &p.vDriRefs );
    Vec_IntFreeP( &p.vVarsCs );
    Vec_IntFreeP( &p.vVarsNs );
    Vec_IntFreeP( &p.vCs2Glo );
    Vec_IntFreeP( &p.vNs2Glo );
    Vec_IntFreeP( &p.vGlo2Cs );
    Vec_IntFreeP( &p.vGlo2Ns );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_CoreExperiment(pInit *Aig_Man_t, pAig *Aig_Man_t, pPars *Gia_ParLlb_t, vResult *Vec_Ptr_t, TimeTarget abctime) int32 {
    var RetValue int32;
    var p *Llb_Img_t;
//    printf( "\n" );
//    pPars->fVerbose = 1;
    p = Llb_CoreStart( pInit, pAig, pPars );
    p.vDdMans = Llb_CoreConstructAll( pAig, vResult, p.vVarsNs, TimeTarget );
    if p.vDdMans == nil {
        if !pPars.fSilent {
            fmt.Printf("Reached timeout (%d seconds) while deriving the partitions.\n", pPars.TimeLimit); }
        Llb_CoreStop( p );
        return -1;
    }
    RetValue = Llb_CoreReachability( p );
    Llb_CoreStop( p );
    return RetValue;
}

/**Function*************************************************************

  Synopsis    [Finds balanced cut.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Llb_ManReachMinCut(pAig *Aig_Man_t, pPars *Gia_ParLlb_t) int32 {
    extern Vec_Ptr_t * Llb_ManComputeCuts( Aig_Man_t * p, int Num, int fVerbose, int fVeryVerbose );
    var vResult *Vec_Ptr_t;
    var p *Aig_Man_t;
    var RetValue int32  = -1;
    var clk abctime  = Abc_Clock();

    // compute time to stop
    if pPars.TimeLimit {
        pPars.TimeTarget = pPars.TimeLimit * int64(time.Second) + Abc_Clock();
    } else {
        pPars.TimeTarget = 0;
    }

    p = Aig_ManDupFlopsOnly( pAig );
//Aig_ManShow( p, 0, NULL );
    if pPars.fVerbose {
    Aig_ManPrintStats( pAig ); }
    if pPars.fVerbose {
    Aig_ManPrintStats( p ); }
    Aig_ManFanoutStart( p );

    vResult = Llb_ManComputeCuts( p, pPars.nPartValue, pPars.fVerbose, pPars.fVeryVerbose );

    if pPars.TimeLimit && Abc_Clock() > pPars.TimeTarget {
        if !pPars.fSilent {
            fmt.Printf("Reached timeout (%d seconds) after partitioning.\n", pPars.TimeLimit); }

        Vec_VecFree( (Vec_Vec_t *)vResult );
        Aig_ManFanoutStop( p );
        Aig_ManCleanMarkAB( p );
        Aig_ManStop( p );
        return RetValue;
    }

    if !pPars.fSkipReach {
        RetValue = Llb_CoreExperiment( pAig, p, pPars, vResult, pPars.TimeTarget ); }

    Vec_VecFree( (Vec_Vec_t *)vResult );
    Aig_ManFanoutStop( p );
    Aig_ManCleanMarkAB( p );
    Aig_ManStop( p );

    if RetValue == -1 {
        Abc_PrintTime( 1, "Total runtime of the min-cut-based reachability engine", Abc_Clock() - clk ); }
    return RetValue;
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

