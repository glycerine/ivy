package bmc

// Line-by-line port of src/sat/bmc/bmcBmc3.c: the "bmc3" engine
// (Saig_ManBmcScalable) with dynamic unrolling, ternary simulation, the
// cut-based CNF of Cnf_DeriveMappingArray hashed by (truth, leaves), and one
// incremental SAT solver (sat_solver, or satoko/glucose with -s/-g).

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
	"github.com/glycerine/abcgo/sat/glucose"
	"github.com/glycerine/abcgo/sat/satoko"
)

// Saig_ParBmc_t contains options and accumulated status for scalable BMC.
type Saig_ParBmc_t struct {
	NStart         int32 // starting timeframe
	NFramesMax     int32 // maximum number of timeframes
	NConfLimit     int32 // maximum number of conflicts at a node
	NConfLimitJump int32 // maximum number of conflicts after jumping
	NFramesJump    int32 // the number of timeframes to jump
	NTimeOut       int32 // approximate timeout in seconds
	NTimeOutGap    int32 // approximate timeout in seconds since the last change
	NTimeOutOne    int32 // timeout per output in multi-output solving
	NPisAbstract   int32 // the number of PIs to abstract
	FSolveAll      int32 // does not stop at the first SAT output
	FStoreCex      int32 // enable storing CEXes in the MO mode
	FUseBridge     int32 // use bridge interface
	FDropSatOuts   int32 // replace sat outputs by constant 0
	nFfToAddMax    int32 // max number of flops to add during CBA
	fSkipRand      int32 // skip random decisions
	FNoRestarts    int32 // disables periodic restarts
	FUseSatoko     int32 // enables using Satoko
	FUseGlucose    int32 // enables using Glucose 3.0
	NLearnedStart  int32 // starting learned clause limit
	NLearnedDelta  int32 // delta of learned clause limit
	NLearnedPerce  int32 // ratio of learned clause limit
	FVerbose       int32 // enables verbose output
	FNotVerbose    int32 // skip line-by-line print-out
	PLogFileName   string
	FSilent        int32 // completely silent
	IFrame         int32 // explored up to this frame
	NFailOuts      int32 // the number of failed outputs
	NDropOuts      int32 // the number of dropped outputs
	timeLastSolved int64 // the time when the last output was solved (Abc_Clock)
	pFuncOnFail    func(int32, *util.Abc_Cex_t) int32
	RunId          int32
	pFuncStop      func(int32) int32
	pBridgeWriter  io.Writer // nil selects stdout, matching ABC's bridge stream
}

// Saig_ParBmcSetDefaultParams sets the default parameters.
func Saig_ParBmcSetDefaultParams(p *Saig_ParBmc_t) {
	if p == nil {
		return
	}
	*p = Saig_ParBmc_t{}
	p.NStart = 0         // maximum number of timeframes
	p.NFramesMax = 0     // maximum number of timeframes
	p.NConfLimit = 0     // maximum number of conflicts at a node
	p.NConfLimitJump = 0 // maximum number of conflicts after jumping
	p.NFramesJump = 0    // the number of tiemframes to jump
	p.NTimeOut = 0       // approximate timeout in seconds
	p.NTimeOutGap = 0    // time since the last CEX found
	p.NPisAbstract = 0   // the number of PIs to abstract
	p.FSolveAll = 0      // stops on the first SAT instance
	p.FDropSatOuts = 0   // replace sat outputs by constant 0
	p.NLearnedStart = 10000
	p.NLearnedDelta = 2000
	p.NLearnedPerce = 80
	p.FVerbose = 0
	p.FNotVerbose = 0
	p.IFrame = -1
	p.NFailOuts = 0
	p.NDropOuts = 0
	p.timeLastSolved = 0
}

const (
	SAIG_TER_NON = 0
	SAIG_TER_ZER = 1
	SAIG_TER_ONE = 2
	SAIG_TER_UND = 3
)

func Saig_ManBmcSimInfoNot(Value int32) int32 {
	if Value == SAIG_TER_ZER {
		return SAIG_TER_ONE
	}
	if Value == SAIG_TER_ONE {
		return SAIG_TER_ZER
	}
	return SAIG_TER_UND
}

func Saig_ManBmcSimInfoAnd(Value0, Value1 int32) int32 {
	if Value0 == SAIG_TER_ZER || Value1 == SAIG_TER_ZER {
		return SAIG_TER_ZER
	}
	if Value0 == SAIG_TER_ONE && Value1 == SAIG_TER_ONE {
		return SAIG_TER_ONE
	}
	return SAIG_TER_UND
}

// Gia_ManBmc_t is the bmc3 manager. The C object accessors are replaced by
// per-object arrays read once from the AIG.
type Gia_ManBmc_t struct {
	// input/output data
	pPars  *Saig_ParBmc_t
	pAig   *aig.Graph
	vCexes *vec.Vec_Ptr_t // counter-examples
	// intermediate data
	vMapping []int32   // mapping (Cnf_DeriveMappingArray)
	vId2Num  []int32   // number of each node
	vTerInfo [][]uint8 // ternary information (one value per object)
	vId2Var  [][]int32 // SAT vars for each object
	vVisited [][]int32 // visited nodes (Vec_Wec_t)
	// time spent on each outputs
	pTime4Outs []int64
	// hash table
	vData     *vec.Vec_Int_t    // storage for cuts
	vHash     *vec.Hsh_IntMan_t // hash table
	vId2Lit   []int32           // mapping cuts into literals
	nHashHit  int32             // hash table hits
	nHashMiss int32             // hash table misses
	nBufNum   int32             // the number of simple nodes
	nDupNum   int32             // the number of simple nodes
	nUniProps int32             // propagating learned clause values
	// SAT solver
	pSat     *bsat.Solver
	pSat2    *satoko.Solver
	pSat3    *glucose.Solver
	nSatVars int32 // SAT variables
	nObjNums int32 // SAT objects
	pSops    [][]byte
	// the AIG
	kind      []aig.NodeKind
	fanin0    []aig.Edge
	fanin1    []aig.Edge
	ciNum     []int32 // CI index of each CI object, else -1
	cis, cos  []aig.Edge
	travIds   []int32
	travId    int32
	nPis      int32
	nPos      int32
	nRegs     int32
	nodeOrder []int32 // AND/EXOR object IDs in increasing order
}

