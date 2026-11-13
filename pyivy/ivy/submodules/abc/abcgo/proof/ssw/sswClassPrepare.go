package ssw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ssw_ClassesPrepareSimple(g *aig.Graph, latchOnly, maxLevels int32) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	inputs := g.Inputs()
	nRegs := int(g.RegisterCount())
	latches := make(map[int]bool, nRegs)
	for _, obj := range inputs[len(inputs)-nRegs:] {
		latches[obj.NodeID()] = true
	}
	for _, obj := range g.LiveEdges() {
		n := g.Node(obj)
		if latchOnly != 0 {
			if !latches[obj.NodeID()] {
				continue
			}
		} else {
			if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && !latches[obj.NodeID()] {
				continue
			}
			if maxLevels != 0 && int32(n.Level) > maxLevels {
				continue
			}
		}
		Ssw_ObjSetConst1Cand(g, obj)
		p.nCands1++
	}
	return p
}
func Ssw_ClassesPrepareFromReprs(g *aig.Graph) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	classes := make(map[int][]aig.Edge)
	for _, obj := range g.LiveEdges() {
		repr := aig.Aig_ObjRepr(g, obj)
		if repr == g.Const1() {
			p.nCands1++
			continue
		}
		if repr != aig.InvalidEdge {
			id := repr.NodeID()
			if classes[id] == nil {
				classes[id] = append(classes[id], repr)
			}
			classes[id] = append(classes[id], obj)
		}
	}
	for _, class := range classes {
		Ssw_ObjAddClass(p, class[0], class, int32(len(class)))
	}
	return p
}

func Ssw_ClassesPrepareTargets(g *aig.Graph) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	outputs := g.Outputs()
	numPOs := len(outputs) - int(g.RegisterCount())
	for _, obj := range outputs[:numPOs] {
		fanin := g.Node(obj).Fanins[0] &^ 1
		Ssw_ObjSetConst1Cand(g, fanin)
		p.nCands1++
	}
	return p
}
func Ssw_ClassesPreparePairs(g *aig.Graph, classes []*vec.Vec_Int_t) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	if len(classes) > 0 && classes[0] != nil {
		constants := classes[0]
		for i := 0; i < vec.Vec_IntSize(constants); i++ {
			id := vec.Vec_IntEntry(constants, i)
			if (i == 0) != (id == 0) {
				panic("ssw: invalid constant class")
			}
			if i == 0 {
				continue
			}
			Ssw_ObjSetConst1Cand(g, aig.Edge(id)<<1)
			p.nCands1++
		}
	}
	for i := 1; i < len(classes); i++ {
		v := classes[i]
		if v == nil {
			continue
		}
		if vec.Vec_IntSize(v) < 2 {
			panic("ssw: class must contain at least two nodes")
		}
		group := make([]aig.Edge, vec.Vec_IntSize(v))
		for i := range group {
			id := vec.Vec_IntEntry(v, i)
			if id < 0 || (i > 0 && id <= vec.Vec_IntEntry(v, i-1)) {
				panic("ssw: class must be ordered")
			}
			group[i] = aig.Edge(id) << 1
			if i > 0 {
				aig.Aig_ObjSetRepr(g, group[i], group[0])
			}
		}
		Ssw_ObjAddClass(p, group[0], group, int32(len(group)))
	}
	return p
}
func Ssw_ClassesPreparePairsSimple(g *aig.Graph, pairs *vec.Vec_Int_t) *Ssw_Cla_t {
	p := Ssw_ClassesStart(g)
	if vec.Vec_IntSize(pairs)%2 != 0 {
		panic("ssw: incomplete pair")
	}
	for i := 0; i < vec.Vec_IntSize(pairs); i += 2 {
		a, b := vec.Vec_IntEntry(pairs, i), vec.Vec_IntEntry(pairs, i+1)
		if a < 0 || a >= b {
			panic("ssw: pair must be ordered")
		}
		repr, obj := aig.Edge(a)<<1, aig.Edge(b)<<1
		aig.Aig_ObjSetRepr(g, obj, repr)
		Ssw_ObjAddClass(p, repr, []aig.Edge{repr, obj}, 2)
	}
	return p
}

