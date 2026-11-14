package gia

/**CFile****************************************************************

  FileName    [gia.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [Scalable AIG package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: gia.h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/vec/vecWec.h"
//#include "misc/util/utilCex.h"

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

// Value is currently used to store several types of information
// - pointer to the next node in the hash table during structural hashing
// - pointer to the node copy during duplication

// new AIG manager

type Gps_Par_t struct {
	FTents    int32
	FSwitch   int32
	FCut      int32
	FNpn      int32
	FLutProf  int32
	FMuxXor   int32
	FMiter    int32
	FSkipMap  int32
	FSlacks   int32
	FNoColor  int32
	PDumpFile string
}

type Emb_Par_t struct {
	NDims      int32 // the number of dimension
	NSols      int32 // the number of solutions (typically, 2)
	NIters     int32 // the number of iterations of FORCE
	FRefine    int32 // use refinement by FORCE
	FCluster   int32 // use clustered representation
	FDump      int32 // dump Gnuplot file
	FDumpLarge int32 // dump Gnuplot file for large benchmarks
	FShowImage int32 // shows image if Gnuplot is installed
	FVerbose   int32 // verbose flag
}

// frames parameters

type Gia_ParFra_t struct {
	NFrames      int32 // the number of frames to unroll
	FInit        int32 // initialize the timeframes
	fSaveLastLit int32 // adds POs for outputs of each frame
	FDisableSt   int32 // disables strashing
	FOrPos       int32 // ORs respective POs in each timeframe
	FVerbose     int32 // enables verbose output
}

// simulation parameters

type Gia_ParSim_t struct {
	// user-controlled parameters
	NWords      int32 // the number of machine words
	NIters      int32 // the number of timeframes
	RandSeed    int32 // seed to generate random numbers
	TimeLimit   int32 // time limit in seconds
	FCheckMiter int32 // check if miter outputs are non-zero
	FVerbose    int32 // enables verbose output
	iOutFail    int32 // index of the failed output
}

type Gia_ManSim_t struct {
	pAig     *Gia_Man_t
	pPars    *Gia_ParSim_t
	nWords   int32
	vCis2Ids *Vec_Int_t
	vConsts  *Vec_Int_t
	// simulation information
	pDataSim    []uint32 // simulation data
	pDataSimCis []uint32 // simulation data for CIs
	pDataSimCos []uint32 // simulation data for COs
}

type Jf_Par_t struct {
	NLutSize           int32
	NCutNum            int32
	NProcNum           int32
	NRounds            int32
	NRoundsEla         int32
	NRelaxRatio        int32
	NCoarseLimit       int32
	NAreaTuner         int32
	NReqTimeFlex       int32
	NVerbLimit         int32
	NDelayLut1         int32
	NDelayLut2         int32
	NFastEdges         int32
	DelayTarget        int32
	FAreaOnly          int32
	FPinPerm           int32
	FPinQuick          int32
	FPinFilter         int32
	FOptEdge           int32
	FUseMux7           int32
	FPower             int32
	FCoarsen           int32
	FCutMin            int32
	FFuncDsd           int32
	FGenCnf            int32
	FGenLit            int32
	fCnfObjIds         int32
	fAddOrCla          int32
	fCnfMapping        int32
	FPureAig           int32
	FDoAverage         int32
	FCutHashing        int32
	FCutSimple         int32
	FCutGroup          int32
	FVerbose           int32
	FVeryVerbose       int32
	NLutSizeMax        int32
	NCutNumMax         int32
	NProcNumMax        int32
	NLutSizeMux        int32
	Delay              word
	Area               word
	Edge               word
	Clause             word
	Mux7               word
	WordMapDelay       word
	WordMapArea        word
	WordMapDelayTarget word
	MapDelay           int32
	MapArea            float32
	MapAreaF           float32
	MapDelayTarget     float32
	Epsilon            float32
	pTimesArr          []float32
	pTimesReq          []float32
}

// AIG construction
// extern void Gia_ObjAddFanout( Gia_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pFanout );

// extern void Gia_ManQuantSetSuppAnd( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void Gia_ManBuiltInSimPerform( Gia_Man_t * p, int iObj );

// Original macro:
// #define Gia_ManForEachConst( p, i )                            \
//     for ( i = 1; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsConst(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachClass( p, i )                            \
//     for ( i = 1; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsHead(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachClass0( p, i )                            \
//     for ( i = 0; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsHead(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachClassReverse( p, i )                     \
//     for ( i = Gia_ManObjNum(p) - 1; i > 0; i-- ) if ( !Gia_ObjIsHead(p, i) ) {} else

// Original macro:
// #define Gia_ClassForEachObj( p, i, iObj )                      \
//     for ( assert(Gia_ObjIsHead(p, i) && i), iObj = i; iObj > 0; iObj = Gia_ObjNext(p, iObj) )

// Original macro:
// #define Gia_ClassForEachObj1( p, i, iObj )                     \
//     for ( assert(Gia_ObjIsHead(p, i)), iObj = Gia_ObjNext(p, i); iObj > 0; iObj = Gia_ObjNext(p, iObj) )

// Original macro:
// #define Gia_ClassForEachObjStart( p, i, iObj, Start )          \
//     for ( assert(Gia_ObjIsHead(p, i)), iObj = Gia_ObjNext(p, Start); iObj > 0; iObj = Gia_ObjNext(p, iObj) )

// Original macro:
// #define Gia_ObjForEachFanoutStatic( p, pObj, pFanout, i )      \
//     for ( i = 0; (i < Gia_ObjFanoutNum(p, pObj))   && (((pFanout) = Gia_ObjFanout(p, pObj, i)), 1); i++ )

// Original macro:
// #define Gia_ObjForEachFanoutStaticId( p, Id, FanId, i )      \
//     for ( i = 0; (i < Gia_ObjFanoutNumId(p, Id))   && (((FanId) = Gia_ObjFanoutId(p, Id, i)), 1); i++ )

// Original macro:
// #define Gia_ManForEachLut( p, i )                                       \
//     for ( i = 1; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsLut(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachLutReverse( p, i )                                \
//     for ( i = Gia_ManObjNum(p) - 1; i > 0; i-- ) if ( !Gia_ObjIsLut(p, i) ) {} else

// Original macro:
// #define Gia_LutForEachFanin( p, i, iFan, k )                            \
//     for ( k = 0; k < Gia_ObjLutSize(p,i) && ((iFan = Gia_ObjLutFanins(p,i)[k]),1); k++ )

// Original macro:
// #define Gia_LutForEachFaninObj( p, i, pFanin, k )                       \
//     for ( k = 0; k < Gia_ObjLutSize(p,i) && ((pFanin = Gia_ManObj(p, Gia_ObjLutFanins(p,i)[k])),1); k++ )

// Original macro:
// #define Gia_ManForEachLut2( p, i )                                      \
//     for ( i = 1; i < Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsLut2(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachLut2Reverse( p, i )                               \
//     for ( i = Gia_ManObjNum(p) - 1; i > 0; i-- ) if ( !Gia_ObjIsLut2(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachLut2Vec( vIds, p, vVec, iObj, i )                 \
//     for ( i = 0; i < Vec_IntSize(vIds) && (vVec = Vec_WecEntry(p->vMapping2, (iObj = Vec_IntEntry(vIds, i)))); i++ )

// Original macro:
// #define Gia_ManForEachLut2VecReverse( vIds, p, vVec, iObj, i )          \
//     for ( i = Vec_IntSize(vIds)-1; i >= 0 && (vVec = Vec_WecEntry(p->vMapping2, (iObj = Vec_IntEntry(vIds, i)))); i-- )

// Original macro:
// #define Gia_LutForEachFanin2( p, i, iFan, k )                           \
//     for ( k = 0; k < Gia_ObjLutSize2(p,i) && ((iFan = Gia_ObjLutFanin2(p,i,k)),1); k++ )

// Original macro:
// #define Gia_LutForEachFanout2( p, i, iFan, k )                          \
//     for ( k = 0; k < Gia_ObjLutFanoutNum2(p,i) && ((iFan = Gia_ObjLutFanout2(p,i,k)),1); k++ )

// Original macro:
// #define Gia_ManForEachCell( p, i )                                      \
//     for ( i = 2; i < 2*Gia_ManObjNum(p); i++ ) if ( !Gia_ObjIsCell(p, i) ) {} else

// Original macro:
// #define Gia_CellForEachFanin( p, i, iFanLit, k )                        \
//     for ( k = 0; k < Gia_ObjCellSize(p,i) && ((iFanLit = Gia_ObjCellFanins(p,i)[k]),1); k++ )

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define Gia_ManForEachObj( p, pObj, i )                                 \
//     for ( i = 0; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachObj1( p, pObj, i )                                \
//     for ( i = 1; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachObjVec( vVec, p, pObj, i )                        \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Gia_ManForEachObjVecReverse( vVec, p, pObj, i )                        \
//     for ( i = Vec_IntSize(vVec) - 1; (i >= 0) && ((pObj) = Gia_ManObj(p, Vec_IntEntry(vVec,i))); i-- )

// Original macro:
// #define Gia_ManForEachObjVecLit( vVec, p, pObj, fCompl, i )             \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManObj(p, Abc_Lit2Var(Vec_IntEntry(vVec,i)))) && (((fCompl) = Abc_LitIsCompl(Vec_IntEntry(vVec,i))),1); i++ )

// Original macro:
// #define Gia_ManForEachObjReverse( p, pObj, i )                          \
//     for ( i = p->nObjs - 1; (i >= 0) && ((pObj) = Gia_ManObj(p, i)); i-- )

// Original macro:
// #define Gia_ManForEachObjReverse1( p, pObj, i )                         \
//     for ( i = p->nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p, i)); i-- )

// Original macro:
// #define Gia_ManForEachBuf( p, pObj, i )                                 \
//     for ( i = Gia_ManBufNum(p) ? 0 : p->nObjs; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )      if ( !Gia_ObjIsBuf(pObj) ) {} else

// Original macro:
// #define Gia_ManForEachBufId( p, i )                                     \
//     for ( i = 0; (i < p->nObjs); i++ )                                     if ( !Gia_ObjIsBuf(Gia_ManObj(p, i)) ) {} else

// Original macro:
// #define Gia_ManForEachAnd( p, pObj, i )                                 \
//     for ( i = 0; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )      if ( !Gia_ObjIsAnd(pObj) ) {} else

// Original macro:
// #define Gia_ManForEachAndId( p, i )                                     \
//     for ( i = 0; (i < p->nObjs); i++ )                                     if ( !Gia_ObjIsAnd(Gia_ManObj(p, i)) ) {} else

// Original macro:
// #define Gia_ManForEachMuxId( p, i )                                     \
//     for ( i = 0; (i < p->nObjs); i++ )                                     if ( !Gia_ObjIsMuxId(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachCand( p, pObj, i )                                \
//     for ( i = 0; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )      if ( !Gia_ObjIsCand(pObj) ) {} else

// Original macro:
// #define Gia_ManForEachAndReverse( p, pObj, i )                          \
//     for ( i = p->nObjs - 1; (i > 0) && ((pObj) = Gia_ManObj(p, i)); i-- )  if ( !Gia_ObjIsAnd(pObj) ) {} else

// Original macro:
// #define Gia_ManForEachAndReverseId( p, i )                              \
//     for ( i = p->nObjs - 1; (i > 0); i-- )                                 if ( !Gia_ObjIsAnd(Gia_ManObj(p, i)) ) {} else

// Original macro:
// #define Gia_ManForEachMux( p, pObj, i )                                 \
//     for ( i = 0; (i < p->nObjs) && ((pObj) = Gia_ManObj(p, i)); i++ )      if ( !Gia_ObjIsMuxId(p, i) ) {} else

// Original macro:
// #define Gia_ManForEachCi( p, pObj, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCis)) && ((pObj) = Gia_ManCi(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachCiId( p, Id, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCis)) && ((Id) = Gia_ObjId(p, Gia_ManCi(p, i))); i++ )

// Original macro:
// #define Gia_ManForEachCiVec( vVec, p, pObj, i )                         \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManCi(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Gia_ManForEachCiReverse( p, pObj, i )                           \
//     for ( i = Vec_IntSize(p->vCis) - 1; (i >= 0) && ((pObj) = Gia_ManCi(p, i)); i-- )

// Original macro:
// #define Gia_ManForEachCo( p, pObj, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && ((pObj) = Gia_ManCo(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachCoVec( vVec, p, pObj, i )                         \
//     for ( i = 0; (i < Vec_IntSize(vVec)) && ((pObj) = Gia_ManCo(p, Vec_IntEntry(vVec,i))); i++ )

// Original macro:
// #define Gia_ManForEachCoId( p, Id, i )                                  \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && ((Id) = Gia_ObjId(p, Gia_ManCo(p, i))); i++ )

// Original macro:
// #define Gia_ManForEachCoReverse( p, pObj, i )                           \
//     for ( i = Vec_IntSize(p->vCos) - 1; (i >= 0) && ((pObj) = Gia_ManCo(p, i)); i-- )

// Original macro:
// #define Gia_ManForEachCoDriver( p, pObj, i )                            \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && ((pObj) = Gia_ObjFanin0(Gia_ManCo(p, i))); i++ )

// Original macro:
// #define Gia_ManForEachCoDriverId( p, DriverId, i )                      \
//     for ( i = 0; (i < Vec_IntSize(p->vCos)) && (((DriverId) = Gia_ObjFaninId0p(p, Gia_ManCo(p, i))), 1); i++ )

// Original macro:
// #define Gia_ManForEachPi( p, pObj, i )                                  \
//     for ( i = 0; (i < Gia_ManPiNum(p)) && ((pObj) = Gia_ManCi(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachPo( p, pObj, i )                                  \
//     for ( i = 0; (i < Gia_ManPoNum(p)) && ((pObj) = Gia_ManCo(p, i)); i++ )

// Original macro:
// #define Gia_ManForEachRo( p, pObj, i )                                  \
//     for ( i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++ )

// Original macro:
// #define Gia_ManForEachRi( p, pObj, i )                                  \
//     for ( i = 0; (i < Gia_ManRegNum(p)) && ((pObj) = Gia_ManCo(p, Gia_ManPoNum(p)+i)); i++ )

// Original macro:
// #define Gia_ManForEachRiRo( p, pObjRi, pObjRo, i )                      \
//     for ( i = 0; (i < Gia_ManRegNum(p)) && ((pObjRi) = Gia_ManCo(p, Gia_ManPoNum(p)+i)) && ((pObjRo) = Gia_ManCi(p, Gia_ManPiNum(p)+i)); i++ )

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== giaAiger.c ===========================================================*/
// extern int                 Gia_FileSize( char * pFileName );

