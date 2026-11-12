package bdc

/**CFile****************************************************************

  FileName    [bdc.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Truth-table-based bi-decomposition engine.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 30, 2007.]

  Revision    [$Id: bdc.h,v 1.00 2007/01/30 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////
 


//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Bdc_Fun_t = Bdc_Fun_t_;
type Bdc_Man_t = Bdc_Man_t_;
   
type Bdc_Par_t struct {
    // general parameters
    nVarsMax           int32;      // the maximum support
    fVerbose           int32;      // enable basic stats
    fVeryVerbose           int32;  // enable detailed stats
}

// working with complemented attributes of objects
func Bdc_IsComplement(p *Bdc_Fun_t) int32 { return int32(((ABC_PTRUINT_T)p & (ABC_PTRUINT_T)01));              }
func Bdc_Regular(p *Bdc_Fun_t) *Bdc_Fun_t { return (Bdc_Fun_t *)((ABC_PTRUINT_T)p & ~(ABC_PTRUINT_T)01);     }
func Bdc_Not(p *Bdc_Fun_t) *Bdc_Fun_t { return (Bdc_Fun_t *)((ABC_PTRUINT_T)p ^  (ABC_PTRUINT_T)01);     }
func Bdc_NotCond(p *Bdc_Fun_t, c int32) *Bdc_Fun_t { return (Bdc_Fun_t *)((ABC_PTRUINT_T)p ^  (ABC_PTRUINT_T)(c!=0)); }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== bdcCore.c ==========================================================*/
// extern Bdc_Man_t * Bdc_ManAlloc( Bdc_Par_t * pPars );

// extern void        Bdc_ManFree( Bdc_Man_t * p );

// extern void        Bdc_ManDecPrint( Bdc_Man_t * p );

// extern int         Bdc_ManDecompose( Bdc_Man_t * p, unsigned * puFunc, unsigned * puCare, int nVars, Vec_Ptr_t * vDivs, int nNodesMax );

// extern Bdc_Fun_t * Bdc_ManFunc( Bdc_Man_t * p, int i );

// extern Bdc_Fun_t * Bdc_ManRoot( Bdc_Man_t * p );

// extern int         Bdc_ManNodeNum( Bdc_Man_t * p );

// extern int         Bdc_ManAndNum( Bdc_Man_t * p );

// extern Bdc_Fun_t * Bdc_FuncFanin0( Bdc_Fun_t * p );

// extern Bdc_Fun_t * Bdc_FuncFanin1( Bdc_Fun_t * p );

// extern void *      Bdc_FuncCopy( Bdc_Fun_t * p );

// extern int         Bdc_FuncCopyInt( Bdc_Fun_t * p );

// extern void        Bdc_FuncSetCopy( Bdc_Fun_t * p, void * pCopy );

// extern void        Bdc_FuncSetCopyInt( Bdc_Fun_t * p, int iCopy );


/*=== working with saved copies ==========================================*/
func Bdc_FunObjCopy(pObj *Bdc_Fun_t) int32 { return Abc_LitNotCond( Bdc_FuncCopyInt(Bdc_Regular(pObj)), Bdc_IsComplement(pObj) );  }
func Bdc_FunFanin0Copy(pObj *Bdc_Fun_t) int32 { return Bdc_FunObjCopy( Bdc_FuncFanin0(pObj) );                                        }
func Bdc_FunFanin1Copy(pObj *Bdc_Fun_t) int32 { return Bdc_FunObjCopy( Bdc_FuncFanin1(pObj) );                                        }


//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

