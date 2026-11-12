package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

// llbImageState retains the graph associated with the legacy iterative image
// API. The BDD manager itself owns all canonical nodes and variable ordering.
type llbImageState struct {
	graph *aig.Graph
	dd    *core.DdManager
}

var activeImage *llbImageState

func llbRelationalImage(g *aig.Graph, dd *core.DdManager, current *core.DdNode) *core.DdNode {
	if g == nil || dd == nil || current == nil {
		return nil
	}
	inputs := g.Inputs()
	registers := int(g.RegisterCount())
	piCount := len(inputs) - registers
	outputs := g.Outputs()
	if registers < 0 || piCount < 0 || len(outputs) < registers {
		return nil
	}
	variables := make([]*core.DdNode, len(inputs))
	for i, edge := range inputs {
		if edge.NodeID() < len(dd.Vars) {
			variables[i] = core.Cudd_bddIthVar(dd, edge.NodeID())
		}
	}
	functions := llbBuildBDDTable(g, dd, variables)
	if functions == nil {
		return nil
	}
	relation := current
	for i := 0; i < registers; i++ {
		next := outputs[len(outputs)-registers+i]
		if next.NodeID() >= len(dd.Vars) {
			return nil
		}
		value := llbBDDForEdge(g, functions, next)
		relation = core.Cudd_bddAnd(dd, relation, core.Cudd_bddXnor(dd, core.Cudd_bddIthVar(dd, next.NodeID()), value))
		if relation == nil {
			return nil
		}
	}
	cube := core.Cudd_ReadOne(dd)
	for _, edge := range inputs {
		cube = core.Cudd_bddAnd(dd, cube, core.Cudd_bddIthVar(dd, edge.NodeID()))
	}
	if cube == nil {
		return nil
	}
	image := core.Cudd_bddExistAbstract(dd, relation, cube)
	permutation := make([]int, len(dd.Vars))
	for i := range permutation {
		permutation[i] = i
	}
	for i := 0; i < registers; i++ {
		permutation[outputs[len(outputs)-registers+i].NodeID()] = inputs[piCount+i].NodeID()
	}
	return core.Cudd_bddPermute(dd, image, permutation)
}

// Llb_NonlinImage computes an exact transition image for the supplied current
// state set. The old score-driven partition scheduler is subsumed by canonical
// BDD conjunction and existential abstraction.
func Llb_NonlinImage(g *Aig_Man_t, leaves, roots *Vec_Ptr_t, varsToQuant *int32, dd *DdManager, current *DdNode, reorder, verbose int32, order *int32) *DdNode {
	return llbRelationalImage(g, dd, current)
}

func Llb_NonlinImageStart(g *Aig_Man_t, leaves, roots *Vec_Ptr_t, varsToQuant *int32, order *int32, first int32, timeout abctime) *DdManager {
	if g == nil {
		return nil
	}
	dd := core.NewBDDManager(int(aig.Aig_ManObjNumMax(g)))
	if dd == nil {
		return nil
	}
	activeImage = &llbImageState{graph: g, dd: dd}
	return dd
}

func Llb_NonlinImageCompute(current *DdNode, reorder, drop, verbose int32, order *int32) *DdNode {
	if activeImage == nil {
		return nil
	}
	return llbRelationalImage(activeImage.graph, activeImage.dd, current)
}

func Llb_NonlinImageQuit() {
	if activeImage != nil {
		core.Cudd_Quit(activeImage.dd)
		activeImage = nil
	}
}

// Retained utility helpers used by the non-linear reachability package.
func Llb_NonlinCreateCube1(p *Llb_Mgr_t, part *Llb_Prt_t) *DdNode { return llbNonlinCube(p, part, nil) }
func Llb_NonlinCreateCube2(p *Llb_Mgr_t, first, second *Llb_Prt_t) *DdNode {
	return llbNonlinCube(p, first, second)
}
func llbNonlinCube(p *Llb_Mgr_t, first, second *Llb_Prt_t) *DdNode {
	if p == nil || p.dd == nil {
		return nil
	}
	cube := core.Cudd_ReadOne(p.dd)
	for _, part := range []*Llb_Prt_t{first, second} {
		if part == nil || part.vVars == nil {
			continue
		}
		for _, id := range vec.Vec_IntArray(part.vVars) {
			cube = core.Cudd_bddAnd(p.dd, cube, core.Cudd_bddIthVar(p.dd, int(id)))
		}
	}
	return cube
}

type Llb_Var_t struct {
	iVar, nScore int32
	vParts       *Vec_Int_t
}
type Llb_Prt_t struct {
	iPart, nSize int32
	bFunc        *DdNode
	vVars        *Vec_Int_t
}
type Llb_Mgr_t struct {
	pAig                       *Aig_Man_t
	vLeaves, vRoots            *Vec_Ptr_t
	dd                         *DdManager
	pVars2Q                    *int32
	pParts                     []*Llb_Prt_t
	pVars                      []*Llb_Var_t
	iPartFree, nVars, nSuppMax int32
	pSupp                      []int32
}