// extern Gia_Man_t *         Gia_AigerReadFromMemory( char * pContents, int nFileSize, int fGiaSimple, int fSkipStrash, int fCheck );

// extern Gia_Man_t *         Gia_AigerRead( char * pFileName, int fGiaSimple, int fSkipStrash, int fCheck );

// extern void                Gia_AigerWrite( Gia_Man_t * p, char * pFileName, int fWriteSymbols, int fCompact, int fWriteNewLine );

// extern void                Gia_DumpAiger( Gia_Man_t * p, char * pFilePrefix, int iFileNum, int nFileNumDigits );

// extern Vec_Str_t *         Gia_AigerWriteIntoMemoryStr( Gia_Man_t * p );

// extern Vec_Str_t *         Gia_AigerWriteIntoMemoryStrPart( Gia_Man_t * p, Vec_Int_t * vCis, Vec_Int_t * vAnds, Vec_Int_t * vCos, int nRegs );

// extern void                Gia_AigerWriteSimple( Gia_Man_t * pInit, char * pFileName );

/*=== giaBalance.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManBalance( Gia_Man_t * p, int fSimpleAnd, int fStrict, int fVerbose );

// extern Gia_Man_t *         Gia_ManAreaBalance( Gia_Man_t * p, int fSimpleAnd, int nNewNodesMax, int fVerbose, int fVeryVerbose );

// extern Gia_Man_t *         Gia_ManAigSyn2( Gia_Man_t * p, int fOldAlgo, int fCoarsen, int fCutMin, int nRelaxRatio, int fDelayMin, int fVerbose, int fVeryVerbose );

// extern Gia_Man_t *         Gia_ManAigSyn3( Gia_Man_t * p, int fVerbose, int fVeryVerbose );

// extern Gia_Man_t *         Gia_ManAigSyn4( Gia_Man_t * p, int fVerbose, int fVeryVerbose );

/*=== giaBidec.c ===========================================================*/
// extern unsigned *          Gia_ManConvertAigToTruth( Gia_Man_t * p, Gia_Obj_t * pRoot, Vec_Int_t * vLeaves, Vec_Int_t * vTruth, Vec_Int_t * vVisited );

