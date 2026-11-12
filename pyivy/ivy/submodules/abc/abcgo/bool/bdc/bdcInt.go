package bdc

/**CFile****************************************************************

  FileName    [bdcInt.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Truth-table-based bi-decomposition engine.]

  Synopsis    [Internal declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 15, 2007.]

  Revision    [$Id: resInt.h,v 1.00 2007/01/15 00:00:00 alanmi Exp $]

***********************************************************************/
 




////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "bool/kit/kit.h"
//#include "bdc.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


const BDC_SCALE = 1000  // value used to compute the cost

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// network types
 type Bdc_Type_t int32
const ( 
    BDC_TYPE_NONE Bdc_Type_t = 0
  // 0:  unknown
    BDC_TYPE_CONST1 Bdc_Type_t = BDC_TYPE_NONE + 1
    // 1:  constant 1
    BDC_TYPE_PI Bdc_Type_t = BDC_TYPE_CONST1 + 1
        // 2:  primary input
    BDC_TYPE_AND Bdc_Type_t = BDC_TYPE_PI + 1
       // 3:  AND-gate
    BDC_TYPE_OR Bdc_Type_t = BDC_TYPE_AND + 1
        // 4:  OR-gate (temporary)
    BDC_TYPE_XOR Bdc_Type_t = BDC_TYPE_OR + 1
       // 5:  XOR-gate
    BDC_TYPE_MUX Bdc_Type_t = BDC_TYPE_XOR + 1
       // 6:  MUX-gate
    BDC_TYPE_OTHER Bdc_Type_t = BDC_TYPE_MUX + 1      // 7:  unused

)

 

type Bdc_Fun_t_ struct {
    Type              int32;         // Const1, PI, AND, XOR, MUX
    pFan0       *Bdc_Fun_t;        // fanin of the given node
    pFan1       *Bdc_Fun_t;        // fanin of the given node
    uSupp         uint32;        // bit mask of current support
    puFunc        *uint32;       // the function of the node
    pNext       *Bdc_Fun_t;        // next function with same support
    union { int      iCopy;        // the literal of the node (AIG)
    void *           pCopy; };     // the function of the node (BDD or AIG)

}

   
type Bdc_Isf_t struct {
    uSupp         uint32;        // the complete support of this component
    uUniq         uint32;        // the unique variables of this component
    puOn        *uint32;         // on-set
    puOff        *uint32;        // off-set
}

type Bdc_Man_t_ struct {
    // external parameters
    pPars       *Bdc_Par_t;        // parameter set
    nVars              int32;        // the number of variables
    nWords              int32;       // the number of words 
    nNodesMax              int32;    // the limit on the number of new nodes
    nDivsLimit              int32;   // the limit on the number of divisors
    // internal nodes
    pNodes       *Bdc_Fun_t;       // storage for decomposition nodes
    nNodesAlloc              int32;  // the number of nodes allocated  
    nNodes              int32;       // the number of all nodes created so far
    nNodesNew              int32;    // the number of new AND nodes created so far
    pRoot       *Bdc_Fun_t;        // the root node
    // resub candidates
    pTable      **Bdc_Fun_t;       // hash table of candidates
    nTableSize              int32;   // hash table size (1 << nVarsMax)
    vSpots       *Vec_Int_t;       // the occupied spots in the table
    // elementary truth tables 
    vTruths       *Vec_Ptr_t;      // for const 1 and elementary variables
    puTemp1        *uint32;      // temporary truth table
    puTemp2        *uint32;      // temporary truth table
    puTemp3        *uint32;      // temporary truth table
    puTemp4        *uint32;      // temporary truth table
    // temporary ISFs
    pIsfOL *Bdc_Isf_t;
IsfOL Bdc_Isf_t;
    pIsfOR *Bdc_Isf_t;
IsfOR Bdc_Isf_t;
    pIsfAL *Bdc_Isf_t;
IsfAL Bdc_Isf_t;
    pIsfAR *Bdc_Isf_t;
IsfAR Bdc_Isf_t;
    // internal memory manager
    vMemory       *Vec_Int_t;      // memory for internal truth tables
    // statistics
    numCalls              int32;
    numNodes              int32;
    numMuxes              int32;
    numAnds              int32;
    numOrs              int32;
    numWeaks              int32;
    numReuse              int32;
    // runtime
    timeCache          abctime;
    timeCheck          abctime;
    timeMuxes          abctime;
    timeSupps          abctime;
    timeTotal          abctime;
}

func Bdc_FunNew(p *Bdc_Man_t) *Bdc_Fun_t { var pRes *Bdc_Fun_t; if p.nNodes >= p.nNodesAlloc || p.nNodesNew >= p.nNodesMax { return nil; }
 pRes = p.pNodes + p.nNodes++; p.nNodesNew++; memset( pRes, 0, sizeof(Bdc_Fun_t) ); return pRes; }
func Bdc_FunWithId(p *Bdc_Man_t, Id int32) *Bdc_Fun_t { assert(Id < p.nNodes, "Id < p.nNodes"); return p.pNodes + Id; }
func Bdc_FunId(p *Bdc_Man_t, pFun *Bdc_Fun_t) int32 { return pFun - p.pNodes; }
func Bdc_IsfStart(p *Bdc_Man_t, pF *Bdc_Isf_t) { pF.uSupp = 0; pF.uUniq = 0; pF.puOn = Vec_IntFetch( p.vMemory, p.nWords ); pF.puOff = Vec_IntFetch( p.vMemory, p.nWords ); assert(pF.puOff && pF.puOn, "pF.puOff && pF.puOn"); }
func Bdc_IsfClean(p *Bdc_Isf_t) { p.uSupp = 0; p.uUniq = 0;                                      }
func Bdc_IsfCopy(p *Bdc_Isf_t, q *Bdc_Isf_t) { var T Bdc_Isf_t  = *p; *p = *q; *q = T;                               }
func Bdc_IsfNot(p *Bdc_Isf_t) { var puT *uint32  = p.puOn; p.puOn = p.puOff; p.puOff = puT;    }

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== bdcDec.c ==========================================================*/
// extern Bdc_Fun_t *      Bdc_ManDecompose_rec( Bdc_Man_t * p, Bdc_Isf_t * pIsf );

// extern void             Bdc_SuppMinimize( Bdc_Man_t * p, Bdc_Isf_t * pIsf );

// extern int              Bdc_ManNodeVerify( Bdc_Man_t * p, Bdc_Isf_t * pIsf, Bdc_Fun_t * pFunc );

/*=== bdcTable.c ==========================================================*/
// extern Bdc_Fun_t *      Bdc_TableLookup( Bdc_Man_t * p, Bdc_Isf_t * pIsf );

// extern void             Bdc_TableAdd( Bdc_Man_t * p, Bdc_Fun_t * pFunc );

// extern void             Bdc_TableClear( Bdc_Man_t * p );

// extern int              Bdc_TableCheckContainment( Bdc_Man_t * p, Bdc_Isf_t * pIsf, unsigned * puTruth );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

