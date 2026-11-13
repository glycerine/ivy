package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"time"
)

type Ssw_Man_t = Ssw_Man_t_ // signal correspondence manager
type Ssw_Frm_t = Ssw_Frm_t_ // unrolled frames manager
type Ssw_Sat_t = Ssw_Sat_t_ // SAT solver manager
type Ssw_Cla_t = Ssw_Cla_t_ // equivalence classe manager

type Ssw_Man_t_ struct {
	// parameters
	pPars   *Ssw_Pars_t // parameters
	nFrames int32       // for quick lookup
	// AIGs used in the package
	pAig          *aig.Graph // user-given AIG
	pFrames       *aig.Graph // final AIG
	pNodeToFrames []aig.Edge // mapping of AIG nodes into FRAIG nodes
	// equivalence classes
	ppClasses *Ssw_Cla_t // equivalence classes of nodes
	fRefined  int32      // is set to 1 when refinement happens
	// SAT solving
	pMSatBmc *Ssw_Sat_t // SAT manager for base case
	pMSat    *Ssw_Sat_t // SAT manager for inductive case
	// SAT solving (latch corr only)
	vSimInfo       *vec.Vec_Ptr_t // simulation information for the framed PIs
	nPatterns      int32          // the number of patterns saved
	nSimRounds     int32          // the number of simulation rounds performed
	nCallsCount    int32          // the number of calls in this round
	nCallsDelta    int32          // the number of calls to skip
	nCallsSat      int32          // the number of SAT calls in this round
	nCallsUnsat    int32          // the number of UNSAT calls in this round
	nRecycleCalls  int32          // the number of calls since last recycling
	nRecycles      int32          // the number of time SAT solver was recycled
	nRecyclesTotal int32          // the number of time SAT solver was recycled
	nVarsMax       int32          // the maximum variables in the solver
	nCallsMax      int32          // the maximum number of SAT calls
	// uniqueness
	vCommon        *vec.Vec_Ptr_t // the set of common variables in the logic cones
	iOutputLit     int32          // the output literal of the uniqueness constraint
	vDiffPairs     *vec.Vec_Int_t // is set to 1 if reg pair can be diff
	nUniques       int32          // the number of uniqueness constraints used
	nUniquesAdded  int32          // useful uniqueness constraints
	nUniquesUseful int32          // useful uniqueness constraints
	// dynamic constraint addition
	nSRMiterMaxId int32          // max ID after which the last frame begins
	vNewLos       *vec.Vec_Ptr_t // new time frame LOs of to constrain
	vNewPos       *vec.Vec_Int_t // new time frame POs of to add constraints
	pVisited      []int32        // flags to label visited nodes in each frame
	nVisCounter   int32          // the traversal ID
	// sequential simulation
	pSml          *Ssw_Sml_t     // the simulator
	iNodeStart    int32          // the first node considered
	iNodeLast     int32          // the last node considered
	vResimConsts  *vec.Vec_Ptr_t // resimulation constants
	vResimClasses *vec.Vec_Ptr_t // resimulation classes
	vInits        *vec.Vec_Int_t // the init values of primary inputs under constraints
	// counter example storage
	nPatWords int32    // the number of words in the counter example
	pPatWords []uint32 // the counter example
	// constraints
	nConstrTotal   int32 // the number of total constraints
	nConstrReduced int32 // the number of reduced constraints
	nStrangers     int32 // the number of strange situations
	// SAT calls statistics
	nSatCalls      int32 // the number of SAT calls
	nSatProof      int32 // the number of proofs
	nSatFailsReal  int32 // the number of timeouts
	nSatCallsUnsat int32 // the number of unsat SAT calls
	nSatCallsSat   int32 // the number of sat SAT calls
	// node/register/lit statistics
	nLitsBeg  int32
	nLitsEnd  int32
	nNodesBeg int32
	nNodesEnd int32
	nRegsBeg  int32
	nRegsEnd  int32
	// equiv statistis
	nConesTotal   int32
	nConesConstr  int32
	nEquivsTotal  int32
	nEquivsConstr int32
	nNodesBegC    int32
	nNodesEndC    int32
	nRegsBegC     int32
	nRegsEndC     int32
	// runtime stats
	timeBmc       time.Duration // bounded model checking
	timeReduce    time.Duration // speculative reduction
	timeMarkCones time.Duration // marking the cones not to be refined
	timeSimSat    time.Duration // simulation of the counter-examples
	timeSat       time.Duration // solving SAT
	timeSatSat    time.Duration // sat
	timeSatUnsat  time.Duration // unsat
	timeSatUndec  time.Duration // undecided
	timeOther     time.Duration // other runtime
	timeTotal     time.Duration // total runtime
}