// extern Gia_Man_t *         Gia_ManPerformBidec( Gia_Man_t * p, int fVerbose );

/*=== giaCex.c ============================================================*/
// extern int                 Gia_ManVerifyCex( Gia_Man_t * pAig, Abc_Cex_t * p, int fDualOut );

// extern int                 Gia_ManFindFailedPoCex( Gia_Man_t * pAig, Abc_Cex_t * p, int nOutputs );

// extern int                 Gia_ManSetFailedPoCex( Gia_Man_t * pAig, Abc_Cex_t * p );

// extern void                Gia_ManCounterExampleValueStart( Gia_Man_t * pGia, Abc_Cex_t * pCex );

// extern void                Gia_ManCounterExampleValueStop( Gia_Man_t * pGia );

// extern int                 Gia_ManCounterExampleValueLookup( Gia_Man_t * pGia, int Id, int iFrame );

// extern Abc_Cex_t *         Gia_ManCexExtendToIncludeCurrentStates( Gia_Man_t * p, Abc_Cex_t * pCex );

// extern Abc_Cex_t *         Gia_ManCexExtendToIncludeAllObjects( Gia_Man_t * p, Abc_Cex_t * pCex );

/*=== giaCsatOld.c ============================================================*/
// extern Vec_Int_t *         Cbs_ManSolveMiter( Gia_Man_t * pGia, int nConfs, Vec_Str_t ** pvStatus, int fVerbose );

/*=== giaCsat.c ============================================================*/

// extern Cbs_Man_t *         Cbs_ManAlloc( Gia_Man_t * pGia );

// extern void                Cbs_ManStop( Cbs_Man_t * p );

// extern int                 Cbs_ManSolve( Cbs_Man_t * p, Gia_Obj_t * pObj );

// extern int                 Cbs_ManSolve2( Cbs_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pObj2 );

// extern Vec_Int_t *         Cbs_ManSolveMiterNc( Gia_Man_t * pGia, int nConfs, Vec_Str_t ** pvStatus, int f0Proved, int fVerbose );

// extern void                Cbs_ManSetConflictNum( Cbs_Man_t * p, int Num );

// extern Vec_Int_t *         Cbs_ReadModel( Cbs_Man_t * p );

/*=== giaCTas.c ============================================================*/
// extern Vec_Int_t *         Tas_ManSolveMiterNc( Gia_Man_t * pGia, int nConfs, Vec_Str_t ** pvStatus, int fVerbose );

/*=== giaCof.c =============================================================*/
// extern void                Gia_ManPrintFanio( Gia_Man_t * pGia, int nNodes );

// extern Gia_Man_t *         Gia_ManDupCof( Gia_Man_t * p, int iVar );

// extern Gia_Man_t *         Gia_ManDupCofAllInt( Gia_Man_t * p, Vec_Int_t * vSigs, int fVerbose );

// extern Gia_Man_t *         Gia_ManDupCofAll( Gia_Man_t * p, int nFanLim, int fVerbose );

/*=== giaDfs.c ============================================================*/
// extern void                Gia_ManCollectCis( Gia_Man_t * p, int * pNodes, int nNodes, Vec_Int_t * vSupp );

// extern void                Gia_ManCollectAnds_rec( Gia_Man_t * p, int iObj, Vec_Int_t * vNodes );

// extern void                Gia_ManCollectAnds( Gia_Man_t * p, int * pNodes, int nNodes, Vec_Int_t * vNodes, Vec_Int_t * vLeaves );

// extern Vec_Int_t *         Gia_ManCollectAndsAll( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManCollectNodesCis( Gia_Man_t * p, int * pNodes, int nNodes );

// extern int                 Gia_ManSuppSize( Gia_Man_t * p, int * pNodes, int nNodes );

// extern int                 Gia_ManConeSize( Gia_Man_t * p, int * pNodes, int nNodes );

// extern Vec_Vec_t *         Gia_ManLevelize( Gia_Man_t * p );

// extern Vec_Wec_t *         Gia_ManLevelizeR( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManOrderReverse( Gia_Man_t * p );

// extern void                Gia_ManCollectTfi( Gia_Man_t * p, Vec_Int_t * vRoots, Vec_Int_t * vNodes );

// extern void                Gia_ManCollectTfo( Gia_Man_t * p, Vec_Int_t * vRoots, Vec_Int_t * vNodes );

/*=== giaDup.c ============================================================*/
// extern void                Gia_ManDupRemapLiterals( Vec_Int_t * vLits, Gia_Man_t * p );

// extern void                Gia_ManDupRemapEquiv( Gia_Man_t * pNew, Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupOrderDfs( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupOrderDfsChoices( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupOrderDfsReverse( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupOutputGroup( Gia_Man_t * p, int iOutStart, int iOutStop );

// extern Gia_Man_t *         Gia_ManDupOutputVec( Gia_Man_t * p, Vec_Int_t * vOutPres );

// extern Gia_Man_t *         Gia_ManDupSelectedOutputs( Gia_Man_t * p, Vec_Int_t * vOutsLeft );

// extern Gia_Man_t *         Gia_ManDupOrderAiger( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupLastPis( Gia_Man_t * p, int nLastPis );

// extern Gia_Man_t *         Gia_ManDupFlip( Gia_Man_t * p, int * pInitState );

// extern Gia_Man_t *         Gia_ManDupCycled( Gia_Man_t * pAig, Abc_Cex_t * pCex, int nFrames );

// extern Gia_Man_t *         Gia_ManDup( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDup2( Gia_Man_t * p1, Gia_Man_t * p2 );

// extern Gia_Man_t *         Gia_ManDupWithAttributes( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupRemovePis( Gia_Man_t * p, int nRemPis );

// extern Gia_Man_t *         Gia_ManDupZero( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupPerm( Gia_Man_t * p, Vec_Int_t * vPiPerm );

// extern Gia_Man_t *         Gia_ManDupPermFlop( Gia_Man_t * p, Vec_Int_t * vFfPerm );

// extern Gia_Man_t *         Gia_ManDupPermFlopGap( Gia_Man_t * p, Vec_Int_t * vFfPerm );

// extern void                Gia_ManDupAppend( Gia_Man_t * p, Gia_Man_t * pTwo );

// extern void                Gia_ManDupAppendShare( Gia_Man_t * p, Gia_Man_t * pTwo );

// extern Gia_Man_t *         Gia_ManDupAppendNew( Gia_Man_t * pOne, Gia_Man_t * pTwo );

// extern Gia_Man_t *         Gia_ManDupAppendCones( Gia_Man_t * p, Gia_Man_t ** ppCones, int nCones, int fOnlyRegs );

// extern Gia_Man_t *         Gia_ManDupSelf( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupFlopClass( Gia_Man_t * p, int iClass );

// extern Gia_Man_t *         Gia_ManDupMarked( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupTimes( Gia_Man_t * p, int nTimes );

// extern Gia_Man_t *         Gia_ManDupDfs( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupDfsOnePo( Gia_Man_t * p, int iPo );

// extern Gia_Man_t *         Gia_ManDupCofactorVar( Gia_Man_t * p, int iVar, int Value );

