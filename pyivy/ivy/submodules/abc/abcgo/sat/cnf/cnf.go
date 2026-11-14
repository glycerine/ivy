package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge
type Vec_Int_t = vec.Vec_Int_t
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Str_t = vec.Vec_Str_t
type Dar_Cut_t struct {
	nLeaves int32
	pLeaves []int32
	uTruth  uint16
	nCuts   int32
}
type Aig_MmFlex_t = struct{}
type abctime = time.Duration

/**CFile****************************************************************

  FileName    [cnf.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [DAG-aware AIG rewriting.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - April 28, 2007.]

  Revision    [$Id: cnf.h,v 1.00 2007/04/28 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "aig/aig/aig.h"
//#include "opt/dar/darInt.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Cnf_Man_t = Cnf_Man_t_
type Cnf_Dat_t = Cnf_Dat_t_
type Cnf_Cut_t = Cnf_Cut_t_

// the CNF asserting outputs of AIG to be 1
type Cnf_Dat_t_ struct {
	Variables  int
	Clauses    [][]int32
	VarNums    []int32
	Mapping    []int32
	Obj2Clause []int32
	Obj2Count  []int32

	pMan        *Aig_Man_t // the AIG manager, for which CNF is computed
	nVars       int32      // the number of variables
	nLiterals   int32      // the number of CNF literals
	nClauses    int32      // the number of CNF clauses
	pClauses    [][]int32  // the CNF clauses
	pVarNums    []int32    // CNF variable by object ID (-1 if unused)
	pObj2Clause []int32
	pObj2Count  []int32
	pClaPols    []byte
	vMapping    *Vec_Int_t
}

// Cnf_DataVarNum, Cnf_DataClauseNum, and Cnf_DataClause expose the read-only
// portions of a CNF needed by packages that cannot access this package's
// source-compatible private fields.
func Cnf_DataVarNum(p *Cnf_Dat_t) int32            { return p.nVars }
func Cnf_DataClauseNum(p *Cnf_Dat_t) int32         { return p.nClauses }
func Cnf_DataClause(p *Cnf_Dat_t, i int32) []int32 { return p.pClauses[i] }
func Cnf_DataSetManNil(p *Cnf_Dat_t)               { p.pMan = nil }

// Cnf_DataSetPackedClauses imports ABC's packed clause representation. offsets
// has one entry per clause plus a final end offset; literals is copied into
// independent clause slices so subsequent Cnf_DataLift calls remain safe.
func Cnf_DataSetPackedClauses(p *Cnf_Dat_t, variables int32, offsets, literals []int32) {
	if p == nil || variables < 0 || len(offsets) == 0 || offsets[0] != 0 || offsets[len(offsets)-1] != int32(len(literals)) {
		panic("cnf: invalid packed clauses")
	}
	clauses := make([][]int32, len(offsets)-1)
	for i := range clauses {
		beg, end := offsets[i], offsets[i+1]
		if beg < 0 || end < beg || end > int32(len(literals)) {
			panic("cnf: invalid clause offsets")
		}
		clauses[i] = append([]int32(nil), literals[beg:end]...)
	}
	p.nVars, p.Variables = variables, int(variables)
	p.nClauses = int32(len(clauses))
	p.nLiterals = int32(len(literals))
	p.Clauses, p.pClauses = clauses, clauses
}

// Cnf_DataClauses returns the CNF's clauses in order. The slices are borrowed
// and reflect literal lifting performed on the CNF.
func Cnf_DataClauses(p *Cnf_Dat_t) [][]int32 { return p.pClauses }

// the cut used to represent node in the AIG
type Cnf_Cut_t_ struct {
	nFanins byte  // the number of leaves
	Cost    byte  // the cost of this cut
	nWords  int16 // the number of words in truth table
	vIsop   [2]*Vec_Int_t
	pFanins []int32
	truth   []uint32
}

// the CNF computation manager
type Cnf_Man_t_ struct {
	pManAig      *Aig_Man_t // the underlying AIG manager
	pCuts        *aig.Aig_ManCut_t
	selectedCuts []*Cnf_Cut_t
	pSopSizes    []byte
	pSops        []string
	aArea        int32         // the area of the mapping
	pMemCuts     *Aig_MmFlex_t // memory manager for cuts
	nMergeLimit  int32         // the limit on the size of merged cut
	pTruths      [4][]uint32
	vMemory      *Vec_Int_t // memory for intermediate ISOP representation
	timeCuts     abctime
	timeMap      abctime
	timeSave     abctime
}

// inlining of macro Dar_ObjForEachCut
func Dar_ObjBestCut(pObj *Aig_Obj_t) *Dar_Cut_t { return nil }

func Cnf_CutSopCost(p *Cnf_Man_t, pCut *Dar_Cut_t) int32 {
	if p == nil || pCut == nil {
		panic("cnf: missing SOP cost input")
	}
	truth := uint16(pCut.uTruth)
	return int32(p.pSopSizes[truth]) + int32(p.pSopSizes[^truth])
}

func Cnf_CutLeaveNum(pCut *Cnf_Cut_t) int32 { return int32(pCut.nFanins) }
func Cnf_CutLeaves(pCut *Cnf_Cut_t) []int32 { return pCut.pFanins }
func Cnf_CutTruth(pCut *Cnf_Cut_t) []uint32 { return pCut.truth }

func Cnf_ObjBestCut(p *Cnf_Man_t, pObj *Aig_Obj_t) *Cnf_Cut_t {
	if p == nil || pObj == nil {
		return nil
	}
	id := (*pObj).NodeID()
	if id < 0 || id >= len(p.selectedCuts) {
		return nil
	}
	return p.selectedCuts[id]
}
func Cnf_ObjSetBestCut(p *Cnf_Man_t, pObj *Aig_Obj_t, cut *Cnf_Cut_t) {
	if p == nil || pObj == nil {
		return
	}
	id := (*pObj).NodeID()
	if id < 0 || id >= len(p.selectedCuts) {
		return
	}
	p.selectedCuts[id] = cut
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                           ITERATORS                              ///
////////////////////////////////////////////////////////////////////////

// iterator over the clauses
// Original macro:
// #define Cnf_CnfForClause( p, pBeg, pEnd, i )                         \
//     for ( i = 0; i < p->nClauses && (pBeg = p->pClauses[i]) && (pEnd = p->pClauses[i+1]); i++ )

// iterator over leaves of the cut
// Original macro:
// #define Cnf_CutForEachLeaf( p, pCut, pLeaf, i )                         \
//     for ( i = 0; (i < (int)(pCut)->nFanins) && ((pLeaf) = Aig_ManObj(p, (pCut)->pFanins[i])); i++ )

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cnfCore.c ========================================================*/
// extern Vec_Int_t *     Cnf_DeriveMappingArray( Aig_Man_t * pAig );

