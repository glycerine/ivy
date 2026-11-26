package saig

// Functional constraint detection by direct implications, ported from
// src/aig/saig/saigConstr2.c: Saig_ManUnrollCOI, Ssw_ManFindDirectImplications,
// Saig_ManCreateIndMiter, Saig_ManFilterUsingIndOne_new and
// Saig_ManFilterUsingInd.

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// saigFramesMap returns ABC's pObjMap: nFrames entries per object, unset
// entries holding aig.InvalidEdge (NULL).
func saigFramesMap(p *aig.Graph, nFrames int32) []aig.Edge {
	m := make([]aig.Edge, int(nFrames)*int(aig.Aig_ManObjNumMax(p)))
	for i := range m {
		m[i] = aig.InvalidEdge
	}
	return m
}

// saigUnrollFrames builds the frames shared by Saig_ManUnrollCOI and
// Saig_ManCreateIndMiter: the PIs of every frame, then the frame-zero latch
// outputs as free inputs, then each frame's nodes in order, the latch inputs
// of frame f feeding the latch outputs of frame f+1.
func saigUnrollFrames(p *aig.Graph, nFrames int32) (*aig.Graph, []aig.Edge) {
	m := saigFramesMap(p, nFrames)
	q := aig.Aig_ManStart(aig.Aig_ManObjNumMax(p) * nFrames)
	q.Name, q.Spec = p.Name, p.Spec
	for f := int32(0); f < nFrames; f++ {
		aig.Aig_ObjSetFrames(m, nFrames, p.Const1(), f, q.Const1())
	}
	inputs, outputs := p.Inputs(), p.Outputs()
	nRegs := int(aig.Aig_ManRegNum(p))
	pis, los := inputs[:len(inputs)-nRegs], inputs[len(inputs)-nRegs:]
	lis := outputs[len(outputs)-nRegs:]
	for f := int32(0); f < nFrames; f++ {
		for _, pi := range pis {
			aig.Aig_ObjSetFrames(m, nFrames, pi, f, q.NewInput())
		}
	}
	for _, lo := range los {
		aig.Aig_ObjSetFrames(m, nFrames, lo, 0, q.NewInput())
	}
	nodes := p.Nodes()
	for f := int32(0); f < nFrames; f++ {
		for id, node := range nodes {
			e := aig.Edge(id) << 1
			if !p.ValidEdge(e) || node.Kind != aig.And {
				continue
			}
			a := aig.Aig_ObjChild0Frames(p, m, nFrames, e, f)
			b := aig.Aig_ObjChild1Frames(p, m, nFrames, e, f)
			aig.Aig_ObjSetFrames(m, nFrames, e, f, q.AND(a, b))
		}
		for i, li := range lis {
			next := aig.Aig_ObjChild0Frames(p, m, nFrames, li, f)
			if f < nFrames-1 {
				aig.Aig_ObjSetFrames(m, nFrames, los[i], f+1, next)
			}
		}
	}
	return q, m
}

// saigManUnrollCOI is Saig_ManUnrollCOI: the frames with every property
// output of the last frame, and no sweeping (dangling nodes stay, as they
// get SAT variables in ABC too). It returns the frames map (p->pObjCopies).
func saigManUnrollCOI(p *aig.Graph, nFrames int32) (*aig.Graph, []aig.Edge) {
	q, m := saigUnrollFrames(p, nFrames)
	outputs := p.Outputs()
	for _, po := range outputs[:len(outputs)-int(aig.Aig_ManRegNum(p))] {
		co := q.AddOutput(aig.Aig_ObjChild0Frames(p, m, nFrames, po, nFrames-1))
		aig.Aig_ObjSetFrames(m, nFrames, po, nFrames-1, co)
	}
	return q, m
}

// Saig_ManCreateIndMiter is saigConstr2.c's: two frames, and for each
// candidate an output that holds when the candidate is true in the first
// frame and false in the second.
func Saig_ManCreateIndMiter(p *aig.Graph, cands FunctionalCandidates) *aig.Graph {
	const nFrames = 2
	q, m := saigUnrollFrames(p, nFrames)
	for _, level := range cands {
		for _, cand := range level {
			n0 := m[nFrames*cand.NodeID()+0]
			n1 := m[nFrames*cand.NodeID()+1]
			fan0 := edgeNotCond(n0, cand.Complemented())
			fan1 := edgeNotCond(n1, !cand.Complemented())
			q.AddOutput(q.AND(fan0, fan1))
		}
	}
	aig.Aig_ManCleanup(q)
	return q
}