// extern Gia_Man_t *         Gia_ManDupCofactorObj( Gia_Man_t * p, int iObj, int Value );

// extern Gia_Man_t *         Gia_ManDupMux( int iVar, Gia_Man_t * pCof1, Gia_Man_t * pCof0 );

// extern Gia_Man_t *         Gia_ManDupBlock( Gia_Man_t * p, int nBlock );

// extern Gia_Man_t *         Gia_ManDupExist( Gia_Man_t * p, int iVar );

// extern Gia_Man_t *         Gia_ManDupUniv( Gia_Man_t * p, int iVar );

// extern Gia_Man_t *         Gia_ManDupDfsSkip( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupDfsCone( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern Gia_Man_t *         Gia_ManDupConeSupp( Gia_Man_t * p, int iLit, Vec_Int_t * vCiIds );

// extern int                 Gia_ManDupConeBack( Gia_Man_t * p, Gia_Man_t * pNew, Vec_Int_t * vCiIds );

// extern int                 Gia_ManDupConeBackObjs( Gia_Man_t * p, Gia_Man_t * pNew, Vec_Int_t * vObjs );

// extern Gia_Man_t *         Gia_ManDupDfsNode( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern Gia_Man_t *         Gia_ManDupDfsLitArray( Gia_Man_t * p, Vec_Int_t * vLits );

// extern Gia_Man_t *         Gia_ManDupTrimmed( Gia_Man_t * p, int fTrimCis, int fTrimCos, int fDualOut, int OutValue );

// extern Gia_Man_t *         Gia_ManDupOntop( Gia_Man_t * p, Gia_Man_t * p2 );

// extern Gia_Man_t *         Gia_ManDupWithNewPo( Gia_Man_t * p1, Gia_Man_t * p2 );

// extern Gia_Man_t *         Gia_ManDupDfsCiMap( Gia_Man_t * p, int * pCi2Lit, Vec_Int_t * vLits );

// extern Gia_Man_t *         Gia_ManPermuteInputs( Gia_Man_t * p, int nPpis, int nExtra );

// extern Gia_Man_t *         Gia_ManDupDfsClasses( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupTopAnd( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManMiter( Gia_Man_t * pAig0, Gia_Man_t * pAig1, int nInsDup, int fDualOut, int fSeq, int fImplic, int fVerbose );

// extern Gia_Man_t *         Gia_ManDupAndOr( Gia_Man_t * p, int nOuts, int fUseOr, int fCompl );

// extern Gia_Man_t *         Gia_ManDupZeroUndc( Gia_Man_t * p, char * pInit, int nNewPis, int fGiaSimple, int fVerbose );

// extern Gia_Man_t *         Gia_ManMiter2( Gia_Man_t * p, char * pInit, int fVerbose );

// extern Gia_Man_t *         Gia_ManTransformMiter( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManTransformMiter2( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManTransformToDual( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManTransformTwoWord2DualOutput( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManChoiceMiter( Vec_Ptr_t * vGias );

// extern Gia_Man_t *         Gia_ManDupWithConstraints( Gia_Man_t * p, Vec_Int_t * vPoTypes );

// extern Gia_Man_t *         Gia_ManDupCones( Gia_Man_t * p, int * pPos, int nPos, int fTrimPis );

// extern Gia_Man_t *         Gia_ManDupAndCones( Gia_Man_t * p, int * pAnds, int nAnds, int fTrimPis );

// extern Gia_Man_t *         Gia_ManDupAndConesLimit( Gia_Man_t * p, int * pAnds, int nAnds, int Level );

// extern Gia_Man_t *         Gia_ManDupAndConesLimit2( Gia_Man_t * p, int * pAnds, int nAnds, int Level );

// extern Gia_Man_t *         Gia_ManDupOneHot( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupLevelized( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupFromVecs( Gia_Man_t * p, Vec_Int_t * vCis, Vec_Int_t * vAnds, Vec_Int_t * vCos, int nRegs );

// extern Gia_Man_t *         Gia_ManDupSliced( Gia_Man_t * p, int nSuppMax );

// extern Gia_Man_t *         Gia_ManDupDemiter( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManDemiterToDual( Gia_Man_t * p );

// extern int                 Gia_ManDemiterDual( Gia_Man_t * p, Gia_Man_t ** pp0, Gia_Man_t ** pp1 );

// extern int                 Gia_ManDemiterTwoWords( Gia_Man_t * p, Gia_Man_t ** pp0, Gia_Man_t ** pp1 );

/*=== giaEdge.c ==========================================================*/
// extern void                Gia_ManEdgeFromArray( Gia_Man_t * p, Vec_Int_t * vArray );

// extern Vec_Int_t *         Gia_ManEdgeToArray( Gia_Man_t * p );

// extern void                Gia_ManConvertPackingToEdges( Gia_Man_t * p );

// extern int                 Gia_ObjCheckEdge( Gia_Man_t * p, int iObj, int iNext );

// extern int                 Gia_ManEvalEdgeDelay( Gia_Man_t * p );

// extern int                 Gia_ManEvalEdgeCount( Gia_Man_t * p );

// extern int                 Gia_ManComputeEdgeDelay( Gia_Man_t * p, int fUseTwo );

// extern int                 Gia_ManComputeEdgeDelay2( Gia_Man_t * p );

// extern void                Gia_ManUpdateMapping( Gia_Man_t * p, Vec_Int_t * vNodes, Vec_Wec_t * vWin );

// extern int                 Gia_ManEvalWindow( Gia_Man_t * p, Vec_Int_t * vLeaves, Vec_Int_t * vNodes, Vec_Wec_t * vWin, Vec_Int_t * vTemp, int fUseTwo );

/*=== giaEnable.c ==========================================================*/
// extern void                Gia_ManDetectSeqSignals( Gia_Man_t * p, int fSetReset, int fVerbose );

// extern Gia_Man_t *         Gia_ManUnrollAndCofactor( Gia_Man_t * p, int nFrames, int nFanMax, int fVerbose );

// extern Gia_Man_t *         Gia_ManRemoveEnables( Gia_Man_t * p );

/*=== giaEquiv.c ==========================================================*/
// extern void                Gia_ManOrigIdsInit( Gia_Man_t * p );

// extern void                Gia_ManOrigIdsStart( Gia_Man_t * p );

// extern void                Gia_ManOrigIdsRemap( Gia_Man_t * p, Gia_Man_t * pNew );

// extern Gia_Man_t *         Gia_ManOrigIdsReduce( Gia_Man_t * p, Vec_Int_t * vPairs );

// extern Gia_Man_t *         Gia_ManComputeGiaEquivs( Gia_Man_t * pGia, int nConfs, int fVerbose );

// extern void                Gia_ManEquivFixOutputPairs( Gia_Man_t * p );

// extern int                 Gia_ManCheckTopoOrder( Gia_Man_t * p );

// extern int *               Gia_ManDeriveNexts( Gia_Man_t * p );

// extern void                Gia_ManDeriveReprs( Gia_Man_t * p );

// extern int                 Gia_ManEquivCountLits( Gia_Man_t * p );

// extern int                 Gia_ManEquivCountLitsAll( Gia_Man_t * p );

// extern int                 Gia_ManEquivCountClasses( Gia_Man_t * p );

// extern void                Gia_ManEquivPrintOne( Gia_Man_t * p, int i, int Counter );

// extern void                Gia_ManEquivPrintClasses( Gia_Man_t * p, int fVerbose, float Mem );

// extern Gia_Man_t *         Gia_ManEquivReduce( Gia_Man_t * p, int fUseAll, int fDualOut, int fSkipPhase, int fVerbose );

// extern Gia_Man_t *         Gia_ManEquivReduceAndRemap( Gia_Man_t * p, int fSeq, int fMiterPairs );

// extern int                 Gia_ManEquivSetColors( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManSpecReduce( Gia_Man_t * p, int fDualOut, int fSynthesis, int fReduce, int fSkipSome, int fVerbose );

// extern Gia_Man_t *         Gia_ManSpecReduceInit( Gia_Man_t * p, Abc_Cex_t * pInit, int nFrames, int fDualOut );

