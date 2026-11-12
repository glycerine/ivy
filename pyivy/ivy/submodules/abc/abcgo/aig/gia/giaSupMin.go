package gia

/**CFile****************************************************************

  FileName    [giaSupMin.c]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [Support minimization for AIGs with don't-cares.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: giaSupMin.c,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

//#include "gia.h"
//#include "bool/kit/kit.h"

//ABC_NAMESPACE_IMPL_START


////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

// decomposition manager
   
type Gia_ManSup_t struct {
    nVarsMax             int32;     // the max number of variables
    nWordsMax             int32;    // the max number of words
    vTruthVars      *Vec_Ptr_t;   // elementary truth tables
    vTruthNodes      *Vec_Ptr_t;  // internal truth tables
    // current problem
    pGia      *Gia_Man_t;
    iData             int32;
    iCare             int32;
    vConeCare      *Vec_Int_t;
    vConeData      *Vec_Int_t;
    pTruthIn       *uint32;
    pTruthOut       *uint32;
}

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/**Function*************************************************************

  Synopsis    [Starts Decmetry manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSupStart(nVarsMax int32) *Gia_ManSup_t {
    var p *Gia_ManSup_t;
    assert(nVarsMax <= 20, "nVarsMax <= 20");
    p = new(Gia_ManSup_t);
    p.nVarsMax    = nVarsMax;
    p.nWordsMax   = Kit_TruthWordNum( p.nVarsMax );
    p.vTruthVars  = Vec_PtrAllocTruthTables( p.nVarsMax );
    p.vTruthNodes = Vec_PtrAllocSimInfo( 512, p.nWordsMax );
    p.vConeCare   = Vec_IntAlloc( 512 );
    p.vConeData   = Vec_IntAlloc( 512 );
    p.pTruthIn    = ABC_ALLOC( unsigned, p.nWordsMax );
    p.pTruthOut   = ABC_ALLOC( unsigned, p.nWordsMax );
    return p;
}

/**Function*************************************************************

  Synopsis    [Stops Decmetry manager.]

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSupStop(p *Gia_ManSup_t) {
    ABC_FREE( p.pTruthIn );
    ABC_FREE( p.pTruthOut );
    Vec_IntFreeP( &p.vConeCare );
    Vec_IntFreeP( &p.vConeData );
    Vec_PtrFreeP( &p.vTruthVars );
    Vec_PtrFreeP( &p.vTruthNodes );
    ABC_FREE( p );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSupExperimentOne(p *Gia_ManSup_t, pData *Gia_Obj_t, pCare *Gia_Obj_t) {
    var iData int32  = Gia_ObjId( p.pGia, Gia_Regular(pData) );
    var iCare int32  = Gia_ObjId( p.pGia, Gia_Regular(pCare) );
    if !Gia_ObjIsAnd(Gia_Regular(pCare)) {
        Abc_Print( 1, "Enable is not an AND.\n" );
        return;
    }
    Abc_Print( 1, "DataSupp = %6d. DataCone = %6d.   CareSupp = %6d. CareCone = %6d.", 
        Gia_ManSuppSize( p.pGia, &iData, 1 ),
        Gia_ManConeSize( p.pGia, &iData, 1 ),
        Gia_ManSuppSize( p.pGia, &iCare, 1 ),
        Gia_ManConeSize( p.pGia, &iCare, 1 ) );
    Abc_Print( 1, "\n" );
}

/**Function*************************************************************

  Synopsis    []

  Description []
               
  SideEffects []

  SeeAlso     []

***********************************************************************/
func Gia_ManSupExperiment(pGia *Gia_Man_t, vPairs *Vec_Int_t) {
    var p *Gia_ManSup_t;
    var pData, pCare *Gia_Obj_t;
    var i int32;
    p = Gia_ManSupStart( 16 );
    p.pGia = pGia;
    assert(Vec_IntSize(vPairs) % 2 == 0, "Vec_IntSize(vPairs) % 2 == 0");
    for  i = 0; i < Vec_IntSize(vPairs)/2; i++  {
        Abc_Print( 1, "%6d : ", i );
        pData = Gia_ManPo( pGia, Vec_IntEntry(vPairs, 2*i+0) );
        pCare = Gia_ManPo( pGia, Vec_IntEntry(vPairs, 2*i+1) );
        Gia_ManSupExperimentOne( p, Gia_ObjChild0(pData), Gia_ObjChild0(pCare) );
    }
    Gia_ManSupStop( p );
}

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////


//ABC_NAMESPACE_IMPL_END

