package bdc

import "fmt"

/**CFile****************************************************************

  FileName    [bdcCore.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Truth-table-based bi-decomposition engine.]

  Synopsis    [The gateway to bi-decomposition.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 30, 2007.]

  Revision    [$Id: bdcCore.c,v 1.00 2007/01/30 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "bdcInt.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Accessing contents of the node.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_ManFunc(p *Bdc_Man_t, i int32) *Bdc_Fun_t { return Bdc_FunWithId(p, i); }
func Bdc_ManRoot(p *Bdc_Man_t) *Bdc_Fun_t { return p.pRoot;            }
func Bdc_ManNodeNum(p *Bdc_Man_t) int32 { return p.nNodes;           }
func Bdc_ManAndNum(p *Bdc_Man_t) int32 { return p.nNodes-p.nVars-1;}
func Bdc_FuncFanin0(p *Bdc_Fun_t) *Bdc_Fun_t { return p.pFan0;            }
func Bdc_FuncFanin1(p *Bdc_Fun_t) *Bdc_Fun_t { return p.pFan1;            }
func Bdc_FuncCopy(p *Bdc_Fun_t) any { return p.pCopy;            }
func Bdc_FuncCopyInt(p *Bdc_Fun_t) int32 { return p.iCopy;            }
func Bdc_FuncSetCopy(p *Bdc_Fun_t, pCopy any) { p.pCopy = pCopy;           }
func Bdc_FuncSetCopyInt(p *Bdc_Fun_t, iCopy int32) { p.iCopy = iCopy;           }

/**Function*************************************************************

  Synopsis    [Allocate resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_ManAlloc(pPars *Bdc_Par_t) *Bdc_Man_t {
    var p *Bdc_Man_t;
    p = new(Bdc_Man_t);
    memset( p, 0, sizeof(Bdc_Man_t) );
    assert(pPars.nVarsMax > 1 && pPars.nVarsMax < 16, "pPars.nVarsMax > 1 && pPars.nVarsMax < 16");
    p.pPars = pPars;
    p.nWords = Kit_TruthWordNum( pPars.nVarsMax );
    p.nDivsLimit = 200;
    // internal nodes
    p.nNodesAlloc = 512;
    p.pNodes = ABC_ALLOC( Bdc_Fun_t, p.nNodesAlloc );
    // memory
    p.vMemory = Vec_IntStart( 8 * p.nWords * p.nNodesAlloc );
    Vec_IntClear(p.vMemory);
    // set up hash table
    p.nTableSize = (1 << p.pPars.nVarsMax);
    p.pTable = ABC_ALLOC( Bdc_Fun_t *, p.nTableSize );
    memset( p.pTable, 0, sizeof(Bdc_Fun_t *) * p.nTableSize );
    p.vSpots = Vec_IntAlloc( 256 );
    // truth tables
    p.vTruths = Vec_PtrAllocTruthTables( p.pPars.nVarsMax );
    p.puTemp1 = ABC_ALLOC( unsigned, 4 * p.nWords );
    p.puTemp2 = p.puTemp1 + p.nWords;
    p.puTemp3 = p.puTemp2 + p.nWords;
    p.puTemp4 = p.puTemp3 + p.nWords;
    // start the internal ISFs
    p.pIsfOL = &p.IsfOL;  Bdc_IsfStart( p, p.pIsfOL );
    p.pIsfOR = &p.IsfOR;  Bdc_IsfStart( p, p.pIsfOR );
    p.pIsfAL = &p.IsfAL;  Bdc_IsfStart( p, p.pIsfAL );
    p.pIsfAR = &p.IsfAR;  Bdc_IsfStart( p, p.pIsfAR );   
    return p; 
}

/**Function*************************************************************

  Synopsis    [Deallocate resynthesis manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_ManFree(p *Bdc_Man_t) {
    if p.pPars.fVerbose {
        fmt.Printf("Bi-decomposition stats: Calls = %d.  Nodes = %d. Reuse = %d.\n", p.numCalls, p.numNodes, p.numReuse);
        fmt.Printf("ANDs = %d.  ORs = %d.  Weak = %d.  Muxes = %d.  Memory = %.2f K\n", p.numAnds, p.numOrs, p.numWeaks, p.numMuxes, 4.0 * Vec_IntSize(p.vMemory) / (1<<10));
        ABC_PRT( "Cache", p.timeCache );
        ABC_PRT( "Check", p.timeCheck );
        ABC_PRT( "Muxes", p.timeMuxes );
        ABC_PRT( "Supps", p.timeSupps );
        ABC_PRT( "TOTAL", p.timeTotal );
    }
    Vec_IntFree( p.vMemory );
    Vec_IntFree( p.vSpots );
    Vec_PtrFree( p.vTruths );
    ABC_FREE( p.puTemp1 );
    ABC_FREE( p.pNodes );
    ABC_FREE( p.pTable );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    [Clears the manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_ManPrepare(p *Bdc_Man_t, vDivs *Vec_Ptr_t) {
    var puTruth *uint32;
    var pNode *Bdc_Fun_t;
    var i int32;
    Bdc_TableClear( p );
    Vec_IntClear( p.vMemory );
    // add constant 1 and elementary vars
    p.nNodes = 0;
    p.nNodesNew = - 1 - p.nVars - (vDivs? Vec_PtrSize(vDivs) : 0);
    // add constant 1
    pNode = Bdc_FunNew( p );
    pNode.Type = BDC_TYPE_CONST1;
    pNode.puFunc = (unsigned *)Vec_IntFetch(p.vMemory, p.nWords); 
    Kit_TruthFill( pNode.puFunc, p.nVars );
    pNode.uSupp = 0;
    Bdc_TableAdd( p, pNode );
    // add variables
    for  i = 0; i < p.nVars; i++  {
        pNode = Bdc_FunNew( p );
        pNode.Type = BDC_TYPE_PI;
        pNode.puFunc = (unsigned *)Vec_PtrEntry( p.vTruths, i );
        pNode.uSupp = (1 << i);
        Bdc_TableAdd( p, pNode );
    }
    // add the divisors
    if ( vDivs )
    Vec_PtrForEachEntry( unsigned *, vDivs, puTruth, i )
    {
        pNode = Bdc_FunNew( p );
        pNode.Type = BDC_TYPE_PI;
        pNode.puFunc = puTruth;
        pNode.uSupp = Kit_TruthSupport( puTruth, p.nVars );
        Bdc_TableAdd( p, pNode );
        if i == p.nDivsLimit {
            break; }
    }
    assert(p.nNodesNew == 0, "p.nNodesNew == 0");
}

/**Function*************************************************************

  Synopsis    [Prints bi-decomposition in a simple format.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_ManDecPrintSimple(p *Bdc_Man_t) {
    var pNode *Bdc_Fun_t;
    var i int32;
    fmt.Printf(" 0 : Const 1\n");
    for  i = 1; i < p.nNodes; i++  {
        fmt.Printf(" %d : ", i);
        pNode = p.pNodes + i;
        if pNode.Type == BDC_TYPE_PI {
            fmt.Printf("PI   ");
        } else {
            fmt.Printf("%s%d &", Bdc_IsComplement(pNode.pFan0)? "-":"", Bdc_FunId(p,Bdc_Regular(pNode.pFan0)));
            fmt.Printf(" %s%d   ", Bdc_IsComplement(pNode.pFan1)? "-":"", Bdc_FunId(p,Bdc_Regular(pNode.pFan1)));
        }
//        Extra_PrintBinary( stdout, pNode->puFunc, (1<<p->nVars) );
        fmt.Printf("\n");
    }
    fmt.Printf("Root = %s%d.\n", Bdc_IsComplement(p.pRoot)? "-":"", Bdc_FunId(p,Bdc_Regular(p.pRoot)));
}

/**Function*************************************************************

  Synopsis    [Prints bi-decomposition recursively.]

  Description [This procedure prints bi-decomposition as a factored form.
  In doing so, logic sharing, if present, will be replicated several times.]
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_ManDecPrint_rec(p *Bdc_Man_t, pNode *Bdc_Fun_t) {
    if pNode.Type == BDC_TYPE_PI {
        fmt.Printf("%c", 'a' + Bdc_FunId(p,pNode) - 1);
    } else if pNode.Type == BDC_TYPE_AND {
        var pNode0 *Bdc_Fun_t  = Bdc_FuncFanin0( pNode );
        var pNode1 *Bdc_Fun_t  = Bdc_FuncFanin1( pNode );

        if Bdc_IsComplement(pNode0) {
            fmt.Printf("!"); }
        if Bdc_IsComplement(pNode0) && Bdc_Regular(pNode0).Type != BDC_TYPE_PI {
            fmt.Printf("("); }
        Bdc_ManDecPrint_rec( p, Bdc_Regular(pNode0) );
        if Bdc_IsComplement(pNode0) && Bdc_Regular(pNode0).Type != BDC_TYPE_PI {
            fmt.Printf(")"); }

        if Bdc_IsComplement(pNode1) {
            fmt.Printf("!"); }
        if Bdc_IsComplement(pNode1) && Bdc_Regular(pNode1).Type != BDC_TYPE_PI {
            fmt.Printf("("); }
        Bdc_ManDecPrint_rec( p, Bdc_Regular(pNode1) );
        if Bdc_IsComplement(pNode1) && Bdc_Regular(pNode1).Type != BDC_TYPE_PI {
            fmt.Printf(")"); }
    
    } else { assert(false, "0"); }
}
func Bdc_ManDecPrint(p *Bdc_Man_t) {
    var pRoot *Bdc_Fun_t  = Bdc_Regular(p.pRoot);

    fmt.Printf("F = ");
    if pRoot.Type == BDC_TYPE_CONST1 { // constant 0
        fmt.Printf("Constant %d", !Bdc_IsComplement(p.pRoot));
    } else if pRoot.Type == BDC_TYPE_PI { // literal
        fmt.Printf("%s%d", Bdc_IsComplement(p.pRoot) ? "!" : "", Bdc_FunId(p,pRoot)-1);
    } else {
        if Bdc_IsComplement(p.pRoot) {
            fmt.Printf("!("); }
        Bdc_ManDecPrint_rec( p, pRoot );
        if Bdc_IsComplement(p.pRoot) {
            fmt.Printf(")"); }
    }
    fmt.Printf("\n");
}

/**Function*************************************************************

  Synopsis    [Performs decomposition of one function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Bdc_ManDecompose(p *Bdc_Man_t, puFunc *uint32, puCare *uint32, nVars int32, vDivs *Vec_Ptr_t, nNodesMax int32) int32 {
    var Isf Bdc_Isf_t;
    var pIsf *Bdc_Isf_t = &Isf;
    var clk abctime  = Abc_Clock();
    assert(nVars <= p.pPars.nVarsMax, "nVars <= p.pPars.nVarsMax");
    // set current manager parameters
    p.nVars = nVars;
    p.nWords = Kit_TruthWordNum( nVars );
    p.nNodesMax = nNodesMax;
    Bdc_ManPrepare( p, vDivs );
    if (puCare != nil) && Kit_TruthIsConst0( puCare, nVars ) {
        p.pRoot = Bdc_Not(p.pNodes);
        return 0;
    }
    // copy the function
    Bdc_IsfStart( p, pIsf );
    if puCare != nil {
        Kit_TruthAnd( pIsf.puOn, puCare, puFunc, p.nVars );
        Kit_TruthSharp( pIsf.puOff, puCare, puFunc, p.nVars );
    
    } else {
        Kit_TruthCopy( pIsf.puOn, puFunc, p.nVars );
        Kit_TruthNot( pIsf.puOff, puFunc, p.nVars );
    }
    Bdc_SuppMinimize( p, pIsf );
    // call decomposition
    p.pRoot = Bdc_ManDecompose_rec( p, pIsf );
    p.timeTotal += Abc_Clock() - clk;
    p.numCalls++;
    p.numNodes += p.nNodesNew;
    if p.pRoot == nil {
        return -1; }
    if !Bdc_ManNodeVerify( p, pIsf, p.pRoot ) {
        fmt.Printf("Bdc_ManDecompose(): Internal verification failed.\n"); }
//    assert( Bdc_ManNodeVerify( p, pIsf, p->pRoot ) );
    return p.nNodesNew;
}

/**Function*************************************************************

  Synopsis    [Performs decomposition of one function.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
var spStatic_Bdc_ManDecomposeTest_Counter int32;
var spStatic_Bdc_ManDecomposeTest_Total int32;
func Bdc_ManDecomposeTest(uTruth uint32, nVars int32) {
    
    
    var Pars Bdc_Par_t ;
    var pPars *Bdc_Par_t = &Pars;
    var p *Bdc_Man_t;
    var RetValue int32;
//    unsigned uCare = ~0x888f888f;
    var uCare uint32  = ~0;
//    unsigned uFunc =  0x88888888;
//    unsigned uFunc =  0xf888f888;
//    unsigned uFunc =  0x117e117e;
//    unsigned uFunc =  0x018b018b;
    var uFunc uint32  = uTruth; 

    pPars.nVarsMax = 8;
    p = Bdc_ManAlloc( pPars );
    RetValue = Bdc_ManDecompose( p, &uFunc, &uCare, nVars, nil, 1000 );
    spStatic_Bdc_ManDecomposeTest_Total += RetValue;
    fmt.Printf("%5d : Nodes = %5d. Total = %8d.\n", ++spStatic_Bdc_ManDecomposeTest_Counter, RetValue, spStatic_Bdc_ManDecomposeTest_Total);
//    Bdc_ManDecPrint( p );
    Bdc_ManFree( p );
}


////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