// extern Gia_Man_t *         Gia_ManSpecReduceInitFrames( Gia_Man_t * p, Abc_Cex_t * pInit, int nFramesMax, int * pnFrames, int fDualOut, int nMinOutputs );

// extern void                Gia_ManEquivTransform( Gia_Man_t * p, int fVerbose );

// extern void                Gia_ManEquivImprove( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManEquivToChoices( Gia_Man_t * p, int nSnapshots );

// extern int                 Gia_ManCountChoiceNodes( Gia_Man_t * p );

// extern int                 Gia_ManCountChoices( Gia_Man_t * p );

// extern int                 Gia_ManFilterEquivsForSpeculation( Gia_Man_t * pGia, char * pName1, char * pName2, int fLatchA, int fLatchB );

// extern int                 Gia_ManFilterEquivsUsingParts( Gia_Man_t * pGia, char * pName1, char * pName2 );

// extern void                Gia_ManFilterEquivsUsingLatches( Gia_Man_t * pGia, int fFlopsOnly, int fFlopsWith, int fUseRiDrivers );

/*=== giaExist.c =========================================================*/
// extern void                Gia_ManQuantSetSuppStart( Gia_Man_t * p );

// extern void                Gia_ManQuantSetSuppZero( Gia_Man_t * p );

// extern void                Gia_ManQuantSetSuppCi( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManQuantUpdateCiSupp( Gia_Man_t * p, int iObj );

// extern int                 Gia_ManQuantExist( Gia_Man_t * p, int iLit, int(*pFuncCiToKeep)(void *, int), void * pData );

/*=== giaFanout.c =========================================================*/
// extern void                Gia_ObjAddFanout( Gia_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pFanout );

// extern void                Gia_ObjRemoveFanout( Gia_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pFanout );

// extern void                Gia_ManFanoutStart( Gia_Man_t * p );

// extern void                Gia_ManFanoutStop( Gia_Man_t * p );

// extern void                Gia_ManStaticFanoutStart( Gia_Man_t * p );

// extern void                Gia_ManStaticFanoutStop( Gia_Man_t * p );

// extern void                Gia_ManStaticMappingFanoutStart( Gia_Man_t * p );

/*=== giaForce.c =========================================================*/
// extern void                For_ManExperiment( Gia_Man_t * pGia, int nIters, int fClustered, int fVerbose );

/*=== giaFrames.c =========================================================*/
// extern Gia_Man_t *         Gia_ManUnrollDup( Gia_Man_t * p, Vec_Int_t * vLimit );

// extern Vec_Ptr_t *         Gia_ManUnrollAbs( Gia_Man_t * p, int nFrames );

// extern void *              Gia_ManUnrollStart( Gia_Man_t * pAig, Gia_ParFra_t * pPars );

// extern void *              Gia_ManUnrollAdd( void * pMan, int fMax );

// extern void                Gia_ManUnrollStop( void * pMan );

// extern int                 Gia_ManUnrollLastLit( void * pMan );

// extern void                Gia_ManFraSetDefaultParams( Gia_ParFra_t * p );

// extern Gia_Man_t *         Gia_ManFrames( Gia_Man_t * pAig, Gia_ParFra_t * pPars );

// extern Gia_Man_t *         Gia_ManFramesInitSpecial( Gia_Man_t * pAig, int nFrames, int fVerbose );

/*=== giaFront.c ==========================================================*/
// extern Gia_Man_t *         Gia_ManFront( Gia_Man_t * p );

// extern void                Gia_ManFrontTest( Gia_Man_t * p );

/*=== giaFx.c ==========================================================*/
// extern Gia_Man_t *         Gia_ManPerformFx( Gia_Man_t * p, int nNewNodesMax, int LitCountMax, int fReverse, int fVerbose, int fVeryVerbose );

/*=== giaHash.c ===========================================================*/
// extern void                Gia_ManHashAlloc( Gia_Man_t * p );

// extern void                Gia_ManHashStart( Gia_Man_t * p );

// extern void                Gia_ManHashStop( Gia_Man_t * p );

// extern int                 Gia_ManHashXorReal( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManHashMuxReal( Gia_Man_t * p, int iLitC, int iLit1, int iLit0 );

// extern int                 Gia_ManHashAnd( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManHashOr( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManHashXor( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManHashMux( Gia_Man_t * p, int iCtrl, int iData1, int iData0 );

// extern int                 Gia_ManHashMaj( Gia_Man_t * p, int iData0, int iData1, int iData2 );

// extern int                 Gia_ManHashAndTry( Gia_Man_t * p, int iLit0, int iLit1 );

// extern Gia_Man_t *         Gia_ManRehash( Gia_Man_t * p, int fAddStrash );

// extern void                Gia_ManHashProfile( Gia_Man_t * p );

// extern int                 Gia_ManHashLookupInt( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManHashLookup( Gia_Man_t * p, Gia_Obj_t * p0, Gia_Obj_t * p1 );

// extern int                 Gia_ManHashAndMulti( Gia_Man_t * p, Vec_Int_t * vLits );

// extern int                 Gia_ManHashAndMulti2( Gia_Man_t * p, Vec_Int_t * vLits );

// extern int                 Gia_ManHashDualMiter( Gia_Man_t * p, Vec_Int_t * vOuts );

/*=== giaIf.c ===========================================================*/
// extern void                Gia_ManPrintMappingStats( Gia_Man_t * p, char * pDumpFile );

// extern void                Gia_ManPrintPackingStats( Gia_Man_t * p );

// extern void                Gia_ManPrintLutStats( Gia_Man_t * p );

// extern int                 Gia_ManLutFaninCount( Gia_Man_t * p );

// extern int                 Gia_ManLutSizeMax( Gia_Man_t * p );

// extern int                 Gia_ManLutNum( Gia_Man_t * p );

// extern int                 Gia_ManLutLevel( Gia_Man_t * p, int ** ppLevels );

// extern void                Gia_ManLutParams( Gia_Man_t * p, int * pnCurLuts, int * pnCurEdges, int * pnCurLevels );

// extern void                Gia_ManSetRefsMapped( Gia_Man_t * p );

// extern void                Gia_ManSetLutRefs( Gia_Man_t * p );

// extern void                Gia_ManSetIfParsDefault( void * pIfPars );

// extern void                Gia_ManMappingVerify( Gia_Man_t * p );

// extern void                Gia_ManTransferMapping( Gia_Man_t * p, Gia_Man_t * pGia );

// extern void                Gia_ManTransferPacking( Gia_Man_t * p, Gia_Man_t * pGia );

// extern void                Gia_ManTransferTiming( Gia_Man_t * p, Gia_Man_t * pGia );

// extern Gia_Man_t *         Gia_ManPerformMapping( Gia_Man_t * p, void * pIfPars );

// extern Gia_Man_t *         Gia_ManPerformSopBalance( Gia_Man_t * p, int nCutNum, int nRelaxRatio, int fVerbose );

// extern Gia_Man_t *         Gia_ManPerformDsdBalance( Gia_Man_t * p, int nLutSize, int nCutNum, int nRelaxRatio, int fVerbose );

// extern Gia_Man_t *         Gia_ManDupHashMapping( Gia_Man_t * p );

/*=== giaJf.c ===========================================================*/
// extern void                Jf_ManSetDefaultPars( Jf_Par_t * pPars );

// extern Gia_Man_t *         Jf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars );

// extern Gia_Man_t *         Jf_ManDeriveCnf( Gia_Man_t * p, int fCnfObjIds );

/*=== giaIso.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManIsoCanonicize( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManIsoReduce( Gia_Man_t * p, Vec_Ptr_t ** pvPosEquivs, Vec_Ptr_t ** pvPiPerms, int fEstimate, int fDualOut, int fVerbose, int fVeryVerbose );

// extern Gia_Man_t *         Gia_ManIsoReduce2( Gia_Man_t * p, Vec_Ptr_t ** pvPosEquivs, Vec_Ptr_t ** pvPiPerms, int fEstimate, int fBetterQual, int fDualOut, int fVerbose, int fVeryVerbose );

/*=== giaLf.c ===========================================================*/
// extern void                Lf_ManSetDefaultPars( Jf_Par_t * pPars );