func (p *Gia_ManBmc_t) isNode(id int32) bool {
	k := p.kind[id]
	return k == aig.And || k == aig.ExclusiveOr
}
func (p *Gia_ManBmc_t) isCi(id int32) bool { return p.kind[id] == aig.Input }
func (p *Gia_ManBmc_t) isCo(id int32) bool { return p.kind[id] == aig.Output }
func (p *Gia_ManBmc_t) isPi(id int32) bool { return p.kind[id] == aig.Input && p.ciNum[id] < p.nPis }
func (p *Gia_ManBmc_t) isLo(id int32) bool { return p.kind[id] == aig.Input && p.ciNum[id] >= p.nPis }

// loToLi is Saig_ObjLoToLi.
func (p *Gia_ManBmc_t) loToLi(id int32) int32 {
	return int32(p.cos[p.nPos+p.ciNum[id]-p.nPis].NodeID())
}

// bmcNormalizeAIG lowers Go's native XOR nodes into the AND/inverter graph
// accepted by Cnf_DeriveMappingArray; C's bmc3 inputs never contain them.
func bmcNormalizeAIG(src *aig.Graph) *aig.Graph {
	nodes, inputs, outputs := src.Nodes(), src.Inputs(), src.Outputs()
	needsCopy := false
	for _, n := range nodes {
		if n.Kind == aig.ExclusiveOr {
			needsCopy = true
			break
		}
	}
	if !needsCopy {
		return src
	}
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	mapped := make([]aig.Edge, len(nodes))
	mapped[0] = dst.Const1()
	for _, input := range inputs {
		mapped[input.NodeID()] = dst.NewInput()
	}
	child := func(e aig.Edge) aig.Edge {
		v := mapped[e.NodeID()]
		if e.Complemented() {
			v = v.Not()
		}
		return v
	}
	for id, n := range nodes {
		switch n.Kind {
		case aig.And:
			mapped[id] = dst.AND(child(n.Fanins[0]), child(n.Fanins[1]))
		case aig.ExclusiveOr:
			a, b := child(n.Fanins[0]), child(n.Fanins[1])
			mapped[id] = dst.OR(dst.AND(a, b.Not()), dst.AND(a.Not(), b))
		case aig.Buffer, aig.Output:
			mapped[id] = child(n.Fanins[0])
		}
	}
	for _, output := range outputs {
		dst.AddOutput(child(nodes[output.NodeID()].Fanins[0]))
	}
	aig.Aig_ManSetRegNum(dst, aig.Aig_ManRegNum(src))
	return dst
}

// Saig_Bmc3ManStart creates the manager.
func Saig_Bmc3ManStart(pAig *aig.Graph, nTimeOutOne, nConfLimit, fUseSatoko, fUseGlucose int32) *Gia_ManBmc_t {
	p := &Gia_ManBmc_t{pAig: pAig}
	nodes := pAig.Nodes()
	nObjs := len(nodes)
	p.cis, p.cos = pAig.Inputs(), pAig.Outputs()
	p.nRegs = aig.Aig_ManRegNum(pAig)
	p.nPis = int32(len(p.cis)) - p.nRegs
	p.nPos = int32(len(p.cos)) - p.nRegs
	p.kind = make([]aig.NodeKind, nObjs)
	p.fanin0 = make([]aig.Edge, nObjs)
	p.fanin1 = make([]aig.Edge, nObjs)
	p.ciNum = make([]int32, nObjs)
	p.travIds = make([]int32, nObjs)
	for id, n := range nodes {
		p.ciNum[id] = -1
		if aig.Aig_ObjType(pAig, aig.Edge(id)<<1) == aig.AIG_OBJ_NONE {
			p.kind[id] = aig.Buffer // deleted: neither CI, CO nor node
			continue
		}
		p.kind[id] = n.Kind
		p.fanin0[id], p.fanin1[id] = n.Fanins[0], n.Fanins[1]
		if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
			p.nodeOrder = append(p.nodeOrder, int32(id))
		}
	}
	for i, e := range p.cis {
		p.ciNum[e.NodeID()] = int32(i)
	}
	// create mapping
	p.vMapping = vec.Vec_IntArray(cnf.Cnf_DeriveMappingArray(pAig))
	// map object IDs into their numbers and section numbers
	p.nObjNums = 0
	p.vId2Num = make([]int32, nObjs)
	for i := range p.vId2Num {
		p.vId2Num[i] = -1
	}
	p.vId2Num[0] = p.nObjNums
	p.nObjNums++
	for _, e := range p.cis {
		p.vId2Num[e.NodeID()] = p.nObjNums
		p.nObjNums++
	}
	for _, id := range p.nodeOrder {
		if p.vMapping[id] > 0 {
			p.vId2Num[id] = p.nObjNums
			p.nObjNums++
		}
	}
	for _, e := range p.cos {
		p.vId2Num[e.NodeID()] = p.nObjNums
		p.nObjNums++
	}
	// create solver
	p.nSatVars = 1
	if fUseSatoko != 0 {
		opts := satoko.DefaultOptions()
		opts.ConflictLimit = int64(nConfLimit)
		p.pSat2 = satoko.New()
		*p.pSat2.Options() = opts
		p.pSat2.SetVariables(1000)
	} else if fUseGlucose != 0 {
		p.pSat3 = glucose.Bmcg_sat_solver_start()
		for i := 0; i < 1000; i++ {
			glucose.Bmcg_sat_solver_addvar(p.pSat3)
		}
	} else {
		p.pSat = bsat.NewSolver()
		p.pSat.SetnvarsC(1000)
	}
	p.pSops = cnf.Cnf_ReadMsops()
	// hash table
	p.vData = vec.Vec_IntAlloc(5 * 10000)
	p.vHash = vec.Hsh_IntManStart(p.vData, 5, 10000)
	p.vId2Lit = make([]int32, 0, 10000)
	// time spent on each outputs
	if nTimeOutOne != 0 {
		p.pTime4Outs = make([]int64, p.nPos)
		for i := range p.pTime4Outs {
			p.pTime4Outs[i] = int64(nTimeOutOne)*util.CLOCKS_PER_SEC/1000 + 1
		}
	}
	return p
}

func (p *Gia_ManBmc_t) satNVars() int32 {
	if p.pSat != nil {
		return p.pSat.NVarsC()
	}
	if p.pSat3 != nil {
		return glucose.Bmcg_sat_solver_varnum(p.pSat3)
	}
	return int32(p.pSat2.VariableCount())
}

