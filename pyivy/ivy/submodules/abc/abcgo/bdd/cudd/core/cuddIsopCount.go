package core

// CountISOPBounded uses the same authoritative Minato recursion as the ordinary
// BDD/ZDD ISOP operations, without building a ZDD cover. It returns nil,-1 when
// the cube or newly allocated BDD-node limit is exceeded. Failed allocations
// are rolled back; existing diagrams and manager allocation limits are retained.
func CountISOPBounded(dd *DdManager, lower, upper *DdNode, cubeLimit, nodeLimit int) (function *DdNode, count int) {
	count = -1
	if !dd.valid(lower, upper) || cubeLimit < 0 || nodeLimit < 0 || !Cudd_bddLeq(dd, lower, upper) {
		return
	}
	defer dd.beginAllocationAttempt(uint64(nodeLimit)).finish(&function)
	result := dd.isopBounded(lower, upper, false, cubeLimit)
	if result.function != nil {
		function, count = result.function, result.cubes
	}
	return
}