// internal SAT manager
type Ssw_Sat_t_ struct {
	pAig         *aig.Graph     // the AIG manager
	fPolarFlip   int32          // flips polarity
	pSat         *bsat.Solver   // recyclable SAT solver
	nSatVars     int32          // the counter of SAT variables
	vSatVars     *vec.Vec_Int_t // mapping of each node into its SAT var
	vFanins      *vec.Vec_Ptr_t // fanins of the CNF node
	vUsedPis     *vec.Vec_Ptr_t // the PIs with SAT variables
	nSolverCalls int32          // the total number of SAT calls
}

// internal frames manager
type Ssw_Frm_t_ struct {
	pAig     *aig.Graph     // user-given AIG
	nObjs    int32          // offset in terms of AIG nodes
	nFrames  int32          // the number of frames in current unrolling
	pFrames  *aig.Graph     // unrolled AIG
	vAig2Frm *vec.Vec_Ptr_t // mapping of AIG nodes into frame nodes
}

// SAT variable numbers are indexed by AIG node ID.
func Ssw_ObjSatNum(p *Ssw_Sat_t, obj aig.Edge) int32 {
	return vec.Vec_IntGetEntry(p.vSatVars, obj.NodeID())
}
func Ssw_ObjSetSatNum(p *Ssw_Sat_t, obj aig.Edge, num int32) {
	vec.Vec_IntSetEntry(p.vSatVars, obj.NodeID(), num)
}
func Ssw_ObjIsConst1Cand(g *aig.Graph, obj aig.Edge) int32 {
	if aig.Aig_ObjRepr(g, obj) == 0 {
		return 1
	}
	return 0
}
func Ssw_ObjSetConst1Cand(g *aig.Graph, obj aig.Edge) {
	if Ssw_ObjIsConst1Cand(g, obj) != 0 {
		panic("object is already a constant candidate")
	}
	aig.Aig_ObjSetRepr(g, obj, 0)
}
func Ssw_ObjFrame(p *Ssw_Man_t, obj aig.Edge, frame int32) aig.Edge {
	return p.pNodeToFrames[int(p.nFrames)*obj.NodeID()+int(frame)]
}
func Ssw_ObjSetFrame(p *Ssw_Man_t, obj aig.Edge, frame int32, node aig.Edge) {
	p.pNodeToFrames[int(p.nFrames)*obj.NodeID()+int(frame)] = node
}
func frameChild(g *aig.Graph, obj aig.Edge, which int, lookup func(aig.Edge) aig.Edge) aig.Edge {
	if obj.Complemented() {
		panic("complemented object")
	}
	n := g.Node(obj)
	if n.Kind == aig.Constant || n.Kind == aig.Input || (which == 1 && (n.Kind == aig.Buffer || n.Kind == aig.Output)) {
		return aig.InvalidEdge
	}
	child := n.Fanins[which]
	mapped := lookup(child &^ 1)
	if mapped != aig.InvalidEdge && child.Complemented() {
		mapped = mapped.Not()
	}
	return mapped
}
func Ssw_ObjChild0Fra(p *Ssw_Man_t, obj aig.Edge, frame int32) aig.Edge {
	return frameChild(p.pAig, obj, 0, func(child aig.Edge) aig.Edge { return Ssw_ObjFrame(p, child, frame) })
}
func Ssw_ObjChild1Fra(p *Ssw_Man_t, obj aig.Edge, frame int32) aig.Edge {
	return frameChild(p.pAig, obj, 1, func(child aig.Edge) aig.Edge { return Ssw_ObjFrame(p, child, frame) })
}
func Ssw_ObjFrame_(p *Ssw_Frm_t, obj aig.Edge, frame int32) aig.Edge {
	x := vec.Vec_PtrGetEntry(p.vAig2Frm, int(p.nObjs*frame)+obj.NodeID())
	if x == nil {
		return aig.InvalidEdge
	}
	return x.(aig.Edge)
}
func Ssw_ObjSetFrame_(p *Ssw_Frm_t, obj aig.Edge, frame int32, node aig.Edge) {
	vec.Vec_PtrSetEntry(p.vAig2Frm, int(p.nObjs*frame)+obj.NodeID(), node)
}
func Ssw_ObjChild0Fra_(p *Ssw_Frm_t, obj aig.Edge, frame int32) aig.Edge {
	return frameChild(p.pAig, obj, 0, func(child aig.Edge) aig.Edge { return Ssw_ObjFrame_(p, child, frame) })
}
func Ssw_ObjChild1Fra_(p *Ssw_Frm_t, obj aig.Edge, frame int32) aig.Edge {
	return frameChild(p.pAig, obj, 1, func(child aig.Edge) aig.Edge { return Ssw_ObjFrame_(p, child, frame) })
}