func Llb_MgrVar(p *Llb_Mgr_t, i int32) *Llb_Var_t {
	if p == nil || i < 0 || int(i) >= len(p.pVars) {
		return nil
	}
	return p.pVars[i]
}
func Llb_MgrPart(p *Llb_Mgr_t, i int32) *Llb_Prt_t {
	if p == nil || i < 0 || int(i) >= len(p.pParts) {
		return nil
	}
	return p.pParts[i]
}
func Llb_NonlinAlloc(g *Aig_Man_t, leaves, roots *Vec_Ptr_t, vars *int32, dd *DdManager) *Llb_Mgr_t {
	manager := &Llb_Mgr_t{pAig: g, vLeaves: leaves, vRoots: roots, pVars2Q: vars, dd: dd}
	if dd != nil {
		manager.nVars = int32(len(dd.Vars))
		manager.pVars = make([]*Llb_Var_t, len(dd.Vars))
		for i := range manager.pVars {
			manager.pVars[i] = &Llb_Var_t{iVar: int32(i), vParts: vec.Vec_IntAlloc(2)}
		}
	}
	return manager
}
func Llb_NonlinStart(p *Llb_Mgr_t) int32 {
	if p == nil || p.dd == nil {
		return 0
	}
	return 1
}
func Llb_NonlinFree(p *Llb_Mgr_t) {
	if p != nil {
		p.pParts = nil
		p.pVars = nil
	}
}
func Llb_NonlinAddPartition(p *Llb_Mgr_t, i int32, f *DdNode) {
	if p == nil {
		return
	}
	for len(p.pParts) <= int(i) {
		p.pParts = append(p.pParts, nil)
	}
	part := &Llb_Prt_t{iPart: i, bFunc: f}
	if f != nil && p.dd != nil {
		support := p.dd.Support(f)
		part.vVars = vec.Vec_IntAlloc(len(support))
		part.nSize = int32(core.Cudd_DagSize(f))
		for _, index := range support {
			vec.Vec_IntPush(part.vVars, int32(index))
			if index < len(p.pVars) {
				vec.Vec_IntPush(p.pVars[index].vParts, i)
			}
		}
	}
	p.pParts[i] = part
	p.iPartFree = int32(len(p.pParts))
}
func Llb_NonlinAddPair(p *Llb_Mgr_t, f *DdNode, part, variable int32) {
	if p == nil || variable < 0 {
		return
	}
	Llb_NonlinAddPartition(p, part, f)
	for len(p.pVars) <= int(variable) {
		p.pVars = append(p.pVars, nil)
	}
	if p.pVars[variable] == nil {
		p.pVars[variable] = &Llb_Var_t{iVar: variable, vParts: vec.Vec_IntAlloc(2)}
	}
	for _, existing := range vec.Vec_IntArray(p.pVars[variable].vParts) {
		if existing == part {
			return
		}
	}
	vec.Vec_IntPush(p.pVars[variable].vParts, part)
}
func Llb_NonlinAddPartitionPair(p *Llb_Mgr_t, part int32, f *DdNode) {
	Llb_NonlinAddPartition(p, part, f)
}
func Llb_NonlinHasSingletonVars(p *Llb_Mgr_t, part *Llb_Prt_t) int32 {
	if p != nil && p.dd != nil && part != nil && part.bFunc != nil && core.Cudd_SupportSize(p.dd, part.bFunc) == 1 {
		return 1
	}
	return 0
}
func Llb_NonlinQuantify1(p *Llb_Mgr_t, part *Llb_Prt_t, subset int32) int32 {
	if p == nil || p.dd == nil || part == nil || part.bFunc == nil {
		return 0
	}
	support := p.dd.Support(part.bFunc)
	if len(support) != 1 {
		return 0
	}
	part.bFunc = core.Cudd_bddExistAbstract(p.dd, part.bFunc, core.Cudd_bddIthVar(p.dd, support[0]))
	if part.bFunc == nil {
		return 0
	}
	part.vVars = vec.Vec_IntAlloc(0)
	part.nSize = int32(core.Cudd_DagSize(part.bFunc))
	return 1
}
func Llb_NonlinQuantify2(p *Llb_Mgr_t, first, second *Llb_Prt_t) int32 {
	if p == nil || first == nil || second == nil {
		return 0
	}
	first.bFunc = core.Cudd_bddAnd(p.dd, first.bFunc, second.bFunc)
	if first.bFunc == nil {
		return 0
	}
	Llb_Nonlin4RemovePart(p, second)
	first.vVars = vec.Vec_IntAlloc(len(p.dd.Support(first.bFunc)))
	for _, index := range p.dd.Support(first.bFunc) {
		vec.Vec_IntPush(first.vVars, int32(index))
	}
	first.nSize = int32(core.Cudd_DagSize(first.bFunc))
	return 1
}
func Llb_NonlinNextPartitions(p *Llb_Mgr_t, first, second **Llb_Prt_t) int32 {
	if p == nil {
		return 0
	}
	var previous *Llb_Prt_t
	for _, part := range p.pParts {
		if part == nil {
			continue
		}
		if previous == nil {
			previous = part
			continue
		}
		if first != nil {
			*first = previous
		}
		if second != nil {
			*second = part
		}
		return 1
	}
	return 0
}
func Llb_NonlinRecomputeScores(p *Llb_Mgr_t)                {}
func Llb_NonlinVerifyScores(p *Llb_Mgr_t)                   {}
func Llb_NonlinReorder(dd *DdManager, twice, verbose int32) {}
