package cut

/**CFile****************************************************************

  FileName    [cut.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [K-feasible cut computation package.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]
  
  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - June 20, 2005.]

  Revision    [$Id: .h,v 1.00 2005/06/20 00:00:00 alanmi Exp $]

***********************************************************************/





////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////



//ABC_NAMESPACE_HEADER_START


const CUT_SIZE_MIN =    3      // the min K of the K-feasible cut computation
const CUT_SIZE_MAX =   12      // the max K of the K-feasible cut computation

const CUT_SHIFT =       8      // the number of bits for storing latch number in the cut leaves
const CUT_MASK =        0xFF   // the mask to get the stored latch number

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type Cut_Man_t =         Cut_ManStruct_t_;
type Cut_Oracle_t =      Cut_OracleStruct_t_;
type Cut_Cut_t =         Cut_CutStruct_t_;
        

type Cut_Params_t struct {
    nVarsMax                int32;          // the max cut size ("k" of the k-feasible cuts)
    nKeepMax                int32;          // the max number of cuts kept at a node
    nIdsMax                int32;           // the max number of IDs of cut objects
    nBitShift                int32;         // the number of bits used for the latch counter of an edge
    nCutSet                int32;           // the number of nodes in the cut set
    fTruth                int32;            // compute truth tables
    fFilter                int32;           // filter dominated cuts
    fSeq                int32;              // compute sequential cuts
    fDrop                int32;             // drop cuts on the fly
    fDag                int32;              // compute only DAG cuts
    fTree                int32;             // compute only tree cuts
    fGlobal                int32;           // compute only global cuts
    fLocal                int32;            // compute only local cuts
    fRecord                int32;           // record the cut computation flow
    fRecordAig                int32;        // record the cut functions
    fFancy                int32;            // perform fancy computations
    fMap                int32;              // computes delay of FPGA mapping with cuts
    fAdjust                int32;           // removed useless fanouts of XORs/MUXes
    fNpnSave                int32;          // enables dumping 6-input truth tables
    fVerbose                int32;          // the verbosiness flag
}

type Cut_CutStruct_t_ struct {
    Num0 uint32; /* C bitfield: 11 bits; review packing and truncation */   // temporary number
    Num1 uint32; /* C bitfield: 11 bits; review packing and truncation */   // temporary number
    fSimul uint32; /* C bitfield: 1 bits; review packing and truncation */   // the value of cut's output at 000.. pattern
    fCompl uint32; /* C bitfield: 1 bits; review packing and truncation */   // the cut is complemented
    nVarsMax uint32; /* C bitfield: 4 bits; review packing and truncation */   // the max number of vars [4-6]
    nLeaves uint32; /* C bitfield: 4 bits; review packing and truncation */   // the number of leaves [4-6]
    uSign           uint32;             // the signature
    uCanon0           uint32;           // the canonical form 
    uCanon1           uint32;           // the canonical form 
    pNext         *Cut_Cut_t;             // the next cut in the list
    pLeaves [0]int32;        // the array of leaves
}