// saigManFilterUsingIndOne is Saig_ManFilterUsingIndOne_new: the candidate
// is inductive when its miter output cannot be 1.
func saigManFilterUsingIndOne(frames *aig.Graph, sat *bsat.Solver, c *cnf.Cnf_Dat_t, nConfs int32, counter int32) bool {
	co := aig.Aig_ManCo(frames, counter)
	lit := bsat.Positive(c.VarNums[co.NodeID()])
	return sat.SolveWithSourceLimits([]bsat.Literal{lit}, int64(nConfs), 0, 0, 0) == bsat.Unsatisfiable
}

// Saig_ManFilterUsingInd keeps the candidates proved inductive, in place.
func Saig_ManFilterUsingInd(p *aig.Graph, cands FunctionalCandidates, nConfs, nProps int32, fVerbose int32) FunctionalCandidates {
	frames := Saig_ManCreateIndMiter(p, cands)
	c := cnf.Cnf_DeriveSimple(frames, aig.Aig_ManCoNum(frames))
	sat := cnf.Cnf_DataWriteIntoSolver(c, 1, 0)
	if fVerbose != 0 {
		fmt.Printf("Filtered cands:  ")
	}
	counter := int32(0)
	for i, level := range cands {
		k2 := 0
		for _, cand := range level {
			keep := sat != nil && saigManFilterUsingIndOne(frames, sat, c, nConfs, counter)
			counter++
			if keep {
				level[k2] = cand
				k2++
				if fVerbose != 0 {
					compl := ""
					if cand.Complemented() {
						compl = "!"
					}
					fmt.Printf("%d:%s%d  ", i, compl, cand.NodeID())
				}
			}
		}
		cands[i] = level[:k2]
	}
	if fVerbose != 0 {
		fmt.Printf("\n")
		aig.Aig_ManPrintStats(frames)
	}
	return cands
}

func edgeNotCond(e aig.Edge, c bool) aig.Edge {
	if c {
		return e.Not()
	}
	return e
}

func saigCandidateCount(cands FunctionalCandidates) int {
	n := 0
	for _, level := range cands {
		n += len(level)
	}
	return n
}

// Ssw_ManFindDirectImplications is saigConstr2.c's: the nodes (and latch
// outputs) whose value is implied at the root once every property output of
// the unrolled frames is asserted, each taken once, latest frame first; then
// only the inductive ones are kept.
func Ssw_ManFindDirectImplications(p *aig.Graph, nFrames, nConfs, nProps, fVerbose int32) FunctionalCandidates {
	// Vec_VecAlloc reserves the levels; Vec_VecPush creates them on demand
	cands := make(FunctionalCandidates, 0, nFrames)
	frames, copies := saigManUnrollCOI(p, nFrames)
	c := cnf.Cnf_DeriveSimple(frames, 0)
	if sat := cnf.Cnf_DataWriteIntoSolver(c, 1, 0); sat != nil {
		inputs, outputs := p.Inputs(), p.Outputs()
		nRegs := int(aig.Aig_ManRegNum(p))
		piNum, poNum := len(inputs)-nRegs, len(outputs)-nRegs
		loIndex := make(map[int]int, nRegs)
		for i, lo := range inputs[piNum:] {
			loIndex[lo.NodeID()] = i
		}
		taken := make([]bool, aig.Aig_ManObjNumMax(p))
		nodes := p.Nodes()
		for f := int32(0); f < nFrames; f++ {
			for id, node := range nodes {
				e := aig.Edge(id) << 1
				if !p.ValidEdge(e) || (node.Kind != aig.Input && node.Kind != aig.And && node.Kind != aig.ExclusiveOr) {
					continue // Aig_ObjIsCand
				}
				if taken[id] {
					continue
				}
				repr := copies[int(nFrames)*id+int(nFrames-1-f)]
				if repr == aig.InvalidEdge {
					continue
				}
				v := c.VarNums[repr.NodeID()]
				if v < 0 {
					continue
				}
				value, assigned := sat.AssignedValue(v)
				if !assigned {
					continue
				}
				taken[id] = true
				if reg, ok := loIndex[id]; ok {
					taken[aig.Aig_ObjChild0(p, outputs[poNum+reg]).NodeID()] = true
				}
				for len(cands) <= int(f) {
					cands = append(cands, nil)
				}
				cands[f] = append(cands[f], edgeNotCond(e, value != repr.Complemented()))
			}
		}
		sat.Delete()
	}
	if fVerbose != 0 {
		fmt.Printf("Found %3d candidates.\n", saigCandidateCount(cands))
		for k, level := range cands {
			fmt.Printf("Level %d. Cands  =%d    \n", k, len(level))
		}
	}
	cands = Saig_ManFilterUsingInd(p, cands, nConfs, nProps, fVerbose)
	if n := saigCandidateCount(cands); n != 0 {
		fmt.Printf("Found %3d constraints after filtering.\n", n)
	}
	if fVerbose != 0 {
		for k, level := range cands {
			fmt.Printf("Level %d. Constr =%d    \n", k, len(level))
		}
	}
	return cands
}
