package framgr

import "github.com/glycerine/abcgo/aig/aig"

// SweepForCEC runs the original FRA combinational manager with the CEC
// driver's current node and miter conflict budgets.
func SweepForCEC(graph *aig.Graph, nodeLimit, miterLimit int32) *aig.Graph {
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	pars.FDontShowBar, pars.FProve = 1, 1
	pars.NBTLimitNode, pars.NBTLimitMiter = nodeLimit, miterLimit
	return Fra_FraigPerform(graph, &pars)
}