func Cut_CutReadLeaveNum(p *Cut_Cut_t) int32 {  return p.nLeaves;   }
func Cut_CutReadLeaves(p *Cut_Cut_t) *int32 {  return p.pLeaves;   }
func Cut_CutReadTruth(p *Cut_Cut_t) *uint32 {  return (unsigned *)(p.pLeaves + p.nVarsMax); }
func Cut_CutWriteTruth(p *Cut_Cut_t, puTruth *uint32) { 
    var i int32;
    for  i = (p.nVarsMax <= 5) ? 0 : ((1 << (p.nVarsMax - 5)) - 1); i >= 0; i--  {
        p.pLeaves[p.nVarsMax + i] = int32(puTruth[i]);
    }
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== cutApi.c ==========================================================*/
// extern Cut_Cut_t *      Cut_NodeReadCutsNew( Cut_Man_t * p, int Node );

// extern Cut_Cut_t *      Cut_NodeReadCutsOld( Cut_Man_t * p, int Node );

// extern Cut_Cut_t *      Cut_NodeReadCutsTemp( Cut_Man_t * p, int Node );

// extern void             Cut_NodeWriteCutsNew( Cut_Man_t * p, int Node, Cut_Cut_t * pList );

// extern void             Cut_NodeWriteCutsOld( Cut_Man_t * p, int Node, Cut_Cut_t * pList );

// extern void             Cut_NodeWriteCutsTemp( Cut_Man_t * p, int Node, Cut_Cut_t * pList );

// extern void             Cut_NodeSetTriv( Cut_Man_t * p, int Node );

// extern void             Cut_NodeTryDroppingCuts( Cut_Man_t * p, int Node );

// extern void             Cut_NodeFreeCuts( Cut_Man_t * p, int Node );

/*=== cutCut.c ==========================================================*/
// extern void             Cut_CutPrint( Cut_Cut_t * pCut, int fSeq );

// extern void             Cut_CutPrintList( Cut_Cut_t * pList, int fSeq );

// extern int              Cut_CutCountList( Cut_Cut_t * pList );

/*=== cutMan.c ==========================================================*/
// extern Cut_Man_t *      Cut_ManStart( Cut_Params_t * pParams );

// extern void             Cut_ManStop( Cut_Man_t * p );

// extern void             Cut_ManPrintStats( Cut_Man_t * p );

// extern void             Cut_ManPrintStatsToFile( Cut_Man_t * p, char * pFileName, abctime TimeTotal );

// extern void             Cut_ManSetFanoutCounts( Cut_Man_t * p, Vec_Int_t * vFanCounts );

// extern void             Cut_ManSetNodeAttrs( Cut_Man_t * p, Vec_Int_t * vFanCounts );

// extern int              Cut_ManReadVarsMax( Cut_Man_t * p );

// extern Cut_Params_t *   Cut_ManReadParams( Cut_Man_t * p );

// extern Vec_Int_t *      Cut_ManReadNodeAttrs( Cut_Man_t * p );

// extern void             Cut_ManIncrementDagNodes( Cut_Man_t * p );

/*=== cutNode.c ==========================================================*/
// extern Cut_Cut_t *      Cut_NodeComputeCuts( Cut_Man_t * p, int Node, int Node0, int Node1, int fCompl0, int fCompl1, int fTriv, int TreeCode );
 
// extern Cut_Cut_t *      Cut_NodeUnionCuts( Cut_Man_t * p, Vec_Int_t * vNodes );

// extern Cut_Cut_t *      Cut_NodeUnionCutsSeq( Cut_Man_t * p, Vec_Int_t * vNodes, int CutSetNum, int fFirst );

// extern int              Cut_ManMappingArea_rec( Cut_Man_t * p, int Node );

/*=== cutSeq.c ==========================================================*/
// extern void             Cut_NodeComputeCutsSeq( Cut_Man_t * p, int Node, int Node0, int Node1, int fCompl0, int fCompl1, int nLat0, int nLat1, int fTriv, int CutSetNum );

// extern void             Cut_NodeNewMergeWithOld( Cut_Man_t * p, int Node );

// extern int              Cut_NodeTempTransferToNew( Cut_Man_t * p, int Node, int CutSetNum );

// extern void             Cut_NodeOldTransferToNew( Cut_Man_t * p, int Node );

/*=== cutOracle.c ==========================================================*/
// extern Cut_Oracle_t *   Cut_OracleStart( Cut_Man_t * pMan );

// extern void             Cut_OracleStop( Cut_Oracle_t * p );

// extern void             Cut_OracleSetFanoutCounts( Cut_Oracle_t * p, Vec_Int_t * vFanCounts );

// extern int              Cut_OracleReadDrop( Cut_Oracle_t * p );

// extern void             Cut_OracleNodeSetTriv( Cut_Oracle_t * p, int Node );

// extern Cut_Cut_t *      Cut_OracleComputeCuts( Cut_Oracle_t * p, int Node, int Node0, int Node1, int fCompl0, int fCompl1 );

// extern void             Cut_OracleTryDroppingCuts( Cut_Oracle_t * p, int Node );

/*=== cutTruth.c ==========================================================*/
// extern void             Cut_TruthNCanonicize( Cut_Cut_t * pCut );

/*=== cutPre22.c ==========================================================*/
// extern void             Cut_CellPrecompute();

// extern void             Cut_CellLoad();

// extern int              Cut_CellIsRunning();

// extern void             Cut_CellDumpToFile();

// extern int              Cut_CellTruthLookup( unsigned * pTruth, int nVars );




//ABC_NAMESPACE_HEADER_END





////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////