// extern Gia_Man_t *         Lf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars );

// extern Gia_Man_t *         Gia_ManPerformLfMapping( Gia_Man_t * p, Jf_Par_t * pPars, int fNormalized );

/*=== giaLogic.c ===========================================================*/
// extern void                Gia_ManTestDistance( Gia_Man_t * p );

// extern void                Gia_ManSolveProblem( Gia_Man_t * pGia, Emb_Par_t * pPars );

/*=== giaMan.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManStart( int nObjsMax );

// extern void                Gia_ManStop( Gia_Man_t * p );

// extern void                Gia_ManStopP( Gia_Man_t ** p );

// extern double              Gia_ManMemory( Gia_Man_t * p );

// extern void                Gia_ManPrintStats( Gia_Man_t * p, Gps_Par_t * pPars );

// extern void                Gia_ManPrintStatsShort( Gia_Man_t * p );

// extern void                Gia_ManPrintMiterStatus( Gia_Man_t * p );

// extern void                Gia_ManPrintStatsMiter( Gia_Man_t * p, int fVerbose );

// extern void                Gia_ManSetRegNum( Gia_Man_t * p, int nRegs );

// extern void                Gia_ManReportImprovement( Gia_Man_t * p, Gia_Man_t * pNew );

// extern void                Gia_ManPrintNpnClasses( Gia_Man_t * p );

// extern void                Gia_ManDumpVerilog( Gia_Man_t * p, char * pFileName, Vec_Int_t * vObjs );

/*=== giaMem.c ===========================================================*/
// extern Gia_MmFixed_t *     Gia_MmFixedStart( int nEntrySize, int nEntriesMax );

// extern void                Gia_MmFixedStop( Gia_MmFixed_t * p, int fVerbose );

// extern char *              Gia_MmFixedEntryFetch( Gia_MmFixed_t * p );

// extern void                Gia_MmFixedEntryRecycle( Gia_MmFixed_t * p, char * pEntry );

// extern void                Gia_MmFixedRestart( Gia_MmFixed_t * p );

// extern int                 Gia_MmFixedReadMemUsage( Gia_MmFixed_t * p );

// extern int                 Gia_MmFixedReadMaxEntriesUsed( Gia_MmFixed_t * p );

// extern Gia_MmFlex_t *      Gia_MmFlexStart();

// extern void                Gia_MmFlexStop( Gia_MmFlex_t * p, int fVerbose );

// extern char *              Gia_MmFlexEntryFetch( Gia_MmFlex_t * p, int nBytes );

// extern void                Gia_MmFlexRestart( Gia_MmFlex_t * p );

// extern int                 Gia_MmFlexReadMemUsage( Gia_MmFlex_t * p );

// extern Gia_MmStep_t *      Gia_MmStepStart( int nSteps );

// extern void                Gia_MmStepStop( Gia_MmStep_t * p, int fVerbose );

// extern char *              Gia_MmStepEntryFetch( Gia_MmStep_t * p, int nBytes );

// extern void                Gia_MmStepEntryRecycle( Gia_MmStep_t * p, char * pEntry, int nBytes );

// extern int                 Gia_MmStepReadMemUsage( Gia_MmStep_t * p );

/*=== giaMf.c ===========================================================*/
// extern void                Mf_ManSetDefaultPars( Jf_Par_t * pPars );

// extern Gia_Man_t *         Mf_ManPerformMapping( Gia_Man_t * pGia, Jf_Par_t * pPars );

// extern void *              Mf_ManGenerateCnf( Gia_Man_t * pGia, int nLutSize, int fCnfObjIds, int fAddOrCla, int fMapping, int fVerbose );

/*=== giaMini.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManReadMiniAig( char * pFileName );

// extern void                Gia_ManWriteMiniAig( Gia_Man_t * pGia, char * pFileName );

// extern Gia_Man_t *         Gia_ManReadMiniLut( char * pFileName );

// extern void                Gia_ManWriteMiniLut( Gia_Man_t * pGia, char * pFileName );

/*=== giaMuxes.c ===========================================================*/
// extern void                Gia_ManCountMuxXor( Gia_Man_t * p, int * pnMuxes, int * pnXors );

// extern void                Gia_ManPrintMuxStats( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupMuxes( Gia_Man_t * p, int Limit );

// extern Gia_Man_t *         Gia_ManDupNoMuxes( Gia_Man_t * p, int fSkipBufs );

/*=== giaPat.c ===========================================================*/
// extern void                Gia_SatVerifyPattern( Gia_Man_t * p, Gia_Obj_t * pRoot, Vec_Int_t * vCex, Vec_Int_t * vVisit );

/*=== giaRetime.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManRetimeForward( Gia_Man_t * p, int nMaxIters, int fVerbose );

/*=== giaSat.c ============================================================*/
// extern int                 Sat_ManTest( Gia_Man_t * pGia, Gia_Obj_t * pObj, int nConfsMax );

/*=== giaScl.c ============================================================*/
// extern int                 Gia_ManSeqMarkUsed( Gia_Man_t * p );

// extern int                 Gia_ManCombMarkUsed( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManCleanup( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManCleanupOutputs( Gia_Man_t * p, int nOutputs );

// extern Gia_Man_t *         Gia_ManSeqCleanup( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManSeqStructSweep( Gia_Man_t * p, int fConst, int fEquiv, int fVerbose );

/*=== giaShow.c ===========================================================*/
// extern void                Gia_ManShow( Gia_Man_t * pMan, Vec_Int_t * vBold, int fAdders, int fFadds, int fPath );

/*=== giaShrink.c ===========================================================*/
// extern Gia_Man_t *         Gia_ManMapShrink4( Gia_Man_t * p, int fKeepLevel, int fVerbose );

// extern Gia_Man_t *         Gia_ManMapShrink6( Gia_Man_t * p, int nFanoutMax, int fKeepLevel, int fVerbose );

/*=== giaSopb.c ============================================================*/
// extern Gia_Man_t *         Gia_ManExtractWindow( Gia_Man_t * p, int LevelMax, int nTimeWindow, int fVerbose );

// extern Gia_Man_t *         Gia_ManPerformSopBalanceWin( Gia_Man_t * p, int LevelMax, int nTimeWindow, int nCutNum, int nRelaxRatio, int fVerbose );

// extern Gia_Man_t *         Gia_ManPerformDsdBalanceWin( Gia_Man_t * p, int LevelMax, int nTimeWindow, int nLutSize, int nCutNum, int nRelaxRatio, int fVerbose );

/*=== giaSort.c ============================================================*/
// extern int *               Gia_SortFloats( float * pArray, int * pPerm, int nSize );

/*=== giaSim.c ============================================================*/
// extern void                Gia_ManSimSetDefaultParams( Gia_ParSim_t * p );

// extern int                 Gia_ManSimSimulate( Gia_Man_t * pAig, Gia_ParSim_t * pPars );

// extern unsigned *          Gia_SimDataExt( Gia_ManSim_t * p, int i );

// extern unsigned *          Gia_SimDataCiExt( Gia_ManSim_t * p, int i );

// extern unsigned *          Gia_SimDataCoExt( Gia_ManSim_t * p, int i );

// extern void                Gia_ManSimInfoInit( Gia_ManSim_t * p );

// extern void                Gia_ManSimInfoTransfer( Gia_ManSim_t * p );

// extern void                Gia_ManSimulateRound( Gia_ManSim_t * p );

// extern void                Gia_ManBuiltInSimStart( Gia_Man_t * p, int nWords, int nObjs );

// extern void                Gia_ManBuiltInSimPerform( Gia_Man_t * p, int iObj );

// extern int                 Gia_ManBuiltInSimCheckOver( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManBuiltInSimCheckEqual( Gia_Man_t * p, int iLit0, int iLit1 );

// extern void                Gia_ManBuiltInSimResimulateCone( Gia_Man_t * p, int iLit0, int iLit1 );