// Saig_Bmc3ManStop deletes the manager.
func Saig_Bmc3ManStop(p *Gia_ManBmc_t) {
	if p.pPars.FVerbose != 0 {
		var nUsedVars, start, delta, ratio, reduces int32
		if p.pSat != nil {
			nUsedVars = p.pSat.CountUsedVarsC()
			start, delta, ratio, reduces = p.pSat.LearntLimitsC()
		}
		fmt.Printf("LStart(P) = %d  LDelta(Q) = %d  LRatio(R) = %d  ReduceDB = %d  Vars = %d  Used = %d (%.2f %%)\n",
			start, delta, ratio, reduces, p.satNVars(), nUsedVars, 100.0*float64(nUsedVars)/float64(p.satNVars()))
		fmt.Printf("Buffs = %d. Dups = %d.   Hash hits = %d.  Hash misses = %d.  UniProps = %d.\n",
			p.nBufNum, p.nDupNum, p.nHashHit, p.nHashMiss, p.nUniProps)
	}
	if p.vCexes != nil {
		p.pAig.SetSequentialCounterexamples(p.vCexes)
		p.vCexes = nil
	}
	if p.pSat != nil {
		p.pSat.Delete()
	}
	vec.Hsh_IntManStop(p.vHash)
}

// Saig_ManBmcMapping returns the offset of the object's cut record, or 0.
func (p *Gia_ManBmc_t) Saig_ManBmcMapping(id int32) []int32 {
	if p.vMapping[id] == 0 {
		return nil
	}
	return p.vMapping[p.vMapping[id] : p.vMapping[id]+5]
}

func (p *Gia_ManBmc_t) Saig_ManBmcLiteral(id int32, iFrame int32) int32 {
	ObjNum := p.vId2Num[id]
	if ObjNum < 0 {
		panic("bmc3: object without a number")
	}
	return p.vId2Var[iFrame][ObjNum]
}

func (p *Gia_ManBmc_t) Saig_ManBmcSetLiteral(id int32, iFrame int32, iLit int32) int32 {
	p.vId2Var[iFrame][p.vId2Num[id]] = iLit
	return iLit
}

var saigBmcTruth = [4]int32{0xAAAA, 0xCCCC, 0xF0F0, 0xFF00}

func Saig_ManBmcCof0(t, v int32) int32 {
	return 0xffff & ((t &^ saigBmcTruth[v]) | ((t &^ saigBmcTruth[v]) << (1 << v)))
}
func Saig_ManBmcCof1(t, v int32) int32 {
	return 0xffff & ((t & saigBmcTruth[v]) | ((t & saigBmcTruth[v]) >> (1 << v)))
}
func Saig_ManBmcCofEqual(t, v int32) bool {
	switch v {
	case 0:
		return ((t & 0xAAAA) >> 1) == (t & 0x5555)
	case 1:
		return ((t & 0xCCCC) >> 2) == (t & 0x3333)
	case 2:
		return ((t & 0xF0F0) >> 4) == (t & 0x0F0F)
	case 3:
		return ((t & 0xFF00) >> 8) == (t & 0x00FF)
	}
	panic("bmc3: bad cofactor variable")
}

// Saig_ManBmcReduceTruth propagates constants into the cut function.
func Saig_ManBmcReduceTruth(uTruth int32, Lits []int32) int32 {
	for v := int32(0); v < 4; v++ {
		if Lits[v] == 0 {
			uTruth = Saig_ManBmcCof0(uTruth, v)
			Lits[v] = -1
		} else if Lits[v] == 1 {
			uTruth = Saig_ManBmcCof1(uTruth, v)
			Lits[v] = -1
		}
	}
	for v := int32(0); v < 4; v++ {
		if Lits[v] == -1 {
			if !Saig_ManBmcCofEqual(uTruth, v) {
				panic("bmc3: truth depends on an absent input")
			}
		} else if Saig_ManBmcCofEqual(uTruth, v) {
			Lits[v] = -1
		}
	}
	return uTruth
}

func dar_CutTruthSwapAdjacentVars(uTruth uint32, iVar int) uint32 {
	switch iVar {
	case 0:
		return (uTruth & 0x99999999) | ((uTruth & 0x22222222) << 1) | ((uTruth & 0x44444444) >> 1)
	case 1:
		return (uTruth & 0xC3C3C3C3) | ((uTruth & 0x0C0C0C0C) << 2) | ((uTruth & 0x30303030) >> 2)
	case 2:
		return (uTruth & 0xF00FF00F) | ((uTruth & 0x00F000F0) << 4) | ((uTruth & 0x0F000F00) >> 4)
	}
	panic("bmc3: bad adjacent variable")
}

func dar_CutTruthSwapPolarity(uTruth uint32, iVar int) uint32 {
	switch iVar {
	case 0:
		return ((uTruth & 0xAAAA) >> 1) | ((uTruth & 0x5555) << 1)
	case 1:
		return ((uTruth & 0xCCCC) >> 2) | ((uTruth & 0x3333) << 2)
	case 2:
		return ((uTruth & 0xF0F0) >> 4) | ((uTruth & 0x0F0F) << 4)
	case 3:
		return ((uTruth & 0xFF00) >> 8) | ((uTruth & 0x00FF) << 8)
	}
	panic("bmc3: bad polarity variable")
}

// Dar_CutSortVars makes the cut inputs positive and sorts them.
func Dar_CutSortVars(uTruth uint32, pVars []int32) uint32 {
	// replace -1 by large number
	for i := 0; i < 4; i++ {
		if pVars[i] == -1 {
			pVars[i] = 0x3FFFFFFF
		} else if pVars[i]&1 != 0 {
			pVars[i] ^= 1
			uTruth = dar_CutTruthSwapPolarity(uTruth, i)
		}
	}
	// permute variables
	for fChange := true; fChange; {
		fChange = false
		for i := 0; i < 3; i++ {
			if pVars[i] <= pVars[i+1] {
				continue
			}
			fChange = true
			pVars[i], pVars[i+1] = pVars[i+1], pVars[i]
			uTruth = dar_CutTruthSwapAdjacentVars(uTruth, i)
		}
	}
	// replace large number by -1
	for i := 0; i < 4; i++ {
		if pVars[i] == 0x3FFFFFFF {
			pVars[i] = -1
		}
	}
	return uTruth
}