// Ssw_ClassesPrepare derives candidate classes through sequential simulation.
func Ssw_ClassesPrepare(g *aig.Graph, framesK, latchOnly, constOnly, outputOnly, maxLevels, verbose int32) *Ssw_Cla_t {
	return sswClassesPrepare(g, framesK, latchOnly, constOnly, outputOnly, maxLevels, verbose, nil)
}

func sswClassesPrepare(g *aig.Graph, framesK, latchOnly, constOnly, outputOnly, maxLevels, verbose int32, pars *Ssw_Pars_t) *Ssw_Cla_t {
	if sswDeadlineExpired(pars) {
		return nil
	}
	p := Ssw_ClassesStart(g)
	p.fConstCorr = constOnly
	nFrames := max(framesK, 4)
	sml := Ssw_SmlStart(g, 0, nFrames, 2)
	Ssw_SmlInitialize(sml, 1)
	if !sswSimulateOne(sml, pars) {
		Ssw_SmlStop(sml)
		Ssw_ClassesStop(p)
		return nil
	}
	sml.fNonConstOut = Ssw_SmlCheckNonConstOutputs(sml)
	if sswDeadlineExpired(pars) {
		Ssw_SmlStop(sml)
		Ssw_ClassesStop(p)
		return nil
	}
	Ssw_ClassesSetData(p, sml,
		func(data any, obj aig.Edge) uint32 { return Ssw_SmlObjHashWord(data.(*Ssw_Sml_t), obj) },
		func(data any, obj aig.Edge) int32 { return Ssw_SmlObjIsConstWord(data.(*Ssw_Sml_t), obj) },
		func(data any, a, b aig.Edge) int32 { return Ssw_SmlObjsAreEqualWord(data.(*Ssw_Sml_t), a, b) })
	candidates := vec.Vec_PtrAlloc(1000)
	inputs := g.Inputs()
	nRegs := int(g.RegisterCount())
	latches := make(map[int]bool, nRegs)
	for _, obj := range inputs[len(inputs)-nRegs:] {
		latches[obj.NodeID()] = true
	}
	selected := map[int]bool{}
	if outputOnly != 0 {
		outputs := g.Outputs()
		for _, out := range outputs[:len(outputs)-nRegs] {
			driver := g.Node(out).Fanins[0] &^ 1
			if aig.Aig_ObjIsCand(g, driver) {
				selected[driver.NodeID()] = true
			}
		}
	}
	for _, obj := range g.LiveEdges() {
		node := g.Node(obj)
		if outputOnly != 0 {
			if !selected[obj.NodeID()] {
				continue
			}
		} else if latchOnly != 0 {
			if !latches[obj.NodeID()] {
				continue
			}
		} else {
			if !aig.Aig_ObjIsNode(g, obj) && node.Kind != aig.Input {
				continue
			}
			if maxLevels != 0 && int32(node.Level) > maxLevels {
				continue
			}
		}
		vec.Vec_PtrPush(candidates, obj)
	}
	Ssw_ClassesPrepareRehash(p, candidates, constOnly)
	rounds := int32(1)
	for ; rounds < 16; rounds++ {
		if sswDeadlineExpired(pars) {
			Ssw_SmlStop(sml)
			Ssw_ClassesStop(p)
			return nil
		}
		vec.Vec_PtrClear(candidates)
		for _, obj := range g.LiveEdges() {
			if Ssw_ObjIsConst1Cand(g, obj) != 0 {
				vec.Vec_PtrPush(candidates, obj)
			}
		}
		if int32(vec.Vec_PtrSize(candidates)) != p.nCands1 {
			panic("ssw: inconsistent constant candidates")
		}
		Ssw_SmlReinitialize(sml)
		if !sswSimulateOne(sml, pars) {
			Ssw_SmlStop(sml)
			Ssw_ClassesStop(p)
			return nil
		}
		sml.fNonConstOut = Ssw_SmlCheckNonConstOutputs(sml)
		if Ssw_ClassesPrepareRehash(p, candidates, constOnly) == 0 {
			break
		}
	}
	if verbose != 0 {
		fmt.Printf("Simulation of %d frames with 2 words (%d rounds).\n", nFrames, rounds-1)
	}
	Ssw_SmlStop(sml)
	Ssw_ClassesCheck(p)
	return p
}