// extern void                Gia_ManBuiltInSimResimulate( Gia_Man_t * p );

// extern int                 Gia_ManBuiltInSimAddPat( Gia_Man_t * p, Vec_Int_t * vPat );

// extern void                Gia_ManIncrSimStart( Gia_Man_t * p, int nWords, int nObjs );

// extern void                Gia_ManIncrSimSet( Gia_Man_t * p, Vec_Int_t * vObjLits );

// extern int                 Gia_ManIncrSimCheckOver( Gia_Man_t * p, int iLit0, int iLit1 );

// extern int                 Gia_ManIncrSimCheckEqual( Gia_Man_t * p, int iLit0, int iLit1 );

/*=== giaSimBase.c ============================================================*/
// extern Vec_Wrd_t *         Gia_ManSimPatSim( Gia_Man_t * p );

/*=== giaSpeedup.c ============================================================*/
// extern float               Gia_ManDelayTraceLut( Gia_Man_t * p );

// extern float               Gia_ManDelayTraceLutPrint( Gia_Man_t * p, int fVerbose );

// extern Gia_Man_t *         Gia_ManSpeedup( Gia_Man_t * p, int Percentage, int Degree, int fVerbose, int fVeryVerbose );

/*=== giaSplit.c ============================================================*/
// extern void                Gia_ManComputeOneWinStart( Gia_Man_t * p, int nAnds, int fReverse );

// extern int                 Gia_ManComputeOneWin( Gia_Man_t * p, int iPivot, Vec_Int_t ** pvRoots, Vec_Int_t ** pvNodes, Vec_Int_t ** pvLeaves, Vec_Int_t ** pvAnds );

/*=== giaStg.c ============================================================*/
// extern void                Gia_ManStgPrint( FILE * pFile, Vec_Int_t * vLines, int nIns, int nOuts, int nStates );

// extern Gia_Man_t *         Gia_ManStgRead( char * pFileName, int kHot, int fVerbose );

/*=== giaSupp.c ============================================================*/

// extern Gia_ManMin_t *      Gia_ManSuppStart( Gia_Man_t * pGia );

// extern void                Gia_ManSuppStop( Gia_ManMin_t * p );

// extern int                 Gia_ManSupportAnd( Gia_ManMin_t * p, int iLit0, int iLit1 );

// extern Gia_Man2Min_t *     Gia_Man2SuppStart( Gia_Man_t * pGia );

// extern void                Gia_Man2SuppStop( Gia_Man2Min_t * p );

// extern int                 Gia_Man2SupportAnd( Gia_Man2Min_t * p, int iLit0, int iLit1 );

/*=== giaSweep.c ============================================================*/
// extern Gia_Man_t *         Gia_ManFraigSweepSimple( Gia_Man_t * p, void * pPars );

// extern Gia_Man_t *         Gia_ManSweepWithBoxes( Gia_Man_t * p, void * pParsC, void * pParsS, int fConst, int fEquiv, int fVerbose, int fVerbEquivs );

// extern void                Gia_ManCheckIntegrityWithBoxes( Gia_Man_t * p );

/*=== giaSweeper.c ============================================================*/
// extern Gia_Man_t *         Gia_SweeperStart( Gia_Man_t * p );

// extern void                Gia_SweeperStop( Gia_Man_t * p );

// extern int                 Gia_SweeperIsRunning( Gia_Man_t * p );

// extern void                Gia_SweeperPrintStats( Gia_Man_t * p );

// extern void                Gia_SweeperSetConflictLimit( Gia_Man_t * p, int nConfMax );

// extern void                Gia_SweeperSetRuntimeLimit( Gia_Man_t * p, int nSeconds );

// extern Vec_Int_t *         Gia_SweeperGetCex( Gia_Man_t * p );

// extern int                 Gia_SweeperProbeCreate( Gia_Man_t * p, int iLit );

// extern int                 Gia_SweeperProbeDelete( Gia_Man_t * p, int ProbeId );

// extern int                 Gia_SweeperProbeUpdate( Gia_Man_t * p, int ProbeId, int iLitNew );

// extern int                 Gia_SweeperProbeLit( Gia_Man_t * p, int ProbeId );

// extern Vec_Int_t *         Gia_SweeperCollectValidProbeIds( Gia_Man_t * p );

// extern int                 Gia_SweeperCondPop( Gia_Man_t * p );

// extern void                Gia_SweeperCondPush( Gia_Man_t * p, int ProbeId );

// extern Vec_Int_t *         Gia_SweeperCondVector( Gia_Man_t * p );

// extern int                 Gia_SweeperCondCheckUnsat( Gia_Man_t * p );

// extern int                 Gia_SweeperCheckEquiv( Gia_Man_t * p, int ProbeId1, int ProbeId2 );

// extern Gia_Man_t *         Gia_SweeperExtractUserLogic( Gia_Man_t * p, Vec_Int_t * vProbeIds, Vec_Ptr_t * vInNames, Vec_Ptr_t * vOutNames );

// extern void                Gia_SweeperLogicDump( Gia_Man_t * p, Vec_Int_t * vProbeIds, int fDumpConds, char * pFileName );

// extern Gia_Man_t *         Gia_SweeperCleanup( Gia_Man_t * p, char * pCommLime );

// extern Vec_Int_t *         Gia_SweeperGraft( Gia_Man_t * pDst, Vec_Int_t * vProbes, Gia_Man_t * pSrc );

// extern int                 Gia_SweeperFraig( Gia_Man_t * p, Vec_Int_t * vProbeIds, char * pCommLime, int nWords, int nConfs, int fVerify, int fVerbose );

// extern int                 Gia_SweeperRun( Gia_Man_t * p, Vec_Int_t * vProbeIds, char * pCommLime, int fVerbose );

/*=== giaSwitch.c ============================================================*/
// extern float               Gia_ManEvaluateSwitching( Gia_Man_t * p );

// extern float               Gia_ManComputeSwitching( Gia_Man_t * p, int nFrames, int nPref, int fProbOne );

// extern Vec_Int_t *         Gia_ManComputeSwitchProbs( Gia_Man_t * pGia, int nFrames, int nPref, int fProbOne );

// extern Vec_Flt_t *         Gia_ManPrintOutputProb( Gia_Man_t * p );

/*=== giaTim.c ===========================================================*/
// extern int                 Gia_ManBoxNum( Gia_Man_t * p );

// extern int                 Gia_ManRegBoxNum( Gia_Man_t * p );

// extern int                 Gia_ManNonRegBoxNum( Gia_Man_t * p );

// extern int                 Gia_ManBlackBoxNum( Gia_Man_t * p );

// extern int                 Gia_ManBoxCiNum( Gia_Man_t * p );

// extern int                 Gia_ManBoxCoNum( Gia_Man_t * p );

// extern int                 Gia_ManClockDomainNum( Gia_Man_t * p );

// extern int                 Gia_ManIsSeqWithBoxes( Gia_Man_t * p );

// extern int                 Gia_ManIsNormalized( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManOrderWithBoxes( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupNormalize( Gia_Man_t * p, int fHashMapping );

// extern Gia_Man_t *         Gia_ManDupUnnormalize( Gia_Man_t * p );

// extern Gia_Man_t *         Gia_ManDupUnshuffleInputs( Gia_Man_t * p );

// extern int                 Gia_ManLevelWithBoxes( Gia_Man_t * p );

// extern int                 Gia_ManLutLevelWithBoxes( Gia_Man_t * p );

// extern void *              Gia_ManUpdateTimMan( Gia_Man_t * p, Vec_Int_t * vBoxPres );

// extern void *              Gia_ManUpdateTimMan2( Gia_Man_t * p, Vec_Int_t * vBoxesLeft, int nTermsDiff );

// extern Gia_Man_t *         Gia_ManUpdateExtraAig( void * pTime, Gia_Man_t * pAig, Vec_Int_t * vBoxPres );

// extern Gia_Man_t *         Gia_ManUpdateExtraAig2( void * pTime, Gia_Man_t * pAig, Vec_Int_t * vBoxesLeft );