// Saig_ManBmcAddClauses derives CNF for one node.
func (p *Gia_ManBmc_t) Saig_ManBmcAddClauses(uTruth int32, Lits []int32, iLitOut int32) {
	var ClaLits [5]int32
	if !(uTruth > 0 && uTruth < 0xffff) {
		panic("bmc3: constant cut function")
	}
	// write positive/negative polarity
	for i := 0; i < 2; i++ {
		if i != 0 {
			uTruth = 0xffff & ^uTruth
		}
		for _, cube := range p.pSops[uTruth] {
			nClaLits := 0
			if i != 0 {
				ClaLits[nClaLits] = iLitOut ^ 1
			} else {
				ClaLits[nClaLits] = iLitOut
			}
			nClaLits++
			CutLit := int32(cube)
			for b := 3; b >= 0; b-- {
				if CutLit%3 == 0 { // value 0 --> write positive literal
					ClaLits[nClaLits] = Lits[b]
					nClaLits++
				} else if CutLit%3 == 1 { // value 1 --> write negative literal
					ClaLits[nClaLits] = Lits[b] ^ 1
					nClaLits++
				}
				CutLit = CutLit / 3
			}
			var ok bool
			if p.pSat2 != nil {
				ok = p.pSat2.AddClause32(append([]int32(nil), ClaLits[:nClaLits]...))
			} else if p.pSat3 != nil {
				ok = glucose.Bmcg_sat_solver_addclause(p.pSat3, append([]int32(nil), ClaLits[:nClaLits]...), int32(nClaLits)) != 0
			} else {
				ok = p.pSat.AddClauseC(append([]bsat.Literal(nil), ClaLits[:nClaLits]...)) != 0
			}
			if !ok {
				panic("bmc3: adding a cut clause failed")
			}
		}
	}
}

// Saig_ManBmcCreateCnf_rec derives CNF for one node.
func (p *Gia_ManBmc_t) Saig_ManBmcCreateCnf_rec(id int32, iFrame int32) int32 {
	var Lits [5]int32
	iLit := p.Saig_ManBmcLiteral(id, iFrame)
	if iLit != ^int32(0) {
		return iLit
	}
	if iFrame < 0 {
		panic("bmc3: negative frame")
	}
	if p.isCi(id) {
		if p.isPi(id) {
			iLit = p.nSatVars << 1
			p.nSatVars++
		} else {
			iLit = p.Saig_ManBmcCreateCnf_rec(p.loToLi(id), iFrame-1)
		}
		return p.Saig_ManBmcSetLiteral(id, iFrame, iLit)
	}
	if p.isCo(id) {
		iLit = p.Saig_ManBmcCreateCnf_rec(int32(p.fanin0[id].NodeID()), iFrame)
		if p.fanin0[id].Complemented() {
			iLit ^= 1
		}
		return p.Saig_ManBmcSetLiteral(id, iFrame, iLit)
	}
	if !p.isNode(id) {
		panic("bmc3: unexpected object")
	}
	pMapping := p.Saig_ManBmcMapping(id)
	for i := 0; i < 4; i++ {
		if pMapping[i+1] == -1 {
			Lits[i] = -1
		} else {
			Lits[i] = p.Saig_ManBmcCreateCnf_rec(pMapping[i+1], iFrame)
		}
	}
	uTruth := 0xffff & pMapping[0]
	// propagate constants
	uTruth = Saig_ManBmcReduceTruth(uTruth, Lits[:4])
	if uTruth == 0 || uTruth == 0xffff {
		iLit = 0
		if uTruth == 0xffff {
			iLit = 1
		}
		return p.Saig_ManBmcSetLiteral(id, iFrame, iLit)
	}
	// canonicize inputs
	uTruth = int32(Dar_CutSortVars(uint32(uTruth), Lits[:4]))
	if uTruth == 0 || uTruth == 0xffff {
		panic("bmc3: constant after sorting")
	}
	if uTruth == 0xAAAA || uTruth == 0x5555 {
		iLit = Lits[0]
		if uTruth == 0x5555 {
			iLit ^= 1
		}
		p.nBufNum++
	} else {
		fCompl := uTruth & 1
		if fCompl != 0 {
			Lits[4] = 0xffff & ^uTruth
		} else {
			Lits[4] = uTruth
		}
		iEntry := vec.Vec_IntSize(p.vData) / 5
		for i := 0; i < 5; i++ {
			vec.Vec_IntPush(p.vData, Lits[i])
		}
		iRes := int(vec.Hsh_IntManAdd(p.vHash, iEntry))
		if iRes == iEntry {
			iLit = p.nSatVars << 1
			p.nSatVars++
			p.Saig_ManBmcAddClauses(Lits[4], Lits[:4], iLit)
			if iEntry != len(p.vId2Lit) {
				panic("bmc3: hash entry out of sync")
			}
			p.vId2Lit = append(p.vId2Lit, iLit)
			p.nHashMiss++
		} else {
			iLit = p.vId2Lit[iRes]
			vec.Vec_IntShrink(p.vData, vec.Vec_IntSize(p.vData)-5)
			p.nHashHit++
		}
		iLit ^= fCompl
	}
	return p.Saig_ManBmcSetLiteral(id, iFrame, iLit)
}

// Saig_ManBmcCreateCnf_iter collects the objects of one frame, and the
// latch inputs of the previous frame, needed for the CNF of pObj.
func (p *Gia_ManBmc_t) Saig_ManBmcCreateCnf_iter(id int32, iFrame int32, vVisit *[]int32) {
	if p.Saig_ManBmcLiteral(id, iFrame) != ^int32(0) {
		return
	}
	if p.travIds[id] == p.travId {
		return
	}
	p.travIds[id] = p.travId
	if p.isCi(id) {
		if p.isLo(id) {
			*vVisit = append(*vVisit, p.loToLi(id))
		}
		return
	}
	if p.isCo(id) {
		p.Saig_ManBmcCreateCnf_iter(int32(p.fanin0[id].NodeID()), iFrame, vVisit)
		return
	}
	pMapping := p.Saig_ManBmcMapping(id)
	for i := 0; i < 4; i++ {
		if pMapping[i+1] != -1 {
			p.Saig_ManBmcCreateCnf_iter(pMapping[i+1], iFrame, vVisit)
		}
	}
}

// Saig_ManBmcRunTerSim_rec recursively performs ternary simulation.
func (p *Gia_ManBmc_t) Saig_ManBmcRunTerSim_rec(id int32, iFrame int32) int32 {
	pInfo := p.vTerInfo[iFrame]
	Value := int32(pInfo[id])
	if Value != SAIG_TER_NON {
		return Value
	}
	if p.isCo(id) {
		Value = p.Saig_ManBmcRunTerSim_rec(int32(p.fanin0[id].NodeID()), iFrame)
		if p.fanin0[id].Complemented() {
			Value = Saig_ManBmcSimInfoNot(Value)
		}
	} else if p.isLo(id) {
		if iFrame <= 0 {
			panic("bmc3: latch output at frame 0 without a value")
		}
		Value = p.Saig_ManBmcRunTerSim_rec(p.loToLi(id), iFrame-1)
	} else if p.isNode(id) {
		Val0 := p.Saig_ManBmcRunTerSim_rec(int32(p.fanin0[id].NodeID()), iFrame)
		Val1 := p.Saig_ManBmcRunTerSim_rec(int32(p.fanin1[id].NodeID()), iFrame)
		if p.fanin0[id].Complemented() {
			Val0 = Saig_ManBmcSimInfoNot(Val0)
		}
		if p.fanin1[id].Complemented() {
			Val1 = Saig_ManBmcSimInfoNot(Val1)
		}
		Value = Saig_ManBmcSimInfoAnd(Val0, Val1)
	} else {
		panic("bmc3: unexpected object in ternary simulation")
	}
	pInfo[id] = uint8(Value)
	// transfer to the unrolling
	if p.isNode(id) && p.vMapping[id] != 0 && Value != SAIG_TER_UND {
		v := int32(0)
		if Value == SAIG_TER_ONE {
			v = 1
		}
		p.Saig_ManBmcSetLiteral(id, iFrame, v)
	}
	return Value
}