// extern Cnf_Dat_t *     Cnf_Derive( Aig_Man_t * pAig, int nOutputs );

// extern Cnf_Dat_t *     Cnf_DeriveWithMan( Cnf_Man_t * p, Aig_Man_t * pAig, int nOutputs );

// extern Cnf_Dat_t *     Cnf_DeriveOther( Aig_Man_t * pAig, int fSkipTtMin );

// extern Cnf_Dat_t *     Cnf_DeriveOtherWithMan( Cnf_Man_t * p, Aig_Man_t * pAig, int fSkipTtMin );

// extern void            Cnf_ManPrepare();

// extern Cnf_Man_t *     Cnf_ManRead();

// extern void            Cnf_ManFree();

/*=== cnfCut.c ========================================================*/
// extern Cnf_Cut_t *     Cnf_CutCreate( Cnf_Man_t * p, Aig_Obj_t * pObj );

// extern void            Cnf_CutPrint( Cnf_Cut_t * pCut );

// extern void            Cnf_CutFree( Cnf_Cut_t * pCut );

// extern void            Cnf_CutUpdateRefs( Cnf_Man_t * p, Cnf_Cut_t * pCut, Cnf_Cut_t * pCutFan, Cnf_Cut_t * pCutRes );

// extern Cnf_Cut_t *     Cnf_CutCompose( Cnf_Man_t * p, Cnf_Cut_t * pCut, Cnf_Cut_t * pCutFan, int iFan );

/*=== cnfData.c ========================================================*/
// extern void            Cnf_ReadMsops( char ** ppSopSizes, char *** ppSops );

/*=== cnfFast.c ========================================================*/
// extern void            Cnf_CollectLeaves( Aig_Obj_t * pRoot, Vec_Ptr_t * vSuper, int fStopCompl );

// extern void            Cnf_ComputeClauses( Aig_Man_t * p, Aig_Obj_t * pRoot, Vec_Ptr_t * vLeaves,
//                            Vec_Ptr_t * vNodes, Vec_Int_t * vMap, Vec_Int_t * vCover, Vec_Int_t * vClauses );