// extern Gia_Man_t *         Gia_ManDupCollapse( Gia_Man_t * p, Gia_Man_t * pBoxes, Vec_Int_t * vBoxPres, int fSeq );

// extern int                 Gia_ManVerifyWithBoxes( Gia_Man_t * pGia, int nBTLimit, int nTimeLim, int fSeq, int fDumpFiles, int fVerbose, char * pFileSpec );

/*=== giaTruth.c ===========================================================*/
// extern word                Gia_LutComputeTruth6( Gia_Man_t * p, int iObj, Vec_Wrd_t * vTruths );

// extern word                Gia_ObjComputeTruthTable6Lut( Gia_Man_t * p, int iObj, Vec_Wrd_t * vTemp );

// extern word                Gia_ObjComputeTruthTable6( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vSupp, Vec_Wrd_t * vTruths );

// extern word                Gia_ObjComputeTruth6Cis( Gia_Man_t * p, int iLit, Vec_Int_t * vSupp, Vec_Wrd_t * vTemp );

// extern void                Gia_ObjCollectInternal( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern word *              Gia_ObjComputeTruthTable( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ObjComputeTruthTableStart( Gia_Man_t * p, int nVarsMax );

// extern void                Gia_ObjComputeTruthTableStop( Gia_Man_t * p );

// extern word *              Gia_ObjComputeTruthTableCut( Gia_Man_t * p, Gia_Obj_t * pObj, Vec_Int_t * vLeaves );

/*=== giaTsim.c ============================================================*/
// extern Gia_Man_t *         Gia_ManReduceConst( Gia_Man_t * pAig, int fVerbose );

/*=== giaUtil.c ===========================================================*/
// extern unsigned            Gia_ManRandom( int fReset );

// extern word                Gia_ManRandomW( int fReset );

// extern void                Gia_ManRandomInfo( Vec_Ptr_t * vInfo, int iInputStart, int iWordStart, int iWordStop );

// extern char *              Gia_TimeStamp();

// extern char *              Gia_FileNameGenericAppend( char * pBase, char * pSuffix );

// extern void                Gia_ManIncrementTravId( Gia_Man_t * p );

// extern void                Gia_ManCleanMark01( Gia_Man_t * p );

// extern void                Gia_ManSetMark0( Gia_Man_t * p );

// extern void                Gia_ManCleanMark0( Gia_Man_t * p );

// extern void                Gia_ManCheckMark0( Gia_Man_t * p );

// extern void                Gia_ManSetMark1( Gia_Man_t * p );

// extern void                Gia_ManCleanMark1( Gia_Man_t * p );

// extern void                Gia_ManCheckMark1( Gia_Man_t * p );

// extern void                Gia_ManCleanValue( Gia_Man_t * p );

// extern void                Gia_ManCleanLevels( Gia_Man_t * p, int Size );

// extern void                Gia_ManCleanTruth( Gia_Man_t * p );

// extern void                Gia_ManFillValue( Gia_Man_t * p );

// extern void                Gia_ObjSetPhase( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManSetPhase( Gia_Man_t * p );

// extern void                Gia_ManSetPhasePattern( Gia_Man_t * p, Vec_Int_t * vCiValues );

// extern void                Gia_ManSetPhase1( Gia_Man_t * p );

// extern void                Gia_ManCleanPhase( Gia_Man_t * p );

// extern int                 Gia_ManCheckCoPhase( Gia_Man_t * p );

// extern int                 Gia_ManLevelNum( Gia_Man_t * p );

// extern int                 Gia_ManLevelRNum( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManGetCiLevels( Gia_Man_t * p );

// extern int                 Gia_ManSetLevels( Gia_Man_t * p, Vec_Int_t * vCiLevels );

// extern Vec_Int_t *         Gia_ManReverseLevel( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManRequiredLevel( Gia_Man_t * p );

// extern void                Gia_ManCreateValueRefs( Gia_Man_t * p );

// extern void                Gia_ManCreateRefs( Gia_Man_t * p );

// extern int *               Gia_ManCreateMuxRefs( Gia_Man_t * p );

// extern int                 Gia_ManCrossCut( Gia_Man_t * p, int fReverse );

// extern Vec_Int_t *         Gia_ManCollectPoIds( Gia_Man_t * p );

// extern int                 Gia_ObjIsMuxType( Gia_Obj_t * pNode );

// extern int                 Gia_ObjRecognizeExor( Gia_Obj_t * pObj, Gia_Obj_t ** ppFan0, Gia_Obj_t ** ppFan1 );

// extern Gia_Obj_t *         Gia_ObjRecognizeMux( Gia_Obj_t * pNode, Gia_Obj_t ** ppNodeT, Gia_Obj_t ** ppNodeE );

// extern int                 Gia_ObjRecognizeMuxLits( Gia_Man_t * p, Gia_Obj_t * pNode, int * iLitT, int * iLitE );

// extern int                 Gia_NodeMffcSize( Gia_Man_t * p, Gia_Obj_t * pNode );

// extern int                 Gia_NodeMffcSizeMark( Gia_Man_t * p, Gia_Obj_t * pNode );

// extern int                 Gia_NodeMffcSizeSupp( Gia_Man_t * p, Gia_Obj_t * pNode, Vec_Int_t * vSupp );

// extern int                 Gia_NodeMffcMapping( Gia_Man_t * p );

// extern int                 Gia_ManHasDangling( Gia_Man_t * p );

// extern int                 Gia_ManMarkDangling( Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManGetDangling( Gia_Man_t * p );

// extern void                Gia_ObjPrint( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManPrint( Gia_Man_t * p );

// extern void                Gia_ManPrintCo( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManPrintCone( Gia_Man_t * p, Gia_Obj_t * pObj, int * pLeaves, int nLeaves, Vec_Int_t * vNodes );

// extern void                Gia_ManPrintConeMulti( Gia_Man_t * p, Vec_Int_t * vObjs, Vec_Int_t * vLeaves, Vec_Int_t * vNodes );

// extern void                Gia_ManPrintCone2( Gia_Man_t * p, Gia_Obj_t * pObj );

// extern void                Gia_ManInvertConstraints( Gia_Man_t * pAig );

// extern void                Gia_ManInvertPos( Gia_Man_t * pAig );

// extern int                 Gia_ManCompare( Gia_Man_t * p1, Gia_Man_t * p2 );

// extern void                Gia_ManMarkFanoutDrivers( Gia_Man_t * p );

// extern void                Gia_ManSwapPos( Gia_Man_t * p, int i );

// extern Vec_Int_t *         Gia_ManSaveValue( Gia_Man_t * p );

// extern void                Gia_ManLoadValue( Gia_Man_t * p, Vec_Int_t * vValues );

// extern Vec_Int_t *         Gia_ManFirstFanouts( Gia_Man_t * p );

// extern int                 Gia_ManCheckSuppOverlap( Gia_Man_t * p, int iNode1, int iNode2 );

// extern int                 Gia_ManCountPisWithFanout( Gia_Man_t * p );

// extern int                 Gia_ManCountPosWithNonZeroDrivers( Gia_Man_t * p );

// extern void                Gia_ManUpdateCopy( Vec_Int_t * vCopy, Gia_Man_t * p );

// extern Vec_Int_t *         Gia_ManComputeDistance( Gia_Man_t * p, int iObj, Vec_Int_t * vObjs, int fVerbose );

/*=== giaCTas.c ===========================================================*/

// extern Tas_Man_t *         Tas_ManAlloc( Gia_Man_t * pAig, int nBTLimit );

// extern void                Tas_ManStop( Tas_Man_t * p );

// extern Vec_Int_t *         Tas_ReadModel( Tas_Man_t * p );

// extern void                Tas_ManSatPrintStats( Tas_Man_t * p );

// extern int                 Tas_ManSolve( Tas_Man_t * p, Gia_Obj_t * pObj, Gia_Obj_t * pObj2 );

// extern int                 Tas_ManSolveArray( Tas_Man_t * p, Vec_Ptr_t * vObjs );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
