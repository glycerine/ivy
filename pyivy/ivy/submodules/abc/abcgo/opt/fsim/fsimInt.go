package fsim

/**CFile****************************************************************

  FileName    [fsimInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Fast sequential AIG simulator.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: fsimInt.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "aig/saig/saig.h"
//#include "fsim.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// simulation object
   
type Fsim_Obj_t struct {
    iNode             int32;        // the node ID
    iFan0             int32;        // the first fanin
    iFan1             int32;        // the second fanin
}

// fast sequential simulation manager
type Fsim_Man_t_ struct {
    // parameters
    pAig      *Aig_Man_t;         // the AIG to be used for simulation
    nWords             int32;       // the number of simulation words
    // AIG representation
    nPis             int32;         // the number of primary inputs
    nPos             int32;         // the number of primary outputs
    nCis             int32;         // the number of combinational inputs
    nCos             int32;         // the number of combinational outputs
    nNodes             int32;       // the number of internal nodes
    nObjs             int32;        // nCis + nNodes + nCos + 2
    pFans0            *int32;       // fanin0 for all objects
    pFans1            *int32;       // fanin1 for all objects
    pRefs            *int32;        // reference counter for each node
    pRefsCopy            *int32;    // reference counter for each node
    vCis2Ids      *Vec_Int_t;     // mapping of CIs into their PI ids
    vLos      *Vec_Int_t;         // register outputs
    vLis      *Vec_Int_t;         // register inputs
    // cross-cut representation
    nCrossCut             int32;    // temporary cross-cut variable 
    nCrossCutMax             int32; // maximum cross-cut variable
    nFront             int32;       // the size of frontier
    // derived AIG representation
    nDataAig             int32;     // the length of allocated data
    unsigned char * pDataAig;     // AIG representation
    unsigned char * pDataCur;     // AIG representation (current position)
    iNodePrev             int32;    // previous extracted value
    iNumber             int32;      // the number of the last object
    Obj      Fsim_Obj_t;          // current object
    // temporary AIG representation
    pDataAig2            *int32;    // temporary representation
    pDataCur2            *int32;    // AIG representation (current position)
    // simulation information
    pDataSim       *uint32;     // simulation data
    pDataSimCis       *uint32;  // simulation data for CIs
    pDataSimCos       *uint32;  // simulation data for COs
    // other information
    pData1            *int32;
    pData2            *int32;
}

func Fsim_SimData(p *Fsim_Man_t, i int32) *uint32 { return p.pDataSim + i * p.nWords;    }
func Fsim_SimDataCi(p *Fsim_Man_t, i int32) *uint32 { return p.pDataSimCis + i * p.nWords; }
func Fsim_SimDataCo(p *Fsim_Man_t, i int32) *uint32 { return p.pDataSimCos + i * p.nWords; }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

func Fsim_Var2Lit(Var int32, fCompl int32) int32 { return Var + Var + fCompl; }
func Fsim_Lit2Var(Lit int32) int32 { return Lit >> 1;           }
func Fsim_LitIsCompl(Lit int32) int32 { return Lit & 1;            }
func Fsim_LitNot(Lit int32) int32 { return Lit ^ 1;            }
func Fsim_LitNotCond(Lit int32, c int32) int32 { return Lit ^ int32((c > 0)); }
func Fsim_LitRegular(Lit int32) int32 { return Lit & ~01;          }

// Original macro:
// #define Fsim_ManForEachObj( p, pObj, i )\
//     for ( i = 2, p->pDataCur = p->pDataAig, p->iNodePrev = 0, pObj = &p->Obj;\
//         i < p->nObjs && Fsim_ManRestoreObj( p, pObj ); i++ )


////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== fsimFront.c ========================================================*/
// extern void                Fsim_ManFront( Fsim_Man_t * p, int fCompressAig );

/*=== fsimMan.c ==========================================================*/
// extern Fsim_Man_t *        Fsim_ManCreate( Aig_Man_t * pAig );

// extern void                Fsim_ManDelete( Fsim_Man_t * p );

// extern void                Fsim_ManTest( Aig_Man_t * pAig );

 


//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

