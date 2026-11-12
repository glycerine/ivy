package bsat

/**CFile****************************************************************

  FileName    [satStore.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Proof recording.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: pr.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/
 




/*
    The trace of SAT solving contains the original clauses of the problem
    along with the learned clauses derived during SAT solving.
    The first line of the resulting file contains 3 numbers instead of 2:
    c <num_vars> <num_all_clauses> <num_root_clauses>
*/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include "satSolver.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

#ifdef _WIN32
#define inline __inline // compatible with MS VS 6.0
#endif

// Original macro:
// #define STO_MAX(a,b)  ((a) > (b) ? (a) : (b))


////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

/*
typedef unsigned lit;
// variable/literal conversions (taken from MiniSat)
static inline lit   toLit    (int v)        { return v + v;  }
static inline lit   toLitCond(int v, int c) { return v + v + (c != 0); }
static inline lit   lit_neg  (lit l)        { return l ^ 1;  }
static inline int   lit_var  (lit l)        { return l >> 1; }
static inline int   lit_sign (lit l)        { return l & 1;  }
static inline int   lit_print(lit l)        { return lit_sign(l)? -lit_var(l)-1 : lit_var(l)+1; }
static inline lit   lit_read (int s)        { return s > 0 ? toLit(s-1) : lit_neg(toLit(-s-1)); }
static inline int   lit_check(lit l, int n) { return l >= 0 && lit_var(l) < n;                  }
*/

   
type Sto_Cls_t struct {
    pNext      *Sto_Cls_t;        // the next clause
    pNext0      *Sto_Cls_t;       // the next 0-watch
    pNext1      *Sto_Cls_t;       // the next 1-watch
    Id             int32;           // the clause ID
    fA uint32; /* C bitfield: 1 bits; review packing and truncation */  // belongs to A
    fRoot uint32; /* C bitfield: 1 bits; review packing and truncation */  // original clause
    fVisit uint32; /* C bitfield: 1 bits; review packing and truncation */  // visited clause
    nLits uint32; /* C bitfield: 24 bits; review packing and truncation */  // the number of literals
    lit             pLits[0];     // literals of this clause
}

   
type Sto_Man_t struct {
    // general data
    nVars             int32;        // the number of variables
    nRoots             int32;       // the number of root clauses
    nClauses             int32;     // the number of all clauses
    nClausesA             int32;    // the number of clauses of A 
    pHead      *Sto_Cls_t;        // the head clause
    pTail      *Sto_Cls_t;        // the tail clause
    pEmpty      *Sto_Cls_t;       // the empty clause
    // memory management
    nChunkSize             int32;   // the number of bytes in a chunk
    nChunkUsed             int32;   // the number of bytes used in the last chunk
    pChunkLast           string;   // the last memory chunk
}

// iterators through the clauses
// Original macro:
// #define Sto_ManForEachClause( p, pCls )      for( pCls = p->pHead; pCls; pCls = pCls->pNext )

// Original macro:
// #define Sto_ManForEachClauseRoot( p, pCls )  for( pCls = p->pHead; pCls && pCls->fRoot; pCls = pCls->pNext )


////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== satStore.c ==========================================================*/
// extern Sto_Man_t *  Sto_ManAlloc();

// extern void         Sto_ManFree( Sto_Man_t * p );

// extern int          Sto_ManAddClause( Sto_Man_t * p, lit * pBeg, lit * pEnd );

// extern int          Sto_ManMemoryReport( Sto_Man_t * p );

// extern void         Sto_ManMarkRoots( Sto_Man_t * p );

// extern void         Sto_ManMarkClausesA( Sto_Man_t * p );

// extern void         Sto_ManDumpClauses( Sto_Man_t * p, char * pFileName );

// extern int          Sto_ManChangeLastClause( Sto_Man_t * p );

// extern Sto_Man_t *  Sto_ManLoadClauses( char * pFileName );



/*=== satInter.c ==========================================================*/
type Int_Man_t = Int_Man_t_;
// extern Int_Man_t *  Int_ManAlloc();

// extern int *        Int_ManSetGlobalVars( Int_Man_t * p, int nGloVars );

// extern void         Int_ManFree( Int_Man_t * p );

// extern int          Int_ManInterpolate( Int_Man_t * p, Sto_Man_t * pCnf, int fVerbose, unsigned ** ppResult );


/*=== satInterA.c ==========================================================*/
type Inta_Man_t = Inta_Man_t_;
// extern Inta_Man_t * Inta_ManAlloc();

// extern void         Inta_ManFree( Inta_Man_t * p );

// extern void *       Inta_ManInterpolate( Inta_Man_t * p, Sto_Man_t * pCnf, abctime TimeToStop, void * vVarsAB, int fVerbose );


/*=== satInterB.c ==========================================================*/
type Intb_Man_t = Intb_Man_t_;
// extern Intb_Man_t * Intb_ManAlloc();

// extern void         Intb_ManFree( Intb_Man_t * p );

// extern void *       Intb_ManInterpolate( Intb_Man_t * p, Sto_Man_t * pCnf, void * vVarsAB, int fVerbose );


/*=== satInterP.c ==========================================================*/
type Intp_Man_t = Intp_Man_t_;
// extern Intp_Man_t * Intp_ManAlloc();

// extern void         Intp_ManFree( Intp_Man_t * p );

// extern void *       Intp_ManUnsatCore( Intp_Man_t * p, Sto_Man_t * pCnf, int fLearned, int fVerbose );

// extern void         Intp_ManUnsatCorePrintForBmc( FILE * pFile, Sto_Man_t * pCnf, void * vCore, void * vVarMap );



//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