// Saig_ManBmcCreateCnf derives CNF for one output in one frame.
func (p *Gia_ManBmc_t) Saig_ManBmcCreateCnf(id int32, iFrame int32) int32 {
	// perform ternary simulation
	Value := p.Saig_ManBmcRunTerSim_rec(id, iFrame)
	if Value != SAIG_TER_UND {
		if Value == SAIG_TER_ONE {
			return 1
		}
		return 0
	}
	// construct CNF if value is ternary
	p.vVisited = p.vVisited[:0]
	p.vVisited = append(p.vVisited, []int32{id})
	for f := iFrame; f >= 0; f-- {
		p.travId++
		p.vVisited = append(p.vVisited, nil)
		vVisit := p.vVisited[len(p.vVisited)-2]
		var vVisit2 []int32
		for _, obj := range vVisit {
			p.Saig_ManBmcCreateCnf_iter(obj, f, &vVisit2)
		}
		p.vVisited[len(p.vVisited)-1] = vVisit2
		if len(vVisit2) == 0 {
			break
		}
	}
	for f := len(p.vVisited) - 1; f >= 0; f-- {
		for _, obj := range p.vVisited[f] {
			p.Saig_ManBmcCreateCnf_rec(obj, iFrame-int32(f))
		}
	}
	Lit := p.Saig_ManBmcLiteral(id, iFrame)
	// extend the SAT solver
	if p.pSat2 != nil {
		if int(p.nSatVars) > p.pSat2.VariableCount() {
			p.pSat2.SetVariables(int(p.nSatVars))
		}
	} else if p.pSat3 != nil {
		for i := glucose.Bmcg_sat_solver_varnum(p.pSat3); i < p.nSatVars; i++ {
			glucose.Bmcg_sat_solver_addvar(p.pSat3)
		}
	} else {
		p.pSat.SetnvarsC(p.nSatVars)
	}
	return Lit
}

// Saig_ManBmcTimeToStop returns time to stop.
func Saig_ManBmcTimeToStop(pPars *Saig_ParBmc_t, nTimeToStopNG int64) int64 {
	var nTimeToStopGap int64
	if pPars.NTimeOutGap != 0 {
		nTimeToStopGap = int64(pPars.NTimeOutGap)*util.CLOCKS_PER_SEC + util.Abc_Clock()
	}
	var nTimeToStop int64
	if nTimeToStopNG != 0 && nTimeToStopGap != 0 {
		nTimeToStop = min(nTimeToStopNG, nTimeToStopGap)
	} else if nTimeToStopNG != 0 {
		nTimeToStop = nTimeToStopNG
	} else if nTimeToStopGap != 0 {
		nTimeToStop = nTimeToStopGap
	}
	return nTimeToStop
}

// Saig_ManGenerateCex derives the counter-example of output i in frame f.
func (p *Gia_ManBmc_t) Saig_ManGenerateCex(f, i int32, fModel bool) *util.Abc_Cex_t {
	pCex := util.Abc_CexMakeTriv(p.nRegs, p.nPis, p.nPos, f*p.nPos+i)
	// without a SAT call (the output is constant 1) there is no model to
	// read (CBUGS_FIXED.md #37)
	if !fModel {
		return pCex
	}
	iBit := p.nRegs
	for j := int32(0); j <= f; j, iBit = j+1, iBit+p.nPis {
		for k := int32(0); k < p.nPis; k++ {
			iLit := p.Saig_ManBmcLiteral(int32(p.cis[k].NodeID()), j)
			if iLit == ^int32(0) {
				continue
			}
			if p.varValue(iLit>>1) != 0 {
				util.Abc_InfoSetBit(pCex.PData, iBit+k)
			}
		}
	}
	return pCex
}

func (p *Gia_ManBmc_t) varValue(v int32) int32 {
	if p.pSat2 != nil {
		return p.pSat2.VarPolarityCexValue(int(v))
	}
	if p.pSat3 != nil {
		return glucose.Bmcg_sat_solver_read_cex_varvalue(p.pSat3, v)
	}
	return p.pSat.VarValueC(v)
}

func (p *Gia_ManBmc_t) setRuntimeLimit(limit int64) {
	if p.pSat2 != nil {
		if limit == 0 {
			p.pSat2.SetDeadline(time.Time{})
		} else {
			p.pSat2.SetDeadline(abcClockToTime(limit))
		}
	} else if p.pSat3 != nil {
		glucose.Bmcg_sat_solver_set_runtime_limit(p.pSat3, limit)
	} else {
		p.pSat.SetRuntimeDeadline(util.Abc_Clock, limit)
	}
}

func (p *Gia_ManBmc_t) satStats() (clauses, conflicts, learnts float64) {
	if p.pSat != nil {
		return p.pSat.StatsC()
	}
	if p.pSat3 != nil {
		return float64(glucose.Bmcg_sat_solver_clausenum(p.pSat3)), float64(glucose.Bmcg_sat_solver_conflictnum(p.pSat3)), float64(glucose.Bmcg_sat_solver_learntnum(p.pSat3))
	}
	return float64(p.pSat2.ClauseCount()), float64(p.pSat2.Stats().AllConflicts), float64(p.pSat2.LearntCount())
}

func (p *Gia_ManBmc_t) satMemory() float64 {
	if p.pSat != nil {
		return p.pSat.MemoryC()
	}
	return 0
}

// Saig_ManCallSolver solves one output literal.
func (p *Gia_ManBmc_t) Saig_ManCallSolver(Lit int32) int32 {
	if Lit == 0 {
		return l_False
	}
	if Lit == 1 {
		return l_True
	}
	if p.pSat2 != nil {
		switch p.pSat2.SolveAssumptionsLimit([]int{int(Lit)}, int64(p.pPars.NConfLimit)) {
		case satoko.SATOKO_SAT:
			return l_True
		case satoko.SATOKO_UNSAT:
			return l_False
		}
		return l_Undef
	} else if p.pSat3 != nil {
		glucose.Bmcg_sat_solver_set_conflict_budget(p.pSat3, p.pPars.NConfLimit)
		return glucose.Bmcg_sat_solver_solve(p.pSat3, []int32{Lit}, 1)
	}
	return p.pSat.SolveC([]bsat.Literal{Lit}, int64(p.pPars.NConfLimit), 0, 0, 0)
}

