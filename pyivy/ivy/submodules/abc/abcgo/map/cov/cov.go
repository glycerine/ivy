package cov

/**CFile****************************************************************

  FileName    [cov.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Mapping into network of SOPs/ESOPs.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
   
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: cov.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/




//#include "base/abc/abc.h"
//#include "covInt.h"


//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                        DECLARATIONS                              ///
////////////////////////////////////////////////////////////////////////

type Cov_Man_t =  Cov_Man_t_;
    

// storage for node information
type Cov_Obj_t struct {
    Min_Cube_t *      pCover[3];    // pos/neg/esop
    vSupp        *Vec_Int_t;        // computed support (all nodes except CIs)
}

// storage for additional information
type Cov_Man_t_ struct {
    // general characteristics
    nFaninMax               int32;    // the number of vars
    nCubesMax               int32;    // the limit on the number of cubes in the intermediate covers
    nWords               int32;       // the number of words
    vFanCounts        *Vec_Int_t;   // fanout counts
    vObjStrs        *Vec_Ptr_t;     // object structures
    pMemory             any;      // memory for the internal data strctures
    pManMin        *Min_Man_t;      // the cube manager
    fUseEsop               int32;     // enables ESOPs
    fUseSop               int32;      // enables SOPs
    // arrays to map local variables
    vComTo0        *Vec_Int_t;      // mapping of common variables into first fanin
    vComTo1        *Vec_Int_t;      // mapping of common variables into second fanin
    vPairs0        *Vec_Int_t;      // the first var in each pair of common vars
    vPairs1        *Vec_Int_t;      // the second var in each pair of common vars
    vTriv0        *Vec_Int_t;       // trival support of the first node
    vTriv1        *Vec_Int_t;       // trival support of the second node
    // statistics
    nSupps               int32;       // supports created
    nSuppsMax               int32;    // the maximum number of supports
    nBoundary               int32;    // the boundary size
    nNodes               int32;       // the number of nodes processed
}

func Abc_ObjGetStr(pObj *Abc_Obj_t) *Cov_Obj_t { return (Cov_Obj_t *)Vec_PtrEntry(((Cov_Man_t *)pObj.pNtk.pManCut).vObjStrs, pObj.Id); }

func Abc_ObjSetSupp(pObj *Abc_Obj_t, vVec *Vec_Int_t) { Abc_ObjGetStr(pObj).vSupp = vVec;   }
func Abc_ObjGetSupp(pObj *Abc_Obj_t) *Vec_Int_t { return Abc_ObjGetStr(pObj).vSupp;   }

func Abc_ObjSetCover2(pObj *Abc_Obj_t, pCov *Min_Cube_t) { Abc_ObjGetStr(pObj).pCover[2] = pCov; }
func Abc_ObjGetCover2(pObj *Abc_Obj_t) *Min_Cube_t { return Abc_ObjGetStr(pObj).pCover[2]; }

func Abc_ObjSetCover(pObj *Abc_Obj_t, pCov *Min_Cube_t, Pol int32) { Abc_ObjGetStr(pObj).pCover[Pol] = pCov; }
func Abc_ObjGetCover(pObj *Abc_Obj_t, Pol int32) *Min_Cube_t { return Abc_ObjGetStr(pObj).pCover[Pol]; }

////////////////////////////////////////////////////////////////////////
///                     FUNCTION DEFINITIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== covBuild.c ==========================================================*/
// extern Abc_Ntk_t * Abc_NtkCovDerive( Cov_Man_t * p, Abc_Ntk_t * pNtk );

// extern Abc_Ntk_t * Abc_NtkCovDeriveClean( Cov_Man_t * p, Abc_Ntk_t * pNtk );

// extern Abc_Ntk_t * Abc_NtkCovDeriveRegular( Cov_Man_t * p, Abc_Ntk_t * pNtk );

/*=== covCore.c ===========================================================*/
// extern Abc_Ntk_t * Abc_NtkSopEsopCover( Abc_Ntk_t * pNtk, int nFaninMax, int nCubesMax, int fUseEsop, int fUseSop, int fUseInvs, int fVerbose );

/*=== covMan.c ============================================================*/
// extern Cov_Man_t * Cov_ManAlloc( Abc_Ntk_t * pNtk, int nFaninMax, int nCubesMax );

// extern void        Cov_ManFree( Cov_Man_t * p );

// extern void        Abc_NodeCovDropData( Cov_Man_t * p, Abc_Obj_t * pObj );

/*=== covTest.c ===========================================================*/
// extern Abc_Ntk_t * Abc_NtkCovTestSop( Abc_Ntk_t * pNtk );



//ABC_NAMESPACE_HEADER_END



////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////



