package sim

/**CFile****************************************************************

  FileName    [simSymSim.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Network and node package.]

  Synopsis    [Simulation to determine two-variable symmetries.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: simSymSim.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "base/abc/abc.h"
//#include "sim.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// static void Sim_SymmsCreateSquare( Sym_Man_t * p, unsigned * pPat );

// static void Sim_SymmsDeriveInfo( Sym_Man_t * p, unsigned * pPat, Abc_Obj_t * pNode, Vec_Ptr_t * vMatrsNonSym, int Output );


////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Detects non-symmetric pairs using one pattern.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsSimulate(p *Sym_Man_t, pPat *uint32, vMatrsNonSym *Vec_Ptr_t) {
    var pNode *Abc_Obj_t;
    var i, nPairsTotal, nPairsSym, nPairsNonSym int32;
    var clk abctime;

    // create the simulation matrix
    Sim_SymmsCreateSquare( p, pPat );
    // simulate each node in the DFS order
clk = Abc_Clock();
    // inlining of macro Vec_PtrForEachEntry
    for  i = 0; (i < Vec_PtrSize(p.vNodes)) && (((pNode) = (Abc_Obj_t *)Vec_PtrEntry(p.vNodes, i)), 1); i++  {
//        if ( Abc_NodeIsConst(pNode) )
//            continue;
        Sim_UtilSimulateNodeOne( pNode, p.vSim, p.nSimWords, 0 );
    }
p.timeSim += Abc_Clock() - clk;
    // collect info into the CO matrices
clk = Abc_Clock();
    // inlining of macro Abc_NtkForEachCo
    for  i = 0; (i < Abc_NtkCoNum(p.pNtk)) && (((pNode) = Abc_NtkCo(p.pNtk, i)), 1); i++  {
        pNode = Abc_ObjFanin0(pNode);
//        if ( Abc_ObjIsCi(pNode) || Abc_AigNodeIsConst(pNode) )
//            continue;
        nPairsTotal  = Vec_IntEntry(p.vPairsTotal, i);
        nPairsSym    = Vec_IntEntry(p.vPairsSym,   i);
        nPairsNonSym = Vec_IntEntry(p.vPairsNonSym,i);
        assert(nPairsTotal >= nPairsSym + nPairsNonSym, "nPairsTotal >= nPairsSym + nPairsNonSym"); 
        if nPairsTotal == nPairsSym + nPairsNonSym {
            continue; }
        Sim_SymmsDeriveInfo( p, pPat, pNode, vMatrsNonSym, i );
    }
p.timeMatr += Abc_Clock() - clk;
}

/**Function*************************************************************

  Synopsis    [Creates the square matrix of simulation info.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsCreateSquare(p *Sym_Man_t, pPat *uint32) {
    var pSimInfo *uint32;
    var pNode *Abc_Obj_t;
    var i, w int32;
    // for each PI var copy the pattern
    // inlining of macro Abc_NtkForEachCi
    for  i = 0; (i < Abc_NtkCiNum(p.pNtk)) && (((pNode) = Abc_NtkCi(p.pNtk, i)), 1); i++  {
        pSimInfo = (unsigned *)Vec_PtrEntry( p.vSim, pNode.Id );
        // inlining of macro Sim_HasBit
        if  ((pPat)[(i)>>5]  & (1<<((i) & 31))) > 0 {
            for  w = 0; w < p.nSimWords; w++  {
                pSimInfo[w] = SIM_MASK_FULL;
            }
        
        } else {
            for  w = 0; w < p.nSimWords; w++  {
                pSimInfo[w] = 0;
            }
        }
        // flip one bit
        // inlining of macro Sim_XorBit
        ((pSimInfo)[(i)>>5] ^= (1<<((i) & 31)));
    }
}

/**Function*************************************************************

  Synopsis    [Transfers the info to the POs.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Sim_SymmsDeriveInfo(p *Sym_Man_t, pPat *uint32, pNode *Abc_Obj_t, vMatrsNonSym *Vec_Ptr_t, Output int32) {
    var pMat *Extra_BitMat_t;
    var vSupport *Vec_Int_t;
    var pSupport *uint32;
    var pSimInfo *uint32;
    var i, w, Index int32;
    // get the matrix, the support, and the simulation info
    pMat = (Extra_BitMat_t *)Vec_PtrEntry( vMatrsNonSym, Output );
    vSupport = Vec_VecEntryInt( p.vSupports, Output );
    pSupport = (unsigned *)Vec_PtrEntry( p.vSuppFun, Output );
    pSimInfo = (unsigned *)Vec_PtrEntry( p.vSim, pNode.Id );
    // generate vectors A1 and A2
    for  w = 0; w < p.nSimWords; w++  {
        p.uPatCol[w] = pSupport[w] & pPat[w] &  pSimInfo[w];
        p.uPatRow[w] = pSupport[w] & pPat[w] & ~pSimInfo[w];
    }
    // add two dimensions
    // inlining of macro Vec_IntForEachEntry
    for  Index = 0; (Index < Vec_IntSize(vSupport)) && (((i) = Vec_IntEntry(vSupport, Index)), 1); Index++  {
        // inlining of macro Sim_HasBit
        if  ((p.uPatCol)[(i)>>5]  & (1<<((i) & 31))) > 0 {
            Extra_BitMatrixOr( pMat, i, p.uPatRow ); }
    }
    // add two dimensions
    // inlining of macro Vec_IntForEachEntry
    for  Index = 0; (Index < Vec_IntSize(vSupport)) && (((i) = Vec_IntEntry(vSupport, Index)), 1); Index++  {
        // inlining of macro Sim_HasBit
        if  ((p.uPatRow)[(i)>>5]  & (1<<((i) & 31))) > 0 {
            Extra_BitMatrixOr( pMat, i, p.uPatCol ); }
    }
    // generate vectors B1 and B2
    for  w = 0; w < p.nSimWords; w++  {
        p.uPatCol[w] = pSupport[w] & ~pPat[w] &  pSimInfo[w];
        p.uPatRow[w] = pSupport[w] & ~pPat[w] & ~pSimInfo[w];
    }
    // add two dimensions
    // inlining of macro Vec_IntForEachEntry
    for  Index = 0; (Index < Vec_IntSize(vSupport)) && (((i) = Vec_IntEntry(vSupport, Index)), 1); Index++  {
        // inlining of macro Sim_HasBit
        if  ((p.uPatCol)[(i)>>5]  & (1<<((i) & 31))) > 0 {
            Extra_BitMatrixOr( pMat, i, p.uPatRow ); }
    }
    // add two dimensions
    // inlining of macro Vec_IntForEachEntry
    for  Index = 0; (Index < Vec_IntSize(vSupport)) && (((i) = Vec_IntEntry(vSupport, Index)), 1); Index++  {
        // inlining of macro Sim_HasBit
        if  ((p.uPatRow)[(i)>>5]  & (1<<((i) & 31))) > 0 {
            Extra_BitMatrixOr( pMat, i, p.uPatCol ); }
    }
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