// extern void            Cnf_DeriveFastMark( Aig_Man_t * p );

// extern Cnf_Dat_t *     Cnf_DeriveFast( Aig_Man_t * p, int nOutputs );

/*=== cnfMan.c ========================================================*/
// extern Cnf_Man_t *     Cnf_ManStart();

// extern void            Cnf_ManStop( Cnf_Man_t * p );

// extern Vec_Int_t *     Cnf_DataCollectPiSatNums( Cnf_Dat_t * pCnf, Aig_Man_t * p );

// extern Cnf_Dat_t *     Cnf_DataAlloc( Aig_Man_t * pAig, int nVars, int nClauses, int nLiterals );

// extern Cnf_Dat_t *     Cnf_DataDup( Cnf_Dat_t * p );

// extern void            Cnf_DataFree( Cnf_Dat_t * p );

// extern void            Cnf_DataLift( Cnf_Dat_t * p, int nVarsPlus );

// extern void            Cnf_DataCollectFlipLits( Cnf_Dat_t * p, int iFlipVar, Vec_Int_t * vFlips );

// extern void            Cnf_DataLiftAndFlipLits( Cnf_Dat_t * p, int nVarsPlus, Vec_Int_t * vLits );

// extern void            Cnf_DataPrint( Cnf_Dat_t * p, int fReadable );

// extern void            Cnf_DataWriteIntoFile( Cnf_Dat_t * p, char * pFileName, int fReadable, Vec_Int_t * vForAlls, Vec_Int_t * vExists );

// extern void *          Cnf_DataWriteIntoSolver( Cnf_Dat_t * p, int nFrames, int fInit );

// extern void *          Cnf_DataWriteIntoSolverInt( void * pSat, Cnf_Dat_t * p, int nFrames, int fInit );

// extern int             Cnf_DataWriteOrClause( void * pSat, Cnf_Dat_t * pCnf );

// extern int             Cnf_DataWriteAndClauses( void * p, Cnf_Dat_t * pCnf );

// extern void            Cnf_DataTranformPolarity( Cnf_Dat_t * pCnf, int fTransformPos );

// extern int             Cnf_DataAddXorClause( void * pSat, int iVarA, int iVarB, int iVarC );

/*=== cnfMap.c ========================================================*/
// extern void            Cnf_DeriveMapping( Cnf_Man_t * p );

// extern int             Cnf_ManMapForCnf( Cnf_Man_t * p );

/*=== cnfPost.c ========================================================*/
// extern void            Cnf_ManTransferCuts( Cnf_Man_t * p );

// extern void            Cnf_ManFreeCuts( Cnf_Man_t * p );

// extern void            Cnf_ManPostprocess( Cnf_Man_t * p );

/*=== cnfUtil.c ========================================================*/
// extern Vec_Ptr_t *     Aig_ManScanMapping( Cnf_Man_t * p, int fCollect );

// extern Vec_Ptr_t *     Cnf_ManScanMapping( Cnf_Man_t * p, int fCollect, int fPreorder );

// extern Vec_Int_t *     Cnf_DataCollectCiSatNums( Cnf_Dat_t * pCnf, Aig_Man_t * p );

// extern Vec_Int_t *     Cnf_DataCollectCoSatNums( Cnf_Dat_t * pCnf, Aig_Man_t * p );

// extern unsigned char * Cnf_DataDeriveLitPolarities( Cnf_Dat_t * p );

// extern Cnf_Dat_t *     Cnf_DataReadFromFile( char * pFileName );

/*=== cnfWrite.c ========================================================*/
// extern Vec_Int_t *     Cnf_ManWriteCnfMapping( Cnf_Man_t * p, Vec_Ptr_t * vMapped );

// extern void            Cnf_SopConvertToVector( char * pSop, int nCubes, Vec_Int_t * vCover );

// extern Cnf_Dat_t *     Cnf_ManWriteCnf( Cnf_Man_t * p, Vec_Ptr_t * vMapped, int nOutputs );

// extern Cnf_Dat_t *     Cnf_ManWriteCnfOther( Cnf_Man_t * p, Vec_Ptr_t * vMapped );

// extern Cnf_Dat_t *     Cnf_DeriveSimple( Aig_Man_t * p, int nOutputs );

// extern Cnf_Dat_t *     Cnf_DeriveSimpleForRetiming( Aig_Man_t * p );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
