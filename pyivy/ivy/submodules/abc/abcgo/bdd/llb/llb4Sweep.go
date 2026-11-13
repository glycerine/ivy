package llb

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func Llb_Nonlin4SweepOrder(g *Aig_Man_t, counter *int32, saveAll int32) *Vec_Int_t {
	return Llb_Nonlin4FindOrder(g, counter)
}
func Llb_Nonlin4SweepOrder_rec(g *Aig_Man_t, e *Aig_Obj_t, order *Vec_Int_t, counter *int32, saveAll int32) {
	Llb_Nonlin4FindOrder_rec(g, e, order, counter)
}
func Llb4_Nonlin4SweepCutpoints(g *Aig_Man_t, order *Vec_Int_t, limit, verbose int32) int32 {
	if g == nil {
		return 0
	}
	return int32(vec.Vec_IntSize(Llb_Nonlin4CollectHighRefNodes(g, 2)))
}
func Llb_Nonlin4SweepPartitions_rec(dd *DdManager, g *Aig_Man_t, e *Aig_Obj_t, order *Vec_Int_t, roots *Vec_Ptr_t) *DdNode {
	if dd == nil || g == nil || e == nil {
		return nil
	}
	vars := make([]*core.DdNode, len(g.Inputs()))
	for i := range vars {
		vars[i] = core.Cudd_bddIthVar(dd, i)
	}
	return llbBDDForEdge(g, llbBuildBDDTable(g, dd, vars), *e)
}
func Llb_Nonlin4SweepPartitions(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, transition int32) *Vec_Ptr_t {
	return Llb_Nonlin4FindPartitions(dd, g, order, transition)
}
func Llb4_Nonlin4SweepBadMonitor(g *Aig_Man_t, order *Vec_Int_t, dd *DdManager) *DdNode {
	return Llb_Nonlin4ComputeBad(dd, g, order)
}
func Llb_Nonlin4SweepVars2Q(g *Aig_Man_t, order *Vec_Int_t, addLis int32) *Vec_Int_t {
	mask := vec.Vec_IntAlloc(0)
	vec.Vec_IntFill(mask, int(aig.Aig_ManObjNumMax(g)), 1)
	regs := int(g.RegisterCount())
	for _, edge := range g.Inputs()[len(g.Inputs())-regs:] {
		vec.Vec_IntWriteEntry(mask, int(vec.Vec_IntEntry(order, edge.NodeID())), 0)
	}
	if addLis != 0 {
		for _, edge := range g.Outputs()[len(g.Outputs())-regs:] {
			vec.Vec_IntWriteEntry(mask, int(vec.Vec_IntEntry(order, edge.NodeID())), 0)
		}
	}
	return mask
}
func Llb_Nonlin4SweepDeref(dd *DdManager, parts *Vec_Ptr_t) {
	if parts != nil {
		vec.Vec_PtrFree(parts)
	}
}
func Llb_Nonlin4SweepPrint(functions *Vec_Ptr_t) {
	if functions != nil {
		fmt.Printf("%d swept BDD functions\n", vec.Vec_PtrSize(functions))
	}
}
func Llb4_Nonlin4SweepBadStates(g *Aig_Man_t, order *Vec_Int_t, variables int32) *DdManager {
	if g == nil {
		return nil
	}
	dd := core.NewBDDManager(int(variables))
	if dd != nil {
		_ = Llb_Nonlin4ComputeBad(dd, g, order)
	}
	return dd
}
func Llb4_Nonlin4SweepGroups(g *Aig_Man_t, order *Vec_Int_t, variables int32, groups **Vec_Ptr_t, limit, verbose int32) *DdManager {
	if g == nil {
		return nil
	}
	dd := core.NewBDDManager(int(variables))
	if groups != nil {
		*groups = Llb_Nonlin4FindPartitions(dd, g, order, 0)
	}
	return dd
}
func Llb_Nonlin4SweepPrintSuppProfile(dd *DdManager, g *Aig_Man_t, order *Vec_Int_t, groups *Vec_Ptr_t, verbose int32) {
	Llb_Nonlin4PrintSuppProfile(dd, g, order, groups)
}
func Llb4_Nonlin4Sweep(g *Aig_Man_t, maxSweeps, maxClusters int32, dd **DdManager, order **Vec_Int_t, groups **Vec_Ptr_t, verbose int32) {
	Llb_Nonlin4Cluster(g, dd, order, groups, maxClusters, verbose)
}
func Llb4_Nonlin4SweepExperiment(g *Aig_Man_t) {
	if g == nil {
		return
	}
	var dd *DdManager
	var order *Vec_Int_t
	var groups *Vec_Ptr_t
	Llb4_Nonlin4Sweep(g, 0, 0, &dd, &order, &groups, 0)
	core.Cudd_Quit(dd)
}
