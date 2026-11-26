package mfs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Port of mfsInt.h.

const MFS_FANIN_MAX = 12

type Mfs_Man_t struct {
	// input data
	pPars     *Mfs_Par_t
	pNtk      *abc.Network
	pCare     *aig.Graph
	vSuppsInv [][]int32
	nFaninMax int32
	// intermeditate data for the node
	vRoots       []*abc.Object // the roots of the window
	vSupp        []*abc.Object // the support of the window
	vNodes       []*abc.Object // the internal nodes of the window
	vDivs        []*abc.Object // the divisors of the node
	vProjVarsCnf []int32       // the projection variables
	vProjVarsSat []int32       // the projection variables
	// intermediate simulation data
	vDivCexes [][]uint32 // the counter-example for dividors
	nDivWords int32      // the number of words
	nCexes    int32      // the numbe rof current counter-examples
	nSatCalls int32
	nSatCexes int32
	// used for bidecomposition
	vTruth            *vec.Vec_Int_t
	pManDec           *bdc.Bdc_Man_t
	nNodesDec         int32
	nNodesGained      int32
	nNodesGainedLevel int32
	// solving data
	pAigWin       *aig.Graph     // window AIG with constraints
	pCnf          *cnf.Cnf_Dat_t // the CNF for the window
	pSat          *bsat.Solver   // the SAT solver used
	pMan          *bsat.Int_Man_t
	vMem          []uint32       // memory for intermediate SOPs
	vLevels       *vec.Vec_Vec_t // levelized structure for updating
	vMfsFanins    []*abc.Object  // the new set of fanins
	nTotConfLim   int32          // total conflict limit
	nTotConfLevel int32          // total conflicts on this level
	// switching activity
	vProbs *vec.Vec_Int_t
	// the result of solving
	nFanins int32 // the number of fanins
	nWords  int32 // the number of words
	nCares  int32 // the number of care minterms
	uCare   [1 << (MFS_FANIN_MAX - 5)]uint32
	// performance statistics
	nTryRemoves    int32
	nTryResubs     int32
	nRemoves       int32
	nResubs        int32
	nNodesTried    int32
	nNodesResub    int32
	nMintsCare     int32
	nMintsTotal    int32
	nNodesBad      int32
	nTotalDivs     int32
	nTimeOuts      int32
	nTimeOutsLevel int32
	nDcMints       int32
	nMaxDivs       int32
	dTotalRatios   float64
	// node/edge stats
	nTotalNodesBeg    int32
	nTotalNodesEnd    int32
	nTotalEdgesBeg    int32
	nTotalEdgesEnd    int32
	TotalSwitchingBeg float32
	TotalSwitchingEnd float32
	// statistics
	timeWin, timeDiv, timeAig, timeGia, timeCnf, timeSat, timeInt, timeTotal int64

	// pCopy / pNext of the network objects (AIG literals of the window),
	// indexed by object ID; C keeps them in the Abc_Obj_t fields.
	objCopy []aig.Edge
	objNext []aig.Edge
	// pData of the CIs when a care set is used (CI index), indexed by ID.
	ciIndex []int32
}

func Abc_MfsObjProb(p *Mfs_Man_t, pObj *abc.Object) float32 {
	if p.vProbs != nil && pObj.ID < vec.Vec_IntSize(p.vProbs) {
		return util.Abc_Int2Float(vec.Vec_IntEntry(p.vProbs, pObj.ID))
	}
	return 0.0
}

func (p *Mfs_Man_t) ensureObj(id int) {
	if id < len(p.objCopy) {
		return
	}
	n := max(id+1, int(abc.Abc_NtkObjNumMax(p.pNtk)))
	p.objCopy = append(p.objCopy, make([]aig.Edge, n-len(p.objCopy))...)
	p.objNext = append(p.objNext, make([]aig.Edge, n-len(p.objNext))...)
}