const (
	l_Undef int32 = 0
	l_True  int32 = 1
	l_False int32 = -1
)

func bmc3Sign(fUnfinished int32) string {
	if fUnfinished != 0 {
		return "-"
	}
	return "+"
}

// abcClockToTime converts an Abc_Clock deadline for satoko.
func abcClockToTime(limit int64) time.Time {
	return time.Now().Add(time.Duration(limit - util.Abc_Clock()))
}

// Saig_ManBmcScalable is the bounded model checking engine "bmc3". It
// returns 0 on a counter-example, 1 when all reachable states are explored,
// and -1 when undecided.
func Saig_ManBmcScalable(pAig0 *aig.Graph, pPars *Saig_ParBmc_t) int32 {
	var pLogFile *os.File
	RetValue, fFirst, nJumpFrame, fUnfinished := int32(-1), int32(1), int32(0), int32(0)
	var f int32
	var clkOther, nTimeUnsat, nTimeSat, nTimeUndec, clkOne int64
	clkTotal := util.Abc_Clock()
	if pAig0 == nil || pPars == nil {
		return -1
	}
	pAig := bmcNormalizeAIG(pAig0)
	nPos := int32(pAig.OutputCount()) - aig.Aig_ManRegNum(pAig)
	nOutDigits := util.Abc_Base10Log(uint32(nPos))
	if pPars.PLogFileName != "" {
		pLogFile, _ = os.Create(pPars.PLogFileName)
	}
	if pPars.NTimeOutOne != 0 && pPars.NTimeOut == 0 {
		pPars.NTimeOut = pPars.NTimeOutOne*nPos/1000 + 1
	}
	if pPars.NTimeOutOne != 0 && pPars.FSolveAll == 0 {
		pPars.NTimeOutOne = 0
	}
	var nTimeToStopNG int64
	if pPars.NTimeOut != 0 {
		nTimeToStopNG = int64(pPars.NTimeOut)*util.CLOCKS_PER_SEC + util.Abc_Clock()
	}
	nTimeToStop := Saig_ManBmcTimeToStop(pPars, nTimeToStopNG)
	// create BMC manager
	p := Saig_Bmc3ManStart(pAig, pPars.NTimeOutOne, pPars.NConfLimit, pPars.FUseSatoko, pPars.FUseGlucose)
	p.pPars = pPars
	if p.pSat != nil {
		p.pSat.SetLearnedPolicy(pPars.NLearnedStart, pPars.NLearnedDelta, pPars.NLearnedPerce)
		p.pSat.SetNoRestarts(pPars.FNoRestarts != 0)
		p.pSat.SetStopFunc(pPars.RunId, pPars.pFuncStop)
	} else if p.pSat3 != nil {
	} else {
		p.pSat2.SetRunID(int(pPars.RunId))
		if pPars.pFuncStop != nil {
			p.pSat2.SetStopFunc(func(id int) int { return int(pPars.pFuncStop(int32(id))) })
		}
	}
	if pPars.FSolveAll != 0 && p.vCexes == nil {
		p.vCexes = vec.Vec_PtrStart(int(nPos))
	}
	if pPars.FVerbose != 0 {
		fmt.Printf("Running \"bmc3\". PI/PO/Reg = %d/%d/%d. And =%7d. Lev =%6d. ObjNums =%6d.\n",
			p.nPis, p.nPos, p.nRegs, aig.Aig_ManNodeNum(pAig), aig.Aig_ManLevelNum(pAig), p.nObjNums)
		fmt.Printf("Params: FramesMax = %d. Start = %d. ConfLimit = %d. TimeOut = %d. SolveAll = %d.\n",
			pPars.NFramesMax, pPars.NStart, pPars.NConfLimit, pPars.NTimeOut, pPars.FSolveAll)
	}
	if pPars.NFramesMax == 0 {
		pPars.NFramesMax = 1000000000 // ABC_INFINITY
	}
	// set runtime limit
	if nTimeToStop != 0 {
		p.setRuntimeLimit(nTimeToStop)
	}
	bridge := bmcBridgeWriter(pPars.pBridgeWriter)
	cexEntry := func(i int32) bool {
		return p.vCexes != nil && vec.Vec_PtrEntry(p.vCexes, int(i)) != nil
	}
	// perform frames
	aig.Aig_ManRandom(1)
	pPars.timeLastSolved = util.Abc_Clock()
	func() { // returning from this closure is C's "goto finish"
		for f = 0; f < pPars.NFramesMax; f++ {
			// stop BMC after exploring all reachable states
			if pPars.NFramesJump == 0 && p.nRegs < 30 && f == int32(1)<<uint(p.nRegs) {
				fmt.Printf("Stopping BMC because all 2^%d reachable states are visited.\n", p.nRegs)
				if pPars.FUseBridge != 0 {
					for i := int32(0); i < p.nPos; i++ {
						if !cexEntry(i) && !(p.pTime4Outs != nil && p.pTime4Outs[i] == 0) { // not SAT and not timed out
							_ = writeBmcBridgeResult(bridge, 1, nil, i)
						}
					}
				}
				RetValue = 1
				if pPars.NFailOuts != 0 {
					RetValue = 0
				}
				return
			}
			// stop BMC if all targets are solved
			if pPars.FSolveAll != 0 && pPars.NFailOuts+pPars.NDropOuts >= p.nPos {
				fmt.Printf("Stopping BMC because all targets are disproved or timed out.\n")
				RetValue = 1
				if pPars.NFailOuts != 0 {
					RetValue = 0
				}
				return
			}
			// consider the next timeframe
			if (RetValue == -1 || pPars.FSolveAll != 0) && pPars.NStart == 0 && nJumpFrame == 0 {
				pPars.IFrame = f - 1
			}
			// map nodes of this section
			vFrame := make([]int32, p.nObjNums)
			for i := range vFrame {
				vFrame[i] = ^int32(0)
			}
			p.vId2Var = append(p.vId2Var, vFrame)
			pInfo := make([]uint8, len(p.kind))
			p.vTerInfo = append(p.vTerInfo, pInfo)
			// prepare some nodes
			p.Saig_ManBmcSetLiteral(0, f, 1)
			pInfo[0] = SAIG_TER_ONE
			for i := int32(0); i < p.nPis; i++ {
				pInfo[p.cis[i].NodeID()] = SAIG_TER_UND
			}
			if f == 0 {
				for i := int32(0); i < p.nRegs; i++ {
					id := int32(p.cis[p.nPis+i].NodeID())
					p.Saig_ManBmcSetLiteral(id, 0, 0)
					pInfo[id] = SAIG_TER_ZER
				}
			}
			if (pPars.NStart != 0 && f < pPars.NStart) || (nJumpFrame != 0 && f < nJumpFrame) {
				continue
			}
			// create CNF upfront
			if pPars.FSolveAll != 0 {
				for i := int32(0); i < p.nPos; i++ {
					// check for timeout
					if pPars.NTimeOutGap != 0 && pPars.timeLastSolved != 0 && util.Abc_Clock() > pPars.timeLastSolved+int64(pPars.NTimeOutGap)*util.CLOCKS_PER_SEC {
						fmt.Printf("Reached gap timeout (%d seconds).\n", pPars.NTimeOutGap)
						return
					}
					if nTimeToStop != 0 && util.Abc_Clock() > nTimeToStop {
						if pPars.FSilent == 0 {
							fmt.Printf("Reached timeout (%d seconds).\n", pPars.NTimeOut)
						}
						return
					}
					// skip solved outputs
					if cexEntry(i) {
						continue
					}
					// skip output whose time has run out
					if p.pTime4Outs != nil && p.pTime4Outs[i] == 0 {
						continue
					}
					// add constraints for this output
					clk2 := util.Abc_Clock()
					p.Saig_ManBmcCreateCnf(int32(p.cos[i].NodeID()), f)
					clkOther += util.Abc_Clock() - clk2
				}
			}
			// solve SAT
			for i := int32(0); i < p.nPos; i++ {
				// check for timeout
				if pPars.NTimeOutGap != 0 && pPars.timeLastSolved != 0 && util.Abc_Clock() > pPars.timeLastSolved+int64(pPars.NTimeOutGap)*util.CLOCKS_PER_SEC {
					fmt.Printf("Reached gap timeout (%d seconds).\n", pPars.NTimeOutGap)
					return
				}
				if nTimeToStop != 0 && util.Abc_Clock() > nTimeToStop {
					if pPars.FSilent == 0 {
						fmt.Printf("Reached timeout (%d seconds).\n", pPars.NTimeOut)
					}
					return
				}
				if pPars.pFuncStop != nil && pPars.pFuncStop(pPars.RunId) != 0 {
					if pPars.FSilent == 0 {
						fmt.Printf("Bmc3 got callbacks.\n")
					}
					return
				}
				// skip solved outputs
				if cexEntry(i) {
					continue
				}
				// skip output whose time has run out
				if p.pTime4Outs != nil && p.pTime4Outs[i] == 0 {
					continue
				}
				// add constraints for this output
				clk2 := util.Abc_Clock()
				Lit := p.Saig_ManBmcCreateCnf(int32(p.cos[i].NodeID()), f)
				clkOther += util.Abc_Clock() - clk2
				// solve this output
				fUnfinished = 0
				if p.pSat != nil {
					p.pSat.CompressC()
				}
				if p.pTime4Outs != nil {
					clkOne = util.Abc_Clock()
					p.setRuntimeLimit(p.pTime4Outs[i] + util.Abc_Clock())
				}
				clk2 = util.Abc_Clock()
				status := p.Saig_ManCallSolver(Lit)
				clkSatRun := util.Abc_Clock() - clk2
				if pLogFile != nil {
					var t1, t2 int32
					if Lit >= 2 {
						t1 = int32(clkSatRun * 1000 / util.CLOCKS_PER_SEC)
						// p->pTime4Outs exists only with -a -H (CBUGS_FIXED.md #38)
						if p.pTime4Outs != nil {
							t2 = max(0, min(pPars.NTimeOutOne, pPars.NTimeOutOne-int32((p.pTime4Outs[i]-clkSatRun)*1000/util.CLOCKS_PER_SEC)))
						}
					}
					fmt.Fprintf(pLogFile, "Frame %5d  Output %5d  Time(ms) %8d %8d\n", f, i, t1, t2)
				}
				if p.pTime4Outs != nil {
					timeSince := util.Abc_Clock() - clkOne
					if p.pTime4Outs[i] > timeSince {
						p.pTime4Outs[i] -= timeSince
					} else {
						p.pTime4Outs[i] = 0
					}
					if p.pTime4Outs[i] == 0 && status != l_True {
						pPars.NDropOuts++
					}
				}
				if status == l_False {
					nTimeUnsat += clkSatRun
					if Lit != 0 {
						// add final unit clause
						Lit ^= 1
						var ok bool
						if p.pSat2 != nil {
							ok = p.pSat2.AddClause32([]int32{Lit})
						} else if p.pSat3 != nil {
							ok = glucose.Bmcg_sat_solver_addclause(p.pSat3, []int32{Lit}, 1) != 0
						} else {
							ok = p.pSat.AddClauseC([]bsat.Literal{Lit}) != 0
						}
						if !ok {
							panic("bmc3: adding the final unit clause failed")
						}
						// add learned units
						if p.pSat != nil {
							for _, unit := range p.pSat.TakeUnitLiterals() {
								if p.pSat.AddClauseC([]bsat.Literal{unit}) == 0 {
									panic("bmc3: adding a learned unit failed")
								}
							}
							// propagate units
							p.pSat.CompressC()
						}
					}
					if pPars.FUseBridge != 0 {
						_ = writeBmcBridgeProgress(bridge, i, f)
					}
				} else if status == l_True {
					nTimeSat += clkSatRun
					RetValue = 0
					fFirst = 0
					fModel := Lit > 1 // the solver ran: a constant-1 output has no model
					if pPars.FSolveAll == 0 {
						if pPars.FVerbose != 0 {
							clauses, conflicts, learnts := p.satStats()
							fmt.Printf("%4d %s : ", f, bmc3Sign(fUnfinished))
							fmt.Printf("Var =%8.0f. ", float64(p.nSatVars))
							fmt.Printf("Cla =%9.0f. ", clauses)
							fmt.Printf("Conf =%7.0f. ", conflicts)
							fmt.Printf("Learn =%7.0f. ", learnts)
							fmt.Printf("%4.0f MB", 4.25*float64(f+1)*float64(p.nObjNums)/(1<<20))
							fmt.Printf("%4.0f MB", 1.0*p.satMemory()/(1<<20))
							fmt.Printf("%9.2f sec  ", float32(util.Abc_Clock()-clkTotal)/float32(util.CLOCKS_PER_SEC))
							fmt.Printf("\n")
						}
						_ = pAig.SetSequentialCounterexample(p.Saig_ManGenerateCex(f, i, fModel))
						return
					}
					pPars.NFailOuts++
					if pPars.FNotVerbose == 0 {
						fmt.Printf("Output %*d was asserted in frame %2d (solved %*d out of %*d outputs).\n",
							nOutDigits, i, f, nOutDigits, pPars.NFailOuts, nOutDigits, p.nPos)
					}
					if p.vCexes == nil {
						p.vCexes = vec.Vec_PtrStart(int(p.nPos))
					}
					pCexNew := util.AbcCexSentinel
					if pPars.FUseBridge != 0 || pPars.FStoreCex != 0 {
						pCexNew = p.Saig_ManGenerateCex(f, i, fModel)
					}
					var pCexNew0 *util.Abc_Cex_t
					if pPars.FUseBridge != 0 {
						_ = writeBmcBridgeResult(bridge, 0, pCexNew, pCexNew.IPo)
						pCexNew0 = pCexNew
						pCexNew = util.AbcCexSentinel
					}
					vec.Vec_PtrWriteEntry(p.vCexes, int(i), util.Abc_CexDup(pCexNew, p.nRegs))
					if pPars.pFuncOnFail != nil {
						var arg *util.Abc_Cex_t
						if pPars.FStoreCex != 0 {
							arg = vec.Vec_PtrEntry(p.vCexes, int(i)).(*util.Abc_Cex_t)
						}
						if pPars.pFuncOnFail(i, arg) != 0 {
							fmt.Printf("Quitting due to callback on fail.\n")
							return
						}
					}
					// reset the timeout
					pPars.timeLastSolved = util.Abc_Clock()
					nTimeToStop = Saig_ManBmcTimeToStop(pPars, nTimeToStopNG)
					if nTimeToStop != 0 {
						p.setRuntimeLimit(nTimeToStop)
					}
					// check if other outputs failed under the same counter-example
					// (only when the solver produced one)
					for k := int32(0); fModel && k < p.nPos; k++ {
						// skip solved outputs
						if cexEntry(k) {
							continue
						}
						// check if this output is solved
						Lit = p.Saig_ManBmcCreateCnf(int32(p.cos[k].NodeID()), f)
						if p.varValue(Lit>>1) == Lit&1 {
							continue
						}
						// write entry
						pPars.NFailOuts++
						if pPars.FNotVerbose == 0 {
							fmt.Printf("Output %*d was asserted in frame %2d (solved %*d out of %*d outputs).\n",
								nOutDigits, k, f, nOutDigits, pPars.NFailOuts, nOutDigits, p.nPos)
						}
						// report to the bridge
						if pPars.FUseBridge != 0 {
							// set the output number
							pCexNew0.IPo = k
							_ = writeBmcBridgeResult(bridge, 0, pCexNew0, pCexNew0.IPo)
						}
						// remember solved output
						pCexDup := util.Abc_CexDup(pCexNew, p.nRegs)
						// without stored CEXes pCexNew is the sentinel (Abc_Cex_t *)1,
						// which Abc_CexDup returns unchanged
						if pCexDup != util.AbcCexSentinel {
							pCexDup.IPo = k
						}
						vec.Vec_PtrWriteEntry(p.vCexes, int(k), pCexDup)
					}
				} else {
					nTimeUndec += clkSatRun
					if pPars.NFramesJump != 0 {
						pPars.NConfLimit = pPars.NConfLimitJump
						nJumpFrame = f + pPars.NFramesJump
						fUnfinished = 1
						break
					}
					if p.pTime4Outs == nil {
						return
					}
				}
			}
			if pPars.FVerbose != 0 {
				clauses, conflicts, learnts := p.satStats()
				if fFirst == 1 && f > 0 && conflicts > 1 {
					fFirst = 0
				}
				fmt.Printf("%4d %s : ", f, bmc3Sign(fUnfinished))
				fmt.Printf("Var =%8.0f. ", float64(p.nSatVars))
				fmt.Printf("Cla =%9.0f. ", clauses)
				fmt.Printf("Conf =%7.0f. ", conflicts)
				fmt.Printf("Learn =%7.0f. ", learnts)
				if pPars.FSolveAll != 0 {
					fmt.Printf("CEX =%5d. ", pPars.NFailOuts)
				}
				if pPars.NTimeOutOne != 0 {
					fmt.Printf("T/O =%4d. ", pPars.NDropOuts)
				}
				fmt.Printf("%4.0f MB", 4.0*float64(f+1)*float64(p.nObjNums)/(1<<20))
				fmt.Printf("%4.0f MB", 1.0*p.satMemory()/(1<<20))
				fmt.Printf("%9.2f sec ", 1.0*float64(util.Abc_Clock()-clkTotal)/float64(util.CLOCKS_PER_SEC))
				fmt.Printf("\n")
			}
		}
		// consider the next timeframe
		if nJumpFrame != 0 && pPars.NStart == 0 {
			pPars.IFrame = nJumpFrame - pPars.NFramesJump
		} else if RetValue == -1 && pPars.NStart == 0 {
			pPars.IFrame = f - 1
		}
	}()
	// finish:
	if pPars.FVerbose != 0 {
		total := float64(util.Abc_Clock() - clkTotal)
		sec := float64(util.CLOCKS_PER_SEC)
		fmt.Printf("Runtime:  ")
		fmt.Printf("CNF = %.1f sec (%.1f %%)  ", float64(clkOther)/sec, 100.0*float64(clkOther)/total)
		fmt.Printf("UNSAT = %.1f sec (%.1f %%)  ", float64(nTimeUnsat)/sec, 100.0*float64(nTimeUnsat)/total)
		fmt.Printf("SAT = %.1f sec (%.1f %%)  ", float64(nTimeSat)/sec, 100.0*float64(nTimeSat)/total)
		fmt.Printf("UNDEC = %.1f sec (%.1f %%)", float64(nTimeUndec)/sec, 100.0*float64(nTimeUndec)/total)
		fmt.Printf("\n")
	}
	Saig_Bmc3ManStop(p)
	if pAig != pAig0 {
		// the normalized copy holds the results for the caller's graph
		if cex := pAig.SequentialCounterexample(); cex != nil {
			_ = pAig0.SetSequentialCounterexample(cex)
		}
		if v := pAig.TakeSequentialCounterexamples(); v != nil {
			pAig0.SetSequentialCounterexamples(v)
		}
	}
	if pLogFile != nil {
		pLogFile.Close()
	}
	return RetValue
}
